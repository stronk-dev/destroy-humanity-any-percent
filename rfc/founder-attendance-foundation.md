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

## Changelog

- 2026-08-05: created (draft) — the shared Founder attendance clock; unblocks Pet C10 + Minigame
  C26; defined as summed run-attended flushed at Exit + a frozen mid-run partial (replay-safe).
