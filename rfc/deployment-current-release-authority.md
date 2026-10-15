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

Pre-backup failed release attempts leave the prior identity known. Once a release has reached
the durable backup checkpoint, a failure makes ordinary release inadmissible until governed
rollback succeeds. Failed rollback remains retryable against the exact same authority; successful
rollback closes that authority and establishes its target as current. No newer/unrelated attempt
may reuse an older backup chain. Installation/recovery prerequisites remain the parent's gates.

### CR2 — Observe the containers and release-owned bytes

Add a mandatory current-observation capability to the release runtime. Absence is an admission
failure, never a no-op adapter. Before accepting an operation, validate the current bundle and
observe the existing project without image pull/load, service creation, `exec`, configuration
reload or database mutation.

Enumerate the exact seven non-one-off `cloud-clicker` service containers. Require one distinct
full container ID per service, canonical project/service ownership, and image-config ID equality
to the appropriate manifest entry. A registry tag, cached image inspection alone, name alone,
or candidate manifest cannot supply this proof. Gameserver/Caddy/Postgres must be running;
other service states are recorded and do not substitute for the parent's operations acceptance.

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
CR1 only after exact bundle/image-reference matching and CR2. Legacy failed post-backup release
or rollback is uncertain, not silently promoted. No-success legacy history cannot authorize release.

Every v2 event has exactly these decoded keys; reject aliases, duplicates, unknown/missing keys
and unintended nulls. Fields nullable below must still be present.

| Field | Rule |
|---|---|
| `schema_version` | literal2 |
| `event` | `operation_started`, `backup_ready`, `operation_finished` |
| `operation_id` | canonical UUIDv4, new per operation, never reused |
| `action` | `release`, `rollback` |
| `server_id`, `operator` | configured canonical server UUID and existing operator-ID grammar |
| `occurred_at` | nonzero UTC RFC3339Nano, nondecreasing event order; an ordinary new operation starts strictly after the previous completion |
| `source`, `target` | non-null identity objects |
| `backup_id` | existing backup-ID grammar or null under the event rules below; no stored path |
| `rollback_until` | UTC date or null under the event rules below |
| `outcome`, `failure_stage` | null until finish; `succeeded`/`failed` and exact failure stage at finish |
| `source_proof_sha256`, `target_proof_sha256` | SHA256-prefixed blob hashes or null under the rules below |

The first v2 row follows the latest v1 completion strictly. Validate the complete prefix before
effects; all subsequent rows keep operation/server/operator/source/target identical. Permit only
one active transition chain: an unfinished release can have only its CR4 governed rollback as
successor, never a competing release or two recovery attempts. Hold the existing operator lock
through observation, writes and effects; append each event and fsync it before crossing its
protected boundary.

- **Started:** release has known source and its source proof, with null backup/deadline/target
  proof. Rollback has the source/target and backup
  of its governing release/checkpoint, null deadline/proofs; a damaged source need not look healthy.
- **Backup ready:** release only, once, after actual host-byte backup validation. Carry its exact
  ID and pending deadline `occurred_at +7days`; target proof/outcome/stage remain null. No drain,
  candidate startup or database change may precede this durable row. An unrecorded produced
  backup remains protected under existing policy, not automatically resolved.
- **Finished:** once, after the real outcome. Success requires the existing identity/smoke gates
  and target binding proof. Successful release requires its checkpoint and repeats its backup ID,
  with completed-release deadline `occurred_at +7days`. A failed post-checkpoint release has the
  same terminal seven-day origin as current completed failures. Pre-checkpoint failure has null
  backup/deadline. Rollback has null deadline and repeats its exact backup ID.
  Failure has no target proof and uses the existing stage vocabulary plus CR4's interruption stage.

The source-proof hash repeats unchanged on every event of its release; it is null throughout
rollback. All outcomes/stages are null before Finished. Finished success has null failure stage;
Finished failure has a non-null stage appropriate to its action and reached phase. Before Backup
ready a release may fail only at supply_chain/preflight/preupgrade_backup or CR4's interruption;
after it, only bounded_drain/startup_migration/epoch_artifact_identity/authenticated_smoke.
Rollback permits its existing post-admission execution stages. No success is permitted without
the referenced target proof and actual parent identity/smoke gates.

