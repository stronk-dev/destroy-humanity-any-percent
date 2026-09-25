# Pet Adoption v1 — implementation log

## 2026-09-25 — Predeclaration (Claude)

**Implemented by:** Claude. **Review:** awaiting Codex designated cross-party review.

Scope is `plan.md` P1–P7, fixture-first, with no production epoch mint.

- **Numbering:** Founder v23, the next free version, since Reputation took v22. Migrations start at
  00079. The replay-inputs version is the next free one.
- **Kernel protocol:** every commit that touches a guarded path bumps `kernel/VERSION` in the same
  commit.

**Recorded reading (scalar Founder chain, C17):** Founder v23 = v22 + `pet_identities`, so a
bundle that pins `pet_species` must also pin `reputation_tree`, which transitively pins
`minigame_api` → … → `pets`. The RFC names only `pets` as a prerequisite. The stronger chain is
forced by the existing scalar version design, not chosen here. PA2's "pet_species requires pets"
stays true.

## 2026-09-25 — P1: the companion tone and lint (Claude)

- Added `companion` to the closed tone union in `copy-pipeline.mjs` and `client/src/copy/index.ts`.
- Added `validateCompanionEntry`, run inside `validateCopySafety`. Every `pet.*` key must use the
  companion tone. Companion text rejects:
  - price tokens (`$`, `0.00`, free, cost, price, buy, purchase);
  - urgency or streak phrases;
  - curtain or disclosure phrasing;
  - statistics.
- **Recorded reading:** a bare "now" is not banned, because PA8.1 itself names the "Not now"
  control. Urgency is matched as phrases ("act now", "right now", "limited", "expires" and so on).
- Generated Go `copykeys.CompanionKeys()`, so the Go loader can check tone (AC1).
- **Candidate copy:** `copy/catalog/pet-candidate.json` holds every PA8.6 key with `PENDING OWNER
  COPY` text. Nothing is minted with it (PA8.6).
- **Recorded reconciliation:** the name-pool keys are `pet.name.server_room_cat.nNN`, because the
  copy-key grammar requires every segment to start with a letter. That still matches the RFC's
  `^pet\.name\.[a-z0-9_]+\.[a-z0-9_]+$`.
