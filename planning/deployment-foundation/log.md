# Deployment Foundation implementation log

Append-only record. A fresh agent should read `plan.md`, this log and the active RFC before acting.

## 2026-08-22 — implementation opened

- Owner accepted the reconciled RFC after Claude's designated cross-party `APPROVED` verdict in
  `cd102d7`. That commit is the implementation baseline.
- Corrected the active index's stale draft label and advanced the RFC to `implementing` because
  this planning directory now exists and work is underway.
- Decomposed the work into six dependency-ordered batches. The clean-host/R-006 claim remains last;
  no component-level success may promote the package to supported self-hosting.
- No product behavior changed in this opening record.

## 2026-08-22 — DP-A predeclaration

**Authority:** RFC DP2–DP4 and AC2/AC6. This batch does not create Compose, release, backup,
observability or public-release claims.

**Expected paths:** the gameserver entry point; a focused deployment configuration package and its
tests; account/bootstrap/public-cursor key adapters only where current/previous runtime support is
missing; `Makefile` only for the shared preflight entry point; and canonical configuration docs.
Exact path expansion must be recorded before commit if inspection discovers a required owner.

**Positive population:** decode the production environment through one shared startup/preflight
path; read secrets from files; validate one canonical HTTPS origin, one proxy hop, UUID server ID,
packaged content root and current plus optional previous keys; prove both current and previous
verification where the runtime consumer already exists.

**Negative population:** unknown deployment key, missing/empty/unreadable secret file, malformed
base64 and wrong key length, HTTP/path/query/fragment origin, production proxy depth other than
one, legacy inline secret in production, half-specified previous pair, duplicate current/previous
ID and duplicate current/previous value. Each family must have a test that fails when its rejecting
branch is severed.

**Authorized claim:** the gameserver and its preflight share a fail-closed production config and
file-secret decoder for implemented key consumers. **Not authorized:** a deployable bundle,
complete rotation ledger, public cursor reader, backup, rollback, observability, R-006 or release
readiness.

## 2026-08-22 — DP-A implementation and Codex first-filter

**Implementation commit:** `a906398` (review baseline `bbff0b6`).

**Actual paths:** `server/deploymentconfig/{config.go,config_test.go}` owns the shared decoder;
`server/cmd/gameserver/{main.go,main_test.go}` consumes it and preserves both rotation pairs;
`server/gameserver/{composition.go,deployment_test.go}` binds the one external origin and trusted
hop to the real transport/account configuration; `Makefile`, `docs/gameserver.md` and
`docs/accounts-and-sessions.md` expose only the behavior implemented in this batch. No account,
bootstrap or public-cursor consumer needed modification.

**Executed evidence (cold):**

- `make test-go GO_PACKAGES='./deploymentconfig ./gameserver ./cmd/gameserver ./account ./transport ./publicapi' GO_TEST_FLAGS='-count=1'` — PASS;
- `make vet GO_PACKAGES='./deploymentconfig ./gameserver ./cmd/gameserver'` — PASS;
- `make test-save-integration SAVE_TEST_PACKAGES='./gameserver ./account' SAVE_TEST_FLAGS='-run Integration' SAVE_TEST_COUNT=1` — PASS against the declared Docker Postgres service (`gameserver` 20.424 s, `account` 1.305 s); the first sandboxed Docker attempt was denied access to the daemon, then the same command was rerun with the approved Docker authority; and
- `make build-gameserver` — PASS.

**Discrimination probes (temporary mutations, all restored; clean diff confirmed):**

1. bypassing unknown `CLOUD_CLICKER_*` rejection made
   `TestProductionConfigRejectsEveryFailClosedFamily/unknown_deployment_key` fail because the
   invalid fixture was accepted;
2. weakening the exact one-hop check made
   `TestDeploymentBoundaryBindsOneOriginToOneTrustedProxyHop` fail on a zero-hop production
   origin; and
3. severing the previous-bootstrap-key adapter made
   `TestCompositionKeysPreserveCurrentAndPreviousRuntimeMaterial` fail with an empty previous map.

**Review by:** Codex. **Recorded by:** Codex. First-filter range `bbff0b6..a906398` reviewed in full:
scope matches DP2–DP4/AC2/AC6, behavior and docs agree, secret-bearing errors were checked, every
production field/negative family has an executable row, and the runtime adapters have direct
tests. Verdict: **APPROVED as first filter; not the designated pass.** Cursor keys are deliberately
validated but not consumed because no public reader is composed; no rotation-ledger or release
claim is made. DP-A is ready for Claude's exact-range designated cross-party review.

## 2026-08-22 — DP-B predeclaration and runtime-closure audit

DP-A awaits its mandatory designated review over `bbff0b6..d7d443f`. DP-B begins as a separate
range; this does not promote DP-A's first-filter verdict.

**Authority:** RFC DP1–DP3, DP8 and AC1/AC2/AC8. DP-B does not implement backup, release/rollback,
rotation-ledger timing, alerting or the final R-006 claim.

**Runtime closure derived from code, not a hand-maintained catalog list:**

- `epochseed.Path` (`balance/epochs/phase0.json`), every artifact path and every epoch changelog
  reference declared by those bytes;
- `balance/transport/phase0.json` and `moderation/guild-names.txt`, the two additional files opened
  by `gameserver.Compose`;
- the existing `deployment/content-manifest.v1.json` identity record;
- the statically linked Linux/amd64 gameserver and built `client/dist` tree; and
- root MIT and generated linked-server/client-runtime third-party notices.

The closure builder must consume the epoch declaration rather than repeat its 19 current artifact
rows. Test-only balance trees, source, planning and a writable checkout are not release inputs.

**Expected tracked paths:** `deployment/` production Dockerfile, Caddyfile, base/rotation Compose,
environment example and generated config/release schemas/notices; one focused release-package
builder/validator plus fixtures/tests under `tools/` or `server/cmd/`; `Makefile`; the smallest
D-014-compatible static CI hook only after its local elapsed time is measured; and canonical
deployment documentation. Any expansion is recorded before its implementing commit.

**Positive population:** build client and static Linux/amd64 server; stage the derived runtime
closure into a fresh output directory; validate every staged byte/hash, exact image digest,
schema/save/epoch/copy identities, private-service topology, Caddy routes, license set and SPDX
SBOM; then build/start from only that output. The base Compose mounts current secrets only; a
reviewed rotation override supplies optional previous pairs, avoiding fake placeholder secrets.

**Negative/severing population:** remove each required closure class in turn; add an undeclared
source dependency; change one staged byte or image digest; use a mutable image tag; publish a
non-Caddy port; expose metrics through Caddy; seed a secret-like value into tracked/built image
inputs; remove root or third-party attribution; forge schema/epoch/copy/version identity; and run
the bundle command without a client build or Linux/amd64 server. Each must fail nonzero.

**Authorized claim:** exact tracked inputs can produce a byte-bound, repository-independent
release candidate with a private one-node topology. **Not authorized:** supported self-hosting,
backup/restore, rollback, operational alerts, RPO/RTO or release readiness until their later
batches and exact-manifest R-006 pass.

## 2026-08-22 — DP-B implementation and Codex first-filter

**Implementation commits:** `5f1751e`, `f5ba299`, `7277b83`, `018a977`, `cda5a10`, `dde5c15`
and `ba643db` (review baseline `f62e0b7`; this record commit is the range tip).

**Actual boundary:** `server/releasepackage` and its four small commands derive/stage the runtime
content closure, validate/render the three-service private Compose topology, inventory the linked
application license set, emit SPDX 2.3, construct/validate the release manifest, inspect an offline
gameserver Docker archive and scan tracked/bundled/image bytes for recognized secret material.
`deployment/` supplies the scratch/nonroot gameserver Dockerfile, Caddy/Compose/rotation inputs,
non-secret example and closed JSON Schemas. `Makefile` exposes explicit build inputs; canonical
behavior and limitations live in `docs/deployment.md`.

The exact candidate source is `ba643db317513e2ed8e0ff666b49cd3526983b30`. Its package contains
53 manifest-bound artifacts. The final clean manifest SHA-256 is
`67ae884e5dc0011fbef1d31e0fef585bcfdd1bb25a7395563db907b482e814e6` and records:

- gameserver config/image ID
  `sha256:afe2f78978f50c226b6dfcd08c891c5024c10dcf2f55ab03a6d0e79cbf0ce7f9`;
- Caddy index `sha256:5f5c8640aae01df9654968d946d8f1a56c497f1dd5c5cda4cf95ab7c14d58648`
  with linux/amd64 config
  `sha256:af555904a0961945f16bb323a501457b13a4f7e9bde969b145b97da80b38ecbe`; and
- Postgres index `sha256:cf78e76683b9ca8c5733cbbdce6c9262b45b6767934dd0a95e671f9a0fc20685`
  with linux/amd64 config
  `sha256:75f5a96988cdf694a215073c3e9c001b706b371e2f94df3967f2efdec2787f6b`.

**Executed evidence:**

- two no-cache exports from the digest-pinned BuildKit v0.24.0 builder, with fixed source epoch and
  timestamp rewriting, were byte-identical: archive SHA-256
  `5b2b7f6e597ff9d4900b13591cec394c75034f2ecb3832c2788b2637243a37b3`;
- the clean assembler accepted the exact linux/amd64 binary/image/config-addressed SBOM population
  and emitted all 53 artifacts; `docker compose config --quiet` accepted the rendered Compose;
- from only the extracted bundle plus operator fixture secrets, real Postgres, the amd64 gameserver
  under emulation and Caddy returned `health=204`, `ready=204`, `spa=200`, `bootstrap=201`;
- Docker Desktop lacks journald, so the startup witness used an external test-only logging override
  to `json-file`. The unmodified release Compose first failed on that unsupported-host condition;
  the override did not alter tracked or manifest-bound bytes. This is DP-B component evidence, not
  the required clean Linux host R-006 rehearsal;
- `make test-go-ci CI_TEST_PACKAGES='./releasepackage ./cmd/assemble-release-bundle
  ./cmd/release-secret-scan'` — PASS cold inside the repository's CI service after the final amd64
  provenance correction;
- `make release-secret-scan GAMESERVER_IMAGE_ARCHIVE=<exact archive>` — PASS over 1,305 tracked
  paths and every saved-image member; and
- `make verify-ci-topology` — PASS, including all 10 negative topology fixtures. No workflow was
  changed: these Go tests already execute in the existing server job, while image/clean-host work
  remains manual as DP8 requires.

**Discrimination and audit findings:**

1. temporarily bypassing the bundle artifact comparison made the changed-site-byte fixture pass
   incorrectly; `TestReleaseManifestBindsEveryBundleByteAndImageSBOM` failed with `tampered bundle
   accepted`, and the mutation was restored;
2. seeded tracked material and malformed/seeded archive bytes are rejected by the scanner; mutable
   image references, wrong amd64 binary, client symlink, missing attribution, changed image/SBOM
   hash, forged schema field set and source-commit/image-label mismatch each have cold negatives;
3. an initial image was correctly exposed as arm64 metadata around an amd64 binary. The final build
   fixes the platform explicitly and archive validation binds both architecture and entry point;
4. ordinary Docker-driver builds remained timestamp-dependent even with `SOURCE_DATE_EPOCH`.
   DP-B therefore moved to the pinned container BuildKit exporter with `rewrite-timestamp=true` and
   proved two no-cache archives byte-equal; and
5. the first upstream SBOM attempt described native arm64 variants. The final manifest records each
   linux/amd64 runtime config digest and requires each SPDX name to identify that config, so a
   native-host or swapped SBOM fails before packaging.

**Review by:** Codex. **Recorded by:** Codex. First-filter range `f62e0b7..ba643db` plus this record
commit reviewed in full. Scope remains DP1–DP3/DP8 and AC1/AC2/AC8 only; no backup, release,
rollback, rotation ledger, operations or R-006 claim is present. Verdict: **APPROVED as first
filter; not the designated pass.** DP-B is ready for Claude's exact-range designated cross-party
review and remains unarchived.

## 2026-08-22 — DP-C predeclaration

DP-A and DP-B await their independent designated reviews. DP-C begins in a new range after
`9b916c2`; neither pending verdict is promoted by this parallel construction.

**Authority:** RFC DP6 and AC4 only. This batch creates backup/restore/retention primitives and
their real Postgres witnesses. It does not perform release replacement, rollback, alert delivery,
rotation timing, operations composition or the final R-006 rehearsal.

**Expected paths:** one focused backup package and command under `server/`; package-owned test
fixtures; release-bundle backup/restore helper entry points; the smallest necessary Compose
backup-worker inputs; `Makefile`; and canonical deployment/operations documentation. Database
migrations and gameplay schemas are not expected to change.

**Positive population:** empty and populated Postgres custom-format dumps; age X25519 encryption
to an operator recipient; plaintext envelope plus encrypted authenticated payload; atomic target
rename; exact release-manifest/epoch/artifact/server/timestamp metadata; restore into a clean
database; identity equality for account, Founder, Company, events, leaderboard and epoch; governed
six-hourly/daily retention preserving newest and unresolved pre-upgrade records; measured RPO/RTO
fields that fail rather than claim success when incomplete.

**Negative population:** truncated/corrupt payload, wrong age identity, wrong release manifest,
partial temporary output, simulated interruption between dump/encrypt/rename, missing/late backup,
non-clean restore target and retention attempts against newest/unresolved pre-upgrade backups. No
fixture may coast past an excluded row or guard.

**Authorized claim:** the exact release package can create, validate, retain and restore encrypted
database backups with identity evidence. **Not authorized:** RPO/RTO acceptance, supported
self-hosting, release/rollback, operational readiness or release readiness until DP-D–DP-F and
exact-manifest R-006 pass.

## 2026-08-22 — DP-C test-boundary expansion before CI-environment edit

The first Linux/amd64 emulation run exposed a QEMU CPU-feature defect in the age payload path:
native arm64 passed, emulated amd64 corrupted its own authenticated chunk, and the same amd64 image
passed when either AVX2 or BMI2 advertisement was disabled. This is not accepted as a product
result. DP-C therefore expands to `compose.ci-test.yml`, `compose.deployment-backup-test.yml` and
`docs/ci.md` before editing them: local Apple-host amd64 emulation disables AVX2, while hosted
GitHub Actions remains native amd64 and exercises its ordinary detected crypto path. The dedicated
Postgres backup population is also pinned explicitly to `linux/amd64`. No workflow, timeout,
acceptance bound or hosted job changes.

## 2026-08-22 — DP-C implementation and Codex first-filter

**Implementation commits:** `23a590d` and `120343c` after predeclaration `86cf4d6` (review
baseline `9b916c2`; this record commit is the range tip).

**Actual boundary:** `server/deploymentbackup` owns the age/X25519 envelope, authenticated package,
Postgres 16 command composition, clean-target and exact-migration checks, schedule/objective state
and governed retention. `cmd/deployment-backup` supplies create/schedule/restore/retention operator
entry points. The release assembler now requires and hashes the static Linux/amd64 helper; the
fourth base-Compose service reuses the pinned Postgres image for exact client tools and remains
nonroot, read-only, private and separately mounted. `compose.deployment-backup-test.yml` owns the
real empty/populated restore population. No gameplay schema, migration, release replacement,
rollback, alert or workflow changed.

The encrypted payload contains the Postgres custom-format archive, strict release manifest, exact
epoch declaration and authenticated metadata. Creation checks the live Goose migration before
dumping, never stages plaintext on the off-host target, syncs encrypted bytes and commits by atomic
rename. Restore authenticates and checks every identity before mutation, verifies the archive,
refuses any non-clean target and checks the restored migration. Retention is derived internally
from the complete target population so a caller cannot forge a deletion plan; invalid entries
block deletion, while newest and unresolved pre-upgrade backups remain protected. Missing/late and
incomplete RPO/RTO observations fail instead of coasting.

**Executed evidence (cold):**

- `make verify-server-core` — PASS across vet, every non-harness Go package, formula drift and API
  generation;
- `make test-go-ci CI_TEST_PACKAGES='./deploymentbackup ./cmd/deployment-backup ./releasepackage
  ./cmd/assemble-release-bundle'` — PASS on emulated Linux/amd64 with cold package tests;
- `make test-deployment-backup` — PASS using a cross-compiled Linux/amd64 test binary inside the
  Postgres 16 Alpine image: empty and populated custom dumps restored into newly created databases,
  and exact Account, Founder, Company/save, event, verified-board and epoch identities matched;
- the same lane passed its live-migration/release-manifest mismatch and occupied-target negatives;
- the rendered four-service release Compose passed Docker Compose's own `config --quiet` parser;
- `make build-deployment-backup-linux-amd64` — PASS; and
- `make release-secret-scan` — PASS over 1,307 tracked paths.

The first emulated amd64 run failed its own age round trip. Disabling either advertised AVX2 or
BMI2 made the same image pass, identifying Docker Desktop/QEMU CPU emulation rather than accepted
backup evidence. Local amd64 Compose now disables AVX2 explicitly; native GitHub Actions retains
ordinary CPU detection, and exact native-host R-006 remains mandatory. Two subsequent Dockerfile
module-download attempts also failed on QEMU TLS before test execution; the final lane avoids that
unrelated network path by cross-compiling through the repository toolchain and mounting the exact
test binary into Postgres.

**Discrimination probes (temporary mutations, all restored; clean diff against the committed
implementation confirmed):**

1. bypassing encrypted/plaintext header comparison made a rewritten `server_id` restore
   successfully; `TestPlaintextEnvelopeIdentityCannotBeRewritten` failed with `rewritten plaintext
   identity accepted`;
2. severing newest-record protection made the sole 31-day-old recovery point deletable;
   `TestRetentionNeverDeletesTheOnlyNewestBackupEvenAfterThirtyDays` failed on the exact path;
3. bypassing clean-target inspection allowed a restore beside an unrelated `occupied` table;
   `TestPostgresRestoreRefusesNonCleanTargetIntegration` failed with `non-clean target accepted`;
4. bypassing Goose/release-manifest equality made the wrong-migration fixture produce a backup;
   `TestPostgresBackupRestoreEmptyAndPopulatedIdentityIntegration` failed with
   `live database/release migration mismatch accepted`; and
5. the permanent Compose negatives independently sever root-user and separate-target rules and are
   rejected by `ValidateCompose`.

**Review by:** Codex. **Recorded by:** Codex. First-filter range `9b916c2..120343c` plus this record
commit reviewed in full. Scope matches DP6/AC4; no later-batch claim is made. The earlier real DP-B
candidate is intentionally not reused because adding the manifest-bound helper changes its bytes;
an exact rebuilt candidate and measured Caddy-complete RPO/RTO remain DP-F evidence. Verdict:
**APPROVED as first filter; not the designated pass.** DP-C is ready for Claude's mandatory exact-
range cross-party review and remains unarchived.

## 2026-08-22 — DP-D predeclaration

DP-C awaits its independent designated review. DP-D begins after `ab70327`; construction may
continue, but no pending batch is promoted and no archival claim is made.

**Authority:** RFC DP4–DP5 and AC3/AC5/AC6 only. This batch owns the operator release/rollback
state machine, the append-only release and rotation ledgers, manifest-bound helper packaging and a
real Caddy HTTP/WebSocket release population. It composes the already-shipped gameserver drain and
DP-C backup/restore paths; it does not add operations metrics/alerts, alter gameplay, claim measured
RPO/RTO or perform the final clean-host R-006 rehearsal.

**Expected paths:** a focused `server/deploymentrelease` package and `server/cmd/deployment-release`
operator command; package-owned unit and Postgres/Caddy integration witnesses; release-package
manifest/bundle/Compose validation needed to ship the static Linux/amd64 helper and durable
operator-state mount; `Makefile`; `compose.deployment-release-test.yml`; and canonical deployment
documentation. Existing gameserver drain code may change only if the external witness exposes a
contract defect. No gameplay schema, balance data, save migration or GitHub workflow is expected.

**Positive population:** governed-clock JWT/bootstrap/cursor rotations append activation/removal
records without secret values, accept old/new verification during the full governed overlap and
remove only after 30 minutes/31 days/366 days. A release controller validates current and candidate
bundles, config, alert receiver and free space; creates an encrypted pre-upgrade backup; verifies
and loads the exact images; signals the old gameserver; observes readiness-down, a real
`server_restarting` WebSocket publication, intent refusal, bounded exit and successful admitted-work
completion; starts the candidate, verifies forward migration plus epoch/artifact identity, then
runs authenticated HTTP and WebSocket smoke through real Caddy before appending success. Rollback
restores that exact backup into a clean Postgres volume, starts the exact immediately previous
manifest within seven days and passes the identical smoke path without running a Down migration.

**Negative/severing population:** premature removal for each key family; duplicate or rewritten
ledger rows; any secret value in either ledger; missing previous bundle/backup; expired rollback
window; wrong image or manifest; irreversible migration/content declaration; failed preflight or
backup; severed courtesy frame, readiness-down/drain wait, migration failure propagation,
epoch/artifact reconciliation or authenticated smoke; non-clean restore target; and a controller
that attempts rollback through a Down migration. Every failed release/rollback appends an explicit
non-success result and can never emit a success record.

**Authorized claim:** the exact release package supplies a fail-closed stop-drain-start release,
seven-day exact-backup rollback and governed key-rotation mechanism with real Caddy-path evidence.
**Not authorized:** operations readiness, supported self-hosting, RPO/RTO, release readiness or a
1.0 claim until DP-E, DP-F, exact-manifest R-006 and both review gates pass.

## 2026-08-22 — DP-D scope expansion: production epoch backup blocker

The first real Caddy + pre-upgrade-backup composition exposed a DP-C fixture defect before DP-D
could proceed: production Compose mounts the canonical `balance/epochs/phase0.json`, but
`deploymentbackup.validateReleaseInputs` decoded an exact two-field projection and rejected the
canonical declaration's required `artifacts` and `epochs` members. DP-C's Postgres witness used the
same reduced synthetic object, so it could not falsify the shipped path.

DP-D expands to the smallest owning correction in `server/deploymentbackup/package.go` and its
fixtures: decode through the canonical `epochseed` authority and replace reduced epoch fixtures
with structurally valid declarations. A permanent negative must show that a reduced or malformed
declaration still fails. This does not alter backup contents, encryption, retention or objectives;
it makes the already-specified exact production epoch input executable. The correction remains in
DP-D's designated-review range and does not retroactively change DP-C's pending verdict.

## 2026-08-22 — DP-D scope expansion: private-database operator boundary

The adversarial production-boundary pass found that the draft host-side release helper opened the
database URL directly for free-space and epoch/artifact checks. That cannot work in DP-B's accepted
topology: Postgres has no published host port and is reachable only from the private `database`
network. The in-process Caddy population exercised the database semantics but could not falsify
this operator-network mistake.

DP-D therefore expands within its existing DP5 authority to add exact database-size and
manifest/epoch/artifact inspection commands to the already shipped `deployment-backup` helper.
The release controller invokes those commands as one-off `backup` service containers on the private
network and strictly decodes their output. The host helper no longer opens the database secret.
Tests must reject a severed/wrong migration, epoch, artifact or database-size result and must show
that candidate preflight runs the candidate image/config rather than `exec`-ing the old container.

## 2026-08-22 — DP-D implementation and Codex first filter

Implementation commit `23adca6` completes the predeclared DP-D mechanism. The packaged static
`deployment-release` command owns strictly forward stop-drain-start release, exact seven-day
backup rollback, retryable failed rollback attempts, serialized append-only release/rotation
ledgers and governed JWT/bootstrap/cursor removal timing. Release inputs are full validated bundles;
candidate config runs in the candidate image, image runtime-config digests are inspected, and
normal release cannot act as an unrecorded version downgrade.

The production topology remains private: database sizing and post-start migration/epoch/artifact
identity run through strict output from one-off `deployment-backup inspect` containers on the
database network. The backup service now mounts the complete manifest-bound content closure. This
also carries the predeclared canonical-epoch correction: real `phase0.json` is accepted and the old
reduced projection is a permanent negative. The application SPDX/notices generator inventories
all three shipped Go binaries plus the client; its previously red ISC license boundary is now
explicitly supported. Both operator binaries build as stripped static Linux/amd64 ELFs and are
required, hashed and validated in the release manifest.

Executed positive evidence on committed implementation content:

- `make verify-server-core` passed cold across vet, all non-harness Go packages and generated
  formula/API drift checks;
- `make verify-ci-topology` passed with all ten negative fixtures;
- focused `make test-go-ci` passed under the repository Linux/amd64 Compose lane for
  deploymentrelease, both operator commands, deploymentbackup, releasepackage, metadata,
  account rotation, deployment config and gameserver composition;
- `make test-deployment-backup` passed empty/populated encrypted Postgres 16 restore identity and
  non-clean-target refusal under an isolated project that cleaned its containers/network/volumes;
- `make test-deployment-release` passed real Postgres, real internal-TLS Caddy, authenticated
  HTTP/WebSocket, held admitted intent, readiness-down, exact drain refusal, courtesy frame, clean
  bounded closure, restart, encrypted pre-upgrade restore to a new database and previous-service
  smoke; its deliberate catalog-byte corruption was rejected by the private inspection boundary;
- `make generate-release-metadata` produced 44 exact dependencies; both helper build targets
  produced static x86-64 ELF binaries; and the staged `make release-secret-scan` inspected 1,333
  tracked files with no finding.

Demonstrated severing failures, all restored before the implementation commit:

1. allowing normal version downgrade failed
   `TestReleaseRejectsVersionDowngradeEvenWhenMigrationIsForwardCompatible` with “normal release
   accepted a semantic version downgrade”;
2. removing exact backup-ID authorization failed the independently open `wrong backup` population
   with “wrong backup accepted”;
3. removing the overlap comparison failed JWT, bootstrap and cursor cases independently with
   “premature removal accepted”;
4. changing the exact courtesy code failed `TestRestartCourtesyRequiresExactSystemEnvelope` with
   “exact restart courtesy frame rejected”;
5. returning to `docker compose exec` failed
   `TestDockerRuntimePreflightRunsCandidateConfigAndPrivateDatabaseInspection` with the exact old-
   container command; and
6. removing both helper commands from metadata discovery failed
   `TestGoDependencyInventoryCoversEveryShippedBinary` because `filippo.io/age` disappeared.

The rollback deadline, wrong backup and wrong previous-manifest negatives each receive a fresh
open release authority; they no longer coast behind a prior successful rollback. Locks reject a
concurrent operator mutation, ledger readers reject discontinuous/reused key IDs and unknown
failure stages, symlinked backup inputs fail, and no Down-migration method or command exists.

**Review by:** Codex. **Recorded by:** Codex. First-filter range `ab70327..23adca6` plus this record
commit reviewed in full; the range includes predeclaration commit `6c626c2`. Scope matches RFC
DP4–DP5 and AC3/AC5/AC6 plus the two recorded production-boundary corrections. No DP-E operations,
RPO/RTO, clean-host R-006, supported-self-hosting or 1.0 claim is made. Verdict: **APPROVED as first
filter; not the designated pass.** DP-D is ready for Claude's mandatory exact-range cross-party
review and remains unarchived.

## 2026-08-22 — DP-E predeclaration

DP-A through DP-D remain ready for their independent designated reviews. DP-E begins after
`3f58fea`; construction may continue while those reviews are pending, but no pending batch is
promoted and no archival or supported-release claim is made.

**Authority:** RFC DP1–DP3 and DP7, AC2/AC7/AC8, and the accepted operations decisions only. This
batch owns the complete provider-off operations profile: private bounded gameserver/Centrifuge
metrics, Prometheus/Alertmanager/node-exporter composition, privacy-safe structured logs, measured
14-day journald capacity/retention enforcement, optional separately bounded seven-day raw-IP sink,
durable backup/release/restore result metrics, successful Alertmanager delivery proof, all seven
blocking alert families and release-bundle image/SBOM/config closure. It may reconcile DP-D's
receiver-health preflight with the accepted private Alertmanager boundary. It does not change
gameplay, product/legal retention, GitHub workflows, CI topology, balance/content, save schemas,
RPO/RTO claims or the final clean-host R-006 rehearsal.

**Expected paths:** a focused `server/operations/` metrics, privacy, textfile, alert-delivery and
host-policy package with cold tests; `server/cmd/deployment-operations/` as the packaged host-side
observation/preflight helper; bounded hooks in `server/gameserver/`, `server/transport/`,
`server/production/`, `server/save/`, `server/cmd/gameserver/`, `server/deploymentbackup/`,
`server/cmd/deployment-backup/`, `server/deploymentrelease/` and
`server/cmd/deployment-release/`; `deployment/operations/` Prometheus rules/config, Alertmanager
template, journald drop-in and host timer/service configuration; production Compose/config/Caddy
templates; release-package manifest, assembler, schema, metadata and fixtures needed to ship six
immutable image/SBOM identities and the operations helper/config bytes; `Makefile`; a dedicated
`compose.deployment-operations-test.yml`; and canonical `docs/deployment.md`. Existing migrations,
gameplay data and `.github/workflows/` are excluded.

**Positive population:** an isolated Prometheus registry exposes process/readiness, bounded
HTTP/WebSocket status and latency classes, Postgres reachability, every named composed job's
success/failure (including credential cleanup), outbox and all dead-letter populations. Backup,
restore and release outcomes atomically update node-exporter textfile metrics, including the last
successful scheduled backup timestamp. Labels are closed enums only. Prometheus scrapes only
private targets and routes the seven exact alert families to a configured Alertmanager receiver;
a nonce-bearing synthetic alert proves a successful receiver delivery before release preflight can
pass. Every alert is tested both firing for its exact duration/population and resolved after the
fault clears. All six runtime images are digest/config/SBOM bound in the release manifest.

The packaged host helper validates and observes journald rather than declaring a convenient byte
ceiling: the evidence names the measured workload population and interval, peak bytes/day,
configured journal capacity, filesystem capacity and collection completeness. It accepts only
exact 14-day time retention with capacity for fourteen measured peak days and a storage-pressure
threshold that fires before journald could evict early. The default configuration emits no raw IP;
an explicitly enabled security sink is separate and capped at exactly seven days.

**Negative/severing population:** public or host-published metrics; a Caddy metrics proxy; any
secret, recovery material, raw IP, account/founder/stream/request identifier or unbounded value in
metric labels; the current founder/stream/detail invariant-log leak; missing/failed credential
cleanup observation; severed Postgres/outbox/dead-letter collectors; stale or failed backup,
restore or release textfile updates; corrupt/non-atomic textfile bytes; mutable/unbound operations
images or SBOMs; absent/malformed receiver config; a health-only receiver check with no delivered
synthetic alert; each severed alert expression, duration or resolved state; journal retention under
or over fourteen days, incomplete measurement, insufficient capacity, an alert threshold after
early eviction, default raw-IP logging or a security sink exceeding seven days. Operations
Compose must fail validation if any non-Caddy service publishes a host port or any operations
service escapes its intended private network.

**Authorized claim:** the exact release package contains a provider-off, privacy-bounded operations
profile whose private signals, measured log-retention contract, receiver delivery and all seven
firing/resolution paths have executable evidence. **Not authorized:** designated approval,
supported self-hosting, release readiness, RPO/RTO, public-service availability or a 1.0 claim
until DP-F, exact-manifest R-006 and both review gates pass.

## 2026-08-23 — DP-E implementation and Codex first filter

Implementation commit `09d5027` completes the predeclared provider-off operations batch. The exact
first-filter range is `3f58fea..09d5027`: 65 paths including predeclaration `e2fc1c4`, 63 paths in
the implementation commit and no `.github/`, migration, gameplay-content or design path.

The gameserver now owns an isolated Prometheus registry with process/readiness, bounded HTTP and
WebSocket route/status/latency, Centrifuge, current-schema Postgres/outbox/dead-letter, five named
periodic-job and invariant signals. Real composition primes and records verification, presence,
clearing, guild sweep and credential cleanup. Logs and command failures expose bounded classes,
not founder/stream/run/intent IDs, payload/detail text, recovery material or arbitrary error text.
Backup, restore, release and host observations use fsync/rename textfiles; later failures preserve
the last success, while over-budget journal use remains observable so its alert cannot coast on a
stale healthy sample.

The bundle now closes six immutable image/config/SBOM identities and all operations helper/config
bytes. Only Caddy publishes ports. Prometheus and node-exporter remain internal-only;
Alertmanager also joins the non-publishing edge network because an internal-only receiver path
would make the RFC's permitted remote receiver unreachable. Release preflight validates the exact
Alertmanager config, and post-start smoke runs the packaged nonce-bearing delivery verifier from
that egress-capable boundary. The real isolated test receives the same nonce through Alertmanager.

The host policy accepts only a complete measured release-workload observation, exact 14-day
retention, fourteen projected peak days of capacity and a storage alert at 80% before eviction.
Rendering produces both the new journald drop-in and the exact budget file consumed by the
one-minute host observer and refuses existing/symlink targets. The shipped stack has no raw-IP
producer or enable switch. The initial boolean-only “separate seven-day sink” branch was vacuous,
so it was removed: raw-IP-enabled observations now fail until a future governed producer, separate
purge mechanism and executable witness actually exist. This enforces the RFC's default-off state
without claiming an unbuilt optional sink.

Executed restored-tree evidence:

- `make verify-server-core`: `go vet ./...`, every non-harness server package cold, pitch gate,
  formulas and API generation all green. The first sandboxed pass was denied an ephemeral
  localhost listener; the exact unrestricted rerun passed and is the claimed result.
- `make test-deployment-operations`: pinned Prometheus `v3.12.0` rule evaluator green; all focused
  changed packages cold; real pinned Linux/amd64 Caddy, Prometheus, Alertmanager and node-exporter
  reached five private scrape targets and delivered the generated proof nonce. An initial
  `host.docker.internal` receiver-health mistake failed and was corrected to the receiver's actual
  private endpoint. A later redundant Compose `platform` hint hit Docker Desktop's cached-manifest
  collision; removing only that hint retained the exact amd64 digests and the lane passed.
- `make test-save-integration SAVE_TEST_PACKAGES='./gameserver ./operations' ...`: current Postgres
  schema collection and composed job-prime observations green on real Postgres 16.
- `make test-deployment-backup` and `make test-deployment-release`: empty/populated restore,
  dirty-target refusal, real Caddy drain/release and exact rollback all green. Their first runs
  exposed stale three-image fixture manifests; the fixtures now bind the same six-image and full
  operations-file closure as production.
- Production Compose passed both the repository validator and Docker Compose's parser; the changed
  Caddyfile passed real Caddy validation; the receiver example passed pinned `amtool check-config`.
  Metadata regeneration found the expected 44 linked dependencies.
- `make verify-ci-topology` retained all ten negative controls, and `make release-secret-scan`
  reported 1,333 tracked files with no findings. No workflow byte changed. A bare
  `make deployment-config-check` was also attempted without the mandatory production environment
  and correctly failed configuration loading; it is not cited as positive evidence.

Discrimination was executed, not inferred: changing the five-minute outage duration to six made
`promtool` fail at exactly 5m; reintroducing `intent_id` made the privacy test reject the emitted
value; weakening the capacity multiplier from fourteen days to thirteen made the insufficient-
capacity fixture pass and the test fail; removing manifest image cardinality made the missing-
Prometheus image/SBOM fixture fail; and accepting an unchanged Alertmanager counter made the
health-only negative fail in about 100ms. Every mutation was restored before the green runs.

First-filter findings fixed before commit included the missing rendered journal-budget artifact,
non-atomic privileged policy output, Alertmanager's initially impossible remote egress, stale
healthy storage metrics above budget, silent database-collector absence, a fake raw-IP-sink
assertion, incomplete fired/resolved semantic validation, and unbounded top-level command errors.
No gameplay behavior, balance, content, save schema, CI topology, RPO/RTO result or R-006 claim
entered the range.

**Review by:** Codex. **Recorded by:** Codex. The exact implementation range
`3f58fea..09d5027` plus this record commit was inspected and executed in full. Verdict:
**APPROVED as first filter; not the designated pass.** DP-E is ready for Claude's mandatory
exact-range cross-party review and remains unarchived. DP-F may be planned, but no Deployment
Foundation archival or supported-self-host/release-ready/1.0 claim is authorized.

## 2026-08-23 — DP-F predeclaration

DP-F planning begins after clean DP-E record `65099e7`. DP-A through DP-E remain unapproved until
their mandatory exact-range Claude reviews land. Construction of missing rehearsal/install
machinery may proceed without treating those batches as approved; the external clean-host run,
R-006 conclusion, lifecycle closeout and every supported-self-host/release-ready/1.0 claim remain
gated on the prior designated verdicts and the exact candidate manifest.

**Authority:** RFC DP1–DP8 and AC1–AC10, research item R-006, owner decisions
D-002/D-003/D-006/D-011/D-014, and no broader release promise. DP-F owns the minimum missing
initial-install path plus the complete rehearsal and evidence machinery needed to evaluate the
already accepted contract. It does not add gameplay, host provisioning/CD, arm64 support,
multi-node delivery, account/export retention, a sunset covenant, CI workflow changes, public
hosting or an availability SLA.

**Four waves:** DP-F1 implements a fail-closed initial-install operation and a strict exact
rehearsal schema/validator/driver with sanitized artifacts and corrupt/forged-evidence negatives.
DP-F2 builds two retained semantic release bundles from explicit commits/timestamps, resolves all
six Linux/amd64 image digests/config identities, generates real per-image SPDX SBOMs with a pinned
tool, scans tracked and image bytes, and proves byte-identical rebuilds. DP-F3 runs only those
bundles on an explicitly authorized clean supported Linux/amd64 host with no checkout. DP-F4 writes
the evidence-backed R-006 dossier/runbook/limitations and reconciles lifecycle records only after
the run and all prior designated verdicts exist.

**Positive population:** the clean host records kernel/distribution/`x86_64`, exact Docker Engine
and Compose versions, zero repository source, the exact manifest/hash and every loaded runtime
config identity. Initial install from bundle-only bytes reaches HTTPS Caddy readiness and a real
browser drives the default Phase-0 surface. Empty and populated databases back up and restore;
account, Founder, Company save, event, leaderboard, epoch/constants/copy and migration identities
match. A declared incident and restore interval measures RPO at most six hours and authenticated
Caddy-complete RTO at most four hours. Candidate release proves courtesy/drain/start; rollback
restores the exact previous manifest/backup without Down migration. Current+previous key overlap,
provider-off operation, private metrics, real receiver delivery, seven alerts, measured journald
budget, license delivery and all six SBOM/provenance bindings remain visible in one evidence set.

**Negative/severing population:** bundle with one catalog/client/license/config/helper removed;
changed image digest/config/SBOM; source checkout made available to a purported clean-host runner;
missing/malformed/current/previous secret families; wrong origin/proxy depth; seeded source/image
secret; truncated/corrupt/wrong-identity/wrong-manifest backup; non-clean restore target; killed
backup writer and gameserver restart during admitted work; wrong epoch/artifact set; missing
previous image/backup, irreversible migration and Down-migration attempt; public metrics;
health-only alert receiver, severed rule/counter path; incomplete/guard-ended journal or objective
observation; RPO/RTO above bound; and a forged successful evidence row. Each must make the
rehearsal or evidence validator nonzero, not merely add a warning.

**Evidence contract:** the driver writes append-only, schema-versioned JSON with UTC step start/end,
command/result class, exact input/output digests, host/tool identities, explicit exclusions and
objective-completion/guard fields. It never records secret values, recovery codes, raw IPs,
account/founder identifiers or database URLs. Stored identity comparisons use release/run fixture
IDs or hashes only. A run terminated by timeout/guard, missing any declared population, reusing a
prior artifact, observing a mutable image, or lacking a named severing result is invalid.

**External-action boundary:** no push, publication, DNS mutation, certificate issuance, remote-host
write, deployment or destructive restore is implied by local construction. DP-F3 will use only a
clean host the owner explicitly places in scope; until then, local build/validator work continues
and the exact host run is honestly blocked. No timeout or acceptance bound is loosened to fit the
available machine.

**Authorized claim before DP-F3:** only that the exact rehearsal machinery is implemented and its
forgery/severing controls discriminate. **Authorized after a valid DP-F3 but before cross-party
review:** R-006 evidence is ready for designated review, not that self-hosting is supported.
Archival and release wording require DP-F4, full implementation-range Codex first-filter, Claude
designated approval covering every batch edge, and the owner's final release call.

## 2026-08-23 — DP-F1 initial-install and evidence-boundary slice

Implemented the previously absent initial-install authority rather than pretending the normal
release command could operate against a fictional current version. `deployment-release install`
loads and verifies the exact six-image bundle, proves the Compose project and named Postgres volume
clean on both sides of candidate config preflight, starts Postgres before the remaining services,
verifies migration/epoch/artifact identity, runs the authenticated Caddy and alert-delivery smoke,
and syncs an append-only install record. Pre-start failures do not delete state. Every post-start
failure removes only the exact newly created Compose project and volumes. A final ledger failure
also tears the stack down instead of leaving an unaudited installation. Corrected bundles may retry
after failed install rows; a success or any non-install transition permanently closes install
authority.

Added the first strict `deploymentrehearsal` evidence boundary and Linux/amd64 validator command.
It requires the exact supported host/tool identities, eleven ordered command classes, every
predeclared positive and severing population, derived (not asserted-only) RPO/RTO durations,
restored identity equality, ten named artifact hashes, explicit non-claims and completed/non-guarded
termination. Unknown/trailing fields and private identifier/credential-shaped fields reject. This
is validator and capture-contract construction only: it does not claim that the external R-006 run
has happened, and DP-F1 remains open for the actual fixed-command driver and bundle integration.

Cold evidence: `make test-deployment-rehearsal` passed all four focused packages; focused `make vet`
passed; the exact Linux/amd64 rehearsal validator binary built; and the unrestricted rerun of
`make test-deployment-release` passed the real Postgres+Caddy drain/backup/restore population. The
first Docker attempt was sandbox-denied and is not cited as a product failure. Discrimination was
executed: removing post-smoke install cleanup made the severed-smoke test fail with no
`abort_install` call, and removing the negative-population severing requirement made the vacuous
negative evidence fixture pass and its test fail. Both mutations were restored before the green
runs.

## 2026-08-23 — DP-F dependency-order correction

The initial predeclaration placed the fixed-command rehearsal driver in DP-F1 and exact bundle
construction in DP-F2. Implementation established that this order would force the driver to invent
commands and artifact identities before the candidate and previous bundles exist. The living plan
now keeps DP-F1's completed install/schema/validator/evidence contract, moves the driver binding
into the exact-bundle DP-F2 range, and retains every population, falsifier and claim gate. No
acceptance criterion or external clean-host requirement moved later or became optional.

DP-F2 supply-chain construction now includes `deployment-rehearsal` and its strict evidence schema
as manifest-hashed release artifacts; an omitted or non-Linux/amd64 helper fails assembly. The
application dependency inventory covers the rehearsal command too. Image SBOM generation is pinned
to Syft v1.51.0 at multi-platform digest
`sha256:678bfa565b60f747aac0f8e964fe5588a24445b8d0a480e91f6efd70020dfbb0` and the release normalizer
replaces Syft's wall-clock/random document header with the exact OCI runtime-config identity and
declared release timestamp while retaining the discovered package graph. It refuses empty graphs,
mutable identities and output overwrite.

The real lane was exercised twice against the pinned Linux/amd64 Caddy 2.11.4 image. Independent
registry scans normalized to byte-identical SPDX JSON, both SHA-256
`633dfe37c5221607d4fb5aaaa346d8fcfb9858a7f7338755c61ed2f3c1b3ee0b`, bound to runtime config
`sha256:af555904a0961945f16bb323a501457b13a4f7e9bde969b145b97da80b38ecbe` and the declared UTC
timestamp. Focused cold package tests and vet passed. This verifies the generator mechanism only;
the exact six-image candidate/previous SBOM set and byte-identical release bundles remain DP-F2
work, and no R-006 or release claim follows.

The retained previous rehearsal bundle is now built at ignored local path
`.cache/release/previous/bundle` from clean source commit
`44a4a72a52ef1b299075389bfa5f365baa00e4c8` as semantic version
`0.1.0-preview.0`. Its manifest contains 70 artifacts and hashes to
`sha256:75fafa3126c9372fe5c70da948d27962f22ca75e805da95fdfd2651095ed169d`; its
offline gameserver archive hashes to
`sha256:701e60d11106d7dc432eb6416b19b5d4ebff5c98db32f5f627626f9e7fe0b006`.
An independent cold rebuild regenerated five Linux binaries, the client, metadata, runtime closure,
gameserver archive and all six real Syft package graphs. The complete accepted bundle trees are
byte-identical; the raw Syft documents differ in their nondeterministic headers while all six
normalized documents match byte-for-byte, demonstrating that the normalizer removes only the
known nondeterminism rather than reusing first-run inputs. The strict, trackable build record is
`release-builds/previous.json` and passes `deployment-rehearsal validate-build`.

The first assembly attempt correctly failed before a manifest existed because the Make wrapper
resolved relative SBOM paths from `server/` instead of the repository root. Its partial directory
was preserved as `.cache/release/previous/bundle.failed-relative-sbom-path`; the wrapper now applies
the same root-relative conversion as every other build input, and assembly into a new empty
destination passed. This is a build-system defect found and fixed, not a softened input rule.

The fixed-command driver boundary is now implemented for the candidate binary. A strict execution
plan is bound to both manifest hashes and must enumerate all 43 predeclared positive/negative
populations in canonical order, with a ruled step, direct argv vector, exact expected exit and
bounded timeout. Shells and sudo are structurally refused; secret-shaped arguments, output beyond
one MiB, timeout/guard exhaustion, wrong exit, non-monotonic time and any nonempty result directory
fail. Each executed check writes one exclusive mode-0600 result containing only command/output
hashes, time, exit, kind and result. Tests execute the complete 43-command positive/severing
population and demonstrate that a vacuous negative, shell command, secret argument, missing check,
wrong order, absent guard, output truncation and overwrite each fail. The DP-F3 plan itself remains
to be generated after the candidate manifest exists and receives review as a hashed evidence
artifact; implementation of the driver does not mark any rehearsal population observed.

The retained candidate rehearsal bundle is now built twice from clean source commit
`469db5e42320b91a9c610c0e015c437aa6356d63` as `0.1.0-preview.1`. Both complete 70-artifact trees
are byte-identical. The manifest is
`sha256:8ab2efa35e8850b974614ce26cb689fc6b29691b4608c1bea771ae9dadc3fb68`, the gameserver archive is
`sha256:37db56def353a38903431628df13ac49230a6b11ed4680bd592bb4cc70e34462`, and its OCI runtime config
is `sha256:fec3fe1812c2af121345c95e2f6ce7038787e0f6f19af92fb5bc671085b4b5b9`. All six normalized SBOM
pairs match independently regenerated package graphs. The strict candidate record is
`release-builds/candidate.json`; the manual rehearsal target validates both candidate and previous
records cold.

DP-F2 has therefore produced the exact two-bundle input population and proven independent rebuilds.
It is not yet complete: the reviewed 43-check command plan must bind these two manifest hashes,
and the complete source/image secret scan must run against the candidate bytes before DP-F3 can
touch an authorized clean host. No component result is being promoted to R-006 evidence.

## 2026-08-23 — DP-F2 typed bundle-probe slice

The fixed-command boundary had a remaining false-positive class before the real plan could be
authored: a negative row expected exit one, so an unsupported command or failed fixture setup could
look identical to the intended gate rejecting a severing. The new `deployment-rehearsal probe`
contract reserves exit zero for accepted subjects, exit one only for a fully prepared named
negative that reaches and is rejected by the release-package validator, and exit two for invalid
input, unsupported populations or setup failure. Cleanup failure also invalidates the result.

Nine of the 43 declared populations now have real bundle-only probes. The positive validates both
exact bundles' six-image, SPDX, license, provenance and complete artifact closure. Eight negatives
hardlink the candidate into a private temporary tree, then remove the catalog, client entry point,
root license, config schema or release helper, or alter the gameserver image digest, an image
runtime-config identity or an image SBOM. Rewrites unlink before writing so the retained hardlinked
candidate cannot be changed. Tests prove every mutation exists, the original remains byte-exact,
an accepted mutation returns zero, missing mutation input cannot count as rejection and unsafe work
placement fails.

The current retained bundles were exercised directly with the native helper: the intact population
exited zero; all eight mutations exited one with `fixture_rejected`; the deliberately unsupported
`rpo_or_rto_above_bound` runtime population exited two with `invalid_probe`. A first `go run`
attempt was denied access to the user Go cache and is not evidence; the root Make build uses the
repository cache and produced the binary used for the recorded runs. The full 43-row plan remains
unwritten and DP-F2 remains open. In particular, no browser, database, recovery, rotation,
operations, RPO/RTO or clean-host observation is claimed by this slice.

## 2026-08-23 — DP-F2 config matrices and semantic bundle-validation repair

Three additional negative rows now execute the production `deploymentconfig.Load` decoder instead
of duplicating its rules. A valid current/previous JWT, bootstrap and cursor baseline must load
first. The missing/malformed matrix then requires missing database material, malformed JWT,
wrong-length bootstrap, a half current cursor and a half previous JWT to reject. The duplicate
matrix requires both repeated identity and repeated value to reject for all three key families.
The origin/proxy matrix requires insecure, path-bearing and noncanonical origins plus both zero and
two trusted hops to reject. If any one severing is accepted, the aggregate probe exits zero and the
negative plan row fails. An injected accepted mutation demonstrates that discriminator. All three
real helper invocations exited one; setup/unimplemented failures retain the separate exit-two path.

The public-metrics population exposed a real bundle validator hole. Its fixture adds a Caddy
`/metrics` reverse proxy and then updates the Caddyfile artifact hash in `release-manifest.json`,
ensuring it gets past byte-integrity validation. The retained-bundle probe initially exited zero:
`ValidateCaddyfile` existed and assembly called it, but `ValidateBundle` did not re-run it. Full
bundle validation now re-runs `ValidateCompose`, `ValidateCaddyfile` and
`ValidateGameserverDockerfile` after exact artifact comparison. The re-bound public route now exits
one. Removing only the new Caddy validation call makes the focused regression fail with
`manifest-rebound public metrics route accepted`; restoring it returns the test green. This is a
production release-boundary fix discovered by the negative population, not a relaxed probe.

The probe registry now has 13 of 43 real populations. The current candidate record still describes
the earlier reproducible input bundle and will be rebuilt after the rehearsal helper/probe surface
stabilizes; no stale manifest is presented as the final R-006 candidate. Runtime and clean-host
populations remain open.

## 2026-08-23 — DP-F2 secret probes and result-binding gate

The seeded source and seeded image rows now run through the real release scanner. The source
fixture is a valid one-file tracked population; the image fixture is a valid tar containing one
seeded member, avoiding the old negative's weaker “malformed bytes after tar EOF” failure. A probe
may return the expected rejection only after the scanner reports exactly the named seeded-fixture
rule. Scanner/parser/setup failure returns exit two, and an injected sleeping no-secrets gate makes
both probes return zero. The native helper returned exit one for both real fixtures. This brings
the implemented probe population to 15/43; it does not replace the already completed full
candidate source/image scan or the final rebuilt-candidate scan.

Auditing `forged_successful_evidence` exposed a separate integrity gap before that row could be
implemented honestly. The standalone evidence validator required plausible hashes but did not
open the reviewed plan or the 43 result files, so a structurally plausible JSON document was not a
cryptographic witness. Final CLI `validate` now requires `--evidence`, `--plan` and `--results`.
It binds run/manifest identities and the exact plan-file digest, rejects missing/extra/symlinked or
non-0600 result files, strictly decodes every result, and matches its name, kind, step, command
hash, expected exit, pass/guard state, interval and raw file hash to the plan and population row.
Tests demonstrate an otherwise valid free-standing evidence document, a forged population hash,
a rewritten result, a missing result, an extra summary and an unsafe result mode all fail. This
closes the first forgery route. Removing only the raw-result/population hash comparison makes the
focused forgery test fail with `forged population evidence hash accepted`; restoring it returns
the test green. Remaining final artifacts and step aggregation still need their own byte bindings
before the forged-evidence population can be marked implemented.

The binding was then extended over the remaining retained surface rather than stopping at result
files. CLI `validate` also requires an exclusive artifact directory containing exactly ten named,
mode-0600 files: both manifests, both ledgers, backup header, browser result, alert delivery,
journal observation, secret scan and supply-chain result. Their raw hashes must equal the evidence
rows; candidate/previous manifest bytes must additionally equal the run identities. The eleven
step summaries are now derived from the earliest/latest underlying result interval and the ordered
command/result-hash lists for their assigned rows. Changed/missing artifacts and a forged step
output hash fail. The test plan fixture was corrected from round-robin to contiguous step
assignment because round-robin made truthful aggregate intervals overlap later steps; no product
plan or criterion changed. Removing only the step-derivation comparison makes the focused forged
step test fail with `forged artifact/step accepted`; restoring it returns the test green.

This still does not mark `forged_successful_evidence` complete: tool-binary identities and schemas
for the retained browser/alert/journal/scan/supply-chain files remain to bind, and the final plan
must avoid a circular dependency between producing evidence and testing its forgery rejection.

## 2026-08-23 — DP-F2 clean-host browser-driver construction

The repository's composed Game UI script cannot serve as R-006's browser tool: it builds the Go
server, starts Vite, reads repository paths and shells into the test Compose database. A clean host
with only Docker/Compose and release bytes needs a different executable boundary. A new static
Linux/amd64 `deployment-browser` uses pinned `chromedp` v0.15.1 and drives Chromium from the pinned
Playwright 1.62.0 Noble amd64 manifest
`sha256:02bbb2155cd7109e3e9c741941097ed1608cf8b6fa44ee2595896da2bdc1f471`; the selected runtime
config is `sha256:50cbb76d250a50002045a95f484c5f40573cde831adbe40c784052c037e36118` and browser path is
`/ms-playwright/chromium-1234/chrome-linux64/chrome`. The multi-platform index initially hit Docker
Desktop's cached-manifest collision after download; resolving and executing the amd64 manifest
directly worked without changing the selected bytes.

The driver records only the exact candidate manifest hash, UTC bounds, desk surface, credential
handoff booleans, completed WebSocket handshake, manual-intent observation/status, page-exception
count and objective/guard state in an exclusive mode-0600 result. It never captures credential
values, response bodies, player IDs or socket frames. Unit negatives cover retained bootstrap,
missing durable credentials, no handshake, no intent, non-200 intent, page exception, incomplete
objective, guard exhaustion, insecure production origin and output overwrite.

A real Linux/amd64 driver binary ran inside the pinned Playwright image against the local isolated
fixture: Chromium navigated, clicked both visible controls, completed the socket handshake and
intent, and wrote a 437-byte sanitized result in 3.35 seconds. Severing only the fixture's WebSocket
upgrade made the same driver exit one and write no success artifact. The first sandboxed fixture
listener was denied and is not evidence; the exact unrestricted local rerun is the claimed result.
This is driver discrimination only, not the real product browser population. Bundle inclusion,
Playwright SBOM/provenance, tool-hash binding and exact Caddy execution remain open, so the probe
count stays 15/43. Two independent trimpath/build-ID-free Linux/amd64 driver builds were
byte-identical.

## 2026-08-23 — DP-F2 browser supply-chain closure

The clean-host browser tool is now part of the exact candidate release closure rather than a
developer-side executable. Assembly requires and copies the static Linux/amd64
`deployment-browser`, binds the exact Playwright Linux/amd64 manifest and runtime-config digest,
validates its independently generated SPDX graph, and records the rehearsal-only image separately
from the six production Compose services. Manifest validation requires the browser binary,
Playwright SBOM and rehearsal-image row to appear together; full bundle validation re-opens the
SBOM and checks it against the selected runtime config. The external image is not promoted into the
production topology.

The application metadata inventory now includes the browser command. A real generator run found
47 linked third-party Go modules, the standard library and three browser-client dependencies (51
dependencies total), and the canonical deployment documentation now states the measured graph and
all six shipped Go commands. Build-record schema v2 makes the rehearsal image mandatory for a
candidate while preserving the already-recorded schema-v1 previous/interim records as immutable
historical inputs; a schema-v1 record cannot smuggle in the new field.

Cold focused tests passed for `deploymentrehearsal`, `releasepackage`, both assembly/metadata
commands, and focused vet passed. Discrimination was executed by removing only the browser copy
from assembly: `TestAssembleBundleBindsBuiltInputsWithoutCheckout` failed with `incomplete
rehearsal browser closure`. Restoring the copy returned the full focused population green.
Removing the retained browser after assembly and supplying missing/mutable browser inputs are also
permanent negative fixtures. This is construction evidence only: the retained candidate and its
independent rebuild still need regeneration from the committed source point, and the real product
browser population remains unobserved. The implemented probe count therefore remains 15/43.

## 2026-08-23 — DP-F2 exact candidate v2 rebuild

The interim candidate is superseded by two independent builds from clean committed source
`83a0e7fc7e814e2dc514939b07e15ddd2671eeeb`, timestamped from that commit at
`2026-08-23T08:51:16Z`. Each build independently regenerated the client, all six static
Linux/amd64 Go commands, staged content, application metadata, the no-cache gameserver image, six
production image package graphs and the Playwright package graph. The exact 72-artifact bundle
trees are byte-identical. Both release manifests hash to
`sha256:9f920670e3674b6ef8fd44639c451c5d80adb1cbda4a1c3f3fe67ff6fda47cf4`;
both offline gameserver archives hash to
`sha256:e4d21ec1cfa3543328bca33903a7bbf3b4f161309cc6e497a9e70edd41e2e474`;
and both gameserver OCI configs are
`sha256:f7d1c0fc076bf246b54c16220360206a4659b16f06f719f2029d1d4c0be4d207`.

All seven normalized SPDX pairs matched byte-for-byte after independent discovery. Their candidate
hashes are Alertmanager `71c14e93…`, Caddy `0d6317a4…`, gameserver `418f5af1…`, node-exporter
`1e21ae83…`, Postgres `da45be36…`, Prometheus `dc52283b…` and Playwright `912af379…`. The two
browser executables also matched at `fa80f3c3…`. The full release secret scan covered 1,381 tracked
source files plus the exact candidate gameserver image and reported no findings.

The first image attempt failed before output because the staged-content primitive was mistakenly
treated as a complete OCI context; the corrected invocation then exposed that content had to live
under the Dockerfile's declared `content/` path. Neither setup failure produced an accepted archive
or evidence row. The independently rebuilt context used the corrected explicit construction and
reproduced the first valid bytes. `make test-deployment-rehearsal` passed cold and validated the
untouched previous build record plus the new schema-v2 candidate record. This completes the exact
candidate input replacement, not R-006: no clean-host runtime population has yet been observed and
the probe count remains 15/43.

## 2026-08-23 — DP-F2 candidate tool-identity binding

Final evidence validation now requires the exact candidate bundle in addition to the evidence,
plan, raw result and retained-artifact directories. The declared `deployment-rehearsal`,
`deployment-release` and `browser-driver` hashes must equal the executable bytes named by the
candidate bundle; missing, changed, non-executable or symlinked tools reject. This removes the path
where a forged evidence file could name plausible tool hashes unrelated to the programs that
actually produced the run.

Cold focused tests and vet passed. The binding discriminator was executed by replacing only the
candidate browser bytes and temporarily neutralizing only the hash comparison: the permanent
negative failed with `unbound candidate tool accepted`. Restoring the comparison returned the full
focused population green. The first mutation edit accidentally removed the comparison in a way
that left unused variables and produced a compile error; that attempt is not cited as the
discriminator. Artifact-specific schemas and the final evidence/forgery execution order remain
open, so `forged_successful_evidence` and the overall probe count remain unchanged.

## 2026-08-23 — DP-F2 typed browser and journal evidence

The final validator now opens the retained browser and journal bytes after their evidence hashes
match. Browser evidence uses the driver's strict decoder, must bind the exact candidate manifest,
and must fall within the run interval. Journal evidence uses the operations package's existing
strict observation decoder and must also fall within the run. Both reject unknown/trailing fields
and invalid objective, guard, timing, retention, capacity and privacy states; no duplicate summary
schema was added in the rehearsal package.

Cold focused tests and vet passed. The discriminator rehashed an invalid browser summary into the
top-level artifact row, then temporarily removed only typed-artifact validation: the permanent
negative failed with `forged artifact/step accepted`. Restoring typed validation returned the full
population green. A separately rehashed invalid journal document also rejects. Alert-delivery,
secret-scan and supply-chain artifacts still need owning structured producers/decoders, so the
forged-evidence population and 15/43 total remain unchanged.

## 2026-08-23 — DP-F2 structured secret-scan evidence

`release-secret-scan` can now write an exclusive mode-0600 result for the final run. The typed
result records the exact candidate manifest, Git-derived source commit, measurement interval,
tracked-file count, mandatory image-archive inclusion, zero findings and completed/non-guarded
termination. The scanner still uses the existing source and Docker-archive rules; it does not
serialize matched paths or bytes. The non-evidentiary developer console mode remains available,
but final rehearsal validation accepts only the structured source-plus-image population and binds
it to the candidate manifest's embedded source commit.

Cold tests passed for `releasepackage`, the scanner command and `deploymentrehearsal`; focused vet
passed. The discriminator rehashed an invalid secret-scan summary into the evidence, then replaced
only the typed secret validation with an unconditional return: the permanent negative failed with
`forged artifact/step accepted`. Restoring validation returned the full population green. A first
mutation removed the entire case and failed at compile time due to the then-unused package import;
it is not cited as the discriminator. The retained candidate predates this producer, so no final
secret artifact is claimed and the candidate must be rebuilt after producer construction ends.
Alert-delivery and supply-chain typed producers remain open; the probe count remains 15/43.

## 2026-08-23 — DP-F2 derived supply-chain evidence

The supply-chain artifact now has an owning producer instead of being an operator-authored JSON
summary. `deployment-rehearsal supply-chain` reopens both complete bundles through the production
validator, decodes both independent-build records, checks their exact manifest and source
identities, consumes the typed source/image secret scan, and derives the six production images,
one rehearsal image, eight SBOM documents and both attribution surfaces. It writes only after all
inputs pass, into an exclusive mode-0600 file. The root Make lane exposes the same direct command.

Candidate and previous build records are now first-class retained run artifacts, increasing the
artifact directory from ten to twelve files. Final validation decodes both records, requires a
schema-v2 candidate, binds them to the retained manifest bytes/source commits, then requires the
supply-chain result to name the hashes of both records, both manifests and the secret scan. Cold
focused tests passed and the producer rejects a failing bundle validator, mismatched scan identity
and output overwrite. The discriminator rehashed an invalid supply-chain summary and temporarily
replaced only its typed validation with an unconditional return; the permanent negative failed
with `forged artifact/step accepted`. Restoring validation returned the focused population green.
Alert-delivery is now the last untyped retained artifact. The retained candidate still predates
these producers, no R-006 run is claimed, and the probe count remains 15/43.

## 2026-08-23 — DP-F2 seven-family alert-delivery evidence

The last retained artifact now has an owning production contract. `deployment-operations
alert-observe` validates the exact bundle, derives its digest-pinned Prometheus image, runs the
bundled `promtool` population, then sends all seven canonical release-floor identities through the
private Alertmanager API. It waits for seven receiver notifications in firing state, posts the same
identities resolved, and waits for seven more notifications. Its exclusive mode-0600 observation
binds the candidate manifest and records ordered per-family firing/resolved delivery without
storing receiver URLs, nonces or payloads. Final rehearsal validation decodes that observation and
binds its manifest and interval.

Cold focused tests passed across operations, the command, release-package manifest loading and the
final evidence validator. The HTTP population proves all seven names traverse both states; a
health-only/single-notification fixture times out instead of passing. The final typed discriminator
rehashed a health-only alert summary and temporarily replaced only alert validation with an
unconditional return; the permanent negative failed with `forged artifact/step accepted`.
Restoring validation returned the population green. The evidence JSON Schema now reflects thirteen
top-level artifact hashes (the reviewed plan plus twelve retained files). All retained artifact
types are now strict, but `forged_successful_evidence` remains uncounted until its non-circular real
execution is authored and run. The overall implemented-probe total remains 15/43.

## 2026-08-23 — DP-F2 predeclaration: non-circular forgery proof

The final open evidence-integrity row cannot lawfully execute inside the same result set it is meant
to authenticate: its result hash would enter the evidence it must first mutate, creating a circular
dependency or forcing a vacuous already-invalid subject. The implementation boundary is therefore
predeclared as a two-stage seal without changing the required 43 populations:

1. the fixed command plan executes the 42 non-forgery populations and produces their exclusive raw
   results plus the twelve typed run artifacts;
2. a base dossier containing exactly those 42 populations and thirteen hashes (the reviewed plan
   plus twelve artifacts) must pass the full plan/result/artifact/tool validator;
3. the forgery producer mutates one named base-dossier binding, requires the same base validator to
   reject it, and writes a typed proof bound to the unmodified base bytes; and
4. the final dossier adds the retained base and proof hashes plus the 43rd
   `forged_successful_evidence` population. Final validation replays the named mutation itself and
   rejects a proof that merely claims success.

The discriminator will remove only the replayed rejection and must make a forged proof pass. A
missing base, pre-invalid base, unknown mutation, mismatched subject hash, self-authored result hash
or proof generated before all 42 rows complete is invalid. This dependency-order correction does
not remove a population, weaken a validator or count the row before the real proof executes.

## 2026-08-23 — DP-F2 non-circular evidence seal implementation

The fixed execution plan now contains exactly the 42 non-forgery populations. A base dossier has a
separate strict decoder and must bind all 42 raw results, the reviewed plan, twelve typed artifacts
and the exact candidate tool bytes before the forgery producer will run. `deployment-rehearsal
forge-proof` mutates only the base dossier's browser-result hash and writes an exclusive mode-0600
proof only when the same base validator rejects that mutation. A pre-invalid base is refused.

The final dossier retains fifteen artifact hashes: the thirteen base hashes plus the immutable base
dossier and forgery proof. Final `validate` requires a separate two-file mode-0600 seal directory,
checks that the final dossier is exactly the base plus those two hashes and the 43rd population,
and independently replays the named mutation. Focused tests cover base/final separation, input and
output binding, pre-invalid base refusal, unknown/claim-only proof rejection and seal modes. The
replay discriminator replaced the real mutation with an identity function; the otherwise-valid
final fixture failed with `invalid deployment rehearsal evidence`. Restoring the mutation returned
the focused population green. The JSON Schema now requires all fifteen final artifact hashes.

This completes the non-circular mechanism but does not count the real
`forged_successful_evidence` population: it must run after the exact 42-row R-006 base exists. The
implemented real-probe count therefore remains 15/43, and the retained candidate must still be
rebuilt after the full DP-F2 command surface stabilizes.

## 2026-08-23 — DP-F2 typed evidence-boundary probes

Six additional negative populations now have fixed native-helper probes. `source_checkout_present`
requires a valid clean Linux/amd64 host baseline before severing the source-checkout-absent field.
The two alert rows start from all seven canonical families delivered in both firing and resolved
states, then sever receipt, rule-fixture or resolution evidence. The two journal rows start from a
complete privacy-preserving fourteen-day-capacity observation, then force early eviction, shortened
retention, incomplete termination or guard exhaustion. `rpo_or_rto_above_bound` starts from a
time-consistent objective record and independently exceeds the ruled six-hour RPO and four-hour RTO.
Every probe calls the production evidence validator; setup and an invalid positive baseline remain
exit two, an accepted severing remains exit zero, and only the named rejection returns exit one.

The native retained-bundle helper executed all six populations and each returned exact exit one
with `fixture_rejected`. Cold focused tests passed for `deploymentrehearsal`, `operations` and the
rehearsal command, and root `make vet` passed. The direct root `go vet` attempt and a nonexistent
`make lint-go` target were invocation errors and are not cited as evidence. For discrimination,
the objective probe's mutation-result check was temporarily disabled; its sleeping-validator
subtest then failed with `sleeping objective gate satisfied negative`. Restoring the check returned
the complete typed-probe population green. Permanent injected sleeping validators cover the host,
alert, journal and objective gates independently.

This brings the implemented probe count to 21/43. These are typed negative fixtures, not runtime
claims: no clean-host installation, browser flow, database recovery, release/rollback, rotation,
positive alert delivery, journal measurement or objective observation is marked complete. The
retained candidate also predates this command surface and must be rebuilt again after DP-F2
construction stabilizes.

## 2026-08-23 — DP-F2 encrypted backup-envelope probes

Four backup negatives now run through the production `deploymentbackup` envelope rather than a
rehearsal-owned imitation. Every restore row first creates and successfully restores a valid age
X25519-encrypted envelope with matching payload bytes. The corruption row then tests both a
truncated file and a changed ciphertext byte; the identity and manifest rows supply independently
valid but wrong values. The interrupted-writer row delivers valid dump bytes before returning an
injected read failure, then requires both failure and an empty target directory so a partial or
apparently complete backup cannot survive.

The rebuilt native helper executed all four rows against the retained candidate/previous bundle
directories; each returned exact exit one with `fixture_rejected`. Cold `deploymentrehearsal` and
`deploymentbackup` tests and root `make vet` passed. For discrimination, the restore loop's
accepted-severing branch was temporarily disabled; the sleeping restore gate then failed with
`sleeping restore gate satisfied negative`. Restoring it returned the full focused population
green. A separately injected sleeping create gate permanently covers the interrupted-writer row.

The implemented probe count is now 25/43. This does not claim the real empty/populated Postgres
restore, non-clean target refusal, measured RPO/RTO, clean-host install or any other runtime row.

## 2026-08-23 — DP-F2 predeclaration: dependency-ordered execution plan

The exact-plan validator currently sorts all 42 non-forgery population names alphabetically. That
order is incompatible with the accepted runtime sequence: it places bounded drain before initial
install, rollback before its authorizing release, and interleaves checks from later evidence steps
into earlier ones. No truthful clean-host plan can obey it without hiding multiple lifecycle steps
behind opaque commands or manufacturing state outside the reviewed plan.

The next batch will replace name sorting with one explicit canonical population sequence grouped
under the eleven already accepted `RequiredSteps`. Every population remains present exactly once,
with its existing positive/negative kind and expected exit. Validation will require both the exact
name at every position and its owning step, so an operator cannot reorder, duplicate, omit or route
a check through a more convenient step. The order will be host preflight, install, browser, empty
restore, populated restore, incident recovery, candidate release, previous rollback, rotation,
operations, then provider-off supply chain. Within stateful release and rollback groups, refusal
fixtures precede the successful transition that would consume their authority.

Permanent tests will require the exact ordered closure and reject a swap, a correct name under the
wrong step and a duplicate. The discriminator will temporarily restore alphabetical comparison;
the canonical lifecycle fixture must then fail. This batch changes plan authority only and does
not implement or count any runtime population.

## 2026-08-23 — DP-F2 dependency-ordered execution contract

The plan contract now owns one explicit 42-row sequence and the exact step for every population.
It follows the accepted lifecycle rather than map/alphabetical order, keeps refusal fixtures ahead
of authority-consuming release/rollback successes, and rejects wrong-step routing as well as
reorder, duplicate and omission. The bound-run validator was also repaired to open result files by
their exact population names: directory enumeration is lexically sorted and therefore cannot be
used as a proxy for the reviewed execution order.

Cold rehearsal-package and command tests passed. The independent order oracle spells out all 42
names rather than deriving its expectation from the production sequence. Temporarily sorting the
production sequence by name made that oracle fail immediately, beginning with
`bounded_drain_and_restart` instead of `source_checkout_present`; restoring lifecycle order returned
the focused test green. Permanent negatives also swap rows, duplicate a row and put the correct
name under the wrong step. No runtime population was implemented or counted by this repair, so the
total remains 25/43.

## 2026-08-23 — DP-F2 predeclaration: strict runtime-scenario input

The remaining seventeen non-forgery rows require shared clean-host state and cannot be represented
honestly by the four probe paths alone. Before orchestration, the next batch will add one strict,
private scenario-input contract containing only non-secret identities and absolute paths: run ID,
candidate/previous bundles, work/artifact/operator/backup/metrics directories, age identity path,
public origin, receiver-health URL, public age recipient and server UUID. Secret values remain in
the production Compose secret files and never enter this document or the reviewed argv vectors.

The decoder will reject unknown/trailing fields, relative or non-clean paths, overlapping mutable
directories, malformed public/receiver URLs, malformed age recipient/server identity, non-0600 or
non-regular input and unsafe directory/file types. A config whose values are syntactically valid
but whose runtime paths do not exist will also reject; orchestration may not create authority by
guessing operator locations. Permanent mutations cover each field class and output overwrite is
not applicable because this batch only establishes input authority. The discriminator will remove
unknown-field refusal and require the strict-decoder test to fail. No population count changes.

## 2026-08-23 — DP-F2 strict runtime-scenario input contract

`deploymentrehearsal` now owns the strict private input boundary for the seventeen stateful rows.
It records only non-secret values and absolute paths, rejects unknown/trailing JSON, and validates
the ruled HTTPS public origin, receiver endpoint, age X25519 recipient and canonical server UUID.
All five mutable directories must be pairwise non-overlapping and outside both immutable bundles;
the identity file must also live outside mutable state. Load-time checks require every directory to
exist as a real directory and the identity to be a private nonempty regular file. The scenario JSON
itself must be a real exact mode-0600 file no larger than 64 KiB.

Cold rehearsal tests cover valid load plus relative, nested, colliding, credentialed/queried URL,
malformed recipient/server, missing directory, symlink and public-file negatives. Removing only
`DisallowUnknownFields` made the strict-input test fail with `unknown field accepted`; restoring it
returned the full package green. Root vet also passes. This is input authority for later runtime
orchestration and does not implement or count a population; the total remains 25/43.

## 2026-08-23 — DP-F2 predeclaration: observed-host and objective artifact bindings

Adversarial tracing found that raw command results and twelve retained artifacts are byte-bound,
but `Evidence.Host` and `Evidence.Objectives` are only shape-validated top-level claims. A dossier
can currently replace a real machine identity or recovery timestamps with any independently valid
values and re-encode itself without changing a retained artifact. That is inadmissible for the
clean-host, RPO and RTO claims.

The next batch adds two owning mode-0600 artifacts: a timed host observation and a timed recovery-
objective observation. Their strict decoders reject unknown/trailing fields, invalid intervals,
unsafe host/objective values and incomplete/guarded termination. Final/base validation will require
their hashes, open their bytes, bind their intervals inside the run and require exact equality with
the top-level host/objective values. The base artifact count becomes fifteen including the reviewed
plan; the final sealed count becomes seventeen. Schema, fixtures, docs and seal expectations change
together. Rehashing different valid host or objective bytes into the dossier must still reject.

The discriminator will temporarily remove one equality binding and require the permanent rehashed-
artifact negative to fail. This strengthens evidence only; no runtime population is counted.

## 2026-08-23 — DP-F2 host and recovery observation bindings

Two owning strict contracts now close the top-level assertion gap. `HostObservation` contains the
exact supported-host fields and a completed/non-guarded interval; `ObjectiveObservation` contains
the incident, newest valid backup, restore start, authenticated smoke, derived RPO/RTO and restored-
identity result. Both have strict decoders and exclusive mode-0600 writers. The retained artifact
directory now has fourteen files, the base dossier binds fifteen hashes including the reviewed
plan, and the final sealed dossier binds seventeen.

Bound-run validation opens both artifacts after hash comparison, requires their intervals inside
the run and requires their decoded values to equal the top-level evidence. Tests mutate each into
a different but independently valid observation, rehash both artifact and base dossier, and still
require rejection. Removing only the host equality comparison made the isolated base-run test fail
with `rehashed different host accepted`; restoring it returned both host and objective negatives
green. Cold rehearsal/command/release-package tests and root vet pass. This adds evidence ownership
but no observed runtime population; the count remains 25/43.

## 2026-08-23 — DP-F2 predeclaration: semantic operator-record authority

The retained `release_ledger` cannot honestly witness both AC1 and AC3–AC5. AC1 requires a clean
install of the exact candidate; the release/rollback lifecycle requires an existing previous
release, a forward previous→candidate transition and a candidate→previous rollback. The production
ledger correctly closes install authority after its first success, so those are two independent
operator histories. Combining them in one append-only file would be an impossible timeline.

The next batch adds a separate retained `install_ledger` and strict byte decoders for install/
release and rotation ledgers plus the backup header. Final validation will require the install
ledger to contain the exact candidate success, while the lifecycle ledger must contain the exact
previous install, candidate release and previous rollback in order, with manifest/version/image/
backup bindings and the seven-day authority. The retained pre-upgrade header must bind the previous
manifest and the release/rollback backup ID. The rotation ledger must contain governed activation
and removal for JWT, bootstrap and cursor with each minimum overlap met. Unknown, malformed,
impossible, mismatched and extra success histories reject; negative-fixture scratch ledgers remain
outside retained authority.

This raises the retained directory to fifteen files, the base dossier to sixteen hashes including
the reviewed plan and the final dossier to eighteen. The discriminator will neutralize one exact
manifest binding and require a rehashed, structurally valid wrong-ledger fixture to fail. No runtime
population is counted until these records are produced by the real clean-host commands.

## 2026-08-23 — DP-F2 semantic operator-record bindings

The retained authority is now split as predeclared. `install-ledger.jsonl` must contain exactly one
successful install of the candidate manifest/version/six images. `release-ledger.jsonl` must
contain exactly the previous install, forward candidate release and exact previous rollback, with
one operator, matching manifest/version/image identities, the same pre-upgrade backup and a full
seven-day authority. The strict decoded backup header must bind that backup to the previous
manifest/epoch and lie inside the run. The strict rotation decoder and semantic gate require the
JWT, bootstrap and cursor activation/removal pairs and their 30-minute, 31-day and 366-day minimums.

`deploymentrelease` now exposes byte decoders using the same append-only validation as file reads;
`deploymentbackup` exposes a strict header decoder and also rejects backward header intervals, a
gap found while making the artifact admissible. The candidate manifest artifact must now equal the
candidate bundle's actual manifest bytes. The retained directory has fifteen files, base evidence
sixteen hashes and sealed final evidence eighteen.

Cold rehearsal, backup and release tests cover strict/trailing data and rehashed wrong candidate
install, extra lifecycle success, wrong backup manifest and shortened rotation authority. Removing
only the install-manifest equality made the isolated base test fail with `rehashed wrong candidate
install manifest accepted`; restoring it returned the population green. Root vet and schema checks
pass. These are validator and authority contracts; the real records remain unobserved and the probe
count remains 25/43.

## 2026-08-23 — DP-F2 predeclaration: clean-host observation producer

The first runtime producer will be `deployment-rehearsal observe-host --config=<private path>`.
It will consume the strict scenario input, validate both exact bundles, observe Linux kernel/
architecture/distribution and Docker/Compose versions through direct commands, prove the named
Compose project has no containers or Postgres volume, reject source-control metadata in every
configured run/bundle directory and reject any optional identity/mail/AI/payment/cloud-provider
credential in its inherited environment. Production database/key/receiver secret-file plumbing is
not an optional-provider credential and its values are never read into evidence.

The producer writes only `host-observation.json` in the exclusive artifact directory, using the
already bound strict mode-0600 contract. Command errors, unknown OS/architecture, malformed distro
identity, nonempty project state, provider credentials, checkout metadata, output overwrite and
invalid/incomplete timing reject. Tests inject exact command output and independently sever each
predicate; the discriminator will make one provider credential invisible and require its permanent
negative to fail. This is producer construction, not an observed clean-host population.

## 2026-08-23 — DP-F2 clean-host observation producer

`deployment-rehearsal observe-host` now consumes the private scenario config and writes the bound
host artifact. It validates both exact bundles, observes `uname`, `/etc/os-release`, Docker Engine
and Compose directly, and rejects a non-Linux/non-amd64 host, malformed distro/version, a nonempty
Cloud Clicker container/volume population, checkout metadata in any configured run directory or a
nonempty optional identity/mail/AI/payment/cloud-provider environment variable. The root Make lane
exposes the exact command; no secret value enters argv or output.

Injected command tests prove valid Debian/Linux/amd64 output and independently reject dirty project
state, checkout metadata, provider credentials, wrong OS/architecture, malformed distribution,
invalid bundle and output overwrite. The provider-family table covers AWS, Azure, Google/GCP,
OpenAI, Anthropic, Stripe, SendGrid, Mailgun, Twilio, SMTP, Sentry, Datadog and New Relic. Removing
only the OpenAI entry made the provider detector fail with `provider credential OPENAI_API_KEY
accepted`; restoring it returned the cold package green. Root vet passes. The producer has not run
on the authorized clean Linux host, so no positive population is counted and the total stays 25/43.

## 2026-08-23 — DP-F2 predeclaration: exact candidate-install producer

The semantic-authority split requires the scenario input body to replace its singular
`operator_state` with distinct `install_operator_state` and `lifecycle_operator_state` directories
and to add a bounded non-secret operator ID. The same edit will reconcile docs and strict-input
tests; leaving the singular field live would contradict the accepted evidence model.

`deployment-rehearsal install-candidate --config=<private path>` will then validate the candidate
bundle and invoke the production `deploymentrelease.Controller.Install` with the production
`DockerRuntime`. Its durable state lives only in the install operator directory. After success, the
producer strictly decodes that ledger, requires the single candidate install identity and copies
its exact bytes exclusively to retained `install-ledger.jsonl`; a controller success with missing,
extra, wrong or malformed authority is failure. It will not tear down the stack because the browser,
database and recovery populations consume that exact installed candidate next.

Tests inject only the controller execution boundary while keeping real bundle/ledger validation and
exclusive output. They cover controller failure, missing/wrong/extra ledger, overwrite and sleeping
identity gate. The discriminator will neutralize the exact candidate-manifest check and require the
wrong-ledger negative to fail. This constructs but does not execute the clean-host install row.

## 2026-08-23 — DP-F2 exact candidate-install producer

The scenario body now has distinct install/lifecycle operator-state directories and a bounded
operator ID; its path-separation, filesystem and checkout scans cover both authorities. The new
`install-candidate` command validates the exact candidate and invokes the production install
controller/Docker runtime with the scenario's origin, receiver, backup, metrics, recipient and
server identity. After success it reopens the durable install ledger, strictly requires one exact
candidate install and exclusively retains those same bytes as `install-ledger.jsonl`. It deliberately
leaves the installed candidate running for the browser/database/recovery sequence.

Tests keep real manifest/ledger validation while injecting the controller boundary. They reject
controller failure, missing ledger, wrong manifest, an extra success and retained-output overwrite.
Temporarily replacing the exact install matcher with shape/image checks made the wrong-manifest test
fail with `invalid candidate install authority accepted`; restoring it returned the cold package
green. Root vet passes. The Make lane exposes the command, but it has not run on Linux; the positive
install row and overall 25/43 count remain unchanged.

## 2026-08-23 — DP-F2 predeclaration: exact product-browser producer

The browser row will execute through `deployment-rehearsal run-browser --config=<private path>` so
the reviewed command cannot substitute an arbitrary Chromium image, driver or output. The producer
will validate the candidate, derive its exact rehearsal-only Playwright image and manifest digest,
then run the candidate's own `deployment-browser` read-only inside that digest-pinned image on the
Linux host network. It will drop capabilities, prohibit privilege escalation, provide only tmpfs
scratch space and mount only the driver plus exclusive artifact directory. No provider or product
secret enters the container.

Success requires the real driver to create `browser-result.json`, after which the producer reopens
it through the strict browser decoder and binds its candidate manifest. Missing output, an already
existing output, wrong image/driver/manifest, Docker failure or incomplete browser result rejects.
Tests inject the Docker command boundary, assert the exact security and identity argv, and write
only a typed result. The discriminator will remove the manifest comparison and require a wrong-
manifest result to fail. This constructs but does not execute the product browser population.

## 2026-08-23 — DP-F2 exact product-browser producer

`deployment-rehearsal run-browser` now validates the candidate, derives its exact manifest and
Playwright identities and runs the candidate's own browser driver inside the digest-pinned image.
The direct Docker argv fixes linux/amd64, host networking, read-only root, tmpfs/shm bounds,
no-new-privileges, dropped capabilities and only the driver/evidence mounts. Loopback HTTP is never
enabled. A zero Docker exit is insufficient: the producer strictly decodes the exclusive browser
artifact and requires its manifest to equal the candidate bytes.

Injected Docker tests assert every security/identity argument and reject Docker failure, missing
artifact, wrong result manifest, invalid candidate and overwrite. Removing only the final manifest
comparison made the wrong-manifest test fail with `invalid product browser result accepted`;
restoring it returned the cold rehearsal/command/browser packages green. Root vet passes. The Make
lane is ready, but the product workflow has not run on the clean Linux host and 25/43 is unchanged.

## 2026-08-23 — DP-F2 predeclaration: canonical recovery identity

The database recovery populations need a semantic identity boundary before orchestration. Counts
alone would let a restore silently exchange one player, Founder, Company, event, board or epoch row
for another while still passing. `deploymentbackup` will therefore expose one strict recovery-
identity observation whose six domains each carry both a row count and a SHA-256 digest over
canonical, deterministically ordered Postgres row bytes. The player domain includes accounts,
email and account-to-Founder ownership; Founder and Company cover their complete stream/revision
histories; events include intent and event records; board covers projection claims and verified
runs; epoch covers catalog bytes, epoch/hash authority and run epoch pins. Secrets and row content
never leave the helper: only counts and hashes are emitted.

The bundled `deployment-backup recovery-identity` command will read the existing file-backed
database URL and emit the strict observation. Empty recovery means zero player/Founder/Company/
event/board rows while retaining nonempty current epoch/catalog authority; populated recovery must
have every domain nonempty. The later rehearsal producer must compare the complete observation,
not independently editable booleans or counts, before and after restore.

Exact implementation paths are `server/deploymentbackup/identity.go`, its unit and real-Postgres
integration tests, `server/cmd/deployment-backup/main.go` plus command tests, and the canonical
backup/rehearsal sections of `docs/deployment.md`. Cold package tests and the declared Postgres
population must pass. The discriminator will remove one canonical row field from the digest input
and require a same-count content mutation to stop changing that domain's identity. This constructs
the evidence boundary only; it does not count an R-006 runtime population and leaves 25/43 honest.

## 2026-08-23 — DP-F2 canonical recovery identity

`deployment-backup recovery-identity` now streams a complete public-table fingerprint plus separate
player, Founder, Company, event, board and epoch/catalog identities from the file-backed private
Postgres connection. Each full JSONB row and table identity is length-delimited into SHA-256; only
hashes and counts leave the backup service. Strict validators distinguish an epoch-bearing empty
player population from a populated one in which all six semantic domains have rows, and the shared
comparison gate requires every field and the complete database identity to match.

The real Postgres 16 lane now restores both populations and compares the structured identity before
and after. Its populated fixture includes a Founder stream, Company stream, event and verified board
row rather than allowing an empty semantic domain to pass. A permanent same-count mutation changes
an account recovery hash and requires both the player and database digests to change. As the
declared discrimination probe, narrowing the production account query to hash only `account_id`
made that integration test fail with `same-count content mutation escaped identity`; restoring the
full-row query returned all three Postgres populations green. Focused cold Go tests, root vet and
`git diff --check` pass. (`make lint-go` is not a repository target; the canonical `make vet` lane
was used.) This is still construction, not a Linux R-006 observation, so the count remains 25/43.

## 2026-08-23 — DP-F2 predeclaration: exact recovery runtime adapter

The clean-host orchestrator must not recreate Compose argv or trust a host-routable database. The
production `deploymentrelease.DockerRuntime` will therefore gain three explicit recovery methods:
create a non-pre-upgrade encrypted backup, observe the strict semantic identity through the
candidate's private backup service, and restore that non-pre-upgrade backup into an already-clean
target. Existing release backup and rollback restore methods will share the implementation but
retain their mandatory `pre_upgrade=true` contract; a scheduled recovery backup cannot become
rollback authority and a rollback backup cannot masquerade as the scheduled recovery population.

The adapter must invoke only the exact bundle Compose file, pinned backup binary and existing
file-backed secrets. It strictly decodes status, path, header and identity output, binds manifest,
epoch, server and pre-upgrade state, requires the host-side encrypted file, and rejects unknown or
trailing output. Exact paths are `server/deploymentrelease/docker.go`, focused Docker-runtime tests,
`docs/deployment.md` and this log. Cold release/backup tests and root vet must pass. The discriminator
will remove the scheduled/pre-upgrade separation and require the wrong-header fixture to fail. No
runtime population is claimed; the count remains 25/43.

## 2026-08-23 — DP-F2 exact recovery runtime adapter

`DockerRuntime` now provides scheduled recovery backup, private semantic identity inspection and
scheduled recovery restore using only the exact bundle's Compose file and backup helper. Shared
creation/restore implementations bind the decoded output to the candidate manifest, epoch, server,
host encrypted file and the caller's exact pre-upgrade class. Release and rollback retain the
pre-upgrade-only methods; rehearsal recovery cannot consume or produce that authority.

Focused command fixtures assert the private database-secret mount, strict/trailing-output rejection,
manifest migration binding, both valid backup classes and cross-class refusal. Replacing the
creation gate's expected class with the returned header's own value made
`TestDockerRuntimeBindsBackupAndRestoreOutputToExactManifest` fail with `pre-upgrade backup accepted
as scheduled recovery backup`; restoring the caller-bound value returned the cold release/backup
packages green. Root vet and `git diff --check` pass. No real host state was observed, so 25/43
remains the runtime count.

## 2026-08-23 — DP-F2 predeclaration: destructive empty/populated recovery producer

`deployment-rehearsal recover-empty --config=<private path>` will consume the installed candidate
after the browser population, require a genuinely populated semantic identity, create its encrypted
scheduled backup, declare the incident, stop the stack, remove the Postgres volume, start only the
candidate gameserver/Caddy core to migrate and seed a genuinely empty player database, and require
the strict empty identity. It then backs up that empty population, destroys the volume again,
restores into the clean target, restarts the core and requires complete identity equality. The core
start deliberately excludes the continuously scheduled backup worker so the observed backup cannot
race an untracked second writer.

`deployment-rehearsal recover-populated` will consume only that strict private checkpoint, destroy
the empty-restored volume, restore the exact populated backup, restart and verify manifest/epoch/
artifact and semantic identity, then complete the authenticated Caddy smoke. It derives RPO from
the populated backup completion to the declared incident and RTO from restore start to smoke pass,
uses `deploymentbackup.MeasureObjectives`, and exclusively writes the retained typed objective
observation. The post-restore smoke may create new session/player state only after before/after
identity equality has been established.

Exact paths are a new `server/deploymentrehearsal/recovery.go` and tests, the rehearsal CLI and
Make lanes, a production `DockerRuntime.StartRecoveryCore` adapter plus focused test, canonical
docs and this log. Tests inject state transitions but retain strict bundle/header/checkpoint/
observation validation. Failure, wrong population, wrong backup class/manifest/path, incomplete
checkpoint, identity mismatch, objective overrun and output overwrite all reject. The discriminator
will neutralize the populated identity comparison and require its mismatch fixture to fail. This
constructs the producer only; no local/macOS execution counts toward the Linux R-006 25/43 state.

## 2026-08-23 — DP-F2 destructive empty/populated recovery producer

The two manifest-bound recovery commands now exist. `recover-empty` refuses to begin destruction
unless every populated semantic domain has rows, captures the exact scheduled backup, records the
incident, rebuilds Postgres from a clean volume using the candidate's core only, requires an
epoch-bearing empty identity, and proves the empty backup by a second destroy/restore/full-identity
comparison. Its mode-0600 checkpoint is written only after every step succeeds. `recover-populated`
then destroys that restored-empty target, restores the exact populated backup, compares the full
database identity before smoke can add state, performs the authenticated Caddy smoke and derives
the retained RPO/RTO observation through the production objective validator.

The production core-start adapter explicitly starts only gameserver and Caddy with `--no-deps`;
the scheduled backup worker and operations containers cannot race the observed backup. Strict
checkpoint validation reopens regular backup files, exact-decodes both headers, binds manifest,
epoch/server/class/path and rejects unknown/trailing state. Unit populations cover the complete
empty/populated sequence, every runtime failure boundary, wrong starting population, wrong backup
class/path, trailing checkpoint, identity mismatch, failed smoke, RTO overrun and pre-destructive
objective overwrite refusal. Neutralizing only the populated comparison by comparing the restored
identity with itself made the mismatch population fail with `mismatched populated identity
accepted`; restoring the before/after comparison returned all focused packages and root vet green.

This construction intentionally exposes a real next gate: the current minimal browser producer
does not create a verified board row, so `recover-empty` will refuse that state as an incomplete
populated recovery fixture. The browser/product population must be strengthened; the recovery gate
will not be loosened. No Linux run occurred and the honest count remains 25/43.

## 2026-08-23 — DP-F2 recovery boundary full-lane verification

After `7bb65d1`, the complete cold `make test-deployment-rehearsal` lane passed: rehearsal,
rehearsal CLI, release runtime and release CLI packages were green at `-count=1`, and both retained
independent-build records still strict-validated at their recorded previous/candidate manifest
identities. This verifies tooling compatibility only; the records predate the new command surface
and remain inputs to be rebuilt after DP-F command construction stabilizes, not evidence that the
new recovery producer ran.

## 2026-08-24 — DP-E/DP8 corrective predeclaration after hosted CI failure

GitHub Actions run `32637134690` failed at pushed HEAD `7b510df` in two blocking jobs. Both exact
leaf commands reproduce locally and therefore expose a local verification/process failure, not a
host-only divergence:

- `make verify-client` rejects pushed commit `09d5027` because its edit to the kernel-watched
  `server/production/intents.go` did not carry a real version bump or append-only correction;
- `make test-game-ui-composed` reaches HTTP readiness but every Centrifuge WebSocket upgrade fails.
  The DP-E operations middleware replaces the response writer with `statusWriter`, which does not
  preserve `http.Hijacker`; the real browser then reports repeated proxy `EPIPE`/`ECONNRESET` and
  times out before the visitor counter.

The repository-wide evidence language was also false: `docs/ci.md` called `make verify` the exact
blocking aggregate, but that target omits `test-game-ui-composed`; the DP-F rehearsal lane includes
neither failed leaf. No single local command was mechanically bound to all six push jobs.

**Review by:** Codex. **Recorded by:** Codex. **Decision:** **CHANGES REQUIRED** for the exact
offending DP-E implementation commit (`09d5027^..09d5027`). This is the implementer's first filter,
not the designated cross-party pass. History is already pushed and remains append-only.

The corrective range is predeclared as:

1. forward `http.Hijacker` through the operations status writer and make an operations-enabled real
   composition WebSocket population permanent; severing only the forwarding method must fail the
   handshake;
2. add `09d5027` to `kernel/history-corrections.json` in the same commit as kernel `0.3.101` across
   all three parity files; removal of the correction or any parity byte must fail the history gate;
3. add `make verify-push` with exactly the six blocking leaf commands and extend the topology guard
   plus its negative fixtures to bind workflow jobs and aggregate dependencies bidirectionally;
4. correct `docs/ci.md` to distinguish the developer aggregate from the exact push aggregate; and
5. run cold focused Go/Postgres tests, both previously red leaf commands, topology negative
   controls and the complete `make verify-push` before handoff. No timeout, hosted workflow,
   gameplay, balance, migration, content or deployment-secret behavior enters this range.

## 2026-08-24 — DP-E/DP8 corrective implementation and first filter

The predeclared correction landed in three concern-separated commits after predeclaration
`e0e2201`: realtime composition `73bb4ee`, append-only kernel correction `32ee744`, and CI parity
gate `5d72591`.

The realtime correction preserves `http.Hijacker` through operations instrumentation and enables
that middleware in the existing real Postgres/account-revocation WebSocket population. The exact
population failed its Centrifuge handshake with HTTP 500 before forwarding and passed cold after
it. Repeated composed execution then exposed two independent client ordering defects instead of
being retried away: an exit-offer decline could release the next click before binding its consumed
revision, and a successor snapshot could advance the cursor past the preceding run's terminal
publication. Permanent browser/runtime witnesses failed before each correction. Severing the
decline refresh made all three browser engines report zero snapshot calls where one was required;
the terminal race fixture received only `transport_recovered` until the bounded immediate-successor
delivery rule was restored. Older snapshot responses now also cannot regress either live revision
coordinate. Three consecutive real Chromium/Vite/gameserver/Postgres/WebSocket composed runs then
passed both terminal variants, next-run continuation and recovery.

Kernel `0.3.101` records pushed offending commit
`09d5027a8f57ddacde270a7ac3fc31e2a183f7d3` without rewriting published history. With that one
correction removed, `make verify-kernel-version` failed against `server/production/intents.go`;
restored, the parity/history gate and its adversarial fixtures passed.

`make verify-push` now contains exactly the six blocking workflow leaves in job order. The topology
guard binds each hosted job to its exact Make command and binds the local aggregate back to the same
set. Thirteen negative controls reject a substituted hosted leaf, missing/extra local leaf, missing
job, exhaustive push work, trigger drift, cache drift and maintenance-evidence weakening.

The first aggregate invocation was invalid evidence: the restricted execution sandbox denied two
legitimate `httptest` localhost listeners with `bind: operation not permitted`. The exact command
was rerun with local test-network authority and exited zero. Executed leaves were cold server core,
fast harness, strict client type/build/unit/boundary/history gates (6,662 unit tests), the complete
three-engine functional browser population (20,049 tests) plus isolated performance, the real
composed browser path, and schema/catalog validation. No workflow timeout, skipped job, hosted-only
exception or retry was changed.

**Review by:** Codex. **Recorded by:** Codex. **Decision:** **APPROVED AS FIRST FILTER** for
`e0e2201..5d72591`; this does not satisfy the designated cross-party gate. The corrective batch is
ready for Claude's adversarial review of that exact range plus this record commit. It is not
archival authority and does not promote DP-E or Deployment Foundation lifecycle state.

## 2026-09-23 — DP-F2 predeclaration: populated browser journey

The current `deployment-browser` stops after one manual intent. That proves only initial Caddy,
bootstrap, WebSocket and intent transport. `recover-empty` correctly refuses its database because
the required verified board domain is absent. The browser producer must not report Phase-0
completion or feed the recovery gate from this fragment.

**Authority and scope:** accepted DP-F/AC1/AC4 and the previously declared clean-host product
population. This slice may change only the rehearsal browser driver, its strict evidence schema,
the binding adapter fixtures, focused tests and canonical deployment documentation. It may not
add a product clock, test-only public endpoint, direct database seed, gameplay balance/content
change, looser recovery identity or a broad hosted run claim. The first manual intent and every
subsequent gameplay intent must originate from an enabled visible DOM control against the real
Caddy origin; read-only state inspection is permitted but is not a substitute for the UI action.

**Positive population:** fresh anonymous bootstrap → Desk and committed credentials → real
WebSocket → manual intent → earn/buy through the default T0 gate → cross gate → first Exit/run-end
→ continue into run 2. The result binds exact manifest, observed controls/surfaces, successful
intent responses, zero page exceptions, completed objective and a bounded elapsed/attempt count.
The recovery producer then observes a *real* populated identity, including its asynchronously
verified board row, before the encrypted backup. A guard expiry or missing projection fails and
cannot be labelled a successful browser or recovery observation.

**Falsifiers:** sever each gate, Exit and continuation DOM action; make an intent receipt reject;
remove WebSocket handshake; force an offer interruption; suppress verified-board projection; and
forge a result with absent transition/terminal fields. Each relevant witness must fail the
appropriate browser or recovery validator. Focused tests must run cold. The complete clean-host
population and timing objective remain unclaimed until DP-F3 executes on the exact bundle.

**Sequence:** first strengthen the browser journey and typed result with local discriminating
tests. In a separate bounded slice, make recovery wait for the real asynchronous board projection
with a fixed guard and test both eventual arrival and permanent absence. Only then rebuild the
candidate bundle and attempt the external rehearsal. No plan checkbox flips on this predeclaration.

## 2026-09-23 — DP-F2 browser-journey construction (unexecuted product population)

The browser driver now uses visible, enabled player controls to bootstrap, perform a manual
action, buy generators at a deliberately spaced cadence so buying cannot indefinitely starve the
gate, cross the T0 gate, reach run-end and continue to run 2. An offer interruption is declined
through its UI control. Read-only production snapshot calls verify run sequence 1 → 2; no direct
gameplay intent, fixture state, clock advance or database seed was introduced. Its strict result
is version 2 and refuses absent gate/terminal/next-run observations, failed intent HTTP responses,
page exceptions, invalid sequence, action/elapsed guard exhaustion and missing transport evidence.
The command's former ten-minute maximum could not contain the already designed first-hour
journey; it now has the same two-hour execution guard as the driver. This guard is not a pacing
target and cannot turn a timed-out run green.

Cold `make test-go GO_PACKAGES='./cmd/deployment-browser ./deploymentbrowser ./deploymentrehearsal'
GO_TEST_FLAGS='-count=1'` and focused `make vet` passed. The complete
`make test-deployment-rehearsal` also passed after its first attempt was invalidated by sandbox
denial of Go module-cache writes; that first attempt is not cited as a product failure or pass.
The strict result test runs negative cases for each new boundary; the policy table exercises
offer interruption, gate, purchase, manual, Exit and continuation selections. The actual
two-hour browser population, DOM-action severing probes and asynchronous verified-board arrival
have **not** been executed. The real clean-host browser row and 25/43 rehearsal population count
remain unchanged. This code is construction ready for product execution, not AC1/AC4 acceptance.

**Review by:** Codex. **Recorded by:** Codex. Exact construction range
`b2fde65..44f5b3f` was inspected after the cold runs. **APPROVED AS FIRST FILTER OF THE
CONSTRUCTION SLICE ONLY.** The validator's new forged-field negatives and the enabled-control
selection matrix discriminate locally. No running-product/browser-action severing, verified-board
arrival, clean-host result or cross-party designated review is imported by this verdict. The
candidate bundle predates these driver bytes and must be rebuilt before any R-006 claim.

## 2026-09-23 — DP-F2 predeclaration: asynchronous board-arrival boundary

The verifier runs every 100 ms and projects a verified run asynchronously. `recover-empty`
currently makes one immediate identity read, so even a legitimate completed browser run can race
the board row. This slice changes only the rehearsal recovery producer, its focused fake-runtime
tests and canonical deployment documentation. It does not alter verification, game state,
backup content, RPO/RTO, the populated-domain rule or the product browser.

Before the first encrypted backup, poll only the explicitly pending-board state: player, Founder,
Company, event and epoch domains must already be populated and structurally valid; only board
rows may still be zero. The safety guard is six minutes, covering the verifier's five-minute
claim lease plus five five-second retry backoffs and its 100 ms cadence, rounded upward only to
avoid a boundary race. This is an observation guard, not a release latency promise or permission
to increase an objective bound. An inspection error, any other missing domain, context cancellation
or guard expiry fails without stopping or resetting the database. After creating the populated
backup, re-inspect and require byte-equal semantic identity before any destructive operation;
a mutation during capture must not produce a checkpoint.

The positive fixture presents board-absent then board-present identities before backup, and the
same identity after capture. Negatives hold the board absent through a short injected guard,
remove a non-board domain, inject an inspection error and mutate an identity during backup. Each
must fail with no destructive runtime call and no checkpoint. Cold focused recovery tests, the
rehearsal aggregate and vet are required. This is local producer proof only; no clean-host board
arrival or AC4 claim follows from a fake runtime.

### Scope correction before implementation commit

Inspection of `deploymentbackup.RecoveryIdentity` found that `Board.Rows` aggregates both
`verification_projection_events` and `verified_runs`. A projection-event-only state can therefore
pass a mere nonempty-board check without the required verified leaderboard row. The board-arrival
slice must also update the canonical recovery identity producer/validator and its private Docker
adapter fixtures: add an explicit `verified_runs` count, bump that rehearsal identity schema, and
require at least one actual verified run for populated recovery. Empty recovery requires zero.
The full board hash still protects content equality. A projection-event-only negative is mandatory.
This is DP6/AC4 evidence-contract tightening, not a gameplay or board-projector change; no
previously accepted migration, backup body or product API is edited.

## 2026-09-23 — DP-F2 board-arrival and exact verified-row construction

`recover-empty` now waits only when player, Founder, Company, event and epoch identities are
already populated but the verifier's actual `verified_runs` row is pending. The six-minute
safety guard and 100 ms poll fail closed on expiry, cancellation, malformed identity or runtime
error. The producer re-inspects after the populated backup and refuses a changed identity before
any stop/reset/restore. The recovery identity and private checkpoint are version 2; the former
reports `verified_run_rows` separately from the complete board-domain content hash. A projection
event alone no longer satisfies populated recovery.

Executed cold evidence: focused `make test-go` for deploymentbackup, deploymentrehearsal and
deploymentrelease; `make test-deployment-rehearsal`; focused `make vet`; and
`make test-deployment-backup` against real Postgres 16 all passed. The Postgres negative first
attempt tried to delete an immutable verified board row and correctly failed at the database
history trigger; the corrected fixture instead starts with a projection event and never inserts a
verified run, and the inspector returns board rows > 0 with verified rows = 0, rejected by the
populated contract. Fake-runtime negatives prove permanent absence and mid-backup mutation leave
no checkpoint and never enter the destructive phase. The earlier red assertion about the number
of legitimate stop calls was corrected before the cold green rerun; it was test expectation drift,
not a production failure.

This construction does not prove that the new browser driver actually reaches a verified run on
the clean host. It does not add an executed R-006 population or change the 25/43 count. The exact
candidate bundle still predates these bytes and must be rebuilt; AC4 and Deployment archival
remain open.

**Review by:** Codex. **Recorded by:** Codex. Exact construction range
`30cd7b9..6d89880` was inspected with the cold focused, aggregate and real-Postgres outputs
above. **APPROVED AS FIRST FILTER OF THIS AC4 CONSTRUCTION SLICE ONLY.** The real
projection-event-only database is a non-vacuous negative for the new verified-row count; the
fake runtime fails on permanent absence and a changed post-backup identity before destructive
calls. This is not the designated cross-party verdict, a clean-host run, or archival authority.

## 2026-09-23 — DP-F2 supported-architecture binary preflight

From clean local source commit `85f8e9f`, the root Make targets built the changed
`deployment-browser`, `deployment-backup` and `deployment-rehearsal` commands as static
Linux/amd64 ELF binaries. Their ignored local outputs have SHA-256 digests respectively
`ee8ce4f2cf4034262c07d0bbef2ab6b348fb72af2c79864e27e84c7505a45fa1`,
`569941aca0908a20a5576bea11dfaa91c2e1c74ee4f62af314eb1ffe840c9454`, and
`b37be504bd30548b9c9e0e92fd890a256a294e29ac172b2ec45d43d31986e563`.
`file` confirmed each is an x86-64 statically linked executable. This checks the supported
architecture compilation, not an independently rebuilt six-image release bundle, browser run,
secret scan or clean-host result. No retained build ledger is updated yet.

## 2026-09-23 — DP-F2 predeclaration: refreshed exact candidate bundle

The retained `candidate.json` and local candidate bundle predate the Phase-0 browser journey and
version-2 verified-row recovery identity. They cannot be used to run or claim R-006. The next
local build will use a clean committed source coordinate after this entry, with a new
`0.1.0-preview.2` *candidate label only* (not an owner release call), the existing pinned
BuildKit/Syft tools, Docker Engine 28.4.0 and Compose 2.39.4-desktop.1, and the six plus one
previously selected immutable Linux/amd64 image references/config digests. The prior `previous`
bundle remains the rollback input; it is not silently rebuilt or relabelled.

Build all six static Linux/amd64 commands and the client from that coordinate, stage the exact
content closure, build the gameserver image/archive, generate all seven real normalized image
SBOMs and assemble a new candidate directory. Repeat in an independent empty output tree with
the same declared source timestamp. Require byte-identical bundle trees, image archive and
normalized SBOMs; record the source/tool/image/output digests and a real source/image secret scan.
Any mismatch, mutable image identity, missing input, network/build failure or dirty source
invalidates the candidate and leaves old retained records untouched. This is DP-F2 construction,
not a clean-host run, supported-self-host claim, status promotion or authorization to push.

## 2026-09-23 — DP-F2/AC5 predeclaration: previous-bundle schema compatibility

The refreshed candidate assembled as `0.1.0-preview.2` with 72 artifacts and validates locally.
The first exact candidate+previous positive probe refused the retained previous bundle even
though its 70 manifest artifact hashes, Compose/Caddy/Dockerfile, application/image SBOMs,
operations profile and gameserver archive independently pass. The remaining mismatch is that
current `validateSchema` unconditionally requires the `rehearsal_images` property in the bundled
release-manifest schema, while the exact previously retained bundle predates the browser lane
and has neither that optional property nor a browser artifact/image. AC5 explicitly needs the
previous bundle to validate for rollback; loosening the new candidate's browser closure would
be the wrong fix.

This narrowly scoped correction may change the release-manifest schema validator and its tests:
accept the legacy schema shape only when the loaded manifest has no rehearsal image and no
browser artifact/SBOM; retain the current property requirement for any browser-bearing bundle
and for the canonical repository schema. The positive population is a generated legacy-shaped
bundle plus the exact retained previous bundle. The negatives remove the property from a
browser-bearing candidate and alter other required schema fields; both must still fail.
No previous bundle byte, migration, production gameplay behavior, image digest or R-006 evidence
is changed. After the fix, rerun cold releasepackage tests and the exact candidate+previous
supply-chain probe; then continue independent candidate rebuilding and scans.

The narrow validator change is implemented with self-contained generated bundle tests. The
legacy-shaped positive removes the browser helper, Playwright SBOM, rehearsal image and optional
schema declaration together; a second test keeps the browser closure while removing only the
schema declaration. A third mutation removes another required schema field from the historical
fixture. `make test-go GO_PACKAGES='./releasepackage ./deploymentrehearsal ./deploymentrelease'
GO_TEST_FLAGS='-count=1'`, `make vet`, and the exact retained candidate-v3 plus previous-bundle
`six_image_sbom_license_provenance` probe passed. The probe previously exited 2 on the unchanged
previous bundle, so this is a demonstrated refusal-to-accept correction, not a merely green
fixture. Candidate-v3 was assembled before the fix and contains an obsolete rehearsal binary;
it is diagnostic only and cannot become R-006 evidence. Rebuild from the corrected committed
source before independent comparison, build records or a clean-host run.

## 2026-09-23 — DP-F2 predeclaration: corrected candidate rebuild

The compatibility correction landed at `578fda0`. Candidate-v3 remains an immutable diagnostic
comparison input, not a candidate to promote. From the next clean committed source coordinate,
build two independent `0.1.0-preview.2` candidate trees in empty ignored output directories
with identical pinned tools, seven immutable image identities and source timestamp. Require
byte-identical binaries, client, image archive, normalized SBOMs and complete bundle trees;
validate both with the corrected helper and run the exact candidate-plus-previous supply-chain
probe. The first copy is only a release candidate if all checks pass and a retained source/image
secret scan and release-build record validate. Neither copy is R-006 clean-host evidence. Any
new source edit invalidates this coordinate and requires another rebuild.

The clean build coordinate was `8621f33f2c2f6d419691a59a2c17ea6249d934fe`, timestamp
`2026-09-23T12:50:59Z`, with Docker Engine 28.4.0, Compose 2.39.4-desktop.1 and the pinned
BuildKit/Syft images in the Makefile. Two fresh ignored trees, `candidate-v4a` and `candidate-v4b`,
each built six static Linux/amd64 binaries, client output, staged content, a no-cache gameserver
archive, and seven separate real Syft scans. The first image-context attempt failed because content
was copied beside, not under, `content/`; corrected fresh contexts succeeded without changing
source. The two archives are byte-identical at
`sha256:3ce0a9d45d5b333817d7b99a778179f3d195809ac79e5265e0be94bc06079205`;
their runtime config ID is
`sha256:34d48155b28d09fc929a0891ce05a09b2a33d8d77fbbb559bb7e4c40c4218ddd`.
All six binaries, client distributions, metadata, content, seven *normalized* SPDX documents and
both 72-artifact bundle trees compare byte-identical. Raw Syft headers differed and are not the
release evidence. Both manifests hash to
`sha256:521e5bcaf4ff32bec47ec79eec7e44e1794c87513c5aefa1db8b04a599334d36`.

The exact candidate-v4a and candidate-v4b versus retained previous-bundle supply-chain probes
each passed. A first structured secret-scan attempt exited 1 because the Make wrapper forwarded
a relative output path to an absolute-path-only result writer; no secret finding occurred. The
rerun with an explicit absolute output scanned 1,407 tracked files and the exact gameserver
archive with zero findings; RP-111 tracks the tooling defect. The candidate build record now
names this exact source/manifest/archive/SBOM closure; `make test-deployment-rehearsal` validated
both candidate and previous records after cold tests. The full supply-chain command reopened the
two bundles, records and structured scan and passed, writing an ignored result. These are local
DP-F2 construction and supply-chain facts, **not** a clean-host install, browser journey,
restore/rollback, R-006 completion, supported-self-host claim, designated review or owner release.

**First-filter review — Review by: Codex; Recorded by: Codex.** Reviewed implementation range
`86dfdbc..578fda0` (rollback schema compatibility, self-contained positive/negative bundle
fixtures and canonical doc). The previous-bundle rejection was reproduced before the fix; the
exact pair probe passes after it. The historical fixture rejects a removed required field and
the browser-bearing fixture rejects the missing new declaration. Cold focused tests, vet and
diff check passed; no product gameplay or historical bundle byte changed. RP-111 is a separate
Make-wrapper defect discovered while generating the build evidence. This is an implementer-side
first filter only; **Claude's designated cross-party verdict for the exact range is absent** and
DP-F cannot archive or claim release acceptance from this entry.

## 2026-09-23 — DP-F2/AC5 follow-up predeclaration: versioned rollback and scan output

RP-112 is a latent seven-day rollback break: `ValidateReleaseManifest` currently requires the
previous bundle's company/founder save versions to equal the *new validator binary's*
`save.Latest…Version` constants. The exact previous bundle happens to pass now because both
versions still match. At the next save-version bump, a byte-exact previous bundle would be
rejected before the mandated backup restore. The accepted DP5/AC5 contract is to validate that
bundle and prove rollback from its pre-upgrade backup, not to pretend its historical version was
rewritten. RP-111 independently makes the root secret-scan target fail for a relative result path.

This follow-up may change only manifest version validation, the root Make secret-scan output-path
forwarding, focused fixtures/tests and canonical deployment docs. A generated historical-shaped
bundle with lower positive company/founder versions must validate, while zero or future versions
must refuse; the builder must still emit both current latest versions. The exact retained
candidate/previous pair must still validate. For the scan wrapper, a repository-relative ignored
output must succeed and an existing output must fail without overwrite; an absolute path still
works. The old failing relative-path run is the pre-change discriminator. No gameplay migration,
schema, historic bundle byte, release image or owner rollback interval changes. After source
changes, candidate-v4 remains evidence for its pinned source commit but is **not** the current
candidate; rebuild two independent exact trees before any R-006 clean-host run.

The local correction keeps `BuildReleaseManifest` current-version-only and changes structural
validation to admit positive versions no newer than the running validator's known versions.
The generated-bundle test covers both lower previous versions and zero/future refusals. Reverting
the validator to its old equality rule made the named previous-version subtest fail; restoring
the rule made it pass. Cold focused releasepackage, rehearsal, release and scan-command tests and
`make vet` passed. The root Make target now resolves `SECRET_SCAN_OUTPUT` against the repository
root: a relative ignored output succeeded, and the exact second invocation failed on exclusive
create while the result SHA-256 remained unchanged. The pre-change relative invocation failed
with `invalid release runtime content`, so this is a demonstrated path correction. The retained
previous and candidate-v4 pair still need revalidation after the commit; candidate-v4 remains
bound to its older source and cannot stand in for the rebuilt release candidate.

## 2026-09-23 — DP-F2 predeclaration: final-source candidate refresh

The rollback-version and Make wrapper correction landed at `b7a9587`. The `candidate-v4` pair
must not be relabelled to this new source. Build fresh `candidate-v5a` and `candidate-v5b` trees
from the next clean commit, with the same pinned Docker/Compose/BuildKit/Syft and seven immutable
image inputs, version label `0.1.0-preview.2`, and that commit's timestamp. Rebuild six Linux/amd64
commands, the client, staged content, no-cache gameserver archives, seven independent real image
SBOM scans and both bundles. Demand byte equality of normalized outputs and full bundle trees;
then rerun both candidate/previous probes, a manifest-bound structured source/image secret scan,
the retained build-record validator and full supply-chain check. Update the candidate build record
and strategic trackers only to the proven v5 manifest. No current-head source edit may be treated
as covered by the old bundle. Clean-host R-006 and designated review remain separate gates.

The clean source coordinate was `7e8aa708ff9e002e7ec4b46242d7d889f76b7fdd` at
`2026-09-23T13:07:52Z`. Both fresh trees built six static Linux/amd64 commands, client output,
staged content, no-cache gameserver archives and seven independent real Syft scans. Their binaries,
client, metadata, content, archives, normalized SPDX documents and complete 72-artifact bundles
compare byte-identical. The gameserver archives hash to
`sha256:660d95778fe9d015a376d97e0024614fc0ed1ede3c4154c8c9d319`, config ID
`sha256:37ba4070b3d2316a600781dd314513437ee58e937d30869fd6e0c136318888d7`;
both manifests hash to
`sha256:677e94bbc51af65e394599ac3951120d324579ac3f129c0dc52bf0ac5247818e`.
Both exact candidate/previous probes passed. The structured scan through the corrected *relative*
Make output path covered 1,407 Git-tracked files and the exact image archive with zero findings.
The candidate build record binds these exact hashes; `make test-deployment-rehearsal` passed cold
and validated both candidate and previous records. The full supply-chain command reopened both
bundles, both records and the scan result and passed. Raw Syft headers are not compared or shipped;
the seven normalized documents are byte-equal. These are local DP-F2 facts, not R-006 runtime
evidence, designated review, an archival gate or an owner release call.

**First-filter review — Review by: Codex; Recorded by: Codex.** Reviewed implementation range
`df8f6d7..b7a9587`: the versioned manifest gate, historical/current/invalid fixture, root
secret-scan path correction, canonical doc and backlog disposition. The old equality-rule mutant
failed the prior-version positive; the corrected rule passed it and rejected zero/future values.
The old relative output failed; the corrected path succeeded and the existing-output negative
failed without changing its digest. Cold focused tests, vet, exact prior pair probe and diff check
passed. This is not Claude's mandatory designated cross-party verdict. The later planning/build
record range `b7a9587..` also remains unreviewed by the designated party.

## 2026-09-23 — bounded designated-review handoff prepared

Prepared `designated-review-handoff.md` for the five DP-A–DP-E batches plus the separate DP-E/DP8
corrective range. Git history confirms the six `base..tip` ranges are contiguous within each batch
and disjoint across the DP-F construction interval: commit counts are 3, 9, 4, 3, 3 and 5. The
handoff points to the accepted RFC, this append-only log, cold root Make lanes and the required
independent severing checks. One cross-party session can review them efficiently, but each batch
still needs its own exact-range verdict. No designated review, DP-F approval, R-006 host result,
archival or release claim is inferred from preparing the packet.

## 2026-09-24 — Claude designated cross-party review: DP-A, DP-B, DP-C, DP-D

**Review by:** Claude — the designated cross-party pass required by AGENTS.md (c). Each range was
inspected by a separate Claude review subagent working in an isolated detached worktree at the
range tip, and the parent Claude session re-checked the load-bearing findings in source.
**Recorded by:** Claude (parent session). Codex first-filter verdicts were read as input only.
**Environment:** macOS 27 arm64 host, go1.27.1, Docker 28.4.0 / Compose v2.39.4. Linux/amd64
containers ran under emulation. This is not a clean supported host, and no R-006 claim follows
from it. Every worktree mutation was restored with `git checkout -- .` and a clean `git status`.
The worktrees and the Docker projects `cloud-clicker-backup-test` and `cloud-clicker-release-test`
were removed afterwards.

### DP-A — exact range `cd102d7..d7d443f` (bbff0b6, a906398, d7d443f) — **CHANGES REQUIRED**

Executed cold at `d7d443f`:
- `make test-go GO_PACKAGES='./deploymentconfig ./gameserver ./cmd/gameserver ./account ./transport ./publicapi' GO_TEST_FLAGS='-count=1 -v'`
  exited 0: 188 PASS, 10 SKIP, 0 FAIL. Every skip is a pre-existing Postgres `*Integration` test
  or the first-hour replay. All seven load-bearing config tests ran, including all 43 fail-closed rows.
- `make vet` exited 0 and `gofmt -l` reported nothing.
- The real built binary, run under `env -i`, exited 1 before any database attempt for a missing
  secret file, an `http://` origin and the legacy `DATABASE_URL`. No secret value appeared in its output.

Severing probes:
- These failed their witnesses as required:
  - proxy depth;
  - the insecure-origin scheme;
  - the half-pair check (bootstrap/cursor);
  - duplicate previous IDs;
  - duplicate previous values;
  - the legacy-variable list.
- These left every test green:
  - the `/run/secrets/` prefix check;
  - path normalization;
  - origin userinfo;
  - the trailing-dot origin check;
  - `gameserver/composition.go` `apiConfig.TrustedProxyHops = trustedProxyHops`;
  - `policy.AllowedOrigins = allowedOrigins`.
- As a carry-over, the same two composition mutations were run in DP-D's real Caddy lane at
  `3f58fea`. The origin severing failed there (WebSocket 403). The proxy-hop severing still passed.

- **B1 (blocking):** the production proxy-hop wiring has no witness that can fail. The log and
  `docs/gameserver.md` say that composition binds one trusted hop into account/IP handling. Nothing
  in this range or in DP-D's lane fails when that assignment is removed, so production could
  silently treat every player as Caddy's single IP.
- **Non-blocking:**
  - The JWT half-pair row passes through the empty-path guard, not the half-pair branch.
  - The four surviving decoder rules above lack rows.
  - Duplicate environment keys are unreachable through `os.Environ()`: Go keeps only the first.
    DP3's duplicate-key rule must therefore be enforced at the `.env`/schema preflight, and the
    `duplicate_environment_key` row does not prove it for `LoadEnvironment`.
  - The gameserver's closed `CLOUD_CLICKER_*` allowlist rejects the other services' variables if
    the env file is shared (source reading).
  - Fail-before-migration ordering has no automated test.

### DP-B — exact range `d7d443f..9b916c2` (f62e0b7…9b916c2, 9 commits) — **CHANGES REQUIRED**

Executed cold at `9b916c2`:
- `make test-go` over releasepackage and the new cmd/config packages exited 0; releasepackage
  passed 19/19 with none skipped.
- The five new cmd packages have no tests.
- These all exited 0: `make stage-release-content` (31 files), `make release-secret-scan`,
  `make render-release-compose`, `make generate-release-metadata`,
  `make build-gameserver-linux-amd64` and a client `vite build`.
- Two `--no-cache` `make build-gameserver-image` runs produced byte-identical archives, so the
  reproducibility claim holds.
- Not run: `docker compose up` from an extracted bundle, and `make deployment-config-check`
  (it needs secrets).

Severing probes:
- These failed as required:
  - removing a closure file;
  - removing the image source-revision check;
  - weakening the seeded-secret quantifier.
- Removing the `scanSecretBytes` call inside `ScanDockerArchive` failed nothing.

- **F1 (blocking, executed):** the image secret scan cannot see inside the image. BuildKit
  `type=docker` layers are gzip blobs and `ScanDockerArchive` scans raw outer tar members.
  - An image built with `content/leaked.txt` holding a PEM private-key header passed
    `make release-secret-scan` with "no findings", although the file is present in the layer.
  - The only image test appends the sentinel after the tar end marker.
  - AC2's image-scan negative is therefore unproven. Parent check: `ScanDockerArchive` at HEAD
    still does not decompress layers.
- **F2 (blocking, executed):** `ValidateBundle` accepts a manifest whose Caddy or Postgres digest,
  `epoch_id`, `constants_hash`, `copy_hash` or `database_migration` has been forged. All six
  single-field forgeries returned nil. It hash-binds files but never cross-checks the manifest
  against `compose.yml` or `content/`. The predeclared "change one image digest" and
  "forge epoch/copy identity" negatives, and AC8, are therefore unmet.
- **F3 (blocking for AC8 attribution, executed):** `third-party-licenses.txt` lists only direct
  client dependencies. The built client bundle also ships `pad-end` (MIT) and `tslib` (0BSD), and
  their notices are missing. 0BSD is also absent from the license allowlist.
- **Non-blocking:**
  - `compose.rotation.yml` is hash-bound but never validated.
  - YAML is decoded non-strictly.
  - The runtime closure hard-codes two runtime reads, with no staged-closure startup test.

### DP-C — exact range `9b916c2..ab70327` (86cf4d6, 23a590d, 120343c, ab70327) — **CHANGES REQUIRED**

Executed cold at `ab70327`:
- The focused `make test-go` exited 0; both Postgres tests skip natively, as designed.
- `make test-deployment-backup` exited 0. Both Postgres tests ran with no SKIP:
  - `TestPostgresBackupRestoreEmptyAndPopulatedIdentityIntegration`, with both subtests;
  - `TestPostgresRestoreRefusesNonCleanTargetIntegration`.
  The same result held on a clean rerun.
- Removing the `GODEBUG` AVX2 workaround still passed 3/3 under this emulator. That is
  informational only.

Severing probes:
- These failed as required:
  - removing the unresolved pre-upgrade protection;
  - removing newest-backup protection;
  - removing the 30-day daily keep;
  - removing the Restore manifest check;
  - making the clean-target check vacuous (real Postgres).
- These left every test green:
  - Restore payload length/sha256 verification;
  - ReadHeader payload length/sha256 verification;
  - the extracted dump sha256 check;
  - temp-file cleanup;
  - future-dated `CompletedAt` invalidation;
  - skipping dot-named temp files in `backupPaths`;
  - the post-restore migration identity check (real Postgres).

- **F1 (blocking, parent-verified in source):** `runSchedule` creates a backup before applying
  retention. Any entry in the target that is not a valid backup makes retention return an error,
  the worker exit and `restart: unless-stopped` re-run it. Examples are a crash-left `.backup-*.tmp`
  (AC4's mid-write/restart case) or `lost+found` on a dedicated ext4 mount. Each restart writes
  another full backup and never prunes, so the off-host target fills. The docs give the operator
  no recovery step, and HEAD has the same loop shape.
- **F2 (blocking):** the DP6 checksum checks have no failing witness. The header checksum is what
  lets schedule and retention recognize a truncated backup whose header is intact. Without a
  witness, such a file can count as "newest valid".
- **F4 (blocking as overclaim):**
  - There is no partial-rename or mid-write/restart fixture; the only crash-temp witness uses a
    non-dot name.
  - The real-Postgres lane has only migration-mismatch and non-clean-target negatives. Truncation,
    corruption, identity and manifest checks are envelope-level unit tests.
- **Non-blocking:**
  - The committed envelope is never re-read and verified before rename, and there is no directory
    fsync.
  - `UpgradeResolved` can never become true, so pre-upgrade backups are kept forever. This is a
    DESIGN-GAP to route.
  - The test-boundary expansion landed in the same commit as its CI-environment edit, so its
    predeclaration order cannot be shown. Hosted workflow behavior is unchanged.
  - The lane leaves its Postgres dependency container running.
  - RPO is measured from `CompletedAt`, not from snapshot time (this matters for DP-F).
  - The 256 MB `/tmp` tmpfs caps the dump size and is undocumented.

### DP-D — exact range `ab70327..3f58fea` (6c626c2, 23adca6, 3f58fea) — **CHANGES REQUIRED**

Executed cold at `3f58fea`:
- `make test-deployment-release` exited 0. `TestCaddyIntegrationDrainReleaseAndExactBackupRollback`
  ran against real Postgres 16 and Caddy (no SKIP). It is the lane's only test.
- The focused `make test-go` over nine packages exited 0: 80 PASS, 7 env-gated SKIP, 0 FAIL.

Severing probes:
- Every RFC-named AC3 severing failed at the controller level:
  - drain validity;
  - migration propagation;
  - epoch identity;
  - smoke;
  - the rotation-overlap minimum for JWT/bootstrap/cursor.
- In the real Caddy lane, these also failed as required:
  - skipping the transport courtesy publish;
  - disabling the artifact comparison;
  - making `/readyz` ignore draining.
- These production `DockerRuntime` observations left every test green:
  - forcing `CourtesyFrame=true`;
  - `VerifyIdentity` returning nil;
  - dropping the `ArtifactsVerified` count check;
  - forcing `ReadinessDown`.

- **F1 (blocking, executed and parent-verified):** `Rollback` runs `StopFailed` →
  `ResetDatabase` (`docker volume rm …postgres_data`) → `Restore`. The backup file and age identity
  file are checked only inside `Restore`. A throwaway fixture with a missing backup and a missing
  identity file removed the volume, then refused at `exact_restore`. A path typo therefore wipes
  the live database. AC5's missing-backup refusal only happens after that destruction, and there
  is no missing-backup fixture. The ordering is unchanged at HEAD.
- **F2 (blocking, executed):** `Release` runs candidate `Preflight` (`compose run gameserver
  validate-config` on the candidate Compose file) before `Prepare` performs `docker load`. The
  manifest pins the gameserver as a bare `sha256:` config ID, which does not resolve before the
  load: `pull access denied`, exit 1. A normal release therefore fails at preflight on a host
  without the image; the unit runner mock hides this. At HEAD, `Install` was reordered, but
  `Release` still preflights first.
- **F3 (blocking, executed):** `composeArgs` passes only `compose.yml`. `compose.rotation.yml`,
  which carries the previous keys, is never included. Any release or rollback during a JWT or
  bootstrap overlap recreates the gameserver without previous keys, which shortens DP4's mandatory
  overlaps. The rotation ledger is not bound to runtime secret state. This is unchanged at HEAD.
- **F4 (blocking):** no test calls `DockerRuntime.DrainCurrent` or `VerifyIdentity`.
  - Three drain properties are derived from one `cleanExit` bit.
  - `waitHTTPState(false)` accepts any non-204, so a Caddy 502 satisfies "readiness down".
  - The lane proves the gameserver drain and backup/restore components, not the helper's state
    machine as the log claims.
- **F5 (blocking for AC5):**
  - There is no irreversible-migration fixture or detection.
  - There is no wrong-image rollback fixture.
  - Rollback never loads or inspects the previous image.
  - The integration "rollback" restores into a new database on a shared server with the same code,
    not a clean volume with the previous version.
- **Non-blocking:**
  - `docs/ci.md` says failed release-lane runs retain containers; the Make target always removes them.
  - Both scope-expansion notes landed in the implementation commit, and the ISC license-allowlist
    expansion is unrecorded.
  - Suspicion, not executed: the `backup` service's `user: 70:70` cannot read an operator-owned
    0600 identity file on rollback.
- **Controller-level gates with no finding:** ledger ordering, locks, rotation continuity,
  downgrade refusal and window arithmetic.

### Consequence

None of DP-A–DP-D is approved; each exact range needs a corrective range from the implementer
and a fresh designated pass over that correction. Several findings persist at HEAD and affect
DP-F:
- the vacuous image-layer secret scan (so the DP-F2 "image scan passes" record proves nothing
  about layer contents);
- the release preflight-before-load failure (a clean-host R-006 release step would fail);
- destructive rollback ordering;
- the ignored rotation overlay.
Per the handoff, Claude records the findings and does not repair them.

## 2026-09-24 — Claude designated cross-party review: DP-E and DP-E/DP8 corrective

**Review by:** Claude, as the designated cross-party pass. A Claude review subagent worked in
detached worktrees at `65099e7` and `cf4ac25`, plus a scratch probe worktree. The parent session
read the C2 hunk in source.
**Recorded by:** Claude (parent session).
**Environment:** the same host as the DP-A–DP-D entry. Docker and host-listener commands ran
outside the sandbox. All worktrees, containers, networks and volumes from these lanes were removed.

### DP-E — exact range `3f58fea..65099e7` (e2fc1c4, 09d5027, 65099e7) — **CHANGES REQUIRED**

Executed cold at `65099e7`:
- `make test-deployment-operations` exited 0. promtool reported SUCCESS, 11 Go packages passed,
  and `TestPrivateOperationsProfileIntegration` passed against real Caddy, Prometheus, Alertmanager
  and node-exporter. Postgres-gated tests in those packages skip in this lane.
- `make verify-kernel-version` exited 2 on `09d5027`'s `server/production/intents.go`. This
  reproduces known defect (a).
- Removing only `statusWriter.Hijack` made the real-Postgres socket-revocation test fail with
  HTTP 500. This reproduces known defect (b).
- Both known defects are remediated by the corrective range below.

Severing probes:
- These failed as required:
  - a Prometheus host port;
  - a Caddy `/metrics` route;
  - a Caddy admin host port;
  - public-outage `for: 6m`;
  - restart `> 3`;
  - cleanup `> 1`.
- These 9 of 13 promtool mutations left the rule tests green:
  - public-outage `for` lowered to 1m;
  - Postgres `for` lowered to 0s and to 30s;
  - the backup deadline raised to 9999999;
  - the backup any-failure clause deleted;
  - restart `>= 2`;
  - the dead-letter consecutive-interval clause deleted;
  - the filesystem threshold changed to 0.5;
  - the journal-budget clause deleted;
  - the `cloud_clicker_ready == 0` clause deleted.

- **F1 (blocking, executed against pinned Prometheus v3.12.0):** alert 6 (cleanup failure) can
  never fire in the shipped stack, for two independent reasons:
  - The counter's `job` label is rewritten to `exported_job` under `job_name: gameserver`
    (default `honor_labels: false`). The rule's `{job="credential_cleanup"}` selector returns
    nothing.
  - The lazily created failure series starts at 1, so `increase()` reports 0 for the first failure
    after a process start.
  promtool feeds synthetic `job=` series, and the integration test scrapes a fixture `/metrics`, so
  neither caught this.
- **F2 (blocking, executed against pinned Alertmanager v0.32.1):** `VerifyAlertDelivery` treats a
  rise in `alertmanager_notifications_total`, which counts attempts, as success, and it is not
  bound to the nonce. Against a receiver that returned 500 for every `/alerts` delivery,
  `deployment-operations alert-test` exited 0. DP7's successful-delivery gate is therefore not
  enforced, and `docs/deployment.md` misnames the counter.
- **F3 (blocking, executed):** `admin :2019` exposes Caddy's full admin API on every Caddy network.
  A peer container read `/config/` and `POST /load`ed a replacement, and the public listener then
  served the injected config. A compromised gameserver, Alertmanager or exporter could rewrite
  public ingress.
- **F4 (blocking as overclaim):**
  - The nine surviving mutations above show the alert tests don't discriminate thresholds,
    clauses or lower duration bounds.
  - No alert is taken through fault → fire → clear, so the claim that the tests prove "firing for
    its exact duration/population and resolved" is unsupported.
  - The backup test covers only the `absent()` branch.
- **F5 (Medium, source-verified):** `host-observe` rewrites `host.prom` only on success, and no rule
  checks textfile freshness or scrape errors. A broken observer therefore serves the last healthy
  storage, journal and restart values indefinitely.
- **F6 (process):** the first filter did not run `verify-client` or the composed browser lane,
  which is how both known defects were pushed.
- **Open item:** AC7's optional raw-IP 7-day sink is refused, not built. It needs an explicit
  open item or DESIGN-GAP rather than a satisfied row.
- **Status at HEAD:** F1–F3 source is unchanged at HEAD. `alert.go` was rewritten in unreviewed
  DP-F work, and the attempt-counter check is still present there.

### DP-E/DP8 corrective — exact range `7b510df..cf4ac25` (e0e2201, 73bb4ee, 32ee744, 5d72591, cf4ac25) — **APPROVED**

Executed cold at `cf4ac25`, all exiting 0:
- `make test-deployment-operations`.
- `make verify-kernel-version`: `0.3.101`; the adversarial fixtures passed.
- `make verify-ci-topology`: 13 negative controls rejected.
- `make verify-push` (5m02s). All six leaves ran with no silently skipped lane:
  - core Go with `-count=1`;
  - the fast harness;
  - client type/build/unit: 6,662 passed, 22 skipped;
  - boundary/kernel/topology/copy checks;
  - three-engine browser: 20,049 passed, 3 skipped;
  - the performance lane;
  - the real composed lane, both terminals plus continuation plus recovery;
  - `verify-schema`.

Severing probes, each of which failed as required:
- removing the Hijack forwarding (real Postgres, HTTP 500);
- removing the `09d5027` history correction (exit 2);
- reverting `client/src/kernel/version.ts` to 0.3.100 (drift, exit 2);
- dropping or reordering a `verify-push` leaf (topology, exit 2);
- removing the snapshot anti-regression guard;
- removing successor-terminal delivery.

Findings recorded with this approval:
- **C1 (process):** `73bb4ee` carries player-visible client transport-ordering changes, which the
  predeclaration's "no gameplay behavior" boundary did not list, and it mixes them with the server
  fix in one commit.
  - They are ordering repairs, not new mechanics.
  - Except for C2, each is witnessed and fails when severed, `docs/game-ui.md` was updated, and the
    composed lane passes.
  - The C1 changes are covered by this verdict because they were executed and severed, not because
    they were predeclared.
- **C2 (Low, required follow-up):** the new early return
  `if (offset > 0 && priorPosition && offset <= priorPosition.offset) return true;` in
  `client/src/game-ui/runtime.ts` makes duplicate or regressed channel offsets count as consumed,
  where before they forced a resync.
  - Removing it fails no unit test, and the log does not describe it.
  - The revision cursor still guards authoritative state, so it does not block this range.
  - A witness (or a documented justification for silently consuming same-channel offset
    regressions) must land in the next client transport change.
- **Suspicion, not executed:** `statusWriter` does not forward `http.Flusher`.

This approval covers exactly `7b510df..cf4ac25`. It does not approve DP-E's AC7 alert or receiver
surface, and it does not approve any DP-F commit.

### Review-union status after these six verdicts

Only `7b510df..cf4ac25` is designated-approved. DP-A–DP-E each need a corrective range and a
fresh designated pass. DP-F (`65099e7..7b510df` and the deployment commits after `cf4ac25`) remains
unreviewed. No archival, release or R-006 claim changes.

## 2026-09-24 — Claude advisory inspection of DP-F construction (not a designated verdict)

**Review by:** Claude, in three advisory subagent passes plus one parent lane run.
**Recorded by:** Claude (parent session).

DP-F has not been handed off for designated review. These are advisory findings for the
implementer and do not approve or reject any range.

**Coverage:**
- The ranges inspected were `65099e7..0693bbd` (15 commits), `0693bbd..7b510df` (29 commits) and
  the four product commits after `cf4ac25` (44f5b3f, 6d89880, 578fda0, b7a9587) with their records.
- Cold runs, all exiting 0: focused `make test-go -count=1 -v` over the rehearsal, release,
  backup, operations, release-package and browser packages; `make test-deployment-rehearsal`
  (both `validate-build` records passed); and `make test-deployment-release`.
- The parent ran `make test-deployment-backup` at HEAD. It exited 0, and all four Postgres tests
  ran with no SKIP, including `TestRecoveryIdentityRejectsProjectionEventWithoutVerifiedRunIntegration`.

**Severing probes that failed their witnesses:**
- population, tool and step-output hash binding;
- install abort on smoke failure, and the clean-state recheck;
- the RPO bound;
- backup-class separation and the objective-overwrite refusal;
- populated-identity validation and `no_image`;
- next-run and gate-crossed journey guards, and `VerifiedRunRows`;
- the post-backup identity recheck;
- the board-wait coast and no-wait variants;
- the historical/current `rehearsal_images` rule;
- both save-version bounds.

**Probes that left tests green (advisory findings):**
- final-bundle manifest-byte and browser-manifest binding (still unwitnessed at HEAD);
- the scanner seeded-rule check;
- `finalExtendsBase` and the proof-before-base check;
- the empty-restore identity comparison;
- the resolved-alert wait;
- the checkpoint-exists refusal before destructive steps;
- the empty-backup-after-incident check;
- the browser DOM labels, whose driving code has never run.

**Advisory findings, all present at HEAD unless noted:**
1. **High.** `ValidateExecutionPlan` does not bind commands to populations. A plan of
   `/usr/bin/true` positives and `/usr/bin/false` negatives executes and records all rows as
   passed.
   - The shell denylist is basename-only: `dash -c`, `env bash -c` and `python3 -c` are all
     accepted.
   - `deployment-rehearsal` usage and validation errors also exit 1, so a typo counts as a caught
     severing.
   - The accepting `boundRunFixture` is itself a complete hand-authored dossier.
   - `forged_successful_evidence` therefore does not prove forgery is rejected, and the docs
     overstate final `validate`.
2. **High.** Seven-family alert-delivery evidence inherits DP-E F1/F2/F4.
   - It marks all 7×2 firing/resolved deliveries true from one aggregate attempt counter.
   - It was accepted with a receiver that failed every delivery, and with only 4 of 7 families
     notified.
   - Alerts are hand-posted, so the unfireable cleanup rule cannot show.
3. **High.** The structured secret-scan and supply-chain evidence inherit DP-B F1.
   - A seeded private key inside a gzip layer gave `findings: 0, image_archive_scanned: true`.
   - The DP-F `seeded_image_secret` probe seeds a flat tar member, which hides the defect.
   - The v5 candidate's "image scan passes" claim therefore says nothing about layer contents.
4. **High (source-verified, not executed).** `recover-populated` will fail at smoke on a real host:
   - Its core start is `--no-deps` gameserver and Caddy only.
   - `AuthenticatedSmoke` then runs the Alertmanager `alert-test` against a service that is not
     running.
5. **Medium.** Build-record checks bind only the manifest hash and source commit.
   - The real `supply-chain` command accepted records with a wrong archive hash, SBOM hash or
     Playwright config.
   - `independent_rebuild` and `normalized_sboms_equal` are self-asserted.
   - Log line 1881 quotes a truncated archive hash.
6. **Medium.** The bundle `removed_*` and `changed_sbom` negatives are satisfied by generic manifest
   byte integrity. Deleting the catalog and its manifest entry passes `ValidateBundle`, because
   the catalog is not a required artifact, so AC1's removed-catalog failure is unproven.
7. **Medium.** RPO is near zero by construction: the incident is declared right after the
   rehearsal's own fresh backup. The six-hour cadence is never measured, and RTO starts after
   stop/reset.
8. **Medium.**
   - `AbortInstall`'s `down --volumes` removes pre-existing `caddy_data` (ACME material) and the
     other named volumes, which contradicts "removes only the newly created" state.
   - The browser result records no origin, and it takes its manifest hash from a flag.
   - `RestoredIdentityMatch` is a literal, and the checkpoint/identities are not retained
     artifacts.
   - The log's "reopens regular backup files, exact-decodes both headers" overclaims an `Lstat`.
9. **Low.**
   - Browser guard exhaustion collapses to `workflow_failed` with no visible guard field.
   - Five early DP-F2 slices relied on the umbrella predeclaration only.
   - The alphabetical plan-order defect in `65099e7..0693bbd` is fixed at HEAD.

**Sound in these passes:**
- The install ordering (load, then preflight, then destructive steps).
- The board-arrival wait is bounded and fails loud.
- Recovery identity hashes full rows.
- Rollback schema compatibility has no older-binary-on-newer-schema path.
- The four post-`cf4ac25` product commits, whose checks discriminated in 10 of 10 severings.

**Consequence:** before DP-F handoff, Codex should repair the evidence-integrity items (1–3 and 5–6)
and the recovery smoke composition (4), in addition to the DP-A–DP-E correctives. Until then, an
R-006 run on the current candidate could produce accepting evidence that proves nothing.

## 2026-09-24 — DP-D corrective R1 predeclaration (Claude implements; Codex designated reviewer)

**Owner direction (Marco, 2026-09-24):** Claude implements accepted-RFC work directly. Under
AGENTS.md (c), Codex is therefore the cross-party designated reviewer for these Claude-authored
corrective ranges, and Claude never approves or archives its own work.

**Authority:** RFC DP5 and AC3/AC5, addressing the findings recorded in the 2026-09-24 DP-D verdict:
- F1: rollback destroys data before validating its inputs;
- F2: release runs preflight before loading the image;
- F5 (partial): rollback never loads or verifies the previous image.

**Change:**
1. `Release`: `Prepare(candidate)` (image load plus digest verification) runs before
   `Preflight(candidate)`, so the candidate-image config check can resolve the loaded image.
2. `Rollback`: after ledger authority, run these non-destructive steps in order before
   `StopFailed`/`ResetDatabase`:
   - `Prepare(previous)`: load and verify the previous release's images;
   - `Preflight(previous)`;
   - a new `VerifyRestoreInputs(previous, backup)`: target-path binding, regular non-empty backup
     file, owner-only age identity, and host-side `deploymentbackup.ReadHeader` (payload
     length/sha256 verified), with the header bound to the backup ID, server, previous manifest,
     epoch and pre-upgrade class.
   `restoreBackup` keeps the same checks by calling the shared helper.
3. Tests:
   - The exact call sequences change.
   - New severed stages `prepare` and `restore_inputs` in rollback must fail before
     `stop_failed`/`reset_database` appear in the call list.
   - A DockerRuntime-level test with a recording runner proves that a missing backup file, a
     group-readable identity file and a corrupt payload each refuse with no `down`/`volume rm`
     command issued.
   - Demonstrated failing-first: each new test fails on the pre-change code.
4. `docs/deployment.md` release/rollback sequence updated in the same commit.

**Not in scope:** rotation overlay (F3), DrainCurrent/VerifyIdentity witnesses (F4),
irreversible-migration and wrong-image fixtures (rest of F5). Those are the next corrective ranges.

## 2026-09-24 — DP-D corrective R1 implemented; R2 predeclaration (rotation overlay)

**R1 implemented at `1978660`** (Claude):
- Release now runs `Prepare → Preflight → pre-upgrade backup`.
- Rollback now runs `Prepare(previous) → Preflight(previous) → VerifyRestoreInputs → StopFailed →
  ResetDatabase → Restore`.
- `VerifyRestoreInputs` reads the host envelope with `deploymentbackup.ReadHeader` (payload
  length/sha256) and binds the backup ID, server, previous manifest, epoch and pre-upgrade class,
  plus a regular owner-only age identity.
- The ledger admits the new `supply_chain`/`restore_inputs` rollback stages.

Cold `make test-go` over deploymentrelease, cmd/deployment-release, deploymentrehearsal and
cmd/deployment-rehearsal (`-count=1`) passed. Severing probes, each restored exactly:
- moving `VerifyRestoreInputs` after `ResetDatabase` failed
  `TestRollbackUsesExactPreviousManifestBackupAndSevenDayWindow` and
  `TestRollbackRecordsEverySeveredExecutionStage`;
- restoring preflight-before-prepare failed `TestReleaseSequenceAndEveryRuledSeveringStage`;
- dropping the header read failed `TestDockerRuntimeVerifiesRestoreInputsWithoutRuntimeCommands`.

**R2 predeclaration (DP4/AC6, DP-D F3):** release, rollback, install and recovery compose only
`compose.yml`, so an open previous-key overlap is silently shortened on recreate.
- Add `RotationOverlay(ledgerPath)`: the overlay applies when both the JWT and bootstrap latest
  rows are `activated`, and the result is invalid when exactly one is. The bundled overlay requires
  both pairs, so a single-family overlap cannot be composed and must fail closed. A per-family
  overlay would change hash-bound bundle contents and break historical-bundle rollback, so it is
  recorded as a follow-up.
- `DockerRuntime` gains an absolute `RotationLedgerPath`, resolved in `normalized()`. Every Compose
  invocation goes through `runtime.composeArgs`.
- The CLI and rehearsal constructors pass the operator-state ledger.
- Witness: a recording-runner test covers no ledger, JWT-only, both open, bootstrap-only and both
  removed, plus a relative-path refusal.
- Severing either the overlay append or the single-open refusal must fail it.
- The docs are updated in the same commit.

## 2026-09-24 — R3: image-layer secret scan (DP-B F1) and process note

**Process note:** R2's predeclaration landed in the same commit as its implementation (`92fab70`).
That is the same ordering defect Claude recorded against earlier Codex batches. It is disclosed
here for Codex's designated review, and later ranges keep predeclaration and implementation in
separate commits.

**R3 change:**
- `ScanDockerArchive` now recursively opens gzip-compressed and plain tar layers, plus nested
  gzip/tar files inside them, and scans every leaf file and tar entry name once.
- zstd, corrupt gzip, truncated tar and depth/size overflow fail closed.
- The DP-F `seeded_image_secret` probe now hides its sentinel inside a gzip layer.

**Evidence:**
- The new `TestSecretScanOpensCompressedAndNestedImageLayers` failed on the prior scanner for the
  gzip-layer, nested-gz and private-key-in-layer cases, and passes now.
- `TestSeededSecretProbesRequireAnActualScannerFinding/seeded_image_secret` fails with the prior
  scanner and passes now.
- Cold `make test-go` over releasepackage, cmd/release-secret-scan, deploymentrehearsal and
  cmd/deployment-rehearsal passed.
- `make release-secret-scan` against the real v5 candidate archive (one gzip layer) passed with no
  false positive.
- The first run caught this batch's own test file carrying a literal PEM header in tracked source;
  that literal is now split like the existing sentinel.
- The structured result still records no per-layer count, so no visible count of opened layers
  exists yet. This is a follow-up; the result schema is bound by the rehearsal evidence validator.

## 2026-09-24 — R4 predeclaration: manifest ↔ bundle cross-binding (DP-B F2)

**Defect:** `ValidateBundle` hash-binds files but never checks the manifest's own claims against
them. Forged Caddy/Postgres digests, `epoch_id`, `constants_hash`, `copy_hash` and
`database_migration` each validated as nil.

**Change:**
1. **Compose images.** `ValidateBundle` decodes the hash-bound `compose.yml`. Each service image
   must equal the manifest reference of the same name, the backup service must equal the Postgres
   reference, and the gameserver must equal its config ID.
2. **Content closure.** `ValidateBundle` re-derives the runtime closure from the bundle's own
   `content/` tree using the same `DeriveRuntimeClosure` authority, then requires:
   - an exact staged file set (`ValidateStagedContent`);
   - manifest `epoch_id`, `constants_hash` and `copy_hash` equal to the derived values.
3. **Database migration.** Migrations are compiled into the gameserver binary, so no bundle byte
   derives `database_migration`. It stays bound at runtime by the private database inspection and
   `VerifyIdentity`. The docs will say this, and no bundle-level claim is made.
4. **Tests.** One negative per forged field (caddy, postgres, prometheus digests; epoch; constants;
   copy) must fail on the current code and pass after the change. A removed content file must fail
   even with a rebound manifest, which is AC1's catalog-removal case that generic integrity cannot
   cover. The retained `previous` and `candidate-v5a` bundles must still validate.

## 2026-09-24 — R4 implemented: manifest claims bound to Compose and content

**Implementation:** `ValidateBundle` adds two checks.
- `bindComposeImages`: the service set and every image equal the manifest (backup equals Postgres).
- `bindContentIdentity`: `DeriveRuntimeClosure` runs over bundle `content/`, followed by
  `ValidateStagedContent`, and the epoch, constants and copy values must equal the manifest's.

The manifest test fixture now stages the real runtime closure instead of a placeholder epoch file.

**Evidence:**
- New `TestValidateBundleBindsManifestClaimsToComposeAndContent` covers seven forgeries: the
  caddy, postgres and prometheus digests; epoch; constants; copy; and a removed catalog with a
  rebound manifest. All seven failed on the pre-change validator.
- The first Prometheus variant reused the fixture's own all-`f` digest, which was a test-driver
  error. It is corrected to a zero digest.
- Cold `make test-go` over releasepackage, deploymentrelease, deploymentbackup,
  deploymentrehearsal, operations and their cmds passed.
- A throwaway validator (created and deleted, never committed) accepted the retained real
  `previous`, `candidate-v4a`, `candidate-v5a` and `candidate-v5b` bundles, so historical-bundle
  rollback is not broken.
- `database_migration` stays runtime-bound, as predeclared.

## 2026-09-24 — R5 predeclaration: operations alert corrections (DP-E F1/F2/F3)

**F1 — the cleanup alert can never fire.**
- Rename the application label `job` to `job_name` on `cloud_clicker_job_runs_total` and
  `cloud_clicker_job_last_success_timestamp_seconds`. `job` is a Prometheus target label, and with
  default `honor_labels: false` the application's value becomes `exported_job`.
- Pre-create every closed-set job's success and failure series at 0, so the first failure is an
  `increase()`.
- Update the rule and promtool fixtures.
- Witnesses:
  - a registry test proves no family exposes a reserved target label (`job`/`instance`), and that
    failure series exist at 0 before any run;
  - the Docker operations lane serves the **real** registry, not a hand-written text fixture;
  - after one real `ObserveJob` failure, Prometheus must report `CloudClickerCleanupJobFailed`
    firing via its API. Severing either the label rename or the pre-creation must fail this.

**F2 — delivery counts attempts.**
- Delivery success is the rise in `alertmanager_notifications_total − alertmanager_notifications_failed_total`,
  summed over integrations, and any rise in `failed_total` during the window fails. This applies to
  both `VerifyAlertDelivery` and `ObserveReleaseFloorAlertDelivery`.
- Unit witness: a metrics fixture where attempts rise but every attempt fails must be rejected.
- Real-lane witness: an Alertmanager receiver that answers 500 must fail `VerifyAlertDelivery`.
- Stated limitation: Alertmanager metrics cannot attribute deliveries to alert families or to the
  nonce without receiver-side evidence, so the per-family booleans remain count-derived. That is
  documented and not claimed as per-family proof.

**F3 — the Caddy admin API is reachable by every peer.**
- Remove `admin :2019` so the admin API keeps Caddy's default localhost-only listener.
- Serve Caddy metrics from a dedicated `:2020` read-only `metrics` site.
- Compose exposes `2020`, and Prometheus scrapes `caddy:2020`.
- `ValidateCaddyfile` rejects a non-localhost admin listener, and `ValidateCompose` requires the
  `2020` exposure.
- Witnesses:
  - template/unit rejection of `admin :2019`;
  - in the operations lane, a peer request to `caddy:2019/config/` must fail, while Prometheus's
    `caddy` target is up on `:2020`.

Docs are updated in the same commits. No hosted CI or timeout change.

## 2026-09-24 — R5 implemented: operations alert corrections

- **F1.** The job metrics use `job_name` instead of `job`, and each closed-set job's success and
  failure series is pre-created at 0. The rule and promtool fixtures are updated. The operations
  lane now scrapes the real registry and requires `CloudClickerCleanupJobFailed` to fire after one
  real `ObserveJob` failure.
- **F2.** Delivery is measured as `notification_requests_total − notification_requests_failed_total`,
  and any rise in failed requests fails. This correction was found by execution: the first
  implementation used the notification-level `notifications_failed_total`, and the lane showed it
  does not move until retries are exhausted, so the rejecting receiver still passed. Pinned
  Alertmanager v0.32.1 exposes the request-level counters, confirmed live.
- **F3.** `admin off` is set, and a dedicated `:2020` site serves Caddy metrics only.
  - Compose exposes `2020` and Prometheus scrapes `caddy:2020`.
  - `ValidateCaddyfile` rejects network, default or duplicate admin listeners and metrics outside
    `:2020`.
  - The lane asserts that `caddy:2019/config/` is unreachable from a peer.

**Evidence (cold):**
- promtool passed (SUCCESS).
- `make test-go` over operations, cmd/deployment-operations, releasepackage, deploymentrelease
  and deploymentrehearsal passed.
- `make test-deployment-operations` passed: `TestPrivateOperationsProfileIntegration` took 66s and
  covered the real registry, delivery, cleanup firing, the rejecting receiver and admin
  unreachability.
- The unit/validator test `TestCaddyRejectsMissingWebSocketRouteAndPublicMetrics` rejects
  `admin :2019`.

**Severing probes, each restored:**
- Removing series pre-creation failed `TestRegistryAvoidsTargetLabelsAndPrecreatesJobFailureSeries`.
- Reverting the label to `job` failed three registry tests. Separately, the Docker integration
  alone, compiled with that mutation, failed with "never scraped the pre-created cleanup failure
  series".
- Restoring `admin :2019` failed Caddyfile/bundle validation.
- The notifications-level counter variant let a rejecting receiver pass (observed above).
- The unit fixture `TestAlertDeliveryRejectsAttemptsThatFailed` rejects attempts equal to failures.

**Limitation (documented):** per-family delivery attribution is not observable from Alertmanager
metrics.

## 2026-09-24 — R6 predeclaration: backup schedule and checksum witnesses (DP-C F1/F2/F4)

**F1 — schedule loop.** A foreign entry such as `lost+found`, a crash-left temp or a corrupt
backup makes retention fail after a new backup has already been written. The worker exits, and
`restart: unless-stopped` repeats that until the target fills.

**Change:** `deploymentbackup.InspectTarget` classifies the target before any create:
- regular `*.ccbackup` files;
- Create's own `.backup-{payload,envelope}-*.tmp` files, split into stale (mtime older than
  1 h, crash debris) and active (possibly a concurrent pre-upgrade `create`);
- foreign entries.

Each round of the schedule worker then:
- removes and reports stale own temporaries;
- defers without a failure record while an active temporary exists;
- records a backup failure (so the any-failure alert fires) and retries in 5 min *without creating
  a backup* while any foreign entry or invalid backup exists;
- otherwise creates, records success and applies retention over the classified backups;
- never exits for a blocked target, so there is no restart churn.

The loop body becomes an injected-create round function so the decisions are unit-testable without
Postgres.

**Witnesses:**
- a foreign directory yields a failure record and zero create calls;
- a stale envelope temp is removed and the round creates;
- an active temp defers with neither a failure record nor a create;
- a corrupt backup blocks.

**F2 — checksum witnesses.** Tests where a backup's header is intact but its payload is truncated
or has a flipped byte must be rejected by `ReadHeader`, marked invalid by
`PlanRetention`/`EvaluateSchedule`, and refused by `Restore`. Removing each respective check must
fail its test.

**F4.** The crash-temp witness uses Create's real dot-prefixed names. Docs gain the operator
guidance that the target must be a dedicated directory, not a filesystem root containing
`lost+found`.

## 2026-09-24 — R6 implemented: backup schedule and checksum witnesses

**Implementation:**
- `deploymentbackup.InspectTarget` and `RemoveStaleTemporaries` classify the target (new
  `target.go`).
- `cmd/deployment-backup` `scheduleRound` inspects the target before creating: it removes stale
  own temporaries, defers on active temporaries, and blocks with a failure record and no create on
  foreign or invalid entries.
- The loop waits five minutes after deferred/blocked rounds and six hours after a created round,
  and never exits for a blocked target.

**Evidence (cold):** `make test-go` over deploymentbackup and cmd/deployment-backup passed. New
tests:
- `TestScheduleRoundNeverCreatesIntoABlockedTarget`: `lost+found` and a corrupt backup block with
  zero creates and a failure record; a stale `.backup-envelope-*.tmp` is removed and the round
  creates; an active `.backup-payload-*.tmp` defers without a failure.
- `TestHeaderChecksumGuardsScheduleRetentionAndRestore`: a truncated payload with an intact header
  is invalid for `ReadHeader`, `PlanRetention` and `EvaluateSchedule`. A same-ID substituted but
  age-valid payload is refused by `Restore` and `ReadHeader`.

**Severing probes, all restored:**
- dropping Restore's length/sha check failed the swapped-payload case;
- dropping ReadHeader's length/sha check failed both cases;
- disabling stale classification failed the stale-temporary round.

**Process note:** `target.go` was untracked during probing, so `git checkout` could not restore
it. The one-line mutation was restored manually, and that line was re-verified before the green
rerun.

## 2026-09-24 — R7: DP-A proxy-hop/origin wiring witness and decoder rows

**Process note:** R7 has no separate predeclaration commit. Its scope is exactly the remedy (a)
named in Claude's DP-A B1 verdict plus the N1/N2 rows, recorded here with its evidence.

**Change:**
- New `TestComposedProductionBoundaryTrustsExactlyOneProxyHopIntegration` in gameserver, on the
  declared Postgres service. It composes with `PublicOrigin` and `TrustedProxyHops: 1`, exhausts
  the unauthenticated limiter for `X-Forwarded-For: 203.0.113.10`, and requires a second forwarded
  client not to be limited. It then requires a WebSocket upgrade from the public origin to succeed
  (101) and one from a foreign origin to be refused (403).
- New deploymentconfig rows:
  - a JWT half-pair missing its ID;
  - readable secrets at a non-`/run/secrets` path and a non-normalized path, via a new
    `movedSecret` helper so a missing file cannot mask the rule;
  - an origin with userinfo;
  - an origin with a trailing dot.

**Evidence (cold):**
- `make test-save-integration SAVE_TEST_PACKAGES='./gameserver' SAVE_TEST_FLAGS='-run TestComposedProductionBoundary -v'`
  passed with no skip.
- Severing P11 (hop assignment) failed with "second forwarded client inherited the first client's
  limit".
- Severing P12 (origin assignment) failed with a 403 on the configured origin.
- `make test-go ./deploymentconfig` passed. P7 (prefix), P8 (normalization), P9 (userinfo),
  P10 (trailing dot) and P3 (half-pair incl. JWT) each now fail their named row.
- All mutated files were restored, with a clean diff against the implementation.

**Carried:** N3 (duplicate env keys are unreachable through `os.Environ`) and N4 (the shared-env
allowlist) remain DP-B/`.env`-preflight items.

## 2026-09-24 — R8 predeclaration: shipped client attribution (DP-B F3)

**Defect.** `gen-release-metadata` attributes only the three direct `client/package.json`
dependencies. The built client also ships `pad-end@1.0.2` (MIT) and `tslib@2.8.1` (0BSD), and
neither notice is delivered.

**Change.**
- Derive the client inventory from the build's own module graph: every `client/dist/**/*.map`
  `sources` entry under `node_modules/`, resolved on disk to the exact package directory. This
  distinguishes, for example, the installed `break_infinity.js` 1.3.0 from the aliased, shipped
  2.2.0.
- Each resolved package still needs a LICENSE/COPYING file whose detected license equals its
  `package.json` license. Fail closed if the build output or its sourcemaps are absent. Over- and
  under-attribution are both impossible by construction: only shipped modules are listed, and all
  of them are.
- **License policy change, explicit:** add `0BSD`, recognized as ISC-form permission text without
  the notice-retention clause. ISC recognition also accepts the common "and/or distribute" wording.
  This expands the audited permissive allowlist by one license, needed for a dependency already
  shipped to players, and needs owner/review visibility.

**Witnesses.**
- A fixture-root unit test: a sourcemap naming a transitive package that is not in
  `package.json` must appear in the inventory. A missing `dist` fails, a license/metadata mismatch
  fails, and 0BSD and ISC texts classify correctly.
- A real run of `make generate-release-metadata` against the current build must list `pad-end` and
  `tslib`.

## 2026-09-24 — R8 implemented: shipped client attribution

**Implementation:**
- `gen-release-metadata` discovers npm packages from `client/dist/**/*.map` sources, resolving
  each to its package root (scoped packages supported) with `package.json` license equal to the
  detected LICENSE.
- `DetectPermissiveLicense` collapses whitespace, and the ISC-form grant becomes `0BSD` when the
  notice-retention clause is absent.
- The SPDX license pattern admits `0BSD`. This is the explicit license-policy expansion
  predeclared at `6214874`.

**Evidence:**
- New unit tests: `TestClientInventoryFollowsShippedModulesNotDirectDependencies` covers a
  transitive and a scoped package, a license mismatch and a missing build.
  `TestDetectPermissiveLicenseSeparatesISCAndZeroClauseBSD` covers both classifications.
- The existing `TestDetectPermissiveLicenseRecognizesISC` initially failed because wrapped real
  ISC text split the notice clause. That failure led to the whitespace normalization.
- A real `make build-client` plus `make generate-release-metadata` produced 53 dependencies. The
  npm set is exactly `@antimatter-dimensions/notations@1.6.0`, `break_infinity.js@2.2.0`,
  `pad-end@1.0.2`, `svelte@5.56.8` and `tslib@2.8.1`.
- Diffing licenseConcluded against the retained v5 candidate SBOM shows only those two additions;
  no Go dependency was reclassified by the normalization.
- Severing probes (both failed, both restored): disabling scoped resolution, and mapping 0BSD to
  ISC.

**Process note:** `releasepackage/bundle_test.go` was committed unformatted in R4 (`751717f`). It
is gofmt'd in this range.

## 2026-09-24 — R9 predeclaration: production release observers (DP-D F4)

**Defects:**
- `DockerRuntime.DrainCurrent` and `VerifyIdentity` have no test.
- `waitHTTPState(..., false)` treats any non-204 or transport error as "readiness down". A Caddy
  502 after the gameserver died therefore satisfies `ReadinessDown`.
- Three drain properties are derived from one exit bit.

**Change:**
1. **Readiness down.** It now requires the gameserver's own `503 Service Unavailable` from
   `/readyz` while draining, polled at 25 ms. A 502/504 proxy error or a transport error is not
   evidence.
2. **Drain evidence.** Derivation moves into a pure `deriveDrainEvidence` from the observed
   readiness, socket (courtesy/closed), stop error, inspect error, exit code and elapsed time
   against the bound.
   - `IntentsRefused`, `AdmittedComplete` and `JobsFlushed` stay derived from a clean exit 0,
     because the gameserver exits 0 only after its admission gate closes and its admitted
     work/jobs/outbox/transport shutdown returns. The real Caddy lane separately witnesses the
     refusal.
   - The docs will state this derivation plainly instead of implying three independent
     observations.
3. **Tests.**
   - A `deriveDrainEvidence` table: each single failing input yields invalid evidence, and only
     the all-good input is valid.
   - A readiness-poller test: 502, then transport error, never counts, while 503 does.
   - A `VerifyIdentity` table against a bundle with the real staged runtime closure: a matching
     inspection passes, and a wrong migration, epoch, constants hash or artifact count each fail.
   - Severing the 503 requirement, or the artifact-count comparison, must fail its test.

## 2026-09-24 — R9 implemented: release observer witnesses

**Implementation:**
- `waitHTTPState(false)` now requires the gameserver's `503` (25 ms poll). A 502/504, a 204 or a
  transport error never counts.
- The drain derivation moves into the pure `deriveDrainEvidence`, with the process-contract
  attestation of three properties stated in code and docs.

**Evidence (cold):**
- `make test-go ./deploymentrelease` passed with three new tests:
  - `TestDeriveDrainEvidenceRequiresEveryObservation`: eight single-failure inputs;
  - `TestReadinessDownRequiresTheGameserverDrainingAnswer`: 502, 504, 204 and a vanished
    upstream are rejected, 503 is accepted;
  - `TestDockerRuntimeVerifyIdentityBindsEveryInspectedIdentity`: real staged closure; wrong
    migration, epoch, constants or artifact count each rejected.
- `make test-deployment-release` passed on real Postgres+Caddy with the stricter readiness rule.

**Severing probes, each failed and was restored:**
- restoring any-non-204 readiness failed the readiness test;
- dropping the artifact-count comparison failed the identity test;
- ignoring the exit code failed the derivation test.

**Remaining DP-D F5:** the irreversible-migration and wrong-image rollback fixtures are the next
range.

## 2026-09-24 — R10 predeclaration: rollback-authority negatives (DP-D F5)

**Wrong image.** Since R1, rollback runs `Prepare(previous)` before any destructive step, and
`Prepare` compares each image's inspected ID to the manifest config identity. The existing
`docker_test.go` wrong-ID case covers the refusal, and R1's severed `prepare` rollback stage proves
the controller stops before `stop_failed`/`reset_database`. No new code is needed; this entry
records the coverage.

**Irreversible change — the interpretation recorded for Codex review.** DP5 makes rollback
restore-based: no Down migration ever runs, and a forward migration or content change is undone
by restoring the exact pre-upgrade backup into a clean volume. A change becomes irreversible
exactly when that backup cannot serve as rollback authority. Today `createBackup` trusts the
backup container's reported header and only `Lstat`s the host file.

**Change.** After the container reports completion, `createBackup` reads the host envelope with
`deploymentbackup.ReadHeader`, which verifies payload length and SHA-256. It then requires the
decoded header to equal the reported one, bound to the current manifest, epoch, server and class.
Otherwise release fails at `preupgrade_backup` before drain, so a release whose rollback authority
is unusable never starts.

**Witnesses.**
- The existing backup/restore test switches to a real age envelope.
- New negatives: a corrupt host payload, and a host header that differs from the reported header.
  Each must refuse, and severing the host re-read must fail them.

## 2026-09-24 — R10 implemented: rollback authority is the host envelope

**Implementation.** `createBackup` re-reads the host envelope with `deploymentbackup.ReadHeader`
and requires equality with the reported header. Release refuses at `preupgrade_backup` otherwise.

**Tests.**
- `TestDockerRuntimeBindsBackupAndRestoreOutputToExactManifest` now uses real age envelopes for
  both the pre-upgrade and recovery classes.
- New negatives: a reported header one second off from the host envelope, and a corrupt host
  payload. Each refuses.

**Severing.** Disabling the re-read made the differing-header case pass, so the test failed. The
change was restored.

**Evidence (cold).** `make test-go` over deploymentrelease and deploymentrehearsal passed.
Wrong-image rollback is covered by `Prepare(previous)` (R1 plus the existing wrong-ID test), as
recorded in the predeclaration.

## 2026-09-24 — R11: discriminating alert fixtures and stale host observation (DP-E F4/F5)

**Process note:** R11 has no separate predeclaration commit. Its scope is the remedy for Claude's
DP-E F4/F5 findings, recorded here with its evidence.

**Change:**
- `CloudClickerStoragePressure` gains stale-observation clauses:
  - `time() - node_textfile_mtime_seconds{file=~".*host\.prom"} > 300`;
  - `node_textfile_scrape_error == 1`;
  - absence of the host textfile.
  Its summary text is extended accordingly. Seven families are unchanged; there is no eighth alert.
- The promtool population gains discriminating and resolution cases (see docs).
- `TestOperationsProfileRejectsMissingAlertAndResolvedEvidence` now removes **every** resolved case
  for the probed alert. The new cleanup resolution case had made its single replacement
  non-discriminating (caught by execution).

**Evidence:**
- promtool passed (SUCCESS).
- Each of 13 mutations was killed:
  - public `for` lowered to 1m;
  - Postgres `for` lowered to 0s and to 30s;
  - the backup deadline raised to 9999999;
  - the backup any-failure clause removed;
  - restart `>= 2`;
  - the dead-letter second interval removed;
  - filesystem 0.5;
  - the journal clause removed;
  - the ready clause removed;
  - the stale-mtime clause removed;
  - the scrape-error clause removed;
  - the cleanup window widened to 60m.
- `make test-go ./releasepackage` passed.
- `make test-deployment-operations` passed (integration 67s).

## 2026-09-24 — R12 predeclaration: bind rehearsal plan rows to their producers (DP-F advisory 1)

**Defect (advisory, verified).** `ValidateExecutionPlan` accepts any non-shell executable per row.
A plan of `/usr/bin/true` positives and `/usr/bin/false` negatives executes and records every row
as passed. Negative rows expect exit 1, which `deployment-rehearsal` also emits for usage and
validation failures, so a typo counts as a caught severing.

**Change.**
1. **Exit code.** `deployment-rehearsal probe` exits **3** only for `ProbeRejected`, the fully
   prepared named fixture rejected by its production gate. Usage, invalid-evidence and setup errors
   keep exit 1 or 2. Plan negatives must expect exit 3.
2. **Producer binding.** `ValidateExecutionPlan` binds every row to its owning producer.
   `Command[0]` must be an absolute path whose basename is `deployment-rehearsal`, and
   `Command[1]` must be that population's subcommand:
   - `probe`, carrying exactly one `--population=<row name>`, for every population `RunProbe`
     supports;
   - `install-candidate`, `run-browser` and `recover-empty` for their positives;
   - `recover-populated` for the populated-identity, RPO and RTO positives.
3. **Unbindable populations.** A population with no producer yet makes every plan invalid. This is
   deliberate: no R-006 plan can pass until each population has a real producer, which matches the
   honest 25/43 state. The unbindable set is listed in the docs.
4. **Witnesses.**
   - The constant-command plan is rejected.
   - A `probe` row naming a different population is rejected.
   - A negative row expecting exit 1 is rejected.
   - The CLI returns 3 for a rejected fixture and not for a usage error.
   - Severing the binding or the exit code must fail these tests.

The existing plan-execution tests switch to bound fixture commands.

## 2026-09-24 — R12 implemented: rehearsal plan rows bound to producers

**Implementation:**
- `ProbeRejectedExit = 3`: the CLI `probe` exits 3 only for `ProbeRejected`.
- `populationProducers` plus `boundToProducer` make every plan row require an absolute
  `deployment-rehearsal` tool, the population's producer subcommand, and exactly one matching
  `--population=` for `probe` rows.
- Plan negatives must expect exit 3.

**Tests:** plan tests now use a bound fixture executable named `deployment-rehearsal`. New
rejections cover constant true/false commands, a probe for another population, a negative
expecting exit 1, the wrong producer for install, and a relative tool path.

**Evidence (cold):**
- `make test-go` over deploymentrehearsal and cmd/deployment-rehearsal passed.
- The real built CLI against the retained v5a/previous bundles gave:
  - `seeded_source_secret` → 3;
  - `removed_catalog` → 3;
  - unsupported `gameserver_restart_during_admitted_work` → 2;
  - typo subcommand → 1;
  - bad flag → 2.
- Severing the binding failed five rejection cases. Relaxing the negative exit failed the
  generic-exit case. Both were restored.

**Remaining DP-F advisory items:**
- 2: alert-delivery attribution is limited as documented in R5; the observation still marks all
  families from a count.
- 4: `recover-populated` smoke composition.
- 5: build-record self-assertions.
- 6: the `removed_*` probes now also fail semantically after R4, but still do not rebind the
  manifest.
- 7: RPO by construction.
- 8: `AbortInstall` volumes.

## 2026-09-24 — R13 predeclaration: recovery smoke composition and install volume safety (DP-F advisory 4 and 8)

**4 — recovery smoke.** `StartRecoveryCore` runs `up --no-deps` for `gameserver` and `caddy` only.
Its `AuthenticatedSmoke` ends in `verifyAlertDelivery`, a `compose run` that targets
`http://alertmanager:9093`, a service that is not running after `StopFailed`'s `down`. The step
would fail on a real host.

**Change:** a single `smokeDependencies` list (`gameserver`, `caddy`, `alertmanager`) drives
`StartRecoveryCore`.

**Witness:** a recording-runner test requires the recovery `up` to include every smoke dependency
and the alert test to target the started Alertmanager. Removing `alertmanager` from the list must
fail it.

**8 — initial-install volumes.** `requireCleanInstallState` checks only
`cloud-clicker_postgres_data`, yet a failed install's `AbortInstall` runs `down --volumes`, which
removes every declared named volume. That includes a pre-existing `caddy_data` holding ACME
material, contradicting the documented "removes only the newly created" guarantee.

**Change:** clean install state requires that none of the five declared project volumes exists
(`caddy_data`, `caddy_config`, `postgres_data`, `prometheus_data`, `alertmanager_data`). An install
over any retained volume refuses before anything starts, so `down --volumes` can only remove
volumes this install created. Operators who want to keep certificates must restore them after a
successful install.

**Witness:** a recording-runner test where only `cloud-clicker_caddy_data` exists must refuse
preflight with no `up` and no `down`. Narrowing the check back to `postgres_data` must fail it.

## 2026-09-24 — R13 implemented: recovery smoke dependencies and install volume safety

**Implementation:**
- `smokeDependencies` (`gameserver`, `caddy`, `alertmanager`) drives `StartRecoveryCore`.
- `requireCleanInstallState` refuses when any of the five declared project volumes exists.

**Test changes:**
- Codex's `TestDockerRuntimeRecoveryCoreExcludesUnobservedBackupWriter` previously required
  Alertmanager to stay stopped, which contradicted the smoke dependency. It now requires
  Alertmanager and keeps the backup writer, Prometheus and node-exporter excluded.
- New `TestRecoveryCoreStartsEverySmokeDependency`.
- A new occupied-host case for a retained `caddy_data`.
- The occupied-host subtests now reuse the clean-host runner, so only the occupied state differs.

**Evidence (cold):** `make test-go` over deploymentrelease and deploymentrehearsal passed.
Severing probes, each failed and was restored:
- removing Alertmanager from `smokeDependencies`;
- narrowing the volume check to `postgres_data`, which failed the retained-certificate case;
- ignoring containers.

**Correction to this batch's own probing:** one severing run was first misread as a survivor. It
was a build failure (an unused import) hidden by an output filter. Re-verification on the
*original* occupied-host harness showed it was **not** vacuous: its no-helper assertion caught the
container severing. The harness change is an isolation improvement, not a repair of a vacuous test.

## 2026-09-24 — R14 predeclaration: bind build records to bundle bytes (DP-F advisory 5)

**Defect (advisory, verified by execution).** `ObserveSupplyChain` binds each build record only
by manifest hash and source commit. The real `supply-chain` command accepted a candidate record
with a wrong `gameserver_archive_sha256`/`rebuild_archive_sha256`, a wrong gameserver
`sbom_sha256`, or a wrong Playwright `runtime_config_sha256`.

**Change.** For both the candidate and the previous record, `ObserveSupplyChain` requires:
- `gameserver_archive_sha256` equal to the SHA-256 of the bundle's `images/gameserver.docker.tar`,
  and `rebuild_archive_sha256` equal to it;
- `rebuild_manifest_sha256` equal to the manifest hash;
- every record image and rehearsal image equal, in order, to the manifest's name, reference,
  runtime config and SBOM hash.

`ValidateBundle` already binds those manifest fields to Compose and the SBOM bytes (R4), so each
record field reduces to bundle bytes.

**Limitation stated, not claimed.** `independent_rebuild` and `normalized_sboms_equal` describe the
build procedure and cannot be recomputed from one bundle. They remain operator-attested, and the
docs say so.

**Witnesses.**
- Unit cases, each one wrong field that the current code accepts: archive, rebuild archive,
  rebuild manifest, a gameserver SBOM hash, and a Playwright config.
- A real `make deployment-rehearsal-supply-chain` run against the retained v5a/previous bundles
  and records must still pass.
- The same run with one mutated record field must fail.

## 2026-09-24 — R14 implemented: build records bound to bundle bytes; retained bundles invalidated

**Implementation.** `bindBuildToBundle` in `ObserveSupplyChain` applies to both the candidate and
the previous record. The manifest images decode as typed `BuildImage` values.

**Tests.** `TestObserveSupplyChainBindsBuildRecordFieldsToBundleBytes` covers:
- a self-consistent wrong archive/rebuild-archive pair;
- a wrong gameserver SBOM hash;
- a wrong Playwright config.

The first draft also mutated single self-consistency fields. Those cases were already rejected by
record decoding, so they did not witness byte binding and were replaced (found by severing).
Severing the candidate binding fails all three cases.

**Real-run finding (important for DP-F).** `make deployment-rehearsal-supply-chain` against the
retained `candidate-v5a`/`previous` bundles now fails on committed HEAD as well as with R14. The
cause is `ValidateBundle`. Since R5 it requires `admin off` and Caddy `expose: ["2020"]`, and
**every retained bundle (v1–v5 candidates and the previous bundle) still exposes the admin API on
`:2019` and is correctly rejected.** Consequences:
- The earlier R4 statement that those bundles validate was true at R4 and became false at R5.
- No retained bundle can be an R-006 candidate *or* previous input. Both must be rebuilt from
  corrected source, the previous one at a lower release version.
- Current security validation also applies to the previous bundle, so a pre-R5 bundle can no
  longer be a rollback target. No instance is deployed, so no real rollback is affected. This is
  a policy choice flagged for Codex/owner review, not silently decided.

A real run of R14 on rebuilt bundles is therefore still owed.

## 2026-09-24 — R15: semantic bundle-mutation probes (DP-F advisory 6)

**Process note:** R15 has no separate predeclaration commit. Its scope is the recorded advisory
remedy.

**Change (`deploymentrehearsal/probe.go`):** `runBundleMutationProbe` now works in three steps.
1. It validates the unmutated hardlinked candidate first; failure is a setup error, never a
   rejection.
2. It applies the mutation, then `rebindManifest` rehashes every artifact and a mutated image
   SBOM's hash. The manifest file is rewritten, never the hardlinked source.
3. It requires the gate to reject on meaning.

`changed_sbom` now changes the SBOM subject (`sha256-<config>`), which is semantically bound by
`ValidateImageSPDX`, instead of appending a JSON-valid space.

**Tests:**
- The mutation-probe test now requires exactly two validator calls, a clean baseline first.
- The setup-failure test permits only the baseline call.

**Evidence:**
- `make test-go` over deploymentrehearsal and cmd/deployment-rehearsal passed.
- The real CLI `removed_catalog` against the pre-R5 v5a bundle now exits **2** (baseline
  invalid). Consequence: the R12 real-run observation "`removed_catalog` → 3" was **vacuous**.
  That bundle was already invalid after R5, so any mutation "was rejected".

**Coverage boundary.** Semantic rejection of each rebound mutation on a *real current-rule* bundle
still needs a rebuilt candidate. R4's releasepackage test already proves the removed-catalog case
with a rebound manifest on an assembled bundle.

## 2026-09-24 — R16: verify the committed envelope; sync the rename (DP-C F3)

**Process note:** there is no separate predeclaration commit. The scope is the recorded DP-C F3
remedy.

**Change:**
- `Create` now reads the closed temporary envelope with `ReadHeader` and requires it to equal the
  header it built before renaming.
- After the rename it fsyncs the target directory.
- An unexported `beforeEnvelopeCommit` test seam (nil in production) lets a test tear the
  envelope.

**Evidence:**
- `TestCreateVerifiesTheCommittedEnvelopeBytes` truncates the closed envelope by one byte. `Create`
  returns `ErrInvalid` with no final path, and the target directory is empty.
- Disabling the verification failed the test; the change was restored.
- `make test-go` over deploymentbackup and cmd/deployment-backup passed.

**Limit:** the directory fsync has no crash-simulation witness; it is stated in the docs, not
proven by a test.

## 2026-09-24 — R17: witness the browser-manifest binding (DP-F advisory 8)

**Process note:** there is no separate predeclaration commit.

**Change (tests only).** `TestBaseRunRejectsRehashedOperatorAuthority` gains two cases:
- a schema-valid browser result for another manifest, with its digest rebound;
- a different candidate-manifest artifact, with its digest rebound.

**Findings from severing:**
- Removing the `browser_result` manifest comparison now fails the new base-run case, so that check
  is witnessed.
- Removing the `candidate_manifest` byte comparison is **not** detected, because another binding
  rejects the same forgery. The check is redundant defence-in-depth and not independently
  load-bearing, consistent with the advisory note. No claim is made that it is.
- The same two cases placed in the *final-run* table
  (`TestRunValidationRejectsForgedArtifactsAndStepAggregation`) passed with either check removed.
  Rewriting the final evidence breaks its seal binding to the unchanged base evidence. The
  existing "rehashed …" rows in that table therefore mostly witness the seal rather than their
  named typed checks.
  - Typed-artifact witnesses belong in the base-run table.
  - Auditing the existing final-run rows individually is a recorded follow-up for Codex.

## 2026-09-24 — Verification of R1–R14 and DESIGN-GAPs for the RFC author

**Cold aggregate.** `make verify-push` exited 0. It started after R14 (`ee551f1`); R15–R17 touch
only rehearsal and backup code, which was tested separately. Results:
- 58 Go packages ok, with cold core and fast harness;
- client unit tests: 6,662 passed, 22 skipped;
- three-engine browser: 20,049 passed, 3 skipped;
- performance lane: 1 passed, 17 skipped;
- the composed Game UI v3 path (both terminals, next-run continuation and WebSocket recovery)
  passed;
- `verify-schema` passed.

**DESIGN-GAPs.** None of these is decided by Claude, per AGENTS.md; each needs an RFC-author or
owner ruling.

1. **RPO semantics (DP6, AC4).** The recovery producer declares the incident immediately after its
   own fresh backup, so the measured RPO is about zero by construction and never exercises the
   six-hour cadence. Should R-006 RPO be measured from the newest *scheduled* backup, possibly
   over a longer or forced-late window? Or should the RFC state that the cadence is proven by the
   backup-missing alert rather than by the RPO row?
2. **Per-family alert delivery (DP7, AC7).** Alertmanager metrics cannot attribute a delivery to
   one alert family or nonce. Proving seven fired-and-resolved families needs receiver-side
   evidence: a rehearsal-owned receiver alongside the operator's, or receiver logs. The current
   observation marks families from a count (R5 documents this).
3. **`UpgradeResolved` lifecycle (DP6).** No code ever sets it, so every pre-upgrade backup is kept
   forever. Something has to record that an upgrade is "resolved", perhaps the rollback window
   closing.
4. **Rotation overlay granularity (DP4).** The bundled overlay binds JWT and bootstrap previous keys
   together, so a single-family overlap now refuses release (R2). Should the bundle ship
   per-family overlays? That changes hash-bound bundle contents.
5. **Security floor for previous bundles (DP5).** Current `ValidateBundle` rules apply to the
   previous bundle, so every pre-R5 bundle, which exposes the Caddy admin API, is refused as a
   rollback target. Nothing is deployed, but the policy for future floor changes should be ruled:
   is it acceptable to refuse rollback to a less-secure previous release?
6. **Asserted host fields (DP-F evidence).** `Host.CleanStart`, `SourceCheckoutAbsent` and
   `ProviderCredentialsAbsent`, the kernel/Docker strings and `Objectives.RestoredIdentityMatch`
   remain producer-asserted values in the evidence schema. Deriving them needs host-side
   observation contracts.

## 2026-09-24 — C2 follow-up closed: consumed-offset witness

This closes the follow-up carried in Claude's approval of `7b510df..cf4ac25`: the
`client/src/game-ui/runtime.ts` early return for a channel offset at or below the consumed position
had no witness and no description.

- **Test.** `test/game-ui-runtime.test.ts` gains "treats an already-consumed channel offset as
  delivered without re-emitting or resyncing". World presence at offsets 1 and 2 is delivered; a
  replay at offset 2 (count 99) and a regression to offset 1 (count 98) emit nothing, the socket
  stays open, and the stored world position remains offset 2.
- **Severing.** Disabling the early return fails the test; the change was restored.
- **Docs.** `docs/transport.md` now documents the rule.
- **Checks.** `make typecheck` passed, and the runtime suite passed 12/12.

This is Claude-authored client test and doc work, for Codex review.

## 2026-09-24 — Diagnostic rebuild under corrected rules; real R3/R4/R12/R14/R15 checks

**What was built.**
- Source: clean HEAD `b60a86e`, source timestamp `2026-09-24T12:50:45Z`.
- One tree per role:
  - `diag-previous` (`0.1.0-preview.3`);
  - `diag-candidate` (`0.1.0-preview.4`).
- Each tree contains the six Linux/amd64 commands, client, staged content, a no-cache gameserver
  archive, a real Syft SBOM for the new gameserver image, metadata (53 dependencies, including
  `pad-end` and `tslib`) and a 72-artifact bundle.
- Assembly ran the corrected `ValidateBundle` and the layer-opening image scan.
- Candidate gameserver config: `sha256:bb1e0575a869299073e843a6e968d991b7c7465b4b3f8ece0ccd2ba15c2366bf`.
- Candidate manifest: `sha256:434c0675c548e4d448883d17a001be5ae1d6653b1b06f709bf1db356ef5bc547`.

**Diagnostic limits (not release evidence):**
- one tree per role, with no two-tree reproducibility comparison;
- the six upstream normalized SBOMs reused from `candidate-v5a` (config IDs unchanged);
- the build records for the supply-chain run written from the actual bytes, with operator-attested
  flags copied, and not committed;
- `planning/deployment-foundation/release-builds/*.json` unchanged, still naming v5, which is now
  invalid.

**Real results on these bundles:**
- **Positive probe.** `six_image_sbom_license_provenance` exits 0, so both bundles validate under
  the corrected rules.
- **Negative probes.** All 24 real negative probes exit **3**. This covers every `removed_*` and
  `changed_*` probe with a rebound manifest (semantic rejection, R15), `seeded_image_secret` inside
  a gzip layer (R3), config, alert, journal, objective, backup and host probes, and
  `interrupted_backup_writer`.
- **Missing probe.** `non_clean_restore_target` exits **2**: it has no probe yet, and fails loudly
  (R12).
- **Secret scan.** `make release-secret-scan` covered 1,446 tracked files plus the new image
  archive with its layers opened, and reported 0 findings.
- **Supply chain (R14).** `make deployment-rehearsal-supply-chain` with byte-true records exited 0.
  A wrong gameserver SBOM, a self-consistent wrong archive pair and a wrong Playwright config each
  failed. These are the three cases the DP-F advisory pass showed being *accepted* before R14.

**Still not done.** There is no clean-host R-006 run, no two-tree reproducible candidate record, no
Codex review of the source, and no owner authority for the host.

## 2026-09-24 — R18: three more R-006 negative producers (DP-F)

**Process note:** there is no separate predeclaration commit. The scope is RFC AC3/AC5 negatives
that previously had no producer.

**New probes:**
- `wrong_epoch_or_artifact_set` is a bundle mutation. It decrements the packaged
  `current_epoch_id`, rebinds the manifest, and `ValidateBundle`'s content-identity binding (R4)
  rejects it.
- `irreversible_or_down_migration` runs the production `deploymentrelease.Controller.Release` over
  the exact `LoadBundle`-validated pair with a guard runtime. The unmodified pair must reach the
  runtime. A candidate migration one below the running release must fail at `compatibility`, with
  zero runtime calls and one failed ledger row.
- `missing_previous_image_or_backup` uses the production `DockerRuntime.VerifyRestoreInputs` and
  `Prepare`, with a scripted docker runner. A real age pre-upgrade envelope bound to the previous
  manifest and the intact image IDs are accepted first. Then the removed envelope and an absent
  image (`image inspect` fails) are both refused.

**Real evidence, CLI against the `diag-*` bundles:** all three exit **3**.

**Severing probes, each restored:**
- the release migration gate severed → `irreversible_or_down_migration` exits 0;
- the image-ID comparison severed → `missing_previous_image_or_backup` exits 0;
- the content-identity binding severed → `wrong_epoch_or_artifact_set` exits 0.

**Also:**
- `make test-go` over deploymentrehearsal and cmd/deployment-rehearsal passed.
- The generic mutation test gained an epoch fixture.
- Probe-backed populations are now 28 of 43: 27 negatives and 1 positive. Still without a producer:
  - negatives: non-clean restore target (Postgres) and restart during admitted work (live
    gameserver);
  - 12 runtime positives, which need the host.

## 2026-09-24 — R19: lifecycle release/rollback producers; rollback-row contract bug fixed

**Process note:** there is no separate predeclaration commit. The scope is producers for two
existing positives, `bounded_drain_and_restart` and `exact_previous_release_rollback`, whose exact
contract `validateOperatorRecords` already defined.

**Bug found by building the producer.** On a successful rollback, `deploymentrelease.Controller.Rollback`
cleared `PreviousVersion` and `PreviousManifestSHA256`. Failed rollback rows kept them. Final R-006
validation (`validateOperatorRecords` → `matchesTransition(lifecycle[2], "rollback", …)`) requires
the rolled-back-from candidate in those fields. A real run could therefore **never** produce
accepting lifecycle evidence; the rehearsal fixtures hand-wrote rows the controller cannot emit.

**Fix:** the successful rollback row keeps its rolled-back-from release fields. The ledger's
validity rules already permit this, and the rollback-authority chain does not read them. This is
witnessed by the new `TestSuccessfulRollbackRecordsTheReleaseItRolledBackFrom`.

**New producers (`deploymentrehearsal/lifecycle.go`, CLI, Make):**
- `lifecycle-release` requires the exact candidate install row and an empty lifecycle state. It
  runs `AbortInstall` for the candidate stack, `Install(previous)` and `Release(previous →
  candidate)` through the production controller, and requires the two exact ledger rows.
- `lifecycle-rollback` re-reads the release's pre-upgrade envelope, which must be bound to the
  previous manifest. It then runs `Rollback(candidate → previous)`, requires three rows, and writes
  the `release_ledger` and `backup_header` artifacts.
- Both are bound as the producers of their plan rows.

**Tests** (fake install runtime, real controller, real age envelope):
- Contract-order call sequences for both producers.
- Produced rows satisfy `matchesInstall`/`matchesTransition` exactly as the final validator applies
  them.
- Refusals with zero runtime calls: release before the candidate install, release over an existing
  lifecycle, and rollback before a release.
- A rollback with a missing backup never reaches `reset_database`.
- Restoring the old clearing fails both the controller test and the lifecycle test.

**Evidence:** `make test-go` over deploymentrelease, deploymentrehearsal and both cmds passed, and
`make verify-ci-topology` passed. Producer-backed populations are now 30 of 43. No host run has
happened.

**DESIGN-GAP 7 (rotation population).** `validateRotationAuthority` needs real JWT, bootstrap and
cursor activation/removal pairs that respect 30 min / 31 d / 366 d overlaps. A real-time rehearsal
cannot complete the 31-day and 366-day overlaps. DP4 allows a governed clock for tests, but R-006
evidence is a real run. The RFC author must rule how rotation is evidenced, for example the
activation plus a governed-clock removal fixture, or a long-running observation.

## 2026-09-24 — R20: non-clean restore producer with attributable refusal

**Process note:** there is no separate predeclaration commit. The scope is the RFC AC4 negative
"non-clean restore target", which had no producer.

**Change:**
- `deploymentbackup.RequireCleanTarget` returns the dedicated `ErrNonCleanTarget`, which wraps
  `ErrInvalid`, so existing callers still fail closed. A query error stays a generic `ErrInvalid`.
- The backup CLI reports it as error class `non_clean_target`.
- `deploymentrelease.IsNonCleanRestoreRefusal` recognizes that structured class.
- The new rehearsal producer `restore-non-clean` (CLI, bound to `non_clean_restore_target`) does
  the following on the installed candidate stack:
  1. requires the exact candidate install row;
  2. inspects the live recovery identity and creates a recovery backup;
  3. restores it in place;
  4. counts as rejected (exit 3) only on the `non_clean_target` class with an unchanged identity.
  An accepted restore exits 0, and any other failure exits 2. The probe removes its own backup.

**Evidence:**
- `make test-deployment-backup` (real Postgres) passed.
  `TestPostgresRestoreRefusesNonCleanTargetIntegration` now requires
  `errors.Is(err, ErrNonCleanTarget)`.
- Unit tests:
  - `TestNonCleanRestoreRefusalHasItsOwnClass`;
  - `TestNonCleanRestoreProducerRequiresTheCleanTargetRefusal`, which uses the exact slog/ExecRunner
    error rendering and covers refusal, acceptance, a refusal for another reason, an identity
    change, and a missing install.
- Severing the class attribution and severing the identity comparison each fail the producer
  test; both were restored.
- Two identity-severing attempts first failed to compile (unused variables). They were not counted.
- `make test-go` over deploymentrehearsal, cmd/deployment-rehearsal and deploymentrelease passed.

**Coverage:** producer-backed populations are now 31 of 43. Still without a producer:
- restart during admitted work (a live gameserver kill);
- key overlap (DESIGN-GAP 7);
- alert delivery (DESIGN-GAP 2);
- journal budget;
- provider-off operation.

No host run has happened.

## 2026-09-24 — R21: host clean-start covers every project volume; DESIGN-GAP 6 corrected

**Correction to DESIGN-GAP 6.** Reading `deploymentrehearsal/host.go` shows `SourceCheckoutAbsent`
and `ProviderCredentialsAbsent` are **already derived**. `sourceMetadataAbsent` (no `.git` ancestor
of any scenario path) and `providerCredentialsAbsent` (a closed environment-variable family list)
gate the observation, and the literal `true` values are written only after those checks pass.
`RestoredIdentityMatch` in the objective observation remains a producer-set literal (advisory F6),
and the kernel/Docker strings are observed command output. Gap 6 is narrowed to
`RestoredIdentityMatch` only.

**Fix.** Clean start checked only `cloud-clicker_postgres_data`. The host observer now refuses any
`cloud-clicker_*` volume, consistent with R13's install rule, so a host carrying retained Caddy
certificates or metrics/alert state is not reported as a clean start.

**Evidence.** A new `retained certificate volume` host negative. Reverting to the Postgres-only
check fails it; the change was restored. `make test-go ./deploymentrehearsal` passed.

## 2026-09-24 — DESIGN-GAP 6 fully re-read

`RestoredIdentityMatch` is also derived. `recovery.go` writes the objective only after
`CompareRecoveryIdentity(checkpoint.PopulatedIdentity, restored)` succeeds; otherwise the producer
fails. So none of the gap-6 fields is a bare assertion.

The residual concern (advisory F6) is **re-verifiability**: the checkpoint and both recovery
identities are not retained artifacts, so final validation cannot recompute the comparison.
Retaining them would change the 16/18-artifact evidence contract. That is recorded as the narrowed
DESIGN-GAP 6 for the RFC author.

## 2026-09-24 — Two-tree reproducibility under the corrected rules

**What was compared.** From clean `c333b17` (source time `2026-09-24T14:15:04Z`), two independent
`0.1.0-preview.4` candidate trees were built: `diag-candidate-a` and `diag-candidate-b`. Each tree
independently rebuilt:
- the client (fresh `client/dist` per tree);
- six Linux/amd64 commands;
- staged content;
- a no-cache gameserver image and its real Syft SBOM;
- metadata (the sourcemap-derived client inventory, R8);
- the assembled bundle.

**Results.** The complete bundle trees compare identical (`diff -r`), as do the binaries, client
output, metadata, image context, gameserver archive (config
`sha256:6eff9bb26876a29cbee33b893d4f83bebc03cece7c6b8b27a2268627c2ce459b`) and normalized
gameserver SBOM. The R3–R20 changes therefore preserve byte reproducibility, including the new
build-graph licence inventory.

**Limits (same as the earlier diagnostic build):**
- The six upstream SBOMs were reused rather than regenerated.
- Nothing was reviewed, and `release-builds/*.json` is not updated.
- This is not R-006.

## 2026-09-24 — R22: restart-during-admitted-work producer

**Process note:** there is no separate predeclaration commit. The scope is the DP-F predeclared
negative "gameserver restart during admitted work", which had no producer.

**Change:**
- `deploymentrelease.DockerRuntime.KillGameserver` sends `docker kill --signal=KILL`, observes
  readiness through the R9 rule (only the gameserver's own 503 counts) and inspects the exit code,
  then derives evidence with the release's `deriveDrainEvidence`. There is no courtesy or socket
  observation.
- The new rehearsal producer `restart-admitted-work` (CLI, bound to
  `gameserver_restart_during_admitted_work`):
  1. requires the exact candidate install row;
  2. kills the gameserver;
  3. always restores service (`Start` plus `AuthenticatedSmoke`);
  4. exits 3 only when the evidence is invalid, the exit is non-zero and service is restored.
  An accepted graceful drain exits 0, and an unrestored host exits 2.

**Evidence:**
- `TestKillGameserverIsNeverABoundedDrain`: a recording runner returns exit 137 and the proxy
  answers 502. The evidence is invalid, with no readiness, courtesy or intent refusal.
- `TestRestartDuringAdmittedWorkRequiresARejectedDrainAndRestoredService` covers: rejected with
  restore and smoke; a graceful drain accepted; an unrestored host not rejected; no install means
  no kill.
- Severing the producer classification and forging the kill derivation each fail their test; both
  were restored.
- `make test-go` over deploymentrelease, deploymentrehearsal and cmd/deployment-rehearsal passed.

**Coverage:** producer-backed populations are now 32 of 43. Still without a producer:
- key overlap (DESIGN-GAP 7);
- alert delivery (DESIGN-GAP 2);
- journal budget, which needs workload/max-use inputs the strict scenario schema does not carry
  (a contract addition);
- provider-off operation (a composite row with no single command defined).

No host run has happened.
## 2026-09-30 — Codex targeted review of Claude's R1 rollback correction

- **Review by:** Codex. **Recorded by:** Codex. **Reviewed range:** `1978660^..1978660`,
  limited to the `VerifyRestoreInputs` safety claim, its Docker fixture and DP5/AC5. This is
  not an approval or complete range review of R1–R22.
- **Decision:** **CHANGES REQUIRED (RP-133).** `Rollback` does call
  `VerifyRestoreInputs` before `StopFailed` and `ResetDatabase`, and the host check binds the
  envelope metadata/checksum, backup path and key-file permissions. But it never tests whether
  that well-formed key decrypts this backup. The later containerized restore is the first
  decryption attempt, after the database volume is removed. This leaves the primary safety
  property of R1 false for a plausible operator key mismatch.
- **Executed negative:** temporarily extended
  `TestDockerRuntimeVerifiesRestoreInputsWithoutRuntimeCommands` with a second valid age X25519
  identity, mode `0600`, for a backup encrypted to the first. Cold
  `make test-go GO_PACKAGES='./deploymentrelease'
  GO_TEST_FLAGS='-run TestDockerRuntimeVerifiesRestoreInputsWithoutRuntimeCommands -count=1 -v'`
  failed in 0.08 s with `wrong but well-formed identity accepted before destructive rollback:
  <nil>`. Removed the temporary test exactly; the original focused lane then passed cold.
  No product/test bytes from the probe remain.
- **Predeclared Codex correction (accepted DP5/AC5, no new mechanic):** before destructive
  rollback, parse exactly the identity population the real restore accepts (one age identity),
  authenticate and consume the backup with that identity while discarding plaintext, then retain
  the existing envelope/bundle checks. The positive case is a valid bound backup/key with zero
  runtime commands; negatives are a distinct valid key, malformed/multiple key file, and
  ciphertext whose untrusted header checksum is recomputed after tampering. These must all
  refuse before stop/volume removal. A severing probe that removes the new decryption call must
  make the wrong-key test fail. `Restore` remains the actual database restore and must still
  recheck its inputs. The preflight may cost an additional full backup read; do not claim the
  four-hour RTO until the clean-host R-006 run measures it. Codex's correction will require
  Claude's independent cross-party review; Codex cannot approve its own fix or archive the RFC.

## 2026-09-30 — RP-133 Codex rollback preflight correction, ready for Claude review

- **Scope:** `DockerRuntime.VerifyRestoreInputs` now parses exactly one age identity and runs
  the backup package's authenticated `Restore` to `io.Discard` before rollback's destructive
  steps. That existing package routine verifies full ciphertext length/hash, successful age
  decryption and the authenticated envelope fields; the returned header is then bound to the
  requested backup, server, previous manifest, epoch and pre-upgrade class. The actual restore
  still runs after the clean-volume reset. The canonical Deployment page states both boundaries.
- **Failing-first evidence:** the test extension initially failed cold on the old
  check with `wrong but well-formed identity accepted before destructive rollback: <nil>`.
  Its negatives also include malformed and multiple identity files and changed ciphertext with
  a recomputed outer SHA-256. The pre-change run demonstrates the wrong-key discriminator; no
  separate post-fix mutation run is claimed. The earlier ephemeral review probe was removed
  before this implementation.
- **Cold checks:** focused `TestDockerRuntimeVerifiesRestoreInputsWithoutRuntimeCommands` passed;
  full `make test-go GO_PACKAGES='./deploymentrelease' GO_TEST_FLAGS='-count=1'` passed; real
  Compose/Postgres/Caddy `make test-deployment-release` passed. `make vet` could not reach its
  package check because the sandbox denied creating the Go module-cache directory for
  `github.com/chromedp`; the focused Go test still ran its package-level default vet. No
  current-head `verify-push` claim: the independent RP-131 kernel history gate remains red.
- **Limits:** this authenticates and reads the whole backup before destruction, but does not
  prove `pg_restore` will accept its logical contents or that a production-sized backup meets
  the four-hour RTO. Those remain clean-host R-006 gates. This is Codex-authored implementation,
  not a designated verdict; Claude must review the exact correction range before it counts.

## 2026-09-30 — RP-133 correction commit coordinate

Codex's predeclaration and implementation span is `5f14a808^..d19b5d8a`. The code/test/docs
landing is `d19b5d8a`. This is the exact range for Claude's designated review; the R1
CHANGES REQUIRED finding and the independent RP-131 CI failure remain open.

## 2026-09-30 — RP-133 full Go vet environment resolution

The earlier `make vet` failure was a sandbox denial while populating Go's normal module cache,
before package analysis. With narrowly scoped approval for the exact root `make vet` target,
the full `go vet ./...` run exited 0. This replaces the environment-limited verification
status above; it does not supply Claude's designated code verdict, a clean-host rollback proof,
or a green `verify-push` while RP-131 remains.

## 2026-09-30 — R2 rotation/rollback interleaving probe predeclaration

- **Question:** can the rotation command change the ledger while `Controller.Rollback` holds its
  release lock, so `StopFailed` executes under a both-open overlay but `ResetDatabase` rereads a
  single-open ledger and refuses after the live service has stopped?
- **Population:** current `deploymentrelease` package, its real `acquireOperatorLock`,
  `ActivateRotation`/`RemovePrevious` ledger path, and `DockerRuntime` with a recording Docker
  runner (no real containers or database). Start with JWT and bootstrap overlaps both open.
- **Arms/controls:** hold the release-ledger lock; first show both-open `StopFailed` and
  `ResetDatabase` accept when the ledger stays unchanged. In the interleaving arm, run
  `StopFailed`, then legally remove previous JWT after its governed 30-minute overlap, leaving
  bootstrap open; ask `ResetDatabase` to continue. As a control, hold the rotation-ledger lock
  and show the same removal is refused, proving the probe actually exercises lock exclusion.
- **Fired condition:** rotation removal succeeds while the release lock is held, the recorded
  `down` occurs, and `ResetDatabase` refuses before its volume action. This establishes a
  possible stranded rollback interleaving and blocks R2's safe composition claim. A mere
  refusal before `down` is not the defect. The probe does not measure real operator timing or
  substitute for clean-host rollback; any repair must preserve independent key-overlap minima.

## 2026-09-30 — Codex targeted R2 review: interleaving fired (RP-134)

- **Review by:** Codex. **Recorded by:** Codex. **Reviewed range:** `92fab70^..92fab70`,
  limited to R2's rotation-overlay/rollback safety claim. This is not a full-range verdict on
  the Deployment RFC.
- **Decision:** **CHANGES REQUIRED.** The static both-open path works, and the existing
  `TestDockerRuntimeComposesRotationOverlayOnlyForAnOpenPair` passed cold. Temporarily removing
  the overlay append made that test fail, proving it catches the static omission. But release
  and rotation commands acquire different lock files. Each DockerRuntime method calls
  `normalized()` and rereads the rotation ledger, so a legal rotation between rollback steps
  changes which Compose overlay is usable after a destructive step.
- **Executed interleaving:** the predeclared temporary test activated both overlaps, held the
  release-ledger lock, and recorded a `StopFailed` `down` with `compose.rotation.yml` present.
  While that lock stayed held, `RemovePrevious` for JWT succeeded at the governed 30-minute
  minimum, leaving bootstrap open. `ResetDatabase` then returned `ErrInvalid` before its
  `docker volume rm` call. The unchanged-ledger control accepted both steps; holding the
  rotation-ledger lock blocked removal. The cold test exited 0 because it asserted the defect.
  The temporary test and static-overlay mutation were both removed exactly; product/test tree
  returned clean.
- **Scope boundary:** this proves a possible post-stop refusal in the recording-runner/lock
  composition, not a real clean-host outage or its frequency. It is distinct from the RFC
  author's existing DESIGN-GAP 4: whether a single-family open overlap should be independently
  composable. Even if the owner keeps fail-closed single-family behavior, rotation must not
  interleave between preflight and destructive release/rollback/recovery steps. A correction
  needs shared serialization or a governed operation-wide immutable overlay decision, plus a
  permanent failing-first interleaving witness and cross-party review. No R2 approval or
  archival is recorded.

## 2026-09-30 — RP-134 Codex correction predeclaration

- **Authority/scope:** accepted DP4/DP5 safe key overlap and stop-drain-start contract. Share
  the existing canonical `release-ledger.jsonl.lock` across Controller install/release/rollback,
  rotation activation/removal and both R-006 recovery producers for their entire operation.
  Rotation also retains its own ledger lock for backwards-compatible mutation exclusion.
  Lock acquisition is nonblocking and refuses before any Compose/volume command. No change to
  the current single-family fail-closed policy; DESIGN-GAP 4 stays with the RFC author.
- **Population/controls:** permanently test a both-open rotation under a held release lock.
  `StopFailed` and `ResetDatabase` must use the overlay; a governed JWT removal must fail while
  that lock is held, then succeed after release. Remove the shared-lock call to make the
  wrong-interleaving test fail. A second test exercises the recovery producer: at its stop
  boundary, a rotation command must refuse while the recovery operation lock is held. Existing
  rotation overlap-time tests and release/recovery suites remain green.
- **Limits:** this serializes commands using the supported helper and operator-state directory;
  it does not prove an external/manual edit of the append-only ledger is impossible. It does
  not resolve whether a deliberate single-family rotation should allow an upgrade. Clean-host
  R-006 and Claude's cross-party review remain mandatory.

## 2026-09-30 — RP-134 Codex shared-lock correction, ready for Claude review

- **Implementation:** the existing canonical `release-ledger.jsonl.lock` now spans the full
  Controller install/release/rollback operation, each rotation activation/removal, and the
  empty/populated recovery producers. Rotation retains its own ledger lock for mutation
  serialization. The shared lock is nonblocking, so a competing rotation refuses before
  changing the ledger; a helper operation then keeps one governed overlay state through its
  destructive steps. The Deployment page records this supported-helper boundary.
- **Fired tests:** `TestRotationCannotChangeBetweenRollbackStopAndReset` and
  `TestRecoveryHoldsOperatorLockAgainstRotationDuringStop` both failed cold on the previous
  code with a successful illegal rotation. They pass with the correction. The rollback test
  also checks activation and removal refusal while locked, successful rotation after unlock,
  and overlay preservation through `StopFailed`/`ResetDatabase`. Removing only
  `RemovePrevious`'s new shared-lock call made **both** tests fail; it was restored and both
  passed cold again. The initial diagnostic probe was removed exactly.
- **Cold gates:** full `./deploymentrelease ./deploymentrehearsal` Go packages passed;
  `make test-deployment-release` passed its real Compose/Postgres/Caddy integration;
  `make vet` exited 0. `make test-deployment-rehearsal` first failed before assertions when the
  sandbox denied an `httptest` loopback listener; the exact root target passed with narrowly
  scoped local-socket access, including its build-record validators. The retained bundles are
  still historical inputs, not rebuilt release candidates.
- **Limits:** direct manual edits to operator-state files are outside helper serialization;
  no clean-host rotation/recovery/rollback or production-sized RTO was proved. The known
  single-family-overlay DESIGN-GAP 4 remains an RFC-author choice. This Codex implementation
  needs Claude's designated review of its exact range and cannot be self-archived.

## 2026-09-30 — RP-134 correction commit coordinate

Codex's predeclaration and implementation span is `8e9b46d0^..2d8ca3f1`; the code/test/docs
landing is `2d8ca3f1`. Claude is the required designated reviewer for this exact range.
R2's original targeted CHANGES REQUIRED verdict is not silently converted to approval by the
existence of this unreviewed correction.

## 2026-09-30 — R3 image scanner legacy-tar probe predeclaration

- **Question:** R3 claims to recurse through plain tar members, but `isTarStream` recognizes
  only the `ustar` magic. Can a valid Go-readable V7 tar member hide a gzip-compressed secret
  because the scanner treats the whole tar as opaque bytes?
- **Population/arms:** a Docker-archive-shaped outer tar containing one layer tar with a
  nested `.gz` file and a generated secret sentinel. The USTAR inner tar is the positive
  control. For the V7 arm, clear only the inner header's USTAR magic/version, recompute its
  checksum, and require Go's `archive/tar.Reader` to accept the header and file before asking
  `ScanDockerArchive` for a finding. A corrupted-checksum arm must not be mislabeled valid.
- **Fired criterion:** USTAR yields a finding, V7 is Go-readable but the scanner returns no
  finding and no error. This would be a parser-recognition gap in R3's stated recursive-tar
  contract. The probe does not establish that the current BuildKit image emits V7 or that
  Docker's runtime accepts this specific archive; those claims require separate evidence.

## 2026-09-30 — R3 legacy-tar probe fired; correction predeclared

- **Executed result:** the first gzip fixture was invalid because its raw compressed bytes
  retained the sentinel, so a scanner that did no recursion could still report a finding.
  Rebuilt with Go gzip best compression and a raw-byte absence assertion. With that control,
  the USTAR layer produced a finding; the modified V7 header failed Go tar parsing until its
  checksum was recomputed; the checksummed V7 header and gzip payload passed Go tar parsing;
  `ScanDockerArchive` returned `findings=[] err=<nil>`. The predeclared criterion fired.
- **Scope:** this demonstrates an R3 parser-recognition gap for a Go-readable nested tar,
  not a claim about current BuildKit output or Docker runtime support for V7. Under accepted
  DP2/AC2, retain the permanent USTAR/V7/malformed-header controls; recognize a valid
  non-USTAR tar via Go's tar reader while preserving the current fail-closed path for
  malformed USTAR. The old recognizer must fail the new V7 test; unrelated source/image
  secret fixtures and a clean image must still pass. Current bundles remain unrehearsed.

## 2026-09-30 — Codex targeted R3 finding and local scanner correction (RP-135)

- **Review by:** Codex. **Recorded by:** Codex. **Reviewed Claude range:**
  `2b6c0cb4^..2b6c0cb4`, limited to R3's recursive image-layer scan claim.
  **Decision:** **CHANGES REQUIRED.** R3's existing gzip/plain/nested fixtures passed cold,
  and reverting its recursion to raw outer-member scanning made the scanner test and the
  seeded-image rehearsal probe fail. That proves its principal fix is real, but the
  predeclared V7 probe established a missed Go-readable tar family carrying a nested gzip
  secret; the accepted DP2/AC2 scanner may not return a false clean result for it.
- **Codex correction:** keep USTAR as a fail-closed tar candidate even with a bad header;
  otherwise ask Go's tar reader whether the first header is valid, recognizing magic-less
  V7 without treating arbitrary bytes as tar. The permanent test has a USTAR positive, a
  V7 header rejected until its checksum is recomputed, a Go-readable V7 payload, and an
  assertion that the compressed bytes do **not** expose the sentinel in raw form. The first
  version of this fixture was invalid because default gzip did expose the sentinel; that
  driver error was corrected before drawing the V7 conclusion.
- **Discrimination and checks:** the permanent test failed on the prior recognizer with
  `findings=[] err=<nil>`, passed on the correction, failed again when the old recognizer was
  temporarily restored, and passed after restoration. Cold full `./releasepackage` and
  `./deploymentrehearsal` Go packages passed; `make release-secret-scan` scanned 1776
  tracked files with no finding; `make vet` exited 0. No current image archive or R-006 bundle
  was scanned in this range. Claude must designated-review the Codex correction; Codex does
  not approve its own implementation or the whole Deployment span.

## 2026-09-30 — RP-135 correction commit coordinate

The predeclaration/result record `d871339d..432624c0` and code/test/docs landing
`a7ab2640` form Codex's exact RP-135 correction span `d871339d^..a7ab2640` for Claude's
designated review. The original R3 targeted CHANGES REQUIRED finding remains until a reviewed
correction closes it; no R-006 bundle is validated by these tests.

## 2026-09-30 — RP-135 post-commit rehearsal gate

At product source `a7ab2640`, the exact root `make test-deployment-rehearsal` target passed
with its local loopback listener permitted: cold `deploymentrehearsal`, both rehearsal/release
command packages and `deploymentrelease`, followed by validation of the retained previous and
candidate build *records*. Those historical records passing structural validation does not
rebuild or validate their old bundles as current R-006 inputs. RP-131 still prevents a green
full `verify-push` claim.

## 2026-10-04 — Codex designated review of Claude R4 manifest cross-binding

- **Review by:** Codex. **Recorded by:** Codex. **Reviewed range:** `751717f4^..751717f4`
  only (R4). **Decision:** **APPROVED** for this bounded range, not for R1–R22, DP-B/DP-F,
  archival, or a release claim.
- Inspected the full R4 diff, accepted R4 predeclaration, `ValidateBundle`, the manifest's
  image-name/order validation, Compose service validation, runtime closure derivation, exact
  staged-file validation, and the matching canonical Deployment documentation. The migration
  claim is explicitly runtime-bound and was not promoted to a bundle-level proof.
- Cold root `make test-go GO_PACKAGES='./releasepackage'
  GO_TEST_FLAGS='-run TestValidateBundleBindsManifestClaimsToComposeAndContent -count=1 -v'`
  passed all seven forged-claim cases; cold whole-package `-count=1` passed after restoration.
  A temporary severing probe removed both new `ValidateBundle` bindings. All seven forgeries
  then passed validation and the focused test failed on every case with `forged manifest claim
  accepted: <nil>`. The exact bindings were restored; `git diff --exit-code` confirmed no
  product/test residue. A scratch-worktree creation was denied by local `.git` permissions,
  so the reversible probe ran in the primary checkout instead.
- This proves the seven checked R4 claims discriminate at package level. It does not validate
  retained historical bundles on this host, the runtime database-migration check, clean-host
  R-006, or any other Deployment batch. Prior R1–R3 CHANGES REQUIRED findings and Codex
  correction ranges awaiting Claude review remain unchanged.

## 2026-10-04 — R5 execution boundary and targeted R6 review finding

- **R5 review status:** no verdict yet on `78ec39a7^..78ec39a7`. Codex inspected the F1/F2/F3
  source, tests, templates and docs. The cold Go population for `operations`, `releasepackage`,
  `deploymentrelease`, `deploymentrehearsal` and `cmd/deployment-operations` passed after granting
  local loopback for `httptest`. The root `make test-deployment-operations` stopped before
  `promtool` could run: Docker pulled the pinned Prometheus image then returned `exec format
  error`. The host is `arm64`, Docker reports `aarch64`, and the pulled image inspects as
  `amd64/linux`. The composed cleanup-firing, rejecting-receiver and Caddy-peer witnesses were
  not executed here. Neither a failure nor an approval of R5's product behavior is inferred.
- **Review by:** Codex. **Recorded by:** Codex. **Reviewed range:** `1833882c^..1833882c`,
  limited to R6's schedule precheck and checksum evidence. **Decision:** **CHANGES REQUIRED
  (RP-136)** for the checksum witness; this is not a full Deployment range verdict. Cold root
  `make test-go GO_PACKAGES='./deploymentbackup ./cmd/deployment-backup'
  GO_TEST_FLAGS='-count=1'` passed. Temporarily severing the schedule's foreign-entry branch
  made the `lost+found` subtest fail because it created a backup; source was restored and the
  package reran green. But both R6 corruption fixtures changed encrypted payload length.
  Temporarily removing only `ReadHeader`'s SHA comparison left both new checksum subtests
  green, proving the test did not discriminate that check. The SHA comparison was restored;
  `git diff --exit-code` confirmed no product/test residue.
- **Predeclared RP-136 correction under accepted DP6:** keep production code unchanged.
  Replace the substituted dump with an equal-length alternate; assert encrypted payload byte
  lengths match and contents differ before writing the forged envelope. The existing rejection
  assertions for `ReadHeader` and `Restore` must pass cold on intact code and fail separately
  when each method's SHA comparison alone is removed. The truncated-payload control remains.
  This proves a bounded checksum branch, not an actual Postgres restore or clean-host R-006.

## 2026-10-04 — RP-136 test-only correction, ready for Claude review

- Changed the substituted dump to equal-length alternate plaintext and asserted the encrypted
  payloads have exactly equal byte lengths, differ in bytes, and match the original header's
  `payload_bytes`. No production code or shipped documentation changed.
- The focused test passed cold on intact code. Removing only `ReadHeader`'s SHA comparison made
  its swapped-payload subtest fail with `substituted payload header accepted: <nil>`; restoring
  that check and removing only `Restore`'s SHA comparison made it fail with `substituted payload
  restored ...: <nil>`. Both changes were restored exactly; the whole backup/command packages
  passed cold, and `make vet GO_PACKAGES='./deploymentbackup ./cmd/deployment-backup'` passed.
- The original Claude R6 range remains CHANGES REQUIRED until this Codex test correction gets
  Claude's designated verdict. Do not treat the local first filter or these package tests as
  a full R6/Deployment acceptance or as clean-host backup/restore proof.
- Codex's predeclaration/finding and test-correction commits are `934420f4` and `ae69a64c`;
  Claude's designated review should cite the exact correction range
  `934420f4^..ae69a64c`. This record commit is coordination only, not a new test or approval.

## 2026-10-04 — Codex designated review of Claude R7 proxy/origin witness

- **Review by:** Codex. **Recorded by:** Codex. **Reviewed range:** `63bed7dd^..63bed7dd`
  only. **Decision:** **APPROVED** for the bounded R7 witness and config-row range, not for
  DP-A as a whole, a public Caddy deployment, R-006 or release.
- Inspected the added production-composition/Postgres test, the decoder negatives and the
  corresponding canonical gameserver wording. Cold root
  `make test-go GO_PACKAGES='./deploymentconfig' GO_TEST_FLAGS='-count=1'` passed.
  The first declared Postgres invocation failed before tests because the local cached
  `postgres:16-alpine` image was `amd64` on an `arm64` host. Refreshing that existing test tag
  for `linux/arm64` (no repository or release-manifest edit) made the declared
  `make test-save-integration SAVE_TEST_PACKAGES='./gameserver'
  SAVE_TEST_FLAGS='-run TestComposedProductionBoundary -v'` pass cold without a skip.
- First proxy-hop severing attempt made `trustedProxyHops` unused and failed compilation;
  that was a probe error, not evidence. The compiling severing variant forced the forwarded
  hop count to zero and failed with `a second forwarded client inherited the first client's
  limit`. A separate origin severing variant supplied a different allowed origin and failed
  with configured-origin upgrade `status=403` instead of 101. Each was restored, and the
  Postgres witness passed again; `git diff --exit-code` confirmed no product/test residue.
  Only the R7 file range is approved. The R7 predeclaration was not a separate prior commit,
  as the original implementation log already discloses; the executed review did not erase
  that provenance limitation.

## 2026-10-04 — R8 packaging-rights probe predeclaration (RP-137)

- **Review by:** Codex. **Recorded by:** Codex. **Target range:** Claude R8
  `1f048de9^..1f048de9`, specifically the claim that the new source-map extractor cannot
  under-attribute a shipped JavaScript package. This entry is a predeclared test, not a verdict.
- Current-source positive: `make build-client` and `make generate-release-metadata` on the fresh
  build produced 53 dependencies, including `pad-end@1.0.2` and `tslib@2.8.1` (0BSD). Current
  `dist` contains two JavaScript assets and matching maps. That does not exercise a partial
  source-map population.
- **Negative population:** a fixture `dist` with one mapped JavaScript asset and a second
  emitted JavaScript asset whose `sourceMappingURL` names a missing map. The existing mapped
  asset carries one recognized npm package; the unmapped asset imports a distinct package.
  `discoverClientDependencies` must reject the build, not return the mapped subset. A companion
  case has an orphan `.map` without its referenced JavaScript asset and must also refuse; a
  complete pair is the positive control. Run the new negative on current code first.
- **Bounded correction if fired:** validate a one-to-one, local `sourceMappingURL` ↔ map pairing
  for every emitted `.js` asset before reading package sources. Do not change actual release
  content or license-policy classification. A severing probe that removes the pairing gate must
  make the missing-map test fail. This proves JavaScript map completeness only; CSS-only
  package attribution is not silently claimed and needs a separate build-graph route before a
  full client-rights closure.

## 2026-10-04 — R8 targeted finding and RP-137 correction

- **Review by:** Codex. **Recorded by:** Codex. **Reviewed range:** `1f048de9^..1f048de9`,
  limited to the JavaScript source-map completeness claim. **Decision:** **CHANGES REQUIRED
  (RP-137)**. The predeclared missing-second-map test failed first on the original code:
  `emitted JavaScript asset without its map accepted: <nil>`. The fresh current build had two
  JavaScript assets with matching maps, so this is an acceptance-set hole, not evidence of a
  missing notice in the present generated output.
- Codex's accepted-DP2 correction validates every emitted `.js`/`.mjs`/`.cjs` asset has its
  sibling map and one exact trailing source-map link. Every map must point back to an existing
  linked JavaScript or CSS asset; symlinks in the build tree refuse. The fixture also rejects
  orphan maps and wrong map links. This does not infer CSS-only package completeness from
  absent CSS maps; canonical Deployment docs now state that limit.
- Cold `make test-go` over `cmd/gen-release-metadata` and `releasepackage`, plus focused vet,
  passed. A fresh root `make build-client` followed by real `make generate-release-metadata`
  still emitted 53 dependencies with the five npm packages; `pad-end@1.0.2` and
  `tslib@2.8.1` (0BSD) were present. Temporarily removing only the new every-script map gate
  made the missing-map test fail with the same acceptance diagnostic; it was restored.
- This is Codex implementation and requires Claude's designated exact-range verdict. It does
  not approve R8 as a whole, settle 0BSD policy beyond the accepted scope, prove CSS-only
  attribution, or validate a final release artifact/clean host.
- The predeclaration/finding commit is `7d9f690c` and the code/test/docs correction is
  `7931d7c5`; Claude's designated review should cite `7d9f690c^..7931d7c5` exactly.

## 2026-10-04 — R9 readiness provenance probe predeclaration (RP-138)

- **Review by:** Codex. **Recorded by:** Codex. **Target:** Claude R9
  `8ff555f0^..8ff555f0` and accepted DP5 readiness-down evidence. This is a predeclared
  probe, not an approval.
- R9's current `waitHTTPState(false)` sees `runtime.PublicOrigin+/readyz` through Caddy and
  accepts any 503. The gameserver `handleReady` returns 503 on not-ready, database ping failure
  and drain, with no distinct response marker. A status-only observation cannot prove the
  gameserver's own orderly drain answer, despite the R9 canonical wording.
- **Negative population:** an HTTP 503 without a drain marker, representing either a proxy
  failure or gameserver non-drain unavailability, must not satisfy readiness-down. A genuine
  draining gameserver response with an exact marker must satisfy it; a ready 204 must still
  satisfy readiness-up. Run the new negative on current code first. The marker must be emitted
  only while the gameserver drain gate is active, not on generic DB/readiness failure.
- **Correction if fired:** under DP5, add a fixed, non-secret response marker to the gameserver's
  draining `/readyz` branch and require it in the Caddy-facing down poll. Verify marker absence
  on non-drain 503 in a server test. Severing the observer marker check must fail the no-marker
  test, and severing the producer marker must fail its server test. This is observer evidence,
  not a new player mechanic. A local package pass does not substitute for the exact amd64
  Caddy/Postgres and clean-host R-006 populations.

## 2026-10-04 — R9 targeted finding and RP-138 observer correction

- **Review by:** Codex. **Recorded by:** Codex. **Reviewed range:** `8ff555f0^..8ff555f0`,
  limited to the readiness-down provenance claim. **Decision:** **CHANGES REQUIRED (RP-138)**.
  The predeclared unmarked-503 fixture failed first: `status 503 counted as the gameserver
  withdrawing readiness`. A separate server test failed because the orderly-drain response
  lacked its marker. R9's code distinguished status 503 from 502/504 but not a proxy or
  non-drain gameserver 503.
- Codex's accepted-DP5 correction shares the fixed header contract through
  `deploymentconfig`. `gameserver.handleReady` emits it only while draining; ordinary
  not-ready/database 503s remain unmarked. The Caddy-facing `waitHTTPState(false)` requires
  both 503 and the exact marker. Its HTTP test covers unmarked 503, 502, 504, 204, marked 503
  and vanished upstream; server tests cover marked drain and unmarked database failure.
  Canonical gameserver and Deployment docs were updated with the narrowed claim.
- Cold root tests over `deploymentconfig`, `gameserver` and `deploymentrelease`, plus focused
  vet, passed. A temporary wrong producer marker failed the server test; removing the consumer
  discrimination failed the unmarked-503 test. Both probes were restored exactly. The first
  R7-style attempt to remove the producer line would have left an unused import, so the
  executed producer severing used a compiling wrong marker instead.
- This is not an approval of R9, the full Deployment range, or the exact Caddy/Postgres path.
  `compose.deployment-release-test.yml` pins its services and test binary to `linux/amd64`;
  this `arm64` Docker host has no working amd64 emulation (`exec format error` observed on the
  earlier pinned Prometheus lane). The real composed drain witness must be run on an amd64
  runner before R9/release closeout. Claude must designated-review Codex's correction range.
- Codex's predeclaration/finding and implementation commits are `c474e0c3` and `c9c9dde7`;
  Claude's designated review should cite the exact correction range
  `c474e0c3^..c9c9dde7`. This routing record is not a new test or approval.

## 2026-10-04 — Codex targeted R10 review: host integrity is not restore authority

- **Review by:** Codex. **Recorded by:** Codex. **Reviewed range:** `1f376ccc^..1f376ccc`,
  limited to the host-envelope preflight claim and its new witness. **Decision:** **CHANGES
  REQUIRED (RP-139)** for the canonical claim, not for the checksum mechanism. The new
  `ReadHeader` call checks real host bytes and rejects a misreported header/corrupt payload;
  cold `deploymentrelease` and `deploymentrehearsal` packages pass. Temporarily substituting
  the reported header for the host re-read made the differing-header test fail with
  `reported header differing from the host envelope accepted: <nil>`. The call was restored,
  `git diff --exit-code` confirmed no code/test residue, and both packages reran green.
- The positive R10 fixture configures an invalid restore identity string while a separate
  generated recipient encrypts the backup; `CreatePreUpgradeBackup` still succeeds. This is
  consistent with DP6 keeping the identity off-host until explicit restore, but contradicts
  the R10 documentation's broad “rather than proceeding without a working rollback” claim.
  Host checksum/header agreement cannot prove recipient-key match or Postgres restoratability.
- Codex corrected only `docs/deployment.md` to state the actual host-byte guarantee and the
  unproved decrypt/restore boundary. The new D-019 owner decision frames what per-release
  recoverability proof is required before a supported release; no extra key mounting or
  release mechanic was invented. The actual clean-host R-006 recovery remains mandatory.
  This documentation/decision correction requires Claude's designated review. R10/full
  Deployment is not approved or archived by the package tests.
- The Codex finding/decision route and canonical-doc correction are commits `c576f8f4` and
  `93ec2c10`; Claude's designated review should cite `c576f8f4^..93ec2c10` exactly. This
  routing record is not a new acceptance result.

## 2026-10-04 — Codex targeted R12 review: producer basename is not executable identity

- **Review by:** Codex. **Recorded by:** Codex. **Reviewed range:** `4a6c8f27^..4a6c8f27`,
  limited to R12's assertion that the plan is bound to real rehearsal producers. **Decision:**
  **CHANGES REQUIRED (RP-140)**. Cold `make test-go` over `deploymentrehearsal` and
  `cmd/deployment-rehearsal` passed. The focused
  `TestExecutionPlanRunsEveryExactPositiveAndNegativeCheck` passed cold too, while its
  `rehearsalFixtureTool` writes a shell script with basename `deployment-rehearsal` that
  returns 0 or 3 on demand. `boundToProducer` checks absolute path, basename and subcommand,
  but no binary hash or relation to the candidate bundle; `ExecutePlan` executes that path.
  Thus the same API that rejects `/usr/bin/true` accepts a renamed constant producer.
- The R12 exit-3 separation is useful and retained: generic CLI failures cannot satisfy
  negative rows. The finding is narrower—an executable chosen by the plan can itself emit
  those exits. `CommandSHA256` and the plan hash bind only command strings, not executable
  bytes or the release manifest's `deployment-rehearsal` artifact.
- **DESIGN-GAP 8 for the DP-F author:** specify the exact authority binding before any
  corrective implementation: how `run-plan` receives the candidate bundle, verifies its
  manifest hash against the plan, and requires the executable to be the bundle's hash-checked
  `deployment-rehearsal` rather than an arbitrary same-basename path. A negative using a
  fake executable with the correct basename/subcommands must fail before any plan row runs;
  the genuine bundle tool must pass structure and fail loudly only on currently unproduced
  populations. This is not license to add an unruled plan schema/path contract. No R-006
  evidence or R12 approval is claimed from the current passing fixture.

## 2026-10-04 — Codex designated review of Claude R13 recovery/install safety

- **Review by:** Codex. **Recorded by:** Codex. **Reviewed range:** `0ce7508e^..0ce7508e`
  only. **Decision:** **APPROVED** for R13's bounded smoke-dependency and initial-install
  volume-preflight correction, not for DP-F/R-006, a real recovery, or a complete Deployment
  range union.
- Inspected the full R13 diff, accepted DP5/DP-F scope, five named volumes in the actual
  release Compose template, recovery `StartRecoveryCore` call and canonical Deployment docs.
  Cold root `make test-go GO_PACKAGES='./deploymentrelease ./deploymentrehearsal'
  GO_TEST_FLAGS='-count=1'` passed. The exact live Docker filter
  `docker volume ls --quiet --filter=name=^cloud-clicker_` returned this project's existing
  cache volumes, so the prefix filter is not merely a recording-runner fiction.
- Removing Alertmanager from `smokeDependencies` failed
  `TestRecoveryCoreStartsEverySmokeDependency` with `recovery start omits smoke dependency
  alertmanager`. Narrowing the volume check to its Postgres slice failed the occupied-host
  test with `retained certificate volume ... occupied host accepted: <nil>`. Both mutations
  were restored; `git diff --exit-code` showed no product/test residue and both packages
  passed cold again.
- The evidence proves the configured Compose preflight and helper command composition; it
  does not demonstrate a host with real retained ACME data, a timed recovery, or the final
  clean-host install/rollback. Those remain R-006 and cross-party range gates.

## 2026-10-04 — Codex targeted R14 review: previous-build binding witness

- **Review by:** Codex. **Recorded by:** Codex. **Reviewed range:** `ee551f13^..ee551f13`
  for Claude's R14 bundle-to-build binding. **Decision:** **APPROVED for the bounded code
  binding**, after the supplemental previous-record witness below; not approval of a
  reproducible real-bundle run, DP-F, R-006, or the full Deployment range.
- Inspected the complete R14 diff, `BuildRecord` validation, manifest decoding, and the
  bundle validator's artifact/Compose/SBOM bindings. Cold root
  `make test-go GO_PACKAGES='./deploymentrehearsal ./cmd/deployment-rehearsal'
  GO_TEST_FLAGS='-count=1'` passed. The R14 negative fixtures initially covered candidate
  records only, although the predeclaration required both roles. Temporarily removing the
  previous-record `bindBuildToBundle` call left those tests green. The call was restored.
- Codex added a self-consistent wrong archive/rebuild-archive pair for the previous record
  in `460cbfd8`. With the previous binding temporarily severed, the focused cold test then
  failed specifically at `previous build record not bound to previous bundle bytes: <nil>`;
  after restoring the call, both packages passed cold. The production source has no
  residual probe mutation. This correction is a Codex-authored first-filter, not its own
  designated cross-party approval; Claude must review the exact correction range.
- The predeclared real supply-chain run was later described against diagnostic rebuilt
  bundles, but those bundles and build records are not retained as reproducible release
  inputs; the tracked `release-builds/*.json` still refer to invalid v5 bundles. This
  review therefore accepts the local binding behavior only. It does not convert the
  historical diagnostic account or operator-attested rebuild flags into release proof.

## 2026-10-04 — Codex targeted R15 review: semantic mutation fixture

- **Review by:** Codex. **Recorded by:** Codex. **Reviewed range:** `dda9f41f^..dda9f41f`
  for Claude's R15 probe change. **Decision:** **APPROVED for the bounded probe-preparation
  mechanism** after a supplemental test; not approval of each real semantic rejection or
  the complete DP-F/R-006 gate.
- Cold root `make test-go GO_PACKAGES='./deploymentrehearsal ./cmd/deployment-rehearsal'
  GO_TEST_FLAGS='-count=1'` passed. Inspected hardlink isolation: both the mutated target
  and manifest are unlinked before rewrite, so the retained candidate is not changed.
  The clean baseline executes before the mutation; a setup failure cannot become a
  negative result. The diagnostic rebuilt-bundle run in the earlier log reported real
  semantic rejections, but its bundle bytes are not retained here for an independent
  rerun.
- The original focused unit test still passed when Codex temporarily removed
  `rebindManifest`: its fake validator rejected on the second call without inspecting
  manifest hashes. Codex added a fixed sentinel artifact and assertions for the
  regenerated artifact hashes and mutated image-SBOM hash in `cad999f4`. Repeating that
  severing then failed seven negative subtests at `untouched artifact not rebound`.
  The production call was restored, and both packages passed cold again. The correction
  is Codex-authored and awaits Claude's designated exact-range review; this record is
  not a self-approval of that commit.

## 2026-10-04 — Codex designated review of Claude R16 envelope commit verification

- **Review by:** Codex. **Recorded by:** Codex. **Reviewed range:** `e6c58afc^..e6c58afc`
  only. **Decision:** **APPROVED** for R16's precommit envelope re-read and directory-sync
  implementation, not for key ownership, a successful restore, DP-F or the release floor.
- Inspected the full R16 diff and `Create`/`ReadHeader` sequence. The temporary envelope is
  closed, then the on-disk header, payload length and checksum are re-read before rename;
  the directory is synced after rename. Cold root
  `make test-go GO_PACKAGES='./deploymentbackup ./cmd/deployment-backup'
  GO_TEST_FLAGS='-count=1'` passed. Removing the precommit `ReadHeader` call caused
  `TestCreateVerifiesTheCommittedEnvelopeBytes` to fail with a reported successful path
  for the torn envelope. The call was restored, `git diff --exit-code` was clean, and
  both packages reran cold green.
- A crash or filesystem fault between rename and successful directory sync is not
  simulated. The code returns an error if sync fails, but the final path may already
  exist and needs operator reconciliation. No durability or decryptability claim is
  inferred from this unit lane or the checksum alone; D-019 remains open.

## 2026-10-04 — Codex designated review of Claude R17 browser-manifest witness

- **Review by:** Codex. **Recorded by:** Codex. **Reviewed range:** `14a82f3c^..14a82f3c`
  only. **Decision:** **APPROVED** for R17's test-only base-run witness and its candid
  redundancy disclosure; not approval of the final-run seal table or DP-F/R-006.
- Cold root `make test-go GO_PACKAGES='./deploymentrehearsal'
  GO_TEST_FLAGS='-count=1 -run TestBaseRunRejectsRehashedOperatorAuthority'` passed. Removing
  the typed `browser_result` versus evidence manifest comparison made the new
  other-manifest subtest fail with `rehashed valid browser result for another manifest
  accepted: <nil>`. Restored.
- Removing the direct candidate-manifest byte comparison left its new subtest green:
  another binding rejects the forged candidate. This is defense in depth, not an
  independently witnessed check, exactly as Claude disclosed. Both mutations were
  restored, the focused population reran cold green, and `git diff --exit-code` is clean.
  The earlier log's follow-up to audit final-run table rows individually remains open.

## 2026-10-04 — Codex targeted R18 review: two negative producers are unwitnessed

- **Review by:** Codex. **Recorded by:** Codex. **Reviewed range:** `82a78b53^..82a78b53`
  for the migration and missing-rollback-input producers. **Decision:** **CHANGES REQUIRED
  (RP-141)** for those two claimed R-006 negatives. The epoch bundle mutation remains in
  the generic probe path; this verdict does not pronounce on every R18 path.
- Inspected the new `probe_release.go`, dispatch, tests and the earlier diagnostic-run
  record. No test invokes either `runIrreversibleMigrationProbe` or
  `runMissingPreviousInputProbe`; the `plan_test.go` rows only list their names. Codex
  temporarily replaced both dispatch arms with unconditional `ProbeRejected, nil`.
  Cold root `make test-go GO_PACKAGES='./deploymentrehearsal ./cmd/deployment-rehearsal'
  GO_TEST_FLAGS='-count=1'` still passed. The original dispatch was restored exactly.
- The 2026-09-24 log records exit 3 for both CLI probes on `diag-*` bundles, but neither
  bundle nor its build record is retained, so this review cannot replay those results.
  A fixture for each producer must show that intact release/rollback inputs reach the
  relevant gate, the named mutation is refused for the intended reason, and severing
  that producer makes a cold test fail. Until then the rows are not independently
  witnessed and may not be counted toward clean-host R-006 acceptance.

## 2026-10-04 — Codex designated review of Claude R19 lifecycle producers

- **Review by:** Codex. **Recorded by:** Codex. **Reviewed range:** `1a8375af^..1a8375af`
  only. **Decision:** **APPROVED** for R19's controller rollback linkage and bounded
  lifecycle producer/control-flow contract, not for a real host drain, database restore,
  full DP-F or R-006.
- Inspected the full R19 diff: Make/CLI routes call the production lifecycle functions;
  the plan names those routes; the controller's successful rollback row retains the
  rolled-back-from candidate; and the lifecycle result checks the exact install,
  release and rollback ledger predicates used by final validation. Cold root
  `make test-go GO_PACKAGES='./deploymentrelease ./deploymentrehearsal
  ./cmd/deployment-release ./cmd/deployment-rehearsal' GO_TEST_FLAGS='-count=1'` passed,
  as did `make verify-ci-topology` and its 13 negative controls.
- Reinstating the old successful-rollback field clearing failed both
  `TestSuccessfulRollbackRecordsTheReleaseItRolledBackFrom` and
  `TestLifecycleProducersRunInstallReleaseAndRollbackInContractOrder`. Omitting the
  lifecycle release transition failed the latter at its exact two-row check. Both
  mutations were restored; all four Go packages reran cold green and `git diff
  --exit-code` was clean.
- The unit fixture uses a fake runtime and minimal manifests; it proves controller
  order/ledger binding but not Caddy drain frames, Postgres data restore, timing,
  Docker image availability or an operator-run release. R12 producer identity and
  R18 negative-producer witnesses remain open, so no R-006 row is promoted here.

## 2026-10-04 — R20 review predeclaration: exact non-clean refusal provenance

- **Question:** can `IsNonCleanRestoreRefusal` classify an unrelated command failure as
  the required dirty-database refusal merely because its error text contains the
  substring `"error_class":"non_clean_target"`? Source currently uses `strings.Contains`.
- **Population and controls:** the production `ExecRunner` failure shape from a Docker
  restore command; a genuine backup CLI JSON error with `command=restore` and the exact
  non-clean class; wrong command/class, malformed or embedded JSON, and an arbitrary
  error containing that substring. The existing real-Postgres dirty-target integration
  test remains the database-side control.
- **Criterion:** the adversarial arbitrary-error case must fail on current code; after a
  correction, only a typed failed Docker command carrying the exact structured backup
  CLI restore refusal may return true. The genuine class must still pass, while wrong
  classes and identity changes stay errors. A temporary severing of the parser must
  fail the permanent test. This authorizes only error-provenance classification, not
  a claim that the R20 host probe or R-006 has run.

## 2026-10-04 — Codex targeted R20 review and local refusal correction

- **Review by:** Codex. **Recorded by:** Codex. **Reviewed range:** `680b7bb1^..680b7bb1`
  for Claude's R20. **Decision:** **CHANGES REQUIRED (RP-142)** for error attribution;
  the real-Postgres and host producer outcomes are not approved by this review.
- The predeclared arbitrary-error fixture failed on original HEAD:
  `IsNonCleanRestoreRefusal` accepted an unrelated error echoing
  `"error_class":"non_clean_target"`. Codex's accepted-DP-F correction makes `ExecRunner`
  retain a typed failed executable, exit status and stderr, and requires one exact JSON
  backup-CLI `restore` error record from a failed Docker command. Tests reject another
  command/class, wrong exit, extra or embedded log text, and arbitrary error strings.
  The genuine structured class still drives the rehearsal negative; a real shell exit
  test confirms `ExecRunner` populates the failure fields.
- Restoring the substring path temporarily made the new arbitrary-error test fail
  again; it was removed. Cold root Go tests over `deploymentrelease`,
  `deploymentrehearsal` and both affected CLIs passed with `-count=1`; focused vet
  passed. The implementation and canonical-doc correction are `d3fde4c0`; Claude's
  designated review must cite `5f394f2a^..d3fde4c0` including the predeclaration.
- `make test-deployment-backup` could not start its pinned `linux/amd64` Postgres
  container here: it exited 255 on this arm64 host before tests ran. No Postgres pass
  is claimed. Strict parsing also needs an actual amd64 composed CLI run to show the
  backup log arrives as the expected single structured stderr record; any extra
  Compose stderr fails closed as setup rather than counting a false rejection.

## 2026-10-04 — Codex designated review of Claude R21 clean-host volume boundary

- **Review by:** Codex. **Recorded by:** Codex. **Reviewed range:** `a1da95c5^..a1da95c5`
  only. **Decision:** **APPROVED** for the bounded all-project-volume clean-start
  correction, not for a clean-host R-006 result or the broader host attestation.
- Inspected the full R21 diff. Cold `make test-go GO_PACKAGES='./deploymentrehearsal'
  GO_TEST_FLAGS='-count=1'` passed. Temporarily narrowing the returned-name check
  back to the Postgres volume left the retained certificate volume visible but made
  `TestHostObservationRejectsEveryUnsupportedOrDirtyBoundary/retained_certificate_volume`
  fail with `unsupported host accepted: <nil>`. The prefix check was restored and
  the package reran green; `git diff --exit-code` is clean.
- The exact `docker volume ls --quiet --filter=name=^cloud-clicker_` command on this
  daemon returned three existing project cache volumes, confirming the filter reaches
  non-Postgres project state. This machine is not a clean host and was not used as
  R-006 evidence. The R21 record's correction to DESIGN-GAP 6 is limited to source
  derivation; retaining raw identities for final re-verification remains an RFC-author
  question.

## 2026-10-04 — Codex targeted R22 review: no admitted work at restart

- **Review by:** Codex. **Recorded by:** Codex. **Reviewed range:** `b706e4df^..b706e4df`,
  limited to R22's claim to produce `gameserver_restart_during_admitted_work`.
  **Decision:** **CHANGES REQUIRED (RP-143)** for that R-006 population. The narrower
  SIGKILL-versus-graceful-drain classifier is not disputed or promoted to AC4 proof.
- Inspected the full R22 diff, `ProbeRestartDuringAdmittedWork`, `KillGameserver`, both
  new tests and the plan binding. The producer requires an install row, runs `docker
  kill`, reads readiness/exit, restarts and smokes. Neither its runtime interface nor
  its fake test contains an admitted intent, a held write, a receipt, or a post-restart
  data/result comparison. Cold root `make test-go GO_PACKAGES='./deploymentrelease
  ./deploymentrehearsal ./cmd/deployment-rehearsal' GO_TEST_FLAGS='-count=1'` passed,
  demonstrating the current tests can be green without that named population.
- RFC AC4 requires mid-write/restart fixtures to fail loudly; the existing composed
  release test holds an admitted intent during governed drain, but R22 does not use
  that fixture for an ungraceful kill. Canonical Deployment docs now explicitly limit
  R22 to crash classification and service recovery. **DESIGN-GAP 9 for DP-F author:**
  specify the exact admitted operation, pause point, allowed client outcome, durable
  post-restart state/receipt check, and severing case. Do not count the current exit 3
  as the AC4/R-006 row until that contract is implemented and executed on the bundle.

## 2026-10-04 — R22 fail-closed plan-route predeclaration

- **Question:** after RP-143, does the canonical execution plan still bind the named
  admitted-work row to a crash-only helper that can exit 3 without admitted work?
- **Population and negative control:** the `gameserver_restart_during_admitted_work`
  plan row with `restart-admitted-work`; the production `probe --population=...`
  fallback, which currently has no implementation and exits as unsupported. The
  existing fake-plan test is not an integration oracle because R12 accepts a
  same-basename constant executable.
- **Criterion:** a cold test fails on the current mapped crash-only producer. After
  correction, the plan cannot bind that command to the admitted-work row, and a real
  bundled CLI must fail loudly until DP-F defines and implements the actual fixture.
  Keep the crash-classification command available for diagnostics; do not rename the
  acceptance population or count it as complete. Severing the refusal route must
  fail the new test.

## 2026-10-04 — R22 fail-closed route implemented, actual mid-write proof still absent

- The predeclared `TestAdmittedWorkRestartRejectsCrashOnlyProducer` failed on original
  code with `crash classification accepted as an admitted-work restart witness`.
  Codex removed only the false `populationProducers` binding. The crash helper stays
  available as a diagnostic, but the named R-006 row now uses `probe`, which has no
  implementation for that population and exits as unsupported rather than claiming
  negative proof. `ValidateExecutionPlan` rejects an exact plan row that still tries
  to run `restart-admitted-work`.
- Cold root `make test-go GO_PACKAGES='./deploymentrehearsal
  ./cmd/deployment-rehearsal' GO_TEST_FLAGS='-count=1'` and focused vet passed.
  Temporarily reinserting the removed mapping failed the new test; it was restored.
  Canonical Deployment docs distinguish diagnostic crash classification from AC4.
- This is a Codex implementation correction under accepted DP-F, not Codex's own
  designated review. Claude must review `95d1bb60^..81a970ce` exactly. The R22
  CHANGES REQUIRED verdict remains for the absent admitted-write/post-restart witness;
  no host R-006 acceptance is claimed.

## 2026-10-04 — R8 CSS attribution diagnostic predeclaration

- **Question:** can the accepted DP1/AC8 client dependency inventory derive and verify
  packages shipped only through CSS, rather than silently treating the JavaScript
  source-map set as complete? The current `client/dist` has one CSS asset with no map.
- **Arms and controls:** enable Vite's CSS sourcemap option for one local build, then
  inspect whether the emitted CSS has a linked map and whether its `sources` identify
  the local imported stylesheet. Existing JS map/package inventory is the positive
  control. A separate fixture with an unmapped CSS asset must show that the current
  inventory accepts it, proving the under-attribution path.
- **Exit:** if the build emits a usable CSS map, specify a narrow every-CSS pairing
  gate and a CSS-only package fixture before implementation. If it does not, restore
  the diagnostic config edit and route a build-graph provenance requirement; do not
  claim R8 or AC8 closed. No released artifact or owner license policy is changed by
  this diagnostic.

## 2026-10-04 — R8 CSS diagnostic fired; build-graph provenance required (RP-164)

- `make build-client` passed with `css.devSourcemap: true` temporarily enabled, but
  `client/dist` still held one `.css` asset and no `.css.map`; the two JavaScript
  assets retained their maps. The diagnostic config edit was removed exactly.
- A temporary `TestDiagnosticClientInventoryRejectsUnmappedCSS` added a second,
  unmapped CSS asset to the existing two-package fixture and expected refusal.
  Cold root `make test-go GO_PACKAGES='./cmd/gen-release-metadata'
  GO_TEST_FLAGS='-count=1 -run TestDiagnosticClientInventoryRejectsUnmappedCSS'`
  failed: `unmapped CSS asset accepted: <nil>`. The diagnostic test was removed.
- Thus R8's current source-map inventory proves the JavaScript package set but not
  CSS-only dependencies. This is a potential under-attribution path, not evidence
  that today's emitted CSS contains an uncredited npm package: current source has
  only a local `./shell/styles.css` import. The accepted DP1/AC8 route needs a
  build-graph-derived CSS provenance producer, a fixture shipping a package only
  through CSS, a falsifying omission/severing case and exact artifact binding.
  Until then no full packaging-rights or AC8 claim is authorized.

## 2026-10-04 — R8 CSS graph correction predeclaration

- A temporary Vite `generateBundle` diagnostic showed the production entry chunk's
  `modules` includes `client/src/shell/styles.css` and 20 Svelte style virtual
  modules, while the emitted CSS still has no map. The diagnostic plugin was
  removed. This is a viable build-graph seam, not yet a rights witness.
- **Implementation scope:** emit a deterministic CSS provenance record from that
  hook: exact emitted CSS asset names/hashes and all contributing CSS module paths
  under `client/node_modules`. The release-metadata generator must require the
  record, compare it against every emitted CSS asset, resolve every package path
  to installed metadata/LICENSE, and union those packages with the JS map set.
- **Controls:** an intact production build still inventories the same five npm
  packages; a Vite-built fixture importing CSS from an npm package not used by JS
  must include its notice. Removing that CSS module from the graph, adding an
  unrecorded CSS asset, changing recorded CSS bytes, or removing the graph must
  fail. A temporary severing of the new graph-consumption guard must fail the
  CSS-only fixture. This proves the build/metadata boundary, not a clean-host
  bundle, license-policy adoption, or final AC8 release proof.

## 2026-10-04 — RP-164 local CSS attribution correction; cross-party review pending

- The first hook hashed CSS bytes during Vite `generateBundle`, but Vite changed
  the emitted CSS before writing it; a real `make generate-release-metadata`
  rejected the mismatch. The hook now records final CSS hashes in `writeBundle`,
  and the generator checks each hash against the actual dist file. This failed
  attempt is not counted as a successful witness.
- The first real Vite test imported package CSS from JavaScript and passed.
  A second fixture put `@import "style-only/theme.css"` inside a local
  stylesheet; it failed with `package_css_modules: []`. The hook now follows
  CSS `@import` dependencies, including from Svelte style sources, and refuses
  an import it cannot resolve or parse. Both Vite-built fixture arms pass.
- The Go generator requires the graph, exact emitted CSS asset set/final-byte
  hashes and installed package stylesheet/manifest/license. A fixture records
  a CSS-only npm package in the dependency list and delivered notice; removing
  the graph, adding an unrecorded CSS asset, changing CSS bytes or removing the
  package stylesheet each fails. Current client build plus real metadata
  generation still reports 53 dependencies (five npm packages, none CSS-only).
- Two temporary severing probes fired and were restored: omitting the Vite
  module-add operation made the real Vite CSS-only test fail at the expected
  package path; dropping the Go union made the notice fixture fail with only
  its two JavaScript packages. Cold full client tests (6910 pass, 83 skip),
  typecheck/Svelte check, root client build, focused Go metadata/releasepackage
  tests, vet, client boundary and CI-topology controls pass. The initial
  `@types/node` install attempt failed on unavailable registry DNS without
  changing the lockfile; the build hook and test use repo-standard `.mjs` with
  a narrow TypeScript declaration for the Vite config instead.
- `make verify-kernel-version` remains independently red at pushed
  `50a3a514` (RP-131). No full `verify-push`, clean-host bundle or AC8 success
  is claimed. This is Codex implementation under accepted DP1/AC8, requiring
  Claude's designated exact-range review. CSS URL asset/package provenance,
  worker-only CSS and final image/bundle rights remain later AC8 audit surfaces;
  this range proves package stylesheets imported into the main build graph.

## 2026-10-04 — RP-164 commit and review coordinate

The bounded implementation, tests, canonical docs and ledger landed together at
`2de5fedc`. Claude's designated review should cite the exact range
`2de5fedc^..2de5fedc`; this coordinate entry is bookkeeping, not a review
verdict. The new Vite fixture runs in `make test-client` (the client CI lane),
and the Go metadata fixture runs in `test-go-core` (the server CI lane). The
pre-existing pushed kernel-history failure still prevents calling aggregate
`verify-client`/`verify-push` green.

## 2026-10-04 — Codex cross-party review of unmerged Claude R-006 draft

- **Review by:** Codex. **Recorded by:** Codex. **Reviewed range:**
  `cd3235a7^..cd3235a7` on the clean Claude worktree branch
  `worktree-agent-a3833e3f9f9400b0b` (not merged into main).
  **Decision:** **CHANGES REQUIRED before acceptance or integration.** This is
  a draft-design verdict, not a verdict on Deployment implementation.
- Read the full 721-line proposed `deployment-r006-evidence-contract.md`, its
  index addition and log entry against accepted Deployment DP4–DP8 and the
  current R12/R18/R22 review records. It offers explicit owner options for
  seven original DESIGN-GAPs plus journal and provider-off producers, without
  pretending those options are accepted. That structure is useful and no
  product byte changed in the branch.
- Its summary and plan delta say the recommended options leave every R-006
  row with a producer. At current HEAD that cannot be true: RP-140/DESIGN-GAP
  8 still permits a same-basename fake plan executable, and RP-143/DESIGN-GAP
  9 leaves `gameserver_restart_during_admitted_work` intentionally unsupported
  after the crash-only route was removed. The draft also predates D-019's
  decryptable-restore authority choice and the R18 negative-producer witness
  finding. Accepting it now would create a false complete-evidence contract.
- Required revision: retain the owner-option format, reconcile the current
  plan population and artifact counts, add explicit DG8/DG9/D-019/R18 routes
  and decide which are in this amendment versus another accepted successor.
  The owner must select actual policy options before body reconciliation.
  Do not cherry-pick the draft as active authority or implement its
  recommendations in Marco's name. The worktree and its commit are preserved.

## 2026-10-04 — RP-164 follow-up predeclaration: CSS resource and worker boundaries

- **Question:** does the new main-build CSS graph omit npm material shipped
  only through a stylesheet `url(...)` resource or a Vite worker's stylesheet?
  The prior exact range explicitly left both surfaces open, so neither is
  counted as packaging-rights closure.
- **Population:** a real Vite fixture with a local CSS file referring to a
  small file inside an installed npm package via `url(...)`, and a separate
  real Vite worker fixture importing a package CSS file. Retain the direct
  and nested `@import` controls from RP-164. Inspect final emitted bytes and
  graph, not just source text; only count a case if the package file's bytes
  actually ship or are inlined into a shipped asset.
- **Firing criterion:** the build succeeds and ships the package bytes, but
  the graph lists no path in that package and the metadata inventory would
  omit its notice. A fixture that cannot build or whose bytes are not shipped
  is invalid, not a green or red rights result.
- **If fired:** under accepted DP1/AC8, make the producer and Go inventory
  capture the missing resource path, bind final asset bytes, and add a test
  that fails when that producer route is severed. If the Vite architecture
  cannot provide a trustworthy path for a case, fail closed rather than
  claiming complete client attribution. No clean-host/image AC8 proof or
  cross-party verdict follows from this local test.

## 2026-10-04 — RP-165 CSS URL probe fired; worker arm did not

- The URL fixture built with Vite and emitted a CSS data URI. The test decoded
  its base64 SVG and found the `fixture-logo` sentinel, proving the package
  bytes shipped; the first diagnostic mistakenly searched the encoded CSS
  text for raw sentinel bytes and was corrected before assessing attribution.
  With that control valid, the v1 graph still had no package path and the
  fixture failed at `expected [] to include node_modules/style-only/logo.svg`.
- The worker fixture also shipped `worker-style-sentinel`, but its generated
  JavaScript source map contained the package stylesheet path in `sources`.
  Therefore the metadata generator's existing map walk could attribute it;
  the predeclared omission criterion did **not** fire for worker CSS. A Go
  worker-map fixture now proves that such a source reaches the delivered
  notice. No worker-specific graph or unsupported-host claim was invented.
- A bounded v2 CSS graph adds `package_css_assets`. The build hook follows
  stylesheet `url()` references to installed package files; the Go generator
  requires v2, resolves the new paths to package manifest/LICENSE and unions
  them with stylesheet and JavaScript-map packages. An asset-only package
  reaches delivered notices; a missing asset and a stale v1 graph reject.
  The final emitted CSS hash still binds the inlined asset bytes. The current
  product build emits no package URL resource; real metadata generation still
  reports 53 dependencies.
- Temporary severing of `resources.add` failed the Vite URL fixture at the
  named asset path; severing the Go v2 asset-path loop failed the asset-only
  notice fixture. Both were restored. The full client suite (6912 pass, 83
  skip) and cold focused Go package pass before final gate aggregation.
  This is local Codex implementation under accepted DP1/AC8, not cross-party
  approval, a final bundle/image rights proof, or R-006 evidence.

## 2026-10-04 — Correction to worker arm: raw map text was a false oracle (RP-166)

- The previous entry's worker conclusion was wrong. Its first check searched
  raw `.js.map` bytes for `style-only/theme.css`; that string came from the
  worker source's import text, not from the map's `sources` array. The stronger
  final gate parsed `sources` and failed: the worker CSS sentinel shipped in
  an emitted asset, but no package stylesheet was listed in either the main
  graph or the map sources. This is RP-166, a real under-attribution route.
- A diagnostic `worker.plugins` hook printed the worker bundle's module IDs
  and showed the exact `node_modules/style-only/theme.css` file. The permanent
  hook now shares package-module/resource sets between the main and worker
  builds; Vite config installs the worker hook explicitly. The real worker
  fixture asserts shipped sentinel bytes, map-source absence and graph
  inclusion. A Go fixture has a worker JavaScript map with only `worker.ts`
  in `sources` and proves the worker CSS package still reaches notices through
  the graph. The diagnostic print hook was removed.
- This correction retracts the prior worker-map claim in the RP-165 entry and
  in its initial planning/roadmap text; canonical docs, backlog, plan and
  current queue now use the executed stronger result. No clean-host or full
  AC8 claim follows, and Claude's designated review remains required.

## 2026-10-04 — RP-166 worker provenance severing controls

- Removing the worker hook's `scanBundle` call (while leaving the main hook
  intact) made the real Vite worker fixture fail exactly at the absent
  `node_modules/style-only/theme.css` graph path. The shipped-byte and
  source-map-absence controls passed first. Restoring the hook restores the
  fixture.
- Separately removing the Go stylesheet-module consumption loop made
  `TestClientInventoryIncludesWorkerCSSOmittedByJavaScriptMap` fail at the
  delivered-notice assertion. The original loop was restored. These probes
  discriminate the worker producer and its metadata consumer independently;
  they do not substitute for a final image or bundle audit.

## 2026-10-04 — RP-165/RP-166 final local gate and handoff

- `make test-client`: 6912 passed, 83 skipped; `make typecheck`: zero errors or
  warnings; cold focused `make test-go GO_PACKAGES='./cmd/gen-release-metadata
  ./releasepackage' GO_TEST_FLAGS='-count=1'`: both packages passed. `make vet`,
  `make verify-client-boundary` and `make verify-ci-topology` passed (13 topology
  negative controls rejected). `make build-client` emitted the v2 graph with
  one final-byte-hashed CSS asset and empty package CSS sets for the current
  product build; real `make generate-release-metadata` returned 53 dependencies.
- The first metadata invocation supplied an abbreviated source commit and was
  rejected as invalid release runtime content. Repeating with the full
  40-character HEAD succeeded; this was an invocation error, not a dependency
  inventory failure.
- The first cold `make test-go-core` run could not bind existing `httptest`
  loopback listeners in the restricted sandbox (`operation not permitted` in
  deployment operations/release/alert tests). The exact same target reran with
  narrowly scoped local-loopback permission and passed every package cold.
  This is not a claim that the aggregate push gate is green: the pushed
  historical kernel-version guard remains red, and the final bundle/image
  and clean-host AC8/R-006 evidence is absent.
- This Codex batch is ready for Claude's designated independent review over
  its exact committed range. No RFC status, archival or release promotion is
  authorized by these local gates.

## 2026-10-04 — RP-165/RP-166 review-range coordinate

- Implementation, fixtures, docs and the preceding gate record landed in
  `2ea9f316` (base `7163fe8f`). The designated reviewer must cover
  `7163fe8f..2ea9f316` plus this coordinate commit, inspect the actual diff,
  rerun the relevant gates and record `Review by:` and `Recorded by:` with
  the exact inspected range. Codex's local probes are a first filter only.

## 2026-10-04 — Current-HEAD Deployment review resumption (no verdict)

- Re-read the current execution queue before treating the older R1–R17
  handoff as pending wholesale: bounded Codex designated verdicts and
  CHANGES REQUIRED findings already exist for individual R batches. The
  historical overlay is not authority to approve, repeat or archive the
  `67fd415..ec5518b` range as one unit. The current queue still names R5's
  composed witness, corrected ranges awaiting cross-party review, exact
  bundle rebuilding and clean-host R-006 as open boundaries.
- At current HEAD `627956be`, `make test-deployment-rehearsal` passed cold:
  both rehearsal/release Go packages and both retained previous/candidate
  *build-record* validators. This is structural record evidence, not
  validation of the now-stale retained bundles.
- `make test-deployment-release` compiled its Linux/amd64 test binary, then
  stopped before assertions because the pinned amd64 `postgres:16-alpine`
  service exited 255 on this arm64/aarch64 Docker host. A direct same-image
  `postgres --version` diagnostic returned `exec format error`; the test
  project's `down --volumes` cleanup completed. This reproduces the current
  queue's architecture limit for composed Deployment tests. It is neither a
  green gate nor an implementation-failure verdict; rerun on an amd64 runner.
- No new designated verdict or RFC status change follows from this checkpoint.

## 2026-10-04 — RP-169 browser-lane discovery predeclaration

- Cold `make test-browser-ci` on the cached ARM64 Linux Playwright image failed before importing `client/test/css-dependency-graph.test.mjs` in Chromium, Firefox and WebKit: Vite injected `__vite__injectQuery` twice into this Node-only fixture (which uses `node:fs`, temporary directories and nested Vite builds). The current browser config collects every test file, while `make test-client` is the intended Node runner for this fixture.
- Bounded correction: exclude this exact Node-only file from the browser population without removing it from the Node population or changing the CSS provenance implementation. A cold Node run must execute its four fixture cases; a cold Linux three-engine browser run must not import it and must report its remaining tests. Deleting or skipping the fixture in both lanes is a failing control, not an acceptable fix.
- The same cold browser run also exposed a separate Cosmetic Shop network-witness failure; that belongs to `planning/cosmetic-shop-v1/log.md`, not this Deployment correction. No R8, AC8, bundle-rights or release status is promoted by fixing test selection.

## 2026-10-04 — RP-169 Node/browser selection result

- The browser config now excludes only `test/css-dependency-graph.test.mjs` using Vitest's default exclusions plus that exact path; future browser-test formats are not globally excluded. The initial unmodified cold Linux run failed import in all three engines. The corrected `make test-browser-ci` completed with exit 0, including Chromium/Firefox/WebKit and the isolated performance lane, without importing this Node-only fixture.
- The file remained in the Node lane: a focused uncached Vitest run executed all four CSS Vite-build cases, 4/4 passing. Full `make test-client` passed 6912 tests with 85 browser-only skips. No CSS producer/consumer or release-bundle byte changed. Claude's designated review of this Codex correction and the wider R8 range remains pending.

## 2026-10-04 — DP6/AC4 × Account AC6 rights supplement (RP-125)

- **Review by:** Codex (implementer first filter). **Recorded by:** Codex. A test-only current-HEAD supplement runs encrypted B_pre and B_post around real Account repository deletion, then restores both into clean Postgres targets. Only B_pre re-creates the deleted account and active Founder/streams. The entire five-test deploymentbackup integration population passes cold on isolated ARM64 Postgres; an account-row-deletion severing fails the new test at source census and production bytes were restored. Method and limits: `planning/platform-alignment/data-rights-restore-resurrection.md`.
- DP6 deliberately excludes account-retention semantics. This is observation of the cross-RFC boundary, not a retention decision or repair. No R-006 supported-host/amd64, credential-revival, full rights, Deployment archival or release claim follows. The exact Codex range awaits Claude's designated cross-party review.

## 2026-10-07 — Current plan/handoff truth reconciliation (RP-306)

At clean `2a785598`, the plan still described Claude R1–R17 as one ready range
with green deployment/aggregate gates, and the dated handoff presented all
findings as closed. Later exact-range verdicts above and the current execution
queue contradict treating those claims as current. The plan row and a clearly
dated handoff correction now point to the actual partial dispositions. Original
submission details and append-only verdicts remain intact.

This is record reconciliation only: no new review of original implementation,
no new test execution, no R11 verdict, no checkbox flip and no range absorption.
The existing bounded approvals remain bounded; Codex correction spans need
Claude, composed amd64 execution is unverified here, and RP-131 is still red.
Fresh exact bundles, DP-F contract/operator authority and R-006 remain open.

Review by: Codex (recorder's local first filter, not designated approval).
Recorded by: Codex. Reviewed the record-only diff against the individual verdict
entries and current queue; no product behavior or acceptance status changes.
The following record commit is another Codex edge requiring Claude where a
closeout range consumes it. No archival, publication or deployment performed.

## 2026-10-07 — RP-141/R18 witness supplement predeclaration

Baseline `fbbf0150` is clean; previous goal turn made progress with RP-304 and
Deployment record reconciliation. No verification handles are live. Re-read
accepted Deployment AC3/AC5/AC10, original R18 CHANGES REQUIRED finding and
current producer/dispatch/bundle validation. No existing test calls the two
functions; the older diagnostic bundles are not retained acceptance inputs.

Scope: new `server/deploymentrehearsal/probe_release_test.go`, per-RFC plan/log,
the existing RP-141 ledger/queue/roadmap routes, and bounded canonical evidence
notes if useful. Use `releasepackage.AssembleBundle` plus actual `LoadBundle`
and `RunProbe`, not an injected always-successful bundle validator. Build
synthetic, hash-bound candidate/previous fixture trees from current tracked
templates/content. They do not become retained real release builds.

Population: both named valid probe outcomes; missing/tampered manifest-bound
previous input for both branches; missing/tampered candidate and equal/reversed
versions for migration; work-directory cleanup and unchanged bundle identities.
Intact pair must reach the production compatibility/runtime boundary before the
migration negative counts. Both missing-input gates must accept intact inputs
first. Each constant-rejection dispatch replacement must fail setup controls;
severing migration, backup refusal and image refusal independently must fail
the valid row. Restore all exact source bytes before cold relevant Go/vet and
the root rehearsal lane. No manual edits while any verification handle lives.

No behavior/kernel/balance/schema/migration/copy/CI or normative author edit is
planned. No R18/DP-F/full-range approval, checkbox, archival, push, deploy or
R-006 claim. If production changes prove necessary, predeclare them separately;
this supplement's new Codex span requires Claude's designated review.

## 2026-10-07 — RP-141/R18 actual producer execution and discrimination

The new 203-line test file builds synthetic candidate/previous bundles through
`AssembleBundle` and loads them through the actual production validator. No
always-successful validator injection or old diagnostic bundle is used. Both
registered `RunProbe` rows reject. Six missing-manifest/tampered-client cases
remain errors, and two equal/reversed-version controls cannot satisfy the
migration negative. Intact manifests remain byte-identical; full bundle
validation and empty probe-work directories check retained fixture integrity.
The production migration producer's positive pair reaches its guard runtime;
the missing-input producer first accepts its real age envelope/manifest and
scripted image identities, then removes the named inputs. These are deliberately
synthetic ELF/image/client/dump inputs, not actual Docker/pg_restore or R-006.

Baseline `7ebfed` /41025, terminal `18157a`: both tests and all ten nested
reported cases pass, package0.780s. Five independent compiling source faults:

- Constant migration rejection: `6cef56` /63696, terminal `03810b`, exit2;
  four invalid-bundle plus two non-forward controls fail (six), other row passes.
- Constant missing-input rejection: `b2207d` /89593, terminal `7743ec`, exit2;
  its two invalid-bundle controls fail, migration controls pass.
- Actual migration comparison omitted: `583887` /40040, terminal `0b7cd0`,
  exit2; valid migration producer fails with outcome0/nil, other controls pass.
- Actual backup gate bypassed: `219186` /65711, terminal `93dad0`, exit2;
  missing-input producer fails with outcome0/nil, migration controls pass.
- Actual image-preparation gate bypassed: `8996af` /9397, terminal `aeab2c`,
  exit2; same intended missing-input producer fails, migration controls pass.

Each fault is restored before the next. `8e9887` and `64702e` prove exact source
restoration; no production delta remains. Source SHAs:
probe.go595a659cc73846eb94146eb77ced22a244010b7b5b93535febfab24714ccafc0;
probe_release.go4790d84a11c83ccc5f7c8711e089026cd98847b6750c25aac910f26438ce3f91;
controller.go36bdb0a42fd1bc8327cb7145fbfed326b38bb87e11b1a8cd18d0f84faf2a62df;
docker.goefd9ea6f2f505ef10ecdff9a5436f5d3d2853861b0222e03a5915bc5d271d35f.
New test SHA0c8c69b94d75db1843e4db707ce375a1bf0aff8d30a13ca9c7dc30112ccaab92.

Final restored checks, all terminal:

- `11ff78` /33407, terminal `6ff64a`, exit0: focused TestR18 population passes,
  package0.647s, no cases skipped.
- Initial declared rehearsal `69baf5` /94131, terminal `35ac86`, exit2:
  rehearsal/CLI tests passed, but an existing release httptest listener was
  denied by sandbox loopback policy before assertions. Exact same root target
  reran with narrowly scoped local-listener permission: `918b6f` /24732,
  terminal `1cd4a1`, exit0; all four packages pass cold, including release2.353s
  and rehearsal2.057s. Both retained build-record validators pass structurally;
  this does not validate/rebuild those historical bundles.
- `b8cfb8` /23590, terminal `40c024`, exit0: cold backup0.535s and
  releasepackage1.240s pass; same-package vet passes. `fd4e65` exits0 for the
  other four-package vet and CI topology's positive/13negative controls.
- Cold normal core lane `575227` /70537, terminal `4f2a77`, exit0: all packages
  selected by `test-go-core` pass (harness explicitly excluded), including
  rehearsal2.026s, release2.533s, production35.238s, transport13.306s. Ordinary
  host SQL skips are not fresh database execution or a hosted Actions verdict.
- `479858` /49028, terminal `a46e49`, exit2: kernel-history checkout and fixture
  gates pass, but full actual history still rejects `50a3a514` against
  `0cf9f7a6` for the six watched minigame paths. RP-131 remains red; no bypass or
  cosmetic bump. A later process-list diagnostic was unavailable (sysmond),
  but the exact test handle already returned terminal exit2; nothing restarted.

All handles are terminal; no temporary fault remains. This test/record supplement
does not approve original R18 or consume any prior independent span. The entire
new range after `fbbf0150` needs Claude, including record edges. No checkbox,
normative owner-body edit, schema/balance/mint/CI/deploy/push/archive or release
promotion. Full nine-tier/platform1.0 goal remains active and this turn progress.

### RP-141 local range review — first filter only

Review by: Codex (implementer, self first-filter). Recorded by: Codex.
Exact inspected range: `fbbf0150..050304de`, all eight changed paths,
including predeclaration, new tests, canonical evidence notes and planning.
Production sources, kernel, CI, balance and migrations have zero delta. Fixtures
use the actual assembler/loader/dispatch, not a substituted validator; invalid
input controls distinguish preparation failures from the named negative.
Five compiling faults fail independently, and exact source hashes restore.
The full cold core/rehearsal results and historical red kernel guard are kept
separate. Diff checks pass; no checkbox or lifecycle state is promoted.

Result: local witness supplement passes the first filter, not designated review.
Claude must cover the entire new span after `fbbf0150`, including this following
record commit. Original R18 CHANGES REQUIRED and all older independent ranges
remain live; this does not approve full DP-F/R-006 or a release. No archival,
deployment or external publication. Full 1.0 goal remains active.

## 2026-10-07 — R11 designated review: alert rules executed, bounded approval

Review by: Codex (cross-party reviewer of Claude's implementation). Recorded by: Codex.
Verdict: **APPROVED for R11 only**, exact range
`1f376cccdabf83359b8d3810fb37ab2c5f0a030b..711071c003f7080299eceb0662f1fafef2e41e16`.
All five changed paths inspected, including the historical evidence entry and canonical docs.
The rule files and affected Go test are unchanged from that range at `54a3ef5f`.

Executed against the real Prometheus evaluator, not `ValidateOperationsProfile`'s structural
checks. The official [Prometheus 3.12.0 release](https://github.com/prometheus/prometheus/releases/tag/v3.12.0)
provides a native darwin/arm64 `promtool`; no global install, release image or Make/CI change.
Downloaded using `gh release download v3.12.0 --repo prometheus/prometheus --pattern
prometheus-3.12.0.darwin-arm64.tar.gz --pattern sha256sums.txt --dir <temporary-directory>`.
Verified archive SHA256 `d758070049a4de5abbeb925b1d4540c5c38a2f40d1356f59aa17a71ac8b48be3`
against both official asset metadata and the release checksum file (SHA256
`251981b6ade521ca72f35878ceee7ff6e9bc6774a8093faffa32776af57dfae9`).
Tool reports 3.12.0, revision `9f27dffc1f93ca23287972f632025879f2d1c658`.

- `<promtool> check rules deployment/operations/cloud-clicker-alerts.yml`: PASS, seven rules.
- `<promtool> test rules deployment/operations/cloud-clicker-alerts.test.yml`: PASS, full
  checked-in population, including duration/threshold controls, stale observation, scrape failure
  and readiness/cleanup fire-then-clear cases.
- Independent adversarial check in a temporary two-file copy: raise only the stale observation
  threshold from 300 to 9999999. The same evaluator exits 1: the expected storage alert at ten
  minutes is missing. No production file was modified; this does not claim independently
  re-executing all thirteen historical probes.
- `make test-go vet GO_PACKAGES='./releasepackage ./operations' GO_TEST_FLAGS='-count=1'`:
  PASS (1.411s / 0.991s), including missing-rule/quiet-evidence/privacy rejection checks.

The seven alert families and privacy boundaries are preserved; stale observation can no longer
silently serve indefinitely healthy values. Native rule-unit evidence is admissible for this
bounded correction, not Linux/amd64 image identity, real scrape-to-receiver delivery, clean-host
R-006, complete DP-E/AC7 or release proof. Original R5 and all other outstanding independent
implementation/record ranges remain open. Next deployment work: the remaining corrective
reviews and exact-artifact/clean-host gates; no archival or publication.

## 2026-10-07 — DP7 native notification integration (supplement, not release proof)

Added `server/operations/alertmanager_native_integration_test.go`, a manual root Make target and
canonical usage docs. This exercises the existing production delivery functions with **real**
Alertmanager and webhook payloads, not a synthetic counter server. The declared test receiver
configuration is retained except its disposable loopback URL. No production, CI, grouping,
interval, timeout, kernel, balance or release-contract change.

Official native Alertmanager 0.32.1 archive SHA256
`4fd3d3d6156cbe76016583b852dc80beda7ee693d28ac27477a9b3da06d87a93` matches the
[release](https://github.com/prometheus/alertmanager/releases/tag/v0.32.1) asset metadata and
published checksum file (SHA256 `11f284aa3b991d61435b6ea6dd0a40262c068ac44b3e1d1aa6fc63464271f018`).
Binary revision `8768aa6f65f1a888b5aa5fbf877cf20ad45d1f61`; archive extracted in a temporary
directory, no global install or tracked binary. Reproduce with
`make test-operations-alertmanager-native ALERTMANAGER_BINARY=<verified absolute native path>`.

- Real integration PASS, 60.52s: preflight's exact nonce received; all seven families delivered
  firing and resolved payloads under one fresh proof nonce; separate healthy receiver returning
  HTTP 500 rejected on notification failure in 0.21s, not timeout. Both process waits completed.
- Missing binary target exits 2; supplying the real Prometheus binary instead also fails the
  version check before starting a service. Neither is a skipped-green integration result.
- Cold `make test-go vet GO_PACKAGES='./operations ./releasepackage' GO_TEST_FLAGS='-count=1'`
  PASS (0.877s / 1.470s). Ordinary optional DB/Compose/native tests skip without their prerequisites;
  this command is not the real-service proof above. `make verify-ci-topology` PASS, including
  its thirteen existing negative controls. No hosted CI claim.
- Excluded initial launcher result: nested Make consumed the regex's dollar anchor and selected
  no tests. Fixed the recipe before the actual run. First selected run was denied a loopback bind;
  identical source then ran with narrow listener authority. No assertion/budget weakening.

Review by: Codex (implementer, first filter). Recorded by: Codex. Scope: the new test, Make target,
docs and owning records after `8f2e29d8`; cross-party Claude review remains required. Real
receiver-side evidence supplements, but does not repair, the runtime record's count-derived
per-family attribution. Original R5, actual scrape/rule-to-receiver/private-network/Caddy and
exact-artifact clean-host gates remain open. Docker overlay rechecked at 0 free / 100%; no cleanup,
deployment, publication, archival or full-1.0 promotion.

## 2026-10-07 — DP7 real cleanup metric → rule → delivery → resolution

Added `prometheus_native_integration_test.go` and `make test-operations-native`; shared native
process/version helpers avoid a second launcher implementation. Production registry, shipped
Prometheus config and unchanged alert rules feed real Prometheus 3.12.0 and Alertmanager 0.32.1.
Only service addresses/rule path are rewritten. Unavailable Caddy/node-exporter and fixture DB
are explicit exclusions, not healthy substitutes. No direct alert injection, counter reset or
shortened rule completes the cleanup journey. The eight-minute completion watchdog accommodates
the declared five-minute rule window and one-minute grouping; it is not a release latency bound.

Execute `make test-operations-native ALERTMANAGER_BINARY=<verified absolute path>
PROMETHEUS_BINARY=<verified absolute path>`; the official binary/archive provenance is in the
preceding R11/native-notification entries. Full target PASS, 399.199s, all selected cases executed:

- Existing real notification population passes again (60.52s), including every injected family's
  firing/resolution and a healthy-but-500 receiver failing on an actual failed request.
- Fresh production zero series scraped at 6.423s; one real `ObserveJob` failure scraped as one
  at 21.447s; actual cleanup firing notification received at 35.748s; natural resolution received
  at 5m35.754s with counter still one. Job labels retain `job=gameserver` and
  `job_name=credential_cleanup`; no `exported_job`/proof nonce substitutes for the metric route.
- Missing binary inputs fail the combined target with exit 2. Cold affected Go tests/vet pass
  (operations 0.900s, releasepackage 1.124s), optional native/DB/Compose tests excluded from that
  ordinary command. Initial unprivileged unit run hit a denied loopback bind; identical source
  passes with narrow listener permission. CI topology and all thirteen existing controls pass;
  no hosted CI claim. All test/process handles are terminal before records are edited.

Review by: Codex (implementer first filter). Recorded by: Codex. The related native-test commits
after `8f2e29d8`, including `7140548c` and this batch's records, form one consolidated Claude
review range. R11's original correction retains its separate bounded designated approval.
Actual private-network services, Caddy isolation, real Postgres/cleanup coordinator, other live
rule producers, release images and R-006 remain unproved here; original R5 and all older ranges
remain open. No acceptance checkbox, CI, runtime, kernel, balance, archive or publication change.

## 2026-10-08 — R5 resumption: private composition still environment-blocked

At `4055b4fe`, re-read the shipped operations Compose profile, its real-service test and DP7/DP8
contract. Read-only inspection finds Docker's overlay at zero available / 100%; the exact pinned
Caddy, Alertmanager and node-exporter images are absent. The pinned Prometheus image is present
as Linux/amd64. An initial node-exporter lookup used an incorrect digest; the subsequent lookup
used the exact Compose digest and also found it absent. No pull, service run or cleanup was
attempted, and no executed test or R5 verdict is claimed. Existing native supplements do not
prove the private network/Caddy boundary. Resume the declared composition when Docker capacity
is restored; do not substitute images or waive its gate. No product or test-tool change.

## 2026-10-08 — Restore refuses schema-only and nonrelational targets (RP-423)

Outcome under DP5/DP6/AC4: the read-only clean-target guard now refuses existing user catalog
objects, not just selected relation kinds. Catalog OIDs use PG16's initdb/user boundary;
explicitly numbered large objects are checked unconditionally. This also catches a user
function in `pg_catalog`, not just public-schema objects. Error class and restore ordering stay
unchanged; no migration, kernel, gameplay, CI or release-policy change.

Old-source50086, real Postgres16 through the declared save-test service, fails five cases:
enum, domain, function, empty schema and large object. Empty/table controls pass. The correction
and extended coverage execute20 arms (empty, table and18 schema-only/nonrelational controls),
checking initial empty admission, the exact refusal and retained seeded object. Each arm owns
and removes a fresh random test database; no existing application/test database is reset.
The OID boundary is verified against PostgreSQL16's primary source:
[transam.h](https://github.com/postgres/postgres/blob/REL_16_STABLE/src/include/access/transam.h).
It is a database-format identity, not a gameplay cap or a convenience measurement threshold.

Final `make test-go-ci CI_TEST_PACKAGES='./deploymentbackup ./cmd/deployment-backup ./deploymentrehearsal' CI_TEST_FLAGS='-v'`
74555 passes all selected packages, including all20 actual SQL arms. Five existing full-backup/
rights/identity integration functions explicitly skip: this service does not supply their
restore/admin URLs or pg_dump/pg_restore16 population. This is NOT full backup integration.
`make test-go vet GO_PACKAGES='./deploymentbackup ./deploymentrelease ./deploymentrehearsal ./cmd/deployment-backup' GO_TEST_FLAGS='-count=1'`
35016 passes all four native packages and vet. Diff/format/whitespace inspected.

Broader Linux39755 is RED: backup, rehearsal and backup CLI pass; release fails
`TestDockerRuntimePreflightRunsCandidateConfigAndPrivateDatabaseInspection` and
`TestDockerRuntimeInitialInstallRequiresCleanStateAndCleansFailedStart`. Focused73724 reproduces
both at their preflight acceptance calls with `ErrInvalid`; the earlier output filter omitted
the detailed failure lines, hence this diagnostic execution. Docker overlay is independently
0 available/100%, consistent with preflight's actual filesystem-space checks. Native equivalents
pass. No floor, fixture, timeout or assertion was weakened; no cleanup, image substitution or
green full-Linux/hosted-CI claim. The full pinned backup/clean-host gates remain separate.

Review by: Codex (implementer diff/first filter). Recorded by: Codex. Exact new range starts
after `ed48bea2` through this code/tests/docs/record commit; designated cross-party review remains
pending alongside earlier Deployment spans. Original R20/DP-C/R-006 are not approved by this
component correction. No archive, release, deploy or push. Next: remaining accepted deployment
and gameplay integration, full restore/private-service gates when capacity permits, and
consolidated review. The full nine-tier1.0 objective stays active.

## 2026-10-08 — Execute the full encrypted Postgres backup/restore population

At clean source `e478b248`, `make test-deployment-backup` PASS. The existing target compiles
the Linux/amd64 test binary and runs the declared Postgres16 service, actual pg_dump/pg_restore
and age encryption, rather than the save-test lane that skipped the restore functions.
All six top-level integration functions execute, zero skips:

- Twenty clean-target arms cover empty admission and refusal/preservation of existing objects.
- Empty and populated databases survive real encrypted dump/restore with exact header and
  restored player/database identity checks. Wrong migration binding is rejected.
- Pre-deletion and post-deletion backups restore their respective account/Founder/stream state.
  The former revives a deleted account: this corroborates RP-125, not a privacy acceptance.
- Recovery validation rejects a projected board event without a verified run, restore refuses
  an occupied target, and same-row-count content mutation changes both identity digests.

The isolated `cloud-clicker-backup-test` project had no containers before execution; its normal
target teardown removes its database container/network, and the final `ps -a` is empty. No
existing development database, unrelated Docker resource or deployed service is modified.
No source/test/assertion change was needed, so no duplicate regression or new review range is
invented. Existing implementation reviews remain pending. This is component integration, not
off-host scheduling, a real release bundle, Caddy smoke, measured RPO/RTO or R-006 clean-host
evidence. DP-C/AC4 remain open. No acceptance checkbox, archive, push or release promotion.

## 2026-10-08 — A concurrent backup cannot replace a completed backup ID (RP-466)

DP6/AC4 atomic commit already rejects existing destinations, but its early `Lstat` followed
by replacing `Rename` leaves a race. Old-source20621 reproduces overwrite with two real encrypted
envelopes: the existing commit seam finishes the second writer between the first writer's check
and rename; both return success and the winning bytes disappear. No sleeps/goroutine timing or
fabricated winning envelope are needed. The regression retains exact winning bytes, authenticated
restore/header identity, losing-writer refusal/temporary cleanup and an already-existing control.

Use atomic no-replace rename: Linux `renameat2(RENAME_NOREPLACE)`, Darwin development
`renamex_np(RENAME_EXCL)`. Existing target collisions remain `ErrInvalid`; other filesystem
failures propagate, with no replacing fallback. Unsupported platforms return unsupported rather
than weaken the guarantee. The already-pinned x/sys0.47.0 becomes a direct dependency, with no
version/sum change. Encryption, receipt/header bytes, sync ordering, retention and objectives
remain unchanged. Additional real-filesystem controls cover absent/file/symlink/directory targets.

Executed: cold complete native backup/CLI/release/rehearsal and vet13818 pass. Declared Linux-
amd64 backup/CLI23256 passes the new collision/rename controls and the real20 clean-target SQL
arms; five full-restore functions explicitly skip in that service. Separately the exact full
`make test-deployment-backup`34023 executes all six integration functions with actual pg_dump/
pg_restore16 and age, zero skips: empty/populated and pre/post-deletion restores, identity mutation
and refusal controls all pass. Its owned disposable project tears down normally. Gofmt/diff checks
pass; every process is terminal before this record edit. No browser/gameplay rerun is relevant.

Review by: Codex (implementer actual-diff first filter). Recorded by: Codex. Exact new range is
`d0fc603a` exclusive through this containing commit, pending designated cross-party review.
This is local backup safety, not off-host filesystem coverage, natural RPO/RTO, a current bundle,
clean-host R-006, full Deployment acceptance or hosted CI. Pre-deletion revival/rights decisions
and other named holds remain. No acceptance flips, archival, push, deployment or release claim.

## 2026-10-09 — Interrupted-backup proof rejects unrelated failures (RP-481)

Outcome under DP6/AC4: the named release negative must reach its injected dump failure,
not count a broken writer or unrelated filesystem/setup error. Old-source94296 fails all
three regression arms: the probe reports `ProbeRejected` with no error after only one call.
It had neither a successful baseline nor proof that the failing read was reached.

The same writer must now create an age envelope that restores the exact baseline payload.
The second call must consume partial bytes and propagate that reader's unique failure.
Only the unchanged valid baseline may remain. Eight new controls reject always-failing/
pre-read/unrelated failures, absent/wrong baseline output, partial leftovers and damaged
baseline, while a wrapped real interruption is recognized. The existing sleeping-writer
control now actually creates the baseline, then wrongly commits the interrupted prefix;
that cannot satisfy the negative. Production backup format/behavior and release bounds
are unchanged. Temporary probe populations are removed; retained release inputs are untouched.

Coverage: focused96881 passes; native15442 runs full `deploymentrehearsal`, rehearsal CLI
and `deploymentbackup` cold tests plus vet, all pass. Declared `make test-go-ci
CI_TEST_PACKAGES='./deploymentrehearsal ./cmd/deployment-rehearsal' CI_TEST_FLAGS='-v'`
97505 passes both full packages on Linux/amd64, zero skips. Initial unquoted regex command
fails in the shell before executing tests; the correctly quoted focused run above executes.
Gofmt and diff checks pass; all handles terminal before this record edit.

These are actual encrypted-envelope/filesystem probe executions, not Postgres dump contents,
off-host storage, RPO/RTO, clean-host R-006 or whole CI. No unrelated browser/gameplay rerun
is claimed. Review by: Codex (implementer actual-diff first filter); Recorded by: Codex.
Exact corrective range starts after `1efea9cf` through this containing source/tests/docs/
ledger/plan/log commit, pending designated cross-party review. No acceptance flip, archive,
push or deployment; continue accepted construction and the existing explicit contract holds.

## 2026-10-09 — Healthy baselines for metrics/secret release probes (RP-482)

Outcome under DP2/DP3/AC2/AC7/AC8: a negative must discriminate the injected defect,
not reward an already-broken subject. Old-source47610 fails four arms: all three
always-rejecting gates count as success, and a real assembled bundle with tampered client
bytes counts as public-metrics isolation. Intact-route and missing-manifest controls pass.

Validate the intact candidate before preparing the metrics route; retain the existing
manifest rebinding and semantic rejection. Source/image probes now scan clean fixtures
through the actual scanners and require the same no-secrets gate to admit them before
scanning their seeded counterparts. No scanner rule, network/runtime configuration,
release bound or existing backup behavior changes. Reuse the real bundle assembler fixture
(synthetic executable/image/client inputs, not runnable release artifacts), not a new framework.
Tests retain input Caddy bytes and require empty probe-work directories.

Focused27871 passes. Full native37774 runs rehearsal, rehearsal CLI and releasepackage cold
tests plus vet, all pass. The declared Linux/amd64 target89854 runs the same three full
packages successfully with zero skips. Gofmt and diff checks pass; all handles terminal.
These are actual bundle-gate/scanner executions, not private runtime reachability,
production-image scans, a clean-host rehearsal or whole CI. No browser/DB rerun is claimed.

Review by: Codex (implementer actual-diff first filter). Recorded by: Codex. Exact new range
begins after `b76d1a08` through this containing source/tests/docs/ledger/plan/log commit;
designated review pending. It may be reviewed together with RP-481 as one coherent range
after `1efea9cf`. No acceptance flips, archival, push or deployment. Kernel-history RFC
adoption has been requested from the owner; no answer or implementation authority is inferred.

## 2026-10-12 — Backup envelope readers reject trailing/unbounded metadata (RP-489)

Under accepted DP6/AC4, inspection and restore must enforce the existing header
contract before treating a file as a valid backup or emitting dump bytes. At
baseline `630f6833`, the standalone DecodeHeader rejects trailing values but both
file paths decode only the first value; restore also lacks bounded line framing.
The new encrypted regressions68719 fail all eight malformed arms: second object,
second scalar, garbage and oversized line in outer or authenticated metadata.
Actual restore emits30/35 dump bytes on those old-code cases; four healthy/JSON-
whitespace controls pass. Encrypted-metadata cases use genuine age ciphertext
and a verified outer checksum, not corrupt-payload shortcuts.

ReadHeader/Restore now share DecodeHeader via bounded ReadSlice framing. Restore
also requires exactly one authenticated metadata object before io.Copy of the
dump. Final14 arms add exact64KiB positive controls alongside the original
healthy/whitespace and malformed cases. No envelope version/format, encryption,
identity fields, checksum, schedule/retention, backup content or operator policy
changes. Canonical docs and RP-489 reconcile in this batch; no new RFC needed for
this contract-conformance correction.

Final cold affected tests15623 pass all four packages: deploymentbackup,
deploymentrelease, deploymentrehearsal and cmd/deployment-backup. Affected vet
74079 passes. Declared `make test-deployment-backup`29449 executes all six
integration functions against real Postgres16 on Linux-amd64, no skips: empty/
populated encrypted dump/restore and identity equality, non-clean-object refusal,
pre/post-deletion backup behavior and identity mutation/refusal. Earlier29958
also passes before the two maximum-line controls were added. The declared target
removes only its disposable test Postgres container/network. No application data
or unrelated Docker resources were removed. Native initial88111 also passes
the affected packages. All handles terminal; formatting and actual diff checked.

This proves the parser correction and retained DB backup path, not clean-host
R-006, production-sized RPO/RTO, rights deletion semantics or hosted/whole CI.
Review by: Codex (implementer first filter). Recorded by: Codex. Exact new range
starts after `630f6833` through this source/tests/docs/ledger/plan/log commit;
designated cross-party review pending. Earlier review debt remains. No feature
acceptance, archival, publication, deployment or release claim.

## 2026-10-12 — Recovery inspection uses one database snapshot (RP-490)

Under DP6/AC4, a recovery identity must describe one coherent database state.
Old-code real-Postgres regression82454 fails: an actual table lock holds the
inspector after hashing accounts, another connection commits a same-count account
change, and one returned identity combines the old Database digest with the new
Player digest. The six existing integration functions pass on that same run.

All queries now use one read-only repeatable-read transaction; helpers require it
and errors return no partial identity. The regression observes the actual blocked
Postgres statement and verifies both digests change across the concurrent commit.
Output/hash definitions, backup bytes, schema and gameplay/retention are unchanged.

The added cancellation observer87372 fails its immediate NOWAIT lock probe. pgx
v5.10.0 PgConn.asyncClose/CleanupDone and database/sql Tx.awaitDone confirm cleanup
is asynchronous. Final observation requires actual database-lock acquisition within
the unchanged30-second invalidity guard, plus context.Canceled and zero identity.
No instantaneous-cleanup promise, retry, production timeout or guard increase.

Final `make test-deployment-backup`25425 passes all seven actual Linux-amd64/
Postgres16 integration functions, no skips, including concurrent-write coherence,
cancelled-query/no-identity and database-observed lock release. The target removes
only its disposable test Postgres container/network; no unrelated/application data.
Final cold `make test-go vet GO_PACKAGES='./deploymentbackup ./deploymentrelease
./deploymentrehearsal ./cmd/deployment-backup' GO_TEST_FLAGS='-count=1'`57464
passes all four packages and vet. Initial corrected integration28087 also passes;
no results inferred from lost handles. Final handles terminal; gofmt/diff checked.

Review by: Codex (implementer first filter). Recorded by: Codex. New exact range
starts after `8ba84996` through this containing commit; designated review pending.
RP-489/490 may share one coherent review after `630f6833`; earlier debt remains.
This is not a snapshot bound to a separate pg_dump, measured RPO/RTO, clean-host
R-006, hosted/whole CI or Deployment acceptance. No archival, push or deployment.

## 2026-10-12 — Designated review: original Claude R5 operations correction

Review by: Codex
Recorded by: Codex
Reviewed range: `78ec39a7^..78ec39a7`
Decision: APPROVED

Bounded to R5's three repairs, not DP-E/AC7 or later implementation. Inspected all
18 original changed paths and the later changes to their consumers at `ffc8bc25`.
The operations registry, delivery code, private fixture and scrape configuration
remain byte-identical to R5; later Caddy payment headers and template/cursor/test
additions are outside this verdict. R5's append-only predeclaration describes the
first attempted counters/default-local admin design; its implementation entry
explicitly records the corrected request counters and `admin off`. Current code
and canonical docs describe the latter, not the superseded attempt.

Executed evidence, cold:

- `make test-deployment-operations-profile`31754 executes the unchanged
  `TestPrivateOperationsProfileIntegration`, PASS61.67s, no skip: all five private
  scrape targets healthy; the actual production registry's zero cleanup-failure
  series observed, then its first `ObserveJob` failure produces a firing shipped
  rule; the receiver gets the preflight nonce; a healthy receiver returning HTTP500
  to notifications causes an explicit failed-notification error; a peer cannot
  reach `caddy:2019/config/`. The test uses a fake DB behind the real registry,
  not a gameserver/database workload. The target removes only its dedicated
  containers/network; independent `compose ... ps -a` confirms none remain.
- `make test-go GO_PACKAGES='./operations ./cmd/deployment-operations
  ./releasepackage' GO_TEST_FLAGS='-count=1'`92089 and affected vet pass. Existing
  failure controls refuse attempted-but-failed delivery and exposed/default/
  duplicate admin listeners; the registry population checks zero series and
  absence of reserved target labels. R5's recorded label/pre-creation severings
  remain historical evidence, not newly rerun probes.
- The pinned Prometheus manifest index `69f52414…60a8ac` declares amd64 child
  `dd4bced05dfaddf23a7ec50f87334993a4149f7fcfbf58456d1c8bafce91cd13`.
  `docker run --rm --platform linux/amd64 --entrypoint promtool -v
  /Users/stronk/repos/cloud-clicker/deployment/operations:/work:ro
  prom/prometheus@sha256:dd4bced05dfaddf23a7ec50f87334993a4149f7fcfbf58456d1c8bafce91cd13
  test rules /work/cloud-clicker-alerts.test.yml`36424 reports SUCCESS. This is
  the exact index's child, not a substituted release/version.
- `make test-go-ci CI_TEST_PACKAGES='./gameserver'
  CI_TEST_FLAGS='-run TestComposedGameserverStartupPrimesAttachedClearingAndSessionGCIntegration -v'`
  93146 executes the actual Linux-amd64/Postgres startup/clearing/session-GC job
  witness, PASS0.65s, not skipped. It checks R5's composed `job_name` output.

Failed attempts/limits are not erased: the first index-based amd64 promtool run
71604 exits125 because Docker cannot overwrite the index digest already stored
for arm64; no rule test executed. Resolving the index's declared child avoids
that image-store collision without deleting any image. The first DB selector
99617 used a single dollar anchor that Make consumed, yielding `[no tests to run]`;
it supplies no evidence. The corrected verbose selector above actually runs.
Docker had2.8GiB available before exact-image fetch55173 completed. No unrelated
images, volumes, containers or application data were pruned. The profile runs
on the ARM64 Docker host: Caddy amd64, Alertmanager/node-exporter/Postgres fixture
images arm64, amd64 test binary, host-selected Prometheus. It is deliberately
not claimed as an all-amd64 supported release-host execution.

The delivery counters still cannot attribute each family/nonce. The actual nonce
receiver observation strengthens this bounded fixture, not the count-derived
seven-family release observation. Full seven-family producer→rule→receiver proof,
retention, all other corrective ranges, exact fresh bundles, R-006 and complete
Deployment acceptance remain open. No archival, publication or deployment.

### Separate Codex test-target extraction — designated review pending

Under DP7/RFC-0000, the existing private Compose recipe is extracted unchanged to
`test-deployment-operations-profile`. Its original aggregate still evaluates
rules, runs the exact same affected Go population, then calls the fixture. This
allows direct execution without the known-red production suite preventing entry;
it does not bypass that suite for aggregate success. `make -n
test-deployment-operations` confirms ordering and cleanup. No test, assertion,
service config, CI workflow, budget or product byte is changed.

`make verify-ci-topology` passes with all13 negative controls; this checks
lane wiring, not hosted job execution or green CI.

Outcome/coverage are the actual fixture result above; docs explain the focused
and aggregate boundaries. Review by: Codex (implementer first filter).
Recorded by: Codex. The new Make/docs/records range starts after `ffc8bc25`
through this containing commit and requires Claude. This original-R5 designated
verdict does not approve Codex's extraction. Diff and whitespace checked; no
whole aggregate, hosted CI, release or feature-completion claim.

## 2026-10-12 — Exact backup objective bounds before report conversion (RP-503)

Outcome under DP6/AC4: `MeasureObjectives` must refuse any duration beyond six-hour
RPO/four-hour RTO, including subsecond overruns. Old code truncated to seconds first;
the existing final rehearsal validator already compared durations. Added one regression
with12 cases: each objective at zero, one nanosecond below, exactly at, one nanosecond
over,999ms over and one second over. Old-source26841 fails precisely the four subsecond
overrun cases, returning a success measurement. Corrected code compares full durations
before constructing the unchanged whole-second report and returns no partial result on
refusal. Backup completion/incident/restore/smoke meanings and all bounds stay unchanged.
Canonical Deployment docs explain this distinction; no new policy or schema.

Executed final checks:7823 cold `make test-go vet` over `./deploymentbackup
./deploymentrehearsal ./cmd/deployment-backup ./cmd/deployment-rehearsal` with
`GO_TEST_FLAGS='-count=1'` passes all affected packages and vet. Native DB prerequisites
skip;66695 `make test-deployment-backup` separately executes and passes all seven actual
Linux-amd64/Postgres16 integration functions, including empty/populated encrypted restore,
non-clean refusals, identity mutation, concurrent snapshot and pre/post-deletion populations.
Its declared isolated services/network are removed by the existing target.89677 declared
`make test-go-ci CI_TEST_PACKAGES='./deploymentbackup' CI_TEST_FLAGS='-v -run TestObjectiveMeasurement'`
passes the new12 cases plus the existing incomplete/late/slow tests verbosely; those objective
tests are pure arithmetic, not SQL proof. All handles terminal; gofmt/diff checks pass.

Review by: Codex (implementer first filter). Recorded by: Codex. Exact correction range
starts after `ad54b258` through this containing commit; Claude designated review pending,
separate from earlier debt. No acceptance checkbox, archival, CI/workflow change, full-CI
green or exact-bundle clean-host claim. RP-131/RP-307, author/operator decisions, rebuilt
bundles and R-006/R-007 remain. Continue accepted construction/repair and consolidate reviews;
the four pending proposals are not adopted by automatic goal continuation. No push/deploy.

## 2026-10-12 — Include stop/reset in populated recovery elapsed time (RP-504)

Outcome under DP6/AC4: time required recovery work, not just dump restoration onward.
`runPopulatedRecovery` sampled restore start after `StopFailed`/`ResetDatabase`; real
Docker adapters shut down the stack, remove its database volume and wait for fresh
Postgres in those phases. The stopwatch now starts after validated inputs/operator lock/
checkpoint, before those actions. A clock preceding the declared incident refuses before
destructive calls. The4h bound, RPO origin, serialized format and final smoke endpoint
remain unchanged; canonical docs state the corrected phase boundary.

Regression uses the existing coordinator fixture with phase-driven time. At unchanged
production code69008 all four cases fail: stop1min/reset2min/smoke3min reports180s,
stop/reset over4h still emit success, and an invalid start reaches destructive work.
Initial66138 also assumed the rehearsal error class for the last arm, but the old path
returns the backup validator's error; removed that unsupported class assumption and
reran unchanged production before correction. Final four cases pass, report360s, consume
the actual written observation through its decoder, and leave no success artifact on
overrun. No source severing/repeated acceptance programme needed for this regression.

Final16318 `make test-go vet` with `GO_PACKAGES='./deploymentrehearsal
./deploymentbackup ./deploymentrelease ./cmd/deployment-rehearsal'` and
`GO_TEST_FLAGS='-count=1'` passes all affected packages/vet. Native DB prerequisites skip.
69411 declared `make test-go-ci CI_TEST_PACKAGES='./deploymentrehearsal'
CI_TEST_FLAGS='-v -run TestRecovery'` executes the new four and all existing recovery
functions successfully on Linux-amd64. Runtime/bundle/backup/identity dependencies are
controlled fixtures, not SQL restore proof merely because the lane supplies a DB URL.
No actual Docker stop/volume reset, host incident or clean-host rehearsal executed here.
All handles terminal, source formatted and diff checked.

Review by: Codex (implementer first filter). Recorded by: Codex. Exact new range begins
after `e6674289` through this containing commit; Claude review pending, combinable with
the RP-503 range after `ad54b258`. Earlier debt remains. No acceptance checkbox, archival,
CI/workflow/budget, operator authority, four-proposal adoption, push or deployment change.
The complete1.0 goal remains active; exact current/previous bundles, clean-host R-006/R-007,
RP-131/RP-307 and author/owner decisions remain separate requirements.

## 2026-10-12 — Outer protection metadata permits unsafe retention (RP-505)

Under DP6/AC4, cleanup must preserve unresolved pre-upgrade backups and refuse corrupt
input. Two independent real age-encrypted populations each retain and restore their
intact31-day protected backup alongside a recent ordinary control. Changing only the
outer `pre_upgrade` or `upgrade_resolved` flag preserves encrypted payload/checksum:
`ReadHeader` accepts and actual `ApplyRetention` deletes the protected fixture, while
authenticated `Restore` refuses before output. Both recent controls survive. Protection
fails2/2; successful diagnostic execution is explicitly NOT retention acceptance.

Executed final checks:8339 cold `make test-go vet
GO_PACKAGES='./deploymentbackup ./cmd/deployment-backup' GO_TEST_FLAGS='-count=1'`
passes affected packages/vet.59504 declared `make test-go-ci
CI_TEST_PACKAGES='./deploymentbackup'
CI_TEST_FLAGS='-v -run TestRetentionOuterProtectionMetadataResearch'` executes both arms
on Linux-amd64 and reports the same2/2 failures. These are actual encryption/filesystem
operations in test-owned temporary directories, not Postgres dumps or a clean-host
rehearsal. Native DB prerequisites may skip; no SQL proof is claimed. Earlier61595/66459
completion output was unavailable after context recovery; final evidence above was
rerun after adding explicit refusal of an unobservable filesystem outcome. All current
handles terminal; gofmt/diff checks pass.

DESIGN-GAP: retention needs independently verifiable protection metadata while the age
restore identity stays off-host. Draft `rfc/backup-protection-metadata.md` routes verification
authority, atomic publication, existing-backup compatibility and the separately unresolved
resolution lifecycle. No mechanism is selected/adopted; runtime bytes, format, key handling
and retention durations are unchanged. Canonical docs now disclose the current limitation.
Next: complete that engineering contract before construction, without treating a draft as
authority or weakening DP6 to fit the implementation.

Review by: Codex (implementer first filter). Recorded by: Codex. Exact new diagnostic/docs/
draft range starts after `09cf45e5` through this containing commit; Claude designated review
pending, combinable with related Deployment corrections. Earlier review debt remains.
No acceptance, archival, hosted-CI, four-proposal adoption, push or deployment claim.

## 2026-10-12 — Complete the RP-505 engineering recommendation (draft only)

The previous goal turn made progress: two executed cases established unsafe deletion
and the committed diagnostic/draft changed the next action. This batch completes a
recommended repair contract rather than selecting an unreviewed runtime fix. BM1–BM5
specify standard Ed25519 over exact metadata/payload identity, independent public trust,
dedicated writer-secret placement, atomic single-file publication and target-owned locked
cleanup. Authenticated legacy copying preserves original bytes/timestamps/protection;
the schedule never gets the age identity. Post-rename sync failure is explicitly
indeterminate durability, not an impossible promise to undo publication.

Source inspection covered Create/ReadHeader/Restore, encrypted package extraction,
Postgres create/restore, target classification, CLI scheduled/retention rounds, Linux
no-replace rename, current Compose mounts and release-state authority. The official
Go crypto/ed25519 documentation supports the selected standard-library API and key/
signature sizes; it is not evidence this proposed integration works. The ciphertext's
authenticated metadata and package format can remain unchanged because neither carries
the outer schema version/payload hash. Existing resolution and actual snapshot-based RPO
remain distinct author gaps; no protection flag is patched or inferred from release success.

Verification is source/specification diff and whitespace inspection. No software/crypto
acceptance tests were run for this normative draft, no private key generated, no backup
converted and no operator configuration changed. Plan/index now point to adoption and
normative review instead of unfinished engineering choices. RP-505 remains OPEN.

Review by: Codex (author first filter). Recorded by: Codex. Exact draft range starts
after `0e9503f2` through this containing commit; Claude designated normative review pending.
Owner adoption is also required for the new format/signing-secret/legacy-copy contract;
neither automatic goal continuation nor this record supplies it. No product implementation,
acceptance/archival, CI-green, publication or deployment claim. Continue safe work under
existing accepted contracts while this bounded amendment is held.

## 2026-10-12 — Package restore refuses trailing data without a pipe deadlock (RP-507)

DP6/AC4 requires corrupt backups to fail visibly, not stall restoration. The exact
four-member archive can reach tar EOF while its authenticated age stream still has bytes.
ExtractPackage then waits for the decryptor, which is blocked writing that unread tail.
New regression20185 on unchanged product fails for an extra byte, a zero block and a
concatenated archive (each reaches the5s deadlock guard); exact-package control passes.
Every fixture independently decrypts successfully with a matching outer checksum, so these
failures exercise package framing, not damaged encryption or missing prerequisites.

Consume the pipe to authenticated EOF after successful tar decoding; any remainder refuses.
Close the reader on all returns. No archive/envelope writer, format, retention, secret,
budget, timing objective, version, CI or operator configuration changes. Corrected
regression72906 passes all four cases in0.08s. The5s test guard detects a deadlock; it is
not a restore-performance measurement or a changed acceptance ceiling.

Final checks:

- `make test-go vet GO_PACKAGES='./deploymentbackup ./cmd/deployment-backup ./deploymentrelease ./deploymentrehearsal' GO_TEST_FLAGS='-count=1'`
  79435: all four affected packages and vet PASS; native missing-URL DB cases are not SQL proof.
- `make test-deployment-backup`67138: all seven Linux-amd64/Postgres16 integration functions
  PASS. Both actual empty/populated pg_dump streams with authenticated extra data refuse;
  RequireCleanTarget still accepts the target, then the intact originals restore with matching
  stored and semantic recovery identities. The declared isolated project cleans itself up.
- `make test-go-ci CI_TEST_PACKAGES='./deploymentbackup ./cmd/deployment-backup ./deploymentrelease ./deploymentrehearsal' CI_TEST_FLAGS='-v'`
  32928: affected Linux-amd64 packages PASS cold count1, including all four new framing cases.
  Its general DB service lacks the dedicated restore/admin URL population;67138 owns that proof.
  The unrelated orphan warning did not authorize cleanup. The RP-505 diagnostic still reports
  both protection failures despite its successful instrument exit; no retention acceptance.
- gofmt and actual diff/whitespace checks PASS; all run handles terminal.

Docs/plan and RP-507 reflect this bounded repair. No full Deployment/Actions, clean-host R-006,
release artifact, nine-tier completion, acceptance checkbox, archival or publication claim.
Review by: Codex (implementer first filter). Recorded by: Codex. Exact new range starts after
`4627861a` through this containing commit; Claude designated review pending, consolidatable
with related Deployment corrections. Existing metadata-binding, snapshot-origin/resolution
author holds and earlier independent ranges remain; no proposal was implicitly adopted.

## 2026-10-12 — Rollback validates the backup package before database destruction (RP-508)

DP5/DP6/AC5 and the existing restore_inputs stage require corrupt input to refuse before
stop/reset. RP-507's real package decoder refuses trailing data, but host rollback preflight
still decrypts only to a discard sink. Regression49581 admits both an authenticated non-package
and a complete package with trailing plaintext. Its healthy control now uses CreatePackage,
not the old raw-dump envelope. Initial98310 has a wrong repository/content path and establishes
no product failure; correction reads the existing root epoch without changing its bytes.

Reuse ExtractPackage inside a private0700 temporary workspace before returning preflight
success; its0600 dump and workspace are removed on return. Existing key/header/path checks
remain, and no runtime command is issued. Add a changed-dump/old-hash refusal and assert no
plaintext workspace survives all tested returns. R18's missing-input diagnostic initially
fails its healthy control because it generates only a raw-dump envelope. Correct that existing
producer to package the actual validated previous manifest/epoch and its explicitly synthetic
dump; do not bypass the guard, count baseline failure as a negative, or claim a real DB restore.
No backup format, secret placement, retention, objective bound, CI or controller ordering change.

Executed evidence:

- Focused corrected host check98794 passes. Final regression/cleanup51651 also passes on
  declared Linux-amd64 with healthy, non-package, trailing-data and wrong-dump-hash populations,
  while retaining existing missing/wrong/multiple identity, manifest/class and ciphertext cases.
- Final `make test-go vet GO_PACKAGES='./deploymentbackup ./deploymentrelease ./deploymentrehearsal ./cmd/deployment-release ./cmd/deployment-rehearsal' GO_TEST_FLAGS='-count=1'`
  25939 passes all five packages and vet. Native unavailable DB cases do not establish SQL proof.
- `make test-go-ci CI_TEST_PACKAGES='./deploymentbackup ./deploymentrelease ./deploymentrehearsal ./cmd/deployment-release ./cmd/deployment-rehearsal'`
  26233 passes all five affected Linux-amd64 packages cold count1, including repaired R18 controls.
  Final test-only workspace assertion is separately covered by51651; no hosted/whole-CI claim.
- `make test-deployment-release`78484 passes the real Linux-amd64/Postgres16/Caddy journey:
  actual pg_dump package passes host preflight; its authenticated trailing-data variant refuses
  before any runtime command; readiness remains; admitted work drains/socket courtesy closes;
  restart, actual intact restore, identity/start/smoke and severed-artifact refusal still pass.
  The declared isolated project cleans up. This is a repository fixture, not a release bundle
  installed on a clean host or actual DockerRuntime rollback destroying a live volume.
- Earlier36355 and the first release-target attempt fail compilation on a missing test import;
 77338 then exposes R18's invalid healthy fixture. These results remain red, not passed gates.
 Final97018/26233 and25939/51651 supply the corrected consumer results. gofmt/diff checks pass;
 all run handles terminal. Unrelated orphan containers were not cleaned.

Docs/plan and RP-508 reflect the guarantee and its limits. Logical pg_restore acceptance and
production-size extraction/RTO remain measured clean-host gates; D-019 still owns the distinct
every-release isolated-restore choice. Metadata-binding/snapshot-origin/resolution holds remain.
Review by: Codex (implementer first filter). Recorded by: Codex. Exact new range starts after
`cd609d07` through this containing commit; Claude designated review pending, combinable with
related Deployment repairs. No full feature approval, checkbox, archival, push, deployment or
nine-tier/release-completion claim follows.

## 2026-10-12 — Bind both backup creation passes to the complete dump (RP-510)

Outcome under DP6/AC4: a completed package must bind the entire supplied dump,
not a suffix hash or changed second pass. Definitive unchanged-source53556
fails four cases: positioned-reader success followed by restore refusal,
EOF-position rejection of a complete readable dump, same-length substitution
accepted, and appended bytes silently omitted. Initial-position zero and stable
input pass; shortened second-pass input already refuses. Initial24545's changed-
source helper accidentally promoted bytes.Reader.WriteTo, bypassing its Read hook;
those three changed-source results are invalid. Corrected helper hides that fast
path and asserts exactly one between-pass replacement before counting evidence.

Repair: rewind before first-pass hashing, hash the actual streamed second-pass
bytes, and require exact EOF before closing the pipe successfully. Existing atomic
envelope publication/cleanup remains authoritative. No backup format, age identity,
metadata/retention choice, timestamp semantics, bound or CI workflow changes.
Corrected89793 passes all seven regression arms; every changed source refuses
with no header/path or retained files, while all initial positions restore exactly.

Executed final checks:95691 root `make test-go vet GO_PACKAGES='./deploymentbackup
./cmd/deployment-backup ./deploymentrelease ./deploymentrehearsal' GO_TEST_FLAGS='-count=1'`
passes all four packages and vet.63312 `make test-go-ci` with the same four
`CI_TEST_PACKAGES` passes cold on Linux/amd64.10138 `make test-deployment-backup`
executes all seven real Postgres16 integration functions without skips: original
empty/populated restore and trailing-data refusal remain, and both actual dumps
repackaged after reading their signature restore with exact semantic recovery
identity. Native/general-CI missing dedicated restore URLs are not that SQL proof.
The declared isolated target cleans itself up; unrelated orphan containers are
untouched. Formatting/diff checks pass; all handles terminal.

Docs/plan and RP-510 record the guarantee and limitations. Review by: Codex
(implementer first filter). Recorded by: Codex. Exact full source/test/docs/ledger/
plan/log range after `e27eb832` through this containing commit needs Claude
designated review, combinable with related backup corrections. No acceptance,
archival, hosted/whole-CI, clean-host/RTO or release claim; signed protection
metadata, snapshot-origin/resolution author holds and D-019 remain separate.
Continue accepted construction/correctness while those specific contracts are
held; full nine-tier1.0 goal remains active. Nothing pushed or deployed.

## 2026-10-12 — Refuse malformed image tar headers without panic (RP-511)

Local release construction at clean `abf4b4ac`:97246 builds all six Linux/amd64
commands;77862 builds the client;11019 stages epoch8 content and generates metadata
for53 dependencies. Outputs are ignored under `.cache/release-current-abf4b4ac/`
and `client/dist/`; no image/SBOM/full bundle or reproducibility claim follows.
The old candidate record still names `7e8aa70`; it was already invalidated by later
validator corrections, so the plan now explicitly distinguishes that historical
build from a usable current candidate. No old bytes or build record relabelled.

During construction-path inspection, ValidateDockerArchive dereferences the tar
header before checking Next's error. Old-source9013 reproduces four panics each
through direct validation and AssembleBundle: short first/later headers and bad
first/later checksums. Empty input and truncated member bodies already refuse;
intact archive and assembled/validated bundle controls pass. These are synthetic
structural archive/ELF fixtures, not an executed release image. Under DP1/AC1/AC8,
check the parser error first and return invalid content with its original cause.
Corrected848691 passes both named tests and all twelve negative arms; archive
failure returns no manifest and leaves no published manifest in the output tree.

Final checks:5883 cold native `make test-go vet` for releasepackage,
cmd/assemble-release-bundle, cmd/gen-release-metadata, deploymentrelease and
deploymentrehearsal passes; assembler command has no separate unit tests, its
actual function is exercised above.74516 `make test-go-ci` for the same packages
passes on Linux/amd64.99190 `make test-deployment-release` executes the real
Postgres16/Caddy backup/drain/restart/rollback integration without a skip; it
retains its helper-fixture boundary, not an installed exact release bundle.
The isolated target removes its own services; unrelated orphans are untouched.
Formatting and diff checks pass; all handles are terminal.

Docs/plan and RP-511 reconciled. Review by: Codex (implementer first filter).
Recorded by: Codex. Full source/test/docs/ledger/plan/log range after `abf4b4ac`
through this containing commit needs Claude designated review, consolidatable
with related packaging corrections. No new RFC, image/format/key/bound/CI change,
hosted/full-CI, clean-host, acceptance or release claim. Next: rebuild exact
current candidate inputs from the committed correction, retaining previous
artifacts unchanged; operator/author holds and full nine-tier1.0 scope remain.

## 2026-10-12 — Fresh two-tree candidate construction; platform input defect (RP-512)

DP1/AC1/AC8 local construction uses actual clean starting source
`102face6f2aeadc8aaa0ff19af0d4d5bd0f63ef1`, timestamp `2026-10-12T21:29:21Z`,
source-date epoch `1791840561` and existing local candidate label `0.1.0-preview.2` (not a
published release). Both independent trees rebuild the client, six static
Linux/amd64 Go commands, epoch8 content, metadata for53 dependencies, no-cache
image archives and seven fresh Syft scans. No source/build input changes occur
between them; the intervening dirty file is only the new RP-512 ledger row.
The Dockerfiles match the repository template; BuildKit/Syft/frontend images
stay pinned. Actual Docker28.4.0 and Compose2.39.4-desktop.1 are inspected, not
inferred from old records. An initial metadata invocation used a mistaken full
commit/time: `.cache/release-current-102face6/metadata/` is invalid and excluded.
Both assembled trees instead use `metadata-corrected/` with the actual Git values.

RP-512: pinned registry index → unique Linux/amd64 child → config.digest and
config architecture/OS verification exposes two wrong identities in **both old
records**. Alertmanager's old23b56bfd… and node-exporter's old84acc55a… are ARM64,
also confirmed by cached-image inspection. Actual amd64 configs are6519630d…
and9bbbca8f…. The other four upstream configs are independently checked too;
single-image Caddy/Playwright manifests are handled without inventing an index.
All fourteen platform-selected image scans execute; no old SBOM is reused.
Documentation now explains explicit platform resolution instead of host-native
image inspection. Old files/records remain unchanged, not silently repaired.

Actual outputs:
- `.cache/release-current-102face6/bundle-a` and
  `.cache/release-rebuild-102face6/bundle-b`, each74 manifest-bound artifacts;
- manifest `sha256:061375c80401d5aaaac172c894837cb4e3ca082a59250e7ad40b635338a46073`;
- gameserver archive `sha256:0ffddcec5503ff6b20691a4a1039057e7eb3ad502e13b65cebc77033fc2e0d4b`;
- gameserver config `sha256:0a6a731624aabc87cf0fb4df42afa8f35ba36b7c50e9455689dc1e48494a9677`.

Root six-command build targets, `build-client`, `stage-release-content`,
`generate-release-metadata`, `build-gameserver-image`, `generate-image-sbom`
and `assemble-release-bundle` complete in each tree. `diff -r` compares the
entire bundled trees byte-identical; separate comparisons also cover all six
binaries, staged content, metadata, archives and seven normalized image SPDX
documents. Raw nondeterministic Syft headers are neither compared nor shipped.
The new small retained `release-builds/candidate-102face6.json` comes from actual
bundle hashes/images, and the native-built rehearsal command's `validate-build`
accepts it. The structured `make release-secret-scan` bound to the above manifest
and exact image completes over2126 tracked files, zero findings, without a guard.
Only the RP-512 tracking row was dirty during that scan; no build input changed.

Both assemblers validate their actual bundle. The rehearsal `probe` command's
healthy supply-chain population exits0 on this **independent candidate rebuild
pair**, not an immediately previous release. Each of nine real-bundle mutations
exits3 (`fixture_rejected`), not setup failure: removed catalog/client/license/
config/helper, changed image/runtime-config/SBOM and wrong epoch/artifact set.
Their healthy-first semantic/rebound checks preserve the originals; a final
whole-tree comparison passes and probe work is empty. All handles are terminal.

Limits/next: `.cache/release/previous/bundle` is absent on this machine; the old
previous record has RP-512's bad architecture identities. Do not substitute the
second candidate as a previous release or fabricate the missing supply-chain
pair. Recover/reconstruct legitimate previous artifacts and resolve their
provenance before the full candidate/previous supply-chain/rollback population.
Full-range designated reviews, producer/author holds, actual CI and explicitly
authorized clean-host R-006 still gate support/release; this construction is not
a deployment, platform acceptance or the full nine-tier1.0 result. No push.
Review by: Codex (first filter of outputs/record/docs, not designated approval).
Recorded by: Codex. The record/docs/plan/ledger/log range after `102face6` through
this containing commit is available for consolidated designated review. No
product, migration, image pin, contract, bound, CI or owner-copy bytes changed.

## 2026-10-12 — Bind actual image layers to configuration identity (RP-513)

DP1/AC1/AC8 repair, not a new packaging format or product ruling. The archive
validator previously checked the config hash and layer-name presence but never
compared actual layer bytes to the config's ordered rootfs identities. The new
plain/gzip population on unchanged production (48227) has ten failing negatives:
changed first/second layer, reversed layer order, absent rootfs and missing DiffID;
both intact controls pass. This establishes an offline validation defect, not a
Docker-loader bypass: release preparation already loads/checks images before drain.

The validator now streams uncompressed hashes and matches every manifest layer in
order to `rootfs.diff_ids`. Plain/gzip encoding and existing512MiB member/1GiB
decompression bounds remain; no layer is silently excluded. Bad gzip checksum,
truncated stream, malformed header, unsupported zstd and source read failure refuse,
with the read-error cause preserved. Synthetic existing fixtures acquire correct
rootfs metadata without changing their payloads; they are still not runnable images.
The new two-layer regression fixtures use actual inner tar archives.

Final executed checks, all terminal:
- 96093: cold verbose ordered-layer population, twelve cases, and all five codec/read
  refusals pass; intact plain/gzip hash controls also pass.
- 31939: cold affected releasepackage, deploymentrehearsal, deploymentrelease and
  their three command packages pass, plus vet. The assembler command has no separate
  tests; the actual assembly function remains exercised by releasepackage tests.
- 19100: the same package boundary passes through the declared Linux/amd64 target.
  This is not evidence of dedicated SQL integration populations executing there.
- 99278: actual Postgres16/Caddy release/drain/exact-backup rollback integration
  passes without skipping (2.38s); the isolated target removes its own services.
- Fresh `make build-deployment-rehearsal`, then actual supply-chain probe against
  both source102face6 bundle trees exits0 under this validator. These are independent
  candidate copies, not a legitimate candidate/previous pair or clean-host R-006.

An interrupted tool output hid earlier check handles/results; only the fully read
fresh final runs above support the final check claim. Read-only process inspection
confirmed no overlapping jobs for these packages before rerunning. Unrelated processes
and the warned game-ui orphan remain untouched. Formatting/diff checks pass.

Docs, plan and RP-513 reconcile in this batch. Review by: Codex (implementer first
filter). Recorded by: Codex. The complete source/tests/docs/ledger/plan/log range
after `13621c7d` through this containing commit needs designated Claude review,
consolidatable with the packaging repairs. No new bound, wire, key, CI or gameplay
change; no self-approval, archival, deployment or release claim. Retained source102
bundle helpers predate this repair: rebuild at the final reviewed source rather
than relabel artifacts. Actual previous artifacts/provenance, contract-author holds,
full-range reviews and explicitly authorized exact clean-host rehearsal remain open.

## 2026-10-13 — Reconstruct the historical predecessor; current floor refuses it (RP-514)

Source `44a4a72a52ef1b299075389bfa5f365baa00e4c8`, epoch1787441542,
`2026-08-22T23:32:22Z`, was rebuilt in an exact managed historical checkout without
product edits. Root Make lanes build all five original static Linux/amd64 commands,
production client, epoch8 content (31 files), metadata for44 dependencies and the
no-cache gameserver image. This pre-browser source has no deployment-browser helper;
do not invent that output or claim the reconstructed bytes are the original binaries.
The compiler is current `go1.27.2 darwin/arm64`, targeting Linux/amd64.

Two independent builds each perform six actual fresh Syft scans and assemble70 artifacts.
Pinned upstream indices/amd64 child manifests/configs are inspected, correcting the old
ARM64 Alertmanager/node-exporter identities only for these new reconstructions. Tooling:
Docker28.4.0, Compose2.39.4-desktop.1, pinned BuildKit0.24.0 and Syft1.51.0. Pins,
historical source and all three existing build records remain unchanged. The small new
`release-builds/reconstructed-44a4a72.incomplete.json` retains actual identities.

Outputs retained in main's ignored `.cache/release-previous-44a4a72/`:
- `a/bundle` and `b/bundle-matched-tag`, byte-identical whole trees;
- manifest `sha256:4fe204c9475091405a0da7f2f260622ebe517703d7f2366f3f8695f4bf974bd4`;
- archive `sha256:8fa5476b61ace42902174aa0a3782cd8bc4f3e8469241c4cb1a9cbec91b7cecd`;
- config `sha256:1a23b95e96fb0db55075af6a1a1d82e22c7fea592402b74f94a8c89e6ddb67ea`.

Exclusion disclosed: initial B used a different Docker tag, changing the archive and its
manifest hash only. Those outputs remain in `b/bundle`. A third no-cache image build from
B's independently rebuilt inputs uses A's same tag, followed by its own fresh gameserver
scan and assembly into `b/bundle-matched-tag`. The corrected full comparison passes;
separate staged-content and metadata comparisons pass. Total13 actual image scans,
three image builds; raw nondeterministic scan headers are neither shipped nor compared.

Fresh current rehearsal helper and temporary test-side diagnostics establish:
- Current source102face6 candidate `ValidateBundle` passes. The reconstructed predecessor
  fails `invalid release Compose`; its actual Caddy also fails the current admin-off/private
  metrics requirement. Old Caddy declares `admin :2019`, not `admin off`; old Compose lacks
  current cursor ID configuration. An in-memory diagnostic replacing2019 with2020 exposes
  that next missing field; no source or artifact is changed. This is a floor refusal, not a
  newly demonstrated live exploit or an assertion that historical cursor absence leaked data.
- The actual prepared candidate/previous supply-chain probe exits2 `invalid_probe`, not
  successful proof or fixture rejection3. Earlier invalid population-name/missing-work-dir
  attempts are setup failures, excluded from the finding.
- Both actual old image archives pass current ordered-layer/config validation (59491).
  Diagnostic runs43945/22322 disclose the bundle refusal, not acceptance; temporary
  diagnostic test files are removed. No safety validator or production byte is changed.
- Fresh whole-bundle diff exits0. Actual record manifest/six SBOM hashes match the retained
  artifacts. `validate-build` deliberately rejects the new incomplete record (exit1,
  `invalid_evidence`), while the existing candidate record passes. This incomplete record
  cannot enter R-006 as a supported predecessor. Existing records diff unchanged.

The artifact-availability investigation is complete; the supported predecessor is not.
Next is a reviewed floor-compliant rehearsal predecessor with explicit provenance, then
full pair proof and exact clean-host rehearsal. Do not grandfather the old management API,
relabel the historical package, use a second candidate as the previous release, or rerun this
reconstruction as if missing files remained the blocker. Candidate helpers also predate
RP-513 and must be rebuilt at final reviewed source. Producer/author holds, full-range
reviews and explicit clean-host authority remain; no hosted CI, R-006, release or1.0 claim.

Review by: Codex (output/record first filter, not designated approval). Recorded by: Codex.
The records/docs/plan/ledger/checkpoint range after `07616e41` through this containing
commit needs designated Claude review, consolidatable with the existing packaging range.
No push, deployment, migration, bound, image pin, contract or owner-copy changes.

Temporary checkout cleanup: removed only the dependency symlink created for these builds;
the historical tree is clean and main's dependencies/retained artifacts remain. The managed
archive operation refused because the worktree is protected by a pinned task or workspace.
Left that checkout intact; no manual deletion or attempt to bypass its protection.

## 2026-10-13 — Same-second backup failures reach the operator (RP-518)

**Outcome / authority:** Deployment DP7/AC7 requires any backup failure to page.
`RecordOperation` wrote whole-second success/failure timestamps; a blocked scheduled
round one millisecond after a healthy round therefore produced equal values, and
the shipped strict timestamp comparison stayed quiet. No new owner policy is needed.

**Correction / regression:** preserve both timestamp files and add an individually
atomic, label-free `<operation>-result.prom` gauge (`1` success / `0` failure) for
the three existing closed operation names. Outcome publication errors propagate;
symlink refusal leaves the protected target unchanged. The backup rule adds its
failure arm and keeps all legacy newer-failure/missing/deadline clauses. Old
helpers cannot overwrite this new file, so the legacy comparator is not suppressed.
Files are individually atomic, not one transaction. No backup format, cadence,
retention, alert threshold, gameplay, kernel identity or CI topology change.

Executed evidence:

- Old source: corrected-selector Go run52906 fails all nine outcome rows (no
  outcome file); scheduler run62320 establishes equal actual timestamps then
  fails at the missing outcome. Its healthy baseline creates a real encrypted
  target with a synthetic dump and the failed round blocks on a foreign entry
  without creating. The old failed assertion does not reach the retry. Exact-image
  promtool27898 separately fails the intended15s same-second alert (`got:[]`).
- Corrected focused94901 passes the nine rows and scheduler regression, including
  removal of the foreign entry and a successful retry clearing the outcome. Final
  `make test-go vet GO_PACKAGES='./operations ./cmd/deployment-backup
  ./cmd/deployment-operations ./deploymentbackup ./deploymentrelease ./releasepackage'
  GO_TEST_FLAGS='-count=1'`47316 passes all six packages and vet. This includes
  failure to publish the outcome; optional integration skips are not proof.
- `make test-go-ci CI_TEST_PACKAGES='./operations ./cmd/deployment-backup
  ./cmd/deployment-operations ./deploymentbackup ./deploymentrelease ./releasepackage'`
  87566 passes all six cold Linux-amd64/Postgres packages. Full custom-dump/restore
  tests requiring the separate backup environment are not inferred from this run.
- Exact pinned Prometheus index69f52414's amd64 childdd4bced0 was independently
  re-read from registry metadata. Its actual `promtool test rules
  /work/cloud-clicker-alerts.test.yml`47202 passes the full checked-in population,
  including equal-time fire/clear, earlier failure timestamp and the retained
  legacy newer-failure arm. No release/version substitution.
- `make test-deployment-operations-profile`45444 passes138.64s, exit0: five real
  private targets, node-exporter textfile collection, healthy backup baseline,
  equal-time failure, received firing after that injection, received resolution
  with that exact firing's start time, synthetic preflight nonce, actual cleanup
  firing, HTTP500 receiver refusal and peer Caddy-admin isolation. Notification
  overflow/truncation fails loudly; unsuccessful Prometheus responses cannot
  establish a quiet baseline. The expanded5m completion guard accommodates the
  unchanged1m grouping transitions plus45s rejection window, not an alert SLA.

Non-evidence attempts retained: initial test parser namespace typo and unquoted
Make regex did not run a valid population. The first index-based promtool attempt
exited125 on Docker's existing architecture/digest collision; the exact index's
verified amd64 child avoids it without deleting images. Initial profile61058
was sandbox-denied before Docker execution; permitted45444 executed in full.

**Coverage / handoff:** the Compose test injects controlled outcomes/times into
the actual production textfile writer; it does not run pg_dump or the scheduled
backup command. The scheduler proof is separate. Services retain the fixture's
host-selected/mixed architecture, not an all-amd64 clean release host. The target
removed only its dedicated fixture project/volume. No all-seven producer journey,
retention, clean-host R-006, whole-CI, Deployment acceptance or archival claim.
Review by: Codex (implementer first filter). Recorded by: Codex. The complete
new range after `24a343b8` through this containing commit requires Claude's
designated review; it may consolidate with related operations corrections.

## 2026-10-13 — Real Postgres collection failure reaches the operator (DP7)

**Outcome / authority:** construct the missing producer→rule→receiver proof under
Deployment DP7/AC7 and the existing reviewed database/collector rule. The private
fixture previously supplied a fake unavailable database. It now owns a pinned
Postgres16 service on its private network, no host port, with tmpfs data and only
declared fixture credentials. Its current embedded migration chain must succeed.
No production, migration, rule, threshold, kernel, provider or CI-workflow bytes change.

**Test / discrimination:** from a healthy baseline, an actual exclusive lock on
`verification_dead_letters` makes the production collector's count query exceed
its existing2s context. Ping and scrape must remain1 while collection becomes0.
The focused check requires unavailable count families to disappear, not return
stale/invented zeros, and requires healthy collection after rollback. The private
fixture additionally requires the unchanged2m rule, a received firing newer than
the fault, and resolution carrying that exact firing's start time after lock release.
The shared notification helper retains backup matching and forbids synthetic nonce
notifications from satisfying these paths.

Executed evidence:

- `make test-go-ci CI_TEST_PACKAGES='./operations'
  CI_TEST_FLAGS='-run TestDatabaseCollectorIntegration -v'`70699 passes the healthy
  current-schema and new blocked-query/recovery populations,0.12s/2.11s, no skips.
- Compiler overlay93504 maps only a task-owned copy of `operations/metrics.go`
  into the declared Linux-amd64/Postgres test service, changing query-error
  collection0 to1. The new test fails2.16s with `actual database health was hidden`,
  showing the actual falsely healthy metric. Production source never changed;
  the exact one-line copy diff was checked. Removed only the temporary source/
  overlay and empty task directory after the probe completed.
- `make test-go vet GO_PACKAGES='./operations ./releasepackage'
  GO_TEST_FLAGS='-count=1'`53874 passes both affected packages and vet. Native
  optional integration skips are not SQL/Compose evidence.
- After the fault probe, `make test-go-ci CI_TEST_PACKAGES='./operations'
  CI_TEST_FLAGS='-v'`70387 passes the complete cold package,3.149s. Both real SQL
  tests execute; optional native binaries and the separate private profile skip
  explicitly. No production restoration or cache-dependent success is needed.
- `make test-deployment-operations-profile`31383 passes330.07s, exit0. All five
  private scrape targets, synthetic nonce, cleanup firing, backup same-second
  firing/matching resolution, actual Postgres collection failure after2m, its
  received matching resolution, receiver500 refusal and Caddy-admin isolation
  execute. Its8m completion guard covers the measured138.64s prior population
  plus the new real2m hold and unchanged1m grouping transitions; no alert bound
  was relaxed. The target removes only its dedicated fixture project/volume.

**Coverage / handoff:** this is a real database/query failure, not simulated
metric text or a stopped Postgres service/network partition. The gameserver
target remains the actual registry served by a test fixture, not a release
gameserver. Service images retain the existing host-selected/mixed architecture;
Linux-amd64 Go/SQL and this private stack are not the clean release-host rehearsal.
Other producer arms, retention, seven-family end-to-end acceptance, exact release
bundles, R-006, current hosted CI and broader corrective reviews remain open.
Review by: Codex (implementer first filter). Recorded by: Codex. The complete
test/fixture/docs/plan/log range after `b4013e99` through this containing commit
needs Claude's designated review, consolidatable with the preceding operations
correction. No archival, publication, deployment or feature-completion promotion.

## 2026-10-13 — Actual database connection refusal and reconnection (DP7)

**Outcome / authority:** complete a distinct database alert producer path under
Deployment DP7/AC7. A timed-out population query is not connection-loss evidence.
The private fixture now supplies a separate maintenance database; it disables
connections to only `cloud_clicker_operations_test` and terminates that subject's
existing sessions. The same production registry/pool must report reachability0
and collection0 while its HTTP scrape stays healthy. The unchanged2m rule must
deliver a fresh firing; access restoration must produce automatic pool reconnect,
healthy collection and received resolution of that exact alert. No direct alert
injection, synthetic metrics, replacement registry or shortened rule interval.

**Focused test / discrimination:** the SQL test creates and migrates its own
generated database, observes healthy→refused→reconnected, and drops only that new
subject. The helper allowlists fixture names, requires a separate control database,
scopes termination by exact database name, and restores before caller cleanup on
failure. Both failed-health families must be0 and unavailable count families absent.
No production, migration, rule, gameplay, kernel, credential or CI-workflow bytes change.

Executed evidence:

- `make test-go-ci CI_TEST_PACKAGES='./operations'
  CI_TEST_FLAGS='-run TestDatabaseCollectorIntegrationReportsConnectionRefusalAndRecovers -v'`
  59410 passes1.19s after the helper cleanup-order correction; initial81533 also
  passed its normal path but did not verify recovery on an assertion failure.
- Compiler overlay22619 maps only a temporary source copy into the declared
  Linux-amd64/Postgres service, falsifying both ping-error health gauges from0
  to1. The test fails0.63s with `actual connection state hidden` and the two
  falsely healthy metrics. Production source never changed; the two-line diff
  was checked. Task-owned overlay/source/directory were removed after completion.
- `make test-go vet GO_PACKAGES='./operations ./releasepackage'
  GO_TEST_FLAGS='-count=1'`55706 passes both cold packages and vet. After the
  probe, `make test-go-ci CI_TEST_PACKAGES='./operations' CI_TEST_FLAGS='-v'`
  22916 passes the full cold package3.864s: all three actual-SQL functions
  execute; optional native binaries/private-profile skips remain explicit.
  Read-only fixture query9b31dc confirms zero generated reachability databases
  remain after both normal and failed runs.
- `make test-deployment-operations-profile`60688 passes511.42s, exit0. It retains
  the backup firing/resolution, actual query-lock firing/resolution, nonce,
  cleanup, receiver500 refusal and admin isolation, then executes actual
  refused connections, new firing after2m, automatic pool reconnection and
  matching resolution through the configured receiver. The new12m completion
  guard covers the completed330.07s preceding population plus the additional
  real2m hold and unchanged1m notification intervals; it is not an acceptance
  bound or CI timeout increase. The target removes only its dedicated fixture.

**Coverage / handoff:** actual Postgres connection refusal, not a service process
restart or network partition. The private stack still uses host-selected/mixed
service architectures and a registry fixture rather than an exact release
gameserver/clean host. Remaining producer families, retention, exact release
pair/rehearsal, hosted CI and corrective reviews stay open. No new defect or
owner-policy choice was required for this accepted acceptance work.
Review by: Codex (implementer first filter). Recorded by: Codex. The full new
test/fixture/docs/plan/log range after `391ae4a5` through this containing commit
needs Claude designated review, consolidatable with the preceding operations
ranges. No feature-completion checkbox, archival, push or deployment.

## 2026-10-13 — Real transport dead-letter growth and resolution (DP7)

**Outcome / authority:** construct Deployment DP7/AC7's transport-queue alert
journey without injecting metric values or alerts. Controlled compensation events
enter the current SQL trigger; the real Node, loaded with the current transport
policy, rejects oversized envelopes. The actual Store/player relay retries five
times and persists dead letters while preserving authoritative payloads. Valid
publication is required before and after failures. No production, migration,
rule, threshold, kernel, provider or CI-workflow bytes change.

**Discrimination:** the focused SQL test checks every persisted retry, fifth-
attempt dead-letter transition, no published/held claim, exact collector counts
and no seeded private material in metrics. A compiler overlay changes only the
transport collector value to 0; the test fails in 0.78 seconds at the first actual dead letter
with `actual outbox population not reported`. Production source never changed;
the one-line copy diff was verified and the task-owned overlay/copy removed.
The private fixture separately requires one isolated increase to stay quiet for 45 seconds,
then grows populations across distinct actual scrapes, receives the firing, and
receives resolution with the same start time after growth stops. All four rows
remain and later valid publication must still succeed.

Executed checks:

- Focused declared Linux-amd64/Postgres SQL run passes in 1.09 seconds. Final
  `make test-go-ci CI_TEST_PACKAGES='./operations ./transport ./releasepackage'
  CI_TEST_FLAGS='-v'` passes: operations 4.987s, transport 14.188s, releasepackage 1.818s. All four operations SQL
  functions run; optional native-service/private-profile skips are explicit.
- `make test-go vet GO_PACKAGES='./operations ./transport ./releasepackage'
  GO_TEST_FLAGS='-count=1'` passes the affected native packages and vet; native
  DB skips are not SQL evidence. A read-only query confirms zero generated
  dead-letter fixture databases remain after passes and the fault probe.
- `make test-deployment-operations-profile` passes in 643.69 seconds, exit 0, within the
  unchanged 12-minute guard. Earlier backup and both database fire/clear arms, cleanup
  firing, nonce, receiver HTTP-500 refusal and Caddy-admin isolation still execute.
  The target removes only its own disposable project/textfile volume.

**Coverage / handoff:** this is the real transport trigger/encoder/relay/SQL/
rule/receiver path with controlled upstream events, not authenticated gameplay,
socket-client delivery, verification/poison producers or a release gameserver.
Mixed service architectures, remaining producer/retention evidence, exact
release pair/R-006, current hosted CI and earlier corrective reviews remain open.
The owning plan's stale per-edit instruction is reconciled to Marco's already
adopted RFC-0000 procedure; feature acceptance and designated review are retained.
Review by: Codex (implementer first filter). Recorded by: Codex. The entire new
test/fixture/docs/plan/log range after `7d0aa087` through this containing commit
needs Claude designated review, consolidatable with the preceding operations
batches. No acceptance checkbox, archival, publication or deployment promotion.

## 2026-10-13 — Actual cleanup startup faults expose volatile alert observation (RP-519)

**Outcome / authority:** test the real composed credential cleanup under DP7/AC7,
not a substituted failing callback. Each healthy/session/bootstrap arm owns a
generated, current-schema Postgres database and uses real account bootstrap and
startup prime. SQL trigger faults propagate their exact error, refuse readiness,
and report failure once; removing each trigger permits startup and clears the
expired credentials. Session pruning is atomic; the bootstrap failure happens
after session pruning commits, which the test checks explicitly.

**Regression / evidence:** the old registry reports failure1 but a fresh registry
reports0. Command source opens HTTP only after prime succeeds. Thus the counter-only
rule cannot observe startup failure; runtime scrape-versus-exit timing is still
unexecuted, not silently promoted. A temporary compiler overlay changing only the
composed job label to `guild_sweep` fails the named SQL oracle in0.73s. Production
source stayed byte-unchanged and the task-owned overlay was removed.

**Coverage:** focused declared Linux-amd64/Postgres run passes all three cases
in2.51s. Final `make test-go-ci CI_TEST_PACKAGES='./gameserver ./operations
./cmd/gameserver' CI_TEST_FLAGS='-v'`55367 exits0; the new population executes
in2.27s, operations4.825s, command0.076s. Optional native-service/private-profile
skips are explicit. `make test-go vet GO_PACKAGES='./gameserver ./operations
./cmd/gameserver' GO_TEST_FLAGS='-count=1'`9249 passes the native packages and vet;
native DB skips are not SQL proof. Post-run SQL census confirms no generated cleanup
databases remain. Corrected docs/test commentary distinguish the previous callback
fixture from actual composed cleanup. No production/migration/rule/CI bytes changed.

**Handoff:** RP-519 remains OPEN. Preserve fail-closed startup and job shutdown;
failure delivery must survive loss of the gameserver endpoint/process. This is
accepted engineering work, not a new owner decision. Actual command-process,
journal/Prometheus/receiver and exact release-host proof remain open. Review by:
Codex (implementer first filter). Recorded by: Codex. Entire range after `f0009bfc`
through this containing commit needs Claude designated review, consolidatable with
preceding DP7 work. No feature acceptance, archival, publication or deployment.

## 2026-10-13 — Preserve cleanup failure observation after endpoint loss (RP-519, DP7)

**Outcome / authority:** repair DP7/AC7's startup observation gap without changing
cleanup, readiness or fatal-job policy. The real command uses the bounded logger
registry; errors emit fixed class/job fields before returning, never underlying
errors or player/credential material. The existing host observer streams its exact
journal directory over the retained14-day population, persists a newly discovered
cursor hash/detection time, and publishes an independent atomic textfile. Seen
failures do not extend the5m alert window; late catch-up is observable. Metrics
publish before the cursor is consumed. Missing/stale collection pages; truncated,
partial, malformed, nonzero-exit, timeout or exit-zero diagnostics cannot refresh
health. The operator's existing serialized oneshot remains the collector owner.

**Regression / executed evidence:**

- Actual composed session-delete/bootstrap-tombstone SQL faults now also require
  the bounded event before startup returns; healthy/recovered startup emits none.
  Focused declared Linux-amd64/Postgres28041 passes2.03s, and full affected53684
  passes gameserver88.236s/operations4.856s/command0.077s/operations-command5.250s/
  releasepackage1.283s. All real SQL cases execute; optional-service skips are explicit.
- Removing just the log emission in a compiler overlay makes the actual session
  startup oracle fail0.75s. Reverting only command construction to its old discard
  registry makes the actual command-startup oracle fail0.85s. Main sources never
  changed for these probes; temporary copies/overlays were removed.
- The actual development-command startup test first failed before its SQL subject:
  my fixture's port0 was correctly refused. A reserved valid loopback address fixes
  setup and prevents accidental serving. Focused90801 passes0.74s; final whole
  command23977 passes0.967s with the new case0.88s. This is `run`, not an OCI process.
- The pinned rule population passes. A scratch old counter-only rule fails the new
  startup-without-counter and absent/stale collection cases. Forced Linux-amd64
  evaluation first fails before tests because Docker cannot overwrite the cached
  manifest's ARM64 image. The explicitly disclosed cached ARM64 evaluator passes,
  including the final summary clarification; no supported-amd64 rule claim from it.
- Full private stack84031 passes700.28s, exit0, within the unchanged12m guard:
  actual logger event→modelled envelope→decoder/textfile→real node-exporter→shipped
  rule→receiver firing and matching resolution, while the fresh process counter0
  remains0. The backup, both database-failure/recovery, dead-letter, nonce, receiver500
  and admin-isolation populations still execute. Its dedicated services/volume are
  removed; post-run inspection is empty. Envelope/timer are modelled, not journald.
- Affected native/vet82340 passes all five packages. Final command-observer49876
  passes5.551s on Linux-amd64, native/vet26622 passes, including a controlled child
  emitting valid output then exit-zero exclusion diagnostics. No healthy refresh.
- Updating the optional native fixture exposed my gzip/plaintext concatenation bug.
  A read-only live Prometheus target query reports `gzip: invalid header`; that owned
  run73078 is stopped as invalid, not credited. Retained ordinary HTTP regression57069
  fails with the same error; the corrected fixture serves a complete plaintext
  response when combining its two producers. Native/vet55375 passes; final cold
  Linux-amd64 operations52107 passes4.909s including that regression and all SQL tests.
  The earlier direct test-binary invocation4493 could not find package-relative config
  and ran no journey. Corrected package-selector19688 passes321.51s within the unchanged
  8m guard: zero at6.403s, failure1/firing at21.522s, natural resolution at5m21.358s with
  counter still1. It uses Linux-amd64 tests plus cached pinned ARM64 service executables,
  not native macOS/all-amd64 release proof. Extracted program SHA256s: Prometheus
  `9016b054ba86f89ad32880084fefb2ca34a457ed4e0e586e1ebb032e6ed5931e`, Alertmanager
  `15c74ada5a4677236016a13f24d430dd3126a14a940c071a2c70799fa1e30f03`.
- Final diff inspection extends the invalid-state cases: preinitialized schema1
  accidentally admitted `{}` without a version. Regression88959 fails at that input;
  decode now starts present files from zero state and refuses missing versions. This
  changes only invalid-state handling, not the valid private/native journeys above.
  Final native/vet94695 passes all five affected packages; final cold Linux-amd64
  operations42286 passes6.073s, including invalid-state/compression controls and
  all four actual-SQL functions. Optional-service skips there are explicit.

**Coverage / handoff:** Docker's documented tag/partial fields were checked against
[the driver documentation](https://docs.docker.com/engine/logging/drivers/journald/) and
[primary source](https://github.com/moby/moby/blob/master/daemon/logger/journald/journald.go).
[journalctl source](https://github.com/systemd/systemd/blob/main/src/journal/journalctl-show.c)
confirms no-match grep is nonzero and corrupt-record exclusions can warn; the collector
uses quiet complete JSON without grep and rejects diagnostics. These are source checks,
not host execution. Current-schema SQL, command wiring and private receiver behavior
are locally proved; actual release-process/Docker-journal delivery, rapid-exit timing,
retention, remaining alert producers, rebuilt exact release pair/R-006, hosted CI and
earlier review debt remain open. No new owner policy, retry, threshold, kernel, migration
or CI change. Review by: Codex (implementer first filter). Recorded by: Codex. Entire
repair range after `a38bcb9c` through this containing commit needs Claude designated
review; consolidate with the diagnosis after `f0009bfc` and earlier DP7 ranges. No
feature acceptance, archival, push, deployment or full1.0 claim.

## 2026-10-13 — RP-520: distinguish orderly job cancellation from cleanup failure

**Outcome / authority:** DP5's normal drain and DP7's genuine failure alert must
coexist. The composition now supplies the job-lifetime context to observation.
An error caused solely by that context ending increments the closed `canceled`
result, preserving the error and last success without a failure log. SQL/other
errors still fail, including joined cancellation plus a real error. Existing
context-free observers keep their failure semantics. No drain, cleanup, migration,
timer, retry, threshold, kernel or CI-workflow change.

**Regression / checks:** old source `5fcb5353` with the new observer-adapter test
fails the false failure counter (39504). The actual one-minute Postgres timer lane
(31179) passes runtime SQL failure but fails orderly drain in121.63s: the scheduled
DELETE entered its real table-lock wait and returned `context.Canceled` after drain,
yet failure count became1. Corrected33117 passes both arms in121.85s: runtime
P0001→failure channel/readiness503/bounded log; fault removal→fresh successful
cleanup; drain cancellation→canceled1/failure0/no alarm/unchanged credential rows.
The spy forwards the real context/error; interval remains60s. Shutdown after a
reported job error is tested as completion of shutdown, not replay of that error;
the failure channel is its actual error authority.

Focused44618 passes the adapter and ten observation cases, including wrapped/joined
cancellation, deadline, real/joined SQL error, live-context cancellation, mismatched
error, same-message impostor and completed success. An initial Make selector22741
ran no tests because `$` was expanded; it is not evidence and was corrected before
the recorded old-source failure. Full cold native33451 passes operations1.561s,
gameserver0.371s, command0.305s, operations-command5.469s; native DB/service skips
are not credited as integration. Vet passes. Full declared Linux-amd64/Postgres17528
passes4.917s/105.575s/0.968s/5.595s respectively, including actual startup SQL and
command witnesses. The manual timed case explicitly skips in that ordinary lane;
its separately executed33117 is the timing evidence. Disposable cleanup DB census0.
`git diff --check` passes.

**Coverage / handoff:** fast regression stays in ordinary package/CI execution;
new `make test-deployment-cleanup-runtime` is included in manual Deployment
operations verification, not push CI. Existing private receiver/rules/journal
fixtures are not rerun or newly credited here; this verifies observation and real
server/SQL lifetime, not OCI exit, host journald, release-host timing/retention,
rebuilt exact release pair/R-006, hosted CI or full AC7. Review by: Codex
(implementer first filter). Recorded by: Codex. Entire range after `5fcb5353`
through this containing commit needs Claude designated review; consolidate with
preceding DP7 ranges. No feature acceptance, archive, push or1.0 claim. Continue
remaining accepted deployment evidence and construction; owner-policy questions
do not block that work.

## 2026-10-13 — DP7 actual bounded-filesystem alert delivery; RP-521 input repair

**Outcome / authority:** build DP7/AC7 storage producer→receiver evidence with
actual allocation, not typed usage constants. The original host-observer syscall
reader moves verbatim to `operations.FilesystemUsage`, shared by the command and
private fixture. Distinct empty20MiB tmpfs subjects limit all writes. Each actual
16MiB allocation produces exactly80% and stays quiet through two evaluations;
17MiB produces85% and reaches the configured receiver. Removing only the owned
file recovers0 and delivers resolution of the exact alert/start/storage label.
No threshold, interval, retention, gameplay, migration, kernel or CI change.

**Regression / checks:** while inspecting that writer, both old NaN-input cases
fail7475 (0.09s test), while infinities already refuse. Explicit NaN rejection
now preserves the previous valid `host.prom` bytes and mtime. This is an input
validation defect, not a reproduced nonfinite filesystem syscall result.
Native99220 passes operations1.595s/command5.464s and vet passes; cold declared
Linux-amd64/Postgres80738 passes4.964s/5.555s. Final native26014 passes1.694s/
5.476s and vet passes; final Linux-amd64/Postgres65070 passes4.874s/5.600s,
including actual SQL collectors and command-journal negatives. All six nonfinite
cases pass. Optional private/native-service skips there are not integration credit.

Focused private storage24913 passes286.35s (subjects119.64s/119.99s). After a
test-diagnostic adjustment directs errors to the current subtest, final default
46290 passes both named populations: existing profile681.05s, storage644.86s
(subjects119.86s/150.01s), exit0. Existing backup, query lock, connection refusal,
dead-letter, cleanup, nonce, receiver-rejection and admin-isolation assertions
remain and execute. The shared run's storage setup is substantially longer than
the standalone run; setup sub-phases are not timestamped, so no cause, latency
claim or budget adjustment is inferred. Each population's12m guard is unchanged.
The private fixture uses cached pinned services and an amd64 Go binary; its Caddy
platform warning is disclosed, not an all-amd64/clean-host claim. Make removes
only its own services/network/textfile volume. `git diff --check` passes.

**Coverage / handoff:** default manual profile runs both distinct test names;
`make test-deployment-storage-profile` focuses only storage, never masquerading
as the full population. Fast invalid-input/reader checks remain ordinary tests.
Restart/journal fields are controls here; actual host-observer process, Docker
inspection/journald, journal pressure/retention, remaining producer journeys,
rebuilt exact release pair/R-006, hosted CI and prior review debt remain open.
Review by: Codex (implementer first filter). Recorded by: Codex. Entire range
after `82c8e57b` through this containing commit needs Claude designated review;
consolidate with preceding DP7 work. No acceptance, archive, push, deployment
or1.0 claim. Continue accepted operations/release construction.

## 2026-10-13 — DP7 first-scrape restart detection (RP-522)

**Outcome / regression:** DP7 requires three restarts inside ten minutes. The
committed counter-increase rule misses a new container already at three on its
first scrape. The full retained population against `db752b0a`'s rule in an
isolated scratch fixture fails80601 at30s/10m15s (expected firing, got none).
The production observer now inspects count and creation time together, rejects
invalid/missing/future metadata, and publishes a bounded birth gauge without
container IDs. Preserve the increase arm; add only count>=3 with known nonfuture
container age<=600s. Lower counts, unknown/future birth and age>600s stay quiet.
Older pre-scrape history cannot be reconstructed from cumulative count alone.
No threshold, cadence, journal policy, migration, kernel or CI-workflow change.

**Checks:** pinned complete promtool population9683 passes; cached exact-index
image is Linux/arm64, not an amd64 rule-evaluator claim. Native46810 passes
operations1.598s/command5.487s and vet passes. Declared Linux-amd64/Postgres50634
passes5.023s/5.744s. Ordinary lanes explicitly skip the opt-in Docker-engine
test; separately executed `make test-deployment-restart-observer`75344 passes
46.34s: real automatic restart count0→3, actual inspected birth, production
parser/filesystem/textfile publication and no emitted container identity.
The earlier fixture lookup adapter matched the wrong argument positions and
failed before the restart objective (22417); it was corrected, not counted as
evidence. Lost earlier process handles have no inferred result; the final checks
above were executed fresh. Exact owned-container census is empty after cleanup.
Formatting/diff checks pass; Make dry-run confirms the observer lane is included
in manual operations verification, not push CI. Aggregate is not newly rerun.

**Coverage / handoff:** the real engine fixture is a scripted exiting process,
not gameserver. Official-project lookup and journal source are controlled;
Docker inspection, filesystem reads and publication are real. This and the
rule regression do not prove live receiver delivery of the new arm, actual
journald/retention, older pre-scrape history, release-host/R-006 or full AC7.
Review by: Codex (implementer first filter). Recorded by: Codex. Entire range
after `db752b0a` through this containing commit needs Claude designated review,
consolidatable with the preceding DP7 batches. Continue accepted operations and
release-artifact construction; no acceptance, archive, push or1.0 promotion.

## 2026-10-13 — Actual restart observation reaches the private receiver (DP7/AC7)

**Outcome:** extend the existing native Docker producer into the existing private
operations fixture. Start monitoring only after real automatic count3 publication;
require at least two scraped samples, minimum3/increase0, shipped-rule firing and
the actual receiver payload. Only after delivery, create a new stable container
and run the production observer again: real count0/later birth, a fresh atomic
file, inactive rule and resolution with the exact original alert start. No typed
restart metric, synthetic family notification, monitoring reset or forced resolve.
A per-test bind overlay replaces only the textfile mount; no Docker socket enters
the private stack. Production/rules/intervals/CI bytes are unchanged.

**Regression / correction:** first actual attempt31033 fails600.53s at its original
ten-minute parent guard. Live queries prove count3/minimum3/increase0 and firing;
the preflight loop consumed the earlier restart notification while seeking its
nonce (RP-523). Retain unrelated notifications within the existing capacity32,
refusing private fields, closed/canceled observation and backlog overflow. Unit
preservation/failure/overflow cases pass72115. Do not shorten the four-hour repeat
interval or manufacture another firing. That failed attempt also exposed our new
runner cleanup omission: killing the Compose CLI leaves its one-off alive; down
leaves the occupied network. The runner self-exited before manual removal, and
only its subsequently verified empty project network was removed. Cleanup now
validates exact project/service/one-off labels before removing that runner and
its anonymous volumes, then downs only the unique project. Other projects remain.

**Finished checks:** corrected first delivery66916 passes126.04s (child76.52s).
Final root `make test-deployment-restart-profile`13739 exits0 after helper extraction:
canceled-CLI negative7.70s, complete delivery134.44s (child82.28s), package142.485s.
The negative actually leaves its sleep runner live after CLI cancellation, then
requires zero project containers/networks after cleanup. Final cold native19517
passes operations2.137s/command5.528s and vet; declared Linux-amd64/Postgres6772
passes5.193s/6.482s, including bounded notification regressions. Those ordinary
lanes skip the opt-in real-Docker populations; the separate root run is their
evidence. An intermediate cleanup extraction had a stray closure caught by
gofmt before any run; repaired, not hidden as a passing attempt. Final fixture
container/project/network censuses are empty and diff checks pass.
No budget/capacity/threshold increase.

**Coverage / handoff:** actual Docker, host-command inspect/parser/syscalls/writer,
private node-exporter/Prometheus/Alertmanager, real SQL-backed metrics and receiver;
scripted exiting process with controlled project lookup/journal, not the game
binary, actual host driver/timer/retention, older pre-scrape restart history or
exact release artifact. Cached Caddy's amd64-on-ARM warning remains disclosed;
not an all-amd64/clean-host proof. The developer filesystem's actual high usage
also produced unrelated storage alerts; receiver matching is family/start-specific,
not an allocation or supported-host storage claim. Existing full/storage profile
populations and the whole manual aggregate were not newly rerun; their literal
selectors remain, with unchanged tests/intervals. The manual aggregate now selects
the combined restart journey, retaining the producer-only focused selector.
Review by: Codex (implementer first filter). Recorded by: Codex. Entire range after
`cdc78050` through this containing commit needs Claude designated review; consolidate
with preceding DP7 work. No acceptance, archival, push, deployment or1.0 promotion.
Continue the remaining genuine producer and release-host boundaries.

## 2026-10-13 — Actual scheduled backup reaches the receiver and restores fresh data

**Outcome / authority:** DP6/DP7 actual execution, reusing the private fixture and
production CLI. No injected backup metric, clock or dump in this named population.
Create a real encrypted Postgres16 baseline; observe its actual completion; add an
owned foreign entry; run the real schedule command. Its target refusal must reach
the receiver without changing baseline/obstruction bytes. After delivery, update
one seeded account row and remove only the obstruction. The same worker waits its
unchanged five-minute retry, creates a new encrypted package, resolves the exact
receiver alert and restores the updated row into a newly created database using
real pg_restore16. No round callback, forced resolution or faster retry. Existing
12m child/parent guards remain; 14m Go timeout allows supervisor cleanup only.

**Checks:** first execution11372 passes376.07s (child373.04s). Final root execution
21704, with explicit obstruction-preservation and named-PASS collection guards,
passes374.04s (child370.10s), exit0. Empty, skipped, wrong-population and log-only
PASS inputs are refused by the cold unit guard. Existing real Postgres backup/
restore population8984 passes all seven top-level tests, including populated
identity, wrong/trailing package, clean-target and snapshot-consistency cases.
Backup metadata builders moved unchanged into `internal/backupfixture` for reuse;
the current migration is read from actual SQL, not a stale fixture version.

Final affected cold native/vet72114 passes operations1.637s, operations-command
5.455s, backup-command0.398s and backup0.607s. Declared Linux-amd64/Postgres51681
passes5.166s/5.569s/0.149s/4.022s; private opt-in skips there are not joined proof.
Initial17614/57698 compile attempts failed on my confusing PackageMetadata's dump
hash with Header's payload hash and a moved unused import; no private execution
ran. Corrected field checks the package, while actual restore owns the stronger
new-data assertion. An initial apply-patch hunk-order failure changed no files.
The final post-run cleanup diagnostic wording edit does not change execution.
All handles are terminal; the final project container/network/volume censuses
are empty, target database cleanup passed before the child verdict, and diff
checks pass. Existing warned Game UI orphan remains untouched.

**Coverage / handoff:** actual production backup CLI, pg_dump16, encrypted package,
scheduled refusal and physical five-minute retry, real private exporter/rules/
Alertmanager/receiver and actual pg_restore with selected SQL data. Manifest/image
hashes are the existing synthetic fixture metadata, not a built-release bundle;
the local stack is mixed-architecture (cached Caddy warning disclosed). No next
six-hour slot, off-host durability, RPO/RTO, metadata-authentication adoption,
host journal/retention, whole manual aggregate or hosted CI/clean-host proof.
Production code, schema, policy, cadence, bounds, content and CI remain unchanged.
Manual aggregate includes the new focus lane; other selectors remain unchanged.
Review by: Codex (implementer first filter). Recorded by: Codex. Entire new range
after `c4279b66` through this containing commit needs Claude designated review;
consolidate with the preceding DP7 work. No acceptance, archival, push, deployment
or1.0 promotion. Continue the remaining real game/host and release-artifact gates.

## 2026-10-13 — Physical Caddy outage fires and resolves through the receiver (DP7)

**Outcome:** reuse the private fixture with the shipped Caddyfile. Start with a
successful proxied health response and healthy scrapes. Validate the exact owned
project/service labels, physically stop Caddy, require an unreachable proxy route
and real caddy-up0 while upstream scrape/readiness remain1. Record real Prometheus
pending time, observe no firing through4m45s and reject an observed firing before
5m; require the actual receiver's Caddy-labelled firing. Only then restart the
same container and require route/scrape recovery plus matching alert resolution.
No injected proxy metric, changed rule, shortened interval or forced resolution.

**Checks:** root `make test-deployment-ingress-profile`72356 passes406.67s, with
child403.37s, exit0. Native controller executes both physical actions in order
and verifies process state; direct intermediate inspection confirms stopped/
exited Caddy. Named child PASS is required, not an empty/skipped selection.
Shared collection guard retains its invalid cases. Cold native11867 passes
operations1.838s/command5.479s and vet; declared Linux-amd64/Postgres5075 passes
5.003s/5.556s. Ordinary private opt-in skips are not integration credit. All
handles terminal; final owned-project container/network/volume censuses are
empty and diff checks pass. Other projects/orphans remain untouched.

**Coverage / handoff:** actual Caddy process/config/route/scrape, shipped rule,
real private Prometheus/Alertmanager/receiver, lower-duration control and matched
recovery. Upstream health204/readiness1 are fixture controls, not game-server
readiness, public DNS/TLS, browser gameplay or clean-host/release-artifact proof.
Cached Caddy's amd64-on-ARM warning remains disclosed. No production, schema,
content, policy, CI, budget or notification changes. Manual aggregate selects
the new lane; its other populations/whole aggregate were not rerun here.
Review by: Codex (implementer first filter). Recorded by: Codex. Entire new range
after `12399dc9` through this containing commit needs Claude designated review,
consolidatable with preceding DP7 work. No acceptance, archival, push, deployment
or1.0 promotion. Continue the remaining genuine game/host observation and
release-artifact boundaries rather than relabel these local fixtures as release.

## 2026-10-13 — Bind the existing DP-E verdict to its history correction

**Review by:** Claude, as identified in the 2026-09-24 designated cross-party entry.
**Recorded by:** Codex (binding transcription only, not a new reviewer).
**Decision:** **CHANGES REQUIRED** — unchanged from that original DP-E verdict.

The completed review's exact range is `3f58fea..65099e7`, containing the offending
implementation (`09d5027^..09d5027`). Its reported cold execution explicitly
reproduces the missing kernel bump and the socket defect before the corrective
range. The original entry attributes inspection to a Claude review subagent and
recording to the Claude parent; that provenance is retained, not replaced by a
claim of fresh Codex inspection or approval.

This addendum supplies the single-commit anchor required by the unchanged history
correction format, using the existing completed review instead of the August24
implementer first-filter paragraph. It edits neither that history nor the
append-only correction row. It approves no DP-E finding/remediation, current
Deployment closeout, new CI guard or release. The old verdict and its executed
evidence remain in their original section above.

## 2026-10-13 — Complete original Claude R3 diff-review disposition

**Review by:** Codex (designated cross-party reviewer of the original Claude range).
**Recorded by:** Codex. **Range:** `2b6c0cb4^..2b6c0cb4`, all five changed paths:
scanner, scanner tests, rehearsal secret producer, canonical Deployment docs and owning log.
**Decision:** **CHANGES REQUIRED**, retaining the earlier RP-135 finding; no descendant approval.

Inspected the complete recursive gzip/tar opener, errors/guards, payload/name assertions,
exclusive synthetic probe-file lifecycle, accompanying coverage claims and historical process
disclosure. The body/gzip correction is real, but the original recognizer misses valid V7 tar
(the earlier executed RP-135 record), and the inherited scan branches omit owner/group/PAX/xattr
and gzip-header metadata. Its scanner-path rendering can also include recognized secret names.
Current digest-valid RP-543 regressions below execute that surviving omission; they are NOT
relabeled execution of the complete historical tree. This completes the original diff
disposition, not all original acceptance experiments, current-image safety, DP-B or R-006.
Later Codex repairs remain separately subject to Claude review; other bounded approvals stand.

## 2026-10-13 — RP-543 archive metadata scanning and secret-safe diagnostics

Outcome under DP3/AC2: inspect outer/nested tar names/links, owner/group, PAX/xattr keys/values
before non-file skipping, and gzip name/comment/extra before recursion. Use existing patterns,
types and bounds. Redact entire secret-bearing finding/member-error paths, not just a marker
that could leave key material. Preserve ordinary locations, payload detection and named probe
results. No public result schema, credential, kernel, workflow, limit or archive-byte change.

Regression95601 on `cf8c6f08` production scanner:31 metadata-only sentinels are missed,
two finding paths and two unreadable-compression errors expose the synthetic name; six controls
pass. Each image fixture first passes actual Docker archive/config/layer-digest validation, so
parser refusal cannot masquerade as secret detection. Initial82679 predates the extra outer-
directory and separate error cases; final95601 is the complete old-source population.

Corrected8039 passes that population. Final supplement adds direct-gate path redaction;
91004 passes all42 metadata/image/error subcases plus the complete cold releasepackage,
deploymentrehearsal and both scanner/rehearsal CLI packages, with vet. 63466 was the preceding
passed package union, not the added direct-gate result. Existing gzip/V7/body/malformed and
named seeded-source/image producer checks remain. Diff/format/whitespace checked.

Both retained102face6 archive copies still hash to
`0ffddcec5503ff6b20691a4a1039057e7eb3ad502e13b65cebc77033fc2e0d4b`, matching its committed build
record. Ordinary root command:
`make release-secret-scan GAMESERVER_IMAGE_ARCHIVE='.cache/release-current-102face6/bundle-a/images/gameserver.docker.tar'`
passes2163 tracked/indexed working files and that actual historical image with no findings.
New tests are staged so the source population
includes them. This dirty-worktree developer mode writes no R-006 artifact and does not pair a
current source claim with an old image. The final same command also passes2163/no findings
after the remaining record edits; neither run produces a structured release evidence artifact.

Limits: recognized-pattern/component and historical-image evidence, not every credential
encoding, current rebuilt package, clean-host install/restore, Docker runtime, SQL, hosted CI,
full Deployment or1.0. Historical bundle bytes/tools are unchanged and lack this new scanner;
rebuild/re-scan the final reviewed artifact. No secrets accessed outside repository/artifact
scope or printed; no daemon, container, publication or deployment operation.

**Review by:** Codex (new repair's implementer first filter). **Recorded by:** Codex.
Exact new correction range starts after `cf8c6f08` through this containing source/tests/docs/
records commit and needs designated Claude review, consolidatable with RP-135. The original
R3 verdict approves none of this repair. All named handles terminal; no acceptance or archival.
Continue the remaining original reviews and authorized product/operations/artifact construction.

## 2026-10-13 — Complete original Claude R6 diff-review disposition

**Review by:** Codex (designated cross-party reviewer of the original Claude implementation).
**Recorded by:** Codex. **Range:** `1833882c^..1833882c`, all six changed paths: target classifier,
scheduler, both test files, canonical docs and owning log. **Decision: CHANGES REQUIRED.**

Inspected the entire target classifier/type/name/age checks, blocked/deferred/created scheduler,
five-minute retry and six-hour success cadence, checksum witnesses and associated claims.
Retain the previously executed RP-136 finding: unequal-length payloads did not isolate SHA-256
checking. RP-544 adds a production safety gap: age alone admits deletion of a still-live creator's
temporary. The unchanged original target logic is also present at `ca447daa`; the current-source
regression below executes it there, not the full historical tree. Existing foreign/corrupt
refusal and active deferral are real improvements, not blanket R6 acceptance. No new checksum
severing is needed; the older equal-length correction and its separate probes retain their own
pending Claude review. This original negative verdict approves none of the later Codex repairs.

## 2026-10-13 — RP-544 exclude live writers from stale backup cleanup

Outcome under DP6/R6: age remains eligibility, not writer-death evidence. Create acquires a shared
nonblocking directory lock before any temporary and holds it through final cleanup. Stale cleanup
binds paths to the inspected directory and acquires exclusive ownership before the existing
name/type/age rechecks. No lockfile, backup-format/JSON-list change or retuned limit. Shared
ownership preserves concurrent creators and their existing atomic no-replacement contract.
Busy cleanup/creation produces a deferred `target_in_use` scheduler result, no false success/
failure metric and the unchanged retry cadence. Other lock errors fail, never degrade quietly.

Old-source regression62878 (`ca447daa`) uses an actual age-encrypted writer paused at its dump
read, controlled mtime older than one hour, and a second OS process invoking the real cleanup.
It fails: `refused=false remains=false`, then the real writer cannot finish. The prior41613
handle was absent on resume and its final result unavailable; it is not claimed as evidence.
The corrected child requires the exact busy error, not just any failure. Corrected99161 passes
the cross-process case, actual authenticated bytes/header restore, subsequent genuine orphan
removal and the unchanged nested/concurrent no-overwrite population.

Final62181: complete cold deploymentbackup, backup CLI, deploymentrelease and deploymentrehearsal
packages pass; corresponding vet passes. 13090 passes focused race execution of cross-process
cleanup, six inspected-target refusals, creator exclusion/failed-dump ownership release,
concurrent no-overwrite and both scheduler ownership directions. The live-writer scheduler
test checks no new creation/no metrics while deferred, restores its exact encrypted bytes, then
executes a successful real encrypted scheduled retry. Six rechecks retain freshened, symlink,
directory, foreign-name, outside-target and unbound-state entries. All fixture deletions are
narrow test-owned paths; no real backup/container cleanup or restart occurred.

The actual CreatePackage tar-producer/pipe boundary also returns the precise ownership refusal
without blocked writer completion or partial files, then retries and authenticates/extracts
the complete dump successfully. These final runs supersede earlier83703/71856 package and
79589 race passes for the added package case. Formatting/diff/whitespace checks pass.

75934 root Linux-amd64 helper build passes into ignored `.cache/bin/deployment-backup-linux-amd64`.
Cross-compilation is not Linux filesystem/runtime evidence. Native integration-only execution
explicitly skips all seven Postgres tests for missing declared DB URLs; no DB restore or R-006
claim. Locks coordinate supported helpers, not privileged uncoordinated writers; validate the
actual supported filesystem and release artifact in the Linux/clean-host population.

**Review by:** Codex (repair implementer first filter). **Recorded by:** Codex.
New source/tests/docs/record range starts after `ca447daa` through this containing commit and
needs designated Claude review, consolidatable with the earlier RP-136 corrections. R6, full
Deployment, hosted CI and full nine-tier1.0 remain open. No owner choice, version, acceptance,
archival, push or deployment implied. All named execution handles above are terminal.

## 2026-10-13 — Complete original Claude R8 diff-review disposition

Review by: Codex (designated reviewer of original Claude implementation). Recorded by: Codex.
Range `1f048de9^..1f048de9`, all seven paths: metadata command/tests, metadata model/tests,
bundle-test formatting, canonical docs and owning log. **Decision: CHANGES REQUIRED.**

Inspected full map discovery, exact package-root resolution, installed identity/license reads,
deduplication, 0BSD/ISC classification and whitespace normalization, SPDX allowlist change,
fixture assertions and recorded limits. Keep RP-137's missing-map finding and later CSS/URL/
worker gaps (RP-164–166); their repairs are separately pending review. The original command's
deduplication also skips later copies' attribution; unchanged logic at `fa2bec58` is executed
below as RP-545, not relabeled a historical-tree run. R8 replaced the existing real Go-graph
guard with client fixtures; this batch restores that test without claiming its two assertions
prove each binary independently. Editorial bundle-test gofmt does not change behavior.
Original improvements are real, but neither the original range nor corrected descendants are
approved by this negative disposition. No new license-policy or legal determination is made.

## 2026-10-13 — RP-545 validate every shipped package copy

Under accepted DP1/AC8, move each resolved npm directory's existing license/metadata validation
before identity deduplication. Permit one output row only for matching detected license and
extracted notice text. Refuse ambiguous attribution for the same name/version rather than drop
another copyright notice. Distinct versions retain separate notices. Preserve classifier,
allowlist, package URLs, output schema and source-map/CSS graph semantics; no kernel/CI change.

Old-source40287 (`fa2bec58`) fails twelve bad second-copy cases across actual Go JavaScript-map
and CSS-graph consumers: missing, empty, unknown, metadata-mismatched, different same-license
notice and different valid license. All four identical-copy/distinct-version controls pass.
Corrected3503 passes the full affected metadata/release packages and vet. Final23616 repeats
those cold checks with all16 cases and the restored real Go `list -deps` helper-only age/
chromedp guard; all pass. The controls invoke the actual delivered-notice generator, asserting
one identical row or both distinct-version notices. Tests use owned temporary package trees,
not installed dependency mutations. Formatting/diff checked; no extra mutation needed for
this old-behavior regression.

Fresh root `make build-client` passes217 modules. Actual root metadata generation from that
working build passes53 dependencies; parsed SBOM and generated notices contain exactly
notations1.6.0, break_infinity.js2.2.0, pad-end1.0.2, svelte5.56.8 and tslib2.8.1 with their
existing MIT/0BSD labels. Diagnostic output `/tmp/cloud-clicker-metadata.AXLs2X` uses version0.1.0,
base source `fa2bec58c323b9e85bbad419981566a0b319457b` and fixed creation time
`2026-10-13T00:00:00Z`; the generator correction is uncommitted during execution. It is not a
release artifact/source-manifest claim. SHA256 SBOM
`55bdd505894439cc2fcea8bcbd0b0cae77608e0d42644af0db4d8ca5af84710e`; notices
`7765454957ce365749015dbbf0ec2493abaaa5aa8b181d553b5a56be520be553`.

Docker version90680 receives no server reply; only that owned diagnostic client is interrupted
(exit130, client28.4.0 metadata). No daemon/container restart, cleanup, Linux/Postgres execution
or clean-host gate is claimed. Existing backup runtime/R-006 and full AC8/release gates remain.

Review by: Codex (new repair implementer first filter). Recorded by: Codex. New source/tests/docs/
records range starts after `fa2bec58` through this containing commit; designated Claude review
pending, consolidatable with prior R8 map/CSS corrections. No original/corrective acceptance,
archival, publication or full nine-tier1.0 completion. All named handles are terminal.

## 2026-10-13 — Complete original Claude R9 diff-review disposition

Review by: Codex (designated reviewer of original Claude implementation). Recorded by: Codex.
Range `8ff555f0^..8ff555f0`, all five paths: Docker runtime/tests, smoke observer, canonical
docs and owning log. **Decision: CHANGES REQUIRED**, retaining RP-138; no descendant approval.

Inspected the complete readiness poll, drain derivation and its production caller, staged
identity comparisons, eight derivation negatives, HTTP controls and accompanying evidence
claims. Its pure derivation/identity tests are useful, and the documentation honestly separates
process-contract attestation from direct observations. An unmarked503 still cannot prove orderly
drain in the original range (the earlier executed RP-138 finding). The unchecked request
construction also survives at current baseline, reproduced below as RP-546, not historical-tree
execution. Test call-site search finds no actual `DockerRuntime.DrainCurrent` invocation: only
interface-fixture implementations, with the Caddy test separately testing server draining.
RP-547 retains this outer-orchestration evidence gap; pure derivation cannot cover disconnected
caller wiring. Current corrections, native checks and old Caddy evidence are not original-range
approval or exact release-artifact/clean-host proof. No accepted ruling is edited.

## 2026-10-13 — RP-546 fail malformed HTTP endpoints before observation/commands

Under DP3/DP5, check the configured receiver's request construction in runtime normalization;
handle contextual construction errors in preflight and readiness polling. Invalid input yields
ErrInvalid or no readiness evidence, never a nil-request panic. Preserve all constructable URL,
receiver/origin, timeout/poll, status/marker and error policies. No player/schema/kernel/CI change.

Old-source51933 (`39019517`) fails all eight malformed URL cases: bad escape, IPv6 bracket,
port and control character in each receiver/readiness path. The old receiver path reaches the
unchecked request after its three preflight commands; the correction asserts zero runner calls.
Corrected92711 executes all new negatives, the ordinary receiver preflight, staged identity,
drain-evidence table, real-localhost marked/unmarked/proxy/vanished readiness and gameserver
admission/committed-work/job/socket ordering controls. Its deploymentconfig selector executes
zero tests, explicitly not a configuration-suite result.

64306 passes all six affected cold packages/vet: deploymentrelease, deploymentconfig,
deploymentrehearsal, both release/rehearsal CLI packages and gameserver. 56567 focused race
passes the eight new negatives and readiness HTTP matrix. Supplemented the existing matrix
with a genuine ready204 success; final38709 repeats the full cold six-package/vet population
and passes with that control. Focused27166 also passes the supplemented HTTP matrix under
the race detector. Linux-amd64 root helper build passes into ignored
`.cache/bin/deployment-release-linux-amd64`; cross-compilation is not Linux runtime evidence.
Formatting/diff checked. Old-behavior negatives discriminate without another source mutation.

Boundary: live localhost HTTP for readiness; controlled receiver and command runner;
gameserver unit doubles, not actual Compose/process exits, SQL or clean-host recovery. Native
DB/declared-integration prerequisites do not execute here. Docker's last owned probe ended
with no server reply; no repeated stalled command, daemon restart, cleanup, push or deployment.

Review by: Codex (repair implementer first filter). Recorded by: Codex. New source/tests/docs/
record range starts after `39019517` through this containing commit; designated Claude review
pending, consolidatable with the preceding RP-138 correction. R9/full Deployment/full nine-tier1.0
remain open. Next: build RP-547's actual outer-method protocol population, preserving R-006's
real Linux/Caddy/Postgres/release-artifact requirements and outstanding cross-party review.

## 2026-10-13 — RP-547 actual drain-method protocol coverage

Under DP5/AC3, invoke the real `DockerRuntime.DrainCurrent`, not a fixture implementation or
just its evidence derivation. Physical localhost HTTP handles bootstrap/state/ready reads;
physical WebSocket handles correlated connect/world-subscribe acknowledgements, the actual
restart publication and closure. Docker commands are recorded, never executed. Check bootstrap
credential propagation, exact smoke→container→stop→readiness→inspection ordering, observed
container identity and the unchanged default20s stop argument.

Eleven cases: complete control; unmarked/wrong readiness; missing courtesy; wrong close code;
unclosed socket; stop/inspection failure; nonzero exit; bootstrap/state constants mismatch.
The latter two must return no evidence and issue no runtime command. Each observable and the
three exit-attested fields must match its actual fixture inputs; invalid evidence must return
ErrInvalid. Every upgraded peer is cancelled/closed/waited for, including refusal paths.
Initial28500 passes all11 in6.02s. Added exact protocol IDs and complete boundary ordering;
final12665 passes cold deploymentrelease/config/rehearsal and both release/rehearsal CLI packages
plus vet. Final92365 runs all11 under race detection in6.05s and passes. Format/diff checked.

Discrimination uses isolated Go compiler overlays in `/tmp/cloud-clicker-drain-probe.oE0JFE`,
copying committed `de0ca191`'s Docker runtime and forcing one caller evidence bit after derivation.
No production file is edited. 49767: the three existing R9 derivation/readiness/identity tests
pass with ReadinessDown forced true. 80742: the new unmarked-readiness case fails with a forged
all-true drain. Independent99824: CourtesyFrame forced true fails the missing-courtesy case,
also exposing all-true evidence. Production `git diff --exit-code` remains clean. The subsequent
correlation/order assertions strengthen the same fixture; the two bit oracles remain unchanged.

Scope: protocol/outer-method unit integration, not a real Account authentication, gameserver,
Docker stop, Postgres, Caddy or release artifact. Bootstrap/state/protocol and exit status are
synthetic. The default20s production bound is NOT shortened: two-second caller-context
cancellation makes deliberately withheld observations finite, and is not a measurement of
release latency or admitted-work/job completion. Those three flags remain process-contract
attestations; actual process and exact clean-host/R-006 evidence remain required. No new schema,
runtime behavior, timeout, retry, accepted policy, image or kernel version.

Review by: Codex (test implementer first filter). Recorded by: Codex. Exact test/docs/records
range starts after `de0ca191` through this containing commit; designated Claude review pending,
consolidatable with the RP-138/RP-546 observer corrections. Original R9 remains CHANGES REQUIRED;
no feature/archival/CI/full nine-tier1.0 or release promotion and no push/deploy. All named handles
terminal; continue remaining accepted implementation/reviews and the exact-artifact gates.

## 2026-10-13 — Complete original Claude R10 diff-review disposition

Review by: Codex (designated original-Claude reviewer). Recorded by: Codex. Exact range
`1f376ccc^..1f376ccc`, all four paths: runtime, runtime tests, canonical docs and owning log.
**Decision: CHANGES REQUIRED**, retaining RP-139; no approval of later corrections.

Inspected the host-path/type and header-equality addition, real age-envelope fixtures,
wrong-manifest/class/symlink and new header/payload negatives, recorded severing and the
documentation's irreversible/working-rollback claim. The checksum/equality improvement is
real; it cannot prove key-recipient match or logical Postgres restore (D-019 remains unruled).
The original positive preserves an invalid restore identity and synthetic command replies;
it is not a restored release. Current inherited creation also lacks invocation timestamp
binding, executed below as RP-548, not historical-tree execution. Existing wrong-image and
pre-destructive restore-input routes remain separately evidenced/reviewed; no prior verdict
is expanded or relabeled by this full original diff disposition.

## 2026-10-13 — RP-548 bind created-backup timestamp claims to invocation

Under DP5's immediately-before-upgrade backup and DP6 creation path, sample UTC before/after
the create command, refuse zero/backwards clocks, and require its reported start/completion
inside that interval. Header validity still enforces start≤completion and exact host agreement.
No arbitrary age tolerance, minimum duration, key mount, format/retention, restore behavior,
schema, kernel, CI or timestamp-model change. Older retained backups remain restorable.

Initial76469 on `dac25388` fails twelve bad reports and passes four controls. Added zero-finish
and same-instant cases; isolated old-source overlay50491 reruns the complete20: fourteen bad
reports accepted, six fresh/exact-edge/same-instant controls pass. Every case first creates,
checksums and authenticates/restores a real age envelope, so bad ciphertext cannot mask temporal
admission. Both pre-upgrade and recovery paths execute; zero-start refuses before any command.
Command results and invocation clock are controlled, not real Docker/Postgres or elapsed-time
measurements. The fixture's one-minute window is input data, not a new operating budget.

Corrected5155 passes initial scope plus existing host binding. Final59384 passes complete cold
deploymentrelease, deploymentbackup, deploymentrehearsal and both release/rehearsal CLI packages,
with vet. Final50902 passes all20 and existing host-binding under race. Root Linux-amd64 helper
build passes into ignored `.cache/bin/deployment-release-linux-amd64`; no Linux runtime claim.
Formatting/diff checked. Original header test now supplies its own invocation clock, deliberately
wide enough that its one-second mismatch remains temporally legal. Isolated host-read-bypass
overlay50314 fails that unchanged oracle: `reported header differing from the host envelope
accepted: <nil>`. It proves freshness did not mask the retained equality discrimination.
Overlays live in owned `/tmp/cloud-clicker-backup-window.bG1cX3`; no live-file probe mutation.

Limits: temporal report consistency, not trusted snapshot time, authenticated editable-header
protection, recipient match, SQL restorability or exact-host/RPO/RTO. Existing metadata/snapshot
author holds and D-019 remain; no new policy inferred from this safety check. Docker remains
unverified from its last no-server-reply probe; no daemon/container action, push or deployment.

Review by: Codex (new repair implementer first filter). Recorded by: Codex. Exact new source/
tests/docs/records span starts after `dac25388` through this containing commit; designated
Claude review pending, consolidatable with the earlier host-byte correction. Original R10,
full Deployment and full nine-tier1.0 remain open. All execution handles are terminal.

## 2026-10-13 — Complete original Claude R12 diff-review disposition

Review by: Codex (designated original-Claude reviewer). Recorded by: Codex. Exact range
`4a6c8f27^..4a6c8f27`, all five paths: CLI, execution-plan source/tests, canonical docs and owning
log. **Decision: CHANGES REQUIRED**, retaining RP-140. No descendant repair approval.

Inspected the dedicated fixture-rejected exit3, absolute basename/subcommand/population binding,
the negative-exit and unsafe-command controls, and the replacement fixture executor. Exit3
correctly distinguishes a prepared refusal from usage/setup failure. The basename check remains
insufficient to prove execution of shipped bytes: its own shell stand-in passes every canonical
row. Final candidate-tool hashing is not a binding to the plan executable. That existing author
contract hold is not resolved here. Current docs now state this limitation without adopting a
new identity policy. The inherited sequential-time gap is executed at current `88868d34` below,
not claimed as an execution of historical R12. Existing verdicts and evidence remain intact.

## 2026-10-13 — RP-549 enforce sequential rehearsal observation times

Under DP-F/AC10's completed non-monotonic-observation refusal, require each command's start
at or after the prior completion. Check before launching the next producer and independently
when base/final validation reads results in canonical plan order. Local intervals still require
completion strictly after start. Exact adjacent boundaries and forward gaps remain valid; no
minimum gap, tolerance, timeout, clock model, result schema or producer identity policy added.

Regression24975 on `88868d34` accepts all three invalid executor sequences: same-step overlap,
same-step reversal and next-step overlap. Both invalid same-step result populations pass the
base and final loaders after all population/step/base/proof hashes are rebound. Every local
interval stays valid and the step's min/max remain unchanged, so these are chronology failures,
not bad-hash refusals. Three executor and two loader controls pass on that old source.
Initial command selection failed shell parsing before any test; the simple ChecksSequential
selector is used for all actual runs. No probe edits production code.

Corrected87321 passes all ten cases, including a marker proving the invalid next producer never
launches, stopped clock sampling and no invalid/later result files. Four cold affected packages
pass in54006; vet passes. Focused race7489 passes all ten. Linux-amd64 rehearsal helper build23310
passes into ignored `.cache/bin/deployment-rehearsal-linux-amd64`; that is not Linux execution.
Full root `make test-deployment-rehearsal`1601 initially fails a sandbox localhost-listen denial
in the existing readiness HTTP test before record validation. Authorized exact rerun53187 passes
all four cold packages and both retained-build-record validators. These validate historical
records, not newly rebuilt or floor-compliant release artifacts. Format and diff checked.

Boundary: real local subprocess/file execution and actual evidence loaders with controlled clock,
stand-in producers and synthetic artifacts/seals; not Docker/Postgres/Caddy, admitted-work recovery,
clean-host install, RPO/RTO or R-006. Existing RP-140/DG8 executable-identity hold, unsupported
producer rows and full exact-artifact gates remain. No daemon/container action, push, deployment,
kernel/CI change or full-nine-tier1.0 promotion.

Review by: Codex (new repair implementer first filter). Recorded by: Codex. Exact new source/tests/
docs/records range starts after `88868d34` through this containing commit; designated Claude
review pending, consolidatable with the earlier DP-F validator corrections. Original R12 remains
CHANGES REQUIRED. All named processes are terminal; continue accepted construction and review,
without treating owner holds or pending review as a global implementation stop.

## 2026-10-13 — Complete original Claude R18 diff-review disposition

Review by: Codex (designated original-Claude reviewer). Recorded by: Codex. Exact range
`82a78b53^..82a78b53`, all five paths: probe dispatch, release-negative producers, unit tests,
canonical docs and original log. **Decision: CHANGES REQUIRED**, retaining RP-141. Full-path
inspection does not approve Codex's later witness supplements.

The controller's forward-compatible control reaches the guard runtime before its migration
negative; the backup/image producer accepts intact inputs first. Existing59258 cold TestR18
population executes both real producers and their invalid-bundle/non-forward controls at current
`5d70f673`; deploymentrelease has no matches in that selection, not a package-wide result.
The original wrong-epoch unit uses an injected validator, supplemented below as RP-550. Earlier
diagnostic bundles are still unavailable; no historical command replay or clean-host claim.

The original log's image-ID-comparison severing claim cannot be inferred from the missing-image
mechanism. At current source, isolated image-comparison omission65583 leaves both TestR18 probes
green: the absent-image runner error still rejects. Independent48645 fails the existing actual
Prepare wrong-ID oracle. This is different-property coverage, not a broken production comparison
or permission to discard either test. No historical record is rewritten.

## 2026-10-13 — RP-550 exercise all packaged mutations against the real release gate

Under DP-F/AC1/AC8, extend the existing validated-bundle fixture and RunProbe population, not a
new harness. All nine mutation rows execute against actual AssembleBundle/ValidateBundle inputs;
each has an intact subject and missing-manifest/unbound-client baselines. Require named rejection
only for the prepared subject, typed content error for bad baselines, preserved manifest/target
bytes, both retained bundle validations and empty temporary work. No production or limit change.

Initial97448 passes all27 cases. Compiler overlays in owned
`/tmp/cloud-clicker-rehearsal-gates.oQjse6` remove only content identity or image-ID comparison
from copied committed source, never edit production files.14720 leaves the old fake wrong-epoch
case green;94251 fails the new actual epoch subject with outcome0/nil. The final typed-error
assertion strengthens invalid-baseline checks without changing that oracle. Final55350 passes
six cold affected packages and vet (rehearsal/release/backup/releasepackage/both CLIs); focused
race76508 passes all27. Formatting/diff and unchanged production files checked. Native backup
tests with unavailable Postgres are not DB evidence; no Docker/CI/clean-host run.

These are actual local assembly/file/gate operations with synthetic ELF/image/client inputs,
not executable release artifacts or logical dump restores. The existing R18 envelope is real age
crypto but its dump and runtime/image runner remain synthetic. RP-140, admitted-work producer,
fresh supported bundle-pair provenance and R-006/whole nine-tier1.0 remain open. No push/deploy,
kernel mint, lifecycle promotion or owner choice inferred from the new passing checks.

Review by: Codex (new test implementer first filter). Recorded by: Codex. Exact test/docs/records
range starts after `5d70f673` through this containing commit; designated Claude review pending,
consolidatable with the prior RP-141/R15 supplements. All execution handles are terminal.

## 2026-10-13 — RP-551 bind backup service files to endpoint query overrides

During current restore-path inspection, found that pgx honors host/port/user/password/database
query overrides but the libpq service/pass writer uses only URI authority/path fields. Under DP6's
exact-target backup/restore contract, apply those overrides before the existing field checks and
write the effective tuple to both files. `database` aliases `dbname`; contradictory aliases
cannot name a stable target and refuse. Invalid port, empty/adjusted host and injected field
values refuse before files are written. No new credential location, SSL policy, backup format,
retention, dump/snapshot/key policy, schema or kernel change. Other URI/transport options are not
covered by this endpoint-field correction or its proof.

Initial46921 fails seven override cases on `a7501ac1`, with its ordinary control green. The final
old-source compiler overlay89724 (owned `/tmp/cloud-clicker-backup-endpoint.5dq3AC`) fails18 of20:
eight endpoint overrides and ten unsafe/ambiguous inputs; ordinary and password-isolation controls
pass. Corrected86036 passes the initial19; final race37421 passes all20 after the host-whitespace
case is added. Expectations come from pinned pgx5.10.0's actual parser plus literal service/pass
checks; no database is connected, and credentials are synthetic test data. No production probe
mutation. One patch-context mismatch made no edit and was corrected before tooling.

43023 passes six cold backup/release/rehearsal/CLI packages plus vet. Root Linux-amd64 backup-helper
build passes into ignored `.cache/bin/deployment-backup-linux-amd64`, not a runtime proof. Extend
the existing Postgres empty/populated test with swapped URI paths and explicit `dbname` overrides,
retaining all original restore/corruption/identity checks. A misplaced test-helper insertion was
caught by diff inspection and moved before compilation. Final60055 cold backup/vet pass; all seven
DB integration functions explicitly skip, including the new populations. Docker's last owned
probe has no server reply; no restart/cleanup or host-DB workaround. The declared
`make test-deployment-backup` execution remains mandatory before DB/AC4 acceptance.

Review by: Codex (repair implementer first filter). Recorded by: Codex. Exact source/tests/docs/
records range starts after `a7501ac1` through this containing commit; designated Claude review
pending. This is a verified parser/file correction, not an executed logical restore or full R20,
Deployment/R-006/1.0 closeout. Existing author/owner and exact-artifact gates remain; no push/deploy.
All execution handles are terminal. Continue complete original R20 inspection and accepted work
while retaining this named integration gate rather than substituting native green for Postgres.

## 2026-10-13 — Complete original Claude R20 diff-review disposition

Review by: Codex (designated original-Claude reviewer). Recorded by: Codex. Exact range
`680b7bb1^..680b7bb1`, all eleven paths: CLI dispatch/classification/tests, backup clean-target
class/DB assertion, rehearsal producer/tests/binding, runtime classification, docs and log.
**Decision: CHANGES REQUIRED**, retaining RP-142; no approval of corrected descendants.

The dedicated error class preserves query/setup failures as generic invalid input. The producer
requires the exact install, compares before/after recovery identity and removes its own backup;
its original class test uses a synthetic runtime and error rendering, not a composed Docker CLI.
Original substring classification remains the blocking flaw, already corrected in its separate
Codex range. The original relations-only cleanliness census also lacks later RP-423's nonrelational
coverage; those existing repairs are not approved here. Current record parsing has a distinct
member-admission gap, executed below as RP-552, not as a historical R20-tree run. Original DB/
severing records are retained, not reclassified as current clean-host proof.

## 2026-10-13 — RP-552 exact non-clean restore refusal members

Under DP-F's attributable prepared-refusal rule, inspect the raw record's decoded member names
before matching values. Require the existing five exact unique string fields; reject duplicate,
escaped-duplicate and case-aliased members, including a conflicting earlier class. Preserve valid
field order/whitespace/escaping and all existing typed-command, exit/cause, value, nonempty-time
and one-record conditions. No generic JSON framework, broader ledger parser, new timestamp rule,
backup/restore behavior, policy, schema, kernel or CI change.

Old2007 at `fc04db24` admits all16 bad records, while four formatting controls pass. Corrected
13780 passes that population, existing attribution negatives, the producer and CLI class control.
Extend the existing producer fixture with a Slog-rendered case alias and conflicting class;
isolated committed-source overlay3842 gives false ProbeRejected/nil for both. Overlay resides in
owned `/tmp/cloud-clicker-refusal-record.MVTS5i`; no production file is probed or restored.
Final50364 passes all six cold release/rehearsal/backup/CLI packages and vet, including both new
producer cases. Final2401 passes the complete focused NonClean population under race. Root
Linux-amd64 release-helper build passes into ignored `.cache/bin/deployment-release-linux-amd64`.
Formatting/diff checked; the regression cases discriminate without another source mutation.

Boundary: actual parser, typed failure records, physical fixture files and real probe orchestration
with synthetic backup/identity/runtime inputs; no Docker/SQL/clean-host execution. Native DB skips
and unavailable declared service remain, including RP-551's unexecuted extension. No current
composed CLI stderr, R-006, full Deployment or nine-tier1.0 promotion; no push/deploy or cleanup.

Review by: Codex (new repair implementer first filter). Recorded by: Codex. Exact new source/tests/
docs/records range starts after `fc04db24` through this containing commit; designated Claude review
pending, consolidatable with RP-142's earlier attribution correction. Original R20 is now fully
dispositioned, not approved. All execution handles are terminal; remaining author/owner, producer,
artifact and full-range review gates stay open while safe accepted work continues.

## 2026-10-13 — Complete original Claude R2 disposition; failed-install regression (RP-554)

Review by: Codex. Recorded by: Codex. Exact original range `92fab705^..92fab705`, all nine
changed paths: docs/log, release CLI, rehearsal install/recovery constructors, command builder,
DockerRuntime, its test and rotation ledger. Decision: **CHANGES REQUIRED**, extending rather
than replacing the earlier targeted RP-134 verdict. This is not approval of any later Codex code.

The constructor wiring and static paired-overlay classifier match the submitted bounded scope;
single-family composition remains DESIGN-GAP4, not a decision made by this review. RP-134 already
proved the original command-lock interleaving. Complete invocation tracing exposes a second
omission: `AbortInstall` validates `normalized()` but discards the returned value. Its cleanup
Compose command therefore loses the overlay R2 adds, while the production CLI's unset Runner
panics. The nil-runner error predates R2; R2's overlay change leaves the same mistake unresolved.
The current controller calls this method after failed install startup/identity/smoke and joins
returned cleanup errors; a panic bypasses that failure-record path.

New tests run against unchanged `b868cd2c`: corrected fixture run4869 fails three of eight
cases—both-open overlay omitted and both production-shaped default-runner calls panic—while
five ordinary/refusal controls pass. Initial58955 additionally failed my removed-pair fixture
because its append order went backwards; repaired the fixture chronology, not production or
assertions, before4869. No new rotation policy, predeclaration requirement or host action.

## 2026-10-13 — RP-554 preserve failed-install runtime defaults and rotation state

Under accepted DP4/DP5 and the existing cleanup contract, retain `normalized()`'s returned runtime
in `AbortInstall`. Same `down --volumes --remove-orphans` target and existing clean-install guard;
no other lifecycle, key timing, retention, schema, balance, kernel or CI-policy change. Tests
require exact base/overlay arguments, refusal before commands for single-open/malformed ledgers,
and propagation of the actual default ExecRunner's exit23 failure. That process runs an owned
recording executable as `docker` on a test-only PATH, never Docker or a daemon.

Finished root checks: focused4939 passes all eight cases;94188 passes them under race.
23067 cold `make test-go GO_PACKAGES='./deploymentrelease ./deploymentrehearsal
./cmd/deployment-release ./cmd/deployment-rehearsal' GO_TEST_FLAGS='-count=1'` passes all four
packages; matching vet exits0. The selected Integration run14155 explicitly skips the Caddy/
Postgres release witness (environment absent); rehearsal reports no tests for that selector,
not integration success. Linux-amd64 release-helper build succeeds at
`/tmp/cloud-clicker-abort-install.g2Zh3G/deployment-release`; not a released bundle/runtime proof.
Formatting and diff checked. Old-behavior regressions discriminate without another mutation.

Review by: Codex (repair implementer first filter). Recorded by: Codex. Exact new correction/
tests/docs/records range starts after `b868cd2c` through this containing commit; Claude designated
review remains pending and may consolidate with the prior RP-134/lifecycle corrections. Original
R2 is fully dispositioned, not approved. Actual failed-install Docker cleanup, clean-host R-006,
author decisions, full range review and nine-tier1.0 remain open. All handles terminal; no push,
deployment or operator cleanup performed. Continue accepted construction and remaining reviews.

## 2026-10-13 — Complete original Claude R1 disposition

Review by: Codex. Recorded by: Codex. Exact original range `1978660^..1978660`, all six changed
paths: Deployment docs, controller/interface/tests, Docker runtime/test and failure-stage ledger.
Decision: **CHANGES REQUIRED (RP-133)**, completing rather than replacing the earlier targeted
review. The image-load-before-in-image-preflight sequence, rollback input ordering and failure
stage records are coherent; tests require preflight/input failures to precede stop/reset. But the
original host preflight's length/checksum/file-permission checks do not establish decryption, as
the earlier executed wrong-but-valid identity regression proved. Its docs/interface overstate
that assurance. Current code uses the later Codex authenticated extraction correction;79587 cold
focused check passes its retained wrong-key/package/hash/refusal cases. That is current first-filter
evidence, not designated approval of Codex's changes or a real logical Postgres restore/RTO.
The exact RP-133 and later package/preflight corrective ranges remain owed to Claude; no archival.

## 2026-10-13 — RP-555 rotation commands cannot poison ledger chronology

Accepted DP4's append-only governed ledger already rejects backwards chronology on read. Both
mutation commands could nevertheless append a backdated row and return success. A legitimate
clock rollback across key families then leaves all later readers/commands unable to use the file.
Require `now` not precede the latest global row while the existing shared/ledger locks are held;
reject before append. Equal timestamps, family-specific overlaps and continuous never-reused IDs
stay unchanged. No history rewrite, new clock policy, cursor alias solution or gameplay attendance
decision. This is writer/reader conformance, not another owner ruling.

Old `3964194b` cold13329 fails six of18 actual file/command cases: earlier activation/removal
for each JWT/bootstrap/cursor returns nil, changes bytes and makes ReadRotationLedger fail.
The12 equal/later controls pass. Each earlier removal still exceeds its own governed minimum,
so the finding is specifically global chronology. Corrected77250 passes all18 and confirms a
refused command leaves identical/readable bytes and can be retried successfully at a later time.
25688 passes the population under race.18960 cold four-package deploymentrelease/rehearsal/
both-command tests pass; matching vet exits0. Linux-amd64 release-helper build succeeds at
`/tmp/cloud-clicker-abort-install.g2Zh3G/deployment-release-chronology`. The initial output-less
build invocation refused at its Make guard before compiling; supplied the required output path.
Formatting/diff checked; ordinary old-behavior regressions discriminate without another mutation.

CI routing inspected: `.github/workflows/ci.yml` runs `make verify-server-core`, whose
`test-go-core` selects every Go package except harness with cold count1. This package/test is
included without a workflow change; that source check is not an executed hosted/Linux CI verdict.

Physical supported-command/ledger proof, not Docker/Postgres, deployed credential/cursor rotation,
release-sized recovery or exact-artifact R-006. No real-service integration executed; existing
environment limits, full nine-tier floor, owner/author holds and independent review remain open.
Review by: Codex (repair implementer first filter). Recorded by: Codex. Exact new code/tests/docs/
records range starts after `3964194b` through this containing commit; Claude designated review
pending, consolidatable with prior rotation/lifecycle fixes. All handles terminal. No acceptance,
archival, push, deployment or operator cleanup; continue safe accepted construction/reviews.

## 2026-10-13 — Operator-state admission before lifecycle effects (RP-556/RP-557)

Under accepted DP4/DP5 and existing private/append-only ledger rules, admit existing release
history before runtime operations. Install/rollback already read it, but only final append
enforced its type/mode and new start chronology; release did not read it until after its runtime
sequence. Shared-lock-held admission now requires a regular0600 file, valid ordered history and
start strictly after the latest completion. Preserve missing-history behavior and rollback's
exact authority/retry chain. Refusals cannot honestly append a row to inadmissible history.
No new ledger schema, cutoff, retrospective adoption or failed-transition policy.

Old `b448069b`52857 fails15 of24 cases: existing malformed/empty/directory release history,
symlink/public-mode history and earlier/equal starts reach runtime before final append refuses.
The rollback arms record stop/reset/restore and the install arms also cleanup; this is why merely
checking the final error would be vacuous. Nine controls pass. Corrected11880 passes all24,
requiring zero runtime calls and unchanged history on refusal plus exact success records on
the three valid sequences. Runtime is the existing recording fixture, not a real database/host.

The same operator-state boundary exposed chmod-before-final-directory-validation. Two old
lock/append cases in58411 refuse but change owned symlink-target permissions0775→0700; four
private-target/real-directory controls pass. Check final directory type before chmod in both
paths. Corrected20001 passes all six: link/target remain untouched on refusal and a real directory
still initializes privately.27597 originally also closed the error result's typed-nil closer;
fixed that test bookkeeping before58411, without changing the observed permission regression.
No new ancestor-link/race sandbox, user-data deletion or external permission change is claimed.

Finished checks after removing the temporary authority diagnostic below:70141 cold four-package
deploymentrelease/rehearsal/both-command tests pass; matching vet exits0.89731 passes all30
cases under race. Linux-amd64 helper build succeeds at
`/tmp/cloud-clicker-abort-install.g2Zh3G/deployment-release-admission`. Formatting/diff checked.
The tests live in the existing server CI population; no workflow change or hosted CI verdict.
No actual Docker/Postgres lifecycle, R-006, future-write durability or mid-operation-clock proof.

Review by: Codex (implementer first filter). Recorded by: Codex. Entire new source/tests/docs/
records range starts after `b448069b` through this containing commit; designated Claude review
pending, consolidatable with preceding operator/lifecycle fixes. All handles terminal. No
acceptance, archival, push, deployment, daemon restart or operator cleanup.

## 2026-10-13 — Current release identity gap, not another generic owner decision (RP-558)

Source audit finds no current-bundle/installed-release binding in Controller.Release. Compare
checks caller bundles; the new history admission above checks admissibility, not their identities.
Temporary two-case `TestCurrentReleaseAuthorityDiagnostic`58696 uses existing fixtureBundles,
validRuntime and physical Append/ReadReleaseLedger. Seed a successful install at12:00–12:01,
then release at12:02 from supplied1.0.0 to1.1.0. Matching recorded1.0.0 and contradictory2.0.0
(different manifest SHA256) both execute `[prepare preflight backup drain start identity smoke]`
and append succeeded1.1.0. The runtime/loader are doubles, so this does not prove a real-host
downgrade bypasses image/database gates. It proves the controller cannot establish the claimed
live-release ordering from its current inputs. Removed the temporary probe exactly before the
final package/race checks; do not retain an unsafe outcome as a normal acceptance assertion.

**DESIGN-GAP10, engineering/author closure:** specify the authoritative current-release tuple
and actual running-artifact binding; define absent history and failed/partially executed
transition handling. Then adopt/review that contract and implement a stale-identity refusal
through the real helper/host before backup/drain. Do not merely compare two caller-selected
versions, infer the current host from a supplied path, or silently invent a new retry/adoption
policy. This is a technical contract to prepare, not a newly delegated product choice. Canonical
docs now limit the implementation claim; the accepted forward-only floor stays unchanged.
Next: prepare that bounded authority contract and continue accepted work; no full1.0 promotion.

## 2026-10-13 — Current-release proposal and actual child-lifetime precursor (RP-558)

Prepared `rfc/deployment-current-release-authority.md` under the existing RFC template, linked
from the active index. It proposes exact recorded/live image and mounted-byte binding, removal
of history-less ordinary release, explicit v2 release/rollback start/checkpoint/finish records,
and governed manual interruption recovery. The existing initial-install v1 path and completed-
release seven-day window remain. These are explicit proposed operational changes, not authority
for an implementation or a Marco ruling. Parent acceptance/body reconciliation and cross-party
normative review remain required. It is **not acceptance-ready**: pinned-engine mounted-reader
and prior-request quiescence questions must be executed and reconciled, not left to an implementer
to invent. Existing D-019, protection metadata, DESIGN-GAPs1–9, code review and full1.0 holds stay.

Source checked against the accepted Deployment body, current seven-service Compose templates,
six-image manifest/validator, actual Prepare/image-ID behavior, controller/ledger and lifecycle
readers. The backup service shares Postgres's image; nine mounted roots need original byte/set
checks. Docker's primary copy/inspection/label/export docs provide partial instrument support,
not runtime proof. The old721-line `cd3235a7` proposal is confirmed unmerged/non-ancestral and its
existing CHANGES REQUIRED verdict remains; this new scoped draft does not import its options.
No claim that all R-006 populations have producers. My first body-refinement patch refused a
stale context atomically; corrected it before source/diff checks, not an executed product failure.

Mechanical precursor: current ExecRunner has no parent-death supervision. Predeclared native
population is intact parent versus killed owned helper while its real child waits on an explicit
gate, with the real operator lock held.74635 completes both handshakes on darwin/arm64. Refined
72139 also checks the actual owned child PID becomes terminal: before kill the parent holds the
lock; after kill the lock is reacquirable while the child has not completed; opening the gate then
makes that child write its completion marker. Normal completion is the control. Both owned child
processes are observed exited. This disproves lock-only quiescence; it proves neither a Linux
supervision solution nor that a Docker daemon request stops when its client dies. The fixture's
five-second handshake/termination guards invalidate an incomplete run, not an operating budget.

Preserved the verbatim executable diagnostic outside normal CI as
`probes/child-lifetime-probe.go.txt`. Reproduction: use apply_patch to temporarily move it to
`server/deploymentrelease/child_lifetime_probe_test.go`, run root
`make test-go GO_PACKAGES='./deploymentrelease'
GO_TEST_FLAGS='-count=1 -run TestExecChildLifetimeDiagnostic -v'`, await its terminal result,
then move it back. Do not turn the observed unsafe survival into a normal correctness assertion;
the file is mechanical evidence for this draft, not feature acceptance. Its source is gofmt-clean.

Fresh direct Docker socket ping19065 terminates exit28 after5006ms with zero bytes. No daemon
restart/cleanup or hanging test launched. Requested explicit restart permission asynchronously;
no approval is inferred from its suggested options. The supported Linux/container experiments
remain unexecuted. No product/schema/balance/CI-workflow or authored player-copy bytes changed.
Documents/index/paths/diff checked; no unrelated software suite for this normative draft/records.

Review by: Codex (draft/research author first filter). Recorded by: Codex. Exact new proposal/
source/evidence/record range starts after `b29fc293` through this containing commit; independent
Claude review is pending, not self-approval. All observed sessions/owned processes terminal.
Next: finish those mechanical questions when the declared engine is usable, reconcile CR2/CR4,
then obtain adoption/normative review; other accepted work remains available. No release-ready,
archival, publication, push, deployment or whole nine-tier1.0 claim.

## 2026-10-13 — Reject malformed backup URL queries before credential preparation (RP-559)

At `0e7a4b95`, the previous turn completed the Headcount oracle investigation. Continue
accepted DP3/DP6 conformance, without adopting the held deployment/gameplay drafts.
`databaseServiceEnvironment` now uses complete `url.ParseQuery` and refuses any error
before either credential file is written. Previously `URL.Query` silently dropped malformed
fields and returned successful service/pass files for the fallback settings. Existing valid
overrides, encoded values, password isolation, SSL handling and supported options stay unchanged.
[Primary Go API documentation](https://pkg.go.dev/net/url#URL.Query) names the silent-discard
behavior; no new URI compatibility or transport policy is proposed.

Initial diagnostic63550 incorrectly required the actual pgx parser to reject those URLs:
all12 preconditions fail, with4 encoded-value controls passing. That falsifies the proposed
cross-parser mismatch and is not old-helper regression evidence. Correct the diagnostic's
population check to complete query syntax, retaining the actual helper/file assertions.
Old-source44999 then fails all12 cases: each returns nil error and writes both credential
files;4 controls pass. Corrected focused race85803 passes the new16 and all existing credential
file cases,1.323s. No source mutation, fixture tolerance or parser-error text/secret logging.

Root `make test-go GO_PACKAGES='./deploymentbackup ./deploymentrelease ./deploymentrehearsal
./releasepackage' GO_TEST_FLAGS='-count=1'`89373 passes0.929/9.216/7.380/1.297s; same-package
vet exits0. Native integration-only33185 explicitly skips all7 database populations (URLs
unset), not SQL evidence. Docker-dependent restore remains unexecuted; no host DB substitute,
daemon restart/cleanup, Linux/hosted CI or R-006 pass. Go formatting and final diff checked.

Review by: Codex (repair implementer first filter). Recorded by: Codex. Full source/tests/docs/
records range after `0e7a4b95` through this containing commit needs designated Claude review,
consolidatable with RP-551 without erasing its unexecuted restore obligation. No accepted body,
balance/kernel, migration, CI workflow, feature checkbox or release claim changes. Next: declared
restore on these bytes when Docker is usable; continue independent accepted work meanwhile.
Nothing pushed, published or deployed; the full nine-tier1.0 goal remains active.

## 2026-10-14 — DP6 binds stale cleanup to inspected identities; Linux/Postgres is available

Outcome: repair cleanup's existing inspected-target boundary at `68eee56b`.
Three old physical-file cases delete uninspected data: renamed/replaced directory,
retargeted directory alias and replaced file with identical name/age. Capture
directory/file identities through os.Root; require the exclusive directory lock
and reopened root to match inspection. Recheck each stale inode/type/name/age
and remove through the bound root, not the mutable pathname. Snapshot JSON,
one-hour stale threshold, locks, retention and encryption formats are unchanged.
This is not a general arbitrary-privileged-file-mutation sandbox.

Regression and coverage: old TestCleanupRejectsReplacement population fails all3
with actual removal of test-owned replacement bytes. Corrected native TestCleanup
passes, including unchanged real/alias directories, deterministic bound-root
rename and existing actual second-process live-writer exclusion/orphan cleanup.
The bound-root test exercises the production destructive helper after binding,
not a timed race or the entry point's lock (the separate lock witness covers it).
Focused race passes2.417s. Cold deploymentbackup/deploymentrelease/
deploymentrehearsal/cmd/deployment-backup and affected vet pass; Linux-amd64 helper
build passes with RELEASE_BACKUP_OUTPUT=.cache/bin/deployment-backup. The first
build call omitted that required output and refused before compiling; no product
failure inferred. The compiled helper is development output, not a release bundle.

Environment changed: read-only Docker ping returns OK/HTTP200 in0.003226s; no
agent restart/daemon cleanup was performed. Inspect the declared backup-test
project as empty before execution. `make test-deployment-backup`77791 exits0
on actual Linux-amd64/Postgres16: all7 integration functions run, including the
empty/populated ordinary and database-query-override restores (RP-551), clean-
target refusals, same-count mutation and snapshot consistency. Existing pre-
deletion restore still reproduces retained/deleted data behavior, not a rights
policy fix. Trailing authenticated data refuses before DB mutation; positioned
real pg_dump packages restore with exact semantic recovery identity.

Then run the compiled test binary in the same declared service with TestCleanup,
count1 and verbose:55144 exits0. New replacement/anchored-root/alias populations
and real second-process ownership/orphan cases execute on Linux, not merely
cross-compile. Both test rounds' owned containers/network/tmpfs are removed;
final project ps is empty. No production database, host DB substitute, external
deployment, off-host durability/RPO/RTO, Caddy/clean-host or full R-006 proof.

Fresh gh metadata also identifies failed published CI38560135778 at be66902b;
failure markers confirm RP-131/RP-307. The CI log owns this separate read-only
checkpoint; local backup success is not whole-CI success. Update current routing
to local Docker verification READY, keeping exact-artifact/author/review/target
gates intact. Historical unavailable executions remain historical, not passes.

Handoff: source/test/docs plus RP-611/RP-551 and material environment/status
reconciliation. Review by: Codex (implementer first filter); Recorded by: Codex.
Full range after68eee56b through this containing commit needs designated Claude
review; earlier exact corrective ranges remain owed. No accepted body, owner
choice, feature checkbox, archival, mint, publication or push. Next: execute the
other declared integration lanes individually and the pinned-engine engineering
experiments under their actual prerequisites; daemon recovery alone closes none.

## 2026-10-14 — DP5 real Caddy release lane and complete smoke-response reads

Start clean7644aed4. Read the entire current-release proposal and inspect the
actual daemon: Linux/aarch64, Engine28.4.0, Compose2.39.4-desktop.1. The versions
match the recorded development pair, but this is not the supported native
Linux/amd64 host. Do not substitute emulation for CR2/CR4's required population
or call that draft acceptance-ready. No mounted-byte/quiescence experiment or
new supervisor/recovery mechanism is implemented/adopted.

Accepted DP5 lane: inspect the declared release-test project as empty, then
run make test-deployment-release.82532 exits0, actual test2.29s: Caddy HTTPS
with its integration-only internal CA/test client, Postgres16, authenticated
bootstrap/state/socket; readiness withdrawal with the
server marker, exact new-intent refusal, completion of an admitted DB-blocked
intent, courtesy+4003 closure, restarted composition and restored-account smoke.
Real encrypted-package preflight accepts the intact dump and rejects authenticated
trailing data before recording runtime commands. Changed catalog bytes refuse.
The test runs the Go composition in its compiled test process with a read-only
checkout mount. Its restore starts the same current composition, not a different
previous release image; the preflight command runner is a double. This is not
controller/CLI identity binding, exact-bundle/previous-manifest or clean-host R-006.

RP-612: the smoke decoder's io.LimitReader creates false EOF at4MiB. Three cold
old-reader negatives accept oversized whitespace, hidden garbage and hidden second
JSON; a real HTTP caller also sends authenticated state after an invalid bootstrap
body. Valid short/at-boundary and deliberately stopped healthy HTTP controls pass.
Read one sentinel byte and refuse overflow without increasing the admitted ceiling.
Do not expose reply bytes. Add an endless-tail control proving the actual reader
consumes only the existing ceiling plus detection byte; no timing-budget claim.

Corrected focused, cold deploymentrelease/deploymentrehearsal (9.478/7.931s),
affected vet and focused race1.665s pass. Atomic counters keep HTTP fixture
observations race-free. Final make test-deployment-release67398 exits0, actual
test2.53s, on corrected helper bytes with all original journey assertions intact.
Both runs remove only their owned test containers/network/tmpfs; final project
ps is empty. No public deployment, operator secret access, full API DTO validation,
native-amd64 host, journal/retention/RPO/RTO, full AC3/AC5 or release approval.

Handoff: helper, regressions, canonical docs, RP-612 and this plan/log; current
verification gets the new real-lane result, not a gameplay milestone. Review by:
Codex (implementer first filter); Recorded by: Codex. Whole coherent range after
7644aed4 through this containing commit needs Claude designated review; earlier
corrective spans remain owed. No archived body, owner choice, feature checkbox,
mint, publication or push. Native-host CR2/CR4 and the full nine-tier/platform floor
remain open; continue the other declared integration populations independently.

## 2026-10-14 — RP-615 bounded minute-timer callback observation

Resume clean `21734b10` under DP7/AC7 and RFC-0000. The manually selected
real-minute test has three bare channel receives that ignore its existing
95-second context and treat a closed channel's zero value as healthy. Factor
those exact old receive semantics for regression:85538 fails two expired
context cases;9709 also fails a closed-channel case. Timely success, SQL error
and child cancellation controls pass unchanged. No production bug is inferred.

Correct the test-only helper: check expiry before receive, select context
cancellation while waiting, reject closed/missing callbacks and check expiry
again before returning an outcome. All three callers use it without changing
actual job/timer, callback forwarding, SQL/readiness/metric assertions, parent
95-second budgets or cadence. Final native cold/vet65272 and race37430 pass
(0.348/1.438s). Formatting/diff checked; final diagnostic wording is editorial.
Cold helper/vet37919 also passes on the final formatted source after that wording edit.

Actual `make test-deployment-cleanup-runtime`52076 exits0 on declared
Linux-amd64/Postgres16: runtime SQL failure60.86s and orderly drain60.52s,
whole121.448s. Both actual60-second ticks execute, not skipped or accelerated.
The trigger failure is supervised, readiness withdraws, credentials remain
transactionally intact and fresh composition recovers after fault removal.
Drain cancels the actual session-table lock wait with unchanged credentials,
canceled=1/failure=0 and no fatal supervision/log alarm. This refreshes the
named current-source manual population skipped by the ordinary Go wave.

Scope is the compiled test composition/private registry and SQL, not OCI
process, receiver delivery, host journald/retention or clean-host R-006/fullAC7.
Test-owned tmpfs Postgres was absent before setup; remove only its service
after terminal checks, preserving volumes/caches/network/unrelated services.
Only that service is removed; final inventory is empty and its disposable test
data is discarded. No durable or user data is removed.
No kernel bump, production byte, new runtime/retention/clock policy, owner
adoption, archival, publish/push or whole1.0 claim. Review by: Codex
(implementer first filter); Recorded by: Codex. Whole test/docs/plan/ledger/log
range after `21734b10` through this containing commit needs designated Claude
review, consolidatable with the previous DP7 corrections. Earlier debt remains.

## 2026-10-15 — RP-125 HTTP credential revival after full-dump restore

Research under the existing rights inventory plan, not a retention/deletion ruling.
The new backup integration test creates subject/bystander and credential pairs via
actual Account HTTP, takes encrypted B_pre, HTTP-deletes the subject, proves account
absence/credential denial, takes B_post and restores each into a fresh test target.
Before token rotation/login can mutate the target, five selected credential SQL
families must match the actual chosen pre/post snapshot. Healthy bystander and
wrong-code/invalid-token refusals prevent dead-server/always-authorize observations.
Every revived credential must read the original Founder. Payloads/keys/IDs/tokens
stay private; the existing dossier holds the counts/status table and all limits.

Final root `make test-deployment-backup`95335 passes all eight integration functions
cold, no skips, using its declared Linux-amd64/Postgres16 image/toolchain and real
pg_dump/age/pg_restore. New function2.50s; all seven existing tests remain intact.
Initial36645 passed before adding credential-row binding. Source after DELETE:
old access/refresh/recovery401; B_pre restores all three to200 for the original
Founder; B_post retains401. Bystander200 in all states. Final native package/vet
78814 pass; native DB skips are not integration credit. Diff/whitespace checked.

The fixture uses the existing constructor-only catalog, fixture epoch/manifest,
unchanged synthetic keys and frozen valid-token clock. It is not the gameserver's
foundation/Guild initializer, gameplay, bootstrap retry, expiry/rotation, Caddy,
off-host storage, RPO/RTO or exact-release clean-host proof. No production, policy,
schema, migration, kernel, CI or content bytes change. RP-125 stays open for
D-009/D-015 and an accepted reconciliation contract; revival is observed, not
required as future policy or described as a public authentication bypass.

All handles terminate; the initially empty backup-test project automatically
removes its owned tmpfs databases and network. No durable/user data or Go caches
removed. Review by: Codex (test/research author first filter). Recorded by: Codex.
Whole seven-path test/plan/dossier/inventory/decision-evidence/ledger/log range
after `42ce00f0` through this containing commit needs designated Claude review.
No feature acceptance, archival, publication, deployment or release promotion.

## 2026-10-15 — Reconcile stale backup-override verification text

Fresh independent corrective reviewer `/root/backup_safety_corrective_review`
finds docs/deployment.md still says RP-551's real-Postgres override restores have
not executed. The DP6 checkpoint at7644aed4 records their actual completion;
the reviewer also independently executes both ordinary/override populations.
Replace only that stale claim, distinguishing parser/file and SQL evidence and
retaining general-URI/exact-artifact/clean-host limits. No software behavior,
accepted contract, test, policy or release gate changes. Parent checks the diff
against the owning recorded execution; no unrelated suite for this editorial repair.
Review by: Codex parent (editorial first filter). Recorded by: Codex parent.
Include this two-path supplement afterf2751639 in the pending backup corrective
review, not as approval of the later data-rights experiment or full DP6 acceptance.

## 2026-10-15 — Backup safety independent corrective review: APPROVED

Review by: fresh independent Codex sub-agent `/root/backup_safety_corrective_review`.
Recorded by: Codex parent, transcribing its verdict under Marco's delegation.
Complete reviewed disjoint ranges (base exclusive, tip inclusive):

- RP-551: `a7501ac1bac149d1fb87949b75b308ef58073479..fc04db242303e3d873b213d2b5a2e597ca809db3` (seven paths).
- RP-559: `0e7a4b95f1daeacac175678883638c1aae339f51..4966e29671783bfec9a8ee5a1ae43612ff6bdc72` (six paths).
- RP-611/environment records: `68eee56b5e37e9454359813c822ea1215c086ef3..7644aed4001c8f17d59b397d1d5eadd25c7d2b59` (eleven paths).
- Stale-doc repair: `f2751639872f09cf8a32b36bc6ae7041d1a3f2f9..22dd54e97c641c11ab6913371eda0cb8cb2fdd55` (two paths).

Verdict: **APPROVED**, all changed code/tests/docs/records inspected, no remaining
findings after the stale canonical verification claim is repaired. Historical-source
compiler overlays independently reproduce18 endpoint/unsafe failures,12 malformed-query
failures and3 destructive replacement-cleanup failures, retaining ordinary/password
controls. Corrected credential/cleanup race tests pass, including anchored roots,
aliases, encoded values, second-process ownership and orphan cleanup. Five affected
packages pass cold and vet; an initial release-package sandbox listener denial is
resolved by the permitted rerun, not counted as product success/failure.

Actual root `make test-deployment-backup` passes Linux-amd64/Postgres16: all seven
original integration functions execute, including empty/populated ordinary/override
restores and refusals. Linux TestCleanup also passes. The current eighth rights-HTTP
experiment executes but is explicitly outside approval. Original relevant source/tests
remain byte-identical at22dd54e9. Range whitespace checks pass. Independent read-only
CI38560135778 metadata/finished failure markers confirm that historical reconciliation;
this is not a new hosted run. Reviewer makes no tracked edits; only review-created
test resources removed and final backup project inventory empty.

No intervening-commit approval, general URI compatibility, rights-policy resolution,
draft adoption, off-host durability, RPO/RTO, full DP6/Deployment, clean-host/R-006 or
release acceptance. Original failures and other corrective review debt remain.

## 2026-10-15 — Operator-state independent corrective review: APPROVED

Review by: fresh independent Codex sub-agent `/root/operator_state_corrective_review`.
Recorded by: Codex parent, transcribing its verdict under Marco's delegation.
Complete reviewed ranges (base exclusive, tip inclusive):

- RP-554: `b868cd2cba70a34162ad608d03a3a4f956393917..3964194b4c4ece27be07267bbaf6b0a9b3980507`.
- RP-555: `3964194b4c4ece27be07267bbaf6b0a9b3980507..b448069beef2b25fca0acc95f61782d00dcfa0ec`.
- RP-556/RP-557: `b448069beef2b25fca0acc95f61782d00dcfa0ec..b29fc293a0913d5c6365c10b997556ea09889510`.

Verdict: **APPROVED**, no actionable findings; all source/tests/docs/plan/ledger/log
paths inspected. Four affected deploymentrelease/rehearsal/both-command packages
pass cold count1 and matching vet. Joined56 corrective cases pass race/count1.
Independent old-source overlays reproduce RP-554: 3/8 failures; RP-555: 6/18;
RP-556: 15/24; RP-557: 2/6, with remaining controls passing. Initial chronology
overlay mixes old ledger/current controller and fails compilation; consistent old
sources then reproduce the six actual failures. Compile failure is not regression
evidence. Formatting/range diff clean; reviewed source/tests unchanged atf2751639.

Reviewer makes no tracked edits or Docker/DB resources; owned overlay evidence
remains `/tmp/cloud-clicker-operator-review.TKXOiZ`, not claimed cleaned. Physical
ledgers and recording subprocesses are not real failed-install cleanup, deployed
rotation, full lifecycle, clean-host/R-006 or release proof. RP-558/current-release
authority, future clock/durability and original Claude CHANGES REQUIRED dispositions
stay open. Missing-history semantics unchanged; other descendants not approved.

## 2026-10-15 — Independent RP-612/RP-615 corrective approvals with actual services

Review by: fresh independent Codex sub-agent `/root/deployment_evidence_corrective_review`,
under Marco's 2026-10-15 corrective-code delegation. Recorded by: Codex parent transcribing
the returned verdict, not implementer self-approval. **APPROVED**, no findings, separately:

- RP-612, all seven paths: `7644aed4001c8f17d59b397d1d5eadd25c7d2b59..cc256422c6bc4579ac588676b7e3192db7075938`.
- RP-615, all five paths: `21734b100b56ef010a1c5c0866478369c05318bc..ea194d8dec3a353c9daefd58c683229cd8e56633`.

Full code/test/docs/records inspected against DP3/DP5/DP7/AC7. Sentinel-byte overflow
detection retains the4MiB ceiling; callback observation refuses expired/closed results while
retaining timely SQL-error and child-cancellation observations. No production timer/budget change.

Independently executed at descendant `801d2fee6b11837a9c26f28e25045b5d81d6dc44`:
`make test-go GO_PACKAGES='./deploymentrelease ./deploymentrehearsal' GO_TEST_FLAGS='-count=1'`
passes9.458s/9.146s; cold Gameserver passes0.280s with DB/manual skips explicit. Named
`TestCleanupRuntimeOutcome` executes all three helper tests; an initial wrong selector matched
no tests and supplies no evidence. Focused race `TestSmokeResponse|TestCleanupRuntimeOutcome`
passes1.661s/1.388s across Deploymentrelease/Gameserver, including overflow, hidden tail,
malformed/endless-tail and authenticated-followup refusal. Affected three-package vet passes.
`make test-deployment-release` passes actual Caddy/Postgres/WebSocket in2.94s: drain/refusal,
admitted work, restart/restore and damaged inputs. `make test-deployment-cleanup-runtime`
passes121.724s, actual SQL-failure60.97s and orderly-drain60.67s; scheduled SQL reports
`context_ended=true cancellation_error=true`. Real60-second callbacks and95-second arm budgets
unchanged. Both exact-range whitespace checks pass; reviewer makes no tracked edits.

These are descendant runs, not historical builds/fresh old-source reproductions. RP-612
source/tests are byte-identical to its commit; RP-615 helper/regressions are unchanged, while
later691e9bac adds route-expiry observations/job-count changes, not approved by these verdicts.
Old failures were inspected in retained logs. Docker28.4.0/Compose2.39.4-desktop.1 daemon is
Linux/arm64; declared amd64 services use emulation. No native supported-host, governed
previous-image rollback, OCI gameserver lifecycle, receiver delivery, journald/retention,
exact-release R-006, full AC7, aggregate CI or release/archival approval follows. Known27
Production failures and historical kernel guard50a3a514 remain unresolved, not rerun here.

Declared release target cleans its resources. Reviewer removes only its twice-inspected,
newly created tmpfs Postgres container, discarding disposable test data; shared networks/cache
volumes and two pre-existing BuildKit containers remain. No publication/deployment/push.
Continue remaining accepted implementation and current-release contract work; these two
corrective ranges no longer wait on Claude, without closing the original feature.

## 2026-10-15 — Exact recovery objectives independently approved

Review by: fresh independent Codex `/root/deployment_recovery_objectives_corrective_review`.
Recorded by: Codex parent under Marco's corrective-code delegation. **APPROVED**, all eight
paths in `ad54b2587f2fcd9fbc19ea3d5fd851710562f1b1..09cf45e5539437b581b9318443632a23a995977e`.
No findings. RP-503 applies full-duration bounds before seconds conversion; RP-504 includes
stop/reset and refuses a clock preceding the incident before destructive calls. Actual Docker
callers/final validator agree; existing6h RPO/4h RTO, clock/report semantics remain unchanged.

At descendant69c84ad4/product77fe2275/kernel221: affected five-package cold tests/vet76325,
backup/rehearsal cold race26990 and declared Linux-amd64 full packages8432 pass. Explicit
Linux72962 executes all12 duration-boundary and four recovery-phase regressions, no skips.
Generic Linux service lacks restore/admin URLs, so its seven restore functions SKIP and are
not restore evidence. Separate actual make test-deployment-backup63275 passes ALL8 current
integration parents with no skips: encrypted empty/populated/query-override restores, clean-
target refusals, identity mutations and snapshot checks. The eighth restored-HTTP-credentials
observation executes as context, NOT rights-policy approval or an expansion of this verdict.
First verbose output partly truncated; explicit regression/prerequisite check supplies totals.

Recovery source/tests byte-identical endpoint; later backup framing/credential/fixture changes
inspected as execution context only. Historical old failures inspected, not independently
reproduced. No reviewer edits/commits. All handles terminal; both Compose inventories empty.
Only newly created inspected tmpfs test Postgres removed; backup target removes its
services/new network. Disposable SQL data lost; persistent volumes/caches/builders preserved.
No actual Docker stop/reset timing, host incident, clean-host R-006/R-007, normative adoption,
full Deployment/hosted-CI/archival/release approval. Remaining authored/artifact/review gates stay.

## 2026-10-15 — Bundle-gate and refusal evidence corrections independently approved

Review by: fresh independent Codex `/root/deployment_refusal_gate_corrective_review`.
Recorded by: Codex parent under Marco's corrective-code delegation. **APPROVED**, no findings,
every changed path inspected, separately for complete scopes:

- RP-550, five paths: `5d70f673388a981a75d4dd2bb74e6096635fac6a..a7501ac1bac149d1fb87949b75b308ef58073479`.
- RP-552, seven paths: `fc04db242303e3d873b213d2b5a2e597ca809db3..a713137e59cbdb8a70d9cda75552e1cea0c5aa16`.

At descendant69c84ad4/product77fe2275/kernel221: six affected packages and vet97540 pass
cold; additional backup CLI5652 passes nine parents including the dedicated refusal-class
control. Focused cold race76714 passes27 real bundle-gate subjects/baselines, malformed-
record producer cases, classifier/format controls and actual ExecRunner provenance.
Formatting/both full-range whitespace checks pass. Nine intact mutation subjects reach actual
assembler/validator;18 invalid baselines must be typed content errors, not successful negative
evidence. Retained candidate/previous bundles remain valid and temporary work empties.
Exact decoded-member validation rejects duplicate/escaped-duplicate and case aliases before
values; four valid formatting controls and outer Docker/exit1/cause/value/time/single-record
requirements remain. Actual probe orchestration rejects malformed Slog-derived records.

Reviewed classifier/tests unchanged from endpoints; unrelated later AbortInstall change is
outside these verdicts. Historical overlays inspected, not rerun or relabelled as independent
execution. All handles terminal, no edits/mutations/worktrees/resources/cleanup by reviewer.
These actual local file/gate/parser/probe checks use synthetic bundle/runtime inputs, not
executable release artifacts, Docker/SQL, clean-host R-006, full Deployment or release proof.
Original R18/R20 dispositions remain CHANGES REQUIRED; normative/archival gates unwaived.

## 2026-10-15 — Authenticated stream framing independently approved

Review by: fresh independent Codex `/root/backup_stream_corrective_review`.
Recorded by: Codex parent under Marco's corrective-code delegation. **APPROVED**, no findings,
all seven paths of `4627861a0df1dce9a84c88d56fee80b4b8cec9e1..cd609d0711eea9a46be0b43fa2ab2ca5bf5b93e2`.
RP-507 drains after tar EOF before joining the decryptor, refuses any plaintext remainder
and closes the reader on failure. Exact-package success remains; no usable dump escapes a
failed extraction, and caller refuses before restore/removes temporary work. No format,
retention, key, timeout or policy change.

Actual descendant38dfc002/product77fe2275/kernel221: four affected packages/vet14023 pass
cold. Focused race1387 executes all7 TestPackage parents including4 authenticated-tail cases;
other selected packages have no matching tests, not additional race coverage. Declared real
Linux-amd64/Postgres backup31073 passes all8 current integration parents, no skips. All4 empty/
populated/direct/query-override populations reject authenticated tails while target remains
clean, then intact restores match recovery identities. Formatting/full-range whitespace pass.
Current extraction/tail hunks retained; later creation/query/fixture changes inspected as
execution context only. Original3 deadlock failures inspected, not independently reproduced.

All handles terminal. Initial inventory only2 existing BuildKit containers; target creates/
removes identified tmpfs7d10afbd… plus its new network. Final service inventory empty; builders
preserved, only disposable SQL data discarded. Reviewer edits/mutates/commits nothing; root's
separate research preserved. No metadata proposal/rights adoption, exact-release/clean-host/
R-006/whole-Deployment/hosted-CI/archival/release approval. RP-505 and other holds remain.

## 2026-10-15 — Creation and rehearsal chronology independently approved

Review by: fresh independent Codex `/root/deployment_chronology_corrective_review`.
Recorded by: Codex parent under Marco's corrective-code delegation. **APPROVED**, no findings,
all paths in both complete baseline-exclusive ranges:

- RP-548,7 paths: `dac25388d3cfbba54037a04e9e518b41c42739fd..88868d347d384d42dc22c6159f992c7a0c132cef`.
- RP-549,8 paths: `88868d347d384d42dc22c6159f992c7a0c132cef..5d70f673388a981a75d4dd2bb74e6096635fac6a`.

New-create report bracketing preserves old-backup restores, host/header equality and lack of
minimum duration. Executor refuses invalid chronology before launching the next producer;
base/final loaders independently check canonical per-command intervals after hash rebinding.
Actual descendant38dfc002: five cold packages55065 pass (release9.257s,backup.778s,rehearsal
7.230s,both CLIs.270/.263s), matching vet pass. Focused cold race19516 executes all20 creation
timestamp cases/retained header binding,6 executor and4 rehashed result sequences through
both loaders (1.700/3.019s). Actual authorized make test-deployment-rehearsal54874 passes4 cold
packages and both retained-build-record validators. Formatting/both full-range checks pass.

Initial unquoted race selector parses no tests; corrected quoted selection supplies evidence.
Initial local rehearsal23056 hits localhost EPERM; exact authorized target passes, not a product
failure. Chronology source/tests and RP-548 timestamp hunks unchanged; later AbortInstall and
classifier changes in shared docker.go are context only. Historical failures/severings inspected,
not rerun. All handles terminal, no edits/mutations/resources/cleanup by reviewer. Actual age/
local subprocess/file/loaders use controlled clocks/synthetic producers, not Docker/SQL or
trusted snapshot/RPO/RTO/clean-host/current-release-artifact proof. RP-140 executable identity,
original R10/R12 CHANGES REQUIRED and normative/full-feature/archival/release gates remain.

## 2026-10-15 — Rollback package preflight independently approved

Review by: fresh independent Codex `/root/rollback_preflight_package_corrective_review`.
Recorded by: Codex parent under Marco's corrective-code delegation. **APPROVED**, no blocking
findings, all eight paths of
`cd609d0711eea9a46be0b43fa2ab2ca5bf5b93e2..a5a2e94a50154df0301110dfa50718b3fbd9c1f4`.
DP5/DP6/AC5 caller checks the authenticated full package before StopFailed/ResetDatabase;
existing identity/path/header checks remain. Private directory/0600 dump/deferred removal
prevent retained plaintext workspace. R18's validated previous manifest/epoch restores its
healthy-first fixture with explicitly synthetic dump, not invented restore evidence.

Actual descendant064c54ee/product77fe2275/kernel221: five affected cold packages/vet6555,
backup/release/rehearsal cold race71420 and focused verbose87574 pass. Actual preflight,
all9 rollback-stage failure cases and R18 controls execute, including encrypted non-package,
trailing-data and changed-dump/old-hash refusal; setup failures cannot count as successful
negatives. Complete-range whitespace check passes. Actual make test-deployment-release33922
passes2.47s: real Linux-amd64/Postgres/Caddy/HTTP/WebSocket backup passes preflight; its
authenticated tail refuses before runtime commands. Readiness/admitted-work drain/courtesy/
socket closure/restart/intact restore/account count/authenticated smoke/severed artifacts pass.

Current RP-508 hunks/consumers retained; later shared docker.go corrections inspected as
context only. Historical wrong-path/missing-import/healthy-fixture failures remain historical,
not reviewer executions relabelled green. All handles terminal; target removes its new
isolated tmpfs Postgres/Caddy/network, final inventory equals initial2 BuildKit containers
with no Compose projects. No persistent storage/cache/builder cleanup, source edits/mutations
or commits. Root's separate Decimal research preserves active arithmetic. No actual live-
volume rollback, clean-host R-006, production-size extraction/RTO, D-019, metadata adoption,
full Deployment/normative/archival/release approval. This correction no longer waits on Claude.

## 2026-10-15 — Complete-dump creation and archive parser independently approved

Review by: fresh independent Codex `/root/backup_creation_archive_corrective_review`.
Recorded by: Codex parent under Marco's corrective-code delegation. **APPROVED**, no findings,
all source/test/docs/records paths inspected, separately for complete baseline-exclusive ranges:

- RP-510,7 paths: `e27eb83211764243166a1d9b640083c33a729356..abf4b4acd56baf6e7e91312cbddb28f9d1525c31`.
- RP-511,6 paths: `abf4b4acd56baf6e7e91312cbddb28f9d1525c31..102face6f2aeadc8aaa0ff19af0d4d5bd0f63ef1`.

Creation rewinds before hashing, verifies actually streamed second-pass hash and exact EOF,
and propagates changed input through atomic encrypted publication refusal. Fixture hides
bytes.Reader.WriteTo and asserts exactly one between-pass change; refusal leaves no header/
path/files. Tar.Next errors are checked before header dereference; direct/actual assembler
refusals publish no manifest. Intact controls remain; no format/retention/time/key/policy expansion.

Actual descendant38cba707/product77fe2275/kernel221: seven cold packages/vet57503 pass;
assembler CLI has no separate tests, actual assembler executes in regressions. Cold race69377
passes3 reader-position/4 between-pass cases, existing lock/pipe-unblock,6 direct malformed/
6 actual-assembly arms and intact controls. Declared Linux-amd64 seven-package78104 passes.
Actual backup76320 executes all8 integration parents, no skips: all4 empty/populated/direct/
query-override positioned pg_dump populations restore with exact recovery identity, authenticated
tails preserve a clean target. Extra deletion/credential observation is context, NOT rights
approval. Actual release49033 passes2.46s with Postgres/Caddy/HTTP/WebSocket: backup preflight,
tail refusal before runtime calls, drain/courtesy/closure/restart/intact restore/account count/
authenticated smoke/damaged artifacts. Both range whitespace/format checks pass.

Creation source unchanged from endpoint; later shared query/fixture additions are context.
RP-511 error-first hunk/regressions retained; later layer-identity validation is not approved
by these scopes. Historical invalid changed-source-helper and compilation results remain
invalid/historical, not new executions. Synthetic archive/ELF inputs are not runnable release
images; composed release helper is not installed current/previous OCI proof. Docker28.4.0/
Compose2.39.4-desktop.1 daemonLinuxarm64, declared amd64 under emulation, not native clean-host.

All handles terminal. Initial/final inventories equal2 existing BuildKit containers, no Compose
projects. Targets remove own isolated resources; separately remove only newly created twice-
inspected tmpfs testc6b19024…; disposable SQL data unrecoverable, persistent storage/cache/
builders preserved. Reviewer edits/mutates/commits nothing. No snapshot/metadata/D-019 adoption,
R-006/full Deployment/hosted-CI/normative/archival/release approval. These corrections leave
the pending-review queue; current artifact, authored and full feature gates remain open.

## 2026-10-15 — Ordered image-layer identity corrective review approved (RP-513)

Review by: /root/image_layer_identity_corrective_review (fresh independent Codex reviewer,
Marco-delegated corrective code). Recorded by: Codex (root, transcription). APPROVED full
13621c7de429b0bc5764cea6cb0061563df40fa4-exclusive through
c7071178c5e2a70b5272f09c47940f638f0596cd-inclusive: one commit/eight paths, including
all production/tests/docs/ledger/plan/log, under accepted DP1/AC1/AC8. Initial source
4d847020/product77fe2275/kernel221, Go1.27.2 Darwinarm64. No new contract adopted.

Independent old-validator overlay reproduces ten changed/reordered/missing-DiffID accepts;
two intact plain/gzip controls pass. Initial overlay extra-brace compile failure is not
regression evidence; corrected substantive run71fb66 establishes the failures without
tracked mutations. Corrected focused77939a passes12 ordered-layer,5 codec/read and6
malformed-assembly cases. Cold six-package70352 passes (assembler CLI no separate tests);
same vet passes. Race73040 passes releasepackage/deploymentrehearsal and27 downstream
bundle mutations/controls; the descendant consumer is supporting evidence, not an enlarged
review scope. Commands use root make test-go/vet, cold selectors and ignored overlays.

Actual retained archive→current ValidateBundle integration3157 independently passes both
74-artifact source102face6f2aeadc8aaa0ff19af0d4d5bd0f63ef1 candidates,0.26/0.27s.
Release0.1.0-preview.2; manifest061375c80401d5aaaac172c894837cb4e3ca082a59250e7ad40b635338a46073,
archive0ffddcec5503ff6b20691a4a1039057e7eb3ad502e13b65cebc77033fc2e0d4b,
8526848bytes; config0a6a731624aabc87cf0fb4df42afa8f35ba36b7c50e9455689dc1e48494a9677.
These helpers predate the correction; final-source rebuild is still required. Native
Caddy/SQL integration explicitly SKIPS (72299), not credited. No new Linux, Docker loading,
runnable synthetic-image, supported predecessor-pair, clean-host R-006, CI or release proof.

Full range whitespace passes. All handles terminal; four ignored review overlay/source
files only, no tracked edits, commits, Docker resources or unrelated cleanup by reviewer.
This closes only RP-513's corrective review; original feature/author gates remain open.

## 2026-10-15 — Same-second backup alert corrective review approved (RP-518)

Review by: /root/backup_alert_outcome_corrective_review (fresh independent Codex, Marco's
corrective-code delegation). Recorded by: Codex (root, transcription). APPROVED complete
24a343b80e2913225aeafc27548833a3265f6e6a-exclusive through
b4013e99c613c31888ecf4ecdca4a68f38ba553a-inclusive, one commit/all11 paths. Source
54eec084cbae491aed4007254ffe0abedc304a61; accepted DP7/AC7, no new operations contract.

Cold six affected packages/vet30831 and race outcome/scheduler17771 pass. Actual old
RecordOperation overlay92331 fails all9 operation/time-order rows, publication-failure and
equal-timestamp scheduler regressions; recovery assertions after failure are not credited.
Original package snapshot68199 passes. Original pinned Promtool84289 passes; old rule24649
fails exactly both15-second failure assertions. Independently rederived amd64 child
dd4bced05dfaddf23a7ec50f87334993a4149f7fcfbf58456d1c8bafce91cd13 from index69f52414….

Original b401 Compose/config/package receiver population44602 passes130.89s in isolated
cloud-clicker-rp518-review: real node-exporter→Prometheus→Alertmanager→HTTP receiver,
equal-timestamp failure and successful retry resolution matched to that firing's start,
rejecting receiver, five targets and Caddy-admin peer refusal. Unrelated startup alerts
cannot satisfy the backup-specific oracle. Outcomes/clock are controlled, not real pg_dump;
scheduler uses actual encrypted-target admission with synthetic dump. Legacy timestamps,
newer-failure and deadline arms remain. Publication is individually atomic, not transactional.

Docker28.4.0 ARM64/Compose2.39.4-desktop.1, mixed service architectures; no all-amd64 clean
host, independent full SQL/restore, seven-family AC7, R-006, CI or feature/release approval.
Initial sandbox denial and unmatched metadata glob ran no tests. Reviewer changed no tracked
files/commits. All handles terminal; removed only inspected isolated services/network/volumes;
final inventory is original two BuildKit containers. Images/ignored reproductions retained.

## 2026-10-15 — Mounted archive reader mechanical experiment (before execution)

Question: can actual Docker archive reads identify all nine CR2 proposed release-owned
mounted roots, rather than trusting host source paths? Manual diagnostic only, no production
observer/CI lane or draft adoption. Validate retained source102face6 bundle-a first; mount
copies of its nine roots into four isolated STOPPED containers using exact manifest image
configuration IDs. Read tar streams via docker cp without following links; compare regular
file sets, lengths and SHA256 plus directory/type admission, and require process completion.
Intact roots must pass; changed site/helper/manifest, missing child, symlink and interrupted
archive must refuse. Derive expectations before mutations; never change retained originals.
No services start, ports, network or credentials. Inspect exact owned IDs before removal.
Docker28.4.0/Compose2.39.4-desktop.1 match retained metadata, daemon linux/arm64 with
linux/amd64 images. This is a bounded Desktop mechanism check, NOT supported native clean-host
CR2 closure, seven-service observation, current-source release evidence, CR4 quiescence or R-006.

Separate lifetime question (before execution): after an actual ExecRunner helper acquires
the operator lock and starts Docker exec, do helper death AND owned CLI death terminate
the already-started daemon-side process? Use one isolated no-network/read-only Caddy-image
container with a writable disposable handshake directory, no game/services/credentials.
Normal completion control plus killed-helper/client case; require closed gate before deaths,
released lock before gate, then explicit release and observe completion/process disappearance.
This tests cancellation as a proposed quiescence mechanism, not an adopted recovery policy.
Same ARM64-daemon/Darwin-client limit; no daemon restart, privileged container or deployment.

## 2026-10-15 — Mounted-byte approach works; Docker-client death is not quiescence

Outcome under unaccepted RP-558 draft: bounded mechanical evidence only. All nine actual
mounted archive roots pass exact expected sets/lengths/hashes using four stopped read-only,
no-network exact-image fixtures. ValidateBundle first accepts retained source
102face6f2aeadc8aaa0ff19af0d4d5bd0f63ef1/74 artifacts. Caddy configaf555904…,
Postgres75f5a969…, Prometheusd2f7aaa3… and Alertmanager6519630d… independently inspect
as linux/amd64. Test expectations derive from the validated original; only disposable copies
are changed. Same-length changed site/helper/original-manifest bytes refuse; restoring each
passes. Missing site child and symlink refuse. Kill actual docker-cp after its first helper
header:32,256/15,298,684 body bytes followed by unexpected EOF and killed process, not a
manufactured test error; subsequent intact read passes. Root make test-go
GO_PACKAGES='./deploymentrelease' GO_TEST_FLAGS='-run TestManualMountedArchiveResearch -v
-count=1', session63811, PASS2.18s. Initial73569 fails before subject execution because Go's
package test cwd makes the initial bundle path wrong; correct to ../../.cache, no oracle change.

Separate real daemon lifetime experiment75927 PASS1.28s: same root selector with
TestManualDaemonLifetimeResearch. Actual ExecRunner under operator lock launches the real
Docker CLI (transparent PID-recording fixture wrapper). Real daemon exec writes entered and
blocks on explicit file gate. Normal completion control passes. Killed case confirms held
lock and absent completion, kills helper AND owned CLI, observes both terminal, reacquires
lock and still sees no completion. Opening the gate THEN produces completion; independent
daemon /proc check observes exec process gone. Thus killing local clients cannot establish
no future daemon mutation, even with stable pre-gate observations. No arbitrary sleep decides
the result; polling guards fail the experiment if exhausted. This is not an adopted supervisor
or recovery mechanism and must not become an automatic test demanding unsafe survival.

Manual sources preserved as probes/mounted-archive-probe.go.txt and
probes/daemon-lifetime-probe.go.txt, outside automatic CI. Temporarily place as *_test.go in
server/deploymentrelease to run the named root Make selectors. All test-only Go sources were
gofmt'd. Draft author reconciles CR4 and evidence section; no production, accepted mechanics,
workflow, thresholds, artifacts, secrets, copy or kernel change. Review by: Codex (instrument
author first filter). Recorded by: Codex. Research/draft amendments are NOT approved by the
corrective-code delegation; designated normative/adoption gate stays open.

Docker28.4.0/Compose2.39.4-desktop.1, Darwinarm64 client/Linuxarm64 Desktop daemon; amd64
images do not turn it into a native clean Linuxamd64 host. No seven-service observer,
command/start-identity/secret-leak matrix, production archive framing/duplicate/traversal
admission, true release request drain/start interruption, CR2/CR4 closure, R-006 or whole CI
claim. Engine metadata/copy/exec documentation are partial support only. Index pulls for
Postgres/Alertmanager report existing digest overwrite errors after layer download; exact
config objects nevertheless inspect correctly. No image/cache deletion or daemon intervention.

All owned fixtures' IDs/labels inspect before removal; lifetime PID1 exits0 normally; only
their disposable anonymous volumes are removed. Final docker ps inventory returns exactly
the two preexisting BuildKit containers. Temp fixture data is disposable/unrecoverable;
retained release bundles, caches, builders and ignored review evidence remain. No deployment,
publication or push. Next: specify/prove a mechanism that accounts for daemon-side outstanding
work before accepting the recovery contract; continue independent corrective reviews meanwhile.

## 2026-10-16 — ExecID completion candidate falsification (before execution)

CR4 mechanical question, not implementation: can a persisted Docker ExecID's terminal
Running=false/ExitCode0 establish that no descendant can mutate the project afterward?
One owned exact Caddy-image/no-network/read-only container, disposable handshake bind only.
Actual local Docker28.4.0/API1.51/Compose2.39.4 Desktop pair, Linuxarm64 daemon/Darwinarm64
client with pinned amd64 image; still NOT native supported clean-host closure.

Run ordinary direct completion and a parent shell that forks one gated child then exits0.
Inspect only ID/ContainerID/Running/ExitCode/Pid through the actual versioned Engine API;
check created/not-started state, terminal parent, and restart409. For the forked arm require
parent terminal before explicit gate opening, absent completion before gate, then actual
child mutation afterward. No arbitrary wait decides success;10s handshake/1min global guards
fail the experiment on exhaustion. This may reject ExecID-only recovery, never approve a
supervisor from sampled metadata. No secrets, daemon restart, privileged container, deployment,
data/service alteration or automatic-CI assertion that unsafe survival is desirable.
Manual probes/exec-terminal-probe.go.txt is exposed only by ignored compiler overlay.

Result74156 terminal0/package2.918s: ordinary control and forked-parent negative execute actual
API create/start/inspect. Both created states return Runningfalse/ExitCode0/Pid0. Ordinary
terminal retainsPid49123; forked parent terminal retainsPid49144, both Runningfalse/ExitCode0.
Both repeat-start requests refuse409. In forked arm the child has entered with gate closed,
parent is terminal and completion absent; explicitly opening gate THEN produces its completion.
Thus a completed exec parent is not a generic no-future-descendant-mutation fence. It does not
show actual pg_dump/Compose recipes leaving descendants or a queued original-start request.

Root command: make test-go GO_PACKAGES='./deploymentrelease'
GO_TEST_FLAGS='-overlay=/Users/stronk/repos/cloud-clicker/.cache/exec-terminal-probe-overlay.json
-run TestManualExecTerminalResearch -v -count=1'. Manual source outside CI; actual API1.51
and Docker28.4.0 arm64 metadata independently read. First metadata format uses wrong ApiVersion
capitalization and partially fails; corrected APIVersion returns28.4.0/1.51/arm64. No failed
format is passed engine evidence. Context discovery reads only the local socket endpoint.
No whole environment/secret configuration/output, daemon restart or privilege change.

Owned fixture opens its gates on cleanup, verifies exact full ID/label twice, observes PID1
terminal0 then removes only it/its disposable anonymous volumes. Disposable fixture data is
unrecoverable; retained bundles/two existing builders/caches preserved. No production, accepted
RFC, CI/kernel/balance, deployment or publication change. Draft author reconciles CR4 and
evidence without claiming mechanism closure. Review by: Codex root (research/draft author
first filter). Recorded by: Codex root. Complete new rangee140405d-exclusive through this
containing commit needs research/normative review outside corrective-only delegation.
Next: settle the actual operation-specific lifetime fence, not another stable-snapshot rule.

## 2026-10-16 — RP-543 review: CHANGES REQUIRED, later gzip metadata missed (RP-675)

Review by: /root/archive_metadata_corrective_review (fresh independent Codex under Marco's
corrective-code delegation). Recorded by: Codex root, transcribing its verdict. Complete6-path
cf8c6f08a349daae43263cc75ccfced5d80457d2-exclusive..ca447daabeb7ae9ae703d9d982542a8fca814706-inclusive.
Decision: CHANGES REQUIRED. All scanner/test/docs/backlog/plan/log paths inspected. Default
gzip multistream consumes later headers without exposing them through the first Reader.Header.
Independent digest-valid image fixtures miss all3 second-member name/comment/extra sentinels;
ordinary multistream control passes, ignored overlay0bd212/0.218s/Make2. No real secrets emitted.
Original metadata omissions/unsafe paths also independently reproduced on historical scanner.
Current original42-case/cold Releasepackage/Rehearsal/two CLIs86632 pass1.621/7.936/.109/.166s,
vet passes, race68756 passes1.514s. Those passes do not rebut the new finding. Whitespace passes.
Executioninitiale140405d, exact scanner/test endpoint bytes; Go1.27.2/Darwinarm64. No Linux,
Docker/SQL, repeated historical-image scan, current artifact, R-006 or release acceptance.
Reviewer has no tracked edits/commits; original R3 negative stays. No delegated normative review.

Root corrective outcome under SAME DP3/AC2: scan every gzip member's header, concatenate
all uncompressed bytes before nested parsing and retain ONE unchanged decompressed ceiling.
Use bytes.Reader's ByteReader boundary with Multistream(false)/Reset; malformed/truncated
later members and trailing junk still refuse, including aggregate-bound crossing. Add actual
digest-valid second/third-member metadata and split-payload controls plus small-limit helper
tests of the same production bound accounting, not a reduced runtime limit. Corrective range
starts afterbdf5add1; no secret pattern, image/artifact, kernel, workflow or accepted policy change.

## 2026-10-16 — RP-546/RP-547 independent corrective approval

Review by: /root/deployment_drain_orchestration_corrective_review (fresh independent Codex
under Marco's corrective-code delegation). Recorded by: Codex root, transcribing its verdict.
Decision: APPROVED, no findings, two complete base-exclusive/tip-inclusive ranges:

- RP-546 39019517b7314e12db478e4ec363301521a3431b..de0ca191ba34dbf2272a581567bc22250821166a (8paths).
- RP-547 de0ca191ba34dbf2272a581567bc22250821166a..dac25388d3cfbba54037a04e9e518b41c42739fd (5paths).

All source/tests/docs/plan/log/backlog inspected. Historical request overlay61154 reproduces
all8 endpoint panics. Compiling ReadinessDown forgery84065 leaves3 old R9 tests green but
fails new unmarked-readiness51103; CourtesyFrame forgery22425 fails missing-courtesy case.
Intact cold six affected packages31659/vet pass; race13006 executes all11 real outer-method
HTTP/WebSocket cases,8malformed cases and readiness matrix,8.945s. Native19067 explicitly
confirms Caddy/DB-backed children/real-minute cleanup skip; parent PASS is not DB credit.
Three initial terminal-dollar regexes are consumed by Make and execute nothing; corrected
selectors supply the counted evidence. Exact-range whitespace passes.

Executione140405d: relevant drain/readiness/new tests endpoint-identical; later backup/
response-decoder changes outside verdict. No tracked edits/main mutation/commits/Docker
resources. Protocol observation/recording-runner coverage, not actual Docker stop/Caddy/SQL/
auth/work completion/job flush/exact artifact/clean host. Original R9 CHANGES REQUIRED,
full AC3/R-006/feature/archival/release gates remain. Other new scanner correction separate.

## 2026-10-16 — RP-675 every gzip header, one unchanged concatenated payload/bound

Root correction: use a ByteReader-backed member reader with Multistream(false) re-applied
after every Reset. Scan every Name/Comment/Extra, accumulate decoded bodies under the SAME
1GiB combined ceiling, then recurse once. Preserve compressed member/depth limits, ordinary
finding paths and redaction/rules; later CRC/truncation/trailing errors still fail closed.
The internal helper takes a bound to test identical accounting with small physical inputs;
the only runtime call supplies the original maximumSecretScanDecompressed constant.

Independent original reviewer repro now passes69741/0.305s. Root baseline20313 uses the exact
bdf5add1 gzip caller branch in current source; the new helper remains unreferenced solely to
compile its tests. This is a historical-caller overlay, NOT a full historical checkout.
All6 second/third-header metadata cases miss, all4 ordinary/split/nested-secret controls pass,
Make2/0.223s. No tracked-source mutation or synthetic parser error. Corrected32628 passes
all20 new metadata/payload/combined-bound/integrity cases,0.272s. Every image fixture first
passes actual Docker archive/config/uncompressed-layer-digest validation. No secret emitted.

Cold affected Releasepackage/Rehearsal/scanner CLI/rehearsal CLI72535 pass
1.446/7.682/.413/.270s, matching vet passes. Race52896 executes SecretScan population1.423s.
Declared Linux-amd64 full Releasepackage5091 executes without skips1.139s, including existing
archive/layer/bundle/metadata tests and all added cases. Its otherwise-unused SQL fixture is
NOT DB coverage. Ordinary make release-secret-scan with retained102face6 bundle-a image69736
passes2366 tracked working files/no findings; this dirty-tree developer scan creates no
structured release result and does not label the historical image current or rebuilt.

Only twice-inspected owned tmpfsfe9e4ae1… stopped/removed after terminal; disposable data
unrecoverable, durable data/caches/two existing builders preserved. Gofmt/diff checked, all
handles terminal. Review by: Codex root (implementer first filter). Recorded by: Codex root.
Complete new source/test/docs/backlog/plan/log rangebdf5add1-exclusive through this containing
commit requires independent corrective re-review. The original RP-543/R3 negatives remain
historical, not relabelled approval. No kernel/gameplay, secret pattern, limit, policy, budget,
workflow, artifact regeneration, full Deployment/CI/clean-host/archival/release approval.

## 2026-10-16 — RP-675 corrective re-review: APPROVED

Review by: /root/archive_metadata_corrective_review (fresh independent Codex under Marco's
corrective-code delegation). Recorded by: Codex root, transcribing its terminal verdict.
Decision: APPROVED, no findings, complete6-path single-commit range
bdf5add19e43eddaac18afd3d45109f79f44e5a7-exclusive..76cba7bcbe30b646e1ca11101abfbc367f915d91-inclusive.
All scanner/tests/canonical docs/backlog/owning plan/log inspected at exact reviewed endpoint.
Runtime1GiB and other patterns/depth/policy preserved, every member header scanned, joined
payload parsed once; corrupt later streams/refusal return no usable partial output.

Independent historical-caller97355f misses all6 later-header cases,4 payload controls pass;
unused new helper compiles tests only, not a historical checkout. Original independent
three-negative/one-control reproff74c9 now passes0.095s. Cold affected4packages52194 pass
1.493/7.701/.082/.162s, matching vet passes, race90591 passes1.515s including20 additions
and old metadata/redaction/body/V7 tests. Full-range whitespace passes. Scanner SHA256
55e965cea0a9950d9dc02254218b9568acb7f0a130f87566c3f32600f944b788;
test4e1a9f11f14161841fd343765a6086bb17709119a739238c1e438496cccb75db.

RP-546/RP-547 record attributions/ranges/limits inspected, no enlarged runtime verdict.
Original RP-543/R3 CHANGES REQUIRED stays historical; RP-675 defect closes. Root's Linux/
historical-image executions remain separately attributed, not independent repeats. No new
Linux/image/Docker/SQL resources, edits, commits or publication by reviewer; all handles
terminal/tree clean. No current-artifact/R-006/clean-host/full Deployment/CI/normative/
archival/release approval. Rebuild/re-scan the actual future artifact under its release gates.

## 2026-10-16 — Actual Compose one-off lifetime (before execution)

CR4 next question follows the actual runtime, not arbitrary exec recipes: docker.go uses
Compose run --rm --no-deps for config validation, backup create/inspect/restore, recovery
identity and alert tests. It has no production Docker exec invocation. The earlier exec/child
negatives reject generic completion assumptions, not an observed leaked backup or full
current-command failure. Does killing the ENTIRE owned local Docker/Compose process group
stop a real one-off job already inside the daemon? Same pinned Engine28.4.0/API1.51/Compose
2.39.4 Desktop pair and exact amd64 Caddy image, Linuxarm64 daemon/Darwinarm64 client.

Two populations: ordinary completion and SIGKILL of a new, exclusively owned CLI process
group after the daemon job enters a closed file gate. Actual temporary Compose project,
network_mode none, read-only rootfs, no ports/credentials/database/persistent volume, one
writable disposable handshake bind. Explicit unique name aids observation; production today
does not assign that name. Require complete parent Wait, actual identified one-off still
running/no completion before release, then explicit gate→completion→automatic removal.
The ordinary control must exit0. Poll every10ms;10s handshake/1min whole-experiment limits
are invalidation guards, not operating bounds or arbitrary no-write sleeps. Runtime payload
is controlled shell, not pg_dump/restore, and cannot authorize backup/release acceptance.
Manual probes/compose-job-lifetime-probe.go.txt remains outside CI; no process-runner fix,
supervisor/protocol adoption, secret access, daemon restart, deployment or publication.

## 2026-10-16 — Actual Compose one-off lifetime: both populations executed

Outcome: process-group cancellation is not sufficient for CR4. Root executed
`make test-go GO_PACKAGES='./deploymentrelease' GO_TEST_FLAGS='-overlay=/Users/stronk/repos/cloud-clicker/.cache/compose-job-lifetime-overlay.json -run TestManualComposeJobLifetimeResearch -v -count=1'`.
Session93532 terminal exit0; ordinary case0.48s, killed-group case0.39s, test0.87s,
package1.250s. Both actual identified one-offs remain behind the closed gate until explicit
release; in the killed case the whole exclusively owned local CLI group has been SIGKILLed
and parent Wait is terminal first. The daemon job is still running, then writes only after
the gate opens. Both containers auto-remove; successful daemon lists confirm absence.

This follows actual `docker.go` one-off commands, unlike earlier generic exec diagnostics.
Explicit unique name is observation plumbing, not existing production identity. Exact image
sha256:af555904a0961945f16bb323a501457b13a4f7e9bde969b145b97da80b38ecbe,
Engine28.4.0/API1.51/Compose2.39.4 Desktop, amd64 image on Linuxarm64 daemon with Darwinarm64
client. Guards did not fire; no skipped population. Post-run complete container inventory
contains only the two pre-existing buildx builders bb56ca0ca798 and17c2013f1491, preserved.
Disposable temp handshakes are removed by the test; no durable data was touched.

Draft CR4/body/evidence/owning plan now exclude process-group-only cancellation. No safe
replacement, delayed-request/whole lifecycle fence, real pg_dump/restore, native-host or
R-006 proof follows. No product code, CI lane, adoption, archival, push or deployment.
Review by: Codex root (research author/first filter). Recorded by: Codex root. Complete range
bda1b086-exclusive through this containing research commit still needs research/normative
review under existing authority; corrective-code delegation does not approve that scope.
