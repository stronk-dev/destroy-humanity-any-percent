# Operations, retention, and preservation release audit

Coordinate: product tree `190a4fa`; audit checkpoint after `571cfe3`; 2026-08-20.

This pass traced process health, fatal-background-work propagation, logging and metrics, request
correlation, scheduled cleanup, persisted-data retention, inactive-account handling, and the
designed-sunset claims through the composed gameserver. It read production composition rather than
assuming that package-local primitives or research recommendations were deployed. No product,
test, schema, design/RFC, canonical-product-doc, player-copy, or retention byte was changed.

## Bottom line

The runtime has a sound bounded supervision core:

- `/healthz` reports process liveness and `/readyz` checks readiness state plus Postgres;
- jobs must prime before readiness, unexpected job exit lowers readiness and reaches the process
  failure channel, and SIGINT/SIGTERM runs the bounded drain sequence;
- expired refresh/access credentials and bootstrap ciphertext are pruned by one composed minute
  job, with startup prime and real-Postgres evidence; and
- the Save layer keeps the latest five ordinary revisions while preserving referenced Founder
  genesis rows, and events deliberately remain append-only replay authority.

Those are useful primitives. They do not make the service operable or its retention posture
coherent:

1. **There is no production metrics surface.** Centrifuge's namespace-label callback is enabled so
   its queue hook can classify frames, but no `/metrics` endpoint or exporter is mounted. The
   Production service is explicitly composed with `InvariantMetrics=nil`; invariant events reach
   JSON logs only. There are no latency/error/queue/GC/connection/business counters, SLOs, alerts,
   dashboards, or production runbook.
2. **There is no request/access correlation trail.** The composed private API has no access-log,
   recovery, timeout, or request-ID middleware. A request-ID primitive exists only in the
   uncomposed public-API policy. Runtime JSON logs cover process termination, save rejection, and
   three invariant families; ordinary requests, job identity, retries, readiness changes, and
   operator actions are not recorded.
3. **The documented 30-day idempotency retention is not scheduled.** `Store.PruneIntentRecords`
   exists and has a package integration test, but its only non-test occurrence is the method
   declaration. `Compose` schedules credential/bootstrap GC, not intent-record GC. Therefore the
   canonical statement that a deployment scheduler owns a 30-day policy is a future contract, not
   current behavior.
4. **There is no complete data-retention schedule.** Events are intentionally append-only and
   dead-letter outbox rows remain operational evidence, while anonymous accounts, archived
   Founders/streams, Guild records, and most projections have no inactivity expiry or capacity
   owner. The compliance research recommends 24–36-month inactive-account purge, short raw-IP
   security retention, aggregate telemetry, and a retention schedule; none is accepted/composed.
   The resettable in-memory IP limiter is not durable abuse control for account/storage creation.
5. **The preservation research materially overstates current reality.** Its option-space table says
   a single Go binary and one `docker-compose up` deliverable already exist and concludes the
   covenant is “not an engineering project.” At this coordinate every Compose file is test-only,
   the binary reads repository files, no client is packaged or served, and export/import bundle,
   bots-default self-host workflow, source/license posture, final-world artifact, mirror, notice
   process, and sunset runbook are absent. The proposal remains valuable research, but its current-
   state premise is false and cannot authorize a promise.

## Producer -> consumer -> operational proof

