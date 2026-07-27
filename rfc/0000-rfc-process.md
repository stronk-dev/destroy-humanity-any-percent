# RFC-0000: The RFC Process

- **Status:** accepted
- **Author:** Marco
- **Created:** 2026-07-27
- **Supersedes / superseded by:** —

## The four tiers of documentation

| Tier | Directory | What it is | Mutability |
|---|---|---|---|
| **Design** | `design/` | Intent and evidence: the game design docs and research corpus. Where ideas come from. | Amended rarely; never the implementation spec |
| **RFC** | `rfc/` | The authoritative spec. Every system that gets built is specified by one or more RFCs before implementation. | Immutable once implemented — changes come as new RFCs that amend/supersede |
| **Planning** | `planning/` | Per-RFC working documents: the implementation plan and a running log. The durable record of long-running jobs. | Living during implementation; archived on completion |
| **Docs** | `docs/` | Canonical description of what actually exists — architecture, data formats, runbooks, balancing values as-shipped. | Always current; updated as part of every implementing change |

**The flow:** design → RFC (spec'd) → planning (how + log) → implementation → docs (canonical) + RFC marked implemented + planning archived.

## RFC lifecycle

`draft` → `accepted` → `implementing` → `implemented` → (`superseded` | `withdrawn`)

- **draft**: under discussion; anything may change.
- **accepted**: scope and approach agreed by Marco; implementation may be planned.
- **implementing**: at least one planning doc exists and work is underway.
- **implemented**: shipped; the RFC text is now frozen history; `docs/` carries the living truth.
- **superseded**: a later RFC replaces it (must link both ways). **withdrawn**: abandoned, kept for the record.

## Rules

1. **Numbering:** four digits, sequential, never reused. Filename `NNNN-short-slug.md`.
2. **Scope discipline:** an RFC should be implementable in bounded work. If scope grows during drafting or implementation, **split**: the new scope becomes a new RFC referencing the parent. Note the split in both.
3. **Amendments:** small clarifications to a not-yet-implemented RFC are edited in place with a changelog line. Anything that changes an *implemented* RFC's behavior is a new RFC.
4. **Design linkage:** every RFC cites the `design/` sections it specifies. Deviations from design docs are called out explicitly in a "Deviations from design" section — the RFC wins once accepted, but the divergence must be visible.
5. **Agent rule:** coding agents implement **RFCs**, not design docs. If needed spec is missing, that's a `DESIGN-GAP` → propose a draft RFC, don't improvise.
6. **The index** (`rfc/README.md`) lists every RFC with status; keep it updated in the same commit as any status change.

## Planning docs & the job log

For each RFC being implemented, create `planning/NNNN-slug/`:

- `plan.md` — the implementation plan: task breakdown, sequencing, acceptance criteria (tests/gates), assignee (human/agent).
- `log.md` — an **append-only running log**: dated entries for decisions made, problems hit, scope changes, review outcomes, handoffs between agents/sessions. This is the long-term memory for jobs that span many sessions — a new agent must be able to resume from `plan.md` + `log.md` alone.

On completion:
1. Distill outcomes into `docs/` (the canonical statement of what now exists).
2. Set the RFC status to `implemented`.
3. Move the planning dir to `planning/archive/NNNN-slug/`. Never delete it — the archive is the project's institutional memory.

## Docs conventions

- `docs/` is organized by system, not by history (`docs/architecture.md`, `docs/economy.md`, `docs/data-formats.md`, `docs/ops.md`, …).
- Every implementing PR/commit that changes behavior updates the relevant `docs/` page in the same change.
- When docs and code disagree, that is a bug in docs; fix it with the next change.
