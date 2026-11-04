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
The complete original plus correction through ee90913f is independently APPROVED
after fresh real57617 and actual-source/refusal/full-pipe controls21d8f5. Original
CHANGES REQUIRED stays recorded. RP-739's instrument defect is closed; RP-738's
production ownership/startup requirement remains open.

### Connection-owned cancellation candidate

The [manual gateway](probes/connection-cancel-gateway.go.txt),
[build/fixture runner](probes/connection-cancel-gateway.py.txt) and
[source-bound observation](connection-cancel-gateway-observation.v1.json) now
provide a candidate that avoids the preceding PID-only signal. A transparent
single-connection gateway registers its connection in memory BEFORE dialing
Postgres. It observes BackendKeyData without printing/persisting its secret,
keeps draining the original server stream after actual pg_dump is killed, and
uses that connection's PID/key pair in the PostgreSQL CancelRequest protocol.
The [PostgreSQL16 protocol](https://www.postgresql.org/docs/16/protocol-flow.html#PROTOCOL-FLOW-CANCELING-REQUESTS)
explicitly distinguishes request delivery from cancellation; the test does too.

Actual35493 completes every predeclared population against pinned PG16.15:

| Population | Actual result |
|---|---|
| Wrong cancellation key, relation lock held | Delivery connection closes normally, yet original dump remains blocked. Original server EOF occurs only after the driver releases the lock. |
| Correct key, same lock held | Original server EOF occurs BEFORE lock release; no PID-only termination is issued. |
| Startup delay8s, no key/activity row yet | Killed client does not clear gateway registration. The original connection remains tracked through actual server EOF, without pretending a key exists or sending a cancellation. |

Every population retains the same continuously connected control PID and
postmaster and exact original/control rows7/42. Actual native control overlay
also rejects truncated/invalid/oversized frames, duplicate/bad BackendKeyData,
unexpected cancel responses, exhausted delivery and late handshakes; partial
writes and ordinary EOF pass. Neither socket-delivery EOF nor census0 is admitted
as original-session completion. The actual subject is pg_dump, NOT the complete
Go helper's pgx exporter plus dump, encrypted publication or production lifecycle.

### Concrete production direction to specify and prove

Select a **fail-closed supervised operation**, not a detached CLI assumed dead
from a lock. This remains a recommended technical direction requiring full
specification, independent review and adoption; no runtime permission follows:

- One operation supervisor must outlive the invoking CLI and own the real helper,
  all descendants, outer metrics and operation jobs. Record its exact native
  identity durably before effects. A per-operation native unit/scope needs real
  supported-host packaging and failure tests; this does not silently introduce a
  permanent deployment service under the draft's current no-service promise.
- Every mutating Engine request needs a durable intent before forwarding and an
  endpoint-specific terminal receipt/resource binding. After cancellation close
  worker admission, keep the supervisor/broker alive to collect already-forwarded
  requests, then stop/reap exact owned job namespaces. CLI exit or an arbitrary
  HTTP error must not erase outstanding work. Unknown/lost authority refuses a
  successor before new effects; it never means absence. Native worker isolation
  must prevent bypassing that broker, not merely set DOCKER_HOST.
- Route ALL operation-owned SQL connections through registered gateways, including
  pgx inspection/exporter and libpq. Close client admission before cancellation;
  wait actual original streams and worker actors terminal. Keys stay in memory;
  loss of the gateway/supervisor cannot be repaired by guessing a PID or accepting
  a reused application name. The single-connection prototype proves neither
  multi-connection completeness nor durable admission. Native network isolation
  must prove workers cannot reach Postgres directly.
- Before-checkpoint abort may stop only those operation jobs/connections/output
  writers while retaining the prior live deployment. Post-checkpoint recovery
  still validates protected backup/exact pair before permitted teardown/reset.
  Account-creating HTTP, all four networks/five volume families, image preparation,
  retention deletion, actual outputs and resource ownership remain in the reviewed
  mutation map. Never apply this read-only SQL proof to those other effects.

Unknown supervisor/gateway/request outcomes remain unfinished and refuse automatic
recovery; the separate disaster-recovery route is not invented by this direction.
The next construction is the real supervisor/broker ownership and admission
protocol with lost-authority refusal, followed by full real-helper binding. Do
not repeat the completed pg_dump controls instead. RP-740 tracks the candidate;
RP-738/734 remain open. Full native-host/R-006/adoption/release gates remain.

The subsequent [CR1–CR6 ordering correction](../../rfc/deployment-current-release-authority.md#cr6--whole-operation-closure-and-supervisor-authority)
is independently technically approved c5419034..0a89f026: healthy-current
admission differs from recovery ownership; predecessor closure differs from
current-helper completion; admitted cleanup precedes recovery_fenced, while
reset/restore follows it; successful descendants derive parent recovery without
fabricating Finished. RP-741–744 close only those draft contradictions. Exact
supervision/work-log/proof schemas, first-install handoff, runtime/native paths
and adoption remain open. This is not a packaged supervisor or completed CR4.

The complete nine-path `20cbb5b8..fd4c5b92` batch is independently APPROVED after
fresh actual2276, fourteen controls42422 and source/runner/binary bindingsbc06cb.
This approves the experimental candidate and stated technical direction, not
production ownership, implementation/adoption or recovery acceptance.

Reproduce the three real populations from the repository root:

```sh
python3 planning/deployment-foundation/probes/connection-cancel-gateway.py.txt
```

Reproduce actual-function controls without changing production inputs on disk:

```sh
python3 - <<'PY'
import json, pathlib
r = pathlib.Path.cwd()
p = r / '.cache/connection-gateway-research'
p.mkdir(parents=True, exist_ok=True)
replacements = {
    str(r / 'server/cmd/deployment-backup/main.go'):
        str(r / 'planning/deployment-foundation/probes/connection-cancel-gateway.go.txt'),
    str(r / 'server/cmd/deployment-backup/main_test.go'):
        str(r / 'planning/deployment-foundation/probes/connection-cancel-gateway-controls.go.txt'),
}
(p / 'controls-overlay.json').write_text(json.dumps({'Replace': replacements}))
PY
make test-go GO_PACKAGES=./cmd/deployment-backup GO_TEST_FLAGS="-overlay=$PWD/.cache/connection-gateway-research/controls-overlay.json -run TestManualGatewayControls -v -count=1"
make vet GO_PACKAGES=./cmd/deployment-backup GOFLAGS="-overlay=$PWD/.cache/connection-gateway-research/controls-overlay.json"
```

### Isolated command-worker mechanism

The [manual worker study](probes/worker-isolation-probe.py.txt) and
[current source-bound observation](worker-isolation-observation.v3.json) test a
Docker-backed worker boundary for OP9. The [initial observation](worker-isolation-observation.v1.json)
binds committed sourceaa765a40 and retains a known
instrument defect: source-binding collection follows its last deadline check
(RP-749). Its actual native outcomes below are recorded, but final deadline
discrimination needs correction before instrument approval.
The corrected source checks AFTER final collection and enforces each command's
local completion deadline, including cleanup. Actual old-source controls accept
61s final collection/deadline60 and11s command/local10; current timely controls
pass and exact/late controls refuse. Fresh actual76328 completes every original
population with unchanged native outcomes; v2 binds source3d672aa9. Both versions
remain. Independent review of that union is CHANGES REQUIRED for RP-751:
actual92969 and diagnostic24761 fail on the same VALID broker mount set returned
in reverse order. Neither completes the after-control; no independent pass is
claimed. Source now compares destination-ordered projections, preserving exact
cardinality/values and rejecting duplicate/altered/extra mounts. Actual old/new
controlab38d7 discriminates both orderings and invalid sets; fresh actual14788
completes every population at source2cfcb55c, retained v3. Original negative/
v1/v2 remain. Complete original-plus-both-corrections union through299ad114 is
independently APPROVED after fresh12798 all populations,dd3dfd permutation/
corruption/binding controls and unchanged independent guard/broker controls.
Both failed partial reviewer runs remain disclosed. RP-749/751 close instrument
defects, not OP9/profile/adoption or actual CLI/supervisor/recovery acceptance.

Actual57766 compares UID65532/cap-drop writable network controls BEFORE and AFTER
an isolated network-none/read-only worker. Both controls reach real pinned
Postgres16.15, receive the expected SSLRequest refusal byteN, and write/read their
local marker. The isolated worker receives ENETUNREACH101 for the SAME live PG,
ENOENT2 for raw daemon paths, EROFS30 for `/tmp` and a world-writable channel file.
Actual kernel effective/bounding capabilities are0, NoNewPrivs1, and allowlisted
Engine inspection matches each image/user/network/mount/security profile.

The isolated worker DOES reach actual Engine200/OK through its private Unix
broker. The correct UID's wrong method and UID0's GET are distinct403 refusals;
broker obtains UID/GID from SO_PEERCRED, not a request field. Fixture socket/file
modes are deliberately0666 so refusal cannot be attributed to inaccessible
filesystem permissions; the worker's channel mount is read-only. The retained
marker remains intact. This supplies a candidate for isolated Docker command
jobs/per-job broker channels without selecting a new native host service.

It is NOT Docker/Compose CLI execution or the Go supervisor. UID alone is not
complete job identity: OP9 must bind each private channel to its exact admitted
job/namespace, prevent unintended mounts, enforce all stage-bound write methods/
bodies and preserve request lifetime/durable receipts. Supervisor/daemon epoch,
bootstrap, actual SQL/HTTP gateways/output publishers and all retained resources
remain unbound. No full bypass inventory, supported Linux/amd64 host, shipping
dependency, profile adoption or recovery acceptance follows from these Desktop
Linux/arm64 Python workers with amd64 PG. Do not repeat the completed native
positives instead of building those missing real producer bindings.

Next construct the complete durable production protocol against this table and
resolve its pre-backup branch/packaging/compatibility, then execute its actual
helper and supported-host interruption population. Do not repeat these complete
controlled populations or declare full CR4/Deployment/1.0 acceptance from them.

### Actual Docker/Compose clients through a Go broker

The [manual runner](probes/cli-worker-probe.py.txt),
[Go broker](probes/cli-worker-broker.go.txt) and
[source-bound observation](cli-worker-observation.v1.json) extend the prior
isolation study to actual Linux/amd64 Docker28.4.0 and Compose2.39.4 clients.
They are research inputs only; no daemon or host CLI is installed and no
shipping dependency changes. Actual cafdcb/418737 exits0. Complete new range
2699ad00..04452299 is independently APPROVED after fresh c8216a/f54d40 actual
population, f3ce08 actual-AST trace/profile/deadline/error and binding controls,
d605ec Linux/amd64 vet and66728a complete scope/archive/static-binary checks.
Exact reviewer project container/network/volume inventories are empty. This
does not extend the review to the still-unbuilt production profile.

An isolated UID65532 worker obtains the real Engine28.4.0 Linux/arm64 version
through the Go Unix/HTTP broker. Compose's real plugin version and config
render execute; config rendering is local, not proof of Compose up/stop/down.
The client's actual ContainerCreate POST reaches the broker and receives403
without forwarding; no requested child exists. UID0's actual version requests
are refused403 by SO_PEERCRED. The same successful Docker executable cannot
use the unmounted raw socket (native ENOENT2). Native profiles and kernel
capability/no-new-privileges observations are retained for both workers.
Eleven actual Go classifier controls include wrong method/query/API version/
unowned image and write-route refusals. Cross-target overlay vet0df9ed passes.

The retained trace contains categories/status/UID only, no request bodies,
environment or credentials. A denied unknown GET is retained rather than
silently excluded; it does not authorize another route. This narrow broker
permits only ping/version and exact cached fixture-image inspection. UID and
a shared fixture channel are not a complete per-job identity or write policy.
There is no durable admission journal, daemon-generation/lock ownership,
actual release-helper binding, whole-operation completion receipt, native
Linux/amd64 host proof or OP9 adoption here. Workers remain on Desktop's
Linux/arm64 kernel, using amd64 static clients via its binary compatibility.

Reproduction requires the already cached Python image/config from the worker
study, Docker Desktop with the observed Engine version, public research clients
and no real credentials. From repository root, prepare ignored tool storage:

```sh
mkdir -p .cache/worker-cli-tools/plugins
curl --fail --location --max-time 180 --output .cache/worker-cli-tools/docker-28.4.0.tgz https://download.docker.com/linux/static/stable/x86_64/docker-28.4.0.tgz
curl --fail --location --max-time 180 --output .cache/worker-cli-tools/plugins/docker-compose https://github.com/docker/compose/releases/download/v2.39.4/docker-compose-linux-x86_64
python3 - <<'PY'
import hashlib, json, pathlib, tarfile
root = pathlib.Path.cwd()
tools = root / '.cache/worker-cli-tools'
metadata = json.loads((root / 'planning/deployment-foundation/cli-worker-observation.v1.json').read_text())['tools']
archive = tools / 'docker-28.4.0.tgz'
assert hashlib.sha256(archive.read_bytes()).hexdigest() == metadata['docker_archive']['sha256']
with tarfile.open(archive) as source:
    member = source.getmember('docker/docker')
    assert member.isfile() and not member.issym() and not member.islnk()
    (tools / 'docker').write_bytes(source.extractfile(member).read())
for item in metadata['binaries']:
    binary = root / item['path']
    assert binary.stat().st_size == item['bytes']
    assert hashlib.sha256(binary.read_bytes()).hexdigest() == item['sha256']
    binary.chmod(0o555)
(tools / 'tools.json').write_text(json.dumps(metadata, indent=2) + '\n')
PY
python3 planning/deployment-foundation/probes/cli-worker-probe.py.txt
```

The Compose binary digest was checked against the official GitHub release
metadata. Docker's archive came from its official HTTPS static index; its local
digest is retained but no separate upstream checksum/signature was checked.
These are testing inputs, not verified production packaging provenance. The
runner checks tool hashes before use, builds the Go binary via the existing
root Make Linux/amd64 target with an ignored compiler overlay, and keeps all
production inputs untouched. Build180s precedes fixture work60s, ordinary
commands10s, cleanup15s per command outside work. Completed output follows
verified owned cleanup. Unknown outcomes and guard failures invalidate.

The earlier interrupted execution has no admissible result: output was lost;
later process/owned-resource inventory was empty and a scratch directory alone
does not establish success. A pre-commit source check found the container-log
assertion expecting3 lines although docker exec's controls stdout is separate.
It now requires the actual2-line ready/trace shape; an old/new shape control
reproduces the false rejection. The retained complete run follows correction.
Do not repeat these compatibility positives instead of constructing the missing
durable per-job producer/admission/receipt bindings.
