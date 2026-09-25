# RFC: Server Garden (wall-clock crossbreeding minigame)

- **Status:** accepted — owner batch acceptance 2026-09-25; implementing
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-09-25
- **Design refs:** `design/03 §1` (Server Garden), `design/03` clock taxonomy (Wall-clock: "real
  time, production-immune"), `design/03` preamble rules 1–5 (distinct clock, economy hook, AI
  fallback, staggered unlock, persistence across Exits), `design/03` unlock stagger (Tier 2),
  `design/03 §12a` (dailies doctrine: no chores, no absence pricing), `design/03 §12b` (tier→minigame
  scaling seam), `design/03 §3` (the FtHoF lesson: no external-predictor meta), `design/02` Fiscal
  Quarters (Investor Confidence buys minigame unlocks and building levels), `design/00 §5`
  (everything on different clocks), `design/11 §2/§7` (one hint + one why-tooltip; tabs, depth ≤ 2;
  accessibility baseline), `design/08 §1` (voice rules; copy keys only here)
- **Depends on:** Minigame Platform Foundation (accepted, implementing; C1–C40 — this RFC reuses its
  faucet grammar, conversion kernel, and attended-day window, and names three platform amendments);
  Minigame & Recovery API + Surface (accepted, implementing; the surface/registry discipline);
  Fiscal Quarters (implemented; unlock sink and building levels); Founder Attendance and
  Founder-scoped transitions (implemented; `ApplyFounderLogged`, attendance sample); Soul
  Foundation (implemented; `soul_gate` vocabulary); API Foundation (operation registry for the
  read endpoint); Accessibility of Player Workflows (draft; the surface acceptance floor);
  First Content Epoch lane (production mint). **Blocked for production mint on Tier-2 content**,
  which does not exist yet (the host generator is owed by it — OD-14).
- **Parent / amends:** content tenant of the Minigame Platform; amends it by the named SG-P1–SG-P3
  below (or, per OD-2, those move to a separate platform successor RFC).
- **Supersedes / superseded by:** —
- **Planning:** `planning/minigame-server-garden/` (once implementing)

## Summary

Server Garden is the garden-equivalent of the minigame suite: a grid of rack slots in which the
Founder plants **strains** (fictional server/software cultivars), which mature on **wall-clock
ticks** that no production rate can accelerate. Mature strains seed adjacent empty slots; rare
adjacency pairings mutate into new strains at sub-1% per-tick odds; harvesting a mature strain adds
it permanently to the Founder's **seed collection** (the Pokédex) and pays a governed one-shot cash
harvest through the platform faucet. A whole-garden **substrate** trades tick speed against harvest
effect and mutation odds. The garden is Founder-scoped and persists across Exits.

The platform cannot host this as a session tenant. Its sessions are bounded, run-pinned,
one-per-Founder, and block Exit while active. The `clock: wallclock` field sketched in MP5 was
never carried into the C37 artifact grammar. D-013 (the deferred composed async lifecycle) covers
`async_snapshot` PvP sessions, and the garden needs neither that mode nor its lifecycle. The garden
therefore ships as a **persistent Founder mechanic** on the Fiscal Quarters/Pet Care pattern. It
reuses the platform only where a platform seam is the law: the faucet governor and the payout
kernel. The minimal platform successor scope that this requires is stated explicitly in SG0.

## Motivation

`design/03` puts the first major minigames at Tier 2 so they land in the mid-arc sag. It also puts
the garden on the only wall-clock time signature besides pet care. The platform has shipped one
real tenant, The Pitch, which runs on a session-skill clock. A second tenant on a genuinely
different clock tests whether the platform's abstractions generalize. The main finding of this
draft is that they do not generalize to persistent wall-clock state, and that the honest fix is
small.

**In scope:** the garden state machine (grid, plant, growth, spread/mutation, harvest, uproot,
substrate); its lazy wall-clock advance; the seed collection; the cash-harvest hook through the
faucet; Founder v22 persistence and replay; the multi-stream harvest coordinator; a read
projection; the UI surface contract and accessibility floor; copy keys; a structural fixture
catalog.

**Out of scope (each a DESIGN-GAP with an owner decision below):** passive strain payloads
(uptime buffs, golden-opportunity frequency, daemon spawn rate); the sacrifice ritual; tradeable
gifts and the community seed census; the ~30-strain launch content and its prose; tier-relative
(rate-denominated) payout. Clout drops are excluded by existing law, not by choice (Platform C1).

## Specification

### SG0 — Platform fit, and the minimal successor scope

**Why the garden is not a platform session.** Every shipped platform lifecycle property conflicts
with the garden:

| Platform property (source) | Garden need |
|---|---|
| Session pins `(company_stream_id, run_seq, constants_hash)` (C13, TP-C14) | Garden persists across Exits (`design/03` rule 5) |
| At most one `active\|claimed` session per Founder (MA-C6) | A permanent garden would lock out The Pitch and every later tenant |
| Exit rejects while a session is active (MA-C12) | Growing a garden must never block Exit |
| Scaling inputs frozen at create (MP1/§12b) | Grid size grows mid-life with building levels |
| Create → play → one terminal certified result (C14, MA-C4) | No terminal; many small harvests over months |
| Modes `solo\|async_snapshot`; MP5's `clock` enum absent from the C37 row | Real-time evolution between commands |

**D-013 is not this gap.** D-013 defers the composed `async_snapshot` lifecycle, where a live player
plays against another player's snapshot, to "the first real async-tenant successor." Server Garden
is solo (`design/03 §1`: "AI fallback: none needed"), so it neither consumes nor unblocks D-013. The
missing platform lifecycle here is a different one: **persistent, Founder-scoped, wall-clock
tenant state**. Neither the platform nor D-013 names it.

**Minimal successor scope.** The garden needs exactly three platform amendments and nothing more.
They follow The Pitch's precedent of named amendments inside a content RFC (TP-C2/TP-C6). OD-2
decides whether they stay here or split out:

