# Axis Stack (Clout v1 PR Interns)

Implements `rfc/clout-v1-and-pr-interns.md` under the owner's 2026-09-25 rulings (Option A,
attainment input). **Fixture-first:** no minted epoch declares the axis stack yet. The only
catalog carrying it is `balance/testdata/axis-stack/economy-v5-fixture.json`, whose CV8 rows are
proposals awaiting owner SHA ratification. `pr_intern_3` is absent until Tier 2 content lands.

## Catalog (economy schema v5)

- **`axis_stack`** `{input, input_cap, cap_reason_key}`:
  - `input` is the closed enum `achievement_attainment_run` (the ruled input) or
    `achievement_score_run` (kept as AC3's contrast arm).
  - `clout_run` always rejects, because no Clout ledger exists under Option A.
  - `input_cap` is a visible hardcap.
  - The block is required if and only if some upgrade carries an axis effect.
- **The axis effect arm** is `{source_id, slot: "axis_stack", target: "all", factor_ppm}`, with
  `factor_ppm` in 1..1,000,000. It may not share an upgrade with the static `upgrades` arm. Each
  axis effect needs one explicit `multiplier_sources` row
  `{slot: axis_stack, target: all, provider: axis_stack}`. The slot and provider pair exactly, and
  an orphan axis source rejects.
- **`axis_at_least {minimum}`** is legal only in an upgrade's `requires`. The loader splits it off
  before the route condition parser runs, so route predicates and gates reject it and the route
  context version is unchanged.
- **Cross-artifact check** (`ValidateAxisInputs` / `validateAxisInputs`): an achievement input
  requires the pinned `achievements` artifact, and `input_cap` must be at least the reachable
  maximum. That maximum is the sum of run-scoped grants for attainment (44 at epoch 8), or all
  grants for the score input.
- **Loader corpus:** the shared Go/TS corpus is `testdata/axis-stack/loader-corpus-v1.json`.

## Company v19 attainment

- **Fields:** Company v19 carries `achievements_attained_run`, a sorted-unique set of run-scoped
  achievement IDs, and the derived `attainment_score_run`.
- **Activation:** v19 exists exactly when the pinned economy declares `axis_stack`, and such a
  bundle must pin `opportunities` (v19 extends the v18 active-play wire). Activation is new-run
  bound, the set starts empty, and it is discarded at Exit. Nothing settles to the Founder.
- **Second hook pass:** the achievement hook runs a second pass (Go `attainRun`, TS `attainRun`)
  against the same pre-achievement observation and proof batch as `NewlyEarned`.
  - It attains run-scoped definitions not yet attained whose condition and proof hold, whether or
    not the Founder owns them for life.
  - It never reads Founder state, and career definitions never attain.
  - An ID newly earned in the transition keeps `achievement_earned.v1`. Every other newly attained
    ID emits `achievement_reattained.v1` `{run_id, achievement_id, score_grant}`, in catalog byte
    order (migration 00081; schema v1 only).
- **Invariants:** every attained ID is run-scoped, the score is their summed grant, and run-scoped
  earned IDs are a subset of attained IDs. Both runtimes reject violations.

## Formula

`x = min(input, input_cap)` is read from Company state only.
`factor_i = (1,000,000 + x × factor_ppm_i) / 1,000,000`, computed in exact integers and then
quantized once. The `axis_stack` slot sits after `milestones` and before `faction`. Contributions
multiply in raw-byte `source_id` order, and each owned axis upgrade contributes even at x = 0
(factor `1e0`). A new input applies from the next accrual interval, because the hook runs after
accrual.

- **Purchase:** `buy_upgrade` enforces `axis_at_least` against the same x. A shortfall is
  `not_eligible/requires`. The RFC says `requirement_not_met`; that is recorded as DESIGN-GAP DG-A.
- **Game UI:** upgrade eligibility agrees with the purchase rule.
- **Publication:** `docs/generated/production-formulas.json` schema 14 publishes the formula. Its
  `pinned` value is null until an epoch declares the stack.
- **Parity vectors:** `testdata/axis-stack/formula-vectors-v1.json` and
  `attainment-vectors-v1.json`, both Go-authored and reproduced byte for byte by TS.

## Gaia law

`server/production/gaia_law_test.go` proves that nothing outside the save codec writes
`CloutLifetime`; under Option A the allowed writer set is empty. No economy resource is a Clout
resource.

## Not yet delivered

- The Desk and snapshot rendering (CV9 / AC11).
- Harness relevance and observation rows (CV10 / AC9).
- A production mint.
- All copy: the keys named in RFC CV8 are owner-authored and still pending.
