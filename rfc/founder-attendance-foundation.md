# RFC: Founder Attendance Foundation (the shared cross-run clock)

- **Status:** accepted (A1-A5 ruled; implementing)
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-05
- **Design refs:** `design/02 §2` (Founder scope, persists), `design/02 §9` (attended vs offline — the banked-time model)
- **Depends on:** Save + Run Genesis + ApplyFounderLogged (implemented), plus the Founder
  genesis/Exit-log slice owned as this RFC's first implementation batch
- **Unblocks:** Pet Care C10 (pet decay clock), Minigame Platform C26 (cross-run faucet quota) — TWO consumers, which is why it's its own primitive, not owned by either.
- **Planning:** `planning/founder-attendance/` (once implementing)
- **Created because:** Pet Care C10 + Minigame C26 both showed a server timestamp is not an attendance authority (elapsed wall time counts offline gaps; the command instant advances by zero). A shared clock two systems need deserves deliberate design, not a ruling squeezed into either bounce.

## Summary

The Company stream derives attended time from the online-session gap between accrual evaluations
(`catchup_ceiling_ms`). Completed-run attendance is already persisted in Founder `age_ms`; Founder
commands are sporadic, so they additionally need one race-safe frozen view of the current run's
partial attendance. This RFC formalizes that existing authority plus the sampled partial once,
replay-safe, for every Founder-scoped consumer.

## Specification

### FA1 — The definition (`age_ms` authority + one frozen current-run partial)

