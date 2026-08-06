# Pet Care

The implemented Pet Care foundation owns the cross-runtime wire grammar, replay-owned Founder v18
state, the pure care transition, and the server-authoritative `care_action` command. Production pet
identity/species rows and combat consumption remain unimplemented and are not claimed here.

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

This state is embedded into replay-owned Founder save v18. The pinned pet artifact is the complete
closed union of stat-grid, action, Trust, mood-threshold, and deterministic behavior-transition
policies; pinning the earlier mood/behavior fixture alone is rejected. v18 requires the v17
minigames artifact to remain pinned, while Company remains v14/v16. Mood stays derived and is
never persisted. Numeric policy rows remain fixture/balance data and no production pet is enabled.

## Care transition and authoritative intent

Go and TypeScript execute the same pinned-catalog transition. A care command first advances decay
on the Founder attended-time grid, preserving an exact remainder and rejecting stale per-pet
watermarks. Stats stop at their declared floors and Trust moves monotonically toward neutral.
Eligibility, diminishing returns, cooldowns, positive-only Trust grants, mood derivation, and the
bounded deterministic behavior FSM then run in the RFC-defined order. Large behavior intervals use
cycle skipping; there is no per-grid server tick loop.

The client sends only `{intent_id,kind:"care_action",expected_revision,pet_id,action_id}`. The
server selects the active Company sibling and freezes its attendance sample into the immutable
Founder log; the client cannot choose a clock context. Applied commands emit
`pet_care_applied.v1` and, only when the public band changes, `pet_status_changed.v1`. The exact
receipt and ordered events are compared against one shared Go-authored replay vector in the
TypeScript suite. Rejected commands use `unknown_id` for missing pets/actions and `not_eligible`
for cooldown, eligibility, or saturation failures.

The fixed-grid integration helper is shared with the minigame faucet and future offline-quality
decay, so interval partitioning cannot change the result. Empty behavior queues remain canonical
arrays across save cloning and persistence; they are never normalized to `null`.

New-schema action rows also declare `soul_gate: essential|recovery|ordinary`. A bundle may use
that schema with Soul only when it pins the same Soul artifact; the server resolves the exact
Founder state and bundle. Near-zero Soul rejects ordinary care but never essential or recovery
actions. Historical pet artifacts keep their old grammar and remain valid only in Soul-less
bundles.

## Combat input seam

The foundation exports the exact-safe pair `{pet_trust_ppm,soul}` as a pure Go/TypeScript producer.
It does not yet bind a combat engine or close combat C5; that closes only when a combat-owned
Obedience table consumes the pair.
