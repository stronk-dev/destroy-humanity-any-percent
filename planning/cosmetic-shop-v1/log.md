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

## 2026-09-25 — C8: the mechanical-isolation property, docs, and handoff (Claude)

- **`server/production/cosmetic_isolation_test.go` (AC7):**
  - It runs the existing `TestIntentPolicyPropertyTwentyFourHoursTwoHundredSeeds` shape, 200 seeds
    × 288 five-minute Company steps (manual batch / buy-max), in two arms.
  - The second arm interleaves seeded Founder cosmetic intents (acquire at a random tier 0/1,
    equip, unequip) on a v24 Founder with an adopted wearer.
  - **Assertions:** identical Company bytes; identical frozen Founder contributions (the
    Founder→production multiplier channel); every non-`cosmetics` Founder byte identical.
  - **Non-vacuity:** all three kinds must actually apply somewhere in the run.
  - **Fixed during authoring:** Founder commands pin the server time to the Fiscal period-open
    instant. Otherwise the automatic Fiscal sweep, which is time and not cosmetics, would
    confound the comparison (first observed as a `fiscal_period_seq` divergence).
  - **Failing case (in-suite):** `TestCosmeticIsolationCatchesAMultiplierLeak` installs a
    test-only arm that raises a Fiscal generator level (a multiplier input) on every cosmetic
    transition and requires the property to report a divergence. My first draft of this test
    passed for the wrong reason: the run errored on the unrelated sweep. It now asserts the error
    text contains "diverged", so a crash cannot satisfy it.
- **Docs (AC16):**
  - new `docs/cosmetics.md`;
  - `docs/founder-transitions.md`: the cosmetic arms and the commit guard;
  - `docs/game-ui.md`: the shelf;
  - `docs/production-engine.md`: replay-inputs v11 carry;
  - `docs/save-layer.md`: Founder v24;
  - `docs/copy-pipeline.md`: the N7 rule;
  - `docs/deployment.md`: the N6 headers.

Evidence (cold): `make test-go GO_PACKAGES='./cosmetic ./save ./production ./replaycatalog ./gameui ./account ./gameserver ./releasepackage' GO_TEST_FLAGS='-count=1'`
passes; gofmt and vet are clean.

**Handoff: ready for Codex designated cross-party review. Not self-approved; not archived; no
mint.**
- **Implementation range:** `8d88248a..` this commit (C1–C8). Kernel 0.3.121 → 0.3.125. Every
  guarded commit bumped in the same commit: `afe6529b`, `8e315569`, `581886a4`, `500d944c`.
- **Open:**
  - AC14 composed witness (needs a `cosmetics`-pinning epoch);
  - G10 pet-panel overlay mount;
  - owner adoption of `copy/catalog/cosmetics-candidate.json`;
  - the RFC author's reconciliation of the two recorded shape notes: the snapshot arm is
    optional, not a "required v4" field (C2); and the presentation is a separate strict file, not
    a presentation-catalog v4 bump.
- **Findings for other lanes:**
  - the TS Founder Exit arm's v22 cap, fixed in C4 (Pet Adoption lane impact);
  - `make deployment-config-check` fails at HEAD without these changes (environment config).
- **`verify-kernel-version`:** it still stops at the pre-existing `50a3a514` history item. None of
  this range's hashes is named.

## 2026-10-04 — Codex C6/AC11 keyboard-evidence review predeclaration

- **Range under review:** Claude's C6 landing `1bba27ba^..1bba27ba`, limited to
  AC11 keyboard acquisition and focus. This is not a verdict on the rest of
  Cosmetic Shop v1 or permission to archive it.
- **Observed gap before experiment:** the C6 log calls
  `cosmetic-shelf-browser.test.ts` a three-engine "keyboard buy" witness, but
  the test focuses the Buy button and invokes `.click()`; the host test does
  the same. Neither operation proves keyboard activation. The passing client
  suite is therefore not AC11 keyboard evidence.
