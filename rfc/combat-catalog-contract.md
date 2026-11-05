# RFC: Combat Catalog Contract

- **Status:** draft — proposal, not implementation/content authority
- **Author:** Codex (for author/owner reconciliation)
- **Created:** 2026-11-05
- **Parent / amends:** Combat Shared Data C1/C2/C4, AC1; no child adoption
- **Design refs:** `design/04-pets.md` §2; `design/03-minigames.md` §10
- **Planning:** `planning/combat-shared-data/`
- **Depends on:** adoption of `combat-care-tables.md`; current strict row readers

## Outcome and source boundary

Make the accepted parent's complete catalog buildable. Existing Go/TS readers
decode species, lane rows and loadout references; they deliberately cannot
invent effects, legal numeric domains, a versioned envelope, cross-row coverage
or pinning. This draft supplies a concrete proposal for those missing contracts.
Nothing here is a minted artifact or evidence a battle can be played.

Sources checked at `d38a4372`: parent C1–C5; both child-engine drafts and the
DU-C1–DU-C8 owner block; `server/combat` / `client/src/combat` readers;
`server/replaycatalog/catalog.go`; `server/production/replay.go`;
`server/save/state.go`'s exact-byte constants identity; Pet Adoption PA9.
Research-only candidate moves and lane figures motivate examples, not acceptance
thresholds or calibrated player values. The research reports do not establish
balance for the still-unbuilt engines.

## Proposed contract

### CC1 — Exact artifact envelope and identities

One UTF-8 JSON artifact, family/name `combat`, schema version1:

```text
{schema_version, tag_vocabulary, runtimes, rng_labels, policy,
 species, moves, lane_units, lane_buildings, lane_spells, care, critical}
```

Every member is required and non-null. Objects have exact unique decoded keys,
including escaped spellings; no extra fields, duplicate keys or defaulting.
`schema_version` is the integer token1. Integer fields admit lexical integer
tokens only, as current Combat row decoders do; fraction/exponent tokens,
booleans, strings and out-of-domain values refuse. Whitespace and object-member
order are immaterial to admission; malformed UTF-8/surrogate text refuses.
Loaders do not repair/rewrite input bytes or return a partial catalog.

Root row arrays are sorted by their ASCII `id` and have no duplicate IDs.
Schema1 requires at least12 species,24 moves,8 units,3 buildings and3 spells,
with at least two species per temperament; the committed fixture has exactly
those minima. This proposes a complete-catalog admission floor, not a rule for
a child's separately specified experimental subset or player loadout.
IDs use `[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)*`, with family prefixes
`combat_species.`, `combat_move.`, `combat_unit.`, `combat_building.` or
`combat_spell.` as appropriate. Lane IDs are unique across all three families.
These are new full-catalog admission rules, not retroactive narrowing of the
already-shipped standalone structural row APIs.

`tag_vocabulary` is a sorted unique array of nonempty IDs in the same grammar.
All row tag arrays are sorted unique declared references. Empty tags/vocabulary
are allowed for the minimum roster; roster growth beyond12 additionally requires
the parent's populated tag vocabulary and the later bots/integration harness
gate. This loader is not proof of tag usefulness or roster balance.

`runtimes` has exactly six `{id,tags}` rows, one for each existing C1 runtime.
Runtime tags and species tags are metadata declarations; this draft adds no
synergy effect, implicit stat modifier or second type chart. Preserve each list;
any eventual engine interpretation of tags requires its own accepted contract.

`rng_labels` is a sorted unique string array containing `bot_policy`, `crit`,
`obedience`. Additional declared labels must be nonempty ASCII IDs in the same
grammar. The existing labeled-stream algorithm/draw semantics are unchanged.

`care` is exactly the object proposed in Combat Care Tables CT2, evaluated by
CT1/CT3. That proposal must be adopted first; this reference is not adoption.

### CC2 — Species and moves

