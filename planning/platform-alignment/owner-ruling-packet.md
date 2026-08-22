# Owner ruling packet — release and program choices

Coordinate: product tree `190a4fa`; evidence through `7e98c40`; owner rulings through 2026-08-21.

This packet turns the audit into choices an implementation agent may not make. The option tables
preserve the evidence and alternatives that produced the ruling. Any item not resolved by the
adopted text below remains an owner question.

## Adopted owner ruling — 2026-08-21

Marco adopted the following bounded program. These choices are authority; the unresolved details
listed below remain blockers and must not be inferred by an implementer.

- **Current public claim:** repository development snapshot. The repository does not claim to be a
  supported playable service, alpha, 1.0 release, or release-ready self-host package.
- **Next release target:** **Cloud Clicker Phase-0 Playable Preview**, explicitly bounded to the
  audited T0–T1 vertical slice and explicitly not roadmap v0.1 or 1.0. No T2+ roadmap content,
  broad social surface, or unbuilt minigame is implied by the name.
- **Hosted public service:** no. Public hosting remains blocked until the recovery, accessibility,
  rights, supported-package/configuration, backup/restore/rollback, privacy, and operations floors
  are implemented and rehearsed. Those obligations are not waived by the preview label.
- **Repository memory:** public repository plus tracked public planning/research. Sensitive
  material must be inspected and sanitized before it becomes tracked; this ruling does not itself
  authorize blindly publishing ignored files or secrets.
- **Recovery credential:** one-time display with copy and download at bootstrap, plus an honest
  recover-existing-account path. Email recovery is deferred. Export scope, deletion semantics,
  and retention remain unresolved under O-003/O-004.
- **Deployment topology:** one supported Docker Compose node. Public origin/proxy boundaries,
  backup cadence, RPO/RTO, rotation ownership, observability destinations, alerts, incident
  ownership, and retention remain unresolved; therefore this is an architecture choice, not a
  release authorization.
- **Deferred from the Phase-0 preview:** Advisor Mode, composed asynchronous minigames, gameplay
  telemetry, and public UGC/social content. Their ruling authors must reconcile any active RFC body
  that still claims them; the deferral does not erase historical intent.
- **R-001 authority:** a narrow harness-observability follow-up RFC may be accepted and implemented
  for measurement only. It may add phase/row progress, registry-aware selection, declared/executed
  work counts, identity, and fail-loud incomplete-run artifacts. It may not change scenarios,
  balance, seeds, horizons, work budgets, timeouts, acceptance bounds, parallelism, sharding, or CI
  topology. D-014 was subsequently ruled on 2026-08-21 after nine consecutive hosted cancellations
  and the accepted local phase observation: strict fast blocking lane, exhaustive scheduled/manual
  evidence.
- **Canonical update owner:** Marco; Codex may record this ruling and implement the expressly
  authorized mechanical R-001 contract. Owner/ruling-author body reconciliations remain with their
  named authors.

## Adopted owner ruling — 2026-08-22

Marco approved the following minimal Phase-0 deployment/operations defaults and delegated the
Deployment RFC/body reconciliation to Codex. This ruling authorizes specification and planning;
implementation still requires the reconciled RFC to reach `accepted` status.

- **Preservation posture (D-003):** the public MIT source remains the current source deliverable.
  Target a supported self-host bundle as well, but make no covenant or self-host-support claim
  until the package passes R-006. Notice period, final-world artifact, preservation mirror and
  player-ward ratchet commitments are deferred to a later accepted covenant RFC.
- **Supported topology:** one Linux Docker Compose node. Caddy is the only public ingress; the
  gameserver and Postgres are private to the Compose network. Production accepts one required
  exact HTTPS public origin and trusts exactly one Caddy proxy hop. Local development may use its
  separately declared localhost origin.
- **Backup and recovery:** create an encrypted off-host backup every six hours and immediately
  before an upgrade. Retain six-hourly backups for seven days and one daily backup for 30 days.
  The supported objectives are RPO 6 hours and RTO 4 hours.
- **Rollback:** support the current and immediately previous release for seven days. Rollback uses
  the pre-upgrade backup and never runs a down-migration against a live database.
- **Operations owner:** Marco owns the official instance, incident response and breach response.
  A self-host operator owns their instance under the same documented checks; this does not create
  an official hosted public service.
- **Secret rotation:** the official operator owns rotation. JWT previous-key overlap is 30 minutes;
  bootstrap-receipt previous-key overlap is 31 days; public cursor previous-key overlap is 366
  days. Current and previous identities/values are deployment configuration, never repository
  data.