- **Population and method:** mount the real Svelte shelf with its T1
  acquirable arm in the existing Vitest browser lane, focus Buy, send real
  `userEvent.keyboard("{Enter}")` and `userEvent.keyboard(" ")` in separate
  cases, and require exactly one `onAcquire("horse_armor")` call per key.
  Retain the existing pending/no-optimistic-ownership and authoritative
  snapshot/focus controls. Run the new cases in Chromium, Firefox and WebKit
  where this host can start each engine.
- **Discriminator:** temporarily sever Buy's `onclick` binding; both keyboard
  cases must fail at zero acquire calls, while the unmodified component
  passes. A keyboard event merely dispatched on the DOM or a `.click()` call
  does not count. If an engine cannot start, record it as unavailable, not
  green. If actual keyboard activation fails, report an implementation
  defect instead of weakening the test.
- **Authority and limit:** accepted Cosmetic Shop §7.3/AC11 authorizes a
  test-only supplement. No player-facing copy, intent semantics, artifact
  pin, RFC status, or owner decision changes in this batch. Codex's
  correction needs Claude's cross-party review before closure.

## 2026-10-04 — Codex targeted C6/AC11 review and keyboard supplement (RP-167)

- **Review by:** Codex. **Recorded by:** Codex. **Reviewed Claude range:**
  `1bba27ba^..1bba27ba`, limited to the AC11 keyboard Buy claim and its
  browser witness. **Decision: CHANGES REQUIRED for evidence.** The shipped
  Buy is a native button, but both C6 browser tests invoked `.click()`; the
  existing green tests did not exercise Enter or Space. This finding does not
  assert that keyboard Buy is broken and is not a verdict on C6's other ACs.
- The test-only supplement changes the shelf browser test to focus Buy and
  send `userEvent.keyboard("{Enter}")`, and adds a separate focused Space
  case. Both require exactly one acquire call; pending still cannot render
  owned before an authoritative snapshot. No product source, copy or schema
  byte changed.
- **Executed discrimination:** with only Buy's production `onclick` binding
  temporarily changed to a no-op, the Chromium run failed both keyboard
  cases at `expected [] to deeply equal ["acquire:horse_armor"]` and
  `["horse_armor"]`; the other two shelf cases passed. The handler was
  restored byte-for-byte (`git diff` on the component empty). Before and
  after severing, the unmodified component passed all four scoped tests in
  Chromium and WebKit. `make typecheck` passed with zero warnings; full
  `make test-client` passed 6912 with 84 browser-only skips. Cold Go
  `./cosmetic ./replaycatalog ./production` packages also passed, as did the
  focused catalog, bundle and transition negatives; these Go checks are
  context for C1/C2, not a full Cosmetics verdict.
- **Firefox limit:** the three-engine run passed 8 tests in Chromium/WebKit
  but Vitest reported an unhandled Firefox session connection timeout after
  60 seconds; a Firefox-only retry reached the same timeout before any test
  started. Both invocations were interrupted after their error summaries
  because Vitest stayed live. They are not green AC11 evidence. This host
  cannot close the three-engine criterion; an executable Firefox lane remains
  required. The initial attempt to combine several Go `-run` regexes through
  a Make variable was parsed as shell pipelines and was discarded; separate
  exact focused Go invocations passed.
- AC14's real-server Buy→reload path, the G10 live pet-panel mount, owner
  copy adoption, full implementation-range review and archival remain open.
  This Codex test correction itself requires Claude's designated cross-party
  review of its exact committed range.

## 2026-10-04 — RP-167 commit coordinate and post-restoration gate

- Predeclaration `b739d295`; test, ledger, plan and queue landed in
  `a4c21816`. Claude's designated review range is
  `b739d295^..a4c21816` plus this coordinate record commit. This is a
  Codex-authored test correction, not Codex approval of itself.
