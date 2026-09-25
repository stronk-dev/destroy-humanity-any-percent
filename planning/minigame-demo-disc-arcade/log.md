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
