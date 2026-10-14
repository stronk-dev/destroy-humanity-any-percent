# Deployment Foundation implementation plan

- **RFC:** `rfc/deployment-foundation.md`
- **Owner:** Marco
- **Implementer:** Codex
- **Started:** 2026-08-22
- **State:** implementing
- **Implementation baseline:** `cd102d7` (accepted after designated cross-party review)

**Current DP7 timed-observation correction, 2026-10-14 (RP-615):** the real-minute test's
startup/failure/drain callback waits now obey its existing95-second context and refuse missing
or late observations. Three factored-old receive regressions fail; corrected cold/vet/race
controls and the actual Linux-amd64/Postgres timer population pass, retaining real60-second
callbacks, SQL failure/recovery and orderly drain. This changes tests only, not credential
retention, timer, runtime or alert policy. The log owns the range after `21734b10`, pending
Claude review; current-head manual evidence is not full AC7 or exact-artifact/clean-host proof.

**Current DP5 smoke correction and release-lane evidence, 2026-10-14:** RP-612 rejects
size-limiter EOF as completion, preserving the existing 4 MiB ceiling. Four old decoder/caller
regressions fail; corrected native/race and the actual declared Caddy/Postgres/WebSocket lane
pass, including admitted work, courtesy/drain, restart and restored-account smoke. The log
records the range after `7644aed4` pending Claude, and the lane's checkout-backed/emulated limits.
Current Docker28.4.0/Compose2.39.4-desktop.1 matches the recorded development pair but its
daemon is Linux/aarch64; CR2/CR4 native supported-host experiments remain unexecuted.

**Current DP6 cleanup correction and restored test environment, 2026-10-14:**
RP-611 binds cleanup to inspected directory/file identities and a directory-rooted handle.
Three old physical-file regressions delete an uninspected replacement; corrected native and
Linux-amd64 populations preserve it, including unchanged aliases and second-process writer
exclusion. Cold affected packages/vet/race and Linux helper build pass. Docker now responds;
`make test-deployment-backup` executes all seven real Postgres integration tests, including
RP-551's empty/populated query-override restores. Exact evidence/range after `68eee56b` is in
the owning log, pending Claude. Other previously unavailable lanes must run individually;
no daemon intervention, current-release proposal adoption or clean-host/R-006 pass is inferred.

**Current backup-input correction, 2026-10-13 (RP-559):** malformed URL queries now
refuse before service/pass file creation, rather than silently dropping invalid fields
and preparing fallback credentials. Twelve actual old-helper failures and four encoded-value
controls precede the correction; all credential-file checks/race, four cold affected packages
and vet pass. The log records the initially falsified cross-parser hypothesis and exact range
after `0e7a4b95`, pending Claude review. This does not extend supported URI options or establish
SQL restore, Docker/Linux/hosted or exact-artifact acceptance; those boundaries remain open.

**Current contract work, 2026-10-13 (RP-558 / DESIGN-GAP10):** the
[Current Release Authority draft](../../rfc/deployment-current-release-authority.md) specifies
live image/mounted-byte identity, one ledger with durable release/rollback events and explicit
interruption recovery. It is not accepted/acceptance-ready; CR2/CR4 require the two named pinned-
engine experiments before mechanical closure and adoption/review. Native actual ExecRunner/lock
evidence confirms child continuation after killed-parent lock release, with normal completion
control and both children observed terminal; preserved manual source is outside CI. The initial Docker
socket ping timed out; it now responds, but the daemon is Linux/aarch64 and no supported-host
mounted-reader/quiescence experiment is claimed. No product byte
changed. Next: execute/reconcile the mechanical questions when the declared engine is usable;
continue other accepted work meanwhile. D-019, other DESIGN-GAPs and full nine-tier floor remain.

