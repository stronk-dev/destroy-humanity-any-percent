# RFC: Founder Attendance Foundation (the shared cross-run clock)

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-05
- **Design refs:** `design/02 §2` (Founder scope, persists), `design/02 §9` (attended vs offline — the banked-time model)
- **Depends on:** Save + Run Genesis + ApplyFounderLogged (implemented — the Founder mutation boundary)
- **Unblocks:** Pet Care C10 (pet decay clock), Minigame Platform C26 (cross-run faucet quota) — TWO consumers, which is why it's its own primitive, not owned by either.
- **Planning:** `planning/founder-attendance/` (once implementing)
- **Created because:** Pet Care C10 + Minigame C26 both showed a server timestamp is not an attendance authority (elapsed wall time counts offline gaps; the command instant advances by zero). A shared clock two systems need deserves deliberate design, not a ruling squeezed into either bounce.

## Summary

The Company stream derives attended time from the online-session gap between accrual evaluations
(`catchup_ceiling_ms`). The Founder stream has no evaluation cadence — Founder commands are
sporadic — so it needs its own attendance authority. This RFC defines it once, replay-safe, for
every Founder-scoped consumer.

## Specification

### FA1 — The definition (Founder attended = summed run-attended, flushed at Exit)

**A founder is "attended" exactly when one of their company runs is attended.** `founder_attended_ms`
is the Founder-stream cursor = the sum of every completed run's Attended Time (P6's already-computed
per-run attended, a `run_ended` fact) PLUS the current run's attended-so-far. This is the honest
definition: pets decay and faucets tick on the founder's REAL presence, which is exactly their
runs' attended time — not care-command frequency (which would make decay meaningless) and not wall
clock (which would rot pets during genuine absence).

### FA2 — Replay independence (the C1 requirement, preserved)

