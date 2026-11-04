# RFC: Deployment Operation Evidence

- **Status:** draft — record envelopes/phase reduction independently technically approved; OP9 native profile/producers and adoption pending, not implementation authority
- **Author:** Codex
- **Created:** 2026-11-04
- **Design refs:** `design/06-tech.md` §stack; `design/00-vision.md` pillar6
- **Parent / amends:** [Current Release Authority](deployment-current-release-authority.md), CR3/CR6; [Deployment Foundation](deployment-foundation.md), DP1/DP5/AC10
- **Planning:** owning `planning/deployment-foundation/plan.md` and append-only log

## Scope and authority

Specify the immutable records that connect an admitted operation to its
supervisor, actual outstanding work and closure. The release ledger remains
the ONLY current-deployment authority. These records cannot create an operation,
change its identities, extend its rollback window or authorize deletion by
themselves. An intact digest is not proof that the described work happened.

This proposal defines record shapes and reduction. The native producer/profile,
non-bypassable Engine/SQL/HTTP admission and exact installation/bootstrap
implementation still require OP9. No production reader may accept an experimental
artifact or arbitrary profile as a substitute. No new signing key, host service,
permission, retention policy or automatic recovery is introduced here.

## OP1 — Common representation and storage

All types below are CLOSED objects: listed keys are exact and required, including
explicit nullable keys. Reject aliases, duplicate keys, unknown/missing fields,
wrong JSON types, invalid UTF-8, trailing data and noncanonical encodings.

- `UUID`: the existing canonical lowercase UUID grammar. Server and host-boot
  identities keep their existing version rules. New operation/work IDs,
  recovery/parent references and covered-operation lists specifically use UUIDv4;
  this does not restrict existing server IDs to version4.
- `Hash`: the existing lowercase `sha256:` plus64hex grammar.
- `Date`: nonzero UTC RFC3339Nano, canonical Go time formatting, ending `Z`.
- `Counter`: a JSON STRING matching `0|[1-9][0-9]*`, parsed as uint64; overflow
  refuses. Sequence numbers begin at1. These are representation limits, not
  operating budgets; exhausting a measured guard still invalidates the work.
- `ID`: the existing operator-ID grammar `[A-Za-z0-9][A-Za-z0-9._-]{0,127}`.
- `CID`: full64 lowercase hex Docker container/network ID. Never a prefix/name.
- `Relative`: nonempty ASCII relative POSIX path using only letters, digits,
  `.`, `_`, `-`, `/`; no leading/trailing slash, empty/`.`/`..` component or
  backslash. A path also must belong to its declared artifact/scope, not merely
  pass this grammar. No absolute host/secret/receiver path is serialized.
- `Identity`: the exact CR1 ReleaseIdentity, not a new version-only identity.

Record strings are ASCII. Compact JSON uses ASCII-lexicographic object-key order,
no optional whitespace/escapes, standard required quote/backslash/control-character
escaping, no HTML escaping, no floats, and exactly ONE final LF. Arrays retain
the specified order. `schema_version` is integer1; ordinary bounded integers
such as native exit status remain integers where explicitly stated. Hashes bind
these exact bytes, including LF; readers re-encode and compare before trusting a
digest. Existing v1/v2 release-ledger bytes are NEVER reformatted by this rule.

Store immutable objects as private
`operation-evidence/<64hex-without-prefix>.json`, regular0600 files beneath a
private0700 operator-state directory. Use exclusive temporary creation,
no-replace publication, file sync and parent-directory sync before reference.
Reject symlink/component substitution, wrong ownership/mode, partial publication,
missing objects and changed bytes. Existing protection policy governs retention;
this proposal creates no automatic garbage collector or deletion permission.

The native deployment operator and approved supervisor are trusted writers,
as with the existing ledger. This is local integrity/provenance, NOT a signature
or protection against a privileged operator forging/deleting everything. A reader
must validate the approved producer/profile and OP8's graph, not just hashes.

## OP2 — Supervisor binding

`Supervisor` has exactly:

