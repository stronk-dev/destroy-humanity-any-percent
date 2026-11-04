# RFC: Native Deployment Supervisor Profile

- **Status:** draft — prior reviews retained, including P01 throughdcc151f2; complete linked adoption package independently technically approved throughed9a8f2a; owner adoption/native construction/conformance pending
- **Author:** Codex
- **Created:** 2026-11-04
- **Parent / amends:** [Operation Evidence](deployment-operation-evidence.md), OP9; [Current Release Authority](deployment-current-release-authority.md), CR4/CR6; [Deployment Foundation](deployment-foundation.md), DP1/DP5/DP6
- **Design refs:** `design/06-tech.md` §stack; `design/00-vision.md` pillar6
- **Planning:** `planning/deployment-foundation/plan.md` and append-only log

## Scope and evidence boundary

Specify one native, one-operation supervisor and its complete admission boundary.
The ledger remains deployment authority; OP1–OP8 remain the record/reduction
contract. This draft supplies the proposed profile behind their hashes, not
permission to implement it before adoption. No new gameplay or retention policy,
signing key, automatic rollback or host provisioning is proposed.

The native carrier and original-SQL-stream studies prove mechanisms, not this
profile. In particular, neither study proves authenticated bootstrap, actual
Docker HTTP ownership, scheduled-service handoff or this complete protocol.
The [reviewed mutation map](../planning/deployment-foundation/interrupted-transition-mutation-map.md)
is the minimum coverage inventory. Every production call must have an exact plan;
an unknown call refuses before dispatch, but an empty/unusable registry fails
conformance rather than qualifying as a safe implementation.
The [adapter companion](deployment-native-supervisor-adapters.md) now supplies
the proposed finite plan expansion, request/body/terminal policies and private
IPC. The corrected core/adapter union is independently technically approved
througha836e92b; the owning log retains exact ranges and original negative verdicts.
The earlier two-tool [delivery companion](deployment-tool-artifact-inventory.md)
and linkage had separate independent technical approval through1648e792. The
current direct upstream-Compose amendment has its own complete independent
technical/research approval765aae44..70acee7f, including both input-construction
arms and the image-environment correction. Prior verdicts are not extended to
new bytes. P01 projection/restart semantics and its isolated rule fixtures have
their own complete independent approval407b641d..dcc151f2. Parent-contract linkage
and the SP8 adoption package below have separate complete technical approval
for dcc151f2..ed9a8f2a. Older verdicts are not extended to those new bytes.
Technical approval precedes owner adoption under SP8; native producer
construction and actual conformance follow adoption, not as a circular prerequisite
to permission to implement. OP9 remains open until its actual acceptance proof.

## SP1 — Packaging, privileges and trust

Use the release-owned `deployment-release` binary in a **one-shot native Linux
supervisor mode**, launched by the existing operator account. It is a separate
session/process, with no parent-death signal and no terminal-owned standard IO.
It exits after completion/handoff or remains the live lock owner while manual
recovery is unresolved. It is not an installed service, timer, socket-activation
unit, root daemon or boot-time agent. No additional resident guardian is hidden
inside another service. The existing scheduled-backup service remains resident.

The operator already needs access to the local Docker socket, its existing
private state, bundles and configured secrets. This profile grants no new host
capability, sudo requirement or host PID namespace to workers. It requires native
Linux/amd64, a direct local rootful Docker Unix endpoint, `/proc`, usable pidfds,
SCM_RIGHTS/SCM_CREDENTIALS and the pinned Engine/Compose implementations. Rootless,
remote TCP/SSH endpoints, socket proxies and Desktop are not this native profile.
These are **proposed explicit prerequisites**, not newly established support
claims. Adoption must reconcile DP1's host floor; R-006 records the actual kernel,
Engine/Compose versions and successful feature checks. Do not infer support from
a version number or the Desktop model alone.

The selected source/target bundle supplies the actual supervisor, backup and
operations helpers, profile and plan table, plus one pinned upstream Compose
executable for direct isolated configuration compilation. No separate CLI or
plugin discovery supplies this worker. The proposed
[inventory/delivery contract](deployment-tool-artifact-inventory.md) supplies their
separate DP1 artifact, license/source payloads, provenance and SPDX inventory.
Its technical review and owner policy/artifact adoption precede implementation;
actual pins/rights must be resolved before shipping eligibility. Fixture downloads
are not approved shipping provenance. Reuse the selected bundle's
Postgres image as the worker filesystem; do not add a seventh image implicitly.
Supervisor/producer helpers and their Docker/Compose tools are read-only mounts
from ONE selected producer manifest under OP2. This restriction applies to that
producer toolchain, not all runtime images. Retained source services and source
pre-upgrade helpers must match the SOURCE identity; target services and restore/
verification helpers match the applicable TARGET identity. Each helper role
declares which exact source/target manifest supplies it; the selected supervisor
must support both declared profiles/formats. Cross-version incompatibility
refuses before effects. Never mix arbitrary uncited helpers/tools, and never
require a running source image to match an unrelated target image.

Native operator/supervisor and approved service/observer code are trusted. Worker
input, responses, names and labels are not authority. Docker access is already
host-powerful; this proposal does not claim protection against a malicious host
operator or daemon. Secrets remain only in existing authorized memory/files;
never serialize secret values, hashes of secret configuration or protocol keys.

## SP2 — Bootstrap, lock and daemon generation

Before Started, the launcher may only validate bundles/history, acquire the
existing real operator lock, create private operator-state metadata/IPC, and
start the selected inert supervisor executable. The supervisor may then perform
the bounded admission observations/input validation specified below. Neither may
load/pull images, create/exec Docker jobs, query Postgres, contact the application,
publish metrics or modify any live service/output before admission. Execute an already opened,
verified regular helper file; a pathname hash followed by an unprotected reopen
is insufficient. The executable descriptor must survive the exec boundary.

The launcher establishes a direct native daemon connection and obtains fresh
responses with SO_PASSCRED. SO_PEERCRED alone can identify a listener creator,
not its current writer. Open a pidfd for the actual response writer, then obtain
a second fresh response from that same writer while the held pidfd is live.
Validate the protected local endpoint, expected daemon UID, supported API/version
and host boot. Reject proxy/ambiguous writers and unrecognized socket activation.
Actual Docker HTTP framing/credentials must pass conformance; the line-protocol
model is not sufficient. Persistent Engine ID is recorded only as metadata.

