# First Content Epoch — draft production content rows (Meters · Achievements · Pet Care)

> **FROZEN HISTORICAL PROPOSAL — NONCANONICAL.** Preserved for the 2026-08-07 mint
> provenance. Its duplicated payloads are not production authority. Canonical adopted bytes live
> under `balance/`; current rulings and status live in `planning/platform-alignment/`.

**Status: DRAFT FOR OWNER REVIEW — nothing committed, nothing in the repo changed.**

Prepared for `rfc/first-content-epoch.md` (epoch 6). These are proposed production bytes for the
three foundations whose artifacts are fixture-only and whose final AC *is* the mint (FCE5.1:
Meters, Achievements, Pet Care). Per FCE2 the default is byte-identical promotion of reviewed
fixture bytes; where this proposal diverges from a fixture literal, the divergence is called out
explicitly with its design grounding, so the owner can accept or revert each one.

**Validation performed (2026-08-07):** all three draft artifacts below were loaded through the
*actual Go loaders* from a scratchpad harness (repo untouched):

- `meters.LoadCatalog` — OK (11 meters), plus `ValidateResourceSeparation` against the production
  economy catalog (`balance/catalogs/phase0.json`) — OK.
- `achievements.LoadCatalog` with `production.FoundationAchievementRegistry(economy phase0)` — OK
  (12 rows) *after* adding the 13 proposed copy keys to the registry; correctly **rejected**
  against the current generated `copykeys.All()` (proves the copy-pipeline entries must land
  before the artifact can load — FCE5.4 ordering).
- `pet.LoadCatalog` — OK (schema 2, 5 actions, `SchemaSupportsSoul() == true`, which
  `server/production/replay.go:174` requires because epoch 6 pins the Soul artifact).
- `balance/meters.schema.json` and `balance/achievements.schema.json` re-checked with ajv — VALID.
  (There is no `balance/pets.schema.json`; the pet artifact is validated by the Go/TS loaders
  only — see cross-cutting notes.)

Proposed production paths (FCE2 repo convention, implementer-chosen):

| Artifact | Path |
|---|---|
| meters | `balance/meters/first-content.json` |
| achievements | `balance/achievements/first-content.json` |
| pets | `balance/pets/first-content.json` |

---

## 1. Meters (`balance/meters/first-content.json`)

### Loader constraints that shaped the rows (from `server/meters/catalog.go`)

- Exactly **11 rows in a fixed order**: `doom.probability`, then
  `trust.{employees,investors,press,regulators,users}.{grievance,standing}` alphabetically —
  `raw.Meters[i].ID != requiredMeterIDs[i]` is a hard load failure, so row order below is
  mandatory, not stylistic.
- Exact keys per row `{id,scope:"company",min_value:0,max_value:100,initial_value,bands,inputs,decay}`;
  unknown fields rejected (`DisallowUnknownFields`); `inputs` must be present even when empty
  (`[]`, not omitted); `decay` must be present (`null` allowed, object here).
- Bands: unique IDs, strictly ascending floors, first floor must be 0.
- Inputs are a closed two-arm union with **exact-key checks**:
  `{kind:"ledger_fact",fact_kind,delta}` (delta ∈ [-100,100], ≠0) and
  `{kind:"contribution_slot",slot,source_id,delta_per_attended_hour}` (slot must be a valid
  `multiplier.Slot`: upgrades|milestones|faction|doctrine|commons|trust|event_buffs|prestige).
  Duplicate source bindings per meter reject. **No cross-reference check** against economy —
  a nonexistent `source_id` loads but is permanently inert.
- `trust_reseed`: all five fields required; `floor ≤ base ≤ ceiling`, denominator ≥ 1.
- Decay: `toward_value ∈ [0,100]`, `rate_per_attended_hour ∈ [1,100]`.
- Composition (`replaycatalog.Load` → `ValidateResourceSeparation`): meter IDs must be disjoint
  from economy resource IDs — verified against phase0 (`company.cash` only).
- Standing-axis `initial_value`s are **dead values**: `meters.NewRunState` overwrites every
  `trust.*.standing` with the clamped Notoriety reseed at new-run assembly. They are kept at the
  fixture's 50 for byte continuity; only the reseed block governs real runs.

### Draft artifact