Species retain the parent's exact row and existing axes. HP and stamina_max are
in `[1, INT32_MAX]`; ATK/DEF/SPD in `[0, INT32_MAX]`. These are fixed identity
ceilings, not care-scaled live stats. The four move slots are ordered declared
references; no sorting or inferred ownership. No new duplicate-selection rule
is imposed on teams, decks or species move slots by this contract.

Moves have exactly:

```text
{id, temperament, power, stamina_cost, kind, effect}
```

Temperament is null or one existing C1 value. Power/stamina_cost are integers
in `[0, INT32_MAX]`. Kind is `strike|coverage|guard|utility`. The effect union is:

| Effect object, exact keys | Parameter admission / declared purpose |
|---|---|
| `{type:"none"}` | No additional effect payload |
| `{type:"guard", damage_numerator, damage_denominator}` | `1 <= numerator <= denominator <= INT32_MAX`; declares a damage-retention ratio, not an ATK/DEF ceiling increase |
| `{type:"restore_hp", amount}` | `1 <= amount <= INT32_MAX`; self-restoration bounded by the species ceiling when an engine resolves it |
| `{type:"priority", amount}` | `1 <= amount <= INT32_MAX`; declares action-priority data |
| `{type:"drain_stamina", amount}` | `1 <= amount <= INT32_MAX`; opponent stamina reduction bounded by zero when an engine resolves it |

Strike/coverage require positive power and `none` or `priority`. Guard requires
power0, null temperament and `guard`. Utility requires power0, null temperament
and `restore_hp` or `drain_stamina`. Typed coverage must be non-null and differ
from every referencing species' own temperament. Every species must declare at
least one same-temperament strike and one off-temperament coverage move; missing
coverage refuses instead of relying on a tag or unreferenced move elsewhere.

The parent loader admits payloads; it does not resolve actions. The effect
semantics above name capabilities only. Priority versus SPD, guard timing and
rounding/store order, faint cancellation, targets at resolution time, stamina
debit and event sequences still require the deferred DU-C4 ruling. No operation
may overflow/wrap or bypass C3 when that engine is specified. Do not implement
an effect executor or infer an ordered transition from this data proposal.

### CC3 — Lane cards, capacity units and bypass coverage

Units/buildings retain the exact existing common row and optional `bypass`.
Spells have that common row **plus required `effect`**; their loader is a new
full-catalog path, not an expansion of the old standalone common-row API.

All costs are **capacity deci-units**, the same units as child L1's
`capacity_dds`: cost10 means1.0 capacity, cost100 means the full10.0 cap. This
explicitly chooses the previously unnamed unit; there is no hidden ×10 at the
comparison/debit boundary. Require `1 <= cost <= 100` for schema1.

HP/ATK/SPD/range are signed-int32 representations with these family constraints:

| Family | HP | ATK | SPD | Range |
|---|---:|---:|---:|---:|
| Unit | 1..INT32_MAX | 0..INT32_MAX | 1..1000 | 0..1000 |
| Building | 1..INT32_MAX | 0..INT32_MAX | 0 | 0..1000 |
| Spell | 0 | 0 | 0 | 0..1000 |

All existing temperament/target/tag rules apply. Spell effect is the closed
exact object `{type:"damage", power, radius}`: power in `[1,INT32_MAX]`, radius
in `[0,1000]`. This holds the damage payload once; common spell ATK is a zero
sentinel, not another power authority. Target selection, placement/aim position,
area boundaries, damage scaling and tick/event ordering are not supplied here;
they remain requirements of the unaccepted Lane engine.

`policy` is exactly `{cheap_cycle_max_cost_dds}` with integer value in `[1,100]`.
This declared threshold is balance data; it is not inferred from ID/name/tag.
Require all three catalog-wide predicates, not merely matching IDs:

1. A lane unit with `bypass:true`, `targets:"buildings"` and positive SPD.
2. A lane unit with `bypass:true` and cost at most the declared cheap-cycle limit.
3. A lane spell with `bypass:true` and the admitted positive damage payload.

