# Pet Care

The implemented Pet Care slice currently owns only the cross-runtime wire grammar. Pet identity,
Founder persistence, care/decay arithmetic, trust changes, mood derivation, behavior transitions,
public status projection, and production species remain unimplemented and are not claimed here.

## Closed Phase-A vocabulary

The four stat IDs are `hunger`, `energy`, `cleanliness`, and `affection`. Status bands are
`floor`, `low`, `normal`, and `high`; moods are `withdrawn`, `restless`, `neutral`, and `engaged`;
behavior states are `idle`, `care_response`, `active`, and `resting`; behavior events are
`grid_tick`, `care_applied`, and `care_rejected`.

Care rejection details are `cooldown`, `ineligible`, `saturated`, `unknown_pet`, and
`unknown_action`. A behavior queue contains at most 8 entries, and deterministic behavior draws
use the label `pet.behavior.v1`. Go and TypeScript consume one shared parity fixture for every
member and the queue boundary. These names are protocol grammar, not balance data.

Thresholds, durations, candidate weights, stat deltas, species, and temperaments remain catalog
content. No production pet row is synthesized from the protocol vocabulary.
