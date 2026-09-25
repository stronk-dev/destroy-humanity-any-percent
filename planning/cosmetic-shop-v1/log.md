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

## 2026-09-25 — C4: the three cosmetic intents, events and migration (Claude)

- **Go:** `server/production/cosmetic_intent.go`
  - `handleFounderCosmetic`. It is the unguarded Founder boundary, because §4.5 says cosmetics are
    never `exclusive_activity`.
  - `resolveCosmeticActiveCompany`, which freezes the active Company's
    `{stream, revision, run_seq, tier}` for acquire only. The client never sends a tier.
  - `applyFounderCosmeticResolved`, which follows the §4.1 prefix (`not_eligible/inactive`,
    `unknown_id/cosmetic_id`) and then the per-intent order:
    - acquire: `owned`, then `locked`;
    - equip: `unknown_id/pet_id`, `not_owned`, `already_equipped`;
    - unequip: `unknown_id/pet_id`, `nothing_equipped`.
  - `checkCosmeticsTransition` runs in `ApplyFounderLogged`'s commit guard. Only the three intents
    may change `cosmetics`; the one exception is activation from absent to empty.
  - `ParseIntent` uses strict exact keys, so `price` or `tier` in a request is terminal
    `invalid/<kind>.fields` (N2).
- **Events:** `save/intent.go` adds `cosmetic_acquired.v1` / `cosmetic_equipped.v1` /
  `cosmetic_unequipped.v1` with strict payloads. `replaced_cosmetic_id` must be present, possibly
  null. Migration `00080_cosmetic_events.sql` extends the closed kind constraint, and the
  migration pin moves to 80.
- **Receipts:** `{intent_id, outcome, founder_revision, kind, cosmetics, event}`. There is no
  amount, price, currency or payment field; the corpus asserts this.
- **TS:** `applyFounderCosmetic`, the parse cases, the event kinds and `checkCosmeticsTransition`
  in `replay.ts`.
- **Corpus:** Go-authored `testdata/replay/cosmetic-v1.json` holds 21 Founder cases across
  `species`/`shop`/`pair` bundles (the two-item fixture catalog reaches replace and `not_owned`),
  plus the Exit case `exit-activates-founder-v24` (current `pet_species` → next `cosmetics`).
  `reputationFounderState` (test-only) gained the v23/v24 defaults so that case can be built.

**Finding, fixed here (pre-existing, Pet Adoption lane):** the TS Founder Exit arm capped
`result_founder_wire_version` at 22 (`safeInteger(…, 1, 22)` and the allowed-version list). Any
Exit whose Founder result is v23 (Pet Adoption) or v24 was therefore unreplayable in TS; no TS
Exit witness at v23 existed to reveal it. Both caps now allow 23 and 24. The new Exit case fails
with "integer outside exact domain" when the cap is put back at 22.

Evidence (cold):
- `make test-go GO_PACKAGES='./cosmetic ./save ./production ./replaycatalog ./pet ./gameui ./account ./gameserver ./releasepackage' GO_TEST_FLAGS='-count=1'`
  passes.
- Client `tsc` is clean, and `vitest run` passes 6792, including all 23 cases of
  `cosmetic-replay.test.ts`.
- Postgres `docker compose -f compose.save-test.yml run --rm test go test -p 1 ./save ./production ./gameserver ./account -run Integration -count=1`
  passes, including the new `TestCosmeticIntegrationPersistsReplayableFounderLog`. It covers:
  - an applied acquire, a byte-identical retry, `idempotency_conflict`, then `owned` and the
    `invalid` price field;
  - exactly one event row, with the rejections storing none;
  - the Company stream bytes and revision left untouched;
  - `UPDATE … kind='cosmetic_purchased.v1'` refused by the constraint;
  - Founder history `ReplayVerified`.
- `TestCosmeticEventPayloadsAreStrict` checks the extra-field, missing-replace and order-0
  payloads. AC9's "database rejects an extra payload field" is enforced by the save layer's strict
  decoder before any insert; the database constrains kinds only (DESIGN-GAP note: no JSON-schema
  check exists in SQL for any event kind).

Severing (all restored):
- **G1:** dropping Go's `locked` check fails the corpus build at `rejects-locked-at-tier-0`. This
  is AC5's named failing case.
- **G2:** disabling the commit guard gives "adopt_pet arm that cleared cosmetics: err=<nil>".
  This is the AC8 failing case.
- **T1:** dropping the TS `locked` check fails 2 corpus cases.
- **T2:** the TS Exit cap back at 22 fails the v24 Exit activation case.

Kernel 0.3.124 → 0.3.125.

## 2026-09-25 — C5: the Game UI `features.cosmetics` arm (Claude)

This is the §7.1 projection as an **optional** v4 arm, like Reputation R9 and Pet Adoption PA7. A
required property would violate API Foundation C2, and the RFC's "v4 plus a required `cosmetics`"
text predates snapshot v4's landing. It is recorded here as a numbering/shape reconciliation, not a
mechanic change.

