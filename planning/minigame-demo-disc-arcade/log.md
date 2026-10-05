# Demo Disc Arcade — implementation log (append-only)

## 2026-09-25 — Predeclaration (Claude)

**Implemented by:** Claude (the kernel-path lane after Clout `527246f1`). Awaiting Codex's
designated review; never self-approved or archived.

Scope is `plan.md` A1–A7: everything below the public wire. AR-P4 (API arms) and AR-P5 (client
registry) stay blocked on the same owner ruling as TT-PA4, because extending the existing v1 minigame
unions contradicts API Foundation C2. Not redone: AR-F1 (fixed in `8add475`), and F9/AR-F3 (Wind Down
preview fixed in `2def3612`; the scripted-collapse freeze is logged for Game UI/Curriculum).

Protocol:
- Every commit touching a guarded path bumps `kernel/VERSION` in the same commit. The new
  `server/arcade/` and `client/src/arcade/` directories are registered when they are created.
- Red-first tests, and a severing probe for each gate.
- Cold `-count=1` runs.
- Fixture-first, with no mint.

## 2026-09-25 — A1–A3: artifact, loaders, both engines, and the shared corpus (Claude)

**Implemented by:** Claude. Awaiting Codex's designated review.

- **Artifact and loaders (A1):** Go (`server/arcade/catalog.go`) and TS
  (`client/src/arcade/catalog.ts`) enforce AR1.2/AR3.1/AR4.1 exactly. The candidate bytes are
  `balance/testdata/arcade-v1.json`; the AR6.6 copy keys are candidate text in
  `copy/catalog/arcade-candidate.json`. Both toy titles are marked `PENDING OWNER NAME` (OD-15),
  with no trade names.
- **Engines (A2/A3):** `mine_grid` and `snake` exist in both runtimes, with the AR3.7/AR4.7 closed
  taxonomies. A Go-generated corpus (`make arcade-corpus`, 16 scenarios, 43 transitions) replays
  byte for byte in TS.
- **Deviation (recorded):** the corpus fixture's Snake board is 6×5, not the RFC's 5×5. A 5×5 grid
  has no Hamiltonian cycle, so a deterministic `cleared` witness is impossible there.
- **Decoding classification, fixed identically in both runtimes:**
  - a malformed `choose_board` is `unknown_preset`;
  - a malformed cell command is `cell_out_of_range`;
  - a malformed `advance`, turn or tick is `advance_window`, but an unknown direction string is
    `invalid_turn`;
  - an unknown command kind, or a malformed `quit`, is `illegal_phase`.
- **Kernel:** 0.3.128 → 0.3.129 in this commit, and `server/arcade/` and `client/src/arcade/` are
  registered in `kernel/affecting-paths.json`.

**Evidence (cold):**
- `make test-go GO_PACKAGES='./arcade ./kernel'` and `make arcade-corpus-check` pass.
- TS `arcade-content-gate.test.ts` passes 6/6. `tsc` and `copy-check` pass.

**Severing:** each probe was restored afterwards.
- G1: letting the flood reveal flagged cells fails the Go corpus gate.
- G2: removing the hidden-information check fails `TestMineGridNeverExposesMinesBeforeTerminal`.
- G3: removing `advance_past_terminal` fails both the corpus and the dedicated Snake test.
- T1: changing TS's food substream label fails the replay.
- T2: a TS tail that never leaves fails it.
- T3: TS without zero cascades fails it.

## 2026-09-25 — A4/A5: platform chain, resolver, composition, and a Postgres witness (Claude)

**Implemented by:** Claude. Awaiting Codex's designated review.

**AR-P1:**
- `CatalogBundle.Arcade` is added. `replaycatalog.Load` loads the `arcade` artifact, which requires
  `minigame_api`; the name is added to `validArtifactNames`.
- `validArcadeChain` enforces:
  - arcade-engine definitions (`mine_grid` or `snake`) exist exactly when the artifact does;
  - they pin engine version `1.0.0`;
  - each is a `minigame_api` tenant;
  - every stage toy resolves to one.
- The TS replay loader mirrors the chain.

**AR-P2:** `CatalogBundle.TenantContent` resolves `(mine_grid|snake, 1.0.0)` to the single
pinned arcade artifact. Unknown pairs still resolve nothing.

**AR-P3:** already generalized by TT-PA3, and reused as is. An arcade start needs only its own
tenant content.

**Composition:** the gameserver registers `arcade.NewMineGridTenant()` and
`arcade.NewSnakeTenant()`.

**Fixtures:**
- `testdata/minigame/pitch-typer-arcade-v3.json` holds the two AR2 rows verbatim: zero faucet,
  inert quality, neutral rating, `always`, `human_hobby`.
- `balance/testdata/minigame-api-arcade-candidate-v1.json` adds the two tenant rows.
- Neither is production bytes (AR8).

**Kernel:** 0.3.129 → 0.3.130 in this commit.

**Evidence (cold):**
- `make test-go` passes for arcade, production, replaycatalog, gameserver, minigame and kernel.
- Postgres (`compose.save-test.yml`) passes for production, minigame, gameserver and
  replaycatalog.
- `typecheck` passes, and `test-client` passes 6861.
- TS `replay.test.ts` passes 87/87.

**A5 witness:** `TestArcadeComposedIntegrationUnlockLockPlayAndZeroCreditResolution` covers:
- near-zero Soul (5) rejecting start with the shipped `human_content_locked`;
- a Tier-0, no-Exit Founder starting both toys (the `always` unlock);
- `mine_grid` choose → reveal → quit, and `snake` crashing into the wall at tick 10;
- each resolution applied with `credited_delta: "0"` and zero forfeit;
- cash, rating (Elo 1000) and offline quality (200000 ppm) unchanged;
- an identical-bytes retry;
- Founder history `ReplayVerified`;
- the Exit block held during the session and released on resolution.

This end-to-end witness confirms AR-F1's zero-credit fix (`8add475`) for a zero-rate faucet.

**Severing** (each restored):
- R1: skipping the tenant check fails "definitions without tenants".
- R2: skipping the stage-toy check fails "stage toy without definition".
- T4: the TS stage-toy check severed fails the TS chain test.
- P1: removing the resolver arm makes the start fail (`minigame tenant returned invalid output`).
- P2: a paying fixture row makes the zero-credit assertion fail (`credited_delta: "5.1e1"`).
- R3 survives, and is logged as redundant rather than vacuous: `CatalogBundle.valid`'s arcade arm
  duplicates `validArtifactNames`, which already rejects an arcade without `minigame_api` before
  `valid` runs.

## 2026-09-25 — A6/A7: client toy boards and docs (Claude)

**Implemented by:** Claude. Awaiting Codex's designated review.

`MineGridBoard.svelte` (AR6.2) and `SnakeBoard.svelte` (AR6.3/6.4) live under the unguarded
`client/src/game-ui/minigame/`. They are mounted only in tests: no pinned `minigame_api` artifact
carries arcade tenants yet, and AR-P5 registration is blocked with the API arms.

`SnakeBoard` takes `submit`/`current` callbacks from its host, so it never touches transport.

**Candidate copy gap (logged):** AR6.6 has no key for the per-response "N cells revealed"
announcement, the visible glyphs, the board labels, the pause state, or the pace label. These are
added as candidate keys and await owner adoption:
- `arcade.mine_grid.cells_open_frame`
- `arcade.mine_grid.glyph.{flag,mine}`
- `arcade.{mine_grid,snake}.board_label`
- `arcade.snake.{score_frame,paused,pace_label}`

The mine glyph is `X` because the copy linter treats `*` as Markdown.

**Evidence (cold):**
- `test/arcade-boards-browser.test.ts` passes 12/12 across chromium, firefox and webkit, twice. It
  covers:
  - axe checks on setup, playing, terminal, snake ready and snake terminal;
  - the roving tabindex, and the F/Arrow/C keys and mode toggle producing exact commands;
  - no mine leak before terminal;
  - quit confirmation;
  - D-pad steering flushing exactly `{through_tick:3, turns:[{tick:1,"up"}]}`, validated by the
    real TS engine;
  - `flushEvery`, blur auto-pause, resync on a rejected flush, and the max-lead freeze.
- The Snake tests are real-time (20–40 ms ticks). Timing margins are logged as a flake risk, not a
  proven stability claim.
- Full lanes pass:
  - `make typecheck build-client test-client verify-client-boundary copy-check test-browser`
    (20772 tests; boundary scan now 19 files);
  - `make test-game-ui-composed`, both the Pitch phase and the v4 lifecycle.

**Severing (chromium, 1 failed | 3 passed each; restored):**
- B1: no freeze.
- B2: no blur auto-pause.
- B3: no roving tabindex.
- B4: no mine glyph.

**Not built (blocked or owner-gated):**
- AR-P4 API arms and AR-P5 registration (the owner's TT-PA4/C2 ruling);
- the AR6.1 arcade surface host and AR6.5 Desk copy, which need the public wire;
- the OD-3 Fiscal `unlock.arcade` retirement and all production bytes (AR8 mint);
- toy names and all prose (OD-15).

## 2026-10-05 — Cold full-browser Snake timing failure (RP-188)

**Observed / recorded by:** Codex. The restored Cosmetic motion correction's full Linux
`make test-browser-ci` failed two cases: Chromium Typer's update-depth error (RP-187) and
WebKit Snake's batching/blur/resync assertion. Snake observed no submitted command after
the fixture's fixed `settle(190)` wait; expected `{kind:"advance", through_tick:4, turns:[]}`.
The run exited Make 2 with 21,097 passing, 2 failed and 3 skipped; the separate performance
case was not reached. This is current red CI-equivalent evidence, not an unrelated-case waiver.

The unchanged case schedules four 40 ms callbacks but treats 190 wall milliseconds as proof
they all ran. That is a candidate instrument assumption, not a verified runtime cause. Before
any correction, isolate the exact scheduled callbacks and actual engine/server outcomes under
accepted AR6.3. Do not raise sleeps/timeouts, add automatic retries, weaken expected commands,
or claim full browser green from a selected passing rerun. No Arcade code changed here.

## 2026-10-05 — RP-188 callback-population predeclaration (Codex)

**Authority:** accepted AR6.3; test/instrument correction only unless a separately recorded
production defect is demonstrated. The prior WebKit failure and unchanged passing rerun are
both retained. No runtime scheduler, pace, batching threshold, content, copy or CI change.

**Question:** does the assertion confuse elapsed wall time with delivery of four scheduled
callbacks? Mount the real child with the real TS engine, controlling only test timeout delivery.
Move the fixture wall clock without delivering timers: zero submitted advances is required.
Then deliver 39 + 1 ms per scheduled tick and verify the exact first advance only on tick 4.
Use DOM head positions and engine/server revisions as independent observations, not a counter
inside a replacement game. Exercise blur over further timer delivery, rejection/resync to the
unchanged server snapshot, and the max-lead/unacknowledged freeze. Retain a real-timer terminal
smoke path. Browser population: Chromium, Firefox and WebKit in the declared Linux CI image.

**Exit criterion:** all named callback/engine assertions pass, with no larger sleeps/timeouts,
retries or weaker commands. A missing batch flush, missing blur pause and missing lead freeze
must independently fail their named populations; restore every probe before broad verification.
If timer control cannot safely coexist with real async engine work/browser tooling, record the
failure and change the instrument explicitly, not the outcome. The experiment can establish
an invalid timing oracle and a test-only correction; it cannot establish that this was the
sole cause of the earlier hosted/local failure or certify CI reliability from selected reruns.
Finish with type/unit/build/boundaries and the full cold browser target, including performance.
Hand off the exact corrective range for Claude review; no archival or full-RFC promotion.

## 2026-10-05 — RP-188 instrument diagnosis and test-only correction (Codex)

**Review by:** Codex. **Recorded by:** Codex. **Original bounded range:** Claude
`fca062a1^..fca062a1`, A6 scheduling evidence only. **Verdict: CHANGES REQUIRED** on the
fixed wall-delay instrument, not a demonstrated runtime batching defect or full A6 verdict.

The retained three-engine population moves the wall clock 190 ms without delivering callbacks:
no advance is submitted and the DOM head stays at genesis. It then delivers each scheduled
40 ms tick as 39 + 1 ms: the head advances once per callback, nothing flushes before tick 4,
and the complete command is exactly `{kind:"advance", through_tick:4, turns:[]}` at tick 4.
The real engine acknowledges tick 4. Blur then prevents both movement and submission over
further delivered time. A rejected tick-8 flush fetches current exactly once, restoring the DOM
head and server tick 4 with the resync notice. The separate unacknowledged population observes
ticks 0→4→5 and the exact singleton tick-5 command, then a frozen head despite further callbacks.
Only test timeout/Date delivery is controlled; TS engine/hash work and mounted DOM are real.
Each test unmounts before restoring timers. The D-pad/terminal smoke retains native timers.

**Executed discrimination (each independent; restored):**

- S1: remove only the nonterminal batch trigger → all three engines fail `[]` vs exact tick-4
  advance (3 failed / 9 filtered out; exit 1).
- S2: remove only the window-blur listener → all three fail the paused-state assertion
  (3 failed / 9 filtered out; exit 1).
- S3: remove only the local-lead freeze → all three fail head 219 vs frozen 215
  (3 failed / 9 filtered out; exit 1).
- S4: remove only rejected-flush resync → all three fail current calls 0 vs 1
  (3 failed / 9 filtered out; exit 1).

The initial corrected Snake population passes 9/9 (3 unrelated Mine Grid cases filtered out).
`git diff --exit-code -- client/src/game-ui/minigame/SnakeBoard.svelte` passes after probes:
zero residual production change. The fixed-delay oracle is invalid under delayed callback
delivery; this does not prove the sole cause of the original WebKit failure or certify hosted
reliability. Broad verification is recorded below when it completes. No increased delay,
retry, runtime fix, weakened command, content, copy, kernel, CI topology or archival change.

**Restored broad verification:** root `make typecheck test-client build-client
verify-client-boundary verify-cosmetic-boundary verify-no-payment` exits 0: zero Svelte errors/
warnings; 6,953 unit tests passed / 86 browser-only skipped; production build and all three
boundaries green. Cold `make test-browser-ci` exits 0: 252 file populations, 21,102 tests passed /
3 intentional performance skips in 33.22 s; then the separate Chromium performance population
passes its one case (20 filtered out). All 12 Arcade cases pass in that full population. The
original red run is not deleted, and complete client CI is still blocked by RP-131's pushed
kernel history, whose repair RFC is draft. This test-only correction needs Claude's designated
review; it does not approve the rest of A6 or the Arcade RFC.

**Corrective review handoff:** exact Codex range `6652e467^..89434711`, test/instrument/docs/
tracking only; production child bytes are unchanged. **READY FOR CLAUDE DESIGNATED REVIEW.**
The original bounded Codex finding concerns only A6 scheduling evidence, not approval of
Claude's full `fca062a1` batch or later Arcade integration. No self-approval or archive.

## 2026-10-05 — RP-189 native keyboard acceptance supplement predeclaration (Codex)

**Review by:** Codex. **Recorded by:** Codex. **Bounded original range:** Claude
`fca062a1^..fca062a1`, A6/AC12 keyboard evidence only. **Verdict: CHANGES REQUIRED.** The
four committed cases dispatch synthetic F/Arrow/C events, click controls programmatically,
never complete Mine Grid via keyboard and never execute Snake quit. This is missing evidence,
not a demonstrated inability of the actual controls to receive keyboard input.

**Authority / population:** accepted AR6.2/AR6.3/AC12; test-only supplement. Native Playwright
keyboard events in Chromium/Firefox/WebKit activate the real children. Use the existing
Go-authored `mine_grid_clear_small` scenario's fixture/seed/commands, not a new solver or
fabricated terminal. Choose Small with Enter, traverse from cell 0 to 12 with real arrows,
activate reveal with Space, bind the real engine response into the same mounted component,
and compare the entire terminal snapshot to the corpus. No pointer or direct gameplay call
is permitted outside the component's host callbacks; keep hidden mine positions absent until
terminal. Snake quit is activated by Enter while initially paused, then its real engine state
must be terminal with the exact singleton quit command and visible terminal announcement.

**Discrimination / exit:** sever Mine Grid's primary-action handler and Snake's quit handler
independently. The old suite's ability to miss the Snake severing is an explicit baseline;
the new named cases must fail in all three engines. Restore probes and run the complete
Arcade and full cold browser populations plus root type/unit/build/boundaries. No production
mechanic, copy, timer, corpus, kernel, wire or CI change. Native focus placement is permitted
as test setup; no claim about complete Tab navigation, screen-reader participants, public
hosted Arcade integration or full AC12/task-matrix closure follows. Claude reviews this Codex
test-only corrective range; nothing authorizes archival or blocked API/mint implementation.

**Separately discovered RP-190 (queued, not corrected in RP-189):** direct Node enumeration
of `testdata/arcade/content-gate-v1.json` prints `{declared:43, attempts:60, applied:43}`.
The TS AR7 gate increments only after applied commands; AR7 says the budget equals the sum
of corpus command counts, including attempted rejections. Verify the Go producer and consumer
and predeclare a separate test/corpus correction. Do not fold this into native keyboard scope.

## 2026-10-05 — RP-189 executed native keyboard supplement (Codex)

With only Snake's Quit callback severed, the old complete Arcade population still passes
12/12. Adding the native keyboard cases under that same probe makes Snake fail in every
engine (`[]` vs `[{kind:"quit"}]`); Mine Grid's new case passes in all three (3 passed /
3 failed / 12 filtered out, exit 1). Restore Snake, sever only Mine Grid's cell primary-action
handler: the named native keyboard completion fails in all three because the cell-12 reveal
is missing (3 failed / 15 filtered out, exit 1). Both child files are restored exactly.