- **SG-P1 — Persistent-tenant faucet identity.** The `minigame_faucet_window` key domain
  `minigame_id` widens from "declared minigames-artifact row" to "declared faucet owner." A
  faucet owner is either a minigames row or a persistent tenant declared by its own pinned artifact.
  The garden is the first such tenant, as `server_garden`. If the window table enforces the key
  domain in a constraint, the change lands as a new append-only migration. The per-window quota,
  remainder, and cap semantics are unchanged.
- **SG-P2 — Non-session payout coordinator.** The window mutation and conversion kernel currently
  run only inside `ApplyMinigameResolutionTransaction`, with the session claim token as their
  exactly-once authority. SG-P2 exposes them to one additional narrow coordinator,
  `save.Store.ApplyGardenHarvestTransaction` (SG6). Its exactly-once authority is the Founder intent
  record. Lock order stays Founder → Company (C38). The kernel's operation order and
  published formula are unchanged.
- **SG-P3 — Founder advance pre-step.** Garden advance (SG3) runs as a pre-step of an enumerated set
  of Founder transitions, including one existing Fiscal arm. This is the same pattern as the
  Fiscal auto-sweep that already precedes every Founder command.

Nothing else from the platform is reused or changed. There is no session row, no tenant
descriptor, no scaling-row grammar, no offline quality, and no rating.

### SG1 — The `server_garden` artifact (schema v1, exact keys)

A new optional artifact `balance/server-garden/<epoch>.json`. It joins constants identity and the
catalog-bundle hash like every artifact, and Go and TypeScript load one shared fixture:

```json
{
  "schema_version": 1,
  "unlock_id": "minigame.server_garden",
  "host_generator_id": "<a fiscal generator_level_rows id>",
  "soul_gate": "unrelated",
  "grid": {
    "max_width": 6,
    "max_height": 6,
    "dimension_rows": [{"min_level": 0, "width": 2, "height": 2}],
    "hardcap_reason_key": "cap.garden_grid"
  },
  "clock": {
    "catchup_cap_ms": 86400000,
    "catchup_reason_key": "cap.garden_catchup",
    "substrate_lockout_ms": 600000,
    "lockout_reason_key": "cap.garden_substrate_lockout"
  },
  "default_substrate_id": "bare_metal",
  "substrates": [
    {"substrate_id": "bare_metal", "tick_ms": 300000, "effect_ppm": 1000000,
     "chance_factor_ppm": 1000000, "name_copy_key": "garden.substrate.bare_metal.name",
     "tooltip_copy_key": "garden.substrate.bare_metal.tooltip"}
  ],
  "species": [
    {"species_id": "strain_a", "maturation_ticks": 3, "harvest_units": 5, "starter": true,
     "name_copy_key": "garden.species.strain_a.name",
     "description_copy_key": "garden.species.strain_a.description"}
  ],
  "recipes": [
    {"recipe_id": "r_strain_a_spread", "child_species_id": "strain_a",
     "parents": [{"species_id": "strain_a", "min_mature_neighbours": 1}], "chance_ppm": 50000}
  ],
  "payout": {
    "credited_resource_id": "company.cash", "sends_per_day": 8, "per_send_cap": 300,
    "conversion_ppm": 1000000, "payout_score_fact_id": "garden.harvest_units",
    "cap_reason_key": "cap.minigame_faucet"
  }
}
```

**Loader rules.** Any violation fails the load in both runtimes. Every rule has a rejecting fixture
(AC1).

1. Exact keys at every level: unknown, missing, or duplicate keys reject. JSON numbers must be exact
   safe integers under the platform grammar.
2. IDs (`species_id`, `substrate_id`, `recipe_id`) match `^[a-z][a-z0-9_]{0,47}$`. Arrays are
   sorted by their ID in raw bytes with no duplicates. `parents` is sorted by `species_id`, has 1–2
   entries, and has distinct species.
3. Domains:
   - `tick_ms` ∈ [60_000, 86_400_000].
   - `effect_ppm` and `chance_factor_ppm` ∈ [0, 10_000_000].
   - `maturation_ticks` ∈ [1, 1_000_000].
   - `harvest_units` ∈ [0, 1_000_000_000].
   - `chance_ppm` ∈ [1, 1_000_000].
   - `min_mature_neighbours` ∈ [1, 8], and the sum over one recipe's parents is ≤ 8.
   - `catchup_cap_ms` ∈ [1, 604_800_000].
   - `substrate_lockout_ms` ∈ [0, 86_400_000].
   - `max_width` and `max_height` ∈ [1, 6].
4. At least one species is a `starter`. `default_substrate_id` exists. Every recipe's child and
   parents exist.
5. **Reachability (the completable Pokédex).** Compute the fixpoint from the starters: a species is
   reachable if it is a starter, or if some recipe producing it has only reachable parents. Every
   species must be reachable, because a strain that cannot be discovered is a load error.
6. **Geometric satisfiability.** Every recipe's parent-count sum is ≤ the largest Moore
   neighbourhood available inside `max_width × max_height`.
7. `dimension_rows`: `min_level` strictly ascending starting at 0; `width` and `height`
   non-decreasing, within [1, max]; the last row equals `(max_width, max_height)`, so the
   hardcap is a visible row.
8. **Evaluation budget.** `ceil(catchup_cap_ms / min(tick_ms)) × max_width × max_height ≤
   2_000_000`. This makes per-command work structurally bounded (SG3).
9. `payout` loads through the platform's existing six-key payout loader unchanged.
   `payout_score_fact_id` must be exactly `garden.harvest_units`, the garden's only score fact.
10. Cross-artifact rules: `unlock_id` ∈ Fiscal `unlock_rows`; `host_generator_id` ∈ Fiscal
    `generator_level_rows`; every copy/reason key resolves in the copy manifest.
11. Across epochs, species, substrate, and recipe IDs are **append-only**. A retune may change
    values but never remove an ID. A save referencing an unknown ID is an invariant failure
    (`constants_mismatch`, fail loud) and never a silent drop.

### SG2 — Founder v22 state (exact keys, replay-owned)

`server_garden` is added to the Founder save at **v22**, the next link on the scalar chain (C36;
OD-19). The rule is: `server_garden` artifact pinned ⟺ Founder floor ≥ 22. v22 requires the whole
lower chain (`minigames`, `pets`, `fiscal`, `soul`, `minigame_api`) to stay pinned. Activation
follows the Fiscal precedent: at a new-run boundary from the pinned next bundle, or at New-Founder
initialization. The Company axis is unchanged.

