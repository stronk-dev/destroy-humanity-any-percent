# Planning — Working Docs & Job Log

> **Start here on a fresh session: [`CURRENT-STATE.md`](CURRENT-STATE.md)** — where the project
> actually is, what is blocking what, what is ratified, and the open debts. Read it after
> `CLAUDE.md` and before touching anything.

One directory per RFC under implementation: `planning/<rfc-slug>/` containing:

- **`plan.md`** — task breakdown, sequencing, acceptance gates, who/what is assigned (human, Claude, Codex).
- **`log.md`** — **append-only, dated** running log: decisions, problems, scope changes, review outcomes, session handoffs. Written so that a fresh agent can resume the job from `plan.md` + `log.md` alone. This is the project's long-term memory for multi-session work.

Lifecycle (RFC-0000): created when an RFC moves to `implementing` → on completion, outcomes
distilled into `docs/`, RFC moved to `rfc/archive/`, and planning moved to
`planning/archive/<rfc-slug>/`. **Never deleted.**

Log entry format:

```
## 2026-07-27 (codex, session N)
- Implemented X per plan §2; deviated on Y because Z (plan updated).
- BLOCKED: ... / DECISION NEEDED: ...
- Next: ...
```