| Key | Type / rule |
|---|---|
| `schema_version`, `artifact_type` | 1, `supervisor` |
| `server_id`, `operation_id` | configured server UUID, new operation UUID |
| `action`, `recovery_of` | CR3 action and nullable ancestry UUID |
| `source`, `target` | exact CR3 identities |
| `anchor_prefix_bytes`, `anchor_prefix_sha256` | Counter and Hash of the COMPLETE accepted release-ledger prefix BEFORE its unique Started; OP7 gives the install-only substitution |
| `profile_sha256` | Hash of the release-bound OP9 producer/isolation profile |
| `producer_sha256` | Hash of the exact release-owned supervisor helper artifact |
| `producer_manifest_sha256` | Hash selecting one exact source/target bundle (target only for install); both helper and profile must be declared artifacts of THAT validated bundle |
| `host_boot_id`, `engine_binding_sha256` | boot UUID and OP9 native daemon-generation binding Hash; persistent Engine ID alone is insufficient |
| `scope_id`, `scope_started_at`, `scope_binding_sha256` | ID, Date and OP9 actual supervisor-scope binding Hash |
| `created_at` | Date, no later than its unique Started; OP7 gives the install-only admission date |

The approved native profile defines how scope/daemon bindings are independently
observed, and how the caller attaches to the actual surviving lock owner. A PID,
label, unit name or self-reported protocol version is not that producer. No
profile is implicitly approved by supplying its hash. For release/rollback,
descriptor publication does not admit worker effects: the exact Started must
be synced first. Installation uses OP7's explicit admission instead.
Actual executing helper bytes and scope must match the selected artifact/profile,
not merely a caller's copy of its expected hash or an unrelated third bundle.
For release/rollback the descriptor's anchor is its ONE Started event, not each later
row repeating supervision_sha256. Later references verify unchanged binding to
the original anchor; they never demand a hash of a later prefix.

Add required non-null `supervision_sha256` to EVERY CR3 v2 event, immutable for
that operation. It refers to this object and matches its entire tuple. The
same surviving supervisor may govern a related recovery chain, but each new
operation gets its own descriptor/UUID; a new descriptor cannot resurrect lost
authority for the preceding work. For CR3 actions before Started, only explicitly inert native
bootstrap/read-only observation is allowed; its complete safety profile is OP9.

## OP3 — Phase work logs

Private append-only JSONL streams are keyed by operation AND phase, not a second
release ledger. Path is `operation-work/<operation_id>.<phase>.jsonl`. Release
uses phase `execution`; rollback uses `fencing` then `execution`; installation
uses `install`. No other phase is accepted. Every row has exactly:

`schema_version:1`, `operation_id`, `phase`, `supervision_sha256`, `seq`,
`previous_sha256`, `observed_at`, `event`, `payload`.

`seq` is Counter, starts1 and increments exactly once. `previous_sha256` is null
only on row1; otherwise it hashes the preceding complete row including LF.
Dates are nondecreasing; operation/phase/supervision are unchanged. Each append is
synced before crossing its described side-effect boundary. Errors and torn
records leave the work unresolved; do not truncate, skip or repair them in place.
Work UUIDs are unique across all phases/operations in the covered graph. A row's
authority prefix must be an exact earlier accepted prefix authorizing its plan,
not a later event retroactively used to authorize an already-admitted effect.
All work-specific events name an already admitted, still nonterminal item;
work_terminal makes it terminal exactly once. A child cannot be admitted after
its parent becomes terminal. Binding/observation rows cannot reopen a terminal
item, and unknown/orphan event references refuse.

| `event` | Exact `payload` keys and rules |
|---|---|
| `work_admitted` | `work_id` UUID, `parent_work_id` nullable UUID, `kind` OP4 enum, `plan_sha256` Hash, `subject_sha256` nullable Hash, `authority_prefix_bytes` Counter, `authority_prefix_sha256` Hash. Parent, if any, is already admitted in the same phase. The exact accepted ledger prefix authorizes the stage/plan before forwarding. A pending create has null subject until binding; no effect precedes this row. |
| `subject_bound` | `work_id` UUID, `subject_sha256` Hash. Exactly once for a previously unbound item, before that resource is used. Never replace a binding/name-resolve it to a new resource. |
| `work_terminal` | `work_id` UUID, `receipt_sha256` Hash matching OP5. Exactly once, only after actual terminal outcome. |
| `observation_failed` | `work_id` UUID, `code` from `transport_lost`, `deadline`, `native_error`, `invalid_reply`, `authority_lost`. NONTERMINAL; no raw stderr/payload. It cannot close work. |
| `worker_admission_closed` | `reason` from `completed`, `failed`, `manual_abort`, `manual_rollback`. Exactly once; worker-origin admission never reopens. |
| `verification_closed` | empty object. Exactly once after trusted verification/cleanup has finished; ALL admission is then closed. No later admitted work, binding or terminal receipt is legal. |