- After restoring the severed production handler, a sequential current-HEAD
  rerun passed all four scoped shelf tests in Chromium and all four in
  WebKit, with the performance sublane green on each invocation. An attempted
  *parallel* Chromium/WebKit rerun was invalid: concurrent Vitest processes
  raced while renaming their shared Vite dependency cache and WebKit failed
  at startup with `ENOTEMPTY`. The sequential rerun resolved that harness
  collision. Firefox remained unstarted after its independent 60-second
  session timeout; no three-engine pass is claimed.

## 2026-10-04 — Codex C7/N5 network-witness predeclaration

- **Question:** does `installNetworkTrap` actually observe every browser HTTP
  request during the shop flow as Cosmetic Shop §10 N5/AC13 claim, or only
  calls through its patched `fetch`/`WebSocket` APIs? The current source
  patches neither `XMLHttpRequest` nor `navigator.sendBeacon`.
- **Population and controls:** in the existing real-browser host test, install
  the trap, issue a `fetch` to a disallowed loopback port as the positive
  control, then issue an `XMLHttpRequest` and a `sendBeacon` to that same
  disallowed local port. No secret or external host is involved. Catch network
  errors; count violations recorded by the trap, not request success. Retain
  the existing `PaymentRequest` and `window.open` controls.
- **Firing criterion:** the fetch attempt is recorded, while an XHR or beacon
  attempt is not. That is a witness-coverage defect, not proof that current
  production shop code sends off-origin traffic. If the criterion fires,
  either make the test's transport coverage accurate with failing cases for
  each supported API or move N5 to a browser-level request observer that can
  support its literal all-requests claim; do not relabel the narrow trap as
  complete merely because production currently uses `fetch`.
- **Scope:** accepted §10 N5/AC13. No payment feature, player copy, endpoint or
  release policy change is authorized. C7's full designated verdict remains
  separate from this focused probe.

## 2026-10-04 — Codex targeted C7/N5 finding and request-audit supplement (RP-168)

- **Review by:** Codex. **Recorded by:** Codex. **Reviewed Claude range:**
  `a4b22429^..a4b22429`, limited to N5's network-witness completeness.
  **Decision: CHANGES REQUIRED for evidence.** `installNetworkTrap` patches
  `fetch` and `WebSocket` but not XHR, `sendBeacon` or resource loads; the
  C7 log's "every HTTP and WebSocket request" claim exceeded its observer.
  No current production shop egress was demonstrated.
- The predeclared Chromium probe used a disallowed local loopback checkout
  URL. It recorded `fetch` but missed both XHR and beacon: expected three
  violation labels, received only `fetch`. The test was not left red.
- A test-only Vitest/Playwright command now attaches request and WebSocket
  listeners to the actual browser page for the duration of the mocked shop
  flow. The original API trap remains for direct PaymentRequest/credentials/
  window.open refusal. A browser negative issues fetch, XHR, beacon and image
  requests to the same-origin `/checkout` path (which the game allowlist
  rejects) and requires all four to appear in the page observer. The real
  shop-flow test requires both the direct trap and page-level request list to
  be clean. Disabling only the new page `request` listener made the negative
  fail at an empty request list; it was restored. Chromium and WebKit each
  passed the three scoped host tests with the listener present.
- The first browser-observer negative used a disallowed *unserved* loopback
  port. Chromium recorded the attempts, but WebKit returned no request events
  for that unreachable destination. Replacing it with a served same-origin
  path outside the game allowlist made all four transport arms observable in
  both engines. This shows path-policy discrimination and request capture;
  it does **not** prove WebKit reports an off-origin attempt that is blocked
  before a network request. The direct trap still catches fetch/WebSocket,
  but XHR/beacon/image off-origin attempt coverage and a real-server AC14
  flow remain open before N5 can be called complete.
