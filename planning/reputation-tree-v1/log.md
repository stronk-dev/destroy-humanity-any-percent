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
