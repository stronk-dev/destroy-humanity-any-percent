# RFC: Demo Disc Arcade (the container and its first two cover-disc toys)

- **Status:** accepted — owner batch acceptance 2026-09-25; implementing
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-09-25
- **Design refs:** `design/03 §8` (Demo Disc Arcade, Tier 0–1, host: the CRT — the evolving
  nostalgia container), `design/03` header (T0–1 minigames are tutorial-tier and free), `design/03
  §Clock taxonomy` (session skill: short skill runs, cooldown-gated), `design/03 §Unlock stagger`
  (Tier 0: Demo Disc Arcade), `design/03 §12a` (dailies doctrine), `design/03 §12b` (scaling seam
  and Fairness Law), `design/03 §5` (2026-08-06 Soul ruling), `design/08 §1` (voice rules),
  `design/08 §2` (era 1995 and 2000 presentation), `design/11 §1b`, `§2`, `§7` (T0 satire beat,
  one hint per system, accessibility baseline), `design/00` (pillars and anti-goals),
  `design/01 §Tier 0` (the `CRT`), `design/BACKLOG.md §Nostalgia arcade sweep`
- **Depends on:** Minigame Platform Foundation (accepted, implementing; C1–C40); Minigame &
  Recovery API + Surface (accepted, implementing; MA-C1–C15, including the MA3 surface
  components not yet built); The Pitch (archived, the template this RFC follows); UI Foundation and
  Game UI Screens (archived); Soul Foundation (archived); Copy Pipeline (archived); First Content
  Epoch (the `minigames`/`minigame_api` artifact chain). It also consumes the draft Accessibility
  of Player Workflows RFC as its accessibility floor if that RFC is accepted first.
- **Parent / amends:** content on the Minigame Platform. It adds **two tenants**, **one pinned
  content artifact**, and **four named platform/composition amendments** (AR-P1–AR-P4). None of
  them changes a ruled platform grammar.
- **Supersedes / superseded by:** —
- **Planning:** `planning/minigame-demo-disc-arcade/` (once accepted)

## Summary

The Demo Disc Arcade is the Tier-0 minigame home. In 1995 it looks like the demo disc from a
magazine cover, playing on the garage CRT. Design gives it three jobs: host tiny era-authentic toys, change
with the era (cover disc → Flash portal → app store), and carry the tier-0 tutorial role. This RFC
specifies:

1. The **container**: its pinned stage manifest, unlock, Soul classification, economy stance,
   exclusivity with the rest of the game, and UI surface.
2. The **first two toys that design names**: the Minesweeper-lineage deduction grid (engine
   `mine_grid`) and a Snake (engine `snake`). Both are specified the same way as The Pitch:
   deterministic pure engines over `(seed, commands)`, Go/TypeScript byte parity, certified
   integer results, and content as exact-key data.

Design's third named toy, the shareware platformer demo that ends with "ORDER NOW: $0.00", is a
declared successor. The later era stages and the rest of the nostalgia sweep are also successors.

The recommended economy stance follows `design/03 §8`'s "pure nostalgia texture". The arcade pays
**zero** into the economy through a zero-rate faucet row and charges no offline quality. It
therefore moves no T0–T1 pacing number, and its rules and correctness can ship before any balance
tuning.

## Motivation

The platform has one real tenant, which sits at Tier 5. Tier 0 has no minigame, but
`design/03 §Unlock stagger` puts the arcade there. `design/11 §1b` names the arcade's ORDER NOW
screen as the ≤10-minute satire beat, and the onboarding research makes the arcade the place where
a player first learns that minigames exist. The capability ledger records the gap directly:
`M-011.02 Play a real arcade tenant from the Demo Disc` is marked `client_or_fixture_only`, and no
arcade tenant is playable.

The platform RFC also has an open question: do arcade toys need full sessions or a lightweight
score-send path? This RFC answers it: **full sessions**. A path where the client sends its own score
is a client-sent result, which violates the server-authority law (`CLAUDE.md` law 2). Full sessions
also keep the arcade ready for the later daily-slot ghost scores.

**Out of scope:** the platformer demo; T2–3 Flash-portal and T5 app-store stages and their toys
(Solitaire/Boss Key, the match-3, block stacker, tower defense, plinko, the 2048-lineage daily
slot); ghost scores and daily seeds; achievement hooks (DG-2); Soul restoration (DG-5); any payout
magnitude; audio; canvas rendering.

## Specification

### AR1 — The container

**AR1.1 Identity and host.** The arcade is a Game UI surface named `arcade`, not a minigame. It
lists toys, and each toy is an ordinary platform minigame definition. The "host: the CRT" in
`design/03 §8` is presentational: the Desk entry point wears the CRT. See DG-6 for the host/tier
mismatch.

**AR1.2 The pinned arcade artifact.** A new artifact `balance/arcade.json` (`schema_version: 1`)
holds the container manifest and the content for both toys. Its exact top-level keys are
`{schema_version, container, mine_grid, snake}`. It joins `CatalogBundle` as `CatalogBundle.Arcade`
under the Pitch precedent (TP-C2/TP-C14), with these rules:

- The canonical hash uses the established constants-hash construction.
- Session content identity lives in exact genesis keys `{arcade_content_hash,
  arcade_schema_version}`. There is no database column.
- Artifact chain: `arcade` requires `minigame_api`, which already requires the full minigame
  content chain. The artifact is added to `validArtifactNames` with the same hard-fail discipline.

