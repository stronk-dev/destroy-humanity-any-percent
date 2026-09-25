# RFC: Cosmetic Shop v1 — $0.00, Horse Armor

- **Status:** draft — not implementation authority
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-09-25
- **Design refs:** `design/07` Phase 1 ("the cosmetic shop ($0.00, Horse Armor)"); `design/04` §4
  (Horse Armor, FREE with the struck-through price, the pet visibly annoyed, Remastered; "Nothing is
  ever purchasable with real money"), §6 (CSS-only sprites, reduced motion); `design/00` satire
  thesis 2 and 4, pillar 6, anti-goals; `design/08` §1 rules 1–3 and 5, §3 (gaming-history suite),
  §7 (honesty appendix); `design/11` §1b (Horse Armor is the Tier-1 shop's opening joke), §2
  (unfold-as-onboarding), §5 (the curtain rule applies to tooltips), §7 (every parody surface
  carries its curtain; the pet panel is a sincerity island; accessibility baseline); `design/01`
  Tier 1 era anchor ("2006 first DLC (free horse armor)"); `design/02` §2c law 3 and §6b (the Gaia
  law); `design/05` §2 trading row (transferability); `design/research/gaming-enshittification.md`
  §1.1 (2006-04-03 row), §6.3, §6.5.
- **Depends on:** archived Save Layer & Migrations, Pet Care Foundation (Founder v18 `pets` map),
  Founder Attendance, Soul Foundation, Copy Pipeline, UI Foundation and Game-UI Screens (all
  implemented); `rfc/minigame-api-and-surface.md` (accepted; owns Founder v21);
  `rfc/accessibility-player-workflows.md` (draft; this RFC consumes its acceptance floor);
  the proposed `pet-adoption-v1` (not yet drafted) **for the equip half only** — see §4.4 and OD-7.
- **Parent / amends:** replaces the static Horse Armor card described in `docs/game-ui.md`
  §Shipped surfaces ("the free Horse Armor shelf") and the `cosmetic_stubs` presentation rows
  (`purchasable: false`, `stateful: false`) from archived T0–T1 Playable Content and Game-UI
  Screens. Neither archived RFC is edited.
- **Supersedes / superseded by:** —
- **Planning:** `planning/cosmetic-shop-v1/` (once implementing)
- **Release manifest row:** `rfc/v0.1-garage-release-manifest.md` G10 (proposed owner of the
  UNOWNED "cosmetic ownership" item; batch C; EC-4 and EC-5).
- **Evidence coordinate:** repository HEAD `c2d9bbc`. This is a static trace of sources, docs and
  catalogs; no product code was run for this draft. `server/save/state.go` declares
  `LatestFounderVersion = 21` and `LatestCompanyVersion = 18` at this coordinate.

## Summary

Today Horse Armor is one static Desk card: a title, a `$0.00` description, and a curtain-pull
disclosure. It has no state and no actions (`client/src/game-ui/GameUIApp.svelte:353`;
`client/src/game-ui/presentation.ts:8` types it `purchasable: false; stateful: false`).

This RFC turns that card into a working **$0.00 cosmetic shop**:

- a declarative cosmetic catalog;
- Founder-scoped ownership that survives every Exit;
- a server-authoritative `acquire_cosmetic` intent. It moves no ledger value, and the receipt
  reads `$0.00`;
- `equip_cosmetic` / `unequip_cosmetic` intents that put an owned cosmetic on a pet;
- a Desk shelf, a parody order receipt, and a pet overlay with the "visibly annoyed" reaction;
- a **curtain-pull contract**. Every dark pattern the shop parodies is enumerated, bound to a
  disclosure copy key, and rendered persistently. A loader or render test fails if any curtain is
  missing.

The catalog ships exactly one item, Horse Armor, because it is the only item the design names for
this shop at a known tier. Cosmetics have zero mechanical effect. The no-real-money law is enforced
**by construction** and proven by negative tests: the grammar has no price field, the intents have
no payment field, there is no payment dependency, and a browser network witness shows no
off-origin call.

## Motivation

`design/07` Phase 1 names the cosmetic shop as a v0.1 deliverable. `design/11` §1b names Horse
Armor as the Tier-1 shop's opening joke. The platform-alignment ledger (`P-004.03`) records the
current path as `partial_integration`: "Disclosure is visible but no acquisition mechanic exists."
The v0.1 release manifest marks cosmetic ownership UNOWNED and proposes this file as its owner.

The joke only lands if the shop behaves like a real one. You press Buy, get a receipt, own the
item, and the pet wears it. Every step then states that it cost nothing and that the pattern is a
parody (`design/00` thesis 2: "reproduce every dark pattern faithfully, price it at zero, set
disclosure to maximum"). The static card cannot be bought, owned or worn, so it only tells the
joke.

### In scope

1. The `cosmetics` simulation catalog (schema v1) and its pinning in the constants bundle.
2. The `cosmetics` presentation rows and the curtain-binding grammar (presentation schema v4).
3. Founder save v22: ownership and per-pet equip state, the migration and the activation rule.
4. Three Founder-scoped intents with receipts and three closed event kinds.
5. The mechanical-isolation contract: cosmetics touch no ledger, multiplier, Clout, Soul, Trust
   or production path.
6. The `game_ui_snapshot.v4` cosmetics projection.
7. The Desk shelf, the parody receipt, and the pet overlay component with its reaction pose.
8. The no-real-money-by-construction contract and its negative tests.

### Out of scope (each needs its own RFC or content change)

- Every other `design/04` §4 item: hats, Unusual effects, the Knife, Default Skin, Surprise
  Mechanic Crates, gacha banners, vaulted countdowns, the AI Slop line, the Battle Pass, Compliance
  Points, and the collection book.
- The exchange shop (`design/09` §5) and Influence-priced world-layer cosmetics (`design/02` §4).
- Gifting and trading, and the transferability flag (`design/05` §2).
- The house and decor (`design/04` §3).
- Slots other than the pet: the founder avatar, buildings, and the cursor.
- The `IDDQD` cheat cosmetic (`design/01` Tier 0). See DESIGN-GAP 2.
- `Horse Armor (Remastered)` unless OD-1 admits it. See DESIGN-GAP 1.
- Achievements that consume the events below (for example research §6.2 "Horse Armor — first
  cosmetic equipped"). See OD-13.
- Pet identity, adoption, and the live pet panel. These belong to `pet-adoption-v1` and
  `garage-player-surfaces`.
- The honesty appendix page (`design/08` §7). This RFC only exports the machine-readable curtain
  list the appendix will consume.
- Export, deletion and retention UI (the D-008/D-009/D-015 successor). This RFC only declares
  the data class (§11).

## Specification

### §1 Invariants (normative; every later section is subordinate)

- **I1 — No real money, in any direction (law 1).** No catalog field, intent field, dependency,
  endpoint, header, or network call can represent, request, collect or transmit a real-money
  amount, payment instrument, or store product. `$0.00` exists only as the presentation constant
  `constant.price_zero`, which Game-UI presentation v3 already owns. The simulation never sees a
  price. §10 lists the enforcement.
- **I2 — Zero mechanical effect.** Owning, equipping or unequipping a cosmetic never reads or
  writes the economy ledger, the multiplier stack, production, gates, Routes, Clout (the Gaia law,
  `design/02` §6b), achievement score, Soul, pet stats, pet Trust, or the pet behavior FSM. Every
  run and replay produces a byte-identical Company state whether or not cosmetic intents
  interleave.
- **I3 — Server-authoritative (law 2).** Clients send intents only. Ownership and equip state
  change only through §4 on the authoritative Founder stream. The client never shows an item as
  owned before an applied receipt arrives (the Client Shell D2 "snap discrete with receipts" rule).
- **I4 — Curtain always pulled (law 10).** Every parodied pattern on the shop surface has a bound
  disclosure. The disclosure is rendered persistently, before and after interaction, in every item
  state. The build fails if one is missing (§8).
- **I5 — The pet stays sincere (law 10, `design/04` tone).** The pet's reaction to the armor is
  rendered as genuine animal behavior. The pet never delivers a joke, and the reaction carries no
  copy in v1.
- **I6 — Hardcaps.** The owned set is bounded by the pinned catalog size. The equipped map is
  bounded by the number of pets, with one cosmetic per pet in v1. No other counter exists.
- **I7 — No shop telemetry.** The shop emits the three gameplay events in §4 and nothing else. It
  records no impression, view, hover, dwell, funnel or conversion metric.

### §2 Simulation catalog — `cosmetics` artifact, schema v1

The artifact is balance data (law 4). It is a new named artifact in the constants bundle,
alongside `pets` and `minigames`, and is pinned by a content epoch.

```json
{
  "schema_version": 1,
  "items": [
    {
      "cosmetic_id": "horse_armor",
      "slot": "pet",
      "unlock": { "kind": "active_company_tier_at_least", "tier": 1 }
    }
  ]
}
```

Validation (Go and TypeScript loaders enforce the same rules):

- Top-level keys are exactly `schema_version` and `items`. Item keys are exactly `cosmetic_id`,
  `slot` and `unlock`. Unlock keys are exactly `kind` and `tier`. Any other key rejects the whole
  artifact. In particular, `price`, `cost`, `currency`, `sku`, `product_id`, `store`, `amount` or
  any payment-shaped key is rejected **as an unknown field**, not by a special case. The negative
  tests in AC1 name them to prove the rejection holds.
- `cosmetic_id` matches the save layer's mechanical-ID pattern
  (`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`), is unique, and rows are sorted by raw bytes.
- `slot` is a closed grammar enum. In v1 its only member is `pet`. A new slot is a schema change in
  a later RFC, not a data row.
- `unlock.kind` is a closed enum. In v1 its only member is `active_company_tier_at_least`, and
  `tier` is an exact integer in `[0, 8]`.
- **Cosmetic IDs are permanent.** The epoch-transition validator rejects a next bundle whose
  `cosmetics` artifact omits any `cosmetic_id` present in the current pinned artifact, so no owned
  item can become unknown. A row's `slot` may never change. Its `unlock` may be retuned.
- The v1 catalog has **exactly one row**, `horse_armor`, with `tier: 1` (`design/11` §1b; `design/01`
  Tier-1 era anchor; research §4 "Indie Studio" stage). The tier value is config (a starting value,
  per the working conventions), not a code constant.

### §3 Founder state — Founder save v22

Cosmetic ownership is Founder-scoped: the company dies, the cat does not (`design/04` preamble).
v22 is v21 plus exactly one required key.

```json
"cosmetics": {
  "owned": ["horse_armor"],
  "equipped": { "<pet_id>": "horse_armor" }
}
```

- `owned` is an array of catalog IDs, strictly ascending by raw bytes. It has no duplicates, every
  ID is present in the pinned `cosmetics` artifact, and its length is at most the catalog row
  count.
- `equipped` maps a pet ID to one cosmetic ID. Every key must be a key of the same state's `pets`
  map. Every value must be in `owned`, and its catalog `slot` must be `pet`. An empty map is
  canonical `{}`, never `null`. Empty `owned` is canonical `[]`, never `null`. The existing
  pet-queue rule applies to both.
- No timestamps, counts, prices or receipt numbers are persisted. The order number shown on the
  receipt is derived (§4.2).
- A v22 state without `cosmetics`, or a v21 or earlier state with it, is invalid. A v22 state is
  valid only under a bundle that pins the `cosmetics` artifact.

**Version axis.** This is a Founder-only mechanic on the scalar Founder chain
(`docs/save-layer.md`). v22 requires the v21 predecessor's pinned artifacts plus `cosmetics`. The
Company axis is untouched. The `cosmetics` artifact registers the declared Founder floor 22
before any code path may write v22 ("an unregistered artifact cannot raise a version by
convention").

If another draft claims Founder v22 first, this RFC takes the next free Founder version at
acceptance. The version number is a coordinate, not a mechanic.

### §4 Intents, receipts, events

All three intents are Founder-career commands. They run through the existing
`ApplyFounderLogged` boundary and immutable `founder_log`, like `care_action` and
`buy_route_hint`, with the standard lowercase UUIDv7 `intent_id` idempotency and a positive
safe-integer Founder `expected_revision`. Decoding is strict: an unknown or extra field is terminal
`invalid`. The request hash follows the existing canonical rule.

#### §4.1 Common evaluation prefix

For every cosmetic intent, in this order:

1. Envelope decoding and revision check (existing semantics).
2. The Founder state is below v22, or the bundle does not pin `cosmetics` → terminal
   `not_eligible` / `inactive`.
3. `cosmetic_id` (where present) is not in the pinned catalog → terminal `unknown_id`.

#### §4.2 `acquire_cosmetic`

Request: `{intent_id, kind:"acquire_cosmetic", expected_revision, cosmetic_id}`.

4. The ID is already in `owned` → terminal `not_eligible` / `owned` (the same shape as
   `buy_upgrade`).
5. Evaluate `unlock` against the **frozen active-Company context**. The server selects the active
   Company sibling and freezes its authoritative tier into the Founder replay envelope, as
   `care_action` freezes attendance. The client never supplies a tier. If the tier is below
   `unlock.tier` → terminal `not_eligible` / `locked`.
6. Apply: insert the ID into `owned` in sorted order. Emit exactly one
   `cosmetic_acquired.v1 {cosmetic_id, order_number}`, where `order_number` is the length of
   `owned` after insertion (exact safe integer, at least 1). IDs are permanent and never removed,
   so this number is stable and unique per Founder.
7. No ledger read or write occurs. The applied receipt carries the standard fields, the complete
   `cosmetics` object, and the event. It carries **no** amount, price, currency or payment field.

Ownership is not consumed by Exit. A Founder who owns Horse Armor keeps it in a Tier-0 run. The
shelf visibility rule (§7.1) follows from that.

#### §4.3 `equip_cosmetic`

Request: `{intent_id, kind:"equip_cosmetic", expected_revision, cosmetic_id, pet_id}`.

4. `pet_id` is not a key of `pets` → terminal `unknown_id`.
5. `cosmetic_id` is not in `owned` → terminal `not_eligible` / `not_owned`.
6. `equipped[pet_id] == cosmetic_id` → terminal `not_eligible` / `already_equipped`.
7. Apply: set `equipped[pet_id] = cosmetic_id`. Emit exactly one
   `cosmetic_equipped.v1 {cosmetic_id, pet_id, replaced_cosmetic_id}`, where
   `replaced_cosmetic_id` is the previous value or `null`.

One owned cosmetic may be worn by several pets at once (OD-15). Ownership is a fact, not a stock,
and the shop does not perform scarcity mechanically.

#### §4.4 `unequip_cosmetic`

Request: `{intent_id, kind:"unequip_cosmetic", expected_revision, pet_id}`. There is no
`cosmetic_id`, so prefix step 3 is skipped.

4. `pet_id` is not a key of `pets` → terminal `unknown_id`.
5. `pet_id` is not a key of `equipped` → terminal `not_eligible` / `nothing_equipped`.
6. Apply: delete the key. Emit exactly one `cosmetic_unequipped.v1 {cosmetic_id, pet_id}`.

**Pet dependency.** At the evidence coordinate, production has no pet identity ("no starter pet or
species is fabricated", `docs/pet-care.md`). Equip and unequip are therefore fully specified and
tested against fixture pets, but in production they always reject `unknown_id` until
`pet-adoption-v1` ships. Acquisition does not depend on pets (OD-7).

#### §4.5 Cross-cutting rules

- **Soul, cooldowns, exclusivity.** None apply. Cosmetic intents are not care actions and carry no
  `soul_gate`. They are not minigame activities and are never blocked as `exclusive_activity`.
- **Events.** A migration expands the closed event-kind constraint with exactly
  `cosmetic_acquired.v1`, `cosmetic_equipped.v1` and `cosmetic_unequipped.v1`. Payloads are strict
  and contain only the fields above. Applied state, revision, events and receipt commit atomically
  through the existing store transaction, and the receipt enters the transport outbox as usual.
  Terminal rejections store only the receipt.
- **Idempotency.** Identical replay returns the original receipt. A corrected payload under the
  same `intent_id` returns `idempotency_conflict` (existing semantics).
- **Exit.** Exit carries `cosmetics` byte-identically into the next Founder revision. No Exit arm
  reads, clears or rewrites it. Pet retirement (`design/04` §5) does not exist yet. When it ships,
  retired pets remain keys of `pets` and their equip entries remain valid.

### §5 Mechanical isolation

- A new server package (mechanical name `cosmetic`) holds the catalog loader, the three pure
  transitions, and the validators. A package gate like the Achievements gate forbids it from
  importing ledger spending, the multiplier stack, production, faction, fiscal, Clout/achievement
  score, Soul mutation, or pet care-transition symbols. It may read the `pets` key set and the
  frozen tier only.
- The mirrored TypeScript module has the same boundary under a `verify-*-boundary` script, and
  that script ships with a failing fixture.
- The balance harness does not model cosmetics. Cosmetic intents are absent from every harness
  policy, and AC7 proves they could not change any harness outcome.

### §6 Save migration and activation

- **Codec.** v21 → v22 adds `"cosmetics": {"owned": [], "equipped": {}}`. The migration is pure
  and never reads the wall clock. v1–v21 remain readable through the existing chain. No v22 →
  v21 down-conversion exists, and neither axis may regress.
- **Activation.** This follows the Soul v20 precedent (`docs/soul.md`). An existing Founder moves
  to v22 **only at a new-run boundary** (Exit into the next run) under a bundle that pins
  `cosmetics`. A Founder created under such a bundle starts at v22. An ordinary mid-run Founder
  write never raises the version. Exit derives the reachable version tuple from the pinned bundle,
  and no client or deployment setting chooses it.
- **Corpus.** `testdata/save-migrations.json` gains named cases: `founder-v21-to-v22-empty`,
  `founder-v22-owned-equipped`, `founder-v22-equipped-unknown-pet` (reject),
  `founder-v22-equipped-not-owned` (reject), and `founder-v21-with-cosmetics` (reject). The case
  count baseline ratchets in the same reviewed change.
- **Pre-activation UI.** While the active Founder is below v22, the Desk renders the existing static
  card unchanged (fail closed). It never shows a Buy button that cannot work.

### §7 Snapshot and UI surface contract

#### §7.1 `game_ui_snapshot.v4`

v4 is v3 plus a required `cosmetics` object. v1–v3 snapshots remain decodable under the schema
version they were minted with, per the existing encrypted-receipt rule.

```json
"cosmetics": {
  "active": true,
  "items": [
    { "cosmetic_id": "horse_armor", "owned": false, "acquirable": true,
      "lock": null, "worn_by": [] }
  ],
  "wearers": [ { "pet_id": "<pet_id>", "worn": null } ]
}
```

- `active` is true exactly when the Founder is at v22 or later under a pinned `cosmetics` artifact.
  When `active` is false, `items` and `wearers` are empty.
- `items` follows catalog order. `owned` comes from Founder state. `acquirable` is advisory and is
  true iff the item is not owned and the frozen active tier satisfies `unlock`. `lock` is `null`
  or `{"kind":"active_company_tier_at_least","tier":N}` for a visible reason. `worn_by` is the
  sorted pet IDs wearing it.
- `wearers` follows `pets` key order. `worn` is a cosmetic ID or `null`.
- The Go projector and TypeScript decoder share one fixture. The decoder rejects internal
  contradictions, for example `owned: true` with `acquirable: true`, or a `worn` value absent from
  `items[].worn_by`.

#### §7.2 Presentation schema v4 (client catalog)

v4 replaces `cosmetic_stubs` with two sections.

```json
"cosmetics": [
  {
    "cosmetic_id": "horse_armor",
    "title_key": "cosmetic.horse_armor_free.title",
    "description_key": "cosmetic.horse_armor_free.description",
    "anchor_key": "cosmetic.horse_armor_free.anchor",
    "render_key": "horse_armor",
    "wearer_reaction": "annoyed",
    "curtains": [
      { "pattern": "paid_cosmetic_dlc", "key": "cosmetic.horse_armor_free.disclosure" },
      { "pattern": "reference_price_anchor", "key": "cosmetic.horse_armor_free.anchor_disclosure" }
    ]
  }
],
"cosmetic_shop": {
  "heading_key": "shop.cosmetics.heading",
  "curtains": [ { "pattern": "checkout_flow", "key": "shop.cosmetics.checkout_disclosure" } ]
}
```

- `wearer_reaction` is a closed enum: `annoyed | none`.
- `render_key` names a CSS overlay layer in the pet sprite system.
- `anchor_key` is nullable.
- The three existing Horse Armor copy keys and their owner-ratified text are reused **unchanged**
  (evidence discipline 6). The `{price}` parameter still binds `constant.price_zero`.
- The build-time presentation check requires the `cosmetics` presentation IDs to equal the pinned
  catalog IDs exactly. At runtime, a catalog ID without a presentation row is withheld, not
  rendered mechanically. This mirrors the existing unknown-slot rule.

#### §7.3 The Desk shelf

The shelf replaces the static card in the same Desk position.

- **Visibility.** The shelf is shown when `cosmetics.active` and (some item is `acquirable`, or
  some item has a `lock` whose tier is at or below the active tier, or some item is `owned`). In
  practice it appears at Tier 1 and never disappears after the first acquisition. At Tier 0 with
  nothing owned it is absent: the shop *appears* (`design/11` §2 unfold-as-onboarding; OD-4). The
  T0 order form is untouched.
- **Item card, unowned state.** The card shows, in order: the title; the description with
  `{price}`; the anchor line (§8, row 2); a Buy button (`shop.cosmetics.buy`, param `price`); and
  the item's curtain small print. There is no confirm dialog, cart, quantity, timer, countdown,
  "limited" badge, urgency toast or upsell. None of those is a pattern this RFC parodies. Adding one
  would be a new parody that needs its own curtain row and a later RFC.
- **Pending.** While the intent is pending, the button is disabled and keeps its label. The owned
  state never renders before the applied receipt (I3). A terminal rejection renders its reason
  through copy. `locked` renders `shop.cosmetics.locked` with the tier.
- **Receipt.** On the applied receipt, a polite `role="status"` region inside the card renders
  `shop.receipt.line` (params `order_number`, `item`, `price`) and `shop.receipt.payment_method`.
  It is not a modal: `design/11` §1 allows at most one return modal, and a receipt has not earned
  one. The receipt line is session-local. On reload the card shows the owned state without
  re-announcing it.
- **Owned state.** The card shows `shop.cosmetics.owned`, the curtain small print (still present),
  and per-wearer controls: `shop.cosmetics.equip` / `shop.cosmetics.unequip` (param `pet`) for
  each entry in `wearers`. If `wearers` is empty, the card shows `shop.cosmetics.no_wearer` in place
  of controls, a visible reason rather than a silently missing button.
- **Shop-level curtain.** `shop.cosmetics.checkout_disclosure` renders once under the shelf
  heading in every state.
- **Accessibility.** The shelf consumes the `accessibility-player-workflows` floor. That means
  keyboard operability, 320 px reflow, and no focus loss to `<body>` when the Buy button is
  replaced; focus moves to the owned-state heading. Each parody control (Buy, the anchor line) has
  `aria-describedby` pointing at its curtain small print. The strikethrough is decorative (`<s>`
  inside an element whose accessible text is the full anchor copy), so screen-reader users never
  depend on a visual strike to understand the anchor. The existing axe WCAG 2.2 AA gate covers the
  shelf in all three engines.
- **Era.** The shelf uses the active era's chrome tokens (`era_2000` at T1). Copy rows may carry
  `era_variants`.

#### §7.4 Where cosmetics render on the pet

- This RFC owns a presentation-only `CosmeticOverlay` component. Its input is
  `{render_key, wearer_reaction}`. It renders a CSS layer inside the existing CSS-only sprite
  system (`design/04` §6: nested divs, custom properties, data-attribute poses, zero image assets).
- `wearer_reaction: "annoyed"` sets a data-attribute pose read as ordinary cat annoyance: ears
  back, a tail flick, a slow blink away. It adds no speech, text or caption (I5). Under
  `prefers-reduced-motion` the pose is static.
- The reaction is **presentation-only**. It never enters pet stats, Trust, mood or the behavior FSM
  (I2; OD-6).
- Mounting the overlay into the live pet panel is the pet surface slice's job
  (`garage-player-surfaces` / `pet-adoption-v1`). That slice must pass `wearers[].worn` through
  this component. This RFC proves the component against a fixture sprite. The integrated witness
  is the manifest's G10 exit criterion, and it cannot close until a pet surface exists.
- In v1, cosmetics render **nowhere else**: not in the feed, house, avatar, cursor, buildings,
  leaderboards or other players' views. There is no social surface, so ownership is private
  gameplay state.

### §8 Curtain-pull disclosure contract (law 10)

Each parodied pattern is a closed presentation enum member with exactly one curtain copy key per
bound element. The rules:

- Every catalog item binds `paid_cosmetic_dlc`.
- An item with a non-null `anchor_key` binds `reference_price_anchor`.
- The shop binds `checkout_flow`.
- `re_release` is bindable only by a row that declares a re-release (reserved for Remastered;
  OD-1).
- An unbound, duplicated or unknown pattern rejects the presentation catalog at build time and at
  load time.

| # | Pattern (enum) | Where it appears | The real-world pattern it reproduces | Curtain key | Curtain must state |
|---|---|---|---|---|---|
| 1 | `paid_cosmetic_dlc` | every item card | selling a purely cosmetic item for real money (research §1.1, 2006-04-03 row) | `cosmetic.horse_armor_free.disclosure` (**existing, owner-ratified**; provenance `gaming.horse_armor_2006`) | that the item once cost real money elsewhere, costs nothing here, and does nothing |
| 2 | `reference_price_anchor` | the struck-through "was" line (`design/04` §4 "FREE (~~200 Microsoft Points~~)") | reference-price / fake-discount anchoring | `cosmetic.horse_armor_free.anchor_disclosure` (**new, pending owner prose**) | that the struck price is a historical reference, not a discount, and nothing is being saved because nothing was ever charged |
| 3 | `checkout_flow` | the Buy button, the order receipt | the storefront purchase flow and receipt | `shop.cosmetics.checkout_disclosure` (**new, pending**) | that no payment is taken, no payment details are ever requested, and the receipt records a $0.00 transaction |
| 4 | `re_release` | Remastered only (OD-1) | re-selling the same cosmetic in a remaster (research §2 "Horse armor" row) | `cosmetic.horse_armor_remastered.disclosure` (pending; only if OD-1) | the "we did it again" re-release, same zero price |

Carrier rules:

- **Persistent small print is the normative carrier.** It is visible without hover or focus, not
  truncated or collapsed, meets AA contrast, and is present in the unowned, pending, owned and
  equipped states. It is visible **before** the first interaction, so the player learns what the
  pattern is before clicking.
- A hover tooltip may duplicate the text but is never the only carrier (OD-9; Deviations).
- The curtain list (the pattern enum, every bound key, and the element it describes) is exported
  as part of the generated presentation artifact, as the input `design/08` §7's future honesty
  appendix enumerates.
- Curtain copy follows `design/08` §1: deadpan corporate, no winking, and no brand name or figure
  unless OD-8 admits one with a provenance row.

### §9 Copy keys

Player-facing text is owner-authored (evidence discipline 6). This RFC fixes **keys, params, tone
tags and intent only**. Implementers must not author shipped prose. Candidate text returns for
explicit owner adoption through the existing copy-candidate ratification path.

| Key | Params | Tone | Status / intent |
|---|---|---|---|
| `cosmetic.horse_armor_free.title` | — | diegetic | existing, ratified, unchanged |
| `cosmetic.horse_armor_free.description` | `price:string` | diegetic | existing, ratified, unchanged |
| `cosmetic.horse_armor_free.disclosure` | — | diegetic | existing, ratified, unchanged (curtain 1) |
| `cosmetic.horse_armor_free.anchor` | — | diegetic | **pending** — the "was" line, full accessible text (OD-8) |
| `cosmetic.horse_armor_free.anchor_disclosure` | — | diegetic | **pending** — curtain 2 |
| `shop.cosmetics.heading` | — | diegetic | **pending** — shelf heading, T1 era voice |
| `shop.cosmetics.checkout_disclosure` | — | corporate | **pending** — curtain 3 |
| `shop.cosmetics.buy` | `price:string` | diegetic | **pending** — Buy label, binds `constant.price_zero` |
| `shop.cosmetics.owned` | — | diegetic | **pending** |
| `shop.cosmetics.locked` | `tier:integer` | diegetic | **pending** — visible lock reason |
| `shop.cosmetics.equip` / `shop.cosmetics.unequip` | `pet:string` | diegetic | **pending** — `pet` is the adopted pet's display name (pet-adoption-v1 owns naming) |
| `shop.cosmetics.no_wearer` | — | diegetic | **pending** — shown when no pet exists |
| `shop.receipt.line` | `order_number:integer`, `item:string`, `price:string` | corporate | **pending** — parody receipt |
| `shop.receipt.payment_method` | — | corporate | **pending** — states that no payment method was used |

- The pet reaction carries **no** copy in v1 (I5). The catalog has no `sincere` tone tag, and
  adding one is Copy Pipeline scope (DESIGN-GAP 3).
- **Copy-linter rule (new, fixture-backed).** No string under `shop.*` or `cosmetic.*` may contain
  a currency amount, meaning a currency symbol followed by a digit, an ISO currency code with a
  number, or "points" with a number. The only exceptions are the `{price}` placeholder and rows
  carrying a verified provenance claim. The detector-configuration change and its failing fixture
  ship with this RFC.

### §10 No real money, by construction

| # | Mechanism | Where |
|---|---|---|
| N1 | The catalog grammar has no price, currency or product field. Unknown fields reject. | §2 loaders (Go + TS) |
| N2 | The intent grammar has no payment field. Strict decoding makes any extra field terminal `invalid`. | §4 decoders |
| N3 | The cosmetic package cannot import ledger or spending symbols. | §5 package gates |
| N4 | **Dependency allowlist.** The client's runtime `dependencies` must equal a checked-in allowlist. A denylist scan of `client/pnpm-lock.yaml` and `server/go.mod`/`go.sum` rejects known payment/IAP/ads SDK package-name patterns (for example `stripe`, `paypal`, `braintree`, `adyen`, `square`, `paddle`, `chargebee`, `recurly`, `revenuecat`, `in-app-purchase`, `google-pay`, `apple-pay`, `adsense`, `admob`). The denylist is a secondary net; the allowlist is the gate. | new `verify-no-payment` script, wired into `verify-client` / `verify-server` |
| N5 | **Network witness.** During the full shop flow in the browser, every HTTP and WebSocket request must hit the same-origin gameserver API allowlist. `window.PaymentRequest`, `navigator.credentials.*` and `window.open` are trapped, and any call fails the test. | composed browser lane |
| N6 | `Permissions-Policy: payment=()` and a CSP whose `connect-src` is `'self'` plus the gameserver WebSocket origin (OD-11) | Caddy config, coordinated with Deployment Foundation |
| N7 | Copy-linter currency rule | §9 |
| N8 | No shop telemetry: the cosmetic package may not import metrics or observability packages | §5 gate |

The existing `make verify-client-boundary` rule against raw network calls in Game-UI components
already applies to the shelf. The shelf's intents go only through `client/src/game-ui/runtime.ts`.

### §11 Data class (for EC-5)

`cosmetics` is a new player-data class inside the Founder save. It is also stored as rows in the
immutable `events` table. It holds no free text, no identifier beyond pet IDs, and no timestamps
beyond event metadata. This RFC registers it with the D-008 export / D-009 deletion / D-015
retention successor. The successor owns the UI and the joined deletion witness. This RFC's own ACs
prove the data is reachable from the Founder save and the events table, which is what that
successor enumerates.

## Deviations from design

1. **Scope is one item.** `design/04` §4 lists the whole parody suite, but `design/07` Phase 1
   names only "the cosmetic shop ($0.00, Horse Armor)". Remastered is excluded by default because
   its tier is unspecified (DESIGN-GAP 1; OD-1).
2. **The curtain carrier is persistent small print, with the tooltip optional.** Law 10 and
   `design/08` §1 rule 2 say "tooltip". `design/11` §7 says "tooltip or small print". Hover-only
   tooltips fail keyboard and touch users and the WCAG 2.2 content-on-hover rules, which the
   accessibility floor enforces. Small print meets the law's intent (the pattern states what it
   is, at maximum disclosure) for every player. OD-9.
3. **The anchor copy may omit the brand and figure.** `design/04` §4 shows
   "~~200 Microsoft Points~~". The ratified Horse Armor copy already chose "no brand, no figure"
   (`planning/archive/t0-t1-content/screen-copy-ruling-v1.md` §6). The default follows that ruling.
   OD-8.
4. **The pet reaction is visual-only and has no mechanical effect.** `design/04` §4 says only
   "visibly annoyed". This RFC does not make wearing armor a care or mood input. OD-6.
5. **The shelf is absent at Tier 0.** Today the static card renders at every tier. The shelf now
   follows `design/11` §2/§1b (the shop *appears* at T1). OD-4.

## Acceptance criteria

Every gate ships a demonstrated failing case (evidence discipline 1) and is run with `-count=1` or
an uncached equivalent. "Both runtimes" means Go and TypeScript against one shared Go-authored
fixture.

- **AC1 Catalog grammar (both runtimes).** The v1 fixture loads. Each of these rejects the whole
  artifact: an added `price`, `currency`, `sku` or `product_id` key; an unknown `slot`; an
  unknown `unlock.kind`; `tier: 9`; a duplicate ID; unsorted rows; a non-mechanical ID. *Failing
  case:* a loader that ignores unknown keys passes the `price` fixture, and this test must go red
  against it.
- **AC2 Permanent IDs.** Epoch-transition validation rejects a next bundle whose `cosmetics`
  artifact drops `horse_armor` or changes its `slot`, and accepts a retuned `unlock.tier`.
  *Failing case:* the drop fixture.
- **AC3 Founder v22 codec (both runtimes).** Round-trips a canonical v22 state. It rejects: unsorted
  or duplicate `owned`; an unknown ID; `equipped` keyed by a non-pet; `equipped` pointing to an
  unowned item; `null` for either collection; v22 without `cosmetics`; v21 with `cosmetics`; v22
  under a bundle without the artifact. *Failing case:* each named reject fixture.
- **AC4 Migration and activation.** The corpus cases in §6 pass, and the case-count baseline
  ratchets. A mid-run Founder write under a cosmetics-pinning bundle keeps the prior version. Exit
  into a new run under that bundle yields v22 with empty cosmetics. A new Founder under that
  bundle starts at v22. *Failing case:* a fixture that writes v22 mid-run is rejected.
- **AC5 `acquire_cosmetic` (both runtimes + Postgres integration).**
  - At T1 it applies, emits exactly one `cosmetic_acquired.v1 {cosmetic_id:"horse_armor", order_number:1}`,
    and leaves the ledger byte-identical.
  - At T0 it rejects `not_eligible/locked`. A second acquisition rejects `not_eligible/owned`.
    An unknown ID rejects `unknown_id`. Below v22 it rejects `not_eligible/inactive`. A request
    with an extra `price` field rejects `invalid`.
  - Identical replay returns the original receipt. A corrected payload under the same ID returns
    `idempotency_conflict`.
  - The Go-authored receipt and event vector matches byte-for-byte in TypeScript.
  - *Failing case:* removing the unlock check turns the T0 case green-applied, and the test must
    fail.
- **AC6 Equip/unequip (both runtimes, fixture pets).** These apply: equip, replace (emits
  `replaced_cosmetic_id`), equipping two pets with one item, and unequip. These reject:
  `unknown_id` pet, `not_owned`, `already_equipped`, and `nothing_equipped`. *Failing case:* a
  fixture that equips an unowned item.
- **AC7 Mechanical isolation.** A property test runs 200 seeded intent policies (the existing
  24-simulated-hour harness shape) twice: once plain, and once with cosmetic intents interleaved
  at seeded points. Company state, ledger, production receipts and non-`cosmetics` Founder fields
  must be byte-identical. The package gates (§5) reject a fixture importing the ledger and a
  fixture importing a metrics package. *Failing case:* a test-only transition variant that adds a
  multiplier contribution on equip must make the property test fail. The recorded demonstration
  lives in the planning log.
- **AC8 Exit carry.** `accept_exit_offer` and `wind_down` both carry `cosmetics` byte-identically.
  *Failing case:* an Exit arm fixture that resets `cosmetics` is caught.
- **AC9 Event persistence (Postgres, `docker compose -f compose.save-test.yml run --rm test`).**
  The migration admits exactly the three new kinds. Applied intents commit state, revision, event
  and receipt atomically. The database rejects an unregistered event kind and a payload carrying
  an extra field. *Failing case:* inserting `cosmetic_purchased.v1` fails.
- **AC10 Snapshot v4 (both runtimes).** The shared projector/decoder fixture passes, and v3/v2/v1
  receipts still decode. The decoder rejects `owned:true, acquirable:true` and a `worn` value
  absent from `worn_by`. *Failing case:* those two fixtures.
- **AC11 Shelf behavior (component + browser, three engines).**
  - The shelf is absent at T0 with nothing owned and present at T1. It stays present in a later
    Tier-0 run once owned.
  - Buy submits exactly one `acquire_cosmetic` through `runtime.ts`. The owned state does not
    render until the applied receipt. The receipt status line renders `order_number`.
  - Keyboard-only acquisition works without focus loss. The axe WCAG 2.2 AA gate passes. The
    `no_wearer` line renders when `wearers` is empty.
  - *Failing case:* an optimistic-ownership fixture renders "owned" before the receipt, and the
    test must fail.
- **AC12 Curtain completeness.**
  - Presentation build and load reject an item missing `paid_cosmetic_dlc`, an anchor without
    `reference_price_anchor`, a shop without `checkout_flow`, a duplicate pattern, and an unknown
    pattern.
  - A render test asserts each bound curtain is visible, not hover-gated, in the unowned, pending,
    owned and equipped states, and is referenced by `aria-describedby` on its control.
  - The exported curtain list equals the bound set.
  - *Failing case:* removing the small print from the rendered card must fail the render test,
    not just the loader.
- **AC13 No real money (negatives).**
  - N1/N2 are covered by AC1/AC5.
  - N4: a fixture lockfile containing `stripe` and a fixture `package.json` adding any runtime
    dependency both fail `verify-no-payment`.
  - N5: the composed witness fails when a fixture build issues
    `fetch("https://example.invalid/checkout")` or constructs `PaymentRequest`.
  - N6 (if OD-11 is adopted): a deployment config test fails when the header is absent.
  - N7: a fixture copy row `shop.cosmetics.buy` containing `$2.50` fails `copy-check`.
  - Each negative's failing run is recorded in the planning log.
- **AC14 Composed witness (real gameserver + Postgres, Chromium).** Ordinary server-side setup
  reaches T1. The visible Buy button acquires. A reload shows the owned state from the
  authoritative snapshot. The whole flow runs under the N5 network trap. *Failing case:* severing
  the producer (the projector omits `cosmetics`) must fail the witness. A green UI over a severed
  backend is not accepted.
- **AC15 Pet overlay (component, three engines).** `CosmeticOverlay` renders the `horse_armor`
  layer and the `annoyed` pose on a fixture sprite, renders a static pose under reduced motion, and
  renders no text node. *Failing case:* a fixture that adds a caption to the overlay fails the
  no-text assertion. Integration into the live pet panel is **not** claimed here; it is the
  manifest G10 exit row.
- **AC16 Docs and archival.** `docs/cosmetics.md` (new) plus updates to `docs/game-ui.md`,
  `docs/save-layer.md` and `docs/production-engine.md` (intent list) land in the implementing
  change. A designated cross-party review verdict covers the full implementation span before
  archival.

## Owner decisions (numbered; recommended default in bold)

- **OD-1 Catalog scope.** **Horse Armor only.** Alternative: also ship `Horse Armor (Remastered)`.
  That requires an owner-picked unlock tier, a `re_release` curtain, and a provenance row for the
  2025 re-sale claim (research §2 "Horse armor" row).
- **OD-2 Ownership scope.** **Founder-scoped and permanent across Exits** (`design/04` "the company
  dies; the cat does not"). Alternative: per-run Company ownership, which contradicts the
  collection model.
- **OD-3 Unlock predicate.** **Active Company tier ≥ 1 at acquisition.** Alternatives: the Founder
  has ever reached T1 (career), or always available.
- **OD-4 Tier-0 visibility.** **The shelf is absent at T0 until something is owned.** Alternative:
  keep a non-interactive teaser card at T0.
- **OD-5 Acquisition model.** **An explicit `$0.00` Buy intent with a parody receipt.** Alternative:
  auto-grant at T1, which loses the checkout parody and curtain 3.
- **OD-6 Pet reaction.** **Presentation-only pose, no stat/Trust/mood/FSM effect, no copy.**
  Alternative: a declared behavior-FSM candidate row, which would need a Pet Care follow-up RFC.
- **OD-7 Sequencing against pet adoption.** **Implement the whole RFC fixture-first now.**
  Acquisition is live on activation. Equip rejects `unknown_id` in production until
  `pet-adoption-v1` supplies pets, and the shelf shows `no_wearer`. Alternative: block the RFC on
  pet adoption.
- **OD-8 Anchor copy.** **Generic "was" anchor with no brand and no figure**, following the
  existing ruling. Alternative: "200 Microsoft Points" with a provenance claim and a trade-name
  rights review (F05).
- **OD-9 Curtain carrier.** **Persistent small print plus `aria-describedby` is normative; a
  tooltip is an optional duplicate.** Alternative: tooltip-only, which fails the accessibility
  floor.
- **OD-10 Artifact identity.** **`cosmetics` joins the constants bundle** (replay integrity;
  catalog changes ride content epochs). Alternative: a separate identity outside
  `constants_hash`, so that adding a cosmetic never segments boards. That needs a new identity
  mechanism like `copy_hash`.
- **OD-11 Payment-blocking headers.** **This RFC adds `Permissions-Policy: payment=()` and the
  `connect-src` CSP to Caddy, coordinated with Deployment Foundation.** Alternative: hand N6 to a
  Deployment follow-up and keep only N1–N5, N7 and N8.
- **OD-12 Receipt.** **An inline polite status with order number = owned count, and no modal.**
  Alternative: a persisted receipt history view (new state; out of v1).
- **OD-13 Achievements.** **Not in this RFC.** The three events are available to the
  `clout-v1-and-pr-interns` achievement catalog (for example research §6.2 "Horse Armor — first
  cosmetic equipped", proof `provenance`). No cosmetic path emits Clout.
- **OD-14 Mechanical ID.** **`horse_armor`, with the existing `cosmetic.horse_armor_free.*` keys
  reused unchanged.** Alternative: rename the keys, which forces owner re-ratification of the copy.
- **OD-15 One item, many wearers.** **Allowed.** Alternative: one wearer per owned item, which
  would perform scarcity mechanically.
- **OD-16 Founder version coordinate.** **v22, or the next free Founder version at acceptance.**

## Open questions and DESIGN-GAPs

- **DESIGN-GAP 1 — Remastered tier.** `design/04` §4 says "at a later tier" without naming one,
  and "better textures" has no meaning in a CSS-only, zero-image sprite system (`design/04` §6).
  The owner must supply both before Remastered can be a row. The grammar already supports it with
  no schema change (the `re_release` pattern is reserved).
- **DESIGN-GAP 2 — `IDDQD` cosmetic.** `design/01` Tier 0 says typing `IDDQD` "gives a cosmetic"
  but names neither the item nor its slot. This RFC's ownership model can host it later through
  a separate grant intent. Out of scope here.
- **DESIGN-GAP 3 — Sincere copy tone.** Pet barks are "sincere (never satirical)"
  (`design/11` §5), but the copy catalog has only `achievement`, `corporate` and `diegetic` tones.
  v1 avoids the gap by giving the reaction no copy. The Copy Pipeline needs a follow-up before any
  pet bark ships.
- **DESIGN-GAP 4 — Slot vocabulary.** `design/04` §4 names hats "for the pet, the founder avatar,
  buildings, and the cursor". v1 defines only `pet`. Other slots, and whether several cosmetics
  can layer on one pet (armor plus hat), belong to the hats RFC.
- **Deferred — Gifting.** `design/05` §2 keeps hand-authored cosmetics giftable. v1 has no
  transfer path. A gifting RFC must add the transferability flag before any crate-derived item
  exists.

## Changelog

- 2026-09-25: created as a draft (not implementation authority) for release-manifest row G10.