```json
"server_garden": {
  "salt_hex": null,
  "tick_anchor_wall_ms": null,
  "tick_seq": 0,
  "substrate_id": "bare_metal",
  "substrate_set_wall_ms": null,
  "plots": [
    {"row": 0, "col": 1, "species_id": "strain_a", "age_ticks": 3, "matured_effect_ppm": 1000000}
  ],
  "seed_collection": ["strain_a", "strain_b"]
}
```

- Activation initializes an empty garden: `plots: []`, `seed_collection` = every starter,
  `substrate_id` = the default, and the salt, anchor, and lockout stamp all `null`.
- The codec validates the following. Violations reject in both runtimes.
  - `salt_hex` is null or 16 lowercase hex characters.
  - The anchor and stamp are null or safe integers.
  - `tick_seq` ∈ [0, MaxExactInteger].
  - Plots are sorted by `(row, col)` and unique, with `row < 6` and `col < 6`.
  - `age_ticks` ∈ [0, 1_000_000], and `matured_effect_ppm` is null or ∈ [0, 10_000_000].
  - `seed_collection` is sorted, unique, and a superset of the starters.
- **Stage is derived and never stored:** `mature ⟺ matured_effect_ppm != null`, following the
  Pet Care rule for derived mood.
- Exit never touches `server_garden`: it survives every Exit byte-identically. A New Founder gets a
  fresh garden, because a New Founder is a genuinely new lifecycle.

### SG3 — Clock and lazy advance (wall-clock, OD-1)

The clock is the **server-authored wall timestamp** that `founder_log` already stamps on every
Founder command (`server_ms`), following the Fiscal Quarters precedent. There is no background tick
loop, no per-player server timer, and no client-supplied time.

**Advance inputs:** state, `server_ms`, the pinned catalog, the Founder's Fiscal level of
`host_generator_id` (which gives the active grid via the last `dimension_rows` row whose
`min_level ≤ level`), and `unlocked`. The garden is unlocked when the Founder's Fiscal unlocked set
contains `unlock_id`. The algorithm is exact and integer-only:

```
if not unlocked: return unchanged                              # garden dormant, clock not started
if salt_hex == null: salt_hex = <server-drawn, frozen in resolved inputs>   # OD-12
if anchor == null: anchor = server_ms; return (0 ticks)
if server_ms <= anchor: return unchanged                        # clock regression: no-op, never reject
elapsed   = server_ms - anchor
forfeited = 0
if elapsed > catchup_cap_ms:                                    # visible hardcap, OD-4
    forfeited = elapsed - catchup_cap_ms; anchor += forfeited; elapsed = catchup_cap_ms
T = tick_ms(substrate_id)
n = floor(elapsed / T)
apply ticks tick_seq+1 .. tick_seq+n (SG4), with the fixed-point skip below
tick_seq += n; anchor += n * T
```

- **Partition invariance.** Randomness is keyed by the absolute `tick_seq`. No PRNG cursor is ever
  stored, which follows the Pitch precedent. Advancing t0→t2 in one step therefore equals
  t0→t1→t2 whenever the catch-up cap does not bind (AC3).
- **Fixed-point skip.** After any tick, the garden may have no growing plant in the active region
  and no empty active plot with an eligible recipe. In that case the remaining ticks are exact
  no-ops: `tick_seq` and `anchor` advance arithmetically. This is exact because no-op ticks consume
  no draws.
- **Visible truncation.** A nonzero `forfeited` appears in the receipt and in `garden_advanced.v1`,
  with `catchup_reason_key`. Silent truncation is forbidden (evidence discipline rule 3).
- **Advance trigger set (SG-P3, OD-20).** The advance runs as a pre-step, after the Fiscal
  auto-sweep and before the command body, of exactly:
  - every garden command (SG5);
  - `spend_fiscal_credit`, which can raise the host level (and therefore the grid) or buy the unlock.

  No other Founder transition mutates a garden input. Any future transition that does must join
  this set, and AC6's property test proves the set is sufficient. If the command rejects, the
  pre-step rolls back with it, as the Fiscal sweep does.

### SG4 — The tick transition (deterministic, Go/TS byte-parity)

**Per-tick PRNG.** For absolute tick `s`:

- `garden_base = Substream(uint64(salt), "garden.founder.v1").Next()`
- `rng_s = Substream(garden_base XOR uint64(s), "garden.tick.v1")`

These use SplitMix64 exactly as in `server/determinism`. The *active region* is rows `< height`
and columns `< width` of the current dimension. Plots outside it are **dormant**: they never grow,
never count as neighbours, and are never spawned into.

**Step 1 — growth.** For each active plot with a growing plant, in `(row, col)` order:
`age_ticks = min(age_ticks + 1, 1_000_000)`. If `age_ticks ≥ maturation_ticks(species)`, set
`matured_effect_ppm = effect_ppm(current substrate)`. The plant matures, and its harvest effect is
frozen at the substrate in force at maturity (OD-8). If a retune lowers `maturation_ticks` below a
plant's age, the plant matures on its next processed tick.

**Step 2 — census.** For each empty active plot, count its **Moore (8-neighbour)** active plots
holding a *mature* plant, per species (OD-6). The census is taken once, after step 1. Seedlings
spawned in step 3 are immature, so they cannot affect other plots in the same tick.

**Step 3 — spread and mutation.** For each empty active plot in `(row, col)` order:

1. `eligible` = the recipes, in `recipe_id` order, for which every parent satisfies
   `census[parent.species_id] ≥ parent.min_mature_neighbours`.
2. If `eligible` is empty, skip the plot. No draw is consumed.
3. Otherwise draw `u = rng_s.Bound(1_000_000)`. Walk `eligible` with
   `eff = min(1_000_000, floor(chance_ppm × chance_factor_ppm(substrate) / 1_000_000))` and
   `cumulative = min(1_000_000, cumulative + eff)`. The first recipe with `u < cumulative` places a
   seedling of `child_species_id` (`age_ticks 0`, `matured_effect_ppm null`) and ends the walk.
   If no recipe matches, the plot stays empty.

