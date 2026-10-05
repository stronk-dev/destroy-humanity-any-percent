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

## 2026-10-05 — RP-175 bounded C1 raw-loader parity correction (Codex)

**Implementer / recorded by:** Codex. The TS `loadCosmeticCatalog(bytes)` scanner now retains
the JSON number lexeme for exactly `schema_version` and `items[].unlock.tier` before
`JSON.parse` erases it, rejecting fractional or exponent spelling for those exact-integer
fields. The object-level `parseCosmeticCatalog(unknown)` still checks safe integer *values*;
it cannot recover a lexeme already discarded by its caller and is not claimed to. The existing
unknown-key rule remains the grammar gate for fields outside those two positions. No Go
cosmetic production bytes, catalog shape, content epoch, payment behavior or player-facing copy
changed. Because `client/src/cosmetic/` is kernel-guarded, the behavior change advances the
shared kernel identity from 0.3.141 to 0.3.142 in the same commit; this is not a claim that
the historical pushed kernel-history failure at `50a3a514` is repaired.

Cold shared-corpus run: `make test-go GO_PACKAGES='./cosmetic ./replaycatalog ./production ./kernel'
GO_TEST_FLAGS='-count=1'` passed; final `make test-client` passed 6,944 tests with 85 browser skips;
`make typecheck` reported zero errors/warnings; `make verify-cosmetic-boundary` rejected its
nine negatives; `make verify-no-payment` rejected six negatives with two accepted near-misses;
`make build-client` produced the client bundle; `make vet` passed; and `make api-check`
regenerated byte-identical API contracts.
With only the new exact-integer-token condition temporarily severed, cold `make test-client`
failed exactly the three new number-lexeme cases (3 failed, 6,940 passed), then the code was
restored by patch. A separate unknown-field control uses `price:1.5` and requires the
unknown-key error rather than a numeric-token error; broadening the new token check to every
number failed that case (1 failed, 6,943 passed), and the path restriction was restored.
This is loader parity evidence, not a complete C1/C2–C8 approval or
Cosmetics archival gate. Claude's designated review of the exact Codex corrective range is due.

## 2026-10-05 — C1 corrective review coordinate and kernel gate

**Recorded by:** Codex. The C1 Codex predeclaration → red shared cases/verdict → corrected
raw loader/tests/docs/kernel identity range is `7644808d^..76a9fe04` (three commits:
`7644808d`, `8c447b80`, `76a9fe04`). The behavior-changing commit itself moves
`kernel/VERSION` and both generated mirrors 0.3.141→0.3.142 with
`client/src/cosmetic/catalog.ts`. The post-commit `make verify-kernel-version` still exits 2
at pre-existing pushed `50a3a514`, before it reaches this new commit; the CI kernel-history
contract and adversarial fixtures pass. The new same-commit version bump is directly checked
against `76a9fe04^..76a9fe04`, not claimed as a green full history walk. This C1 range needs
Claude's designated adversarial review. The separate C5 corrective range and RP-174 ruling-
author body conflict remain open; no Cosmetics archival approval is asserted.

## 2026-10-05 — Codex designated C2 replay-bundle review predeclaration

**Review by / recorded by:** Codex. **Claude range under review:** `8e315569^..8e315569`,
limited to C2/OD-10's `cosmetics` artifact membership, constants identity, scalar Founder
dependency on `pet_species`, and the permanent-ID settlement hook. This is not approval of C1
or C3–C8. At current HEAD, run cold `make test-go
GO_PACKAGES='./replaycatalog ./production' GO_TEST_FLAGS='-count=1'` and `make test-client`.
The positive control is a full Go/TS bundle containing the pinned Horse Armor fixture; negative
arms remove cosmetics, remove `pet_species`, add a payment-shaped field, and drop a current ID
from the next bundle. Independently sever the Go and TS `cosmetics ⇒ pet_species` loader checks:
each should fail its own C2 test on the missing-species arm. Independently sever the server
permanent-ID settlement check: the dropped-artifact or dropped-ID test must fail. Restore every
mutation and rerun cold. Inspect whether hash/byte binding witnesses can fail if cosmetics is
omitted from identity. Any vacuous negative or cross-runtime discrepancy is CHANGES REQUIRED;
otherwise record a bounded C2 verdict with exact range and limits. No product mechanic, epoch
pin, body text or copy change is authorized by this review.

## 2026-10-05 — Codex designated C2/OD-10 verdict: APPROVED (bounded)

**Review by:** Codex, the cross-party designated reviewer of Claude's C2.
**Recorded by:** Codex. **Reviewed range:** `8e315569^..8e315569`.
**Decision:** APPROVED for C2's replay-bundle wiring, identity membership, scalar dependency
and permanent-ID settlement hook only. Source and range diff were inspected against accepted
Cosmetic Shop §2/OD-10 and the existing generic constants-hash/`CatalogBundle.valid` contract.
At current HEAD, cold `make test-go GO_PACKAGES='./replaycatalog ./production'
GO_TEST_FLAGS='-count=1'` and `make test-client` pass (6,944 client tests, 85 browser skips).

Independent temporary mutations were each restored before the final cold pass:

- Removing Go's `cosmetics ⇒ pet_species` artifact-name check made
  `TestLoadCosmeticsRequiresItsChain` fail, "cosmetics loaded without pet_species".
- Removing the corresponding TS replay-artifact check made the client cosmetic-bundle negative
  fail because the missing-species bundle loaded.
- Removing `cosmetic.ValidateTransition` from server settlement made
  `TestSettleRejectsCosmeticIDDrop` fail on the dropped-artifact arm with a nil error.
- Omitting `cosmetics` from Go `ConstantsHashArtifacts` made the C2 identity test fail with
  "cosmetics does not join constants identity". Omitting it from the TS replay hash made the
  focused two-test client C2 population fail at the complete-bundle positive with "replay
  artifact label mismatch" (one failed, one passed). The broader client suite also failed
  56 dependent replay tests under that mutation; the focused result is the cited C2 oracle.

The restored source diff is empty. This approval is not C1/AC1 or C3–C8 review, not proof a
production epoch pins Cosmetics, and not Cosmetics archival authorization. RP-175/C1 and
RP-173/C5 Codex corrections still need Claude's designated review; RP-174 still needs its
ruling-author RFC body reconciliation.

## 2026-10-05 — Codex C3/AC3–AC4/AC8 migration-corpus review predeclaration

**Review by / recorded by:** Codex. **Claude range under review:** `581886a4^..581886a4`,
limited initially to §3/§6 Founder v24 codec, activation and the exact AC4 migration-corpus
requirement; C4's later Exit corpus is a dependency witness, not retroactive proof that a named
`testdata/save-migrations.json` case exists. At current HEAD, enumerate the actual corpus names
and baseline count, execute cold `make test-go GO_PACKAGES='./save ./production'
GO_TEST_FLAGS='-count=1'` and `make test-client`, and inspect the `RestoreState`/new-run
activation boundary. Compare the accepted §6 names after OD-16's next-free-version substitution
to real files and tests. A missing named acceptance population or a contradictory migration
instruction is a C3 CHANGES REQUIRED finding; do not silently replace it with a different test
or invent mid-run activation. A bounded correction may proceed only if the accepted authority
unambiguously defines the expected transition. Ruling-author body edits remain with that author;
no product code or owner-authored content change is authorized by this audit.

