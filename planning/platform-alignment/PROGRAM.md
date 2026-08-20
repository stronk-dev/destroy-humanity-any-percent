# Repository alignment program

This program is the repository-wide control plane for turning design ambition into proven user
outcomes. It does not replace RFC-0000 or per-RFC planning. It exists because those records can be
individually plausible while disagreeing with each other and with `HEAD`.

## Authority chain

Every material claim must travel through this chain without skipping a gate:

```
observation / complaint
        -> BACKLOG defect or question
        -> reality audit at a named commit
        -> bounded research or owner decision
        -> reconciled design intent
        -> accepted RFC and dependency owner
        -> implementation
        -> integrated user-workflow proof
        -> code + docs + RFC + plan + ledger closeout in one reviewed range
```

Research answers what is true. Owner decisions choose between legitimate product or risk
positions. Design states intent. RFCs specify authorized implementation. Tests and runtime
evidence prove outcomes. None substitutes for another.

## Classification

| Class | Meaning |
|---|---|
| **Proven integration** | Producer, consumer, real content/workflow, and a current executable witness all exist. |
| **Mechanical fragment** | A schema, package, route, component, or test fixture exists, but the real user workflow is incomplete. |
| **Claimed only** | A current record says it exists, but the trace is missing or contradicted at `HEAD`. |
| **Absent** | No implementation for the named outcome exists. |
| **Blocked** | The next gate is named and cannot be passed by implementation judgment alone. |

An archived RFC proves only the bounded behavior it specified. It does not automatically promote a
larger product capability to proven integration.

## Program gates

1. Audit claims against the exact commit and, for hosted behavior, the exact remote run.
2. Put every discovered defect, contradiction, unknown, and owner choice in the internal
   `design/BACKLOG.md` and mirror repository/release findings in tracked `backlog.md` until D-002
   establishes a durable shared store.
3. Trace intent -> producer -> consumer -> real data/content -> executable witness.
4. Predeclare measurements and negative controls before running them.
5. Treat a fired criterion or negative result as completed evidence.
6. Do not draft an RFC that silently decides a research or owner question.
7. Mark only currently authorized, dependency-satisfied work `READY`.
8. Close code, canonical docs, RFC/index state, plan, and ledgers transactionally.
9. Preserve the cross-party review and full-range-union archival gates in `AGENTS.md`.

## Current audit coordinate

- Local and remote repository commit at this wave's start:
  `cb162a3bd8b00ea7378293c5d7179995688151ec`.
- Last product-behavior commit: `190a4fa04958cc2a3b4e689804cd55682f6c6420`; the intervening
  `cb162a3` changes planning/research only, so product capability verdicts stay pinned to `190a4fa`.
- Audit dates: 2026-08-20 through 2026-08-21.
- Product changes during audit: none.
- Existing dirty work preserved: `AGENTS.md` only.
- Hosted evidence: current-head push run `32009994004` and nightly runs `32096019304`,
  `32212696707`, and `32328790752` all ended cancelled; the harness job exhausted its 30-minute
  budget while `balance-harness -mode=check` was still running.

## Artifact index

- `plan.md` — exhaustive eight-wave audit plan and exit gates.
- `inventory.md` / `inventory.tsv` — counted repository populations and reconciliation state.
- `design-capability-ledger.tsv` — stable design-outcome IDs and preliminary trace state.
- `active-acceptance-ledger.tsv` — all 111 true active-RFC acceptance rows with current bounded
  verdicts; five remain open on exact review/provenance closeout.
- `acceptance-audit.md` — criterion-level classification, downgrades, and execution batches.
- `test-evidence-audit.md` — cold command outcomes, populations, oracle gaps, and invalid runs.
- `capability-map.md` — product outcomes and their actual stage.
- `backlog.md` — tracked interim defect/question ledger while the internal design ledger is ignored.
- `release-platform-audit.md` — release/readiness evidence at the audit coordinate.
- `deployment-foundation-lifecycle-audit.md` — artifact/configuration/operations trace and the
  draft Deployment RFC's five criterion verdicts.
- `account-rights-release-audit.md` — player credential/recovery/export/deletion workflow trace.
- `accessibility-release-audit.md` — task/viewport/focus/reduced-motion evidence and failed probes.
- `operations-retention-preservation-audit.md` — runtime supervision, observability, cleanup,
  retention, and sunset current-state trace.
