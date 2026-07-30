# RFC: Run Genesis & Replay Verification

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-07-30
- **Design refs:** `design/05 §6`, `design/08 §6` (verification-is-replay, shipped validator), `design/research/speedrun-governance.md`
- **Depends on:** Leaderboards & Balance Epochs (implementing — run_log, pins, boards), Prestige & Exits (implementing — run lifecycle), Account Bootstrap (implementing — imported exclusion)
- **Closes:** the "no immutable initial run state" DESIGN-GAP recorded in `docs/leaderboards-and-epochs.md` and `planning/websocket-transport-and-fanout/log.md` — the last blocker before L4's verifier can honestly exist
- **Planning:** `planning/run-genesis-and-replay/` (once implementing)

## Summary

Replay verification needs a replay *starting point*. Catalog initials can't reconstruct it (later
runs fold in Founder-carried Network/Reputation effects), and revisions prune to the last 5. This
RFC adds the missing object — an immutable genesis snapshot per run — and specifies the verifier
that consumes it, completing L2/L4's contracts.

## Specification

### R1 — The genesis record

`run_genesis(company_stream_id, run_seq, PRIMARY KEY (company_stream_id, run_seq), state bytea, version int, constants_hash, created_at)` + the standard `reject_immutable_change` trigger.

- **Written in the same transaction as the run's pin**, both sites: account creation (run 1 —
  the initial revision bytes are the genesis) and the Exit transaction (run N+1 — `newEncoded`,
  the same bytes committed as the first revision of the new run). Genesis is definitionally
  byte-identical to revision `first(run)` — asserted, then the revision may prune freely.
- Import (A-D4a) writes through the same seam: the normalized imported state IS that run's
  genesis (the founder is ranked-excluded, but the record keeps the invariant total: **every
  pinned run has exactly one genesis**, no special cases).
- Founder-scope state at run start is NOT snapshotted: replay of a single run needs the company
  transition function only — Founder effects are already baked into the genesis bytes by D6
  assembly. (Cross-run Founder verification is a named non-goal; the Founder stream's own
  events/log remain the audit surface.)

### R2 — The verifier (completes L4)

`verify(genesis, runLog, catalogBytes, kernelVersion) → verdict` in the shared kernel, both
runtimes:

1. Decode genesis under `constants_hash` catalog (exact bytes from `catalog_artifacts`).
2. Apply the run log in seq order through the SAME transition entry the live engine uses —
   evaluation-then-mutation per canonical payload, comparing each computed receipt byte-for-byte
   with the logged receipt.
3. Terminal check: `max(seq) == run_ended.terminal_seq`, final state's run facts match the
   `run_ended` payload (RTA, attended, tier, lifetime value).