Before worker closure, trusted supervisor and worker work are distinguished by
the exact approved `plan_sha256`; callers cannot invent a plan. After worker
closure only supervisor-owned cleanup/read-only verification plans may admit
work. After verification_closed only a closure artifact and the referring ledger
append are allowed. Neither log end nor an observation_failed row is a seal.
This closure is phase-specific; only CR3's synced recovery_fenced boundary can
permit the distinct rollback execution phase. It never reopens the fencing log.

## OP4 — Plans, subjects and outstanding work

`kind` is exactly one of `engine_request`, `job_namespace`, `runtime_scope`, `sql_stream`,
`http_stream`, `host_publication`, `host_deletion`, `native_job`.

Each `plan_sha256` names a specific plan in the OP9 release-bound profile. The
plan must define its stage/actor, allowed actions, exact subject derivation,
all child work it can create, terminal observer and required native receipt
schema. Arbitrary command/URL/path interpolation and unknown plans refuse before
forwarding. No payload, credential, cancellation key, session token, whole
environment or secret-config digest is recorded to make a plan look precise.

Subjects are immutable OP9 typed binding objects: full container/image/start
identity, full network identity, protected storage generation, declared host
scope/relative target or registered connection identity as appropriate. Native
unit work, if selected, must bind both request/job and resulting invocation;
it does not disappear merely because the initiating CLI exited. Exact bindings
must be specified by OP9; opaque ID/label-only implementations are unsupported.

Admitting a request and receiving its HTTP response does NOT close child work.
Examples: ContainerCreate returns a CID but its separately admitted job namespace
remains outstanding; a delivered CancelRequest does not close its original
sql_stream; a parent ExecID's exit does not close its descendants. Required
children are admitted BEFORE the parent request can create their effects, then
bound to returned native identities. An unreturned create remains outstanding,
not absent. Every item must be terminal before verification_closed.

## OP5 — Terminal receipt envelope

`Receipt` has exactly:

`schema_version:1`, `artifact_type:"work_receipt"`, `operation_id`, `phase`, `work_id`,
`supervision_sha256`, `kind`, `plan_sha256`, `subject_sha256`, `outcome`,
`native_evidence_sha256`, `observed_at`.

IDs/hashes/types match the admitted item and its bound subject. Subject is
non-null except a typed `refused` receipt for an unbound creation whose approved
native observer proves no resource/effect can subsequently appear. This is not
permission to convert an unreturned/unknown create into absence.
Outcome is `completed`, `refused`, or `retained`. The last is allowed
ONLY for runtime_scope with verified source/target ownership in OP6; it cannot
close a temporary job or unknown writer. Refusal is terminal ONLY when the
native observer proves no continuing effect; transport errors/unknown replies
are not terminal refusals. A failed job can be terminal without being successful.
`native_evidence_sha256` refers to the exact typed receipt schema/producer in OP9,
not free text or a boolean. Terminal Date cannot precede admission/binding or
its native evidence; work_terminal cannot precede its referenced receipt.
Phase closure then precedes its Closure, which precedes its unique ledger
anchor. These checks supplement each stream's nondecreasing dates.

Minimum observer obligations are normative:

