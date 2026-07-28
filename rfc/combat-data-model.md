# RFC: Combat Data Model (shared: pet battles + the Lane)

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-07-28
- **Design refs:** `design/04 §2` (pet battles, as re-decided 2026-07-28), `design/03 §10` (the Lane), `design/05 §4` (PvP table)
- **Research:** `design/research/creature-battler.md §8` (the minimum viable spec + simulations), `design/research/lane-pusher-design.md` (shared-data-model finding, bypass verbs, Nash results)
- **Depends on:** RFC-0002 (catalog patterns), Save Layer & Migrations (accepted)
- **Planning:** `planning/combat-data-model/` (once implementing)

## Summary

One combat **content layer**, two **engines**. Both combat games — the turn-based pet battler and the continuous-time Lane — share a single data model: the six-Temperament type chart, one stat schema, one balance harness, one sprite family. The engines stay separate (~600 lines each; a duel and a lane are different physics). This RFC specifies the shared layer and the two engine contracts; both simulations that validated it are reproducible from the research files.

## Motivation

Two independent simulators (~29,000 matches total) converged on the same laws; encoding them as spec prevents both games from re-deriving or violating them. Out of scope: matchmaking/ratings service (PvP RFC), tournament seasons, cosmetics, the pet care sim itself (cattery port — already specced in `design/04 §1`).

## Specification

### D1 — The type layer

- **Six Temperaments** (= the cattery personalities: lazy/playful/curious/sassy/shy/chaotic) on a **6-cycle chart**: each beats the next two, loses to the previous two.
- **Multipliers 1.3× / 0.77×** — never 2× (simulated: 2× produces 100% win rates in 1v1) — **plus +1 stamina on advantage** (independently reproduces Cassette Beasts).
- Units/pets are typed `Temperament × Runtime` (the swappable second axis) → **36 identities** sharing sprites via the CSS-palette system.
- ⚠️ **Synergy-tag layer required before any roster >12 ships** (`liveservice-idle-tier.md`'s gap): without decorrelated team axes, the Lane's Nash-support-1 result predicts one best bench everywhere. Tags are balance data; the harness gate below enforces the outcome.

### D2 — The stat contract

- **Hardcapped identical stat ceiling for every pet/unit of a given identity** (owner decision 2026-07-28). Care/investment buys **options, consistency, tempo**: Trust→Obedience (smooth 50%→30% across Trust 1.00→0.80), crit-rate doubling, survive-at-1-HP, stamina regen.
- **Soul modulates Obedience in both engines** (the On-Call Leader ability in the Lane; command reliability in duels).
- **All arena math is int32.** No `Decimal` in combat. Replays are bit-exact across Go and TS by construction.

### D3 — Engine contracts

| | Duel engine (pets) | Lane engine |
|---|---|---|
| Time | Turn-based, simultaneous commitment, no switching | Continuous, fixed tick |
| Player | Live or async-vs-snapshot | **Live only**; opponent is async snapshot + bot |
| Depth source | Coverage moves (mandatory day one — they convert a 100% matchup cliff to 67–76%) | Real-time tempo; **bypass verbs mandatory** (buildings-only unit + cheap cycle + ≥1 spell) |
| Resource | Stamina (Temtem overexertion), not PP | Elixir-style flow |
| Resolution | `(seed, teams, choice list)` → deterministic replay | `(seed, decks, timed placements)` → deterministic replay |

### D4 — Verification & bots

- **A battle is its inputs.** Server re-simulates every submitted match (cheap: the research measured ~13,000× real time); PvP results are derived from replays, never client-reported.
- Bots: duel = expectimax over the small action space with personality-flavored policies; lane = the greedy policy family with **published behaviour-flag difficulty manifests**. **Real ratings, rating-based matchmaking, no rubber-band** (owner decision 2026-07-28). Bots never cheat.
- **Matchmaking payouts carry the punch-down multiplier: 200% punching up, 5% four tiers down.**

## Deviations from design

None — this encodes the 2026-07-28 decisions already applied to `design/03 §10` and `design/04 §2`.

## Acceptance criteria

1. Golden replays: identical `(seed, inputs)` produce byte-identical battle logs in Go and TS, both engines.
2. Chart property tests: no 2× path exists; every identity has ≥2 winning and ≥2 losing matchups.
3. **Harness gate (the anti-Nash-1 test):** across the sampled deck/team space, the best loadout's field win rate ≤ 65% and Nash support ≥ 3 — run per balance change to combat data.
4. Obedience curve matches spec across the Trust range; Soul modulation visible in logs.
5. Lane decisiveness ≥ 40% core-kills with the mandated bypass verbs present (regression: removing them drops it below 10%, asserting the verbs are load-bearing).
6. A bot at published manifest level `n` beats level `n−1` ≥ 60% of the time (difficulty monotonicity).

## Open questions

- Synergy-tag vocabulary (D1) — balance data, needs a content pass; blocks roster growth, not the engines.
- Season/rotating-weakness mechanics (`design/04 §2` tournament note) — follow-up.

## Changelog

- 2026-07-28: created (draft).
