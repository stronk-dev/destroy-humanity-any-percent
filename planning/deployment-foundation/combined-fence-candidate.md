# Combined interruption candidate: contained clients and resource generations

2026-11-04. Research for draft CR4; **not an adopted or implemented release
protocol**. The source-bound [three-population observation](combined-fence-observation.v3.json)
comes from the [manual prototype](probes/combined-fence-prototype.py.txt).
The [original observation](combined-fence-observation.v1.json) is historical and
bound to source98af4c8d. Its volume-inspection branch could conflate ordinary CLI
failure with absence (RP-735). The corrected source accepts only actual Engine404
as absence, validates200 ownership, rejects other failures and retains typed
responses. [Version2](combined-fence-observation.v2.json) preserves that correction's
actual run, but independent review finds generic late-start failure acceptance and
missing final collection-deadline refusal. Current source records the typed,
exact-container/retired-network start result and checks the deadline AFTER final
collection. Final full execution reproduces the same three outcomes below; old
source/observations remain historical rather than silently rewritten.

## Candidate and actual experiment

Contain the old helper and its descendants in one removable process namespace;
remove known daemon jobs/application and database process namespaces; bind delayed
jobs to a full retired network ID; never reuse the old volume name or writable
output directory for the replacement operation. A private proxy first receives
complete old create/start/volume-delete requests, then holds them until AFTER the
driver writes the replacement-admission marker. No receiver or production data is used.

The old database is actually executing pg_restore COPY behind an observed SQL
advisory lock. HTTP/file actors are controlled models, **not Account/bootstrap or
the real release CLI**. Replacement SQL is a new table/row42, not an encrypted
game-backup restore. The volume sentinel is an unattached prepared named volume,
not a running Postgres data mount. These distinctions limit the result.

| Population | SQL rows after old gates open | Current metric/checkpoint/HTTP marker/volume |
|---|---|---|
| Kill only helper parent | 7,13,14,42 | old / deleted / old / deleted |
| Remove contained clients and known jobs/old DB, but reuse names and output path | 14,42 | old / deleted / new / deleted |
| Same containment, plus full network ID, new volume name and separate output path | 42 | new / retained / new / retained |

In the first population, local descendants, known daemon writer and already-
admitted HTTP writer all write after admission. In the second they do not, but
the delayed create uses the recreated network name and can still reach the new
database/output path; the delayed delete removes the reused volume. The complete
candidate preserves every tested replacement value. Actual positive SQL/HTTP and
volume writes precede admission, so absence of old mutation is not inferred from
an unavailable target. Exact rows/bytes/start results are retained.

The late create returns201 in ALL populations. In the two contained populations,
the dead client never receives that ID and cannot submit its next request.
Parent-only descendants can survive and receive it, but the actor does not issue
a subsequent start. The driver deliberately
forces an ADDITIONAL start of that container as a stronger adversarial control:
it succeeds in the weak populations and refuses for the retired network ID.
Do not describe this forced start as a request the dead client had submitted.
The proxy delay is BEFORE daemon submission, not a witness of queued execution
inside the native daemon. The separate admitted COPY supplies the SQL lifetime.

This is on Docker Desktop Engine28.4.0 Linux/arm64/Compose2.39.4-desktop.1, using
the existing amd64 Postgres16 child and cached Python3.12 arm64 config. It is
mixed-architecture local evidence, not the supported Linux/amd64 host. Python's
cached config is `d2924abe506373b6f722045c48f98350077cc5b56a50491105c1abb027dce9b2`
(observed repository digest `python@sha256:e6dae589c1c05976188fd3bd019afd091445fc533ed2ea29001beddc7368a9c8`).
No new shipping image or dependency is approved by using it as a fixture.

## Isolation is different from literal quiescence

Even the successful population observes late ContainerCreate201 and old-generation
VolumeDelete204 AFTER admission. Therefore this is NOT proof that every old
daemon action became physically quiescent. It supports isolation of the tested
replacement resources, with terminal local processes/database, while allowing
quarantined old-generation metadata effects.

CR4 currently requires quiescence. An isolation-based production proposal must
explicitly reconcile that contract and justify every allowed late effect, with
independent review/adoption. It cannot close the existing criterion by silently
changing the meaning of "quiescent". Unknown effects on current/recovery resources
must still refuse before recovery mutation. No criterion is waived here.

## Production binding still required across the whole mutation map

The [reviewed map](interrupted-transition-mutation-map.md) remains the scope.
The prototype is one combined execution of effect classes, not the implemented
Controller/Compose protocol. Before a production proposal is acceptance-ready:

