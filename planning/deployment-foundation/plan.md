# Deployment Foundation implementation plan

- **RFC:** `rfc/deployment-foundation.md`
- **Owner:** Marco
- **Implementer:** Codex
- **Started:** 2026-08-22
- **State:** implementing
- **Implementation baseline:** `cd102d7` (accepted after designated cross-party review)

**Projector contract validation,2026-11-04 (RP-757):** before native-profile
adoption, make the existing proposed15s/300s diagnostics executable. The present
three gauge names do not state whether projection_success is observation success
or business outcome, and host.prom freshness must not stand in for a ledger read.
Predeclare actual pinned Prometheus3.12 promtool rule tests covering fresh idle,
failed/successful/pending/conditional-owner observations, stopped/cache-only
observation at age300 versus315s, missing/error/unknown-owner gauges, invalid
numeric values/future timestamps, duplicate series and fresh unrelated host
metrics. The rule must distinguish a failed release from a failed projection.
Also predeclare one private rule control deleting the age predicate: the same
stale fixture must fail. No fuzzy comparisons, timeout/threshold increase,
shipping rule/metric/CI change or new native producer. A60s per-command guard is
containment only; any guard/setup/missing dependency is INVALID, not evidence.
Use only the existing pinned evaluator image, no network/Engine/secret channel,
read-only public fixtures and owned ephemeral containers. A pass supports a
reviewed proposal/truth table, not implemented cache/ledger ownership, receiver
delivery, native privacy/SP8/R-006 or owner adoption. Actual cache/restart handoff
still requires the constructed producer after its adopted contract.

Executed first filter: actual pinned Prometheus35 metric-series cases/37 asserted
evaluations pass, including age300 clear/315 firing, failed-release versus projection
failure, missing/invalid/duplicate gauges and non-float histogram refusals. The
initial32-case predicate missed three added histogram controls; actual baseline
fails all3, and absent(metric+0) checks correct that concrete type gap. Deleting
only the final age predicate makes the same full fixture fail precisely the
cache-only/fresh-host stale case. Original nonfinite repeat shorthand caused a
parser setup failure; explicit literals retain all cases/expectations. No software
or producer pass was inferred. All owned containers are absent; guards never fire.

Draft P01 now fixes diagnostic/result meanings, two atomic metric files with
supervisor status-only versus retained-service business ownership, original-prefix/
cache ordering and eight actual restart/handoff populations. It explicitly proposes
DP7's additional native-only alert and C01 logging exception without changing
accepted parent text. The draft's circular before-adoption native-cache-test line
is reconciled with SP8: isolated predicate proof before adoption; actual producer/
restart/receiver/native proof after adoption before acceptance. Complete new
research/tool/technical-spec union407b641d..dcc151f2 is independently APPROVED
with full actual35/37 and both controls, after factual rollback-mapping and stale-
ledger-sequencing corrections. No owner adoption or native/release proof follows.
Do not repeat pending adoption questions.

**Contract-graph adoption preparation,2026-11-04:** complete P01 review exposes
stale CR6/Open-question and OP9 text still calling now-proposed schemas/profile
unspecified or demanding actual producers before permission to construct them.
Author reconciles those own-draft bodies to the linked concrete contracts while
preserving ALL actual native/clean-host acceptance. SP8 now enumerates the exact
proposed DP1/DP5/DP6/DP7/DP8 changes and full control-contract adoption scope,
including rootful local Linux/amd64 prerequisites, no new resident service,
paused scheduler/verified-result ownership, C01 logging exception and extra
projection alert. Accepted parent and owner-authored rulings remain untouched.
The complete eight-path linkage uniondcc151f2..ed9a8f2a is independently
TECHNICALLY APPROVED after reconciling CR6's old outer-outcome-writer alternative.
One concrete architecture/parent-reconciliation adoption question is now delivered
and unanswered. It covers the linked CR/OP/SP/adapters design and explicit SP8
parent/platform changes, not construction/native acceptance or release approval.
Static tool/MPL, signing-metadata, host access and other pending questions remain
separate; do not duplicate, infer or absorb them. After applicable adoption,
reconcile parent bodies under the actual delegation and implement adopted
contract dependencies with their full producer/native/clean-host proof. No
product implementation or goal pause is claimed at this checkpoint.

**Current exact-artifact construction, 2026-11-05:** exact clean source
`28aefd29` has fresh independent trees `.cache/release-current-28aefd29/bundle-a`
and `.cache/release-rebuild-28aefd29/bundle-b`, matching all75 files/modes and74
bound artifacts. The [new retained record](release-builds/candidate-28aefd29.json)
includes approved RP-768/RP-767 at kernel0.3.226, epoch8 unchanged. Both complete
builds and two fresh gameserver scans pass. Six immutable upstream/rehearsal
inventories per tree explicitly reuse prior scans, with exact old checksums,
producer-output equality and fresh registry/config/platform checks.
The actual shipped Linux helper passes intact/build-record consumers and all9
exact-exit semantic tampering checks; source/image secret scan is clean. This
new four-path evidence batch needs independent review. Neither tree is a
predecessor or native/clean-host/R-006/rights/current-hosted/full-feature proof.
The local preview label is unchanged; full nine-tier1.0 remains the objective.

**Direct upstream-Compose proposal, 2026-11-04 (RP-759):** the current
[tool companion](../../rfc/deployment-tool-artifact-inventory.md), native SP1/SP3/
SP8 and adapter C01 now propose one upstream compiler executable, no worker
Engine/secret descriptor, explicit validated interpolation inputs and unlogged
private output through original E12. The existing application policy, exact
manifest/source/notices/covered-source closure and every parent acceptance floor
remain. The unused unadopted local build system is withdrawn; another publisher/
build mode needs its own concrete accepted contract if selected.

Earlier core/adapter/tool reviews cover their cited historical bytes, not this
amendment. The reviewed v3 configuration study is below; two new input arms also
complete the actual public population. Complete11-path765aae44..70acee7f now has
independent combined research/tool and technical-spec APPROVAL after the exact
exec-environment/image-default and withdrawn-runner corrections. Next is an
explicit owner adoption request for the reviewed tool/policy scope, now delivered
and unanswered. Do not repeat that question or infer acceptance. Static
inventory construction can be adopted separately from
native runtime dispatch; private handling/native/SP8/R-006 must still be built
and proved before feature acceptance. No shipping version or rights clearance,
production/schema implementation or whole-CI claim is made.

**Predeclared bounded study — direct Compose configuration:** pinned Compose source
`cmd/main.go` and `cmd/compatibility/convert.go` have an explicit standalone path.
Question: can the actual cached2.39.4 executable supply the same configuration
outputs/refusals as CLI28.4.0 plus that plugin, without a separate CLI executable
or Engine channel? Predeclare minimal public fixture, current production base
and rotation inputs with synthetic public values, invalid YAML/type/required-env/
missing-file cases, and removed/swapped executable controls. Compare complete
parsed models in memory and actual version/exit outcomes; never persist private
resolved models. Each case uses the existing isolated read-only/nonroot/no-cap/
no-network fixture boundary and prior command/work/cleanup guards; a guard,
missing dependency or incomplete population is INVALID, never a refusal/pass.
Hash-check actual tools; make external/system CLI and daemon sockets unavailable
in direct mode; no PATH fallback. Preserve all case results, not just survivors.
A pass authorizes a smaller artifact/profile proposal for review and owner
choice, not selection, implementation, rights clearance or waived conformance.
Desktop/emulation research is not native Linux/amd64 SP8/R-006. A partial or
negative result stays bounded and cannot silently change the two-tool draft.

Executed first filter: all3 complete valid models match; both modes reject all4
invalid cases; missing/swapped binary controls and an actual parsed policy-field
change refuse. RP-764's original Python bool/int equality defect is corrected:
v2 rejects numeric1 in the actual boolean field and accepts object-key reordering.
Its bounded independent approval stands; root subsequently found float decoding
could merge distinct large numeric tokens. Current v3 carries raw model JSON to
a lossless typed/token comparator, adds precision/string/duplicate/nonfinite
controls and passes the full actual population with3 byte-identical pairs.
Exact binaries run inside checked nonroot/read-only/no-cap/no-network boundaries
without Engine sockets; direct mode has no CLI mount and no PATH fallback.
Only synthetic public model output is used: Docker stdout/logs may persist it,
while final observations omit models. No private-model transport is measured.
Owned cleanup passes. Complete eight-path539e1bb4..7a5368b5 is independently
APPROVED as bounded research/tool, including reproduced old failures, exact
numeric transport, all current controls and a full fresh Docker run. RP-764 is
CLOSED for this instrument correction. All observations/verdicts remain intact.
At that historical study boundary, the next action was to draft a single directly
invoked compiler amendment; the study itself did not change the then-reviewed
two-tool contract. The current amendment and its review status are at the top
of this plan. The old research version is not selected as a shipping pin, and no
native-host release proof is claimed.

**Compiler input construction study, predeclared 2026-11-04:** before specifying
the direct worker's grants, test whether an explicit environment plus an empty
dotenv file preserves all3 current configuration outputs, then whether those
outputs can compile with every referenced secret-file path absent. Use the same
actual tools/templates and full valid/invalid/control population, guards and
owned cleanup. New process_environment arm must match retained v3 valid stdout
hashes; process_environment_absent_secrets must match within direct/wrapped
pairs and actually expose6/8 absent secret references in base/rotation models.
Direct mode also gets a private empty auths-only Docker config, no discovery.
Values are synthetic public data. A pass supports proposing no secret-value FD
or Engine channel for this compiler; it does not prove native environment/path
binding, private output handling, lifecycle closure, artifact eligibility or
any other worker's grants. Unknown/guards/skips invalidate, not count as refusal.

Executed29b5a5/33190→e2cbc0 and2ec074/93818→97a190: each arm passes3 full
direct/wrapped model pairs,8 invalid-input refusals,2 binary controls and all
typed/token controls with owned cleanup. Process-environment outputs match all6
retained v3 positive mode/case hashes/lengths (047f2b). Absent-secret references
are actually0/6/8 in minimal/base/rotation for BOTH modes; the direct Docker
configuration has empty auths only. The new selector rejects unknown mode before
work (5b36c0). The complete source/data and corrected C01 translation are
independently APPROVED765aae44..70acee7f; the reviewer reran both actual arms.
This is not owner adoption or private/native acceptance.

**Nested attribution correction,2026-11-04 (RP-761):** independently APPROVED
complete six-path957dfa6b..a8168b63. Actual old imported-GDDO BSD omission
reproduces; corrected full affected tests/vet and fresh53-dependency generator
pass. Imported ancestor scopes retain their license/NOTICE text under relative
paths, with known-grant union and complete refusal controls. Reviewer output hash
matches recorded attribution. RP-761 is CLOSED for this application defect;
external-tool/MPL/source-rights, native/adoption and full release gates remain.
The prior seven-path nested-observation batch1cc6762a..7623cf93 is independently
approved; its v1 evidence is12positive/25UNKNOWN. RP-762 now corrects the missed
escaped YAML names with a retained v2 probe:15positive/22UNKNOWN, independently
APPROVED complete nine-path a6592af3..571b0d8c. RP-762 is CLOSED within that scope.
The matching Compose source query selects17 nested scopes and no package under20;
neither zero result is binary absence or license clearance. RP-760 stays
closed within its original root-NOTICE scope.

**Application attribution correction,2026-11-04 (RP-760):** independently
APPROVED complete8-path9cde3a64..0c746ca1. Old actual Prometheus NOTICE omission
reproduces; corrected full metadata packages/vet and committed-source generator
pass. Independent generated attribution matches the recorded SHA256, preserving
SPDX/allowlist and requiring complete browser-copy agreement. RP-760 is CLOSED
for this root-NOTICE defect; canonical docs describe delivery. Nested/MPL/tool/
source-rights, native architecture/adoption and full release gates remain separate.

**Native profile construction,2026-11-04:**
The preceding core/adapter proposal was independently TECHNICALLY APPROVED, exact union
c58f8e48..224fc0e2 plus4118c4d8..a836e92b; Headcount editorial excluded.
The owning log preserves all negative verdicts and actual diagnostic execution.
The current direct-Compose amendment above still needs its own review/adoption.
SP1–SP8 plus the [adapter companion](../../rfc/deployment-native-supervisor-adapters.md)
now define native bootstrap, exact stage/Engine/SQL/HTTP/file/native plans,
immutable creation/start, source/target identities, worker FD/control/results,
first-volume binding, scheduler freeze/import/cache, protected staging and final
ledger projection. This is technical coherence, not implemented/runtime evidence.