Pass the ORIGINAL operator-lock file, original connected socket and writer pidfd,
along with the exact selected executable/validated inputs, over inherited private
IPC. Validate descriptor kinds and native identities, matching source/target,
prefix, profile and producer; no independently reacquired lock or PID reopen.
The supervisor acknowledges only after owning the native handles and verifying
the entire tuple. The launcher then closes its copies. An interrupted/incomplete
handoff cannot mint a descriptor or enable protected work. The supervisor owns
the ledger writer and lock through all operation publishers, closure and final
append. SP6 replaces the current outer release/install metric with a retained
service's ledger projection; it cannot be moved before Finished as a shortcut.
The foreground CLI is thereafter only a status/control client; its disappearance
does not become proof that the actual worker died or trigger automatic rollback.

`engine_binding` is an OP1 object with exact keys `schema_version:1`,
`artifact_type:"native_engine"`, `host_boot_id`, `writer` (SP4 Process),
`socket_cookie` Counter, `engine_id` ID, `api_version` ASCII version and
`observed_at` Date. Its hash binds OP2. It describes a retained LIVE native handle;
reading these bytes cannot reconstruct one. New upstream connections require
fresh native writer checks against that original held generation; no reconnect
or pool retry may silently cross authority loss. Existing requests stay unknown
until actually observed terminal. Loss of supervisor/daemon/gateway authority
before sealing invokes CR6 refusal, not a reconstructed success or new window.

Pre-admission observation is a native supervisor protocol, NOT an OP3 execution
log/phase before Started. Under the real lock and held daemon generation, allow
only read-only Engine version/info, container list/inspect/declared-public-file
archive, network list/inspect, volume list/inspect and image inspect. The exact
resources derive from the accepted ledger/source/target manifests. CR2's full
mounted-byte/current-container proof must complete here for ordinary release;
rollback uses CR4's damaged-scope ownership observation, not healthy CR2. Initial
install performs the parent's read-only clean-state check, including prior OP7
descriptors even if history is empty. No exec, start, pull, DB/HTTP probe or
arbitrary archive path is an observation. Every original response and read-only
handler must be terminal before Started; timeout or transport loss refuses.
Actual archive-handler cleanup/HTTP terminal classification remains part of
the exact adapter construction gate, not an inferred response-EOF guarantee.

Rollback also authenticates its exact protected ciphertext/manifest/key BEFORE
Started using the selected trusted Go helper IN PROCESS: current
VerifyRestoreInputs/ExtractPackage semantics, existing private scratch placement
and no subprocess/SQL/Engine job. This is input validation, not a restore. Track
the original private scratch handles and finish decryption/verification/removal
before admitting rollback. Never hash/persist plaintext, keys or secret config.
Scratch/error cleanup failure blocks admission. This does not adopt D-019's
separate placement/protection policy or add a new host plaintext destination.
The current VerifyRestoreInputs ignores its context; it does not already prove
the proposed guard. Specify cancellable/guarded processing in the adapter table.
A fired guard invalidates admission and keeps the original supervisor/lock until
all native/in-process work actually ends; it cannot abandon a still-writing
goroutine and call cleanup complete.

Before each such activity sync a private pending marker containing its bootstrap
UUID, process/scope and closed activity enum; complete one final OP1
`native_admission` object only after ALL original activities/scratch are terminal.
Its exact keys are `schema_version:1`, `artifact_type:"native_admission"`,
`bootstrap_id` UUIDv4, `server_id`, `producer_manifest_sha256`, `profile_sha256`,
`anchor_prefix_bytes`, `anchor_prefix_sha256`, `engine_binding_sha256`,
`process` Process, `mode` clean_install/ordinary_release/rollback,
`observation_sha256` Hash, `input_validation` not_required/verified,
`scratch` none/removed, `activities` Hash array in native activity order,
`observed_at` Date. Observation is the complete CR2 proof,
or exact read-only install/CR4 ownership inventory with the same prefix/tuple;
the companion A0 defines its exact inventory/activity shapes (review pending). Rollback requires verified;
other modes not_required. No other fields/nulls. This object is private bootstrap
evidence, not a new ledger or permission by itself. SP7 binds its hash; it does
not reference that later scope hash, avoiding a self-hash cycle. Unresolved
markers prevent a clean-install/new-operation bypass. Lost native authority
cannot reconstruct a successful bootstrap from marker/PID/cache bytes.

After OP7 descriptor and initial install work admission (or release Started),
the NATIVE supervisor can issue the fixed admitted image-load/pull/inspect plans
directly through its own Engine transport to acquire the exact worker image.
It does not need a Docker worker to create that worker's filesystem. Bind full
response/config identity before creating any worker. No pull occurs before
the relevant admission, and no cached-image assumption qualifies as this proof.

The supervisor binds private control IPC only inside the operator's0700 state.
Attachment checks native peer credentials, held process generation, selected
executable, one fresh challenge and exact descriptor/ledger tuple. Caller-supplied
PID, socket path or operation UUID alone cannot select a lock owner. Manual
abort/rollback goes to the real live owner; never kill it to obtain the lock.
After loss, stale IPC/PID records do not authorize a replacement supervisor.

## SP3 — Inescapable worker and service boundaries

Each command worker is a distinct admitted Docker job: private PID/IPC namespace,
read-only root, no restart policy, no-new-privileges, all capabilities dropped,
no host socket, ports, host PID/cgroup namespace or writable host outputs. It
receives read-only selected inputs/authorized secret
files, private scratch and only its declared gateway network. The preflight
Compose compiler has NetworkMode=none and NO Engine channel or secret-value
descriptors; C01 fixes its public source files, nonsecret interpolation inputs
and private output handling. All Engine calls, including ping/version, belong
to the native supervisor under the companion table. No
opaque Compose up/down/run command supplies hidden mutation authority. No Docker context,
plugin, credential helper, environment override or mount may supply another
daemon or executable. Reject privileged/devices/extra mounts/network connections,
Docker exec into retained services and daemon API hijacks outside fixed plans.

