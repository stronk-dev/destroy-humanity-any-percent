# Reputation Tree v1 — implementation log (append-only)

## 2026-09-25 — Predeclaration B1–B2 (Claude)

**Implemented by:** Claude. Awaiting Codex designated cross-party review for every range below;
nothing here is self-approved.

B1 predeclares: a strict `reputation_tree` loader in Go and TS enforcing R2 rules 1–8 (rule 8's
copy-key registration is enforced against the declared copy-key set, as Soul/Pitch do; the
`copy/references.v1.json` registration lands with the production artifact at mint), accounting
helpers (available = level − spent; derived unlock ppm), and the R3 factor
`1 + level×per_level_ppm×unlock_ppm / 1e12`. One Go-authored corpus drives both runtimes:
`testdata/reputation/tree-fixtures-v1.json` (one rejection fixture per rule) and
`testdata/reputation/bonus-vectors-v1.json` (AC5 vectors). Failing cases to demonstrate: deleting
each rule's check makes its fixture load; computing from `available` instead of `level` fails a
vector.

B2 predeclares the bundle wiring described in the plan. The fixture tree (the RFC's proposed R2
table) lives in `balance/testdata/reputation-tree/fixture-v1.json`; no production artifact or
epoch is created.

## 2026-09-25 — B1 landed (Claude)

**Implemented by:** Claude; awaiting Codex designated review.

- **Delivered:**
  - `server/reputation/tree.go` and `client/src/reputation.ts`, both registered as kernel-guarded
    paths. The kernel version bump is in the same commit.
  - `curriculum.ValidateStarter`, exported as a pure wrapper so the tree reuses the exact starter
    grammar.
  - The fixture tree, the 24-case rejection corpus and the 10 Go-authored bonus vectors.
  - Placeholder node copy.
  - `docs/reputation-tree.md`.
- **Evidence, all runs cold:** `go test -count=1 ./reputation ./curriculum ./kernel` passes, and
  `vitest run test/reputation.test.ts` passes 4/4.
- **Severing probes:** each breaks one check, and each turned its test red; every file was then
  restored.
  - Go: rule 1 (whole declaration, provider only), rule 2, rule 3, rule 4, rule 5, rule 6
    (monotonic, final rung), rule 7 (whole headroom, resource branch, generator branch, duplicate
    upgrade), rule 8, and the AC5 mutant that computes from available instead of level.
  - TS: rules 1–8 and the AC5 mutant.
- **Finding during probing:** the TS rule-1 provider-only mutant first survived, because no fixture
  bound a declared row that had the wrong provider. I added `bonus_source_wrong_provider`, which
  uses `fiscal.hoard`; both the Go and TS provider mutants now fail.
- **DESIGN-GAP RT-DG-A:** R2 rule 8 names `copy/references.v1.json` registration. The copy
  reference registry is keyed to epoch artifact schemas, and this artifact has no epoch or schema
  yet. The loader enforces the declared-copy-key half now. Registration, and
  `balance/reputation-tree.schema.json`, land with the production artifact at the mint (B-mint),
  and the gap is recorded here rather than improvised.

## 2026-09-25 — B2 landed (Claude)

**Implemented by:** Claude; awaiting Codex designated review.

**What landed:**
- The Go and TS bundle loaders accept `reputation_tree`, which requires the `minigame_api` chain.
  The pairing rule applies in the loaders and in Go's `valid()`.
- The frozen-contribution resolver accepts provider `reputation_tree`.
- Tests: `server/replaycatalog/reputation_test.go` and
  `server/production/reputation_contributions_test.go` (both against the live epoch-8 artifact
  set), plus `client/test/reputation-bundle.test.ts`.
- Kernel version 0.3.108 → 0.3.109, bumped in the same commit.

**Evidence (cold):** `go test -count=1` passes for `./reputation`, `./replaycatalog`,
`./production`, `./curriculum` and `./economy`. `make typecheck test-client` passes (6702 tests).

**Severing probes.** Each check below was disabled and turned its test red, then was restored:
- loader pairing and loader chain, in Go and in TS;
- the `valid()` pairing check, in Go;
- the resolver's provider acceptance, in Go.

**Probe corrections:**
- The first resolver sever failed only because the build broke (an unused import). I redid it as a
  mutant that compiles and passes vet, and it still turned the test red.
- My first `valid()` pairing test was vacuous: the hash/count mismatch masked it. I replaced it with
  an isolated case. A live bundle stays valid until only its Economy is swapped for one that
  declares the source.

**Deliberately not done:** B2 raises no Founder version floor, because v22 lands in B3. No epoch
pins a tree, so this is not a live hole.

**Unrelated finding:** `server/replaycatalog/catalog_test.go` at HEAD (committed in `391beb76`) is
not gofmt-clean. Its owner or review should fix it; I left it untouched.

## 2026-09-25 — B3 landed (Claude)

**Implemented by:** Claude. Awaiting Codex designated review.

**What landed:**
- **Founder save v22.** The next free version: HEAD's `LatestFounderVersion` was 21, and no
  other stream claimed 22.
  - Go codec: `server/save/state.go`.
  - Floor, activation and pinned-tree validation: `server/production/replay.go`,
    `foundations.go` and `founder_replay.go`.
  - TS restore/encode/activation: `client/src/replay.ts`.
- **Tests:**
  - `TestFounderV22ReputationTreeRoundTripAndInvariants` in `save/state_test.go`.
  - `production/reputation_activation_test.go`: floor 22; run-boundary activation through
    `settleAndActivateFoundations` (epoch 5 → epoch 8 → tree); the replay activation arm; the
    mirror and accounting validation; the carry failing closed.
  - `client/test/reputation-founder-state.test.ts`.
- **Kernel version:** 0.3.109 → 0.3.110, in the same commit.

**Evidence (cold):**
- `go test -count=1` passes for save, production, replaycatalog, reputation, releasepackage,
  deploymentbackup, gameui and account.
- Postgres integration passes for `./production`.
- `./save` integration passes when run alone.
  - My first combined run failed `TestStoreIntegrationRevisionLifecycle` (outbox claims) while
    another agent's `test-run` container shared the Postgres service.
  - After that container exited, `./save` integration passed alone. I'm logging this as
    interference, not as a pass of the combined run.
- `make typecheck test-client` passes 6702 tests; the new TS test passes 2/2.

**Severing probes:**
- Go codec: 3 of 4 turned red (spent over level, the pre-v22 unlock mirror, the required-field
  check). The one survivor is explained below.
- Go production: 6 of 6 turned red (floor, settle initialization, mirror, accounting, the replay
  pre-activation unlock, the carry failing closed).
- TS: 4 of 5 turned red (accounting, mirror, pre-v22 mirror, encoding the mirror). The one
  survivor is explained below.
- **The two survivors are redundant layers, not vacuous checks.** Removing the load-side
  sortedness check alone (Go `sortedUniqueMechanicalSlice` on restore; TS
  `sortedUniqueMechanical`) still rejects, because `validateFoundationState` / the tree's
  `UnlockPPM` independently enforce sortedness. Each invariant is enforced; the load-side layer is
  defense in depth.

**Probe corrections:**
- `TestPinnedTreeValidatesTheUnlockMirrorAndAccounting` did not match the probe `-run` filter, so
  my first mirror and accounting probes ran no test for them. The mirror "FAIL" was also a compile
  failure.
- I renamed the test to `TestReputation…` and re-ran both as compiling mutants. Both turned red.

**DESIGN-GAPs:**
- **RT-DG-B.** R7 names cases in `testdata/save-migrations.json`. That corpus is the v1–v9 codec
  corpus, restoring and expecting a v5-shaped result; it has no Founder v15+ arm. Founder v17–v21
  were witnessed by codec and activation unit tests, and v22 follows that precedent:
  - `founder-v21-to-v22` is the activation tests;
  - `founder-v22-spent-over-level` and `founder-v21-nonzero-unlock` are codec cases;
  - `founder-v22-unlock-mismatch` is the pinned-tree validation.

  The baseline manifest ratchet is therefore unchanged.
- **RT-DG-C.** The Founder carry in replay inputs has no tree fields yet. A v22 carry fails closed,
  in Go and TS, until B5/B6 add them with the next replay-inputs version.

**Not yet witnessed:** the TS Founder-log Exit activation arm (`resultVersion >= 22`) has no TS
test. It needs a TS Exit-replay fixture with a tree bundle, which lands with the B5/B6 corpus. B3's
plan box therefore stays unchecked until then.

## 2026-09-25 — B4 landed (Claude)

**Implemented by:** Claude; awaiting Codex designated review.

**What landed:**
- The R5 intent: `server/production/reputation_intent.go`, with `reputation.Tree.Purchase` as the
  pure evaluator.
- The replay arm in `founder_replay.go`.
- Intent parse and dispatch in `intents.go`.
- The event kind and strict payload validator in `save/intent.go`.
- Migration `00075_reputation_node_purchased.sql` (next free number; the event-kind constraint).
- The TS parse, `reputationPurchase`, and the `applyFounderReputationPurchase` replay arm.
- Tests:
  - the corpus test, the rejection table, and the replay-tamper test in `production/reputation_purchase_test.go`;
  - the Postgres witness in `production/reputation_purchase_integration_test.go`;
  - `client/test/reputation-replay.test.ts`.
- Kernel version 0.3.110 → 0.3.111.

**Evidence, all run cold:**
- Go tests pass for reputation, production, save, replaycatalog, gameui, account and gameserver.
- TS: `reputation-replay.test.ts` passes 22/22 (every corpus case byte-matches Go on receipt,
  events and post-state), and 6726 client tests pass.
- Postgres: `TestReputationPurchaseIntegrationRecordsReplayableFounderLog` passes. It covers an
  applied purchase, an idempotent retry that returns identical bytes with `Replay`, the owned
  rejection as a recorded founder_log row, one persisted event, a `VerifyFounderHistory` result of
  `verified`, and the persisted Founder state.

**Severing probes.** Each mutant compiled and was restored after the run.

| Area | Mutant | Result |
|---|---|---|
| Go | requires check | red |
| Go | owned check | red |
| Go | inactive arm | red |
| Go | owned_before comparison | red |
| Go | unaffordable at `available+1` | survived the first corpus (see finding 1), then red |
| TS | cost off by one | red, 11 cases |
| TS | requires check | red |
| TS | receipt literal | red |
| TS | inactive detail | red |
| TS | owned_before comparison | red |
| TS | unaffordable at `available+1` | red |
| Postgres | event validator narrowed so a valid `direct` event is rejected | red, so the validator is on the real persistence path |

**Findings:**
1. The unaffordable boundary mutant survived the first corpus, because no row had
   `cost == available + 1`. I added `rejects-cost-one-over-available` (level 1, cost 2).
2. **The migration probe is invalid.** Removing the kind from the uncommitted 00075 still passed
   because the shared test Postgres had already recorded goose version 75, so the edited file was
   never re-applied. I did not restart the shared service while other agents were using it.
   Positive evidence only: the event persisted under the migrated database, and the pre-00075
   constraint (00072) does not list the kind.

**RFC notes:**
- R7 item 3, extending the founder_log resolved-arm check, is N/A: no such database constraint
  exists (only the events kind and schema-version checks do).
- R5 step 2's Fiscal sweep runs through the existing ApplyFounderLogged prelude and decorator.

## 2026-09-25 — B5a landed: frozen bonus row and pin completeness (Claude)

**Implemented by:** Claude; awaiting Codex designated review.

**What landed**
- `FrozenFounderContributions`, wired into the Exit path and `FounderInitializer`.
- Migration `00076_reputation_frozen_contribution.sql`: `CREATE OR REPLACE` of the completeness
  function, with Down restoring the 00065 body.
- `TestReputationFrozenFounderContributionsAddTheBonusRow`.
- The Postgres test now also covers AC6 and AC7:
  - AC6: a pin missing the reputation row and a pin with an extra row both fail with the new
    message, which proves 00076 is live on the test database; the complete pin commits.
  - AC7: a purchase leaves `FrozenContributionProvider` output for the pinned run byte-identical,
    while `FrozenFounderContributions` for the next run changes.
- Kernel version bumped 0.3.111 → 0.3.112.

**Evidence (cold)**
- Go tests pass for production, save and reputation.
- The Postgres test passes.

**Severing probes (both turned the tests red; restored afterwards)**
- Omitting the reputation row.
- Computing the factor from available instead of level.

**Not yet done**
- **AC7's failing case** ("a live-Founder-read implementation changes the current projection") is
  not demonstrated as a mutant, because no production path reads Founder state for a run's
  contributions. The positive byte-identity check is recorded instead.
- **B5b remains:** the R4 starter application at new-run assembly (a replay-inputs carry
  extension, RT-DG-C), `run_started` v2, and the TS side of the frozen row.

## 2026-09-25 — B5b landed: starters, run_started v2, replay-inputs v9 (Claude)

**Implemented by:** Claude; awaiting Codex designated review.

**What landed:**
- Go:
  - `applyReputationStarters` in `prestige.go` (R4 step 4) and the run_started v2 emit;
  - `reputation.Tree.OwnedStarters`;
  - the replay-inputs v9 carry extension in `replay.go` (closes RT-DG-C);
  - the run_started v2 payload/schema validation in `save/intent.go`;
  - migration `00077_run_started_v2.sql`.
- TS (`client/src/replay.ts`): v9 accepted, carry parse/advance with the tree fields,
  `applyReputationStarters`, and the run_started v2 emit.
- The shared corpus `testdata/replay/apply-logged-v1.json` was regenerated with `make
  replay-fixture`. The diff is only the `v` field (8 → 9); no receipt, event or state byte moved.
- Kernel version 0.3.112 → 0.3.113.

**Process finding (mine):** `releasepackage.TestCurrentMigrationIsContiguous` pins the latest
migration number (74).
- My committed `541da96e` (00075) and `7ea6b942` (00076) broke that pin without my noticing,
  because I ran `./releasepackage` only before those migrations existed.
- The coordinator flagged it. This commit updates the pin to 77, the contiguous head.
- Other `DatabaseMigration: 74` literals are fixture values, not pins, so they were left as-is.

**Evidence (cold):**
- Go: `go test -count=1` passes for save, production, replaycatalog, reputation, gameui,
  account, gameserver and harness (harness ran the full 30 s).
- Go: releasepackage, deploymentbackup, deploymentrelease and deploymentrehearsal pass.
- Postgres integration passes for `./save` and `./production`, with 00077 applied.
- Client: `make typecheck test-client` passes 6726 tests; existing replay parity is intact under
  v9.

**Not yet witnessed:**
- Go and TS starter application across a real Exit with a tree bundle.
- run_started v2 Go/TS byte parity.
- AC8, the burnout curriculum plus `generated_beige_tower` totalling 15 generated.

These need an Exit cross-runtime case on a tree bundle; that case is the next commit. B5's plan box
stays open until then.

## 2026-09-25 — B5 witness: AC8 Exit cross-runtime case (Claude)

**Implemented by:** Claude. This awaits Codex designated review.

**What landed.** The reputation corpus gains `exit`. It is a scripted-first burnout Exit on a tree
bundle whose next epoch pins `curriculum-v2`. The Founder is at v22, level 6, spent 6, and owns
`p05`, `cash_small` and `generated_beige_tower`.

**Go assertions:**
- The burnout curriculum grants 10 generated beige towers; the `generated_beige_tower` node adds 5,
  for exactly 15 generated and 0 purchased.
- `cash_small` grants cash, which stands at `1e3`.
- `run_started` is emitted at schema 2, listing the two applied starters in tree order, with a
  `bonus_factor` of `1.003e0`. That value is 6 × 1% × 5%. My first hand-computed expectation was
  wrong, and the test caught it.

**TS replay.** `test/reputation-replay.test.ts` byte-matches the Go receipt, the Founder output,
the new Company, and the started events. The file passes 23/23.

**Severing probes.** Each compiled and was restored afterwards.
- Assignment instead of addition (the RFC's named AC8 failing case) fails in Go and TS.
- Dropping the v2 emit fails in Go.
- Computing the TS factor from available instead of level fails.

**Kernel version:** no bump. Only `_test.go` and testdata change, which the guard exempts.

**Still open:**
- The TS Founder-log Exit activation arm (`resultVersion >= 22` in `applyFounderExit`, from B3)
  has no TS witness yet.
- R6, the Exit-attached plan, is the next batch (B6).

## 2026-09-25 — B6 (Go) landed: Exit-attached plan (Claude)

**Implemented by:** Claude; awaiting Codex designated review.

**C2 ruling check.** `/api/v1/intents` is not an operation in the generated API registry
(`docs/generated/api.json` has no such path), so the optional key widens no v1 request union.
There is no conflict, and no compatibility pin was touched.

**What landed:**
- `parseReputationPlan` in `intents.go`.
- Pre-validation and application in `finishExitResolved`; `reputationPlanRejection` and
  `applyReputationPlan` in `reputation_intent.go`.
- The `exit.v2` audit arm, derived from the decision's `exit_plan` events, in
  `founder_exit_replay.go`.
- The Founder replay `exit.v2` arm, which re-derives and compares purchases.
- History linking for `exit.v2`.
- Migration `00078_founder_exit_v2.sql`.
- Migration pin 77 → 78.
- Kernel 0.3.113 → 0.3.114.

**Corrections to my own earlier log:**
- **B4's "R7 item 3 is N/A" was wrong.** The founder_log arm constraint exists: `00057`/`00062`/
  `00069` `founder_log_multistream_source_shape`. I missed it by grepping for the wrong identifier.
  Migration 00078 extends that constraint.
- **Real defect found by the live witness:** `applyFounderReplayOutput` (the live Exit parity
  expected state) did not copy the v22 tree fields. Any live Exit that carried a plan or activated
  v22 would have failed parity. It is fixed, and this also covers B3's live activation path.

**Evidence:**
- **AC9 (`TestReputationExitPlanIntegration`), real Postgres through `Service.Handle` with
  `wind_down`:**
  - A plan whose last entry is unaffordable is rejected as
    `unaffordable / reputation_plan.reputation`. Both streams stay unchanged: same run, spent 0, no
    owned nodes, one exit record.
  - A valid plan (with an in-plan prerequisite ordering) applies:
    - run 3 starts with `generated_beige_tower` applied (5 provisioned);
    - spent is 6 and unlock ppm is 50,000;
    - three `exit_plan` events are persisted;
    - the next run's frozen bonus is `1.003e0`;
    - the Founder log arms read `[exit.v1 exit.v2]`, which is live proof that 00078 was applied,
      since the insert would otherwise fail its constraint;
    - `VerifyFounderHistory` returns `verified`.
- **Severing probes (Postgres), both turned red and were restored:**
  - removing pre-validation;
  - dropping the plan events from Founder replay.
- **Cold runs:** Go tests pass for save, production, replaycatalog, reputation, gameui, account,
  gameserver and releasepackage. The client passes 6727 tests.

**Not yet done:**
- The TS side of R6: the Company-log Exit with a plan, and the Founder `exit.v2` arm.
- A TS cross-runtime plan case.

These are the next commit.

## 2026-09-25 — B6 (TS) and B3 witness landed (Claude)

**Implemented by:** Claude; awaiting Codex designated review.

**What landed (TS):**
- `takeReputationPlan`, `reputationPlanRejection` and `applyReputationPlan` in `client/src/replay.ts`.
- The Company-log Exit now pre-validates the plan before finishing (rejection path) and applies it
  after v22 activation.
- The Founder arm accepts `exit.v2`, re-derives the plan's purchases, and compares them against the
  recorded ones.
- **B3 defect fixed:** `applyFounderExit` capped `result_founder_wire_version` at 21, so the TS v22
  activation branch was unreachable. The cap is now 22.

**Kernel:** 0.3.114 → 0.3.115, bumped in this commit (`client/src/replay.ts` is guarded).

**Corpus (Go-authored):** `exit_cases`, each a Company-log Exit plus its Founder arm:
- a plan with an in-plan prerequisite (`exit.v2`);
- an unaffordable plan (whole Exit rejected);
- a plan with no tree in the next bundle (`not_eligible / reputation_plan.tree_inactive`);
- an Exit activating Founder v22 (`exit.v1`, epoch 8 → tree bundle; the B3 witness);
- an activating Exit that also applies a plan (`exit.v2`).

**Evidence (cold):**
- `client/test/reputation-replay.test.ts`: 29/29, byte-matching Go on:
  - Company receipt, Founder output, new Company and events;
  - Founder receipt, events and post-state.
- Tamper tests:
  - TS: altered `exit.v2` purchases are refused.
  - Go (`TestReputationExitV2RefusesTamperedPurchases`): altered purchases, and a planned Exit
    downgraded to `exit.v1`, are both refused.
- Go tests pass for production, save, replaycatalog and reputation; the client passes 6733 tests.

**Severing probes (TS):**
- Red, as expected:
  - the B3 cap put back to 21 (3 cases fail);
  - pre-validation removed;
  - Company plan events dropped;
  - Founder unlock mirror not updated.
- Survived, redundant:
  - the activation `owned = []` initialization. A v21 restore already yields `[]`.

**Severing probe (Go):**
- Red: the exit.v2 comparison checking node id only, not cost.

**Plan boxes:** B3 and B6 flip in the following planning commit. Their tests are in this range.

## 2026-09-25 — B7a landed: server-derived `features.reputation` arm (Claude)

**Implemented by:** Claude. Awaiting Codex designated review.

**Snapshot versioning decision (per the coordinator's instruction).** R9 predates the Garage
lane's v4. Adding a new *required* `reputation` property to the pinned v4 `GameUIFeatures` fails
the API compatibility gate. I tried it: `gen-api` rejected it with "schema BootstrapResponse:
invalid API schema", because response mode forbids new required properties. That rejection is this
gate's demonstrated failing case.

The lawful additive route is an **optional** v4 key, `features.reputation: GameUIReputationArm |
null`:
- the server always emits it;
- the compatibility pin (`docs/generated/api-compat-v1.json`) is unchanged and was **not**
  re-pinned;
- no v5 is needed.

**Deviations from R9, recorded for the RFC author:**
- The block lives in the `features` arm pattern rather than at top level.
- The fact is `feature.reputation_tree`, following the v4 `feature.*` convention, instead of
  `founder.reputation_tree_active`.
- `tree_active` is implied by the arm being non-null.

**What landed:**
- **Server:** `server/gameui/reputation.go` (`projectReputation`):
  - fields: level, spent, available and unlock ppm;
  - `bonus_factor_this_run`, read from the run's pinned contributions, or null;
  - `bonus_factor_next_run`;
  - per-node state (owned, available, locked or unaffordable), derived on the server.
- **Schema:** `GameUIReputationArm` and `GameUIReputationNode` in `account/game_ui_api_schema.go`.
- **Generated files:** regenerated.
- **Client:** the TS v4 parser accepts and strictly validates the optional arm: `available` must
  equal `level - spent`, and every `requires` entry must name an earlier node.

**Evidence:**
- `TestReputationArmIsServerDerived` covers every node state, a null arm for a tree-less bundle or a
  pre-v22 Founder, and rejects an overspent Founder.
- Go `./gameui`, `./account` and `./gameserver` pass cold; the client passes 6733 tests.

**Severing probes:**
- Removing the `locked` requires-check fails the test.
- The unaffordable boundary mutant (`available+1`) at first survived, because no row had
  `cost == available + 1`. I added that boundary row, and the mutant now fails.

## 2026-09-25 — B7b landed: Reputation tree surface and Exit plan panel (Claude)

**Implemented by:** Claude; awaiting Codex designated review.

**What landed**
- `client/src/game-ui/ReputationTreeSurface.svelte`
- `client/src/game-ui/ReputationPlanPanel.svelte`
- Game UI wiring:
  - surface row `reputation_tree`, unlocked by fact `feature.reputation_tree`;
  - a nav tab, and the surface closes when the arm goes null;
  - the Founder-scope purchase intent, a rejection map, and a snapshot refresh after an applied
    purchase;
  - `withPlan` for `wind_down` and `accept_exit_offer`, which omits the key when the plan is empty;
    the plan resets on run continuation.
- Copy: the R9 key list in `copy/catalog/reputation-candidate.json`, every entry marked
  `PENDING OWNER COPY` (OD-11).
- Docs: `docs/game-ui.md`.
- Nothing here touches a kernel-guarded path.

**Evidence**
- `test/reputation-surface-browser.test.ts` passes 9/9 across Chromium, Firefox and WebKit. It
  covers:
  - axe checks;
  - no mechanical ids rendered;
  - server states rendered as text;
  - only an available node has a control;
  - Buy → Confirm (focus moves) → Escape cancels (group gone, focus back on Buy);
  - Confirm purchases;
  - controls disabled when Founder controls are unavailable;
  - the plan panel keeps tree order, gates on prerequisite and budget, and cascades deselection.
- Full lanes pass:
  - `make build-client test-client test-browser verify-client-boundary copy-check`, including
    20340 browser tests;
  - `make test-game-ui-composed`, both lanes.

**Severing probes (chromium)**

| Mutant | Result |
|---|---|
| Buy without confirm | red |
| Button shown on locked nodes | red |
| Plan budget gate removed | red |
| Cascading deselection removed | red |
| Escape handler removed | first survived, then red |

The Escape mutant first survived because my assertion was vacuous: it checked focus on "the first
button", which was the still-open Confirm. I tightened it to require that the confirm group is gone
and only Buy remains; it now fails as it should.

**Not done**
- The composed real-server purchase through the UI. No composed epoch pins a tree, so it cannot be
  reached honestly until a mint. This is recorded as a gap, not a pass.
- The Run End `[NEW ROUTE]` rendering of `run_started` v2. The Game UI event decoder does not
  surface `run_started` yet (DESIGN-GAP RT-DG-E).

## 2026-09-25 — B8a landed: H1, H2, H3 and the OD-2 threshold measurement (Claude)

**Implemented by:** Claude; awaiting Codex designated review. `server/harness` is not a
kernel-guarded path.

**Harness fidelity finding, fixed here.** The first-hour runner advanced production without the
served prestige accrual hook. As a result `company.LifetimeValue` never grew, and every simulated
Exit computed Reputation from lifetime 0. The first H1 run recorded `lifetime_value: 0` in all 97
runs, which is exactly the "looks like data" failure the evidence rules forbid.

The fix is `firstHourLifetimeHook`, which calls the served `prestigecore.AccumulateLifetimeValue`.
It deliberately omits the served hook's offline-span bookkeeping, because the runner records
session gaps itself and the ratified milestone clock must not move. The earlier epoch-8 first-hour
evidence stands, as H3 below shows: nothing it measured depends on Reputation.

**What landed:**
- **H1.** `FirstHourRunResult.reputation_exits` records, for each Exit: `run_seq`, `exit_type`,
  `lifetime_value`, level before and after, `reputation_delta` and available after. A run that
  reached its elective Exit without a recorded collapse sample fails as
  `reputation_exit_unrecorded`.
- **Re-run report.** `planning/reputation-tree-v1/first-hour-reputation.v1.json`: 97 runs, all
  completed, using the same epoch-8 command and knobs as `first-hour-epoch8-report.v1.json`.
- **H3, first-hour non-regression.** Against the epoch-8 evidence, every run's milestones and ending
  and the whole aggregate are byte-identical. The only differing key is the new
  `reputation_exits`. (This was checked by script; the check is recorded here and is not yet a
  pinned test.)
- **Measured lifetime values.** Scripted first Exit p50: Chaos 2.34e7, Casual 3.92e6. Elective
  collapse p50: Chaos 2.24e8, Casual 5.33e7. The RFC's "~1e4 cash" estimate referred to *cash on
  hand*, not lifetime value.
- **H2 and OD-2.** `harness/reputation_threshold.go` re-derives the paid Reputation at the first
  elective Exit under any threshold, using `prestigecore.ReputationDelta`: run 1's scripted_first
  credit sets the level, and run 2's collapse pays against it. This is exact because runs 1–2 are
  threshold-independent.
- **Pinned measurement.** `threshold-measurement.v1.json` covers a 30-point grid from 1e3 to 5e12.
  - **Thresholds satisfying the envelope** (paid in [3, 10] at Chaos p50 and Casual p50):
    **2e4, 5e4, 1e5, 2e5**.
  - The RFC's 1e8 placeholder pays 0.
  - The live 1e12 pays 0 for every persona. This is H2's demonstrated failing case, and the test
    asserts it.
- **Nothing is ratified or minted.** The report is owner SHA-ratification input (OD-2/OD-10).

**Evidence:**
- `TestReputationThresholdMeasurement` pins the measurement to its source report and asserts the H2
  failing case.
- `TestReputationThresholdMeasurementFailsLoud` rejects missing H1 samples, a failed run, and a
  missing gated persona.
- `TestFirstHourRecordsReputationAtEachExit` is a live Chaos run with two ordered samples, each
  lifetime above 1e5 and each delta 0 under 1e12.
- `make test-go GO_PACKAGES=./harness -count=1` passes.

**Severing probes (each turned the test red; restored):**
- the lifetime hook made a no-op;
- the elective sample dropped (the first attempt did not compile and was redone as a compiling
  mutant);
- the run-1 level ignored in the re-derivation.

**Not yet done:**
- **H4, the runs 1–3 career scenario.** The first-hour runner stops at the first elective Exit
  (run 2). H4 needs a third run with tree purchases and starters applied, under a fixture threshold
  taken from this measurement.
- **H5, per-node relevance across runs.**

These are the next batch.

## 2026-09-25 — B8b landed: H4 runs 1–3 career (Claude)

**Implemented by:** Claude; awaiting Codex designated review.

**Self-correction to B8a.** The lifetime-hook edit in `b522c577` attached the hook to the discarded
`commandApplies` probe clone instead of `apply()`. That commit's summary ("the served
`AccumulateLifetimeValue`") was therefore incomplete. This commit adds the hook to `apply()` and
re-measures cold. The first-hour report came out **byte-identical** to the committed B8a report:
every transition follows a same-instant advance, so no accrual is folded into transitions. The B8a
threshold measurement stands.

**What landed:**
- **Harness.** `harness/reputation_career.go` adds career mode to the first-hour runner. At the
  run-2 elective Exit it:
  - credits Reputation under a fixture threshold;
  - buys nodes by the declared policy: `cheapest` for Casual and Reference, `seeded_uniform` for
    Chaos, and `none` for the control arm;
  - assembles run 3 through the served `production.ApplyReputationStarters`, a new thin export of
    R4 step 4 (kernel-guarded, so kernel 0.3.115 → 0.3.116 in this commit);
  - carries the frozen Founder bonus into run-3 production (checked via
    `production.ResolveFrozenContributions`).

  The whole career runs on the tree bundle, which differs from the suite bundle only by the R2
  declaration row. `runExperiment` is unchanged in behaviour: it calls `runWithCareer` with no
  career.
- **Report.** `career-h4.v1.json` covers all 97 seeds, each with a treated and a control arm,
  at fixture threshold 1e5, which is in the OD-2 satisfying set.
  - 93 seeds are gated.
  - 3 Casual seeds are **excluded, and visible**, because the run-3 gate lies beyond the ratified
    2-hour horizon in both arms. I did not extend the horizon to fit.
  - Time saved on the run-3 Garage gate (min/p50/max): Chaos 38 / 118 / 212 s; Casual
    15 / 80 / 350 s.

**H4 verdict: FAIL, recorded rather than loosened.** In 6 Casual seeds, run 3 with starters reaches
the Garage gate at *exactly* the same attended time as control:

| Seed | Gate time (both arms) | Starters |
|---|---|---|
| 1 | 275 s | `cash_small` |
| 6 | 425 s | `cash_small` |
| 8 | 405 s | `cash_small` + `generated_beige_tower` |
| 11 | 265 s | `cash_small` |
| 24 | 360 s | `cash_small` |
| 25 | 430 s | `cash_small` |

Casual's session-quantized decision schedule absorbs the small starter advantage. Under the
proposed R2 data (OD-10 is provisional), the RFC's "strictly sooner at every seed" gate is not met.

This is owner/RFC-author input. Options:
- stronger starters;
- a different Casual node choice;
- stating the gate on a distribution rather than every seed.

Each of these is a ruling, not something I can choose. The report records `h4_gate_passed: false`
and the violations. The test pins the measurement, so drift fails.

**Evidence and severing probes:**
- `TestReputationCareerGateRejectsTiesAndMissingTreatment`: the gate rejects ties, a slower
  treatment, and a missing treatment, and accepts a strictly sooner one. Severing `>=` to `>` makes
  it fail.
- `TestReputationCareerStartersShortenRunThree` regenerates and pins the report. It fails outright
  if no seed is gated, which would be vacuous. Severing the served starter call makes it fail.
- `make test-go GO_PACKAGES='./harness ./production' -count=1` passes.

**Still open:** H5, the per-node relevance report across runs.

## 2026-09-25 — B8c landed: H5 per-node relevance; B9 status (Claude)

**Implemented by:** Claude; awaiting Codex designated review.

**H5.** Career mode gained an `Exclude` leave-one-out mask. The report, `relevance-h5.v1.json`,
covers 97 seeds × (baseline + 9 masks) at fixture threshold 1e5. Δ is the change in the following
run's Garage-gate time when a node is withheld, counted only over runs whose baseline actually
bought that node. It took 618 s cold.

| Node | Result |
|---|---|
| `starter.cash_small` | relevant (Casual p50 +70 s, Chaos +102 s) |
| `starter.generated_beige_tower` | relevant (+15 s, +30 s) |
| `starter.upgrade_continuous_feed_paper` | relevant (Chaos +40 s) |
| `unlock.p05`, `unlock.p25` | `owner_exempt:OD-5` |
| `starter.cash_large`, `unlock.p50`, `unlock.p75`, `unlock.p100` | `unreachable_in_horizon` (never bought) |

On `p05` and `p25`: withholding them makes run 3 *faster* (Chaos −24 s and −46 s), because the
freed Reputation buys starters instead. At these levels the bonus is +0.03% to +0.3%, which matches
the RFC's DG-7 expectation that these nodes cannot be relevant in Phase 1.

**Gates:**
- A starter that is bought but moves nothing fails the test.
- An exclusion without a reason fails.
- `TestReputationRelevanceClassification` pins the classification rules. Widening the OD-5
  exemption to starters fails it.

**DESIGN-GAP RT-DG-F:**
- R10 names ε without giving a value. I used 1 ms, i.e. any strict p50 improvement, and recorded it.
- The first-elective-Exit dimension of H5 needs a run-4 horizon, which the ratified scenario does
  not have.

**B9 status:**
- `make formulas-check` is clean. No epoch pins a tree, so the formula artifact correctly carries no
  `reputation_tree` provider yet. The provider regeneration (R8, AC14) lands in the owner-gated mint
  commit.
- AC15 (a composed real-Postgres career across three runs) also needs a composed epoch pinning the
  tree. I did not fake it with a fixture epoch in the composed lane. Postgres witnesses cover the
  same transitions piecewise: purchase, Exit plan, activation, and frozen row.

## 2026-09-25 — Fast-harness timeout regression fixed (Claude, orchestrator)

**Implemented by:** Claude. Awaiting Codex's designated review.

`TestReputationTreeRelevance` (875 s) and `TestReputationCareerStartersShortenRunThree` (201 s)
pushed `make test-harness`, and with it `verify-harness-fast` and the 5-minute push harness, past
Go's default 600 s timeout. The fix moves this evidence to its own lane without loosening any bound:
- Both tests now call `requireReputationExhaustive`. Without
  `CLOUD_CLICKER_REPUTATION_EXHAUSTIVE=1` they skip, with a named `t.Skip` pointing at the lane
  that runs them.
- A new `make reputation-harness-check` target runs exactly those two tests with a 60-minute
  timeout.
- A new maintenance job, `reputation-evidence` (60-minute budget), runs that target.
  `client/tools/ci-topology.mjs` now requires the job, its budget and its run line, and three Go
  jobs caching modules only.

Evidence, all cold:
- `make test-harness` passes in 32.8 s.
- `make reputation-harness-check` passes: career 145.69 s, relevance 645.47 s, package 791 s,
  exit 0.
- `make verify-ci-topology` passes, and its negative controls reject 13/13.
- **Severing:** deleting the `make reputation-harness-check` line from `maintenance.yml` fails with
  "maintenance must run the exhaustive Reputation harness evidence". It was restored afterwards.

## 2026-10-06 — bounded R2 starter-wire parity diagnosis predeclaration (Codex)

Source coordinate 6947ebe3, clean tree. Previous goal turn completed the Clout
cross-party guard verdict and authoritative record reconciliation: progress,
not a waiting loop. This separate accepted Reputation R2/AC1 lane targets only
the nested starter closed-key grammar from Claude's original a522fdf1; not a
designated verdict on its entire twenty-one-path B1 commit or later B2–B9.

Source lead: Go decodes the shared curriculum struct then tests nonzero values;
TypeScript checks the starter arm's exact keys. Predeclare actual fixture-based
controls for all three starter kinds: untouched legal arms must load; add each
other arm's field with its zero/empty and null representations, and require
ErrInvalidTree/TS rule-7 refusal. Source inspection is not yet an executed finding.
No caller bypass, invented economy/copy literals or live player exploit claim.

If these populations confirm the discrepancy, accepted R2's closed union
authorizes a narrowly scoped Go raw-key check, shared test evidence and canonical
docs. Predeclare that correction separately after observing failure. Preserve
curriculum's archived behavior, starter effects/values/order/headroom, threshold,
save/migration/API/auth/CI/workflow/mint/copy bytes and all plan checkboxes. Any
kernel-watched runtime change must honestly advance all three kernel identities.
Final selected/full root client and Go checks must execute; restore a single
raw-key-check severing to prove discrimination, with all handles terminal before
each mutation/restore. Docker remains full; no new workload or cache deletion.
Claude reviews the complete resulting Codex span; no self archival/acceptance.

### 2026-10-06 — RP-243 confirmed; correction predeclared

Actual Go 11207 exits Make 2/package FAIL (-count=1): all twenty malformed
starters load with nil error, each reported as `cross-arm starter key admitted`.
The untouched fixture loads all three legal kinds first. Root client 62669
passes 84 files/7367 tests, with 17 files/134 existing skips; the new shared
twenty-case TypeScript population rejects all malformed rows at rule 7.
No DB/browser/served-game evidence, no whole-CI claim. Both handles terminal.
RP-243 entered immediately; no reinterpretation of these failures as green.

Correction authority: accepted Reputation R2's exact nested closed union/AC1.
Allowed scope: server/reputation raw starter-key validator before shared struct
decode, same twenty-case corpus and Go/TS tests, docs, all three kernel version
identities 0.3.153→0.3.154, records. Do not change shared curriculum loader,
legal starter arithmetic, ordering, economy/balance/copy/mint/API/save/CI bytes.
Require the Go negatives to pass without changing their inputs/oracle; temporarily
remove only the new raw-key call and require the same twenty failures, restore
exactly, then cold related Go/vet and full root client/type/build/boundaries.
Run root composite too: prior RP-131 is expected to remain a separately reported
historical refusal, never bypassed or described as complete CI success. No plan
checkbox or B1/whole-RFC approval. Claude's designated review is mandatory.

### TypeScript oracle control predeclaration, 2026-10-06

All Go/composite/copy handles are now terminal. Also exercise the new TS
assertion's discriminator: temporarily let only rule-7 exact-key calls return
their unchanged input, retaining semantic validators and all other exact-key
checks. Full root test-client must fail the new cross-arm refusal test; restore
client/src/reputation.ts SHA cb36e5e69ba9addea25f639f858c8840f3ae81238ff660a1d6219b402a636922
exactly and rerun full client tests. No persistent TypeScript runtime change,
admission-policy weakening, skip, oracle relaxation or CI change authorized.

## 2026-10-06 — RP-243 correction and executed first filter (Codex)

**Review by:** Codex (self/first-filter of this correction, not designated approval).
**Recorded by:** Codex.
**New scope:** after 6947ebe3 exclusive, including dc9e6fc6 predeclaration,
8379f96d failing-first shared fixtures/tests and correction predeclaration, the
runtime/docs/kernel/records change and following exact pin. Literal tip follows.

Targeted original B1 finding concerns R2/AC1 on `a522fdf1^..a522fdf1` only:
Go and TS starter parsing and their tests were inspected, with shared curriculum
semantics as dependency. Blame confirms the challenged decode block originated
in a522fdf1. This is **not** a verdict on all twenty-one paths or the original
B1/full R2/AC1 acceptance. Original scope remains unapproved; this Codex repair
needs Claude independently, not a recorder-relabelled cross-party verdict.

The Go loader now validates the chosen nested arm's exact raw key set before
decoding the shared curriculum struct. Zero, empty or null values cannot erase
other arms' key presence. Existing semantic/headroom checks remain unchanged.
Only Reputation admission changes; shared curriculum and TypeScript runtime
remain byte-unchanged. All three kernel identities honestly move to 0.3.154.
No balance/copy/epoch/mint/scenario/threshold/save/migration/API/auth/CI bytes,
plan checkboxes, owner choices or player-facing prose changed.

Executed evidence, all terminal before any source/record edit:

- 11207 original Go baseline: legal fixture's three kinds load, twenty malformed
  cases incorrectly load; package FAIL/Make 2. TS root 62669 passes the same
  twenty refusals inside the new test, total 7367 pass/134 existing skips.
- 99537 corrected whole reputation package: all old tests and twenty new
  negative subcases pass cold, package 0.230 s.
- Single Go call severing (1ed8c1 tool chunk): twenty new cases refail with nil
  admission error; Make 2. Restore SHA
  78f908fb1fbdf0da05e0bb2a5d9e5a705919ce17d405a24c1f3bafd52eda4b30.
- 70103 related reputation/curriculum/replaycatalog/production/save/kernel tests
  and selected vet pass, -count=1 (production 37.326 s). Host mode includes
  dependency-skipped DB tests; not executed Postgres evidence.
- 71532 root verify-client: typecheck/strict TS/Svelte zero errors/warnings,
  built client, 84 files/7367 tests pass (17 files/134 existing skips), shell
  boundaries and checkout controls pass; composite **FAILS** unchanged historical
  kernel guard at 50a3a514 / RP-131. No whole-CI green claim or guard bypass.
- 34617 separately executes all remaining root client gates: topology/thirteen
  negative controls, combat/meters/achievements/cosmetic/no-payment boundaries,
  copy and content-manifest check pass. Copy retains 610 orphan warnings,
  657 keys; no copy regeneration or warning suppression. Standalone kernel
  guard adversarial fixture population 16255 also passes, not history closure.
- 74883 full root verify-server-core passes: vet ./..., all non-harness Go
  packages cold (-count=1), Pitch content, formula and API generation drift,
  Routes and Commons boundaries. No generated artifact diff. Its host-mode
  dependency skips are not real-DB acceptance; exhaustive harness is not run.
- 95141 TS rule-7-only exact-check bypass: full root client fails precisely the
  new cross-arm test, first row resource_grant generator_id empty, while 7366
  old tests pass. The loop stops at that first counterexample: not twenty
  independently fired TS mutations. Restore runtime SHA
  cb36e5e69ba9addea25f639f858c8840f3ae81238ff660a1d6219b402a636922.
- 73164 final restored root client: 84 files/7367 tests pass, 17 files/134
  existing skips; no pending handle or persistent mutant. Go source hash still
  matches corrected 78f908fb; final diff-check passes.

Read-only Docker capacity recheck: writable overlay 100%, 39,784 KiB available,
while DB tmpfs/shared memory have headroom. No workload/restart/deletion or
assumed cleanup approval. RP-236 remains independent of this successful CPU lane.
Full Reputation still needs its complete range review, ruled measurement/mint,
real career/replay/default player surfaces and platform obligations; no archive,
push, threshold retune or shrinkage of full nine-tier 1.0 follows from this fix.

### Exact corrective span pin, 2026-10-06

Substantive Codex span `6947ebe3..4f1e6839`: three commits / fourteen paths,
including dc9e6fc6 predeclaration, 8379f96d fired original baseline/shared tests,
and 4f1e6839 runtime/kernel/docs/ledger/board reconciliation. This following
pin edge also belongs to the requested designated review; closing relay names
its exact literal tip. Not the original a522fdf1 B1 union or any earlier scope.
Postcommit 96739 reruns reputation/kernel from 4f1e6839 cold, both pass (0.196 /
0.167 s); all handles terminal, diff-check clean. Claude pending, no self-verdict,
checkbox, archival or push. Next accepted CPU scope: R1/R3 bonus/accounting
discrimination, preserving all current owner, DB/browser and CI blockers.

## 2026-10-06 — R1/R3 earned-level bonus evidence predeclaration (Codex)

Source coordinate 5467f575, clean main. Previous turn made progress: RP-243
runtime correction/shared failing fixtures, whole server-core and client checks,
two restored runtime probes, truthful kernel 0.3.154 and reconciled records.
This next accepted R1/R3/AC5 lane must not consume that pending Claude review.

Population: existing ten Go-authored canonical bonus vectors plus a new shared
parameter matrix: levels 0,1,2,4,552,1000,MaxExactInteger; per-level ppm
1,10000,1000000; unlock ppm 0,1,50000,250000,500000,750000,1000000. This is
147 parameter triples, not exhaustive safe-integer proof. For each, distinct
legal spends among 0,1,floor(level/2),level give 462 cases: available must equal
level-spent while bonus equals the zero-spend bonus. Zero-level/zero-unlock
must be neutral. Preserve independently hand-checked anchors already pinned.
Eight shared invalid integer-domain tuples must reject in both pure helpers'
applicable validation, without DB/save/API or caller-coercion claims.

Controls: independently replace the earned-level numerator with available
(level-spent) in Go and TS, one at a time; require actual existing/new tests to
fail. Independently remove each runtime's spent>level admission guard and require
the invalid tuple test to fail. Restore exact source SHA after each terminal
handle; no editing under a live check. Only tests/shared fixture/docs/records
persist: no runtime, kernel, copy, balance, save, schema, CI, mint or checkbox
change. If another runtime defect fires, record it and separately predeclare
any accepted-contract fix; do not smooth it into the evidence batch.

Run cold root selected Go/vet and root full client/type/build/boundary/topology.
No fresh Docker workload while capacity approval is pending. Designated Codex
review may assess only original Claude a522fdf1 R1/R3 primitive arithmetic, not
its full twenty-one-path B1/loader/copy or later codec/transaction/UI/mint scope.
Our new tests/records remain Codex first-filter and require Claude's exact-span
review. Full Reputation and full nine-tier 1.0/platform floor stay unchanged.

## 2026-10-06 — R1/R3 sampled proof and bounded original-property verdict

**Review by:** Codex (designated other party for the original Claude primitive;
self/first-filter only for the new Codex tests/records).
**Recorded by:** Codex.
**Original coordinate:** `a522fdf1^..a522fdf1`, limited to Go Available/BonusFactor,
TS reputationAvailable/reputationBonusFactor and their accounting/ten-vector
tests plus bonus-vectors-v1.json. Current-source comparison confirms those
helpers and the ten vectors are byte-unchanged from that commit. The intervening
raw starter repair and later Purchase/OwnedStarters additions are not reviewed
or approved by this verdict.
**Verdict:** APPROVED for those bounded primitive properties only. **Not** a
full commit/path-union archival verdict, full B1, R1 codec/owned-id/mirror
acceptance, R3 frozen-row/next-run transaction proof or full Reputation approval.
Original twenty-one-path B1 and RP-243/new-test review requests remain open.

Population under ce6a9d6b: the existing ten canonical cross-runtime vectors
plus shared bonus-domain-v1.json. Its sampled 147 parameter triples execute
462 distinct legal spend cases in each pure runtime, with neutral zero inputs.
Eight invalid bonus tuples reject; the four accounting-invalid rows also
reject Available while bonus-only-invalid rows leave valid Available usable.
This is not exhaustive safe-integer coverage or a byte comparison of all 147
baseline factors across runtimes; those cross-runtime bytes are checked by the
existing ten-vector corpus. No alternate rounding oracle or acceptance bound
was invented. No fixture/vector regeneration or experimental balance adoption.

Actual executed sequence, every handle terminal before mutation/restore:

- 50783 baseline accounting/vectors pass cold, package 0.141 s; full original
  client 44641 passes 7367 tests/134 existing skips.
- 17427 augmented whole reputation package passes cold, 0.249 s, including
  all twenty retained RP-243 cases and eight named new invalid-domain subcases.
  Client 75094 passes 7369/134, 84 passing files/17 existing skipped files.
- Go earned→available numerator 84675: Make 2, new spending test fails at
  (level1,spent1,per_level1,unlock50000), expected 1.00000005e0 versus 1e0;
  four existing canonical vectors independently fail (indices 3–6). Restore.
- Go spent>level guard removed (865b94 result): Make 2; existing accounting,
  new spent_over_level and existing invalid bonus vector assertions fail.
  Restore exact original current source before starting TS probes.
- TS earned→available 29742: Make 2, five tests fail. New spending and existing
  vector tests fail, plus three existing new-run/Exit-plan replay event checks
  report changed run_started bonus factors. These are actual fixture consumers,
  not real DB/default player/production-mint evidence. Restore.
- TS spent>level guard removed 39844: Make 2, four tests fail, including the
  new invalid-domain test, both existing pure assertions and the existing
  Founder-state restore rejection. Restore; no codec implementation changed.
- Final 39481 cold reputation/kernel and selected vet pass (0.143/0.062 s).
  Final root typecheck/build/client/boundary/topology 64285 passes: zero TS/
  Svelte errors/warnings; 7369 tests, 134 existing skips, 84/17 files; thirteen
  topology negatives reject. No fresh composite/historical guard/whole-CI or
  exhaustive harness/real Postgres/browser claim in this wave.

Restored unchanged runtime SHA:
Go tree.go 78f908fb1fbdf0da05e0bb2a5d9e5a705919ce17d405a24c1f3bafd52eda4b30;
TS reputation.ts cb36e5e69ba9addea25f639f858c8840f3ae81238ff660a1d6219b402a636922.
New shared domain fixture SHA:
0bf5cca86888354eaec0eff6986c374bef5509c9c46eab89967493e52f698b17.
All four probes fired; none omitted, no compilation-only failure or surviving
probe reclassified. Existing strict-loader defect is not silently promoted by
these arithmetic successes. Current kernel remains 0.3.154: new tests only,
Go _test.go explicitly excluded by the existing guard; no false version signal.

RP-244 and canonical/current/execution/roadmap records retain the evidence.
New Codex test/doc/record span begins 5467f575 exclusive, including this wave's
predeclaration ce6a9d6b; exact tip follows. Claude independently reviews this
span, not substituted by Codex's limited original-property verdict. No runtime,
balance/copy/save/schema/API/CI/mint/owner-choice/checkbox/archive/push change.
The full nine-tier 1.0/platform objective and H4/measurement/mint/capacity,
real DB/browser/previous review blockers all remain unconsumed.

### Exact R1/R3 test-evidence span pin, 2026-10-06

Substantive Codex range `5467f575..cbb5a998`: two commits / ten paths,
predeclaration ce6a9d6b and shared tests/docs/verdict/ledger/board records.
This following pin edge also belongs to the review; closing relay names its
literal tip. Claude must review this new test/record work independently of
the bounded verdict on Claude's original helpers and the earlier RP-243 span.
Postcommit 77279 cold reputation/kernel rerun passes from cbb5a998; all handles
terminal, source remains restored, diff-check clean. No full B1/RFC/archival
or release approval. Next bounded accepted lane: R4 additive starter/new-run
fixture consumers, without inventing a mint, threshold or owner decision.

## 2026-10-06 — R4/AC8 additive starter witness review predeclaration (Codex)

Clean source coordinate e365e0da. Inspect the complete four-path Claude witness
range `7d130b89^..7d130b89`, not the earlier B5 producer implementation or later
Exit-plan changes. Execute its current Go corpus producer and TS replay consumer
with committed artifacts, without fixture regeneration. Required existing case:
scripted-first burnout curriculum assigns ten provisioned Beige Towers, then
the owned Reputation starter adds five; purchased remains zero; canonical
receipt/state/run_started-v2 bytes and ordered applied ids agree.

Controls: independently replace the Reputation generated starter's addition
with assignment in Go and TS. Each existing witness must fail on the actual
15-to-5 regression, not a compilation error. Restore exact source hashes after
each terminal check; no edits while any check handle lives. Run selected cold
production tests/vet and the full root client population before and after probes.
No database, browser, minted-content or full-RFC claim follows from these local
transition/replay fixtures. No runtime change persists, and no new balance,
copy, kernel, CI, schema, owner choice or checkbox change is authorized. Any
unexercised starter arm/order/cap property is recorded as coverage debt, not
inferred from the required additive case. A real runtime defect needs a separate
accepted-contract correction predeclaration before implementation.

Codex can independently review Claude's four-path witness range; this new Codex
predeclaration/evidence/record span remains first-filter only and awaits Claude.
No archive, push, cache deletion or weakening of the complete 1.0 objective.

### R4 review extension predeclared after the two terminal probes

49641 original selected Go passes cold; 21057 original full client passes
7369/134. Go assignment control 3002 fails with generated=5, purchased=0,
cash=1e3; TS assignment 16762 fails the exact receipt's 15-versus-5 difference
(one failing test, 7368 passing, 134 existing skips). Both source hashes restored.

The existing required AC8 case owns cash_small/generated only. Its two starter
ids happen to be both tree-sorted and byte-sorted, so this case cannot distinguish
those orders and does not exercise preowned_upgrade or cumulative resource grants.
Extend this test-only wave with three shared expected cases on the same committed
bundle/Exit setup: all nine fixture nodes plus one unknown retired id (four known
starters in tree order, cash 1.01e5, fifteen provisioned/zero purchased, owned
Continuous Feed Paper, factor 6.52e0); unknown id only; empty owned set. Last two
retain only curriculum's ten units, cash zero, no upgrade/starter ids, factor 1e0.
Level 552 and known costs/spends use the accepted R2 table; no new balance adopted.
Each runtime drives its real ApplyLoggedExit on frozen inputs, not a test-side
starter implementation. These shared expectations add semantic assertions, not
new all-case canonical Go/TS bytes or DB/default-player evidence.

Additional independent controls in each runtime: suppress preowned upgrade;
sort the emitted applied starter ids bytewise instead of retaining tree order.
Require new cases to fail, restore exact hashes, rerun cold selected Go/vet and
root client/type/build/boundaries/topology. Resource-grant additive semantics are
asserted by the two known cash grants, but not separately mutation-probed in
this declared wave. Cap refusal, already-owned idempotency and next-tree-removal
cases remain separate coverage debt. Tests/shared expectations/docs/records only
persist. Original four-path witness verdict cannot approve these Codex additions.

### Original envelope compatibility refinement, before final record

Full cold server-core 42344 is terminal and passes, as do selected Go/vet
40518 and client/type/build/boundaries/topology 33816. The original four-path
diff is completely inspected: prior version/bundles/purchase cases unchanged,
all forty duplicated artifact strings identical to the prior tree bundle, and
all seven canonical output-string fields agree with their structured fields.
The current original AC8 fixture differs from 7d130b89 in exactly one leaf:
replay_inputs.v rose from 9 to 12 under later carry-version commits; its Go
producer and TS AC8 assertion body otherwise remained unchanged before this
wave. Do not call that historical corpus byte-unchanged.

Add an explicit Go legacy-v9 replay against the same pinned receipt/state/events,
and run the original TS AC8 assertion at both v9 and current v12. This refines
the original-coordinate review to demonstrated legacy consumption, without
approving the later carry-version implementation spans. The three new semantic
effect cases remain current-v12 cases. No fixture regeneration/runtime change.
Also reconcile this system's canonical stale staging statements: implemented
Founder v22 floor/carry no longer await B3/R6; latest supported Founder version
is 25 and Company replay inputs 12, while legacy v9 remains accepted. Source
coordinates are save/state.go, save/runlog.go and production/replay.go version
floors; this edits technical status only, not an owner ruling or design body.

## 2026-10-06 — R4/AC8 witness verdict and full starter-effect first filter

**Review by:** Codex (designated other party for Claude's original witness;
self/first-filter only for this wave's Codex supplements and records).
**Recorded by:** Codex.
**Exact original range:** `7d130b89^..7d130b89`, all four paths inspected:
Go corpus producer, TS consumer, JSON corpus addition and planning log.
**Verdict:** APPROVED for that witness addition. Not the earlier B5 producer
implementation, later carry-version/Exit-plan commits, full B5/R4/RFC acceptance
or an archival range union. Source execution depends on e365e0da plus this
wave's supplemental tests; no intervening implementation is implicitly approved.

The original JSON adds only exit: prior version/bundles/cases unchanged. All
forty embedded artifact strings duplicate the prior bundle; seven canonical
output strings agree with their structured representations. Current original
Exit differs only in envelope v9→v12 (later carry versions), not reward/event
expectations. Existing Go producer and original TS assertion body were unchanged
before this wave. Explicit new Go legacy-v9 replay compares seven output fields;
TS original AC8 case now compares receipt/Founder/new-Company/started-event bytes
at both v9 and v12. All pass without corpus regeneration. Do not cite these
checks as current historical commit execution or approval of later dependencies.

RP-245's three shared semantic cases now execute real ApplyLoggedExit in both
runtimes. Full ownership includes all nine accepted fixture nodes plus a retired
id: the four known starters emit in tree order, credit 1e3+1e5 cash to 1.01e5,
produce fifteen free/zero purchased Towers, own Continuous Feed Paper and record
the run_started summary factor at 6.52e0. Unknown-only/empty-owned cases keep
the curriculum's ten units, zero cash, no upgrade, empty applied ids, unit bonus.
Founder level/spent and retired ownership stay unchanged. The supplement is a
shared semantic expectation table, not new full-byte Go-authored output corpus.

Actual controls/checks, every handle terminal before source edit/restore:

- 49641 original selected Go cold passes (0.431 s), 21057 original full client
  passes 7369/134. Existing assignment regression: Go 3002 fails generated=5
  versus required fifteen; TS 16762 fails exact receipt (one failure, 7368 pass).
- Initial augmented Go c8ee40 is a test-instrument compilation failure: wrong
  outcome constant, Ledger Balance bool treated as error, and pluralized State
  field. Corrected to actual APIs; not a product defect or fired mutation.
  34945 then passes selected production tests cold (0.421 s); augmented client
  52370 passes 7373/134.
- Go upgrade no-op 28795 fails the new full case: empty ownership versus the
  expected upgrade. Restore, then Go emitted-id sort 52394 fails with cash_large
  first rather than last. Neither is a compilation-only failure.
- TS upgrade no-op 28545 fails the same new case; original AC8 still passes.
  Restore, then TS emitted-id sort 21668 fails the full case's ordered event
  array; original two-node case again passes. Both have one failing test,
  7372 passing/134 existing skips. All six declared controls fire; none hidden.
- Final selected full packages/vet 40518 pass cold: production 36.163 s,
  reputation 0.247, save 0.253, kernel 0.174. Host DB-dependent skips are not
  a real-Postgres population. Root client/type/build/boundary/topology 33816
  passes 7373/134, zero type/Svelte errors/warnings, thirteen topology negatives.
- Full root verify-server-core 42344 passes vet/all non-harness Go cold,
  formulas/API generation with no tracked drift and import boundaries. Its
  Pitch sub-target reports cached, but the preceding whole cold Pitch package
  executes (0.319 s); not a cold claim for that cached sub-target alone.
- After explicit legacy-v9 tests: 38789 selected production/reputation pass
  cold (0.368/0.085 s) and selected vet passes. Its name-filtered kernel reports
  no tests to run, not new kernel evidence (full kernel already executed above).
  26759 typecheck/full client passes 7374/134, 84 passing/17 skipped files.

Exact restored source SHA: prestige.go
6d07581c74822e57d63c09b0ed139d98b0b1f44aba566f8d654ad2f4213cb691;
replay.ts ebe2e60186f8abbd28828a4a2bdfee13d216b2d4f00b125fdf663b52da16dc94.
Existing corpus f9b129e36af5b536f67c7eddb8d6088c76ce5cb0b170cfeec6a5172e8a009782
unchanged; new semantic table
bde253bdc99d4430b8ec45f968bbe94742e8cb3a58bf9bf6fafddce94a19000d.
No runtime/kernel/balance/epoch/save/API/CI/copy or owner-choice change persists.

Canonical docs reconcile the actual Founder floor/latest version and implemented
carry instead of retaining obsolete B3/R6 gaps. New Codex test/docs/record span
begins e365e0da exclusive, includes dd887626, 7ff39c12, 8e89164c and final
evidence/pin; Claude pending, independent of this original-witness verdict.
No checkbox flip, archival or push. Cap refusal, idempotent existing ownership,
next-tree-removal cases, frozen persisted rows, real career/default player,
mint/H4/author measurements, full prior range review and platform/CI/capacity
obligations remain. The complete nine-tier 1.0 goal remains active, not narrowed.

### Exact R4/AC8 supplemental span pin, 2026-10-06

Substantive Codex range `e365e0da..97c558b2`: four commits / ten paths,
dd887626 and 7ff39c12 scope declarations, 8e89164c legacy/doc refinement, and
97c558b2 tests/shared expectations/docs/evidence/boards. This final pin edge
also belongs to the review, including the canonical paragraph's clarification
that a tree means at least v22, not always exactly v22 with later features.
Postcommit 19027 reruns the complete reputation/kernel packages cold from
97c558b2, both pass (0.244/0.166 s), unlike the earlier filtered no-kernel run.
All handles terminal, no source mutation or dirty fixture remains. Closing relay
names the exact literal tip including this edge; Claude pending, no self-approval,
checkbox/archive/push. Next accepted local work: remaining R1/R7 codec/mirror
and R4 cap/idempotency/next-tree evidence, retaining every CI/capacity/owner and
full-player/1.0 obligation.

## 2026-10-06 — remaining R4 boundaries predeclared (Codex)

Source coordinate 70f8c8c8, clean main. Previous goal turn was progress: original
Claude AC8 witness reviewed, shared effect/legacy tests and six fired controls,
full server-core/client gates, canonical staging reconciliation. No live handles.
This wave supplements accepted R2/R4/OD-7 only; no production or policy change.

Four shared fixture-only next-bundle cases, all with the accepted all-owned
Founder (earned/spent552, unlock1e6) and actual Exit replay:
1. Remove generated and its dependent upgrade node from the next tree: ten
   curriculum units, no upgrade, both cash grants, only two applied starter ids.
2. Pivot curriculum grants the same Continuous Feed Paper already granted by
   the tree: pre-run cash1e3 selects pivot; one upgrade remains owned, five free
   Towers, both cash grants. Existing union/upgrade, no authored copy changed.
3. Generated grant MaxExactInteger-10: curriculum ten plus tree reaches exactly
   MaxExactInteger without purchases. Strict loader must admit the boundary.
4. Small-resource starter grants 24 company.permits: exact current permit cap;
   cash_large still grants1e5. Strict loader and actual transition must admit.
Every variant derives a new hash from its actual changed artifact bytes and
loads the changed catalogs strictly; no epoch/mint or fixture regeneration.

Negative artifact controls: increase each exact-cap grant by one; strict Go/TS
loaders must reject rule7. Go served starter helper additionally receives an
already valid Company at one beyond remaining headroom (generated Max-4 plus5,
or permit1 plus a valid24 grant): require typed ErrInvalidEngineState, no clamp.
TS private helper is exercised through real Exit replay with an explicitly
fault-injected COPY of the fully parsed next bundle, grant increased by one.
That copy intentionally no longer matches its artifact semantics, is not loader-
accepted/persistable/minted evidence, and must fail specifically at the runtime
guard (provisioned-hardcap / above_hardcap). Original parsed objects stay frozen.

Eight independent source controls, all compiled and restored before continuing:
Go and TS provision guard bypass; strict resource grant switched to saturation;
starter caller switched from next to current tree; preowned grant toggled instead
of idempotently set. Each must fail the relevant new assertion, not a build error.
Existing declared-catalog negatives remain independent of runtime guard tests.
Do not require helper-level whole-batch rollback: the served transaction applies
starters to a private new Company. Only failed target credit/count is checked.

Run cold selected production/reputation/kernel/vet and root client/type/build/
boundaries/topology. No source/test/record editing under live verification handles.
Only tests/shared expectations/docs/records persist; kernel remains0.3.154.
No checkbox, archival, push, Docker deletion/workload, balance/copy/CI/schema or
owner-ruling change. New Codex span starts after70f8c8c8, Claude independently
required; previous spans and full 1.0/platform/CI/capacity obligations remain.

## 2026-10-06 — R4 caps, idempotency and next-tree evidence (Codex)

**Review by:** Codex (self/first-filter on new tests/docs/records only).
**Recorded by:** Codex.
**Verdict:** first-filter PASS, ready for Claude's designated review. No new
designated verdict on the earlier B5 producers or on this Codex work; no archive.

Under e5a8f7f3, four shared starter-boundaries-v1.json profiles execute Go/TS
actual Exit replay against separately hash-derived, strictly loaded next
fixtures. Retired generator/upgrade nodes no longer grant, while their ids and
earned/spent Founder values carry unchanged. Pivot + tree granting the same
upgrade stays one owned upgrade. Generated Max-10 + curriculum10 lands exactly
at MaxExactInteger, unpurchased. The retargeted small grant of24 permits lands
exactly on the existing permit cap, with cash_large's1e5 unchanged. These are
fixture transformations, not balance adoption, copy rewrites or a mint; Go/TS
serialize their changed artifacts independently, so this is shared semantic
expectation evidence, not four shared canonical artifact/output byte pairs.

Each runtime rejects two raw over-cap variants at rule7. Two Go served-helper
inputs violate remaining headroom and reject typed ErrInvalidEngineState, leaving
the failed count/credit target unchanged, not asserting whole helper rollback.
Two TS parsed-bundle COPIES increase a grant beyond cap and reject specifically
at the actual runtime provision guard / above_hardcap. Those copies intentionally
do not match their admitted artifact semantics: defensive fault injection only,
never an admitted pin/DB/default-player claim. Original parsed trees stay frozen.

Baselines:34474 cold selected production passes0.326s;26159 full type/client
passes7383 tests/134 existing skips,85 passing/17 skipped files. Eight probes
all compile, fail assertions and restore before any following check:

- Go generated guard bypass acb412: EngineGuards/generated_cap wrongly returns
  ids with nil error; test fails. Initial orchestration then restores correct
  predicate without indentation (SHA mismatch) and cannot match the next patch's
  substring. Instrument stops before a second probe; exact indentation/hash
  restored at e4e317. Not a product defect, extra mutation or hidden survivor.
- Go resource saturation e5aaf5: invalid permit credit clamps and returns nil;
  EngineGuards/resource_cap fails. Wrong-current-tree7b7467 fails three real next
  profiles (retire/generated_cap/resource_cap). Upgrade toggle76f24a fails
  idempotent's missing upgrade. Each restore verifies exact SHA.
- TS generated guard35cc1b fulfills instead of rejecting, emitting unsafe
  9007199254740992 provisioned units. Resource saturation282687 fulfills with
  permits clamped24 instead of rejecting. Each has one failure/7382 pass/134 skips.
- TS wrong-current-tree5b5e99 fails seven tests: three next profiles, two parsed
  faults and two existing v22 activation/Exit-plan replay cases.7383 total gives
  7376 pass/134 skips, not seven new tests. Upgrade toggle6f7a82 fails only
  idempotent (one failure/7382 pass/134 skips). All source restores byte-exact.
  Verbose mutation output retrieval is truncated in aggregate; stored result
  summaries retain all named failures/totals. No claim of complete diagnostic
  text for the truncated wrong-current-tree output.

Final55929 passes typecheck/build/full client (7383/134), shell boundaries and
all thirteen topology negatives. Full root verify-server-core50811 passes
vet/all non-harness packages cold (production35.794s, reputation0.191s,
kernel0.168s), formulas/API generation without drift and import boundaries.
Pitch sub-target is cached but its complete preceding package executes cold;
host DB skips do not prove integration. No fresh browser/harness/full CI claim.

Restored prestige.go SHA6d07581c74822e57d63c09b0ed139d98b0b1f44aba566f8d654ad2f4213cb691;
replay.ts ebe2e60186f8abbd28828a4a2bdfee13d216b2d4f00b125fdf663b52da16dc94.
Original corpus f9b129e36af5b536f67c7eddb8d6088c76ce5cb0b170cfeec6a5172e8a009782
unchanged; new boundary table5d438535f04dfeea74a946d587e6f26e4f97a7055236143b8b8701d7950255c1.
Only tests/shared expectations/docs/records persist; kernel remains0.3.154.
RP-247 records this narrowed evidence improvement. RP-246 remains a static
curriculum branch-grammar question, not a proven defect or in-scope correction.

New Codex range starts70f8c8c8 exclusive and includes e5a8f7f3 plus final
tests/docs/evidence/pin; Claude pending. Previous spans remain separate.
Next accepted local lane: R1/R7 Founder codec/mirror admission and activation.
Persisted frozen rows/career/default browser, H4/mint/author decisions, historical
CI/capacity and full B1/B5/RFC range review remain open. No checkbox, archive,
push, Docker deletion, source/kernel/copy/owner change or full1.0 scope reduction.

## 2026-10-06 — starter-boundary review-range pin (Codex)

**Review by:** Codex (self/first-filter only).
**Recorded by:** Codex.
**Designated reviewer:** Claude, pending; no archival authority claimed.

Complete new range: `70f8c8c8..HEAD`, where HEAD is the commit containing this
pin, not a future moving endpoint. It comprises predeclaration `e5a8f7f3`,
tests/shared expectations/docs/evidence `4ca549e6`, and this record-only pin:
three commits, ten paths. The predeclaration is not the test implementation.
These ranges do not absorb previous Codex batches or earlier Claude producers.

Committed-source cold check 338549 passes the production Reputation population
in 0.469s. Its filtered reputation package reports no tests to run and is NOT
counted as coverage. Separate unfiltered check 9f30b6 passes the entire
reputation and kernel packages cold (0.099s/0.162s). Earlier full server/client
checks and all eight fired controls retain their recorded scopes and limits.
No test/runtime content changed after those full gates. Pending designated
review does not authorize a status promotion, archival or push.

## 2026-10-06 — R1/R7 encoder counterfact and correction predeclaration

Baseline `6a8ccaf1`, clean. Previous goal turn made concrete progress: committed
R4 tests, eight fired controls and synchronized records. This is a new accepted
R1/R7 lane, not a repeated capacity wait or an expansion of that review span.

Static observation RP-248: `encodeFounderReplayState` copies level/spent/unlock/
owned fields without the Go structural codec checks. Existing TS tests exercise
load rejection, not encode rejection. Three actual production call sites wrap
encoding with catalog-bound restoration/history comparison; this investigation
does not assume corrupted TS bytes were admitted by the authoritative server.

Question: does the encoder refuse the accepted structural accounting domain,
including legacy state leakage, without rewriting valid owned sets or bytes?
Before correction execute 13 active negative cases: over/negative/fractional/
unsafe spent; negative/fractional/unsafe level; negative/fractional/over-one-
million unlock; unsorted/duplicate/nonmechanical owned ids. Execute three v21
negative cases independently carrying nonzero spent, owned ids or unlock.
Positive cases: empty owned, fully spent, unknown retired owned id, exact maximum
earned/spent and legacy clean v21; retain every existing canonical replay corpus.
All must encode unchanged; load remains independently tested against the tree.

Arms: unmodified HEAD vs the same cases after the minimal R1/R7 TS encoder
correction; no output permissiveness/normalization, new schema or new mechanics.
Baseline is defective if any invalid case emits a value rather than throwing.
Each positive retains its exact Reputation values/order and round-trips under a
compatible strictly loaded bundle when its mirror is consistent. Unknown ids
remain permanent facts, not reasons to debit/refund or delete ownership.

Conditional correction authority: accepted R1 load AND encode accounting, R7
pre-v22 corruption rejection. Only add encoder structural validation, tests,
docs, mandatory kernel identity bump and records. No tree lookup without a
pinned bundle: structural codec and catalog-derived mirror are separate layers
in Go; document TS encoding/restoration similarly rather than claim a bare
encoder checks an unavailable artifact. Mirror checks and activation will be
reviewed separately; do not call this whole AC2/AC10/AC11 acceptance.

Controls after a confirmed correction: independently bypass active structural
validation and legacy leakage guard, expecting their encode negatives to fail.
Restore byte-exact before another probe/check. Run current Go structural codec
tests cold, pinned-mirror/activation tests cold, full client/type/build/boundary/
topology and server-core/vet. Retain existing DB skips, RP-131 historical CI red,
Docker capacity and review limits. No source edits while verification handles
live. No new DB workload, cleanup, mint, owner copy, balance, migration mutation,
checkpoint completion, archival or push. Claude designated review pending for
the entire new span after `6a8ccaf1`, independent of earlier ranges.

## 2026-10-06 — R1/R7 TS encode defect reproduced and corrected (Codex)

**Review by:** Codex (self/first-filter on this correction only).
**Recorded by:** Codex.
**Verdict:** first-filter PASS for RP-248's bounded structural correction;
Claude designated review pending. No full B3 or RFC approval/archival claimed.

Predeclaration 8094e914, before new tests or runtime changes. Baseline 470bda
exits2: all 13 active negative cases and all three independent legacy leaks
emit instead of throwing (16 failures/7387 passes/134 existing skips). All four
new valid cases and the existing clean v21 control pass. Cold Go baseline
3fe84c passes save's original v22 encode/load structural tests and production's
pinned-mirror/activation tests (0.336s/0.184s), not a DB run.

Minimal TS correction validates available from exact earned/spent, unlock
integer domain and sorted unique mechanical ids before serialization. Before
v22 it requires zero spent, empty internal owned ids and zero unlock. No valid
id sorting/repair, refund, purchase-cost rederivation, save-schema change or
automatic activation. Pinned-tree derivation remains the catalog-bound
restoration/validation layer, not a guessed value in a context-free encoder.
Docs now distinguish those layers explicitly rather than imply a bare codec
has artifact authority. Kernel0.3.154→0.3.155 in all three identity files.

7403 tests/134 skips pass after correction (ebdc3a), typecheck zero errors and
warnings. Two compiling mutants, each restored before the next check:

- Active encode check removal 808dbb exits2 with exactly the 13 active negatives
  failing and 7390 passing; all legacy/positive tests remain green.
- Legacy leakage guard bypass d5e52b exits2 with exactly the three legacy
  negatives failing and 7400 passing; all active/positive tests remain green.

Both restores verify corrected replay.ts SHA
834dc5b265da8a119e24ea686f6063f1ac4251ff5ee485f77e913dec22a0c9a0.
New test SHA93b5e35a8658a9e7a7625f8737e49fc255d2eceb0bde5e7a6a9021d712646f29.
No other runtime source, production balance, owner copy or DB migration changes.

Final full verify-client reaches type/build/tests (7403/134) and shell boundaries
green, then fails849de2 at the unchanged pushed50a3a514 against0cf9f7a6 history
violation (RP-131). Not a green composite; no rewrite, exemption or bypass.
Its later gates execute separately and pass7e1181: topology13 negatives,
combat/meters/achievements/cosmetic boundaries, no-payment6 negatives/2 near
misses, copy657keys with610 orphan warnings and deployment content manifest.
Full verify-server-core832782 passes vet/all non-harness packages cold
(production34.014s,save0.277s,reputation0.224s,kernel0.168s), formulas/API
generation without drift and import boundaries. Pitch subtarget is cached,
but the preceding complete Pitch package ran cold0.278s. Host DB skips are
not integration; no fresh harness/browser/real PG/complete CI claim.

RP-249 independently verifies R7/AC10's absent corpus work: the shared corpus
baseline is still11 legacy v1–v9 cases, not the four named Founder cases.
Original RT-DG-B openly substituted codec/activation unit tests; it did not
record an owner waiver. B3's plan text wrongly included the corpus and ratchet,
now corrected without flipping its box. Next accepted R7 work must satisfy
the actual corpus and baseline, with activation/pinned-mirror consumers and
fired controls, rather than relabel existing units. No automatic pre-v22
migration on mere load. This correction does not close full AC2/AC10/AC11.

The new review span after6a8ccaf1 includes predeclaration, correction/tests/
docs/records, and the final range pin. Prior ranges remain independent. Owner/
author/mint/H4, persisted rows/career/default browser, RT-DG-B, historical CI,
Docker capacity and full nine-tier1.0 obligations remain. No Docker deletion,
new DB workload, archival, checkbox completion, push or release promotion.

## 2026-10-06 — Founder encode correction range pin (Codex)

**Review by:** Codex (self/first-filter only).
**Recorded by:** Codex.
**Designated reviewer:** Claude, pending.

Complete new span: `6a8ccaf1..HEAD`, HEAD meaning the commit containing this
pin, not any later endpoint: predeclaration `8094e914`, implementation/tests/
docs/evidence `e06762fd`, this pin (three commits, thirteen paths). Original
Claude B3 and all earlier Codex spans remain independent; no union promotion.
After implementation committed, 759398 runs save/reputation/kernel completely
cold and passes (0.207s/0.099s/0.061s). a5558d passes full client7403/134.
No source/test edits after final gates, identities agree at0.3.155 and both
mutations remain restored. Historical RP-131 composite failure is still open.

Next accepted action is R7's actual four-case corpus/activation and pinned-mirror
evidence plus baseline ratchet, not rewriting the save schema or activating on
load. RP-249 records the missing gate rather than treating prior unit tests as
its substitute. Whole B3/AC2/AC10/AC11 and full1.0 are NOT claimed complete;
no archive/push, capacity consent, mint or owner-ruling substitution.

## 2026-10-06 — R7 shared corpus completion predeclaration (Codex)

Clean baseline39912364. Previous turn made concrete progress: RP-248 actual
encoder defect corrected,16 baseline failures/two fired controls and records
committed. RP-249 is accepted R7/AC10 work, not an owner question disguised as
implementation. Existing RT-DG-B's substitute is not completion authority.

Extend testdata/save-migrations.json to corpus9 with a distinct founder_cases
arm. Keep the eleven legacy case objects byte-unchanged. Add exactly the four
accepted names: founder-v21-to-v22; founder-v22-spent-over-level;
founder-v22-unlock-mismatch; founder-v21-nonzero-unlock. Ratchet baseline total
11→15 and required names together. The current Go-authored Reputation replay
corpus supplies the full encoded input and compatible artifact/Exit contexts;
bind it by explicit source path, SHA and case name rather than copying its
large embedded artifacts. Modern patches and exact expected Reputation fields
are declared in the shared migration corpus, not independently in each runtime.
This is a referenced corpus, not four newly independent full-save byte fixtures.

Consumers: old save corpus runner checks the combined census and still executes
all eleven legacy cases; a save_test external-package adapter exercises public
Go save/production/replaycatalog APIs (no production exports/import cycle).
TS consumer strictly loads the same artifacts and exercises both actual Company
Exit and Founder-log replay for the positive activation; rejects the three
negative loads at their actual structural/pinned boundaries. No silent skip or
empty modern population is a pass; version/count/names/source pin must fail
on malformed/absent rows. Both replay paths must retain earned4, spent0, owned[],
unlock0 at v22 and preserve the original complete canonical result bytes.
Loading the v21 source must leave it at v21; activation is a recorded Exit.

Controls predeclared: independently default spent to level in Go live settlement
and Go Founder replay activation; suppress Go spent guard, pinned mirror and
pre-v22 leakage checks; suppress TS spent/load mirror/pre-v22 guards and alter
TS Founder activation default. Each compiling control must fail the targeted
case, or be recorded as surviving defense-in-depth, never hidden. Removing a
required modern case must fail the census. Restore exactly before any next run.
No runtime mutation persists in this test-only wave; kernel remains0.3.155.

Run new Go population and old migration corpus cold, full server-core/vet and
client/type/build/boundaries/topology. Historical RP-131 client composite RED
stays open; no new DB workload/cleanup permission, minted tree or default-player
claim. No migration body, save schema, owner copy/balance, checkbox completion,
RFC status, archival, push or policy amendment. All new files/tests/docs/records
after39912364 need Claude's designated range-union review, independent of prior
spans. Full R1 mirror authority, B3/AC2/AC11 and the whole nine-tier/platform
goal still require their own evidence; this wave claims only actual R7 corpus.

Instrument refinement before its negative probes: explicitly remove one required
modern row (census must fail), falsify the source SHA (pin must fail), and add an
unknown modern-row key (closed shape must fail). Each runs both Go consumers and
the TS lane, then restores corpus bytes exactly. Runtime controls are already
terminal/restored; no simultaneous source/test edits or deferred restore.

## 2026-10-06 — actual R7 corpus/ratchet supplied (Codex)

**Review by:** Codex (self/first-filter on the new tests/corpus/docs/records).
**Recorded by:** Codex.
**Verdict:** first-filter PASS; Claude designated review pending. Not full B3,
earlier-version activation-chain, R1/R7/RFC or archival approval.

Under a499a5b9, corpus9 adds exactly four named Founder rows in a modern arm;
baseline15 requires them together with the eleven legacy rows. Old row bytes
are unchanged (278c8d, exact textual comparison against39912364), not just their
decoded values. The modern arm uses source path/SHA/case reference to the full
existing Go-authored input/catalog/Exit contexts, plus shared input patches and
exact Reputation expectations. It is NOT four independent new full-save byte
fixtures. Source SHA remainsf9b129e36af5b536f67c7eddb8d6088c76ce5cb0b170cfeec6a5172e8a009782;
corpus SHA f6cc5be4aa0c7af330b0c52023245e24da673872d49e5d56a8d0f574651af389;
baseline SHA2b4913637114b46b55e8084f402434f437abf77be064136b760004c4a80fc4c7.

The save package's legacy runner validates the combined census but executes
only its eleven old rows. The external save_test adapter strictly loads real
catalogs via replaycatalog and executes all four modern rows using public
save/production APIs; no test-only production exports/import cycle. TS executes
the identical modern rows. v21 load remainsv21; Company Exit/live-settlement
and Founder-log replay both activatev22 with earned4/spent0/owned[]/unlock0.
Original canonical Founder post-state, Company genesis and receipt bytes match;
this wave does not newly assert every ordered event or DB transaction boundary.
Overspend/pre-v22 corruption reject typed Go ErrInvalidState; mirror corruption
is structurally legal but rejects at CatalogBundle.ValidateFoundationState with
ErrInvalidEngineState. TS catalog-bound restoration rejects corresponding inputs.

Initial instrument defects, not product findings: 9c8060 is a new adapter compile
failure (used exit.Outcome/Receipt instead of exit.Decision); fixed before
4486ab, which runs both parents/all fifteen subcases cold and passes0.332s.
34620e is one TS test's wrong error regexp ('accounting' vs actual 'invalid
reputation state'); fixed without runtime edits. Then87505a passes7408/134.

Nine runtime controls all compile and fail the expected new corpus row:

- Go live default spent=level31244f and Founder replay default2fe00b independently
  report spent4 rather than0 on founder-v21-to-v22.
- Go spent guard23acfc admits over-level input; mirror guard24d618 admits the
  false pinned mirror; pre-v22 guardc3ffb4 admits nonzero unlock. Each isolates
  its targeted modern row and exits2.
- TS spentd90bea, mirror86092a and legacy9a42de each fail two tests: the targeted
  new corpus row AND an existing codec test (7406 pass/134 skips).
- TS Founder default3b591c fails the new activation row and two existing Exit
  replay cases (7405 pass/134 skips). No survivor or compile-failure-as-severing.

Three corpus corruptions also fail in both lanes: removed row179c33/87dd7e
reports14of15 and3modern, TS1fail/7406pass/134skip (denominator7541);
wrong SHA7a68e5/8fb30e fails source pin (TS1fail/7407pass/134skip);
unknown modern key722d58/d9390f fails closed shape (same TS totals). Nine source
mutations plus three distinct corpus corruptions across two lanes = fifteen
executed negative lane runs. Every restore verifies exact SHA before any next
run; all runtime files match the baseline hashes, kernel remains0.3.155.

Final root client/type/build/boundaries/topology5669c5 passes7408/134, zero
typecheck errors/warnings,213module build, shell14/8/22 and13topology negatives.
Final full verify-server-corebc63c3 passes vet/all non-harness Go packages cold
(production34.343s,save0.171s,reputation0.197s,kernel0.171s), formulas/API
generation without drift and import boundaries. Pitch subtarget is cached;
complete preceding Pitch package ran cold0.313s. Host DB skips are not PG
integration. No fresh browser/harness/completeCI run; RP-131 client composite
RED from the prior turn remains unresolved, not relabelled by this subset.

RP-249's actual four-case corpus/ratchet is locally supplied, unlike RT-DG-B's
prior substitution. Plan text says so without flipping any checkbox. This is
not all earlier-version activation-chain or complete codec/pinned admission
coverage. Next full original B3 producer review and remaining R1/R7 evidence;
full B3/AC2/AC11, career/frozen-row/DB/default-player, H4/mint/owner/author/CI/
capacity and complete nine-tier/platform1.0 remain. New review span begins
after39912364 through its final pin, independently pending Claude. No persistent
runtime, balance, save schema, migration body, owner copy or CI edits; no new DB
workload, Docker deletion/consent, archival, status promotion or push.

## 2026-10-06 — shared migration corpus range pin (Codex)

**Review by:** Codex (self/first-filter only).
**Recorded by:** Codex.
**Designated reviewer:** Claude, pending.

New complete range: `39912364..HEAD`, with HEAD meaning the commit containing
this pin, not later history: predeclaration `a499a5b9`, test/corpus/docs/evidence
`c3274c6f`, this pin (three commits, fourteen paths). Independent of RP-248 and
all earlier ranges; no approval/archival union inferred. No runtime file is in
the persistent diff, kernel identity remains0.3.155. Eleven legacy row bytes
and the referenced source are unchanged; baseline total is now15.

After implementation committed, cold full save/kernel98af9d passes0.756s/0.059s;
full clientc6b590 passes7408/134. No source/test changes after final full gates.
Historical RP-131 composite failure, full earlier-version activation/pinned
admission/B3 producer review, real DB/career/default-player, H4/mint/author/
owner/capacity and complete nine-tier/platform1.0 remain open. Next original
B3 range audit, then remaining admitted R1/R7 evidence; not self archival,
push or an implied Docker-cleanup permission.

## 2026-10-06 — full original B3 range review (Codex)

**Review by:** Codex (cross-party review of Claude's original change).
**Recorded by:** Codex.
**Reviewed range:** `fb0ab3b1^..fb0ab3b1` (all thirteen paths, 597 insertions,
35 deletions, including the three kernel identities, docs and planning claims).
**Verdict:** CHANGES REQUIRED for original B3 completion; no archival approval.

The Go structural/accounting and pinned-mirror boundaries are present and the
new-run default preserves earned Reputation. The original TS encoder did not
validate accounting before emission: RP-248's executed sixteen failing inputs
demonstrate that defect, not an inference from a green decoder test. The original
doc's load/encode rejection claim was therefore false. R7/AC10's four shared
migration rows and baseline ratchet were absent; RT-DG-B disclosed substituting
unit tests, not an owner waiver. This is RP-249, not a second ledger defect.
Original Founder-carry refusal and deferred TS Exit witness are disclosed
incremental dependencies, not evidence of full R6/AC11 acceptance.

Later Codex ranges `6a8ccaf1..39912364` and `39912364..80519365` locally repair
RP-248 and RP-249. Their separate Claude review is pending: this verdict does
not approve my corrections by folding them into Claude's thirteen-path range.
Fresh current-tree cold execution 0f56b8 passes the four migration subcases and
three activation/accounting parents (save0.277s, production0.307s). It is not an
execution of historical fb0ab3b1 or real Postgres. Prior mutation evidence is
explicitly the later ranges' executed record, not a new historical probe.

Earlier-version direct activation remains narrower than R7's stated chain:
the original full-state live test crosses v14 to v21, then v22; its separate
replay helper starts at v21. No full seven-source or catalog-bound encode
verdict is inferred. Full R1 mirror encode authority, all B3/AC2/AC11 consumers,
DB/career/default-player and release remain open. Main is clean; original
history stays append-only. Two unsuccessful lookup commands this session used
nonexistent guessed filenames/unmatched glob; no file was changed by them.

## 2026-10-06 — predeclare earlier Founder activation matrix (Codex)

Authority: accepted R7's every-earlier-chain-step accounting preservation and
new-run-only activation. Test-only supplement, no save-version/mechanic change.

Population: seven legal writable Founder sources v14,16,17,18,19,20,21, each
encoded/restored under its matching internally valid artifact bundle. v15 is
historical decode-only, not an invented writable intermediate; pre-v14 legacy
upgrades remain the existing migration corpus, not this population. Six older
fixture bundles plus current epoch8's v21 bundle target the fixture-only v22
tree. No epoch is minted and no claim of production/default-player integration.

Arms: actual `settleAndActivateFoundations` run-boundary kernel with an old and
new Company, and public `ApplyFounderLogged` with valid exit.v1 audit evidence.
Loading alone must retain its source version/accounting. Both result arms must
encode/restore under the next economy and pass its pinned validation; earned11
remains available11 with spent0, non-null empty owned and unlock0. Whole encoded
Founder result bytes must agree between the two arms after the expected Exit
history append. Existing age/knowledge fields must survive, and prerequisite
feature state must be populated sufficiently for the full v22 codec.

Controls: independently default spent to earned in live and Founder replay,
then independently bypass the v17 and v20 replay activation steps. Each mutation
must compile and fail this new population at its affected earlier sources;
restore exact source SHA before any subsequent run. A surviving control is a
finding, not permission to broaden assertions after observing it. No edits
while any verification/probe is running. Initial setup errors will be recorded.

Exit: seven sources pass both arms cold and all four controls discriminate;
full server-core/vet plus client/type checks cover unchanged consumers. This
does not supply shared TS older-source cases, Company-log/DB write atomicity,
all legacy migrations, full R1 mirror encode authority or whole AC11. New test/
docs/records require Claude designated review. No checkbox flip, balance/copy/
schema/migration/CI edit, archival, push or Docker cleanup/DB workload.

Instrument refinement before controls: the six older bundles retain identical
pitch IDs/rating season across the boundary, so an unrelated added-minigame
migration cannot mask the save-version subject. Their target is a minimal v22
fixture retaining the existing unlock-chain nodes with unchanged values; no
starter is owned or applied. The v21 arm retains epoch8 and the original full
tree fixture. Neither target is a production epoch. Setup run3ead8a rejected
the original full tree because its starter upgrade is absent from the legacy
fixture economy; this is an instrument incompatibility, not a product defect.
After limiting that old-economy target to the unchanged unlock chain,0b1491
passes all seven sources and both arms cold0.547s. No runtime change or
acceptance-bound relaxation. Source and next-floor census added before probes.

## 2026-10-06 — earlier activation evidence and review range checkpoint

**Review by:** Codex (self/first-filter on new tests/docs/records; NOT their
designated reviewer).
**Recorded by:** Codex.
**Designated reviewer:** Claude, pending.
**Complete new range:** `80519365..HEAD`, where HEAD is the commit containing
this checkpoint, not later history. Includes original B3 review/predeclaration
ef54bb45, compatible-population refinement2811ae65 and this test/evidence commit.
The original Claude B3 verdict inside the range is separately bounded to
`fb0ab3b1^..fb0ab3b1`; no retrospective approval union with my new work.

Seven sources v14,16,17,18,19,20,21 pass the live new-run foundation kernel and
public ApplyFounderLogged Exit arm. Sources encode/restore without automatic
activation; outputs encode/restore under the next economy, pass its pinned
validation and match in complete Founder bytes after the independently expected
Exit-history append. Earned11/available11, spent0, owned[]/unlock0, age12345 and
route-knowledge9 are checked, not inferred from a version number. Positive final
census runc5511d passes all seven cold0.358s. v15 remains historical decode-only;
pre-v14 upgrades remain the legacy corpus. Six old minimal fixture bundles
target a stable-key v22 unlock-only fixture; the v21 source targets the original
full tree fixture. This is not all immutable historical artifact populations or
real Company-log/DB/default-player evidence. No TS older-source corpus yet.

Four independent final controls all compile, fail and restore both runtime
SHA values exactly before any next run:

- live spent=earned2620a1 fails all seven sources (expected0, actual11);
- Founder replay spent=earnedc07513 fails all seven sources;
- skipped v17 initialization3c8c58 fails v14/v16 pinned minigame key-set admission,
  while the five already-active minigame sources pass;
- skipped v20 initialization9f2ebf fails v14/v16/v17/v18/v19 valid Exit replay,
  while the two already-active Soul sources pass.

Driver mistake disclosed: the first four runs001be0/7144a6/81d7f8/8cfe36 also
failed, but the first two restore patches reduced indentation on their default
assignment lines. Printed SHA results were not enforced by that driver before
the subsequent probes, violating the declared exact-restore sequence. Diff
inspectiona0c206 showed whitespace only. Immediate patch andf008ab/8b3c6a
restore exact hashes/no runtime diff; then all four probes were rerun with
mandatory hash comparisons before and after each mutation. The first set is
disclosed exploratory evidence, not the final controlled gate. No concurrent
edit, leftover mutation, survivor or compile-failure-as-control.

Restored hashes: foundations.go
7bdeb4e5db7dd180645f3b0d5971a45b7ba9e0d392369c894f16856f520520e6;
founder_replay.go
ecbb5da078af284bafac96ca2929a32e8362a7df0f8157fb90b7ccdb8af17f63.
Complete verify-server-core8f77c6..85dff0 passes vet/all non-harness Go packages
cold: production33.605s,save0.272s,reputation0.202s,kernel0.264s. The separate
Pitch subtarget is cached, while the complete preceding Pitch package ran
cold0.340s. Formulas/API generation is byte-unchanged. db3548..57ff99 passes
typecheck with zero errors/warnings,213-module client build and7408/134 client
tests. Host DB skips are not real Postgres. Full verify-client's historical
RP-131 RED remains unresolved; no fresh browser/harness/full-CI claim.

No persistent runtime, kernel(0.3.155), migration, save-schema, copy/balance or
CI edit; no checkbox completion. Original B3 review is now recorded, its two
corrections and this supplement await their own Claude ranges. Next shared
earlier-source TS replay evidence and remaining pinned load/encode reader/writer
audit under R1/R7. Complete AC2/AC11, transaction/career/default-player, H4/mint/
owner/author/CI/capacity and full nine-tier/platform goal stay open. No archival,
push or Docker cleanup/DB workload; all verification/probe handles are terminal.

## 2026-10-06 — predeclare shared earlier-source Founder replay proof

Previous goal turn made concrete progress: original B3 range reviewed and the
seven-source Go supplement committed at29ed56e8. Main is clean, no running
verification/probe or new external review/push observed. This is the next
accepted R7/R8 task, not a smaller replacement for the full nine-tier goal.

Population: the exact seven previously registered writable sources v14,16–21,
same input states, artifact hashes and two fixture targets. Go authors one new
deduplicated corpus under testdata/replay (nine source/target bundles, seven
rows), including complete pre/post state, canonical command/inputs, receipt,
ordered events and result hash. Existing eleven-case legacy and four-case R7
corpora remain byte-unchanged. Production epochs, balances, owner copy, schema,
runtime and kernel identity are not changed by this test wave.

The existing root Go test flag `-update-replay-fixture`, narrowly selected to
the new activation test, generates the new artifact only after every source's
live/replay/codec assertions pass. Subsequent unflagged runs require byte-exact
regeneration equality; this is a Go-authored expectation, not a hand-maintained
second TS byte table. TS loads the corpus artifacts through its actual strict
loadReplayCatalogBundle, loads without activation, then invokes public
applyFounderLogged and requires complete Go result state, receipt, ordered
events and result hash equality. TS output must restore under the next pinned
bundle and retain earned11/available11/spent0/owned[]/unlock0 and age/knowledge.

Controls, independently with terminal wait and mandatory exact source restore:
TS activation spent=earned; bypass v17 minigame initialization; bypass v20 Soul
initialization. Each must compile and fail the affected registered rows, not
pass via fixture adaptation. Corpus controls: remove one registered row, then
tamper one expected receipt; both Go regeneration and TS consumers must fail.
Restore artifact SHA before the next run. Normal runs never regenerate; a
negative run never uses the update flag. Any setup/compiler mismatch or survivor
is disclosed and routed, not hidden behind a generic green count.

Exit: all seven rows execute in both runtimes, exact regeneration is stable,
three runtime and two cross-lane instrument controls fire, then cold Go source
population/full server-core/vet and full client/type/build pass. No real DB,
Company-log/transaction, career/default-player or all historical artifact claim;
remaining R1 catalog-bound encode consumers/AC2/AC11 and mint/H4/author/owner/CI/
capacity require their own evidence. No checkbox flip, archival, push or Docker
cleanup authority. New range after29ed56e8 needs Claude independently of all
earlier Codex spans; self/first-filter only, no self designated approval.

Generation command refinement before corpus creation:5c6c60 is a command-order
setup failure, not a test/control result. Root test-go places custom flags before
the package selector, so Go's custom update flag made it select server's empty
root. Add a narrow root `reputation-activation-corpus` authoring target beside
the existing replay-fixture precedent, with package before custom flag. This is
generation tooling only, not a CI lane/workflow/count/exclusion change. Normal
root test-go runs stay unflagged and validate byte equality. Makefile is not a
kernel-affecting prefix; no artificial version bump. This extra scoped path also
belongs to Claude's review range. No fixture was written by the failed command.

Instrument refinement before an additional corpus control: the generated rows
each contain exactly one FounderAdvanced event. R8's full ordered-event byte
comparison is exercised, but this population cannot discriminate a multi-event
permutation; no such broader ordering claim is authorized. Independently erase
v14's expected events_json list in a scratch corpus mutation. Go regeneration
and the TS actual-event comparison must fail; restore exact corpus SHA before
subsequent checks. This supplements the predeclared missing-row/receipt controls
without changing runtime, the seven populations or their expected behavior.

## 2026-10-06 — shared earlier-source replay evidence and complete range

**Review by:** Codex (self/first-filter on new corpus/tests/generation/docs/records).
**Recorded by:** Codex.
**Designated reviewer:** Claude, pending.
**Complete new range:** `29ed56e8..HEAD`, where HEAD is the commit containing
this checkpoint, not later history: f1ca58db predeclaration and this implementation/
evidence/checkpoint commit. This range does not approve prior Codex corrections
or the original Claude B3 producer. No archival range union inferred.

Corpus schema1 contains exactly seven registered rows and nine deduplicated
bundles,328345 bytes, SHA
61e5f8f8354e9945088b4dca5023ced850617a486e57268722ae7155bc395e53.
874f0d executes Go assertions and generates it via the narrow root authoring
target. Subsequent unflagged a44487 executes all seven cold0.243s and requires
byte-exact equality with the generated expectations. TS86e972 runs all eight
new tests: exact census plus the seven rows,7416pass/134existing skips, zero
type errors/warnings. All actual strict artifact loads and public Founder Exit
replay pass, including complete pre-state roundtrip without activation and
result state, receipt, event-array and constants-hash equality. The target
restoration checks accounting and age/knowledge. All seven rows have one event,
so this does not newly discriminate a multi-event permutation.

Three independent TS runtime controls compile and fail, then restore exact
client/src/replay.ts SHA834dc5b265da8a119e24ea686f6063f1ac4251ff5ee485f77e913dec22a0c9a0:

- spent=earned66ad4a fails all seven new rows and three existing activation
  cases,10fail/7406pass/134skip;
- bypassed v17 initialization5d9304 fails new v14/v16 and the old minigame
  activation case,3fail/7413pass/134skip; five already-active new rows pass;
- bypassed v20 initialization46e6ea fails new v14/v16/v17/v18/v19,
  5fail/7411pass/134skip; two already-active new rows pass.

Three corpus corruptions fail BOTH lanes (normal Go runs, never update flag):

- missing v14 rowbb3d6b fails exact Go regeneration; TSb2efb4 fails its seven-row
  census,1fail/7414pass/134skip. Denominator7549 reflects the deliberately removed
  row, not a passing full population;
- forged v14 receipt7b7dc2 fails Go; TS7195c2 fails the actual receipt comparator,
  1fail/7415pass/134skip;
- erased v14 expected eventfcb314 fails Go; TS409071 fails the actual event
  comparator with the same1fail/7415pass/134skip.

Three runtime controls plus three corruptions in two lanes = nine negative
lane runs. All mandatory before/after SHA comparisons pass; corpus restored
byte-exact before any next run. No survivor, concurrent edit or hidden failed
population. Initial5c6c60 is the documented command-order setup failure, not
research or a runtime defect. The new Make target only authors this corpus after
the existing assertions pass; normal Go/CI tests compare and never regenerate.

Final58bf0d..4835df verify-server-core passes vet/all non-harness Go packages
cold (production34.398s,save0.274s,reputation0.193s,kernel0.171s,transport13.269s).
Pitch subtarget is cached; full preceding Pitch package ran cold0.316s. Formula/
API generation has no byte drift.83b585..22df5a passes typecheck with zero
errors/warnings,213-module build,7416pass/134skips, shell14/8/22 and13 CI-topology
negative controls. No fresh real Postgres/browser/harness/full CI claim; the
historical RP-131 composite RED remains. All verification/probe handles terminal.

Old migration corpus/baseline and reputation-tree-v1 replay source retain their
previous hashes (2c4206), no persistent runtime, kernel0.3.155, migration-body,
schema, epoch, balance/copy or CI edit. This supplies earlier-source shared TS
replay, not full R1 catalog-bound encode admission, full B3/AC2/AC11/DB/career/
default-player or release acceptance. Next remaining R1/R7 pinned reader/writer
audit and full B4 purchase-contract range review. New range awaits Claude;
whole nine-tier/platform goal stays active. No checkbox completion, archival,
push, Docker cleanup consent or new DB workload.

## 2026-10-06 — predeclare pinned Reputation admission counterfacts

Previous goal turn made concrete progress at501000ea; tree clean, no live check
handles or new Claude verdict. Original B4 has seventeen paths, not yet fully
reviewed: this session inspected its current purchase/resolved/dispatch and
save callback paths, not the whole original corpus/range. No B4 approval claimed.

Candidate R1 defect: structural Go restoration/encoding cannot check a tree
mirror. The live purchase resolver and public ApplyFounderLogged currently lack
pinned Reputation input admission, so a purchase may silently recompute a false
mirror instead of refusing it; a non-purchase arm may record against it. Output
admission also appears absent from the shared successful Founder boundary.
This is an unexecuted inference until the following cases run. Authority is
accepted R1's checked mirror/load-and-encode invariants, not a new policy.

Population: full structurally legal v22 states under the existing tree fixture:
empty owned/spent0 with false mirror50000, and owned p05/spent1 with false mirror0.
Both must already fail pinned validation but pass the bare structural codec.
Each runs the live resolver, public purchase replay with independently frozen
legitimate cost/ownership evidence, and public recorded-invalid-command arm.
Every entry must return a typed invalid-state error, no receipt/event and exact
pre-state preservation, not an ordinary user rejection or silent repair.

An existing deliberate Founder transition test hook injects a false mirror after
a valid starter purchase. Successful output admission must fail and restore the
complete pre-command state without receipt/events. Positive controls cover
empty, known unlock and retired unknown ownership; valid purchase expectations
and existing Go/TS canonical corpora must remain identical. A tree-inactive
pre-v22 control retains R5's ordinary recorded rejection and no mutation.

If baseline fires, record RP-250 and minimally enforce existing pinned checks
at the live resolver and shared public Founder input/output boundaries, only for
activated v22+ state; do not force inactive legacy state through a v22 validator.
Kernel bump in the same runtime change is mandatory. No new category, purchase
price, refund/repaired save, schema, migration body or provider/content changes.
Bare artifact-free codecs remain structural, not falsely labelled pinned.

Controls after the fix: remove live admission (only live negatives should fail
while replay remains defended), remove shared input admission (the recorded-
invalid-command negatives must fail even if purchase remains defended), remove
shared successful-output admission (the deliberate post-state fault must fail).
All must compile, wait terminal, restore exact SHA before next run. Any survivor
or command/setup error is disclosed. Cold full server-core/vet and client/type/
build/boundaries/topology plus unchanged-corpus byte checks follow. No new DB
workload, cleanup consent, checkbox flip, archive, push, full R1/B4/RFC/CI/1.0
claim. New range after501000ea needs Claude independently of prior work.

Baseline d21b4b confirms all seven negatives fail (six input arms plus the
post-state hook), while all four valid/retired/inactive controls pass. After
the minimal checks, f21383 passes all eleven cases.4ea870 passes the seven
earlier activation sources and exact shared corpus regeneration, so output
admission uses the result epoch correctly rather than rejecting legal activation.

Control detail before the remaining output probe: live-guard removal1ef64f
fails exactly the two live negatives; all replay/output/positive controls pass.
Shared input-guard removala37745 fails four assertions, not the driver's expected
two: both recorded-invalid arms actually admit corruption, while both purchase
arms remain defended by the live-resolver check but report ErrInvalidReplayInputs
instead of the required ErrInvalidEngineState. Those two are error-class
discrimination, not accepted purchases. The driver stopped on the unexpected
count after restoring exact SHA; no survivor or source residue. Preserve this
observation rather than rewriting the oracle/count. Next independent output
removal, with the same mandatory before/after hashes; no test/behavior edits.

## 2026-10-06 — pinned command admission evidence and complete range

**Review by:** Codex (self/first-filter on new runtime/tests/docs/records).
**Recorded by:** Codex.
**Designated reviewer:** Claude, pending.
**Complete new range:** `501000ea..HEAD`, where HEAD is the commit containing
this checkpoint, not later history:9a5dd7c2 predeclaration plus this runtime/
test/docs/evidence checkpoint. This does not approve previous Codex corrections
or Claude's original B4. No archival range union inferred.

Baseline d21b4b fails all six pinned-invalid input arms and the injected output
fault while the four positive controls pass. f21383 passes all eleven after the
minimal live/shared input/output checks.4ea870 passes all seven earlier-source
activation cases and exact Go corpus regeneration; output checks the new result
pin when Exit changes epochs. No valid replay byte, price, owned-node refund,
save schema, applied migration body or fixture epoch changes.

Independent controls all compile and fail: live admission1ef64f fails only the
two live negative cases; shared inputa37745 fails four assertions as disclosed
above (two actual recorded-invalid admissions, two wrong typed errors on
purchases still defended); outputd767da fails only the injected post-state
fault. Each restores exact SHA before any next run;00b14a confirms final
sources dd5d94db42daa17d936d05c6ec916684e6b48e60d2dc4bfc77f62b6a64312c99
(reputation_intent.go) and
c54e596ef71226e37199d8d62ff974ec26e3e8339f49be50cbef538165bfa249
(founder_replay.go). No surviving guard removal or hidden compiler/control error.
The input driver stopped on its unexpected failure count after restoring;
neither oracle nor runtime was loosened to fit that prediction.

Final b11ee2..65ec56 verify-server-core passes vet and all non-harness packages
cold: production33.558s,save0.310s,reputation0.196s,kernel0.190s,transport13.434s.
Full Pitch package ran cold0.330s; its later separate subtarget is cached.
Formula/API generation is byte-unchanged, import boundaries pass.
2a6589..3489ae passes zero-error/warning typecheck,213-module build,
7416 client tests/134 existing skips, shell14/8/22 and13 CI-topology negatives.
These are not real Postgres, browser, harness, full verify-client or hosted CI;
historical RP-131/50a3a514 RED is not promoted. All handles terminal before
checkpoint edits. a7e020/a2a2ee confirm migration/baseline, original and earlier
Reputation corpora, TS replay and generated contracts remain unchanged.

Kernel0.3.155→0.3.156 in all three identity files accompanies the real invalid
state acceptance change. Bare artifact-free codecs remain structural; this
fix covers the command boundaries named here, not every repository reader/
writer or a real DB transaction. RP-250 is locally corrected with designated
review pending; full R1/R7/AC11 and original B4's seventeen-path review remain.
RP-251 records only static next-epoch mirror/append-only-ID observations; next
bounded tests must execute same-ID effect retune and prohibited removal under
accepted R1/OD7 before any defect/ruling or runtime change is claimed. No new
refund, balance/copy, schema, CI policy, checkbox completion, archive, push or
Docker cleanup authority. Full nine-tier/platform goal remains active.

## 2026-10-06 — predeclare owned-effect cross-epoch mirror counterfacts

RP-250 is committed at8ab5a2c4; tree clean, all verification handles terminal.
Its complete range501000ea..8ab5a2c4 awaits Claude, not promoted by this wave.
This next bounded wave takes only RP-251's current-epoch owned-effect mirror
observation under accepted R1/OD-7. Append-only-ID admission remains a separate
next population; do not silently permit node removal or invent a refund.

Fixture-only next artifacts keep every ID but retune p05 unlock from50000 to
60000 and40000, strictly inside the existing increasing ladder. Its new cost3
must NOT rewrite historical spent4. Owned known p05 plus a retired unknown ID
remain unchanged; earned11 stays11. A same-artifact control keeps50000.
Strict loading and actual hash derivation are mandatory; no parsed fault copy,
production balance/copy/epoch edit or public content mint.

Go population: three targets × actual live settleAndActivateFoundations and
public ApplyFounderLogged Exit arms = six cases. Each uses a valid old pinned
Founder and Company, checks next mirror, unchanged level/spent/owned/metadata,
next pinned admission, available7 and the next frozen bonus factor. Full encoded
Founder results from both arms must match after the separately expected Exit
history append. Baseline must demonstrate which retunes fail; an unrelated
fixture/setup/compiler failure cannot establish the candidate defect.

TS population: the same three targets through strict artifact loading and
actual public Founder Exit replay, using the previously Go-authored v21 target
Founder as a source fixture, not pretending the new effects are already a new
shared full-byte corpus. Check unchanged ownership/historical spent/accounting
and next mirror/pinned full encode→restore. A future Company replay/DB/default-
player population stays separately named, not inferred from Founder replay.

If the retunes fail, minimally re-derive the persisted mirror from next-tree
owned nodes at the new-run boundary, in Go live, Go Founder replay and TS
Company/Founder epoch-advance paths. Do not recompute spend from current costs,
change a current run's frozen row, repair bad input or bypass RP-250 admission.
Successful same-pin and existing canonical replay bytes must remain identical.
Kernel bump in all three identities belongs to this real behavior correction.

Controls: independently omit live Go rebinding, Go Founder rebinding and TS
Founder rebinding; the two affected retunes must fail and identity control pass.
All controls must compile, become terminal and restore exact source SHA before
the next run. The TS Company path needs its own executed population/control
before any Company coverage claim; include it only with a further predeclaration.
Final cold server-core/vet and client/type/build/boundaries/topology; existing
corpora byte-unchanged. No full CI, DB, R1/OD-7/B4/AC11/RFC/release claim or
checkbox/archive/push/cleanup authority. New complete span after8ab5a2c4 needs
Claude independently. Earlier path/glob scan misses were read-only setup errors,
not experimental results; required files are now located via rg.

Baseline1ffcf0..b4fb21 compiles and fails all four Go retune arms; both unchanged
arms pass. f8b706..307e19 typecheck is clean, then the two TS Founder retunes
fail specifically at the result-pinned mirror restore. Unchanged Founder and
all previous tests pass:7417pass/2fail/134existing skips. RP-251's mirror defect
is now executed, not static; no source mutation/fix has yet occurred.

Further predeclaration before touching TS Company advance: use the existing
Go-authored scripted-first Company Exit fixture with the same three target
mirrors, strictly loaded actual next artifacts, unchanged IDs and historical
spent4/earned11/known p05 plus retired unknown ownership. Execute the real TS
public Company Exit and require next Founder accounting plus run_started's
factor1.0055/1.0066/1.0044. This is three more cases, not a new shared full-byte
cross-runtime corpus or DB commit. Independently omit only TS Company mirror
rebinding after the fix; both retunes must fail and unchanged pass, then exact
restore. No current-run frozen-contribution mutation is authorized or claimed.

Company baseline3ed799..6789f5 also compiles/typechecks cleanly and fails both
retunes specifically in next-pinned Founder carry parsing. Both unchanged new
cases and all previous tests pass:7418pass/4fail/134skips. Alongside the four
Go failures, all eight retuned arm counterfacts fire, four identity arms pass.
Only the existing R1/OD-7 mirror behavior is corrected below; no ID-removal
guard, refund policy, fixture epoch mint or spent-cost reinterpretation.

## 2026-10-06 — owned-effect mirror correction and complete range

**Review by:** Codex (self/first-filter on new runtime/tests/docs/records).
**Recorded by:** Codex.
**Designated reviewer:** Claude, pending.
**Complete new range:** `8ab5a2c4..HEAD`, where HEAD is the commit containing
this checkpoint, not future history:4e11f890 predeclaration plus this runtime/
test/docs/evidence checkpoint. RP-250 and previous spans remain independently
pending; this is not the complete original B4 review or an archival verdict.

Executed baselineb4fb21 fails four Go retunes,307e19 the two TS Founder
retunes,6789f5 additionally both TS Company retunes. Unchanged controls pass;
all failures are specific next-pinned mirror errors, no fixture/compiler failure.
After minimal correction,a8b934 passes cold selected Reputation0.454s and
312505 all7422 client tests/134existing skips with clean typecheck.

New population: higher60000/lower40000 versus unchanged50000 unlock, known
p05 plus retired unknown ownership, historical spent4 (next price3), earned11
and available7. Go live and public Founder replay each pass all three targets,
next-pinned full encode→restore and frozen factors1.0055/1.0066/1.0044; full
encoded Founder bytes agree after the separately expected history append.
TS Founder uses an earlier Go-authored state fixture and verifies next-pinned
full encode→restore/accounting/history. TS Company uses the existing actual
scripted-first Exit fixture and verifies carry/accounting and run_started's
new bonus factor. This is twelve semantic cases, not a new shared all-retune
canonical byte corpus or a real DB/service commit/default-player witness.

Four independent corrections omitted, each compiling and discriminating only
its own two retunes while unchanged and other-arm controls pass:

- Go live28527c..173bac: two live failures, both replay retunes pass;
- Go Founder89a451..c5430a: two replay failures, both live retunes pass;
- TS Founder863b30..769312: two Founder failures,7420pass/134skip;
- TS Company870844..a4a9ae: two Company failures,7420pass/134skip.

Mandatory SHA comparison after every restore passes (a2720d,5f585f,e1f427,
0f23a6). Final exact sources:

- foundations.go: `6c36f9cc3dba979e92baf8ea1556bae646cf9a3cffa299dd64fecb6be675c575`;
- founder_replay.go: `bcf482e3fab2bd4b8d154ad6c8eb214ceb465da299b03582829fd44e8827e7f5`;
- TS replay: `def54f36d6aa4b0da93bd45ea0354198e83db73bd08fc9014eb5da3705fa44a0`.

No survivor, leftover mutation, concurrent edit or acceptance loosening.

Final597627..6d95ee verify-server-core passes vet/all non-harness packages
cold: production35.321s,save0.265s,reputation0.202s,kernel0.169s,transport13.326s.
Full Pitch ran cold0.309s; later separate subtarget is cached. Generated formula/
API bytes remain unchanged. ff5a27..f88fb4 passes zero-error/warning typecheck,
213-module build,7422tests/134existing skips, shell14/8/22 and13 topology controls.
Full cb985f..6ea9e8 verify-client re-executes these first checks then fails
historical RP-131/50a3a514 kernel-history enforcement; no green CI claim.
Separatee4d880..b2efbf passes remaining combat/meter/achievement/cosmetic,
payment/copy/content-manifest checks (22cosmetic negatives,6payment negatives,
2near misses;657copy keys/610existing orphan warnings).7d81e9 confirms all
previous corpora and generated contracts unchanged. A redundant poll of already
completed39915 returned Unknown process id after its recorded exit0; that is a
tool bookkeeping miss, not a server failure. Every handle terminal before edits.

Kernel0.3.156→0.3.157 in all three identity files, no applied migration body,
schema, actual balance/epoch/copy, current-run frozen-row or CI policy changes.
RP-251 mirror part is locally corrected, designated review pending. Paired-
epoch append-only-ID admission remains static/unexecuted; previous positive
retirement fixtures prove mechanical fallback only, not removal authority.
Next predeclare actual removal/identity/append tests under OD-7, then remaining
pinned readers/writers and full B4. No full R1/OD-7/AC11/RFC/DB/CI/release,
checkbox flip, archive, push, cleanup consent or new DB workload. Full goal active.

## 2026-10-06 — predeclare append-only Reputation epoch admission

Previous goal turn made concrete progress at cf458b2b (RP-250/251 mirror fixes).
Grounding now confirms clean main, ahead64, no new designated verdict or live
verification handle. AGENTS/process/index and accepted OD-7 rechecked; scope is
the existing append-only-node contract, not a new refund or product decision.

Negative population: independently remove each of the nine fixture node IDs,
strictly reload the next artifact and derive its actual hash. Repair only fixture
prerequisite links and the terminal unlock ladder so each candidate is valid
in isolation; this must test a forbidden transition, not invalid tree syntax.
Founder owns only a retired unknown ID with historical spent4/earned11/mirror0,
so refusal cannot be vacuously supplied by an owned unlock mirror mismatch.
Go: linked CatalogBundle validity/resolver, direct live activation, public Founder
Exit replay, nine cases per arm. TS: linked bundle loader, and public Company/
Founder Exit on deliberately direct-linked bundles bypassing the convenience
loader, nine per arm. All reject without a receipt/event or state mutation.
Whole-artifact withdrawal is a separate linked-loader negative in both runtimes;
do not claim every runtime arm newly lacks a floor/withdrawal guard.

Positive linked controls: unchanged tree, legal same-ID effect/price retune,
append one valid starter node, initial tree activation, and both epochs inactive.
Existing epoch-retune/earlier-source/current corpus and retired-unknown fallback
populations must keep passing. No global chronology assumption: a standalone
historical artifact remains loadable; only a supplied previous→next pair can
enforce OD-7. Shared semantic census names all nine IDs and five controls;
it is not a full shared canonical byte corpus or a real Postgres/epoch mint.

If baseline confirms admission, add pure previous/next tree validators in Go/TS,
require them at Go linked bundle validity and direct live boundary, TS linked
loader plus both direct replay boundaries. Public Go replay already uses linked
bundle validity; it must not need an independently maintained second ID table.
The TS Founder verifier's existing catch must contain paired-loader refusal;
retain its current failure taxonomy, no unhandled rejection. Kernel bump in all
three identities accompanies this real acceptance change. No immutable migration,
save schema, owned-node refund, live balance/copy/epoch or CI-policy edit.

Existing RP-247 retirement positive contradicts OD-7. Preserve its two removed
IDs as a separately consumed forbidden-transition population in the shared
boundary table; keep the three idempotency/exact-cap positives byte-for-byte.
Update both consumers to assert refusal and no mutation, not silently drop the
row or weaken the guard. Pure unknown-ID fallback remains independently tested
by the earlier three starter-effect cases; it is not transition permission.

Controls after correction: independently remove Go linked guard (only loader/
public Founder negatives should lose admission), Go direct-live guard (only its
live negatives), TS linked guard (only loader negatives), TS Company guard and
TS Founder guard (only the corresponding deliberately direct-linked arm).
All compile, wait terminal, restore exact SHA before next run. Also omit one
semantic census ID: both consumers must reject that incomplete population,
then restore exact table bytes. Any surviving or setup-failing arm is retained.
Cold full server-core/vet and client/type/build/boundaries/topology, unchanged
canonical corpora, full client composite and remaining guards follow. No fresh
Docker workload or unapproved cleanup on the existing capacity hold. New range
after cf458b2b awaits Claude independently of all prior spans; no full R1/OD-7/
B4/AC11/RFC/CI/DB/player/release or checkbox/archive/push promotion. Next full
original B4 review and remaining pinned consumers stay on the queue.

Baseline5ad78b..53269a compiles and fires all28 Go negative cases (nine IDs ×
three arms plus whole-tree linked withdrawal); all five legal linked controls
pass. 11e392..7b934b typecheck is clean, then all28 TS negatives fail because
linking throws nothing or replay actually resolves applied. Six new positives
(census plus five linked controls) and all earlier tests pass:7428pass/28fail/
134existing skips. The tool truncates7937 tokens of repetitive applied-object
diagnostics; every failed population title and the final denominator are retained.
This is admission evidence, not a setup/compiler failure or real DB exploit.

Instrument refinement before controls: replace TS rejects.toThrow's enormous
resolved-object diagnostic with a small explicit caught-error assertion, keeping
the same refusal property and additionally requiring RangeError. No assertion
is removed; before fixing, all nine admissions in each replay arm must still
fail on missing error. Set next price3 in the TS positive retune too, matching
the predeclared legal effect/price population. Retain the first baseline above.

Compact baseline a48cec..d07c39 reproduces the same28 TS failures without
truncation: undefined is not RangeError in both replay arms, linking throws
nothing;7428pass/28fail/134skips. Initial corrected runs:
26720/759f0b passes selected Go Reputation cold (production
0.443s),53622/613270 passes clean types and7456tests/134existing skips.

Instrument repair before omission controls: the new Go retirement refusal was
initially passed the old fixture next hash. Supply the actual removed-tree hash
in resolved inputs instead, so a mismatched hash cannot supply a vacuous refusal.
No threshold/assertion weakened; full final cold population below covers the
repair. One combined apply_patch was rejected for duplicate operations on the
same file before writing; merged hunks succeeded. Read-only path misses for
CURRENT-STATE and migration corpora were corrected via rg --files; no false
unchanged-file claim rests on the missing paths.

Five independent compiled omission controls, every handle terminal before edits:
- Go linked admission removed: ef3fb7..7ec466 fires19 new cases (nine linked,
  nine public Founder, one withdrawal). Starter retirement also fires on the
  public error class: live still refuses with ErrInvalidEngineState instead of
  ErrInvalidReplayInputs. This twentieth failure is defense-in-depth/taxonomy,
  not a twentieth admitted transition. d35aff confirms exact source restoration.
- Go live admission removed: cb4bd1..fb2a29 fires all9 direct-live negatives;
  linking/Founder and all legal controls remain green. 7d5e17 exact restore.
- TS linked admission removed: df2379..fd16f0 fires10 new linked/withdrawal
  cases plus the retained retirement row:7445pass/11fail/134skip, types clean.
  The first transition to the Company omission restored the line but omitted
  the SHA check before that next probe. This bookkeeping miss is retained:
  after all other probes, repeat from exact8a55d0/d61224 restoration as
  8586d6..b96ce9, same11fail, then42d13a confirms exact SHA again.
- TS Company admission removed:3b320e..f6a493 fires9 direct Company cases plus
  retirement's actual applied result:7446pass/10fail/134skip. Typecheck clean;
  tool truncates5128tokens of the retirement applied-object diagnostic, while
  all failed titles/counts and the applied refusal failure are visible. 8a55d0
  confirms exact restore; no unseen raw diagnostic is credited as evidence.
- TS Founder admission removed:84a8f4..3faeb4 fires all9 direct Founder cases,
  7447pass/9fail/134skip; other arms and legal controls green. d61224 exact restore.

Census control removes p05 only:934d58..8d8342 Go rejects incomplete population;
799040..b265bd TS fails8-versus9,7452pass/1fail/134skip. Its denominator shrinks
three because each removed ID generates three cases:7587 rather than7590.
This invalid population is not a full-run pass. 646e77 restores exact source/
census SHA: replayGo a2bf1bf6...,foundations e701fb66...,replayTS1790fb96...,
table3db3a6da.... No mutation survives into the final gates.

Final cold d8ad19..a4b78e server-core/vet passes (production34.448s,save0.270s,
transport13.274s; full Pitch0.296s cold, separate content alias cached). Formulas
and generated API unchanged. b008e7..c88146 full verify-client passes types,
build213modules,7456tests/134existing skips and14/8/22shell/UI boundaries, then
fails historical pushed50a3a514 under RP-131. It is RED, not replaced by separate
greens. 473483..af84b2 separately passes topology13negative fixtures, remaining
combat/meter/achievement/cosmetic boundaries (22cosmetic negatives), payment
6negatives/2near misses,657copykeys/610existing orphan warnings,content-manifest.
Canonical migration baseline and both Go-authored Reputation corpus hashes are
unchanged (12c1f6,d4b9ff); generated contract diff b908fe is empty. 879e35 confirms
three legal starter expectation rows unchanged. Every verification handle terminal.

Review by: Codex (implementer first filter). Recorded by: Codex. Source/diff,
population, executed baselines and controls reviewed; NOT designated approval.
Complete new span starts after cf458b2b, includes155f6c17 and the implementation/
records checkpoint, and needs Claude against its literal committed tip. Kernel
0.3.157→0.3.158 in all three identities accompanies real admission narrowing.
No immutable migration, schema, actual balance/epoch/copy or CI-policy change.
The former retirement positive is corrected under existing OD-7, not a refund
ruling; three legal rows and independent unknown-owned fallback remain.
No global chronology, real Postgres/browser/harness/CI, full R1/OD-7/B4/AC11,
checkbox flip, mint, archival, push or Docker cleanup. Next full original17-path
B4 review and remaining pinned consumers. Full nine-tier/platform goal active.

## 2026-10-06 — original B4 review and event-admission predeclaration

Previous goal turn made concrete progress at822774df (OD-7 admission). Current
main is clean, ahead66, no live handle/new Claude verdict. AGENTS/process,
accepted Reputation RFC and binding vision/tech rechecked. Review scope is the
original Claude range541da96e^..541da96e:17paths, including all code/test/docs/
record/migration/kernel changes and the Go-authored purchase corpus. Its20
purchase rows and two bundle objects are unchanged at current HEAD (321060),
while later Exit supplements and Codex corrections are separate authority.
Initial long corpus projection and combined read output truncated; re-read
authority/code and compactly project every original row instead of claiming
unseen output reviewed. A read-only Exit projection assumed nullable Founder
rows non-null and threw; it is not runtime evidence or a census claim.

R5 strict event validation appears weaker than its claim (RP-252): plain struct
decoding defaults absent/null numbers, admits case aliases/duplicates, and adds
unbounded cost to spent. First prove/refute at validateIntentDecision, the entry
called before event writes by ordinary and Exit persistence paths. Do not infer
that a malformed player request can manufacture such an event.

Predeclared negative population: all eight required payload fields independently
omitted, null, case-aliased and duplicated (32); lower/upper numeric domains for
cost, level, spent-before, spent-after and unlock (10); wrong sum and sum-over-
earned relationships (2); signed-int64 wrap with valid earned level and negative
earned level (2); unknown kind/source, invalid mechanical ID and extra key (4);
null/array/empty-object/trailing-object/trailing-garbage shapes (5). Total55,
through the actual decision validator with a valid applied receipt/envelope.
Six valid controls cross both kind/source enums and exact-safe arithmetic ends;
all eleven original direct-purchase corpus events must remain admitted as well.
Use exact typed ErrInvalidStream on every refusal; no test SQL/DB evidence claim.

Run unmodified existing Go purchase/replay and client populations cold, then the
new55case population against unchanged production. If confirmed, implement a
Reputation-only exact/non-null/duplicate-free payload decoder and checked exact-
safe domains/relationship. Do not change general JSON decoding or other event
arms, payload schema/enums, producer bytes, immutable migration00075, any real
balance/epoch/copy or CI policy. Kernel bump all identities for admission change.
Demonstrate compiling independent omissions of exact-shape, numeric-domain and
relationship guards; assert affected refusals fail and valid controls still pass,
then restore exact SHA before each next run. Cold server-core/vet, full client
composite and remaining separate guards follow. No edits with live checks.

Separate verified review finding RP-253: original AC3 Postgres witness records
only applied and owned decisions, not every R5 rejection row. Current additions
cover frozen rows/Exit plans, not the missing direct taxonomy census. Preserve
Claude's historical Postgres evidence and invalid migration probe; no fresh DB
or all-row persistence verdict without an actual declared Postgres run. Existing
capacity/cleanup hold persists. This event repair cannot close RP-253, AC3/4/8,
full B4 or the career. Record the original designated review only for its exact
range; any new Codex correction requires Claude independently. No boxes/archive/
mint/push/owner-copy changes. Full nine-tier/platform goal remains active.

## 2026-10-06 — designated original B4 review: CHANGES REQUIRED

Review by: Codex (designated cross-party reviewer of Claude's original batch).
Recorded by: Codex. Exact reviewed range:541da96e^..541da96e, all17paths,
3323insertions/7deletions. Verdict: CHANGES REQUIRED, not full B4/AC3/4 approval.
This verdict cannot approve any subsequent Codex correction or uncovered range.

Diff inspection covered purchase evaluator/resolver/dispatch, shared replay arm,
TS parser/evaluator/replay, payload validation/enum, immutable00075 Up/Down,
all unit/DB/TS tests, docs/log and the three consistent kernel identities111.
Every original corpus row was compactly inspected after the first long projection
truncated:20rows/two bundles unchanged at current HEAD321060. Actual cold Go
reproduction71e74e..f0e7f5 passes the generated corpus and tamper tests; both
Reputation integration tests explicitly SKIP without TEST_DATABASE_URL. Clean
TS/types b2274e..5b1f88 passes7456/134, preserving the original20-row comparisons.
These executions are at current HEAD with later dependencies, not falsely
labelled a checkout/reproduction of the original111 kernel.

Finding A / RP-252: the original event validator (still unchanged in that arm at
baseline) admits22of55malformed decisions. c4f1df..c073ea reports actual nil
errors for all eight case aliases/eight duplicate fields, missing/null spent-
before/unlock (four), and two signed-overflow cases. Other33negatives refuse;
six valid controls and eleven original direct producer events pass. This is a
real strict-event admission defect at validateIntentDecision, not a demonstrated
malformed-player-request exploit or fresh SQL commit.

Finding B / RP-253: original real-Postgres witness records applied and owned
only, while AC3 demands every R5 rejection row. Later test additions do not
supply that direct taxonomy census. Claude's historical DB/retry/history pass
remains bounded evidence; its migration severing probe was invalid because the
shared database already applied75. No current all-row DB verdict is invented.

Finding C / RP-254: original Go case type/runner and TS consumer compare
state/receipt/events but never explicitly assert purchase result constants hash
against the shared Go-authored bundle pin, as R8 requires. This is an observed
oracle-scope gap, not yet an executed wrong-hash mutation. Next bounded evidence
can use the existing corpus pin without inventing a new schema or epoch.

Historical R7 nuance: B4's N/A-constraint statement is false, as Claude already
corrected in its later B6 log (b767dc):57/62/69 have founder_log_multistream_source_shape,
and78 extends it. Do not repeat the false N/A as a current claim or rewrite its
historical entry. Later constraint/Exit implementations and their review remain
outside this original B4 verdict. R1 pinned-command repair RP-250 and other
subsequent Codex work likewise remain separately pending Claude, not absorbed.

## 2026-10-06 — RP-252 event repair evidence and implementer first filter

Review by: Codex (implementer first filter, NOT designated review of this fix).
Recorded by: Codex. New complete span starts after822774df, includesef37e83c and
this implementation/records checkpoint; Claude must review its literal committed
tip independently of the original verdict and every previous correction.

The Reputation-only token check requires eight non-null/exact-case/duplicate-free
fields, then existing typed decode/closed enums and bounded accounting. Every
integer's exact-safe domain is checked before adding cost to spent, preventing
signed wrap from satisfying equality. No general decoder or other event arm,
producer payload, schema or applied migration is changed. Three kernel identities
158→159 mark this real admission narrowing. Initial patch's temporary final
false disjunct was removed before any verification; no result depends on it.

064342..7e1acb passes all55negatives, six legal controls and eleven producer
events (save0.298s), plus full selected Reputation production0.461s. The six
controls prove schema/accounting admission, not catalog-bound node effect/price
consistency; the eleven actual producer fixtures are a separate population. Independent
compiling omissions, every handle terminal before edits:
- exact fields: a60cd2..dbfc73 fires20key ambiguity/default cases; six/eleven
  positives remain passing.93a5c1 exact source restoration.
- numeric domains:85c04c..b4056e fires7cases (cost0, excessive level, negative
  spent-before, both unlock bounds and two overflows). Other domains remain
  protected by arithmetic; no claim every redundant bound independently fires.
  d0e2b4 exact restoration.
- sum/earned relationship:7aad27..b580be fires2cases.2e5b11 exact restoration:
  intent.go292d60335fe1575090beb2f50e7dc67abd6ecaf9ee7aeee74883f04efc97ab3f;
  new helper801a6959ffd4b8be96631e850f5c6f91674110fb2d92bb24ae50d88b7dd14308.

Final99f24b..700c67 server-core/vet passes cold (production34.041s,save0.204s,
transport13.320s; full Pitch0.298s cold, separate content alias cached). Generated
formulas/API diffs stay empty.69e77c..d4ee2d full client composite passes clean
types, build213modules,7456tests/134existing skips,14/8/22shell/UI boundaries,
then is RED at historical pushed50a3a514 (RP-131).0415bd..049747 separately
passes topology13negatives, remaining package boundaries22cosmetic negatives,
payment6negatives/2near misses,657copykeys/610existing orphan warnings and
content-manifest. Original Reputation/migration corpora and all applied migration
files unchanged (65488d,2c0eb8); no replay production bytes change in TS.

Capacity revalidated read-only:e8cd8c lists both healthy owned Postgres services
and the existing tabiya builder; afdb0d shows100%/39784KiB free.9c46bc reports
143volumes/39.36GB total,36.78GB reclaimable; that broad pool is NOT the earlier
narrow cleanup request, not deletion authority. No cleanup, SQL mutation,
service restart or new Docker/DB workload. Every gate/probe handle terminal
before records. Next RP-254 hash evidence/remaining pinned readers/closed replay
inputs, with RP-253 actual Postgres taxonomy still required. Full B4/AC3/4/R8/
career, mint/H4/owner/author/privacy/accessibility/deployment/CI and nine-tier
1.0 obligations stay open. No checkbox flip, archival, push or goal completion.

## 2026-10-06 — predeclaration: RP-254 purchase replay result pins

Authority: accepted R8 and the original B4 Finding C. Scope is test-only:
assert the direct purchase result hash against the existing Go-authored bundle
pin, plus the three existing paired Founder Exit results against their recorded
next bundle. No fixture/schema, runtime, kernel, balance, epoch or copy change.

Population: all20 direct rows (11applied,9rejected), including the automatic
Fiscal sweep in the chain, and all3 paired Founder arms among5 R6 Exit rows.
Go generation and TS consumption must census these populations and compare
hash strings exactly, in addition to unchanged state/receipt/ordered events.
The two rejected Company Exit rows have no Founder result; they are not silently
counted as hash observations. Existing corpus bytes must remain identical.

Controls, declared before execution: corrupt the actual Go inactive purchase
return pin and TS applied-purchase return pin, retaining state/receipt/events;
the selected consumers must fail specifically on hash mismatch. With the new
purchase assertion temporarily omitted, those unchanged old comparisons must
pass the same fault (otherwise disclose defense-in-depth, not oracle proof).
For R6, inject the old pin into Go's returned Founder transition at the test
boundary, and replace TS Exit's result pin with its input pin; the two activation
rows must fail while the same-pin Exit remains valid. The Go injection proves
the oracle, not a production defect or a bypass of its output-state guard.
Each probe runs cold, terminates before any edit, and restores exact source SHA.

Success: all23 assertions per runtime pass honestly, the declared faults fail,
old direct comparison controls survive their hash-only fault, corpus bytes and
production sources restore unchanged, cold server-core/vet and full client
checks executed. Historical RP-131 is reported RED, never waived. No actual
Postgres workload under the capacity hold; RP-253 remains required. First-filter
only for this new Codex span afterc355fb7d; Claude designated review required.
No whole R8/B4/AC3/4, reader/writer census, career/player/release, archival,
checkbox, push, cleanup or goal-completion claim follows.

Supplemental population controls, before execution: independently remove one
direct row and one paired Founder observation from the in-memory test census
in each runtime. Each must fail its population check; these are test-boundary
denominator probes, not corrupt artifact/runtime claims. Original bytes restored.

## 2026-10-06 — RP-254 result-pin proof and implementer first filter

Review by: Codex (implementer first filter, not designated review).
Recorded by: Codex. Reviewed prefix c355fb7d..eddddc61 plus the complete working
test/docs/ledger/board/queue/checkpoint diff committed with this entry. Claude
must independently cite the full literal c355fb7d..new-checkpoint range; no
original Claude-batch verdict or prior correction approval covers this work.

The two consumers now compare direct purchase result hashes with the existing
Go-authored bundle pin and paired Founder Exit result hashes with the recorded
next pin. The census is20direct (11applied/9rejected),5Company Exit rows and3
paired Founder results:23 hash observations per runtime. The two rejected
Company Exits have no Founder result and are not counted as hash observations.
No new hash authority, fixture field, runtime semantics or epoch is invented.

Honest selected baseline: TS00c8ad..636469 passes35tests. Initial Go2cb86f/352359
selected no tests because Make consumed the single-dollar regex suffix; it is
invalid evidence. Corrected d1f3ad..d5e883 runs TestReputationPurchaseCorpus
verbosely and passes cold. No gate claim relies on the empty selection.

Declared wrong-hash controls, each handle terminal before edits:

- Go inactive purchase return:025997..29df02 fails at the new result-pin
  comparison with the deliberately all-zero hash. Omitting only that assertion
  while retaining the same runtime fault (7de712..5773c4) passes the generated
  corpus byte comparison. The existing state/receipt/events cannot see this fault.
- TS applied purchase return:bf4782..a3be94 fails all11applied rows at result-pin
  equality; nine rejections and two selected metadata/tamper tests pass. Omitting
  only the new assertion (f9286a) passes22selected tests/13selection skips with
  the same bad return hash. These are selection skips, not missing populations.
- Go paired Exit: inject the old pin after the actual runtime result. Initial
  b7e87e..e0e9cf stops at the first activation case. Replacing only the hash
  assertions' fail-fast diagnostic with Errorf allows the whole population to
  execute:47eb33..80c7cc fails both activation rows, not the same-pin Exit.
  This final diagnostic is retained. The injection proves the test oracle,
  not a runtime defect or a bypass of the existing Go output-state guard.
- TS paired Exit runtime return:08376a..078e8c fails exactly the two activation
  rows at hash equality; the other five selected tests pass. Input/result pins
  coincide for the ordinary plan Exit, so that positive remains green.

Supplemental denominator controls: remove the last direct row only:
ffffa6..93b142 Go reports19/10/9/5/3;0f6d9a..fd66ce TS reports19instead of20.
Remove the last Founder pairing only:44e09c..99c5a4 Go reports20/11/9/5/2;
ba1428..21feb5 TS reports2instead of3. Both consumers fail each independently.
These are in-memory test-boundary probes, not corrupt corpus artifacts.

Exact final restoration d14077:

- Go runtime reputation_intent.go:
  dd5d94db42daa17d936d05c6ec916684e6b48e60d2dc4bfc77f62b6a64312c99.
- TS runtime replay.ts:
  1790fb96006b721372a3cb5a596225708754e271ab972bb7631851652cb54ee7.
- Final Go test:
  4950c995d0305ba7c70d06114f1acd9d89ef524e510b1f5405141ae05cb63616.
- Final TS test:
  159ce3a23aa70e186c46fc900936c687439cd24208624b49d593e4100959cba7.
- Unchanged Reputation corpus:
  f9b129e36af5b536f67c7eddb8d6088c76ce5cb0b170cfeec6a5172e8a009782.

Final cold fb709f..fd7ff3 make verify-server-core CORE_TEST_COUNT=1 exits0:
vet and core pass (production40.217s,save0.263s,transport13.303s), formulas/API
regeneration has no diff. The main Pitch package was cold0.293s; its separate
content alias is cached and is not called another cold population. 2c7c51..fddfb2
full verify-client passes types0errors/0warnings, build213modules,7457tests/
134existing skips and14/8/22shell/UI boundaries, then exits2 at historical
pushed50a3a514 (RP-131). Full CI is RED, not waived or called green.
99a6d1..bc4541 separately passes topology13negatives, remaining boundaries
(cosmetic22negatives), payment6negatives/2near misses, copy657keys/610existing
orphan warnings, and deployment content-manifest. All handles are terminal
before records. An unnecessary read-only ps diagnostic was sandbox-refused;
no escalation or claim depends on it. a2a08f and final diff inspection confirm
runtime/kernel/corpora/applied migrations/generated contracts remain unchanged.

RP-254 is locally covered, awaiting Claude. RP-253 still needs the complete
persisted rejection taxonomy and an actual declared Postgres run; capacity/
cleanup hold unchanged, no Docker workload or deletion attempted. R8 history/
verifier and full B4/career/player/RFC acceptance remain open. No boxes, archive,
mint, push, owner-copy or release promotion. Full nine-tier/platform goal active.

Next bounded accepted work is RP-255: static inspection cd27fb/3f77e7 finds
the R5 Go resolved arm uses a struct decoder with unknown-key refusal but no
explicit exact/non-null/duplicate-key census, while TS uses exactKeys plus
typed fields. Do not infer actual admission from inspection. Predeclare raw-
wire and corresponding parsed-object populations before execution; TS receives
objects, so duplicate JSON keys already lost during parsing are not a TS wire
rejection claim. No generic decoder or other intent fix is authorized by this
observation. Remaining pinned-reader/writer, owner/author/mint/H4/rights/privacy/
accessibility/deployment/CI and full Transcendence obligations persist.

## 2026-10-06 — predeclaration: RP-255 frozen purchase-input admission

Previous goal turn made concrete progress at4d690f29. Grounding923934 confirms
clean main ahead70, no new Claude verdict or live verification/probe handle.
Accepted R5 freezes six resolved fields and R8 requires parity; save/founderlog.go
explicitly leaves the exact closed resolved union to feature packages. This
authorizes a bounded Reputation-only replay admission check, not a generic
decoder policy change or a new public request/API contract.

Population, before measuring: all18 original corpus cases whose resolved kind
is purchase_reputation_node (11applied/7ordinary rejections). The two invalid-
request cases use the separate invalid arm and are explicitly excluded, not
counted as purchase-arm observations. Add one zero-earned control derived from
the existing cost-one-over case by setting earned level to0 in the valid pre/post
Founder state and frozen inputs; its unchanged ordinary unaffordable receipt,
empty events and resulting input pin must match. This supplies a real zero-level
default counterexample, not a population chosen to avoid it.

19controls ×6fields ×4single-field mutations =456Go raw-wire negatives:
missing, null, upper-case alias and duplicate-identical key. All must fail with
ErrInvalidReplayInputs, no receipt/events and exact pre-state restoration. TS
receives parsed objects:342missing/null/alias mutations must refuse;114duplicate
raw JSON mutations normalize during JSON.parse and must reproduce the original
full state/receipt/ordered events/result pin. These114 are parsing-limit controls,
NOT TS duplicate-wire refusals. All19 unmodified controls byte-match separately.

Author a shared fixture from Go using the existing corpus's source SHA and
explicit control IDs/raw resolved strings. Existing replay/migration corpora
remain unchanged. A root authoring target may use the existing explicit
-update-replay-fixture flag; ordinary checks compare, never regenerate. Source
SHA, exact population/cartesian census and generated row equality are required;
no hidden duplicate-key collapse in the stored raw strings. Before admission
claims, cold Go and TS execute the controls and negatives on unchanged runtime.

If baseline admits malformed frozen inputs, minimally enforce the existing six
fields only in the Go Reputation resolved arm. No other arm/envelope/decoder,
schema, migration, balance, epoch, provider or player copy change; a real runtime
admission narrowing bumps all three kernel identities159→160. TS need not change
if its object-level refusal already matches. Success requires all declared
negatives and controls, full existing corpus parity, exact rollback, independent
compiling Go gate omission with its formerly admitted negatives firing, and a
TS object-key gate omission demonstrating its alias fixtures fail. Census/source
SHA/row-corruption negatives must fail both fixture consumers. Restore exact
SHAs after each probe, and terminate all handles before editing any source or
record. Cold full server-core/vet, types/client/build and independent boundary/
topology checks follow; historical RP-131 remains honestly RED.

This is replay admission evidence, not an executed malformed player request or
fresh Postgres commit. RP-253 actual persisted taxonomy and the Docker capacity/
cleanup hold remain. New complete span after4d690f29 needs Claude independently;
no checkbox, archive, mint, push, cleanup, full B4/R8/career/RFC/CI/player/release
or nine-tier-goal promotion. Remaining pinned reader/writer audit still required.

Population refinement before any experiment or fixture authoring: a renamed
key may remain defended by TS's independent typed-field checks even without
exactKeys. Add alias_extra as a fifth mutation: retain the canonical field and
also supply its uppercase alias with the identical value. This specifically
exercises exact-key admission without relying on a different error class.
Final population supersedes the four-mutation counts above:19×6×5=570Go
raw negatives;456TS missing/null/alias/alias_extra negatives,114parsed-duplicate
positive controls and19unmodified controls. No error-class-only TS rejection
claim: gate omission should admit the114extra aliases; other rows can stay
defended and must be disclosed. The added fields change neither expected state
nor receipt/event/hash. All other scope, controls and held obligations unchanged.

## 2026-10-06 — RP-255 correction, controls and implementer first filter

Review by: Codex (implementer first filter, not designated review).
Recorded by: Codex. Reviewed4d690f29..4ca6dc7d and the complete current
implementation/test/fixture/docs/ledger/queue/board checkpoint diff. Claude must
review the full literal span after4d690f29 through this checkpoint, including
761462fc and4ca6dc7d. No previous designated verdict or self review covers it.

The original token stream is now checked for the six R5 keys, all required,
non-null, exact-case and duplicate-free, before its existing strict typed decode
and recomputation. This is a localized Go purchase-arm admission narrowing;
generic replay/envelope decoding, producers, other arms, schemas and applied
migrations are untouched. Existing input shapes and state/receipt/event/pin
bytes stay identical. Three kernel identities159→160 record the real narrowing.
The TS runtime remains byte-identical; its parsed-object checks already refuse.

The shared fixture uses19controls:18original purchase-arm cases and one valid
zero-earned unaffordable profile. Two original invalid-request arms are outside
this union and explicitly excluded. Each control has six fields × five mutations:
570raw Go negatives,456TS malformed-object negatives and114TS normalized-
duplicate positives. Nineteen unmodified controls compare full canonical state,
receipt, ordered events and result pin, including the automatic Fiscal sweep.
The Go constructor authors raw JSON strings without erasing duplicates. The TS
census independently reconstructs every raw row and verifies actual source bytes
with browser-compatible WebCrypto. Ordinary tests compare, never regenerate;
only the named root authoring target uses the existing explicit update flag.

Instrument construction failure disclosed:2479d9..5054cb failed all19positive
controls at Founder canonical-command admission because the test supplied
indented fixture presentation bytes. The constructor correctly refused to write
the expectation. Canonical command normalization was corrected in the test
only (3b12c5); cad878..8d094f then authoritatively passes all19before authoring.
That initial setup failure is not replay-defect or negative-population evidence.
Earlier misspelled read-only fixture paths and truncated broad source projections
were corrected with rg discovery and targeted actual-file reads, not treated as
evidence that the missing files or projected claims existed.

Baseline e99fc2..134127 executes all570 on unchanged runtime and reports:
376admitted,194refused; admissions are114aliases,114extra aliases,114duplicates,
17missing and17null. All19positive controls pass. 94f566..bd694f executes590TS
tests successfully:456refusals,114duplicate normalization controls,19original
results and the census. JSON.parse has already erased duplicate keys before the
public TS API: those114 are never called TS raw-wire rejections. Neither result
is a malformed-player-request exploit or an actual corrupt SQL write.

Correction2156bf..9d1fc6 rejects all570 with ErrInvalidReplayInputs, no receipt/
events and exact complete pre-state rollback. All19controls still byte-match.
8dc904..2138c5 types pass0errors/0warnings. Independent compiling controls,
every handle terminal before source or record edits:

- Go omits only the new gate:2869bf..02b139 admits the same376/570 again with
  the original group counts;194remain defended. The unmodified controls pass.
- TS omits only existing R5 exactKeys:49738f..d2c5bb fails114extra-alias rows
  because the replay promise resolves instead of refusing;476other tests pass.
  The342other malformed objects stay defended by field/type/dispatch checks.
  Detailed repeated diagnostics were tool-output-truncated; final114/476count
  and actual resolved-promise diagnostic were visible, not a claimed full text
  dump. No TS product change is retained.206b48 verifies both runtime SHAs.
- Forge source hash:ad153f..93cea6 Go detects drift;244ce9..48827d TS compares
  zero hash with the actual f9b129e3 source digest and fails. de3e7f exact restore.
- Remove one case:4e6388..739fde Go detects drift;5a776a..08fe83 TS requires570
  and finds569. This has588selection skips, not missing runtime coverage.
  f42561 exact restore before the next distinct probe.
- Alter a raw row from p05 to p25 without changing its metadata:7e0b6c..b2230d
  Go detects drift; e77fd4 TS's independent expected raw row fails.1f8648 exact
  restore. No control regenerates the fixture around the fault.

Exact restored SHAs (b97290/206b48/1f8648/d04c4c/0092cf):

- corrected Go reputation_intent.go:
  11fe1b6a202f45b142acbf6de1375522dd21011fa108a04db79cb72a4fba8189;
- unchanged TS replay.ts:
  1790fb96006b721372a3cb5a596225708754e271ab972bb7631851652cb54ee7;
- new Go test:d46b218318b86a94c53e47b344f9c840efa4b5415cb3d68eb2ec9657c790eeaa;
- new TS test:ed0373a13d4ce5d8e2b4fa5868d5888b8ae194e4d1992b5acb5aa4a6b7337b4e;
- new shape fixture:e860f214a4df64c152d06ebdb2e1178c3a16e885ef15ff5153ba2b13c8675686;
- unchanged source corpus:
  f9b129e36af5b536f67c7eddb8d6088c76ce5cb0b170cfeec6a5172e8a009782.

Final f5070c..1b3426 cold make verify-server-core CORE_TEST_COUNT=1 exits0:
vet/core pass, production34.319s/save0.272s/transport13.288s, formulas/API no
diff. Main Pitch0.292s is cold; separate content alias is cached and is not a
second cold observation.4790c8..c9587c full verify-client passes clean types,
build213modules,8047tests/134existing skips and14/8/22shell/UI boundaries, then
exits2 at historical pushed50a3a514 (RP-131). No full CI green claim.008f1e..
5fa203 separately passes topology13negatives, remaining boundaries (cosmetic22),
payment6negatives/2near misses, copy657keys/610existing orphan warnings and
deployment content-manifest. d5efa4/81abf3/0092cf and final diff confirm old
replay/migration corpora, applied migrations, generated contracts and TS runtime
unchanged. New authoring target changes no CI workflow or test selection policy.

Supplemental native browser execution / RP-256:7fb609..d25885 runs the new file
on Darwin arm64 (df7f4c). Chromium/WebKit each pass590tests (1180total). Firefox
never executes: connection timeout60s, launch timeout180s, with plugin-container
sandbox-extension Operation not permitted AND SWGL framebuffer diagnostics.
The initial suggestion that Codex's execution sandbox alone explains it is not
established. After that exact handle exits2, an approved narrowly escalated root
make test-browser command selects only Firefox; b2191c..77b30d again executes
0tests and fails with the same two diagnostics. This is a verified changed-
environment check, not a blind retry on an observation timeout. No timeout,
browser config, assertions, security bypass or skipped engine is changed.
Both red attempts remain evidence; no native three-engine or Linux/hosted CI
promotion. Every handle is terminal before these records; no live probe remains.

Capacity rechecked read-only:c1b7d2 lists both healthy owned Postgres services
and the existing tabiya builder;906fe0 shows100%/39784KiB free. No prune, volume
delete, restart, new Docker workload or SQL mutation. Narrow unused-volume
cleanup request remains unanswered, not inferred from the goal. RP-253 still
requires retained complete direct taxonomy plus actual declared Postgres and
fired persistence controls; R8 history/verifier/career and remaining pinned
reader/writer audit stay required. Next prepare that accepted test population
without calling a host skip persistence evidence; native launch diagnosis is a
separate open environment obligation. Full nine tiers/platform/mint/H4/rights/
privacy/accessibility/operations/release remain, no boxes/archive/push or goal
completion. New correction and every preceding Codex span need Claude separately.

## 2026-10-06 — RP-253 persisted purchase taxonomy predeclaration

Scope: accepted R5/R8/AC3/AC4, test-only supplement after86c96035. Retain all
twenty pinned direct-purchase profiles (eleven applied/nine recorded rejections),
plus a tree-present unactivated v21 Founder and an overdue-Fiscal rejection.
The original corpus bytes/hash remain fixed. Each of the22 independent persisted
profiles must run through Service.Handle, then assert exact outcome/category/
detail, full rejected-state preservation, ordered committed events, one Founder
log/intent record/outbox entry, immutable Company state, identical retry, and
Founder-history verification. Applied accounting/ownership/unlock must match
the pinned post-state independently of the live resolver. Database-generated
timestamps are retained and replayed, never replaced by a fabricated wall clock.
Each profile also attempts stale CAS and changed-body idempotency conflicts:
44 unrecorded refusals must add no revision, event, log, intent or outbox row.

Preparation may execute the source SHA/name/outcome census and construct valid
profiles locally, including fired source-hash, missing-row and changed-outcome
controls. These do not prove persistence. Actual declared-compose Postgres
execution and compiling severing controls (requires gate, rejection log write,
rejection rollback, idempotency lookup) remain mandatory and pending until the
Docker capacity hold clears. Rechecked28d8db/199cbb/9c87ea: clean main ahead73,
two healthy Postgres services, Docker filesystem100%/39784KiB free. No deletion,
prune, restart, new workload or SQL mutation is authorized by this plan.

No runtime/balance/schema/migration/kernel/CI policy or player copy changes,
no expectation regeneration, box flip, verdict/archival/mint/push/release
promotion. Invalid instrumentation is corrected and disclosed, not counted as
a product failure. Host database skips stay explicit. No edits while a check
handle lives; independent Claude review must cover the complete new span.

Pre-authoring precision: the single outbox entry above means one receipt row;
applied event rows have their own outbox deliveries and remain required. Profile
Fiscal opened time is rebased to the measured preparation clock (the overdue
negative moves it back one AutoMS); actual database command timestamps govern
any sweep. Purchase state/receipt/event expectations still come from the pinned
source, with only intent/revision coordinates and independently calculated
Fiscal sweep fields adapted. No timing-dependent sweep is silently discarded.

Preparation finding before any SQL run:6e1ea3..a26acf compiles but profile
construction rejects the added tree-present/v21 pair at the pinned version
floor. This is an invalid instrument, not a product defect. All twenty original
profiles remain mandatory; the extra valid overdue-Fiscal rejection yields21
recorded profiles and42 unrecorded CAS/idempotency controls. The invalid v21/tree
pair is retained separately as a pinned-foundation refusal, never inserted by
bypassing policy. Strengthen the test store resolver to apply the actual bundle
foundation policy and use a pinned valid Company fixture, not the old thin
integration resolver's faction-only state policy. This corrects the predeclared
population's invalid premise; it waives no R5 row or SQL/negative-control gate.

## 2026-10-06 — RP-253 retained taxonomy and preparation first filter

Review by: Codex (implementer first filter, not designated review).
Recorded by: Codex. Reviewed `86c96035..2e9cc4c4` and the complete current
test/docs/ledger/plan/queue/board checkpoint diff. Claude must independently
review the literal full span after86c96035 through this checkpoint; it includes
5425e684, 373a6e1a and2e9cc4c4. No prior verdict covers the new work.

Retained `reputation_taxonomy_integration_test.go`: all twenty original source
cases (eleven applied/nine rejected), plus an overdue-Fiscal unknown-node
rejection. The original source SHA is fixed; its purchase state, complete
receipt and purchase-event expectations are not regenerated. Independent
intent/revision coordinates are adapted to each new stream. Fiscal opened
time is rebased, and the actual DB command timestamp governs sweeps; expected
Fiscal effects use the existing catalog sweep/wire helper, not an independent
new arithmetic proof. Rejection expectations are the complete pre-state,
including Fiscal, with no event. The full nine-node chain and exact-budget
controls remain in the population, not sampled away.

The prepared SQL population checks each actual Service.Handle result against
those expectations; one immutable Founder log/intent/receipt-outbox row;
applied revision and ordered event/outbox delivery counts; history verification;
identical receipt retry; no Company mutation; and42 unrecorded CAS/idempotency
conflicts with no further writes. The test store applies the pinned foundation
policy, not the thin integration resolver's faction-only validation. The Company
fixture is restored from the matching pinned Exit case and passes that policy.
Its presence does not claim an executed Company run or composed player career.

Initial instrument failure6e1ea3..a26acf: the extra v21/tree state violated the
pinned floor, preventing profile construction; three source controls passed,
but the purchase population did not run. Not a product rejection/result. The
predeclaration correction2e9cc4c4 retains the mismatch as a separate foundation
refusal;21 valid purchase profiles remain, no R5 row removed. Subsequent
db4e93..dfe7ba passes all21 locally plus that refusal and three census controls.

Compiling population-validator omission ef4a88..7dc1bb fails all three forged
source-hash/missing-row/changed-outcome controls with "forged taxonomy population
accepted"; all21 honest profiles and the floor refusal remain green. No runtime
mutation or expectation regeneration.104b98 confirms exact new-test SHA restored:
`cea42f5e7b7427f9e23c4d30323aa3838f7afccf9be06dd08b15ed4a3459b0d0`.
These are instrument controls, not the deferred SQL severing cases.

Final2ed3a1..358b82 `make verify-server-core CORE_TEST_COUNT=1` exits0:
vet/core cold; production34.609s, save0.297s, transport13.339s. Pitch0.316s is
cold in the core; the separate content alias is cached, not a second cold claim.
Formulas/API regeneration has no diff. Final9b2ef8..6b9747 focused cold verbose
run passes all25 preparation subtests (21 profiles,3 corruptions,1 invalid
pair refusal), and explicitly SKIPs the SQL test: "persisted taxonomy NOT
EXECUTED". Every handle is terminal before these records. No TS/browser/full
CI rerun or fresh SQL evidence; prior RP-131/RP-256 red limitations persist.

e18b92/836929 verify original purchase corpus SHA f9b129e3..., corrected Go
runtime11fe1b6a..., unchanged TS1790fb96..., kernel/schema/applied migrations
and CI workflows unchanged. Only the new test and documentation/tracking are
retained. The first combined tracking patch missed a full-line context and
failed atomically;451d4c/7c01c1 confirm no partial edits before the corrected
patch. No false completion was committed.

RP-253 remains OPEN for actual declared Postgres execution and four compiling
persistence severings (requires, rejection log, rollback, idempotency). Capacity
remains100%/39784KiB free with two healthy databases, read-only checked; no
cleanup or new workload. Test preparation may be independently reviewed, but
cannot close AC3/B4 or replace R8 career/run verifier/AC15, mint/H4 or any
nine-tier/platform/release obligation. Next safe accepted work: remaining
pinned reader/writer audit. No checkbox, archive, push, deletion or goal status
promotion.

## 2026-10-06 — R1/R3/R9 projection readers predeclaration

Clean HEAD4030c895, ahead77. Previous turn is progress: RP-253 tests retained,
not SQL completion. Read-only search traces writers/encoders/epoch paths,
run-frozen producers, server UI projection and TS snapshot consumer. Broad
output truncated; targeted actual sources were read. Store.LoadLatest performs
structural RestoreState but does not invoke StatePolicyValidator; production
writes do apply runtimeCatalogs.ValidateState. The UI preview checks Company
transition eligibility, not the Founder's checked Reputation mirror. These
observations do not authorize a generic Store policy change or prove a SQL
corruption. R3 frozen and starter producers already call the pinned validator;
current Company rate remains run-frozen. Harness and remaining replay/history
consumers still need their own complete audits.

Scope: RP-257 and RP-258 only, accepted R1/R3/R9. Before measuring, declare:

- Go7 corrupt active profiles: nil owned collection, duplicate, unsorted and
  nonmechanical owned IDs, empty ownership with a nonzero mirror, p05 ownership
  with zero mirror, and p05 ownership with a different nonzero mirror. Execute
  both the isolated arm and public transaction-local InitialGameUISnapshot:
  14 refusals, no output. Three consistent profiles (empty, p05, unknown historical
  id plus p05) remain legal in both paths; no node removal is authorized.
  Complete pinned source states/artifacts come from the unchanged R8 fixture;
  valid controls must reach the public projection before negatives are evidence.
- TS12 invalid string values per factor field and per public reader
  (parseFeatures and parseGameUISnapshot):48 refusals. Values are empty, NaN,
  Infinity, -Infinity,1,1e+0,01e0,1.0e0,0,0e0,-1e0,5e-1. Two numeric rather
  than string fields per reader supply4 already-defended controls. Three legal
  unit/nonunit/null-current profiles per reader supply6 positives. R3's factor
  is canonical and >=1; null current factor stays legal. Legacy optional arm,
  null/inactive behavior, schema/pins and owner copy are untouched.
- If reproduced, use existing pinned UnlockPPM derivation to refuse inconsistent
  ownership/mirrors locally; use the existing canonical Decimal parser plus
  >=1 factor domain in the TS Reputation arm. Independent compiling guard
  omissions must fire affected negatives; unchanged valid controls and exact
  source SHAs must remain. No silently normalized invalid input.

These presentation-only paths are outside kernel/affecting-paths.json; do not
bump the numeric kernel for a non-kernel change. No balance, applied migration,
API schema/pin, canonical corpus, generic save/decoder, CI policy or authored
copy edits. No source/record writes while any check/probe lives. No fresh DB or
browser/full native proof, whole R9/B7/AC12/career/1.0, box/archival/mint/push
claim. Capacity, RP-131 and RP-256 remain. Claude must independently review the
complete new span after4030c895; self filter cannot replace that gate.

## 2026-10-06 — R1/R3/R9 projection reader corrections executed

Authority336503b3; accepted presentation-admission scope only. Cold unchanged
Go baseline5e17c1 fails all14 negative observations, while all three valid
profiles pass both isolated/public transaction-local projection. TS's first
pnpm invocation never executes tests:3555c6 fails package-manager dependency
resolution on registry DNS. No install/network bypass; the existing root
`make test-client` lane executes normally. Baselinea36c2c fails all48 malformed
or below-one string refusals, with six positives, four already-defended numeric
inputs and census passing (8058 other tests/134 existing skips). Neither
baseline is a persisted corruption, public HTTP or player exploitation claim.

Go now requires non-null ownership and derives unlock ppm using the pinned
tree's existing sorted/mechanical/unique validation before comparing the mirror.
TS uses the existing canonical Decimal parser and R3's >=1 domain; null current
factor stays legal, and the client does not derive node eligibility or formula.
No normalization or generic Store policy change. Corrected94297e and048130
pass all17 Go subtests (six positive observations,14 refusals) and all8106
client tests with134 existing skips, including the59new tests.

Three independent compiling controls, no assertion changes:

- Nil-set guard omission40fe54..98c71a fails exactly2nil-owned observations;
  the other12 negatives and all valid profiles stay green.
- TS factor guard omissionf3fb45..913054 fails exactly48string refusals;
  numeric controls, six positives and census remain green (8058/134).
- Pinned derivation/mirror guard omissionb3301d..d3389b fails exactly12
  observations; nil-owned remains defended and all valid profiles stay green.

c85aa9 confirms exact corrected source/test SHAs restored:
Go projector92edef8bb8e92687e3d98f9f823951bc1480969f029430b2921af035c0451e17;
TS contractsf8c126d15d0e6e6ea71bd7ab73e68e21e2a2ce747a5ed550cf97ef08334e3fe2;
Go test8225abd76e0b6c0a7f0e9c315cf93c5c92d1928c07061e3647ed29707fce2324;
TS testac1eee94c9db428d18f3e6187d07c608db45cf6a50ed9359542bd74c60bd3ea2.
Original purchase corpus staysf9b129e3...; no expectation regeneration.

Final1dcd48..59f670 `make verify-server-core CORE_TEST_COUNT=1` exits0,
vet/core cold: gameui0.188s, production35.447s, save0.278s, transport13.347s;
separate Pitch alias cached, not another cold claim. Formulas/API regenerate
without diff. Final94c714..d2db4c `make verify-client`: types0errors/0warnings,
build213modules, tests8106/134 and shell boundary pass; composite exits2 at
the unchanged historical50a3a514 kernel-history defect (RP-131). Separate
f3348b..be2632 topology and remaining package/no-payment/copy checks exit0:
657keys/610existing orphan warnings; deployment content manifest unchanged.
All handles terminal before records. No browser/full native/CI or SQL rerun.

The first combined tracking patch missed a full-line context and failed
atomically;f27b58 confirms no partial tracking edits before the corrected patch.
RP-257/RP-258 are locally corrected, not designated-closed. Kernel0.3.160,
schema/pins, canonical corpora, applied migrations, CI policy, balance and owner
copy are unchanged; these presentation paths are outside watched prefixes.
Full new span after4030c895, including predeclaration, needs Claude independently.
Original B7/R9 body reconciliation, minted purchase-through-UI/AC12, actual
RP-253 SQL/persistence controls, R8 history/career and remaining pinned/harness
consumers remain. Docker capacity and RP-256 persist; no cleanup, box, archive,
mint, push or release promotion. Full nine-tier/platform goal remains active.

## 2026-10-06 — Projection reader range first filter

Review by: Codex (implementer self-review, FIRST FILTER ONLY).
Recorded by: Codex.
Reviewed range: `4030c895..b364b75f`, all thirteen paths, including the
`336503b3` predeclaration and the implementation, tests, docs and tracking.
Verdict: first filter passed; NOT the designated cross-party approval.

I inspected the reader/test diffs, then the committed documentation and full
planning span (764299/7d62d5). The ten runtime lines stay in the declared
presentation lane and reuse pinned derivation/canonical parsing. Tests prove
baseline admissions, corrected refusals and three independent compiling
omissions while preserving positive controls. Corrected files restore exact
SHAs; kernel/schema/corpus/migrations/CI/owner copy are unchanged. The records
carry the actual composite failure rather than substituting unit green.
Whitespace checks pass and the committed tree is clean (4878e6/f10a42).

Claude must review the full new span after4030c895, including this record commit;
no batch or RFC archive is authorized. Actual SQL, browser/default-player and
whole B7/R9/AC12 remain unproved. Next safe accepted lane remains the other
pinned/history/harness consumers; no goal completion or authority expansion.

## 2026-10-06 — R8 portable history consumer predeclaration

Previous turn is progress: RP-257/RP-258 runtime corrections, fired regressions
and synchronized records, committed through b6c7ebca; tree now clean, ahead80.
Read AGENTS/process and the accepted Reputation RFC. Broad reference output
truncated and an unmatched shell glob aborted one search; targeted actual
history, verifier, storage reader and harness sources were subsequently read.
No execution or conclusion comes from those incomplete searches.

VerifyFounderHistory resolves pinned bundles, invokes ApplyFounderLogged, checks
receipt/events/source coordinates and compares the head. Existing Reputation
integration tests verify purchase/rejection and plan/activation histories on
Postgres; they do not supply currently executed evidence while capacity is held.
The R8 two-Exit purchase career and non-unit Company run verifier remain separate
mandatory populations. The harness performs policy purchases manually through
Tree.Purchase and starter assembly; its comment's "served transitions" is not
evidence of a served transaction or stored Founder history. No harness change or
whole-harness verdict is authorized by this observation.

Bounded RP-259 supplement, test-only, under accepted R8:

- Consume unchanged source SHA f9b129e3... via the existing source validator and
  pinned bundle helpers. Build exactly24 histories: all20 direct cases, all
  three paired Founder Exit cases and the contiguous nine-node purchase chain.
  Each single-case wrapper rebases ONLY FounderLogSeq to1 (a new local history
  beginning at that row's existing revision). Canonical command whitespace is
  normalized; resolved values, timestamps, receipt, events, full states and
  expected pins remain source-derived. The nine-row chain preserves original
  log coordinates and adjacent full-state equality. No replay-generated oracle.
- Public VerifyFounderHistory must return verified for every honest history.
  Six one-field/one-record corruptions per history (144 refusals): wrong head
  mirror, wrong head pin, altered receipt outcome, extra event, first log
  sequence2, and toggled linked-source presence. Expect state_divergence except
  the sequence corruption's log_gap. Honest controls must pass first.
- Independently compile omissions of head-state, head-pin, receipt, event and
  source-presence comparisons; respective controls must fail, while honest
  populations remain green. Sequence has two defenses: probe them separately
  and together; record surviving defense honestly. Restore exact source SHA
  after every terminal probe. No assertion edits or expected-byte regeneration.
- Population closure must reject source-hash, missing-case and missing-chain
  corruptions, with a demonstrated validator omission. No new shared corpus,
  schema, applied migration, runtime, numeric kernel, balance, CI or owner-copy
  change. If the honest histories cannot verify, report setup vs product failure
  before expanding authority or changing expectations.

No record/source edit while a check/probe lives. Cold focused and full server-core
verification; no fresh SQL, browser, full CI, LoadFounderHistory execution,
two-Exit career, Company-run verification, AC3/AC15 or whole R8/RFC/1.0 promotion.
Original B4 CHANGES REQUIRED and capacity/RP-131/RP-256 remain. Full new span
after b6c7ebca needs Claude; no box, archive, mint, push, cleanup or status change.

## 2026-10-06 — R8 portable history consumer executed

Authority e3378e96, RP-259 test-only. The first attempt (1df092) does not compile:
the new test assigned a fixture string to save.EventKind. Corrected only that
conversion; no tests ran and no product defect follows from the setup error.
565f4c..1fff7f then passes all 171 subtests: 24 honest histories, 144 corrupt
histories and three source/population controls. Verbose tool display truncates
some lines, not test execution; exit0 and the closed source-derived loop/census
cover the declared population. No SKIP path exists in the new tests.

Expected full states, receipts, ordered events and pins come from unchanged
source bytes. Single-case wrappers rebase ONLY log sequence to1, preserving
original revision/command semantics; the complete nine-purchase history checks
adjacent full-state/revision/pin equality and uses its original coordinates.
All three paired Exits load their original current/next bundles, including both
v21-to-v22 activations. No live output is used as the expected replay head.

Independent compiling comparison omissions, assertions unchanged:

| Probe | Executed handles | Result |
|---|---|---|
| Head-state comparison | f78ee7..6c8b4b | 24 forged mirrors become verified; all 24 honest histories stay green. |
| Head-pin comparison | 636c47..32890f | 24 forged pins become verified; honest histories green. |
| Receipt comparison | 1331d7..ac3ba5 | 24 altered outcomes become verified; honest histories green. |
| Event comparison | b51c45..01d71c | 24 extra-event histories become verified; honest histories green. |
| Source-presence comparison | cb1bd3..f8af67 | 22 bad source arms become verified; both activation cases still refuse through missing next-bundle resolution. Honest histories green. |
| Entry-order check alone | e0b3ad..776613 | Survives: wire-to-entry log sequence still refuses all 24 corrupt sequences. All 171 subtests green. |
| Wire-to-entry sequence check alone | dd455e..7e48a7 | Survives: expected entry order still refuses all 24 corrupt sequences. All 171 subtests green. |
| Both sequence checks | 6151c0..d159b7 | 24 corrupt sequences become verified; honest histories green. |
| Population validator | fa75c0..ca2c4e | All three source-hash/missing-case/missing-chain controls fail; honest histories and other corruption controls remain green. |

Each probe finishes before restoration or the next edit. Temporary source
plumbing for the omitted source-presence block retains `_ = linked` so it
compiles. Exact runtime/test SHAs restore after every probe; final818102:
runtime2ffb54e186655740776ed891c69718d3d3c9b73fb8e41302482548bdde679fbc;
testb0b564a498abe88a77abb6ce72582ae6c9004c8218aa65a2c07ed80d9e8d7c33;
original corpusf9b129e36af5b536f67c7eddb8d6088c76ce5cb0b170cfeec6a5172e8a009782.
b468b3/78d355 confirm no runtime diff, only the new test before records. No
source omission or expectation regeneration is retained.

Final d12d87..a64390 `make verify-server-core CORE_TEST_COUNT=1` exits0:
vet/core cold, production35.097s/save0.276s/transport13.290s; Pitch0.305s cold
in core, separate alias cached. Formulas/API regenerate with no diff. Every
handle terminal before these records. No TS/browser/full CI/SQL rerun; prior
RP-131/RP-256 failures remain, not greened by a Go test-only wave. Go `_test.go`
files are explicitly exempt from kernel-affecting semantics (6190a0); version
0.3.160 unchanged. Schema/migrations/balance/owner copy/CI policy also unchanged.

Docker read-only bab33b/a79aa4: same two healthy Postgres containers, same
100%/39784KiB free. No new DB workload, restart or cleanup. RP-259 supplies
portable public consumer proof, NOT SQL LoadFounderHistory, a two-Exit purchase
career, the Company-run verifier, stored transaction provenance or AC15/full
R8/B4 acceptance. Original CHANGES REQUIRED and cross-party obligations remain.
Full new span after b6c7ebca needs Claude, including predeclaration and records.
Next safe accepted work: Company-run verifier and remaining harness consumers;
no checkbox, archive, mint, push, deployment, cleanup or 1.0 promotion.

## 2026-10-06 — R8 first-filter finding and bounded refinement

Review by: Codex (implementer self-review, FIRST FILTER ONLY).
Recorded by: Codex.
Reviewed range: `b6c7ebca..8d816b9c`, all nine paths, including e3378e96.
Verdict: CHANGES REQUIRED in this first filter; not a designated review.

Committed diffs fcdef8/5e0a46 and append-only log check915d64 agree with executed
scope; runtime unchanged and whitespace clean. One instrument gap remains:
validateReputationHistorySource requires exactly three paired Exits, but the
three declared corruption controls never remove one. The full population
validator omission fires the other controls; that is not a missing-pair case.

Before measuring this refinement: retain a fourth population control which
removes one paired Founder arm from a copied parsed source without changing
the raw SHA or purchase rows. Normal validator must refuse it; omitting only
the paired-count check must admit it and turn that control red. Then omit the
whole validator and require all four population controls to fail. Honest24
and corrupt144 histories stay unchanged and green. Restore exact test/runtime
SHAs; cold focused/core runs, no new data or product behavior. Earlier three-
control runs remain historical evidence, not a claim the fourth ran then.
Same accepted R8/test-only authority; full new span still needs Claude.

## 2026-10-06 — R8 missing paired-history control executed

Authority: c7c7a0fb first-filter finding/refinement, accepted RFC R8. Only the
test population and records change. The fourth control clones ExitCases and
removes one paired Founder arm without changing raw bytes or purchase rows.
acfebe..e1a6be passes the refined 172 subtests: 24 honest histories, 144 corrupt
histories and four population controls; no SKIP. The earlier 171-subtest runs
remain historical evidence, not retroactively relabelled four-control runs.

5f9661..fd201a omits ONLY the paired-count check. Both honest/corrupt-history
parents pass; missing-paired-exit alone fails with `corrupt history population
admitted`, exit2. Restore493a0b matches the refined test SHA. b8335e..eefcb1
then bypasses the whole population validator: all four population controls
fail with that diagnostic, while honest24/corrupt144 stay green, exit2. Each
handle reaches terminal before restoration or another edit; no assertion changed.
Final restoration e2a25e:
test01f060e4a3e637ccc583aae9ffefaad8f26320b8c680527db79fee96a7fbb9a9;
runtime2ffb54e186655740776ed891c69718d3d3c9b73fb8e41302482548bdde679fbc.

The first final core attempt e2a25e..1b8f83 exits2: existing httptest listeners
cannot bind `[::1]:0` under the execution sandbox, in deployment-operations,
deploymentrelease and operations. Production itself passes35.327s, but this
does NOT make the aggregate green. A scoped `make verify-server-core` escalation
permits the unchanged tests' local ports. 636886..a65ce5 then passes the complete
`make verify-server-core CORE_TEST_COUNT=1`: vet/cold core, production35.085s,
save0.178s/transport13.110s; those three listener populations pass. Pitch0.185s
cold inside core; the separate alias is cached, not another cold claim.
Formulas/API regeneration has no diff. Final a88e57..6084c2 focused run confirms
exactly172 passing subtests, zero failures, production0.292s, after restoration.
All handles terminal before these records; whitespace clean.

Read-only path/glob lookups with missing guessed paths produced no evidence;
actual tracked paths were then located and inspected. No retained source
omission, runtime/kernel/balance/corpus/schema/migration/CI policy change or
new SQL/browser/whole-CI evidence. RP-131/RP-256/capacity holds and actual SQL
reader/two-Exit career/Company-run/full R8 obligations remain. No checkbox,
archive, release, mint, push, deployment or cleanup. Full new span after
b6c7ebca, including predeclarations, refinement and records, still needs Claude.

## 2026-10-06 — R8 portable-history first filter after refinement

Review by: Codex (implementer self-review, FIRST FILTER ONLY).
Recorded by: Codex.
Reviewed range: `b6c7ebca..695da194`, all nine paths and all four commits
e3378e96, 8d816b9c, c7c7a0fb and 695da194.
Verdict: FIRST FILTER PASSED; c7c7a0fb's missing-control finding is locally
addressed by 695da194. This is NOT designated approval or RFC acceptance.

Earlier full test/docs diff inspection fcdef8/5e0a46 and the final refinement
inspection b1ee48/593ff3/ba5180 cover the whole range: source-derived heads,
unchanged original corpus, copied ExitCases removal, independently executed
omissions and exact restoration. Final focused172 and cold core/vet pass;
the sandbox-denied aggregate and scoped permitted rerun are both disclosed.
Logs append at EOF, live census is four, historical three-control runs are
unchanged. Net scope contains no runtime/kernel/data/schema/CI change or new
checkbox. No surviving probe or missing SQL/browser proof is relabelled green.

Claude's designated review must cover the COMPLETE span after b6c7ebca through
the commit recording this entry; the implementer's filter cannot substitute.
Actual SQL reader, two-Exit purchase career, Company-run verification, remaining
harness consumers and full R8/AC15 remain. No archive, mint, push or release.

## 2026-10-06 — R8 Company-run consumer predeclaration

Previous goal turn: progress (695da194/721c0ee1, fourth history control and
executed cold verification). Current tree761128 clean at721c0ee1; no live
check handle. Accepted R8 is authority, not the long-term board by itself.
Grounding: existing source f9b129e36af5b536f67c7eddb8d6088c76ce5cb0b170cfeec6a5172e8a009782
contains the terminal `reputation-starters-after-burnout`, not a subsequent
full Company run. Its source new Company is run2/tier0/cash1e3/generated15/
purchased0; its run_started v2 bonus is1.003. Existing plan Company rows begin
at tier1 with no starters or frozen bonus. R8 explicitly requires this consumer;
RP-260 is a proof gap, not a reproduced runtime/SQL defect.

Test-only bounded implementation, predeclared before generation or measurement:

- Start from the original source Exit's exact new-Company bytes and complete
  pinned tree bundle. Re-execute that source Exit only as a byte-checked setup,
  comparing receipt, ordered events and full Company output to the immutable
  source. Use its actual post-Exit Founder and complete frozen rows.
- Two profiles: source bonus1.003, and a unit1e0 frozen-row counterfactual with
  identical genesis. The unit control is synthetic, not a claimed stored career.
  Freeze contributions for the whole current run; later Founder/next-run facts
  must not re-materialize it.
- Exactly three applied commands per profile: one manual action after7000s,
  cross gate.t0_to_t1 one second later, then Wind Down one second later. These
  are explicit test clocks inside the existing24h horizon, not changed pacing,
  bounds, active-play policy or a claim of a served player workflow. Actual
  public command/transition boundaries execute; no direct tier/cash/gate edits.
- Retain a separate versioned Go-authored Company-run fixture with source SHA,
  exact genesis/pin, both profiles' canonical commands/frozen inputs, receipts,
  ordered events and terminal full states. Normal verifier tests READ retained
  expected bytes; they may not generate their own expected answers. An explicit
  root Make test-selector generation flag creates it once, and a separate
  regeneration-byte-equality test detects drift without rewriting normally.
- Independent first-action cash check uses the source's15 units/base1 rate,
  source1e3 grant, elapsed7000s, one base1 action and the declared frozen factor.
  This arithmetic oracle is not another invocation of the replay transition.
- Both honest profiles must return public verified and exact terminal state.
  Nine corruptions per profile must refuse: removed starter cash, removed five
  generated units, altered first factor, altered terminal factor, missing
  terminal, log gap, wrong constants pin, altered ordinary receipt, extra event.
  Expected verdicts: log_gap for missing terminal/sequence; constants_mismatch
  for wrong pin; state_divergence for the others.
- Four fixture controls: bad source SHA, missing profile, missing command,
  altered genesis. They must reject and fire when the fixture validator is
  bypassed. Independently compiling runtime omissions: force Reputation replay
  factor to1; omit ordinary receipt comparison; omit ordinary event comparison;
  omit terminal requirement. Honest controls must discriminate, including the
  unit arm staying green under the bonus omission. Restore exact SHAs after
  EACH terminal probe; no assertion/fixture regeneration during omissions.

Scope: one new production _test.go, one additive testdata/reputation fixture,
canonical docs and tracking. No retained runtime/kernel/balance/original-corpus/
schema/migration/CI/copy change, epoch mint or acceptance-bound relaxation.
Every check/probe handle terminal before source OR record edits. Full cold core
and focused tests after restoration; local listener permission if required,
not substituted no-op checks. Full new span after721c0ee1 needs Claude.
Actual LoadFounderHistory, stored two-Exit career/AC15, browser/RP-131/RP-256,
capacity, H4/H5 and remaining harness/full-nine-tier/platform work stay open.

## 2026-10-06 — R8 generation setup and arithmetic-oracle correction

4d4846f2 generation setup fb9a82/180d48 executes no tests: the custom Go test
flag must follow the package selector and `-args`; the initial `$` regex also
passes through Make expansion. Correct root invocation places the custom flag
after `./production -args` via GO_PACKAGES, with the ordinary selector/count
in GO_TEST_FLAGS. 224434 then compiles and finds a test-only nonexistent
decimal.FromInt64 symbol; FromFloat64 is the actual existing exact constructor
for these small integer constants. No product defect follows from those errors.

1ff47a..c183a5 executes setup and the first action but rejects the independent
arithmetic oracle before writing any fixture: got cash1.06316e5, want1.06316003e5.
Grounding3749dd/92f4c1: the manual action matches only its explicit target in
contributionFactorForTarget; the all-target prestige row contributes to generator
production, not that base manual action. R3 specifies production. The erroneous
oracle multiplied the one-action grant by the bonus. This is an oracle error,
not authority to change production or accept an unexplained difference.

Before the next measurement, correct the independent formula to:
source1e3 + source15 * base1 * elapsed7000 * declaredFrozenFactor + baseAction1.
Both populations, commands, timings, refusal counts and public/full-state gates
remain exactly as predeclared. No fixture exists yet, no expectation is rewritten
to hide drift, no acceptance bound is loosened. All handles terminal before this
record; no retained runtime edit. Subsequent output must still be byte-retained.

## 2026-10-06 — R8 Company final-head control refinement

After the retained fixture, honest2/public corrupt18/population4 and five
compiling removals execute, Codex self-inspection finds a missing instrument
control: the detailed verifier's terminal full-state equality is asserted for
honest profiles but has no retained forged expected-head case. This is not a
runtime defect or designated review. First cold core7198ec..32aa14 passes;
final focusedddb8a7..415623 passes24subtests plus fixture equality. Those runs
do not retroactively execute the additional controls below.

Before measurement: factor the terminal-state predicate into a test helper
consumed by the honest tests; add two false-expected-head controls, one per
profile, replacing only expected final cash with0 in a copied JSON object.
Normal helper must refuse; omit only its canonical final-state comparison and
both controls must fail, while honest2/public corrupt18/population4 stay green.
No change to the retained fixture or public-input mutation population, and no
claim the public verdict accepts a stored-head parameter. Restore exact test
SHA and rerun cold focused/core. This additive test-only refinement remains
inside accepted R8 and the complete span after721c0ee1 needs Claude.

## 2026-10-06 — R8 portable Company-run proof executed

Authority4d4846f2, oracle correctionf7822250, head-control refinementd23da963;
RP-260. d16f97..a0c3ac rejects the test's missing required cross_gate.route_id
before generation, not a product defect. Adding explicitnull satisfies the
existing request contract. 671a3b..956a73 then writes the additive Company-run
fixture once. Original corpus SHA f9b129e36af5b536f67c7eddb8d6088c76ce5cb0b170cfeec6a5172e8a009782
stays unchanged. 8ba2ba's first retained-reader run refuses both honest profiles
because fixture indentation was passed as canonical command bytes. Fix944482
normalizes presentation whitespace only, as other committed replay readers do;
9a5076..408601 then passes honest2/corrupt18/population4 plus fixture equality.
No public production defect is asserted from either setup error.

The generator first byte-checks the original source Exit receipt, ordered
events, final Company and new Company, then uses its actual post-Exit Founder
and three frozen rows (Fiscal generator/hoard plus Reputation). Exact source
genesis is run2/tier0/cash1e3/generated15/purchased0. Both profiles issue manual
at7000s, ordinary Garage gate at7001s and Wind Down at7002s, with no direct
cash/tier/gate edit. Unit factor is an explicitly synthetic frozen control,
not stored provenance. Independent first-action formula is recorded inf7822250.
Normal verifier tests read retained bytes; generation is a separate equality
gate and was NOT run during any omission. Ended-run cash: source6.34609e3,
unit6.031e3; lifetime source1.0534509e5/unit1.0503e5. Full states, not these
summaries alone, are compared by the retained-reader test.

Executed omissions (the reader-only selector excludes fixture generation):

| Probe | Handles | Executed discrimination |
|---|---|---|
| Force Reputation replay factor to1 | d45b56..ddc321 | Source-bonus honest run fails; unit honest run passes. Unit first/terminal-factor corruptions become verified. Source missing-terminal/log-gap are stopped earlier by state divergence, recorded rather than called a sequence defect. |
| Ordinary receipt comparison | 69c28e..7ecc8f | Four corrupt runs become verified: two forged receipts and two missing starter-cash cases. Honest controls pass. |
| Ordinary event comparison | b7d3ce..f3a162 | Both extra-event cases become verified; honest controls pass. |
| Terminal requirement | b0908d..cdf8ad | Both missing-terminal cases become verified; honest controls pass. |
| Fixture population validator | 1342bf..30b26c | All four population controls fail with corrupt-population-admitted; honest2/corrupt18 pass. |
| Detailed final-head comparison | aa402b..94595e | Both retained false-head controls fail; honest2/corrupt18/population4 pass. |

Probe plumbing failure disclosed: reversing the removed terminal block with
an empty context misplaced it at file start (aa3430 SHA mismatch); the next
fixture-validator attemptab6c47 fails compilation and executes NO tests. This
is excluded from discrimination evidence. 1bb246 inspection catches it;
cd347d restores verifier byte-exact, then1342bf..30b26c is the genuine compiling
validator omission above. All handles terminal before each repair/record.
Other restoration checks6328de/5f9058/6c5030/222048 and final223bf6 match:
replaya2bf1bf6a6338b0548587d067dc96569a8355363cc50f1b2905782a7edc2b6ca;
verifier0d6a643abdac495ba0e115fe7cbd9e0ab41202a3794998eeab82ea70523510cd;
final test518641286a8d152abe568471705c52a4d1cc6153cd5c8b0dbdc971a2aa98f4db;
new fixturee48187d7afd153a950c3e1cbc369c5da206020e2aefdc64e3f24b50e1a505aea.
No assertion omission, probe rewrite of expected fixtures or runtime omission
is retained.

New final-head controls c71947..e45136 pass: 26subtests (honest2, false-head2,
public-corrupt18, population4), plus the fixture-equality parent. Initial
24-subtest evidence stays historical. Final5c07ba..d209ff cold
`make verify-server-core CORE_TEST_COUNT=1` passes: vet/core, production35.060s,
save0.174s/transport13.175s, Pitch0.269s cold in core (separate alias cached),
formulas/API regeneration byte-unchanged. No fresh TS/browser/whole-CI or SQL
proof. _test.go and additive testdata/reputation are outside kernel semantics;
kernel0.3.160 unchanged. No balance/schema/migration/CI/copy/old-corpus changes.

Docker read revalidation: sandbox denies socketd55283, scoped read4c07ba
confirms the same two healthy Postgres containers; read-only04fe31 capacity
still100%/39784KiB free. No DB workload, restart or cleanup. These portable
proofs do NOT replace actual LoadFounderHistory, persisted two-Exit purchase
career or real-Postgres AC15. Original review objections, all prior designated
review obligations, H4/H5, RP-131/RP-256 and full nine-tier/platform scope remain.
The first tracking patch misses a CURRENT-STATE context and fails atomically;
e00e5d confirms no partial diff, then corrected7175d7 updates the actual lines.
Next safe accepted work: remaining R10 harness consumer audit, no optimization
or policy/threshold/bound waiver. Full new span after721c0ee1 needs Claude;
no checkbox, archive, mint, push, deployment, cleanup or release promotion.

## 2026-10-06 — Company-run checkpoint first filter

Review by: Codex (implementer, self-review FIRST FILTER ONLY).
Recorded by: Codex.
Reviewed range: 721c0ee1..d2247d67, all four commits and all ten changed paths.
Verdict: PASSED first filter; NOT designated independent approval.

Checked the source-bound generator, retained-input public verifier, detailed
full-head predicate, every input/population control, generated fixture and the
complete docs/tracking diff. The source profile preserves the actual starter
genesis and non-unit bonus; the unit control is expressly synthetic. Expected
bytes are retained independently of normal reader execution. Six compiling
omissions discriminate as recorded above; the invalid compilation attempt and
oracle/setup mistakes remain disclosed, not counted as evidence.

Final committed-HEAD run1a4b43..8b2135 (session85727, terminal exit0) passes
26 subtests plus fixture equality with -count=1, production0.499s. Final cold
core5c07ba..d209ff remains the broader server proof; no SQL/browser/whole-CI
claim is added. No production omission or old corpus edit survives. Range
stat7e48c3 is ten paths1552 insertions/7 deletions; clean starting checkpoint
fcd458 has no unrelated dirty work. No unresolved first-filter finding in
this bounded test/fixture range. Actual persisted career and all separate
review obligations remain open. Claude must review the complete span after
721c0ee1 through this record commit before any archival eligibility.

## 2026-10-06 — R10 career input admission predeclaration

Clean source7da200f0; accepted R10 H4 declares exactly cheapest, seeded_uniform
and none, while H5's leave-one-out mask is a tree node. Evidence rule3 requires
invalid measurement inputs to fail loudly. Source lead a95197/482b3d: policy
admission occurs inside the affordable-node loop; an unknown Exclude is never
matched and may silently behave like the baseline. This is not yet an executed
defect or a designated review of B8.

Bounded population: actual loaded first-hour suite and existing tree fixture,
Chaos seed0 under the ratified two-hour horizon, unchanged experiment. Six
negative inputs: empty policy, unknown policy, and unknown exclusion, each at
fixture threshold1e5 and live threshold1e12. Require ErrReputationCareer; no
report regeneration, output expectations generated from the tested result,
instrumented budget, horizon extension or change to purchase choice. Four
positive inputs at1e5: the three declared policies, plus cheapest excluding
the actual reputation.starter.cash_small node. Require completed two-Exit
career/run3 gate; no-purchase arm buys/applies nothing and freezes unit bonus;
purchasing arms buy nodes; excluded node is absent from purchases and starters.

If baseline admits a malformed input, retain the red test and enforce only
closed policy/known exclusion admission before simulation. Demonstrate each
new guard's removal independently with the same tests, restore exact source
SHA before the next probe, and abort sequencing on a restoration mismatch.
Cold focused harness tests, root vet/fast harness and core must then pass;
existing exhaustive H4/H5 reports stay byte-untouched and their recorded FAIL/
epsilon/horizon questions remain. Harness-only source is outside kernel watched
paths. No balance/runtime/schema/CI policy/owner copy/old corpus or threshold
literal change, mint, archival or full AC13 claim. Complete new span after
7da200f0 needs Claude. Actual DB/capacity and all other release blockers remain.

## 2026-10-06 — R10 career input admission executed

RP-261, authority9337f4b8, source7da200f0. Baseline f6bc98..855771
(session40584, terminal exit2) executes six malformed cases and four legal
careers. Four malformed cases return nil error: empty/unknown policy at1e12,
unknown exclusion at both1e5/1e12. The two malformed policies already fail
after entering the affordable-node loop at1e5; that defense is preserved.
All four legal controls complete. This is an instrument-input defect, not
a live player path, report corruption or authority to change balance.

The twelve-line harness repair admits only the three declared policies and
known nonempty node exclusions before simulation. Patched e28e17..2097e1
(session81036) passes all ten subtests cold; all four legal result predicates
pass, package7.773s. No claim these structural
controls alone prove full byte equality of every valid career.

Independent compiling omissions, same full ten-case selector:

- Policy admission333499..e4845c (session8885, exit2): only the two1e12 policy
  refusals fail. The late1e5 policy defense and all four legal controls survive.
- Known-node admission2fc6af..8250f4 (session41987, exit2): only the two unknown
  exclusions fail. All policy refusals and all four legal controls survive.

000417 and b44800 verify exact restoration before proceeding; mismatch aborts
the orchestration rather than allowing another probe. No compiling/restore
failure in this wave. Final780bff source/test SHA:
career c1fef55d7a38ac761020e37d5a727e365060f56299eeccdce7f87ab605784bc1;
test4ce61cf3f301219fce2cc97275c57268a91eb1ac26ce06af43dc3cb6d6d89fcb.
All handles terminal before source or record edits.

Cold gates (fast/core ran concurrently, no edits while either lived):

- `make verify-harness-fast HARNESS_TEST_COUNT=1`, ab4425..76546f
  (session11053), exit0: full fast harness98.786s, role activation0.208s,
  Commons invariance0.472s and guard mode. The new ten-case tests are
  unconditional; exhaustive H4/H5 are still explicitly separate.
- `make verify-server-core CORE_TEST_COUNT=1`, b4d700..fa50f8
  (session45078), exit0 under the existing narrow local-listener permission:
  vet/core, production44.782s, save0.270s, transport13.314s. Pitch0.376s
  cold in core; separate alias cached. Formulas/API regeneration byte-unchanged.
- CI topology b8755e: declared fast/maintenance lanes pass, all13 negative
  controls reject. No workflow change or fresh hosted/whole-CI claim.

Full `make reputation-harness-check`, 98b545..c06ec1 (session2471,
terminal exit0), actually runs both exhaustive populations: H4's97treated/
control pairs165.53s; H5's97seeds × (baseline+nine masks)757.23s;
package922.866s. Both retained reports byte-reproduce, with no update flag.
H4 still records exactly the same six Casual ties (seeds1/6/8/11/24/25),
not a passing balance gate. H5's epsilon/run-4 questions remain. Final report
SHA780bff matches pre-runfc80b7/470c26:
H4 648f36d685c07471830692bf4c55810fc6c95ed4dcd00306d75a36697c2a9aa8;
H5 4ed79054baa69389fff34cf850327c07b882b6b4aa507c25dc4d3d13808ab63c;
threshold46a8fea612ff7e777a38c54166e542dff33e92ec18380fdf09ff4bdd1df72808.

Read-only leads found during the exhaustive handle are recorded now that it
is terminal, not hidden by the edit freeze: RP-262's mandatory H3 multiplier
control was not found (af79b4/8c38ac); RP-263's fixture identity divergence
is a source lead (482b3d/20704b/f4c914/5ddf74), not an executed pin finding;
RP-264's exclusion-oracle lead needs a malformed-row experiment (482b3d).
None has new measurement or fix authority beyond a bounded R10 predeclaration.
Next safe work: RP-262, without guaranteeing an unmeasured tiny multiplier
can move discrete milestones or weakening the criterion if it cannot.

External Git bookkeeping changed during the long check: 9cb045 confirms
HEAD and origin/main both9337f4b8; remote reflog05ebbe says update by push.
Codex performed no push/fetch, and local HEAD did not change. The new guard
and test remained uncommitted at80cd5e; no unrelated source changes appeared.
The narrowed ps readb3d789 was sandbox-denied, contributes no liveness proof
and was not escalated. Missing guessed source path098c56/glob2b1868 yielded
no facts; actual catalogf4c914 and seedf71804 were read instead.

Harness-only, outside kernel watched paths; kernel0.3.160 unchanged. No
production/balance/report/corpus/schema/migration/CI policy/owner-copy changes.
No fresh SQL/browser or AC13/archival/mint/release promotion. Docker/actual
career and all prior reviews remain held. Complete new span after7da200f0
needs Claude; the original B8 range is not designated-approved by this work.

## 2026-10-06 — Career input checkpoint first filter

Review by: Codex (implementer self-review, FIRST FILTER ONLY).
Recorded by: Codex.
Reviewed range: 7da200f0..3f5cc158, both commits and all ten changed paths.
Verdict: PASSED first filter; NOT designated independent approval.

Reviewed the two upfront admission guards, all ten retained controls and the
complete predeclaration/docs/tracking/log changes. Guards precede simulation;
the existing late purchase defense remains. Policy and known-node omissions
independently discriminate at the expected populations, without changing
legal-arm predicates. Retained report reproduction is separately executed,
not inferred from those predicates. Report/threshold/kernel/balance/CI bytes
are unchanged (a1f53f); scoped diff5c93cd is whitespace-clean. Append-only
log additions5aa880 are at EOF. Next source leads are clearly unexecuted and
do not authorize a retune, report rewrite or owner-choice substitution.

Committed-HEAD focused5d0739..09439f (session60848, terminal exit0) passes
all ten controls cold, harness7.517s. Fast/core/vet/topology and full report
reproduction remain as recorded above; no new SQL/browser/whole-CI proof.
Follow-up broader H3 string searchfaf407 finds only the RFC clause in server,
client/src and the Reputation planning scope; first attemp ta611a6 includes
a nonexistent client/tests directory and is not exhaustive evidence itself.

No unresolved first-filter defect in this bounded admission range. All prior
reviews, actual persisted career/capacity, H4/H5 and original B8 remain open.
Claude must review the complete new span after7da200f0 through this record
commit before any archival eligibility. No source probe or live handle remains.

## 2026-10-06 — H3 prerequisite: complete career fixture identity

Previous turn is progress, not a waiting loop: RP-261 is corrected and full
H4/H5 reproduce. Source6963692b is clean (d7405c/aa94e3); no live handles.
Typographical clarification to the preceding entry: "first attemp ta611a6"
means the first attempt a611a6; faf407 is the successful broader search.

Grounding the RP-262 multiplier experiment finds two dependencies. RP-263's
fixture helper changes parsed catalogs but retains the base artifact map/hash
(482b3d/20704b/f4c914). RP-265's Reference planner omits external frozen rows
from ranker advances/transitions/projections (bef2ed/e732ea/acd43d/8fdeb7).
A three-persona multiplier experiment must not hide either omission behind a
milestone result. No numerical failure or full B8 verdict is inferred here.

First bounded wave is RP-263, under accepted R10's fixture-first measurements
and R2's complete paired tree/economy artifact contract. Before measurement:
load the existing ratified first-hour suite and its current fixture helper;
require retained artifact bytes for the exact existing tree and its paired
economy declaration, a freshly computed complete constants hash distinct from
the base epoch, and agreement with a public replaycatalog.Load roundtrip.
The untouched base suite/artifacts/hash must remain exact. Public career output
must identify the actual supplied fixture hash, not the old base hash.
No fixture threshold/value or tree/economy disk bytes change.

Retain three source-bound checks (tree bytes, economy declaration, composed
hash/loader) and one actual Chaos seed0 no-purchase career identity check;
normal cold execution must fail on the current helper/coordinate if inconsistent.
If confirmed, fix only helper composition through existing strict loaders and
the measurement's run-key coordinate. Add copied-input negatives for removed
tree, undeclared source and false hash; each must refuse rather than relabel.
Prove helper/coordinate omissions independently, exact restoration before the
next probe, then cold fast/core/vet. Full H4/H5 numeric reports must reproduce
without update flags; their historical bytes and FAIL/gaps remain intact.

No assertion that this makes H3 or RP-265 complete, no source-pinning of a live
epoch, report regeneration, policy change, kernel bump, threshold retune, budget/
horizon waiver, SQL workload/cleanup, checkbox, archival or release promotion.
Next dependency is RP-265 before the full RP-262 measurement. Complete new span
after6963692b needs Claude; independent review is not supplied by this work.

## 2026-10-06 — Career fixture identity corrected and reproduced

Executed26e97e4b's RP-263 wave, source6963692b. The initial focused command
21bdc3..9bc4b1 (session23829) executes zero tests: Make consumed the trailing
selector dollar. It is not evidence. Corrected selector e3ccf1..424a41
(session51861, terminal exit2) executes all four predeclared checks and fails
all four. The helper retains neither the tree nor paired economy artifact,
and both its hash and the actual completed Chaos seed0 career's run key name
the base epoch instead of the independently reconstructed fixture.

Base catalog identity:
`sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6`.
Complete paired tree/economy fixture identity:
`sha256:3625eddb73da494574a2031fd483e93af236aa9b1f14482e5b51ad6ea3d0f7b0`.
Threshold1e5, Chaos seed0, CareerNone and the existing experiment remain exact;
these are separate inputs, not extra fields hidden inside that catalog hash.

Correction copies every base artifact, retains exact existing tree bytes and
the already-proposed economy declaration, computes ConstantsHashArtifacts and
uses strict replaycatalog.Load. RunReputationCareer copies that bundle hash
into the career suite's RunKey coordinate. No arithmetic, purchase policy or
natural gameplay transition changed. The retained test compares sources and
the full public-loader roundtrip, executes a completed two-Exit/unit-factor
no-purchase career, and checks that the base suite remains untouched. Three
copied-input controls (missing tree, undeclared source, false hash) refuse at
the public loader, not a new second-authority validator.

Normal0e4a87..f4e3e4 (session62184) passes all seven controls cold. Compiling
omissions, with no source/record edit while any test handle lives:
- 675437..f5238b (session29530, exit2): restore old artifacts/hash on the
  composed helper; all four identity checks fail, three refusals stay green.
  a9f855 confirms exact restoration before the next probe.
- 397ddc..0776f4 (session62061, exit2): omit only careerSuite.ConstantsHash;
  actual-career identity fails and all six other controls stay green.
  d8520f confirms exact restoration before broader tests.
- 097251..11900b (session45303, exit2), only after the full study terminates:
  omit helper artifact copying and mutate the original base map. All seven
  children pass, but the base-preservation cleanup correctly fails the test.
  This independently exercises the predeclared untouched-base assertion.
  e6de93 confirms exact restoration;84a6e1..f37a9b (session39063, exit0)
  passes all seven normally, harness2.172s.

Restored SHA256s: career.go
`b3a815a4009416f0b8e67f8818fda59b9d27e135da1dcb4f44acacee5df9dca4`;
career_test.go
`b43b2ab9b6468bcb4e97aac3673bec0735df77c8f21ee8fa92b747f003912917`;
career_identity_test.go
`6bcfa145619d11d278039559b675a15b13f5f928ace7ea7dd904a06aaf73e625`.
Restoration checks abort before another probe on mismatch.

Cold make verify-harness-fast0ba097..8769b2 (session16414, exit0):
harness56.089s, role population0.251s, Commons0.544s and guard mode pass.
Cold make verify-server-core a47117..40c418 (session34331, exit0) uses the
existing narrow local-listener permission; complete core/vet pass, including
production49.279s, save0.297s and transport13.379s. Formula/API regeneration
is byte-unchanged. Its separate Pitch alias is cached; the main cold core
Pitch population passes0.362s. Topology afdee3 passes all13 negative controls.

Full make reputation-harness-check21abec..39b4db (session76332, terminal
exit0): H4's97 treated/control pairs reproduce in179.03s, H5's970 arms in
660.93s; total840.123s. No update flags. H4 remains FAIL on Casual seeds
1/6/8/11/24/25 (6cfdf4); a successful reproduction is not a passing criterion.
H5's epsilon and run-4 dimensions remain unresolved. Report hashes7e0bea
remain H4 `648f36d685c07471830692bf4c55810fc6c95ed4dcd00306d75a36697c2a9aa8`,
H5 `4ed79054baa69389fff34cf850327c07b882b6b4aa507c25dc4d3d13808ab63c`,
threshold `46a8fea612ff7e777a38c54166e542dff33e92ec18380fdf09ff4bdd1df72808`.

All handles terminal before these docs/ledger/board/log edits. Initial combined
record patch rejected an incorrect roadmap-log anchor; e8690b confirms zero
partial edits, then the exact patch succeeds. Exploratory missing-path searches
are not executable evidence; actual paths resolve through029655/539c4a.
No kernel160, live runtime, balance, original corpus, schema, CI, owner copy,
budget, horizon, report rewrite or epoch change. No fresh SQL/browser/whole-CI
claim. RP-263's report-envelope provenance remains open; its helper/key portion
is corrected locally only. Next: predeclare/execute RP-265 Reference planner
frozen-input lead, then RP-262 H3 and RP-264 exclusion oracle. No numerical
planner finding, box flip, archival, mint, push, cleanup or release promotion.
Whole new span after6963692b requires Claude; all earlier reviews remain open.

## 2026-10-06 — Fixture identity first filter

Review by: Codex (implementer self-review / first filter).
Recorded by: Codex.
Reviewed range: 6963692b..2a21fe20, both commits and all eleven changed paths.
Verdict: PASSED first filter; NOT designated independent approval.

Reviewed predeclaration26e97e4b, complete helper composition, run-key assignment,
all seven controls, canonical docs, partial RP-263 disposition and the current
queue/board/plan/append-only records. The expected sources are reconstructed
independently of the helper's parsed objects. Public loader refusal does not
claim that RunReputationCareer validates arbitrary caller-constructed catalogs.
The full-hash field is catalog identity, not a hash of threshold/policy/exclusion.
Compiling fixture-identity, run-key and base-copy omissions each fail the named
assertions; restoration aborts on mismatch. Historical report provenance stays
open, H4's fired criterion remains explicit and RP-265 remains source-only.

Committed-HEAD focused47320e..5a62f2 (session85907, terminal exit0) passes
all seven controls cold, harness2.014s. Broader cold fast/core/vet/topology and
full report reproduction are the executed runs above, not a new whole-CI or
SQL/browser verdict. bcfb33 confirms clean main ahead4, whitespace-clean range
and EOF-only log additions. Kernel160, balance, reports, original corpus,
live runtime, generated contracts, CI and owner copy remain unchanged.

No unresolved first-filter defect in this bounded helper/run-key range.
Complete new span after6963692b through this record commit still needs Claude;
all prior reviews and actual persistence/career/AC13/AC15 obligations remain.
Next accepted work is RP-265 before RP-262, with RP-264 separately bounded.
No live handle, source mutant, checkbox, archival, mint, push or release claim.

## 2026-10-06 — Reference frozen-input consumer diagnosis predeclared

Previous turn is progress: RP-263's helper/key correction and full report
reproduction are committed. d18d4e09 is clean main ahead5 (4239fb), with no
live handle or new designated verdict. Current accepted work remains R10.
Grounding c30666/545bec/ab4165 confirms nil external inputs in the Reference
ranker and actual bank advance; no fresh numeric failure is inferred yet.

RP-265 diagnosis population, before measurement: strictly compose the current
paired fixture through RP-263's helper. A hypothetical earned6 Founder buys
unlock.p05, cash_small and generated_beige_tower through the actual tree
purchase function, then NewRunState and ApplyReputationStarters assemble run3
at the existing harness Epoch. Require cash1e3, five generated/zero purchased
Beige Towers and the actual tree-derived frozen factor1.003. This is an
isolated legal headless fixture, NOT a persisted or naturally earned career.
Resolve its declared contribution through the public frozen-input resolver.

Compare three explicit arms: no external row, a synthetic unit row and the
actual non-unit row. For each, retain four observations through the actual
Reference ranker/runtime: projected-milestone denominator, one-second candidate
advance, a one-unit generator purchase whose lazy accrual spans that second,
and the actual bank branch. Independent witnesses are ProjectRates and
SimulateAdvance/SimulateTransition supplied the same frozen input; compare full
encoded candidate states where applicable. Require the no-row/unit output rate
5 and non-unit rate5.015, not merely inequality against the broken consumer.
The bank arm explicitly replaces its cash ledger with the ordinary initial
zero-cash ledger to make every purchase unreachable in the one-second decision
horizon; it does not alter the retained starter/source artifact or any scenario.
Its production, frozen factor and declared policy remain identical. This is a
counterfactual fixture branch, not a claim about an observed player trajectory.
Retain a generated-Beige-Tower effect mask and a zero-production control.

Cold baseline must execute all arms. A consumer mismatch is a finding; an
oracle/fixture setup error is not. This FIRST range is diagnosis/test only:
no product or solver correction is authorized by its own start entry. If a
failure is confirmed, record it and predeclare the exact corrective range
under accepted R10 before production/solver changes. In particular the current
SimulateResourceRate API accepts no external argument; it cannot be labelled
a defective implementation of an input it never promised. A canonical masked
rate producer and consumer binding need explicit planning, kernel protocol,
source isolation, failing controls and a separate report-drift disposition.
Do not rewrite historical H4/H5 reports to absorb a changed instrument.

Additional source lead RP-266: the AST caller map omits SimulateResourceRate
(145c4f/01eed4). Current discovered callers are harness/tests; this is not an
executed forbidden-caller finding. A corrected rate producer must stay inside
the same simulation-only boundary, with an executed negative. No broader
boundary redesign is implied. Initial jq guesses selected nonexistent nested
shapes (01eed4/4a34ff); b8c7cb resolves them. The old corpus is untouched and
none of those queries is proof of a run3 population.

No unit/non-unit milestone verdict (RP-262), H4/H5 completion, retune, literal
ratification, budget/horizon waiver, report regeneration, epoch mint, SQL or
cleanup, owner copy, box, archival, push or release promotion. Independent
review of the complete new span afterd18d4e09 remains mandatory. Goal active.

## 2026-10-06 — Reference consumer failure and corrective range predeclared

Baseline31e2c3..16647e (session83250, terminal exit2) never reaches consumer
comparisons: ProjectRates rejects a v14 headless state paired with a modern
active-play bundle. This is my oracle setup error, not16product defects.
The first-hour runner constructs legacy CurrentVersion14 state and does not
simulate active play (7f6a05). The corrected independent rate projection uses
the same economy/external input with no active-play bundle, explicitly requires
that v14 contract and retains the exact5/5.015 expected rates. No state version
or game contract was changed. That alignment does not turn the headless
population into a modern played/persisted Company career.

Corrected cold d117ba..7b2ccb (session55650, terminal exit2) executes all16
observations:12controls pass, four tree/non-unit consumers fail. Projection
5 instead of5.015; candidate advance cash1005 instead of1005.015; lazy purchase
cash995 instead of995.015; actual bank cash5 instead of5.015. Healthy independent
production calls and no-row/unit inputs discriminate. Full encoded states
are compared, not merely those cash strings. RP-265 is now a reproduced
instrument-consumer defect, not a claim that the live game drops the bonus.
No test/probe handle remains before this record.

Corrective authority is accepted Reputation R10 H3/H4/H5 consuming R3's frozen
bonus through the existing canonical production stack; this is missing input
binding, not a new formula, policy or balance choice. Separate bounded range:
1. Give RelevanceSuite an internal simulation-only frozen contribution set;
   the career Reference factory copies the actual runtime input into it.
   Pass it to every existing advance/transition arm and the actual bank arm.
   Standalone suites keep nil; no public report or wire field is introduced.
2. Add an explicit SimulateResourceRateWithContributions producer sharing the
   existing canonical masked assembly. Existing SimulateResourceRate retains
   its nil-input contract by delegation. Bind planner rate projections and
   first-hour legal-command rates to the matching actual external input.
   Preserve ablation/validation and do not reimplement rates in the harness.
3. Execute RP-266's forbidden old-rate reference and alias before adding both
   old/new rate names to the existing AST boundary. Retain literal/alias/decoy
   controls for both; do not expand the permitted caller set.
4. Retain producer unit/non-unit/masked/no-production/invalid-input controls
   and all16Reference observations. Omit factory, advance, transition,
   projection, bank and rate-assembly bindings independently, require compiling
   failing evidence and exact restoration. A blocked secondary assertion is
   not independent evidence; exercise candidate projection separately if needed.

The production simulation/boundary files are watched: the implementation
commit must bump kernel160→161 in source, Go and TS together. This is a real
new simulation-input capability, not a false version signal or a change to
live production arithmetic. Cold fast/core/vet, Go/TS vectors/client and
boundary/topology checks follow. Historical RP-131 stays visible; no rewriting,
guard waiver, CI-policy change or all-CI-green inference.

Before/after the correction, explicitly observe the existing Reference seed0
treated/control career at threshold1e5 and the original experiment/policies;
require completed Exits and visible gate/null outcomes, log complete source
coordinates and report numeric drift honestly. Keep v1H4/H5 bytes unchanged.
If their strict reproduction fails on a changed Reference row, record it as
stale instrument evidence, not a new balance verdict or permission to update
the expected bytes. Other personas/policies/thresholds/horizons are unchanged.
Do not run simultaneous probes or edit while any test handle lives.

No H3 milestone (RP-262), AC13/AC15, epoch/literal/copy ratification, SQL or
cleanup, box, archival, push or release promotion. Diagnose/repair only these
input seams and preserve all previous verdicts. New complete span after
d18d4e09 still needs Claude; implementer review remains a first filter.

## 2026-10-06 — Reference frozen inputs corrected and measured

Resume eb215d/a6cf49 confirms main ahead7 ataa4c241b with only this unfinished
correction dirty; no intervening local Claude commit/verdict is present. No
live handle remains from the saved checkpoint. Required authority is unchanged.
The complete corrective predeclaration is aa4c241b; diagnosis is3fb50327.

Pre-correction actual Reference career b52bd1..c2c45e (session48716, exit0)
reproduces the retained seed0 row: treated355000ms, control357000ms, factor1.001,
only unlock.p05 purchased and no applied starter. Both careers complete with
two Exits. The complete paired catalog hash is
sha256:3625eddb73da494574a2031fd483e93af236aa9b1f14482e5b51ad6ea3d0f7b0;
scenario hash18a6f16a6d9b4a66469750c728599d32768136ee22f5b9178832d19b627f85ba,
policy hashe5e5de7051beb0340e54f7013ce7d4a48c35bfcc3343220310290478445d10c3.
Threshold1e5 and the original experiment (purchased minimum200, burnout2,
route50, seed capital1e4, generated10) remain unchanged. Full outputs and
coordinates are in that executed observation, not reconstructed from memory.
RP-266 baseline b47ea2..b751df (session60256, exit2) misses exactly the old
rate name's direct and alias references; its quoted decoy passes.

The correction copies runtime.external into an internal RelevanceSuite slice,
binds all four existing solver transition sites plus action-free advance and
projection, and binds actual bank waiting and legal-command rate projection.
The new simulation-only SimulateResourceRateWithContributions shares canonical
validated/masked assembly; the old method delegates with nil and retains its
contract. Both names join the existing caller detector; permitted callers are
unchanged. No copied arithmetic, public report/wire field or live formula change.
The watched simulation/boundary changes carry kernel160→161 in all three sources.

New producer tests initially had setup failures, excluded from defect evidence:
b51180..cd006a (session22025) used a nonexistent Decimal constructor;
c44e5e..6af733 (session50696) used an incomplete encode fixture lacking its
cursor; 5dcf64..84a629 (session62674) caught the now-unused import. These are
my test-authoring errors, not product failures or successful severing probes.
The fixture now uses actual prestige.NewRunState rather than patching a
partial state into an assumed-valid one. All handles were terminal before edits.

Normal cold 4f9d3a..bd77ab (session77023, exit0) executes16Reference
observations, nine valid producer profiles, eleven invalid-input refusals,
six boundary controls and the actual repository caller scan. This preserves
nil/unit/non-unit, ordinary/masked/zero production, full encoded state equality,
pure rate projection and fail-closed contribution/mask/resource admission.

Independent compiling omissions, each followed by SHA-exact restoration before
the next probe (all commands use root make test-go and -count=1):

| Removed binding/guard | Executed output / handle | Fired observations |
| --- | --- | --- |
| Factory frozen-input copy | e9ecb4..07fe24 /30809 | Three non-unit comparisons: projection, advance, purchase; actual bank remains independently bound |
| Candidate advance input | 4fcd2e..799710 /42946 | One non-unit advance comparison |
| Candidate purchase input | bf75de..809384 /83757 | One non-unit purchase-state comparison |
| Projected-milestone input | 159165..f60c87 /88244 | Projection and independently reached candidate-projection assertion; candidate state still matches |
| Actual bank input | d11d32..999d86 /50800 | One non-unit bank-state comparison |
| New producer assembly input | 325983..185d48 /71092 | Two Reference projections, one non-unit producer profile and seven malformed-contribution refusals |
| Old rate name in detector | f86d26..202ad5 /38197 | Exactly old direct/alias; decoy and new-name controls pass |
| New rate name in detector | 74e65f..e1c67c /88922 | Exactly new direct/alias; decoy and old-name controls pass |
| Producer mask input | 95885c..92b7ad /52266 | All three masked-rate profiles and the invalid-mask refusal |

Every probe exits2 on actual assertions, not a compile failure. The transition
probe isolates rankCandidate; it is not independent severing evidence for each
of the other three transition sites. Candidate projection fires independently
when projection alone is severed:6.006 vs6.024018 after a correctly applied
purchase. A candidate-state failure is not cited as proof of that assertion.
Restoration hashes d6a026:
runner250b46f34c3b8b1920b25d030f96fd13b7054f41f302869c0cf1c9c232de6090,
solver3cc00ebf6e723951845cd28d78995cf6dbf8c6908001e49913d0ccb1182df39e,
simulationb02a8bd012793efb7be8ea4eba2313106969bb7ac0daa550f0b6f3c0c943abf1,
boundary450816c0aeb7e8817708bc27c30b145f10a2bbbd3b56a197e59243fba5ea89a1.

Corrected actual Reference observation b8f8f5..7b05b8 (session79577, exit0,
harness2.419s) retains the same source coordinates/experiment, both completed
careers and both Exits. Treated gate357000ms versus old355000; control stays
357000. Treated transition count16962 versus baseline16965; control16991.
The actual1.001 factor and ownership/starters are unchanged. This is an observed
changed policy trajectory, not proof that bonuses universally slow players or
that H4's starter-specific gate applies to this no-starter Reference row.
Observation mode current reports drift without writing or approving reports.
Mode baseline asserts the retained row; it does not execute old source.

Cold broader checks after all exact restores:
- Fast290086..c18249 (session39797, exit0): harness61.872s; role0.158s;
  Commons0.409s and harness guard pass. Exhaustive tests and the new opt-in
  observation explicitly skip here; the dedicated runs are separately recorded.
- Core4de7a2..af8dd7 (session56679, exit0): full core -count=1 and vet pass;
  production40.789s, save0.274s, transport13.255s. Narrow approved listener
  escalation only. Decimal golden vectors and version-source test execute in
  this core population. Separate legacy Pitch corpus alias is cached; its
  same test already executes cold in the full Pitch package. Formulas/API bytes
  regenerate unchanged; route/Commons boundaries pass. No fresh SQL claim.
- Composite14e741..56b36a (session98309, exit2): typecheck zero errors/warnings,
  build,8106client tests/134existing skips and shell boundary pass; TS decimal
  vectors execute. Kernel source parity/history-check setup reaches unchanged
  historical RP-131/50a3a514 and fails. No whole-client/CI green claim.
- Remaining1d168f..c7e977 (session74340, exit0): topology plus13negative
  controls, combat/meters/achievements/cosmetic/payment boundaries and copy pass;
  copy has657keys and610existing orphan warnings. Deployment manifest unchanged.

Complete exhaustive43361e..f8ba37 (session48713, terminal exit2), not cancelled:
H4's97treated/control pairs finish in145.25s; the same six Casual seeds
1/6/8/11/24/25 still fail strict-sooner. The report-byte comparison also fails
because the corrected instrument's Reference row differs. All970H5baseline/
leave-one-node-out arms finish; H5 then rejects retained report drift in641.10s.
Total786.459s. No node-classification error is emitted, but this is NOT H5
completion: epsilon/run4/provenance obligations remain. No regeneration flag
is enabled. Failed strict reproduction does not authorize retuning or updating
expected bytes. Separate report-provenance/refresh work must be predeclared.

ed8fd2 confirms unchanged H4/H5/threshold/original-corpus SHA256 respectively:
648f36d685c07471830692bf4c55810fc6c95ed4dcd00306d75a36697c2a9aa8,
4ed79054baa69389fff34cf850327c07b882b6b4aa507c25dc4d3d13808ab63c,
46a8fea612ff7e777a38c54166e542dff33e92ec18380fdf09ff4bdd1df72808,
f9b129e36af5b536f67c7eddb8d6088c76ce5cb0b170cfeec6a5172e8a009782.
No source/record edit occurs while any handle is live; all are terminal before
canonical docs, ledger, board, queue and plan reconciliation. No box is flipped.
Legacy v14/isolated generated-starter fixtures are not modern played/SQL careers.

RP-265 and RP-266 are corrected locally, not designated-approved. Complete span
afterd18d4e09, both predeclarations included, requires Claude. All earlier review
obligations and CHANGES REQUIRED findings remain. Next safe accepted work:
RP-262's actual H3 multiplier witness, then RP-264's false-exclusion oracle.
Report provenance/refresh is a distinct range; no AC13/15, literal/epoch mint,
owner copy, schema/migration/CI policy change, SQL/cleanup, archival, push or
full nine-tier/platform/release promotion. Goal active.

Final restored-source cold102cfb..d30649 (session53885, terminal exit0) again
executes all16Reference/20producer/six boundary controls and actual caller scan:
harness0.356s, production0.387s. Whitespace/gofmt checks ac364d/7d6be3 pass.
All subsequent record edits are after that handle ends; source bytes unchanged.

## 2026-10-06 — Reference correction checkpoint first filter

Review by: Codex (implementer self-review / first filter).
Recorded by: Codex.
Reviewed range: d18d4e09..6794a6b4, all three commits and all nineteen paths,
including both predeclarations, implementation, kernel sources and records.
Verdict: PASSED bounded first filter; NOT designated independent approval.

Reviewed the canonical masked producer's validation and nil-wrapper contract,
all four solver transition sites/advance/projection, copied factory input,
actual waiting/legal-command rate, all42new observations and opt-in career
observation. No balance/literal/report/schema/CI-policy/owner-copy changes;
watched files carry the real kernel161 simulation-input capability in the same
commit. No box flips, archived moves or existing verdict rewrites. Source/rate
scope does not prove modern active-play, SQL, natural careers, H3 or full R10.
Recorded report drift and historical RP-131 remain failing, not redefined gates.

Committed-HEAD cold544ae8..43e023 (session63797, terminal exit0) executes all
16Reference/20producer/six boundary controls and actual caller scan again,
harness0.190s, production0.247s. The broader cold and complete exhaustive
populations are the executed runs above, not a new whole-CI verdict. 77c18b
confirms clean main ahead8, whitespace-clean range;4976d3 confirms EOF-only log
addition. 3e544e reconfirms retained report/threshold/original-corpus hashes.

One separate source lead is registered, not silently corrected: RP-267.
1b4408/8bde66 show Reference adopts ranker candidate states and performs actual
bank advances with Routes-only dependencies. Ordinary runtime advances/intents
use runtime.lifetimeHook, whose AfterAccrual calls the canonical
prestigecore.AccumulateLifetimeValue. Missing paid-production accounting here
could change H1/payout/threshold/career evidence. No new numerical mismatch is
inferred. RP-265's independent rate/state calls intentionally match the same
Routes-only contract; they cannot be cited as a lifetime-accounting witness.
The next safe accepted R10 work therefore predeclares and measures this lead
before RP-262 H3, then RP-264. No extra implementation in this record.

No unresolved first-filter defect inside the bounded frozen-input correction;
whole new span afterd18d4e09 through this record still needs Claude. All earlier
reviews, report provenance/refresh and actual persistence/AC13/15/release gates
remain. No live handle, mutant, checkbox, mint, archival, cleanup, push or
release promotion. Goal active.

## 2026-10-06 — RP-267 Reference lifetime-accounting diagnosis predeclared

Previous turn is progress: RP-265/266's correction and report drift are committed
through973d983c. 37e121/349844 confirms clean main ahead9, no intervening change
or live handle. RP-267 remains a source lead, not an executed finding.
48d31a/e10cbc confirms accepted R10 H1's observable paid Reputation depends on
canonical prestige lifetime accrual; the existing simulation dependencies already
support the exact firstHourLifetimeHook used by the ordinary runner.

Diagnosis/test-only population before measurement: reuse the complete-source
RP-265 isolated legal v14 run3 fixture (actual tree/starters, cash1e3, five
generated/zero purchased Beige Towers). Three explicit frozen-input arms are
no row, synthetic unit and actual tree-derived1.003. For each, exercise actual
Reference candidate adoption after one second of lazy production and actual
bank dispatch under the declared zero-cash counterfactual. Determine the
Reference's chosen legal candidate with its existing policy, then construct
the independent intent through catalog class/upgrade lookup (not a copied
prefix rule). The oracle is canonical SimulateTransition/SimulateAdvance with
runtime.lifetimeHook and identical frozen input, cursor, revision and selected
intent. Compare complete encoded Company states; explicitly require accrued
LifetimeValue5/5.015, not merely disagreement with the consumer. Candidate
selection/optimality is not tested by taking its chosen ID as the coordinate.
The Reference consumes that candidate state; the observation is not just a
hypothetical candidate left unadopted. Ordinary advances/intents are controls.

Retain three direct zero-elapsed candidate controls, one per frozen-input arm:
canonical hook and no-hook outputs must agree when no production elapsed.
Starter grants remain uncredited to lifetime production; do not fabricate a
paid starting balance. No naturally earned, SQL or modern active-play claim.
Any fixture/admission error is setup, not an accounting defect. Cold baseline
must execute the six actual-branch comparisons and three zero-elapsed controls.
Before a repair, separately observe the current actual Reference seed0 treated/
control career with the original threshold1e5, experiment, source coordinates,
both Exits and complete outcomes; the existing opt-in current mode writes no
reports. It does not imply H4 or full first-hour distributions passed.

First range is diagnosis only: no source correction based on this start entry.
If the numerical omission is confirmed, record it and predeclare the exact
corrective range under accepted R10 before wiring hooks. In particular RP-265's
existing independent full-state oracles deliberately had Routes-only dependencies;
if actual Reference begins carrying the ordinary hook, those calls must be
reconciled to that same canonical dependency, without dropping state assertions
or claiming the earlier range proved lifetime accounting. Hook changes must
preserve simulation ablation semantics and remain absent from standalone suites.
Historical first-hour/H4/H5/threshold reports and formulas stay untouched.

No H3 verdict, threshold/purchase-policy/balance change, report regeneration,
epoch mint, owner copy, SQL/cleanup, checkbox, archival, push or release
promotion. New complete span after973d983c still requires Claude. Goal active.

## 2026-10-06 — RP-267 accounting omission reproduced; repair predeclared

Cold353e97..5d7320 (session37907, exit2) executes all nine initial observations:
six actual adoption/bank comparisons fail, three zero-elapsed controls pass.
The complete control population e49818..d6f5a1 (session65431, exit2) executes
twelve observations: the same six failures, plus all three ordinary-intent and
three zero-elapsed controls passing. No setup error or skipped comparison.
No-row/unit actual Reference credit0 instead of5; tree credit0 instead of5.015.
Full encoded states disagree, while the known-value canonical hook and actual
ordinary intent controls agree. Candidate selection is only a chosen legal
coordinate; no optimality claim. All handles end before this record.

Pre-repair actual Reference observation223458..7ce8a3 (session66184, exit0,
2.438s) records original experiment/threshold1e5 and complete fixture hash
sha256:3625eddb73da494574a2031fd483e93af236aa9b1f14482e5b51ad6ea3d0f7b0.
Both careers complete with two Exits. In both, scripted lifetime1.26604673417e6
credits2; elective lifetime3.34889972124e6 credits0. Treated owns onlyunlock.p05,
factor1.001, no starters and gate357000ms; control factor1 and gate357000ms.
Transitions16962/16991. All seven preceding milestones are recorded in the
full output, not reduced to the run3 gate. Existing v1 comparison is already
false after RP-265; this baseline does not recover report authority.

Corrective authority is accepted R10 H1/H2/H4's actual paid-Reputation observation
through the existing canonical lifetime accrual, not a new payout formula.
Separate bounded implementation range, predeclared before source changes:
1. Add an internal optional production.AccrualHook to RelevanceSuite. The
   composed Reference factory supplies runtime.lifetimeHook. Standalone suites
   retain nil, with no new wire/report or configuration artifact.
2. Bind this hook at all four solver transition sites and action-free advance;
   actual bank advances supply runtime.lifetimeHook directly. Reuse the existing
   simulation wrapper and ablation policy; do not rebuild lifetime arithmetic,
   served offline bookkeeping, receipts, quotes or policy selection.
3. Reconcile exactly the three independent full-state production calls in
   RP-265's tests to the same now-required hook. Keep complete state equality,
   frozen-rate assertions, masks and zero controls; do not suppress LifetimeValue
   differences. The earlier Routes-only proof remains honestly limited.
4. Retain all twelve diagnosis observations. Add per-arm direct ranker advance
   and effect-masked advance controls (three each), comparing canonical state
   with explicit5/5.015 and0 lifetime credit. No elapsed-time or starter credit
   fabrication. Ordinary intents remain healthy controls.
5. Independently omit factory hook, candidate-purchase hook, advance hook and
   actual-bank hook. Require compiling failures and exact SHA restoration
   before the next probe. Also sever the shared first-hour hook to a no-op:
   explicit known-value assertions must fail despite the shared oracle path.
   Omit advance's mask in a separate probe to show the masked controls fire.
   Run normal focused checks afterward; no editing while any handle lives.

Before/after the repair, execute the existing actual Reference seed0 career
observation with complete outcomes/coordinates, Exits/lifetime/deltas,
ownership/starters/bonus and all seven recorded milestones. Report drift even
if gate clocks do not move. Then run cold fast/core/vet, client/type/build and
boundaries, and the full97pairedH4/970armH5 strict reproduction. Existing reports
are expected to be potentially stale; no regeneration flag or expected-byte
edit is authorized. A red full run is not a balance verdict or permission to
weaken its criteria. No claim that a Reference observation proves all H3
distributions, unit/non-unit milestone sensitivity or all persona accounting.

Only harness code/tests and docs/records are intended. They are outside the
current watched kernel prefixes, so kernel161 stays; no false bump. If actual
scope needs a watched runtime edit, stop and predeclare it instead. Historic
RP-131 stays red; no CI-policy waiver, SQL/capacity/cleanup, owner copy,
threshold/policy/literal change, report refresh, box, mint, archival, push or
release promotion. Full span after973d983c needs Claude. Goal active.

## 2026-10-06 — RP-267 observer correction executed

Corrective predeclaration821b867b follows the executed diagnosis underb15d8eef.
The internal optional RelevanceSuite hook receives runtime.lifetimeHook from
the composed factory. All four existing solver transition sites and action-free
advance bind it; actual bank advance binds the same observer directly. Standalone
suites keep nil. The exact existing simulation wrapper and canonical
prestigecore.AccumulateLifetimeValue are reused; no new arithmetic, served
offline bookkeeping, quote/policy or public report/wire field. This is harness-
only; no watched kernel prefix changes, so161 remains unchanged.

The three RP-265 full-state oracle calls now explicitly carry the same observer
as the corrected consumer. They still compare complete encoded states; no
LifetimeValue suppression or looser rate assertion. The earlier Routes-only
proof is not retroactively represented as accounting evidence. Three direct
ranker advance and three effect-masked advance controls complete the new
18observation population. All zero/ordinary/frozen-input controls remain.
Normal cold ae7784..a3e568 (session52309, exit0) executes all18lifetime and
16frozen-input observations, harness0.576s. No setup, compile or skip failure.

Six independent compiling probes, each root make test-go with -count=1 and
exact SHA restoration before the next probe. Every handle ends with exit2:

| Omission | Executed output / handle | Fired lifetime observations / RP-265 comparisons |
| --- | --- | --- |
| Factory hook | b46bb2..3235c2 /98595 | Six adopted/advance observations and six prior advance/purchase comparisons |
| Candidate purchase hook | 45add2..e38ba6 /74814 | Three adopted-candidate observations and three prior purchase comparisons |
| Candidate advance hook | 731442..086280 /35929 | Three direct advance observations and three prior advance comparisons |
| Actual bank hook | 55c19c..d7f0f7 /26295 | Three actual bank observations and three prior bank comparisons |
| Shared hook → no-op | 9f94db..c07cca /42397 | All twelve explicit nonzero known-value assertions fail; the old16comparison tests still pass |
| Advance mask | ebfff3..196b8f /26161 | Exactly three masked-advance observations; all prior16comparisons pass |

The shared-hook probe demonstrates why comparing two calls to the same broken
observer alone is insufficient. New known5/5.015 values fail in the oracle
before consumer equality, while zero-elapsed/masked controls survive. It does
not claim an independently broken served hook; that source is unmodified.
The candidate transition probe isolates rankCandidate; it is not severing
evidence for the other three transition sites. Prior diagnostics report equal cash when LifetimeValue
alone differs; the new lifetime diagnostics explicitly print the affected field.
Exact restoration hashes3d4463:
runner9ec0e95f593eb0f237f7e2a3b00f14dae256bc9def2e404b5cf8f709d623a39a,
solverfbed47e2f8b14358e4bbe846e9ae46562d96edee3dcd4e026a4c6c39d71ee994,
suite655f989892f13f4505766b9b7e9242233af2bc7b94bc03019ec63cc6bd368865,
old-diagnostic2c1e1300ccf592adcbe1392f1e452f3eeb4241d6de64871b19169a9605b5e329,
new-test98f0e9cb9802017b9bf5f3650ef92bcf8e0bf9b9c5e7fede08d85aa9512dffbb.

After-repair actual Reference observation2b188b..22639a (session97228, exit0,
3.802s) has exactly the same source coordinates/experiment as the pre-repair
223458..7ce8a3. Both treated/control careers complete with two Exits. Scripted
lifetime1.26604673417e6→1.4605083614e6; elective3.34889972124e6→3.54431965065e6.
Paid deltas remain2/0 and Founder level2. Treated keeps unlock.p05, no starters,
factor1.001 and available1; control keeps no nodes/starters, factor1 and
available2. Run3 gates both357000ms, transition counts16962/16991. All seven
preceding clocks remain identical for this seed:0/10000/66992/356000/900000/
346000/4500000. Ending/ownership observation is unchanged. These are not all
H3distributions, natural SQL careers or a tiny-factor sensitivity verdict.

Broader executed checks, after all exact restores:
- Fastcee5eb..e60645 (session22338, terminal exit0): harness92.066s,
  role0.174s, Commons0.504s and guard pass. Exhaustive tests and explicit career
  observation skip here; their actual dedicated runs are separately recorded.
- Core3883b4..c55b78 (session31201, terminal exit0): full core -count=1 and
  vet pass; production47.445s, save0.299s, transport13.400s. Narrow approved
  HTTP-listener escalation only. Go numeric vectors/source version execute;
  the separate legacy Pitch alias is cached but its full package was cold.
  Formulas/API regenerate byte-identical; route/Commons boundaries pass.
- Composite15cc0c..3a3ab8 (session69208, terminal exit2): typecheck zero
  errors/warnings, build and8106client tests/134existing skips plus shell pass.
  TS numeric vectors execute. Historical RP-131/50a3a514 still fires; no
  whole-client/CI green verdict and no guard/budget waiver.
- Remaininga78fd5..393c0b (session60510, terminal exit0): topology and13
  negatives; combat/meters/achievements/cosmetic/payment/copy pass. Copy has
  657keys and610existing orphan warnings. Deployment manifest unchanged.

Complete exhaustive811f69..c66bd3 (session36620, terminal exit2), not cancelled:
all97treated/control H4pairs finish in212.58s. The same six Casual seeds
1/6/8/11/24/25 still violate strict-sooner at275000/425000/405000/265000/
360000/430000ms. Retained H4byte comparison remains red at report drift.
All970H5baseline/leave-one-out arms finish in623.44s; its retained byte comparison
also remains red. No node-classification error is emitted, not full H5 proof.
Total836.130s; no report regeneration or expected-byte update. These fired
comparisons do not authorize retuning. H4's genuine criterion and H5's epsilon/
run4/provenance remain open.

5dee84/692bde additionally shows the retained first-hour Reference H1 fields
contain pre-correction lifetime values. The threshold test reads this retained
source and recalculates; it does not rerun the corrected producer. Its green
test is historical-source consistency, not fresh H1/H2/calibration proof.
The observed career uses threshold1e5/paired tree; the retained first-hour
report uses the live1e12/base bundle, so this is not a byte-reproduction claim
for that distinct first-hour experiment. Fresh source/report lineage and full
current-producer reproduction stay in the separate RP-263 route. No claim
that Casual/Chaos paid distributions or threshold recommendations changed.

13a877 confirms unchanged first-hour/H4/H5/threshold/original-corpus SHA256:
1f7c774d40e0e6b88dc23c37fa6e734c8a0c9d5e8e05416eb29c6f342180b7a6,
648f36d685c07471830692bf4c55810fc6c95ed4dcd00306d75a36697c2a9aa8,
4ed79054baa69389fff34cf850327c07b882b6b4aa507c25dc4d3d13808ab63c,
46a8fea612ff7e777a38c54166e542dff33e92ec18380fdf09ff4bdd1df72808,
f9b129e36af5b536f67c7eddb8d6088c76ce5cb0b170cfeec6a5172e8a009782.
No source/record edit while any handle is live; all end before canonical docs,
ledger, queue, board and per-RFC plan reconciliation. No checkbox changes.

RP-267 is corrected locally, not designated-approved; whole span after973d983c
includes both predeclarations and requires Claude. Previous range afterd18d4e09
and all earlier reviews remain separate obligations. Next safe accepted work
is RP-262's honest H3 unit/no-row/tiny-factor population, then RP-264's false-
exclusion oracle. No source scope drift, kernel/live formula/balance/schema/
migration/CI policy/owner-copy change, SQL or capacity cleanup, report refresh,
literal/epoch mint, AC13/15, archival, push or full1.0 promotion. Goal active.

Final restored-source cold rerun9cf13a..d48ec1 (session94035, terminal exit0)
executes all18 lifetime-accounting and16 frozen-input observations in0.393s.
All six compiling mutations were restored before this run. No source or record
edit occurred during the live handle. This is a pre-commit check, not a
designated verdict or a committed-HEAD claim.

## 2026-10-06 — RP-267 bounded implementer first filter

Review by: Codex (implementer, self-first-filter).
Recorded by: Codex.
Reviewed range:973d983c..9743dcb7, all three commits and all14 changed paths,
including both predeclarations, five harness source/test files, canonical docs
and synchronized ledger/queue/board/plan records. Verdict: bounded first filter
passes; this is NOT the designated independent cross-party verdict.

The committed-HEAD cold focused run53f483..307cb5 (session19657, terminal exit0,
harness0.329s) executes all18 lifetime-accounting and16 frozen-input cases.
All six prior compiling negative probes and exact restores are recorded above.
Known-value assertions catch the shared observer fault even when all16 earlier
comparison tests survive. Candidate-purchase severing covers rankCandidate,
not all four transition sites. The optional hook stays internal; standalone
nil behavior, frozen inputs and effect masks remain. No served runtime, watched
kernel prefix, formula, balance, schema, migration, CI policy or report bytes
changed. Kernel161 remains honest. No checkbox was promoted.

Actual Reference lifetime increases, while this observed seed's paid deltas,
ownership, ending and clocks remain unchanged. That is not SQL/player proof,
fresh H1/H2 population calibration or H3 tiny-factor proof. Full97-pair/970-arm
study remains RED on retained report drift; six genuine Casual strict-sooner
ties and H5 epsilon/run4/provenance remain. Composite client remains RP-131 RED.
Full973d983c..9743dcb7 needs Claude; prior ranges remain separate obligations.
No archive, mint, release promotion or push. Continue accepted RP-262, then
RP-264; goal remains active.

## 2026-10-06 — RP-262 H3 current-producer sensitivity study predeclared

Resume from clean ee1064a0 after the RP-267 correction and bounded first
filter. Authority is accepted Reputation Tree v1 R10 H3, not permission to
retune the tree or relax the tiny-factor criterion. Source: the existing H3
test reads two retained reports; normal first-hour runtime.external returns
nil outside career run3. There is no executed explicit unit/tiny population.

Question: do fresh empty-plan/no-purchase first-hour milestone distributions
match epoch8, does an explicit prestige factor1 remain a no-op, and does the
required factor1.000001 move an actual milestone? This is a synthetic input
sensitivity experiment, not a naturally earned Founder bonus or SQL career.

Population and controls declared before source changes:

1. Use the owner-ratified first-hour scenario/policies unchanged: all32 Casual,
   64 Chaos and1 Reference seeds, horizon7200000ms, all seven must-reach clocks.
   Reuse the retained epoch8 experiment: purchased minimum200, burnout2,
   route bonus50, seed capital1e4 and10 generated towers. Live threshold1e12;
   no reputation purchases, no career or Tier2 mode.
2. Five complete arms (485 seeded runs,3395 milestone observations): fresh
   epoch8 base/nil, complete paired Reputation fixture/nil, paired explicit
   prestige factor1, paired factor1.000001, paired factor2. The last is a
   separate strong positive control, never a substitute for tiny sensitivity.
   Resolve the three explicit inputs through the fixture's declared frozen
   Founder source, preserving all other catalog bytes and threshold1e12.
3. All runs must complete, carry exactly seven named/non-nil clocks and two
   valid Exit samples with zero paid Reputation at this threshold. Enumerate
   every declared policy/seed exactly once; reject empty/duplicate/missing
   rows, null clocks, source/experiment mismatch, failed or truncated runs.
   Report full row coordinates and clocks, not just selected successes.
4. Fresh base and paired/nil must match retained epoch8 clocks and ending
   samples; unit must match paired/nil clocks, ending, exits and transition
   counts. Compare all per-seed clocks and sorted per-policy distributions,
   not only the seven existing aggregate envelope entries.
5. Factor1.000001 must change at least one actual clock and its associated
   per-policy distribution; factor2 must independently do so. If the tiny
   criterion fires because discrete clocks do not move, complete and record
   the RED study. No lowering tolerance, larger substitute, hidden changed
   clock, finer invented clock or silent population/horizon expansion.

Bounded harness-only input seam: an unexported optional FirstHourSuite
contribution slice, copied once into each ordinary runtime; runtime.external
returns it only for that diagnostic mode. Nil preserves every current caller.
Reject a non-nil diagnostic input combined with career or Tier2 before running
so it cannot overwrite a real frozen career bonus. No exported API, artifact
identity/schema/wire field or new live multiplier authority. Validation still
uses canonical production admission. The study logs source/factor coordinates
separately; existing RunKey is not falsely claimed to encode this extra input.

Opt-in observation selector analogous to the existing Reference observation,
outside the push fast lane; no CI change. A selector other than the declared
full study fails; default explicitly skips. A fast helper population exercises
coverage/clock/sensitivity refusal using non-production synthetic rows.
Demonstrate compiling failures for diagnostic-input omission, wrong factor,
and census/oracle severing as applicable; exact restore each before the next.
Run actual no-row/unit/strong controls to distinguish dormant wiring from a
tiny-factor criterion failure. No edit while any handle lives.

Only harness sources/tests and canonical docs/records; watched kernel prefixes
stay unchanged, so161 remains. Cold focused checks, fast harness/core/vet and
the full opt-in study are required. Existing full H4/H5 RED evidence remains
separate; no need to rerun970 arms solely for a default-nil first-hour seam
after its ordinary/career controls and fast population pass. Existing client/
CI RP-131 remains RED; no waiver. Reports/corpus/balance/policy/owner copy stay
byte-identical. RP-264, report lineage/H1/H2 fresh calibration, H5epsilon/run4,
SQL/browser and prior Claude reviews remain separate. No box/mint/AC13/15,
archive, push or 1.0 claim; complete new span afteree1064a0 needs Claude.

## 2026-10-06 — RP-262 full study RED; bounded probe control predeclared

All handles ended before this record. The complete opt-in study680609..8f803f
(session64870, terminal exit2,341.867s) executes all485 runs /3395 clocks.
Epoch8, paired/nil and explicit unit reproduce the seven per-seed clocks and
ending samples; unit also matches Exit observations and transition counts.
No comparison error was emitted. Every arm has97 valid completed runs,
679 clocks, two observed Exits per run and zero paid Reputation at threshold1e12.
Tiny factor1.000001 moves0/679 clocks and no distribution; strong factor2
moves289/679. The exact H3 tiny criterion fires; no stronger substitute or
clock/cadence/tolerance change is adopted.

The tiny input was consumed: Reference seed0 scripted/elective lifetime moves
1.4605083614e6/3.54431965065e6 →1.46050991148e6/3.54432330408e6, with clocks
0/10000/66992/356000/900000/346000/2700000 unchanged. Its transition count
11916→11915 is not called a milestone. Strong Reference clocks are
0/10000/44072/182000/900000/177000/2700000, lifetime
3.5053281185e6/7.66194816889e6 and8947 transitions. These are synthetic
counterfactual observations, not naturally earned bonuses or SQL acceptance.

Additional small control declared before editing: execute these four paired
Reference seed0 inputs (none/unit/tiny/strong) under the same full2hour ratified
policy, comparing all seven clocks, both lifetime samples and transition counts
with the just-completed observation. This is a fast input-consumption regression
control, explicitly NOT the485-run study or an H3 passing verdict. It leaves
the full opt-in selector/full-population criterion untouched. Its small row
is for inexpensive independent compiling probes: omit runtime input-copy,
omit runtime.external's diagnostic consumer, replace only the control's strong
input with unit. Each must fail tiny/strong observations rather than rely on
the already-fired full tiny criterion. Separately omit the oracle's expected
row deletion and admit unchanged sensitivity; malformed-duplicate and unchanged
negative controls must respectively fail. Exact SHA restores before each next
probe and cold normal run afterward; no edits while handles live.

Baseline focused3168f1..c2d674 passes in3.233s:27 oracle/refusal children,
the existing live Reference first-hour control, live Chaos Exit observer and
18 lifetime profiles. Fasta39ae2..7e1c11 exit0: harness63.896s, role0.153s,
Commons0.383s and guard; opt-in study is explicitly skipped there. Core3e515b
..c77ce1 exit0: vet/full cold core (production47.108s, transport13.334s,
save0.293s), numeric vectors, generated API/formulas byte-identical. No SQL.
Original predeclaration64badd7b remains; only this narrow new test supplement.

## 2026-10-06 — RP-262 discrimination and checkpoint

Supplement3d7c93ff's four actual Reference input-consumption controls pass
ea1807..5be0d9 (session51958, terminal exit0,4.770s). The27 oracle/refusal
children include23 malformed-population refusals, unchanged sensitivity,
one valid changed-clock control and two career/Tier2 override refusals.
The ordinary healthy synthetic population also passes admission. These
synthetic oracle rows are not production evidence; four Reference rows are
actual headless runs, not SQL/player workflows or the all-seed study.

Five independent compiling omissions, each with exact four-file SHA restoration
before the next. All root -count=1 runs terminate exit2, no build errors:

| Probe | Executed chunks | Fired controls |
| --- | --- | --- |
| Runtime diagnostic-input copy omitted | ff09b4..196fe9 | Actual tiny and strong lose their observed production/clock behavior |
| runtime.external diagnostic input omitted | 0d881b..b0fa6e | Actual tiny and strong lose their observed production/clock behavior |
| Strong control factor changed2→1 | 282c85..07d2a7 | Only actual strong row fails; a test-input corruption, not a production mutation |
| Expected census row consumption omitted |121579..caf4d5 | Duplicate-row negative control is silently admitted and fails |
| Unchanged sensitivity admitted |e059ae..7b1087 | Unchanged-population negative control fails with fabricated sensitivity |

The small probes do not rerun485 observations or pretend to be that study.
The full sensitivity criterion remains RED as recorded above. No threshold,
clock, policy or data retune makes the probes succeed. All source hashes after
restoration68c030 match6f744e: first-hour suite
7fc67a7926e183575e08529500dff79b5d5ae2d84295e6031df70e30b7bf07f8,
runner a68584945d354eaa1f9bf963b0224a647f87d09984a8d377f4b821601d7fbd95,
career6fdbd87b2a0954057533516f931ca5e2c6cdfd6e355cafe0c3f4991af9739edf,
new-test11c225df64a0a30d7d5c7cbccde7d53443a8c4c7e301c7fe7d8a7077d5259b1a.

Restored cold527bb1..aa9698 (session48978, terminal exit0,4.863s) executes
all31 new named children plus34 previous lifetime/frozen-input observations.
Final fast0f1a37..a3d421 (session92037, terminal exit0): harness59.220s,
role0.339s, Commons0.362s and guard pass. Full opt-in study explicitly skips,
not green evidence. Final narrow vet e6eba5 exit0 covers the added test.
Earlier full core/vet result remains on the same restored harness sources;
only this small test was added afterward. No new client/SQL/hosted/browser
claim: earlier composite client remains RP-131 RED and SQL/capacity is held.

Default selector973ba9..d8d729 (session91047) explicitly SKIPS then exits0;
unknown single selector dc99b2 rejects exit2 before any seed execution. Neither
can masquerade as the full study. No CI configuration, budget or lane change.
68c030 also confirms first-hour/H4/H5/threshold/original-corpus bytes unchanged;
their hashes match the RP-267 checkpoint. No report refresh or epoch mint.

Reconciled canonical docs, backlog RP-262, current-state/long-term board,
execution queue and per-RFC plan with the fired criterion, not H3 approval.
All handles ended before these source/record edits; no checkbox changes.
No served runtime, watched kernel prefix, live math/balance/schema/migration/
CI policy or owner-copy change. Kernel161 remains unchanged. Complete new span
afteree1064a0 includes both predeclarations and needs Claude. Previous RP-267
973d983c..9743dcb7 and all prior spans remain separate review obligations.

R10 H3 author reconciliation is required before closeout; no weakened criterion
is inferred. Next safe accepted work RP-264 false-exclusion/accounting oracle;
H1/H2 fresh calibration/report lineage, H4's six Casual ties, H5 epsilon/run4,
SQL/browser/reviews and full1.0 remain. Goal active; no archive or push.

## 2026-10-06 — RP-262 bounded implementer first filter

Review by: Codex (implementer, self-first-filter).
Recorded by: Codex.
Reviewed range:ee1064a0..fab074a1, all three commits and all13 changed paths:
both predeclarations, four harness source/test files, canonical docs and
ledger/queue/board/plan/log reconciliation. Verdict: bounded instrument and
record first filter passes; exact R10 H3 acceptance criterion remains RED.
This is NOT the designated independent cross-party verdict.

Committed-HEAD cold5e3c3a..84fd14 (session70251, terminal exit0,4.866s)
executes65 named children:27 new oracle/refusals, four actual Reference
consumer controls and34 prior lifetime/frozen-input observations. The five
earlier compiling probes and exact restores cover input-copy/consumer,
wrong-factor, duplicate-census and unchanged-sensitivity refusal. They do not
pretend to rerun the485-run study. That complete study was executed on the
same three instrument source files before the small control was added; it
is not relabelled as a new committed-HEAD full-study run.

The diagnostic input is unexported, copied per ordinary runtime, and rejected
with career/Tier2. Nil preserves existing callers. Canonical frozen admission
and production math remain; no report-key authority is fabricated. Historical
reports/corpus and kernel161 stay unchanged. Source/record editing occurred
only after all handles ended. No checkbox, retune, refresh, mint or archive.

Unit/no-row neutrality passes; tiny0/679 versus strong289/679 records the fired
criterion honestly. Changed lifetime or transition count is not a milestone.
R10 H3 author reconciliation, H4/H5, report provenance/fresh H1/H2, SQL/capacity,
RP-131 and all previous reviews remain. Completeee1064a0..fab074a1 needs Claude;
earlier973d983c..9743dcb7 remains separate. Next RP-264; goal remains active.

## 2026-10-06 — RP-264 false-exclusion/accounting diagnosis predeclared

Previous goal turn made progress: observer correction9743dcb7 and full H3
studyfab074a1; H3's exact tiny criterion remains RED. Resume clean db8398a3
(50b1ee/ba0215); no live handle. Accepted R10 H4 requires strictly sooner
at every eligible starter seed. Current evaluateReputationCareerGate skips
any nonempty Excluded before inspecting clocks/reason, then dereferences
SavedMS in a successful two-clock row. RP-264 is still source-only.

Additional source finding RP-268 (bfdcb1/b7dc5e/901ebe): the accepted named
career-scenario-v1.json path is absent, and the only authored tree testdata
is fixture-v1.json. Existing tests compose the ratified first-hour population.
Recorded in the ledger immediately; the author must reconcile this artifact
authority separately. No new scenario/policy or author-body rewrite here.

Test-only diagnosis, before changing the existing oracle:

- Synthetic rows exercise the exact current helper, never called production
  measurements. Legal controls: earlier two-clock treatment with exact saving,
  earlier treatment/control unreached with no finite saving, documented
  both-unreached exclusion, tie and slower treatment (both must violate),
  missing treatment/control reached (must violate), and a no-starter row
  outside H4 timing eligibility. Unknown/no/false exclusion reasons never
  become owner exemptions.
- Corrupt exclusions: documented reason with tie/slower/earlier finite clocks,
  either one-arm-missing clock, unknown reason with both clocks absent, missing
  reason with both absent, and unknown reason on a no-starter row. Exclusion
  admission applies to every row; no-starter is not permission for a false
  label. The existing valid no-starter row still stays outside timing counts.
- Corrupt accounting: missing or wrong SavedMS for finite clocks (including a
  claimed exclusion), finite SavedMS on a one-arm-unreached or both-unreached
  row, and negative observed clocks. Require a visible violation rather than
  panic or silently admitted/suppressed values. Panic is a separately disclosed
  baseline failure, not discarded as fixture setup.
- Mixed/repeated census: accepted earlier finite row, valid both-unreached
  exclusion, tie and no-starter; counters/saving populations must derive from
  this invocation. Seed stale counters and call twice to detect accumulation.
  Invalid rows must produce violations and no successful saving distribution;
  no stale counters, fake exclusions or invented zero saving may hide them.

Predeclare exact matrix in the new diagnostic test; run cold against the
existing helper, record which corruptions were admitted/panicked and which
healthy controls passed. Do not interpret intentionally malformed rows as
numerical balance findings. If confirmed, record it and separately predeclare
the bounded test-side oracle repair before editing the old helper.

The only current exclusion route is the existing documented literal
run3_gate_beyond_ratified_horizon_in_both_arms with both clocks nil and no
finite saving. Preserve that report wording and the ratified7200000ms horizon.
Clock nil means the completed producer did not reach the gate in that horizon,
not a manufactured future timestamp. Do not compare a run-attended clock to
the global wall horizon or extend it. No statistical weakening, retune,
report regeneration, new public schema/resource/balance or live runtime change.

After a lawful repair: compiling exclusion/reason/clock/accounting/counter
omission probes with byte-exact restores, cold fast/core/vet and the complete
existing97-pair H4 and970-arm H5 strict study; retained reports remain untouched
and expected drift/real H4 ties must be recorded RED. Source/test/record edits
only after every handle terminates. Harness test files stay outside kernel
watch; kernel161 remains. Whole span afterdb8398a3 needs Claude. H3 author
reconciliation, RP-268, report lineage/fresh H1/H2, H5epsilon/run4, SQL/capacity,
all prior reviews and full1.0 remain. No box, mint, archive, cleanup or push.

## 2026-10-06 — RP-264 reproduced; test-side oracle repair predeclared

Diagnosis7b0046..89bb3a (session65357, terminal exit2,0.295s) executes all28
profiles on the original evaluateReputationCareerGate undera0aab5a6. Eight
healthy controls pass. Eighteen malformed rows and both mixed/repeated census
controls fail. Twelve malformed rows return no violation (five falsely
excluded finite/one-arm cases, unknown both-nil and no-starter exclusions,
wrong finite saving, invented saving with unreached control, saving on a
both-unreached exclusion, negative treatment and no-starter wrong saving).
Five already have timing violations but carry invalid gated counts/diagnostics:
unlabelled both-unreached, wrong tie saving, invented saving with missing
treatment or both missing, negative control. Missing saving on a sooner row
panics at the dereference; disclosed separately, not discarded as setup.

Mixed false exclusion hides the tie alongside a healthy sooner row: zero
violations, gated1/excluded2 instead of violation1/gated1/excluded1. A report
seeded with counters99/77 grows to101/78 instead of deriving2/1 from the four
current rows. The eight controls distinguish a dormant oracle from these gaps.
No production/balance conclusion follows from deliberately corrupt test rows.

Corrective predeclaration before editing the old helper; accepted R10 H4 and
the existing visible both-arms horizon exclusion are the bounded authority:

1. Test-only private row-admission helper runs before eligibility. Reject
   negative observed clocks. A nonempty reason must equal the existing exact
   horizon literal and both observed clocks must be nil. Both-nil with no
   reason is invalid, not an invented observation that control reached it.
2. With both finite clocks, SavedMS must be present and exactly control minus
   treatment, even for a tie/slower/outside-starter-eligibility row. With any
   absent clock it must be nil: no manufactured saving for an unknown gate.
   Violations are visible and do not enter accepted/excluded/savings counts.
   Integer differences are safe after nonnegative int64 clock admission.
3. After admission, preserve starter eligibility and exact strict-sooner rule.
   A treated reached/control unreached remains sooner with no finite saving;
   a valid both-unreached starter row remains excluded and visible. No-starter
   rows remain outside H4 timing counts, not outside artifact honesty checks.
4. Reset the two derived report counters on every invocation, then count only
   admitted eligible rows. Preserve savings aggregation, report schema,
   original wording, policies,7200000ms horizon and non-vacuity caller guard.
   Centralize the existing reason constant in test-side producer and consumer
   without changing bytes; diagnostic oracle retains an independent literal.
5. Retain every diagnostic expectation. Independently omit reason guard,
   both-clock guard, saving presence/equality, nil-clock saving refusal,
   negative-clock guard, caller admission, counter reset and strict-sooner
   comparison as applicable. Require compiling failures; nil dereference is
   separately labelled if it occurs, never called a semantic refusal. Exact
   two-test-file SHA restores before each next probe, cold normal after.
6. Cold focused/fast/core/vet and full97-pair H4/970-arm H5 strict reproduction.
   Do not regenerate retained reports; record existing drift/six Casual ties
   RED, with unchanged exclusions and census if the producer still meets row
   admission. No H4 balance promotion follows from repairing the gate.

Only two harness test files plus docs/records. No product source, watched
kernel prefix, formula, balance, schema, migration, CI policy, owner copy,
report or corpus changes; kernel161 remains. No edits while any handle lives.
RP-268 author/artifact reconciliation and H3's fired tiny criterion are separate
requirements. Whole span afterdb8398a3 needs Claude; all prior reviews remain.
Report lineage/fresh H1/H2, H5epsilon/run4, actual SQL/capacity/browser and
full1.0 stay open. No box, mint, archive, cleanup, push or goal completion.

## 2026-10-06 — RP-264 full study ends; direct census logging predeclared

All handles terminal before this record. Restored focused646055..3cc565
(session20726, exit0,0.162s) passes all28 profiles plus the original gate test.
Nine compiling probes discriminate, exact e49c252c/03bda3a3 hashes restored;
saving-presence and ignored-admission probes cause a caught missing-saving
panic, explicitly not a semantic refusal. Other omissions fire semantic/
count comparisons; no compiler failure. Cold fast ef9ebe..09a725 exit0:
harness64.918s, role0.157s, Commons0.396s and guard. Cold core/vet59dee6
..e083d6 exit0: production38.528s, transport13.262s, save0.273s, numeric
vectors and generated API/formulas byte-identical. No fresh SQL/client/browser.

Complete8f5352..e9c740 (session88822, terminal exit2,787.881s): all97 H4
pairs finish131.03s, with the same six Casual ties and no new row-admission
violation. H4 retained byte comparison remains report-drift RED. All970 H5
arms finish656.69s, retained byte comparison remains RED; no classification
error emitted, not full H5 proof. No report update flags. Source/old reports
restore/match481d66. The denied read-only ps check0d8b84 yields no process
evidence; same confirmed live handle was polled to terminal, not restarted.

Observation gap: the current producer did not print GatedSeeds/ExcludedSeeds.
471517 independently inspects the historical97-row report (93gated/3excluded,
Casual4/14/21 valid both-nil exclusions) but is not a current-producer census.
Do not silently substitute that historical result for the required current
accounting observation.

Bounded additional test-side observation, declared before source changes:
after the existing gate computes its report, log total produced rows, derived
gated/excluded counts and each excluded row's actual reason/clocks/saving.
Observation only: no new gate, schema, population, policy, horizon, arithmetic,
selection or regeneration. Execute only the complete97-pair H4 test with the
existing explicit exhaustive flag through root make test-go, not a shortcut
or fake970-arm repeat. Rerun cold focused/fast/vet after adding these logs.
The existing full H5 result remains separately dated; no need to repeat970
arms for stdout-only logging. Report drift/six real ties remain RED.

Read-only14531b/d540c8/80731d additionally finds a threshold-source admission
lead: MeasureReputationThresholds checks outcomes but not the complete seed
population/source coordinates, and PaidReputationAtFirstElectiveExit takes
the first matching kinds without sequence/order/duplicate admission. No bad
input has been executed; record RP-269 as a source lead, not a confirmed
measurement defect, and predeclare diagnosis separately before changing it.
No source/record edits while any handle lived; previous goal progress remains.

## 2026-10-06 — RP-264 exclusion admission repaired; measured failures retained

All handles terminal before this record. Predeclarations a0aab5a6 (diagnosis),
5e3064fb (bounded repair) and8ef50ccb (direct census observation) precede their
changes. No product byte moved: only two harness test files, docs and records.
Private row admission validates nonnegative clocks, exact both-unreached reason,
finite-pair saving presence/equality and absent saving when either clock is
unreached, before starter eligibility. Counters reset every invocation. Existing
strict-sooner, treated-only and no-starter semantics remain; no new exemption,
policy, horizon or report schema. All28 synthetic profiles and the legacy gate
pass after stdout-only observation:2e55e4..5b70f7, session2113,exit0,0.295s.

Baseline7b0046..89bb3a had8healthy passes,18 malformed and2accounting failures:
12 malformed admissions,5already-timing-violating but invalid census/savings
rows,1missing-saving panic. Nine independent compiling probes, each cold and
terminal before restoration/next edit:

| Omission | Output range | Demonstrated fault |
|---|---|---|
| exact reason | a734ff..dbe3c9 | two unknown-reason refusals lost |
| both-clock exclusion requirement |32f2e4..061d69| five false labels plus mixed population |
| saving presence |e8cb09..636681| missing-saving panic caught, not semantic refusal |
| saving equality |c2b0b7..1c8b24| three inconsistent-saving profiles |
| finite saving on unreached clock |a95627..d1bdb4| three nil-clock profiles |
| negative clock refusal |7eb9df..ff7762| two negative-clock profiles |
| caller honors admission error |b50f14..4de345| eighteen bad profiles plus mixed; includes caught panic |
| census reset |58b2e3..43c9ff| repeated invocation counts |
| strict >= comparison |35ad92..74e9c2| tie, census and original gate test |

All nine exit2, no compiler failure. No separate omission of the unlabelled
both-unreached branch was executed; do not claim every guard was severed.
Probe-phase source SHA256 e49c252c1fa56cb14f68faf8c2812878254917bf5694422cdc583bc189ce943f
and diagnostic03bda3a3bece2b701ef3f72761d1fdc9a94c1a146d99c465d2cb1cadab0f4227
restore exactly (0ded61/481d66). The subsequent predeclared logging-only addition
changes the career-test SHA to502d0a344a8485b3318442641f711ea60dd8e9aa53a5238fd34db9eafa0bbcb9
(c88617); guards and diagnostic are unchanged. Do not relabel probe restoration
as the later source hash.

Full8f5352..e9c740, session88822,terminal exit2,787.881s:97 H4 pairs131.03s,
970 H5 arms656.69s. Same Casual ties:1=275000,6=425000,8=405000,11=265000,
24=360000,25=430000ms. No new admission violation/classification error emitted;
both strict retained-byte comparisons remain RED. Absence of an H5 error is not
full H5 proof. No update flags or old-report changes.

Actual census after observation supplement7719c1..369a03, session50842,
terminal exit2,146.952s (test146.72s): all97 current pairs,93 counted comparisons,
3valid exclusions:Casual4/14/21, each exact known reason and both clocks/saving
nil. Same six ties and retained-report drift remain RED. This is the current
producer, not the historical471517 census, and not another970-arm H5 run.

Cold core59dee6..e083d6 exits0 including vet, production38.528s,transport13.262s,
save0.273s, numeric vectors and unchanged generated API/formulas. Final fast
after logging5c19d2..d569ec exits0:harness72.051s,role0.264s,Commons0.416s and
guard. Final narrow vet174b98 exits0. No fresh SQL/client/browser/hosted claim;
historical composite RP-131 remains RED. Failed read-only ps0d8b84 was not
process evidence; the same live full-study handle was polled to terminal.

Kernel161/live math/balance/schema/migrations/CI policy/owner copy/reports/corpus
unchanged. No checkbox flips. Whole span afterdb8398a3, all three predeclarations
and implementation/records, needs Claude independently of all prior ranges.
Next accepted diagnosis RP-269 source/Exit admission and RP-268 artifact-authority
grounding. Fresh H1/H2/lineage, H3 author reconciliation, H4/H5, actual SQL/capacity
and full1.0 remain. Goal active; no waiver, retune, mint, archive, cleanup or push.

## 2026-10-06 — RP-264 committed-HEAD first filter

Review by: Codex (implementer self-first-filter).
Recorded by: Codex.
Reviewed range:db8398a3..97d916eb, all four commits/all10 changed paths:
three predeclarations, two test files, canonical docs, shared ledger and
current-state/queue/board/per-RFC plan/log reconciliation. Verdict: bounded
instrument repair passes first filter, NOT a designated independent verdict.
No implementation path or old report changed; kernel161 unchanged. Admission
precedes eligibility, saving subtraction follows nonnegative-clock admission,
and counters reset per call. No plan checkbox was flipped or bound weakened.

Committed-HEAD cold3024e2..b0d7cd (session22294,terminal exit0,0.207s) passes28
diagnostic profiles and the legacy gate. Earlier nine compiling probes, exact
probe-phase restoration and later stdout-only hash supplement are separately
recorded; they are not a second committed-HEAD mutation/full-study claim.
274734 rechecks final test hashes and all three retained reports unchanged.
Full source census remains93/3 with six real failures; H4/H5 drift remains RED.
Whole db8398a3..97d916eb requires Claude, independently of all prior ranges.
No mint/archive/push or full1.0 promotion. Goal active; next RP-269 diagnosis.

## 2026-10-06 — RP-269 threshold-source/Exit admission diagnosis predeclared

Resume clean9f5b81a0, no live handles. Accepted R10 H1/H2 and existing
fail-loud measurement rule authorize instrument admission, not a new balance
decision. Source f9863f/a6d45b: the payout reader selects first matching kinds
without sequence/order/duplicate checks; the pinned H2 test calls the generic
calculator without admitting its report against the declared study population.
3f9649/9086c9 supplies the existing H3 full-source oracle and ratified scenario.
Do not require97 rows from every generic calculation or assume new seed values.

Predeclared test-only diagnosis, before changing either implementation:

1. Exercise PaidReputationAtFirstElectiveExit using a clearly synthetic ordered
   pair: run1 scripted_first lifetime1e6, run2 collapse lifetime1e8, candidate
   threshold1e6. Compute the hand expectation from published cube-root level
   and policy modifiers. Legal controls additionally use zero lifetimes and
   the retained complete report (historical, not a fresh producer). Corrupt
   sequence/order, duplicate scripted/collapse samples, an extra/unknown sample,
   missing either sample and malformed lifetime values. The first-hour report
   stops at exactly these two Exits; this is not a general multi-run career API.
   Record typed refusal vs raw error vs panic, and the value from any admission.
2. Exercise the exact generic calculation used by the pinned H2 caller on
   independently deep-cloned retained inputs. Corrupt report/run source
   coordinates, aggregate source/count, missing/extra/duplicate runs, unknown
   seed/persona, noncompleted/invariant-failed runs, missing ending/transitions
   or required clocks, and empty report. No updates/regeneration. The admission
   failures describe a missing complete-study boundary, NOT a requirement to
   narrow the generic calculator to a hardcoded97-row population. Legal generic
   subset and reordered-population controls remain calculable.
3. Cross-check every study-source corruption against the already executed H3
   population oracle; disclose any gap it does not catch rather than claiming
   it validates more than its assertions. Any corrective scope follows measured
   outcomes in a separate record before code changes. Bind admitted source to
   the actual pinned caller, not a diagnostic-only unused helper.

Run focused cold root Make tests (-count=1); all handles terminal before any
edit. Diagnosis alters only a new test file plus records. No source/catalog/
kernel/formula/balance/schema/migration/CI/owner-copy/report/corpus changes.
No H2 pass or fresh H1 report inferred from historical arithmetic. RP-263
lineage, RP-268 artifact authority, fired H3/H4/H5, actual SQL/capacity and
prior reviews remain. Whole new span after9f5b81a0 needs Claude. No box, mint,
archive, push, cleanup or release claim; proper1.0 goal stays active.

Predeclaration correction before any diagnosis:7b03fa reads the actual prestige
math, which uses floor cube-root, not square-root. The original wording was
my recall error and is reconciled above, not an owner formula change. At the
declared synthetic1e6/1e8 pair and threshold1e6, levels1/4 yield elective
floor((4-1)*0.75)=2. Policy bytes d0fab4 give scripted1.0/collapse0.75.

## 2026-10-06 — RP-269 baseline and bounded correction scope

All handles terminal before this record. 3f0118..ad0dcb,session97278,
terminal exit2,0.365s:42 named children. Synthetic known-paid2 and zero
controls pass. Eight of16 corrupt Exit profiles still pay2: reversed, wrong
scripted0/2 sequence, wrong elective1/3 sequence, duplicate each kind and
unknown extra. Eight missing/kind/lifetime cases already refuse: four typed
ErrReputationMeasurement, two raw canonical-decimal errors and two raw prestige
arithmetic errors; no panic. No runtime/product outcome was measured.

Study-side generic invocation admits19 of21 corrupted sources; only failed
and empty reject. Existing full-source H3 oracle rejects20 but admits aggregate
source corruption, now RP-270. Three legal controls pass: retained full input,
generic two-persona/two-row subset and reordered full input. Historical report
arithmetic is not fresh H1 production or threshold ratification.

Bounded correction, declared before source edits:

- Payout reader admits exactly the ordered first-hour pair: index0/run1/
  scripted_first, index1/run2/collapse. No duplicates/extras/searching for a
  later collapse. Preserve canonical parsing, nonnegative arithmetic, zero
  lifetime legal control, formula/modifier/threshold and returned payout.
- A test-side study wrapper admits source through the existing full-population
  oracle, additionally matching aggregate schema/scenario/hash/constants
  coordinates to the declared suite. Bind the actual pinned measurement caller
  and source-corruption tests to it. The generic calculator retains arbitrary
  legal populations; subset control continues to use it directly. Do not alter
  the shared H3 oracle in this range: RP-270 stays a separate instrument route.
- Existing retained report and30-point calculation must remain byte-identical.
  No report regeneration/source refresh flags. No claim aggregate value arrays
  are re-derived or the historical producer freshly ran. No new public schema.
- Cold focused diagnostics/legacy threshold/H1/H3 tests, fast harness/core/vet.
  Independently compiling Exit cardinality/sequence/kind, full-source and
  aggregate-coordinate omissions must fail and restore exactly; if a guard
  survives, disclose and repair its fixture before claiming discrimination.
  No source or record edit while any handle lives.

Only reputation_threshold.go and test-side measurement/admission files plus
canonical docs/shared tracking. Harness-only, kernel161 unchanged; no actual
product source, balance/math/migration/schema/CI/owner copy/corpus/report edits.
H3/H4/H5 acceptance remains RED/open; no need to relabel a costly career repeat
as required evidence for this H1/H2-only boundary. Fresh H1/H2/lineage, RP-268,
actual SQL/capacity and prior reviews remain. Complete new span after9f5b81a0
needs Claude. No checkbox, mint, archive, push or full1.0 promotion.

## 2026-10-06 — RP-269 admission corrected; historical arithmetic preserved

All handles terminal before this record. New test-side study wrapper binds the
actual pinned measurement caller to the existing full-source population oracle
and aggregate schema/scenario/hash/constants admission. Generic calculator
population contract stays unchanged; a legal two-row subset remains calculable.
Payout helper requires exactly ordered run1/scripted_first then run2/collapse;
canonical parsing and existing prestige arithmetic remain. No report refresh.

Focused9a2e23..513a3b,session80222,terminal exit0,2.063s:45 diagnostic children
(18 Exit,27 study), legacy threshold/payout/H1/H3 tests and unrelated matched
integer-drift test pass. Actual one-Seed Chaos H1 takes1.68s, not a full97-run
fresh H1 report. The retained30-point measurement reproduces byte-identically.
Three additional aggregate schema/id/hash corruptions refine the declared
aggregate-coordinate cases, after the original21-source baseline. All four
are still admitted by the separate H3 oracle; RP-270 records that honestly.

Six independent compiling probes, cold and terminal before restoration:

| Omission | Output range | Demonstrated failure |
|---|---|---|
| exact two-Exit count weakened to at least two |c443ea..40e38c| duplicate each kind and unknown extra |
| ordered run sequences |59fa1f..12e746| four wrong sequence profiles |
| required Exit kinds |449a2e..389650| two unknown-kind profiles |
| study honors population oracle |1d8503..618651| eighteen malformed source profiles |
| aggregate coordinates |76306f..cc0874| all four aggregate corruptions |
| prior Founder level ignored |8b9d53..3d04c2| hand paid2 control and pinned report fail |

All exit2, no compiler failure or panic. The extra arithmetic probe confirms
the preserved formula/level dependency, not authority to change it. All three
source hashes restore exactly before each next probe: threshold.go
020593de605da8d7bdb372633b12d17c00fc78e2917111939e7fa7d2fde4f3be;
threshold_test.go633b676a8becf7363c53d47b3787fcbe16a0a90e99e3d76c915f51297ae20bd0;
admission_test.go272cf54f050e73c0258b5882df99a70d1a8d92aba962716966ad3a688afd4671.
Normal restored focused855e6b..14b798,session47498,terminal exit0,1.739s.
07e525/d17362 independently confirm exact restoration; report SHA46a8fea6
and retained H1 SHA1f7c774d unchanged.

Cold fast590f6d..039c63..3e1c11,session36474,terminal exit0:
harness57.543s, role0.159s, Commons0.374s and guard. Cold core732b24..37e269,
session65975,terminal exit0 includes full vet/cold core, production36.517s,
transport13.240s,save0.281s, numeric vectors/source version and unchanged
generated API/formulas. The redundant Pitch alias is cached; its full package
already ran cold. Narrow vet03845f exits0. No fresh SQL/client/browser/hosted
check, full CI pass or complete-career repeat claimed; RP-131 remains historical
RED and prior97/970 career study remains six-tie/report-drift RED.

The first corrective apply_patch failed atomic context verification (cabcbc/
5e7d7f confirm no partial source changes); exact current context was reread and
the patch reapplied before any test. No source/record edit during live handles.
Harness-only; afe1d9 confirms no kernel-watched prefix moved, kernel161 remains.
No live formula/balance/schema/migration/CI policy/owner copy/report/corpus change.
Whole span after9f5b81a0, both predeclarations and implementation/records, needs
Claude independently of all prior spans. No box, waiver, retune, mint, archive,
cleanup or push. Next RP-270 shared H3 aggregate admission, then fresh H1/H2/
lineage and RP-268 artifact authority. H3/H4 failures, H5 gaps, SQL/capacity and
the proper full1.0 goal remain. Goal active; this turn made concrete progress.

## 2026-10-06 — RP-269 committed-HEAD first filter

Review by: Codex (implementer self-first-filter).
Recorded by: Codex.
Reviewed range:9f5b81a0..677deb07, all three commits/all11 changed paths:
both predeclarations, three harness source/test files, canonical docs and
ledger/current-state/queue/board/plan/log reconciliation. Verdict: bounded
first filter passes, NOT designated cross-party approval or H2/AC13 closeout.

Committed-HEAD coldeaea38..e33bcf,session75202,terminal exit0,1.980s passes45
diagnostic children and legacy threshold/H1/H3 checks. Source inspection shows
the actual pinned caller consumes the guarded wrapper, not a diagnostic-only
unused helper. No separate caller-bypass mutation was executed or claimed.
Generic subset control stays on MeasureReputationThresholds; complete/reordered
controls and all source corruptions use the same wrapper as the pinned caller.
All six earlier compiling probes restored exact source hashes; retained H1/
threshold reports remain unchanged. No source/record edit while a handle lived.

H3 helper reuse proves only its actual assertions. The four aggregate-coordinate
gaps remain registered RP-270 on the separate H3 path. Stored-historical source
admission is not fresh production, semantic authenticity of every aggregate
value, full report lineage or owner balance adoption. Kernel161/live product/
balance/schema/CI unchanged; no checkbox or epoch mint. Whole9f5b81a0..677deb07
needs Claude independently of RP-264 db8398a3..97d916eb and prior ranges. Goal
active; next accepted RP-270, then fresh H1/H2/lineage and RP-268 grounding.

## 2026-10-06 — RP-270 shared H3 aggregate admission predeclared

Previous work made concrete progress:97d916eb repairs H4 evidence admission
and677deb07 repairs first-hour Exit/H2 study admission, with cold checks and
negative controls. Resume clean0a486d79; no live handles. Accepted R10 H3 and
fail-loud instrument rules authorize admission, not changing the fired criterion.

RP-269 directly demonstrates four aggregate schema/id/hash/constants corruptions
accepted by the existing shared H3 population oracle while H2's local guard
refuses them. Add those same four fields as independent mutations to H3's own
synthetic population test, execute the unchanged oracle cold, and record the
four baseline failures rather than inferring them from the H2 path alone.

Then, only if reproduced, add aggregate coordinate admission to the shared H3
helper. Existing full-source/count/seed/outcome/clock/Exit checks stay unchanged;
do not rederive aggregate value arrays or require altered milestone distributions.
The H2 local guard remains an independent defense, not proof its shared H3
dependency is correct. One compiling omission of the new combined coordinate
guard must fire these four H3 mutations, exact restoration required.

Cold focused H3/H2/H4 admission tests and fast/core/vet; execute all five current
producer arms/all97 seeds (485 runs/3395 clocks) with the existing explicit
observation selector. Same factor1.000001 and strong2 control, same ratified
scenario/policies/horizon. No selector shortcuts, thresholds, waivers, update
flags or old report changes. A still-failing tiny criterion stays RED; neither
changed lifetime nor the strong control substitutes for it. No edits while any
handle lives; poll to terminal before records/probes/restoration.

Scope one existing harness test file plus canonical docs/shared tracking.
No product/math/balance/kernel/schema/migration/CI/owner-copy/report/corpus edits;
kernel161 stays. Whole new span after0a486d79 needs Claude. Previous ranges,
fresh H1/H2/lineage, RP-268, H3 author reconciliation, H4/H5, actual SQL/capacity
and full1.0 remain. No checkbox, mint, archive, cleanup or push; goal active.

## 2026-10-06 — RP-270 baseline fixture refinement, before guard changes

All handles terminal.4a8098..6e35c8,session16198,exit2,0.346s prints four new
aggregate refusals lost, while the existing27 controls pass. Source inspection
then notices the synthetic positive fixture set only Aggregate.RunCount; its
schema/id/hash/constants were zero/empty. In particular the schema mutation was
0→1, not a clean valid→invalid single-field mutation. This first attempt is not
claimed as the four isolated clean corruptions or evidence a new guard passes.

Before touching the oracle, refine only the synthetic fixture: initialize its
aggregate schema1/scenario id/hash/constants from the same loaded suite as its
report, plus actual generated RunCount. No aggregate values/milestone/population
change, report source regeneration or new authority. Then rerun the unchanged
oracle: legal fixture must still pass and all four isolated coordinates must
fail refusal. Only after that terminal evidence add the predeclared guard and
execute its omission/full study. Kernel161 and all live/report bytes unchanged.

Clean diagnostic74bcfc..954115,session16589,terminal exit2,0.353s confirms
all four isolated coordinate refusals lost from the now-valid synthetic source;
existing27 controls pass. Add only the predeclared shared aggregate-coordinate
guard. This is source admission, not aggregate-value recomputation or a change
to the exact tiny sensitivity criterion. H2 retains its independent guard.

## 2026-10-06 — RP-270 shared admission repaired; full study criterion still RED

All handles terminal before this record. Four new H3 coordinate cases plus
properly initialized positive aggregate fixture and shared coordinate guard
are the only source edits. Producer/math/selection/horizon/copy/CI unchanged.
Focused380b18..168c24,session48036,terminal exit0,0.436s passes31 H3,45 H2,
28 H4 diagnostic children and the three direct legacy gates. Guard omission
f507bb..e75fa9 compiles and exits2,0.576s; exactly four H3 corruptions lose
refusal, all27 prior controls pass. Source1c6c17e45356fc0d52368ce71a17924b0ae5fa3d56cf726b4a465ca44e1f3ade
restores exactly (533367); normal1cb2a0 exits0,0.217s.

After the full study, repeat that same omission with H2 controls included:
d11d0a..accfc1 exits2 with exactly four H3 failures while all45 H2 diagnostic
children and both legacy threshold tests stay green. This is the same probe
repeated for defense-in-depth, not two independent omissions. H2's independent
aggregate guard survives the missing shared H3 guard; it does not establish
that shared H3 admission is present. f9b208 restores identical source hash;
normal42ccd2 exits0,0.299s. No compiler failures/panics or hidden source drift.

Cold fast0204ec..2a8c53..2ba373,session55230,terminal exit0:harness67.552s,
role0.175s,Commons0.453s and guard. Core76b91b..17ad03,session52391,terminal
exit0:vet/full cold core, production43.989s,transport13.222s,save0.287s,
numeric vectors and unchanged generated API/formulas. Redundant Pitch alias
cached after its full package ran cold. Narrow veta25dd2 exits0.

Full current-producer study56adee..96dfcc,session19726,terminal exit2,
313.732s (test313.61s), not cancelled. All five arms complete and pass strengthened
source/aggregate admission: each97 runs/679 clocks, total485/3395. Retained-
epoch8/paired-none/unit neutrality passes; tiny changes0 clocks, strong289.
Only exact tiny-factor sensitivity criterion fires. No larger-factor substitute.
Independent full-log parser confirms each arm97/679 and epoch8/unit/tiny0
clock changes vs none, strong289. Raw outputs retained during the turn; logged
chunk coordinates support replay, not a newly generated JSON lineage artifact.

Source headers unchanged: base epoch8baa89050, paired fixture3625eddb,
scenario18a6f16a,policye5e5de70,threshold1e12. Actual Reference none/unit
lifetime1.4605083614e6/3.54431965065e6; tiny1.46050991148e6/3.54432330408e6;
strong3.5053281185e6/7.66194816889e6. All paid0/0. Tiny clocks remain
[0,10000,66992,356000,900000,346000,2700000], strong
[0,10000,44072,182000,900000,177000,2700000]. H3 author reconciliation remains
required; lifetime/transition changes are not a substitute milestone.

3bc738 verifies source hash and retained H1/H4/H5/H2 report SHA values unchanged.
No fresh SQL/client/browser/hosted or full CI claim; RP-131 remains historical
RED, prior H4 six ties/H5 drift remain. Kernel161/live product/math/balance/
schema/migration/CI/owner-copy/corpus unchanged. No edits while any handle lived.
Whole span after0a486d79, both predeclarations and implementation/records, needs
Claude independently of all previous spans. No checkbox/mint/archive/push.

Read-only authority sweep during live measurement (no edits):3d47c7/bc5bef/
d81848/482566 shows career purchases occur only at run2 collapse, not applyEnding's
scripted-first credit/reset. R10 H4 says at each Exit, but85cb09 explicitly rules
command-replaced scripted_first never carries a plan (R6);482208 also preserves
next-run-only effects. Record RP-271 immediately after handles terminate as
normative coverage ambiguity, NOT a confirmed runtime defect. Do not invent a
scripted plan or freeze the current assumption into RP-268's absent career data.
The RFC author must reconcile plan-bearing-vs-all-Exit intent. No owner ruling
was inferred and no message sent to Claude. Other accepted work remains: predeclare
fresh H1/H2/report lineage under RP-263 with old v1 bytes preserved. Goal active;
this turn made three bounded repairs and fully measured their remaining failures.

## 2026-10-06 — RP-270 committed-HEAD first filter

Review by: Codex (implementer self-first-filter).
Recorded by: Codex.
Reviewed range:0a486d79..bc315ecf, all three commits/all9 changed paths: both
predeclarations, the one test-side helper/fixture file, canonical docs and all
ledger/current-state/queue/board/plan/log updates. Verdict: bounded first filter
passes, NOT designated cross-party approval, H3/AC13 acceptance or archival.
The criterion remains RED. No kernel/runtime/math/balance/old-report byte moved.

Committed-HEAD cold3e628b..a14ab5 (session40660,terminal exit0,0.306s) covers31 H3,
45 H2 and28 H4 diagnostic children plus direct legacy gates. Earlier compiling
guard omission and its H2 defense-in-depth repeat each restore the same exact
1c6c17e4 source hash. The completed485-run study was executed on that identical
source before docs/records; it is not relabelled as a new committed-HEAD rerun.
The initially invalid synthetic fixture is disclosed and refined before the
clean baseline; no green guard proof derives from the invalid first attempt.

No aggregate value-array recomputation/producer authenticity claim. RP-271 is
an unexecuted normative ambiguity, not permission for a scripted-first plan.
Whole new designated-review handoff must include this self-first-filter record
as well as0a486d79..bc315ecf; the designated verdict must cite its actual full
reviewed HEAD/range, not discard uncovered edge records. Prior ranges/records
remain independent, and no self-review substitutes for Claude. Next safe work
fresh H1/H2/report lineage under RP-263. Goal active; no box/mint/archive/push.

## 2026-10-06 — RP-263 fresh H1/H2 and lineage predeclaration

Authority: accepted Reputation R10 H1/H2, existing ratified first-hour scenario/
policy and OD-2 measurement-only threshold study. Baseline c352370a, clean main.
This is research instrumentation/evidence, NOT balance adoption, H3/H4/H5 closure,
minting, archival or a new product contract. Preserve all historical v1 reports.

Population: all97 ratified seeds (64 Chaos,32 Casual,1 Reference), unchanged
7200000ms horizon; experiment (minimum200, factor2e0, route50, seed1e4,
towers10), workers8. Load the unmodified epoch-8 suite; derive H2 using its
exact Prestige catalog, all30 existing 1/2/5e3..12 candidates, p50 envelope3..10
for Casual and Chaos. The live1e12 row must exist, fail the envelope and pay
zero for every persona. Candidates remain proposals, even if any satisfy.

Allowed new paths: server/harness/reputation_current_measurement_test.go and
planning/reputation-tree-v1/{first-hour-reputation,threshold-measurement,
measurement-lineage}.2026-10-06.v1.json; canonical docs/tracking only otherwise.
No production, balance, kernel, scenario, policy, old-report, CI or copy edits.
The three output paths are absent before this declaration. Existing generic
Harness Observability is for live objective instrumentation, not immutable H1/
H2 provenance; keep this small companion private/test-side, no shared schema.

Closed opt-in test selector: off (explicit fresh-run skip), record or verify;
unknown values fail. Commit the instrument before record so its producer commit
and complete server/balance Git tree identities name actual committed code/data.
Record requires clean tracked server/balance inputs and no untracked inputs in
those trees. Record runs the real full producer before H2, validates complete
population and recomputed aggregate including failure arrays, then exclusively
creates all three declared outputs (never overwrite). A partial write is invalid,
not completion. The companion binds raw H1/H2 SHA-256, producer commit/tree ids,
kernel version, runtime identity and census. Metadata alone is not freshness.

Static verification resolves recorded Git objects and rederives aggregate/H2
from the admitted report; clearly labelled historical-artifact validation, not
a fresh run. Full verify requires the same clean source/data trees, reruns all97
and byte-compares both outputs. Different producer trees must fail, not silently
refresh provenance. No timing thresholds or limited populations substitute.

Negative controls: altered raw bytes/hashes, schema/count/source metadata,
changed aggregate values, missing/duplicate runs, changed H2 result/envelope/
grid and producer identity must refuse. Demonstrate compiling guard omissions
for report binding, aggregate/H2 recomputation and producer-tree admission;
restore exact source hashes only after the same handles terminate. Fast harness,
cold core/vet and focused cold committed-HEAD tests follow; full CI/SQL/browser
remain unclaimed. Generated artifacts are the explicitly authorized measurement
outputs while tests run; no manual source/tracking/probe edits while any handle
lives. Designated Claude full-span review remains separate. No box flip.

## 2026-10-06 — RP-263 instrument checkpoint before recording

New private test file only; no existing producer/math/catalog changed. The
instrument uses the exact suite Prestige, recomputes full aggregate and H2,
strictly decodes H1/lineage and binds raw report bytes. Record/verify check clean
inputs before/after, and source/data trees again after production. Verify also
validates retained lineage against its recorded Git objects before running.
The existing CI harness checkout fetch-depth0 supplies those objects; no CI
change. Full replay requires unchanged trees; artifact-only checks do not.

First command d9c8cc..1314ac selected no tests: Make consumed trailing dollar
and space, forming Admission-count=1. It is invalid evidence, explicitly not a
green gate. Corrected27cd79..bbd474 executes all24 refusal children plus healthy
historical-source/commit-advance controls, exit0,0.185s. No fresh producer yet.

Four independent compiling omissions, each terminal exit2 before restoration:
raw binding d4f5e8..d50f94 fails h1-hash/h2-hash/raw-h1 (raw-h2 remains defended
by exact H2 bytes); aggregate f8f652..be207d fails aggregate while explicit
failure arrays remain defended; H2 6c9612..089ad9 fails result/envelope/grid;
producer-tree 6ddee2..d25e2c fails changed server/balance/kernel. No compiler
errors/panics. Each exact restored SHA947a957736fd346d11ccdeab29c65d5fcffb790ec3acb12747eef5b81dbfc0bc.
Restored0e0e05..803f17 exit0,0.284s. Unknown selector396284 exits2 immediately;
quoted Go end-of-text anchor correctly selects the one fresh-measurement test.

This is an intermediate committed instrument, not completion: the new artifact
gate intentionally requires the still-absent three outputs. Next record once
from this committed code, then full byte-identical replay, focused/fast/core/vet,
canonical docs/board/ledger reconciliation. No old report overwrite, candidate
adoption, kernel bump, mint, box, archive, message to Claude or push. Full range
afterc352370a including declarations/code/reports/records needs designated review.

## 2026-10-06 — RP-263 fresh reports executed and reproduced

Producer instrument committedda512cd39b1e49e5fc734cd0e05be6a11a73fb2e under
c095e2b5 predeclaration. Server tree7db46f7ed63ab07402fc40ab66e7d710fb56d144,
balance treecd982b7c58a53cac0ab4c91a773705033930414d, kernel0.3.161;
runtimego1.27.1/darwin/arm64 (not hosted runner evidence). Record953c02..5cc430,
session74897, terminal exit0,69.118s (test68.91); full verifyd99f93..a6ee7b,
session77792, terminal exit0,62.416s (test62.30). Each executes all97 seeds and
679 clocks at the original horizon/tuple. Strong source/census, recomputed
aggregate including failures, pinned exact Prestige and all30 H2 rows pass;
full H1/H2 bytes match on independent replay. Metadata alone is not freshness.

New H1 SHA3ce87b08dcebebcc55f19bff73e93473c61fcb052825ec01179bfea9332d02ac;
H2 d5979c934b64d867834580a24e96d6b643336704507db50687b898b4edd773f4;
companion2377a2d4e0a862be93b9cc61dc69c9b99dc345960704e53bb63f5436dca88421.
Outputs are exactly the three absent paths declared, never old v1 replacements.
Actual second record289595..376dd4,session39877, exits2 before production at
the first existing path. Hashes remain identical. Partial writes would not
constitute a valid three-file measurement; no partial write occurred here.

Independent JSON comparison: only Reference's reputation_exits differ in H1;
its lifetime1.26604673417e6/3.34889972124e6 becomes1.4605083614e6/
3.54431965065e6. All other96 rows and the complete aggregate are semantically
identical; Reference's clocks/ending/transition fields are unchanged. H2 has
three real Reference changes:1e4 paid0→1,2e4/5e4 paid1→0. Not merely formatting
drift. Casual/Chaos statistics and satisfying proposals2e4/5e4/1e5/2e5 remain
unchanged. Live1e12 remains zero for every persona and fails the envelope.
No candidate is adopted, minted or substituted for the live balance contract.

Cold fast48d74f..94bb9d,session42381,terminal exit0:harness66.395s,
role0.170s,Commons0.365s, guard. Cold core4a9fb5..56ada1,session60403,
terminal exit0:full vet/non-harness core, production35.115s,transport13.260s,
save0.275s, numeric vectors and unchanged generated formulas/API. The extra
Pitch alias was cached after its whole package ran cold0.327s. No fresh SQL,
browser, hosted CI or full verify-push claim. Local harness CI uses full Git
history by existing workflow declaration; no workflow bytes moved.

Focused713145..a4eca2,session94340,terminal exit0,0.311s validates retained
artifacts and all24 refusal children; the fresh test explicitly SKIPS by default.
This is static artifact validation, never a third full run. All four compiling
omissions and the invalid first command remain disclosed above. Final source
SHA947a9577 unchanged. Old H1/H4/H5/H2 hashes match1f7c774d/648f36d6/4ed79054/
46a8fea6. One guessed career-report path was absent; actual career-h4.v1.json
was resolved with rg and checked, not silently counted as the failed read.

Canonical docs reconcile stale live H1/H2 pending statements, not just append;
board/queue/ledger now name this bounded evidence and H4/H5 provenance next.
Historical logs remain append-only. No source/tracking edits while any handle
lived (only the predeclared generated outputs). Kernel/live math/balance/copy/
schema/corpus/CI unchanged. Full new span afterc352370a includes predeclaration,
instrument, all outputs and all records and needs Claude's designated pass.
Prior review ranges stay independent. H3 tiny criterion, H4 six ties, H5
epsilon/run4, RP-268/RP-271 author boundaries, SQL/capacity and whole1.0 remain.
Goal active; no checkbox, criterion waiver, retune, mint, archive, cleanup,
publication, message to Claude or push.

## 2026-10-06 — RP-263 fresh-measurement committed-range first filter

Review by: Codex (implementer self-first-filter).
Recorded by: Codex.
Reviewed range: c352370a..422892c0, all three commits/all12 changed paths:
predeclaration, sole new private test file, all three dated outputs, canonical
docs and all ledger/queue/plan/board/checkpoint records. Verdict: bounded local
first filter passes, NOT designated cross-party approval, AC13 or archival.

Committed-HEAD coldce6e9f..3c7139,session67347,terminal exit0,0.430s runs retained
artifact validation and all24 refusals; fresh test explicitly skips by default.
Full record/replay were run before the evidence-record commit on identical
server tree7db46f7e/balance treecd982b7c; no claim of a third full replay at the
new HEAD. Current trees match the recorded producer. Scope contains no live
math/balance/kernel/schema/CI/copy/corpus or old-report edit. No checkboxes flipped.
Invalid initial command, compiling failures and real overwrite refusal are
preserved on record, not absorbed into green evidence.

The designated handoff must include this record edge along with the entire
c352370a..422892c0 span, cite its actual complete reviewed HEAD, and independently
cover prior ranges. This self-filter cannot authorize archive/mint/adoption.
Next accepted work H4/H5 report-envelope provenance audit; author boundaries
and actual H3/H4/H5 failures remain. Goal active; no push or message sent.

## 2026-10-06 — RP-263 career/report source-binding predeclaration

Previous turn is progress: dated H1/H2 artifact/replay and local gates completed.
Baseline dfc2fb8c, clean main. Accepted R10 H4/H5 measurement authority and
RP-263's open report-envelope source route, NOT a career-policy/epsilon ruling.
Read-only audit shows RunReputationCareer overrides Prestige threshold after
loading the complete paired bundle. Its catalog RunKey omits this effective
policy, experiment tuple, purchase policy, exclusion and effective horizon;
H4/H5 projections also discard these inputs. Reproduce that identity collision.

Allowed code: server/harness/reputation_career.go (observation metadata only),
existing career/relevance test report producers and one private diagnostic test;
canonical docs/tracking otherwise. Add actual career measurement_source:
complete paired RunKey, first-hour policy hash, SHA-256 of canonical serialized
effective loaded Prestige policy, experiment tuple, horizon, purchase policy,
excluded node. It is a measurement identity, NOT an epoch/hash rewrite or live
game contract. Bind both report callers to expected source before projection;
retain both H4 arm sources and all H5 arm sources in the in-memory report.
No historical file writes/update env flags; the old report comparison remains
RED after drift rather than relabelling a new shape as historical completion.

Before producer edits, add a JSON-side diagnostic that compiles against the
existing result and runs actual Chaos seed0 under base/none/live-threshold/
excluded-node configurations. Record missing source and same catalog keys,
actual earned purchases/observations. Then add source metadata/checks without
changing policy execution, balances, choices, clocks, lifetimes or math.
Healthy no-exclusion and actual excluded-node sources must bind separately;
malformed fields must reject through both caller paths. Demonstrate compiling
policy-hash/tuple/mask omissions and caller-admission severing, restoring exact
hashes only after terminal handles. Cold focused/fast/core/vet gates follow.
Full H4/H5 observation, if run, preserves failed criteria and stale-file RED;
no new reports or asserted full-population reproduction without an executed run.

This cannot resolve RP-268's missing declared career data or RP-271's each-Exit
wording; it observes today's instrument, not adoption of that policy. H3 tiny
criterion, H4 six ties, H5 epsilon/run4, source-code companion provenance and
designated reviews remain open. No kernel bump, product/schema/copy/balance/CI/
corpus change, box, mint, archive, cleanup, deployment, message or push. No manual
source/tracking edits while any handle lives. Full span afterdfc2fb8c requires
Claude's separate designated review; all prior ranges stay independent.

## 2026-10-06 — RP-263 actual source-collision baseline

JSON-side diagnostic3c1486..87d87f,session99978,terminal exit2,8.648s:
all four actual Chaos seed0 careers complete, share the identical paired catalog
RunKey, and fail the missing-source assertion. Cheapest/none/live1e12/excluded-
cash differ in purchases/payout/gate. No synthetic or compiler-error baseline.
Refined logging e3b003..0b22ea,session67564,terminal exit2,7.910s reruns the same
four failures and captures whole semantic result SHA (excluding metadata),
not four additional distinct defects:
base fc02332820547f0bc437fa0c342e361673cf66be53363e9c3b97edacf6c5125c,
none d0ce1b7ee0a102dc1a5258d775e04d9d3e707835a3f7c39601b9735d0fdd9d1f,
live225832823eab3e83fc9eb2595660e113c14645a51605aa66b14d442ea72bcd65,
excluded8be070f54cb5ee12f1e32cafe6b83dca9d0856b7eaf6fc113f0fb6878b56cc7a.
Gates respectively320000/442000/442000/440000ms; cheapest earned5/5 and buys
p05/cash_small/generated_beige_tower/upgrade_continuous_feed_paper; none buys
nothing with the same earned payouts; live pays0/0 and buys nothing; exclusion
buys p05/p25. Effective Prestige policy hash at1e5 isab4edf83, at1e12caa0dd54.
These full-result fingerprints are the metadata-only non-regression controls.

Also register RP-272 immediately: read-only H5 loop increments purchased counts
but drops nil-clock pairs from its median without recording finite/unreached
denominators. Controlled reproduction and explicit censoring authority remain
separate; this range does not impute censored clocks or silently bless1ms.
No source/tracking edits while either baseline handle lived. Next the declared
observation/caller binding, negative controls/probes, cold gates and records.

## 2026-10-06 — RP-263 source/caller instrumentation checkpoint

Observed inputs now return with actual careers; catalog RunKey/epoch authority
unchanged. Effective policy hash names the loaded overridden policy, not base
catalog Prestige. H4 projections bind/retain both arm sources, H5 projections
bind each baseline/mask source and retain all sources in the report header.
These are private measurement report fields, not a live/save/wire schema change.
Four real careers preserve the baseline complete semantic fingerprints;36
synthetic corruption children cover both H4 arms and H5. Healthy projections
retain distinct baseline/mask sources. Normal1e831f..15f170 exit0,8.438s;
after header extraction efafd0..eceab4 exit0,8.388s (returned output truncated,
not a truncated measurement; the full preceding36-case run is recorded).

Nine independent compiling omissions, terminal exit2 before exact restoration:
effective policyc57b66..a20968 fails base/none/mask; live policy stays healthy;
experiment83b18e..84d68e fails all4; mask227644..5c5f28 fails excluded only.
All four semantic fingerprints stay unchanged even on those metadata failures.
H4 treated56ca66..63f46e, control2ade19, H5ecf825 each admit all12 malformed
profiles for that arm when its admission result is ignored. H4 retentionddbc15..
99e4ca, H5 outcome retentiondb1929 and H5 report retentionfbd3b8..c65a58 fail
their healthy known-source assertions, not compiler errors/panics. Exact
restored .go SHA3b691d40; consumer hashes5355742f/bc16ab13 and diagnostic806c69a1.
Final restored focused5b3dae..db335b exits0,7.796s, includes4 actual/36 source/
28 H4 profiles and legacy H4 gate. Source-binding completeness is not H4 timing
acceptance or full H5 relevance proof.

After all handles ended, add one observation-only census line to each full
report producer so a subsequent run reports its admitted arm count. No guard,
formula or metadata behavior changes; probe hashes above precede those two
logging lines. Next cold fast/core and declared full H4/H5 observation. Old
report update flags remain off; stale shape/timing and fired criteria must stay
RED. No new reports or full-population claim at this checkpoint. RP-272 source
finding remains unexecuted. All current and earlier designated-review spans
remain independent. Kernel161/live math/balance/copy/corpus/CI unchanged; no box,
mint, archive, owner ruling, cleanup, deployment, message or push. Goal active.

## 2026-10-06 — RP-263 full source admission and remaining report findings

Implementation committed831fa9b3 afterec71a33c/8f51c0a5. Exact restored sources
then plus two admitted logging lines: .go3b691d40, H4 test1d79b174,
H5 test0c0b4148, diagnostic806c69a1. No probe/checkpoint edits during live runs.

Complete97-pair H4/970-arm H5 run8ff774..6a52a7,session70877,terminal exit2,
636.768s, not cancelled or restarted. H4 logs194 admitted sources,97 rows,
93 timing comparisons/3 valid exclusions, Casual4/14/21. Its six ties remain
at Casual1/6/8/11/24/25 and275000/425000/405000/265000/360000/430000ms.
ce4476 records those and the preserved report comparison RED after129.55s.
H5 completes506.93s,970 admitted arms/970 retained sources; only its old-report
byte comparison fires, no source or node-classification error emitted. This is
not H5 acceptance: unruled epsilon/run4, denominators, policy/data authority,
software-provenance artifacts and actual SQL still remain. No v1 file updated.

Cold fastc4a989..cf7daa,session34615,terminal exit0:harness77.901s,
role0.161s,Commons0.408s and guard. Cold core4729d7..702695,session90827,
terminal exit0:full vet/cold non-harness tests, production41.781s,
transport13.273s,save0.282s,numeric vectors and unchanged API/formula generation.
Pitch alias cached only after its whole package ran cold0.327s. This is local
gate evidence, not hosted CI/client/browser/SQL/full verify-push evidence.

Full H1 verifydf6d21..a8ecf4,session76735,terminal exit2,0.280s intentionally
refuses changed server tree61f1370e before production. Its dated producer
da512cd3/tree7db46f7e remains immutable historical evidence. Static artifact
validation stayed green in cold fast; no relabelling or overwrite. Old H1/H4/
H5/H2 SHA1f7c774d/648f36d6/4ed79054/46a8fea6 and dated H1/H2 SHA3ce87b08/
d5979c93 remain unchanged.

Read-only resource checks: docker system df8658e2..233c6a,session35296,
terminal exit0,342 images/78.3GB,143 volumes/39.36GB, build cache2.634GB.
Resolve game-ui postgres9402b6ad7eb7 and verify compose projectcloud-clicker/
servicegame-ui-postgres. Its data tmpfs has8108540KiB free, but actual root/tmp
overlay still100%/39784KiB free. Do NOT confuse tmpfs with usable build capacity.
No fresh Docker tests started. Narrow cleanup approval requested asynchronously,
not granted, preserving all volumes/running containers/release/rollback artifacts.
Tag/source-label filters identify no Cloud Clicker image candidate; do not
delete another project's images or infer authority from reclaimable sizes.
One compounded listing invocation is sandbox-denied; standalone retry succeeds.
No secret env read, cleanup, new container or deployment. Capacity hold remains.

Read-only report audit while the full run lived (no edits): historical H4 nil
baseline Casual4/14/21 owncash_small/generated_beige_tower. Historical H5
publishes Casual bought32/23 and medians70000/15000ms without finite/unreached
counts. This is not a current mask-arm census or shared-software provenance proof
(RP-272). Current H4 again observes those nil baselines; proper census research
still precedes any censoring or median policy.

Register RP-273/274 immediately after handles terminate: source shows H4 appends
savings only for strictly-faster finite pairs; independent historical arithmetic
has28 finite eligible Casual pairs/six ties, all-finite min/p50/max0/80000/350000,
versus22 faster pairs/published15000/80000/350000ms. No current full-summary
value claim beyond observed ties. R10 also requires per-node min/p50/max, whereas
the schema/writer exposes only per-policy aggregates plus compound starter rows;
no reconciled per-node attribution found. Need controlled denominator diagnosis
and author clarification before calling package savings isolated node effects.
No new statistic, epsilon, horizon, mechanic or owner ruling was adopted.

Docs/ledger/board/queue/plan now reflect observed source completion and next
RP-272/273 diagnosis, with RP-274/RP-268/271/H3/H5 author routes held. Full new
span afterdfc2fb8c, declaration/baseline/code/docs/records, requires Claude's
designated review independently of every prior range. Kernel161/live math/
balance/CI/corpus/retained reports/owner copy unchanged. No checkbox, acceptance,
mint, archive, publication, cleanup, message to Claude or push. Goal active.

## 2026-10-06 — Career source binding self-first-filter

Review by: Codex (implementer; self-first-filter, not designated).
Recorded by: Codex.
Reviewed range: `dfc2fb8c..850ba69e` — all four commits, all 12 paths,
including declaration, reproducer, implementation and evidence reconciliation.
Verdict: ready for designated review, with H3/H4/H5 acceptance still red/open.

Inspected the full source diff: metadata is calculated after effective policy
loading; both consumers admit the declared source before projecting results;
H4 retains both sources and H5 retains every arm in declared order. Gameplay
and the existing gate/classifier branches are unchanged. Four independently
pinned complete gameplay results, 36 malformed-input refusals and nine
compiling omissions are recorded above; synthetic projection controls are not
represented as earned careers. Historical artifacts and author boundaries
remain explicitly separate. Reviewed all tracking changes against those limits.

Committed-HEAD cold focused check at850ba69e: root `make test-go`, harness,
source/consumer/gate/static-artifact selectors, `-count=1`; chunks9acc50/510099,
session38003, terminal exit0,7.972s. No live test handle remained before this
record edit. Full exhaustive/fast/core evidence above used the identical
committed831fa9b3 source; subsequent changes were tracking only, not a new full
run. This record edge must also be included in Claude's designated range; it
does not self-approve archival or cover any earlier independent review span.

## 2026-10-06 — RP-272/273 controlled population observation predeclaration

Authority: accepted R10 H4/H5 reporting and evidence discipline. Question:
which gate pairs actually contribute to the existing reported statistics, and
which purchased/eligible careers are omitted because one or both clocks are
unreached? This is observation, not a censoring-policy decision.

Scope: existing H4/H5 test report producers, one new test-side population helper
and diagnostic file, canonical docs and tracking. Keep gameplay/threshold/
horizon/policies/masks/epsilon/classifier/strict-sooner gate unchanged. Preserve
historical JSON and update flags off; do not mint a new report or epoch. Keep
the existing H4 faster-only statistic, explicitly describe its population, and
add a separately named all-finite descriptive statistic including ties/slower
pairs. Neither statistic assigns a value to an unreached clock or attributes a
package effect to an isolated node (RP-274 remains an author route).

Predeclared populations: H4 every admitted row, partitioned into starter/no-
starter rows and, for starter rows, finite pairs (faster/tied/slower), treated-
only, control-only and both-unreached. H5 every persona/node baseline career,
partitioned bought/not-bought, and bought pairs partitioned into the same clock
states. Publish counts even for zero bought or zero finite pairs. Finite H5
median remains conditional on bought AND both clocks finite; name that fact.

Controls: deterministic synthetic pairs spanning all clock states, including
one finite positive effect amidst three censored bought pairs and one not-
bought career; H4 includes positive, zero and negative savings plus a non-
starter. Assert exact population counts, conservation and finite statistics,
and that existing gate/classifier outcomes remain unchanged. Also retain the
current source-admission and malformed-H4 controls. These are deliberately
synthetic statistical controls, not real career prevalence measurements.

Before each separate compiling omission, pin/restore exact files and await
the same test handle to terminal. Demonstrate that dropping finite, tie,
unreached, bought/not-bought or all-finite observations makes controls fail;
restored tests pass. No tracking/probe edits during any live test handle.
Then cold focused/fast tests and the complete existing97-pair/970-arm lane,
logging current counts from the real producer. Finish one full run; no restart
or horizon/threshold adjustment. Expected old-report comparisons remain RED
and H4's strict criterion remains RED unless the unchanged producer contradicts
that expectation. A RED criterion is a result, not permission to loosen it.

Exit: accountable current denominators and exercised observations, with exact
source/commands/limits recorded. No H5 epsilon/run4/censoring adoption, H4 gate
waiver, per-node causal conclusion, fresh artifact provenance, acceptance box,
SQL/browser claim, CI edit, archive or push. All work aftera3e36f94 needs its
own full designated range, independently of the preceding source-binding work.

## 2026-10-06 — RP-272/273 controlled census observation implementation

Under69c79d28, only test-side report producers/helpers change. H4 retains its
original strict gate/faster-only savings and adds all-finite descriptive package
savings plus a complete row/clock-state partition. H5 retains bought counts,
finite-only deltas and classifier, adding all-career/bought/clock-state counts.
The JSON field names and notes state the conditional populations. No imputation,
per-node causal attribution, policy/horizon/epsilon or old-report change.

Synthetic H4 seven-row control covers positive/zero/negative savings, all three
unreached states and no starter; the strict gate still emits3 violations.
H5 includes one positive finite pair among4 bought pairs, one not bought,
an all-not-bought persona, and a separate finite faster/tied/slower population.
Exact counts/savings/deltas, conservation and both serialized populations pass.
Also reject inconsistent H4 clocks/savings. Initial focused runb19675..5bacc3,
session67094,terminal exit0,0.403s; refined report-retention/finite controlse519b2..
9bf899,session95046,terminal exit0,0.341s. Neither is a real career prevalence.

Eleven separate compiling omissions all exit2 semantically: pair-total60fd85,
finite9c6428, tief23b58, both-unreached22df89, treated-onlyc64a55,
control-onlyb506f6, not-bought66f868, bought9d8c1c, all-finitec2841d,
H4 JSON98c5ef and H5 JSON290740. Each same handle terminates before restoration.
After each restore, exact source SHA verifies: population helper
e756486f61e1889a1575923ae3cf68a26413c5affe7e5243583ee195024cbe80,
H4d3c898faef782e5355676c944950ba1739f835aac048d68e38929673be4b37bf,
H502b4f12264b658f6dd3d62224ce452acc1b0a449a41ba4a5f495e2295e3daa59.
Restored focused source/gate/classifier/population check769339,terminal exit0,
0.118s. Full cold fast and97-pair/970-arm current census are next, not yet
claimed. Kernel161/live math/balance/CI/reports/corpus unchanged; no box,
acceptance, mint, author adoption, archive, cleanup, SQL/browser claim or push.

## 2026-10-06 — RP-272/273 complete current population evidence

Committed producerb22ce51e, server tree71f436514376317b7b5675dc15105f54b10c56d4,
balance treecd982b7c58a53cac0ab4c91a773705033930414d, kernel0.3.161 unchanged.
Source hashes exactly match the three pinned/restored files above. All test
handles are terminal before this tracking edit; the working tree stayed clean
throughout both checks. No update flags or source changes during measurement.

Root `make reputation-harness-check`,84be52..f315af,session46547,terminal exit2,
547.497s. H4 completes119.83s and H5 completes427.55s; no cancellation/restart.
H4 current census4a8f73: Casual32 starter pairs,28 finite (22 faster/six tied),
one treated-only/three both-unreached. Chaos64 starter pairs,all64 finite/faster.
Reference1 row/no starter. All-finite min/p50/max Casual0/80000/350000ms and
Chaos38000/118000/212000ms.194 sources admitted. Strict H4 still fires its six
ties and original v1 comparison is RED, not new acceptance.

H5 terminalf315af admits/retains970 sources and logs all9nodes ×3personas.
The following table retains every nonzero purchased group; fields faster/tied/
slower refer to finite unmasked-baseline versus masked comparisons, NOT causal
per-node effects. Baseline careers are32Casual/64Chaos/1Reference for every node.

| Node suffix | Persona | Bought | Finite | Faster/tied/slower | Unmasked-only/masked-only/both-unreached | Conditional p50 ms |
|---|---|---:|---:|---|---|---:|
| unlock.p05 | Casual | 32 | 29 | 1/23/5 | 0/0/3 | 0 |
| unlock.p05 | Chaos | 64 | 64 | 6/8/50 | 0/0/0 | -24000 |
| unlock.p05 | Reference | 1 | 1 | 0/0/1 | 0/0/0 | -109000 |
| starter.cash_small | Casual | 32 | 28 | 22/6/0 | 1/0/3 | 70000 |
| starter.cash_small | Chaos | 64 | 64 | 64/0/0 | 0/0/0 | 102000 |
| starter.generated_beige_tower | Casual | 23 | 20 | 17/3/0 | 0/0/3 | 15000 |
| starter.generated_beige_tower | Chaos | 53 | 53 | 44/6/3 | 0/0/0 | 30000 |
| unlock.p25 | Chaos | 42 | 42 | 0/0/42 | 0/0/0 | -46000 |
| starter.upgrade_continuous_feed_paper | Chaos | 22 | 22 | 21/0/1 | 0/0/0 | 40000 |

The other18 node/persona groups have zero purchases/finite pairs and no median.
Specifically cash_large/unlock.p50/unlock.p75/unlock.p100 are never bought by
any persona; Casual buys none ofp25/continuous_feed_paper, Reference buys only
p05. Not-purchased careers equal each group's baseline count minus bought,
not an exclusion from the declared cohort. Descriptive arithmetic over the27
logged groups:873 node-pair exposures,333 bought/540 not bought;323 finite bought
pairs (175 faster/46 ties/102 slower), one unmasked-only and nine both-unreached.
These are repeated node comparisons, NOT873 independent careers. H5 emits no
node-classification error; preserved report comparison stays RED. No epsilon,
run4, censoring decision or whole H5 acceptance follows.

Root `make verify-harness-fast`,8f820b..ad4b37,session29123,terminal exit0:
harness72.287s/role0.329s/Commons0.406s/guard; runs alongside exhaustive study,
so wall times are local observations, not performance baselines. Root `make vet`,
67432c,terminal exit0. Read-only historical H4 census08676b separately agrees
on Casual28 finite/six ties; it is not current masked-arm evidence.

Source/status/hash rechecks521a5f/6541ca remain clean; original H4/H5 SHA
648f36d6/4ed79054 and dated H1/H2 SHA3ce87b08/d5979c93 unchanged. No other runtime/
balance/corpus/CI/kernel edit or fresh hosted/SQL/browser/deployment claim.
RP-272/273 now have measured denominators, with censoring and per-node author
routes held. Next separately predeclare dated H4/H5 artifact provenance under
RP-263; do not silently refresh old reports or adopt the unratified career
fixture. Full new span aftera3e36f94 needs designated Claude review, independently
of every earlier span. No checkbox, mint, archive, cleanup, publication or push.

## 2026-10-06 — Population observation self-first-filter

Review by: Codex (implementer; self-first-filter, not designated).
Recorded by: Codex.
Reviewed range: `a3e36f94..a768864b` — all three commits and all11 paths,
including predeclaration, observation implementation and full evidence records.
Verdict: ready for designated review; no H4/H5 acceptance or author ruling.

Inspected all source changes: only test-side reporting changes; old strict H4
gate and H5 classifier/delta conditions are unchanged. H4 validates rows and
distinguishes no-starter/unreached observations from all-finite descriptive
package savings. H5 retains the original bought/finite estimator, with explicit
baseline/not-bought and all clock-state counts. Tests check exact synthetic
values, old gate/classifier outcomes and serialized observations. Eleven
compiling omissions fire with exact restores; full current census has no source
admission failure, but preserved comparisons and the six ties remain RED.
All counts/table arithmetic and zero-purchase groups cross-checked against
terminal4a8f73/f315af, including873 repeated node exposures, not independent
careers. This record clarifies a short board sentence to distinguish one
treated-only pair from three both-unreached; no statistic or rule changes.

Committed-HEAD cold focused check ata768864b, root `make test-go`, harness,
population/source/gate/classifier/static-artifact selectors, `-count=1`:
a622fc/e9f26e,session10374,terminal exit0,0.323s. All handles terminal before
this edit; source unchanged from full b22ce51e run. Range checkddbbed clean.
This wording/record edge must also join Claude's complete designated span.
Prior spans remain independent; no self-archival, checkbox, mint or push.

## 2026-10-06 — Dated career evidence preparation predeclaration

Previous goal turn made concrete progress: RP-272/273 source/report population
observations, full current census and tracking are committed through5e083766.
Revalidated clean HEAD and accepted R10. No new Claude verdict, cleanup approval
or release authority inferred. One unsuccessful read guessed types.go; rg
resolved RunKey to harness.go before use. No live test/probe handles at start.

RP-275 is registered immediately: H5 discards raw clocks/purchases, so its
retained aggregates cannot be recomputed from the report alone. This turn is
the bounded preparation stage for RP-263's dated career evidence: add raw H5
observations and shared test-side report composition; do NOT generate the new
dated artifacts yet. Later recording/replay must be separately predeclared on
the exact committed producer, preserving all old reports and author boundaries.

Authority: accepted R10 H4/H5 observation and evidence discipline. Scope:
existing H4/H5 test report producers, one diagnostic file and canonical docs/
tracking. No runtime/kernel/balance/CI/data/policy/horizon/epsilon change.
Retain existing strict H4 and finite-only H5 statistic/classification. Preserve
historical JSON, with update flags off. Share report composition rather than
creating a second report algorithm for a later opt-in instrument.

Before correction, an independent JSON-side diagnostic must compile against
the current producer and fail for absent raw H5 observations. Then retain each
admitted arm's source/gate/purchase list, copying mutable data so an input edit
cannot silently revise retained evidence. Recompose H4's full gate/population/
savings and H5's nodes/populations/medians from declared complete arm groups;
the measured source tuple is not policy adoption. Test synthetic coherent
baseline/mask groups, finite and unreached observations, unchanged summaries,
serialization and independent retained recomposition. Refuse incomplete,
duplicate, unknown-mask, mismatched-source and inconsistent purchase/clock
populations. These are statistical fixtures, not naturally earned careers.

Show separate compiling omissions for raw retention/copying, group admission
and aggregate composition, awaiting each same handle to terminal and restoring
exact bytes. Cold focused/fast/vet checks and one complete existing97-pair/
970-arm invocation; finish without restart/retune. Expected H4 six ties and
both historical comparisons remain RED. Full results change that expectation
only through observed evidence, never a loosened criterion.

Exit: report data can be recomposed with executed controls and full current
caller coverage. No fresh dated provenance claim at this preparation stage,
censoring decision, per-node causal attribution (RP-274), H3/H5 author ruling,
career artifact/policy ratification (RP-268/271), epoch/acceptance/archive/SQL/
browser/deployment/cleanup/push. Full new span after5e083766 needs Claude,
independently of all earlier ranges; no delegated/self archival.

## 2026-10-06 — RP-275 shared report composition preparation

Baseline at1f6a09ca,bb2a72/e60598,session70948,terminal exit2,0.319s: independent
JSON-side check compiles and fails absent raw arms (0 versus2). Corrected raw
retention and existing source/population controls3c22aa/323360,session71380,
terminal exit0,0.319s. Refined composition controls416ffc/ab3ec5,session82209,
terminal exit0,0.343s: coherent synthetic20-arm/two-seed groups, ten malformed
groups and eight retained-report corruptions; H4 positive/tie/exclusion fixture
keeps strict FAIL and both savings populations. None is a naturally earned or
full ratified cohort. An added internally consistent zero-effect bought starter
fires existing H5 report admission; a6aabe/c899db,session93977,terminal exit0,
0.338s after restoration. Final focused raw/composition/source/population check
bb8d03 follows the one logging-line addition, not a new full-population claim.

Eleven separate compiling omissions all fail semantically: raw retention18c0a3,
clock copyf38592, purchase copyd805d2, group census5ded9c, paired source51b347,
negative clock2040c3, purchased-node admission7e7d30, mediancd98bf,
retained equality5f5ea2, H4 gate status9a19d8 and H4 passing summaryf0a930.
Each same handle terminates, then exact source restores: H40c3f4795,
H5b99323a7, diagnostic28d2899f. No compilation or panic failure counted.
Subsequent H5 criterion admission82d96926/diagnostic345787ca gets its own
compiling guard omission911cb9: the internally consistent invalid report is
wrongly admitted and the test fails. Restores exactly82d96926. One final
observational log line yields H5b8aadfa5; unchanged H40c3f4795 and diagnostic
345787ca. No live-run edit or undocumented restoration delta.

H4 measurement/gate/statistics now share a test-side constructor; its original
criteria and default v1 comparison remain. H5 shares composition across raw
arms and retained-report recomposition. It still uses purchased baselines and
finite-only deltas, the same median index/epsilon/classifier. Full measurement
calls the raw recomposition check before reporting; invalid source/group data
cannot silently become an aggregate. Existing invalid H5 node criterion remains
FAIL (now also refused during recomposition before ordinary diagnostic output).
No runtime/balance/kernel161/CI/policy/corpus/old-report byte change. Cold
fast/vet and full97-pair/970-arm caller run next; no dated artifact generated,
acceptance, author adoption, archive, cleanup, SQL/browser/deployment or push.

## 2026-10-06 — RP-275 full current caller and recomposition evidence

Committed producerd96bfaf1, server tree560c4976d35e361043a10a12d08350d34170120e;
balance treecd982b7c58a53cac0ab4c91a773705033930414d/kernel0.3.161 unchanged.
Final source SHA: H40c3f47954867cd4aa1c13948ec29b38ae0b5bd8e6f86cfba1df25344616793d8,
H5b8aadfa5e54499c27c3352878b226ac83d3b6e4227a8d068842f826597ab4c68,
diagnostic345787ca288430e29de07b3bc4932892fbc5b0013f23af6b6da29510b0476b34.
Final restored focusedbb8d03/9bc6b4,session78083,terminal exit0,0.271s.

Root `make reputation-harness-check`,222aab..d73e16,session56244,terminal exit2,
641.626s. H4 completes130.03s;3c1729 reproduces all97 rows/194sources,
93 timing comparisons/3exclusions, identical savings and the same six Casual
ties. Original career comparison RED. H5 completes511.44s;d73e16 explicitly
logs970 raw observations retained/recomposed and970 admitted/retained sources.
All27node/persona population groups and conditional medians match the preceding
table (b22ce51e); no source/group/recomposition/node-classification error emitted.
The original relevance comparison remains RED. Full run not cancelled/restarted;
no update flags, edits or record changes while either check handle lived.

Root `make verify-harness-fast`,20306d..d819b6,session3681,terminal exit0:
harness73.940s/role0.266s/Commons0.401s/guard. Root `make vet`,df0047,terminal
exit0. Both are cold/local and fast ran alongside the exhaustive measurement;
wall times are not a performance comparison. A read-only ps diagnostic2fc0b0
was sandbox-denied, not a missing/terminal test handle; continued polling the
same PTY through its actual exit. No workaround or restart.

Source/status/hash checksff782e/9ebe37 remain clean; original H4/H5 SHA648f36d6/
4ed79054 and dated H1/H2 SHA3ce87b08/d5979c93 unchanged (3da497). All test
handles now terminal before these tracking edits. RP-275 raw recomposition is
locally exercised; it proves internal consistency, not authentic earned
synthetic populations or a new dated producer artifact. No raw report written
at this preparation stage. Next separately predeclare exact-cohort/source-
bound dated H4/H5 artifacts and complete replay under RP-263. H3/H4/H5 author
criteria/policy routes, RP-268/271, RP-274 and SQL/browser/deployment/capacity/
review union remain. Full new span after5e083766 needs Claude, independently
of every prior span. No kernel/live math/balance/CI/corpus/owner-copy change,
new dated report, box, acceptance, mint, archive, cleanup, publication or push.

## 2026-10-06 — Raw report evidence preparation self-first-filter

Review by: Codex (implementer; self-first-filter, not designated).
Recorded by: Codex.
Reviewed range: `5e083766..e6067ae1` — all three commits and all11 paths,
including predeclaration/baseline, shared builders and full evidence records.
Verdict: ready for designated review; preparation only, no acceptance or archive.

Inspected the complete source/test/tracking diff. Raw H5 fields preserve the
admitted source/clock/purchases and copy mutable values. Ordinary checks now
share H4 gate/statistics and H5 composition with retained-data recomposition;
old comparisons and update flags remain, unchanged/off in executed runs.
Finite-only estimation, strict ties and classifier are not loosened. New group
admission validates complete baseline/mask structure and matched sources;
internal consistency explicitly does not authenticate the cohort or producer.
Nineteen corrupt/invalid controls and twelve compiling omissions are recorded
with honest synthetic limits and exact restoration; the subsequent H5 criterion
guard/log-line deltas are separately disclosed. Full current caller execution
retains/recomposes970 arms and reproduces the prior census/medians, still RED.
No private observation field is claimed as a public schema or live epoch.

Committed-HEAD focused check ate6067ae1, root `make test-go`, harness,
raw/composition/static-artifact selectors, `-count=1`:160322/f447e9,
session82333,terminal exit0,0.279s. Source identical to full d96bfaf1 run;
later changes are records/docs, not a new full measurement. Range whitespace
check6267ba and status eb8c35 clean. No live handle remains before this edit.
This record edge also belongs in Claude's full designated span; every earlier
span remains independent. Goal active; no box, author ruling, new dated report,
mint, archive, cleanup, publication, deployment, message to Claude or push.

## 2026-10-06 — RP-263 dated career observation predeclaration

Baseline5c93ff6f, clean. Continue accepted R10's measurement lane using the
shared H4/H5 producers, not new gameplay or acceptance criteria. Private opt-in
`-reputation-career-measurement=off|record|verify`; unknown selectors fail,
default explicitly skips fresh measurement. No Make/CI lane change. Outputs:
`career-h4.2026-10-06.v1.json`, `relevance-h5.2026-10-06.v1.json` and
`career-measurement-lineage.2026-10-06.v1.json` in this directory. Record refuses
every existing/unresolved output before production and writes exclusively;
partial writes fail loudly. Original/different dated reports remain untouched.

Declare the pinned suite's ordered64 Chaos/32 Casual/1 Reference seeds. H4
retains97 treated/control pairs (194 sources); H5 retains97 baseline groups,
each followed by the nine fixture node masks (970 sources/raw observations).
Admit exact expected RunKey/paired fixture hash, effective1e5 Prestige hash,
first-hour policy, experiment, horizon, purchase policy and mask. Recompose
whole H4 and H5 reports and cross-bind each H4 treated observation to its H5
baseline gate/purchases/source. Bind both raw report hashes to committed server/
balance trees/kernel and native Go/OS/architecture. Dirty or changed inputs
before/after production fail; record-only advancement may not change inputs.

Controls: complete synthetic97/970 admission fixture is not an earned cohort;
missing/duplicate/reordered coordinates, semantic source mutations, report/
census/gate/median corruption with rebound raw hashes, contradictory baseline,
lineage identity/hash/count corruption, strict JSON and selector/overwrite
refusal. Demonstrate compiling guard omissions and exact restorations, with
all test handles terminal before any edit. Execute cold focused/fast/vet lanes,
then record and independently execute full verify (97 paired careers plus970
arms each), requiring byte-identical H4/H5 output. Static retained-artifact
validation is not fresh execution or producer authenticity on its own.

Success is trustworthy reproduction, including truthful H4 FAIL/six ties;
never promote a negative measurement to accepted H4/H5. No epsilon/horizon/
policy/threshold/mask/classifier change, censoring adoption, per-node attribution,
RP-268/271 author reconciliation, balance/kernel/runtime/corpus/CI/owner-copy
change, acceptance box, mint, archive, SQL/browser/deployment/cleanup or push.
New full range after5c93ff6f and its record edges need designated Claude review;
every earlier range remains independently pending. Goal stays active.

## 2026-10-06 — RP-263 dated career instrument and admission evidence

Only two new harness test files; shared gameplay/builders unchanged. Complete
synthetic97/970 cohort is independently declared from pinned inputs, not earned.
Thirteen lineage corruptions,36 report corruptions (including six rebuilt,
internally consistent forgeries), two strict JSON controls, positive admission
and truthful one-tie negative admission pass. An independent JSON-side
descriptor checks every declared arm. Initial admission f79d76/6174f2,
session14571,exit0,0.971s. Refined selector d1b93d/ec7dd8,session66716,exit0
ran NO tests because escaped alternation matched literally; not counted as proof.
Corrected prefix3b50ec/96470d,session37690,exit0,9.007s explicitly executes
admission/declaration and all four real source fingerprint controls. Artifact
check explicitly skips before the recording commit; no artifact claim here.

Eight separately compiling guard omissions fire semantic failures: identity
655e64, raw hash9794e9, census bca955, H4 coordinates2f8726, H5 mask order
5978f9, H4/H5 baseline809a3b, H4 recomposition3d6ff6, applied-starter admission
f6584b. Rebound hashes/rebuilt reports are wrongly admitted under their respective
omissions, proving independent discrimination rather than incidental checks.
Every same PTY terminates exit2 before exact restoration; no build/panic failure
counted. Restored sources1d9c5f2bb4ec5ecb88a117aae0a3096fb67a62ded222d634545c03c455886495 /
3d3ba067f9bc88b5b9313db52874846857004f5987688d6b90566b27d62953d4.

Actual unknown selector42c2d6/f93029 exits2 before production. Actual dirty
record3008f4 exits2 before production, listing only the two new uncommitted test
files. Restored cold focused0ab644/03509f,session50802,exit0,8.964s.
Root `make verify-harness-fast`,346fe9..5141ae,session3535,terminal exit0:
harness66.711s, role0.324s, Commons0.351s and guard. `make vet`,471264,exit0.
All handles terminal before these edits. Whitespace81ad2a clean. No body/literal/
runtime/math/balance/kernel161/CI/corpus/old-report change or checkbox flip.

Next commit this instrument, then record and replay the complete real population
with exact source identity. No current dated career evidence, acceptance,
author-policy adoption, RP-268/271/274 resolution, mint, archive, SQL/browser/
deployment/cleanup/publication/message to Claude or push. Complete span after
5c93ff6f plus later record edges still needs designated Claude review, separately
from all earlier ranges. Goal remains active.

## 2026-10-06 — RP-263 dated career reports recorded and fully reproduced

Committed producer6bd24e8d9e211b68aa0bb04dd18f4f02653616a8,
server treec3d6616e9b33ff94e71ea452e0c1d5ffddebbef5,
balance treecd982b7c58a53cac0ab4c91a773705033930414d/kernel0.3.161 unchanged;
native Go1.27.1/darwin/arm64. Root `make test-go`, harness with `-args
-reputation-career-measurement=record`, exact CurrentMeasurement selector,
`-count=1 -timeout 60m -v`,1b50f4..c6d633,session5519,terminal exit0,
676.102s. Complete97 pairs/194 sources/970 baseline-mask arms are admitted,
retained and recomposed; H4/H5 baseline observations agree. GatePassedfalse,
six existing Casual ties,93 comparisons/3 valid exclusions, all27 H5 groups
and medians unchanged. This is completed negative research, not H4/H5 success.

Root same complete selector with verify,c5328d..9e1da4,session81662,terminal
exit0,469.043s. Fresh97 paired careers and970 node-study arms reproduce both
whole report files byte-identically; all same input guards pass before/after.
No cancelled/restarted run, update flags, manual edits or record/source changes
while any test handle lived. Timings are not a performance comparison/budget;
fast check ran concurrently with replay. Companion records exact producer/runtime.

Raw SHA H4:
29d0e6d2758502db3e82578db349cc1e080535afb74be0ae77d5df0e0ca2f3f7
Raw SHA H5:
8044667158f80ccd920cce5861bd029fa3d46339c77fdb4a7230b6d3e9b404fe
Companion SHA:
ba9fcc7b458e3d4bc2e5038e893d72731af52d040571609d1c18ae7612676ad7
H4/H5 sizes263025/2257335 bytes (3a0d60); no output truncation/reduced cohort.
Actual overwrite8a8d7d/44ba75,session30606,terminal exit2, refuses existing H4
before production. Retained-artifact validation now executes rather than skips:
2e85ff/88a91a,session6158,terminal exit0,9.102s, including51 refusal children,
full declaration and the four unchanged real source fingerprints. Source/datum
SHAs603ab8 and committed trees07b2ec unchanged after all runs.

Root post-artifact `make verify-harness-fast`,cc553d..289ccb,session32025,
terminal exit0: harness98.632s, role0.375s, Commons0.501s and guard. Earlier
vet471264 remains the unchanged-source cold pass, not relabelled as another run.
Read actual CI harness checkout/entrypoint0d21b2/d965c1: full Git history and
same fast target, not hosted execution or full CI proof. Existing report SHAs
dfbdb7 stay648f36d6/4ed79054 and dated H1/H2 stay3ce87b08/d5979c93. They
remain earlier producer evidence, not freshness at this changed server tree.

Read-only Docker recheck8ad30e/0f562c/40e8a0 still reports100% root,39784KiB
free, DB tmpfs8108540KiB free;342 images78.3GB/143 volumes39.36GB/cache2.634GB.
Reclaimable is not deletion authority. No resources removed, no Docker tests
or browser/deployment attempts. All handles terminal before tracking edits.
Failed multi-file patch context verification left the tree untouched (853c63)
before the corrected patch; no hidden partial edit.

RP-263's local dated career observation gap is now reproduced, not whole R10
acceptance. RP-268/271 author/data intent, RP-274 attribution, H3 tiny criterion,
H4 strict failures, H5 epsilon/run4/censoring, owner SHA adoption/mint and real
SQL/default-player proof remain. Next safe work grounds remaining accepted
R9/AC12 client-consumer coverage from B7 and HEAD before any bounded probe;
do not silently resolve these author holds or absent Firefox/DB populations.
Full range after5c93ff6f, including these artifacts/records and final edge,
needs designated Claude review; every earlier range is separate. No production
math/balance/kernel/CI/corpus/owner-copy change, checkbox, acceptance, archive,
cleanup, deployment, publication, message to Claude or push. Full1.0 goal active.

## 2026-10-06 — Dated career observation self-first-filter

Review by: Codex (implementer; self-first-filter, not designated).
Recorded by: Codex.
Reviewed range: `5c93ff6f..5e9b9c6a` — all three commits/all14 paths,
including predeclaration, both test files, three generated reports and records.
Verdict: ready for designated review; no acceptance or archival approval.

Read full source/tracking diff (c5cbbf/b09c12/c1868e), confirmed scope fe8972/
153e5a and whitespace3389ae/bfefc0. Only new Go test files; shared producers,
strict gates/estimators and all old report/data/runtime/kernel/CI bytes remain.
Exact-declaration admission, whole recomposition and cross-report baseline
binding discriminate independently through compiling omissions. Synthetic
controls are never labelled earned. Both complete native executions and exact
raw-byte equality authenticate this declared observation; they do not create
the missing named career artifact, owner policy, public epoch, run4/SQL/player
workflow, independent review or whole-platform readiness. H4 FAIL is retained,
not disguised by the green measurement test. Earlier H1/H2 remain historical
producer evidence; no freshness restamp. Exclusive outputs and dirty/unknown
selectors were actually exercised. No mutation/edit under a live handle.

Committed-HEAD retained-artifact check at5e9b9c6a,root `make test-go`, harness,
exact Artifacts selector,`-count=1 -v`:5c4916/84cdc1,session74634,terminal
exit0,0.651s (validation, not another full producer execution). Server/balance
trees de848f remain c3d6616e/cd982b7c. All handles terminal before this record.
This record edge also belongs in Claude's full designated span, independently
of every earlier pending range. Next accepted R9/AC12 consumer grounding stays
queued; no owner-copy change, checkbox, acceptance, archive, cleanup, deployment,
publication, message to Claude, push or goal completion. Full1.0 goal active.

## 2026-10-06 — R9 purchase focus diagnostic predeclaration

Baseline020a25c6, clean. Previous goal turn is progress (dated career artifacts
recorded/reproduced, not release acceptance). Re-ground accepted R9/AC12 B7 at
HEAD. cb95dc shows confirm clears the inline pair, calls onPurchase, then tries
to focus an unmounted Buy button; existing a081af test asserts submission but
never checks post-submit focus. RP-276 records the suspected loss, not a runtime
verdict yet. This bounded diagnostic is actual Svelte/native browser keyboard,
not real server/receipt/SQL/epoch/player integration.

Test-only reactive Svelte5 harness passes existing props to the real component;
its exposed setters supply controlled pending and authoritative-arm replacements.
No internal Svelte runtime mocks or production fixture seam. Native Enter/Space
Buy→Confirm cases in both supported eras; held pending disables all buy controls;
injected applied/owned and rejected/available arms retain the exact row and
focus. Native Escape cancels to Buy with no purchase callback. Preserve original
three tests and their axe checks. Attempt declared Chromium/Firefox/WebKit via
root make test-browser on the installed native runtime. Any non-launch/missing
population is invalid for that engine; never weaken deadlines or substitute
another engine, Docker cleanup, browser flags or a passing node-only skip.

First execute baseline unchanged production. If actual focus loss reproduces,
separately predeclare the accepted R9 correction before changing production:
stable row focus after DOM settlement, no new copy, eligibility, API, receipt,
balance/kernel, schema or transport policy. Demonstrate a restored compiling
severing, retain Enter/Space/Escape and owned/no-enabled-button controls, then
cold client/type/build and independent boundaries. Host/public/mint/whole AC12,
other row feedback/presentation requirements, native Firefox/RP-256, SQL/capacity,
H3/H4/H5/author/review holds and full nine-tier1.0 remain separate. All new range
after020a25c6 requires Claude; no box, archive, cleanup, publication or push.

## 2026-10-06 — R9 focus failure reproduced; correction predeclared

Unchanged production, native all-engine run9ca0fc/73f15a/36bef9/a1ac62,
session85290 terminal exit2: Chromium and WebKit each execute13 cases,
eight purchase-focus failures and five passes. Firefox executes zero: native
launch fails sandbox extension/SWGL, times out180s (RP-256), never counted green.
Both engines show actual activeElement body while the submitted row's sole Buy
control is disabled. Escape and existing rendering/axe/plan cases pass.

Test refinement explicitly captures initialArm via public untrack (avoids the
Svelte initial-capture warning), requires exactly one pending control before
the disabled census, and keeps the injected owned arm's dependent tower state
consistent with the lower budget. Repeated unchanged-production diagnostic:
root make test-browser with exact file and explicitly selected chromium/webkit,
ed0be4/296615,session39864 terminal exit2; same16 failures/10 passes. Selection
is a bounded two-engine diagnostic, not full AC12. No source edits under live
handles; original three-engine failure remains on record.

Correction authority is accepted RFC R9's post-purchase row focus. Predeclare
only a stable keyed row attachment, programmatic tabindex=-1 (no added Tab stop),
and focus of that row after Svelte DOM settlement when confirming. Keep callback,
server-derived eligibility, pending policy, owned/no-control rendering and
Escape-to-Buy unchanged. No transport/receipt promise seam, copy, balance, save,
kernel, schema or CI changes. Canonical UI docs follow this bounded behavior.
Execute the same26 two-engine cases; then temporarily sever only the row-focus
call in compiling source, require the eight cases per engine to fail, restore
exact source bytes and rerun. Run cold types/build/client and declared independent
boundaries. Full Firefox, real host/SQL/purchase receipt/minted epoch, other R9
requirements and whole AC12 remain separate. Full span after020a25c6 requires
Claude's designated review, including these tests and record edges. No checkbox,
acceptance, archival, cleanup, publication or push.

## 2026-10-06 — R9 row-focus correction and executed discrimination

Under770a97bb/930b69d9, confirm now awaits DOM settlement then focuses the
stable keyed row. Generic typed attachment registration supports the row's
tabindex=-1 without adding a Tab stop. No callback/eligibility/copy/pending/
receipt/schema/balance/kernel/CI change. docs/game-ui.md describes this behavior.

Root native exact-file chromium/webkit run eafdf9/d3a434,session38824 terminal
exit0:26/26 (13 per engine). Mandatory root-target follow-on performance lane
also passes its one selected Chromium case with22 explicitly skipped; this is
not complete performance/Worker/default-player acceptance. Typecheck6e070a/
69ebb1,session58899 terminal exit0,zero errors/warnings.

Compiling severing removes only rows.get(id)?.focus():1af790/3e58a9,
session24246 terminal exit2,16 purchase failures/10 passes, matching baseline
body-focus failure. Restore9654d1 byte-exact SHA256
442d53a68c6224c8487edc0d5c9f4714c9aa20e75a6d6672a70c17673c857f10
equals564b35. Restored native486a1d/3f6fe1,session72116 terminal exit0:
26/26 plus the same selected performance case. No edits under a live handle.

Cold root build-client/test-client and client/topology/combat/meters/achievements/
cosmetic/no-payment boundaries/copy-check: a61878/1966b9/88c5a5,
session60079 terminal exit0. Build succeeds;8106 unit tests pass,144 visibly
skip across91 passed/17 skipped files. Those native-only skips are not counted
as browser executions. Copy657 keys and deployment content manifest pass.
Kernel-history899256/af34b2,session34953 terminal exit2: CI checkout/negative
fixtures pass, historical50a3a514 against0cf9f7a6 still fails RP-131. Thus NOT
whole verify-client/verify/CI green; current renderer is outside watched kernel
prefixes and kernel161 remains unchanged, no false retrospective bump.

One authorized outside-sandbox Firefox-only retry8bc575/06507b/ad6501,
session76800 terminal exit2,zero tests: same sandbox-extension/SWGL diagnostics
and180s launch timeout. No browser flags/deadlines/config changed. This rules
out claiming the execution-tool sandbox escape as a working workaround, not
the underlying host cause. RP-256 remains; no third-engine AC12 claim.

Read-only R9 reconciliation finds separate gaps (c5ccee vs actual247e63/
9cc54c/04fab8/5946c1/923d0c): pending row has no aria-busy (RP-277);
rejection is host-global, no row input, revision conflict uses shared notice
(RP-278); cost only appears in available Buy text, not every row via Amount
(RP-279). These are source-contract findings, not newly executed host proofs.
Next predeclare RP-277's bounded pending-row baseline first; do not silently
extend this focus correction to other requirements. Full new span after020a25c6
needs Claude, separately from all earlier ranges; no checkbox, acceptance,
archive, cleanup, publication, deployment, push or full1.0 promotion.

## 2026-10-06 — R9 focus self-first-filter checkpoint

Review by: Codex (implementer, self-first-filter; not designated).
Recorded by: Codex.
Reviewed range:020a25c6..055cc70f — all four commits/all11 paths,
including predeclaration, test-only baseline, renderer/docs and tracking edges.
Verdict: ready for designated review, not RFC acceptance or archival approval.

Full code/test/docs/log diff524790 and tracking diff3e1182 read. Whitespace
e66bae/8b1efa and full-range check pass. Tests require exactly one disabled
pending control, exact retained keyed row, native submission, both era/outcome
arms and callback cardinality; original tests/axe remain. Baseline and single
compiling focus-call removal independently fail at the intended focus assertion;
restoration is byte-exact. Typed attachments and tick change only the focus
target/timing, with no extra Tab stop or new transport/API/copy/eligibility seam.
Server/balance trees88e82f remain c3d6616e/cd982b7c. Owner copy, kernel161,
CI, corpus and research reports unchanged; no checkbox flips. Typecheck/build/
units/boundaries/copy pass but historical RP-131 and zero-executed Firefox remain
RED, as records consistently say. Controlled prop injection is not a real
receipt/revision/database/default-player proof or whole AC12.

RP-277/278/279 are explicitly source-contract findings, not fixed by this range.
Next bounded RP-277 diagnostic is safe accepted R9 work, not blocked on policy;
all prior author/mint/capacity/review obligations remain. This self-record edge
must also be included in Claude's full span after020a25c6, separate from every
earlier range. All handles terminal before edits/record. No archival, cleanup,
publication, deployment, push or full1.0 completion. Goal turn is concrete
progress; full nine-tier/platform objective remains active.

## 2026-10-06 — R9 purchasing-row diagnostic predeclaration

Baseline5bf6017f, clean main; previous goal turn was progress (RP-276 focus
reproduced/corrected/discriminated, records reconciled). Re-read root process/
AGENTS and active accepted index; RFC/design authority unchanged since020a25c6.
Current surface898d69 has no aria-busy. Host4f3251 returns an existing act
Promise which can await a preceding refresh before setting global pending;
pending can also outlive a conflict receipt during refresh. Thus "every row
busy whenever global pending" or retaining a last node indefinitely is wrong.
RP-277 remains a source finding until executed; no host-proof inference.

Test-only native diagnostic: two eras × Enter/Space × synchronous/delayed
host-pending start × controlled owned/available arm =16 cases per executed
engine. Use a legal-shaped two-available-row fixture and controlled returned
task; no receipt/transport implementation. Non-vacuous assertions require only
the submitted row busy, all Buy controls disabled during its returned task and
parent pending, retained focus, no extra submission, and no stale busy marking
under unrelated global pending after task/snapshot settlement. A subsequent
purchase must mark its own row, not the previous one. Keep all prior13 cases
and axe controls unchanged except callback braces if stronger return typing
later requires them. No production change before baseline execution.

Use root exact-file chromium/webkit selection explicitly: Firefox remains zero
after both normal/outside-sandbox attempts last turn; no new installation,
security flag or deadline workaround, no three-engine/full AC12 claim.
If red, separately record/predeclare bounded correction before production:
typed existing callback may return void or the host's existing Promise<void>;
track the submitted row while that task OR parent pending remains and clear
after both finish. Preserve global pending/transport/revision/refresh semantics,
server-derived eligibility, owner copy and RP-276 focus. Only local presentation
controls enforce R9's existing one-in-flight rule. Execute baseline/restored
suite, compiling busy-attribute and stale-attribution negative probes with exact
restoration, cold types/build/client/boundaries/copy. No live-handle edits.
RP-278/279, host/SQL/mint/full AC12, author/data/H3/H4/H5/R11, Firefox/capacity,
historical RP-131 and cross-party reviews remain separate. Full new span after
5bf6017f needs Claude, including final record edges; no box, acceptance, archive,
cleanup, publication, deployment or push. Full nine-tier/platform1.0 active.

## 2026-10-06 — R9 busy baseline fails; presentation correction predeclared

Unchanged5bf6017f renderer, root native exact-file chromium/webkit baseline
6d3ccc/567d1a,session1237 terminal exit2:32 new busy failures/26 existing passes
(16/13 per engine). Each new case fails empty busyRows vs exact submitted row;
delayed-start screenshot also shows enabled Buy controls while its returned task
is held. This is controlled props/task evidence, not executed host/SQL proof.
Typecheck555c4a/f9444e,session25841 terminal exit0,zero errors/warnings.
Refine the second submission to hold parent pending after task settlement, so
the same population covers both settling orders; baseline first failure unchanged.

Correction, now before production: explicit callback void|Promise<void> (host
already returns4f3251 act); local submitted-node and task flag; await that existing
task without moving transport or awaiting before RP-276's focus handoff. While
local task or shared parent pending is true, disable Buy/Confirm and mark only
the submitted row aria-busy. Clear attribution only after both settle; guard a
duplicate Confirm against the same existing one-in-flight/disabled policy.
Sync legacy callbacks remain valid; test push callbacks get braces solely for
void typing. No new receipt/error/revision/refresh/eligibility policy, copy or
host transport change. Update canonical UI docs with this exact local behavior.

Require58 native cases green. Demonstrate independently compiling busy-attribute
removal, stale-attribution-clear removal, task-await removal and row-equality
removal; each must produce the declared native assertion failure, not a build
failure. Restore exact bytes only after each handle is terminal and rerun. Cold
types/build/client/independent boundaries/copy; kernel history remains separate
RP-131, renderer outside watched prefixes. No new Firefox launch workaround or
whole client/CI/AC12/minted-player claim. Full span after5bf6017f needs Claude;
all prior review/policy/mint/capacity/full1.0 holds remain. No checkbox, archival,
cleanup, publication, deployment or push.

## 2026-10-06 — R9 purchasing row correction and executed evidence

Under367fe467/27ff4cd5, explicit void|Promise<void> callback typing reflects
the host's existing returned act task (host/runtime bytes unchanged). Local
submitted-row/task state plus shared pending disable Buy/Confirm, mark only
the submitted row busy, and clear attribution after both settle. Confirm guards
the existing one-in-flight/controls floor. Awaiting the task comes AFTER tick/
row focus, preserving RP-276. Canonical docs updated; two prior test callbacks
use void braces, fixture typing follows the public prop, no new transport seam.

Root native exact-file chromium/webkit00ce66/937c5a,session46360 terminal
exit0:58/58 (29 per engine). Typed78b446/1ffc18,session23076 terminal exit0,
zero errors/warnings. Every new case covers both task-first and parent-first
settlement, unrelated pending before/after purchase, and second-row attribution;
old13 native cases/axe/focus/Escape remain. These are controlled tasks/arms,
not executed receipt/host/SQL/mint/whole AC12 evidence.

Four independently compiling source probes, same root exact-file population,
each terminal exit2 with32 new failures/26 prior passes:
- e2e37e: omit aria-busy attribute, exact submitted-row assertion fires.
- 80efc8: omit attribution clearing, unrelated pending revives completed row.
- 093e1d: omit task await, busy/disabled lifetime ends before held task.
- c02796: omit row equality, unrelated parent pending marks all four rows.
After EACH terminal probe, source restored to934cc5 SHA256
6db914d640d65abeb504bf82445e207f82b0c3be71b63cc188dcb95fb4855314
(146853/e6ff19/43522f/44fca8). Final restored020938/55a5b6,
session81884 terminal exit0:58/58. Both green root runs' automatic performance
follow-on passes one selected Chromium case/22 skipped, not full Worker/perf
acceptance. No edits under live handles, no restored-by-syntax-error probe.

Cold root types/build/unitclient and client/topology/combat/meters/achievements/
cosmetic/no-payment/copy checks2cba33/98d195/9706ef,session14516 terminal
exit0:zero type errors/warnings, successful build,8106 unit passes/160 visible
native skips across91 passed/17 skipped files; boundaries/negative fixtures,
657-key copy pipeline and deployment content manifest pass. Native-only Node
skips are not browser successes. Kernel historyd8a73b/19887c,session68838
terminal exit2: CI checkout/negative fixtures pass; historical50a3a514 vs
0cf9f7a6 remains RP-131 RED. Renderer outside affecting paths324558; kernel161
unchanged. NOT whole verify-client/verify/CI green. Prior Firefox zero-execution
is not retried or counted green this turn; no security/deadline/config workaround.

RP-277 corrected locally, not full R9/AC12. Next ground RP-278's actual host
outcome/resync/row path before predeclaring correction; RP-279 remains separate.
Prior H3/H4/H5 author/data/SHA/mint, SQL/capacity/Firefox, full nine-tier/platform
and all designated-review holds remain. Full new span after5bf6017f needs Claude
including record edges, independently of every earlier range. No checkbox,
acceptance, archive, cleanup, publication, deployment, push or goal completion.

## 2026-10-06 — R9 purchasing row self-first-filter

Review by: Codex (implementer; self-first-filter, not designated).
Recorded by: Codex.
Reviewed range:5bf6017f..e7b68f84 — all four commits/all11 paths,
including predeclaration, red baseline, renderer/docs and tracking edges.
Verdict: ready for designated review, not RFC acceptance or archival approval.

Complete source/fixture/docs diffdf4dbb, test diff016bb1, log0384a8 and
tracking e3352f inspected; full-range whitespace977d01 passes. Host callback
already returns4f3251 task; no host/runtime/receipt/refresh bytes changed.
Void-compatible callbacks and both settling orders pass; local attribution
clears only after task/shared pending, not by interpreting outcomes or deriving
server state. Submitted-row cardinality, non-vacuous disabled-control census,
unrelated refresh and second purchase assertions discriminate through four
compiling negative probes; restoration exact, previous focus/Escape/axe intact.
Native58 pass and types/build/unit/boundaries/copy are genuine executions;
Node native skips, selected performance case, Firefox absence and historical
kernel guard RED are not relabelled whole AC12/CI/host/SQL/mint proof.

Scope e1c245/c92dc8 remains one UI renderer plus tests/docs/tracking; server
c3d6616e and balance cd982b7c unchanged, kernel161 outside watched UI path,
owner copy/artifacts/research reports/CI untouched. No checkbox flips. RP-278
actual host outcome/resync/row diagnosis is next, not replaced by a child-only
error stub. Prior owner/author/mint/capacity/all-review/full1.0 holds remain.
This final self-record edge must join Claude's full span after5bf6017f, separately
from earlier spans. All handles terminal; no self-archive, cleanup, publication,
deployment, push or goal completion. Current goal turn is concrete progress;
complete nine-tier/platform objective stays active.

## 2026-10-06 — RP-278 host/runtime diagnostic predeclaration

Resume at b8ee639f, clean main, all preceding handles terminal. R9 row rejection
feedback is next; RP-279 cost remains separate. Grounding reads actual act,
noticeForOutcome/noticeForError, runtime HTTP parsers and subscription protocol.
Do not put revision_conflict into SurfaceRejections: its early return would
silently remove the shared refresh effect. The documented B7a optional eight-key
arm / feature.reputation_tree convention stays unchanged, not a new wire edit.

First use canonical buildCopyArtifact to census the five R9 rejection keys.
Source inspection finds four; the required revision_conflict key is absent.
Record that separately as RP-280 if executed census confirms. No owner prose
invention: the accepted owner block expressly permits clearly marked placeholders;
any eventual declaration must use the existing PENDING OWNER COPY convention.

Add a native host diagnostic using GameUIApp and the real browser runtime with
controlled HTTP Responses and Centrifuge protocol messages (NOT a live server,
WebSocket, database or minted tree). Wire-valid arm, distinct Company/Founder
revisions, two available rows, both supported eras, native Enter/Space. Hold the
POST to observe busy/disabled/no optimistic ownership, then deliver each of five
HTTP-200 rejection categories plus typed HTTP409 conflict. Assert exact inline
copy on submitted row only, no raw detail/ID, correct Founder revision/unique
intent ID. Hold conflict refresh to prove one additional GET and disabled controls;
deliver new Founder revision, retry a different row and prove revision/attribution
and old feedback clearance. Applied-control test holds authoritative refresh;
no optimistic state and existing global polite result remain required.

Run unchanged-production baseline before correction. Missing conflict declaration
is an explicit failing prerequisite, not generic intent.conflict acceptance.
Native chromium/webkit only; Firefox prior zero-execution remains RP-256. Include
existing child focus/busy and shared GS0.2 host rejection as separate controls.
No source edits while any handle is live. No whole AC12/AC15/SQL/CI proof, mint,
retune, kernel bump, checklist flip, archive, cleanup, push or publication.
Correction scope and compiling omission probes will be declared after baseline.
Full new range after b8ee639f needs Claude's designated pass including record edges.

## 2026-10-06 — RP-278 baseline and bounded correction predeclaration

Canonical copy census51e972 exits1: four keys present, revision_conflict absent
(RP-280). New native host diagnostic3d9906/3193fa,session65189 terminal exit2:
48 rejection cases fail absent inline row status, eight applied controls pass.
Both conflict HTTP200 and typed409 execute one held authoritative refresh with
disabled controls before failing the inline assertion. Actual runtime parsers,
native Worker/host/keyboard run over controlled Response/socket protocol input;
no live network/SQL/mint claim. Types1bcca0/78b614,session95572 terminal exit0,
zero errors/warnings. All handles now terminal before correction.

Correction scope under accepted R9: one optional act result/error observer with
closed typed inputs, called inside its existing task, so feedback is bound to the
submitted node rather than inferred from mutable global notice after settlement.
Reputation-only wrapper clears old feedback at submission; four existing mapped
rejections and revision conflict map to declared Reputation keys. Preserve shared
notice/effect and refresh ordering exactly, including HTTP409 handling; do not
add a conflict entry to SurfaceRejections. Render a stable polite row status
matching only that submitted node. Applied result/global status remains existing.
No authoritative state computation, schema/security/session/transport change.

Separate key scope: declare ONLY missing revision_conflict with standard explicit
PENDING OWNER COPY placeholder; all other prose unchanged. Root copy generation
updates required six copy outputs and content manifest. Generated Go All() list
changes server producer tree; retained dated career artifacts remain valid for
their recorded historical tree, NOT current-tip proof. Do not restamp or overwrite
reports, mint/ratify content, or assert harness/epoch acceptance. Kernel affecting
paths exclude these UI/copy bytes; no behavior-identical false numeric bump.

Require56 new native host cases plus58 existing child cases green, and shared
GS0.2 host rejection control. Demonstrate compiling omissions of row feedback
rendering, host outcome observation, conflict refresh and row ID matching; latter
must fail attribution without changing response behavior. Restore exact production
bytes after EACH terminal probe, rerun; no edits under any live handle. Cold
types/build/client/boundaries/copy, focused Go copy registry via root -count=1,
and separate kernel-history guard. No full native Firefox/CI/SQL/AC12/AC15 claim.
No checklist flips, authored prose adoption, archive, cleanup, publication or push.
Full range after b8ee639f needs designated cross-party verdict including records.

Presentation refinement before source: when a known Reputation rejection has
its inline polite status, suppress only the duplicate global status text while
that surface is mounted. Shared intentNotice/effect remains intact, unknown and
transport failures retain shared presentation, and applied global result remains.
Add an exact empty-global assertion for known rejection to avoid double speech.

## 2026-10-06 — R9 host correction initial green and probe extension

Root host+child native0ab5fd/c5b717,session70392 terminal exit0:114/114
(56 new host/58 child). Root shared GS0.2 rejection control15d693/9771ec,
session67142 terminal exit0:two selected passes/44 other cases visibly skipped.
Both root runs' automatic performance follow-on passes one selected Chromium
case/22 skips; not full Worker/perf/native acceptance. Types2b59f7/10bbc9,
session31584 terminal exit0, zero errors/warnings. New full run reached row
attribution/retry/new-Founder-revision checks; baseline failed before those.

Cold build/client/boundaries/topology/copy ff5f76/1b63b8/6803f1,
session12842 terminal exit0:8106 Node passes/188 native skips (91 pass/18 skip
files); boundaries/negative controls and658-key copy/manifest pass. Key explicitly
pending owner copy. Copy hash now a5df8920bc1573c8b0543b0c6a3c4c2de955b580626c8e1d130fc9e1adeb0e5e;
constants hash unchanged. Go copykeys649a17 exits0 but has NO test files (compile,
not test evidence); actual gen-content-manifest/releasepackage a718bb/aa5785,
session81201 terminal exit0, -count=1. Kernel6a04bc/68c495,session26630 exit2:
checkout/adversarial fixtures pass; historical50a3a514 vs0cf9f7a6 still RP-131 RED.
Read-only ps was sandbox denied; no escalation/termination/claim inferred.

All handles terminal. Extend predeclared compiling probes to also omit feedback
clearance and duplicate-global suppression, exercising those two new safeguards
explicitly rather than accepting green tests alone. Six probes total, individually
terminal then exact restoration. No assert/population/deadline/engine-policy edit.
The act option type formatting is mechanical, not a new behavior. No checkbox,
archival, retune, mint, owner-prose adoption, report restamp, CI green or push.

After the six terminal probes/restorations, predeclare one independent omission
of options.failed observation, isolating the typed409 connection from HTTP200
observation. Same56 native cases; no fixtures/assertions/bounds changed. Require
only409 inline failures with applied/ordinary rejection controls preserved.

All seven probes now terminal/restored. Test-teardown inspection finds that an
early assertion failure with a held POST can make cleanup's applied receipt start
a new refresh after the held-snapshot drain. Before editing, scope its correction
to resolving held POSTs with the same valid invalid-rejection body already used
by the diagnostic; no new production behavior or assertion/population change.
Rerun restored host+child and shared control, cold types/client. This is teardown
robustness, not a passed gameplay or receipt-acceptance criterion.

## 2026-10-06 — R9 host rejection correction / executed discrimination

RP-278 locally corrected under af744e65/defa9679/441013a5: act's optional
outcome/error observers bind feedback to the submitted node inside the existing
task. Four existing keys plus conflict presentation are inline on only that
row, with a stable polite status. Next submission clears previous feedback;
known inline text suppresses duplicate global speech only on this surface.
Applied global result/shared unknown/network handling are unchanged. Conflict
is NOT added to SurfaceRejections; HTTP200 and typed409 preserve the existing
authoritative refresh, held controls and next Founder revision. Server-derived
state remains authoritative, no optimistic purchase or transport/schema change.

Seven independently compiling source omissions, same56 host cases, all terminal
exit2 with assertion failures (not type/build failure):

| Omission | Run / terminal output | Failed / passed | Fired observation |
|---|---|---|---|
| Row text rendering | 54db28 / 1abdae, session29756 | 48 / 8 | Exact inline text empty |
| HTTP200 outcome observer | e119cc / f05f06, session17665 | 48 / 8 | Ordinary first rejection empty; typed409 first works but retry's invalid rejection empty |
| Conflict refresh in both shared branches | 9d4d16 / 851c87, session30270 | 16 / 40 | Snapshot calls remain1, required2 |
| Row ID equality | 0f1baa / b96dba, session15915 | 48 / 8 | Unrelated row incorrectly receives text |
| Previous feedback clearance | b3d6fe / 9d7970, session18366 | 48 / 8 | Old row still speaks during second submission |
| Duplicate-global suppression | 03b49e / 89795e, session44797 | 48 / 8 | Global status nonempty beside inline status |
| Typed409 error observer | 63c843 / ee437c, session74228 | 8 / 48 | Only HTTP409 inline text empty |

After EACH terminal probe, exact production restoration is verified by
b15afa/45a721/4ca52d/57a434/ace9d6/76f60f/543231. Both sources match441013a5:
GameUIApp SHA256 ea9609226388b0ca1abfb1db2020557684a9132a3b10e630a6a8a75fda0557a7;
ReputationTreeSurface SHA256 059bfaf66960be8347b0d6f8edd3c90c5a8d95d22b603779dd9b62552279743d.
No edits under live handles; no assertion/population/launch/security/deadline loosening.

Final restored host+child f2af8a/ac53d4,session61050 terminal exit0:114/114
(56 host/58 child). Teardown-only repair resolves held requests with existing
invalid rejection instead of opening a new applied refresh after drain; gameplay
assertions/population untouched. Final shared GS0.2 control042a69/67d999,
session46428 terminal exit0:two selected cases/44 skips. Automatic performance
follow-ons each pass one selected Chromium case/22 skips, NOT full Worker/perf
acceptance. Final cold types/client36ea00/dd0401,session4965 exit0:zero errors/
warnings,8106 unit passes/188 visible native skips,91 pass/18 skip files.
Earlier build/boundaries/topology/copy/Go-manifest executions remain unchanged-
production passes above; kernel-history RP-131 is explicitly RED, not waived.

RP-280 declaration corrected separately in441013a5: required key plus explicit
PENDING OWNER COPY placeholder, six generated-output computation and dependent
manifest. Only five generated files changed; code-reference output was already
byte-identical (registry is Go-only). Census0f158d confirms all five keys and
all prior source entries byte-unchanged. Copy658, hash a5df8920bc1573c8b0543b0c6a3c4c2de955b580626c8e1d130fc9e1adeb0e5e;
611 orphan warnings visible, not exhaustive mounted-client discovery. No owner
prose adoption. Generated Go All() changes server tree to4882868df62e81847347a791f84c474d62d6794b;
balance cd982b7c58a53cac0ab4c91a773705033930414d / constants hash unchanged.
Prior career artifacts remain historical evidence for their recorded producer,
not current-tree runs. Files untouched, no restamp or fresh H1/H2/H4/H5 claim.
Kernel161 unchanged, these paths outside kernel-affecting list e3e720.

Actual client host/runtime/parsers/native keyboard/Worker execute against
controlled HTTP Responses/socket protocol messages; NO live server/WebSocket/
Postgres/minted-tree or full AC12/AC15 claim. Firefox remains zero-execution
RP-256; capacity and all owner/author/statistical/review/full1.0 holds remain.
Next ground/predeclare RP-279 persistent Amount costs across every row state
and era. Full range after b8ee639f requires Claude's designated pass through
final tracking/self-record edge; earlier spans remain independent. No checkbox,
RFC acceptance/archive, cleanup, retune, mint, publication/deployment/push or
goal completion. Concrete host behavior and discriminating evidence progressed.

## 2026-10-06 — Self-first-filter fixture coordinate finding

Full code/test diff3ec1d9 inspection finds controlled rejection metadata always
uses current_revision7: that is wire-valid but incoherent for a first conflict
whose subsequent snapshot is8, and for the second rejected intent sent at8.
Before editing, bound repair to adding a fixture revision argument: conflict
receipt current_revision8, retry invalid receipt at its request's exact expected
revision. No assertion/source/effect/copy change. Earlier baseline/probe failures
remain evidence of their named UI defects, not server-coordinate validation.
Rerun host+child114, shared control, types and unitclient before final self-filter.

Coordinate-corrected native d5439f/94a261,session4183 terminal exit0:114/114;
shared040f72/8e79fe,session30505 exits0:two selected/44 skips. Both automatic
performance follow-ons one selected pass/22 skips. Cold types/client fa7ce2/
7baf90,session51086 exits0:zero errors/warnings,8106 passes/188 native skips,
91 passed/18 skipped files. Earlier probe metadata limitation remains disclosed;
assertions/source unchanged, no retrospective server-coordinate proof claimed.
Whole-root vet f4e10c/54b4cb,session99210 terminal exit0.

Read-only formatter check e6c70b lists server/copykeys/generated.go. Baseline
b8ee639f streamed through gofmt -d f2d80f also exits1 with All/CompanionKeys body
formatting differences. This is inherited generator-output debt RP-281; do not
format only generated output and break canonical generation, or quietly add a
Go executable dependency to client CI. Separately predeclare paired template/
output canonicalization and generation negatives in its tooling lane. This R9
range changes only All() membership, verified666cf4; not gofmt/all-CI green.
All handles terminal. Next accepted UI work remains RP-279 costs, with RP-281
explicitly queued rather than hidden. Cross-party/full1.0 holds remain.

## 2026-10-06 — R9 host rejection self-first-filter

Review by: Codex (implementer; self-first-filter, NOT designated).
Recorded by: Codex.
Reviewed range: b8ee639f..3b1e620b — all five commits/all18 paths, including
predeclaration, diagnostic/correction, generated artifacts, teardown/coordinate
repairs, docs and ledger/board/queue edges. Verdict: ready for designated review,
not RFC acceptance/archival approval or full CI/formatter-green verdict.

Source/test/docs/candidate full diff3ec1d9, tracking/log diffcfd90d, generated
small diff18b0ef and generated Go/types coordinate check666cf4 inspected.
Follow-up fixture/debt diffb80567 inspected; no new source/assertion or gate
relaxation. Full-range whitespace passes. Initial fixture revision finding is
corrected and rerun114; original failures/probes retain their actual fixture
version/provenance, not retroactively relabeled server-coordinate proofs.
Inherited generator formatter debtRP-281 is recorded, not hidden or unscopingly
repaired. All sources restored exactly, all test/probe/tool handles terminal.

The optional observers bind the particular node to its parsed result inside
act's task; no reading global mutable notice after completion. Shared effect/
refresh branches are byte-unchanged; no conflict entry removes resync. HTTP200/
409, both eras/native keys, held POST/refresh, row cardinality, retry/Founder
revision and applied controls execute. Seven compiling negatives discriminate
independently; final native114/sharedtwo and types/unit8106 pass. Runtime/parsers
and Worker are actual; HTTP/socket input is controlled, not real network/SQL/
mint/whole AC12. Native Node skips and selected performance are not full proof.
Unknown/network policy remains shared; only known inline double speech is hidden.

The accepted owner placeholder exception authorizes exactly one missing key,
not prose adoption; all previous entries unchanged. Copy658 and manifest match;
Go All() adds only that key. Generated-Go tree changes4882868d and older dated
career records stay historical, not restamped/current. Balance/constants/kernel161
and all runtime math/transport/security/schema/CI/report files stay untouched.
Historical RP-131 RED, Firefox/capacity/owner/author/mint/full1.0 and every prior
designated-review hold remain. No checklist flip or self-archive. Final self-record
edge must join Claude's full range afterb8ee639f; all earlier spans independent.
Next safe accepted work: ground/predeclare RP-279 all-state Amount costs, with
RP-281 tooling repair separately queued. No cleanup, publication/deployment,
push, goal completion or reduced nine-tier/platform1.0 promise.

## 2026-10-06 — RP-279 persistent row cost diagnostic predeclaration

Resume at22e03946 clean main. Previous goal turn is progress: RP-278/280 host
feedback/key correction, executed native/negative evidence and tracking commits.
AGENTS/process fully reread; active index still accepts Reputation. RFC/design
refs unchanged since020a25c6. R9 explicitly requires title/body/Amount cost/
requirements/state/control in that order. Existing renderer has cost only inside
available Buy copy, no Amount; Confirm removes that only cost. RP-281 formatter
debt is a separate tooling lane, not permission to broaden this UI range.

Ground actual Amount, canonicalString, Standard notation goldens and shared100ms
render scheduler. Follow the existing Desk's standalone Amount pattern: no
invented cost caption, reused Buy-as-label on owned nodes, new prose/key or
formatting rule. Owner copy remains pending. Numeric/Amount/scheduler themselves
must stay byte-unchanged; consume their published boundary rather than String(cost).

Add controlled component native diagnostic (NOT host/HTTP/SQL/minted content).
Two coherent arm populations: small costs1/2/3/5, and notation-boundary
999/1000/12345/999950 with independent literal expected output999/1.00 K/12.3 K/
1.00 M. Each includes owned/available/locked/unaffordable rows, valid prerequisite
order, adequate available budget for its available node and a truly unaffordable
last node. Diagnostic large costs are not candidate data/adoption. Both eras.

Static cases require exactly one Amount output per row in artifact order,
visible/non-hidden, correct independent literal text, before requirements/state/
controls; only available row has a Buy control. Lifecycle population adds native
Enter/Space × owned/available authoritative outcomes in both eras/populations.
Verify cost persists through Confirm, Escape, resubmission, held purchase/refresh,
authoritative replacement and task settlement; row/focus/state remain server-prop
derived and published formatting remains unmodified.20 cases/engine,40 new total.
Run unchanged-source baseline first; existing58 child/56 host are later controls.

No source/manual file edits while any test/probe/tool handle is live. Native
chromium/webkit only; unexecuted Firefox remains RP-256, not waived/skip-added.
Then separately predeclare correction and compiling negatives before source.
Root cold types/build/client/boundaries/copy and separate historical kernel guard;
no full CI/AC12/AC15/mint/SQL or assistive-user-study claim. Full new span after
22e03946 needs Claude including final record edges; all prior ranges independent.
No numeric/kernel bump, balance/report/copy/transport/CI/schema/auth change,
checklist flip, archive, cleanup, publication/deployment/push or goal completion.

## 2026-10-06 — RP-279 baseline / bounded renderer correction predeclaration

New diagnostic b7e064/cddb53,session36994 terminal exit2:all40 cases fail
missing first row Amount output. This baseline confirms missing rendering,
not yet lifecycle/order/notation acceptance (assertion stops before those).
Types e64a2e/3bddeb,session51863 terminal exit0, zero errors/warnings.
All handles terminal before correction. Remove one unused test type import only.

Correction scope: import existing Amount and canonicalString; mount one Amount
immediately after body and before requirements/state, outside all available/
confirm/pending conditions. Its value comes from that exact server-prop node.cost
converted through the existing canonical boundary. Existing Buy text unchanged;
no new cost caption/prose/key, tooltip, pricing, state, eligibility, task/focus,
host/runtime, numeric formatting or scheduler contract. Canonical UI docs updated
in same source change. No core/kernel change; renderer outside watched paths.

Require40 cost +58 old child +56 host native cases green. Five independently
compiling negatives before closing local claim: omit cost; render available-only;
hide while Confirm/pending; bind another row's cost; duplicate the amount.
Additionally move cost after state to demonstrate reading-order discrimination.
Six total. Hold source bytes fixed under every live handle; after each terminal
result restore exact bytes and verify SHA, then final full restored population.
The small/notation literal oracles are independent of formatter implementation.
Visibility/cardinality/lifecycle are bounded native DOM evidence, not human AT
or whole AC12/SQL/mint proof. Cold root types/build/client/boundaries/copy,
separate RP-131 history; no Firefox/CI green, report restamp/epoch adoption,
checklist flip, archive, cleanup, deployment/publication/push or goal completion.
Whole new span after22e03946 needs Claude; prior ranges remain independent.

## 2026-10-06 — RP-279 initial restored green / visibility probe refinement

Initial corrected native c94f0e/7bfbcf,session74638 terminal exit0:154/154
(40 cost/58 prior child/56 host). Native lifecycle now reaches every phase;
notation literals/order/geometry all pass. Root automatic performance follow-on
one selected Chromium pass/22 skips, not full Worker/perf acceptance.
Types1645fc/0b94a9,session86942 exits0, zero errors/warnings.
All handles terminal. Before probes add one hidden-wrapper omission/control
to explicitly discriminate visible cost, not merely DOM presence: seven probes
total. No fixture/assertion/bound/engine/security changes. Component-only new
cost evidence; prior host controls still controlled-network, not SQL/mint/AC12.

## 2026-10-06 — RP-279 executed negatives / final restored gates

Seven independent compiling mutations on the same40 cost cases, both engines:

| Probe | Terminal execution | Actual outcome |
|---|---|---|
| Omit standalone Amount | 3b17e4/ca89df,session43696 exit2 | 40 fail, row0 Amount cardinality0 |
| Available-only Amount | b7d01b/df3ba1,session54035 exit2 | 40 fail, owned row0 cardinality0 |
| Hide during Confirm/pending | 61cf2f/81c52f,session3082 exit2 | 32 lifecycle fail at Confirm,8 static pass |
| Bind row0 cost on every row | 336452/430b4d,session69760 exit2 | 40 fail wrong row1 literal:1 vs2 or999 vs1.00 K |
| Duplicate Amount | 241f99/87605b,session80947 exit2 | 40 fail cardinality2 |
| Move Amount after state | 204137/04375b,session62034 exit2 | 40 fail document reading order |
| Hidden wrapper | 522167/0902fa,session91142 exit2 | 40 fail hidden ancestor, despite DOM amount presence |

Every case exits by assertion rather than compile error. No test fixtures,
assertions, deadlines, browser flags or criteria change between probes. Each
terminal result is followed by explicit source restoration and SHA check before
the next mutation. The first restoration initially used wrong indentation
(SHA3d16e43b); diff241fc9 corrected it immediately before any new test/probe.
Every subsequent pre-probe/final check matches committed renderer exactly:
1b8917a122bc275ad55ffde751e44ba1835a9848d443190dc30d2a5363cd1ef7.

Final restored native0c3094/970bb1/2d9847,session14751 terminal exit0:
154/154 across six engine/files,40 new cost/58 child/56 host. Root automatic
performance follow-on one selected Chromium pass/22 skips, not complete perf.
Root cold type/build/unit/boundary/topology/no-payment/copy/manifest command
9b3ead/d9a6ac/31c89f/b657da/770a4b,session57097 terminal exit0. Types zero errors/
warnings,213 modules build,8106 unit passes/208 skips,91 files pass/19 skip.
Browser-only Node skips are visible, not native evidence. Shell boundaries,
13 topology/22 cosmetic/six payment negatives pass. Copy658/hash unchanged,
611 orphan warnings visible; Go-only discovery remains partial. Manifest passes.
Separate kernel3a64d5/ad9cf0/803075,session84161 terminal exit2 at the unchanged
RP-131 commit50a3a514 vs0cf9f7a6; history-checkout contract and fixtures pass.
Not full verify-client/CI green. A read-only process listing28ef81 was sandbox
denied; no escalation/termination attempted. All handles terminal before records.

Correction is exactly three renderer lines plus same-range canonical UI docs,
new diagnostic and tracking. Numeric/Amount/scheduler/copy/catalog/hash/Go/
kernel161/live math/balance/schema/auth/transport/CI/corpus/reports unchanged.
No new authored labels or reinterpretation of Buy text. R9's four states,
notation and held tasks execute on coherent synthetic props; not a host/SQL/
minted player/whole AC12 or manual assistive proof. Full span after22e03946,
including these record edges, requires Claude; all earlier ranges independent.
No checkbox flipped, self-archive, report restamp, owner adoption, cleanup,
publication/deployment/push, goal completion or reduced full1.0 objective.

Remaining R9 grounding while gates ran (read-only): e55c08 is exact R9.
2cf2cd/6c93a4/94ddd1/496db3 show child selected starts empty, parent exitPlan
persists until continueRun; navigation/remount can disagree. File RP-282 as a
source finding, not executed failure or invented persistence policy. 496db3/
4d607d show Wind Down previewDelta0 versus eligibility-only transition DTO,
while Offer has authoritative preview; RP-283 requires producer/contract
grounding, not client math. 142c97/496db3/94ddd1 show RunEnd receives only
ended and renders payout without examined available/next-route consumer;
RP-284 needs event/snapshot ownership grounding. Neither new source finding
expands this cost correction or authorizes a wire/payout-policy change.
Next safe accepted lane: bounded native actual-host RP-282 reproduction,
after this checkpoint/self-first-filter; RP-281 tooling stays separate.

## 2026-10-06 — RP-279 range self-first-filter (not designated review)

Review by: Codex (implementer/self-first-filter).
Recorded by: Codex.
Inspected range:22e03946..d1a247a6 (four commits). This record edge must join
the later Claude range; no designated verdict or archival eligibility claimed.

Full source/docs diffa8513e:exactly two imports/one persistent Amount mount,
outside row controls. New diagnostic fully read39ba6b:independent literal cost
oracles, visible/cardinality/order assertions, both eras/native keys, all four
states and held task/authoritative replacement; actual Amount/formatter used,
controlled props not host/SQL/mint. Tests were not weakened to fit source.
Seven compiling probes independently discriminate, restore exact SHA; final154
native and8106 cold unit/type/build/boundary/copy/manifest pass with explicit
skips. RP-131 guard RED/Firefox zero execution remain, not whole CI acceptance.
Full records8beaf3/tail073014 inspected, whitespace passes4b8b73/this range.
Scope10 paths; canonical docs/tracker/ledger/log agree. Initial indentation-only
restore slip disclosed/corrected before next execution; no residual probe.

No numeric/Amount/scheduler/copy/kernel161/CI/server/balance/report/policy
bytes moved. RP-282/283/284 are explicitly source findings with separate routes,
not executed claims or scope extensions. No inherited B7/full AC12/1.0 closure,
checkbox flip, self-archive, owner adoption, mint, cleanup, deployment/push or
goal completion. Next bounded actual-host plan reproduction is safe accepted
work while every independent prior designated-review obligation remains live.

## 2026-10-06 — RP-282 actual-host plan consistency diagnostic predeclaration

Start eb258a7a clean main, prior cost range closed locally only. Accepted R6/R9
requires advisory selected plan, default empty/unchanged one-action Exit, and
exact server re-validation. Ground2cf2cd/6c93a4/94ddd1/496db3/4176bb/4bebf6:
panel local selected starts empty; parent remembers exitPlan through navigation
and Offer replacement; withPlan forwards it. Reuse existing controlled actual
runtime/parser/host fixture, not test-only host setters or direct gameplay API.
This is a consistency diagnostic, not new persistence semantics/owner ruling.

Extend only existing host diagnostic helper with opt-in eligible Wind Down
(default false preserves old population), and controlled valid company Offer
publication at revision14 after initial13/Founder7. Offer uses collapsed type
with preview delta2, eligible Desk has current available4, nodes costs1/2.
Fixture wire remains declared v4; no new preview field/schema/parser/source.
Existing actual native Worker and controlled HTTP/socket boundary execute.

Population:both eras × native Enter/Space × five flows,20 per engine/40 total:
untouched Desk default empty; selected Desk forwards both nodes in artifact
order; Desk→Settings→Desk remount resets visible empty selection; selected Desk
replaced by authoritative Offer resets visible empty panel; selected Offer
forwards both visible selections using preview delta2. Select second then first
to independently check ordered serialization, project4→1 or6→3. Default flows
must omit the reputation_plan field, not send[]; selected flows match actual
checked rows exactly. All gameplay intents originate native controls and one
held POST, unique UUID, Company expected_revision13/14 and Founder7. Reject
known invalid with coherent Company revision; no need to change/save/mint game.
Assert one initial snapshot/no unexpected requests, transport readiness and
visible collapsed/open disclosure. No broad AC12/server/SQL/default-player claim.

Run unchanged production first. Classify baseline as fixture invalid if event
parser/resync or prior-step control fails; only a proper enabled native action
reaching a wrong captured payload proves hidden spend. Existing empty-remount
UI is the bounded reference behavior, not permission to invent persistence.
If confirmed, separately predeclare minimum synchronization and compiling
negative/control probes before product edits. No RP-283 preview math/wire,
RP-284 event consumer, RP-281 generator, copy/prices/kernel161/CI/security/
schema/balance/reports, acceptance box/mint/owner adoption/archive/push change.
All handles terminal before edits. Full new span aftereb258a7a needs Claude,
independently of prior ranges. Full nine-tier/platform1.0 goal remains active.

## 2026-10-06 — RP-282 first diagnostic / Offer fixture refinement

0e6f64/4ddf9a,session23344 terminal exit2:24 fail/16 pass/56 old cases
selector-skipped. Eight Desk remounts genuinely reach the outgoing intent and
fail hidden reputation_plan while checkboxes are empty. Sixteen Offer cases
instead stop at expected_revision14 vs actual13:the injected live event advances
runtime's stream cursor, not the host's snapshot coordinate. That expectation
was unsupported without an authoritative refreshed snapshot. Do not call those
plan failures or a product revision defect. Types966845/ea0d1c exits0.

Refine controlled Offer population before further measurement:deliver the
normal trailing receipt, let actual runtime/host request its snapshot, then
deliver declared v4 revision14/Founder7 with unchanged reputation arm and
eligible Wind Down. Existing receipt→refresh behavior4176bb/94ddd1 remains
source-unchanged. Assert two snapshots only in Offer populations, otherwise one.
This completes the controlled protocol population instead of relaxing revision
or plan assertions; Company14/Founder7 then becomes coherent. All40 cases stay,
same native keys/eras/flows/plan oracles. No gameplay API shortcut/source edits.
Baseline must be rerun before choosing any plan correction. All handles terminal.

## 2026-10-06 — RP-282 confirmed hidden plan / synchronization predeclaration

Refined unchanged-production d07a7c/224459,session11003 terminal exit2:
16 remount/Offer-replacement cases fail hidden spending;24 empty/selected
controls pass;56 old selected-out cases remain skips, not failures/green proof.
Both eras/native keys deliver actual wrong POST after visible reset, now with
coherent Company13/14 and Founder7. All handles terminal. This is controlled
network/native host evidence, not real server/SQL/mint execution or acceptance.

Minimum correction consumes the existing child's current selected state on
mount through its existing onChange callback. Child selection starts empty and
still reports all user changes; parent must not retain an invisible previous
mount's plan. No persistence policy change, selection caching, payout math,
feature/event/parser/schema/transport code or copy/pricing change. Add Svelte
onMount plumbing in ReputationPlanPanel only; GameUIApp and withPlan unchanged.
Canonical docs updated in same source change. Also fix diagnostic-only cleanup
to use the captured request's actual expected_revision, not Founder7 for every
Company request. It changes no assertion/population or gameplay source.

Require40 plan/56 old host/58 child/40 cost native cases green =194 total.
Before local closeout demonstrate two compiling mutations on same40 plan cases:
omit mount notification (expected16 fail with24 control pass), and notify a
nonempty first-node plan on otherwise empty mount (expected24 empty-flow fail,
16 explicit-selection controls pass). Restore exact source SHA after terminal
each, then full final194. Cold root types/build/unit/boundaries/copy/manifest,
separate inherited RP-131 guard remain; no full CI/Firefox/AC12/SQL/mint claim.
No numeric/kernel161/balance/reports/copy/CI-policy change, owner/author ruling,
checkbox flip, archive, cleanup, push/deployment or goal completion. Whole new
span aftereb258a7a needs Claude, incl fixture refinement and final record edges.

## 2026-10-06 — RP-282 initial corrected native green

3330f9/1776e1,session4043 terminal exit0:194/194 (40 plan/56 old host/
58 child/40 cost). Automatic performance follow-on one selected Chromium
pass/22 skips, not full native/performance. Types7d9bce/ca9cf8,session25489
terminal exit0, zero errors/warnings. Four-line mount plumbing plus canonical
docs inspected0d039d; callback consumes current child selected state, user
toggle serialization/eligibility/preview computation unchanged. All handles
terminal before probes. No GameUIApp/runtime/parser/server/copy/prices/kernel
bytes change, no invisible persistence semantics or promotion to full AC12.

## 2026-10-06 — RP-282 executed negatives / final restored gates

Same40 plan cases (both eras/native keys/five flows), assertions fixed:

| Probe | Terminal execution | Actual outcome |
|---|---|---|
| Drop mount callback | 81b253/940577,session7806 exit2 | 16 remount/Offer empty-plan failures,24 controls pass |
| Notify first-node selection on empty mount | d46444/59168f,session77973 exit2 | 24 empty-flow hidden-plan failures,16 explicit-selection controls pass |

Both compile and reach the wrong outgoing POST; no compiler exception as
discrimination. Old56 host cases are selector-skipped during probes only.
Source restored after each terminal to exact committed plan SHA
251343ff0583a62a2b3ca5faa5332f8d80c8de6a3c8e7b3e161f52c1ed7de861
(54f97c/88d697/5eeb9d). No criteria/fixture/flags changes between probes.
Final456872/5076e9,session28947 terminal exit0:194 selected native passes,
40 plan/56 old host/58 child/40 cost. Root automatic perf one selected
Chromium pass/22 skips, not full performance. Cold0c4c4a/bfdd5e/0f9ed3,
session10449 terminal exit0:zero type errors/warnings,213 build modules,
8106 unit passes/228 skips,91 files pass/19 skip. Boundaries/13 topology/
22 cosmetic/six payment negatives pass; copy658/hash unchanged,611 orphan
warnings; Go manifest passes. Node skips are not native execution evidence.
Separated8f536/e2cd53,session23827 terminal exit2 at the unchanged RP-131
50a3a514/0cf9f7a6; checkout contract/history fixtures pass. Firefox unexecuted,
not whole client/CI green. All handles terminal before tracking edits.

Correction changes only plan mount synchronization, canonical UI docs, the
existing host diagnostic and tracking. Existing helper defaults preserve old
host population; cleanup now uses captured scope revision. No host/runtime/
event decoder/transport/server/SQL/schema/kernel161/numeric/price/copy/balance/
report/CI policy bytes moved. Full range aftereb258a7a requires Claude including
fixture refinement/record edges, independently of all earlier ranges. New
native evidence is actual runtime/host/Worker over controlled network inputs,
not server/SQL/mint/full AC12 or human assistive proof. No new persistence
policy:an existing empty panel cannot carry another mount's invisible selection.
No checkbox flip, self-archive, owner adoption, mint, cleanup, report restamp,
publication/deployment/push, goal completion or narrower1.0 objective.

Remaining source grounding (read-only during gates): RP-2832df91c/2e4f65/
9e67ef/a0138e confirms projector, registered API schema and exact client parser
all expose eligible only. DESIGN-GAP: R9 demands payout-aware Wind Down plan
but does not define the authoritative preview bridge. Proposed draft/amendment
topic:Reputation UI evidence bridge, covering server-authored payout preview,
closed versioned API ownership and producer→consumer/refusal proof. Author must
reconcile before implementation; do not use client math or silently add fields.
RP-284595c22/abfee9/6da5b4 confirms run_started v2 already produces factor/
applied starter IDs, but examined UI event union ignores it and RunEnd lacks
available-balance input. Producer/consumer ownership still needs grounding,
not inferred balances from payout or eager next-run lifecycle suppression.
RP-2859ded02/e55c08 identifies header tabindex=-1 despite R9 sequential Tab
requirement and no host heading-focus path. Source finding only, needs native
unchanged-source diagnostic before correction. Next safe accepted lane is that
bounded header/row path; RP-281 generator repair remains separately available.

## 2026-10-06 — RP-282 first-filter population limitation

Read-only producer census153165 finds server/gameui/transition_preview.go
offers Wind Down only at tier>=1. The helper's opt-in eligible Tier0 projection
is deliberately manufactured to exercise both copy eras through the host;
it is wire/accounting/coordinate-valid, not a reachable production terminal
population. Tier1 independently reproduces the same remount/Offer hidden-plan
failure (eight of baseline16), while Tier0 supplies additional renderer behavior
only. No producer/default-workflow/eligibility claim may cite the full194.
Existing default-false helper/old host population is unchanged. Add a fixture
comment and current-board disclosure; no fixture/assertion/source change or
rerun relabeling. This limitation does not negate the Tier1 defect or permit
shipping manufactured eligibility. Real SQL/mint/default-player AC12 remains
held independently; no policy change to make Tier0 eligible.

## 2026-10-06 — RP-282 range self-first-filter (not designated review)

Review by: Codex (implementer/self-first-filter).
Recorded by: Codex.
Inspected range:eb258a7a..0d8029e4 (four commits). This record edge must join
the later Claude range; no designated approval or archival eligibility claimed.

Full source/docs083e73:four plan lifecycle lines and canonical empty-panel
description, no GameUIApp/runtime/parser/server or payout/selection-policy
rewrite. Complete test diffaeeb0e inspected:helper defaults preserve old
population, controlled v2 Offer→receipt→v4 snapshot coordinates are explicit,
both eras/native keys/five flows capture exactly one DOM-originated POST.
Hidden-spend field absence and explicit selected artifact order discriminate;
24 positive controls prevent an always-empty wire from passing the full lane.
Native baseline16/two compiling probes16/24/full restored194 are executed;
source restoration SHA exact. Cold unit8106/type/build/boundary/copy/manifest
pass; Node228 skips, selected perf22 skips and inherited RP-131 RED/Firefox
unexecuted are visible. No whole CI/AC12 claim. Every handle terminal before
record edits; whitespace6c2090/bc0e6b passes, ten-path scope4f5126.

Records7d9e26/63fbd6 and append-only log inspected. First-filter caught the
manufactured Tier0 eligible projection, confirmed by full producer read899fcf,
and disclosed it in fixture/comment/current boards plus the log. It exercises
copy-era host rendering only, not reachable production Wind Down; the Tier1
subset independently reproduces the defect. Initial Offer coordinate omission
is disclosed and corrected by completing the protocol, not loosening assertions.
No true backend eligibility/default-player/mint/SQL coverage claimed.

RP-283 preview DESIGN-GAP/author route, RP-284 unexamined next-route/available
bridge and RP-285 unexecuted Tab source finding are not quietly implemented or
promoted. Existing owner/author/data/mint/capacity/deployment/all-range reviews
remain; no checkbox flip/self-archive/push/cleanup/report restamp/owner copy or
goal completion. Kernel161/price/copy/balance/CI unchanged. Next safe accepted
work:predeclared native R9 header/row Tab diagnosis RP-285. Proper full nine-tier/
platform1.0 goal is active, not complete/paused/blocked or reduced to a preview.

## 2026-10-06 — RP-285 sequential Tab diagnostic predeclaration

Resume8a3bfb70 clean main, no inherited live tool/test handles. Previous goal
turn is progress:RP-279/RP-282 runtime corrections, executed negative/native
evidence and synchronized records. AGENTS/process fully reread1fc519/3c942e;
active indexf269dd still accepts Reputation. Exact RFC/design refs unchanged
since020a25c6 (6bbc50; use actual02-economy-balancing/11-ux-writing filenames).
R9 fully re-grounded2dab8a, componentc34f8b/old native tests7db2cf read fully.
Kernel watched pathsb2431c exclude Game UI renderer; no numeric/kernel bump.

R9 explicitly requires Tab to reach header, then each row's single control in
artifact order. Current h1 tabindex=-1 and no host heading-focus path are source
findings only (RP-285), not native proof. Extend existing component diagnostic,
not host fixture setters/SQL/minted/default-player or human AT evidence.

Population:two eras × three coherent arms × three control states =18 cases per
engine/36 total. Arms:mixed four states/one available control; adequate budget
with two available rows; no available rows with zero budget. Use exact declared
arm shape (do not propagate old fixture-only tree_active extra field). Control
states:ready,pending,unrecovered controls false; these are controlled props,
not actual transport-state producers. Costs/level/spend/bonus and prerequisite
states remain coherent; diagnostic values do not adopt balance content.

Mount actual component/theme between two test-only sentinel buttons. Only the
preceding sentinel gets programmatic focus. Native Tab must visit header→each
enabled Buy in artifact order→following sentinel; native Shift+Tab must reverse
exactly, including disabled/no-buy cases. Never focus the header or row controls
directly to hide the defect. Assert four-state/title census, no row Tab stops,
no accidental purchase/confirm. Independent expected order comes from input,
not discovered current browser focus order. Existing58 child/96 host/40 cost
cases are later regression controls. No new player prose/labels or mechanics.

Run unchanged production baseline first; if the setup/browser key fails, do not
call it a product failure. If confirmed, separately predeclare minimum heading
Tab-stop correction and compiling probes before source changes. Final target
230 Chromium/WebKit cases; Firefox remains RP-256 unexecuted, not waived.
Root cold types/build/client/boundaries/copy/manifest and separate RP-131 history
guard, all existing skips/limits visible. No file edit while any handle live;
terminal followed by exact SHA restoration before next mutation. Whole new
span after8a3bfb70 needs Claude, every earlier range independent. No full AC12/
AT/SQL/mint/1.0 closure, checkbox, owner-copy/price/balance/CI-policy change,
archive, cleanup, deployment/publication/push or goal completion. RP-281 tooling,
RP-283 author bridge and RP-284 producer/consumer remain separate lanes.

## 2026-10-06 — RP-285 first baseline / native harness control refinement

e72c63/3c3f5a,session80438 terminal exit2:36 fail/58 prior cases selector-
skipped. Chromium18 reach Buy or trailing sentinel instead of heading, a direct
sequential omission. WebKit18 instead reach body; this is not yet isolated
heading evidence because implicit sentinel/button Tab reachability was assumed.
Typescffd0d/9e7aee,session15257 terminal exit0,zero errors/warnings. All handles
terminal; no production correction yet.

Before relying on WebKit baseline, add a standalone native sentinel control:
two explicitly tabindex0 sentinel buttons, Tab forward and Shift+Tab reverse,
no component. Make component sentinels explicit0 too, without focusing heading/
rows or changing any product/browser flags/criteria. Original36 cases remain,
two engine control cases add38 total; final regression target232. If the control
fails, that engine's instrumentation is invalid, not product acceptance/failure.
If it passes, rerun unchanged production to separate heading/implicit-control
focus from the browser boundary. Also simplify unnecessary empty concat/type
cast syntax without changing oracle or input. No hidden workaround/waiver.

## 2026-10-06 — RP-285 controlled baseline / heading-only observation

b3039e/0b4b14,session27352 terminal exit2:36 heading failures/two standalone
native sentinel controls pass/58 old cases selector-skipped. Both engines now
reach an explicit trailing sentinel rather than header; native Tab boundary
control discriminates setup from product. Chromium ready cases reach Buy;
WebKit skips implicit buttons. No production change yet, all handles terminal.

Predeclare temporary heading-only observation, not a committed finished fix:
change only h1 tabindex -1→0, run same38 targeted native cases to determine
whether enabled row controls are reached next. Do not change flags/security/
browser preferences. If row focus fails, record a distinct accepted R9 defect
and extend diagnosis before a product correction. After terminal, restore exact
original renderer SHA1b8917a122bc275ad55ffde751e44ba1835a9848d443190dc30d2a5363cd1ef7.
No test edit/record while the observation handle lives. No acceptance promotion.

## 2026-10-06 — RP-286 confirmation diagnostic / correction predeclaration

Heading-only observation abdfb9/d2ac5c session75555 terminal exit2:4 WebKit
ready mixed/two-buyable row failures,34 controls pass,58 older selector skips.
Header now reached; WebKit goes to explicit trailing sentinel instead of Buy.
Both standalone native sentinel controls pass. Original source restored exactly
SHA1b8917a122bc275ad55ffde751e44ba1835a9848d443190dc30d2a5363cd1ef7
(c470c6). Record RP-286 separately from RP-285. AGENTS/process and R9 reread
1424ba/173be5/295935. No live handles during edits.

Extend diagnosis with both eras × mixed/two-buyable × Enter/Space =8 cases per
engine/16 total. Direct Buy focus is branch setup only, not sequential-header
proof: native key opens Confirm; Tab reaches Cancel then next Buy or sentinel;
Shift+Tab reverses Cancel→Confirm; Escape returns Buy without submission;
reopen then native Tab/key activates Cancel and returns Buy. No pointer or
synthetic DOM key dispatch. Existing sequential tests never directly focus
heading/rows. Combined new54 baseline first on restored original production;
old194 controls retained. Final selected Chromium/WebKit target248.

If baseline confirms, minimum correction under accepted R9: header tabindex0,
Buy/Confirm/Cancel explicit tabindex0. Disabled controls remain skipped; keyed
rows remain tabindex-1 for programmatic post-receipt focus. Document header's
intentional noninteractive Tab stop mandated by R9 with a narrowly scoped
Svelte accessibility annotation, not an altered role or blanket lint suppression.
No formula/state/transport/copy/price/schema/balance/numeric/kernel/CI change.

Predeclared compiling probes after corrected native pass: (1) header back to-1
must fail sequential tests, (2) rows tabindex0 must fail no-extra-stop census,
(3) remove explicit tabindex from three buttons must fail WebKit row/confirmation
order with Chromium/sentinel controls retained. Run new54 each; restore exact
corrected SHA after every terminal result before next edit. Finally all248
plus root types/build/client/boundaries/copy/manifest and inherited history guard.
No edits with live test handles. Negative results remain visible; no acceptance
bound/flag/preference/skip relaxation. This is controlled native component
evidence, not host/SQL/mint/default-player/manual AT or whole AC12. Whole range
after8a3bfb70 still needs Claude; earlier ranges remain independent. No checkbox
flip/archive/owner copy/cleanup/push or reduction of full1.0 goal.

## 2026-10-06 — RP-285/RP-286 combined baseline

a82043/341a5c session88542 terminal exit2:44 fail/10 pass/58 old selector skips.
Original heading fails all36 sequential cases. WebKit's8 confirmation cases
skip Cancel; Chromium's8 pass; standalone sentinel controls2 pass. Actual native
keys isolate the two omissions with controls green. Source remains original
SHA1b8917a122bc275ad55ffde751e44ba1835a9848d443190dc30d2a5363cd1ef7.
No live handle. Diagnostic commit precedes minimum predeclared correction.

## 2026-10-06 — RP-285/RP-286 native correction and discrimination

Minimum correction: heading and Buy/Confirm/Cancel explicit tabindex0; keyed
rows remain-1. One heading-only Svelte annotation documents accepted R9's
intentional noninteractive focus stop, with no fabricated role/lint policy.
Docs describe actual traversal in the same behavior change. No callback,
purchase/error/pending/state/formula/copy/price/numeric/browser/CI change.

084ac5/9ad53a session94349 terminal0:54 native cases pass,58 older selector
skips; selected perf1 pass/22 skips. Corrected source SHA27703516b76799c60c4f166f8a55822dc855b741f663e8c9ddcc8f1d3aa1834d.
Three compiling probes, each restored that exact SHA before next mutation:
- Heading back to-1:52187d/fab091 session42486 terminal2,36 sequential fail,
  18 confirmation/sentinel controls pass.
- Rows made0:741e1a/9e45ee session33967 terminal2,52 fail/2 standalone controls
  pass. The row census catches all36; native confirmation also hits the extra
  locked-row stop in16. This is stronger discrimination, not a changed criterion.
- Explicit button stops removed:f4bd32/38b733 session98403 terminal2,12 WebKit
  row/confirmation fail,42 controls pass (all Chromium and remaining WebKit).
Restoration16b767/c779ae/028b22 exact each. No edit while any test handle lived.

Final restored full population2a7d9f/ce10ff session57972 terminal0:248 pass,
zero selected skips (112 child/96 host/40 cost),96 module requests finish with
no pending/truncation. Follow-on selected perf1 pass/22 skips. Actual native
component/host over controlled inputs, not live SQL/mint/default-player or
human AT; Firefox RP-256 remains unexecuted, no browser flags/preferences changed.

Cold root83d28f/8e9833/3cd326/3997b1 session65627 terminal0:types0 errors/
warnings;build213 modules;units8106 pass/255 Node skips,91 files pass/19 skip.
Shell boundary passes;CI topology13/cosmetic22/no-payment6 negative controls
reject;copy658/hash a5df8920bc1573c8b0543b0c6a3c4c2de955b580626c8e1d130fc9e1adeb0e5e
unchanged,611 orphan warnings;deployment manifest passes. Server copykeys5691ee
compiles cold but has NO test files, not runtime producer proof.
Separate bfe2ba/a705f6 session27515 terminal2:historical RP-131 still fails
50a3a514 against0cf9f7a6;CI history checkout/fixtures pass before it. Not whole
verify/client/CI green. Kernel161/old dated career producer unchanged. All
handles terminal before record/commit. Whole span after8a3bfb70 requires Claude,
independently of earlier ranges. No checkbox/full AC12/RFC/archive/push/cleanup/
owner copy/mint/1.0 promotion. Next safe lane is RP-281 paired tooling diagnosis
or RP-284 producer/consumer grounding; RP-283 remains author contract gap.

## 2026-10-06 — Native Tab range self-first-filter

**Review by:** Codex (implementer self-first-filter, NOT designated review).
**Recorded by:** Codex.
**Reviewed range:** `8a3bfb70..5d9211da`, all six commits/ten paths including
predeclarations, diagnostic refinements, production+docs and tracker edge.
**Verdict:** first-filter passes; ready for Claude's designated full-range pass,
including this record edge. No acceptance/archive/independent approval.

Full source/test/docs/tracker diff442d8c and log diff inspected; whitespace
passes. Runtime change is four focus attributes plus one contract-specific
annotation; no role deception, focus trap, global lint suppression, state/math/
copy/browser policy change. Disabled buttons still excluded and keyed rows
stay programmatic-only. Native54 diagnosis/248 regression/three compiling
severings run, not merely read; exact restoration held. Standalone control
prevented attributing initial WebKit setup failure to the product. Row-stop
probe also fails confirmation16; actual52 failures disclosed rather than
pretending only the anticipated36 fired. Types/build/unit/boundary/copy gates
pass at unchanged copy658; inherited RP-131/Firefox/Node skips explicit.
No test assertions removed/checkbox flip/AT/default-player/SQL/mint/full AC12
claim. Prior independent ranges remain open. Goal active; RP-281 tooling and
RP-284 producer grounding are safe next work; RP-283 needs author contract.

## 2026-10-06 — RP-281 generator formatting predeclaration

Resume3df3ff32 clean main/no inherited live handles (d54f7c/3c5aa1). Previous
goal turn is progress: RP-285/286 source/test/docs/tracker committed, full native
248 and fired probes; independent review still pending. AGENTS/process fully
read3281af/db4106; active indexcc1c38. Existing generator/function/verifier
f4a042/2fe4ca/16790b, canonical docs4d40cf and archived CP1-C10/frozen amendment
f4cbcf read. This is existing-tool formatting repair under AGENTS gofmt law,
not new copy behavior or a rewrite of archived authority. Kernel registry01088f
does not watch copykeys/template; kernel161 remains unchanged.

Baseline2bb330 lists generated.go; e622c4 gofmt-d exits1:only All/CompanionKeys
one-line bodies split by formatter. Existing server producer tree787441 remains
4882868df62e81847347a791f84c474d62d6794b until regenerated file changes.
No hand-format-only repair. Extend existing copy verifier with a standalone
Node-only generation fixture import: exact independent whole-output goldens for
empty, single/default, ordered multi, independent all/companion sets, and
companion-only input. Keep collision refusal. Six corrupted golden-output
controls must reject (one-line body, indent, order, omission, companion binding,
constant identity). Run unchanged generator first; expected formatting failure,
not weakened golden. Then minimum multiline function template and avoid extra
blank line for zero constants; regenerate through root make copy-generate.

Acceptance: generated Go equals gofmt(old committed Go) byte-for-byte, all
other generated outputs and copy658/hash/manifest unchanged; fixture goldens
pass, current Go gofmt-d emits zero; Go cold core tests/vet, root types/build/
units/boundaries/copy plus inherited history guard recorded separately. No new
Go executable dependency in Node generation/verifier/client CI, no workflow/
Make lane/timeout/skip changes. Fixture import extends existing generation
verification, not content or CI topology authority.

Three compiling template severings predeclared: All body back to one-line,
Companion body back to one-line, or Companion binds allKeys. Each must fail
standalone fixture before history walk; restore exact template SHA each after
terminal result, never edit while any handle live. Also test current generated
drift rejects a temporary output-only corruption before restoring exact file.
Finish regenerated output+docs in same source change and reconcile all trackers.
Changing Go bytes makes current producer tree different even though behavior is
formatter-identical: dated career artifacts remain their original producers,
not restamped or claimed fresh. Full new span after3df3ff32 needs Claude;
all earlier ranges remain independent. No checkbox/full AC12/AT/SQL/mint/owner
copy/epoch/RFC/archive/push/cleanup/balance/formula/price/CI/kernel promotion.
Full proper nine-tier/platform1.0 goal remains active.

## 2026-10-06 — RP-281 unchanged-generator diagnostic

1fa064 terminal1: independent empty whole-output golden fails on both one-line
function bodies and the extra blank before All. The baseline is fail-fast;
remaining four goldens/collision/corruptions do not execute yet, not claimed
passed. No Go dependency, generator/output correction or live handle. Commit
fixture and existing-verifier import before changing the template.

## 2026-10-06 — RP-281 paired correction and executed proof

Template emits multiline All/CompanionKeys directly and one blank after package
when there are no constants. Root9992e1 generation writes all declared outputs;
only generated.go changes. Five goldens/collision/six corruption controls pass
2d2b33/2ee1ae. Local independent formatter check5da1fe proves actual generated
file equals gofmt(3df3ff32's committed file) byte-for-byte and all five synthetic
populations are gofmt-stable. Current gofmt-d74d6e1/b16966 emits nothing/exit0.
Other five generated copy artifacts and deployment manifest byte-identical to
baseline0e4f46. No English/key/constant/list/order/companion identity changed.

Three compiling template severings fire, never a syntax/build failure:
fb21e2 All one-line → empty golden failure;0be59e Companion one-line → empty
golden failure;ed5e0a Companion bound to allKeys → single/default failure.
Restore exact template SHA0e48a169dfbc29568ba602c3608b27ae3ce26ae4b91615577c806e6b4faf9cfc
after each terminal result98bdc4/5229cb/b284d1. Output-only wrong constant
cb77f2 is rejected by actual existing verifier as generated artifact drift,
after five golden controls pass; exact output restoreda76c73 to
a461181e70f79a436e1ebf0fdb71e33e6b1d8faeaa7c1c7fbb0577bf8eb82604.
Node-only fixture/template import adds no Go/formatter invocation/dependency.
No test assertions/history guards removed or policy/deadlines/skips altered.

Cold02219d/763d7a/3295d1 session45090 terminal0: types0 errors/warnings,
build213 modules;units8106 pass/255 Node skips,91 files pass/19 skip. Shell/
topology13/cosmetic22/no-payment6 controls pass. Actual copy check runs new
fixture import, copy658/hash a5df8920bc1573c8b0543b0c6a3c4c2de955b580626c8e1d130fc9e1adeb0e5e
unchanged/611 orphan warnings;manifest passes. No native/SQL/host/harness career
rerun is claimed for this formatting-only range.

Cold coreeebdbb/801a88/1dfe4f/813c75/48bb01 session29154 terminal2:
three deployment/operations packages fail httptest loopback binding denied by
sandbox; all other listed non-harness packages pass, with usual Postgres tests
not exercised without DB. Original make target DID NOT pass; chained vet was
not reached. Separate root vetebec66 terminal0. After original handle terminal,
narrowly escalated root make test-go347cfc/d8c1a7 session61567 terminal0 reruns
all three failed packages cold:5.313/2.390/0.876s,all pass. This gives successful
core package coverage by union, not an exit0 claim for the original command,
and no DB/deploy proof. Local permission repaired execution, not test semantics.

Separate history572061/1dc21a session12596 terminal2: checkout contract/fixtures
pass then same historical RP-13150a3a514 vs0cf9f7a6 fails. Not whole client/CI
green. All handles terminal before any record/commit. Generated Go producer
tree changes despite formatter-equivalent behavior; old dated career artifacts
remain historical at their producers. No numeric/kernel161/epoch/copy/balance/
formula/price/CI/security or owner-authority change, box flip, archival/push/
cleanup/release claim. Full span after3df3ff32 needs Claude independently of
earlier ranges. Next ground RP-284 Run End/new-route consumer and author bridge;
proper nine-tier/platform1.0 remains active.

## 2026-10-06 — Copy generation range self-first-filter

**Review by:** Codex (implementer self-first-filter, NOT designated review).
**Recorded by:** Codex.
**Reviewed range:** `3df3ff32..903c901d`, four commits/twelve paths, including
predeclaration, test/verifier wiring, paired template/output/docs and all trackers.
**Verdict:** first-filter passes; requires Claude's independent exact full span
including this record edge. Earlier ranges remain independent; no archive.

Complete non-generated diff2856f8 and scopea46154 inspected. All generated Go
bytes inspected by independent exact gofmt(baseline)-equivalence, not a visual
sample of the giant list. Full check also confirms all other generated artifacts
unchanged. Goldens independently hardcode output; input order/subset separation/
empty/default/collision are covered; imported fixture runs before ordinary
generation/history guard and adds no formatter process. Actual template and
output faults fire, not merely corruptions of the fixture itself. No prior
assertion removed, history/owner-prose/CI/numeric policy modified, or report
producer restamped. Baseline/core sandbox failure and narrower successful
reexecution retain distinct results; usual SQL skips and RP-131/Firefox remain.
Tracker append initially matched an older repeated Evidence line; detected by
diff before commit, moved to EOF, and39de4e byte-prefix check proves both logs
append-only. Whitespace f20b75 passes; current tree otherwise clean. No checkbox
flip or full game/platform acceptance. Next separate RP-284 consumer/author
bridge grounding; full1.0 goal stays active.

## 2026-10-06 — RP-284 producer ground and bounded reader predeclaration

Clean start `0ae1fa11`. Accepted R7 already emits run_started v1 (tree absent)
or v2 (null/object). Production prestige.go writes the post-plan frozen factor
and applied starters in artifact order. save/intent.go validates exact fields,
canonical state-valued factor >=1 and unique mechanical IDs. Outbox migration42
marks ordinary events advance, not historical. Client events.ts ignores
run_started entirely; runtime additionally suppresses an event whose Company
revision was already sampled by HTTP. Existing run_ended has a bounded
same-Founder/exact-successor exception. No summary consumer currently exists.

R9's balance display is a separate DESIGN-GAP: ended has payout delta, not
available balance; founder_advanced has delta too; receipt state is Company
wire, not the v4 Founder feature. Existing authoritative v4 snapshot reads
latest Company then sibling Founder, not an Exit-bound balance row. The older
Game UI archived GU-C3 explicitly forbids a snapshot parameter on RunEndSurface
and a compile-time negative enforces it. R9 requires available after Exit but
does not reconcile that boundary. Do not relabel an arbitrary later Founder
balance or old balance plus advisory plan as the exact Exit balance; route
bridge/body reconciliation to the author. Also docs/game-ui.md's claim that
next-run snapshot binds only on Continue is stale: act already refreshes every
applied Exit; bindSnapshot leaves the terminal surface intact. Source finding,
not a reproduced navigation failure. RP-283 remains an independent payout gap.

Proceed with only the R7 reader/delivery supplement predeclared in plan.md.
Tests first: absent/null/object positives with intentionally non-lexical
starter order; exact-shape/type/identity/canonical-factor/duplicate-ID refusals.
Controlled real runtime covers both HTTP/event orders, wrong-Founder/old/future
duplicate controls, channel-offset dedup and malformed summary resync. Baseline
must fail; compiling reader/validation/race omissions must fail independently.
No UI fixture or whole browser/SQL/mint claim. Root client checks after exact
restoration; inherited RP-131/Firefox and all owner/data/platform/review holds
remain. Full span after0ae1fa11 needs Claude; no checkbox/copy/kernel/numeric/
schema/CI/security/archival/publication/deployment/push or goal promotion.

## 2026-10-06 — RP-284 reader baseline

Root make test-client (48c7a4/30897d, session25261 terminal2) executes41 new
cases:38 fail, three duplicate-cohort controls pass; existing8106 pass and255
Node browser skips remain. Reader returns undefined for all three payload
versions and admits malformed ignored events without resync. No production
change yet. The initial snapshot-first fixture sampled the event's exact
revision: PlayerRevisionCursor reset has an empty seen set, so that arm alone
does NOT prove duplicate suppression. Refine the population before production
edits to cover both equal and advanced sampled revisions; negative duplicate
cohort controls use the advanced revision. No assertion relaxed, and no claim
that the first run demonstrated the advanced-revision defect.

Refined baseline a18d24/4a637f (session76902 terminal2):42 new cases,39 failures
and three controls pass; existing8106/255 skips unchanged. Both sampled-revision
arms now execute separately. Commit test-only diagnostic before correction.

Reader-only correction bc85d3/54ad6a (session20349 terminal2) reduces failures
to two. The advanced-revision race now fails independently while equal-revision
and ordinary delivery pass. The other failure was a bad diagnostic assumption:
1e1001 exceeds a resource hardcap, NOT Decimal's state exponent limit. Replace
that refusal input with literal1e9000000000000000 and retain1e1001 as a positive
control; do not tighten numeric behavior to satisfy a false fixture. Add exact
start-time duplicate cohort control before bounded runtime correction.

Corrected reader-only diagnostic80c60f/36963b (session14390 terminal2):44 new
cases, only advanced-revision delivery fails;43 new controls and existing8106
pass/255 skips. Apply a Company-only duplicate exception for exact sampled
Founder/run sequence/start time, alongside unchanged preceding-run terminal
exception. No cursor reset/offset/gap/reconnect/auth policy change, UI navigation
or snapshot prop. Canonical docs accompany reader/runtime behavior and correct
the stale snapshot-only-on-Continue claim without changing the old RFC body.

Initial correction954cb0/fce803/727bfd (session67806 terminal0):44 new cases
pass, total8150/255 skips; types0 errors/warnings, build213 modules; shell,
CI topology13/cosmetic22/no-payment6 controls, copy658/hash unchanged and
manifest pass. No native browser/SQL/CI-run proof. Self-inspection finds the
new duplicate exception could re-deliver the same summary at a new channel
offset, unlike the ordinary cursor's event-ID dedup. Extend all three delivery
arms with that assertion before touching production again. Add only bounded
single-current-summary identity memory if diagnosed; do not grow an unbounded
event-ID set or change other events' policy. This is a supplement-specific
regression check, not authority to repair unrelated existing terminal replay.

Extended duplicate diagnostic dc1807/d562ef (session41907 terminal2):both
snapshot-first arms fail duplicate delivery at a new offset; event-first and
all other controls pass. Add one last-delivered run_started event ID, scoped
to the subscription; only this event is deduplicated by the new memory.
No growing set or other lifecycle replay policy change. Test/probe this guard
as a fourth compiling omission, in addition to the original three probes.

## 2026-10-06 — RP-284 discriminating reader/runtime probes

Final population adds the next distinct run and later snapshot-reset duplicate
controls:45 new cases;66461f terminal0 total8151/255 skips. Four source-only
compiling severings execute the unchanged root unit population:
reader disabled at nonnegative event revisions b3a927 terminal2 fires41;
below-one factor bound removed9bbc96 terminal2 fires2 (reader and runtime
resync); current-run race exception removeddeda8f terminal2 fires1 (advanced
snapshot only); repeated-event guard removedff4442/4de7fd session28228 terminal2
fires3 (all delivery orders). No syntax/compiler failure used as a gate result.
Events source restores after each to SHA256
93d0ae593d3ec388cf899ec17876db3c65ffde30d7d56f5326865e09494375b7;
runtime restores to f8f9bae2348f103cb57b98a9bee6f64d7057dbcde732f142229be5706ab95dba
(c7922f/ffe13f/e4023f and final restoration). No test assertion removed.

All probe handles terminal before restoration. Execute final types/build/units/
boundaries and the inherited historical kernel guard separately. Add browser
execution of this exact controlled-runtime population plus existing Game UI
screen regressions on Chromium/WebKit through the existing root Make lane;
its existing isolated performance gate remains enabled. This is browser-side
reader/regression evidence, NOT native Exit action/real network/Postgres/mint/
manual AT/whole AC12. Firefox's existing unexecuted hold is not waived. No
flags/preferences/timeouts/security or CI policy changes.

Final local e36c7f/c576a4 session71419 terminal0:types0/build213/units8151,
255 Node skips and boundaries/control populations pass. Separate history
0974f8/f82cd0 session77327 terminal2:checkout/fixtures pass, unchanged RP-131
50a3a514 vs0cf9f7a6 fails; not whole client/CI green. Browser4fee09/dabe0f
session62008 terminal2 runs134 functional cases:133 pass, one Chromium recovery
assertion fails; two performance-selector skips visible. WebKit's samecase
passes. All36 observed Worker requests complete, zero pending. The isolated
performance target was NOT reached because the functional target failed.
Failure is the diagnostic's arbitrary10-microtask wait for Response.json, not
a recovery-semantic finding: resync notification/second fetch/close all fire,
but browser parsing has not finished. Replace counting microtasks with awaiting
the actual runtime snapshot callback and assert its schema/revision/Founder/run
as well as zero event delivery. Existing test deadline remains unchanged; no
delay/timeout increase, production edit, reduced oracle or hidden browser miss.
All handles terminal before this test-only correction.

## 2026-10-06 — RP-284 final reader proof and remaining display route

Callback-bound final fixture13f2f4/ae2058 session68480 terminal0:types0 errors/
warnings, units8151 pass/255 Node skips. Browser ee6cec/01c341 session65020
terminal0:134 functional Chromium/WebKit cases pass, two performance-selector
skips; isolated unchanged performance gate1 pass/22 selector skips. All36
functional Worker requests and the one isolated request complete, zero pending.
New45 cases execute in both browsers (90); the other44 functional cases are
existing screen regressions. This is controlled HTTP/socket reader execution,
not a real server/socket/SQL, user-driven Exit or whole AC12 claim. Earlier
133-pass/one-failure arm remains recorded. Post-fixture below-one omission
8bc74f terminal2 still fires reader/resync cases; exact restorationf20851
passes total8151 again23ce27. Events/runtime SHAs above remain exact; no edit
while any test handle live. Build/boundaries/copy/manifest from the unchanged
production source are the separately recorded passing arms, not one whole
verify-client/CI verdict. Historical guard remains RP-131 RED.

Reconcile backlog/plan/current-state/roadmap/execution and append-only checkpoint
logs in the source+docs range. Add precise author finding in decision queue:
current-vs-Exit-bound available balance is unspecified across R9 and GU-C3;
no snapshot prop, latest-balance interpretation or event/schema addition is
self-authorized. The accepted reader is a mechanical fragment, not a rendered
feature. Next safe accepted work is recovered-publication R7 identity/delivery
checks, while RP-283/RP-284 display wait for author reconciliation. Full new
span after0ae1fa11 needs Claude independently of previous ranges. No boxes/
numerics/kernel161/balance/price/prose/epoch/CI/security/owner/RFC/archive/
cleanup/publication/deployment/push/report producer change; full1.0 stays active.

## 2026-10-06 — RP-284 reader supplement self-first-filter

**Review by:** Codex (implementer self-first-filter, NOT designated review).
**Recorded by:** Codex.
**Reviewed range:** `0ae1fa11..47ccc003`, all three commits/twelve paths:
predeclaration004f1c27, test-firstdf327d81, paired source/docs/tracking47ccc003.
**Verdict:** first-filter passes; Claude must independently cover the full range
including this record edge. No acceptance/archival; earlier ranges independent.

Complete test613717/source+docs bf216d/tracker12df9b inspected. Reader mirrors
the existing R7 producer validator, preserves absent/null/object and artifact
order, and does not touch payout/available math or the terminal component's
compile-time boundary. Runtime exception is Company-only and exact current
Founder/run/start time; one last event identity prevents its own repeat without
unbounded memory. Existing generic cursor/recovery/auth policy is unchanged.
Controls cover ordinary/equal/advanced HTTP order, same/new offsets, later reset,
other cohorts, next distinct run and malformed recovery. Four compiling faults
discriminate; final callback-based fixture still catches below-one omission.
Node/browser results, original invalid numeric fixture and Chromium async miss
remain separate and disclosed. No fixture skip/deadline change or evidence
promotion; native, SQL/mint/full AC12/Firefox and historical RP-131 holds stay.
ff608a verifies both logs append-only from0ae1fa11; full whitespace check passes.
No Go/copy/kernel artifact changes, stale report restamp, checkbox or RFC body
edit. Precise display gap is routed to its author, not silently closed by the
reader. Next accepted supplement: recovered-publication R7 identity checks.

## 2026-10-06 — R7 disconnect/recovery predeclaration

Previous goal turn was progress:47ccc003/61d6c8eb committed the reader/delivery
supplement, actual browser proof, corrected fixtures and synchronized tracking.
Fresh52b134 is clean61d6c8eb; no inherited live handles. AGENTS164eea/process
2acc51 and entire accepted RFC2309b0/aa25fd/93dfec reread. Bound design refs
remain unchanged from020a25c6. Actual runtime dd69b8 and existing recovery tests
058748 ground persisted-offset request, replay loop, fresh-state fallback and
the per-subscription last start-event identity. R7/R9 remain accepted; no new
authority for Run End display/available bridge or RP-283 payout.

Plan above predeclares a controlled real-runtime recovery population, preserving
all45 existing tests and all deadlines. Exact callback promises replace timing
guesses. Source risk RP-287 enters shared backlog immediately: delivering a new
start replaces the only remembered ID while an old snapshot can still authorize
reviving an older start at a new offset. Test live/recovered-batch order before
correction; if confirmed, bound monotonic Company start revision in existing
memory, not a growing set. No generic cursor semantics, auth/session refresh,
transport reconnect/drain/epoch policy, component/copy/wire/balance/epoch/CI/
kernel161 changes. Failure outcomes and competing controls must be recorded.
Source-only compiling probes after positive run, exact restoration after every
terminal handle; root unit/types/build/boundaries/copy and selected browser/
isolated performance. Full span after61d6c8eb needs Claude independently of
earlier ranges. Full nine-tier/platform1.0 objective remains active; no shortcut,
acceptance/archival/publication/deployment/push or owner content adoption.

## 2026-10-06 — R7 recovery unchanged-source diagnostic

Helper extension alone preserves old45/ec7fd8 terminal0 total8151/255 skips.
New55-case population749b56/c35ea5 session88688 terminal2: two stale-start
cases fail with a real extra run2 delivery after run3 (live/recovered batch).
Four fresh-state controls time out because their diagnostic snapshot incorrectly
put the new run's start after server_now_ms; parser correctly rejects it and
runtime closes instead of delivering a snapshot callback. Those four failures
are INVALID FIXTURE evidence, not product-recovery defects. Correct the sampled
server/evaluated time to the new start and validate the fresh fixture through
the actual parser before injection. Keep all recovery/close/read/event oracles
and5-second test deadlines unchanged. Production bytes are still untouched.
Move the new plan subsection before the original Batches heading so old B1–B10
do not appear owned by this supplement; no checkbox or task authority changed.

Refined ae6f0b terminal2 executes55 cases:two RP-287 live/recovered failures,
53 controls passing; total8159 pass/255 skips. Fresh-state four arms now execute
their intended snapshot/reconnect/refusal assertions and finish, rather than
timing out on invalid data. The old45 assertions remain unchanged; actual1006
reconnect retains its original1-second delay. Commit diagnostic before source.

## 2026-10-06 — R7 recovery correction and probe declaration

Test-first f91e9b0e retains the reproduced defect. Bounded Company start-revision
memory corrects RP-287 without changing the generic cursor, offsets, reconnect,
drain, auth or epoch policy. Corrected55-case cf25e8 terminal0:8161 unit passes/
255 unchanged skips. Four recovered wrong-cohort controls plus actual1006
high-water retention and full-sync identity retention extend the population to61.
2d58c3/33dbf7/c9d098 session95543 terminal0: types0 errors/warnings, build213
modules,8167 unit passes/255 unchanged skips; client boundary, CI topology,
cosmetic/payment controls, copy pipeline and Go content-manifest check pass.
The copy history scan completed; no timeout, restart or truncation. This is not
whole verify-client/CI: inherited RP-131 history and Firefox remain separately
red/unavailable. No SQL, live server/native Exit, display or whole AC12 claim.

Corrected runtime SHA256 before probes:
9aef7b48ad2b72becee3ef1f874e6672d304bcf20aec16e8afbbe5c64db26c2e.
Predeclare three serial compiling source-only faults: omit recovered-publication
iteration; reset both start-memory fields when connect begins; remove the
Company-start revision high-water predicate. Each runs the unchanged root unit
lane, records fired cases and controls, waits for its terminal handle and then
restores exact bytes. No source/test/record edits with any handle live. Restored
selected Chromium/WebKit functional cases and unchanged isolated performance
lane follow. These faults test replay delivery, reconnect memory and RP-287
supersession respectively, not the producer or real transport/server behavior.

Fresh Docker observations861222/1c948f:both declared Postgres containers healthy;
database tmpfs136ad0 has8,091,276 KiB available (1% used). Root capacity3e88af
still100%,39,784 KiB free. Earlier compound-command socket denial5f3893 was a
permission result, not capacity evidence; the standalone read succeeds. No
cleanup authorized/performed. Real SQL/mint proofs remain blocked by writable
Docker capacity, not by the controlled-client successes.

## 2026-10-06 — R7 recovery discrimination and restored browser proof

Recovered-publication omission ac804f/cc5d3c session27872 terminal2: six new
recovery cases plus one existing generic history case fail,8160 units/255 skips
remain. Identity reset at connect7ba7a7/2ea725 session9115 terminal2: all three
actual1006/full-sync retention controls fail with extra delivered starts;8164
units/255 skips remain. High-water predicate omission ac28e5/268215 session4116
terminal2:live, recovered-batch and lagging-HTTP reconnect supersession fail,
8164 units/255 skips remain. All failures are semantic assertions, not compiler
errors. SHA checks f72498/400044/f6ea28 verify exact corrected runtime after
each terminal probe; no handle cancelled/restarted and no tests/oracles relaxed.

Restored db1d4b/540e87 session83109 terminal0: types0 errors/warnings,213-module
build,8167 unit passes/255 unchanged skips, shell/UI boundary, CI topology13,
cosmetic22 and no-payment6 negative controls pass. Copy/manifest passed on these
same source bytes in session95543 above; not rerun or claimed as a whole CI gate.
Browser6f240c/43c7a8 session78119 terminal0:166 functional Chromium/WebKit
passes, two performance-selector skips;122 are all61 next-run cases in both
browsers,44 existing screen cases. Unchanged isolated performance1 pass/22
selector skips. Functional Worker36/36 and isolated1/1 requests finish, zero
pending. Real browser runtime execution over controlled socket/HTTP inputs,
not a real server, SQL career, minted tree, native user Exit or whole AC12.

RP-287 is locally corrected and its former baseline/probe failures stay recorded.
RP-284 remains an unresolved Run End/post-Exit balance input contract; RP-283
remains the Wind Down authoritative preview gap. No checkbox, authored copy,
wire/schema, balance/kernel161, transport/auth/CI policy, owner/author ruling,
report restamp, mint, cleanup, publication/deployment/push or archival changed.
Whole span after61d6c8eb through its final record edge needs Claude independently
of earlier ranges. Next safe accepted work is remaining R9 component/host states
and advisory-plan consumer grounding, not bypassing the author/data/SQL holds.
Full nine-tier/platform1.0 remains active; this turn made concrete runtime progress.

## 2026-10-06 — R7 recovery supplement self-first-filter

**Review by:** Codex (implementer self-first-filter, NOT designated review).
**Recorded by:** Codex.
**Reviewed range:** `61d6c8eb..8eb10d98`, all three commits/ten paths:
bb9edc41 predeclaration,f91e9b0e diagnostic,8eb10d98 source/docs/tracking.
**Verdict:** first-filter passes; Claude must independently review the complete
span through this record edge. No acceptance/archival; earlier spans independent.

Complete range source/tests/docs and all tracker/log diffs inspected. Runtime
change is one per-subscription Company-start high-water plus its delivery guard;
generic revision cursor, offset persistence, replay loop, reconnection/drain,
auth and full-sync behavior are unchanged. Stored memory remains bounded and
survives those paths. Original45 assertions are intact; helper defaults preserve
their original behavior. New16 cases execute replay compatibility/cohorts,
superseded ordering, actual1006/full-sync retention and four fresh-state refusals.
Snapshot fixtures use the actual parser before fault injection. No arbitrary
microtask budget, skipped failure or increased deadline. Three compiling source
faults discriminate the named properties; restored Node/browser evidence and
original invalid fixture outcomes remain individually recorded.

Append-only comparison from61d6c8eb passes both logs; whole-range whitespace
check and clean tree pass. Records distinguish local RP-287 correction from
unresolved RP-283/RP-284 contracts and actual SQL/mint/full AC12. No checkbox,
Go/copy/kernel/epoch/CI/ruling/owner prose or release status changed. Fresh
capacity observation is not a SQL success or cleanup permission. Next separately
ground remaining accepted R9 component/host/advisory-plan behavior; full1.0 active.

## 2026-10-06 — Full-tree advisory-plan predeclaration

Previous goal turn is progress:8eb10d98/e312b4d7 commit the actual R7 correction
with executed discrimination/browser proof and synchronized tracking. Fresh
clean e312b4d7, no inherited live handles. AGENTS/process and entire accepted
Reputation RFC reread; bound design refs unchanged since020a25c6. Actual plan/
tree components and host controls inspected; existing host plan tests use two
independent nodes, not the full declared DAG or sequential disclosure/checkbox/
Clear path. Earlier native WebKit tree-control findings make implicit plan
stops a source risk, RP-288, not a claimed new browser failure.

Plan predeclares both-era full nine-row advisory state/order/budget/cascade and
native keyboard/axe populations. Use the actual declaration in controlled
component props, not a minted epoch or live server. Keep RP-283 authoritative
Wind Down payout and RP-284 available-balance/display contracts separate.
Baseline before source; if native stop failure occurs, minimal R9 accessibility
correction only. Existing algorithms/content/owner prose/server/wire policies
remain. Demonstrated compiling source faults, exact restoration and terminal
handle discipline; root/client/browser gates and unchanged perf lane follow.
All old45/61 reader, host/surface/cost controls and deadlines preserved. Complete
new span aftere312b4d7 needs Claude independently of previous spans; full nine-
tier/platform1.0 active, no shortcuts, box flip, acceptance/archive/cleanup/push.

## 2026-10-06 — Full-tree plan unchanged-source diagnosis

Declaration path is balance/testdata/reputation-tree/fixture-v1.json, not the
RFC's future balance/reputation-tree/phase1.json (which does not exist before
ratification/mint). No fixture/data adoption or source rewrite. Type/unit
2a5018/11e400 session33584 terminal0:0 errors/warnings,8167 passes/273 skips
(18 new native tests are correctly Node-inapplicable).

Native b2c3ff/3a5794 session45875 terminal2:28 controls pass, four WebKit native
paths skip the first checkbox and land on the after-sentinel; both eras and
Enter/Space reproduce RP-288. Four additional axe failures name the test-only
unstyled before-sentinel's target size, not a panel element. Those are INVALID
FIXTURE evidence, not a product-accessibility defect. Remove only external
sentinels from the component axe profile; leave its zero-violation oracle/tag
set, every panel control and keyboard population/deadlines unchanged. Re-run
baseline before source. Production SHA251343ff0583a62a2b3ca5faa5332f8d80c8de6a3c8e7b3e161f52c1ed7de861
unchanged. Refine minimally with explicit checkbox stops first; if Clear then
fails independently, add its explicit stop too. Summary already natively works.

Refined unchanged-source4a5eff/ae4952 session23883 terminal2:all32 non-WebKit-
keyboard controls pass including both-era zero-violation component axe. Exactly
four native WebKit checkbox paths fail; no sentinel failure remains. Commit
this diagnostic before the source correction; old surface/host tests untouched.

Checkbox-only refinement cc6f5e/311636 session92300 terminal2:the same four
WebKit paths now reach and operate both enabled boxes, then skip enabled Clear.
All32 other cases still pass. This independently establishes both missing
control stops; Summary already works, so leave it unchanged. Add explicit
zero Tab indices only on checkbox and Clear. No algorithm, outgoing plan,
component input, copy, browser preference or test changes.

Corrected b3627e/3db8fe session53352 terminal0:36 native Chromium/WebKit cases
pass and unchanged isolated perf1/22 selector skips pass. Component SHA now
77a4b842f989b29d748195e405e29fefb3fc2a3f0a3a650f23db969b99fd3d6f.
Before final gates, predeclare five compiling source-only faults with unchanged
36-case oracles: remove selectable prerequisite requirement; remove selectable
budget condition; keep click order instead of artifact order; treat each
deselection prerequisite as satisfied; remove both new native stops. Each root
selected-browser run must reach terminal before exact source restoration, then
SHA check. These test the component's declared advisory contract, not server
atomicity/payout or adopted data. No editing with a live handle, deadline/skip/
budget/rule relaxation or browser preference change.

## 2026-10-06 — Plan probes and native screenshot-copy stall

Prerequisite omission7a9fae/1a1fff session25584 terminal2:20 failures/16 controls
pass across Chromium/WebKit. Source restored SHA77a4b842...fd3d6f before next
probe. Budget omission9747d6 session92636 prints14 Chromium assertion failures,
then stalls: no completed WebKit/whole-population verdict. Do not count it as a
normal36-case result. Same handle repeatedly polled; no restart due to timeout.
Read-only ps confirms owned Make18744→shell18757→Node18761 and live browser
children, Node98–100% CPU. Two one-second samples231c61/e2b325 and1000ed
(full result retained in tool store planBudgetSample) show main loop idle and
one fs worker in pread/sendfile. lsof identifies a4,254-byte screenshot source
and .vitest-attachments destination, distinct inodes; both offsets4,254.
Host filesystem has160,000,000 KiB free; this is not the Docker-capacity hold.

This is a verified native copy stall, RP-289, not just an elapsed observation.
Narrow graceful termination999834 closes browser/server (zero pending modules)
but same Node remains100% CPU; fresh ps confirms it. Targeted kill8f5983 then
c43176 session92636 terminal2/Error137. No unrelated process stopped, file
deleted or dependency/CI config changed. Only after terminal restore exact
component SHA77a4b842f989b29d748195e405e29fefb3fc2a3f0a3a650f23db969b99fd3d6f.
Sample handles95591/12531 both terminal0; no handle remains live.

Predeclare diagnostic-only adjustment for remaining negative probes: root
test-browser selects the same36 cases with --browser.screenshotFailures=false,
whose installed Vitest CLI declares that option. It omits automatic failure
image/attachment generation AFTER an assertion fails; it does not alter DOM,
browser security/preferences, native keyboard, oracles, deadline, skips or
error guard. Stdout must still show semantic failures and terminal counts.
The previous incomplete arm stays on record; repeat the budget fault as a
new declared capture arm, not a replacement success. No permanent option,
workflow/Make change or green CI claim. Final restored positive browsers run
DEFAULT screenshot behavior and existing perf lane. Native copy cause remains
unassigned; no spec or owner ruling is inferred from this instrumentation.

Capture-syntax correction:bd996e/03c577 session7055 terminal2 completes28
budget failures/8 controls, BUT --browser.screenshotFailures=false did not
disable capture: output still names screenshots and actual screenshot mtime
advances21:00:23/17,787 bytes. Do not credit that run as a no-capture mechanism.
Installed tester checks option truthiness. Correct CLI negation is
--no-browser.screenshotFailures; a8b338/eb5407 session6742 terminal2 produces
12 click-order failures/24 controls, no failure-image artifact output. Semantic
DOM/plan oracles unchanged. Repeat budget under the correct negation, then
remaining cascade/keyboard probes. Final positives remain DEFAULT capture.

## 2026-10-06 — Full-tree plan correction verified locally

Correct CLI-negation budget repeat0bd765/session23428 terminal2:28 failures/
8 controls pass, no failure-image artifact output. Cascade omission86b3ef/
session66529 terminal2:8 failures/28 controls. Native-stop omissionbed6b5/
b1d76c session55997 terminal2:4 WebKit failures/32 controls. Together with the
earlier prerequisite20 and artifact-order12 failures, all five compiling
source faults discriminate their declared properties. Every arm reached
terminal before source restoration; final component SHA is exactly
77a4b842f989b29d748195e405e29fefb3fc2a3f0a3a650f23db969b99fd3d6f.
No oracle, deadline, browser preference, security, skip or assertion relaxed.
The incomplete native-copy arm, invalid axe fixture and mistaken '=false'
capture syntax remain recorded above, not replaced by the successful reruns.

Restored DEFAULT-capture browser44171b/session55816 terminal0:328 functional
Chromium/WebKit passes across new plan and existing surface/host/cost/screen
populations, two performance-selector skips; Worker131/131 requests finished,
zero pending. Unchanged isolated performance1 pass/22 selector skips, Worker1/1
finished, zero pending.36 new full-tree cases cover both eras, budgets,
prerequisites, exact nine-node cost552, canonical order, cascading deselection,
native forward/reverse Tab and Enter/Space, and component axe zero violations.
This controlled component/host proof is not live SQL/server/mint/default-player
Exit, Firefox, manual assistive technology or whole AC12 acceptance.

Root gate9a6337/session50431 terminal0 at a93a3c: types0 errors/warnings,
213-module build,8167 unit passes/273 skips (18 new native cases are Node-
inapplicable), shell/UI boundary, CI topology13, cosmetic22 and no-payment6
negative controls pass. Copy658 and its SHA remain unchanged; five generator
goldens/collision/six corruptions, append-only copy history and content manifest
check pass. Existing611 orphan warnings remain visible, not newly resolved.
No Go runtime, schema, kernel161, authored copy, balance/epoch or CI policy
changed; this is not a complete green CI claim (RP-131/Firefox remain separate).

RP-288 is locally corrected with only explicit Tab0 on checkbox and Clear;
Summary and all advisory algorithms are unchanged. RP-289 native screenshot
copy stall remains OPEN: diagnostic-only no-capture syntax enabled completed
negative arms, while restored positives used normal capture. No dependency or
permanent screenshot-policy fix is claimed. RP-283 authoritative Wind Down
preview and RP-284 post-Exit balance/display still require author contracts.
Whole span aftere312b4d7 through final record edge needs Claude independently
of earlier ranges. No checkbox, acceptance, archival, cleanup, mint, report
restamp, publication/deployment/push or full nine-tier/platform1.0 promotion.
Next ground R9 host inactive/offline/in-flight/refreshing purchase controls,
retaining the actual runtime/parser/DOM boundary and existing deadlines.

## 2026-10-06 — R9 plan supplement self-first-filter

**Review by:** Codex (implementer self-first-filter, NOT designated review).
**Recorded by:** Codex.
**Reviewed range:** `e312b4d7..82ac87b0`, all three commits/ten paths:
d44d209c predeclaration,6dff42fc test-first,82ac87b0 source/docs/tracking.
**Verdict:** first-filter passes; complete span through this record edge needs
Claude independently of every earlier range. No acceptance or archival verdict.

Complete source/test/docs/tracker diffs and appended log inspected. Production
diff is two explicit Tab0 attributes only; eligibility, selected-state/callback,
ordering, budget and cascading algorithms remain byte-identical. New fixture
uses the actual nine-row test declaration, explicit controlled balances/preview,
both eras and real native events; only the sequential path claims Tab reach and
it focuses solely its sentinel. Axe removes test-only sentinels, not a product
element or violation. Original host/surface/cost tests, assertion deadlines,
browser security/preferences/error handling and CI configuration are unchanged.
Five compiling faults catch the declared properties; every restored arm is
terminal before edits. Interrupted copy stall and mistaken CLI syntax remain
disclosed. Final browser run uses default capture, not negative-probe diagnostics.

Append-only comparison frome312b4d7 passes both logs; whole-range whitespace
check passes. Kernel161/copy658/manifest/Go/wire/epoch are unchanged. Records
retain RP-289 tooling, RP-283/RP-284 author contracts and all full1.0/SQL/mint/
Firefox/review holds rather than promoting component proof to acceptance.
Next separately ground accepted R9 host state controls; full goal remains active.

## 2026-10-06 — R9 host readiness predeclaration

Previous turn is progress:82ac87b0/b5ca3e7d commit minimal actual keyboard
correction, executable negatives and synchronized records. Fresh clean
b5ca3e7d, no live inherited handles. AGENTS/process/index and accepted
Reputation RFC reread; bound design refs unchanged since020a25c6. Transport
RFC fully read for existing network-drop/drain/full-sync authority. Actual
host/component/runtime and original native tests inspected. Reputation's
founderControls omits transportReady; mounted tree has no offline notice,
RP-290 source risk. Runtime's ordinary close path reconnects but never tells
the host it is not recovered, RP-291 source risk. Neither is an executed defect
yet. Existing host tests auto-ack all subscriptions, masking these populations.

Plan predeclares controlled actual-runtime/native both-era state transitions
and all disabled/busy/focus/revision/copy/nonoptimistic oracles. Hold subscription
responses, not timers; native reconnect retains actual one-second delay and
existing poll deadlines. Minimal possible source correction bounded to R9
binding and internal transient notification under Transport D2/D4/T4, never
terminal-unsubscribe or new auth/recovery/wire policy. Keep old tests intact;
test-first baseline and compiling severings before final local gates. No manual
edits with live test/probe handle; no acceptance checkbox, author/body/copy
adoption, schema/balance/kernel/CI/mint, cleanup, push/deploy or archival change.
Entire new span afterb5ca3e7d needs Claude, earlier ranges independently pending.
Full nine-tier/platform1.0 active; old author/data/SQL/browser/release holds remain.

Test fixture extension keeps the original boundary's default automatic replies
and every old assertion. New40 native cases explicitly hold individual channel
acks and exercise actual one-second recovery, not fake timers/host fixture
mutators. Typecheck06e769/e5f614 session22339 terminal0,0 errors/warnings.
Predeclare diagnostic baseline and compiling negative arms with actual CLI
--no-browser.screenshotFailures because RP-289 already verified a native
failure-attachment copy stall. This omits post-failure diagnostics only, not
assertions/native events/deadlines/security/error detection or test populations.
Retain the option/result on record; final restored full positives use DEFAULT
capture. No Make/Vitest/CI configuration is edited and no tooling fix is claimed.

Unchanged-source7d1e6f/d4463a session3813 terminal2:64 failures/16 controls
pass,96 old cases selector-excluded (not skipped defects). Each browser has
startup4/drain4/post-full-sync8/ordinary1006 drop4 readiness failures and
terminal4001/4002 eight/failed-refresh four missing-offline-notice failures.
Both-era Enter/Space controls execute. Inactive initial/arm-loss and successful
in-flight/authoritative-refresh controls pass. Worker80/80 finished, zero
pending; no attachment-copy stall/cancellation or incomplete population.
RP-290/RP-291 are now executed findings, not source hypotheses. No test fixture
or oracle refinement is required by this baseline. Commit test-first before
adding scoped Reputation readiness/offline binding and a distinct nonterminal
runtime notification (terminal auth/replaced disposal remains unchanged).

Corrected default-capture45803c/64fa9e/e2d31d session43479 terminal0:all80
new native cases pass,96 original cases selector-excluded; isolated perf1/22
selector exclusions pass. Worker80/80 and1/1 finished, zero pending. No
fixture/oracle/deadline changes between red and green. Correction adds scoped
Reputation readiness and existing offline copy plus an internal nonterminal
recovering message; auth/replaced terminal disposal is unchanged.

Before final gates, refine the predeclared compiling source probes precisely:
omit the Reputation binding's transportReady; omit runtime's transient
notification; omit the existing-copy offline notice; omit host refreshPending
from shared pending. Run the ENTIRE host population, not just new cases, so
the fourth also retains the original held revision-conflict refresh controls.
The new applied path awaits its action through refresh independently, and is
not claimed to be its sole discriminator. Other pending/runtime controls must
remain. Each arm diagnostic-only no-capture (RP-289), terminal before exact
source restoration; final restored positive DEFAULT capture. No new policy,
budget/skip/deadline/authority relaxation. Initial RP-290 ledger's per-browser
readiness count was mistyped12; correct16 (=startup4+drain4+full-sync8), plus
12 notice failures and four RP-291 failures. Baseline log/count64 is unchanged.

All four unchanged-oracle compiling probes terminal2, whole176-case host:
readiness omissioncdc04e/a08cd0/bee293/abcf35 session6563:32 fail/144 pass;
recovering-notification omission80bbe6/e3a3bc/87c8a8 session99556:8 fail/168 pass;
offline-notice omission53b57b/b318d6/21acf6/b48213 session89434:64 fail/112 pass;
shared refreshPending omissionf96362/2702f1 session31152:16 fail/160 pass.
Last discriminator is the original held HTTP200/409 revision-conflict population,
not a fabricated claim that new applied cases alone catch it. Each worker176/176
finished, zero pending. No screenshot artifacts/copy stall or handles cancelled.
After EACH terminal result restore exact three-file SHA:
GameUIApp a471dea6e8be0725ccd762c64f7829c2568fba8ce11a1cad2cacc50a06a8df27;
runtime 0a8c420eb9968aa76a3e9ece91c918a43e3d8c4530773c8818bce5104c0cb416;
ReputationTreeSurface e97971c1a4b00eb3416bcd7933808e94bd68135dc9df7638be76a9bbc14d3125.
No source probe remains. Final gate/browser positives now DEFAULT capture,
including prior61 recovery/reader assertions in the ordinary client suite.

First final gate08658a/7979ef session48840 terminal2:types0 errors/warnings,
213-module build pass; Node8165 pass/313 skips plus two failures in original
61-case recovery population. Exact total-message equality correctly rejects
the newly declared nonterminal notification. No payload/cursor/high-water
failure. Predeclare strict oracle reconciliation: include exactly one recovering
message at the close, require it immediately BEFORE awaiting reconnect, retain
every old payload/order/positions/read-count/stale-socket assertion and compare
the full widened sequence; do not filter the new message or weaken equality.

Final browser219119/session95561 accidentally omitted project selectors and
thus exercised default three-engine launch.067b71 reports408 Chromium/WebKit
functional passes/two performance exclusions,212/212 worker requests finished,
zero pending, but ONE unhandled Firefox browser-session failure at its unchanged
60-second deadline, ZERO Firefox tests. Whole command stays RED, not browser CI
approval. Same handle continued polled; it remained in browser teardown. Fresh
read-only ps533447 identifies owned Make70027→shell70028→Node70029→Nightly70101,
Firefox98%CPU. One-second sample5c7d6a/4a0480 session24065 terminal0 confirms
native main-thread activity; no source/test/record edits while handle live.
Owned Firefox alone gracefully terminatedadfd88 after reported launch failure;
same runnerd0829f terminal2 releases it and prints actual sandbox-extension
denial plus SWGL framebuffer failure, confirming unresolved RP-256. No user
Firefox or other process touched, no security/preference/deadline/config changed.
This is failed launch with hung cleanup, not restart on an observation timeout.
Re-run only the PREDECLARED Chromium/WebKit population, retain this failed
all-engine arm and Firefox's mandatory acceptance hold. Original61 expected
sequences now explicitly include the new internal message, otherwise intact.

Corrected root539c32/c397b5/4cac79 session30828 terminal0:types0 errors/warnings,
213-module build,8167 unit passes/313 skips, all shell/boundary/topology/cosmetic/
no-payment negative controls and copy658/history/generation/manifest checks
pass.313 skips include40 new native-only cases; no failure was newly skipped.
Prior611 orphan warnings remain visible. Restored default-capture selected
browserd7a234/a73e83 session22443 terminal0:408 Chromium/WebKit functional
passes/two performance-selector exclusions,212/212 workers finished, zero
pending. Isolated performance1 pass/22 exclusions,1/1 workers, zero pending.
Whole three-engine attempt remains RED as above; Firefox acceptance NOT proved.
No inherited handle remains live. Fresh declared-Postgres df6b2803 shows root
100%/39,784KiB free and DB tmpfs1%/8,091,276KiB free. No cleanup or SQL proof.

Predeclare one additional source-only omission of the recovering notification
against the corrected FULL Node suite: the two widened exact sequences must
fail before reconnect, with all old payload assertions retained. Native source
omission already fails eight; this confirms the added immediate-order oracles
independently. Restore exact runtime SHA after terminal before records/commit;
then final FULL client run. No relaxed types, filters, deadlines or result claims.

Immediate-message omissione4a129/f2aeff session72563 terminal2:the two exact
recovery tests fail at the newly added close-time equalities (before awaiting
reconnect);8165 other passes/313 skips. Runtime then restored exact SHA above.
Final FULL client7fff3e/9b3cfe session75448 terminal0:8167 pass/313 skips.
No probe or handle remains live. Final type/build/boundary/copy/manifest and
native positives above ran on these byte-identical production sources.

RP-290/RP-291 locally corrected, NOT designated-reviewed or whole AC12. R9
binding adds readiness/existing offline notice; the ordinary network-drop path
adds only internal nonterminal status, preserving subscriptions and prior
delay/history/offset/cursor/full-sync/auth classifications. Existing exact R7
message sequences widen explicitly; all old event payload/dedup/position/read/
stale-socket assertions remain, with immediate-order checks added, not filtered.
Source changes no authoritative balance, purchase/Exit plan or content rules.
Full new range afterb5ca3e7d through its final record edge needs Claude separately
from previous ranges. No checkbox, Go/schema/kernel161/copy658/balance/epoch/CI,
owner/author ruling, mint/report restamp/cleanup/push/deploy/archive changed.
Next accepted grounding: R9 advisory panel under authoritative ownership/budget/
prerequisite replacement; diagnose before inventing an unspecified reselection
policy. RP-283/RP-284 author contracts, actual SQL/two-Exit/mint/AT/Firefox,
RP-131/H3/H4/H5, independent reviews and full nine-tier/platform1.0 remain open.

Final record check3aefea rejected the uncommitted roadmap checkpoint because
its insertion matched an earlier repeated evidence link rather than EOF.
Removed only the newly inserted block and appended it at the actual end;
existing history is untouched. Exact comparison against b5ca3e7d now passes
both logs (1a0e31), whitespace check16ad62 passes. This detector failure is
retained; no previously committed record was rewritten.

## 2026-10-06 — R9 host-readiness implementer first filter

Review by: Codex (implementer self-first-filter).
Recorded by: Codex.
Reviewed range: b5ca3e7d..07e78a95 (120b8cb6, 9a13aebf, 07e78a95).
Verdict: locally approved first filter ONLY; designated Claude pass pending.
This record edge also needs that pass; earlier ranges remain independent.

Inspected the full thirteen-path range, including predeclaration, test-first,
production/docs and all records. The default fixture still auto-acknowledges
subscriptions for the original tests; new cases hold real runtime replies.
No fixture mutation of host state or direct purchase callback substitutes for
native controls. Subscription recovery, held applied/conflict refresh and
terminal closure oracles retain the native delays and original assertions.
R7 message equality expands by exactly one internal recovering notification,
with explicit pre-reconnect equality; old payload/dedup/cursor assertions stay.
Production changes are scoped to Reputation readiness/copy and ordinary-close
host notification; no unsubscribe on transient loss, new wire/auth policy,
reconnect delay/history change, optimistic receipt balance or owner prose.

Executed evidence and failed arms above, not this self-review, support the
local claim: restored full client8167/313;408 selected native plus isolated
performance; type/build/boundary/copy/manifest checks; four native compiling
faults and two exact Node failures on the notification omission. All probes
restore exact SHA38bb9b; no live test handle remains. Whole-range whitespace
d87695 and exact append-only checks da9890 pass. The earlier failed three-
engine attempt remains RED with zero Firefox execution, native copy tooling
and Docker capacity holds unresolved. No full AC12, SQL/mint/manual AT, whole
CI or release approval inferred. No acceptance box/status/archive, kernel,
balance, schema, epoch or copy-content change. Next separately ground accepted
R9 authoritative plan input replacement; full nine-tier/platform goal active.

## 2026-10-06 — R9 authoritative plan-input replacement predeclaration

Fresh clean00335488, previous recovery range b5ca3e7d..00335488 remains ready
for designated Claude review, not archived. R9 plan paragraph, full component,
existing full-tree browser population, actual host snapshot/plan bindings and
existing controlled-prop fixture inspected. Guessed standalone plan/Desk/Offer
filenames do not exist; file discovery resolves actual GameUIApp ownership,
no missing-file claim inferred from that search error. Current selected array
does not reconcile on arm/preview replacement; unknown reselection policy
registered RP-292 as source risk, not executed defect or implementation license.

Predeclare test-only reactive public-prop wrapper and both-era native component
population in plan.md. Separate valid formula/affordability/ownership replacement
from descriptive observations of invalidated selections. Measure stale hidden
owned ids and negative budget, preserve native Clear escape; never encode an
unruled automatic clear/prune policy as acceptance. Three compiling source
faults must falsify the registered formula/ownership dependencies with exact
restoration after terminal results. Original tests/deadlines/security remain;
positive default capture, negative diagnostic no-capture only for RP-289.
No Go/copy/balance/epoch/kernel161/wire/CI/owner-text/checkbox/archive/push or
SQL/mint/full AC12 claim. Full new span after00335488 needs Claude, prior ranges
independent; full nine-tier/platform goal and existing release holds remain.

Controlled wrapper/test added without production changes or old-test edits.
Typecheck afc569/2f3535 session86012 terminal0: zero errors/warnings.
Default-capture7076ef/9a1aab session24318 terminal0: all20 native cases pass
(ten declarations, two eras across Chromium/WebKit); isolated performance1
pass/22 selector exclusions. Component worker0/0, performance1/1, zero pending.
Six functional declarations cover independent formula/ownership replacement
and valid selections; four are explicitly RP-292 CHARACTERIZATION ONLY.
Both eras reproduce hidden owned id retained in callback with double-subtracted
cost, and retained over-budget enabled checkboxes with negative projection.
Native Clear restores empty callback/current projected budget in both cases.
These passing descriptive observations are not desirable-UX acceptance.

RP-292 now executed and routed in the author queue: reset, deterministic pruning
or explicit invalid/refusal state requires reconciled R9 behavior, not guessed
mechanics or new prose. Wrapper exports only public arm/preview replacement;
native events own selection. Controlled component, not actual concurrent host/
SQL/player proof. Source SHA2a856f is unchanged77a4b842f989b29d748195e405e29fefb3fc2a3f0a3a650f23db969b99fd3d6f.
The new ledger row's accidental blank table separator in036fc924 is removed
forward in this range; no committed history rewrite. Next run the predeclared
compiling dependency severings against unchanged tests, restore exact bytes,
and final root/client/native gates. No product correction or policy adoption.

Three predeclared compiling faults, unchanged20-case oracles, all terminal2:
preview omission5328d8/b9cabd session59458:8 fail/12 pass;
available omissionbfab4d/4eba12 session25012:20 fail;
ownership dependency omissionfeb163/f94427 session25653:8 fail/12 pass.
Every component worker0/0, zero pending; no screenshot-copy stall or cancelled
handle. Preview fires after replacement, available also fires initial budget
controls, ownership fires after replacing the public arm. The observations
are not claimed to prove all server or host behavior. Restore exact original
SHA after EACH terminal arm (4afa55,7c2e52 and final check); tests unchanged.
No product diff remains. Proceed to restored default-capture whole Reputation
native population plus original Game UI screen/performance and root gates.

Final restored root7b2760/a79f97/2695fe/17a5d5 session83872 terminal0:
types zero errors/warnings,213-module build,8167 units/323 skips, boundary/
CI-topology13/cosmetic22/no-payment6 negatives, copy658 unchanged SHA and
generation/history checks, deployment manifest pass. Ten new Node skips are
the ten native-only declarations, not disabled browser failures.611 existing
copy orphan warnings remain visible. Read-only ps diagnostic80b39b was sandbox-
denied; no escalation needed or command cancelled, same verification handle
continued to terminal success. No source/test/record edit while a gate was live.

Default-capture restored57127a/3bff1e/898558 session40349 terminal0:
428 Chromium/WebKit functional passes/two isolated-performance selector
exclusions, worker212/212 zero pending; separate performance1 pass/22
exclusions, worker1/1 zero pending. Exact component SHA92c8d0 matches00335488;
f02589 confirms no source diff. All probes and gates terminal, no mutation or
handle remains. Firefox still unproved/earlier all-engine arm RED, RP-289 native
copy and Docker capacity unresolved. No SQL/mint/manual AT/whole AC12/CI claim.

Canonical docs now disclose the measured plan limitation; decision queue names
the author action rather than editing the accepted body or choosing mechanics.
Full new span after00335488 needs Claude through final record edge; prior
b5ca3e7d..00335488 and earlier spans independent. Test-only progress, no product
behavior or owner-copy change, checkbox/epoch/kernel161/balance/migration/
archive/publication/push/deploy/cleanup/restamp. Full nine-tier/platform goal
active; next safe R9 header accounting/frozen-next bonus/published-formula
checks remain independent of RP-292 policy, RP-283 preview and RP-284 display.

## 2026-10-06 — R9 plan-update implementer first filter

Review by: Codex (implementer self-first-filter).
Recorded by: Codex.
Reviewed range:00335488..407f35a7 (036fc924,6d72aa17,407f35a7).
Verdict: locally approved test/record first filter ONLY; not designated review.
This final record edge and the complete range require Claude. The recovery
range b5ca3e7d..00335488 and earlier ranges retain separate obligations.

Inspected all eleven paths. No production/Go/balance/copy/epoch/CI/RFC body
diff; old tests unchanged. New wrapper replaces only public arm/preview props,
not internal selection. Real browser keyboard owns select/Clear. The six
functional declarations check registered formula/ownership updates; four
explicit CHARACTERIZATION ONLY declarations measure the invalidated-plan
limitation, not a desired-policy or release gate. Their later reconciliation
must follow an authored R9 rule; green characterization cannot close RP-292.
No host/concurrent-player/SQL claim follows from supplied component props.

Executed source faults fail8/20/8 with unchanged oracles and terminal exact
restoration. Default restored native428 includes eight RP-292 characterization
executions; isolated performance passes separately. Full client8167/323, root
type/build/boundary/copy/history/manifest evidence recorded above. Original
copy orphan warnings, Firefox/SQL/capacity/tooling and full1.0 holds retained.
Append-only71399a and whitespace4a6023/27abd6 pass; net production-scope check
cb814a is empty. All live handles resolved before record edits; no mutation
remains. No box/status/archival/release promotion. Full goal stays active;
next R9 accounting/frozen-next-factor/formula checks are safe and independent.

## 2026-10-06 — R9 header predeclaration

Previous goal turn is progress: actual recovery correction and executed public-
prop plan observation are committed through2db69792. Fresh clean b7d728 and
de413c, no inherited live handle. AGENTS/process/index and full accepted RFC
reread; bound design sources fully read earlier and unchanged since020a25c6
on their ACTUAL paths (ee2e7b). Initial guessed economy/playstyle/UX and source
filenames were wrong; empty comparisons/search errors on them are NOT evidence.
File discovery resolves actual references and gameui/reputation.go, client
contracts and copy source/generation. Header component/projector/parser,
original producer tests, Go-authored vectors and copy parameter checks inspected.

RP-293 source-contract gap: R9 requires percentage names with integer types;
catalog/generator/component instead pass raw ppm under perlevel/unlock. R2
admits1ppm, fraction0.0001 percent. No unit conversion, parameter rename/type
change or owner prose invented; author reconciliation required. Header formula
placeholder is NOT a published formula. Plan predeclares exact census, native
accounting/factor/null/replacement scope and five compiling faults. Supplied
trees pass actual TS loader; Go-authored factors are independent expectations,
not UI-computed expected values. Controlled component evidence, never live
Founder/Company transaction, server/SQL/mint/manual AT/full AC12/CI approval.
Full new span after2db69792 needs Claude, prior ranges independent. All9-tier/
platform1.0 holds remain; no box/kernel161/balance/copy658/epoch/archive/push.

Actual metadata census ba2e8f terminal1: five balance/factor controls pass,
formula fails with required per_level_percent/unlock_percent versus actual
perlevel/unlock. Added narrowly scoped reproducible
`node planning/reputation-tree-v1/header-copy-contract-check.mjs`:13b4fd
terminal1 reproduces the same five/one. It reads current R9 definitions and
current candidate catalog; no saved green interpretation or alternate catalog.
`--self-test` f0d762 terminal0 is explicitly SYNTHETIC ONLY: six aligned
metadata rows and four individually rejected name/type/missing/duplicate-key
faults. Neither path assesses prose, actual UI parameter units or owner adoption.

Default native header2e4268/dc0258 session86690 terminal0:32 cases pass across
both eras and Chromium/WebKit; separate performance1/22 selector exclusions.
Worker0/0 and1/1, zero pending. Typecheck5754b9/490a79 session80994 terminal0,
zero errors/warnings. Actual TS loader admits the diagnostic1ppm tree; native
formatter rejects0.0001 as an integer parameter. Census stays RED and formula
text is only current raw-placeholder characterization, never R9 formula proof.
All source is unchanged; next commit tests/instrument before compiling faults.
The predeclaration's new ledger-row blank separator is repaired forward to
retain one table, without rewriting committed history. No acceptance change.

Cold existing Go producer/arithmetic population: first command a92818 failed
before Go execution because the Make-expanded regex was not recipe-shell-quoted.
No tests ran in that arm. Correctly quoted b58645/658ea9 session15944 terminal0
uses -count=1 -v, executes exactly three gameui Reputation projection tests
(including admission/refusal subcases) and three reputation bonus tests.
No skip, warm-cache claim or SQL/integration substitution. Go bytes unchanged.

Five compiling unchanged-oracle header faults, each terminal2 before restoration:
balances conflated0232ed/d238f5 session98364:32 fail;
frozen replaced by next4a9d64/42fca9 session3922:20 fail/12 controls;
next replaced by frozen057e35/482f6f session87037:20 fail/12 controls;
null replaced by unit1d75f2/7d4742 session57449:8 fail/24 controls;
formula line omitted9d4f25/03dc9e session39494:32 fail.
Worker0/0, zero pending in every arm; no cancelled handle or copy stall.
Restore original component SHA e97971c1a4b00eb3416bcd7933808e94bd68135dc9df7638be76a9bbc14d3125
after EACH terminal (8abcc1,1ee39d,b53986,41966e and final check).
Formula-line probe proves current placeholder presence ONLY; census remains
RED and does not become percentage/prose acceptance through a passing DOM test.
All sources restored, no product diff; final root/native gates use DEFAULT
capture. Next inspect actual Docker capacity read-only while gates run; do not
silently substitute component proof for blocked real-Postgres verification.

Final restored8055b5/cc2c82/c0fae6 session85818 terminal0: typecheck zero
errors/warnings,213-module build,8167 units/339 reported skips, boundaries/
topology13/cosmetic22/no-payment6 negatives, copy658 unchanged SHA
a5df8920bc1573c8b0543b0c6a3c4c2de955b580626c8e1d130fc9e1adeb0e5e,
generation/history and deployment manifest pass. Sixteen additional Node skips
are the native-only header declarations; no browser failure was disabled.
611 existing copy orphan warnings remain. Restored default native32db48/4f769c
session11249 terminal0:460 selected Chromium/WebKit checks (including eight
RP-292 descriptive executions), two performance-selector exclusions, workers
212/212 zero pending; isolated performance1/22 exclusions, worker1/1 zero
pending. RP-293 actual census still RED; full AC12/Firefox/manual AT/SQL/CI or
owner-prose adoption cannot be inferred. Source1b119c exactly restored and
e4be80 confirms no product diff; all handles terminal before record edits.

Read-only storage550d83: Docker reports2.634GB reclaimable build cache; large
reclaimable images/volumes also exist but ownership/data is not established
and deletion NOT authorized. Fresh postgres dfe0d0c1:root100%/39,784KiB free,
DB tmpfs1%. Builder listdb5f7d distinguishes unrelated tabiya and inactive
release builder; default desktop-linux is the selected cache target only.
Actual du JSONb348b5/9f8cf8:816 reclaimable records,113 unshared, zero in use.
No host-level cleanup, image/container/volume/database deletion or build-policy
change. User asked asynchronously for permission to prune UNUSED BUILD CACHE
ONLY on desktop-linux; accepted question is not an answer/approval. No prune
performed. Existing declared Compose services7e6936 healthy; default image
listing83f3c0 excludes profiled test, not proof its Go image is missing. Actual
Compose file and root test-save-integration alias read before the next scope.

Canonical docs now distinguish valid accounting/factor rendering from the
unimplemented published percentage formula. RP-293 author queue names exact
representation/domain/copy reconciliation without changing another author's
body or choosing prose. All earlier review and nine-tier/platform1.0 holds
remain; full new span after2db69792 needs Claude including final record edge.
No checkbox/Go/schema/kernel161/balance/copy/epoch/mint/report restamp/archive/
publication/deploy/push. Next safe accepted action is exact cached declared-
Postgres preflight and existing R8/AC15 career population if runnable; preserve
any capacity failure and never promote a narrower run to complete SQL proof.

### Header proof first-filter — 2026-10-06

Review by: Codex (implementer self-first-filter).
Recorded by: Codex.
Reviewed range: `2db69792..acfdba0f` (predeclaration, tests/instrument and records).
Verdict: approved as a local test/instrument/record first filter ONLY; not the
designated independent review, R9 formula acceptance or archival authority.
Inspected all eleven changed paths, including the actual census and native
test declarations. Executed evidence and five compiling faults are recorded
above. The unchanged-source positive population passes; RP-293's actual
metadata census remains RED. Synthetic census controls cannot substitute for
the actual catalog, and current raw-ppm placeholder rendering cannot substitute
for published percentage arithmetic or adopted prose. No production, schema,
balance, copy, epoch, CI or RFC-body bytes changed. The independent Go factors
and actual loader's fractional-percent admission remain visible. Both logs
preserve the prior committed prefix; whitespace check passes. Claude must
review the full span after `2db69792`, including this final record edge; earlier
pending ranges remain independent and are not swept into this approval.

Next preflight is read-only: declared profiled Go image/cache availability and
the exact existing R8/AC15 SQL population. Cache-prune question has no user
answer and grants no deletion authority. Neither Docker fullness nor healthy
Postgres alone establishes whether this narrow cached population can execute.

### R8/AC15 SQL grounding — predeclared 2026-10-06

Start `ffd1b673`, clean tree. The declared profiled `golang:1.26` image exists
locally (arm64), Compose owns the existing `cloud-clicker_go-cache`, and its
disposable Postgres service is healthy. No pull, cache deletion or host URL.
Run the existing exact three-test production population through the root
`test-save-integration` target with `-count=1 -v`: purchase persistence/retry/
frozen completeness, Exit-plan atomicity, and complete persisted R5 taxonomy.
Every named test must execute without skip; failures remain evidence and no
timeout, oracle or capacity policy is loosened. This is NOT full CI or AC15.
Source census finds no composed real-SQL scripted-first → elective-with-plan →
run-3 career: existing Exit-plan test seeds run 2 and prior Exit history. R8
corpus/history and Company-run verifier controls are separate populations.
After the existing SQL baseline, predeclare and add the missing composed career
under accepted R8/AC15 if the declared environment executes. No production,
balance/mint/owner-copy or other-author body change; independent review stays
mandatory and all prior holds remain.

SQL baseline: first command b50c08 terminal2 before execution (Make consumes
the regex's dollar anchor and leaves an unmatched quote); no tests ran.
Corrected selector without that anchor04e34e/5a5ce3 session45351 terminal2:
purchase0.51s and Exit-plan0.14s PASS, taxonomy0.13s FAIL at fixture setup,
`epochs_one_current_idx` duplicate current epoch. No taxonomy subcase executed.
The cached declared environment DOES execute SQL despite root-capacity hold;
do not cite capacity as a blanket SQL blocker. Compose's unrelated-service
warning is not cleanup authority. A glob search also failed before reading;
subsequent `rg` against existing paths supplies the actual source census.

RP-294 predeclaration: test-only repair within R5/AC3. The source population
requires both tree and inactive bundles; register their immutable catalogs
and artifacts under ONE disposable fixture epoch's accepted hash set, rather
than creating a second current epoch. Assert exactly one epoch/current epoch
and the complete unique source-hash population. Preserve all21 taxonomy
profiles, eleven applied source controls, all rejection, replay, no-write,
event/outbox/state checks, the real DB constraint and production bytes.
Rerun unchanged three-test population cold; remove the second-bundle binding
temporarily as a compiling negative and require the fixture population oracle
to fail before restoring exactly. No broad shared-helper change. Full new
range afterffd1b673 needs Claude; career and full CI remain separate/unproven.

First repair d33287/be6898 session12191 terminal2 passes the epoch population
check but exposes RP-295: taxonomy service omits mandatory current constants
hash. Purchase and Exit-plan controls pass0.11/0.10s; taxonomy constructor
refuses before subcases. NewService's explicit current/policy preconditions
are correct and stay unchanged. Predeclare fixture-only current tree-bundle
selection from the already admitted profile set, passed via existing option;
inactive profiles retain their original pins and both registered hashes.
No fallback/default/resolver/domain change. Test setup, not player behavior.

277c0f/3059e2 session4663 terminal2 executes all21 subcases, all refuse at
ParseIntent; both other controls pass. RP-296: ParseIntent's canonicalRequest
intentionally removes intent_id for stored/hash identity. The test mistakenly
resubmits that canonical object as incoming wire. Preserve complete parsed
request bytes separately and use them for initial/retry Handle; conflict
envelopes must explicitly restore original intent_id before their own mutation.
Stored canonical payload, hash, full-state/receipt/replay/count oracles stay
byte-unchanged. Predeclare this fixture-envelope correction only; do not
change parser/storage identity or accept missing intent ids. No source body
or balance/migration/epoch release change. All three original REDs stay visible.

Restored fixture baseline beba41/571a96 session4297 terminal0: all three
declared tests PASS, taxonomy executes all21 named subcases with no skips.
Purchase/Exit controls0.11/0.09s, taxonomy0.26s. Source fixture hash, receipts,
canonical-storage/state/event/outbox/no-write/retry/conflict/history oracles
unchanged; incoming wire retained separately. This repairs dormant SQL test
setup, not runtime behavior. Commit test+records before the predeclared missing-
hash-binding fault; no acceptance box/full AC15 or CI claim.

Missing-hash probe8dc949/56c822 session98279 terminal2 fails the exact fixture
population oracle (2 bundles/1 epoch/1 current/1 hash), before any profile.
Restored92c940 SHA4faa45f861df402e070891534b5ca5a53ebb5fd971a06405fc591520c2684905;
baac74 empty test diff. All handles terminal. Earlier missing-current and
missing-wire failures independently demonstrate those constructor/envelope
requirements without changing production. Final broader gates follow the
separate career population, not inferred here.

### Composed Reputation career — predeclared 2026-10-06

Test-only accepted R8/AC15. Use current-content-plus-tree diagnostic bundle,
actual Store/Service/Postgres, declared minigame repository/frozen provider and
one fixture epoch. Seed one valid run-1 genesis at curriculum attendance
threshold, empty owned tree and earned Founder level6: an explicit diagnostic
budget, NOT a measured/default-user/pacing claim or threshold retune. Subsequent
state changes must all be Service.Handle commands; no seeding/replacing run2,
Exit history, frozen factors or new-run genesis after the first seed.

Drive actual scripted first Exit → run2; direct unlock purchase with exact
retry and unchanged current frozen row; accrue/cross the real T0→T1 gate;
elective wind_down with in-plan starter prerequisite → run3. Assert exact Exit
types/count, ownership/accounting, committed plan events, run pins/genesis,
starter inventory and non-unit next frozen row. Replay BOTH completed Company
runs from stored genesis/log/events and the full persisted Founder history.
Run3 remains in progress: execute one real command and independently compare
its logged replay and full persisted head, never call an incomplete run a
completed verifier population. No gameplay mutation outside Handle.

Controls: retained source-bonus vectors, exact receipt retry and current-run
frozen-row equality. Corrupt one existing frozen Reputation factor byte in
copied persisted run2 replay inputs (actual production interval present), and
require the unchanged public Company verifier to return state_divergence;
corrupt Founder head and require Founder state_divergence. Never update
immutable SQL evidence. Exact population assertions prevent a missing run or
log from passing vacuously. If a runtime defect fires, record it before repair
and retain all oracles. No full AC15 checkbox until designated review, other
criteria/platform gates unchanged. No production/balance/mint/copy/CI change.

Career instrument repairs before success:631adc/a3dd2b session59248 terminal2
is a compile failure, not runtime evidence (nonexistent guessed State branch
field). Replaced with actual stored resolved branch;1a5ef1/1efa0c session60158
terminal2 reveals selected_branch is a string, not object. Corrected actual
wire path. f68e5d/2ebabe session86999 terminal2 correctly refuses the test's
dependent-before-prerequisite plan. Reordered the test request to the accepted
topological plan contract, no runtime relaxation or hidden rejection. These
are my test-construction errors, not product defects or accepted evidence.

162c87/991724 session93465 terminal0: composed SQL career passes0.19s,
scripted burnout → direct unlock/exact retry → accrued gate → elective plan →
run3. Both completed Company verifiers and whole three-entry Founder history
verify; copied existing factor byte and copied Founder head independently
return state_divergence. Run3 independent non-unit starter production and full
head/log replay match. Added explicit unit-row/nonempty and exact pin controls
before final gates. No SQL evidence mutated and no state seeded after genesis.
Initial earned6 is diagnostic, not reachability/H1/pacing/minted UI proof.
Next root vet, persisted production integration population and cold Reputation
unit/history/Company-replay population; acceptance boxes remain unchanged.

Restored whole production Integration population27e395/ede0cd session26105
terminal0 in5.921s, verbose output includes the strengthened career and all21
taxonomy subcases, no SQL skip. Root vet75b7ee terminal0. Cold host selected
Reputation/Bonus population5c9664/27020f session3468 terminal0 across production,
reputation and gameui; verbose output was truncated by display, so no exact
host population count is claimed. Host SQL skips are not evidence; actual SQL
population above supplies persistence proof separately. No full CI/history
guard/Linux-amd64/browser/package-union claim.

Additional first-filter probes, predeclared before touching source: temporarily
omit Tree.Purchase's prerequisite rejection and run the exact persisted21-case
taxonomy; then restore source SHA. Separately compute BonusFactor from level-
spent instead of earned level and run the composed career; require failure of
the independent run3 frozen-factor/output criterion, then restore exactly.
Run no source edits while handles live; no committed product mutant, kernel
bump or altered oracle. Final restored root/SQL gates follow both terminals.

Live prerequisite omission4eecdf/073269 session60930 terminal2: two persisted
taxonomy cases fail (wrong applied revision and wrong refusal precedence),
19 controls pass. Exact tree.go SHA restored38e265. Wrong bonus base08476f/
b1a0a6 session87182 terminal2: career fails independently at run3 frozen
factor1e0 instead of1.003e0. Exact tree.go SHA restored748e69:
8c2a4788154768668eb159455a000367228dfc368b75f3ae382b10a7f9d2d291;
5c145d confirms empty source diff. Both compiling source faults used unchanged
test oracles; no mutant committed or live-handle edit.

Final restored SQL44277c/6cae6b session54595 terminal0: whole production
Integration population passes cold in7.872s. Final host265905/2f86e8
session16258 terminal0: production1.098s, reputation0.089s, gameui0.157s;
focused vet passes. Earlier full root vet75b7ee passes. Host SQL skips are
preparation only; verbose actual SQL27e395/ede0cd includes the complete
executed taxonomy and career with no skip. All handles terminal before
records. No full CI/history/Linux/browser/mint/harness or release claim.

Docs and live trackers now distinguish portable corpus from the newly
executed composed SQL career. They remove the stale never-executed taxonomy
claim and name the diagnostic earned6 limit. Initial fixture is seeded once;
all subsequent gameplay comes through Handle. Run3 continuation is explicitly
not a completed Company-verifier population. RP-294–296 are locally corrected,
not designated-approved; previous reviews remain independently pending.
No cleanup/production/kernel161/balance/copy658/epoch/CI/owner-prose/ruling/
checkbox/archive/publish/deploy/push. Cached SQL is runnable despite the
capacity hold; unused-cache question still has no approval. Next accepted
scope is R6/AC9 plan-specific rollback at actual write boundaries, not just
unaffordable prefix refusal. Full1.0 goal remains active with genuine progress.

### Real-SQL supplement first-filter — 2026-10-06

Review by: Codex (implementer self-first-filter).
Recorded by: Codex.
Reviewed range: `ffd1b673..b7f68beb` (six commits; all ten changed paths).
Verdict: approved as local test/fixture/docs/planning first filter ONLY.
Inspected both complete Go test files and the range's predeclarations, retained
failures, canonical docs, ledger and live tracking. The test repair keeps
source-bound canonical/state/receipt/hash oracles; only incoming envelopes and
disposable fixture setup change. The new career seeds once, uses Handle for
later gameplay, reads actual immutable SQL evidence and compares independent
inventory/accounting/factor/production outcomes. Full Founder replay retains
its Fiscal prefixes; Company terminal replay retains plan events and excludes
only those Founder-owned automatic prefixes, per existing replay ownership.
No later-state injection, synthesized receipt/event expectations, selected
failure waiver, copied-evidence database mutation or hidden completed-run claim.
The diagnostic earned6 and unfinished run3 limitations remain explicit.

Executed SQL positives and prerequisite/bonus/hash/factor/head negatives are
recorded above. Exact source restore and empty product diff verified; both
logs preserve their committed prefixes and whitespace check passes. No
production/migration/balance/copy/kernel/epoch/CI or accepted body changed.
Designated Claude review must cover the full span after `ffd1b673`, INCLUDING
this record edge. This first filter cannot archive, close all AC3/AC15 or
approve itself as cross-party evidence. Header `2db69792..ffd1b673` and all
earlier pending ranges remain independent. Next accepted work: predeclare
plan-specific R6/AC9 actual SQL write-boundary rollback; goal remains active.

### R6/AC9 plan write-boundary proof — predeclared 2026-10-06

Previous goal turn: progress (RP-294–296 SQL repair and composed career,
committed through82830dcb). Fresh tree clean, no live inherited handles.
AGENTS/process reread; active index still accepts Reputation. Accepted R6/AC9
and full Store write/fault path inspected; prior full RFC/design readings
remain applicable, no authority file changed. Existing plan test only checks
unaffordable rejection, not faults at applied-plan write boundaries.

Test-only scope after82830dcb: share the existing plan fixture through a
test helper without changing its original population/oracles; add a separate
declared-Postgres test for all14 applied logged-Exit fault stages. Use existing
Store fault injection and the EXACT live service.applyLoggedExit callback.
No new production hook/options. This is persistence-boundary proof, not a new
HTTP/socket/UI/guard/actor population; normal Service.Handle success remains
the end-to-end control. Valid plan includes in-plan cash/tower prerequisite.

Before any logged command, append byte-identical valid diagnostic revisions
2..8 for each initial fixture stream. They are fixture setup, not gameplay
history. Founder genesis begins at the actual revision8; no fake earlier log.
This gives retention actual old revisions to delete, rather than observing a
no-op pruning stage. Full table-row snapshots and decoded latest states must
be unchanged after EACH fired fault, including revisions, events, both logs/
genesis, pins, frozen rows, intent receipts, outbox and verification queue.
Require exact injected sentinel and actual hook firing; an unrelated early
failure cannot count. All14 stages must run in the positive population.

Final normal Handle using the same request must commit exactly once, apply
all plan purchases, freeze non-unit next bonus, prune eligible old revisions,
verify Founder/Company replay and return identical retry with zero new rows.
No immutable evidence update. Predeclare compiling negative: at retention's
injected error, commit before returning it; target only retention subcase and
require full-row rollback oracle failure (not merely wrong error). Restore
exact source SHA after terminal. Final cold whole production SQL Integration,
focused tests/vet and scope/append-only checks. No checkbox/RFC archival or
full AC9 acceptance before designated review; earlier ranges independent.

Initial declared SQL328a6a/5bcde6 session12432 terminal0: original plan control
passes0.10s; new population passes all14 named applied-plan stages0.37s, each
exact sentinel/hook, complete table rows and both decoded heads unchanged.
Normal Handle positive applies all three nodes, freezes1.003, prunes genuinely
eligible old revisions, verifies both replay consumers and retries exactly
without row changes. Revision/genesis anchoring is explicit (first Founder
genesis at8), not a fabricated earlier gameplay log. Twelve tables cover the
Exit path's row writes; SQL sequences are not claimed transactional row state.
Read retry snapshot once rather than twelve times before final gates. Commit
test+records before predeclared retention commit-on-error probe; no product
or acceptance checkbox change and full new span after82830dcb needs Claude.

Compiling retention negative5c1271/54e719 session98179 terminal2: existing
fault branch commits its SQL transaction before returning the exact injected
sentinel. The unchanged full-row oracle rejects changed values in11 tables
(all except save_streams), then decoded Founder head differs. This is actual
rollback discrimination, not a compile error or merely wrong error text.
The narrow `/retention` selector additionally fires the complete-population
guard (1 versus14); that is disclosed, not used as the rollback evidence.
Founder Fatal stops the later Company decoded comparison in this negative;
full table rows already include Company changes. Positive all14 checks both.
No positive control runs after a failed population. Source restored after
terminal: ab6a98 SHA87159fc54645cbb25c2de1286bc7207547c6153ed806a504f22a3a444414b15f,
62f59a empty server/save/exit.go diff. No mutant committed.

Final declared SQLd7c57b/5c9e19 session40207 terminal0: whole production
Integration selector passes cold11.915s,34 top-level and117 subcases, zero
skips/failures, including all14 new stages and previous taxonomy/career.
This is not all packages/fullCI/Linux-amd64/browser/harness/release proof.
Host00fee5/5e8d86 session49652 terminal0 passes production/save/gameui/vet,
but `Reputation|Exit` selects NO reputation-package tests. Corrected0bf514/
efb77a session68646 terminal0 adds `|Bonus`: production1.459s/save0.179s/
reputation0.115s/gameui0.151s and focused vet pass. Host SQL skips remain
preparation only. All handles terminal before docs/record edits.

Full new test and original wrapper diff inspected; old plan oracles unchanged.
Exit insert path inspected: twelve-table snapshot covers writes including
player receipt outbox/queue; no world outbox write in this callback. Docs and
live trackers now record this persistence proof and its limits, not fullAC9.
R6/AC9 source/spec census identifies next accepted scope as request-validation
and Accept Offer coverage, with missing populations to be predeclared first.
One exploratory glob failed without running rg; corrected bounded recursive
search follows. Two guessed source filenames did not exist; actual source
intents.go/prestige.go identified. Neither is evidence of a product failure.
No production/migration/kernel/balance/copy/epoch/CI/owner-body/checkbox/status/
cleanup/archive/publish/deploy/push change. Cache-only permission still has no
answer; cached SQL works despite fullness. Full new span after82830dcb needs
Claude including final edge; earlier reviews independent. Proper1.0 goal active.

### Applied-plan SQL supplement first-filter — 2026-10-06

Review by: Codex (implementer self-first-filter).
Recorded by: Codex.
Reviewed range: `82830dcb..8531832a` (three commits; all nine changed paths).
Verdict: approved as local test/docs/planning first filter ONLY.
Inspected full new test, original fixture wrapper diff, predeclaration and
record/doc changes. Original population is retained; separate fault mode uses
exact live callback/existing Store hook, not a test substitute. Every error
requires hook and exact sentinel, full rows/head comparisons and all14 count.
Diagnostic old revisions precede Founder genesis; retention success proves
real pruning. Normal Handle control/replay/retry binds live persistence, not
an assumed receipt. Source commit-on-error negative fires row oracles, with
its narrow-selector guard separately disclosed; exact restore verified.
Complete cold SQL and corrected host/vet evidence retained; original narrower
selector is not misrepresented. Product diff empty, whitespace clean, both
logs append-only relative82830dcb. No new actor/guard/route/default-player/
pacing/fullAC9/CI/release claim or stale authority promotion.

Designated Claude review must cover full span after `82830dcb`, INCLUDING
this first-filter edge. Earlier SQL/header/browser/runtime and other spans
remain independent. This record cannot approve its own cross-party gate or
archive. Next accepted R6 request/offer census; goal remains active/progress.

### R6 request boundary — predeclared 2026-10-06

Fresh clean6dc16a58, no live handles. Accepted R6 and AC9 reread; actual
intents.go canonical/parser path inspected. Server-wide test census confirms
RP-297: no dedicated size/type/uniqueness/syntax/order/hash identity population.
Portable helper can configure an offer, but none of its Reputation cases does
so; actual SQL plan/career/write-fault populations use Wind Down. Those are
distinct gaps, not evidence that production is wrong. Log/ledger immediately.

Test-only paired Wind Down/Accept Offer shape cases and literal canonical/hash
oracles; absent versus empty semantics/identity, ordered/reversed plans,
64 accepted/65 refused, non-arrays/null/mixed elements/invalid IDs/duplicates.
Extra field and invalid Founder revision refuse without confusing plan syntax.
IPO/decline/scripted-first cannot carry plan. Canonical identity retains both
revisions/Offer ID/order/key presence, excludes intent ID, ignores JSON key
order/whitespace. Syntactic node IDs are not live-tree membership/affordability.
No fixtures/golden/copy/epoch/RFC rule rewritten.

Predeclare compiling negatives: bound64→65; remove duplicate; remove ID regex;
drop reputation_plan in canonicalRequest; reverse parsed plan; permit IPO plan
parsing. Run unchanged tests cold after each, restore exact SHA and empty
production diff before next. Then cold focused parser/Reputation/Exit/Bonus/
vet and whole production SQL Integration. No edits while test handles live.
No tests-skipped/compile-error as discriminating runtime success. Full range
after6dc16a58 needs Claude including record edges; all earlier spans independent.

Initial request populationd61ab0/07527d session27465 terminal0: three top-level
tests and59 reported subcases pass cold0.356s. Both plan-bearing commands run
all20 shape cases. Literal canonical bytes and independent SHA-256 bind exact
key presence/order/revisions/Offer ID; JSON key order and whitespace and intent
ID equivalence are explicit controls. Forbidden commands and unrelated closed-
field/Founder revision checks retain exact refusal detail.64 syntactic IDs
are deliberately not an admitted/registered64-node tree. Test-only commit
before compiling source probes; no acceptance checkbox or production change.

All six compiling source probes use unchanged complete request population:
bound eb2627/3778c4 session87751 terminal2 fails two65 cases/57 controls;
duplicate8236f3/1a04cb session56825 terminal2 fails2/57 controls;
syntax f3526d/291b9c session83337 terminal2 fails12/47 controls (including
JSON null element decoding as invalid empty ID, not a silently accepted node);
canonical-plan omission8a7e3c/229211 session97629 terminal2 fails40 reported
subcases/12 controls; identity parent failures prevent its seven nested field
controls in that negative, disclosed rather than counted as passes;
parsed reversal2030bb/1bfa82 session92285 terminal2 fails6 shapes/53 controls;
IPO admission29e8fe/a8a148 session79724 terminal2 fails2/57 controls.
Every negative is a real runtime assertion failure, not compilation or skip.
Direct order vectors catch reversal while the separate inequality identity
control survives (both orderings reversed are still different); it is not
cited as the reversal oracle. Exact SHA restored and empty source diff after
each terminal arm:35fe7a/df9ea8,14742d/0ec7c6,26fd84/425e7a,ee49f5/b2cd85,
291b74/0517dc,179e1f/fe4b96. SHA remains
a0c5fb975771c6f4c4b5c50ea8ee67ebaedb85348bcfdc9561808fd9546ae955.

Final restored cold04add8/1e30a6 session64713 terminal0: production1.072s,
save0.325s/reputation0.201s/gameui0.264s, focused vet passes. Selected
Reputation/Exit/Bonus/ParseIntent; host SQL skips are preparation only.
Declared SQL2fb5a2/c3b496/a73f96 session30756 terminal0 passes production
Integration7.132s,34 top-level/117 subcases/no skips/failures. Compose reports
an existing game-ui-postgres orphan; no cleanup/ownership inference or
--remove-orphans. All handles terminal before records. Source and product
diff empty; no migration/kernel/balance/copy/epoch/RFC/CI or checkbox change.

Docs/ledger/live board/queue reconcile request proof versus actual offer-plan
SQL absence. RP-297 locally supplemented, not designated-approved; syntax64
IDs do not imply registered/affordable64 nodes. Full new span after6dc16a58
needs Claude through records; earlier independent spans pending. All prior
author/owner/Firefox/history/capacity/harness/mint/manual AT/clean-host/release
holds remain. No cleanup/archive/publish/deploy/push; proper1.0 goal active.

Record-placement correction: first-filter's actual zero-context diff shows
d51a95e4 inserted the new roadmap checkpoint after an older identical evidence
link, not EOF. Own new entry relocated in a forward correction; preexisting
log bytes are restored as the exact prefix relative6dc16a58. No historical
entry rewritten, no hash amendment. This is a process slip caught by inspection,
not an assertion that the original commit was append-only. Exact final range
must include the correction and final first-filter edge.

### R6 request supplement first-filter — 2026-10-06

Review by: Codex (implementer self-first-filter).
Recorded by: Codex.
Reviewed range: `6dc16a58..2d97c7ec` (four commits; all nine changed paths).
Verdict: approved as local test/docs/planning first filter ONLY.
Inspected full208-line test and all ledger/docs/plan/live records. Independent
literal canonical bytes bind field order/presence; independent SHA expectation
does not call the production helper. Pairing covers both live plan commands,
and closed arms reject. Runtime lookup/affordability is deliberately outside
syntax population. Reversal vectors, not merely inequality, catch wrong parsed
order. All six compiling negatives fire with controls retained and exact source
restoration. Nested controls absent from hash negative are not green-labelled.
Focused cold and declared SQL populations run; no fullCI or browser claim.

Actual diff caught and forward-corrected d51a95e4's misplaced new roadmap
entry. Final both logs retain exact committed prefixes relative6dc16a58;
whitespace and empty production/migration/balance/copy/epoch/CI/RFC diff checks
pass. No acceptance checkbox/status promotion. No authority or policy was
invented to make green tests. Designated Claude review must cover full span
after6dc16a58 INCLUDING this edge and record correction; earlier independent
spans remain pending. This first-filter cannot archive or approve itself as
cross-party review. Next accepted scope: separately predeclare real SQL
Accept Offer with plan/replay population. Proper1.0 goal remains active/progress.

### R6 Accept Offer with payout-funded plan — predeclared 2026-10-06

Previous goal turn progress (all14 applied-plan faults, RP-297 request proof),
clean8ee45af3 and no live inherited handles. AGENTS/process fully reread,
active index still accepts Reputation; unchanged full RFC/design readings
remain applicable and R6/R8/AC9 specific clauses reread. Actual offer producer,
plan dry-run/apply, logged-Exit refusal and both replay consumers inspected.
No new skill/delegation required. Current tracking correctly marks missing
offer-plan SQL as next; parser proof and configured-but-unused portable helper
do not substitute. Earlier SQL faults only used already-earned diagnostic6.

Test-only population per plan above: two deterministic offer kinds, initial
zero Reputation, live gate-generated offer, payout-funded plan6, unaffordable
last-entry refusal preserving heads/offer/game tables and logged exact retry.
Initial tier2/run2/cash1e9/lifetime8000×threshold/prior-history is explicitly
diagnostic, not natural progression or OD-2 pacing. Offer kind fixture selection
uses existing bounded deterministic setup search; both kinds must execute,
none silently omitted. All later commands use Handle, no later head/offer
injection or direct immutable evidence edits. Pinned payout modifiers must
match independent18/20 controls, not derive expectations from ComputeTerms.
Full current-run/verifier history, ordered events, run3 state/pin/summary,
exact retries and changed-plan same-ID conflict are required.

Predeclare source negatives: disable actual offer generation; validate plan
before including credited payout; change accepted-resolution payload; skip
recorded-hash conflict check. Copied Founder purchase cost and Company canonical
plan entry/order must independently return state_divergence without changing
SQL evidence. Restore exact source hash and empty product diff after each
terminal; then cold full production SQL Integration/focused/vet and docs/live
tracking/append-only reconciliation. No production hooks/rules/epochs/copy/
balance/kernel/CI or acceptance checkbox change. Full new span after8ee45af3
needs Claude including records; earlier designated review gates independent.

Initial declared SQL931334/807ff6 session22010 terminal2 does not execute:
my new test compared plain string to typed save.EventKind, a compile error.
Correct test scan/vector to the actual EventKind type; no production defect
or failing-case evidence inferred. Temporary retry helper was completed before
this compile; no placeholder lands. Whole test is new, old tests untouched.

Next178919/669ea8 session77256 terminal2 executes both real offer producers
and returns the exact expected unaffordable rejection, but my new receipt
helper expected flat category/detail rather than the established nested
`rejection` object. Correct helper to actual wire shape, retain exact outcome/
category/detail and require absent rejection for applied. This is test
construction, not a wire or product correction. No weaker substring oracle.

Nextdb3904/5c2f5e session49211 terminal2 reaches applied state in both kinds,
but my owned-ID assertion conflates request purchase order with R1's byte-
sorted owned set. Added diagnostics only,3b904e/84d4d6 session67555 terminal2:
all revisions/earned18or20/spent6/unlock50000/Exitkind/run3/offer cleared/
generated5/purchased0 match; owned set is correctly cash,tower,unlock, not
unlock,cash,tower. Correct independent sorted literal; keep purchase/event
order assertion unchanged. These initial construction failures stay recorded,
not labelled production defects or demonstrated severings.

Initial complete positive0d58ac/8b3036 session16802 terminal0 passes both
offer kinds0.317s: real gate producer, zero initial Reputation, independent
payout18/20, last-entry rejection with full heads preserved, valid plan6,
cash1e3/generated5/purchased0/non-unit frozen1.009or1.01, ordered Founder
purchases, accepted offer resolved before ended, run_started v2 summary,
three-entry completed Company replay and rejected+applied Founder history.
Both copied cost/order corruptions return state_divergence, both exact retries
and same-ID changed-plan conflict preserve complete table values. No SQL skip.

Before source probes, strengthen independent applied receipt head-revision/
count oracle and explicitly compare complete SQL rows before/after copied
evidence negatives. No old assertion removed, production hook added or budget
changed. Re-run complete positive before test commit. Bounded fixture search
uses existing offerFixtureFounder ceiling1,000,000; no subset quietly excluded
and no gameplay population/pacing/statistical inference from selected IDs.

Strengthened positivefece30/b307eb session76533 terminal0 passes both kinds
0.316s, including receipt committed-head revisions/count and explicit SQL
immutability around copied negatives. Commit test+records before predeclared
source probes. No production change or acceptance checkbox; full new span
after8ee45af3 needs Claude independently of earlier request/rollback spans.

First altered-resolution probe da2088/6955fb session84092 terminal2:
`resolution:declined` compiles but the existing save payload validator refuses
it before applied/event assertions, both kinds. Defense-in-depth failure,
not evidence that the new event oracle itself fired. Exact source restored
(recorded with forthcoming complete probe evidence). Refine the same declared
altered-resolution family to a different VALID UUIDv7 offer_id while retaining
resolution:accepted; require the new stored identity oracle to fail. No
expectation/control/production acceptance set change, new fixture or waiver.

Executed live source faults, all compiling against unchanged test8a4b0784:
producer c24b3f/a1d454 session61774 terminal2 fails both cases at actual gate-
offer presence; exact prestige restore6bd3fa/8c22a1.
prospective-credit omission4995d4/7f78b6 session98604 terminal2 keeps initial
unaffordable refusal/retry controls but rejects valid plan because it ignores
new payout, both kinds; restoref09e05/dfd910.
invalid resolution literal da2088/6955fb session84092 terminal2 fails earlier
save validator, not new stored-event oracle; restore2eba03/64e2dd.
valid wrong offer-ID refinement ca4c50/835078 session47240 terminal2 reaches
the committed valid event and fails both exact identity assertions: expected
offer-ID query has no matching value, NULL scan is the oracle's missing-row
failure, not connection/schema failure; restore43f843/33e8af.
recorded-hash bypass9e7bf7/d15df0 session83173 terminal2 reaches final changed-
plan same-ID attempt; old applied receipt/replay=true violates required
idempotency_conflict, both kinds. Final source restoreb6ca3c/b22493.
Prestige SHA6d07581c74822e57d63c09b0ed139d98b0b1f44aba566f8d654ad2f4213cb691;
Store SHA87159fc54645cbb25c2de1286bc7207547c6153ed806a504f22a3a444414b15f.
All restores exact/empty diff, no mutant committed or edit while handle live.

Final cold9f6859/fd1485/f8d5aa session53729 terminal0 runs COMPLETE relevant
host packages (no selector): production35.080s/save0.461s/reputation0.263s/
gameui0.403s; focused vet passes. Host integration skips are preparation, not
SQL evidence. Declared SQL6883f2/831325 session32275 terminal0 executes entire
production Integration selector8.245s,35 top-level/119 subcases/no skips or
failures. Compose orphan notice retained, no --remove-orphans or cleanup.
Client baseline ef8a5a/88cf2c session13669 terminal0:92 files/8167 tests pass,
22files/339 cases skip,4.91s. Node browser skips are not browser proof. No
new offered-plan TS fixture is claimed by those existing units; source/test
census4ce165/8d10d0/950ff9 shows current shared fiveExit/threeFounder corpus
has WindDown/activation, not this offer population. R8 separate shared Go/TS
offered-plan parity is the next safe accepted scope, predeclare first and
preserve existing historical bytes/authority. No generation or fixture edit
made during this census.

All handles terminal before record edits. New357-line test inspected in full;
no old test weakened. Independent fixed payout18/20 and factor1.009/1.01,
byte-sorted owned set versus ordered plan, exact applied receipt revisions,
rejection and full persisted/replay outcomes remain bound. No blind approve
of test-construction failures. Docs/ledger/live trackers reconcile actual SQL
proof and missing cross-runtime boundary. Full span after8ee45af3 needs Claude,
earlier spans independent; no fullAC9/RFC/CI/browser/pacing/mint/release claim.
No production/migration/kernel/balance/copy/epoch/owner-body/CI/checkbox/status/
cleanup/archive/publish/deploy/push. Proper1.0 goal active/progress, all earlier
author/environment/release holds retained and cache-only approval unanswered.

### Offer-plan SQL supplement — local first-filter 2026-10-06

Review by: Codex (implementer self-first-filter)
Recorded by: Codex
Reviewed range: 8ee45af3..71029ba2 (three commits, nine paths).
Verdict: locally approved; NOT designated cross-party approval.

Full new357-line test, predeclaration, canonical docs, ledger and live tracking
reviewed. Actual gate producer rather than seeded offer; initial earned0 versus
independent credited18/20; ordered plan versus byte-sorted ownership; exact
receipt fields against persisted heads; saved resolution/purchases/run summary;
both complete replay histories; refusal/exact retries/changed-plan conflict;
copied negative evidence with unchanged SQL all remain independently asserted.
Initial compile/wire-shape/owned-order construction mistakes remain disclosed,
not production defect claims. Invalid resolution fault hits the earlier save
validator; valid wrong-ID refinement reaches the new persisted identity oracle.
Producer, prospective-payout and hash-conflict source probes discriminate as
recorded. All source restores exact, no mutant committed, all handles terminal.

Final relevant complete host packages/vet, declared SQL35/119/no skips and
client8167 pass baseline retain their stated limits. No offered-plan TypeScript,
browser/default progression, full CI or release proof inferred. Fresh checks:
8be514 range statistics agree; 41f8a5 whitespace clean; 5b1b3e production/
migration/balance/copy/epoch/CI/RFC/testdata diff empty. Both append-only logs
retain original prefixes (44b60c/11d1a5). No acceptance checkbox/status changed.

Claude's designated review must cover the entire new span after8ee45af3,
including this record edge; earlier pending ranges remain independent. This
self-first-filter cannot archive or relabel itself cross-party. Next accepted
work is separately predeclared R8 shared Go/TS offered-plan parity, preserving
existing historical corpus bytes. Proper1.0 goal remains active/progress.

### R8 offered-plan parity — predeclared 2026-10-06

Previous goal turn progress: offer-plan SQL supplement committed through
c3ab42d0, clean tree/no live handles. AGENTS/process reread; accepted index,
R6/R8/AC4/AC9 and design vision/tech revalidated. Existing historical shared
corpus is20 purchases/5 WindDown-or-activation Exits/3 Founder arms. Actual
Go helper supports stored offers and TypeScript replay contains promise and
prospective-credit paths, but declarations alone prove no parity population.

Implement the ten-case supplement above, no existing corpus rewrite. Use one
diagnostic pinned bundle and explicit stored promises18/20, initial earned0,
plan6/refusal23/absent/empty/promise-floor for each offer kind. Add explicit
generation-only root target, independent fixed state/event/receipt controls,
both runtime byte comparisons and copied-input negatives before source faults.
Compiling runtime mutations must fail unchanged tests; generation must not
replace expectations on failing controls. Restore exact source bytes before
final full relevant packages/vet/client/type gates and tracking reconciliation.
No SQL/default player/browser/pacing/mint proof inferred; earlier holds remain.
No delegation/new skill or source mechanic change authorized. Full new span
afterc3ab42d0 including records requires Claude's designated pass; no archival.

Initial generator4e464d terminal2 fails compilation: my new test called a
nonexistent decimal.FromInt64 helper. Replace with existing exact Decimal.New
(8,3), preserving independent8000 population. No fixture generated, no runtime
execution, product defect or discrimination claim from this construction error.

Next ebfd10/98e632 session26281 terminal2 executes but my independent event
kind literal omitted the established `.v1` suffix. Correct to exact existing
EventReputationNodePurchased and separately assert schema1. No fixture written;
no production defect claim or relaxation of ordered id/cost/source controls.

Adding copied plan-order control exposed another test-only helper assumption:
e61b15 terminal2, nonexistent mustMarshal. Use explicit checked json.Marshal;
generator again did not execute or write. All construction failures retained.

Generator79af50/4467ac session60153 terminal0 emits ten-case supplement0.287s
after independent controls. First replay52b5a0/ac2bf8 session71432 terminal2:
Go byte regeneration passes, copied cost/promise refusals pass, but copied
request order calls external ParseIntent with stored canonical bytes (which
omit transport intent_id), so all four order arms fail before replay. Add the
recorded command identity to copied external request, then ParseIntent strips
it as usual; no original canonical fixture or runtime changed. Client baseline
9c494b/c330e3 session43027 terminal0:8184 pass/339 skips, including all17 new
offered replay tests. Browser skips remain unexecuted, not acceptance proof.

Corrected Go037271/eef8a6 session53733 terminal0: corpus and allfour copied-
negative arms pass0.316s. Typecheck91a97e/c6aec5 session60711 terminal2 catches
my new test passing optional ReplayCatalogBundle fields where concrete meter/
achievement catalogs are required. Add explicit missing-catalog refusal and
typed restore helper, no cast-away or runtime source change. Existing historical
Reputation corpus diff da6e7b empty; new fixtureSHA0cc495bf7170d391fbd763719b1782c183097d689220655035db0101f35787bc.

Restored test helper positive1f99cb/6beed1 session70090 terminal0: strict TS
and Svelte0errors/0warnings. Full client b6b63c/1a472e session61643 terminal0
8184 pass/339 skip5.14s, including17 new parity tests. Go baseline above stays
green; no live handles before commit/source faults. Generation compares exact
ten-case bytes, eight paired applied Founder arms, independent refused state/
events and payout/starter/bonus/order controls. Both runtime copied cost/promise/
prerequisite-order negatives run for allfour planned applied cases. Existing
corpus preserved. Commit tests/generation lane before unchanged-test source
probes; initial scaffold/type errors remain explicitly retained, not sourcebugs.

Executed source faults against committed abc019c4: first7cd880/62d7d0
session80698 terminal0 ran NO tests because a single `$` anchor was consumed
by Make expansion. This is invalid evidence, not a pass. Correct selector
d2adc7 terminal2 executes and catches omitted prospective payout at the first
acquihire applied oracle. Serial builder stops there; remaining kinds/arms
were not executed in this negative. Actual generator cf8d1e terminal2 fails
the same independent control BEFORE writing; f514c7 fixture SHA unchanged.
Exact Go restore6b6d2a/f06c2f. TS cost+1 in actual plan event producer:
63c08b/7998a5 session16638 terminal2, four new exact ordered-event byte
comparisons fail, plus two existing historical plan comparisons;8178 pass/
339 skip. This proves new byte oracle discrimination, not just earlier guards.
Exact TS restore e7df24/8e97ee; no source mutation committed/live handle left.

Further actual Founder audit census4c2b87/df8a42 finds a small coverage gap
in my initial supplement: rejected offers also have an exit.v1 Founder audit
arm, while the reused fixture helper intentionally stops before emitting it.
Do not call eight applied arms full paired parity. Predeclare additive extension
now: keep same ten Company cases, add both rejected Founder arms via actual
buildFounderExitAudit/ApplyFounderLogged, compare exact rejection receipt,
unchanged full Founder state, zero events and same input result pin in Go/TS.
Population becomes ten Company/ten Founder, eight applied/two rejected on each.
Generate only the NEW supplement after independent rejection controls; preserve
original corpus. Add copied rejected Reputation-delta1 refusal for both runtimes.
This is accepted R8 evidence extension, no rejection policy/runtime change.

Refinement generator856b21/b43652 session14921 terminal0 emits ten Company/
ten Founder arms0.379s. a65ee4 terminal0 runs normal compare and allsix copied
negative groups0.231s: four planned arms plus both rejected delta arms. Full
type/client384f5e/2dc715 session20028 terminal0, TS/Svelte0errors0warnings,
8186 pass/339skip4.15s (19 new tests). SupplementSHA d062a758c917aafc45ac9d149722dc7bd4ea6ee906f5f536df071298cc30e50e.
Original corpus preserved. Commit extension before final source faults.
Predeclare additional exact Go rejected-Founder delta-zero guard omission:
both copied delta1 negatives must fail, unmodified positive rejected records
remain accepted. This is input refusal discrimination, not a SQL mutation.
Re-run prospective-credit/generator and TS cost faults against final tests,
restore every source exactly before final complete relevant packages/vet/
type/client baselines. No expectation regenerated on a fault, no sourcecommit.

Final faults against b03ca9d4: generator7c3e1b/0311cc session67328 terminal2
catches prospective-credit omission at first serial acquihire applied arm;
later arms unexecuted in this negative, not silently claimed. d8e001 confirms
supplementSHA d062a758c917aafc45ac9d149722dc7bd4ea6ee906f5f536df071298cc30e50e
unchanged. Go rejected-delta omission0bb8c4/a2715b session51300 terminal2:
both new rejected-Founder delta controls fail, four other copied groups pass.
Exact Go restores38ca44/675b80: Prestige6d07581c74822e57d63c09b0ed139d98b0b1f44aba566f8d654ad2f4213cb691;
Founder replaybcf482e3fab2bd4b8d154ad6c8eb214ceb465da299b03582829fd44e8827e7f5.
Final TS cost+1 b5ffcc/069574 session47046 terminal2: four new exact byte
comparisons fail at ordered Company-log Founder events, plus two older plan
comparisons;8180 pass/339 skip. These tests stop before later paired Founder
comparisons in the four failed cases; no later comparison failure inferred.
Exact TS restorea9e838/db0ccc SHA1790fb96006b721372a3cb5a596225708754e271ab972bb7631851652cb54ee7.
No production diff, mutant commit or fixture write under source faults.

Final restored complete host package union b4c6aa/23604d/12d293/c29bf2
session69922 terminal0: production37.045s/save0.374s/reputation0.195s/gameui
0.296s and focused vet pass. Host integration skips are preparation, not SQL
proof. Declared SQL0378fb/37b9f7 session88557 terminal0,10.269s: entire
production Integration selector35 top-level/119 subcases/no skips. Existing
Compose orphan warning retained; no cleanup or ownership inference. Strict
TS/Svelte plus full client353112/a5671c session57762 terminal0:0errors0warnings,
8186 pass/339 skips4.72s. No native browser or fullCI/kernel-history/1.0 claim.
All handles terminal before docs/tracker edits; full339-line Go test inspected,
142-line TS test and deterministic fixture producer/consumer controls reviewed.

RP-298 ledger, canonical docs, current state, executable queue, per-RFC plan
and full1.0 board/log reconcile the actual offered replay proof. Preserve all
historical corpus bytes and previous live SQL/request/rollback obligations.
Next safe accepted step: census remaining R6 refusal and next-bundle activation
populations before predeclaring gaps. Full new span afterc3ab42d0 including
records requires Claude; original implementation and earlier supplements remain
independently pending. No production/schema/migration/balance/copy/epoch/mint/
kernel/CI/checkbox/owner-body/cleanup/archive/publish/deploy/push change.
Previous goal turn progress; this turn progress, proper full1.0 goal active.
All prior owner/author/environment/release holds and unanswered cache request
remain; a running cached SQL service does not erase the capacity problem.

### Offered-plan replay supplement — local first-filter 2026-10-06

Review by: Codex (implementer self-first-filter)
Recorded by: Codex
Reviewed range: c3ab42d0..a7ce3f7a (five commits, twelve paths).
Verdict: locally approved; NOT designated cross-party approval.

Read full339-line Go producer/negative test and142-line TS consumer, all scope
declarations/refinement, Make generation-only lane and docs/ledger/live records.
Normal Go compares the entire generated9020-line fixture byte-for-byte; TS
compares ten Company and ten Founder arms, eight applied/two rejected each,
complete receipts/state/events/pins. Independent fixed payout/starter/bonus/
ownership/offer/refusal controls survive source restoration. Initial scaffold,
canonical-vs-external-request and strict typing errors remain disclosed. Initial
no-tests selector is rejected as evidence. Rejected-Founder arms were genuinely
added after census/predeclaration, not retroactively claimed by eight-arm work.

Prospective-credit/generation, TS cost and rejected-Founder delta source faults
discriminate as recorded; exact hashes restore. Serial first-case stopping and
later Founder comparisons not reached by TS fault are explicitly bounded.
Full relevant cold Go/vet, strict TS/Svelte,8186 client pass/339 skips and real
SQL35/119/no skips retain their scope; browser/fullCI/release remain unproved.
Original shared corpus unchanged9cd79e; production/migration/balance/copy/epoch/
kernel/CI/RFC diff empty7522b4; whitespace clean88a503. Append-only Reputation
and roadmap logs preserve committed prefixes (ee4cf9/c26048). Clean7f3592.
No acceptance checkbox/status/owner copy change or source defect claim.

Claude's designated pass must cover full new span afterc3ab42d0 INCLUDING
this record edge; earlier independent implementation/supplement spans remain
pending. This first-filter never archives or labels itself cross-party. Next
accepted work: remaining R6 refusal/next-bundle activation census, then
predeclare missing evidence. Proper full1.0 goal remains active/progress.

### R6/R7 persisted activation/refusals — predeclared 2026-10-06

Previous goal turn progress (offered parity through e5a69731), clean/no live
handles. Accepted R6/R7/R8/AC9/AC11 revalidated against actual next-bundle
resolver, dry-run/apply, seed/pin/epoch integrity and both history consumers.
Pure/shared WindDown activation exists, and diagnostic same-tree SQL offer/
career/rollback evidence exists. No dedicated Reputation persisted old v21
pin→new tree boundary or full plan-prefix refusal matrix found in actual
production populations. RP-299 registered immediately, not a claimed defect.

Use24 profiles above, three commands × eight arms. Nine activation positives
with mid-run inactive purchase control, fifteen explicit whole-Exit refusals,
all exact retries. Old epoch closes normally and next fixture epoch appends
only in disposable DB; no same-epoch artifact addition/minted source edits.
Initial earned6/stored promises0/run2/tier3 are diagnostic, not naturally earned
progression. Actual saved heads/pins/frozen/events/logs and public Founder/
completed Company verifiers are mandatory; refused current run remains
nonterminal, compared directly instead of promoted to completed verification.
Compiling current-vs-next, skipped application and refusal-detail source faults
must fire unchanged tests; exact restores and broad cold gates before closeout.
No new skill/delegation, product rules/copy/epochs/CI or acceptance boxes.
Full span aftere5a69731 including final records needs Claude; earlier holds
and full proper1.0 goal remain active. Next pure cross-pin Go/TS refusal census
stays separate from this real-Postgres proof, not silently claimed by it.

Initial declared SQL8bf826/e5e18e session62868 terminal2 fails compilation:
unused fmt import in my new test. No cases execute and no source/SQL defect
inferred. Remove it; before first executed baseline, strengthen exact applied
receipt revisions, both-head mid-run invariance, full next-Company replay/head
equality, v2 starter summary and independently ordered Founder purchase events.
No production edit or weakened expectation/acceptance bound.

First executed fa841f/964eda session84538 terminal2 reaches all24 cases:
nine activation and all five WindDown refusal cases pass; ten offer refusals
fail the final recorded Company replay compound oracle with nil error. Earlier
exact taxonomy/persisted heads/game rows/public Founder history pass in those
cases; final completed counters truthfully9applied/5refused, not24green.
Add component diagnostics only to isolate outcome/receipt/events/full-state
difference before deciding test-format issue versus product defect. No weakened
oracle, silent omission or runtime edit authorized by this red observation.

Diagnostics59aa04/780e6f session9118 and f0796c/14f4fb session19394 terminal2
each run only the selected offer refusal;24-case population guard also fires
(executed1), not a complete run. Outcome/rejection receipt/zero events all
match. Raw TermsJSON differs only PostgreSQL jsonb spaces/key order versus
typed codec order; exact-number full JSON canonicalization is equal. This is
my replay-state formatting oracle, not a confirmed product defect. Change only
this replay comparison to complete canonical state values (UseNumber), retain
raw full persisted head/row before-after comparisons, and add copied tier+1
negative in every refusal. No field omitted/rounded/defaulted, runtime change,
weak substring comparison or acceptance threshold waiver.

Corrected full SQLfd811c/9dfc81 session78784 terminal0: all24 execute/pass,
nine applied/fifteen refused,1.620s. Nine mid-run inactive purchase controls
preserve both full heads under the newer available tree. Actual two-epoch
activation preserves old run pin, installs new22/18 heads/new run3 next pin,
unit or1.003 frozen factor, starters, ordered Founder events and summary;
completed old-pin Company run/full Founder history verify. Wrong current-hash
substitution refuses both consumers. Fifteen first-failure prefixes preserve
full heads/game rows/offer and record both logs/receipt/outbox; direct Company
refusal replay/full canonical head and copied-tier negatives pass. All24 exact
retries preserve twelve full tables. Diagnostic initial earned6/stored offers,
not naturally earned progression/default browser/minted-release proof.
Commit test before the predeclared unchanged-test production source probes.

Source probes against43999875, full24-case matrix each, all compiling:
current-bundle instead of next e0626d/9603ce session12809 terminal2 rejects
allthree activation-plan cases with tree_inactive; six other activation and
fifteen refusals pass, population guard9→6. Restore selector, isolate subsequent
skip-application diff7b0016 (no residual first fault). 5d6dc7/0a5726 session31171
terminal2 fails allthree planned activations at EXISTING live Founder Exit
parity guard before new saved-accounting oracle. Earlier defense-in-depth, not
new-oracle discrimination. Prestige exact restore162c44/2811b6.
Wrapped refusal-detail suffix713d2a/6b736f session78987 terminal2 fails all12
unknown/owned/requires/unaffordable exact details; nine positives/three inactive
refusals pass, population guard15→3. Restore exact originals before refinement.

Predeclare same post-plan accounting family refinement: keep application/events/
purchase audit valid but reset Spent to0 after shared applyReputationPlan finishes.
Both live and Founder replay consume that shared helper, so the existing live-
parity guard may agree with the wrong state. New independent persisted spent6
oracle must fail the three activation-plan cases; other21 controls should pass.
This is test-evidence faulting only, not a product accounting change. No tests,
data, bounds or prior bad-result record modified. Restore exact source before
broader cold baseline and closeout; full span still needs Claude.

Refined shared-helper spent reset275b89/55bdbc session83421 terminal2 reaches
saved state and fails allthree independent spent6 accounting oracles; other21
controls pass, final completed counter6/15. Earlier live parity agrees with
the shared corrupted state, so this is the new oracle's discrimination, unlike
the prior skipped-application probe. Restore exact source before next work.

R6's always-open door also deserves the actual follow-up, not inference from
a plan rejection. Predeclare three added sequential controls in existing
next-inactive profiles: after full refused/retry proof, actual WindDown without
plan must complete the same run into run3 under the absent-tree pin, retain
Founder v21/earned6/no purchases/no Reputation frozen row, and verify both
two-entry histories. Exact fallback retry must preserve all twelve table rows.
No reseeded later heads. Add explicit completed fallback count3 to whole matrix.
Demonstrate compiling omission of empty-plan early return: these three fallback
requests must wrongly reject while other controls stay defended. No new plan
policy/epochs/owner decision or meaning-change to the initial24-case population.

Fallback-positive146d0d/5fd999 session5762 terminal0: all24 original profiles
pass,9 initial applications/15 initial refusals and3 sequential plan-free
fallbacks,2.826s. Each fallback completes the previously refused run without
reseeding, keeps the absent-tree version/pin/no purchases, verifies both
two-entry histories and preserves12tables on exact retry. Commit positive
test before the predeclared empty-plan omission; no production fault active.

Empty-plan omission againstb36f2777 a336b8/ecd5cb session2860 terminal2:
all24 profiles execute; exactlythree next-inactive follow-up requests wrongly
refuse reputation_plan.tree_inactive at the new no-plan-door oracle. Other21
controls pass, completed counters9/12/0 instead9/15/3. Source compiles;
no assertion/corpus/threshold changed during the live handle. Restore exact
reputation_intent.go before broader baselines. Prior fault evidence remains
bounded to43999875, this additive fallback span still needs Claude.

Exact source restorec0fe39/c1482f: prestige.go SHA256
6d07581c74822e57d63c09b0ed139d98b0b1f44aba566f8d654ad2f4213cb691;
reputation_intent.go SHA256
11fe1b6a202f45b142acbf6de1375522dd21011fa108a04db79cb72a4fba8189.
Both match pristine source, empty production diff. Broad handles all terminal:
d1fb2b/f6c66a/60d37b/391b07 session92863 exit0, full production38.091s,
save0.327s/reputation0.188s/gameui0.295s with-count=1 and scoped vet. Host SQL
skips are not persisted proof. e5f3ce/e41aa7 session73640 exit0, strict TS and
Svelte0errors0warnings;93 client files pass/22skip,8186 tests pass/339skip4.94s.
716f64/21539c session90418 exit0, declared production Integration12.814s,
36 top-level/143 subcases/no skips independently counted from PASS lines.
New24 matrix passes3.19s including allthree follow-up fallbacks. Docker orphan
warning not cleanup authority; no files edited while any broad handle live.

Reconcile RP-299 ledger/canonical evidence docs/current board/executable queue/
per-RFC plan/append-only roadmap log. New test428lines, no existing test removed
or weakened, production/artifact/corpus/CI/epochs unchanged. Initial compile/
raw-json formatting oracle mistakes preserved above; full replay-state canonical
comparison retains every exact-number field and copied tier negative, while
actual persisted head/row comparisons remain raw byte-exact. Initial earned6/
stored zero-promise offers are diagnostic, not natural progression/live producer/
default browser/AT/minted-release/fullAC9/CI/1.0 proof. No checkbox/status/archive/
push/cleanup/owner-body edit. Full new range aftere5a69731 including final
records needs designated Claude review, earlier independent spans still pending.
Next accepted safe work: census separate R8 cross-pin/first-failure Go/TS evidence,
predeclare missing cases; this SQL test is not that portable corpus. Full
nine-tier/platform goal remains active/progress; all previous holds intact.
