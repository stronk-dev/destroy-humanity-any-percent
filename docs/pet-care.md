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

The two Phase-A catalog row families now have exact cross-runtime grammar. Mood thresholds are
`{mood_member, floor_ppm}` rows containing every closed mood exactly once with strictly ascending
ppm floors. Behavior candidates are `{from_state, event, to_state, duration_grid_ticks}` rows over
the closed behavior state/event unions, with positive exact tick durations and no duplicate
transition tuple. The persisted behavior queue remains hardcapped at eight; thresholds and
durations in the shared fixture are test data, not production balance data.

## Replay-owned mutable state

The isolated state validator now fixes the C14 mutable JSON keys without activating a Founder save
version. Each pet has complete four-stat ppm and decay-remainder maps, declared-action cooldown
cursors, persistent Trust and its remainder, current behavior state/entry cursor, a queue of
declared behaviors bounded to eight entries, and the behavior PRNG cursor. Mood is derived and is
rejected if stored; writable bonds remain absent. Go and TypeScript consume one shared state
fixture and enforce the same exact domains.

This state is not yet embedded into the Founder save. The pinned pet-artifact/Founder-only version
transition remains the C16 contract, so no deploy-current catalog or half-replayable state is
accepted.