Use per-job internal network(s) for worker-to-gateway traffic. A worker must not
join the application's, database's, operations' or edge network directly. The
trusted per-job opaque SQL relay alone bridges its worker network and exact DB
generation. It admits every connection BEFORE upstream dial/client forwarding,
binds the native backend before protocol bytes, and retains original stream and
forwarders plus backend pidfd through terminal observation. It does not terminate
TLS or parse keys. Preserve configured hostname verification/client certificates;
worker-side alias/dial routing must not replace the authenticated DB hostname.
No unrestricted CONNECT, arbitrary upstream or unregistered reconnect is allowed.
Actual pgx AND libpq paths, startup and verified TLS must pass this routing proof.

The SQL native observer uses the exact owned Postgres PID namespace, PostgreSQL
UID and no capabilities/no-new-privileges. It opens the backend pidfd BEFORE
validating full accepted tuple/socket inode/UID/direct postmaster parent; rejects
the postmaster, ambiguity, reuse and unrelated connections; signals ONLY a held
validated reference. Same-namespace access belongs only to trusted observer
code, never the command worker. A Python fixture is not the packaged Go producer.

Every worker scratch volume is a unique recorded generation. Backup ciphertext
is staged there; only a trusted publisher may copy the exact admitted encrypted
object into the final backup target, validate it and sync publication. Metrics
are generated/validated by the trusted publisher, not copied from arbitrary
worker bytes. Plaintext dump/credentials remain private scratch, never a host
backup/metric mount or recorded digest. Scratch deletion waits for all writers
and streams to end. No unmeasured tmpfs-size assumption supplies capacity proof.

Retained services are a DIFFERENT trusted scope. Their ordinary gameplay,
monitoring and scheduled backup activity can continue only under explicit source/
target ownership. Before replacing a service, close its operation-owned ingress
and prove its entire old scope and dependent effects terminal. For ambiguous
application HTTP after checkpoint, terminate the exact owned app/DB scopes under
CR4 before recovery; before checkpoint no mutating app request is permitted.
Pre-backup abort preserves live game, Caddy, Postgres and their data.

## SP4 — Typed subjects and native observations

Use OP1 canonical objects/grammar. These are CLOSED exact-field types; nullable
keys are explicit. Counters are strings. Lists below are ordered as stated,
unique, and complete. IDs/hashes derive from actual inspected state plus prior
admission, never labels alone.

- `Process`: `host_boot_id` UUID, `pid` Counter, `start_ticks` Counter,
  `pid_namespace_inode` Counter, `uid` Counter. Positive process/start/namespace;
  UID may be0. It is observation metadata tied to a retained pidfd, not permission
  to reopen/signal that numeric PID later. The PID is interpreted in the named
  observer PID namespace; never compare integer PIDs across namespace views.
  Live kernel credentials/held handles and actual container init mapping, not a
  self-reported PID, bind cross-namespace producer identity.
- `Container`: `container_id` CID, `image_config_sha256` Hash, `role` declared
  service/job enum, `generation` UUIDv4, `declared_networks` CID array sorted,
  `mounts` Hash array sorted. This immutable CREATED-container binding has no
  mutable PID/start fields. Inspect/validate actual creation config before binding;
  declared networks are allowed endpoints, not a claim all are already attached.
- `Start`: `container_sha256` Hash, `start_work_id` UUIDv4, `started_at` Date,
  `process` Process, `networks` CID array sorted, `native_scope_sha256` Hash.
  This is a distinct native start observation, not a replacement Container.
  Reserve start/native scope work BEFORE starting; append its once-only binding
  on actual native return, then use the running resource. Every subsequent start
  has a separate identity/observation; do not mutate an earlier subject or reopen
  terminal work. Startup child effects need prior admission under the already
  bound created-container scope; unknown start outcomes remain outstanding.
- `Mount`: `scope` a Storage binding, the inspector-only CreatedVolume reference
  defined in adapter V00, or a read-only input binding, `target`
  Relative relative to the container root, `access` read_only/read_write. Input
  binding has exact keys `class` release_artifact/secret_input/host_observation,
  `role` declared logical input ID, `device`, `inode`, `owner_uid` Counter,
  `artifact_sha256` nullable Hash (required only for release_artifact). The
  actual opened source and mount mode must match; no secret value/digest/path.
  Each Container.mounts entry hashes one OP1 Mount object, in hash order.
- `Storage`: `family` one of five persistent DP1 volume families, `job_scratch`,
  `backup_target`, `metrics_target`, `operator_state`, `service_cache`,
  `protected_staging`; `generation` UUIDv4, `driver` enum `local_volume`/`host_directory`,
  `native_name` nullable ID, `device` Counter, `inode` Counter, `owner_uid` Counter.
  `engine_created_at` nullable Date is also required; non-null only for a Docker
  volume. Native name required only for a Docker volume. Device/inode observed through
  the approved trusted mount inspector. No absolute path is persisted.
  owner_uid is the observed state at binding, not permission for an unrecorded
  ownership change. V00 defines first inspection, private-scratch initialization
  and exact approved-image population, including Postgres initialization and the
  monitoring images' ownership. Its separate current storage observation must
  validate each permitted transition without rewriting the original binding.
  Other changes refuse. A provisional CreatedVolume is never a complete Storage
  binding or permission for a service/backup/restore to consume the volume.
- `Network`: `network_id` CID, `generation` UUIDv4, `role` one of `edge`,
  `application`, `database`, `operations`, `job_gateway`, `internal` boolean.
- `Socket`: `connection_id` UUIDv4, `database_container_sha256` Hash,
  `relay_container_sha256` Hash, `backend` Process, `socket_inode` Counter,
  `local_address`, `remote_address` canonical internal IP strings,
  `local_port`, `remote_port` Counter in1..65535. No player/external receiver IP,
  URL, authentication bytes, cancellation key or secret-config hash.

Each subject has the common exact keys `schema_version:1`, `artifact_type:
"native_subject"`, `operation_id`, `phase`, `work_id`, `kind`, `binding`.
`binding` is exactly the applicable closed object:

