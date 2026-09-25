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
