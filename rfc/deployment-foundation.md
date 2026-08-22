# RFC: Deployment Foundation

- **Status:** implementing — accepted at `cd102d7`; implementation opened 2026-08-22
- **Author:** Marco (initial draft by Claude; 2026-08-22 reconciliation delegated to Codex)
- **Created:** 2026-08-03
- **Design refs:** `design/06 §stack` (single-node Go/Postgres/Caddy Compose), `design/00`
  pillars 2/6/7 (bot fallback, provider-off honesty, server authority)
- **Depends on:** Gameserver Composition (implemented); Game UI Screens (implemented); Transport
  drain/recovery (implemented, active RFC lifecycle); Account Bootstrap and API Foundation
  configuration primitives (active); CI Baseline and D-014 fast/slow topology (implemented,
  lifecycle active)
- **Parent / amends:** —
- **Supersedes / superseded by:** —
- **Planning:** `planning/deployment-foundation/`

## Summary

This RFC defines the supported Cloud Clicker Phase-0 Playable Preview self-host package: one Linux
Docker Compose node running Caddy, the gameserver and Postgres, with an operator-owned backup,
restore, rollback, rotation, metrics and alert path. The public repository and hosted CI already
exist. They are useful source/verification facts, not evidence that a deployable product exists.

The package becomes a supported self-host deliverable only after its real clean-host and recovery
rehearsal passes R-006. This RFC makes no public-hosting or sunset-covenant promise.

## Motivation and scope

At current HEAD the gameserver binary, migrations, readiness, bounded drain and hosted CI are real.
Every Compose file is test-only; the binary depends on repository files; no production container,
static-client/Caddy bundle, backup, restore, rollback, rotation runbook, license delivery or fired
alert path exists. The old draft incorrectly called the repository local-only, described the push
as future work, claimed catalogs were embedded and treated `make verify` as one hosted job.

### In scope

- a reproducible Linux/amd64 Phase-0 release bundle and pinned container set;
- one-node Compose networking and Caddy TLS/static/reverse-proxy behavior;
- repository-independent runtime content and client/license delivery;
- fail-closed deployment configuration and current/previous key composition;
- stop-drain-start release and previous-release rollback;
- encrypted off-host backup, restore, retention and RPO/RTO rehearsal;
- private metrics, bounded local logs, alert evaluation and operator delivery;
- provider-off installation and manual/scheduled release evidence; and
- exact negative fixtures proving every gate can fail.

### Out of scope

- an official hosted public service, CD or host provisioning;
- multi-node/rolling deployment, Redis, Kubernetes or automatic failover;
- email/identity/analytics/AI/payment providers;
- player export/import UI or account deletion/retention semantics;
- gameplay telemetry, public UGC or later-tier content;
- a sunset covenant, notice period, final-world artifact, archive mirror or ratchet promise; and
- unsupported host/architecture claims. Linux/arm64 may be added only after the same rehearsal.

## Specification

### DP1 — Supported release bundle

A versioned release bundle is the unit an operator downloads and retains. It contains:

1. `compose.yml` for the exact production services and private networks;
2. a pinned gameserver OCI image containing the Go binary plus the immutable epoch/catalog,
   moderation and transport/API policy files it consumes at runtime;
3. a pinned upstream Caddy OCI image plus the immutable Caddy configuration and built client
   assets in the bundle;
4. a pinned Postgres 16 image;
5. the backup/restore/release helpers and operations-profile configuration;
6. `.env.example`, a machine-readable non-secret configuration schema and a preflight command;
7. `third-party-licenses.txt`, the root MIT license, image/source provenance and an SBOM; and
8. `release-manifest.json`, binding release version, commit, image digests, schema floor, epoch ID,
   artifact hashes, supported Docker Engine/Compose versions and supported `linux/amd64` platform.

The gameserver image may package immutable files on disk; this RFC does not require Go `embed.FS`.
It must start without a repository checkout or writable source tree. Caddy serves the SPA and
proxies `/api/`, `/connection/websocket`, `/healthz` and `/readyz` to the private gameserver.

Every image uses an immutable digest in the released Compose file. The implementation records the
exact Docker Engine and Compose versions used by R-006; “recent Docker” is not a support contract.

### DP2 — One-node network and origin boundary

Only Caddy publishes host ports, normally `80/tcp` and `443/tcp`. Gameserver, Postgres, Prometheus,
Alertmanager, node-exporter and backup services remain on private Compose networks and publish no
host port. Postgres has a named persistent volume. Backup output uses a separately mounted target;
the Postgres data volume is not a backup destination.

Production configuration accepts exactly one canonical `https://` public origin. Caddy appends the
forwarded chain, and the gameserver trusts exactly one proxy hop. Direct gameserver access is
structurally unavailable from outside Compose. A separate development profile may permit the
declared localhost origin and zero trusted hops; those values are rejected in production mode.