| Kind | Binding keys |
|---|---|
| engine_request | `engine_binding_sha256`, `request_id` UUIDv4, `action` plan action enum, `resource_sha256` nullable Hash, `reserved_children` UUID array in admission order |
| job_namespace | `container` Container, `scratch` Storage array in declaration order |
| runtime_scope | `owner` source/target, `identity` CR1 Identity, `containers` Container array in service order, `networks` Network array in DP1 order, `storage` Storage array in DP1/host-scope order |
| sql_stream | `socket` Socket, `runtime_origin_sha256` nullable Hash (null for a newly dialed operation stream; original service-round admission/handoff proof for imported streams under A6) |
| http_stream | `request_id` UUIDv4, `action` fixed health/bootstrap/state/socket/alert/readiness enum, `route` H01/H02/H03/H04/H05/alert_health/alert_metrics/alert_post, `server_scope_sha256` nullable Hash, `children` UUID array; null server scope ONLY for the configured read-only receiver-health probe |
| host_publication | `scope` Storage, `target` Relative, `publisher_work_id` UUIDv4, `format` backup_ciphertext/operation_metric/service_cache |
| host_deletion | `scope` Storage, `target` Relative, `target_sha256` Hash, `policy_sha256` Hash, `publisher_work_id` UUIDv4 |
| native_job | `role` supervisor-dispatch/observer/publisher/restore-validation enum, `container_sha256` nullable Hash, `process` Process, `runtime_origin_sha256` nullable Hash, `invocation` nullable object with `invocation_id` UUIDv4 and `entry` stream_end/scratch_remove/freeze_scheduler/export_scheduler/start_services/go_snapshot/opaque_forwarder/private_scratch_cleanup/publish_cache/publish_metric/publish_ciphertext/delete_retained_object, `children` UUID array |

Container.role is exactly one of the seven service names from the Compose
template, or `engine_cli`, `backup_helper`, `sql_relay`, `native_observer`,
`publisher`, `mount_inspector`, `restore_validator`. The once-bound Container is
never replaced by its Start. Native scope/exit/removal evidence checks the exact
admitted start(s), no unseen restart, and the whole namespace. Retained ordinary
service restart policy and fresh current observation must be explicitly handled
by the exact plan table; they are not permission to restart a closed job.
Network.internal matches the approved template (edge=false,
application/database/operations=true); job gateways must be internal. Shared
physical storage/networks can appear as references in
several service bindings but have ONE exact generation/ownership disposition.
The union of retained bindings must match the full inventory; partitions cannot
omit a resource or assign contradictory owners to it.

`native_scope` (hashed by job_namespace evidence) has exact keys
`schema_version:1`, `artifact_type:"native_job_scope"`, `container_sha256`,
`host_boot_id`, `mode` private_pid_namespace/trusted_go_leaf, `namespace_device`,
`namespace_inode` Counter, `init_process` Process, `observed_at` Date.
For private_pid_namespace retain the actual private PID-namespace
FD and ORIGINAL namespace-init pidfd, bound before releasing the fixed shim.
Kernel namespace type must be PID, distinct from the host, and the exact native
init must be PID1 inside that namespace. Require original init pidfd EXIT plus
exact CID removal/restart prevention and every external Engine/SQL/HTTP/file
child terminal. Linux kills all members when this namespace init terminates and
cannot admit new processes there; this is not a generic parent/ExecID exit claim.
Prove worker cannot enter ancestor namespaces or obtain a host/Docker escape.
Actual native-host descendant/blocked-IO controls remain mandatory. This replaces
the earlier unproved cgroup-v2 populated=0 requirement with a specific kernel
namespace lifetime contract; it does not waive whole-scope terminality. Shared
Postgres-PID-namespace observers/retained backup are NOT this private job scope;
A6 defines their distinct trusted native origins and closure obligations. A temporary native_observer alone may use trusted_go_leaf in the exact
Postgres PID namespace: pinned CGO-free Go entry observe_native, no exec/fork/
plugin/healthcheck or accepted Docker-exec route, one initial thread group and
only in-process tasks. Verify that entry/closed membership in source and native
conformance; require original init pidfd group exit and all external children
separately terminal. It cannot use the Postgres PID1 exit test or kill that
namespace. No arbitrary CLI/serial tool/worker may select this trusted leaf mode.
For not_started only, native_scope hash is null and the daemon observer must
prove no admitted/pending start and the exact created container disposed.

Native evidence has exactly `schema_version:1`, `artifact_type:"native_evidence"`,
`operation_id`, `phase`, `work_id`, `kind`, `subject_sha256`, `producer_sha256`,
`profile_sha256`, `observed_at`, `result`. Hashes/tuple match OP5 and the selected
producer. `result` is the corresponding CLOSED object below, not free text:

| Kind | Result keys and required interpretation |
|---|---|
| engine_request | `request_id`, `response_status` integer100..599, `response_end` complete/not_dispatched, `handler_end` ordered_reply/server_eof/readonly_hijack_eof, `writer` Process, `socket_cookie` Counter, `bound_children` UUID array. Entire original handler/response observed under held daemon; exact native status classified by plan. Every reserved child bound or proved natively uncreated. Error/EOF/404 alone is not an unknown-create disposition. |
| job_namespace | `container_sha256`, `disposition` removed/not_started, `exit_code` nullable integer0..255, `native_scope_sha256`, `children` terminal UUID array. Whole declared native scope terminal and restart prevented; not merely parent PID/ExecID exit. |
| runtime_scope | `binding_sha256`, `disposition` removed/retained_source/retained_target, `observations` Hash array in subject order. Covers ALL containers/networks/storage and actual service ownership; unknown or omitted scope refuses. |
| sql_stream | `socket_sha256`, `backend_end` pidfd_exit, `stream_end` eof/reset, `forwarder_jobs` terminal UUID array. Both original held native process exit AND original server stream/forwarders terminal; cancellation-delivery socket is not this stream. Other errors/guards remain unresolved. |
| http_stream | `request_id`, `response_status` nullable integer100..599, `stream_end` response_complete/socket_closed/server_scope_removed, `disposition` known_handler_complete/owned_scope_terminal, `children` terminal UUID array. Plan-specific actual handler/SQL completion or full owned server-scope cleanup, never client cancel alone. |
| host_publication | `scope_sha256`, `target`, `format`, `content_sha256` nullable Hash, `bytes` Counter, `publication` no_replace/replace_metric/cache_snapshot/not_published, `directory_sync` completed, `publisher_receipt_sha256`, `protected_staging_sha256` nullable Hash. Successful publication requires actual final identity/non-null content. not_published requires proven no new final object, all owned temporaries disposed or explicitly protected, no remaining publisher/IO; null content/bytes0 cannot become backup/metric success. An existing untouched foreign target is never deleted. |
| host_deletion | `scope_sha256`, `target`, `target_sha256`, `policy_sha256`, `unlink` completed, `directory_sync` completed, `publisher_receipt_sha256`. Exact generation/protection decision, confirmed removal and no remaining delete request. |
| native_job | `process`, `disposition` exited/invocation_returned, `exit_code` nullable integer0..255, `invocation_id` nullable UUIDv4, `children` complete bound UUID array. External effect children remain independently outstanding; only local contained descendants/IO must end before this native process/invocation receipt. Process exit requires exact native scope/null invocation. Record actual wait/daemon status when known (signaled Unix status as128+signal); null means unknown business outcome and can NEVER satisfy success, even when physical terminality is proved. Only the eight declared trusted supervisor entries, the I05 trusted service cache-writer leaf, or the three exact I04 imported source-routine entries with valid runtime origin, may return while their original producer lives: non-null matching invocation/null exit, actual entry return with closed child-spawn registry and every native/IO task settled. An arbitrary helper's reply/parent exit cannot use this case. |

