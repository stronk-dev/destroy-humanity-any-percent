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