Rejected requests before a valid Started row return a typed error without manufacturing a tuple
or an operation record. Failed writes remain failures; no success, rollback authority, or backup
resolution may be inferred from a partly written/unsynced event. Final event failure leaves the
attempt unfinished even if services started. Unknown clocks fail rather than inventing dates.

### CR4 — Explicit interrupted-operation recovery

No ordinary install/release starts while an operation is unfinished or current state uncertain.
Do not automatically resume, retry, finish success, or treat an unlocked lockfile as completion.

Manual `release-abort --operation-id` is permitted only for an unfinished release without Backup
ready. Revalidate exact prior identity via CR2 and current database/epoch identity before appending
Finished failed with `failure_stage:"interrupted_before_backup"`; no running-service/data mutation
(the parent's read-only one-off database inspector is permitted). If
identity cannot be established, refuse and require the separately governed recovery route.
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
rollback closes both chains. A failed/interrupted rollback permits only exact governed retry.
An unfinished rollback must first be explicitly closed as failed `interrupted_rollback` by that
retry command, without claiming service recovery or changing its release deadline; never create
two live recovery attempts. Its new attempt keeps the same tuple and receives a new operation ID.
Before any abort/retry/rollback, prove prior helper/client actions quiescent; shared lock acquisition
alone is insufficient. The supported-engine experiment below must settle this mechanism before
acceptance. No PID/lock-file heuristic or two stable metadata samples may substitute for that proof.
Killing the helper and its Docker CLI is also insufficient: the bounded daemon experiment below
observes an already-started command completing afterward. Recovery must account for daemon-side
work, not merely prove the local processes exited. The actual quiescence mechanism remains open.
An observed terminal Docker ExecID alone is also insufficient as a general fence: the bounded
diagnostic below observes successful parent completion/restart refusal before an explicitly
gated descendant later writes. A proposed mechanism must account for the actual operation's
descendant/process and outstanding-request lifetimes, not silently equate parent exit with
that guarantee. This diagnostic does not show a shipped backup command leaking descendants.
No new backup decryption placement, protection-metadata resolution or deletion policy is adopted.

### CR5 — Consumers and compatibility

One decoder/reducer in `deploymentrelease` serves the controller, CLI and rehearsal readers.
R-006 consumes full v2 events and binding blobs, reducing them to the exact install→release→rollback
operations; it must reject missing checkpoints/proofs, identity substitution, duplicate/unclosed
operations and byte corruption. Three terminal summaries alone are not v2 evidence.

The exact candidate and predecessor helper artifacts used for this amended rehearsal must both
support v2 reading/writing and the existing security floor. Do not claim pre-RP-556 helpers fail
closed before effects or copy current helpers into a historical bundle without rebuilding its
manifest/provenance. Unsupported helper combinations refuse; legacy journal bytes are preserved.
This requirement does not settle the other previous-bundle-floor or producer-identity questions.

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
  can mutate the project afterward, or fail without a new operation/data mutation. Test delayed
  old-request completion after the first stable observation as the negative control. Predeclare
  the actual supported Engine/Compose pair and retain process/request ordering, not an arbitrary
  sleep. Results authorize the exact supervisor/recovery mechanism, never automatic rollback or
  a changed deployment window. If that mechanism cannot be established, this proposal must change.

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

## Changelog

- 2026-10-13: source-grounded RP-558/DESIGN-GAP10 proposal. No product implementation, owner
  adoption, acceptance, archival, push or deployment.
- 2026-10-15: retain bounded Desktop mounted-reader and daemon-lifetime evidence; CR4 explicitly
  rejects helper/CLI cancellation as proof. Native-host mechanism/adoption remain open.
- 2026-10-16: actual ExecID parent completion still permits a gated descendant write. Reject
  generic ExecID-only quiescence; actual-operation/native-host proof and adoption remain open.
