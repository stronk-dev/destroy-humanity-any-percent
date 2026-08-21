# D-002 Targeted Artifact Review

**Review date:** 2026-08-21  
**Population:** all nine Class-B artifacts in `publication-sensitivity-audit.md`  
**Effect:** disposition only. No ignored source was edited, added, deleted, adopted, or published.

## Verdict

All nine targeted artifacts now have an exact disposition. None is eligible to be blindly added
to the public repository in its current location and authority role.

| Disposition | Count | Artifacts |
|---|---:|---|
| Ruling-author reconciliation | 2 | `design/research/README.md`; `planning/coverage-map/decisions-log.md` |
| Public-safe synthesis or bounded rewrite | 2 | `design/research/cattery-reusables.md`; `design/research/cicd-deploy.md` |
| Frozen historical archive | 2 | `planning/codex-fixes-2026-07-30.md`; `planning/coverage-map/mint-content-rows-proposal.md` |
| Verified redundant; retain only the canonical production copy | 3 | the three `planning/coverage-map/draft-artifacts/*.json` files |
| **Total** | **9** | |

## Per-artifact findings

### Ruling-author reconciliation (2)

#### `design/research/README.md`

The normative repository-disposition section still says the project is one private repository,
that research must never be published, and that sanitization would destroy its value. That is the
opposite of the owner's 2026-08-21 D-002 ruling: public repository plus tracked, inspected and
sanitized public planning/research.

This is not a wording cleanup for an implementation agent. The ruling author must reconcile the
body, retain the historical sequence, define the public-source/private-source boundary, and name
the durable private store if any raw source remains excluded.

#### `planning/coverage-map/decisions-log.md`

The log correctly preserves the 2026-08-06 private-repository decision and its 2026-08-07
correction to a public repo with ignored/history-filtered research. It does not yet contain the
2026-08-21 superseding shared-memory ruling. The ruling author must append that ruling and mark the
former final disposition superseded without rewriting the historical entries.

### Public-safe synthesis or bounded rewrite (2)

#### `design/research/cattery-reusables.md`

The raw dossier is not public-eligible. It contains an absolute workstation path, a live hostname,
port numbers, service topology, route and endpoint inventory, authentication gaps, token/header
details, exact sibling filenames and line-level extraction recipes. Those details are unnecessary
to preserve the reusable conclusions and expose an adjacent system outside this repository's
publication authority.

Create a source-neutral synthesis of the reusable patterns: localhost binding, reverse-proxy
ownership, authenticated webhook shape, single-flight/debounce behavior, health checks and
separation of reusable patterns from project-specific configuration. The raw dossier requires an
explicit sibling-publication/private-store ruling or a later verified cleanup; it must not be
force-added.

#### `design/research/cicd-deploy.md`

This file combines useful public-source CI/deployment research with material that should not cross
the public boundary unchanged:

- the author's workstation architecture, core count, toolchain and machine-relative paths;
- exact sibling-repository scripts, topology, ports, headers and deployment recipes;
- time-sensitive July-2026 platform/version/pricing claims;
- benchmark estimates and CI-cost conclusions superseded by the complete R-001 instrument;
- proposed architecture and player-facing copy that are research recommendations, not shipped
  behavior.

Apply a bounded rewrite: preserve the public-source comparisons and Cloud Clicker conclusions,
remove sibling and machine identity, label dated vendor facts, and replace the old cost premise
with R-001's current measured result (974.510 seconds locally; standard pacing 739.441 seconds,
active relevance 212.927 seconds). Do not convert recommendations into accepted deployment
authority; D-006, D-011 and D-014 remain the decision/RFC route.

### Frozen historical archive (2)

#### `planning/codex-fixes-2026-07-30.md`

This is a valuable record of an earlier corrective queue, but it states that the repo is
local-only and mixes completed, superseded and once-current assignments in a shape that resembles
a live execution queue. Track it only as a frozen historical artifact with an explicit through
date and pointers to the current platform-alignment backlog, decision queue and execution queue.
Do not refresh individual rows or allow it to authorize new work. If the owner instead chooses
removal, first prove every still-material finding has a canonical successor and obtain explicit
cleanup approval.

#### `planning/coverage-map/mint-content-rows-proposal.md`

The proposal is useful provenance for the mint-content sequence and appended owner rulings, but
its header says draft, it contains duplicate JSON payloads, and it is not the current content
authority. Preserve it with the dated coverage-map historical archive, add a noncanonical banner
and point readers to the adopted production artifacts and current rulings. Do not track it as a
live proposal and do not use its duplicated rows as a content source.

### Verified redundant canonical duplicates (3)

The ignored drafts are byte-for-byte identical to already tracked canonical production artifacts:

| Ignored draft | Canonical artifact | Result |
|---|---|---|
| `planning/coverage-map/draft-artifacts/achievements-first-content.json` | `balance/achievements/first-content.json` | `cmp -s` equal |
| `planning/coverage-map/draft-artifacts/meters-first-content.json` | `balance/meters/first-content.json` | `cmp -s` equal |
| `planning/coverage-map/draft-artifacts/pets-first-content.json` | `balance/pets/first-content.json` | `cmp -s` equal |

Tracking the ignored copies would create a second authority with no preservation benefit. The
canonical production bytes are already durable. Keep the duplicates ignored until an explicit
owner-approved cleanup removes them; this review does not authorize deletion.

## Required transaction

After the two author-owned disposition texts are reconciled:

1. produce and review the two public-safe research rewrites;
2. move the two historical artifacts with an explicit frozen-through banner and current-authority
   pointers;
3. leave the three redundant drafts untracked and point historical prose at canonical production
   data;
4. change `.gitignore` only for artifacts approved in that same reviewed range;
5. run the fresh-clone authority gate before claiming shared memory is durable.

No step authorizes publication, push, deployment, player-copy adoption, product implementation, or
destructive cleanup.

## Execution status — 2026-08-21

The two ruling-author reconciliations and two public-safe derivatives are complete in
`publication-disposition-execution-01.md`. The two frozen historical moves are complete in
`publication-disposition-execution-02.md`. The raw sibling/machine-specific dossiers and three
canonical duplicates remain ignored and noncanonical; no file was force-added or deleted.
