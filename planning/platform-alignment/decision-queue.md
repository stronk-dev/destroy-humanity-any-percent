# Owner decision queue

Implementation agents may gather evidence and frame options; they may not infer these choices.

| ID | Owner choice | Evidence required first | Canonical home | Blocks |
|---|---|---|---|---|
| **D-001** | Define the exact label and floor for the next public milestone: development preview, playable alpha, or 1.0. | R-002 baseline, capability map, content scope. | `design/07-roadmap.md` and release RFC metadata. | Every “release-ready” claim and Deployment acceptance. |
| **D-002** | Reconcile repository disposition: GitHub/source is public while research/backlog/coverage records are local-only and internal prose says the repo is private. Decide which artifacts are public, which durable private store carries the rest, and how both agents receive them on a fresh clone. | Current GitHub visibility, `.gitignore`, legal/research corpus review. | Research disposition, ignore policy, and ops docs. | Trustworthy publication/privacy stance and shared-memory model. |
| **D-003** | Adopt, narrow, or reject the proposed Sunset Covenant and decide whether the self-host deliverable is source, binary bundle, or both. | Designed-sunset verification refresh; R-002. | `design/00`, `design/08`, new accepted RFC. | Export, preservation, self-host packaging. |
| **D-005** | Choose the recoverability posture for the one-time recovery code: display/download/print, optional email, device-only risk, and loss disclosure. | R-003 plus threat model. | Account UX design and copy ruling. | Account UI, recovery, deletion/export workflow. |
| **D-006** | Choose supported production/self-host topology, operator, backup retention, RPO/RTO, and secret rotation ownership. | R-006 plan plus hosting constraints. | Deployment RFC and ops docs. | Packaging, backups, restore, deploy. |
| **D-007** | Define content scope for the chosen milestone: T0–T1 vertical slice or the designed nine tiers and ending. | Capability map, pacing evidence, workload estimate. | `design/07` roadmap. | Content RFC DAG and honest release label. |
| **D-008** | Decide whether player data export includes only current state, all save revisions/events, leaderboard history, or a portable self-host import bundle. | Data inventory, privacy/legal review, R-003. | Account/data-portability design. | Export schema and UI. |
| **D-009** | Rule the account deletion disclosure: which anonymized gameplay records remain and for how long. | Current schema trace and legal review. | Account design/copy. | Player-facing deletion workflow. |
| **D-010** | Decide the lifecycle location for withdrawn RFCs and completed non-RFC maintenance threads. | Active RFC audit. | RFC-0000 amendment if needed. | A mechanically checkable active-work index. |
| **D-011** | Choose the production observability/incident floor: named operator, metrics/log destinations, privacy budget, alert/SLO set, escalation and breach ownership. | Operations dossier, R-007 plan, chosen topology. | Deployment/operations RFC and runbooks. | Monitoring, request correlation, retention failure handling, release operations. |

Already-blocked authored action: the author of GU-C25–GU-C28 must reconcile Game UI AC1's
normative body. This is not a new option for an implementer to choose.

Resolved input, not an owner question: `design/11 §1b` adopted silent server-anonymous as the
default and local-only play as the labeled outage fallback. `design/06` and Account D4 still use a
broader “may run fully offline” formulation. Their authors must reconcile the bodies and specify
the fallback contract; an implementation agent may not re-open the already chosen default.