**Current operator-state admission, 2026-10-13 (RP-556/RP-557):** lifecycle commands now
reject existing inadmissible history/start times before runtime operations; lock/append paths
reject final directory symlinks before changing their targets. Old regressions and all30
corrected physical-file/recording-runtime cases, focused race, four cold packages/vet and Linux
helper build pass. New range after `b448069b` needs Claude review; not real Docker/SQL/R-006.
**Next engineering contract: RP-558 / DESIGN-GAP10**—define the actual current-release identity
binding, including missing history and failed transitions, before claiming forward-only live
upgrades. The current controller admits a contradictory install-history diagnostic; no live-host
downgrade is proved. Preserve existing accepted behavior until its contract closure is adopted;
this is agent specification work, not a request that Marco design a technical parameter.

**Preceding rotation chronology correction, 2026-10-13 (RP-555):** activation/removal refuse
before appending a timestamp the existing reader would reject. Six old-source poisoned-ledger
cases and12 equal/later controls establish the regression; all18 corrected cases preserve
readability and valid retry, across all three key families. Cold four-package tests/vet,
focused race and Linux-amd64 helper build pass. New range after `3964194b` needs Claude review;
overlap minima, cursor contract holds and exact-host rotation proof remain unchanged. Original
R1's full six-path CHANGES REQUIRED disposition is now recorded; its Codex descendants still
need independent review. Continue accepted construction without relabelling native proof R-006.

**Preceding failed-install correction, 2026-10-13 (RP-554):** `AbortInstall` now retains its
normalized runtime rather than dropping the default command runner and rotation overlay.
Three old-source failures and five controls precede the correction; all eight now pass,
including actual ExecRunner subprocess failures through an owned recording executable.
Cold four-package tests/vet, focused race and Linux-amd64 helper build pass. Original R2's
complete nine-path disposition is CHANGES REQUIRED (RP-134/RP-554), not approval of Codex
corrections. New range after `b868cd2c` needs Claude review; Docker/SQL failed-install and
exact-artifact clean-host proof remain open. No change to rotation timing or cleanup targets.

**Current DP-F refusal-record correction, 2026-10-13 (RP-552):** the non-clean restore classifier
now rejects duplicate and case-aliased members rather than allowing Go's overwrite/case-fold
semantics to create valid-looking evidence. Sixteen old-parser regressions and two actual-probe
orchestration counterexamples fail; corrected cold affected packages/vet, race and Linux helper
build pass. Full original R20's eleven-path disposition is CHANGES REQUIRED (RP-142), not approval
of its corrective descendants. New range after `fc04db24` needs Claude review. Actual Postgres,
composed CLI stderr and exact-host/R-006 gates remain; unchanged valid records stay admissible.

**Current DP6 endpoint correction, 2026-10-13 (RP-551):** service/pass files now honor the
database URL's endpoint query overrides instead of targeting its unoverridden URI fields.
Old-source20-case population fails18, retaining ordinary/password-isolation controls; corrected
focused/race, affected cold packages/vet and Linux helper build pass. The existing actual-Postgres
empty/populated restore test now includes swapped URI paths with database-query overrides; it
initially only compiled/skipped locally. The 2026-10-14 declared Linux/Postgres run now passes
both empty/populated override populations; the latest owning log records it. New range after
`a7501ac1` still needs Claude review. This
does not close DP-C/AC4, general URI options, original R20 review or exact-host/R-006.

**Current DP-F producer evidence, 2026-10-13 (RP-550):** all nine packaged-artifact mutations now
execute against the actual assembled-bundle validator, with18 invalid-baseline controls. The
content-identity omission fails the new epoch witness while the prior fake-validator test passes.
Cold affected packages/vet and focused race pass. This test-only range after `5d70f673` needs
Claude review; original R18's complete five-path disposition remains CHANGES REQUIRED (RP-141),
not approval of Codex's later witnesses. Synthetic image/client/dump inputs and recorded runtime
commands are not exact-artifact/clean-host R-006, or permission to execute external deployment.

**Current DP-F chronology correction, 2026-10-13 (RP-549):** sequential command results must not
overlap or run backwards, even inside one lifecycle step. The executor now refuses before the
next producer launches; base/final validators reject rehashed overlapping results. Ten focused
cases preserve exact adjacent boundaries and forward gaps; old-source failures and corrected
cold/race checks are recorded in the owning log. New range after `88868d34` needs Claude review.
Original R12's full five-path disposition remains CHANGES REQUIRED (RP-140); basename/command
shape is not executable identity, and no synthetic result or seal closes R-006.

