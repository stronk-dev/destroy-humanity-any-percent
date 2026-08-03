# RFC: Meters Foundation (Trust · Externality · Soul · p(doom))

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-03
- **Design refs:** `design/02 §7` (moral axis — not spendable; Trust 5 constituencies + Externality ledger), `design/02 §8` (Soul, the personal ledger), `design/09 §2` (pressure meters, EU4 disaster model), `design/10` (Ethical% consumes the moral stack)
- **Research:** `events-playstyles.md §1` (pressure-meter architecture — visible forecastable bars driven by the player's own choices), `morality-systems.md`, `billionaires-decay.md` (Trust as a resource that only spends, never buys)
- **Depends on:** Production + Run Genesis (implemented — meters mutate inside `ApplyLogged`); Events L1 (drafted — Layer-2 meters ARE recurring hidden events on the same evaluator)
- **Owner ruling honored:** breadth-first foundation — the meter MECHANICS, not the events or endings that consume them.
- **Planning:** `planning/meters-foundation/` (once implementing)

## Summary

`MeterBands` exists as a save-state map with no mechanics. This RFC gives it laws: the closed
meter registry, the not-spendable invariant enforced in code, band transitions as evented facts,
and the two scopes (Company-run moral axis, Founder-persistent Soul). The moral axis is the
structural spine of Ethical% and every ending — it must be a real quantity before content leans
on it.

## Specification

### M1 — The meter registry (closed catalog family)

`balance/meters/*.json`: `{id, scope: company|founder, min_ppm, max_ppm, bands: [{id, floor_ppm}],
inputs: [closed contribution union], decay: {toward_ppm, rate_ppm_per_attended_hour}|null,
spendable: false}`. Phase-0 meters: the five **Trust** constituencies (users, employees,
regulators, investors, public — Company scope, `design/02 §7`), the **Externality** ledger
(Company, accumulates — never decays, the point), **Soul** (Founder scope, persists across Exit,
the §8 personal ledger), **p(doom)** (Company pressure meter). Values are ppm integers; bands are
the legible forecastable state (EU4 model).

### M2 — The not-spendable invariant (enforced, not documented)

**No intent may debit a meter as payment.** Meters move ONLY via declared `inputs` — production
stack side-effects, ledger-fact emissions, event-option effects, choice consequences — computed
inside `ApplyLogged` from committed state. The loader rejects any catalog wiring a meter as a
`cost` resource; a lint/loader assertion forbids meter IDs in the ledger's spendable set. This is
the moral axis's defining property: you can act to change your Trust, never *pay* with it
(`billionaires-decay.md`'s Institutional Trust — "can't be bought, only spent" — made
structural).

### M3 — Band transitions as facts

Crossing a band floor emits `meter_band_changed {meter_id, from_band, to_band, direction}` (event
kind, registry-registered) — the hook for pressure-meter events (Layer 2: a meter entering its
`crisis` band IS the trigger of a hidden recurring event, per Events L1's E3 note), for
Ethical%'s `facts_disjoint` predicate (a forbidden band entered = a ledger fact), and for UI
dials. Band state is save-persisted (`MeterBands`, already present — this RFC populates it with
meaning). All transitions replay-deterministic (inputs are state-derived; nothing new in
replay_inputs).

### M4 — Decay & the attended clock

Trust constituencies decay toward a neutral band on attended time (the reseed formula
`clamp(90−0.35·Notoriety,55,90)` already ruled for the Founder-side reseed — this is the live
Company decay); Externality NEVER decays (accumulation is the satire); Soul decays only via
declared drains (crunch, Faustian contracts — `design/08` beats), recovers only via
touch-grass activities (design/02 §8), both as declared inputs. Decay uses the same
attended/offline split as everything else — meters move on attended time; an idle founder's Trust
doesn't rot while they're away.

### M5 — Scope & reset

Company-scope meters reset with the run (D6 assembly); the moral reseed (`10 §5`) sets the new
run's starting Trust from Founder Notoriety — already implemented in the Exit transaction; this
RFC declares the meter-init side. Soul persists (Founder scope). Save-version bump adds the
per-meter ppm fields + remainders; corpus fixtures both scopes.

## Acceptance criteria

1. Registry round-trip; loader rejects a meter wired as a cost resource (the not-spendable law,
   seeded-violation test).
2. Meter mutation through a declared input is byte-parity Go/TS inside `ApplyLogged`; a masked/
   ablated run leaves meters untouched (they're state-derived, not input).
3. Band transitions emit the event with correct direction; a fixture crosses up and down;
   Externality accumulation never decreases across a decay interval.
4. Decay: Trust decays toward neutral on attended time only (offline-span fixture proves no
   rot-while-away); Soul drain/recover through declared inputs.
5. Reset: Company meters reinit from the moral reseed at Exit; Soul survives; save migration +
   corpus both scopes.

## Open questions

- Exact Phase-0 band counts and floors per meter — balance data, harness-gated.
- Whether p(doom) is Company or World scope (recommend Company for Phase 0; World-aggregate
  p(doom) is a World Layer foundation concern).

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
demotes it to dated, addressed ledger facts rendered on the world map; the research calls that
distinction load-bearing. A scalar would erase where the harm landed and recreate the threshold-
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
- 2026-08-03: Codex acceptance review recorded C1–C12. Implementation is blocked pending owner
  rulings and reconciliation with the two-axis Trust model, Externality ledger, and Founder Soul
  transaction boundary.
