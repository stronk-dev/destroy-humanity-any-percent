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