Recipes whose child equals a parent are the design's "mature units seed adjacent slots." Recipes
with a different child are the crossbreeding mutations. Chaos Monkey's "×3 mutation" is its
`chance_factor_ppm` (3_000_000), applied to all recipe chances, which is the Cookie Clicker
wood-chips rule. The cumulative clamp at 1_000_000 in canonical order is published (OD-21).

**Invariants.** Plants never die or decay (OD-5). The garden therefore evolves monotonically
between commands toward a fixed point, and absence never costs anything already grown
(`design/03 §12a`, `design/00 §6`).

### SG5 — Founder commands (closed union, server-authoritative)

All four commands join the existing authenticated Founder intent route, the one carrying
`care_action` and the Fiscal intents, as additive arms. Schemas are exact, and extra fields
(including any time, tick, outcome, or species-result field) reject at decode:

- `{intent_id, kind: "garden_plant", expected_revision, row, col, species_id}`
- `{intent_id, kind: "garden_uproot", expected_revision, row, col}`
- `{intent_id, kind: "garden_harvest", expected_revision, plots: [{row, col}]}`: 1–36 entries,
  sorted by `(row, col)`, duplicate-free.
- `{intent_id, kind: "garden_set_substrate", expected_revision, substrate_id}`

`row` and `col` ∈ [0, 5], checked at decode. `expected_revision` is the Founder revision.

**Validation order** (after the Fiscal sweep and the garden advance). Each rejection uses the
existing closed categories with new closed `detail` members, and mutates nothing:

1. Artifact not pinned or Founder < v22 → `not_eligible/garden_inactive`.
2. Unlock not owned → `not_eligible/fiscal_unlock_required` (the Pitch detail, reused).
3. `soul_gate: human_hobby` and near-zero Soul → `not_eligible/human_content_locked`. This check is
   inert under the recommended `unrelated` (OD-13).
4. Per command:
   - **plant:**
     - Plot outside the active region → `not_eligible/plot_dormant`.
     - Plot occupied → `not_eligible/plot_occupied`.
     - Unknown species → `unknown_id/garden_species`.
     - Species not in `seed_collection` → `not_eligible/seed_not_collected`.
     - Otherwise place a seedling (`age_ticks 0`). Planting is free (OD-9).
   - **uproot:**
     - Empty plot → `unknown_id/garden_plot`.
     - Otherwise remove the plant, with no payout and no seed. Dormant plots are allowed.
   - **harvest:**
     - Any listed plot empty → `unknown_id/garden_plot`.
     - Any listed plant not mature → `not_eligible/plant_not_mature`.
     - Otherwise run SG6. Dormant mature plants may be harvested.
   - **set_substrate:**
     - Unknown substrate → `unknown_id/garden_substrate`.
     - Equal to the current substrate → `not_eligible/substrate_unchanged`.
     - Lockout active (`substrate_set_wall_ms` non-null and `server_ms − substrate_set_wall_ms <
       substrate_lockout_ms`) → `not_eligible/substrate_lockout`.
     - Otherwise set the substrate, re-anchor at `server_ms` (the in-progress partial tick is
       forfeited and reported as `partial_tick_forfeited_ms`), and stamp `substrate_set_wall_ms`.

Plant, uproot, and set_substrate commit through `ApplyFounderLogged` with the resolved-input arm
`garden_command.v1`: `{server_ms, advance: <SG8 advance summary>}`. Harvest commits through SG6.

### SG6 — Harvest payout (multi-stream, through the faucet)

**Founder side** (pure, in both runtimes):

- `units_i = floor(harvest_units(species_i) × matured_effect_ppm_i / 1_000_000)`
- `total_units = Σ units_i` (checked; bounded < 2^53 by the SG1 domains)

The harvest removes every listed plant. It adds each harvested species not yet in the collection
to `seed_collection`; these are reported as `seeds_discovered`. "Seeds permanent once harvested":
the collection is never reduced.

**Coordinator** (SG-P2). `save.Store.ApplyGardenHarvestTransaction` is modeled on
`ApplyMinigameResolutionTransaction`. Under one Postgres transaction it:

1. locks Founder, then the active Company sibling;
2. validates both expected revisions;
3. freezes the server timestamp and the Founder attendance sample (`ValidateFounderAttendanceSample`);
4. runs the Founder transition (advance pre-step + harvest);
5. if `total_units > 0`, applies the platform conversion kernel with `score = total_units`, zero
   fallback reduction (solo), and the window `(founder_id, "server_garden", attended_day)`;
6. credits `company.cash` with the saturating Decimal credit;
7. appends both logs, both revisions, the events, the window update, and the Founder intent record.

**Zero harvest.** A `total_units = 0` harvest still commits both revisions, with the Company arm
recording `faucet_applied: false`. It never touches the window, so it consumes no daily send. Seeds
are recorded regardless of quota: a harvest beyond the daily send cap still discovers seeds, and it
forfeits units with the visible `cap_reason_key`.

**Idempotency.** The authority is the Founder intent record, keyed by `(founder stream, intent_id)`
with a canonical request hash. A retry returns the stored receipt without re-running the transition
or the faucet. The Company internal intent ID equals the Founder `intent_id`.

**Cross-stream binding** (the C39 pattern). The harvest hash is
`harvest_hash = SHA-256(canonical {intent_id, plots:[{row,col,species_id,units}], total_units})`.
Both logs bind it:

- Founder arm `garden_harvest.v1`: `{server_ms, attendance_sample, advance, plots, total_units,
  seeds_discovered, harvest_hash}`.
- Company internal command `{kind: "credit_garden_harvest", intent_id, harvest_hash}`. Its
  resolved arm is `{policy_hash, selected_units, faucet_before, faucet_after, credited
  (canonical Decimal string), forfeited_units, cap_reason_key | null, faucet_applied,
  founder_log_coordinate}`.

**Fault injection** follows every write and proves all-or-none (AC8).

Payout magnitudes are fixed integers through the shipped kernel, at parity with The Pitch. A
tier-relative ("N seconds of rate") harvest needs a new platform payout arm (OD-10).

