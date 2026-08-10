# RFC: Combat — Duel Engine (pet battles)

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-07-29 (split from Combat Data Model per Codex review blocker #1)
- **Design refs:** `design/04 §2`
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

### D4 — Pre-acceptance hardening (2026-08-08 — closing the edges an acceptance review would bounce)

1. **The commitment hash (D2's `committed(hash only until reveal)`):** `fnv1a` over the C4
   canonical serialization of the committed action — both constructions already exist in the
   parent; nothing new. Engine-level commitment is LOG-SHAPE ONLY: the pure function receives
   both choice lists as inputs, so there is no hidden information at this layer — commitment
   secrecy (challenge flows, live play) is the integration RFC's wrapper concern, stated here so
   nobody "fixes" the engine into holding secrets.
2. **Choice-list exhaustion (RULED):** the pure function consumes one action per side per round;
   a `choice_list` exhausted before the battle reaches a terminal is a **terminal engine error**
   (`illegal_action`), exactly like an illegal move — callers (bots, the wrapper) must supply
   enough actions; the engine never repairs input. A list longer than the battle is fine (excess
   ignored). Negative test per AC3.
3. **Stamina saturation:** upkeep gains and rest restoration saturate at the species'
   `stamina_max` (catalog, parent C2) — stamina is bounded `[0, stamina_max]` and the AC5 fuzz
   invariant covers both bounds.
4. **Snapshot v1 exact keys (D3, pinned):**
   `{v: 1, team: [species_id × 3], care_derived: {crit_table: <parent C2 crit-table shape>,
   obedience_inputs: <parent C5 input shape>}}` — byte-sorted keys, canonical serialization per
   C4; captured at challenge time; the shapes are the PARENT's catalog objects by reference,
   never re-declared here.

## Open questions

- Status-effect vocabulary beyond `guarding` — content pass, grows the closed union by RFC.

## Changelog

- 2026-07-29: created from the four-way split; answers parent-review blocker #5 (duel half).
- 2026-08-06: non-normative reference cleanup for publication; no spec change.
- 2026-08-08: D4 pre-acceptance hardening — commitment-hash binding (log-shape only), choice-list
  exhaustion ruled terminal, stamina saturation bounds, snapshot v1 exact keys.

## Codex acceptance-review blockers (2026-08-10 — DU-C1–DU-C8)

D4 closes several edges, but the duel cannot yet be implemented without inventing the parent
catalog and most transition bytes.

### DU-C1 — The required parent catalog is not implemented

The shipped Combat foundation owns arithmetic/RNG only. There is no `balance/combat` artifact,
strict loader, 12-species/24-move fixture, move-effect union, or Trust/Soul table. The parent RFC
and canonical docs explicitly say those pieces remain open DESIGN-GAPs.

**Proposed contract:** block this child on completing Combat Shared Data C2/C5 first. The parent
must acceptance-review and land the strict Go/TS catalog, full fixture, exact effect union, input
tables, artifact registration, and content hash before duel code starts.

### DU-C2 — The pure function input envelope is not exact

`(seed, teams, choice_lists)` omits the pinned catalog bytes/hash/version and does not define team
instance state versus species definitions. D3's species-ID-only snapshot cannot by itself supply
HP, moves, stats, or the per-pet care inputs used during resolution.

**Proposed contract:** enumerate `duel_input.v1` with exact keys for match seed, catalog identity,
two ordered teams of three instance rows, frozen care-derived inputs, and two choice lists. The
engine receives an already-validated immutable catalog object as a separate argument and asserts
the input catalog hash; no deploy-current lookup is permitted.

### DU-C3 — Choice and commitment bytes remain ambiguous

The list container, round association, canonical action field names, move-ID validation order,
hash width/encoding, and whether rejected choices emit a partial committed event are unspecified.
“fnv1a” is not enough to distinguish FNV-1a32/64 or text/binary output.

**Proposed contract:** define choice lists as exact arrays indexed by zero-based round, each action
using the D1 closed exact-key union. Commitment is lowercase 16-hex FNV-1a64 over the parent's
canonical action JSON bytes. Both commitments are appended slot 0 then slot 1 before either reveal;
an invalid/exhausted choice returns a typed engine error and no battle log.

### DU-C4 — Move effects and simultaneous-resolution semantics are missing

`guard`/`utility` effects are not declared by the parent, and D1 does not say when stamina is
debited, how guarding modifies damage, whether a slower action executes after its pet faints,
which target a move hits, or how simultaneous forfeits resolve. These choices change every golden
log.

**Proposed contract:** after the parent effect union is ruled, append one ordered resolution table
covering validation, obedience, debit/rest conversion, priority/speed, target selection, guard,
damage, faint cancellation, entry, dual forfeit, and upkeep. Every branch names its emitted event
sequence; there is no implicit “usual turn-based behavior.”

### DU-C5 — Disobedience-to-rest can mean two different economies

A failed obedience check “converts the move to rest,” but it is unclear whether the original move
spends stamina, whether rest restores half stamina, and whether `rested` accompanies `disobeyed`.

**Proposed contract:** rule one literal arm. Recommended: obedience is checked before debit; failure
spends no move stamina, performs the normal saturating half-rest, and emits `disobeyed` then
`rested` then any `stamina_changed`. This preserves the stated conversion rather than creating a
third action type.

### DU-C6 — The closed event union has names but no payload grammar

No event has exact keys, round/slot coordinates, before/after values, state-hash placement, or
ordering rules. `battle_ended{winner,reason}` leaves the reason union open. Byte-identical replay is
therefore untestable.

**Proposed contract:** enumerate every event payload and the terminal reason union, with one shared
JSON fixture consumed by both loaders. Pin total ordering for commit/reveal, speed ties, multi-faint
cascades, stamina changes, and terminal state hash. The outer C4 battle log exact keys remain
unchanged.

### DU-C7 — Snapshot v1 references undefined parent shapes

`crit_table` and `obedience_inputs` are described as parent catalog objects “by reference,” but C5
defines tables over integer inputs and Pet Care currently exports only `{pet_trust_ppm,soul}`. The
snapshot also cannot carry only species IDs if it freezes care per pet.

**Proposed contract:** snapshot rows carry each pet instance ID/species ID plus the exact frozen
two-integer Pet Care output; catalog policy tables remain solely in the pinned combat artifact.
Delete duplicated tables from the snapshot. Bind identity/ownership and privacy to the later
integration RFC, while the pure engine consumes this exact immutable shape.

### DU-C8 — Error/result and fuzz contracts are not reproducible

“Terminal engine error” could mean a returned error, a terminal battle result, or a partial log.
The 10k fuzz gate has no seed, generator domain, shrink corpus, or invariant list beyond three
examples.

**Proposed contract:** return the closed union `{ok:{log}} | {error:{category:"illegal_action",
detail,round,slot}}`, with errors producing no log. Pin one fuzz seed and exact generation bounds,
then enumerate HP/stamina bounds, active-index validity, one terminal event, event/state agreement,
round monotonicity, catalog immutability, and termination within 100 rounds.