| Outcome | Producer / primitive | Production consumer | Current proof | Verdict |
|---|---|---|---|---|
| Process liveness/readiness | Server state, Postgres ping, job-health state | `/healthz`, `/readyz` | Cold gameserver unit suite exercises healthy/unhealthy/draining states | **Proven bounded primitive** |
| Fatal worker supervision | Prime/run contract and failure channel | `cmd/gameserver` waits on `Server.Failures()` | Unexpected clean/error exits are unit-tested | **Proven integration** |
| Graceful shutdown | intent gate, worker stop, relay flush, socket drain | signal-driven entry point | Unit lifecycle ordering and composed drain witness | **Proven integration** |
| Expired credential cleanup | Account session/bootstrap pruners | minute `sessionGCJob` in `Compose` | Package and composed Postgres fixtures | **Proven integration** |
| 30-day intent-record cleanup | `Store.PruneIntentRecords(cutoff)` | none | Package method test only; no production call site | **Orphaned primitive / claimed policy** |
| Inactive-account/storage cleanup | none | none | no policy, job, quota, or rehearsal | **Absent** |
| Runtime metrics/alerting | optional Production counter; Centrifuge internal metrics config | Production passes nil; no exporter/route/collector | source/call-site trace only | **Absent as operator capability** |
| Request correlation/access audit | public-policy request-ID primitive | public policy itself is uncomposed; private API has none | primitive unit tests only | **Absent in runtime** |
| Backup/restore/rollback | save migrations and replay restoration | no operator tool/runbook | schema tests only | **Absent** |
| Player/service preservation | sunset research and architecture principles | no accepted covenant or deliverable | no clean-host/export/sunset rehearsal | **Claimed-only research intent** |

The restored cold command was:

- `make test-go GO_PACKAGES='./gameserver ./cmd/gameserver ./account ./save' GO_TEST_FLAGS='-count=1'`
  — all four packages passed. This establishes that the observed gaps are composition/operations
  gaps rather than a red primitive baseline. A repository-wide reference trace found no production
  caller for `PruneIntentRecords` and no metrics HTTP/export binding.

## Retention inventory at this coordinate

| Data family | Actual behavior | Missing decision/proof |
|---|---|---|
| Refresh/access credentials | Expired rows deleted in bounded batches; empty families removed | Operator visibility and failure alerting |
| Bootstrap retry receipts | Ciphertext tombstoned after refresh expiry | Tombstone retention/deletion policy |
| Save revisions | Latest five ordinary snapshots retained; referenced genesis may remain | Export/restore expectations and archived-stream retention |
| Intent receipts | Indefinite in production despite documented 30-day scheduler policy | Schedule, batch/error semantics, idempotency disclosure |
| Authoritative events/run/founder logs | Append-only/immutable by design | Legal/product retention and export/deletion boundary |
| Outbox/dead letters | Published/dead-letter rows retained | Retention and operator inspection/repair workflow |
| Anonymous accounts and archived Founders | No inactivity purge | D-009 plus abuse/storage and recovery policy |
| Guild/moderation/projection history | Mixed anonymization and immutable history | Complete data map, legal basis, expiry, export/deletion disclosure |
| Application logs | stdout JSON, no configured sink/rotation | IP/identifier minimization, access, retention, incident policy |

## Smallest honest next order

1. D-001 chooses the milestone. D-006 chooses topology, operator, RPO/RTO, backup and secret
   ownership. D-009/D-015 choose retained-account/history disclosure and complete retention.
   D-003 chooses covenant posture.
2. Complete a bounded data inventory and legal/product retention ruling before adding cleanup.
   Deleting immutable replay authority by convenience would break accepted history semantics.
3. Reconcile `docs/save-layer.md` so the 30-day scheduler is explicitly future until an accepted
   deployment/operations RFC composes and observes it.
4. Accept one operations contract covering metrics/export, request correlation, structured job
   identity, alert thresholds, retention jobs, failure escalation, incident/backup/restore runbooks,
   and privacy-safe log handling. Every counter/alert/cleanup needs a fired negative fixture.
5. Re-verify the sunset sources and rewrite the proposal's “already exists / near-zero” premise
   before adopting a covenant. Build and rehearse the deliverable under Deployment plus the future
   sunset RFC; research prose is not a release artifact.

This audit does not choose a retention duration, logging vendor, monitoring stack, license,
notice period, RPO/RTO, or release label. Those choices belong to the named owner gates.