Caddy owns TLS. COOP/COEP headers remain off unless a shipped tenant requires them; enabling them is
a later reviewed configuration change because they alter browser resource isolation.

### DP3 — Fail-closed configuration and secrets

The release provides one canonical configuration schema and a `validate-config` preflight using
the same decoder/validators as startup. Unknown keys, duplicate keys, malformed values, insecure
production origin, proxy depth other than one, missing files and unreadable secrets fail nonzero
before migration or listener startup.

The production names and ownership are exact:

| Name | Source | Rule |
|---|---|---|
| `CLOUD_CLICKER_DEPLOYMENT_MODE` | environment | literal `production`; any other value selects a non-release profile |
| `CLOUD_CLICKER_PUBLIC_ORIGIN` | environment | one canonical `https://` origin, no path/query/fragment |
| `CLOUD_CLICKER_TRUSTED_PROXY_HOPS` | environment | literal `1` in production |
| `CLOUD_CLICKER_CONTENT_ROOT` | environment | read-only packaged path `/opt/cloud-clicker/content` |
| `CLOUD_CLICKER_SERVER_ID` | environment | canonical UUID |
| `LISTEN_ADDR` | environment | private container listener, default `:8080` |
| `DATABASE_URL_FILE` | environment → secret path | mounted file containing the Postgres connection URL |
| `CLOUD_CLICKER_JWT_CURRENT_ID` / `_KEY_FILE` | environment + secret path | required current HS256 identity/value |
| `CLOUD_CLICKER_JWT_PREVIOUS_ID` / `_KEY_FILE` | environment + secret path | optional pair; both or neither |
| `CLOUD_CLICKER_BOOTSTRAP_CURRENT_ID` / `_KEY_FILE` | environment + secret path | required current AES-256-GCM identity/value |
| `CLOUD_CLICKER_BOOTSTRAP_PREVIOUS_ID` / `_KEY_FILE` | environment + secret path | optional pair; both or neither |
| `CLOUD_CLICKER_CURSOR_CURRENT_ID` / `_KEY_FILE` | environment + secret path | required when public cursor readers are composed |
| `CLOUD_CLICKER_CURSOR_PREVIOUS_ID` / `_KEY_FILE` | environment + secret path | optional pair; both or neither |
| `CLOUD_CLICKER_BACKUP_TARGET` / `CLOUD_CLICKER_AGE_RECIPIENT` | environment | separately mounted target and public X25519 recipient |
| `CLOUD_CLICKER_ALERTMANAGER_CONFIG` | environment | mounted receiver config; no receiver credential in `.env` |

The current legacy single-key variables may remain available to local/test profiles while callers
migrate, but production preflight rejects them. Non-secret configuration also includes:

- deployment mode, public origin, listen address and server ID;
- repository-independent packaged-content root;
- Postgres database/user/host/name without password;
- current and previous JWT/bootstrap/cursor key IDs;
- backup schedule/target/age recipient and retention values fixed to DP6;
- release/rollback window fixed to DP5; and
- Prometheus/Alertmanager paths plus the named receiver-health check.

Secret values enter through Compose secrets mounted under `/run/secrets`; `.env`, command-line
arguments, logs, images and the release bundle contain no secret value. Runtime composition must
support:

- Postgres password/URL material;
- JWT current key and optional previous key;
- bootstrap-receipt current key and optional previous key;
- public cursor current key and optional previous key; and
- the restore-only age identity, which is never mounted into the continuously running gameserver.

Current/previous IDs must be distinct and match the supplied values. Supplying only half of an
optional previous pair or duplicating current/previous IDs or values fails closed.

### DP4 — Secret rotation contract

Marco owns rotation for an official instance; each self-host operator owns their instance.
Rotation is an operator sequence, not in-band client negotiation:

1. deploy a new current key while moving the former current key/value to previous;
2. confirm readiness and old/new verification fixtures;
3. retain the previous JWT key for at least 30 minutes, bootstrap key for at least 31 days and
   cursor key for at least 366 days; then
4. remove the previous key and prove an old credential/cursor rejects while current material works.

The release helper maintains an append-only `rotation-ledger.jsonl` in the operator-state volume,
recording key family/IDs, activation/removal times and operator identity, never values. It refuses
previous-key removal before the applicable minimum; tests use its governed clock rather than
waiting in wall time. These overlaps cannot be shortened by implementation convenience. Public cursor
rotation remains inactive until the public reader runtime exists, but its config contract must not
be replaced by an ephemeral restart secret.

### DP5 — Stop-drain-start release and rollback

There is no automatic production deployment. The supported release helper performs:

