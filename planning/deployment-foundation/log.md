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
