# RFC: Combat Shared Data & Arithmetic

- **Status:** draft (split parent, 2026-07-29 — was "Combat Data Model"; the four-way split Codex's review required)
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-07-28
- **Design refs:** `design/04 §2` (pet battles, as re-decided 2026-07-28), `design/03 §10` (the Lane), `design/05 §4` (PvP table)
- **Research:** `design/research/creature-battler.md §8`, `design/research/lane-pusher-design.md`
- **Depends on:** RFC-0002 (catalog patterns), Save Layer (implemented), Balance Harness (implemented — SplitMix64)
- **Children:** `combat-duel-engine.md`, `combat-lane-engine.md`, `combat-bots-and-integration.md`
- **Planning:** `planning/combat-shared-data/` (once implementing)

## Summary

The shared content layer both combat engines consume: the type chart, stat schema, catalog objects, exact integer arithmetic, and the determinism contract. Engines, bots, and statistical gates live in the three child RFCs. Nothing here runs a battle; everything here is what both battles agree on.

## Specification

### C1 — The type layer (unchanged laws)

- **Six Temperaments** (lazy/playful/curious/sassy/shy/chaotic) on a 6-cycle: each beats the next two, loses to the previous two.
- Advantage **×13/10**, disadvantage **×10/13 — both exact rationals** (see C3; "0.77" was display shorthand), **+1 stamina on advantage**. Never 2×.
- **Runtime axis (closed, enumerated):** `{container, vm, bare_metal, serverless, edge, mainframe}` → 36 identities `Temperament × Runtime`, sprites via the CSS-palette system. Runtime carries no chart interaction at Phase 0 — it is the synergy-tag carrier (each Runtime declares its tag list in the catalog; tags are balance data). **Roster >12 remains gated on a populated tag vocabulary** (harness gate in the bots RFC).

### C2 — Catalog objects (loadable, versioned, strict)

New catalog file family `combat/*.json`, loaded by the strict loader like every catalog:

- `species`: `{id, temperament, runtime, stats: {hp, atk, def, spd, stamina_max}, moves: [move_id × 4], tags: [string]}` — stats are the **hardcapped identity ceiling** (int32; owner decision 2026-07-28: care buys options/consistency/tempo, never stats).
- `moves`: `{id, temperament | null (neutral), power int, stamina_cost int, kind: "strike"|"coverage"|"guard"|"utility", effect: closed union}` — `coverage` moves are typed off-temperament by definition (day-one mandatory; they convert 100% matchup cliffs to 67–76%).
- `lane_units` / `lane_buildings` / `lane_spells`: `{id, temperament, cost int, hp, atk, spd, range, targets: "units"|"buildings"|"any", tags}` — the mandated bypass verbs are catalog rows flagged `bypass: true` (≥1 buildings-only unit, ≥1 cheap cycle card, ≥1 spell — loader-validated presence).
- `teams` (duel: exactly 3 species refs) and `decks` (lane: exactly 8 refs) are player data validated against the catalog; wire schema `{v: 1, ids: [...]}` exact-key.
- **Minimum Phase-0 fixture committed with this RFC:** 12 species (2/temperament), 24 moves, 8 lane units, 3 buildings, 3 spells — enough for every child RFC's tests, explicitly not balance-final.

### C3 — Exact integer arithmetic (the parity contract)

- All combat state is **int32**; every intermediate is computed in int64 then **saturated to int32 on store** (no wrap, ever).
- Rational multipliers are exact: advantage `x*13/10`, disadvantage `x*10/13`, **integer multiply first, then floor-divide** — one rounding site per multiplier application, in declared order: `base power → attacker atk scaling (×atk/64) → chart multiplier → crit (×3/2) → floor at each step`.
- **Damage minimum 1** after all multipliers if base power > 0. Stamina clamps to `[0, stamina_max]`. HP clamps to `[0, hp_max]`.
- No `Decimal`, no floats, no native `/` anywhere in combat paths (the economy-kernel.ts:334 lesson is a lint rule here: TS combat modules use a `idiv(a,b)` helper; direct `/` in `combat/` fails CI).

### C4 — Determinism contract

- **PRNG: the harness's SplitMix64, exactly** (declared dependency). Seed expansion: `battle_seed = splitmix(match_seed)` per battle; each consumer draws from its own labeled substream `splitmix(battle_seed ⊕ fnv1a(label))` — labels are catalog-declared strings (`"crit"`, `"obedience"`, `"bot_policy"`), so adding a consumer never shifts another's draws.
- Bounded draws use **rejection sampling** (multiply-shift bias is forbidden); the helper is shared kernel code with golden vectors.
- Canonical input ordering: duel = the choice list as committed (simultaneous choices ordered by player slot 0,1); lane = placements sorted `(tick, player_slot, placement_seq)`.
- **Battle log schema:** `{v: 1, engine: "duel"|"lane", seed, catalog_hash, inputs, events: [...]}` with a closed per-engine event union (owned by each engine RFC) and a terminal `state_hash` = fnv1a over the canonical serialized end state. Golden replays assert byte-identical logs Go↔TS.

### C5 — External inputs: fixture-only boundaries (blocker #8 resolved)

Trust→Obedience (smooth 50%→30% across Trust 1.00→0.80) and Soul modulation are **functions over two int inputs `(trust_ppm, soul)` declared here** (integer piecewise-linear tables in the catalog); *where those inputs come from* is the pet-care RFC and Prestige P1 respectively — until those land, engines consume committed fixtures. Matchmaking/rating identity is the bots/integration RFC's boundary. **No combat RFC blocks on the care sim; the seam is these two integers.**

## Acceptance criteria

1. Strict-loader round-trip of the full Phase-0 fixture; every closed union exact-key validated; bypass-verb presence loader-enforced.
2. Golden vectors for C3 arithmetic (chart multipliers at extremes, saturation, damage minimum, crit ordering) asserted byte-identical in Go and TS.
3. PRNG substream isolation: adding a labeled consumer leaves all other substreams' golden draws unchanged (regression fixture).
4. Chart property tests: no 2× path; every identity ≥2 winning and ≥2 losing matchups.
5. Obedience/Soul tables match spec across the input ranges (golden vectors, both runtimes).
6. The `/` lint rule fails a seeded violation in `combat/`.

## Open questions

- Synergy-tag vocabulary content pass — blocks roster growth, not this RFC.
- Runtime-axis mechanical meaning beyond tags — future RFC, deliberately inert now.

## Changelog

- 2026-07-28: created (draft) as "Combat Data Model".
- 2026-07-29: Codex acceptance review required a four-way scope split and recorded seven further executable-contract gaps.
- 2026-07-29: split executed — this file is now the shared data/arithmetic parent (blockers #2, #3, #4, #8 answered here as C2–C5); engines, bots, and statistical gates moved to the three child RFCs.
