# RFC: Fiscal Quarters Foundation (the wall-clock meta-currency)

- **Status:** accepted — F1-F15 ruled (exact artifact rows, runidentity/Substream seed, the compound
  Founder-command wrapper, `run_frozen_contributions` storage with NO Company bump, v19 save shape,
  byte-enumerated wire, Prestige-offer scope split); implementing. Founder v19.
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-05
- **Design refs:** `design/02 §5` (Fiscal Quarters — the rate-immune clock: the Earnings Call ripens on
  real time, yields Investor Confidence, spent on building levels / minigame unlocks / special
  quarter types; the hoard bonus; immune to production rate).
- **Depends on:** Save + Run Genesis (implemented), Founder Attendance Foundation (implemented — the
  Founder stream + `ApplyFounderLogged` this rides), the Founder save-version chain (minigames v17 /
  pets v18 — this proposes v19). **NOT the Company `ApplyLogged` path** — this is a Founder-scoped
  mechanic on its own wall-clock.
- **Owner ruling honored:** breadth-first — the ripen/harvest/spend MECHANICS and wire, not the
  special-quarter catalog or the balance numbers (those are content/data).
- **Planning:** `planning/fiscal-quarters-foundation/` (once implementing)

## Summary

The third-clock meta-currency, made deterministic and server-authoritative: a **real-wall-clock**
currency (`fiscal_credit`, flavor "Investor Confidence") that matures on a fixed real-time
period (`fiscal_period`, flavor "Earnings Call"), is **immune to production rate and to attended
time**, and is spent on permanent Founder-side investments. It is the long-tail pacing device — a
clock nothing in the game can accelerate — so it lives on the **Founder scope** (persists across
Exit) and mutates through `ApplyFounderLogged`, activating New-Founder-forward at **Founder save
version 19** under a pinned `fiscal` artifact. This RFC specifies the maturation clock, the
harvest/spend intents, the closed spend-target grammar, the hoard bonus, and the byte-parity wire —
not the special-period catalog or any rate/threshold number.

## Motivation

Every other clock in the game is either the Company production clock (closed-form over real elapsed,
but gated by production/offline rules) or the Founder attended clock (advances only while present).
The design calls for a THIRD clock that is immune to both — a currency you cannot farm faster by
playing more, whose only input is real time passing. That is the pacing floor that keeps the very
long tail honest. Getting three things right is the whole job: (1) the clock
must be **closed-form and lazy** — never a server tick loop (binding law); (2) it must be
**deterministic under replay** despite depending on wall time; (3) its persistence scope must be
settled, because it reshapes the Founder save axis.

Out of scope (content / later RFCs): the special-period catalog (Golden Quarter and its siblings),
the exact ripen period / fail chance / hoard cap / building-level costs (all balance data), and any
spend target beyond the two named here.

## Specification

### FQ1 — The maturation clock (`fiscal_period`), closed-form and wall-clock

A single maturing period per founder. State carries `period_opened_wall_ms` — the server wall
timestamp at which the current period began (the previous harvest, or founder genesis). Maturity is
a **pure function of `(now_wall_ms − period_opened_wall_ms)`** against catalog thresholds — there is
NO stored "ripeness" counter and NO tick loop; ripeness is computed on read (lazy), exactly like
production accrual reads `last_evaluated_at`. The clock is **immune to production rate and to
attended time**: neither output nor presence changes it; only wall time does.

Catalog thresholds (STRUCTURE ruled; the millisecond NUMBERS are balance data) define three windows
against elapsed wall time — the harvest-timing model:
- `early_ms` (design: ~20 h) — harvest is *permitted but risky* from here;
- `guaranteed_ms` (design: ~23 h) — harvest always succeeds from here;
- `auto_ms` (design: ~24 h) — if not harvested, the period **auto-reports** (auto-harvest) at the
  next Founder command that reads the clock, crediting deterministically.
Invariant: `0 < early_ms ≤ guaranteed_ms ≤ auto_ms`, loader-validated under the pinned artifact.

### FQ2 — `fiscal_credit`: the Founder-scoped ledger

`fiscal_credit` is an integer ledger on the **Founder save** (persists across Exit). It is a
**spendable** currency (unlike the moral meters) but is minted by exactly one source — the harvest
of a matured `fiscal_period` (the one-mint discipline, cf. Clout). It cannot go negative (typed
rejection on overspend). A matured harvest yields `credit_per_period` (balance data; the mint is
one unit family, no other surface emits `fiscal_credit`).

**SCOPE DECISION (the one that reshapes the save axis — see Open Questions):** `fiscal_credit` and
everything it buys are **Founder-scoped and persist across Exit** — both the credit and the
building levels it buys survive the run reset. Proposed: this adds **Founder save
version 19** (chain: minigames 17 → pets 18 → `fiscal` 19), the `fiscal` artifact biconditional with
floor ≥ 19, activating New-Founder-forward — the same scalar-chain discipline ruled for C36. The
Company axis is untouched.

### FQ3 — Harvest (with the deterministic early-harvest risk)

Harvest is a Founder intent: `harvest_fiscal_period {intent_id, expected_revision}` — C1-envelope,
evented, replay-logged, through `ApplyFounderLogged`. The server stamps the resolved arm with the
authoritative wall timestamp (`now_wall_ms`, exactly as the attended cursor is server-stamped), and
**that timestamp is logged**, so replay reproduces maturity from the logged value, not a live clock.