- `make verify-no-payment` passed with six built-in rejected fixtures and
  two near-miss controls; `make copy-check` passed; full `make test-client`
  passed 6912 with browser-only skips; `make typecheck` reported zero errors
  or warnings. These are C7 context checks, not a full AC13 or release verdict.

## 2026-10-04 — RP-168 final local gate and review coordinate

- Predeclaration `459858f6`; test-only observer, fired negatives, ledger and
  queue landed in `f58a0f6e`. Claude's designated review must cover
  `459858f6^..f58a0f6e` plus this coordinate record commit. Codex did not
  self-approve the C7 correction or Claude's original C7 range.
- Sequential final current-HEAD runs passed seven scoped host+shelf tests in
  Chromium and seven in WebKit; each invocation's performance sublane passed.
  Full client unit suite passed 6912 with 85 browser-only skips; typecheck
  reported zero errors or warnings. `make verify-no-payment` and
  `make copy-check` passed. Firefox's independent browser-session timeout
  from RP-167 remains unresolved on this host; it was not rerun for this
  C7 supplement. AC14, off-origin pre-request observation and copy adoption
  remain open.

## 2026-10-04 — RP-170 cold Linux request-audit predeclaration

- The cached ARM64 Linux Playwright image bypasses this macOS 27 host's Firefox app-data launch denial without altering the local Firefox directory. Cold `make test-browser-ci` launched all three engines but failed the new RP-168 four-transport negative: Chromium reported only one request, and Firefox reported only `fetch`, rather than all four. The full browser gate is red; prior macOS Chromium/WebKit green results do not override it.
- Diagnose whether the observer misses requests or the fixture's fire-and-forget XHR/beacon/image requests are not guaranteed to reach Playwright before audit stop. Preserve the browser-level request observer and direct payment/API traps. A correction must exercise four transport/resource attempts in the Linux browser population, require all four observed labels, retain a fired listener-severing negative, and not claim blocked-before-network off-origin attempts or real-server AC14. Avoid a fixed delay as the oracle; wait for observable completion or a bounded explicit timeout.
- This is test/evidence scope only. CSS fixture collection failure from the same run belongs to the Deployment log; neither is a product Cosmetics defect yet. Claude's cross-party designated review remains required.

## 2026-10-04 — RP-170 cold Linux request-audit result

- Isolated Linux execution of the original four-transport test passed in all three engines; the cold full suite failed Chromium and Firefox. Thus the fixed 100 ms sleep was not a valid completion condition under the full population. The correction gives each transport URL a fresh nonce, asks the Playwright-side audit to wait (bounded at five seconds) until all four exact URLs have been observed, and fails with both expected and observed URL lists on timeout. It keeps the direct payment trap and actual page request/WebSocket observer.
- Cold `make test-browser-ci` on the cached ARM64 Linux image now exits 0 with 20,976 browser tests passed (3 skipped) across Chromium, Firefox and WebKit, plus the isolated Chromium performance case. Full `make test-client` passed 6912; typecheck had zero errors/warnings. A temporary removal of the page `request` listener made the focused Linux test fail in all three engines with `observed=[]`; the listener was restored and the focused 9/9 tests passed. The Linux run demonstrates browser-engine behavior despite this macOS 27 host's local Firefox launch denial, which matches a Playwright-tracked OS app-data access problem; it does not prove an actual hosted CI job.
- This remains a test-only correction. Off-origin attempts stopped before network dispatch, the real-server Buy→reload AC14 path, G10 live pet overlay, owner copy adoption and Claude's designated review remain open. No Cosmetics or 1.0 release status changes.

## 2026-10-04 — AC14 real-server browser witness predeclaration