- **Server:**
  - `server/gameui/features.go` `projectCosmetics`. Items follow catalog order, with `owned`, an
    advisory `acquirable` (not owned and the active Company tier ≥ unlock), `lock` (non-null only
    while locked and unowned) and sorted `worn_by`. `wearers` follow the byte-sorted pets keys.
  - The arm is emitted exactly when `cosmetics` is pinned; `active` means Founder ≥ v24.
  - Fact `feature.cosmetics`, which the client `GAME_UI_FACT_IDS` also registers.
- **Schema:** `server/account/game_ui_api_schema.go` adds `GameUICosmeticsArm`/`Item`/`Lock`/
  `Wearer`. `make api-generate` regenerated `docs/generated/api.json` and the client types.
  `docs/generated/api-compat-v1.json` is **unchanged**: the additive optional field passes the
  compatibility gate with no re-pin.
- **Client:** `parseCosmeticsArm` in `client/src/game-ui/contracts.ts` uses exact keys and rejects:
  - owned together with acquirable;
  - a lock on an owned or acquirable item;
  - an unowned item that is worn;
  - a `worn` value absent from its `worn_by`, and the reverse;
  - an inactive arm that is non-empty.
- **Shared fixture:** the Go-authored `testdata/gameui/cosmetics-arm-v1.json` has 5 cases
  (locked at T0, acquirable at T1, owned with no wearer, worn, not worn).

Evidence (cold):
- `make test-go GO_PACKAGES='./gameui ./account ./gameserver' GO_TEST_FLAGS='-count=1'` passes.
- Client `tsc` is clean, and `vitest run` passes 6794.

Severing (all restored):
- Dropping the TS owned+acquirable and lock checks fails `cosmetics-arm.test.ts`. This is AC10's
  named failing case.
- A Go `acquirable` that ignores the lock fails with "tier-0 row".

No kernel-guarded path was touched, so there is no version bump.

## 2026-09-25 — C6: Desk shelf, parody receipt, overlay, curtain contract (Claude)

- **`client/src/game-ui/cosmetics/presentation.{json,ts}`:** the §7.2/§8 presentation and
  curtain-binding contract. It is a strict separate presentation file, like the Garage lane's
  `features-presentation.json`, not a presentation-catalog v4 bump (a recorded deviation in form
  only). The loader rejects:
  - a missing `paid_cosmetic_dlc`;
  - an anchor without `reference_price_anchor`, or the reverse;
  - `checkout_flow` on an item;
  - a shop that does not bind exactly `checkout_flow`;
  - a duplicate or unknown pattern;
  - `re_release` (reserved, unbound in v1);
  - an undeclared copy key;
  - any extra key.

  It exports `curtainList` for the future honesty appendix, and it loads eagerly so a violation
  fails every build and test.
- **`CosmeticShelf.svelte`:**
  - per-item curtain small print in every state, which Buy and the anchor reference through
    `aria-describedby`;
  - the shop-level checkout curtain under the heading;
  - the anchor as `<s>` inside readable text;
  - Buy, `locked` with its tier, owned, per-wearer equip/unequip, and `no_wearer`;
  - an inline `role=status` receipt (`shop.receipt.line` + `payment_method`), no modal;
  - focus moves to the owned heading after the authoritative snapshot;
  - no cart, confirm, quantity, timer, badge or upsell.
- **`CosmeticOverlay.svelte`:** a CSS layer plus the `annoyed` pose (ears back, tail flick) with
  no text node, `aria-hidden`, and static under reduced motion. Mounting it into the live pet panel
  remains manifest row G10 (§7.4); it is not claimed here.
- **`GameUIApp.svelte`:** it shows the shelf iff `features.cosmetics.active` and some item is
  acquirable or owned. That makes it absent at T0 until something is owned (OD-4), and it persists
  once owned. Below v24 or unpinned, the static card is unchanged (§6, fail closed). Intents are
  Founder-scoped through `runtime.intent` with `COSMETIC_REJECTIONS`, and the receipt order number
  comes from the applied receipt's event payload.
- **Copy:** `copy/catalog/cosmetics-candidate.json` holds 19 keys: the §9 keys plus 7 inline
  rejection keys. It is **candidate text for owner adoption**, with no currency literal (price
  only via `{price}` ← `constant.price_zero`). The three ratified Horse Armor keys are reused
  unchanged.

Evidence (cold):
- `make typecheck build-client test-client verify-client-boundary copy-check test-browser` all
  pass: 6796 unit, boundary scan over 16 components, 20556 browser.
- `make test-game-ui-composed` passes (v4 features lane).
- New tests:
  - `cosmetic-presentation.test.ts`: the curtain contract, and presentation IDs equal to the
    pinned catalog IDs.
  - `cosmetic-shelf-browser.test.ts`, in 3 engines:
    - keyboard buy, with no owned state while pending and focus on the owned heading;
    - the receipt order number;
    - curtains visible (not hover-gated, non-zero height) in unowned, pending, owned and equipped;
    - axe;
    - the lock and no-wearer paths, and 320 px reflow;
    - the overlay: no text node, pose, and static under reduced motion.
  - `cosmetic-host-browser.test.ts`: T0/T1/owned visibility, the pre-activation static card, and an
    exact-key `acquire_cosmetic` request at the Founder revision.

