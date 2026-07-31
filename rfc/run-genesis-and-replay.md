# RFC: Run Genesis & Replay Verification

- **Status:** implementing
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

## DESIGN-GAPs blocking acceptance (live-surface pass, 2026-07-31)

RA/RB name the right ownership boundary, but their declared closed inputs cannot represent values
the shipped transition demonstrably consumes. Because `replay_inputs` becomes immutable history,
these are owner contracts, not implementation freedom:

### C1 — command/execution identity

`canonical_payload` deliberately excludes `intent_id`; Company `State` contains neither
`stream_id` nor `owner_id`; receipts/events also consume current revision and terminal
`run_log_seq`. Specify the exact execution envelope available inside `ApplyLogged`.

**Proposed:** RA adds `command: {intent_id, company_stream_id, founder_id, revision, run_seq,
run_log_seq}` with exact UUID/safe-integer validation. These are authoritative coordinates, not
player inputs. The four arguments remain unchanged because the envelope is inside `replayInputs`.

### C2 — evaluation mode

`online|offline` changes accrual efficiency, cap, and Compute Credits. It is absent from RA.

**Proposed:** required `evaluation_mode: "online"|"offline"` beside `evaluated_at_ms`; clock
validation compares that instant to the state cursor inside `ApplyLogged`.

### C3 — catalog means an artifact bundle

The live path consumes economy, Routes, Commons, Prestige, faction, and Guild artifacts under the
same constants hash. A singular economy catalog cannot apply several accepted intent kinds.

**Proposed:** RB's third argument is a closed `CatalogBundle` resolved from exact
`catalog_artifacts`, with all six typed loaders and a required shared constants hash. Missing or
extra artifact kinds fail before transition.

### C4 — closed per-intent resolved policy

The proposed generic policy object omits live inputs: Commons participation weight, Guild
settlement batch/membership, compact tithe band, and the exact resolved values used by offer and
route paths. `registry_decisions: [...]` has no item schema.

**Proposed:** replace `policy` with a discriminated `resolved` union keyed by intent kind. Every
arm has exact keys; common accrual keys are Commons weight (nullable), ordered Guild settlement
batch (possibly empty), contributions, and route-context version. Offer/cross-gate arms add the
closed offer inputs. Empty arrays remain explicit. New arms require a new replay-input version.

### C5 — terminal Exit and Founder carry

An Exit receipt contains the next Company snapshot. Its construction consumes Founder
Reputation/Network/Advisor state, executed-route projection facts, and the server's current hash;
R1 explicitly does not snapshot Founder state. The ended run therefore cannot reproduce its final
logged receipt from the stated four arguments.

**Proposed:** terminal `resolved` carries the minimal closed Founder carry view consumed by
`ComputeTerms`/`NewRunState`, executed route IDs, selected exit terms inputs, and next constants
hash. `ApplyLogged` still computes both the final current-run state and next-run snapshot; it does
not log the resulting receipt or next state as an input. Specify whether Founder mutation itself
is verified or remains the stated cross-run non-goal.

### C6 — pre-column rows versus NOT NULL

“No backfill; pre-column runs return `log_gap`” conflicts literally with adding a table-level
`NOT NULL` column to existing rows.

**Proposed:** historical rows retain SQL NULL; a database trigger/check rejects NULL on every new
insert after the migration. The reader maps legacy NULL to `log_gap`. Archive storage preserves
the same nullable legacy marker.

### C7 — events are transition output

The live shared result includes events, but RB returns only state and receipt. Silent event drift
would corrupt projections while receipt parity stayed green.

**Proposed:** `ApplyLogged` returns `(state', receipt, events)` with canonical event envelopes;
run-log replay compares receipt bytes and event bytes/order against immutable `events` rows. If
events are intentionally excluded, name a separate owned event verifier and its join key.

### C8 — accrual-hook order and extension closure

The shipped chain is constructed as Prestige → faction → optional Guild → ServiceOption extras;
only the first two are tested. Commons is currently an unnamed extra even though it mutates saved
Solidarity and emits events.

**Proposed:** declare and test the Phase-0 order Prestige → faction → Guild → Commons, forbid
unregistered extras in replayable production, and grow a closed hook-kind registry by RFC. The
same ordered registry constructs live Go and replay Go; TS implements the identical order.

## Owner ruling on C1–C8 (2026-07-31)

**All eight proposed contracts are ACCEPTED as written and are now normative**, with one addition:

- **C5's open question is ruled: cross-run Founder verification remains a non-goal** (R1's stance
  stands). The founder-carry view in the terminal `resolved` arm exists so the Exit receipt —
  including the next-run snapshot it embeds — reproduces byte-for-byte; the receipt comparison
  therefore covers every founder-DERIVED output. The Founder stream's own integrity remains the
  audit surface of its events and log; verifying Founder mutations end-to-end would require a
  founder genesis + founder log and is a successor RFC if the need ever materializes.
- C7 is confirmed in its strong form: `ApplyLogged` returns `(state', receipt, events)` and replay
  compares event bytes AND order against the immutable `events` rows — the guild rounds proved
  event drift is a real, silent corruption channel; receipt-only parity is insufficient.
- C8's declared Phase-0 order **Prestige → faction → Guild → Commons** also discharges the
  faction-round LOW ("pin accrual-hook chain order"); the closed hook registry grows by RFC like
  every other registry.

RA/RB are amended by these contracts wherever they conflict; the RA `replay_inputs` schema version
starts at the post-C1–C8 shape (there is no pre-C1 producer to preserve).

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
- 2026-07-31: live-surface acceptance pass found C1–C8: command identity, evaluation mode, artifact bundle, per-intent resolved inputs, terminal Founder carry, legacy NULL semantics, event parity, and hook-order closure remain owner decisions before RA/RB can be persisted.
- 2026-07-31: owner ruling — all eight proposals accepted as normative; C5's Founder question ruled (cross-run verification stays a non-goal; the carry view exists for receipt byte-parity); C7 confirmed strong-form; C8 discharges the hook-order LOW.
- 2026-07-31: RA/RB Go landing reviewed — reroute APPROVED; rulings C5a (founder-hash coherence: live fail-closed + carry carries founder_constants_hash asserted in-boundary), C4a (non-empty batches representable), C4b (route_hint arm deleted), RB-1 (prestige runtime required), C3a (bundle hash recomputed over bytes), RB-2 (stock resource derived in-boundary); run_log gains the standard immutability trigger.