**Current DP5/DP6 creation correction, 2026-10-13 (RP-548):** newly created backup reports must
place start/completion inside their command invocation; invalid clocks refuse. Old-source20-case
population admits14 bad reports while six legal controls pass. Corrected actual encrypted
envelopes, cold affected packages/vet, race and Linux-amd64 helper build pass; existing host/header
oracle still rejects its isolated bypass. The range after `dac25388` needs Claude review. Original
R10's full four-path disposition remains CHANGES REQUIRED; D-019, authenticated metadata/snapshot
meaning, logical restore, clean-host/RPO/RTO and exact-artifact gates are not resolved by this check.

**Current DP3/DP5 correction, 2026-10-13 (RP-546):** invalid receiver/readiness request syntax
now refuses rather than panics; configured receiver construction is checked before runtime
commands. Old-source eight failures, ordinary receiver/ready/drain/identity controls, affected
cold packages/vet/race and Linux-amd64 helper build pass. Original R9's complete five-path
disposition remains CHANGES REQUIRED; this correction after `39019517` needs Claude review.
RP-547 now adds actual `DockerRuntime.DrainCurrent` coverage through native HTTP/WebSocket and
recording commands: eleven success/failure cases, two caller-level falsification overlays,
affected cold packages/vet and race pass. Its test/docs/record range after `de0ca191` needs Claude
review. No production behavior or drain bound changed; synthetic identity/exit replies are not
real process, Caddy/Postgres/Linux or clean-host/R-006 proof. Continue remaining accepted original
reviews and exact-artifact construction while those environment/review gates stay explicit.

**Current DP1/AC8 correction, 2026-10-13 (RP-545):** every shipped npm copy's license/notice is
validated before name/version deduplication, including CSS-derived copies. Matching attribution
merges; conflicting/missing evidence refuses without partial inventory. Two-route old-source
negatives, delivered-notice controls, cold affected packages/vet and the freshly built client's
actual53-dependency metadata pass locally. The original Go helper-dependency guard is restored.
Original R8's full seven-path disposition remains CHANGES REQUIRED; this new correction after
`fa2bec58` and prior map/CSS repairs need designated Claude review. No classifier/policy/schema,
image provenance, clean-host, Docker/Linux-runtime, full AC8 or release promotion.

**Current DP3/AC2 correction, 2026-10-13 (RP-543):** archive secret scanning now
includes decoded tar owner/group/PAX/xattr fields and gzip header metadata; recognized
secret-bearing finding/error paths are redacted. Digest-valid old-source negatives and
clean controls discriminate; cold packages/vet and an ordinary scan of the retained
102face6 image pass. Its bytes remain historical and its embedded tool lacks this fix.
The log records the complete original R3 five-path CHANGES REQUIRED disposition and the
new correction range after `cf8c6f08`, pending designated review. Rebuild/re-scan the final
reviewed artifact; no new scanner rule/limit, R-006, full Deployment or1.0 acceptance.

**Latest DP7 actual proxy outage, 2026-10-13:** the shipped Caddyfile's private
route/scrape now has a physical stop→real pending five-minute rule→receiver
firing→same-container restart→matching resolution witness. Upstream readiness is
an explicit control, not a gameserver proof. Root lane passes406.67s, cold affected
native/vet and Linux-amd64/Postgres checks pass. The log owns the range after
`12399dc9`, pending designated review. Continue actual game/host observation,
journal/retention, older restart history and exact-artifact evidence; no full AC7.

**Latest DP6/DP7 actual backup execution, 2026-10-13:** the production CLI now
has a joined actual Postgres dump→scheduled target refusal→receiver firing→real
five-minute retry→matching resolution→fresh-database restore witness. The retry
restores a deliberately updated account row, not the old dump. No direct backup
metric or fake dump is used in this population. Shared fixture metadata remains
synthetic, and the log owns final checks/range after `c4279b66`, pending designated
review. Continue actual game/host observation, retention, older-container restart
history and exact-artifact evidence; this is not off-host/RPO/RTO or full AC7/R-006.

