# Pet Adoption

Pet Adoption v1 (`rfc/pet-adoption-v1.md`) adds the species catalog, immutable pet identity, and
the Founder-scoped adoption path to the Pet Care foundation. It is fixture-first: no epoch pins
`pet_species` until the owner adopts the name pool and the species copy (PA8.6).

## The `pet_species` artifact

`server/pet/species.go` and `client/src/pet/species.ts` load the strict schema-v1 artifact
`{schema_version, max_pets_per_founder, species[]}`. It is exact-key at every level.

- **Species rows** are byte-sorted by the namespaced `pet_species.*` id. Exactly one row has
  `availability: "starter"`. `visual_family` is the closed member `cat`.
- **Temperaments** are a non-empty list in the canonical combat order: lazy, playful, curious,
  sassy, shy, chaotic.
- **Palettes and names:** `palette_ids` and `name_keys` are sorted and unique.
- **Copy:** every name, species-name and description key must be registered and use the
  `companion` tone. Go checks this through the generated `copykeys.CompanionKeys()`.

The fixture is `balance/testdata/pet-species/fixture-v1.json`. The shared loader corpus is
`testdata/pet/species-fixtures-v1.json`.

A bundle may pin `pet_species` only together with `pets` and `reputation_tree`. On the scalar
Founder chain, Founder v23 extends v22's Reputation state. The artifact joins constants identity.

## Draws (published formula, PA4.3–PA4.5)

The server draws a 16-byte `adoption_nonce` from its CSPRNG once, when the command first executes,
and freezes it in the resolved input. The client-chosen `intent_id` never enters any hash input.

- `d = SHA-256("pet.id.v1" ‖ founder_uuid[16] ‖ nonce[16])`.
- **Pet ID** bytes: `b[0..5]` = the command server time in unix ms, 48-bit big-endian;
  `b[6] = 0x70 | d[0]&0x0f`; `b[7] = d[1]`; `b[8] = 0x80 | d[2]&0x3f`; `b[9..15] = d[3..9]`.
- **Temperament** = `allowed_temperaments[ uint64_be(SHA-256("pet.temperament.v1" ‖ nonce)[0..8])
  mod len ]`.
- **Palette** uses the same recipe with the label `"pet.palette.v1"` over `palette_ids`.

The modulo bias is at most 6/2⁶⁴ and is not corrected. `testdata/pet/adoption-draw-vectors-v1.json`
comes from an independent implementation and pins Go/TS parity.

## Founder v23 state

Founder v23 = v22 + the required `pet_identities` map, keyed exactly like `pets`. Each identity is
exact `{species_id, temperament, palette_id, name_key, adopted_at_ms, adopted_at_attended_ms}`.

Both codecs reject:
- a key-set mismatch with `pets`, in either direction;
- a non-UUIDv7 pet ID;
- missing or extra identity keys;
- identities before v23;
- v23 without the map, and Company v23.

Under a pinned `pet_species`, identities must also resolve: the species, temperament, palette and
name must belong to the row, and the count must stay within `max_pets_per_founder`.
`adopted_at_attended_ms` never exceeds the care watermark.

**Activation** happens at a new-run boundary into a `pet_species` bundle, and New-Founder-forward.
It sets `pet_identities: {}`. It is legal only while `pets` is empty, because an identity is never
synthesized. The Company replay carry holds `pet_identities` from replay-inputs v10 onward when the
pinned Founder floor is at least 23. v10 is otherwise byte-identical to v9 apart from the version
field.