**Question:** does the accepted Cosmetic Shop acquisition actually work through the default browser runtime and a composed gameserver/Postgres when a bundle pins Cosmetics, and does a reload recover ownership from the server? Existing C6 tests use an injected runtime; they are not this proof. **Population:** a test-only epoch root based on the current phase-0 artifact bytes plus the accepted Reputation declaration/tree, Pet Species and Cosmetics fixtures. Its constants hash is computed with the production length-framed algorithm and is accepted only by that test-only epoch; no production epoch or catalog file is minted. Use the normal gameserver binary, real Postgres, Vite proxy and Chromium, with a fresh anonymous bootstrap. Ordinary server-side cash setup may make the T0→T1 gate eligible; Buy and all acquisition gameplay actions must originate in visible UI controls.

**Arms and required observations:** (1) before T1, the shelf is absent and the server snapshot reports the Cosmetics arm active but locked; (2) after a visible T1 cross-gate, the visible enabled Buy button emits exactly one `acquire_cosmetic` intent through `/api/v1/intents`, returns an applied receipt/event, and the owner-facing shelf changes only after the authoritative result; (3) a full page reload shows `horse_armor` owned from a fresh authoritative snapshot, not a retained session-local receipt. The entire browser flow must run under a page-level request/WebSocket origin/path audit and direct PaymentRequest, credential API and window.open traps; any payment/off-origin/unknown served path fails. **Negative:** temporarily sever the server Game UI Cosmetics projector so the browser witness fails at its Cosmetics-arm/shelf assertion, then restore exact source bytes. A test that still passes on a severed producer is invalid. Measure the added runtime against D-014's push topology; the target must remain suitable for the existing `game-ui-composed` CI job, not silently create a local-only witness.

**Limits:** this synthetic epoch is fixture-first, not a production content mint or a v0.1 manifest adoption. The witness does not prove live pet-panel overlay, owner copy adoption, keyboard/AT task completion, off-origin attempts stopped before browser request dispatch, public hosting, or a supported release bundle. If the currently accepted wire/PA7 conflict prevents server startup or snapshot decoding, record the failing boundary rather than weakening AC14.

The first driver run reached T1, applied Buy and recovered ownership on reload, but its N5 audit failed on three Vite development HMR WebSockets to `ws://localhost:5174/?token=...`; Vite also printed proxy EPIPE. These requests are injected by the test server, not the product, but whitelisting them would dilute the predeclared API-only rule. Correction before claiming a pass: build the client and serve the resulting bundle through Vite preview with the same API/WebSocket proxy. The audit should then allow the compiled `/assets/` path, not development `/@vite`/`/src` paths, and continue to reject HMR, checkout and off-origin traffic. Re-run the full flow and a severed projector on that built-client population. The first run is invalid as N5 acceptance evidence.

The built-client Vite-preview run reached Buy→reload but emitted proxy EPIPE; adding the existing composed lane's visitor-counter handshake assertion made it time out after 30 seconds. Thus the first built-client green result was HTTP-only and is retracted as a complete default-runtime witness. Vite preview's WebSocket proxy is not a valid transport carrier in this environment. Keep the compiled client and strict N5 audit, but serve it from a narrowly scoped test-only same-origin HTTP/static/WebSocket proxy that forwards the actual upgrade to the composed gameserver. Require the visitor counter before any Buy assertion. This is another pre-acceptance instrument correction, not an AC14 threshold change or a product-code fix.

The custom proxy's first run also failed its visitor handshake because the test UI used port 5174 while the checked-in transport policy permits only `http://localhost:5173`. The driver now uses that permitted origin. The full live handshake, T0 locked state, DOM T1 gate, visible Buy, applied receipt, server-owned reload and N5 request audit pass; temporarily replacing the server Cosmetics projection with nil fails at the first arm assertion and exact production bytes were restored. A later sequential run found Buy visible but transiently disabled immediately after the authoritative T1 snapshot; the driver now waits boundedly for Playwright's actual actionable Buy control, without an extra click or relaxed outcome. The exact composed Make target passes in 17 seconds on local ARM64 using an explicit Postgres image override. This is not owner copy, G10, or full AC13 approval.

