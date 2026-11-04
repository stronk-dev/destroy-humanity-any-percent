# RFC: Current Release Authority and Interrupted Transitions

- **Status:** draft — RP-558 / Deployment DESIGN-GAP10; proposed safety contract, not implementation authority
- **Author:** Codex
- **Created:** 2026-10-13
- **Design refs:** `design/06 §stack`; `design/00` pillar6 (honest claims)
- **Depends on:** Deployment Foundation DP1/DP4/DP5; its supported bundle/host floor
- **Parent / amends:** [Deployment Foundation](deployment-foundation.md), DP5 and AC3/AC5/AC10
- **Supersedes / superseded by:** —
- **Planning:** shared owning `planning/deployment-foundation/plan.md` and append-only log

## Summary

Bind an ordinary release to the actual current deployment, not two caller-selected version
strings. Use the existing operator ledger as the single authority, adding explicit durable
in-progress/checkpoint/finished events so an interrupted release retains truthful rollback
authority. Keep the one-node manual workflow, seven-day completed-release window and exact
pre-upgrade restore; add no deployment service, automatic retry or automatic rollback.

This is one recommended engineering proposal. Its new admission, interruption and journal
rules need adoption and designated normative review before code changes. Existing v1 behavior
is not silently reinterpreted by this draft.

## Motivation and limits

RP-558's two-case controller diagnostic recorded a successful2.0.0 installation yet admitted
1.1.0 from supplied1.0.0. Both cases executed all seven recording-runtime stages. That proves
missing controller binding, not a real-host downgrade. RP-556 now preflights history admissibility;
it deliberately did not invent the missing current-release policy.

Six manifest images back seven Compose services: `backup` uses the Postgres image. Caddy and
helper services also mount release-owned files. Image-only matching therefore cannot establish
the bundle being used. Current terminal-only v1 records also leave no completed release row if
the helper dies between backup/drain/start and final append. Rejecting a stale path alone would
leave that interrupted operation without a governed recovery route.

Out of scope: host provisioning/CD, multi-node delivery, legacy-host adoption, automatic recovery,
new secret access, retention/backup-resolution policy, D-019 decryption-before-release policy,
the unmerged R-006 proposal and DESIGN-GAPs1–9. None is closed by this amendment. Full nine-tier
1.0 and the exact-artifact clean-host floor remain unchanged.

## Specification

### CR1 — One current identity, derived from governed history

A `ReleaseIdentity` is an exact object with `release_version`, `manifest_sha256` and `images`.
Version/hash grammar is the existing release-manifest grammar. `images` is the six-element
manifest-ordered array of exact `{name, reference, runtime_config_sha256}` objects, with the
existing image-name/reference/config-hash rules. The manifest hash covers the original manifest
bytes; tuple equality is not version-string equality. Image hashes come from the validated
manifest, not a new handwritten table.

Under the existing shared operator lock, reduce the admitted ledger to either:

- **known:** exact identity established by a successful install/release/rollback;
- **in progress:** one admitted unfinished operation, with its durable checkpoint if any; or
- **uncertain:** a failed post-backup release/rollback whose live state is not established.

An ordinary release requires `known`, exact supplied-current equality, and CR2's current
observation. Candidate version must be strictly greater and migration nondecreasing relative
to that identity. Empty/missing history permits clean initial install only, not an ordinary
release. Do not synthesize an install row, trust a supplied directory, or add an adoption override.

Pre-backup failed release attempts leave the prior identity known only after
whole-operation closure and preserved-prior identity are proved under CR3/CR6.
An unfinished or unclosed failed helper remains in progress/uncertain even if
its last completed database read was harmless. Once a release has reached
the durable backup checkpoint, a failure makes ordinary release inadmissible until governed
rollback succeeds. Failed rollback remains retryable against the exact same authority; successful
rollback closes that authority and establishes its target as current. No newer/unrelated attempt
may reuse an older backup chain. Installation/recovery prerequisites remain the parent's gates.

### CR2 — Observe the containers and release-owned bytes

Add a mandatory current-observation capability to the release runtime. Absence is an admission
failure, never a no-op adapter. Before accepting an ORDINARY RELEASE, validate the
current bundle and observe the existing project without image pull/load, service creation, `exec`, configuration
reload or database mutation.

For that healthy-current observation, enumerate the exact seven non-one-off `cloud-clicker` service containers. Require one distinct
full container ID per service, canonical project/service ownership, and image-config ID equality
to the appropriate manifest entry. A registry tag, cached image inspection alone, name alone,
or candidate manifest cannot supply this proof. Gameserver/Caddy/Postgres must be running;
other service states are recorded and do not substitute for the parent's operations acceptance.

This healthy-current test also verifies the preserved prior deployment at a
pre-backup abort's conclusion and the restored target at successful rollback.
It is NOT rollback admission: a failed source may have missing/stopped or mixed
source/target containers. Recovery instead observes ownership against the exact
governing source/target tuple and durable resource intents/bindings. Absence is
recorded, not treated as a healthy service or invented ownership. Existing
resources need exact admitted IDs and allowed bundle/config bindings; registered
one-off jobs are checked through their operation ownership, not mistaken for
extra steady-state service containers. Third/unowned resources refuse cleanup.
Failure to pass the full healthy-current observation keeps ordinary release and
success unavailable; it does not by itself forbid governed cleanup of owned
failed resources. Exact recovery-observer/proof producers remain CR6 gates.

Read these existing release-owned paths from those identified containers:

