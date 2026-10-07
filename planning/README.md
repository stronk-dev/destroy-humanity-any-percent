# Planning — Working Docs & Job Log

Start with `AGENTS.md` (also `CLAUDE.md`), RFC-0000's delivery procedure, and the assigned RFC's
current plan/latest checkpoint. For whole-product priorities, consult
[`roadmap-1.0.md`](roadmap-1.0.md). Older `CURRENT-STATE.md` and audit records are historical
context; don't require every agent to re-audit the repository before a bounded fix.

The accepted RFC and its per-RFC plan authorize implementation.
[`platform-alignment/`](platform-alignment/PROGRAM.md) retains audits and optional cross-system
routing; it is not a second approval queue or a green-progress gate.

One directory per RFC under implementation: `planning/<rfc-slug>/` containing:

- **`plan.md`** — task breakdown, sequencing, acceptance gates, who/what is assigned (human, Claude, Codex).
- **`log.md`** — one concise, append-only checkpoint per coherent batch or useful handoff.
  Capture commands/results and remaining limits, not every tool invocation. Keep detailed
  evidence here; other boards link to it and change only when their material status changes.

Lifecycle (RFC-0000): created when an RFC moves to `implementing` → on completion, outcomes
distilled into `docs/`, RFC moved to `rfc/archive/`, and planning moved to
`planning/archive/<rfc-slug>/`. **Never deleted.**

Log entry format:

```
## YYYY-MM-DD — outcome
- Scope/range: ...
- Checks: command → result (including failures/skips).
- Limits/blockers: ...
- Review by: ...; Recorded by: ...; status/range: ...
- Next: ...
```

Routine debugging/fixes need no separate predeclaration commit or new research dossier.
Do not mirror each checkpoint into BACKLOG, roadmap, execution queue and every audit ledger.
Use RFC-0000's risk table to choose tests; tracking consistency is never product acceptance.
