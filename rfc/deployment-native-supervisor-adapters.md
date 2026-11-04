# Native supervisor: request and producer adapters

Normative companion proposal to [SP1–SP8](deployment-native-supervisor-profile.md).
**Draft; direct upstream-Compose amendment pending technical review/adoption;
no implementation authority.** This enumerates the proposed
plan population and adapters. Native conformance remains unproved. A successful
table/schema check cannot establish a real handler, filesystem or service result.

## A0 — Admission-only objects and native protocol

These activities precede Started under SP2; they are NOT OP3 phases or permission
to create a job. Exact activity enum: daemon_observe, container_list,
container_inspect, container_archive, network_list, network_inspect, volume_list, volume_inspect,
image_inspect, service_status, validate_restore_input. Use only E00/E03/E04/E05/E06/E15/E20 below
and in-process protected-input validation. Clean install/rollback volume/network
enumeration uses GET /volumes and GET /networks with no filters/body; entire200
result is locally restricted/validated against accepted bindings. No other
pre-admission Engine path is allowed. Every original read uses the same native
handler classifier below; no exec, SQL, application HTTP or image acquisition.
I00 below additionally permits read-only native status/ledger-FD renewal for an
already retained exact service; no freeze/grant/cache mutation is a bootstrap
command. This is metadata attachment under existing ownership, not new work.

Pending marker has exact keys `schema_version:1`, `artifact_type:
"native_admission_pending"`, `bootstrap_id`, `sequence` Counter, `activity`,
`process` Process, `engine_binding_sha256`, `anchor_prefix_bytes`,
`anchor_prefix_sha256`, `observed_at`. Each is an immutable0600 OP1 object synced
BEFORE activity, under the0700 state. Partial/fired activity remains unresolved.
Successful activity result has exact keys `schema_version:1`, `artifact_type:
"native_admission_result"`, `pending_sha256`, `projection_sha256`, `observed_at`.
Projection is only validated public native metadata/CR2 file observation or the
validation object below, never raw inspect/protocol/config data. native_admission
adds required `activities` Hash array in sequence order, referencing ALL complete
results; readers verify markers, exact prefix/native process and no omitted
pending activity. Matching a JSON hash is not physical completion evidence.

For clean_install/rollback, `observation_sha256` refers to a closed object with
exact keys `schema_version:1`, `artifact_type:"native_admission_inventory"`,
`bootstrap_id`, `server_id`, `mode`, `source` nullable Identity, `target` Identity,
`ledger_prefix_bytes`, `ledger_prefix_sha256`, `services`, `networks`, `storage`,
`jobs`, `pending_work`, `observed_at`. service entries are exact `{service,
container, current_start, owner}` in seven-service order; container/start may be
null only for observed absence/never-started, owner source/target/absent/unowned.
Network entries are exact `{binding, owner}` in SP7 inventory order. Storage
entries are exact `{binding, owner, engine_observation_sha256}`;
binding is nullable SP4 Network/Storage only for actual absence. Jobs are exact
`{container, current_start, owning_operation_id}` in CID order. pending_work is
the complete ancestry-ordered UUID array from accepted open work logs, not a
census. Every ownership mapping must match accepted intent/binding ancestry and
actual native generation. An unowned related resource refuses; partial/missing
OWNED failed resources remain observable and do not require healthy CR2.
Clean install requires source=null and all project resources/jobs/pending work
absent, including earlier unfinished OP7 attempts. Ordinary release instead
uses the existing complete CR2 proof unchanged; no second healthy-current type.
The configured empty host targets and inert bootstrap operator metadata are not
live project resources; their presence is validated, not mistaken for a failed
clean-install check. A prior unfinished attempt is never such an exemption.

Native pre-admission volume ownership uses the original ADMITTED, non-reused
generation/physical binding plus fresh E20 driver/name/creation observation.
Do not invent a fresh device/inode stat by calling E20: it exposes no such stat.
`native_engine_volume_observation` has exact keys `schema_version:1`,
`artifact_type:"native_engine_volume_observation"`, `engine_binding_sha256`,
`native_name`, `driver` local, `engine_created_at`, `storage_binding_sha256`,
`observed_at`. The Hash is required for an existing local volume, otherwise null.
The closed prior producer/creation graph and enforced no-name-reuse rule are
essential; name/labels/two metadata samples alone cannot establish ownership.
After Started, BEFORE mounting/reset/deletion, actual trusted mount inspection
must revalidate device/inode and permitted current ownership. No new observer
job is smuggled before Started. Unsupervised/unknown/recreated generations refuse
rather than synthesizing a physical binding. Privileged out-of-protocol host
filesystem replacement is outside the operator-trust boundary, not something
this metadata observation claims to detect atomically.

Authenticated input projection has exact keys `schema_version:1`, `artifact_type:
"native_admission_input"`, `bootstrap_id`, `backup_id`, `manifest_sha256`,
`ciphertext_sha256`, `validation` authenticated, `scratch_scope` UUIDv4,
`scratch_device`, `scratch_inode` Counter, `scratch_end` removed, `observed_at`.
It follows actual existing protected-envelope/key/manifest validation and complete
private workspace cleanup, not outer metadata alone. No plaintext/key/config
hash or absolute scratch path. The current context-ignoring function needs a
guard-aware reader/lifecycle implementation; guard exhaustion is invalid even
if a later return succeeds. No unjoined goroutine can write after terminal.

## A1 — Exact plan population and authority

Use OP1 types and SP4 subjects/results. In the tables, `S` means supervisor,
`W` worker and `V` retained service. Stage lists mean their EXACT finite expansion;
they are not patterns accepting new stages. Plan ID is
`<stage>.<actor>.<action>.<kind>`, using full actor names. The following rows are
the ENTIRE population, including the lifecycle rows in A2. No Cartesian product
of all actions/kinds/stages is allowed. Every child keeps its parent's operation,
phase, resource/command authority; matching a plan ID alone never grants access.

| Stages | Actor | Action | Kind | Adapter / permitted subject |
|---|---|---|---|---|
| prepare,install | S | image_load | engine_request | E01, exact bundle game archive |
| prepare,install | S | image_pull | engine_request | E02, exact selected manifest digest |
| prepare,preflight,checkpoint,drain,start,verify,fencing,restore,finalize,install,service_handoff | S | image_inspect | engine_request | E03, source/target authorized config |
| prepare,preflight,checkpoint,drain,start,verify,fencing,restore,finalize,install,service_handoff | S | daemon_observe | engine_request | E00, original bound daemon only |
| prepare,preflight,checkpoint,drain,start,verify,fencing,restore,finalize,install,service_handoff | S | container_inspect | engine_request | E04, explicit admitted CID or current source/target inventory |
| prepare,preflight,checkpoint,drain,start,verify,fencing,restore,finalize,install,service_handoff | S | container_list | engine_request | E05, complete node project/job inventory, never caller-selected filters |
| prepare,preflight,checkpoint,drain,start,verify,fencing,restore,finalize,install,service_handoff | S | container_archive | engine_request | E06, declared public release files only |
| preflight,checkpoint,drain,start,verify,fencing,restore,finalize,install,service_handoff | S | container_create | engine_request | E07, exact reserved C-command job or R-service |
| preflight,checkpoint,drain,start,verify,fencing,restore,finalize,install,service_handoff | S | container_start | engine_request | E08, exact created CID/start reservation |
| preflight,checkpoint,drain,start,verify,fencing,restore,finalize,install,service_handoff | S | container_stop | engine_request | E09, owned temporary job; retained service only under A3 |
| preflight,checkpoint,drain,start,verify,fencing,restore,finalize,install,service_handoff | S | container_remove | engine_request | E10, exact terminal owned CID |
| preflight,checkpoint,drain,start,verify,fencing,restore,finalize,install,service_handoff | S | container_wait | engine_request | E11, admitted temporary job |
| preflight,checkpoint,start,verify,restore,install | S | container_attach | engine_request | E12, temporary helper stdout/stderr, stdin=false |
| prepare,preflight,checkpoint,drain,start,verify,fencing,restore,finalize,install,service_handoff | S | container_logs | engine_request | E13, non-following admitted job/service observation |
| preflight,checkpoint,drain,start,verify,fencing,restore,finalize,install,service_handoff | S | network_create | engine_request | E14, exact new generation |
| prepare,preflight,checkpoint,drain,start,verify,fencing,restore,finalize,install,service_handoff | S | network_inspect | engine_request | E15, exact full network ID |
| start,fencing,restore,install | S | network_connect | engine_request | E16, admitted CID/network pair |
| preflight,checkpoint,drain,start,verify,fencing,restore,finalize,install,service_handoff | S | network_disconnect | engine_request | E17, exact owned pair after ingress closed |
| preflight,checkpoint,drain,start,verify,fencing,restore,finalize,install,service_handoff | S | network_remove | engine_request | E18, exact terminal owned generation |
| preflight,checkpoint,start,verify,fencing,restore,install | S | volume_create | engine_request | E19, declared persistent or per-job scratch generation |
| prepare,preflight,checkpoint,drain,start,verify,fencing,restore,finalize,install,service_handoff | S | volume_inspect | engine_request | E20, recorded unique native name |
| preflight,checkpoint,drain,start,verify,fencing,restore,finalize,install,service_handoff | S | volume_remove | engine_request | E21, exact empty owned generation, never generic prune |
| preflight | S | receiver_health | http_stream | H01, configured read-only receiver health |
| drain,verify,install | S | bootstrap | http_stream | H02, exact source/target app and registered request |
| drain,verify,install | S | state | http_stream | H03, same probe account/session |
| drain,verify,install | S | socket | http_stream | H04, same probe session/world subscription |
| drain,start,verify,fencing,restore,install | S | readiness | http_stream | H05, exact owned app readiness only |
| verify,install | S | alert_test | http_stream | H06, fixed test through exact Alertmanager/receiver |
| preflight,checkpoint,start,verify,fencing,restore,install | S | stream_forward | sql_stream | Q01, original registered opaque relay/backend |
| service_handoff,fencing | S | stream_adopt | sql_stream | Q03, exact original service-round admission/handoff under A6 |
| service_handoff,fencing | S | native_adopt | native_job | I04, exact original retained native scope, never a new historical execution claim |
| preflight,checkpoint,start,verify,fencing,restore,install | S | stream_end | native_job | Q02, exact held backend/forwarding cleanup, never PID reopen |
| checkpoint,restore,install | S | publish_ciphertext | host_publication | F01, authenticated encrypted output in declared backup scope |
| checkpoint,restore,install | S | publish_ciphertext | native_job | F01, trusted native ciphertext-writer leaf invocation |
| prepare,preflight,checkpoint,drain,start,verify,fencing,restore,finalize,install,service_handoff | S | publish_metric | host_publication | F02, actual helper outcome or projection status only; never whole-release success |
| prepare,preflight,checkpoint,drain,start,verify,fencing,restore,finalize,install,service_handoff | S | publish_metric | native_job | F02, trusted native metric-writer leaf invocation |
| fencing,restore,install | S | delete_retained_object | host_deletion | F03, exact admitted public object + accepted protection policy |
| fencing,restore,install | S | delete_retained_object | native_job | F03, trusted native unlink/sync leaf invocation |
| preflight,checkpoint,start,verify,fencing,restore,finalize,install | S | scratch_remove | native_job | F04, exact job-private scratch, no plaintext hash |
| preflight,checkpoint,drain,start,verify,fencing,restore,finalize,install,service_handoff | S | scope_observe | runtime_scope | R01, full owned source/target inventory or partition |
| drain,start,fencing,restore,install | S | scope_remove | runtime_scope | R02, only A3-authorized exact scopes |
| start,verify,finalize,install,service_handoff | S | retain_service | runtime_scope | R03, complete retained transfer, conditions in I05 |
| service_handoff,fencing | S | freeze_scheduler | native_job | I03, exact source controller and durable freeze |
| service_handoff,fencing | S | export_scheduler | native_job | I04, complete frozen handle population |
| service_handoff,fencing,finalize,install | V | publish_cache | host_publication | I05, exact freeze/conditional retained-service cache |
| service_handoff,fencing,finalize,install | V | publish_cache | native_job | I05, trusted source/target cache-writer leaf invocation |