The existing composed Game UI driver, which precedes AC14 in that Make target, failed intermittently on a Pitch Start/Unlock control with no emitted request. It passed from a clean DB and on several subsequent runs, then failed again; the rendered Fiscal unlock was enabled and `main[aria-busy=false]` after the timeout. The Cosmetics fixture leaves a different epoch in the same ephemeral database, so the existing driver now resets that named test DB before its own run. This removes cross-driver epoch contamination but does **not** explain the no-request Pitch flake. Do not call the combined CI target reliable from one pass. A bounded focused diagnostic is needed: temporarily instrument the Game UI `act` path for Pitch Fiscal unlock, capturing whether the click entered the handler, whether it was dropped on `actionTask`/`refreshTask` guards, and whether a request was attempted; restore every diagnostic production byte. This diagnoses an existing acceptance path; it does not authorize changing its contract under the Cosmetic RFC.

The bounded `act` instrumentation was run four times and all four passed; each observed
`spend_fiscal_credit` enter with snapshot/Founder revision present and no active action or
refresh task, then reached the post-wait guard. The intermittent failure did not fire in that
instrumented population, so it cannot establish handler entry or dropped-task cause. The
instrumentation and page-console hook were removed byte-exact. The ordinary driver had already
failed on both locked Start and Fiscal Unlock with no request, and the full two-driver Make target
failed once at Fiscal Unlock after a separate 17-second pass. RP-172 retains the defect without a
speculative product fix or test retry. The AC14 driver then tightened its direct trap to reject
disallowed fetch/WebSocket calls instead of merely recording them; its live run remained green.
`make test-client` passed 6912 tests (85 browser-only skips), typecheck reported zero errors or
warnings, and `make verify-ci-topology` passed all 13 negative fixtures. Hosted CI has not run.

**AC14 landing coordinate:** `0cf3f719^..0cf3f719` carries the test, Make/CI wiring,
docs, ledger, roadmap and AC14 plan checkbox together. **Review by:** Codex (implementer-side
first filter). **Recorded by:** Codex. The executed positive, producer-severing failure,
restored-source check, exact local Make target, typecheck/client/topology and no-payment results
above are evidence for handoff, not the mandatory designated cross-party verdict. Claude must
adversarially review this exact range plus this coordinate record before any Cosmetics archival.

## 2026-10-05 — Codex designated C5/AC10 two-wearer decoder review predeclaration

**Review by:** Codex. **Recorded by:** Codex. **Claude range under review:**
`1a477d9e^..1a477d9e`, limited to the C5 snapshot/projector/decoder contract and its own
fixtures. This is not approval of C1–C4 or C6–C8, and not an archival verdict for Cosmetics.

The Go projector sorts each `worn_by` list and `wearers` by pet ID, but the current shared fixture
has at most one wearer. The TypeScript decoder checks wearer-row order and cross-reference
membership, yet appears not to check each `worn_by` list for strict order or uniqueness. Before
changing decoder behavior, add a two-pet valid control to the client AC10 test and derive two
negative arms: reverse the same two `worn_by` IDs while preserving the reciprocal wearers, and
duplicate one ID while retaining its wearer. The accepted §7.1 requires sorted pet IDs, and a
duplicate cannot represent a set of pets. Both negatives must fail at the decoder; if they pass,
record a C5 CHANGES REQUIRED finding and make the bounded fail-closed decoder correction under
the accepted AC10 contract. Also check `lock.tier=9` against the registry's `[0,8]` domain and
the accepted catalog. Controls: the original Go-authored fixture cases still parse, and the
two-pet valid arm parses. An independent temporary severing of the new strict-order check must
make at least one negative fail, then restore exact bytes. Run cold Go Game UI and client tests,
typecheck, API compatibility generation/check; no production epoch, schema shape, copy or payment
behavior change is authorized. The normative §7.1 required-v4 wording conflicts with the
accepted API Foundation C2 optional-arm contract and the C5 implementation; file that as a
separate ruling-author body-reconciliation finding, not an implementer rewrite.