| Service | Container path | Validated bundle source |
|---|---|---|
| caddy | `/etc/caddy/Caddyfile`, `/srv` | `Caddyfile`, complete `site/` tree |
| backup | `/opt/cloud-clicker/deployment-backup`, `/opt/cloud-clicker/release-manifest.json`, `/opt/cloud-clicker/content` | `deployment-backup`, original manifest, complete `content/` tree |
| prometheus | `/etc/prometheus/prometheus.yml`, `/etc/prometheus/cloud-clicker-alerts.yml`, `/opt/cloud-clicker/deployment-operations` | the corresponding two `operations/` files and helper |
| alertmanager | `/opt/cloud-clicker/deployment-operations` | the helper |

Use a read-only container archive stream, not host-side bind-path assertions. Compare exact
file sets, lengths and SHA256 against the validated bundle; original manifest bytes are checked
against `manifest_sha256`. Missing/extra/repeated paths, links, traversal, changed body lengths,
read/process errors or incomplete streams refuse. Stream rather than buffer an unbounded archive;
content limits derive from the validated files. No fallback to an unmounted directory or file count.
Do not use container export: it excludes mounted volume contents. Docker's archive/inspection
mechanisms still need execution on the supported pinned engine; documentation is not that proof.
([copy](https://docs.docker.com/reference/cli/docker/container/cp/),
[export](https://docs.docker.com/reference/cli/docker/container/export/))

Verify the actual command/entrypoint and read-only mount destinations against the validated
Compose/image configuration, including absence of mounts shadowing the gameserver binary/content.
Inspect only allowed fields. Never emit whole environment/secret/receiver configurations. Reinspect
container IDs, image IDs and start identities after reads; replacement/restart during observation
invalidates it. Use the existing cancellation contract, not automatic retries or a new time budget.
([inspection](https://docs.docker.com/reference/cli/docker/container/inspect/),
[project/service ownership](https://docs.docker.com/reference/compose-file/services/#labels))

The evidence blob has exactly `schema_version:1`, `server_id`, `identity`, `observed_start`,
`observed_end`, `containers` and `files`. `identity` is CR1's object. `server_id` is the configured
UUID, matched to the selected gameserver's sole `CLOUD_CLICKER_SERVER_ID` value through an
allowlisted projection, never a whole environment dump. Add the same required non-secret
`ServerID` to the controller, using the CLI's existing configured value.

`containers` has seven service-ordered exact `{service, container_id, image_config_sha256,
running, started_at, config_sha256}` records. IDs are64 lowercase hex; config hashes bind the
compact JSON plus newline of the exact `{entrypoint, command, read_only_rootfs, mounts}`
projection checked above. Entrypoint/command are ordered arrays (nil normalized to empty);
mounts are destination-ordered `{destination, type, read_only}` objects, with no source path.
The observer compares that projection against the validated Compose configuration and pinned
image defaults before emitting its hash. `started_at` is null only for a
nonrunning container that has never started; otherwise nonzero UTC RFC3339Nano. `files` is the
sorted exact `{service, bundle_path, sha256}` array described above. Observation dates are
nonzero UTC RFC3339Nano; start must not exceed end. Only declared relative bundle paths appear,
never raw mount-source/secret paths. Reject extra/duplicate fields and all unintended nulls.
Store its compact JSON plus newline under private `release-bindings/<sha256-hex>.json`, syncing
before the referring ledger event. Hash and decode it on reuse. Container copy output is not stored.

This proves a supported-helper observation, not an atomic snapshot against a privileged manual
administrator or a replacement Docker daemon. It does not waive real HTTP/WebSocket, data identity,
network, secret or loaded-service behavior checks in the parent/R-006.

### CR3 — Durable events in the existing ledger

Extend `release-ledger.jsonl` with v2 release/rollback events; never rewrite v1 rows or introduce
a second current pointer. Initial installation retains its current v1 terminal record and clean-host
rules, before any v2 history. New release/rollback writers emit only v2. Existing valid v1 prefixes
remain readable with their existing grammar; after the first v2 event, a v1 append is invalid.
A v1 successful terminal row can anchor
CR1's identity only after exact bundle/image-reference matching and CR2; it is
not a whole-helper completion proof. Before any new protected work, CR6's
predecessor fence is mandatory even when the preceding history is v1 or Finished
success. The governed first-install/v1 handoff proof producer remains an explicit
acceptance blocker; this draft does not grandfather an unlocked historical helper.
Legacy failed post-backup release
or rollback is uncertain, not silently promoted. No-success legacy history cannot authorize release.

Every v2 event has exactly these decoded keys; reject aliases, duplicates, unknown/missing keys
and unintended nulls. Fields nullable below must still be present.

| Field | Rule |
|---|---|
| `schema_version` | literal2 |
| `event` | `operation_started`, `backup_ready`, `recovery_fenced`, `operation_finished` |
| `operation_id` | canonical UUIDv4, new per operation, never reused |
| `action` | `release`, `rollback` |
| `server_id`, `operator` | configured canonical server UUID and existing operator-ID grammar |
| `occurred_at` | nonzero UTC RFC3339Nano, nondecreasing event order; an ordinary new operation starts strictly after the previous completion |
| `source`, `target` | non-null identity objects |
| `backup_id` | existing backup-ID grammar or null under the event rules below; no stored path |
| `rollback_until` | UTC date or null under the event rules below |
| `outcome`, `failure_stage` | null until finish; `succeeded`/`failed` and exact failure stage at finish |
| `source_proof_sha256`, `target_proof_sha256` | SHA256-prefixed blob hashes or null under the rules below |
| `recovery_of` | null for release; the immediate governing predecessor's operation UUID for rollback/retry, under CR4's ancestry rules |
| `predecessor_fence_sha256` | CR6 proof hash; non-null on ordinary release Started and repeated thereafter; rollback null at Started, non-null from recovery_fenced onward |
| `completion_proof_sha256` | CR6 whole-operation closure proof hash at Finished when closure is established; null on earlier events and permitted uncertain post-checkpoint failures/interrupted-rollback administrative finishes |
| `supervision_sha256` | required non-null hash of the exact operation's Supervisor object under the proposed Operation Evidence contract; immutable across its events |

The first v2 row follows the latest v1 completion strictly. Validate the complete prefix before
effects; all subsequent rows OF THAT OPERATION keep operation/server/operator/source/target identical. Permit only
one active transition chain: an unfinished release can have only its CR4 governed rollback as
successor, never a competing release or two recovery attempts. Hold the existing operator lock
through observation, writes and effects; append each event and fsync it before crossing its
protected boundary.

- **Started:** release has known source and its source proof, with null backup/deadline/target
  proof. Rollback has the source/target and backup
  of its governing release/checkpoint, null deadline/identity proofs and non-null
  recovery_of; a damaged source need not look healthy. Rollback Started authorizes
  only CR4's bounded fencing cleanup, not reset/restore/replacement startup.
- **Backup ready:** release only, once, after actual host-byte backup validation. Carry its exact
  ID and pending deadline `occurred_at +7days`; target proof/outcome/stage remain null. No drain,
  candidate startup or database change may precede this durable row. An unrecorded produced
  backup remains protected under existing policy, not automatically resolved.
- **Recovery fenced:** rollback only, exactly once after Started and CR4's cleanup,
  before reset/restore/replacement startup. It has non-null predecessor_fence_sha256,
  the same ancestry/identity/backup tuple, null deadline/outcome/stage/target proof/
  completion proof. This proves predecessor closure, not completion of this recovery.
- **Finished:** once, after the real outcome. Success requires the existing identity/smoke gates
  and target binding proof AND non-null completion_proof_sha256. A successful
  rollback also requires its earlier recovery_fenced row. A pre-backup failure
  requires a completion proof and preserved-prior identity before Finished can
  be appended; absent those, it stays unfinished and needs the governed abort
  or separate recovery route. A post-checkpoint failed finish
  without completion proof reports failure only and cannot authorize a successor's
  protected work. Successful release requires its checkpoint and repeats its backup ID,
  with completed-release deadline `occurred_at +7days`. A failed post-checkpoint release has the
  same terminal seven-day origin as current completed failures. Pre-checkpoint failure has null
  backup/deadline. Rollback has null deadline and repeats its exact backup ID.
  Failure has no target proof and uses the existing stage vocabulary plus CR4's interruption stage.

The source-proof hash repeats unchanged on every event of its release; it is null throughout
rollback. All outcomes/stages are null before Finished. Finished success has null failure stage;
Finished failure has a non-null stage appropriate to its action and reached phase. Before Backup
ready a release may fail only at supply_chain/preflight/preupgrade_backup or CR4's interruption;
after it, only bounded_drain/startup_migration/epoch_artifact_identity/authenticated_smoke.
Rollback permits its existing post-admission execution stages plus
`recovery_fence` for failed cleanup before recovery_fenced and
`interrupted_rollback` for CR4's administrative failed closure. No success is permitted without
the referenced target/completion proofs and actual parent identity/smoke gates.
All rows repeat recovery_of unchanged. Release predecessor_fence_sha256 is
unchanged; rollback changes it only once, null to its recovery_fenced value,
then repeats it. Completion proofs are never inferred from a failure result.

Rejected requests before a valid Started row return a typed error without manufacturing a tuple
or an operation record. Failed writes remain failures; no success, rollback authority, or backup
resolution may be inferred from a partly written/unsynced event. Final event failure leaves the
attempt unfinished even if services started. Unknown clocks fail rather than inventing dates.

### CR4 — Explicit interrupted-operation recovery

No ordinary install/release starts while an operation is unfinished or current state uncertain.
Do not automatically resume, retry, finish success, or treat an unlocked lockfile as completion.

Manual `release-abort --operation-id` is permitted only for an unfinished release without Backup
ready. Read-only admission first validates the exact operation and its surviving
supervisor or a proved sealed handoff. It may then seal worker admission and
stop ONLY proven operation-owned jobs/connections/output writers. This cleanup
may not stop/recreate the live game, Caddy or Postgres, remove persistent data,
or modify gameplay state. Unknown ownership refuses without widening cleanup.
After that cleanup, revalidate exact prior identity via CR2 and current database/epoch identity before appending
Finished failed with `failure_stage:"interrupted_before_backup"`; no running-service/data mutation
(the parent's read-only one-off database inspector is permitted). If
identity cannot be established, refuse and require the separately governed recovery route.
The inspector is itself tracked work and must finish before the CR6 completion
proof; it cannot reopen general worker admission after the fence is sealed.
Interrupted initial install or rollback is not covered by this abort command; it remains explicit
recovery/rollback work, not authority for a clean-install bypass. Initial installation remains
the parent's v1/clean-host workflow; this amendment creates no installation pending event.

Manual rollback may consume a pending release's durable Backup ready row. Match operation,
candidate/previous identities, server and backup ID; check the pending seven-day deadline from
that checkpoint. Preserve the completed-release seven-day origin for finished attempts. Pending
recovery never fabricates a later completion merely to extend its window. Parse/authenticate
the backup and restore key before stop/reset, then use the existing exact clean-volume restore,
forward-only startup, identity/smoke and target-binding sequence. Source containers may be absent
or a mixture of the exact source/target attempt; a third/unowned release refuses before teardown.

A rollback Start event references the pending/failed release through its source/target/backup
tuple; it is the only allowed successor to that recoverable unfinished release. A successful
rollback gives its governing release a DERIVED `recovered` disposition; it does
not fabricate that release's missing Finished row, mark it succeeded, or rewrite
an existing failed result. The successful rollback retains its own complete
event chain. CR5 readers must recognize this precise ancestral closure.
A failed/interrupted rollback permits only exact governed retry.
An unfinished rollback must first be explicitly closed as failed `interrupted_rollback` by that
retry command, without claiming service recovery or changing its release deadline; never create
two live recovery attempts. Its new attempt keeps the same tuple and receives a new operation ID.
Every rollback records recovery_of: the governing release for its first attempt,
or the immediately preceding failed rollback for a retry. Following that ancestry
must reach exactly one checkpointed release; source is that release's target,
target is its source, and backup/deadline stay bound to THAT release. Reject
cycles, unrelated parents, changed tuple/backup, overlapping recovery children
and a second successful recovery of the same chain. A failed/interrupted child
does not recover its ancestors or create a fresh seven-day window.

Recovery has FOUR distinct permissions, in order:

1. Read-only admission establishes ancestry, exact source/target/resource
   ownership and the existing authenticated backup/key authority. It does not
   stop services or create a replacement. A live lock-owning supervisor must
   accept the authenticated manual request itself, or complete a proved sealed
   handoff; a new CLI cannot recover by killing it merely to reacquire its lock.
2. Durably append rollback Started, then seal old worker admission and perform
   bounded fencing cleanup of PROVEN owned work. After a valid checkpoint/input
   authority this may stop the failed application/database namespace while
   retaining persistent volume bytes. It may not reset/delete the database
   volume, restore data, or start replacement services. Waiting old requests
   and native jobs remain tracked until actual terminal outcomes.
3. Prove the predecessor fence under CR6 and durably append recovery_fenced.
   No reset/restore/replacement startup is lawful before this boundary. A
   Started row, client exit, delivered cancellation or empty census is not it.
4. Execute the existing clean-volume reset/restore/forward-only startup and
   identity/smoke sequence, then establish THIS recovery's completion proof
   before Finished. Predecessor closure does not prove current-helper closure.

Before protected successor work, prove prior helper/client actions quiescent;
the narrow cleanup in step2 is how that fence is established, not permission to
skip it. Shared lock acquisition alone is insufficient. The supported-engine experiment below must settle this mechanism before
acceptance. No PID/lock-file heuristic or two stable metadata samples may substitute for that proof.
Killing the helper and its Docker CLI is also insufficient: the bounded daemon experiment below
observes an already-started command completing afterward. Recovery must account for daemon-side
work, not merely prove the local processes exited. The actual quiescence mechanism remains open.
This also holds for the shipped one-off command shape: killing an exclusively owned local
Docker/Compose process group does not stop an already-running `compose run --rm --no-deps`
container. The controlled job below completes only after its gate is opened after CLI death.
Process-group cancellation alone therefore cannot authorize recovery. This is not a real
backup/restore interruption or proof about every lifecycle command.
An observed terminal Docker ExecID alone is also insufficient as a general fence: the bounded
diagnostic below observes successful parent completion/restart refusal before an explicitly
gated descendant later writes. A proposed mechanism must account for the actual operation's
descendant/process and outstanding-request lifetimes, not silently equate parent exit with
that guarantee. This diagnostic does not show a shipped backup command leaking descendants.
Removing the exact one-off container is insufficient for database work already
admitted by Postgres too. RP-726's real pg_restore COPY waits in its identified
backend; client-container removal succeeds, yet opening the database gate then
commits the row. CR4 must account for outstanding database sessions/statements
as well as Engine requests and client descendants. The controlled separate
backend-termination arm prevents that row, but earlier restored DDL remains:
session termination is neither rollback nor a complete operation fence. No
unscoped backend termination, application-name-only authority, new database
permission or recovery implementation is authorized by this observation.
The separate [pre-backup observation](../planning/deployment-foundation/prebackup-live-database-observation.v2.json)
on pinned PostgreSQL16.15 also finds a live startup backend absent from
pg_stat_activity, surviving exact client removal temporarily. An empty session
census cannot establish physical quiescence. In the same disposable fixture,
separately identified pg_dump backend termination preserves the exact running
database and continuously connected control. This is a SQL-side candidate only:
the real operation's durable session ownership, startup completion, pending
requests and generation-safe targeting remain unresolved. Matching backend_start
in a query does not make PID-only termination atomic against PID reuse. No
production cancellation policy or database privilege follows from the probe.
The subsequent [connection-owned gateway candidate](../planning/deployment-foundation/combined-fence-candidate.md#connection-owned-cancellation-candidate)
tracks an actual libpq connection before dialing, uses its BackendKeyData only
in memory and distinguishes cancellation delivery from original-session EOF.
Wrong-key/correct-key/startup controls support that narrow candidate without a
PID-only signal. They do not bind every actual helper connection or establish
durable supervisor ownership. CR4 remains unresolved until the full operation
protocol/packaging, lost-authority refusal and supported-host proof are complete;
no gateway or supervisor is adopted/implemented by this evidence paragraph.
The network-generation experiment below supplies a candidate boundary for NEW
database clients: a worker bound to a retired full network ID fails to start even
when that network's human-readable name is recreated; a name-only worker reaches
the replacement network and writes. Any candidate using network retirement must
therefore bind the actual immutable ID before submission rather than resolve its
name again afterward. This does not establish that the current Compose commands
provide that binding. It also does not fence an existing database backend, host
file/volume mutations, or every outstanding Engine request. The full mechanism
remains open; do not substitute network removal for the required proof.
The source-bound [mutation map](../planning/deployment-foundation/interrupted-transition-mutation-map.md)
now enumerates the existing Controller release/rollback paths. It explicitly
includes mutating HTTP bootstrap in drain/smoke, four configured networks, named
volumes, startup migrations/jobs, backup/operations host writers and scheduled
temporary/backup DELETION. It also includes runRelease's outer outcome metric,
published after Controller returns and outside its operator-lock lifetime;
runRollback has no corresponding wrapper write. A terminal Controller ledger row
therefore cannot by itself prove the whole helper finished publishing. A combined
candidate must supply a lifetime or isolation argument for every in-scope row;
SQL-session termination plus one retired network is not whole-operation proof.
The map distinguishes upstream Compose's ID-addressed stop/remove calls from
creation/volume name handling; no claim about queued daemon execution follows
from a top-level CLI command string or source inspection alone.
The [combined candidate](../planning/deployment-foundation/combined-fence-candidate.md)
now executes contained clients, actual admitted pg_restore work, controlled
HTTP/host writers and delayed create/start/volume-delete in ONE three-population
experiment. The generation population preserves all tested replacement values;
parent-only and reused-name controls do not. This is not the actual release
helper, encrypted restore or four-network/five-volume production topology.
It also allows late old-generation create/delete effects: that is isolation,
NOT literal quiescence. An isolation-based production proposal must explicitly
reconcile this contract with full ownership/durable ordering and review/adoption;
the experiment does not silently satisfy or waive the present quiescence clause.
The candidate table names the remaining real bindings, including the pre-backup
abort whose no-service/data-mutation rule cannot use the prototype's DB teardown.
No new backup decryption placement, protection-metadata resolution or deletion policy is adopted.

### CR5 — Consumers and compatibility

One decoder/reducer in `deploymentrelease` serves the controller, CLI and rehearsal readers.
R-006 consumes full v2 events and binding blobs, reducing them to the exact install→release→rollback
operations; it must reject missing checkpoints/fences/completion proofs, identity
substitution, duplicate/unclosed operations and byte corruption. A pending parent
release is closed only by CR4's exact successful descendant recovery rule, not
by ignoring an arbitrary unclosed row. Failed/interrupted recovery never closes
that parent. Three terminal summaries alone are not v2 evidence.

The exact candidate and predecessor helper artifacts used for this amended rehearsal must both
support v2 reading/writing and the existing security floor. Do not claim pre-RP-556 helpers fail
closed before effects or copy current helpers into a historical bundle without rebuilding its
manifest/provenance. Unsupported helper combinations refuse; legacy journal bytes are preserved.
This requirement does not settle the other previous-bundle-floor or producer-identity questions.

### CR6 — Whole-operation closure and supervisor authority

CR1's identity result and the operator lock are not whole-helper completion.
Before EVERY successor's protected work, validate a predecessor fence bound to
the exact accepted ledger prefix, identity/operation ancestry and supervised
resource generation. Ordinary release needs it before Started/runtime preparation;
governed rollback obtains it through CR4 before reset/restore/replacement startup.
No successful Finished, apparent healthy deployment or v1 terminal record waives
this gate. Initial install/v1 handoff needs a real governed producer, not an
operator assertion; until specified/proved, that bootstrap remains blocked.

The supervisor must outlive the invoking CLI and own the shared operator lock,
controller, command descendants, admitted daemon/native jobs, operation SQL/HTTP
streams and host publications/deletions. The runtime map includes the OUTER
CLI outcome metric: publish it before the final completion proof/Finished, or
include its whole lifetime in that proof. No protected writer remains outside
the supervised scope just because Controller returned. Long-lived source/target
services must have an explicit verified ownership disposition; they cannot be
silently counted as vanished jobs or excluded unknown writers.

Close worker admission monotonically before cleanup. Keep the authority that
knows outstanding requests alive while collecting their actual terminal results,
stopping/reaping owned jobs and finishing trusted read-only verification.
Every side-effecting request needs durable intent BEFORE forwarding and an
endpoint-specific terminal receipt AFTER completion. Native unit start/stop jobs
are also asynchronous work if that packaging is selected. Native worker cgroup
emptiness alone cannot erase queued native starts, Docker work or database
streams. Prevent bypass of Engine/SQL/HTTP admission through real isolation;
environment variables or a private socket name alone are not that guarantee.
Secrets and cancellation keys never enter intents, receipts, logs or proof blobs.

A predecessor fence and a completion proof must be separately typed, hash-bound
private artifacts. Readers must verify exact producer/supervisor identity,
covered ledger prefix/operations, sealed admission, complete work-log/receipt
bindings, terminal operation actors/streams and explicit retained-service/output
ownership. Neither counts, self-reported booleans, a label nor file presence is
sufficient. Once completion is sealed, only its final ledger append and inert
control-plane teardown may follow; no old metric writer or delayed protected
request can run afterward. The proof is synced before its referring event.
Its covered-prefix hash ends immediately BEFORE that referring event; it never
hashes an event containing its own hash. Reusing an earlier completion proof at
a later admission needs a verified predecessor-fence binding to the current
prefix, not a circular hash or an unbound sidecar. Prefix changes, missing work
receipts and any later unsealed protected intent refuse.

**Not acceptance-ready:** the exact artifact/request-log schemas, native profile,
first-install/v1 handoff and supported-host producers/readers still need complete
specification and independent execution. These semantic fields are not permission
to insert placeholder proofs, a no-op verifier, a fake certificate or a tracking
validator as implementation. The existing prototype artifacts are research,
never these new proof types. No new resident service, native-unit prerequisite
or privilege is adopted by naming the supervisor. Selecting such packaging must
reconcile this draft's no-service promise and the parent's host floor explicitly.
The proposed [Operation Evidence](deployment-operation-evidence.md) specifies
the descriptor, phase logs, receipt envelope, closure and install-handoff shapes.
Its OP9 native plan/subject/receipt producers and profile are STILL required;
the proposed shapes are neither accepted authority nor a completed supervisor.
Rollback fencing and subsequent execution use separate sealed phase logs, so
its predecessor proof does not prematurely seal the execution it authorizes.

If the supervisor/gateway or request authority is lost before a valid seal,
missing processes/sockets or an unlocked file do NOT reconstruct it. Refuse
in-place successor mutation, preserve pending/uncertain history and backup
protection, and require the separately governed recovery route. A retained valid
seal may support a proved handoff; a partially written artifact cannot. This is
not automatic rollback, permission to reboot a host/daemon, or a disaster-recovery
implementation. Clean-host/RPO/RTO and actual useful recovery paths remain gates;
universal refusal cannot pass the successful-abort/rollback populations.

## Deviations and adoption

No gameplay, infrastructure provider, retention bound, normal seven-day window or live Down
migration changes. Proposed operational changes are explicit: recorded/observed current identity
becomes mandatory; history-less ordinary release is removed; ambiguous post-checkpoint failure
requires rollback; and v2 pending/checkpoint records add manual interruption recovery. These are
not implemented improvements presumed approved because they are safer.

Adoption must reconcile the parent's DP5 body, acceptance and operator instructions in one
normative range with designated review. Keep D-019, protection metadata, other DESIGN-GAPs,
independent implementation reviews and all existing clean-host populations open.

## Acceptance criteria

1. Controller regressions refuse contradictory version/manifest/image tuples before runtime
   preparation or backup. Empty/no-success history, uncertain states and missing verifier refuse;
   valid install/release/rollback and pre-checkpoint-failure retry still work.
2. On the supported pinned Linux engine, observe all seven actual containers and mounted bytes.
   Wrong/duplicate/one-off containers, a correct image with wrong mounted site/helper/manifest,
   command shadowing, restart/replacement and broken/truncated copy streams refuse. Seeded secret
   environment values and malformed metadata never enter output/logs. No host-path-only fallback.
3. Actual process-kill/file-write faults before/after Started, backup checkpoint, drain/start and
   Finished demonstrate the precise journal state and supported manual recovery. Before-checkpoint
   abort preserves data; checkpointed rollback restores actual database identities. Lost final
   append never becomes a completed release. Corrupt/wrong-key backups refuse before destruction.
4. V1→v2 fixtures preserve original bytes; every decoder rejects reordered/duplicate events,
   substituted proofs, reused operation IDs and fabricated windows. Candidate/predecessor helpers
   both read real evidence; neither a fake basename nor protocol self-report proves identity.
5. R-006 runs the exact reviewed pair without checkout/providers, including a stale-current
   refusal, wrong mounted artifact and interrupted release recovery. Retain original AC1–AC10,
   RPO/RTO and manual/publication authority. Designated review covers code/tests/docs/records;
   no acceptance from these draft criteria or an isolated unit pass.
6. Real supervision/reducer populations prove all ordering branches: an ordinary
   successor refuses a finished-but-unsealed helper; a valid preserved-prior
   abort completes without live-service/data mutation; a partial/mixed failed
   source admits ONLY governed cleanup, then refuses reset until recovery_fenced;
   correct rollback success derives ancestral recovery without fabricating a
   parent Finished row. Failed recovery, wrong ancestry, changed deadline and
   lost authority do not recover that parent. A late outer metric, queued native
   start, Docker request or SQL stream must prevent the appropriate proof.
   First-install/v1 handoff and successful real paths are mandatory, alongside
   all refusal controls; a paper trace or a no-op verifier is not execution.

## Open questions

The proposal selects a concrete policy for review rather than asking Marco to choose data shapes.
It is **not acceptance-ready**. Adoption/designated normative review and these bounded mechanical
questions remain; the author must reconcile their results into CR2/CR4 before acceptance:

- **Mounted-byte reader:** on the supported pinned engine, run the same observer against all nine
  listed mounted roots. Intact controls must return exact sets/hashes; one changed site/helper/
  manifest, missing path, link and deliberately interrupted archive must refuse. No truncation,
  host-only substitute or stdout capture of secrets. Failure means select and specify another
  real-container reader, not waive mounted-byte binding.
- **Interrupted-command quiescence:** compare a normally completed helper operation, helper
  termination while an owned Docker client is active, and termination after a Docker request
  starts but before its response. Use an owned Linux test project, real child processes and
  real daemon observations. The recovery command must either establish that no earlier request
  can mutate protected successor state afterward, or refuse that successor before
  reset/restore/replacement startup. CR4's explicitly admitted fencing cleanup is
  permitted to establish this boundary; it is not permission for successor data
  mutation or unowned teardown. Test delayed
  old-request completion after the first stable observation as the negative control. Predeclare
  the actual supported Engine/Compose pair and retain process/request ordering, not an arbitrary
  sleep. Results authorize the exact supervisor/recovery mechanism, never automatic rollback or
  a changed deployment window. If that mechanism cannot be established, this proposal must change.
  Include already-admitted Postgres work: job-container removal alone cannot fence
  that work (RP-726). Bind every operation-owned connection before mutation and
  establish its terminal outcome, while also preventing late client/Engine actions
  from creating replacement work. Exact recovery ownership/permissions and the
  complete mechanism remain to be specified and proved; the experiment's synthetic
  application tag/PID selector is not that production protocol.
  The network-generation candidate also requires a complete mutation-path mapping:
  pending creates/starts, project-wide Compose operations, volume/name reuse and
  host-mounted outputs cannot silently inherit the SQL worker's narrower result.
  Persisting a generation label or full network ID alone is not this mapping.
  The source inventory is now retained in the mutation map; the unresolved task
  is the COMBINED mechanism and its execution, not another inventory of the same
  root calls. Include admitted mutating application HTTP, post-controller CLI
  metrics, every helper's host writer and scheduled deletion of retained material.
  Separately name any excluded install/rehearsal/rotation path rather than
  silently extending an ordinary release/rollback proof to it.
  The controlled combined prototype now exists; do not repeat its complete
  populations in place of the durable real-helper protocol. Resolve isolation
  versus literal quiescence explicitly, pre-backup non-destructive abort,
  packaged supervisor, all production resources and checkpoint authority before
  promoting the mechanism. Native supported-host execution remains required.
- **Closure/protocol producers:** CR6 now closes the normative ordering gaps, not
  its own construction gates. Specify exact fence/completion/work-log schemas,
  initial-install/v1 handoff, supervisor ownership/lock attachment, non-bypassable
  admission and native job receipts. Prove actual successful completion as well
  as lost-authority refusal; neither the experimental booleans nor a universally
  refusing implementation can satisfy this contract. If native services are
  selected, reconcile both host prerequisites and the no-service promise before
  adoption, not after product code is written.

Docker documentation/source inspection is partial support for those instruments, not a runtime
witness. Parent holds remain separate; no claim that every R-006 population has a producer.

Native precursor evidence: the preserved [manual source](../planning/deployment-foundation/probes/child-lifetime-probe.go.txt)
calls the actual ExecRunner under a real operator lock. On darwin/arm64, the child writes its
completion marker after the parent is killed and the lock is reacquired; the intact-parent
control completes normally. Both owned child processes are observed terminal. This directly
rejects lock-only quiescence, but supplies neither a Linux supervision mechanism nor a daemon
request witness. It is not a normal CI assertion that child survival is desirable.

Bounded Docker evidence, 2026-10-15 (not native supported-host closure): Engine28.4.0,
Compose2.39.4-desktop.1, Darwin/arm64 CLI and Linux/arm64 Desktop daemon. Exact pinned
linux/amd64 image configurations and validated retained source102face6 bundle, NOT a current
release build. The [mounted archive diagnostic](../planning/deployment-foundation/probes/mounted-archive-probe.go.txt)
reads all nine roots from four isolated stopped containers. Intact sets/lengths/hashes pass;
changed site/helper/manifest, missing child, symlink and actual killed-copy unexpected EOF
refuse. Host copies only supply mounts/expectations; observations come from real container
archive streams. This supports the reader approach, not complete seven-container ownership,
command/start-identity/secret exclusion or native clean-host acceptance. No production reader
is implemented; missing/extra/duplicate/traversal and framing admission remain production gates.

The separate [daemon lifetime diagnostic](../planning/deployment-foundation/probes/daemon-lifetime-probe.go.txt)
uses actual ExecRunner/operator lock and a real daemon-side exec in an owned no-network
fixture. Intact helper completes normally. In the killed case, both helper and owned CLI are
observed terminal and the lock reacquired while the work gate is still closed; opening it then
produces the daemon-side completion. That process is separately observed terminal. This rejects
client cancellation as a quiescence proof, including stable pre-gate observations, but does not
establish a safe Linux supervisor/recovery mechanism. The proposal remains not acceptance-ready;
do not substitute restart/retry or an operator assertion. Exact commands/limits are in the log.

Further bounded evidence, 2026-10-16: the [manual ExecID diagnostic](../planning/deployment-foundation/probes/exec-terminal-probe.go.txt)
uses the same exact-image/no-network Desktop fixture and actual versioned Engine API1.51.
An ordinary direct command completes normally. A shell that forks one deliberately gated child
also reaches Running=false/ExitCode0, and restarting that same ExecID refuses409, yet opening
the child's gate afterward produces its write. The created-but-not-started state also reports
Running=false/ExitCode0, with Pid0 rather than the terminal parent's retained nonzero PID.
Those states must not be conflated. No helper-death/request-queue experiment or actual
pg_dump/Compose descendant claim is made by this narrower diagnostic; it rejects the generic
ExecID-only inference, not every operation-specific mechanism. No supported native-host,
complete lifecycle fence, production recovery implementation or adoption follows.
Docker's documented [create/start/inspect endpoints](https://docs.docker.com/reference/api/engine/version/v1.51/)
and pinned [exec implementation](https://github.com/moby/moby/blob/v28.4.0/daemon/exec.go)
are partial source support, not substitutes for the runtime observation.

The [actual Compose one-off diagnostic](../planning/deployment-foundation/probes/compose-job-lifetime-probe.go.txt)
uses the runtime's `compose run --rm --no-deps` shape, with an explicit unique observation
name, on the same pinned Desktop Engine/Compose pair. Both ordinary completion and killing
the entire exclusively owned CLI process group execute. In the killed case, parent Wait is
terminal while the identified daemon job remains running behind a closed gate. Opening that
gate then produces its completion; both populations automatically remove their containers.
The source inventory matters: current backup/validation/identity/alert jobs use Compose one-offs,
not Docker exec. This is a controlled shell payload, not pg_dump/restore, native supported-host
proof, a delayed-request fence or evidence of a complete safe recovery mechanism. It excludes
process-group cancellation as the sole solution without inventing an accepted replacement.

The [restore-backend diagnostic](../planning/deployment-foundation/probes/restore-backend-fence-probe.py.txt)
adds a real database boundary on the same Desktop pair, using the exact amd64
child of the pinned Postgres16 image and three disposable databases. An actual
custom dump contains a synthetic CHECK function whose transaction advisory lock
holds COPY after the target connection/DDL are established. Every arm observes
the exact COPY/advisory wait and zero committed rows before intervention. Normal
completion restores row7. Removing the exact owned client leaves its backend
alive; releasing the lock afterward still commits row7. Removing the client and
then terminating that observed backend leaves zero rows, while the table created
before COPY remains. Completion follows backend disappearance and owned-project
cleanup, not a sleeping absence check. Initial catalogue-lock attempts are invalid
and excluded in the owning log; no guard was raised.

This refutes a client-container-only fence and narrows the proposed mechanism.
The workload is controlled SQL content through actual pg_dump/pg_restore, not
the encrypted application helper, game dataset, native supported host or complete
pending-request/recovery proof. PostgreSQL16's documented optional
[connection checks](https://www.postgresql.org/docs/16/runtime-config-connection.html#GUC-CLIENT-CONNECTION-CHECK-INTERVAL)
and [backend termination semantics](https://www.postgresql.org/docs/16/functions-admin.html#FUNCTIONS-ADMIN-SIGNAL)
are consistent with the observation, not substitutes for it. CR4/adoption remain open.

The [network-generation experiment](../planning/deployment-foundation/probes/network-generation-fence-probe.py.txt)
and [source-bound observation](../planning/deployment-foundation/network-generation-observation.v1.json)
execute a constructive candidate on that same Desktop pair and exact amd64
Postgres image. An intact ID-bound SQL worker writes row11. A second worker is
created but not started with the full original network ID. After disconnecting
only the owned database, network removal succeeds; the same network NAME is then
recreated with a different ID and the same database reconnected. Starting the old
worker fails with the old-network-not-found error. Creating another worker with
the retired ID succeeds, but its start fails too. A name-only control starts and
writes row14; final exact rows are11/14, with no12/13. Thus creation success is
not startup or connectivity, and name reuse is not immutable identity.

Commands and SQL outcomes are sequentially ordered, not inferred from a sleep.
The workers directly invoke `psql`; this is no delayed request already inside the
daemon, killed helper/Compose process, pg_restore or full lifecycle experiment.
Earlier RP-726 establishes the separate already-admitted backend problem. All
owned resources are removed before complete output. Docker's documented
[disconnect](https://docs.docker.com/reference/cli/docker/network/disconnect/)
and [removal](https://docs.docker.com/reference/cli/docker/network/rm/) operations
support the fixture setup, not a general pending-request fence. CR4's complete
mechanism, native-host proof and adoption remain open.

## Changelog

- 2026-10-13: source-grounded RP-558/DESIGN-GAP10 proposal. No product implementation, owner
  adoption, acceptance, archival, push or deployment.
- 2026-10-15: retain bounded Desktop mounted-reader and daemon-lifetime evidence; CR4 explicitly
  rejects helper/CLI cancellation as proof. Native-host mechanism/adoption remain open.
- 2026-10-16: actual ExecID parent completion still permits a gated descendant write. Reject
  generic ExecID-only quiescence; actual-operation/native-host proof and adoption remain open.
- 2026-10-16: actual Compose one-off survives whole local CLI process-group termination;
  reject process-group-only recovery fencing. Controlled Desktop result, not backup acceptance.
- 2026-11-04: RP-726 actual pg_restore COPY commits after its client container is
  removed. Require database-side quiescence too; identified backend termination is
  a bounded positive control, not a complete recovery mechanism or adoption.
- 2026-11-04: controlled old-ID pending/late worker starts refuse after network
  retirement, while the reused-name control writes. Retain exact source/rows;
  distinguish this candidate admission boundary from a complete operation fence.
- 2026-11-04: retain the source-bound whole release/rollback mutation map,
  including HTTP bootstrap and host writers. Complete mechanism/proof remains
  open; upstream ID/name distinctions are source support, not daemon execution.
- 2026-11-04: execute a combined controlled candidate with two discriminating
  weaker controls. Retain actual after-admission mutations/absence and explicitly
  distinguish generation isolation from this draft's unfulfilled quiescence rule.
- 2026-11-04: independent review identifies four normative ordering/closure
  contradictions. Scope healthy-current admission, require whole-helper proof,
  separate admitted fencing cleanup from destructive execution, and define
  explicit recovery ancestry/derived parent closure. Add semantic proof fields
  and recovery_fenced; concrete schemas/first-install handoff/supervisor/native
  producers remain explicit acceptance blockers. This is not implementation,
  owner adoption, host-prerequisite selection or recovery acceptance.
