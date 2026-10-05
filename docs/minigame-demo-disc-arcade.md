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

Stage `min_tier` is a nonnullable integer in both loaders. Go checks a pointer-decoded raw
row before decoding its integer field, so null cannot silently become zero. Legal zero,
integer `-0` and surrounding whitespace remain accepted. This loader-only correction is
kernel 0.3.147; no content, mechanics or wire fields changed.

The candidate bytes are `balance/testdata/arcade-v1.json` (the RFC's provisional v1 rows). The
corpus fixture is `testdata/arcade/corpus-fixture-v1.json`: small boards, and a 6×5 Snake board,
which supports the current cycle-based clearing driver. That driver does not supply AR7's
specified 5×5 clearing population by itself; lacking a cycle is not proof that a different
legal command trace cannot clear 5×5. RP-196 adds a separate exact 5×5 fixture and witness
without replacing or relabelling the 6×5 corpus.

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
growth, clocks, descriptor and snapshot field sets are unchanged (decoder correction introduced
in kernel 0.3.146; current kernel 0.3.147 includes the separate catalog null-tier fix).

`testdata/arcade/snake-snapshot-negatives-v1.json` supplies 45 shared raw negatives and seven
legal controls over actual genesis, moved/grown playing and cleared terminal states. Go and
TS execute their real engines to construct each source and match the literal corpus bytes;
both direct decoders and direct apply entries refuse the mutated inputs before transition.
Legal spellings preserve canonical snapshot/result execution; the largest safe tick remains
usable and produces its exact certified fact. These are grammar controls, not proof of all
semantic reachability invariants or a public saved-state exploit. Claude review remains required.

## Verification

The actual Go/TS loaders execute matched 22-case negative populations, including extra keys,
unsorted presets, mine bounds and equal/descending stage tiers, with valid candidate/fixture,
inclusive bounds and ascending test-only-stage controls. The real TS stage selector observes
tiers 8/9 on the structural control; no later-stage content or gameplay is minted. Both bundle
loaders recompute full hashes and reject missing toys and wrong-engine definitions. Independent
tier/preset/cap/key/binding policy severings fail; probes are restored. This bounded RP-199
evidence repair awaits Claude review, not complete A1/A4 acceptance or raw JSON grammar parity.
RP-200's separate loader correction refuses null in candidate and fixture and through actual
freshly hashed complete Go/TS bundle loaders. Legal zero spellings still load. Tests fail
before the fix and again when only the null guard is severed; that probe is restored.
The correction awaits Claude's designated review, not full raw JSON parity or archival.

`mine-grid-sampling-v1.json` supplies a real rejected-draw population on the unchanged large
9×9 fixture. Inverting SplitMix64 and its two published substream labels constructs seed
`15581846558861750132`; both actual RNGs verify its first draw is zero, below threshold 16
at eligible-cell bound 72, then consume the accepted second draw. Go records actual registry
create/choose/reveal/quit and terminal retry bytes (four attempts / five states); TS independently
executes its placement and every apply step against those literals. A Mine Grid-only modulo
shuffle changes the actual mine list and first-reveal bytes, and both comparisons fail.
Threshold removal and a forged artifact also fail. Root Arcade generation/check aliases
include this witness without changing the older corpora, fixture, mechanics or production
content. This bounded RP-198 evidence repair awaits Claude review; it does not establish all
A2/AR7 or public integration. The existing flood/exclusion/chord witnesses already catch their
required gameplay mutants independently in both runtimes and are preserved.

`snake-rejection-atomicity-v1.json` supplies six actual commands at both genesis and after an
accepted move: terminal overshoot, a turn after death, late same/opposite turns, an empty window
and nonascending turns. Go registry and independent TS apply reject the exact codes; direct
transitions preserve their actual caller-owned objects. Successful advance, quit and terminal
refusal remain usable. The pure population uses the separate 5×5 fixture, not the original 6×5.
Future-turn and partial-input mutation probes fail these comparisons.

`TestArcadeRejectedAdvanceIntegrationPreservesPersistedSessionAndCommands` submits the same
twelve negatives through actual Service Play against real Postgres and the 20×20 candidate.
It compares saved state/genesis/result/revision and the complete ordered SQL command rows,
then checks active status and released claims. UpdatedAt is excluded because legitimate claim
acquisition/release changes it. One accepted move and quit/resolution commit exactly two rows;
rejections commit none. Rejected-persistence and stuck-claim probes fail independently. This is
DB/library composition, not public API receipts, sockets or full AC8 acceptance. Claude review
remains mandatory for this bounded RP-197 test-only correction; production rules are unchanged.

`testdata/arcade/snake-5x5-fixture-v1.json` differs from the original corpus fixture only in
Snake width (6→5). `snake-5x5-gate-v1.json` records all 1,024 Go registry-driven strategy
observations and the selected real clearing trace. The strategy follows a 24-cell cycle
excluding cell 0, entering that cell only at the legal final growth step. It clears seed 455
at tick 134, with score 24 and all 25 body cells. The other 1,023 observations are explicitly
excluded-food strategy failures, not claims those seeds or the board are unwinnable. Largest
observed tick is 209; guard exhaustion invalidates measurement instead of excluding a seed.

The selected Go trace records 26 attempted commands and 27 literal snapshot observations,
including nonterminal food/growth outputs, an overlong terminal-window refusal that preserves
state/result, exact terminal and post-terminal phase refusal. TypeScript independently
byte-compares actual genesis and every attempted snapshot/result. Validators refuse the
6×5 substitute, incomplete bodies and fabricated outcome/facts; independently severed final
exit and intermediate output probes fail. This is the named 5×5 clearing population, not
all A3/AR7, public-wire or archival acceptance. Claude review remains mandatory.

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