- Elapsed `< early_ms`: reject `not_eligible` / `period_not_ripe` (F7 — the closed taxonomy; no
  state change).
- `early_ms ≤` elapsed `< guaranteed_ms`: the risky window. The success draw is a **pure function of
  immutable inputs** (F2): shared **SplitMix64**, substream `fiscal.early_harvest.v1`, seeded from
  `founder_id` + the persisted monotonically-increasing `fiscal_period_seq`, compared to
  `early_success_ppm` (balance data, design ~50%). Success credits `credit_per_period` and opens a new
  period (`period_opened_wall_ms = now_wall_ms`); failure ("missed estimates") credits nothing and
  **still opens a new period** (F1-ruled: you spent the quarter, you missed the estimate — the risk is
  real). Every close (success/failure/auto) increments `fiscal_period_seq`, so a retry/re-open can
  never reuse a draw.
- Elapsed `≥ guaranteed_ms`: guaranteed success, credits and re-opens.
- Auto-report (F1): any Founder command computes `n = floor((now − opened)/auto_ms)` BEFORE its own
  action, mints `n·credit_per_period`, and advances `opened += n·auto_ms` (never to `now` — phase
  preserved, so an absent founder never loses matured periods); one aggregated auto-report event
  precedes the command event.

The draw is seeded from `(founder_id, fiscal_period_seq)` via SplitMix64 (F2 — there is NO reusable
Founder RNG stream; the behavior FSM is deterministic), NEVER from `now_wall_ms` — wall time decides
*ripeness*, the seed decides *the coin flip*, both logged/derived deterministically.

### FQ4 — Spend: the closed target grammar

Spend is a Founder intent: `spend_fiscal_credit {intent_id, expected_revision, target}` — C1,
evented, replay-logged, `ApplyFounderLogged`. **There is NO client `amount` (F3 — cost is
server-derived, so a client amount is an authority hole):** the server resolves the cost from the
pinned artifact + current Founder state and debits `fiscal_credit` integer-exact (cannot go negative;
overspend rejects `unaffordable`/`fiscal_credit`). `target` is a **closed union** (STRUCTURE ruled;
catalogs/costs are data):
- `generator_level {generator_id, levels}` — the "building levels" sink. A **Founder-permanent**
  per-generator level bonus (design: +1% that generator per level). Cost for level `L` buying `k` is
  the triangular `k·(2L+k+1)/2` (F3, wide intermediates, safe-integer). **Frozen into each new
  Company's genesis for the whole run (F4 — freeze-at-genesis; mid-run changes affect the NEXT run
  only), materialized as genesis contribution rows so Company replay needs no ambient Founder read.**
  Because
  generators are Company-scoped and reset, the LEVELS are a Founder-side modifier keyed by
  `generator_id`, re-applied each run — persistence lives on the Founder, effect lands on the run.
- `unlock {unlock_id}` — the staggered minigame/system unlock sink (a Founder-permanent flag).
Special-period types (Golden Quarter and its siblings) are a **declared successor** target family,
not in Phase A — named so the union is known to be open, their catalog deferred as content.

### FQ5 — The hoard bonus

A passive production modifier derived from the **unspent** `fiscal_credit` balance:
`+hoard_ppm_per_credit` per unspent unit, hardcapped at `hoard_cap_credits` unspent (design: +1% per
unspent, cap 100). It is DERIVED (a read over the Founder ledger) and **frozen into Company genesis
for the run (F4 — like the levels; mid-run hoard changes affect the NEXT run)**, never a stored
second value — no duplicate authority, no ambient Founder read in production. The hoard-vs-spend tension is the meta decision:
banked credit passively boosts production up to the cap, or is spent on permanent levels. Numbers are
balance data; the *derivation* (linear in unspent, hardcapped) is ruled.

### FQ6 — Wire, receipt, events (byte-parity)

Founder intents: the two above, exact fields as written (spend carries NO `amount`, F3). Resolved
arms: `{kind:"harvest_fiscal_period", now_wall_ms, period_opened_wall_ms_before, periods_swept,
seq_before, draw_ppm, outcome}` and `{kind:"spend_fiscal_credit", target, resolved_cost}` — the
server owns the resolved cost/seq/draw; the pinned artifact owns thresholds/costs. Receipts name the
before/after `fiscal_credit`, the new `period_opened_wall_ms`, and (harvest) the swept-period count +
success/fail outcome; (spend) the target applied + authoritative debit, or a `cap_exceeded`/target-ID
or `unaffordable`/`fiscal_credit` rejection (F7 — the closed taxonomy; NOT the invented `not_ripe`/
`insufficient_credit`). Manual vs automatic harvest is distinguished by a required closed `source`
field. Register `fiscal_period_harvested.v1` and `fiscal_credit_spent.v1`. Go/TypeScript shared
vectors compare state, receipt, and ordered event bytes — including the maturity-window boundaries
(`early_ms−1`, `early_ms`, `guaranteed_ms`, `auto_ms`), MULTIPLE auto periods, the seeded fail path,
and every semantic rejection.

## Deviations from design