## 2026-10-05 — Codex targeted C5/AC10 verdict: CHANGES REQUIRED

**Review by:** Codex (designated cross-party reviewer of Claude's C5).
**Recorded by:** Codex. **Reviewed range:** `1a477d9e^..1a477d9e`.
**Decision:** CHANGES REQUIRED, limited to C5/AC10; C1–C4 and C6–C8 remain unreviewed by this
entry. A valid two-pet arm parsed. Three independently named negatives also parsed when they
must reject: reversed `worn_by` pet IDs, a duplicate `worn_by` ID, and `lock.tier=9`. Cold
`make test-client` failed 3/5 AC10 cases (6,912 other tests passed), so this is executable
evidence, not a source-only suspicion. The Go producer sorts the lists; the defect is that the
public-wire TS reader accepts forms that the accepted contract and API schema reject. The
bounded correction is ordered pet-ID validation and `[0,8]` tier validation in
`parseCosmeticsArm`, preserving the Go-authored cases and schema shape. It needs Claude's
independent review as a Codex-authored corrective range before C5 can close.

**Separate body conflict:** accepted Cosmetic Shop §7.1 says v4 gains a *required* top-level
`cosmetics` object, while the actual C5 patch adds an *optional* `features.cosmetics` arm under
the already existing v4 snapshot. The C5 log cites accepted API Foundation C2 as the reason, and
that optional shape is the right producer/consumer contract at HEAD. But the normative RFC body
was not reconciled in the same edit. Per the owner/ruling-author body-reconciliation rule, the
RFC's author must edit §7.1; this review cannot treat the log's explanation as a body amendment.

## 2026-10-05 — RP-173 bounded C5 decoder correction and restored checks (Codex)

**Implementer / recorded by:** Codex. This is a correction to the CHANGES REQUIRED C5/AC10
finding, not an approval of Claude's wider Cosmetics range. The Go producer fixture now contains
two pets simultaneously wearing the same cosmetic; its projector assertion checks both
`worn_by` and `wearers` are byte-sorted. The TS valid control clones that exact shared case.
Three client negatives first accepted reversed `worn_by`, a duplicate ID, and `lock.tier=9` on
the original decoder (cold `make test-client`: 3 failed, 6,912 passed). The decoder now requires
strictly increasing pet IDs within every item and a safe integer lock tier in 0–8; no wire
shape, producer behavior, product content, payment behavior or player-facing copy changed.

Final local cold checks on the generated two-wearer fixture: `make test-client` 6,915 passed,
85 browser tests skipped; `make typecheck` zero errors/warnings;
`make test-go GO_PACKAGES='./gameui ./account ./gameserver' GO_TEST_FLAGS='-count=1'` passed;
`make api-check` regenerated byte-identical contracts. With the new strict-order check
temporarily severed, `make test-client` failed exactly the reversed and duplicate cases (2 failed,
6,913 passed). With the upper-tier bound separately severed, it failed exactly tier 9 (1 failed,
6,914 passed). Both source mutations were restored by patch; the final checks and Git diff
confirm the restored implementation. `make build-client`, `make vet`,
`make verify-cosmetic-boundary` (nine negatives) and `make verify-no-payment` (six negatives,
two accepted near-misses) also pass. This is local decoder evidence, not a composed browser
journey, broad C5 verdict, or Cosmetics archival authorization. RP-174's RFC §7.1 contradiction
remains for the ruling author. Claude must designated-review the exact Codex corrective range.

## 2026-10-05 — C5 corrective review coordinate

**Recorded by:** Codex. The complete Codex predeclaration → red tests/verdict → corrected
implementation/test/fixture/docs/ledger range is `ebb88bab^..547ed7f0` (three commits:
`ebb88bab`, `44ce40c3`, `547ed7f0`). The worktree is clean at the coordinate. This is ready for
Claude's designated adversarial review; no approval is asserted. Claude's verdict must cite
the actual range it inspects. RP-174's ruling-author body reconciliation remains separate.

## 2026-10-05 — C5 reader correction composed smoke after commit

**Recorded by:** Codex. From committed `dd7dccc6`, the exact root target
`make test-game-ui-composed GAME_UI_COMPOSE_FILES='-f compose.game-ui-test.yml -f compose.game-ui-arm64.yml'`
passed against real Postgres/WebSocket and the built browser client. Both composed Game UI
drivers passed, as did Cosmetics AC14: T0 locked → visible T1 Buy → one applied server intent →
server-owned reload, with 29 N5 requests and no violation. The target took 26.6 seconds.
This confirms the corrected decoder does not break that integrated path; it does not make the
synthetic Cosmetics test epoch a production pin or prove RP-173's reversed/duplicate negatives
through a browser transport. The test-only `game-ui-postgres` service was stopped afterward;
the separate existing `postgres` service was not touched. Claude's designated review and
RP-174 body reconciliation remain open.

## 2026-10-05 — Codex designated C1/AC1 raw-number parity review predeclaration

**Review by / recorded by:** Codex. **Claude range under review:** `afe6529b^..afe6529b`,
limited to C1 catalog loading, its shared corpus and AC1's exact-integer/Go–TS parity claim.
This is not review or approval of C2–C8. The Go loader decodes `unlock.tier` into `int64` from
raw JSON, whereas the TypeScript loader calls `JSON.parse` before checking `Number.isSafeInteger`.
Predeclared hypothesis: TS admits alternate JSON number lexemes `1.0` and `1e0` for a tier that
Go rejects, and may do the same for `schema_version:1.0`; the accepted §2 contract calls for
an exact integer and both loaders to enforce the same grammar. Add these as shared negative
corpus arms, retaining the ordinary `tier:1` and `schema_version:1` positive controls. Run cold
`make test-go GO_PACKAGES='./cosmetic' GO_TEST_FLAGS='-count=1'` and `make test-client` before
changing either loader. A Go-pass/TS-fail split is a C1 CHANGES REQUIRED finding. If the split
exists, correct the raw TS loader only, preserving the object-level parser's stated limits;
require the shared controls, client typecheck, package-boundary/no-payment gates, and a
temporary severing that makes at least one new negative fail again. Do not invent a pricing
mechanic, edit owner-authored text, mint a production content epoch, or treat a narrow C1
finding as full Cosmetics acceptance.

## 2026-10-05 — Codex targeted C1/AC1 verdict: CHANGES REQUIRED

**Review by:** Codex (designated cross-party reviewer of Claude's C1).
**Recorded by:** Codex. **Reviewed range:** `afe6529b^..afe6529b`.
**Decision:** CHANGES REQUIRED for C1/AC1 only. A positive integer-token catalog remains
valid. The shared corpus gained three raw-number negative arms: tier `1.0`, tier `1e0`, and
schema `1.0`. Cold `make test-go GO_PACKAGES='./cosmetic' GO_TEST_FLAGS='-count=1'` passes all
cases. Cold `make test-client` initially stopped at the first TS acceptance. Changing the
corpus runner to independent per-case tests exposed all three: 3 failed, 6,940 passed,
85 browser tests skipped. The TS `JSON.parse` normalizes all three lexemes to numeric 1 before
`Number.isSafeInteger`, so this is loader-parity and exact-grammar failure, not a payment or
mechanics path. A bounded correction may validate integer lexemes in the raw TS loader;
`parseCosmeticCatalog(unknown)` cannot reconstruct lost lexical form and should not claim to.
The correction requires its own cross-party designated review by Claude; this verdict does not
approve C2–C8 or archive Cosmetics. RP-175 tracks the defect.
