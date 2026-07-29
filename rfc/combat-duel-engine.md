# RFC: Combat — Duel Engine (pet battles)

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-07-29 (split from Combat Data Model per Codex review blocker #1)
- **Design refs:** `design/04 §2`
- **Research:** `design/research/creature-battler.md §8`
- **Depends on:** Combat Shared Data & Arithmetic (parent — C1–C5 are normative here)
- **Planning:** `planning/combat-duel-engine/` (once implementing)

## Summary

The turn-based engine: 3v3, simultaneous commitment, no switching, stamina economy. Pure function `(seed, teams, choice_lists) → battle_log` in both runtimes; no I/O, no clock, no identity — the integration RFC wraps it.

## Specification

### D1 — State machine (closed)

Phases per round: `commit → reveal → resolve → upkeep`. Battle state: `{round, sides: [{active_idx, pets: [{species_id, hp, stamina, status}]}]}`. Status is a closed union `{none, guarding, exhausted}` (grows by RFC).

- **Commit:** each side commits one action for its active pet: closed union `action = {move: move_id} | {rest} | {forfeit}`. Legality: move must be one of the species' four and `stamina ≥ stamina_cost`; an illegal commitment is a **terminal engine error** (`illegal_action` — callers validate first; the engine never repairs input). Zero-legal-moves ⇒ only `{rest}` is legal (Temtem overexertion: rest restores `stamina_max/2`, floor).
- **Reveal/resolve:** both actions resolve in `spd` order (tie → slot 0 first — deterministic, published). Damage per parent C3; crit chance from care-derived crit table, drawn on the `"crit"` substream; Obedience check (parent C5) on the `"obedience"` substream — a failed check converts the move to `{rest}` and logs `disobeyed`.
- **Upkeep:** faint check (hp 0 → next pet in team order auto-enters; no switching verb exists), stamina +1 to both actives, +1 more to chart-advantaged active (parent C1), round++.
- **Terminal:** all pets of a side fainted, `{forfeit}`, or round = 100 (hard cap → winner = side with more surviving pets, then more total hp, then **draw** — draws are a real result).

### D2 — Event union (the log)

`round_started, committed(hash only until reveal), revealed, moved, disobeyed, damaged, fainted, entered, stamina_changed, guarded, rested, battle_ended{winner: 0|1|draw, reason}` — closed; the log's `state_hash` per parent C4.

### D3 — Async-vs-snapshot

An async duel runs the same pure function; the opponent's `choice_list` is produced by a bot policy (bots RFC) over the **defender snapshot**: `{team (species ids only), care_derived: {crit_table, obedience_inputs}}` — no hidden state exists in a duel beyond uncommitted choices, so the snapshot is exactly the defender's team sheet. Snapshot schema versioned `{v: 1, ...}`, captured at challenge time.

## Acceptance criteria

1. Golden replays: fixed `(seed, teams, choices)` fixtures (≥ 6, covering crit, disobey, faint-cascade, rest-lock, round-cap draw, forfeit) byte-identical Go↔TS.
2. Coverage-move law reproduced: the parent's fixture roster shows no 100% matchup at the team level (property test over all 3v3 pairings).
3. Illegal actions produce `illegal_action` terminally; no repair path exists (negative tests per union arm).
4. Round-cap tiebreak and draw path exercised by fixture.
5. Fuzz: 10k random battles, zero invariant violations (hp/stamina bounds, monotonic round, terminal reached).

## Open questions

- Status-effect vocabulary beyond `guarding` — content pass, grows the closed union by RFC.

## Changelog

- 2026-07-29: created from the four-way split; answers parent-review blocker #5 (duel half).
