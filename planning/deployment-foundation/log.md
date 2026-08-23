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
