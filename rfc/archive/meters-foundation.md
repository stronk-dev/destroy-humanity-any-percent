# RFC: Meters Foundation (Trust · p(doom))

- **Status:** implemented — mechanics, production epoch-6 catalog, activation, replay, and events
  are shipped and designated-review covered.
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-03
- **Design refs:** `design/02 §7` (moral axis — not spendable; Trust 5 constituencies + Externality ledger), `design/02 §8` (Soul, the personal ledger), `design/09 §2` (pressure meters, forecastable disaster windows), `design/10` (Ethical% consumes the moral stack)
- **Depends on:** Production + Run Genesis + Purchasable Content (implemented — meters mutate inside `ApplyLogged` after Purchasable Content)
- **Owner ruling honored:** breadth-first foundation — the meter MECHANICS, not the events or endings that consume them.
- **Planning:** `planning/archive/meters-foundation/`

## Summary

`MeterBands` exists as a save-state map with no mechanics. This RFC replaces that misleading field
with a closed Company-meter registry, a structural not-spendable boundary, and deterministic band
transitions. Founder Soul remains the existing separate read-only carry value; addressed
Externality remains a ledger owned by World Layer. The moral axis must be a real quantity before
content leans on it, without collapsing those deliberately different systems back into scalars.

## Specification

### M1 — The meter registry (closed catalog family)

`balance/meters/*.json` is strict schema v1. Each row is exact-key
`{id,scope:"company",min_value:0,max_value:100,initial_value,bands,inputs,decay}`. Phase A requires
exactly eleven rows: ten independent Trust axes
`trust.{users|employees|regulators|press|investors}.{standing|grievance}` plus
`doom.probability`. Bands are unique `{id,floor_value}` rows strictly sorted by floor, beginning at zero;
numeric values—not band IDs—are the sole persisted authority. Externality and Soul are forbidden
as meter IDs. The production artifact supplies literal floors/initials as balance data.

### M2 — The not-spendable invariant (enforced, not documented)

**No intent may debit a meter as ledger payment.** Meter IDs and economy resource IDs are disjoint
registries validated at catalog composition; cost targets accept economy resources only. The
meters package cannot import ledger spend APIs. Meters move only through the declared input union:
`ledger_fact {fact_kind,delta}` applies once to a newly emitted fact, and
`contribution_slot {slot,source_id,delta_per_attended_hour}` integrates only while that committed
contribution is non-neutral. Rows carry no decorative `spendable:false` flag. A declared input may
decrease a value; that consequence is not a purchase.

### M3 — Band transitions as facts

After the complete meter transition, meters emit at most one byte-ordered
`meter_band_changed.v1` event each:
`{run_id,meter_id,from_band,to_band,direction,value_before,value_after}`. A multi-band jump reports
the prior and final derived bands, not every intermediate floor. Band changes are presentation and
trigger events only; they never create moral ledger facts. Ethical% remains gated by dated
`darkpattern.*`/`externality.*` acts. All values and bands resolve from the run-pinned catalog.

### M4 — Decay & the attended clock

Decay and rate inputs use exact carried integer arithmetic on attended time. For rate `r`, elapsed
attended milliseconds `dt`, and prior remainder `q`, `numerator=r*dt+q`, whole steps are
`floor(numerator/3_600_000)`, and the remainder is the exact modulo. Decay moves linearly toward
its declared target and clears its remainder when the target saturates; input remainders are keyed
by `(meter_id,input_index)`. Hook order is decay first, then newly caused ledger facts, then active
contribution inputs; final values clamp to `[0,100]`, then band events emit by meter ID. Offline
mode supplies `dt=0`, so Trust does not rot while the founder is away. Externality and Soul are
not mutated by this hook.

### M5 — Scope & reset

Save v15 replaces `meter_bands` with complete-key `meter_values`, `meter_decay_remainders`, and
`meter_input_remainders` maps. Legacy percent values migrate exactly; missing/extra keys reject
after v15. New-run assembly resolves the new run's pinned meter artifact. Every Standing axis uses
the published clamped Notoriety reseed; every Grievance axis and p(doom) use literal catalog
initials. Founder Soul survives untouched. The assembly function is shared by live Exit and replay.

## Owner rulings on C1–C12 (2026-08-03)

- **C1 — accepted, my Trust shape was wrong:** ten Company Trust IDs
  `trust.{users|employees|regulators|press|investors}.{standing|grievance}`, each unsigned
  independent value; `trust.public.*` rejected; the full ten-row set required. (design/02 §7 is
  binding; my five-single-bar `public` model is withdrawn.)