- **Observability:** expose metrics only on the private network; write structured logs to a
  size-rotated operator destination retained 14 days. Raw IP material, if collected for bounded
  security handling, is retained no longer than seven days.
- **Blocking alerts:** endpoint/readiness failure, missed or failed backup, Postgres failure, disk
  use above 80%, restart loop, cleanup failure and dead-letter growth must reach the named operator.
  Provider-off operation remains valid: the supported stack may use a local alert manager with an
  operator-configured receiver and must not require a cloud monitoring vendor.
- **Still open:** account inactivity, immutable-history/deletion, moderation-history and other
  product/legal retention semantics under D-009/D-015. This ruling does not invent them.

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

The 2026-08-22 ruling above fills the Deployment/operations fields needed for the package. Product
data-family retention remains open only where that ruling explicitly says so; Deployment must not
choose it by convenience.

R-006 then rehearses the package/backup/restore result. R-007 rehearses fired alerts, cleanup,
privacy, disclosure, and failure handling. Neither study is allowed to invent the values above.

## O-005 — sunset/source posture (D-003)

The tracked `design/research/designed-sunset-public.md` brief completed the required
verification/current-state repair without promoting the ignored raw dossier. The 2026-08-22 owner
ruling retains public MIT source, targets a supported bundle only after R-006, and defers covenant,
notice, final-world artifact, mirror and ratchet commitments. No current self-host or preservation
promise may be inferred from that target.

## O-006 — unresolved product/acceptance choices

| Decision | Options requiring owner/ruling-author selection | Current recommendation/constraint |
|---|---|---|
| **D-012 Advisor Mode** | Required in chosen milestone with exact toggle/behavior, or explicitly deferred and removed from current acceptance claims. | Defer from a Phase-0 preview unless its accessibility value is part of D-001's floor; do not ship orphaned state/math/labels. |
| **D-013 Minigame async scope** | Implement a composed async lifecycle now, or reconcile it out of the current platform criterion into a successor. | Defer until a real minigame needs it; enums are not evidence. |
| **D-014 CI topology/latency** | **RULED:** preserve the sub-five-minute push/PR target and move exhaustive harness/numeric evidence to bounded scheduled/manual jobs with an uploaded observation. | Current-head hosted success and one complete maintenance observation validate the implementation; they do not reopen the topology choice. |
| **D-016 gameplay telemetry** | None, privacy-preserving aggregate only, or a larger disclosed gameplay set. | Current preview can use none; community milestones cannot ship until a ruled aggregate instrument is measured and disclosed. Operational metrics are separate. |
| **D-017 UGC/social scope** | No UGC, curated/guild-scoped phrases/names, or broader public feed. | Keep absent for a Phase-0 preview; any public UGC requires the notice/reason/moderation/account-rights contract before the surface. |

## Decision readiness

| Decision | State | What is still needed |
|---|---|---|
| D-001/D-007 milestone/content | **PARTIALLY RULED** | Development snapshot now; Phase-0 Playable Preview next. Exact final content/surface manifest still belongs to its release RFC. |
| D-002 repository disposition | **RULED AND EXECUTED** | All 96 artifacts were dispositioned; the fresh-clone authority gate passed and rejected three forged states. |
| D-005/D-008 account recovery/export | **RECOVERY RULED; EXPORT OPEN** | Build the ruled one-time copy/download posture; choose export semantics before R-003. |
| D-009/D-015 deletion/retention | **PARTIALLY RULED; PRODUCT/LEGAL FAMILIES OPEN** | Log/IP/backup windows are ruled; complete account/history/moderation data-family review and adopt exact deletion disclosure. |
| D-003 sunset | **RULED** | Public MIT source now; target a supported bundle after R-006; defer covenant/notice/artifact/mirror/ratchet promises. |
| D-006/D-011 topology/operations | **RULED FOR DEPLOYMENT** | One Compose node, proxy/origin, backup objectives, rollback, key overlap, metrics/logs, alerts and operator are exact. |
| D-010 lifecycle locations | **READY FOR OWNER/PROCESS AUTHOR** | Choose one archive home for withdrawn RFCs and completed non-RFC threads. |
| D-012/D-013 product scope | **DEFERRED; BODY RECONCILIATION PENDING** | Advisor and async are out of the preview; ruling authors must reconcile the active bodies. |
| D-014 CI contract | **RULED 2026-08-21 — FAST BLOCKING / SLOW EVIDENCE SPLIT** | Nine consecutive cancellations plus the accepted local observation supplied the decision evidence. |
| D-016/D-017 telemetry/UGC | **DEFERRED FROM PREVIEW** | No gameplay telemetry or public UGC/social content in the Phase-0 preview; revisit only through a later owner ruling. |