The [tool provenance audit](native-tool-provenance-audit.md) now executes that
first inventory: actual Compose publisher graph covers180 modules but leaves182
licenses unknown; actual CGO-enabled CLI has no dependency module list. RP-759
keeps source/native closure and license/NOTICE delivery OPEN. Existing app SBOM
cannot silently cover either new binary. Source-bound manual tool/audit is independently APPROVED
ad1d23c1..3fd7483c; no clearance/distribution or shipping version adopted.
The complete source follow-up now verifies all180 archive h1 identities and
retains237 scoped notice files, including37 nested. Actual current detector
recognizes176 roots and refuses4 exact MPL2.0 roots; the manual diagnostic
correctly exits nonzero. Complete ten-path4af4fe6b..9c90bc2e batch is independently
APPROVED as research/tool; no license-policy widening or clearance.
The CLI source follow-up now executes the matching Go1.24.7 Linux/amd64/CGO1
query:623 packages (83 CLI,338 vendor across89 exact modules,202 stdlib),3,203
source files and116 named scoped notice candidates. Exact tree/SDK/source bytes
are checked; private incomplete/unowned/missing-command controls refuse. Complete
seven-path66625d72..ac45914c independently APPROVED as bounded research/tool;
fresh observation is byte-identical and all seven controls pass. This is source selection, not the
published binary's compiler/link map. Three declared CGO packages and distinct
inline OASIS header terms leave native/rights closure open; see the audit.
The Compose matching-source collector now binds1,290 packages/180 module h1
identities and6,962 inspected source/text files to Git/SDK/module archives;
245 named candidates include main/stdlib and every prior237 module text.
Four real-query source/graph faults plus no-overwrite refuse. Full overlaid
metadata tests/vet and actual corrected90,355-function probe pass. Complete
source/symbol batch independently APPROVED a6592af3..571b0d8c, with both complete
observations byte-identically reproduced; see latest log/probes.
The companion above proposes covered-source/notices delivery. Actual inline/
header terms, published-binary/source correspondence and CLI's native-component
closure remain shipping prerequisites; source selection is not a rights decision.

CLI native follow-up now inspects807 CUs/35,689 function symbols from the actual
ELF:212 non-Go units,330 executable-backed native ranges and549 native-range
function symbols. Keep8 no-range native units,13 zero-size symbols,2 unassigned
function extents and unmatched file symbols explicit.192 musl-labelled paths
have upstream1.2.5 counterparts;12 runtime/cgo paths match SDK/compiler source;
8 generated units remain unknown. Exact Alpine patches/CRT/native source identity
and rights are not established. Source counterparts and extent/binding/unknown
controls execute. Complete nine-path a590aa2d..23c646b4 is independently APPROVED
as bounded research/tool, with both reports reproduced byte-for-byte. Next narrow
actual source/patch identity through line-file checksums and original build-package
provenance, plus CRT/generated/header counterparts and full delivery terms; another
positive census alone cannot complete native closure or authorize packaging.
The checksum follow-up now maps all795 line tables/212 native references and
57 checksum records;55 match source candidates,2 generated records stay unknown.
Musl's192 tables have no checksum records, so that route cannot settle exact
source-body patches. RP-763 identifies a fortify header: the pinned Alpine1.1-r5
recipe reconstructs its actual recorded checksum, while unpatched upstream does
not. Root license evidence, source/patch hashes and all no-checksum/unknown rows
remain explicit. Complete Go/LLVM/real-ELF controls pass; complete nine-path
5d73b715..ef2780e4 is independently APPROVED with exact fresh LLVM/ref/checksum
reproduction. RP-763 closes its bounded identification/instrument route; RP-759
remains open. No package identity, rights, build or source-body closure follows.

Publisher OCI counterpart now has a complete offline descriptor/subject/diff-id/
member check. Its executable sections and330 known native ranges match, but its
whole binary differs from the static download in four sections. Real wrong-
subject/data-only identity controls discriminate. Full metadata overlay/vet pass;
complete six-path dc0c652c..53280165 is independently APPROVED as bounded research/tool with exact report and executed controls. Do not attach the image's
statement to the original or silently substitute the artifact. DP1 does not
implicitly demand universal signed/reproduced upstream builds; finish truthful
per-tool inventory, terms/delivery and provenance, then the concrete artifact/
policy proposal and applicable adoption. Further path/code equality cannot do that.

The tool-delivery and P01 predicate/contracts now have their own exact technical
approvals; the full linked CR/OP/SP8 adoption package is independently reviewed
above. Await the explicit architecture/host-floor and separate tool-policy
adoption, then reconcile parent bodies before dependent production implementation. Actual selected-tool rights and inventory must be settled before
shipping eligibility. Then construct native producers/adapters/service
control with complete dependent review and actual SP8/native-host/manifest R-006
proof. The currently cached tools remain research-only, not approved shipping
provenance. No empty registry/table validation substitutes for that work.
RP-754/756/757 and host/adoption/producer/R-006 gates remain OPEN.

RP-758 is independently reproduced: matching-positive local HTTP delivery
passes, but unrelated delivery with ZERO matching nonce also returns helper
success. Manual probe/observation source-bound; real helper plus modeled-manager/
local-webhook only, not real Alertmanager. H06 proposes native/operator attribution,
not a silent runtime remedy. That exact-delivery defect stays OPEN; aggregate
counts cannot establish nonce delivery. No feature or release gate promoted.

**Carrier result,2026-11-04:** both native populations completed in the
[source-bound observation](authority-carrier-observation.v1.json). The actual
operator lock and original stream survive launcher exit in the separate
supervisor namespace. Original writer pidfd exit plus original EOF precede
same-path/same-ID replacement refusal; lock stays held. Omitting the lock FD
refuses before a supervisor probe; fresh actual lock acquisition succeeds after
fixture exit. Independent research/tool review APPROVED0f3a9a7c..695540b7; exact
verdict is in the owning log. This is model evidence, not an actual
Docker-daemon/authentication/bootstrap or supported-host result. Next bind this
carrier and the reviewed SQL observer into the complete eight-kind supervisor
profile/admission/receipt protocol against the existing mutation map; no contract
adoption, production implementation or release acceptance follows from this run.

**Supervisor authority-carrier experiment,2026-11-04:** source/contract work
exposes daemon-generation/lock/bootstrap lifetime as the next missing binding.
Before inventing a permanent service or assuming native handles survive the
caller, test real Linux SCM_RIGHTS transfer of the existing exported
AcquireOperatorStateLock FD, a connected native Unix stream and its writer pidfd
into a separate-PID-namespace supervisor. Model socket activation by creating
the listener in one process and passing it to a different-UID server; compare
SO_PEERCRED with actual SO_PASSCRED/SCM_CREDENTIALS from fresh replies. Open the
writer pidfd and obtain a second fresh reply before transfer. Complete transfer
must keep the REAL operator lock busy and the same stream usable AFTER launcher
exit. Native server exit and replacement at the same path/stable ID must leave
the supervisor refusing new dispatch while keeping the lock. A missing lock FD
must refuse before any supervisor probe. Actual library lock acquisition must
work after fixture shutdown. Two owned model populations,60s per population,
ordinary handshakes10s; build180s outside work, cleanup15s PER Docker command.
No real Docker-daemon restart/host PID exposure, host settings, image pull/ports,
keys, production code or deployment. Same-namespace launcher/server and separate
supervisor use only exact owned container namespaces. This is native carrier
mechanism evidence, not full daemon HTTP behavior, trusted peer authentication,
inert bootstrap/OP9 adoption, packaged supervisor or clean-host recovery proof.

**Next whole-connection ownership question:** the proven key-reading gateway is
plaintext-only; do not weaken existing verified-TLS backup support to fit it.
Before adding a TLS-terminating/authenticating proxy, test whether an observer
inside the exact owned Postgres namespace can bind an accepted socket's full
native tuple/inode to a backend pidfd BEFORE signaling that same process. A
pidfd alone is not ownership: prove socket association, exclude the postmaster/
unrelated children, handle startup before activity rows and process exit/reuse,
retain original EOF observation, and preserve a continuously active bystander.
This is a research candidate, not approved production PID signaling, a selected
host prerequisite, or permission to inspect/modify an external host. Check
actual kernel/seccomp/UID support in disposable fixtures; unsupported/unknown
identities must refuse. If viable, compare TLS and startup populations against
the existing original-stream and prior-data invariants before proposing OP9.

Readiness d8c9ea passes on the disposable Desktop fixture: same PostgreSQL UID70,
CapEff/Prm/Bnd0 and NoNewPrivs1 can open self/postmaster pidfds, check permission
with signal0 only, read socket FD links and the namespace TCP table. This does
not authorize a real termination or prove tuple ownership. Remove that fixture.
Next bounded experiment reuses the unchanged actual backup/restore package and
the existing manual child harness. An opaque TCP relay must register accepted
backend ownership before forwarding client bytes, retaining the original stream
through EOF without TLS termination/key parsing. Native observer uses the exact
owned Postgres PID/network namespace, opens pidfd BEFORE validating socket inode/
full tuple/UID/direct parent, excludes postmaster/ambiguous subjects and retains
that reference. Positive verified-TLS encrypted roundtrip; interrupted actual
TLS helper with wrong-binding refusal then exact pidfd termination; first-helper
startup held before activity rows. Preserve continuous bystander/Postgres/rows,
verify old references return ESRCH, and reject incomplete/error/changed ownership.
Setup30s, each of3 populations60s, ordinary observation10s; build180s/actor and
observer240s are outside population budgets; Docker cleanup15s per command.
No image pulls/ports/real secrets or production signaling. Initial tuple-to-job
authorization, adversarial kernel tuple/inode reuse, durable restart/native-host/
full CLI/OP9 adoption remain separate even if these native populations pass.

The initial EOF-only criterion FIRES3eb05d/901c0e after actual TLS backup child
success: original native TCPread is ECONNRESET. Preserve original9bf8fdba/v1;
restore/interruption/startup did not run. New experimental producer requires
BOTH retained original backend pidfd exit AND exact original EOF/reset plus
ended forwarding actors. Resets alone never close work, all other read errors
remain invalid. This is additional native evidence, not an adopted OP9 change.
Final cb4352/5cfaec exits0 (v2): verified-TLS encrypted roundtrip, interrupted
actual helper and held pre-activity startup all execute. Seven distinct native
bindings close with retained-pidfd exit: four positive streams reset, interrupted
two/startupone EOF. Wrong binding refuses without signaling; actual SIGTERM2,
old references yield ESRCH3; same continuous control/postmaster/rows7 then7/42.
All11 new terminal plus unchanged14 framing/4 deadline controls pass, no skips.
Roota8517b binds v2 to final sources;2b7282 Linux/amd64 overlay vet passes.
Full original-plus-native-producer range needs independent review. Native tuple/
job admission, adversarial pre-binding inode/tuple reuse, trusted observer
permissions, complete CLI/outer writers/daemon lifetime/native host/adoption
and R-006 remain. No TLS interception or BackendKeyData parsing is introduced.

Independent original-plus-native-producer review is CHANGES REQUIRED for RP-755:
actual d0d69c/23b33f accepts11.021604s queueing because the10s RPC clock starts
after mutex; e02ea3/0f2837 also publishes an already-cancelled call. Actual native
8476a3/b67fac and private no-signal failure f05842/15edaa otherwise corroborate.
Preserve verdict/v1/v2. Forward correction starts at entry with context-aware
TryLock and guards before publication. Fresh98a13e/ba1f57 v3 passes all3 native
populations,6 queued/timely/cancelled/expired/encoding/write-error controls plus
the11 native terminal and reused framing/deadline controls; veteb13f0 passes.
Complete corrected union needs independent re-review. No deadline raised;
already-published work cannot be unsent by cancellation or erased from scope.

Complete original-plus-native-producer-plus-correction union through2e20afde is
independently APPROVED. Fresh2246cb/bc6565 all native populations,428566/888f63
original11s/cancel controls,ca8c63/e57d11 timely queue,758e37 bindings and2d5ad5
vet pass;6f90a8 confirms full10-path scope/negative preservation. Own reviewer
fixtures removed. RP-755 closes instrument only; RP-754 stays OPEN. This remains
Python native-observer research, not a release-packaged Go producer. Next bind
the original held socket/admission and native producer to the complete trusted
supervisor profile/whole mutation inventory, with supported-host proof before
adoption. No repetition of these unchanged positives or full recovery claim.

**RP-753 existing TLS-input correction,2026-11-04:** URL certificate paths reach
Go but were dropped from libpq service config. Valid old-code25bca5 proves Go
verified TLS/client identity while actual psql fails at default root.crt instead
of the supplied CA. Preserve sslrootcert/sslcert/sslkey with line-injection/
unrepresentable whitespace refusal before credential writes, under existing DP6
database-URL authority; no new TLS deployment policy. Real corrected740b52/f0375a
passes matching Go/libpq TLS client identity (including paths with spaces), full
encrypted backup/restore identity and explicit wrong-CA verification failure.
Complete declared Postgres integrationff08c7/045adc, cold native packagecc2aa5/
2a2e55 and vet211db7 pass. Native SQL skips are not that integration evidence.
Initial key-permission and DN-format setup misses are preserved in log, not
the bug reproduction. Complete2a02536f..857569fa is independently APPROVED:
private old-source4dadf9/359e4a reproduces, corrected8d6c90/6a2cd5 all10 real
tests/no skips,1ec9ae package/ac833a vet andd9f529/df028b scope checks pass.
Reviewer fixtures removed; RP-753 closes only this defect. This does not close
the experimental gateway TLS/native-host
profile, protected-metadata/CR4 or clean-host/R-006 acceptance.

