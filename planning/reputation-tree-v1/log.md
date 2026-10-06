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