- `reality-audit.md` — producer/consumer/content/proof traces.
- `active-rfc-audit.md` — lifecycle truth for every active RFC.
- `*-lifecycle-audit.md` — full criterion/producer/consumer/range traces for each reconciled active
  implementation RFC.
- `research-queue.md` — predeclared empirical and verification work.
- `r001-harness-diagnosis.md` — first-wave hosted/local harness evidence and instrument gap.
- `decision-queue.md` — choices implementation agents may not infer.
- `owner-ruling-packet.md` — exact release/program options, sequencing, recommendation, and
  readiness for the owner/ruling authors.
- `rfc-graph.md` — dependency and resource ownership.
- `dependency-resource-ledger.tsv` — machine-readable producer/consumer/transformation/refusal/
  witness ownership for the dependency graph.
- `server-package-inventory.tsv` — all 45 second-level Go package directories, their production
  consumers, runtime position, evidence class, and current integration gap.
- `route-operation-inventory.tsv` — all 24 exposed HTTP/WebSocket operations, authority, backend
  witness, production consumer, and exact player/operator gap.
- `migration-inventory.tsv` — all 74 contiguous migrations, domain ownership, structural Up/Down
  presence, current cold migration evidence, and the remaining semantic rollback limit.
- `client-source-inventory.tsv` — all 82 client source files, entry/build reachability, mounted
  consumer, evidence class, and exact gap.
- `client-workflow-inventory.tsv` — 25 bounded default/failure/accessibility workflows from backend
  producer through production client, real data, witness, verdict, and route.
- `balance-file-inventory.tsv` — exact boundary classification and consumer/gap for all 91 balance
  files.
- `catalog-family-inventory.tsv` — 19 current-epoch and four platform families through schema,
  loader, server/client consumers, and capability verdict.
- `copy-file-inventory.tsv` — all ten copy source/control/report files, shipped merge behavior,
  record counts, consumers, and evidence limits.
- `make-target-inventory.tsv` — all 72 phony targets classified as setup, mutator, bounded check,
  partial/invalid aggregate, measurement, alias, or manual tool with CI consumer and exact limit.
- `ci-job-inventory.tsv` — all seven hosted jobs, triggers, budgets, commands, current verdicts, and
  capability limits.
- `client-test-artifact-inventory.tsv` — all 43 tracked client test sources plus 13 ignored local
  captures, their subject, production relationship, and evidence limit.
- `server-test-file-inventory.tsv` — all 151 Go test files with package, top-level Test/Fuzz counts,
  dependency signal, and structural kind.
- `server-test-skip-inventory.tsv` — all 40 explicit skip sites in 28 server test files and the lane
  required to execute them.
- `runtime-concurrency-inventory.tsv` — every explicit deployed goroutine/job lane plus absent
  actor/scheduler claims, with trigger, message boundary, shutdown, witness, and exact route.
- `event-family-inventory.tsv` — exact closed event/protocol families and their producer,
  persistence, consumer, and capability limit.
- `archived-rfc-risk-plan.md` — predeclared 46-row archive population, risk score, mandatory deep
  strata, fired criteria, exit conditions, and authority limit.
- `archived-rfc-risk-inventory.tsv` — all 46 archive/planning pairs with structural risk inputs,
  selection reason, exact score, canonical docs/RP routing, and replay verdict.
- `archived-rfc-risk-audit.md` — the 20-row, ten-domain deep replay, commands, invalid run, fired
  criteria, and authority-bounded consequences.
- `copy-key-consumption-plan.md` — predeclared 208-key population, seven reference lanes, mounted-
  workflow verdicts, controls, failure rules, and authority limit.
- `planning-thread-inventory.tsv` — all 23 top-level planning directories with tracked/local file
  counts, authority, current state, and exact closeout gap.
- `docs-file-inventory.tsv` — all 38 canonical/generated docs with system owner, artifact kind,
  current truth class, and evidence/repair route.
- `execution-queue.md` — the only presently authorized queue.
- `ready-batch-manifest.tsv` — exact accepted-scope READY batches, negative controls, cold gates,
  forbidden scope, conflicts, and review protocol.
- `review-handoff.md` — bounded cross-party review draft; it is not the final audit handoff until
  the remaining Wave-2/3 semantic ledgers and contradiction pass close.
- `log.md` — append-only program history.