The Founder stream never READS live Company state (C1's law). Instead:
- At **Exit**, the multi-stream transaction (which already touches both streams) advances
  `founder_attended_ms += run_ended.attended_ms` — a Founder-stream write of a Company-stream
  FACT, at the one boundary that legitimately spans both. So the Founder cursor is durably correct
  as of the last Exit.
- **Between Exits** (during a run), a Founder command that needs current attendance reads the
  cursor's last-Exit value PLUS `current_run_partial_attended` — which is passed as a
  RESOLVED INPUT into `ApplyFounderLogged` (frozen in the founder replay_inputs, the RA pattern),
  computed server-side from the live run at command time. The founder command never reaches into
  the Company stream; the partial is a resolved input, exactly as combat's scaling inputs are.

Founder replay is therefore independent: it consumes `founder_attended_ms` (Founder state) +
the frozen partial (Founder replay_inputs) — never a live Company read.

### FA3 — Offline correctness

Because the source is run Attended Time (which already subtracts offline spans, P6), offline is
correct by construction: a founder away for a week accrues zero attended time (their runs weren't
attended), so pets don't rot and faucet quotas don't tick. No separate offline computation.

### FA4 — The consumer contract

Consumers (pet decay, faucet quota) read `founder_attended_ms` (+ the frozen partial where they
evaluate mid-run) and compute deltas on the fixed attended grid (the provision-grid
partition-invariance). Adding a consumer adds no new attendance source — one clock, many readers.

## Acceptance criteria

1. `founder_attended_ms` advances by exactly `run_ended.attended_ms` at each Exit (fixture); a
   founder with N completed runs has the sum.
2. Mid-run Founder command: the resolved partial is frozen in founder replay_inputs; replay
   reproduces the same attendance without a live Company read (independence proof).
3. Offline: a founder idle across a week accrues zero attended (their runs weren't attended);
   no pet decay, no faucet tick over that span.
4. Both consumers (a pet-decay fixture, a faucet-quota fixture) read one clock; adding the second
   reader introduces no second source.
5. Migration adds the cursor to Founder save; corpus; the Exit advance is byte-parity Go/TS.

## Acceptance blockers (Codex review, 2026-08-05)

The shared-clock direction is correct, but the draft is not yet safe to implement. The existing
engine already persists completed-run attended time, and the mid-run snapshot crosses the same
Founder/Company lock seam as Exit. The contracts below close those facts rather than adding a
second clock or relying on timing.

### A1 — `age_ms` already is the completed Founder-attendance authority

`finishExitResolved` already computes `prestigecore.AttendedMS(company, now)` and adds it to
`founder.AgeMS`; the canonical wire field is `age_ms`, and achievements already consume it as a
career counter. Adding a separately persisted `founder_attended_ms` cursor would make two fields
claim authority over the same sum, with no invariant capable of repairing drift after one write
path misses the other. AC5's proposed migration therefore creates risk without adding state.

**Proposed contract:** retain `age_ms` as the one persisted sum of completed-run attended time and
formally declare that meaning here. Go and TypeScript expose a mechanically named
`completedFounderAttendedMS` accessor/projection; `age_ms` remains the compatibility wire name and
achievement counter. `founder_attended_ms` means the effective sampled value
`age_ms + current_run_partial_attended_ms`, never a second stored Founder field. AC1 and AC5 become
semantic/parity coverage of the existing field rather than a migration that duplicates it.

### A2 — A mid-run snapshot can race Exit and count one run twice

Computing a Company partial before entering `ApplyFounderLogged` is not sufficient. Exit may lock
and advance the Founder between that read and the Founder command's lock; combining the newly
advanced `age_ms` with the stale partial then counts the completed run twice. Conversely, reading
Company after locking Founder would create an ambient cross-stream read inside the Founder
boundary and risks reversing the declared Founder→Company lock order.

**Proposed contract:** resolution freezes the tuple
`{company_stream_id, run_seq, company_revision, company_constants_hash,
completed_attended_ms, current_run_partial_attended_ms, effective_founder_attended_ms}` before the
Founder transition. `ApplyFounderLogged` locks Founder and accepts the tuple only when its persisted
`age_ms == completed_attended_ms` and the request's expected Founder revision still matches. A
concurrent Exit produces the ordinary typed stale-revision result; the service then re-resolves the
tuple against the next run before retry. The transition never reads Company state. A two-order
real-Postgres fixture (Founder command wins / Exit wins) proves the interval is counted exactly
once in both schedules.

### A3 — The partial's offline-aware derivation is not specified

`AttendedMS(company, now)` only subtracts spans already recorded in the Company state. When `now`
is later than `evaluated_through`, the unresolved gap must first be classified with the pinned
prestige `catchup_ceiling_ms`; otherwise a week-long dormant gap becomes attended merely because a
pet command was the first request back. FA3 currently assumes this classification has happened.

**Proposed contract:** the resolver loads the authenticated session's exact Company stream and
pinned catalog bundle, clones the Company state, canonicalizes database/server time, applies
`RecordOfflineSpan(clone, evaluated_through, now, pinned_catchup_ceiling_ms)`, and then computes
`AttendedMS(clone, now)`. It does not persist the clone. No active Company run resolves a partial of
zero; missing, ambiguous, stale-run, or unresolved-catalog contexts fail closed with typed internal
errors. Fixtures cover a sub-ceiling reconnect, a 25-hour return, an Exit race, and replay under a
new deploy-current epoch while the run remains pinned to its old epoch.

### A4 — Founder replay/genesis is a prerequisite, not an implemented dependency

The dependency line says ApplyFounderLogged is implemented, but Pet C9 correctly records the
remaining work: Founder genesis does not yet exist and Exit's Founder mutation is not yet a
`founder_log` entry. Attendance-consuming Founder commands cannot claim career-long replay until
that history is complete; revision pruning would otherwise remove the only starting state.

**Proposed contract:** this RFC depends on the Pet-C9 Founder replay slice landing first, or owns it
as its first implementation batch. That slice defines one immutable Founder genesis row created in
the same transaction as the first Founder-log activation, plus an Exit Founder-log command whose
resolved input contains the immutable Company `run_log` identity and the exact Founder-relevant
Exit facts. Exit appends the Founder row in the existing multi-stream transaction. Founder replay
starts from genesis and compares state, receipt, and ordered Founder events through both ordinary
Founder commands and Exits; no Company stream is scanned during replay.

### A5 — Bounds, monotonicity, and consumer cursor semantics are implicit

The sum is an exact integer but the draft supplies no overflow rule, no definition for the partial
across a run boundary, and no rule preventing consumers from persisting a cursor ahead of the
effective clock. Those omissions would let Go and TypeScript disagree at the numeric boundary or
let a failed transition burn future attended time.

**Proposed contract:** every component is an integer in `[0, MaxExactInteger]`; addition rejects
overflow before mutation. The partial is the total attended-so-far for the named current run, not
a delta since the previous Founder command. The effective value is monotonic across commands and
across Exit (`old age + final partial == new age + zero next-run partial`). Consumer cursors advance
only in the same transaction as their effect, never during resolution, and must satisfy
`0 <= cursor <= effective_founder_attended_ms`. Shared Go/TS vectors cover zero, boundary carry,
run transition, stale resolution, and overflow rejection.

## Changelog

- 2026-08-05: created (draft) — the shared Founder attendance clock; unblocks Pet C10 + Minigame
  C26; defined as summed run-attended flushed at Exit + a frozen mid-run partial (replay-safe).
- 2026-08-05: Codex acceptance review found five blockers: reuse the existing `age_ms` authority,
  make the frozen partial race-safe against Exit, specify offline-aware resolution, land Founder
  genesis/Exit logging first, and close exact bounds/cursor semantics. Status remains draft.