### SG7 — Unlock, Soul, scaling, fallback, rating, offline quality

- **Unlock.** Purchased with Investor Confidence through the Fiscal `unlock` target
  `minigame.server_garden` (Fiscal owns the cost row). Host-tier eligibility is **not enforceable
  today**: Fiscal unlock rows carry only `{unlock_id, cost}`, a gap shared with The Pitch (OD-15).
- **Soul gate.** `unrelated` by default (OD-13). The garden is corporate infrastructure, not the
  Founder's hobby, and it is not a recovery activity because it pays (`design/03 §5`).
- **Scaling (§12b).** One input, `garden.grid_dimension`, sourced from the host generator's Fiscal
  level. Its class is breadth/resource-pool; the garden is unranked, so the Fairness Law is not
  engaged. It is resolved at every advance from replay-owned Founder state rather than frozen at a
  session create (Deviation 5). It uses the garden's own `dimension_rows`, because the platform's
  closed scaling-source union has no Fiscal-level arm.
- **Fallback.** Solo by design (`design/03 §1`). **Rating:** none. **Offline quality:** not
  applicable, because the garden has no automation destination.
- **Exit and platform sessions.** The garden is not a platform session, so it never triggers
  MA-C12's Exit rejection and never occupies the one-active-session slot.

### SG8 — Events, receipts, replay

These are new closed event kinds (schema v1), registered in the event registry and backed by one
append-only migration extending the DB kind constraint (the active-play `00067` precedent):

- `garden_advanced.v1` (Founder): `{ticks_applied, tick_seq_after, catchup_forfeited_ms,
  catchup_reason_key|null, matured:[{row,col,species_id}], spawned:[{row,col,species_id,
  recipe_id}]}`. Emitted iff `ticks_applied > 0` and (matured or spawned is non-empty), or iff
  `catchup_forfeited_ms > 0`. It precedes the command's own event.
- `garden_planted.v1`: `{row,col,species_id}`.
- `garden_uprooted.v1`: `{row,col,species_id,was_mature}`.
- `garden_substrate_set.v1`: `{from_substrate_id,to_substrate_id,partial_tick_forfeited_ms}`.
- `garden_harvested.v1` (Founder): `{plots:[{row,col,species_id,units}], total_units,
  seeds_discovered, harvest_hash}`.
- `garden_harvest_credited.v1` (Company), plus the ordinary resource event: `{harvest_hash,
  selected_units, credited, forfeited_units, cap_reason_key|null, faucet_applied}`.

The *advance summary* inside resolved inputs is the `garden_advanced.v1` payload.

**Replay.** Both runtimes replay Founder history from genesis. They recompute every advance from
`(state, row server_ms, pinned catalog, Founder Fiscal level, unlocked)` and compare state,
receipt, and ordered event bytes. The salt is read from the resolved input of the row that
initialized it. The Company run log replays `credit_garden_harvest` from its resolved arm and
checks the shared `harvest_hash`. A missing artifact reports `constants_mismatch`; any byte
difference reports `state_divergence`.

### SG9 — Read surface

`GET /api/v1/garden/current` (authenticated; Founder identity from the token only) returns HTTP 200
with one of three shapes:

- `{kind: "inactive"}`
- `{kind: "locked", unlock_id}`
- `{kind: "active", founder_revision, server_ms, garden}`

In the `active` shape, `garden` is exact:
`{width, height, max_width, max_height, substrate_id, substrate_lockout_until_ms|null,
next_tick_wall_ms|null, tick_seq, pending_catchup_forfeited_ms, seed_collection, species_total,
plots:[{row, col, species_id, stage: "growing"|"mature", age_ticks, maturation_ticks, dormant}]}`.

- The server computes it by running the SG3 advance on a **discarded clone** at database `now()`
  (the Game UI projector precedent). A read never commits.
- The projection is advisory, and the command receipt is authority. Because draws are keyed by
  `tick_seq`, a later commit's ticks are a superset-prefix-extension of the projected ticks.
- The response never contains `salt_hex`, `garden_base`, any draw, or any future tick (OD-12).
- The endpoint registers through the API Foundation operation registry and generator (MA-C14
  precedent; no handwritten router), using the exact error pairs of SG5 where they apply.

### SG10 — UI surface contract (consumed by a Game-UI successor)

A `garden` surface is a Tier-2-era tab in the host (office/campus) chrome. It is visible only when
the read returns `locked` or `active`, and navigation depth stays ≤ 2 (`design/11 §7`).

- **Props:** the generated `GardenView` DTO and the Copy/Presentation catalog.
- **Callbacks:** `plant(row,col,species_id)`, `uproot(row,col)`, `harvest(plots[])`,
  `setSubstrate(substrate_id)`. They are dispatched only through `client/src/game-ui/runtime.ts`;
  components import no transport.
- **"Harvest all mature"** is client composition of one `garden_harvest` over the view's mature
  plots, not a new command.
- **Closed UI states:** `locked | loading | active | stale | error`.
- **No client simulation.** The surface re-reads at `next_tick_wall_ms` while visible (the
  Page Visibility API gates it), and after every command receipt. Growth is tick-granular, so
  nothing interpolates.
- **DOM-first.** The garden is a grid of native buttons, with no canvas (`design/06`).
- **One contextual hint and one why-this-matters tooltip** (`design/11 §2`); `?` opens the codex key.

**Accessibility floor.** This applies the Accessibility of Player Workflows A1–A5 floor once that
RFC is accepted. AC13 is blocked until then.

1. The grid is `role="grid"` with a roving tabindex. Arrow keys move between plots, and
   Enter/Space opens the plot's action menu (native buttons). There are no drag, hover-only, or
   pointer-only gestures. Substrate tooltips (the curtain text) are reachable by focus and tap.
2. Every plot's accessible name comes from the copy template `garden.plot.label`
   (`{row} {col} {species} {stage}`). Stage and dormancy are conveyed by text plus a non-hue glyph
   shape, never by color alone.
3. Tick changes are announced once per refresh in a polite live region via
   `garden.announce.tick` (counts of matured and spawned plants). There is no per-plot announcement
   flood.
