# Clout v1 + PR Interns — implementation log

## 2026-09-25 — Predeclaration (Claude)

**Implemented by:** Claude, on the owner's direction that Claude implements accepted RFCs.
Awaiting Codex's designated cross-party review; nothing here is self-approved.

**Authority:** `rfc/clout-v1-and-pr-interns.md`, accepted 2026-09-25 at recommended defaults.

**Numbering at predeclaration (HEAD `c20d23f1`):** kernel 0.3.125; economy catalog schema 4 is
current and no artifact or candidate claims v5 (`balance/testdata/t2/economy-candidate-v1.json` is
schema 4), so this RFC takes **economy v5**; `LatestCompanyVersion = 18`, so this RFC takes
**Company v19**.

**Scope under A:** CV1–CV5, CV8 (fixture rows only), CV9 producer, CV10, AC4 with an empty writer
set. CV6/CV7 are not built (A rejects them). `pr_intern_3` stays out of all data (OD-5: it lands
with Tier 2 content or is dropped).

**Readings recorded before code:**
- **R1 — Input enum.** The loader accepts `achievement_attainment_run` (the ruled input) and
  `achievement_score_run` (needed as AC3's discriminating contrast arm; it needs no save change).
  `clout_run` always rejects, because no CV6 ledger exists under A (AC1 row "clout_run without the
  CV6 ledger").
- **R2 — `axis_at_least` storage.** It is extracted from an upgrade's `requires` before the
  remainder reaches `routes.ParseConditions`, and stored as the upgrade's own axis minimum. The route
  condition union never learns the kind, so route predicates and gates reject it by construction and
  the route context version is unchanged (CV1 item 3).
- **R3 — Rejection detail.** CV1 says an axis shortfall is `requirement_not_met`, while the shipped
  `buy_upgrade` reports every `requires` failure as `not_eligible/requires`. DESIGN-GAP DG-A: the
  implementation keeps the shipped `not_eligible/requires` pair (the shortfall is a `requires`
  failure; a new detail would widen the rejection taxonomy without a ruling). Owner/RFC author to
  confirm.

## 2026-09-25 — P1: economy schema v5 loaders (Claude)

**Implemented by:** Claude. Awaiting Codex's designated review.

- **What landed:**
  - Go `server/economy` and TS `client/src/economy-kernel.ts` accept schema 5.
  - The `axis_stack` block (closed input enum; `clout_run` always rejects under A).
  - The axis effect arm (`slot: axis_stack`, `target: all`, `factor_ppm` 1..1e6).
  - Mixed-arm rejection.
  - The upgrade-only `axis_at_least` requirement, split off before the route parser (R2).
  - Exact `axis_stack` slot/provider pairing, the explicit `multiplier_sources` row per axis effect,
    and the orphan-source and "iff" rules.
  - `multiplier.Order` gains `axis_stack` after `milestones` (OD-6).
  - `ValidateAxisInputs`/`validateAxisInputs` wired into both bundle loaders. The pinned
    achievements' run-scoped sum is 44 today.
- **Numbers:** `pr_intern_1` (1e6, `axis_at_least 6`, 25,000 ppm) and `pr_intern_2` (5e7, 10,
  20,000) exist only in `balance/testdata/axis-stack/economy-v5-fixture.json`. They are generated
  additively from the epoch-8 bytes by `client/tools/generate-axis-stack-fixtures.mjs`. No
  production artifact changed.
- **Neutral until P2:** static contribution assembly skips axis effects in both runtimes.
- **Corpus:** `testdata/axis-stack/loader-corpus-v1.json` has 22 economy, 4 cross-artifact and 2
  route cases. Go `TestAxisStackLoaderCorpus` and TS `axis-stack.test.ts` (30/30) both pass.
- **Severing:**
  - S1 (Go: drop the mixed-arm check) → `mixed_arm_upgrade` fails.
  - S2: the first attempt did not compile (unused variable), so it doesn't count. The redo with a
    compiling mutant that disables the declaration check → `axis_upgrade_without_multiplier_source`
    fails. `wrong_provider`/`wrong_slot` still reject through the independent pairing rule; that is
    a redundant layer, not vacuous.
  - S3 (TS: `<` → `<=` on the cap bound) → `cross/input_cap_equal_reachable_attainment` fails.
  - S4 (TS: no `axis_at_least` extraction) → 9 cases fail.
  - All mutants were restored.
- **Cold runs:**
  - `make test-go GO_PACKAGES='./economy ./multiplier ./production ./replaycatalog ./routes
    ./achievements ./gameui ./fiscal ./reputation' GO_TEST_FLAGS=-count=1` passes after the slot
    order test is updated for OD-6.
  - Client `vitest run` passes 6826.
  - `make formulas-check` fails, as expected, on the new slot; the regeneration follows in its own
    commit (AC10).
- **Kernel:** 0.3.125 → 0.3.126.

## 2026-09-25 — P2 + P3: axis formula, Company v19 attainment, re-attainment event (Claude)

**Implemented by:** Claude. Awaiting Codex's designated review.

**P3: CV2, CV4, CV5**
- **Save:** Company save v19 (`LatestCompanyVersion` 18 → 19) adds `achievements_attained_run` and
  `attainment_score_run`.
  - Both are required at v19 and rejected before v19 and in Founder scope.
  - The Go codec has a new `companyStateV19` arm; TS restore/encode are mirrored.
- **Derivation check:** `production.validateAttainmentState` / TS `validateReplayAttainment` require
  that every attained ID is run-scoped, that the score equals their summed grants, and that run
  earnings are a subset of attainment. It runs whenever the pinned floor is v19.
- **Version floor:** the Company floor becomes 19 exactly when the pinned economy declares
  `axis_stack`, and such a bundle must also pin `opportunities` (Go `bundle.valid`, TS loader).
- **Activation:** the new-run assembly in `settleAndActivateFoundations` resets attainment to `{}`
  under an axis economy, and to nil otherwise, so nothing carries across an Exit.
  `initializeActivePlayState` no longer lowers a v19 Company to 18.
- **v18 checks now inclusive:** every Go/TS `WireVersion == 18` meaning "active play present" is
  now `>= 18`.
- **Second hook pass:** `attainRun`, which never reads Founder state, and TS `newlyAttained` +
  `attainRun`. It uses the same pre-achievement observation and the same proof batch as the first
  pass.
- **Events:** `achievement_reattained.v1` is emitted for IDs attained without being newly earned. It
  is registered in the Go and TS registries, with migration **00081** (the next free number) and
  the contiguity pin moved to 81.
- **Snapshots:** the receipt snapshot (`wireSnapshot`, both runtimes) and Soul-suppression restore
  carry the v19 fields.

**P2: CV3**
- `AxisInput` gives x = min(input, cap) and `saturated`, from Company state only.
- The axis contribution is `countPPMFactor(x, factor_ppm)`, in slot `axis_stack`, target `all`.
- `buy_upgrade` enforces `axis_at_least` against the same x (`not_eligible/requires`, R3), and the
  Game UI projector's upgrade eligibility agrees.

**Evidence (cold)**
- **AC5:** `testdata/axis-stack/attainment-vectors-v1.json` (9 Go-authored vectors) is replayed
  10/10 by TS `attainment-vectors.test.ts`.
- **AC2:** `testdata/axis-stack/formula-vectors-v1.json` (8 vectors) is replayed 9/9 by TS
  `axis-formula-vectors.test.ts`. x=8 gives ×1.2 and the cap gives ×2.1 / ×1.88, matching RFC
  CV8.
- **AC3:** `TestCarryRuleIsStructural`. A fresh and a veteran Founder on byte-identical Company
  states produce identical attainment and axis product under the attainment input. The veteran's
  events are `achievement_reattained.v1` ×2 and the fresh Founder's are `achievement_earned.v1` ×2.
  Under `achievement_score_run` the veteran's product diverges, which is the discriminating case.
  Receipts are not byte-identical, because the shipped `achievements_earned_run` field
  legitimately differs by Founder. The RFC's "identical receipts" wording is recorded as
  DESIGN-GAP DG-B for the author.
- **AC7:** `TestCompanyV19AttainmentRoundTripAndRejections` (save) covers the wire, missing fields,
  v18 carrying the fields in both directions, the Founder scope and a negative score.
  `TestAttainmentDerivationIsValidated` covers a tampered score, a career ID, an unknown ID,
  earned-not-attained and a nil set. `TestAxisStackOwnsCompanyV19Activation` covers new-run
  activation in the real Exit order.
  - The migration corpus (`testdata/save-migrations.json`) has no v15+ arm, the same gap Reputation
    logged as RT-DG-B, so v19 is witnessed by these unit tests.
  - No existing run replays differently: every pre-axis bundle keeps its semantics, and all
    existing suites pass unchanged.
- **AC8:** `TestAxisStackIntegrationReattainsAndBuysPRIntern` runs on Postgres, through
  `Service.Handle` → store → Postgres, for a veteran Founder:
  - the purchase re-attains `generators_purchased_1` (one `achievement_reattained.v1` row with the
    exact payload, no earned row for it);
  - the receipt shows `attainment_score_run` 8;
  - `pr_intern_2` is rejected `not_eligible/requires` at x=8 < 10, and `pr_intern_1` applies;
  - the constraint rejects an unregistered kind and schema version 2.
- **Severing:**
  - T1 (TS includes career definitions) → 1/10 vectors fail.
  - G1 (Go emits re-attained for newly earned IDs) → the vector golden file fails.
  - G2 (Go drops the earned⊆attained check) → the `earned_not_attained` case fails.
  - AC3 mutant (attain only newly earned IDs, which is equivalent to reading Founder lifetime) →
    `TestCarryRuleIsStructural` fails.
  - AC8 mutant (purchase ignores `axis_at_least`) → the integration test fails.
  - AC2 additive-stacking mutant → 4/9 TS vectors fail.
  - **AC2 float-division mutant: SURVIVES, and is an equivalent mutant.** A search over x ∈ {3, 7,
    11, 13, 29, 37, 43, 44, 97, 1,000,003, 123,456,789} × 127 `factor_ppm` values found no
    difference after the mandated canonical quantization. AC2's named float mutant is therefore not
    observable in this domain; this is recorded as DESIGN-GAP DG-C for the RFC author. It is not
    claimed as a check.
- **Suites:**
  - `make test-go GO_PACKAGES='./save ./production ./replaycatalog ./economy ./achievements
    ./gameui ./multiplier ./routes ./fiscal ./reputation ./gameserver ./account ./releasepackage'
    GO_TEST_FLAGS=-count=1` passes.
  - Docker Postgres `./save ./production ./gameui ./gameserver` passes.
  - Client `vitest run` passes 6845.
  - `validate-migrations` passes.
- **Kernel:** 0.3.126 → 0.3.127.

## 2026-09-25 — P4: Gaia-law structural test (AC4) and AC10 failing case (Claude)

- **AC4:** `server/production/gaia_law_test.go` parses every non-test Go file under `server/` (a
  sanity floor of more than 100 files). It collects every assignment, increment or composite-literal
  write of `CloutLifetime` outside the save codec (`save/state.go`, which only round-trips the
  value). Under Option A the allowed writer set is **empty**. The pinned epoch-8 economy and the
  fixture economy declare no Clout resource, so no reward union (minigame payout, Fiscal,
  opportunities, Commons), all of which bind to catalog resources, can name a Clout arm.
  - **Seeded failing case:** three synthetic writers (`+=`, `++`, literal key) are all reported.
  - **Real severing:** adding `state.CloutLifetime += definition.ScoreGrant` inside `attainRun`
    fails the test at `production/axis_stack.go:101:3`. It was restored afterwards.
- **AC10:** the formula publication landed in its own commit `63aa0b62` (schema 14, `axis_stack`
  section, new authorities `AxisInput`/`axisFactor`/`attainRun` fingerprinted, `pinned: null`
  because no epoch declares the stack). **Failing case:** moving `SlotAxisStack` after `faction`
  without regenerating makes `make formulas-check` exit 2. Restored, it exits 0.

## 2026-09-25 — Docs and plan record; hand-off state (Claude)

- **Docs:** new `docs/axis-stack.md`, plus pointers in `docs/achievements.md`,
  `docs/purchasable-content.md`, `docs/production-engine.md`, `docs/save-layer.md` and
  `docs/balance-harness.md`.
- **Plan boxes:** P1–P4 are flipped against their test-bearing commits (`ace4e67e`, `f256b235`,
  `4c089f00`). Formula commits `80f1bd47` and `63aa0b62` sit in the same range.
- **Cold runs at this coordinate:** `make test-harness` passes (41 s).
  `make verify-kernel-version` stops only at the pre-existing `50a3a514` history item. None of this
  range's commits is named: `ace4e67e` bumps to 0.3.126 and `f256b235` to 0.3.127, and the other
  commits touch only test, planning, formula, tool or doc paths.

**Open (not done in this range):**
- **P5 / AC11:** the snapshot `axis_stack` producer and Desk PR-row progress, plus the composed
  witness.
- **P6 / AC9:** harness relevance for PR rows, the dead-row (`factor_ppm = 1`) fixture, the
  time-to-first-PR observation, and the `axis_input_within_cap` invariant. The scenario-bundle
  rejection already holds, because `ValidateAxisInputs` runs in `replaycatalog`.
- **Copy:** the RFC CV8 keys are owner-authored and pending (OD-7).
- **DESIGN-GAPs for the RFC author:**
  - DG-A: the `requirement_not_met` detail versus the shipped `not_eligible/requires`.
  - DG-B: AC3's "identical receipts" cannot hold literally, because the shipped
    `achievements_earned_run` differs by Founder.
  - DG-C: AC2's float-division mutant is equivalent after canonical quantization.
  - The migration-corpus v15+ gap, shared with Reputation's RT-DG-B.

## 2026-09-25 — P5: snapshot producer and Desk PR rows (Claude)

**Implemented by:** Claude. Awaiting Codex's designated review.

**Producer:** `server/gameui/axis_stack.go` `projectAxisStack` builds the optional v4 arm
`features.axis_stack`: `{input_kind, input_value, input_cap, cap_reason_key, saturated, product,
contributions[{source_id, upgrade_id, factor}], attained[{achievement_id, earned_this_run}],
interns[{upgrade_id, minimum, factor_ppm, owned, factor}]}` (CV5/CV9).
- Every factor comes from `production.AxisFactors`/`AxisProduct`, the same private arithmetic as
  live rates.
- A new fact `feature.axis_stack` is added; the arm is null or omitted unless the pinned economy
  declares the stack.
- API schema `GameUIAxisStackArm` is an optional field (C2). `make api-generate` accepted it with
  `docs/generated/api-compat-v1.json` byte-unchanged; no re-pin.

**Client:**
- `parseAxisStackArm` rejects contradictions: a saturation flag against input/cap, a contribution
  for an unowned intern or with a factor different from its intern's, an owned intern without a
  contribution, unsorted rows, a non-canonical product, an unknown input kind, and extra fields.
- `AxisStackPanel.svelte` on the Desk shows the input against the visible cap, the saturation
  notice with the cap reason, the product, and each intern's progress bar (unowned) and factor.
  Everything is server-derived.
- PR rows in the upgrade list resolve through `axis-presentation.json` (a strict, byte-sorted
  map, with duplicates of pinned upgrades rejected, following the Garage lane's
  features-presentation precedent). An unknown ID still fails loudly.
- Copy: `copy/catalog/clout-candidate.json` (13 keys) is candidate text awaiting owner adoption
  (OD-7). The "PR Intern" naming collision with the `design/01` Intern autoclicker and
  `nephew_intern` stays with the owner.

**Witness:**
- `TestAxisStackArmIsServerDerived` pins the arm for a fixture Company v19 (x=8: product 1.2e0;
  pr_intern_2 factor 1.16e0 with minimum 10; saturation at 60 → 2.1e0; epoch-8 → null) and writes
  the golden `testdata/axis-stack/game-ui-arm-v1.json`.
- `client/test/axis-stack-arm.test.ts` (9) decodes that exact golden and rejects 8 contradictions.
- `client/test/axis-stack-browser.test.ts` (3 browsers × 2) mounts `GameUIApp` with that golden. It
  checks the readout text, the progress bars, the PR titles in the upgrade list, that no mechanical
  ID leaks, axe, and that there is no panel without the arm.
- **Composed lane:** a real fixture-bundle composed run is not possible without a mint (the
  composed gameserver serves only the pinned epoch); this matches Reputation's B7 note. The lane now
  asserts the real server withholds `axis_stack` and reports `feature.axis_stack=false` on the
  pinned epoch. It passes.

**Severing (each red):**
- P5a: removing the Desk panel wiring → browser 1 failed.
- P5b: removing the decoder owned-count check → 1 failed.
- P5c: the producer contributes unowned interns → the Go golden fails.
- Composed: removing the `feature.axis_stack` fact makes the composed lane throw.

**Cold gates:**
- `make test-go GO_PACKAGES='./gameui ./account ./gameserver ./production' GO_TEST_FLAGS=-count=1`
  passes.
- `verify-client-boundary` (17 component files), `copy-check`, `svelte-check`, `formulas-check`
  and `test-game-ui-composed` all pass.

**Kernel:** 0.3.127 → 0.3.128, because `server/production/axis_stack.go` gained the read helpers.

## 2026-09-25 — P6 (partial): harness refusal and the axis_input_within_cap invariant (Claude)

**Implemented by:** Claude. Awaiting Codex's designated review.

**Finding (blocks the rest of CV10):** both harness runtimes execute pre-foundation Company
semantics. `newFirstHourCompany` builds a v14-shaped state, and no achievement hook runs, so the
run-local attainment input cannot be evaluated inside the harness. Severing probe H1 demonstrates
the consequence: with the refusal below removed, a first-hour run on the v5 fixture economy
**completes with a silently neutral stack**. The policy never buys a PR row and nothing reports it.
PR-row relevance (AC9 ANY/ALL), the `factor_ppm = 1` dead-row fixture and the time-to-first-PR
observation all need the harness to execute the attainment pass. Two choices:
- (a) the harness runs the served foundation hook (a larger harness change that affects every
  pacing baseline); or
- (b) a harness-local attainment observer over simulated decisions.

This is a harness-architecture decision for the RFC author/owner (DESIGN-GAP DG-D); nothing was
improvised.

**What landed (`server/harness/axis_stack.go`):**
- `refuseAxisStack`: the Phase-0 `newSuite` and the first-hour `runWithModes` refuse any economy
  declaring `axis_stack` (`ErrAxisStackUnevaluated`). This is AC9's "a v5 economy must fail loudly
  rather than run neutral".
- `CheckAxisInputWithinCap`, the `axis_input_within_cap` invariant, is wired into
  `validateFirstHourCompany`. It is inert without a Company v19 set, so it is currently
  unreachable in scenarios.
- Tests:
  - `TestHarnessRefusesUnevaluatedAxisStack`: Phase-0 and first-hour both refuse; the epoch-8
    control completes.
  - `TestAxisInputWithinCapInvariant`: 8 is accepted, 60 saturates and is accepted, -1 is
    rejected, and the check is inert without the set.

**Severing:**
- H1 (removing the first-hour refusal) → the test fails, and shows the neutral completion above.
- H2 (removing the invariant's own negative check) **survives**: `production.AxisInput` already
  rejects a negative input, so this is a redundant layer, not a vacuous check. The explicit check is
  kept as documentation.

**Push harness:** `make test-harness` takes 42.5 s cold. Nothing multi-minute was added, so no
exhaustive flag is needed.

**Not done:** PR-row relevance, the dead-row fixture, the time-to-first-PR observation, the pacing
envelopes on a minted bundle, and ratcheting the invariant into the scenario `required_invariants`
registry. All wait on DG-D; the registry ratchet would change the ratified scenario bytes.

## 2026-10-06 — predeclare bounded cross-party P6 guard review

Review scope: Claude 527246f1^..527246f1, all six paths inspected, not the full
Clout span/P6/AC9. Current source b76980a7 is the dependency coordinate for cold
execution; guard/test files are unchanged from the reviewed commit. Read full
accepted RFC, CV10/AC9, exact diff and current call sites/production AxisInput.
DG-D still requires author/owner choice, no harness architecture inferred.

Run cold existing two-test population plus relevant harness suite. Predeclare
independent actual-source probes: sever first-hour refusal only, Phase-0 refusal
only, and the cap-invariant function (return nil). Each scoped test must fail;
restore exact source hashes before next/final run. Separately remove only the
negative-score guard: expected defense-in-depth survival through AxisInput is
recorded honestly, not a new discrimination claim. No live verification handle
at edit time; only one temporary source probe at a time, no mutation persists.
Do not label fixture scalar states/default non-axis control as real PR purchases,
attainment hook, pacing/relevance/time-to-first-PR or integrated release proof.
No constants/balance/epochs/scenario/production policy/CI/checkbox/archival/push
changes. Review by: Codex (designated other party for exact Claude scope).
Recorded by: Codex. The new review/predeclaration/record range is separate from
all earlier Codex scopes, which still await Claude; no self-approval of those.

## 2026-10-06 — designated P6 guard review: APPROVED, partial scope only

**Review by:** Codex (designated other party; Claude implemented the reviewed commit).
**Recorded by:** Codex.
**Reviewed range:** `527246f1^..527246f1` — all six changed paths, no exclusions.
**Decision:** APPROVED for the bounded refusal/scalar-invariant addition only.

Not an approval of P1–P5, full P6/AC9, rendered player journey or Clout archival.
Execution dependency coordinate 792a04f1; guard/test files byte-identical to the
reviewed commit. Current caller files include later changes; no verdict on those
later ranges inferred from this review. P5 plan edit names its already-landed
proof commit, not a new checkbox flip. P6 stays unchecked and explicitly partial.

Executed existing population 64440 cold: both named tests pass, supported
epoch-8 first-hour control completes. Initial selector command failed at the
shell before Go due quoting/Make dollar expansion; not a baseline or mutation
result. Correctly quoted selector runs both observed tests, -count=1.

Independent probes, one at a time, every handle terminal before restore:

- First-hour refusal removed (51883): witness fails, actual unsupported axis
  scenario returns `outcome=completed failures=[]`. This directly corroborates
  the misleading neutral completion DG-D warned about, not a source-only claim.
- Phase-0 refusal removed (95041): witness fails on `invalid commons catalog:
  decode: EOF` instead of ErrAxisStackUnevaluated. Its incomplete constructor
  fixture discriminates the typed early refusal only; **not** proof of a valid
  Phase-0 neutral run. The diagnostic says "accepted" but the actual result is
  this different error, disclosed rather than relabelled as successful admission.
- Scalar invariant replaced with no-op (56228): test fails `negative attainment
  accepted`. This is a synthetic scalar fixture, not a valid derived save/run.
- Only explicit negative-score guard removed (16592): witness survives through
  production.AxisInput's independent negative rejection. Expected redundancy,
  not a claimed firing probe or omission from the adversarial record.

Every source restores exact SHA before next/final gate:
first_hour_runner.go 12c48b5bf506ef9b499c66c915e0faabfc712da25c46ce3e71c65dcc3fd55769;
harness.go 139c3ee8410744323e05dd6adff961b12b2f77a8f18a07db92f9fcbfa4a8c78a;
axis_stack.go 613fef3d16fabbdb03f5b0d425a8b6688e31a03b0d579301db2703694ce90ef2;
axis_stack_test.go 37de652ab634d3b48bfda7418f85351cba8603f4898d94c1008e5e0a52a59a5d.
Final root test-harness/selected vet 9433 passes, -count=1, 40.486 s; default
command, not exhaustive maintenance/pacing population or minted PR relevance.
No persistent source/test/balance/kernel/CI mutation, DB/browser/deployment or
release evidence. No cache deletion, archival or push.

RP-241/D-021 now routes existing DG-D centrally: shared foundation hook versus
explicit observer semantics requires author/owner acceptance; neither is chosen
by this verdict. Relevance/dead-row/purchase observation/scenario ratchet remain
unimplemented. RP-242 reconciles canonical docs: replaycatalog pairing rejection
is distinct from runtime's all-axis refusal; fixture snapshot/panel code exists,
but served activation/default PR journey is unproved. Cold existing Game UI
projection witness 10310 passes; component presence is source-confirmed, no fresh
browser claim. These record/doc repairs are Codex work, not secretly covered by
the designated verdict on Claude's old six-path commit. Proper full 1.0 stays open.

### Record-span pin, 2026-10-06

New Codex review/reconciliation span `b76980a7..3709b5db`: two commits / nine
paths, including 792a04f1 predeclaration and all docs/decision/ledger/board records.
This following pin edge also belongs to that span; the closing relay supplies
its exact literal tip. It does not expand the designated-approved Claude range
`527246f1^..527246f1`, supply missing P6 measurements, or consume any earlier
Codex review request. Final record-only diff-check passes, source remains restored
and clean; no verification handle is live. No checkbox/archive/push change.

## 2026-10-06 — RP-303 predeclaration: accessible PR progress association

Implementer: Codex. Baseline b2b2d5a5; clean tree, no live test handles.
Accepted Clout CV9 accessibility and its server-derived progress contract own
this bounded consumer correction. Static census finds native progress bars
without a label; the existing axe rule aria-progressbar-name selects explicit
role attributes, not these native progress elements. Its green result does not
establish a native accessible name. Native reproduction is still pending.

Population: existing Go-authored diagnostic arm mounted through GameUIApp,
both PR rows unowned, then replacement with one owned. Query actual native
progressbar by exact existing translated intern name; assert exact row binding,
value/maximum and contextual description. Refresh must preserve the remaining
association, with no fabricated intents. Existing original tests remain intact.
Failing controls: remove the label binding; bind the second bar to the first
intern. Each must fail the new oracle; restore exact source before final gates.
No source/test/record edits while a verification handle is live.

Only panel/test/docs/ledger/planning changes allowed. No owner text authored or
adopted, producer/arithmetic/generated contract/balance/epoch/kernel/CI change,
mint, cleanup, archive or push. This is not a full designated P5 review, default
served-epoch journey, manual assistive-user study, or full accessibility proof.
All earlier Codex spans remain independently review-pending. Cold local browser,
types/client/build/copy/boundary checks are not hosted CI success.

### RP-303 unchanged-production baseline

After c3606a96 predeclaration, the new native role/name case fails in Chromium
and WebKit: `native progressbar named PR Intern: expected null not to be null`.
Run bc5cb7 / session76051, terminal e6a861, exit2: 2 failed / 4 original cases
passed, 2.32s. Failure is the new association oracle, not a launch or compile
failure; the later refresh assertions were not reached. Make's dependent
performance lane did not execute on this failed invocation. No handle remains.
Local correction binds each bar to its existing translated title and links its
numeric progress sentence as a description. No prose or math is changed.

### RP-303 executed correction and discrimination

Initial corrected run f1a31b /56983, terminal c4be0d: six Chromium/WebKit
cases pass (1.31s); dependent performance lane one pass/22 filtered skips.
Corrected panel SHA256:
`2e3c38c8713f031b3a3bc7d1f7899ee640260ee9c95a66e4e542e45953afc0a7`.

- Missing-label compiling source control: ffa805 /18761, a0391c terminal
  exit2, two new failures/four originals pass (1.36s), exact source restored
  f33259. Later refresh assertions were not reached.
- Wrong-row label compiling control: 2ababf /1901, 68914c terminal exit2,
  two new failures/four originals pass (1.33s). Native strict role/name lookup
  reports two different bars named PR Intern. This fails before the later
  Senior lookup, not a claim of executed later assertions. Exact restore d98c11.
- Final restored run381f77 /52502, b0d480 terminal exit0: six cases pass
  (1.86s), including two named rows, exact value/max, contextual descriptions,
  owned-row removal after refresh and no intents. Performance separately one
  pass/22 filtered skips (1.21s).
- Cold root types/client/build/boundary/copy/topology run3cf837 /82253,
  terminal283091 exit0: Svelte0errors/0warnings, 8241unit pass/340visible
  Node skips, production build213modules, UI boundary22GameUI files, copy658
  keys/hash unchanged plus five generation goldens/six corruptions; content
  manifest unchanged, topology positive/13negative controls. The new browser
  case is a Node skip, never counted there as executed accessibility evidence.
- Firefox0f6012 /13232 emits session-connect timeout after60.03s, no tests,
  1unhandled error (570594). Its process then remains live; only that owned
  handle is gracefully interrupted via Ctrl-C, terminal e6afce exit130. No
  Firefox success, skipped-case substitute, launch workaround or full CI claim.
  Diagnostic `ps` attempt946d9a was denied; no escalation/cleanup inferred.

Existing Actions browser job calls `make test-browser` and its default config
discovers this file in all three projects. No workflow/discovery flag changed;
local selected-engine checks do not prove the hosted Linux run. Previous
RP-131/historical CI red and other independent holds remain unchanged.

Only three production markup lines change, outside watched kernel paths; no
math, eligibility, copy bytes, generated API, balance or epoch moves. The native
role query proves DOM accessibility-name association, not a human screen-reader
study. Full CV9 hints/codex/notices, served-epoch purchase/rendering, P6 authority,
mint and wider release proof remain open. P5 checkbox is unchanged and its text
now explicitly describes presence, not full acceptance. New work afterb2b2d5a5
including these record edges independently requires Claude; no archive/push.

### RP-303 local diff review — first filter, not designated approval

Review by: Codex (implementer, self-first-filter). Recorded by: Codex.
Exact inspected range: `b2b2d5a5..807e719b`, two commits/eight paths,
169 insertions/five deletions. Full diff inspected in599fe7. Scope holds:
three markup bindings, one new browser case, canonical docs and tracking only.
Original cases/assertions remain intact; failed baseline and both compiling
negative controls distinguish absent/wrong association. Source restored exactly,
all verification handles terminal; no production source fault remains. The
new checkpoint is appended at the end of the long-term log (draft placement
mistake caught and fixed before commit). Cold local proof/Firefox failure and
no hosted/AT/default-player acceptance are accurately separated.

Result: local first filter passes. **Claude's designated review remains required
for the entire new span afterb2b2d5a5, including this following review-record
commit.** No existing Claude range is retrospectively approved here, no other
pending Codex span consumed, no checkbox/status/archival promotion. Goal active.

## 2026-10-06 — RP-304 predeclare bounded P1 role review and correction

Review by: Codex. Recorded by: Codex. Original range: ace4e67e^..ace4e67e,
bounded to CV1 item5's Go/TS upgrade-role cardinality and fixture proposals,
not a full seventeen-path P1 designated verdict. Baseline b30808e5 is clean;
no verification handles live. Previous turn was progress (RP-303 fixed).

Static finding: Go rejects nil roles but accepts [], TS checks Array.isArray
then maps it without a nonempty test. Neither branch subsequently validates
upgrade-role cardinality. Accepted CV1 item5 requires ≥1 typed role on axis
upgrades; execution will test this inference rather than trusting the read.

Population: a separate shared role corpus based on the original SHA-locked v5
fixture, empty roles on intern1/intern2/both, and positive unchanged/one-role/
static-empty controls. Each selected id must exist exactly once; actual loaders
execute every case and refusal checks must identify the axis-role floor. Original
loader corpus/fixture remains untouched. After confirmed failure, the narrow
loader guard rejects only roleless axis upgrades. Kernel161→162 is required
because the acceptance set genuinely changes; no spurious balance/schema bump.
Independent compiling Go/TS guard omissions must fail the three roleless cases,
all controls retained; restore exact sources before cold final gates. No manual
edits during any live test handle. Applicable root verification, full client,
types/build and historical kernel-guard outcome must be recorded honestly.

Does not authorize a new role vocabulary or invent a synergy pool/rate. CV1's
truthful executed pool-binding question is separate and cannot be resolved by
adding a nonempty label. No owner text, mint, archive, push, deployment, CI change,
full P1/AC1 or 1.0 approval. New Codex range needs Claude, separately from the
original bounded finding and every previously pending correction.

### RP-304 role-branch designated finding: CHANGES REQUIRED

Review by: Codex. Recorded by: Codex. Original Claude range
`ace4e67e^..ace4e67e`, limited to CV1 item5's role-cardinality branch; no
full-P1 verdict. The static hypothesis is independently reproduced at current
013eb7b1 plus new tests: actual Go loader accepts all three roleless mutations
unchanged (fe6913 exit2, 0.226s; three valid controls pass), and actual TS loader
accepts the same three (e9ab00 /94844, terminal668509 exit2). Complete client
run: three new failures/8245pass/340skip, no other failures,4.14s. Both tests
pin the unchanged original fixture and check selected ids actually exist.

This test-only checkpoint precedes the correction. Remedy is the accepted
nonempty-axis-role rule only; static [] remains legal. It does not establish
whether the proposed synergy_feed labels have real pool bindings. Neither
normative owner text nor the original corpus is rewritten; later Codex runtime
repair and record edges need Claude, never covered by this original finding.

## 2026-10-07 — RP-304 narrow correction and restored evidence

Test-only failing checkpoint ec080fcf precedes the runtime change. The Go guard
and TS guard reject empty roles only when axis effects exist. Existing enum and
duplicate rules stay unchanged; legacy static[] still loads. Kernel161→162 in
the same implementing change records a genuine catalog acceptance-set change,
not a balance retune. Canonical axis docs update alongside behavior.

Independent compiling omission controls:

- Go guard removed:58526a exit2,0.072s, three new roleless failures/three legal
  controls pass. Go restored before the TS fault, SHA verified in4d167e.
- TS guard removed:4d167e /49145, terminal6c9a27 exit2; same three new failures,
  8245other passes/340skips,5.04s. Restore76ff53 verifies both exact source SHAs:
  Go d6c376b31bd7a0c33d4e6698bd780331e5760e9f550222460a8b0ab6e26c4667;
  TS 338f0c47c9b4a0aee26e2318b07caf7ac7ce0f391404a208162894c2ba64cf86.

Final cold checks, after all production faults restored:

- 4ad5af /30717, terminal1c7ba5 exit0: complete economy/achievements/multiplier/
  replaycatalog/kernel/production/gameui packages pass, production37.112s,
  gameui0.294s; vet for the same seven packages passes. Ordinary host SQL skips
  are not counted as database execution.
- ac280f /85834, terminal5ba866 exit0: types/Svelte0errorswarnings,
  8248unit passes/340visible skips, build213modules, formulas byte-unchanged,
  component boundaries and topology with13negative controls pass.
- 52e031 /90270, terminalaee96b exit0: actual declared Postgres production
  Integration package12.969s and gameui0.187s pass; visible public-projection
  test executes. Output is truncated, so no fresh exact population/skip census
  is claimed from this capture. Orphan warning is not cleanup authority.
- 5778b9 /67719, terminal1f5e8f exit0: original loader corpus, seven new role
  tests and three panel cases all execute in Chromium/WebKit,80pass; dependent
  performance1pass/22filtered skips. Firefox is not rerun or claimed covered;
  prior RP-303 session timeout remains unexecuted evidence.
- 32f3fa /72299, terminal1f2736 exit2: CI kernel-history checkout and adversarial
  fixtures pass, but actual complete history guard again rejects50a3a514
  against0cf9f7a6 for six watched minigame files without a bump. RP-131 remains
  RED. Our three kernel constants agree at162; this is NOT a green complete CI
  claim, and no history correction or guard exception is invented.

The old fixture and loader corpus retain their exact SHAs:
f8c67aed784ceaae1335bb40c0e0487cc0d9b78a50b9e63a73569270df7511d7;
2a05d90054144599351d06f42d2b08e93147214fccd5da57f4cc569bc8015ee2.
No golden/report restamping, copy, production epoch, formula, schema or CI edit.
Every verification handle is terminal; no source fault remains.

RP-305/D-022 is a separate observed content-contract hold: c99da6's read-only
fixture census returns no synergy-pool source for either named PR upgrade,
although both declare synergy_feed. CV1 item5 explicitly requires truthful
execution binding; the role-floor fix alone does not deliver it. Owner/RFC
author must reconcile real binding/ratified values versus vocabulary semantics.
No normative body or balance value is changed here.

All new implementation/records afterb30808e5 require Claude independently; the
bounded original role finding is not a full seventeen-path P1 verdict or a
review of Codex's repair. No boxes/status/archival or fullAC1/mint/release claim.
The public Typer/Arcade API-versioning decision was asked non-blockingly; no
answer is assumed or recorded as authority. Full nine-tier/platform goal active.

### RP-304 local range review — first filter only

Review by: Codex (implementer, self first-filter). Recorded by: Codex.
Exact inspected range: `b30808e5..bb3f6714`, all sixteen changed paths,
including predeclaration, failing-first shared cases, runtime repair and records.
The paired tests exercise the actual loaders, pin the fixture and require each
mutation target exactly once. Guards scope the new floor to axis upgrades;
static-empty admission, role vocabulary and original artifacts remain intact.
All three kernel constants advance together with the real acceptance change.
Executed baseline, omission controls, exact restores and terminal checks are
recorded above; historical kernel-history failure remains visible. Diff checks
and restored source/artifact hashes agree. No balance, copy, schema, CI, checkbox
or archival edits occur in the range.

Result: bounded local correction passes the first filter. This is NOT Claude's
designated pass. Claude must review the entire new span after `b30808e5`,
including this following review-record commit; no older independent span is
absorbed. RP-305/D-022, D-021, mint and full P1/AC1 holds remain live.

## 2026-10-07 — RP-307 timing/partition predeclaration

Baseline07bb3d9d, clean tree, no live verification handle. The P3 row names
AC6 but original axis tests provide isolated formula/attainment vectors and one
SQL purchase, not a seeded interval-partition or retroactive-factor witness.
Bounded original f256b235^..f256b235 AC6 evidence finding: CHANGES REQUIRED;
this is not a verdict on all28 changed paths or the other P2/P3 criteria.
Review by: Codex. Recorded by: Codex. Static inventory only so far.

Population/method/controls/limits are predeclared in plan.md's RP-307 section.
Exact full-state comparison and independent purchase-interval arithmetic must
not be softened to produce green. New tests first run against unchanged source;
any numeric contract gap is routed without improvising persistence or balance.
No product/kernel/CI/copy/mint/owner body change or checkbox flip authorized by
this test-only range. New Codex span needs Claude separately; earlier pending
review spans and author/owner holds remain. Full nine-tier/platform1.0 active.

Initial instrument attempt e9bde8 /86171, terminal61fecd exit2: all64 timing
arms reach the purchase and x=8, but my Founder fixture owns only purchase_1;
other eligible run definitions correctly mint first-earn score6. That invalidates
the instrument's claimed veteran-earned0 control, not the product. Correct the
fixture to own all run definitions, with their exact derived lifetime score;
preserve the observed failure and the timing arithmetic/assertion unchanged.

Instrument follow-ups: cc60ba exit2 is a test-authoring compile error
(Definitions is a slice, not a method), corrected before measuring. ba455c
/1439, terminal00c35d exit2 executes all192 new cases: the timing setup still
omits lifetime career achievements and correctly first-earns score4. Make the
diagnostic veteran own the complete pinned achievement catalog with exact sum;
do not alter the independent timing oracle or production hook. The separate
128 partition arms are admitted and execute:27 exact full-state divergences,
101 equality controls. A representative online end3114/cut1553 gives
cash1.00035846811e4 vs1.00035846812e4. Whole-second and tiny-cut controls pass;
the seeded population still disproves AC6. This is not a skip, floating tolerance
or threshold waiver. Further investigate and route the numeric contract before
any runtime repair; preserve the literal property failure.

Corrected unchanged-source timing run4c5dca /37795, terminale2cca5 exit0,
0.295s:64 applied online/offline arms. The fully awarded Founder and its exact
derived lifetime score isolate attainment from first-earn output; this is an
admitted diagnostic fixture, not evidence of naturally earned progression.

Temporary future-input control: first54555b exit2 is an invalid probe compile
(AxisStack returns value/bool, not a pointer), not discrimination. Corrected
f677a8 /74802, terminal2c0768 exit2,0.411s: timing fails on the independent
triggering-interval arithmetic (example online cash9.9935048e3 versus required
9.9933046e3). Output truncates in that capture; it is not a fresh full mutation
population census. c0862d restores the production file byte-exactly before any
other check: replay.go SHAa2bf1bf6a6338b0548587d067dc96569a8355363cc50f1b2905782a7edc2b6ca;
engine.go SHAb872d4746779a097f68f88475cef66214a2a7f8970817298f788ffad515ae875.
No production/kernel change is retained.

Restored59ee91 exit2 and stronger receipt witness46ef08 /72550,
terminal3f9387 exit2 reproduce only the partition property failures. Timing64
passes including new_revision2/snapshot attainment8 and exact reattainment ID;
128 partition arms yield27 failures/101 equal controls, independent seeds and
online/offline populations unchanged. No SQL or TypeScript partition execution
is inferred. Sourceinspection identifies economy.Ledger.apply's per-commit
Quantize(12); Numeric Core K3 explicitly requires that boundary. The observed
one-rounding-unit difference is a counterexample to AC6, not a proved exploit
or permission to remove quantization. Its repair needs an explicit compatible
accumulation contract, scoped empirically by R-012 before a follow-up RFC.

The literal new regression stays RED in the ordinary Go suite: no skips,
tolerances, env switches, build tags, filtered fixtures or CI gate bypasses.
This is an intermediate failing-proof checkpoint, not completed implementation.
Original f256b235's bounded AC6 evidence remains CHANGES REQUIRED; neither the
new timing witness nor a future local first-filter covers the full original
28-path range. Entire Codex span after07bb3d9d needs Claude independently.
Earlier review/owner/API/pool/harness/mint holds and full1.0 scope remain.

Final unchanged-source verification:

- 44c064 /63336, terminaleb98d8 exit2: full cold Go core completes. Production
 35.944s fails only the new27 partition arms; other selected core packages pass,
 transport13.321s. Harness is intentionally outside this root target, and host
 SQL skips are not counted as database evidence. This is a RED core gate.
- 07c436 exit0: vet production/economy/decimal.
- 8c11fe /10793:8248 client passes/340 visible skips,4.69s; types/Svelte
 zero errors/warnings. Terminal473135 exit2 at the unchanged historical kernel
 guard50a3a514 vs0cf9f7a6. Checkout/adversarial guard fixtures pass; actual
 history remains red. Later topology target in that command never runs.
- Separately b9a6a2 exit0: CI topology plus13 refusing negative controls.
- 9bee70 /32657, terminal98c99c exit0: declared real Postgres executes
 TestAxisStackIntegrationReattainsAndBuysPRIntern (PASS,0.10s; package0.107s),
 not skipped. This is one existing persistence witness, not SQL coverage of the
 new partition/timing population. The orphan warning is not cleanup authority.

All verification handles terminal, temporary source fault absent, kernel162 and
production/epoch/copy/schema/CI bytes unchanged. R-012 is predeclared READY
research, not already executed or permission to relax AC6. Updated shared
ledger/docs/queue/roadmap keep the failed gate and next research visible. The
new failing regression intentionally affects ordinary core/CI execution; no
green current/hosted CI, full Clout, archival or1.0 claim. No push or deployment.

### RP-307 local diff review — failing-proof checkpoint, not acceptance

Review by: Codex (implementer, self first-filter). Recorded by: Codex.
Exact inspected range: `07bb3d9d..2d492926`, all nine changed paths,
including predeclaration and final evidence records. Test fixture admission and
actual ApplyLogged purchase precede the independent old/new interval arithmetic;
the partition oracle compares complete encoded states, never only helper factors.
The32 timing draws and64 partition draws each run both modes with seed307;
no source faults, skipped new cases or conditionally disabled gate remain.
Historical artifact/balance/owner/production/CI bytes stay unchanged. Plan
presence is explicitly distinguished from acceptance, with no checkbox flip;
append-only logs and boards match the executed red/green populations.

Decision: CHANGES REQUIRED for AC6 acceptance; timing supplement is locally
validated and the exact partition counterexample is preserved, not fixed.
The structural/evidence first-filter is not Claude's mandatory designated pass.
Claude must inspect the complete new span after07bb3d9d, including this following
record edge; no original28-path or older independent span is absorbed. R-012
remains research-before-contract, and no tolerance/persistence change is adopted.

## 2026-10-07 — R-012 conserved-state first-wave predeclaration

Previous goal turn: progress (committed actual timing proof and partition failure).
Current baselineb6a3c48d is clean; no live verification handle or source fault.
Exact542-case Go/TS population, seed, arms, eight refusals, controls, authority
and excluded domain are predeclared in plan.md before measurement. The existing
RP-307 literal regression and historical kernel guard remain RED, not removed.

Static golden inspection changes the instrument design: existing midpoint
1.234567890135e0 is canonically1.23456789013e0 in the shipped float runtime,
not the ideal rational-half-even result1.23456789014e0. This is recorded numeric
authority, not a new defect or license to restamp it. Exact rational accumulation
must project through the existing quantizer and pass the pinned controls; an
unbounded arbitrary-precision production rewrite is neither needed nor authorized
by the bounded prototype. Full1.0 goal remains intact; no new owner decision or
public API delegation is inferred from this automatic continuation.

## 2026-10-07 — R-012 bounded conserved-state observation

Review by: Codex (implementer; no designated verdict). Recorded by: Codex.
Predeclared61f01349 afterb6a3c48d. New test-only Go/Rat and independent
TS/BigInt models, actual unchanged Go Evaluate baseline, actual TS scalar
accrual/projection, selected source identities and new complete artifact.
No existing balance/golden/runtime/kernel/CI/copy/save representation changes.

Generation2448f8 /15127, terminal4b3366 exit0,0.393s:542 primary cases/eight
refusals complete. Current45/520 full-state differences; one-shot/reference
cash0 differences, discarded carry45 hits, retroactive8 hits, cap4 controls.
Conserved final wire+residue agrees exactly with the reference throughout the
declared population. Whole-second controls pass; reference residue97 negative/
99 positive/324 zero. This is a signed correction, not merely hidden positive
currency. No SQL/served TS/browser/full-domain proof is inferred.

Cold nongenerationd1c1d3 /52561, terminald41791 exit0,0.297s: byte-identical
artifact reproduction. Full client/type7a3190 /45281, terminald770ee exit0:
8799pass/340visible skips,4.59s; types/Svelte zero errors/warnings.

False-promotion discrimination changes only new artifact acceptance_status to
PROVEN. Go a0df70 /16958, terminalba705f exit2 rejects artifact drift; client
81b510 /14517, terminal0d33eb exit2 fails exact NOT_PROVEN metadata (one failed,
8798pass/340skip; other550 research cases pass). Both handles terminal before
exact patch restoration. No source mutation or faulted generation.

Final restored cold252086 /49242, outputsfd9fb6 and terminal3f99c3 exit2:
production36.956s fails only existing27/128 RP-307 partition arms; economy6.131s
and decimal0.220s pass. New research passes but acceptance stays RED. Final
bc8157 /56284, terminal0fff69 exit0:8799client passes/340skips,4.57s; types/
Svelte zero errors/warnings. All handles terminal before record edits. Earlier
RP-131 historical guard remains separately red; full/hosted CI not claimed.

partition-research.md records exact population, observed arms, reproduction,
discrimination and explicit restore/cap/full-domain/settlement limits. Queue,
ledger, docs and long-term board reconcile to the same bounded result without
checkbox or status promotion. Next is predeclared representation/restore research,
not speculative production fields or a tolerance. Entire new span afterb6a3c48d
needs Claude; older independent spans remain live. Full nine-tier/platform1.0
active, no mint/owner-body/API ruling/archival/push/deployment or release claim.

Final record-bound replaye73760 /94126, terminal77f2cd exit0,0.412s reproduces
the original artifact exactly; vet7940b4 exit0 on production/economy/decimal.
All verification handles are terminal. No temporary artifact/source fault remains.

### R-012 local range review — bounded research, no implementation acceptance

Review by: Codex (implementer; self first-filter). Recorded by: Codex.
Exact inspected range: `b6a3c48d..d73fd408`, all twelve changed paths, including
predeclaration and evidence records. Generator loops implement the declared
population; artifact rows are audited by cold Go byte reconstruction and actual
TS independent arithmetic/source/population checks, not manual inspection of
15,049 JSON lines. Selected hashes are explicitly not full transitive provenance.
Original data/goldens/runtime/save/kernel/CI/specification bytes remain unchanged.
Full-state Go versus scalar TS scope, experimental debit reset and bounded
enormous-range/restore/cap exclusions are explicit rather than implied away.

The implemented negative arms and executed false-promotion failure discriminate
within this population. Source-selected artifact updates require explicit valid
re-observation; they cannot close AC6, which remains literally red in ordinary
tests. All record homes agree on counts and holds; no checkbox/status promotion.

Decision: locally validated bounded observation; no production repair or Clout
acceptance. This is not Claude's designated verdict. Claude must cover the full
span afterb6a3c48d, including this following record edge, independently of all
earlier pending ranges. Next predeclared representation/restore research is
allowed; a persistence implementation still requires a buildable accepted
contract. Full nine-tier1.0 active; no archive/mint/push/deploy/release action.

## 2026-10-07 — R-012 second-wave predeclaration

Previous goal turn: progress (first bounded paired research completed/committed).
Currentb54fc7ef clean; no live verification handle or source/artifact fault.
AGENTS/process/current accepted Clout body and unchanged accrual primitives/goldens
re-read. The next wave is predeclared in plan.md:615 paired primary cases,
16 restore refusals each, three Go-only carry-codec diagnostics, unchanged-
primitive anchor versus per-evaluation rebase, enormous exponent/max-time edges
and explicit observed settlement/cap limits. No new persistence contract inferred.
Original first-wave artifacts/instruments and acceptance redtest remain untouched.
New range needs Claude; no prior human question is answered by this automatic
goal continuation. Full proper nine-tier/platform1.0 active.

First instrumentb46a8d exit2 is a test-authoring compile error: the prior
research struct names EndBefore/CutBefore, not BeforeEnd/BeforeCut (and likewise
after). Correct those references before measuring. No artifact was written or
product failure inferred; population/oracle/predeclared expectations unchanged.

Correcteda8639a /84353, terminal2f8a57 exit0,0.291s:615 primary cases/16
restore refusals complete, rebase45/domain4/golden5 refused/prior-boundary0diff/
max223bytes. Go-only carry diagnostics admit inconsistent projection and trailing
JSON, refuse below-cap quantity projecting to cap (residue-1/500000000). Those
are experimental model limits, not a production save defect. Client4be18b
/97110, terminalfa3128 exit0:9434pass/340skip/types zero errors/warnings.
An initial combined-run invocation9909bb fails in the shell before Go executes
(unquoted regex/Make dollar expansion); corrected5251bc /15694, terminalfebac9
exit0 runs BOTH new anchor and unchanged first-wave tests cold,0.341s.

After every handle is terminal, strengthen selected provenance to pin the actual
old Go carry helper directly, and size census to include all1215 valid snapshot
instances, not only frozen/golden/domain rows. This changes instrumentation
identities, so explicitly re-observe rather than silently restamp the source hash.
No case/policy/golden/production/first-wave-byte change or population narrowing.

Re-observationc096ca /51849, terminal866a50 exit0,0.359s preserves all counts
and223-byte maximum. Next predeclared test-only instrument fault: override the
recomputed visible wire with the serialized wire before validating reconstruction,
in Go and TS separately. The valid-but-wrong wire/rate negatives must actually
fire (not only the self-source SHA). Faulted Go generation must leave artifact
SHA68984669c75a530330b1112e7c951ac8d51cb0c5b86d576c44c40784febc0548 untouched.
Original instruments Go86dd7588b35393eb947ba977125787177f85f42754b0d19534677da0141900d2;
TS1bcb9a1b1415b59b40b24020ffa9ab80354c99c5bef013a061355e6166e3702d.
Both handles must become terminal before exact-byte source restoration and
restored cold verification. No production mutation or designated review implied.

Go instrument fault71f7d6 /83567, terminal04836a exit2 fails before generation
at named wrong-valid-rate admission,0.363s. Map order selects the first failure;
do not claim all16 Go negatives completed under the fault. Clientbf5db4 /35773,
terminal3c2c27 exit2:three failures (source SHA plus BOTH wrong-valid-rate and
wrong-valid-wire refusal assertions),9431pass/340skip. Thus the semantic checks,
not merely artifact identity, discriminate. 05ffeb confirms both artifacts'
original hashes unchanged. Both handles terminal before exact source restoration.

Restorationb61217 matches BOTH original instrument hashes and BOTH artifact
hashes exactly. Final950f6a /98871, terminalc0ac98 exit2:production37.245s
fails only existing27/128 AC6 cases; economy6.367s/decimal0.236s pass. Separate
d9b3a2 /37445, terminala9cbd8 exit2,0.183s independently executes the same
128 acceptance arms and counts27 failures. New research stays ordinary/default,
and the production regression is neither bypassed nor diluted.
Client179d1d /46932, terminal4600c4 exit0:9434passes/340visible skips,5.18s;
types/Svelte zero errors/warnings. Vet3489e0 exit0 on all three packages. Every
verification handle terminal before record edits; no source fault retained.

New anchor-research.md names actual results and restricted schema/domain,
carry-codec diagnostics and real semantic fault failures. First-wave dossier
links the new evidence without changing its old code/artifact. Shared ledger,
queue/docs/plan and full1.0 board stay synchronized, no acceptance/status/checkbox
promotion. Next empirical seams are live raw-rate serialization and real jsonb
round-trip; canonical source lists/byte-framed self-serialization are NOT
presumed full producer or Postgres compatibility. Offline/provision/migration/
whole-state/receipt/replay and owner/author/API/content/environment/release gates
remain. Full new Codex span afterb54fc7ef needs Claude independently; older
ranges remain live. No runtime/save/kernel/CI/golden/balance/copy/owner-body
changes, archival/mint/push/deploy or full nine-tier/platform1.0 completion claim.

Final record-bound cold8e4a11 /73430, terminal92151a exit0,0.518s executes
BOTH original first-wave and new anchor observers, reproducing each artifact
exactly. All verification handles terminal. First-wave code/corpus remain
byte-unchanged; no temporary probe or runtime change retained.

### R-012 anchor local range review — research only

Review by: Codex (implementer, self first-filter). Recorded by: Codex.
Exact inspected range: `b54fc7ef..ba4d3d70`, all thirteen changed paths, including
predeclaration, two new test instruments, generated artifact and every record.
Generator/population/codec/negative dispatch reviewed;22,069 artifact lines are
verified by actual cold reconstruction, source census and independent TS
execution, not claimed individually read. Original instrument/golden/corpus,
runtime/save/kernel/balance/copy/CI/RFC bodies remain byte-unchanged.

The primary/refusal/diagnostic populations are distinguished; Go-only carry
findings are not represented as client/production findings. Scalar candidate
state is not full Company state. Restricted source count/size and sampled
numeric domain are explicit. Actual reconstructed-wire faults fire semantic
assertions before successful measurement/promotion; source/artifacts restore
exactly. Records agree and no acceptance checkbox/status flips.

Decision: locally validated bounded comparison, not a production repair or
designated approval. Live rate/context fidelity, jsonb, offline/provision,
migration/receipts/full state/replay remain required. Historical research gates
are tied to the measured baseline; a future accepted repair must explicitly
reconcile them to changed implementation while preserving counterexample and
negative-control evidence, never quietly remove AC6 or tolerate its failures.
Claude must cover the entire new span afterb54fc7ef, INCLUDING this following
record edge; older independent spans remain live. Full nine-tier/platform1.0
active, no owner/body ruling, archive/mint/push/deploy or release action.

## 2026-10-07 — R-012 producer/SQL third-wave predeclaration

Previous goal turn: progress (paired anchor/codec comparison committed).
Baseline9050fe4d clean; no live verification handle or probe. Actual rate
assembly retains intermediate precision; existing save storage is jsonb. Both
seams are declared before measurement:64 actual producer profiles,1215 temporary
SQL round-trips/16 SQL negatives, separate artifacts/writers, exact validation
and typed refusal provenance, scope/limits in plan.md. Read process/current
index/instructions and relevant unchanged producers/save/database/Compose code.
No owner question answered by automatic continuation; full1.0 active. No
production or old-corpus edit/acceptance/kernel/balance/body/CI change authorized.

Third-wave instrument finding before landing: the first SQL replay compares its
entire artifact byte-exactly, including observed linux/arm64 and Postgres16.15.
That would manufacture a failure on a valid linux/amd64/other16.x runner despite
identical semantic observations. Reconcile the instrument (not CI): keep actual
writer environment provenance, require Postgres16/Linux supported architecture,
and explicitly exclude ONLY version/architecture from replay equality. Sources,
complete populations and all normalized JSON/statuses remain exact. Add local
metadata/refusal fixtures; these are comparator proof, not executed AMD64 CI.

### R-012 third-wave executed observation and fault probes

CPU writer e2e43a/30197 terminal36cfc5 exit0:64 profiles complete; raw bits
change64, early-rounded accrual differs5, existing Encode/Restore context exact64,
actual Evaluate equals raw primitive. Artifact has fourteen selected source pins.
Client fbe314/59753 terminal0263ab exit0:9499pass/340visible skips, types/Svelte
zero errors/warnings. No natural progression/served TS producer claim.

SQL first attempt eab446 exit2: sandbox socket denial before connection, NOT
executed. Narrow declared-test retry0f516d/45348 terminal9d734c exit0:real
Postgres16.15/linux-arm64, all1215valid and16negative payloads complete. Strict
refuses1215, logical full state preserves1215;12semantic/domain refusals,
one typed22P02, three representation-only normalized inputs. Ten selected source
pins; separate writer/artifact. Only transaction-local TEMP table/savepoints,
rollback; no live save/migration/cleanup/write authority.

Fault probes ran only after all handles terminal, on NEW TEST instruments:
5e24a1/92108 terminal7b9414 exit2 substitutes rounded accrual for raw and
fails context equality before CPU writer. TS755d69/93689 terminalf15ab5 exit2:
five exact semantic payout failures plus one source check, not merely source
identity rejection. SQL1364a3/9780 terminal6545bf exit2 silently reconstructs
wire and fails wrong-valid-rate admission after actual SQL casts, before writer.
All three faulty sources restored byte-exactly (a30318); both artifacts retain
their pre-probe hashes. No temporary production mutation or changed acceptance.

After restoration: declared SQL9a6b5c/92742 terminal28ffae exit0, exact artifact
replay; client19afca/32814 terminal7a5084 exit0,9499/340/types clean; relevant
vet14cc5d exit0. Cold production/economy/decimal29d91d/34957 terminalce42e0
exit2:production37.802s fails ONLY original AC6; economy6.141s/decimal.232s pass.
Focused cold456d91/67726 terminale97602 executes128cases,27fail/101pass,.347s.
No complete CI-green claim. Duplicate verbose invocation876445 also red; it
was truncated in displayed output and is not the census source.

Metadata fix is instrument reconciliation before landing, not new CI authority.
Final SQL writerf8c7bf/8601 terminal2f5fc8 exit0,real16.15/arm64,1215/16complete;
new source pin/artifact after this explicit re-observation. Missing-DB generation
eecdef/29481 terminala26e2d fails loud before writer. Host cold185aea/20596
terminal087345 passes all three arithmetic observers plus environment comparator;
SQL visibly NOT EXECUTED there. Earlier research corpora reproduce unchanged.
Topologydc9da0 exit0,13negative controls. Local declared AMD64 CI03e3d3/86550
terminale619cd exit2:pull succeeds, Go fails before execution (exec format error).
No AMD64/hosted pass, emulation bypass, CI edit or Docker cleanup inferred.

Results/next empirical questions in rate-and-sql-research.md. No new save format,
raw diagnostic wire, runtime/K3/balance/copy/body change, status/checkbox, archive/
mint/push/deployment or full1.0 acceptance. CPU context and SQL scalar prototypes
remain separate populations. All new after9050fe4d need Claude independently,
including later record edges; earlier spans and owner/author questions remain.

Final verification/environment closeout:vet9d09c6 exit0. SQL replayb8f0b9
fails before Go after the AMD64 pull changed the shared golang:1.26 tag;
image inspectionf105a0 confirmsamd64/3343c365. Native declared-service pull
6d0c52/59215 terminal37f41c succeeds, no deletion or Compose edit;8c3278
confirmsarm64/b6081f19. Final cold336006/4038 terminal00c828 exit0,.458s:
ALL five research tests execute on real Postgres16.15/linux-arm64, none skip;
542/615/64 arithmetic and1215/16SQL populations reproduce unchanged. Native
image restoration is NOT an AMD64 CI bypass or passing AMD64 observation.
All handles terminal; no probe remains. Old artifacts retain their pinned hashes.

### R-012 producer/SQL local range review — research only

Review by: Codex (implementer, self first-filter). Recorded by: Codex.
Exact inspected range: `9050fe4d..8d0d9ce4`, all fifteen changed paths, including
predeclaration, three new test instruments, separate corpora, manual generation
lane and every synchronized record. Instruments, population/decoder/writer
dispatch and record diffs inspected;8,667 artifact lines verified by complete
actual cold reconstruction and independent TS checks, not individually read.

The raw producer/canonical-rate arms genuinely differ; existing saved-context
proof is limited to admitted diagnostic Go profiles. SQL uses the actual
declared database and no persistent table; logical equality covers every field.
Representation-only normalization is not corrupt-field admission. Typed syntax
failure and missing-DB generation cannot masquerade as successful measurement.
Environment comparison preserves exact semantic/source evidence and explicitly
records actual provenance; its local fixtures are not AMD64 execution proof.

Both original research artifacts remain byte-unchanged; diff verifies no runtime,
save/kernel/balance/copy/RFC body/CI configuration change. The new manual Make
target has no verify/CI dependency. Mutation failures are semantic, not solely
source-pin mismatches; exact restoration and final actual SQL replay verified.
No checkbox/acceptance/status promotion. Production AC6 and local AMD64 execution
remain red/unexecuted respectively; those failures are explicit in the records.

Decision: locally validated bounded research, NOT a production repair or
designated approval. Offline/banking/boost/provision, full-state/action/replay,
activation/migration and Service/Store proofs remain prerequisites for a repair
contract. Historical research gates must be explicitly reconciled in any accepted
future change, without erasing the counterexamples or weakening AC6. Claude must
cover the ENTIRE span after9050fe4d, INCLUDING this following record edge; all
older independent spans/owner/author holds remain. Full nine-tier/platform1.0
active; no archival, mint, push, deployment or release call.

## 2026-10-07 — R-012 policy-boundary predeclaration

Previous goal turn:progress (64actual producers and1215actual SQL snapshots
committed; unsafe transplant assumptions rejected). Baseline d61a5248 clean,
no live handles or new designated verdict. Re-read full current instructions,
process/accepted Clout body and relevant canonical engine policy/source/helpers.
Fourth-wave plan declares44actual Go full-state paired observations with strict
existing save restoration, independent integer clock/bank/burst/provision bindings
and four negative probes, separate artifact. No TS engine export or copied engine
is invented. Offline cap-per-call is distinguished from one absence episode;
observations cannot change intended offline policy or rule the author's AC6 text.
No owner answer inferred from automatic continuation; full nine-tier1.0 active.

### R-012 fourth-wave execution and separate author question

Cold writer908ec1/50662 terminal4c50ca exit0,.666s:44paired profiles complete,
18different final states. Full corpus132actual phases and220exact restore
points;11selected source identities. Offline12/8different, burst16/6different,
provision12/4different, clock4/0different; all integer bindings hold. All
eighteen differ in cash; five offline cases also differ in credit. Full-state
not scalar comparison. Provider counts/carry agree in both branches.

RP-308 added immediately after executed observation/first live probe terminal:
offline cap-per-call histories cannot be conflated with internal partitioning
of one catch-up episode. Asked owner to delegate wording-only reconciliation
while preserving all current offline rewards, banking and reconnect behavior;
no answer received/preselected option not authority. No CV3/AC6 body edit,
new episode field, tolerance or player exploit inferred. Decision/research queue
route the author question. RP-307's numeric failure remains independently red.

Five separate NEW-OBSERVER faults, each cold writer expected to fail:

- Bank2a1143/7012 terminal9eb2e2 exit2:uncapped reported bank/state at48h with
  credit259199999 fails integer binding before restore/writer.
- Burstf5b3e8/57032 terminal9ddd09 exit2:retain old remaining burst1 at3114ms
  fails integer binding.
- Provisionb58257/94751 terminalc36103 exit2:omit generated count/carry fails
  provider1/minute60000 integer binding.
- Clock201205/9799 terminal802969 exit2:advance zero-work cursor fails complete
  state check atonline/end0.
- Census460aad/62147 terminal0a52ae exit2:remove oneprofile, incomplete44guard
  fails before any observation/writer.

These are mutations of reported observer state/census, not production mutations.
Each handle terminal before the next edit. Final367ed5 confirms exact source
71c0f5a1/artifactb4262aa2 match pre-probe hashes; no old artifact modified.

Restored cold default692c1b/26058 terminaldfb2d0 exit2:production38.605s fails
ONLY original27AC6cases; economy6.095s/decimal.228s pass. Relevant vet864cce
exit0. Client4afb8a/79606 terminala8161d exit0:unchanged9499pass/340visible
skips/types/Svelte clean. Declared native3e0150/20077 terminal302b47 exit0,
1.078s:all SIX research tests execute, no skips; old SQL1215/16 and every
previous corpus reproduce exactly. New policy observer is still Go-only, not
SQL-backed policy persistence or actual TS evaluation merely because it runs
inside that Docker service. AMD64/history/hosted holds remain unchanged.

Record preparation caught an erroneous six-credit-differences sentence; actual
artifact census770dcb verifies five and the uncommitted text is corrected.
A multi-file patch had a mistyped final context; it failed atomically, then
the records were applied with verified context. Neither is a product finding.

Observation/limits/next contracts in policy-boundary-research.md. All handles
terminal, no temporary probe. No runtime/kernel/balance/copy/save/body/CI/status/
checkbox/archive/mint/push/deploy change. Entire new span afterd61a5248 needs
Claude; prior spans and unanswered owner/author questions remain independent.

### R-012 policy local range review — research only

Review by: Codex (implementer, self first-filter). Recorded by: Codex.
Exact inspected range: `d61a5248..a08eec2d`, all twelve changed paths, including
predeclaration, new Go observer/corpus and every synchronized record. Source,
population/phase/restore/integer-policy/writer dispatch and record diffs reviewed.
19,426 artifact lines are validated by complete cold reconstruction, not claimed
individually read. Actual source/full-state bindings and four policy/missing-row
negative failures inspected; exact restoration verified.

The observer reports eighteen differences, not eighteen equivalent defects.
Offline histories obey legitimate per-call policy; RP-308 requests author scope
reconciliation. Ten other cash differences remain numerical findings. All integer
provision/burst/bank/clock controls and220full-state restores hold. Separate
Go-only engine evidence is not relabelled TS, SQL Service/Store, player history,
mixed-boundary coverage, accepted anchor, migration or release proof.

Whole inspected span changes no runtime/numeric/save/kernel/balance/copy/RFC
body/CI/Make bytes. Old corpora stay byte-identical. Tracking agrees, no checkbox
or acceptance promotion. Local record review caught a draft count error and
misplaced new decision paragraph before commit; prior decision bodies remain
byte-unchanged. No owner reply or designated verdict was fabricated.

Decision: locally validated characterization, NOT a repair or designated
approval. RP-308 delegation remains unanswered. Paired actual logged Go/TS,
action/buff/mode/nonzero-resource and persistence/replay/activation remain
required. Claude must review the ENTIRE span afterd61a5248 INCLUDING this
following record edge; all prior independent spans/holds remain. Proper full
nine-tier/platform1.0 active; no archive/mint/push/deploy or release call.
### R-012 fifth-wave start — actual logged policy replay

At clean HEAD7effff9a, predeclare24 actual logged Company cases and9 copied
catch-up refusals per runtime, plus missing-row/forged-state controls, in plan.md.
No measurement has run for this population. Go and TypeScript real ApplyLogged
producers are the subject, not an arithmetic proxy. Synthetic combined inventory
must admit unchanged; observed refusals/mismatches invalidate a claim of parity.
All previous red AC6 findings and independent designated-review spans remain;
owner/author RP-308/API/cleanup questions are unanswered. No product/body change.

### R-012 fifth-wave result — actual logged Go/TS seam fails

Predeclaration6141af35 at7effff9a. New Go/TS observers and separate source-pinned
logged-policy-research-v1 corpus; every old corpus/source unchanged. Go actual
ApplyLogged24purchases/57full v19 restores/9explicit catchup refusals pass.
All24TS initial restoration comparisons pass; all actual logged calls then fail
before accrual at active-play presence. RP-309 immediately ledgered: TS recognizes
onlyv18 there, Go and the actual restore codec recognize the v19 extension.
Six TS missing/from catchup negatives are MASKED, not target refusals; three
to-coordinate negatives refuse in parser unchanged; census passes. Thirty new
red tests retained. No output-parity, repair, policy or AC6 claim is made.

Writer259032/41578 terminal3b3c59 exit0; typo correction re-observation7621c6/
73159 terminal216d84 exit0. Missing-row writer probeeb7fc6/36300 terminal3ac966
exit2 at incomplete24guard BEFORE writer. TS forged initial-state expectation
434912/95055 terminal6792ed exit2:24positive failures move to exact initial
assertion rather than active-play dispatch. That proves initial comparator,
not unreachable final comparator. Both probe edits restore exact pre-probe
hashes at2e4515; no production mutation or final-output proof inferred.

Self-check then caught a nonzero-permits string oracle comparing against0e0
instead of emitted0. Corrected to actual Decimal positivity with zero-resource
controls, plus corrected TS Ms spelling, then complete re-observation75d933/
99195 terminalcba3e3 exit0. Final3df7c6 hashes: Go4a82177d, TS1b6372f4,
artifactb400abe4. This is a corrected instrument lineage, not unchanged probe
hashes falsely claimed across the later correction; no expected output changed.

Final cold4fb470/50050 terminald9fc30 exit2: production41.833s fails ONLY
original27AC6cases; economy6.094s/decimal.114s pass. Final client98efcf/78550
terminal7180d6 exit2:9503pass/30newfail/340visible skips (newfile4pass/30fail).
Earlier make test-client typecheck vet stopped at client red, so NOT types/vet
proof. Independent final046f5e/96513 terminalef7377 exit0:types/Svelte clean,
full go vet pass. Native3d58b7/66014 terminald08639 exit0,1.437s:all SEVEN
research tests execute, no skips; old SQL1215/16 and other corpora unchanged.
New logged observer remains in-memory Go even inside declared Postgres service.
All verification handles terminal before edits; orphan warning did not authorize
cleanup. AMD64/history/hosted holds unchanged, no restart/tolerance/skip.

Full observation/limits and actual Go samples in logged-policy-research.md.
Separately predeclare accepted-CV4 RP-309 runtime repair (kernel protocol,
old-v18 compatibility,missing/unexpected evidence refusal,exact outputs).
Nothing in this research range changes runtime/kernel/save/balance/copy/CI/body/
checkbox/status/archive/mint/push/deploy. RP-308 delegation unanswered. ENTIRE
new span after7effff9a needs Claude including records; prior spans independent.
Proper nine-tier/platform1.0 active; this turn adds a concrete integrated defect,
not a reason to proclaim paired numeric research complete or abandon the goal.

### R-012 logged-policy range review — local first filter only

Review by: Codex (implementer, self first-filter). Recorded by: Codex.
Exact inspected range: `7effff9a..b55f2131`, all twelve changed paths, including
6141af35 predeclaration and all implementation/record changes. New observer
source and full record diffs inspected;12,866artifact lines validated through
complete Go reconstruction and declared-population census, not claimed read
individually. Actual TS attempts expose RP-309, not paired payout success.

Explicit check of source scope: only test files/artifact/dossier/planning/backlog
and canonical limitation documentation changed; no runtime/old corpus/kernel/
balance/save schema/copy/CI/RFC-body byte moved. No checkbox or acceptance flips.
24Go purchases/57restores/9refusals hold; final client30failures remain exact
upstream v19 refusal and six masked negatives, not relabelled successes. Wrong
zero spelling/property typo corrected and re-observed, not hidden as product
findings. Initial comparison fault is not claimed to prove final comparison.

Decision: scoped reproducible research and an actionable accepted-CV4 defect,
NOT runtime repair, AC6 pass, independent approval or archival authority. Entire
span after7effff9a INCLUDING this following record edge requires Claude's
designated pass; all prior independent spans/holds remain. Next bounded RP-309
repair requires its own predeclaration/kernel protocol/re-observation. Goal
active; RP-308 unanswered; no mint/archive/push/deploy/release call.

### RP-309 correction start — accepted CV4, separate from R-012 research

At clean974c1a45, preceding goal turn was progress: actual logged checks expose
v19 active-play refusal, with failing evidence committed. Existing accepted CV4
is authority to repair the dispatch, not an unanswered product choice. Predeclare
the one-line two-predicate correction, honest kernel162→163, unchanged24/9
outputs with explicit source re-observation, five companion controls and two
actual guard mutants in plan.md BEFORE implementation. Nothing absorbs or
approves the preceding research spans; RP-308 wording question remains unanswered.

### RP-309 bounded ordinary-dispatch correction — deeper RP-310 exposed

Predeclaredff50031e after974c1a45. Runtime diff ONLY the two ordinary v18-only
presence/version predicates to>=18, plus honest kernel0.3.162→0.3.163 in source
and both mirrors. Companion TS tests use actual old pinned v18/v16 fixtures,
not demoted axis state. No scheduler/terminal or numeric body/code changed.

First actual39-test execution9564b9/6403 terminal07cd5a exit2:all nine catchup
target refusals and all FIVE compatibility/refusal companions pass; all24
positive calls now fail at scheduler's separate v18-only state guard. RP-310
immediately ledgered. Terminal guard also tests exactly18:SOURCE finding only,
not an executed terminal failure or authority to expand this presence-only range.

Old guard mutantc941d4/9873 terminal364c59 exit2:31newfile failures, including
missing-v19 evidence wrongly applying a purchase. Permissive mutant80d1d8/86392
terminal7e9c90 exit2:28newfile failures; all FOUR companion refusals fail (three
apply, unexpectedv16 throws the wrong scheduler error). Positive calls continue
to hit unchanged scheduler. Both probes compile and restore exactly atccc181:
runtime92da4b6e/observer040c48de/artifactb400abe4. No live handle during edits.

Explicit complete Go writerccc181/84678 terminal3ec00b exit0,0.471s. JSON
comparison408220 against974c1a45:ONLY runtime/TS-observer source identities
change, ALL bundle/profile/payload/input/receipt/event/poststate/negative counts
unchanged. Current artifactf8154b7e, prior red source/artifact retained atb55f2131.
No output restamp/tolerance or old research artifact rewrite.

Final clientb1b346/77367 terminale6d214 exit2:9514pass/24fail/340visible skips;
newfile15pass/24fail, all failuresRP-310before output comparisons. Final Go
c65515/3218 terminal2393eb exit2:production42.813s fails ONLY original27AC6;
economy6.241s/decimal.225s/kernel.214s pass. Independentc72804/87214 terminal
6d02c0 exit0:types/Svelte zero errors/warnings, fullvet, topology13controls pass.
Native64e87a/98043 terminal9ad025 exit0,1.431s:all SEVEN research tests execute,
old SQL1215/16complete. New observer itself still in-memory, not SQLtransactions.

Kernel742ea9/44773 terminal9e100f exit2:CI checkout contract and fixtures pass,
full kernel history remains red at50a3a514(RP-131), not this source163 correction.
Independent version fixtures3cdc46/95578 terminalb3000f exit0. A read-only ps
check was sandbox denied; no escalation/restart, actual handles confirm terminal.
AMD64/hosted/history and all other author/owner/content/release holds remain.

Records synchronized; no checkbox/RFC approval/archival/mint/push/deploy. Entire
new span after974c1a45 needs Claude including record edges; preceding research
and all prior obligations independent. Next predeclare scheduler correction and
actual terminal v19 population under acceptedCV4. Goal active; this turn repairs
one real runtime rejection and proves retained refusal boundaries, not full1.0.

### RP-309 local range review — ordinary presence correction only

Review by: Codex (implementer, self first-filter). Recorded by: Codex.
Exact inspected range: `974c1a45..e2a605b2`, all fifteen changed paths, including
ff50031e predeclaration, one runtime guard, same-commit kernel source/mirrors,
five companion tests, explicit source re-observation and all synchronized records.
Source/artifact diffs inspected: exactly two report source identities changed,
zero bundle/context/payload/input/receipt/event/poststate/negative-count changes.
Earlier research corpora and all numeric/save/balance/copy/CI/RFC bodies untouched.

Both predicate obligations remain enforced: evidence iff Company>=18, and active
replay version>=5. Actual v18 good receipt/events remain exact; four invalid
companions refuse at the named guard unchanged. Old guard lets missing-v19
evidence apply a purchase; permissive guard fails all four refusal checks.
Actual executed mutations, not a source-only claim of discrimination. Restored
runtime92da4b6e is the single authorized change and kernel163 is a real behavior
signal, not a correction of historicalRP-131 or a full-history pass.

Decision: bounded local repair validated, NOT designated approval or complete
CV4/Clout acceptance. All24positive calls still stop atRP-310scheduler before
output comparison; terminal check is only a source finding. Counts/limitations
agree across docs/ledger/queues/roadmap. No checkbox or archival/status promotion.
Claude must review ENTIRE new span after974c1a45 INCLUDING this following edge;
all older independent spans remain. Next accepted-CV4 scheduler/terminal work
must be separately scoped; proper full1.0 goal remains active.

### RP-310 scheduler start — accepted CV4, new independent range

At clean685debe7, previous goal turn progress: ordinary presence repair is
committed and discriminator-tested, and actual calls expose scheduler rejection.
Predeclare ONLY scheduler admission predicate, honest kernel163→164, nine
before/after-evidence refusals, two actual mutations and unchanged24outputs
in plan.md. Terminal exact-v18 remains separately unexecuted source finding;
this scope does not guess a full-terminal fix or waive accumulated-state red.
No measurement for the corrected scheduler population has yet run.

### RP-310 scheduler result — ordinary full Go/TS parity, not terminal acceptance

Predeclared483fa5ed after685debe7. Runtime diff ONLY scheduler state admission
!==18→<18, with honest kernel163→164 and mirrors in same change. Existing
catalog/before/after/attended/draw/expiry/compound checks unchanged. Terminal
predicate/numeric/save/balance/copy/CI/RFC bodies untouched.

Actual first restored48case populationa20434/37331 terminal9dcb86 exit0:
all24Go/TS full receipt/events/encoded poststates match and restore; all9new
scheduler before/after faults refuse with complete initial rollback, including
combined25hcatchup; prior9catchup/5companions/census pass. No expectations changed.

Old scheduler mutant8a1e35/95046 terminal025b9b exit2:newfile27fails (24valid
calls refuse and3after-sequence controls hit WRONG earlier guard). Severed
after-sequence checka8d0e7/41625 terminalcff578 exit2:exactly3controls fail,
corrupt after_sequence+1 inputs APPLY purchases; all valid outputs still match.
Both compile, real runtime mutants, not whole-red-exit inference. Restored at
1a2f34:runtimec37164b6/observerc562de1d/artifactf8154b7e exactly pre-probe.

Full Go re-observation53b65f/59396 terminal241710 exit0,.363s. Comparison
7cb802 with685debe7:ONLY two selected source identities change; ALL24original
bundle/profile/payload/input/receipt/event/poststate/negative counts unchanged.
Current artifactc5c56c9c; prior failing code identities preserved in Git. Old
research corpora untouched; no output restamp/rounding tolerance/new save format.

Final clientef4381/2079 terminal5d829b exit0:9547pass/340visible skips;
newfile48pass. Independentaa95a7/22046 terminal1f5c01 exit0:types/Svelte
zero errors/warnings,fullvet,topology13controls pass. Native294194/32331 terminal
ad880b exit0,1.536s:all SEVEN research tests execute/no skips, oldSQL1215/16.
New ordinary observer itself remains in-memory, not a SQLtransaction claim.
Cold14319f/33627 terminal5e4186 exit2:production46.037s fails ONLY original
27AC6cases; economy7.005s/decimal.280s/kernel.161s pass.

Declared formula/API verificationb0af15/33693 terminal7b77b6 exit0:generator
outputs byte-unchanged. Kernelc02227/44135 terminal82ffc1 exit2:CI checkout
contract/fixtures pass, full history still fails historicalRP-131/50a3a514.
Independent version fixturesb5f0c8/65820 terminal22d225 exit0. No source/record
edits while verification handles live, no restart, cleanup or mutant left.
AMD64/hosted/all prior author/owner/content/release holds remain explicit.

Records synchronized. Ordinary24case parity is now real and bounded, NOT
partition/mode-switch/buff/full-domain/terminal/Exit/persistence/natural-player/
release proof. Next predeclare actual terminal v19 producer/replay/next-run
population under acceptedCV4; terminal predicate is still only a source finding.
RP-308 delegation unanswered, no owner/body edits or accumulation waiver.
Entire new span after685debe7 needs Claude including record edges, prior ranges
independent. Goal active; no checkbox/status/archival/mint/push/deploy/release call.

### RP-310 local range review — scheduler admission only

Review by: Codex (implementer, self first-filter). Recorded by: Codex.
Exact inspected range: `685debe7..65eb0cdf`, all fifteen changed paths, including
483fa5ed predeclaration, the single scheduler predicate, honest kernel164 and
both mirrors, nine refusal tests, source-only re-observation and linked records.
Runtime diff admits Company>=18 without removing any evidence validation. The
terminal guard and all numeric/save/balance/CI/RFC body bytes are unchanged.

Executed evidence: all48newfile cases pass, all24original Go receipts/events/
poststates match exactly. The original guard rejects24valid replays; severing
after-sequence validation lets three corrupt inputs apply purchases. Both real
runtime mutants are restored. Artifact comparison changes ONLY two source pins,
not outputs or population. Full client9547pass/340visible skips; types/vet/
topology/formulas/API/native research pass. Original27GoAC6 and historicalRP-131
remain red; this local correction does not waive them or establish hosted CI.

Decision: bounded repair validated, NOT designated approval or full CV4/Clout
acceptance. Counts and limitations agree across the dossier/docs/ledger/queues/
roadmap. Terminal/next-run behavior is not yet measured. Claude must inspect the
ENTIRE span after685debe7 INCLUDING this record edge before any archival claim;
all prior independent review obligations remain separate. No status promotion.

### RP-310 terminal research start — separate test-only scope

At clean32116b14, scheduler first-filter record complete. Predeclare six actual
v19 wind_down/next-run pairs and three refusal companions in plan.md before
execution. Current terminal exact-v18 guard remains SOURCE finding. No product
fix in this research range; actual Go/TS results will determine the next bounded
repair. Existing helper restoration hardcodes v18, so this observer must use
actual v19 restoration/admission rather than demote or alter old fixtures.

### RP-310 terminal result — valid exits rejected, missing evidence applies

Test-only predeclarationf7a6e92a after32116b14. New Go observer/TS companion/
six-pair JSON/dossier; no production/numeric/save/balance/copy/CI/RFC body bytes.
Go actual six wind_down transitions preserve final attainment8, reset newrun3
v19 to empty/0, initialize scheduler, then apply first manual action/cash1e0.
Three invalid terminal evidence controls refuse full initial state unchanged.
Go writer8d7979/32346 terminald4cca4 exit0,.293s. No legacy fixture demotion.

TS initial9c86b9/92755 terminal26ba63 exit2:16newfile declarations8pass/8fail,
client9555pass/8fail/340skip. Six valid calls reject terminal presence; missing
current evidence APPLIES; nextsequencecontrol masked at wrong earlier guard.
Six direct restored-Go-next-run cases match full outputs/restore, NOT TS
terminal continuity. Missingnext andcensus pass. Terminal comparators unreached.

Go truncatedpopulationbcabfe/19511 terminal40b414 exit2 beforewriter. Forged
nextpoststate6ebdd6/22869 terminal1db72f exit2 adds six direct-next failures,
14newfile failures total. Original8remain separate. Exact restorationc8eeb9:
Go43bdaa7f/TS3b1c218b/runtimec37164b6/artifact1c62bc71. No handle live at edits.

Final gates:cliente97d54/84146 terminal77b5bc exit2,9555/8fail/340skip.
Go05fb3c/63242 terminal105133 exit2:production41.496s ONLY original27AC6,
economy6.170s/decimal.222s/kernel.167s pass. Types/vet/topologyf8e4f8/36301
terminal9edd6c exit0:zero diagnostics/13controls. Nativea5a383/66237 terminal
54bb21 exit0,2.759s:all EIGHTresearchtests execute,oldSQL1215/16complete.
Newterminalobserver in-memory, not newSQLtransaction witness. Kernelcbd1f2/
44231 terminal6f2bdf exit2, historicalRP-131/50a3a514, checkout/fixturespass;
independent8e16d1/78507 terminal1b0b07 exit0. All handles terminal, no mutants.

Records synchronized. Next separate acceptedCV4terminal guard repair, honest
version signal/legacy companions/real guard probes/source-onlyreobs. RP-308
unanswered; no body/accumulation waiver or CI/AMD64/hosted/release claim. Entire
newspan after32116b14 including records needs Claude; older spans independent.
Goalactive/progress, no checkbox/status/archive/mint/push/deploy/release call.

### RP-310 terminal research local range review

Review by: Codex (implementer, self first-filter). Recorded by: Codex.
Exact inspected range: `32116b14..3bb4e154`, all twelve changed paths including
f7a6e92a predeclaration, new observer/TS checks/corpus/dossier and all records.
No production byte moved. Independent3ad749 verifies all15source identities,
six applied pairs, exact new-state/next-input joins andthree refusal count.
Population comes from real v19 restore/builders/transition, not v18 demotion.

Executed dropped-row and forged-next-state probes fail their own reached checks;
restoration hashes exact. Go exits6/next6/refusals3 complete. TS terminals6red,
missingcurrent wronglyAPPLIES, sequencecontrolmasked; directnext6passes cannot
prove terminal continuity. Counts/limits/old27AC6/history failures consistent
across records. Decision: valid bounded research and concrete repair finding,
NOT CV4/Clout acceptance or designated approval. Claude must inspect ENTIRE
new span after32116b14 INCLUDING this edge; all older ranges remain separate.

### RP-310 terminal correction start — accepted CV4

At clean0c4d6481, six actual valid exits reject and missingcurrent wrongly
applies. Predeclare ONLY terminal presence/version predicates, honest kernel165,
five old-version companions, real guard probes and unchanged outputs after both
reports' explicit source re-observation. New runtime range distinct from the
preceding test-only research and scheduler correction. No measurement of the
corrected terminal population has run yet.

### RP-310 terminal correction result — bounded full outputs and continuity

Predeclared72c9983b at0c4d6481. Runtime diff ONLY two Company===18terminal
predicates→>=18, honest kernel164→165/mirrors. Every other evidence, draw,
next-bundle, claim, scheduler, state/output/numeric/save/balance/copy/CI/RFC
body byte unchanged. Five real pinnedlegacy companions added, oldfixtures
unchanged. All21terminalchecks pass, original six fullGooutputs andfirstnext
actions exact. Original16declarations retained. Ordinary48cases still pass.

Initialcorrected957fce/29972 terminal4dbc02 exit0,9568/340skip. Actual oldguard
b5ec5f/84402 terminal0e83df exit2:exactly8newfilefails,6valid calls refuse and
missingcurrent wronglyAPPLIES;sequencecontrolmasked. Permissive26a166/32630
terminal2f132e exit2:6refusalcontrols fail; some wronglyapply, others escape to
wrong later diagnostics. Sequence+1 still hits next-schedule validator. Both
real compiling runtime mutants restored;9e0849/312a33 sourcehashes exact.

First342225writer failed SHELLselector syntax beforeGo; no corpuswritten.
Correctedbf51a6/72005 terminal3b5814 exit0,.385s explicitly re-observes both
whole populations. Independentbbd1e7 with0c4d6481 excludes ONLYsource_sha256:
loggedchanges1pin,terminal2pins, ALL observed bundles/profiles/commands/evidence/
receipts/events/states/refusalcounts identical. Older corpora untouched.
Runtimeba2f2ea8,TSterminalobservere37a2813,terminalartifact687e43f2,
loggedartifactea238d75; oldsource/artifactidentities preserved at0c4d6481.

Final client938498/30265 terminalfb625b exit0,9568pass/340visible skips.
Types/vet/topology1418b3/72593 terminal35fcab exit0,zero diagnostics/13controls.
Native d3908b/55080 terminala1fe44 exit0,1.497s:allEIGHTresearchtests execute,
oldSQL1215/16exact. Newterminalpopulation still in-memory, notSQLtransactions.
Cold14f875/28131 terminala2ff3c exit2:production43.020s fails ONLY original
27AC6;economy6.089s/decimal.122s/kernel.171s pass. Clientbuild9521d5 exit0.
Kernel6ba675/39623 terminal11a924 exit2:checkout/fixturespass, historical
RP-131/50a3a514 remains. Independentfe57e4/49134 terminaleb18ff exit0.
All handles terminal beforeedits; no restart/cleanup/mutantleft. A multi-file
recordpatch failed context matching and wrote NOTHING; corrected patch applied.

Records synchronized. Full outputs/TS-created next-run continuity now proven
for this six-case same-bundle synthetic population, NOT allpreactivation/buff/
mode/partition/persistence/natural-player/release behavior. Next separately
predeclare actual pinned pre-v19→v19 activation and relevant seams underCV4.
RP-308 delegation unanswered; no owner-body/accumulationwaiver. Fullnewspan
after0c4d6481 needs Claude including everyrecordedge; older spans independent.
Goalactive/progress, no checkbox/status/archive/mint/push/deploy/release call.

### RP-310 terminal correction local range review

Review by: Codex (implementer, self first-filter). Recorded by: Codex.
Exact inspected range: `0c4d6481..7bc300c3`, all seventeen changed paths,
including72c9983b predeclaration, two runtime predicates, kernel165/mirrors,
five legacy companions, two source-only report re-observations and all records.
Single runtime line changes only Company===18→>=18 twice. No other transition,
numeric/save/balance/copy/CI/RFC body or historical fixture change. Source and
artifact diffs8a2ac1 inspected; bbd1e7 whole reports identical except pins.

Six complete terminal/next-run sequences andfive actual legacy companions now
execute; three original refusals plusfour legacy/version refusals retain exact
diagnostics/full unchanged initial states. All21checks pass. Original guard
eight failures/permissive guard six failures demonstrate actual rejected-valid
and admitted-invalid cases, with later independent refusals explicitly retained.
Restoration and counts agree across code/dossier/docs/ledger/queues/roadmap.

Decision: bounded local repair validated, NOT designated approval or fullCV4/
Clout/1.0 acceptance. Original27AC6 andhistoricalRP-131 remain red; AMD64/
hosted/preactivation/action-buff-mode/persistence/natural-player holds not waived.
Claude must inspect ENTIRE newspan after0c4d6481 INCLUDING this record edge;
preceding independent spans remain separate. No checkbox/status/archival claim.
Initial uncommitted review-record placement matched an earlier repeated line;
diff inspection caught it and the record moved to EOF before commit. Existing
log bytes must remain an exact prefix; no prior entry is edited/reordered.

### CV4/AC7 activation population predeclaration

Resumed clean65c3a34c. Read-only grounding confirmed actual next-bundle
resolution, foundation reset, existing Store fault seam and replay readers.
Next accepted test-only population is recorded in plan.md: six actual pinned
epoch8 v18→unminted v19 sequences with fresh/veteran lifetime state; one actual
two-epoch Service/Store/Postgres sequence with fourteen full-row rollback
faults, retry and old/new/Founder replay. No experiment or mutation has run.
Observer failure is a finding, not permission to change runtime/balance/body.
Old source/corpora unchanged. All previous cross-party and release holds stay.

### Activation instrument corrections before final observation

Go6/TS7 pass, full client9575/340 visible skips. SQL first attempt28911→815e8a
failed BEFORE gameplay on two current epochs (test seed helper is single-epoch).
Explicit fixture old-epoch closure preserves pin before next seed. Second
10813→d6286f hit live Founder parity: fixed noon fixture was in the database
clock's FUTURE and Fiscal refuses sweeping before its period opened. Use actual
database clock minus1h for SQL-only initial fixture; deterministic corpus stays
unchanged. Third53793→d25cc9 reaches/fires ALL14 rollback faults and completes
normal Exit/retry/new actions, then fails the provisional replay oracle.

Read actual replay ownership: generic old test reader includes automatic Fiscal
prefixes, but Company replay does not own those; existing reputationCareerReplay
does the correct separation, independently checked by full Founder history.
Also VerifyReplayRun requires a TERMINAL run; newrun3 is deliberately open.
Clarify the predeclared "old/new history replay" means completed old verifier
PLUS full actual new ApplyLogged receipt/event/head equality, with explicit
ReplayLogGap for the unfinished run. No gameplay gate weakened, no runtime
edited; this fixes two instrument assumptions rather than hiding failures.
These corrections are provisional until the actual corrected SQL run executes.

### Pinned activation result — actual producers and transactional evidence

Predeclaredea43ef5b at65c3a34c. Corrected86481/82638f SQL path passes0.300s.
Six complete Go/TS four-transition sequences preserve oldfloor18/newfloor19,
empty initial attainment then independently re-attain grant2 without lifetime
duplication. Full outputs/restores exact; next browser state actually comes
from browser Exit, not a supplied Go shortcut. No production byte changed.

All14faults reach exact sentinels/full12table/head rollback; normal Handle
commits/retries once; actual old completed/Founder histories verify, new OPEN
run logged receipts/events andfull saved head match. Retention pruning is NOT
claimed. Initial single-current/futureclock/wrongreader assumptions remain
disclosed in preceding entry and activation-research.md; none is a product bug.

Go literalfloor mutant29441b BUILDfailure not counted; real floor-minus1
86111/2bb4a6 fails pinnedartifact/saveversion at firstactualExit. TSfloor
35774/977e0b fails six newcomplete receipt oracles +previousterminal6 (12).
Census14609/bda9cd fails despite missingcase enumeration; forgednextpurchase
20793/f4735b fails fullpoststate. All exact original sources restored before
writer/finalchecks; no handle edits/restarts/mutantleft.

Final newwriter83135/c9eeb2 .362s; native10227/85d1213.247s ALL10research/
integration populations execute/noSkip, SQL1215/16oldreportsexact. Client
13077/b89da3 9575pass/340skip. Types/vet/topology/build72913/141726 zero
diagnostics/13negativecontrols/213modules399ms. Cold45068/4d3ef1 production
38.883s fails ONLYoriginal27AC6;economy6.211/decimal.234/kernel.174pass.
HistoricalRP-131/AMD64/hosted remain, no wholeCIgreen claim. Existing Actions
server PG/core andclientVitest discover the newtests; source routing not hosted
execution. Kernel165/oldcorpora unchanged; selected sourceSHA pinned newreport.

Docs/ledger/queues/roadmap synchronized; next separately predeclare actual
action/buff/mode seams. RP-308 delegation unanswered, no accumulationwaiver.
Whole newspan after65c3a34c INCLUDING allrecords needs Claude independently;
older spans remain. Goalactive/progress; no checkbox/status/archive/mint/push/
deploy/release call. No fullCV4/Clout/natural-player/1.0 acceptance.

### Pinned activation range self first-filter

Review by: Codex (implementer, self first-filter). Recorded by: Codex.
Exact inspected range: `65c3a34c..a540dab9`, thirteen paths, including
ea43ef5b predeclaration, both Go observers, TS full sequence comparator,
complete source-pinned corpus and docs/ledger/queue/roadmap/append-only records.
Executed comparisons and failure controls are recorded above; staged source
and all canonical-record diffs inspected directly. Independent e3eab7 verifies
all22 selected source SHA256 identities and six rows; artifact SHA256
`b185b69dc7bebdedc6f69f3267a20d1de4a846451a6527c98fcf105102363f3f`.
gitdiff confirms zero residual runtime/kernel/oldcorpus change. 65fecb proves
both prior HEAD log files are exact byte prefixes. Initial uncommitted roadmap
entry matched a repeated link and was inserted before the previous terminal
entry; diff review caught it, moved it to EOF and verified prefix before commit.

One record clarification in this edge: research queue's former "not Service/
Store proof" now explicitly refers to the EARLIER ordinary observer; the new
separate bounded SQL witness exists, without promoting all persistence breadth.
No observed result or downstream acceptance changes.

Decision: bounded pinned activation/rollback evidence locally validated, NOT
designated approval, full CV4/AC7/P3/Clout or 1.0 acceptance. Source hypotheses
initially failed through instrument clocks/reader boundaries and were disclosed;
no production defect inferred from them. Original27AC6 and all author/content/
environment/release holds remain. Claude must cover ENTIRE newspan after
65c3a34c INCLUDING this record/clarification edge; older independent spans
unchanged. No checkbox/status/archive/mint/push/deploy/release call.

### Action/claim/buff/mode sequence predeclaration

Resumed clean0d866359, previous goal turn PROGRESS: actual activation evidence
and transactional rollback committed, not an unchanged-status turn. Authority,
current implementation and actual claim/clock/compute builders re-derived.
Plan now declares sixteen eight-command sequences covering all four actual
opportunity effects, two wall gaps and two explicit evaluation-mode paths,
forty-eight claim refusals and real contribution/burst/census/event controls.
No experiment run or source mutation yet. Accepted CV3/CV4 only, no production
repair/accumulation contract/owner-body delegation inferred. Prior scopes and
review spans independent, full1.0 objective active.

### Sequence constructor failure — provision instrument assumption

18836/779169 fails before new report exists. Diagnostic50940/a4dc60 identifies
firstbuilding/3114/online-offline-online: permits2.37797654855e-2, purchased101,
provisioned0. Actual engine/catalog code materializes provisions at the pinned
tick, not per second, and this short declared sequence crosses no tick.
Plan correction predeclares exact-zero short controls versus positive long
controls, all16 original sequences retained. No observed output edited, no
source/runtime/balance waiver or corpus published. Corrected run not yet made.

### Sequence constructor failure — attended-clock instrument assumption

4068/be498e and diagnostic66723/122058 both terminate exit2 before corpus
publication. Diagnostic original building buff activated3262/expires5262
survives the long-gap sequence: final5001ms is greater than the declared
5000ms catchup ceiling, so RecordOfflineSpan correctly pauses attended time
across that gap. Actual policy read directly; no runtime defect inferred.
Plan now predeclares exact5000ms final delay and report/observer identity
checks plus prestige source pins. Keep expiry assertion, all16 populations,
128 command attempts and48 refusals. No production/balance/owner-policy edit.
This correction is not yet executed or claimed successful.

### Sequence constructor failure — first-spawn attended-clock setup

95078/a0a26d exits2; diagnostic29978/4df0ff identifies click/3114/online-offline-
online: attainment12/purchased100/earned0 but pendingnil, actual firstspawn6684.
Jumping straight there exceeds5000ms, so this is again the instrument's clock,
not a scheduler defect. No corpus exists. Plan predeclares actual short-gap
online manual preludes, full continuous outputs and explicit per-row/total
command census, retaining all effects/founders/eight core commands/refusals.
No seed filtering, fabricated pending buff, production edit or gate waiver.

### Actual sequence observation — full continuous replay and real failure controls

9974f1d3 at0d866359; corrections eabebf80/83d553ff/eca45c0c precede corrected
experiments.40767/0bee34 publishes FIRST valid report0.637s:16 sequences/136
actual command attempts/48 claim refusals, all original effects/founders/gaps/
modes retained. TS11397/bc1a60 fullclient9640/340skip. Full receipts/ordered
events/pre-poststates/restores agree; actual PR2/bank/burst/buff expiry/permits
and0 short/1440 long provisions bind. Initial tick/clock errors retained.

Discrimination: omitted row80026/5fa473 fails census despite61declarations;
forged claim54716/a5e385 fails full event equality. Compiling actual Go
contribution39951/1e2cc9 fails non-neutral semantic assertion before source
pins. Actual TS omission46736/83ad61 fails10 receipts, while two short click
arms survive because buff expires before their manual command and four Lucky
arms correctly survive buff-only mutation. Actual TS burst91166/64b2f0 fails
ALL16 new receipts plus existing doctrine corpus1. One patch-context rejection
wrote nothing/executed no experiment; corrected target read directly. Exact
runtime/source restore after terminal handles; no compiler failure counted.

Final33205/ee803b client9640/340skip;57401/131d98 types0warnings/errors/full
vet/build213modules713ms/topology13controls pass. Native28013/cb57bf ALL11
research/integration populations/noSkip5.476s, including priorSQL1215/16 and
activation14faults. Cold87688/5dabf9 production66.536s fails ONLYoriginal27AC6;
economy7.401/decimal.267/kernel.169pass. Longer concurrent walltime is not a
performance finding (checks ran concurrently, no causal benchmark inferred).
Process-list read denied; no inferred diagnosis or rerun
from timeout. Source-only CI routing verified; hosted/AMD64 not executed and
historicalRP-131 unchanged. Existing orphan warning not cleanup authority.

6e181d verifies18sourcepins/exact136attempt/48negativecensus, reportSHA
6ec5ed1ba93ab2ebcd6f50f3c7c0aac1458748bfc9823695b3a47b0f5cf81c54.
Prior artifacts/kernel/runtime unchanged. Canonical docs/ledger/currentqueues/
roadmap/log reconciled. Next separately predeclare live Service/Store claim/
burst persistence/retry/history; not all seeds/combo/natural/minted journey,
fullCV3/CV4, accumulation repair or AC6 acceptance. RP-308 delegation unanswered.
ENTIRE span after0d866359 including all records needs Claude independently;
earlier spans and owner/content/environment/release holds remain. Goalactive/
progress; no checkbox/status/archive/mint/push/deploy/release call.

Final non-writing observer58247/be57d8 passes0.618s, complete report byte-exact.
e0de35 checks both append-only logs against0d866359,120applied/16rejected and
parses all full observed JSON. Full staged sources/doc/record diffs inspected;
zero residual runtime mutation, only13 declared paths. Reproduction commands
and explicit writer boundary now retained in sequence-research.md.

### Sequence range self first-filter

Review by: Codex (implementer, self first-filter). Recorded by: Codex.
Exact inspected range: `0d866359..785c088a`, all13 paths and five commits:
9974f1d3/eabebf80/83d553ff/eca45c0c predeclaration/corrections, then actual
observer, full corpus, TS consumer and canonical records785c088a. Complete
source/doc/record diffs read; report checked byte-exact by actual cold producer,
all136 full transitions executed in TS,18 selected source hashes recomputed,
both log prefixes and120/16 outcomes checked separately. Executed failures,
surviving short-click/Lucky probes and all limits recorded above and in dossier.
No existing assertion removed; original27AC6 stays red. Zero production/kernel/
balance/RFC/workflow/oldartifact diff against baseline; no residual mutant.

Decision: bounded continuous action/claim/buff/mode evidence locally validated,
NOT designated approval, fullCV3/CV4/Clout/AC6/CI/1.0 acceptance. Claude must
review ENTIRE newspan after0d866359 INCLUDING this edge; prior spans remain
independent. Next distinct accepted scope is separately predeclared live claim/
burst persistence/retry/history. No owner/body/representation waiver inferred.
One compound Git bookkeeping call was metadata-denied; standalone authorized
add/commit succeeded with no escalation or rewrite. All handles terminal,
goalactive/progress; no checkbox/status/archive/mint/push/deploy/release call.

### Persisted claim/burst predeclaration — 2026-10-07

Resumed cleanf038a467; previous goal turn PROGRESS: committed16 continuous
sequence proofs and discriminating mutations, not a status-only continuation.
AGENTS/RFC-0000/entire accepted Clout contract and relevant binding design/
runtime/save producers re-read. No live handles, no concurrent dirty work.
Plan now predeclares actual16 Service/Store/SQL sequences,168 new logged attempts,
168 later retries/32 conflicts,48 exact six-stage DB faults, full twelve-table
rollback/census/outbox/genesis-to-head replay and real retention deletion using
explicit seven-snapshot diagnostic padding. No experiment or mutation yet.
Accepted CV3/CV4/AC8 only; no inferred owner/body/accumulation waiver, product
repair or CI publication. Proper full1.0 objective active, prior holds unchanged.

### Persisted adapter failure — no product command executed

62521/be1591 fails all16 rows before Handle: canonical replay payload excludes
intent_id by design; the envelope carries it. Diagnostic a38ce5 reads the real
first payload and actual parser. No applied/refused/retry/conflict/fault count
or new shared artifact exists. Plan correction declares exact ID reconstruction
and raw request return, plus body-only conflicts preserving original revision.
No runtime, parser, expected gameplay output, epoch or population changed.

### Persisted outbox instrument failure — events are not receipts

11110/cd71d6 exits2 after16 actual first commands, before any claim/fault/retry:
outbox totals5 or2 differ from claimed receipt-only1. The census query forgot
message_kind; live migrations40/42 also enqueue events. ea4700 confirms actual
schema. Plan now retains exact receipt census, adds full event/outbox bijection
and16 separately counted transactional corruption controls, and targets the
late RECEIPT insertion for the sixth fault. No legitimate event dropped, no
population/production/schema/expected-gameplay-output change. Instrument-only
correction is predeclared before rerun; no completed persistence claim yet.

### Actual persisted claim/burst sequences — 2026-10-07

Corrected1595/ce7c1e passes8.116s:16 actual Handle/Store/Postgres paths,168
logged attempts(120applied/48refused),168 later identical retries,32 body-only
same-ID conflicts,48 exact DB faults and16 corrupt-event-outbox refusals.
Full twelve-table rollback/retry equality, latest-five actual rows, receipt
and event outbox payload/identity, immutable resources and all stored logged
transitions to the exact head pass. Diagnostic seven-snapshot padding is
explicit; actual retention deletes old rows. Runs remain OPEN/log_gap, not
claimed completed histories. Earlier adapter/census instrument errors retained.

Compiling claim omission64184/0f44a2 fails1.868s on evidence admission BEFORE
DB writes; not a claimed successful rollback-stage population. Lookup bypass
90276/b30b4c fails2.681s on16 later retries(revision_conflict/Replay=false),
while all48faults execute. Retention omission72787/2a0349 fails.771s on16
first applies:8actualrows[1,8]not5[4,8]. Restored exactly after terminal handles;
no residual mutation, compiler/source-pin failure or expected-output restamp.

Final broad production Integration33766/1b9130 passes23.538s; focused
67854/18f02a executes ALL12research/integration populations/noSkip6.866s,
newSQL4.70s/exactcensus. Unquoted selector pipe8f326d was a shell launch error,
not product evidence; quoted rerun executes correctly. Client55740/9cf38a
9640/340skip;5047/8facb9 types0/vet/build213/topology13controls pass.
Cold72556/d7fd17 fails ONLYoriginal27AC6production39.331s, economy7.070/
decimal.244/kernel.160pass. Concurrent durations not performance evidence.
NativePG16.15arm64, not hosted/AMD64; CI source discovery is not wholeCIgreen.
Orphan warning not cleanup authority; historicalRP-131 remains unchanged.

sequence-persistence.md retains sources/commands/oracles/failures/limits;
canonical docs/ledger/currentqueues/roadmap/log reconciled. Next distinct work
is reconcile remaining accepted CV4 migration-corpus and CV5/AC3 receipt
obligations before selecting implementation; author/owner contradictions cannot
be silently inferred away. Representation/RP-308 and all previous product/
platform/content/independent-review holds remain. Entire newspan afterf038a467
INCLUDING predeclarations/corrections/records requires Claude. No checkbox/
acceptance/status/archive/mint/push/deploy/release call. Goalactive/progress.

Final5468cb recomputes the unchanged report SHA and all18 selected source
hashes,16distinct row IDs/136planned actions, and verifies both log prefixes
append-only againstf038a467. Complete new test/dossier and all eleven scoped
path diffs inspected; no production/oldartifact/accepted-body residual diff.

### Persisted sequence range self first-filter

Review by: Codex (implementer, self first-filter). Recorded by: Codex.
Exact inspected range: `f038a467..09aeb13a`, all eleven paths/four commits:
fe0e4196 predeclaration,132d2925 request correction,6c3c4173 outbox correction,
then09aeb13a actual SQL test/dossier/canonical records. Full test and records
inspected; live failures/rollback/retries/retention executed above,18 unchanged
source hashes/report SHA and both append-only log prefixes checked separately.
No existing assertion removed, production/oldartifact/kernel/catalog/RFC/copy/
workflow diff or remaining mutant. Original27AC6 remains red, no hidden skip.

Decision: bounded persisted claim/burst/retry/retention evidence locally
validated, NOT designated approval, fullCV4/Clout/AC6/CI/1.0 acceptance.
Claude must independently review ENTIRE newspan afterf038a467 INCLUDING this
record edge; older review spans remain independent. Next reconcile remaining
CV4 migration-corpus and CV5/AC3 receipt obligations against actual contracts.
Owner/author/representation/RP-308/content/platform/release holds unchanged.
All handles terminal; proper full1.0 goalactive/progress. No checkbox/status/
archive/mint/push/deploy/release call or goal completion.

### Exact CV4 Company migration predeclaration — 2026-10-07

Resumed clean7ca728ca/no live handles. Previous goal turn PROGRESS, actual SQL
proof09aeb13a/edge7ca728ca. AcceptedCV4 promises five absent names; corpus9
currently11legacy/4Founder/baseline15. Plan declares real Go/TS Company consumers
inside native migration lane, existing SHA-pinned activation source, exact
20case ratchet, public load/Exit/replay/structural-derived-superset controls,
full outputs and discrimination. No experiment/product change yet. One patch
context rejection wrote nothing (status verified); corrected planning edit
only, no experiment or evidence inferred from that tooling error.

RP-311 filed immediately: actual Go/TS wireSnapshot and committed full receipt
keys526059/4865fb confirm absent CV5-derived axis_stack. GameUI feature arm
does not discharge applied receipt contract. Separate accepted-CV5 lane after
TEST-ONLY corpus work. DG-B/owner-author body holds remain independent.
Goalactive/progress; no checkbox/status/archive/mint/push/deploy/release call.

### Migration instrument failures — exact stage/partial Founder distinction

97094/a32f81 exits2:3Go negatives/11legacy/4Founder pass,2positive cases fail
because adapter tried full save encoding on partial replayed Founder. Direct
public projection, not new production save/export, is predeclared before rerun;
COMPLETE existing Founder output comparison remains mandatory.77033/657ed0
exits2:9645clientpass/1new early-field diagnostic mismatch, actual message
'save v18 fields are not exact', not invented phrase. Plan fixes exact stage
text, retaining5cases/population/source/expected outputs/other oracles. Both
terminal; no product finding or green migration claim from these failed runs.

### Actual shared Company migration corpus — 2026-10-07

Corrected46870/788c51 save.277s and56122/b865b4 client9646/340skip pass.
All five named cases execute real load/Exit/replay and three rejection stages;
full outputs and unchanged-state/minimally corrected companions bind.
Missing Company row04d68c fails Go; pnpm wrapper60661 produced no result and
was explicitly interruptedab5391/exit130, never accepted test evidence. Existing
full Make26854/377217 then fails both Company and Founder census9643/2.
Forged ONLY sourceSHA11623/39b215 Go and67146/e08ed4 TS fail. Compiling real
Go derivation omission80733/c0f5f7 and TS89670/b6657c fail ONLY the exact named
derived-negative, other four Company cases pass. All mutants/corpus corruptions
restored AFTER terminal handles; no production or old artifact diff remains.

Orchestration error: final native42533/dcc783 and27272/45b818 accidentally
overlapped on the SAME disposable DB. Their missing stream/epoch/outbox failures
are invalid final observations, not newly inferred product defects. Reran SAME
commands/populations serially:85332/11f907 native migrations save.536s, then
91919/1600d6 all3actual activation/persistence populations/noSkip5.803s pass.
These include old14fault Exit and16paths/168attempts/168retries/32conflicts/
48faults/16outbox controls. No DB cleanup/population/assertion relaxation.

Final90047/79d652 client9646/340visibleSkip,73307/57ec45 types0/vet/build213/
topology13negative controls pass. Cold18820/eb73c9 production37.414s fails
ONLY original27AC6; save.275/economy6.119/decimal.229/kernel.174pass. Durations
not performance evidence. No fullCI/hosted/AMD64/RP-131 repair. Read-only hash
checkbbc990 used nonexistent artifact_hashes, failed without mutation; corrected
f54834 verifies actual source_sha256 ALL22 files/source SHA and unchanged
11legacy/4Founder/source/baseline15. Existing Make names corrected in dossier
before publishing; no command claimed executed solely from illustrative text.

migration-corpus.md retains complete method/controls/failures/reproduction/limits.
Canon/ledger/currentqueues/roadmap/log reconciled. Next separately predeclare
RP-311 accepted-CV5 applied receipt axis_stack producer repair, not reuse GameUI
feature projection as proof. AC3/DG-B author, R-012 representation/RP-308 and all
owner/content/platform/fullnine-tier/review holds independent. Entire newspan
after7ca728ca including all records needs Claude; no checkbox/status/archive/
mint/push/deploy/release call. All handles terminal; goalactive/progress.

### Company migration range self first-filter

Review by: Codex (implementer, self first-filter). Recorded by: Codex.
Exact inspected range: `7ca728ca..90e102a8`, all18 paths/three commits:
36924143 predeclaration/finding,44c0f8a8 reader correction,90e102a8 tests/corpus/
dossier/canon/records. Complete Go398line/TS124line new readers and ALL existing
reader/data/record diffs inspected. Original11legacy/4Founder/source/baseline15
and ALL22 source hashes verified; both logs append-only29bb13. Runtime/corpus
negative controls execute and restore, final native populations serialized.
No existing assertion/case removed or product/oldartifact/kernel/RFC/copy/CI diff.
Original27AC6 remains red, no newly hidden skip or wholeCI/Clout claim.

Decision: bounded five-case migration proof locally validated, NOT designated
approval or RFC acceptance. Claude must independently review ENTIRE span after
7ca728ca INCLUDING this record edge; older spans remain independent. Next
separately scope RP-311 acceptedCV5 derived receipt object. Author/owner/
representation/RP-308/content/platform/release gates unchanged. Goalactive/
progress; no checkbox/status/archive/mint/push/deploy/release call.

### RP-311 CV5 receipt repair predeclaration — 2026-10-07

Clean3868e3d2, no live handles. Migration proof90e102a8/edge3868e3d2 complete
locally, independent review still needed. Direct acceptedCV5/body and actual
Go/TS ordinary/refresh/Exit producers reread; GameUI has extra attained/intern
rows and cannot discharge this receipt contract. Scope declares private actual
producer/caller errors, kernel166, tests-first missing-object baseline and
independent factor/ownership/legacy controls. Four affected report generators
must execute; allowed delta strictly source pins plus new applied receipt object,
not states/events/payloads/inputs/populations/mints. Company sourceSHA updates
only after actual observation. DG-B/AC3 and all prior holds stay independent.
No experiment/production change yet, no inferred owner amendment or promotion.

Record placement correction:79285a14 matched an earlier identical suffix and
inserted this NEW entry inside an older session.9e5f56 prefix check caught it
before any experiment. This forward correction removes only that new insertion
and appends it here, restoring the ENTIRE3868e3d2 log prefix byte-exact. No old
entry changed or history rewritten; both commits remain in the review range.

### RP-311 actual red baseline

62911/54c115 compiles, all7 Go ordinary/activation cases fail EXACT missing
axis_stack (.375s).97291/1f49b2 client9647pass/30newfail/340visibleSkip:
24 actual ordinary and6 actual old→new Exit producers all omit the object.
Independent literal expected factors/product/closed object fail on absence,
not compilation or stale source pins. Legacy assertions preceding Exit pass;
Go refresh path is not yet reached because ordinary assertion fails first.
Invented test helper caught by source search BEFORE execution and replaced by
actual save.IntentDecision; no compiler failure counted. Both handles terminal.
Next declared producer/signature/error propagation and kernel166 repair.

### RP-311 producer repair / actual observations

69500/6ea5a2 passes7 actual Go ordinary/refresh/Exit paths; expanded63021/e00d25
passes direct zero/one/two/cap/above-cap projection/contrast and typed ordinary/
refresh errors. Above-cap state patched only for projection, not admitted save.
TS additionally admits diagnostic latched inventory at44 with both interns,
literal2.1/1.88/product3.948/not-saturated; not natural progression.
6817/303fcc pre-refresh client60oldreceipt-fail/9617pass/340skip, all31new pass.
Existing COMPLETE receipt oracles retain their discrimination, no field stripping.

49893/198aac actual generators execute3 populations:24logged/6terminalpairs/
6activation; selector's nonexistent sequence name does not prove that population.
81335/99f19e separately executes actual16 action/buff/mode sequences.86d922
full-report comparison to3868e3d2 proves ONLYsource pins and228actual receipt
object/JSON additions changed. ALLstates/events/inputs/bundles/populations/
refusals unchanged; all recorded source hashes verified. Initial bd5619 read
hit Node1MiB buffer on measured2,083,215byte report, not completed proof; full
read succeeds with suitable buffer. SourceSHA patch first failed atomically on
one mistyped old digest, no edit; corrected exact patch updates ONLYCompanySHA.

Corrected8570/1500fb client9678/340skip and16189/b76780 cold save/production/
kernel ALL20 migration and all4observer/newreceipt populations pass. Go omission
58791/62cab1 fails7actual+direct/contrast; TSunowned62204/5f774b fails90total,
new30 fail while census and fully-owned cap survive. TS omission56425/d3d2c5
fails91total including31new object assertions. First Go factor303c5f compiler
failure(unused factors) is NOT discrimination. Corrected compiling50851/54ff4b
forces factor1 and fails ordinary/one/two/cap/above-cap/contrast; empty and6actual
empty new-run Exit cases survive correctly. All mutants restored after terminals,
no artifact generation around a mutant or remaining temporary source changes.

Final17416/07898b client9678/340visibleSkip,88456/92b73c types0/vet/build213/
topology13controls pass. SERIAL14630/07464c native migrations save.591s, then
19214/e3dafe production16.438/gameui.181 Integration pass. Cold48431/b0ab92
save.269/economy6.338/decimal.229/kernel.070 pass; production47.083s ONLYoriginal
27AC6red. PG16.15/arm64, not hosted/AMD64/fullCI or performance claim.
Canonical producer docs/ledger/queues/roadmap/proof reconciled; no save/schema/
migration/event/gameplay/catalog/balance/copy/API schema/CI/accepted-body change.
Kernel166 belongs in same runtime commit. HistoricalRP-131 stays independent.

receipt-projection.md retains complete commands/method/failures/limits. Next
reconcile remaining Clout acceptance and a distinct unblocked accepted lane,
not re-prove already covered boundaries or silently implement unruled numeric/
policy/content contracts. ENTIRE newspan after3868e3d2 including records needs
Claude; older spans independent. AC3/DG-B/representation/RP-308/owner/platform/
full-nine-tier/release holds remain. All handles terminal; goalactive/progress.
No checkbox/status/archive/mint/push/deploy/release call or goal completion.

### Refresh oracle tightened before extra control

Self-inspection catches a potentially vacuous happy refresh assertion: ordinary
receipt already correct, noop could pass that assertion. Seed ONLYthe input
receipt's axis object to valid-shaped empty0, keep actual state8/expected8 and
all existing errors/oracles. Predeclare actual compiling noop refresh mutation;
run focused tests before/after. No output restamp/population/game rule changed.

Normal36679/e7e1d2 passes. Actual compiling refresh noop8006/ad0869 fails
the EARLIER ordinary assertion on input6 not8, before the seeded direct refresh;
negative refresh-error test also fires. Thus the isolated happy assertion was
weak, but calling the whole ordinary test vacuous would be false: real ApplyLogged
already uses refresh AFTER attaining. No claimed independent seeded-direct
failure from that probe. Restored exactly;70636/84e6d9 final focused receipt
tests .211s pass, full vet1e7415 passes again. Production/corpus unchanged by
oracle refinement. Earlier full cold/native/client results retained at their
actual run boundaries, not falsely said to execute this later test refinement.

Final8a3b1a recomputes ALL four report source maps plus migration sourceSHA;
900469 verifies both full log prefixes append-only against3868e3d2 and EXACT
corpus-only Company sourceSHA delta/baseline unchanged. Full runtime/test/canon/
record diffs and whole dossier inspected; generated full outputs independently
compared via86d922, not accepted from summaries. No remaining mutation.
Re-executed root verify-kernel-version91511/294dd4: checkout contract/fixtures
PASS, history verifier exits2 at historical50a3a514 missing bump(RP-131).
This does NOT validate the later dirty range's guard, since history fails first;
same-commit165→166 mirrors and scoped runtime paths are verified separately.
Denied read-only ps82fa6e adds no test evidence, no escalation/cleanup attempted.

### RP-311 full runtime range self first-filter

Review by: Codex (implementer, self first-filter). Recorded by: Codex.
Exact inspected range: `3868e3d2..33d4c534`, all27paths/four commits:
79285a14 predeclaration,e55487fb forward placement correction,03f513ef actual
red tests,33d4c534 runtime/tests/observations/canonical records. All runtime/
test/dossier/record diffs inspected; four complete generated reports compared
against3868e3d2 with ONLYsource pins/new receipt object allowed. Original states/
events/inputs/bundles/populations/cases/baseline retained; source hashes restored.
32a0bd/35e801 verify exact range/path inventory,ec1505 verifies real kernel
165→166 in SAME runtime commit/all3mirrors and append-only log prefixes.

Ordinary/Exit/refresh/cap/owned/order/product/errors and all five compiling
controls execute as recorded. Compiler-only probe excluded, legitimate survivors
retained, no seeded-direct refresh failure falsely claimed. Final focused receipt/
vet supplement follows the last oracle-only refinement; preceding broader cold/
client/native observations remain at their actual boundaries. Original27AC6 and
re-executed historicalRP-131 guard are still RED, not tolerated or bypassed.

Decision: RP-311 locally repaired/bounded compatibility verified, NOT designated
approval, fullCV5/AC3/Clout/CI/1.0 acceptance. Claude must independently inspect
ENTIRE newspan after3868e3d2 INCLUDING this record edge; prior spans independent.
Next reconcile residual accepted Clout criteria and a distinct unblocked accepted
lane; do not repeat evidence or infer unruled numeric/policy/content authority.
All owner/author/representation/RP-308/platform/full-nine-tier/release gates stay.
All handles terminal/clean checkpoint; goalactive/progress. No checkbox/status/
archive/mint/push/deploy/release call or goal completion.

## 2026-10-07 — residual acceptance map and CV10 mask predeclaration

Resumed clean at64f836c2; previous goal turn made implementation progress.
Read actual accepted CV1–CV10/AC1–AC12 and public simulation/first-hour paths.
`acceptance-reconciliation.md` maps all twelve criteria, preserving red AC6,
author DG-B/DG-C, role-binding D-022 and owner mint/copy gates. Next distinct
accepted lane is CV10, not another migration/receipt population.

Plan predeclares test-only public effect masking: four admitted states/four
masks/online+offline, exact literal rates/full state, live no-mask companion and
invalid masks. Actual content-guard omission must fail after normal execution;
restore before cold gates. First-hour axis refusal remains until a separately
predeclared runtime really executes pinned attainment. No AC9/pacing/relevance/
mint/archival/release promotion. All new edges need designated Claude review;
Codex executes implementation and first filter directly, not a substitute gate.

### CV10 diagnostic amendment before the complete repeat probe

28346/597968 unchanged runtime passes16rates/32advances and two invalid-mask
rows (four refused entrypoints). Actual mask-guard omission38263/378c61 exits2
and visibly fails both rates and full states. Full-save diagnostics produce
25,498 output tokens, exceeding the22,000 return limit: not complete cell-count
evidence. Source restored exactly(c0b25b) after terminal handle;30120/5c0de6
normal production+harness refusal/invariant pass, veteb27c6 passes. Replace ONLY
failure messages with cash and full-state hashes, retaining exact byte equality
and every arm; repeat normal then compiling fault with complete diagnostics.
No observation budget, assertion, population or runtime contract is loosened.

### CV10 executed proof and authority correction

31778/145913 unchanged-production repeat PASS. Compiling content mask guard
omission48457/f324ea exits2 with COMPLETE2774token diagnostics: eight literal
rate failures and sixteen full-state online/offline failures, eight rate and
sixteen advance controls legitimately survive. Both unknown/duplicate mask
guards survive as expected. Source restored exactly after terminal handle
(efb8cb). Final44233/4505bb root selected production+harness PASS(.318/2.268s)
including actual source-boundary gates; vet16daeb PASS. Combined AC6 repeat
09ba26 shows27original failures; full cold production22905/787fb3 after35.727s
likewise ONLYoriginal27AC6red. No test removed, tolerance or kernel change.

Forward correction of Codex's41dd8c9f proposed-next note: actual decision
register D-021 and original DG-D still select NEITHER shared served hooks nor
a harness-local observer. Initial acceptance map overlooked that hold. Corrected
current body/queue/plan before ANY runtime implementation; retained original
append-only entry, no inferred delegation. New mask tests do not depend on an
architecture selection and remain inside accepted CV10's simulation primitive.
No product changes or falsified rejection-to-completion claim occurred.

Complete AC1–AC12 reconciliation now identifies the specific proof, missing
authority or empirical failure per row. Next distinct unblocked agent-side lane
is designated review of Claude Garage GS0.3 `301728c8^..301728c8` (14paths);
later remainder/current Codex corrections separate. Read full Garage RFC before
starting. No Clout/AC9/wholeCI/1.0/AT/hosted/AMD64/defaultmint promotion. All
prior review/owner/author/numeric/platform holds remain. Current tests+canon+
records after64f836c2 INCLUDING edges need designated Claude review.

Review by: Codex (implementer-side self first-filter). Recorded by: Codex.
Inspected entire new test160lines, actual live assembly and mask guard, exact
literal expected factors/deltas, full-state/projection immutability and public
mask rejection; independently executed normal/fault/restored populations.
Full first probe output limitation and provisional authority-map mistake
disclosed; neither accepted as complete evidence or implementation authority.
Decision: bounded mask primitive VERIFIED LOCALLY, not designated approval.
All handles terminal, source restored; no checkbox/status/archive/mint/push/
deploy/release call. Long-term full-nine-tier goalactive/progress.

### CV10 full-range review coordinate

Review by: Codex (implementer, self first-filter). Recorded by: Codex.
Exact inspected range: `64f836c2..4509257c`, two commits/all eleven paths:
41dd8c9f predeclaration/map and4509257c test/canon/ledger/queue/records.
Exact complete test diff inspected8f8314; full record/canon diff867027 and
range map/plan/log/ledger a98338. 8a9fbd independently checks exact path
allowlist and unchanged full log prefixes against64f836c2. No runtime/kernel/
oldcorpus/balance/CI byte changed. All normal/probe/final handles terminal,
source restored and cold fullproduction retains only original27AC6failures.
Decision: bounded mask proof locally verified; NOT designated approval.
Claude must inspect ENTIRE newspan after64f836c2 INCLUDING this coordinate
edge; prior receipt/migration/persistence and other ranges remain separate.
Next GarageGS0.3 pending exact-range review, not unruled harness construction.
No checkbox/status/archive/mint/push/deploy or release/goal completion claim.

## 2026-10-07 — meaningful research regressions, not historical checkout equality

Outcome (RP-382, existing CV3/CV4/AC7 and RFC-0000 verification procedure): execute the current
activation, SQL and action/buff/mode observers against their exact retained outputs without
requiring every historical source byte to remain unchanged. Preserve recorded provenance;
no observation regenerated, no production formula/save/kernel/workflow/Make change.

Regression: native activation23807 fails on source/results drift; declared Postgres SQL73203
fails likewise. Existing comparator cannot distinguish those from changed outcomes. New narrow
comparator validates the same source-path population/canonical digests, reports changed sources,
and compares every remaining report byte exactly. Explicit writers retain actual current hashes.
Ten controls exercise unchanged/source-only, wrong receipt including simultaneous source change,
truncation, false acceptance, version and source/digest defects; three corrupt artifacts refuse.
Current report sources restore on every comparison. Test-only helper; no historical-proof claim.

Coverage: native12127 passes activation/comparator. Actual Postgres82087 passes the repaired SQL
and activation, but reveals the same source-only coupling in the action observer; that caller is
included in the coherent repair, not ignored. Final declared real-DB70077 executes all ten Axis
research tests plus comparator, PASS1.948s: SQL1215/16, six activation sequences, action16/136
commands/48 refusals, all other retained observers unchanged. TS56249 passes772 cases/four
files; vet88678 passes. Cold full real-DB production47199 takes83.795s and fails ONLY the original
27 AC6 partition cases. This is linux/arm64/Postgres16.15, not amd64/hosted CI. Two initial native
selector mistakes execute no intended population (Make's dollar handling/unquoted regex pipe);
corrected quoted selector executes the named tests. No green credit from those attempts.

Handoff: Review by: Codex (implementer first filter); Recorded by: Codex. Exact range after
`cd0c8581` through this repair's commit needs designated review; older spans remain separate.
Full diff inspected; historical artifacts and production bytes unchanged. Records/docs identify
local correction without hiding current hosted red, RP-307/RP-131, numerical-policy/owner/author
holds or missing late-game/release proof. All process handles terminal before records/commit.
Next: remaining accepted integration/review work; numerical repair still needs its compatible
accepted representation/episode contract. No acceptance checkbox, mint, archive, push or release.

## 2026-10-08 — show the current PR multiplier instead of a rounded resource amount

Outcome: repair RP-387 under CV9's current-factor/product rendering contract. The Desk panel
used Amount's ruled below-1000 integer formatting, rendering all three supplied values as `1`.
Axis-only output now preserves the validated canonical coefficient with text decimal placement;
when expansion needs extra zeros it retains canonical notation. No factor calculation, general
Amount/notation-policy change, server math, balance, copy, content mint or CI change.

Regression: new independent literal assertions against the Go projector golden and an
authoritative fixture refresh fail in both Chromium/WebKit (65716: four failures/four passes):
`[1,1,1]` instead of `[1.2,1.2,1.16]` / `[1.3,1.3,1.24]`. The initial sandbox attempt cannot
listen and executes nothing. Initial Decimal.toString replacement passes these small cases,
but direct diagnostic2e997f exposes `1.005e1 -> 10.049999999999999` and
`1.11111111111e1 -> 11.111111111100001`; discarded before commit. Exact text formatting retains
both values, small effects, neutral/cap examples and large canonical exponents; invalid inputs
reject. No new numeric engine or change to the shared numeric module.

Coverage: Node12171 passes52 formatter/decoder/formula/unchanged Amount goldens. Native4106
passes14 Axis/actual shell browser checks, including literal factors at inputs8/12/44,
progressbar associations, refresh, absence, axe and unchanged Amount cadence/cap behavior.
An earlier selector28697 included a nonexistent Amount-browser filename: only eight Axis
executions ran, not Amount coverage; final4106 uses the actual shell-browser file. Types,
production build215 modules and UI boundaries61248 pass. Final indentation-only test edit
changes no assertions. All verification handles terminal before records.

Review by: Codex (implementer diff/first filter). Recorded by: Codex. Exact new range after
`48a7e5ed` through this batch's final code/test/docs/record commit needs designated review.
No AC/archival/served-epoch/AT/full-CI/1.0 completion claim. RP-307's original27 arithmetic
failures and the unaccepted accumulation/episode contract remain; this is fixture-panel proof.
Next: reuse the existing declared real-service journey and axis fixture for PR purchase,
authoritative multiplier refresh and reload persistence, without claiming a release content mint.

## 2026-10-08 — PR purchase reaches the actual built browser and database

Outcome under CV9/AC11 and CI Baseline D2: reuse the existing Cosmetic service/built-client driver with explicit
`--axis-stack` / root `make test-clout-composed`, selecting the existing unminted economy fixture.
Only test cash is seeded. The actual DOM gate earns2; native Enter Buy Max earns12 with the
exact four run IDs; native Enter PR Buy emits one exact Company intent and a bound applied
receipt at the next revision. Snapshot and DOM progress/factors/product match literal expected
values at input0/2/12, product1→1.3. Reload retains Founder/run/ownership. Direct Postgres checks
the exact Company19 head, attainment set/score, owned upgrade and unique purchase event.
No production, schema, numeric, balance, copy, content mint or Actions job/timeout change.
The existing composed CI target retains its ordinary main/Cosmetic populations and adds this
fixture variant; `make test-clout-composed` is the focused local selector, not a manual-only gate.

Discrimination: healthy65767 passes the selected PR and existing Cosmetic/care populations.
AC11's compiling producer-severing17563 then exits2 with `real snapshot must carry axis_stack`.
Product source restored before subsequent runs; Git hash stays exactly
`f88eee3d271f85a73e7c28d9622c3846900e4387` and its diff is empty. Restored72963 passes.
Unknown CLI7237ce exits1 before test DB/temp setup; syntax and CI topology/13 negatives pass.
Initial sandbox Docker denial executes no test; the permitted root runs own declared test ports.
Test DB capacity preflight reports105.5MiB used/7.7GiB available on its tmpfs; no container pruning.

Stronger63825 passes PR/SQL but fails the existing care projection equality: RP-388 stays OPEN.
Public-coordinate diagnostic added without weakening the check; actual helper healthy/four
mismatch controls d8a8d4 pass and omit a private sentinel. Diagnostic91893 passes the complete
selected fixture (94 requests/7.699s), not a unique cause or reliability repair. Ordinary
`make test-game-ui-composed`99343 passes its actual CI population: asset observer5, refresh8,
all7 persisted parent tests, production-built main gameplay/Pitch/early endings/recovery and
default Cosmetic/care/reloads (main101 requests; Cosmetic73/6.664s). No hosted/Linux-amd64/full
CI/AT/default PR content proof substituted.

Final expanded `make test-game-ui-composed`70861 is RED in its unchanged first driver, BEFORE
either Cosmetic variant. Asset5, refresh8 and all7 persisted parents pass; actual locked Pitch
409 has the expected not_eligible/fiscal_unlock_required, but its visible notice times out at
the unchanged30s bound. No Clout failure or new-stage execution is inferred. RP-389 records
this independently of earlier ordinary99343 PASS. Failed DB preserved before reset: latest
Company36/v18/T1 has an Acquihire offer; spawn and gate events share revision36. The first
read-only query used wrong exit_offer/id field names and provides no valid offer finding;
corrected0bbbb6 uses offer_state/occurred_at/event_id. No failed browser DOM capture remains,
so the offer is a lead, not unique cause. Local start22:53:45→terminal observation22:54:59 UTC
is an incomplete74s attempt, not a completed latency observation or hosted-budget proof.
Syntax, full diff and final CI topology/13 negatives pass; all process handles terminal.

Review by: Codex (implementer diff/first filter). Recorded by: Codex. Exact new range after
`00a6f43d` through this batch's final driver/Make/docs/records commit needs designated review;
the prior display correction and older Clout spans remain separate. No checkbox, archival,
full CV9/Clout/1.0, mint or publication promotion. RP-307's numeric red and unaccepted repair
contract, harness binding and remaining projected-factor/hint/Codex/notice obligations remain.
Next: diagnose RP-389's actual locked-Pitch notice/lifecycle boundary and retain the failed DB
until useful evidence is collected; then verify the expanded target without weakening its gates.
Consolidate this fixture proof with the display correction for cross-party review; genuine
numeric/harness/content/UI contracts remain explicit, not defaulted by a passing fixture.

## 2026-10-08 — Existing producer context survives real jsonb

Baseline `3d1171b3`; predeclared in this plan before measurement. Test-only
R-012 follow-up, not a repetition of the completed scalar-anchor SQL study.
`TestAxisRateContextSQLIntegrationResearch` executes all64 admitted Company
profiles through Encode -> Postgres16.15 jsonb -> existing RestoreState ->
foundation validation -> actual Rates/Evaluate. All canonical restored states,
raw mantissa bits/exponents, full evaluation results and final states are exact;
cash equals original raw AccrueConstant. All64 admitted changed-count controls
produce different context, rates and cash. SQL128 rows are transaction-local;
explicit rollback succeeds. Actual Go runtime linux/arm64, not AMD64 evidence.

Commands executed cold from root:

- `make test-save-integration SAVE_TEST_PACKAGES=./production SAVE_TEST_FLAGS='-run TestAxisRateContextSQLIntegrationResearch -v'`: PASS,0.161s.
- `make test-go GO_PACKAGES=./production GO_TEST_FLAGS='-run TestAxisContextRawRateComparator -v -count=1'`: PASS; exact companion and six negatives, including a low-bit difference hidden by identical canonical rate text.
- `make test-save-integration SAVE_TEST_PACKAGES=./production SAVE_TEST_FLAGS='-run TestAxis.*Research'`: PASS,2.047s; related research regression population unchanged.
- `make vet` and `git diff --check`: PASS.

Initial ordinary host run explicitly skipped as NOT_EXECUTED without DB; it is
not SQL proof. Self-inspection fixed helper closures to receive the subtest T
before SQL execution. No production/save fields/kernel/balance/content/CI or
old source-pinned artifact bytes changed. No new frozen corpus/framework.

Conclusion: existing context reconstruction survives SQL for this bounded
population. It does not adopt a conserved representation, settle offline episode
meaning or repair RP-307's original27 partition failures. Store/replay, served
TS producer, natural progression and whole-feature/release acceptance are not
claimed. Next R-012 work must address the remaining representation/policy question,
not repeat these completed fidelity waves. D-021/D-022 and author holds remain.

Review by: Codex (implementer first filter). Recorded by: Codex. Exact new range
after `3d1171b3` through this containing test/record commit needs designated
cross-party review; earlier ranges remain separate. No acceptance/archive/push.

## 2026-10-08 — Route completed research into an accounting contract

Baseline `10d94e70`. Current cold checks:

- `make test-go GO_PACKAGES='./production' GO_TEST_FLAGS='-count=1 -run TestAxisAccrualPartitionProperty'`:
  RED, 27 of the original 128 seeded cases fail (exit 2). Population, cuts and assertions unchanged.
- `make test-go GO_PACKAGES='./production' GO_TEST_FLAGS='-count=1 -run TestAxisAttainmentChangesOnlyFutureAccrual'`:
  PASS, all 64 timing arms. This does not repair partition accounting.

Completed carry, anchor, raw-rate, policy, logged-sequence and SQL evidence now routes to
`rfc/production-accrual-conservation.md`, indexed as a draft. It identifies a frozen-producer
origin direction and explicitly leaves the exact state fields, boundary/receipt algorithm,
versioned activation and operating bounds unresolved. Rounded rate serialization is rejected
by the existing five payout differences; resetting at every provision boundary is not assumed
equivalent to the old one-shot calculation. No completed experiment or artifact was regenerated.

RP-308's internal-chunks versus distinct-reconnect distinction was put to Marco; no answer or
body-reconciliation delegation is recorded. RP-307 remains red. No accepted body, runtime,
schema, balance, kernel, content, CI or completion checkbox changed. Next: resolve those actual
contract questions before implementation; do not repeat completed SQL fidelity studies.

Review by: Codex (draft/diff self-inspection). Recorded by: Codex. The new draft range after
`10d94e70` through this containing commit needs designated cross-party review before acceptance;
it does not approve earlier implementation ranges. Documentation links and `git diff --check`
verified before commit. No archival, push or release claim.

## 2026-10-08 — Actual evaluator/ledger origin feasibility

Baseline `32f7518c`; bounded population declared in the owning plan before execution.
New test-only `axis_origin_projection_test.go` restores an unchanged whole-Company origin,
uses the actual contributions/evaluator, and advances visible balances only through
`Ledger.ApplyAccrual`. Every cut checks exact targets, reproducible receipt deltas, the
five non-ledger evaluation outputs, full encoded Company equality and ordinary restoration.
All 128 original seeded profiles pass this candidate; the ordinary rebased control retains
27 differences. Twelve provider/burst/mode intersections execute six cuts, positive permits
and literal provision counts/carry. Four single-episode offline cases check cap/bank floor
and saturation; two debit/re-anchor cases require literal payouts without capped overflow.

Executed root checks: focused `make test-go GO_PACKAGES='./production'
GO_TEST_FLAGS='-count=1 -run TestAxisOriginProjection -v'` PASS (0.676s, all 146 profiles);
`make vet GO_PACKAGES='./production'` PASS. Finished package
`make test-go GO_PACKAGES='./production' GO_TEST_FLAGS='-count=1'` remains RED (36.422s,
exit 2), with only the original 27 partition failures reported. No regression is skipped,
tolerated or replaced. `git diff --check` passes. This host run is not SQL or hosted CI proof.

The draft now states the bounded algorithm/write set and a material integration constraint:
pre-advancing before unchanged ApplyLogged would suppress its elapsed-driven accrual hooks;
running hooks on every cut would instead add semantic events. Neither is an allowed shortcut.
The accepted repair must bind logical settlement results/once-only hooks and the closed
persisted mutation boundary. The experiment is Go-only, with no external/active-buff inputs;
whole-Company JSON is diagnostic, not a proposed production schema or storage bound. No
production, numeric, save/replay version, content, CI or recorded-artifact bytes changed.

Next: resolve the minimal durable context and settlement-result contract, then obtain
acceptance before implementing it. RP-308 adoption remains unanswered; RP-307 remains red.
Review by: Codex (test/draft first filter). Recorded by: Codex. Exact new range after
`32f7518c` through this containing commit needs designated cross-party review; earlier
ranges remain independent. No feature completion, archival, push or release claim.

## 2026-10-08 — Correct the assumed persistence requirement

Baseline `6559a3d5`. Source tracing finds GameUISnapshot projects committed states without
committing accrual; ordinary Handle, Exit, offline catch-up and suppression belong to logged
mutation boundaries. Simulation and content diagnostics operate on hypothetical states.
This does not establish that a conservation origin must span authorized saved heads.
Codex's own draft assertion that durable state was necessarily required was unsupported.
The body and current next-action pointer now distinguish operation-local versus durable
origins, pending author interval/settlement resolution. No accepted contract is narrowed;
the original partition property and all 27 failures remain, and no version/field is adopted.

Under existing GU-C9/read-only projection and Reputation AC7, the actual persisted public
projection test now performs four extra reads before the same later Exit. Each checks its
advancing display time, unchanged revisions/full resource and generator views; the existing
observer binds unchanged full rows in twelve tables. The later Exit/continuation and exact
retry checks still execute. This is the real projector/Store/Postgres path with candidate
Reputation content, not every public HTTP/read path or a Clout production repair.

Executed root checks: focused real-DB projection PASS (0.199s); final
`make test-save-integration SAVE_TEST_PACKAGES='./gameui' SAVE_TEST_FLAGS='-run Integration -v'`
PASS (0.404s, all three named populations including four Fiscal arms, no skips);
`make test-go GO_PACKAGES='./gameui' GO_TEST_FLAGS='-count=1'` PASS (0.430s);
`make vet GO_PACKAGES='./gameui'` and diff checks PASS. Host tests alone are not DB proof;
the declared Postgres run is. No production, kernel, save, balance, content or CI bytes changed.

Next: settle the accounting origin's required lifetime before choosing a storage format;
prove action/hook integration for the selected contract. Do not infer new persistence from
the scalar defect or weaken accepted invariance to avoid it. RP-308 remains unanswered.
Review by: Codex (test/draft first filter). Recorded by: Codex. Exact new range after
`6559a3d5` through this containing commit needs designated cross-party review; the earlier
origin experiment/draft and all older implementation ranges remain independent. No archival,
feature/release promotion or push.

## 2026-10-08 — CV9.4 native PR Codex help (RP-412)

Baseline `143174da`. Accepted CV9.4 already requires `?` to open `codex.axis_stack`;
the panel and catalog lacked both. Native details/summary now provides keyboard
open/close, visible focus, a 44px target and wrapping content. Snapshot replacement
preserves the open state; help submits no gameplay command. The accepted Garage
manifest's 2026-09-25 placeholder allowance owns the explicitly `PENDING OWNER COPY`
label/body. No final prose is adopted. Generated catalog/types/Go keys and deployment
copy hash change together; constants hash, kernel, balance, save and CI are unchanged.

Verification: old production fails all12 new keyboard cases (31733), with existing
cases deselected. Final `make test-browser-focused BROWSER_TEST_FLAGS='test/axis-stack-browser.test.ts
--project=chromium --project=webkit'` passes20 (19890): all three current eras,
Enter/Space, Tab/Shift-Tab, 320px reflow, fresh-revision snapshot preservation,
axe serious/critical floor and zero intents. `make test-clout-composed` passes
(30766): actual built-client keyboard help, then PR purchase/render/reload/SQL
and the existing Cosmetic/adoption/care journey. No HTTP or authoritative action
was replaced with a fixture callback. Node driver syntax passes.

Intermediate evidence retained: a test-side button-role assumption made two
composed attempts fail (15001/62609) and12 mounted cases fail (86824). The native
summary existed; the accessibility snapshot was a group. [W3C's summary rule](https://www.w3.org/WAI/standards-guidelines/act/rules/2t702h/)
explicitly distinguishes native summary from button-role queries. Tests now select
the native summary and bind its label plus real keyboard outcome; no explicit
role override, retry, budget increase or changed product behavior hides that error.
The initial unsorted catalog also failed generation and was sorted before rerun.

Finished source types/build/boundary/copy checks pass (32257); final test-source
typecheck also passes (85483, zero Svelte errors/warnings). Generation verifies
661 keys and the exact copy hash, retaining614 visible orphan warnings. Current
constants hash is unchanged. These are bounded local checks, not full/hosted CI,
manual assistive-technology acceptance, adopted Codex content, projected factors,
once-only hint, coalesced notices, content mint or whole Clout/1.0 completion.
RP-307's27 partition failures and the other owner/author holds remain separate.

Review by: Codex (implementer first filter). Recorded by: Codex. Exact new range
`143174da` exclusive through this containing implementation/test/record commit
requires designated cross-party review; earlier ranges remain independent.
No completion checkbox, acceptance/archive, push or release claim.

## 2026-10-08 — CV9.5 coalesced re-attainment notices (RP-425)

Baseline `24cefcd5`. The accepted producer already emits Company
`achievement_reattained.v1`; the browser decoder ignored it. The exact decoder
now admits its three fields and rejects malformed/Founder payloads through the
existing resync path. The host announces once per Company stream/revision,
ignores trailing previous-run events, preserves focus and snapshot authority,
and updates the existing live DOM even for identical text at later transitions.
No timer, factor arithmetic or optimistic state is added. The CV8 notice key is
explicitly `PENDING OWNER COPY`; generated outputs and deployment copy hash move
together, with constants hash/kernel/save/balance/CI unchanged.

Executed evidence: old decoder population884822 fails14 cases (two controls pass),
and old mounted host33142 fails all four new native cases. Final selected decoder/
runtime Node population15238 passes69, including actual simulated reconnect replay,
distinct same-revision events and malformed-successor resync. Final affected root
browser population76730 passes1014 Chromium/WebKit cases, with four existing
performance-only skips. New cases bind actual live-region mutations, same-transition
coalescing, identical later text, new Company/reused revision, previous-run refusal,
ordinary earned notices, unchanged axis values/focus and zero intents at320/1280px.
Types (zero errors/warnings), production build, copy/manifest and client-boundary
checks96798 pass; driver syntax and diff checks pass. Copy verification retains621
visible orphan warnings rather than claiming catalog adoption.

`make test-clout-composed`15804 passes against real Postgres and WebSocket with the
production-built browser. Its controlled setup now seeds four prior-earned Founder
IDs/score12 in addition to cash, never Company attainment or events. Native first
gate and Buy Max generate four actual stored re-attainment events in two revisions;
the browser records exactly two notice updates. This is not a played prior run.
Existing PR help/purchase/receipt/render/reload/SQL and Cosmetic/adoption/care remain.
Final whole `make test-game-ui-composed`84866 exits0:35 driver controls, real-DB
refresh/Fiscal/content/board/composition populations, production-byte early journey,
both normal/Clout cosmetic-care journeys and both intended no-payment failures.
Expected reconnect proxy resets are visible; no assertion, retry or deadline is
weakened. Focused and whole fixture passes are not full/hosted CI or reliability proof.

Next: remaining CV9 projected factors/contextual hint and owner content, alongside
the separately held accounting/harness contracts. RP-307's27 partition failures,
offline-episode meaning, mint, manual AT and whole Clout acceptance remain open.
Review by: Codex (implementer first filter). Recorded by: Codex. Exact new range
`24cefcd5` exclusive through this containing implementation/test/record commit
requires designated cross-party review; earlier ranges remain independent.
No completion checkbox, acceptance/archive, push or release claim.

## 2026-10-08 — Put PR progress/factors beside the purchase (RP-426)

Baseline `627e0f9e`. CV9.1 requires the existing upgrade rows to carry Intern
progress/current factors; they instead lived in a separate panel. Actual PR
purchase cards now render the server-supplied factor and unowned unlock bar.
Bars name their own card heading; Buy descriptions include factor/progress plus
existing refusal/pending reasons. Owned cards retain factor/owned text without
unlock progress. The shared panel keeps one product/formula/cap/help readout,
not duplicate Intern rows. No new math, copy, API, save, kernel or CI behavior.

Regression65028 fails both unchanged-source Chromium widths at the missing card
factor. Initial46155 incorrectly selected the price output and used two wrong
copy literals; it also ended with a browser-connection error after two failures,
so is not credited as a complete baseline. Corrected selectors target the factor
inside the actual purchase card, not arbitrary outputs. Fixture ownership now
agrees across upgrade and Intern arms instead of hardcoding the first as owned.
Final5086 passes all36 Clout Chromium/WebKit cases; final affected78060 passes236
with four existing separate-performance skips. New cases cover all three current
eras at320/1280px, independent literal factors, actual card/name/description
bindings, refresh, non-PR isolation, no duplicate readout and axe/reflow floors.
Existing fractional precision, progress refresh, help and notice checks remain.
Types (zero errors/warnings), production build and boundaries12281 pass; copy/
manifest5638 pass byte-unchanged, retaining621 orphan warnings. Diff checks pass.

Real `make test-clout-composed`15331 exits0: production-built purchase cards bind
each actual projected factor/progress, native PR purchase and owned reload retain
exact SQL checks, real re-attainment notices and existing Cosmetic/care pass.
No new whole/hosted CI, natural pacing, manual AT, content or feature acceptance
claim follows. RP-307 and all earlier holds/review ranges remain independent.

RP-427 names two genuine missing definitions: projected-factor target and the
hint's once-only lifetime. Owner questions delivered, not answered/adopted; the
decision queue records options. Do not invent projection inputs or persistence.
Next: resolve those definitions while continuing other accepted construction.
Review by: Codex (implementer first filter). Recorded by: Codex. Exact new range
`627e0f9e` exclusive through this containing implementation/test/record commit
requires designated cross-party review. No acceptance/archive or push.

## 2026-10-08 — RP-462 producer observation freshness

At `1fe602e2`, the broader client check55351 fails exactly the R-012 producer
source binding (10318 pass/977 explicit skips); ordinary Go75894 independently
fails `research source/results drifted`. Cosmetic7f94671a changed the pinned
save decoder; no diagnostic census source changed in the intervening Typer fix.

Re-executed the existing predeclared64-profile observation, with no changed
population, thresholds, instrument or production source:
`make test-go GO_PACKAGES='./production'
GO_TEST_FLAGS='-run TestAxisRateSerializationResearch -v -count=1'
UPDATE_RATE_RESEARCH=1`43429 passes. Structural comparison to committed artifact
confirms only `source_sha256[server/save/state.go]` changes. All64 rows, fourteen
source names, fixture hash,64 precision changes/five payout differences and
`NOT_PROVEN: production AC6 remains red` are unchanged. No manual hash edit.

Ordinary cold producer replay/vet1991 passes. Full `make test-client`22451 exits0:
10319 pass/977 explicit skips. Declared `make test-go-ci
CI_TEST_PACKAGES='./production'
CI_TEST_FLAGS='-run "TestAxisRateSerializationResearch|TestTyperAPIStartIntegrationUsesPinnedTierAndExit" -v'`
76993 passes both named populations on Linux-amd64/real Postgres, none skipped.
This is affected CI-environment evidence, not full server/hosted CI, a new SQL
measurement, natural progression or closure of RP-307/AC6. Diff check passes.

Review by: Codex (measurement implementer first filter). Recorded by: Codex.
Exact new artifact/dossier/record range `1fe602e2` exclusive through this
containing commit requires Claude's designated review. All earlier review debt
remains. No runtime, kernel, API, balance, CI-workflow, acceptance or release
change. Next safe work: remaining accepted integrations/reviews; do not repeat
the completed64-profile study or infer the held numeric/owner rulings.

## 2026-10-08 — RP-462 policy-observation sibling reconciled

During Garden closeout at `8fd83b40`, cold full production80654 reproduces the
27 known RP-307 partition failures plus policy-observation source drift. The
44-profile sibling still pins the old save decoder, independently of the earlier
64-profile repair. Re-executed its existing producer with
`make test-go GO_PACKAGES='./production'
GO_TEST_FLAGS='-run TestAxisPolicyBoundaryResearch -v -count=1'
UPDATE_POLICY_RESEARCH=1` (terminal exit0), then the same cold command without
the update switch (exit0). All44 full-state/restore pairs execute,18 differ as
before, and AC6 remains NOT_PROVEN. The entire artifact diff is one source hash;
all measurements, population, thresholds and other provenance bytes are unchanged.
No manual hash replacement or runtime/policy/test change. The earlier full-package
red is retained, not relabelled green by the focused replay. No client consumes
this policy artifact; no unrelated browser run was needed for its refresh.

Review by: Codex (measurement implementer first filter). Recorded by: Codex.
New range begins after `8fd83b40`, requires designated review, and closes only
this stale-observation sibling. RP-307/RP-308 and prior review debt remain open.

## 2026-10-12 — Origin difference-write fails the lower admitted domain (RP-488)

Baseline `237c2f17`. The plan predeclares20 actual Company projection pairs at five
admitted generator-rate exponents, two modes and1000→1001ms /2000→2001ms. The
hash-bound in-memory fixture keeps its existing1.001 frozen multiplier and removes
the owned PR upgrade; zero initial cash isolates the producer. Current Company
admission, pinned catalog loading, codec restore, Evaluate and ledger all execute.
This is not adopted content or a change to the numeric domain.

Cold observation87060 completes all20: nine non-state differences/refused commits,
ten exact commits and one unchanged pair; zero wrong-target commits. For the floor
online1000→1001ms pair, both `1.001e-8999999999999999` and
`1.002001e-8999999999999999` are admitted targets, but `target.Sub(before)` is NaN
in Go. Actual `Ledger.ApplyAccrual` refuses with `ErrInvalidTransaction`, preserving
the complete encoded state and an empty receipt. The retained test now requires
the exact census, literal primary pair, refusal cause/non-state delta, untouched
state and exact positive receipt re-additions. An always-refusing ledger or wrong
target cannot pass the observation.

Five literal client arithmetic cases14040 pass: floor-online/near-floor-offline
refusals plus near-floor-online and ordinary online/offline exact controls.
These use actual numeric helpers but do not claim a TS Company evaluator, ledger
or SQL boundary. The original anchor research source/artifacts are untouched.
Affected client91424 passes1208 tests in four files: the new research, existing
anchor/partition research and production accrual vectors. Types32235 pass with
zero errors/warnings. Final cold Go origin population35365 and vet pass, including
all146 prior candidate profiles and the added20-cell census. Earlier verbose83341
records the same outcomes and original27 rebasing control.

Full cold `make test-go GO_PACKAGES='./production' GO_TEST_FLAGS='-count=1'`51740
exits2 (Go exits1) with precisely the original27 RP-307 partition divergences.
It is RED, not a full-package pass or proof DB prerequisites ran. No runtime,
kernel, save, balance, rounding, episode, receipt or CI-workflow byte changes.
No browser/SQL/Linux/full or hosted CI claim is made for this test-only extension.

The decision draft's C2 body now explicitly rejects the difference-write candidate
as a general algorithm. Resolving a viable ledger/receipt algorithm is technical
work, not an owner waiver. Episode/settlement and compatibility decisions still
require their proper authority; neither this negative result nor the prior146
successes adopts a production format or closes Clout AC6. Next: a viable accounting
contract that retains these counterexamples and the unchanged numeric/receipt laws.

Review by: Codex (research implementer first filter). Recorded by: Codex.
Exact test/research/draft/record range `237c2f17` exclusive through this containing
commit requires designated cross-party review. Earlier review debt remains;
no acceptance, archival, push, deployment or release claim.

## 2026-10-12 — Floor increment constraint and deferred settlement (R-012 / RP-488)

Baseline `8e7eecc6`; question/population predeclared in the owning plan. Extended
the existing test-only Company-origin instrument and client arithmetic controls.
Actual ledger zero/minimum/three larger probes retain literal results: zero stays
below the floor target, minimum positive accrual already overshoots it, receipts
reproduce the real committed results, and sub-floor numeric values refuse. The
normalized-domain bound explains the specific pair's unavailable positive
increment; this is not an exhaustive library monotonicity or served-content claim.

Separately all20 retained lower-domain cells project scratch copies while leaving
the original Company untouched, then settle once through the actual ledger.
Complete final encodings match one-shot Evaluate, with reproducible receipts.
Origins have zero cash; arbitrary origins, saved intermediate heads, ApplyLogged/
hooks/actions/SQL/TS Company integration and production field bounds remain open.
The draft now requires a clear scratch-cut/authorized-settlement distinction,
without adopting it or redefining AC6. No production/numeric/receipt/domain,
kernel/save, balance, episode, authored copy or CI-workflow bytes changed.

Cold `make test-go GO_PACKAGES='./production'
GO_TEST_FLAGS='-count=1 -run TestAxisOrigin -v'`91799 passes every existing and
new origin population, retaining the original27 rebasing discrepancies and the
20-cell refusal/exact census. Four affected client files pass1209 checks; types
48711 and affected vet pass. Initial client control incorrectly expected RangeError
from canonical parsing; the existing parser's SyntaxError was confirmed from source
and corrected in the test only. Its five earlier cases already passed; no assertion
or production parser weakened. Final client suite passes all six new/existing arms.

Full cold production48338 exits2 (Go exit1) with exactly the original27 RP-307
partition failures, not new failures or a green package. No DB/browser/Linux/
hosted/full-CI execution is claimed. All handles terminal, formatting/diff checked.
Review by: Codex (research implementer first filter). Recorded by: Codex. Exact
new test/draft/ledger/record range starts after `8e7eecc6` through this containing
commit; designated review pending, alongside earlier debt. No acceptance, archive,
push or release. Next: settle the interval/authorized-head contract, then prove a
complete candidate through the actual closed transition boundary—not another
search for a sub-floor delta that the current domain cannot represent.

## 2026-10-12 — Scratch cuts through the actual logged purchase boundary (R-012)

Question/population predeclared in the owning plan at `c2bd064e`. All128 original
seed307/end/cut/mode profiles now exercise actual logged purchases from admitted
nonzero-cash origins. Two internal projections leave the committed pre-state untouched;
the final purchase reproduces the ordinary full Company, exact receipt and ordered events.
Literal generator/reattainment and positive prestige lifetime accrual gates prevent cash
alone from standing in for the closed transition. In all128 deliberately pre-advanced
controls, the real purchase still applies and cash matches, but prestige lifetime accrual
is lost. The draft C3 body now cites that bounded integration finding, not an implemented
accounting algorithm or adoption of scratch-cut/settlement meaning.

Initial test-only compilation referred to a nonexistent Ledger.MustBalance helper, so no
experiment executed; corrected to existing canonical snapshot/parse helpers. Focused89108
then passes all128 profiles. Full cold origin78165, affected vet and declared Linux-amd64
focused37014 pass. The Linux test is in-memory, not a new SQL persistence population.
Full cold production58475 exits2 (Go exit1) with precisely the original27 RP-307 partition
failures; no new failure, skip/tolerance repair or whole-package green claim. All handles
completed; source formatted and diff inspected. No client change/TS parity, full hook
population, lower-floor purchase, SQL/restart, production codec, hosted CI or release proof.

Review by: Codex (research implementer first filter). Recorded by: Codex. Exact new
test/draft/plan/ledger/log range starts after `c2bd064e` through this containing commit and
needs Claude review; no self-approval or waiver of older debt. Next: complete producer-context/
compatibility specification and the remaining closed-hook boundaries under the unresolved
interval/authorized-head contract, not repeating scalar studies or pre-advancing live state.
No production, balance, save/kernel version, policy, accepted criterion, archival or push change.

## 2026-10-12 — External producer context through a logged manual action (R-012)

Predeclared at `d9359e88` in the owning plan. Extended the existing origin instrument,
not production: actual Commons and frozen Fiscal producer outputs plus the declared neutral
Guild membership signal, online/offline, six cells. The existing replay-accrual codec retains
every resource's ordered raw Decimal rates. Four scratch projections leave the committed
Company untouched; the subsequent actual logged manual action matches the complete ordinary
state, exact receipt and ordered events. Positive gates require Prestige accrual and, where
present, the Commons sample and positive Guild tithe. Every omission control changes projected
cash; every pre-advanced control keeps cash but loses those hooks. Completed-cell census is
incremented only after all assertions, preventing invalid arms from printing a success census.

The first120001ms setup (75317) refused all six before action at the existing active-play
due-transition ceiling. It is invalid accounting evidence, not a conservation success or
authority to raise the ceiling. That observation remains recorded. The plan explicitly
narrows this experiment to an admitted2001ms interval with cuts1/1000/2000/2001; it makes
no longer-interval claim. Focused54294 then passes all six, followed by exact raw-rate gates.

Final execution: cold native `make test-go vet GO_PACKAGES='./production'
GO_TEST_FLAGS='-count=1 -run TestAxisOrigin'`7471 passes all old/new origin research and vet.
Declared Linux-amd64 focused29525 passes all six new cells verbosely. It is in-memory
research, not SQL persistence proof merely because the Docker lane supplies a DB URL.
Full cold `make test-go GO_PACKAGES='./production' GO_TEST_FLAGS='-count=1'`68371 remains
RED: exactly the original27 `TestAxisAccrualPartitionProperty` failures, 38.455s; no new
failure or AC6 repair. Native DB prerequisites skip. No TS/browser/hosted/full-CI claim.
Formatting and diff checks pass; all handles terminal.

The Codex-authored decision draft now distinguishes encoded Company state from existing
external contribution rows, participation weight, Guild settlement coordinates and the
closed transition's other authorities. This is evidence constraining that pending design,
not a new save format, arithmetic algorithm, chosen lifetime/absence policy or accepted RFC.
Nonempty Guild settlements, active-buff temporal boundaries, live capture/SQL/TS and general
conservation remain unproved. No production, balance, kernel/save, policy, owner copy,
accepted contract, CI workflow or acceptance checkbox changed.

Review by: Codex (research implementer first filter). Recorded by: Codex. Exact new
test/draft/plan/log range begins after `d9359e88` through this containing commit; designated
Claude review pending, separate from earlier debt. Next: finish the remaining producer and
settlement contract (especially active temporal inputs and nonempty Guild batches) under
the unresolved interval/lifetime ruling, then implement a reviewed viable algorithm. Do not
repeat ordinary scalar profiles, pre-advance live state or call this green server CI.
No archival, push, deployment, cleanup or release claim; the complete1.0 goal remains active.

## 2026-10-12 — Nonempty Guild clearing through the closed transition (R-012)

At `a598167b`, predeclared six cells: single/ordered/nonzero-base clearing batches ×
online/offline. Extended the existing origin instrument and its manual input helper;
the helper now preserves nonempty replay batches instead of dropping them. No production
byte changed. Actual `resolveReplayAccrual` and `StockConsumptionProvider` execute with
fixed membership/pending-batch lookup; observed full pre-provider state proves clearing
precedes rate resolution, while the original Company remains exact after resolution.
Existing context encoding retains the batch and contribution rows. Four fresh scratch
copies followed by a real logged manual action reproduce ordinary whole state, exact
receipt and ordered events, including positive Prestige/Guild hooks.

Negative controls: omitted clearing keeps cash equal under the pinned zero Guild bonus
but loses stock/consumption/watermark state; duplicate batch application refuses without
mutation. A real known-generator purchase reaches its exact `unaffordable` receipt after
accrual/hooks, rolls back the full Company, then the valid manual action reproduces the
ordinary result. Census completes only after every assertion: six, no exclusions.
Initial23473 instead refused all six at state admission because the fixture seeded stock
without incorporation; invalid research evidence, not a product defect. Corrected to the
existing bootstrapper incorporation fields and zero cash. An unknown manual action was
also unsuitable for late rollback (source refuses before evaluation), so the instrument
uses the specified unaffordable purchase; initial admitted61026 passes all six.

Final checks: native78283 `make test-go vet GO_PACKAGES='./production ./faction ./guild'
GO_TEST_FLAGS='-count=1 -run TestAxisOrigin'` passes all origin tests and vet (the selector
finds no tests in Faction/Guild). Linux-amd6446428 `make test-go-ci
CI_TEST_PACKAGES='./production' CI_TEST_FLAGS='-v -run TestAxisOriginDeferredGuildSettlementResearch'`
passes all six verbosely. In-memory test, not SQL proof despite that lane's DB URL.
Whole cold48993 `make test-go GO_PACKAGES='./production ./faction ./guild'
GO_TEST_FLAGS='-count=1'`: Faction/Guild pass; production remains RED with exactly the
original27 RP-307 partition failures (38.246s), no new failure. Native DB prerequisites
skip; no TS, hosted, full-CI or AC6 success claim. All handles terminal; formatted/diff checked.

The Codex-authored draft C2/C3 now records this bounded once-only/rollback evidence.
General clearing workloads, live membership/SQL capture, temporal active buffs, TS and
the unadopted episode/origin-lifetime contract remain. Review by: Codex (implementer first
filter). Recorded by: Codex. Range after `a598167b` through this containing commit needs
Claude review; no self-approval or waiver of earlier review debt. Next: specify and test
temporal producer boundaries, then the viable accounting/compatibility algorithm under
explicit episode/lifetime authority. No balance, save/kernel, policy, owner copy, criterion,
accepted contract, workflow, archival, push or deployment changes; full1.0 goal active.

## 2026-10-12 — One contribution per owned PR Intern (RP-509)

**Review by:** Codex. **Recorded by:** Codex. **Decision:** CHANGES REQUIRED
on the contribution-identity check in original Claude P5 `2f2c5a04^..2f2c5a04`.
This is a bounded decoder finding, not a verdict covering all 30 original paths
or approval of later Codex repairs. At clean `a5a2e94a`, the original count-only
check is still present: two sorted unique source rows may repeat one owned
upgrade and omit the other, while both factors match the named intern and the
row count equals the owned count. Old-source `bf08f0` runs the13-case decoder
population: both repetition directions FAIL with missing expected rejection;
the other 11 cases pass. No producer emitting such a malformed arm or public
client gameplay injection is alleged.

Outcome under accepted CV9: reject that contradictory breakdown before it
replaces the retained Game UI snapshot. The repair tracks contribution upgrade
identities, leaving the existing ownership/factor/order/count checks intact.
No locally calculated factor/product, API schema, Go producer, content/copy,
save/kernel, gameplay arithmetic, owner choice or CI workflow changes. Tests
retain distinct two-owned contributions, equal factors on distinct interns,
the Go-projected golden, both duplicate directions and input immutability.
The existing browser runtime test additionally traverses its actual generated
snapshot client with a mocked HTTP response: refusal preserves retained
revision and credentials; the healthy companion still loads. It is not a real
HTTP/Postgres claim.

Executed final checks:
- `a92518`: four affected Node files (`axis-stack-arm`, `game-ui-runtime`,
  `game-ui`, `axis-factor-format`), 185 passing tests.
- `7298`: `make typecheck build-client`, zero Svelte errors/warnings and a
  successful production build.
- `88555`: root `make test-browser-focused` selecting `axis-stack-arm`,
  `axis-stack-browser`, `game-ui-runtime` in native Chromium/WebKit, 170 tests
  across six file executions pass. Initial sandbox local-listen refusal ran
  no browser tests; the same narrow command then executed successfully.
- `22116`: `make test-clout-composed` exits 0. Built client, real game service,
  Postgres and WebSocket execute attainment 2→12, native first PR purchase,
  product 1.3/current factors, owned reload, exact SQL head/purchase and four
  real re-attainment events→two notices. Existing Cosmetic purchase/adoption/
  equip/unequip and care populations also pass. This fixture has controlled
  setup cash and only one owned PR Intern; two-owned identity controls above
  are decoder tests, not two live purchases or pacing/mint evidence. Initial
  sandbox Docker refusal ran no journey; unrelated orphan containers untouched.

All handles terminal; diff checked. **Review by:** Codex (repair implementer
first filter). **Recorded by:** Codex. New full source/test/docs/ledger/record
range from `a5a2e94a` through this containing commit needs designated Claude review,
consolidatable with related Clout corrective debt. No whole-P5/AC11/Clout,
Firefox/Linux/hosted/full-CI, archival, push or release claim. Next: continue
accepted outstanding implementation/review work; pending drafts remain pending
and do not block unrelated accepted repairs. Full nine-tier 1.0 goal stays active.

## 2026-10-12 — Both PR purchases through the built client

Outcome under CV1/CV3/CV9/AC11: extend the existing composed fixture from one
owned PR Intern to both, without changing product behavior or content. The
original first-purchase/reload/SQL assertions remain. The native Senior PR Buy
now checks its exact request and bound applied receipt, both distinct server
contributions, literal ×1.612 product in receipt and rendered UI, disabled
rebuy, both-owned reload and exactly one stored purchase event per Intern.
Both individual factors remain ×1.3/×1.24; progress bars disappear only for
owned rows. Setup cash changes from `1e8` to `1e10` so both existing prices are
affordable after native Buy Max. This is explicit test arrangement, not earned
cash, a pacing-bound change or content ratification. No attainment, ownership,
purchase/notice event is seeded; controlled prior-earned Founder history remains.

Executed: `31758` `make test-clout-composed` passes the new two-owned population
and unchanged Cosmetic/care population (12.89s driver); initial sandbox Docker
refusal ran no test. `32616` full `make test-game-ui-composed` exits 0: its
mandatory refresh 8/8 and ten real-DB parent tests execute; main production
client, two terminal states, next-run/recovery, default Cosmetic/Reputation,
new two-owned Clout, both N5 failure fixtures, Garden harvest and Tier-2
candidate journeys pass. Deliberate process-disconnect proxy errors remain
visible; recovered outcomes are required. Syntax and diff checks pass; all
handles terminal and unrelated orphan containers untouched.

Review by: Codex (test implementer first filter). Recorded by: Codex. Exact
new driver/docs/plan/log/ledger range after `0c2d1ef4` through this containing
commit needs Claude designated review; it may be consolidated with the related
decoder correction beginning after `a5a2e94a`. No full P5/AC11/Clout approval,
Linux/hosted/full-CI, natural pacing, mint, archival, push or release claim.
Next: continue accepted implementation/review work while pending owner/author
contracts remain explicit. Full nine-tier 1.0 goal remains active.

## 2026-10-12 — R-012 temporal context, without a new accounting policy

Completed the bounded measurement predeclared in the owning plan at `b9dabdc4`.
The existing origin test file now runs42 admitted cells: production frenzy,
building special and their combination, seven offsets, online/offline. Buff
history and cash are controlled arrangement; pinned durations/production bytes
are unchanged. Typed replay-input roundtrip preserves exact bytes. Scratch
schedule/evaluation leaves the committed origin untouched; its ledger plus one
existing manual action matches the real logged transition. Retained context
preserves full state, receipt and ordered events, positive lifetime accrual,
exact surviving buff identities and the literal long-absence span.

All cells complete without exclusions or guard termination. Frozen-at-origin
modifiers change cash in24 expired cells; omitting modifiers changes cash in24
still-active cells. Legitimate neutral controls survive. All84 forged-attended/
missing-schedule variants refuse unchanged. The important compatibility finding:
current command scheduling removes expired buffs BEFORE accruing that command's
interval. This proves endpoint-context fidelity, NOT prorated window payout.
Short ModeOffline intervals still advance attended time;6000ms crosses the
existing5000ms classifier and records the entire absence in either mode. Neither
observation authorizes a new absence policy or silently changed buff semantics.

Executed: final15199 root `make test-go vet GO_PACKAGES='./production ./activeplay'
GO_TEST_FLAGS='-count=1 -run TestAxisOrigin'` passes all origin tests and vet;
Activeplay has no selected tests. Final28529 root `make test-go-ci
CI_TEST_PACKAGES='./production' CI_TEST_FLAGS='-run TestAxisOriginDeferredTemporalContextResearch'`
passes on declared Linux/amd64; in-memory population, not SQL proof. Earlier
verbose native27652 and Linux35835 execute all42 individually. Cold broad74591
`make test-go GO_PACKAGES='./production ./activeplay' GO_TEST_FLAGS='-count=1'`
retains exactly the original27 RP-307 failures (production46.756s); Activeplay
passes. That broad run precedes the final test-only absence-span/typed-roundtrip
oracle strengthening; the final affected population and vet rerun pass. No
tolerance, excluded failing fixture, production, balance, kernel, save, owner
copy or CI-workflow edits. No hosted/TS/full-CI/AC6 success claim. All handles
terminal; unrelated orphan containers untouched; formatting/diff checked.

Codex-authored draft C2 now records this bounded evidence and remaining general
temporal/compatibility obligations. Next: complete the viable accounting and
compatibility proposal under explicit interval/origin-lifetime authority, rather
than repeat these bounded successes or treat them as a general numeric repair.
Review by: Codex (implementer first filter). Recorded by: Codex. Exact full
test/plan/draft/log range after `b9dabdc4` through this containing commit needs
Claude designated review; earlier review debt remains. No adoption, acceptance,
archival, push, deployment or release call. Full nine-tier 1.0 goal stays active.

## 2026-10-13 — R-012 temporal replay compatibility across runtimes

Predeclared in the owning plan at baseline `5bc0e222`. Extend the existing42 Go
temporal cells, not the accounting policy or the population: three controlled buff
histories, both modes, seven offsets. Record their actual retained inputs and complete
canonical state/receipt/ordered-event outputs through the existing artifact helper.
TypeScript's actual pinned-bundle loader, Company codec and `applyLogged` reproduce
all42 outputs and restored results. Positive lifetime accrual, exact kept buffs and
the literal absence span match; both runtimes refuse all84 forged-attended/missing-
schedule variants without state mutation. Existing Go24 frozen/24 omitted differences
remain. No scalar prototype, direct ledger overwrite or successful-subset comparison.

Initial Node93121 fails43 checks through two instrument mistakes: the metadata
expects null where Go's existing codec omits the schedule, and the positive checks
compare lifetime against the already-mutated successful input. All42 canonical output
comparisons had passed before that latter assertion;84 refusal checks passed, but the
run is not counted as completed research. Correct to actual omission and captured
pre-action lifetime/buffs, retaining full output equality and refusal nonmutation.
Do not turn those errors into production findings. The first uncommitted generated
artifact duplicated84 invalid contexts; regenerate once with an actually counted
84-refusal census and construct those two exact variants independently in TS.
The completed artifact is719594 bytes,42 rows,12 selected current source hashes;
SHA256 `eb4194d330f17243189f683f82b231d5cfd34d1a549c364e6d09c714b48001c2`.
This is bounded evidence size/provenance, not a production storage budget or full
transitive/release identity. No historical committed artifact is restamped.

Executed root checks: generation31487/64267 passes every Go control; ordinary
origin regression/vet80299 and final48377 (also the existing artifact-comparison
negative controls) pass without regeneration. Node75691 passes230 affected temporal/
sequence/migration/receipt cases; native16847 passes254 temporal Chromium/WebKit
executions,2.90s. Types54992 has zero errors/warnings. Full Node35971 passes10795,
with1022 explicit existing skips,8.82s. Original timing/partition run67492 remains
RED with exactly27 partition failures; no criterion/tolerance/skips changed.
All handles terminal; formatting/diff checked. Normal existing CI collection owns
these tests; no workflow/job/timeout/concurrency change or new CI lane.

Only tests, the new observation and planning/own Codex-authored draft evidence
change. This is not a conservation implementation, SQL/live buff acquisition,
Linux/Firefox/hosted-CI or release-artifact proof. C1/C1a interval/authorized-head
meaning and the viable C2/C3 ledger algorithm remain unresolved; general external/
Guild/temporal compatibility is not claimed. Do not repeat this bounded population
as the next repair step. Review by: Codex (implementer first filter). Recorded by:
Codex. Full range after `5bc0e222` through this containing commit needs designated
Claude review. No adoption, acceptance/archival, mint, push, deployment or1.0
promotion; next is the complete viable accounting/compatibility contract under
the pending interval/lifetime authority.

## 2026-10-13 — Full original P5 integration review

Review by: Codex (designated independent reviewer of Claude). Recorded by: Codex.
**CHANGES REQUIRED**, exact `2f2c5a04^..2f2c5a04`, all30 changed paths. Provenance:
Claude's owning implementation record and co-author trailer. Inspect producer/
read helpers/projector, registered schemas, decoder/host/panel/presentation, tests/
golden/composed-driver addition, generated API/copy, three version markers,
content manifest and owning records. Canonical docs at that coordinate still
list CV9 rendering as undelivered; the original checkbox does not prove full AC11.
The current plan already distinguishes presence from full acceptance.

Existing findings directly block original approval: RP-303 unnamed native progress
bars, RP-387 resource formatting concealing fractional factors, RP-426 progress/
factors outside their purchase cards, RP-509 duplicate owned-intern identities.
Fresh in-memory execution uses the actual historical decoder/helpers and numeric
parser: both duplicate directions escape refusal; eleven original/healthy controls
pass. No production mutation or current decoder substitution. Original panel/
Amount/formatter source traces the rounding path; retained65716 native red evidence
owns its executed failure. A direct historical formatter probe here cannot load
the notation package's legacy CommonJS dependency and is excluded, not a browser
or product failure. No dependency installation/patch. Projector source is byte-
unchanged from the original, but the checks below run current dependencies.

Readonly original generated checks: four new schema names plus one optional
Features property, with every preceding API field/path/definition preserved;
compatibility pin byte-unchanged. All13 candidate rows equal generated entries,
all496 preceding copy rows preserved;509 TS/catalog/Go-All keys agree,462 orphan
keys unique/sorted/known, hash references agree and constants hash unchanged.
All three version markers move127→128 consistently. Initial checker counted the
following34 companion keys as part of `All()`; scope its actual function and
rerun successfully. This is consistency, not a fresh historical compiler/hash
derivation or prose adoption. Candidate copy remains pending owner action.

Current source `59974458`: cold root72639 producer/registered-envelope tests and
gameui/account vet pass; the stored-projection SQL test explicitly skips for
missing TEST_DATABASE_URL. Node59957 passes86 decoder/factor/runtime cases;
native54013 passes36 current Clout cases in Chromium/WebKit,3.59s, including
purchase-card bindings, exact factors, progress names and current help/notices.
Native log is tool-truncated; terminal census is present, not a full trace claim.
No historical Svelte execution, fresh SQL/composed, Firefox/AT/performance/full
CI or whole-feature approval. All handles terminal; no product/test/copy changes.

Later Codex repairs retain their own designated-review ranges; this original
negative verdict approves none of them. Projected-factor/contextual-hint meaning,
content adoption/mint, full acceptance, arithmetic and the complete nine-tier
goal remain open. Continue remaining original reviews and accepted construction;
do not rerun this completed original P5 review or silently close corrective debt.

## 2026-10-13 — RP-536 contained Clout presentation failure

Outcome under accepted CV9/GS0.5: a missing or parameter-incompatible cap-reason
copy row must produce the existing region-level error, not crash Desk or render
as if required presentation exists. Preflight the row's zero-parameter signature;
retain the heading and `axis-why` description target, hide the unavailable readout/
Help, report one fixed invariant per mounted error episode and recover on healthy
data. Other Desk controls and navigation remain usable. No copy, factors, caps,
decoder admission, server requests, schema, kernel or CI semantics change.

Old source `30050a93`: typed-clean initial59103 fails8 missing-key executions;
extended76058 fails all16 missing/parameterized × below-cap/saturated × first-open/
mounted-refresh executions, including uncontained RangeError/TypeError. These are
actual decoder-admitted wire fixtures, not proof of a valid minted producer state
or a public exploit. Retain healthy controls, zero uncaught errors, exact alert/
diagnostic, unchanged-tick deduplication, recovery/new-error episodes, native tab
navigation, enabled Manual/three upgrade cards, no intents,320px reflow and axe.

Executed root commands: `make test-browser-focused` with axis-stack and game-ui-
screens files in Chromium/WebKit passes262 functional cases (10 performance cases
excluded by the existing lane),17.28s. After adding the independent-controls
assertions, final axis-stack file32192 passes all52 executions,5.01s. Node axis-
stack-arm/axis-factor-format/game-ui-runtime35512 passes86. Final `make typecheck`
1391 has zero errors/warnings; `make build-client`88137 and `make verify-client-
boundary` pass. `make test-game-ui-performance`58035 runs alone and passes the
existing5 render-budget cases,105 deselected,5.84s. First passing51917 overlapped
other checks and is not counted as the isolated measurement. No bound is changed.
Final diff checked; all named handles terminal. Large browser tool output is
clipped; terminal results are retained, not a complete raw-trace claim.

Coverage is mounted UI/runtime and the existing local performance population,
not fresh Postgres/composed, Linux/Firefox, assistive technology, pinned manual
profile, hosted/full CI, CV9/AC11 or release proof. RP-131/RP-307 and all earlier
corrective reviews remain open. Review by: Codex (implementer first filter).
Recorded by: Codex. Full range after `30050a93` through this containing commit
needs designated Claude review, consolidatable with related CV9 repairs. Original
P5's negative verdict does not approve this fix. Next: remaining original reviews
and accepted construction; no archival, content mint, push or1.0 promotion.

## 2026-10-13 — Full original P4 Gaia-law review

Review by: Codex (designated independent reviewer of Claude). Recorded by: Codex.
**CHANGES REQUIRED**, exact `4c089f00^..4c089f00`, both changed paths:
`server/production/gaia_law_test.go` and this log's original P4 entry. Claude
provenance is explicit in that entry and the commit's co-author trailer. Inspect
the whole test/collector, its three direct seeded writes, two catalog checks and
the original reported source/formula severing. The helper/test remains byte-
identical today; historical execution is not claimed for current dependencies.

RP-537: the AST visitor only matches a direct SelectorExpr on assignment/inc-dec
and named literal keys. An independent temporary test type-checks three source
fixtures through Go's actual parser and go/types before invoking that unchanged
helper: `founder.CloutLifetime += 2` is found, but
`(founder.CloutLifetime) += 2` and
`p := &founder.CloutLifetime; *p += 2` each return an empty writer inventory.
Cold root `make test-go GO_PACKAGES='./production' GO_TEST_FLAGS='-count=1 -run
TestGaiaWriterReviewProbe -v'`91754 exits2: two failed probes, one passing direct
control. These are type-checked source fixtures, not executed production mints.
Remove only that own temporary test after the terminal result; no runtime bytes
are mutated and no failing acceptance check is disabled. Its exact bodies above
retain the bounded reproduction against `cloutWriters`.

Existing `TestGaiaLawNoCloutWriterUnderOptionA`28063 passes cold despite these
escapes. The entire `save/state.go` exclusion is a separate source finding:
the collector cannot enforce its claimed round-trip-only exception. No codec
mutation is executed or credited. Current source search finds lifetime storage/
round-trip/validation references, not a demonstrated live mint. The recursive
import/source ownership and closed-union requirements are not established merely
by finding no matching field writes or no Clout resource in two catalogs.

After probe removal, cold26463 reruns the unchanged Gaia test successfully;
`make vet GO_PACKAGES='./production'` passes. All handles are terminal and the
diff contains only records, not tests or production changes. These green checks
do not close the demonstrated oracle gap.

The original formula-check anecdote is retained as Claude's executed evidence,
not a fresh Codex AC10 review in this two-path range. This verdict approves none
of the later Codex corrections/research. Next: repair the guard's mutation/codec
boundary and retained discriminating negatives under accepted AC4, then review
that corrective range; remaining original P1/P2/P3 and full feature gates remain.
No product, numeric, reward, copy, version, CI, mint, archival or release change.

## 2026-10-13 — Full original P1 loader review

Review by: Codex (designated independent reviewer of Claude). Recorded by: Codex.
**CHANGES REQUIRED**, exact `ace4e67e^..ace4e67e`, all17 changed paths. Claude
provenance is explicit in the original log and co-author trailer. Inspect both
loaders, achievement maxima, Go/TS bundle checks and static-contribution skips,
slot order/test, generator, shared loader tests/corpus/fixture, three version
markers and the original record. Existing RP-304's executed empty-role findings
block original approval; RP-305's truthful role/pool binding remains separate.

RP-538 newly exposes source ownership: both original branches verify presence
but reduce effects to a set without rejecting cross-upgrade repetition. Current
old-behavior Go/TS regressions each admit four collisions (both owners, equal/
different factors) while retaining two distinct-source controls. This is fresh
execution of the current unchanged ownership branches, not a historical compiler
or full original runtime execution. The original helper's "exactly one" comment
does not establish that invariant; later source-keyed factor production and
duplicate-contribution refusal depend on it.

Readonly in-memory execution of the actual historical generator reproduces both
original outputs byte-identically. Removing only the two proposed PR upgrades/
declarations, axis block and schema5 increment restores the complete original
v4 catalog. Corpus22 economy/4 cross/2 route names are unique; all three kernel
markers agree at126 after125. No filesystem generation or balance mutation.
These are artifact consistency checks, not content ratification, full route/
reward/SQL proof or approval of the separate formula-publication commit.

Later Codex role fixes and the following ownership correction need their own
designated Claude review; this original negative verdict approves none. Original
P1/P4/P5 reviews are complete; P2/P3 and whole-feature gates remain. No archival,
content mint, historical rewrite, CI relaxation, push or release promotion.

## 2026-10-13 — RP-538 unique axis contribution ownership

Outcome under CV1 and existing contribution identity: reject two upgrades sharing
one declared axis source before gameplay/replay. Retain distinct IDs even with
equal factors. Four old-loader refusals fail per runtime; two legal controls pass.
Reuse the existing mutation helpers, fixture and corpus; no new test framework.
Both loaders now track axis source ownership explicitly. Kernel186→187 is a real
admission-set change, not an attempt to repair RP-131; all three markers change
together. Formula shapes, catalogs, values, copy, slots, schema and pins stay put.

Executed root checks: affected economy/multiplier/achievements/replaycatalog/
routes/decimal/kernel81433 and vet96593 pass cold, including Go golden vectors.
Node4919 passes376 in eight actual files; the requested nonexistent numeric.test
selector runs nothing and gets no vector credit. Corrected full `make test-client`
16529 includes decimal-vectors and passes10801, with1030 explicit existing/browser
skips (eight newly added Clout browser declarations explain the increase).
Native2924 passes104 loader/role/formula executions in Chromium/WebKit,1.65s.
Types/build/client-boundary/formulas4921 pass; generated formulas and original
fixture/corpus/API pin stay byte-unchanged. All handles terminal; final diff checked.

Broader cold production/gameui36590 exits2: production43.061s fails exactly the
original27 RP-307 partition cases; gameui passes. `make verify-kernel-version`
76631 also exits2 at historical50a3a514, with its checkout/adversarial controls
passing first. No assertion, tolerance, guard, retry, timeout or CI lane is loosened.
Native package runs do not supply fresh SQL, Linux/Firefox, hosted CI, content
mint, full AC1/AC6 or release evidence; Docker-dependent journeys remain separate.

Review by: Codex (implementer first filter). Recorded by: Codex. Complete corrective
range after `e4054647` through this containing commit, including these records,
needs designated Claude review, consolidatable with related CV1 fixes. Original
P1's negative verdict is not approval of this range. Continue accepted work and
remaining original P2/P3 reviews; the full nine-tier1.0 goal remains active.

## 2026-10-13 — Full original P2/P3 formula and save review

Review by: Codex (designated independent reviewer of Claude). Recorded by: Codex.
**CHANGES REQUIRED**, exact `f256b235^..f256b235`, all28 changed paths. Claude
provenance is explicit in the original record and co-author trailer. Inspect the
entire Go/TS attainment/formula/serialization path, activation and Exit boundaries,
active-play and Soul/recovery consumers, eligibility/read projection, event
registry/migration, release migration pin, tests/vectors, version markers and log.

Existing acceptance gaps remain: RP-307's absent original timing/partition proof
and actual27 divergences; RP-309/310's missed TypeScript v19 active-play predicates;
RP-311's absent derived receipt axis object. The original log's "every Go/TS v18
check is inclusive" claim is false at its own source. Retain its openly reported
surviving float-division probe and AC3 receipt-equivalence contradiction as author
reconciliation, not accepted waivers or permission to change legitimate earnings.
Unselected score-input fixtures do not prove production compatibility for that
alternative; the ruled release input remains attainment.

RP-539 newly reproduces the original/current encoder's misplaced guard: v14/v16/
v17 return before its pre-v19 prohibition, so9 field-presence/score refusals fail.
V18's3 negatives and4 healthy legacy controls pass; Go's16 matching cases pass.
This is fresh execution of the current unchanged control-flow boundary, not a
full historical compiler/service run. The correction below is a separate scope.

Readonly original artifact checks: eight formula and nine attainment vectors
have unique names; both files and applied00081 remain byte-unchanged. Migration
Up adds exactly re-attainment to the preceding kind set; Down restores that set.
Prior schema constraints keep this kind atv1; no applied migration is edited.
All three original kernel markers agree at127 after126. This is source/artifact
consistency, not a fresh SQL migration/constraint proof. Current cold11486 passes
activation, derivation, both vector generators without regeneration, carry contrast,
v19 codec and new legacy cases, plus release migration contiguity. The actual
Postgres test explicitly SKIPS for missing TEST_DATABASE_URL. Existing source
search finds no generic events UPDATE trigger masking its kind/version checks;
no such defect is asserted without evidence. Original reported SQL/severing remains
Claude's historical evidence, not a fresh Codex execution.

Original P1–P5 implementation dispositions are complete. Later Codex corrections,
compatibility research, migration corpus and service/browser supplements require
their own designated reviews; this negative verdict approves none. P6/harness,
accounting, author/copy/mint and full feature/release gates remain open.

## 2026-10-13 — RP-539 refuse attainment before every legacy encoding return

Outcome under CV4: non-null attainment (even an empty set) or nonzero attainment
score must refuse before encoding any pre-v19 Company, not disappear from its
wire output. Move the existing TypeScript guard to the function entry. No Go
runtime codec, serialized healthy bytes, replay chronology, arithmetic, field
schema, artifacts or balance changes. Kernel187→188 and both generated markers
record the real refusal change. No raw observation is regenerated/restamped.

Old TypeScript14017 fails9 regressions,7 controls pass; Go's first attempt fails
compilation because my test omitted fmt, not a product finding. Add that import;
cold38065 passes all16 matching Go cases. Fixtures shape earlier-version wire
controls from the unchanged activation observation; they are codec-boundary
inputs, not a claim of real historic migrations or a server persistence exploit.
Existing v19 migration/Exit/receipt cases retain their exact bytes and behavior.

Executed root checks: Node40270 passes6608 including shared decimal vectors and
full replay/attainment/migration/receipt cases; native46760 passes146 in Chromium/
WebKit,13.20s. Full `make test-client`52599 passes10817,1030 explicit existing/
browser skips. Cold save/kernel/decimal51440 and affected vet56352 pass; selected
production/save/release11486 passes with its SQL skip explicit. Type/build/client
boundary/formulas40079 succeed; its following migration target fails on sandbox
socket permission before a DB check. Authorized retry61903 produces no test output
and remains live after82seconds: verified make20820→docker20828→compose20830.
Deliberately interrupt only that owned Compose client, then observe terminal
Make exit2 / child130. No Docker restart, daemon shutdown or unrelated cleanup.
Migration/DB acceptance is unverified, never promoted from that attempt.

All named handles are terminal; diff checked. RP-131/RP-307, Docker availability,
Linux/Firefox, hosted/whole CI, AT and exact-release evidence remain separate.
Review by: Codex (implementer first filter). Recorded by: Codex. Complete correction
after `ed129ec5` through this containing commit, including tests/records, needs
designated Claude review, consolidatable with related CV4 fixes. The original
P2/P3 verdict is not approval of this range. Next: RP-537's guard boundary and
remaining accepted construction; no archival, mint, push or1.0 promotion.

## 2026-10-13 — RP-537 checked mutation routes and codec exceptions

Outcome under accepted AC4: close the demonstrated parenthesis/field-alias gaps
and stop exempting the whole codec file. Old guard67617 misses six type-checked
write forms; three direct-write and two value-read controls pass. Unwrap LHS
parentheses and report field-address escapes as mutation-capable routes, without
claiming each address is an executed mint. Ordinary value reads stay allowed.

The actual repository collector now includes `save/state.go`. Its only exceptions
are exactly one `normalized.CloutLifetime` copy in `EncodeStateVersion` and one
`source.CloutLifetime` copy in `RestoreState`. Changed values, missing/extra copies,
other codec writers and aliases refuse. Eight typed codec cases retain a healthy
baseline; three in-memory additions to the actual collected codec each produce
one specific violation. No production source is mutated for these probes.
The existing Founder v17/v18 test additionally round-trips 1, 999 and MaxExactInteger
through the real encoder/restorer with identical bytes; the zero control remains.
This tests actual permitted persistence, not only an AST shape or field census.

Final root84779 `make test-go GO_PACKAGES='./production ./save'` with cold
`Test(Gaia|FounderV17AndV18RoundTripWhileCompanyRejectsThem)` passes: all four Gaia
tests (314 current non-test server files, including both copies) and real codec
controls. Affected production/save vet passes. Initial corrected48883 preceded
the final codec cases and is not complete-batch credit; final45597/84779 cover
them. Gofmt/diff checked, all handles terminal. Runtime/source markers, fixtures,
schema, balance and kernel0.3.188 are byte-unchanged. Only tests, canonical guard
description and owning records change; no CI lane or new framework.

This remains a conservative explicit-write/field-address guard, not whole-program
alias/reflection/external-library proof. Its two catalog absence checks do not
establish every reward/effect union. Full AC4, original P4 acceptance, Docker/SQL,
hosted CI, accounting and the full nine-tier goal remain open. Review by: Codex
(implementer first filter). Recorded by: Codex. Full range after `e7ad3591`
through this containing commit needs designated Claude review; consolidate with
the related corrective range after `30050a93` instead of reviewing each record
separately. Original negative verdicts approve none of these fixes. Next: remaining
accepted ownership/reward boundaries and construction, with no archival or push.

## 2026-10-15 — Prepare actual external/Guild replay observations

The old serialization/SQL studies are complete; do not repeat them as new progress.
Declare the missing12-cell Go/TS compatibility question in the owning plan before
measurement. Extract observation-only returns from the existing two Go producers,
capturing the actual request/input rather than rebuilding it after the action.
Keep all omission, pre-advance, duplicate-clearing and late-purchase rollback checks.

`TestAxisExternalReplayEvidenceCapture` verifies the exact twelve names and replays
each capture through actual Go ApplyLogged, comparing full state, receipt and ordered
events. Cold67241 also passes the existing20 lower-domain,128 purchase,42 temporal
and6+6 external/Guild populations; production vet passes. Verbose capture is truncated,
so only retained parent verdicts/census assertions, not a complete per-child transcript,
are claimed. Old temporal results match with historical source differences disclosed;
no corpus or recorded digest is overwritten. Formatting/diff checked.

Review by: Codex (test implementer first filter). Recorded by: Codex. Preparation
after `12208257` starts the related research range; consolidate its later Go/TS capture,
controls and results into one designated review. No runtime/kernel/save/wire/balance,
policy, content, CI or author body changes. TS comparison and new artifact generation
remain next, not completed; original27 failures and all1.0 gates remain open.

## 2026-10-15 — Complete bounded external/Guild Go–TypeScript replay comparison

Outcome: all twelve declared actual Go transitions reproduce complete Company,
receipt and ordered events through TypeScript replay and restoration. Each of the
six nonempty Guild cells also carries omission, duplicate-refusal and late-purchase
rollback controls. Omission preserves neutral-rate cash but changes settlement
state, so cash-only comparisons cannot pass this requirement. Three forged expected
outputs (state, receipt, events) are caught without mutating the fixture.

Reuse existing producers/research helpers and one hash-bound bundle; the new shared
fixture is397164bytes with twelve actions, six control sets and eight selected source
digests, all checked against final source bytes. It is a replay fixture, not a runtime
storage bound or complete dependency manifest. Generation uses root
`make test-go GO_PACKAGES=./production GO_TEST_FLAGS='-run TestAxisExternalReplayCompatibilityResearch' UPDATE_EXTERNAL_REPLAY_RESEARCH=1`.
The earlier writer handle21049 is now missing and its final output was lost; do not
credit that terminal result. Fresh cold43851 regenerates and compares all result
bytes without writing, passes in0.927s, and production vet passes. It includes the
existing lower-domain,128 purchase,42 temporal and external/Guild controls.

Current Node comparison passes161 cases (34 new,127 unchanged temporal). Native
Chromium/WebKit56689 passes68 executions in1.21s, no selected skips. Typecheck12222
reports zero errors/warnings. The first browser attempt cannot bind localhost under
the sandbox and runs no tests; permitted retry produces the native result above.
Earlier instrumentation errors are retained: parsing a canonical replay payload as
a complete incoming intent failed six Go controls; use existing parseLoggedIntent
with its captured command ID. Typecheck caught a test callback returning replay
state; the async callback now awaits the comparison without returning that value.
Neither correction changes production behavior or weakens assertions.

No runtime, kernel, save/replay schema, arithmetic, balance, CI or owner-authored body
changes. The Codex draft's evidence is updated, not adopted. This is bounded existing
compatibility, not conservation, live provider/SQL acquisition or a browser-to-server
player journey. Original27 partition failures remain evidenced by the current full
server checkpoint, not rerun or waived here. Review by: Codex (implementer first
filter). Recorded by: Codex. The whole research range after `12208257` through this
containing commit, including preparation `6ed5ac79`, needs designated Claude review;
no archival or feature/release promotion. Next: viable accounting/interval/origin
contract and remaining accepted construction, not another replay of this study.

## 2026-10-15 — RP-637 full Linux native disclosure failure and diagnostics

Accepted CV9.4: native Enter/Space must operate the PR Codex and preserve help/focus
through authoritative refresh. Clean product `b248d26a` `make test-browser-ci`14637
exits2: Firefox Tier2 Enter leaves details.open=true at Close (old line422).
Broad477 files:476 pass/one fail;39800 ordinary pass,9 expected failures,24 skips,
141.34s. Later native/performance phases do not execute. Final assertion/diff/text
is retained; inspected screenshot is blank after teardown, not cause evidence.

Add passive stage/mounted/focus/trusted key/click observations and intermediate
focus assertions, preserving native input and every old outcome assertion. Focused
Chromium/WebKit62408 passes12 cases (40 selector skips); same-image Linux Firefox
50613 passes6 (20 selector skips). Full diagnostic46034 still exits2: same Firefox
Tier2 Enter now leaves details.open=false at Open (new line422), with the same
broad result counts in140.67s. Its live diagnostic is lost in truncated stdout.
This is an observation-delivery defect, not authority to blame axe, relax the
assertion, refocus the summary, increase a timeout or add gameplay retries.

Bind the diagnostic to the final thrown Error with its original cause. Capture
state immediately and after one existing settle cycle **only after failure**, always
rethrowing it. Full16929 then exits0: broad477 files/39801 ordinary passes,9 expected
failures,24 skips in137.97s; native3 files/87 ordinary and6 expected in3.34s;
Chromium performance5 pass/105 selector skips in6.29s. Test-file SHA256 at this
observation:01f2f27568f1eaa61c3b7510d2a3360e2189b552562af246e18bcae3a4ff8eb2.
No production byte changed between these runs; that pass does not explain or repair
the failures. No repeated full run is used to manufacture a reliability claim.

Final controlled negative prevents only the native Close key's default action.
It must fail at Close with connected/current/focused elements, two trusted Enter
keydowns, open state before/after settle, and original assertion/error provenance.
A failure at any other stage cannot satisfy it; ordinary cases still require exact
open/close, preserved focus and zero gameplay requests. Final same-image cold
Linux Chromium/Firefox/WebKit62785 passes the whole affected file:81/81,7.31s,
no selected skips. Typecheck67227 reports zero errors/warnings; diff checked.
These final controls follow the full observation, not a new complete-job pass.

Playwright image inspection proves Linux/ARM64, not hosted AMD64 or exact runner
timing. Native keyboard provider's conditional frame-focus code is a hypothesis,
not cause evidence; no frame-focus workaround lands. Other suites/features are not
closed by these browser counts. All handles terminate and declared browser runners
auto-remove; final inventory is empty, no durable data/cache/network removed.

Review by: Codex (test implementer first filter). Recorded by: Codex. Full test/
ledger/linked status/record range after `b248d26a` through this containing commit
needs designated Claude review, consolidatable with CV9 corrections. RP-637 stays
open; no product/copy/schema/kernel/budget/worker/policy/CI-topology change, owner
decision, archive, push or1.0 promotion. Next: use retained final-error evidence on
recurrence to distinguish actual key delivery/default-toggle delay and frame focus;
do not rerun this unchanged population for another green or reopen completed studies.

## 2026-10-15 — RP-637 parent-focus control refuted

Controlled probe asks whether focusing a real parent-frame control leaves the
child summary as activeElement while native input misses it. Native Chromium/
WebKit90034 and same-image Linux Firefox72480 all fail that premise: child
activeElement is BODY immediately after parent focus, not the retained summary.
No keyboard action or later recovery arm executes. Typecheck25953 passes, but
that is not behavioral proof. Withdraw the entire uncommitted probe with an exact
empty source diff; do not restate its premise as a browser/provider defect or
land a focus workaround. Other routes to frame-focus loss remain untested.

Clean-source full Linux gate75582 at `fc398997` subsequently completes all phases;
the CI log owns results. No natural failure recurs, so there is no new diagnostic
causation evidence or RP-637 repair. Existing final-error diagnostics and original
failures remain. Next use actual recurrence evidence, not repeated unchanged runs
for reassurance. Record-only outcome; no new software review range or product
claim. Previous diagnostic implementation still needs its designated review.

## 2026-10-15 — R-012 live Guild/SQL input capture and actual Go/TS replay

Execute the bounded declaration in the owning plan at `639edf19`. Reuse the existing
HTTP/Postgres/socket/Guild worker parent: real account/Guild creation and membership,
three clearing ticks and its existing authenticated manual action. Faction/stock are
still controlled fixture setup. Capture its committed pre-head and read the actual
canonical command, resolved replay inputs, applied revision, receipt and ordered
Company events from SQL after commit. Require exactly three ordered clearings with
positive credit. Pure Go replay receives only that pre-head, the pinned bundle and
stored inputs; complete Company/receipt/events match. Omission loses settlement
state and duplication refuses with unchanged state/no output. JSON comparisons are
typed logical/jsonb equality, not a claim to preserve SQL formatting; the command
is separately byte-bound to the actual HTTP request's canonical form.

Direct declared Postgres20072 passes the whole parent1.962s. Reuse the existing
stdin/Vite replay runner with one explicit private-observation mode, not a new
framework. The same actual capture passes through TS restore/apply, complete output
comparison, restored result and omission/duplicate controls. No live provider is
passed to replay and no private payload/artifact is printed or persisted. The root
composed driver requires the exact new marker; ordinary Go explicitly gives no TS
credit without the existing opt-in seam. Default public-download mode is unchanged.

Final full `make test-game-ui-composed`61907 exits0:35 driver controls pass; both
persisted populations execute8/8 refresh leaves and ten required parents with no
captured test-header skips/failures, plus the new Guild/TS marker twice. Main
transitions/endings/recovery, Cosmetic/Reputation, Clout, both payment negatives,
Garden harvest and Tier-2 candidate journeys pass with their existing controlled
cash/maturity/content limits and exact built assets/Worker. Expected proxy errors
occur during deliberate SIGKILL/restart; both recovery objectives pass. Actions
ci.yml calls this same root target, not a new CI job or a hosted verdict. Final
cold Gameserver/vet84470 passes0.431s. Formatting/diff checked; removed only both
inspected owned tmpfs Postgres services, preserving networks/caches/durable data.

The earlier twenty lower-domain settlement cells revalidate38521 in0.390s while
tracing the candidate; that is not new research or conservation progress. This new
one-profile capture closes only its live-input/SQL/cross-runtime question. It does
not establish positive idle-production fidelity: the reused consumer buys no
generators. It does not cover Clout-v19 activation, every external/buff/offline population, a SQL fault/
restart protocol, the numerical write algorithm or C1/C1a. Original27 partition
failures remain untouched, as do the known under-floor write-set conflict and
release/owner/author gates. Research/tests only; no product/kernel/schema/balance,
source-pinned research artifact, policy, acceptance or archival change. Review by:
Codex (implementer first filter). Recorded by: Codex. Complete seven-path test/
driver/plan/log/draft-evidence range after `639edf19` through this containing commit
needs designated Claude review, consolidatable with R-012. No push or deployment.

## 2026-10-15 — R-012 positive idle producer in the live Guild capture

Under the bounded follow-up declared at `b2d77add`, strengthen the same profile,
not its conclusion. Old zero-generator88127 fails the new positive idle/lifetime
assertion while the original parent checks still pass. Add exactly one controlled
owned beige tower to the existing faction/stock setup write; no purchase event,
earned budget, data retune or natural progression is fabricated. The real HTTP
manual action now settles idle production and increases Prestige lifetime value.
Complete saved Company/receipt/ordered events still match actual Go and TS replay.
In fresh replay origins, remove only that producer: clearing/consumed units remain
but lifetime no longer increases. This separates idle-hook coverage from manual
output and clearing. Existing batch omission/duplicate controls remain mandatory.

Declared Postgres34835 passes the full parent1.684s. Final `make
test-game-ui-composed`55690 exits0:35 driver controls; both8/8 refresh populations,
all ten persisted parents twice, zero captured test-header skips/failures and two
stronger Guild/TS markers. Main/endings/recovery, Cosmetic/Reputation, Clout,
payment negatives, Garden and Tier-2 journeys pass with their prior fixture limits.
Cold `make test-go vet GO_PACKAGES='./gameserver'`42052 passes0.459s and vet.
Formatting/diff checked; removed only the two inspected owned tmpfs Postgres
services, preserving networks/caches/durable data. No raw private capture stored.

This proves one current-content online positive-producer profile, not Clout-v19,
natural acquisition/pacing, general provider/temporal fidelity or conservation.
Original27 partition failures and the floor/interval/lifetime/compatibility
contracts remain unchanged. Test/research records only; no runtime/kernel/schema,
balance, content mint, owner choice or acceptance/archival/push/deployment change.
Review by: Codex (implementer first filter). Recorded by: Codex. Complete seven-path
test/driver/plan/log/draft-evidence range after `b2d77add` through this containing
commit needs designated Claude review, consolidatable with the earlier capture.

## 2026-10-15 — R-012 origin-relative cumulative receipt feasibility

Predeclared in the existing plan at `fe1fa51d`: test a different receipt
representation, not another subtraction implementation. The test-only reference
restores one trusted encoded origin, runs actual Evaluate and retains its receipt
relative to that origin. Before returning a candidate, reconstruct both complete
previous/projected Company states and bind origin/catalog identity, expected time/
mode and every EvaluationResult field. Hashes identify trusted inputs; they do not
authenticate their provenance. No live state or ledger is replaced by the verifier.

All20 existing lower-domain cells produce exact one-shot candidates and reproducing
origin-relative receipts. All nine previous-head additive deltas still refuse with
unchanged state. Fourteen altered-field cases and a non-forward interval return no
candidate. Both cap/debit modes reject reuse of the old origin even with a rehashed
post-debit before image; a new origin gives the existing literal payout without
recovering discarded overflow. These are22 bounded Go profiles, not a full mutation
algorithm, Company wire, hook protocol, SQL/TS integration or new valid incremental
receipt. The trusted pinned bundle is an input, not a verified download/release.

Focused `make test-go GO_PACKAGES='./production' GO_TEST_FLAGS='-run
TestAxisCumulativeReceipt -v'`77167 passes0.377s. Final repeated/shuffled reference
run33823 with `-count=2 -shuffle=20261015` passes0.439s. Full affected cold
`make test-go GO_PACKAGES='./production ./economy ./decimal'`95139 remains red:
Production36.781s reports exactly the original27 partition failures; Economy6.197s
and Decimal0.226s pass. Same-package vet passes separately. Native DB skips are not
integration credit. No unchanged browser, Docker or hosted run is claimed.

Reconcile this author-owned draft's C3/body/deviation: cumulative Delta is relative
to the origin, not the preceding head. If saved intermediate heads are required,
that is an explicit alternative receipt contract needing adoption, an atomic
ledger-owned checked mutation and complete replay/hook/activation specification.
If cuts are internal only, retain the closed settlement route; C1/C1a still require
author/owner resolution. No generic approval can make the unchanged additive rule
express the retained floor increment. No new production authority or repaired AC6.

Formatting/diff checked; no services started or resources removed. Test/draft/plan/
ledger/log only; product, numeric domain, kernel, balances, content and accepted
criteria unchanged. Review by: Codex (research implementer/draft author first
filter). Recorded by: Codex. Complete five-path range after `fe1fa51d` through this
containing commit needs designated Claude review; no acceptance, archival, push or
deployment. Next specify the selected complete accounting contract, not repeat this
completed feasibility population or report the research green as server/CI green.

## 2026-10-15 — RP-652 temporal cumulative targets can decrease

Predeclare at `aa2f97df` in the owning plan before measurement: four unchanged
production/building expiry profiles ×online/offline, edge−1ms versus edge, plus
matched no-buff controls. Use real schedule resolution/application and Evaluate;
independently calculate literal payouts from the pinned factors and base rate.
No retune, debit, cap, new claims or origin mutation may explain the difference.

Focused45418 PASS0.295s: all eight buffed pairs decrease, all eight controls
increase. Online production projects `1.0008049992e4` at999ms then
`1.000115115e4` at1000ms. Every origin-relative receipt is nonnegative and
reproduces its own target; all eight representable negative previous-head deltas
refuse through actual ApplyAccrual with complete state rollback. The first
attempt compared a Decimal field to a string and did not compile; the corrected
accessor precedes the actual measurement. No production code changed.

This falsifies naive monotone **saved-payment** installation under these changing
endpoint contexts, not fixed-context conservation or valid one-shot behavior.
Fresh scratch copies are not two real player payments. The earlier constant-
context proof verifier is not extended/admitted as a temporal mutation API.
Reanchoring at expiry would change existing payout semantics, not just rounding.
Reconcile my draft's C3 with this constraint, leaving C1/C1a and full algorithm/
context/hook/activation closure open; no approval or author delegation is inferred.

Existing `TestAxisCumulative` population passes three shuffled uncached runs88645
in0.464s, retaining lower-floor refusals, proof alterations and cap/debit controls.
Cold full `make test-go GO_PACKAGES='./production ./economy ./decimal'`21606
fails only the unchanged27 TestAxisAccrualPartitionProperty leaves: Production
56.449s; Economy11.643s and Decimal0.215s pass. All failure output retained and
counted; no other failure headers. Affected vet passes. No SQL/TS/browser/Linux/
hosted or full-CI proof is claimed for this Go-only candidate constraint.

All handles terminal; no Docker services or durable resources touched. Existing
artifacts, accepted Clout body, runtime, kernel/version and original criteria stay
unchanged. Formatting/diff checked. Review by: Codex (research implementer/draft
author first filter). Recorded by: Codex. Complete five-path test/draft/plan/ledger/
log range after `aa2f97df` through this containing commit needs designated Claude
review. No adoption, acceptance, archival, push/deployment or1.0 completion.
Next close the selected context/settlement protocol, not repeat this population.

## 2026-10-15 — RP-308 explicit author delegation and wording reconciliation

Marco directly answers the pending question with "Delegate that reconciliation
(recommended)". The question expressly preserves the 24-hour offline cap, 90% rate,
banking and separate-return-visit behavior, excludes extra saved payments from
internal calculations/display reads, and delegates wording only. This is a received
human answer, not a preselected option or inferred approval.

Reconcile Clout CV3 and AC6 in the body: computational cuts of the same absence do
not renew its allowance, introduce player actions or create saved payments; genuine
separate return histories retain their existing policy. Preserve the original128
seeded cases, exact equality and all27 known failures. Reconcile my draft's C1/C1a
and the live decision/queue/ledger status with the actual delegation. No origin
lifetime, episode identifier, ledger/receipt algorithm, save version, buff proration
or implementation authority is adopted. RP-307/RP-488/RP-652 remain technical work.

Verification: inspect the seven-path documentation-only diff and factual policy
bindings against `docs/production-engine.md`; `git diff --check` passes. No software
suite is run for wording-only changes under RFC-0000. No runtime, test, kernel,
balance, CI or artifact bytes change; no services started or resources removed.
Review by: Codex (delegated reconciler/draft author, implementer first filter).
Recorded by: Codex. Complete range `fdb9a79d` through this containing commit needs
designated Claude normative review; this entry is not that verdict or acceptance
of the conservation draft. No push, archival, deployment or release promotion.
Next complete the exact accounting/context/settlement/hook/compatibility contract,
then obtain its acceptance and independent review before production implementation.

## 2026-10-15 — specify the operation-local candidate and real caller gap

After `7eb64e37`, trace actual Evaluate/ApplyLogged/Exit/catchup/claim/suppression,
the replay envelope and Company clone, and the TypeScript evaluator. Add draft C1b:
existing command tuple and pinned inputs own private context; retain Guild→catchup→
active-play→ordinary-stage order, original receipts, once-only hooks and Store rollback.
No persistent episode ID, intermediate payment, cumulative receipt or buff proration.

Explicitly retain the call-contract mismatch: the original test's cut mutates its
ledger/cursor, unlike a pure projection. A clone-only green test does not repair it.
Static tracing finds actual action-free `SimulateAdvance` callers in relevance and
first-hour waits, plus the real `contentDynamicsAccrueWindow` split check. A live
command-only design cannot claim those solved. Specify simulation branch/context/
hook ownership and the real API/caller route before acceptance; do not change reports.
Archived Relevance R10/R11 confirms the action-free wait authority and current
`{canonical_state_hash, virtual_ms}` node key. An origin-carrying branch needs future-
equivalence proof for merged nodes or an explicit reviewed key-contract amendment;
do not silently change search identity, population or goldens to land accounting.

The bounded two-profile check declared in the plan before measurement verifies this
candidate constraint through actual Company encode/restore/Evaluate. Both origin
pairs give the same complete visible state/time at the cut. Original3114/cut1553
future cash differs `1.00035846811e4` versus `1.00035846812e4`; whole-second control
2000/cut1000 retains equal full future state and cash `1.00023023e4`. All origins
remain unchanged. Record RP-653; no actual beam collision, published-report failure
or served-player loss is claimed for the current origin-free implementation.

Focused82238 passes0.310s. New identity check plus existing128-profile origin
reference pass three shuffled uncached runs54533 in0.753s; affected vet passes.
One repeat command failed shell parsing because its Make-passed regex lacked inner
quotes; corrected before the actual repeated measurement. Original partition check
remains red: complete untruncated recheck reports27 failures/101 passes in0.199s.
No broader Go/SQL/TS/browser/CI pass is claimed; no fixture, report, population,
arithmetic, runtime or version changes. Inspect the five-path test/draft/plan/ledger/
log diff; `git diff --check` passes. No services started or resources removed.
Review by: Codex (research implementer/draft author first filter). Recorded by:
Codex. Complete range after `7eb64e37` through this containing commit needs designated
Claude review, consolidatable with the preceding wording amendment. Draft remains
unaccepted; original27 failures and full1.0 obligations remain. No push, archival
or deployment. Next specify the actual API and simulation context/hook ownership,
not repeat this identity counterexample or implement the unaccepted proposal.

## 2026-10-15 — RP-654 preserve the simulation counterfactual's source version

While tracing simulation ownership, find hardcoded v14 restoration in the stock-rate
counterfactual. Under accepted Clout CV4/Relevance R10, test an admitted Company19,
owned tower and real faction hook. Old72626 refuses with unknown achievement field;
masked/no-hook controls pass. Capture the source version before the hook and restore
its pre-hook bytes at that version. Corrected17402 passes0.330s, including existing
legacy role controls: real stock1/progress60, masked stock1/progress0, no hook zero,
attainment/ownership/version retained and resulting Company19 round-trips. Kernel
0.3.216→0.3.217 records the genuine admission correction; no formula changed.

First broad runs expose three Go/two TS research source-pin failures because I put
the regression in the fingerprinted timing instrument. Move the independent test
into `simulation_version_test.go`; original timing bytes match HEAD exactly. No
artifact regeneration, changed outputs or weakened source checks. Final focused47417
passes all three research parents plus new/legacy role checks in1.458s. Final full
client75984 passes12297 with1088 explicit skips; numeric goldens included. Cold
decimal passes0.231s. Full `make verify-harness-fast`18949 passes: harness86.095s,
role population0.250s, Commons0.381s and actual baseline/history guard. Affected vet
and unchanged `make formulas-check` pass.

Final cold Production/Save/Kernel/Faction22374: Production37.220s remains red only
at the original27 partition leaves; Save0.280s, Kernel0.062s, Faction0.199s pass.
The kernel history verifier still refuses the existing50a3a514 miss, not this bump;
no global/hosted CI green claim. Native DB skips are not SQL evidence. No unchanged
browser/Postgres journey is rerun for this in-memory simulation correction. The
separate harness clone helper stays legacy-only; held axis constructors are not enabled.

Docs/plan/ledger and diff checked; no services or resources touched. Review by:
Codex (implementer first filter). Recorded by: Codex. Complete nine-path range after
`94b8af8d` through this containing commit needs designated Claude review; no feature
acceptance, archival, push or deployment. The conservation draft remains unadopted;
next finish its actual API and simulation context/hook ownership contract.

## 2026-10-15 — define the candidate calculation/branch lifecycle

At `f03cd934`, inspect actual SimulateTransition/SimulateAdvance, simulation hooks,
first-hour lifetime adapter, candidate purchase path, beam and completion cache.
Own draft C1c now specifies begin/fork/project/real-action ownership: detached views
cannot become new origins; real intents execute from the original pre-state and
run normal hooks once; refusal leaves the parent reusable. Full hook-owned state
and fresh role collectors belong to each scratch evaluation, not shared closures.

Cover all source-observed reuse paths: proposed origin/context identity accompanies
beam equality, ranked-completion cache and dominance, not beam alone. Do not hash
opaque callbacks or assume decreasing endpoint payouts support monotone affordability
search. The actual Guild/Commons chain returns events that the legacy wait seam
refuses: event ownership and once-only real delivery need an explicit contract,
not dropped events or a live outbox from a discarded projection.
Exact typed hook adapters/context encoding, temporal-search correctness and
compatibility/work bounds still require engineering closure; this is not acceptance-
ready or permission to implement. R10/R11 bodies, current APIs, report bytes, numeric
domain and original27 failures remain unchanged.

Three-path draft/plan/log diff and source bindings checked; `git diff --check` passes.
No software suite rerun for prose-only work under RFC-0000. No services/resources or
runtime/data/test/kernel bytes touched. Review by: Codex (draft author first filter).
Recorded by: Codex. Complete range after `f03cd934` through this containing commit
needs designated normative review, consolidatable with earlier accounting-draft work.
No adoption, feature completion, archival, push, deployment or1.0 promotion.

## 2026-10-15 — RP-655 event ownership and exact adapter/context proposal

Declare the two-mode Company19/Compact/Guild population in the plan at `cc626b20`
before measurement. Actual Evaluate plus the closed replay chain returns ordered
guild_tithe_accrued/compact_sampled values, positive lifetime/progress/sample state,
schema1 payloads and no invented intent. Actual SimulateAdvance on separate clones
refuses with ErrInvalidEngineState; no-hook controls advance but lack hook effects.
Origin bytes remain unchanged. This verifies the legacy wait/API mismatch, not a
current bug, new projection API, actual event delivery or repaired conservation.

Focused40590 passes0.280s. After that measurement, strengthen the positive oracle
to require positive progress and the exact Guild-tithe event instead of permitting
the zero-XP activity arm. This tightens the original ordered-event/lifetime/sample
population; it was not an extra criterion predeclared before the first run.
Final36785 runs this test, the existing six external-context profiles and RP-654
controls three shuffled uncached times: PASS0.471s. Affected vet and diff check pass.
No full Go/SQL/TS/browser/hosted rerun is claimed; unchanged runtime still has the
original27 partition failures and known historical kernel guard hold.

Own draft C1d/C1e proposes closed production-owned hook recipes, private event values
versus once-only real settlement, and strict private context/identity encoding.
Bind original operation pre-state as well as stage entry, full needed state versions,
artifacts, actual policy/seed/absence history, existing replay inputs and ablation.
No callback labels as authenticity, discarded events, rounded rates, partial commits
or changed old receipt law. Reconcile C1c/C3 wording in the body with these proposals.
Strict codec/adapter reconstruction, real Store delivery/retry, temporal search,
compatibility and measured bounds remain; no adoption or production implementation.

Five-path test/draft/plan/ledger/log range after `cc626b20` through this containing
commit needs designated review. Review by: Codex (research implementer/draft author
first filter). Recorded by: Codex. No source-pin/artifact restamp, runtime/kernel/
balance/CI changes, services or removed resources. No push, archival or deployment.

## 2026-10-15 — RP-656 bounded temporal affordability search

At `374e1899`, inspect the real horizon-first/lower-bound search and declare in the
plan five unchanged RP-652/no-buff profiles ×both modes, every endpoint0..1501ms.
Execute actual schedule resolution/application and Evaluate on fresh Company19
copies; independently check each cash amount from literal pinned fixture factors.
The hypothetical resource thresholds are not catalog purchases or new balance data.
First86628 completes15020 endpoints in2.394s: earliest production497/building1422/
combined452ms in both modes, plain869/0ms. All six buffed horizon predicates are
false despite earlier qualifying endpoints; whole-horizon monotone search misses
them. Chronological expiry-regime search agrees with the exhaustive earliest oracle.

Before the next run, prospectively declare two no-buff unreachable controls without
changing any of the original ten populations. Final67694 runs all twelve/18024
endpoints twice, shuffled, alongside existing RP-652/RP-655 parents: PASS4.719s.
Four ordinary/time0 controls and two unreachable controls retain correct tagged
results. Every known regime is exhaustively checked monotone, every endpoint's
payout matches and origins remain byte-identical. Affected vet and diff check pass.
No production per-ms scan, actual migrated harness, real purchase/SQL/TS or generic
monotonicity proof is claimed. Last full cold runtime population remains the recorded
RP-30727 failures; this test-only batch does not rerun or claim that gate green.

Draft C1f specifies the proposed boundary-aware algorithm and complete-enumerator,
hook/quote predicate and work-proof obligations. Keep C1b–C1e contracts, old API,
numeric domain, cap/buff semantics, accepted R10/R11, reports and all original red
cases unchanged. No new origin resets, proration, tolerances, hidden exclusion or
budget increase. Do not repeat this completed fixture study as general search proof.

Review by: Codex (research implementer/draft author first filter). Recorded by:
Codex. Complete five-path test/draft/plan/ledger/log range after `374e1899` through
this containing commit needs designated review. No runtime/kernel/artifact/CI
change, services or removed resources; no push, acceptance, archival or deployment.

## 2026-10-15 — separate producer and search specifications/readiness gates

At `8bef2695`, RFC-0000's scope/lifecycle rules show that the growing proposal now
amends two distinct contracts. Split search/reuse/caller migration into draft
`accrual-aware-harness-search.md`, linked from producer/index/plan. Move C1f's complete
temporal algorithm/evidence there and retain the old anchor as a redirect; move
canonical beam/cache/dominance rules into its H2. No study, result or consumer gate
is dropped. Archived Relevance R10/R11 stays byte-untouched, not silently amended.

The producer recommends operation-local origins without adoption, preserves old
saved/receipt semantics and states the proposed no-new-save-field compatibility
route. Distinguish exact caller/owner/compatibility engineering and owner/designated
RFC acceptance from later codec/SQL/Go–TS/workload implementation acceptance. Neither
an unbuilt codec nor an imaginary SQL pass can prove a draft already implemented;
neither owner approval nor the scope split waives that actual proof. The failed
intermediate-commit arithmetic remains refused; no domain restriction or tolerance.

Successor still needs the complete temporal enumerator/predicate justification,
diagnostic-owner binding and exact scenario/report dispatch before acceptance.
Full AC6/accounting closeout retains the successor's affected-consumer proof.
No accepted architecture, feature checkbox, numeric/report/runtime/CI bytes change.
Five-path documentation diff, links and moved-section coverage checked;
`git diff --check` passes. No software rerun for this documentation-only batch.
Review by: Codex (draft author first filter). Recorded by: Codex. Complete range
after `8bef2695` through this containing commit needs designated normative review;
prior review debt remains. No resources/services touched, push, archival or deployment.

## 2026-10-15 — RP-657 actual diagnostic-owner sampling and binding

At `0f4e0f16`, inspect the real ContentDynamicsActivePlayInput and natural
spawn/claim/window path. It samples contributions at each subinterval's start,
not the logged path's endpoint. Declare the registered epoch8 production draw,
real pre/post-claim branches and the5000ms literal comparison before executing.
Focused70645 PASS0.314s: the actual public diagnostic and private window helper
retain claimed35.035/control5.005/bonus30.03; actual endpoint scheduling/Evaluate
on separate copies gives5.005 in both branches. Complete origins remain unchanged.
An endpoint adapter would erase this diagnostic's bonus, not repair rounding.

Draft C1g supplies exact diagnostic owner/input/branch/window/source bindings and
window-start contribution sampling, without a fake harness scenario identity.
Reuse the existing production-owned constructor/claim prefix during implementation,
not a second authority. Reconcile C1e union/stage fields and the harness H1 body,
which had incorrectly promised endpoint sampling for this caller. Logged semantics
stay endpoint-selected. Profile1 does not enable unsupported axis mechanics.

Final focused50080 repeats this/temporal-search/event-ownership parents three times,
shuffled and uncached: PASS8.904s. Production/harness vet passes. `make content-harness`
executes both registered populations through the immutable baseline writer: existing
bytes match and none change; a differing result would refuse, not overwrite. No
new source-pin/golden/epoch bytes. No full server, SQL, TS/browser, hosted CI or
repaired27-case arithmetic claim; the owner factory and strict codec remain unbuilt.

Six-path test/draft/plan/ledger/log diff checked; `git diff --check` passes. Review
by: Codex (research implementer/draft author first filter). Recorded by: Codex.
Complete range after `0f4e0f16` through this containing commit needs designated review.
No runtime/kernel/CI changes, services or removed resources. No adoption, feature
completion, archival, push or deployment. Next caller compatibility and temporal/
dispatch closure; do not repeat this completed sampling-profile comparison.

## 2026-10-15 — close proposed sampling/identity and legacy dispatch ambiguity

At `57906e5d`, trace actual replay target binding, simulation revision inputs and
current family formats. Parent C1h now distinguishes fixed logged stage, unchanged
fixed-external primitive, explicit modern endpoint reconstruction and diagnostic
window-start recipes. Captured simulation owner/stream/log coordinates cannot be
inferred from Company bytes; a fixed logged target is not an endpoint-independent
envelope. Return ordered real stage evaluations, not an invented merged receipt.
Modern reconstruction follows the existing Guild→catchup→schedule→ordinary order
on fresh copies; no live lookup, saved projection or unsupported mechanic is enabled.

Successor H4 proposes explicit old/new protocol dispatch and schema-family pairing.
Required profile tags bind each family to its actual sampling recipe; content
Fiscal/Pitch/permit methods retain their existing non-window contracts, not an
invented global active-play policy or a report-side override.
Source actually has Relevance reports6, policy/scenario families1–2, first-hour and
content-dynamics formats1: do not invent a common v2 or retag old bytes. Preserve
base Relevance grammar without a fake balance/policy mint, existing ratified-hash
authority and RunKey shape when unchanged. A proposed registry2 shares immutable
manifest resources with one registry owner, never duplicate entry ownership or an
overwrite; scenario/report ownership and mismatch refusals remain explicit.

Four-path draft/plan/log diff and source bindings checked; `git diff --check` passes.
No software suite rerun for prose-only work. No actual schema/version/hash, artifact,
runtime/test/kernel, CI or existing data changes. All code/codec/stage/dispatch proof
remains owed after accepted specifications; full temporal correctness and producer
bindings still need engineering closure, not owner approval as a substitute.
Review by: Codex (draft author first filter). Recorded by: Codex. Complete range
after `57906e5d` through this containing commit needs designated normative review.
No services/resources touched, adoption, feature completion, push or deployment.

## 2026-10-15 — RP-658 long-absence stage reconstruction

Declare two48h Company19 profiles at `2c9f8833`: owned tower/PR, active production
buff, Compact weight and one controlled Guild clearing, online/direct offline.
First72228 is invalid: the short-interval input helper omits required catchup
metadata, and my hand-calculated cash literals are wrong. Before the next run,
use actual buildOfflineCatchup and exact integer rate115115/100000 ×86400 ×9/10
(×7 direct offline) plus initial10000. No population, factor, cap, banking,
clock or production bytes change. The original declaration/correction stay in plan.

Corrected64550 PASS0.408s. A forwarding Commons-last observer leaves actual
ApplyLogged's whole final state/receipt/events unchanged. Fresh prefix reconstruction
matches complete observed post-hook state/result/contributions and ordered prefix
events. Online has two evaluations (ordinary zero) and one full hook call; direct
offline one. Both produce24h and bank43,200,000ms, with distinct current-context
cash99,513.424/636,593.968. Clearing literals stock17/consumed10/window7/seq1 hold.
Omitting clearing keeps cash but loses full state; neither scratch path consumes
the original. This is not a mode-equivalence or served-player-loss finding.

Final24629 repeats this plus existing six Guild and42 temporal profiles three
times, shuffled/uncached: PASS1.244s. Affected vet and diff check pass. No full
server/CI, real Guild worker/SQL/TS/browser or new owner/codec proof is claimed;
runtime and original27 arithmetic failures are unchanged. Draft C1i and RP-658
record exact scope, including the invalid first attempt and omitted-clearing control.

Five-path test/draft/plan/ledger/log range after `2c9f8833` through this containing
commit needs designated review. Review by: Codex (research implementer/draft author
first filter). Recorded by: Codex. No artifact/source-pin restamp, runtime/kernel/
CI changes, services or removed resources; no adoption, archival, push or deployment.

## 2026-10-15 — remove an unnecessary protocol from the unaccepted proposal

At `d0947b4b`, recheck actual SimulateAdvance, Relevance advance/search,
contentDynamicsAccrueWindow and ApplyLogged ownership. The proposed calculation
API/codec has no runtime implementation; these consumers do not persist/transmit
origin handles. Withdraw our private JSON/hash protocol rather than build another
codec for in-memory state. C1e now specifies typed immutable ownership and opaque
run-local identity; forks share it, real action/context boundaries replace it.
Reconcile C1c/C1g/C1h, compatibility/acceptance wording and successor reuse rules.
Live ApplyLogged remains unchanged reference code, not a new wrapper's customer.

No accepted Clout wording, accounting, source, test, artifact or CI bytes change.
Retain original128 cases/27 failures, all research counterexamples, actual Go/TS/
SQL and complete-workload/budget gates. Conservative identity is not budget proof.
Both RFCs remain drafts; temporal boundary/predicate and actual caller/landing
engineering still precede adoption. Historical records remain intact. Checked full
diff and source bindings; documentation-only, no software tests rerun/claimed.
Four-path range after `d0947b4b` through this containing commit needs designated
normative review. Review by: Codex (draft author first filter). Recorded by: Codex.
No acceptance, archival, publication or release. Next resolve real-caller search
correctness rather than repeat completed fixture measurements or expand the codec.

## 2026-10-15 — RP-660 distinguish actual attendance and catchup boundaries

At `c4303ceb`, trace actual candidateCost/gateRequirementsMet, active-play schedule,
RecordOfflineSpan and buildOfflineCatchup. Current affordability uses pinned unit
prices; gates use resource-at-least conjunctions, not an invented dynamic-discount
system. Attendance's5000ms threshold is not automatic catchup's24h threshold.
Declare18 actual logged buffed profiles at the exact expiry/attendance/cap edges,
both modes. Initial50413 passes0.472s. Then prospectively add18 same-offset no-buff
controls, preserving the original18 and every expectation.

Final36-cell population plus existing long-absence and temporal-search parents
passes three shuffled cold runs67626 in9.064s; production vet passes. All payouts
match exact integer nanocash then existing canonical quantization. The forwarding
last hook leaves complete final state/receipt/events unchanged. At5001ms attendance
freezes but no automatic catchup occurs. At24h+1 online catchup changes pre-action
payout even without buffs; direct-offline cash stays equal. Both bank1ms atcap+2;
origins stay unchanged. These independent transitions are not consecutive saved
payments, a live-loss finding or a proposal to change either current policy.

Draft successor H3a binds actual predicate/clock owners; keep fixed-external and
diagnostic callers separate. This establishes a mandatory extra search boundary,
not a complete enumerator, general numeric monotonicity proof, actual migrated
search, Go/TS/SQL integration, conservation implementation or repaired27 cases.
No runtime/kernel/data/artifact/CI bytes change; no whole package/server/hosted
green claimed. Diff/formatting checked. Review by: Codex (research implementer/
draft author first filter). Recorded by: Codex. Entire five-path test/draft/plan/
ledger/log range after `c4303ceb` through this containing commit needs designated
Claude review, consolidatable with calculation work. No services/resources touched,
adoption, acceptance, archival, push or deployment. Next close the remaining actual
enumerator/numeric/work proof, not repeat these completed clock-edge profiles.

## 2026-10-15 — Derived clock-regime reference; reject the flat payout shortcut

At `42e54cfe`, declare four real-logged-probe searches with a fixed hypothetical
100000-cash threshold and24h+2ms horizon. Derive starts from original attended time,
buff expiries and the two actual policy thresholds; preserve the existing12/18024
expiry population and independently declared literal cuts. No price/content change,
production API or per-ms long-horizon scan. The independent integer ceilings are
11168956ms online buffed,78182687ms online plain,12409951ms offline buffed and
unreachable plain offline, with actual adjacent probes required.

Initial37823 stops three search arms on exact payout differences; the12/18024
parent passes. Those are valid counterexamples to my flat payout premise, not a
production repair or successful complete search. Before revising the question,
diagnostic1236 runs three retained literal pairs through actual schedule/Evaluate
versus existing AccrueConstant+ledger. Two buffed whole-rate results match the ideal
but not Evaluate; the plain whole-rate matches Evaluate and not the ideal. Counts
are unchanged. RP-661 retains all three exact negative pairs/contrasts; do not
replace the production integrator, alter payouts or add a tolerance.
The scaled-integer reference uses Decimal.New/Quantize, not exact-rational rounding;
the earlier "ideal" label is not a claim of mathematical ground truth.

Prospectively revise only the research claim: exact threshold-search results and
qualification at every probe, not a false claim that every canonical cash amount
equals the rejected ideal. Retain/report cash-model disagreement separately.
Revised54130 PASS3.791s: all four outcomes/neighbors match, both whole-horizon and
missing-cap online controls miss, both offline controls hold.186 predicate calls
include13 cash-model mismatches (repeated points count), all with identical exact
qualification. The rejected cash model has not become successful arithmetic proof.
Original12/18024 profile/payoff/census assertions remain; the new derived boundaries
also match their independent literal set. No original test population is removed.

Final70993 repeats these plus36 actual clock-edge and long-absence parents three
times shuffled/cold: PASS11.260s; production vet passes. Inspect save admission's
past/nonoverlapping spans and total-offline-duration checks; do not generalize the
offset reference past actual admitted history/canonical-clock/overflow proof.
H3b supplies the concrete candidate and both bounded positive/negative findings;
general producer/numeric/work proof and actual consumer migration remain open.
After clarifying the reference terminology, final95972 reruns all five selected
parents cold: PASS3.916s plus vet. The plan now leads with the actual next action;
completed declarations/results are retained under their history heading.

Six-path tests/draft/plan/ledger/log diff/formatting checked. Runtime/kernel/data/
artifacts/CI unchanged; no full Production/SQL/TS/browser/hosted rerun or repaired27
partition claim. Review by: Codex (research implementer/draft author first filter).
Recorded by: Codex. Entire range after `42e54cfe` through this containing commit
needs designated Claude review, consolidatable with the calculation lane. No
services or resources touched, adoption, acceptance, archival, push or deployment.

## 2026-10-15 — Numeric ordered-transform conformance, not a search proof

At `5c69475d`, inspect RFC-0001's accepted property table and the actual dedicated
Go/TS suites. Round-trip/invalid-domain properties exist, but ordered arithmetic
transforms are absent. RP-662 adds eighteen shared literal input pairs across
addition, positive multiplication/division, Quantize12 and deterministic sum.
Require exact raw and committed order, declared equal inputs, strict ordinary
controls, state validity and TS input immutability. Invalid/reversed results,
including a reversal hidden by canonical rounding, must fail the order oracle;
legitimate equal-result plateaus must pass. No production or golden output changes.

First Go compile fails on the new test's int versus float64 comparison; correct
the test type before execution. Initial Node runs98 cases:97 pass, minimum-domain
division fails because my proposed representable-output population included a
quotient below the allowed domain. Preserve that original input as an explicit
separate characterization: Go returns canonical zero; pinned JS returns raw
mantissa5/exponent-9000000000000000, rejected by isStateValue/canonicalString.
RP-663 remains OPEN: this is a real cross-runtime discrepancy, not agreement or
a repaired arithmetic defect. For the representable minimum-domain order pair,
use2e-min/4e-min divided by2; do not silently discard the initial failing input.
No assertion is relaxed with a tolerance or a new gameplay numeric policy.

Final cold `make test-go GO_PACKAGES='./decimal' GO_TEST_FLAGS='-count=1'` passes
0.236s, including all existing numeric goldens. Affected decimal vet passes.
Node `client/node_modules/.bin/vitest run --root client
test/decimal-vectors.test.ts test/numeric-monotonicity.test.ts` passes6404 tests
(100 new),0.663s. Typecheck/Svelte diagnostics pass with zero errors/warnings.
Native all-engine startup first fails sandbox localhost binding; approved local
server execution then passes Chromium/WebKit12806 tests at the earlier revision
but reports Firefox connection timeout (4/6 files, not whole-browser green).
Interrupt its lingering teardown, exit130. After the final extra reversal control,
`make test-browser-focused BROWSER_TEST_FLAGS='test/decimal-vectors.test.ts
test/numeric-monotonicity.test.ts --project=chromium --project=webkit'` passes
12808 tests/4 files in2.26s. Firefox has no passing result for this batch.

This is bounded arithmetic conformance, not whole-domain monotonicity, producer/
search proof, new accounting, Clout AC6 or a repair of the original27 failures.
No production/kernel/version/old vectors/balance/CI changes; no unrelated full
journey, DB, performance, Linux or hosted run is needed or claimed. Archived RFC/
planning remain frozen. Diff/formatting checked. Seven-path test/fixture/docs/
plan/ledger/log range after `5c69475d` through this containing commit needs
designated review, consolidatable with the calculation lane. Review by: Codex
(implementer first filter). Recorded by: Codex. No acceptance, archival, services,
resource deletion, push or deployment. Next continue actual caller/predicate and
landing closure; do not cite these eighteen pairs as a general algorithm proof.

## 2026-10-15 — RP-662 Firefox execution follow-up

The [CI-owned shared Linux runner batch](../scaffolding-and-ci/log.md#2026-10-15--shared-fullfocused-linux-browser-execution)
now executes the unchanged numeric vectors/ordered-transform files in all three engines,
including Firefox, and the complete default browser population. Native Firefox's earlier
connection failure and RP-663 remain; no arithmetic/accounting or general search proof changes.
The CI log owns commands, results, limits and the separate designated-review range.

## 2026-10-15 — RP-664 positive aggregation counterexample and actual producer

At `d81713b5`, inspect RP-663 against the Numeric contract, actual pinned library,
Go arithmeticResult and vector generator. The generator explicitly filters finite
non-state raw outputs; its passing corpus does not resolve the lower-edge state
discrepancy. Do not choose a different underflow policy by an assumed one-line fix.
RP-663 stays open; no runtime changes.

Debugging the prospective monotonicity argument finds a different valid-domain
counterexample. Seed663 XorShift exploration uses four positive15-digit mantissas
at exponents-5..8 and varies9.999999999999998→10. In100000 cases,767 raw reversals
include one Quantize12 reversal (index40044). Repeating the matched12-digit input
experiment finds none; that is a bounded control, not a whole-domain theorem.
The first attempted exploratory call mixes ESM/CJS Decimal constructor identities
and errors before measuring; correct the import before the reported counts.
Discovery counts are exploratory, not an acceptance population or operating bound.
Retain only the reduced exact inputs/results as the executable durable witness.

All four fixed raw terms can be produced exactly by canonical12-digit base rates
times ordinary exact integer counts. A variable canonical base2e-15 times count
4999999999999999/5000000000000000 produces the two raw boundary terms. The new
ephemeral five-generator catalog passes the actual loader; a legacy Company save
is encoded/restored with complete matching counts/purchase total before execution.
Actual Rates preserves the source and reproduces the exact raw terms. Actual
one-second online Evaluate plus ledger/receipt yields respectively
1.01245957117e1 and1.01245957116e1: a one-canonical-unit decrease with more ownership.
Go's raw sums are1.012459571165003e1 and1.012459571165e1. The shared TS witness
reproduces the producer arithmetic prefix and those committed values, preserving
inputs, equal/permutation and ordinary increasing controls. It is not full TS replay.

This compares different ownership contexts at the same time. My initial commentary
that clock/buff cuts were therefore insufficient for time search was too strong:
this does not prove a decreasing fixed-origin temporal predicate or current harness
miss. H3c explicitly reconciles that distinction. Within a provision bucket raw
rate aggregation is fixed; the actual caller proof must follow the integration/
rounded-segment/ledger inputs that vary, not assume a false universal helper lemma
or require a proof unrelated to that caller. Numeric repair remains separately open.

`TestCalculationSumOrderResearch` passes cold0.279s, reproducing both actual evaluator
outputs. Final cold selection with the unchanged temporal/long-clock/rejected-model
parents plus decimal goldens/ordered checks passes Production3.396s/Decimal0.212s.
Affected production/decimal vet and strict types/Svelte diagnostics pass. Node vectors,
ordered tests and retained counterexample pass6405 tests0.693s; actual focused Linux
Chromium/Firefox/WebKit pass19215 tests/9 files4.74s after cold setup. These green
characterization tests preserve a real negative result; they do not repair arithmetic.

Eight-path test/fixture/docs/draft/plan/ledger/log range after `a421860a` through
this containing commit requires independent review under its existing research/
normative gate; Marco's corrective-code reviewer delegation does not adopt this
draft or approve new arithmetic. Review by: Codex (research implementer/draft author
first filter). Recorded by: Codex. Diff/formatting checked. No runtime, kernel,
version, active catalog, authored balance, historical artifact, settlement or receipt
contract changes; no natural-player-loss, complete replay/SQL/full-CI/AC6, new owner
ruling, archival, push or deployment claim. Original27 partition failures remain.

## 2026-10-15 — RP-654 designated corrective-code review: APPROVED

Review by: fresh independent Codex sub-agent `/root/simulation_version_corrective_review`.
Recorded by: Codex parent, transcribing the returned verdict under Marco's explicit
corrective-review delegation (`1b00f8ab`); not a relabelled implementer review.
Reviewed complete nine-path range:
`94b8af8dacc8d6c06b5384bda95239a0204ff933..f03cd9346dc8052b913ab2d6612096101f9ce5b8`
(base exclusive, tip inclusive). Verdict: **APPROVED**, no actionable findings.

The reviewer independently inspects the whole diff, source-version restoration,
real faction/role regression, controls, version mirrors and records. Five focused
cold leaves PASS; a temporary overlay with exact pre-fix simulation source fails
only Company19 `real_stock_role` with unknown `achievement_score_lifetime`.
Masked/no-hook and legacy controls still pass; no repository source/test mutation.
Full cold Production/Save/Kernel/Faction/Decimal retains exactly the original27
Production partition failures; the other four packages PASS. Cold fast harness
PASS (harness86.635s, role0.178s, Commons0.372s, actual repository guard included).
Affected vet, formulas, gofmt and range diff checks PASS. Full client12398 PASS/
1088 explicit skips; typecheck has zero errors/warnings. These current-tree checks
include subsequent test-only research outside the reviewed range; affected source,
regression, version mirrors and original legacy/timing controls match `f03cd934`.
Actual `make verify-kernel-version` remains FAIL at historical `50a3a514`.

This closes only RP-654's corrective review. It does not waive either existing red
gate, adopt accounting/specification proposals, enable axis scenarios, approve
new features or other ranges, establish SQL/browser/hosted evidence, or authorize
archival, release, publication or deployment. The current plan/ledger now point
to this verdict; no product bytes change in this record update.

## 2026-10-15 — RP-664 aligned aggregation candidate rejected

At187cf426 inspect accepted Numeric Core and Deterministic Aggregation fully:
the latter prescribes grouping/combination order, not just an abstract sum.
Replacing it requires an explicit follow-up/compatibility contract. Before candidate
execution declare original midpoint pair, signed/domain controls, unchanged accrual
fixtures and100000 seeded positive profiles, separating raw/committed reversals.
The [dossier](aggregation-candidate-research.md) retains the exact method and limits.

Exact Python/Go rational diagnosis of represented inputs places both original
sums above10.12459571165. The test-only aligned Neumaier candidate returns
1.01245957117e1 for both and matches eleven valid accrual-fixture formulas. All
100000 profiles execute: current Go876 raw/1 committed reversal, Node and each
Linux browser875/1; candidate0/0 throughout. Initial JS census assertion improperly
requires Go876 and fails1/6409 cases. Correct the invented raw-equality premise,
retain each actual count and exact committed controls; no ordering tolerance added.

The added prospective signed-tail control rejects the candidate in Go/TS:
current1e300−1e300+1e-300 retains1e-300, candidate returns0 after scale underflow.
Actual exported-function overlay fails that unchanged current-result assertion
(Make2,0.212s). Its body matches the retained prototype byte-for-byte; no tracked
runtime file is patched. Earlier full overlay run96136 passes Decimal0.372s and
Economy6.139s, but Production41.194s fails27 partitions,44-profile policy-observer
drift, three changed clock-reference pairs and changed RP-664 upper payout. No
golden/report/source fingerprint is refreshed or old failure hidden.

Final active-source cold Decimal/vet pass0.378s. Original policy/clock/Rates-Evaluate
parents plus candidate controls pass23131 (Decimal0.234s/Production0.626s), with
the original one-shot outputs intact. Final affected Node6409 pass29634; types/
Svelte diagnostics zero54153. Declared Linux browser92428 passes19227/12 files,
all three engines6.57s; cold setup time excluded. Initial socket EPERM runs no
browser tests; narrow approved invocation supplies actual evidence. Final browser
Compose inventory empty; no DB/user data or durable resources touched. Ignored
compiler-overlay files remain in .cache, not a shipped implementation.

Review by: Codex (research implementer first filter). Recorded by: Codex. Complete
seven-path test/dossier/docs/plan/ledger/log range after187cf426 through this
containing commit needs designated research review, outside the corrective-code
delegation. No production/kernel/vector/balance/content/CI/specification behavior
change, numeric repair, full monotonicity/search/AC6/SQL/replay/hosted or release
claim. RP-664 and the original27 remain open. Next viable candidate must retain
cancellation tails, justify full-domain work and establish lawful replay/version
compatibility; do not repeat this rejected candidate or its positive census.

## 2026-10-15 — RP-664 sparse exact-input successor, unapplied

At35011479 construct a private Go/TS sparse radix prototype over exact binary64
input coefficients. Preserve cancellation before rounding; compress long borrow
gaps rather than visiting exponent distance. The [dossier](sparse-aggregation-research.md)
contains the proposed domain semantics, rounding/work argument, predeclared
population, independent dense oracles and actual caller census. This does not
adopt arithmetic or conflate an across-ownership counterexample with time search.

Fourteen literal controls cover empty/signed/full-domain cancellation and borrow,
remote tails around even/odd rounding ties, intermediate overflow and final-domain
refusal. All2048 seed664 bounded profiles match independent Go rational/TS exact
decimal models at raw precision, with reversal/input-ownership/work controls.
Original RP-664 pair matches both oracles at1.01245957117e1. Complete input/output
witness SHA256 ef6cdc7626f91ff237961488f84f07a022b5ac746542d53dfb68fb258768047f
agrees in native Darwin/arm64 Go1.27.2, declared Linux-amd64 Go, Node and Linux
Chromium/Firefox/WebKit. This finite evidence does not prove every admitted input.

Final `make test-go vet GO_PACKAGES='./decimal' GO_TEST_FLAGS='-count=1'` passes
0.449s; `make test-go-ci CI_TEST_PACKAGES='./decimal' CI_TEST_FLAGS='-v'` passes
0.751s. Verbose legacy vectors exceed tool output; named new populations, digest
and terminal package pass are retained, not a full per-vector skip census. Node
five affected files6412 PASS; declared focused Linux browsers19236 PASS/15 engine-
file executions6.01s. `make typecheck` has zero errors/warnings. No hosted claim.

Actual compiler overlay replaces only aggregate.go, using the same prototype
helpers after nominal suffix renaming (byte-compared). Ignored overlay source
SHA256 f83a37cfbf80d52ad22a20db64f062b53be5a40c1c27e3b845bdc8787feba9fb.
`make test-go GO_PACKAGES='./decimal ./economy ./production'
GOFLAGS='-overlay=/Users/stronk/repos/cloud-clicker/.cache/sparse-aggregation-go-overlay.json'
GO_TEST_FLAGS='-count=1'` exits Make2: Decimal0.835s/Economy9.655s PASS,
Production45.140s RED. Original27 partitions remain; the44-profile observer,
three retained clock-reference pairs and the RP-664 upper payout change. No
internal guard fires, golden/report is refreshed or performance improvement claimed.
This unversioned overlay is an experiment, not an accepted manifest/kernel.

Declared Compose inventories end empty. Only this run's tmpfs Postgres container
is removed with the declared service command; no durable DB volume/user data is
removed. Ignored compiler-overlay files remain. Seven-path test/dossier/docs/plan/
ledger/log range after35011479 through this containing commit needs designated
research review, outside Marco's corrective-code delegation. Review by: Codex
(research implementer first filter). Recorded by: Codex. Production/kernel/vector/
balance/CI/RFC bytes remain unchanged. RP-664, RP-663,27 partitions, H3c and AC6
stay open. Next actual caller/version/activation and workload closure, respecting
existing Company drift/deferral rather than inventing a legacy arithmetic fork;
separate minigame/catalog/report identities still need their own disposition.

## 2026-10-15 — RP-664 actual caller compatibility and cost

Afterf3280ff2 test two genuine activation unknowns, without changing production.
The [sparse dossier](sparse-aggregation-research.md#caller-compatibility-findings-after-f3280ff2)
owns method, literal fixture, source-bound compatibility constraints and timings.
Seed665's declared at-most20000 four-card search reaches a disagreement at384;
the candidate matches independent rational rounding. Shared literal metrics/factor
enter both actual catalog loaders. Go registry Create/Apply and TS create/apply
keep the real seed2 hand with synthetic admitted hack ownership. Current1.0.0
clears funding at5.97255552034e3; the actual Go overlay instead yields
5.97255552033e3/playing and fails the named regression (Make2). This disproves
unchanged minigame semantics, not reachable-player loss or a published-content bug.
No tracked runtime byte is mutated; original helper/overlay hash stays unchanged.

Actual AccrueConstant/Evaluate benchmarks execute all declared intervals and ledger
credits, with fresh save restoration per Evaluate iteration. Two clean active/
candidate rounds,three100ms samples per five rows, show material allocation and
constant-cost overhead; eight-source accrual medians345–353 versus2961–2973ns,
4 versus157 allocations. One intervening active round overlaps package checks/vet
and is excluded; a prior one-source-only smoke is not a matched comparison.
No budget, pacing limit, CI timeout or acceptance threshold is adopted/loosened.
Initial benchmark invocation loses its lone dollar to Make and executes no tests;
the retained command uses `-run=^$$`. No invalid invocation is counted as evidence.

Cold native full Decimal/Pitch/Commons/Replayverify pass and affected vet passes.
Final native Decimal/Pitch full passes0.492/0.333s before the final literal-metadata
binding; final focused Decimal0.301s/Pitch0.081s covers that binding, with affected
vet clean. Production compiles but has no selected ordinary tests in that focused
run; its real caller execution is the benchmark above. Node six affected files
118 PASS; strict types/Svelte diagnostics zero; Linux Chromium/Firefox/WebKit
12 PASS/6 engine-files2.52s with unchanged2048-profile digest. Initial native listen
EPERM and sandbox Docker denial run no browser tests; the declared CI-focused
Docker target supplies actual evidence after narrow escalation. Declared Linux-
amd64 Go runs all sparse populations and the actual Pitch transition cold:
Decimal0.322s/Pitch0.069s, including the same profile384 and raw witness digest.
That run precedes only the final fixture-metadata binding assertion, not any math
or Pitch input change. No SQL behavior is claimed. Its own verified tmpfs Postgres
container is removed; both declared Compose inventories finish empty. No durable
volume/user data removed. Ignored compiler-overlay files remain.

Ten-path test/fixture/dossier/docs/plan/ledger/log range afterf3280ff2 through this
containing commit requires designated research review, not corrective-code
self-approval. Review by: Codex (research implementer first filter). Recorded by:
Codex. No runtime, kernel/version, catalog/balance, RFC, CI or baseline bytes change.
No AC6/monotonicity/full-workload/SQL/hosted/release claim. Next numeric step is
an explicit activation follow-up preserving actual identities and measuring cost
after proven-equivalent reductions; the separate H3c caller-specific time proof
remains available without waiting for a universal aggregation repair.

## 2026-10-15 — H3c actual predicate reduction and burst carry attack

At288ff35c trace all three real Relevance predicates, actual integration and the
closed hooks. Quotes/thresholds are fixed during a wait; completed bucket deltas
are fixed within endpoint context, while the last bucket's canonical delta varies.
The closed hooks do not subsequently mutate the resource ledger. H3c now states
that concrete remaining numerical obligation, rather than requiring the false
arbitrary-raw-rate sum lemma. This is source-bound reduction, not its final proof.

Before measurement declare20000 seed666 near-flat burst/carry profiles. All20000
cross at adjacent milliseconds; none has a decreasing grouped/quantized total.
The test refuses invalid/zero aggregates, missing crossings and incomplete real-
subject controls. First/last profiles match four actual Evaluate/ledger results
from unchanged origins under admitted retuned v4 catalogs and v17 saves. The
initial two-stage instrument's actual-subject check fails on missing Company
achievement ownership in its promoted v17 save; initialize that required empty
map before the successful run, not an assertion/domain relaxation.

Cold related clock/long-search/whole-rate/temporal-search/sum-order parents plus
new experiment pass Production4.049s; Decimal compiles with no selected tests,
not a claimed golden-vector rerun. Affected Production/Decimal vet passes. Final
new experiment including positive/valid aggregate checks passes0.474s and vet.
Commands use root make test-go with count1 and the exact names in the current
plan/RFC. No DB/container, TS/browser/hosted run, exhaustive-ms scan, independent
full arithmetic oracle or current Relevance migration is claimed. No runtime,
kernel, balance, schema, golden, receipt or CI behavior changes; the original27
partition failures and AC6 remain open.

Four-path test/draft/plan/log range aftera0550e14 through this containing commit
needs designated research/normative review, outside the corrective-code delegation.
Review by: Codex (research implementer/draft author first filter). Recorded by:
Codex. Both drafts remain unaccepted. This bounded attack is complete, not a
reason to repeat or enlarge a positive census; next establish the actual canonical
partial-prefix/ledger order argument or specify a correct alternative with its
work/compatibility proof. No new owner product ruling, archival, push or deployment.

## 2026-10-15 — RP-663 lower-domain arithmetic conformance correction

Accepted Numeric Core Boundary Hardening §3/§4 makes the pinned client the diagnostic
oracle and excludes lower-domain results from state/wire. Under RFC-0000's repair
route, preserve finite invalid diagnostics in arithmeticResult instead of returning
valid zero. Normalize its coefficient independently, preserving canonical zero and
later scaling that rescues a representable final result. Constructor/state domains,
quantization, aggregate algorithm, balance, saves and minigame identities do not change.
Kernel0.3.218 records this genuine arithmetic behavior correction; no accounting draft,
origin/receipt protocol, sparse candidate or general power/exp fallback is adopted.

Parent first runs the new regressions at015ae5e4, whose runtime matches the9d49e0bd
review base (tool session51794): nine raw diagnostic rows and the original ordered-boundary case fail
with Go(0,e0); both actual Ledger.Apply/ApplyAccrual arms wrongly return nil; three
actual AccrueConstant refusals report got0/want error. Controls pass. Node already
passes the same raw/accrual cases. These are executed parent failures, transcribed
here; no separate raw on-disk artifact or independent reproduction is claimed.
Final seven-case accrual supplement also checks later scaling back into domain;
that extra control was not in the initial old-code run. Reuse existing Go/TS accrual
runner; keep the original16 primitive vectors and6296 numeric goldens byte-unchanged.
The initial three failures run under TestAccrualVectors before the supplement is
separated into TestAccrualLowerDomainVectors, preserving the old16-case corpus.

Full affected checks expose real consequences, not just a green narrow suite:
the original lower-domain research had counted silent-zero projections as successes.
Keep all20 cells and explicitly assert new refusals/no receipt/unchanged complete
Company state:19 cumulative proofs plus one refused projection;18 complete deferred
work-cut sequences plus two refusal sequences. The additive census is now one refused
projection, nine invalid additive deltas and ten exact commits, no wrong target.
Historical success entries remain evidence for their old kernel, not current claims.
The initial full run is RED on these stale assumptions, original27 partitions and
three source-bound report identities; no claim of only27 failures for that run.

Re-execute existing instruments using required UPDATE_PARTITION_RESEARCH/UPDATE_RATE_RESEARCH
switches on root make test-go, then UPDATE_ANCHOR_RESEARCH. All original report results
are identical; JSON diffs change only Decimal source hashes and anchor's dependency hash.
The original16 accrual corpus remains unchanged rather than expanding its historical
research population. This is conformance-test maintenance, not new research adoption.

Verification (cold Go count1, repository-root commands):

- Full Decimal/Economy/Pitch/Commons/Kernel/Save/Replayverify/Faction/Guild/Commonsbinding
  pass; full Production42.273s remains RED only on the original27 partition cases after
  the explicit correction/re-observation. Affected vet passes. No weakened bound or
  repeated unchanged partition run to seek a last green. Focused corrected regressions
  and complete20-cell refusal populations pass natively and in declared Linux-amd64.
- Node nine files7908 pass; same numeric, ordered/lower-domain, accrual, Economy, replay
  and three research files pass23724 across Chromium/Firefox/WebKit in declared Linux
  browser container. Initial sandbox Docker denial runs no tests; narrow escalation
  supplies actual evidence. Containers approximate Linux, not hosted hardware/full CI.
- make vectors-check regenerates6296 byte-identically; make fuzz-ci passes4875945
  executions/31.208s. Types/Svelte zero errors/warnings, build/UI boundaries pass.
- Declared Linux-amd64/Postgres scoped Minigame start/resolution/history/rollback/retry
  population also passes on corrected Decimal. This is an affected consumer check,
  not a new underflow-through-SQL witness. Selected Pitch has no matching test in that
  narrow run; its full native suite above supplies Pitch coverage.
- make verify-kernel-version remains RED at historical50a3a514/RP-131 after checkout
  contract/adversarial fixtures pass. Current version mirrors agree; no guard exemption,
  rewrite, historical fix, whole CI or release claim. Final diff/whitespace checked.

Review by: Codex (implementation first filter). Recorded by: Codex. Complete corrective
range after9d49e0bd through this containing code/tests/fixtures/version/docs/record
commit is submitted to fresh independent `/root/numeric_lower_domain_corrective_review`
under Marco's delegation. Independent verdict pending; no self-approval or archival.
Next retain that verdict and continue accepted construction; RP-664, AC6/27 partitions,
draft adoption, full-nine-tier scope, hosted CI and exact-release proof remain open.

## 2026-10-15 — RP-663 scoped independent approval; fallback follow-up remains

Review by: fresh independent Codex sub-agent `/root/numeric_lower_domain_corrective_review`.
Recorded by: Codex parent transcribing its returned verdict under Marco's corrective-code
delegation. **APPROVED**, no actionable findings, complete21-path exact range
`9d49e0bd842fcf7044555fc2e715511474b43bb9..3c521cab1bda8b008ac5c07d6ad1212cfa4bab49`.

Reviewer independently executes full cold Decimal/Economy/Kernel and focused actual
accrual/lower-domain/partition/anchor/rate parents, affected vet, types/Svelte0/0,
Node7908 tests, vectors-check6296 byte-identical and exact-range whitespace. Full cold
Production45.252s is RED only on27 original partition divergences. Kernel guard's
checkout/adversarial fixtures pass; historical scan remains RED at50a3a514. Initial
unmatched Node filename selections are corrected with actual vector/Economy files;
no claim of tests from absent files. Tested Decimal hash58190097 matches committed
source/report bindings; final record edits are inspected. Old failures remain parent
transcript evidence; reviewer verifies identical old preimages, does not reexecute
them. No independent browser/Postgres/fuzz/build/full-consumer/hosted execution is
claimed; parent evidence above retains its own provenance and boundary.

This approves the shared arithmeticResult fix and maintenance, not every underflow
path. A subsequent parent pinned-Node check actually returns raw(1,-9000000000000002)
for `1e-4500000000000001` squared and(1,-17999999999999998) for minimum-exponent
input squared. Untouched Go Pow/pow10 fallback source predicts valid zero there;
no cold Go reproduction of those two fallback paths has yet run. Keep RP-663 OPEN
for that follow-up rather than claiming blanket parity. Next add/execute its bounded
regression under the same accepted diagnostic contract and determine the correction;
no owner policy is inferred. No accounting/sparse candidate adoption, partition
repair, original-feature acceptance, archival, push or release approval follows.

## 2026-10-15 — RP-663 representable power/exp fallback correction

Resume cleanf4158b6c; prior turn made implementation/review progress. Accepted Numeric
Core/Boundary Hardening §3/§4 governs this repair, not either accounting draft. Before
runtime edits, root cold regressions38338 expose11 new numeric failures and six new
actual Ledger.Apply/ApplyAccrual mistaken accepts; original division controls pass.
`(4e-6000000000000000)^1.5` is not merely lost to zero: old Go returns valid
`1e-8999999999999998`, while pinned Node returns invalid raw(8,-9000000000000000).
Other retained safe/unsafe-integer exponent, sign, fractional, pow10 and exp cases
return old Go zero. Node25 passes before correction. Failures are root-executed
tool-transcript evidence, not a claimed independent historical reproduction.

Preserve the lower fallback result when its normalized exponent fits int64, using
JS Number exponent-shift addition before the conversion. Existing in-domain paths
stay unchanged. Reached diagnostics require safe Add gap comparison and checked
Mul/Div exponent arithmetic: no wrap back into valid state; genuine zero and valid
quotient/Add/Sub recovery still work. Kernel0.3.219 records genuine numeric change.
The24-case shared fixture has20 literal raw checks and four explicitly classified
transcendental diagnostics (finite/nonzero/sign/normalized/invalid state+wire), not a
loosened valid-value tolerance or replacement for6296 goldens. Initial int64-edge
test assumes coefficient1; actual normalized coefficient1.298... disproves that
instrument premise. The final test checks the required normalized diagnostic and
its actual arithmetic, not invented bit equality for a transcendental intermediate.

Reexecute existing partition/rate then anchor instruments with their UPDATE switches
on root make test-go. All numerical report outputs stay identical; only source and
partition-dependency hashes change. Current lower-domain refusal populations and
all original27 partition failures remain; no research/balance/accounting adoption.

Verification: final cold Decimal/Economy/Kernel pass; broad affected consumer run
passes Pitch/Commons/Kernel/Save/Replayverify/Faction/Guild/Commonsbinding, Production
RED only at27 partition failures64.553s. That broad run precedes only the final Add
diagnostic-gap guard; final Decimal/Economy/Kernel and affected vet pass after it,
and fresh reviewer separately executes the full final Production population. Final
root Node nine numeric/Economy/replay/report files7920 pass; same selection23760
passes in declared Linux Chromium/Firefox/WebKit. Final source-bound report checks
use Decimal0f20973d. Linux-amd64 selected raw diagnostics, diagnostic arithmetic,
actual ledger refusals and accrual controls pass. This is native ledger/Company math,
not underflow-through-SQL; selected Kernel has no matching tests, its full native
suite passes. Declared Compose starts disposable Postgres even though these selected
cases do not use it. Full types/Svelte0/0, build/UI boundaries and whitespace pass.
Vectors-check6296 byte-identical; fuzz-ci6188653 executions/30.341s passes (before
final Add guard, unchanged normalization/round-trip boundary). No whole CI green,
hosted hardware, pacing, deployed/user loss or release proof is claimed.

**DESIGN-GAP remains beyond int64.** A temporary observation file, removed after its
cold execution, runs minimum-exponent input.Pow(2048) and FromString(-1e20).Exp:
Go raw(0,e0), finite=true/state=true for both. Pinned Node instead returns finite
raw(1,-18431999999999997952) and(1,-43429448190325178368), invalid state. The
observation asserted current outputs, not conformance; do not count its pass as
correctness. No private diagnostic representation, guessed sentinel, NaN policy or
domain expansion is adopted. RP-663 stays open for an explicit follow-up contract;
the scoped code correction cannot resolve a representation/classification choice.

Review by: Codex (implementer first filter). Recorded by: Codex. Complete range
afterf4158b6c through this containing code/test/fixture/version/docs/record commit
is submitted to fresh independent `/root/power_fallback_corrective_review` under
Marco's corrective-code delegation. Final verdict pending; no self-approval,
archival, content mint, push or deployment. Next record that verdict and propose
the small exponent-limit follow-up while continuing other accepted1.0 work.

## 2026-10-15 — power correction independently approved; representation draft only

Review by: fresh independent Codex sub-agent `/root/power_fallback_corrective_review`.
Recorded by: Codex parent transcribing its returned verdict under Marco's narrow
corrective-code delegation. **APPROVED**, no actionable findings, every changed path
in exact15-path range
`f4158b6c5521b59c3deeb966f9c590558d21a15d..1d1b244a82ff14fe141e95af632cbf79c404a265`.

Reviewer inspects complete final code/tests/fixture/version/docs/report/records
against pinned source and accepted numeric contracts. Independently cold full
Decimal/Economy/Kernel and affected vet pass; final-source full Production50.457s
is RED only at original27 partition cases. Eleven selected Node files7767 pass;
types/Svelte0/0,6296-vector byte-identical regeneration, final ledger regression and
range whitespace pass. Kernel checkout/adversarial tests pass, historical scan
still RED at50a3a514. Final Decimal source0f20973d matches refreshed report bindings.
Old failures remain root transcript evidence; reviewer does not reexecute them.
No independent browser/DB/fuzz/build/hosted result is claimed. Approval is only the
int64-representable correction, not fullRP-663, accounting, RP-664, AC6, release or
any future normative exception. Parent records and preserves these boundaries.

Separately draft `rfc/numeric-diagnostic-exponent-limits.md` and index its real status.
Recommend an explicit NaN diagnostic exception for lower results beyond Go's exponent
representation, preserving exact rejection/state/wire and all representable-result
rules. Alternative is an expanded representation with measured compatibility/cost;
neither is silently chosen. The draft names the two executed counterexamples,
conversion/normalization/zero/atomicity/cross-runtime acceptance and existing holds.
No runtime, version or research output changes in this record/proposal commit.
Review by: Codex (draft author, NOT designated normative review). Recorded by: Codex.
Normative review/adoption remain required before implementation; the independent
corrective verdict above does not cover/adopt this subsequent draft. RP-663 remains
open. Continue accepted work; no whole1.0, archival, push or deployment claim.

## 2026-10-15 — Bound the missing canonical partial-bucket argument

H3c's completed20000-profile attack used at most287 completed buckets; catalog admission
permits positive safe-integer provision ticks, not that287 bound. Do not repeat the positive
census or extrapolate it to a proof. Next bounded diagnostic asks whether a canonical partial
delta crossing10 can decrease the actual grouped/Quantize12 sum of an unchanged canonical
prefix. Use prefix counts100/1000/10000/100000/1000000 and literal full-bucket values
20,20.0000000001,20.0000000003,23.4567890123,29.9999999999; partial9.99999999999→10.
Every input must be canonical, positive and unchanged. Actual client sum/quantize is the
subject, not a replacement implementation. Include known RP-664 raw reversal and increasing/
equal controls; count all25 pairs, fail on incomplete execution or invalid output. This is
a structural aggregation counterexample diagnostic, not a temporal evaluator experiment,
benchmark, whole-domain proof, adopted algorithm or production change. Any reversal must
then be checked in Go and for realizability through an admitted fixed-origin evaluator;
no-find closes only this diagnostic, not H3c. Do not grow the population after seeing results.

Executed directly in Node26.11.1 against actual client numeric.ts/pinned break_infinity2.2.0
(strip-types; no reconstructed summation), source312f6e02: all25 pairs finish, no decrease;
every projected pair is canonically equal. Full input mantissa/exponent arrays stay unchanged.
Three separate controls pass: increasing/equal pairs and the existing raw RP-664 reversal
1.01245957117e1→1.01245957116e1 from numeric-sum-order-research-v1.json. An initial control
invocation mixed independently loaded ESM/CommonJS Decimal constructors and throws before
its comparisons; corrected CommonJS identity supplies the three control results. The25-case
run uses objects produced by numeric.ts's own parseCanonical and is unaffected by that mistake.
Routine output is complete. No source/fixture/test/report/kernel edits, Go/Evaluate/SQL/browser
or workload claim. Close this bounded diagnostic without expanding it; it does not supply the
still-required actual caller proof or alternative search algorithm. Both drafts remain
unaccepted;27 partition failures and all compatibility/budget/review gates remain.

## 2026-10-15 — Narrow the temporal arithmetic proof without another census

Inspect actual Go aggregate.go and TS numeric.ts coefficient folds. H3c now gives a
source-level sublemma: for positive finite coefficients staying in one exponent group,
the sorted ordinary-addition fold is nondecreasing in one variable. Fixed-position
composition is monotone; at a crossed fixed coefficient both orderings coincide exactly,
so the inequalities join. This proof does not rely on the sampled287-bucket limit and
allows plateaus; it proves neither exact summation nor complete payout monotonicity.
Normalization, Decimal alignment/addition, group/zero/grid transitions, scalar integration
and the final ledger/caller still need their actual argument or a correct alternative.
RP-664's cross-group counterexample and all27 accounting failures remain. This is an
unreviewed engineering refinement to the unaccepted draft, not implemented search,
normative adoption or a new empirical study. No production/test/fixture/report/version
bytes change and no software rerun is warranted solely for this prose. Review by:
Codex (draft author, NOT designated normative review). Recorded by: Codex. Next is the
remaining actual normalization/group-boundary proof, not expansion of positive samples.

## 2026-10-15 — Declare bounded sparse-candidate identity-path optimization

RP-664's unapplied candidate has measured48 allocations for one-source accrual; its exact
zero/single-input identity needs no sparse integer accumulator. Test a Go-only fast path for
length0/1, validating before return and preserving canonical positive zero. Multiple inputs
and the entire arithmetic/rounding fallback remain unchanged. Preserve the original overlay
and its hash as the comparison arm, not overwrite historical evidence. This is prototype
engineering, not activation, a new numeric contract or progress on H3c's missing caller proof.

Before measurement: require existing14 edge controls,2048 independent-rational/raw profiles
with unchanged witness digest, original RP-664 pair and Pitch compatibility counterexample.
Add raw identity/invalid/signed-zero/full-exponent controls against zero-padded general sum;
no payout tolerance or input exclusion. If these pass, use the existing five actual-caller
benchmark rows,100ms/sample,count3, two sequential rounds of old candidate→optimized candidate
→active implementation. Wait for other owned Go/browser/SQL work to finish before timing;
record complete rows, medians and allocation counts, machine/source identities and limitations.
No complete-workload, hosted or adoption claim from short native measurements. No repeated
positive census or hidden production/CI/budget/golden/kernel change is authorized.

## 2026-10-15 — Sparse identity-path result: cheaper, still unadopted

The Go test-only prototype now short-circuits length0/1 with validation/canonical positive
zero; all multi-input arithmetic stays byte-unchanged. Exact identity supplies the proof of
this transformation. Raw tests compare a forced-general zero-padded sum and literal input,
including signed zero, invalid/unnormalized/diagnostic values and full-domain endpoints.
Focused sparse native passes0.260s; full cold Decimal0.337s/vet pass. Declared Linux-amd64
30477 passes0.234s: identity/14edges/2048 rational+reverse+ownership/original pair/retained Pitch
counterexample, unchanged ef6cdc7626… raw witness digest. No new TS/browser/runtime behavior.

Original ignored overlay f83a37cf… is untouched; mechanically copied optimized overlay SHA256
60dbfd2966571baa51a38054160c278fe24e494d05104849eb42d84b6679842a replaces only aggregate.go
in manual builds. Both actual Company/Pitch witness pairs30036/55358 fail IDENTICALLY on old/
optimized overlays, preserving known incompatibilities; active87195 passes both. A mistaken
whole sparse-test overlay run makes the current/candidate comparison identical and correctly
refuses a missing disagreement; it is invalid comparison evidence. First narrowed dollar
regex fails shell parsing before tests; corrected four-parent overlay selection passes0.163s.

After all reviewers explicitly release Go/SQL/browser work, six sequential benchmark arms
complete (two old→optimized→active rounds,5rows×3samples×6arms=90), no omitted/failed row.
Actual Darwinarm64/Go1.27.2/AppleM3Max metadata from benchmark; hardware sysctl denied.
One-source candidate48→0 allocations,1s median1021–1028→57.47–57.97ns,24h1030–1041→62.01–62.54ns.
Restore/Evaluate220/242→76 allocations,4853–5004/4912–4975ns optimized medians. Eight-source
157 allocations stays157 and medians3048–3199ns versus active356.9–360.9ns: major remaining
cost, not adoption readiness. Dossier retains full table/recipe; ignored raw artifact SHA256
19026d95da52dfd9a21e14deb0ba847f111a28745b0b9ead63f37eec588a868d retains every output/command.
Short sequential samples are not confidence intervals, hosted timing or complete workloads.

Review by: Codex (prototype/test implementer first filter, NOT designated review).
Recorded by: Codex. Full research/test/record range begins38dfc002-exclusive through this
containing commit. No product/canonical-doc/kernel/CI/golden/owner-copy changes; actual
27 partitions/history failures stay open, not rerun or fixed by these measurements. Caller/
version/catalog/Commons/harness dispositions, independent review/adoption and actual full
workloads remain before activation; H3c's general temporal proof remains separate. All test/
benchmark handles terminal. Initial service inventory empty; after all checks, inspected
95c2085e… has this checkout's exact Compose labels, no mounts and tmpfs-only DB. Stop/remove
only that created service; final inventory empty. Disposable test data is unrecoverable,
no SQL used in this experiment; caches/volumes/networks/preexisting builders preserved.

## 2026-10-15 — Declare exact sparse scratch-storage reuse

The previous identity-path candidate still allocates157 times for actual eight-source
accrual. Reuse call-local big.Int scratch for coefficient/radix division and carry normalization;
retain distinct owned map entries and copy each digit before reuse. Replace only exact
10^r for Euclidean r0..8 by its bounded int64 product, not the arbitrary-width coefficient
or accumulated digits. Reuse exact threshold/remainder storage in final rounding. Local
Go1.27.2 math/big docs explicitly permit receiver/operand aliasing; no shared mutable cache,
fixed-width accumulation, changed rounding, work bound or arithmetic contract is introduced.

Before measuring: the full existing identity/14edge/2048 rational+reverse+ownership/pinned-
digest/original-pair/Pitch counterexample population must pass, native and declared Linux;
affected full native Decimal/vet too. Preserve both previous overlays, create a separate
ignored exact-helper overlay, and compare actual Company/Pitch compatibility outcomes.
After reviewer Go/browser/SQL work terminates, run two sequential identity→reuse→active
benchmark rounds with unchanged5 real-caller rows,100ms,count3. Retain all90 rows and
allocation counts, no selective exclusions or CI/performance/adoption claims. This is
unapplied prototype engineering; H3c and caller/version/full-workload gates remain open.

## 2026-10-15 — Scratch reuse preserves exact outputs and reduces multi-input cost

Full native Decimal/vet passes0.435s. Focused sparse0.159s, overlay four-parent0.261s and
declared Linux-amd64 46891/0.230s pass. All identity/14edge/2048 rational/reverse/ownership
profiles and original pair remain exact, ef6cdc7626… witness unchanged; same profile384
Pitch disagreement retained. Actual Company/Pitch overlay92918 stays RED on the same two
incompatible outputs, not a production regression or adopted fix. General scalar/prefix
monotonicity proof remains separate. Original/identity overlay digests f83a37cf…/60dbfd29…
unchanged. New mechanically generated ignored reuse overlay SHA256
4b9e897a9c79a1b38fda998f06a240e0c5596e56e85aaa525c3b255110849304 matches all test helpers.

After reviewer release and root Linux completion/cleanup, all six sequential benchmark arms
finish with90/90 rows, no failures/exclusions. NativeDarwinarm64/Go1.27.2/AppleM3Max,100ms/count3,
two identity→reuse→active rounds. Eight-source medians3011–3024→1791–1803ns,157→41 allocations,
3832→1160bytes. Active355.3–358.4ns/4 allocations, so substantial cost remains. One-source
zero allocations and Evaluate76 remain; small timing changes there are not improvement proof.
Dossier has full table/recipe; ignored raw artifact SHA256
b5dd0226955041a4a8aad8ebdaaf724a7860f3fd52576661c6c1eb964b128784 retains all outputs/commands.
Short sequential native samples are not confidence intervals/hosted or full-workload evidence.

Only local scratch/exact bounded power/storage reuse changes in the Go test prototype;
no active source/TS/fixture/golden/kernel/CI/threshold change. Review by: Codex (prototype
implementer first filter, NOT designated review). Recorded by: Codex. Complete new research/
test/record range starts064c54ee-exclusive through this containing commit; remains unadopted.
Next close caller/version/catalog/Commons/harness dispositions, full workloads and independent
review/adoption, not another uncontrolled timing run or silent global replacement. All handles
terminal. Initial service inventory empty; root-created8b5abfec… verified exact checkout labels,
no mounts/tmpfs-only, stopped/removed after tests. Disposable unused SQL data unrecoverable;
final service inventory empty, caches/volumes/networks/existing builders preserved. No release.

## 2026-10-15 — Specify the exact-aggregation caller and activation proposal

Review by: Codex (draft author first filter, NOT designated normative review). Recorded by: Codex.
At38cba707/product77fe2275/kernel221, rederive13 direct Go/7 direct TS product call sites
and their actual ledger/tenant/catalog/report/projection boundaries. New
rfc/exact-represented-aggregation.md proposes a checked exact operator, explicit migration
table and two STATIC pinned grouped consumers (catalog admission and Pitch1.0.0). Reuse
accepted Company drift; no permanent per-run arithmetic fork, global replacement, new
engine version, private origin protocol or owner-authored copy change.

Two important concrete bindings: resourceTotals currently has no error result, so invariant
errors must propagate rather than become an empty/partial valid report. Commons refreshScope
also smooths health/writes evaluated_at; the proposed startup capacity cutover must update
ONLY derived capacity transactionally before admission, preserving all other fields/events.
These are source-derived engineering constraints, not new deployed-loss findings.

New-format Harness producer identity uses actual kernel authority, not another algorithm
registry. Code-only baseline migration needs this explicit prospective guard amendment,
separate artifact-only commits, real preceding behavior/version change, historical producer
bindings and complete same-population before/after proof; unchanged current guards stay
binding. Old report bytes/readers/inputs remain intact. Reconcile own UNACCEPTED parent/search
text in the same batch: same-kernel one-shot compatibility versus separately adopted numeric
references; wait/accounting protocol versus numeric producer identity. Neither draft accepts
another, changes today's reference or fixes27 partitions. Full-domain review, workload,
actual Go/TS/SQL/retry/cutover and feature evidence remain required.

RFC readiness is explicit agreement/normative review/owner adoption of these contracts;
full implementation/activation evidence is not demanded before code may be authorized.
Check actual source-to-proposal links, current old baseline law and diff/whitespace. No
software suite for proposal-only Markdown; no code/test/fixture/report/kernel/CI/secret/
deployment/publication change. Exact eight-path draft/record/index range begins38cba707-
exclusive through this containing commit; designated normative review/adoption PENDING.
No material gameplay/release milestone or new implementation authority. Next pursue the
still-missing real temporal proof and normative review, not more positive sample censuses.

## 2026-10-15 — Refuse infinity-class constant-accrual recovery (RP-671)

Outcome: align the client constant-rate helper with accepted Production Accrual Math's
non-finite-intermediate refusal and Numeric Boundary Hardening's pinned classification.
Six new shared upper-bound vectors first pass unchanged Go; two fail unchanged TS
(c0937a, expected throw but accepted). Maximum valid rate×60/1000 seconds reaches
the positive-infinity class, then nonzero1e-6 efficiency incorrectly recovers client
state. Existing Go refuses. Correct only the client primitive: preserve validated zero
efficiency, check the accrued intermediate's classification before scaling. Do not
reject finite lower-domain intermediates or change pinned Decimal arithmetic.

The earlier suspected ordinary burst operation-order discrepancy did NOT reproduce:
4096 seed671 profiles at default speed, then the same4096×six speeds, have no raw
effective-seconds difference. Division by1000 shifts the exponent in the representable
range; no rounding defect or general proof follows. Initial malformed generated rate
and direct Node extensionless-import failures were instrument failures, not product
evidence. Burst replay's upper-bound caller is NOT yet reproduced/adjudicated here.
This helper feeds prediction (efficiency1); current replay owns a separate formula.
No receipt/event/persisted encoding or watched path changes, so no kernel bump, registry
extension, golden restamp, balance retune or historical-guard exception is introduced.

Executed final checks, cold Go via root Make:
- Native Go six upper vectors pass (3729); all29 production vectors and existing
  exponent-boundary controls pass (34811). Full Decimal passes0.315s; Decimal/production
  vet passes. No production Go arithmetic changed. Full production's known AC6 failures
  are not rerun or claimed repaired.
- Node seven affected files, including shell prediction, all numeric goldens, lower
  diagnostics and existing partition/anchor/rate consumers:7628 pass (60218).
- Native Chromium/WebKit six suites/12718 executions pass (5555). Initial sandbox
  port EPERM ran no tests; permitted local-server run above succeeded. No native
  Firefox or Linux/browser/performance/full-journey/hosted-CI claim.
- make typecheck build-client passes (68015): zero diagnostics and actual worker/client
  build. Initial focused Node run exposes three stale source identities (1321 other
  assertions pass); re-execute existing Go partition/rate then anchor generators via
  UPDATE_PARTITION_RESEARCH/UPDATE_RATE_RESEARCH/UPDATE_ANCHOR_RESEARCH Make variables.
  All complete, preserving542/64/615 populations and every result. Independent parsed
  comparison removes ONLY source_sha256 and requires deep equality against HEAD:
  unchanged observations; only production.ts and dependent partition-artifact identity
  change. These are new-source re-observations, not relabelled old raw measurements.

Review by: Codex (implementer first filter, NOT designated review). Recorded by: Codex.
Complete correction range begins4d847020-exclusive through this containing commit;
source, shared vectors, tests, three re-observations, docs and records remain pending
fresh independent corrective-code review. RP-663, RP-664, H3c and27 partition failures
remain separate. No feature acceptance, normative adoption, deployment or publication.

## 2026-10-15 — RP-671 independent corrective review approved

Review by: /root/constant_accrual_upper_corrective_review (fresh independent Codex,
Marco-delegated corrective code). Recorded by: Codex (root, transcription).
APPROVED full4d847020aca6650ff6c294c4d1a6b6b5759c9e53-exclusive through
76c1a57d0a8e39c945fa18ea3d489148c880fbd0-inclusive, one commit/all11 paths.
Accepted Production Accrual Math and Numeric Boundary Hardening govern this correction;
no new contract, watched path or receipt encoding change. No kernel bump required.

Independent in-memory old source accepts exactly the two forbidden payments
6e8999999999999994 and1e8999999999999996; corrected source matches all six rows.
Initial import lacked TS stripping, ran no population; corrected stripTypeScriptTypes
run establishes reproduction. No tracked source mutation. Cold root Make executes29
Go accrual vectors; Decimal in that selector runs0 and is NOT credited. Separate full
Decimal0.365s/vet pass. Read-only three research generators reproduce542/615/64 rows
and committed bytes. Independent parsed comparison preserves all observations except
source_sha256, and every current source hash verifies.

Node affected six files1324 plus numeric goldens6304 pass. Native Chromium/WebKit four
files/eight suites12728 executions pass, including prediction-worker boundary. Initial
port EPERM runs0; approved local bind succeeds. Typecheck/build zero diagnostics and
actual worker/client build. Exact range whitespace passes. All reviewer handles terminal,
no tracked edits/commits/Docker/publication. Root's later burst work is outside this range.
No Firefox/Linux/hosted/performance/DB/natural-player loss/burst replay/fullProduction/AC6/
new-feature/normative/archival/release approval. RP-671 correction closes, other gates remain.

## 2026-10-15 — Correct infinity-class burst replay recovery (RP-672)

Outcome: conform the separate client burst-accrual caller to Doctrine/Compute Credit D7's
byte parity and existing Numeric Boundary Hardening infinity classification, not change
the core's diagnostic representation. Six small shared rows retune the existing unminted
doctrine corpus, rehash/admit the actual bundles and restore v17 saves before ApplyLogged.
They are extreme conformance inputs, not current balance or a natural-player loss finding.

Old TS64045 resolves both forbidden transitions with actual receipt/events and payments
6e8999999999999994 and1.2e8999999999999994; three positive controls pass. First Go58316
does NOT reach arithmetic: preserved pretty-printed fixture payload is not canonical.
Canonicalize with existing helper, then22894 executes all five actual Go transitions:
two accrual refusals with identical encoded state/no output, three successful controls.
Do not count the first instrument failure as refusal evidence. Add the sixth zero-efficiency
control; both runtimes pass6713/8781. Refusal checks require accrual error, not any failure.

Correction: same seconds-first calculation as Go, rate-sum validation and existing zero
behavior, then reject infinity-class effective seconds or pre-efficiency payment. Keep
finite diagnostic recovery intact. The valid one-second maximum-speed control explicitly
prevents rejecting a transient milliseconds overflow that Go's order avoids. Normal and
near-edge positive payments/cursor/burst consumption remain exact. Existing ordinary
ApplyLogged rollback restores full state on these throws. No new API/guard/balance/copy/
sum algorithm or parent accounting adoption. True replay behavior correction bumps kernel
221→222 in all three sources, not to disguise a runtime-neutral file edit.

Executed final boundary checks:
- Full Node85728:145 files/12493 assertions pass,23 files/1094 skipped NOT credited.
  This includes full191-entry-test replay population and all numeric golden consumers.
- Native Go69614 cold: numeric goldens0.440s, six real burst cases/existing burst+offline
  cap/unchanged full Go-authored replay corpus0.704s, version-source check0.162s. Vet passes.
  Verbose golden output truncates displayed middle rows; final package/exit0 is read.
- Declared Linux-amd6454534 executes29 accrual vectors, six actual ApplyLogged cases,
  burst/offline cap, unchanged full replay corpus and version source: PASS. Decimal's
  selector runs0 here and is NOT extra numeric coverage. This target's Postgres starts,
  but no SQL test is selected; no persistence/HTTP/default-player journey claimed.
- Frozen-install declared Linux browser12206: four files×Chromium/Firefox/WebKit,
  12 suites/19650 executions pass, including all191 replay tests and numeric goldens/
  lower-domain/refusal controls. DaemonLinuxarm64, not hosted runner hardware.
  Initial socket EPERM creates no resources; narrow allowed Docker run above succeeds.
- Typecheck/build44707 zero diagnostics, actual worker/client artifacts. Vet passes.
  verify-kernel-version86920 remains RED at historical50a3a514, no waiver or CI-green
  claim. Checkout-contract fixtures pass before it; separately42493 passes full guard
  adversarial fixtures. New version-source test passes; no history rewrite/guard change.

Review by: Codex (implementer first filter, NOT designated review). Recorded by: Codex.
Full code/shared fixture/tests/docs/version/records range begins3e6e6e68-exclusive through
this containing commit; pending fresh independent corrective-code review. All commands
terminal. Initial/final Compose inventory empty, same two existing BuildKit containers.
Root-created3b6ce4a3… twice inspected exact checkout/service labels, no mounts/tmpfs-only;
stopped/removed only it after completion. Disposable unused SQL data is unrecoverable;
preexisting builders/cache/volumes preserved. No new general temporal proof, AC6/27-partition
repair, draft adoption, full CI, feature acceptance, clean-host, deployment or publication.

## 2026-10-15 — RP-672 independent corrective review approved

Review by: /root/burst_replay_upper_corrective_review (fresh independent Codex, delegated
corrective code). Recorded by: Codex (root). APPROVED all11 paths/one commit,
3e6e6e681ebf1cd2fb67e23dddace7382ebcde33-exclusive through
137f7a8d7c4fcafe19536b23271945fb7090352e-inclusive; execution at exact endpoint.
Accepted D7/classification conformity, not a new numeric contract. Three identities222 bind
real behavior change. Safe Vite loader of exact baseline SHA380147f1675aab4a6125d818fccfaa59902b03cb31cd7242135c8d91f51b467a
independently reproduces two wrong receipts/events (6e8999999999999994,1.2e8999999999999994)
with four valid controls. Initial relative-config startup and wrong Node filenames are
instrument failures, explicitly corrected; no main-source mutation or extra coverage claimed.
Cold root Make verifies six Go transitions, burst/cap/full replay fixture/goldens/kernel,
separate full Decimal and vet; actual five-file Node6605, native Chromium/WebKit replay382,
typecheck/build and complete-range whitespace pass. Initial browser EPERM ran0; narrow
retry passed. Full history guard stays RED at50a3a514; its checkout/adversarial fixtures pass.
All reviewer commands terminal; ignored diagnostic config only, no tracked edits/commits/Docker.
No independent Firefox/Linux/SQL/default-player/natural-loss/AC6/normative/feature/archival/
release approval. RP-672 corrective review closes; other accounting/1.0 gates remain.

## 2026-10-15 — Check the remaining canonical group-carry proof boundary

H3c's fixed-exponent sublemma does not cover a canonical partial term crossing9.99999999999→10
into the next exponent group. Diagnose that specific boundary, not repeat seed666's positive
burst census: five canonical full-bucket coefficients from burst fractions250/333/500/667/750
of1000ms, with equal prefixes2^16−1/2^16/2^16+1. Compare actual grouped sums before/after,
raw and Quantize12, with input-order controls. A reversal would refute the proposed proof
step; it becomes a temporal finding only after actual same-origin admitted Evaluate confirms
reachable partial terms. No-find is no general proof and does not authorize search or arithmetic.

Initial Node diagnosis completes15/15: six raw reversals, zero committed reversals.
For coefficient3.003003003e1/prefix65536, raw coefficients at exponent6 fall
1.96805804804446→1.9680580480444572; both commit1.96805804804e6. Thus raw monotonicity
is already false with canonical inputs, though the player predicate is not disproved.
Next isolate the quantization seam at that pair with one fixed positive shim near5.4e-7
(the gap to its next Q12 midpoint), fixed literals5.38/5.39/5.40/5.41/5.42e-7. This
changes the proof question, not a substitute larger temporal census: confirm any committed
counterexample through the actual aggregator, then determine whether real bucket histories
can produce that prefix. A synthetic prefix is not itself a reachable gameplay history.

All five small-shim cases align to the same raw1.968058048045e6 in both arms: the extra
Decimal Add removes the normalization-path distinction. Instead perturb one EXISTING
full-bucket term in the same exponent group, retaining65536 prefix entries and the same
partial pair. Target3.003003057e1 and its two adjacent canonical values each side
(±1/2 units of1e-10 in value), derived from the same midpoint gap. This directly tests
whether that already-observed raw reversal can survive Q12 without an additional alignment.

Those five cases retain raw reversal but Q12 rounds the exact lower endpoint tie toward
even…04, so both still agree. Test the SINGLE predicted next midpoint by adding1e-5 to
that modified prefix bucket (3.003004057e1): the tie should now round toward…06 while
the smaller raw arm remains below it. This is a literal rounding-parity prediction,
not a wider search population; retain the outcome whichever direction it takes.

The predicted odd-tie case DOES reverse Q12:1.96805804806e6→1.96805804805e6.
Retain it in one compact shared fixture and actual Go/TS tests with reversed-order,
input-ownership, positive/equal and independent decimal-literal rational controls.
Reachability check is ONE derived profile, not a census: rate3.003003003e-11,
speed1e12,1000ms ticks,65536 full buckets,333ms burst portion in the partial bucket.
Find its adjacent partial canonical9.99999999999→10 pair, compare actual raw sums,
then reuse admitted v17 Evaluate/ledger controls at both endpoints. Expect raw reversal
but equal committed payments for the unmodified prefix. The synthetic odd-tie case's
extra unequal fixed bucket is NOT silently assumed reachable from this origin.

## 2026-10-15 — Canonical group-carry proof shortcut refuted; real boundary retained

Shared65537-term fixture repeats65535×3.003003003e1, fixed3.003004057e1 and partial
9.99999999999→10. Actual Go/TS grouped Q12 falls1.96805804806e6→1.96805804805e6.
Independent exact decimal-literal sums196805804805661999/1e11→196805804805662000/1e11
increase; reversed-order/immutable-input and positive/equal controls pass. Native raw
upper coefficient differs from Node by one representational step; no raw byte-parity claim.
One DERIVED actual admitted v17 origin with equal full buckets reaches raw reversal at
adjacent499/500ms; fresh Evaluate/ledger projections both pay1.96805804804e6, origin
unchanged. The synthetic unequal prefix's reachability and actual decreasing resource
predicate remain UNPROVED. H3c/plan now reject the generic canonical-input proof shortcut,
not the still-required actual caller proof. Sparse candidate is not activated or benchmarked.

Executed: full native Decimal0.479s/vet; native two Production research tests0.932s, including
unchanged20000 population solely because the shared admission helper gained a speed argument.
Node6306/50950 and types7342 pass. Linux-amd6495886 executes both new cases0.158/1.322s;
Linux Chromium/Firefox/WebKit85245 executes both retained research tests,6 passes. No selected
skip. New witness is research/test-only: no numeric golden, source math, report, CI, balance,
kernel222, owner copy or accounting implementation change. No full Production/27-partition,
closed modern replay/hook, SQL, natural-player loss, H3c or feature-completion claim.
Review by: Codex (research/draft author first filter). Recorded by: Codex. Complete research/
test/draft/record range startsda38175d-exclusive through this containing commit; designated
research/normative review/adoption remains outside corrective-review delegation and pending.
All handles terminal. Only root-owned twice-inspected tmpfs72cd9ece… stopped/removed after
tests; unused disposable SQL data unrecoverable. Initial/final Compose empty, same two
builders, caches/volumes preserved. No publication, deployment, archival or release approval.

## 2026-10-15 — Clout cap-presentation containment independently approved (RP-536)

Review by: /root/clout_cap_presentation_corrective_review (fresh independent Codex under
Marco's corrective-code delegation). Recorded by: Codex root, transcription. APPROVED complete
30050a9364209feadfecf2a7877405eef176077a-exclusive through
373396d439fdf6eaa2d28cb36e85e30a24cb03d6-inclusive, one commit/all6 paths. Accepted
CV9/GS0.5: actual zero-parameter copy signature, existing contained alert/descriptions,
once-per-error reporting/healthy recovery and operational independent Desk controls. No new
copy, numeric, decoder, request or policy. Current panel/host/copy resolver match endpoint;
later keyboard diagnostics in test are inspected context, not enlarged approval.

Independent ignored Vite overlay loads exact original panel via git show without tracked
mutation: all16 native Chromium/WebKit correction executions fail, saturated RangeError/
TypeError and below-cap missing alerts. Initial config/import errors and sandbox listener
refusal run no tests. Old completed run's ancillary screenshot filesystem denial is not
the behavioral failure cause. Corrected Linux ARM64 axis/screens54143 includes all24 new
three-engine correction cases; axis81 CLI passes split78 ordinary+3 expected blocked-Close
controls (NOT deadlines). Node105, types0/0, shipped client build and boundary verification
pass. Isolated native render budget16788 passes5,105 unrelated deselected,5.58s.

Retained RED: combined Linux54143 exits unsuccessfully,392 ordinary+3 expected controls
pass, one ordinary Firefox failure and15 performance exclusions. Screens `late http-conflict
completion cannot refresh an unmounted host`, line235, sees requests0 rather than1 before
unmount/late-completion assertions. No axis_stack means the corrected panel is unmounted;
failure cause remains unknown. No reassurance rerun/assertion adjustment. Root records RP-673
in backlog/CI log; it is not a passed consumer file or full CI. Native Enter/readiness/lifetime
must be diagnosed rather than inferred from absent dispatch. Exact census clarified from
definitions/terminal summaries, no rerun.

Real browsers/DOM/runtime with synthetic decoder-admitted wire arms, not actual minted
producer state/public exploit, SQL/composed or assistive technology. Linux daemon ARM64,
not native hostedamd64. No full CI, authored content, numeric/accounting adoption,27-partition
repair, whole Clout, archival or release approval. All handles terminal; reviewer changes
no tracked bytes/commits/source; ignored reproduction retained. Final Docker inventory is the
same two preexisting builders. This closes only RP-536 corrective review; full1.0 remains open.

## 2026-10-15 — Target exact decimal midpoints in the actual equal-prefix carry (before measurement)

H3c/RP-664 question: does the admitted equal-prefix burst caller's known raw decrease cross
a committed Q12 boundary at analytically selected prefixes? Keep rate3.003003003e-11,
speed1e12, tick1000,333boosted partial ms and adjacent499/500ms from the retained control.
For nominal full bucket30.03003003, at sum exponent6 the Q12 unit is1e-5. Literal
N*30.03003003+10 is exactly halfway when3N≡500(mod1000), hence N≡500(mod1000).
Execute exactly53 prefixes33500..85500 step1000, all inside the existing24h horizon.
Use actual accrueBoostedConstant/grouped reducer; count raw reversals and committed differences,
no excluded cases. A committed decrease must then reproduce via real admitted v17
catalog/save/Evaluate/ledger and immutable origin before any caller claim. Positive/equal
results do NOT establish monotonicity; stop this bounded falsification instead of extending
a passing census. Research only, no active algorithm, accounting adoption, balance or golden change.

Result65465: root make test-go GO_PACKAGES='./production'
GO_TEST_FLAGS='-run TestCalculationHalfwayCarryResearch -v -count=1' exits0;
all53 selected profiles execute0.10s,19 raw reversals, ZERO committed reversals.
Thus this midpoint selection does not expose a committed decrease. The conditional real
Evaluate confirmation runs ZERO times; no new admitted-origin/ledger proof is claimed.
This is native Go1.27.2/Darwinarm64 only, no TS/Linux/SQL/cross-runtime claim and no
general monotonicity conclusion. Stop the population here: further positive prefixes would
not supply H3c's missing actual-caller proof. Existing canonical negative and equal-ledger
origin stay unchanged. Preserved manual source probes/halfway-carry-probe.go.txt is outside
normal CI, gofmt-clean; not a new passing acceptance test. No product, math, kernel222,
balance, fixture/golden, algorithm/adoption or normative text changed. Review by: Codex
(research author first filter). Recorded by: Codex. Range9a551e61-exclusive through this
research commit awaits research review outside the corrective-only delegation. Next remains
the general reachable-prefix/quantization/ledger argument or separately adopted arithmetic
route, not a larger midpoint census. All handles terminal; no Docker resources for this probe.

## 2026-10-15 — Source ownership, legacy encoding and writer guard corrections approved

Review by: /root/clout_admission_guard_corrective_review (fresh independent Codex under
Marco's corrective-code delegation). Recorded by: Codex root, transcription. APPROVED all3
bounded corrections, complete contiguous e4054647746af3a09683252851a3027dea1d3407-exclusive
through a7e33b73a39a435157c7049a8d13abdd68a63c4d-inclusive, all16 changed paths:
ed129ec5ec08c2835758919c18385198ce127a21/RP-538/CV1,
e7ad35910cd027faf365b38911330a12ede4471a/RP-539/CV4,
a7e33b73a39a435157c7049a8d13abdd68a63c4d/RP-537/AC4. All code/tests/docs/records and
3 markers inspected: real admission changes187/188; guard-only batch no bump. No balance,
schema/migration/workflow/authored-copy change, new mechanic or original feature approval.

Independent old Go loader overlay reproduces4 collision refusals with2 distinct-ID controls.
Old TS shim bound to e4054647 reproduces4 ownership+9 legacy refusals,9 selected controls pass.
Old Gaia helper fails6 parenthesis/address cases;3 direct-write/2 ordinary-read controls pass.
Its failure on new codec baseline is NOT further old-regression evidence: old collector
excluded the codec. Corrected typed codec cases and actual collected-source mutations test
the new boundary. No tracked mutation; ignored read-only shims/overlays retained.

Actual current source9a551e61: cold economy/save/decimal/kernel81444 and full multiplier/
achievements/replaycatalog/routes28225 pass, Decimal vectors included. Cold four Gaia/16legacy/
real Founder codec8790 passes; scanner345 non-test sources, nonzero v17 values1/999/
MaxExactInteger retain bytes and existing v18 control passes. Actual activation/derivation/
formula-attainment/carry/receipt3019 passes. Initial dependency selectors ran0: only full28225
gets credit there. Node74054 passes6617 across8 actual files; nonexistent initial axis-stack-
vectors selector gets no credit, corrected axis-formula-vectors executes. Declared Linux
ARM64 browser82813 passes327/15engine-file populations, all3engines. Vet/types0/0/build/API/
formulas/client boundary/gofmt/complete-range whitespace pass. Wrong first old-TS config path
and sandbox-denied browser attempt execute no tests; corrected invocations supply evidence.

Regression/guard files and TS ownership loader remain endpoint-identical. Later Go admission
and TS replay changes retain these exact guards; activation discloses historical replay-source
drift, not a fresh historical checkout. No fresh SQL/composed player/hostedAMD64/full CI,
whole-program alias/reflection/complete reward-union/AC4/full Clout/archival/release approval.
Original P1–P5 negatives,27 accounting failures, history guard and research/adoption holds
stay untouched. All handles terminal; reviewer edits/commits no tracked bytes, disposable
browser gone, original2BuildKit preserved. These three bounded ranges leave corrective debt.

## 2026-10-16 — One analytically selected nonaligned-origin carry (before measurement)

H3c next bounded question: can a reachable initial partial grid segment expose a committed
decrease hidden by the equal-full-prefix control? One profile only, not a larger positive
census: unchanged rate3.003003003e-11/speed1e12/tick1000, start cursor667ms after run start,
first333ms fully boosted,65500 full buckets, then adjacent499/500ms with333ms boosted.
First and lower final segments both canonicalize to9.99999999999; full buckets30.03003003.
The exact decimal sum lies at the Q12 midpoint neighborhood because65500≡500(mod1000).
This targets a nonuniform prefix admitted by actual grid alignment, not an arbitrary list.

Manual probes/offset-carry-probe.go.txt stays outside normal CI and is added through an
ignored compiler overlay. Require valid retuned v4 catalog, encoded/restored v17 origin,
both actual Evaluate results/receipt cash equal the declared grouped/quantized sequence,
unchanged source origin and no invalid/guard cases. Count both projections even if no
decrease. A decrease is bounded admitted-origin evidence, not natural acquisition or a
current harness miss; equality is not a proof. No further profiles authorized by a positive
result. No active math/kernel/balance/golden/CI change or draft adoption.

Result61491: the declared root Make/ignored overlay selector executes one profile and TWO
actual admitted/restored v17 Evaluate/receipt projections, terminal exit0/package0.849s.
Raw totals equal(1.96698696696338,e6); both committed cash amounts1.96698696696e6. Thus no
committed decrease in this analytically selected initial-offset control. Complete source
origin remains unchanged; no guards/exclusions. Stop, do not extend the passing population.
Native Go/Darwinarm64 only; no TS/Linux/SQL/natural acquisition/current-search/universal proof.
This does not establish that the earlier unequal-prefix reversal is reachable or repair any
of27 partition failures. Manual source is gofmt-clean/outside CI; no active production/kernel222,
balance/golden/protocol or normative text changes. Review by: Codex root, research author first
filter. Recorded by: Codex root. Complete new research range8f4b50d0-exclusive through its
containing commit requires research review outside the corrective-code delegation.

Final source61973 additionally executes an independent math/big.Rat decimal-literal reference:
1966986.96696499998→1966986.96696499999, both immediately below midpoint1966986.966965.
Actual grouped raw totals/committed payouts remain equal; package0.906s, two actual Evaluate
projections again execute. This exact reference explains the selected inputs, not accepted
engine arithmetic or a further population. Reproduction uses a Go overlay mapping virtual
server/production/calculation_offset_carry_probe_test.go to the retained manual source, then
make test-go GO_PACKAGES='./production' GO_TEST_FLAGS='-overlay=/Users/stronk/repos/cloud-clicker/.cache/offset-carry-probe-overlay.json -run TestCalculationOffsetCarryResearch -v -count=1'.

## 2026-10-16 — Exact-run prototype study: retain patch, do not replace the reference

Sparse dossier declares the one bounded optimization/method before execution: coalesce
adjacent equal-exponent signed binary significands exactly, then unchanged digit/rounding
algorithm. No Decimal intermediate or cache. Full native Decimal0.472s and vet pass;
focused existing/new tests0.269s retain identity/14edge/2048 rational/reverse/ownership,
original pair and profile384 consumer mismatch. Raw witness still
ef6cdc7626f91ff237961488f84f07a022b5ac746542d53dfb68fb258768047f.
New65537-input positive/cancellation/invalid-tail and full-domain controls pass, ≤3 digit
splits. Declared Linux-amd6495819 repeats actual focused checks0.250s, terminal0.
This is not SQL evidence; its otherwise-unused service creates only owned tmpfs data.

New ignored aggregate overlay94837f76fd64b91003539b7ae7e71294d266d7047fc0690522879661253d7c7a
mechanically copies experimental declarations with suffix renaming/unchanged panic wrapper.
Affected overlay vet passes. Actual Company/Pitch48132 intentionally FAIL at the same retained
payout/funding differences; no compatibility repair or activation. Original compatibility
tests/goldens/production are not edited. A first filename glob fails before execution and
a guessed-file search fails; neither supplies test evidence. All credited selectors execute.

Only after reviewer/heavy work terminates, execute existing BenchmarkAggregationCallerResearch
via root Make, NoTests/100ms/count3/benchmem, sequential old reuse→new runs→active, all45 rows.
Raw artifact .cache/sparse-aggregation-runs-bench-0c8ff6c1.json SHA256
205b555d33e632b613b27951332ed748ed586c1620e2cdbf5bf498308043ef7f retains commands/all rows,
all3 terminal0 and derived medians. Sparse dossier holds table/limitations: eight-source
41→43allocations,1160→1216bytes;1780→1722ns is not a statistically established speedup.
No long-list wall-time/full-workload/hosted estimate. No further short tuning census.

Do not adopt this variant by default. Restore ordinary test-prototype source exactly to
5ec45ccbb841e346cdb25f6d09e6bda73a43a6ab08aba9b0ea230fdcb51abd4d; actual production and TS
stay unchanged. Preserve manually applicable probes/sparse-runs-candidate.patch SHA256
292a74432338d797cd7a0670fdf6afb3848db310b972a2f839fe07526a533ca1, based on0c8ff6c1.
git apply --check passes; compiler-overlay candidate/raw evidence retained ignored. The patch
is outside ordinary test collection, not a new CI population. No kernel222, balance, report,
golden, architecture/adoption, H3c/27-partition or release promotion. Only root-owned twice-
inspected tmpfs1eb1d24b… stopped/removed after terminal; its disposable data is unrecoverable,
two existing builders and durable data/caches preserved. All handles terminal.

Review by: Codex root (research author first filter). Recorded by: Codex root. Complete
research range8597577d-exclusive through this containing commit needs research review outside
the corrective-only delegation; no self-approval. Continue actual caller/workload/compatibility
closure, not more short microbenchmarks. Independent arithmetic/contract adoption remains held.

## 2026-10-16 — One provision-generated unequal prefix (before measurement)

Different actual caller shape from completed equal-prefix/offset censuses: a provision
boundary changes ownership once and then reaches its hardcap. One analytically selected
v4 catalog/v17 origin, TWO independent actual Evaluate projections only. Tick1000ms,
origin1ms after run start; first999ms,65535 full buckets, adjacent final499/500ms with333ms
boosted. Speed1e12. Beige base3.003003003e-23, purchased999999352334, one provider initially
adds647666 provisioned units at the first boundary, target hardcap647666. Its own positive
production goes to permits, not cash. First canonical cash delta should be2.999998057e1,
then all full buckets3.003003003e1 and partial9.99999999999→10. The first value is the retained
synthetic unequal bucket minus0.03006, selected to preserve its odd-midpoint parity while
remaining below the maximum first partial. No growing prefix or rate/balance search.

Require real catalog admission, encoded/restored origin, actual provision counts/cap, exact
declared segment/sum/evaluator/receipt agreement and unchanged origin in BOTH projections.
Report raw and committed order whether equal/increasing/decreasing; a differing sorted-fold
result is not a failed measurement. A passing no-decrease stops this profile, not a universal
proof. A decrease supplies admitted-origin evidence, not natural acquisition/current beam
failure. Manual source stays outside CI under ignored compiler overlay. No active math,
kernel222, balance/content/golden/CI, accounting implementation or draft adoption.

First execution15872 terminal0/package0.944s: BOTH actual admitted/restored Evaluate/receipt
projections execute. Raw(1.968058017995,e6)→(1.9680580179949971,e6); committed cash
1.968058018e6→1.96805801799e6. RP-676 records this immediately. The original v4 fixture
still has purchased-count milestone/feed bonuses that direct Evaluate does not assemble;
do not mislabel it an executed full harness wait/factory. All actual provision counts and
the unchanged source-origin check pass. No natural acquisition or live-payment-loss claim.

Before further execution: preserve that result, add independent literal-rational order and
one wrapper compatibility control, not a larger time/rate census. Same numeric profile and
TWO same endpoints through actual SimulateAdvance, with nil hook/default mask and no external
inputs, in a separately admitted catalog clone removing only the active target's ladder and
manual/feed roles/feed membership. Keep its existing stock role, all other rows, provision
and numeric inputs. Require actual content contribution assembly empty and the same known
cash decline/receipt/provisioned cap. This neutral clone prevents secretly bypassing derived
contributions. It is a different controlled catalog, not today's published economy or the
complete modern endpoint factory/hook/search. Stop after these two wrapper projections.

Wrapper result35143 terminal0/package1.676s: both original Evaluate and both actual neutral
SimulateAdvance projections execute, same decline, same provision cap/role and receipt.
Actual content assembly is empty as required. Independent literal-rational reference increases
196805801799661999/1e11→196805801799662000/1e11. No source-origin mutation. This now rejects
a generic monotonicity claim even at the real action-free wait primitive, not the complete
modern factory/hook/reuse/search or a current published report.

Next bounded counterfactual, before execution: bind the retained UNAPPLIED scratch-reuse
exact represented-input aggregate via its compiler overlay, leaving this numeric/profile
input unchanged. Separate explicitly named ExactCounterfactualResearch test expects both
cash amounts1.968058018e6 and preserves literal/receipt/ownership/provision controls. Two
Evaluate plus two neutral SimulateAdvance endpoints only, no benchmark or wider population.
The ordinary research entry retains today's decreasing pair and cannot pass as an exact
candidate result. This can identify whether the separate numeric proposal addresses THIS
counterexample; not full temporal monotonicity, candidate review/adoption or caller activation.

Counterfactual32616 terminal0/package1.590s: exact prototype yields equal raw
(1.9680580179966198,e6) and equal cash1.968058018e6 in all four real projections.
Its bound source .cache/sparse-aggregation-reuse-candidate.go SHA256
4b9e897a9c79a1b38fda998f06a240e0c5596e56e85aaa525c3b255110849304 is the retained
scratch-reuse prototype, not the later rejected default replacement or active math.
Final ordinary-source84950 terminal0/package1.544s again executes all four projections,
preserving the decreasing pair. No skipped/conditional/guard population or benchmark.

Exact reproduction from repository root: create an ignored Go overlay mapping virtual
server/production/calculation_provision_carry_probe_test.go to the committed manual
probes/provision-carry-probe.go.txt (both absolute paths). Ordinary command:
`make test-go GO_PACKAGES='./production' GO_TEST_FLAGS='-overlay=/Users/stronk/repos/cloud-clicker/.cache/provision-carry-probe-overlay.json -run TestCalculationProvisionCarryResearch -v -count=1'`.
For the counterfactual add the absolute server/decimal/aggregate.go→the SHA-bound prototype
mapping and select TestCalculationProvisionCarryExactCounterfactualResearch instead.
The exact prototype's source derivation remains in sparse-aggregation-research.md; its
private panic wrapper/changed Company/Pitch compatibility/cost are NOT adopted authority.
Final manual source SHA256d4148e2df242488a64007946ed716366b6eec2eae4c3f26ddd0bacc633b9e44f.

Reconcile OWN draft H3/H3b/H3c, live plan, defect ledger and current strategic summaries:
the context-only monotonicity proof route is false for this admitted family. Next is a
correct alternative search/partition or separately adopted numeric route with actual full
caller/compatibility/work proof; no repeated positive census or altered acceptance bounds.
This does not prove natural acquisition, today's complete search/report failure, full
modern factory/hooks, TypeScript/Linux/SQL, repaired accounting/27 partitions, full CI or
any tier/release. No active product/kernel222/balance/golden/CI bytes or resource mutations.
Review by: Codex root (research author/first filter). Recorded by: Codex root. Complete
56f8a278-exclusive through this containing research commit needs research/normative review
outside the corrective-only delegation. All handles terminal; formatting/diff checked.

## 2026-10-16 — RP-676 TypeScript public-replay boundary (before execution)

Next test is TWO independent public TypeScript ApplyLogged calls at the same499/500ms
endpoints, not another rate/prefix census or private-function export. Reuse the neutral v4
catalog recipe and admitted v17 ownership/cursor/burst coordinates. Foundation fields come
from the existing doctrine genesis and pinned foundation artifacts from first_content_exit;
the freshly hashed complete10-artifact subset still passes actual bundle validation.
Spend1 banked compute millisecond after accrual (no cash debit/credit), require applied,
known cash/receipt pair, actual provision cap, restored independent origins,4999 credits and
new1ms burst. This includes actual replay hooks but is not the Go full logged caller, SQL,
modern v19 factory, current search, natural acquisition or live loss. Manual test is mapped
onto the existing replay test slot by ignored Vite loader; no tracked client/Go source changes.

First92505 fails before bundle loading: fs path assumed client cwd although the root command
keeps root cwd. Correct the data path, not the environment. Next9691 admits the bundle but
fails meter-state validation before either ApplyLogged: the doctrine genesis's input-remainder
keys belong to its OTHER pinned meter catalog. Neither run supplies a logged outcome. Use
actual newRunMeterState(bundle.meters,0) for the new pinned catalog's complete values/decay/
input-remainder sets; do not weaken state validation or splice in missing ad-hoc fields.

17902 then executes the whole single test/two public ApplyLogged calls, terminal0/1.47s.
Final source63838 repeats both after fixture typing/immutable-artifact construction cleanup,
terminal0/1.49s (861ms test work). Both applied: exact cash/receipt decline, provision cap647666,
unchanged source genesis,4999 remaining credits and new1ms burst; actual neutral contribution
assembly is empty. Fresh10-artifact bundle admission executes. No private export/instrumentation
of the evaluated module. The ordinary source origin's foundation fields are NOT relabelled as
this separately initialized full v17 state. Actual module hooks execute, but no complete
Go/TS state/event equivalence, SQL, v19 factory/search, natural acquisition or live loss claim.

Exact root command:
`client/node_modules/.bin/vitest run --root client --config /Users/stronk/repos/cloud-clicker/.cache/provision-carry-replay.config.mjs`.
The ignored config merges client/vite.config.ts, includes only test/replay.test.ts and loads
the committed manual probes/provision-carry-replay.test.ts.txt into that existing test slot.
It does not replace product sources or run the ordinary full replay population. Source SHA256
1f838350e540dd096d39f5feae375dbdf034171804938dd20765b1e4dbb4bccb. Gofmt not applicable;
source/diff checked, no claim that .txt research was included in the project's TS typecheck.
No product/math/kernel/balance/CI/artifact-baseline change or new proof population for CI.

Own draft/plan/RP-676 reconcile the stronger bounded evidence. Review by: Codex root (research
author/first filter). Recorded by: Codex root. Complete eac09e1c-exclusive through this
containing research commit needs research/normative review outside corrective delegation.
No independent approval, adoption, full accounting/27-partition repair or release milestone.

## 2026-10-16 — Exact candidate: first full standard-pacing measurement declared

After final TS63838 terminates, declare the single full300 native exact-candidate phase
in sparse-aggregation-research.md. No competing heavy check/reviewer is running. The manual
profile overlays only the existing benchmark to retain its full canonical result outside
the timer, with no-overwrite output; actual suite/12worker dispatch/300 distinct keys and
all ordinary completion/invariant/aggregate checks remain. Ordinary CLI run mode also
traverses registered relevance, so do NOT mistake it for a standard-phase-only selector.
No active optimization, schema/baseline/report promotion, timeout increase or CI change.

Actual root measurement command started session98864, still LIVE, not a pass/result:
`make test-go GO_PACKAGES='./harness' GO_TEST_FLAGS='-overlay=/Users/stronk/repos/cloud-clicker/.cache/exact-standard-35cde078-overlay.json -run NoTests -bench BenchmarkStandardPacingProfile -benchtime=1x -timeout=60m -cpuprofile=/Users/stronk/repos/cloud-clicker/.cache/exact-standard-35cde078-candidate-cpu.pprof -o=/Users/stronk/repos/cloud-clicker/.cache/exact-standard-35cde078-candidate.test -v'`.
Startup confirms actual Go1.27.2/Darwinarm64/M3Max/14logicalCPUs,12workers,
scenario10aa5c8a2dd8a5c79a5902e74f4e55a1ca3ab46a2d35c9bf3dd25fc729dbb07a,
constantsbaa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6.
Profile source SHA2561090d8124b0e9cbca42bc62323c51b9107c47ad66fa86c6f083536126a7addd3;
original ordinary benchmark218991f980444a817a94cbc24adc24e5be4dfc61f208accfe79ab8c669f2d83c
is unchanged. Combined overlay maps that benchmark to the committed manual profile plus
aggregate.go to the previously SHA-bound scratch-reuse source. No rate traversal overlay.

Only read-only source/artifact inspection and editorial work proceed during this measurement;
no competing heavy checks/reviewers launched by this chat. Report path is CPU-profile path
plus .report.json, written with O_EXCL. Resume98864, never restart on an empty observation.
On terminal, inspect all300 unique keys, outcomes/invariants/aggregate and full digest before
using timing, then compare actual retained report with historical artifacts under their own
producer limits. The recorded old timings are not a matched control. No current-source full
standard-phase result, matched speedup, relevance completion, baseline/adoption or CI-green
claim yet. This checkpoint merely preserves resumable live work through a crash/context cut.

Read-only retained-baseline inspection changes the follow-up, not the live run: current
golden-seed.json has TWO runs, pacing-baseline.json has300-run aggregate/no failures/warnings.
Their SHA256s4479ada5aa2b782c3f59966a41902aa6ed1789236eb674b51480e18db55ad472 and
e11e077e2536d61b9fb21b9be05d24c79d044de126096fa8cc9885edae4f0122 do not constitute a
full300 before report. The combined two-seed/aggregate JSON hash27171bc2…a01 is NOT the
historical full1ffa47cf…e826d and is not used as that control. Predeclare one full active
source control AFTER candidate98864 terminates, same measurement source1090d812…addd3,
toolchain/current runtime dependencies/12workers/strict300 checks/60m guard, no aggregate
overlay. Retain its complete JSON for actual300-key field comparisons. No repeated candidate,
extra seeds, report regeneration or inference from only the two golden rows. A single
sequential pair is bounded local evidence, not statistical/hosted timing or full acceptance.

## 2026-10-16 — Exact standard candidate completes; full active control starts

Session98864 is terminal0: benchmark770707897625ns/op (770.707897625s timed),
package771.156s, actual300 runs/op. Instrument checks passed: every governed key once,
every outcome completed, no invariant or aggregate failures. Independent retained-JSON
inspection confirms300 distinct keys/all300 completed/zero invariant failures and aggregate
run_count300/no warnings or failures. This is the full standard phase, not registered
relevance, the complete check, caller compatibility, accounting repair or adoption.

Retained ignored artifacts beside the declared overlay, SHA256:
- exact-standard-35cde078-candidate-cpu.pprof.report.json:
  c4254dd5f1213031deb0bb98246782ad96bd331295adc69190d39e9872b39426
- exact-standard-35cde078-candidate-cpu.pprof:
  f47df34d915f899788ec56c67e9d39834f8c1b831b0b50564abc329f2ef046f1
- exact-standard-35cde078-candidate.test:
  124056fa405738d2f1f2a06758c8e2c3ef4ee84bddc83017add48b8161466a18
Candidate source remains4b9e897a9c79a1b38fda998f06a240e0c5596e56e85aaa525c3b255110849304;
manual instrument1090d8124b0e9cbca42bc62323c51b9107c47ad66fa86c6f083536126a7addd3.
No candidate restart or competing heavy workload was launched.

After that terminal result, start the predeclared SINGLE active-source control as
session77856, currently LIVE, not a pass/result. Ignored active overlay maps ONLY the
existing standard_pacing_profile_test.go to the same committed manual instrument; active
aggregate.go SHA256270578489486b6dbd43fce0b53d51dbd3cffb088dd3f18308b6cb5978ae24f52.
Exact root command (source HEAD37a70479; no changed product/runtime bytes since35cde078):
`make test-go GO_PACKAGES='./harness' GO_TEST_FLAGS='-overlay=/Users/stronk/repos/cloud-clicker/.cache/exact-standard-35cde078-active-overlay.json -run NoTests -bench BenchmarkStandardPacingProfile -benchtime=1x -timeout=60m -cpuprofile=/Users/stronk/repos/cloud-clicker/.cache/exact-standard-35cde078-active-cpu.pprof -o=/Users/stronk/repos/cloud-clicker/.cache/exact-standard-35cde078-active.test -v'`.
Same12 workers, declared population/guards, no arithmetic or traversal overlay. Full
report will be CPU profile path plus.report.json. Resume77856; never restart from empty
output. Corrective media-identity reviewer still holds heavy execution until this control
is terminal. Read-only inspection/editorial work only while measurement runs.

Next: validate the control's entire population and retain its artifact identities, then
compare all300 per-run fields and aggregate against the candidate. No equivalence/speedup
claim yet. Existing compatibility and normative-review holds remain; live math/kernel222,
goldens, balance data and CI budgets are unchanged. Review by: Codex root (measurement
author/first filter). Recorded by: Codex root. This is a factual research checkpoint,
not designated approval or a release milestone.

## 2026-10-16 — Full standard control and actual300-row comparison

Session77856 terminates0:948033621291ns/op (948.033621291s timed), package948.494s,
300 runs/op. Both retained reports have300 distinct matching keys/all300 completed,
zero invariant failures, aggregate run_count300/no failures or warnings. The active full
report exactly reproduces historical digest1ffa47cfd4f85f6daf1b340cf6e43cfa7d534cb39308cdf80e955c9d28be826d.
No guard, exclusion, dropped row, candidate restart or competing heavy workload.

Full JSON comparison sorts by complete RunKey and compares every field of every row,
not merely the two golden rows or aggregate. Aggregate is identical, but36/300 rows
(all chaos.phase0) change company.cash final balance and final_state_hash. All other
resource balances and all other per-run fields are identical;264 full rows are identical.
Example chaos seed10: active cash1.42967585038e13 versus candidate1.4296758504e13.
This is observed semantic drift in the complete standard workload, not a behavior-preserving
optimization, an adopted report producer or proof that a kernel bump disposes of compatibility.

One ordered native pair: candidate770.707897625s versus active948.033621291s timed,
18.704581745% lower elapsed time for this candidate run. Not statistical speedup, a CI/hosted
gain or all-workload feasibility. The short multi-source microbenchmark's higher reducer
cost did not predict this whole standard phase; neither measurement can replace the other.
No additional repetition, shortened horizon, relaxed acceptance bound or baseline regeneration.
Registered relevance/search/content/Commons workloads and the draft's complete static
caller/exception/producer/cutover gates remain unproved. The global aggregate overlay does
not implement the proposed checked API or pinned static catalog/Pitch exceptions; this is
not a measurement of that finished migration. Kernel222/active math is unchanged.

Retained active artifacts in.cache (same filename prefix exact-standard-35cde078-active):
- -cpu.pprof.report.json SHA2561ffa47cfd4f85f6daf1b340cf6e43cfa7d534cb39308cdf80e955c9d28be826d
- -cpu.pprof SHA2560697ec5504eab3a561ec9b5682927d5c09cff7d80fa3e569ba03851aa82c3ad1
- .test SHA256c451638fc1f39f6ff9f824c1d4ea9fba2f06dad11f99ccb381a217a04568b761
- -overlay.json SHA2567ae23330a45426657c0918f75e8ff7edfc85311330f4773f76ee50c5a91c894b
The previous entry binds candidate artifacts and the shared instrument. Read-only Git
comparison35cde078..ad40b0f0 confirms no server/client/kernel/balance/testdata source changes.

Both measurements now terminal: release the serialized execution slot ONLY to independent
RP-535 reviewer api_media_identity_corrective_review. RP-524 and RP-502 reviewers have inspected
their complete corrective scopes read-only and remain held for their subsequent slots.
No reviewer verdict inferred from source inspection. Own draft A7's factual cost note now
distinguishes this standard phase from still-missing complete affected workloads. No normative
adoption, feature approval, archival, publication or1.0 milestone. Review by: Codex root
(research author/first filter). Recorded by: Codex root. The complete standard-measurement
research range35cde078-exclusive through this containing commit remains research review,
outside the corrective-only delegation; existing normative/research holds are unchanged.

## 2026-10-16 — Targeted harness search version-boundary probe, before execution

Question: can actual RelevanceSuite.earliestAffordable clone an independently admitted v17
Company, or does cloneState's hardcoded v14 restore block it before arithmetic is evaluated?
Use the unchanged real fixture suite loader/catalog and two admitted origins: legacy v14 with
one alpha generator, and v17 with the same generator plus1000ms of burst. Exact fixed cost101
from initial100 should first be affordable at500/250ms respectively under the fixture's rates.
Require version/encoded-byte preservation, source independence and an unchanged source after
actual search. A refusal at clone/restore is its own result, not proof of a numerical search
miss. Manual source is probes/harness-clone-version-probe.go.txt, outside ordinary CI.
No arithmetic, active accounting, catalog, balance, baseline or kernel change is authorized.
If refused, first distinguish unsupported integration from a bounded clone-conformance defect;
do not silently alter a version to reach the desired counterexample. This is a two-origin
boundary diagnostic, not a new census, full modern factory, natural play or whole-report proof.

## 2026-10-16 — RP-678 actual harness clone/search version refusal repaired

The targeted diagnostic finds a separate consumer defect before numerical search. Initial63193
passes legacy control but refuses the v17 FIXTURE before admission because its meter maps are
uninitialized; excluded. Initialize its required empty maps, not the clone/predicate. Corrected
old-source69377 independently encodes/restores v17 successfully, then BOTH cloneState and actual
earliestAffordable fail on v14's unknown achievement_score_lifetime field. Admitted v14 control
returns the expected500ms and leaves source untouched. No rounding decrease/search miss is
inferred from a pre-search decode error; RP-678 is distinct from RP-676 and RP-654's other clone.

Bounded conformance correction under Balance Harness FoundationD1 and the existing Save version
contract: cloneState restores save.VersionForState(source), exactly the version its encoder
uses, rather than v14. The diagnostic is promoted from the temporary .go.txt location into
ordinary regression server/harness/clone_state_test.go (not duplicated in a manual/CI corpus).
It admits each source before cloning, checks version/complete bytes/non-aliasing and actual
earliestAffordable: legacy14 reaches cost101 from100 at500ms; v17's preserved1000ms burst
reaches it at250ms. Neither consumes origin. Focused83645 PASS0.259s; full
`make test-go vet GO_PACKAGES='./harness' GO_TEST_FLAGS='-count=1'`95174 PASS82.645s plus vet.
No timeouts, exclusions or existing tests changed. These are ordinary package populations,
not a new full300 standard measurement, full relevance run or axis-enabled factory proof.

Canonical harness docs/own plan/ledger reconcile; kernel affecting-paths excludes server/harness,
so no false numeric version bump. No price, accrual, simulation formula, state format, balance,
scenario constructor/guard, hook, golden/baseline or active accounting algorithm changes.
Review by: Codex root (implementer first filter). Recorded by: Codex root. Complete range
d9b5b07b-exclusive through this containing repair commit requires fresh independent corrective
review. All root handles terminal before reviewer execution. No DB/browser/hosted/full-CI,
Clout AC6/27-failure repair, original feature acceptance, archival, publication or release claim.
Next: after independent correction review, follow the actual affordability method with the
admitted RP-676 profile and a bound real quote, without labelling it a complete modern factory
or natural report failure. Continue the correct search/partition contract, not a false generic
monotonicity proof or another passing census.

## 2026-10-16 — RP-678 clone correction independently approved; canonical version wording

Review by: /root/harness_clone_version_corrective_review (fresh independent Codex under
Marco's corrective-code delegation). Recorded by: Codex root, transcribing terminal verdict.
Decision: APPROVED, no findings; complete six-path one-commit range
d9b5b07b5beffb22a5fa8992c0be08048d5d77f1-exclusive..de0008ec2443e087dbc58054b9dbac3b9a63c17a-inclusive.
Actual restore matches the encoder's existing VersionForState(source), not a new schema or
activation. Reviewer also checks accepted Doctrine/Compute D9's independent Company v17 codec.

Independent exact old harness.go ignored overlay74050 fails BOTH admitted-v17 clone and
actual search at unknown achievement_score_lifetime; v14 control passes500ms. Corrected
`make test-go vet GO_PACKAGES='./harness' GO_TEST_FLAGS='-count=1'`49151 PASS81.853s plus vet.
Separate verbose27341 PASS3.483s: v14/v17 version/complete bytes/nonalias/source preservation,
actual search500/250ms, existing relevance fixture, first-hour reference consumer and unchanged
axis refusal/legacy control, no selected skips. Gofmt/full-range diff clean. Go1.27.2 DarwinARM64;
ordinary package checks, not opt-in exhaustive measurements or hosted execution.
Reproduce old regression with root Go overlay selector and TestCloneStatePreservesSaveVersionThroughSearch:
.cache/review-rp678-old-harness.go SHA139c3ee8410744323e05dd6adff961b12b2f77a8f18a07db92f9fcbfa4a8c78a;
.cache/review-rp678-old-overlay.json SHA0b190e548b954a7bc12681aa67951dfd279d66226ed2cf7289e3102af6b09adc.
All handles terminal/slot released; no reviewer tracked edits. No standard/relevance full run,
modern axis integration, RP-676/accounting/Clout/CI/new-feature/archival/release acceptance.

Separately, reviewer reports canonical docs/save-layer.md's unqualified "Company rejects
v17/v18" versus its own independent Company v17 paragraph. Root records RP-679 and clarifies
the Founder-only payload context and bundle-conditional mixed tuple. These are implemented
codec distinctions, not a schema/activation/ruling amendment. Documentation-only follow-up
lies outside the six-path verdict and needs no software rerun; diff/source checked. No accepted
RFC or archived text changed. Continue actual-search caller analysis after this bounded repair.

## 2026-10-16 — Before execution: actual quoted affordability-window consumer

Use one RP-676 neutral economic recipe, not another monotonicity census. Retune only its
controlled catalog and one existing upgrade's cost/resource requirement to the earlier cash;
load actual scenario/routes/policy/catalog/hash via LoadRelevanceSuite and independently admit
the v17 origin. Compare actual candidateCost, SimulateAdvance/receipt at499/500ms of the last
bucket and actual earliestAffordable with those two horizons from the SAME untouched origin.
Hypothesis: shorter horizon yields a valid affordable witness, extending by1ms falsely returns
unreachable despite containing that witness. Use the scenario's unchanged transition budget;
errors/refused admission/guard exhaustion are not search misses. Preserve receipt/provision,
full source-byte and within-scenario-horizon controls. Manual source
probes/relevance-provision-window.go.txt stays outside ordinary CI. This tests current private
search with the real suite loader and a real unminted catalog quote; not full RunRelevance,
modern axis/hook factory, natural acquisition/report drift, SQL or a repaired algorithm.

First declared method probe57020 PASS9.321s: real admitted suite/quote, shorter horizon65536499
returns65536333 with27 charged transitions; horizon65536500 returns unreachable with1, nil error.
Direct cash/receipt499→500 match RP-676. Before final execution, extend this SAME profile to
actual rankCandidate for both horizons: require short arm to apply the real quoted upgrade,
not merely compare a hypothetical threshold, and check whether long arm loses that candidate.
No broader population, changed budget, source mutation, asserted global earliest proof or
RunRelevance/published-report claim is added. This bounded caller extension may fail on a
different admission/eligibility boundary; such failure would not count as a ranking miss.

## 2026-10-16 — RP-680 quoted search and actual purchase-ranking counterexample

Final47483 PASS17.863s executes both declared method and caller populations, including
explicit quoted-balance spending control: cash equals the actual quote before
purchase and0 afterward (no substitute ownership-only assertion). Current
LoadRelevanceSuite admits exact retuned economy/routes/policy/scenario and computes constants
sha256:7cdc13340749a78b075e1aede620faf9af0290c642e45c0e19c9b9496800e5d1.
Independently admitted v17 late origin, real candidateCost for upgrade.beige_tower_cache at
1.968058018e6, actual SimulateAdvance/receipt/provision-cap controls and unchanged encoded
source all execute. Horizon65536499 finds65536333 using27 actual transitions. Extending by1ms
to65536500 returns unreachable with nil error after1 transition, despite the valid witness
remaining inside the range. Direct endpoint cash matches1.968058018e6→1.96805801799e6.
Actual rankCandidate short arm spends the real quote/applies the upgrade at65536333,
revision2,29 charged transitions; extended arm suppresses the same candidate with nil error.
No diagnostic guard or scenario budget exhausts. RP-680 records this current consumer defect;
RP-676 remains the separate payout-decrease evidence, not relabelled as natural player loss.

Reproduce from repository root:
`make test-go GO_PACKAGES='./harness' GO_TEST_FLAGS='-count=1 -overlay=/Users/stronk/repos/cloud-clicker/.cache/relevance-provision-window.json -run TestRelevanceProvisionWindowResearch -v'`.
Overlay maps server/harness/manual_relevance_provision_window_test.go to committed manual
source probes/relevance-provision-window.go.txt, SHA256
81faecc91284845f58eb2f8b4335dec2d62f43a41a6704915507c0fcf34322c0;
overlay SHA1b3b31d842df8534f02f27b2eb8f3d54967fc0b23d444a31ee384d063e19bcfd.
All57020/74339/47483 terminal; first run tests method only, second adds the ranking consumer,
final strengthens it with observed post-purchase cash0. No earlier failure is relabelled green.
Runtime temp inputs use t.TempDir and are removed by Go; tracked source/content unchanged.
No routine CI population, arithmetic, search, budget, baseline, scenario or kernel is altered.

This is deliberately unminted retuning/seeded late origin with nil-hook standalone factory,
not natural acquisition, full RunRelevance, modern axis/hook integration, paired TS/SQL,
published-report drift or lost saved payment. The short result is a verified valid witness,
not an exhaustive global-earliest oracle. It suffices to falsify the longer unreachable result.
Owning plan, ledger and own draft's H3c reconcile these precise findings; no new algorithm or
accepted contract follows. Review by: Codex root (research author/first filter). Recorded by:
Codex root. Complete range8f1e399e-exclusive through this research checkpoint remains subject
to research/normative review outside corrective-code delegation. Next: correct search/partition
proof and complete caller/work/compatibility closure; neither a window-specific special case
nor a generic approval repairs this defect or the27 original accounting failures.

## 2026-10-16 — Before execution: positive aggregation enclosure component

Question: can a conservative rounding enclosure contain current positive aggregation,
including RP-676's decreasing pair, without asserting that current payouts are monotone?
Derive an exact-rational upper bound for represented inputs, not their rounded String
forms. Proposed slack is (N+16G)*2^-52 + G*5e-15, with an additional
2*2^-52+5e-12 for Q12. N/G are uniform source-term/positive-exponent-group ceilings,
not endpoint-specific counts reused as an interval proof. Require 0<=G<=N<=86400002,
actual counts within those ceilings and slack<1. The ceiling is this experiment's
bounded domain, NOT a new production limit or a claim about every admitted producer.

Manual Go source will execute actual SumDeterministic and Q12 against independent exact
binary-mantissa rational sums on: the two retained65537-term RP-676 lists; zero/empty,
singleton, group carry, exponent gaps17/18 and Q12 tie neighborhoods. Verify required
Pow10 constants against exact powers, and worst-domain slack without allocating86M terms.
Negative/special/non-state inputs, unsupported executable-reference exponents outside
[-32,32], undersized ceilings and over-domain ceilings must refuse loudly. No random
positive census, tolerance, altered comparison, production math/search/API or ordinary CI
population. A failure invalidates the proposed component and is retained, not excluded.

This is a bounded executable challenge to a source-derived candidate inequality, not a
universal proof by sampling. The derivation must expose normalization/power and finite-normal
arithmetic assumptions. Even a valid aggregation enclosure does not supply a uniform bound
on all time-varying producer outputs, the search algorithm, real caller completeness or
workload feasibility. RP-680/accounting remain OPEN; research/normative review and adoption
are outside the corrective-code delegation. No release or feature status promotion.

Before first execution, make normalization's library side condition finite and explicit:
in addition to40 local exact-rational power comparisons, check every632 finite Pow10
argument's outer magnitude and the two saturation edges. With positive normalization
input in [1,4e15], a valid nonzero normalized result cannot use a shift outside[-2,16]
under those checks; an underflowed zero cannot violate a positive upper bound. No
monotonicity or accuracy assumption about Log10's selected integer is needed. The
manual proof is for64-bit Go; no32-bit/client/hosted guarantee is inferred.

Native first execution03707c PASS .437s package/.21s manual population: both retained
decrease endpoints fit; exact-input envelope with the same ceilings increases. Both
omitted-margin controls discriminate,12 refusals and all finite power premises pass.
Before environment follow-up: execute the SAME source/population in the declared Linux
AMD64 Go service with --no-deps (no database boundary in this component), using its
/workspace overlay mapping. Also finish the native decimal package with the overlay
and ordinary vet. Linux emulation is not native hosted hardware or TS parity.

## 2026-10-16 — Positive aggregation enclosure: bounded component result

Native full decimal package with the manual overlay59651c PASS .601s plus ordinary
decimal vet. Declared Linux AMD64 service43509/e3b009 PASS .535s package/.50s manual
population, no selected skips; same exact gaps as native. It ran --no-deps, created no
Postgres service and its --rm test container is terminal. Docker Desktop emulation on
ARM64, not native hosted Actions. No runtime/source edits occurred during either run.

Both retained Q12 values stay1.968058018e6/1.96805801799e6. With the SAME N65537/G2,
upper-minus-current is .000035134178807512/.000045134189634029; the rational upper
envelope itself increases. Ten named edge cases execute raw and Q12 comparisons;
alignment-round-up exceeds the zero-slack bound and Q12-round-up exceeds the raw-only
bound. All12 explicit refusal cases and40 exact-power/632 range/2 saturation premises
pass. Worst declared Q12 slack .000000758144133707 is derived, not a raised threshold.
The source-derived candidate argument is in own draft H3d; its positive-operation
path/product bound makes no generic actual-payout monotonicity claim.

Reproduce native from root:
`make test-go GO_PACKAGES='./decimal' GO_TEST_FLAGS='-count=1 -overlay=/Users/stronk/repos/cloud-clicker/.cache/positive-aggregation-enclosure.json -run TestPositiveAggregationEnclosureResearch -v'`.
Overlay maps server/decimal/manual_positive_aggregation_enclosure_test.go to committed
probes/positive-aggregation-enclosure.go.txt SHA256
2ba2ca24645bfb470b1ab0136cbdf77482dde975fd29ea9bb450a0534ed71315.
Ignored native overlay SHA2bb84a39039afdcca51392e2c2fc108adcf0fcd9b6089fd4b3f23c11e67b9d2f;
Linux overlay SHA472a75647d4d12b323bfb8b7599d9b7755dc7561556dd928efd0cbd42946a8cf replaces
the same paths under /workspace. Linux reproduction:
`docker compose -f compose.save-test.yml -f compose.ci-test.yml run --rm --no-deps test go test -p 1 -overlay=/workspace/.cache/positive-aggregation-enclosure-linux.json -run TestPositiveAggregationEnclosureResearch -v ./decimal -count=1`.

Only manual source, own draft and owning plan/log change. This bounded candidate is
neither the temporal producer interval certificate nor complete search/caller/work
proof. Exact rational arithmetic here is independent reference/certificate research,
NOT adoption of Exact Represented Aggregation as payouts. No production/math/kernel,
canonical behavior docs, ordinary CI, content, reports or acceptance labels changed.
RP-680 and27 AC6 failures stay open. Next derive uniform producer/ledger envelopes
before any new actual search implementation. Review by: Codex root (research author/
first filter). Recorded by: Codex root. Complete four-path range4cf683d1-exclusive
through this containing checkpoint requires research/normative review, outside the
corrective-code delegation; no designated approval, feature closeout or release claim.

## 2026-10-16 — Before execution: fixed-prefix production-cell enclosure/search

Previous goal turn PROGRESS: f7a2ad7b derived/executed the positive aggregation
component. Next test its composition with the real varying partial producer and
ledger, NOT another positive monotonicity census. Manual source will reuse RP-676's
four actual Evaluate/SimulateAdvance anchor controls. Within the retained1000ms
bucket use its exact fixed65536-term prefix, actual accrueBoostedConstant partial,
actual SumDeterministic/Q12 and actual Ledger.ApplyAccrual. Check every endpoint
0..1000 against a rational whole-prefix/partial/ledger upper bound. Additional
complete1001-endpoint controls: ordinary90% efficiency with nonzero starting balance,
zero rate and zero efficiency. No synthetic rate or changed input is called a
naturally acquired state or full Evaluate population.

For this bounded grid, prepare a prefix-maximum table of the ACTUAL integer-ms
seconds conversion once. Bound positive Mul/Add/Q12 with the source-derived rounding
factors, retain the real constant raw total rate and uniform N/G for prefix+partial.
Table preparation is explicit1001 work, not hidden per-query or full-domain proof.
Compare chronological certificate-pruned search with the exhaustive earliest
endpoint inside the cell for the real retained threshold; extend its horizon to500
where the endpoint is unaffordable. Controls: already affordable and unreachable,
the old endpoint-only pseudo-certificate, and a deliberately omitted burst factor.
Errors/refusals/work exhaustion invalidate a result rather than mean unreachable.

All4004 pipeline endpoints and all search controls must execute. Count actual point
and certificate work separately from exhaustive reference/table/prefix construction.
Do not charge these diagnostic counters as real harness transitions. No production
algorithm, payout, kernel, budget, content, report, CI or accepted contract change.
One fixed cell is not whole-horizon completeness, hook/axis/caller integration or
workload feasibility. Own draft engineering/research remains outside corrective-code
review delegation; RP-680/AC6 remain open regardless of the bounded result.

First launch86601/5d51b1 selected NO tests: Make consumed the unescaped trailing
regex dollar and joined -v into the selector. Excluded as unexecuted, not a pass.
Use the unique test-name selector without the dollar; no source/assertion change.

First actual execution8664/bdd181 PASS2.747s, all4004 pipeline endpoints and four
real anchor projections. Exhaustive earliest cell point333; chronological search
at horizons499/500/1000 finds333 with14/15/16 nodes and one fresh actual pipeline
point each,4/5/5 prunes. Preparation/reference work remains separately counted.
Before final execution, add explicit propagation of a REAL negative-rate producer
refusal, unavailable/nil/negative certificate refusals, unsupported grid and work
exhaustion; verify the calculated upper function itself never decreases across
the complete declared cell. These make validity assumptions visible; they do NOT
prove that an unobserved general range cannot contain a producer/ledger refusal.

Final native44767/b0f40f PASS2.806s, ordinary production vet also passes; all original
4004 endpoints/four actual anchors/search controls retained, added definedness/refusal
and calculated-upper monotonicity checks pass. Before environment follow-up: run the
identical final source in declared Linux AMD64 Go --no-deps with /workspace mappings.
No DB boundary, broader production suite, fullCI, native hosted or TS claim intended.

## 2026-10-16 — Fixed-cell producer/ledger composition result

Final Linux AMD6459606/e1982f PASS5.994s package/5.93s manual population, same
4004 endpoints, four actual anchors and search counts as final native44767/b0f40f.
Both final sources include actual negative-rate refusal propagation, three invalid
certificate controls, unsupported-grid and work-exhaustion refusals; all execute.
Both omitted-burst and endpoint-only pseudo-certificates exclude the known reachable
threshold incorrectly. Correct chronological search matches exhaustive earliest333
for horizons499/500/1000 with14/15/16 certificate nodes and one fresh pipeline point;
already affordable0 and unreachable controls also pass. Scalar1001/prefix65536 and
full4004 reference work are explicit setup, not hidden in the search query counts.

Committed manual source probes/production-cell-enclosure.go.txt SHA256
59931e558fe89b8a3d45f168f9a2558ccc23f545081fd20065868524cc46f0ef;
unchanged reused provision-carry-probe.go.txt SHA
d4148e2df242488a64007946ed716366b6eec2eae4c3f26ddd0bacc633b9e44f.
Native ignored overlay .cache/production-cell-enclosure.json SHA
ff397f6718f3c7e2ceb76edc22b893aa94f147a6caf920212177a2712e179147
maps both server/production/manual_*_test.go names to the two manual sources.
`make test-go vet GO_PACKAGES='./production' GO_TEST_FLAGS='-count=1 -overlay=/Users/stronk/repos/cloud-clicker/.cache/production-cell-enclosure.json -run TestProductionCellEnclosureResearch -v'`.
Linux overlay .cache/production-cell-enclosure-linux.json SHA
2cd42febfb5f17aa1b6d805549e4c91891fd1397924a23c0eafac5a8a075663f
maps the same files under /workspace; use the declared --no-deps Docker command from
the previous checkpoint, substituting this overlay and TestProductionCellEnclosureResearch.
Both final handles terminal; --rm Linux run creates no DB fixture. No source edits
during runs, and no native hosted, DB/TS, fullCI or broader-suite proof is inferred.

Own draft H3e supplies the composition and definedness caveat; owning plan points
next to real cell/context construction and setup/work accounting. No actual harness
method, production arithmetic, kernel, payout, balance/content, report, ordinary CI
or accepted contract changed. RP-680 and27 AC6 failures stay OPEN. The manual search
is not whole-horizon/caller completeness or performance acceptance. Review by: Codex
root (research author/first filter). Recorded by: Codex root. Complete four-path
rangef7a2ad7b-exclusive through this containing checkpoint needs research/normative
review outside the corrective-code delegation. Full1.0 goal remains active/progress;
no feature completion, archival, push, deployment or release approval.

## 2026-10-16 — Before execution: source-driven wait cell constructor

Previous turn PROGRESS aabe9ce0: bounded partial/ledger search, not a whole wait.
Build a manual constructor from actual simulationPolicyFor/assembleContributions,
ratesWithProvisionedAndPolicy/materializeProvisionBoundaryWithPolicy and actual
accrueBoostedConstant. No hand-authored completed prefix: retain actual canonical
segment deltas on an append-only shared slice and its running rational sum/count/
exponent set. Cash alone is the certificate subject; real leaf projections still
execute SimulateAdvance from a separately restored untouched origin. Boundary
ownership must cover every integer endpoint exactly once; offline cap adds a cash
plateau, not another payment or reset. Validate source bytes unchanged.

Populations: default60000ms grid/cursor+17/burst70000, horizon120005, twelve declared
edge points; retuned1000ms grid/cursor+17/burst1500, horizon3001, EVERY0..3001;
same grid offline with cap1500/burst1700 and same whole population; same online
grid with provider masked, same whole population; existing no-tick Phase0 origin,
horizon3001, same whole population. Counts25/provider1/remainder900000 in the
provisioned ordinary populations are seeded, codec-admitted origins, not acquisition.
Default edge offsets:0,1,59982,59983,59984,69999,70000,70001,119982,119983,119984,120005.
Retained RP-680 neutral retuning, full65536499ms elapsed horizon, all cells built;
actual reference offsets0,1,998,999,1000,65536332,65536498,65536499. Its quoted
threshold must retain the valid point despite the unaffordable final endpoint.

Compare exact reconstructed cash and its cash receipt change with actual full wait
at each reference point, plus the conservative bound. In each exhaustive small
population, compare chronological cell search with its independently scanned earliest
threshold result. Long retained population has selected references/conditional
certificates, NOT an exhaustive65M-endpoint oracle or complete refusal-domain proof.
Record all rate/provision/delta/prefix/scalar/certificate and real point work, plus
setup/search time separately; counters are not existing harness transitions/budgets.
Negative controls must detect omitted provision materialization, wrong origin grid
and omitted offline cap; unsupported context/grid/cell size and errors refuse loudly.

Manual experiment guard: at most100000 cells, scalar table at most60001 entries;
unsupported shapes fail, never truncate/fall back or imply unreachable. This is NOT
a proposed runtime/content/work limit or a substitute for full admitted-domain/work
closure. No live API/math/kernel/harness/CI/payout changes, research/adoption gates
unchanged. Actual hook/modern endpoint context, general definedness and complete
workloads remain outside this experiment, not claimed fixed by its constructor.

First execution71339/aa8d83 FAIL4.093s: all12028 honest actual-wait references and
cash/receipt/bound checks execute, but the omitted-cap negative control cannot build
its3001ms uncapped horizon: the EXISTING provision-horizon guard correctly refuses
it. Not a numerical discriminator, not a green experiment. Retained long constructor
does find65536332 elapsed, with65537 cells,65551 search certificates/one actual point;
timings remain preliminary, not comparative performance proof.

Instrument correction before rerun: keep ALL honest populations unchanged. Run the
omitted-cap negative control on0..2000, a declared subrange of the already observed
offline population. Here an uncapped2000ms projection satisfies the original guard
(2000/1000 <=1500/1000+1), so a mismatch after1500 can genuinely detect missing cap
application without bypassing another source guard. Its points above2000 are outside
that negative control's declared domain, not exclusions from the honest population.
Keep the original horizon guard and require a numeric mismatch, never an error.

Corrected20692/a2835a PASS4.094s plus ordinary production vet. Grid/provision
faults differ at59984(default)/984(short); omitted cap now differs numerically at1501.
Before final runs, make the diagnostic search's work refusal explicit too: its
certificate-node ceiling is the finite binary-tree bound2*(horizon+1), NOT an adopted
harness budget. Inject a one-node allowance and require an exhaustion error. This
retains every declared endpoint and all original source guards; no runtime limit is
raised and no incomplete search may return an unreachable result.

13594/d9b764 PASS4.162s plus vet, including explicit one-node exhaustion refusal.
Before environment run, correct a diagnostic label: short_grid's printed search
certificate count included that later one-node negative control (15 rather than14).
Snapshot the main query counter immediately after search; total work still records
the control. No outcome/assertion/population or numerical code changes. Rerun final
native source and the identical declared Linux AMD64 --no-deps population; emulation
does not establish native hosted/TS/DB/fullCI behavior.

15292/63c705 native PASS4.425s plus vet; Linux47726/9706e2 PASS8.150s,
same12028 actual references and control outcomes. Final first-filter gap: whole-cell
search success is tested, but unlike the earlier single-cell tool it has no direct
already-affordable/unreachable controls. Add BOTH for each of the six profiles
(quote0 =>0; quote1e12 =>unreachable), using fresh actual point callbacks and the
same derived diagnostic ceiling. Preserve all12028 reference points and snapshot
main-query node/point counts separately from these controls before final reruns.
This is extra negative coverage, not a changed result, work ceiling or authority.

## 2026-10-16 — Source-driven wait constructor result

Final84886/c906e3 native PASS5.895s plus ordinary production vet; declared Linux
AMD6442109/bd0bc4 PASS10.812s, no selected skips, same source/populations/outcomes.
All12028 real SimulateAdvance cash AND cash-change receipt references match the
source-driven reconstruction and fit their cell bounds. Four small complete
populations match their independently scanned earliest endpoints. All six profiles
also pass quote0 already-affordable and quote1e12 unreachable controls.

| Population | Actual references | Cells | Scalar setup | Main query certificates | Main actual points |
|---|---:|---:|---:|---:|---:|
| default grid | 12 | 3 | 60001 | 11 | 1 |
| short grid | 3002 | 4 | 1001 | 14 | 1 |
| offline plateau | 3002 | 3 | 984 | 20 | 1 |
| masked provider | 3002 | 4 | 1001 | 14 | 1 |
| no tick | 3002 | 1 | 3002 | 22 | 1 |
| retained long wait | 8 | 65537 | 1001 | 65551 | 1 |

Retained constructor calls actual rate assembly65537/provision materialization65536/
segment accrual131074, retaining65537 cash terms. Main query returns65536332 elapsed,
absolute65536333 from cursor1ms, despite unaffordable final65536499 elapsed. Native
setup329.477ms/search1.898746s; emulated Linux697.912ms/3.659344s. Single ordered local
observations, not an old/new comparison, full scenario result or performance gate.
Totals include references and later controls:131107 certificates/two actual search
points, plus eight real reference calls/196616 reconstruction terms. No setup work
or second control query is disguised as the main one-point query's cost.

Grid/provision omission faults differ at59984(default)/984(short); missing cap at1501
on its guard-admissible0..2000 negative subrange. Honest offline remains complete0..3001.
Hook/query/scalar/cell guards and injected one-node exhaustion refuse explicitly.
The first bad cap instrument/main-query counter correction and all earlier runs remain
above, not relabelled final green. Origins stay byte-unchanged. Final handles terminal;
--no-deps/--rm Linux run starts no database. Emulation is not native hosted Actions.

Reproduce from root using .cache/production-wait-cells.json, which maps the three
server/production/manual_*_test.go files to the committed wait-cell, prior cell and
provision manual sources:
`make test-go vet GO_PACKAGES='./production' GO_TEST_FLAGS='-count=1 -overlay=/Users/stronk/repos/cloud-clicker/.cache/production-wait-cells.json -run TestProductionWaitCellConstructorResearch -v'`.
New source probes/production-wait-cells.go.txt SHA256
16b1be49c4e5d916083a66a1dc23d6728c13d1e5f9632d5ba09eb1023a9d2dd1;
native overlay SHAb19f1bfeabdadbbd505f8f0d4f8ed5abf95ada834424849b7b832471c3996c46.
Linux overlay .cache/production-wait-cells-linux.json SHA
5a9785133538f2105fe51ad0f9929fea502244dbe5e57d66061f0e11606ddd82
maps the same files under /workspace. Reproduce with declared Docker --no-deps
command above, selecting that overlay and TestProductionWaitCellConstructorResearch.
Reused manual sources are byte-unchanged; no product/schema/math/kernel/harness/
content/ordinary CI/report/golden bytes or accepted contracts changed.

Own draft H3f explains cell ownership/plateau/source binding and limits; the living
plan consolidates H3d–H3f into one current next action, preserving the append-only
history here. Next: actual all-resource/ledger refusal-domain proof and modern
context construction, then complete workloads with setup charged. The long population
is selected references plus conditional bound evidence, not an exhaustive65M-endpoint
or general definedness proof. No full-state/other-resource receipt, SQL/TS/fullCI,
actual private harness replacement, normative adoption or RP-680/AC6 closure claimed.
Review by: Codex root (research author/first filter). Recorded by: Codex root. Complete
four-path rangeaabe9ce0-exclusive through this containing checkpoint needs research/
normative review outside corrective-code delegation. Full1.0 goal active/progress;
no feature acceptance, archival, push, deployment or release approval.

## 2026-10-16 — Before execution: all-resource partial arithmetic definedness

Previous turn PROGRESS5aa403c6 built source-driven waits; its partial-domain caveat
is unresolved. Attack: admitted v17/tick1000/zero-burst origin with provider1 producing
permits at1e-8999999999999999, no cash production. A full1000ms segment may succeed
while1..999ms fail lower-domain accrual. Test actual SimulateAdvance at every0..1000,
expected typed engine refusal/unchanged failed states versus successful endpoints.
Show the previous raw constructor/cash-only search can return unreachable for cash1
without evaluating any point; this is a RESEARCH prototype defect, not a newly
executed current private harness miss or permission to change production arithmetic.

Add a manual partial-arithmetic certificate before prototype search construction:
retain all resource rate vectors and execute the actual accrueBoostedConstant on
every integer partial duration in each distinct numeric context. Reuse ONLY after
all rates are validated and total-rate raw mantissa bits/exponent, efficiency,
bonus factor and burst schedule agree. Full/no/partial burst are separate contexts;
cache coverage may extend, never silently skip. Actual raw total rate suffices only
because the pure helper reads validated rates solely through that exact sum. Rounded
String is not an identity. Count unique contexts, actual checks and reuse explicitly.

The negative context must fail with its exact resource/duration and actual accrual
refusal, not a diagnostic guard. Controls: provider rate1e-8999999999999996, provider
mask and zero offline efficiency; every0..1000 actual wait executes for each. A raw
coefficient-collision control (same Q12 String, distinct IEEE value), exponent and
equivalent validated rate-vector controls test cache identity/coverage. Retain and
rerun H3f's full12028 actual references and all prior faults, nodes/guard controls.
Native and declared Linux AMD64 --no-deps, final source identical in both.

Finite partial-domain checks are constructor research, not a per-player tick or a
production per-ms wait scan. Their entire preparation cost must be visible and full
workload feasibility remains unproved. This certifies PARTIAL arithmetic only, NOT
final multi-segment aggregation, resource caps/receipt reproduction, progress/hooks
or modern states. No production/kernel/math/CI/content/RFC adoption or defect closure.

Old prototype baseline37367/ce41ac FAIL .733s exactly at the new guard oracle:
secondary permits rate1e-8999999999999999 is catalog/codec-admitted; actual waits0
and1000 succeed, every1..999 refuses with typed engine/accrual boundary and unchanged
complete failed state. Cash-only constructor accepts; its search returns-1/nil after
zero actual points for cash1. Other three complete1001-point companions pass,
including zero efficiency: do NOT invent a raw-exponent rejection that overrides
the actual helper's valid zero behavior. Record RP-681 as a prototype research defect,
not a newly demonstrated actual harness defect; add the all-resource partial check.

Protected focused89534/94f3a5 PASS .532s. Broader2141/6ab2c8 PASS6.145s plus vet:
all12028 prior references/faults and4004 new actual waits execute. Retained long
prototype checks4004 partial durations across5 raw contexts, reusing131068 complete
resource-cell ranges after131074 actual rate sums. This is numeric partial-domain
evidence only; no final aggregate/ledger or complete-search error policy follows.

Before final execution, strengthen cache controls beyond counter identity: safe
rate1e(min+3) followed by a valid raw coefficient just below10e(min+2) has the SAME
Q12 String but must fail actual partial arithmetic at1ms. A same-mantissa/different-
exponent pair must also fail, as must a2ms burst where1ms burst succeeds near the
upper exponent. These are pure-helper contexts, not asserted admitted saves.
The correct key must expose the real refusal, not reuse another context's pass.

Final native51516/2eb320 PASS6.187s plus ordinary production vet: all12028 prior
real references,4004 new actual waits, five cache-control populations and all prior
faults/guards execute. Semantic same-String/different-exponent/burst-duration aliases
produce the exact expected helper refusals instead of reusing passing coverage.
Before environment follow-up, freeze both manual sources and run the identical
declared Linux AMD64 --no-deps population via /workspace's four-file overlay.

## 2026-10-16 — RP-681 partial arithmetic domain guarded in manual prototype

Final native51516/2eb320 PASS6.187s plus ordinary production vet; declared Linux
AMD6410837/987ac4 PASS11.990s, same source/outputs, no selected skips. All12028 prior
real references and4004 new waits execute: the bad origin has999 expected typed
refusals/two successes, three companions3003 successes. Every failed Company save
encoding is byte-unchanged; no assertion about unencoded transient fields. The raw
old constructor still reproduces unreachable/nil with no point call. Protected
construction reports actual company.permits duration1/ErrInvalidAccrual instead.

Five cache-control populations pass, including semantic counterexamples: same Q12
String or omitted raw exponent can hide an actual1ms refusal; merging1ms/2ms burst
schedules hides upper-domain overflow at2ms. Equivalent validated vectors, checked
coverage extension/subranges and full/no/partial burst controls also execute.
No canonical String identity, cross-run protocol or raw-exponent rejection policy.

Partial preparation checks/contexts for the six original factories:
default120008/3, short2986/3, offline1502/2, masked2986/3, no-tick3002/1,
retained4004/5. Retained reuse131068 ranges after131074 actual raw rate sums; all
extra checks are explicit preparation work, not hidden in one-point query counts.
Native retained setup316.982ms/search1.849397s; emulated Linux713.262ms/3.459715s.
These single observations are not a matched performance/full workload conclusion.
The source constructor now retains all produced rate vectors; its former upper-only
body is explicitly named buildRawWaitCellsResearch for the executed negative.

Reproduce from root with ignored four-file overlay .cache/wait-cell-partial-domain.json:
`make test-go vet GO_PACKAGES='./production' GO_TEST_FLAGS='-count=1 -overlay=/Users/stronk/repos/cloud-clicker/.cache/wait-cell-partial-domain.json -run WaitCell -v'`.
New probes/wait-cell-partial-domain.go.txt SHA256
4c5a3c820638f94677d22c90e3c331e863de5d556137d2f281c205807bec243f;
updated probes/production-wait-cells.go.txt SHA
58a4c69c76de702580cb741ce0c9d2dd84ad8d8d225588187fa5515bfcd151ed;
native overlay SHA9966da48602d9b319a49ddf5fb4995172ad776d16ec23aaf05dae0cb39fae4d5.
Linux overlay .cache/wait-cell-partial-domain-linux.json SHA
1d7a82ac920fef4592d2f1af6eb458b33054ecbc81560eeff57add9902f324b1
maps all four manual sources under /workspace; declared Docker --no-deps command
above reproduces it. Prior H3e/provision sources stay unchanged; previous H3f results
remain bound to16b1be49 at5aa403c6, not reclassified as the new guard's evidence.
All handles terminal; --rm Linux run starts no DB. No native hosted/TS/SQL/fullCI claim.

Own draft H3g and RP-681 reconcile the measured result; living plan advances next
to final aggregation/ledger refusal handling and modern contexts, then full workloads.
This preflight certifies ONLY partial arithmetic; it neither adopts an eager error
policy over a valid earlier result nor shrinks accepted caller support. No production,
kernel/math/payout/content/ordinary CI/report/accepted-contract change. RP-680 and27
AC6 failures stay open. Review by: Codex root (research author/first filter). Recorded
by: Codex root. Complete six-path range5aa403c6-exclusive through this containing
checkpoint requires research/normative review outside corrective-code delegation.
Full1.0 goal active/progress; no feature completion, archival, push or release call.

Before final checkpoint, retain a reproducible compiling no-check control for the
new oracle: an IGNORED Go overlay removes only the wrapper's partial-certificate
call, keeping all current constructors/helpers/tests present. The unchanged domain
regression must fail at accepted bad context, not compiler/fixture errors. Tracked
sources stay frozen; this is isolated evidence, not a source rollback or a new run
of the production implementation. It complements the actual pre-guard red above.

Compiling no-check64144/d9453b FAIL .529s at precisely the unchanged bad-context
guard assertion; all three valid companions pass. No compiler, fixture, source-pin
or diagnostic-bound failure. Reproduce by selecting .cache/wait-cell-partial-no-check.json
with TestWaitCellPartialDomainResearch on root make test-go. Ignored source SHA
1c9d5ef1db08f7bd5dd529516f95e5c46517309fa84e75592ce92e528ca2569c;
overlay SHAd06a1de16f658fc2f82090d85b5d97cf8390da99c099b195d203d73efa9e09a4.
Only the wrapper's call to certifyWaitCellPartialsResearch is omitted. Tracked
sources never changed for this probe; prior intact native/Linux results stand,
not a reason to rerun unrelated suites. Handle terminal; expected failure retained.

## 2026-10-16 — Before execution: final multi-segment aggregation definedness

Previous turn PROGRESS468f8212 guards pure partial arithmetic; final aggregation
is explicitly unproved. New attack: v17/tick1000/provider1 produces permits6eMAX,
no cash/burst, horizon2000. Every partial segment up to1000ms may succeed, but the
combined payout should overflow from elapsed1667 onward BEFORE permits' cap24 can
apply. Execute every0..2000 through actual SimulateAdvance; require typed engine/
non-state-ledger-entry failure, distinct from partial-accrual refusal, and unchanged
failed encoded saves. Companion rates4eMAX and4.99999999999eMAX, provider masking,
and zero offline efficiency retain all2001 endpoints each. Controlled inputs are
catalog/codec-admitted seeded origins, not natural acquisition or minted balance.

Prototype constructor must not label these errors unreachable. Retain pre-aggregate
guard miss and actual partial-certificate pass. Proposed manual aggregate check:
capture immutable per-resource canonical prefixes/max exponents, retain partial
certificate output exponent ceilings, then use a SOURCE-DERIVED interior-positive-
sum lemma for <=100000 terms/max input exponent<=MAX-8. This leaves room for grouped
normalization, positive alignment rounding and Q12; no signed/cancellation case.
If that proof is inconclusive, execute actual SumDeterministic/Q12 over each partial
duration with the REAL prefix, not its previously rounded sum. No domain rejection
merely because a sufficient bound fails. Count fast proofs/exact checks/terms.

This is aggregate validity only, not ledger/cap/receipt, progress, modern context,
error ordering or full workload feasibility. Exact fallback can be expensive; do not
hide it in a point count, raise a budget or claim production per-ms scanning is
authorized. Challenge the source lemma with100000-term upper-interior/lower-edge/
mixed-exponent and zero controls, keeping universal proof separate from finite
tests. Retain all prior16032 actual waits/cache/fault/guard controls. Native and
declared Linux AMD64 --no-deps; no runtime/math/kernel/CI/content/adoption change.

Old partial-only prototype74432/b3c1d6 FAIL .964s exactly at missing aggregate
guard: all10005 actual references execute. Bad case succeeds1667/refuses334;
all four2001-point companions succeed, including near-boundary max-exponent input.
Partial certification passes1001 checks, then old cash search returns-1/nil with
zero points. Actual refusal is engine-wrapped non-state ledger entry, not partial
accrual; cap24 cannot repair the already-invalid sum. RP-682 records this prototype
gap. No domain restriction, arithmetic correction or actual harness defect inferred.

Focused aggregate39644/e1ab9c PASS .778s: new guard rejects the real overflow;
all valid near-boundary/headroom/mask/zero-efficiency companions still admit.
Before comprehensive runs, retain raw+partial's unsafe search explicitly, require
the two valid extreme-domain companions to execute every2001 exact aggregate
fallback/3000 input terms, and execute the declared100000-term interior challenges.
No passing sample substitutes for the source-derived sufficient-condition argument.

Comprehensive native82731/c930e4 PASS6.671s plus ordinary production vet: prior16032
actual waits and new10005 all execute with their expected outcomes. Five interior
challenge populations pass; each valid extreme companion executes2001 actual
aggregates/3000 terms, not a rejected or silently excluded domain. Retained long
factory has131074 aggregate fast certificates and ZERO exact fallback checks;
all original partial/cache/fault/query/guard controls remain. Freeze both manual
sources for declared Linux AMD64 --no-deps follow-up using its five-file overlay.

Linux30121/f337be PASS11.578s, same complete26037 actual references/expected
refusals, interior/cache/fault/guard controls and fallback counts. Before final
checkpoint, run one compiling IGNORED-overlay falsification: make the sufficient
interior predicate claim true unconditionally, leaving constructors/metadata/tests
unchanged. The real overflow subpopulation must fail its guard oracle. This probes
proof discrimination, not a production mutation or permission to widen the bound.

## 2026-10-16 — RP-682 aggregate domain guarded in manual prototype

Final native82731/c930e4 PASS6.671s plus ordinary production vet; declared Linux
AMD6430121/f337be PASS11.578s, same sources/populations/outputs, no selected skips.
All prior16032 and new10005 actual-wait references execute: the new bad case has
1667 successes/334 engine/non-state-entry refusals, all four companions succeed.
The partial-only retained path still reproduces unreachable/nil with no point call;
protected construction returns aggregate certificate/company.permits/duration667.
Failed encoded saves stay unchanged; no transient-memory claim. Valid permits
payouts cap at24, but cap application cannot rescue an invalid aggregate upstream.

Both valid extreme companions execute complete2001 exact aggregate checks over3000
input terms; neither is excluded by the sufficient bound. Five interior challenge
shapes (100000-term upper/lower/mixed/zero and empty) pass. Draft H3h derives the
positive-sum exponent margin and makes the pinned normalization premise explicit;
these finite tests are not the universal proof. Immutable per-resource prefixes
share backing storage; maxima come from actual prefix/checked partial values.

Original six factories use aggregate fast certificates3/4/3/4/1/131074 respectively,
and no exact fallback. Retained constructor captures131074 resource terms, keeps
4004 partial checks/five contexts and finds elapsed65536332 with65551 main search
certificates/one actual point. Native setup359.924ms/search1.875361s; emulated Linux
595.281ms/2.805186s. Single observations, not comparative/full-workload feasibility.
Exact fallback remains potentially costly and all its checks/terms are counted.

Compiling forged-proof6767/65d04e FAIL .425s at the real334-refusal guard oracle,
not compilation/fixture errors. Only positiveAggregateInteriorResearch is changed
to unconditional true in ignored .cache/wait-cell-aggregate-forged.go SHA256
1fa59eacdf7188918e1f6109db5f41ac10c12d925a57fa7adf62c32321fa0469;
overlay .cache/wait-cell-aggregate-forged.json SHA
a8fce52c182e9f88e250e8bf00c88a4626df9bc57abe166d034eddf393d130ee.
Select TestWaitCellAggregateDomainResearch/secondary_aggregate_overflow on root
make test-go with that overlay to reproduce. Tracked sources were never changed
for the mutation, so normal final results remain valid; no unrelated rerun needed.

Normal reproduction:
`make test-go vet GO_PACKAGES='./production' GO_TEST_FLAGS='-count=1 -overlay=/Users/stronk/repos/cloud-clicker/.cache/wait-cell-aggregate-domain.json -run WaitCell -v'`.
New probes/wait-cell-aggregate-domain.go.txt SHA256
201fce325af8b5b6d0c92f38e65f312ce087b87362181bcb63eb9d1ac2374e80;
updated probes/production-wait-cells.go.txt SHA
0c1b884275f128e1f16d5ef3509d0ca46a5755bac49f1b21b46575e72fd4a02e;
native five-file overlay SHA282fd00cf70ea7ecfcc30ad161ab43efe5764c3e40ae033d80fa1a2c98d11864.
Linux .cache/wait-cell-aggregate-domain-linux.json SHA
7656fcabc7015c14aa0b3c2be2ce6f7cfedb7c9d27d6ab5ff84db38af2fddce9
maps the same files under /workspace; declared Docker --no-deps command above
reproduces it. Earlier manual sources remain unchanged except this constructor;
prior H3g results stay bound to58a4c69c at468f8212, not restamped as current proof.
All handles terminal; --rm Linux starts no DB. No native hosted/SQL/TS/fullCI claim.

Own draft H3h/ledger/plan reconcile the result. Next actual ledger Add/cap/receipt
refusal proof, then progress/modern/hooks and complete workload/caller closure.
This is not an adopted eager-error policy, a complete wait validity certificate or
permission to narrow caller support. No production/core math/kernel/content/CI/
baseline/accepted-contract changes; RP-680 and27 AC6 failures remain open. Review
by: Codex root (research author/first filter). Recorded by: Codex root. Complete
six-path range468f8212-exclusive through this checkpoint requires research/normative
review outside corrective-code delegation. Full1.0 goal active/progress; no feature
acceptance, archival, push, deployment or release approval.

## 2026-10-16 — Before execution: ledger Add/cap/receipt definedness

Previous turn PROGRESS63658744 guards aggregate validity, not the ledger. Test
eleven complete0..1000 actual-wait populations (11011 references), with one owned
provider producing permits and cash0, admitted v17/tick1000/zero burst. Retune only
in-memory catalog/initial save. MAX=8999999999999999, MIN=-MAX. Profiles:
Add overflow B6eMAX/rate6eMAX/cap9eMAX (hypothesis refuse667..1000); already-capped
B9eMAX/same rate/cap (refuse167..1000); headroom B4eMAX/rate4eMAX/cap9eMAX;
zero B/rate6eMAX/cap9eMAX; already-capped safe B9eMAX/rate1e(MAX-1)/cap9eMAX;
the existing ledger R1 numeric recipe transplanted to permits (B5.6765610215e6,
rate9.81579561656e8,cap9.87256122677e8); tiny no-change B1e6/rate1e-6/cap9eMAX;
zero offline efficiency on the first overflow recipe; receipt gap B1eMIN/
rate1e(MIN+3)/cap1.00000000001eMIN (refuse every1..1000); its zero-start companion;
and safe tiny headroom B1eMIN/rate1e(MIN+3)/cap2eMIN. All other profiles must succeed.

Failure classes must be actual engine/ledger-result or receipt-reproduction errors,
not partial or aggregate admission. Failed encoded saves stay unchanged. Every
successful receipt is parsed and replayed through actual strict Ledger.Apply from
the original snapshot, comparing complete scoped balances. R1 at1000 must preserve
its existing adjusted Delta9.81579561655e8, not echo nominal9.81579561656e8.
Already-cap/tiny/zero-efficiency no-change observations remain explicit.

Candidate manual ledger certificate uses source-derived sufficient cases (no entry,
zero before, safe finite Add with nonbinding cap or already-capped before); all
inconclusive cases execute actual resource aggregate and Ledger.ApplyAccrual, not
reject supported input. Cache only successful exact ledger checks by resource and
raw canonical delta, constructor-local with fixed original balances/catalog. Count
aggregate terms, actual ledger calls/reuse and fast certificates separately. Source
ledger validates resource effects independently before committing, so per-resource
checks must not be mistaken for progress/hook/full-state proof. Retain all26037 prior
actual references and controls. Native + declared Linux AMD64 --no-deps; no new
runtime/math/kernel/CI/balance policy or adopted eager-error precedence.

Old aggregate-only baseline63671/00a2fa FAIL .948s exactly in the three new ledger
guard oracles. All11011 real waits execute: Add failures334, already-cap Add834,
receipt-gap1000;8843 successes all replay through strict Ledger.Apply. Eight full
companions succeed. Old constructor returns-1/nil with zero points in each bad
context. RP-683 records this prototype gap, not incorrect production refusal.
Before correction, tighten the R1 oracle to require its one actual final receipt
(the existing per-change assertion alone would tolerate no receipt); retain all
other populations/values and explicit successful cash0 observations.

Protected focused8195/04df30 PASS .907s: all original11011 actual waits and strict
receipt replays retain outcomes. Before comprehensive execution, add one declared
cache-reuse population: the same R1 recipe offline/efficiency1/cap1000ms through
wall2000ms (all2001 actual waits), so its plateau reuses the identical canonical
payout's successful ledger check. Total new references13012, expected2168 refusals/
10844 strict-replayed successes. Also a pure-ledger synthetic flat-cell control:
the same6eMAX nominal delta succeeds for cash (small before/cap) but overflows for
permits (before6eMAX/cap9eMAX). It must NOT reuse across resource IDs. This is a
component control, not an asserted acquired Company/wait. Existing eleven
populations/values remain unchanged; the twelve-profile census is explicit.

Comprehensive native1644/6a3745 PASS7.408s plus ordinary production vet. All prior
26037 and new13012 actual references/expected errors execute, along with8843+2001
strict receipt replays. R1 plateau hits one exact-delta cache reuse; the cross-
resource component control preserves the permits refusal after successful cash.
Retained long profile uses131074 ledger fast certificates, one initial restore
and no ledger exact/fallback aggregate work. Freeze both manual sources for
declared Linux AMD64 --no-deps follow-up using its six-file overlay. This is not
native hosted/DB/TS/fullCI or complete-workload acceptance.

Linux25269/c73739 PASS15.069s, same complete39049 actual references and all
expected failures/replays/controls. Before final checkpoint, retain two compiling
ignored-overlay controls: force ledgerFastResearch to unconditional true (run the
WHOLE new ledger population so its census remains valid), and separately remove
resource ID from the successful-delta cache key (run the cross-resource component
oracle). Both must fail semantically, not by compilation or skipped populations.
Tracked normal sources remain frozen; these are not production mutations.

## 2026-10-16 — RP-683 ledger validity guarded in manual prototype

Final native1644/6a3745 PASS7.408s plus ordinary production vet; declared Linux
AMD6425269/c73739 PASS15.069s, identical sources/populations, no selected skips.
All39049 actual references execute (prior26037 + new13012). New populations have
2168 expected ledger refusals and10844 strict-replayed successes. Add errors334,
already-capped Add834, receipt-gap1000 are distinct from partial/aggregate failure.
Failed encoded saves stay unchanged; not a claim about transient in-memory fields.
The retained raw+partial+aggregate path still misses each bad context with-1/nil
and zero point calls. Protected construction preserves the actual ledger error.

Every successful emitted receipt is parsed and applied through strict Ledger.Apply
from the original scoped balances; all scoped balances match the actual wait.
R1 explicitly requires one complete receipt at1000 and throughout its offline
plateau, Delta9.81579561655e8 rather than nominal9.81579561656e8. No-change cases
keep balances/receipt empty as declared. The plateau executes1000 real ledger
checks,1002 fallback aggregate evaluations/1001 terms and one cache reuse.
Cross-resource component control proves a cash success cannot bless permits.

Own draft H3i supplies the sufficient-case argument (no entry, zero before,
finite positive Add/nonbinding or already-equal cap), Q12 idempotence and real
fallback/resource independence; proof review remains open. Inconclusive cases
execute actual Ledger.ApplyAccrual rather than reject the supported value. The
shared exact aggregate reader is factored once for aggregate and ledger phases.
Six original factories use ledger fast certificates3/4/3/4/1/131074, one initial
restore each and no ledger exact/aggregate fallback work. Retained native setup
379.981ms/search1.822364s; emulated Linux776.520ms/3.417951s. These are single
observations, not a matched performance/full-workload conclusion. Other inputs
can incur expensive prefix recomputation, exposed by ledger-specific counters.

Two compiling isolated controls fail semantically:
- Forged ledgerFastResearch2399/0e88ab FAIL1.014s: all13012 actual references still
  execute; all three real-refusal guard oracles fire. Plateau's cache-reuse oracle
  also fires because the forged fast path bypasses its exact check.
- Cache resource erasure25961/f14304 FAIL .272s: the component observes one real
  ledger check/one erroneous reuse and no permits error, firing its typed oracle.
No compiler/fixture/census error substitutes for discrimination. Tracked sources
were not mutated; intact final results remain valid without unrelated reruns.
Ignored control source/overlay SHA256 pairs:
fast-forged2e74c3fcb8aaea4cd582bf94679e12f6e666aaa9311dc438f3a2c1cbda07a609 /
94f383661ba0ce492a7e7657626077b46d171213eb512f9fd84c872b6675b039;
resource-alias37c3fae767cb96f6ca57212b58c2a7bed710a27f4379e0d422572d6165fce03e /
6cb7c90f7042f2269376d58eb45921c1e3d2cdd352bc355bc6b02267838974e3.
They are .cache/wait-cell-ledger-{fast-forged,resource-alias}.{go,json}; select the
respective whole TestWaitCellLedgerDomainResearch / TestWaitCellLedgerCacheResearch
with root make test-go. Both handles terminal, expected red results retained.

Normal reproduction:
`make test-go vet GO_PACKAGES='./production' GO_TEST_FLAGS='-count=1 -overlay=/Users/stronk/repos/cloud-clicker/.cache/wait-cell-ledger-domain.json -run WaitCell -v'`.
New probes/wait-cell-ledger-domain.go.txt SHA256
d0ce8e67ed9a4965905a4c962a16fafbe69af0546b9ed6c1af5407fe9e28bc01;
updated probes/production-wait-cells.go.txt SHA
9e54d5943be1bc81442e144d28f10ed8f3c53bfe1aed3b8def1f07155f10a7cb;
native six-file overlay SHA17b91debb6914531efa3d42dd027bfa356be3cf8913fefbe856349c4ad927650.
Linux .cache/wait-cell-ledger-domain-linux.json SHA
4cd6e83b602c4803a19cb97bea503b5855c7661fe424574b8ea46d68e17b0298
maps the same sources under /workspace. Reproduce via declared Docker --no-deps
command above. Prior evidence stays bound to0c1b8842 at63658744, not restamped.
All normal handles terminal; --rm Linux starts no DB. No native hosted/SQL/TS/fullCI
claim. Earlier manual sources are unchanged except the constructor.

Ledger/plan/own draft now point next to progress/modern-context/hook validity,
error precedence and complete workload/caller closure. This is not an adopted
eager-error policy, complete wait proof, production implementation or permission
to narrow caller support. No product/core math/kernel/content/CI/baseline/accepted
contract changes; RP-680 and27 AC6 failures stay open. Review by: Codex root
(research author/first filter). Recorded by: Codex root. Complete six-path range
63658744-exclusive through this checkpoint needs research/normative review outside
corrective-code delegation. Full1.0 goal active/progress; no feature acceptance,
archival, push, deployment or release approval.

## 2026-10-16 — progress-domain follow-up, declared before execution

Continue the manual wait prototype from c366b9ba. Question: can a catalog/codec-
admitted composite progress coordinate refuse after otherwise valid accrual and
ledger commit, while the cash certificate reports unreachable without a point?
Use the v4 fixture retuned in memory: v17 Company, one beige_tower_v2 producing
permits at1/second, cash0, tick1000, tier0 composite resource_log permits/target1e6
weighted1eMIN plus count_fraction(total_owned/1) weighted1. Compare complete0..1000
fresh actual SimulateAdvance calls. Expected first positive progress underflows
the tiny weighted term; zero-time succeeds. Headroom weight1e(MIN+700), masked
producer, zero offline efficiency and a direct resource_log coordinate are full
companions. Require exact stage evidence, receipts/balances and unchanged origin;
inspect failed cloned state rather than assume atomic failure. First run the
unchanged constructor to establish discrimination. Then add a bounded progress
certificate with actual fallback for inconclusive cases, without rejecting valid
tiny-weight inputs, changing live math or settling eager-error precedence. Charge
all preparation work. Native and declared Linux AMD64 overlay runs must retain
the existing populations. Research/spec adoption and complete workloads stay open.

Before execution add one valid near-edge companion: the same minimum weight with
resource_log target5e-15; all positive production endpoints saturate that coordinate
to1 before weighting, so the tiny weighted term itself stays valid. Six populations,
6006 actual waits,1000 expected progress refusals; no narrowed accepted weight range.

Initial23794/0cea20 is an INVALID probe: progress JSON used resource_id instead
of resource, so all six catalogs refused and zero actual waits ran. Correct the
fixture key only, not population/thresholds. Baseline99775/c3b392 then executes all
6006 actual references:1000 declared post-ledger progress refusals and five complete
companions. All preceding partial/aggregate/ledger certificates pass; the existing
constructor fails its new oracle by hiding those1000 refusals. Failed cloned waits
retain the independently asserted permits/cursor changes and encode differently;
the returned evaluation is empty. This is not an externally committed atomicity
defect: the real harness uses clones. RP-684 records the manual prototype gap.
Next sufficient proof uses the generous float64-origin coordinate exponent bound
and fixed purchased counts; inconclusive tiny weights use actual post-ledger
tierProgressPPM calls. Preserve the valid minimum-weight companion, expose setup
counts and retain early-result/error-policy work as an unresolved separate gate.

Protected six-profile6955/bdcb4a PASS .611s. Before full verification extend the
population once: direct count_fraction/required2 with one purchased provider and
its provision rate1e6ppm. All0..1000 waits keep progress500000/delta0, including
the ending boundary that creates one generated tower. This tests the source
distinction between purchased counts and generated counts. Seven populations,
7007 actual references/1000 expected refusals. Tighten valid receipt assertions
and bind protected failure to progress cell0/at1. The three valid tiny-weight
companions must execute1000 actual fallback progress/ledger checks each; the
headroom/direct-coordinate cases use one sufficient certificate. No scope change
to production, math, hooks/modern support or error-precedence policy.

## 2026-10-16 — RP-684 progress guard verified, live search still unchanged

Native93029/c7a000 PASS7.832s plus ordinary production vet; declared Linux
AMD6480633/3823b4 PASS15.224s. Both terminal, no selected skips, --rm container
started without DB (--no-deps). The complete WaitCell population executes46056
actual references: earlier39049 plus7007 new. Seven new complete0..1000 populations
have1000 expected post-ledger progress refusals and6007 successes. Successful
progress deltas and exact permits receipts are asserted; the failed clone's
non-atomic mutation is explicit, not generalized to persisted state. Origin bytes
stay unchanged. The valid minimum-weight, mask and zero-efficiency companions each
execute1000 actual fallback ledgers/progress checks. Headroom/direct cases use the
sufficient guard; provision creates one generated tower without changing purchased-
count progress500000. No excluded near-edge weight population substitutes for proof.

Own draft H3j derives the loose float64-origin coordinate bound, fixed purchased
counts, weighted-product headroom and positive aggregation argument. This remains
an unaccepted research proof, not an implementer-issued normative approval. The
guard calls actual initial progress once, then either sufficient cell checks or
complete actual-ledger/progress fallback with all work counted. Prior six factory
profiles require progress fast checks3/4/3/4/1/65537, one initial check each and no
exact progress fallback. Retained long native setup378.820ms/search1.875967s;
emulated Linux762.946ms/3.272349s. Single observations, not a matched performance
comparison, native hosted measurement or complete workload acceptance. There is
no new cache, narrowed weight domain, live arithmetic/engine/kernel/schema/content/
normal-CI/budget change. The genuine old-source failure above establishes the
regression; no additional per-assertion mutation ceremony is needed.

Normal reproduction from root:
`make test-go vet GO_PACKAGES='./production' GO_TEST_FLAGS='-count=1 -overlay=/Users/stronk/repos/cloud-clicker/.cache/wait-cell-progress-domain.json -run WaitCell -v'`.
Linux: `docker compose -f compose.save-test.yml -f compose.ci-test.yml run --rm --no-deps test go test -p 1 -overlay=/workspace/.cache/wait-cell-progress-domain-linux.json -run WaitCell -v ./production -count=1`.
Ignored overlays retain the previous six manual-test mappings and add
server/production/manual_wait_cell_progress_domain_test.go mapped to the new probe;
the Linux overlay uses /workspace for both sides. Sources are tracked, overlays
are not ordinary CI inputs. SHA256:
- probes/production-wait-cells.go.txt: ecdc86e7f4934ae82a6e7ad3d27b294b7e040cb8144493fde8f58f531f546b35
- probes/wait-cell-progress-domain.go.txt: f2c71476f826b7142b336cca2bf1b1d687efbba5e9518e58174779ee8bc4bfde
- native overlay: 8cef4d64e3bec0f00273b581f4586ce4de83c7bcb1290ab0bda30b4a995178d4
- Linux overlay: aa17ef7c47a84416bc5da95c32d59f5638cd785b313fff5fce8daee18082c77b

Next: earlier-result/error precedence (including the existing already-affordable
shortcut), modern endpoint contexts/hooks, remaining whole-wait/caller identity
and complete workloads. An all-range failure guard is not yet the correct final
query algorithm; no owner decision is silently inferred from it. RP-680 and the27
AC6 failures remain open. Review by: Codex root (research author/first filter).
Recorded by: Codex root. Complete six-path range c366b9ba-exclusive through this
checkpoint requires research/normative review outside the corrective-code
delegation. Full1.0 goal active/progress. No feature closure, archival, push,
publication, deployment or release approval; no SQL/TS/full CI/hosted claim.

## 2026-10-16 — chronological error-ordering experiment, declared before execution

Continue from6ca24449. The whole-horizon certificate is not a query contract:
it may reject a later endpoint before returning an earlier valid result. Preserve
the real earliestAffordable initial-balance shortcut (no transition consumed).
Test one unchanged admitted v17 origin per profile with complete fresh actual
0..horizon waits: cash producer1/second, controlled secondary permits producer,
neutralized cash ladder/synergy, tick1000 except the raw-partial profile tick2000.
Profiles: ledger Add overflow (before/rate6eMAX, horizon1000, first refusal667),
final aggregate overflow (before0/rate6eMAX, horizon2000, first refusal1667),
raw full-segment/partial overflow (same values, tick2000), and minimum-weight
progress underflow (permits rate1, horizon1000, first refusal1). Every profile
also queries an already-affordable0 quote, an early qualifying quote and a quote
whose qualifying point would fall after the first error; retain a cash-unreachable
error case and a valid/masked unreachable control. Compare query result/error
ordering against the chronological exhaustive actual references, not endpoint
monotonicity. First execute the old eager constructor behavior.

Proposed manual mechanism: keep raw producer failure as a stop boundary, never
omit it; certify ranges before pruning; refine an inconclusive/refused range in
chronological order to an actual leaf. A real failing leaf returns its actual
error, not unreachable; an earlier result requires all earlier ranges exhausted
without a hidden error. All setup/repeated certification work is counted and
diagnostic node exhaustion fails explicitly. No per-ms production scan or adopted
search policy, changed payout, modern/hook support or workload feasibility is
authorized. Existing sources/populations remain, native and declared Linux verify
the completed research batch. Broader normative adoption/review stay open.

Before execution include future-rate construction failure: neutral cash producer
dot_matrix_queue1/second, permits-producing beige_tower base1eMAX, purchased1,
provider generating9 towers/boundary. First boundary1000 remains valid; from1001
the new count10 makes the actual rate non-state. Add the declared masked valid
companion as a sixth complete population. Total9006 actual waits/3002 expected
refusals. Four quotes per profile:0, early, after-error (or unreachable),1e12.
The initial helper deliberately composes the unchanged eager constructor/search
so the new chronological oracle must first fail for the real ordering defect.

Initial81085/52a3a7 invalid: removed all cash-generator roles, which v4 admission
forbids; zero actual references. Keep an inert stock_rate role instead. Next
93958/3cbf76 invalid/incomplete: removing a dot_matrix pool source while keeping
its synergy role refused five catalogs; only the future-rate2001 references ran.
Keep that source unless its role is actually replaced. Neither partial run is
the declared census. Baseline6904/675762 then runs all9006 real waits/3002 expected
refusals and fails the intended chronological oracles: time0 and earlier500/1500
results are suppressed by later certificate/raw producer failures. The masked
companion is valid but wrongly constructs/consumes one point for quote0. This
is a prototype composition finding, not a fresh private-harness defect claim.
RP-685 records it. Implement the manual chronological query and retain eager
construction as the explicit control; live search/math remain unchanged.

First corrected focused79439/08e378 PASS .800s: all9006 references/3002 expected
errors,24 chronological queries; earlier500/1500 results survive, all first-error
queries call the actual endpoint and return its error, and zero quotes construct
nothing. Future rate failure is retained at1001, not mistaken for completed source
coverage; raw-partial's stopped endpoint2000 does not hide its earlier1667 error.
Before complete verification add budget1 refusal/zero-budget already-affordable
controls and bind the new query to the existing six complete constructor profiles
via a shared runner, including the retained nonmonotone long wait and zero/
unreachable controls. Select TestWait: it runs the shared12028 reference population
once through the new wrapper (not the old test entrypoint too), retaining all
other34028 references plus new9006, total55062 enumerated actual references.
Expose global and repeated range checks, refinements, source-stop calls and actual
points. Failed proof means refine/check, not a semantic error or unreachable.

Intermediate55386/4b2649 PASS12.568s plus production vet, full55062 references.
Before freezing, close the zero-future/source-first edge: add an admitted origin
with purchased target10 so rates fail from1ms before any cell exists, and assert
that a horizon0 query with insufficient cash returns unreachable without future
preparation/budget/point. This is the same chronological-query boundary, not a
new product decision. Seven profiles10007 references/4002 expected errors, total
56063 with prior populations. A one-node budget is sufficient for the initial
source-stop witness, unlike the other declared refinement-exhaustion cases; assert
that distinction instead of demanding a fabricated budget failure. Run the new
empty-domain oracle before correcting its branch. Earlier results remain dated
intermediate evidence, not final source verification.

Empty-domain baseline19167/156168 FAIL .856s executes all10007 references and
only the seven new horizon0 assertions fail: budget/point validation happens
before recognizing the empty future. Add that explicit zero-horizon branch
after the unchanged initial-balance shortcut and before future work. The new
initial-rate profile's actual1ms stop already passes, including its one-node
budget. Preserve all values/census/expected error identities for final execution.

Final native11370/ac3b47 PASS16.499s plus ordinary production vet; declared Linux
AMD6478995/32eb01 PASS28.924s. Both terminal; all56063 enumerated references and
query/negative/guard populations run without selected skips. These runs overlap
on this machine; their single timing observations are NOT a comparative benchmark.
Keep the three tracked manual sources frozen. Before checkpoint execute two
compiling ignored-overlay controls on the whole TestWaitQueryErrorOrderingResearch:
force whole-domain definedness true (invalid ranges may then be pruned), and
separately drop the producer stop barrier. Both must fail the real first-error
oracles, not compilation/admission/census. Neither changes production or normal
source bytes. The original eager/empty-domain red runs retain separate provenance.

Controls77420/e33a4d FAIL .917s and74204/a4d34d FAIL .990s as intended, all10007
references: bypassing definedness hides/steps past first ledger/aggregate/partial/
progress errors; dropping source stops falsely reports unreachable for future
and initial rate failures. No compiler/admission/census error substitutes.

Final source inspection finds one more future failure still raised eagerly in
the raw header: online provision horizon floor(production/tick)>floor(cap/tick)+1.
Before correction add an eighth profile, admitted cap1000/tick1000/horizon4000,
first actual refusal3000. The same cash1/second has an earlier500 result; later
3.5/1e12 queries must hit the real3000 error. This is RP-685's same boundary, not
a new release/product rule. Total14008 references/5003 errors, full60064. Run it
red first, then defer the header's known failure boundary and preserve earlier
valid cells. Previously normal runs/controls remain bound to their intermediate
sources and seven profiles, not relabelled final after this extension.

Header baseline69474/f20196 FAIL1.225s: all14008 actual references run; only the
new horizon_guard profile fails (lost500, no actual3000 error, wrong budget result).
Deferral uses first invalid duration (floor(cap/tick)+2)*tick, clips prepared
production to one millisecond earlier, and retains the mandatory actual stop
point. An earlier producer failure replaces that later stop. Ordinary eager
construction keeps its original guard; the query does not call this a cash
plateau or expand the accepted online/offline production allowance.

## 2026-10-16 — RP-685 chronological research query verified

Final native29892/524e35 PASS22.244s plus ordinary production vet; declared Linux
AMD6463770/31670e PASS38.927s. Both terminal, all60064 enumerated actual references
and retained controls run, no selected skips. Eight new profiles have14008
references/5003 expected refusals.32 main queries compare earliest results/first
actual errors/unreachable against exhaustive chronological references. Time0 and
empty-future controls need no preparation/budget/callback; node exhaustion is a
distinct failure. The actual initial/future rate stops at1/1001 and online-horizon
stop3000 are mandatory, not missing tails. Raw-partial's earlier1667 error beats
its stopped full-width endpoint2000. Earlier500/1500 purchases survive later
failures; failed proof alone supplies neither an error witness nor unreachable.

The shared six-profile runner retains all12028 original references/fault controls
and binds the new query/main/zero/unreachable calls without duplicating that census.
The retained nonmonotone query finds65536332 with65551 nodes, one actual leaf and
one global domain certificate; its4004 partial checks/other setup remain counted.
Repeated inconclusive-range certificates are also counted (not cheap by assumption).
Native/Linux full runs overlap; timings are single contended observations, NOT
a matched benchmark, full workload result or justification for new work ceilings.

Final controls62468/37672a FAIL1.502s and37443/d3dc66 FAIL1.623s intentionally,
all14008 references. Definedness bypass hides/steps past the real first errors;
dropped stops falsely report unreachable at the initial/future-rate and new
horizon barrier. No compilation, admission, truncation or census failure is the
oracle. Normal tracked source bytes were unchanged; no repeat of normal runs
is needed because only ignored alternative helpers were executed. All handles
terminal; --rm Linux starts no DB. No production mutation, TS/SQL/hosted/fullCI
or release claim. RP-680 and27 AC6 partitions stay open.

Commit the portable nine-file probes/wait-query-overlay.json rather than require
this machine's ignored absolute overlays. Relative mappings use the server cwd
already owned by root Make and the declared Docker service. Consumer verification:
native78483/125821 PASS1.019s; Linux4777/15852d PASS2.049s, complete new14008-
reference population on the same frozen sources. Full reproduction from root:
`make test-go vet GO_PACKAGES='./production' GO_TEST_FLAGS='-count=1 -overlay=../planning/clout-v1-and-pr-interns/probes/wait-query-overlay.json -run TestWait -v'`.
Linux: `docker compose -f compose.save-test.yml -f compose.ci-test.yml run --rm --no-deps test go test -p 1 -overlay=../planning/clout-v1-and-pr-interns/probes/wait-query-overlay.json -run TestWait -v ./production -count=1`.
Full normal final executions above used ignored absolute manifests mapping the
identical source bytes; portable consumer executions verify actual relative-path resolution.
This is not a separately executed fresh-clone or clean-host release gate.

Final tracked source SHA256:
- production-wait-cells.go.txt: f8c01cf3bb1e55e92ee24ff274add0c40193eccb78f52da2ea7cf52cbe927b46
- wait-query-order.go.txt: 03c316c25ce5aa62c07bc30c6dd87d3bf9d6e4e71e419c209ab785760e309728
- wait-query-order-test.go.txt: 4abca278df87f84336b4d5a3da5fbf26981ba1573c556b4adabf19a4b6ffce0b
- wait-query-overlay.json: fb2d780a13620f6fb6103eecf2e9f62d37dc23a11b5fcf0e0f81c333ee9c2e2b
Ignored normal native/Linux overlay hashes1611dc70212f64a44c7bfef3ae3dd562b4bbd888d99fbf0dbc86c254920ebda2 /
e7c947747f9dee38b91d853eed9e52cc152a20588526c89f772fd9a8ded1e836.
Control helper/overlay pairs (.cache/wait-query-{proof-bypass,stop-dropped}.{go,json}):
fb5cb646f27be8447ce2f38ac2005ef96cddf6b8c87b9c1f812981dfbed72240 /
f4bf35dfb2f5b965dc9b7e0d0c653f0c4efac0cefeb02606f51bb7ded4cdbc7e;
7b9e01cd593f662ddc6f93eb2fb822f8506d23f66af991bceed4f69439dab77b /
c36db484626ee1f589e2b72c045c9c100dbed4bb26322bce8f16823da14c6b40.
They replace only the query helper with the documented one-line faults; prior
intermediate source identities/results are not restamped as final.

Own draft H3k states the conditional ordering proof and limits. Next: modern
endpoint contexts/hooks, remaining global/caller/resource/state semantics and
error-policy compatibility, complete workloads under accepted budgets and lawful
adoption. No accepted error policy is inferred from this experiment. Review by:
Codex root (research author/first filter). Recorded by: Codex root. Complete eight-
path range6ca24449-exclusive through this checkpoint requires research/normative
review outside corrective-code delegation. Full1.0 goal active/progress; no feature
acceptance/archival/push/publication/deployment/release or active arithmetic change.

## 2026-10-16 — stock-hook range definedness, declared before execution

Continue from c1162c29, without enabling existing legacy harness constructors.
Question: can a successful later v19 faction-stock wait hide earlier event-bearing
refusals, and can a source-bound attendance split certify the hook domain safely?
Use actual axisContentBundle/axisTimingState, codec+foundation-admitted Company19,
faction bootstrapper, stock cap-1, progress interval-1500, one purchased stock-role
tower. Pinned role1000ppm and catchup ceiling5000 imply saturation at1499ms;
masking that tower leaves the base stock rate and saturation at1500ms. Full fresh
SimulateAdvance references0..6001 in both modes, four populations: no hook,
saturating hook, masked saturating hook, already-capped hook. Total48016 reference
waits,14006 expected event refusals (1499/1500 through5000), with successful5001+
controls. These are direct fixed-context simulation primitives, NOT logged active-
play scheduling/catchup or a migrated modern harness.

Probe the real concrete faction hook through its actual simulation wrapper on
fresh origins; verify its stock read/write set, version/attainment preservation,
event reason and agreement with full waits. Compare actual no-hook cash controls
and full encoded states at named anchors after normalizing only the three stock
fields. Thirteen named intervals per profile/mode must match independently counted
whole-reference validity, including0, threshold neighbors, cutoff and mixed ranges.
First retain an unsafe endpoint-only certificate so the mixed-range oracle fails;
then split using the actual strict cutoff and probe each regime's end. Source
argument uses the hook's nonnegative exact-integer rate/overflow/event thresholds,
not a false monotonicity assertion about production payouts. Expose all helper and
range probes; no arbitrary callback purity, new event policy, runtime change or
numeric/shared-vector change. Native and declared Linux execute this new subject;
unchanged prior60064 references are not needlessly rerun or restamped.

First compile2b1ef3 invalid (unused decimal import); no tests executed. Remove
that unused import only. Endpoint-only baseline72098/6df21f FAIL14.573s executes
all48016 full wait references,14006 expected event refusals,64 additional full-
state anchors and104 range checks. Only the eight mixed-cutoff assertions fail
(two ranges × two saturating populations × two modes). Every actual stock
read/write-set, event reason, cash, role and origin/version check passes. RP-686
records this proposed proof gap, not a new live bug or the guarded legacy
constructor admitting hooks. Correct the certificate by splitting at the actual
catchup ceiling, without changing any source hook, wait event policy or population.

## 2026-10-16 — RP-686 concrete stock-hook component verified

Final native51196/7bffd7 PASS15.896s; declared Linux AMD6446283/b125ff PASS34.374s.
All48016 actual full-wait references,14006 expected event refusals,64 additional
no-hook full-state anchors,104 positive/negative range checks and24 invalid-range
controls execute. No selected production skips. The native selector also built
./faction but selected no faction-package tests; do NOT cite that as a green
faction suite. Ordinary vet passes ./production ./faction. Both handles terminal,
--rm/--no-deps Linux starts no DB; overlapping single timings are not performance
comparisons, hosted CI or full workload evidence. No TS/SQL/shared-math/fullCI claim.

Saturation remains an actual emitted event and therefore an existing wait refusal,
not a newly accepted success/private-event policy. The corrected range component
checks the attended branch's end before any absent branch. All no-hook and capped
valid ranges stay certifiable; endpoint-only's eight mixed-cutoff failures are
the retained old-source regression. No extra per-assertion mutation is necessary.
Actual hook read/write-set probes match every full wait's stock fields and returned
stock-role observations. Masked base production still reaches saturation at1500;
unmasked reaches it at1499; both succeed again after5000. Cash controls and64
complete encoded-state comparisons normalize only the three stock fields, including
event-refused clones' mutation. Origins/version/attainment/masks remain intact.

Wrapper-probe work is explicit: six hook-bearing populations each6001 reference
probes; their13 range checks need12 more probes when saturating or14 when capped,
and two attendance splits. Two none populations execute zero wrapper probes.
The wrapper still owns its actual stock-role counterfactual work; these are not
counts claiming each inner call free or an accepted production work bound. Hook-
only range definedness cannot authorize pruning without other component proofs.

Own draft H3l derives the exact-integer factor/overflow/saturation argument and
conditional pre-hook encoding/read-set obligations. It binds existing fixed-
context SimulateAdvance only, not complete logged scheduling/catchup/private-event
ownership. Existing legacy constructors still reject these modern/hook contexts;
no DG-D/axis hold or arbitrary-hook safety assumption is bypassed. No production,
kernel, catalog, schema, normal-CI or accepted-contract bytes change. The previous
60064 reference checkpoint remains bound to c1162c29; it was not rerun/restamped.

Reproduce from root with committed relative overlay:
`make test-go GO_PACKAGES='./production' GO_TEST_FLAGS='-count=1 -overlay=../planning/clout-v1-and-pr-interns/probes/wait-query-overlay.json -run TestWaitStockHookDomainResearch -v'`;
`make vet GO_PACKAGES='./production ./faction'`.
Linux: `docker compose -f compose.save-test.yml -f compose.ci-test.yml run --rm --no-deps test go test -p 1 -overlay=../planning/clout-v1-and-pr-interns/probes/wait-query-overlay.json -run TestWaitStockHookDomainResearch -v ./production -count=1`.
New probe SHA25684ab36c7b0ca92ac3dd911c9a89e168085aab9f38fd218bed41a6aa3b32bdce9;
updated ten-file overlay SHAe88bbf54634f74b3500603a73d7f6e6f0a3ebe112354ff7bb288a5d52db37f66.
The overlay adds only this manual source; old source bytes are unchanged.

Next: integrate the modern producer/query factory with its exact operation and
remaining concrete hook families, complete global/caller/context/error-policy and
workload proof, then obtain lawful review/adoption. Do not repeat passing stock
populations or treat this component as full modern/hook support. RP-680 and27 AC6
failures remain open. Review by: Codex root (research author/first filter).
Recorded by: Codex root. Complete six-path range c1162c29-exclusive through this
checkpoint requires research/normative review outside corrective-code delegation.
Full1.0 goal active/progress; no feature acceptance, archival, push, deployment or
release approval. Meaningful progress is the new actual hook-domain counterexample
and corrected source-bound certificate, not the reference count alone.

## 2026-10-16 — fixed modern producer/query binding, declared before execution

Continue from e8016a01. Compose the real Company19 fixed-context producer with
H3k query ordering and H3l's concrete stock-hook domain, not logged scheduling or
the parent draft's unbuilt private-event API. Existing legacy entrypoints must
retain their version/hook refusals. A separate manual fixture owner reloads the
already-admitted pinned artifact bytes with the existing test loader, owns the
encoded origin and copied external/mask inputs, and exposes detached full actual
selected states/results, not a cash-only surrogate. This is not a production
bundle-admission API or permission to enable modern live harness constructors.

Twelve complete0..6001 populations (six × online/offline) use admitted axis state:
none; saturating stock; masked saturating stock; already-capped stock; a validated
fixed external guild.stock_consumption factor2 under none; and cap1500ms plateau
under none (retuned fixture/hash, not minted content). Total72024 full references,
14006 expected stock-event refusals. Four quotes each: original balance, actual
no-hook cash at750, at2000 (use original+1 when masked), and1e12. Exhaustive
chronological references own earliest result/error/unreachable. Verify entire
selected encoded state, actual evaluation/role data, source19, no state on error/
unreachable, zero-time semantics and unchanged source. Mutate returned scratch
and original input state/mask/external/catalog buffers to test independent reuse.
Invalid source version/recipe/mask/contribution/hash must refuse, not fall back.

First run the combined producer with its stock-domain binding intentionally
omitted: the new oracle must expose hidden/late hook failures. Then connect the
existing component, counting producer and hook preparation separately and avoiding
repeated producer certification when only the hook range is inconclusive. Keep
all other global/caller/clock/replay/encoding and complete workload/adoption gates
explicit. Finish native and declared Linux against affected legacy consumers too;
do not turn this bounded fixed simulation into full modern/logged support.

Unbound integration97585/15de8d FAIL9.317s, all72024 references/14006 refusals and
48 queries. Only stock/masked first-error queries fail: unmasked probes error at
2000 instead of1499, while high quotes and masked queries return unreachable/nil.
The earlier valid producer/query results pass. RP-687 records this manual binding
gap. Before final verification, self-review catches that selected-state comparison
reuses owner.project, although the exhaustive outcome oracle is independent.
Replace that comparison with the same direct, caller-bundle reference function,
including independently constructed faction hook; do not cite the old comparison
as proof against a faulty projection method. Then bind H3l, add declared ownership/
refusal controls and retain a projection-hook-disconnection control to test the
full-state oracle independently of range safety. No live implementation changes.

Bound focused16135/1a8a7b PASS9.470s, all72024 references and48 queries after
independent selected-state comparison. Before complete verification add the
declared mutable-input/returned-state/role-buffer isolation, zero-time/empty-
future/budget and seven invalid-context controls. These do not add another
reference census or establish a general untrusted bundle-admission API: the
manual owner reloads already-admitted fixture artifacts with the existing test
loader. H3l's event reason now uses the state's actual declared faction rather
than a bootstrapper literal; its original population stays unchanged. Freeze
all manual sources for full affected legacy/stock/new query verification.

Full native40648/7bde11 PASS37.644s plus ordinary production vet; corrected declared
Linux83896/71485b PASS83.262s, all180104 enumerated actual references (prior108080
+ new72024), plus ownership/refusal and existing controls. Initial Linuxb16825
ran no tests: a mistyped overlay path omitted the hyphen in pr-interns. Correct
the command only; do not count that exit as a code failure or the later pass as
a retry hiding one. All sources frozen throughout the real runs, handles terminal.
Full runs overlap; elapsed figures are not a performance comparison or full
scenario workload proof. Before checkpoint execute the prepared compiling ignored
overlay that disconnects only owner.project's hook, leaving the real stock-domain
guard and independent direct reference untouched. It must fail the full selected-
state oracle as well as first-error cases; no production mutation or guard waiver.

## 2026-10-16 — RP-687 owned fixed-modern composition verified

The final native/Linux results above cover all180104 enumerated references:
prior60064 legacy +48016 stock +72024 new, plus distinct anchor/query/control calls.
New twelve-profile population has14006 expected event refusals/48 main queries;
ownership, zero-time, empty-future, budget and seven refusal controls pass. Full
selected encoded state/evaluation/all roles match independent direct caller-
bundle SimulateAdvance references. Producer certificates stay reusable when only
the hook domain requires refinement; actual leaf hooks are never omitted.

Point-only hook disconnection10615/104f98 FAIL18.167s intentionally, all72024
references/48 queries retained. It fires SIX full-state mismatches at750/2000 and
EIGHT first-error mismatches despite intact domain safety and matching cash. All
three none populations remain green. No compilation/admission/census failure is
the oracle. Normal source bytes were never mutated; the compiling alternative is
.cache/wait-modern-no-point-hook.{go,json}, replacing only owner.project's first
recipe branch, not query guard or independent references. All handles terminal;
no normal rerun is necessary on unchanged source. No production/TS/SQL/hosted/
shared-math/fullCI or full-workload claim; overlapping timings are not benchmarks.

The fixture owner copies state, all mask slices/external values and artifact
bytes, reloading parsed objects with the existing trusted test loader. Mutating
caller inputs/faction/catchup objects or returned Company/role buffers cannot
change a subsequent query's full result, owned origin or pinned hash. This is
NOT general untrusted bundle admission or full C1e fork/action/cache identity.
Begin exposes one reload,49498 artifact bytes (49494 cap fixture),2816/2817 origin
bytes and time. Query restores/leaves and producer/guard/hook work are separate;
actual contribution assembly occurs at validation and source construction too.
There is no new accepted work ceiling or permission to hide preparation costs.

Full reproduction from root:
`make test-go vet GO_PACKAGES='./production' GO_TEST_FLAGS='-count=1 -overlay=../planning/clout-v1-and-pr-interns/probes/wait-query-overlay.json -run TestWait -v'`.
Linux: `docker compose -f compose.save-test.yml -f compose.ci-test.yml run --rm --no-deps test go test -p 1 -overlay=../planning/clout-v1-and-pr-interns/probes/wait-query-overlay.json -run TestWait -v ./production -count=1`.
Focused selectors: TestWaitQueryModernBindingResearch, TestWaitQueryModernOwnershipResearch,
TestWaitQueryModernRefusalsResearch. Final tracked SHA256:
- production-wait-cells.go.txt: 3c9b17dd80b0048ad181cb27b826b763d8109ec55161d5e552d527d83ee1c644
- wait-query-order.go.txt: 6c2f420e1750f2347e51c72e7b1e7be67d9048bee9140c5c0df586c7a797bf97
- wait-stock-hook-domain.go.txt: 206c9772485800e4ddacd71ddb2a967b7e29b3abe10a3e496c0c807242121324
- wait-modern-binding.go.txt: 495e2cb33d10c6d9049fd849f36cc43b920402ce0e5e3f1e22ab1472dae54e3d
- eleven-file wait-query-overlay.json: 721ff8f2e79d3af845ffd866bb2fdfc1723dd7040559319f0681831a741b924b
Ignored fault source/overlay SHAa6a57ed2cb07a17fb86bbc1d701108a2b5572480114319ac1ea8cbe6953ef986 /
325aefa0e82e5f90e64d4b7d45fc7fb21225253aa070fa1dcc4fa69760a7cc16.

Own draft H3m records the bounded composition. Existing legacy/axis guards and
active product/harness/kernel/catalog/schema/normal-CI/accepted contracts remain
unchanged. Other hook families, logged stages/clock, global future-time/encoding/
state/caller/resource and lifecycle proof, complete work/compatibility and lawful
adoption stay open. RP-680 and27 AC6 failures remain open. Review by: Codex root
(research author/first filter). Recorded by: Codex root. Complete nine-path range
e8016a01-exclusive through this checkpoint needs research/normative review outside
corrective-code delegation. Full1.0 goal active/progress, not preview/review-only.
No feature acceptance, archival, push, publication, deployment or release approval.

## 2026-10-16 — real first-hour input grounding, declared before execution

Continue from9dd996b2, toward the first_hour_lifetime adapter. Do not copy its
math into another authority or claim synthetic v19 cases are real caller binding.
Use the already tested reference.greedy seed0 experiment tuple (200,2e0,50,1e4,10)
with ratified scenario/policy and current pinned bundle. A manual harness test
constructs the same initial runtime, drives ACTUAL step(boundary), then requires
its complete result AND command script to equal public RunExperimentScript.
No production observer, source rewrite or harness behavior change is authorized.

Capture three distinct post-step states: first run's first purchased generator,
first run reaching ten purchased generators, and the new second run. Validate
with the actual first-hour state checker. For each, call the actual reference
ranker on its captured cursor over the ratified1000ms decision cadence, with one
real generator quote and the first unowned upgrade quote. These are post-step
diagnostic queries, NOT an invented record of the calls inside the previous step
or proof a purchase would pass all intent/gate requirements.

Each snapshot also has1001 fresh full ranker.advance references at0..1000, real
lifetime hook included; retain cash/lifetime values and full state/evaluation/role
anchors0,1,500,1000 plus any actual selected result. Require current ordinary
external context nil, canonical quotes, preserved source version/origin and no
source mutation. Expect3 snapshots/6 quote queries/3003 enumerated references;
failure to reach a declared population invalidates the measurement, not exclusion.
Only after complete validation write one generated ignored .cache input packet
for a named production-prototype consumer. Include artifacts and existing save
encoding, scenario/policy/constants identity, complete source results and query
provenance. This is an identified research producer/consumer handoff, not a new
runtime origin schema, saved-payment protocol or fulfilled C1e API. Native and
declared Linux must reproduce it; no live fixture/body/budget/CI/schema change.

Capture execution note: initial compile0119b0 ran no tests (unused time import,
removed). Session36338 ended without a recoverable terminal result during context
compaction; no packet existed, so no success is claimed. Fresh658296/52491 ran
the test and FAIL1.491s: generator.beige_tower's actual internal quote does NOT
roundtrip through its canonical display string. This is an honest instrument
failure, not a product regression. Prospectively preserve the real normalized
mantissa's IEEE bits as a hex string and exponent, asserting exact reconstruction;
keep cost string only as display. No quantization of the queried price. This is
a research-packet encoding detail, not a change to game wire/save precision.

Revised capture f2bd03/15618 PASS2.770s: all three actual source14 states,
3003 full advances, six actual quotes; full public/manual report and19 commands
equal,11916 original transitions. Packet228362 bytes SHA256
f8a80a6c0d1f3d1ecc7c041edc00e26784676fda16e6f237af8462dbdc66f921.
All six catalog quotes are honestly unreachable inside this one-second window.
Do not disguise them as positive-selection coverage. The named production
consumer will compare all3003 cash/lifetime observations, all12 full anchors and
these six queries using exact raw prices. Add two explicitly DIAGNOSTIC thresholds
per snapshot: original cash and actual500ms cash, not catalog purchases. Require
at least two positive-time selections overall; these exercise selected full state.

The separate legacy-lifetime adapter preserves source14 and uses the real
prestige.AccumulateLifetimeValue, not copied math or the offline-span hook. Its
sufficient range guard only certifies canonical nonnegative original lifetime
and a nonnegative cash cap with BOTH exponents<=MAX-3. Given certified canonical
nonnegative cash receipts reproducing after<=cap, receipt delta exponent<=cap
exponent (positive Add cannot reduce its maximum input exponent). Positive Add
and Q12 can carry at most two decades, so lifetime remains valid. Missing/extreme
caps or lifetime are inconclusive, forcing actual chronological leaves, never
silently invalidating supported inputs. No global lifetime proof/adopted budget.
Freeze sources for native/Linux and a point-hook-disconnected negative control;
the external real-harness packet remains the independent oracle.

## 2026-10-16 — first-hour lifetime research binding verified

Actual producer nativef2bd03/15618 PASS2.770s; declared Linuxb5341e/96604
PASS5.205s, byte-identical228362-byte packet. Same real3 origins/3003 references,
12 full anchors/six actual prices,19 commands and11916 runner transitions.
Source14 cursors10000,53575,900000ms and lifetimes0,1.7420122e2,0. Public/manual
result AND complete command-script comparison pass. The packet is generated
ignored data, reproducible from the committed capture, not another catalog.

Initial production consumer918429/54248 PASS0.473s matches3003 observations,
12 full independent harness anchors/12 queries, including two diagnostic500ms
selections. Add the explicit nonnegative-before premise to the sufficient guard
and six direct domain controls (zero/interior/extreme/negative/noncanonical,
including a valid actual ignored delta despite an inconclusive certificate).
Then freeze both source files and the thirteen-file portable overlay.

Full affected native14cfea/2939 PASS38.509s plus production vet; declared
Linux115626/46906 PASS82.623s. Prior180104 enumerated actual references retained
plus3003 new production comparisons =183107, with separate real producer runs,
anchors, selected-state checks and controls. No selected tests skip; sources
frozen throughout. Runs overlap and are not comparative benchmarks. No TS/SQL/
hosted/full harness workload/fullCI or shared numeric change claim.

Compiling ignored point-hook disconnectionabf66c/56035 FAIL0.488s intentionally:
all3003 points/12 anchors/12 queries retained,2000 lifetime mismatches and EIGHT
full-state/selection mismatches, while actual cash/query outcomes still agree.
Zero-production run2 companion remains green. Independent harness packet and
query guard are untouched. Normal source was never mutated; all handles terminal.
The failed canonical-price capture above remains disclosed, not overwritten.

Reproduction, in order from repository root:
`make test-go GO_PACKAGES='./harness' GO_TEST_FLAGS='-count=1 -overlay=../planning/clout-v1-and-pr-interns/probes/wait-query-overlay.json -run TestWaitFirstHourCaptureResearch -v'`,
then `make test-go vet GO_PACKAGES='./production' GO_TEST_FLAGS='-count=1 -overlay=../planning/clout-v1-and-pr-interns/probes/wait-query-overlay.json -run TestWait -v'`.
For Linux use declared compose.save-test.yml+compose.ci-test.yml service, --no-deps,
`go test -p 1 -overlay=../planning/clout-v1-and-pr-interns/probes/wait-query-overlay.json -run TestWaitFirstHourCaptureResearch -v ./harness -count=1`,
then the same flags with `-run TestWait -v ./production -count=1`.
Focused production selectors: TestWaitFirstHourLifetimeBindingResearch and
TestWaitFirstHourLifetimeDomainResearch. A missing generated packet fails loudly.
Negative command uses -overlay=../.cache/wait-first-hour-no-hook.json and
-run TestWaitFirstHourLifetimeBindingResearch; only the adapter's real helper call
is replaced with a compiling no-op. Exact SHA256:
- capture:0c77efaa69f9a7a8151bb1a50aeab4f805248c90580e5bee82ae39d3c5cb0b82
- consumer:1a9bb8a3887a1398c0af5f566a55bdb36283493b4c83ee796da1bc5b51a27166
- overlay:48aa5b92b466e20ac3883666014f00f4b1e693fae9ef1985eee4c3c844562322
- packet:f8a80a6c0d1f3d1ecc7c041edc00e26784676fda16e6f237af8462dbdc66f921
- fault source:1ad0e06ee94f49b5b231043f1fbd8760a11c7f79c2f7fa8c5c8a69a481293802
- fault overlay:87fe0d56dd92aa75331fd3990a1bc1993ae8781f8ab62bf544ba5cee40f73034

Draft H3n/own plan capture the bounded result and next remaining closed replay
recipe/logged stages. Existing modern owner/legacy guards, production math,
kernel222/epoch8, game data, client, saves and CI remain untouched. Exact raw
price identity is a research transport constraint, not a discovered game-wire
bug or reason to mint another defect row. RP-680 and27 AC6 failures stay open.
Full domain/caller/lifecycle/workload/research-normative adoption gates remain.
Review by: Codex root (research author/first filter). Recorded by: Codex root.
Six-path range9dd996b2-exclusive through this checkpoint is NOT designated
approval; research/normative review is outside corrective-code delegation.
No implementation acceptance, archival, push, publication, deployment or release
claim. Full nine-tier1.0 goal remains active/progress; meaningful safe work remains.

## 2026-10-16 — closed replay fixed-query reference scope

Continue from6fe8e3b4. C1d/C1h remain unadopted: no production change or removal
of SimulateAdvance's intentional event refusal. Research projector uses actual
simulation policy/contribution/evaluation and simulationHook(closedReplayAccrualHook),
retaining ordered private EventWrite values. Independent reference uses the real
SimulateAdvance with a test tap that returns no events to its legacy envelope,
captures ONLY the main scratch state's events, and forwards the actual full chain
and errors. Counterfactual hook calls must not duplicate main projection events.

Eight complete0..6001 populations: online/offline x full Guild+Commons+stock chain,
generator-masked companion, plain Prestige-only effects, and missing Commons
weight refusal. Modern19 pinned axis fixture, source time12:59:59.500, frozen
Guild factor1 and Commons weight500000, EMPTY pending clearing batch. Full/masked
stock starts cap-1 with interval-1500 progress, so actual attendance/stock/hour
boundaries are included. Plain has no faction/Guild/Compact membership. Invalid
weight is deliberately absent on a Compact member, not silently substituted.
For all48016 points compare full encoded state/result/roles/ordered event payloads
or identical actual errors; no excluded population. Query origin cash, independent
750/2000ms cash and1e12 thresholds, preserving first-error/zero-time/unreachable
distinctions. Thresholds are diagnostic, not catalog purchases or pricing changes.

No closed-chain range proof yet exists. Its explicit inconclusive guard MUST
force chronological real leaves, with work visible under the existing diagnostic
node ceiling. This reference is not an efficient final search, accepted workload
budget or completed domain proof; do not label leaf enumeration a producer proof.
Require selected full state/events against the independent real primitive, owned
weight/input/result-payload isolation, no intent IDs/publisher/Store, and legacy
event refusal unchanged. A compiling projection-only Commons disconnection must
fail while the reference and source preparation stay intact. Native/Linux affected
checks precede checkpoint. No live logged stage/catchup/settlement/SQL claim.

Clarification before execution: constructor requires valid Compact weight binding.
The missing-weight population deliberately severs that dependency AFTER a valid
research construction and compares the real primitive's error propagation. It
is not a claim the proposed BeginCalculation may admit missing required inputs.

## 2026-10-16 — closed replay fixed-query reference checkpoint

Initial projection/reference smokece3397/46296 PASS47.048s:48016 complete points,
12002 expected refusals,24004 event-bearing references,32 queries and24012 actual
oracle counterfactual calls. It did not prove a returned-query ownership API.
Move result selection/clearing into the owned research query, count real main/
counterfactual hooks and restores, add mutable weight/input/result-payload controls.
Replace the draft's self-canonicalization origin check with a separately copied
pre-query origin, and compare full evaluation/roles on the repeated selection.
Standalone ownershipff2510/76451 PASS0.899s precedes that final assertion tightening.
Final source/portable overlay frozen before both architecture runs and fault run.

Final nativea2baaf/76756 PASS49.762s plus production vet: all eight complete
populations, ownership/refusal/budget/empty-future controls and the existing actual
event-ownership/long-absence-stage tests execute. The latter retain both modes,
actual full ApplyLogged comparison, original24h cap/12h banking and missing-clearing
negative. No need to rerun unchanged183107 prior references from6fe8e3b4: neither
their helper sources nor ordinary production changed. These focused checks are
not a full production/harness, TS, SQL, hosted or CI claim.

Compiling projection-only Commons disconnection211cde/57126 FAIL55.340s as
required, with the full48016 points/32 queries retained and36016 semantic mismatches.
Cash/nominal valid purchase times still agree; full state/events do not. More
importantly, severing Commons loses its actual1ms missing-input refusal and instead
returns positive purchases869/1738ms online,966/1931ms offline. Normal source,
reference chain and preparation are untouched. No compile/admission failure is
counted as discrimination. Negative source/overlay remain ignored .cache files.
Native/fault runs overlap the declared Linux run, so timings are not performance
comparisons; the linear reference is explicitly NOT accepted workload feasibility.

Real cost is exposed, not excused by successful pruning: the unreachable6001ms
query executes6002 full projections/12003 query nodes/12004 guard calls. Positive
750/2000ms queries execute751/2001 points; first-error queries execute0 then1ms.
For full/plain profiles each owned reference executes14756 projections,29512
restores,29504 hook calls including14752 counterfactual calls across its complete
enumeration and queries. The masked profile avoids counterfactuals but still
executes12004 projections. These are diagnostic work observations, not new caps
or evidence that a65M-endpoint query can afford this strategy. A source-derived
closed-chain range proof and complete real workload measurement are next gates.

Declared Linux046f15/91975 PASS224.698s, all48016 paired points/32 queries and
ownership controls; existing event/long-absence stage populations also pass in
both modes. This is emulated Linux AMD64 Go1.26, not hosted hardware or a timing
comparison with native Go1.27.2 ARM64. All owned handles terminal. Sources remained
frozen and hash-identical across final native/Linux/fault runs. No tests skipped.

Root reproduction, focused affected boundary:
`make test-go vet GO_PACKAGES='./production' GO_TEST_FLAGS='-count=1 -overlay=../planning/clout-v1-and-pr-interns/probes/wait-query-overlay.json -run "TestWaitClosedReplay|TestAxisCalculationEventOwnershipResearch|TestAxisCalculationLongAbsenceStagesResearch" -v'`.
Linux: `docker compose -f compose.save-test.yml -f compose.ci-test.yml run --rm --no-deps test go test -p 1 -overlay=../planning/clout-v1-and-pr-interns/probes/wait-query-overlay.json -run 'TestWaitClosedReplay|TestAxisCalculationEventOwnershipResearch|TestAxisCalculationLongAbsenceStagesResearch' -v ./production -count=1`.
These selections need no generated first-hour packet; a full -run TestWait still
requires H3n's capture first. Fault command selects TestWaitClosedReplayBindingResearch
with -overlay=../.cache/wait-closed-no-commons.json; copy only the manual source/
overlay, then slice the projection's actual four-hook accrualHookChain to[:3]
before its observer wrapper. No production source mutation. Final SHA256:
- closed reference:fb60ffc90eaaba6c7a64b4ace0a3429fdc3823774d06d139e1f7728473283c95
- fourteen-file overlay:281c78b4e597e124676cd5f8cfa05c5d72dd302a1940db0fb42a2f20c2d9c9ba
- fault source:b71636bb91355ea51368a1533b1abfc4345b448d9359f6347c6ed1c82d9a5720
- fault overlay:3c5bd9cd220dfeb8bdad391f9baf83148855154f1651d8f2cc740d752b429ce0

Own draft H3o and current plan describe the reference and unresolved complete
closed-chain domain proof. Prior observed boundaries/negative results remain;
active production, ordinary tests, CI, numeric kernel222/epoch8, schemas and data
are unchanged. No new game defect inferred from intentional wait-event refusal.
RP-680 and27 Clout AC6 failures remain OPEN. Logged stage/clock, actual clearing/
SQL/event delivery/idempotency, complete caller/global/lifecycle/domain, workload/
compatibility and research/normative adoption stay mandatory. Fixed-sampling
success is not permission to omit modern endpoint or full nine-tier1.0 scope.
Review by: Codex root (research author/first filter). Recorded by: Codex root.
Five-path range6fe8e3b4-exclusive through this checkpoint needs research/normative
review outside corrective-code delegation. Not designated approval, implementation
acceptance, archival, push/publication/deployment or release. Goal active/progress;
next source proof must cover all-resource Commons capacity, not just cash/lifetime.

## 2026-10-16 — closed-chain range proof, declared research population

Continue fromc61c4eb2. The next component must bound the Commons hook's actual
receipt arithmetic across ALL resources. A positive valid producer/ledger delta
can still underflow when multiplied by Compact tithe; cash safety proves nothing
about a secondary resource. Predeclare four complete0..1000 populations on a
rehash/reloaded modern19 axis fixture with secondary permits production and zero
initial permits: online rate1e(MIN+3), offline rate1e(MIN+4), and both modes at
1e(MIN+6). Default Compact tithe, real full closed chain, no pending clearing/live
resolver. All4004 no-hook producer points must succeed. Low-domain arms must have
an actual early Commons-capacity refusal and later success; exact endpoint counts
are measurements (the real axis factor is retained), not assumed1/19ms folklore.
Headroom companions must succeed; source admission/population failure is invalid,
not an exclusion. Check first actual error against unreachable cash queries.

Carry minimum positive exponents alongside existing maximums in the research
partial/prefix/aggregate certificates, without changing their arithmetic or guards.
Positive aggregation cannot lower its smallest positive input exponent. Receipt
reproduction bounds nonzero-before delta below by before.exponent-17; zero-before
delta copies after, bounded below by aggregate minimum or positive binding cap.
Upper receipt exponent is bounded by the nonnegative cap. Derive Commons Mul and
sequential positive Add/Q12 headroom from those bounds; inconclusive is not invalid.
Compose with the closed lifetime certificate, actual offline-span helper, stock
read-set certificate permitting PRIVATE saturation events, fixed Guild validation
and actual fixed Commons enclosure/compliance checks. No generic hook claim.

Bind the concrete guard to a globally certified producer; producer-inconclusive
ranges keep the prior actual chronological fallback. All original48016 closed
references/32 queries must retain full state/events/error equality; valid whole
ranges should now prune unreachable queries without omitting selected real hooks.
The missing-weight fault must still reach its real1ms error. Retain prior legacy/
stock/modern/first-hour populations for changed shared-source verification. A
compiling control removing only Commons' lower-exponent safety must fail the
new admitted-refusal oracle. No product/kernel/data/schema/CI or accepted contract
change; global timestamp/encoding/caller/workload/adoption gates remain open.

Initial4004-point capture421a9a/45231 FAIL2.221s: both hypothesized low arms had
ZERO hook refusals. Do not call this a successful negative or omit the population.
The subsequent first-point diagnostic2d6d0b keeps that refusal requirement unchanged.
Source inspection explains the missing witness: Decimal.Mul retains a FINITE
out-of-domain lower diagnostic; Commons checks final capacity, and a normal cash
term can absorb that tiny permits term through Add's17-decade gap. There is no
proved product defect in that acceptance. Prospectively isolate the actual
capacity boundary with a loader-admitted unminted cash cap1e6 and original cash
at that cap (so no cash receipt is emitted). Use rateMIN+3 in BOTH low arms:
the real1.15115 axis factor times offline.9 keeps1ms production valid, while the
current default tithe is.1, not the earlier hypothetical.05. Keep the same four
1001-point populations and headroom arms; no-hook producer success, actual closed
refusal then later success, and first-error query assertions remain mandatory.

Revised54c4c3/53000 PASS2.170s: all4004 producer points succeed; actual closed
refusals are1..8ms online and1..9ms offline (17 total), then later waits succeed.
Both headroom arms succeed. Guarded composition d23ca2/15511 PASS1.863s preserves
those first errors with one real leaf and certifies headroom with zero leaves.
Initial guard compilec614ef ran no tests (unused multiplier import; removed).
Before final verification add two declared2002-point headroom-prefix companions
(both modes): provision tick1000, offline cap1500, horizon2001, exercising complete
prefixes and the flat offline tail. Original4004 populations stay unchanged.
Require every actual receipt exponent inside the certificate and compare ALL8008
full closed states/evaluations/roles/events/errors with independent real
SimulateAdvance+event tap, not the projector's own output alone.

Broader nativef061bc/67524 FAIL80.805s solely in the old ownership budget fixture:
its unreachable1e12 query is NOW lawfully proved in one node, so budget1 is not
exhausted. All other selected populations pass. Do not weaken the budget or
force inefficient work: after the currently running Linux handle21072 terminates,
keep budget1 but use the existing750ms reachable quote so another node really is
needed. The shared proof/helper source stays frozen through that running check.
Lower-safety disconnection5a0b6d/28272 FAIL6.975s correctly: both admitted low arms
lose their actual1ms error as unreachable/nil; all four headroom/prefix companions
pass. Full8008 reference population retained. No production mutation.

Also add two declared1001-point VALID absorption companions after that handle
ends: same online/offline MIN+3 permits rate, but starting cash1e4 below cap1e6.
Real cash capacity absorbs the lower diagnostic, so actual full projection and
query must succeed despite an inconclusive sufficient proof. This retains the
initial negative research result as an executed supported case, not only prose.
Final complete population becomes241133 enumerated reference points (239131 above
plus2002 companions), with separate no-hook references/anchors/query controls.

Linux9b258e/21072 is now terminal FAIL245.088s for that SAME obsolete budget
fixture alone; all other selected populations pass. Only after terminal state,
change its quote to the already independently evidenced750ms target, keep budget1,
and add the two absorption controls as a separate manual source. No proof/helper
or product behavior changes for that fixture correction. The capture970abe/43276
PASS3.250s regenerated the identical H3n packet before its consumer ran; all normal
source bytes were frozen through those first full runs. Final full native/Linux
verification will include the corrected meaningful budget case and both added
supported controls, not conceal the earlier red runs.

Absorption control4f8433/10626 PASS3.012s: both complete1001-point populations
agree with the independent real primitive, including a finite out-of-domain
permits term at1ms which the real cash capacity absorbs. The sufficient guard
correctly remains inconclusive; query executes1000 positive leaves and succeeds
as unreachable/nil, with no leaked selection. It does not invent a new refusal.
Final native56c3a0/30068 PASS82.906s plus production vet; all241133 enumerated
reference points retained, with full closed states/events and separate no-hook
receipt references/anchors/query/ownership/refusal/budget controls. Sixteen-file
overlay and all manual source bytes frozen; no ordinary product source changed.

The new concrete bind only consumes globally certified per-resource producer
bounds. Complete successful producer prefixes/partials now carry BOTH positive
minimum and maximum exponents. Actual receipts from all8008 new source points
stay inside their bounds; the prefix and1500ms offline-tail fixtures exercise
nonempty and flat prefixes, not only a one-cell partial. This extends proof data,
not arithmetic or the earlier typed failure guards. Source review retains the
existing raw-bit cache key, copies min/max with each immutable prefix snapshot
and never treats a failed certificate as invalid product input.

In the ordinary closed profiles, the former751/2001 positive query projections
become ONE full leaf each; unreachable6002 becomes ZERO. Query proof still pays
for6002 scalar/partial checks (when production active), one extra contribution
assembly, one Commons catalog load when member, receipt bounds, two stock helper
probes and one span restore. This is a source-bound proof/work count, not a claim
that complete preparation is free or the full65M/300-report workload meets budget.
Private event acceptance is enabled ONLY in this closed research recipe; original
stock and live simulation event refusals remain unchanged. Every selected leaf
executes the actual full chain, including its stock-role counterfactuals.

## 2026-10-16 — RP-688 conditional closed-chain proof verified

Final declared Linux1f788e/29849 PASS283.216s on the SAME frozen sources; all
241133 enumerated reference points execute, including the corrected budget case,
new17 real Commons refusals and both valid absorption controls. Native result
above remains82.906s plus vet. They overlap and differ in compiler/architecture;
no timing or hosted/full-workload performance conclusion follows. No tests skip,
all owned handles terminal. The earlier native/Linux budget-fixture reds and
the two failed initial population attempts remain recorded, not overwritten.

Reproduction from root: run H3n's TestWaitFirstHourCaptureResearch command first
and await completion; then
`make test-go vet GO_PACKAGES='./production' GO_TEST_FLAGS='-count=1 -overlay=../planning/clout-v1-and-pr-interns/probes/wait-query-overlay.json -run TestWait -v'`.
Declared Linux: `docker compose -f compose.save-test.yml -f compose.ci-test.yml run --rm --no-deps test go test -p 1 -overlay=../planning/clout-v1-and-pr-interns/probes/wait-query-overlay.json -run TestWait -v ./production -count=1`.
Focused selectors: TestWaitClosedReceiptDomainResearch, TestWaitClosedAbsorbedLowerResearch,
TestWaitClosedReplayBindingResearch and TestWaitClosedReplayOwnershipResearch.
The final full selection retains183107 earlier legacy/stock/modern/first-hour
points +48016 closed-chain points +8008 capped/prefix points +2002 absorption
points. Separate producer, anchor, selected, ownership and invalid-input calls
are additional, not double-counted as new populations.

The compiling lower-safety fault retains the identical8008-point source-test body
and guard code, replacing only `if low+factor.Exponent() < -edge` with
`if false && low+factor.Exponent() < -edge`. Its15-file ignored overlay predates
the separate absorption source, which is not selected by that negative command.
No negative rerun or coverage of that extra test is claimed/needed: normal guard
math and the fault's selected population did not change. Root command uses
-overlay=../.cache/wait-closed-unsafe-lower.json -run TestWaitClosedReceiptDomainResearch.
Both real early refusals become unreachable/nil and fail, while all headroom/
prefix companions remain green; it is not compilation or admission discrimination.
Normal source was never mutated. Final SHA256:
- production-wait-cells:67af8acb2426bf9a7e6d6ac34ad962affb5d1967d12730e825a89e396e7f8f6b
- wait-query-order:c5e665b72235a3be0401fa51ccd4d02f4cd5b948ce058c0943e4d4a27284ba18
- wait-stock-hook-domain:b8efea99d9d4985325ba361c7d08734935c0a4a7eed2462edb126a893caefaa5
- wait-closed-domain:d8a6181a4a9dfb013f4e79db1c31658dfba90d0412901613a9fcbdc735a33e9f
- wait-closed-absorption:5fabdcfb7b91f9c20620c8e996f91ebb6bcdfe96c4c0f729b3bd82e4bf94f9f0
- wait-closed-replay:9fd6d81625c0426485d1a0b69e75dfb9f39c771a054706fc3f7a7a055fadbd38
- sixteen-file overlay:178711c548198b904aba29ee7da8152ce8909bba998c11c069910fdb5adae432
- ignored fault source:bc1242c40303f5b4ddc7296d72e4a2e248e516a5de26230c69186a05c1c56787
- ignored fault overlay:889c27f753e7e3d06520366d3e376eb02bc9d0478615598f4c1cf54d534c276e

RP-688 records an admitted semantic constraint on the RESEARCH certificate, not
a product defect or a proposed new runtime refusal. Own draft H3p/body/plan
reconcile the now-conditional proof with H3o's historical unproved reference.
The private guard's input ranges come only from its bound producer's cells;
this is not a standalone arbitrary-range admission API. Global pre-hook encoding,
timestamps/state/caller/identity/lifecycle, exact logged stage/clock, real clearing/
SQL/event delivery/idempotency, compatibility/full workloads and normative
review/adoption remain open. RP-680 and27 AC6 failures remain OPEN. No accepted
contract, kernel222/epoch8, product, client, active data, schema or CI bytes changed.
Review by: Codex root (research author/first filter). Recorded by: Codex root.
Complete eleven-path rangec61c4eb2-exclusive through this checkpoint needs research/
normative review outside corrective-code delegation. No designated approval,
implementation acceptance, archival, push/publication/deployment or release claim.
Full nine-tier1.0 goal remains active/progress; next complete global pre-hook
encoding/timestamp/state obligations before claiming the producer+hook query sound
over its whole supported context, then exact logged stages and real settlement.

## 2026-10-16 — pre-hook cursor/codec boundary research

Continue from5b758f50. Inspect the ACTUAL encoder and simulation hook, not merely
the catalog's arithmetic domain. Evaluate changes ledger, credits, burst,
provisioned/remainder maps and cursor. An installed simulationHook first encodes
that state, then may restore a stock-role counterfactual at the same source version.
UTC/millisecond syntax alone does not prove the formatted cursor is readable or
nonzero. Keep actual no-hook behavior distinct: it does not call that encoder.

Predeclare12 complete0..1501 populations: two clocks x online/offline x none,
first-hour lifetime, generator-masked lifetime. Upper clock uses the real v4 T0T1
catalog, admitted Company14 at9999-12-31T23:59:59Z with purchased stock-role beige1.
Lower clock uses the existing v3 phase0 catalog (no provision clock), admitted
Company14 at0000-12-31T23:59:59Z, RunStartedAt zero as its existing codec permits.
Both cross a boundary at1000ms: RFC3339 year10000 or Go's zero time. These are
seeded primitive-domain diagnostics, NOT natural player dates, a changed server
clock, active-play enablement or a deployed save-loss finding. Population/admission
failure invalidates the proposed instrument rather than exclude a row.

At all18024 points execute actual fresh SimulateAdvance; retain exact errors and
whole state/results/roles. Observe encode/restore failures SEPARATELY from actual
primitive errors: do not turn a successful no-hook wait into a fabricated error.
For complete comparison only, encode a shallow state view with its cursor reset
to the original cursor AND compare the true cursor independently; this does not
mutate the projected object or hide its actual codec observations. Query original
cash, independently sampled500/1250ms no-hook cash and1e12; all48 queries compare
against complete chronological actual-error/affordability references and full
selected state. Earlier valid results must survive a later error; a later valid
endpoint cannot erase an earlier error.

Proposed repair is manual research only: closed recipe owners declare when the
actual primitive requires pre-hook encoding. Source-derived canonical relative
time/codec range checks may certify within that owned interval, otherwise refine
to real endpoints. A no-hook caller keeps its actual behavior. Do not silently
reject dates or invent a runtime calendar limit. Preserve existing producer and
hook proof reuse; expose temporal checks separately. Initial hypothesis predicts
1008 actual errors (upper unmasked1004, lower hook cases4), but source execution
owns truth. Native/Linux affected checks and a compiling bypass of ONLY this new
temporal guard must discriminate before checkpoint. No production/codec/kernel/
schema/data/CI or accepted-contract edit; other global/caller/logged gates stay.

Baselinefeacf3/52701 FAIL1.498s exactly at the query boundary: all18024 actual
points execute with1008 real primitive errors; TEN queries hide the first1000ms
error or report it later at1250ms. Independent codec observations are six encode
refusals and3012 restore refusals, NOT3018 additional primitive failures. The
no-hook and upper masked companions remain successful under their actual contract.
Correctedebbf5a/66781 PASS1.539s, same population/48 queries, full selected state/
true cursor/evaluation/roles unchanged. Add thirteen direct temporal-certificate
controls, including valid endpoints straddling the zero sentinel, unaligned/
non-UTC origins, reversed and existing-instrument-range bounds. Then freeze all
sources for full affected native/Linux and the compiling temporal-bypass control.
Initial Linuxb29c84 ran ZERO tests: an overlay path typo used pr-interns without
its first hyphen. Correct only that command; do not call it a code failure or
silently count it as a passing architecture gate.

## 2026-10-16 — RP-689 temporal proof checkpoint

Native86e70f/7220 PASS83.233s plus production vet, all259157 enumerated reference
points (prior241133 +18024 temporal), plus independent producer/selected/full-state/
ownership/certificate controls. The thirteenth direct time control is included;
no test skips. Source helpers/context bindings frozen before all final runs.
Corrected Linux2d37a8/91607 remains the single live owned handle until its terminal
result; no restart because an observation wait yields.

Compiling ignored bypass8a8e65/92461 FAIL1.502s with all18024 points/1008 actual
errors/48 queries retained and the exact TEN first-error misses reproduced. Only
`timeDefined := !context.preHookEncoding` becomes `timeDefined := true`; numerical
producer proof, hook guard, actual primitive and ownership constructors remain
intact. Valid none/upper-masked populations still pass. This is semantic query
discrimination, not a compile failure or a counter-only assertion. Normal source
never mutated. No full producer/Codec rewrite or artificial date error is involved.

The complete source write-set audit is recorded in own draft H3q: successful
ledger certificates preserve finite canonical balances; validated bank/credit
caps and burst subtraction preserve integer bounds; actual provisioning writes
existing declared targets with safe hardcaps and modulo remainders; all other
pre-hook metadata is unchanged, with monotone cursor advance preserving its
existing order constraints. This argument REQUIRES a fully admitted RESTORED
origin/pinned catalog/mask. A valid initial ledger alone is insufficient. Do not
promote it to arbitrary-input, absent-key mask, post-hook codec, lifecycle, all-
caller or complete logged-stage proof. Those remain the next global obligations.

The temporal obligation is selected mechanically by the closed research recipe
owner, not accepted from an external callback claiming purity. No-hook waits keep
their actual lack of pre-hook encoding; installed lifetime/stock/closed hooks
pay for temporal range checks. QueryTimeGuards exposes this work without changing
existing research node limits or pretending this proves full workloads feasible.
The zero sentinel can lie inside two valid endpoints, so endpoint-only sampling
is expressly insufficient. A failed sufficient proof refines to actual source
results, retaining masked-year10000 successes instead of rejecting them by fiat.

Final declared Linux2d37a8/91607 PASS265.424s, all259157 enumerated reference
points and thirteen direct temporal controls, with unchanged ordinary populations.
Native83.233s/vet above and Linux share frozen source hashes; no skips. All owned
handles terminal. Compilers/architectures and concurrent load differ, so timings
are NOT performance or hosted-CI evidence. No production/full-harness/TS/SQL/
full-workload claim follows. The original unguarded and intentional-bypass reds
remain exact evidence; the failed Linux path lookup ran no tests.

Root reproduction: H3n's producer must already have regenerated its ignored
first-hour packet (its source/artifacts are unchanged since the prior wave), then
`make test-go vet GO_PACKAGES='./production' GO_TEST_FLAGS='-count=1 -overlay=../planning/clout-v1-and-pr-interns/probes/wait-query-overlay.json -run TestWait -v'`.
Linux: `docker compose -f compose.save-test.yml -f compose.ci-test.yml run --rm --no-deps test go test -p 1 -overlay=../planning/clout-v1-and-pr-interns/probes/wait-query-overlay.json -run TestWait -v ./production -count=1`.
Focused selectors: TestWaitStateTimeResearch and TestWaitStateTimeCertificateResearch.
Negative command uses -overlay=../.cache/wait-time-bypassed.json and
-run TestWaitStateTimeResearch; ignored copy changes only the global temporal
boolean in the shared query helper, so per-range checks are bypassed too. All
actual reference populations, owner-selected recipe flags and other proof stages
stay intact. SHA256:
- wait-state-time:8ad44ce51211e78387705610cecd2fdf2d47aae7ae6222c2c498d18b7f01becc
- production-wait-cells:858bce15a64237ca0f60f2d61b6d9a4ad3690f5ba2b240a1a74447f86eda7169
- wait-query-order:e64d16db988e2346c49c8ff6ef4e9469a946ecb431377990bbfb6283310f9321
- wait-modern-binding:a0b4d1566a68f7b7755474db4c5a5aac68b343d423234b351d59e181235b48c5
- wait-first-hour-lifetime:7ac0a2ac2db6db6548ca3a18225d2fc3bb4d2feefaaa0ace1f77f14c05dd890e
- wait-closed-replay:748da29d33af72f602136d688dbde80fd2fa27f81c0b34db014e4ae7fe91ebb7
- seventeen-file overlay:a2e901bf8ea4937f8959ddc633cc96a75775808dff988c8744cacee850137853
- ignored bypass source:ea062056e4349d7f056a5adae24bcb26a14a868cdd53e52c929cebe6df6b664b
- ignored bypass overlay:0ff2112692180f7b5d8bc18bf3a8f1344179ca5ac4999a3810afa305fb3b4c70

RP-689/own H3q/own plan preserve this bounded result and its input/state limits.
Source-version, normal producer/hook error behavior, kernel222/epoch8, live
simulation/codecs, active data/schema and normal CI remain unchanged. No runtime
calendar restriction, invented error class, generated origin wire format, extra
saved payments or author-content change. General input/mask/map membership and
post-hook state preservation, lifecycle/identity, exact logged stages/clock, real
settlement/SQL/event delivery/idempotency, full workload/compatibility and normative
adoption remain mandatory. RP-680 and27 AC6 failures stay OPEN.
Review by: Codex root (research author/first filter). Recorded by: Codex root.
Eleven-path range5b758f50-exclusive through this checkpoint remains outside
corrective-code delegation and needs research/normative review. No designated
approval, feature acceptance, archival, push/publication/deployment or release.
Full nine-tier1.0 goal active/progress. Next inspect remaining input/mask/state
preservation, rather than assume unchanged metadata from a cash-only predicate.

## 2026-10-16 — RP-690 simulation mask restoration, accepted-contract correction

Resume923c694a. Purchasable Content Foundation P5/C10 requires output masking with
units/costs preserved; Relevance R10 uses that same real hook-bearing wait. This
is a routine conformance repair under those ACCEPTED contracts, not permission
to implement H3/C1 drafts. The current helper restores absent IDs as present0.
An unminted loader-admitted Founder generator is a valid catalog-level mask but
absent from the codec-admitted Company. Real faction-hook waits leave it present,
then fail the actual encoder (purchased/provisioned map shape mismatch).

Old9e20d1/18866 FAIL0.342s: six cases fail (two absent-key unit success/error plus
four actual SimulateAdvance online/offline success/error), while four existing-
zero/nonzero companions pass. Actual catalog/mask/Company codec admission runs
first. The failure-arm policy resolver is deliberately unavailable; expected
post-evaluation scratch effects are compared to the SAME unmasked failure, never
misrepresented as persisted mutation or a transaction atomicity defect.

Preserve original value AND presence; do not insert an absent masked ID during
the hook, and delete any such ID the hook writes on return. Existing purchased
counts restore on both success/error, while real hook effects/events/errors pass
through. No new mask rejection, scope restriction, formula or hook recipe. The
watched production behavior changes, so honestly bump all three kernel identities
222→223 in the same batch (no history rewrite, false identity bump or baseline
regeneration). Run focused old-failing regressions, affected production/harness/
kernel/numeric checks and declared Linux; keep known Clout/guard failures explicit.
Canonical harness docs are updated; independent corrective review is delegated
to a fresh Codex agent after the complete implementation range is committed.

Implementer verification (not designated approval): corrected focused simulation
and version-source tests PASS (production0.435s/kernel0.170s); production vet PASS.
Full native decimal/economy/kernel/harness tests and vet PASS, including harness
85.703s. Full client suite PASS:145 files/12494 tests,23 files/1095 tests explicitly
skipped; typecheck/svelte-check report zero errors/warnings. Declared Linux/AMD64
Go1.26 `TestSimulation` population PASS0.428s, including all ten new cases.
The full native production suite remains RED51.685s solely on the27 existing
Clout AC6 partition subcases; no new failure population appears. The full kernel
history guard remains RED at immutable50a3a514 (missing historical version bump),
after checkout/adversarial history checks pass. Neither red is waived or repaired
by this change. Current three-file version identity is223. Full declared Linux
decimal/economy/kernel/harness execution is still running at this checkpoint;
its result must be recorded separately, not inferred from native success.
No browser journey, Postgres integration, exhaustive pacing, hosted measurement
or manual research population was claimed from these focused runs. Existing
ordinary simulation tests exercise action paths; the new real-hook regressions
exercise the affected wait and codec boundary with success/error controls.
Review by: Codex root (implementer first filter). Recorded by: Codex root.
Exact implementation range starts after923c694a; fresh independent corrective
review remains required. RP-680,27 AC6 failures and unaccepted research contracts
remain open; no feature archival, release or publication follows.

## 2026-10-16 — RP-690 independent corrective review: APPROVED

Review by: independent Codex subagent `/root/simulation_mask_corrective_review`.
Recorded by: the same reviewing subagent. Authority: Marco's2026-10-15 corrective-
code delegation, not cross-model review or new-feature/specification adoption.
Exact complete reviewed range: `923c694a..7ec0873c` (nine paths).

Independently inspected the full range and accepted Purchasable Content P5/C10
and Relevance R10. The helper now restores original membership and values without
narrowing catalog-level mask admission; actual hook effects/events/errors remain
forwarded. Canonical docs describe that limited correction and the honest watched-
behavior kernel bump; all three identities independently checked at0.3.223.

Executed `make test-go vet GO_PACKAGES='./production ./kernel'
GO_TEST_FLAGS='-count=1 -run TestSimulation\|TestVersionSourceOfTruth -v'`:
PASS production0.424s/kernel0.155s, then vet PASS. This includes all ten new
subcases, existing action/output/role tests and the simulation source-boundary
guard. The real-hook cases actually admit the mixed-scope catalog and Company
codec, execute online/offline success and stock-policy-error paths, compare the
complete encoded state against the corresponding unmasked control, and restore
the result. Unknown-mask refusal remains unchanged.

Independently recreated the old helper ONLY in an ignored source copy, using
Go `-overlay` through root Make. Its SHA256 matched
`git show 923c694a:server/production/simulation.go` exactly:
`bbde8606cd3cc7517c02da137af93ac0aca0650c84a4964524563ffad9fc29c2`.
With the two new regression test names and `-count=1`, old-source execution FAILS
0.283s: both absent-key unit cases plus all four actual wait/codec cases fail;
the four present-zero/nonzero controls pass. The failure is the real phantom
Founder count and real encoder refusal, not a fixture admission failure. No
tracked source was mutated. `git diff --check` passes.

APPROVED for this correction only. No independent rerun/approval of whole Clout,
Postgres, browser, exhaustive research, hosted CI or release is implied. Existing
27 AC6 failures, the historical50a3a514 guard failure and pending H3/C1 research
adoption remain open. The implementer's separate full Linux suite result is not
an independently rerun claim of this review. No archival or publication.

### RP-690 terminal Linux result and closeout

The previously running declared Linux/AMD64 Go1.26 full decimal/economy/kernel/
harness command terminates PASS:0.661s/8.563s/0.028s/196.325s respectively; Docker
wait confirms actual container exit0. This is implementer execution, not another
independent-review claim. Independent verdict190a4790 covers the complete
923c694a..7ec0873c correction. RP-690 is closed; kernel223/epoch8, no authored
data or accepted mechanic changes. Current-state/roadmap pointers now name this
bounded correction rather than222; no whole-product milestone advances.
Next: remaining admitted-input/mask/state obligations and actual logged stages
for the proposed accounting/search route. RP-680,27 AC6 partition failures,
historical version guard and research/normative adoption remain open. Full1.0
goal remains active; no push, release, archival or deployment.
Review by: Codex root (execution recorder/closeout, not designated reviewer).
Recorded by: Codex root.