**Latest DP7 restart receiver construction, 2026-10-13:** the actual Docker
observer's first-scrape count3/increase0 now fires through the private services
and reaches the receiver; an actual replacement count0 resolves the same alert.
The final root lane also proves canceled-CLI one-off cleanup. Its early-notification
loss finding (RP-523), failed attempt, final checks and range after `cdc78050`
live in the log, pending designated review. No production bytes or alert policies
change in this supplement. Continue remaining actual backup/game-process/host
journal/retention, older-container restart history and exact release-artifact
evidence; full AC7/R-006 remain open.

**Latest DP7 restart correction, 2026-10-13 (RP-522):** actual Docker count and
container birth now reach the host textfile together. A new-container lifetime
arm catches three restarts before the first scrape, preserving the original
counter-increase rule and ten-minute threshold. Old rule regression fails;
corrected pinned rules and actual automatic-restart/observer witness pass.
The log owns checks and the range after `db752b0a`, pending designated review.
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
pass. The log owns the range after `82c8e57b`, pending designated review. Continue
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
the range after `5fcb5353`, pending designated review. Continue remaining actual
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
regressions and the new range after `a38bcb9c`, pending designated review; consolidate
with that diagnosis and earlier DP7 work. Continue those engineering proofs.

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
The owning log records the corrective range needing designated review and remaining gates.

**Latest bounded correction (RP-511):** malformed gameserver archive headers now return
invalid content rather than panicking in the validator/assembler. Eight old-source failures
become refusals; intact controls and all twelve malformed-input cases pass. Affected cold
native/vet and Linux packages plus existing real Postgres/Caddy integration pass. The log
owns the pending designated range and fixture limits. Six Linux binaries, client, metadata
and content build at `abf4b4ac`, but these pre-correction outputs are not a current full bundle.

**Latest bounded correction (RP-510):** both package-creation passes now bind the
same complete dump under DP6/AC4. Four old-source regressions fail; rewind before
hashing plus streamed hash/EOF validation prevents publishing an inconsistent
package. Actual positioned empty/populated Postgres dumps restore with matching
recovery identity; affected native/vet and Linux checks pass. No format, retention,
clock, key or CI contract change. The owning log records results and the full
corrective range needing designated review; broader Deployment gates remain open.

**Latest bounded correction, 2026-10-12 (RP-508):** rollback input preflight now uses the
actual authenticated package decoder before stopping services/resetting Postgres. The old
method admits encrypted non-packages and packages with trailing data; focused controls now
also reject a changed dump with its old internal hash. Plaintext workspaces are private and
removed on return. The R18 missing-input producer now creates a full package, retaining its
healthy-first control and synthetic-dump limitation. Real Postgres/Caddy and affected native/
Linux checks pass; the owning log holds the pending review range. D-019's separate per-release
restore choice, logical restore/RTO, clean-host and broader Deployment gates remain open.

**Latest bounded correction, 2026-10-12 (RP-507):** package extraction now consumes the
authenticated stream after tar EOF and refuses any remaining bytes before returning a usable
dump. The old implementation stalls for one extra byte, a zero block and a second archive;
the exact-package control passes. The existing empty/populated Postgres restore population
also proves refusal leaves its target clean before the intact backup restores with matching
identities. The owning log records affected checks and the pending designated range. No
metadata-binding draft adoption, retention/format change or clean-host acceptance follows.

**Latest bounded correction, 2026-10-12 (RP-504):** populated recovery records its
start after input/lock/checkpoint validation but before stopping the failed stack and
resetting/restarting Postgres. Those required phases now count toward the existing4h
RTO. Four old-source failures become passes: complete elapsed time, stop/reset overruns,
and invalid-clock refusal before destructive work. Affected cold packages/vet and declared
Linux-amd64 recovery tests pass. Controlled runtime evidence, not an actual clean-host
restore; the owning log holds the new pending review range, combinable with RP-503.

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
integration functions pass. The owning log holds the exact range requiring Claude review.
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
corrected cancellation observer and review range after `8ba84996`; it can share
review with RP-489. Separate dump snapshots and full Deployment/clean-host gates remain.