- **Mechanical naming:** design's flavor names ("Investor Confidence", "Earnings Call", "Golden
  Quarter") become localization keys over mechanical fields (`fiscal_credit`, `fiscal_period`, a
  special-period-type enum) per the naming law — so the satire retunes without a refactor.
- Nothing else diverges; the ripen/harvest/hoard shapes follow `design/02 §5` exactly, numbers as data.

## Acceptance criteria

1. Maturity is closed-form over `(now_wall_ms − period_opened_wall_ms)` with NO tick loop and NO
   stored ripeness; a fixture advancing wall time across `early/guaranteed/auto` boundaries yields the
   ruled outcomes; the harvest wall timestamp is logged and replay reproduces byte-identically.
2. Determinism: the early-harvest fail draw is a pure SplitMix64 function of `(founder_id,
   fiscal_period_seq)` (F2 — NOT wall time, NOT a nonexistent Founder RNG stream); replay of a logged
   harvest reproduces success/fail exactly; multi-period auto-report at `auto_ms` is deterministic and
   phase-preserving (F1).
3. `fiscal_credit` is Founder-scoped, one-mint (only harvest emits it), integer-exact, never negative;
   spend carries NO client `amount` (server-derived cost, F3); overspend rejects `unaffordable`/
   `fiscal_credit` (F7); visible catalog hardcaps + reason keys (F6).
4. Spend targets apply: `generator_level` (server-derived triangular cost) and hoard bonus are
   **frozen into Company genesis for the run (F4 — freeze-at-genesis), materialized so replay needs no
   ambient Founder read**; `unlock` as a Founder flag; all replay-owned.
5. Founder save v19 activation: `fiscal` artifact biconditional with floor ≥ 19, New-Founder-forward,
   chain requires v18 pinned; Company axis rejects v19; migration + corpus.
6. Byte-parity Go/TS shared vectors on the window boundaries, the seeded fail path, spend, and
   overspend.

## Open questions (resolve before `accepted`)

- **Scope: RESOLVED 2026-08-05 → Founder-scoped, save version 19.** Owner ruling: a wall-clock
  meta-currency that persists across the prestige reset is what drives the months-long tail, while a
  long-tail mechanic that does NOT persist across prestige feels wasted the moment the player is on
  an Exit cadence (a documented failure mode this scope avoids). The chain is minigames 17 → pets 18 → `fiscal` 19; the
  `fiscal` artifact is biconditional with floor ≥ 19, New-Founder-forward, requires v18 pinned; the
  Company axis rejects v19. No longer an open question.
- **Early-harvest failure:** does a failed risky harvest open a new period (proposed — real risk) or
  leave the current one to keep maturing? Balance-adjacent but structural (it changes the state
  transition), so ruled here, not deferred to data.
- **Auto-report timing:** proposed lazy (auto-harvest on the next Founder command past `auto_ms`).
  Confirm no background job is expected — a lazy auto-report keeps the no-tick-loop law; a scheduled
  one would need a job contract.

## Acceptance-review blockers (2026-08-05)

Founder v19 is the correct scope/version and the wall-clock calculation can be implemented lazily.
The proposed early-harvest failure opening a new period is also coherent: it makes the risky action
consume the period and gives the failure real cost. The remaining gaps below prevent one canonical
transition and replay result.

### F1 — Lazy auto-report does not define multi-period catch-up or command ordering

If a founder returns after several `auto_ms` intervals, crediting one period and setting opened time
to `now` loses matured periods, contradicting "an absent founder never loses matured periods."
Crediting all periods requires exact phase-preserving arithmetic. It is also unstated whether the
auto-report runs before or after the triggering command, which events/receipt contain it, and what
happens when that command rejects (the Founder boundary normally rolls all state back on rejection).

**Proposed contract:** no background job. On every valid Founder command, compute
`n=floor((now-opened)/auto_ms)` before the command-specific action, mint `n*credit_per_period`, and
advance `opened += n*auto_ms` (never set it to `now`, so phase is preserved). Emit one aggregated
auto-report event with count and total credit before the command event. Define whether a later
semantic rejection commits this sweep; recommended: the compound transition is applied when
`n>0`, with the attempted sub-action outcome represented in its receipt, because mutating under an
ordinary rejected outcome violates the existing log contract. The special harvest intent at/after
auto must consume the already-auto-reported result rather than double-harvest.

### F2 — The claimed Founder-seeded RNG does not exist

Pet behavior is deterministic and requires `behavior_prng_cursor == 0`; it is not a reusable Founder
random stream. Prestige offer RNG is Company-run-derived and not a career-long Founder source. The
RFC names neither seed derivation, substream label, draw cursor, nor which period identity prevents
retries/re-opens from reusing a draw.

**Proposed contract:** make the risky outcome a pure function of immutable inputs, using the shared
SplitMix64 algorithm and a named `fiscal.early_harvest.v1` substream over a specified seed derived
from founder ID plus a persisted monotonically increasing `fiscal_period_seq`. Increment the period
sequence whenever any manual/automatic period closes; log the before sequence and resolved draw
PPM. Alternatively store a PRNG cursor, but the exact seed/cursor wire and Go/TS vectors are
mandatory. `now_wall_ms` must never supply randomness.

### F3 — Spend accepts a client amount although costs are server-derived

