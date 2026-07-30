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

## DESIGN-GAPs blocking acceptance

- R4 states that the evaluation instant is in the canonical payload. The implemented canonical
  request payload contains only client intent fields; `evaluated_at` is server-authored in the
  receipt. Replay therefore cannot call the live transition with the original time from the
  declared input tuple.
- Live transitions consume server-side multiplier contributions, Route/Commons projection state,
  and other resolved policy inputs that are not reconstructible from `(genesis, canonical payload,
  catalog)`. A logged receipt can detect disagreement only after the verifier has a normative way
  to reconstruct or snapshot those inputs; it cannot serve as both missing input and oracle.
- The “same transition entry in both runtimes” is not currently an owned interface: Go owns the
  authoritative transition while TypeScript owns presentation prediction. The RFC must define the
  replay-input record and precise shared-kernel boundary before AC2 can be implemented without a
  second, drifting engine.

The immutable genesis table is technically separable, but R1 requires it at every pin site and R2
defines its consumer. Landing only storage would not close the stated replay DESIGN-GAP.

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
