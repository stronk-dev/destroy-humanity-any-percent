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
