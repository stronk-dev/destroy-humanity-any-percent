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