`container` has one exact key, `stages`: a list of rows `{stage_id, min_tier, title_copy_key,
toys}` sorted by strictly ascending `min_tier`.

- `toys` is a byte-sorted list of `minigame_id`s.
- At bundle composition, every toy ID must resolve to a `minigames` definition whose `engine_ref`
  is `mine_grid` or `snake`.
- **v1 has exactly one stage:**
  `{"stage_id":"cover_disc","min_tier":0,"title_copy_key":"arcade.stage.cover_disc.title","toys":["arcade.mine_grid","arcade.snake"]}`.
- The client shows the stage with the greatest `min_tier ≤ authoritative tier`.
- The stage is presentation. Server-side permission to start a toy comes only from that toy's
  `unlock_condition`. Tier-gated later stages need a server-enforced unlock arm first (DG-4).

**AR1.3 Era evolution under the UI laws.** Platform C9 removed `era_skins`, and UI Foundation
allows one token theme per era, never per-feature skins. Era evolution is therefore expressed only
through:

- (a) the active stage row, which selects the toy list and title key;
- (b) the Copy Pipeline's existing `era_variants` on arcade copy keys, so the same key reads as a
  1995 cover disc or a 2000s download page;
- (c) the global era theme tokens, `era_1995` and `era_2000`.

Tier 1 keeps the `cover_disc` stage. The Flash portal (T2–3) and the app store (T5) are successor
stage rows (DG-4).

**AR1.4 Unlock.** Both toys use `unlock_condition: {"kind":"always"}` (OD-3), which matches
"Tier 0–1 minigames … are tutorial-tier and free". The epoch-6 Fiscal row
`{"unlock_id":"unlock.arcade","cost":5}` contradicts that line (finding AR-F2). This RFC does not
bind it, and OD-3 recommends retiring it in the arcade mint epoch.

**AR1.5 Soul.** Both toys declare `soul_gate: "human_hobby"` (OD-4). `design/03 §8`'s "Hobby
framing applies" is implemented as the lock half only: near-zero Soul rejects start with the
shipped `not_eligible/human_content_locked`. Soul *restoration* is not implemented. The
2026-08-06 §5 ruling says paying hobbies never restore Soul and that `§5c` cozy activities are the
sole recovery source, and `§8`'s parenthetical contradicts it. That contradiction is the
owner-authored RP-017, which this RFC must not resolve (DG-5).

**AR1.6 Economy stance: texture only (OD-2).** Every arcade definition carries:

- a zero-rate faucet row
  `{credited_resource_id:"company.cash", sends_per_day:0, per_send_cap:0, conversion_ppm:0, ...}`;
- an inert offline-quality row: one curve row whose grade equals the neutral floor, and zero
  decay;
- an inert neutral rating row, with certified `rating_delta: null` (C40).

A resolution therefore credits exactly zero, forfeits exactly zero (so no cap reason appears), and
leaves rating and offline quality unchanged. It still commits the ordinary resolution records
(both logs, events, session receipt) through the shipped composer. This depends on AR-F1 being
fixed, because the shipped composer currently fails closed on a zero credit.

**AR1.7 Clock.** `design/03`'s clock taxonomy puts the arcade on "session skill, cooldown-gated".
The skill runs are the sessions. The platform's only cooldown is the daily faucet quota, and
nothing guards a zero faucet, so v1 has **no play cooldown**. A cooldown on play itself would be a
new platform mechanic (DG-3).

**AR1.8 Exclusivity and Exit.** Arcade sessions are ordinary platform sessions and inherit two
rules:

- (a) one active minigame session per Founder. Starting any minigame while one is active rejects
  `not_eligible/exclusive_activity`.
- (b) MA-C12: while a session is `active|claimed`, the Exit path rejects
  `not_eligible/minigame_session_active`. That path includes `cross_gate`, which drives the scripted
  first-company ending (`server/production/replay.go:629`).

A T0 toy can therefore block the ~15-minute scripted collapse. Three mitigations are normative:

- Every arcade engine has a `quit` command that reaches a terminal state from any non-terminal
  phase, so the player can always clear a session without playing it out.
- The Desk renders the `minigame_session_active` rejection with copy that names the arcade and
  links back to the running toy (AR6.5).
- The arcade entry shows when a session is active.

**AR1.9 Persistence across Exits (`design/03` rule 5).** A session cannot cross an Exit (AR1.8b).
Rating and offline-quality maps persist but stay neutral. An optional **local-only** best-score
display may be kept (OD-11). It follows the gate-splits precedent (`docs/game-ui.md`): a client
display record that never feeds an intent, receipt, or board.

**AR1.10 AI fallback.** Both toys use `fallback: {"kind":"solo"}`; solo play is the design. Design's
"daily-slot ghost scores" belong to the T5 daily slot, a successor.

### AR2 — Definition rows (C37 grammar, literal)

Two rows are added to the pinned `minigames` artifact (schema v3). Both copy the Pitch shape and
change only the literals below. All integers are provisional balance data unless the text marks a
value as structural.