**Actual backup-helper connection study,2026-11-04:** the next unknown is the
whole existing CreatePostgresBackup/RestorePostgresBackup connection population,
not another standalone pg_dump. Compile a manual test-only overlay in the owning
package so existing fixture metadata/identity helpers and unchanged production
functions are reused. A multiple-connection gateway must fsync experimental
admission BEFORE each native dial, retain every actual pgx/libpq stream through
server EOF and keep BackendKeyData only in memory. Positive control must make an
actual encrypted backup, restore it and compare canonical identity plus explicit
fixture rows through the gateway. Interrupted population must observe the actual
Go helper's exporter and blocked pg_dump, kill its owned process group, show
wrong-key delivery leaves the original dump blocked, then correct-key delivery
produces ALL original server EOFs BEFORE releasing the lock. Preserve the same
live DB/control connection/rows. Test closed admission and actual journal-write
failure as refusals, not missing-session inference. Setup30s, each of2 populations
60s, ordinary handshake10s; compilation180s/actor240s are observation guards,
cleanup15s per Docker command outside work. Guards/errors/incomplete populations
invalidate. Use declared disposable Postgres service, no ports/pulls/real secrets.
Synthetic release metadata remains labeled; this is actual package-function
binding on Desktop, not the complete CLI/metrics/encryption publication fence,
non-bypassable worker isolation, durable authority reconstruction, TLS profile,
OP9 adoption or native Linux/amd64 acceptance. No production/CI/API change.

Final actual b6e37d/1637f4 exits0, sourceca26d970: unmodified package functions make
and restore a real encrypted backup through2 registered sessions each, with
full canonical identity and explicit restored row7 equal. Actual interrupted
Go helper also uses2 streams; parent/group death and wrong-key delivery leave
the original dump blocked, then correct-key delivery yields ALL original EOFs
before lock release. Same control/postmaster retained, rows7 then7/42. Closed
admission yields exact111; real journal-write failure occurs before native dial.
All14 terminal/framing plus4 deadline controls pass, no test skips. Final
Linux/amd64 overlay vet5bd1d6 passes;62d424 binds retained v1 to final
source/runner and excludes
fixture password/private key in observation. Complete86b82187..efc4d402 is
independently APPROVED: fresh ee94bc/55acca all populations; private correct-key
severing dd9f6f/32439f fails unfinished-session observation;783b19 guards/error/
artifact corruptions and94b5cf vet pass. All reviewer fixtures removed.
Next full supervisor/OP9 binding must include actual CLI/outer metrics and
interrupted output publication, not relabel this package-function proof as those.

**Actual CLI/Go broker experiment,2026-11-04:** download only public Linux/amd64
Docker28.4.0 client and Compose2.39.4 into ignored research storage; extract no
daemon, install no host tool or new shipping dependency. Official Docker index
lists the archive; GitHub release metadata binds Compose binary SHA256
7af95166a730b87e172d4fc9aefea8725d3c6c7327d59149267b452114ddb7d4.
Record URL/bytes/digests and the archive's lack of an independently checked
upstream checksum, without calling TLS/local hashes signed provenance. Public
download/build guards180s each are setup, not runtime/RTO budgets.
Compile a manual Go Unix/HTTP broker through the existing root Make overlay.
In the previously proved Docker worker boundary, actual Docker version and
Compose version/config must execute, while default raw-daemon connection fails
and an actual owned ContainerCreate attempt receives broker403 with zero child
creation. Broker permits only read-only ping/version and exact fixture-image
inspection, checking actual SO_PEERCRED; wrong UID0 must be refused. Keep native
API trace categories/status/UID only, never whole bodies/env/credentials.
One60s measured fixture guard, <=10s per ordinary command/handshake, cleanup15s
PER command outside work. Typed setup/permission/transport/guard failures are
invalid, never evidence of policy refusal. No root-cap/port/image-pull/production
change. This proves actual client/Go-proxy compatibility, not admitted API writes,
durable supervisor/daemon lifetime/OP9 receipts, native Linux/amd64 support,
adoption or full recovery. Desktop kernel is Linux/arm64 with amd64 clients.

Actual cafdcb/418737 exits0: Docker28.4.0 linux/amd64 sees Engine28.4.0
linux/arm64 through the Go socket broker; Compose2.39.4 version/config execute.
Actual ContainerCreate reaches POST and is refused403 without forwarding or
child creation; UID0 is refused403. Native profiles match network-none/read-only/
cap-drop/no-new-privileges, and raw socket is ENOENT2. Eleven actual Go policy
controls pass; Linux/amd64 overlay vet0df9ed passes. Retained v1 binds exact
sources/binary/tools. Complete2699ad00..04452299 is independently APPROVED:
fresh c8216a/f54d40 reproduces all populations, f3ce08 refuses trace/profile/
deadline/error corruptions and validates bindings; d605ec vet and66728a full
scope/static-binary/archive checks pass. Reviewer cleanup inventories empty.
This is compatibility research, not durable write admission/native-host/adoption/CR4 proof.
Earlier interrupted output is inadmissible: terminal output was lost, no probe
process/resources remain, and its retained scratch directory alone proves no
outcome. A pre-commit source check corrected the broker stdout cardinality from3
to2 (docker exec controls has separate stdout); old/new actual-shape control
reproduces the error. No guard or policy was weakened to obtain this pass.

**Worker-isolation experiment,2026-11-04:** test a Docker-backed candidate rather
than assume an untested native host service. Use a uniquely owned declared tmpfs
Postgres fixture and cached config-pinned Python image; no pulls/host ports/new
shipping dependency. Compare same-UID/cap-drop network-attached writable controls
before/after an isolated UID65532 worker with network none, read-only root,
all capabilities dropped and no-new-privileges. A private Unix-socket broker on
an owned volume must permit that worker's exact GET _ping to the actual Engine,
reject a wrong method and reject UID0 by actual SO_PEERCRED. Deliberately broad
fixture socket mode lets the wrong UID reach the policy check. The worker must
fail direct owned-Postgres TCP with ENETUNREACH, raw daemon paths with ENOENT,
and writes to world-writable root/volume targets with EROFS; arbitrary errors
are invalid, not policy proof. Controls must actually write and receive the
same PostgreSQL SSLRequest refusal byte before/after, proving a live target.
Inspect exact container image/user/network/read-only/capability/security/mount
metadata as well as worker output. One60s measured-work guard, <=10s ordinary
commands/handshakes; cleanup<=15s PER command outside work, no aggregate/RTO
claim. Incomplete actors/guards/errors invalidate. This is Desktop Linux/arm64
kernel mechanism evidence with amd64 PG, not actual Docker/Compose CLI workers,
the Go supervisor/broker, API write authorization, all bypasses, Linux/amd64
supported-host/OP9 adoption or full recovery acceptance. No real secrets/data.

Actual57766 completes the full boundary/control population: exact ENETUNREACH101,
ENOENT2 and EROFS30; actual Engine200/OK through the channel, typed403 route/UID
refusals, CapEff/CapBnd0 and NoNewPrivs1, expected native image/user/network/mount
profiles, before/after actual PG refusal byteN and successful local writes.
All owned actors exit/cleanup. Root545c35 passes exact/invalid errno controls but
finds RP-749: final source-binding collection can cross60 after its last guard
and still claim completed. Preserve original source/v1, then forward-correct
the collection boundary and rerun the changed instrument. RP-750 is a narrow
native mechanism candidate, not a complete supervisor/profile or adoption.

Original source/v1 is preserved aa765a40. Roota41a20 additionally shows local
command timeout alone accepts a simulated11s matching return under global60.
Correction checks command/cleanup local deadlines AFTER output collection and
global work AFTER all final metadata/hash collection. Actuald110f3 reproduces old
acceptance, passes timely, refuses exact/late final/command/cleanup. Fresh76328
exits0, all original native outcomes unchanged, v2 binds3d672aa9. Actualdc6858
passes broker allow/UID/route/native-error and native-profile corruption controls.
Complete original-plus-correction review remains; no bounds/populations changed.

Independent union review is CHANGES REQUIRED for RP-751: actual92969/diagnostic
24761 both exit1 at broker native Mounts order, before after-control. Same exact
destinations/type/RW appear reversed. Original negative remains, no independent
full pass inferred. Forward correction sorts by destination without dropping
entries/values and rejects duplicate destinations; actualab38d7 shows old refusal,
current both-order pass and invalid-mount refusals. Fresh14788 exits0, complete
native population unchanged, v3 binds2cfcb55c. Guard functions are AST-identical
to independently checked4f0c5464. Complete containing union needs re-review.

Complete original plus both correction ranges through299ad114 is independently
APPROVED research/tool scope. Fresh12798 completes EVERY population;dd3dfd checks
old/current permutations, corruption and all source bindings. Prior independent
guard/error/errno/broker controls apply to verified unchanged functions/actors;
082e80 verifies full union/scope/negative preservation/cleanup. RP-749/751 close
instrument defects only; RP-750 and OP9 remain OPEN. Next execute actual
Docker/Compose command binaries and the Go broker in this per-job boundary,
bind full job/channel identity and stage-aware admitted write routes/receipts,
then complete supervisor/daemon/bootstrap/native-host ownership. No repeat of
the completed Python isolation population absent a source change/new concern.

**Operation evidence proposal,2026-11-04:** new draft
`rfc/deployment-operation-evidence.md` specifies canonical private hash-bound
Supervisor, phase work-log, receipt-envelope, predecessor/completion and
initial-install handoff records. Rollback fencing and execution have separate
sealed logs, preventing a fence from closing the restoration it authorizes.
Explicit ancestry/prefix/child/retained-runtime/attempt coverage and producer
trust are required, not counts or experimental booleans. Parent CR3 gains the
immutable descriptor reference; no production/schema generator is implemented.
Whole proposal needs independent technical review. OP9 still requires exact native
plan/subject/receipt producers, profile/permissions, bootstrap, non-bypassable
admission and successful/refusal runtime/native proof before adoption/building.
Next specify/prove OP9 against the complete mutation map, not another raw
pg_dump repeat, invented certificate or standalone tracking validator.

Original six-path b74d02a7..7a7e7b39 review is CHANGES REQUIRED for RP-746–748:
proof references need unique anchors, installation needs explicit non-Started
admission substitutions, and failed-before-fence rollback needs stage-dependent
phase coverage. Forward body correction addresses all three in proposal/parent,
including receipt/proof temporal order and historical-reference semantics.
Complete original-plus-correction needs independent re-review; no format/proof
is adopted or implemented, and OP9's native producer/profile gates remain.

Complete original six-path plus corrective five-path union through09c1c792 is
independently APPROVED as a SCOPED SERIALIZATION CONTRACT. Original negative
remains; independent0dd9b8 checks all ranges/references/whitespace and affected
positive/failure/install/retry traces. RP-746–748 close draft defects; RP-745
stays OPEN for OP9/native producers. Next choose/specify and prove a concrete
native supervisor/worker profile and its exact plan/subject/native-receipt types,
including daemon/scope authority, bootstrap and non-bypassable admission. Bind
the actual helper/all mutation-map resources. No format-only acceptance or
universal-refusal implementation substitutes for useful real recovery paths.

**Recovery-contract review,2026-11-04:** independent review at c5419034 finds
RP-741–744: universal healthy-current admission conflicts with partial-source
rollback; Finished lacks whole-helper closure; cleanup has no lawful pre-fence
phase; pending/retried ancestry closure is ambiguous. Forward draft amendment
reconciles CR1–CR6 rather than appending a waiver: separate recovery ownership,
predecessor/completion proof semantics, admitted cleanup→recovery_fenced→destructive
execution, explicit recovery_of/derived recovery, authenticated surviving lock
owner and native-job tracking if selected. This is source/contract work only.
Full amended range needs independent technical review; exact proof/request-log
schemas, first-install/v1 handoff and native producers remain named construction
gates. No boolean/no-op substitute, legacy grandfathering, new OS service,
runtime change, owner adoption or completion claim. Next complete those artifact/
supervisor contracts and actual helper binding, not repeat the approved gateway.

Complete four-pathc5419034..0a89f026 is independently APPROVED as a TECHNICAL
ORDERING CORRECTION. Original CHANGES REQUIRED remains. Reviewer directly checks
ordinary success, mixed-source cleanup, preserved-prior abort, retry ancestry,
derived recovery, lost authority, noncircular prefixes and blocked v1/bootstrap;
scope/whitespace999b10 passes. RP-741–744 close only their normative defects.
Next specify exact hash-bound supervision/work-log/fence/completion artifacts and
their producer/consumer/first-install handoff, with a concrete permission/profile
proposal before any implementation. Runtime, non-bypassable isolation, native
successful/refusal paths and owner adoption remain. No prose check is runtime proof.