| Boundary | Concrete unresolved binding/proof |
|---|---|
| Admission and durable ownership | The driver config/admission file is not CR3's journal. Specify durably recorded supervisor identity, resource intents/bindings and recovery ordering BEFORE effects; unknown creates must not derive authority from labels alone. |
| Local helper and post-controller metrics | Supply a supported packaged supervisor/process scope for the actual Go CLI and all descendants. Successor admission must cover the outer writer even after a terminal Controller record; lock acquisition alone is insufficient. The fixture Python scope is not that artifact. |
| Container actions and all four networks | Bind every submitted create/start/stop/remove to exact ownership, full required network IDs and non-reused generations. Prove multi-network startup and all real Compose/helper paths; this fixture has only one data network plus its control network. |
| Persistent volumes | Name/address every one of the five persistent families and preserve required data/certificates/state. Test prepared and attached volumes, actual clean-volume rollback, pending removal and lifecycle cleanup. An unattached sentinel does not prove Postgres data preservation or clone-space/RTO feasibility. |
| SQL and application HTTP | The candidate kills the entire old DB/application namespace. Reconcile ordinary release, rollback and CR4's PRE-BACKUP abort separately: abort currently forbids running-service/data mutation, so the destructive prototype branch cannot implement it unchanged. Bind all actual connections/requests; no application-name-only authority. |
| Host output and scheduled deletion | Bind every writable mount and any publication to the authorized generation. Include real backup/metric writers and scheduled temporary/retention deletion, not just the modeled metric/checkpoint files. Preserve signed protection-metadata/D-019 holds; neither algorithm is tested or adopted here. |
| Image preparation | Classify late load/pull effects and prove immutable config consumption in actual startup. No image-load/pull interruption is exercised by this prototype. |
| Checkpoint/previous release authority | Validate actual protected encrypted backup/key and exact release pair before destructive reset. Row42 and a text checkpoint are not that authority. Keep the seven-day window, forward-only migration and clean-host gates. |

### Non-destructive pre-backup SQL observation

The [manual pre-backup probe](probes/prebackup-live-database-probe.py.txt) and
[source-bound observation](prebackup-live-database-observation.v2.json) address
one part of the incompatible destructive branch above. On the pinned image's
actual PostgreSQL16.15, a libpq client held by `post_auth_delay=8` has a live
backend process but ZERO rows for that PID in pg_stat_activity. Exact client
removal leaves that same process temporarily alive; it subsequently exits and
never executes its offered INSERT. Empty session metadata alone is therefore
not proof that every connection process has ended. This is a controlled startup
delay, not an observed normal backup or production incident.

A real pg_dump blocked on ACCESS SHARE survives exact client removal too. In
this isolated fixture, termination of its separately observed backend with a
3000ms wait removes that process. Releasing the exclusive-lock control then
preserves row7; a continuously connected independent control inserts row42.
The exact database container, postmaster start time and control PID/kernel-start
identity remain unchanged. No entire-DB restart or application disconnect is
necessary for THIS known-backend cancellation. SQL success/failure, OS process
identity and actual data are observed, rather than inferred from client exit.

This does NOT solve production ownership. The predicate combines observed PID,
backend_start, address/port and application name, but pg_terminate_backend
ultimately targets a PID; the predicate does not remove its check/signal PID-reuse
race. The startup process's log/address association is also only an owned-fixture
instrument, not durable operation registration. No new database permission,
unscoped termination or adoption is inferred. Pending starts/new connections,
authentication/startup completion, real pgx snapshot exporter, actual backup
helper/encryption/publication, image preparation and output generations remain
part of the complete production protocol. This probe neither implements abort
nor proves the full quiescence criterion. RP-738 tracks the new startup-census
finding and these remaining requirements.

The [original v1](prebackup-live-database-observation.v1.json) binds source83abd009;
its independent real run reproduces these outcomes, but original review is
CHANGES REQUIRED: interactive stdin writing was outside the local10s handshake
deadline (RP-739). Current source uses nonblocking pipe writes/select and one
deadline through writing, reading and final matching-output acceptance. Original
actual-source late acceptance is retained, new exact/late I/O controls refuse,
and fresh actual79437 reproduces both populations under corrected source7b93d963.
Neither observation is a whole recovery protocol or a runtime adoption.

Next construct the complete durable production protocol against this table and
resolve its pre-backup branch/packaging/compatibility, then execute its actual
helper and supported-host interruption population. Do not repeat these complete
controlled populations or declare full CR4/Deployment/1.0 acceptance from them.
