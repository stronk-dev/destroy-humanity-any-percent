# RFC: Pet Adoption v1 — Species Catalog, Adoption Intent, and the Cattery Port Surface

- **Status:** draft — not implementation authority
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-09-25
- **Design refs:** `design/04` §1 (the pet: starter server-room cat at Tier 0, named later variants,
  personality, Trust as the Soul proxy), §2 (the hardcapped ceiling and the tone guard; conflict
  avoidance only), §5 (no-death, fur palette), §6 (the CSS-only sprite port, reduced motion),
  §Neopets adoptions (no-death canon); `design/03` §Unlock stagger (the T2 variants row) and §10b
  (pet battles; conflict avoidance only); `design/07` Phase 1 ("pet adoption + care (cattery
  port)"); `design/11` §1 (at most one modal), §5 (pet barks: "short, sincere (never
  satirical)"), §6 (`{pet_name}` slot), §7 (sincerity islands, accessibility baseline);
  `design/08` §1 (voice rules, used only to exclude the pet layer from them); `design/13` §2 (the
  T0 garage scene: "you + the cat"); `design/00` pillar 4 (the Bogost rule: "The pet genuinely
  loves you")
- **Depends on:** Pet Care Foundation (archived; `docs/pet-care.md`); Founder-scoped transitions
  (`docs/founder-transitions.md`); Save Layer (`docs/save-layer.md`); Production Engine replay
  inputs (`docs/production-engine.md`); Copy Pipeline (`docs/copy-pipeline.md`); Game UI
  (`docs/game-ui.md`). Coordinates with `rfc/v0.1-garage-release-manifest.md` (G14, VQ-9, EC-4,
  EC-5), `rfc/accessibility-player-workflows.md` (D-018, F03), and the not-yet-present drafts
  `rfc/garage-player-surfaces.md` (care panel) and `rfc/cosmetic-shop-v1.md` (Horse Armor pet
  reaction).
- **Parent / amends:** follow-up to the archived `rfc/archive/pet-care-foundation.md`. It
  discharges that RFC's carried debt ("starter species/temperament and pet acquisition belong to
  successor content", `planning/archive/pet-care-foundation/plan.md`). It amends C11/C14's
  `pet_records` table wording. See Deviation DV-1 and OD-5.
- **Supersedes / superseded by:** —
- **Planning:** `planning/pet-adoption-v1/` (once accepted)
- **Evidence coordinate:** repository HEAD `34462f7`; balance epoch 8. This draft is a static
  trace of the tree. No product code was run for it. Every "exists"/"absent" claim cites its file.

## Summary

The care backend is shipped: Founder-scoped care state, the `care_action` intent, and the epoch-6
care-policy artifact. No Founder can own a pet yet. `docs/pet-care.md` says: "Pet
identity/species acquisition … remain unimplemented" and "no starter pet or species is
fabricated". Every production `care_action` therefore ends in `unknown_id`. This RFC specifies the
missing half of the v0.1 item "pet adoption + care":

1. a strict, declarative **`pet_species` catalog artifact**. Its v1 content is exactly one row,
   the server-room cat that `design/04` §1 names as every Founder's Tier-0 companion.
2. a replay-owned, immutable **pet identity** in Founder save **v22**.
3. one Founder-scoped **`adopt_pet` intent**, with exact wire, validation order, receipt, event
   and server-side draws that the client cannot grind.
4. its exact relationship to the existing care state: adoption seeds a care record, and care is
   otherwise untouched.
5. the **adoption surface contract**. It is the first player-visible piece of the cattery port.
   It is sincere, keyboard- and AT-operable, and honors reduced motion.

The later variants that design names (robot vacuum, rubber duck, keychain blob, the red-eyed
horse) are listed and deliberately **not minted**. Each one has an unresolved DESIGN-GAP. The
catalog shape already fits them.

## Motivation

**Why now.** The v0.1 manifest (`rfc/v0.1-garage-release-manifest.md` §1a G14) lists adoption as
**UNOWNED** and proposes this filename. Adoption is also an upstream producer in that manifest's
DAG: `cosmetic-shop-v1` needs pet identity for the Horse Armor reaction (`design/04` §4), and
`garage-player-surfaces` needs a pet before its care panel can show anything. Until a Founder can
own a pet, the shipped care transition cannot run in production.

**What exists (at `34462f7`).**

- Founder v18+ carries the replay-owned `pets` map of `CareState`, keyed by a UUID pet ID
  (`server/pet/state.go`; `server/save/state.go`, where `LatestFounderVersion = 21`). Epoch 8
  activates it empty.
- The `pets` artifact (`balance/pets/first-content.json`, schema v2) holds species-agnostic care
  policy only. That covers the stats, the five actions with `soul_gate`, Trust, mood and
  behavior.
- The care intent is `{intent_id,kind:"care_action",expected_revision,pet_id,action_id}`, and
  `pet_id` must be a UUIDv7 (`server/production/intents.go`).
- Replay-inputs v6 `founder_extensions` carries pet care state (`docs/production-engine.md`).
- The Temperament enum is the six combat values, in the order lazy, playful, curious, sassy, shy,
  chaotic (`server/combat/arithmetic.go`; `docs/combat.md`).

**What is absent.**

- There is no pet identity anywhere. The C11/C14-ruled `pet_records(pet_id, founder_id,
  species_id, temperament, created_at)` table was **never implemented**: no migration under
  `server/save/migrations/` creates it, and no Go or TS code references it.
- There is no species catalog and no acquisition path.
- There is no pet surface. `docs/game-ui.md` lists none.
- There is no typed pet sprite contract. Pet Care AC5's "sprite contract typed for UI" has no
  artifact in `client/src` or `docs/` (see GAP-8).

**Out of scope.** This RFC does not cover:

- the care panel and its Soul greying, which belong to `garage-player-surfaces`;
- guild-visible pet status (C7 projection successor);
- bonds, the house, breeding and elder retirement (`design/04` §3, §5);
- pet battles and any pet-to-combat binding (`design/04` §2, `design/03` §10b; see PA9);
- hats, cosmetics and the Horse Armor reaction (`cosmetic-shop-v1` consumes this RFC's identity);
- temperament-qualified FSM transitions and the care-history ledger
  (`research/believable-pet-personality.md` §9 II–III; C17 reserves them for a future ruling);
- the recognition/greeting beat;
- life stages;
- renaming and release. There is no release verb at all (see PA4.8).

## Specification

### PA1 — Authority boundary and invariants carried unchanged

1. **Sincerity law (binding).** `design/00` pillar 4, `design/04` §1 ("play it straight, never as
   a joke") and §2 tone guard, and `design/11` §5/§7. The pet is never the target of a joke, is
   never priced, and never carries a curtain disclosure. It is never used as a dark pattern: no
   urgency, no guilt copy, no streak loss. Parody chrome recedes in every pet surface. These are
   testable surface rules (PA8, AC12–AC13). They are not aspirations.
2. **No-death / no-loss (carried).** A pet never dies, leaves, is released, or is deleted by
   gameplay. Adoption only ever adds. There is no verb that removes an identity.
3. **Hardcap law (carried).** Species carry no stat, care-rate, Trust or combat modifier. Every
   pet shares the one pinned care policy and the combat identity ceiling (`design/04` §2). Species
   and temperament are identity and presentation only in v1.
4. **Server-authoritative (carried).** The client sends an intent. It never sends a pet ID,
   temperament, palette, timestamp or clock context.
5. **Founder boundary (carried).** Adoption is a Founder-only command through
   `ApplyFounderLogged`. It reads no live Company state beyond the frozen Founder-attendance sample
   that care already uses, and it spends no resource.
6. **No real money (law 1).** Adoption is free. The surface shows **no price at all**, not even
   `$0.00`. The pet is not merchandise, and the `$0.00` joke belongs to the shop.

### PA2 — The `pet_species` catalog artifact

A new named epoch artifact **`pet_species`** at `balance/pet-species/first-content.json`, with a
JSON schema at `balance/pet-species.schema.json`. It is separate from the existing `pets`
care-policy artifact (OD-6). The C17-pinned `pets` top level stays byte-stable, so historical care
replay is untouched.

**Exact shape (schema_version 1).** Every object is exact-key. Unknown, duplicate or missing keys
fail load in Go and TS.

```json
{
  "schema_version": 1,
  "max_pets_per_founder": 1,
  "species": [
    {
      "species_id": "pet_species.server_room_cat",
      "availability": "starter",
      "visual_family": "cat",
      "allowed_temperaments": ["lazy", "playful", "curious", "sassy", "shy", "chaotic"],
      "palette_ids": ["pet_palette.fur_00", "…"],
      "name_keys": ["pet.name.server_room_cat.00", "…"],
      "name_copy_key": "pet.species.server_room_cat.name",
      "description_copy_key": "pet.species.server_room_cat.description"
    }
  ]
}
```

**Field grammar (wire; enums and keys are normative):**

| Field | Rule |
|---|---|
| `schema_version` | Exactly `1`. |
| `max_pets_per_founder` | Safe integer ≥ 1. A visible hardcap (law 5), shown on the surface. v1 value `1` (OD-12). Balance data. |
| `species` | Non-empty array, sorted by `species_id` byte order, with unique `species_id`. **Exactly one** row has `availability:"starter"`. This is load-validated, so "every Founder gets a companion" cannot be broken by data. |
| `species_id` | Mechanical ID matching `^pet_species\.[a-z][a-z0-9_]*$`. The `pet_species.` namespace is mandatory (PA9.1). |
| `availability` | Closed enum. v1 has exactly one member: `starter`. Later members (a tier unlock, for example) need a successor ruling (GAP-1). An unknown member fails load. |
| `visual_family` | Closed enum. v1 has exactly one member: `cat`. It selects the sprite family (PA8.5). |
| `allowed_temperaments` | Non-empty and duplicate-free. Each member is one of the six combat Temperaments, and the list is in canonical combat order (lazy, playful, curious, sassy, shy, chaotic). Its order is the draw domain (PA4.4). |
| `palette_ids` | Non-empty and duplicate-free, matching `^pet_palette\.[a-z0-9_]+$`, and sorted. Each ID resolves to a presentation swatch row (PA8.5). IDs are wire; swatch colors are presentation data (OD-7). |
| `name_keys` | Non-empty and duplicate-free, sorted copy keys matching `^pet\.name\.[a-z0-9_]+\.[a-z0-9_]+$`. Every key must exist in the copy catalog with tone `companion` (PA8.6). The pool is owner-authored (OD-4). |
| `name_copy_key`, `description_copy_key` | Registered copy-bearing paths in `copy/references.v1.json`. They must resolve with tone `companion`. |

The artifact joins constants identity (`docs/save-layer.md` §State format). A bundle that pins
`pet_species` must also pin the `pets` artifact. Both runtimes reject `pet_species` without `pets`.

**v1 minted content: one row, `pet_species.server_room_cat`** (`design/04` §1: "default: the
**server-room cat** (industry-canon)"; `design/13` §2 T0: "you + the cat"). The artifact is
minted into a production epoch only after the owner adopts its name pool, species name and
description copy. Until then it exists only as a fixture and activates nothing (AC14).

**Named in design, NOT minted in v1 (OD-1).** These are recorded so the shape is proven to fit
them. None may be added without resolving its gap.

| Design name (`design/04` §1) | Design unlock | Why not minted | Rights note |
|---|---|---|---|
| Robot vacuum "with a personality" | T2 (`design/03` §Unlock stagger) | GAP-1: there is no Founder-owned tier fact, and a Founder command may not read Company tier. The `visual_family` and sprite family are unauthored. | — |
| Rubber duck ("debugging canon") | T2 | GAP-1, as above | — |
| "Tamagotchi-style keychain blob" | T2 | GAP-1, as above | "Tamagotchi" is a third-party mark. Mechanics only: the ID and copy must never use the mark (`research/believable-pet-personality.md` legal line). |
| "The Blucifer-red-eyed horse" | "conspiracy-tier unlock" (the conspiracy layer is Tier 5, `design/07`) | GAP-1, plus the tier is far past v0.1 | "Blucifer" is the popular nickname of a real public artwork. It needs a verify-before-shipping and rights pass (F05). The ID must be mechanical, not the nickname. |

This RFC mints no IDs for these rows. Choosing them is successor work.

### PA3 — Pet identity state (Founder v22)

Founder save **v22** adds exactly one required map, `pet_identities`, keyed by pet ID. It is
replay-owned and mutated only through `ApplyFounderLogged` (C14 discipline: no second mutable
authority). Mutable care stays in `pets`, unchanged.

```json
"pet_identities": {
  "0199….-7…-….-…": {
    "species_id": "pet_species.server_room_cat",
    "temperament": "curious",
    "palette_id": "pet_palette.fur_03",
    "name_key": "pet.name.server_room_cat.07",
    "adopted_at_ms": 1790000000000,
    "adopted_at_attended_ms": 5400000
  }
}
```

**Invariants.** They are enforced by both codecs and, where an invariant needs the catalog, by
catalog-aware validation under the pinned bundle.

1. **Biconditional key sets.** `keys(pet_identities) == keys(pets)` at v22. A care record without
   an identity, or an identity without a care record, is invalid.
2. **Exact keys** as shown. `adopted_at_ms` is the command's server timestamp: a UTC
   whole-millisecond safe integer. `adopted_at_attended_ms` is the Founder-attendance sample total
   at adoption: a safe integer that is ≤ the pet's `evaluated_through_attended_ms`.
3. **Catalog resolution.** `species_id` exists in the pinned `pet_species` artifact. `temperament`
   is in that row's `allowed_temperaments`. `palette_id` is in `palette_ids`. `name_key` is in
   `name_keys`.
4. **Cap.** `len(pet_identities) ≤ max_pets_per_founder`, taken from the pinned artifact.
5. **Immutability.** Every Founder transition other than `adopt_pet` leaves `pet_identities`
   byte-identical. That includes care, Fiscal, minigame, Soul, route hints and **Exit**. An applied
   `adopt_pet` produces exactly the prior map plus one new key. The transition layer checks this
   after the callback, so a violating feature arm fails the authoritative transaction
   (`docs/founder-transitions.md`: "a live/replay semantic split fails the authoritative
   transaction").
6. **Pet ID format.** Each pet ID is a lowercase canonical UUIDv7 (PA4.3). This is stricter than
   the v18 `petIDPattern`, which accepts versions 1–8, and it matches the care intent's
   UUIDv7-only `pet_id`.

### PA4 — The `adopt_pet` intent

**PA4.1 Wire.** Founder intent, exact keys:
`{intent_id, kind:"adopt_pet", expected_revision, species_id, name_key}`.
`expected_revision` is the Founder revision. The client supplies no pet ID, temperament, palette,
Company, run or clock coordinate. Extra keys yield `invalid_intent`.

**PA4.2 Validation order (normative; the first failure wins).**

| # | Check | Outcome |
|---|---|---|
| 1 | Exact-key decode; `species_id` and `name_key` are well-formed mechanical IDs or copy keys | `invalid_intent` + field detail |
| 2 | Founder at v22 with `pet_species` pinned in the Founder's resolved bundle | `not_eligible` + `adoption_inactive` |
| 3 | `species_id` exists in the pinned artifact | `unknown_id` |
| 4 | `name_key` ∈ that row's `name_keys` | `unknown_id` |
| 5 | Row `availability == "starter"` (the only v1 member) | `not_eligible` + `species_locked` (reachable only through a successor member; kept so the union is closed now) |
| 6 | `len(pet_identities) < max_pets_per_founder` | `not_eligible` + `adoption_cap_reached` |
| 7 | Founder-attendance resolution | Identical resolution and failure behavior to `care_action`. The server selects the sole active Company sibling (C21), and the client cannot choose it. |
| 8 | Apply (PA4.3–PA4.6) | `applied` |

Adoption rejection details form their **own closed union**:
`adoption_inactive | species_locked | adoption_cap_reached`. The C12a care-rejection set is not
extended. Terminal rejections append a `founder_log` row and consume no revision, as usual.

**Soul.** Adoption has **no Soul predicate** (OD-13). A companion is essential by construction.
The `soul_gate` classification belongs only to care-action rows.

**PA4.3 Server entropy and the pet ID (grinding-proof).** On first execution the callback draws a
16-byte `adoption_nonce` from the server CSPRNG and freezes it in the resolved input. Every other
derived value comes from the nonce **and never from `intent_id`**. The client chooses
`intent_id`, so an `intent_id`-derived temperament could be ground offline before sending (AC3's
failing case).

`d = SHA-256("pet.id.v1" ‖ founder_uuid_bytes[16] ‖ nonce[16])`. The pet ID bytes are:

- `b[0..5]` = the 48-bit big-endian command server time in unix ms;
- `b[6] = 0x70 | (d[0] & 0x0f)`;
- `b[7] = d[1]`;
- `b[8] = 0x80 | (d[2] & 0x3f)`;
- `b[9..15] = d[3..9]`.

The ID is rendered as a lowercase hyphenated UUID. If the ID collides with an existing key (it
cannot within the cap, but the check is cheap), the transition fails closed as an internal error.
It does not re-draw.

**PA4.4 Temperament draw.**
`t = allowed_temperaments[ uint64_be(SHA-256("pet.temperament.v1" ‖ nonce)[0..8]) mod len(allowed_temperaments) ]`.
The modulo bias is at most 6/2⁶⁴ and is documented, not corrected. The exact byte recipe is the
parity contract. The draw is uniform over the row's list (OD-3). The formula is published in
`docs/` (law-9 spirit) and is not shown on the sincere surface.

**PA4.5 Palette draw.** The recipe is the same with the label `"pet.palette.v1"` over
`palette_ids`.

**PA4.6 State effect (single transaction).**

1. Insert `pet_identities[pet_id] = {species_id, temperament, palette_id, name_key,
   adopted_at_ms: server_time_ms, adopted_at_attended_ms: A}`, where `A` is the frozen attendance
   sample total.
2. Insert `pets[pet_id]` as the **canonical initial care record** under the pinned `pets`
   artifact:
   - `stats_ppm[s] = stat_policy.stats[s].initial_ppm` for all four stats;
   - every `stat_decay_remainders_ppm` = 0;
   - `cooldown_until_attended_ms = {}`;
   - `trust_ppm = trust_policy.initial_ppm`;
   - `trust_decay_remainder_ppm = 0`;
   - `evaluated_through_attended_ms = A`;
   - `behavior_state = "idle"`;
   - `behavior_entered_at_attended_ms = A`;
   - `behavior_queue = []`;
   - `behavior_prng_cursor = 0`.

   Seeding the watermark at `A` means that attended time before adoption never decays the new pet
   (AC5).
3. Increment the Founder revision.

**PA4.7 Resolved input, receipt, event.**

- **Resolved arm, exact:** `{kind:"adopt_pet", attendance, adoption_nonce}`. `attendance` is the
  same frozen A1–A5 sample object the care arm records. `adoption_nonce` is 32 lowercase hex
  characters. The canonical payload owns `species_id` and `name_key`, and the pinned artifacts own
  policy. Replay **re-derives** the pet ID, temperament and palette from the frozen nonce and
  byte-compares them. It never trusts a stored derived value.
- **Receipt, exact:** `{intent_id, outcome, founder_revision, pet_id, species_id, temperament,
  palette_id, name_key, adopted_at_attended_ms, status_band, eligible_action_ids}`.
  `status_band` and `eligible_action_ids` are computed by the existing care projection on the new
  record, so the surface can render immediately without a raw-stat leak. Rejected receipts carry
  `{intent_id, outcome, category, detail}` per the existing Founder receipt grammar.
- **Event:** register **`pet_adopted.v1`**, exact payload `{pet_id, species_id, temperament,
  palette_id, name_key}`. It is Founder-scoped, uses the player outbox only, and is published to
  no feed, guild or world channel (D-017 posture). A new append-only SQL migration extends the
  closed event-kind constraint (precedent: `00059_pet_care_events.sql`). No `pet_status_changed.v1`
  is emitted at adoption, because there is no prior band.

**PA4.8 Idempotency and non-verbs.**

- A retried identical intent returns its recorded receipt without invoking the callback, so no
  second nonce is ever drawn (`docs/founder-transitions.md` idempotency). A reused `intent_id` with
  different bytes yields `idempotency_conflict`.
- There is **no** release, abandon, rename, re-roll or transfer verb. Re-rolling a temperament is
  therefore impossible within a Founder. Starting a New Founder is a whole-career cost (OD-10).

### PA5 — Relationship to the existing care state

1. **Care is unchanged.** `care_action`, its receipt, its events, its C18–C21 arithmetic and its
   no-death carve-out are not modified. The only behavioral change is that a production Founder can
   now have a `pet_id` that resolves. `unknown_id` still covers IDs that were never adopted.
2. **Species-agnostic care.** v1 applies the single pinned `pets` care policy to every species. It
   has no per-species decay rate, action variant or life stage. See Deviation DV-2 and GAP-4.
3. **Temperament is stored, not consumed.** Care, the FSM, mood and the combat-input producer
   ignore it (C17: temperament enters the FSM only by a future ruling). Its v1 consumers are the
   surface (presentation) and later combat (PA9).
4. **The combat seam is untouched.** `{pet_trust_ppm, soul}` is produced exactly as before, per pet.
   Pet Care AC3 stays carried to the duel engine. This RFC neither closes nor weakens it.
5. **Exit.** Identity and care are Founder-scoped and survive every Exit byte-identically (PA3.5).
   The Exit arm's Founder-facts update must not touch `pet_identities`. "The company dies; the cat
   does not."
6. **New Founder.** `POST /api/v1/founder` archives the old Founder with its streams, so its pet
   identity and care stay archived with it, never deleted by gameplay. The new Founder starts with
   an empty map and adopts its own starter. No transfer happens in v1 (OD-10, GAP-6).

### PA6 — Save migration, activation, replay

1. **Founder v22 codec (Go + TS).** v22 = v21 plus the required `pet_identities`. Both runtimes
   make these rejections:
   - Company v22;
   - `pet_identities` present before v22;
   - v22 without the key;
   - any PA3 invariant violation;
   - v22 without `pet_species` pinned.

   `LatestFounderVersion` becomes 22. The Company axis is unchanged (independent axes, C16).

   **Version collision.** The sibling draft `rfc/cosmetic-shop-v1.md` §3 also claims Founder v22
   (for `cosmetics`), and it takes "the next free Founder version" if another draft claims v22
   first. On the scalar Founder chain (C17), a higher version requires every lower version's
   artifact. Cosmetic equip keys on pet IDs (`equipped` must be a subset of `pets`), and the
   manifest DAG runs `pet-adoption-v1 → cosmetic-shop-v1`. So this RFC recommends **pets identity
   = v22, cosmetics = v23**, with v23 requiring `pet_species` (OD-15). Whichever RFC is accepted
   second rebases its number in the same acceptance edit.
2. **Migration v21 → v22.** It is only legal when `pets` is **empty**. It adds
   `pet_identities: {}`. A v21 state with any `pets` entry is **rejected**: identity cannot be
   synthesized, and the loader refuses rather than fabricating one (C2 "no invented
   species/temperament"). No production v21 Founder can hold a pet (`docs/pet-care.md`), so this
   rejection guards fixtures and imports, not players. Corpus cases are added to
   `testdata/save-migrations.json` and ratchet the baseline manifest.
3. **Activation (OD-14).** v22 activates on the same two paths as its precedents (Fiscal v19, Soul
   v20):
   - **at a new-run boundary**, when an Exit enters a bundle pinning `pet_species`;
   - **New-Founder-forward**, when account creation, New Founder or import under such a bundle
     creates the Founder directly at v22.

   The Exit per-axis version floors gain the `pet_species` artifact with Founder floor 22. The
   artifact is resolved from the run's pinned hash, never from deploy-current data. Pre-epoch
   Founders synthesize nothing until their next boundary.
4. **Replay inputs.** Company replay-inputs `founder_extensions` (v6 today) must carry
   `pet_identities` when the pinned Founder floor is ≥ 22, so that same-epoch Founder state cannot
   disappear across Exit. The mechanism is a **replay-inputs v7** (precedent: v6 introduced
   `founder_extensions`). Historical v6 stays readable for its pinned semantics.
5. **Founder replay.** `LoadFounderHistory` replay in Go and TS gains the `adopt_pet` arm. The
   failure classes are:
   - a missing `pet_species` bytes, reported as `constants_mismatch`;
   - a tampered nonce or a derived-value mismatch, reported as `state_divergence`;
   - a tampered receipt or event bytes, reported as `state_divergence`.

   `client/src/replay.ts` adds the arm and the `pet_adopted.v1` event kind.
6. **Economy isolation.** Adoption touches no Company state. The balance-harness pacing outputs
   and Company replay vectors must be byte-identical with and without an adoption in the Founder
   log (AC15).

### PA7 — Read model (snapshot projection)

The owner's own Game UI snapshot gains two fields in the next snapshot schema version. The version
number is coordinated under OD-11.

- `pets`: an array sorted by `(adopted_at_attended_ms, pet_id)` of exact
  `{pet_id, species_id, temperament, palette_id, name_key, status_band, eligible_action_ids}`.
  **No raw stats, Trust, remainders, cooldown cursors, behavior queue or mood.** This keeps C20's
  projection-privacy rule ("snapshots expose ONLY the band + eligible action IDs"). The identity
  fields are the owner's own immutable facts, not care internals.
- `pet_adoption`: `null` when the Founder is below v22 or the artifact is unpinned. Otherwise it is
  exact `{cap, count, starter_species_id, name_keys}`, derived from the pinned artifact.
  `cap`/`count` let the surface show the visible hardcap.

Nothing new becomes guild-, world- or feed-visible. The guild-visible status projection stays the
C7 successor's job.

### PA8 — Adoption surface contract (the cattery port, first slice)

This RFC owns the **adoption** surface and the **pet visual contract**. The care panel stays with
`garage-player-surfaces`. That RFC consumes PA7 and PA8.5 and must not fork them.

**PA8.1 Placement and flow.**

- When `pet_adoption` is non-null and `count < cap`, the Desk shows an **inline, non-modal**
  adoption card. The one-modal budget belongs to the ripe Fiscal Quarter (`design/11` §1).
- The card shows the starter species name and description copy, the name pool as a choice, a
  primary **Adopt** control and a secondary **Not now** control.
- **Not now** collapses the card to a small persistent entry point on the Desk. It never
  disappears, re-nags, badges, counts down or escalates. It sets no local timer, because the pet
  can always wait (no-loss canon).
- On an applied receipt, the card is replaced by a welcome state showing the pet's name, species,
  temperament label and sprite. After that, the care panel (the successor surface) takes over.

**PA8.2 Sincerity rules (binding, testable).** Inside the adoption surface and every `pet.*`
copy key:

- no corporate/satire voice, curtain disclosure, statistic, price token (`$`, `0.00`, "free",
  "cost"), urgency ("now", "limited", timers), guilt or streak copy, era parody chrome, or joke at
  the pet's expense;
- the surface uses the calmer "sincerity island" styling (`design/11` §7), with era chrome
  receding;
- temperament labels describe disposition warmly and never as a defect;
- pet barks are non-verbal or sincere (`design/11` §5; research §9 I: "affect, not dialogue").

**PA8.3 Accessibility (under D-018 / F03).**

- Every control is a native, keyboard-operable DOM control.
- The name pool is a single labelled radiogroup with arrow-key navigation. The first name is
  preselected, and Adopt is never disabled merely because of the default choice.
- The visible focus order is species description, name choice, Adopt, Not now.
- After an applied receipt, focus moves to the welcome heading, and **one** polite live-region
  announcement is made (`pet.adoption.announce.adopted` with `{pet_name}`). Retries and resyncs
  never re-announce.
- A rejection shows its reason text inline, associated with the Adopt control. Focus stays put.
- The sprite is `aria-hidden`. An adjacent text alternative states name, species and temperament.
- Status is never color-only: the palette is decorative, and the band is also shown as text.
- The surface reflows at 320 CSS px without horizontal scroll, and targets are at least 24×24 CSS
  px.
- Browser tests fail on any uncaught error (evidence discipline 1).

**PA8.4 Reduced motion.**

- The OS `prefers-reduced-motion` query, and the in-game setting if D-018 establishes one, drive
  the sprite.
- Under reduced motion the pet renders a **static pose**. It is not hidden, because the pet is the
  surface's subject (OD-9 and DV-5; cattery hid the stage entirely).
- Under reduced motion there is no arrival or idle loop, no transition between poses, and no
  parallax.
- In any motion mode, nothing flashes more than 3 times per second.
- Entering reduced motion mid-animation snaps to the static pose of the current authoritative
  band. Leaving it affects only future transitions, per the accessibility draft's A4 posture.

**PA8.5 Pet visual contract (ports `design/04` §6).**

- One pure TS function,
  `petVisualSpec(identity, status_band, reduced_motion) → {family, palette_id, pose, animate}`,
  closed over `family ∈ {cat}` and `pose ∈ {content, low, withdrawn}`.
- `status_band` maps to pose as follows: `high`/`normal` → `content`, `low` → `low`, `floor` →
  `withdrawn`.
- `animate` is `false` whenever `reduced_motion` is on.
- The renderer is a CSS-only nested-div sprite driven by custom properties and `data-*` attributes,
  with zero image assets.
- Palette swatches live in a presentation table keyed by `palette_id`. Each swatch pair must pass
  the non-text contrast check against its background in every era theme.
- This contract re-implements the cattery *pattern* (`research/cattery-reusables-public.md`). No
  sibling source is imported.
- Behavior-state poses are excluded because behavior state is not in the snapshot (PA7). Adding
  them is a successor.
- `cosmetic-shop-v1` needs a wearer reaction pose (`annoyed`). It extends the `pose` union in its
  own RFC, and this contract stays the single function it extends, with no parallel renderer. Its
  reaction must stay genuine animal behavior (PA8.2), and under reduced motion it must be a
  static pose.

**PA8.6 Copy keys (keys only; all prose is owner-authored and pending).** The tone is `companion`
(OD-8). It is a new member of the closed `CopyTone` union (`client/src/copy/index.ts:9`), and the
copy linter gets a companion-tone rule set that enforces PA8.2. The keys are:

- `pet.species.server_room_cat.name`, `pet.species.server_room_cat.description`;
- `pet.name.server_room_cat.00` … `.NN` (the pool; size is the owner's choice under OD-4);
- `pet.temperament.{lazy,playful,curious,sassy,shy,chaotic}.label`;
- `pet.adoption.card.title`, `pet.adoption.card.body`, `pet.adoption.name_choice.label`,
  `pet.adoption.action.adopt`, `pet.adoption.action.later`, `pet.adoption.entry_point`;
- `pet.adoption.welcome.title` (`{pet_name}`), `pet.adoption.welcome.body` (`{pet_name}`),
  `pet.adoption.announce.adopted` (`{pet_name}`), `pet.adoption.sprite.alt`
  (`{pet_name}`, `{species}`, `{temperament}`);
- `pet.adoption.reject.adoption_inactive`, `.species_locked`, `.adoption_cap_reached`;
- `pet.adoption.cap.label` (`{count}`, `{cap}`).

The implementer adds keys with placeholder status only. No English text lands without owner
adoption (evidence discipline 6). The artifact cannot be minted while any referenced key lacks
adopted text.

### PA9 — Conflict avoidance with pet battles (`design/03` §10b, `design/04` §2)

1. **Namespace.** Combat already has a `species` concept: `{id, temperament, runtime, stats,
   moves}`, 36 identities (`rfc/combat-data-model.md`). Pet-layer species IDs are namespaced
   `pet_species.*`, and **no equality or implicit mapping** exists between a pet species and a
   combat species. The binding of a pet to a combat identity (its runtime axis, its moves) is
   GAP-5, owned by the combat integration successor.
2. **Temperament is shared.** The pet temperament *is* the combat Temperament enum (one identity
   system, PC4/C15). A pet's temperament is immutable after adoption, so a future duel snapshot
   can freeze it.
3. **Team size.** Duel teams are exactly three species (`rfc/combat-data-model.md`), while the v1
   pet cap is 1. This RFC does not resolve how owned pets populate a team. The cap is data, and
   raising it is a balance change, not a wire change.
4. **No power.** Nothing adopted here carries a stat. Species, palette and name are cosmetic, and
   temperament is a chart position, not a strength (`design/04` §2 hardcap).
5. **Tone guard.** Battles are play-fights, and nothing in this surface frames the pet as a
   fighter.

### PA10 — Data rights (manifest EC-5)

- Pet identity and care are Founder-scoped player data: a new v0.1 data class.
- Export (D-008) includes `pet_identities` and `pets`.
- Deletion and anonymization (D-009) follow Founder archival.
- Retention (D-015) follows the Founder save policy.
- With OD-4's curated names there is **no free text and no PII** in pet data.
- If D-008/D-009/D-015 are still open at implementation, AC16 is carried under the release floor.
  It is not waived.

## Deviations from design

- **DV-1 (from archived Pet Care C11/C14).** Identity lives in Founder save `pet_identities`
  (replay-owned). It does not live in a relational `pet_records` table. Rationale: C14's own
  discipline ("no second mutable authority beside the state the transition boundary owns")
  applies, Founder replay stays self-contained, and the table was never built. The recorded
  `founder_id` and `created_at` are implied by the owning stream and `adopted_at_ms`. OD-5 offers
  the table instead.
- **DV-2 (`design/04` §1).** v1 has no care-action variant overrides and no life-stage decay
  (kitten/adult/elder). All species share one care policy. This is required by the hardcap-law
  reading and by the absence of any stage model (GAP-4).
- **DV-3 (`design/04` §1, `design/03` §Unlock stagger).** The four later variants are not minted
  (OD-1, which matches manifest VQ-9).
- **DV-4 (archived Pet Care C11 ruling text).** C11 had starter creation run "in the New-Founder
  transaction". Here it is an explicit `adopt_pet` intent instead: a meeting, not an allocation.
  The C2 proposal it came from allowed "a separately ruled acquisition command". See OD-2.
- **DV-5 (`research/cattery-reusables.md`).** Under reduced motion the pet renders a static pose
  rather than the cattery's hidden stage (OD-9).
- **Pre-existing, not introduced here:**
  - `design/03` §Clock taxonomy calls pet care "wall-clock", but shipped care runs on attended
    time. Adoption follows the shipped attended clock.
  - `design/04` §1 names the stats Fed/Happy/Energy/Clean, while the wire IDs are
    `hunger|energy|cleanliness|affection`. Display labels are copy.

## DESIGN-GAPs (surfaced, not improvised)

- **GAP-1** Variant unlock predicates ("T2", "conspiracy-tier") need a Founder-owned tier or
  career fact, because a Founder command cannot read Company tier (C1). The same gap blocks every
  non-starter `availability` member.
- **GAP-2** Design does not say how a pet's temperament is set. PA4.4 is OD-3's recommended
  answer, pending ruling.
- **GAP-3** Design does not say how a pet is named. It shows only the `{pet_name}` slot
  (`design/11` §6). See OD-4.
- **GAP-4** The life stages and elder retirement (`design/04` §1, §5) have no state model.
- **GAP-5** The binding of a pet to a combat identity (runtime axis, moves, the three-pet duel
  team) is unspecified.
- **GAP-6** Design does not address what happens to the pet on New Founder. See OD-10.
- **GAP-7** `design/04` has no `§tone` heading. The sincerity law is stated in §1 (Trust bullet)
  and §2 (tone guard). Citations elsewhere to "§tone" have no anchor. That is an owner design
  edit, not one for this RFC.
- **GAP-8** Archived Pet Care AC5 claims "the sprite contract typed for UI", but no typed pet
  sprite/visual contract exists in `client/src` or `docs/` at `34462f7`. PA8.5 supplies it, and
  the discrepancy should be recorded by the designated reviewer.

## Acceptance criteria

Every criterion names a **failing case that must be demonstrated red** before it counts (evidence
discipline 1). Gate claims use `-count=1`.

1. **AC1 Catalog loader (Go/TS parity).** One shared fixture set loads identically in both runtimes
   and joins constants identity. *Failing cases:*
   - an unknown key;
   - zero or two `starter` rows;
   - a temperament outside the six, or out of canonical order;
   - an unsorted or duplicate `palette_ids` or `name_keys`;
   - `max_pets_per_founder: 0`;
   - `pet_species` pinned without `pets`;
   - a `name_key` absent from the copy catalog or not tone `companion`;
   - a `species_id` outside the `pet_species.` namespace.

   Each must fail load in both runtimes.
2. **AC2 Founder v22 codec.** *Failing cases:*
   - an identity/care key-set mismatch in either direction;
   - `pet_identities` at v21;
   - v22 missing the key;
   - Company v22;
   - over-cap;
   - an unknown species, palette or name under the pinned artifact;
   - a non-v7 pet ID;
   - v21→v22 with a non-empty `pets` map (must reject, not fabricate).
3. **AC3 Draws are nonce-derived and grinding-proof.** Shared Go/TS byte vectors for pet ID,
   temperament and palette. *Failing cases:* a variant that includes `intent_id` in any hash input
   must turn the "same nonce, different `intent_id` ⇒ identical draws" vector red. A variant using
   digest bytes other than `[0..8]`, or a wrong version/variant nibble, must turn the vectors red.
4. **AC4 Validation order and receipts.** One vector per PA4.2 row, with exact receipt bytes.
   *Failing cases:* swapping rows 3 and 6 (unknown species checked after cap) must turn the
   combined unknown-species-at-cap vector red. A second adoption at cap 1 must reject with
   `adoption_cap_reached`.
5. **AC5 Initial care record and care integration.** An adopted pet is byte-equal to the PA4.6
   canonical record. `care_action` on it applies. A pet adopted after `A` ms of attendance shows
   zero decay for that span. *Failing case:* seeding `evaluated_through_attended_ms = 0` must turn
   the first care vector red through retroactive decay.
6. **AC6 Identity immutability and Exit carry.** `pet_identities` is byte-identical across care,
   Fiscal, minigame, Soul, route-hint and Exit transitions. *Failing cases:* a test-only arm
   mutating a temperament, and an Exit that drops the map, must each fail the authoritative
   transaction.
7. **AC7 Founder replay.** Go and TS replay of a history containing an adoption reaches the saved
   head. *Failing cases:* a tampered nonce gives `state_divergence`, stripped `pet_species` bytes
   give `constants_mismatch`, and tampered `pet_adopted.v1` payload bytes give `state_divergence`.
8. **AC8 Replay-inputs v7.** Company replay carries `pet_identities` in `founder_extensions` at
   Founder floor ≥ 22, and v6 stays readable. *Failing case:* an Exit replay whose extension omits
   identities must reject.
9. **AC9 Event registry and migration.** `pet_adopted.v1` is registered in the Go and TS event
   unions and in the database constraint through a new append-only migration. *Failing case:*
   inserting an unregistered kind is rejected by the constraint in the Postgres suite (`docker
   compose -f compose.save-test.yml run --rm test`).
10. **AC10 Idempotency.** A retried identical intent returns the byte-identical receipt and draws
    no nonce. *Failing case:* an instrumented nonce source that counts draws must read 1 after a
    retry. A build that re-executes the callback on retry must fail. A different payload under the
    same `intent_id` yields `idempotency_conflict`.
11. **AC11 Snapshot projection.** Schema tests pin the PA7 exact keys. *Failing cases:* a projector
    emitting `stats_ppm`, `trust_ppm`, `mood` or `behavior_queue` in `pets` must fail the schema
    test. `pet_adoption` must be `null` pre-v22.
12. **AC12 Surface accessibility and reduced motion.** A browser test covers keyboard-only
    adoption, focus landing on the welcome heading, exactly one live-region announcement, 320 px
    reflow, axe at WCAG 2.2 AA, and a static pose with `animate:false` under emulated
    reduced motion. *Failing cases:*
    - a sprite carrying an animation class under reduced motion must fail;
    - a second announcement on resync must fail;
    - any uncaught page error must fail the run even if assertions pass.

    Full R-005 evidence follows the D-018 matrix.
13. **AC13 Sincerity and no-price lint.** Every `pet.*` key has tone `companion`. The companion
    linter rejects price tokens, urgency tokens, statistic-detector hits and curtain/disclosure
    phrasing. *Failing cases:* seeded `companion` rows containing "$0.00" and "limited time" must
    fail `make copy-generate`/verify. A DOM test must find no price text in the adoption surface.
    The owner's sincerity review is recorded in the planning log (manifest EC-4).
14. **AC14 Activation boundary.** Pre-epoch Founders stay v21 with nothing synthesized. Exit into
    a `pet_species` bundle yields v22 with `{}`. New Founder under that bundle is created at v22.
    *Failing case:* resolving `pet_species` from deploy-current data instead of the run's pinned
    hash must fail the mixed-epoch fixture.
15. **AC15 Economy isolation.** Harness pacing outputs and Company replay vectors are
    byte-identical with and without an adoption. *Failing case:* a test-only adoption arm that
    writes any Company field must turn the comparison red.
16. **AC16 Data rights.** Export contains identities and care, and the deletion witness covers
    them. This AC is carried under the release floor if D-008/D-009/D-015 are unruled.
17. **AC17 Canon.** `docs/pet-care.md` (acquisition now exists), `docs/save-layer.md` (v22),
    `docs/founder-transitions.md` (adopt arm), `docs/production-engine.md` (replay-inputs v7) and
    `docs/game-ui.md` (adoption surface) are updated in the implementing change. The RFC index is
    updated at every status move. Archival follows the cross-party designated-review gate.

## Owner decisions (numbered; recommended default first)

- **OD-1 Scope.** **Recommend: the starter server-room cat only**, which matches manifest VQ-9.
  The four named variants are deferred until GAP-1 has a Founder-owned tier fact. The alternative
  (T2 variants in v0.1) requires first ruling GAP-1 and authoring three new sprite families.
- **OD-2 Acquisition shape.** **Recommend: the explicit `adopt_pet` intent at T0** (DV-4). It
  makes a sincere meeting beat, leaves the account-creation transaction untouched, uses one path
  for new and existing Founders, and gives replay an ordinary log row. The alternative is
  auto-creation atomic with New Founder, per C11 as written.
- **OD-3 Temperament.** **Recommend: a server-nonce uniform draw over the species'
  `allowed_temperaments`**, published in docs and not shown on the surface. The research says a
  pet's character should be *met*, not configured, and the nonce makes grinding impossible.
  Alternatives: the player chooses (no RNG, but temperament becomes a combat-chart pick), or a
  fixed temperament per species.
- **OD-4 Naming.** **Recommend: pick one name from an owner-authored curated pool** (suggest 12
  names), immutable in v1. That means no free text, no moderation, no PII, and a `{pet_name}` slot
  that is always safe. Alternatives: no name (the species label only), or free text. Free text is
  **not recommended**: it would need moderation, and its privacy and guild-visibility impact runs
  into D-008/D-009/D-017.
- **OD-5 Identity storage.** **Recommend: Founder-save `pet_identities`** (DV-1). The alternative
  is the C11 relational `pet_records` table, append-only and trigger-protected, written in the
  same transaction and equality-checked against the save.
- **OD-6 Artifact split.** **Recommend: a separate `pet_species` artifact.** The alternative
  extends the `pets` artifact to schema v3 with a `species` key, which re-opens C17's pinned top
  level.
- **OD-7 Palette identity.** **Recommend: include `palette_id`**, with 10 swatches ported from the
  cattery fur palette as contrast-checked presentation data. Recognition ("my cat is the orange
  one") is cheap attachment. The alternative drops it until breeding (`design/04` §5).
- **OD-8 Copy tone.** **Recommend: a new `companion` tone member with its own linter rules.** The
  alternative reuses `lore_card` (the silent register), which carries no pet-specific lint.
- **OD-9 Reduced motion.** **Recommend: a static pose.** The alternative hides the sprite, as
  cattery did.
- **OD-10 New Founder.** **Recommend: the pet stays archived with its Founder, and the new Founder
  adopts a fresh starter.** No transfer happens in v1. The alternative is a transfer, which needs
  cross-Founder identity rules. Any copy for the New Founder confirmation is the owner's.
- **OD-11 Snapshot version.** **Recommend: the PA7 fields are additive, and whichever of this RFC
  or `garage-player-surfaces` lands first takes the next snapshot version**, with the other
  rebasing. The alternative bundles both into one pre-agreed snapshot bump.
- **OD-12 Pet cap.** **Recommend: `max_pets_per_founder = 1`**, shown as a visible cap. Raising it
  waits for variants and the combat team binding.
- **OD-13 Soul gate on adoption.** **Recommend: none.** A companion is essential, and Soul drain
  acts through recognition and Trust, not through access.
- **OD-14 Activation path.** **Recommend: new-run boundary plus New-Founder-forward** (the Fiscal
  v19 and Soul v20 precedent). The alternative activates at the first Founder command under a
  pinned catalog, which the C16 ruling also allows.
- **OD-15 Founder-version order with `cosmetic-shop-v1`.** **Recommend: pet identity = v22, then
  cosmetics = v23**, with v23 requiring the `pet_species` artifact. That follows the manifest DAG
  and the cosmetics equip dependency on pet IDs. The alternative puts cosmetics at v22 and pets at
  v23. That forces pet adoption to require the cosmetics artifact, which inverts the dependency.
  Deferring both to the feature-vector envelope (C17's named successor) is not recommended for
  v0.1.

## Open questions

- Owner-authored text for every PA8.6 key, including the name pool. This blocks minting, not
  fixture implementation.
- D-018 matrix and in-game reduced-motion setting (AC12 full evidence).
- D-008/D-009/D-015 (AC16).

## Changelog

- 2026-09-25: created (draft). Claude drafted it for the v0.1 "Garage" manifest item G14.