**Connection-owned cancellation experiment,2026-11-04:** test a transparent
single-connection gateway between actual pg_dump/libpq and the pinned PostgreSQL16
fixture. Register its in-memory connection BEFORE dialing; retain the protocol's
BackendKeyData in memory only, never output its secret. Three complete populations:
wrong cancellation key while dump is lock-blocked; correct key while equally
blocked; startup delayed8s before BackendKeyData. Kill/wait the actual pg_dump
process, keep draining the server connection, and distinguish cancellation delivery
EOF from actual original-session EOF. Wrong-key delivery must leave the original
query blocked behind its still-held lock; correct-key arm must end that session
BEFORE lock release. Startup must remain registered until terminal server EOF,
even with no key/session row. Same continuously connected control/postmaster and
exact data must survive. One60s guard per population (180s maximum measured work),
<=10s command/handshake within it; cleanup outside work <=15s per command.
Guard/error/unfinished actor invalidates, never raises the bound. Native Go build
via the existing root Make target with an ignored compiler overlay; run only in
a uniquely owned declared tmpfs Compose fixture and cached pinned amd64 PG image.
No production source, Docker socket inside fixture, host port, pull or CI change.
This can authorize a concrete SQL-side candidate, not actual multi-connection Go
backup-helper coverage, durable supervision/registration, an entire operation
fence, native-host acceptance or adoption. Wrong-key requests concern generated
disposable-fixture keys only; no real secret access or key publication.

Actual35493 exits0 with all three populations: wrong-key delivery does NOT clear
the locked query, correct-key original EOF precedes lock release, and key-less
startup remains registered through EOF. Same control PID/postmaster and rows7/42
in each. Manual root-Make overlay controls46325 pass14 subcases after an initial
sandbox listener denial (no logic failure or assertion change). RP-740 tracks
the narrow connection-owned candidate; current candidate document now selects a
fail-closed supervised-operation direction across the whole mutation map, with
durable native ownership, non-bypassable Engine/SQL admission and critical-loss
refusal STILL to specify/prove. Do not infer actual helper/production/native-host
coverage from the one-connection gateway. Complete research/tool/draft batch
needs independent review; original runtime/CI/operator policy remain unchanged.

Complete nine-path20cbb5b8..fd4c5b92 is independently APPROVED after fresh2276
all3 real populations, controls42422 and source/runner/binary bindingsbc06cb.
Exact verdict/cleanup lives in the log. RP-740/738/734 remain OPEN for the full
mechanism, not instrument errors. Current next batch is the complete journal-bound
supervisor/request-ownership protocol: native invocation/boot identity, admission
before side effects, terminal receipts, retained outstanding work and refusal on
critical authority loss. Bind actual multi-connection backup and all mutation-map
resources; a successful toy proxy does not authorize bypassing any of those.
Candidate document includes portable root reproduction commands. No unchanged
experiment repetition or new CI lane is required for this editorial closeout.

**Pre-backup branch experiment,2026-11-04:** preserve the exact live Postgres
container/postmaster and a continuously connected control session while examining
two populations: a libpq client held by `post_auth_delay=8` before ReadyForQuery,
and an actual pg_dump waiting for an exclusive table lock. Observe actual backend
process identity as well as pg_stat_activity; empty activity is not assumed to
mean no connection. Remove only each exact owned worker. For the already-visible
dump, test separately identified backend termination with a bounded wait, then
release the lock and verify unchanged data and the same control connection.
All commands/errors and incomplete populations must fail loudly. Use only the
cached pinned amd64 Postgres16 child, a unique declared tmpfs Compose fixture,
no ports/pulls/production services. One60s measured-work guard, <=10s ordinary
commands; cleanup <=15s PER command outside it. Record actual outcomes even if
the hypothesis fails; guard exhaustion invalidates, never raises the bound.
This tests a narrow SQL-side candidate, not queued creates/new connections,
production session ownership/PID reuse, the actual backup helper/publication,
all pre-backup effects, native-host or CR4 acceptance. No recovery implementation
or new database permission is authorized.

Actual42009 completes both populations at source7091c5cf: startup process live
with census0 and still alive immediately after client removal, later absent
without executing its INSERT; actual pg_dump remains lock-blocked after client
removal, then bounded exact-fixture backend termination removes it. Same DB
container/postmaster and continuously connected control survive, exact rows7/42.
RP-738 adds startup/process lifetime to the protocol; SQL PID predicates are NOT
generation-safe production ownership. Actual-source guard/error controls pass.
Both populations and cleanup finish; no whole recovery or acceptance promotion.
This complete research/tool/draft batch needs independent review.

Original complete49b6113e..83abd009 review is CHANGES REQUIRED for RP-739:
interactive stdin write/flush precedes its local deadline and final reply only
checks global60. Independent real38051 reproduces outcomes, but actual8069e8
accepts an11s handshake. Forward correction uses nonblocking pipe writes and
one deadline across select/write/queue read/final acceptance. Actual25892b
reproduces old acceptance, passes timely/partial writes, rejects exact/late I/O
and propagates ordinary failures. Fresh79437 exits0, both populations unchanged;
v2 binds source7b93d963. Original negative/source/v1 remain. Complete containing
correction requires independent re-review, not a full mechanism/adoption claim.

Complete original seven-path plus corrective seven-path union through ee90913f
is now independently APPROVED: fresh57617 reproduces both real populations and
21d8f5 verifies actual old/new deadline/error/partial/would-block/full OS-pipe
controls and both source bindings. Original negative remains. RP-739 closes its
instrument defect; RP-738 remains OPEN for production startup ownership and
generation-safe cancellation. Do not rerun these completed populations without
a correction. Next specify the durable real-helper ownership/containment protocol,
including the non-destructive pre-backup branch, rather than adopting this
fixture's PID query or treating an empty census as completion. Native-host and
all other whole-operation/release gates remain.

**Interrupted restore boundary, 2026-11-04 (RP-726):** real pg_restore COPY
commits after exact owned client-container removal when its held database lock
is released. The separately identified backend-termination arm prevents that
row but preserves earlier DDL. Draft CR4 now explicitly requires database-side
quiescence too; no production recovery policy is adopted. Both successful
three-arm runs and invalid initial instruments are in the owning log. Five-path
research/draft range retains its original CHANGES REQUIRED for RP-727/728;
the complete corrected union through12d935e5 is independently APPROVED after
fresh six controls and final actual SQL outcomes1/1/0. RP-727/728 close; RP-726's
complete mechanism remains open. Next specify/prove the
complete pending-Engine/client/database fence, not more client-death positive checks.

**Next CR4 engineering experiment,2026-11-04:** test a concrete generation boundary,
not another client-death observation. In a uniquely owned project on actual
Engine28.4.0/Compose2.39.4-desktop.1, pin a pending SQL worker to the full network ID.
After an intact write, disconnect only the owned DB and attempt old-network removal.
If removal refuses, record that refusal and do not claim a fence or start replacement
work. If it succeeds, recreate the SAME network name with a different ID, reconnect
the same DB, then execute the delayed old worker, a newly created old-ID worker,
and a name-only control. Exact SQL rows distinguish late old-ID access from merely
unavailable DB. Preserve all terminal/error outcomes; hypothesis failure is valid
research. Measured work has one absolute60s deadline, each ordinary command at
most10s within it. Cleanup is outside that work deadline, at most15s PER cleanup
command; there is no claimed aggregate15s cleanup or60s end-to-end wall bound.
Complete output requires successful cleanup too. Guard expiry is invalid,
never grounds for raising bounds. Use the same pinned amd64 Postgres16 image as
RP-726. This is a network-admission mechanism study, not already-admitted backend
termination, arbitrary queued Engine-request quiescence, a full restore or native
supported-host proof. No production mechanism/adoption is inferred.

The declared experiment now completes twice: original84016 and final source-bound
88928 both yield exact SQL rows11/14, no12/13. Retirement succeeds despite a
created/not-started ID-bound worker; both that worker and a newly created old-ID
worker fail to start with network-not-found. The name-only worker writes through
the same-name replacement, so database unavailability is not the explanation.
Retained JSON binds final source/image/Engine/Compose and exact error outcomes.
This narrows the candidate to immutable network identity for NEW clients, not a
complete fence. Next map all runtime mutation paths and settle pending-create/
volume/host-file identities plus already-admitted SQL, rather than rerun these
completed controls or promote the narrow network result. Complete original6path
plus2path wording-correction union througha4959f48 is independently APPROVED after
fresh actual Docker/Postgres and source guard/classifier checks. This approves
research/tool/draft reconciliation, not deployment acceptance or adoption.
During independent review, the original shorthand `whole run60s`/`cleanup15s` is
found ambiguous and reconciled above to the ORIGINAL source's exact semantics.
No source, measured bound or retained observation changes; no aggregate elapsed-
time claim is inferred from those successful runs. The owning log preserves the
original wording and review finding rather than backdating a different instrument.

**Whole-operation map,2026-11-04:**
`interrupted-transition-mutation-map.md` now traces every Runtime call from the
existing release/rollback controllers, its nested SQL/HTTP/host effects and
explicitly excluded install/rehearsal/rotation entry points at sourcea0dc53bc.
Drain and authenticated smoke create accounts; backup/restore update host metrics;
the outer release CLI also publishes outcome metrics outside Controller's lock
lifetime, and Start's scheduled backup deletes eligible temporary/backup files;
the production topology has four networks and five persistent volume families.
Upstream Compose stop/remove submits selected IDs, while creation/volume names
need different treatment. These are source findings, not queued-request/runtime
proofs. Do not repeat the completed root-call inventory or independent network/
backend positives. Next construct and prove ONE combined mechanism across the
whole map, with explicit refusal for unknown ownership or surviving mutation.
Original5-patha0dc53bc..12aa01ae is CHANGES REQUIRED for those two completeness
omissions, now explicitly reconciled in map/draft/plan. Complete original plus
corrective5-path12aa01ae..8825019c union is independently APPROVED; original
negative verdict remains. No actual late overwrite/deletion or runtime-policy
adoption is inferred; full CR4/R-006 and owner adoption remain. The owning log
records checked source boundaries and unexecuted runtime/host gates.

**Combined prototype predeclaration,2026-11-04:** build one controlled Docker/
Postgres interruption experiment with three complete populations: parent-only
cancellation; contained clients/removed known workers but reused resource names;
and contained clients plus non-reused resource generations. A private forwarding
proxy must receive complete old Docker requests before interruption, hold them,
and forward them only AFTER the new-operation admission marker. Explicitly this
is proxy-held work, not a claim of delay inside Docker's daemon. Include actual
pg_restore COPY already waiting on a SQL lock, local descendant/daemon host
writers, late container start/create and named-volume deletion. Open all old
gates after admission. Exact SQL rows, current metric/checkpoint bytes and a
replacement-volume sentinel distinguish old work from a dead target.
Use cached config-pinned Python3.12 Linux/arm64 and the existing exact amd64
Postgres16 child on Engine28.4.0/Compose2.39.4-desktop.1. Each independently reset
population keeps a60s measured-work guard (three populations,180s maximum measured
work), <=10s ordinary commands within it; cleanup <=15s PER command outside the
work guard. This is not a performance/RTO budget. Exhaustion/unfinished actors
invalidate that population and the overall observation; no population dropping
or raised guard. No new images/builds/host ports/privileged containers; private
proxy alone accesses the existing daemon socket and forwards only exact owned
fixture requests. No production deployment, game dataset, real receiver, secret
or recovery implementation. The experimental protocol and observed limits must
be reconciled into CR4 before any adoption/whole-mechanism claim.

19985 completes all three predeclared populations: parent-only yields SQL7/13/14/42
and changes/deletes current files/volume; contained-but-name-reusing yields14/42
and current file/volume loss; generation isolation retains only42 and all exact
current values. Actual old COPY, complete proxy-held requests, local/daemon/HTTP
writers and replacement positive controls execute. Both delayed create201 and
old-volume delete204 still occur after admission even in the good population:
this is isolation, NOT literal daemon quiescence. `combined-fence-candidate.md`
now names the concrete remaining production/durable bindings and the incompatible
pre-backup destructive branch. Full real-helper/native-host/adoption/compatibility
remain OPEN. Root actual-AST control finds RP-735: ordinary volume-inspect failure
could be recorded as absence. Original source/v1 observation is preserved98af4c8d.
Correction uses typed actual Engine404, checks owned200 and rejects other replies.
Seven failure/shape/ownership controls refuse; full corrected70426 reproduces all
three outcomes and retains404/404/owned200. Final v2 hash binds exact corrected
source. Original guard/error AST is unchanged and its previous controls apply;
all6 project container/network/volume inventories are independently empty after
both runs. Independent complete original-plus-correction review remains pending.