Mine Grid's retained witness activates Small with native Enter, moves focus through real
arrows to cell 12, and activates with native Space. Both commands enter solely through the
real child's callback and are applied by the actual engine; its response updates the same
mounted component. The full terminal snapshot must equal the existing Go-authored seed-202
`mine_grid_clear_small` scenario, with all 24 safe cells and the terminal mine shown. No mines
appear in the initial playing response. Snake's retained Enter-on-Quit case requires its
actual engine terminal at tick 0/revision 2, exact singleton quit command, terminal text and
removed controls. Terminal axe checks run for both. Focus placement is fixture setup, not
proof of whole-task Tab navigation. No production bug is alleged from this evidence gap.

Root type/unit/build/boundaries pass: 6,953 unit tests / 88 browser-only skips; zero Svelte
errors/warnings, production build and shell/Cosmetic/no-payment boundaries green. Full cold
browser verification follows when complete. RP-190 remains separately queued, not repaired
in this native keyboard range. No runtime/copy/content/engine/kernel/wire/CI change or archive.

**Restored broad browser evidence:** cold root `make test-browser-ci` exits 0 with all 18
Arcade cases green, 252 file populations / 21,108 tests passed / 3 intentional performance
skips in 37.58 s, followed by the separate Chromium performance case (1 passed / 20 filtered
out). Both production child files have zero diff. All new gameplay starts inside native-key
component callbacks; setup only creates the test session. This is **READY FOR CLAUDE
DESIGNATED REVIEW**, not a Codex self-approval or full A6/AC12/release acceptance.

**Native keyboard corrective review handoff:** exact Codex range `9026bd40^..27622885`.
It contains tests/harness/docs/tracking and a separately queued RP-190 finding, not product
or corpus changes. Claude's designated review is required; original A6 remains unapproved.

## 2026-10-05 — RP-190 AR7 budget correction predeclaration (Codex)

**Review by:** Codex. **Recorded by:** Codex. **Bounded original range:** Claude
`508fe19a^..508fe19a`, AR7 corpus budget only. **Verdict: CHANGES REQUIRED.** Both the Go
generator and TS replay consumer count applied commands only; direct fixture enumeration is
43 applied / 60 attempts. AR7 requires the sum of command counts, not applied-state transitions.

**Authority / population:** accepted AR7. Tests and generated fixture metadata only; no engine,
catalog, seeded scenario, expected snapshot/result, command, kernel, policy or production byte.
Add independent Go and TS equality checks against each scenario's step count before correcting
the producer. They must fail first on declared/generated 43 vs 60. Then count every attempted
command in the Go generator and TS execution consumer and regenerate via the existing root
Make target. JSON comparison against the prior committed fixture must show only the budget
changed (43→60), with every other value identical.

**Exit / negative controls:** cold Go Arcade and full client/type checks green; separately
restore applied-only Go counting and applied-only TS execution counting, each must fail its
corresponding gate. Restore each probe exactly. Regeneration must be byte-reproducible and
the full cold browser lane, including performance, must pass. Use `-count=1` for Go evidence;
do not treat a cached corpus check as cold proof. Root vet/build/boundaries remain required.
No added retry/bound increase or fabricated content coverage. This repairs the observation
denominator only; it does not establish missing gameplay scenarios or approve the rest of
Claude's A1–A7/AR7 range. Claude independently reviews the exact Codex correction.

## 2026-10-05 — RP-190 exact attempted-command budget correction (Codex)

The new Go equality fails first on the unchanged generator: `budget=43 must count all 60
attempted commands (AR7)`; cold focused Make exits 2. The new TS metadata equality also fails
first on the unchanged fixture: `expected 43 to be 60`; full client Make exits 2 with one
failed / 6,953 passed / 88 browser-only skips. These are executed failures, not read assertions.

The generator now counts each scenario's complete step list, and TS increments before every
attempt, rejected or applied. Root `make arcade-corpus` regenerates metadata 43→60. Structural
comparison to committed HEAD confirms every other value—including scenarios, commands,
expected terminal/results and the content hash—is identical. Regenerating a second time is
byte-identical (SHA-256 `61994942904ab0cc03f608ed8e60471f825f49b3a554cf43a4f7f344af54af28`).

**Independent restored negatives:** reinstate only Go's applied-only producer → the cold
Go gate fails at 43 vs 60 (Make 2); restore. Reinstate only TS's applied-only execution counter
→ its metadata check stays green but actual replay count fails 43 vs 60 (Make 2; one failed /
6,953 passed / 88 browser-only skips); restore. This tests distinct producer and execution
consumers, not two copies of the same assertion.

Restored cold `make test-go GO_PACKAGES='./arcade' GO_TEST_FLAGS='-count=1'` passes, as does
root type/client/build/boundaries/vet (6,954 unit tests / 88 browser-only skips; zero Svelte
errors/warnings). The full cold browser result is recorded below when complete. No runtime,
scenario, content, balance, copy, kernel, wire or CI bytes change, and no archive is authorized.

**Broad browser evidence:** restored cold root `make test-browser-ci` exits 0: 252 file
populations / 21,111 tests passed / 3 intentional performance skips in 38.01 s, followed by
the separate Chromium performance case (1 passed / 20 filtered out). This correction is
**READY FOR CLAUDE DESIGNATED REVIEW**, not a Codex self-approval or wider AR7 acceptance.
The full kernel-history CI gate remains red on RP-131; its repair RFC remains draft.

**Separately queued RP-191:** the TS AC4 case named "never exposes a mine position in a
non-terminal snapshot" reads only corpus `expected_terminal` metadata and asserts terminal
phase; it observes no nonterminal transition or forged-state refusal. Go
`TestMineGridHiddenInformation` and the new RP-189 browser playing-state check supply actual
positive controls, so this is not a claim that all AC4 proof is absent or mines leak. Next,
predeclare a bounded TS intermediate-state and leaked-state-refusal population with precise
severing probes. No RP-191 fix is included in this budget range.

**Budget corrective review handoff:** exact Codex range `593ee762^..bcdee28d`, tests/generated
fixture budget metadata/docs/tracking only. **READY FOR CLAUDE DESIGNATED REVIEW.** No
engine, candidate content, owner copy, kernel or production file changed in this range.
Reference correction to the immediately preceding RP-191 note: the actual Go positive-control
name is `TestMineGridNeverExposesMinesBeforeTerminal` in `server/arcade/engine_test.go`, not the
shorthand `TestMineGridHiddenInformation`. It ran in the cold Arcade package population above.

## 2026-10-05 — RP-191 nonterminal Mine Grid witness predeclaration (Codex)

**Authority:** accepted AR3.4/AC4; bounded A2 TS test correction on Claude's original
`508fe19a^..508fe19a`. **Review by:** Codex. **Recorded by:** Codex. **Verdict: CHANGES
REQUIRED** on the named TS hidden-information instrument, not a demonstrated production leak.
The existing case reads terminal fixture metadata only. Go's executed
`TestMineGridNeverExposesMinesBeforeTerminal` and RP-189's browser playing-state control remain
positive evidence; no claim that the entire protection is missing.

**Question / population:** run all eight existing Go-authored Mine Grid corpus scenarios
through the real TS create/apply engine, observing genesis and after every attempted command,
including rejected attempts. Inspect raw JSON output independently of the decoder being
tested. Every nonterminal output must have empty `mine_cells` and `exploded_cell:-1`.
Require actually observed setup, unplaced playing, placed playing, terminal-unplaced and
terminal-placed populations, and prove the observation count covers every scenario/step.
Terminal placed states must contain the actual derived sorted mine list; unplaced quit stays
empty. Compare each final state to the existing corpus and preserve rejection outcomes.

**Refusal controls:** for every actual placed/nonterminal state, independently forge only
`mine_cells` (real derived positions) or `exploded_cell` (one real mine); both the exported
decoder and engine apply entry must reject as SyntaxError. Keep the exact clean input intact
and continue the scenario to its existing terminal. No replacement engine, invented seed,
new corpus, public API call or fabricated terminal is allowed.

**Discrimination:** first remove only the TS decoder's hidden-state guard and run the old
named test: its terminal-only population is expected to remain green. Restore before adding
the new witness. Then separately expose actual mine positions in a placed nonterminal engine
output, disable only mine-list refusal, and disable only exploded-cell refusal. The new named
population must fail each independent mutation; restore source byte-exactly before broad
verification. Exit requires cold Go, client/type/build/boundary/vet checks and full cold Linux
browser CI including performance. Test-only correction unless a new real defect is separately
reproduced and recorded. No gameplay, kernel, catalog, copy, schema, wire, deployment, archival
or release promotion. Claude designated review is mandatory for the exact corrective range.

**Instrument refinement before the additional probe:** standalone decoder refusal and engine
entry refusal are distinct asserted consumers. In addition to the three declared mutations,
temporarily bypass only the decoder call inside `applyMineGrid` with JSON parsing: standalone
refusal must stay green while the engine-entry refusal fails. Restore before broad verification.
This remains a test-only range and does not authorize retaining the source mutation.

## 2026-10-05 — RP-191 actual hidden-state observations and refusal proof (Codex)

**Executed old-instrument control:** removing only the decoder hidden-state guard leaves the
entire old client population green (6,954 passed / 88 browser-only skips). The attempted
`CLIENT_TEST_FLAGS` selector is not consumed by this Make target; this was a full run, not a
selected-case run. The guard was restored before the replacement witness. This does not erase
Go's real `TestMineGridNeverExposesMinesBeforeTerminal` or RP-189's browser positive control.

The replacement executes all eight existing Mine Grid scenarios: 44 raw genesis/attempt
observations, including rejections, require the five declared phase/placement populations.
It checks nonterminal hiding and terminal disclosure, accepts every clean snapshot through
the decoder, and separately forges actual mine positions and an actual explosion at each
placed/nonterminal state. Both exported decoder and apply entry must reject; untouched replay
must still reach each exact corpus terminal. No fixture, command or production bytes change.

**Independent executed negatives (full client population each time):**

- S1: emit the real mine list in placed/playing output → the raw observation fails on
  `mine_grid_preset_large`, revision 3. The existing corpus replay also fails later through the
  decoder (defense in depth, not exclusively new detection): 2 failed / 6,952 passed, Make 2.
- S2: remove only mine-list refusal → the named decoder mine-list assertion fails:
  1 failed / 6,953 passed, Make 2.
- S3: remove only explosion refusal → the named decoder explosion assertion fails:
  1 failed / 6,953 passed, Make 2.
- S4: bypass only the decoder call inside apply with JSON parsing → standalone decoder
  refusal remains green, but engine refusal fails because the promise resolves a real quit
  terminal: 1 failed / 6,953 passed, Make 2.

All four probes are restored exactly; `git diff --exit-code --
client/src/arcade/mine-grid.ts` exits 0. Each negative population has 88 browser-only skips.
Restored cold `make test-go GO_PACKAGES='./arcade' GO_TEST_FLAGS='-count=1'` passes. Root
type/client/build/client-boundary/cosmetic-boundary/no-payment/vet passes: zero Svelte
errors/warnings, 6,954 unit tests / 88 browser-only skips. Cold root `make test-browser-ci`
exits 0: 252 file populations / 21,111 tests passed / 3 intentional performance skips
(32.65 s), then the separate Chromium performance case (1 passed / 20 filtered out).

**READY FOR CLAUDE DESIGNATED REVIEW**, not a self-approval, wider A2/AC4 acceptance, public
integration or archival. RP-131 still blocks complete green kernel-history CI. Next bounded
accepted-lane diagnostic: snapshot grammar parity between Go and TS decoders; reproduce before
claiming a defect, predeclare before retained corrections, and keep it outside this test-only
range. Owner wire/copy/mint and broader release gates remain independent obligations.

**Hidden-state corrective review handoff:** exact Codex range `24b7b9d7^..fe49c408`,
test/docs/tracking only. **READY FOR CLAUDE DESIGNATED REVIEW.** No runtime, fixture,
candidate content, owner copy, kernel, schema or CI file changed. This range does not approve
Claude's original full A2 implementation or authorize an archival move.

## 2026-10-05 — Mine Grid snapshot value-grammar parity predeclaration (Codex)

**Authority:** accepted AR3.2/AR3.4 snapshot coordinates, numeric fields and revealed-row
shape; AR7 Go/TS parity. Original Claude A2 range `508fe19a^..508fe19a`. This is a separate
range from RP-191's test-only witness. No defect is claimed before execution.

**Population:** one shared mutation table applied to real Go and TS fixture-engine snapshots
(seed 7, `large` board), unplaced after choose, placed after revealing cell 40, and terminal
after quit. Independently modify revision type/fraction/null, dimension/mine type/fraction,
preset type, flag/revealed/mine cell type/fraction, adjacency type/fraction and an extra
revealed-row key. All inputs are valid JSON but violate declared value/row grammar. Each
decoder must reject; each real apply entry must refuse without yielding a terminal/result.
For every forged input preserve and independently apply the exact clean base as a positive
control, comparing its output with the unmodified baseline. Go must report invalid tenant;
TS must report SyntaxError. Record separately which existing later guards already refuse.

**Exit / authorization:** if unchanged Go rejects while TS accepts, record the reproduced gap
immediately in BACKLOG before a bounded TS type/integer/row-shape correction. No new mechanics,
balance data, public wire, persisted schema, content epoch or owner text. A retained watched
engine correction must honestly advance shared kernel identity and retain golden-vector parity.
Independently remove numeric, list-element and nested-row validation: each declared population
must fail; restore before cold Go/client/type/build/boundary/vet/vector and full Linux browser
verification. Exact Claude corrective review remains mandatory; no archival or push.

**Limits:** this population does not establish raw JSON token/duplicate-key parity, every
semantic impossibility, Snake validation, public-route acceptance or real storage migration.
Those are distinct follow-up diagnostics, not silently covered by a value-grammar test.

**Executed finding RP-192 (before production edits):** 19 shared negatives produce 18 TS
decoder failures and three apply failures: fractional flags, string flags and an extra
revealed-row key survive to real quit outputs. Full client run: 21 failed / 6,971 passed /
88 browser-only skips, Make 2. The remaining apply cases already refuse through identity,
catalog or phase guards; do not claim they all yielded corrupt results.

**Go instrument correction:** all 19 decoder negatives reject, but the first Go test wrongly
requires `ErrInvalidTenant` from the registry too. The registry canonical/input/tenant-error
boundary returns `ErrTenantDivergence` for 11 malformed shapes instead; it does not accept them.
Record the initial cold Make 2 as an oracle error, not a runtime defect. Refine the population
to separately require direct tenant Apply's `ErrInvalidTenant`, registry refusal as one of its
two fail-closed classes, and unchanged clean controls. This names the exact boundary without
accepting arbitrary errors or modifying production. No TS defect is inferred from that Go
oracle failure; the TS resolved quit outputs independently reproduce the defect.