| Kind | Required terminal meaning |
|---|---|
| engine_request | Entire known response/handler outcome for the bound request plus all declared child bindings; a delivery/timeout alone is insufficient. Child actors still have their OWN work items. |
| job_namespace | Exact admitted whole namespace terminal with restart/new admission prevented; parent PID or ExecID alone is insufficient. |
| runtime_scope | Whole namespace terminal, or actual verified transfer to the declared retained source/target release scope. Merely running or bearing a project label is insufficient. |
| sql_stream | Original registered server stream and forwarding actors terminal, including key-less startup; cancellation-delivery EOF is not original-stream EOF. No PID-only substitution. |
| http_stream | Exact response/stream and known server-side effect disposition; a canceled client/error response cannot erase outstanding Account/actor/SQL work. Unknown completion requires governed owned-namespace cleanup. |
| host_publication | Exact permitted scope/target, final data identity and publication/directory-sync result, with no surviving publisher. Backup-format/authentication authority is still separate. |
| host_deletion | Exact permitted target generation and accepted protection/deletion policy, confirmed unlink/sync and no surviving delete request. This proposal does not adopt RP-505/D-019 remedies. |
| native_job | Exact admitted native request/job/invocation terminal and descendants disposed under its approved profile; no stable-metadata/PID-only inference. |

## OP6 — Fence and completion objects

`Closure` has exactly:

`schema_version:1`, `artifact_type`, `server_id`, `operation_id`,
`supervision_sha256`, `ledger_prefix_bytes`, `ledger_prefix_sha256`,
`covered_operations`, `legacy_handoff_sha256`, `logs`, `retained_scopes`,
`observed_at`.

`artifact_type` is `predecessor_fence` or `operation_completion`; they are NOT
interchangeable. Each proof has ONE first-introduction anchor:

| Object | Unique anchor |
|---|---|
| Ordinary-release predecessor_fence | That operation's Started, where its hash first becomes non-null |
| Rollback predecessor_fence | That operation's recovery_fenced, where its hash first becomes non-null |
| operation_completion | That operation's Finished, where its hash first becomes non-null |

Prefix byte count/Hash cover the exact accepted ledger bytes immediately BEFORE
THAT unique anchor. Later CR3 rows repeating the immutable hash validate it
against the original anchor/prefix and exact tuple, WITHOUT reanchoring. A later
operation needs its own new current-prefix proof; repetition within one operation
does not permit stale-proof reuse across operations. No own-event/self-hash cycle.
Repeating the reference is historical validation, not a demand to recreate a
predecessor runtime that lawful release/restore has replaced. Recheck native
facts at the admission/verification boundary that requires them, while retaining
the original proof's anchored observation.
`covered_operations` is the nonempty ancestry-ordered unique UUID list covered
by the proof; each has its required phase logs in `logs`. For a pure initial v1 predecessor
handoff it is empty, `logs` is empty and legacy_handoff_sha256 is non-null.
Otherwise legacy_handoff_sha256 is null.
Closure Date is no earlier than its supervisor binding, included legacy handoff
and any included row/receipt/native observation, and
no later than its unique anchor event. Validate the accepted prefix through that
boundary; a valid hash of a stale, unrelated or longer prefix is insufficient.

Each `logs` element has exactly `operation_id`, `phase`, `supervision_sha256`, `bytes`,
`sha256`, `last_seq`. It binds the ENTIRE finalized work stream, ending at its
verification_closed row, not a selected prefix. Array order matches covered
operations and then phase order fencing-before-execution. Reader checks every row/item/receipt/required child; no outstanding
work or extra bytes. Existing closed logs cannot be silently appended/reopened.

Each `retained_scopes` element has exactly `binding_sha256`, `owner`,
`identity`. Owner is `source` or `target` and identity is the corresponding exact
CR3 Identity; array is unique/hash-ordered. The OP9 native ownership observer
must cover the complete retained runtime/storage set from the mutation map,
not just the scopes convenient to enumerate. Normal gameplay/service work may
continue under that declared release ownership; operation-only jobs, delayed
publishers/requests and unknown writers cannot be relabelled as retained service
work. Identity/health/smoke proofs remain separate CR2/parent requirements.

A predecessor_fence is bound to the referring operation's complete predecessor
chain/current prefix and authorizes only its CR3 admission boundary. For rollback
it ALSO seals that recovery's fencing log; current operation is last in covered
operations, but its execution log is NOT yet opened. For ordinary release no
current-operation work log may precede Started. Outstanding predecessor work
receives its final receipts in its original still-open phase log; newly admitted
recovery cleanup belongs to the new recovery's fencing log and binds its own
Started authority. Never append to a previously sealed log.