```json
{
  "schema_version": 1,
  "trust_reseed": {
    "base_value": 90,
    "notoriety_numerator": 35,
    "notoriety_denominator": 100,
    "floor_value": 55,
    "ceiling_value": 90
  },
  "meters": [
    {
      "id": "doom.probability",
      "scope": "company",
      "min_value": 0,
      "max_value": 100,
      "initial_value": 50,
      "bands": [
        { "id": "low", "floor_value": 0 },
        { "id": "high", "floor_value": 70 }
      ],
      "inputs": [
        { "kind": "ledger_fact", "fact_kind": "externality.emitted", "delta": 3 }
      ],
      "decay": { "toward_value": 50, "rate_per_attended_hour": 2 }
    },
    {
      "id": "trust.employees.grievance",
      "scope": "company",
      "min_value": 0,
      "max_value": 100,
      "initial_value": 0,
      "bands": [
        { "id": "low", "floor_value": 0 },
        { "id": "high", "floor_value": 70 }
      ],
      "inputs": [],
      "decay": { "toward_value": 0, "rate_per_attended_hour": 1 }
    },
    {
      "id": "trust.employees.standing",
      "scope": "company",
      "min_value": 0,
      "max_value": 100,
      "initial_value": 50,
      "bands": [
        { "id": "low", "floor_value": 0 },
        { "id": "high", "floor_value": 70 }
      ],
      "inputs": [],
      "decay": { "toward_value": 50, "rate_per_attended_hour": 2 }
    },
    {
      "id": "trust.investors.grievance",
      "scope": "company",
      "min_value": 0,
      "max_value": 100,
      "initial_value": 0,
      "bands": [
        { "id": "low", "floor_value": 0 },
        { "id": "high", "floor_value": 70 }
      ],
      "inputs": [],
      "decay": { "toward_value": 0, "rate_per_attended_hour": 1 }
    },
    {
      "id": "trust.investors.standing",
      "scope": "company",
      "min_value": 0,
      "max_value": 100,
      "initial_value": 50,
      "bands": [
        { "id": "low", "floor_value": 0 },
        { "id": "high", "floor_value": 70 }
      ],
      "inputs": [],
      "decay": { "toward_value": 50, "rate_per_attended_hour": 2 }
    },
    {
      "id": "trust.press.grievance",
      "scope": "company",
      "min_value": 0,
      "max_value": 100,
      "initial_value": 0,
      "bands": [
        { "id": "low", "floor_value": 0 },
        { "id": "high", "floor_value": 70 }
      ],
      "inputs": [],
      "decay": { "toward_value": 0, "rate_per_attended_hour": 1 }
    },
    {
      "id": "trust.press.standing",
      "scope": "company",
      "min_value": 0,
      "max_value": 100,
      "initial_value": 50,
      "bands": [
        { "id": "low", "floor_value": 0 },
        { "id": "high", "floor_value": 70 }
      ],
      "inputs": [],
      "decay": { "toward_value": 50, "rate_per_attended_hour": 2 }
    },
    {
      "id": "trust.regulators.grievance",
      "scope": "company",
      "min_value": 0,
      "max_value": 100,
      "initial_value": 0,
      "bands": [
        { "id": "low", "floor_value": 0 },
        { "id": "high", "floor_value": 70 }
      ],
      "inputs": [],
      "decay": { "toward_value": 0, "rate_per_attended_hour": 1 }
    },
    {
      "id": "trust.regulators.standing",
      "scope": "company",
      "min_value": 0,
      "max_value": 100,
      "initial_value": 50,
      "bands": [
        { "id": "low", "floor_value": 0 },
        { "id": "high", "floor_value": 70 }
      ],
      "inputs": [],
      "decay": { "toward_value": 50, "rate_per_attended_hour": 2 }
    },
    {
      "id": "trust.users.grievance",
      "scope": "company",
      "min_value": 0,
      "max_value": 100,
      "initial_value": 0,
      "bands": [
        { "id": "low", "floor_value": 0 },
        { "id": "high", "floor_value": 70 }
      ],
      "inputs": [],
      "decay": { "toward_value": 0, "rate_per_attended_hour": 1 }
    },
    {
      "id": "trust.users.standing",
      "scope": "company",
      "min_value": 0,
      "max_value": 100,
      "initial_value": 50,
      "bands": [
        { "id": "low", "floor_value": 0 },
        { "id": "high", "floor_value": 70 }
      ],
      "inputs": [
        { "kind": "contribution_slot", "slot": "commons", "source_id": "commons.member", "delta_per_attended_hour": 1 }
      ],
      "decay": { "toward_value": 50, "rate_per_attended_hour": 2 }
    }
  ]
}
```

### Provenance annotations

| Value | Provenance |
|---|---|
| `trust_reseed` = base 90, 35/100, clamp [55,90] | **(design/10-playstyles.md §"the game remembers")** — published formula `clamp(90 − 0.35·Notoriety, 55, 90)`, exactly; also the reviewed fixture bytes (`balance/testdata/meters-catalog-parity-v1.json`, `testdata/replay/apply-logged-v1.json`). Keep verbatim. |
| Band set `low@0 / high@70` on every meter | (fixture — `meters-catalog-parity-v1.json`; **default — no design source, flag**. Design names no band IDs/floors. `high@70` is loosely consistent with routes phase0's `trust.regulators.standing min 70` predicate literal, but route predicates compare numeric values, not bands, so bands are presentation/trigger-only.) |
| `doom.probability` initial 50 | (fixture; **default — no design source, flag**. design/02 §7 and design/09 §3 define p(doom) direction/feeders but no initial.) |
| `doom.probability` decay toward 50 @ 2/attended-h | (fixture; **default — no design source, flag**) |
| `doom.probability` input `externality.emitted +3` | (fixture, kept; design/02 §7: Externality "feeds the Planet drain and conspiracy/outrage pressure meters". **Inert at epoch 6** — no live intent emits any `externality.*` Company ledger fact yet; see DESIGN-GAP M1.) |
| Fixture's second doom input (`contribution_slot upgrades/generator.example -2/h`) **dropped** | **Deliberate divergence from fixture.** `generator.example` does not exist in the production economy catalog (phase0 has only `generator.beige_tower`); the loader would accept it but it can never activate. Design's "lowered by safety spending" (design/02 §7) has no bindable source at epoch 6 — see DESIGN-GAP M2. |
| Standing initial 50 (all five) | (fixture, kept; **dead value** — reseed governs; harmless byte continuity) |
| Standing decay toward 50 @ 2/attended-h | (fixture; **default — no design source, flag**. "Trust decays toward neutral without maintenance" is a defensible reading of the moral-axis intent, but no design number exists.) |
| Grievance initial 0, decay toward 0 @ 1/attended-h | **Deliberate divergence from fixture** (fixture: initial 50, decay toward 50). Grounding: design/02 §7's derivation rule — "every moral quantity is derived from the production stack, never awarded" — and design/09 §3's quiet-days-reduce-pressure shape. The fixture's decay-toward-50 would *spontaneously regenerate grievance* toward 50 in any run where inputs pushed it below 50, i.e. the meter would manufacture grievance no act caused. A fresh 1995 garage has no aggrieved constituencies: initial 0. Rate 1 (slower than standing's 2) so accrued grievance fades over ~days of attended play. **Numbers are provisional (flag); the direction (toward 0, start 0) is the design-grounded part. Owner may instead rule byte-identical fixture promotion per FCE2 — this is the one meters row-family where I recommend not doing that.** |
| `trust.users.standing` input `commons.member +1/attended-h` | (design/02 §7: Ethical% is "viable mainly via the MMO commons" — commons membership is the one live, design-named Trust feeder that exists at epoch 6; `commons.member` is a real multiplier source in `balance/catalogs/phase0.json`, slot `commons`. The +1/h magnitude is **default — flag**.) |