## 2026-10-05 — Codex targeted C3/AC4 verdict: CHANGES REQUIRED

**Review by:** Codex (cross-party designated reviewer of Claude's C3).
**Recorded by:** Codex. **Reviewed range:** `581886a4^..581886a4`, limited to the §3/§6
codec/activation and AC4 corpus boundary. **Decision:** CHANGES REQUIRED for the accepted
AC4 evidence contract; this is not a demonstrated live Founder activation failure and not a
verdict on C4–C8. The actual `testdata/save-migrations.json` and its baseline contain exactly
11 legacy cases, the newest from v9. Two independent `jq -e` exact-name checks exit 5: all
five literal §6 v21/v22 names are absent, and all five OD-16-adjusted v23/v24 names are absent.
Cold `make test-go GO_PACKAGES='./save ./production' GO_TEST_FLAGS='-count=1'` and
`make test-client` pass (6,944 client tests, 85 browser skips); that green population never
asserts the missing cases.

The implementation does have current-coordinate evidence: `TestFounderV24CosmeticsRoundTripAndInvariants`
rejects malformed v24 saves, `TestCosmeticsOwnsFounderV24Activation` activates at settlement,
and the later `testdata/replay/cosmetic-v1.json` case `exit-activates-founder-v24` carries the
Go/TS Exit arm. Those do not silently replace the RFC's explicit `testdata/save-migrations.json`
deliverable. The existing save-corpus harness calls `RestoreState(data, version, economyCatalog,
scope, baseline)` and compares a re-encoding; it has no pinned content bundle. `RestoreState`
decodes v23 and v24 separately and rejects Cosmetics before v24, while §6 also requires
activation **only** at a new-run boundary. Interpreting `founder-v23-to-v24-empty` as an ordinary
read-time migration would violate that rule. The ruling author must reconcile §6/AC4's corpus
home, version coordinates and expected transition before an implementer can honestly close it.
RP-176 tracks this distinction. C3 is not approved or archival-eligible on this verdict.

## 2026-10-05 — Codex designated C4/AC9 database-payload review predeclaration

**Review by / recorded by:** Codex. **Claude range under review:** `500d944c^..500d944c`,
limited to §4.5/AC9's claim that Postgres rejects an extra cosmetic event payload field; the
rest of C4's intent and replay corpus is not approved by this bounded review. The current
`events` SQL has an object payload check and a closed kind list, while the C4 log says Go's
strict decoder, not the database, rejects extra fields. Before a correction, add a retained
real-Postgres test inside the existing Cosmetic integration population. For each of acquired,
equipped and unequipped, a valid payload/kind update in a rollback-only transaction must
succeed, and an otherwise identical payload with an extra `price` or `amount` key must fail at
the SQL boundary. Run cold through the declared Compose Postgres service. If the valid control
passes and the extra update succeeds, record C4 CHANGES REQUIRED. Under accepted AC9, a
correction may add a NEW migration (never edit applied `00080`) constraining exact key sets
for only those three kinds; it must leave other event kinds unchanged and prove valid cases,
all three extras, and a temporary severing failure. Also run migration validation, focused
Go/client replay, and the relevant existing integration population. Do not infer full C4,
Cosmetics or release approval from this database-only boundary.

## 2026-10-05 — Codex targeted C4/AC9 verdict: CHANGES REQUIRED

**Review by:** Codex (cross-party designated reviewer of Claude's C4 database claim).
**Recorded by:** Codex. **Reviewed range:** `500d944c^..500d944c`, limited to AC9's
Postgres extra-field rejection. **Decision:** CHANGES REQUIRED. A retained test inside
`TestCosmeticIntegrationPersistsReplayableFounderLog` uses a transaction that is rolled back
for each cosmetic event kind. The valid acquired, equipped and unequipped payload updates all
succeeded; each otherwise identical `price`/`amount` extra-field update also succeeded, so
three named subtests failed cold on real Postgres. Go's `validateEventPayload` still rejects
the same extras before insert; it cannot satisfy AC9's explicit *database* assertion.

The first Compose attempt did not run the test: cached `postgres:16-alpine` was amd64 on this
ARM64 host and exited 255 with `exec format error`. A test-only
`compose.save-test-arm64.yml` override selects the already-local ARM64 Postgres 16 image,
leaving the Go test image and hosted x86 Compose default unchanged. The same declared
`docker compose ... run --rm test go test ... -count=1` population then reached Postgres and
failed at exactly the three extra-field assertions. RP-177 records the defect. Corrective
scope is a new post-00083 migration with exact cosmetic payload key sets and positive/negative
database evidence; applied `00080` is immutable. This is not full C4 approval.

## 2026-10-05 — Codex bounded C4/AC9 corrective implementation

The retained AC9 database witness from the red review now runs against append-only migration
`00084_cosmetic_event_payload_keys.sql`. The new constraint requires exactly the accepted key
set for each of `cosmetic_acquired.v1`, `cosmetic_equipped.v1`, and `cosmetic_unequipped.v1`; it
does not change other event kinds or validate cosmetic value semantics, which remain in Go.
Applied `00080` was not edited. `docs/cosmetics.md` now describes both DB and Go boundaries.

Cold real-Postgres evidence: all three valid payloads pass; each extra `price`/`amount` key and
each missing required key rejects; an unrelated `generator_purchased` payload remains accepted.
The existing Founder replay history still verifies after every rollback-only probe. A temporary
`ALTER TABLE events DROP CONSTRAINT events_cosmetic_payload_keys_check` inside the acquired
test's own transaction made its extra-field assertion fail with `database accepted an extra
field on acquired payload` (exit 1); rollback restored the constraint, and the unmodified
verbose run passes all six named extra/missing subtests. `Integration|Migration` on `./save`
and `Integration|Cosmetic` on `./production` passed cold through the declared Postgres Compose
service with the local ARM64 image override. Focused `./save ./production ./replaycatalog`
passed with `-count=1`; `make test-client` passed 6,944 tests with 85 browser skips;
`make vet`, `make typecheck`, `make api-check`, `make verify-no-payment`,
`make verify-cosmetic-boundary`, `make verify-ci-topology`, and the replay-fixture check passed.
The first full CI server-core run against real Postgres found another migration consumer:
`server/releasepackage/manifest_test.go` still asserted the previous maximum `83`, so
`TestCurrentMigrationIsContiguous` failed with `migration=84` and the target exited 2. The
bounded correction updates that pin to `84`. A second cold full CI server-core run through the
declared Postgres Compose service passed (exit 0): vet; every non-harness Go package at
`-count=1`, including production, save, releasepackage and transport against real Postgres;
pitch content, formula generation/check and API generation/check. The command used was:

```sh
docker compose -f compose.save-test.yml -f compose.save-test-arm64.yml run --rm test sh -c 'cd /workspace && make verify-server-core'
```

It reproduces the CI server job's target locally on ARM64 with only the Postgres image
override; hosted Actions has not run on this unpushed
correction. No green *whole-CI* claim follows.

This is a **Codex implementation and self/first-filter evidence**, not the designated
independent review of Codex's range. Claude must review its exact commits before AC9 is
eligible for closure. Codex's bounded CHANGES REQUIRED verdict on Claude C4 remains in force
for the original range, and the rest of C4 is still unreviewed by that verdict. C3/AC4 remains
blocked on RP-176; no production Cosmetics pin, RFC archival, or release claim follows.

## 2026-10-05 — Codex C4/AC5–AC6 intent/replay review predeclaration

**Review by / recorded by:** Codex. **Claude range:** `500d944c^..500d944c`, now examined
at current HEAD after the bounded RP-177 database correction. This review is limited to AC5
and AC6's acquisition/equip/unequip semantics and Go/TS replay parity, not AC9's pending
Claude review of Codex's SQL correction or the rest of C4. Read §4 and AC5–AC6 against the
actual request grammar, server authority, event and receipt producers, fixture population and
TS consumer. Verify cold Go/TS shared corpus and real-Postgres intent/retry path. Predeclared
attacks: remove the T0 unlock rejection in each runtime and require the locked corpus case to
fail; independently sever equip's owned check or pet-key check and require a corresponding
fixture to fail. If the corpus cannot distinguish a named AC5/AC6 failure, record a defect,
not an approval. Any new test or correction stays in a separate bounded Codex range for Claude
cross-party review. A passing bounded verdict will not archive Cosmetics while C1/C3/C5,
AC9's correction, owner copy and G10 remain open.

## 2026-10-05 — C4/§4.5 active-recovery probe addendum

The C4 source claims cosmetic intents never receive `exclusive_activity`, but
`Service.Handle` runs `soulRecoveryExclusivity` before the cosmetic branch, and production
composition supplies `WithSoulRecovery`. Predeclare a real-Postgres diagnostic before editing
that router: use the existing C4 integration fixture as the no-session control; on the same
service with a real active Soul recovery session started through `StartSoulRecovery`, submit
the otherwise eligible Tier-1 Horse Armor `acquire_cosmetic`. §4.5 predicts an applied
cosmetic intent and one event, not `exclusive_activity`. If the active-session arm fails
there, record C4 CHANGES REQUIRED on AC5/§4.5 and retain the failing-first test. Any fix
must preserve exclusivity for ordinary gameplay intents, have a severing failure, and receive
Claude's designated review. This remains a bounded C4 review, not permission to change Soul
recovery policy broadly.

## 2026-10-05 — Codex C4/AC5–AC6 targeted verdict: CHANGES REQUIRED

**Review by:** Codex (designated cross-party reviewer of Claude C4).
**Recorded by:** Codex. **Reviewed range:** `500d944c^..500d944c`, limited to §4.5's
exclusive-activity claim and AC5's eligible acquisition; AC9's Codex correction and the rest
of C4 remain outside this verdict. **Decision:** CHANGES REQUIRED on RP-178. Production
`server/gameserver/composition.go` supplies `WithSoulRecovery`, and `Service.Handle` checks
Soul-recovery exclusivity before dispatching to `handleFounderCosmetic`. The latter's
unguarded `ApplyFounderLogged` call cannot override the earlier rejection. The existing
no-session Postgres C4 witness passed at `ce7688a7`; adding a real service-started active
recovery session to that integration population makes Tier-1 Horse Armor acquisition fail cold
with `not_eligible/exclusive_activity` (exit 1), rather than apply as §4.5 says.

The other predeclared attacks are informative but not a C4 approval: removing the Go unlock
check makes the Go corpus fail at `rejects-locked-at-tier-0`; removing the TS unlock check
makes two TS cases fail. Removing the Go equip pet-key or owned check also fails the corpus,
but via the saved-state validator rejecting an equipped unknown pet or unowned item, so those
failures are defense in depth rather than proof that the named rejection oracle fired. All
temporary production mutations were restored and `git diff --exit-code` verified that.
One initial Go selector invocation ran **no tests** because Make consumed the `$` anchor in
`GO_TEST_FLAGS` (the already-ledgered RP-025 class); the corrected selector without `$`
executed and failed as stated. Do not cite the no-op invocation as evidence.

## 2026-10-05 — Codex C4/§4.5 recovery-exemption correction

The correction exempts the three `isCosmeticIntent` kinds from `Service.Handle`'s Soul-recovery
preflight under accepted §4.5. Ordinary commands retain the preflight and transaction guards.
The retained real-Postgres integration test uses production's `WithSoulRecovery` option and
starts the pinned `defrag` activity through `StartSoulRecovery`. It checks that an ordinary
manual command returns `exclusive_activity`, then acquisition applies with its original retry,
conflict and rejection cases. A fixture pet permits equip and unequip while the same session
is active. Server-clock progress beats complete the pinned recovery duration; resolution
persists the terminal session, advances the Founder to revision 5, carries the final cosmetic
state unchanged and leaves Founder history `ReplayVerified`. Canonical Cosmetics docs now
describe the dispatch exemption.

The first added resolution attempt failed with `invalid replay inputs: Fiscal sweep`: its
September fixture clock preceded the database-clock Founder command. That was an invalid
test-time coordinate, not evidence of a live recovery defect. The fixture now anchors to
`SELECT clock_timestamp()` like the command producer, and the full acquisition → equip →
unequip → recovery-resolution path passes cold. It uses the pinned recovery duration and beat
ceiling rather than reducing either to make the check cheap.

Three restored severing probes reject the intended outcomes: removing the entire new exemption
fails acquisition as `exclusive_activity`; exempting only acquisition fails equip; exempting
acquisition and equip fails unequip. Every failed probe exits 1 on real Postgres. The original
exemption is restored. Kernel `0.3.142` → `0.3.143` and both mirrors accompany the guarded
`server/production/intents.go` behavior change in this correction.

The cold Postgres CI server-core target passes: vet and every non-harness Go package at
`-count=1`, including the existing Soul recovery population, production, save, kernel and
transport; formula and API artifacts are unchanged. The final verbose C4 integration run
passes all six SQL payload subtests and the completed recovery workflow. `make test-client`
passes 6,944 tests with 85 browser skips, `make typecheck` reports zero errors/warnings,
and the production client build passes.
This is local implementation and first-filter evidence. Hosted CI has not run on this
unpushed range, and the historical kernel-history RP-131 failure remains separate.

**READY FOR CLAUDE DESIGNATED REVIEW:** the corrective span begins at `380d854b^` (the retained
red test/verdict) and ends with this implementation plus its exact-range checkpoint. Codex
does not approve its own correction. C4's remaining acceptance evidence, RP-176/§6 and
RP-174/§7.1 body reconciliation, other Cosmetic corrections, copy adoption and G10 remain
open; this entry authorizes no archive or release.

## 2026-10-05 — Exact corrective ranges ready for Claude

**Review needed by:** Claude (designated cross-party reviewer). **Recorded by:** Codex.
The SQL/RP-177 correction is `fca686c5^..ce7688a7`: predeclaration, the retained red SQL
witness/ARM64 override, new append-only migration 00084, required/extra/unrelated controls,
release-package migration pin, docs and tracking. The Soul-dispatch/RP-178 correction is
`380d854b^..646daa08`: retained red active-recovery witness/verdict, the bounded exemption,
all three live cosmetic intents, ordinary-intent control, recovery completion/history, database
clock fixture, kernel 0.3.143 mirrors, docs and tracking. These exact spans cover the tests,
code and records; Codex's own executed checks remain the first filter. Neither range carries
a Claude verdict yet.

## 2026-10-05 — C4/AC6 second-wearer evidence gap and predeclaration

**Review by / recorded by:** Codex. **Claude range:** `500d944c^..500d944c`, bounded to
AC6's explicit two-pet equip criterion. The current shared corpus has 20 Founder cases
(plus one Company/Founder Exit pair), with a maximum of one pet and one equipped wearer.
The original C4 log's "21 Founder cases" count was inaccurate. Go's
`TestOneCosmeticManyWearers` validates an already-built shape; the later two-wearer snapshot
fixture exercises C5's projection/decoder, not AC6's applied transition in both runtimes.
RP-179 records this missing acceptance evidence; there is no demonstrated live sharing bug.

Before changing product code, predeclare a test-only correction under accepted §4.3/OD-15
and AC6: pin an independently named `applies-equip-second-wearer` case in the shared TS
population, which must initially fail while the row is absent; add a second fixture pet through
the Go adoption transition, then equip the same owned item and assert that both pets retain
it. Regenerate the Go-authored corpus and require byte-equal TS replay. A temporary Go and
TS equip mutation that removes earlier wearers of the same item must fail the new population;
restore each mutation before committing. Run cold focused Go and the full client population.
No simulation or catalog change is authorized by this evidence correction. Its exact Codex
range requires Claude designated review, and the original C4 remains CHANGES REQUIRED on
its named findings.

## 2026-10-05 — C4/AC6 shared second-wearer evidence correction

The retained required-row check first failed the full client population because
`applies-equip-second-wearer` was absent (6,943 passed, one failed, 85 skipped). The test-only
supplement adds five Founder transitions on an independently rehashed fixture bundle: acquire
Horse Armor once, adopt two distinct pets using the existing two nonces, equip the first and
equip the second. Go asserts exactly one owned item, two retained wearers, and a non-replacing
second equip event. The shared TS replay byte-matches all receipts, events and states, with
an additional explicit one-wearer → two-wearer assertion. The corpus now has 25 Founder
cases plus the unchanged Company/Founder Exit pair.

The standard fixture's adoption cap is one. Only this new population's in-memory artifact
has cap two; both ordinary loaders consume its rehashed artifact. No production catalog,
existing fixture source, mechanical behavior or copy changed. The old cases and bundles
remain unchanged in the generated JSON; five rows and one bundle were added.

Selective severing was executed separately: a temporary Go equip loop deleting earlier
wearers fails `TestCosmeticCorpus` with `second equip did not preserve both wearers of the
single owned item`; the corresponding TS mutation fails the second-wearer shared receipt
and the explicit two-wearer assertion (two failed, 6,948 passed, 85 skipped). Both mutations
were restored; `git diff --exit-code` confirms no residual diff in either production file.

After restoration, cold focused Go Cosmetic tests and the full client population pass
(6,950 passed, 85 browser skips). Typecheck has zero errors/warnings. The Go generator
regenerates and then verifies the pinned corpus. These are local first-filter checks;
the earlier cold Postgres server-core run remains the evidence for the separate SQL and
recovery corrections. This test-only range does not claim fresh hosted CI, repair the
historical kernel-history failure, or approve the wider C4 range.

**READY FOR CLAUDE DESIGNATED REVIEW:** predeclaration commit `4b319313` plus this test/fixture
and tracking supplement; the next checkpoint pins the final hash. No Codex self-approval,
archival or live sharing bug is asserted.

## 2026-10-05 — RP-179 exact test-only review span

**Review needed by:** Claude (designated cross-party reviewer). **Recorded by:** Codex.
The RP-179 corrective range is `4b319313^..144f5e8e`: predeclared acceptance gap and exact
earlier correction ranges, retained required-row negative, Go corpus generator/adoption fixture,
five additional shared transitions, explicit TS second-wearer check, and synchronized tracking.
No product code or existing content artifact changes. The demonstrated Go/TS mutations were
restored before the implementation commit. This checkpoint does not supply a designated verdict.

## 2026-10-05 — C8/AC7 mechanical-isolation review predeclaration

**Reviewer:** Codex, designated cross-party review of Claude `c20d23f1^..c20d23f1`,
bounded first to AC7 and its package isolation gates. Verify the actual two-arm fixture,
all three intent application guards, non-cosmetics state comparison, the frozen
Founder→Company contribution comparison and how Company commands consume that channel.
Run the 200-seed policy and existing injected-leak negative cold. Independently perturb
the contribution producer conditional on cosmetic ownership, rather than only a Founder
field, and require the honest population to fail on the ruled isolation outcome. Probe the
package boundary gate with a disallowed dependency in a scratch fixture. Restore mutations;
no product correction or C8 approval follows from reading the tests. If a probe survives,
ledger the exact population/oracle failure before changing its contract.

## 2026-10-05 — C8 isolation probe and C7/N8 gate finding

**Review by / recorded by:** Codex. The independent contribution-producer mutation changed
the Reputation factor to `2e0` only when a Founder owns a cosmetic. The honest 200-seed
population immediately failed at seed 0 with `frozen Founder contributions diverged`.
The mutation was restored. An initial parenthesized Make test selector was rejected by the
shell before execution; the corrected simple selector runs the intended population. One
verbose broad run overlapped editing the contribution probe and is not used as clean baseline
evidence; the final restored run is required instead. Its DB test skips without Postgres, and
no integration claim follows from that host run.

RP-180 records a different failure in the package gate, authored by Claude in
`a4b22429^..a4b22429` (C7): a temporary `client/src/cosmetic/boundary-probe.ts` function
returning `import("../economy-kernel")` passes `make verify-cosmetic-boundary` with exit 0
and the false output `3 client files import siblings only`. The actual file was removed;
the existing production package has no such import. The current regex omits dynamic imports,
and its starts-with check also does not resolve `./../` escapes. This is targeted
**CHANGES REQUIRED** on C7/N8's isolation gate, not a blanket verdict on C7 or C8.

Predeclare the accepted §5/N8 correction before editing the gate: use the already-installed
TypeScript compiler's syntax tree (as the existing Combat boundary gate does), covering
import declarations, re-exports, dynamic import calls and import-equals. Reject computed
specifiers and package escapes after normalization; scan nested TypeScript sources instead
of only the immediate directory. Retain the former accepted dynamic-import fixture and
path-escape negatives; demonstrate they fail first. After correction, recreate the actual
disallowed source probe and require a nonzero gate exit, then restore. Run the gate, client
population and typecheck cold. No runtime dependency, mechanic or kernel source changes;
Claude designated review of this tooling correction remains mandatory.

## 2026-10-05 — RP-180 syntax-aware tooling correction

Two retained negatives failed separately before correction: dynamic
`import("../economy-kernel")`, then the otherwise static `./../economy-kernel` escape
(by reordering the fixtures so one red did not mask the other). The corrected gate uses
the existing installed TypeScript dependency, with parse errors/computed specifiers failing
closed. Import declarations, re-exports, dynamic/require calls, import-equals and import
types share a normalized package-contained path check. Recursion covers nested JS/TS module
extensions, rejects symlinks and refuses an empty source population. Nested committed fixtures
include both an allowed package-local parent import and an escaping dynamic import.

After correction the actual top-level temporary economy-import source fails the full gate;
a second actual nested `.mts` probe also fails. Both files were removed. With probes restored,
the gate passes 22 negatives and the real two-file package. No-payment passes its six negatives
and two near-misses. Cold full client tests pass (6,950 plus 85 browser skips); typecheck has
zero errors/warnings. The client Actions job already calls `make verify-client`, which invokes
this gate; no workflow or runtime dependency change was needed. This is local evidence, not
a hosted CI result; the separate historical kernel-history failure remains.

The final restored cold verbose Go Cosmetic run passes AC7's 200-seed population and its
existing injected Fiscal-input negative. The independent cosmetic-owned contribution mutation
previously failed `frozen Founder contributions diverged`. Company arms still invoke the same
Company commands with no contribution provider; their equality alone is not integrated proof
of the frozen-channel consumer. The contribution comparison and source boundary provide the
named narrow evidence here; do not quote the Company equality as an integrated multiplier
witness. This entry does not approve the full C8 docs/range or supersede C7's targeted gate
finding. The host run explicitly skips Postgres, and no new DB result is claimed.

**READY FOR CLAUDE DESIGNATED REVIEW:** the RP-180 correction begins at `c09e7e9a^` and
includes this tooling/docs/fixtures/tracking commit; the next checkpoint names its final hash.
No runtime source, kernel source, production content, owner copy or RFC body was changed.

## 2026-10-05 — RP-180 exact tooling review span

**Review needed by:** Claude (designated cross-party reviewer). **Recorded by:** Codex.
The RP-180 correction is `c09e7e9a^..ad789790`: the recorded actual-source bypass and
predeclaration, syntax-aware scanner, dynamic/normalized-path/recursive fixture negatives,
docs and synchronized tracking. The recreated disallowed source probes and contribution
mutation were restored. This exact-range checkpoint is not a review verdict and grants no
archival authority. RP-179's separate test-only range remains `4b319313^..144f5e8e`.

## 2026-10-05 — Fresh local client CI-target result

After `ad789790`, `make verify-client` passes typecheck, the production Vite build,
6,950 client tests (85 browser skips), shell/UI boundary and CI checkout-history fixtures.
It then fails the kernel-history gate at the same historical pushed commit
`50a3a5141d5c4c1cd25cbf3c9c39d9e3cf8e2444` against `0cf9f7a6`: six minigame paths changed
without a real kernel bump. Exit 2; **the complete client CI target is not green**. This
reconfirms RP-131 at the current checkpoint rather than inheriting a past summary. Make stops
there, so later prerequisites are not claimed as executed through this target; the new
Cosmetic boundary and no-payment gates were separately executed and passed. No history
rewrite, silent gate exception, push or hosted-run claim follows.

## 2026-10-05 — C3/C4 AC8 Exit-carry evidence predeclaration

**Review by / recorded by:** Codex, bounded designated review of Claude C3
`581886a4^..581886a4` and C4 `500d944c^..500d944c` on AC8. RP-181 records missing
named-path evidence: the shared corpus only exits with empty pre-activation cosmetics;
the nonempty test invokes an output-copy helper directly, and C4's reported failing
case mutates adoption rather than an Exit arm. This is not a proven live persistence bug.

Predeclare a test-only correction under accepted §4.5/AC8, independent of RP-176's
unreconciled migration-corpus home: pin independently named `exit-wind-down-preserves-owned-equipped`
and `exit-accept-offer-preserves-owned-equipped` cases, requiring both names before generation.
Each uses a v24 Founder with a valid adopted pet and nonempty owned/equipped cosmetics;
the Company terminal and Founder audit arm must both retain exactly the input cosmetics.
Keep the original activation Exit unchanged. Generate via Go and replay byte-equal in TS.
Add real-Postgres Service.Handle coverage for the two accepted intent kinds, persistence,
idempotent retry and verified Founder history. Mutating an Exit arm to reset cosmetics
must fail independently in Go, TS and the persisted witness; restore every mutation.
Run focused cold Go, client/type checks and the real-Postgres population. No product,
catalog, copy or ruling-author RFC body change is preauthorized by this evidence supplement.
Claude designated review remains required before closure.

## 2026-10-05 — C3/C4 AC8 targeted verdict and persisted proof supplement

**Review by:** Codex (designated reviewer of Claude). **Recorded by:** Codex.
**Reviewed ranges:** `581886a4^..581886a4` and `500d944c^..500d944c`, limited to AC8's
two named Exit paths. **Verdict:** CHANGES REQUIRED on RP-181's missing acceptance evidence,
not a proven live ownership-loss bug and not a blanket C3/C4 verdict. The retained required-row
guard initially failed the full client population because the nonempty wind-down case was
absent (6,950 passed, one failed, 85 browser skips). The original corpus had only the empty
activation Exit; a supplied-state copy test and an adoption mutation did not prove AC8's paths.

The test-only correction now retains `exit-wind-down-preserves-owned-equipped` and
`exit-accept-offer-preserves-owned-equipped` in the Go-authored shared corpus. Both have a
valid fixture pet and one owned/equipped item. The Company terminal's output and the Founder
audit replay's saved bytes independently preserve that same nonempty input. Go has separately
named subtests, and TS byte-matches both output arms and checks the nonempty input. The existing
25 Founder cases and original activation Exit are unchanged; only two Exit pairs were added.
The optional test-factory configuration keeps the existing Reputation corpus unchanged.

`TestCosmeticExitCarryIntegration` proves each path through real Postgres. The Founder starts
with a seeded adopted pet but **empty ownership**; actual service acquisition and equip create
the input, before the named Exit advances Company run 2→3 and Founder revision 3→4. Reload
preserves cosmetics exactly, retry returns the same receipt without extra history entries,
and both the persisted Founder history and Company run log verify. The service uses the real
database-backed frozen-contribution provider and minigame activity repository. Its first
attempt lacked the activity resolver and failed `minigame activity resolver unavailable`;
that was an invalid test composition, not a game defect, and the correction supplies production's
actual resolver rather than loosening the guard. The clock is anchored to Postgres as in the
earlier recovery proof. Controlled offer state and the adopted pet remain test fixtures, not
claims about live content minting or a default-browser journey.

Independent reset probes were run and restored:

- Clearing cosmetics inside Go `finishExitResolved` fails **both** pure named-path subtests
  with `Company terminal lost cosmetics`. The Founder audit arm by itself still preserved the
  state, which is why the independent Company-output assertion matters.
- The same production reset fails **both** persisted service subtests at the existing live
  Founder/Exit parity guard (`parity at cosmetics`), rather than silently committing a loss.
- Resetting the TS Company Exit carry fails **both** shared parity cases at the Founder-output
  comparison (two failed, 6,951 passed, 85 skips).

All source mutations were removed; `git diff --exit-code` confirms no residual production
change. Final cold local checks pass: real-Postgres `make verify-server-core` (vet and every
non-harness Go package at `-count=1`), focused Go corpus checks, 6,953 client tests plus 85
browser skips, typecheck with zero errors/warnings, 22 Cosmetic boundary negatives and the
no-payment gate. Formula/API generation leaves no diff. The auxiliary Pitch-specific recipe
reported cached, but the same Pitch tests ran cold in the package population; no warm recipe
is counted as a new independent proof. Hosted CI was not run, and RP-131's historical
kernel-history failure remains separate.

**READY FOR CLAUDE DESIGNATED REVIEW:** the test-only correction starts at `e7bc1c8b^` and
includes this supplemental test/fixture/docs/tracking commit; the next checkpoint records the
exact final hash. Codex's corrective checks are a first filter, not self-approval. RP-176's
ruling-author §6/AC4 reconciliation, remaining Cosmetic reviews, copy and G10 are still open.

## 2026-10-05 — RP-181 exact Exit-carry review span

**Review needed by:** Claude (designated cross-party reviewer). **Recorded by:** Codex.
The test-only RP-181 corrective span is `e7bc1c8b^..66cfa3db`: recorded gap/predeclaration
and retained missing-row negative, two shared Exit pairs, per-path Go assertions, actual
Postgres acquisition/equip/Exit/retry/history journey, docs and synchronized tracking.
Neither production Exit source file has a residual mutation. This checkpoint only names the
range; it is not a designated verdict and does not close C3/C4, RP-176 or Cosmetics.

## 2026-10-05 — C8/AC7 receipt and consumer evidence predeclaration

**Review by / recorded by:** Codex, bounded to Claude `c20d23f1^..c20d23f1` on AC7.
RP-182 records the missing explicit production-receipt assertion and the disconnected Company
arms (nil contributions). The final bonus-producer comparison is useful and independently
mutation-proven; it does not turn identical nil-input Company runs into a consumer witness.

Predeclare a test-only accepted-AC7 correction: a receipt-only probe changes the cosmetic arm's
Company receipt without changing its state or Founder, and must initially make the missing
receipt assertion visible. Retain 200 seeds × 288 five-minute steps and all three applied
cosmetic kinds. Compare every Company receipt/outcome/state, not only final Company bytes.
Use the pinned fixture economy and the real FrozenFounderContributions→ResolveFrozenContributions
channel for Company math. Bonuses stay frozen within each run; after the cosmetic population,
a fresh next-run consumer checks the resulting bonuses through actual Company math. Never
refreeze mid-run. Retain independent negatives that change only a receipt and only a consumed
bonus, requiring named divergence rather than any error. Keep the non-cosmetics Founder and
producer comparisons. Cold Go/client and server-core checks, no production/catalog/RFC changes,
Claude designated review before closure.

## 2026-10-05 — C8/AC7 targeted verdict and receipt/consumer supplement

**Review by:** Codex. **Recorded by:** Codex.
**Reviewed range:** Claude `c20d23f1^..c20d23f1`, bounded to AC7's policy witness.
**Verdict:** CHANGES REQUIRED (RP-182), not a verdict on the full C8 docs/range.
The committed receipt-only probe in `52bd6963` changed one arm's receipt while leaving both
states unchanged; the original property returned nil and its retained negative failed with
`receipt-only leak was not caught by its named oracle: <nil>` (Make exit 2).
The existing producer mutation proof remains valid, but the Company arms supplied nil bonuses.

The accepted-AC7 test-only correction keeps all 200 seeds × 288 five-minute steps and all
three applied cosmetic kinds. It compares every outcome and receipt, complete encoded Company
state after each step, and non-cosmetics Founder bytes after each cosmetic transition. The
previous policy clone copied only selected Company fields; the new clone uses the real save
codec so purchase totals and the other current fields survive each step. Company state uses
the same pinned fixture economy as the Founder and real new-run/foundation initialization.

Both arms consume FrozenFounderContributions→ResolveFrozenContributions, including a seeded
valid non-unit Fiscal generator level. Bonuses remain immutable during the simulated run.
Afterward each Founder feeds a fresh next run; a real generator purchase and lazy accrual
consume the resulting bonus rows. That separate consumer is not a mid-run refreeze. The
existing multiplier-input negative is retained, alongside receipt-only and consumed-bonus
negatives requiring their specific divergence text.

**Invalid first fixtures disclosed:** the first strengthening attempt supplied a legacy prior
Company to an active-foundation settlement and assumed Reputation level 1 alone was a non-unit
bonus. It failed on `foundation mechanics cannot disappear between epochs` and the explicit
non-unit population guard. These were test-fixture errors, not production defects or successful
negatives. The prior Company now has the actual active floor, meters, achievement state and
active-play initialization; the Founder has a declared valid Fiscal level before adoption.
No accepted bounds, guards, runtime code or content were loosened to repair the fixture.

**Executed evidence (all temporary mutations restored):**

- Cold focused negative population passes all three tests; the full positive passes at
  200 seeds, 24 hours (29.13 seconds locally).
- Disabling only the per-step receipt comparison makes the receipt-only negative fail with
  `<nil>` (Make exit 2). Unchanged Company bytes cannot substitute for that receipt oracle.
- Setting only the next-run consumer contributions to nil makes the consumed-bonus negative
  fail with `<nil>` (Make exit 2). A producer comparison alone cannot satisfy the new consumer.
- `docker compose -f compose.save-test.yml -f compose.save-test-arm64.yml run --rm -w
  /workspace test make verify-server-core` passes: vet, cold non-harness Go packages against
  real Postgres (production 55.758 seconds), API/formula no-drift and boundaries. The auxiliary
  Pitch single-gate recipe reports cached; the complete Pitch package also ran cold above.
- `make typecheck test-client verify-cosmetic-boundary verify-no-payment` passes:
  0 type/Svelte diagnostics, 6,953 client tests, 85 intentional browser skips in the unit lane,
  22 package negatives and 6 no-payment negatives. This is not a browser result.

No live cosmetic bonus bug was established. This controlled pure-engine population still
uses fixture eligibility tiers and pins Founder command time to isolate Fiscal sweeps; it is
not an integrated browser/service claim. No production bytes, catalog, copy, RFC body, kernel
identity or CI workflow changed. The full client history gate remains red at RP-131's pushed
`50a3a514`, per the last full-target run at `ad789790`; green selected checks are not green CI.
Claude must designated-review this Codex correction. No C8/Cosmetics closure, archival or push.

## 2026-10-05 — AC7 exact correction range and C6/AC12 predeclaration

**Review needed by:** Claude. **Recorded by:** Codex.
The test-only RP-182 span is `52bd6963^..7c16d890`; this entry is a checkpoint, not a
designated review. It includes the missing-receipt negative, per-step/full-state assertions,
correct frozen producer/consumer inputs, independent firing negatives, docs and tracking.

Next Codex designated review is bounded to Claude C6 `1bba27ba^..1bba27ba` on AC12.
RP-183 records that only Buy references the bound visible disclosure text; Equip and Unequip
do not. Predeclare an all-controls browser assertion in the existing unowned/pending/owned/
equipped render population. Execute it against unchanged source first. Then change only the
two missing component references to the existing canonical describedBy helper. Run the cold
Linux three-engine component population and current type/unit/boundary checks; remove one
link again to demonstrate failure, then restore. No new copy, behavior rule, payments, pricing,
catalog or schema. Record actual keyboard/browser evidence, not an AT participant study.
The component correction requires Claude's designated cross-party review before closure.

## 2026-10-05 — C6/AC12 targeted verdict and wearer-control link correction

**Review by:** Codex. **Recorded by:** Codex.
**Reviewed range:** Claude `1bba27ba^..1bba27ba`, bounded to AC12's control/disclosure binding.
**Verdict:** CHANGES REQUIRED (RP-183), not a complete C6 acceptance or AT-study verdict.
The retained all-controls assertion at `e5c63de5` fails against the unchanged component in
Chromium, Firefox and WebKit: `owned Put it on Mittens curtain references: expected undefined`
instead of the three bound curtain IDs. The visible paragraphs existed; the links did not.

The bounded accepted-§8/AC12 correction adds the existing describedBy helper to Equip and
Unequip. No copy or new disclosure is authored. The retained render population now checks
every button's exact references alongside the rendered paragraphs' existing visibility/text
checks, including owned-pending and equipped-pending states with disabled controls.

Executed evidence:

- Cold declared Linux browser composition, focused shelf file: initial 3 failures/9 passes,
  then 12/12 pass across all three engines after the correction.
- Remove only Unequip's link, leaving Equip correct: all three engines fail specifically at
  `equipped Take it off Mittens curtain references: expected undefined`. The probe is restored.
- Cold complete `make test-browser-ci` passes: 252 file populations, 21,099 tests, 3 intentional
  performance-case skips across the three-engine non-performance run (91.09 seconds), followed
  by the separate Chromium performance population (1 pass/20 filtered skips). These counts
  include shared vector/unit tests; they are not counts of integrated user journeys.
- `make typecheck test-client verify-client-boundary verify-cosmetic-boundary verify-no-payment`
  passes: 0 type/Svelte diagnostics, 6,953 unit tests (85 browser skips), shell/UI boundaries,
  22 cosmetic negatives and 6 no-payment negatives. The production Vite build also passes.

The first attempted focused container command asked the declared Playwright image to run
`make`; that image has no Make binary and exited 127 after installing its frozen dependencies.
This was an invalid launcher attempt, not a product failure or a successful negative. The
subsequent focused runs use the declared composition's Corepack/frozen-pnpm setup and its
actual Vitest command, selected by file. The full CI composition uses its unchanged command.
Existing orphan Postgres services are preserved, never removed as cleanup.

No callback, ownership/revision semantics, style, generated copy, catalog, kernel path or CI
workflow changed. Existing axe assertions pass but this range does not expand their severity
selection or claim manual screen-reader/AT evidence. Copy adoption and G10 live overlay remain
open. Codex's correction is a first filter awaiting Claude designated review; no archive/push.
Copy-history and kernel-history command results are recorded in the following checkpoint.

## 2026-10-05 — C6 exact correction span and current local CI checkpoint

**Review needed by:** Claude (designated cross-party reviewer). **Recorded by:** Codex.
RP-183's bounded correction is `e5c63de5^..e02f6560`: predeclaration/retained failing
all-controls assertion, two component links, pending-state tests, docs and tracking.
The preceding RP-182 correction remains `52bd6963^..7c16d890`. Neither is self-approved.

The remaining command results are now known: `make build-client copy-check` passes, including
657 copy keys, the existing 610 orphan warnings, append-only copy-history checks and deployment
content-manifest identity. `make verify-kernel-version` passes the checkout-contract fixtures
then fails with Make exit 2 on exactly pushed `50a3a5141d5c4c1cd25cbf3c9c39d9e3cf8e2444`
against parent `0cf9f7a6aba4038fadcdf35e5b94f56986164af7`, naming the six historical minigame
paths. RP-131 remains open; `rfc/kernel-history-guard-integrity.md` still says draft/not
implementation authority. No exception, fake kernel bump, history rewrite or CI bypass was
added. This local result is not a fresh hosted run; the complete CI cannot be called green.
The two temporary Go-test probes and Unequip link probe are all restored. No push or archive.

## 2026-10-05 — RP-184 live pet overlay reconciliation and integrated proof predeclaration

**Recorded by:** Codex. Current tree is clean at `6b1c5def` before this scope.
The live consumer already exists in Claude Garage `7a61e4b6^..7a61e4b6`: PetCareSurface
mounts CosmeticOverlay from wearers[].worn, and GameUIApp supplies that snapshot arm. Garage's
host fixture tests a literal equipped pet. Cosmetics docs nevertheless say mounting is still
missing. RP-184 records this contradictory capability claim, not a demonstrated live mount bug.

Under accepted Cosmetic §4/§7.4, predeclare a test-only extension of the existing actual
gameserver/Postgres/built-client/WebSocket AC14 driver. Keep its explicit test-only epoch,
Company cash prerequisite seed and Buy/reload/N5 checks unchanged. Afterward adopt a real pet
through its visible DOM control, equip through the shelf, navigate to the actual pet surface,
check the CSS-only annoyed/no-text overlay, reload and check persisted wearing, exercise the
browser's actual reduced-motion preference on a fresh mount, then unequip through the shelf
and verify the pet overlay disappears while ownership remains. Verify exactly one intent per
DOM action, actual Founder revisions and applied receipts/server projections. Never seed
pet identity, ownership or equip state; no direct gameplay-intent API calls.

Cold honest focused composed run, then sever only the PetCareSurface overlay consumer in a
temporary probe and require the composed path to fail on the missing live overlay. Restore
before checking the full composed CI target, type/unit/build and browser lanes. Reconcile docs,
plan, live execution queue and the 1.0 board against actual results. No product, catalog, copy,
RFC body or kernel change is planned. This is an integrated controlled-fixture witness, not a
production mint or approval of the unresolved GS4×PA7 contract. Claude designated review of
Codex's evidence supplement remains mandatory; no archival, status promotion or push.

## 2026-10-05 — G10 honest path and additional motion-proof predeclaration

The first cold real-server extended driver passed in 11.849 seconds, with 67 observed requests.
It adopted a server-created identity through the DOM, equipped the persisted item, mounted the
live annoyed/no-text overlay, reloaded wearing, selected the actual browser reduced-motion
preference on a fresh mount and unequipped with ownership preserved. Disabling only the
PetCareSurface overlay mount then failed specifically with `cosmetic G10 live pet overlay
missing after equip` (exit 1); that source mutation is restored.

The motion assertion now asks for the ruled computed static outcome, not a particular internal
animate prop: the CSS preference and the host prop are independent defenses, so either alone
may lawfully keep the pose static. Predeclare a combined temporary probe that defeats both
those defenses and require the actual reduced-motion population to fail on its computed
animation, then restore both files. A surviving single-defense probe would not be a defect.
The initial default-sandbox Docker start failed before running; the same named disposable
service was started under narrow approved local-test escalation. No operator database is used.

## 2026-10-05 — RP-184 persisted pet workflow evidence and truthful capability reconciliation

**Implemented / recorded by:** Codex. **Review needed by:** Claude (designated cross-party).
This is a test-only supplement and docs correction, not a production mount fix. The mount
already landed in Garage `7a61e4b6`; the stale absence claim is removed from Cosmetics docs.
G10 release acceptance remains open, distinct from implementation presence and fixture evidence.

The retained existing composed driver now keeps its Buy/reload checks and continues through
DOM adoption, equip, the actual pet nav/sprite, worn reload, the browser's real reduced-motion
preference on a fresh mount, unequip and unworn reload. Each new gameplay action must emit
exactly one exact-key POST with the current Founder revision and applied incremented receipt.
Pet identity comes from the server adoption receipt, never a seed. Equip/unequip events and
persisted projections must match it; the item stays owned after removing the overlay. Every
gameplay write originates in the DOM. Read-only authenticated snapshots inspect results.

Executed discrimination (all production mutations restored):

- Honest first actual composed path passes (11.849 seconds, 67 requests).
- Disable only the PetCareSurface consumer: fails `cosmetic G10 live pet overlay missing after
  equip` (exit 1), despite successful persisted acquisition/adoption/equip.
- Defeat both independent reduced-motion defenses (host animate prop and overlay CSS media
  rule): fails after actual browser preference/reload with `reducedMotion:true` and non-static
  computed `animation:"svelte-5vo3xf-flick"` (exit 1). The assertion checks the ruled static
  outcome, not a specific internal implementation flag.
- Restored complete `make test-game-ui-composed GAME_UI_COMPOSE_FILES='-f
  compose.game-ui-test.yml -f compose.game-ui-arm64.yml'` passes both drivers. The original driver
  observes 19 actual manual clicks/claimed active.production, a Fiscal→Pitch terminal with 5
  UI commands, both terminal states, continuation and WebSocket recovery. The extended Cosmetic
  driver then passes in 6.408 seconds with 66 requests. Both drivers build and run the actual
  client/server, and the declared temporary Postgres database is reset by the existing setup.
- Cold full `make test-browser-ci` passes 252 file populations/21,099 tests (3 intentional
  performance-case skips), then the separate Chromium performance case (1 pass/20 filtered
  skips). These include shared unit/vector cases, not 21,099 distinct integrated journeys.
- Typecheck, 6,953 client unit tests (85 browser skips), shell/UI, cosmetic-package and
  no-payment boundaries pass. `node --check` and `git diff --check` pass.

The synthetic epoch and Company cash prerequisite seed remain explicit controlled test setup;
no content was minted and no pet/ownership/equip state was injected. The pet/adoption labels
still use the generated candidate copy, not owner-adopted final prose. This does not reconcile
GS4×PA7, approve broader Garage/Cosmetics ranges, close G10 for release or provide a manual AT
study. The current CI still has RP-131's historical kernel-history failure, not a fresh hosted
green result. No runtime byte, kernel identity, catalog, copy, RFC body or workflow changed.

## 2026-10-05 — RP-184 exact designated-review handoff

**Review needed by:** Claude. **Recorded by:** Codex.
The test-only driver/docs/tracking span is `b717b0db^..e4e6e567`: predeclaration and ledger,
retained actual DOM/persisted pet workflow, both restored severing proofs, truthful capability
docs, plan, execution queue and 1.0 checkpoint. This is not a designated verdict. The original
Garage `7a61e4b6` range is not approved by this Codex first-filter supplement; wider original
Garage/Cosmetics review and all release gates still apply. No product mutation remains.

## 2026-10-05 — C6/AC15 actual browser-preference predeclaration (RP-185)

**Review by / recorded by:** Codex. **Original range:** Claude `1bba27ba^..1bba27ba`,
bounded to AC15's component evidence. The existing static arm mounts `animate:false`, so it
proves the component prop rather than a user's actual reduced-motion preference in three
engines. RP-184's real-server Chromium preference proof does not cover Firefox or WebKit.

Predeclare a test-only correction under accepted §7.4/AC15. First sever only the overlay CSS
preference rule temporarily and run the original focused overlay case; a surviving case proves
this evidence exclusion, not a live motion defect. Restore it before editing the retained test.
Add a test-provider command that changes the actual Playwright page preference and always
restore no-preference in test cleanup. With `animate:true`, require a real preference match,
static computed pose under reduce, and resumed computed animation under no-preference.
Retain the explicit `animate:false` defense, overlay layer/reaction/no-text checks, and mount
inside a bounded CSS fixture sprite. The component population is Chromium, Firefox and WebKit,
using the declared cold Linux browser service; it is not a manual assistive-technology study.

Then sever only the preference CSS rule: the new actual reduce arm must fail on animation.
Independently add a temporary nonblank caption: the no-text arm must fail. Restore exact
production bytes after each, then run full browser CI, typecheck, client tests and boundaries.
No runtime fix, catalog/copy change, kernel bump, CI topology change or RFC-body reconciliation
is authorized/planned. Claude must cross-party review the resulting Codex test supplement;
this review does not approve the whole original C6 range or close G10/release acceptance.