```json
{
  "minigame_id": "arcade.mine_grid",
  "engine_ref": "mine_grid",
  "engine_version": "1.0.0",
  "modes": ["solo"],
  "result_score_fact_ids": ["mine_grid.cells_revealed", "mine_grid.cleared"],
  "scaling": {"schema_version": 1, "scaling_inputs": [
    {"destination": "minigame.arcade.mine_grid", "destination_class": "breadth", "source_kind": "literal",
     "source_ref": "1", "op": "identity", "operand": 0, "clamp_min": 1, "clamp_max": 1}]},
  "payout": {"credited_resource_id": "company.cash", "sends_per_day": 0, "per_send_cap": 0,
             "conversion_ppm": 0, "payout_score_fact_id": "mine_grid.cells_revealed",
             "cap_reason_key": "cap.minigame_faucet"},
  "fallback": {"kind": "solo"},
  "offline_quality": {"score_fact": "mine_grid.cells_revealed",
    "grade_curve": [{"score_threshold": 0, "grade_ppm": 200000}],
    "decay_grid_ms": 3600000, "decay_ppm_per_grid": 0, "neutral_floor_ppm": 200000,
    "automation_destination": "minigame.arcade.mine_grid"},
  "rating_policy": {"starting_elo": 1000, "elo_floor": 0, "elo_ceiling": 3000,
                    "provisional_games": 10, "season_member": "s1"},
  "unlock_condition": {"kind": "always"},
  "soul_gate": "human_hobby"
}
```

The `arcade.snake` row is identical except for:

- `engine_ref: "snake"`;
- `result_score_fact_ids: ["snake.food_eaten", "snake.ticks_survived"]`;
- payout and quality `score_fact`: `snake.food_eaten`;
- scaling destination and automation destination: `minigame.arcade.snake`.

The rows sort byte-wise (`arcade.mine_grid` < `arcade.snake` < `pitch`). The only scaling input is
the literal breadth 1, so the Fairness Law holds trivially.

### AR3 — Engine `mine_grid` 1.0.0 (the deduction grid)

**AR3.1 Content (`arcade.json#/mine_grid`).** Exact keys: `{presets}`. The rows are
`{preset_id, width, height, mines, copy_key}`, byte-sorted by `preset_id`. Loader bounds are
`5 ≤ width ≤ 30`, `5 ≤ height ≤ 30`, and `1 ≤ mines ≤ width*height − 9`. The v1 rows
(provisional, OD-12) are:

- `large {30, 16, 99}`
- `medium {16, 16, 40}`
- `small {9, 9, 10}`

Their copy keys are `arcade.mine_grid.preset.<preset_id>`.

**AR3.2 Coordinates.** Cell index `i = y*width + x`, zero-based and row-major. Neighbors are the
up-to-8 cells at Chebyshev distance 1. Every list of cells is sorted ascending.

**AR3.3 Randomness (no stored cursor).** The engine uses these streams:

- `run_seed = Substream(seed, "mine_grid.run.v1").Next()`
- mine stream: `Substream(run_seed, "mine_grid.mines.v1")`

Mines are placed lazily at the first `reveal`, as follows:

1. Build the ascending eligible list: every cell except `first_cell` and its neighbors.
2. Shuffle it with the shipped descending Fisher–Yates: for `k` from `n−1` down to 1, draw
   `j = Bound(k+1)` using the rejection-sampled `determinism.SplitMix64.Bound`, then swap. This is
   byte-identical to Pitch's `deal`.
3. The mines are the first `mines` entries after the shuffle, sorted.

Because placement is a pure function of `(seed, preset, first_cell)`, the mine set is **never
stored**. The engine re-derives it on every `Apply`.

**AR3.4 Snapshot (exact thirteen keys).**

```
{phase: "setup"|"playing"|"terminal", preset_id: string|null, width, height, mines,
 first_cell, revealed: [{cell, adjacent}], flags: [cell], exploded_cell, mine_cells: [cell],
 revision, arcade_content_hash, arcade_schema_version}
```

- Genesis values: `phase:"setup"`, `preset_id:null`, `width/height/mines: 0`, `first_cell: -1`,
  `exploded_cell: -1`, and empty lists.
- **Hidden-information invariant:** `mine_cells` is `[]` in every non-terminal snapshot. The
  snapshot is the API response (MA-C7) and the stored platform state, so it must never carry mine
  positions before the game ends. At terminal it holds every mine, sorted.

**AR3.5 Commands (closed union, discriminator `kind`).**

| Command | Legal phase | Effect |
|---|---|---|
| `choose_board {preset_id}` | `setup` | Copies the preset dimensions and enters `playing`. |
| `reveal {cell}` | `playing` | Rejects with `cell_flagged` or `cell_revealed`. If `first_cell = -1`, sets it and places mines (AR3.3). A mine sets `exploded_cell = cell`, and the game ends `detonated`. Otherwise the engine flood-reveals (below). |
| `toggle_flag {cell}` | `playing` | Rejects with `cell_revealed`. Otherwise toggles membership in `flags`. There is no flag cap. The display shows `mines − len(flags)`, which may go negative, as in the lineage. |
| `chord {cell}` | `playing` | Requires `cell` to be revealed with `adjacent > 0` and flagged-neighbor count equal to `adjacent`; otherwise rejects `chord_unsatisfied`. Reveals every unflagged hidden neighbor. If any of them is a mine, the game ends `detonated` with `exploded_cell` = the smallest such index. Otherwise each neighbor flood-reveals. |
| `quit {}` | `setup`, `playing` | Terminal `quit`. |

**Flood reveal.** The revealed set is the closure from the start cell: reveal each cell with its
adjacent-mine count. If the count is 0, add every hidden, unflagged neighbor. Flagged cells are
never auto-revealed. The result is order-independent, and the stored list is sorted.

**Win.** When `len(revealed) = width*height − mines`, the game ends `cleared`.

