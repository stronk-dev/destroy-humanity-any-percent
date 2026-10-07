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
