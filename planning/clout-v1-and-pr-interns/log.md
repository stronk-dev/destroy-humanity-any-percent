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