4. `prefers-reduced-motion` is honored at mount and on change. Growth animations become static
   stage glyphs.
5. At 320 CSS px, a 6×6 grid of ≥ 24×24 targets fits without horizontal page scroll.
6. **No time limits.** The garden requires no timed input and never punishes absence (WCAG 2.2.1
   by construction). The next-tick countdown is informational and server-relative.

### SG11 — Copy keys (keys only; all prose owner-authored, pending)

The keys below are shipped. Fixture rows carry placeholder text that is marked unratified under
the copy pipeline, and no line ships as final copy without owner adoption (evidence discipline
rule 6):

- `garden.title`, `garden.hint.first`, `garden.why`, `garden.codex`
- `garden.species.<species_id>.name`, `garden.species.<species_id>.description`
- `garden.substrate.<substrate_id>.name`, `garden.substrate.<substrate_id>.tooltip`
- `garden.stage.growing`, `garden.stage.mature`, `garden.plot.empty`, `garden.plot.dormant`,
  `garden.plot.label`
- `garden.action.plant`, `garden.action.uproot`, `garden.action.harvest`,
  `garden.action.harvest_all`, `garden.action.set_substrate`
- `garden.collection.progress` (`{count}/{total}`), `garden.announce.tick`
- `garden.next_tick` (countdown)
- Reason keys: `cap.garden_grid`, `cap.garden_catchup`, `cap.garden_substrate_lockout`, plus the
  reused `cap.minigame_faucet`
- `error.garden.<detail>` for every SG5 detail

The design's example strain names (`Slackware Bonsai`, `Kube Vine`, `Legacy Perl Bramble`) are
owner flavor. Whether real software names may appear in species copy is an owner/legal call (OD-16).

### SG12 — Fixture catalog (structural, provisional values)

This implementation ships a **fixture-only** artifact that exercises every rule. All values are
provisional balance data, not launch content:

- **Substrates** (tick/effect values from `design/03 §1` and the `research/cookie-clicker.md §4.1`
  soils):

  | Substrate | `tick_ms` | `effect_ppm` | `chance_factor_ppm` |
  |---|---|---|---|
  | `bare_metal` | 300_000 | 1_000_000 | 1_000_000 |
  | `containerized` | 180_000 | 750_000 | 1_000_000 |
  | `mainframe` | 900_000 | 1_250_000 | 1_000_000 |
  | `chaos_monkey` | 300_000 | **750_000 (design: "reduced", value unspecified — owner)** | 3_000_000 |

- **Species:**

  | Species | Starter | `maturation_ticks` | `harvest_units` | Purpose |
  |---|---|---|---|---|
  | `strain_a` | yes | 3 | 5 | |
  | `strain_b` | yes | 4 | 8 | |
  | `strain_c` | no | 6 | 20 | |
  | `strain_d` | no | 8 | 40 | |
  | `strain_e` | no | 5 | 0 | the zero-harvest path |

- **Recipes:**

  | Recipe | Child | Parents | `chance_ppm` |
  |---|---|---|---|
  | `r_a_spread` | `strain_a` | `strain_a` ≥ 1 | 50_000 |
  | `r_b_spread` | `strain_b` | `strain_b` ≥ 1 | 50_000 |
  | `r_c_cross` | `strain_c` | `strain_a` ≥ 1, `strain_b` ≥ 1 | 5_000 |
  | `r_d_pair` | `strain_d` | `strain_c` ≥ 2 | 2_000 |
  | `r_e_cluster` | `strain_e` | `strain_a` ≥ 3 | 8_000 |

- **Dimension rows** (Cookie Clicker Farm progression, re-based to level 0):

  | Level | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 |
  |---|---|---|---|---|---|---|---|---|---|
  | Grid (w×h) | 2×2 | 3×2 | 3×3 | 4×3 | 4×4 | 5×4 | 5×5 | 6×5 | 6×6 |

- **Clock:** `catchup_cap_ms` 86_400_000; `substrate_lockout_ms` 600_000.
- **Payout:** the SG1 row as shown. The Fiscal fixture gains the unlock row
  `{unlock_id: "minigame.server_garden", cost: <provisional 3>}`. The host is the fixture generator
  row, until Tier-2 content supplies the real one.

The ~30-strain launch catalog (`design/03 §1`), its recipe graph, and its names are an owner
content batch, due before the production mint (OD-16). The loader's reachability and
satisfiability rules gate it.

### SG13 — Activation and mint

The implementation is fixture-first (TP-C18 precedent). Production activation requires all of:

1. a protocol-compliant mint through the epoch lane (BALANCE-CHANGE discipline);
2. Tier-2 content, supplying the host generator and its Fiscal level row;
3. the owner-ratified launch catalog and copy;
4. OD-15's tier eligibility.

The garden is **not** in the Phase-0 preview, which is T0–T1 only (D-007). No harness pacing claim
is made: the balance and relevance harnesses do not model the garden, and a faucet-capped payout is
bounded by construction.

## Deviations from design

1. **Clout drops removed.** `design/03 §1` lists them as a payload. Platform C1 and Achievements C1
   forbid any second Clout mint; Feed/Social owns it.
2. **Passive payloads deferred** (uptime buffs, golden-opportunity frequency, daemon spawn rate).
   Each needs a mid-run Founder→Company effect seam that does not exist (the Fiscal effects are
   run-frozen at run start), and daemons are unimplemented Tier-3 content. v1 ships only the
   one-shot cash harvest (OD-10).
3. **Sacrifice ritual deferred.** Its "permanent bonus" is unspecified, and its "+ Confidence" would
   be a second `fiscal_credit` mint (OD-11).
4. **MMO hooks deferred.** No trading layer exists for tradeable gifts, and the seed census needs a
   World/Commons aggregate plus a privacy ruling (D-015).
5. **Scaling input not frozen at session create.** No session exists; the input is resolved from
   replay-owned Founder state at each advance (SG7).
6. **Not a platform session tenant**, despite the coverage map's "all minigames on the platform."
   The garden rides the platform's faucet only, plus three named amendments (SG0).
