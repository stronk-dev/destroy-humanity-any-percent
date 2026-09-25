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