After the exact recovery_fenced event is synced, rollback may open a NEW execution
log; it may not reopen fencing or change its predecessor fence. Thus establishing
old-work closure does not deadlock the restoration it authorizes. A predecessor
proof does NOT declare current execution complete. An operation_completion covers
ONLY its own operation: release execution, or the rollback phases actually
authorized by history under the matrix below,
and verified retained disposition before its Finished; it cannot
stand in for a later admission's current-prefix predecessor_fence. Later proofs
may reference immutable earlier sealed logs after validating the full ancestry,
not copy an earlier hash into an unrelated field.

Required phase coverage is derived from the exact accepted history, never from
which files the producer happens to list:

| Covered operation and reached stage | Required finalized phase logs |
|---|---|
| Release with Started, including pending/pre-backup failure | execution |
| Rollback with Started but NO recovery_fenced | fencing only; execution work is forbidden |
| Rollback with recovery_fenced, even if execution has admitted no work yet | fencing AND execution |
| Supervised initial-install attempt under OP7 | install |

For the CURRENT rollback's predecessor_fence only, include its completed fencing
phase, not its future execution. For its later operation_completion and as a
predecessor covered by a retry, apply the full reached-stage matrix. A pre-fence
failed/administratively interrupted rollback can therefore have a fencing-only
completion; it still does not close its governing release or permit reset.
Once recovery_fenced exists, missing execution evidence cannot be omitted under
that exception. A no-work execution phase needs an explicitly sealed empty-work
log from the valid live authority; file absence is not proof of no work.

The native profile must account for the complete phase inventory. Reject missing,
unlisted/unsealed authorized logs, execution admission before recovery_fenced,
and extra phase work not authorized by history. A valid supervisor may seal an
authorized no-work phase with its required closure records; lost authority or a
torn stream cannot be repaired into an empty log. This matrix applies equally
to recovery_fence failure, interrupted_rollback administrative failure, inherited
predecessor closure and current completion.

## OP7 — Initial installation handoff

`InstallHandoff` has exactly `schema_version:1`, `artifact_type:"install_handoff"`,
`server_id`, `identity`, `v1_prefix_bytes`, `v1_prefix_sha256`,
`supervision_sha256`, `attempts`,
`retained_scopes`, `observed_at`.

The prefix ends at the exact successful INITIAL INSTALL v1 row, not any
historical successful release chosen by version. Its producer is the exact
reviewed supervisor-aware installer/native profile. Each `attempts` element has
exactly `v1_record_sha256`, `operation_id`, `supervision_sha256`, `work_log_bytes`,
`work_log_sha256`, `last_seq`. It binds one complete original v1 row including LF
to its unique install-phase log. Attempts cover EVERY v1 install row in that
prefix in ledger order; earlier attempts failed and the final one succeeded.
Non-install rows, omissions or duplicate bindings refuse. Every log is fully
sealed under OP3/OP8 and proves the whole installer/outer writers terminal; the handoff
must also match healthy CR2 observation at later use. No v1 row is edited and no
fake v2 install/pending event is created. An old helper cannot mint this proof
by self-reporting compatibility. Unsupervised historical installations remain
outside this handoff; legacy-host adoption needs its separately accepted route.

The install-only supervision descriptor uses action `install`, recovery_of null,
source null and target equal to installed identity. It also REPLACES OP2's
Started-dependent prefix/date/publication conditions: anchor is the complete
accepted v1 prefix immediately before that attempt's first work admission
(possibly empty), created_at is no later than that admission, and the descriptor
must be synced before it. That admission is allowed only by the governed
clean-host install workflow under the shared lock/native profile, never merely
by a descriptor's presence. Its authority prefix matches that anchor. No protected
installer effect precedes descriptor publication AND this admitted work record.
No CR3 v2 Started is invented; its unique operation UUID identifies the work stream,
not a fabricated release-ledger event. Fresh installation must check prior
supervised install descriptors/work logs under the operator lock: empty v1
history does not erase an unfinished installer or permit a clean-install bypass.
Existing pending-install recovery remains outside this amendment. OP9 must
specify bootstrap before this can be implemented.
The InstallHandoff itself anchors THROUGH the final successful v1 install row,
not before a nonexistent Started. Its Date is no earlier than all covered
attempt receipts/rows; a later predecessor_fence references it only after full
OP7 validation, without moving that v1 anchor.