**Every applied command** advances the revision by exactly one. A rejected command mutates
nothing. `cell` must satisfy `0 ≤ cell < width*height`, or the command rejects
`cell_out_of_range`.

**AR3.6 Terminal result.** The terminal transition writes `phase:"terminal"`, fills `mine_cells`,
and emits:

- `outcome`: one of `cleared | detonated | quit`;
- score facts, byte-sorted: `mine_grid.cells_revealed` (safe cells revealed) and
  `mine_grid.cleared` (`1` only on `cleared`, otherwise `0`);
- `rating_delta: null`.

**AR3.7 Rejection taxonomy (sorted).** `cell_flagged | cell_out_of_range | cell_revealed |
chord_unsatisfied | illegal_phase | unknown_preset`.

### AR4 — Engine `snake` 1.0.0

**AR4.1 Content (`arcade.json#/snake`).** Exact keys, all provisional (OD-13):

| Key | v1 value | Loader bound |
|---|---|---|
| `width` | 20 | 5–30 |
| `height` | 20 | 5–30 |
| `start_length` | 3 | 2 to `width/2` |
| `growth_per_food` | 1 | 1–8 |
| `max_ticks_per_advance` | 64 | 1–256 |
| `presentation_tick_ms` | 125 | 50–1000 |

`presentation_tick_ms` is **presentation-only**. The Go engine validates its range and never reads
it. Walls are solid, with no wraparound.

**AR4.2 Randomness.**

- `run_seed = Substream(seed, "snake.run.v1").Next()`
- `food_seed = Substream(run_seed, "snake.food_seed.v1").Next()`

`food_seed` is **published** in the snapshot as a canonical decimal string (OD-5). Food number
`k` goes on the ascending list of empty cells at index
`Substream(food_seed XOR uint64(k), "snake.food.v1").Bound(len(empty))`.

**AR4.3 Snapshot (exact fourteen keys).**

```
{phase: "playing"|"terminal", tick, direction: "up"|"down"|"left"|"right", body: [cell]
 (head first), pending_growth, food_cell, food_index, food_seed, score, width, height,
 revision, arcade_content_hash, arcade_schema_version}
```

Genesis:

- head at `(width/2, height/2)` (integer division), with body cells extending west to
  `start_length`;
- `direction: "right"`, `tick: 0`, `pending_growth: 0`, `score: 0`;
- food 0 placed, then `food_index: 1`.

**AR4.4 The tick step** (tick `t = tick + 1`, in this exact order):

1. **Turn.** If a turn is scheduled at `t`, it becomes `direction`. A turn must be a 90° change;
   the same direction or the opposite direction rejects the whole command with `invalid_turn`.
2. **Move.** Compute `next = head + delta`. If `next` is out of bounds, the game ends `crashed`.
3. **Tail.** `growing = pending_growth > 0`, and `kept = growing ? body : body[:-1]`. The tail
   cell that is leaving is free this tick. If `next ∈ kept`, the game ends `crashed`.
4. **Commit.** `body = [next] ++ kept`. If `growing`, decrement `pending_growth`.
5. **Food.** If `next = food_cell`: add 1 to `score` and `growth_per_food` to `pending_growth`,
   then place food `food_index` (AR4.2) and increment `food_index`. If no empty cell remains, the
   game ends `cleared` and `food_cell` becomes -1.
6. `tick = t`.

On `crashed`, `tick` is the death tick and the body is left unmoved.

**AR4.5 Commands.**

- **`advance {through_tick, turns: [{tick, direction}]}`**
  - Requires `tick < through_tick ≤ tick + max_ticks_per_advance`; otherwise `advance_window`.
  - Every turn tick must lie in `(tick, through_tick]`; otherwise `advance_window`.
  - Turn ticks must be strictly ascending; otherwise `turns_not_ascending`.
  - The engine simulates ticks in order. If a terminal state occurs at tick `d`, the command must
    have `through_tick = d` and no turn after `d`; otherwise it rejects `advance_past_terminal`
    **without mutation**. An honest client simulating locally always submits exactly up to the
    terminal tick. The server never silently discards input.
- **`quit {}`** gives terminal `quit`.

**AR4.6 Terminal result.**

- `outcome`: one of `cleared | crashed | quit`;
- facts, byte-sorted: `snake.food_eaten` (`score`) and `snake.ticks_survived` (`tick`);
- `rating_delta: null`.

**AR4.7 Rejection taxonomy (sorted).** `advance_past_terminal | advance_window | illegal_phase |
invalid_turn | turns_not_ascending`.

**AR4.8 Wall-time trust.** Tenants are pure and receive no clock. The server certifies an input
schedule, not the real time it took to play, so a scripted client could play "slow-motion" Snake.
With zero payout, no boards, and no ghosts this changes nothing (OD-6). Any future board, ghost,
or payout on Snake needs a server-stamped pacing contract as a platform successor.

### AR5 — Platform and composition amendments (named, explicit — the Pitch precedent)

- **AR-P1 — `CatalogBundle.Arcade`.** Artifact name, path, and loader chain as in AR1.2, with a
  loader in both Go and TypeScript that passes one shared fixture. Cross-artifact checks at
  composition cover the stage toys, the engine refs, and copy-key existence.
- **AR-P2 — content resolver generalization.** `ReplayCatalogSet.ResolveTenantContent` currently
  accepts only `pitch` (`server/production/replay.go:114`). It becomes a closed table:
  `(pitch, 1.0.0) → pitch`, `(mine_grid, 1.0.0) → arcade`, `(snake, 1.0.0) → arcade`. Unknown
  pairs still resolve nothing.