`readiness`, `container_list`, `container_archive`, `stream_adopt`, `native_adopt` and `initialize_storage` are added to
SP5's action enum; this finite population supersedes its earlier family-only enumeration.
Each plan's subject/result type is the corresponding SP4 kind binding/result.
Allowed outcomes are completed/refused, plus retained ONLY for R01 (continued existing ownership) and R03 runtime_scope.
Refused means actual non-dispatch/proven-no-effect, not any error status.
Completion of a failed native job is distinct from success of its parent stage.

Only Q02, F04, I03, I04, F01/F02/F03 writer leaves and the start_services orchestration below use SP4's
trusted in-supervisor invocation variant.
Allocate invocation UUID and sync admission before calling its fixed entry;
close its child-spawn capability, join all local tasks/IO and return before
emitting invocation_returned. Its external effects remain separately owned:
signal return cannot close Q01, freeze ACK cannot close active round work.
I05's trusted service cache-writer leaf and the three imported I04 source-routine
entries likewise use invocation_returned with their exact original producer/
origin, never generic worker acknowledgment. All other native_job plans require their actual process/whole scope to exit.
No generic CLI/worker success message can use the invocation-returned variant.

Service cache I05 is admitted operation work until R03's conditional transfer;
after that it is the explicitly retained service's work. No new OP3 admission is
written after verification_closed. The normal scheduled rounds/result projector
are governed retained activity, not extra execution plans for a closed operation.

The additional native plan rows are exactly
`start.supervisor.container_start.native_job`,
`install.supervisor.container_start.native_job`, and
`restore.supervisor.container_start.native_job`, with invocation entry
start_services. They execute the fixed A3 service DAG through
E07/E08 directly. Their admission precedes native start, their child-spawn registry
closes when dispatch returns, and their receipt waits for every declared child
(including startup SQL/retained-scope disposition) to close. No new worker or
child can be admitted after that native item is terminal.

**One Engine dispatcher:** workers have NO Docker API capability.
The selected Go supervisor executes every E01–E21 request through its owned
transport and fixed native body compiler. Compose remains the release model/
configuration authority, not an opaque lifecycle command hiding native requests.
The preflight config compiler directly invokes the selected Compose executable
under C01, with no Engine broker/network. E00 is supervisor-only. Any worker
Engine request or attempted Engine descriptor grant refuses before dispatch.
The H06 alert HTTP channel remains scoped HTTP, never an Engine channel.
This explicitly replaces current lifecycle CLI calls, preserves
the full Compose service/network/storage behavior and requires parent adoption.

Before Started, strict in-process loading of the FIXED release Compose model
with the pinned Compose model-loader version is inert input validation; no
plugin/process/Engine write is allowed. After admission the actual pinned Compose
config worker must match the native complete typed model in memory. Reject any
unsupported field/template extension or mismatch. Do not hash/persist the
resolved secret-bearing model. This is a concrete compiler compatibility gate,
not a new handwritten parallel configuration authority.

### C01 — Direct configuration worker

The selected ONE producer supplies `tools/docker-compose`, mounted read-only
at `/tools/docker-compose`; tools from another bundle or host never substitute.
The job has NetworkMode=none, empty networking endpoints and no gateway/Engine
socket or secret-value descriptor. Native supervisor E00 still establishes the actual
daemon identity/version. Compiler success does not replace that observation.

WorkingDir is `/inputs/release`. Model-source mounts are ONLY the exact bundle's
public `compose.yml` and, when the existing governed rotation selection requires it,
`compose.rotation.yml` at those same names, read-only and manifest-bound. Model
source may be the declared SOURCE or TARGET; compiler identity stays the selected
producer. Do not normalize/shorten/merge away source fields before handing the
original files to Compose. Source/target compatibility must pass for both actual
model populations. This does not resolve remaining rotation/schema/API holds or
change DP4's key families/minimum overlaps.
The common W00 control leaf/private scratch and selected tool mounts remain;
this does not expose an operator-state directory or other source files.

The fixed command is an argv array:

```
/tools/docker-compose --project-name cloud-clicker --env-file /run/private/compose/empty.env --file /inputs/release/compose.yml [--file /inputs/release/compose.rotation.yml] config --format=json
```

Brackets denote the one governed optional overlay, never literal/user argv.
After native child binding/release, the fixed gate prepares the public empty
dotenv file at the exact argv path and Docker config
`/run/private/compose/docker/config.json` with `{"auths":{}}`, each0600 in its
UID-owned0700 scratch, then executes this command. No operator dotenv/Docker
context/credential file or
discovery directory is read by this worker. The supervisor constructs one exact
child environment from the effective validated interpolation values and C01's
fixed execution controls. It supplies that map in Config.Env AND in the
authenticated C01 release below; the gate explicitly replaces its inherited
environment at exec rather than treating Config.Env as an inheritance barrier.
These are nonsecret settings/key IDs/file-path references, not secret contents.
Both the native loader and actual compiler use the SAME values, original file
bytes, virtual root and rotation selection. Held-root translation is checked
before comparing source/target paths; an arbitrary worker path is no authority.

The allowed interpolation names are the fixed template's PUBLIC_ORIGIN,
SERVER_ID, JWT_CURRENT_ID, BOOTSTRAP_CURRENT_ID, AGE_RECIPIENT, BACKUP_TARGET,
OPERATIONS_METRICS, HTTP_PORT and HTTPS_PORT, each prefixed CLOUD_CLICKER_;
POSTGRES_DB/POSTGRES_USER; and the six existing CLOUD_CLICKER_ file references
DATABASE_URL_SECRET_FILE, POSTGRES_PASSWORD_SECRET_FILE, JWT_CURRENT_SECRET_FILE,
BOOTSTRAP_CURRENT_SECRET_FILE, CURSOR_CURRENT_SECRET_FILE, ALERTMANAGER_CONFIG.
The governed rotation variant additionally uses CLOUD_CLICKER_JWT_PREVIOUS_ID,
CLOUD_CLICKER_BOOTSTRAP_PREVIOUS_ID, CLOUD_CLICKER_JWT_PREVIOUS_SECRET_FILE and
CLOUD_CLICKER_BOOTSTRAP_PREVIOUS_SECRET_FILE. Preserve existing defaults and
validation; corresponding accepted rotation fields must be reconciled into the
operator schema before admission. Unknown interpolation names, missing required
values or unsupported variants refuse; this list cannot waive a parent's
required rotation behavior.

Add only fixed execution controls: PATH=/run/private/compose/no-tools (no lookup),
DOCKER_CONFIG=/run/private/compose/docker (the auths-only config directory),
DOCKER_HOST=unix:///run/private/compose/no-engine.sock (must be absent),
DOCKER_CLI_HINTS=false. Encode the child environment as a sorted unique array of
`{name,value}` string objects, with no NULs, duplicates or unknown names. Omitted
optional interpolation values stay omitted so that the original template's
defaults apply identically in both loaders; image defaults never fill that gap.
The gate execs the absolute binary with exactly this environment, not an overlay
on os.Environ. COMPOSE_*, other DOCKER_*, proxy/plugin/credential helpers and caller
extras cannot change it. No interpolation value is a shell fragment. There is
no compose_input or Engine broker FD: public files arrive through fixed mounts,
values through the validated environment. Other helpers retain their actual
secret/SQL/HTTP grants. Docker's ordinary container configuration can retain
these nonsecret environment values on the trusted node; no memory-only claim
is made about that store. Actual credential values never enter Config.Env.