1. config, receiver and free-space preflight;
2. an encrypted pre-upgrade backup under DP6;
3. image pull plus digest, SBOM, license and release-manifest verification;
4. SIGTERM-driven gameserver replacement, consuming the shipped sequence: readiness down,
   `server_restarting`, new-intent refusal, admitted-work completion, jobs/outbox flush and bounded
   socket close;
5. new gameserver startup, forward-only migrations, epoch/artifact reconciliation and readiness;
6. an authenticated HTTP/WebSocket smoke path through Caddy; and
7. an append-only local release record with version, digests, backup ID, timestamps and result.

Phase-0 is stop-drain-start and may have downtime. The helper fails without marking success if
drain exceeds its bound, migration fails, epoch/artifact bytes disagree, readiness never rises or
the smoke path fails.

The current and immediately previous release bundle are supported for seven days. Rollback never
runs a Down migration against the live database. It stops the failed release, restores the exact
pre-upgrade backup into a clean Postgres volume, starts the manifest-bound previous images, checks
epoch/artifact identity and runs the same smoke path. A migration or content change that cannot be
restored this way blocks release rather than silently abandoning rollback.

### DP6 — Backup, restore and objectives

The backup worker creates an encrypted off-host backup every six hours. The release helper creates
one immediately before every upgrade. Each backup contains:

- a consistent Postgres custom-format dump covering every database object and row;
- the release manifest, epoch declaration and immutable artifact hashes needed to interpret it;
- creation/start/end timestamps and the source server ID; and
- a plaintext envelope identifier plus encrypted authenticated payload and checksum.

The payload uses the standard age file format and an operator-supplied X25519 recipient. The age
identity stays off-host and is mounted only for an explicit restore. Backup success is committed
only after dump, encryption, checksum verification and off-host-target rename complete atomically.
A partial temporary file is never a valid backup.

Retain the six-hourly population for seven days and one completed daily backup for 30 days. A
retention pass never deletes the newest valid backup or an unresolved pre-upgrade backup. Missing,
late, failed, truncated, corrupt, undecryptable and wrong-manifest backups are observable states,
not candidates to skip quietly.

Supported objectives are RPO 6 hours and RTO 4 hours. These are acceptance bounds over the
rehearsed package, not a public availability SLA. R-006 measures from declared incident/restore
start through the authenticated Caddy smoke path and verifies restored account, Founder, Company,
epoch, event and leaderboard identity.

### DP7 — Provider-off observability and alerts

The operations profile runs local Prometheus, Alertmanager and node-exporter services on the
private network. No cloud monitoring account is required. The gameserver exposes a private metrics
endpoint; Caddy must not proxy it publicly. Metrics and alert labels must not contain recovery
codes, credentials, raw IPs, account/founder IDs, request payloads or unbounded user-controlled
values.

The supported metrics cover at least process/readiness, HTTP/WebSocket error and latency classes,
Postgres reachability, job success/failure, composed credential cleanup, outbox/dead-letter
counts, last successful backup timestamp, restore/release result, filesystem use and restart count.
This is operational instrumentation, not gameplay telemetry.

All containers write structured logs through the host journald driver. The supported-host preflight
requires time-based retention through day 14 and purge after day 14, with enough reserved journal
budget for the measured release workload. Storage pressure that threatens earlier eviction fires
before deletion; it cannot silently shorten the evidence window. Raw IP material is disabled by
default; if the operator explicitly enables bounded security logging, its separate sink retains
through day 7 and purges after day 7. No log field may contain a secret or recovery code.

Alertmanager must have a configured receiver and a successful test delivery before an instance is
called operational. For the official instance that receiver reaches Marco; a self-host operator
names their receiver. The receiver may be local or remote, so provider-off operation remains valid.
The following alerts are blocking release-floor alerts and need fired fixtures:

1. public endpoint or readiness unavailable for five consecutive minutes;
2. a scheduled backup missing its declared six-hour completion deadline or any backup failure;
3. Postgres unreachable for two consecutive minutes;
4. persistent-volume or backup-target use above 80%;
5. gameserver restart loop (three restarts inside ten minutes);
6. any composed cleanup job failure; and
7. dead-letter population growth across two consecutive collection intervals.

Alert resolution is also recorded. R-007 later validates incident delivery and the still-open
product-retention families; this RFC does not decide account/history/moderation deletion policy.

### DP8 — CI, release evidence and publication boundary

The existing D-014 topology remains binding:

- push/pull-request CI keeps its strict sub-five-minute measured target and fast blocking jobs;
- exhaustive harness/numeric work remains scheduled/manual with bounded uploaded artifacts; and
- no workflow deploys or receives deployment secrets.