Independent review of that union is CHANGES REQUIRED: actual-source controls
accept an arbitrary failed late start and collection that crosses its deadline.
Both findings are now corrected: the extra start uses a typed Engine response,
bound to the actual late-created owned image/container and exact old network
refusal; final deadline is checked after all collection and before completion.
Six permission/transport/target/network/unexpected-success controls refuse, and
actual old/new tail execution discriminates late collection. Full final29543
again completes all3 populations, now with typed start204/204/404 and volume
404/404/owned200; v3 binds exact final source. Original negative verdict/source/
v1/v2 remain. Complete corrected-union re-review is pending, not a new mechanism
adoption or acceptance. No bound/population/replacement criterion changes.

Complete original7-path plus both6-path correction ranges through954abe09 are
now independently APPROVED after fresh final40535 all3 real populations and
actual guard/classifier/source-binding07c8c8. Original CHANGES REQUIRED and v1/v2
remain; RP-735–737 close only their instrument defects. The owning log records
exact reviewed union and cleanup. RP-734's controlled isolation evidence is
supported, but the durable production protocol, non-destructive pre-backup branch,
packaged supervisor, all real resources, backup/pair authority and native-host
adoption/R-006 remain OPEN. Do not repeat complete controlled populations absent
a new correction; construct the real protocol rather than promote this prototype.

**Fresh exact candidate, 2026-11-04:** replace neither historical artifacts nor
the predecessor with a fabricated pair. Two independent committed-source
`df84ce1a` rebuilds now match all74 bundle artifacts, six helpers, image archive,
application metadata and seven freshly scanned normalized image inventories.
The new retained `release-builds/candidate-df84ce1a.json` validates; healthy package,
nine actual rebound tampering populations and bound source/image secret scan pass.
Complete four-pathdf84ce1a..593e54b4 is independently APPROVED after fresh
six-binary source rebuild/equality, actual bundle probes and registry checks.
That verdict approves this evidence, not full DP-B/Deployment acceptance.
Continue legitimate floor-compliant predecessor/current-release authority and
supported clean-host proof; neither candidate rebuild is a previous release.
The log owns exact identities, commands, initial setup mistakes and platform limits.

**Actual journal-file collection, 2026-11-04 (RP-725):** extend accepted DP7
verification beyond controlled journalctl output. New manual
`make test-deployment-journal` runs both production collectors against files
written by actual systemd255 journald in an owned networkless Linux fixture.
Eight selected records, native failure-cursor identity, unrelated-service
exclusion, empty/different/missing directories and unknown-job refusal pass.
Complete six-path351124e4..927b26ed is independently APPROVED after fresh actual
native-journal execution and cold owners/vet. This mixed-architecture local
test/tool proof changes no runtime/policy and does not establish host logging-driver,
14-day retention/full AC7/R-006 acceptance. The owning log records exact execution,
initial selector error and sandbox denial. Continue those separate accepted gates.

**Migration/dump consistency correction, 2026-11-04 (RP-724):** actual old-source
regression creates a completed backup that refuses restoration when a schema/version
transaction commits between validation and pg_dump. Check migration in one read-only
repeatable-read transaction and export its snapshot to pg_dump. Final real restore,
initial-mismatch/failure/cancellation and observed-lock-release controls pass with
the eight prior integration functions. This is an accepted DP6 conformance fix;
no format/crypto/retention/RPO policy changes. Complete six-path0179415e..9cf66afb
is independently approved after fresh all-nine SQL functions and exact historical
production overlay failure. Root's additional real Caddy release/rollback target
passes2.54s at unchanged source. The log preserves all provenance and the invalid
first synchronous-cancellation oracle; no full Deployment/clean-host promotion.

**Backup metadata proposal, 2026-11-04 (RP-505):** complete seven-path
09cf45e5..ce61b82a is independently TECHNICALLY APPROVED, ready for owner
adoption. Fresh native and declared Linux diagnostic populations both reproduce
two unsafe flag-based deletions; those passes are diagnostic execution, not
retention acceptance. Signed outer v2, external trust and authenticated legacy
copy remain unadopted; no production/key/format change. Governed resolution,
D-019, snapshot-based RPO and clean-host gates stay separate. Exact scope/evidence
is in the owning log. One concrete adoption question is delivered2026-11-04;
no answer is recorded. Implement only after adoption; do not repeat the question.

**Actual game-process availability proof, 2026-11-04 (RP-721):** construct the
missing DP7 producer-to-receiver path using the current gameserver executable,
real Postgres and existing private monitoring/Caddy stack. Actual process stop,
unchanged five-minute rule, exact alert resolution after executable restart and
owned cleanup are required. It replaces no production code or rule and does not
claim full AC7/clean-host acceptance. Complete eleven-patha636b315..b23a139b is
independently approved: fresh FINAL-source actual23396 passes401.085s, including
exact original-alert resolution and owned cleanup. Cold owners/vet and all22
observer controls pass; private guard omission reproduces both cancellation
failures. Root's earlier service pass remains separately attributed. Continue
host/journal/retention, older restart-history and exact-release work.

**Journal collection correction, 2026-11-04:** DP7's measurement command now reads
the requested directory and refuses exit-zero diagnostics before emitting evidence.
Original six-pathb6744226..c09bb854 retains CHANGES REQUIRED: newly supplied
--quiet hides access warnings. The complete original-plus-seven-path follow-up
c463a5bb..e3ed4467 is independently approved. Actual old/current flag-aware child
controls discriminate both commands; twelve corrected native/Linux-amd64 arms,
cold owners and vet pass, preserving empty cleanup collection. No host-retention
or exact-artifact acceptance follows; the log owns exact scopes and limits.

**Backup parser/snapshot corrective review,2026-11-04:** complete eight-path
`630f6833..bc5d5f5d` (RP-489/490) is independently approved. Fresh exact-old parser
discriminates all eight malformed cases; corrected14 arms, cold four owners/vet
and all eight real Linux-amd64/Postgres backup functions pass without skips,
including concurrent snapshot/cancellation/lock-release observations. The log
separates fresh execution from historical old-SQL evidence. Separate dump binding,
production RPO/RTO, metadata authority and clean-host/full-feature gates remain.
DP7 writer/storage RP-521 has complete independent approval too. Original receiver
RP-523 retains CHANGES REQUIRED, and its complete original-plus-six-path
RP-714/715 follow-up union is independently approved throughb0fd04ad after actual
restart-profile execution, cold owners/vet and race. Do not repeat completed
backup/storage captures or call old storage evidence execution of the new receiver.
Scheduled-backup and ingress proof ranges also have complete independent approval
(latest log): fresh actual retry/restore and stopped-Caddy pending/firing/recovery
populations pass. Continue remaining accepted game/host/journal/retention work;
original full AC7, exact release pair and clean-host gates remain open.

**Drain/restart corrective checkpoint, 2026-11-02:** independent review approves
RP-138's complete `1908762f..c9c9dde7` marker correction and RP-143's complete
`8f83e4c5..81a970ce` removal of the false admitted-work acceptance route. Cold
affected/race/vet checks and three private overlay failures execute. The existing
Linux-amd64 Caddy/Postgres release lane passes2.73s, including marked drain,
held intent completion, restart and encrypted restore. Actual CLI refuses the
unsupported admitted-work row with exit2. The log owns full hashes and limits;
original R9/R22, real crash-during-write, exact-release/clean-host and full
Deployment acceptance remain open. These corrective scopes no longer await Claude.

**Safety corrective review checkpoint, 2026-11-01:** fresh independent Codex
approves complete RP-466 atomic backup commit, RP-544 live-writer cleanup,
RP-545 shipped-copy attribution and RP-481/RP-482 rehearsal-baseline ranges.
The owning log holds each exact range/reviewer and executed checks. Actual Linux
filesystem/cross-process tests and all eight current encrypted Postgres backup/
restore functions pass; client metadata is regenerated from a fresh build.
Only these corrective reviews close. Original R6/R8 rejections, other ranges,
current-release/backup-metadata authority and exact-artifact clean-host gates
remain separate; no full Deployment or release approval.

**Current corrective reviews, 2026-10-15:** fresh independent Codex approves complete
RP-503/RP-504 duration/recovery-clock range, RP-550 real bundle-gate tests and RP-552 exact
restore-refusal fields. Full hashes/provenance and executed cold/race/Linux/real-Postgres
results are in the owning log. These scopes no longer wait on Claude. They do not close
original R18/R20, backup metadata authority (RP-505), current-release identity, exact-artifact/
clean-host R-006/R-007, or full Deployment acceptance. Continue remaining accepted work.
Further fresh independent approvals cover complete RP-507 authenticated-stream/deadlock,
RP-508 pre-destructive rollback package validation,
RP-548 create-report invocation and RP-549 sequential-observation scopes. The owning log
records actual clean-target SQL refuses/restores, local producers/loaders and exact limits;
original author/artifact/feature holds are unchanged.
Complete RP-510 dump-creation and RP-511 archive-parser ranges also have fresh independent
approval; actual positioned Postgres restores and declared Caddy release checks execute.
The owning log records exact source contexts/limits, not installed-release acceptance.
RP-518 same-second backup outcome/rule correction is also independently approved across
its complete range. Original real-service firing, matching recovery and receiver refusal
execute; this is controlled-outcome evidence, not a scheduled Postgres backup or full AC7.

**Current DP7 timed-observation correction, 2026-10-14 (RP-615):** the real-minute test's
startup/failure/drain callback waits now obey its existing95-second context and refuse missing
or late observations. Three factored-old receive regressions fail; corrected cold/vet/race
controls and the actual Linux-amd64/Postgres timer population pass, retaining real60-second
callbacks, SQL failure/recovery and orderly drain. This changes tests only, not credential
retention, timer, runtime or alert policy. Independent corrective review now APPROVES
`ea194d8d^..ea194d8d`; the log owns the exact range and actual real-minute rerun.
Current-head manual evidence is not full AC7 or exact-artifact/clean-host proof.

**Current DP5 smoke correction and release-lane evidence, 2026-10-14:** RP-612 rejects
size-limiter EOF as completion, preserving the existing 4 MiB ceiling. Four old decoder/caller
regressions fail; corrected native/race and the actual declared Caddy/Postgres/WebSocket lane
pass, including admitted work, courtesy/drain, restart and restored-account smoke. The log
records independent corrective approval of `cc256422^..cc256422`, the actual declared-lane
rerun, and its checkout-backed/emulated limits.
Current Docker28.4.0/Compose2.39.4-desktop.1 matches the recorded development pair but its
daemon is Linux/aarch64; CR2/CR4 native supported-host experiments remain unexecuted.

**Current DP6 cleanup correction and restored test environment, 2026-10-14:**
RP-611 binds cleanup to inspected directory/file identities and a directory-rooted handle.
Three old physical-file regressions delete an uninspected replacement; corrected native and
Linux-amd64 populations preserve it, including unchanged aliases and second-process writer
exclusion. Cold affected packages/vet/race and Linux helper build pass. Docker now responds;
`make test-deployment-backup` executes all seven real Postgres integration tests, including
RP-551's empty/populated query-override restores. Exact evidence/range after `68eee56b` is in
the owning log. Independent corrective review now APPROVES RP-551/RP-559/RP-611
and the stale-doc supplement, with exact ranges and actual SQL/cleanup evidence.
Other previously unavailable lanes must run individually;
no daemon intervention, current-release proposal adoption or clean-host/R-006 pass is inferred.

**Current backup-input correction, 2026-10-13 (RP-559):** malformed URL queries now
refuse before service/pass file creation, rather than silently dropping invalid fields
and preparing fallback credentials. Twelve actual old-helper failures and four encoded-value
controls precede the correction; all credential-file checks/race, four cold affected packages
and vet pass. The log records the initially falsified cross-parser hypothesis and exact range
after `0e7a4b95`, now independently APPROVED through4966e296. The separate RP-551/RP-611
review executes actual local Linux/Postgres restores, not general URI compatibility,
hosted or exact-artifact acceptance; those boundaries remain open.

**Current contract work, 2026-10-13 (RP-558 / DESIGN-GAP10):** the
[Current Release Authority draft](../../rfc/deployment-current-release-authority.md) specifies
live image/mounted-byte identity, one ledger with durable release/rollback events and explicit
interruption recovery. It is not accepted/acceptance-ready; CR2/CR4 require the two named pinned-
engine experiments before mechanical closure and adoption/review. Native actual ExecRunner/lock
evidence confirms child continuation after killed-parent lock release. New bounded Desktop
experiments read all nine mounted roots from four stopped exact-image containers and reject
changed/missing/link/interrupted inputs; actual daemon exec also completes after BOTH helper
and CLI death/released lock. Sources are preserved outside CI. The daemon is Linux/arm64,
not native supported-host closure. No production reader/recovery mechanism or product byte
changed. The 2026-10-16 actual Compose one-off experiment also observes job completion after
the entire exclusively owned local CLI process group is killed. Process-group cancellation
alone is insufficient; the controlled shell is not a backup/restore or delayed-request proof.
Next: settle daemon-side quiescence before the draft can be accepted, and retain
native-host seven-service/metadata coverage rather than calling the Desktop sample acceptance;
continue other accepted work meanwhile. D-019, other DESIGN-GAPs and full nine-tier floor remain.