**Further Go oracle correction:** the direct tenant also deliberately maps invalid stored
snapshots to `ErrTenantDivergence` (`mine_grid.go` Apply's decoder/identity/catalog boundary).
The second cold run refuses every input but fails all 19 on my still-wrong expected error
identity. Refine direct Apply to that exact error; keep decoder `ErrInvalidTenant` and
registry's two precise classes. This is recorded as my instrument error, not a product fix.
Clean direct-vs-registry outputs must match exactly, including the real terminal result.

## 2026-10-05 — RP-192 Mine Grid snapshot value grammar corrected (Codex)

**Review by:** Codex. **Recorded by:** Codex. **Verdict: CHANGES REQUIRED**, bounded A2
snapshot-value grammar on original Claude `508fe19a^..508fe19a`. The unchanged TS decoder
accepted 18 of 19 shared malformed snapshots, and real apply carried fractional/string flags
and an extra revealed-row key into quit results. Other apply cases already failed via later
guards. Go refuses all 19 at decoder, direct-tenant and registry boundaries after the two
disclosed test-oracle error-class corrections above; cold package run exits 0. No public
route exploit or missing server protection is alleged.

The bounded TS correction requires numeric integer types for revision/dimensions/mines,
string-or-null preset identity, integer list cells, and exact typed `adjacent`/`cell` rows.
The shared 19-row fixture mutates actual fixture-engine states, not hand-invented snapshots.
Clean decoder and direct/registry apply controls retain their existing terminals or exact
terminal-phase refusal. The watched runtime correction honestly advances shared kernel
identity 0.3.143→0.3.144 in all three version files; no balance, content epoch, copy, schema,
public API or CI file changes.

**Independent executed severings (full client runs, each Make 2 / 88 browser-only skips):**

- Remove only the new top-level revision/dimension/mine/preset guards → 9 named decoder
  failures / 6,983 passed. Existing later apply guards still refuse those inputs.
- Remove only integer validation from sorted lists → 6 failures / 6,986 passed: four decoder
  cases and the two flag apply cases that again resolve real quit terminals.
- Remove only exact revealed-row keys → decoder AND apply extra-key cases fail:
  2 failed / 6,990 passed.
- Remove only adjacency integer validation → both adjacency decoder cases fail:
  2 failed / 6,990 passed. Catalog-aware apply remains defense in depth here.

All probes are restored; retained production diff is only the stated validation and version
change. Restored cold Go `./decimal ./arcade ./kernel ./minigame ./replaycatalog` passes with
`-count=1`. Root type/client/build/boundaries/no-payment/vet/vectors-check exits 0: 6,992 client
tests / 88 browser-only skips, zero Svelte errors/warnings, 6,296 numeric vectors regenerated
byte-identically. Broad browser and fresh kernel-history outcomes follow when terminal; no
selected rerun or guard bypass substitutes for them. **READY FOR CLAUDE DESIGNATED REVIEW**
only after those outcomes are recorded; no full A2/AR7 approval or archival follows.

**Broad browser outcome:** restored cold root `make test-browser-ci` exits 0: 255 file
populations / 21,225 tests passed / 3 intentional performance skips in 35.37 s; separate
Chromium performance case passes (1 / 20 filtered out). This is local, declared Linux CI-lane
evidence, not hosted reliability. The population grows by the new 38 unit cases run in each
browser; it is not 114 new player journeys. **READY FOR CLAUDE DESIGNATED REVIEW**.

**Fresh complete-history outcome:** root `make verify-kernel-version` exits 2 at the unchanged
pushed RP-131 commit `50a3a5141d5c4c1cd25cbf3c9c39d9e3cf8e2444`, parent
`0cf9f7a6aba4038fadcdf35e5b94f56986164af7`, on its six minigame-prefix files. The CI checkout
contract and its negative fixtures pass first. The walk stops before judging this new range;
do not call it green because version 0.3.144 and local kernel parity tests pass. No history
rewrite, exception, artificial version signal or guard bypass was introduced. The history
repair RFC remains draft and separate from accepted Arcade engine work.

The separately executed root `node client/tools/verify-kernel-version-fixtures.mjs` exits 0
(`kernel history guard adversarial fixtures ok`). That validates the guard's negative controls,
not the failing repository history. No bypass is retained.

**Snapshot-value corrective review handoff:** exact Codex range `8c01a8f5^..9288d1cc`.
Watched TS decoder plus shared kernel 0.3.144, Go/TS tests, negative fixture, docs and tracking.
**READY FOR CLAUDE DESIGNATED REVIEW.** All four probes were restored before the broad
verification above; no production Go engine, gameplay/copy/content/schema/public wire/CI byte
changed. The fresh historical red gate remains on record. No full A2/AR7 acceptance or
archival; raw token/duplicate-key and nullable/missing nested-value parity is the next separately
predeclared diagnostic, not evidence carried by this range.

## 2026-10-05 — Mine Grid raw snapshot grammar predeclaration (Codex)

**Authority:** accepted AR3.2/AR3.4 declared numeric/list/row shape and AR7 cross-runtime
parity, original Claude A2 `508fe19a^..508fe19a`. Separate from RP-192's value-only correction.

**Question / population:** derive genesis, chosen-unplaced, placed (seed 7, fixture `large`,
reveal 40), and quit-terminal snapshots through the real engines. A shared raw replacement
table preserves source distinctions that JSON.parse erases: duplicate top/nested keys
(including escaped-equivalent keys), decimal/exponent integer tokens, null numeric scalars/
array elements/nested rows, and missing declared revealed-row fields. Include unsafe revision
and malformed/trailing JSON controls. Every replacement must match exactly once, with actual
clean source validated independently; fail the instrument on an unmatched/ambiguous needle.
Require decoder AND direct pure engine refusal for every malformed raw input, with exact
existing error classes (Go decoder invalid tenant, direct apply divergence; TS SyntaxError).
Do not confuse the Go registry's canonical-input boundary with direct-engine admission.

**Positive controls:** every unmodified actual state decodes; nonterminal clean Apply reaches
the same quit output on repeated pure execution, terminals retain illegal_phase. Valid raw
whitespace/reordered fields, escaped strings/keys, integer-valued coordinate boundaries and
empty lists remain legal. Refusing every noncanonical spelling is NOT the objective: declared
integer token/type and unique/exact-key grammar are. Negative-zero integer tokens keep the
existing Go decoder's behavior rather than inventing a canonical-token law.

**Authorization / exit:** reproduce and register each actual admission gap before production
edits. A bounded decoder-only correction may enforce declared nullability, exact row fields,
unique keys and integer tokens without changing mechanics/payout/content/copy/wire/schema or
making a production mint. Shared numeric fields must stay safely representable in TS; no
convenience ceiling or balance cap. A retained watched change honestly advances kernel identity.
Independently sever duplicate detection, token-integrality checking, Go nonnullable-number
checking and Go exact revealed-row checking; each targeted population must fail. Restore
before cold focused Go, whole client/type/build/boundary/vet/vector and complete Linux browser
verification. No self-approval, archival, history repair, push or public integration claim.

**Limits:** this is Mine Grid stored snapshot grammar, not Snake, command/catalog grammar,
every semantically impossible state, public API/storage acceptance, migrations or clean-host
release evidence. Those remain separately scoped obligations.

**Executed RP-193 before production edits:** the shared 25-negative/six-positive raw table
matches each actual source needle exactly once. TS admits 11 cases at BOTH decoder and direct
Apply: seven decimal/exponent tokens and four duplicate/escaped-duplicate keys. Initial full
client run: 22 failed / 7,027 passed / 88 browser-only skips, Make 2. Go's cold package run
fails 12 raw subcases: null numeric fields/elements/row, missing row fields and unsafe revision.
Seven yield actual quit outputs after implicit zero normalization. Null terminal mine passes
decode but reaches illegal_phase (not a new result); the remaining malformed Go apply inputs
already refuse through later guards. All six legal spelling/boundary controls pass unchanged.
These direct-engine counterexamples establish grammar drift, not public API/storage admission.

**Parser instrument refinement before depth probing:** Mine Grid's declared compound shape is
at most root object → revealed list → revealed row (three compound levels). A raw scanner must
not recurse arbitrarily into malformed input that the old value decoder would reject. Add an
actual deeply nested-array negative and require SyntaxError rather than RangeError/stack
exhaustion. A three-compound-level bound derives from AR3.4 grammar, not a convenience budget;
legitimate fields cannot be nested deeper. Demonstrate the unbounded draft fails first, then
sever the derived bound separately and restore before broad verification.

## 2026-10-05 — RP-193 raw snapshot grammar corrected in both engines (Codex)

**Review by:** Codex. **Recorded by:** Codex. **Verdict: CHANGES REQUIRED**, bounded original
Claude A2 `508fe19a^..508fe19a` on executed raw-grammar counterexamples. Go's silent
null/missing→zero normalization and TS's JSON.parse-erased duplicate/token distinctions are
now refused before transition. Go checks declared raw number/list/row fields and shared safe
integers. TS scans unique decoded keys and integer tokens before its existing value grammar.
No rule, payout, content, owner copy, API/schema/CI byte or production epoch changes. Watched
decoder changes advance shared kernel identity 0.3.144→0.3.145.

Six valid spelling/boundary controls prove this is not reject-all or canonical-spelling-only
validation: whitespace, escaped key, escaped phase string, negative-zero integer, flags at
0/80, and reordered revealed-row fields still decode and execute correctly. Raw replacement
assertions reject unmatched AND ambiguous needles. Clean direct execution remains exact.

**Executed independent severings, each Make 2:**

- TS duplicate detection only → four duplicated/escaped-duplicate cases fail at decoder AND
  direct apply: 8 failed / 7,042 passed / 88 browser-only skips.
- TS decimal/exponent-token refusal only → seven cases fail at decoder AND direct apply:
  14 failed / 7,036 passed / 88 browser-only skips.
- Go nonnullable-number refusal only → eight named raw subcases fail; explicit missing-row
  shape checks still hold. Six direct quit outputs again normalize null to zero; terminal
  mine reaches phase refusal, not a new result. Cold package run.
- Go revealed-row raw-validation loop only → five nested null/missing subcases fail at decode,
  two again yielding quit outputs. Other top-level/list guards remain. Cold package run.
- Go shared safe-integer bound only → unsafe revision decoder case fails; identity checking
  still prevents its direct apply. Cold package run.
- TS shape-depth accounting only → malformed 10,000-array nesting produces RangeError instead
  of SyntaxError: 1 failed / 7,049 passed. Restore. The initial unbounded draft failed the
  same retained test BEFORE the shape-derived bound; no arbitrary depth budget was adopted.

All probes restored. Two failed local patch-context attempts applied no changes; the actual
depth severing changed recursive depth accounting, NOT the guard comparison. Restored cold Go
`./decimal ./arcade ./kernel ./minigame ./replaycatalog` passes with `-count=1`.
Root type/client/build/boundaries/no-payment/vet/vectors/corpus-check exits 0: 7,050 client
tests / 88 browser-only skips, zero Svelte errors/warnings; numeric vectors and existing
Arcade corpus remain byte-unchanged. Full browser and kernel-history outcomes are recorded
when terminal, not substituted by a selected rerun. No full A2/AR7 or archival acceptance.

**Full cold browser outcome:** root `make test-browser-ci` exits 0: 255 file populations /
21,399 passed / 3 intentional performance skips in 57.24 s, followed by separate Chromium
performance (1 passed / 20 filtered out). These include existing numeric/unit populations
executed in three browsers, not 21,399 distinct player journeys. No selected retry was used.
**READY FOR CLAUDE DESIGNATED REVIEW**; complete-history outcome remains separately recorded.

**Next separately queued RP-194:** AR7 requires every snapshot/result byte comparison; corpus
steps currently contain only command/outcome (`content_gate_test.go`), while the TS primary
replay compares only each final scenario snapshot/result. Intermediate result values can be
overwritten before that assertion. This is a read-derived evidence gap, not a demonstrated
runtime fault; next, predeclare and execute an intermediate-only severing control, then a
test-side Go per-command witness if the gap reproduces. No RP-194 correction is in this
raw-decoder range. Existing real Mine Grid hiding and terminal/corpus evidence is preserved.

**Fresh complete-history outcome:** root `make verify-kernel-version` exits 2 at unchanged
pushed RP-131 commit `50a3a5141d5c4c1cd25cbf3c9c39d9e3cf8e2444`, parent
`0cf9f7a6aba4038fadcdf35e5b94f56986164af7`, on its six minigame-prefix paths. CI checkout
contract and its negative fixtures pass first. The walk stops before this new range; version
0.3.145/local parity do not make history green. No exception, rewrite or bypass was added;
the history repair RFC remains draft. This does not prevent the accepted decoder correction
from being handed off honestly, but blocks any complete-green-CI or release claim.

**Raw-grammar corrective review handoff:** exact Codex range `d326fb2c^..45c35800`.
**READY FOR CLAUDE DESIGNATED REVIEW.** Both watched Mine Grid decoders, local TS raw scanner,
shared kernel 0.3.145, Go/TS tests, raw fixture, docs and tracking only. No gameplay, balance,
production epoch, owner copy, persisted schema, public wire, CI or history exception changes.
Do not archive or treat this as complete A2/AR7 acceptance. The fresh complete-history failure
remains explicit. Record correction: there were THREE failed patch-context attempts, not the
two stated above; none applied edits. Executed probe ranges and outcomes are unchanged.

## 2026-10-05 — RP-194 every-command byte-parity witness predeclaration (Codex)

**Authority:** accepted AR7, original Claude A2/A3 content-gate range
`508fe19a^..508fe19a`. Prior raw/value corrections do not prove this criterion.

**Population:** every existing scenario and every attempted command in the current 16-scenario,
60-attempt Go corpus (43 applied, 17 rejected), plus all 16 actual genesis snapshots. No new
scenario/seed/command/content/result is authorized. First temporarily return a fabricated
nonterminal result from TS choose_board, then separately return semantically identical but
noncanonical snapshot bytes for that command. Run the old whole client population and record
whether its final-only comparison survives each. Restore before retaining any instrumentation.
Existing browser/Go oracles may discriminate other properties; do not claim universal invisibility.

**Correction if reproduced:** promote the TEST corpus to v2 with actual Go-generated genesis
and every post-attempt snapshot/result witnesses. Preserve v1 as the historical comparison
artifact, with only v2 active generation/replay. TS compares exact bytes at genesis and after
every attempt; rejected attempts compare unchanged snapshot, result and revision. Applied
attempts compare actual nullable output results, not a result inferred from phase. Go generation
clones captured bytes/results so later steps cannot overwrite witnesses.

**Exit:** removing only added v2 witness fields and version from the generated data must recover
the exact v1 structure (metadata/budget/scenarios/commands/terminal/results/content identity).
Regeneration is byte-identical. Independently repeat both TS output mutations against the new
case; each must fail. Tampering or omitting a generated post-attempt witness must fail the cold
Go regeneration gate. Restore all probes before cold Go/client/type/build/boundary/vet and
complete Linux browser verification. Corpus instrument and tests only: no retained runtime,
balance, owner copy, kernel, content epoch, schema, public wire or CI change. Exact Claude
corrective review is required; no full AR7/designated approval or archival is inferred.

**Executed old-instrument controls:** fabricated nonterminal choose_board result AND a
separate semantically identical/noncanonical choose_board snapshot each survive the old
whole client population: 7,050 passed / 88 browser-only skips, Make 0 each. Both are restored;
the retained Mine Grid source has zero diff. These are direct TS output mutants, not claims
that original production returned them or every independent browser/Go witness is blind.

**Correction to the immediately preceding control record (Codex):** the fabricated result
really passed the whole old suite (7,050 / 88 skips, Make 0). The reordered snapshot did NOT:
its completed whole-suite run exits 2 with 2 failed / 7,048 passed / 88 skips. Both failures
are the raw-grammar fixture's exact trailing-key needle, not the original final-only corpus
comparison, which remained green. I recorded the second result before observing its terminal
output; that premature claim was mine. A focused original replay is run separately below.
The finding is about the corpus instrument, not universal invisibility to all other checks.

**Focused original replay:** with only snapshot ordering severed, the original named
Go-generated scenario replay passes (1 passed / 6 filtered skips, exit 0). The initial pnpm
launcher produced no output and was interrupted (130); the successful command uses the
installed repository Vitest directly with `--root client`, from the repository root. Neither
is substituted for the disclosed whole-suite 2-failure result. The probe is restored exactly.

**New instrument controls:** same fabricated choose_board result now fails the new per-step
result equality at attempt 1 (1 failed / 7,049 passed / 88 skips, Make 2). Same noncanonical
snapshot now ALSO fails its new per-step snapshot equality at attempt 1 (3 failed / 7,047
passed / 88 skips, Make 2): the other two failures are the already disclosed raw fixture
needles. Both production probes are restored. The Go v2 generator strips back to EXACT v1
structure, with 16 scenarios / 60 attempts / 43 applied / 17 rejected and unchanged content.
Generation followed by regeneration is byte-identical; v2 SHA-256
`b02c433d0df123e77aec848c337b90ecef65899b59f396a97cf20ea25f23d6e7`.
Retained copy-ownership checks use the existing first corpus scenario's seed/commands and
mutate only test-owned session memory; no scenario or gameplay is added to the corpus.

**Go instrument discrimination:** sharing captured genesis/step bytes and nested results with
session memory fails all THREE ownership subcases in a cold package run (Make 2). Restoring
ownership but omitting the per-step snapshot witness fails the cold generation gate at
`mine_grid_preset_large step 1: missing actual snapshot witness` (Make 2). Both probes are
restored; generated corpus bytes were not overwritten during these negative runs.

**Literal-byte refinement before final verification:** the first v2 draft compared reconstructed
JSON objects. Preserve those readable witnesses, but also capture actual Go genesis/snapshot
strings and the actual `json.Marshal` result (including `null`) as literal strings. The TS
primary comparison consumes those strings directly, so fixture parsing cannot normalize a
token spelling or key-order difference. No population/command/runtime change. Earlier controls
and broad green runs above concern the first draft; repeat both output controls, the omitted
literal witness control and the complete gates on this final instrument before claiming them.

**Final literal instrument controls:** fabricated result fails `literal Go result bytes`
(1 failed / 7,049 passed / 88 skips, Make 2). Reordered snapshot fails `literal Go snapshot
bytes` (3 failed / 7,047 passed / 88 skips, Make 2; same two independent fixture failures).
Omitting only the literal result fails `missing actual result witness`; separately omitting
only literal snapshot fails `missing actual snapshot witness`, cold Go Make 2 each, despite
readable witnesses remaining present. All probes are restored exactly before final gates.

**Separately queued RP-195:** read-derived Snake snapshot-grammar question: several TS numeric
fields use coercive comparisons without integer validation, while Go's typed decoder can
normalize nulls. The present corpus exercises valid transitions, not a shared malformed-input
population for this second engine. Predeclare and execute direct decoder/apply counterexamples
next; do not treat this observation as a proven runtime fault or widen RP-194's test-only range.

**Bounded original review verdict — CHANGES REQUIRED (RP-194). Review by: Codex.
Recorded by: Codex.** Original Claude range `508fe19a^..508fe19a`, bounded to AR7's
per-command snapshot/result instrument; the original source at that hash was independently
read. Executed controls use current HEAD with earlier recorded corrections, not an untouched
historical build. Final-only equalities persisted until this correction. No full original-range
acceptance or rejection is inferred for unrelated engine/platform criteria.

**Final retained verification:** cold Go `./arcade ./kernel ./minigame ./replaycatalog` passes
with `-count=1`. Root type/client/build/boundaries/no-payment/vet/vectors/corpus-check exits 0:
7,050 client tests / 88 browser-only skips, zero Svelte errors/warnings, unchanged 6,296 numeric
vectors. Full cold Linux browser target exits 0: 255 file populations / 21,399 passed /
3 intentional performance skips (44.88 s), plus separate Chromium performance (1 passed /
20 filtered). These are cross-browser test executions, not distinct player journeys.
Final generation/regeneration is byte-identical at SHA-256
`8a1b3ac84c93f0b5ba51cb339533266378e0c2c5ca3aed72311620ad297dbaa9`.
Removing added fields/version reproduces the unchanged v1 structure exactly.

Fresh complete-history verification still exits 2 at pushed RP-131 commit
`50a3a5141d5c4c1cd25cbf3c9c39d9e3cf8e2444` against
`0cf9f7a6aba4038fadcdf35e5b94f56986164af7`, after checkout-contract fixtures pass. No guard
exception, bypass, rewrite or CI change was made. No retained product bytes changed; kernel
0.3.145, content/balance/copy/public-wire bytes and the original v1 fixture remain untouched.
**READY FOR CLAUDE DESIGNATED REVIEW** of this test/corpus correction, not full AR7,
complete CI or archival acceptance. Exact corrective span is pinned in the following checkpoint.

**RP-194 corrective review handoff:** exact Codex range `c48ce7f2^..92ed5e68`.
**READY FOR CLAUDE DESIGNATED REVIEW.** Test-side generator, v2 fixture, active TS test readers,
docs and tracking only. No production, kernel, balance, copy, mint, public wire or CI changes.
The v1 artifact remains byte-unchanged. The final literal outputs, negative controls, generation
identity and cold verification above belong to this range; the draft controls are distinguished.
RP-195 is only a queued, read-derived question here. No self-approval, archival, push or deployment.

## 2026-10-05 — RP-195 Snake snapshot-grammar parity predeclaration (Codex)

**Authority:** accepted A3/AR4.3–AR4.6 and AR7; original Claude `508fe19a^..508fe19a`.
Starting tree is clean at `85c18535`. RP-194's command corpus is evidence for valid transitions,
not malformed-input admission. No changes to movement, food, growth, outcomes, payout or clock.

**Population, fixed before execution:** 45 shared raw mutations and seven legal controls.
Actual states are driven through the existing v2 corpus: snake_rejections genesis and step 8,
snake_cleared step 1 (grown playing) and step 5 (terminal). Each baseline must match its
literal Go witness. Negatives: all nine integer fields null and string-valued; seven fractional
groups (score/food-index jointly); three boolean counters; four unsafe-integer groups; null body
and null/decimal/exponent body elements; repeated and escaped-equivalent keys; decimal/exponent
tick tokens; trailing JSON; array-valued seed/hash; missing tick and extra root field. Legal
controls cover whitespace, escaped key/direction, negative-zero tick, body whitespace, reordered
root keys and the shared safe-integer tick boundary. Cases bind exact raw needles and fail if
absent/ambiguous. No arbitrary malformed-state denominator or fixture-only clean control.

**Method/criterion:** first execute unchanged production against every negative at direct decoder
and direct tenant apply. Go must report invalid-tenant at decode and tenant-divergence at apply;
TS must report SyntaxError before transition. Record decoder-only admission separately from
actual quit outputs and terminal phase refusal. Re-run clean controls; legal spellings normalize
to identical execution, while the safe tick-boundary control checks its actual certified fact.
Do not treat an identity refusal or terminal-phase error as decoder admission being safe.

**Conditional correction:** if reproduced, tighten only both Snake snapshot decoders: reuse the
existing raw TS parser; validate actual typed safe integer fields and string identity values;
Go rejects null/integer-token/unsafe scalar and body-element admission before typed decode.
Keep existing body/food/phase rules, schema, descriptor and engine version. No new reachability
rule, body-adjacency rule, pacing cap, migration, public wire, content/mint or owner copy.
The watched acceptance-set correction requires an honest kernel bump from 0.3.145 to 0.3.146.

**Exit:** all 45 negatives refused and all seven legal controls usable; independent raw-parser,
TS numeric-type and Go null/body/safe-bound controls fire, then restore exactly. Go/TS valid
corpus and numeric vectors remain byte-identical. Cold Go/client/type/build/boundary/vet and
complete Linux browser verification are required, with complete-history failure reported
separately if RP-131 persists. Exact Claude review remains mandatory; no self-approval/archive.

**Separately queued RP-196:** the existing Snake fixture is 6×5, not AR7's specified 5×5; its
cleared terminal has 30 body cells. This read-derived mismatch belongs to a separate population
construction/research range, not a change to this grammar experiment or a claim of engine failure.

**Unchanged-production outcome:** full TS client Make 2, 56 failed / 7,092 passed / 88 skips:
32 malformed decoder admissions, 23 direct-apply failures (21 actual quit outputs, two terminal
phase refusals rather than grammar refusal), plus ONE legal-control oracle mistake of mine.
The negative-zero output serializes identically, but object equality distinguishes -0 from 0;
fix that oracle to compare actual snapshot and serialized result bytes as predeclared, not
rewrite valid engine behavior. Null/string/fraction/boolean facts and oversized counters reach
TS quit outputs. Other decoder-only admissions are refused by later identity checks.
Cold Go Make 2: eight malformed decoder admissions, seven actual quit outputs; oversized
revision still hits the identity check. Null tick/food/body entries normalize to zero. All
seven Go legal controls pass. Re-run the corrected oracle on unchanged TS production before
retaining the separately authorized decoder correction.

**Corrected-oracle unchanged-production rerun:** 55 failed / 7,093 passed / 88 skips, Make 2.
All seven legal controls now pass; malformed admissions remain exactly 32 decoder and 23 apply
failures (21 outputs / two terminal phase refusals). The negative-zero test mistake is resolved
in the instrument only. Proceed with the predeclared watched decoder correction; no new mechanic.

**Independent correction severings, each Make 2:**

- TS raw parser bypass only: 12 failed / 7,136 passed / 88 skips (six raw-key/integer-token
  cases at decoder AND apply; terminal cases reach phase refusal, not new outputs).
- TS numeric-type guard bypass only, raw parser retained: 23 failed / 7,125 passed / 88 skips
  (13 decoder admissions and ten actual quit outputs). Fractional/unsafe tokens still reject.
- TS string-identity type guards bypass only: two decoder cases fail / 7,146 pass / 88 skips;
  direct apply still rejects array-valued identities. That later guard is defense in depth,
  not a replacement for correct snapshot grammar.
- Go null-token refusal bypass only: five named cases fail at decoder AND direct apply,
  again producing normalized quit outputs; later schema/revision/food-index checks still hold.
- Go body-element raw-validation loop bypass only: the null body element fails at decoder
  AND direct apply; typed decode continues to reject decimal/exponent elements.
- Go shared safe-integer bound bypass only: unsafe revision/tick/growth fail at decoder;
  tick/growth again produce quit outputs. Revision is still refused by its identity check.

An unmatched needle seeded in the shared fixture fails the cold Go instrument at
`raw mutation must match exactly once`, not a fake grammar pass. It is restored, along with
all six production probes, before the full retained verification. The matched/ambiguous/empty
TS mutation controls also execute in the retained instrument. No arbitrary limit was changed.

**Bounded original designated review — CHANGES REQUIRED.** Review by: Codex.
Recorded by: Codex. Exact original Claude span: `508fe19a^..508fe19a`. This verdict concerns
the Snake snapshot admission contract in A3/AR4.3–AR4.6/AR7 only, not an approval of the full
historical A3 range or A1–A7. The original source was checked independently against the
unchanged starting decoder; the executed baseline is at the current repository population,
not a claim to have rebuilt every historical dependency. The actual decoder/apply outputs
and corrected-oracle baseline above reproduce the defect. The separate 6×5/5×5 population
mismatch is RP-196, not evidence that the rules cannot clear an odd board.

**Retained cold primary verification:** root Go `-count=1` passes `./decimal ./arcade ./kernel
./minigame ./replaycatalog`. Typecheck (zero errors/warnings), all 7,148 client tests (88 browser
only skips), build, client/cosmetic/no-payment boundaries and their negative controls, vet,
numeric-vector regeneration and Arcade-corpus regeneration pass. The original v1/v2 corpus
and 6,296 numeric vectors remain byte-unchanged. The complete Linux browser target passes
21,693 executions across the three engines (three intentional performance skips), then the
separate Chromium performance population passes; these are execution counts, not distinct
player workflows. No selected retry substitutes for the complete target.

`make verify-kernel-version` passes checkout/negative-fixture checks but exits 2 on the SAME
pushed historical RP-131 commit `50a3a5141d5c4c1cd25cbf3c9c39d9e3cf8e2444` (parent
`0cf9f7a6aba4038fadcdf35e5b94f56986164af7`). The honest new 0.3.146 identity does not repair
that missing historical bump. No exception, bypass, rewrite or CI edit; complete CI is not green.

One attempted docs patch had stale context and applied no changes; the subsequent context-
matched patch succeeded. Canonical docs also remove the unsupported inference that absence
of a Hamiltonian cycle proves 5×5 cannot clear. The original 6×5 evidence is preserved.

**Additional real-DB regression, initial environment failure:** declared root
`make test-save-integration SAVE_TEST_PACKAGES='./production'
SAVE_TEST_FLAGS='-run TestArcadeComposedIntegration -v'` exits 2 BEFORE any Go test runs.
The dedicated `cloud-clicker-postgres-1` logs `exec format error`: its cached Postgres image is
linux/amd64 while the Docker host is linux/aarch64. A normal pull of the same declared
`postgres:16-alpine` tag for linux/arm64 restores native local execution without changing
Compose/CI/source, removing orphan containers or touching production data. The cold declared
test lane is rerun below; the failed initial attempt is not a test pass.

**Real-DB rerun outcome:** the exact same declared cold production selector now executes
`TestArcadeComposedIntegrationUnlockLockPlayAndZeroCreditResolution` against real Postgres,
verbosely PASS (0.29 s), Make exit 0. This covers actual DB/platform library composition and
the existing 20×20 candidate Arcade states, not the owner-blocked public wire/AC8 host.
No test assertions, Docker configuration, runner workflow or production data were changed.
The earlier startup failure remains disclosed. RP-195 is **READY FOR CLAUDE DESIGNATED
REVIEW**, not self-approved or archival-eligible; exact corrective span follows this commit.

**RP-195 corrective review handoff:** exact Codex range `4f6173be^..45fcf3ae`.
**READY FOR CLAUDE DESIGNATED REVIEW.** The range includes the predeclaration, shared raw
45/7 instrument, decoder-only source correction, honest kernel 0.3.146, docs and tracking.
The negative cases, six independent guard severings, full retained browser checks and real-
Postgres rerun above belong to this range. RP-196 is a separately recorded population gap,
not implemented or approved here. No self-review gate, archival, push or deployment.

## 2026-10-05 — RP-196 exact 5×5 clearing population predeclaration (Codex)

**Authority:** accepted A3/AR7, fixture-first AR8; original Claude `508fe19a^..508fe19a`.
Starting clean HEAD `8bfb32d9`. This is test-side population construction, not a permission to
change Snake, food placement, growth, engine/kernel identity, public wire or production data.
Preserve the existing 6×5 fixture and both content-gate corpora byte-for-byte.

**Question/arms:** can the real rules clear an exact 5×5 fixture using a legal deterministic
trace? New test-only fixture is the existing corpus artifact with ONLY Snake width 6→5.
Construct a 24-cell cycle excluding top-left cell 0: the standard cycle on columns 1–4,
with two left-column two-cell detours. Pin all 24 cells and verify uniqueness, orthogonal
adjacency, bounds and exact excluded cell. The actual genesis remains head 12/body [12,11].
Follow the cycle, breaking into cell 0 only if it is the published food, the actual body
length is 24 with pending growth and the head is adjacent. No input food/body injection.

**Fixed research population:** seeds 1 through 1,024 inclusive, all executed even after the
first success. Each probe starts via the real tenant registry and advances one actual tick.
Observe every seed as cleared, crashed or excluded-food strategy failure. If cell 0 is food
before the declared final exit is legal, stop that strategy and record its current actual
tick/body/score; do not mislabel it a crash, a truncated successful run or global impossibility.
No seed selection based on an invented terminal. A maximum 578 ticks is derived from at most
24 food events × 24 cycle edges, plus two genesis-alignment steps; exhausting it invalidates
the experiment and fails loudly, never expands the ceiling or drops the seed.

**Witness/criteria:** select the lowest actually cleared seed only after all 1,024 probes
finish. Record complete denominator/outcome counts and largest observed tick in a new
test-only versioned 5×5 artifact. Replay the selected trace from real registry genesis,
capturing literal snapshot/result bytes after every attempted command. Batch at the existing
64-tick advance limit, flush after each food and at the exact terminal; include a deliberately
overlong last command that must refuse without mutation, then the exact legal terminal and
post-terminal phase refusal. TS independently replays every attempted command and matches
the Go literal genesis/snapshot/result bytes. Actual final body must contain all 25 unique
cells, food -1, phase terminal and outcome cleared; facts must match actual score/tick.
Use observed intermediate length/score changes to prove eat/grow, not terminal metadata alone.

**Failing controls:** the population validator must reject the existing 6×5 clearing witness,
a 24-cell/not-full terminal and falsified terminal/fact claims; cycle validator must reject
duplicate, missing and nonadjacent routes. Separately sever the final-exit policy (strategy
must fail to supply a clearing witness) and mutate one actual intermediate TS result/snapshot
to prove the new literal comparisons fail. Restore all probes before cold verification.
Generation/check must discriminate stale artifact bytes; no rewrite of existing evidence.

**Exit/routing:** success supplies the exact bounded 5×5 population, not all A3/AR7 or public
integration acceptance. Zero successes completes the bounded strategy research negatively;
do not infer impossibility or silently use 6×5 instead. Cold Go, whole client, type/build,
root boundaries/vet/vectors/corpus and complete browser target plus the real-DB selector are
retained checks. Historical RP-131 remains separately reported. Claude designated review
must cite the exact new range; no self-approval, archival, push or deployment.

**Research executed:** all 1,024 seeds finish with explicit observations: one cleared
(lowest/only seed 455), zero crashed and 1,023 excluded-food strategy failures. Largest
observed tick is 209, below the independently derived 578-tick guard. The clearing seed
finishes at tick 134, score 24, body length 25 with all cells 0..24, food -1 and pending
growth 1. The new separate fixture changes only Snake width, and its content hash is
`sha256:3bae7a0744d8e2795b88a93bdef180414a6db238f9dbd68e259c8cd0ce103bef`.
The generated artifact records every seed, all counts and the selected real registry trace:
26 attempted commands (24 applied / two refused), 27 literal state observations. Actual
nonterminal outputs show food consumption and growth; terminal overshoot refuses at tick
133 without mutation, exact tick 134 clears, and subsequent quit is illegal_phase.
TS independently replays every attempt and byte-compares actual genesis/snapshot/result.
No property of the other 1,023 seeds is inferred beyond this strategy's recorded failure.

**Discrimination/restoration:** cycle and population validators execute their retained
duplicate/missing/nonadjacent and 6×5/incomplete-body/false-outcome/false-fact controls.
Removing only the final-exit policy executes the full Go population, then fails Make 2:
`bounded strategy found no clearing witness`. A prematurely exhausted probe guard fails
Make 2 at seed 1 with `observation invalid`, not a silently excluded seed. Changing the
artifact's version makes complete regeneration/check fail Make 2 as stale. An invented
nonterminal TS result on only width-5 states fails the new first literal result comparison;
reversed nonterminal JSON key order independently fails the first literal snapshot comparison
(each focused installed Vitest exit 1). Production probes, artifact and guard are restored.

One first final-exit probe used an incorrectly escaped anchored Make selector: the expanded
command ran NO TESTS and exited 0. That invocation is invalid evidence, not a successful
severing. I corrected the selector, reran the actual complete seed experiment and recorded
its failure above. No production correction followed this instrumentation error.

The existing root Arcade generation/check aliases now include the new test-only artifact,
always cold; no CI job/topology change. Both existing corpora and their fixture remain
byte-identical, as do numeric vectors, kernel 0.3.146 and every production file. Final
retained cold verification follows; Claude review is still mandatory.

**Bounded original designated review — CHANGES REQUIRED.** Review by: Codex.
Recorded by: Codex. Exact original Claude span: `508fe19a^..508fe19a`. This finding concerns
AR7's named Snake 5×5 clearing population only. The independently inspected original fixture
is width 6 and the existing generated clearing body has 30 cells; it cannot serve as the
named population. This is not a whole-A3 verdict or a finding that the pure rules cannot
clear an odd board. The separate successful construction supplies the missing bounded
evidence locally and still requires Claude's exact-range designated review.

**Retained cold verification:** root Go `-count=1` passes `./decimal ./arcade ./kernel
./minigame ./replaycatalog`; all 7,151 client tests pass (88 browser-only skips), with zero
typecheck/Svelte errors or warnings. Build, client/cosmetic/no-payment boundaries and their
negative controls, vet, byte-identical numeric vectors and complete Arcade regeneration
pass. The declared real-Postgres production selector executes verbosely PASS (0.19 s),
not skipped. It remains DB/library composition, not owner-blocked public-wire acceptance.

Fresh `make verify-kernel-version` passes checkout and adversarial fixtures, then exits 2
on unchanged pushed RP-131 commit `50a3a5141d5c4c1cd25cbf3c9c39d9e3cf8e2444` against parent
`0cf9f7a6aba4038fadcdf35e5b94f56986164af7`. No exception, bypass, rewritten history or CI edit.
The complete history/CI claim remains red despite this test-only range's green behavioral
checks and unchanged honest kernel identity.

The first complete browser invocation reported 21,702 suite passes plus its separate
Chromium performance pass. My aggregation failed to retain the terminal metadata before
the completed session expired, so I reran the ENTIRE target for an explicitly retained
exit status, not a selected test retry or a claim that the prior invocation was red.

**Complete browser confirmation:** the full Linux target rerun is terminal Make exit 0,
21,702 passing executions across 261 file populations in Chromium/Firefox/WebKit, three
intentional performance skips, followed by the separately executed Chromium performance
case (one pass / 20 filtered). The terminal metadata is retained. Probes remain restored;
`git diff --check` and unchanged existing corpus/fixture/vector/kernel/production-byte checks
pass. **READY FOR CLAUDE DESIGNATED REVIEW** of this exact bounded test-only population
construction. No full A3/AR7, hosted reliability, complete green CI or archival promotion.

**RP-196 corrective review handoff:** exact Codex range `329fc994^..25906f01`.
**READY FOR CLAUDE DESIGNATED REVIEW.** Predeclaration, new test-only 5×5 fixture and
fully observed strategy artifact, Go registry generator, independent TS literal replay,
root generation/check aliases, docs and tracking only. Original v1/v2 Arcade evidence and
all production/kernel/balance/copy/public-wire/CI-workflow bytes remain unchanged. The
content-fixture commit uses AR7's required BALANCE-CHANGE subject without minting a
production epoch. No whole-A3 approval, self-archive, push or deployment.

## 2026-10-05 — A3 rules / RP-197 rejection-atomicity predeclaration (Codex)

**Authority:** accepted A3, AR4.4–AR4.7, AR7 and AC5/AC6; existing A5 composition owns the
real-Postgres test-side path. Original Claude range `508fe19a^..508fe19a` for pure rules;
`b9b3e9aa^..b9b3e9aa` is NOT assumed as the A5 hash: derive its actual commit before any
designated verdict. Starting clean HEAD `90a2437a`. No rule/content/schema/wire/mint/copy
change; independent correction review remains mandatory.

**Inventory/read-derived question:** current v2 covers all eight original Snake scenarios,
tail chase/self collision, food/growth, four walls, all rejection names and plain terminal
overshoot. RP-196 adds exact 5×5 clearing and another plain overshoot. Neither supplies an
explicit turn after the terminal tick, or an invalid same/opposite turn after earlier ticks
have already executed. A5's real-DB witness contains only successful play/resolution; no
durable rejected-command row comparison. This is RP-197, not proof of a production defect.

**Stage 1, existing-instrument review:** cold Go Arcade and whole TS client baseline. Independently
sever tail-vacate handling, empty-cell food selection, reversal refusal and terminal-window
refusal in each pure engine; restore after each, record which actual gate fires. Separately
change only the guard for a command containing a future turn after death, leaving plain
overshoot refusal intact. Run the OLD full Go Arcade and client populations on that conditional
probe before adding tests. Survival is an instrument finding, not permission to retain a
wrong engine. Partial working-copy mutation that cannot escape the tenant is not automatically
a durable-state defect; identify the actual observed boundary.

**Conditional supplement, fixed before execution:** use actual current corpus genesis
and grown tail/self-collision sources plus the existing exact 5×5 fixture. Explicit advance
cases: future turn after east-wall death, late same-direction turn, late reversal, invalid
window, non-ascending turns, successful advance/quit and post-terminal refusal. Refused
attempts compare literal state/result and revision; successful controls execute real output.
Do not rewrite existing corpus/fixture or conflate the 6×5 source with the 5×5 source.

The real-Postgres arm uses the existing candidate 20×20 Arcade bundle, a real Founder and
actual minigame Service Play. Before/after each rejection, compare repository-loaded state,
genesis, result and revision; SQL command-row aggregate (including payload/sequence/revision/
server timestamp); active status and cleared claim/token. Exclude UpdatedAt only because
claim/release legitimately change it. Execute at both genesis (empty command history) and
after a successful advance (nonempty history); late turns and future terminal turns must
refuse with their exact taxonomy. Every error is inspected, not ignored. Successful play
must advance revision and append exactly one row; eventual quit/resolution must remain usable.
This is DB/library composition, not the owner-blocked public API/socket/AC8 route.

**Discrimination:** conditional future-turn discard must fail the new pure/DB witnesses.
Independently force rejected-command persistence in the service error path and suppress claim
release: the new DB comparison must fail each, without changing the tenant error itself.
Restore every production probe before complete retained cold verification. If the existing
instruments already discriminate a hypothesis, record that positive result and add no redundant
test merely to inflate counts. Any actual runtime rule defect needs its own conditional
accepted-contract correction range and honest watched identity; this declaration authorizes
no new mechanic or speculative production edit.

**Exit:** cold selected Go, whole client/type/build, root boundaries/vet/vectors/corpus,
complete three-browser target and declared real-DB selector. Report historical RP-131
separately, with no bypass. Bounded original verdict + exact Codex supplement span; no full
A3/AR7 promotion, self-review archival, push or deployment.

**Predeclaration identity correction:** `git log -- server/production/arcade_integration_test.go`
derives A5's actual original Claude commit as `e1c71d7c`; its bounded reviewed span is
`e1c71d7c^..e1c71d7c`. My preceding negative "NOT assumed" sentence included an unnecessary,
unverified placeholder hash. It is not a source/review identity and must not be cited. The
only A5 identity used from here is the Git-derived one above. No original verdict has been
issued yet.

**Existing-instrument execution:** unchanged cold Go Arcade and all 7,151 client cases pass.
Each of the four broad AC5 severings fails independently in Go AND TS: tail collision fires
the actual tail-chase corpus; all-cell food changes actual genesis/outputs; reversal is
wrongly applied instead of invalid_turn; removed terminal-window refusal admits the plain
overshoot (including RP-196's overshoot). All production probes are restored.

The narrower future-turn-only discard probe survives the OLD full Go Arcade population
(Make 0, 19.25 s), all 7,151 OLD client cases (Make 0), and the OLD real-Postgres A5 witness
(verbosely PASS, 0.17 s, Make 0). Its plain overshoot guard remains effective. This confirms
RP-197's bounded instrument gap: no future turn is actually submitted. It is not evidence
the unchanged production engine is wrong. Proceed with the predeclared test-only supplement;
do not replace the existing controls that already discriminate their properties.

**New unchanged-production controls:** twelve shared pure-engine negative populations
(six commands × genesis/moved) pass in Go and TS; actual transition objects remain unchanged,
and valid next advance, quit and terminal refusal remain usable. The first real-Postgres
run passes its six genesis negatives, then fails MY accepted-move positive oracle: I used
x-coordinate 11 where the snapshot contains row-major cell 211 on the 20×20 candidate.
This is not an engine/admission failure. Correct only the test to compare the actual prior
head + one cell and prior tick + one, then rerun unchanged production. A combined patch with
stale log context failed atomically and applied no changes; the corrected patch follows.

**Corrected unchanged-production DB control:** the exact declared selector now executes
all twelve genesis/moved negatives, a real accepted advance and quit/resolution (Make 0,
verbose PASS, 0.15 s). Only those two accepted commands enter SQL history. The test inspects
actual repository state/genesis/result/revision, the complete ordered command-row aggregate,
active status and cleared claim. UpdatedAt remains deliberately excluded because acquiring
and releasing a claim legitimately updates it. This is not public API/receipt/socket proof.

**New discrimination:** the same conditional future-turn-only discard now fails the new Go
test at both genesis/moved cases (Make 2), the independent TS test at the same two cases
(Vitest 1), and real Postgres at the submitted future turn (Make 2: unexpected certified
terminal, followed by the legitimately retained terminal claim blocking further commands).
Those cascading busy errors are not separate defects. Plain overshoot still refuses.

Independent service probes fail the real DB witness: persisting the rejected command produces
`revision 1→2 commands 0→1`, even though the tenant rejection is retained; suppressing error-path
claim release fires `rejected advance did not release its actual claim` before subsequent busy
errors. Independent direct-transition probes also fire: copying partially executed Go work
back before invalid_turn fails all four late-turn cases; aliasing the TS caller's input instead
of copying it fails eight overshoot/late-turn cases. These observe the actual mutable objects,
not unchanged caller-owned encoded strings. All three production files are byte-restored.

The shared six-command fixture is applied at genesis and after an accepted move in both pure
engines and the real candidate 20×20 DB composition. The pure arm uses RP-196's actual 5×5
fixture and seed 455; it does not relabel either board. All pure valid-next-advance, quit and
post-terminal controls pass. A stale descriptive comment in the old crash test is corrected
from 5×5/head x2 to its actual 6×5/head x3; no command or expectation changes.

**Bounded original designated review — CHANGES REQUIRED.** Review by: Codex.
Recorded by: Codex. Exact original Claude pure-rule span `508fe19a^..508fe19a`, and exact
original composed-platform span `e1c71d7c^..e1c71d7c`. This verdict concerns only AC5/AC6's
future/late-turn and durable rejection evidence. The old suites demonstrably admit the
conditional future-turn discard probe; the old DB witness never submits a rejected command.
Conversely, all four broad AC5 probes already fail independently in BOTH runtimes, and that
positive evidence is preserved. There is no claim the unchanged engines have those defects,
or that every other A3/A5 criterion has now received designated approval. The separately
predeclared test-only repair still requires Claude's exact-range cross-party review.

**Retained cold verification:** root `-count=1` Go passes `./decimal ./arcade ./kernel
./minigame ./replaycatalog ./production` (the ordinary production lane is not substituted
for DB evidence). Whole client: 7,164 pass / 88 intentional browser-only skips; typecheck and
Svelte report zero errors/warnings. Build, client/cosmetic/no-payment boundaries and their
negative controls, vet, byte-identical numeric vectors and full Arcade regeneration pass.
The declared Postgres selector executes BOTH the existing composed witness and the new
twelve-rejection witness verbosely PASS (0.15/0.16 s), terminal Make 0. Complete Linux browser
CI is terminal Make 0: 21,741 passes across 264 three-engine file populations, three intentional
performance skips, followed by the separate Chromium performance case (one pass / 20 filtered).

Fresh `make verify-kernel-version` passes checkout/adversarial fixture checks and fails Make 2
on unchanged pushed RP-131 hash `50a3a5141d5c4c1cd25cbf3c9c39d9e3cf8e2444` against parent
`0cf9f7a6aba4038fadcdf35e5b94f56986164af7`. No bypass, exception, false bump or rewritten history.
Production probes, both older corpora, both older fixtures, clearing artifact, numeric vectors
and kernel 0.3.146 remain byte-identical. `git diff --check` passes. This is **READY FOR CLAUDE
DESIGNATED REVIEW**, not a claim of complete green CI, hosted reliability, full AC6/AR7,
public wire or archival eligibility. The next safe accepted review is A2/AC3's four Mine Grid
rule mutants; no other authority or release scope is inferred.

**RP-197 corrective review handoff:** exact Codex span `8ea62952^..5e82b3d3`.
**READY FOR CLAUDE DESIGNATED REVIEW.** Predeclaration, shared six-command fixture, actual
Go/TS transition and real-Postgres atomicity witnesses, docs and tracking only. The existing
crash-test comment is reconciled to its actual unchanged 6×5 fixture. All production/kernel/
balance/copy/public-wire/CI-workflow bytes remain unchanged. No full A3/A5 acceptance,
self-archive, push or deployment. This checkpoint records the span; it is not a review verdict.

## 2026-10-05 — A2/AC3 rule-discrimination predeclaration (Codex)

Starting clean HEAD `5176f978`. Authority: accepted A2, AR3.2/AR3.3/AR3.5, AR7 and AC3;
original Claude range `508fe19a^..508fe19a`. This is the four named failure populations,
not a whole-A2 or public-wire review. Existing v2 literal Go outputs and independent TS
replay remain the instrument. No production, rule, RNG, content, copy, mint or wire changes
are authorized by this declaration. The first attempted append used stale context and
failed atomically without changing any file; this is the actual predeclaration.

**Before adding tests:** cold whole Go Arcade and client baselines. Independently mutate
Go/TS flood expansion to cardinal neighbors only (retain eight-neighbor adjacent counts),
allow mines in the excluded first-reveal region, and let chords detonate on flagged mines
instead of excluding them. Restore after each actual execution and identify the first
existing witness that fails. Do not add redundant tests when the existing gate discriminates.

Separately replace only Mine Grid's TS Fisher–Yates draw with next()%bound, leaving shared
SplitMix64.bound and unrelated combat tests intact. Execute the OLD whole client suite before
changing the instrument. Rejection sampling may have no rejecting draw in the existing small
board seeds; passing ordinary seeds is not proof that the loop rejects an inadmissible draw.
If this survives, file the bounded instrument finding immediately, not a runtime defect.

**Conditional supplement fixed before execution:** derive a real uint64 seed by inverting
SplitMix64 and the two published Mine Grid substream labels so the first actual shuffle draw
is zero at a non-power-of-two eligible-cell bound. Use a legal board/preset from the existing
fixture, not changed production bounds. Independently verify inversion by executing actual
Go/TS RNG: zero is below the computed threshold, the second draw is accepted, and both normal
engines produce the same exact mine list and real create/choose/reveal/quit outputs. Record
literal Go bytes, not a test-generated self-oracle. Require the complete real entry path,
not just a mocked RNG or a high-bound combat-only test. A test-only scripted-draw control may
explain rejection consumption, but may not substitute for the real-seed population.

The Mine Grid-only modulo mutant must fail the added witness; any supplemental observation
validator also needs a failing case. Restore all probes, run cold Go/client/type/build/root
boundaries/vet/vectors/corpus, full three-browser and existing real-Postgres Arcade tests.
Historical RP-131 remains separate, no guard bypass or CI edit. Report bounded original
verdict and exact Codex supplement span; retain Claude cross-party review, no self-archive,
push or deployment. If all existing instruments already discriminate, record that evidence
instead of manufacturing a correction. All full 1.0 obligations remain.

**Executed old instruments:** unchanged cold Go Arcade and all 7,164 client cases pass.
Cardinal-only flood (with unchanged eight-neighbor counts), excluded-region placement and
flagged-mine chord probes fail independently in BOTH runtimes. The first actual TS corpus
differences are large-board first reveal (flood/placement) and honest chord at cell 29 (flagged
mine 20 incorrectly detonates). Go regeneration/policy checks also fail. No extra tests are
needed for those already-discriminating populations; collateral raw-fixture needle failures
are not cited instead of the primary gameplay differences. Probes are restored.

Mine Grid-only TS next()%bound survives the entire OLD suite: 7,164 pass / 88 intentional
skips, Make 0. Shared combat sampling is unchanged. This is RP-198's missing rare-draw
population, not a production RNG defect. Independent read-only inverse calculation derives
seed `15581846558861750132`: actual run seed should be `2387092019343320515`, first shuffle
draw zero, next `16294208416658607535`. On the existing large 9×9/10-mine preset, first cell
40 excludes nine cells, leaving bound 72 and threshold 16. Verify these with actual engines
before accepting the seed or its outputs. The preliminary calculation is not yet evidence
of real placement or byte parity. All production probes remain restored.

**Executed real-seed construction:** retained Go inversion derives the independent seed
above, then actual two-substream RNG verifies run seed, zero below threshold 16 at bound 72,
accepted draw `16294208416658607535` and consumption of both initial draws. The actual registry
creates/chooses large/reveals cell 40/quits, then refuses a terminal retry: four attempts /
five literal snapshot observations, all nullable result bytes retained. Its terminal mines
are `[4,8,11,16,28,33,45,46,51,63]`. No RNG substitution, synthetic snapshot or changed board.

Independent TS executes the same real RNG, its own placement and actual create/apply entries,
byte-comparing every Go snapshot/result. Unchanged baseline is 7,167 passes / 88 intentional
skips, zero typecheck/Svelte errors/warnings. The Mine Grid-only modulo probe now fails Go's
fresh literal artifact comparison and TWO new TS cases: placement differs and first-reveal
bytes differ. All 7,164 old client cases continue passing under that probe; the new real-seed
population is exactly the missing discrimination, not unrelated collateral. Separately
removing TS Bound's threshold fails the RNG-consumption case (actual 0 vs accepted residue 7)
and both placement/transition comparisons. Forging only the witness version fails Go's cold
regeneration check. All engine/RNG/artifact probes are restored exactly.

Root `arcade-corpus` / `arcade-corpus-check` include this deterministic real-seed witness,
cold, without changing a workflow or regenerating older population bytes. No balance/content
artifact changes: the existing 9×9 fixture is unchanged. The larger uint64 seed is a string
in the artifact and BigInt in TS, never a lossy JavaScript Number. Kernel remains 0.3.146.

**Bounded original designated review — CHANGES REQUIRED.** Review by: Codex.
Recorded by: Codex. Exact original Claude span `508fe19a^..508fe19a`, concerning only AC3's
four named rule-discrimination populations. The three gameplay mutants already fail both
runtimes and that positive evidence is preserved. The named Mine Grid-only TS sampling
mutant demonstrably survives the OLD complete suite. RP-198 supplies its missing bounded
population locally; it still needs Claude's designated correction review. No unchanged-
production RNG defect or complete A1/A2/AR7/public-integration approval is inferred.

**Retained cold verification:** root Go `-count=1` passes `./decimal ./arcade ./kernel
./minigame ./replaycatalog ./production`. Whole client: 7,167 passes / 88 browser-only skips;
typecheck/Svelte zero errors/warnings. Build, client/cosmetic/no-payment boundaries and their
negative controls, vet, byte-identical numeric vectors and full three-artifact Arcade
regeneration pass. Both declared real-Postgres Arcade witnesses execute verbosely PASS (the
new rejection population 0.62 s; original composition 0.41 s), terminal Make 0. This remains
DB/library evidence, not public wire. Complete Linux three-browser CI is terminal Make 0:
21,750 passes across 267 file populations / three intentional performance skips, then the
separate Chromium performance case (one pass / 20 filtered). Terminal metadata retained.

The fresh history target passes checkout/adversarial fixtures, then fails Make 2 on unchanged
pushed RP-131 hash `50a3a5141d5c4c1cd25cbf3c9c39d9e3cf8e2444` against parent
`0cf9f7a6aba4038fadcdf35e5b94f56986164af7`. No false bump, bypass, exception, rewrite or CI edit.
All production probes, older corpora/fixtures, numeric vectors and kernel 0.3.146 remain
byte-restored. `git diff --check` passes. **READY FOR CLAUDE DESIGNATED REVIEW** of the bounded
test-only sampling supplement, not full A2/AR7, complete green CI, public integration or
archival eligibility. Next safe accepted review is A1/AC1's named artifact-loader negatives
and their existing A4 catalog-chain consumers; no new owner wire/copy/mint authority.

**RP-198 corrective review handoff:** exact Codex span `be0c0e00^..859f5216`.
**READY FOR CLAUDE DESIGNATED REVIEW.** Predeclaration, deterministic inverse real-seed
generator, literal Go witness, independent TS RNG/placement/entry replay, root generation
aliases, docs and tracking only. Older corpora/fixture and every production/kernel/balance/
copy/public-wire/CI-workflow byte remain unchanged. No whole-A2/AR7 acceptance, self-archive,
push or deployment. This hash-pinning checkpoint is not a designated verdict. Its first
append attempt used incomplete line context and failed atomically with no file changes;
the exact-source-context append here is the retained checkpoint.

## 2026-10-05 — A1/AC1 loader and A4 chain predeclaration (Codex)

Starting clean HEAD `9707d48e`. Authority: accepted AR1.2, AR3.1, AR4.1, AR-P1/P2 and AC1;
original Claude loader span `508fe19a^..508fe19a`, chain span `e1c71d7c^..e1c71d7c`.
Read-derived inventory: Go loader tests contain more negative categories than TS despite
the latter's "same loader defects" title. Neither explicitly submits two stage rows with
nonascending tiers; both original chain tests submit missing toys, not a definition bound
to a wrong engine. These are questions about the instrument, not proven runtime defects.

**Old gates first:** cold Go Arcade/replaycatalog and whole client baseline. Independently
remove only stage-tier ascending refusal in each loader, leaving all other bounds/ID/toy/
copy checks intact; execute OLD populations, restore and record any survival immediately.
Then submit the named AC1 population in actual Go/TS loaders and chain consumers: unsorted
presets, too many mines, unknown toy/missing definition, wrong-engine definition, root/nested
extra keys, equal/descending tiers. Positive controls load unchanged candidate and fixture,
inclusive bounds and structurally valid ascending test-only stage rows. Multi-stage test
rows neither mint content nor authorize later-stage gameplay or owner copy.

**Conditional evidence repair:** only if named failure cases are absent or a severing survives,
add matched Go/TS actual-input tests to the existing loader/chain witnesses. Do not produce
a second catalog authority, weaken assertions or infer a whole-A1/A4 approval. Re-run the
same ascending mutant against the new gate; independently sever existing preset sorting,
mine cap, exact-key and stage-to-definition checks to identify real readers and refusals.
Binding controls must exercise complete current bundles with freshly computed hashes,
not stale-hash rejection or a frontend-only fragment. Inspect all setup/errors.

The read-derived Go integer-defaulting question (null min_tier becoming zero) is separately
empirical within AR1.2: submit an otherwise valid actual candidate with only that value
changed, compare TS, and retain legal zero as its control. If either admits malformed input,
file the executed finding and predeclare its accepted-contract runtime correction separately
before changing watched source; honest kernel identity required. Raw duplicate/token/depth
parity beyond named cases is not silently claimed or expanded here.

**Exit:** cold Go/client/type/build/boundaries/vet/vectors/corpus, complete Linux browser and
declared Postgres Arcade selectors; report unchanged historical RP-131 separately. Bounded
original verdicts and exact independent-review range, no self-archive/push/deployment. No
wire/copy/mint, content-bound or mechanics changes; all full 1.0 obligations remain.

**Executed old-gate finding:** unchanged cold Go Arcade/replaycatalog and all 7,167 client
cases pass. Removing only ascending stage-tier refusal independently from both loaders
survives those same entire populations (Go 18.24 s / Make 0; client 7,167 / Make 0).
RP-199 records the missing multiple-stage instrument. All production probes are restored;
no unchanged-runtime ordering failure has been alleged. Continue the conditional test-only
supplement, preserving existing sorting/cap/key and binding controls.

**Separate executed runtime finding RP-200:** a temporary diagnostic submits the actual Go
candidate with only min_tier=null. It logs `MALFORMED INPUT ADMITTED: null min_tier normalized
to 0`; the legal-zero control loads. Native Node executes the actual TS loader with the
generated canonical COPY_KEYS: zero loads, null throws SyntaxError. The diagnostic's Go exit
0 is not negative acceptance proof; it explicitly reports an admission. The temporary test
is removed after execution. No production change retained here. RP-200's separately
predeclared watched correction is next, not silently mixed into RP-199's test-only repair;
do not use exactInteger in a way that newly rejects Go's legal negative-zero spelling.

**RP-199 retained instrument:** both actual loaders now execute the same 22 negative
categories, including equal/descending stage tiers and every original Go category missing
from TS. Both keep valid candidate/fixture and inclusive-bound controls. Structurally valid
test-only tiers 0/9 load; the real TS stage selector chooses the proper row at tiers 8/9.
These rows do not ship content, new gameplay, copy or an unlock policy. Freshly hashed Go/TS
bundles also reject exactly one actual definition rebound to pitch, not a corrupted hash.

**Discrimination/restoration:** removing ascending refusal now fails both new tier cases;
independent preset-sort, mine-cap and exact-key policy probes each fail the actual matching
negative in Go and TS. Go's exact-key probe disables BOTH its key-count and strict-decoder
barriers, disclosed as a two-barrier policy severing rather than pretending one layer owns
the whole refusal. Binding removal makes both missing-definition AND wrong-engine bundles
load; both tests fail. The first TS probe stops on the existing missing-toy assertion and
dumps its whole resolved bundle; the test now executes all negative arms and reports only
admitted case names. The rerun identifies both cases explicitly. Go uses individual
subtests, also demonstrating both failures. No runtime correction is retained; all four
production source files are byte-restored. This evidence is not a full raw JSON/depth audit.

**Bounded original designated review — CHANGES REQUIRED.** Review by: Codex.
Recorded by: Codex. Original loader span `508fe19a^..508fe19a`; original chain span
`e1c71d7c^..e1c71d7c`. Verdict concerns AC1's missing tier-order/wrong-engine populations
and misleading matched-loader coverage claim. It does not reject the valid existing sorting,
cap/key or missing-toy controls, nor infer that production accepts wrong bindings unchanged.
RP-199's test-only repair needs Claude's exact-range review. The separate actual null-tier
admission RP-200 remains OPEN pending its own watched correction; do not close it via prose.

**RP-199 final cold verification:** all six selected Go packages pass with `-count=1`;
the whole client passes 7,168 cases / 88 intentional browser-only skips. Typecheck/build,
client/cosmetic/no-payment boundaries and their controls, vet, unchanged numeric vectors and
all three Arcade artifact regeneration checks pass. The declared real-Postgres production
selector executes BOTH Arcade integration functions verbosely, not skipped (rejected-advance
and original composed witnesses); this is DB/library proof, not socket/API acceptance.
The complete Linux Chromium/Firefox/WebKit target exits 0 with 21,753 passes / three
intentional performance skips, then its separate Chromium performance case passes (20
filtered). Terminal metadata retained. The cold history guard exits Make 2 at unchanged
pushed RP-131 `50a3a5141d5c4c1cd25cbf3c9c39d9e3cf8e2444` versus parent
`0cf9f7a6aba4038fadcdf35e5b94f56986164af7`, after checkout/adversarial controls pass.
No exception, false bump, history rewrite or workflow edit. Production probes are restored,
kernel stays 0.3.146 and `git diff --check` passes. Ready for Claude's bounded designated
review, not full A1/A4 acceptance or archival. Next is separately predeclared RP-200.

**RP-199 exact corrective handoff:** `8b00b6e8^..3b9fd0d8`, READY FOR CLAUDE DESIGNATED
REVIEW. This checkpoint is not that verdict and does not archive the work.

## 2026-10-05 — RP-200 nonnullable catalog stage tier correction predeclaration (Codex)

Starting committed test-evidence source `3b9fd0d8`. Authority: accepted AR1.2/AR-P1/AC1's
typed integer stage tier and both-runtime artifact chain, not a new mechanic. Prior executed
diagnostic proves actual Go admits null as zero while TS refuses it; retain an acceptance
test that fails on unchanged Go before fixing the loader. Test actual candidate AND fixture,
and complete freshly hashed composition, not hash drift. Keep legal zero, whitespace and
negative-zero spelling controls. No fixture, content, copy, public wire or mint changes.

Repair only Go's raw stage tier admission using a nonnullable integer decode before the
ordinary typed catalog decode. Do not reuse canonical-only exactInteger, which would newly
reject legal JSON -0. The watched runtime behavior change requires kernel 0.3.146 → 0.3.147
in all three identity files; numeric vectors remain byte-identical. Demonstrate removing
this guard admits null and fails the new tests, then restore exactly. Run cold selected Go,
whole client/type/build/vet/boundaries/vectors/corpus, declared real-DB Arcade selector and
full Linux browser checks. Report historical RP-131 separately without bypass or rewrite.
Update canonical docs and records, pin a distinct Claude review span, no self-archive.
This bounded null correction does not claim exhaustive duplicate/token/depth parity.

**Executed RP-200 red/fix/severing:** on unchanged runtime, both candidate and fixture
subtests fail `null stage tier must refuse, got <nil>`; the freshly hashed full Go chain
also fails its named null-tier population. Actual TS direct-loader/chain cases pass (96
focused tests). Pointer inspection before the ordinary decode fixes both Go populations.
Removing ONLY the nil guard then reproduces both direct failures and the full-chain
admission, while the chain's other six refusal arms remain green. The guard is restored.
Legal 0/-0/whitespace controls pass for both artifacts in both runtimes. This is an actual
loader behavior correction with honest kernel 0.3.147, not a balance/schema/mint change.

The first docs/plan patch failed atomically on stale plan context; no partial changes.
The corrected-context patch is retained. No owner text or production fixture changed.

**RP-200 final cold verification:** six selected Go packages pass with `-count=1`; the whole
client passes 7,169 / 88 intentional browser-only skips. Type/Svelte check reports zero
errors/warnings; build, boundaries/negative controls, vet, unchanged numeric vectors and
all three Arcade regeneration checks pass. Both real-Postgres Arcade integration functions
execute verbosely and pass, not skipped (twelve rejected advances plus valid play/resolution,
and the original composed function). This remains DB/library, not public socket/API proof.
Full Linux Chromium/Firefox/WebKit verification exits 0: 21,756 / three intentional performance
skips, followed by the separate Chromium performance case (one pass / 20 filtered). All
terminal metadata retained. The complete kernel history target again exits Make 2 at the
same pushed RP-131 hash/parent, after checkout/adversarial controls pass. No exception, bypass,
rewrite or workflow change; this honest current behavior bump does not erase old history.
`git diff --check` passes. Ready for Claude's bounded designated review; no self-approval,
archival, full A1/A4/raw grammar, mint, push or deployment claim.

Next safe accepted work: predeclare the remaining A4/AC7 resolver/start composition audit.
Execute actual tenant resolution and Pitch-less Arcade starts, retain Pitch positive controls
and unknown-pair refusal, and sever the specific AR-P2/AR-P3 consumers before inferring
completion. Owner-gated public wire/copy/mint and full 1.0 obligations remain unchanged.

**RP-200 exact corrective handoff:** `df0ed871^..6f5715dc`, READY FOR CLAUDE DESIGNATED
REVIEW. Includes separate predeclaration, red/fix/severing loader and complete-chain evidence,
the eight-line runtime refusal, honest kernel 0.3.147 and canonical/tracking reconciliation.
The original RP-199 review remains separately `8b00b6e8^..3b9fd0d8`. Neither checkpoint
substitutes for Claude's verdict or authorizes archival. Work proceeds on accepted A4/AC7.

## 2026-10-05 — A4/AC7 resolver and atomic-start predeclaration (Codex)

Starting clean HEAD `13cfe371`. Authority: accepted AR-P1/P2/P3 and AC7; original Claude
A4/A5 range `e1c71d7c^..e1c71d7c`. Inspect actual bundle resolution and both start methods.
Read-derived question: AC7's Pitch-less bundle conflicts with MA-C15's explicit owner-ruled
`minigame_api → pitch` dependency, enforced in Go/TS composition and replay validity.
Submit otherwise complete freshly hashed bundles with only Pitch removed in both runtimes;
keep unchanged complete-bundle positive controls. A refusal is a routed contract conflict,
not permission to weaken the earlier ruling or relabel a bypassed test as integrated proof.

Cold old Go Arcade/production/replaycatalog and declared Postgres Arcade plus atomic Pitch
start populations first. Independently sever each Arcade content-resolver pair and execute
the existing chain gate; preserve working witnesses. Then temporarily reject only Arcade
definitions in StartMinigameAPISession and run the old populations to check whether they
exercise that actual atomic start entry (the original Arcade witness uses the older start).
Restore all probes. If the new-entry population is absent, retain test-only real-Postgres
starts for BOTH toys from actual Founder v21 streams, server-owned seed/sequence, persisted
genesis/state and idempotent retries; preserve existing v20 helper behavior and Pitch proof.
Re-run the same Arcade-only start severing against the supplement; no acceptance based on
test names, skipped DB cases or handcrafted expected genesis.

No runtime, kernel, corpus, balance/content, copy, public-wire/mint or CI changes. This proves
an internal coordinator/DB seam, not public socket/API exposure, Pitch-less acceptance or
all AC7. Route any executed contract conflict to the ruling authors/owner. Final cold root
checks and exact Claude review handoff; historical RP-131 remains separately red/unbypassed.

**RP-201 executed contract conflict:** otherwise complete candidate catalogs load in both
actual runtimes. Removing only Pitch and recomputing the complete hash refuses in Go
(`invalid replay inputs`) and TS (SyntaxError). Temporary diagnostics
are removed after execution. This verifies the MA-C15 dependency and disproves an unqualified
AC7 Pitch-less claim; it is not authority to change either owner-authored contract. Route to
the ruling authors/owner for a named dependency amendment or AC7 body reconciliation.
No synthetic resolver or manually invalid bundle may substitute for the missing population.

**RP-202 old-gate finding:** cold Go Arcade/production/replaycatalog and the three declared
real-Postgres functions pass. Removing Mine Grid's or Snake's resolver arm independently
makes the existing chain test fail, so those witnesses are preserved. Temporarily rejecting
only Arcade definitions in StartMinigameAPISession survives the same entire Go population
and all three DB functions (both old Arcade functions and atomic Pitch start). The Arcade
population never calls that method. Probe restored exactly. Proceed with the predeclared
test-only internal-start supplement, not a public route or RP-201 dependency change.

**RP-202 retained supplement:** both Mine Grid and Snake start through the actual atomic
coordinator using real Postgres, Founder v21 and the unchanged complete candidate bundle.
The old helper continues to seed v20 by default; an explicit version helper supplies v21
without rewriting old setup semantics. Stored seed must match the server sequence/run
derivation; canonical persisted genesis/state and create receipt must match actual pinned
tenant creation. Retry supplies a different proposed session ID but returns identical bytes;
SQL retains one session/receipt, Founder advances once to revision 2/sequence 1, Company
revision stays 1 and actual Founder replay verifies. Both cold subcases pass.

Repeating precisely the old Arcade-only atomic-start rejection now fails BOTH named subcases
with `invalid production intent`. Source restored exactly; no production/kernel/content/
schema/copy/wire/CI bytes retained. The resolver's original both-arm controls also fail
independently and remain unchanged. RP-201 is still an unruled dependency conflict.

**Bounded original designated review — CHANGES REQUIRED.** Review by: Codex.
Recorded by: Codex. Original A4/A5 span `e1c71d7c^..e1c71d7c`: the original DB witness does
not exercise AR-P3's named atomic start entry for either Arcade toy. This does not allege
a runtime failure in unchanged code, invalidate the older method's actual gameplay proof,
or overrule MA-C15; RP-201 is routed separately. RP-202 requires Claude's exact corrective
review before acceptance; neither full A4/A5 nor public AC8 is promoted.

**RP-202 final cold verification:** all six selected Go packages pass with `-count=1`.
Whole client: 7,169 passes / 88 intentional browser-only skips; zero TS/Svelte errors or
warnings. Build, boundaries/negative controls, vet, unchanged numeric vectors and all three
Arcade regeneration checks pass. The declared real-Postgres selector executes FOUR functions
verbosely: both old Arcade witnesses, both new atomic-start subcases and atomic Pitch start;
all pass, none skipped. Full Linux browser target exits 0 with 21,756 passes / three
intentional performance skips across 267 populations, then separate Chromium performance
passes (one / 20 filtered). Terminal metadata retained. Kernel history again exits Make 2
at unchanged pushed RP-131 hash `50a3a5141d5c4c1cd25cbf3c9c39d9e3cf8e2444` against parent
`0cf9f7a6aba4038fadcdf35e5b94f56986164af7`, after checkout/adversarial controls pass.
No exception, false bump, bypass, rewrite or CI edit; kernel remains 0.3.147.
`git diff --check` passes; all production probes and temporary diagnostics are restored/removed.
READY FOR CLAUDE DESIGNATED REVIEW of this test-only supplement; no full A4/A5/AC7/AC8
acceptance or archival. D-020's contract choice is presented to the owner asynchronously;
no reply or ruling is inferred. Next safe accepted work audits remaining A5/AC8–AC10
gameplay/resolution/Exit/Soul evidence, preserving the blocked public-wire and mint boundary.

**RP-202 exact corrective handoff:** `deda7e1e^..15fe1ccc`, READY FOR CLAUDE DESIGNATED
REVIEW. Test-only version-preserving helper and real-Postgres both-toy atomic-start witness,
docs/tracking, and the separate RP-201/D-020 contract-conflict routing. Every production,
kernel, corpus, balance/content, copy, public-wire and CI-workflow byte remains unchanged.
No self-approval, full A4/A5/AC7/AC8, archival, push or deployment. The hash-pinning checkpoint
is not a designated verdict. Next accepted evidence audit remains on the execution queue.

## 2026-10-05 — A5 zero-reward and atomic Soul-gate predeclaration (Codex)

Starting clean HEAD `2ead3a1d`. Authority: accepted AR1.5/AR1.6 and AC8/AC10;
original Claude A5 range `e1c71d7c^..e1c71d7c`. Current new atomic-start witness is
`15fe1ccc`; it supplies normal-Soul starts, not atomic low-Soul refusals. The older
low-Soul Arcade population calls a different method and covers only Mine Grid. The older
zero-credit assertions omit cap_reason_key even though AR1.6/AC8 name its absence.
These are instrument questions, not claimed runtime defects.

Cold old Go Arcade/production/replaycatalog and declared real-Postgres Arcade plus atomic
Pitch start population first. Independently (never combined) disable ONLY the atomic start
Soul check, then force ONLY Arcade resolution receipt cap_reason_key to the declared cap key
while keeping zero credit/forfeit, neutral grades/rating and replay input unchanged. Execute
old populations and restore. Record survival/fired controls before any test repair.

If absent populations survive, retain test-only both-toy low-Soul atomic starts over real
Founder v21/Company streams, exact named error, empty response, unchanged full state/revision/
sequence and zero session/create-receipt rows, with normal-Soul positive controls. Add the
explicit no-cap-reason receipt assertion to the existing zero-credit composed witness; do
not replace its real engine commands, certified resolution or retry. Repeat each independent
probe against the respective new witnesses, restore all runtime bytes. A quality attendance
stamp is bookkeeping, not evidence that the neutral economic grade changed; do not assert
whole quality-map immutability where platform C40 explicitly updates that stamp.

AC9's direct before/after Wind Down and cross_gate population is absent in the current
Arcade witness, which observes only repository activity. Keep that full requirement open
for the next separately declared genuine Exit-action witness, not a repository-boolean or
Tier-0-ineligible substitute. Public socket/receipt exposure and the C2 wire conflict remain
blocked; internal proof cannot close public AC8. RP-201/D-020 remains unruled.

No production/kernel, fixture/content, copy, wire/mint or CI changes. Final cold root Go/
client/type/build/boundaries/vet/vectors/corpus/browser and real-DB selectors, exact Claude
review span and tracking closeout. Historical RP-131 remains red without bypass or rewrite.

**RP-203 discovery:** cold original Go and all four actual DB functions pass. Removing ONLY
the atomic start human_hobby check survives the same four DB functions; the selected Go
population is still running at this discovery checkpoint. The old Mine Grid Soul negative
calls the unaffected older method, and normal-Soul atomic starts cannot discriminate this
refusal. Record the gap now, not a runtime defect in unchanged code. Await the same live Go
handle before restoring or changing source; conditional both-toy low-Soul supplement follows.

The same live Go Soul probe finishes green for all three complete packages (production
34.431 s), confirming survival in both old populations. Atomic-start source is restored
exactly before beginning the independent cap-reason probe; no combined mutant is used.

**RP-204 discovery:** forcing only Arcade resolution receipt cap_reason_key to the existing
declared cap key survives all four old real-Postgres functions, preserving zero credit/
forfeit, grade/rating and retry. Selected complete Go is still running at discovery. The
existing witness never inspects that field; do not call this a current runtime defect.
Await the same live Go handle, restore, then add the predeclared receipt-field assertion.

The same live old cap-reason Go probe finishes green in all three complete packages
(production 34.481 s). The receipt producer is restored exactly before retaining tests.
All four declared DB functions pass cold with the new tests on unchanged production.

**RP-203 retained witness and discrimination:** low-Soul atomic starts now execute for BOTH
toys on actual Founder v21/Company streams. Require the named error, empty response, unchanged
complete encoded state and revision/hash for both streams, and zero SQL session/create-receipt
rows. Existing normal-Soul starts, server genesis/sequence, retry and replay remain positive.
Removing only the atomic human_hobby gate fails BOTH locked subcases: admitted receipts,
changed Founder state/revision, and one persisted session/receipt each. Both normal starts
still pass. The source is restored exactly before the independent receipt probes.

**RP-204 retained witness and discrimination:** the existing real-play/resolution/retry
population now requires explicit empty cap_reason_key on each actual zero-credit receipt.
Independent Mine Grid-only and Snake-only producer corruptions each fail on that toy's named
no-cap-reason diagnostic, with credit/forfeit and other receipt fields unchanged. Both producer
probes are restored exactly; no production, kernel, balance/content, corpus, copy, wire or CI
change is retained. The quality attendance timestamp remains legitimate C40 bookkeeping,
not an economic-grade mutation or grounds for a full-map immutability assertion.

**Bounded original designated review — CHANGES REQUIRED.** Review by: Codex.
Recorded by: Codex. Original Claude A5 span `e1c71d7c^..e1c71d7c`: AC10 lacks the actual
atomic low-Soul population; AR1.6/AC8's no-cap-reason field is unobserved. The executed old
populations survive both independent corruptions. These are instrument defects, not claimed
runtime defects in the unchanged product. The test-only corrections require Claude's exact
cross-party review; no full A5/public AC8 promotion, archive, push or deployment follows.
AC9 still needs genuine eligible Wind Down/cross_gate before/after-quit actions, not the
existing repository activity booleans. RP-201/D-020 and the public-wire conflict remain open.

**RP-203/204 final cold verification:** all six selected Go packages pass with `-count=1`
(`decimal`, `arcade`, `kernel`, `minigame`, `replaycatalog`, `production`). Root typecheck
reports zero errors/warnings; whole client has 7,169 passes / 88 intentional browser-only
skips. Build, boundaries including negative controls, vet, unchanged vectors and all three
Arcade regeneration comparisons pass. The declared Postgres selector executes FOUR functions
verbosely: rejected-advance atomicity, real composed play/resolution, atomic Arcade starts
(both low-Soul negatives and both normal starts), and atomic Pitch start. All pass, none skip.
Full cold Linux browser target exits 0: 21,756 passes / three intentional performance skips
across 267 populations, followed by separate Chromium performance (one pass / 20 filtered).
Terminal statuses are observed, not inferred from partial logs.

The full kernel guard exits Make 2 at the same pushed RP-131 commit
`50a3a5141d5c4c1cd25cbf3c9c39d9e3cf8e2444` against parent
`0cf9f7a6aba4038fadcdf35e5b94f56986164af7`, after checkout/adversarial controls pass.
No false bump, historical exception, rewrite or CI change; kernel stays 0.3.147. Production
probe diff is empty and `git diff --check` passes. READY FOR CLAUDE DESIGNATED REVIEW of the
bounded test-only correction, not full A5/AC8/AC9/AC10 or archival. Next accepted work is the
separately predeclared genuine Exit-action population; all later 1.0 obligations remain.

**RP-203/204 exact corrective handoff:** `d798e709^..83008a9f`, READY FOR CLAUDE
DESIGNATED REVIEW. Covers the committed predeclaration and test-only low-Soul atomic-start/
zero-cap-reason receipt corrections, plus docs/tracking. Runtime/kernel/content/copy/public
wire/CI bytes are unchanged. This hash-pinning checkpoint is not a verdict or archival gate;
no full A5/public AC8/AC9/AC10 acceptance, push or deployment. The next genuine Exit-action
audit stays separately scoped; RP-201/D-020 is still unruled.

## 2026-10-05 — AC9 actual Exit-action predeclaration (Codex)

Starting clean HEAD `e58260ff`; previous goal turn made progress (`83008a9f` test-only repairs,
exact review handoff pinned). Authority: accepted AR1.8/AC9, API MA-C12's reject-Exit ruling,
and the current T0–T1 curriculum contract. Original Claude A5 span `e1c71d7c^..e1c71d7c`.
The old Arcade witness checks repository activity without calling any Exit intent. Current
curriculum drives the first ending after the first gate and 900,000 attended milliseconds;
AR1.8's cross_gate reference predates that integration. Do not infer current behavior from
the historical reference or treat ordinary progression as an Exit without executing it.

First run cold complete Go Arcade/production/replaycatalog and declared real-Postgres Arcade,
atomic Pitch start, Typer Exit and current curriculum populations. Independently sever only
the replay Exit active-session check; execute the old population, recording existing working
controls honestly rather than claiming a missing Arcade row means all checks are vacuous.
Restore it before constructing tests.

Build a test-only, freshly hashed complete current-epoch bundle retaining every original
artifact/definition/tenant and adding only the two accepted Arcade candidate rows/artifact.
No live mint, Typer substitution, curriculum deletion, changed gate/clock/economy or synthetic
activity resolver. Use actual Postgres, repository, platform tenants and production Handle.
Establish eligible no-session controls for Wind Down, cross_gate and the default due company
action before asserting active-session refusals. Distinguish an ordinary gate transition from
an Exit; record exact outcomes, error, run sequence and persisted state. Do not choose a later
tier merely to get a green substitute for the current first ending.

If the actual intended due Exit path is supported, retain both-toy active refusal, real quit
from setup/playing, and released Exit with replayable persisted histories. If current
cross_gate fails even without a session or conflicts with a prior explicit contract, record
the failure immediately and route the unresolved scope instead of silently editing owner
text or broadening runtime authority. Test-only instrumentation may be retained; any genuine
production defect requires its own predeclared authorized correction. Repeat independent
guard and per-toy quit severings against the added controls; restore all production probes.

Final cold selected Go, client/type/build/boundaries/vet/vectors/corpus, real-DB and complete
Linux browser targets; historical RP-131 remains separate and unbypassed. Exact corrective
Claude range and docs/ledger/queue/roadmap closeout. No acceptance checkbox promotion, full
A5/AC9/public AC8, archive, copy/content mint, wire/schema, CI or push/deployment changes.

Baseline cold selected Go and the corrected SIX-function real-Postgres selector pass.
The first selector spelling `TestTyper.*Exit.*Integration` matched no Typer function; it
executed five functions and is not Typer evidence. Replacing that arm with the actual
`TestTyperComposedIntegration` executes its real Wind Down before/after end_run population
alongside the three Arcade functions, atomic Pitch start and current automatic curriculum.
No source change occurred between baselines. The broad guard probe now runs independently.

The broad replay Exit-guard severing is caught by existing whole-Go literal replay output,
atomic Pitch start (exact detail changes from active-session to tier), and Typer's actual
Wind Down (it wrongly applies). All original Arcade functions and the no-session automatic
curriculum control still pass. This is positive evidence for existing controls, not old
whole-suite survival. Both tool handles exit Make 2; source restored exactly before new tests.

**RP-205 discovery:** the first current-epoch-plus-Arcade execution passes Wind Down and
ordinary due company-action controls and both toys' block/quit/released-Exit cases. The
due cross_gate control returns `invalid production engine state` with no receipt. It names
the already-crossed first gate, so this is NOT yet an eligible gate-transition positive or
sufficient runtime-defect evidence. Record now, then execute a genuinely eligible ordinary
gate control before/after the attendance threshold. Keep any later-tier diagnostic distinct
from the default first-ending acceptance population. No production correction is authorized
by this ambiguous first control alone.

**RP-205 confirmed paired runtime failure:** a genuinely eligible current-artifact
`gate.t2_to_t3` transition at Tier 2, cash `1e10`, first gate already crossed, run 1 and no
Founder Exit succeeds before the threshold (run remains 1, Tier becomes 3). The otherwise
matched due state refuses with `invalid production engine state` and no receipt. This is
a diagnostic later-tier population, NOT a default first-hour acceptance replacement. With
the actual route projector and resolver the positive still passes and due case still fails.
The first pair lacked WithRouteCatalogs, so its route-runtime error was a setup mistake;
the next actual-projector attempt passed the wrong resolver interface and did not compile.
Both are disclosed and corrected, not product defects or acceptance evidence. The live
`applyLoggedExit` explicitly refuses CrossGate before freezing the curriculum branch.
Separate accepted-contract/runtime replay correction is required; none belongs in this
test-only repair. Remove the temporary failing due diagnostic after recording its exact
population; retain successful actual due Wind Down/company-action evidence without full AC9.

**RP-206 discovery:** all original Arcade functions survive the broad Exit-guard severing,
while existing Pitch/Typer/replay gates catch it. The specific Arcade population is absent,
not the global gate. New current-curriculum tests actually block both toys' Wind Down and
ordinary due company actions, then execute real quit/resolution and released Exit with
identical retry and verified Founder history. Phase/claimed-state and firing controls follow.

**RP-206 retained population:** eight subcases pass on unchanged production: two genuine
no-session due-Exit controls, then Wind Down and default ordinary company action for Mine
Grid setup/playing and Snake playing. Each actually starts through the atomic coordinator,
requires exact rejected/not_eligible/minigame_session_active at both active and claimed status,
compares full Founder/Company state and revisions/hash plus session state/genesis/result/token
and complete SQL command rows, executes real quit/resolution, and then persists the first
ending/run 2. Retry bytes match and both Founder and Company histories verify.

Test setup mistakes are disclosed: an incorrectly named Status type did not compile; using
the pre-start clock for later resolution could precede the atomic start's DB-authored Fiscal
clock and caused Fiscal-sweep refusal. Resolution now uses an actual later server time.
The Exit-only historical helper included Founder events in nonterminal Company replay,
causing state_divergence; the new test reads Company events for those commands and preserves
both-stream event order for actual Exit entries. No production policy or replay verifier
was changed to make these setup errors green. Two cold executions of all eight now pass.

The same broad Exit-guard severing now fails all SIX actual Arcade action/phase cases at
the active-session refusal, while both eligible no-session Exit controls pass. Repeated
probe source restored exactly before isolating claimed-status visibility in the actual
repository resolver. This next probe excludes only claimed, never active, in its read-only
SQL predicate; it must reach and fail the later claimed-state observation.

The claimed-only visibility probe fails all SIX cases specifically at the claimed-session
rejection, after active-session checks and real quits passed. No-session controls remain
green. Repository source restored exactly; independent per-toy quit probes follow.

Independent Mine Grid-only quit removal fails all FOUR setup/playing action cases at the
actual terminal requirement; Snake and no-session controls pass. Restored it before removing
only Snake's quit: both Snake cases fail at the actual terminal requirement while all four
Mine Grid and no-session controls pass. All engine probes now restored exactly. Every retained
acceptance assertion remains; no permissive fake terminal, content or public route was added.

**Bounded original designated review — CHANGES REQUIRED.** Review by: Codex.
Recorded by: Codex. Original Claude A5 span `e1c71d7c^..e1c71d7c`: repository booleans do
not exercise AC9's actual Exit/quit/released-Exit behavior. RP-206 supplies the missing
current-curriculum internal action population, not public wire/default UI or full AC9.
RP-205's separately confirmed due-cross-gate runtime error remains open for its own
predeclared accepted-contract/replay correction. Existing broad guard witnesses are credited,
not invalidated. Claude must review the exact Codex supplement; no self-approval or archival.

**RP-206 final cold verification:** six selected Go packages pass with `-count=1`; production
finishes at 61.889 s, not a guessed completion from its earlier package output. Root typecheck
has zero errors/warnings; whole client 7,169 passes / 88 intentional browser-only skips.
Build, boundaries with negative controls, vet, unchanged vectors and all three Arcade
regeneration comparisons pass. Declared real Postgres executes SEVEN functions verbosely:
four Arcade functions (including the new eight-subcase Exit population), atomic Pitch start,
Typer composed Exit and current automatic curriculum. All pass; none skips. Complete cold
Linux browser target exits 0 with 21,756 passes / three intentional performance skips across
267 populations; separate Chromium performance passes (one / 20 filtered).

The complete kernel guard still exits Make 2 at pushed RP-131 hash
`50a3a5141d5c4c1cd25cbf3c9c39d9e3cf8e2444` against parent
`0cf9f7a6aba4038fadcdf35e5b94f56986164af7`, after checkout/adversarial controls pass.
Kernel remains 0.3.147; no historical exception, false bump, rewrite or CI edit. All temporary
runtime/repository/engine probes are byte-restored; `git diff --check` passes. Fixture preparation
is test-only, retains every current row/artifact and does not mint or alter default player data.
The ledger placement is reconciled into the main RP sequence, not the unrelated feature sweep.
READY FOR CLAUDE DESIGNATED REVIEW of this bounded supplement. RP-205 still requires a
separate accepted-contract live/shared-replay correction; RP-201/D-020/public wire/copy/mint
remain distinct. No full AC9/A5/public journey, archive, push, deployment or 1.0 completion.

**RP-206 exact corrective handoff:** `b57b95df^..bb5c6ac1`, READY FOR CLAUDE DESIGNATED
REVIEW. Covers the committed predeclaration, actual current-curriculum test-only Exit/quit
population and RP-205's separate executed runtime finding, plus docs/tracking reconciliation.
No production/kernel/content/copy/wire/CI changes. This checkpoint is not a verdict or
archive gate; RP-205 correction is the next separately predeclared accepted-contract task.

## 2026-10-05 — RP-205 due-cross-gate correction predeclaration

Baseline: `593f4b30`, clean tree. Accepted AR1.8/AC9 and the archived T0–T1
first-failure replacement contract authorize correcting the confirmed runtime refusal;
they do not authorize a new ending, gate rule, public API or content mint.

Population: retain the actual current-curriculum no-session and both-toy active/claimed
Exit controls. Add a genuinely eligible `gate.t2_to_t3` action with matched before-threshold
and due attendance. Before threshold it must cross normally in run 1; when due it must
persist scripted_first/run 2 without applying the requested gate or spending its cost.
This later-tier diagnostic is explicitly not a substitute for the default first-hour journey.
The stored original cross_gate payload, identical retry, both histories and terminal
events must remain exact. Both toys must still reject this action while active AND claimed,
then allow the replaced action after actual quit/resolution.

Add a Go-produced, TS-consumed literal terminal fixture with the current complete curriculum,
exact receipts, Founder carry, final/new Company state and all three event batches. Preserve
the old no-curriculum cross-gate and existing manual/explicit-Exit fixtures unchanged. Red-first
Go/live tests precede the fix; TS must fail on the new fixture before its correction. Negative
trigger/branch controls must still fail and active-session evidence must still reject.

Runtime scope: `production` live branch freezing and shared Go/TS Exit replay only. A frozen
curriculum branch distinguishes replacement from historical no-curriculum gate execution;
never rewrite the canonical player's intent to Wind Down or fabricate a crossed gate.
Inspect terminal foundation hooks before changing execution. Bump all kernel identities
honestly to 0.3.148 if these watched semantics change. No balance, owner copy, schema, public
wire, migration, CI, history rewrite or RP-131 guard exception.

Discrimination: independently restore the live CrossGate refusal, the replay's old gate
execution path, and the TS CrossGate exclusion; each new witness must fail. Repeat actual
active/claimed blocking with the new cross_gate population. Restore every probe exactly.
Cold Go, client/type/build/boundaries/vet/vectors/corpus/replay fixture, declared Postgres and
Linux browser checks are required; record RP-131 separately if the complete guard stays red.
Update canonical docs and tracking in-range. No whole AC9/A5 promotion, self-approval or
archival: exact corrective range goes to Claude for designated cross-party review.

**Executed red-first:** the new Go terminal fixture fails with scripted curriculum trigger;
actual Postgres passes the genuinely eligible before-threshold gate and all eight old action
cases, but fails the due no-session gate and all three active toy/phase gate cases with the
confirmed engine error. Corrected Go/live execution passes all thirteen. Go generation adds
only curriculum_cross_gate_exit; every old top-level fixture remains byte-identical. The new
TS case then fails at the old selected-branch CrossGate refusal while all 7,169 old cases pass.

Setup errors are not product findings: the new test initially named an internal balance field
instead of the Ledger, and an overly broad patch matched ordinary ApplyLogged rather than
ApplyLoggedExit. Both failed compilation, were corrected before semantic evidence, and neither
is retained. Make consumes an unescaped final dollar in GO_TEST_FLAGS; use the named selector.
The TS restore helper required explicit nonnullable foundation catalogs; corrected typing,
without weakening any replay assertion.

**RP-207 instrument finding / bounded test-only repair predeclaration:** an unchanged old
Mine Grid/Wind Down case and then the new gate case intermittently fail Company replay.
Retained first-divergence diagnostics identify an extra persisted Founder automatic Fiscal
harvest event, not a receipt/state difference. ApplyFounderLogged decorates the Exit with
that prefix; applyFounderExitLive explicitly validates it separately from decision.FounderEvents.
The historical persistence helper aggregates both streams by intent and falsely attributes
that separately owned prefix to the Company Exit. Founder history already verifies the full
actual prefix. Repair only this test's event scoping; require deterministic automatic-harvest
coverage, exact complete Founder verification and an independently failing no-filter control.
Do not suppress unknown events, alter runtime/replay policy or delete the Company event gate.
The attempted repeat flag was overridden by SAVE_TEST_COUNT=1; it was one run, not twenty.
Use the existing SAVE_TEST_COUNT selector for the actual repeated cold population.

**RP-205 retained correction:** live branch preview accepts the original CrossGate and
freezes the existing branch selection; Go/TS shared Exit replay treats that branch as the
replacement discriminator. It does not execute the requested gate or its debit/events.
The historical no-curriculum CrossGate path and every old shared-fixture field remain
unchanged. The new literal fixture uses the actual current complete curriculum and compares
receipt, Founder carry, final/new Company and every Founder/ended/started event batch in TS.
Matched forged first-gate, prior-Exit and branch evidence refuses; active-session evidence
rejects without state mutation. Original player intent is never relabelled Wind Down.
All three kernel identities advance honestly to 0.3.148.

**RP-207 repair:** the test reader retains all actual Company events and terminal base
Founder Exit events; only the separately owned automatic Founder Fiscal prefix is excluded
from Company replay. Full actual Founder-history verification remains mandatory. The due
no-session gate now forces exactly one actual Fiscal harvest; removing only that prefix
from a cloned Founder history fires state_divergence. First-divergence diagnostics remain.
All thirteen subcases pass twenty actual cold DB repetitions (15.546 s).

**Executed independent severings:**

- Restoring only live CrossGate refusal fails due no-session and all three active gate cases;
  the genuine before-threshold gate and all eight old action cases remain green.
- Restoring only Go's gate execution in terminal replay fails with Tier 3, crossed gate and
  cash 9e9 instead of unchanged Tier 2/un-crossed gate/cash 1e10. The assertion discriminates
  the actual requested action effects, not merely a renamed error.
- Restoring only TS's CrossGate branch exclusion fails the new case while all 7,169 old
  client cases pass. Independently restoring TS's gate execution fails exact receipt bytes
  (including the missing branch starter), again with the old population green.
- Disabling the broad Exit guard fails all nine real session/action cases, including the
  three new gate cases; all four genuine no-session/before-threshold controls pass.
- Hiding only claimed status in the actual repository fails all nine specifically after
  active rejection and real quit passed. No-session/before-threshold controls stay green.
- Removing only the test reader's Fiscal scope fails its deterministic due-gate Company
  replay, while the other twelve cases pass. No runtime or verifier is changed by this probe.

Every runtime/repository/test-reader probe is byte-restored against saved fixed-source Git
object hashes before final verification; the minigame repository has no residual diff.
The extra retained removed-prefix Founder control was added after those hash comparisons.
No content, owner text, public wire, migration or CI workflow changed.

**Final cold checks executed so far:** six Go packages pass with `-count=1`, including actual
production completion at 63.082 s and regeneration of the shared fixture. Root typecheck
has zero errors/warnings; all 7,170 client cases pass / 88 intentional browser-only skips.
Build, boundaries/negative fixtures, vet, unchanged 6,296 vectors and all three Arcade corpus
regeneration checks pass. Declared real Postgres executes SEVEN functions verbosely, including
all thirteen actual Exit subcases, atomic Pitch, Typer composed and current automatic curriculum;
all pass with no skips. Complete local ARM64 Linux three-browser verification passes 21,759
cases / three intentional performance skips across 267 populations; separate Chromium
performance passes (one / 20 filtered). This is local Linux evidence, not hosted CI.
The complete kernel guard is still running; do not infer its terminal result from silence.

**Terminal guard result:** Make exits 2 at the unchanged pushed RP-131 hash
`50a3a5141d5c4c1cd25cbf3c9c39d9e3cf8e2444`, parent
`0cf9f7a6aba4038fadcdf35e5b94f56986164af7`, after checkout-contract and adversarial
fixture controls pass. No bypass, false bump, history rewrite or hosted-green claim.
RP-205/RP-207 are locally corrected, READY FOR CLAUDE DESIGNATED REVIEW; the exact
predeclaration-through-implementation range will be pinned by the following checkpoint.
The original full A5/AC9, default public journey, public wire/copy/mint, D-020 and full 1.0
remain separate. No self-approval, archival, push or deployment.

**Exact RP-205/RP-207 corrective handoff:** `efdc2dbc^..4b8abb89`, READY FOR CLAUDE
DESIGNATED REVIEW. This span covers the committed accepted-contract predeclaration,
watched live/Go/TS correction and honest kernel 0.3.148, literal shared replay fixture,
actual thirteen-case gate/active/claimed/quit/retry/history proof, separately recorded
Fiscal event-reader repair, and canonical docs/tracking. The final retained population,
including the added removed-prefix Founder negative, passes a fresh twenty cold repetitions
(13.147 s), not merely the earlier twenty before that extra control. All probes restored;
RP-131 remains red at its unchanged pushed hash. This is a handoff, not an approval/archive.
Next accepted work is the original A6/A7 test-only child/copy/docs review, with public
host/wire, D-020, owner copy and mint still blocked under their named gates.

## 2026-10-05 — A6 delayed-acknowledgement review predeclaration

Baseline `5b876886`, clean tree; original Claude A6/A7 range `fca062a1^..fca062a1`.
Accepted AR6.3 explicitly requires one command in flight and an immediate terminal flush.
Current Snake quit awaits flush(), but flush returns immediately when an advance is already
outstanding. A local terminal step also returns from that guard; after the outstanding
acknowledgement, nothing visibly drains that terminal suffix. These are candidate runtime
defects, not yet executed findings. Existing RP-188/RP-189 callback/keyboard proofs are retained.

Population: mounted actual SnakeBoard plus real shared TS engine, fake delivered presentation
callbacks (never wall-time-as-progress), and a host that delays only the first acknowledgement.
Case one reaches a genuine first nonterminal advance, issues Quit via its DOM button while
that response is pending, and must never overlap requests; acknowledge and require actual
terminal Quit with exact ordered commands. Case two reaches a real wall crash during that
pending response; acknowledge and require the complete terminal advance and real server
terminal state without another player action. Immediate-ack controls and existing freeze,
rejection/resync, native-key quit and actual-engine controls must remain green.

Run the unchanged component red-first. If confirmed, ledger each defect immediately and
correct only owned request sequencing/terminal drain under AR6.3, not host/public API, timing,
engine math, copy, surface registration, balance/mint or CI. No kernel bump for unguarded
presentation code unless the actual correction crosses a watched semantic boundary.
Include destroy/rejected-response controls and independently sever serialization and terminal
draining. Native available browsers are partial evidence; final complete Linux target must
execute all three engines. Update docs/tracking and require exact Claude corrective review;
the original A6/A7 remains unapproved until the designated verdict covers its whole scope.

**Executed RP-208/RP-209 red baseline:** Chromium and WebKit each fail exactly the two new
delayed-response cases while their six old child cases pass (4 failed / 12 passed total).
DOM Quit submits two requests before the first acknowledgement, not one. The local terminal
case submits only advance-through-1; its required automatic advance-through-3 is absent after
acknowledgement. Both hosts execute delivered commands through the actual shared TS engine.
The delayed host does not fabricate a terminal or relax engine validation. Native Firefox
is not in this population; complete Linux verification is still required after correction.
The failed browser target does not run its separate performance tail, so none is claimed here.

**Bounded original A6 designated review — CHANGES REQUIRED.** Review by: Codex.
Recorded by: Codex. Original range `fca062a1^..fca062a1`, targeted Snake AR6.3 request/terminal
behavior: RP-208 and RP-209 violate its explicit one-in-flight and immediate-terminal-flush
clauses. This finding is not an approval of every A6/A7 path, accessibility or candidate-copy
claim. Corrective implementation stays within the above committed predeclaration and requires
Claude's exact cross-party review; no self-approval or archival.

**Retained correction:** flush now returns the actual shared outstanding Promise; Quit pauses,
joins/drains the bounded local lead, then submits once. Repeated Quit/steering/resume cannot
start another sequence while it is pending. An acknowledged earlier advance drains a buffered
local ending automatically. Failed recovery returns failure instead of retrying blindly;
unmount prevents new recovery/drain requests and ignores late responses. No shared-engine,
pace, copy, public wire or content change. This child is outside kernel/affecting-paths.json;
kernel remains 0.3.148, without a false version signal.

Six new child/real-engine cases retain exact request order, concurrency peak, terminal state,
tick and revision. They cover pending Quit, pending terminal suffix, repeated partial-lead Quit,
failed advance plus failed recovery, and accepted/rejected delayed responses after unmount.
Controlled callback rejection is explicitly an instrumented host failure, not a claimed real
public API refusal. All normal delivered commands still execute through the actual TS engine.
All existing six cases and their RP-188/RP-189 proofs remain intact.

The first corrected run exposed a test timing assumption: one Svelte tick did not certify the
extra joined-promise continuation. The test now waits on its actual expected submission,
under the framework's bounded wait, while the game is paused/terminal. No wall-time progress
claim or fabricated command replaces an assertion. Duplicate unmount produced visible Svelte
warnings in the added cleanup controls; they now track and unmount the instance exactly once.

**Independent restored probes (Chromium + WebKit):** bypassing Quit's join/drain loop fails
pending Quit, repeated partial-lead Quit and failed-recovery cases (6 failures / 18 passes).
Disabling only automatic terminal drain fails precisely the pending-ending case in both
engines (2 / 22). Disabling only the pre-recovery destroy guard fails the rejected late-response
case with one forbidden current call instead of zero (2 / 22). The final component's Git
object hash matches its saved fixed baseline exactly before complete verification.

The complete kernel guard exits Make 2 at unchanged RP-131 hash
`50a3a5141d5c4c1cd25cbf3c9c39d9e3cf8e2444`, parent
`0cf9f7a6aba4038fadcdf35e5b94f56986164af7`, after checkout and adversarial fixture controls
pass. No bypass, pushed-history rewrite, false kernel bump or complete green-CI claim.

**Final cold RP-208/RP-209 verification:** six Go packages pass with `-count=1`, with actual
production completion at 45.246 s. Root typecheck has zero errors/warnings; all 7,170 client
cases pass / 94 intentional browser-only skips. Build, boundaries/negative fixtures, copy/
content-manifest checks, vet, unchanged 6,296 vectors and all three Arcade corpus checks pass.
The declared real-Postgres selector executes SEVEN named functions, including the thirteen
actual Exit cases, atomic Pitch, Typer composed and current automatic curriculum; all pass,
none skips. Complete local ARM64 Linux Chromium/Firefox/WebKit target passes 21,777 cases /
three intentional performance skips across 267 populations, with separate Chromium performance
one pass / 20 filtered. Native Firefox was not substituted for the actual Linux execution.
All delayed-response cases are executed in the full population, not counted from skipped
Node declarations. Cleanup warnings are gone; no source probe remains.

READY FOR CLAUDE DESIGNATED REVIEW of the bounded correction, with exact range pinned next.
The original A6/A7 still needs its remaining claim review; public host/wire/copy/mint and full
1.0 remain separate. No self-approval, archive, push or deployment.