7. **Catch-up hardcap** (if OD-4 is adopted). The design says "real time" with no cap. The cap is an
   added visible number that bounds compute and mirrors the offline law.
8. **Host-tier unlock not enforced** in the fixture (OD-15).

## Acceptance criteria

Every criterion names the discriminating failing case that must be demonstrated red before it counts
as green (evidence discipline rule 1). Go gates run with `-count=1`.

1. **Loader.** Every SG1 rule rejects its own fixture in Go and TS. That includes an unreachable
   species, a recipe unsatisfiable within the max grid, an over-budget clock, a non-final hardcap
   row, a missing fiscal unlock or host row, and a removed ID versus the prior artifact.
   *Failing case:* with the reachability check disabled, the unreachable fixture loads, and the
   test must go red.
2. **Activation.** `server_garden` pinned ⟺ Founder floor ≥ 22. A bundle pinning it without
   `minigame_api` rejects. v21 rejects garden state and v22 rejects its absence. Activation happens
   only at a new-run boundary or at New-Founder initialization.
   *Failing case:* a v21 save carrying `server_garden` must reject.
3. **Partition invariance.** The shared Go/TS corpus shows that advancing t0→t2 equals t0→t1→t2 for
   every t1 on a grid of cases below the cap.
   *Failing case:* a demonstration variant with a stored cursor-based PRNG breaks the property.
4. **Tick golden vectors (byte parity).** The corpus covers maturation, effect freezing at maturity,
   spread, two-parent cross, `min > 1`, a diagonal-only neighbour, dormant plots, the chaos factor,
   the cumulative clamp, the fixed-point skip, and a retune below age.
   *Failing case:* a 4-neighbour implementation fails the diagonal vector.
5. **Catch-up cap.** An elapsed time above the cap reports exact `catchup_forfeited_ms` with its
   reason key in the receipt and the event.
   *Failing case:* a truncating implementation without the field fails schema validation.
6. **Trigger-set sufficiency.** A property test compares advancing before *every* Founder command
   with advancing only at the SG3 set, and they must produce identical garden state.
   *Failing case:* removing `spend_fiscal_credit` from the set diverges on a vector where a
   mid-interval level purchase widens the grid.
7. **Commands.** Every SG5 detail is produced by a vector, and a rejection mutates nothing,
   including the rolled-back advance.
   *Failing case:* a variant that commits the pre-step on rejection is caught by a revision/state
   comparison.
8. **Harvest transaction.**
   - Faults injected after every write leave all-or-none across both revisions, both logs, the
     events, the window, and the intent record.
   - A retry returns identical receipt bytes without a second credit.
   - A harvest past the daily sends forfeits with `cap.minigame_faucet` and still records seeds.
   - A zero-unit harvest consumes no send.

   *Failing case:* each fault boundary is demonstrated to leave partial rows when the transaction
   is split.
9. **Replay.** Founder history and the Company run log replay byte-identically in Go and TS, and
   both logs bind the same `harvest_hash`.
   *Failing case:* a tampered hash reports `state_divergence`.
10. **Hidden information** (under OD-12 option b). An enumeration test finds no salt, base, or draw
    in any API response, event, outbox row, or receipt.
    *Failing case:* a planted leak in the read projection must be caught.
11. **Exit and New Founder.** The garden is byte-identical across Exit, Exit is never rejected on
    account of the garden, and a New Founder starts empty with starters.
    *Failing case:* an Exit arm that resets the garden fails.
12. **Server authority.** The intent decoder rejects any client time, tick, outcome, or salt field,
    shown by grep plus a schema test, and the client has no garden simulation module.
    *Failing case:* a request carrying `tick_seq` must reject.
