# Platform alignment program plan

The 2026-08-20 checkpoint at `cb162a3` created the control-plane structure and a principal-risk
scan. It did **not** complete the exhaustive repository audit. Product evidence remains pinned to
`190a4fa`; planning-only commits after that coordinate do not upgrade any capability verdict.

## Evidence contract

Every audit claim must record:

- the exact intent/design source and bounded user outcome;
- the backend producer or primitive, if any;
- the frontend or external consumer, if any;
- the real catalog/content/data row and default workflow that exercise it;
- an executable witness run cold with a fixture capable of falsifying the claim;
- failure/refusal/offline/accessibility behavior where relevant;
- the owning backlog, research, decision, RFC, or closeout route.

Reading code can establish mechanical presence. It cannot establish working integration. A test
name or prior verdict is a lead until the test is inspected for discrimination and executed at the
current product coordinate.

## Waves and gates

| Wave | State | Population and work | Required artifact | Exit gate |
|---|---|---|---|---|
| 0. Authority and coordinate | complete | Process law, design authority, active RFC index, current state, remote coordinate, dirty-work preservation. | `PROGRAM.md`, initial `log.md` | Authority chain and immutable product coordinate named. |
| 1. Exhaustive inventory | complete at audited coordinate | Every design capability/promise; every active and archived RFC acceptance claim; every live/archive planning record; canonical docs; server/client packages; balance/copy catalogs; migrations; HTTP/WebSocket routes; UI surfaces; build/deploy/CI entry points. | `inventory.md` plus machine-readable row ledgers | All declared populations reconcile; later semantic waves attach verdicts without shrinking the frozen denominators. |
| 2. Capability trace | complete at audited coordinate | All 121 parents map to 433 sequential children; each has intent, producer, consumer, current data/workflow, witness, failure, verdict, limit, and authority route. Evidence attachment split one invalid 432-row compound before closeout. | `capability-outcome-ledger.tsv`, `capability-outcome-audit.md`, expanded `reality-audit.md` and `capability-map.md` | Validator rejects dropped/duplicate/vacuously promoted/missing-route rows; manual promoted/partial and mandatory-family pass complete. |
| 3. Acceptance and test audit | complete at audited coordinate | Re-walk every active RFC AC and the predeclared risk-ranked archived sample. Inspect fixtures/oracles; run cold gates; add mutation/negative probes only under accepted RFC authority. Record caches, exclusions, guard exhaustion, truncation, and hosted/local differences. | `acceptance-audit.md`, `test-evidence-audit.md`, lifecycle dossiers, archive replay, 802-row oracle ledger | Every current completion claim has a discriminating executable witness or is downgraded/filed. Current-head gates reach valid objectives or fail loudly. |
| 4. Runtime/release audit | complete at the audited coordinate | Browser workflows, account rights/recovery, offline/outage behavior, accessibility, provider-off operation, packaging/licensing, secrets, deploy/rollback, backup/restore, observability, data retention, self-host/sunset. | Expanded `release-platform-audit.md`; four bounded release dossiers | Each release claim has an integrated workflow proof or a named research/decision/RFC blocker. |
| 5. Research and owner rulings | prepared; owner/ruling-author actions pending | Convert genuine unknowns to predeclared studies; collect choices evidence cannot make; reconcile ruled design bodies before RFC work. | `research-queue.md`, `decision-queue.md`, `owner-ruling-packet.md`, adopted design updates | No RFC is being asked to answer an empirical or owner question. |
| 6. Dependency and resource graph | complete at the audited coordinate; author repairs pending | Bind producers, consumers, shared schemas, content, migrations, versioning, refusal paths, accessibility, and acceptance witnesses. | `rfc-graph.md`, `dependency-resource-ledger.tsv` | DAG has no consumer without producer/content/proof ownership and no duplicate shared-resource owner. |
| 7. Executable program | complete at the audited coordinate | Rank only accepted, dependency-satisfied work; define transactional closeout and review ranges. | `execution-queue.md`, `ready-batch-manifest.tsv`, reconciled backlog/registers | READY means implementable now; every other row names the exact blocker. |
| 8. Contradiction and independent review | Codex contradiction pass complete; designated Claude review pending | Cross-check all ledgers/indexes/docs against the product coordinate; obtain designated cross-party adversarial review over the complete research/planning range. | Finalized `review-handoff.md`; append-only verdict in `log.md` | No uncovered commit range, contradictory normative text, stale current-status claim, or unowned finding. |

