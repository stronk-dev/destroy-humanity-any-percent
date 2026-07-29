# RFC: Combat — Lane Engine

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-07-29 (split from Combat Data Model per Codex review blocker #1)
- **Design refs:** `design/03 §10`
- **Research:** `design/research/lane-pusher-design.md`
- **Depends on:** Combat Shared Data & Arithmetic (parent — C1–C5 normative)
- **Planning:** `planning/combat-lane-engine/` (once implementing)

## Summary

The continuous-time engine: one lane, elixir flow, timed placements, core-kill win. Pure function `(seed, decks, placements) → battle_log`; live-only for the human side (integration RFC owns sessions), opponent = snapshot + bot.

## Specification

### L1 — Time and space (the literals blocker #3 demanded)

- **Tick = 100 ms** of match time; match cap **1800 ticks (3:00)**, then sudden-death 600 ticks (first core damage wins), then tiebreak: more core hp → winner, equal → **draw**.
- Lane = integer positions `0..1000`; player 0's core at 0, player 1's at 1000; each side has one mid-lane tower at 300/700. Movement/range in position units per tick, all int (catalog fields).
- **Elixir:** ×10 fixed-point int (`elixir_dds`), +1 dds per 3 ticks, cap 100 dds (10.0); double-rate in sudden death.

### L2 — Placements and legality

`placement = {tick, card_id (from the 8-card deck), position}`; legal iff `tick` within match, elixir ≥ cost at that tick (evaluated in canonical order, parent C4), position on own half `[0,450]` / `[550,1000]`, card in deck with cycle rule: the played card goes to the back of the 8-card rotation, only the front 4 are playable (Clash cycle). Illegal placement = terminal `illegal_action` (same doctrine as the duel engine: callers validate, engines never repair).

### L3 — Resolution per tick (closed order)

`elixir accrual → spawns (this tick's placements, canonical order) → spell effects → targeting (closed rule: nearest legal target by `targets` field, tie → lower position) → movement (unblocked units advance) → attacks (all simultaneous, damage per parent C3, chart multiplier per unit temperaments) → deaths (simultaneous) → tower/core damage → terminal check`. One rounding site per C3; no float velocity — position advances by int speed per tick.

### L4 — Event union

`tick (implicit, not logged), placed, spawned, targeted, moved (logged on change only), attacked, destroyed, tower_destroyed, core_damaged, sudden_death_started, battle_ended{winner, reason}` — closed; `state_hash` per parent C4.

### L5 — Snapshot

Defender snapshot `{v: 1, deck (8 ids), care_derived}` at challenge time — deck composition is the only hidden information and it is revealed by the snapshot by design (asymmetry favors the live player; the bots RFC's difficulty manifests compensate; this is the published policy, not a leak).

## Acceptance criteria

1. Golden replays (≥ 6: cycle rule, sudden death, draw, bypass-unit core kill, spell-only kill, tower race) byte-identical Go↔TS.
2. **Decisiveness gate:** ≥ 40% core-kills across the bots RFC's sampled population with bypass verbs present; removing them (fixture catalog minus `bypass: true` rows) drops below 10% — asserting the verbs are load-bearing (moved here from the parent's AC5; the sampled population is the bots RFC's G-contract).
3. Elixir/cycle legality negative tests per union arm.
4. Fuzz: 10k random matches, zero invariant violations (position bounds, elixir bounds, simultaneous-death consistency, terminal reached).

## Open questions

- Second lane / spatial widening — explicitly out; a future RFC changes the map, not this engine's laws.

## Changelog

- 2026-07-29: created from the four-way split; answers parent-review blockers #3 (lane literals) and #5 (lane half).