An unbound refusal uses a separate result with EXACT keys `reason` (one of
admission_closed/stage_forbidden/subject_mismatch/unsupported_profile/native_precondition),
`dispatch` `not_dispatched`, `reserved_children`
UUID array. Subject hash is null ONLY under OP5's proven-uncreated exception.
Already forwarded requests cannot use this shape merely because a response is
missing. Null exit/status is never mapped to success. Child arrays must equal
the admitted graph, including dynamically admitted streams, not just a count.
For publisher graphs the order is publication/deletion → job_namespace →
native_job leaf; the native job exits, its namespace is removed, then the
publication observer validates final bytes/sync and references that namespace's
terminal receipt. F01/F02/F03's trusted in-supervisor writers and I05's trusted service
cache writer instead reference their leaf invocation receipt, without a namespace
that would require the whole trusted authority process to die. A publisher receipt cannot depend on its own parent publication
receipt. All receipt/child references must be acyclic; no self-hash cycle.

## SP5 — Finite stage plans and Engine mediation

The release-bound profile file has exact keys `schema_version:1`, `artifact_type:
"native_supervisor_profile"`, `name:"linux_local_operator_v1"`, `platform:
"linux/amd64"`, `producer_artifacts` (manifest-relative path/Hash objects sorted
by path), `plans` (array sorted by plan ID). Each plan has exact keys `id` ID,
`stage` fixed stage enum, `actor` supervisor/worker/service, `kind` OP4 enum,
`action` fixed action enum, `subject_type`, `native_result_type`, `children`
(declared plan-ID array), `allowed_outcomes` OP5 enum array. The profile hashes
exact OP1 bytes; it contains no own hash. References must resolve and cannot
authorize arbitrary command/path/URL interpolation. Code emits the finite table
below and exact endpoint/body adapters; adoption cannot accept a placeholder
table or invented plan merely because its JSON is valid.

The EXECUTION plan stage enum is `prepare`, `preflight`, `checkpoint`, `drain`, `start`,
`verify`, `fencing`, `restore`, `finalize`, `install`, `service_handoff`.
These are plan stages, not new ledger events: prepare/preflight/checkpoint require
release Started or OP7 install admission; drain/start/verify require backup_ready
for release, recovery_fenced for rollback, or OP7 admission for install. Fencing
requires rollback Started (or a governed CR4 pre-backup abort); restore requires
recovery_fenced. Finalize requires the applicable real terminal outcome; no plan
can satisfy an unmet stage by naming it. service_handoff requires current Started/
install admission and is constrained to freeze/observation before checkpoint,
then the governed retained handoff. No rollback reset can hide in that family.

Every plan ID is `stage.actor.action.kind` from these closed tables. `action`
is exactly one of: image_load, image_pull, image_inspect, daemon_observe,
compose_config, validate_config, validate_alert_config, inspect_database,
receiver_health, backup_create, bootstrap, state, socket, container_create,
container_start, container_stop, container_remove, container_inspect,
container_wait, container_attach, container_logs, network_create, network_inspect,
network_connect, network_disconnect, network_remove, volume_create,
volume_inspect, volume_remove, verify_identity, alert_test, restore_validate,
restore_database, scope_observe, scope_remove, stream_forward, stream_end,
publish_ciphertext, publish_metric, delete_retained_object, scratch_remove,
freeze_scheduler, export_scheduler, retain_service, dispatch_cli, observe_native,
publish_cache. The stage/caller table below restricts which actions can be
instantiated; resource/role arguments derive only from the validated manifest,
fixed service template and admitted bindings. `subject_type` is the SP4 binding
for kind and `native_result_type` its SP4 result, not caller-defined strings.

Children are permitted only by the finite effect graph: engine_request may
reserve job_namespace/runtime_scope/native_job; job_namespace may own native_job,
sql_stream/http_stream/host_publication/host_deletion children; sql_stream may
own its forwarding native_job; runtime_scope may own the declared startup/freeze/
export observations. host_publication/host_deletion may own EITHER their declared
publisher job_namespace OR the exact F01/F02/F03/I05 trusted native writer leaf,
never an arbitrary native callback. A start_services native_job may own ONLY its
A3 service-DAG engine_request and runtime_scope/startup-observation children.
The F04 scratch_remove native invocation may own ONLY exact E20/E21 requests
for its bound terminal private generation; it waits for their actual disposition.
The I03/I04 control invocation may own ONLY its I05 cache-publication parent
and exact declared import/observation children. The cache parent owns its own
publish_cache native leaf; a control request's receipt follows leaf then cache
then control completion. F01/F02/F03/I05 writer leaves have no external effect
children and join their local IO before returning. No reciprocal parent edge,
receipt cycle or arbitrary native_job child admission is allowed. V00's pending
volume-create may own ONLY the E07/E08/E11/E10/E04 requests for its exact
first-inspection job (the one narrow engine_request→engine_request edge), whose
E07 owns that inspector namespace. No recursion/new-volume child is allowed:
the inspector needs no scratch volume. That inspector mounts the CreatedVolume
observation, never its parent's still-unavailable terminal receipt. No child action may exceed its parent's stage/
resource authority. Each actual connection
is a distinct admission; this is a repeatable edge, not a fixed stream-count
assumption. A plan/outcome enum constrains the verifier; it is not evidence that
an effect occurred or ended. All plans default to no dispatch unless present in
the reviewed release-bound generated table AND their actual route/body adapter.

