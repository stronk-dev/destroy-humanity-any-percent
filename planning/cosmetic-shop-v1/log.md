# Cosmetic Shop v1 log

## 2026-09-25 — Predeclaration (Claude)

**Implemented by:** Claude, in the kernel-path lane. It is the only lane touching kernel-guarded
paths during this work. Every guarded commit bumps `kernel/VERSION` + `server/kernel/version.go` +
`client/src/kernel/version.ts` in the same commit to HEAD + 1.

Base: `3a5234d4`, kernel 0.3.121, latest migration 00079.

Batches C1–C8 are listed in `plan.md`. The known numbering deviation is recorded there
(Founder v24). `server/minigame/session.go` (left unformatted by the Typer lane) is gofmt'd inside
the first kernel-bumping commit, never in a standalone commit.
