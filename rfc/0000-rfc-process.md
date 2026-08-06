# RFC-0000: The RFC Process

- **Status:** accepted
- **Author:** Marco
- **Created:** 2026-07-27
- **Supersedes / superseded by:** —

## The four tiers of documentation

| Tier | Directory | What it is | Mutability |
|---|---|---|---|
| **Design** | `design/` | Intent and evidence: the game design docs. Where ideas come from. | Amended rarely; never the implementation spec |
| **RFC** | `rfc/` | Active implementation specs. Every system that gets built is specified by one or more RFCs before implementation. | Living until implemented, then frozen and moved to `rfc/archive/` |
| **Planning** | `planning/` | Per-RFC working documents: the implementation plan and a running log. The durable record of long-running jobs. | Living during implementation; archived on completion |
| **Docs** | `docs/` | Canonical description of what actually exists — architecture, data formats, runbooks, balancing values as-shipped. | Always current; updated as part of every implementing change |

**The flow:** design → active RFC (specification) → planning (how + log) → implementation →
canonical docs + RFC/planning archives. Once a feature ships, `docs/` is the only current
description of its behavior; archived RFCs explain why and how it arrived.

## RFC lifecycle

`draft` → `accepted` → `implementing` → `implemented` → (`superseded` | `withdrawn`)

- **draft**: under discussion; anything may change.
- **accepted**: scope and approach agreed by Marco; implementation may be planned.
- **implementing**: at least one planning doc exists and work is underway.
- **implemented**: shipped; canonical behavior has been distilled into `docs/`, and the frozen
  RFC has moved to `rfc/archive/`.
- **superseded**: a later RFC replaces it (must link both ways). **withdrawn**: abandoned, kept for the record.

## Rules

1. **Identity:** use a descriptive, stable filename such as `save-layer.md`. Existing numbered
   RFCs retain their historical identifiers, but new RFCs do not need a global sequence number.
   Follow-ups use a descriptive name and declare `Parent`/`Amends` metadata rather than
   pretending to be an unrelated top-level system.
2. **Scope discipline:** an RFC should be implementable in bounded work. If scope grows during drafting or implementation, **split**: the new scope becomes a new RFC referencing the parent. Note the split in both.
3. **Amendments:** small clarifications to a not-yet-implemented RFC are edited in place with a
   changelog line. Anything that changes implemented behavior is a follow-up RFC linked to the
   archived parent. The archived parent remains immutable; after the follow-up ships, `docs/`
   incorporates the new canonical behavior.
4. **Design linkage:** every RFC cites the `design/` sections it specifies. Deviations from design docs are called out explicitly in a "Deviations from design" section — the RFC wins once accepted, but the divergence must be visible.
5. **Agent rule:** coding agents implement **RFCs**, not design docs. If needed spec is missing, that's a `DESIGN-GAP` → propose a draft RFC, don't improvise.
6. **The index** (`rfc/README.md`) lists active work and the archive, including parent/follow-up
   relationships; keep it updated in the same commit as any status or location change.

## Planning docs & the job log

For each RFC being implemented, create `planning/<rfc-slug>/`:

- `plan.md` — the implementation plan: task breakdown, sequencing, acceptance criteria (tests/gates), assignee (human/agent).
- `log.md` — an **append-only running log**: dated entries for decisions made, problems hit, scope changes, review outcomes, handoffs between agents/sessions. This is the long-term memory for jobs that span many sessions — a new agent must be able to resume from `plan.md` + `log.md` alone.

On completion:
1. Distill outcomes into `docs/` (the canonical statement of what now exists).
2. Set the RFC status to `implemented`.
3. Move the RFC to `rfc/archive/` and its planning directory to `planning/archive/`. Never
   delete either — together they are the project's institutional memory.

## Docs conventions

- `docs/` is organized by system, not by history (`docs/architecture.md`, `docs/economy.md`, `docs/data-formats.md`, `docs/ops.md`, …).
- Every implementing PR/commit that changes behavior updates the relevant `docs/` page in the same change.
- When docs and code disagree, that is a bug in docs; fix it with the next change.
- Current code must be understandable from `docs/` without reconstructing a chain of RFCs.

## Changelog

- 2026-07-27: accepted as the initial numbered, lifecycle-tracked RFC process.
- 2026-07-27: amended by owner direction to rotate implemented RFCs into an archive, make
  `docs/` the sole canonical current description, and allow descriptive follow-up RFCs without
  consuming a global number.
- 2026-08-06: non-normative reference cleanup for publication; no spec change.