## Wave 1 inventory slices

| Slice | State | Reconciliation target |
|---|---|---|
| Design promises and release obligations | 121 parents -> 433 evidence-traced children complete | All parents reconcile; 112 split and only nine remain atomic. Mandatory controls cover T2, Founder currencies, minigame contracts, guilds, security, Run End, and content families without granting inherited proof. Only three integration rows survive, all views of one bootstrap path. |
| RFC lifecycle and acceptance | active population plus archived structural/deep replay complete | Every active/archive file appears exactly once; 111 active-directory rows are reconciled; all 46 archive pairs score exactly and the 20-row/ten-domain deep sample routes every fired criterion. |
| Server producers and persistence | package/operation/migration/actor/event/worker boundary complete; semantic rollback pending | All 45 package directories, five commands, 24 operations, 74 migrations, 11 explicit deployed goroutine instances, six jobs, 48 player event kinds, and protocol families have bounded rows; isolated semantic rollback remains. |
| Client consumers and workflows | complete at audited coordinate | All 82 sources and 25 default/failure/accessibility workflows trace entry/build reachability, HTTP/WebSocket use, storage, surfaces, real data, witnesses, exact gaps, and row-level oracle limits. |
| Declarative content and copy | complete at audited coordinate | All 91 balance and ten copy files, 23 live/platform families, 208 generated keys, and 579 deploy-current gameplay units reconcile. Row result: 0 proven mounted, 173 partial-mounted, 180 backend-active, 141 dormant, 55 measurement-only, 21 zero/empty, nine contradicted. |
| Executable evidence | complete at audited coordinate | All 72 phony targets, seven CI jobs, 56 local client artifacts, 151 server test files, 592 Go functions, 210 client declarations/helpers, and 40 skip sites reconcile. The 802 row verdicts are zero integrated, 171 bounded, 533 positive-only, 43 fixture/mock, 51 dependency-conditional, one non-discriminating, one invalid/guarded, and two helpers. |
| Release and operations | bounded coordinate pass complete; implementation/rehearsal blocked on owner and RFC gates | Build/package artifacts, dependency rights, provider requirements, secrets, deploy, backup/restore, rollback, monitoring, export/deletion, preservation. |
| Planning/docs/shared memory | complete at audited coordinate; author repairs remain externally routed | All 23 planning threads, 38 docs, 238 tracked planning files, and 25 ignored planning files are classified after final contradiction closeout; owner/ruling-author body fixes and archival moves remain routed, not inferred. |

## Required negative controls

- At least one backend-only primitive must remain non-proven without a user consumer.
- At least one component-only UI must remain non-proven without a real producer/content path.
- At least one intentionally absent capability must be detected as absent.
- Each audited gate must demonstrate that a relevant mutation or failing fixture makes it fail.
- Any timeout, guard exhaustion, exclusion, truncation, or skipped population invalidates the
  measurement rather than shrinking its denominator.
- Offline/provider-off/recovery claims must be tested with the dependency actually unavailable.
- Accessibility claims must include task-level keyboard/assistive workflows, not axe alone.

## Commit and review protocol

- Commit inventory/research waves separately from product implementation.
- Do not change product behavior during truth establishment.
- Append findings and command outcomes to `log.md` in the same research commit.
- Do not flip implementation-plan completion boxes from audit inference.
- No planning/research wave authorizes archival. The full range requires the designated
  cross-party review required by `AGENTS.md`.
