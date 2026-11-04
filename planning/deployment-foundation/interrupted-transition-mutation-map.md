# Interrupted release/rollback: mutation ownership map

Source audit at `a0dc53bcc25e16860d5017f7af28be4842afd0d9`, 2026-11-04.
This maps the existing release/rollback call graph. It is research for draft CR4,
not a recovery implementation, dependency-binary attestation or adoption.

The network experiment supplies one NEW-client admission boundary. Combining it
with a terminated restore backend is insufficient to classify the whole release:
HTTP smoke creates an account, container startup runs migrations/jobs, and helpers
write host-mounted backup/operations files. Each must have an explicit disposition.

## Existing operation sequence

[`Controller.Release`](../../server/deploymentrelease/controller.go) calls Prepare →
Preflight → CreatePreUpgradeBackup → DrainCurrent → Start → VerifyIdentity →
AuthenticatedSmoke → append terminal ledger. Rollback calls Prepare → Preflight →
VerifyRestoreInputs → StopFailed → ResetDatabase → Restore → Start → VerifyIdentity
→ AuthenticatedSmoke → append terminal ledger. All occur under the operator lock;
that lock does not supervise Docker jobs or database backends.

## Runtime boundaries

The entries below are **direct source observations**, not new runtime executions.
The Runtime implementation is [`docker.go`](../../server/deploymentrelease/docker.go);
[`command.go`](../../server/deploymentrelease/command.go) supplies the actual child
process and fixed `cloud-clicker` project selection.

| Call | Actual effect and target | Required interrupted-operation disposition |
|---|---|---|
| Prepare | Docker archive load, pinned upstream pulls, image inspection | Classify pending image operations separately from live deployment mutation. Any claim that a late load/pull is harmless needs immutable config-ID consumption, not a tag-only assumption. |
| Preflight | Compose config; gameserver validate-config and amtool check-config one-offs; backup inspect one-off; receiver-health HTTP GET; local bundle/free-space reads | Even validation spawns daemon containers and database clients. Bind and account for jobs and connections; a health response is not job quiescence. |
| CreatePreUpgradeBackup | Compose backup create; exported read-only SQL snapshot plus pg_dump; encrypted backup publication; operations textfile update | Own job/container, snapshot/dump sessions, host output paths and metrics writer. Terminal pg_dump alone is not completion of its enclosing Go backup helper. |
| DrainCurrent | **HTTP bootstrap creates an account**; state GET; connected/subscribed WebSocket; Compose gameserver stop; readiness polling; exact-container exit inspection | Treat bootstrap as a mutating HTTP request, not a read-only probe. Account for admitted server/SQL work, socket and service shutdown as well as the Docker client. |
| Start | Compose force-recreates six non-Postgres services; their entrypoints start migration/epoch reconciliation, background jobs and host-mounted writers | Account for container create/start requests, every configured network, volume and bind target, and the resulting processes. A late create that returns a container ID is not evidence that startup was prevented. |
| VerifyIdentity | Local pinned epoch load plus backup inspect one-off querying the live DB | Bind the observing job/connection; identity agreement is not proof that no old mutation remains. |
| AuthenticatedSmoke | Another mutating HTTP bootstrap/state/WebSocket session, then an alertmanager one-off posting the test alert | Include application requests and the alert-delivery job. A healthy GET or closed local socket does not prove all remote work terminal. Do not execute a real external receiver merely to fill this map. |
| VerifyRestoreInputs | Reads credential/backup; decrypts/extracts into a private local temporary directory, then removes it | Own local process and scratch lifetime. This is not a Docker/SQL mutation and does not adopt D-019's separate decryption-placement decision. |
| StopFailed | Compose down with remove-orphans, including one-offs | Need exact selected resource identities and pending-request disposition. Project spelling alone is not that proof; do not assume every underlying Engine call is name-addressed either. |
| ResetDatabase | Removes the fixed named volume `cloud-clicker_postgres_data`, then Compose starts Postgres | Volume-name reuse and database initialization are explicit boundaries. A retired network ID neither identifies the volume generation nor terminates an already-admitted DB backend. |
| Restore | Compose backup restore; private decrypted dump; clean-target checks; pg_restore writes; post-restore migration check; operations textfile update | Bind job, backend, target DB/storage generation and host writer. RP-726 proves client removal alone is insufficient. Cancellation is not rollback of earlier committed DDL. |
| Success/failure ledger append | Local Go writes/fsyncs terminal JSON line; the helper also owns the operator lock | Durable event ordering belongs to CR3. A dead helper/unlocked file is not a completion record; torn/unsynced writes cannot mint recovery authority. |

