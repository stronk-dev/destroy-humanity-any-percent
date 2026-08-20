# Owner ruling packet — release and program choices

Coordinate: product tree `190a4fa`; evidence through `71d2594`; 2026-08-20.

This packet turns the audit into choices an implementation agent may not make. It does not adopt
any option. “Recommended” means the smallest internally coherent route supported by current
evidence, not a ruling.

## O-001 — next public milestone and exact floor (D-001/D-007)

Current evidence directly rules out calling the product 1.0: only T0–T1, one real minigame, one
early Exit, and a Desk-reaching browser path exist. Current-head CI does not reach a verdict and
the public-service release floor is incomplete.

Choose exactly one next label:

| Option | Meaning | Minimum honest exit |
|---|---|---|
| **A. Repository development snapshot** | Source is inspectable; no claim of a supported playable service or release. | Current limitations remain prominent; no “self-hostable,” “alpha,” or “release-ready” language. CI still needs repair before a quality claim. |
| **B. Phase-0 playable preview — recommended** | A bounded T0–T1 player release, explicitly not roadmap v0.1 or 1.0. | Current-head CI verdict; complete browser first run/Exit/run-2; durable sessions and honest outage behavior; recovery/export/delete; task accessibility; supported package/config/backup/restore/rollback; privacy/legal/operations floor; exact T0–T1 content statement. |
| **C. Roadmap v0.1 “The Garage”** | The first public milestone already defined by `design/07`. | Everything in B plus complete T0–T2, Fiscal/Clout/shop, three named minigames, pet adoption/care, presence/feed/counters, era UI and launch copy slice. |
| **D. v1.0 “Transcendence”** | The designed game reaches its ending. | Everything in C plus roadmap phases 2–5, tiers 3–8, all endings/variants, complete category/challenge/retrospective/honesty surfaces, and a coherent preservation posture. |

A hosted public service triggers the rights, safety, accessibility, recovery, deployment, and
operations work regardless of whether its label says “preview.” Renaming the release cannot defer
behavior players and operators actually depend on.

Required ruling text:

```text
Next milestone label: [A/B/C/D + exact public name]
Hosted public service: [yes/no]
Content in: [exact tiers, endings, minigames, social surfaces]
Content deferred: [exact families]
Release-floor obligations in: [rights, accessibility, deployment, operations, preservation]
Release-floor obligations explicitly deferred: [exact items and why the chosen label permits it]
Canonical update owner: [name]
```

## O-002 — repository/shared-memory disposition (D-002)

Verified facts: GitHub is public; the research matrix says the repository is private; the required
design backlog and most research/coverage artifacts are ignored and absent from a fresh clone.

| Option | Consequence |
|---|---|
| **A. Public repository and public planning/research — recommended absent sensitive material** | Track the control plane and publish/sanitize only content the owner accepts as public; both agents receive one repository memory. |
| **B. Public code plus named durable private store** | Track a public index/stub for every private artifact and define the exact connector/location, access, freshness, and failure behavior in `AGENTS.md`. Local ignored files alone do not qualify. |
| **C. Private repository** | Change remote disposition and reconcile publication/licensing/deployment assumptions. This is an external owner action, not an agent edit. |

The current half-state is not an option: “repository is shared memory” is false when mandatory
memory exists only in one checkout.

## O-003 — account posture (D-005/D-008/D-009)

These choices must precede the R-003 participant acceptance study; the previous queue was circular
because it asked participants to use absent, undecided workflows.

1. Recovery credential: display/download/print at bootstrap, optional email, device-only risk, or a
   specified combination. The smallest recoverable posture is explicit one-time display plus
   download/copy and a rehearsed recover-existing-account branch; email can remain absent if said.
2. Export: current account + current canonical Founder/Company state; current plus history/events;
   or a signed portable self-host import bundle. State exactly whether rankings/evidence are
   included and whether imports retain provenance.
3. Deletion: enumerate deleted, anonymized, and retained data families plus retention duration or
   event trigger. “Account deleted” is not adequate disclosure while immutable ranked/replay
   history remains.