Verdict = the closed six-cause union: `verified | log_gap | state_divergence | constants_mismatch
| clock_violation | engine_mismatch` (L2's five + `verified`). Version rule per L2b: any
`run_version_drift` row → `engine_mismatch` without replaying. Pre-timer runs (P6a) verify
normally but are **excluded from time-keyed boards at projection** (the exclusion this RFC makes
executable: `run_ended.pre_timer` → time-board rows rejected, count boards fine).

### R3 — Verification is a projection

`verify_and_project(run)` runs server-side post-`run_ended` (queue: the existing outbox/claim
pattern over a `verification_queue` table — claim, verify, project via `ProjectVerifiedRun`,
mark; poison → dead-letter with InvariantSink, per the relay's remediated discipline). The
shipped player validator is the same kernel function over the same fixtures (L4 unchanged);
`run_log_archive` compaction happens at mark time (verified runs), completing L1's retention
story.

### R4 — Determinism obligations this surfaces

Replay equality holds only if the transition is a pure function of `(state, canonical payload,
catalog, evaluation instant)`. The evaluation instant is IN the canonical payload (server-stamped
`evaluated_at` — already part of receipts); anything the live path reads outside that tuple
(projections, commons snapshots, registry state) must already be reflected in logged receipts —
the verifier compares receipts, so any hidden input shows up as `state_divergence` on honest
replays. AC5 hunts these before ship.

## Executable contracts (answering the 2026-07-30 bounce)

### RA — The replay-input record (closes bounce items 1 and 2)

`run_log` gains `replay_inputs jsonb NOT NULL` — the **closed, versioned record of every
server-resolved input the transition consumed**, written in the same transaction as the log row:

```json
{"v": 1, "evaluated_at_ms": int,
 "contributions": [{"slot": string, "ppm_or_value": canonical, "source": string}],
 "policy": {"commons_modifier": canonical|null, "route_context_version": int,
            "registry_decisions": [...], "drift_declined_count": int}}
```

Field set grows by RFC exactly like the event-kind registry; the loader/validator exact-key
rejects unknown fields per version. The correct mental model: **canonical payload = what the
player said; replay_inputs = what the server resolved; receipt = what happened.** Inputs and
oracle are now distinct objects — the both-at-once problem is dissolved. Backfill: none (no
verifier consumed old rows; rows predating the column are unrankable-by-construction and the
column is NOT NULL from its migration forward — pre-column runs verify `log_gap`).

### RB — The owned transition boundary (closes bounce item 3)

New shared-kernel entry, **the only legal way to apply a logged mutation**:

`ApplyLogged(state, canonicalPayload, catalog, replayInputs) → (state', receipt)`

- **The live Go path is refactored to call `ApplyLogged` itself**: it computes `replayInputs`
  from live sources (projections, contribution providers, clock), persists them to the log row,
  and applies through the same function the verifier uses. One engine, no drift possible on the
  Go side — the live path and replay path are the same code by construction, which is the
  strongest available answer to "same transition entry".
- TS implements `ApplyLogged` for the shipped validator; parity is enforced by AC2's full-run
  golden fixture plus the existing golden-vector regime (the TS shell's *prediction* path remains
  presentation-only and unchanged — the validator is a separate consumer of the kernel).
- Anything the Go live path reads OUTSIDE `ApplyLogged`'s four arguments is definitionally a bug
  (AC6's hidden-input hunt becomes a structural rule, not a test-time hope).

Sequencing note: RA/RB land FIRST (a refactor of the live engine + log schema), then genesis
storage, then the verifier — the bounce's "storage alone would not close the gap" is accepted and
built into the implementation order.

## Acceptance criteria

1. Genesis invariants: every pinned run has exactly one genesis, byte-identical to its first
   revision (property over the integration matrix); immutability trigger; fault injection between
   pin and genesis leaves neither.
2. A full recorded run (fixture: ≥50 mixed intents incl. offers, gates, manual batches, an
   accrual-heavy gap) verifies `verified` in Go AND TS from the same fixture bytes.
3. Each failure cause has a committed corpus fixture (tampered log, gap, wrong catalog, drifted
   version, clock regression) asserted by both suites.
4. Pre-timer run: verifies, projects to count boards, rejected from time boards.
5. Queue: crash between verify and mark re-verifies idempotently; poison run dead-letters with a
   report; boards never contain an unverified run (structural: projection only via the queue).
6. A deliberately-hidden-input regression (a transition reading wall clock) is caught by the
   receipt comparison fixture.

## Open questions

- Verification queue latency targets and batch size — implementation freedom.
- Category terminal predicates (L7 catalog) consume the verified terminal state; their content
  pass is Leaderboards follow-up, not this RFC.

## Changelog

- 2026-07-30: created (draft) — closes the initial-run-state DESIGN-GAP; completes L1/L2/L4.
- 2026-07-30: Codex bounce answered — RA (replay_inputs record: payload=said, inputs=resolved, receipt=happened), RB (ApplyLogged as the owned shared-kernel boundary the LIVE path itself calls; implementation order RA/RB → genesis → verifier).
