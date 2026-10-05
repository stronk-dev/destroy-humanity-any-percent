# Cosmetics — the $0.00 shop

Cosmetic Shop v1 (`rfc/cosmetic-shop-v1.md`) turns the static Horse Armor card into a working
$0.00 shop. Cosmetics have **zero mechanical effect** and are **never purchasable with real money**.
Both properties are enforced by construction and by gates, not by convention. The shop is
fixture-first: no epoch pins `cosmetics` yet (`balance/testdata/cosmetics/fixture-v1.json`).

## Catalog (`cosmetics` artifact, schema v1)

`{schema_version, items: [{cosmetic_id, slot: "pet", unlock: {kind: "active_company_tier_at_least", tier}}]}`.

- **Loading:** `server/cosmetic` and `client/src/cosmetic` load it with exact keys at every level,
  so `price`, `currency`, `sku`, `product_id` and any other key reject as unknown fields. They also
  enforce mechanical, unique, byte-sorted IDs, the closed `slot` and unlock enums, a tier in
  `[0,8]`, duplicate-key rejection, and canonical integer JSON tokens for `schema_version`
  and `unlock.tier` (so `1.0` and `1e0` do not become `1` only on the client). The shared corpus is
  `testdata/cosmetic/catalog-fixtures-v1.json`.
- **Bundle:** the artifact joins the constants bundle and requires `pet_species` on the scalar
  Founder chain.
- **Permanence:** IDs are permanent. `settleAndActivateFoundations` rejects a next epoch that drops
  a pinned ID, changes its slot, or drops the artifact. Only `unlock` may be retuned.

## Founder state (Founder v24)

`cosmetics: {owned: [sorted ids], equipped: {pet_id: id}}`. Empty collections are canonical `[]`
and `{}`, never `null`.

- **Validation:** every equipped key must be a `pets` key, and every equipped value must be owned
  and use the pet slot. One owned item may be worn by several pets.
- **Activation:** v24 activates only at a new-run boundary (or for a new Founder) under a
  `cosmetics`-pinning bundle, with nothing owned. It never activates mid-run.
- **Exit:** Exit carries `cosmetics` byte-identically. The Company replay carry holds it from
  replay-inputs v11 when the Founder floor is at least 24.
  The shared replay corpus covers owned/equipped carry through both `wind_down` and
  `accept_exit_offer`. A real-Postgres witness acquires and equips through the service before
  each Exit, checks reload/retry, and verifies both Company and Founder histories.
- **Version numbering:** the RFC named "v22"; numbering is landing-order, so this landed as v24.

## Intents and events

`acquire_cosmetic {cosmetic_id}`, `equip_cosmetic {cosmetic_id, pet_id}` and
`unequip_cosmetic {pet_id}` are Founder-scope intents through `ApplyFounderLogged`. They are never
blocked as `exclusive_activity`, including while Soul recovery is active. The top-level service
preflight exempts only these cosmetic kinds; ordinary gameplay intents remain exclusive.

- **Validation order:**
  - the prefix: `not_eligible/inactive`, then `unknown_id/cosmetic_id`;
  - acquire: `owned`, then `locked`, evaluated against the active Company tier the server freezes
    into the replay inputs (the client never sends a tier);
  - equip: `unknown_id/pet_id`, `not_owned`, `already_equipped`;
  - unequip: `unknown_id/pet_id`, `nothing_equipped`.
- **Events:** `cosmetic_acquired.v1 {cosmetic_id, order_number}`, where `order_number` is the owned
  count after insertion; `cosmetic_equipped.v1 {cosmetic_id, pet_id, replaced_cosmetic_id|null}`;
  `cosmetic_unequipped.v1 {cosmetic_id, pet_id}`. Payloads are strict. Migration `00080` admits the
  three kinds; append-only migration `00084` requires exactly those keys for each cosmetic kind
  at the Postgres boundary, leaving other event kinds unchanged. The Go decoder separately checks
  the event payload values before persistence.
- **Receipts:** they carry the complete `cosmetics` object and the event, and no amount, price,
  currency or payment field.
- **Commit guard:** only the three cosmetic intents may change `cosmetics`. Any other Founder
  transition that touches it fails the transaction.
