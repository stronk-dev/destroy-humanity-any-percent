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