- **AR-P3 — start precondition generalization.** `StartMinigameAPISession` requires
  `bundle.Pitch != nil` for every minigame (`server/production/minigame_start.go`). That becomes
  "the definition's tenant content resolves through AR-P2". Pitch keeps its present behavior, and
  an arcade start no longer depends on the Pitch artifact.
- **AR-P4 — `minigame_api` rows and API arms.** The `minigame_api` artifact's `tenants` gains
  `{mine_grid,1.0.0,arcade.mine_grid}` and `{snake,1.0.0,arcade.snake}`. These are new artifact
  bytes, minted with AR8. Following MA-C7/MA-C10, the generated API adds two discriminated snapshot
  arms and two command arms. The closed rejection-detail enum adds the AR3.7/AR4.7 details under
  the same `{status, category, detail}` mapping that Pitch's tenant rejections use. All of this is
  additive-only under the v1 compatibility baseline.
- **AR-P5 — client tenant-surface registry** (MA-C9) registers two children keyed by
  `(engine_ref, engine_version)`. It is blocked on the MA3 `minigame_session` surface being
  implemented.

### AR6 — UI surface contract

**AR6.1 Surfaces.** A Game UI surface row `{surface_id:"arcade", mount_id:"game.surface.arcade",
unlock:{kind:"always"}}`. Its closed states are:

- `menu`: the active stage's toys, each with a title, a description, and a Play control.
- `session`: mounts MA3 `minigame_session` with the tenant child.
- `blocked`: another minigame is active; shows copy plus a control that returns to that session.

The Desk gets one entry control, which wears the CRT. The first time the arcade is opened, it shows
exactly **one** contextual hint (`design/11 §2`; `arcade.hint.first_open`), and nothing else
teaches. On mount, the surface calls `GET …/sessions/current` and reopens an active arcade
session. The arcade uses DOM only, with no canvas (`design/06`; Game UI "no canvas in Phase A").
It has no audio in v1.

**AR6.2 `mine_grid` child.**

- **Structure.** A `role="grid"` of buttons with roving tabindex. Each cell's accessible name comes
  from copy keys: position, then hidden / flagged / revealed with its adjacent count.
- **Keyboard.** Arrow keys move. Enter or Space acts in the current mode. `F` toggles a flag, and
  `C` chords.
- **Pointer and touch.** A visible **Reveal/Flag mode toggle** replaces right-click and long-press,
  so no action depends on a pointer-only gesture (Accessibility A2.1). Targets are at least 24 CSS
  px, and larger boards scroll inside a named two-axis container (A3.1's two-dimensional
  exception).
- **Non-color state.** Every state has a digit or glyph plus text; the numbers are never
  hue-coded only.