- **Isolation:** `TestCosmeticsAreMechanicallyIsolated` checks that cosmetic intents interleaved
  into 200 seeded 24-hour policies leave every Company decision outcome/receipt, complete
  Company state and every non-cosmetics Founder byte identical. Both arms consume the pinned
  Founder contributions, including a non-unit fixture bonus, frozen once for that run.
  A separate next-run purchase/accrual checks the resulting bonus consumer; it does not refreeze
  bonuses mid-run. Retained receipt-only, multiplier-input and consumed-bonus negatives require
  their named divergence. This is a controlled pure-engine policy, not a browser or persisted
  service workflow; the cosmetic arm's eligibility tier and Founder Fiscal clock are fixtures.

## Snapshot and UI

- **Snapshot arm:** the optional arm `features.cosmetics` is
  `{active, items: [{cosmetic_id, owned, acquirable, lock, worn_by}], wearers: [{pet_id, worn}]}`,
  plus the fact `feature.cosmetics`. The client decoder rejects internal contradictions,
  including duplicate or out-of-order pet IDs in each `worn_by` list and lock tiers outside
  the catalog's 0–8 domain. The shared Go-projected fixture includes two simultaneous wearers
  so those reader checks are exercised against a real producer shape.
- **Desk shelf (`client/src/game-ui/cosmetics/CosmeticShelf.svelte`):**
  - it replaces the static card when the arm is active and an item is acquirable or owned, so it is
    absent at Tier 0 until something is owned;
  - it shows Buy, the lock reason, the owned state, per-wearer equip and unequip, and the no-wearer
    line;
  - the parody receipt is an inline `role=status` line, never a modal.
- **Fail-closed card:** below v24 the static card renders unchanged.
- **Overlay:** `CosmeticOverlay.svelte` is the CSS-only layer and the "annoyed" pose. It has no
  text, is presentation-only, and is static under reduced motion. Mounting it into the live pet
  panel is still open.
- **Curtain contract:** `client/src/game-ui/cosmetics/presentation.{json,ts}`:
  - every item binds `paid_cosmetic_dlc`;
  - an anchor binds `reference_price_anchor`;
  - the shop binds `checkout_flow`;
  - `re_release` is reserved;
  - curtains render as persistent small print in every state and are the controls'
    `aria-describedby`, including Buy, Equip and Unequip while pending or applied;
  - the loader rejects a missing, duplicated or unknown curtain, and exports `curtainList` for the
    honesty appendix.

## No real money (N1–N8)

- **N1/N2:** the catalog and intent grammars have no price or payment field.
- **N3/N8:** `make verify-cosmetic-boundary`. `server/cosmetic` is standard-library only, and
  `client/src/cosmetic` imports only within its own package. The gate parses import/re-export
  syntax, including dynamic imports and import-equals, rejects computed specifiers and normalized
  path escapes, and recursively checks nested script files. Literal/comment text is not code.
- **N4:** `make verify-no-payment`. Runtime dependencies must equal
  `client/tools/runtime-dependency-allowlist.json`, and a denylist scans the lockfile and Go modules
  for payment, IAP and ads SDKs.
- **N5:** the browser network trap (`client/test/network-trap.ts`) wraps the shop flow.
- **N6:** Caddy sends `Permissions-Policy: payment=()` and
  `Content-Security-Policy: connect-src 'self' wss://{host}`, which `ValidateCaddyfile` requires.
- **N7:** the copy-check currency rule: no currency amount in `shop.*` or `cosmetic.*` copy.

## Still open

- **Copy:** all shop copy is candidate text for owner adoption
  (`copy/catalog/cosmetics-candidate.json`).
- **Composed witness (AC14):** `make test-game-ui-composed` uses a test-only epoch that pins
  `cosmetics` and runs the built browser client through real Postgres, gameserver, and WebSocket
  transport. It observes T0 locked → visible T1 Buy → applied intent → server-owned state after
  reload under the request/payment trap; severing the server Cosmetics projector fails the run.
  This fixture is not a production content mint, and its Codex range awaits cross-party review.
- **Integrated pet-panel overlay:** release-manifest row G10.