| Stage / existing caller | Required allowed plan families and effect constraints |
|---|---|
| Prepare | Archive load of exact validated bundle artifact; pull exact manifest digests; inspect exact image configs. Track entire original load/pull response and any native job; consume verified immutable config IDs afterwards. No arbitrary tag/image/registry credential helper. |
| Preflight | Compose config; fixed gameserver validate-config and amtool check-config jobs; backup inspect read-only SQL; configured receiver-health GET. Jobs/streams have their own admissions. No bootstrap, account mutation, service start or final output write. |
| Pre-upgrade backup | Fixed actual backup helper; all exporter/libpq streams; encrypted staging, trusted ciphertext publication and metric. Current source exact identity/header/authentication and durable checkpoint precede drain. |
| Drain | Fixed source bootstrap/state/WebSocket plans only AFTER checkpoint; source gameserver stop by exact CID, native exit/readiness/socket observations. Pending remote effects must end before fence. |
| Start / install | Exact selected Compose-model service creates/starts with full network IDs and allowed mounted generations. Reserve job/runtime children BEFORE each create. Include startup migration/epoch jobs, backup controller and all background services. No automatic recreation of an absent/unowned resource inferred from names. |
| Verify / smoke | Exact identity-inspect helper/SQL, fixed application bootstrap/state/socket and alert test job. Response and service-side effects must complete. Actual alert delivery remains the parent's separate receiver proof; aggregate metrics alone are not attribution. |
| Rollback fencing | Close old worker admission; account for original requests/jobs, scheduled service, SQL/HTTP, output/deletion. Validate exact protected backup and source/target ownership BEFORE destructive scope cleanup. Pre-backup abort has read-only/own-job cleanup plans only. |
| Rollback execution | Only after synced recovery_fenced: remove exact failed owned scopes/DB generation, start clean target DB, restore actual protected backup, inspect identity, start/verify/smoke. Preserve other declared storage as required; no name-reused volume substitution. |
| Finalization | All operation-owned helper metrics/publishers BEFORE verification_closed/operation_completion/Finished. Remove runRelease/runInstall's outer write. Entire-release/install result is projected by the retained service only from actual final ledger authority (SP6), never pre-published as success. |
| Initial-install failure | Governed OP7 attempt/log, exact admitted cleanup including volumes only under parent install contract. Empty v1 history does not bypass unfinished descriptor/logs. No invented v2 install row or legacy adoption. |

Every native Engine request is built as a fixed supported API route/action with exact
query/body validation, not a generic method/path allowlist or worker mutation proxy. Enforce manifest
image config, exact service template, generation/network/mount binding and the
current synced ledger stage in memory before dispatch. Translate only approved
worker input-root paths into pinned host bundle/secret/scratch roots; no worker
path becomes a host mount by string substitution. Bind original response to its
request, persist each returned native identity before further use, and drain the
whole response even if the associated job/caller disappears. Reserve child effects
before dispatch, including create/start/exec/network/volume/image handlers.
Unknown/unreturned creates remain outstanding. API errors can be terminal only
when the exact handler's native contract proves no later effect.

Network and volume generations never reuse a retired destructive target name.
Existing fixed Compose names must be reconciled to recorded generation bindings,
and all four networks/five persistent volume families must be covered. A delayed
delete/create cannot select a successor's reused name. Merely isolating a late
effect does NOT satisfy CR4's literal quiescence: await its actual terminal
handler/native outcome before recovery admission. The research's late201/204
observations stay counterexamples, not an adopted exception.

If the target Postgres image/config changes, Start must explicitly stop the
owned source DB AFTER checkpoint/drain and start the exact target Postgres16
image against the governed data generation, then prove identity/readiness.
The current Start method recreates six non-Postgres services; copying that call
unchanged would fail this case. Rollback still uses the exact protected dump and
clean target DB, never a live down-migration or unsupported major-version swap.
If the image/config is identical, retain/recheck the exact owned Postgres scope.
Startup migration/epoch work is operation-owned until its actual transaction/
job completes. Close its dedicated SQL streams before retaining the steady-state
service pool; do not mark an operation sql_stream retained or infer transaction
completion from a still-running service. This may require splitting the current
startup connection lifecycle; real startup and steady-state integration must
prove it. Normal gameplay DB work is then owned by the retained runtime scope.

## SP6 — Scheduled backup ownership without another resident service

Extend the EXISTING backup controller with private authenticated freeze/export/
handoff IPC and a durable service cache outside the backup target's file-scan
root. It retains PostgreSQL UID70, has no raw Docker socket and receives only its
exact Postgres PID namespace for trusted native observation. Its internal dump
children remain governed; namespace sharing does not authorize postmaster or
other-player signaling. Native backend handles/opaque stream ownership are kept
for each active round, including key-less startup. Direct target/metric writes
remain confined to trusted service code and its tracked publishers.

For cross-UID IPC, prepare an exact leaf directory beneath the operator-owned
0700 state parent, mount ONLY that leaf into the selected backup container, and
give UID70 sufficient leaf permissions to create its socket/cache. The parent
remains0700 and is not mounted; another login/container cannot traverse it.
Never grant the service the whole operator directory or a Docker socket.
Validate leaf native identity/mount before use. The host operator accesses that
leaf through its own parent; service access comes through the explicit bind
mount, not a claim its UID can traverse the host parent. Keep this leaf outside
the backup file-scanning root. Companion A6 specifies the exact leaf modes, message/FD shapes and source/cache
adapters; that proposal is not permission to expose general IPC.

Authenticate the source service against the inspected full container ID/image/
current native init generation, a retained pidfd and two fresh native challenge
replies; match actual writer credentials, not a stale socket name/UID alone.
The service likewise accepts only the private channel provisioned by the selected
supervisor and exact descriptor/ledger/source-target tuple. Transfer originals
via SCM_RIGHTS; regular metadata inputs are read-only already-open file FDs,
native stream/pidfd/lock capabilities checked by kind. Reject missing/extra FDs,
truncation, unexpected peer/generation and changed bindings before action.

