# Demo Disc Arcade

Source RFC: `rfc/minigame-demo-disc-arcade.md` (accepted 2026-09-25). **Fixture-first:** no epoch
pins the `arcade` artifact yet, and the public API arms (AR-P4) and client tenant registration
(AR-P5) are blocked on the owner's ruling about extending v1 minigame unions (the TT-PA4 conflict
with API Foundation C2).

## The pinned artifact

`arcade` (`schema_version: 1`) has exact keys `{schema_version, container, mine_grid, snake}` and
loads through `server/arcade/catalog.go` and `client/src/arcade/catalog.ts`.

- **Stages:** `container.stages` are sorted by strictly ascending `min_tier` (0–9) and carry
  byte-sorted toy IDs. The client shows the stage with the greatest `min_tier` at or below the
  authoritative tier. Stages are presentation only; start permission comes from each toy's
  `unlock_condition`.
- **`mine_grid` presets:** width and height are 5–30, and mines are at least 1 and at most
  `width*height − 9`. The copy key must be `arcade.mine_grid.preset.<preset_id>`.
- **`snake`:** width and height are 5–30, `start_length` is 2 to `width/2`, `growth_per_food` is
  1–8, and `max_ticks_per_advance` is 1–256. `presentation_tick_ms` (50–1000) is presentation-only;
  the engine never reads it.

The candidate bytes are `balance/testdata/arcade-v1.json` (the RFC's provisional v1 rows). The
corpus fixture is `testdata/arcade/corpus-fixture-v1.json`: small boards, and a 6×5 Snake board,
because a 5×5 board has no Hamiltonian cycle and so cannot witness `cleared`.

## Engines

Both engines are pure `minigame.Tenant`s. They are solo only, take the literal scaling breadth
`1` at `minigame.arcade.<toy>`, return `rating_delta: null`, and have exact snapshot key sets that
carry `arcade_content_hash` and `arcade_schema_version`.

- **`mine_grid` 1.0.0 (AR3):**
  - Mines are placed lazily at the first reveal, excluding the first cell and its neighbours, by a
    descending Fisher–Yates over the eligible cells.
  - Placement is re-derived on every `Apply` and never stored. `mine_cells` is `[]` and
    `exploded_cell` is `-1` in every non-terminal snapshot, and a snapshot that violates this is
    invalid.
  - Commands: `choose_board`, `reveal` (flood fill that stops at flags), `toggle_flag`, `chord`,
    `quit`.
  - Facts: `mine_grid.cells_revealed` and `mine_grid.cleared`.
- **`snake` 1.0.0 (AR4):**
  - Movement is a deterministic tick simulation. `food_seed` is published as a decimal string.
  - The only commands are `advance {through_tick, turns}` and `quit`. A command whose terminal tick
    falls before `through_tick` rejects `advance_past_terminal` without mutation.
  - Facts: `snake.food_eaten` and `snake.ticks_survived`.

## Verification

`make arcade-corpus-check` regenerates `testdata/arcade/content-gate-v1.json` from the Go engines
and compares it byte for byte. The corpus has 16 scenarios and 43 applied transitions. Between
them they cover every preset, first-reveal safety, a flood that stops at a flag, chord
success/detonate/unsatisfied, a clear-board win, all four walls, self-collision, a legal tail
chase, a `cleared` board, and every rejection code of both engines.
`client/test/arcade-content-gate.test.ts` replays it byte for byte in TypeScript, and checks that
every rejection leaves the snapshot unchanged.