For `generator_level`, cost depends on current level and the requested number of levels; for
`unlock`, cost belongs to the pinned catalog. A client-supplied `amount` is either redundant or an
authority hole. No exact triangular multi-level equation, overflow order, already-unlocked result,
or target eligibility is stated.

**Proposed contract:** remove `amount` from the intent. The target carries
`generator_level {generator_id, levels}` or `unlock {unlock_id}`; the server resolves the cost from
the pinned artifact and current Founder state. For level `L` buying `k`, define
`cost=sum(L+1..L+k)=k*(2L+k+1)/2` with wide integer intermediates and safe-integer result. Unknown
targets reject `unknown_id`; duplicate unlock and zero levels use existing closed rejection
category/details. Receipt reports the authoritative debit.

### F4 — Founder-derived production effects have no cross-stream contract

Hoard credit and generator levels live on the Founder stream but change Company production. "At run
assembly" could mean frozen for the run, while harvesting/spending mid-run suggests immediate
changes. The current Company replay boundary only reproduces Founder effects that are explicitly
carried/resolved; reading live Founder state inside production would violate `ApplyLogged`.

**Owner ruling required:** choose update semantics. Recommended: freeze both level and hoard
contributions into each new Company's genesis for the entire run; mid-run Founder changes affect the
next run only. Declare the exact multiplier slot/source IDs and materialized Company fields or
genesis contribution rows so replay needs no ambient Founder read. If effects are immediate, every
Company command must instead resolve and log a Founder carry with revision/hash and define the
cross-stream race—substantially larger scope.

### F5 — The pinned `fiscal` artifact is not an exact grammar

Thresholds alone are insufficient. The loader also needs credit mint/cap, early-success ppm, hoard
policy, generator-level policy/eligible IDs, unlock rows/costs, hardcap reason keys, and exact
cross-artifact validation against economy generator IDs. No Go/TS schema or artifact-bundle
registration is enumerated.

**Proposed contract:** enumerate one exact schema-v1 root containing clock policy, credit policy,
hoard policy, generator-level rows, and unlock rows. Require exact keys, sorted unique mechanical
IDs, safe-integer bounds, `early<=guaranteed<=auto`, ppm bounds, and all referenced generator IDs in
the same pinned economy artifact. Add `fiscal` to the epoch seed/artifact authority; presence is
biconditional with Founder floor >=19 and requires minigames+pets as already ruled.

### F6 — No declared hardcaps protect the integer state

`fiscal_credit`, period count, generator levels, and auto-catch-up multiplication can reach the safe
integer ceiling. Treating `MaxExactInteger` as a surprise cap violates the announced-hardcap law;
wrapping or rejecting an automatic mint could brick every later Founder command after a long gap.

**Owner ruling required:** declare visible catalog hardcaps + reason keys for credit and each level
row, and an accrual-only saturation policy for automatic/manual mint. Wide intermediates must prove
`period_count*credit_per_period` and triangular cost cannot overflow. Decide whether period sequence
saturates (recommended: no; reject artifact horizons that can exhaust it within the timestamp
domain) and include boundary vectors.

### F7 — Wire/rejection/event grammars are not closed

`not_ripe` and `insufficient_credit` are written as rejection categories, but the implemented
taxonomy is closed and neither category exists. FQ6 does not enumerate exact receipt/event keys,
auto-report payload, period count/sequence, or event ordering. `saturated` is named without saying
whether it is an outcome, receipt field, or event.

**Proposed contract:** reuse existing categories (`not_eligible` with `period_not_ripe` /
`already_unlocked`; `unaffordable` with `fiscal_credit`; `cap_exceeded` with the target ID), then
enumerate exact canonical intent, resolved, receipt, and event objects. Add distinct manual and
automatic harvest event semantics or one kind with a required closed `source` field. Shared vectors
must compare raw state/receipt/ordered-event bytes, including multiple auto periods and every
semantic rejection.

### F8 — Genesis/activation and wall-clock rollback are unstated

Version 19 can arise at a current-epoch New Founder or at an Exit that activates the next pinned
bundle. The RFC does not say which server timestamp initializes `period_opened_wall_ms`, nor how an
older/equal `now_wall_ms` is handled. A wall-clock regression could make the period immature for an
arbitrary duration or permit a timestamp before Founder genesis.

**Proposed contract:** initialize period opened time from the same server-authored timestamp that
commits v19 genesis/Exit activation, recorded in the Founder log where applicable. Every command
requires `now_wall_ms >= period_opened_wall_ms`; equality is a valid zero-elapsed read, regression
fails as `internal_invariant` without state change. Enforce the timestamp safe-integer domain in Go,
TS, save codec, and artifact-boundary fixtures.

## Changelog

- 2026-08-05: created (draft) — Wave-A foundation from the coverage-map gap backlog; the wall-clock
  meta-currency, `design/02 §5`.
- 2026-08-05: Codex acceptance review — Founder v19 and lazy closed-form direction confirmed;
  implementation blocked on F1-F8 (catch-up/order, RNG, spend authority, cross-stream effects,
  artifact grammar, caps, wire taxonomy, and activation/rollback).
- 2026-08-06: non-normative reference cleanup for publication; no spec change.

## Owner rulings on F1-F8 (2026-08-06)

