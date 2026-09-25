# Demo Disc Arcade — implementation plan

RFC: `rfc/minigame-demo-disc-arcade.md` (accepted 2026-09-25, recommended defaults). Fixture-first;
no production mint (AR8). Implemented by Claude; every range awaits Codex designated review.

- [ ] A1 — `arcade` artifact grammar (AR1.2/AR3.1/AR4.1): Go + TS loaders, one shared fixture,
  loader-bound rejections, candidate copy keys (AR6.6) in `copy/catalog/arcade-candidate.json`.
- [ ] A2 — `mine_grid` 1.0.0 pure engine (AR3) in Go and TS; Go-generated corpus TS replays
  byte-for-byte (AR7); hidden-information invariant (AC4) with a failing case.
- [ ] A3 — `snake` 1.0.0 pure engine (AR4) in Go and TS; corpus incl. mid-window terminal,
  `advance_past_terminal` without mutation, every rejection (AR7).
- [ ] A4 — AR-P1/P2/P3: `CatalogBundle.Arcade`, loader chain (`arcade` requires `minigame_api`),
  toy/definition/tenant cross-checks, resolver arms `(mine_grid|snake, 1.0.0) → arcade`, gameserver
  tenant registration, TS replay loader chain; fixture `minigames` + `minigame_api` candidates with
  the two AR2 rows and tenant rows.
- [ ] A5 — composed platform witness (Postgres): `always` unlock, `human_hobby` lock at near-zero
  Soul, start → play → terminal → zero-credit applied resolution, `quit` releases the session.
- [ ] A6 — client children `MineGridBoard` / `SnakeBoard` under `client/src/game-ui/minigame/`
  with the AR6.2/AR6.3/AR6.4 contract, test-only mounted (no pinned tenant row yet), axe in three
  browsers.
- [ ] A7 — docs (`docs/minigame-demo-disc-arcade.md`, platform pointers).

**Blocked (owner ruling pending, not built):** AR-P4 public API arms and AR-P5 tenant-registry
registration. Adding arms to the existing v1 minigame unions conflicts with API Foundation C2 (the
Typer lane's finding in `planning/minigame-terminal-typer/log.md`); the owner has not yet chosen
between new game-specific v1 operations (recommended), `/v2`, or a re-pin. Also not built here:
the arcade Game UI surface host (AR6.1/AR6.5), which needs the public wire; the OD-3 Fiscal
retirement and all production artifact bytes (AR8 mint).
