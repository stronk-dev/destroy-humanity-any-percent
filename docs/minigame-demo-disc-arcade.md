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
which supports the current cycle-based clearing driver. That driver does not supply AR7's
specified 5×5 clearing population (RP-196); lacking a cycle is not proof that a different
legal command trace cannot clear 5×5. The exact odd-board witness remains missing.

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

Snake snapshot decoding validates all nine numeric scalar fields and every body element as
nonnullable integers in the shared safe-integer range. Content identity and `food_seed` are
strings, not values coerced to strings. Both decoders refuse duplicate keys, decimal/exponent
integer tokens and malformed field types before transition. Valid whitespace, escapes,
reordered keys and integer negative-zero spelling remain usable. Movement, food placement,
growth, clocks, descriptor and snapshot field sets are unchanged (kernel 0.3.146).

`testdata/arcade/snake-snapshot-negatives-v1.json` supplies 45 shared raw negatives and seven
legal controls over actual genesis, moved/grown playing and cleared terminal states. Go and
TS execute their real engines to construct each source and match the literal corpus bytes;
both direct decoders and direct apply entries refuse the mutated inputs before transition.
Legal spellings preserve canonical snapshot/result execution; the largest safe tick remains
usable and produces its exact certified fact. These are grammar controls, not proof of all
semantic reachability invariants or a public saved-state exploit. Claude review remains required.

## Verification

`make arcade-corpus-check` regenerates `testdata/arcade/content-gate-v2.json` from the Go engines
and compares it byte for byte. The corpus has 16 scenarios and a fixed budget of 60 attempted
commands: 43 applied and 17 rejected. Go generation and TS replay both count every attempt;
independent checks reject an applied-only budget. Between
them they cover every preset, first-reveal safety, a flood that stops at a flag, chord
success/detonate/unsatisfied, a clear-board win, all four walls, self-collision, a legal tail
chase, a `cleared` board, and every rejection code of both engines.
`client/test/arcade-content-gate.test.ts` compares literal Go snapshot bytes at every genesis and
after every attempted command, plus literal serialized result bytes (including `null`) after
each attempt: 76 snapshot observations. Rejections preserve snapshot, result and revision.
Readable JSON witnesses accompany the literal strings; the primary comparison does not
reconstruct bytes from parsed fixture objects. Go captures own their snapshot/result memory,
and missing witnesses or stale generated bytes fail the generation gate. Independent wrong
intermediate-result and reordered-snapshot probes fail the new comparison. The historical v1
artifact is retained, not actively regenerated or replayed; removing v2's added witness fields
and resetting its version recovers the exact v1 population and gameplay metadata. This closes
RP-194's instrument gap locally, not the full AR7/designated review or public-route acceptance.
The Mine Grid hidden-information witness also executes all eight Mine Grid scenarios and
inspects raw output at genesis and after every attempted command (44 observations). Setup,
unplaced/placed playing, and unplaced/placed terminal states must actually occur. Nonterminal
outputs contain no mine positions or explosion; placed terminals disclose the derived mines.
For each placed nonterminal state, independently forged mine lists and explosions must be
refused by both the decoder and engine entry, while clean replay still reaches its corpus
terminal. This is engine-level evidence, not public-route or assistive-technology acceptance.
`testdata/arcade/snapshot-value-negatives-v1.json` supplies 19 shared value/row mutations of
real unplaced, placed and terminal Mine Grid snapshots. Go and TS decoders and engine entries
refuse malformed numeric fields, cell lists and revealed rows; clean controls remain usable.
The TS decoder validates integer values and exact revealed-row keys before transition. These
cases are not proof of complete raw JSON token/duplicate-key or semantic-state parity.
The separate raw-snapshot population in `testdata/arcade/snapshot-raw-negatives-v1.json`
checks 25 malformed inputs and six legal spelling/boundary controls against actual genesis,
unplaced, placed and terminal states. Both Mine Grid decoders reject duplicate (including
escaped-equivalent) keys, decimal/exponent integer tokens, null numeric fields/elements and
missing revealed-row fields. Numeric fields fit the shared safe-integer range. Valid whitespace,
escapes, reordered fields, negative-zero integer tokens and coordinate boundaries still work.
The TS scanner rejects compounds deeper than the declared root/list/row shape before recursing;
the test demonstrates malformed nesting cannot cause stack exhaustion. These are decoder/direct
engine proofs, not public-route/storage acceptance or proof of every semantic state.

## Platform chain

`CatalogBundle.Arcade` joins the replay bundle. The `arcade` artifact requires `minigame_api`, and
`validArcadeChain` (Go) and the TS replay loader enforce three things:
- arcade-engine definition rows exist exactly when the artifact does;
- each such row is a `minigame_api` tenant at engine version `1.0.0`;
- every stage toy resolves to one of those rows.

`CatalogBundle.TenantContent` resolves both `(mine_grid, 1.0.0)` and `(snake, 1.0.0)` to the one
arcade artifact, and the gameserver registers both tenants.

The fixture rows are `testdata/minigame/pitch-typer-arcade-v3.json` and
`balance/testdata/minigame-api-arcade-candidate-v1.json`. They carry AR2 verbatim: zero faucet,
inert quality, neutral rating, `always`, `human_hobby`.
`TestArcadeComposedIntegrationUnlockLockPlayAndZeroCreditResolution` witnesses on Postgres:
- the near-zero-Soul lock;
- Tier-0 play of both toys;
- zero-credit applied resolutions that leave cash, rating and quality unchanged;
- identical-bytes retries and verified Founder history;
- release of the Exit block.

## Client toys (test-only mount)

`client/src/game-ui/minigame/MineGridBoard.svelte` (AR6.2) renders server snapshots only:
- a `role="grid"` of buttons with roving tabindex;
- arrows move, Enter/Space act in the visible Reveal/Flag mode, `F` flags, `C` chords;
- cell names come from copy keys, with glyph text rather than colour;
- one polite live-region message per response;
- a quit confirmation;
- no timer.

`SnakeBoard.svelte` (AR6.3) runs the shared TS engine locally at `presentation_tick_ms` divided by
the chosen pace. It flushes `advance` every 16 ticks and immediately at a terminal tick, with one
command in flight, and freezes at `max_ticks_per_advance` while unacknowledged. A rejected or
mismatched flush resyncs from `current()`. It auto-pauses on hidden, blur and focus loss, pauses
on `P`/`Esc`, starts paused, steers by arrows, WASD or a D-pad, and has no decorative animation.

Both components pass axe in three browsers. They are not in the tenant registry: no pinned
`minigame_api` artifact carries the arcade tenants, and the public wire is blocked (see the top of
this page).

The Snake batching/blur/resync and lead-limit browser witnesses control timeout delivery, not
the game or engine. Wall-clock movement alone must leave the board unchanged; the exact batch
is required on callback four, blur must stop further movement/submission, rejection must
restore the server head/tick, and an unacknowledged command must freeze at the advance limit.
The D-pad/terminal witness retains real browser timers. This replaces fixed wall-delay guesses
in the two scheduling witnesses (RP-188), without changing the runtime scheduler or pace.

Native keyboard witnesses additionally choose Small with Enter, traverse the Mine Grid with
arrows and clear the existing seed-202 corpus board with Space, updating the same mounted
child from the engine response. Snake's Quit is activated by Enter and must reach the real
engine terminal. Each witness fails independently when its action handler is severed (RP-189).
These are component/engine checks with controlled initial focus, not public Arcade integration,
whole-task Tab navigation or participant assistive-technology evidence.