After the owner adopts these semantics, an accepted Account/UI RFC may build a prototype and R-003
tests it with nontechnical users. Participant success validates the chosen posture; it cannot choose
the posture.

## O-004 — deployment, operations, and retention (D-006/D-011/D-015)

Name the supported topology and responsible operator before drafting implementation:

```text
Supported host/topology:
Public origin and trusted-proxy boundary:
Single-node/multi-node floor:
Backup cadence and retention:
RPO / RTO:
Rollback support window:
JWT/bootstrap/cursor rotation owner and overlap:
Metrics/log destination and identifier/privacy budget:
Blocking alerts and recipient/escalation:
Incident and breach owner:
Inactive-account trigger:
Intent receipt, event/log, dead-letter, archive and moderation-history retention:
```

R-006 then rehearses the package/backup/restore result. R-007 rehearses fired alerts, cleanup,
privacy, disclosure, and failure handling. Neither study is allowed to invent the values above.

## O-005 — sunset/source posture (D-003)

Do not rule this from the current sunset dossier without its verification/current-state repair.
The dossier's premise that a supported compose bundle already exists is false at this coordinate.
After refresh, choose source at launch, binary/self-host bundle, both, or no covenant yet; separately
choose export, notice, final-world artifact, mirror, and ratchet obligations. Any adopted promise
must depend on a rehearsed deliverable, not architecture potential.

## O-006 — unresolved product/acceptance choices

| Decision | Options requiring owner/ruling-author selection | Current recommendation/constraint |
|---|---|---|
| **D-012 Advisor Mode** | Required in chosen milestone with exact toggle/behavior, or explicitly deferred and removed from current acceptance claims. | Defer from a Phase-0 preview unless its accessibility value is part of D-001's floor; do not ship orphaned state/math/labels. |
| **D-013 Minigame async scope** | Implement a composed async lifecycle now, or reconcile it out of the current platform criterion into a successor. | Defer until a real minigame needs it; enums are not evidence. |
| **D-014 CI topology/latency** | Preserve an exact blocking budget and move slow evidence elsewhere, or adopt a measured longer blocking budget/topology. | Wait for authority-preserving R-001 measurement; never ratify the current 30-minute kill by inertia. |
| **D-016 gameplay telemetry** | None, privacy-preserving aggregate only, or a larger disclosed gameplay set. | Current preview can use none; community milestones cannot ship until a ruled aggregate instrument is measured and disclosed. Operational metrics are separate. |
| **D-017 UGC/social scope** | No UGC, curated/guild-scoped phrases/names, or broader public feed. | Keep absent for a Phase-0 preview; any public UGC requires the notice/reason/moderation/account-rights contract before the surface. |

## Decision readiness

| Decision | State | What is still needed |
|---|---|---|
| D-001/D-007 milestone/content | **READY FOR OWNER** | Choose O-001; evidence baseline is complete. |
| D-002 repository disposition | **READY FOR OWNER** | Choose O-002; inspect any sensitive files before public tracking. |
| D-005/D-008 account recovery/export | **READY FOR OWNER POSTURE** | Choose semantics, then prototype and run R-003. |
| D-009/D-015 deletion/retention | **NEEDS LEGAL/OWNER RULING** | Complete data-family review and adopt exact retention/disclosure. |
| D-003 sunset | **BLOCKED ON RESEARCH REPAIR** | Reverify official sources and remove false current-state claims. |
| D-006/D-011 topology/operations | **NEEDS OWNER INPUT** | Host/operator constraints and exact objectives. |
| D-010 lifecycle locations | **READY FOR OWNER/PROCESS AUTHOR** | Choose one archive home for withdrawn RFCs and completed non-RFC threads. |
| D-012/D-013 product scope | **READY FOR OWNER/RULING AUTHOR** | Include or explicitly defer; reconcile the bodies. |
| D-014 CI contract | **BLOCKED ON ACCEPTED R-001 INSTRUMENT + MEASUREMENT** | Do not choose from cancelled runs. |
| D-016/D-017 telemetry/UGC | **BLOCKED UNTIL CHOSEN MILESTONE NEEDS THEM** | Rule before milestone/social RFCs, not during implementation. |