## OP8 — Producer and consumer order

1. Validate accepted ledger/bundles/profile and read-only native authority.
   Under the shared operator lock publish the descriptor. An ordinary release
   first publishes/verifies its predecessor fence from PRIOR sealed work and the
   current prefix, then syncs Started. A rollback syncs its admitted Started
   before opening fencing work. For installation use OP7's governed producer.
2. Before ANY protected effect, sync the work admission/child reservations.
   On native identity return, sync binding before further resource use. Retain
   live unknown requests rather than treating client cancellation as completion.
3. After worker admission closes, the authenticated surviving supervisor finishes
   permitted cleanup/verification, collects each native receipt and syncs each
   terminal row. Close ALL admission for that phase with verification_closed and
   sync its log. The rollback execution phase opens only after recovery_fenced;
   it is never a reopening of the sealed fencing phase.
4. For rollback, seal fencing, publish its predecessor fence, sync recovery_fenced,
   then open execution for reset/restore/start. For every successful final outcome,
   seal execution and publish operation_completion before Finished. Predecessor
   closure and current completion remain distinct; apply CR3's exact failure/
   administrative-finish exceptions without pretending they sealed open work.
5. A reader validates approved profile/producer and native authority or valid
   sealed handoff; then exact immutable graph, ledger ancestry/prefixes, row chain,
   child completeness, native receipt types/outcomes and retained scope coverage.
   Finally apply CR1–CR6's stage rules. No missing verifier defaults to success.

Native historical receipts rely on the approved trusted producer and its
independently executed conformance proof; offline bytes cannot independently
replay an old TCP EOF. This trust boundary is explicit. Native facts still
observable at admission must be rechecked under the profile. A forged summary,
matching counts or a valid hash from an unapproved producer is never sufficient.
Supervisor/gateway authority loss before sealing remains uncertain and refuses
in-place successor mutation under CR6. No retry, new window or key reconstruction.

## OP9 — Remaining profile and acceptance gates

The [native supervisor profile](deployment-native-supervisor-profile.md) now
proposes the concrete one-shot packaging, all eight typed subjects/results,
stage mediation, native lock/daemon binding and scheduled-service handoff.
It is a draft pending technical review, adoption and actual producer conformance;
its existence does not satisfy the gates below. Its explicit cgroup/native-host,
tool-artifact and release-metric projection changes require parent reconciliation.

This RFC is **not acceptance-ready** without a concrete native producer/profile
amendment. It must define exact plan/subject/native-receipt schemas and producer
bindings for EVERY work kind, supervisor/daemon generation, admission isolation,
lock attachment/handoff, retained scope inventory and inert bootstrap. Record
all prerequisites/privileges and reconcile parent no-service/host-floor promises
before adoption. Unknown profiles/plans/native evidence refuse; this document
does not supply an empty registry as a usable implementation.

The [isolated command-worker study](../planning/deployment-foundation/combined-fence-candidate.md#isolated-command-worker-mechanism)
is preliminary mechanism evidence for per-job Docker namespaces/private broker
channels. Its Python/readonly-ping population is not an OP9 Go/CLI profile,
complete job identity, daemon generation binding or write-authorization producer.

Acceptance uses actual helper/Engine/Postgres/HTTP/publication/native lifetimes:
successful install handoff→release→rollback and preserving abort; killed workers,
lost authority, pending creates/native starts, original SQL/canceled HTTP and
outer writers must discriminate. Reject substituted descriptors, extra/missing/
reordered/duplicate rows/children, false terminal refusals, wrong native receipt
types, swapped fence/completion objects, prefix/self-hash substitution, reopened
logs and omitted retained storage/service scopes. Secret-seeded observations
must emit no key/token/config value. Static fixtures and prototype booleans do
not replace runtime conformance or the exact-manifest clean-host/RPO/RTO gates.

## Adoption and implementation

Requires independent technical review, complete OP9 profile/producers and owner
adoption with parent-body reconciliation. Existing v1 artifacts/journals remain
historical; no retroactive completion or archive/publication authorization. The
entire nine-tier1.0 and existing Deployment/backup/retention gates stay intact.