**Current operator-state admission, 2026-10-13 (RP-556/RP-557):** lifecycle commands now
reject existing inadmissible history/start times before runtime operations; lock/append paths
reject final directory symlinks before changing their targets. Old regressions and all30
corrected physical-file/recording-runtime cases, focused race, four cold packages/vet and Linux
helper build pass. Independent review APPROVES b448069b..b29fc293; not real Docker/SQL/R-006.
**Next engineering contract: RP-558 / DESIGN-GAP10**—define the actual current-release identity
binding, including missing history and failed transitions, before claiming forward-only live
upgrades. The current controller admits a contradictory install-history diagnostic; no live-host
downgrade is proved. Preserve existing accepted behavior until its contract closure is adopted;
this is agent specification work, not a request that Marco design a technical parameter.

**Preceding rotation chronology correction, 2026-10-13 (RP-555):** activation/removal refuse
before appending a timestamp the existing reader would reject. Six old-source poisoned-ledger
cases and12 equal/later controls establish the regression; all18 corrected cases preserve
readability and valid retry, across all three key families. Cold four-package tests/vet,
focused race and Linux-amd64 helper build pass. Independent review APPROVES3964194b..b448069b;
overlap minima, cursor contract holds and exact-host rotation proof remain unchanged. Original
R1's full six-path CHANGES REQUIRED disposition remains; other Codex descendants still
need review. Continue accepted construction without relabelling native proof R-006.

**Preceding failed-install correction, 2026-10-13 (RP-554):** `AbortInstall` now retains its
normalized runtime rather than dropping the default command runner and rotation overlay.
Three old-source failures and five controls precede the correction; all eight now pass,
including actual ExecRunner subprocess failures through an owned recording executable.
Cold four-package tests/vet, focused race and Linux-amd64 helper build pass. Original R2's
complete nine-path disposition is CHANGES REQUIRED (RP-134/RP-554), not approval of Codex
corrections. Independent corrective review APPROVES b868cd2c..3964194b; Docker/SQL failed-install and
exact-artifact clean-host proof remain open. No change to rotation timing or cleanup targets.

**Current DP-F refusal-record correction, 2026-10-13 (RP-552):** the non-clean restore classifier
now rejects duplicate and case-aliased members rather than allowing Go's overwrite/case-fold
semantics to create valid-looking evidence. Sixteen old-parser regressions and two actual-probe
orchestration counterexamples fail; corrected cold affected packages/vet, race and Linux helper
build pass. Full original R20's eleven-path disposition is CHANGES REQUIRED (RP-142), not approval
of its corrective descendants. Complete `fc04db24..a713137e` is independently approved.
The earlier RP-142 provenance correction `94f0ee74..d3fde4c0` also now has its own
independent corrective approval (2026-11-02 log), including historical corrected /
broken-classifier overlays, actual shell-exit controls and cold affected checks.
Actual composed CLI stderr and exact-host/R-006 gates remain; valid records stay admissible.

**Current DP6 endpoint correction, 2026-10-13 (RP-551):** service/pass files now honor the
database URL's endpoint query overrides instead of targeting its unoverridden URI fields.
Old-source20-case population fails18, retaining ordinary/password-isolation controls; corrected
focused/race, affected cold packages/vet and Linux helper build pass. The existing actual-Postgres
empty/populated restore test now includes swapped URI paths with database-query overrides; it
initially only compiled/skipped locally. The 2026-10-14 declared Linux/Postgres run now passes
both empty/populated override populations; the latest owning log records it. New range after
`a7501ac1` throughfc04db24 is now independently APPROVED. This
does not close DP-C/AC4, general URI options, original R20 review or exact-host/R-006.

**Current DP-F producer evidence, 2026-10-13 (RP-550):** all nine packaged-artifact mutations now
execute against the actual assembled-bundle validator, with18 invalid-baseline controls. The
content-identity omission fails the new epoch witness while the prior fake-validator test passes.
Cold affected packages/vet and focused race pass. Complete test-only range
`5d70f673..a7501ac1` is independently approved; original R18's complete five-path disposition remains CHANGES REQUIRED (RP-141),
not approval of Codex's later witnesses. Synthetic image/client/dump inputs and recorded runtime
commands are not exact-artifact/clean-host R-006, or permission to execute external deployment.

**Current DP-F chronology correction, 2026-10-13 (RP-549):** sequential command results must not
overlap or run backwards, even inside one lifecycle step. The executor now refuses before the
next producer launches; base/final validators reject rehashed overlapping results. Ten focused
cases preserve exact adjacent boundaries and forward gaps; old-source failures and corrected
cold/race checks are recorded in the owning log. Complete `88868d34..5d70f673` is independently approved.
Original R12's full five-path disposition remains CHANGES REQUIRED (RP-140); basename/command
shape is not executable identity, and no synthetic result or seal closes R-006.

**Current DP5/DP6 creation correction, 2026-10-13 (RP-548):** newly created backup reports must
place start/completion inside their command invocation; invalid clocks refuse. Old-source20-case
population admits14 bad reports while six legal controls pass. Corrected actual encrypted
envelopes, cold affected packages/vet, race and Linux-amd64 helper build pass; existing host/header
oracle still rejects its isolated bypass. Complete `dac25388..88868d34` is independently approved. Original
R10's full four-path disposition remains CHANGES REQUIRED; D-019, authenticated metadata/snapshot
meaning, logical restore, clean-host/RPO/RTO and exact-artifact gates are not resolved by this check.

**Current DP3/DP5 correction, 2026-10-13 (RP-546):** invalid receiver/readiness request syntax
now refuses rather than panics; configured receiver construction is checked before runtime
commands. Old-source eight failures, ordinary receiver/ready/drain/identity controls, affected
cold packages/vet/race and Linux-amd64 helper build pass. Original R9's complete five-path
disposition remains CHANGES REQUIRED; independent corrective review APPROVES complete
39019517..de0ca191, with exact provenance/executed evidence in the log.
RP-547 now adds actual `DockerRuntime.DrainCurrent` coverage through native HTTP/WebSocket and
recording commands: eleven success/failure cases, two caller-level falsification overlays,
affected cold packages/vet and race pass. Independent corrective review APPROVES complete
de0ca191..dac25388. No production behavior or drain bound changed; synthetic identity/exit replies are not
real process, Caddy/Postgres/Linux or clean-host/R-006 proof. Continue remaining accepted original
reviews and exact-artifact construction while those environment/review gates stay explicit.

**Current DP1/AC8 correction, 2026-10-13 (RP-545):** every shipped npm copy's license/notice is
validated before name/version deduplication, including CSS-derived copies. Matching attribution
merges; conflicting/missing evidence refuses without partial inventory. Two-route old-source
negatives, delivered-notice controls, cold affected packages/vet and the freshly built client's
actual53-dependency metadata pass locally. The original Go helper-dependency guard is restored.
Original R8's full seven-path disposition remains CHANGES REQUIRED; this correction
`fa2bec58..39019517` is now independently approved. Prior map/CSS repairs retain
their separately recorded dispositions. No classifier/policy/schema, image
provenance, clean-host, full AC8 or release promotion.

**Current DP3/AC2 correction, 2026-10-13 (RP-543):** archive secret scanning now
includes decoded tar owner/group/PAX/xattr fields and gzip header metadata; recognized
secret-bearing finding/error paths are redacted. Digest-valid old-source negatives and
clean controls discriminate; cold packages/vet and an ordinary scan of the retained
102face6 image pass. Its bytes remain historical and its embedded tool lacks this fix.
The log records the complete original R3 five-path CHANGES REQUIRED disposition and the
new correction completecf8c6f08..ca447daa now has independent CHANGES REQUIRED: later gzip
member headers are missed (RP-675). The correction afterbdf5add1 scans every header while
joining payloads under the unchanged combined ceiling; old six misses/four controls,
corrected affected native/race/Linux and ordinary historical-image scan belong in the log.
Fresh independent corrective review APPROVES completebdf5add1..76cba7bc, resolving RP-675
without relabelling the original negative verdict. Rebuild/re-scan the final
reviewed artifact; no new scanner rule/limit, R-006, full Deployment or1.0 acceptance.

**Latest DP7 actual proxy outage, 2026-10-13:** the shipped Caddyfile's private
route/scrape now has a physical stop→real pending five-minute rule→receiver
firing→same-container restart→matching resolution witness. Upstream readiness is
an explicit control, not a gameserver proof. Root lane passes406.67s, cold affected
native/vet and Linux-amd64/Postgres checks pass. The log owns the range after
`12399dc9`. Complete eight-path12399dc9..9b48c9c6 now has independent approval;
fresh actual profile6671 passes398.132s, including the unchanged five-minute bound.
Continue actual game/host observation,
journal/retention, older restart history and exact-artifact evidence; no full AC7.

**Latest DP6/DP7 actual backup execution, 2026-10-13:** the production CLI now
has a joined actual Postgres dump→scheduled target refusal→receiver firing→real
five-minute retry→matching resolution→fresh-database restore witness. The retry
restores a deliberately updated account row, not the old dump. No direct backup
metric or fake dump is used in this population. Shared fixture metadata remains
synthetic, and the log owns final checks/range after `c4279b66`. Complete twelve-path
c4279b66..12399dc9 now has independent approval; fresh actual profile20190 passes
372.264s, including five-minute retry and updated-row restore. Continue actual
game/host observation, retention, older-container restart
history and exact-artifact evidence; this is not off-host/RPO/RTO or full AC7/R-006.

**DP7 restart receiver construction, 2026-10-13, review finding2026-11-04:** the actual Docker
observer's first-scrape count3/increase0 now fires through the private services
and reaches the receiver; an actual replacement count0 resolves the same alert.
The final root lane also proves canceled-CLI one-off cleanup. Its early-notification
loss finding (RP-523), failed attempt, final checks and range after `cdc78050`
live in the log. Complete original `cdc78050..c4279b66` requires changes: its
requeued early notifications can be overtaken by later resolution, and a canceled
ready control can falsely pass. The complete original-plus-follow-up union is now
independently approved throughb0fd04ad: RP-714/715 isolate nonce delivery from the
unchanged real-alert FIFO and check cancellation; fresh cold/race and actual restart
44339 pass137.936s including matching resolution/owned cleanup. No production
bytes or alert policies change. Continue remaining actual backup/game-process/host
journal/retention, older-container restart history and exact release-artifact
evidence; full AC7/R-006 remain open.

**Latest DP7 restart correction, 2026-10-13 (RP-522):** actual Docker count and
container birth now reach the host textfile together. A new-container lifetime
arm catches three restarts before the first scrape, preserving the original
counter-increase rule and ten-minute threshold. Old rule regression fails;
corrected pinned rules and actual automatic-restart/observer witness pass.
Independent Codex corrective review APPROVES complete `db752b0a..bd91a9a6`;
the log records actual Docker count0→3 and pinned amd64 rule execution.
Continue live producer-to-receiver restart proof, actual journal/retention and
exact release-artifact construction; no full AC7/clean-host/1.0 completion.

**Latest DP7 storage construction, 2026-10-13 (RP-521):** two distinct20MiB
filesystems feed the actual shared production usage reader/textfile writer.
Exactly80% stays quiet;85% fires through node-exporter and reaches the receiver;
freeing the owned file delivers matching resolution on each storage subject.
The focused population passes286.35s. Final default profile selects both named
populations and passes681.05s/644.86s respectively, each within its unchanged12m
guard; longer combined storage setup is recorded without attributing its cause.
Old NaN writer-input cases fail; all six corrected nonfinite cases preserve the
last valid file/mtime. Cold native/vet and declared Linux-amd64/Postgres packages
pass. Complete `82c8e57b..db752b0a` is independently approved2026-11-04 with fresh
physical storage44377 at captured sourcee1aa4c50:305.83s, both subjects pass.
That binary predates the receiver follow-up and does not prove its correction.
The log owns exact evidence. Continue
actual host-command/journal/restart/retention and exact release-artifact/R-006
construction; no CI change, new owner policy or full AC7/1.0 completion.

**Current bounded DP5/DP7 correction, 2026-10-13 (RP-520):** the real one-minute
cleanup timer exposed a false failure event when normal drain cancels its waiting
SQL. Observation now retains the original error and records canceled execution
without advancing success or emitting an alarm. Real SQL errors still withdraw
readiness and emit failure; a joined SQL/cancellation error is not suppressed.
The old real-Postgres drain arm fails; the corrected two-arm timed lane passes
121.85s, including runtime-error recovery. Cold affected native/vet and declared
Linux-amd64/Postgres packages pass. This timed lane is manual, outside push CI;
the fast adapter regression remains in ordinary tests. The log owns evidence and
complete `5fcb5353..82c8e57b`, independently APPROVED with the preceding RP-519
range. Continue remaining actual
release-host observation/retention and AC7/R-006 construction; no new owner policy.

