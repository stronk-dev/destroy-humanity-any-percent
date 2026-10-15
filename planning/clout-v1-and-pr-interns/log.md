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
