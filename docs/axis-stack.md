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
- **Ordinary logged replay:** both Go and TypeScript require resolved active-play
  evidence for Company versions18and19 and replay input version5or later. Earlier
  Company versions reject unexpected evidence. RP-309 corrects the client's
  former v18-only check; kernel0.3.163 records that replay behavior change.
- **Scheduler replay:** v19 uses the existing active-play scheduler, with the
  same catalog, sequence/cursor, clock, expiry and draw checks as v18. RP-310's
  separate scheduler admission correction is recorded by kernel0.3.164; it
  does not change those checks or the terminal evidence contract.
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

Applied receipt snapshots for Company19 include the CV5-derived `axis_stack`
object: `{input_kind,input_value,input_cap,cap_reason_key,saturated,contributions,
product}`; each owned contribution is `{source_id,upgrade_id,factor}`. Raw input
is displayed while factors use the clamped input; saturation is strictly above
cap. Contributions sort by source ID and product uses the live Axis fold.
Ordinary, refreshed-offer and Exit/new-run receipts share this producer; errors
propagate rather than dropping the object. Legacy receipts stay unchanged.
Kernel0.3.166 records this receipt/replay behavior change.
[Executed producer proof and limits](../planning/clout-v1-and-pr-interns/receipt-projection.md).

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

## Axis upgrade roles

Each axis-scaled upgrade must declare at least one role from the existing closed
vocabulary; empty roles reject in Go and TypeScript. This additional floor does
not change legacy static-upgrade admission. A declared role name alone is not
evidence of an executed pool binding: CV1's separate truthful-role requirement
and content review still apply.

## Fixture panel accessibility

The product and each intern's current factor render the supplied canonical value
without resource-amount rounding. The panel places the decimal point within the
coefficient using text (for example, `1.2e0` → `1.2`, `1.005e1` → `10.05`), and
retains canonical notation when expansion would require extra zeros. It neither
recomputes factors nor converts them through a floating-point display value.
The general Amount component's ruled integer/Standard notation remains unchanged.
RP-387's browser regression observes literal product/factor values at attainment
8, 12 and 44; the previous panel displayed the fractional values as `1`.

The Desk panel's native PR Intern progress bars reference their existing
translated intern titles as accessible names and their numeric progress sentences
as descriptions. The controlled browser witness checks distinct row associations
and a refresh that makes one intern owned. This is fixture-panel evidence, not
the served-epoch journey or manual assistive-user acceptance. See RP-303 in the
[implementation log](../planning/clout-v1-and-pr-interns/log.md).

## Not yet delivered

- Full v19 replay coverage beyond the separately scoped RP-309 ordinary dispatch
  correction. The original [24-case observation](../planning/clout-v1-and-pr-interns/logged-policy-research.md)
  records the former rejection honestly; correcting that gate alone does not
  prove all numeric/receipt/event/state, migration, Exit or persistence behavior.
  RP-310's separate scheduler correction now passes all24logged Go/TS receipt,
  event and complete poststate comparisons, plus nine scheduler refusals with
  full rollback (including25hcatchup), nine catch-up and five compatibility
  companions. This is synthetic ordinary-action evidence, not the natural
  player journey, terminal/Exit, accepted accumulation representation or SQL
  transaction proof. The separately predeclared kernel0.3.165 correction now
  enforces v18-and-later terminal evidence and the v5 minimum. All six synthetic
  same-bundle v19 exits match Go receipts, Founder output, three event streams,
  final/new Company states and following first manual actions. All21terminal
  checks pass, including unchanged-state refusals and genuine v16/v18 legacy
  companions. Both real guard mutants fail and restore exactly. See the
  [terminal observation and correction](../planning/clout-v1-and-pr-interns/terminal-research.md).
  Subsequent [pinned activation observation](../planning/clout-v1-and-pr-interns/activation-research.md)
  covers six actual epoch8v18→unmintedv19 four-transition Go/TS sequences and
  one two-epoch Service/Store/Postgres path. All14 existing write faults roll
  back complete rows/heads; normal Exit/retry, old terminal and independent
  Founder histories, plus the new OPEN run's full logged head match. Actual
  floor-reset/census/false-state probes discriminate. This is bounded fixture
  activation evidence, not natural progression, all mode/buff seams, a minted
  default player journey or full CV4/AC7 acceptance. Claude review remains.
  The separately predeclared [action/buff/mode observation](../planning/clout-v1-and-pr-interns/sequence-research.md)
  now covers16 continuous sequences/136 actual command attempts and48 malformed
  claim refusals in each runtime. Real claims, PR2 purchases, compute spending,
  buff expiry, two mode paths,25h catchup and actual provision ticks match full
  receipts/events/states/restores. Census/forged-event/contribution/burst probes
  fail and sources restore exactly. This is synthetic in-memory integration,
  not all seeds/combo caps, live claim/burst persistence or natural/default mint.
  Original27AC6 failures and independent review obligations remain unchanged.
  Subsequent [persisted sequence witness](../planning/clout-v1-and-pr-interns/sequence-persistence.md)
  executes these16 actual Service/Store/Postgres paths with168 logged attempts,
  168 later retries,32 conflicts,48 exact six-stage write faults and16 event
  outbox corruption controls. Full receipts/events, twelve-table rollback and
  retry equality, immutable resources, latest-five retention and all stored
  transitions to the exact saved head pass. Seven initial identical snapshots
  are explicit diagnostic padding, not prior gameplay; runs remain OPEN.
  Actual claim-evidence/lookup/retention mutants fail and restore exactly.
  This adds bounded live claim/burst persistence proof, not completed-history
  breadth, natural/default mint, fullCV4 or AC6 acceptance. Independent Claude
  review, remaining receipt contracts and original27AC6 remain open.
  The [shared Company migration corpus](../planning/clout-v1-and-pr-interns/migration-corpus.md)
  now executes all five exact CV4 names in public Go/TS readers: actual v18
  replay/Exit/new-v19 reset/re-attainment, early-field rejection, derived-score
  rejection and attained-superset rejection. Corpus10/baseline20 preserves the
  original11legacy/4Founder cases. Missing-row/false-SHA/real derivation-check
  omissions fail; the native migration lane passes. This is bounded fixture
  coverage, not fullCV4/Clout acceptance. RP-311's missing receipt object is
  now locally repaired in a separate kernel166 range; real ordinary/refresh/
  Exit tests, cap/ownership oracles and compiling omission/factor controls pass.
  Four reports were fully re-observed; only receipt axis objects and source pins
  changed, not game states/events/inputs/populations. Independent review, AC3
  author reconciliation and all other Clout acceptance holds remain.