Severing (chromium, all restored):
- **S1:** curtains removed once owned fail the render test. This is AC12's named failing case
  (the render test fails, not just the loader).
- **S2:** "owned" rendered while pending fails. This is AC11's optimistic-ownership failing case.
- **S3:** the shelf shown at T0 when nothing is owned fails the host test.

**Carried (AC14, like Pet Adoption's composed witness):** the composed real-server Buy → reload
witness needs a bundle that pins `cosmetics`. No epoch pins it (fixture-first, no mint), so the
composed lane cannot render the shelf yet. It is carried to the mint epoch, together with the N5
network trap run on that flow (C7 adds the trap to the component-level flow).

No kernel-guarded path was touched, so there is no version bump.

## 2026-09-25 — C7: no real money, by construction (Claude)

- **N4:** `client/tools/verify-no-payment.mjs` (`make verify-no-payment`, now in `verify-client`).
  - **The gate:** the client's runtime `dependencies` must equal
    `client/tools/runtime-dependency-allowlist.json` exactly.
  - **The secondary net:** a boundary-aware denylist scan of `client/pnpm-lock.yaml`,
    `server/go.mod` and `server/go.sum` for payment, IAP and ads SDK names.
  - **Built-in fixtures:** 6 rejected (an extra runtime dependency, `/stripe@`,
    `@stripe/stripe-js`, `github.com/stripe/stripe-go`, `in-app-purchase`, `@paypal/paypal-js`),
    and 2 near-misses accepted (`squared-distance`, `stripes-css`).
- **N5:** `client/test/network-trap.ts` traps fetch, WebSocket, `PaymentRequest`,
  `navigator.credentials.*` and `window.open` to same-origin allowlisted paths.
  - It wraps the whole host-level shop flow in `cosmetic-host-browser.test.ts` (acquire through
    `runtime.intent`, snapshot, owned).
  - Its own failing case is recorded: off-origin checkout, `PaymentRequest` and `window.open` are
    all logged.
  - Because the host test uses a fixture runtime, the trap witnesses that **the shop UI code**
    issues no network or payment calls. The composed real-server run is carried with AC14.
- **N6:** `deployment/Caddyfile` adds `Permissions-Policy "payment=()"` and
  `Content-Security-Policy "connect-src 'self' wss://{host}"`. `ValidateCaddyfile` requires each
  exactly once.
  - `TestCaddyRequiresPaymentBlockingHeaders` rejects: a missing Permissions-Policy, `payment=(self)`,
    a missing CSP, a widened `connect-src`, and a second CSP header.
  - The real-Caddy `make test-deployment-release` passes. Its fixture uses
    `deployment/Caddyfile.release-integration`, **not** the shipped file, so I dropped an
    attempted header assertion there: it would have witnessed the wrong file. The shipped file is
    covered by the validator test above.
- **N7:** `copy-pipeline.mjs` `validateShopCurrency`. No `shop.*`/`cosmetic.*` text may contain a
  currency symbol or ISO code with a number, or "N [word] points", unless it carries provenance;
  placeholders are exempt. It ships with 4 failing fixtures and 3 accepted ones in
  `verify-copy.mjs`.
- **N8 / §5 / N3:** `client/tools/verify-cosmetic-boundaries.mjs` (`make verify-cosmetic-boundary`,
  in `verify-client`). `go list -deps ./cosmetic` must be standard library only, and
  `client/src/cosmetic/*.ts` may import siblings only. It ships with 9 rejected fixtures (economy,
  production, save, prometheus, operations; `../replay`, `../economy-kernel`,
  `../game-ui/runtime`, a telemetry package).

Evidence (cold):
- `make typecheck test-client verify-client-boundary verify-ci-topology verify-combat-boundary verify-meters-boundary verify-achievements-boundary verify-cosmetic-boundary verify-no-payment copy-check`
  passes.
- `make test-go GO_PACKAGES='./releasepackage' GO_TEST_FLAGS='-count=1'` passes.
- `make test-deployment-release`, `test-deployment-rehearsal` and `test-deployment-operations` all
  pass.

AC13 failing runs (all restored):
- **N5:** an off-origin `fetch("https://example.invalid/checkout")` injected into the shelf's Buy
  fails the host test ("fetch https://example.invalid/checkout").
- **N7:** `shop.cosmetics.buy` = "Buy for $2.50" fails `make copy-check` ("shop copy may not
  contain a currency amount").
- **N4:** adding a `left-pad` runtime dependency fails `make verify-no-payment` ("differ from the
  allowlist").
- **N8:** `client/src/cosmetic/catalog.ts` importing `../replay` fails
  `make verify-cosmetic-boundary`.
- **N6:** the five mutations in `TestCaddyRequiresPaymentBlockingHeaders`.

**Pre-existing, not caused here:** `make deployment-config-check` fails identically at `1bba27ba`
without these changes. It validates environment-provided runtime configuration, which this
workspace does not set.

No kernel-guarded path was touched, so there is no version bump.