The same unit may satisfy1/2. Absence and explicit false do not satisfy either.
The bypass marker does not authorize deploy-anywhere behavior, ignore body
blocking or prove decisive matches. Those rules and the exact removal population
belong to the Lane engine/bots gates. Full acceptance must measure actual
decisiveness; a present flag is not evidence that a bypass verb works.
The existing child proposal's “catalog minus bypass rows” baseline cannot be
admitted as an ordinary catalog by this required presence check. Its later
gate contract must specify a controlled ablation population without pretending
an invalid artifact passed this loader. Do not disable validation or silently
change that child criterion while implementing the parent.

### CC4 — Critical-chance data without a snapshot ruling

`critical` is exactly `{chance_by_trust, factor_by_soul}`. Each value uses CT2's
point-table grammar/domain, with the same monotonicity, endpoint clamping and
int64/BigInt interpolation bounds. Compute a separate critical-chance result:

```text
critical_ppm = floor(lerp(chance_by_trust, pet_trust_ppm)
                    * lerp(factor_by_soul, soul) / 1_000_000)
```

This is a proposed care-derived probability, not a change to C3's fixed3/2 crit
damage multiplier or to CT3's exact `{obedience_ppm}` result. Keep its pure
function separate; do not add an unannounced CT2 field. Draw order, eligibility
to roll, per-instance input binding, snapshot inclusion and event emission stay
with the future child ruling. Technical approval does not adopt this policy.

### CC5 — Concrete test-fixture recipe, not live content

Build the parent's committed synthetic fixture on implementation, under
`balance/testdata/combat/fixture-v1.json`. It must not be added to a production
epoch or copy catalog. This recipe fixes every row so the first implementation
does not choose additional balance values while writing tests:

- Temperament index `t=0..5` follows the existing C1 order; runtime index uses
  `container,vm,bare_metal,serverless,edge,mainframe`. Declare the six
  `runtime.<runtime>` tags; each runtime row lists its corresponding tag.
- Twelve species `combat_species.s00`..`s11`: row index `2*t+j`, j in0/1;
  temperament t, runtime `(t+3*j) mod6`, stats `{hp:128,atk:64,def:64,spd:16,
  stamina_max:10}`, tags empty. Ordered moves are strike_t, coverage_(t+3 mod6),
  guard_t, utility_t. Runtime does not change those ceilings.
- Twenty-four moves, four for each temperament. IDs
  `combat_move.<kind>_<temperament>`; cost2 for strikes/coverage and1 for
  guard/utility. Strike power24, own temperament, effect `none` for even t
  and `priority{amount:1}` for odd t. Coverage power18, own temperament,
  effect `none` (referenced by opposite-temperament species). Guard power0,
  neutral, ratio1/2. Utility power0, neutral: `restore_hp{amount:12}` for
  even t and `drain_stamina{amount:2}` for odd t.
- Eight units `combat_unit.u00`..`u07`: temperament `index mod6`, HP64,
  ATK16, SPD10, range20, empty tags. u00 costs30, targets buildings and has
  bypass true; u01 costs10, targets any and has bypass true; u02..u07 cost40,
  target any and omit bypass. Cheap-cycle limit10.
- Three buildings `combat_building.b00`..`b02`: temperament `index`, cost40,
  HP128, ATK16, SPD0, range100, targets any, empty tags, bypass omitted.
- Three spells `combat_spell.s00`..`s02`: temperament `index`, cost30,
  HP/ATK/SPD0, range1000, targets any, empty tags, bypass true;
  effect `damage{power:24,radius:50}`.
- Care uses CT4's exact synthetic fixture. Critical Trust points
  `(0,62500),(800000,125000),(1000000,125000)` and Soul points
  `(0,0),(100,1000000)` are synthetic coefficient controls, not research-tuned
  rates or an adopted doubling threshold. RNG labels are the three required
  labels. Sort all root ID arrays and tag/label sets by ASCII after construction.