**A founder is "attended" exactly when one of their company runs is attended.** `founder_attended_ms`
is a computed sample: persisted Founder `age_ms` (the sum of every completed run's Attended Time)
PLUS the frozen current run's attended-so-far. It is never a second stored Founder field. This is
the honest definition: pets decay and faucets tick on the founder's real presence, which is
exactly their runs' attended time — not care-command frequency and not wall clock.

### FA2 — Replay independence (the C1 requirement, preserved)

The Founder stream never READS live Company state (C1's law). Instead:
- At **Exit**, the multi-stream transaction (which already touches both streams) advances
  `age_ms += run_ended.attended_ms` — the existing Founder-stream write of a Company-stream fact,
  at the one boundary that legitimately spans both. So the completed-run authority is durably
  correct as of the last Exit.
- **Between Exits** (during a run), a Founder command that needs current attendance reads the
  locked `age_ms` value PLUS `current_run_partial_attended_ms` from the A2 tuple. The tuple is
  passed as a resolved input into `ApplyFounderLogged` (frozen in `founder_log.replay_inputs`),
  computed server-side before the Founder transition, and accepted only while its completed base
  still matches. The transition never reaches into Company state.

Founder replay is therefore independent: it consumes `age_ms` (Founder state) plus the frozen
partial (Founder replay inputs) — never a live Company read.

### FA3 — Offline correctness

Because the source is run Attended Time (which already subtracts offline spans, P6), offline is
correct by construction: a founder away for a week accrues zero attended time (their runs weren't
attended), so pets don't rot and faucet quotas don't tick. No separate offline computation.

### FA4 — The consumer contract

Consumers (pet decay, faucet quota) read the computed `founder_attended_ms` sample and compute
deltas on the fixed attended grid (the provision-grid partition-invariance). Adding a consumer
adds no new attendance source — one clock, many readers.

## Acceptance criteria

1. Existing `age_ms` advances by exactly `run_ended.attended_ms` at each Exit; a founder with N
   completed runs has the sum, and no second stored attendance field exists.
2. Mid-run Founder command: the complete A2 tuple is frozen in Founder replay inputs; replay
   reproduces the same effective attendance without a live Company read, and both Exit race orders
   count the interval exactly once.
3. Offline: a founder idle across a week accrues zero attended (their runs weren't attended);
   no pet decay, no faucet tick over that span.
4. Both consumers (a pet-decay fixture, a faucet-quota fixture) read one clock; adding the second
   reader introduces no second source.
5. No Founder cursor migration is added. Shared corpus coverage proves the existing `age_ms` Exit
   advance, sampled partial, bounds, and stale-resolution behavior byte-identical in Go/TypeScript.

## Owner rulings on A1-A5 (2026-08-05)

- **A1 - accepted, no new field: `age_ms` IS the authority.** I proposed a duplicate cursor over
  a sum `age_ms` already holds (finishExitResolved adds AttendedMS to founder.AgeMS; achievements
  already consume it). **Ruling: `founder_attended_ms` is NOT a stored field** - it is the effective
  sampled value `age_ms + current_run_partial_attended_ms`, computed, never persisted twice. Go/TS
  expose a `completedFounderAttendedMS` accessor; `age_ms` stays the wire name + achievement
  counter. AC1/AC5 become semantic/parity coverage of the existing field, not a migration. (FA1's
  'cursor' language is superseded by this ruling.)
- **A2 - accepted, the race is real:** a partial read before the Founder lock can race an Exit that
  advances age_ms in between, double-counting the completed run. **Ruling: resolution freezes the
  tuple {company_stream_id, run_seq, company_revision, company_constants_hash, completed_attended_ms,
  current_run_partial_attended_ms, effective_founder_attended_ms} BEFORE the Founder transition;
  ApplyFounderLogged accepts it only when persisted `age_ms == completed_attended_ms` AND the
  expected Founder revision still matches** - a concurrent Exit yields the ordinary typed
  stale-revision result, the service re-resolves against the next run and retries. The transition
  never reads Company state. A two-order real-Postgres fixture (command-wins / Exit-wins) proves
  the interval counts exactly once in both schedules.
- **A3 - accepted, offline-aware resolution:** the resolver clones the Company state, canonicalizes
  time, applies `RecordOfflineSpan(clone, evaluated_through, now, pinned_catchup_ceiling_ms)`, THEN
  computes `AttendedMS(clone, now)` - so an unresolved dormant gap is classified offline BEFORE it
  can become attended (the week-dormant-then-pet-command case). Clone never persisted; missing/
  ambiguous/stale/unresolved-catalog contexts fail closed typed. Fixtures: sub-ceiling reconnect,
  25h return, Exit race, replay under a new deploy-current epoch while the run stays pinned.
- **A4 - accepted, the ordering dependency: this RFC's FIRST implementation batch is the Pet-C9
  Founder replay/genesis slice** (Founder genesis row created with the first Founder-log activation
  + Exit-as-a-founder_log-command carrying the immutable Company run_log identity + Exit facts).
  Attendance-consuming commands cannot claim career-long replay until that history is complete;
  this RFC owns landing it first. Founder replay starts from genesis, scans no Company stream.
- **A5 - accepted:** every component is int in `[0, MaxExactInteger]`, addition rejects overflow
  before mutation; the partial is total-attended-so-far for the named run (NOT a delta);
  monotonicity holds across Exit (`old_age + final_partial == new_age + 0`); consumer cursors
  advance only in the same transaction as their effect, bounded `0 <= cursor <= effective`. Shared
  Go/TS vectors: zero, boundary carry, run transition, stale resolution, overflow rejection.

The through-line: I invented a duplicate of `age_ms` (A1) and under-specified the Exit race (A2) -
both corrected; the clock is now the EXISTING field sampled race-safely, not a new stored cursor.

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

## Implementation blockers after A4 genesis landing (Codex, 2026-08-05)

Migration 00056 and `6e51318` complete the immutable Founder-genesis half of A4. Implementing the
Exit row and verifier exposes a narrower wire decision that the old Run Genesis RFC deliberately
left out: the existing Exit receipt contains a next-Company snapshot, so a Founder-only replay
cannot reproduce it without violating A4's no-Company-scan rule. Embedding that receipt as a
resolved oracle would make comparison meaningless.

### B1 — Exit needs an exact Founder-log receipt and fact union

Ordinary Founder commands store their scope-local deterministic receipt in `founder_log`. Exit's
client receipt is Company-scoped and contains `new_revision` plus the complete next-run Company
snapshot. A Founder verifier cannot generate those bytes from Founder state and Founder facts.
The RFC also says “exact Founder-relevant Exit facts” without naming their closed wire.

**Proposed contract:** the Company `intent_records`/outbox/run-log receipt remains unchanged. The
Exit row in `founder_log` stores a separate audit-only Founder receipt:
`{intent_id,outcome,founder_revision,result_constants_hash}` when applied, or
`{intent_id,outcome,current_revision,rejection:{category,detail}}` when rejected. Its resolved arm
is `exit.v1` with exact keys: immutable Company command identity; result constants hash; reputation,
Route Knowledge, attended-time, and achievement-score deltas; sorted added network slots, ledger
facts, and lifetime-achievement IDs; the exact appended Exit record; and resulting Founder wire
version. Rejected rows require the neutral/empty fact shape. The Founder transition derives its
receipt and `founder_advanced` event from that arm and byte-compares both; the client never receives
a duplicate Founder audit receipt for the same Exit.

### B2 — “Company run-log identity” must be relational, not an unchecked JSON claim

Putting `{company_stream_id,run_seq,run_log_seq}` only inside `replay_inputs` does not prove the
referenced command exists. Both rows are authored in the same Exit transaction, so the database can
enforce the relationship without introducing a read during replay.

**Proposed contract:** a forward migration adds nullable
`(source_company_stream_id,source_run_seq,source_run_log_seq)` columns to `founder_log`, a unique key
on the equivalent `run_log` coordinates, and a deferrable composite foreign key. The three columns
are all-null for ordinary Founder commands and all-present only for the `exit.v1` resolved arm; a
trigger enforces the biconditional. Exit inserts both linked rows in its existing transaction, so a
fault at either insertion commits neither. Founder replay consumes the already-joined identity from
the Founder row but never scans Company state or events.

### B3 — Hash transition and verifier surface need one declared meaning

An ordinary Founder row starts and ends under one constants hash. Exit starts under the old Founder
hash and may activate a new epoch/save version. `founder_log.constants_hash` has one column, and no
current Go/TypeScript Founder verifier or row-loading contract says whether that column identifies
the input or output bundle.

**Proposed contract:** `founder_log.constants_hash` always identifies the input/pre-command bundle;
`exit.v1.result_constants_hash` identifies the applied output bundle (equal to input on rejection).
`applied_revision` identifies the resulting Founder revision. Founder replay restores genesis under
its hash, requires each row's input hash to equal the replay state's current hash, applies the closed
`buy_route_hint.v1|exit.v1` union through the same Go transition used live, then advances to the
result hash/version only on an applied Exit. It byte-compares state, scope-local audit receipt, and
ordered Founder events in Go and TypeScript. Unknown arms, hash gaps, sequence gaps, or unavailable
artifacts fail closed with typed verdicts; no deploy-current catalog is read.

## Owner rulings on B1-B3 (2026-08-05)

- **B1 - accepted:** the Company intent_records/outbox/run-log receipt is UNCHANGED (Exit's
  Company-scoped receipt stays the Company boundary's). The Exit row in `founder_log` stores a
  SEPARATE audit-only Founder receipt over Founder facts only - a closed `exit.v1` union
  (founder-relevant Exit facts: attended_ms advanced, age_ms before/after, any Founder-scope state
  the Exit touched) that a Founder verifier CAN regenerate from Founder state alone. The two
  receipts are distinct by design - Company replay verifies the Company receipt, Founder replay
  verifies the Founder audit receipt; neither pretends to generate the other's bytes.
- **B2 - accepted, relational not JSON-claim:** a forward migration adds nullable
  `(source_company_stream_id, source_run_seq, source_run_log_seq)` columns to `founder_log`, a
  unique key on the equivalent `run_log` coordinates, and a DEFERRABLE composite FK - so the
  referenced Company command provably exists, enforced by the database in the SAME Exit
  transaction that authors both rows (no cross-stream read during replay - the DB proves the link
  at write time). The columns are non-null exactly for Exit rows, null for ordinary Founder
  commands (a CHECK ties presence to the exit kind).
- **B3 - accepted, one declared meaning:** `founder_log.constants_hash` ALWAYS identifies the
  INPUT/pre-command bundle (uniform with ordinary rows); `exit.v1.result_constants_hash` identifies
  the APPLIED OUTPUT bundle (equal to input on rejection, differs when Exit activates a new epoch/
  save version - the activation-boundary law's crossing). The Founder verifier reads input-hash
  from the column, output-hash from the exit fact; no column overload.

These are the Exit-as-Founder-log wire contracts; with them the A4 Founder-replay slice is fully
specified. Exit-as-founder_log-command implementation proceeds after this ruling.

## Changelog

- 2026-08-05: created (draft) — the shared Founder attendance clock; unblocks Pet C10 + Minigame
  C26; defined as summed run-attended flushed at Exit + a frozen mid-run partial (replay-safe).
- 2026-08-05: B1-B3 ruled — separate audit-only Founder Exit receipt (exit.v1), relational deferrable-FK Company-log link, input-hash-in-column/output-hash-in-fact. Exit-logging unblocked.
- 2026-08-05: Codex acceptance review found five blockers: reuse the existing `age_ms` authority,
  make the frozen partial race-safe against Exit, specify offline-aware resolution, land Founder
  genesis/Exit logging first, and close exact bounds/cursor semantics. Status remains draft.
- 2026-08-05: immutable Founder genesis landed. B1-B3 bounce the exact scope-local Exit receipt,
  relational Company-log link, and pre/post-hash verifier semantics required to reverse the older
  cross-run-Founder non-goal without embedding Company state or a self-confirming oracle.
