# Cosmetic Shop v1 log

## 2026-09-25 — Predeclaration (Claude)

**Implemented by:** Claude, in the kernel-path lane. It is the only lane touching kernel-guarded
paths during this work. Every guarded commit bumps `kernel/VERSION` + `server/kernel/version.go` +
`client/src/kernel/version.ts` in the same commit to HEAD + 1.

Base: `3a5234d4`, kernel 0.3.121, latest migration 00079.

Batches C1–C8 are listed in `plan.md`. The known numbering deviation is recorded there
(Founder v24). `server/minigame/session.go` (left unformatted by the Typer lane) is gofmt'd inside
the first kernel-bumping commit, never in a standalone commit.

## 2026-09-25 — C1: `cosmetics` catalog grammar (Claude)

- **Go:** `server/cosmetic/catalog.go`, holding `Load` and `ValidateTransition`.
- **TypeScript:** `client/src/cosmetic/catalog.ts`, holding `loadCosmeticCatalog` /
  `parseCosmeticCatalog` / `validateCosmeticTransition`.
- **Shared corpus:** `testdata/cosmetic/catalog-fixtures-v1.json` has 25 cases. The accept cases
  are v1 and two sorted rows. The reject cases are:
  - each of `price`/`currency`/`sku`/`product_id`/`cost`/`store`/`amount` on an item;
  - `price` at the top level and inside `unlock`;
  - an unknown slot or unlock kind;
  - tier 9, -1 and 1.5;
  - a duplicate id, unsorted rows, or a non-mechanical id;
  - schema 2, empty or null items, a missing unlock;
  - a duplicate JSON key, trailing data.
- **Pinned fixture artifact:** `balance/testdata/cosmetics/fixture-v1.json`, which holds only
  Horse Armor at tier 1.
- **TS duplicate keys:** the TS loader carries its own recursive duplicate-key check, because
  `JSON.parse` collapses duplicates and the Go loader rejects them.
- **Registration and kernel:** `server/cosmetic/` and `client/src/cosmetic/` are registered in
  `kernel/affecting-paths.json`. Kernel 0.3.121 → 0.3.122. `server/minigame/session.go` is
  gofmt'd in this same bumping commit (one blank line; behavior-identical).

Evidence (cold):
- `make test-go GO_PACKAGES='./cosmetic' GO_TEST_FLAGS='-count=1'` passes, 3 tests.
- `vitest run test/cosmetic-catalog.test.ts` passes 3/3, deciding every shared case identically.

Severing (all restored afterwards):
- **G1:** Go `exactObject` without the key-count check fails
  `reject_item_price: accepted`. This is AC1's named failing case.
- **T1:** TS `exactObject` checking only required keys fails `reject_item_price`.
- **T2:** TS skipping the duplicate-key check fails `reject_duplicate_key`.

## 2026-09-25 — C2: `cosmetics` joins the replay bundle (Claude)

- **Go:**
  - `production.CatalogBundle.Cosmetics` joins `valid()`, with the artifact count and the
    requirement `withCosmetics ⇒ withPetSpecies ∧ bytes`.
  - `replaycatalog.Load`, `validArtifactNames` (the name set and the `cosmetics ⇒ pet_species`
    chain rule) and `want++`.
  - `settleAndActivateFoundations` now enforces §2's permanent-ID rule first, before any state
    validation. A next bundle that drops a pinned id or drops the artifact cannot settle an Exit.
- **TS:** `ReplayArtifacts.cosmetics`, the allowed name, the chain rule, and
  `loadCosmeticCatalog`.

Evidence (cold):
- `make test-go GO_PACKAGES='./cosmetic ./replaycatalog ./production' GO_TEST_FLAGS='-count=1'`
  passes.
- Client `tsc` is clean, and the full `vitest run` passes 6767.
- New tests: `TestLoadCosmeticsRequiresItsChain`, `TestSettleRejectsCosmeticIDDrop`, and
  `client/test/cosmetic-bundle.test.ts`.

Severing (all restored):
- Dropping the Go chain rule fails with "cosmetics loaded without pet_species".
- Dropping the TS chain rule fails the TS bundle test.
- Disabling the settle check fails with "dropped artifact: settle accepted".

Kernel 0.3.122 → 0.3.123.

## 2026-09-25 — C3: Founder v24 `cosmetics` (Claude)

Numbering is landing-order, so the RFC's "v22" (OD-16: next free) is **Founder v24**, and the
replay-inputs carry is **v11** (v9 went to Reputation, v10 to Pet Adoption).

What changed:
- **Go:**
  - `server/cosmetic/state.go` holds `State`, `NewState`, `Clone`, `Equal`, `ValidateShape` and
    `ValidateAgainst`.
  - `save/state.go`: `LatestFounderVersion=24`, the `stateV24` wire with pointer-decoded
    `owned`/`equipped` (null or missing rejects), encode/decode, and "cosmetics before v24"
    rejection. From v24 on, the shape is validated against the `pets` key set.
  - `production`:
    - the version floor is 24 when `cosmetics` is pinned;
    - `validateFounderCosmetics`: present iff pinned, and resolving under the catalog;
    - activation in `settleAndActivateFoundations` and `activateFounderFeatureState`, which
      rejects any pre-existing value;
    - the `applyFounderReplayOutput` Exit carry;
    - replay-extension `cosmetics`, which the carry validity requires at floor ≥ 24 with
      wire ≥ 11 and forbids below that.
  - `save.ReplayInputsVersion` is 11, and the parser still accepts 10.
- **TS:**
  - `client/src/cosmetic/state.ts`, plus the `replay.ts` twins: the floor, the save keys and
    their stripping, restore/encode, activation (the carry and replay arms), the `FounderExtensions`
    parse, and the v11 envelope and carry guard.
- **Regenerated Go-authored fixtures:** `testdata/replay/apply-logged-v1.json` via
  `make replay-fixture`, and `testdata/replay/reputation-tree-v1.json` via
  `REPUTATION_UPDATE_FIXTURE=1`. The diff is only the replay-inputs `"v": 10` → `"v": 11`
  (137 lines, nothing else).

Evidence (cold):
- `make test-go GO_PACKAGES='./cosmetic ./save ./production ./replaycatalog ./pet ./gameui ./account ./gameserver ./harness'`
  with `-count=1 -timeout 45m` passes.
- Client `tsc` is clean, and `vitest run` passes 6769.
- New tests: `TestFounderV24CosmeticsRoundTripAndInvariants` (15 decode rejections, 4 encode
  rejections, canonical `[]`/`{}` bytes), `TestCosmeticsOwnsFounderV24Activation`,
  `TestPinnedCosmeticsValidatesAndCarries` (catalog-aware rejections, the AC8 Exit output carry,
  the v11 carry rule), and `client/test/cosmetic-founder-state.test.ts` (TS codec rejections plus
  artifact binding in both directions).

Severing (all restored):
- **S1:** dropping "cosmetics before v24" gives "encode accepted cosmetics before v24".
- **S2:** dropping the settle activation makes the activated Founder invalid.
- **S3:** the TS path without "cosmetics artifact requires Founder v24" fails the binding test.

**Carried to C4:** the TS Exit-activation witness. As with Reputation, it is proven through a
Go-authored Exit corpus case, which C4's cosmetic corpus will include.

Kernel 0.3.123 → 0.3.124.