### Why the meter set is otherwise input-empty

The design's named Trust/doom feeders (moderation cuts, scandals, dark-pattern stages, crunch,
e/acc tree, safety spending — design/09 §3 table) all require content (upgrades, events, ledger
facts) that does not exist in any epoch-6 artifact. Binding placeholder sources would violate the
no-improvised-mechanics law; the two inputs above are the only ones with either reviewed-fixture
provenance (`externality.emitted`) or a live design-named source (`commons.member`). Everything
else is epoch-7+ retune-lane work as real feeders land.

---

## 2. Achievements (`balance/achievements/first-content.json`)

### Loader constraints that shaped the rows (from `server/achievements/catalog.go` + `server/production/foundations.go`)

- Rows must be **strictly byte-ascending by ID** (`source.ID <= lastID` rejects).
- `copy_key` (and possession `justification_copy_key`) must be present in the registry's copy
  keys, which in production is exactly `copykeys.All()` (generated from `copy/catalog/*.json` by
  `make copy-generate`). **The copy entries must land before this artifact can load** — verified:
  the draft loads with the proposed keys added and is rejected against today's generated set.
- The production registry (`FoundationAchievementRegistry`) is *narrow*:
  - run counters: `generators_purchased_total`, `tier` only;
  - career counters: `age_ms`, `notoriety` only;
  - generator IDs / resource IDs: from the pinned economy artifact — phase0 gives exactly
    `generator.beige_tower` and `company.cash`;
  - provenance sources: `counter:run:generators_purchased_total → [generator_purchased]`,
    `counter:run:tier → [gate_crossed]`, `counter:career:age_ms → [founder_advanced]`,
    `counter:career:notoriety → [founder_advanced]`,
    `exit_count → [founder_advanced, run_ended]` — and **no `fact:*` entries at all**, so any
    `fact_present` condition with a provenance proof is *unloadable in production* (this is why
    the parity fixture's `fact_present gate.tier_1` row was NOT promoted; the apply-logged replay
    fixture's `counter tier ≥ 1` form was used instead).
- Provenance proofs: `event_kinds` sorted strictly ascending, all in `save.AllEventKinds`, and
  must cover every provenance-source key derived from the condition tree; forbidden when the
  condition contains `owns_generator_at_least`.
- `possession` proofs: run-scope only, require an `owns_generator_at_least` condition and a
  registered justification copy key. `burn` proofs: run-scope only, canonical positive decimal
  minimum; satisfied only when the earning transition's same batch contains the declared event
  *and* an action debit of ≥ minimum on the declared resource (`achievementProofSatisfied` +
  `actionDebits` derived post-accrual — gate crossing debits its requirement, so a 1e9 gate burn
  is provable).
- `exit_count_at_least` / career `counter_at_least` are career-scope; `owns_generator_at_least`
  run-scope; `all_of` needs 2–16 children, depth ≤ 4, ≤ 64 nodes.
- Tier mechanics reality check (`server/production/prestige.go tierForGate`): tier advances ONLY
  by crossing a `gate.tN_to_tN+1` gate; routes phase0 ships `gate.t2_to_t3` (1e9 cash),
  `gate.t4_to_t5` (1e15), `gate.t7_to_t8` (1e24). A run therefore jumps tier 0 → 3 at its first
  gate. `counter tier ≥ 1` and `≥ 3` latch at the same moment today; `≥ 1` (the reviewed fixture
  literal) is kept for the first-gate row so it stays correct if tier-1/2 gates are ever minted.

### Draft artifact (12 rows, uniform +4 score — design/02 §6 milk model)

```json
{
  "schema_version": 1,
  "achievements": [
    {
      "id": "achievement.career_attended_day",
      "condition_scope": "career",
      "condition": { "kind": "counter_at_least", "counter": "age_ms", "minimum": 86400000 },
      "proof": { "kind": "provenance", "event_kinds": ["founder_advanced"] },
      "score_grant": 4,
      "copy_key": "achievement.career_attended_day"
    },
    {
      "id": "achievement.career_attended_hour",
      "condition_scope": "career",
      "condition": { "kind": "counter_at_least", "counter": "age_ms", "minimum": 3600000 },
      "proof": { "kind": "provenance", "event_kinds": ["founder_advanced"] },
      "score_grant": 4,
      "copy_key": "achievement.career_attended_hour"
    },
    {
      "id": "achievement.exit_count_5",
      "condition_scope": "career",
      "condition": { "kind": "exit_count_at_least", "count": 5 },
      "proof": { "kind": "provenance", "event_kinds": ["founder_advanced", "run_ended"] },
      "score_grant": 4,
      "copy_key": "achievement.exit_count_5"
    },
    {
      "id": "achievement.first_gate",
      "condition_scope": "run",
      "condition": { "kind": "counter_at_least", "counter": "tier", "minimum": 1 },
      "proof": { "kind": "provenance", "event_kinds": ["gate_crossed"] },
      "score_grant": 4,
      "copy_key": "achievement.first_gate"
    },
    {
      "id": "achievement.gate_burn_t3",
      "condition_scope": "run",
      "condition": { "kind": "counter_at_least", "counter": "tier", "minimum": 3 },
      "proof": { "kind": "burn", "event_kind": "gate_crossed", "resource_id": "company.cash", "minimum": "1e9" },
      "score_grant": 4,
      "copy_key": "achievement.gate_burn_t3"
    },
    {
      "id": "achievement.generators_owned_100",
      "condition_scope": "run",
      "condition": { "kind": "owns_generator_at_least", "generator_id": "generator.beige_tower", "count": 100 },
      "proof": { "kind": "possession", "justification_copy_key": "achievement.possession_warning" },
      "score_grant": 4,
      "copy_key": "achievement.generators_owned_100"
    },
    {
      "id": "achievement.generators_owned_300",
      "condition_scope": "run",
      "condition": { "kind": "owns_generator_at_least", "generator_id": "generator.beige_tower", "count": 300 },
      "proof": { "kind": "possession", "justification_copy_key": "achievement.possession_warning" },
      "score_grant": 4,
      "copy_key": "achievement.generators_owned_300"
    },
    {
      "id": "achievement.generators_purchased_1",
      "condition_scope": "run",
      "condition": { "kind": "counter_at_least", "counter": "generators_purchased_total", "minimum": 1 },
      "proof": { "kind": "provenance", "event_kinds": ["generator_purchased"] },
      "score_grant": 4,
      "copy_key": "achievement.generators_purchased_1"
    },
    {
      "id": "achievement.generators_purchased_25",
      "condition_scope": "run",
      "condition": { "kind": "counter_at_least", "counter": "generators_purchased_total", "minimum": 25 },
      "proof": { "kind": "provenance", "event_kinds": ["generator_purchased"] },
      "score_grant": 4,
      "copy_key": "achievement.generators_purchased_25"
    },
    {
      "id": "achievement.generators_purchased_25_tier_3",
      "condition_scope": "run",
      "condition": {
        "kind": "all_of",
        "conditions": [
          { "kind": "counter_at_least", "counter": "generators_purchased_total", "minimum": 25 },
          { "kind": "counter_at_least", "counter": "tier", "minimum": 3 }
        ]
      },
      "proof": { "kind": "provenance", "event_kinds": ["gate_crossed", "generator_purchased"] },
      "score_grant": 4,
      "copy_key": "achievement.generators_purchased_25_tier_3"
    },
    {
      "id": "achievement.old_hand",
      "condition_scope": "career",
      "condition": { "kind": "exit_count_at_least", "count": 1 },
      "proof": { "kind": "provenance", "event_kinds": ["founder_advanced", "run_ended"] },
      "score_grant": 4,
      "copy_key": "achievement.old_hand"
    },
    {
      "id": "achievement.tier_5",
      "condition_scope": "run",
      "condition": { "kind": "counter_at_least", "counter": "tier", "minimum": 5 },
      "proof": { "kind": "provenance", "event_kinds": ["gate_crossed"] },
      "score_grant": 4,
      "copy_key": "achievement.tier_5"
    }
  ]
}
```

### Row provenance & reachability

| Row | Provenance | Reachable |
|---|---|---|
| `first_gate` | Reviewed replay-fixture row (`testdata/replay/apply-logged-v1.json`), kept byte-compatible (condition/proof/score identical; production copy key instead of the fixture's placeholder `category.any_percent`). | First gate = `gate.t2_to_t3` @ 1e9 cash (design/02 §11 pacing: day 2–4). |
| `old_hand` | Reviewed replay-fixture row, kept — **score changed 6 → 4** (design/02 §6: "every achievement grants +4" — flat milk-model grant; the fixture's 6 was test variance). | First Exit: 45–90 min (design/02 §11). |
| `generators_owned_300` | design/02 §6 names "own 300 of a generator" verbatim; parity-fixture shape (`generator_hoard`, possession proof) rebound from the fixture-only `generator.clickfarm` to the real `generator.beige_tower`. Score 8 → 4 (same flat-grant rule). | Late (cumulative cost ≈ 1e17, ~T4–T5). |
| `generators_owned_100` | (default — no design source, flag) T0–T1-reachable possession companion (~2e7 cumulative). | T1-adjacent. |
| `generators_purchased_1` | (design/02 §11: "first generator < 60 s") | T0, first minute. |
| `generators_purchased_25` | 25 = the tier-1 progress-coordinate requirement (`balance/catalogs/phase0.json` tier-1 composite `count_fraction required: 25`). | T0–T1. |
| `generators_purchased_25_tier_3` | (default — flag) `all_of` demonstration row combining the two live run counters. | With first gate. |
| `gate_burn_t3` | Burn discipline demo (CA1: "burn — default for prestige achievements"); the 1e9 minimum is exactly the `gate.t2_to_t3` requirement debit in routes phase0, so the proof is honest and provable in-batch. | With first gate. |
| `career_attended_hour` / `career_attended_day` | (default — no design source, flag) career-age ladder on the only live career counter; `age_ms` advances at Exit, so these latch at exits. | Session 1 / day 2+. |
| `exit_count_5` | (default — flag) career ladder over `old_hand`. | Multi-session. |
| `tier_5` | Gate `gate.t4_to_t5` exists in routes phase0 (1e15). | Mid-game (weeks). |
| `score_grant: 4` everywhere | **(design/02 §6)** — "every achievement grants +4"; carried onto `achievement_score` after the C1 Clout split (see DESIGN-GAP A3). |

Deliberately **excluded** (with reasons — these are findings, not oversights):

- Any `fact_present` row — no `fact:*` provenance sources exist in the production registry and no
  live Company intent emits ledger facts today (DESIGN-GAP A1).
- Notoriety rows — no notoriety producer exists; a `notoriety ≥ N` row would be dead content
  (DESIGN-GAP A4).
- Click/manual rows — no manual-action counter is registered (DESIGN-GAP A5).
- Route/exit-type rows (e.g. "took the acquihire") — exit-type facts like `exit.acquihire` are
  fact-namespace, blocked by A1.

---

## 3. Pet Care (`balance/pets/first-content.json`)

### Loader constraints that shaped the rows (from `server/pet/full_catalog.go` + `grammar.go` + `server/production/replay.go`)

- Exact top-level keys `{schema_version, stat_policy, actions, trust_policy, mood_policy,
  behavior_policy}` (C17 wire grammar); exact keys checked per row family; duplicate JSON keys
  rejected.
- **`schema_version: 2` is mandatory for epoch 6**: the bundle pins Soul, and
  `server/production/replay.go:174` requires `Pets.SchemaSupportsSoul()` when Soul is present —
  every action row must then carry a valid `soul_gate` (`essential|recovery|ordinary`).
- Stats: exactly the four fixed IDs, `floor_ppm ≤ initial_ppm ≤ 1_000_000`, decay ∈ [0, 1e6].
- Actions: strictly ascending `action_id` (mechanical-ID pattern), `delta_ppm ∈ [1, 1e6]`
  (**positive, single-stat only**), `min_eligible_ppm ≤ the target stat's floor_ppm` — the
  loader enforces C19's no-death carve-out (every action must remain eligible at the floor).
- Trust: `floor ≤ neutral ≤ initial ≤ cap ≤ 1e6`.
- Mood: all four moods exactly once, floors strictly ascending from 0.
- Behavior: closed state/event enums, `duration_grid_ticks ≥ 1`, unique `(from_state, event)`
  transition pairs (deterministic FSM — C17 removed weighted choice).

### Draft artifact

```json
{
  "schema_version": 2,
  "stat_policy": {
    "grid_ms": 360000,
    "stats": [
      { "stat_id": "hunger", "initial_ppm": 750000, "floor_ppm": 100000, "decay_ppm_per_grid": 2000 },
      { "stat_id": "energy", "initial_ppm": 750000, "floor_ppm": 100000, "decay_ppm_per_grid": 2000 },
      { "stat_id": "cleanliness", "initial_ppm": 750000, "floor_ppm": 100000, "decay_ppm_per_grid": 2000 },
      { "stat_id": "affection", "initial_ppm": 750000, "floor_ppm": 100000, "decay_ppm_per_grid": 2000 }
    ],
    "diminishing_threshold_ppm": 900000,
    "diminishing_factor_ppm": 500000
  },
  "actions": [
    { "action_id": "care.feed", "stat_id": "hunger", "delta_ppm": 400000, "cooldown_attended_ms": 3600000, "min_eligible_ppm": 100000, "soul_gate": "essential" },
    { "action_id": "care.groom", "stat_id": "cleanliness", "delta_ppm": 250000, "cooldown_attended_ms": 3600000, "min_eligible_ppm": 100000, "soul_gate": "essential" },
    { "action_id": "care.pet", "stat_id": "affection", "delta_ppm": 200000, "cooldown_attended_ms": 600000, "min_eligible_ppm": 100000, "soul_gate": "ordinary" },
    { "action_id": "care.play", "stat_id": "affection", "delta_ppm": 250000, "cooldown_attended_ms": 1800000, "min_eligible_ppm": 100000, "soul_gate": "ordinary" },
    { "action_id": "care.rest", "stat_id": "energy", "delta_ppm": 250000, "cooldown_attended_ms": 3600000, "min_eligible_ppm": 100000, "soul_gate": "essential" }
  ],
  "trust_policy": {
    "initial_ppm": 500000,
    "neutral_ppm": 500000,
    "floor_ppm": 0,
    "cap_ppm": 1000000,
    "gain_ppm_per_effective_action": 20000,
    "decay_ppm_per_grid": 42
  },
  "mood_policy": [
    { "mood_member": "withdrawn", "floor_ppm": 0 },
    { "mood_member": "restless", "floor_ppm": 250000 },
    { "mood_member": "neutral", "floor_ppm": 500000 },
    { "mood_member": "engaged", "floor_ppm": 750000 }
  ],
  "behavior_policy": [
    { "from_state": "idle", "event": "care_applied", "to_state": "care_response", "duration_grid_ticks": 1 },
    { "from_state": "idle", "event": "care_rejected", "to_state": "resting", "duration_grid_ticks": 1 },
    { "from_state": "idle", "event": "grid_tick", "to_state": "active", "duration_grid_ticks": 2 },
    { "from_state": "active", "event": "care_applied", "to_state": "care_response", "duration_grid_ticks": 1 },
    { "from_state": "active", "event": "grid_tick", "to_state": "resting", "duration_grid_ticks": 2 },
    { "from_state": "care_response", "event": "grid_tick", "to_state": "idle", "duration_grid_ticks": 1 },
    { "from_state": "resting", "event": "care_applied", "to_state": "care_response", "duration_grid_ticks": 1 },
    { "from_state": "resting", "event": "grid_tick", "to_state": "idle", "duration_grid_ticks": 2 }
  ]
}
```

### Provenance annotations

| Value | Provenance |
|---|---|
| `grid_ms: 360000` (6 min) | (**default — flag**; chosen so design/04's per-hour rates are *exact* integers per grid: 2 pts/h = 2000 ppm/grid with zero rounding, and FSM `grid_tick` cadence is minutes-scale, appropriate for visible ambient behavior. Design specifies rates, never a grid.) |
| stat `initial_ppm: 750000` | **(design/04 §1)** — "Four stats 0–100 (default 75)". |
| stat `decay_ppm_per_grid: 2000` (= 2 pts/attended-hour) | **(design/04 §1)** — adult decay 2/h. Kitten 3/h and elder 1/h have no age axis in the ruled artifact — DESIGN-GAP P4. Decay runs on the *attended* clock per the ruled mechanics (PC1/C4/C10), not design/04's elapsed clock — see P5. |
| stat `floor_ppm: 100000` | (fixture — `testdata/pet/care-transition-v1.json` & `apply-logged-v1.json`; **default — no design source, flag**. Design says stats floor and stay (no-death), never where; 10 % keeps the pet visibly alive rather than at zero.) |
| `diminishing_threshold_ppm: 900000`, `factor 500000` | **(design/04 §1 / RFC PC2)** — ">90 → 0.5×", exactly. **Divergence from fixture** (fixtures use 700000, test data); design contradicts fixture, so design wins. |
| `care.feed +400000` | **(design/04 §1)** — "feed (+40 Fed)"; hunger = Fed. |
| `care.pet +200000` / `care.play +250000` | **(design/04 §1)** — "+20 Happy" / "+25 Happy"; affection = Happy. Their −5/−15 Energy secondary costs are NOT representable — DESIGN-GAP P2. |
| `care.groom +250000` | **(design/04 §1)** — "+25 Clean". |
| `care.rest +250000` | (design/04 names "passive rest" with no number; **default — flag**.) |
| cooldowns (1h feed/groom/rest, 30m play, 10m pet) | (**default — no design source, flag**. Design's only scarcity is diminishing returns; the ruled grammar requires a cooldown per action (0 allowed). Values sized so care is a check-in rhythm, not a click-spam loop.) |
| `soul_gate` assignment (feed/groom/rest essential; pet/play ordinary) | (interpretive from **design/02 §8** — "the pet stops recognizing you… UI literally greys out": welfare actions stay available at near-zero Soul (no-death), affection interactions are what the drained founder loses. **Flag** — direction is my reading, DESIGN-GAP P9. No action uses `recovery`; touch-grass recovery lives in the Soul artifact's recovery activities.) |
| `min_eligible_ppm: 100000` (= floor) | (fixture; loader requires ≤ floor, i.e. always-eligible — the no-death carve-out.) |
| trust `initial = neutral = 500000` | (**default — flag**; design gives no trust initial. Starting at neutral means the relationship is *earned* — consistent with the trust-gates-the-good-stuff intent. Fixtures used 850000/500000 test values.) |
| trust `cap_ppm: 1000000` | **(design/04 §1)** — "cap 100". |
| trust `floor_ppm: 0` | (**default — flag**; design range is 0–100.) |
| trust `gain_ppm_per_effective_action: 20000` (+2) | **(design/04 §1, degraded)** — design gives per-action gains (+2 feed / +3 play / +1 pet / +1 groom); the ruled grammar has ONE uniform gain (DESIGN-GAP P3). +2 (the feed value, the median) chosen. |
| trust `decay_ppm_per_grid: 42` (≈ 1 pt/attended-day) | **(design/04 §1, adapted)** — "decays 1/day of absence", re-based to the ruled attended clock (P5); 10000 ppm/day ÷ 240 grids = 41.67 → 42 (rounding **flag**; exact value would need a day-divisible grid). Decay only pulls values above neutral toward neutral (ruled monotone bound, C18). |
| mood floors 0 / 250000 / 500000 / 750000 | (fixture — `testdata/pet/catalog-grammar-v1.json` (the reviewed grammar fixture) and both full-catalog fixtures agree; **default — no design source, flag**. Even quartiles over the min-stat care scalar.) |
| behavior rows | (superset of the reviewed `care-transition-v1.json` FSM (its 5 rows kept in spirit: idle/care_response/resting cycle) plus rows so `active` and `resting` respond to care and `active` exits on tick — without an `active`+`grid_tick` row the pet would stick in `active` forever on ticks. Durations 1–2 grids (6–12 min) — **default, flag**.) |

### Species roster (CANNOT be minted — see DESIGN-GAP P1)

The task asked for a 2–4 species launch roster. **The ruled pets artifact has no species or
temperament rows** — C17 pinned the artifact's exact top-level and it is care/trust/mood/behavior
policy only; `docs/pet-care.md` states "Production pet identity/species rows … remain
unimplemented," and pet acquisition/starter creation is an explicitly deferred successor (RFC open
question + C2). `pet_records.species_id`/`temperament` exist in the ruled wire with no catalog
authority behind them. So the roster below is a **proposal for the successor
acquisition/starter-content RFC**, not a mintable row set — minting it now would require an
owner-ruled grammar extension:

| species_id | Source | Notes |
|---|---|---|
| `species.server_room_cat` | **(design/04 §1)** — "default: the server-room cat (industry-canon)" | Proposed sole starter at launch. |
| `species.robot_vacuum` | (design/04 §1 — "a robot vacuum with a personality") | Later adoptable variant. |
| `species.rubber_duck` | (design/04 §1 — "a rubber-duck (debugging canon)") | Later adoptable variant. |
| `species.keychain_blob` | (design/04 §1 — "Tamagotchi-style keychain blob") | Later adoptable variant. |

(The Blucifer horse is a conspiracy-tier unlock — out of launch scope by design.)
Temperaments are already a closed six-member combat authority (`lazy|playful|curious|sassy|shy|
chaotic`, `server/combat/arithmetic.go`); how a starter's temperament is assigned (choice? seeded
draw?) is undefined — part of P1.

---

## 4. Copy keys required (must ship through `copy/catalog/*.json` + `make copy-generate` + `verify-copy` BEFORE the mint — FCE5.4)

Achievements (13 keys — the loader hard-requires them in `copykeys.All()`):

1. `achievement.career_attended_day`
2. `achievement.career_attended_hour`
3. `achievement.exit_count_5`
4. `achievement.first_gate`
5. `achievement.gate_burn_t3`
6. `achievement.generators_owned_100`
7. `achievement.generators_owned_300`
8. `achievement.generators_purchased_1`
9. `achievement.generators_purchased_25`
10. `achievement.generators_purchased_25_tier_3`
11. `achievement.old_hand`
12. `achievement.tier_5`
13. `achievement.possession_warning` — the possession-justification key (the pull-the-curtain
    tooltip: states that this is an ownership check and why it's honest). Family name taken from
    the reviewed parity fixture's registry.

Meters: **none** (the meter schema carries no copy fields; band-change copy is a client concern
keyed off band IDs, out of artifact scope).

Pet Care: **none in the artifact** (rejections use the closed detail enum, not copy keys). The
copy pipeline's `copy/references.v1.json` must register the achievements artifact's copy-bearing
paths (`achievements[].copy_key`, `achievements[].proof.justification_copy_key`) if it has not
already; that registration is part of the copy landing, not this row set.

Copy TEXT for all 13 keys is deliberately not drafted here — player-facing text goes through the
flavor bible voice rules (`design/08 §1`) and the copy pipeline's provenance gates.

---

## 5. DESIGN-GAP register (everything a row needed that design does not cover — nothing below was silently improvised)

### Meters

- **M1 — No live producer for any meter-feeding ledger fact.** No Company intent emits
  `externality.*` / `darkpattern.*` (or any) ledger-fact kinds at epoch 6 (`state.LedgerFactKinds`
  is only ever written on the Founder stream at Exit). `doom.probability`'s `externality.emitted`
  input is retained from the reviewed fixture but is inert until a producer RFC lands.
- **M2 — "Safety spending lowers p(doom)" (design/02 §7) has no bindable source** — no
  safety-spend upgrade/content exists; the fixture's `generator.example` binding referenced a
  nonexistent source and was dropped rather than promoted.
- **M3 — Grievance initial value and decay direction are unspecified in design/02.** The fixture
  placeholder (initial 50, decay toward 50) manufactures grievance spontaneously; this proposal
  diverges (initial 0, decay toward 0 @ 1/h) on derivation-rule grounds. **Owner ruling needed:
  promote fixture bytes per FCE2 default, or accept the divergence.**
- **M4 — Band IDs/floors have no design source** (`low@0/high@70` fixture placeholder kept for
  all 11 meters).
- **M5 — p(doom) initial (50) and all decay targets/rates have no design numbers** — fixture
  literals kept, provisional (epoch-7 retune lane).
- **M6 — Standing catalog initials are structurally dead** (Notoriety reseed overrides at new-run
  assembly) — kept at 50; a future schema could drop them, but that is a grammar change, not
  balance.
- **M7 — Design's Trust/pressure feeders** (moderation cuts, scandals, crunch, board pressure,
  regulatory heat inputs — design/09 §3 table) **have no representable sources at epoch 6**; the
  meter input sets are near-empty by necessity, not choice.

### Achievements

- **A1 — The production registry has no `fact:*` provenance sources**
  (`FoundationAchievementRegistry`), so `fact_present` + provenance rows are unloadable, and no
  dated-fact achievements (exit types, dark-pattern acts) are possible until the registry grows a
  fact vocabulary. Registry growth is code (structural authority), not balance data — needs its
  own reviewed change.
- **A2 — design/02 §6 targets "~600 achievements at launch across all systems"**; this set is 12.
  The rest of the milk-model camp (pet-trust achievements, minigame wins, "trigger AGI In Two
  Years five times") requires condition arms that do not exist in the ruled Phase-A union
  (no pet/minigame/meter/event predicates). Successor content as those unions grow.
- **A3 — Score grants were never re-specced after the C1 Clout split.** design/02 §6's "+4 Clout
  per achievement" is carried onto `achievement_score` as a flat 4 (this also normalizes the two
  reviewed fixture rows' 6/8 test values). Owner should confirm the flat-grant reading.
- **A4 — Notoriety has no producer** (no harm-ledger content), so career-notoriety achievements
  would be dead rows; none drafted.
- **A5 — No manual-click counter is registered** — the genre-canonical "first click" achievement
  cannot exist yet.
- **A6 — No tier-1/tier-2 gates exist in routes phase0** (tier jumps 0→3 at `gate.t2_to_t3`), so
  a true "Tier 1 (~15 min)" achievement (design/02 §11 pacing) is impossible; `first_gate`
  (tier ≥ 1) is written to latch correctly if such gates are ever minted.

### Pet Care

- **P1 — No species/temperament slot exists in the ruled pets artifact** (C17 pinned top-level;
  `pet_records.species_id`/`temperament` have no catalog authority; starter creation is a
  deferred successor). The proposed roster (§3) cannot be minted without an owner-ruled grammar
  extension or the acquisition successor RFC. **This is the largest gap between the task's ask
  ("2–4 species roster") and what is mintable.**
- **P2 — Multi-stat action effects are not representable** (design/04: pet −5 Energy, play −15
  Energy; the grammar is single-stat, positive-delta only). Drafted as single-stat rows;
  secondary costs dropped, flagged.
- **P3 — Per-action trust gains are not representable** (+2/+3/+1/+1 by action in design/04; the
  grammar has one uniform `gain_ppm_per_effective_action`). Uniform +2 used.
- **P4 — Age stages (kitten 3/h, adult 2/h, elder 1/h) have no axis** in the artifact or state;
  adult rate used for all pets.
- **P5 — design/04's decay clocks conflict with the ruled attended clock**: "elapsed-time decay"
  for stats and trust "1/day of absence" both price absence; the ruled mechanics (PC1, C4, C10,
  and the offline-progress law) decay on *attended* time only — attended-time equivalents used.
  Design/04 §1 should eventually be reconciled to the ruled clock.
- **P6 — Personality/food-preference gating** (sassy refuses pellets, play gating by energy,
  item variants) is not representable — no temperament qualifiers in the deterministic FSM or
  action grammar (C17: correct-by-omission, revisit via ruling if wanted).
- **P7 — Cooldown values have no design source** (design's only scarcity is diminishing returns);
  the grammar requires them — defaults flagged.
- **P8 — Stat floor value (no-death floor) unspecified** — fixture 10 % kept.
- **P9 — Soul-gate classes per action are interpretive** (essential vs ordinary split direction
  from design/02 §8) — needs owner confirmation.
- **P10 — `grid_ms` is unspecified by design** — 360000 chosen for exact per-hour integer rates;
  the trust rate then rounds (41.67 → 42 ppm/grid ≈ 1.008 pts/day).

---

## 6. Cross-cutting mint observations (outside the three foundations but found during this work — flag to the mint owner)

1. **Doctrines fixture will not compose with unchanged routes bytes.** The reviewed doctrines
   parity fixture (`balance/testdata/doctrines-catalog-parity-v1.json`) declares
   `transition.t3_to_t4` with `gate_id: "gate.t3_to_t4"`, which does not exist in
   `balance/routes/phase0.json`; `doctrine.Catalog.ValidateRoutes` hard-fails on a missing gate at
   bundle composition. Since FCE1 requires all 8 artifacts (pitch → … → doctrines chain) and FCE3
   keeps the 7 base artifacts byte-unchanged, the doctrine lane must either supply a different
   reviewed doctrine artifact or the routes artifact needs a reviewed `gate.t3_to_t4` addition
   (which is a base-artifact byte change — contradicting "bytes unchanged" — or an epoch-6 routes
   revision). Not resolvable inside the meters/achievements/pets row sets.
2. **Copy ordering**: the achievements artifact cannot load until the 13 copy keys are generated
   into `server/copykeys/generated.go` — the copy landing precedes (or accompanies) the mint
   commit, and `deployment/content-manifest.v1.json`'s `copy_hash` moves with it (FCE3.4).
3. **No `balance/pets.schema.json` exists** — the pet artifact is guarded by the Go/TS loaders
   only; if the schema gate (`client/tools/verify-schema.mjs`) is expected to cover every minted
   artifact family, pets (and any other schema-less family) need either a schema or an explicit
   exemption note in the mint changelog.
4. **Meters/achievements must mint together** (loader biconditional `meters ⇔ achievements`,
   `validArtifactNames`) — already FCE1, restated because these two row sets are a pair, not
   independently promotable.
5. The draft artifacts in this proposal are at
   `scratchpad/draft/{meters,achievements,pets}-first-content.json` next to this file, exactly as
   validated; the validation harness is at `scratchpad/validate/`.

---

## Owner rulings — 2026-08-07 (Marco, via decision round)

1. **FCE-B1 → EXTEND ROUTES.** The routes artifact gains the missing `gate.t3_to_t4` row
   (base-byte change: re-accept + content gate + review). Doctrine row unchanged.
2. **Grievance → initial 0, decay toward 0** (the draft's divergence stands; re-runs the meters
   content gate as a recorded retune). **Marco's EU4-estates idea recorded and ROUTED:** grievance's
   long-term form is an EU4-estates-style equilibrium (decay target DERIVED from current posture/
   modifiers, not a static number). The ruled meters grammar has static decay targets only, and the
   estates mechanic is exactly design/09 §Layer-2 (EU4 disaster/pressure model: visible meters
   driven by player choices with contributing factors listed) — so the equilibrium form lands with
   the Layer-2 pressure evaluator RFC, not epoch 6. Recorded there as a design note, not a gap.
3. **Achievement scoring → TIERED 2/4/8** (overrides the flat-4 draft; design/02 §6's milk model
   gets a tiered amendment note). Applied to the draft artifact: 2 = trivial/first-touch
   (attended_hour, first_gate, purchased_1); 4 = session milestones (attended_day, owned_100,
   purchased_25, old_hand); 8 = long-arc (exit_count_5, gate_burn_t3, owned_300,
   purchased_25_tier_3, tier_5). NOTE: first_gate (reviewed fixture row) changes 4→2 — the whole
   artifact goes through the content gate + review as a recorded retune anyway.