**Current bounded DP7 correction, 2026-10-13 (RP-519):** fail-closed cleanup remains
unchanged, but the command now emits a bounded failure event before returning the
job error. The existing host observer consumes the retained journal and persists
new detection independently of the gameserver endpoint; re-reading does not extend
the alert window and incomplete collection cannot refresh health. Real composed
SQL and actual development-command startup witnesses pass. The private full stack
passes in700.28s within the unchanged12m guard, including journal-driven firing and
matching resolution with the fresh process counter0. Journal envelopes/timer are
modelled there: actual release-host driver/command, rapid-exit timing, retention,
remaining alert producers and AC7/R-006 remain open. The owning log holds checks,
regressions and independently APPROVED complete `f0009bfc..82c8e57b`, including
the diagnosis and RP-520. Continue the remaining release-host engineering proofs.

**Latest DP7 acceptance construction, 2026-10-13:** real transport dead-letter
growth now reaches the private fixture's configured receiver. Current SQL event
trigger → real Node encoder refusal → five relay retries → actual collector →
shipped rule → receiver firing/matching resolution; isolated growth stays quiet
and stopping growth retains history. The full fixture passed in 643.69 seconds
within its unchanged 12-minute guard; affected native/vet and Linux-amd64/Postgres checks pass.
A collector-zero compiler fault fails the focused real-SQL oracle. The owning
log holds the new range after `7d0aa087`, pending designated review. Continue
remaining producer/retention/release-host evidence; no full AC7/archival claim.

**Latest DP7 acceptance construction, 2026-10-13:** the private fixture now also
tests actual connection refusal, distinct from the preceding blocked-query arm.
A separate maintenance database disables access to only the disposable subject
and terminates its existing sessions; the real pool reports both health signals0
while metrics stay reachable. The unchanged2m rule delivers firing, and access
restoration makes that same pool reconnect and delivers matching resolution.
Full fixture60688 passes511.42s; focused/full Linux-amd64 SQL and native/vet checks
pass, and a false-health compiler overlay fails the new oracle. All work is test/
fixture/docs only. The new range after `391ae4a5` needs designated review and can
consolidate with the two preceding operations batches. No process-restart/network
partition, seven-family, retention, clean-host, whole-CI or archival claim.

**Latest DP7 acceptance construction, 2026-10-13:** the private operations fixture
now owns a disposable, current-schema Postgres16 database instead of a fake
unavailable database. Its real count query is blocked by an exclusive table lock:
ping and scrape stay healthy, collection reports failure, the unchanged two-minute
rule reaches the configured receiver, and lock release delivers matching resolution.
The full fixture passes330.07s, including backup fire/clear and receiver rejection.
The focused actual-SQL failure/recovery check passes cold on Linux-amd64; a compiler
overlay falsely returning collection success makes it fail without touching production
source. This is the collector-failure arm, not a Postgres process/network outage or
the complete seven-family release population. The owning log holds the new range
after `b4013e99`, consolidatable with RP-518; designated review and broad DP-E/AC7
acceptance remain open. Continue the remaining producer paths and release-host proof.

**Latest bounded correction, 2026-10-13 (RP-518):** helper observation now keeps
an atomic bounded last-result gauge alongside the unchanged historical timestamps.
The backup alert no longer loses a failure in the same whole second as success;
legacy newer-failure and six-hour deadline arms remain. The actual scheduler
encrypted-target regression, cold affected native/vet and Linux-amd64/Postgres
packages, and exact pinned rule evaluator pass. The private Compose fixture also
proves production textfile→node-exporter→Prometheus→configured receiver firing
and matching resolution, preserving the real receiver-rejection check. This uses
controlled outcome/clock input, not a Postgres backup command or clean release host.
The owning log records the range after `24a343b8` requiring Claude review; all
other DP-E/AC7 and Deployment acceptance/archival gates remain open.

**Historical predecessor reconstruction (2026-10-13, RP-514):** exact source44a4a72
now has two independently rebuilt, byte-identical70-artifact trees, six fresh image
SBOMs per build and corrected Linux/amd64 config identities. This restores artifacts,
not a supported rollback target: current validation refuses the old Compose/Caddy
floor (including its enabled2019 management API). The candidate validates, but the
actual pair probe exits2 `invalid_probe`. The retained
`release-builds/reconstructed-44a4a72.incomplete.json` deliberately marks the supported
predecessor objective incomplete. Establish a reviewed floor-compliant predecessor
with explicit provenance before full pair/rollback proof; do not weaken safety checks,
relabel the historical source or repeat this already completed reconstruction. The
owning log records exact output identities, exclusions and pending designated review.

**Current local candidate construction (2026-10-12):** two independently built candidate trees at
`102face6` compare byte-identical, including74 artifacts, six Linux commands, no-cache image
archives and seven independently scanned normalized image SBOMs. New retained record:
`release-builds/candidate-102face6.json`; the old candidate/previous records stay historical.
RP-512 corrects two ARM64 identities in the old records for this new build, not those old
bytes. A manifest-bound source/image scan and nine real-bundle mutation probes pass.
This is local construction evidence, not a supported candidate/previous rollback pair,
current CI, independent review, clean-host R-006 or any1.0 release acceptance.

**Latest bounded correction (RP-513):** archive validation now binds ordered actual
uncompressed layer hashes to the config's `rootfs.diff_ids`, rather than merely checking
that layer names exist. Ten old-source negatives fail; corrected plain/gzip controls and
compression-error refusals pass. Affected cold native/vet and Linux packages and real
Postgres/Caddy integration pass. Both retained source102face6 candidate bundles still
validate under the corrected native helper, but their bundled helpers do not contain this
later repair. Rebuild at the final reviewed source; do not relabel those immutable artifacts.
Fresh independent Codex review APPROVES the complete13621c7d..c7071178 corrective range,
including independently reproduced old accepts and actual retained archive validation.
Native SQL/Caddy skip is not integration evidence. The owning log holds exact provenance
and limits; supported predecessor, current-source rebuild and clean-host gates stay open.

**Latest bounded correction (RP-511):** malformed gameserver archive headers now return
invalid content rather than panicking in the validator/assembler. Eight old-source failures
become refusals; intact controls and all twelve malformed-input cases pass. Affected cold
native/vet and Linux packages plus existing real Postgres/Caddy integration pass. The log
owns independent approval of complete `abf4b4ac..102face6` and fixture limits. Six Linux binaries, client, metadata
and content build at `abf4b4ac`, but these pre-correction outputs are not a current full bundle.

**Latest bounded correction (RP-510):** both package-creation passes now bind the
same complete dump under DP6/AC4. Four old-source regressions fail; rewind before
hashing plus streamed hash/EOF validation prevents publishing an inconsistent
package. Actual positioned empty/populated Postgres dumps restore with matching
recovery identity; affected native/vet and Linux checks pass. No format, retention,
clock, key or CI contract change. The owning log records results and the full
independently approved complete `e27eb832..abf4b4ac` corrective range; broader gates remain open.

**Latest bounded correction, 2026-10-12 (RP-508):** rollback input preflight now uses the
actual authenticated package decoder before stopping services/resetting Postgres. The old
method admits encrypted non-packages and packages with trailing data; focused controls now
also reject a changed dump with its old internal hash. Plaintext workspaces are private and
removed on return. The R18 missing-input producer now creates a full package, retaining its
healthy-first control and synthetic-dump limitation. Real Postgres/Caddy and affected native/
Linux checks pass; complete `cd609d07..a5a2e94a` is independently approved, including actual
declared release-lane execution. D-019's separate per-release
restore choice, logical restore/RTO, clean-host and broader Deployment gates remain open.

**Latest bounded correction, 2026-10-12 (RP-507):** package extraction now consumes the
authenticated stream after tar EOF and refuses any remaining bytes before returning a usable
dump. The old implementation stalls for one extra byte, a zero block and a second archive;
the exact-package control passes. The existing empty/populated Postgres restore population
also proves refusal leaves its target clean before the intact backup restores with matching
identities. The owning log records independent approval of complete `4627861a..cd609d07`. No
metadata-binding draft adoption, retention/format change or clean-host acceptance follows.

**Latest bounded correction, 2026-10-12 (RP-504):** populated recovery records its
start after input/lock/checkpoint validation but before stopping the failed stack and
resetting/restarting Postgres. Those required phases now count toward the existing4h
RTO. Four old-source failures become passes: complete elapsed time, stop/reset overruns,
and invalid-clock refusal before destructive work. Affected cold packages/vet and declared
Linux-amd64 recovery tests pass. Controlled runtime evidence, not an actual clean-host
restore; the owning log holds independent approval of complete `ad54b258..09cf45e5` with RP-503.

2026-10-12 bounded retention integrity diagnostic at `09cf45e5`: two independent real
encrypted backup populations, each with a recent normal backup and31-day unresolved
pre-upgrade backup. Intact controls must retain both and restore the protected dump.
Change only the outer `pre_upgrade` or `upgrade_resolved` boolean, preserving ciphertext/
length/checksum. Observe actual ReadHeader, Restore, PlanRetention and ApplyRetention;
record protection failure when cleanup deletes the protected local fixture despite an
authenticated restore refusal. No real backup/operator data is used. This is a diagnostic,
not retention acceptance: report unsafe-case census explicitly, preserve existing policy/
format/key placement, and route a confirmed authority gap rather than invent a new key.

Diagnostic completed: both intact controls retain/restore; both outer-flag mutations
are admitted by classification and delete the protected fixture. RP-505 remains an
open DP6/AC4 defect. The metadata-binding draft now proposes exact signed-v2 metadata,
external trust, target-owned cleanup and explicitly authenticated legacy copying.
Adoption/normative review are pending; the missing resolution lifecycle remains a
separate author hold. No implementation authority follows from this proposal. Owning
log records diagnostic execution, design inspection and pending review ranges.

**Latest bounded correction, 2026-10-12 (RP-503):** objective measurement now applies
DP6/AC4's exact six-hour/four-hour bounds before whole-second report conversion. The
12-case regression exposes four old-source subsecond-overrun false passes and passes
after repair, including zero/below/exact/whole-second-over controls. Affected cold native
packages/vet, declared Linux-amd64 objective tests and all seven real-Postgres backup
integration functions pass. The owning log holds independent approval of complete
`ad54b258..09cf45e5` with RP-504, including current actual declared backup execution.
Report fields and clock origins are unchanged; no full Deployment, CI or clean-host claim.

**Latest bounded checkpoint, 2026-10-12:** original Claude R5
`78ec39a7^..78ec39a7` has bounded Codex designated approval. The unchanged private-service
witness executes through `make test-deployment-operations-profile`: five healthy scrapes,
cleanup firing, received nonce, HTTP-500 delivery refusal and peer admin isolation.
The exact pinned Prometheus index's amd64 child passes rule evaluation; the affected actual
Linux-amd64/Postgres startup/job test and cold packages/vet pass. The owning log discloses
the mixed-architecture fixture, Docker digest-cache collision and rejected empty test
selection. The new Make extraction needs Claude review after `ffc8bc25`; no aggregate,
seven-family end-to-end, DP-E/AC7 or clean-host approval follows.

**Preceding bounded checkpoint, 2026-10-12:** RP-490 reads each recovery identity in one
read-only repeatable-read transaction. A real locked-table/concurrent account-write
regression fails on the old implementation; all seven declared Linux-amd64/Postgres
backup integration functions now pass, including cancellation with no partial identity
and observed lock release. Affected cold Go/vet pass. The owning log records the
corrected cancellation observer. The complete shared `630f6833..bc5d5f5d` range
is independently approved2026-11-04; fresh evidence is at the top of this plan.
Separate dump snapshots and full Deployment/clean-host gates remain.

**Preceding bounded checkpoint, 2026-10-12:** RP-489 unifies backup-file header admission with
the existing single-object decoder and bounds outer/decrypted metadata framing.
Eight encrypted malformed arms fail on old code; final14 arms retain healthy,
whitespace and exact64KiB controls. Affected cold Go/vet and all six declared
Linux-amd64/Postgres backup integration functions pass. The owning log holds the
range after `630f6833`, now independently approved throughbc5d5f5d with RP-490.
Full Deployment and clean-host gates remain.

**Bounded checkpoint, 2026-10-09:** RP-482 adds healthy baselines to the public-metrics and
source/image-secret probes. Actual assembled-bundle tampering and always-rejecting gates
fail the old checks; affected native/Linux-amd64 packages pass after correction.
Complete combined `1efea9cf..1b484c50` now has independent corrective approval;
no DP-F acceptance follows.

**Preceding bounded checkpoint, 2026-10-09:** RP-481's interrupted-writer probe now requires a real
encrypted create/restore baseline and the exact consumed dump failure, with no partial files
or damage to the baseline. Old-source regression and affected native/Linux-amd64 checks
pass as recorded in the owning log. Complete combined `1efea9cf..1b484c50` is now
independently approved; this does not close DP-C/DP-F or substitute for the actual
clean-host rehearsal.

