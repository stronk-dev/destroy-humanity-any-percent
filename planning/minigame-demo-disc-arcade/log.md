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