- **Evidence:** `make copy-check` passes (457 keys). The seeded companion rows ("$0.00", "limited
  time", "Free", disclosure, "50%") and a `pet.*` key in the diegetic tone all fail. "Not now"
  passes. Severing the `validateCompanionEntry` call made `verify-copy` fail on the "$0.00" fixture.

## 2026-09-25 — P2: the pet_species loader and adoption draws (Claude)

- **Go:** `server/pet/species.go` has `LoadSpeciesCatalog` (exact keys; PA2 grammar; exactly one
  starter; canonical combat temperament order; name and species copy keys registered and
  companion tone) and `DrawAdoption` (PA4.3–PA4.5). **TS:** `client/src/pet/species.ts` is the
  byte twin.
- **Fixtures:**
  - `balance/testdata/pet-species/fixture-v1.json`: one cat, 12 names, 10 palettes.
  - `testdata/pet/species-fixtures-v1.json`: 18 loader cases, each shared by Go and TS. They cover
    every AC1 failing class except "pet_species without pets", which is a bundle rule (P3).
- **Vectors:** `testdata/pet/adoption-draw-vectors-v1.json` has 79 vectors, computed by an
  independent Python implementation of the PA4 recipe. They cover all 6 temperaments and 10
  palettes, and include pairs that differ only in `intent_id` but expect identical draws (AC3).
  Go and TS both match them.
- **Temperament parity:** `TestTemperamentOrderIsTheCombatEnum` checks the order against
  `server/combat`. Its negative probe shows that combat rejects an unknown member.
- **Severing:** each of these turned a test red.
  - Go: S1 digest bytes [1:9]; S2 version nibble 0x60; S3 starter count `< 1`; S4 the companion
    check removed; S5 the temperament order ignored.
  - TS: the ID digest shifted a byte; the companion check removed.
- **Evidence (cold):** `make test-go GO_PACKAGES=./pet GO_TEST_FLAGS=-count=1` passes, `vitest
  test/pet-species.test.ts` passes 4/4, and `tsc` is clean.
- **Kernel:** `server/pet/` and `client/src/pet/` are guarded, so this commit bumps
  `kernel/VERSION`.

## 2026-09-25 — P3: pet_species bundle wiring (Claude)

- **Go:** `CatalogBundle.PetSpecies`. `replaycatalog.Load` loads `pet_species` against the
  generated copy and companion registries. `validArtifactNames` allows it only with
  `reputation_tree` and `pets`, and counts it in the artifact set. `valid()` rechecks the chain.
- **TS:** `ReplayArtifacts.pet_species` and `bundle.petSpecies`, with the same chain rule.
- **Tests:**
  - `server/replaycatalog/pet_species_test.go`: complete chain loads and resolves; `pet_species`
    changes the constants hash; loading is rejected without `reputation_tree` and without `pets`,
    and for an invalid artifact.
  - `client/test/pet-species-bundle.test.ts`: mirrors the Go test.
  - `client/test/pet-fixture-bundle.ts` is a shared fixture chain for later TS tests.
- **Severing:**
  - Go: removing the loader chain rule failed the test ("pet_species loaded without
    reputation_tree").
  - TS: removing the chain rule failed 1 of 2.
  - Go `valid()` recheck (**survived, recorded, not vacuous**): severing it did not fail. The
    artifact count plus the Reputation declaration pairing already reject every forged bundle I
    could build. It stays as defense in depth; the same precedent is in the Reputation log.
- **Evidence (cold):** `./replaycatalog` and `./production` pass; the client passes 6739 tests; tsc
  and vet are clean.
- **Out of scope:** `server/replaycatalog/catalog_test.go` has a pre-existing gofmt drift from
  Typer's committed edits. This change doesn't touch that file.

## 2026-09-25 — P4: the Founder v23 codec, activation, and the replay-inputs v10 carry (Claude)

**Go:**
- `save.LatestFounderVersion = 23`, with a `stateV23.pet_identities` pointer-field wire, so a
  missing key never becomes a zero value.
- `pet.ValidateIdentityShape` (key sets equal to `pets`, UUIDv7, grammar, adoption coordinate no
  later than the watermark) and `pet.ValidateIdentitiesAgainst` (catalog resolution, cap). The
  second runs from `CatalogBundle.ValidateFoundationState`.
- Activation, in both `settleAndActivateFoundations` and the replay `activateFounderFeatureState`,
  is legal only with empty `pets`.
- `versionFloors` sets founder 23 when `PetSpecies` is pinned.
- `replayFounderExtensions.PetIdentities` is required at floor ≥ 23 with wire ≥ 10, and forbidden
  otherwise. `save.ReplayInputsVersion` goes 9 → 10.
- The Exit live/replay parity copy (`prestige.go`) carries identities at v23.
- `pet.InitialCareState` implements PA4.6.2.

**TS:** the same rules in `replay.ts` and `client/src/pet/identity.ts`. The Founder v23
restore/encode, floor, carry and activation are all twins of the Go side.

**Recorded numbering:** the RFC's "replay-inputs v7" maps to the next free version, v10. The golden
fixtures `testdata/replay/apply-logged-v1.json` and `reputation-tree-v1.json` were regenerated
(`make replay-fixture`, `REPUTATION_UPDATE_FIXTURE=1`). The diff is exclusively `"v": 9` →
`"v": 10`: 137 lines. `client/test/reputation-founder-state.test.ts` gained `petIdentities: {}`,
because the field is now required.

**Tests:**
- `server/save/pet_identity_state_test.go` (AC2 shape) and
  `server/production/pet_adoption_activation_test.go`: AC14 run-boundary, New-Founder-forward and
  replay activation, with a pet-without-identity refusal; AC2 catalog-aware half; AC8 carry.
- `client/test/pet-founder-state.test.ts` (AC2 twin and artifact binding).

**Severing:**
- TS T1 (key-set check removed) and T2 (catalog resolution removed): each failed 1 of 2.
- Go G2 (v10 gate removed): `TestFounderCarryHoldsPetIdentitiesAtFloor23` failed.
- Go G1 (**survived, recorded, not vacuous**): with the activation guard's `len(Pets) != 0` clause
  removed, the case is still refused. The post-activation identity-shape check rejects `{}`
  identities against one care record. The guard is a redundant earlier refusal.

**Evidence (cold):**
- Go: `./production ./save ./pet ./replaycatalog ./gameui ./account ./gameserver` all pass.
- Client: 6739 tests pass, and tsc is clean.
- `./harness` hit the 600 s `go test` timeout inside a combined `test-go` run, which is an
  instrument limit and not an assertion; it is not counted either way. It is rerun through its own
  targets later.

Docs: new `docs/pet-adoption.md`; `docs/pet-care.md` pointer; `docs/production-engine.md` stale
carry sentence corrected (v9 Reputation, v10 pets).

## 2026-09-25 — P5: the adopt_pet intent (Claude)

**Go** (`server/production/pet_adoption_intent.go`):
- the handler: attendance resolution, the Soul-guarded store path, and one CSPRNG nonce drawn
  inside the callback;
- `applyFounderAdoptionResolved`: the PA4.2 order, the draws, and the single-transaction effect;
- the receipt and `pet_adopted.v1` event;
- `checkPetIdentityTransition`, run by `ApplyFounderLogged`'s defer for every applied arm
  (PA3.5);
- `save.EventPetAdopted` with strict payload validation;
- migration `00079_pet_adopted.sql`, with the contiguity pin moved to 79;
- parser case and routing.

**TS:** the `adopt_pet` parser case, the `applyFounderAdoption` arm, and the same immutability
check in `finish`. `pet_adopted.v1` joins `REPLAY_EVENT_KINDS`.

**Recorded readings** (filling what the RFC leaves unspecified):
- **Rejection details:** the RFC names `unknown_id` for rows 3 and 4 without a detail. The details
  are `unknown_species` and `unknown_name`, following care's `unknown_pet`.
- **Soul exclusivity:** adoption uses care's Soul-recovery exclusivity guard (`exclusive_activity`).
  The RFC says attendance behaves "identically to care_action" and is silent on exclusivity, so the
  sibling's rule is mirrored rather than inventing an exemption.
- **Row 5 (`species_locked`)** is unreachable in v1, because the loader admits only starter rows.
  It is implemented but has no vector.

**Tests:**
- `pet_adoption_test.go`:
  - the corpus: rows 1–4 and 6, apply, the unknown-species-at-cap ordering, and care on the
    adopted pet (AC4/AC5);
  - a tampered nonce cannot reproduce the receipt (AC7);
  - direct unit coverage of `checkPetIdentityTransition` (AC6);
  - `TestTransitionLayerRejectsIdentityMutatingArms`, via an explicit test-only
    `founderTransitionTestArm`: an arm that mutates a temperament, or one that drops the map,
    fails with `ErrInvalidEngineState` and restores state (the AC6 failing cases).
- `pet_adoption_integration_test.go`, real Postgres, run verbosely and confirmed to have
  executed:
  - applies through `Service.Handle`;
  - the retry is byte-identical with `draws == 1` (AC10);
  - a changed body yields `idempotency_conflict` with no draw;
  - the event persists, and the constraint rejects an unregistered kind (AC9);
  - the second adoption is refused at cap;
  - `VerifyFounderHistory` returns `ReplayVerified`.
- `client/test/pet-adoption-replay.test.ts`: TS byte-matches all 10 corpus cases; a tampered
  nonce fails to reproduce the receipt; a stripped nonce throws (AC7).

**Severing:**

| Probe | Change | Result |
|---|---|---|
| S1 | cap check moved before the species lookup | corpus build failed ("unknown-species-at-cap: adoption_cap_reached") |
| S2 | watermark seeded at 0 | failed |
| S3 | immutability check disabled | `TestTransitionLayerRejectsIdentityMutatingArms` failed |
| S4 | draw at server time + 1 | the draw/receipt check failed |
| TS T1 | cap check before species | failed 1 of 12 |
| TS T2 | watermark 0 | failed 1 of 12 |
| AC10 | an extra nonce draw outside the callback | the integration test failed with `draws=3` |

**Invalid probe, recorded:** severing migration `00079` in place isn't meaningful, because the
shared Postgres has already applied it. The constraint witness itself (an unregistered kind is
rejected) is live.

**Evidence (cold):**
- Go: `./production ./save ./pet ./replaycatalog ./gameui ./account ./gameserver ./releasepackage`
  pass.
- The full `docker compose -f compose.save-test.yml run --rm test` suite passes.
- Client: 6753 tests pass, and typecheck is clean.

**Housekeeping:** `server/save/pet_identity_state_test.go`, committed in P4 unformatted, is
gofmt'd here.