## Nested helper and persistent-resource boundaries

[`deploymentbackup/postgres.go`](../../server/deploymentbackup/postgres.go) opens
database connections AND invokes pg_dump/pg_restore child processes. Its private
credential/dump workspaces are separate from the published backup target.
[`backup.go`](../../server/deploymentbackup/backup.go) publishes through a
no-replace rename and directory sync. Those protect competing backup commits;
they do not supervise a surviving daemon job.

[`cmd/deployment-backup/main.go`](../../server/cmd/deployment-backup/main.go)
records backup/restore success or failure through
[`operations/textfile.go`](../../server/operations/textfile.go), whose temporary
file is renamed over the final metric file. Include that host writer even when
the failed operation never reaches a successful SQL result. This audit does not
claim that an actual old job overwrote a newer metric.

The release [`Compose template`](../../deployment/compose.template.yml) has FOUR
networks: edge/application/database/operations. The database is not the sole
connectivity boundary. Five persistent volume families and host-mounted backup,
metrics, configuration/content and secret inputs also need their own treatment.
Read-only inputs must remain distinguished from writable targets. A single
retired default network in a fixture does not cover this production topology.

Initial install/AbortInstall additionally creates Postgres and can down the
project **with volumes**. CR4's proposed abort does not cover interrupted initial
install. StartRecoveryCore/CreateRecoveryBackup/InspectRecoveryIdentity/KillGameserver
are rehearsal/recovery entry points outside the two Controller call sequences;
they are not silently included in a proof of ordinary release/rollback. Rotation
shares the operator lock but has separate policy/secret ownership. Extending the
mechanism to those actions requires explicit coverage, not reuse of the table's title.

## What upstream Compose source narrows—and does not prove

At upstream **v2.39.4**, stop enumerates containers, then submits ContainerStop
using the selected container ID. Down also removes selected containers by ID;
network removal uses an inspected network ID. Thus "Compose always reselects a
new container by service name inside the daemon" would be false. This source
inspection is not a runtime witness for the installed desktop.1 binary.
[stop source](https://github.com/docker/compose/blob/v2.39.4/pkg/compose/stop.go),
[down source](https://github.com/docker/compose/blob/v2.39.4/pkg/compose/down.go).

Creation's default network settings use the configured network name; volume
removal accepts its configured name. Those are different identity boundaries from
container IDs. Do not infer from the network-ID experiment that the existing
Compose path pins that ID, or that name-addressed volume removal has an equivalent
generation guard. [create source](https://github.com/docker/compose/blob/v2.39.4/pkg/compose/create.go),
[volume-removal source](https://github.com/docker/compose/blob/v2.39.4/pkg/compose/down.go).

## Combined mechanism still required

A candidate must give every table row either a proven terminal lifetime or a
proven inability to affect the recovery generation. Merely putting network removal
and backend termination next to each other is not that argument. In particular:

1. Stop further submissions from the entire owned local client process scope;
   a root PID, lock or process-group assertion is not sufficient without its proof.
2. Bind already-submitted Engine actions to exact non-reused resource identities;
   handle unacknowledged creates, service processes and network/volume replacement.
3. Account for already-admitted SQL AND mutating HTTP requests after client loss.
4. Fence host-mounted writers and preserve encrypted backup/checkpoint authority.
5. Only after those conditions hold may the governed recovery operation mutate
   the project. If ownership/terminal outcome is unknown, refuse before effects.

No candidate is declared complete here. The next construction is one combined
protocol with this whole mapping, followed by a controlled interruption whose
delayed old actions are opened AFTER recovery admission. Keep the old admission,
backend, name-only and host-writer negatives; do not promote independent green
pieces into a whole-operation proof. Native supported-host and exact-artifact
R-006 remain mandatory after technical review/adoption and implementation.