All accepted (Codex's proposed contracts are sound). Full contracts + the two owner-calls (F4/F6).

- **F1 — accepted (multi-period catch-up + ordering).** No background job. On every well-formed
  Founder command, BEFORE the command's own action, compute `n = floor((now_wall − opened) / auto_ms)`,
  mint `n * credit_per_period`, and advance `opened += n * auto_ms` (NEVER set to `now` — phase is
  preserved so no matured period is lost). Emit one aggregated auto-report event `{count, total_credit}`
  ordered before the command's event. The fiscal intents are **compound transitions**: the time-based
  auto-sweep is a valid sub-transition that commits when `n > 0`, and the triggering sub-action's
  outcome (including a semantic rejection) is recorded in the receipt — the sweep is NOT "mutation
  under rejection" (it is the deterministic passage-of-time mint, independent of the sub-action). A
  manual `harvest` at/after `auto_ms` CONSUMES the auto-reported result (no double-harvest).
- **F2 — accepted; fixes the nonexistent-RNG error in my draft.** There is NO reusable Founder RNG
  stream (the behavior FSM is deterministic, `behavior_prng_cursor==0`). The early-harvest risk is a
  PURE FUNCTION of immutable inputs via the shared **SplitMix64**, named substream
  `fiscal.early_harvest.v1`, seeded from `founder_id` + a persisted monotonically-increasing
  `fiscal_period_seq`. `fiscal_period_seq` increments whenever ANY period closes (manual or auto), so
  a retry/re-open can never reuse a draw. Log the before-seq + resolved draw ppm. `now_wall_ms` never
  supplies randomness. Exact seed/draw wire + Go/TS vectors mandatory. (FQ3 body reconciled.)
- **F3 — accepted; `amount` removed from the intent.** Cost is server-derived, so a client `amount`
  is an authority hole. The target carries `generator_level {generator_id, levels}` or
  `unlock {unlock_id}`; the server resolves cost from the pinned artifact + current Founder state. For
  level `L` buying `k`: `cost = Σ(L+1 .. L+k) = k*(2L+k+1)/2` with wide intermediates + safe-integer
  result. Unknown target → `unknown_id`; duplicate unlock / zero levels → existing closed rejection.
  Receipt reports the authoritative debit. (FQ4 body reconciled.)
- **F4 — RULED: FREEZE-AT-GENESIS (the cross-stream contract).** Founder-derived production effects
  (generator levels + hoard bonus) are **frozen into each new Company's genesis for the entire run**,
  materialized as genesis contribution rows / Company fields with declared multiplier slot + source
  IDs — so Company replay reproduces them with NO ambient Founder read (the `ApplyLogged` boundary is
  never violated). **Mid-run harvest/spend affects the NEXT run only** ("the founder lands softly";
  your permanent investments apply to the run you START). The immediate-effect alternative (log a
  Founder carry on every Company command) is rejected as far larger scope. FQ4/FQ5 "at run assembly"
  is now definitively "frozen at Company genesis."
- **F5 — accepted (the `fiscal` artifact grammar).** Enumerate schema-v1 root `{clock_policy,
  credit_policy, hoard_policy, generator_level_rows, unlock_rows}` with exact keys, sorted-unique
  mechanical IDs, safe-integer bounds, `early_ms ≤ guaranteed_ms ≤ auto_ms`, ppm bounds, and a
  cross-artifact check that every referenced generator ID resolves in the SAME pinned economy
  artifact. Register `fiscal` in the epoch seed/artifact authority; presence biconditional with
  Founder floor ≥ 19; requires minigames(17)+pets(18) pinned (the chain).
- **F6 — RULED (visible hardcaps, the announced-cap law).** Declare VISIBLE catalog hardcaps + reason
  keys for `fiscal_credit` and each generator-level row; auto/manual mint uses accrual-only saturation
  at those caps (never a surprise `MaxExactInteger` cap). Wide intermediates must prove
  `period_count * credit_per_period` and the triangular cost cannot overflow. `fiscal_period_seq` does
  **NOT** saturate — instead the loader REJECTS any artifact whose `auto_ms` horizon could exhaust the
  seq within the timestamp safe-integer domain (a saturating seq would break the F2 draw-uniqueness).
  Boundary vectors included.
- **F7 — accepted; reuses the closed taxonomy (my invented categories removed).** `not_ripe` and
  `insufficient_credit` do NOT exist in the closed rejection taxonomy. Reuse: `not_eligible` with
  detail `period_not_ripe` / `already_unlocked`; `unaffordable` with `fiscal_credit`; `cap_exceeded`
  with the target ID. Enumerate exact canonical intent/resolved/receipt/event objects; manual vs
  automatic harvest either distinct event kinds or one kind with a required closed `source` field.
  Shared vectors compare raw state/receipt/ordered-event bytes incl. multiple auto periods + every
  rejection. (FQ6 body reconciled.)
- **F8 — accepted (genesis/activation + wall-clock rollback).** `period_opened_wall_ms` initializes
  from the SAME server-authored timestamp that commits v19 genesis/Exit activation (recorded in the
  Founder log where applicable). Every command requires `now_wall_ms ≥ period_opened_wall_ms`;
  equality = a valid zero-elapsed read; a REGRESSION fails as `internal_invariant` with NO state
  change (a rolled-back clock can never make the period immature or pre-genesis). Enforce the
  timestamp safe-integer domain in Go, TS, the save codec, and artifact-boundary fixtures.

F1-F8 fully ruled — catch-up, RNG, spend authority, freeze-at-genesis, artifact grammar, caps, wire,
and activation are executable. Numbers (thresholds/rates/caps/costs) and the special-period catalog
remain data / successors.

## Implementation acceptance recheck (2026-08-06) — blocked F9-F15

F1-F8 choose the gameplay behavior, but implementation against the shipped Founder replay and
Company contribution boundaries exposes seven residual contracts. No Fiscal mechanic code started:
choosing these shapes in Go would create new persistence, replay, and wire authority by accident.

### F9 — The schema-v1 artifact still has no exact nested rows or multiplier identity

F5 names five root families but does not enumerate their keys. In particular, neither the hoard nor
generator-level rows name the multiplier slot/source/target that F4 requires, and the economy loader
cannot prove that the generated contributions are declared. `unlock_rows` likewise has no cost key.

**Proposed contract:** exact root `{schema_version:1, clock_policy, credit_policy, hoard_policy,
generator_level_rows, unlock_rows}`. Clock is `{early_ms, guaranteed_ms, auto_ms,
early_success_ppm}`. Credit is `{credit_per_period, hardcap, hardcap_reason_key}`. Hoard is
`{ppm_per_credit, cap_credits, slot, source_id, target}`. Generator rows are raw-byte sorted exact
`{generator_id, ppm_per_level, level_hardcap, hardcap_reason_key, slot, source_id}`. Unlock rows are
raw-byte sorted exact `{unlock_id, cost}`. Rows and IDs are unique; safe-integer/ppm/order rules from
F5-F6 apply; `credit_per_period <= hardcap`; hoard target is `all`; generator targets are their own
IDs. Slot/source/target must exactly match a multiplier-source declaration in the same economy
artifact. Factors are Decimal-exact `1 + min(credit,cap_credits)*ppm_per_credit/1e6` and
`1 + level*ppm_per_level/1e6`, canonical-quantized once. Go/TS share one mutation corpus.

### F10 — `(founder_id, fiscal_period_seq)` is not an exact seed derivation

The repository has two relevant primitives: `runidentity.Seed` (SHA-256 first-u64 xor sequence) and
`determinism.Substream` (FNV-1a label xor SplitMix64 seed). F2 names the inputs and label but does not
choose the framing/hash/endian rule, so Go and TypeScript can legally draw different PPM values.

**Proposed contract:** `base = runidentity.Seed(founder_id, fiscal_period_seq)` exactly (UTF-8
founder ID, SHA-256, first eight bytes big-endian, xor uint64 sequence); draw =
`determinism.Substream(base, "fiscal.early_harvest.v1").Bound(1_000_000)`. The TypeScript port uses
the byte-identical SHA-256/FNV-1a/SplitMix64/rejection-sampling sequence. Shared vectors include seq
0, seq 1, a high safe-integer seq, and both sides of an `early_success_ppm` comparison.

### F11 — A rejected sub-action cannot currently commit the ruled auto-sweep

`save.ApplyFounderLogged` intentionally persists no state/events/revision for `IntentRejected`, and
`ApplyFounderLogged` restores its snapshot for every non-applied outcome. F1 instead requires a
time-based sweep to commit even when the triggering semantic action rejects. It also applies to
every Founder command (care, route hint, minigame resolution, Exit), whose exact receipts do not
have a place for sweep evidence. The current boundary cannot represent that result.

**Owner ruling required:** define the compound envelope. Recommended: under active Fiscal v19 every
Founder command returns one exact wrapper `{intent_id, outcome, founder_revision, fiscal_sweep,
action}`. `fiscal_sweep` is null or `{periods, credit_before, credited, credit_after,
opened_before_ms, opened_after_ms, seq_before, seq_after, saturated, hardcap_reason_key}`; `action`
contains the command's existing canonical receipt. If the sweep is non-null, the outer outcome is
`applied` and the revision commits even when `action.outcome` is `rejected`; only the automatic event
is emitted before an applied action's events. With no sweep, existing applied/rejected semantics
remain. A harvest after its pre-action sweep uses a closed action outcome `consumed_by_auto` and
cannot close another period. Alternatively add rejected-with-mutation semantics to the save store,
but that weakens a project-wide invariant and is not recommended.

### F12 — There is no immutable run-owned home for frozen Founder contributions

Company save state has no generic frozen-contribution field. Adding one changes the Company wire
axis, contradicting F4's “Company axis untouched”; reading current Founder levels/credit from the
contribution provider would make effects immediate and ambient, also contradicting F4. The existing
`run_genesis` row stores only Company state bytes/version/hash.

**Proposed contract:** add immutable `run_frozen_contributions` rows keyed by
`(company_stream_id, run_seq, source_id)` with exact `{slot,target,factor}` and a deferrable parent
link to the run pin/genesis. New-Founder creation and Exit write the complete set in the same
transaction as the run's genesis; an immutability trigger rejects update/delete. A composed
`ContributionProvider` reads only those run-keyed rows using the current Company revision identity,
and the existing replay-input contribution array carries the same bytes into `ApplyLogged`. Thus
live play has no Founder read, replay has no DB read, mid-run Fiscal changes cannot alter the run,
and no Company save-version bump is needed. Fault injection covers genesis-without-rows and
rows-without-genesis; retry requires byte-identical rows.

### F13 — Founder v19 initialization is directional but not an exact save shape

F8 names the timestamp but the save has no v19 fields/key completeness contract. The Exit activation
helper currently receives no server timestamp, and New-Founder creation uses a separate path. Maps
for eligible generator levels/unlocks need one declared completeness rule or Go/TS restore can drift.

**Proposed contract:** Founder v19 appends exact keys `fiscal_credit`,
`fiscal_period_opened_wall_ms`, `fiscal_period_seq`, `fiscal_generator_levels`, and
`fiscal_unlocks`. Credit/seq/timestamp are safe integers; levels is a complete object keyed by every
artifact generator row with zero-or-capped values; unlocks is a raw-byte-sorted unique array limited
to artifact unlock IDs. Activation initializes `{credit:0, opened:logged_server_ms, seq:0,
all_levels:0, unlocks:[]}`. Pass the already-recorded Founder command timestamp into Exit activation;
New Founder uses the timestamp that commits its genesis. Pre-v19 state must contain none of these
values; v19 requires all keys. Fiscal presence iff Founder floor >=19, requiring minigames+pets.

### F14 — Fiscal intents/resolved receipts/events are described, not byte-enumerated

FQ6 lists fields but leaves the spend-target JSON union, nullability, the saturated/manual/automatic
payloads, and F11's compound receipt open. The current `IntentRequest.Target` is a string, so it
cannot carry either ruled Fiscal target without a new exact parser. “Every semantic rejection” is
not a fixture until each byte shape is unique.

**Proposed contract:** harvest request has exact keys `{intent_id,kind,expected_revision}`. Spend has
`{intent_id,kind,expected_revision,target}` where target is exactly one of
`{kind:"generator_level",generator_id,levels}` or `{kind:"unlock",unlock_id}`. Resolved harvest is
exact `{kind,now_wall_ms,period_opened_wall_ms_before,periods_swept,seq_before,draw_ppm,outcome}` with
nullable draw only outside the risky manual window and outcome in
`auto_reported|early_succeeded|early_failed|guaranteed|consumed_by_auto|rejected`; resolved spend is
exact `{kind,target,resolved_cost}`. Ratify F11's wrapper first, then enumerate its action receipt and
the two event payloads from these same fields; shared vectors compare raw state/receipt/event order.

### F15 — Prestige requires Quarter harvest as an offer-spawn site, but no owner can mutate it

`rfc/prestige-and-exits.md` normatively says offer spawn checks occur at tier thresholds **and Quarter
harvests**. Fiscal is Founder-only while offer state is Company-owned. A Founder event cannot mutate
Company state, and neither RFC assigns a multi-stream coordinator for this trigger. Shipping Fiscal
silently without it would leave an implemented RFC contradicting the new mechanic.

**Owner ruling required:** recommended scope split: this foundation emits the immutable
`fiscal_period_harvested.v1` fact only and explicitly routes Quarter-triggered offer spawning to a
successor multi-stream/event-consumer RFC; amend Prestige's current wording to name that dependency.
If Quarter harvest must spawn offers now, Fiscal intents must instead become Exit-style multi-stream
transactions with lock order, replay logs, idempotency, and rejected-command behavior specified — a
material scope increase.

Until F9-F15 are ruled and reconciled into FQ1-FQ6, implementation would invent persistence and
wire semantics. The accepted F1-F8 gameplay decisions remain intact.

## Owner rulings on F9-F15 (2026-08-06) — the exact persistence/wire under F1-F8

All accepted (Codex's proposed contracts are executable and sound). Owner-calls: F11, F15.

- **F9 — accepted (exact artifact rows).** Root `{schema_version:1, clock_policy, credit_policy,
  hoard_policy, generator_level_rows, unlock_rows}`. `clock_policy {early_ms, guaranteed_ms, auto_ms,
  early_success_ppm}`; `credit_policy {credit_per_period, hardcap, hardcap_reason_key}` with
  `credit_per_period ≤ hardcap`; `hoard_policy {ppm_per_credit, cap_credits, slot, source_id,
  target=all}`; generator rows raw-byte-sorted exact `{generator_id, ppm_per_level, level_hardcap,
  hardcap_reason_key, slot, source_id}` (target = own ID); unlock rows raw-byte-sorted `{unlock_id,
  cost}`. Slot/source/target must match a multiplier-source declaration in the SAME economy artifact.
  Factors Decimal-exact `1 + min(credit,cap_credits)*ppm_per_credit/1e6` and `1 +
  level*ppm_per_level/1e6`, canonical-quantized once. Shared Go/TS mutation corpus.
- **F10 — accepted (exact seed derivation).** `base = runidentity.Seed(founder_id, fiscal_period_seq)`
  (UTF-8 founder ID, SHA-256, first eight bytes big-endian, xor uint64 seq); draw =
  `determinism.Substream(base, "fiscal.early_harvest.v1").Bound(1_000_000)`; the TS port is the
  byte-identical SHA-256/FNV-1a/SplitMix64/rejection-sampling sequence. Shared vectors: seq 0, seq 1, a
  high safe-int seq, and both sides of the `early_success_ppm` comparison. (This pins F2's "how".)
- **F11 — RULED (the compound envelope).** Under active Fiscal v19 EVERY Founder command returns one
  exact wrapper `{intent_id, outcome, founder_revision, fiscal_sweep, action}`. `fiscal_sweep` is null
  or `{periods, credit_before, credited, credit_after, opened_before_ms, opened_after_ms, seq_before,
  seq_after, saturated, hardcap_reason_key}`. **If the sweep is non-null the OUTER outcome is `applied`
  and the revision commits even when `action.outcome == rejected`** — this preserves "rejected action =
  no action state change" at the ACTION level while the time-based sweep (a valid independent
  sub-transition) commits; it does NOT add rejected-with-mutation semantics to the save store (that
  weakens a project-wide invariant — rejected). The automatic event emits before an applied action's
  events. A harvest after its own pre-action sweep uses action outcome `consumed_by_auto` and closes no
  further period. With no sweep, existing applied/rejected semantics are unchanged.
- **F12 — accepted (frozen contributions; NO Company save bump).** Add immutable
  `run_frozen_contributions` rows keyed `(company_stream_id, run_seq, source_id)` with
  `{slot, target, factor}` + a deferrable link to the run pin/genesis. New-Founder creation and Exit
  write the COMPLETE set in the same transaction as genesis; an immutability trigger rejects
  update/delete. A composed `ContributionProvider` reads only these run-keyed rows at the current
  Company revision identity, and the existing replay-input contribution array carries the same bytes
  into `ApplyLogged` — so live play has NO Founder read, replay has NO DB read, mid-run Fiscal changes
  can't alter the run, and **the Company save axis is untouched (no v-bump)**. Fault injection covers
  genesis-without-rows and rows-without-genesis; retry requires byte-identical rows. (This is F4's
  freeze-at-genesis made concrete.)
- **F13 — accepted (Founder v19 exact save shape).** v19 appends `fiscal_credit`,
  `fiscal_period_opened_wall_ms`, `fiscal_period_seq`, `fiscal_generator_levels` (a COMPLETE object
  keyed by every artifact generator row, zero-or-capped), `fiscal_unlocks` (raw-byte-sorted unique,
  artifact unlock IDs only). Activation initializes `{credit:0, opened:logged_server_ms, seq:0,
  all_levels:0, unlocks:[]}` from the recorded Founder-command timestamp (Exit) / genesis timestamp
  (New Founder). Pre-v19 has NONE of these; v19 requires ALL; presence iff floor ≥ 19 (requires
  minigames+pets).
- **F14 — accepted (byte-enumerated wire).** Harvest request `{intent_id,kind,expected_revision}`;
  spend `{intent_id,kind,expected_revision,target}` with target ∈ `{kind:"generator_level",
  generator_id,levels}` | `{kind:"unlock",unlock_id}`. Resolved harvest exact `{kind,now_wall_ms,
  period_opened_wall_ms_before,periods_swept,seq_before,draw_ppm,outcome}` (draw nullable only outside
  the risky manual window; outcome ∈ `auto_reported|early_succeeded|early_failed|guaranteed|
  consumed_by_auto|rejected`); resolved spend `{kind,target,resolved_cost}`. The action receipt + two
  event payloads enumerate from these fields inside F11's wrapper; shared vectors compare raw
  state/receipt/event-order.
- **F15 — RULED (scope split; Prestige amended).** This foundation emits the immutable
  `fiscal_period_harvested.v1` FACT only; **Quarter-triggered offer-spawning is routed to a successor
  multi-stream/event-consumer RFC** (Fiscal does NOT become Exit-style multi-stream now — that's a
  material scope increase, rejected). `rfc/prestige-and-exits.md`'s "offer spawn at Quarter harvests"
  wording is amended to name that successor dependency (done in the same edit — the reconcile-the-body
  law across RFCs).

F9-F15 fully ruled. FQ3/FQ4/FQ6 are refined (not contradicted) — the exact rows/seed/wire are additive
enumerations of the F1-F8 shapes. Numbers stay data.

## Correction to F11 (2026-08-06) — aligned with Active-Play A11: ROLLBACK, not the wrapper

Active-Play A11 surfaced that the F11 compound-wrapper (sweep commits even when the action rejects,
via a universal `{outcome, fiscal_sweep, action}` envelope on every Founder command) is
over-engineered and inconsistent with the project-wide "rejected intent = no state change" invariant.
**REVISED: the auto-sweep is part of the ordinary applied command and ROLLS BACK with any semantic
rejection** — nothing is lost, because the sweep is deterministic (`n = floor((now−opened)/auto_ms)`)
and re-triggers on the next APPLIED command (`opened`/`fiscal_period_seq` advance only on commit). An
APPLIED command that swept records the sweep in its resolved arm/receipt (`fiscal_sweep`: periods,
credit deltas, opened/seq before/after — for replay/transparency); a REJECTED command records NOTHING
and re-triggers next time. There is NO "outer applied / inner rejected" compound outcome and NO
universal wrapper on every command. F14's receipts simplify accordingly (the harvest is always an
applied outcome; other commands carry a nullable `fiscal_sweep` arm only when they actually swept).
This supersedes the F11 wrapper's commit-under-rejection clause; the rest of F11 (the sweep math,
`consumed_by_auto`, phase preservation) stands.