Docker DOES merge image environment defaults into container configuration. The
selected image's full Config.Env must therefore match the finite, reviewed
per-image producer template under SP5, including exact names/values; no duplicate,
malformed, credential-bearing or unenumerated default is accepted. That template
comes from the selected manifest-bound image config, not a worker/host override.
Compute the expected effective container environment with the pinned Engine's
merge rules (requested names override image defaults) and compare the complete
E04 result after E07, BEFORE E08/shim start. Refuse any unexpected key/value or
unsupported default; do not start a shim under unchecked image settings.
The pinned [create path](https://github.com/moby/moby/blob/v28.4.0/daemon/create.go)
and [merge implementation](https://github.com/moby/moby/blob/v28.4.0/daemon/commit.go)
establish this distinction. A matching merged container environment still does
not become the compiler environment: only the authenticated C01 map is passed
to exec. Image defaults that remain in ordinary nonsecret daemon configuration
are excluded from the compiler process. This adds no new image or generic
environment-setting capability to another role.

Private compiler output MUST use LogConfig.Type=none with empty Config.
Config.AttachStdout/AttachStderr are true; stdin and Tty remain false. Admit
E12 and establish its original live stdout/stderr attachment after shim start/
binding but BEFORE releasing the compiler command. E12 logs=0 prevents replay;
missing attachment refuses command release. Attachment readiness requires
validated101/200 headers and a live original stream reader. This readiness does
not complete E12; its work stays outstanding. Decode the pinned non-TTY Docker
stdout/stderr multiplex framing; malformed/truncated frames fail observation.
The [pinned daemon attach path](https://github.com/moby/moby/blob/v28.4.0/daemon/attach.go)
registers output readers before its stream handoff and requests stored logs only
when logs=1; that source fact does not replace actual native conformance.
Collect complete stdout privately with a bounded reader derived from admitted
model/input size and pinned output format; never truncate into a passing model.
Drain stderr privately without unbounded buffering or retained content. Require
actual zero child exit, both stream EOFs
and original native scope end, then compare the complete typed model with the
native loader. The shim/supervisor must not persist model/stderr output; no raw
model, stderr, environment or inspect payload/hash enters diagnostics, metrics,
receipts or artifacts. Only the ordinary nonsecret container configuration store
disclosed above is permitted. E13 is forbidden for this
job. The public research fixture's Docker logs are not this implementation.

Native process termination is observed independently of the later E12/namespace
receipts; E12 joins that observation with original EOF, and namespace completion
consumes all required dispositions. No receipt waits on itself. Loss/truncation,
late/cancelled attach, nonzero exit or incomplete output never counts as valid
configuration. A guard produces INVALID measurement under SP8, not successful
refusal evidence. Remaining shim/descendant/lifetime/fence rules are unchanged.

Acceptance must exercise the actual packaged job with a seeded private marker
in valid operator settings and check no marker/model/hash enters Docker logs,
retained daemon job logs, operator evidence/diagnostics or output artifacts.
This logging claim does not deny the ordinary nonsecret Config.Env store above.
Corrupt LogConfig, delay/sever attach before release, truncate model output,
swap source/target/env/path data and force child failure: all must fail the
required observation. Seed missing-binary host/PATH fallbacks and attempt worker
Engine/secret grants; no fallback or extra privilege may execute. Current public
input studies use explicit exec environment replacement, not suppression of
Engine's image merge. Native acceptance must additionally seed an unrecognized
image default and altered effective container environment (refuse before start),
and a declared benign image-only default (absent from the actual compiler
environment and unable to fill an omitted interpolation value). A release map
that differs from admitted inputs must refuse at supervisor dispatch; malformed
or extra gate fields/fixed-control changes must refuse before compiler exec.
The studies establish only their recorded compatibility facts. Native
Linux/amd64, full admitted-input conformance and SP8/R-006 stay mandatory.

## A2 — Exact command/lifecycle plans

For EVERY row below, admit the job_namespace plan
`stage.supervisor.action.job_namespace` before its E07 create and the native_job
plan `stage.worker.action.native_job` before native execution. These are the ONLY
additional plan rows. A job namespace owns its command/native child and its actual
Q01/F01/F02/F03/F04 children as permitted here; publication/deletion instead owns
its publisher namespace, whose native child is a leaf. No receipt hash cycle.

| Stage | Action | Image/helper identity; exact command semantics | Permitted effect children |
|---|---|---|---|
| preflight | compose_config | producer standalone Compose, C01 exact original model/validated environment and fixed config --format=json | no Engine/secret descriptors; complete unlogged E12 model compared to native compiler |
| preflight | validate_config | applicable source/target gameserver, validate-config | no DB/HTTP/publication |
| preflight | validate_alert_config | applicable Alertmanager, amtool check-config of authorized secret FD/path | no HTTP/publication |
| preflight | inspect_database | applicable backup helper, inspect | Q01 read-only inspection only |
| checkpoint | backup_create | SOURCE backup helper, create --pre-upgrade, SOURCE manifest/epoch | Q01 exporter/dump; staged ciphertext; F01/F02 |
| verify | verify_identity | TARGET backup helper, inspect/recovery-identity as parent's named gate | Q01, no publication |
| install | verify_identity | TARGET backup helper, same gate | Q01, no publication |
| verify | alert_test | TARGET operations helper, alert-test, configured endpoints only | H06, no arbitrary receiver/body |
| install | alert_test | TARGET operations helper, same fixed test | H06 |
| restore | restore_database | TARGET backup helper, restore exact governed backup/manifest and authorized identity | Q01 clean-target/restore/inspection, F02, F04 |
| preflight,checkpoint,start,verify,fencing,restore,finalize,install | observe_native | producer native_observer, or mount_inspector ONLY for V00; fixed scope/mount/stream observation named by parent | observations only; no unadmitted signal/mutation |
| preflight,checkpoint,start,verify,fencing,restore,finalize,install | initialize_storage | producer mount_inspector, fixed initialize_private_storage entry; UID0, private namespace, cap_drop ALL | V00 private scratch/staging root initialization only, no secrets/service data/SQL |

Each command is an argv array built by the trusted supervisor from these named
semantics and existing helper flag decoders. No shell/-c, user command, extra
Compose file/project/env/plugin/receiver or arbitrary mount is accepted. Fixed
input paths are scoped aliases resolved through held roots; actual secret values
stay in authorized file descriptors/memory and never become argv evidence.

Worker Compose inputs are restricted operation inputs, not a writable copy of
the node's original compose.yml. All required networks/volumes are supervisor-
created and externally referenced by full ID/unique generation name. Helper jobs
use private gateways/scratch and no writable host output. Live service inputs
retain their approved service template, secret mounts and declared capabilities.
Replace current helper --rm with supervisor-owned exact-CID cleanup: AutoRemove
is false, restart is no. Commands cannot auto-remove before native observation.

Each temporary job starts the pinned Go entry shim as its namespace PID1. The
shim reports native identity/handles and waits; bind the created container and
new start/native invocation BEFORE releasing its fixed command. Native PID/start
observation is not a second Container binding. Finish/kill the WHOLE scope,
collect original streams and exact native exit evidence, then remove by CID.
Ordinary persistent-service startup is separately governed by A3/R03, not this
temporary namespace's completion rule.

Transient helper secret inputs use authorized read-only file descriptors granted
only after native role/start validation. Adapt Go readers to consume those FDs
directly; reopening /proc/self/fd/N does NOT bypass source-file DAC permissions.
Each consumer receives an independent offset/read capability. Do not record
contents/hashes or use a directory FD to claim create permission. libpq credential
and TLS file inputs, where its pathname API requires them, are materialized0600
by the receiving UID in its private0700 scratch and removed after actual writers
end, preserving the exact TLS bytes/hostname policy. The same rule applies to
path-only configuration readers such as amtool; C01 Compose receives no secret
FD. Inherited file FDs are not a claim that /proc/self/fd reopening bypasses DAC.
Age identity may be consumed
from its original read FD without another persistent key copy. Staged ciphertext
is likewise granted as a read FD to the operator-owned F01 writer; writable
backup/metric targets must already be accessible to the existing operator.
No sudo/chown/permission bypass, new key or silent sslmode downgrade is assumed.

## W00 — Temporary-job control and descriptor transport

This is DISTINCT from A6's service leaf. For each admitted temporary namespace,
the supervisor creates operator-owned mode0711 `job-ipc/<generation>` under
its0700 root, binds its operator-owned0666 SEQPACKET socket before E07, and mounts
ONLY that exact leaf read-only into
that job at /run/job-control. No service-cache/state/daemon/ledger/lock directory
or handle is exposed. Connecting to the Unix socket does not require a writable
mount; unlink/create through this mount must fail. Close/unlink the host-owned
leaf only after original peers and the complete job scope are terminal. UID70/
65532 get traversal of this exact leaf, not listing/write access or traversal of
the unmounted0700 host parent. Native conformance must connect under each actual
job UID and refuse create/unlink/parent traversal; socket0666 alone is not proof.
Unknown mount or inode substitution refuses before dispatch.

Canonical messages have exact keys `schema_version:1`, `operation_id`, `phase`,
`job_work_id`, `command_work_id`, `sequence`, `challenge`, `command`, `payload`.
The reply replaces command with result and echoes the exact tuple/sequence/
challenge. Challenge is a fresh memory-only256-bit value for this one connection;
apply A6's native framing/type/truncation/deadline rules, including the deadline
from CALL ENTRY, before waiting for admission and again before acceptance. No
late/queued/cancelled response or successful send is completion evidence.

The supervisor starts each connection with command challenge, empty payload and
no FDs. The fixed Go entry connects before any command or secret read. Its hello
payload is exactly `{role, generation, scope_mode, namespace_pid, descriptors}`.
For private_pid_namespace it is the PID1 shim: namespace_pid is1 and descriptor
roles are `[self_pidfd, pid_namespace]`, those two original SCM_RIGHTS FDs in order.
For trusted_go_leaf it is ONLY observe_native in the exact Postgres namespace:
namespace_pid is its actual positive NON-1 PID; descriptors are
`[self_pidfd, pid_namespace, postgres_init_pidfd]`. The third is the original
pidfd opened for namespace PID1; in the supervisor's host view it must name the
exact Engine State.Pid/original native Start of the admitted Postgres CID.
Validate Engine PidMode=container:<that exact PG CID>, the actual own namespace
FD and pinned no-fork entry; never require this observer to BE Postgres PID1.
Other roles/modes refuse. The supervisor obtains actual kernel writer UID/PID
through SO_PASSCRED, validates expected UID/no user remapping and exact CID's
Engine State.Pid in the SAME host observer namespace. Validate the pidfd's live
native target through its fdinfo in that view and the namespace FD's native PID
namespace type/device/inode; held pidfd, current Engine init and actual writer
must be the SAME original process. Bind SP4 Process in the observer's namespace;
never compare the shim's integer1 to a host PID. Require a second fresh challenge
reply from that same held live process/connection before binding Start: command
confirm, payload exactly `{role, generation, scope_mode}`, no FDs, matching
kernel writer. The namespace_pid self-report never overrides native mapping. These
transferred self handles avoid assuming host ptrace access to another UID's
/proc/<pid>/ns or reopening a numeric PID. Unreadable/ambiguous mapping refuses.

Reply `bound` has payload `{container_sha256, start_sha256, native_scope_sha256}`,
no FDs. The shim may now fork ONLY the fixed blocked command gate under the
already admitted namespace/command item. Before release, PID1 sends child_created
with payload `{command_work_id, descriptors}` and descriptors exactly
`[command_pidfd, pid_namespace]`, those original SCM_RIGHTS FDs in that order.
Reply registered echoes the command_work_id and no FDs. Root validates original
live PID1 writer and its pinned direct-child producer contract, then retains the
child pidfd/namespace. That child sends `command_ready` on its own connection with payload `{action, descriptors}` and the same two self-handle
roles. Authenticate its actual kernel writer and original child pidfd against
the pinned shim's live direct-child registry and exact namespace, then bind the
command's native_job BEFORE it executes any A2 command. No arbitrary sibling,
reported numeric PID, user argv or caller-declared role can select this gate.
The trusted_go_leaf mode has no PID1 shim/child gate/fork: its own bound original
process receives the fixed in-process observation release directly. It may keep
its authenticated control connection until its own fixed entry/local IO returns;
it does not exec a command or gain another role's output/mutation capability.

`release` payload is `{action, inputs}`, except C01 compose_config requires
exactly `{action, inputs, environment}` with inputs=[] and the exact C01 sorted
name/value array. The supervisor derives it from the same admitted native model
inputs as the create request and validates the complete encoded map against
those inputs before sending; the fixed gate validates its closed names/types/
uniqueness and fixed controls, then explicitly supplies exactly that array to
exec. It never reads/merges inherited variables to construct compiler inputs.
Unknown/missing/extra environment fields refuse; other roles retain their old
payload and cannot receive this field. No map/value/hash enters retained control
diagnostics or receipts. This uses the existing authenticated control message,
not a new descriptor or environment file. For the other roles, inputs is an ordered array of exact
`{role, fd_index, access}` objects. access is read_only/duplex. The closed role
population is release_manifest, epoch, alert_config, database_url,
libpq_password, tls_rootcert, tls_clientcert, tls_clientkey, age_identity,
backup_ciphertext, sql_relay, and worker_broker. Supply ONLY roles actually read
by that fixed A2 command's existing flags/configuration: compose gets no input
FDs under C01; validators their own config; database helpers
manifest/epoch/DSN and configured TLS/password/relay; restore additionally its
backup/identity; alert helper only its fixed scoped HTTP channel (role
worker_broker, HTTP-only H06). No Engine broker role exists. No generic host
file/socket or extra role is admitted. File
inputs have independent offsets and read-only regular FDs; only the exact Q01/
worker broker role may be duplex. FD indices are contiguous in inputs order;
reject duplicates/extra/missing/non-native descriptors. No secret/hash enters
persisted evidence. The fixed gate installs only those declared descriptors,
closes ALL control/other descriptors on exec and executes the fixed command;
the PID1 shim retains only its own control/observation capabilities.

For private-mode backup_create and mount_inspector ONLY, PID1 creates a local
AF_UNIX SOCK_SEQPACKET result socketpair before the blocked child gate. It keeps
one endpoint; the child inherits ONLY the other as fixed CLOUD_CLICKER_JOB_RESULT_FD
in addition to its declared data inputs. This is not the Root control socket or
another release input role. Enable SO_PASSCRED; validate the actual child writer
against the original direct-child pidfd registered above. No other action gets
a result FD. Messages have exact keys `{schema_version:1, command_work_id, kind,
report}` with ONE SCM_RIGHTS FD. kind is staged_ciphertext with report=null for
backup_create, or mount_observed with the exact native_mount_report below for
mount_inspector. Reject extra messages/FDs, wrong native descriptor type/identity,
sender, work or kind. Parent reply is exactly `{schema_version:1, command_work_id,
result:accepted}` with no FDs; it only acknowledges local custody, not success.
The child closes its result endpoint and all writers/IO and returns independently
of any Root receipt. PID1 waits for actual child exit/local IO settlement BEFORE
forwarding the sole result over its own authenticated W00 channel. It closes the
local pair and copied FD after transfer. Loss, truncation or unknown child/writer
state leaves work outstanding; stdout and a supplied hash cannot carry a native
FD or close the effect. Command exec still inherits NO supervisor-control FD.

The mount_inspector alone may send `mount_observed` with exact payload
`{created_volume_sha256, storage_binding_sha256, report}` and ONE actual
mounted-root O_PATH FD. The Storage hash is null ONLY for first inspection.
The inline canonical report has exact keys `schema_version:1`,
`artifact_type:"native_mount_report"`, `created_volume_sha256`, `observer_work_id`,
`device`, `inode`, `owner_uid`, `mode`, `observed_at`. Counters use OP1 strings.
Validate report against actual FD fstat, native sender/CID/mount identity and
selected work before constructing Storage/current-observation objects. This
report has no unavailable future Storage hash, terminal receipt or secret value.
No other action/role may use this message or return a directory/root handle.
Root closes that O_PATH FD after native validation and before scope completion;
it is observation-only, not a retained mount/write capability. No scratch volume
or output file is needed for this first inspection.

A shim `output` message has payload `{role, fd_index}` with role
staged_ciphertext, fd_index0 and ONE read-only regular FD rooted in the admitted
encrypted-only staging generation. It is valid only for backup_create after the
actual writer has ended; validate native mount/object identity and F01 package
before publication. This is not a backup-success reply. `command_status` has
payload `{command_work_id, wait_status}` with actual integer wait status; it is
corroboration only. `end` has empty payload and no FDs; it permits fixed shim
shutdown only after command/local IO joins. Neither reply nor IPC EOF closes a
job: independently require original namespace INIT pidfd EXIT, exact CID removal,
restart prevention, and every external child terminal. For a trusted Go leaf use
its separate SP4 scope rule. Lost/late/ambiguous channels leave work outstanding.

## V00 — First physical volume observation and initialization

E19 first reserves a fresh generation/name and rejects any existing name before
dispatch. Its complete actual create response plus E20 observation produces the
immutable OP1 object with exact keys `schema_version:1`, `artifact_type:
"created_volume"`, `operation_id`, `phase`, `create_work_id`,
`engine_binding_sha256`, `generation`, `family`, `native_name`, `driver` local,
`engine_created_at`, `observed_at`. This is a creation observation, NOT Storage,
not a terminal receipt, and never permission for normal service/data use.
Original native daemon/handler and no-name-reuse authority remain mandatory.

The pending E19 item may admit its first fixed observe_native job through A2.
ONLY its mount_inspector role may reference `{created_volume_sha256}` as a Mount
scope, at /scope, read-only with VolumeOptions.NoCopy=true. No application,
backup/restore or generic worker may mount this provisional scope. The inspector
has no SQL/HTTP/secret or other storage inputs and performs only open/fstat of
the mounted root. W00 authenticates its exact CID/start/native process; it returns
the actual O_PATH root FD plus its read-only observation over that channel.
Native FD type/device/inode/UID and the exact approved mount/creation mapping
must agree; no claim the operator can traverse Docker's private mountpoint.
After the inspector/local IO exits and its scope is disposed, write the complete
immutable SP4 Storage binding from that actual observation, then the initial
native_storage_observation referencing the new binding. E19 can then close
its child and creation work. Normal use requires this binding; abandoned/unknown
inspection stays outstanding and cannot be adopted by name.

Current observation is a closed object with exact keys `schema_version:1`,
`artifact_type:"native_storage_observation"`, `storage_binding_sha256`,
`created_volume_sha256` nullable Hash (non-null for local volumes only),
`observer_work_id`, `device`, `inode`, `owner_uid`, `mode` Counter,
`transition` initial/private_scratch_init/image_population/postgres_init,
`transition_work_id` nullable UUIDv4, `image_config_sha256` nullable Hash,
`observed_at`. For transition=initial, transition_work_id and image_config_sha256 are null;
transition itself remains the literal initial. Other transitions require their
actual OP3 work, same physical root and the exact governed case below. This is
an observation history, never a rewrite of original Storage/Mount subjects.

For new job_scratch/protected_staging ONLY, after complete physical binding the
A2 initialize_storage job mounts that exact root read-write/NoCopy. UID0, private
namespace, cap_drop ALL/NNP and fixed pure-Go initialize_private_storage may
verify native root identity/owner0/empty, chmod ONLY that root to01777 and return.
No chown/capability/host mountpoint access or arbitrary file mutation is assumed.
The receiving job UID then creates its own fixed0700 workspace and0600 files;
no other job shares the generation. Observe the actual new mode/identity before
use. Root owner stays0, no secret copied to the shared outer root, and cleanup/
protection covers the complete generation. Existing/nonempty/foreign roots cannot
use this initializer or be recursively chmodded/chowned.

For the five persistent families, first inspection never copy-ups or mutates.
After physical binding ONLY the exact approved service E07 may populate an empty
volume from its manifest-bound image at the template's matching mount target,
preserving Compose's approved image-directory contents/ownership. Record its
image_population observation under that E07 work before service start; unknown
ownership/content transitions refuse. Postgres's existing governed startup may
then initialize/change ownership to its pinned runtime UID, observed separately
under that exact startup work. Monitoring image-owned directories likewise use
their actual pinned image ownership, never an arbitrary UID chosen by a worker.
No applied-data reinitialization, unobserved image swap or service read/write is
allowed before the corresponding binding/transition. All these native/empty-root/
copy-up/UID populations require actual conformance; this table is not that proof.

## A3 — Stage resource rules

The stage table is necessary but not sufficient. Apply these exact restrictions:

- Gateway network creation/removal and scratch-volume removal are permitted in
  their owning active stage ONLY for that job's exact temporary generations.
  The same plan ID cannot create/remove a retained network/volume before A3's
  live-resource boundary. All fixed R01 observation stages have their required
  E03/E04/E05/E06/E13/E15/E20 read plans; no reader reopens a sealed phase.
- Preflight/checkpoint jobs cannot stop/recreate/delete retained source services,
  volumes or networks. They may clean only their exact private jobs/scratch and
  report real helper failure. Add their E09/E10/F04 cleanup to the same stage.
- Before backup_ready, release may perform only prepare/preflight/checkpoint,
  scheduler freeze/export and preserved-prior observations. No bootstrap/drain,
  startup, live DB reset or target service handoff. A failed pre-backup release
  closes original work and conditionally restores SOURCE scheduler ownership.
- Drain after backup_ready may stop the exact source gameserver. Start replaces
  only admitted owned services, stopping old backup/publishers before new grants.
  Changed target Postgres16 config/image requires post-checkpoint source stop and
  exact target start; unchanged Postgres is retained/rechecked. No major upgrade.
- Rollback Started admits fencing ONLY: input/static/read-only checks and exact
  owned cleanup after authenticated backup authority. Do not require healthy
  seven-service CR2 or a readable failed DB as a condition of cleanup. Current
  Preflight's DB-size query cannot be copied unchanged as that admission gate.
  Space checks use authenticated protected input plus actual target-storage
  observation; missing/unknown measurements fail rather than guessed capacity.
- Only recovery_fenced admits restore: exact failed DB generation removal,
  clean target DB initialization, protected restore and target startup. Never
  down-migrate the live DB. Preserve the declared non-DB data/certificate scopes.
- Install uses OP7 before ANY first pull/create. Native supervisor issues E01/E02
  directly; no worker image is assumed to exist. Failed install cleanup is limited
  to bindings created by that attempt and its unfinished ancestors under OP7.
- Finalize admits no worker gameplay/SQL/HTTP or new service start. Finish existing
  operation helper metrics/publications/deletions and retained grants, seal phase,
  publish completion then Finished. Whole-release metrics are normal retained
  projection of the final row, never a pre-commit operation success write.
- Exact retained-service restarts under an already granted identity are normal
  runtime activity, not permission to restart a closed temporary job. Observe a
  fresh Start without changing historical Container/receipts. While an operation
  freezes/replaces that service, prevent its old restart policy from creating a
  new writer: close service ingress/restart admission and bind all native starts
  before declaring its old scope terminal. Unexpected restarts remain outstanding.

## A4 — Engine wire and terminal adapters

Engine routes below use API1.51, HTTP/1.1 and the original native generation;
only /_ping and /version may be unversioned for initial negotiation. Reject
ambiguous encodings/duplicates, extra query/body fields, API downgrade, auth/
plugin/context overrides and redirects. Never persist raw request/response,
environment, inspect JSON, error text or a digest of secret configuration.
The [pinned API](https://raw.githubusercontent.com/moby/moby/v28.4.0/docs/api/v1.51.yaml)
is a wire reference, not permission to expose its other endpoints.

| Adapter | Method/path; exact request derivation | Success response / required child disposition |
|---|---|---|
| E00 | GET /_ping, /version, /info; no body/query |200, exact native generation/version/info; no mutation |
| E01 | POST /images/load?quiet=1; original opened validated archive stream only |200, full progress stream without embedded error, inspect expected immutable config |
| E02 | POST /images/create?fromImage=<manifest repo>&tag=<manifest sha256 digest>&platform=linux/amd64; no body, no credentials |200, full progress without embedded error, inspect exact requested digest/config |
| E03 | GET /images/<exact config Hash>/json; no query/body |200, actual ID/platform matches authorized manifest |
| E04 | GET /containers/<CID>/json; no query/body |200, exact native ID/config/start/mount/network state;404 is absence observation only |
| E05 | GET /containers/json?all=1; no body/worker filters |200, COMPLETE locally filtered project plus recorded job inventory; labels are candidates, not ownership |
| E06 | GET /containers/<CID>/archive?path=<declared public artifact/tree>; no body/compression |200, complete exact expected file set (and directories), no links/traversal/extra entries, file/tree digest matches validated source/target bundle |
| E07 | POST /containers/create?name=<reserved generation name>&platform=linux/amd64; C-body below |201 exact returned CID; inspect/bind created-container children before use. Future start/native children are reserved and outstanding until actual E08/handshake return; never fabricate a Start at create |
| E08 | POST /containers/<CID>/start; no body/query |204, exact Start observed/bound; startup children already admitted; response alone does not close runtime |
| E09 | POST /containers/<CID>/stop?t=<existing drain/cleanup bound>; no body |204, actual whole native scope terminal, no pending restart;304 alone is not that proof |
| E10 | DELETE /containers/<CID>?force=false&v=false; no body |204 plus exact absence/native terminal/restart prevention; no extra volume deletion |
| E11 | POST /containers/<CID>/wait?condition=not-running; no body |200 complete typed status/error; corroborate original native scope, nonzero exit is failed but may be terminal |
| E12 | POST /containers/<CID>/attach?stream=1&stdout=1&stderr=1&stdin=0&logs=0; no input bytes |101 raw-stream mode or200 per pinned negotiated transport, original server EOF plus owned job terminal; no application stdin |
| E13 | GET /containers/<CID>/logs?stdout=1&stderr=1&follow=0&timestamps=0&tail=all; no body; forbidden for compose_config CID |200 entire multiplexed output, decode public required result only, never persist sensitive output |
| E14 | POST /networks/create; exact Name, Driver=bridge, Internal per role, EnableIPv6=false, IPAM default, Labels profile generation; no options/plugins |201 full network ID, inspect/bind before connect/use |
| E15 | GET /networks/<CID>; no body/query |200 exact bridge/internal/endpoint generation,404 observation only |
| E16 | POST /networks/<CID>/connect; exact Container CID and approved aliases, no IPAM/MAC override |200 plus actual exact connection; no user-selected network/name |
| E17 | POST /networks/<CID>/disconnect; exact Container CID, Force=false |200 plus exact endpoint detached, ingress/children already fenced |
| E18 | DELETE /networks/<CID>; no body/query |204 plus exact native absence, no continuing endpoint/request |
| E19 | POST /volumes/create; exact unique Name, Driver=local, empty DriverOpts, exact generation Labels |201 immutable CreatedVolume observation under V00; request remains outstanding through its declared first-inspection child and complete Storage binding. Existing-name collisions reject, never adopt |
| E20 | GET /volumes/<recorded native_name>; no query/body |200 exact local generation metadata; physical identity requires V00 mount observation, never an E20-invented stat;404 observation only |
| E21 | DELETE /volumes/<recorded native_name>?force=false; no body |204 plus exact generation absence and all mount/job users terminal |

C-body is built ONLY by the supervisor from ONE admitted A2 command or approved
service template. There is no worker mutation-body forwarding/normalization path.
It contains ONLY the Docker Config/HostConfig/NetworkingConfig fields populated
by the closed mapping below from the pinned, validated typed model. All worker
mutation requests refuse; only the supervisor-built body may reach the daemon.
Resolve image to exact config ID; bind sources through held declared roots;
networks by full ID, storage by non-reused generation. Service defaults derive
from the authorized image config, not arbitrary daemon-side tag defaults.
For jobs enforce the A2 shim, exact command/input roles, no restart/AutoRemove,
no ports/devices/host namespaces, read-only root, cap_drop=ALL, no-new-privileges,
private PID/IPC, no writable host output. Live services use their separately
approved template; do not accidentally strip Postgres's existing required init
permissions or give those permissions to a command worker. Unsupported Compose
fields/features refuse before dispatch, not silently pass through.

The C-body converter has this closed mapping over the pinned API1.51 types (not
an unspecified future converter). Config.Image is manifest config ID; User,
Entrypoint, Cmd, WorkingDir, Healthcheck, StopSignal/StopTimeout and ExposedPorts
come from the selected image defaults overridden ONLY by the exact approved
role/command template. Environment merges those same defaults/template into one
unique-key map, encoded in key order. C01's request uses its explicit map, then
validates the actual image-default merge before shim start and replaces the child
exec environment through its authenticated release; Config.Env alone does not
remove inherited defaults. Nonsecret values may enter ordinary trusted daemon configuration; actual
credentials remain authorized FD inputs. No environment/inspect payload or
private-config hash is copied into operator evidence. Jobs replace
Entrypoint with the fixed shim and carry the fixed command from A2. AttachStdin,
OpenStdin, StdinOnce and Tty are false; stdout/stderr attach is allowed only for
the bound temporary job. Labels are the exact computed project/service/oneoff/
generation metadata of that role; no arbitrary caller label or ownership claim.

Temporary roles explicitly disable inherited image healthchecks (Test=["NONE"])
and Docker --init/AutoRemove; no unadmitted daemon healthcheck/exec process can
appear in a trusted leaf or after a temporary worker finishes. Only the fixed
pure-Go native_observer entry may join the exact Postgres PID namespace under
SP4 trusted_go_leaf; all CLI/backup/restore/publisher workers use private PID
namespaces. Its no-fork/exec closed membership is a producer contract requiring
source/native review, not a label a caller may set.

HostConfig.Mounts is the exact target-ordered held input/storage projection,
ReadonlyRootfs/Tmpfs/RestartPolicy/LogConfig come from the approved role. C01
compose_config requires LogConfig Type=none/empty Config, NetworkMode=none and
empty NetworkingConfig; validate these before E07. Existing service logging/
retention and other fixed job roles remain. PortBindings are permitted ONLY for
Caddy's two already configured ports.
For other roles, NetworkMode and NetworkingConfig contain only declared full network IDs/
aliases, including all four retained network roles where required. Workers never
select host PID/network namespaces or user remapping; they have private IPC/cgroup namespaces, Privileged
false, no Devices/DeviceRequests/VolumesFrom/extra groups, CapDrop ALL, no added
capability and no-new-privileges. Trusted Postgres-namespace observer/retained-backup roles
alone may select that exact Postgres PID namespace. No arbitrary namespace path.
Other recognized API1.51 Config/HostConfig fields may be absent or the pinned
schema's zero/default ONLY, never a new nonzero capability/resource/route. Unknown
fields fail even if the daemon would ignore them. Compare absent/zero by that
explicit pinned type; reject duplicates and ambiguous representations. No Docker
build, plugin, privileged exec, arbitrary archive import/export or generic proxy.

**Common native terminal classifier:** own/join the original upload and response,
parse entire framing/body (including image progress errors), then issue a fresh /_ping
on THAT SAME ordered HTTP/1.1 connection and verify native writer against held
pidfd. This establishes a handler-order barrier only for non-hijacked handlers;
it does not close reserved resources/streams. The inference is grounded in
[Go's serial connection loop](https://raw.githubusercontent.com/golang/go/go1.25.1/src/net/http/server.go)
and must be confirmed for the actual pinned daemon build, not assumed for HTTP/2,
proxies or another implementation. Archive resource cleanup is inside the
[pinned handler](https://raw.githubusercontent.com/moby/moby/v28.4.0/api/server/router/container/copy.go).
The [image handler](https://raw.githubusercontent.com/moby/moby/v28.4.0/api/server/router/image/image_routes.go)
waits on its backend and can report errors inside a successful HTTP stream.

If the daemon deliberately closes a non-hijacked connection, an actual ORIGINAL
raw server EOF after complete response/upload is an alternate handler-end
observation; response-body EOF at Content-Length, client Close/cancel or a local
read limit is not raw server EOF. Record handler_end as ordered_reply/server_eof;
for E12 use readonly_hijack_eof. Include writer Process and original socket_cookie
in the native result; never silently replace that connection/generation.

E12 has no reusable HTTP barrier: require ORIGINAL upstream EOF after the exact
temporary namespace/forwarders end. It carries no stdin or mutating child route;
client disappearance is not upstream EOF. Read-only logs cannot be followed
indefinitely or treated as complete after a caller-side byte cap. Any truncation
or unavailable terminal observation invalidates the result.

No forwarded4xx/5xx,304, timeout, cancellation or404 is a blanket terminal refusal.
A non-dispatched local rejection may emit SP4's typed refusal. Forwarded errors
retain admitted children and become terminal only after the same complete native
handler/child checks, with a FAILED business outcome. Lost writer/connection,
unknown return ID or uncertain native job remains observation_failed/outstanding.
No automatic retry of a mutating request. Each reserved child must bind or have
independent proof it was never dispatched/created; never infer it from a response.

## A5 — SQL, HTTP and filesystem producers

**Q01:** fix upstream to the exact admitted Postgres generation and worker/role.
Before dial sync admission; before client bytes bind the observer's held native
backend pidfd/full socket tuple/inode/UID/direct-parent check. Preserve TLS
hostname/client-cert semantics, never parse keys or change verify-full. All
actual pgx/libpq connections are individually counted/admitted, including early
startup. Native terminal requires original backend pidfd exit AND original EOF/
reset AND all forwarders ended; other errors remain unknown. Q02 can signal only
that original validated pidfd after worker ingress closes; wrong binding refuses
before signal, and signal delivery is not a terminal receipt. Bystanders survive.

**Q03:** import only an original stream already admitted under the verified
retained source controller's normal runtime authority. Its runtime origin is a
closed OP1 object with exact keys `schema_version:1`, `artifact_type:
"retained_stream_origin"`, `source_identity`, `service_container` Container,
`service_start` Start, `retained_grant_sha256`, `round_id`, `stream_id` UUIDv4,
`socket` Socket, `native_admitted_at`, `native_bound_at`, `freeze_generation`
UUIDv4, `export_cut` Counter, `observed_at`. Dates are real original producer
observations, not permission inferred from their ordering. Validate the live
trusted source producer, original native handles and full retained grant; no
source crash/reconstruction or label/activity-based substitute.

Sync current-phase Q03 admission/binding BEFORE takeover/signal/cleanup, linking
the original source ownership proof through runtime_origin_sha256. This admits
NEW cleanup ownership, not the historical dial; do not falsely claim an OP3 row
preceded an effect that occurred under prior normal retained-service authority.
I04 freezes the complete snapshot/cut while both parties verify; source retains
its original handles until the supervisor acknowledges durable complete import.
Thereafter source admits no next-stage work (including failure-metric writes)
for those rounds. Already-started forwarders/writers remain explicitly tracked
until joined/terminal; no concurrent reader may consume the same server stream
merely because a duplicate FD was transferred. Native EOF observation stays with
the actual forwarding producer until it ends, then the receiver may drain its
original handle. Scope cleanup cannot classify a copied/truncated TLS buffer as
success of the original backup. Q03 closure still requires Q01's full terminal
conditions and corresponding original forwarder/writer observations.

**Private application observer:** extend the EXISTING gameserver, no new public
endpoint/service. Provision I01/I02 below at gameserver UID65532. Before H02–H04,
register the operation/request and a one-use in-memory nonce over that channel;
the public probe carries it only in a dedicated operation-observation header.
It grants no game/API permission and never replaces normal authentication.
Do not persist the nonce/token, bootstrap key, response payload or secret digest.
The trusted server binds it once to the actual handler/socket and records its
awaited transaction/task disposition. Reuse/wrong tuple/peer/endpoint refuses.

| Adapter | Fixed request | Native terminal / result |
|---|---|---|
| H01 | GET exact configured receiver-health URL; no redirects/body | full2xx response under its read-only contract; no claim of alert delivery |
| H02 | POST /api/v1/bootstrap, existing JSON grammar with fresh in-memory key | full201 valid registry response plus observer handler-return/all owned SQL tasks settled; failure/cancel is unknown until scoped cleanup |
| H03 | GET /api/v1/founder/state, same bootstrap session | full200 valid expected constants/founder plus observer completed-task disposition |
| H04 | WSS /connection/websocket with canonical Origin; only existing connect token and world subscribe commands | matched connect/subscription responses, native server disconnect callback AFTER subscription/task removal, original socket close/readers ended; client Close alone insufficient |
| H05 | GET /readyz on exact governed app; existing ready204 or draining503+required marker | full response; this is readiness observation, not drain/job completion |
| H06 | fixed existing alert test to exact Alertmanager, configured receiver only | native operation helper terminal plus parent's actual nonce-attributed receiver proof; aggregate counters alone cannot authorize acceptance |

H06 admits each actual HTTP request separately with route alert_health,
alert_metrics or alert_post. Health is the exact configured read-only URL;
metrics is GET /metrics on the exact target Alertmanager; post is POST
/api/v2/alerts with the fixed existing CloudClickerReceiverTest/info labels and
proof_nonce annotation. Root generates/binds the fresh128-bit hex nonce before
forwarding; do not accept an old/self-selected proof or arbitrary alert payload.
Root worker transport is a private Unix HTTP channel with a fixed per-job route
map; it has no general CONNECT/proxy, redirect or caller-selected upstream.
Current helper gets this scoped RoundTripper/profile in the proposed implementation;
its current direct-network client is not a fulfilled adapter. No worker joins
the operations/edge network to obtain unconstrained HTTP access.

Alertmanager may own the deliberately submitted notification under its declared
retained runtime scope. Successful submission alone is NOT proof of delivery.
The current VerifyAlertDelivery waits on aggregate attempts/failures, which cannot
attribute an increment to this nonce; do not claim it already supplies this gate.
Business acceptance additionally consumes a closed `receiver_delivery_observation`
with exact keys `schema_version:1`, `artifact_type:"receiver_delivery_observation"`,
`operation_id`, `supervision_sha256`, `target_manifest_sha256`, `nonce`,
`receiver_scope` configured_receiver, `producer` native_capture/operator,
`operator` nullable existing operator ID (required for operator), `capture_binding_sha256` nullable Hash (required for native_capture),
`submitted_at`, `received_at`, `observed_at` Date, `outcome` received.
Dates/tuple/nonce must match the exact current submission, never an older test.
No receiver credential/address/payload is stored. Native capture is the actual
controlled receiver's observed request under the declared conformance producer;
production receivers without such an interface use authenticated operator
attestation of the actual received nonce. Status/control accepts that mechanical
observation as a read-only regular FD, copies validated canonical bytes privately,
and never calls it an independent automated proof. No new provider/key/receiver
API is required or silently invented. This is an explicit proposed manual
checkpoint for those receivers, requiring DP7/runbook adoption; the caller's
mere existence/health acknowledgment is not it. Missing proof leaves smoke
incomplete and the operation's existing guard/RTO gate unchanged.

Successful normal handler work on the retained service's pool must be actually
settled before the observer acknowledges; an idle pool is normal runtime-owned
work, not an operation sql_stream falsely marked retained. Any private Q01
connection still requires its original native terminal proof. A lost/canceled
handler with unknown Account/actor/SQL disposition cannot acknowledge success;
after checkpoint CR4 cleanup of the exact owned app/DB scopes is the fallback.
No such mutating request is allowed before checkpoint. Observer lists must match
registered request/children, not a caller's claimed zero count.

Backup helpers use TWO separate scratch scopes: private plaintext/credential
work under /tmp, and encrypted output-only protected_staging under /staging.
Never retain a mixed plaintext/dump volume as a backup. Completed or uncertain
produced ciphertext is subject to DP6's EXISTING unrecorded-backup protection
until its governed resolution, even if final host copy/checkpoint fails. Publish
it successfully or transfer that encrypted-only storage under R03 to explicit
source/target ownership with actual protection/object identity before ending
its job. Retain no process/job namespace. Include every such generation in SP7
inventory; no automatic eviction, retention shortening or new signing key.
Private plaintext cleanup still requires original writers ended and actual F04.

**F01:** a trusted in-supervisor publish_ciphertext leaf reads only the staged ciphertext FD from the ended job scope,
validates package/header/accepted protection authority and expected release/server,
publishes no-replace, syncs file+directory and verifies exact final ciphertext
digest/size. No plaintext/key digest. If publication fails before commit, a typed
not_published result can close FAILED work only after actual publisher/IO end,
own temporary cleanup+directory sync and proof no new final object appeared.
Possible completed ciphertext stays protected as above; unknown FS/remote IO
cannot use this variant. It records null final digest/bytes0 and the non-null
protected staging binding when any object was preserved. This creates no backup
Ready/success authority. Existing target collision is unchanged/unowned, never
cleanup permission. No false terminal refusal from an exception string.

**F02:** a trusted in-supervisor publish_metric leaf (no worker/container) generates the closed helper
metric from actual stage outcome, replace+file/directory sync, no arbitrary metric
names/body. Backup success additionally requires F01 validated final ciphertext
publication; a worker staging success is not a completed backup. Projection status
uses only the fixed diagnostic gauges below, never a release-success result. The
leaf closes all IO before its receipt; host publication references that leaf
receipt, with no parent cycle. **F03:** a trusted in-supervisor delete_retained_object leaf revalidates exact generation/target identity and accepted
checkpoint/retention decision immediately before unlink; sync directory and end
publisher. No signed-protection rule/key is adopted by this adapter. **F04:**
remove only exact private scratch after all native writers/streams end, record
scope disposition without hashing its secret contents. A symlink/mount/inode
change, policy error, publisher error or surviving request never closes work.

**R01–R03:** observe complete SP7 inventory using held native identities and E04/
E06/E15/E20, not label-only discovery. Removal requires whole source scope,
admitted starts, streams, local descendants and writers terminal. Temporary jobs
must satisfy SP4's bound private namespace INIT-exit proof, not generic parent
exit. The [Linux PID-namespace contract](https://man7.org/linux/man-pages/man7/pid_namespaces.7.html)
provides whole-membership termination/no-new-process semantics; native conformance
must bind the actual namespace/init and defeat setsid/double-fork/blocked-IO
escape controls. A cgroup path/404/populated flag is not used as a substitute.
For retained services bind actual current Start and exact source/target policy,
all mounted bytes/generations, scheduler/cache/restart state and closed operation
ingress. Retained ordinary service work is explicit; temporary jobs cannot be
renamed retained. Historical observations are immutable, current admission fresh.

## A6 — Private IPC and service cache adapters

All messages are OP1 canonical JSON in Unix SOCK_SEQPACKET with exact keys below,
plus SCM_RIGHTS in the declared order. Reject unknown/duplicate/trailing fields,
wrong version/tuple/sequence, MSG_TRUNC/MSG_CTRUNC, extra/missing ancillary/FDs,
wrong native types and late replies. Each connection has a fresh256-bit challenge
and monotonic Counter sequence; challenges remain memory-only. No silent partial
export, numeric PID reopen, fabricated empty activity or replayed acknowledgments.

I01 prepares `service-ipc/<service>/<generation>` under the operator's0700 root.
Each leaf is operator-owned0777, mounted ONLY into that exact trusted service;
service creates its0700 cache child and0600 regular cache files plus0666 socket.
The outer parent stays0700/unmounted and cannot be traversed by another login.
Workers receive no leaf. Reject symlink/component/native-mount substitution.
The operator need not chown another UID's files; read-only metadata is granted by
already-open file FD, cache writes by the trusted service's own UID. No claim
that an open directory bypasses mode checks. Preserve exact leaf bindings across
allowed service restart; a new generation needs its own admitted leaf.

I00 bootstrap_status is a separate admission-only message: exact common keys
`schema_version:1`, `bootstrap_id`, `sequence`, `command` bootstrap_status,
`challenge`, `payload`. Payload has exact keys `source_identity`,
`container_sha256`, `start_sha256`, `producer_manifest_sha256`, `profile_sha256`,
`ledger_prefix_bytes`, `ledger_prefix_sha256`, `ledger_fd_index`. Reply result is
observing with exact `bootstrap_id`, `sequence`, `challenge`, `payload`; payload
has `service`, `container_sha256`, `start_sha256`, `cache_sha256`, `mode`,
`accepted_prefix_bytes`, `accepted_prefix_sha256`, `ledger_device`, `ledger_inode`.
It cannot freeze/resume a scheduler or create new identity/ownership. Check actual
native peer/generation and prior retained grant before accepting a read-only
ORIGINAL ledger FD; its device/inode/prefix must match that granted ledger.
The service may renew its in-memory read capability and perform its ALREADY
retained metadata observations, not a new operation mutation. No new Started,
execution log or fake supervision hash is used for this bootstrap exchange.
A0 service-status projection is the same reply payload plus schema_version1,
artifact_type native_admission_service, bootstrap_id and observed_at. Unknown
mapping, stale/missing original capabilities or unsupported source profile fails
admission/diagnostics rather than using a cached prefix as a current read.

Every message has exact common keys `schema_version:1`, `operation_id`,
`supervision_sha256`, `sequence`, `command`, `challenge`, `payload`; reply replaces
command with `result` and echoes tuple/sequence/challenge. challenge is IPC-only,
never an evidence object. I02 authenticates two fresh challenge replies against
actual expected service init PID/held pidfd and full CID/config/native start,
plus actual writer credentials. Service authenticates the prepared supervisor
channel and matching descriptor/prefix. Validate UID mappings explicitly; unknown
user-namespace remapping refuses, never assume container UID equals host UID.

| Adapter / command | Exact payload keys | Reply result/payload; descriptors |
|---|---|---|
| I02 hello | `service`, `container_sha256`, `start_sha256`, `profile_sha256`, `ledger_prefix_bytes`, `ledger_prefix_sha256` | authenticated with same fields; no FD |
| I03 freeze | `owner` source/target, `identity`, `generation`, `cache_work_id`, `publisher_work_id` | frozen with `generation`, `cache_sha256`, `cache_publisher_receipt_sha256`, `active_rounds` UUID array, `settled_host_work` Hash array; fsynced pause, closed new effect admission and ALL in-flight host publication/deletion IO joined before reply |
| I04 export | `generation`, `round_id` | exported with `round_id`, `items` exact array below; one atomic frozen population, no new admission |
| I05 grant | `owner` source/target, `identity`, `expected_operation_id`, `condition` preserved_prebackup_failure/target_success/install_success, `ledger_fd_index`, `metadata` Hash array, `cache_work_id`, `publisher_work_id` | conditional with same owner/identity/operation/condition, `cache_sha256`, `cache_publisher_receipt_sha256`; first FD original read-only ledger, then each exact metadata object in hash-array order |
| I05 proof | `hashes` Hash array | metadata with same hashes; read-only regular metadata FDs in order, no effect permission |
| I05 status | empty object | state with `generation`, `mode` frozen/conditional/retained, `cache_sha256`, `accepted_prefix_bytes`, `accepted_prefix_sha256`; no FD |
| I04 adopted | `generation`, `round_id`, `export_cut`, `imports` exact array of `origin_sha256`, `work_id`, `subject_sha256` | adopted with same fields; source closes only its exported ownership copies, no new work for those rounds; receiver already synced all imports |
| I04 round_status | `generation`, `round_id`, `export_cut` | observing with same fields, `items` array of exact `item_id`, `state` pending/settled/unknown, `native_result_sha256` nullable Hash; settled requires actual original tasks/IO joined, not producer disappearance |
| I04 cancel_round | `generation`, `round_id`, `export_cut` | cancelling with same fields; fixed original context cancellation, NOT a terminal result |
| Q02 signal_stream | `work_id`, `binding_sha256`, `signal` terminate | signal_status with same work/binding and `native_signal` sent/already_exited; only original held validated pidfd, actual SIGTERM/ESRCH, no PID reopen/arbitrary signal |
| I02 register_probe | `request_id`, `action` H02/H03/H04, `nonce` in-memory one-use256-bit value | registered with `request_id`, `action`; no FD |
| I02 probe_status | `request_id` | pending, settled or unknown; payload `request_id`, `handler` pending/returned, `socket` not_applicable/connected/disconnected, `children` UUID array, `unsettled` UUID array; no FD |
| operator submit_delivery | `nonce`, `observation_sha256` | recorded with same fields; one read-only regular observation FD, only on authenticated operator control channel for this pending test |

A plain file/operator channel can attest only producer=operator. Native_capture
requires the actual registered controlled-receiver context and capture binding
from the separately governed R-006 conformance producer; a file/flag/hash naming
that producer cannot instantiate it. Until that producer/profile is accepted,
ordinary operations use actual operator attestation. Neither path is inferred
from Alertmanager counters. This preserves manual/native provenance explicitly.

No native receipt may be minted from pending/unknown/cancelling or nonempty unsettled. A
settled reply is trusted in-process producer observation over the authenticated
channel, not a user boolean. H04 requires disconnected and registered callback
completion. Source abort grant requires actual failed pre-checkpoint Finished,
non-null complete closure/preserved-prior proof and exact SOURCE identity;
target grant requires successful complete target Finished. install_success instead
requires the actual successful v1 install row AND complete OP7 InstallHandoff/
attempt graph; it never waits for an invented v2 install Finished. Publishing
that OP7 handoff after its v1 anchor is permitted metadata finalization, not a
new worker/effect admission. An unresolved failure never becomes a success grant.

For a pre-transfer cache write, sync OP3 cache publication and its native
publisher leaf admissions before the I03/I05 request; work IDs above must name
that exact parent/leaf. The trusted service leaf entry publish_cache writes,
syncs, closes all IO/write capabilities and returns before its native receipt.
It is NOT a worker/CLI or a requirement that the whole service process exit.
Root validates the original authenticated producer/native start and exact final
cache bytes/sync, then closes leaf → cache publication → requesting control job.
No receipt depends on its own parent publication. After a valid conditional R03
transfer, ordinary service cache/projection writes are retained service work;
old outstanding operation writers cannot use that exception.

Each exported item has exact keys `item_id` UUIDv4, `kind` sql_stream/native_job,
`runtime_origin_sha256`, `binding_sha256`, `fd_indices` Counter array. The
binding is the complete SP4 object obtained as a separate read-only metadata FD,
then original handles in order: SQL original relay stream/backend pidfd/forwarder
scope; native child pidfd/native scope. Indices are unique and exhaustive; ALL active
rounds/items must be exported and acknowledged before stop/replacement. Existing
service-trusted settled objects may be copied only with their full native proof.
Unknown/unexportable backend or active deletion refuses; never reconstruct it
from pg_stat_activity or a pathname. Exact native FD kind validation follows SP4;
all claimed native scope formats/producers still require actual conformance.

Freeze closes admission BEFORE joining the source's currently executing file
publication/deletion tasks. Those are prior retained-service effects, not new
OP3 effects; they must actually settle while still source-owned. Export cannot
skip them by returning a zero count. Native source receipt `retained_host_result`
has exact keys `schema_version:1`, `artifact_type:"retained_host_result"`,
`source_identity`, `service_container`, `service_start`, `retained_grant_sha256`,
`round_id`, `item_id`, `kind` host_publication/host_deletion, `scope` Storage,
`target` Relative, `target_sha256`, `policy_sha256` nullable Hash (required for deletion), `entry` publish_ciphertext/publish_metric/
delete_retained_object, `io_end` returned_and_synced, `observed_at`. Validate the
trusted source's original tracked IO-entry return, full final object/policy/sync,
closed write capability and native live producer; no worker reply/timestamp alone.
All such receipts are in settled_host_work in item-ID order. A blocked/unknown
write/deletion means freeze has NOT completed, regardless of SQL readiness.
No source fail/finalizer may open a new metric or delete after this cut.

For a native child, runtime_origin_sha256 instead names a closed object with
exact keys `schema_version:1`, `artifact_type:"retained_native_origin"`,
`source_identity`, `service_container`, `service_start`, `retained_grant_sha256`,
`round_id`, `item_id`, `freeze_generation`, `export_cut`, `native_admitted_at`,
`native_bound_at`, `observed_at`, `process` Process, `scope_sha256`,
`children` UUID array; no Socket fields.
source_identity in these origin objects means the ORIGINAL retained-service
identity; it may match source OR target of the CURRENT rollback, as derived from
accepted ancestry. I03 owner/identity must match that exact binding; an ordinary
release may freeze only its source, while recovery may fence either proven owned
side. Unrelated third identities and image-only matching refuse.

It binds the original actual child/descendant scope. Source Go routines can
report settled only after actual IO/tasks join; an unfinished routine is not a
dead OS process. Current cleanup keeps that original source producer live until
these routine/forwarder tasks settle, or fails the preserving-abort path. It
cannot terminate the whole shared Postgres PID namespace to clear one backup
child. No unproved cgroup/PID-metadata substitute satisfies native scope closure.

An exported retained-service scope is distinct from a temporary worker's whole
job_namespace. Its closed object has exact keys `schema_version:1`,
`artifact_type:"retained_native_scope"`, `container_sha256`,
`service_start_sha256`, `kind` serial_postgres_client/service_routine,
`process` Process, `entry`, `round_id`, `invocation_id`, `children`, `observed_at`.
For serial_postgres_client, entry is pg_dump_serial/pg_restore_serial with the
EXACT fixed nonparallel A2/source helper argv and source image/helper bytes;
children must be empty under that verified executable contract. This exception
is not available to arbitrary commands/shells/CLI workers or unreviewed source
versions. Prove that whole executable/descendant contract on the pinned clients,
not just waitpid on a generic child. Its FD is the original child pidfd. For
service_routine, entry is go_snapshot/opaque_forwarder/private_scratch_cleanup;
process is the original SERVICE, whose exit cannot masquerade as routine success.
Use its held pidfd and authenticated I04 invocation status; source child-spawn/IO
registry must be closed and every actual child joined. No free-form callback or
untracked goroutine qualifies. This is a proposed trusted-code producer boundary,
explicitly requiring source/compatibility and native conformance review.

SQL export includes original socket and backend pidfd followed by retained-source
process pidfd; forwarding scope identities are immutable metadata referenced by
the binding and observed through that same authenticated I04 producer. Serial
child export includes its original pidfd; routine export includes original service
pidfd. Metadata FD is always first. Validate descriptor count/type against this
exact kind; never treat the shared Postgres namespace FD as a private worker.
Source may keep observation-only native copies until actual closure; it can signal
only a current authenticated Q02 request and cannot admit fresh work. Check native
pidfd exit independently for OS/backend subjects. A returned routine status needs
its original live producer and fully joined tasks; authority loss is unknown.

Cache has exact keys `schema_version:1`, `service`, `generation`, `mode`
frozen/conditional/retained, `owner`, `identity`, `operation_id`, `condition`,
`accepted_prefix_bytes`, `accepted_prefix_sha256`, `active_rounds` UUID array, `objects` Hash array sorted.
Its copied immutable objects are the full required verified graph, not summaries.
Before ANY normal scheduled round effect, sync that round UUID into active_rounds.
Clear it only after ALL actual native streams/children/IO and outer overrun/failure
metrics settle and the cache sync completes. After process/container restart,
nonempty active_rounds means lost original authority and MUST stay paused even
if mode was retained. Never reconstruct a completed round from an old cache or
new empty in-memory registry. A live producer's exact freeze/export registry must
match its durable active rounds; torn/missing markers invalidate ownership.
Private no-replace object publication and atomic snapshot replacement both sync
file+directory. After restart only a complete retained cache with matching service
identity and no durable freeze can resume ordinary work; frozen/conditional/torn
states stay paused. A copy cannot authorize new ledger operations. Parent retain
proof includes this policy; service's normal cache/result projection may continue
after Finished, but no operation worker/admission reopens.

The result projector validates the original ledger/required cache graph and
strictly advances accepted prefix under ONE publisher lock. It never rewrites
newer result from an older prefix. Old source publisher must be fully terminal
before target ownership. No fresh metric on stale/missing authority. Missing/
blocked projection must be covered by the adopted DP7 observation-staleness gate;
existing host.prom freshness alone is not proof this new projector is working.

## A7 — Construction and conformance boundary

This table is proposed policy, not a claim its producers exist. Before technical
approval/adoption, review the exact expansion, closed body construction, native
terminal classifiers and bootstrap/control graph together with SP/OP/CR. Before
feature acceptance execute the complete SP8 population with these actual adapters,
not table counts. In particular, handler-order barrier, actual namespace-init death and no descendant escape,
same-ID restarts, cross-UID leaf/FD transfers, private application completion,
full frozen-round export, cache/reboot/Finished cuts, first uncached image pull,
unknown creates and all source-preserving/recovery paths need real proof.

No guard is an operating budget, no transient error proves absence, no cgroup/
namespace metadata alone proves physical termination, and no API success closes
its children. Keep actual native supported-host/R-006 and exact-artifact gates.


Projection diagnostics use the already configured15s scrape/evaluation cadence
and300s stale-observation floor as explicit proposed DP7 bindings, subject to
adoption/native measurement; do not increase those thresholds on a failed run.
Emit closed gauges projection_observed_timestamp_seconds, projection_success
and transition_owner_known under the cloud_clicker_release_ prefix, without
per-user labels. Only a fresh valid ORIGINAL ledger read advances observation;
a cached old prefix is not fresh. Success decisions require complete sealed
proof, while valid failure/pending decisions keep CR3's open-work semantics and
cannot become completion/recovery authority. Page for absent/stale observation,
projection error or unknown active transition owner under that existing floor;
host.prom alone does not cover it. A source restart lacking its original live
ledger/owner capabilities cannot claim current observation from cache alone.

The source projector freezes/joins with the rest of I03 host IO. During an
admitted transition the supervisor may publish ONLY projection-status gauges
through its tracked F02 jobs; it cannot publish release last-result success.
Stop/join this loop and every publication before worker admission closes and
R03 transfers the single projector owner. Conditional target/source retention
permits ordinary metadata projection of pending/failure history; scheduled SQL/
retention remains blocked until the exact I05 condition. No dual publisher,
no timer IO after verification_closed, no extra post-Finished CLI writer.