- Exact interval-partition invariance (AC6). RP-307's actual Go timing witness
  passes and rejects a retroactive-factor fault, but the unchanged production
  engine fails 27 of128 seeded millisecond partitions under full encoded-state
  comparison. The observed cash differences are one canonical rounding unit
  (1e-7 at the sampled1e4 balance). This is a failed acceptance criterion, not
  permission to restrict cuts or introduce a tolerance. Numeric Core's existing
  per-commit12-digit quantization still applies; research R-012 must establish a
  compatible repair contract before persistence/replay changes. Its first bounded
  Go/TS research wave finds45/520 current differences; conserved test-only models
  match in that population, not across the full numeric domain or real saves.
  [Limits and next contract questions](../planning/clout-v1-and-pr-interns/partition-research.md).
  A separate frozen-anchor experiment extends that sampled range and tests
  reconstruction, not the production implementation. Its
  [producer/persistence limits](../planning/clout-v1-and-pr-interns/anchor-research.md)
  remain prerequisites for a buildable repair.
  Subsequent [actual producer/SQL research](../planning/clout-v1-and-pr-interns/rate-and-sql-research.md)
  shows that early rate serialization changes five sampled payouts and jsonb
  invalidates strict byte framing. Existing context restore and a test-only
  logical reader pass separate declared populations; neither is a production
  repair, adopted save format or complete integrated persistence proof.
  [Actual policy-boundary characterization](../planning/clout-v1-and-pr-interns/policy-boundary-research.md)
  preserves full Company state and integer clock/bank/burst/provision rules.
  It finds eighteen differences in44paired Go profiles, including a separate
  offline-episode scope question(RP-308). No offline reward policy or accepted
  invariant is changed by that research; actual logged TS/persistence remains
  outside this observer's proof.
  The retained
  regression is intentionally red, so the current Go/CI tree is not green.
- Served-epoch activation and the default composed PR purchase/rendering journey
  (CV9 / AC11). Fixture-only snapshot producer and Desk panel code exist;
  their presence does not complete that criterion.
- Harness relevance and observation rows (CV10 / AC9).
- A production mint.
- All copy: the keys named in RFC CV8 are owner-authored and still pending.

The current harness deliberately refuses declared axis stacks rather than
simulating attainment incorrectly. Its cap helper has direct scalar tests, not
an active axis-enabled scenario. DG-D/D-021 requires an accepted evaluation
contract before PR relevance, dead-row and purchase-time measurements can run.
The bounded `527246f1^..527246f1` guard commit is independently reviewed; no
full Clout or archival approval follows. See the [review and pending work](../planning/clout-v1-and-pr-interns/log.md).

The public Go simulation effect mask is separately checked on admitted latched
Company fixtures:16literal rate projections and32online/offline advances retain
ownership/attainment and all other saved fields, while masking only the chosen
PR contribution. Removing the actual mask guard fails rates and full-state
accrual checks. This is a prerequisite proof, not an axis-enabled harness run.
The [criterion map and mask evidence](../planning/clout-v1-and-pr-interns/acceptance-reconciliation.md)
keep all twelve accepted criteria and DG-D/D-021's unselected architecture visible.