Immediately after Started, BEFORE any checkpoint or service replacement, the
supervisor freezes source scheduler admission and syncs that service's paused
state. Atomically enumerate its entire active round population: SQL streams,
native children, publications and retention/stale-temporary deletions. Transfer
the original native handles/stream observations before stopping anything; admit
the exact inherited scope/cleanup under the current phase. Finish or terminate
those OWNED rounds and obtain real receipts. An empty/racy enumeration or a
stopped backup container is not proof of no remote SQL/deletion. A crashed,
unattachable or authority-lost controller refuses transition; no scan of activity
rows or reconstructed PID substitutes. Before-checkpoint abort may resume the
unchanged source controller only after preserving-prior proof and closing all
operation work; it must not restart/replace live game, Caddy or DB.
Specifically grant SOURCE retention conditionally before closing the phase;
resume only after the actual pre-checkpoint failed Finished has non-null valid
completion/preserved-prior proof, null backup authority, and the exact original
source identity. A post-checkpoint failure cannot use this path. This is a
scheduler-control change only, not live application/DB restart.

Replacement backup starts paused, with read-only source/target identity/record
capabilities and the expected operation tuple. The trusted supervisor grants
conditional retained ownership before phase closure; the service may resume
normal scheduling only after verifying the corresponding complete Finished/
completion graph in the ORIGINAL ledger and its exact target identity; initial
installation instead requires the successful v1 row and complete OP7 InstallHandoff,
never a nonexistent v2 install event. Pass
read-only already-open metadata descriptors over the private channel, not a
claim that an open directory bypasses UID permissions. Persist the verified
service-cache snapshot/paused state using trusted service code; wrong tuple,
torn cache, missing proof or restart before handoff remains paused. Cache holds
only public evidence/identity, never credentials. It is part of retained storage,
protected from ungoverned deletion; it is not another deployment ledger.

Read-only status/proof-descriptor transfer is not new worker/effect admission and
does not reopen a sealed work log. After verification_closed the supervisor may
only publish its closure/OP7 install handoff, append its referring ledger row, serve such read-only
metadata and close handles/exit. It cannot publish an extra metric, resume an
operation worker or issue a new effect to make handoff look complete. Service
cache/scheduling is the explicitly transferred retained service's work, never
an old operation-only publisher relabelled as retained. Conformance must execute
both successful handoff AND restart/failure at every freeze/grant/Finished edge.

This changes schedule control, not the six-hour cadence, retention tiers,
seven-day checkpoint protection or encryption. RP-505's protection-metadata
amendment remains a separate adoption/implementation prerequisite; this profile
does not give unsigned outer metadata more authority or introduce a key.

The existing backup service also owns a serialized **release-result projector**.
It publishes the existing release last-result/timestamp metrics only from a
new verified final install/release/rollback ledger decision; failure/administrative
rows keep their exact CR3 semantics and cannot mint success. A pending, torn or
unverified row never advances the metric. Guard against stale-prefix overwrite:
the one publisher validates a monotonic accepted ledger prefix immediately
before publication, and an old service's entire publisher scope must be terminal
before its replacement receives ownership. No success before Finished and no
operation-owned writer after it. Ordinary backup/restore helper metrics remain
scoped to those actual helper outcomes, not whole-release completion.

This explicitly changes the current runRelease/runInstall reporting rule:
unadmitted CLI validation/lock failures produce diagnostics, not an unauthorised
shared outcome write. A missing/unavailable projector leaves stale/absent
observations and must fire the operations-floor check, never synthesize fresh
success. The companion P01 fixes the diagnostic meanings, atomic business/diagnostic snapshots,
exact proposed alert and restart matrix; its isolated pinned-rule checks precede
adoption. Reconcile DP7/runbooks/alerts as proposed there. Actual delayed/missing
projector and reboot/cache producers are constructed after adoption under SP8
and must pass the complete native acceptance population before feature closeout.
This resolves the earlier draft's circular demand for unbuilt native-cache proof
before permission to construct it; it waives no accepted release criterion.
The source cache records freeze before
protected effects; after restart it may resume only a completely verified prior
retained decision with no durable freeze. A frozen/torn/unhanded cache remains
paused. New target service obtains the actual Finished graph through read-only
descriptor transfer and syncs its cache as retained-service work. Neither a cache
nor metric becomes deployment authority. No new resident service is introduced.

## SP7 — Complete retained inventory and cleanup

The scope binding referenced by OP2 contains exact keys `schema_version:1`,
`artifact_type:"native_supervisor_scope"`, `operation_id`, `scope_id` ID,
`scope_started_at` Date, `process` Process,
`producer_sha256`, `control_generation` UUIDv4, `lock_device`, `lock_inode`
Counter, `engine_binding_sha256`, `admission_sha256`, `observed_at` Date. A live producer checks the
actual retained lock file/handles and private control generation. Serialized
matching metadata alone is insufficient. The supervisor itself is OP2 authority,
not a job that must die before it can write its own final closure.
Scope ID/start match OP2 exactly and are immutable from the original native
bootstrap; a new process cannot recreate that scope from its serialized fields.

Retained runtime inventory includes seven services in template order; four
networks; caddy_data, caddy_config, postgres_data, prometheus_data and
alertmanager_data; declared backup and metric targets; operator/service-cache
state and every encrypted-only protected staging generation; and every read-only release/config/secret mount's allowed logical scope.
Do not hash secret contents. Inventory read-only inputs separately from writable
storage. Subject generation, actual mount mode/source and release-owned bytes
must match. Missing/unowned/extra services, mounts, pending jobs or resources
refuse rather than disappear from the proof. Clean-host installs create all
bindings through OP7; unsupervised old installations need their separate route.

Close worker admission, stop the WHOLE exact worker scopes, dispose every
admitted Engine/SQL/HTTP/native/file effect, observe verified retained service/
storage ownership, then seal the full phase graph. Never synthesize empty logs
from absent files. For rollback, seal fencing then recovery_fenced BEFORE opening
execution. Successful finalization includes all operation-owned publishers
before completion/Finished; the release-result projector is explicitly retained
service activity under SP6. Sealed old receipts remain historical; current
admission rechecks observable runtime facts under the original proof anchor.