- **Announcements.** One polite live-region announcement per command response ("N cells
  revealed", outcome). Cascades are never announced cell by cell.
- **No timer.** The certified result has no time, so the grid shows none. A display timer would
  imply a scored clock that does not exist.
- The client renders server snapshots only and never simulates (MA3).

**AR6.3 `snake` child.**

- **Board.** A DOM grid. Steering uses arrow keys, WASD, and a visible four-button D-pad (at least
  24 px; 44 px recommended).
- **Pause.** `P`/`Esc` and a visible Pause button pause the game. It also auto-pauses on
  `visibilitychange`→hidden, window blur, and focus leaving the board. Pausing is fair by
  construction (AR4.8): the server never reads wall time.
- **Local simulation.** The client runs the shared TypeScript engine locally at
  `presentation_tick_ms`. Every 16 ticks it flushes `advance` (provisional, OD-13), and it flushes
  immediately on terminal. It keeps at most one command in flight (MA-C8). If its local lead
  reaches `max_ticks_per_advance` while unacknowledged, it freezes the simulation.
- **Divergence.** A rejected or diverged flush triggers `current` and a resync to the server
  snapshot, with the `arcade.resync.notice` status. The server snapshot is the only truth.
- **Pace.** Where OD-7 is adopted: a local, labeled pace setting (100% / 75% / 50%) scales only
  `presentation_tick_ms`.
- **Screen readers.** Snake is inherently a real-time visual game. It is declared an optional
  toy with a timed-input exception (Deviations D7), and the untimed `mine_grid` beside it is the
  accessible arcade path. The live region announces only food eaten (throttled) and the outcome.

**AR6.4 Reduced motion, CRT effects, photosensitivity.**

- Reduced motion (`prefers-reduced-motion`, and the era_1995 zero-motion budget) removes every
  decorative transition: no tweening between Snake cells, no screen shake, no reveal cascade
  animation, and no explosion or death animation.
- Snake's discrete one-cell steps are the game's content, not decoration, and they remain. Their
  pace control is OD-7.
- The same media-query path drives both CSS and component logic (A4.1). A CSS-only reduction
  fails.
- **CRT effects (OD-8):** v1 ships **no CRT visual effect** beyond a bezel built from the existing
  era border and chrome tokens. Animated effects (flicker, rolling bar, glow pulse, curvature
  warp) are prohibited in `era_1995` by its zero-motion budget, and under reduced motion in every
  era. Any static scanline or curvature treatment requires a UI Foundation theme-schema amendment
  that adds governed tokens, never component literals.
- **Photosensitivity:** no content flashes more than 3 times in any 1-second period, in any mode
  (WCAG 2.3.1), and there are no full-field red flashes.

**AR6.5 Exit interplay copy.** The Desk renders `not_eligible/minigame_session_active` from a Wind
Down or gate-crossing receipt with `arcade.exit_blocked` and a control that opens the running
session. This is a named Game UI consumer amendment.

**AR6.6 Copy keys (mechanical; all prose owner-authored and pending, OD-15).**

- `arcade.container.title`
- `arcade.stage.cover_disc.title`
- `arcade.hint.first_open`
- `arcade.toy.{mine_grid,snake}.{title,description}`
- `arcade.mine_grid.preset.{large,medium,small}`
- `arcade.mine_grid.mode.{reveal,flag}`
- `arcade.mine_grid.cell.{hidden,flagged,revealed}` (params `row`, `column`, `adjacent`)
- `arcade.mine_grid.flags_remaining` (param `count`)
- `arcade.snake.dpad.{up,down,left,right}`
- `arcade.snake.pace.{full,three_quarter,half}` (only if OD-7 is adopted)
- `arcade.action.{play,pause,resume,quit,quit_confirm}`
- `arcade.outcome.{cleared,detonated,crashed,quit}`
- `arcade.blocked.session_active`
- `arcade.exit_blocked`
- `arcade.resync.notice`
- `arcade.error.rejected`

T0 voice follows `design/11 §1b`: the earnest webmaster, warm and never mocking. The era-boundary
banned-word linter applies. Player-facing toy names **must not** use trade names or trade dress
(`design/03 §8` legal rule). "Minesweeper" is a Microsoft title, and the owner names both toys.

### AR7 — Verification corpus (the TP-C25 pattern)

A versioned arcade content-gate corpus is checked in, using a fixture artifact whose extra presets
and Snake dimensions (5×5) make the edge cases reachable. Each scenario is an exact
`(seed, commands, expected terminal or intermediate snapshot)`.

**`mine_grid` scenarios:**

- every preset;
- first-reveal exclusion;
- a zero-cascade flood that stops at flags;
- `chord` success, `chord` detonate, and `chord_unsatisfied`;
- flag toggle on and off;
- a clear-board win;
- `quit` from `setup` and from `playing`;
- every rejection code.

**`snake` scenarios:**

- a crash into each of the four walls;
- self-collision;
- a legal tail chase;
- eat and grow;
- `cleared` on the 5×5 board;
- every rejection code;
- an `advance` whose terminal falls mid-window.

Go and TypeScript byte-compare every snapshot and result. The fixed transition budget equals the
sum of corpus command counts. Content changes regenerate the corpus in a reviewable balance-change
commit.

### AR8 — Delivery and mint

**Fixture-first** (TP-C18 option a). Engines, artifact, amendments, API arms, and surface ship
against fixture bundles. The production mint of `arcade.json`, the two definitions, and the new
`minigame_api` bytes is a later owner-gated epoch (a `BALANCE-CHANGE:` commit under the epoch
protocol), which also carries owner-adopted copy and the OD-3 Fiscal retirement.

Work slices, each reviewed separately:

1. AR-P1–P3 plus the AR-F1 fix (platform-owned; must be red-first);
2. `mine_grid`;
3. `snake`;
4. AR-P4 API arms;
5. AR6 surfaces, after MA3.

## Deviations from design

- **D1 — No Soul restoration** despite `§8`'s "Hobby framing applies (Soul restoration)". This
  follows the later §5 ruling. RP-017 is the owner's to reconcile in `design/03` (DG-5).
- **D2 — No play cooldown** despite "cooldown-gated" (AR1.7). The zero faucet leaves nothing to
  govern (DG-3).
- **D3 — Achievement hooks deferred** (DG-2). The achievements catalog has no condition arm over
  minigame results.
- **D4 — Only the `cover_disc` stage ships.** The Flash portal and the app store are successors
  (DG-4).
- **D5 — CRT is a token-built bezel, not a per-feature skin.** UI Foundation and platform C9 forbid
  per-feature skins (AR6.4, OD-8).
- **D6 — Research, not design, says the arcade teaches that minigames "pay into the main economy"**
  (`research/onboarding-retention.md §2c`). The recommended zero payout follows design's "pure
  nostalgia texture" instead (OD-2, DG-1).
- **D7 — Snake is a timed-input game.** It conflicts with the draft Accessibility A2.2 ("no timed
  key sequence") and is declared an optional-toy exception, mitigated by pause and the local pace
  setting. The untimed `mine_grid` is the accessible arcade path (OD-7).
- **D8 — Unlock is `always`,** not the epoch-6 Fiscal `unlock.arcade` purchase (AR-F2). The
  Fiscal row deviates from design; this RFC follows design.

## Design gaps (DESIGN-GAP — not improvised)

- **DG-1 "The tier-0 tutorial hides here"** (`§8`) vs `design/11 §2` "No tutorial mode". Options:
  - (a) *recommended:* the arcade's tutorial role is the single first-open hint plus the act of
    playing, with no tutorial mechanic;
  - (b) a small payout, so the arcade teaches "minigames pay" (needs harness evidence against the
    `[45,90]`-minute first-Exit envelope);
  - (c) a dedicated tutorial toy (a `pitch.demo`-style variant was already rejected in TP-C6).
- **DG-2 Achievement hooks.** Options:
  - (a) *recommended:* an Achievements successor adds a `minigame_fact_at_least {minigame_id,
    fact_id, minimum}` condition over the `minigame_resolved.v1` event;
  - (b) arcade-local cosmetic badges (a new system; not recommended).
- **DG-3 A play cooldown** for session-skill toys. This would be a platform successor if a paying
  arcade toy ever needs it.
- **DG-4 Later stages** need, first, Game UI eras beyond tier 1 (currently fail-closed) and,
  second, a **server-enforced** tier unlock arm. `fact_equals` is parsed by
  `server/minigame/catalog.go` but never evaluated at start: only `fiscal_unlock` and Soul are
  checked in `minigame_start.go`.
- **DG-5 RP-017** (Soul restoration across hobbies, Go, and the Arcade). This is owner-authored
  design text; the owner reconciles it.
- **DG-6 Host tier.** `design/01 §Tier 0` lists a `CRT` generator at T0, but the shipped
  `generator.crt_wall` is `tier: 1`, costs 2.48832e6, and is unaffordable in the first 15 minutes.
  Options:
  - (a) *recommended:* the host is presentational and the unlock is `always`;
  - (b) unlock on owning `crt_wall` (moves the arcade to late T1 and needs a new server unlock
    arm);
  - (c) retune the generator ladder (a balance mint).

## Findings routed to other owners (evidence; "run it, don't read it" still applies)

- **AR-F1 (Minigame Platform, severity high): a zero-credit resolution fails closed and strands the
  session.**
  - The chain, from reading the code:
    1. `economy.Ledger.ApplyAccrual` omits unchanged balances from `Receipt.Changes`
       (`server/economy/ledger.go`, `if before.Eq(after) { continue }`).
    2. `server/production/minigame_resolution.go` then hits `len(ledgerReceipt.Changes) != 1` and
       returns `(MinigameResolutionDecision{}, ledgerErr)` with a **nil** error.
    3. `server/save/minigame_resolution.go` rejects the empty receipt ("invalid minigame resolution
       receipt"), and the transaction rolls back.
  - Every zero-credit resolution is affected. That includes **Pitch** after `sends_per_day` (5)
    is exhausted on an attended day, and any payout while `company.cash` is at its hardcap. The
    claimed session cannot resolve, so MA-C12 then rejects Exit.
  - This has not yet been executed. The fix must land red-first: a failing composed test for a
    6th Pitch send, then the fix. This RFC's AC8 depends on it.
- **AR-F2 (Fiscal / First Content Epoch): `{"unlock_id":"unlock.arcade","cost":5}`** is minted in
  `balance/fiscal/first-content.json`. It originates from the fiscal-foundation fixture
  (`53d1b4a`) and contradicts `design/03`'s free T0–1 arcade. It is currently unbound and
  unpurchasable (there is no Fiscal player surface).
- **AR-F3 (Game UI / Curriculum):** MA-C12's Exit rejection also blocks the `cross_gate` that
  triggers the scripted first ending. Pitch (T5) never met this; a T0 toy does. AR1.8's quit
  command and Desk copy mitigate it. Whether the scripted ending should instead auto-quit an
  active session is a platform/curriculum question this RFC does not decide.

## Acceptance criteria (each has a demonstrated failing case)

1. **Artifact loader.** Go and TypeScript accept the fixture and the v1 `arcade.json`
   byte-identically. **Fails on:** unsorted presets, `mines > width*height−9`, a stage toy with no
   definition or a wrong `engine_ref`, an extra key, or a non-ascending `min_tier`. Each of these
   must reject in both runtimes.
2. **Definitions.** Both rows load under schema v3. A policy test asserts that every minted arcade
   row has `conversion_ppm = 0` and an inert quality curve. **Fails on:** a fixture with nonzero
   conversion fails the test, and a `power` scaling row fails load.
3. **`mine_grid` rules and parity.** The AR7 corpus is byte-identical across Go and TypeScript.
   **Fails on:** each of these seeded mutants turns it red:
   - a 4-neighborhood flood;
   - placement that ignores the first-reveal exclusion;
   - a chord that reveals flagged cells;
   - a TypeScript-only `Bound` without rejection sampling.
4. **Hidden information.** Every non-terminal snapshot in the corpus, and every non-terminal API
   response in the composed test, has `mine_cells: []`. No stored session state contains mine
   positions. **Fails on:** a mutant that stores or exposes mines early.
5. **`snake` rules and parity.** The AR7 corpus is byte-identical across Go and TypeScript, and the
   client's local predictor is the same TypeScript module. **Fails on:** each of these mutants:
   - tail-chase counted as a collision;
   - food drawn from all cells instead of empty cells;
   - a reversal accepted;
   - turns after the terminal tick silently discarded.
6. **Rejection without mutation.** Every AR3.7/AR4.7 code is reached. Revision, command rows, and
   the snapshot are unchanged. **Fails on:** a mutant applying a partial `advance` before
   rejecting.
7. **Platform amendments.** The resolver returns arcade bytes for exactly the two new pairs, and
   Pitch is unchanged. An arcade start succeeds in a bundle without the Pitch artifact but with
   `arcade`. **Fails on:** removing `arcade` from the bundle makes the arcade start reject, and
   reverting AR-P3 makes the Pitch-less bundle reject.
8. **Full composed path.** For both toys, over a real socket: create, commands, then terminal
   auto-resolution returning a receipt with credited `"0"`, no cap reason, and unchanged
   rating/quality. A retry returns identical bytes. **Fails on:** the pre-fix AR-F1 path, shown red
   first.
9. **Exit interplay.** With an active arcade session, Wind Down and `cross_gate` reject
   `minigame_session_active`. After `quit`, both succeed. **Fails on:** a build without `quit`
   leaves no player-reachable terminal from `setup`, and the test demonstrates that.
10. **Soul gate.** Near-zero Soul rejects the start `human_content_locked`. **Fails on:**
    classifying the toy `unrelated` passes the start, so the test turns red.
11. **API.** The generated schema passes the additive-only compatibility baseline, and the new
    details map exactly. **Fails on:** removing or renaming an existing arm breaks
    `make api-check`.
12. **Surface and accessibility.** axe WCAG 2.2 AA shows zero serious/critical findings in
    Chromium, Firefox, and WebKit for `menu`, `mine_grid` (setup/playing/terminal), and `snake`
    (playing/paused/terminal). A keyboard-only browser test completes a `mine_grid` game and a
    Snake quit. Reduced-motion tests assert zero decorative durations, no inter-cell tween, and
    live media-query switching. **Fails on:** a CSS-only reduced-motion mutant, a tween mutant, a
    pointer-only flag action, and a hue-only cell state.
13. **Photosensitivity and CRT.** No arcade component defines an animation in `era_1995`, and a
    static check rejects component-literal CRT styles. **Fails on:** a seeded flicker keyframe.
14. **Copy.** Every AR6.6 key resolves with owner-adopted text, and the denylist contains the
    trade names. **Fails on:** a missing key, and a seeded "Minesweeper" string rejected by the
    linter.
15. **Fixture-first boundary.** The live epoch manifest does not pin `arcade` until the AR8 mint.
    **Fails on:** a test that pins it without the mint record.

## Owner decisions (numbered; recommended defaults)

1. **OD-1 — v1 roster.** *Recommended:* `mine_grid` + `snake`, both named in `§8`. Alternatives:
   `mine_grid` only (Snake deferred with OD-5–7); or also the platformer (not recommended: a
   realtime platformer is a much larger engine).
2. **OD-2 — Economy.** *Recommended:* zero payout and inert quality ("pure nostalgia texture").
   Alternative: a small `company.cash` faucet, which requires harness evidence against the T0–T1
   envelopes.
3. **OD-3 — Unlock.** *Recommended:* `always`, and retire Fiscal `unlock.arcade` in the arcade mint
   epoch. Alternative: keep the 5-IC purchase (contradicts `design/03`).
4. **OD-4 — Soul.** *Recommended:* `human_hobby` (lock only, no restoration) pending RP-017.
   Alternative: `unrelated`.
5. **OD-5 — Snake transport.** *Recommended:* batched `advance` with the published `food_seed`.
   Alternatives: server-revealed food only, with the client flushing on every eat (visible
   rubber-banding at the round trip); or defer Snake.
6. **OD-6 — Snake wall time.** *Recommended:* no pacing validation while payout, boards, and ghosts
   are all absent. Any of those later requires a platform pacing contract.
7. **OD-7 — Snake accessibility.** *Recommended:* ship as an optional toy with pause (always), a
   local pace setting (100/75/50%), and the D7 exception. Alternatives: no pace setting; or defer
   Snake until the Accessibility RFC rules.
8. **OD-8 — CRT visuals.** *Recommended:* only the token bezel in v1. Alternative: a UI Foundation
   theme amendment for static scanline tokens. Animated CRT effects stay prohibited either way.
9. **OD-9 — Tutorial role (DG-1).** *Recommended:* option (a).
10. **OD-10 — ORDER NOW placement.** *Recommended:* it stays on the Desk where Game UI shipped it.
    The platformer successor owns the in-arcade "demo ends → ORDER NOW" frame.
11. **OD-11 — Best score.** *Recommended:* a local-only display record (the gate-splits precedent),
    labeled as local. Alternative: none.
12. **OD-12 — `mine_grid` literals.** *Recommended:* presets 9×9/10, 16×16/40, 30×16/99;
    first-reveal 3×3 exclusion; chord included. All provisional.
13. **OD-13 — `snake` literals.** *Recommended:* 20×20 with walls, start 3, growth 1, 125 ms, 64
    ticks per advance, flush every 16 ticks. All provisional.
14. **OD-14 — Delivery.** *Recommended:* fixture-first, with the production mint in a later
    owner-gated epoch.
15. **OD-15 — Copy.** The owner authors every AR6.6 string and both toy names. No trade names.
16. **OD-16 — Platform open question.** *Recommended:* close the platform RFC's "lightweight
    score-send path" question as **rejected** (it is a client-sent result), citing this RFC.

## Open questions

- AR-F3: should the scripted first ending auto-quit an active minigame session instead of
  rejecting? That is a platform/curriculum ruling. This RFC ships the quit-plus-copy mitigation
  regardless.
- Future stages, ghosts, and daily slots are named successors (DG-3, DG-4, OD-6). Nothing here
  pre-builds them.

## Changelog

- 2026-09-25: created (draft) — container, `mine_grid`, and `snake` on the platform, following the
  Pitch template; findings AR-F1–AR-F3 filed.

## Owner acceptance (2026-09-25)

Marco accepted this RFC in the 2026-09-25 batch (the answer to a direct batch question in-session,
recorded by Claude). Every owner decision this RFC lists is **ruled at its stated recommended
default**. Where the text offers alternatives, the recommended option is the normative one and the
alternatives are rejected. Provisional numbers stay provisional and are ratified by harness
measurement and SHA, as the RFC already requires. Owner-authored copy stays owner-authored:
implementation ships copy keys, with any placeholder or candidate text clearly marked.