- **C2 — accepted, Externality is a LEDGER not a meter (I wrongly restored it):** removed from
  the meter catalog/save/decay/band entirely; the `externality.*` ledger-fact namespace is
  registered as an INPUT source only; World Layer owns addressed Externality records. AC3 →
  "no meter decay mutates an `externality.*` ledger fact."
- **C3 — accepted, the cross-scope ruling:** `FounderState.Soul` (existing int64) stays the sole
  Soul authority — no duplication into meter maps. Phase-A meter inputs may OBSERVE Soul (Pet
  Care reads it through the carry seam) but **an ordinary Company transaction cannot mutate a
  Founder-scope meter** (ApplyLogged commits the Company stream only). Soul-changing verbs need a
  successor owning a two-stream write OR a Company `pending_soul_delta` settled atomically at Exit
  (the exit transaction is already multi-stream). **This RFC removes Soul drain/recover** (and
  AC4's Soul half); Soul is read-only here. This makes Meters Foundation Company-scope-only in
  practice — cleaner.
- **C4 — accepted:** `doom.probability` is one Company-scope pressure meter; World Layer may
  publish a separate aggregate, never reusing the field.
- **C5 — accepted:** meter values stay in the implemented `[0,100]` integer domain (NOT ppm — my
  ppm was gratuitous and would have forced a route-context-version migration); the field is
  renamed to store numeric meter values, not band IDs, with the save migration named.
- **C6 — accepted, scope-corrected:** the input union ships ONLY the arms whose mechanics exist
  at Phase A — production-side-effect and ledger-fact inputs (Events L1's event-effect arm is
  added WHEN Events L1 lands, not inferred now). Each arm is a typed row with target/sign/formula/
  source-uniqueness, causal like the Purchasable-Foundation activation rule.
- **C7 — accepted:** decay is fixed-grid partition-invariant on the attended clock, exactly the
  provision-grid pattern (aligned absolute buckets, `advance(a+b)==advance(advance(a),b)`).
- **C8 — accepted:** band events are presentation/trigger output only, never moral acts; one prior
  → final event per meter, with numeric state as the only persisted authority.
- **C9 — accepted:** the catalog owns complete initial values and the published Notoriety reseed;
  new-run assembly resolves the new run's pinned artifact and is shared by live/replay.
- **C10 — accepted:** ablation removes only causally masked inputs; independent decay remains. The
  Trust↔Soul correlation gate is registered as content-blocked Relevance work, not falsely claimed
  green before Soul-changing content exists.
- **C11 — accepted:** meters is a strict schema-v1 epoch artifact added by a protocol-compliant
  mint; replay bundles carry immutable bytes; hook order appends Meters after Purchasable Content;
  formulas and the sequential Go/TS corpus grow in the same semantic landing.
- **C12 — accepted:** not-spendable is structural registry/package separation, not a duplicate
  spendable flag or handwritten spendable set.

## Acceptance criteria

1. Registry round-trip and exact eleven-row set; composition rejects meter/resource ID collision
   and meter cost targets; package-boundary fixture prevents ledger-spend imports.
2. Both declared input arms and decay are byte-identical Go/TS inside `ApplyLogged`, partition-
   invariant, saturating, and offline-stable. Ablation removes causal masked inputs only.
3. Up/down/multi-band transitions emit one correctly ordered event and no moral fact; unchanged
   bands emit nothing. Externality facts and Founder Soul remain byte-untouched.
4. Save-v15 migration has complete-key corpus coverage; live/replay new-run assembly initializes
   all eleven values from the new pinned artifact and preserves Soul.
5. Meter bytes join epoch identity/replay bundles; formula artifacts publish decay/input order;
   kernel version and shared sequential corpus cover input, decay, reset, and offline cases.
6. Relevance registers the Trust↔Soul correlation report as explicitly content-blocked until real
   Soul-changing content and an owner-ruled bound exist.

## Open questions

- Exact Phase-A band IDs/floors, Trust grievance initials, p(doom) initial, decay targets/rates,
  and seed input bindings remain literal balance data required before the epoch mint. The engine,
  schema, save, replay, and fixture work may land first against discriminating test catalogs.

### C13 — Save v15 cannot activate before the meter artifact exists

Implementation reached a dependency the plan's “mint-only” content blocker understated. V15 maps
must have the complete meter-catalog key set, reset/migration needs the pinned artifact, and replay
must identify those bytes. Every current run is pinned to an epoch with no `meters` artifact;
upgrading its next ordinary write to v15 would either reject an incomplete map or initialize axes
from an unpinned/current catalog. The legacy `meter_bands` placeholder may contain zero, two, or
arbitrary mechanically valid keys, and this RFC never defines how absent axes migrate. Filling them
with guessed initials would invent the owner-gated balance rows and make old-run replay depend on
deploy timing.

**Proposed contract:** meter activation is new-run-bound, like other immutable run identity.
Pre-meter runs remain save v14 and execute no meter hook through their terminal transition. The
first run whose pinned epoch contains `meters` is assembled directly as v15 from that artifact:
all Standing axes use the published Notoriety reseed; Grievance and p(doom) use catalog initials;
the legacy `meter_bands` placeholder is deliberately not imported into authoritative v15 state.
Thereafter v15 complete-key validation is mandatory. The migration corpus proves (a) a pre-meter
v14 run remains byte/replay-stable through ordinary intents, (b) Exit starts a complete v15 run
under the new hash, and (c) no v15 state can load without its exact pinned meter artifact. If
legacy placeholder values must instead survive, owner must define an immutable artifact source for
every old constants hash and the missing-axis rule before implementation.

## Owner ruling on the activation boundary (Meters C13, 2026-08-04) — ACCEPTED, generalized

Accepted as proposed, and lifted to a **reusable law** because every save-version foundation
hits it: **new-mechanic activation is NEW-RUN-BOUND at the first epoch whose PINNED catalog
carries the mechanic's artifact.** A run pinned to a pre-artifact epoch finishes under its pinned
save semantics and executes NO hook for the new mechanic (no retroactive gain — the same shape as
L2b version-drift and P6a pre-timer runs). The first Exit into an epoch containing the artifact(s)
assembles the new save version's complete state IN THE NEW-RUN TRANSACTION from that pinned
artifact (Standing axes from the Notoriety reseed, everything else from catalog initials);
subsequent ordinary writes then require the pinned artifact. Replay reads the run's PINNED
artifact bytes, never deploy-current — so activation never depends on deploy timing.

**Meters v15 and Achievements v16 activate ATOMICALLY** at the first Exit into an epoch containing
BOTH `meters` and `achievements` artifacts — one new-run transaction assembles v15 meter state +
empty/derived v16 achievement state together; no run is ever v15-with-meters-but-v14-achievements.
The migration corpus proves: old-run replay through Exit, atomic v14→v16 new-run assembly,
derived-score closure, and no retroactive earning.

This law is now the template for Pet Care and every subsequent save-version mechanic.

## Acceptance blockers (Codex review, 2026-08-03)

The existing save seam is real, but the draft conflicts with binding design and does not define
an implementable transition. The following proposed closures keep the foundation bounded while
making every persisted value and replay effect exact.

### C1 — The Trust model is the superseded single-bar shape

M1 names five Trust meters and substitutes `public` for the design's `press`. Binding
`design/02 §7` requires five constituencies—users, employees, regulators, press, investors—each
with independent Standing and Grievance bars. The current Exit code seeds only regulator Standing
and Grievance, which is an implemented placeholder, not authority for dropping the other eight.

**Proposed contract:** the registry contains exactly ten Company Trust IDs at Phase A:
`trust.{users|employees|regulators|press|investors}.{standing|grievance}`. Standing and Grievance
are unsigned independent PPM values, never complements except where an explicit transition changes
both. Catalog validation requires the full ten-row set and rejects `trust.public.*` and any
constituency missing either axis.

### C2 — Externality is a ledger, not a meter

The draft reintroduces Externality as a Company meter with decay rules. `design/02 §7` explicitly
demotes it to dated, addressed ledger facts rendered on the world map; that distinction is
ruled load-bearing. A scalar would erase where the harm landed and recreate the threshold-
optimization problem the redesign removed.

**Proposed contract:** remove Externality from the meter catalog, save values, decay, and band
events. This RFC may register/validate the existing `externality.*` ledger-fact namespace as an
input source for other meters, but World Layer owns addressed Externality records and aggregation.
AC3 becomes “no meter decay mutates an `externality.*` ledger fact.”

### C3 — Soul already has an incompatible Founder representation

Soul is an existing Founder `int64` field allowing the exact-safe signed range, while M1 describes
all meters as bounded PPM and M5 proposes adding per-meter fields. The RFC does not say whether
Soul migrates, aliases, or duplicates that field. More importantly, ordinary production intents
commit the Company stream only: `ApplyLogged` receives Founder carry as read-only input and cannot
persist a Founder Soul mutation atomically.

**Proposed contract:** retain `FounderState.Soul` as the sole Soul authority and define its closed
range/initial value explicitly; do not duplicate it in Company meter maps. Phase A meter inputs
may **observe** Soul but cannot mutate it during an ordinary Company transaction. Soul-changing
verbs require a successor owning a two-stream transaction (or an explicit Company pending-delta
settled atomically at Exit). Until that owner lands, remove Soul drain/recover from this RFC and
AC4. Pet Care may consume the read-only Soul value through the existing carry seam.

### C4 — p(doom) scope cannot remain open at acceptance

Its scope controls reset, save placement, event ownership, and future World aggregation. A loader
cannot accept an unresolved scope.

**Proposed contract:** Phase A `doom.probability` is one Company-scope pressure meter. World Layer
may publish a separately named aggregate derived from committed Company samples; it never reuses
the Company ID. Initial value and reset are catalog literals. This is a pressure meter, not part
of the moral correlation gate.

### C5 — PPM conflicts with the implemented 0–100 predicate boundary

`MeterBands` currently stores integers in `[0,100]`; route predicates and both Go/TS/schema
validators use inclusive percent values. M1 silently changes values to PPM without naming the
save migration, route context version, or catalog migration. The field is also misnamed: it stores
numeric meter values, not band IDs.

**Proposed contract:** add save v15 fields
`meter_values_ppm: {company_meter_id:uint32}` and
`meter_decay_remainders: {company_meter_id:int64}` with complete catalog key sets. Migrate existing
`meter_bands` percent values by exact `value*10_000`, then remove the old field at v15. Route
context v3 changes `meter_band min/max` to PPM and migrates every active route literal in the same
mint; v2 remains replay-loadable through its pinned catalogs. Go, TS, JSON Schema, replay fixtures,
and route vectors move together. If percent is preferred, reject PPM in M1 instead—one unit system
must govern all boundaries.

### C6 — The input union has no executable arms

“Production side-effects, ledger facts, event effects, choice consequences” gives no row shape,
target, sign, formula, ordering, source uniqueness, or causal evidence. Events L1 is unimplemented,
yet the queue declares Meters upstream and self-contained. Allowing arbitrary event deltas now
would invent the future effect vocabulary.

**Proposed contract:** Phase A ships a closed two-arm catalog union only:
`{kind:"ledger_fact",fact_kind,delta_ppm}` and
`{kind:"contribution_slot",slot,source_id,delta_ppm_per_attended_hour}`. Fact inputs apply once on
the newly emitted fact in byte-order by `(meter_id,input_index)`; contribution inputs integrate
only while the named committed contribution is non-neutral. Duplicate source bindings reject.
Events L1 adds its own arm by successor RFC. The catalog supplies literal bindings; code contains
no flavor IDs or balance deltas.

### C7 — Decay is not partition-invariant as specified

“Rate per attended hour toward neutral” could mean proportional decay, a constant step, or an
evaluation-relative floor. Without a formula/remainder rule, splitting the same interval across
intents can change the result—the exact bug class Purchasable Content C5 eliminated.

**Proposed contract:** decay is linear and exact. For elapsed attended milliseconds `dt`, direction
`sign(toward-value)`, numerator is `rate_ppm_per_attended_hour*dt + remainder`; step is
`floor(numerator/3_600_000)`, remainder is modulo `3_600_000`; clamp at `toward_ppm`, and clear the
remainder on reaching/crossing target. Inputs apply before decay or after decay by one declared
hook order (recommend decay first, then new consequences). Offline mode contributes `dt=0`.
Partition-invariance, saturation, direction reversal, and exact-hour vectors are mandatory in Go
and TS.

### C8 — Band transitions and moral facts are conflated

The draft says entering a forbidden band emits a ledger fact used by Ethical%. Binding design says
meters modulate while dated **acts** gate; a score threshold must never become “20 more evil
points.” It also does not specify whether a multi-band jump emits one event or every crossed band,
or how persisted band state avoids drifting after a catalog retune.

**Proposed contract:** band changes are presentation/trigger events only and never write moral
ledger facts. Ethical% continues using `darkpattern.*`/`externality.*` act facts. The numeric value
is sole persisted authority; current band is derived from the run-pinned catalog at evaluation and
snapshot time—no duplicated band ID in save state. A jump emits exactly one
`meter_band_changed.v1` from prior derived band to final derived band, with
`{run_id,meter_id,from_band,to_band,direction,value_before_ppm,value_after_ppm}`; unchanged bands
emit nothing. Events order by meter ID after the complete meter transition and before later hook
families.

### C9 — The reset contract is incomplete

The current Exit path seeds two regulator values only. M5 does not define the other eight Trust
values, p(doom), migration defaults, or how a catalog retune affects a run pinned to an older
bundle. “From the moral reseed” is not enough to build D6 assembly.

**Proposed contract:** the meter catalog declares `initial_ppm` for every Company pressure meter
and `trust_reseed:{base_ppm,notoriety_numerator,notoriety_denominator,floor_ppm,ceiling_ppm}` for
Trust Standing. Every Standing axis starts at the published clamped result; each Grievance axis
starts from its own catalog literal (not implicitly `1-standing` unless explicitly ruled). p(doom)
uses `initial_ppm`. Exit resolves the **new run's** pinned catalog and initializes all keys in one
assembly function shared by live and replay. Save migration uses the catalog pinned to that run;
missing catalog keys fail closed.

### C10 — Ablation and correlation acceptance are misstated or missing

An ablation mask removes selected purchasable effects; it does not stop time decay or unrelated
ledger-fact inputs. “A masked run leaves meters untouched” is therefore false. Conversely, the
binding Trust↔Soul decorrelation CI law is absent, even though this RFC is its natural owner.

**Proposed contract:** AC2 compares each masked simulation to a causal counterfactual and requires
only inputs sourced by masked effects to disappear; independent decay remains identical. Register
a deferred correlation report in Relevance Harness now: once T0–T1 supplies real Soul-changing
content, its baseline must report Trust Standing and Soul correlation across personas and fail at
an owner-ruled bound. Until then, a fixture proves the report plumbing and the docs state the gate
is content-blocked rather than claiming coverage.

### C11 — Catalog identity and replay-hook order are absent

Meters add a new catalog family consumed inside `ApplyLogged`, but the RFC does not register it
with the epoch seed/replay bundle, define schema version, or place it in the already pinned hook
order. A current-catalog read would make historical verification deployment-dependent.

**Proposed contract:** `meters` is a strict schema-v1 epoch artifact and adding it is a
protocol-compliant mint. Replay bundles carry its immutable bytes and both kernels validate them.
The declared hook order grows from Prestige → faction → Guild → Commons → Purchasable Content to
→ **Meters** (with C7 defining internal decay/input order); unregistered hooks remain forbidden.
Kernel version bumps with semantic behavior, the sequential cross-runtime corpus includes upward,
downward, decay, reset, and offline cases, and formula artifacts publish every decay/input formula.

### C12 — Not-spendable enforcement needs one authority

Meters are not economy resources today, so existing cost validation already rejects them by
absence. A second handwritten “meter spendable set” can drift, while `spendable:false` on every
row is a constant masquerading as data.

**Proposed contract:** remove `spendable` from rows and make the type boundary structural: meter
IDs and economy resource IDs are disjoint registries, validated jointly at catalog composition;
cost/effect targets accept only economy resource IDs unless a successor adds a declared meter
effect arm. Go package imports forbid `meters` from economy ledger spend APIs. Seeded composition
tests prove ID collision and a meter cost both reject. Docs phrase the guarantee narrowly: no
intent can use a meter as ledger payment; declared meter-transition inputs may decrease values.

## Changelog

- 2026-08-03: created (draft) — the moral-axis mechanics; not-spendable enforced in code.
- 2026-08-03: Codex acceptance review recorded C1–C12; all proposed closures were owner-approved.
  The normative body was reconciled to the two-axis Trust model, addressed Externality ledger,
  read-only Founder Soul seam, percent value domain, causal input union, deterministic decay,
  band-event boundary, epoch/replay ownership, and structural not-spendable rule. Literal Phase-A
  balance rows remain a named pre-mint content gap; implementation is otherwise unblocked.
- 2026-08-04: C13 records the save-v15 activation gap: pre-meter runs have no pinned meter artifact
  and cannot be upgraded deterministically. Proposed new-run-only activation awaits owner ruling.
- 2026-08-06: non-normative reference cleanup for publication; no spec change.