Rotation and the rehearsal-only recovery commands remain separate entry points.
This profile must either route them through separately adopted plans with their
own governed authority or reject concurrent execution under the shared lock.
They cannot run a raw Docker/SQL/output bypass while a supervised operation owns
the node. This draft does not silently enlarge CR3's action enum or claim those
commands have passed the install/release/rollback proof.

## SP8 — Adoption, construction and acceptance

Technical review must inspect the completed profile, proposed exact adapter
companion and tool inventory/delivery companion together against the actual call map,
OP1–OP8 and CR1–CR6, including cycles in bootstrap/finalization/service handoff.
Then obtain owner adoption of the architecture/explicit host and artifact floor,
with parent-body reconciliation. Technical approval alone authorizes no runtime
change. Implement profile validation, native producers, all stage adapters and
service control together as reviewable dependent batches under the adopted
contract; keep existing v1 history append-only.

### Exact proposed adoption and parent-reconciliation scope

The proposed control design is CR1–CR6, OP1–OP9, SP1–SP8 and this profile's
adapter companion (including C01/P01). Adoption authorizes construction of that
substantial native deployment-control subsystem; it is not a claim the subsystem
exists. The existing tool-packaging/terms-policy question is separate and remains
unanswered until Marco answers it. Selecting native control does not silently
adopt the tool-only MPL policy, backup signing metadata, a shipping artifact or
another pending proposal.

The adopting/delegated normative edit must reconcile these parent subjects in
one reviewed range; this table is the concrete proposed change, not permission
to edit the accepted owner-authored body before adoption:

| Parent subject | Proposed reconciliation |
|---|---|
| DP1 supported-host/tool floor | SP1's native Linux/amd64, direct local rootful Docker Unix endpoint, `/proc`, usable pidfds and SCM_RIGHTS/SCM_CREDENTIALS become explicit prerequisites for this profile. Desktop/rootless/remote/proxied endpoints are not supported by it. Use the one-shot operator-launched release binary, existing backup service and existing Postgres image; no new installed daemon, seventh runtime image or cloud account. The bundled standalone compiler/tool inventory still requires its separate adopted contract; actual pins/rights gate shipping eligibility. |
| DP5 current release and recovery | CR1–CR6/OP/SP replace caller-supplied current identity and terminal-only completion with the one existing ledger's governed identity, durable operation events, actual work ownership and typed full closure. Keep manual release/rollback, strict candidate/migration ordering, seven-day rollback and no live Down migrations. Preserve v1 bytes. Before-checkpoint abort must preserve live services/data; post-checkpoint recovery fences before restore. Lost authority refuses in-place successor mutation; no automatic rollback, host/daemon reboot permission or invented disaster-recovery route. |
| DP6 backup scheduler/handoff | Preserve six-hour backup cadence, existing retention tiers, protected-checkpoint policy, encryption and original RPO/RTO. Add SP6/A6's paused startup, durable active-round/freeze markers, exact stream/host-IO ownership transfer and guarded restart. The15s result-observation loop is not a new backup cadence. Protected staged ciphertext remains protected. RP-505's signing/protection contract and unresolved retention subjects remain separate; native adoption introduces no new key or deletion permission. |
| DP7 results/alerts/logs | P01's retained projector owns event-backed release-result metrics; only diagnostic gauges belong to the admitted supervisor's status loop. Add its exact availability alert alongside all seven existing families, with real firing/resolution required. Apply C01's narrowly scoped unlogged private compiler output exception; retain structured journald/service logs and14-day evidence, optional raw-IP7-day policy and existing pressure alert. No metric or cache becomes release authority. |
| DP8/AC10 and full release floor | Preserve current CI topology/sub-five-minute target, no deployment secrets/workflow deployment, all existing acceptance thresholds and exact-manifest supported-host R-006. This contract is not current CI green. Local unit/rule/prototype evidence cannot replace real helper/backup/receiver/private-output/native/clean-host proof or any nine-tier1.0 criterion. |

No new permanent OS service is selected: the one-shot supervisor and retained
backup service roles above preserve the no-additional-resident-service promise.
The host/platform prerequisites and changed scheduler/metrics behavior remain
explicit proposed changes requiring Marco's adoption. The accepted parent is
unchanged until that decision and the applicable body-reconciliation delegation.
Construction may proceed only on adopted contract dependencies; acceptance still
requires every actual producer/consumer below, not a linked spec or static table.

Acceptance is an ACTUAL packaged workflow: clean supervised install and all v1
attempts → exact-current ordinary release → post-checkpoint failure/rollback,
plus preserving pre-checkpoint abort. Use the real helpers, pinned standalone Compose,
Engine, verified-TLS Postgres, application bootstrap/WebSocket, encrypted backup,
publishers, scheduler and all persistent resources. Prove original bystander
data/services preserved where required. Compare restored canonical identity/data,
not merely container/receipt counts. Native Linux/amd64 and exact-manifest R-006
with original RPO/RTO/space/backup/alert thresholds are still mandatory.

Required discriminating populations include caller and worker death; full
descendant escape attempts; actual daemon authority loss/socket activation;
incomplete/replayed descriptor transfer; log sync failure before each effect;
unreturned create/start/delete; original startup/blocked/TLS SQL; canceled
bootstrap/socket and outstanding server SQL; scheduled round/retention freeze,
controller death and restart at every handoff edge; delayed metric/ciphertext/
delete opened after proposed fence; all four networks and five volume families;
partial/mixed recovery; protected-backup tampering; lost/torn evidence/cache;
and swapped/missing/extra native graph objects/receipt kinds. Include C01's
same-source/environment/path comparison, missing binary with seeded fallbacks,
attempted worker Engine/secret grants, and unlogged complete private output
including attach races, truncation and child failure. The public configuration
research cannot satisfy that privacy/lifetime population. New critical
oracles need a real negative or targeted severing. Unknown/guard exhaustion must
invalidate the population and must not qualify as terminal refusal.

Retain existing weak-control results. A profile that only refuses, uses fixture
boolean receipts or passes on Desktop alone does not meet acceptance. No
optimization, CI topology change, publication, deployment, archival or release
approval follows from this proposal.