Static config/schema, secret-scan and reproducible-bundle checks may enter blocking CI only if the
complete hosted workflow still meets D-014. Clean-host, destructive restore, rollback and full
alert rehearsals run in a bounded manual/scheduled release lane and produce retained artifacts.
This RFC does not authorize a timeout increase or moving exhaustive work back into push CI.

The repository is already public and MIT-licensed; that fact is retained as AC9 evidence. A release
may not say “supported self-host,” “release-ready” or promise a sunset covenant until AC1–AC8 and
R-006 pass for the exact release bundle. Public source is not substituted for that proof.

## Deviations from design

- `design/06` says production uses `go:embed`; this RFC permits immutable files packaged inside the
  image because the user outcome is repository-independent, byte-bound runtime content. Requiring
  a compile-time embed would add no operator guarantee.
- `design/06` calls for one game binary plus Postgres/Caddy. The required operations profile adds
  local Prometheus, Alertmanager, node-exporter and a backup worker without changing the gameplay
  data plane or introducing a mandatory provider.
- Multi-node/rolling delivery remains deferred even if measured concurrency later approaches the
  design estimate.

## Acceptance criteria

Every executable criterion includes a demonstrated severing/failure case in the reviewed range.
Warm-cache success is not accepted where a cold or clean-host population is named.

1. **AC1 — reproducible clean-host package:** on a clean supported Linux/amd64 host with only the
   documented Docker/Compose prerequisites and release bundle, the pinned stack reaches readiness
   and the Phase-0 browser flow works through Caddy without a repository checkout. Removing one
   packaged catalog/client/license/config artifact fails build or startup.
2. **AC2 — network/config/secrets:** only Caddy exposes host ports; production origin and one-hop
   proxy behavior pass through real HTTP/WebSocket clients. Every required-secret/config family has
   missing, malformed, duplicate-ID/value and insecure-origin negatives. Image and tracked-source
   scans reject a seeded secret.
3. **AC3 — release/drain:** the helper performs preflight, pre-upgrade backup, real
   `server_restarting` delivery, bounded drain, migration/epoch sync, readiness and authenticated
   smoke. Severing the courtesy frame, drain wait, migration failure propagation, epoch check or
   smoke result fails the release artifact.
4. **AC4 — backup/restore/RPO/RTO:** empty and populated databases back up and restore within 6-hour
   RPO/4-hour RTO. Restored player/Founder/Company, events, board and epoch identities match.
   Truncation, corruption, wrong age identity, wrong release manifest, partial rename and
   mid-write/restart fixtures fail loudly.
5. **AC5 — rollback:** the immediately previous release and pre-upgrade backup restore on a clean
   volume and pass the same smoke path without a Down migration. Wrong-image, missing-backup and
   irreversible-migration fixtures refuse rollback/release.
6. **AC6 — rotation:** current+previous JWT/bootstrap/cursor fixtures work for their governed
   overlap; removal rejects old material. Missing runtime wiring, shortened overlap, logged secret
   and restart-ephemeral cursor mutations fail.
7. **AC7 — operations:** private metrics are reachable only inside Compose; ordinary logs remain
   available through day 14 then purge, and optional raw-IP security logs remain through day 7 then
   purge. Early-eviction, stale-retention and storage-pressure fixtures discriminate. All seven
   alerts fire through the configured receiver and their severed counter/rule/receiver paths fail
   the Deployment rehearsal. R-007 may consume that evidence later but is not this RFC's gate.
8. **AC8 — provider-off and supply chain:** install, play, backup, restore, alert and rollback work
   with no identity/mail/analytics/AI/payment/cloud-monitoring credential. Image digests, SBOM,
   MIT/third-party license delivery and source/image provenance validate; removing attribution or
   changing an image digest fails.
9. **AC9 — existing external-state fact:** retain the already-public origin and hosted CI evidence
   without treating them as proof of AC1–AC8.
10. **AC10 — claim gate and closeout:** R-006 passes against the exact release manifest; canonical
    deployment/operations docs and limitations are current; backlog/queue/RFC/log records reconcile
    transactionally; the complete implementation range receives Codex first-filter and Claude
    designated cross-party approval before archival.

## Open questions

None for acceptance. Account export/deletion and product-data retention are explicitly outside this
RFC and remain blocked under D-008/D-009/D-015. A sunset covenant remains a later owner decision
after the supported deliverable exists.

## Changelog

- 2026-08-03: created around a deploy pipeline and then-pending push.
- 2026-08-06: non-normative reference cleanup for publication.
- 2026-08-22: owner-delegated rewrite after D-003/D-006/D-011/D-014. Removed the obsolete push and
  embedded-artifact claims; specified the one-node package, origin/proxy boundary, fail-closed
  config, rotation overlaps, release/rollback, encrypted backup objectives, provider-off operations,
  alert set, claim gate and discriminating acceptance population. Status remains draft pending
  owner acceptance.