13. **Surface** (blocked on the Accessibility RFC's acceptance). The component passes keyboard grid
    navigation, non-color stage signaling, reduced motion, 320 px reflow, and axe with zero
    serious/critical findings, plus the A5 evidence record.
    *Failing case:* a color-only stage variant fails the structural assertion.
14. **Copy.** Every SG11 key resolves, `verify-client-boundary` finds no literal player text, and
    fixture copy is marked unratified.
    *Failing case:* a literal string inserted into the garden component fails the boundary scan.
15. **Read projection.** At equal `server_ms`, the projected view equals the state the next command
    commits.
    *Failing case:* a projection that omits the fixed-point skip's `tick_seq` arithmetic diverges.

## Owner decisions (numbered, with recommended defaults)

1. **OD-1 — Clock.**
   - (a) **Wall-clock** from the server-authored `founder_log` timestamp, as Fiscal Quarters does.
   - (b) Founder attended time, as Pet Care chose. This would not grow while offline, and the
     garden is a check-in toy.

   **Recommend (a).** It matches `design/03`'s taxonomy and `design/00 §5`.
2. **OD-2 — Where the platform amendments live.**
   - (a) As the named SG-P1–SG-P3 in this RFC (the Pitch precedent).
   - (b) Split into `rfc/minigame-platform-persistent-tenants.md`, with this RFC depending on it.
   - Forcing the garden into sessions is rejected, per the SG0 table.

   **Recommend (a).** The scope is three small seams with one consumer. If a second persistent
   tenant appears (for example the Market's holdings), promote them to (b).
3. **OD-3 — Payout-policy home.**
   - (a) The garden's own `server_garden` artifact, reusing the six-key payout grammar.
   - (b) A minigames-artifact schema v4 with a `lifecycle: session|persistent` discriminator.

   **Recommend (a).** It leaves the accepted C37 grammar and the session API untouched, and the
   session API can then never create a garden session.
4. **OD-4 — Offline catch-up.**
   - (a) A visible hardcap `catchup_cap_ms`, provisionally 24 h to mirror the offline law.
   - (b) Uncapped, with a per-command tick budget that carries deferred ticks forward (the
     active-play due-transition precedent).
   - (c) Uncapped and unbounded.

   **Recommend (a).** Because plants never die (OD-5), a long absence loses only extra mutation
   rolls, never anything already grown.
5. **OD-5 — Plant lifespan.**
   - (a) No death or decay.
   - (b) Cookie-Clicker lifespan and decay.

   **Recommend (a).** It is the no-FOMO law (`design/00 §6`, `§12a`); the design is silent on death.
6. **OD-6 — Adjacency.**
   - (a) Moore, 8 neighbours.
   - (b) Von Neumann, 4 neighbours.

   **Recommend (a).** Cookie-Clicker parity, and the research file's "orthogonally or diagonally."
7. **OD-7 — Substrate swap.**
   - (a) A catalog lockout (provisionally 10 min, as in CC), plus forfeiture of the partial tick on
     change.
   - (b) No lockout.

   **Recommend (a).** Without it, fast-then-slow swapping becomes a degenerate loop.
8. **OD-8 — When substrate effect applies to harvest.**
   - (a) Frozen at the maturity tick.
   - (b) The substrate at harvest time.

   **Recommend (a).** Under (b), switching to Mainframe right before harvesting is a free +25%.
9. **OD-9 — Planting cost.**
   - (a) Free.
   - (b) Company cash, which needs a Company-debit arm from a Founder command.

   **Recommend (a).** The faucet already governs the economy.
10. **OD-10 — v1 payload scope.**
    - (a) Cash harvest only, with passive payloads and a rate-denominated payout arm deferred to a
      "Garden Payloads" successor. That successor requires a mid-run Founder→Company contribution
      seam.
    - (b) Also grant an active-play buff window on harvest now, which amends the Company
      active-play arm.

    **Recommend (a).**
11. **OD-11 — Sacrifice ritual.**
    - (a) Defer to the Payloads successor, which needs a defined "permanent bonus" and an explicit
      ruling on a second Investor-Confidence mint.
    - (b) Specify it now.

    **Recommend (a).**
12. **OD-12 — RNG predictability.**
    - (a) A publicly derivable seed (`runidentity.Seed(founder_id, 0)`), following the Pitch
      precedent.
    - (b) A hidden per-Founder salt: drawn by the server at garden initialization, frozen in that
      Founder-log row's resolved inputs, stored in the save, and never projected to any client or
      export.

    **Recommend (b).** `design/03 §3`'s FtHoF lesson applies directly: with a public seed, an
    external tool could find the ticks where sub-1% mutations fire. Replay stays a pure function of
    persisted inputs. **D-008** (export) must exclude the salt.
13. **OD-13 — Soul gate.**
    - (a) `unrelated`.
    - (b) `human_hobby`, which locks at near-zero Soul.

    **Recommend (a).** The garden is a corporate asset, not a hobby.
14. **OD-14 — Grid growth source.**
    - (a) The Fiscal generator level of the office/campus host generator (the CC Farm-level
      analogue, and Founder-scoped).
    - (b) The purchased count of the host generator, which is Company state and would need a
      cross-scope read.

    **Recommend (a).** Tier-2 content must supply the host ID, and it must have a Fiscal
    `generator_level_rows` entry. The concurrent draft `rfc/tier2-content.md` proposes
    office-category T2 generators (for example `generator.open_plan_floor`). The owner should
    choose the host from whatever that RFC accepts, and not from this draft.
15. **OD-15 — Host-tier eligibility.** Fiscal unlock rows have no tier predicate, which also affects
    The Pitch. **Recommend:** a Fiscal successor amendment adding `min_career_tier`, evaluated over
    Founder Exit history, before production mint. The fixture stays ungated.
16. **OD-16 — Launch content.** **Recommend:** this RFC ships only the SG12 fixture. The
    ~30-strain launch set, its recipe graph, and all names and prose come as an owner content batch
    before mint. The owner or legal must rule whether real software names (Slackware, Kubernetes,
    Perl) may appear in strain copy.
17. **OD-17 — Harvest batching.**
    - (a) One faucet send per `garden_harvest` command covering any set of mature plots.
    - (b) One send per plot.

    **Recommend (a).**
18. **OD-18 — Read path.**
    - (a) A per-family `GET /api/v1/garden/current` (the MA-C6/C10 pattern).
    - (b) A new field in a Game UI snapshot v4.

    **Recommend (a).** It keeps a single owner for the garden read.
19. **OD-19 — Founder version.** **Recommend: v22**, with the `server_garden` artifact as its
    activation authority. Ordering against any other pending Founder mechanic follows C36's scalar
    chain; no other draft currently claims v22.
20. **OD-20 — Advance trigger set.**
    - (a) Garden commands plus `spend_fiscal_credit`, proven sufficient by AC6.
    - (b) Every Founder command (the Fiscal-sweep pattern).

    **Recommend (a).** It is less coupling, and AC6 detects any future omission.
21. **OD-21 — Probability overflow.**
    - (a) A cumulative clamp at 1_000_000 in canonical `recipe_id` order, published.
    - (b) Reject at load any catalog where some reachable neighbourhood could exceed 1_000_000.

    **Recommend (a).** Option (b) is combinatorially expensive to check.
22. **OD-22 — Provisional balance values.** These are the SG12 tick, effect, factor, and dimension
    values, the catch-up cap, the lockout, the payout row, the unlock cost, and Chaos Monkey's effect
    (unspecified in the design). **Recommend:** ratify them as fixture-only; production values ride
    the mint.

## Open questions

- Whether a garden **forecast panel** (the design's "the predictor is in the game") belongs in the
  Payloads successor if players ask for planning tools. This is moot under OD-12(b).
- Whether the collection should feed Achievements (a personal "complete the collection" row).
  That is Achievements content, not this RFC.

## Changelog

- 2026-09-25: created (draft — not implementation authority). Platform-fit analysis: the garden
  needs a persistent Founder-scoped wall-clock lifecycle that the platform does not provide and
  D-013 does not cover. The minimal successor scope is SG-P1–SG-P3.

## Owner acceptance (2026-09-25)

Marco accepted this RFC in the 2026-09-25 batch (the answer to a direct batch question in-session,
recorded by Claude). Every owner decision this RFC lists is **ruled at its stated recommended
default**. Where the text offers alternatives, the recommended option is the normative one and the
alternatives are rejected. Provisional numbers stay provisional and are ratified by harness
measurement and SHA, as the RFC already requires. Owner-authored copy stays owner-authored:
implementation ships copy keys, with any placeholder or candidate text clearly marked.