**Preceding checkpoint, 2026-10-08:** RP-423 corrects false-clean restore targets containing
schema-only or nonrelational objects. At `e478b248`, the declared Linux/amd64 Postgres16 backup
target now executes all six integration functions without skips, including real encrypted
empty/populated dump/restore, recovery identity, non-clean-target refusal and pre/post-deletion
backups. Its isolated project is removed by the target. This supersedes the earlier missing
full-backup execution, not the broader Linux release preflight failures or private operations
composition gate. Pre-deletion restore still revives the deleted account (RP-125), as the test
explicitly records. Original R20/DP-C, measured RPO/RTO, exact release/clean-host criteria and
designated review remain open; the owning log holds the bounded result.

## Delivery sequence

**Current bounded R6 correction, 2026-10-13 (RP-544):** live backup writers own shared directory
locks; aged-temporary cleanup requires exclusive ownership and rechecks the inspected target.
Busy ownership defers scheduled rounds without false result metrics. The actual second-process
regression, encrypted restore, orphan cleanup, scheduler retry and concurrent no-overwrite cases
pass locally, with cold affected packages/vet/race. Independent review now approves
`ca447daa..fad86116` and executes actual Linux-amd64 filesystem plus real encrypted
Postgres backup/restore tests. This is not off-host or clean-host evidence.
Original R6's full six-path disposition remains CHANGES REQUIRED (RP-136/RP-544);
other correction/review obligations remain distinct. See the owning log; no age,
cadence or retention change or whole-R6 acceptance.

The release package is the integration boundary. Work is split so each batch has one authority,
its own executable failure cases and a bounded review range. No batch may claim a later criterion.

| Batch | Scope | Acceptance witness | State |
|---|---|---|---|
| DP-A | Canonical production config, file-backed secrets, origin/proxy boundary, current/previous signing-key composition and shared startup/preflight decoder | Cold Go tests cover every required field plus missing file, malformed secret, insecure origin, wrong proxy depth, half-pair and duplicate ID/value negatives | CHANGES REQUIRED — Claude designated review 2026-09-24 (log); corrective range needed |
| DP-B | Repository-independent gameserver/client/Caddy image inputs, production Compose topology, machine-readable schema, license/SBOM/release manifest generation and validation | Bundle builds from tracked inputs; clean extracted bundle reaches startup without a checkout; removed artifact and changed digest fixtures fail | CHANGES REQUIRED — Claude designated review 2026-09-24 (log); corrective range needed |
| DP-C | Encrypted off-host backup, restore, retention and manifest binding | Empty/populated Postgres restore witnesses plus corrupt/truncated/wrong-identity/wrong-manifest/partial-file negatives | CHANGES REQUIRED — Claude designated review 2026-09-24 (log); corrective range needed |
| DP-D | Stop-drain-start release helper, release record, seven-day previous-version rollback without Down migrations | Real Caddy HTTP/WebSocket release and rollback population; severed courtesy frame, drain, migration, epoch and smoke paths fail | CHANGES REQUIRED — Claude designated review 2026-09-24 (log); corrective range needed |
| DP-E | Private operations profile, metrics, journald policy and seven blocking alerts | Private reachability and retention fixtures; all seven alerts fire and severed metric/rule/receiver paths fail | CHANGES REQUIRED — Claude designated review 2026-09-24 (log); corrective range needed |
| DP-E/DP8 corrective | Preserve WebSocket upgrades through operations instrumentation, append-only repair of the pushed kernel-history miss, and one mechanically bound local push-CI aggregate | Operations-enabled real WebSocket witness; kernel correction/version guard; topology mutation proving a missing local CI lane fails | APPROVED — Claude designated review 2026-09-24 over `7b510df..cf4ac25` |
| Claude corrective R1–R22 | Owner-directed Claude corrections; the dated handoff is a historical review input, not a verdict | Individual exact-range verdicts and their executed evidence in `log.md`; historical aggregate runs do not establish green CI at current HEAD | partial bounded review only: R4/R5/R7/R11/R13/R16/R17/R19/R21 approved; R5's real private Compose witness now passes with its platform limits in the log. R11's prior native rule verdict is supplemented by exact-index amd64 rule execution, not full clean-host delivery. R14/R15 mechanisms and Codex witness supplements now have bounded independent approval (2026-11-02 log); other listed findings/corrections remain open. No full-range or DP-A–DP-E approval; retained bundles must be rebuilt |
| Codex RP-164 / DP-B CSS attribution | Build-generated CSS package provenance and final asset hashes, consumed by release metadata and delivered notices | Direct and nested CSS-only npm fixture, missing/stale graph and stylesheet negatives, Vite/Go severing probes, real build+metadata | bounded independent corrective approval of ef1a4658..f6e4d0ff and7163fe8f..627956be; no AC8 or DP-B closure |
| Codex RP-165/RP-166 / DP-B CSS URL and worker resources | Versioned CSS graph adds package assets inlined via `url()` and worker-only package stylesheets through a shared worker hook | Real Vite inlined SVG and worker-CSS shipped-byte controls; Go URL-only and worker-only notices, missing-file/v1 negatives; producer/consumer severing | bounded independent corrective approval in the same two ranges; exact verdict/limitations in log, no AC8 or DP-B closure |
| Codex native operations witnesses / DP7 | Supplemental local delivery and production-registry→Prometheus→Alertmanager journeys using pinned native services | Preflight nonce, seven firing/resolution pairs, HTTP-500 rejection and actual first cleanup failure firing/natural resolution without reset or injected resolution | original8f2e29d8..a42e680e independently APPROVED, historical target391.431s. RP-704 current startup correction independently APPROVED ae040d1d..38a3297b; reviewer actual native target422.000s and discriminating controls pass, journal-only observations explicitly uncredited. Not original R5, private Compose/Caddy/DB, Linux/full AC7/R-006 or release acceptance |
| DP-F | Exact-manifest R-006 clean-host, provider-off, supply-chain and recovery rehearsal; canonical docs and lifecycle closeout | Clean Linux/amd64 release bundle proves AC1–AC8, RPO/RTO and rollback, then receives both required review gates | planning; construction may start, external run/closeout gated |

## Batch protocol

**Historical reconciliation (2026-10-07):** use the individual verdicts and the
Deployment rows in `planning/platform-alignment/execution-queue.md`, not the
September handoff's original claim that every finding is closed. Corrective ranges
require designated independent review under AGENTS.md's current delegation, not
necessarily Claude. New-feature acceptance retains its existing gates. R11 now
has a bounded Codex designated verdict in the latest log entry. The historical kernel guard
is red at `50a3a514` (RP-131); an old green aggregate does not supersede that
current result. Pinned amd64 composed lanes have not executed on the current
arm64 host. DP-F identity/operator decisions, fresh exact bundles and clean-host
R-006 remain separate gates. No acceptance checkbox is changed here.

For coherent batches, follow RFC-0000's adopted risk-based delivery procedure:

1. identify the accepted outcome, changed boundary and meaningful regression/failure case in
   the existing plan or work update; routine repairs and acceptance construction do not need
   a separate predeclaration commit or research program;
2. update canonical `docs/` in the same commit as behavior;
3. iterate focused, then run cold (`-count=1`) affected boundary checks; retain explicit feature
   acceptance and empirical-research gates, rather than repeating them for every edit;
4. record one concise outcome, coverage/limitations and review range in the owning log; update
   other boards only when their status or next action changes, and never call a first-filter
   the designated pass;
5. hand the exact range to the designated independent reviewer under AGENTS.md's current
   authority before treating the batch as approved; and
6. keep clean-host/destructive/release lanes outside push CI unless measured D-014 headroom and a
   separately authorized CI change permit otherwise.

## Acceptance map

| RFC criterion | Owning batch | Required proof |
|---|---|---|
| AC1 reproducible clean-host package | DP-B, final DP-F | Extracted exact bundle, no checkout/writable source tree, real Caddy browser path, removed-artifact failures |
| AC2 network/config/secrets | DP-A, DP-B | One public Caddy hop; production negatives; tracked/image seeded-secret scan |
| AC3 release/drain | DP-D | Real courtesy frame and bounded drain through release helper; five severing failures |
| AC4 backup/restore/RPO/RTO | DP-C, final DP-F | Empty and populated restore identity; six named corruption/interruption failures; measured objectives |
| AC5 rollback | DP-D, final DP-F | Previous manifest plus pre-upgrade backup on clean volume; no Down migration; wrong-input refusals |
| AC6 rotation | DP-A, DP-D | Current/previous runtime verification and governed overlap/removal ledger mutations |
| AC7 operations | DP-E | Private metrics, exact retention, storage-pressure evidence and seven fired/resolved alerts |
| AC8 provider-off/supply chain | DP-B, final DP-F | No optional provider credentials; digest/SBOM/license/provenance checks with tampering failures |
| AC9 existing public/CI fact | DP-F | Retain evidence without promoting it to deployment proof |
| AC10 claim gate/closeout | DP-F | Exact-manifest R-006, current docs/registers/logs and full-range dual review before archival |

## Final closeout gates

- [ ] AC1 reproducible clean-host package passes with a demonstrated removed-artifact failure.
- [ ] AC2 network/config/secrets passes with every named negative and a seeded-secret failure.
- [ ] AC3 release/drain passes with all five severing cases.
- [ ] AC4 backup/restore passes with identity equality, corrupt inputs rejected, RPO ≤ 6 h and RTO ≤ 4 h.
- [ ] AC5 rollback passes against the previous exact manifest without a Down migration.
- [ ] AC6 rotation passes for JWT/bootstrap/cursor and refuses premature removal.
- [ ] AC7 operations passes retention, privacy and all seven fired/resolved alert fixtures.
- [ ] AC8 provider-off and supply-chain checks pass, including digest and attribution failures.
- [ ] AC9 remains accurately described as an existing external-state fact only.
- [ ] AC10 exact-manifest R-006 and transactional closeout receive Codex first-filter and Claude designated approval.

## DP-F delivery waves

| Wave | Scope | State |
|---|---|---|
| DP-F1 | Initial-install command plus exact rehearsal schema/validator and sanitized evidence contract | implemented locally at `b9d7c8a`; review remains in the DP-F range |
| DP-F2 | Reproducible exact candidate/previous bundle build, seven image SBOMs/config identities, source/image secret scan, retained build ledger, fixed-command driver and non-circular final-evidence seal | two full byte-identical local candidate builds at `102face6` with seven independent fresh image scans, correct amd64 config identities, retained record and structured scan. Historical `7e8aa70` builds remain invalid. Source44a4a72 is now independently reconstructed, but fails the current Compose/Caddy floor (RP-514); establish a reviewed floor-compliant predecessor rather than retry artifact recovery. Full supported pair/supply-chain proof and remaining producer/author gates stay open; no R-006 run or designated verdict. |
| DP-F3 | Clean Linux/amd64 host install, browser flow, backup/restart/restore/rollback, alert delivery, journal measurement and RPO/RTO observation with no checkout/providers | requires the refreshed exact candidate/retained previous pair from DP-F2, explicit clean-host target authority and required designated review boundaries |
| DP-F4 | R-006 dossier, canonical runbook/limitations, backlog/queue/RFC closeout and full-range first-filter | blocked on DP-F3 and designated approvals for DP-A–DP-E |

## RP-141 — R18 producer witness supplement

At baseline `fbbf0150`, add test-only current-repository bundle fixtures through
the production assembler and validator, and call both registered `RunProbe`
branches. Migration probe must reject a genuinely lower schema only after the
intact release pair passes compatibility; reversed/equal versions and invalid
bundles must remain setup/control failures. Missing-input probe must verify a
real age envelope and the actual image-preparation gate before removed inputs
can count. Both bundles must remain byte-bound and work directories empty.

Compiling constant-rejection dispatch faults must fail invalid-input controls;
separately sever the real migration comparison, backup gate and image gate to
make the valid producer tests fail. Restore exact source bytes before cold
focused/full Go and vet checks. No source/test/record edits while a test handle
is live. No runtime change is planned; discoveries requiring one get a separate
predeclaration. Synthetic ELF/image/client fixtures are not executable release
artifacts or clean-host R-006 evidence. Existing history/author/operator holds
and all independent review ranges remain. New Codex tests/records need Claude.

Executed locally: both registered probe rows reject, six invalid-bundle controls
remain errors, and equal/reversed version controls cannot count. Five compiling
faults fire independently (constant migration/missing-input dispatch, migration
comparison, backup gate, image gate); exact source restoration verified. Cold
focused, full Go core, declared rehearsal, relevant vet and topology lanes pass.
The initial rehearsal run lacked loopback permission; the exact target passes
with narrowly scoped local-listener access. Full kernel history stays red at
`50a3a514`. No production/kernel byte changes or full R18/DP-F approval. The
complete eight-path/three-commit `fbbf0150..07bb3d9d` supplement is independently
APPROVED by fresh Codex under Marco's corrective-review delegation; the
2026-11-02 owning log records exact hashes, cold package checks and five private
overlay failures. Original R18 remains CHANGES REQUIRED. Current-source synthetic
fixture execution is not retained-release, real Docker/SQL or clean-host proof.