Every union arm appears in this fixture; this is still not a fun/balanced roster
or a representative match population. The twelve combat species have **no**
implicit mapping to `pet_species.*`. Pet Adoption PA9/GAP-5 still owns that
explicit future binding and the one-pet/three-pet team question; do not raise
the pet cap, fabricate owned pets or change adoption distributions here.

### CC6 — Registration and exact-byte identity

Register optional artifact name `combat` with the existing replay-bundle loader
and both-runtime owner validation. Its `catalog_hash` is lowercase64-hex SHA256
of **original artifact bytes**, not canonical reserialization, the FNV state
hash or the named-set `constants_hash`. The existing named-set hash also covers
those exact bytes; neither hash is stored inside the artifact it identifies.

Old bundles without combat retain exact admission, floors, replay and bytes.
Adding the optional artifact alone invents no Company/Founder fields or version
floor, minigame engine ref, tenant, HTTP resource or player entry point. Missing
or malformed combat on a path that explicitly requires it fails closed, never
falls back to deploy-current content. The future accepted engine/integration
must bind that path's identity, persistence and replay before live activation.

Register authoritative Combat runtime prefixes and make the genuine kernel
version change when integrating the loader into existing authority. Allocate
no epoch/schema/replay identity in this proposal; inspect the actual current
maxima at implementation. Do not modify historical artifacts/applied migrations
or add a production pin as part of fixture construction.

### CC7 — Acceptance of the implemented parent

Execute strict Go/TS fixture loading with exact identity, row counts and all
references, not independent decoders that never see the complete artifact.
Verify all36 identity pairs through existing chart consumers while preserving
the full fixture's two species per temperament. Invalid populations include
every union arm's unknown/extra/missing/null/wrong-type/domain cases, duplicate
IDs/keys, unknown tags/references, wrong slots, absent/mistyped/off-type coverage
and each required bypass predicate. Bypass negatives must preserve row counts
and every other admission constraint, isolating one failed predicate while the
other two still hold. From CC5, independently change u00's target to `any`,
raise u01's cost from10 to11, and set every spell's bypass flag to false.
Ordinary row-deletion negatives may supplement these controls, but cannot
replace them: deleting a minimum row also breaks the count floor and fails to
discriminate bypass validation. Test valid escaping and
semantically valid reordered object keys without rewriting hash input bytes.

Exercise actual replay-bundle admission with/without combat, hash mismatch,
same semantics/different original bytes and a malformed optional artifact;
old admitted bundles must remain admitted with their exact historical bytes.
Use actual Go/TS loaders and existing hash authorities. A schema-shaped mock,
row count or successful source scan is not that integrated reader proof.

Run affected Go/vet, Node/native-browser catalog consumers, both division
boundaries and typecheck. Complete CT5 and existing parent arithmetic/RNG/chart
acceptance; obtain the full implementation-range designated verdict. Original
feature acceptance is not supplied by corrective/specification-review delegation.
No battle engine starts until owner DU-C1's parent gate is actually satisfied.
Afterward rule the deferred child contracts against the now-real catalog;
actual battle logs, statistical gates, bot fairness and player integration
remain required. Test coefficients cannot become live content by test passage.

## Adoption and body reconciliation

Request explicit adoption of the proposed row domains, closed effect vocabulary,
capacity-cost unit, metadata/care/critical objects, synthetic fixture recipe and
pinning route, plus delegation to reconcile parent C1/C2/C4/AC1. These include
gameplay-policy structure and are **not** routine wording fixes. Technical
review cannot make that choice. Parent bodies, deferred child rulings, player
copy and current artifacts stay untouched until the applicable authority.

## Changelog

- 2026-11-05: initial complete parent-catalog proposal. Preserve Pet Adoption's
  identity/team boundary and deferred engine transitions; no implementation.