**Preceding bounded checkpoint, 2026-10-12:** RP-489 unifies backup-file header admission with
the existing single-object decoder and bounds outer/decrypted metadata framing.
Eight encrypted malformed arms fail on old code; final14 arms retain healthy,
whitespace and exact64KiB controls. Affected cold Go/vet and all six declared
Linux-amd64/Postgres backup integration functions pass. The owning log holds the
range after `630f6833`; designated review, full Deployment and clean-host gates remain.

**Bounded checkpoint, 2026-10-09:** RP-482 adds healthy baselines to the public-metrics and
source/image-secret probes. Actual assembled-bundle tampering and always-rejecting gates
fail the old checks; affected native/Linux-amd64 packages pass after correction. This range
after `b76d1a08` can share designated review with RP-481; no DP-F acceptance follows.

**Preceding bounded checkpoint, 2026-10-09:** RP-481's interrupted-writer probe now requires a real
encrypted create/restore baseline and the exact consumed dump failure, with no partial files
or damage to the baseline. Old-source regression and affected native/Linux-amd64 checks
pass as recorded in the owning log. The corrective range after `1efea9cf` awaits designated
review; this does not close DP-C/DP-F or substitute for the actual clean-host rehearsal.

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
pass locally, with cold affected packages/vet/race and a Linux-amd64 build. This is not Linux
runtime, Postgres or clean-host evidence. Original R6's full six-path disposition remains CHANGES
REQUIRED (RP-136/RP-544); the new repair after `ca447daa` and older checksum corrections need
designated Claude review before R6 closes. See the owning log; no age, cadence or retention change.

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
| Claude corrective R1–R22 | Owner-directed Claude corrections; the dated handoff is a historical review input, not a verdict | Individual exact-range verdicts and their executed evidence in `log.md`; historical aggregate runs do not establish green CI at current HEAD | partial bounded review only: R4/R5/R7/R11/R13/R16/R17/R19/R21 approved; R5's real private Compose witness now passes with its platform limits in the log. R11's prior native rule verdict is supplemented by exact-index amd64 rule execution, not full clean-host delivery. R14/R15 mechanisms have bounded approval but Codex witness supplements need Claude; other listed findings/corrections remain open. No full-range or DP-A–DP-E approval; retained bundles must be rebuilt |
| Codex RP-164 / DP-B CSS attribution | Build-generated CSS package provenance and final asset hashes, consumed by release metadata and delivered notices | Direct and nested CSS-only npm fixture, missing/stale graph and stylesheet negatives, Vite/Go severing probes, real build+metadata | locally implemented; Claude designated review required; no AC8 or DP-B closure |
| Codex RP-165/RP-166 / DP-B CSS URL and worker resources | Versioned CSS graph adds package assets inlined via `url()` and worker-only package stylesheets through a shared worker hook | Real Vite inlined SVG and worker-CSS shipped-byte controls; Go URL-only and worker-only notices, missing-file/v1 negatives; producer/consumer severing | locally implemented; Claude designated review required; no AC8 or DP-B closure |
| Codex native operations witnesses / DP7 | Supplemental local delivery and production-registry→Prometheus→Alertmanager journeys using pinned native services | Preflight nonce, seven firing/resolution pairs, HTTP-500 rejection and actual first cleanup failure firing/natural resolution without reset or injected resolution | combined target locally passed; consolidated Claude review required for the new test range after `8f2e29d8`. Not original R5, private Compose/Caddy/DB or release acceptance |
| DP-F | Exact-manifest R-006 clean-host, provider-off, supply-chain and recovery rehearsal; canonical docs and lifecycle closeout | Clean Linux/amd64 release bundle proves AC1–AC8, RPO/RTO and rollback, then receives both required review gates | planning; construction may start, external run/closeout gated |

## Batch protocol

**Historical reconciliation (2026-10-07):** use the individual verdicts and the
Deployment rows in `planning/platform-alignment/execution-queue.md`, not the
September handoff's original claim that every finding is closed. Codex repairs
and record edges still require Claude's exact-range designated review. R11 now
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
5. hand the exact range to Claude for the cross-party designated review before treating the batch
   as approved; and
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
entire new span after `fbbf0150`, including records, needs Claude independently.
