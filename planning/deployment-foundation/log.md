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

## 2026-10-16 — RP-169 browser-selection correction independently approved

Review by: /root/cosmetic_keyboard_network_corrective_review (fresh independent Codex under
Marco's corrective-code delegation). Recorded by: Codex root. APPROVED complete
8b4db5a95377df5631961832a84727a070a938c3-exclusive..c3311dbd04c92b084c632182d46691b987712fe2-inclusive;
all10 paths inspected. Exact exclusion removal reproduces the CSS fixture's duplicate import
failure in all three Linux ARM64 engines; fixed current105 browser checks exclude only the
Node file while its actual4 Node fixtures pass. Full commands/provenance/limits are in the
Cosmetic log's same-date keyboard/request-observer verdict. No CSS provenance behavior,
DeploymentAC8/rights/clean-host or whole CI/release gate is approved by this test-config repair.

## 2026-11-01 — Backup, attribution and rehearsal safety corrections approved

All named complete ranges APPROVED, no blocking findings. Recorded by: Codex root,
transcribing fresh independent reviewers under Marco's corrective-code delegation.
Every source/test/docs/dependency/ledger/plan/log edge in each range inspected;
no intervening commits approved. All verification runs at descendant `a9892a48`,
not historical checkouts. No tracked reviewer edits or redundant source mutations;
retained old-source failures establish the inspected regressions' discrimination.

**Review by: `/root/backup_writer_safety_corrective_review`.** Separate nine-path
ranges `d0fc603a..9265049b` (RP-466) and `ca447daa..fad86116` (RP-544), base
exclusive/tip inclusive. Atomic no-replace commit preserves winning encrypted
bytes/restorability. Shared creator/exclusive cleanup locks preserve live aged
temporaries; failed writers release ownership, package pipes retry without deadlock
and scheduler contention defers without false result metrics. Cold complete backup/
helper/release/rehearsal packages PASS1.233s/0.486s/9.918s/27.162s; matching vet PASS.
Focused native race PASS2.563s/1.369s. Actual declared Linux-amd64 filesystem tests
PASS0.402s/0.232s, including real second-process cleanup, no-replace destinations,
target rechecks, failed-owner release, package extraction/retry and both scheduler
directions. `make test-deployment-backup` executes all eight current real Postgres
integration functions with zero skips: actual encrypted pg_dump/pg_restore,
empty/populated/clean-target/identity/deletion-restore cases. Its disposable project
tears down normally. Later RP-611 identity strengthening is preserved, not newly
approved here. Ordinary native DB skips receive no persistence credit.

```sh
make test-go GO_PACKAGES='./deploymentbackup ./cmd/deployment-backup ./deploymentrelease ./deploymentrehearsal' GO_TEST_FLAGS='-count=1'
make vet GO_PACKAGES='./deploymentbackup ./cmd/deployment-backup ./deploymentrelease ./deploymentrehearsal'
make test-go GO_PACKAGES='./deploymentbackup ./cmd/deployment-backup' GO_TEST_FLAGS="-count=1 -race -run 'Test(CreateDoesNotOverwrite|EnvelopeRename|Cleanup|CreateExcludes|PackageCreationUnblocks|ScheduleDefers)'"
docker compose -f compose.save-test.yml -f compose.ci-test.yml run --rm --no-deps test go test -p 1 -count=1 -v -run 'Test(CreateDoesNotOverwrite|EnvelopeRename|Cleanup|CreateExcludes|PackageCreationUnblocks|ScheduleDefers)' ./deploymentbackup ./cmd/deployment-backup
make test-deployment-backup
```

**Review by: `/root/shipped_attribution_corrective_review`.** Complete six-path
`fa2bec58..39019517` (RP-545). Every physical JS-map/CSS package copy is validated
before deduplication; identical attribution merges, conflicting/missing evidence
refuses without partial rows, distinct versions retain notices. All16 duplicate-
copy cases (12 negative/four controls) pass. Cold metadata/release packages
PASS2.043s/4.196s; vet and restored real Go dependency guard PASS (1.20s).
Fresh client build PASS217 modules/1.86s; actual metadata generation PASS53
dependencies and all five npm notices exactly once. Diagnostic SPDX/notices at
`.cache/attribution-review.NWSSal/` SHA256 respectively
`f3582e666474947d4055222f9fa9cb00c11db356481db75b408416c018015308` /
`c6a56ea62b5c232be61a2bf23dd33531a965299bdc28a3bf7b3d3b727ab554de`.
Generator/model/tests/Vite/manifest/lock inputs match reviewed endpoint; other
descendants are outside approval. No classifier/allowlist/schema/legal-policy change.

```sh
make test-go GO_PACKAGES='./cmd/gen-release-metadata ./releasepackage' GO_TEST_FLAGS='-count=1 -v'
make vet GO_PACKAGES='./cmd/gen-release-metadata ./releasepackage'
make build-client
make generate-release-metadata RELEASE_METADATA_OUTPUT=.cache/attribution-review.NWSSal RELEASE_VERSION=0.1.0 RELEASE_COMMIT=a9892a48f8fc9c5721e08cace21c8402e6436ff3 RELEASE_CREATED_AT=2026-11-01T00:00:00Z
```

**Review by: `/root/deployment_probe_baselines_corrective_review`.** Complete
seven-path/two-commit `1efea9cf..1b484c50` (RP-481/RP-482). Interrupted backup
first authenticates/restores the real healthy envelope, then requires the injected
read cause and unchanged baseline/no partial files. Metrics mutation follows
intact-bundle validation; source/image negatives follow clean scans through the
same gate. Always-accepting/rejecting gates and unrelated failures discriminate.
Focused populations PASS0.414s/0.606s; cold full rehearsal/CLI/releasepackage and
matching vet PASS. Reviewed probe.go/probe_test.go match endpoint; later fixture/
consumer changes remain outside approval. No Docker/SQL executed for this scope.

```sh
make test-go GO_PACKAGES='./deploymentrehearsal ./cmd/deployment-rehearsal ./releasepackage' GO_TEST_FLAGS='-count=1 -v'
make vet GO_PACKAGES='./deploymentrehearsal ./cmd/deployment-rehearsal ./releasepackage'
```

All handles terminal and exact-range/final diff checks pass. Only these five
corrective reviews close. Original R6/R8 CHANGES REQUIRED dispositions, other
repair ranges and current-release/backup-metadata authority remain distinct.
No production artifact/image authenticity, natural scheduling/off-host filesystem,
RPO/RTO, clean-host R-006, runtime metrics exposure or full Deployment/CI acceptance
is established by these bounded checks. No feature checkbox, archival, new policy,
publication, deployment or release approval; existing accounting/history/1.0 gates stay.

## 2026-11-01 — First-scrape restart correction independently approved (RP-522)

APPROVED complete one-commit12-path `db752b0a..bd91a9a6`, no blocking finding.
Review by: fresh independent Codex `/root/restart_alert_corrective_review`.
Recorded by: Codex root, transcribing its verdict under Marco's delegation.
All source/test/rule/docs/record edges inspected against DP7's unchanged threshold.

At clean `9feb826a`, cold `make test-go GO_PACKAGES='./operations ./cmd/deployment-operations'
GO_TEST_FLAGS='-count=1 -v'` passes1.616s/5.394s; matching vet passes. Actual
`make test-deployment-restart-observer` passes65.132s: Docker count0→3, exact
inspected creation time and age53.674502908s published without container identity.
Pinned promtool tests pass using explicitly fetched/inspected linux/amd64 image
`sha256:d2f7aaa363e1c220487d776f21a5d7a3416834826038c1d2b38f1adb9079ed93`.
This fresh rule execution differs from the historical cached ARM64 evaluator.
Go1.27.3 darwin/arm64; Docker28.4.0 linux/arm64 daemon. Gofmt/range whitespace
pass; post-run owned restart-container label census empty. Native optional SQL/
monitoring/private fixtures skip and are not credited.

Runtime Go/unit bytes match the reviewed tip. Current rule population retains
restart cases and adds later cleanup rules; current observer factors its witness
into a helper, executed with profile=false. Later receiver/profile scope is NOT
approved. Nonblocking historical hygiene finding: the original Postgres fixture
omitted `docker rm --volumes`; its image declares an anonymous data volume.
`c4279b66` already corrects that cleanup and current execution uses it. This does
not imply approval of that broader later range or deletion of any old volume.

This closes only RP-522 corrective review. Actual journal/retention, older
pre-scrape history, broader alert receiver/AC7, exact artifact/clean-host R-006
and feature/release approval remain open. No tracked reviewer changes, shared
Postgres use, broad cleanup, kernel/data/CI change, archival or push.

## 2026-11-01 — Cleanup startup/drain corrections independently approved

APPROVED RP-519/RP-520, complete contiguous `f0009bfc..82c8e57b`, all three
commits and 21 paths. Review by: fresh independent Codex
`/root/cleanup_lifetime_corrective_review`. Recorded by: Codex root, transcribing
its verdict under Marco's delegation. No actionable corrective-scope findings.
All source/test/rule/docs/record edges, actual DP5/DP7/AC7 boundaries and retained
old failures inspected; intervening/later construction is not approved.

Independent cold native five-package tests/vet and declared Linux-amd64/Postgres
operations/gameserver/commands/releasepackage checks pass. Real startup SQL and
command/logger cases execute; ordinary lanes explicitly skip the manual timer.
`make test-deployment-cleanup-runtime` passes 121.92s: genuine runtime SQL failure
and recovery (61.01s), orderly canceled lock-wait drain (60.90s). An ignored
compiler-overlay probe also exercises actual PostgreSQL SQLSTATE22012 alone and
joined with cancellation under an ended context: original errors retained,
failure1/canceled0 and bounded log, PASS0.14s. Probe files removed.

Pinned amd64 promtool passes current rules and exact historical tip snapshots.
`make test-deployment-operations-profile` exits0: primary profile PASS685.57s,
including journal-driven cleanup firing/matching natural resolution with fresh
process counter0, receiver500, real DB and dead-letter controls. The current
target also selects storage: PASS614.86s, both actual filesystem resolutions.
Sequential nonce preflight was delayed but passed without intervention; a
deduplication explanation remains inference, not a discovered defect or fix.
Diagnostic `.cache/cleanup-review-repeated-preflight.json` SHA256
`2d0c899a2027b145de5e493b67cded1b0b08fc19abc113309f2af20291591b0b`.

Executions use descendant runtime source `9feb826a`, not historical builds or
fresh old-source regressions; RP-615's separate approval is retained, not renewed.
Journal envelopes/timer are modeled, development `run` is not an OCI process,
and no supported-host journald/retention, exact-artifact R-006, full AC7 or hosted
CI is proved. The target removed its owned private services/network/textfile
volume; cleanup DB census0, exact twice-validated tmpfs Postgres container
`acc427bf1c2fc66c2c469c845aa750651074635a2527a7a6b735d23718715d95`
removed. Other projects/data/caches and root's Garden test edit preserved.
No reviewer tracked edits, kernel/data/CI changes, feature/archival/release
approval, deployment or push; existing accounting/history/full1.0 gates remain.

## 2026-11-02 — CSS resource provenance corrective reviews approved (RP-164–166)

Review by: fresh independent Codex `/root/deployment_css_provenance_corrective_review`.
Recorded by: Codex root, transcribing its verdict under Marco's delegation.
APPROVED complete ranges
`ef1a46589b5d67e7c33a362fe9aea8016a33edb1..f6e4d0ffbb45eac27727c1c9e2a9cc339705c5f6`
and `7163fe8ffd03b0f31a9c56efaa7bc8cae48f14c1..627956be6a6c2c4624f5557fdba319b76def9848`.
The intervening unrelated record is excluded; both implementation/record edges are covered.

Reviewer executes the exact historical Vite fixtures (two/four passing), historical Go
consumers via compiler overlays, and six independent omission probes: package stylesheet,
URL and worker producers plus their three Go consumers. Shipped decoded SVG/worker bytes
are absent from parsed map sources and present in the graph. Historical reconstruction
initially retained later `duplicate_client_test.go` and reproduced the known RP-545
per-copy-license failures. Those failures are disclosed, not approved: excluding only that
later test reconstructs the original population. RP-545 has its own separate verdict.

At clean `ae040d1d`, cold metadata/releasepackage checks, types, production client build,
vet, boundaries/topology and real 53-dependency metadata generation pass. Client producer
bytes match the historical tip; later Go consumer work remains outside this approval.
Range whitespace passes; reviewer made no tracked changes. This closes only these bounded
corrective reviews, not original R8, exact image/bundle rights, AC8/DP-B, R-006 or release.

## 2026-11-02 — R14/R15 witness supplements independently approved

Review by: fresh independent Codex `/root/deployment_rehearsal_witness_corrective_review`.
Recorded by: Codex root, transcribing its verdict under Marco's delegation.
APPROVED complete test-and-record ranges
`de014f6ec6359e683a50d425cd4f058a1081b9e5..6302d0ace79eddf2b2b4f33678dcc38da0b4dd26`
and `6302d0ace79eddf2b2b4f33678dcc38da0b4dd26..4a1a6577da9af8255be0ab6fd821730806e22855`.
Original Claude R14/R15 mechanisms already have their separate bounded verdicts.

At clean `ae040d1d`, reviewer executes four build-binding cases, nine mutation cases,
complete cold `./deploymentrehearsal ./cmd/deployment-rehearsal` and vet. Removing previous
binding through a compiler overlay fails the new previous-record negative with `<nil>`;
removing manifest rebinding fails all seven non-manifest cases at `untouched artifact not
rebound`. Final no-overlay focused checks and whitespace pass; temporary probes removed.
This executes current source, not historical checkouts; later changes are context only.
Mock validators establish bounded binding/probe preparation, not real bundle rejection,
rebuilding, provider-off/clean-host R-006, full AC8, archival or release approval.

## 2026-11-02 — Original native operations witnesses approved; current regression retained

Review by: fresh independent Codex `/root/deployment_native_operations_corrective_review`.
Recorded by: Codex root, transcribing its verdict under Marco's delegation.
APPROVED exact two-commit/eight-path historical range
`8f2e29d8770aefa589cdbb085861b152b7c970aa..a42e680ef8e777bb3e2a0d2b6cdbe5c0636e297c`.
The unchanged exact tip archive uses the existing repository Go cache; no main-tree mutation.

Official darwin/arm64 archives match published checksums: Prometheus3.12.0
`d758070049a4de5abbeb925b1d4540c5c38a2f40d1356f59aa17a71ac8b48be3`,
revision `9f27dffc1f93ca23287972f632025879f2d1c658`; Alertmanager0.32.1
`4fd3d3d6156cbe76016583b852dc80beda7ee693d28ac27477a9b3da06d87a93`,
revision `8768aa6f65f1a888b5aa5fbf877cf20ad45d1f61`.
Historical combined native target PASS391.431s: seven-family delivery/resolution and
healthy-but-500 refusal; actual cleanup zero→one→metric firing→natural resolution at
5m30.540s, counter still one. Cold operations/releasepackage PASS0.934s/1.378s,
vet, pinned promtool rules, historical topology/13 controls pass. Missing inputs and wrong
actual binary fail. Initial sandbox listener denial is excluded, not journey evidence.

Separately, actual current `ae040d1d` fails84.557s: its later missing-journal rule legitimately
delivers `{alertname:CloudClickerCleanupJobFailed,severity:page}` before the metric notification,
and the old waiter rejects it. Filed RP-704; the current correction is explicitly OUTSIDE this
historical approval. Native handles are terminal. Verified binaries remain in the reviewer's
private temporary directory for the parent rerun; no global installation or tracked edits.
No private Compose/Caddy isolation, actual DB/cleanup producer, other live producers, Linux
release artifact, R-006/full AC7, feature acceptance, archival, deployment or release follows.

## 2026-11-02 — RP-704 separate startup journal alerts from required metric evidence

Accepted authority: DP7, not a new observability policy. The retained independent current
failure above exposes a test assumption: one alert name can have several legitimate causes.
The native waiter now logs the exact two-label journal branch as non-credit evidence and
continues waiting for the actual metric's firing/resolution. It rejects wrong/missing metric
labels, any proof-nonce key, invalid status and missing/closed deliveries. A valid payload
cannot hide a malformed cleanup payload later in the same delivery. No shipped rule, grouping,
cadence, timer, metric producer, CI or product byte changes.

Focused cold `make test-go GO_PACKAGES='./operations' GO_TEST_FLAGS='-count=1 -run TestNativeCleanup -v'`
passes0.317s: compressed scrape plus firing/resolution, journal-only, mixed/unrelated delivery,
malformed/injected/empty-nonce, opposite-status, empty delivery and cancellation controls.
Cold complete `./operations ./releasepackage ./cmd/deployment-operations` pass
1.707s/1.541s/5.540s; matching vet and whitespace pass. Optional native/DB populations skip in
those ordinary package checks and are not credited. The explicit combined native target's
first attempt is sandbox-denied before listening (exit2), not a product regression or journey.
Its permitted rerun is active using the same verified binaries above; actual seven-family
delivery/refusal passes60.53s and the separate startup journal alert has been logged without
credit. Final natural resolution and fresh independent corrective review remain required;
the next entry records the terminal result, not an inferred pass.

Review by: Codex root, implementer first filter only. Recorded by: Codex root.
The preceding three independent verdict transcriptions close older review debt only; none
approves this new correction, any broader deployment claim or release.

## 2026-11-02 — RP-704 current real-service rerun completes

Current correction `5c43a93100c32c9fb139f37c1daff91250cfdc3c`, no subsequent code changes.
The permitted `make test-operations-native` run64238 is terminal PASS421.421s using the
same checksum-verified native services above. Seven-family firing/resolution and real
HTTP500 refusal PASS60.53s. The real Prometheus cleanup journey PASS360.70s: zero baseline
7.334s, counter one22.374s, actual metric firing1m0.375s and natural resolution6m0.384s,
counter still one. Exact journal-only firing/resolution deliveries are explicitly logged
without credit. The real five-minute window, fifteen-second cadence, one-minute grouping
and eight-minute guard remain unchanged; this duration is not an operating latency claim.
Owned process cleanup completes with the test; no reset, injected resolution or retry of
an active run. The prior denied listener attempt remains excluded.

Review by: Codex root, implementer first filter only. Recorded by: Codex root.
Fresh independent `/root/native_cleanup_source_corrective_review` is checking the complete
six-path implementation and executing its own real-service journey; approval remains pending.
This is native darwin/arm64 supplemental evidence, not actual cleanup SQL/journald/exporter,
private Compose/Caddy, supported Linux artifacts, R-006/AC7 or whole CI/release acceptance.

## 2026-11-02 — RP-704 independent corrective approval

APPROVED complete two-commit/six-path range
`ae040d1d7155fa939aa90006b97f1660534546e0..38a3297b3d8466487652ed8640b4ffddfbb9f0fc`.
Review by: fresh independent Codex `/root/native_cleanup_source_corrective_review`.
Recorded by: Codex root, transcribing its verdict under Marco's delegation.
No blocking finding. Both code and records inspected; production/rule/timing/grouping/CI
bytes unchanged. Historical verdict transcriptions retain their separate scopes.

Independent cold affected packages PASS1.752s/1.525s/5.409s and vet; focused controls
PASS0.401s, final unmodified oracle0.298s. Two private compiler-overlay probes fail as
required: crediting journal-only observations defeats both firing/resolved negatives;
early success before whole-message validation defeats eight malformed-after-valid cases.
No main-tree mutation; owned probe copies removed.

Independent actual combined native run15679 PASS422.000s: Alertmanager60.46s,
Prometheus361.30s. Zero19.445s→one34.478s→metric firing1m1.059s→natural resolution
6m1.072s with counter still one. Exact journal-only firing/resolution are logged without
credit. No reset, injected resolution, shortened window, guard change or active-run restart.
Retained archive hashes match the supplied published checksum files and actual pinned
versions/revisions. Native handles terminal; shared verified binaries retained. Final tree
and range whitespace clean at reviewed tip; reviewer made no tracked edits.

This closes RP-704's corrective review only. Native darwin/arm64 is not actual cleanup
SQL/journald/exporter, private Compose/Caddy, supported Linux release artifacts, full AC7,
R-006, hosted CI, feature acceptance, archival or release approval. Other accounting,
history-guard, contract-adoption and full nine-tier 1.0 obligations remain open. No push.

## 2026-11-02 — R18 corrective witness supplement independently approved

Review by: fresh independent Codex `/root/deployment_r18_witness_corrective_review`.
Recorded by: Codex root, transcribing its verdict under Marco's corrective-review
delegation. APPROVED complete eight-path/three-commit range
`fbbf01505160f3f5a69a6eae30bbd5a5ae923fd6..07bb3d9d43e95bda76eb3bff363fe74f51ce0332`.
No blocking findings; reviewer confirms no previous independent verdict covers it.
Original Claude R18 remains CHANGES REQUIRED, not silently approved here.

At current source `37f696af1f001f9c18ddda6ca249d83f5a7c93d8`, cold TestR18 passes0.743s:
both actual named producers, six invalid-bundle and two non-forward controls run
without skips. Complete cold affected packages pass: rehearsal7.954s, rehearsal
CLI0.249s, release9.674s, release CLI0.355s, backup0.986s, releasepackage1.422s;
matching six-package vet passes. Root Make selectors use `-count=1`.

Five independent compiling private overlays each exit Make2: constant migration
rejection fails six controls; constant missing-input rejection fails its two
invalid-input controls; omitted migration comparison, backup verification and image
preparation each defeat their corresponding prepared subject. No main-tree source
mutation. Four production source hashes remain unchanged; reviewer tree/format/
range-whitespace checks pass and every execution handle is terminal. Private
overlays are in `/tmp/cloud-clicker-r18-review.DO3LHM`; no service/DB started.

Execution is current-source, not a historical-checkout replay. Original assertions
are unchanged; later fixture scaffolding supplies layer/diff-ID compatibility for
today's validator. Actual local bundle assembly/validation and age cryptography
are exercised, but ELF/client/image-layer/dump inputs and Docker responses remain
synthetic. This does not prove executable artifacts, Docker retrieval, logical
Postgres restore, a retained release pair, wrong-image-ID comparison, clean-host
R-006, full R18/Deployment, archival or release. Only this bounded review debt
closes. Other adoption/CI/full-nine-tier gates remain; nothing pushed.

## 2026-11-02 — R20 refusal-provenance correction independently approved

Review by: fresh independent Codex `/root/deployment_r20_provenance_corrective_review`.
Recorded by: Codex root, transcribing its verdict under Marco's corrective-review
delegation. APPROVED full six-path/two-commit range
`94f0ee74e7e5fcb06328800ac14580c1767f7272..d3fde4c0c0db169ce4faeae2e19902c0f4524372`
(the base is `5f394f2a^`). No prior independent verdict covers it; no new findings.
Actual ExecRunner owns command, exit, stderr and underlying cause separately;
callers retain that failure rather than inferring a class from rendered text.

At current source `37f696af1f001f9c18ddda6ca249d83f5a7c93d8`, cold complete release /
rehearsal and both CLI packages pass9.596s/7.525s/0.156s/0.161s; relevant vet passes.
Private overlays reproduce the byte-identical historical corrected classifier and
original producer fixture: arbitrary-error, exact-record, actual shell-exit and
producer controls pass0.374s/0.304s. A separate original substring-classifier
overlay fails the permanent unattributed-error regression, Make2. Current focused
cold race checks pass1.450s/1.415s/1.312s for classifier/member-format controls,
shell exit, orchestration and backup CLI refusal classification. Reviewed-range
whitespace passes. These are root Make cold selectors, not a cached gate claim.

Historical-source overlay execution uses current surrounding code, not a complete
historical checkout. ExecRunner remains byte-identical to the reviewed tip.
RP-552's separately repaired duplicate/case-alias weakness retains its own
`fc04db24..a713137e` approval; this verdict does not claim the earlier parser
already satisfied that later requirement. Original Claude R20 remains CHANGES
REQUIRED. Actual composed CLI stderr, Postgres/clean-host producer, complete
Deployment, archival and release are not established by local shell/fixture checks.

All reviewer handles terminal; no reviewer edits/commits, main mutation, services
or DB runs. Private overlays remain in `/tmp/cloud-clicker-rp142-review.8EWzY4`.
The observed parent Markdown diff belongs to R18's separate transcription.
Only this corrective review debt closes. No runtime, policy, schema, kernel,
artifact, CI or release status is changed; goal remains active and nothing pushed.

## 2026-11-02 — Drain provenance and unsupported restart route independently approved

Review by: fresh independent Codex `/root/deployment_drain_route_corrective_review`.
Recorded by: Codex root, transcribing its verdict under Marco's corrective-review
delegation. APPROVED separately, complete ranges (base exclusive, tip inclusive):

- RP-138, eleven paths:
  `1908762fbe91417f55ee5995b57545ac652c1490..c9c9dde7acba909ba749a2ee981941eb7a40f5e1`.
- RP-143, four paths:
  `8f83e4c5f4f87f674cca4a77e39b09a347f4f185..81a970ce6850a40a9eaf0c95dc2f64c8b2b38db3`.

No earlier independent verdict covered these corrections; full diffs and current
producer/observer/plan/dispatcher/CLI inspected. At `a043b5410ffe2a6eb42a90eef1ff008007b21a28`,
cold complete config/Gameserver/release/rehearsal/rehearsal-CLI and matching vet pass.
Focused cold race covers readiness matrix, unmarked non-drain failures, marked
drain and crash-only route refusal. Three private compiling overlays fail the
correct assertions: omitted observer marker, wrong server marker and reinserted
crash-only producer mapping. No main-tree mutation; range whitespace passes.

Exact root `make test-deployment-release` PASS: real Linux-amd64 Postgres/Caddy
`TestCaddyIntegrationDrainReleaseAndExactBackupRollback`,2.73s, not skipped.
It observes marked readiness through Caddy while a SQL lock holds an admitted
intent; new work refuses, held work finishes200, courtesy/socket closure arrive,
server restarts and a real encrypted backup restores. Exclusively coordinated
fixed Compose project was initially empty; target cleanup and direct post-run
inventory confirm no remaining containers. `make build-deployment-rehearsal` passes;
actual compiled CLI probe for `gameserver_restart_during_admitted_work`, with
distinct existing input/work directories, exits2 `invalid_probe`, not false proof.

All execution handles terminal, tracked tree unchanged by reviewer; private overlays
remain in `/tmp/cloud-clicker-drain-route-review.9rlIk3`. This is current-source
integration, not historical-checkout or exact-manifest/clean-host evidence.
Compose pins architecture but uses mutable image tags; no production Docker
stop orchestration or actual crash-during-admitted-write/post-restart proof is
established. Original Claude R9/R22 remain CHANGES REQUIRED, RP-143's named
population remains unsupported, and AC3/AC4/R-006/full Deployment/release gates
remain. Only these corrective reviews close; no feature/archival/release approval,
runtime change or publication. Goal remains active.

## 2026-11-04 — Backup header and recovery snapshot corrections approved

Review by: fresh independent Codex `/root/backup_parser_snapshot_corrective_review`.
Recorded by: Codex root. APPROVED, no findings, complete eight-path range
`630f6833f57f4acc98e0df3bc6bece11e55a60cf..bc5d5f5d25e6d988ec00e3760b06cc4db7b395f5`
(base exclusive, tip inclusive), including both production/test files and original
docs/ledger/plan/log. Outer and authenticated metadata use bounded single-object
admission before dump output. One read-only repeatable-read transaction supplies
the complete recovery identity; failures return no partial identity. Formats,
hash/output definitions, retention and backup policy stay unchanged.

Fresh atfe29da5e: cold four owners/vet79525 terminal0; parser80992 terminal0,
all fourteen arms pass (eight malformed/six valid, whitespace and exact64KiB).
Private exact630f6833 parser overlay41343 exits expected Make2: all eight malformed
arms admit and emit30/35 dump bytes, while all six valid controls pass. Main is
never replaced. Declared Linux-amd64/Postgres backup93438 terminal0 executes
ALL eight integration functions, no skips, including encrypted empty/populated
restores, concurrent-write snapshot coherence, cancellation/no identity and actual
database lock release; snapshot case1.00s. Exact-range whitespace passes.

Old SQL82454 remains REUSED implementer evidence, not reexecuted. Reviewer verifies
identity.go equals reviewed tip and exact snapshot test is unchanged, SHA256
`233fe86500bd1ab057fe37e246c3be327332a38013367ce0436d56d194674394`.
Later locking/parser safeguards retain separate scopes. All handles terminal;
target output records creation/removal of dedicated tmpfs Postgres/network and final
project inventory is empty. No prelaunch inventory was captured; none is claimed.
Peer resources untouched, no tracked edits/commits; private overlay retained at
`/private/tmp/cloud-clicker-backup-review.ek90rg15`.

These corrective reviews close, not separate dump-snapshot binding, rights-policy
adoption, production RPO/RTO, clean-host R-006, full Deployment/CI, archival or
release. Next remaining DP7 writer/receiver corrective scopes RP-521/RP-523.
Root reconciles current statuses only; no software suite for editorial records,
no push/adoption/deployment, full nine-tier goal remains active.

## 2026-11-04 — Storage writer/physical proof approved; receiver review finds two defects

Review by: independent Codex `/root/operations_writer_receiver_corrective_review`.
Recorded by: Codex root. RP-521 APPROVED, complete thirteen-path
`82c8e57b514e692b68af993193f8b7b663beeab9..db752b0a19c849a7c5c989946b90f74d35a2f711`.
At sourcee1aa4c50, cold owners/vet56185 and targeted66868 pass. Private old
finite-guard72752 fails both NaN cases while four infinity controls refuse;
corrected six controls preserve valid file bytes/mtime, not a syscall-NaN finding.
Actual focused storage44377 terminal0:305.83s, persistent119.73/backup150.01s;
both20MiB subjects prove exact80%quiet→85%receiver firing→zero/exact resolution.
Initial67775 Docker denial ran none. Fixed project initially empty; own services/
network/volume removed, final inventory empty, pre-existing BuildKit preserved.
Captured binary is pre-follow-up e1aa; this is NOT proof of RP-714/715's receiver.

Original twelve-path `cdc780502bd37d09be14adf5f3a554cc27afcb41..c4279b6621d54e3b65a91aa337db0da46e396221`
has CHANGES REQUIRED finding RP-714: independent exact-helper overlay4938 and
root old-code77137 both fail firing→nonce→later-resolution ordering. Tail
requeue inverts observations and can lose resolution while seeking the firing.
Independent5528 additionally finds RP-715: canceled context plus queued valid
nonce can return success; old canceled-empty control cannot trigger it. Original
delivery/cleanup records stay source-bound; neither observation proves product
alert loss or authorizes altered thresholds/intervals/guards. Final union review
requires these defects corrected, not reinterpretation of the old verdict.

## 2026-11-04 — Receiver preflight preserves real FIFO and cancellation (first filter)

Under DP7/AC7, the actual HTTP fixture handler now demultiplexes structured
ReceiverTest controls to a separate one-message channel. The real32-message FIFO
is never drained/requeued, including firing and later resolution around the nonce.
Exact annotation identity, mixed/invalid/private inputs, all queue overflows,
cleanup mirror bytes, receiver500 and64KiB bound have actual-handler regressions.
Context is checked before reading and before success; canceled-ready controls
retain their queued nonce. No relay goroutine, runtime alert/interval/threshold,
real-alert capacity, production, migration, kernel, copy or CI change.

Root old-code77137 exits Make2 at the ordering assertion. Initial corrected
focused48717 passes; first full71164/vet and race34197 pass before the added
canceled-ready finding. Final cold owners/vet63140 terminal0:operations1.713s,
command5.387s. Final focused race53698 terminal0:all nine receiver parents and
their negative cases pass1.371s, including32 canceled-ready schedules. Ordinary
opt-in service skips are not integration evidence; actual restart-profile and
designated independent union review remain pending. Gofmt/diff checks pass.
No storage/full-profile rerun is inferred; existing e1aa storage capture above
keeps its source identity. Complete follow-up range is e1aa4c50-exclusive through
this containing source/tests/docs/ledger/plan/log commit. Review by: Codex root,
implementer first filter only. Recorded by: Codex root. No self-approval, original
AC7/feature acceptance, R-006, full CI, archival, push, deployment or release.

## 2026-11-04 — Corrected receiver union independently approved

Review by: independent Codex `/root/operations_writer_receiver_corrective_review`.
Recorded by: Codex root. APPROVED, no remaining findings, complete original-plus-
follow-up contexts (base exclusive, tip inclusive), all12/6 paths inspected:

- `cdc780502bd37d09be14adf5f3a554cc27afcb41..c4279b6621d54e3b65a91aa337db0da46e396221`.
- `e1aa4c507e659ca6e204e1103399e4fdbdbbdf7a..b0fd04adfbbde860969e07e49063f01fd36d8a39`.

Original CHANGES REQUIRED and failures4938/5528 remain intact; correction does
not retrospectively approve the original helper. Actual HTTP handler preserves
real FIFO while preflight reads only structured controls. Cancellation cannot
accept an already-ready control. Cleanup mirror/rejection/body/overflow guards stay.

Fresh atb0fd04ad: cold78352 operations1.522s/command5.373s and vet pass; race62219
all nine receiver parents/negative cases plus32 canceled-ready schedules pass1.239s.
Actual declared restart44339 terminal0: canceled-CLI cleanup negative0.88s;
complete journey136.90s, child87.56s, package137.936s. Real automatic count3 and
inspected birth reach production observer/private services/new receiver; actual
replacement count0 delivers resolution of the SAME alert. Exact-range whitespace
passes. No extra old-binary restart or full operations profile is represented.

Live pinned Alertmanager config confirms five-minute resolve timeout, alertname
grouping, zero group wait, one-minute grouping interval, four-hour repeat and
resolved delivery. VerifyAlertDelivery posts one nonce. Room for its consumed
firing plus one later resolution is a CONFIGURATION/SOURCE INFERENCE, not observed
natural expiry or duplicate/retry tolerance. Extra pending controls fail loud.
RP-521 storage44377 retains pre-fix e1aa source; not repeated/relabelled as this
new receiver. This review closes corrective/tool scopes, not full AC7, actual
gameserver/journald retention, R-006, hosted/full CI, archival or release.

All handles terminal. Exact fixture-container and both unique-project container/
network inventories empty; only two pre-existing BuildKit remain. No tracked
reviewer edits/commits. Root reconciles statuses with this verdict; next remaining
scheduled-backup/ingress proof review. No suite for editorial records or publication;
full nine-tier1.0 goal remains active.

## 2026-11-04 — Scheduled backup and ingress proof ranges independently approved

Review by: independent Codex `/root/scheduled_backup_ingress_proof_corrective_review`.
Recorded by: Codex root. APPROVED, no findings, both complete ranges (base
exclusive, tip inclusive), all source/tests/Make/Compose/docs/original records:

- Scheduled backup, twelve paths:
  `c4279b6621d54e3b65a91aa337db0da46e396221..12399dc9248bffdf53c055c2da314c1b543848ff`.
- Ingress outage, eight paths:
  `12399dc9248bffdf53c055c2da314c1b543848ff..9b48c9c603ae66d84deb05351c017e4e02342ac0`.

Fresh operations sourcea6fa32ac, affected children/supervisors byte-identical to
reviewed tips; later receiver correction has its separate approval above. Backup
20190 terminal0: child369.00s/parent372.13s/package372.264s. Actual encrypted
Postgres baseline, scheduled target refusal and receiver firing preserve both
baseline and owned obstruction; same worker waits unchanged five-minute retry,
matching alert resolves, fresh database restores updated account row `after-retry`.
Initial7441 Docker denial ran no private population and supplies no pass evidence.

Ingress6671 terminal0: child394.60s/parent397.88s/package398.132s. Owned Caddy
actually stops; route/scrape fail while upstream controls stay healthy. Independent
inspection confirms Exited0 and live Prometheus PENDING. Existing4m45 quiet and
five-minute lower bound pass, followed by receiver firing, same-container restart
and matching resolution. No injected ingress metric or shortened interval.

Cold four owners/vet37161 pass; native opt-in skips are not integration credit.
Population guard73338 passes healthy and refuses empty/skipped/wrong/log-only
outputs. Private actual pending-observer27451 exits expected Make2: healthy passes;
six malformed/HTTP/query/premature-firing/zero/stale-start controls fail through
own diagnostics. Both exact-range whitespace checks pass. Caddy and Prometheus
warn amd64-on-ARM: mixed-architecture local fixtures, not supported clean-host proof.

All handles terminal; both owned projects' containers/networks/volumes are absent,
only two pre-existing BuildKit containers remain. Reviewer has no tracked edits or
commits. Concurrent CI changes touch no affected operations/server/Make/Compose
bytes. Private controls remain `/private/tmp/cloud-clicker-backup-ingress-review.Gx6s4V`.
Synthetic release metadata and selected-row restore do not establish full semantic
recovery, off-host durability, six-hour scheduling, production RPO/RTO, actual
gameserver readiness, host retention, R-006, full AC7/CI or release approval.
Root updates current plans only; no suite for editorial records, no publication
or feature acceptance. Continue remaining accepted operations construction; the
full nine-tier1.0 goal remains active.

## 2026-11-04 — Bind journal workload collection to its directory and complete output

DP7 requires measured journal capacity and honest incomplete-collection refusal.
`journal-observe` previously used its supplied directory only for filesystem size,
while journalctl read the default host journal. Exit-zero stderr diagnostics were
discarded. It now supplies that exact directory, requests all JSON fields with
quiet/no-pager output, and refuses diagnostics before publishing any observation.
Diagnostic text is not exposed. Retention, capacity calculation, schema, service
filters, runtime logging and alert policy are unchanged.

Actual command-path regression17910 exits expected Make2 on old production:
the healthy control lacks directory/all/quiet arguments, and the warning control
emits468 bytes of supposedly complete evidence. Failed, truncated and empty
controls already refuse. Corrected42655 passes all five arms. The fixture launches
a real controlled child through PATH and checks actual arguments, exit status,
stderr and observation bytes; it is not systemd/journald evidence.
Cold60592 passes operations1.631s and command5.626s; selected vet passes. Focused
race23999 passes5.668s. Existing opt-in skips supply no integration credit.
Gofmt/diff checks pass. No SQL or browser boundary changes; no unrelated full
operations rerun. Complete six-path corrective range starts afterb6744226 through
this containing code/test/docs/ledger/plan/log commit; independent review remains.
Review by: Codex root, implementer first filter only. Recorded by: Codex root.
Actual host collection, retention/eviction, workload ratification, full AC7 and
exact-release R-006 remain open. No self-approval, archival or publication.

## 2026-11-04 — Journal collection review finds suppressed access warnings

Review by: independent Codex `/root/scheduled_backup_ingress_proof_corrective_review`.
Recorded by: Codex root. CHANGES REQUIRED for complete six-path
`b6744226..c09bb854` (base exclusive, tip inclusive). Original five-arm52675 passes,
but private flag-aware67925 fails: child suppresses its access warning iff --quiet,
and current command emits468 observation bytes with nil error. Exact old-main
overlay46386 separately reproduces missing directory/all/quiet and warning
false-complete. Root's initial green fixture did not model warning suppression.

Primary journalctl documentation explicitly identifies inaccessible-journal warning
suppression by --quiet:
https://www.freedesktop.org/software/systemd/man/latest/journalctl.html
Systemd v257 acquire_journal passes arg_quiet with directory selection, and
journal_access_check_and_warn skips the EACCES notice while readable files allow
success: src/journal/journalctl-util.c and src/shared/journal-util.c at
https://github.com/systemd/systemd/tree/v257/src . This is source/controlled-child
evidence, not an actual unreadable host-journal experiment.

Actual declared Playwright Noble image, systemd255.4-1ubuntu8.16: owned networkless,
read-only container with empty tmpfs journal directory executes JSON/all/no-pager
with and without quiet. Both exit0, stdout0/stderr0; no non-JSON no-entry marker.
The --rm container finishes. Cached Go image lacks journalctl (exit127), no journal
credit. This supports preserving healthy empty JSON cleanup without quiet only
for the observed image/version; it proves neither access refusal nor retention.

## 2026-11-04 — Preserve access warnings in both journal collectors (first filter)

Remove --quiet from workload collection and existing cleanup collection. Existing
diagnostic refusal now sees access warnings. Flag-aware real-child tests propagate
the actual arguments; they do not merely assert a changed flag. The cleanup
fixture preserves full row parsing and empty/no-failure collection. No retention,
service-filter, capacity arithmetic, runtime logger, timer or alert policy change.

Root51545 exits expected Make2 before correction: workload access-warning emits468
bytes and cleanup access-warning returns a cursor with nil error. Ten other arms
pass. Corrected2070 passes all twelve arms; cold59660 operations1.810s/command5.742s
and selected vet pass. Focused race9241 passes12.048s. Gofmt/diff pass; no unrelated
SQL/browser or long-profile repeat. Complete seven-path follow-up starts after
`c463a5bb` through this containing main/two tests/docs/ledger/plan/log commit.
Review by: Codex root, implementer first filter only. Recorded by: Codex root.
Original CHANGES REQUIRED remains; final independent original-plus-follow-up
review pending. Actual release-host collection/retention/eviction, full AC7,
exact-artifact R-006, full CI and release stay open. No archival or publication.

## 2026-11-04 — Complete journal collection corrective union independently approved

Review by: independent Codex `/root/scheduled_backup_ingress_proof_corrective_review`.
Recorded by: Codex root. APPROVED, no remaining findings, complete contexts:

- Original six paths: `b6744226ef263cfe031e5dc4daa79ac171701bec..c09bb854a02729af24f4dd5c4c0fb29025d5a484`.
- Follow-up seven paths: `c463a5bb42f068139ff27c47f4aa94b92966b14f..e3ed44678a8fa559f4b6096d939d92616fcdf60c`.

Original CHANGES REQUIRED remains. Both commands omit warning suppression while
retaining exact directory/service selection, complete JSON, diagnostic refusal
and healthy empty cleanup. Fresh exact-old main overlay94454 exits expected Make2:
only two access-warning arms fail (468 workload bytes/healthy cleanup cursor),
ten companion controls pass. Current native34804 passes all twelve,0.649s;
declared Linux-amd64 service41984 passes all twelve,0.974s. These are actual
controlled child processes, not actual journald. Cold42632 operations1.717s,
command5.760s and vet pass; gofmt/both range-whitespace checks pass.

Earlier actual systemd255 empty-directory JSON observation is explicitly REUSED
for unchanged flags: zero output/diagnostics, exit0 without quiet, observed image
and version only. No unreadable-journal or retention claim. All handles terminal,
no tracked reviewer edits/commits. Only newly created inspected tmpfs Postgres
16cd8619 is removed. Network2d8ccc54 gains a peer browser between empty inspection
and removal; Docker refuses removal, reviewer preserves the peer/network and
coordinates with the browser owner. Unrelated parent010be676 changes no affected
source/Make/Compose bytes. Root updates plan/defects; no suite for editorial records.
Host-journal completeness, workload/capacity ratification, retention/eviction,
full AC7, R-006, hosted/full CI and release remain open. No archival/publication.

## 2026-11-04 — Actual gameserver outage and private alert recovery (first filter)

Accepted DP7/AC7 still lacks actual game-process producer evidence: the approved
proxy population controls its upstream response/registry. RP-721 adds a separate
named manual population, reusing the private stack, receiver and owned-project
cleanup. It launches the current compiled gameserver command with a closed set
of disposable development credentials, real migrated Postgres and read-only
checkout content. The controlled metrics/HTTP upstream is not started. Real204
readiness, private scrape and ready gauge establish baseline; SIGTERM must finish
with successful process exit, while Caddy remains reachable and returns502.

Actual gameserver pending timestamp, unchanged4m45 quiet/five-minute firing bound,
receiver delivery, new executable startup and exact original-alert resolution
must all complete. Shared pending parsing retains the Caddy wrapper, binds either
fixed producer and refuses incomplete/trailing JSON, HTTP/query/malformed/stale/
premature-firing observations. Both producer controls execute, including wrong/
missing producer cancellation. Named-pass guard refuses empty/skipped/wrong/log-only
output. No production/rule/interval/kernel/version/content/copy/workflow changes.
The existing twelve-minute observation guard/fourteen-minute cleanup timeout
remain; this is outside push CI and cannot claim all of AC7.

Cold31718 passes before observer controls;78089 operations1.672s/command5.706s
and selected vet pass. Focused race40332 passes both producers/negative controls,
1.350s. Static CI topology/diff checks pass. Initial16346 is Docker access denied,
no private tests. Authorized manual71301 terminal0: child392.94s/parent395.93s/
package396.085s, unique project
`cloud-clicker-gameserver-profile-j7ljhycltdx22yzdhc3pcxxdev`. Source uses actual204
readiness. External private inspection confirms stopped game process (only
operations.test remains), real pending alert jobgameserver activeAt
2026-11-04T01:48:40.964276279Z, and new gameserver PID75014 after restart. Receiver
firing and same-alert resolution complete. Caddy/Prometheus warn amd64-on-ARM:
mixed-architecture local fixture, not supported clean-host evidence. All owned
containers/network/volume are removed; direct exact-project inventories are empty.

Independent PRE-INSPECTION finds a matching-response cancellation gap in shared
pending parsing, also present in the extracted Caddy logic: a transport cancels
the request and still returns matching JSON, which succeeds before the next loop.
Root13815 fails both producer controls with nonzero timestamp/nil error. Check
ctx.Err before returning the timestamp. Final cold56458 operations1.700s/command
5.790s and vet pass; final focused race32716 passes1.362s, all22 controls across
both producers. No binary/source mutation occurred during71301; its real-service
pass PRECEDES this cancellation refinement and a Logf-only cleanup. Do not cite
it as execution of the final revision. Final-source actual service run remains
part of designated independent review, rather than an unchanged reassurance rerun.

Review by: Codex root, implementer first filter only. Recorded by: Codex root.
Complete eleven-path batch starts aftera636b315 through this containing Make/
Compose/three operations tests/shared fixture/parent test/docs/ledger/plan/log
commit. Independent exact-range review remains; pre-inspection is not approval.
Browser/auth/key continuity, automatic restarts, host journald/retention,
full AC7 and exact-artifact clean-host R-006 remain separate obligations.

## 2026-11-04 — Actual gameserver availability proof independently approved

Review by: independent Codex `/root/scheduled_backup_ingress_proof_corrective_review`.
Recorded by: Codex root. APPROVED, complete eleven-path range, base exclusive/tip
inclusive:
`a636b3155de6bcab81fe1340903782f9bb709486..b23a139b24a9123c21d69c299bf7897d4a18c1d3`.
All source/tests/Make/Compose/docs/backlog/plan/log inspected. Pre-inspection's
canceled matching-response issue is corrected in the exact reviewed revision.

Fresh final-source23396 terminal0, actual root manual target: child397.21s,
controller400.91s, package401.085s. Actual gameserver204 readiness→SIGTERM/clean
exit→Caddy502/up1 and game scrape0→real pending/unchanged five-minute rule→receiver
firing→new executable readiness→matching original-alert resolution all execute.
Independent live inspection confirms stopped process and actual pending series.
Root71301's earlier real-service pass remains separately attributed, not restamped.

Cold98627 operations1.714s/command5.809s and vet pass. Focused6697 passes all22
observer controls plus actual named-population guard. Private overlay6708 removes
only the ctx guard: both canceled matching-response cases fail,20 companion
controls stay green; no main-tree mutation. Gofmt/full-range whitespace pass.
Runtime, production rules/thresholds/intervals and CI jobs remain unchanged.

All reviewer handles terminal; direct unique-projectq5ryhjuzyh4i55kw2tp5yjvpp4
container/network/volume inventories empty. No tracked reviewer edits or commits.
Cached Caddy/Prometheus are amd64 on ARM; this is the declared mixed-architecture
private executable population, not production configuration, browser/persistence
or key continuity, automatic restart, host journal/retention, full AC7, clean-host
R-006 or release acceptance. Root reconciles current scopes only; no software
suite for editorial records, archival, publication or goal pause.

## 2026-11-04 — Backup protection amendment technically approved, defect remains

Review by: independent Codex `/root/scheduled_backup_ingress_proof_corrective_review`.
Recorded by: Codex root. TECHNICAL REVIEW APPROVED, complete seven-path range:
`09cf45e5539437b581b9318443632a23a995977e..ce61b82a5a87f3c6f4989f05b17a369396b0b652`.
Includes original diagnostic and complete BM1–BM5 recommendation, test/docs/
backlog/plan/log/index, not just the RFC. Current Create/ReadHeader/Restore/
ExtractPackage/retention/target locks/schedule/Compose/release readers inspected.

No technical blockers: external trust/dedicated Ed25519 seed, exact canonical
signed bytes, outer-v2/authenticated-inner-v1 mapping, target-owned full-census
cleanup, verified publication and authenticated legacy copy are coherent.
Originals/protection remain immutable; governed resolution is still a separate
author hold. Official standard-library API supports the proposed signature use,
not completion of an integration that is not implemented.

Fresh native63927 terminal0,0.318s, and declared Linux-amd6414146 terminal0,0.184s,
both execute both flag mutations and explicitly report protection failures2/2.
Current classification accepts and cleanup deletes each protected synthetic
fixture; authenticated restore refuses before output. Actual age/filesystem,
synthetic dump bytes: NOT SQL restore or retention acceptance. Formatting and
complete-range whitespace pass. All handles terminal; only own created Postgres
987ed25fc175 and confirmed-empty network removed, caches preserved. No reviewer
tracked edits/commits. Root reconciles current records without a software suite.

Ready for owner adoption, not adopted: no signed format/new seed/trust/operator
configuration/legacy conversion or production repair is authorized by this
verdict. RP-505 remains OPEN; D-019, governed resolution, snapshot-based RPO,
exact artifacts and clean-host R-006/full Deployment/release stay separate.

## 2026-11-04 — Bind migration validation and pg_dump to one snapshot (first filter)

Accepted DP6/AC4 requires a consistent database interpreted by its release manifest.
RP-724: CreatePostgresBackup closes its migration-check connection before pg_dump
opens its own snapshot. Actual new regression14205 at0179415e completes a backup
after another transaction commits real DDL, next Goose version and an account-row
update; actual encrypted RestorePostgresBackup then refuses declared migration.
All eight previous integration functions pass in that same red target. This is a
production false-completion defect, not merely two standalone census timestamps.

Check the migration inside a read-only repeatable-read transaction, export its
snapshot to pg_dump16 and hold it until the real dump succeeds. Commit before
packaging; all error paths roll back. Restore's existing check stays unchanged.
No package/envelope/API/key/retention/formula/version/migration-file changes.
PostgreSQL16's documented snapshot export/import supplies the mechanism; it is
not a new application schema or owner policy. Separate recovery-census/dump
binding, snapshot timestamp/RPO, concurrent-DDL refusal and RP-505 stay distinct.

First corrected63193 proves successful actual encrypted restore of the original
complete row/migration identity, but its new cancellation cleanup assertion fails:
NOWAIT incorrectly assumes server lock release precedes a cancellation return.
Existing RP-490 witness and Go database/sql awaitDone/rollback confirm asynchronous
teardown; no production leak is inferred. Follow that existing witness: observe
actual exclusive-lock acquisition under the same30-second completion guard.
This does not add a sleep, raise a bound or redefine restore success.

Final actual declared Linux-amd64/Postgres84225 terminal0 executes all nine
integration functions without skips. New witness3.03s restores pre-migration schema,
version and complete recovery identity after concurrent committed changes; initial
mismatch refuses before dump, ordinary dump failure/cancellation publish nothing,
and actual exclusive-lock acquisition proves exporter release. Existing empty/
populated/query-override, corruption/clean-target, snapshot and deletion/restore
populations also execute. Target removes owned Postgres and its network.
Cold four owners23631 pass; selected vet passes. Native prerequisite skips are
not SQL evidence. Gofmt/diff checks pass; no unrelated browser/full-CI rerun.

Complete six-path batch starts after0179415e through this containing production/
integration test/docs/backlog/plan/log commit. Review by: Codex root, implementer
first filter only. Recorded by: Codex root. Independent corrective review remains;
no self-approval, original R6/AC4, real-release/clean-host or1.0 acceptance,
archival, publication or goal pause.

## 2026-11-04 — Deliver the concrete metadata adoption question

After complete independent technical approval above, root sends ONE async question
linking BM1–BM5: adopt the reviewed signed-metadata/dedicated-seed/external-trust/
authenticated-legacy-copy repair, or keep draft. Tool accepts delivery; no answer
is recorded and preselection is not approval. AGENTS reserves specification
adoption; technical review alone cannot authorize the new format/key contract.
No operator key generated/accessed, backup converted, signing code implemented,
retention duration changed or original discarded. Current RP-724 correction uses
the already accepted DP6 consistency contract independently of this pending choice.
Root reconciles plan/backlog/owner routing only, no software suite for records.
Do not repeat the question; full nine-tier1.0 and all separate release gates remain.

## 2026-11-04 — Migration/dump snapshot correction independently approved

Review by: independent Codex `/root/scheduled_backup_ingress_proof_corrective_review`.
Recorded by: Codex root. APPROVED, complete six-path range, base exclusive/tip
inclusive:
`0179415e5d221dc96aec6febcc0301ed058489b1..9cf66afbe14943097fcc3cd6cd7d51406d1e1284`.
All production/test/docs/backlog/plan/log inspected; no blocking findings. Checked
migration and dump share an exported repeatable-read snapshot held until dump
completion, consistent with PostgreSQL16's documented API.

Fresh independent24259 terminal0 executes all nine real Linux-amd64/Postgres
integration functions without skips. New witness2.01s restores pre-change schema/
version/complete row identity despite a concurrent committed change; initial
mismatch, injected dump error/cancellation and observed actual lock release pass.
Cancellation is AT THE DUMP DISPATCH boundary, not proof of cancellation of an
already-running pg_dump. Four affected backup/release owners39308 pass cold plus
vet. Private exact0179415e production overlay with current regression7267 exits
expected1,1.66s: backup reports completion after DDL/version/row commit, actual
restore refuses its declared migration. No main mutation or false compile failure.
Formatting/full-range whitespace pass. The invalid NOWAIT oracle remains on record;
observing asynchronous server lock release is the justified correction.

Both reviewer-owned Compose projects have no remaining containers/networks/volumes;
no tracked reviewer edits/commits. Parent's editorial metadata-question record
changes no affected source. Additional ROOT caller evidence91910 at2ea094d2:
actual declared Caddy/Postgres release target passes2.54s, including admitted work,
marked drain, real HTTP/WebSocket recovery and encrypted restore. Target removes
Caddy/Postgres/network; direct exact-project container/network inventories are
empty. This caller execution is not relabelled as independent or exact-artifact
clean-host proof. Underlying composition source matches reviewed correction.

RP-724 corrective defect closes, not RP-505, separate recovery-census/archive
binding, snapshot-based RPO, full AC4/Deployment, R-006 or release. No operator
key/configuration change, archival, publication or goal pause. Root reconciles
current states and the bounded whole-product checkpoint; no suite for records.

## 2026-11-04 — Actual native journal files reach both collectors (RP-725)

Outcome: fill DP7's file/reader compatibility gap without treating controlled
journalctl output as an actual Linux journal. New manual root target
`make test-deployment-journal` builds the current command's Linux-amd64 test
executable and owns one labelled, networkless, read-only Noble container using
the existing declared Playwright1.62 image. Its private `/etc` anonymous volume,
runtime/tmpfs journal and test-binary-only read-only checkout mount touch no host
journal. Real systemd255 journald receives logger native messages, syncs and
flushes actual journal files; unchanged production commands invoke actual
journalctl against those files. Exact ownership is checked before force-removing
the fixture and its anonymous volume. Native ordinary package runs explicitly
skip the opt-in parent/child; those skips receive no integration credit.

Coverage: healthy empty cleanup input, all seven identifiers plus one actual
cleanup-failure message (eight selected rows), unrelated service's malformed
message excluded, exact SHA256 of the actual journal cursor, requested empty
directory isolation, missing-directory refusal and an actual unknown-job message
refused. Workload errors emit no observation. A derived reserve uses the actual
short synthetic population's byte rate; the private64GiB tmpfs ceiling allocates
no such backing memory and is not a release-host budget. Collection crosses the
real whole-second query cutoff before reading persisted microsecond records.
No source mutation is needed: actual wrong-directory, unrelated/malformed and
unknown-job populations discriminate the changed test/tool boundary.

Execution by: Codex root, base351124e4 plus this complete six-path batch.
Exploratory networkless native writer/read probe exits0 with one real cleanup
record and string cursor/timestamp. A prior compacted output was lost; exact
image inventory showed no remaining fixture before repeating it. First Make
selector30510 exits0 with **no tests run** due recursive Make consuming an anchor;
corrected to the existing unanchored unique-name convention, no credit for that
attempt.47731 then fails at Docker socket permission before fixture creation;
normal narrow development escalation runs92445 to terminal0. Actual child1.09s,
native owner1.74s, package1.904s; both named PASS lines confirmed. Cleanup's exact
label inventory is empty. Cold affected packages61611 terminal0: operations1.641s,
command5.674s; selected vet exits0. These ordinary package runs are separate from
the actual manual integration, not falsely credited for skipped fixture tests.
Formatting and diff whitespace pass. Final change after execution is test log
wording only; no assertion/runtime change.

Review by: Codex root, implementer first filter only. Designated independent
review requested for complete six paths after351124e4 through this commit:
Makefile, new journal integration test, canonical deployment docs, backlog,
owning plan and this log. No workflow/production/rule/retention change. Host
Docker journald-driver delivery, 14-day retention/purge, supported-host workload
sizing, exact release artifacts, full AC7/R-006 and release approval remain open.
No unrelated service population rerun or broader milestone promotion.

## 2026-11-04 — Native journal compatibility independently approved (RP-725)

Review by: fresh independent Codex `/root/journal_file_compatibility_corrective_review`.
Recorded by: Codex root, transcribing its executed verdict. APPROVED, no findings,
complete six-path range (base exclusive, tip inclusive):
`351124e48ce2a5c0cce5dc73efe76c9f5b43e394..927b26ed02537dda486edf45702bcd491a166954`.
All test/tool/docs/backlog/plan/log inspected. Existing ordinary invalid populations
discriminate service filtering, exact directory selection and unknown-job refusal;
actual native cursor identity and named population checks pass. No source mutation
is needed or claimed. Initial no-tests and sandbox-denied invocations remain
non-evidence; no historical failure is restamped green.

Fresh independent63894 terminal0: actual root manual target, native child0.61s,
owner1.12s, package1.273s, both named populations explicitly PASS. Cold affected
packages97957 terminal0: operations1.612s/command5.755s. Selected vet, formatting
and complete-range whitespace pass. Final escalated exact fixture-label inventory
is empty; removal includes anonymous volume. No tracked reviewer edits/commits or
live handles. Cached Noble/systemd image is Linux/arm64, while collector executable
is Linux/amd64: this mixed-architecture Desktop fixture is explicitly local.

RP-725's file/reader evidence gap closes. Host Docker journald-driver delivery,
release-workload capacity measurement, actual14-day availability/purge,
supported-host installation, full AC7/R-006 and release gates remain open.
Root reconciles only owning plan/ledger and this verdict; no suite for records,
runtime change, full feature promotion, archival, publication or goal pause.

## 2026-11-04 — Rebuild the current candidate in two independent artifact trees

DP1/AC1/AC8 local construction updates the retained candidate after the reviewed
runtime corrections. Source remains clean, exact
`df84ce1a86780135633b1d388541a1a5016bdf1d`, commit epoch1793760497,
creation2026-11-04T02:48:17Z, existing LOCAL label0.1.0-preview.2. This label is not
a publication or release approval. Both trees rebuild the client, six reproducible
static Linux/amd64 helpers, epoch8 content and53-dependency application metadata;
both no-cache image builds use the unchanged pinned BuildKit/frontend recipe and
identical local tag. Fourteen actual platform-selected Syft scans execute, not old
SBOM copies. Registry index→uniqueamd64 child→config digest and actual Linux/amd64
configuration are checked afresh for all six upstream/rehearsal images. Actual
development engine28.4.0/ARM64 and Compose2.39.4-desktop.1 are inspected; initial
sandbox-denied engine inspection provides no server observation, corrected narrow
read supplies it. No unsupported native Linux/amd64 host claim.

Retained local artifacts:

- `.cache/release-current-df84ce1a/bundle-a` and
  `.cache/release-rebuild-df84ce1a/bundle-b`, each74 manifest-bound artifacts;
- manifest `sha256:6bf4d32e022942fc88c327565a648cf3a350adb73f18ee8311bbe2e9db3f9b7c`;
- gameserver archive `sha256:ffb767c584207185a3dbbf60c5d88ea4c0ee4be95f2faa635ffa346071d46334`;
- gameserver config `sha256:22c923e220aa6f1a004db0f65617706791483f34eb6dbbdbb54ebf5d5ce04d54`.

Root Make binary/content/image/metadata/SBOM/assembly commands all complete.
First commands86750/36092/46110, second57196/5267, images70846/54892; all handles
terminal0. Whole bundled trees `diff -r` exits0. All six separate binaries and
archives, application metadata and seven normalized SPDX documents also compare
byte-identical; raw Syft headers are not reproducible outputs or shipped. Both
assemblers execute their actual bundle validator. Build record is derived from
these actual file hashes/manifest identities, not copied old digests; native current
`deployment-rehearsal validate-build` accepts the new retained record. Original
102face6 record/trees and inadmissible historical predecessor remain untouched.

Two setup mistakes explicitly receive no evidence credit. Second metadata run
was dispatched before its live client-build handle was confirmed terminal;
`.cache/release-rebuild-df84ce1a/metadata` is excluded. After actual client completion,
fresh `metadata-complete` is generated and used in bundle-b. First healthy probe
124c8f exits2/error_class=invalid_probe because its required work directory was
absent; mkdir followed by the exact same selector exits0. This is setup correction,
not a changed product assertion or rejection proof. That healthy population is
`six_image_sbom_license_provenance` against the independent candidate rebuild pair,
NOT the immediately previous release. All nine subsequent current-native-binary
probes exit exact3/fixture_rejected: removed catalog/client/license/config/helper,
changed image/runtime config/SBOM, wrong epoch/artifact set. Each validates its
healthy hardlinked fixture first, then semantically rebounds hashes before testing;
post-probe full equality holds and work directory is empty.

Structured `make release-secret-scan`92573 terminal0 binds the above exact manifest
and archive:2464 tracked files, image_archive_scanned=true, findings0,
objective_completed=true, guard_exhausted=false. No build/source input changed;
the retained record is added only afterward. All disposable scanner containers
are absent by exact Syft-ancestor inventory; pre-existing shared builders/caches
are retained. No actual release service is started, secrets provisioned or host
configuration changed. This produces current artifact evidence, not hosted green,
runtime gameplay/rollback, floor-compliant previous release, full DP-B/AC8,
supported-host R-006, original-feature approval or the full nine-tier1.0 result.

Review by: Codex root, output/record first filter only. Recorded by: Codex root.
Complete four-path evidence range afterdf84ce1a through this commit: new retained
record, canonical docs, owning plan/log. Needs independent evidence review, no
self-approval or archival. Continue real predecessor/authority/clean-host and
accepted gameplay work; pending owner adoption questions are not answered by build.

## 2026-11-04 — Current candidate build evidence independently approved

Review by: fresh independent Codex `/root/candidate_rebuild_evidence_corrective_review`.
Recorded by: Codex root. APPROVED, no findings, complete four-path range (base
exclusive/tip inclusive):
`df84ce1a86780135633b1d388541a1a5016bdf1d..593e54b4da2f4afdd6189ddac73b6a3e64854715`.
Retained record, canonical docs, owning plan/log all inspected. Both trees match
all75 files (74 manifest-bound artifacts plus manifest), including actual record
hashes/source timestamp/archive configuration/seven normalized inventories.
Fresh independent registry reads confirm all six upstream/rehearsal Linux/amd64
configuration identities. Current validator and healthy package probe pass;
all nine actual tampering probes exit exact3/fixture_rejected. Independent current
source six-binary rebuild83170 terminal0 matches retained bytes; private scratch
removed. Post-probe full equality holds and work directory is empty.

Root's fourteen actual scans are inspected, NOT independently rerun: distinct raw
scan times and matching package populations. Inspected secret-scan output binds
the exact source/manifest/archive,2464files/zero findings, no guard exhaustion.
Both initial setup mistakes remain excluded. All review handles terminal; no
tracked reviewer edits/commits. This is build-evidence approval, not a candidate/
previous pair, full Deployment/hosted CI/supported-host R-006 or release readiness.
Root reconciles owning plan and bounded whole-product artifact checkpoint only;
no new suite for records or full-feature/archival promotion. Continue the full
nine-tier1.0 scope and all separate source/product/operator/acceptance gates.

## 2026-11-04 — Predeclare restore-client/database quiescence experiment

Question for draft CR4: does exact client-container removal fence an already
admitted real pg_restore database mutation, and does separately terminating the
identified restore backend change that outcome? Current docker.go sends the
restore through a Compose one-off; deleting the client is a proposed containment
step, not an adopted implementation. Earlier local-process/ExecID observations
cannot answer the database boundary. Actual inspected pair: Engine28.4.0 ARM64
Desktop daemon, Compose2.39.4-desktop.1, Linux/amd64 pinned Postgres16 image;
this is a local precursor, not the supported native host.

Three arms on separate disposable target databases: normal completion; exact
owned restore-client removal while its CREATE TABLE waits on a held pg_class
lock; the same removal plus termination of its identified database backend before
unlock. Use actual custom-format pg_dump and the runtime's pg_restore flags,
not a mocked runner, an altered production helper or a pre-existing target table.
Observe exact backend/application/database/query and lock wait before intervention,
then release the real database lock. The normal control must create its table and
row. Record whether each interrupted arm's table exists, not an expected success
chosen in advance. If the target query never reaches CREATE TABLE/lock wait, or
any handshake10s/whole-experiment1m guard fires, the observation is INVALID.
No sleeping absence test or raised guard may replace the backend/command outcomes.

Only the unique labelled local Compose project/client IDs may be removed. No
official project, provider, secret, production dataset or external host is used.
Results may choose/refuse a CR4 operation-specific containment proposal; they
cannot establish all pending Engine-request/Compose-lifecycle quiescence, a real
backup/rollback rehearsal, new runtime authority or adoption. Retain failures and
controls, review any draft body change independently; full1.0 goal remains active.

### Instrument correction before further measurement

Initial34516 fails before database startup with its command diagnostic hidden.
The diagnostic-only repeat reveals Docker's `cannot overwrite digest` cache
collision: the declared index resolves locally to ARM64. Resolve/verify/pull its
EXACT existing amd64 child075f7ba6/config75f5a969 (no version/image-policy change,
shared ARM image not deleted). Initial catalog-lock run54193 is INVALID at its
unchanged10s handshake; diagnostic repeat34649 also INVALID, last tagged backend
empty. Neither measures an admitted restore statement, so no client/backend
quiescence outcome follows and no time bound is raised.

Revised instrument, predeclared before execution: the actual custom dump contains
a synthetic table CHECK calling a source-defined function that obtains one
transaction-scoped advisory lock. Hold that lock from the target's own separate
fixture session, without adding any target object/trigger or changing pg_restore.
All three arms must reach an observed active COPY/Lock backend in that otherwise
ordinary restore. Normal completion must restore row7; interrupted arms record
actual persisted row count after unlock and backend disappearance. Same10s/1m
guards, three interventions, unique ownership/cleanup and precursor-only limits.
This is controlled database-object instrumentation, not current game content or
the complete encrypted restore helper. A missing COPY wait invalidates this
replacement too; no sleeping absence or altered expected outcome is allowed.

## 2026-11-04 — Actual restore COPY survives client-container removal (RP-726)

Outcome changes CR4's next mechanism requirement. Actual three-arm7871 terminal0
observes normal row1, removed-client row1 and removed-client+backend-termination
row0; backend survives removal in both interrupted arms. Stronger causal
instrument13042 terminal0 additionally requires actual COPY/advisory wait and
zero committed rows BEFORE each intervention, then publishes complete JSON only
after successful cleanup. All three arms reproduce the same results. Normal
control restores exact row7. Both interrupted arms retain the table created
before COPY; session termination is demonstrably not rollback. The lock is opened
only after full client-ID removal, with backend disappearance awaited before the
final row observation. No arbitrary delay supplies the absence claim.

The preserved Python source runs actual declared Compose Postgres, pg_dump and
pg_restore with the runtime's exit-on-error/no-owner/no-privileges flags and
separate disposable target databases. Its synthetic source-defined CHECK function
provides the database gate; no target trigger or production source is installed.
Exact project/service/full-ID and image config ownership is checked before client
removal. Third arm terminates only the observed PID under its unique database/
application predicate, then independently observes its disappearance. This is a
bounded positive containment control, not a proposed production PID/application
authority or all-session/request fence. Source/balance/migrations are untouched.

Initial startup and catalog-lock errors remain in the preceding predeclaration/
instrument-correction entries. No successful outcome is inferred from them and no
10s/1m guard is raised. Final source additionally reports incomplete/guard fields
on exceptions and ensures project cleanup is attempted even if local-client
cleanup fails; these error/cleanup-only refinements follow13042's data run.
Every owned invalid/completed project was removed. Final exact name inventory
has no restore-fence containers; script checks its network (and final source its
volume inventory) before emitting completion. All handles terminal, no official
project/host/database/key operation or Docker cache deletion. No full Go/browser
suite for this manual SQL research; execute its real subject rather than trackers.

Primary PostgreSQL16 docs explain optional connection-loss polling and distinguish
signal delivery from confirmed backend termination; actual observed data, not
those descriptions, establishes this result:
[connection checks](https://www.postgresql.org/docs/16/runtime-config-connection.html#GUC-CLIENT-CONNECTION-CHECK-INTERVAL),
[termination](https://www.postgresql.org/docs/16/functions-admin.html#FUNCTIONS-ADMIN-SIGNAL).
Docker's force removal uses SIGKILL, which is insufficient for work already sent
to another service; its [remove reference](https://docs.docker.com/reference/cli/docker/container/rm/)
does not promise a database fence. The real experiment supplies the cross-service
counterexample. No global Postgres setting, signal privilege or recovery mechanism
is adopted merely because a local fixture can terminate its own backend.

Draft author reconciles CR4/body/open-question/evidence, not just an appended
warning. RP-726 remains OPEN until the complete Engine/client/database lifetime
mechanism is specified, supported-host proved and adopted. Review by: Codex root,
research implementer/draft author first filter only. Recorded by: Codex root.
Exact five-path range after5b9dbb2c through this containing commit needs fresh
independent technical review: draft, new probe source, backlog, owning plan/log.
This is no production correction, release acceptance or narrower1.0 goal.

## 2026-11-04 — Original restore research review: CHANGES REQUIRED

Review by: independent Codex `/root/restore_backend_fence_research_review`.
Recorded by: Codex root. Complete five-path range:
`5b9dbb2c8409111984bc68f334b61dff0c0b9062..fad986789aa475ebda35e10e791854e620d785d2`.
CHANGES REQUIRED. Fresh actual30254 terminal0 independently confirms exact
COPY/advisory waits, zero baselines, normal/client-removed row1 and identified
backend-terminated row0. Earlier DDL remains in all arms; no full recovery fence.
Exact owned project container/network/volume inventories empty; no reviewer edits.

RP-727: exact AST-extracted56d223 accepts a matching check returning at11 with
deadline10, because the local wait tests truth before its guard. Fast actual
result is not invalidated, but the tool's promised guard must discriminate.
RP-728: collection catches every RuntimeError as guard exhaustion, including
ordinary Docker failure; the outer report can then contradict that classification.
Preserve this verdict and forward-correct; no guard increase or adoption.

## 2026-11-04 — Forward-correct the research guard and failure class

The actual wait now checks its unchanged deadline before invoking a check and
again before accepting a matching result. Explicit ResearchGuard distinguishes
exhaustion from ordinary command failure; the collection catch handles only that
class. Retained `python3 planning/deployment-foundation/probes/restore-backend-fence-probe.py.txt --guard-controls`
executes two timely matches, exact10/late11/whole-deadline5→6 refusals and a genuine
ordinary Docker error through MAIN'S ACTUAL AST-extracted collection catch, not
just the wait helper. All six controls pass. Separate exact historical/current
catch execution reproduces old false guard=true and corrected no guard claim;
the same sentinel RuntimeError propagates unchanged. No source mutation or mock
database result is used for the real experiment.

Corrected actual53207 terminal0 repeats all three real SQL arms with zero baselines,
advisory COPY waits and results1/1/0; full-ID removal/backend survival and true
termination match prior observations. Owned project is removed before publishing
completion. Subsequent edits only strengthen the retained failure-class control
using the actual catch AST; real SQL/data assertions are unchanged. Whitespace and
Python execution pass. No unrelated Go/browser tests for this instrument repair.

Review by: Codex root, correction implementer first filter only. Recorded by:
Codex root. Complete original five paths plus this four-path forward range
(basefad98678 exclusive through this commit) needs independent union re-review.
Original negative verdict remains. RP-726's complete mechanism/CR4/adoption and
full product/source/host/release gates stay open; no runtime change or goal pause.

## 2026-11-04 — Corrected restore research union independently approved

Review by: independent Codex `/root/restore_backend_fence_research_review`.
Recorded by: Codex root. APPROVED, no remaining findings, complete union:

- Original five paths:
  `5b9dbb2c8409111984bc68f334b61dff0c0b9062..fad986789aa475ebda35e10e791854e620d785d2`.
- Forward four paths:
  `fad986789aa475ebda35e10e791854e620d785d2..12d935e56ccae386416307a6f06ab4ac790f9e8a`.

Original CHANGES REQUIRED is retained. Fresh independentef6c64 terminal0 executes
all six retained controls: two timely, three boundary/late refusals and ordinary
failure through the actual collection catch. Deadline checks precede acceptance;
ordinary errors no longer claim guard exhaustion. Fresh actual62105 terminal0
executes FINAL-source real SQL fixture: all zero-row/COPY/advisory baselines,
normal row1, removed-client row1 with surviving backend, identified termination
true/row0, earlier DDL retained. Complete-range whitespace passes. Exact owned
project container/network/volume inventories empty; all handles terminal, clean
checkout, no reviewer edits or commits.

RP-727/728 tool defects close. RP-726's database-side lifetime requirement and
bounded draft reconciliation are supported, but complete Engine/client/database
fencing remains OPEN. No encrypted game-restore/rollback, native supported-host
proof, full original proposal approval/adoption or release acceptance follows.
Root reconciles only owning plan/ledger and this verdict, no unrelated suite,
runtime change, publication, archival or goal pause. Continue the full1.0 objective.

## 2026-11-04 — Constructive network-generation boundary study (RP-732)

After the Account/Guild corrections, inspect the still-accepted implementation
plans before selecting the next task. Catalog serving, complete P1/Headcount and
Combat catalog/engine paths retain their named technical/content/adoption holds;
do not implement an easier partial endpoint or change authored effects. Continue
CR4's genuine engineering unknown instead: can an immutable network identity
prevent delayed NEW clients from reaching a replacement database network?

Predeclare the full population/60s whole-run/10s per-command/15s cleanup guards
in the owning plan BEFORE84016. Actual supported-for-this-diagnostic population:
Darwin/arm64 CLI, Engine28.4.0 Linux/arm64 Desktop, Compose2.39.4-desktop.1,
exact declared amd64 Postgres16 child075f7ba6/config75f5a969. This is not the
native supported deployment host. Reuse the declared tmpfs backup-test Compose
service in an unpredictable exclusively owned project; no production services,
secrets, published ports, privileges, shared caches or images are modified.

Four ordered worker populations use actual `psql` and one test table:

| Population | Actual result |
|---|---|
| Intact full-ID client | starts/exits0, commits row11 |
| Created/not-started full-old-ID client, started AFTER retirement/recreation | start1, old-network-not-found, no row12 |
| Client created AFTER retirement using full old ID | create0, start1 old-network-not-found, no row13 |
| Name-only client after SAME name recreation | starts/exits0, commits row14 |

Retire only after verifying the owned DB and original project network. Disconnect
the owned DB, remove the old network, confirm absence, create the SAME NAME with
a different full ID, reconnect the SAME DB and then execute delayed clients.
Retirement succeeds despite the not-yet-started worker. Exact final rows11/14
and the name-only control prove that the two missing rows are not merely a dead
database. Creation success alone provides no started/connectivity evidence.
This controlled ordering is not a request delayed inside the Docker daemon or a
helper/CLI death test. RP-726's admitted backend evidence remains separate.

84016 exits0 with those outcomes. Final-source88928 also exits0; the source-bound
`network-generation-observation.v1.json` records all full IDs, create/start/exit
outcomes and errors, original and replacement network IDs and exact rows. It binds
probe SHA256015e28bf396ec9f32a31fa41c1b8c3270f7042d2d6200d025330c34ec05a9b70,
image/config/platform/Engine/Compose and objective=true/guard=false. Final source
separates create/start diagnostics and reports observed row absence rather than
calling a narrow experiment a complete supported mechanism. The original outcome
remains valid; source metadata/cleanup simplification do not invent a second
production experiment. Exact JSON/source/result checks pass. Both owned projects
verify empty container/network/volume inventories before complete output; root
also queries final-project container/network inventories afterward, empty.
No process/test handle remains live. No new bound or ordinary CI lane is added.

Draft CR4 now states the derived conditional: any candidate relying on network
retirement must bind the full immutable ID before submission, rather than resolve
its human-readable name later. Docker's primary disconnect/removal references
support setup, not this runtime inference or a general fence. Current
`DockerRuntime` still uses Compose one-offs/up/down/project selection and fixed
volume names. This experiment does NOT prove that those paths bind the ID, fence
already-admitted SQL, protect host outputs/volumes, or account for all outstanding
Engine/Compose work. Full mutation-path mapping and a combined mechanism remain
required; do not substitute another few positive/negative metadata samples.

Review by: Codex implementer first filter, not independent approval.
Recorded by: Codex root. Complete six-path range afterfb2ae8f4 through this
containing commit requires fresh independent research/tool/technical review.
No runtime behavior, schema, kernel0.3.225, epoch8, balance, deployment/recovery
policy or authored copy changes. No adoption, full CR4/Deployment/1.0 acceptance,
clean-host/R-006 proof, push, archival or host deployment. The goal remains active.

## 2026-11-04 — Interim independent guard-wording finding, RP-732

Review by: fresh independent Codex `/root/network_generation_fence_research_review`.
Recorded by: Codex root, transcribing its interim finding, NOT a final verdict.
The review identifies ambiguity in the plan/log shorthand above: original source
has one absolute60s MEASURED-WORK deadline, ordinary commands at most10s within
that deadline, then cleanup outside it at most15s PER cleanup command. It does
not impose aggregate15s cleanup or60s total wall runtime. Complete observation
still requires successful cleanup. This is the original instrument, not a newly
raised deadline or permission to treat exhausted measured work as completed.

Reconcile the living plan to those exact semantics, while preserving the earlier
log wording and original source/hash/observations. No tool/runtime/data/source
change or rerun is needed for this editorial correction. Independent reviewer is
still executing/checking the source and coherent original-plus-correction union;
no final approval or full CR4 mechanism is inferred from this interim exchange.

## 2026-11-04 — Independent network-generation research review — APPROVED

Review by: independent Codex `/root/network_generation_fence_research_review`.
Recorded by: Codex root, transcribing its returned final verdict.
Complete reviewed union, base exclusive/tip inclusive:

- Six paths: `fb2ae8f4dd463a533a2b94233faebcfa86cb9093..a4ff30ff4dadcd247eb008fd454148c2b4406493`.
- Two-path clarification: `a4ff30ff4dadcd247eb008fd454148c2b4406493..a4959f48ec53510b739ddbe4e5515e8fd972042f`.

Verdict: **APPROVED**, bounded research/tool and draft-reconciliation scope,
no remaining findings. Fresh70664 exits0: exact rows11/14, both old-ID start1/
network-not-found, late create0, intact/name-only exits0. Actual-source AST controls
f435cd pass timely acceptance, boundary/late/pre-start guard refusal and ordinary-
error classification. Retained observation/source/outcome bindingcf71f9 passes;
full-range whitespace passes. Guard wording now matches ORIGINAL measured-work
and per-command cleanup semantics; no source/data/bound change is inferred.

All reviewer handles terminal, private scratch removed, no tracked edits/commits.
Own project`cc-network-generation-27ba938fe187c97b` container/network/volume
inventories are empty. Root projects are already removed as recorded above.
RP-732's bounded evidence is independently supported; the full CR4 mechanism
remains OPEN. Next resolve complete mutation-path ownership and combined Engine/
client/database fencing, not another repetition of the completed worker controls.
No native supported-host/restore acceptance, specification adoption, production
implementation, archival, push or release readiness follows. Goal stays active.

## 2026-11-04 — Whole release/rollback mutation ownership map (RP-733)

Continue from the independently approved network-generation experiment by tracing
the WHOLE existing operation, rather than adding its green result to RP-726 and
calling the pair a recovery fence. Working source is clean ata0dc53bc; current
diff against that coordinate is empty for all audited runtime/helper/template
paths. Retain `interrupted-transition-mutation-map.md` with the exact source
coordinate, both Controller call sequences, every root Runtime call, nested
effects, explicit exclusions and required interrupted-operation disposition.

Source observations that change the next engineering action:

- DrainCurrent and AuthenticatedSmoke invoke HTTP bootstrap, which creates an
  account. A Docker/SQL-only inventory would incorrectly treat smoke as read-only.
- Backup/restore additionally write encrypted backup/operations textfile targets;
  the textfile writer renames over its canonical file. No actual late overwrite
  is asserted: these are source effects that a candidate must cover.
- The production template has edge/application/database/operations networks and
  five persistent volume families. The retired default-network fixture is not
  that whole topology, and ResetDatabase removes a fixed volume name.
- Upstream Composev2.39.4 stop/remove submits selected CONTAINER IDs; network
  removal also uses an inspected ID. Creation's default networking and configured
  volume removal involve names. Thus blanket claims that every Compose request
  late-binds a service name inside the daemon would be false. These primary-source
  observations are not attestation of Desktop2.39.4-desktop.1's binary or runtime
  queued-request behavior. Direct upstream source links are in the map.
- Ledger writes and local decryption scratch have local-helper ownership, while
  daemon jobs and database backends have separate lifetimes. Initial install,
  rehearsal helper entry points and rotation are explicitly outside these two
  Controller sequences; no larger acceptance is inherited from the map.

Checked actual Controller/Runtime/ExecRunner/HTTP-smoke call sites, nested backup
runner/publication/metrics paths and the release template. Relevant source-file
hashes were computed and frozen-coordinate diff checked; the map cites that
Git revision rather than claiming a new runtime population. No Docker job,
database, provider, receiver or secret is accessed by this source audit. No code,
test, schema, kernel0.3.225, epoch8, release artifact or CI byte changes; no product
suite is run solely for prose. Existing bounded runtime findings remain as their
own named evidence, not reruns of this audit.

Draft CR4/Open questions now require a combined mechanism to account for every
in-scope map row, including mutating application HTTP and host writers, and clearly
distinguish the completed inventory from the still-unbuilt protocol. Unknown
ownership/terminal outcome must refuse before recovery effects. Do not add more
root-call censuses or isolated network/backend passes in place of that protocol.

Review by: Codex implementer first filter. Recorded by: Codex root.
Complete five-path range aftera0dc53bc through this containing commit requires
fresh independent research/technical-draft review. Source facts do not adopt a
new runtime mechanism or authorize implementation. Full combined fence, native
supported-host proof, specification adoption, original Deployment acceptance and
R-006/release gates remain OPEN. No live host/deployment, publication, archival,
false status promotion or goal pause. The nine-tier1.0 objective stays active.

## 2026-11-04 — Mutation-map original review: CHANGES REQUIRED; source repair

Review by: independent Codex `/root/transition_mutation_map_technical_review`.
Recorded by: Codex root, transcribing the returned verdict/findings.
Reviewed range: `a0dc53bcc25e16860d5017f7af28be4842afd0d9..12aa01aeee362dfe1c542a50823ad5a720a082d0` (all5paths).
Verdict: **CHANGES REQUIRED** for two completeness omissions, no runtime incident claim:

1. `cmd/deployment-release/main.go`'s OUTER runRelease publishes the outcome metric
   after Controller.Release returns, OUTSIDE its operator-lock lifetime. If an
   operation acquires the lock it is already closed; early validation/admission
   failures may never acquire it. runInstall also writes there; runRollback does
   not have the corresponding wrapper write. Runtime/ledger completion therefore
   does not cover every whole-helper host writer.
2. Start launches backup schedule, whose round calls RemoveStaleTemporaries and
   ApplyRetention, deleting eligible files. Creation/metrics alone omit a possible
   writer class affecting backup/checkpoint material. No actual late deletion is
   asserted; RP-505's separate metadata issue is not resolved or duplicated.

Reviewer independently extracts Controller sequences, checks root/nested source
and all8 original local links, reads fresh upstream2.39.4 source, confirms ID/name
distinctions, and checks source/runtime diff is empty. No software suite, services,
secret/receiver access or tracked reviewer edits for prose; all handles terminal,
no scratch/resources remain. Other original root rows/source claims match.

Correct map/draft/plan/ledger to explicitly cover both omitted effects and preserve
this original negative verdict. This is source-only reconciliation, not a runtime
fix or evidence either race actually occurred. The corrective exact range through
the containing commit needs coherent union re-review. No source/test/CI/kernel/
artifact changes or repeated runtime studies; full mechanism/adoption/release and
the nine-tier1.0 goal remain open/active.

## 2026-11-04 — Corrected mutation-map union independently APPROVED

Review by: independent Codex `/root/transition_mutation_map_technical_review`.
Recorded by: Codex root, transcribing its returned final verdict.
Complete reviewed union, base exclusive/tip inclusive:

- Original5paths: `a0dc53bcc25e16860d5017f7af28be4842afd0d9..12aa01aeee362dfe1c542a50823ad5a720a082d0`.
- Corrective5paths: `12aa01aeee362dfe1c542a50823ad5a720a082d0..8825019c79e1ebe9a1f4bdf296c6ba28ecc704cc`.

Verdict: **APPROVED**, bounded research/technical-draft reconciliation. Original
CHANGES REQUIRED remains. Reviewer verifies both omitted writer classes now
explicit: post-controller CLI metrics outside lock lifetime, and scheduled
temporary/backup deletion. Independent consistencya8dfad exits0: Markdown-only
ranges, unchanged runtime sources, nine valid local links, append-only original
log preservation and coverage of all Controller Runtime calls. Full-range
whitespace passes. Upstream ID/name distinctions were independently checked
against primary source in the original review, not a Desktop-binary/runtime test.

No software suite, services, receiver/secret access or tracked edits/commits;
all handles terminal, no scratch/resources remain, checkout clean. Reviewer asks
one non-blocking editorial clarification, applied in this closeout: `pending
adopted remedy` becomes `proposal awaiting adoption`, preserving the unadopted
backup-metadata status without a new review cycle or source change.

RP-733's source-map completeness is independently supported. The COMBINED
recovery protocol, delayed-action execution, actual runtime implementation,
native supported-host and exact-artifact R-006 remain unproved/unbuilt. This
verdict adopts no specification, operator behavior or release. No publication,
archival, deployment, new playable-tier claim or goal pause. Continue constructing
the complete mechanism under the full nine-tier1.0 objective.

## 2026-11-04 — Combined controlled candidate execution; instrument correction needed

Predeclare the three complete populations in the owning plan BEFORE19985. One
combined candidate uses a contained helper scope, known job/application/DB
namespace removal, full network identity, non-reused volume name and writable
output path. The private request proxy receives complete requests before helper
termination and forwards them only after it reads the new-operation admission
marker. This is BEFORE-daemon proxy delay, not a claimed native daemon queue.
Actual pg_restore COPY is separately observed waiting at an advisory lock with
zero committed rows. Controlled HTTP/local/daemon file actors and an unattached
named-volume sentinel cover effect classes, not the real Go CLI, Account API,
encrypted backup/protection algorithm or complete production topology.

19985 exits0 and completes all three populations under unchanged60s measured-work
per-population guards (three independently reset populations,180s maximum work;
ordinary commands<=10s within each deadline, cleanup<=15s PER command outside it).
No end-to-end wall/RTO/performance budget claim. Engine28.4.0 Linux/arm64 Desktop,
Compose2.39.4-desktop.1, exact amd64 Postgres16 config75f5a969 and cached Python3.12
arm64 configd2924abe; no pull/build/new shipping image, privileged container, host
port or production/receiver/secret access. The allowlisted proxy alone mounts the
existing daemon socket and forwards exact project-owned requests. Parent/helper
and entire-container cancellation are different explicitly exercised policies.

| Population | Actual current rows | Metric / checkpoint / HTTP marker / volume sentinel |
|---|---|---|
| parent_only | 7,13,14,42 | old / absent / old / absent |
| reused_names | 14,42 | old / absent / new / absent |
| generations | 42 | new / checkpoint-current / new / new-volume |

Parent-only cancellation leaves local/daemon/HTTP writers active. Whole containment
removes them, yet name reuse still lets a delayed create reach replacement SQL/
host targets and a delayed volume deletion remove the prepared replacement volume.
Generation isolation retains every tested replacement value; old known start
returns404, late create201 still occurs and old-volume delete204 still occurs.
The driver-forced late start is EXTRA adversarial work: a dead contained caller
never received that new ID and could not have submitted that start itself.
The source-bound v1 observation retains all identities/results and source hash
9ffc51755474a90cbd8a5a8e32be25643bfc04edb4188bac63ea9e5b35a5956b.
Successful new SQL/HTTP/volume writes precede admission, so absence is not inferred
from an unavailable replacement. Exact9 post-run project container/network/volume
inventories independently rechecked empty. All handles terminal.

Actual-source AST guard/classifier controlsb5c23c pass four command completions,
four handshake boundary/late cases, pre-start refusal and four exception classes.
Further actual-AST audit finds RP-735: the volume observation maps ANY nonzero CLI
inspection to absent sentinel, even injected ordinary permission failure. This is
a real instrument defect, not an error observed in19985 (whose volume Delete204
results are retained). Preserve this original source/observation before repairing;
next require typed404 versus ordinary failure and rerun final source. No instrument
approval/complete mechanism may be inferred from the initially successful run.

`combined-fence-candidate.md` names concrete remaining durable real-helper,
packaging, four-network/five-volume, host writer/retention, pre-backup abort and
checkpoint/pair bindings. Late old-generation mutations are ISOLATION, not the
draft's literal quiescence; explicit contract reconciliation/review/adoption is
required rather than redefining an acceptance word. No source/runtime/normal-CI,
kernel0.3.225/epoch8, game artifact, authored copy or operator policy is changed.
Review by: Codex implementer first filter. Recorded by: Codex root. Complete range
after59ea5552 through the containing commit plus the forthcoming correction needs
fresh independent research/tool/technical review. Full CR4/native-host/R-006,
original Deployment/whole1.0 acceptance and owner adoption remain open; no push,
archival, host deployment or goal pause.

## 2026-11-04 — RP-735 typed-volume correction and final combined execution

Preserve original7-path59ea5552..98af4c8d and source-bound v1 unchanged. Its actual-
AST permission controlf4e689 becomes None without raising, confirming RP-735's
error/absence conflation. The corrected proxy exposes only two allowlisted READ-
ONLY volume paths in addition to the predeclared exact owned mutation requests.
It returns the actual Engine HTTP status and only name/project metadata for200;
the driver accepts exact404 as absence, exact owned200 as present, and rejects
ordinary failures, malformed shape or ownership mismatch. No whole environment,
mount path or secret response is emitted. Unknown actor roles also fail loudly.

Actual-source327ed4 controls: owned200/typed404 accepted; seven403/500/missing/
wrong-name/wrong-owner/extra-error/string-status cases refuse. No deadline,
population, hypothesis or replacement-state criterion is changed. Initial source
patch is caught by direct inspection before execution because its new handler
landed in the wrong local class; it is immediately moved to the proxy before any
Docker population runs. No invalid run or correction evidence is invented for
that unexecuted edit.

Final70426 exits0 on ALL three populations, with same exact row/file/HTTP/volume
outcomes as19985. Actual volume responses are404/404/owned200, now retained in
`combined-fence-observation.v2.json`, binding source SHA256
ee9d482ca10c0a351295e4570686c2900b3d875b37df47f341b428822e8be2a6.
Both weak controls mutate replacement state; generation population retains42,
new metric/checkpoint/HTTP marker/volume sentinel. Late create201 and old-volume
delete204 STILL occur, so literal quiescence is not proved. Extra forced late
start fails with the retired-ID error only in the generation population, under
the candidate's stated stronger adversarial framing. Originalv1 remains historical.

ed2d2c independently checks v1 against exact98af4c8d source and v2 against current
source, ordered complete three-population denominators/objective flags, and
identical actual run/wait/outer-error AST. Priorb5c23c guard controls therefore
apply unchanged; no repeated unrelated suite. All9 final-project container/
network/volume checks are empty, as the initial9 checks were. All execution
handles terminal; no retained task cache/build image/shared resource changes.

The complete original7-path range plus correction through this containing
commit requires fresh independent tool/research/technical-draft review. RP-735
is corrected, not independently closed yet; RP-734's controlled isolation result
does not authorize a real protocol, pre-backup teardown, new shipping supervisor,
protected-backup metadata change or production implementation. Concrete production
bindings in the candidate table, native supported-host/actual helper/backup/pair
and R-006/adoption/full1.0 gates remain OPEN. No push, archival, deployment,
authored-copy change or goal pause.

## 2026-11-04 — Original combined-prototype union review: CHANGES REQUIRED

Review by: independent Codex `/root/combined_fence_prototype_research_review`.
Recorded by: Codex root, transcribing its returned verdict.
Complete reviewed union, base exclusive/tip inclusive:

- `59ea5552135a1bdfde21d5544db237131389a14d..98af4c8d68ce31c9f384441dd4da07a6d845ce8c`.
- `98af4c8d68ce31c9f384441dd4da07a6d845ce8c..811db80d0aa43222726a08223666be2759254b2d`.

Verdict: **CHANGES REQUIRED**, two instrument-refusal findings. Actual-ASTa78aa5
accepts injected late-start permission denial because ANY nonzero CLI result was
allowed; unchanged state can still set candidate_held. Actual-tailc3d2a0 advances
final file collection59→61 at deadline60 and still sets normal_complete/held.
RP-736/737 record these distinct failures. Neither occurred in actual measured
runs. Independent full actual66964 exits0 and reproduces all3 expected states and
volume404/404/owned200. V1/v2 bindings, corrected volume classifier/historical
RP-735 reproduction, run/wait/outer-error guards and range whitespace pass.
All handles terminal; all9 owned-project container/network/volume inventories
empty. No tracked reviewer edits/commits or retained scratch, no full mechanism/
runtime/adoption/release approval. Preserve this negative verdict.

## 2026-11-04 — Typed late-start and final deadline corrections; final execution

The extra adversarial start now passes through the fixture proxy, bound to the
actual late-create response ID and independently inspected full ID/owned project/
Postgres image. It records the actual Engine status/container/message, not a
synthetic CLI exit. Weak controls require204; generation refusal requires typed
404/500 plus the exact expected retired-network message bound to that container.
All ordinary permission/transport/wrong-container/wrong-network and unexpected
success responses fail the observation rather than count as fencing. No arbitrary
Docker endpoint is exposed; the extra route is limited to that one known create.

Actual-source2434cf accepts ordinary start and both declared typed retired-network
forms, refuses six error/target/network/unexpected-success controls. Actual-tail
6aafa4 reproduces old late acceptance and verifies corrected timely59.5 acceptance,
exact60 refusal and61 refusal. The final deadline check follows ALL result/file
collection immediately before normal_complete; cleanup's separate<=15s PER-command
allowance is unchanged. No fixture/guard/population/target criterion is weakened.

Full final29543 exits0, all3 populations: exact rows7/13/14/42,14/42,42 and the same
host/HTTP/checkpoint/volume effects as before. Typed extra-start statuses are now
204/204/404; generation404 contains the exact retired network ID. Volume statuses
remain404/404/owned200. New source-boundv3 binds SHA256
27ab7d15908660ead682dcef54fa7b3d4ae4f3ec879e09bc51ddb18c8502d9b8;
originalv1/v2 and their exact committed sources remain intact. Positive replacement
writes, old COPY and complete proxy-held requests all execute; no incomplete
population or late result is promoted. All root handles terminal; exact final
project cleanup checks and source/record binding precede corrective commit.

The original union plus this containing corrective range needs independent
coherent re-review. RP-735–737 are instrument corrections, not full CR4 recovery.
The successful controlled scheme STILL permits quarantined old-generation create/
delete after admission and is not literal quiescence or production implementation.
Real helper/durable ownership/pre-backup non-destructive branch/all resources,
protected encrypted backup/pair, native host and adoption/R-006/whole1.0 remain.
No runtime/CI/kernel0.3.225/epoch8, authored-copy, acceptance or operator-policy
change, push, archival, host deployment or goal pause.

## 2026-11-04 — Corrected combined-prototype union independently APPROVED

Review by: independent Codex `/root/combined_fence_prototype_research_review`.
Recorded by: Codex root, transcribing its returned final verdict.
Complete reviewed union, base exclusive/tip inclusive:

- `59ea5552135a1bdfde21d5544db237131389a14d..98af4c8d68ce31c9f384441dd4da07a6d845ce8c`.
- `98af4c8d68ce31c9f384441dd4da07a6d845ce8c..811db80d0aa43222726a08223666be2759254b2d`.
- `811db80d0aa43222726a08223666be2759254b2d..954abe091392762a91394aeaca3f2c4a26779339`.

Verdict: **APPROVED**, corrected research/tool and technical-draft scope. Original
CHANGES REQUIRED remains. Bound owned-container Engine start/retired-network
classification and the final collection-deadline acceptance boundary resolve both
blocking findings. Fresh independent40535 exits0: all3 real populations match
SQL/file/HTTP/volume outcomes, starts204/204/404, volumes404/404/owned200.
Actual-source07c8c8 accepts declared results and rejects permission/transport/
target/network/unexpected-success errors; old late collection reproduces, corrected
timely passes/boundary-late refuse; all3 artifact/source bindings pass. Full-range
33d154 passes complete ranges, preserved negative verdict/artifacts, unchanged
product/CI source and whitespace. Root executions are not relabeled reviewer work.

All reviewer handles terminal; nine exact owned-project resource inventories empty;
no edits/commits/scratch, checkout clean. Reviewer requests one non-blocking prose
precision, applied in this editorial closeout without source/rerun: only the TWO
contained populations guarantee the late ID was not received; parent-only request
descendants may survive, although they never issue a start by actor design. The
driver's explicit stronger adversarial start framing is unchanged.

RP-735–737 close their instrument defects; RP-734's controlled isolation evidence
is independently supported. Late old-generation create/delete STILL occur, so no
literal quiescence, real recovery protocol, adoption, native supported-host or
release readiness follows. Construct durable full real-helper ownership/protocol,
pre-backup branch, packaging/all resources and checkpoint/previous-pair binding;
do not rerun completed controls for reassurance or promote the prototype into
acceptance. Whole nine-tier1.0 stays active; no push/deployment/archival/goal pause.

## 2026-11-04 — Examine the pre-backup branch without restarting live Postgres

After49b6113e, address the combined candidate's explicitly incompatible pre-backup
DB teardown, not another unchanged three-population repetition. Predeclared the
two populations/60s measured-work guard and exclusions in the owning plan before
execution. New manual `probes/prebackup-live-database-probe.py.txt` uses only a
uniquely owned declared tmpfs fixture, exact cached amd64 image/config, no pulls,
ports, real data, application deployment or privilege change. Engine28.4.0 on
Linux/arm64, Compose2.39.4-desktop.1; actual cached Postgres reports16.15.

Actual42009 exits0 and cleans up. During an explicit libpq post_auth_delay=8,
the observed process exists with ZERO pg_stat_activity rows for its PID. Exact
worker removal leaves the same PID/kernel-start identity temporarily alive;
it subsequently disappears without executing its offered INSERT. This adds
RP-738: a successful empty session census does not establish physical quiescence.
Delay is a controlled test population, not an observed ordinary backup incident.

Actual pg_dump reaches an observed ACCESS SHARE relation lock wait. Exact worker
removal leaves that same backend/process blocked. In this owned fixture only,
matching its observed PID/start/address/port/application and requesting bounded
pg_terminate_backend(...,3000) removes it. After lock release, original row7 and
new control row42 are exact. A continuously connected separate control retains
its PID/kernel start, while the exact database container and postmaster start
remain unchanged. This demonstrates known-backend cancellation without whole-DB
restart; it is neither the actual Go backup helper nor a complete abort.

Actual-source f487d1 controls accept timely completion, reject exact/late guard
completion and propagate permission/daemon errors. Retained v1 JSON matches
executed source SHA2567091c5cf0e371d0f1bd75e4631db3b6e458c60d58804edf3857be1355851e5e4.
Syntax, whitespace and retained binding pass. Root independently rechecks all3
exact project resource inventories empty; all execution handles terminal and
temporary fixture directory removed. Initial ordinary Docker version access was
sandbox-denied before fixture execution; permitted exact probe succeeds. No
invalid population, raised guard, incomplete measurement or suite-green claim.

Production ownership remains unresolved: pg_terminate_backend ultimately signals
a PID; its query predicate does not eliminate check/signal PID reuse. Log/address
association here is an instrument, not durable registration. Startup completion,
new/late clients, actual pgx snapshot exporter/helper/publication, all pre-backup
effects and supported-host proof remain. Candidate/draft/plan/RP-738 reconcile
these facts without adopting a recovery policy or changing runtime/CI/kernel.
Review by: Codex root, implementer first filter. Recorded by: Codex root.
Complete after49b6113e research/tool/draft range needs fresh independent review;
full CR4/Deployment/1.0 acceptance stays open. A text-only native Linux/amd64
disposable-host availability question was delivered earlier this turn; unanswered,
no host or external deployment authority inferred. Local work continues.

## 2026-11-04 — Pre-backup review finds an unbounded interactive write; correct it

Review by: independent Codex `/root/prebackup_live_database_research_review`.
Recorded by: Codex root, transcribing the returned verdict. Complete seven-path
49b6113ead8e37c8c0e2d464ebf3877661592897..83abd00992cab08f9d608e0bfd609a70230f6e9e
is **CHANGES REQUIRED**. Actual-AST8069e8 advances write/flush to11s then returns
a matching line; old send() accepts because writing precedes the local timeout
and final check uses global60. RP-739 records this instrument defect. No recorded
real population is observed exceeding that handshake bound; the original source
and v1 remain historical, never silently rewritten.

Independent38051 exits0 reproducing both actual populations and exact rows7/42.
df2551 actual wait/Docker wrapper/error/source-binding controls pass; db0ec0
actual process-parser/acceptance-tail controls reject live backend, wrong data,
changed control and late global completion. Initial control40b306 lacked a test
namespace variable; corrected execution passes. Reviewer exact-project
dcfee114337423d3 container/network/volume inventories33955b/2c8dc2/8dc220 empty;
all handles terminal, temp directory removed, no tracked changes or commits.

Forward correction uses a genuinely nonblocking stdin pipe, select/write and
queue-read bounded by ONE local deadline established before any write; all
matching completions are checked against it, capped by the global deadline.
No background blocked writer or buffered flush survives the timeout. Actual
25892b compiles the ORIGINAL nested send from committed83abd009 and reproduces
11s acceptance. Current actual send_query accepts timely and partial writes,
refuses exact/late select/write/read, propagates permission/broken-pipe errors,
and rejects an exhausted read. No guard/population/criterion is raised.

Fresh actual79437 exits0: startup census0/process survives client removal then
exits without INSERT; pg_dump survives client removal, bounded known-backend
termination removes it; same DB/postmaster/control process+connection and rows7/42.
V2 binds exact current source7b93d963a2d2604547bd19294bd08f5aa3cde132bf5e2eebc9ae34e0aa99a924.
Both real populations and cleanup complete; syntax/diff/source bindings checked
before commit. This is still a narrow disposable-fixture candidate, not durable
production PID ownership, actual backup helper or native-host/full CR4 proof.
Complete original plus containing corrective range needs independent re-review.
Review by: Codex root, corrective author first filter. Recorded by: Codex root.

## 2026-11-04 — Corrected pre-backup research union independently APPROVED

Review by: independent Codex `/root/prebackup_live_database_research_review`.
Recorded by: Codex root, transcribing its final returned verdict. Complete union:

- `49b6113ead8e37c8c0e2d464ebf3877661592897..83abd00992cab08f9d608e0bfd609a70230f6e9e`.
- `83abd00992cab08f9d608e0bfd609a70230f6e9e..ee90913f6f195dd7e972c5eee4d7a3f101a4c1eb`.

Verdict **APPROVED**, corrected research/tool/draft scope, no remaining blocking
findings. Original CHANGES REQUIRED remains. Nonblocking writes/readiness/queue
reads/final acceptance now share one local deadline; RP-739 closes its instrument
defect. Independent actual57617 exits0 reproducing both PostgreSQL16.15 populations,
preserved DB/postmaster/control connection and exact rows7/42. Independent21d8f5
exits0 reproducing original11s acceptance, accepting timely/partial/would-block,
rejecting exact/late select/write/read/global deadline, permission/broken-pipe,
zero-progress and exhausted-read controls. An actual full nonblocking OS pipe
times out without a buffered writer; historical/current source bindings pass.
Independent9f5335 passes full union scope/whitespace/negative preservation and
clean checkout; product and CI source remain untouched.

Reviewer own projectd0ddd27bcd998cc8 container/network/volume inventories
181d5d/9aab4e/5b259e are empty; all handles terminal, no tracked edits or commits.
This editorial closeout records that verdict without changing the reviewed tool
or rerunning unchanged checks. RP-738 stays OPEN: durable production ownership,
startup completion and generation-safe targeting/real helper remain unresolved.
Next build/specify that complete ownership/containment protocol, including all
real pre-backup effects and pending clients; do not adopt the fixture's PID query
or repeat narrow completed positives. No full abort/CR4/native-host/adoption,
release acceptance, push, archival, deployment or goal pause follows.

## 2026-11-04 — Connection-owned gateway supplies a pre-backup cancellation candidate

After20cbb5b8, advance RP-738's missing ownership direction rather than repeating
PID termination. Predeclare wrong-key/correct-key/startup populations in the
owning plan. New manual Go source, control overlay and Python runner use an
ignored compiler overlay with EXISTING root build/test targets and repository
Go cache; tracked production main/main_test and CI stay byte-unchanged. The
compiled binary runs in a uniquely owned declared tmpfs fixture with the exact
cached amd64 Postgres image/config, no pull/ports/real data/Docker socket in jobs.
Local Docker Desktop Engine28.4.0 Linux/arm64 and Compose2.39.4-desktop.1 run
amd64 Postgres16.15. No native supported-host claim.

The transparent gateway registers one in-memory connection before upstream dial,
observes the PG16 BackendKeyData secret ONLY in memory, and retains/drains the
original server stream after actual pg_dump is killed/waited. It sends a
connection PID/key CancelRequest rather than pg_terminate_backend or OS PID kill.
No generated key enters output or tracked artifacts. PostgreSQL16 protocol's
documented delivery-versus-query-outcome distinction is tested, not inferred.

Actual35493 exits0, ALL3 populations: wrong-key delivery closes normally while
the same original query remains lock-blocked, then server EOF follows explicit
lock release; correct-key server EOF precedes lock release; startup delayed8s
has no key/activity row yet remains registered after client death until actual
server EOF. Each retains same connected control PID/postmaster and exact rows7/42.
No complete result is emitted on an incomplete population. Per-population60s,
total180s work guards and<=10s handshakes remain; build<=180s and outer actor
observation<=240s are NOT work/RTO budgets; cleanup<=15s PER command outside work.

Actual46325 exits0: fourteen manual actual-function controls cover partial/failed/
zero-progress writes, timely/expired matching handshake, ordinary EOF, malformed/
truncated/oversized frame, bad/duplicate BackendKeyData, known closed frontend,
typed cancel delivery EOF versus unexpected response and timeout. Initial18389
could not bind loopback under sandbox; its socket populations never execute.
Same target with permitted listeners passes without source/assertion changes.
Control overlays replace compiler inputs only; no production source mutation.
Gofmt/retained bindings/whitespace and root overlay vet pass; existing build73943 exits0. Recorded JSON
binds Go bce2a8e7c9f1677049689b9a1a9ec7f3232fa695c8fc7223155149857a03d6d2,
runner ac923c40d1d8aac1b61c7d36d93faa2d3a96f4ed2511cd7be15bd6abbfd01821 and
executed binary a6f646b4221402b2d4bc273eff540f31c78bf17ed626df032bb0e9bad2dee4de.
Root verifies exact fixture container/network/volume inventories4a6fe0/24ce8b/
ac04ba empty, fixture work removed, actor/control handles terminal.

RP-740 is an OPEN candidate, not production authority. Candidate document now
chooses a fail-closed supervised-operation direction over the whole reviewed
mutation map: durable native supervisor identity, non-bypassable Engine admission/
terminal receipts and ALL actual helper SQL streams, no guessed PID on authority
loss, full existing checkpoint/backup/pair/deletion holds. These are requirements
to SPECIFY/PROVE, not existing packaged mechanisms. Ephemeral native scope/service
packaging needs explicit contract reconciliation/adoption; do not silently add a
permanent deployment service. This one-connection experiment excludes actual pgx
snapshot exporter/helper/encryption/publication, all Engine actions and native
host. Full CR4/R-006/1.0 stay open. Next construct durable supervisor/broker
protocol and critical-loss refusal, then bind the complete real helper, not more
identical pg_dump positives. Review by: Codex root, author/first filter.
Recorded by: Codex root. Complete after20cbb5b8 research/tool/draft range needs
fresh independent review. No push, deployment, archival or goal pause.

## 2026-11-04 — Connection-owned cancellation candidate independently APPROVED

Review by: independent Codex `/root/connection_cancel_gateway_research_review`.
Recorded by: Codex root, transcribing returned final verdict. Complete nine-path
range20cbb5b8c7639bc3af960fe479d04c07feaec4e8..fd4c5b923bf7107bf6d59a053d039a9ce1c0db55
is **APPROVED**, no blocking findings. Independent actual2276 exits0 reproducing
all3 real populations, unchanged control/postmaster and exact rows7/42.
Independent42422 exits0, all14 actual-function controls;bc06cb passes complete
scope/whitespace, retained Go/runner/executed-binary bindings and scratch removal.
Production and CI source remain untouched. Review did not rely on root's run
as its own execution.

Reviewer's exact owned project6ac92350866d4b2e container/network/volume inventories
ff9051/fc60d4/093fc7 empty; all handles terminal, no tracked edits/commits.
This editorial closeout adds portable root reproduction commands and reconciles
review status only; no reviewed source, observation or criterion changes.
No unrelated suite or identical actual population rerun is necessary.

RP-740/738/734 remain OPEN. Approval covers the experimental candidate and stated
technical direction, not durable supervision, complete actual helper connection
ownership, non-bypassable isolation, full CR4/native-host or adoption. Next specify
the complete journal-bound supervisor/request-ownership protocol and critical-loss
refusal, then bind the actual helper/all reviewed mutation classes. Keep the full
nine-tier1.0 active; no publication, deployment, archival or release promotion.

## 2026-11-04 — Independent draft review identifies recovery-ordering contradictions

Review by: independent Codex `/root/recovery_admission_contract_review`.
Recorded by: Codex root, transcribing returned review. Reviewed object is the
Current Release Authority RFC atc5419034f7bb94189670caa5f8a99029412defdc plus its
relevant controller/CLI sources, NOT a reversal of the approved gateway experiment.
Verdict **CHANGES REQUIRED before specification acceptance**:

- RP-741/P1: universal CR2 healthy-seven-container admission blocks CR4's allowed
  partial/mixed failed-source rollback.
- RP-742/P1: CR1/CR3 successful Finished establishes known without whole-helper
  closure; actual controller.go220 precedes cmd/deployment-release/main.go84's
  outer publication. Finished/helper death/new ordinary release can bypass a
  fence if the condition only names abort/retry/rollback.
- RP-743/P2: requiring quiescence before ANY recovery leaves no explicit lawful
  phase for the fencing cleanup that establishes it.
- RP-744/P2: successful rollback supposedly closes a pending release, yet no
  Finished exists and CR5 rejects unclosed rows; ancestry/orientation/retry
  windows and derived closure need exact normative rules.

Reviewer also requires authenticated attachment to the live lock-owning
supervisor or a proved sealed handoff, and tracking asynchronous native unit
jobs if that packaging is selected. No tracked edits/fixtures/runtime claims;
all reviewer handles terminal. These are draft defects, not shipped executions.

Forward correction reconciles the BODY in one range: CR2 healthy observation
now scopes to ordinary release, abort conclusion and restored success; recovery
instead requires exact admitted resource ownership. CR1/CR3/CR6 require every
successor's predecessor fence and success's full completion proof, including
outer writers. Pre-backup failure cannot be finalized without closure/prior
identity; uncertain post-checkpoint failures never mint ordinary admission.
V1 identity is not a completion proof: concrete first-install/handoff producer
is an explicit remaining construction gate, not grandfathering or a fake sidecar.

CR4 separates read-only admission, checkpoint/input-authorized bounded cleanup,
durable recovery_fenced, then destructive reset/restore/start. Pre-backup abort
may clean only operation-owned work and preserves live services/data. Explicit
recovery_of ancestry fixes source/target/backup/window, retries follow the failed
child, and successful recovery derives parent recovered without fabricated or
rewritten Finished. Source refusal/cleanup and restored-target success remain
different observations. CR3 gains semantic proof/ancestry fields and the new
boundary event; CR5 reads them, and acceptance includes all successful and
refusal traces. Prefix binding excludes the referring event to avoid hash cycles.

This is NOT a completed serialized supervisor protocol: exact proof/work-log
schemas, first-install/v1 handoff, native profile, non-bypassable admission and
actual producer/consumer executions remain explicit CR6/Open-question gates.
No new OS service/privilege/adoption is inferred; selected packaging must
reconcile the no-service promise and host floor before acceptance. No product,
canonical behavior, CI, old evidence, authored copy or owner ruling changes.
Root checked the complete normative diff and concrete trace ordering against
reviewed source; no software suite/mirror validator for a prose-only amendment.
Review by: Codex root, corrective author first filter. Recorded by: Codex root.
Complete afterc5419034 four-path correction needs independent technical review;
full CR4/Deployment/1.0 remain open. No push/deployment/archival/goal pause.

## 2026-11-04 — Recovery ordering correction independently APPROVED

Review by: independent Codex `/root/recovery_admission_contract_review`.
Recorded by: Codex root, transcribing its returned verdict. Complete four-path
c5419034f7bb94189670caa5f8a99029412defdc..0a89f026e6aee83f4c3607884a2aec3eb863d2ee
is **APPROVED — technical ordering correction**, no remaining blocking finding
in this amendment. Original CHANGES REQUIRED remains. RP-741–744 close ONLY
their normative ordering defects, not any production implementation or larger
mechanism tracked by RP-734/738/740.

Reviewer independently inspects the entire diff and derives ordinary-success,
partial/mixed-source, preserved-prior abort, pre/post-fence failed-retry,
interrupted-parent, lost-authority, prefix-hash and v1/bootstrap traces directly
from the revised clauses. Old success cannot waive a predecessor fence; cleanup
does not authorize reset; failed descendants do not close the parent; first
install's absent handoff remains a blocker rather than grandfathering. Scope/
whitespace999b10 exits0. No mirror validator or unrelated runtime suite for this
prose-only amendment. No tracked edits/commits/fixtures; all handles terminal.

This editorial closeout reconciles status/index/candidate links without changing
the approved normative body. Next specify exact supervision/work-log/fence/
completion artifacts, native permission/profile and first-install producer,
then implement actual helper/runtime paths after the applicable acceptance/
adoption gates. Non-bypassable admission, useful successful recovery/native
execution and whole1.0 remain open. No owner authority, OS prerequisite, service,
runtime behavior, publication, deployment, archival or goal pause is inferred.

## 2026-11-04 — Specify operation evidence envelopes and phase reduction

Afterb74d02a7, split the concrete record contract from the still-unselected native
producer/profile. New draft `rfc/deployment-operation-evidence.md` defines closed
canonical private objects for supervisor binding, phase work logs, terminal receipt
envelope, predecessor/completion closures and initial-install handoff. Parent CR3
adds required immutable supervision_sha256; index and owning next action link the
proposal. No production decoder, generated schema or tracking validator is added.

The format fixes ordinary byte/hash/UUID/counter/date rules without changing v1
journal bytes. Native plans/subjects/evidence remain EXACT OP9 prerequisite types,
not free-form extensible claims; unknown producers refuse and a universally
refusing reader is explicitly not useful acceptance. Trust is the approved local
operator/supervisor plus independently executed native conformance, not a new
signature or an assertion that offline bytes replay historical socket EOF.

Rollback fencing and execution use separately sealed phase streams: the first
authorizes recovery_fenced, the second opens only afterward. Old outstanding
items remain in their original open logs; newly admitted cleanup has its own
Started authority. Work children are reserved before effects, native identity
binds once, terminal delivery differs from actual actor/stream closure, and
retained runtime services have explicit verified ownership rather than being
pretended dead or relabelled temporary publishers. Exact prefix/hash boundaries
exclude referring events; no self-hash or stale-proof reuse. Initial handoff
covers EVERY supervised v1 install attempt through successful install, preserves
original rows, and does not erase an unfinished installer from empty history.

OP9 still needs complete native profile/plan/subject/receipt schemas, supervisor
bootstrap and daemon/scope authority, non-bypassable Engine/SQL/HTTP admission,
permissions, live lock-owner attachment/sealed handoff and all actual producers.
Selecting new services/prerequisites requires explicit reconciliation/adoption;
this record proposal selects none. Actual helper/native successful and refusal
paths plus clean-host/RPO/RTO remain gates; the approved toy gateway is not an
OP2/OP6 producer. No retrospective certificate or v1 grandfathering is permitted.

Root reads the complete proposal/parent diff and traces ordinary admission,
checkpointed recovery's phase split, initial handoff, typed refusal/unknown work,
child bindings, closed logs and retained-service ownership. Whitespace/path/ref
checks pass. Prose-only normative proposal: no software suite/mirror validator
or unchanged pg_dump experiment for a document. Review by: Codex root,
proposal author/first filter. Recorded by: Codex root. Complete afterb74d02a7
six-path proposal needs independent technical review; OP9/profile and owner
adoption remain, not product implementation. RP-745 records the prerequisite.
No code/CI/kernel/copy/policy change, push, deployment, archival or goal pause.

## 2026-11-04 — Evidence-contract review finds anchor/install/phase conflicts

Review by: independent Codex `/root/operation_evidence_contract_review`.
Recorded by: Codex root, transcribing its returned verdict. Complete six-path
b74d02a735403e79c2e90147c402fe1ffe5ce8ba..7a7e7b39a71f29dec28c0f70039060a4a550a3e2
is **CHANGES REQUIRED — serialization contract**. Original verdict is preserved.

- RP-746/P1: the same immutable predecessor proof repeats at later CR3 events
  but OP6 demanded a new immediate prefix at every reference.
- RP-747/P1: OP7 installation creates no Started yet failed to substitute all
  OP2 Started-dependent prefix/date/publication conditions.
- RP-748/P2: failed-before-fence rollback cannot open execution but OP6 required
  both logs for its completion/inherited retry evidence.

Independent6c3e0d exits0 for complete scope, references and whitespace; product/
CI untouched. Whole normative diff/parent reviewed directly, no mirrored
validator/unrelated suite. No tracked edits/commits/fixtures; all handles terminal.

Forward correction gives every descriptor/closure its ONE unique introduction
anchor, scopes native reobservation to the proper boundary rather than repeated
historical fields, and forbids stale reuse across operations. Initial-install
descriptor publication/first work admission explicitly substitutes for Started;
handoff separately anchors THROUGH the actual final v1 install row. Protected
effects still require admitted authority. Stage-derived phase coverage requires
only fencing before recovery_fenced and BOTH logs after it, with the current
predecessor proof's cleanup-only exception. Authorized no-work phases require
real sealed empty-work logs; omitted/torn/lost streams never become empty.
Receipt→terminal row→phase closure→proof→unique ledger-anchor dates are explicit.
Parent's CR3/CR6 bodies reconcile the anchor/coverage rule, not append a waiver.

OP9 remains deliberately open; no native receipt/profile, permission, actual
helper conformance, adoption or usable recovery implementation follows. Root
checks full containing diff and all affected ordinary/failure/retry/install
traces; no software suite for this prose correction. Review by: Codex root,
corrective author first filter. Recorded by: Codex root. Complete original plus
containing five-path correction needs independent technical re-review.

## 2026-11-04 — Corrected evidence-format union independently APPROVED

Review by: independent Codex `/root/operation_evidence_contract_review`.
Recorded by: Codex root, transcribing its final returned verdict. Complete union:

- `b74d02a735403e79c2e90147c402fe1ffe5ce8ba..7a7e7b39a71f29dec28c0f70039060a4a550a3e2`.
- `7a7e7b39a71f29dec28c0f70039060a4a550a3e2..09c1c792ce8de0795fb64880186c6d1c2fdac6a6`.

Verdict **APPROVED — scoped serialization contract**, no blocking findings.
Original CHANGES REQUIRED remains. Independently checked unique anchors and
historical repetitions, installation's non-Started admission, reached-stage
phase coverage, valid empty-work versus missing/torn/lost logs, inherited/current
completion/retry/administrative interruption and temporal order. Independent
0dd9b8 exits0 for original/corrective/union scope/whitespace, local references
and negative-verdict preservation. No software suite/mirror validator warranted
for prose semantics. No tracked edits/commits/fixtures; all handles terminal.

RP-746–748 close draft-format defects only. RP-745/OP9 remains OPEN: native
schemas/profile/producers, permissions, non-bypassable admission, actual helper/
runtime conformance, clean-host execution and adoption are not established by
record-format review. This editorial closeout reconciles status only, not the
approved normative rules. Next construct the native supervisor/worker profile
and exact producer/subject/receipt semantics, then prove it against actual work.
No product/CI/kernel/copy change, publication, deployment, archival or goal pause;
the full nine-tier1.0 remains active.

## 2026-11-04 — Actual isolated worker boundary, with final-collection finding

After9a17aa2a, predeclare a Docker-worker/private Unix-broker mechanism population
in the owning plan. New manual source runs cached config-pinned Python actors
and declared tmpfs PG16 fixture, unique project, no ports/pulls/new shipping
dependency/production data. Native Desktop Engine28.4.0 Linux/arm64 and Compose
2.39.4-desktop.1, amd64 PG16.15; not supported native Linux/amd64 proof.

Actual57766 exits0. Same-UID/cap-drop networked writable controls before/after
receive actual PG SSLRequest refusalN and read their written local marker.
Isolated UID65532 network-none/read-only/cap-drop/no-new-privileges has exact
ENETUNREACH101 to that PG, ENOENT2 for raw socket paths, EROFS30 at `/tmp` and
the deliberately world-writable channel marker. Kernel CapEff/CapBnd0 and
NoNewPrivs1 observed. Allowlisted native image/UID/network/mount/read-only/
capability/security/PID/IPC/privilege projection matches the intended actors;
no whole environment/config is inspected or emitted.

Private Unix GET reaches actual Engine200/OK; correct-UID POST and UID0 GET
receive distinct typed403 refusals. Broker uses SO_PEERCRED, not supplied UID;
deliberate socket/file0666 modes force the policy boundary to execute. Intact
channel marker is preserved. All actor waits succeed and exact owned resources
clean up before complete output. V1 binds exact source
edcbf8df9e3962f6f8662c1de40870d7346657bf5e0a3be1b6668cf92eba5d86.

Actual-source545c35 checks three permitted errno classes and refuses wrong
permission/timeout/unknown classes or unexpected success. It also confirms a
ROOT INSTRUMENT DEFECT, RP-749: final source-hash/metadata collection advances
59→61 after the last guard, yet completed=True with deadline60. Same error
class as RP-737 in a new probe; record it rather than hide the repeated mistake.
No actual late collection is observed in the retained real run. Preserve first
source/v1, then forward-correct final collection and rerun changed controls/
subject. No bound or population is raised/dropped.

RP-750 is a narrow isolation candidate, not full OP9. Actual Go/Docker/Compose
workers, per-job channel authority beyond UID, all write routes/bodies, durable
supervisor/daemon generation/bootstrap, complete SQL/HTTP/output lifetimes,
native host/adoption/whole recovery remain open. Candidate/OP9 links reflect
that limit. Review by: Codex root, implementer first filter. Recorded by:
Codex root. Complete after9a17aa2a range plus containing correction needs fresh
independent research/tool/technical review. No runtime/CI/kernel/copy change,
push/deployment/archival or goal pause.

## 2026-11-04 — Correct worker probe's final and command completion boundaries

Preserve original source/v1 ataa765a4026fce0b5311fbcbe4401ec092eab1c68.
Root actual-wrappera41a20 also exposes that requesting subprocess timeout10
does not itself reject a matching return observed at11 under global60. RP-749
now records both missing post-completion boundaries. No real timed-out native
event occurred in either retained actual population.

Forward correction: local deadline starts before each Docker invocation, covers
output collection and is checked afterward; cleanup has its original15s local
deadline outside work. Global work guard is checked AFTER metadata/source-hash
collection, before completed=True. Actuald110f3 compiles the historical/current
functions and reproduces old61s acceptance, passes timely collection and9s/14s
ordinary/cleanup, rejects exact60/late61 final, exact10/late11 ordinary and
exact15/late16 cleanup. No guard or population is increased/dropped.

Fresh actual76328 exits0, every original worker/control/broker population matches:
typed101/2/30/30, Engine200/OK, UID/route403, native/kernel profiles, before/after
PGN and local-marker writes. V2 is the normalized actual second output and binds
source3d672aa9771eb1d70455aea4daa648b78e9584fd5cf91bdcadad607815b042c6;
only project/network/source identities differ from v1. Actualdc6858 exercises
the actual broker AST: allow/UID/route discrimination, native500/bad body and
permission failures cannot turn into policy refusals. Actual native-profile AST
accepts intended projection and refuses altered user/network/read-only/caps/
privilege/mounts. Earlier545c35's actual errno negatives still apply unchanged.

All root actor handles terminal and source cleanup verifies its entire owned
project removed. Syntax/bindings/whitespace pass before corrective commit;
independent resource inventories are rechecked too. Complete original seven-path
plus containing correction needs fresh independent review. Review by: Codex root,
corrective author first filter. Recorded by: Codex root. RP-749/750 remain open
pending review/full-profile construction respectively; no OP9/adoption/runtime/
supported-host/release claim or goal pause.

## 2026-11-04 — Independent worker review fires native mount-order mismatch

Review by: independent Codex `/root/worker_isolation_research_review`.
Recorded by: Codex root, transcribing returned verdict. Complete union
9a17aa2a2f0269598815d43af0ddc5e66a875ade..aa765a4026fce0b5311fbcbe4401ec092eab1c68
plus aa765a4026fce0b5311fbcbe4401ec092eab1c68..4f0c5464aaa94ee219780ce604d8800603e6afe6
is **CHANGES REQUIRED — research instrument**. RP-751/P2: actual92969 exits1
at broker inspect_profile. Diagnostic24761 also exits1 and captures native
/daemon.sock then /broker versus expected reverse, with identical type/RW/
destination values. Actual-function0ba47b reproduces the same false refusal.
Neither real run reaches after-control, so no independent full pass is claimed.

Independentf1efb8 passes old/current bindings, original late acceptance and fixed
final/command/cleanup/ordinary-error/errno controls. Independent394f4d passes
actual broker/native-profile corruption controls, full scope/whitespace and
append-only log. Product/CI untouched. First owned inventories326c8c/35c4c6/80cf1d
and diagnostic inventoriese44f45/914d07/2b6eef are empty; a237a3 removes both
retained temporary directories. All handles terminal; no tracked edits/commits.

Forward correction compares destination-ordered native/intended projections,
preserving all cardinality/field/value checks and refusing duplicates explicitly.
Actualab38d7 compiles old/current inspect_profile: old reversed-order input
refuses; both current orderings pass; duplicate, altered-RW and extra-host mounts
refuse. Actual bounded/run ASTs equal the independently tested4f0c5464 bodies;
no guard change. No retry, looser policy, population omission or raised bound.

Fresh actual14788 exits0 with before/after PGN and successful marker controls,
native101/2/30/30 refusal classes, actual Engine200/OK, route/UID403, kernel/
native profile checks and retained channel marker. V3 normalizes the actual
output and binds source2cfcb55c7f0966e2657ccb00bc7586cf58bdd7b53195f3a3df5c956278bab3b5.
Original source/v1/v2 and independent negative remain. Root fixtures/resources
are rechecked empty before corrective commit. Complete original plus both
corrections needs independent re-review, not OP9/adoption/recovery acceptance.
Review by: Codex root, corrective author first filter. Recorded by: Codex root.

## 2026-11-04 — Corrected worker-isolation union independently APPROVED

Review by: independent Codex `/root/worker_isolation_research_review`.
Recorded by: Codex root, transcribing final returned verdict. Complete union:

- `9a17aa2a2f0269598815d43af0ddc5e66a875ade..aa765a4026fce0b5311fbcbe4401ec092eab1c68`.
- `aa765a4026fce0b5311fbcbe4401ec092eab1c68..4f0c5464aaa94ee219780ce604d8800603e6afe6`.
- `4f0c5464aaa94ee219780ce604d8800603e6afe6..299ad11456bd0b6b5fc0d11b77388b420b7d5129`.

Verdict **APPROVED — corrected research/tool scope**, no blocking findings.
Original CHANGES REQUIRED and both partial failures remain. Independent actual
12798 exits0, ALL before/after/isolated/wrong-UID/broker populations complete:
PG16.15N/control markers, exact101/2/30/30 refusals, actual Engine200/OK,
route/UID403, kernel/native profiles and intact channel marker.

Independentdd3dfd exits0: old reversed mount refusal reproduces, all4 native/
intended order permutations pass; duplicate/missing/extra/altered mounts and
user/network/read-only/caps/security/privilege/PID/IPC corruptions refuse;
all3 artifact/source bindings pass. Prior independentf1efb8 guard/error/errno
and394f4d broker controls remain applicable: actual functions/embedded actors
verified AST/byte-identical. Independent082e80 exits0 checks all ranges/union,
nine-path scope, append-only negative preservation, refs, scratch and checkout.
Production and CI untouched. Root's executions are not relabelled reviewer work.

Own reviewer project09b322e8648d6ae3 container/network/volume inventories
c72573/9c448c/4bdff7 empty; all handles terminal, no tracked edits/commits.
This editorial closeout changes no reviewed tool/observation/criterion and needs
no identical rerun. RP-749/751 close instrument defects only. RP-750/OP9 stays
OPEN: actual Go/CLI, full job identity/write authorization, supervisor/daemon
lifetime, native Linux/amd64, adoption and useful recovery remain unproved.
Next actual Docker/Compose command workers plus Go broker/admission/receipt
binding, not another Python-ping repetition. Full nine-tier1.0 stays active;
no publication, deployment, archival, new permission or goal pause.


## 2026-11-04 — Actual Docker/Compose workers and Go read-only broker

Research only under the existing OP9 candidate plan; no production, CI,
shipping dependency, service installation or native-host deployment change.
Public static Docker28.4.0 client and Compose2.39.4 are ignored research inputs;
exact URLs/bytes/hashes are in cli-worker-observation.v1.json. Compose SHA was
checked against official GitHub release metadata; Docker archive has official
HTTPS origin plus local digest, NOT a separately checked signature/checksum.
No daemon extracted or run. Broker builds via root Make compiler overlay.

Actual cafdcb/418737 exits0, complete population: isolated UID65532 uses actual
Docker28.4.0 linux/amd64 through a Go Unix/HTTP broker to read real Engine28.4.0
linux/arm64. Actual Compose2.39.4 version/config render, pinned fixture fields
match. Config is local rendering, not Compose up/down or admitted write proof.
Raw socket native ENOENT2 plus same-binary successful proxy use; actual POST
ContainerCreate receives403, never forwarded, no child created. Actual UID0
version/ping requests receive403 by SO_PEERCRED. Native profiles match exact
image/user/mount/network/read-only/cap-drop/security/PID/IPC expectations; kernel
CapEff/CapBnd0 and NoNewPrivs1. Eleven actual Go policy controls pass inside the
fixture. Broker trace retains every request category/UID/status/error, including
one denied unknown GET, without recording bodies/env/credentials. Owned actors
and channel volume removed; all project container/network/volume inventories
empty before completed output. Root AST/hash/old-new log-shape control692628
passes; overlay Linux/amd64 vet0df9ed exits0. Both builds ede506/c2a3ed pass.

Interrupted earlier output is inadmissible: no session/terminal output retained.
Read-only process/inventory acdf97/2af141 finds no live probe/owned resources;
only cc-cli-worker-15e3e58c33dd6661 scratch remains. No claim of execution success
or exact runtime failure follows. Pre-commit source inspection found container
stdout expected3 lines though docker exec controls has its separate stdout;
corrected to actual2-line ready/trace. Actual-shape old/new control reproduces
false rejection. The source-bound retained complete run is after correction.
No population/guard/policy loosened. Original interrupted scratch is retained.

RP-750 remains OPEN: compatibility is one prerequisite. No durable work journal,
full per-job identity, stage-bound write authorization, whole-helper binding,
daemon/lock lifetime, complete bypass inventory, supported Linux/amd64 host,
OP9 adoption, actual recovery or Deployment/1.0 acceptance is established.
Next bind durable job admission and receipts to real producers instead of
repeating these unchanged compatibility populations. Complete range from
2699ad00 through this batch requires fresh independent delegated review.

Review by: Codex root, research/tool author first filter. Recorded by: Codex root.


## 2026-11-04 — Actual CLI/Go compatibility independently APPROVED

Review by: independent Codex `/root/cli_worker_research_review`.
Recorded by: Codex root, transcribing returned verdict.
Exact complete range:
`2699ad00d1be30753a08d6b11c60ba5c697c2286..04452299fca7592ee66fe7a1ad6c8b54b132ba82`.
Verdict **APPROVED — research/tool scope**, no blocking findings.

Independent c8216a/f54d40 exits0: actual Docker/Compose, Go11 policy controls,
actual ContainerCreate403 without forwarding, wrong-UID403, raw-socket ENOENT2
and exact native/kernel profiles reproduce. Output retained in ignored
.cache/cli-worker-broker-research/independent-execution-v1.log. Independent
f3ce08 executes reviewed AST: nine trace corruptions refuse; mount ordering
passes while twelve profile/mount corruptions refuse; root/worker exact and late
command deadlines refuse; unrelated permission/transport errors cannot satisfy
UID refusal. Both observation/source/tool/binary bindings pass. d605ec overlay
Linux/amd64 vet passes. 66728a checks complete eight-path scope, append-only log,
actual archive hash, static ELF amd64 identities, diff and clean tracked state.

Exact own reviewer project cc-cli-worker-ea72b248f89dd0bc container/network/volume
inventories64619d/43771a/89a6b8 empty. All handles terminal; no tracked edits or
commits. The lost-output limitation remains, and local Compose config is not
misrepresented as daemon lifecycle execution. Root's pass is not the reviewer's.
This approves compatibility research only. Durable job admission, complete job
identity/write authorization, supervisor/daemon lifetime, actual helpers,
supported native host, adoption/recovery/R-006 and the full nine-tier1.0 remain
open. Editorial closeout changes no tool/observation/criterion; no repeated
identical run required. Goal stays active; no push, deployment or archival.


## 2026-11-04 — Bind actual encrypted backup/restore functions to tracked streams

Scope: manual research/tool overlay, no production/CI/normal-test/contract change.
Existing package fixtures/metadata and unchanged CreatePostgresBackup,
RestorePostgresBackup, InspectRecoveryIdentity/CompareRecoveryIdentity are reused.
New multiple-connection gateway fsyncs experimental admission BEFORE native dial,
retains originals through server EOF and rereads exact acknowledged journal rows.
Secret BackendKeyData stays memory-only. RP-752 tracks remaining full binding.

Initial development actual25d5b1/697f01 exits0 with both real populations (2.48s).
After adding fourteen meaningful terminal/framing controls and exact source-commit
binding, development fb5d9e/d39e8a exits0 (2.51s integration). Pre-commit review
then explicitly bounds native dial10s and checks the clock after local dial, SQL,
cancel and child output collection, with4 actual future/past/cancelled/missing
deadline controls. Final b6e37d/1637f4 exits0 (2.68s integration), no skips.
Positive unmodified package functions create/restore an actual encrypted backup;
complete canonical identity and explicit restored row7 equal. Each uses2 streams,
pgx plus actual libpq child, each durably admitted in this experimental file and
observed through real upstream EOF. Existing synthetic release metadata remains
clearly labeled, not an executable release bundle.

Interrupted actual helper shows idle snapshot exporter plus pg_dump blocked on
the controlled table lock. Close admission, kill owned fixture process group,
observe both frontend ends; wrong-key delivery leaves the original query blocked
and full terminal classifier false. Correct key yields ALL original server EOF
BEFORE lock release. Same continuously connected control PID/postmaster retained;
prior row7 unchanged, then control writes42. No backup/output completion report.
Closed listener is exact ECONNREFUSED111. Real closed-journal write fails before
native dial or row acknowledgement. Fourteen actual classifier/framing controls
refuse incomplete/open/error/deadline and malformed/truncated/oversized/duplicate
frames; real EOF/complete population pass. Final5bd1d6 Linux/amd64 overlay vet exits0.
Root62d424 binds retained v1 to final sourceca26d970 and runner42fc4f82; observation
contains neither fixture password nor generated age private key. Source base is
86b82187919846408152a21f22e7d919fe48bf20, kernel0.3.225/epoch8 unchanged.

All root projects520ee9baf95fa415/02e69d2d4d73d952/52300926d0df29c8/
d5c3c5ec2cfb48bc are removed by the runner;
owned container/network/volume inventories empty before completed output.
Build180s/actor240s are observation guards; setup30s, each of2 populations60s,
ordinary handshakes10s, root collection10s after hashes/encoding; Docker cleanup
15s PER command outside work. No production latency/RTO inference. Build inputs
only change through ignored compiler overlay; every command uses root Make.

Limits: SSL disabled fixture only; loopback gateway is not non-bypassable worker
isolation. Process-group signal is test-local, not production lifetime ownership.
Row fsync is not adopted OP1/OP3, parent-directory/power-loss reconstruction or
recoverable supervisor authority. Successful encrypted publication is observed,
but interruption precedes dump completion and does not fence later encryption/
publication. Full backup CLI/outer metrics, release/Engine/HTTP/host writers,
all resources, TLS, daemon identity/bootstrap, supported native host, adoption
and clean-host/R-006 remain. Complete new range from86b82187 needs independent
review; root checks are first-filter only. Next complete those real producer
bindings, not another unchanged two-session positive. Full nine-tier1.0 stays
active; no push, host deployment, archival, release claim or goal pause.

Review by: Codex root, research/tool author first filter. Recorded by: Codex root.


## 2026-11-04 — Actual-helper gateway study independently APPROVED

Review by: independent Codex `/root/backup_operation_gateway_review`.
Recorded by: Codex root, transcribing returned verdict.
Exact complete range:
`86b82187919846408152a21f22e7d919fe48bf20..efc4d402298488ce9dc5ab3c23dba6f9c598e79a`.
Verdict **APPROVED — research/tool scope**, no blocking findings.

Independent ee94bc/55acca exits0: actual encrypted backup/restore identity,
both2-connection populations, wrong/correct-key distinction, original EOFs BEFORE
lock release, preserved continuous control/Postgres/rows7/42 and all controls,
without skips. Private cancellation severing dd9f6f/32439f exits1: substituting
wrong key for correct key fails at observation deadline: INVALID; no completion
from unfinished originals. Tracked sources untouched. Independent783b19 passes
six actual-AST command deadline/error cases and13 observation corruptions;
root/independent source/runner/binary bindings match. Neither output contains
fixture password or age private key.94b5cf Linux/amd64 overlay vet passes;
69cd49 verifies complete7-path scope, append-only log, exact private mutation and
clean tracked checkout. Reviewer made no tracked edits/commits.

Both reviewer projects' container/network/volume inventories empty:
d2cbd0/2e8d18/f2e4d0 and cd7c8b/5cd140/cf1669. All handles terminal. Root's result
is not relabeled as independent evidence. Research approval leaves RP-752 full
producer/profile binding OPEN: full CLI/outer publication, non-bypassable
production isolation, restart authority, TLS, supported host, OP9 adoption and
CR4/R-006 remain. Editorial closeout changes no tool/criterion and needs no
identical repetition. Full nine-tier1.0 stays active; no push/deploy/archive.


## 2026-11-04 — RP-753 correct existing Go/libpq TLS-input mismatch

Outcome: preserve the database URL's explicit CA/client certificate/key paths in
libpq service configuration so inspection and actual backup/restore use the same
TLS inputs. Existing Deployment DP6 DATABASE_URL_FILE contract, not a new TLS
policy or use of unadopted OP9. Production change is bounded to service-file
construction; passwords remain in0600 passfile, paths in0600 service file.
NUL/CR/LF and leading/trailing whitespace paths refuse before credential writes.
Internal spaces work in the actual native consumer. No key content in args/env.

Valid old-source regression25bca5 exits1: Go verified TLS and the generated
client CN, then actual psql fails because /root/.postgresql/root.crt is missing,
ignoring the actual supplied CA path. The original production source is baseline
2a02536f. Corrected740b52/f0375a exits0: Go and actual libpq report matching TLS
and client certificate identity, actual encrypted dump/restore has equal full
canonical database identity, and wrong CA produces explicit certificate verify
failed. A generic transport error cannot count as that refusal. The same test
runs in the existing full declared Postgres lane; no new CI/target/framework.

Setup misses are NOT the regression: initialcda4a0 exits1 because COPY-created
server key follows the0644 umask, so PostgreSQL refuses SSL reload (45b96f logs).
Fixture now uses its fixed owned /tmp path with umask077 from the first byte.
Subsequentf2faa1 exits1 on the fixture's CN display assumption: native PG emits
/CN=cloud_clicker rather than CN=cloud_clicker. Both exact single-CN encodings
are admitted; actual libpq must match the observed Go identity byte-for-byte.
Only25bca5, after those setup controls pass, reproduces the production defect.
Initial unused test imports were caught/corrected in development buildf38fcf.
Generated test CA/server/client keys are confined to declared disposable Docker
service/test tempdirs, never existing secrets. Test's superuser COPY TO PROGRAM
runs only fixed umask/cat on its generated /tmp path; settings reset on every
exit. This does not require a new production database privilege or host command.

Finished batch: cold native packagecc2aa5/2a2e55 exits0, vet211db7 exits0; the
native run can skip DB tests and is not cited for SQL. Complete declared real
Postgres integrationff08c7/045adc exits0, all10 top-level tests/no skips, including TLS,
concurrent migration snapshot, rights HTTP/restore, clean-target and identity
populations. Full output retained ignored .cache/backup-tls-full-integration.log.
Owned projectcc-backup-tls-regression-a04 is removed with volumes after terminal
checks; final container/network/volume933964/e5848a/556e65 inventories empty.
No product/CI/whole-platform
status promoted. New code/tests/docs/records range from2a02536f needs fresh
independent corrective review before treating RP-753 closed. Gateway TLS/full
supervisor/native-host/adoption/retention/R-006 and nine-tier1.0 remain open.

Review by: Codex root, corrective author first filter. Recorded by: Codex root.


## 2026-11-04 — RP-753 TLS corrective range independently APPROVED

Review by: independent Codex `/root/backup_tls_corrective_review`.
Recorded by: Codex root, transcribing returned verdict.
Exact complete range:
`2a02536fc72005b2232d4fbb94bda0380f7d6c29..857569facd50caeb8af4ebc240906aa88884175b`.
Verdict **APPROVED — corrective-code scope**, no blocking findings.

Private exact baseline overlay4dadf9/359e4a exits1 after Go TLS/client identity
passes, with libpq ignoring supplied CA and using missing default root.crt.
Corrected full Postgres8d6c90/6a2cd5 exits0: all10 top-level tests/no skips,
including independently executed matching Go/libpq client identity, actual
encrypted backup/restore equality and explicit wrong-CA refusal. Cold native
package1ec9ae and vetac833a pass. d9f529/df028b check exact private baseline,
complete7-path scope, append-only log, whitespace and clean tracked checkout.
No tracked edits/commits; all reviewer handles terminal. Exact own project
cc-review-backup-tls-b7d4 removed, container/network/volumeba44ca/33feda/390e26
inventories empty. Root's evidence remains author first-filter, not relabeled.

RP-753 closes the bounded real configuration defect. This does not establish
supervisor/gateway TLS, native supported-host deployment, protected metadata,
CR4/R-006 or full release acceptance. Editorial closeout changes no code/test/
criterion; no identical rerun required. Retained candidate df84ce1a predates this
helper correction as well as the already-recorded Account/Guild runtime fixes;
rebuild before any current-bundle behavior claim. Full nine-tier1.0 remains
active. No push, deployment, archival or goal pause.


## 2026-11-04 — PIDfd/socket candidate initial TLS criterion fires (RP-754)

Manual research only; no production/test/CI contract or signaling authority.
Readiness d8c9ea: same PG UID70/CapEff-Prm-Bnd0/NoNewPrivs1 opens self and
postmaster pidfds and checks signal0 permission (no signal), reads FD links/TCP
namespace. Exact readiness fixture removed7cbd88. Current candidate observer
enters only the owned disposable PG PID/network namespaces. It opens pidfd BEFORE
checking full actual accepted socket tuple/inode, UID70 and direct-parent1,
excludes PID1, rejects ambiguity and holds the reference. Go relay supplies its
actual native upstream endpoints BEFORE forwarding any client protocol bytes.

Initial actual3eb05d/901c0e exits1. Existing14 framing/terminal plus4 deadline
controls pass, but the new population stops after successful actual verified-TLS
backup child, at original server read ECONNRESET rather than required EOF.
Root1864fa/96d0ea identifies exact location and two bound native response objects,
PID107/108 and unique socket inodes; both responses signals_sent0. No SIGTERM
was sent by the observer. Restore, interruption and startup populations did NOT
execute and get no credit. This is a fired empirical criterion, not a complete
candidate or a production defect. Source-bound v1 preserves original code and
partial observation; no invalid run is labeled successful.

All owned resources removed by the runner after failure; source work directory
retained for diagnosis. New producer direction explicitly adds independent
retained-pidfd native process exit plus exact original connection end state
(EOF or typed reset) and ended forwarding actors. A reset/error alone remains
unfinished, never refusal or completion. OP5's actual-stream/actor terminal
meaning stays; this proposed experimental kernel producer is NOT an adopted
replacement, and PIDfd alone is not ownership. Preserve this EOF-only negative
when changing the instrument. No TLS decryption or PG key parsing is introduced.

Root source/runner compiled through manual package overlay14e543/ddcf15, no
production inputs changed. Exact Go/helper/observer source hashes are in v1.
The whole original-plus-next-producer range needs independent review; no full
CLI/supervisor, arbitrary tuple/inode reuse proof, job admission authority,
supported Linux/amd64 host, OP9 adoption, recovery or nine-tier1.0 gate closes.

Review by: Codex root, research/tool author first filter. Recorded by: Codex root.


## 2026-11-04 — Stronger opaque-stream + retained-native-exit candidate

Preserve original failed9bf8fdba/v1. The experimental producer now requires
retained ORIGINAL pidfd exit for EVERY registered connection, in addition to
exact original server EOF or ECONNRESET and terminal forwarding actors. It
never reopens a saved integer PID. No reset/error alone closes work; unknown
read errors still invalidate. The observer exposes bounded observe_exit on the
same held native reference. A local forwarding close is recorded distinctly
from peer EOF. This is additional independent native proof, not relabeling
ECONNRESET as EOF or weakening an adopted release/product criterion. OP9 remains
unadopted and incomplete. No TLS decryption or BackendKeyData parsing.

Development actuala3a138/4b5ccd exits0, all3 populations; rootc69a55 records
actual end-state distributions. Add11 actual terminal-method controls, then
final cb4352/5cfaec exits0. Existing14 framing and4 deadline controls also pass,
no skips. Positive full actual encrypted backup/restore identity equal with
verify-full TLS/client certificate. All4 positive streams reset and all4 held
original pidfds observe exit. Interrupted actual helper has2 real bound streams;
wrong binding hash refuses before any signal and actual dump remains blocked.
Correct retained dump reference receives SIGTERM; all2 native processes AND
streams end BEFORE unlocking. Both original streams here are EOF. Startup is
held BEFORE client protocol forwarding: accepted native backend/socket exists
with activity-row count0; killed actual helper and exact retained SIGTERM lead
to native exit and original EOF. Startup is not a completed TLS handshake.

Both terminated old references subsequently give ESRCH3 using signal0. This is
an exited-reference control, NOT forced PID reuse. Exactly7 subjects and2 actual
SIGTERMs; all7 references observe exit before observer stop. Same continuous
control PID and postmaster retained, row7 then control inserts42. Neither killed
helper publishes its completion report. Roota8517b validates final source/runner/
dependency bindings and retains v2, no fixture password/age private key. Final
Go source d62c5fb4, observer5bd33c27, runner1220dc41. Prior child helper remains
ca26d970 and normal TLS fixture2315262e. Native Linux/amd64 overlay vet2b7282
passes. Root-owned a3031788fa22e897/ba5ab7836c2ff1ac fixtures removed, every own
container/network/volume inventory empty before completed output. All handles
terminal; no production/normal test/CI/kernel/balance byte changed.

Limits: trusted same-UID observer shares exact PG PID/network namespaces with
no capabilities and NoNewPrivs1; not an untrusted worker permission. Public
experimental IPC files are not OP1/OP3/authenticated job admission or durable
restart authority. Initial tuple/job authorization, adversarial pre-binding
inode/tuple reuse, daemon/container lifetime, all effects/outer CLI publishers,
full native supported host/permissions and OP9 adoption remain. No production
PID signaling, new host prerequisite, TLS interception, release authority or
CR4/R-006 completion follows. Existing mutation-map/protected-metadata/full
nine-tier1.0 gates remain. Complete e3beb00d..9bf8fdba plus this next producer
range requires fresh independent delegated review; root is first-filter only.

Review by: Codex root, research/tool author first filter. Recorded by: Codex root.


## 2026-11-04 — Independent native study review fires RPC admission bug

Review by: independent Codex `/root/pidfd_backend_research_review`.
Recorded by: Codex root, transcribing returned verdict. Complete union:
- `e3beb00dc55db24ab56432b54bc56b0c7ce60954..9bf8fdbadd71ae87a4f4c05cd16458180db35b86`.
- `9bf8fdbadd71ae87a4f4c05cd16458180db35b86..8272f5863fa5f65bf7ae4b2d18060b5ba1572f0f`.

Verdict **CHANGES REQUIRED — research instrument**, RP-755/P2. Actual private
control d0d69c/23b33f holds nativeRPC mutex11s under parent60, then supplies a
matching reply: call accepts after11.021604s because its local10s begins AFTER
mutex acquisition. Independent e02ea3/0f2837 additionally shows already-cancelled
call writes request-000001.json before refusing. Bound from call entry, refuse
expired/cancelled admission before publication, retain controls/no raised bound.

Independent actual8476a3/b67fac DOES reproduce all3 native populations,7 retained
references,2 SIGTERMs, TLS roundtrip/bystander preservation and controls. Private
no-signal f05842/15edaa exits1 at native deadline, so unfinished work cannot close.
9fa0cb checks historical/current bindings; Linux/amd64 vet1bd53b passes;5ffb9f
checks9-path scope, append-only negatives, exact private mutation and clean tree.
These actual measured runs remain evidence within their executions, not overall
instrument approval. Initial queue-control aad031/92bd33 accidentally selects no
tests through Make dollar expansion; excluded, corrected executed controls above
are the findings. Reviewer made no tracked edits/commits. Both own projects'
container/network/volume2d09b6/e9cd95/d87a4d and694abf/c27048/295397 empty;
all handles terminal. No native-host/admission-authority/adoption/recovery gate.

## 2026-11-04 — Forward-correct RPC queue budget and pre-publication cancellation

Start local10s at call entry. Context-aware TryLock wait can return while the
mutex remains held; deadline checked again after admission, encoding and staging
before atomic publication. Already-cancelled/expired, cancelled-while-queued and
cancelled-while-encoding calls publish no native request. This does not promise
cancellation can unsend an already-published command; unknown work stays invalid.
Six actual admission controls cover those paths, timely reply and native write
failure (not a refusal). Eleven native terminal and unchanged14 framing/4 deadline
controls remain, no guard raised or populations dropped.

Fresh actual98a13e/ba1f57 exits0: full TLS encrypted identity roundtrip, actual
interrupted helper and pre-activity startup complete with all7 retained native
exits, exact4 reset/3 EOF ends,2 real SIGTERMs, wrong-binding refusal, old-reference
ESRCH3 and same continuous Postgres/bystander/rows7 then7/42. No skipped tests;
six new controls execute (queued expiry returns before mutex unlock). Rootf7d661
binds retained v3 to Go a710094f, observer5bd33c27 unchanged, runner82b6fe7c,
prior child ca26d970 and normal TLS fixture2315262e unchanged. No fixture password
or age private key in output. Final Linux/amd64 overlay veteb13f0 passes. Exact
own projectb25edc1998c487a1 removed, all owned container/network/volume inventories
empty before completed output. Original v1/v2/negative verdict remain.

Whole original-plus-native-producer-plus-correction union needs independent
re-review. No production/normal-test/CI/kernel byte changed. RP-755 remains open
until that verdict; RP-754 full native profile/admission/reuse/lifetime/supported
host/OP9 adoption and whole CR4/R-006 remain OPEN. Root is first-filter only.
Full nine-tier1.0 stays active; no push/deploy/archive or goal pause.

Review by: Codex root, corrective research/tool author first filter. Recorded by: Codex root.


## 2026-11-04 — Corrected native ownership study independently APPROVED

Review by: independent Codex `/root/pidfd_backend_research_review`.
Recorded by: Codex root, transcribing final returned verdict. Full union:
- `e3beb00dc55db24ab56432b54bc56b0c7ce60954..9bf8fdbadd71ae87a4f4c05cd16458180db35b86`.
- `9bf8fdbadd71ae87a4f4c05cd16458180db35b86..8272f5863fa5f65bf7ae4b2d18060b5ba1572f0f`.
- `8272f5863fa5f65bf7ae4b2d18060b5ba1572f0f..2e20afde947ba436314717c30ca232a24b2cbd0b`.

Verdict **APPROVED — corrected research/tool scope**, no remaining blocking
findings. Original CHANGES REQUIRED and fired EOF-only v1 remain preserved.
Independent2246cb/bc6565 exits0: all3 real populations and admission controls,
all7 retained references exit,2 SIGTERMs, actual TLS encrypted identity equality
and continuous bystander preservation.428566/888f63 exits0: reviewer's ORIGINAL
11s queue and cancelled/no-publication controls now pass. ca8c63/e57d11 exits0:
independently held timely queue completes after unlock.758e37 verifies v1/v2/v3
source bindings and private-key/password exclusion;2d5ad5 Linux/amd64 overlay vet
passes.6f90a8 checks full10-path union, append-only negatives, unchanged native
observer and clean checkout. Earlier private no-signal failure remains applicable
to unchanged observer/native completion logic; no fresh pass inferred for an
altered mechanism. Root's observations remain author first-filter.

Exact own reviewer project cc-pidfd-backend-6370075e9b10c1e9 removed; container/
network/volume c9036a/5af3d0/618a7c inventories empty. All handles terminal, no
tracked edits/commits. RP-755 closes instrument defect only. RP-754 stays OPEN
for full trusted admission/profile, adversarial pre-binding reuse, durable
lifetime, native supported host, adoption and CR4/R-006. Current prototype uses
Python observer, not a release-packaged Go producer or complete supervisor.
Editorial closeout changes no reviewed source/observation/criterion; no identical
rerun needed. Full nine-tier1.0 stays active; no push, deployment or archival.


## 2026-11-04 — Original native authority and operator-lock carrier

Outcome: establish the missing native handoff primitive before claiming a
packaged supervisor. Manual Go/runner sources and source-bound
`authority-carrier-observation.v1.json` touch no production/normal-test/CI bytes.
Two60s populations use exact disposable Linux namespaces/native volume, no ports,
network, host PID namespace, added capabilities, pulls or daemon restart.

Root e13e58/session80725 retained final objective_completed=true in
`.cache/authority-carrier-research/execution-complete.log`; resumed read f1e16b
confirms both completed arms after the terminal session expired. Linux/amd64 Go
binary on Desktop Engine28.4.0 Linux/arm64. SCM_CREDENTIALS identifies UID0 writer
versus SO_PEERCRED's UID65532 listener creator. A retained writer pidfd, original
connected stream and actual AcquireOperatorStateLock file pass via SCM_RIGHTS to
a separate-PID-namespace UID65532 supervisor. After launcher exit the original
stream completes probes3/4 and actual competing lock returns EAGAIN11. Native
writer exit plus original EOF precede a same-path/same-ID live replacement;
supervisor refuses authority_lost, replacement sees only its independent control,
and the real lock stays held. Fixture-only supervisor finish permits actual new
lock acquisition. Missing third FD returns missing_lock_fd before any supervisor
probe (original server count remains2), then releases the original launcher lock.
Kernel CapEff/Prm/Bnd0 and NoNewPrivs1 checked; exact own resources removed and
all container/network/volume inventories empty before completed output.

Development failures remain failures: initial Mac bind-mount Unix-socket chmod
returned EINVAL (da514e/2fb1fb; preceding runs hid stderr); switching fixture IPC
to a native Linux volume allowed transfer.18719b/0f44c0 reached lock/loss checks
but failed at finish because the controller ran inside the namespace it asked to
exit. Moving that controller to the owned anchor produced the full two-arm run.
Neither partial population is completion evidence or a production defect.
Root vet21c858 initially included unrelated normal main_test.go against the
replacement research main and failed undefined scheduleRound; corrected exact
file selector61688b passes Linux/amd64 vet. No bounds raised/assertions relaxed.

Reproduce from root: `python3 planning/deployment-foundation/probes/authority-carrier.py.txt`.
Vet: `make vet GO_PACKAGES=./cmd/deployment-backup/main.go GOFLAGS=-overlay=$PWD/.cache/authority-carrier-research/overlay.json GOOS=linux GOARCH=amd64 CGO_ENABLED=0`.
Observation binds source ddd92e1b, runner02c59114, binary bdbbec59 and actual lock
source2d701b2a; build180s/handshake10s/cleanup15s per command/final collection10s.

RP-756 stays OPEN full production binding. Fixture metadata is not fsynced OP1/
OP3 authority, UID equality is not authenticated bootstrap, and model protocol
is not actual Docker daemon HTTP/lifetime or native supported-host evidence.
Next construct the complete supervisor profile/admission/receipt proposal over
all eight mutation kinds, using this carrier and reviewed SQL observer; retain
all adoption, clean-host/R-006 and full nine-tier1.0 gates. Independent research/
tool review pending, no specification adoption or feature acceptance inferred.
No push/deploy/archive or goal pause.

Review by: Codex root, research/tool author first filter. Recorded by: Codex root.


## 2026-11-04 — Native authority-carrier model independently APPROVED

Review by: fresh independent Codex `/root/authority_carrier_research_review`.
Recorded by: Codex root, transcribing returned final verdict.
Exact full six-path range:
`0f3a9a7c4dac96e7b3c11269cb2c08961d75ed75..695540b7e1121c880f67c7dc989effb20a684dee`.

Verdict APPROVED research/tool scope, no blocking findings. Actual independent
runner a2e485/751b84 exits0 with both populations complete. Original stream and
real lock survive launcher exit; native writer exit/EOF precede replacement
refusal, and missing-lock refusal/reacquisition pass. Private copied-source
51c41c/357ac0 closes the transferred lock after launcher exit: unchanged real
lock-contention assertion fails, exit1/objective_completed=false; no tracked
source mutated.044eff Linux/amd64 exact-file overlay vet passes.7b7111 checks
committed source/runner/lock hashes against observation;547b0e whitespace passes.
Exact owned successful projects cc-authority-carrier-9a554e86e6ebd331 and
cc-authority-carrier-713180aab92cbac2 and mutation project
cc-authority-carrier-af51016e3aed4998 have empty container (including exited),
network and volume inventories. All execution handles terminal. Initial sandbox
Docker-socket denial created no fixture and supplies no completion evidence.
Private mutation remains /private/tmp/cloud-clicker-authority-review-22w0if5q.

RP-756 stays OPEN full production binding. Actual daemon protocol/authentication,
bootstrap durability, complete native supervisor/profile, OP9 adoption, supported
host, whole recovery/R-006 and release approval remain separate. Editorial
closeout changes no source/criterion/observation, so no identical software rerun.
Full nine-tier1.0 goal remains active; no push/deployment/archival or goal pause.


## 2026-11-04 — Whole native supervisor profile proposed, not another mechanism census

Previous goal turn made authoritative progress: carrier model committed and
independently approved through695540b7, recordedc58f8e48. Current clean checkout
190b21 confirms that baseline. This batch constructs the missing OP9 proposal:
`rfc/deployment-native-supervisor-profile.md`, links in OP9/index and current plan,
and RP-757. No production/test/CI/kernel/balance byte changes.

SP1–SP8 define one-shot native operator-owned bootstrap (no new resident service),
original daemon/lock handles, private attachment, broker/gateway containment,
all eight typed work subjects/native results, finite stage/action families,
complete retained inventory and install/release/rollback/abort acceptance. It
explicitly includes scheduler freeze/handle export/restart/cache, Postgres image
changes, dedicated startup SQL and publishers/deletions. Linux/cgroup-v2/native
features, Docker/Compose artifact provenance and host floor are proposed extra
bindings, not inferred support from the Desktop studies. Operator capabilities
and secrets are not expanded silently; native feature tests remain mandatory.

Source outcomes change the design: actual command output a3b6b4 shows the release
metric means latest SUCCESS/FAILURE, not verification-in-progress. Existing outer
runRelease/runInstall metric follows the controller; pre-publishing success
before Finished would make a failed final append look successful. The proposed
existing-service ledger projector removes that writer and follows only verified
final decisions, with an explicit change for unadmitted CLI failures and a missing/
stale-projector alert obligation. This is NOT implemented or adopted. Current
16e795/9d6c1e source confirms Start excludes Postgres; target-image change gets an
explicit post-checkpoint stop/start path, not an unchanged-call completeness claim.
87ad12 confirms scheduled rounds also remove temporaries/apply retention; controller
freeze/export covers those effects, not only dump sessions. Current startup uses
save.Migrate in gameserver/composition.go; operation startup SQL cannot be silently
retained as a steady pool to escape OP5's terminal requirement.

Verification: inspected current contracts/call graph/template/textfile producer,
complete eight-kind subject/result mapping and before/after finalization order.
Existing carrier/SQL observations are reused only for unchanged mechanisms,
not actual profile conformance. No software suite for a normative Markdown draft;
new publishers, routing, cgroup authority, packaged Go producers and full journey
have not executed and remain acceptance/adoption prerequisites. Designated
technical review pending for this whole proposal range. RP-754/756/757 stay OPEN;
no owner adoption, ready-to-ship profile, full recovery or1.0 promotion. No
push/deployment/archival or goal pause.

Review by: Codex root, specification author first filter. Recorded by: Codex root.


## 2026-11-04 — Native profile CHANGES REQUIRED; forward ordering/identity corrections

Review by: independent Codex `/root/native_supervisor_profile_review`.
Recorded by: Codex root, transcribing the returned final verdict.
Exact six-path reviewed range:
`c58f8e4854e13edff24cbb79bd95e46bbc085284..61f2461f70cee861fb3928d508f5d0db8fcacdbd`.
Verdict CHANGES REQUIRED. Root's earlier "complete" language exceeded the actual
specification; corrected in current plan/index/body, original log preserved.

Four substantive findings: (1) no lawful pre-Started route for required CR2 source
observation or CR4 authenticated backup validation/scratch, and no first-worker
image-acquisition route; (2) once-bound Container incorrectly changes null start/
process to actual started identity; (3) SP5 action/family prose is not the exact
stage/actor/action/kind population, request-body/terminal classifiers and cross-UID
IPC adapters demanded by OP9; (4) one-selected-manifest producer rule wrongly
applied to both source/target runtime. Reviewer inspected complete proposal,
OP/CR/DP contracts, actual controller/runtime/CLIs/backup/native/textfile/template/
startup lifecycle; exact-range diff whitespace passed, clean endpoint verified.
No software suite warranted or profile/runtime conformance claimed. Projector
is a plausible ordering-cycle resolution, not a tested/adopted implementation.

Forward author correction defines inert native admission observation plus current
in-process protected-input validation before Started, with private pending marker/
final admission object and no fake execution phase or live mutation. Current
VerifyRestoreInputs ignores context: this is explicitly not a proved guard;
invalid/unfinished work keeps original authority/lock. After OP7 admission or
Started, native supervisor acquires worker image directly, avoiding worker-before-
image circularity. Container becomes immutable creation identity; separate Start
and native observations describe invocation without second subject_bound.
Producer toolchain alone has one selected manifest; each source/target runtime/
helper has its own authorized identity/compatibility. IPC uses a private mounted
leaf under0700 parent/native peer-generation checks, not UID/DAC wishful access.
Source scheduler resume is explicitly conditioned on preserved-prior completed
pre-backup failure; post-checkpoint failure cannot take that route.

These corrections are author first-filter, not independent resolution. Finding3
stays OPEN: exact adapter/plan population, native Engine/HTTP handlers and their
terminal classifiers, complete IPC wire schemas and current service restart/start
routes still need construction/review. Current bootstrap observation inventory
also needs those exact schemas; it is not an empty-registry implementation route.
Next deliver that concrete specification table, not another positive mechanism
census or premature adoption request. RP-754/756/757, actual packaged Go/native
producer conformance, supported host and clean-host R-006 remain OPEN. No product,
test/CI/kernel/balance bytes; no software rerun for normative Markdown. Diff checked.
No push/deployment/archival or goal pause; full nine-tier1.0 remains active.


## 2026-11-04 — Concrete native adapter proposal and fired delivery-attribution diagnostic

Previous goal turn made progress through61f2461f/9c4f1b18, preserving the
independent CHANGES REQUIRED verdict. The9c4f1b18 checkout was clean. This
batch supplies `rfc/deployment-native-supervisor-adapters.md` A0–A7 and reconciles
SP/OP9/index/current plan; no production, normal-test, CI or kernel/balance bytes.

The table is an exact proposed finite expansion with E00–E21, command/lifecycle,
SQL/HTTP/file/native, cross-UID IPC/cache and bootstrap schemas. Native Go is the
sole Docker mutation dispatcher; actual pinned Compose model compilation remains
inert/checked, workers get no mutation channel. Explicit source/target producer
bindings, create-versus-start, v1 install grants, source ordinary-work origin
versus new cleanup admission, cache receipt DAG and projection ownership are
spelled out. Native process/invocation completion does not erase independent
external children. Private job scope uses bound PID namespace INIT death plus
restart prevention (not generic parent exit); shared-PG observers/source routines
have explicitly distinct trusted-code contracts. These replace the unproved
cgroup-v2 zero-populated assumption, not whole-scope proof. Native conformance,
source/compatibility review and original full SP8/R-006 gates remain mandatory.
Pre-admission E20 is not a fabricated fresh device/inode stat: combine admitted
non-reused physical generation with current Engine metadata, then physically
recheck before mount/reset/delete after Started. Unknown/unsupervised cases refuse.

Primary sources checked: pinned Moby28.4.0 API1.51 and container/image/network/
volume handlers, Go serial HTTP connection loop, Linux PID namespace manual.
These justify proposed wire/inference boundaries only, not runtime evidence.
Native terminal requires full original upload/response plus same-connection
ordered reply or actual server EOF, with children separate; no generic status/
transport error becomes proof. Closed body construction, pure-Go observer trust,
private application completion and exact source export still require independent
challenge and actual packaged implementation. No "table passes => feature works".

RP-758 now has EXECUTED evidence. Manual probe calls unchanged actual
operations.VerifyAlertDelivery with modeled manager endpoints and actual local
webhook deliveries. Final33a7fa/3364bd: matching_positive PASS (actual received
nonce equals returned); unrelated_negative FAILS as required by the diagnostic:
helper returns nil/32-character nonce with submitted1/unrelated1/MATCHING0.
Make exit2 / Go exit1, not green. Both receiver health and notification are the
actual local receiver. Context3s per population; actual total0.21s. No actual
Alertmanager, Compose, external receiver, credentials or native release proof.
Source-bound observation retains a673816c helper and891a43aa probe hashes.
Initial94cdad/0b96c0 ran ZERO tests because Make consumed the dollar-end selector;
excluded. Corrected a9fcff/c0813e first reproduced the negative; final adds positive
control and exact receiver health. New exact H06 observation/manual checkpoint is
PROPOSED and needs DP7 adoption, not a silently implemented fix/new provider API.

Reproduce by mapping virtual server/operations/alert_delivery_attribution_probe_test.go
to the manual .txt via a temporary Go compiler overlay, then root
`make test-go GO_PACKAGES=./operations GO_TEST_FLAGS='-count=1 -run TestDiagnosticVerifyAlertDeliveryRejectsUnrelatedNotification -v' GOFLAGS=-overlay=<absolute-overlay.json>`.
Final overlay vetf9a2a2 exits0. All execution handles terminal; httptest servers
closed, no Docker/global/native-host resource created. Documentation expansion
sanity d08e0f checks unique plan IDs/E00–E21/worker-only negotiation; it is not
producer/runtime proof. Whole proposal + diagnostic range needs independent
technical/research re-review. Original negative verdict remains; RP-754/756/757/
758 and host/producer/adoption/R-006 remain OPEN. Full nine-tier1.0 remains active;
no publication, deployment, archival, goal pause or release promotion.

Review by: Codex root, proposal/probe author first filter. Recorded by: Codex root.

Final author consistency pass makes ciphertext publication/deletion trusted
in-supervisor leaves (no cross-UID output-write assumption), closes authorized
secret-FD/private libpq-file handling, and separates plaintext scratch from
encrypted-only protected staging. Failed publication cannot fabricate backup
success or discard an unrecorded backup. Source cache restart keeps incomplete
round markers paused; source/target recovery ownership remains explicit. These
are proposed producer contracts, not executed native conformance.

## 2026-11-04 — Adapter re-review: four executable-path omissions, forward correction

CHANGES REQUIRED, complete nine-path union
c58f8e4854e13edff24cbb79bd95e46bbc085284..
224fc0e2e17f8d8993f0971ade6fb5fc9903f66d.
Review by: independent Codex /root/native_supervisor_profile_review.
Recorded by: Codex root, transcribing its verdict. Unrelated Headcount editorial
4118c4d8 is excluded. Original admission/image bootstrap, immutable creation and
producer/runtime distinction are substantively corrected, but workers lack a
concrete FD/control route, new-volume inspection has a binding cycle, SP5's closed
child graph rejects new invocation paths/E07 prematurely demands Start, and
fencing cannot admit I03's required cache publisher. No technical approval.

Reviewer independently executes the exact overlay diagnostic: matching_positive
PASS, unrelated_negative FAIL with nil helper error/nonce32, actual unrelated1/
matching0; Go1/Make2,0.21s. Overlay operations vet passes, both retained hashes
match, complete range whitespace passes. This independently confirms RP-758's
bounded helper defect, not actual Alertmanager/external delivery. No tracked
reviewer changes or native-host/producer/feature/release proof.

Forward author correction adds W00's per-job read-only mounted private socket,
exact handshake/FD order, host-view kernel credential/pidfd/namespace mapping,
blocked child gate before secret/command release, exact output/report/status
messages and physical scope closure. It is separate from service state and
assumes neither inherited Docker-create FDs nor host ptrace/DAC bypass. V00 adds
immutable CreatedVolume, a read-only/no-copy first-inspector exception and native
root-FD report before complete Storage binding. Private scratch initialization
is UID0/cap-drop chmod of its new empty root only, then receiving-UID0700 work;
service image population/initialization has explicit observed ownership history.
No arbitrary chown/privilege or application use of provisional storage.

SP5 now includes the exact bounded writer/start/cache/private-scratch and pending
first-inspection child edges; no reciprocal receipts. E07 binds created children,
reserves future Start/native children, then binds those on actual native return.
Both fencing cache parent/leaf rows exist before I03. These are author first-filter
corrections requiring complete original-plus-correction re-review. All native
producer/compatibility/adoption/SP8/R-006 gates stay OPEN. No software/CI/balance/
copy bytes or goal pause; no push/deployment/archival. Diff checked; no unrelated
software suite for normative Markdown. Research negative and both verdicts remain.

## 2026-11-04 — Worker transport re-review and final forward correction

CHANGES REQUIRED, reviewed native union c58f8e48..224fc0e2 plus
4118c4d8dfd48be110f99c21f79e068b9004ef07..
34a60697b9a257396cc53ad232d9d308e9d09412; Headcount editorial excluded.
Review by: independent Codex /root/native_supervisor_profile_review.
Recorded by: Codex root. Prior volume bootstrap/closed child edges/E07 timing/
fencing cache findings are substantively corrected. W00 still incorrectly makes
the shared-Postgres Go leaf PID1, omits leaf traversal permissions, and closes the
only possible child result-FD channel on exec. V00 also ambiguously calls the
initial transition itself null. No technical approval; original verdicts remain.

Forward correction names operator-owned0711 job leaf/0666 socket beneath0700
host parent, read-only worker mount and real cross-UID positive/negative controls.
Hello is mode-discriminated: private init PID1 versus fixed non-PID1 Go observer
in the admitted PG namespace, carrying native self/namespace and original PG-init
pidfds; native host/CID/peer mapping is mandatory. No host ptrace bypass.
Only backup/mount-inspection children receive a distinct local SEQPACKET result
pair, with exact work/kind/one-FD messages and native child-writer validation.
The original PID1 shim forwards AFTER actual child/IO exit; no supervisor control
FD survives command exec. Shared Go leaf stays fixed in-process/no-fork. The
initial observation keeps transition=initial and precisely names its two null
keys. These are author first-filter corrections, not runtime evidence.

Reviewer checked both exact-range whitespace and unchanged production/probe/
observation bytes; preceding independently executed diagnostic/vet remain applicable.
No redundant software rerun or new runtime/CI claim. Next complete corrected-union
independent re-review; then actual producer/native conformance and explicit
adoption, never table-only acceptance. RP-754/756/757/758 and R-006 remain OPEN;
no host/secret/publication/deployment/archival action or goal pause.

## 2026-11-04 — Complete native adapter proposal technically approved

APPROVED, technical specification and bounded research-tool scope only.
Review by: independent Codex /root/native_supervisor_profile_review.
Recorded by: Codex root, transcribing its completed verdict. Exact complete union:
c58f8e4854e13edff24cbb79bd95e46bbc085284..
224fc0e2e17f8d8993f0971ade6fb5fc9903f66d plus
4118c4d8dfd48be110f99c21f79e068b9004ef07..
a836e92b1dc7840b49e28f48345855bc0d46735f.
Unrelated Headcount editorial excluded; all earlier negative verdicts retained.

Reviewer inspects complete construction/callgraph/OP/CR/DP: prior admission,
immutable binding, producer identity, volume bootstrap, closed child graph,
fencing cache and W00 shared/private modes, cross-UID leaf and post-exec result
transport are resolved. No further blocking positive/error-path contradiction.
Both exact-range whitespace checks pass; checkout clean. Unchanged actual alert
helper/probe/artifact preserve independently executed matching PASS/unrelated
negative FAIL and overlay vet PASS; no redundant suite or new runtime evidence.

Approval is technical coherence, not native producer conformance, actual Docker/
namespace/FD behavior, packaging, adoption, implementation authority, feature/
R-006 acceptance, archival or release. RP-757 remains open through architecture/
parent/host/artifact adoption and construction; RP-758 remains an unfixed bounded
helper defect. Next prepare exact tool provenance/rights/artifact inventory and
DP7 projection/reboot-cache adoption evidence required by SP1/SP6, then obtain
explicit architecture/host-floor adoption and reconcile parents before production
implementation. Supported native host and full SP8/R-006 stay required; no
version/CI/gameplay/copy change, publication, deployment or goal pause.

## 2026-11-04 — Native tool provenance/rights audit: partial SBOM is not clearance

Previous goal turn made concrete progress throughad1d23c1: complete corrected
native profile technical approval, no adoption/runtime claim. Current clean
checkout revalidated. This batch audits SP1's actual cached research-only CLI/
Compose fixtures; no product/test/CI/kernel/balance/copy bytes.

Actual binary inspection finds CLI28.4.0 Go1.24.7/CGO1/pkcs11 and ZERO dependency
module rows, not zero dependencies. Compose2.39.4 Go1.24.7/CGO0 has180 module
rows. Fresh publisher release metadata and actual SBOM/provenance downloads match
binary/statement hashes and subjects. All180 exact linked name/version rows are
in the183-package SPDX2.3 graph, but182 declared licenses/all183 concluded licenses
are NOASSERTION. Top-level Apache LICENSE/NOTICE files are captured separately;
no dependency terms/native link closure is inferred from them. CLI's local archive
hash lacks an independently checked publisher checksum; its unsigned tag only
corroborates embedded short revision. Compose full native revision matches its
observed tag. Publisher metadata is not verified signed provenance/rebuilt source.

Retain native-tool-provenance-audit.md, source-bound v1 observation and the manual
probe. Finalc94da8 exits0: INSPECTION completed, distribution_approved=false and
native_dependency_closure_known=false. Actual-Go/ELF inspection is the population,
not a mocked version fixture or release validator. Private416630 controls rebind
the forged SBOM's release-asset hash, then wrong subject/missing actually linked
compose-go still refuse; truncated two-tool metadata refuses. Shared binary inputs
remain unchanged, private copies removed. Finalb01f15 existing-output attempt
exits1 and preserves prior bytes. Probe/source hashes match. Existing application
generator covers six own Go commands/browser, not these external binaries.

GitHub connector asset URLs returned404; direct read-only curl inputs completed
with exact publisher-file hashes. Optional build-script connector query in exec985
remained live and was explicitly terminated without completion evidence; no
missing-file conclusion. Required Dockerfile/metadata were separately downloaded
by bounded completed curl calls. No execution/session/resource remains live.

RP-759 is OPEN packaging/adoption evidence, not a shipped-tool defect or license
clearance. Next establish actual vendor/Go/native closure and corresponding
notice/provenance delivery; prepare exact tool artifact/SBOM inputs, then obtain
applicable adoption before runtime/package changes. DP7 projector/cache evidence
and supported host/SP8/R-006 remain separate gates. Diff checked; no unrelated
software suite, source build, Docker/host operation, publication or goal pause.
This whole audit/tool range requires independent research review.
Review by: Codex root, implementer first filter. Recorded by: Codex root.

## 2026-11-04 — Native tool audit independently approved, rights closure still open

APPROVED, bounded research/tool scope, exact six-path
ad1d23c118746600451e651047ebd46692cf09c8..
3fd7483c7ae8c3ed2743b59da5be6c2ae059afe6.
Review by: independent Codex /root/native_supervisor_profile_review.
Recorded by: Codex root, transcribing its completed verdict. Actual probe exits0;
stable facts and every retained input/probe hash match. Independent rebounded
wrong-subject/missing-compose-go controls, truncated-tool metadata and existing-
output refusal pass with no new output/unchanged prior bytes. All shared inputs
stay hash-identical. Actual license/notice/revision and existing six-command/
browser generator boundaries checked; range whitespace passes and checkout clean.

No distribution/rights clearance, signed/reproduced provenance, native closure,
shipping-tool selection, profile adoption/runtime/R-006/release approval. RP-759
stays open. Next collect every exact declared module source by its binary h1
identity and corresponding scoped license/NOTICE bytes, then establish CLI's
Go/vendor/native link closure. Read-only cache census found59 exact Compose source
directories and121 missing; it is availability only, not authenticated source or
partial release clearance. The full180-module population remains the required
collection scope; no package/term may disappear to make an inventory pass.
No product/CI/balance/copy change, publication or goal pause.

## 2026-11-04 — All180 Compose source identities collected; actual detector refuses four MPL roots

Previous goal turn's provenance audit was completed/independently approved through
4af4fe6b; clean current state revalidated. This batch advances RP-759 with actual
source, not inferred license labels. Full population180 from the reviewed binary
inventory; root/nested notice paths remain distinct. No production, ordinary-test,
CI, dependency lock, kernel/balance or authored-copy bytes changed.

Initial1a8bd8 v1 exits1 after59 verified/121 URLError; no guard exhaustion or
exclusions. Diagnostic8e0d34 confirms Python DNS failure; direct curl2d4093 works.
Preserve exact v1 source/hash and failed artifact. Separate ordinary root curl
calls (exec1005, all121 successful/terminal) stage missing archives; current
collector is offline-only. Final7e05f9/28740→b04e35 exits0: all180 full source
archives match their binary h1;59 cached/121 staged;106,533,767 archive bytes.
Collect237 named notice files/1,487,398bytes (200 root,37 nested); no missing
root filename. Source collection is complete, license applicability is not.
No source code executed/installed and no module/compiler cache setting changed.

Actual Hash1 corresponds to primary Go x/mod/dirhash;500MiB compressed/expanded
safety limits follow Go module zip contract. Original600s collection guard kept.
Final0ae26d private controls refuse changed full source, duplicate entry, expired
guard, corrupt existing notice cache, incomplete179 population and preexisting
output; originals unchanged, private control directories removed. Earlierbc65ec
only completes source-corruption refusal: its duplicate builder reused mutable
ZipInfo and hit an invalid CRC fixture. That remainder is not completion evidence;
copying ZipInfo produces the six final discriminating controls above.

Manual Go overlay invokes ACTUAL existing licenseAt/detector against every
source-bound root. Final1b60c5/74430→738e61 executes180, recognizes176/refuses4,
Go1/Make2; retained report binds full v2 bytes. All4 are exact Hashicorp MPL2.0
roots: errwrap1.1.0, go-cleanhttp0.5.2, go-multierror1.1.1, go-version1.7.0.
No false all-clear or policy widening. First424d80 invocation ran ZERO tests
because Make places GO_TEST_FLAGS before package selection and -args consumed
that package; excluded. Required wrapper uses GO_PACKAGES= and places the explicit
package BEFORE -args in GO_TEST_FLAGS; ordinary root target/owned GOCACHE retained.

Independent-of-classification corruption controls16bf51/49869→9216da andca562f/
15245→04c2d1 fail at duplicate-index and exact notice-byte checks, respectively;
neither publishes a report. Overlay vet0a6d5e passes. Existing cold positive ISC/
0BSD tests54f975/53262→b6239a pass, alongside actual176 recognized roots. All
handles terminal; private verified corpus/overlay retained for reproduction.
Classifier/probe/source/artifact hashes checked; original helper bytes unchanged.

Primary MPL2.0 §3 and Mozilla FAQ Q7–Q8 were read; exact-source availability and
recipient notices need a concrete tool-specific delivery treatment. No general
legal clearance, signed/reproduced build, vendor-patch correspondence,37 nested
scopes, main-source/stdlib/CLI-native closure, shipping selection or distribution
permission follows. Next review this complete source/tool/diagnostic range, then
finish those exact remaining scopes and proposal. RP-759/adoption/host/SP8/R-006
stay OPEN. No product/CI/architecture adoption, publication/deployment/archival or
goal pause. Diff checked; no unrelated software suite.
Review by: Codex root, implementer first filter. Recorded by: Codex root.

## 2026-11-04 — Complete Compose source/diagnostic batch independently approved

APPROVED, bounded research/tool scope, exact ten-path
4af4fe6b95044b263c34d4d3ba81fe50c690c856..
9c90bc2efeaa8dd554a33f21a491464843996ce0.
Review by: independent Codex /root/native_supervisor_profile_review.
Recorded by: Codex root, transcribing the completed verdict. Independent full
collector reproduces every stable180/237 h1/archive/notice fact and both source
bindings; v1 stays incomplete. All six private corruption/guard/population/cache/
overwrite controls refuse as intended. Final population/elapsed checks inspected.
Own verified200-root-file corpus and actual Go overlay produce byte-equivalent
classification JSON:176 recognized/4 exact MPL2.0 refusals, Go1/Make2. Duplicate
index/changed notice refuse before publication; overlay vet and ISC/0BSD controls
pass. Exact range whitespace passes and checkout clean. No tracked reviewer edits.

No complete rights clearance, vendor-patch/main/stdlib/CLI-native/nested-scope
closure, shipping selection, architecture adoption or release proof. RP-759
remains OPEN. Source follow-up is reviewed; next complete those real scopes and
exact source/notice delivery. Separate read-only follow-up notices the existing
application generator's root NOTICE omission; confirm against an actually linked
dependency before treating it as a corrective defect. No policy widening,
product change, publication or goal pause follows from this research verdict.

## 2026-11-04 — Correct application NOTICE omission under existing DP1/AC8

RP-760 is an ACTUAL application metadata defect, separate from proposed tool
shipping/rights policy. Existing generator licenseAt reads LICENSE/COPYING only;
actual linked prometheus/client_golang1.23.2 has a required separate NOTICE absent
from its ThirdPartyNotices output. New ordinary regressionf78381/38967→0a6839
fails old behavior at exactly that omission; actual module identity, complete
six-command dependency discovery and actual output formatter are used.

Implement separate Dependency.NoticeText, root NOTICE/NOTICE.* collection for
Go/selected stdlib/browser physical sources, delivery beside license text, and
notice-aware browser-copy deduplication. License classification/validation still
uses only existing license text/allowlist; a NOTICE cannot supply a missing grant
or reclassify an Apache mention. Invalid present notice (empty/directory/symlink)
refuses, absent optional notice remains valid. No gameplay/schema/dependency/
kernel/version/CI or MPL-policy change; canonical Deployment docs updated.

Real regression nowpassesf500b9/13427→89ad42. Full affected cold package run
f7cd71/68702→ba1e6d passes ./releasepackage and ./cmd/gen-release-metadata;
vetb14041 passes. Tests cover extra separate notices, license independence,
invalid/absent notice and JS/CSS equal/missing/conflicting physical NOTICE cases.
This advances accepted attribution delivery, not full rights or supply-chain
clearance. Next execute actual generator from committed source with the current
built browser inputs, then independently review the coherent correction range.
RP-760 remains review-pending, not closed by this first filter. No publication,
deployment, archival, owner adoption, release promotion or goal pause.
Review by: Codex root, implementer first filter. Recorded by: Codex root.

## 2026-11-04 — Committed-source application metadata generation confirms NOTICE delivery

At clean committed4f17a1f84c8106ad6e6f223a652a2e7dd36802eb, root
make generate-release-metadataa31442/20519→43d0f1 completes53 dependencies with
existing built-client inputs. Explicit diagnostic version0.0.0-notice-check,
commit4f17a1f8, created2026-11-04T13:31:19Z; this is metadata generation, not a
release bundle/build/rehearsal. Owned output directory is recorded locally in
.cache/worker-cli-tools/notice-metadata-output.txt for reviewer reproduction.

Actual output2a7640 contains full selected Prometheus NOTICE in that exact
module's section, with SPDX declared/concluded Apache2.0 unchanged. SPDX2.3 has
53 dependency packages plus application. Output identities:
sbom.spdx.json41,615bytes SHA256b8a4b05a251f172d5fe45906cfdabbdbfa941680aebe1bc75554ef52ef6a0a3f;
third-party-licenses.txt155,397bytes SHA2567d4186dee61c36cf63d6de4e11afe205063f3af94aa97cae0271a67f98133769.
No external tool/MPL clearance or exact release/candidate proof inferred.
Next independent corrective review of full9cde3a64..this checkpoint, including
old-behavior discrimination and all code/tests/docs/records. RP-760 remains OPEN
until that verdict; other rights/source/native/adoption/R-006 obligations remain.

NOTICE-only control984d6a/89119→7dc759 additionally passes: even valid MIT grant
text in NOTICE cannot supply a missing LICENSE. All focused notice cases execute
cold; final plan wording cleaned up. Production bytes remain those exercised by
full package/vet and committed-source generation; no redundant full rebuild.

## 2026-11-04 — Root NOTICE correction independently approved

APPROVED, scoped corrective code under existing DP1/AC8. Exact complete eight-path
9cde3a64ea2b356473de0ac467a97d02e6d56f2b..
0c746ca13d5e4cf56bf438edd39a1442f76908d5.
Review by: fresh independent Codex /root/notice_delivery_corrective_review.
Recorded by: Codex root, transcribing the completed verdict. No actionable finding.

Reviewer executes cold full affected packages/vet, current actual Prometheus
regression, private old-source compiler overlay and fresh root metadata generator.
Old source FAILS at exactly required NOTICE omission; corrected source passes.
An initial Make-consumed anchored selector executed no tests and is discarded.
Actual53-dependency output retains full target NOTICE with unchanged Apache2.0
SPDX and byte-identical third-party attribution SHA256
7d4186dee61c36cf63d6de4e11afe205063f3af94aa97cae0271a67f98133769.
Root NOTICE/optional-invalid cases, JS/CSS physical-copy agreement, missing-license
refusal, selected stdlib directory and unchanged no-NOTICE formatting inspected.
Gofmt/range whitespace clean; no tracked reviewer edits. Private review evidence
is retained under /private/tmp/cloud-clicker-notice-review-ung5q6d1/.

RP-760 closes this actual root-NOTICE defect. Full nested attribution, MPL/tool
redistribution/source-rights closure, architecture adoption, native-host/R-006
and whole1.0 remain OPEN. No release acceptance, publication/deployment/archival
or goal pause follows. Next resume those concrete source/delivery scopes under
RP-759 and the reviewed native-profile prerequisites; do not reopen the corrected
root NOTICE omission or infer clearance from its closure.

## 2026-11-04 — Nested source/binary observation finds another actual attribution omission

Previous turn closes RP-760 under independent corrective review through1cc6762a;
clean state revalidated. New manual Go overlays only; no production/normal-test/
CI/kernel/balance/copy bytes in this research batch. All inputs are existing
public source corpus and pinned actual Compose binary.

Actual267f24/54722→d7a581 parses90,355 Go functions from the actual ELF's code/
function table, checks input SHA and all180 module identities on the SAME byte
slice, then observes all37 nested candidates.12 have positive function-prefix
matches;25 have no such evidence and remain explicitly UNKNOWN, never absent or
excluded. Samples limited5 with total/truncation fields. A main-package positive
and unrelated-prefix zero control run. Go vet673772 and full-overlay173c27 pass.
Private wrong-binary8308dc/10870→e26e02 and wrong-textf68a44/84498→fad2d8 refuse
before any report; originals unchanged. All handles terminal; private fixtures
retained under the path in .cache/worker-cli-tools/nested-scopes-work.txt.

Buildx bake/hclparser's nested LICENSE contains MPL2.0, but zero function evidence
here cannot rule out inline/data contribution. Root's4 MPL refusals are therefore
not claimed exhaustive. The37 filename candidates also include workflow/code;
no bare filename or detector label becomes legal applicability. No distribution,
source/build correspondence, signature, policy adoption or release clearance.

Actual application diagnostic6d256a/68527→615980 executes the existing six-command
Go import graph and metadata producer: imported client_golangv1.23.2 internal
GDDO httputil lacks its nested BSD/license copyright in delivered attribution.
Go1/Make2 at the exact missing-text assertion. Earlier023cbb/76862→d2c7d3 was an
INVALID probe: it passed the unsorted discoverGoDependencies slice directly to
ThirdPartyNotices. Corrected probe applies the actual SortDependencies boundary;
only the final exact omission is defect evidence. RP-761 records the finding.
Source-read-only import/text inspection corroborates it; RP-760 is not reopened.

Next independent research/tool review, then repair actual linked Go scope
attribution under existing DP1/AC8. Keep root-NOTICE's approved behavior; do not
use all unknown tool scopes as a reason to stall an accepted corrective fix.
Full tool/native/rights/architecture adoption and R-006/whole1.0 gates remain.
Diff checked; no unrelated suite, host/secret/Docker/publication or goal pause.
Review by: Codex root, implementer first filter. Recorded by: Codex root.

## 2026-11-04 — Nested observation and actual omission independently approved

APPROVED, bounded research/tool scope, exact seven-path
1cc6762a74978f25c1c9782309eb9aecd501ef87..
7623cf9367aa93ecfb29930afcd6aa0d7da61da7.
Review by: independent Codex /root/native_supervisor_profile_review.
Recorded by: Codex root, transcribing completed verdict. Independently reproduces
exact90,355/180/37 report,12positive/25UNKNOWN and complete sample/truncation fields.
Wrong binary/text aliases refuse before publication; shared inputs unchanged.
Actual six-command import→discovery→sorting→formatter reproduces precise nested
GDDO BSD omission, Go1/Make2. Earlier unsorted failure is explicitly not evidence.
Full overlay vet/range whitespace pass; checkout clean and no tracked reviewer edits.

RP-761 remains OPEN for correction under DP1/AC8; its actual producer failure is
independently confirmed. Next implement compiler-linked Go attribution scopes,
retaining approved root NOTICE behavior and refusing unknown nested grants. No
complete provenance map, fifth compiled MPL claim, remedy/rights/tool/adoption/
distribution/release acceptance or goal pause follows from this research verdict.

## 2026-11-04 — Correct linked Go scope attribution under existing DP1/AC8

RP-761 correction is separate from tool/MPL architecture adoption. New ordinary
real-import regression14de4b/65166→14ed88 reproduces the exact nested GDDO BSD
omission on old behavior, using actual six-command compiler graph, source file,
metadata discovery/sorting and that module's formatter output. This prevents
an unrelated dependency's copy of a license from satisfying the witness.

Replace module-root-only TSV discovery with typed compiler-package JSON. Preserve
module identity/root agreement, group actual imported directories, then collect
root+imported-ancestor license/NOTICE scopes within each module and selected
stdlib root. Symlink-resolved source directories must remain inside that root;
missing/unknown package metadata refuses. Relative file headings retain scope.
Each licensing file is classified before combining recognized grants; a valid
root/nearby grant cannot mask an unsupported nested one. Named nonregular/empty
files or missing root grant fail. Unimported subtrees cannot widen/poison the
inventory. Existing root/browser NOTICE behavior and license allowlist remain;
no gameplay, public/save schema, dependency/kernel/version/CI or MPL-policy change.
Canonical Deployment docs now describe the actual compiler-import scope.

Actual real-import regressione1158e/45471→bb7096 passes. Full affected packages
bc8236/15580→959d03 pass cold, vet205dc8 passes. Focused fixtures additionally
cover duplicate imported directories/ancestor dedup, root+nested notices and
combined MIT/BSD, an unrelated unsupported subtree, unsupported licensing file
alongside a known file, outside root, missing root, symlink grant and empty notice.
No per-file permission/unknown failure is relabelled absent. Existing model/source
research and RP-760 verdicts remain historical within their exact scope.

Next actual metadata generation from committed source, then fresh independent
corrective review of the complete code/tests/docs/records range. RP-761 remains
review-pending; full attribution/rights/tool/native/adoption/R-006 and1.0 remain.
No publication, deployment, archival or goal pause. Diff/gofmt checked.
Review by: Codex root, implementer first filter. Recorded by: Codex root.

## 2026-11-04 — Committed scoped metadata generation retains actual copied BSD

Actual clean committedef9e1ed811c18bb77816537b48ff7e2ab4e4b5f7 root generator
b40fb5/31440→9e6d1a completes53 dependencies with existing built-client inputs.
Diagnostic version0.0.0-scoped-attribution-check, explicit commit and
2026-11-04T14:15:06Z; no release bundle/exact-manifest/R-006 claim. Local owned
output path is in .cache/worker-cli-tools/scoped-attribution-metadata-output.txt.

Actual output85d2ec includes full original GDDO BSD text/2013 copyright in the
Prometheus dependency's section with relative internal/github.com/golang/gddo/
LICENSE heading. Its declared/concluded expression now correctly combines
Apache2.0 AND BSD3. The stdlib likewise records0BSD AND BSD3 from its selected
imported scopes; existing supported license classes only.53 dependencies plus
application, SPDX2.3. Output identities:
sbom.spdx.json41,725bytes SHA256e072ef559cbb5b30400fd1eae73f2dbd58af0a1a41374f4001c3f75ca37a5057;
third-party-licenses.txt172,622bytes SHA25621900c836ec78e2cb516c3e2cc8d4d93fa450073290ecd87c89793bae0823c35.
Tree stays clean, all handles terminal. Next fresh independent corrective review
of complete957dfa6b..this checkpoint, including old-code real-import failure,
full affected checks and source-scope refusal controls. RP-761 remains OPEN;
no blanket nested/tool/MPL/source-rights, adoption or release promotion.

## 2026-11-04 — Imported Go attribution correction independently approved

APPROVED, scoped corrective code under existing DP1/AC8, complete six-path
957dfa6b83f96d62b2e767bcfafdbdcd62db2204..
a8168b63d6275046f15dec87684c6dbcbd4c5563.
Review by: fresh independent Codex /root/go_scope_attribution_corrective_review.
Recorded by: Codex root, transcribing completed verdict. No material findings.

Independent cold full affected packages/vet pass, including imported-ancestor/
dedup/unrelated-subtree and all five refusal cases. Private base-main overlay
executes real imported-GDDO regression and fails exactly on missing nested BSD;
current suite passes. An initial Make-consumed anchored selector ran no tests and
is discarded, per reviewer's progress disclosure. Production tree untouched.
Fresh reviewed-HEAD generator completes53 dependencies plus application, relative
BSD/copyright heading and correct Apache/BSD Prometheus plus0BSD/BSD stdlib
expressions. Notices SHA256 exactly matches recorded
21900c836ec78e2cb516c3e2cc8d4d93fa450073290ecd87c89793bae0823c35.
Gofmt/range whitespace/final clean-tree checks pass; no tracked reviewer changes.

RP-761 closes the actual application imported-scope attribution defect. Scope
bounds/root grant/regular files/per-file known grants/separate notices and existing
browser semantics inspected; no allowlist widening. Full attribution/source-rights,
external tools/MPL adoption, native profile/host/SP8, AC8/R-006 and whole1.0 remain
OPEN. Next continue RP-759's concrete tool-source/notice delivery and remaining
native prerequisites; do not turn zero function evidence into exclusion. No
publication, deployment, archival, release approval or goal pause.

## 2026-11-04 — Recover the pinned CLI's source dependency inventory

Continue RP-759/SP1 research without selecting shipping tools or changing code,
CI, licenses, hosts or the accepted deployment floor. Actual Docker CLI28.4.0
fixture still has zero embedded dependency-module rows; that is not zero deps.
The public source checkout resolves to d8eb465f86cfceeb57f8582e373d41a558d35503,
tree c22925ab20127fcb06c19dff169e2547086d7c70, matching the tag/short revision.
Normal clone DNS failed; narrow approved public Git clone completed separately.
Go's published darwin/arm64 SDK1.24.7 checksum/size matches. The final manual
probe verifies every extracted SDK entry and every inspected CLI/vendor blob
against those pinned inputs; no global toolchain installation/settings change.

Matching linux/amd64/v1, CGO1, osusergo/pkcs11, PIE and upstream GOPATH vendoring
query succeeds:623 packages =83 CLI +338 vendor +202 stdlib. All338 map to89
module/version owners. The graph must close over the command's622 dependencies;
no unmapped/partial input is excluded. Selected-source manifest covers3,203
files. Named scoped license/NOTICE candidates total116 files/760,643bytes,
11 nested. First exploratory suffix matching missed three hyphenated Apache
filenames; corrected before this retained observation, never license clearance.

Final actual collector673650/42699→b91081 exits0. Observation92,006bytes,
SHA25682759175c5c47e26971da16f53b1f51df49c3b925f25321bd5871c919641d773;
probe15,342bytes, SHA256269905fa322c89d57ba593a28cc7c4eac8be0467e4349fdbe86889c533fd45df.
Final private controls c108a7/73941→84cc8c pass all seven refusals: unowned
vendor, incomplete graph, missing command, missing dependency, replacement,
duplicate package, existing-output/no-overwrite. Four controls actually query
the real SDK/source before altering only returned bytes in the private process.
No shared files mutated. Commands are the two retained Python .txt probes from
repository root; collector --output requires a new path. All handles terminal.

Three declared CGO populations request pthread/resolv/dl. Selected miekg/pkcs11
v1.1.1 headers pkcs11.h/f.h/t.h contain separate OASIS Open2016 IPR-policy terms;
their exact hashes and other declared native-source hashes are retained. A root
BSD filename does not resolve these references. Source selection/link requests
do not prove published-binary correspondence, exact native libraries or rights.
Collector is a manual research tool, not an eligibility/CI framework. No game
tests run for this research-only batch. Dossier/plan/RP-759 updated; historical
RP-761 diagnostic now accurately states the application omission is corrected.

Next independent bounded research/tool review of the complete seven-path
66625d7205f031625d6a5fc8834d5aedc28ef1f6..this checkpoint. RP-759 stays OPEN:
inline/header and Compose scoped terms, complete final delivery, binary/source/
native correspondence, applicable adoption, SP8/native-host and R-006 remain.
No product status/acceptance promotion, push, deployment, archival or goal pause.
Review by: Codex root, implementer first filter. Recorded by: Codex root.

## 2026-11-04 — CLI source inventory independently approved; next Compose graph

APPROVED, bounded research/tool only, exact complete seven-path
66625d7205f031625d6a5fc8834d5aedc28ef1f6..
ac45914c5eca8c8c33060ad4785bc2586c2ff20f.
Review by: independent Codex /root/native_supervisor_profile_review.
Recorded by: Codex root, transcribing completed verdict. No material findings.

Fresh independent collector output matches all92,006 retained bytes and SHA256
82759175c5c47e26971da16f53b1f51df49c3b925f25321bd5871c919641d773.
623 packages/89 vendor modules/3,203 files/116 named scopes reproduced; all
seven invalid-input/no-overwrite controls run successfully. Reviewer inspects
SDK publisher metadata/archive/extracted bytes, tracked Git blobs, ownership,
closure/rechecks and the actual CGO requests/OASIS header hashes. Exact-range
whitespace and unchanged production/tests/CI/specification paths verified.
No shared tracked reviewer changes. This approves the source inventory, not
binary correspondence, native linkage, inline terms or distribution/adoption.
RP-759 remains OPEN; RP-760/761 application corrections remain closed.

Separate next-work preparation (NOT included in that verdict): public Compose
v2.39.4 source clone c89cd6 resolves actual a32dc3da722e6b2e4fbfc72e40b1a697c4686460,
tree e702be38be3400928f7ad9e94470ad0cd6cff512. Actual matching Go1.24.7
Linux/amd64/CGO0/e2e source query e993ed/63508→09dfae exits0. Its download
messages are retained diagnostics, not errors or a silent exclusion. The normal
Go module cache receives public modules through the official proxy/sumdb;
no task-specific compiler cache or global Go configuration is introduced.
Raw graph .cache/worker-cli-tools/compose-source-graph.initial.json contains
1,290 complete package rows:24 main,221 stdlib,1,045 dependency packages,
exactly180 module/version identities matching the published binary, zero CGO.
Checkout stays clean. Temporary reproducible query source is
.cache/worker-cli-tools/compose-source-query.py; stderr is alongside raw graph.
This preliminary query has not yet bound every selected file to the prior h1
archives or resolved37 nested scopes. Next retain that actual scope correlation,
including main/stdlib terms, with meaningful incomplete/changed-source controls
and independent review. No shipping build selection, host/native/SP8/R-006,
release/archival/push approval or goal pause follows. All root handles terminal.

## 2026-11-04 — Bind Compose source scopes and repair escaped binary-name matching

Continue RP-759 with actual pinned source/SDK rather than inferring rights from
module names or empty symbol matches. Retained offline collector now freshly
queries1,290 packages:24 main,221 stdlib,1,045 dependency packages, exactly180
binary-matching module/version/h1 identities and no CGO. Main source is exact
Git a32dc3da/tree e702be38; SDK is the same publisher-checked Go1.24.7.
All6,962 inspected source/text files bind to those Git blobs, SDK archive or
complete h1-verified module archives, then their physical bytes are rechecked.
Archive hashing and subsequent reads share immutable snapshots. No replacement,
partial graph, missing owner, source mismatch or guard exhaustion becomes success.

Source observation32bc603fbff0635cc2efb76753cabcdaa8912921ca01f098468a6b9815ea018b,
204,162bytes; collector b36293daeb85b67ff900984c4af4d1bfd161c2f0e1591c6b766830ed83654fd0,
17,569bytes. Final positive c7f4ef/95211→a8b883 exits0. Full237 module candidate
texts plus main2/stdlib6 =245; all37 nested candidates remain,17 source-selected/
20 with no package selected. Buildx hclparser remains no selected package here,
not proof of absence from the published binary or authority to discard terms.
Private controls562b15/72436→41dbeb execute the actual query before four faults:
remove dependency, forge module h1, alter selected-source read bytes, alter text
in an unselected scope. All refuse exactly; existing output refuses unchanged.
No shared inputs, global Go settings, compiler cache, normal tests or CI edited.

RP-762 records a real v1 instrument miss: Go escapes periods in terminal path
components, so the raw goyaml.v2 prefix missed actual goyaml%2ev2 functions.
Old matcher plus new regression339c1a/74525→c41ba4 fails exactly on actual YAML,
percent and UTF-8 cases (Go1/Make2). Keep v1 source/result intact. V2 implements
pinned Go1.24.7 objabi path encoding separately for exact/subtree prefixes;
six regression/control cases include ordinary methods and sibling refusal.
Full cold overlaid metadata package fb0bcc/31122→e4827d passes; overlay vet
db7f0f passes. Same actual ELF still has90,355 validated functions/all37 scopes:
15positive/22UNKNOWN. Each of three YAML text scopes has184 function matches,
not three separate function populations. Two other newly source-positive K8s
spec texts still have no binary-function match. No inline/data absence inference.

Actual v2 binary report676c0dd7b01d0ec652f022c23ff78244751ba30f422099c6aff8f2aa3d5f4949,
22,234bytes; Go probe2c1d81065c1d683a61cf015c9b0049aa7a1f504afec8ef9954296fca731a54b3,
8,622bytes. A first ad-hoc summary printer assumed nil samples were lists and
raised TypeError; actual Go run/report passed, and corrected parsing preceded
retention. All handles terminal. Manual Go overlay maps virtual
server/cmd/gen-release-metadata/manual_compose_scopes_test.go to the retained
v2 .txt file; local overlay recipe/result pointer is
.cache/worker-cli-tools/compose-scopes-v2-work.txt. Root make test-go uses
GO_PACKAGES= and GO_TEST_FLAGS with explicit ./cmd/gen-release-metadata BEFORE
-args -nested-scope-root <root> -nested-scope-output <fresh-path>; vet uses the
same -overlay on its package selector. No production source mutations.

Complete nine-path a6592af3..this checkpoint awaits independent research/tool
review. RP-762 remains review-pending; RP-759 rights/inline/covered-source and
final delivery, CLI native/OASIS, binary correspondence, shipping/artifact/profile
adoption and real SP8/native-host/R-006 remain OPEN. Source selection/function
presence does not adopt terms or ship tools. No normal-CI, product status,
release/archival/push approval or goal pause. Review by: Codex root, implementer
first filter. Recorded by: Codex root.

## 2026-11-04 — Compose scope/matcher batch independently approved

APPROVED, bounded research/tool only, exact complete nine-path
a6592af336795e3e8b383e2e17c6998117b37a67..
571b0d8ccf9955139b554a438662eca1f01648d2.
Review by: independent Codex /root/native_supervisor_profile_review.
Recorded by: Codex root, transcribing completed verdict. No material findings.

Independent source report matches all204,162bytes/SHA32bc603f; binary report
matches all22,234bytes/SHA676c0dd7. All five real-boundary/private controls pass;
old matcher fails exactly actual escaped YAML/percent/UTF-8, corrected full
metadata overlay suite/vet pass. Exact ranges, no shared-input changes and no
production/normal-test/CI/RFC edits confirmed. Archive snapshot/check/use,
final-file checks, package/module closure and compiler encoding inspected.
All37 scopes remain:17 source-selected/20 no selected package;15 binary-positive/
22 UNKNOWN, never absence or exclusion. RP-762 closes the bounded instrument
miss; RP-759 rights/native/binary/delivery/adoption gates remain OPEN.

Reviewer investigated generic sumdb direct fallback, but found no reachable
counterexample under the exact non-workspace read-only query's preexisting-sum
checks. It is not retained as a finding or a general network-sandbox claim.
An initial reviewer overlay command used an unresolved placeholder and ran no
tests; excluded, then corrected executions provide the actual evidence.

Separate next-work preparation is outside that verdict. Actual hash-bound CLI
ELF inspection98a9c4 finds native producer comments (Alpine clang/LLD19.1.4 and
GCC14.2.0) and209 STT_FILE symbols, including crtstuff.c. A raw compressed-string
peek yields no matches; decompression9f70dd reveals a musl-1.2.5 build-directory
string. No raw zero is absence evidence. Private typed debug/elf+debug/dwarf
probe09b85e/52510→300a88 passes cold, using the SAME binary byte slice:807 compile
units,595 Go and212 other-language units.192 report directory
/home/buildozer/aports/main/musl/src/musl-1.2.5;12 report Go runtime/cgo;8 report
/tmp/go-build. Native units expose330 executable-section ranges,0 invalid ranges;
8 units have no ranges and are not positive code-presence witnesses. No crtstuff.c
compile unit appears, so ELF file symbols and DWARF do not yet provide a complete
native-source map. Source paths/producers are metadata, not source hash/patch or
authentication proof; no exact Alpine package revision/libgcc closure is inferred.

Ignored preparation source .cache/worker-cli-tools/cli-native-cus_test.go is
1,602bytes/SHA6230d7d084f635a2a2b5acb6ab1cd85b3294f345a2ff3ab94a30517f63f6985d;
raw807-unit report cli-native-cus.initial.json is254,668bytes/
SHA7be5ab30cc7cb79b64100d87f91cac727ebac7e2358a5cbe8776368f78e7a5eb.
The local overlay is cli-native-cus-overlay.json in the same ignored directory.
No shipping/native conformance claim. Next retain a bounded native-component
observation with actual code-range/symbol/source correlation, exact remaining
unknowns and source/notice counterparts. Then resolve tool delivery/policy and
applicable adoption before packaging/native producers. All handles terminal;
no push, deployment, archival, release approval or goal pause.

## 2026-11-04 — Inspect CLI native code ranges and preserve source counterpart gaps

Continue RP-759 research; no shipping tool/build or license policy is selected.
Actual CLI bytes stay SHA27f239f9. Retained Go probe parses build info, ELF,
symbols and DWARF from the SAME byte slice. All807 compile units (595 Go/212
other) and1,085 ranges validate against executable sections. Native330 ranges:
192 musl-labelled units/310 ranges (8 units without ranges),12 runtime/cgo units/
12 ranges,8 generated-CGO units/8 ranges. Paths/producers are metadata, not
source/patch hashes or authenticated identity. No binary code is executed.

Complete35,689 STT_FUNC census:35,125 Go CU-only,549 native CU,0 mixed,13 zero
size,0 undefined,2 without CU range match. Entire symbol-type population counted;
all unknown/zero-size function records retained. Six of209 file-symbol entries
lack any CU basename match, including crtstuff.c twice; basename equality itself
is not object identity. Known libc entry point maps to specific
src/env/__libc_start_main.c native CU. Wrong/one-bit-changed binary, six extent
cases and cross-unit/actual-unit controls execute. Final cold full metadata
package92012d/66750→771271 passes; overlay vet677a50 passes. Initial full run/
vet12363f/89172→95d01d and567315 also pass before strengthening that specific-CU
assertion; final hashes/results above are authoritative.

Native report257,712bytes/SHA c13ae17e80399773cef5c95aa14937a10e28ddbcbfe89a9cfba90f64864c8047.
Go probe12,450bytes/SHA71f1af3d6141d7e43a2d677ee145064ea6ddaad288f528dfe3422069ad8533e4.
Overlay recipe pointer .cache/worker-cli-tools/cli-native-components-work.txt;
virtual server/cmd/gen-release-metadata/manual_cli_native_components_test.go maps
to retained .txt source. Root make test-go GO_PACKAGES= uses explicit owning
package before -args -native-cli-root <root> -native-cli-output <fresh-path>;
vet uses the same overlay in its package selector. No source edits while live.

Public musl release page fetched by ordinary curl after browser fetch failed;
actual linked release/archive/signature URLs are in the dossier. Page cache
musl-releases.html SHA5153be19; archive1080786bytes/SHAa9a118bbe84d8764da0ea0d28b3ab3fae8477fc7e4085d90102b8596fc7c75e4;
detached signature490bytes/SHAd9030116fd03e4acfa0b665a13a5de46110296b4e30bb8e67be1f08af29f6306.
Both transfers60216→30cd7c and28899→5f9596 complete. Signature NOT verified;
archive hash is local identity, not an authenticated publisher checksum or proof
of the actual Alpine patch revision. No source install/extraction/execution.

Offline counterpart collector699519 exits0: all212 native units retained;
192 upstream musl SAME-PATH counterparts,12 SDK/compiler-selected path matches,
8 generated/unresolved. Full musl COPYRIGHT6,204bytes/SHAf9bc4423 retained in
ignored source cache, plus SDK LICENSE hash. All candidate-source bytes come
from immutable hashed archive snapshots. Fifteen source records carry terms
markers; not legal classification or exhaustive applicability. Report101,655bytes/
SHA795f944e336b3a95e9d2d606958d47174b4b8542bb1f96b00ebbff1c1ffb5bc8;
collector7,394bytes/SHA62a5f0874c36e129bd7764c24c4d322a411f631dbd5e7bb2b88cd984fe78eb89.
Controls6c2788 pass: changed native report/archive refuse before output, existing
output preserved, private post-parse missing counterpart stays UNKNOWN with all
212 units and false clearance/closure flags. Shared inputs never mutated.

Complete nine-path a590aa2d..this checkpoint awaits independent research/tool
review. RP-759 remains OPEN for exact Alpine/native/CRT/generated/header scope,
actual source correspondence, rights/full notice/source delivery and applicable
adoption. Source counterparts never establish exact patched code. No production,
normal test/CI, RFC, dependency or kernel byte changes; no native-host, SP8,
AC8/R-006, release/archival/push approval or goal pause. All handles terminal.
Review by: Codex root, implementer first filter. Recorded by: Codex root.

## 2026-11-04 — Native metadata/counterpart research independently approved

APPROVED, bounded research/tool only, exact complete nine-path
a590aa2d076d0b581b571406db95cd6f0342ad19..
23c646b403409dfb34e4376b1c7b59a363516d64.
Review by: independent Codex /root/native_supervisor_profile_review.
Recorded by: Codex root, transcribing completed verdict. No material findings.

Independent native report matches all257,712bytes/SHA c13ae17e; counterpart
report matches all101,655bytes/SHA795f944e. Full metadata overlay tests/vet and
extent/cross-unit/wrong-binary/specific-libc-CU controls pass. Python changed
report/archive, no-overwrite and explicit-UNKNOWN missing-counterpart controls
pass. The entire35,689-function census closes; zero-size/no-range/unknown/file
symbol gaps remain explicit. All212 units retained (192 upstream-only paths,
12 SDK-selection matches,8 unresolved);15 markers imply no rights conclusion.
Same-buffer parsing, full extents, immutable archives, source bindings, actual
release-page links/notices and all limitations inspected. Exact-range whitespace
and exclusion checks pass; no tracked reviewer edits, reviewed HEAD unchanged.

RP-759 remains OPEN. This is not source authentication, exact Alpine/native/CRT/
header closure, vulnerability clearance, distribution permission or artifact/
profile adoption. Next investigate native line-file checksums and original build
package provenance for actual source/patch correlation, then missing CRT/generated/
header counterparts and final delivery terms. Do not rerun a positive census as
closure or substitute upstream same-path text for the actual patched source.
No production, CI, RFC, dependency or kernel change; no native-host, SP8/R-006,
release/archival/push approval or goal pause. All root/reviewer handles terminal.

## 2026-11-04 — Correlate native line-file checksums; identify patched fortify header

RP-763 records a newly identified referenced header, not a shipped vulnerability.
Do not infer checksum absence from Go's LineFile API: pinned SDK line.go414
explicitly discards it. Actual Apple LLVM dwarfdump21.0.0 full --debug-line
--verbose decode9e3bb9/45501→c3686b completes0 with224,493,910 raw bytes, SHA
adc8278363292b87d16ac98b3598afb123d111550ddeb66aa58065913c4183d9. Raw stays
.cache/worker-cli-tools/cli-debug-line.raw.txt, outside ordinary Git. A first
Mach-O control is exploratory only; definitive ELF64/x86-64 control uses
clang --target=x86_64-linux-gnu -g -gdwarf-5 -c checksum-positive.c, then the
same dwarfdump flags. Known source is exactly:
int checksum_positive(int input) { return input + 1; }
with final LF; compiler and dump calls e5c051/025dab pass. Parser recovers its
actual MD5 17192cccdbb473f7998eaa5f04865623. Neither ELF is executed.

Independent typed Go CU references:807 CUs,795 distinct line tables,12 Go units
without one, all212 native units mapped and file-header counts checked. Retained
refs61,833bytes/SHA3922232a03bb7b6223a7225856716198040b7b5b52dd5a064f22eccc1a49ee91.
Full cold metadata overlay7a5ecf/20263→425688 and vetafa591 pass. Overlay recipe
pointer .cache/worker-cli-tools/cli-line-refs-work.txt maps the reviewed native
probe plus new line-refs .txt. Root Make flags use explicit package before -args,
-native-cli-root <root> -native-cli-output <fresh> -native-line-output <fresh>.
The companion consumes the same hash-bound binary/native-unit identities.

Complete795-table LLVM/Go set matches; all57 MD5 records/212 native header
censuses retained.13 native tables have checksums:11 runtime/cgo with51 records,
2 generated with6.192 musl tables have NONE; that source-body patch route is now
bounded negative evidence, not reason for more path-only matching.55 records
match candidate bytes,2 different _cgo_export.c records at the same pathname
remain unknown. MD5 is consistency evidence, not source/build authentication.
SDK13 source/header paths match;9 musl headers and the generated x86-64
bits/alltypes.h match too. Generation follows the exact source Makefile sed
rule/3 archive inputs, not make/install hooks; generated header11,446bytes/
SHAa8990c268ef62ec2e6963e757a0efad856ed4fd3611f6c28e7df17962697cc4c.

Fortify primary recipe fetched at aports commit175895ac5c1d9d62add7fe3c5a7f39548b8fd715
(the literal repository is alpinelinux/aports; no branch-time identity inferred).
Cached fortify-APKBUILD1,222bytes/SHA4aa89ccb625f28f125457ef44f761e0a4911f24138e6328131e4277ea12a8af4.
HTTPS Alpine source mirror transfer e92965 passes; all archive+3 patch SHA512s
match that pinned recipe. Collector reconstructs in private temp with patch,
never APKBUILD/make code. Recipe declares0BSD; complete root license retained.
Patched string.h5,059bytes/SHA c05d44a91b69102e7719ddf5f353117251220c7685ae6b738018fde909970d16
has actual compiler MD5c54fcf2fc92e57ea908ab6fbcdb3a56e. Unpatched upstream
4,475bytes/MD5921d971e2293868f599cdfb494e744c5 does NOT match. No unique installed
package/build/rights conclusion; patch contributors/final delivery remain.

Actual final collector89c9ab/17226→a82a45 exits0; report38,380bytes/
SHAced4bff65aa504bd86508525cc8fcbf88c4ce96605fd095df09b32c98d9f3c3b.
Probe14,520bytes/SHA c15950aa6905288458e9bc62bca230d484ba3edb748cf9fbbd88b7844da89584.
Six controls204345/20700→bc6840 pass: remove table/checksum, alter native header
census, corrupt the actual ELF positive's parsed checksum, substitute real
unpatched header, existing-output preservation. These parse complete real
inputs first, then alter private typed data/source variant; shared bytes never
change. No source/test edits while handles live. All handles terminal.

Complete nine-path5d73b715..this checkpoint awaits independent research/tool
review. RP-763 is review-pending; RP-759 overall native/CRT/generated/source-body,
header/terms/full source-notice delivery and adoption remain OPEN. Latest
summaries now link to the owning audit rather than replay every historical
research range in the ledger. No production, ordinary tests/CI, RFC, dependencies,
kernel, artifact/profile/native-host or release status changed. No push,
deployment, archival, release approval or goal pause.
Review by: Codex root, implementer first filter. Recorded by: Codex root.

## 2026-11-04 — Checksum/header research approved; publisher OCI counterpart located

APPROVED, bounded research/tool only, exact complete nine-path
5d73b715b9f71f5903c3bca11925eecef53540f9..
ef2780e4ad9bcbe8bf18cdbf8417a0c46641d49e.
Review by: independent Codex /root/native_supervisor_profile_review.
Recorded by: Codex root, transcribing completed verdict. No material findings.

Fresh refs61,833bytes/SHA3922232a and checksum38,380bytes/SHAced4bff6 match
exactly. Reviewer independently reruns full224,493,910-byte LLVM decoding with
SHAadc82783, cross-compiles actual literal-source ELF positive and recovers its
checksum. Full metadata overlay/vet and all6 controls pass. All57 records,
2 distinct generated unknowns and192 no-checksum musl tables remain. Four
recipe SHA512s/order, immutable archives and musl sed rule inspected. Exact
range/exclusion/whitespace/clean-tree checks pass; no tracked reviewer edits.
RP-763 closes only bounded header identification/instrumentation; RP-759 rights,
actual native/source/CRT/generated/patch/header/delivery and adoption stay OPEN.

Separate next-work preparation, outside that verdict: pinned CLI workflow
actually publishes dockereng/cli-bin. Public Hub28.4.0 metadata is fetched;
source tag lookup in docker-ce-packaging returns404, not proof of absent builds.
A first registry query0a80ca fails DNS before writing any targets. Narrow approved
public-read retry66ed10/91763→389426 succeeds using fresh ANONYMOUS pull token
only in memory/curl stdin, curl config disabled, no Docker config/daemon or
installed user credentials. No token is persisted or printed; all reads are
public manifests/statement/blob. Ignored source cli-oci-query.py3,115bytes/
SHAde0fcebc37b3d24f0447dfda6d423f5d0ad68d1260ae1d46847d2a9f760c382c.

All fetched objects check against their requested SHA256 digests. Public index
8c8f13345422afb0a34cd132fccc2998b9a88f83380828cdb816185f8aa3e3e2 names
Linux/amd64 manifest1aee659fbf601566d6432e9a377df50c7ed41fdd61ddf75e92b421e326eb74e1,
config f1b615441afaf9f42b8ed108c2fd888e23202f33315897baec01abb51e13eede,
single layer1c8e9c3661f1ff145d0897f08a3b2c1cc58ecfcbf8dd30cc67a89f295889a652
(20,431,116bytes), and linked attestation manifest e5d38eb37181f937871a3af3225124c260ee932c35f93a4e7775ad9d0b89f7f2.
Its21,129-byte statement is fd8346457f26d95eee6b70493d651cdb68e7eb9db11a0149f4e3cc1115fc9586,
SLSA provenancev0.2 subject is the IMAGE manifest, not a standalone executable.
Publisher metadata claims source d8eb465, Alpine/Go1.24.7, base golang material
sha256:dedeeec59c120b5b73a2db6918e90b05d59e3410580dee2e7c5cc2b48ae0c328;
reproducible:false. No signature/authentication verification performed.

The sole /docker member is43,556,144bytes/SHA507681e01b00c28dc74523c0108c16075f552eb01877d6f240d8f3536575d8b0:
DIFFERENT from actual static-tar fixture27f239f9. Do NOT assign this image's
attestation to the original binary or silently replace that artifact. Separate
preliminary section/range comparison662aee finds .text, .debug_info, .debug_line,
.comment identical, .rodata different, and all330 known native code ranges byte-
identical at the same addresses. This is an unreviewed counterpart comparison,
not proof of original build/source/patch or complete native correspondence.
Raw index/manifests/config/statement/layer live as cli-bin-28.4.0-* in the existing
ignored research directory; no filesystem extraction or container execution.

Next retain/adversarially check this exact-artifact distinction and evaluate
publisher build/base-package evidence without treating a counterpart as original
provenance. No more path/checksum census can recover the absent musl source-body
checksums. Full rights/source delivery, artifact/profile adoption and actual
native/SP8/R-006 remain. All root/review handles terminal; no push, deployment,
archival, release approval, product promotion or goal pause.

## 2026-11-04 — Verify OCI counterpart scope without laundering original provenance

Continue RP-759 with the accepted DP1/AC8 requirements, not an invented universal
signature/reproduction gate. New offline Go probe verifies exact cached index,
selected Linux/amd64 descriptor, linked attestation, all selected digest/size
bindings, runtime config, uncompressed diff-id and sole regular /docker member.
One image/attestation is selected from20 index descriptors; no other platform
is claimed covered. Immutable snapshots, no network/daemon or Linux execution.

All41 ELF sections compare (virtual NOBITS reported as metadata, compressed
sections compared after standard ELF decompression). All executable sections
and330 known native ranges match at identical addresses. Whole executable bytes
DIFFER: original27f239f9 versus counterpart507681e0. Differences are exactly
.rodata/.go.buildinfo/.note.go.buildid/.note.gnu.build-id. Image statement
subject is its1aee659f manifest, not a standalone original executable.
Publisher source/base materials are claims; reproducible:false and no verified
signature/source authentication. Base material cannot inventory later apk ops.

Full cold metadata overlay383d20/74092→ba0c4d passes; overlay vet0a6252 passes.
Foreign parsed subject rejects. Actual counterpart rejects original binding;
its exact own bytes accept identity. Real-ELF private rodata mutation preserves
code but still rejects whole identity. This prevents promoting code equality
or a matching revision string to original-artifact provenance. Shared inputs
remain unchanged. Original static fixture is neither replaced nor relabelled.

Report17,870bytes/SHA78f764f39578b88144998de2422ef46c1b3b25159381940383d185960f94fd00;
probe13,419bytes/SHA7cc9f1eeae9a39ef5b347f2d7fc62d40ea8aaee57b4d1dc09dd3aaada0ed4663.
Overlay pointer .cache/worker-cli-tools/cli-oci-counterpart-work.txt maps the
reviewed native probe plus new OCI .txt into virtual owning-package test files.
Root make test-go GO_PACKAGES= uses explicit ./cmd/gen-release-metadata before
-args -native-cli-root <root> -native-cli-output <fresh1>
-oci-counterpart-output <fresh2>; same overlay in root vet package selector.
Cached public OCI inputs and acquisition provenance are in prior checkpoint;
ordinary CI does not fetch/run these studies. All handles terminal.

Complete six-path dc0c652c..this checkpoint awaits independent research/tool
review. RP-759 remains OPEN for selected tools/terms/full source-notice delivery,
artifact-specific source/image provenance, concrete artifact/policy proposal and
applicable adoption. Exact source-body unknowns remain honest; this study does
not invent signature/reproduced-build owner requirements or solve actual native
package/header/CRT scope by naming a base. No production, normal tests/CI, RFC,
dependency or kernel change. No packaging/native/SP8/R-006, release/archival/
push approval or goal pause. Review by: Codex root, implementer first filter.
Recorded by: Codex root.

## 2026-11-04 — OCI scope comparison independently approved; terms preparation

APPROVED, bounded research/tool only, exact complete six-path
dc0c652cb3eb4f7b4d8fec5c17e890c49bc59e90..
53280165100e7305e7ceab0bb8dceb349c3eb6a9.
Review by: independent Codex /root/native_supervisor_profile_review.
Recorded by: Codex root, transcribing completed verdict. No material findings.

Full metadata overlay/vet pass; report17,870bytes/SHA78f764f3 matches exactly,
prior native snapshot unchanged. Foreign subject, actual different artifact,
exact self-binding and private data-only/unchanged-code controls all execute.
All41 sections/330 native ranges reproduce; independent ELF flag/address checks
agree. Four sections differ; equal code remains insufficient for original
provenance. Digest/size/image-subject/config/diff-id/sole-member checks inspected.
Signature/reproduced-upstream gates are not invented. Exact-range whitespace,
production/normal-test/CI/RFC exclusion and clean-tree checks pass; no tracked
reviewer changes. RP-759 inventory/terms/delivery/provenance/adoption remains OPEN.

Separate ignored-only terms preparation, not included in this verdict: primary
GNU Runtime Library Exception3.1 and OASIS IPR policy pages fetch through direct
curl after browser fetch errors. Cached GNU HTML15,177bytes/
SHAd97c26b0fc53c12eac9e58822885b3e33506c5abcb6f5e5fc1235b557c10457c;
OASIS HTML202,793bytes/SHAa7bc6bab44b783decc18786d196c4997d9c81b4ff62a131759e585e1518a5fbc.
Plain-text derivatives are convenience views, not substituted source documents.
The current OASIS policy does not by itself settle the version/deliverable grant
applicable to the2016 PKCS11 headers. No legal applicability or grant inferred.

Primary GCC mirror releases/gcc-14.2.0/libgcc/crtstuff.c candidate downloaded,
26,399bytes/SHA5859380c2242ce108b323fe216847df6f2be052fa98ca56eb6284daacda3befd.
Its own notice specifies GPLv3-or-later plus Runtime Library Exception3.1;
the exception text conditions its additional permission on eligible compilation
and applies only to files bearing the notice. This is a SOURCE CANDIDATE for an
unmapped basename, not demonstrated original CRT source or an automatic license
conclusion about our artifact. Do not label it MIT/Apache or widen the app's
permissive detector. Source URLs: https://www.gnu.org/licenses/gcc-exception-3.1.html,
https://www.oasis-open.org/policies-guidelines/ipr/,
https://github.com/gcc-mirror/gcc/blob/releases/gcc-14.2.0/libgcc/crtstuff.c.
No binary/source/policy edits or actual CRT identification result from fetching.

Next make the per-tool inventory and source/notice delivery proposal concrete,
using known module/MPL/native/header evidence and explicit unresolved terms or
artifact choices. The accepted DP1 floor requires honest provenance and delivery,
not endless path-only matching or new cryptographic gates. Resolve applicable
adoption before packaging/native producers; preserve host/SP8/R-006/full1.0
requirements. All handles terminal; no push, deployment, archival, release
approval, tool substitution, product promotion or goal pause.

## 2026-11-04 — Concrete tool inventory and delivery companion proposed

RP-759 now routes to `rfc/deployment-tool-artifact-inventory.md`, a draft contract
for the two tools under SP1/DP1/AC8. It keeps manifest v1 as byte authority and
separates tool SPDX/component inventories from the app/image scopes. Fixed paths,
raw grant/NOTICE delivery, covered source-file hashes, truthful acquisition/build
subjects and actual isolated CLI/plugin consumption are specified. The proposed
tool-only MPL policy requires owner adoption; unknown native/header/OASIS terms
remain unresolved, and no current fixture/version is selected for shipping.

Native SP1/SP8 linkage now explicitly preserves review → adoption → construction
→ real native/R-006 acceptance. The prior core/adapters verdict througha836e92b
is scoped to its reviewed bytes, not this new companion. Actual source/terms
closure still gates shipping; absent signatures/reproduced vendor builds are not
invented universal requirements. The standard SLSA envelope retains compatible
parsing, while the local build recipe/type is concretely source-commit-bound.

First filter: inspected current manifest/assembler/metadata code and parent
DP1/AC8, OP2 and SP1/SP8 against the proposal; checked primary MPL2 §§3.1–3.4
and SLSA1.2 build-provenance terms. `git diff --check` passes; the new draft's
three local links resolve. No runtime/schema/CI/product change and no
software suite just for Markdown. Complete seven-path range after070829cf needs
independent technical-spec review; no adoption or completion inferred.
Review by: Codex root, implementer first filter. Recorded by: Codex root.

Headcount ruling was also checked on resume: both accepted bodies already
record paid hiring/producing-redirecting roles with independent translation
approval. No duplicate reconciliation, price/rate/copy or balance change made.
Goal stays active; no publication, deployment, archival or release approval.

## 2026-11-04 — Tool-delivery review finds missing member and build oracles

CHANGES REQUIRED, independent technical-spec review of complete seven-path
070829cf7c5aded8996f5ff4a3db39d8cd5e48e0..
7efca4b8816ec1277c4484c1e8a7edcfa0eefba5.
Review by: independent Codex /root/native_supervisor_profile_review.
Recorded by: Codex root, transcribing completed verdict. No review handles remain.

The first draft omitted a typed identity-bound scope map and complete archive-
member/terms accounting, and lost original grant/notice locations behind hashed
payload names. It also named local-build URLs without a fixed runner identity
or executable recipe/initiation. The reviewer inspected current consumers and
primary standards; range whitespace and all local references pass. No software
suite applies to the Markdown-only change. All other checked boundaries remain
coherent; no adoption, artifact clearance or feature acceptance follows.

Correction specifies the mandatory binary/source-bound map, original source
origins and total regular-member associations; the runtime can now compare
declared coverage with actual archive contents. Independent discovery still
checks that the declaration itself is complete. Fixed assemble-release-bundle
build-tool mode, a closed vendor-ready/pure-Go recipe, exact argv/environment/
input/output and an illustrative JSON recipe replace unspecified build steps.
This v1 mode does not claim CGO/patch/codegen support or reproduce the existing
native fixture. Published acquisition remains available. Acceptance includes
unlisted real archive member, foreign origin/map and same-commit wrong-runner
refusals. The complete original-plus-correction range needs re-review.

Correction first filter: diff whitespace passes; illustrative recipe parses with
all17 declared keys, sorted source/SDK roles and digest syntax, and all3 local
links resolve. These are document checks, not a tool build or software-suite
pass. No runtime/CI/kernel/dependency/accepted-parent bytes changed.
Review by: Codex root, implementer first filter. Recorded by: Codex root.

## 2026-11-04 — Tool recipe source identity corrected after re-review

CHANGES REQUIRED, independent technical-spec review of complete seven-path
070829cf7c5aded8996f5ff4a3db39d8cd5e48e0..
b828abe57cb1800e1509e98ad03257853ef7c4d4.
Review by: independent Codex /root/native_supervisor_profile_review.
Recorded by: Codex root, transcribing completed verdict. Original member/origin
and runner findings are substantively repaired. New finding: the recipe had
no repository/revision tuple while conflating repository URI with archive URI.
Document checks also expose my inaccurate17-key count: the old example/list
had16. Both are record mistakes, not execution evidence or successful builds.

Correction adds explicit source_repository/source_commit expectations, distinct
from actual archive acquisition; the runner records those honestly, and the
post-build T3 producer verifies actual Git/import/terms correspondence before
the final binary-bound map is accepted. No cyclic pre-build final map is needed.
The actual reviewed tool-builder revision now cross-binds fixed build-type/
runner URLs and internal revision independently of later unrelated app commits;
the final bundle still hashes the exact receipt, recipe and executable.
Complete original-plus-corrections needs independent re-review. No runtime,
schema implementation, dependency or CI change; no build/feature acceptance.

Current first filter parses the example and checks its actual18 keys, exact
declared key set, source revision/digest syntax, sorted source/SDK roles and all3
local links. Log-prefix check confirms earlier committed bytes are untouched;
an unintended blank during editing was caught and removed before commit. Diff
whitespace passes. No actual build ran. Review by: Codex root, implementer first
filter. Recorded by: Codex root.

## 2026-11-04 — Tool-delivery proposal independently technically approved

APPROVED, technical specification only, complete seven-path union
070829cf7c5aded8996f5ff4a3db39d8cd5e48e0..
1648e792bce72d2bc5c7e5177c26d8ad5434e627.
Review by: independent Codex /root/native_supervisor_profile_review.
Recorded by: Codex root, transcribing completed verdict. No blocking findings.
Both prior negative verdicts above remain intact. Typed total member/origin
coverage, bounded fixed runner, separate source expectations/archive identity
and noncyclic post-build T3 binding now agree with existing consumers and parent
roles. Actual18-key example, references, whitespace and production/normal-test/
CI exclusion pass. No actual tool build/runtime run was executed or claimed.

Status/index/plan/RP-759 now reflect that bounded verdict; this does not adopt
the MPL policy, select versions/artifacts, resolve rights, implement the native
profile or prove supported-host/SP8/R-006/1.0 acceptance. Tool data and the normal
application license policy are unchanged. No push/deployment/archival or pause.

Next avoid another untested assumption before artifact adoption. The pinned
Compose source explicitly implements a standalone entry path, but source alone
does not prove required behavior. The owning plan predeclares a finite real
direct-versus-CLI/plugin configuration comparison, with complete parsed-output
comparison, failure cases and missing/swapped binary controls in the existing
isolated fixture. This may support removing an unnecessary separate wrapper
and its discovery/native-component burden; it does not authorize that change.
Preserve every actual config/compatibility outcome and native proof requirement.

## 2026-11-04 — Real standalone Compose matches configuration population

At539e1bb4, execute the predeclared study with cached Compose2.39.4 and CLI28.4.0.
New manual probe invokes the actual root release renderer for current production
base plus rotation, alongside a minimal fixture and4 invalid inputs. All3 full
parsed models match direct versus wrapped/plugin invocation; each mode rejects
all4 invalid inputs. The real base model contains7 services/4 networks/5 volume
families. Missing executable and wrong-binary-hash controls refuse; mutating the
actual parsed gameserver read_only field fails the complete-model comparator.
Direct workers receive no mounted CLI or Engine socket and have no executable
PATH fallback. Actual UID65532, CapEff/CapBnd0, NoNewPrivs1, read-only roots,
network:none and exact read-only bind sets are checked. No service model applied.

Initial sandbox run failed before container creation and its cleanup diagnostic
masked the socket denial. Instrument now exposes stderr and avoids cleanup API
calls when no create was attempted; the next ordinary run explicitly reports
Docker Unix-socket operation-not-permitted. Narrow approved probe-launcher run
cca9eb/97512→965d00 completes, all4 owned containers removed. Existing-output
controlb8af0e rejects before work;449728 confirms exact bytes unchanged.

Retained observation12,729bytes/SHA1b5ebbc46ac0ebffd0560606e409f890f88aef7bb72f3d1532fb4b0d5bb3edbe;
probe16,695bytes/SHA3952fe52dd31e058a838d1a6e5768bc2689317959ce92e78b36bed7b9f9e480c.
Source binding matches. Reproduce from root:
python3 planning/deployment-foundation/probes/compose-standalone-comparison.py.txt
--output <fresh existing-parent JSON path>. Uses existing cached public image/
tools and root Make renderer; guards unchanged, no module/cache workaround.
Raw resolved models stay in memory; retained inputs are synthetic public data.

Complete six-path range after539e1bb4 needs independent research/tool review.
Next propose a direct-Compose-only compiler/profile if that verdict holds;
current two-tool draft remains unadopted/unchanged. Embedded Docker CLI source
components in Compose still require attribution. No selected shipping version,
rights/policy closure, native Linux/amd64/SP8/R-006 or release acceptance follows
from this ARM64 Desktop fixture. No production, CI or kernel change, publication,
deployment, archival or pause. All root handles terminal. Review by: Codex root,
implementer first filter. Recorded by: Codex root.

## 2026-11-04 — Standalone study review catches a typed-comparison defect

CHANGES REQUIRED, independent research/tool review of complete six-path
539e1bb4b8500f75a2e6d93e1daa7a6beac59ff0..
6b8c892124e5472515323fcafca52ca4159b0f75.
Review by: independent Codex /root/native_supervisor_profile_review.
Recorded by: Codex root, transcribing completed verdict. Reviewer re-executes
the complete real Docker population/cleanup and no-overwrite control. All3 actual
successful stdout pairs are byte-identical, independently of the comparator;
no actual direct/wrapped divergence is observed. Random project/diagnostic hashes
vary between runs as expected, not a silently rewritten observation.

RP-764: Python equality wrongly accepts read_only:true versus read_only:1.
Reviewer executes that exact helper counterexample; root018a7e reproduces before
correction. The false-only control did not discriminate JSON type changes.
Correction uses sorted finite JSON serialization, retains the actual false-field
refusal, adds actual numeric1 replacement refusal and object-key-order acceptance.
V1 remains historical, not deleted or restamped; current probe will emit v2.

Reporting correction: my preceding 'Raw resolved models stay in memory' claim
was wrong. Public fixture models traverse worker stdout and Docker logs, which
the daemon may persist. Only the final observation omits full models. No operator
private input was used; no private-model transport guarantee is measured. The
bounded study stays public-fixtures-only; no new private transport is invented.
No normal-product/CI/kernel change. Complete original-plus-correction needs
re-review after the actual corrected population completes.

Corrected actual helper786932 rejects the earlier bool/int counterexample and
accepts reordered keys. Full corrected Docker study4c8973/59764→548d34 passes
all3 model pairs (also byte-identical),8 bad-input refusals,2 binary controls,
both false/numeric1 policy mutations and key-order acceptance. Owned cleanup
passes; no guards changed. V2 observation12,990bytes/
SHA333bda655008b8614feed44ea679dcb4ba95a107d839e338fdeb05ae5d87ccca;
probe17,733bytes/SHA748cfd8aa25f4e3c024632fd3e9528c204d064c0518c4fda43e51f6331696e18.
06124d verifies current source binding and v2 complete/cleanup/control fields.
Current dossier/plan explicitly distinguish logged public models from omitted
final models. The retained v1 is not restamped. All handles terminal; complete
seven-path original-plus-correction after539e1bb4 awaits independent re-review.
Review by: Codex root, implementer first filter. Recorded by: Codex root.

## 2026-11-04 — V2 bounded study approved; root finds numeric parsing edge

APPROVED bounded research/tool, complete seven-path union
539e1bb4b8500f75a2e6d93e1daa7a6beac59ff0..
5f6aefa65f6136b95a99c1e730891463297dac37.
Review by: independent Codex /root/native_supervisor_profile_review.
Recorded by: Codex root, transcribing completed verdict. Full actual Docker
population/cleanup, old bool/int failure/current refusal, key-order/nonfinite
controls and no-overwrite independently execute. Actual3 positive stdout pairs
are byte-identical. Source/observation bindings and history-prefix checks pass.
No tool/profile/rights/native/release adoption or private-model proof follows.

Subsequent root counterexamplee36f36 extends RP-764: default float parsing loses
the difference between JSON9007199254740992.0 and9007199254740993.0 before the
v2 comparator sees them. The actual observed compiler pairs still match bytes;
this is another oracle edge, not observed compiler divergence. V2 verdict and
data stay intact rather than being silently rewritten or overstated.

Correction preserves original model JSON text across worker stdout transport,
then parses numbers into distinct exact-token nodes. Structural comparison
ignores only object-key order; boolean/string/number kinds and numeric token
text remain distinct. Duplicate keys and nonfinite constants refuse instead of
being silently discarded/coerced. Helper43599a rejects precision, bool/number
and number/string counterexamples while accepting reordered object keys.
Full v3 population and complete original-plus-corrections review remain next.

Full actual v3 run23b44b/60658→3e6b15 completes all3 model pairs/8 bad-input
refusals/2 binary controls and owned cleanup. Both real policy-field mutations,
key-order acceptance, exact-number/number-string/spelling changes and duplicate/
nonfinite controls execute; all3 real positive stdout pairs remain byte-identical.
V3 observation13,194bytes/SHA68865f685539c10c1185acee18e1e774375f107b3b5a3a59ff64b483277df8df;
probe19,743bytes/SHAc292c825bb1cd5995f222c1e573b7aa2d89d15edf5ffac8dcb56f24c74bcd0d0.
0acf36 verifies source binding and complete/control fields. No guard or product
criterion changed. Current dossier/plan name v3; v1/v2 remain historical evidence.
All root/review handles terminal. Complete eight-path original-plus-corrections
after539e1bb4 needs independent review; RP-764 stays OPEN until that verdict.
Review by: Codex root, implementer first filter. Recorded by: Codex root.

## 2026-11-04 — Complete standalone comparison and instrument correction approved

APPROVED, bounded research/tool only, complete eight-path union
539e1bb4b8500f75a2e6d93e1daa7a6beac59ff0..
7a5368b55650692e3f7857e6fa0003c6b9b2e3f5.
Review by: independent Codex /root/native_supervisor_profile_review.
Recorded by: Codex root, transcribing completed verdict. No remaining blockers.

Reviewer executes the entire v3 Docker population/owned cleanup, reproduces v2's
precision loss, checks raw numeric text through the JSON envelope, and verifies
precision/type/spelling/array-order/missing-extra-key/duplicate/nonfinite refusal
plus nested object-key-order acceptance. Actual3 positive stdout pairs match
hashes and lengths. Fresh v3 differs only in random project identity and required-
environment diagnostic hashes. Source/report bindings, no-overwrite, prior
observations/log prefix, whitespace and product/normal-test/CI/RFC exclusions
pass. Fresh independent report13,194bytes/SHAbfcaa50ab2658f2cfeb9185e03254023c80c38b116184f7055c1379262c1d9d1
is retained at /private/tmp/compose-standalone-independent-v3-20261104.json.

RP-764 closes this instrument correction; current plan/dossier/RP-759 route to
the direct-Compose-only compiler amendment next. Its actual embedded dependency
inventory/terms, policy/artifact/profile adoption, native producers and supported
Linux/amd64/SP8/R-006 remain. No other commands, private-model handling, exhaustive
image census, rights clearance or release acceptance is claimed. No shipping pin,
product/kernel/CI change, publication, deployment, archival or goal pause. All
root/review handles terminal; the full nine-tier objective remains active.

## 2026-11-04 — Check the direct compiler's input construction before specifying it

The next normative amendment needs an executable input contract, not merely
renaming docker compose. Existing W00 compose_input/broker wording and Engine
E12/E13 output paths must reconcile together; the broker role is also H06 HTTP
and cannot be removed globally. No production/profile change is made yet.

Owning plan predeclares two input arms on the existing real population: explicit
process environment plus empty dotenv must preserve v3 outputs; referenced-secret
paths absent must still compile equal models with6/8 confirmed file references.
The new selector is closed and echoed by actual workers. Direct Docker config
contains only empty auths; no discovery or Engine access. Input handling uses
each case's actual env fixture for every outcome, not the expected success flag.
The selected input arm and exact producer source enter the observation. Prior
typed/token controls, guards, cleanup and no-private-data claims stay intact.
This can justify a proposed grant reduction, not adoption/native acceptance.

Both actual arms complete:29b5a5/33190→e2cbc0 process_environment and
2ec074/93818→97a190 process_environment_absent_secrets. Each runs all3 complete
direct/wrapped pairs,8 bad-input refusals,2 binary and all typed/token controls,
with owned cleanup. 047f2b compares all6 positive process-env hashes/lengths to
v3: identical. Absent-secret arm checks0/6/8 actual references in minimal/base/
rotation for BOTH modes; no referenced file exists. Unknown selector5b36c0 exits2
before work. Current embedded/parent syntax checks pass; guards unchanged.

Probe21,760bytes/SHA319f70dd6f6e0c9cc88a8994059047144c3e24d54fe43e6390a1ba1ab3969a35.
Process observation13,583bytes/SHAc57c42c3079e46c749d95bda0225e7dfb05ca56e6ef43b975cbb22c90ce736ad;
absent-secret observation13,940bytes/SHA865c4393288b0f0c8e3292e2b030614d1df2e301ec5df8cfeb642be2af0d1ee1.
Both are retained with this plan. Public synthetic values do not establish
production operator validation, private output handling or native-host support.

Normative translation revises the unadopted tool/native/adapter drafts together:
one upstream Compose file, exact release/API asset binding, preserved scoped
source/notices/MPL obligations, no discovery or worker Engine/secret FD. H06's
HTTP broker remains. C01 fixes original source files, effective interpolation,
clean environment and LogConfig none plus original E12 attached BEFORE command
release; no E13/private-log shortcut or receipt cycle. Ordinary nonsecret daemon
Config.Env persistence is disclosed, not called memory-only. Full actual private/
native/SP8/R-006 acceptance remains required. The unused unadopted self-build/
archive/OCI acquisition alternatives are withdrawn from this compiler profile;
image provenance and every accepted release criterion remain in force.

No runtime/API/schema implementation, dependency, CI, kernel or authored content
change. Prior reviewed ranges remain scoped to their historical bytes. Current
combined research/tool and technical-spec range after765aae44 needs independent
review before any adoption request. No shipping version, rights clearance,
publication, deployment, archival or goal pause. Review by: Codex root,
implementer first filter. Recorded by: Codex root.

First-filter construction checks for this revision:73fafb resolves12 local RFC
links including C01's anchor, matches all21 declared interpolation names against
the actual base/rotation templates, verifies both source/worker-selector bindings
and absent-reference counts, and confirms invalid selector created no output.
Log prefix remains byte-identical to HEAD; whitespace passes. Pinned Moby28.4.0
daemon/attach.go and container/config.go were read from the official repository:
stdio readers precede stream handoff, logs are requested only for logs=1, and
AttachStdout/AttachStderr are explicit fields. C01 names their non-TTY framing
and original stream readiness/terminal distinction; no private/native run was
performed. No unrelated software suite is claimed for Markdown changes.

## 2026-11-04 — Direct compiler review catches image-default inheritance

CHANGES REQUIRED, complete11-path765aae446dc78d52f72dbb22a2ba137ef87565c9..
5950a28420043ce842af48a86015541a59d40f28. Review by: independent Codex
/root/native_supervisor_profile_review. Recorded by: Codex root. Both new
actual Docker arms independently pass their complete valid/invalid/binary/token
populations and owned cleanup; all6 process-env positive hashes/lengths match
v3, and absent references are0/6/8 in both modes. Fresh reports at
/private/tmp/compose-process-env-independent-20261104.json SHA
 aa2245ff2b004083bd18fcce1d9cbeb1c7d75366b0517c550ffb43f5ce9aa2b1 and
/private/tmp/compose-absent-secret-independent-20261104.json SHA
 1e66d5540255ab0adb5125ab2c774d8729efc71af5e08df43b595e47ab3c1403.
Selector/no-overwrite, retained bindings, paths/names, whitespace and historical
log/observation preservation also pass. All review handles terminal.

Two normative findings prevent approval. Moby's actual create/merge code appends
image ENV defaults; a clean Config.Env request is not a clean compiler environ.
C01/W00 now require an explicit exact environment on exec through the existing
authenticated release, plus a reviewed exact image-default population and full
effective-container comparison before shim start. No other role gets the new
field; no input FD, environment file or generic channel is introduced. The child
map comes solely from admitted input/control values, never inherited defaults.
Daemon nonsecret storage stays disclosed. Actual native acceptance must check
image/default/container/control tampering and observed child inheritance; no
public research is relabelled private/native proof. The second finding removes
T2a's stale invocation of T5's withdrawn runner, preserving full source-member
coverage and requiring separate future build adoption. The earlier plan's then-
next draft/two-tool statements are explicitly historical.

Root read the pinned Moby28.4.0 create.go/commit.go through the official repository
and confirmed the merge. This revision changes only proposal/plan/log text;
research script and both actual reports stay unchanged. No additional Docker
rerun or unrelated software suite is needed for these normative corrections.
The original negative verdict is retained; the complete amended range requires
independent re-review. No adoption, implementation, artifact pin/rights clearance,
private/native/SP8/R-006 acceptance, CI promotion, publication, deployment,
archival or goal pause follows. Review by: Codex root, implementer first filter.
Recorded by: Codex root.

## 2026-11-04 — Corrected direct compiler proposal independently approved

APPROVED, technical-spec and bounded research/tool only, complete11-path union
765aae446dc78d52f72dbb22a2ba137ef87565c9..
70acee7fdfb8865361c74f4c785fc7cb9db85f82. Review by: independent Codex
/root/native_supervisor_profile_review. Recorded by: Codex root, transcribing
completed verdict. Both findings resolved; no remaining technical blockers.

Reviewer rechecked exact C01-only authenticated exec environment, pinned image/
effective container validation before start, defaults/order/source-target binding,
original attach/native completion, retained H06 and private-output limits. Probe
and both reports are byte-identical to its previously executed complete successful
Docker arms, so those actual results are reused rather than needlessly rerun for
text-only corrections. Source/report hashes,21 names,12 local RFC paths, negative
verdict/historical log prefix and whitespace pass. Normal product/tests/CI bytes
remain excluded; tree clean and all review/root handles terminal. The root casing
assertion was corrected against unchanged text, not called a software pass.

Current statuses now cite this complete range. Next: ask Marco to adopt the
reviewed tool-inventory/upstream-acquisition and tool-only MPL/source-delivery
contract so static packaging implementation can proceed. AGENTS and RFC-0000
reserve specification/policy adoption to Marco; independent review supplies no
such decision. Native-profile/parent adoption, actual selected version/rights,
private/native/SP8/R-006 construction and full1.0 remain separate. No runtime
implementation, feature acceptance, publication, deployment, archival, release
approval or goal pause follows from this checkpoint.

The concrete async tool-packaging/MPL-policy adoption question is now delivered,
following0b0c726e, and unanswered. It authorizes static packaging only if adopted;
native-profile/parent adoption remains separate. Do not repeat the question or
begin dependent implementation without the answer. Goal remains active.

## 2026-11-04 — Rebuild a current candidate containing the reviewed runtime fixes

DP1/AC1/AC8 local construction, not adoption of the pending native/tool proposals.
Prior goal turn made progress (reviewed specification/input study). Source is clean
893bb1e7f9e7dcd92abbf1b21926c5486fec21f7, epoch1793820849/2026-11-04T19:34:09Z;
existing LOCAL0.1.0-preview.2 label is retained, not a release/version decision or
a narrower1.0 objective. df84ce1a artifacts predate reviewed Account/Guild, backup
and application root/nested attribution corrections, so they cannot prove those
current bytes. Historical records/trees stay untouched.

Root Make targets rebuild client, all6 static Linux/amd64 executables, epoch8
content and54-package application metadata twice into separate fresh roots.
Build1 handles24495/39116 complete1c9c21/a78e18; metadata6974ed; image94227→6fd288;
assembly91679→e909c5. Build2 handles9085/58987 complete d048d1/c5e578; metadata/
native validator26981→62f03f; image79922→a34b11; assembly91170→dd5945. Both image
builds are no-cache under unchanged pinned BuildKit/frontend and same local tag.
All14 actual pinned Linux/amd64 Syft invocations finish, each pair with distinct
raw creation timestamps and byte-identical normalized inventory. Counts per
image are220/179/41/50/51/433 plus663 Playwright packages. Actual registry index
(where present)→uniqueamd64 child→runtime-config and actual config os/architecture
are freshly verified for all6 upstream/rehearsal references through actual
docker buildx imagetools inspect raw-manifest and image-config reads. First
inspection attempt assumed every pinned digest was an index;
Caddy/Playwright are direct manifests. That attempt is not verification; all live
read handles were polled terminal before corrected manifest-or-index reads pass.
No version/reference changes or registry publication.

Both roots compare all75 files/74 actual bound artifact hashes. Fresh application
notices SHA21900c836ec78e2cb516c3e2cc8d4d93fa450073290ecd87c89793bae0823c35
matches the independently approved correction. Prometheus declares/concludes
Apache-2.0 AND BSD-3-Clause; stdlib0BSD AND BSD-3-Clause. No external-tool/MPL or
image-rights clearance follows. Real normalized inventories are scanned, not
copied from old trees; compiler caches remain the Makefile's normal shared cache.

New manifest SHAba620fb466f79f58b950be095e42d6dceff9acb81da095e0e1ee71e4e070cfae;
archive SHAda47007d6dced9d4a222827d9afb0134edd8dc64cbb900498e9e7ce17497e705;
gameserver config SHA8189a568da56051f1a542c4a7fac5613a7935c52039547e69d9fcac12ebd2e53.
Intact actual package population46029→749623 passes. Current native validator's
all9 semantic-rebound mutations return exact3/fixture_rejected: missing catalog,
client, license, config, helper; image/config/SBOM substitution; wrong epoch.
Each validates its healthy hardlinked fixture first. Post-probe full equality
holds, private work directory empty, exact scanner-ancestor inventory empty.
Secret scan96449→7de5d6 passes2552 tracked files plus actual image, findings0,
objective_completed=true/guard_exhausted=false, bound to exact source/manifest.

Derived new candidate-893bb1e7.json from actual outputs; native validate-build
695a5d passes. Artifacts remain ignored local trees, not a published release.
Engine28.4.0/ARM64 Desktop and Compose2.39.4-desktop.1 are actually inspected;
this is local cross-building/scanning, not native Linux/amd64/clean-host proof.
Both builds are one candidate, never an immediately previous release pair.
Current hosted CI,27 accounting failures/history guard, genuine predecessor,
native/private/SP8/R-006 and all nine-tier gameplay/rights/accessibility/preservation
work remain. No running game service, provisioned secrets, host changes, product
code, CI, content mint, publication, deployment, archival or goal pause.

Review by: Codex root, implementer/output first filter. Recorded by: Codex root.
Complete new four-path evidence range after893bb1e7 through this containing commit
needs independent corrective evidence review: retained record, canonical docs,
owning plan/log. All root handles terminal, tracked source stayed fixed until
these evidence records. Pending adoption questions remain unanswered and must
not be repeated or inferred accepted.

## 2026-11-04 — Current candidate evidence independently approved

APPROVED, scoped corrective build evidence, complete four-path
893bb1e7f9e7dcd92abbf1b21926c5486fec21f7..
40d33bb7332c89247250f966e8ad5d9403e2ac7f. Review by: independent Codex
/root/current_candidate_evidence_review. Recorded by: Codex root, transcribing
completed verdict. No material findings.

Independent actual75-file/74-artifact/mode/source-epoch/archive/record checks pass.
Fresh root-Make six-executable rebuild is byte-identical, with actual static
Linux/amd64 ELF/build metadata. All6 upstream/rehearsal runtime configs/platforms
corroborated. Intact package/build-record consumers and all9 actual semantic
mutations execute: exact3/fixture_rejected; originals still compare afterward.
Corrected attribution,54 application packages and precise secret-scan bindings
pass. Reviewer inspects all7 pairs of distinct raw scan timestamps/counts and
identical normalized outputs, not a fresh rerun of14scans. Append-only log and
whitespace pass; HEAD/tree unchanged, private review artifacts removed and all
handles terminal. Full verdict supplies no predecessor pair, rights clearance,
current CI, native/SP8/R-006, feature/archival or release acceptance.

Root updates only actual artifact/review status and the material whole-product
artifact checkpoint. Pending native/tool/copy/metadata/accounting/Headcount
adoptions remain questions, not blockers to all accepted work. Full nine-tier1.0
stays active. Next reconcile newly observed hosted maintenance failure at exact
published source, preserving incomplete measurements and original budgets;
no source change, publication, deployment or goal pause.

## 2026-11-04 — Make the proposed projector diagnostic contract testable

Previous goal turn made concrete progress: independently approved current exact
artifact and digest-checked failed hosted evidence. Clean HEAD407b641d, all prior
handles terminal. Tool-packaging/native-host and other owner questions remain
unanswered; no repeated question or inferred adoption. Continue the existing
RP-757/SP6/A6 proposal, not production code under an unadopted contract.

Source check confirms native draft already proposes15s evaluation/300s staleness,
three release projection gauges and one writer/cache policy, but does not define
projection_success versus release business outcome or an exact fail-loud rule.
Actual current host collector freshness cannot establish original ledger reads.
Owning plan now predeclares the actual pinned Prometheus rule population and
age-arm deletion control before execution. It tests the future alert predicate,
not the unbuilt service/cache/native producer. No new policy is adopted or accepted
release criterion waived. Parent DP7 reconciliation/adoption and full actual
cache/reboot/producer/receiver/native acceptance remain explicit.

Before the final rule run, first-filter type review adds three unsupported native-
histogram cases, one per required float gauge. Ordinary float comparisons can
silently discard histogram samples; expected behavior remains an alert. Predeclare
actual unchanged-rule baseline followed by a concrete float-presence check if the
baseline misses them; retain all32 previous cases/34 evaluations and budgets.
These exercise type admission in the rule, not a claimed production exposure.

Actual pinned evaluator results: first17148→98a464 exits1 because nine nonfinite
series used unsupported NaNx40/Infx40 shorthand. It runs no valid whole population;
original fixture is preserved ignored with SHA3bda18504af68c148ec8b720111e852bffdf7f86c673e032f23c5b86d8e16e31.
Only repeat syntax changes to41 explicit literals;32cases/34evaluations350b2d
pass. Private age-arm deletion controle8850d fails exactly the stopped-observation/
fresh-host case at5m30s (expected alert, actual empty). No expectations relaxed.

First-filter type follow-up actually confirms the gap: unchanged-rule baseline
fe2430 fails all3 native-histogram cases with expected alert/actual empty. Final
predicate adds absent(metric+0) for each required float gauge; actual35bf0a
passes all35cases/37evaluations. Final-source private age-arm controlf38f3c still
fails only the stale case, with all35cases retained. These are genuine pinned
PromQL executions, not tests of a copied reference classifier. Existing ordinary
operators/unit-rule documentation is cited; no histogram/parser behavior is merely
assumed from those pages. Rule1914bytes/SHA2356ba57b271b233a4af92d65ef197d2bf218504ea9e1dd077ed70414df2df1e;
fixture37613bytes/SHA881a9b675493e5836b72fd52f6e237286515780534f022a5773fda709051a050.

Reproduce from root with the existing pinned Prometheus image and Docker run:
--rm --platform linux/amd64 --network none --read-only --user65532:65532,
--cap-drop ALL --security-opt no-new-privileges, private16m tmpfs/tmp, entrypoint
/bin/promtool, workdir/work and read-only mount of the owning probes directory;
arguments test rules /work/release-projection-rules.test.yml.txt. Use exact image
prom/prometheus:v3.12.0@sha256:69f5241418838263316593f7274a304b095c40bcf22e57272865da91bd60a8ac.
Actual cached configd2f7aaa3 is Linux/amd64 (0a3bf4); no Engine/socket/network/
secret grant, normal Go/cache/dependency/workflow changes or host service. All
commands complete below the predeclared60s containment guard, and exact owned
label inventory2be703 is empty. Desktop/emulation supplies no native-host timing.

P01 is a concrete new draft amendment, not covered by old through70acee7f review.
It separates observation_success from operation outcome, defines exact gauge/
float/age/cadence semantics and metric ownership. Diagnostic and business files
are separate atomic units: F02 supervisor jobs remain status-only; the retained
service alone writes the event-backed business snapshot, then healthy diagnostics.
No cross-file transaction is claimed. The restart matrix explicitly keeps lost/
torn/frozen/conditional/active-round states paused until their real authority.
All32 initial metric cases plus3 type cases are series-level predicate tests;
cache names do not pretend an actual cache/ledger/restart/receiver was executed.

The draft also proposes explicit parent DP7 projection-alert and C01 private-
compiler logging reconciliations, preserving the seven original alert families,
service logs and14-day floor. Accepted parent/owner copy is untouched. SP6's earlier
unbuilt-native-cache-proof-before-adoption demand contradicted SP8's construction-
after-adoption rule; this proposed correction now requires isolated predicate
proof first and ALL actual native cache/reboot/receiver proof after adoption,
before acceptance. This is a visible draft sequencing correction, not a waived
accepted criterion or host proof. One initial patch fails its context check and
changes no file; corrected edit applies. No unsupported source claim is inferred.

Review by: Codex root, implementer first filter. Recorded by: Codex root. Complete
eight-path research/tool/technical-spec range after407b641d through this containing
commit needs independent review: two proposal fixtures, native core/adapters/index,
owning plan/log and existing RP-757 row. No production rules/metric/cache/runtime,
CI, schema, kernel/content, parent adoption, native/SP8/R-006, feature/release,
publication/archival or goal pause. All root handles terminal; pending owner
questions remain pending, not grounds to ask them again.

## 2026-11-04 — P01 review requires two factual/sequencing repairs

CHANGES REQUIRED, complete eight-path
407b641dffa169f07c931164c6211bb0db8de631..
b5a7fb36f09f4fccb455df2d0bd3d241eb78b019. Review by: independent Codex
/root/native_supervisor_profile_review. Recorded by: Codex root. Actual pinned
35cases/37evaluations pass; independent age-arm deletion fails only the stale
fresh-host case at5m30s, and type-check deletion fails exactly3 histogram cases.
Evaluator config/platform, fixture hashes, unchanged controls, local links,
append-only log, whitespace, scope and owned-container absence pass. No actual
cache/ledger/restart/publisher/receiver/native/adoption proof is inferred.

P01 wrongly described rollback's release-result mapping as existing CLI behavior.
Root reads actual runRollback103–104: it returns directly without the install/
release writer. Body now calls rollback's shared-family mapping newly proposed.
Second, RP-757 kept its old pre-adoption projection/reboot-cache gate despite
SP6's corrected sequencing. Existing ledger sentence now agrees: reviewed tool
contract/isolated predicate proof before adoption; actual producer/cache/reboot/
receiver proof after construction and before acceptance. Actual selected payload/
rights closure stays the SP1/T6 shipping gate, not an invented pre-construction
clearance requirement. No accepted parent/owner text is altered.

No rule/fixture/threshold/expectation changed; their real executed evidence stays
applicable. Only two body/ledger corrections plus this append-only record. Full
original-plus-correction union needs independent re-review. All root/review handles
terminal. No software suite/rerun merely for these text repairs, inferred adoption,
publication, release or goal pause.

## 2026-11-04 — P01 approved; reconcile the full proposed contract graph

APPROVED, technical-specification and bounded rule-fixture only, complete8-path
407b641dffa169f07c931164c6211bb0db8de631..
dcc151f277c862c2b78a1d5d16befc97de6fbaac. Review by: independent Codex
/root/native_supervisor_profile_review. Recorded by: Codex root. Both text findings
resolved, no remaining blockers. Rule/fixture bytes unchanged; independent35/37
PASS, one exact age-arm failure and three exact histogram failures remain valid.
Original negative verdict/log prefix, links, whitespace, exclusions and clean
tree pass; all handles terminal. No duplicate Docker run for text-only corrections.
This is neither actual producer/restart/receiver/native proof nor owner adoption.

Before asking for architecture adoption, root inspects the linked CR/OP bodies.
CR6/Open questions still say the native record/profile contract needs specification;
OP9 both links that proposed contract and still demands completed native producers
before adoption. Those are stale own-draft gates, not an excuse to implement or
weaken acceptance. Bodies now distinguish concrete linked schema/profile contracts,
applicable review/adoption/parent reconciliation BEFORE construction, and complete
actual producer/native/R-006 proof BEFORE acceptance. Historical counterexamples,
quiescence/ownership, useful successful paths and lost-authority refusals remain.
No inventory or completed mechanism experiment is repeated as a substitute.

SP8's adoption package now names CR1–CR6/OP1–OP9/SP1–SP8/adapters and enumerates
exact proposed DP1/DP5/DP6/DP7/DP8 reconciliation. It explicitly discloses the
substantial native supervisor, rootful/direct Linux/amd64 prerequisites, reused
backup service/image, no additional resident service, paused scheduler/guarded
restart, status-only versus business-metric writers, P01 alert and C01 narrowly
unlogged compiler. Existing seven alerts,14-day evidence, RPO/RTO/retention,
manual rollback/seven-day window, no Down migrations, provider-off and full1.0
remain. It does not adopt tool-only MPL, backup signatures/keys, any shipping
artifact or another pending proposal, and does not silently edit the accepted
parent/owner-authored body. All construction depends on adopted contracts.

Verification for this new normative reconciliation is direct full-diff/source/
contract trace and local reference/whitespace checks, not a software suite for
Markdown or a table validator pretending native effects. P01 rule/fixture bytes
stay fixed; no public/runtime/schema/kernel/CI/content or accepted-parent edit.
One initial patch context miss applies nothing; corrected exact-context edit
succeeds. New complete eight-path range afterdcc151f2 through this containing
commit requires independent technical review, including the whole linked adoption
scope and absence of weakened accepted gates. Pending owner questions remain
unanswered and are not repeated. No adoption request until that concrete package
passes review. Full nine-tier goal active; no publication/deployment/archival/pause.
Review by: Codex root, author first filter. Recorded by: Codex root.

## 2026-11-04 — Linked review catches the old outer-writer alternative

CHANGES REQUIRED, complete8-path
 dcc151f277c862c2b78a1d5d16befc97de6fbaac..
7b32e5738822981ee1c07277a6b62fb08ddf0a2c. Review by: independent Codex
/root/native_supervisor_profile_review. Recorded by: Codex root. Full linked
CR/OP/SP/adapters/tool/accepted-parent inspection finds one remaining conflict:
CR6 still permits an outer CLI outcome writer before completion, while SP6/P01
retire it and use retained-service business projection after the governed event.
Other host/service/parent/dependency/sequencing clauses have no additional blocker.
Links, append-only log, whitespace, exclusions/clean tree and unchanged rule/fixture
hashes pass; actual35/37 plus both controls reused, no unrelated rerun. All handles
terminal, no tracked reviewer edits or inferred acceptance.

CR6 now selects exactly SP6/P01's ownership in its BODY: no outer outcome writer;
only admitted F02 diagnostics, fully joined before seal; retained-service business
projection after the governed final event/required graph, with OP7 install handoff.
Other real backup/restore/helper publications remain covered by whole-operation
closure and actual receipts. The seal restriction explicitly scopes to operation-
owned effects, preserving ONLY verified transferred retained-service activity.
Acceptance/Open-question text now treats reintroduced outer outcome writing as
refusal, not a legitimate pre-Finished alternative. Initial-install wording names
the now-proposed producer without pretending it is built/proved. Current-runtime
source counterexamples remain factual/historical, not erased or claimed fixed.

Two-path CR/log correction only; no rule/fixture, accepted parent, runtime or other
contract changes. Original negative verdict preserved; complete linked union
needs independent re-review. No native/software rerun for this normative fix,
owner adoption, new permission, publication, archival, release or goal pause.

## 2026-11-04 — Linked package approved; concrete architecture adoption requested

APPROVED, linked-contract/adoption-package technical scope ONLY, complete8-path
 dcc151f277c862c2b78a1d5d16befc97de6fbaac..
ed9a8f2ad7b57f8ed9be420c2be8259e30272f05. Review by: independent Codex
/root/native_supervisor_profile_review. Recorded by: Codex root, transcribing
completed verdict. CR6 conflict resolved, no remaining blockers. The complete
proposed adoption scope/parent changes/dependencies are concrete for owner choice.
CR6 retires the outer outcome writer, closes diagnostic jobs before seal and
preserves transferred retained-service business projection/other helper receipts.
The final I04 clarification keeps later admitted imports while forbidding old
operation writers to masquerade as retained work. Links, original negative/
verdict log prefix, whitespace, clean tree and exclusions pass. P01 source hashes
unchanged; independent35/37 and both actual discriminating controls reused.
No native/Docker/software rerun for text fixes; all handles terminal, no tracked
reviewer edits. No implemented/native/rights/adoption/release proof is implied.

After that completed review, ONE concrete async question is delivered: adopt the
linked deployment-control design and delegate matching parent-body reconciliation.
It names construction of release identity/operation evidence/native supervision,
Linux/amd64+local rootful Docker, existing backup service/restart-safe ownership/
verified results, extra availability alert and private-compiler log exception,
while retaining14-day service logs. The substantial subsystem is disclosed, not
called a routine fix. SP8 is the reviewable scoped proposal. AGENTS/RFC-0000
reserve this architecture/specification adoption to Marco; technical review does
not supply it. The question remains unanswered; do not repeat it or implement
unadopted dependencies. Existing tool/MPL, signing-metadata, native-host access
and release decisions remain separate unanswered questions, not absorbed approvals.

Root reconciles current status only. Accepted parent/owner-authored body remains
untouched pending actual adoption/delegation. The full1.0 goal stays active;
proper construction and all actual private/cache/restart/receiver/native/SP8/R-006,
CI/gameplay/rights/accessibility/ending gates still remain. No new tier, publication,
deployment, archival, feature/release acceptance or goal pause.

## 2026-11-04 — Refresh the exact artifact after socket-authentication corrections

DP1/AC1/AC8 local construction at clean source
9768d681ae0227d59126a060b22f3b481f740236, epoch1793829295 /
2026-11-04T21:54:55Z. The older893bb1e7 artifact predates approved RP-765/RP-766;
this rebuild includes them. RP-767 remains OPEN/unapplied. Kernel0.3.225/epoch8
and the LOCAL0.1.0-preview.2 label are unchanged; this is no release call.

Fresh roots `.cache/release-current-9768d681/bundle-a` and
`.cache/release-rebuild-9768d681/bundle-b` match all75 files, modes and74 bound
artifact hashes. Root Make targets rebuild client, six static Linux/amd64 ELF
executables, content and54-package application metadata twice. Compilation uses
the normal shared Make cache. Both pinned no-cache image builds yield config
SHA2e8f584a59c7a0f4ba59b89faa32998ac591c069f0ee6b8c4d0fe587f84b3003,
archive SHA847070a3fc7909aed5672bab299036e58fb2bb1ceff947fdb8baecc7808d8ece,
manifest SHAaf8408c7737c7dd34816301899f818adc52d189c754a9018972a3c141ab174b2.
Only the gameserver executable differs from the previous candidate's six binaries.

All14 actual pinned Syft scans complete:220/179/41/50/51/433 packages plus663
rehearsal-only Playwright packages. Each pair has distinct raw creation times
and identical normalized inventories. All6 upstream/rehearsal registry/config
Linux/amd64 identities are freshly inspected. The fresh native helper accepts
the build record and intact package; all9 rebound package mutations return exact
exit3/fixture_rejected. These two identical candidate trees are not a legitimate
predecessor pair. Source/image secret scan passes2560 tracked files plus actual
archive, zero findings, objective_completed=true/guard_exhausted=false, bound
to the exact source/manifest. Local comparison.json, package-probes.json,
registry-proof.json, scan/build logs and secret-scan.json retain the details.

Invalid setup is disclosed: the first frozen pnpm install was sandbox-denied,
and compilation started before that result was inspected. That output is retained
separately as release-provisional-9768d681 and receives no verification credit.
The identical install then succeeds; both counted roots are freshly rebuilt
afterward. First image attempt is denied writing Docker builder activity before
building; the same target succeeds under its narrow authorization. All counted
handles terminate, probe work is empty, scanner-container inventory is empty;
tracked source stays fixed through verification. No production byte changes here.

The new candidate-9768d681.json is derived from actual outputs and validates.
Engine28.4.0/ARM64 Desktop and Compose2.39.4-desktop.1 are observed: these are local
cross-build/package checks, not native clean-host/runtime acceptance. Genuine
predecessor, current hosted CI, rights, proposed native architecture, SP8/R-006,
all nine tiers and other gameplay/accessibility/preservation requirements remain.
No push, deployment, archival, policy adoption or goal pause. Complete four-path
9768d681..this commit needs independent corrective artifact-evidence review.
Review by: Codex root (implementer first filter). Recorded by: Codex root.

## 2026-11-04 — Refreshed authentication-fix artifact independently approved

APPROVED, complete four-path
9768d681ae0227d59126a060b22f3b481f740236..
40969759d59aa969219cced212e2a6ff9c636869.
Review by: independent Codex /root/auth_refresh_artifact_evidence_review.
Recorded by: Codex root, transcribing the completed verdict. No material findings.

Reviewer freshly rebuilds all6 static Linux/amd64 executables byte-identically;
executes75-file/hash/mode and74-artifact/source/epoch/archive/config checks,
verifies packaged binary/all32 content files, and runs build-record/intact/nine
semantic-tampering consumers (each negative exact3/fixture_rejected). Fresh
registry selection/config/platform rereads match all6 upstream/rehearsal images.
Post-probe equality and whitespace pass. Reviewer independently inspects root's
14 actual scans, not a rerun: distinct raw pair timestamps, identical package/
relationship graphs and normalized outputs;54 application packages, notices hash
and exact source/manifest secret-scan bindings agree. A reviewer inspection first
assumes the wrong image content prefix; corrected Dockerfile-owned
/opt/cloud-clicker/content/ verifies every actual file. This is an inspection
assumption, not an artifact defect. All processes finish, owned review artifacts
removed; tracked tree stays clean at40969759.

Root now reconciles only review status and material whole-product artifact pointers.
RP-767 stays OPEN/unapplied; approval supplies no new adopted contract, native
clean-host/runtime proof, legitimate predecessor, current hosted green, rights
clearance, complete feature/archival or release acceptance. Full nine-tier1.0
remains active; pending questions are not repeated or inferred accepted.

## 2026-11-05 — Rebuild the current version226 artifact; reuse immutable upstream proof

Previous goal turn makes progress: RP-767's complete correction independently
approved71d8e15c and recorded4c5f662c. The accounting/search authority check is
substantive: accepted R10 mandates binary search; both replacements still need
pending adoption, unlike RP-767's routine correction. Editorial28aefd29 corrects
only the already-completed wording-review status; no algorithm changes/adoption.

Accepted DP1/AC1/AC8 local construction at clean source
28aefd29b19221e2d909f44add2fa899713a33c5; epoch1793833991 /
2026-11-04T23:13:11Z. Kernel0.3.226/epoch8; existing LOCAL0.1.0-preview.2 label,
not a release call or narrowed1.0 floor. Prior9768d681 predates RP-768/RP-767.
Frozen pnpm install terminates0 before builds (11.15.1, already up to date).
Full root Make client/six static Linux-amd64 executable/content/metadata builds
44430/82520 terminate0 into separate fresh roots; compiler cache stays the normal
shared Make cache. Assemblies28810/23665 terminate0; all75 files/modes and74
manifest-bound artifacts match. Only gameserver differs among six old/new
executables; application notices retain SHA21900c836ec78e2cb516c3e2cc8d4d93fa450073290ecd87c89793bae0823c35.

Both no-cache image builds24964/88646 use unchanged pinned BuildKit/frontend.
Config SHA5acc501a4fdd0c7412849042004ff0f3da22cd6dad43488cc9fd981107ceca84;
archive SHA3da8398dc9295a4ad51fd7c3000cf13aed801c737d30406414f86fb7459648d4;
manifest SHA00398bbcf2d773ccb7ccf878e04dd3cada9a1cf3c489c5890d4354a1bba248c5.
Fresh actual gameserver Syft scans6708/15966 both pass,41 packages with distinct
raw timestamps23:17:44/23:21:18Z. Application metadata has54 packages/53 dependencies.

Upstream evidence reuse is explicit, not14 fresh scans: all6 pinned registry/
selected Linux-amd64 config identities are freshly reread and unchanged. Each
root reuses its own six prior9768d681 raw scan inputs (220/179/50/51/433 plus663
rehearsal-only Playwright packages). Prior normalized bytes match the retained
record's digests; raw/canonical copies and original creation timestamps remain
in retained-scan-inputs and retained-upstream-inventories.json. Existing CLI
normalization73756/21034 binds the new release timestamp. Independently replaying
the producer on the old checksum-bound canonical inventories47190/52311 yields
byte-identical outputs (strict cmp), proving no package-graph change was hidden
by raw reuse. All7 normalized inventories match between current/rebuild trees.
No scanner/version/image reference, license policy or software gate changes.

Actual DELIVERED Linux helper (not only host-native source consumer) executes
inside a disposable Linux-amd64 test image config593dc3550e0514c6ec0f7861c9bf7bf415aeabc9f4852f83f7469bfe06db1335:
network none, readonly root, no capabilities/new privileges, UID501:20, one owned
fixture mount/TMPDIR, existing emulation AVX2 fallback and60s containment. A common
private parent holds byte/mode-identical hardlinked candidate/rebuild copies,
required by the existing hardlink probes. Intact package/build-record pass0;
all9 semantic mutations return exact3/fixture_rejected, never generic failure.
All4 original/fixture trees retain their75 hashes/modes afterward; work is empty.
These are identical current candidates, NOT a legitimate previous release pair,
native profile, clean host, running stack or R-006 acceptance. No daemon/secret/
source-workspace mount or game service. All disposable probe containers exit.

Manifest/source-bound secret scan1644 passes2564 tracked files plus actual image,
zero findings/objective_completed=true/guard_exhausted=false. New record derives
from actual outputs and validates with both native and delivered Linux consumers.
Engine28.4.0 ARM64 Desktop/Compose2.39.4-desktop.1 observed. Local build/scan/probe/
reuse/comparison logs and JSON remain under the two28aefd29 roots and dedicated
release-linux-probes-28aefd29 fixture. Source stays fixed until these records;
all root handles terminal. Four-path28aefd29..this commit needs fresh independent
corrective artifact-evidence review. No fresh full server/browser/hosted run,
rights, predecessor/native architecture/clean-host/feature/release acceptance,
publication, deployment, archival, owner-answer inference or goal pause.
Review by: Codex root (implementer/output first filter). Recorded by: Codex root.

## 2026-11-05 — Current kernel226 artifact evidence independently approved

APPROVED, no findings, complete four-path base-exclusive/tip-inclusive range
28aefd29b19221e2d909f44add2fa899713a33c5..
d3e525719d7aa1b1b31ef2fd87f15aa2ce5aa1cb.
Review by: fresh independent Codex /root/kernel226_artifact_evidence_review.
Recorded by: Codex root, transcribing the reviewer's executed verdict.

Independent root-Make rebuild matches all six delivered Linux-amd64 executables.
All75 file hashes/modes,74 bound artifacts, source epoch1793833991, epoch8,
kernel0.3.226, static ELF, image config/archive/manifest agree. Native AND actual
shipped Linux helpers pass build-record/intact consumers0 and reject all9 semantic
mutations exact3/fixture_rejected. Fresh six registry/selected-config/platform
reads pass. Actual normalizer replay from both old raw AND checksum-bound canonical
inputs in both trees passes24 strict comparisons. Application54packages/53 actual
DEPENDS_ON relationships and notices match. Original/private bundles retain all75
hashes/modes after probes; work empty. Whitespace/source freeze checks pass.

Reviewer inspects, does not rerun, root's two fresh gameserver scans, reused
upstream scan chains, builds and manifest/source-bound zero-finding secret scan.
Scans are not silently relabelled independent or fresh upstream measurements.
Sandbox-only daemon/network attempts execute no successful probes; scoped
escalation resolves them. One reviewer diagnostic first counts all SPDX relations
as dependencies; actual relationship-type inspection resolves that assertion error.
All handles terminal;11 owned containers absent/private reviewer directory removed;
shared caches/pre-existing BuildKit containers preserved.

This closes corrective artifact-evidence review only. Editorial current pointers
now refer to source28aefd29; old records/logs remain historical. No native profile,
clean host/R-006, current hosted green, rights, legitimate predecessor, complete
feature/archival or release approval follows. Full nine-tier1.0 remains active;
pending owner questions are neither repeated nor inferred accepted.

## 2026-11-05 — Current reviewed runtime rebuilt into matching local candidates

Accepted DP1/AC1/AC8 construction at clean source
b80424502d9c92c2e5d5f5598a0e33be9bec2775 / epoch1793843642 /
2026-11-05T01:54:02Z. Prior28aefd29 predates approved RP-769 relay and7c850359
reader changes; this candidate includes them, kernel0.3.226/epoch8 unchanged.
The local0.1.0-preview.2 label is not a release call or reduced nine-tier floor.

Root Make builds74998/81459 independently reconstruct client/six static Linux-
amd64 executables/content/metadata in current/rebuild-b8042450 roots. Host Go1.27.3,
Node26.13.0; declared pnpm11.15.1/frozen dependencies confirmed after permitted
install. Initial sandbox cannot write its ordinary manager cache: install99610
fails and version14769 falls back to global12.16.1; no build uses that fallback.
The shared client/dist overwrites the unretained first output: do NOT claim that
as a retained comparison. Capture actual B output, separately rebuild A client67002,
retain/compare all7 file hashes/modes, and restore B snapshot before B assembly.
Six executable pairs, both staged trees/metadata and both no-cache images match.
Actual builder image pin matches unchanged Make definition/frontend.

Image25632/47203 config61fc24a5ae78af67a821f9c6d47afaec6872fc20b3db3423e664e5c35bf279a0;
archive4455c6ca6404ee678bbcded8c229aa43cf99d620510ee222acd7f5fb313c0e61.
Two FRESH gameserver Syft scans6795/73024:41 packages, actual raw times02:02:56/
02:03:59Z. App metadata54 packages/53 DEPENDS_ON; notices SHA21900c836ec78e2cb516c3e2cc8d4d93fa450073290ecd87c89793bae0823c35
unchanged. All6 immutable upstream/rehearsal manifest/config/platform identities
freshly reread/hash-checked;2 roots explicitly reuse their OWN prior28aefd29 raw
and checksum-bound canonical inventories. Normalizer89872 completes12 outputs
and12 strict canonical replays;7 normalized inventories match. These are not14
fresh scans. Initial index-only parser refuses direct Caddy/Playwright manifests;
correct parsing proves both direct and indexed identities, no omitted population.

Assemblies54209/29223 match all75 files/modes and74 manifest-bound artifacts.
Manifest9cb75f9f6cccd8fc44d55661f91b414581e1c65403a5a24a9073b1059c6dd59e.
Native freshly built helper60029 validates build record/intact package. Actual
DELIVERED Linux helper executes in cached config593dc355... with network none,
readonly root, no caps/new privileges, UID501:20, only one owned common fixture
mount/TMPDIR and existing emulation AVX2 fallback. Build-record and intact results0;
all9 semantic tampering populations return exact3/fixture_rejected, never guard/
setup refusal. All4 original/private75-file trees remain unchanged; work empty,
all11 containers absent. They are two identical candidates, NOT predecessor proof.

Manifest/source-bound secret scan71863 passes2,568 tracked files plus actual image,
findings integer0/objective_completed=true/guard_exhausted=false. An inspection
initially expects an array; correct scalar inspection confirms the producer's
actual zero result. All root handles terminal; source fixed until these records.
Raw logs/source/client snapshots/reuse/probe/scan summaries remain in the two
b8042450 roots and release-linux-probes-b8042450. New retained candidate JSON,
canonical docs/plan/log form the complete four-path review batch afterb8042450.
Designated corrective artifact-evidence review PENDING; no self-approval, native/
clean-host/R-006, full current CI, rights, legitimate predecessor, full-feature or
release acceptance. RP-770 remains open. No pending-question repeat/adoption,
push, deployment, archival or goal pause. Shared builders/caches preserved.
Review by: Codex root (implementer/output first filter). Recorded by: Codex root.

## 2026-11-05 — Current runtime artifact independently approved

APPROVED, no findings, complete four-path base-exclusive/tip-inclusive range
b80424502d9c92c2e5d5f5598a0e33be9bec2775..
2527574db410293f9d2892aad4d4ea244ee0b330.
Review by: fresh independent Codex `/root/current_runtime_artifact_evidence_review`.
Recorded by: Codex root, transcribing the reviewer's executed verdict.

Fresh independent root-Make rebuild matches all6 static executables AND7 client
files against delivered package and BOTH retained snapshots. All75 hashes/modes,
74 manifest bindings and actual image-layer binary/content payload agree.
Native record/intact pass; shipped-Linux build/intact0 and all9 semantic cases
exact3. Both raw/canonical chains complete24 independent strict normalizer
comparisons; fresh6 root/selected registry/config/platform identities agree.
Reviewer inspects/hash-verifies root's two fresh Syft and source/image scans,
without claiming independent scanner execution. No fixture, guard or source
change supplies the approval; exact source remainsb8042450.

Proof .cache/current-runtime-artifact-review-b8042450/review-results.json. All
reviewer handles terminal,11 owned containers absent, private fixture removed;
shared caches/builders and unrelated work preserved. No tracked reviewer edits;
clean2527574d before this editorial closeout. Current artifact pointers and the
live Deployment coverage row are reconciled; original build/review logs remain
append-only. No new software test for these status edits. Do not repeat this
reconstruction while runtime bytes remain unchanged.

This closes local corrective artifact-evidence review only. RP-770, accounting/
search adoption, historical guard/current-hosted CI, rights, legitimate predecessor,
native supervisor/clean-host/R-006, full features and release approval remain open.
Full nine-tier1.0 goal ACTIVE; no question repeated, adoption inferred, push,
deployment, archival or release call.

## 2026-11-05 — HTTP credential restore instrument independently approved

APPROVED, corrective test/research instrument only, full seven-path range
42ce00f0d2d1d081fcd7acf6c3837111f84f7ee5..
604f2a826b652a39f2fa6d96d7a3fae19f4af40a.
Review by: fresh independent Codex `/root/backup_credential_observation_review`.
Recorded by: Codex root, transcribing its independently executed verdict.

At current source0bb8f381, unchanged original instrument passes the declared
`make test-deployment-backup` cold (8526): ten integration functions, zero skips;
HTTP observation2.55s. Deleted source has0 account rows and401 for original
access/refresh/recovery; B_pre restore1/200/200/200; B_post0/401/401/401.
Healthy bystander200 throughout; wrong credentials rejected; every successful
credential reaches the original Founder. All five credential SQL populations
match their selected backup before HTTP can mutate them. Native package34187
and vet pass, with native DB/child-process skips explicitly excluded from DB
credit. These are current-source corroboration, not execution of historical604f2a82;
that original eight-function run remains separately attributed above.

All seven original paths inspected; no earlier designated verdict covered them.
No additional mutation needed beyond the meaningful retained negative controls.
Reviewer leaves no tracked edits; all handles terminal, owned Compose containers
and network absent, serial SQL/shared binary ownership released. Existing native
retention-metadata research still reports2/2 protection failures despite package
green. RP-125 remains open: constructor-only catalog/no-gameplay seam and frozen
unchanged synthetic keys do not cover composed bootstrap/Guild, expiry/rotation,
off-host operations or clean-host release. No D-009/D-015, feature acceptance,
archival or release approval; next useful boundary is actual composed bootstrap.

## 2026-11-05 — Composed bootstrap/delete/restore rights observation

Account D5/AC6 + Deployment DP6/AC4; baseline9a4bf86f. Predeclare the bounded
composed supplement in the existing inventory plan before measuring. New manual
`make test-account-restore` and a small companion Compose file reuse declared
Postgres16 dump/restore tools with a separate project/binary and only read-only
balance/changelog/moderation mounts. No production/balance/schema/migration/CI
bytes change. Actual Compose/Server.Start/readiness, Foundation initializer and
Guild deletion wiring replace the earlier constructor/no-gameplay seam.

Subject/bystander originate through bootstrap, Founder/initial Company reads
work, encrypted B_pre/B_post bind all six credential-table populations before
restored startup. Final cold42376 passes3.64s with no skips: deleted source and
B_post0 account/401 access/401 refresh/401 recovery/409 bootstrap_expired;
B_pre1/200/200/200/201 exact original response. All successful credentials and
the retry reach the original Founder; bystander200 and wrong-code/token/key
controls discriminate, no extra account created. Epoch8/migration85/hashbaa89050...
is actual content; fixture manifest with those coordinates is not a release.

First SQL22496 also passes3.86s. Final adds explicit manual-lane configuration
failure: missing source URL8873 exits1 immediately, not skipped; owned project
then removed. Native package42103 and selected vet pass, DB/manual populations
skip natively and are not credited as SQL. Initial native36202 accidentally
loses `$` through Make's flag expansion and runs no tests; native6980 cannot
open an existing httptest listener in sandbox. SQL69043/27725 cannot access
Docker in sandbox. These are invalid/no-evidence attempts, not hidden passes;
permitted runs above finish successfully. All handles terminal; disposable
containers/network removed, shared builders/caches preserved. Current artifact
b8042450 remains valid for unchanged runtime; no rebuild performed.

RP-125 remains open. This closes only composed initialization/bootstrap-retry
evidence, not expiry/rotation, populated Guild membership, gameplay actions,
Caddy/off-host/RPO/RTO, exact-release/clean-host R-006, rights decisions or full
feature acceptance. Complete eight-path batch after9a4bf86f needs designated
corrective test/research review. Review by: Codex root (implementer first filter).
Recorded by: Codex root. No owner adoption, push, deployment or goal pause.

## 2026-11-05 — Composed bootstrap restore instrument independently approved

APPROVED, no blocking findings, corrective test/research instrument only;
complete eight-path range
9a4bf86f1ef1215ff61296e8c6f3c0efa228bf9c..
8c7ccf175c8a89861ae7a6a9df18c8f72743ae49.
Review by: fresh independent Codex `/root/composed_account_restore_review`.
Recorded by: Codex root, transcribing its executed verdict.

Fresh declared real Linux-amd64/Postgres16 target75053 passes4.61s, one
integration function/three observed states, zero skips. Deleted source/B_post
remain0/401/401/401/409; B_pre1/200/200/200/201 with exact original bootstrap
bytes; bystander200 throughout. Six SQL populations bind before restored
composition; every successful/replayed credential reaches the original Founder;
invalid controls refuse and no extra account appears. Independent missing-URL
control immediately exits1, not skip. Full native87025 and vet pass;15 top-level
manual/DB skips receive no integration credit. Formatting/range whitespace clean.
Private assertions do not print credential/response contents.

Reviewer makes no tracked edits; all handles terminal, owned project containers/
network absent, SQL/binary ownership explicitly released. Its editorial suggestion
is applied here: restore the pre-existing host/journal/mixed-architecture sentence
beside its original backup-profile paragraph. Dossier/backlog reconcile scoped
approval; no software rerun for these factual/editorial records. RP-125 and
D-009/D-015 remain open, as do expiry/rotation, populated Guild behavior, gameplay,
off-host/clean-host and full feature/release gates. Full nine-tier1.0 goal stays
active; no artifact rebuild, inferred adoption, push, deployment or archival.

## 2026-11-05 — Restored credential lifetimes do not remove recovery-code access

Account D2/D5 + Deployment DP6; baseline0f3a5318. Predeclare two application-clock
offsets16min/31days over BOTH selected B_pre/B_post dumps in the existing rights
plan. Extend the approved composed observer, preserving original three states,
six-family SQL binding before each startup, actual composition/cleanup, exact
receipt bytes, original-Founder binding, invalid controls and no-new-account
check. Aged bystanders obtain fresh sessions through their recovery code, so
their own expired access cannot make an unhealthy service look protective.

Real declared Linux-amd64/Postgres16 target20771 passes6.68s, seven states/zero
skips. B_pre16min: access401/refresh200/recovery200/bootstrap201; its exact
replayed response contains expired access, not fresh usable access. B_pre31days:
access401/refresh401/bootstrap409, but original recovery200 issues a fresh pair
reaching the ORIGINAL Founder. All B_post offsets deny access/refresh/recovery
401 and bootstrap409; bystanders200 throughout and no extra account appears.
Epoch8/migration85/contenthashbaa89050... unchanged. These measured populations
refute token expiry alone as a remedy for observed recovery after an older
full-dump restore. They do not adopt account expiry or any deletion policy.

Native full gameserver93433 and selected vet pass; manual/DB skips excluded from
integration credit. Only a clarifying test comment changes after SQL execution;
runtime assertion/fixture bytes remain those executed. All handles terminal;
owned Compose containers/network removed and independently inventoried absent.
No production/balance/schema/migration/CI/deadline changes or artifact rebuild.
Keys stay unchanged; application time offsets are not elapsed wall time, key
rotation, arbitrary lifetime or native clean-host evidence. Populated Guild,
gameplay/browser renewal, off-host/RPO/RTO/R-006 and D-009/D-015 remain open.

Complete six-path batch after0f3a5318 needs designated corrective test/research
review; this is no new feature/policy acceptance. Review by: Codex root
(implementer first filter). Recorded by: Codex root. Goal remains active; no
repeated owner question, inferred adoption, push, deployment or archival.

## 2026-11-05 — Restored credential lifetime instrument independently approved

APPROVED, no findings, corrective test/research instrument only; complete
six-path range0f3a53186466502dd2ba454e53e55dfc0578f580..
6df67ab206abaf69f74c5a15e2ec12c082b7a3d3.
Review by: fresh independent Codex `/root/restored_credential_lifetime_review`.
Recorded by: Codex root, transcribing its independently executed verdict.

Fresh real target85435 passes7.31s/all seven states/no skips. B_pre16min is
401/200/200/201; B_pre31days401/401/200/409; B_post all three ages denies
401/401/401/409. Successful pairs reach original Founder; freshly authenticated
bystanders stay200, six SQL families bind before startup, invalid inputs refuse,
no extra account appears and aged receipt replay is not called usable access.
Native6871/vet/format/whitespace pass;15 native manual/DB skips get no SQL credit.
Reviewer leaves no tracked edits; all handles terminal, owned containers/network
absent, project/binary ownership released. No repeat software run for editorial
closeout. RP-125 stays open: token expiry alone does not remove this measured
recovery-code access, but no permanent-validity or retention/deletion policy is
adopted. Other lifetimes/rotation/Guild/browser/off-host/clean-host gates remain.
The live execution-queue header is reconciled to current routing; its October
snapshot stays historical. No source/artifact rebuild, full-feature acceptance,
push, deployment, archival, repeated question, inferred adoption or goal pause.

## 2026-11-05 — Reviewed Exit-credit runtime reconstructed into local candidates

Accepted DP1/AC1/AC8 packaging at clean committed source
c5130acfcd52d07c9bb7b913bd9be5aff4b3e447, SOURCE_DATE_EPOCH1793853149,
created2026-11-05T04:32:29Z. It includes independently approved RP-772 and
migration00086; kernel0.3.226/epoch8 unchanged. Prior b8042450 packaging omits
that runtime change. Keep its record/proof historical; no release/push/deployment.
The existing local0.1.0-preview.2 label does not narrow the full nine-tier1.0 goal.

Declared pnpm11.15.1 frozen install97441 passes; Go1.27.3 darwin/arm64 and
Node26.13.0 observed. Root Make builds4223/5996 separately reconstruct six static
Linux-amd64 executables, client/content/metadata in fresh current/rebuild-c5130acf
roots. Capture A's seven client files BEFORE B build; both retained snapshots,
all six binaries, content and metadata match. Restore/verify each actual client
snapshot before its own assembly. No first-build output is silently overwritten.
Pinned existing BuildKit0.24.0 builder retained, never pruned. Both no-cache image
builds14785/54062 pass; image configbf04f1e89d33bcef75091b9b1ac3f5835d93d3fea91cf5b9f04e39cd31aa3a9f,
archive6d9e74b36a0d133a7abbe86a8ee4eb7b709d936b91be3301107ec092e2284588.
Actual33-file image payloads equal each fresh binary/content build including
ownership/modes; compressed blob AND uncompressed diff-ID hashes are verified.

Root's first payload precheck incorrectly compares compressed bytes with a
raw-layer diff-ID and fails. Root then incorrectly starts scan6945 despite that
failure, passing an unverified hard-coded source IDbd8052b2.... Syft/normalizer
exits0 but that is INVALID provenance, not an image-identity proof. Retain raw,
normalized output/log and invalid-scan.json separately; exclude them from both
bundles and valid scan counts. Correct compressed/raw checks, read actual config
bf04f1e8... from each archive, then fresh scans33834/3712 pass with41 packages each.
As an actual consumer control, feed the quarantined inventory into otherwise
unchanged assembly: Make2 at gameserver SPDX binding, partial output retained
without a release manifest. No assertion or source change makes that refusal.

All6 unchanged upstream/rehearsal root/selected manifest, config and Linux-amd64
identities freshly verified61145. Each new root explicitly reuses its own prior
b8042450 raw/canonical inventories with checked SHA chains and original scan times;
these are NOT twelve new upstream scans. Make normalization44118 performs twelve
raw normalizations plus twelve prior-canonical replays; all12 equality checks pass,
and all7 normalized inventories match across roots. Two new gameserver scans are
fresh. Assemblies12224/49315 match all75 files/modes and74 manifest bindings;
manifestce953309d3c60d995a3f70268fbf619c2c538259a270dddac1287f0dc9112c63
explicitly declares migration86 and exactc5130acf source.

Fresh native helper9714 validates build record/intact package (87848). Actual
DELIVERED Linux helper76829 runs offline/read-only/no caps/new privileges in the
existing593dc355... amd64 runner with UID501:20 and its existing AVX2-emulation
accommodation. Build record/intact return0; all9 semantic tampering populations
return exact3/fixture_rejected, never guard/setup refusal. Four75-file original/
private fixture trees stay hash/mode-identical; work empty,11 owned containers
absent. Shared builders/caches/unrelated resources preserved. These two identical
candidates are not a legitimate predecessor. Secret scan44236 binds exact source/
manifest/image, reports2,576 tracked files/findings0/objective true/guard false.

All root handles terminal; source frozen through execution. Raw inputs/results
are ignored in current/rebuild-c5130acf and release-linux-probes-c5130acf roots;
new small retained candidate JSON and canonical docs/plan/queue reconcile. Complete
five-path batch afterc5130acf awaits independent corrective artifact-evidence
review. Review by: Codex root (implementer/output first filter, not designated
approval). Recorded by: Codex root. No native-profile adoption/construction,
rights/predecessor/hosted/full-feature/clean-host/R-006 or release acceptance.
RP-770 and other held owner decisions/CI failures remain; no goal pause.

## 2026-11-05 — Current Exit-credit runtime artifact evidence independently approved

APPROVED complete five-path range
c5130acfcd52d07c9bb7b913bd9be5aff4b3e447..21a889b726b5a0e46b001b8dcaf9c0714c6bd3d9.
Review by: fresh independent Codex `/root/exit_credit_artifact_review`.
Recorded by: Codex root, transcribing the independently executed verdict.
Own six-executable/seven-client rebuild matches both snapshots/bundles; all75
hashes/modes,74 bindings, migration86/source/epoch/config/archive and both actual
33-file image payloads/owners/modes agree. Own28 strict normalizer comparisons,
6 fresh immutable registry/config/platform reads, native record/intact and actual
shipped-Linux record/intact pass. All9 semantic cases return exact3, never guard/
setup refusal. Own valid assembly matches; quarantined invalid inventory fails
at gameserver SPDX binding before a manifest. Root's two valid41-package scans
and exact-source/image zero-findings scan are hash-inspected, NOT independently
rerun; the invalid unverified-ID scan stays excluded. Unsupported builder --format
returns125/setup only; actual builder-container Config.Image read verifies the
pin. Do not turn that setup error into an artifact/test finding or passing proof.

Evidence `.cache/exit-credit-artifact-review/review-results.json`
SHA64c7ed27ce87c915f0d9022f5e05389f7bd5b9e67435acf95f5e462fac9ba4c0
is digest-checked by root. All handles terminal,11 owned containers absent,
private reviewer fixtures removed; original evidence/shared resources preserved,
no SQL used, clean21a889b7 before this editorial record. Source remainsc5130acf;
no rebuild for documentation. This is local corrective artifact-evidence approval,
not full feature/rights/hosted CI/predecessor/native/clean-host/R-006/release
acceptance. No publication/deployment/archival, pending-question adoption or goal pause.

## 2026-11-05 — Reviewed frozen-age runtime rebuilt into local candidates

Accepted DP1/AC1/AC8 packaging at clean committed source
`aa5e16e8e3e7ee5b0d31278378d19b3008ed4815`, SOURCE_DATE_EPOCH1793861549,
created2026-11-05T06:52:29Z. Includes independently approved RP-773
`c195c0ae..a12218ab`, kernel0.3.227 and appended migration00087, epoch8 unchanged.
Prior c5130acf kernel226/migration86 packaging remains historical. The local
0.1.0-preview.2 label is unchanged; no release, push or deployment is authorized.

Frozen pnpm11.15.1 install passes; actual Go1.27.3 darwin/arm64, Node26.13.0,
Engine28.4.0 and Compose2.39.4-desktop.1 match record. Existing builder inspect
confirms pinned BuildKit0.24.0, preserved rather than recreated/pruned. Root Make
build4913 and60006 separately reconstruct six static Linux-amd64 binaries,
client/content/metadata. Preserve A's actual seven-client-file snapshot before
B runs; both snapshots, six binaries,32 content files and two metadata files
compare equal. Restore/verify each snapshot before its own bundle assembly.
No source/test/metadata edits occur while these checks read their inputs.

No-cache image builds37795/11094 both pass. Before scanning, hash every archive
blob, verify the uncompressed layer diff-IDs separately from compressed blob
hashes, inspect actual runtime user/entrypoint/source/version labels, and compare
all33 image payload files to fresh binary/content bytes/modes/owners. Both match:
config97a529277755b23bd320c808d76a305ad147fea3f4491f0d04e1ae6398102239,
archivec34d4c591b810bd9ce16a1ee330241e1fd0db63951f7b9a0bd6dc5172b920436.
Only after that passing identity check, fresh actual Syft4962/22468 scans each
archive:41 packages, same normalized inventory772a6f14.... No invalid scan is
included or relabelled in this reconstruction.

Fresh registry31348 re-reads all6 unchanged upstream/rehearsal root/selected
manifest/config/platform identities, including fresh Caddy config output, and
matches every pinned Linux-amd64 identity. Each root explicitly reuses its own
prior c513 raw/canonical inventory chain with checked SHA and original scan time;
these are twelve reused inventories, NOT twelve new upstream scans. Normalizer
24551 executes24 commands/12 equality checks; all7 normalized inventories match
across roots. Two new gameserver scans above are fresh. Assemblies94772/25180
match all75 files/modes and74 manifest bindings. Manifest
f5ce06393fb27c41b01dbfbce30002df9e2ed0d37e230dd6c51f25d6eb271010
names exactaa5e16e8 source/migration87/epoch8. All artifact paths/digests are
independently recomputed in the local comparison, not merely counted.

Fresh native helper19439 validates build record/intact package (exit0 each).
Actual DELIVERED Linux helper62974 runs offline/read-only/no caps/new privileges,
UID501:20 in existing593dc355... amd64 runner with its established AVX2-emulation
accommodation. Record/intact return0; all9 semantic tampering populations return
exact3/error_class=fixture_rejected. No timeout/guard/setup refusal is counted.
All4 original/private75-file trees retain bytes/modes; native and Linux work empty;
all11 named owned containers absent, shared builders/networks/caches preserved.
These identical candidates are one artifact, not a legitimate predecessor pair.
Manifest-bound source/image secret scan exits0,2592 tracked files/findings0,
objective true/guard false. Emulated Linux helpers on this ARM64 Desktop host do
not constitute supported native-host or R-006 proof.

All root handles terminal and source remained clean/frozen until recording.
Ignored evidence is in `.cache/release-current-aa5e16e8`, matching rebuild root
and `release-linux-probes-aa5e16e8`; exact small candidate JSON is retained.
Canonical deployment docs/plan and current execution routing are reconciled.
Complete five-path record/docs batch afteraa5e16e8 needs fresh independent
corrective artifact-evidence review. Review by: Codex root (implementer first
filter). Recorded by: Codex root. No independent approval is self-supplied.
RP-770, Clout/history guard/current hosted CI, privacy/rights/accessibility,
native-profile adoption/construction, predecessor/clean-host/full-feature and
release gates remain. No repeated owner question or goal pause.

## 2026-11-05 — Frozen-age runtime artifact evidence independently approved

APPROVED complete five-path `aa5e16e8..e404cce4`.
Review by: fresh independent Codex `/root/frozen_age_artifact_review` under
Marco's corrective-review delegation. Recorded by: Codex root. Exact source is
stillaa5e16e8/kernel0.3.227/epoch8/migration87; this record adds no runtime.
Reviewer independently rebuilds6 Linux binaries,7 client files,32 content and2
metadata files through root Make/shared compiler caches; all match delivered
artifacts. Actual two75-file/mode bundles/74 bindings,33 image payload files each,
compressed blobs/uncompressed diff-IDs/owners/source/epoch/kernel/migration all
match.28 raw/canonical inventory normalizations compare equal;6 fresh registry
manifest/config/Linux-amd64 reads match the pinned identities. No finding.

Fresh native helper record/intact return0; actual delivered Linux helper record/
intact0 and all9 semantic fixtures exact3/fixture_rejected, guards never fire.
Root's2 fresh Syft41-package scans and source/image2592-file/findings0 scan are
independently inspected and source/binary/hash-bound, NOT independently rerun.
No third image build is claimed: actual payloads match the independent rebuild.
Upstream inventory scans remain explicitly reused with original provenance.

Reviewer setup disclosure: first fixture copy collides with its own normalizer
output directory named rebuild; exits1 before any helper run. Separate fixture
subtree corrects setup, no product/assertion/timeout changed, no credit for that
attempt. Root validates review manifest `.cache/frozen-age-artifact-review/review-results.json`
SHA256 `15b15bd4f4cefb5aa83dc23c8526be427fa6b4bc61a8571bbfb3ea36a47d33be`.
All handles terminal, original4 trees/client dist preserved byte/mode-exact,
reviewer's own fixtures preserved/work empty/11 owned containers absent; no SQL,
tracked reviewer edits or shared-builder/cache cleanup. Inputs released.
Canonical docs/board/current pointers now name this exact approved artifact.
Previous c513 record stays historical. No further rebuild for editorial records.
This remains one local candidate, not a predecessor pair, supported native host,
R-006, hostedCI/privacy/rights/accessibility/full-feature or release acceptance.
All current holds and the full nine-tier1.0 goal remain; no push/deploy/archival.


## 2026-11-05 — Current reviewed runtime reconstructed into local candidates

Accepted DP1/AC1/AC8 packaging, frozen clean source
`a7d5a4a4972cc7cfbdd00f1c7770f3b958a4d670`, SOURCE_DATE_EPOCH1793868133,
created2026-11-05T08:42:13Z. Includes reviewed RP-77404642f4c..a12346b0
and RP-776184e1dfc..512024c7, kernel228/migration87/epoch8. Source delta
sinceaa5e16e8 is those production/UI guards, version mirrors and tests/research;
no balance, deployment runtime, schema or new World mechanic is introduced.
Local0.1.0-preview.2 label stays unchanged; no publication/deployment.

Toolchain preflight finds system pnpm12.16.1. Initial sandbox install58786
fails pinned-manager cache writing, no build credit. Permitted frozen install
60573 explicitly executes pnpm11.15.1 successfully; Node26.13.0/Go1.27.3
and actual Engine28.4.0/Compose2.39.4-desktop.1 are recorded. Existing builder
remains pinned BuildKit0.24.0; preserved, not recreated/pruned. Root Make61297/
99287 then94132 separately build six staticLinux-amd64 binaries, seven client
files,32content and2metadata files. A's real client output is snapshotted before
B; both snapshots are restored/verified before their own assembly. Every byte
and mode matches. Shared normal compilation caches are used; no clean-host claim.

No-cache image builds44464/77502 pass. Actual archive blobs AND uncompressed
layer diff-IDs,33 payload files/modes/owners, runtime UID/entrypoint/source/version
labels match fresh builds. Both configfa3de094e2edb2cf994eea3502acf7fdb89e5a1e943e9896ccd11bc1e1b9c9e0,
archive0d8bffb74d600bdca6868fd3f343e5aee2b461d6fbaf92073f4fcf51dec1b984.
Only after identity passes,2 fresh Syft scans49388/3458 each find41packages,
normalizedc1cc7825.... Six fresh registry25558 reads match immutable upstream/
rehearsal root/selected/config/platform identities. Each root reuses its own prior
aa5 raw/canonical inventory chain, checksum/original-scan-time bound. These12
upstream inventories are REUSED, not new scans.24normalizer commands/12equality
checks95803 and all7 cross-root inventories match.

Assemblies53285/27714 match all75files/modes and74manifest bindings. Manifest
4c1140f3ef5700e348f3193dcc6a69f3c4e91d3b7943dfc911196962ccff5adf binds
exacta7d source, migration87/epoch8 and current bytes. Initial comparison assumed
an absent manifest kernel_version key and aborts before any record; corrected
source/mirror/fresh-build check confirms228 without inventing a manifest field.
Native helper build33016/record pass;2 relative-path intact invocations are
invalid_probe/exit2 and NOT acceptance. Actual required absolute-path invocation
passes. No assertion, package or time bound changes to obtain that pass.

Delivered-Linux helper10177 runs offline/read-only/cap-drop/no-new-privileges as
UID501:20 in existing593dc355...amd64 runner with the recorded AVX2 emulation
accommodation. Record/intact return0; all9 semantic mutations return exact3/
fixture_rejected. No guard fires; all4 original/private75-file trees preserve
bytes/modes, both work dirs empty and all11 task-named containers absent. Shared
builders/caches/networks preserved; no SQL services started. Manifest-bound
source/image scan30412 passes2606trackedfiles/findings0/objective true/guardfalse.

All root processes terminal; source stayed clean/frozen throughout. Ignored raw
records in release-current-a7d5a4a4/release-rebuild-a7d5a4a4/release-linux-probes-
a7d5a4a4 under.cache; small candidate-a7d5a4a4.json retained. Complete5-path
record/docs batch aftera7d needs fresh independent corrective artifact-evidence
review. Review by: Codex root (implementer first filter). Recorded by: Codex root.
Earlier exactaa5e16e8 approval remains historical, not relabelled. These identical
builds are ONE candidate, not a legitimate predecessor pair or native supported
host/R-006. Known Clout/history-guard/hostedCI, RP-770, privacy/rights/accessibility,
fullfeature/later-tier/endings and release/owner adoption gates remain open.


## 2026-11-05 — Current runtime candidate evidence independently APPROVED

Review by: fresh independent Codex `/root/current_runtime_artifact_review`.
Recorded by: Codex root, transcribing actual corrective artifact-evidence verdict.
**APPROVED**, no actionable findings, complete5-path range
`a7d5a4a4972cc7cfbdd00f1c7770f3b958a4d670` exclusive through
`beef6cff87763b0516a688942fc1199589348fa9` inclusive. Runtime/source is still
exacta7d, kernel228/migration87/epoch8; this verdict adds no runtime bytes.

Independent root-Make rebuild of6Linux binaries,7client files,32content and2
metadata files matches delivered bytes/modes. Both75-file bundles/74bindings
and actual33-file image payloads match, including blobs/uncompressed diffIDs,
owners/modes/source/version.28inventory normalizations/comparisons and6fresh
registry manifest/config/platform reads pass. Parent's2fresh Syft41-package scans
and source/image2606-file/findings0 scan are inspected and hash/source-bound,
NOT independently rescanned. Original upstream scan times/chains stay reused.

Fresh native and actual delivered-Linux consumers accept record/intact0/0;
all9semantic fixtures exact3/fixture_rejected, no guard fires. Parent4trees and
client output preserved, work empty,11review containers absent. All reviewer
processes terminal, zero tracked edits or SQL resources, shared builders/caches
preserved. One nonblocking editorial nit (“Current that kernel227”) is corrected
at ordinary closeout; no behavior/check changed. Manifest
`.cache/current-runtime-artifact-review/review-results.json`, SHA256
0af4d6220f46dc43a820c09b900c66e8332917e0e504f207c5c7d071308469c7.

Canonical/current pointers now use this exact independently approved local
artifact. Older records/verdicts remain historical; no rebuild for editorial
closeout. One candidate remains one candidate, not a legitimate predecessor pair,
R-006/native-host proof, rights/privacy/CI/full-feature or release approval.
All accepted full-nine-tier1.0 and owner/adoption gates remain open; no push,
deployment, archival or new World/Headcount adoption is authorized or inferred.


## 2026-11-09 — Preserve the incomplete kernel229 package handoff

The November5 component builds at source `85e1d158bf20e1d9eeebd660f5c9cd11b6781928`
(epoch1793876307, created2026-11-05T10:58:27Z) are revalidated today: six
Linux-amd64 executables, seven client files,32 content files and two metadata
files match bytes/modes across both roots. Twelve reused upstream inventory
chains passed24 normalizations/12 equality checks; their retained inputs and
outputs still match. The source-only scan recorded2,612 tracked files/no findings.
This is not an image or manifest-bound scan. Initial setup's SPDX-name assertion
and unsupported structured source-only invocation failed and receive no credit.

Docker's Engine socket again times out at5s. Earlier version/status checks failed;
Desktop start reported already running. Restart permission remains unanswered.
No image build, fresh image/registry observation, bundle, delivered-Linux check
or artifact review has run. Approved a7d5a4a4/kernel228 remains current.

Resume from `.cache/release-current-85e1d158/partial-checkpoint.json` and the
matching rebuild root. All consumers/reviewers are terminal; release the source
freeze for this editorial checkpoint. Revalidate source/dependency identity and
pin inputs before construction resumes. Plan/log diff and retained factual
claims checked; no runtime, acceptance, publication or deployment change.


## 2026-11-10 — Finish the local kernel 0.3.229 package after Docker recovery

Docker 28.4.0 (Linux/arm64 VM) and Compose 2.39.4-desktop.1 respond after the
authorized force-stop and relaunch. The initial 45-second startup wait timed
out while the backend was starting. Its logs recorded a module manifest fetch
timeout and an interrupted shell environment lookup. The original failure's
root cause remains unconfirmed.

Resumed the two component builds at source `85e1d158bf20e1d9eeebd660f5c9cd11b6781928`.
All later commits before this evidence were editorial planning changes. Checked
the retained binaries, client, content, metadata and image contexts before
using the pinned BuildKit builder for two separate no-cache builds. Verified
both archive digests, compressed blobs, uncompressed layer digests and all
33 payload files, including ownership and modes. Two fresh Syft scans each
found 41 packages and normalize to identical bytes. The twelve upstream
inventory chains retain their prior scan provenance. Six fresh registry
identity checks pass after correcting the local script's initial reference to
the as-yet unwritten candidate record. The failed setup produced no checks.

Both bundles match all 75 file hashes and modes. All 74 manifest bindings pass.
Manifest `sha256:497f8ce2ce90f31990404e989b6bc478b9c7a507a66b59a9c81a0a3af9f4d27e`.
Image archive `sha256:b69b21594a5edbd5bf39b45c186372d6bfd6eb396e39a4631a74487e430f75bf`.
Native and delivered Linux helpers accept the build record and intact bundle.
All nine delivered Linux tampering cases return exit 3 with
`fixture_rejected`. Four original and probe bundle trees retain their hashes
and modes, both work directories are empty and all eleven probe containers
are gone. The manifest-bound scan of an exact local source checkout and the
image archive reports 2,612 tracked files and zero findings.

The [build record](release-builds/candidate-85e1d158.json) captures this local
candidate. Raw checks remain under `.cache/release-current-85e1d158`,
`.cache/release-rebuild-85e1d158` and `.cache/release-linux-probes-85e1d158`.
Independent artifact review is pending. The older approved candidate
remains the latest reviewed package. Supported-host R-006, predecessor
compatibility and the release decision remain open.


## 2026-11-10 — Independent local package approval

Review by: fresh independent Codex `/root/kernel229_artifact_review`.
Recorded by: Codex root. **APPROVED**, no actionable findings, for all six
paths in `6afc0b5e7aaafef545b9419d14d068c3bf51d129..d36ea78a4c52266cd32fbb1f6c386a0c39eb2db8`.

The reviewer verified both no-cache BuildKit receipts, both 33-file image
payloads, four matching 75-file trees and all 74 manifest bindings. A fresh
build matched six Linux binaries, 32 content files and two metadata files.
The first client invocation ran from the repository root and changed Svelte
scope hashes. The normal client working directory reproduced all seven asset
names and six files byte for byte. The remaining source map differed only
in the relative paths required by the private output directory. All 139
entries resolve to the same sources and the contents and mappings match.

The reviewer verified six registry identity chains and twelve prior inventory
chains and ran 26 independent normalizations. Original Syft scans and registry
fetches were inspected. A fresh manifest-bound scan covered 2,612 exact-source
files and the image archive with zero findings. The delivered Linux helper
passed both acceptance cases and returned exit 3 with `fixture_rejected` for
all nine tampering cases. Review containers were removed and tracked files
were clean at the end of review.

Source `85e1d158` is now the latest independently approved local package.
Supported-host R-006, predecessor compatibility and the release decision
remain open.


## 2026-11-10 — Fix the migration contiguity witness (RP-780)

The hosted `releasepackage` test on `6afc0b5e` reports
`migration=87 err=<nil>` because it expects a literal 85. The cold focused
test reproduces that failure locally. The expected version now comes from
the number of entries in the checked-in chain, which starts at 1 and must
be contiguous. A separate fixed fixture rejects the 1, 3 chain and accepts
version 3 after 2 is added. No migration or package production code changed.

`make test-go vet GO_PACKAGES=./releasepackage` passes. An initial anchored
Make selector matched zero tests after expansion. The corrected selector
ran the failing test and the final full package run executed its replacement.
Independent review is pending.


## 2026-11-10 — Approve the migration witness correction

Review by: independent Codex `/root/migration_witness_review`.
Recorded by: Codex root. **APPROVED** for all three paths in
`c1d342c60238001ca485855202064c64bff5066a..1e649b8799a18542b42c20b288e3aab6519eb155`.

The reviewer reproduced the exact old version-87 failure using a private
overlay. The full cold release package and vet passed. An overlay that
allowed migration gaps left the live-chain positive green but failed the
new missing-version fixture. RP-780 is closed. The affected Linux package
also passed in the combined local test. Hosted CI and R-006 remain open.

## 2026-11-11 — Require explicit browser observations in R-006 evidence (RP-784)

Under Deployment AC10, the browser result must establish that the journey
reported zero failed intents and exceptions and did not exhaust its guard.
The old decoder accepted an omitted count, null observations, a case alias
and a duplicate that overwrote a failed-intent count of one. The old base-run
validator also accepted the result with the missing count after its digest
was updated in the dossier.

The shared decoder now requires each declared member exactly once, with its
exact name and a non-null value, before checking types and outcomes. Both
the browser runner and dossier verifier use it. The producer's existing
schema and serialized output stay the same.

The five new decoder negatives and the rehashed-dossier negative failed on
the old code. The corrected cold browser and rehearsal packages and their
commands pass, along with vet and diff checks. This is a local check of the
evidence readers. The approved `85e1d158` bundle predates this correction.
Review by: Codex root (implementer check).
Recorded by: Codex root. Independent corrective review pending.

## 2026-11-11 — Independent approval of the browser evidence correction

APPROVED complete six-path `c651cbde..bb8d3694`.
Review by: fresh independent Codex `/root/browser_rehearsal_evidence_review`.
Recorded by: Codex root.

An overlay of the exact old decoder reproduced all five malformed-input
acceptances and the base dossier's acceptance of the rehashed result.
The cold corrected browser and rehearsal packages and affected commands
pass, as do vet and the range diff check. The reviewer confirmed that the
runner and both dossier validators use the shared decoder and that valid
v2 producer output supplies all 21 observations. RP-784 is closed.
Artifact rebuilding and the clean-host rehearsal remain open.

## 2026-11-11 — Rebuild the local package with the shutdown and evidence fixes

Clean source `00ee0e69ea070a0ed6f7a572645a811f90506426`, kernel 0.3.229,
migration 87 and epoch 8. Separate builds matched all six Linux executables,
seven client files, 32 staged content files and two metadata outputs. Compared
with `85e1d158`, the gameserver contains the RP-781 shutdown fix and both
rehearsal executables contain the RP-784 decoder change. The generated client
and staged content match the previous package.

Both no-cache Linux/amd64 image builds used the pinned BuildKit builder and
produced identical archives. Blob, uncompressed layer and 33 payload-file
checks bind the image to the fresh executables and content, including modes
and ownership. Two fresh Syft scans each found 41 gameserver packages and
normalize to the same bytes. The six unchanged upstream image inventories
in each build retain their scan provenance. Six fresh registry identity
checks, 24 normalizations and 18 within-build and cross-build comparisons passed.

The bundles match all 75 file hashes and modes and all 74 manifest bindings.
Manifest `sha256:10f2bfb9bc454c908c175c58ab7fb8159cc70fd4a27a74a99ee6127f04403b0e`.
Image archive `sha256:5f88e986ea5192ed828332e28cf0fdba74c95001f7e6e0ca7b23a7573eac456e`.
Native and delivered Linux helpers accepted the build record and intact
package. All nine delivered Linux tamper cases returned exit 3 with
`fixture_rejected`. The four original and probe trees retained their
hashes and modes. Both scratch directories are empty and the eleven probe
containers are gone. The manifest-bound scan of 2,613 exact-source tracked
files and the new image archive returned zero findings.

The first staging attempt named the Dockerfile incorrectly and the first
native package probe supplied relative paths. Both were corrected, and
only the completed runs count. Raw evidence is under
`.cache/release-current-00ee0e69`, `.cache/release-rebuild-00ee0e69` and
`.cache/release-linux-probes-00ee0e69`. The [candidate build record](release-builds/candidate-00ee0e69.json)
is ready for independent review. The prior package remains the latest
independently approved one until that verdict. Predecessor compatibility
and R-006 on a supported clean host remain open.

## 2026-11-11 — Independent approval of the refreshed local package

APPROVED complete two-path `00ee0e69..6bb51594`.
Review by: fresh independent Codex `/root/release_refresh_evidence_review`.
Recorded by: Codex root.

The reviewer rebuilt the three changed Linux executables and matched their
delivered bytes. Six copies retained the same 75 file hashes and modes.
Independent checks covered all 74 manifest bindings, both 33-file image
payloads, seven SBOM bindings, six registry identity chains and twelve
reused inventory chains. The reviewer inspected the original Syft runs,
eleven Linux results and manifest-bound scan. Independently executed
Linux checks accepted the record and intact package and rejected both the
removed client and changed image digest with exit 3 and `fixture_rejected`.
The independent results SHA-256 is
`050933d08dfacdfc62e21a24ac629878d02dc525d3741572f6aada5305ee5ea5`.
The scratch directory is empty and the test containers are gone.

`00ee0e69` is now the latest approved local candidate. The two builds use
the same source and version. A floor-compliant predecessor, supported-host
R-006, and release approval remain open.

## 2026-11-11 — Browser startup and restart against the exact packaged runtime

The disposable `cc-package-smoke-00ee0e69` project started the approved
`00ee0e69` gameserver image, pinned Caddy and Postgres images, and the
package's unchanged Caddyfile and site. The read-only gameserver mounted
only synthetic secret files. Postgres and the gameserver published no host
ports. The test overlay bound Caddy to loopback, selected the cached amd64
image IDs and used JSON logging on Docker Desktop. Only these three services
ran. The client verified Caddy's local TLS chain and hostname before pinning
the verified leaf key in its disposable Chromium process.

Native Enter on Begin returned HTTP 201, moved focus to the Desk, committed
credentials and opened the WebSocket. Chromium loaded the exact bundled
HTML, JavaScript, CSS and prediction worker, and started the worker. A
visible Fix Computer click applied one intent and advanced authoritative
revision 1 to 2. After normal SIGTERM, the packaged gameserver exited 0.
A passive authenticated, subscribed socket through Caddy received
`server_restarting` and close 4003. The app socket called close and recorded
1005. On restart, the same page opened a new socket, read revision 2, and
committed a second visible Fix Computer action at revision 3. No external
browser requests or page exceptions were observed.

The first observer incorrectly required the app's own socket to see 4003
and failed at 1005. The second run captured both the passive socket and the
app close call. Compose's `start --wait` also refused the unsupported flag
before the successful plain start. The local driver and logs are retained
under `.cache/release-browser-smoke-00ee0e69/`. Corrected browser log SHA-256
`b5c331cecb8641f9afafdace3b09e0732ea6e7b4248c3a3b67a788e83b71f694`.

This covers packaged startup, a persisted player action and same-page
recovery on the local emulated stack. The delivered Linux browser driver,
natural progression to the next Company, backup/rollback, operations and
supported-host R-006 remain outside this observation.

## 2026-11-11 — Independent package restart rerun exposes a recovery defect (RP-785)

CHANGES REQUIRED for the one-path evidence range `ef5bbb83..5a71dbe0`.
Review by: fresh independent Codex `/root/package_browser_restart_evidence_review`.
Recorded by: Codex root.

The exact independent driver reproduced bootstrap, revision 2, graceful
gameserver exit 0 and passive courtesy/4003, then exited 1 because the app
failed to reconnect within 60 seconds. The original recovery result remains
one successful observation. Root's subsequent exact-package wire capture
showed an omitted publication offset, envelope revision 0 and an advertised
15,000 ms delay. The app called close and issued a state read that returned
502 during shutdown. No new socket opened within the 20-second diagnostic
guard. The independent log hash is
`1162b2231408432d779e6dee43a1d80b4b13eff3c6963b7fec5d4844c68baf73`.
The wire diagnostic log hash is
`d9bc8f9f7e22650c994eab698d199562f8ee4bd7ba7b44b7b5b3eeecebe7aaf2`.

The previous no-external-requests statement covers the routed requests
of the test browser context and the page's WebSocket events. The test did
not inspect all process or container egress.

With only the corrected built client substituted at Caddy's site mount,
both sockets receive close 4003 and the app makes no local close call.
No snapshot read returns an error. The browser opens a new socket
15,019 ms after the courtesy frame, retains revision 2 and commits the
next visible action at revision 3 on the same page. The fixed log hash is
`ef1a26e0f98d9afc69a58a0d0450d66cc562cb56343566a304dce23d802f9cb4`.
The exact `00ee0e69` bundle still contains the old client. A fresh bundle
and rerun on its unchanged site remain necessary. The Transport log owns
the corrective review.

## 2026-11-11 — Refresh the candidate and verify restart on the bundled site (RP-785)

Two builds from clean source `2f366a907ffd2b7bb8aee68419df9eb73ac6ff1e`
produced the same 75 file hashes and modes. All 74 manifest bindings pass.
The six Linux executables match `00ee0e69`. The seven client files match
the independently reviewed RP-785 build. Both no-cache image archives
match, with verified blob digests, layer diff IDs and 33 payload files.
Two new Syft runs each find 41 gameserver packages. All seven normalized
inventories match across builds. The unchanged upstream inventories retain
their original scans, with six fresh registry identity checks and 24
normalizer executions.

Manifest `sha256:b082b7b19b6f5c513fa72319dee40933ffb82996e408e4a9b037b37f7725470b`.
Image archive `sha256:9a9c5682c14306c31916c156442c6e976dcaad17484dc75d8a935bf35bb090fd`.
The native and delivered Linux validators accept the record and intact
package. All nine Linux tamper cases return exit 3 with `fixture_rejected`.
The manifest-bound scan covers 2,614 tracked source files and the new
gameserver image with zero findings. Both build trees and both probe copies
retain their hashes and modes.

The loopback test used the candidate's image, Caddyfile and read-only site,
with the pinned Caddy and Postgres image IDs. Its local override only
selects cached image IDs, amd64 execution, loopback publication and JSON
logging. The real database applies migration 87. Chromium loaded and
verified the exact bundled HTML, CSS, entry script and prediction worker
over TLS. Both sockets received the live courtesy notice with the omitted
offset and closed with 4003. The app made no local close call. It reopened
15,021 ms after the notice, with 15,000 ms advertised and all observed
state reads returning 200. The same page recovered revision 2 and committed
revision 3 through a visible action.

The [build record](release-builds/candidate-2f366a90.json) and browser test
are ready for independent review. Browser log SHA-256
`1879401df0e8aff226d5d2027baaac2ccb09accefb035f823ca0675e518ee286`.
The driver and raw evidence are in the ignored
`.cache/release-browser-smoke-2f366a90`, `.cache/release-current-2f366a90`
and matching rebuild/probe directories. This is local Chromium evidence
on the exact package, with emulated amd64 services. The compliant
predecessor, supported-host R-006 and full player/release acceptance remain open.

## 2026-11-11 — Independent approval of the packaged restart evidence

APPROVED complete two-path `2f366a90..f90e2a2d`.
Review by: fresh independent Codex `/root/packaged_restart_release_review`.
Recorded by: Codex root.

The reviewer rebuilt the standard client and matched all seven files,
checked six 75-file trees and all 74 manifest bindings, and verified both
image payloads, the six unchanged executables and the seven SBOM bindings.
The reviewer inspected the root scans, registry and inventory chains, and
the eleven original probe results. The delivered Linux helper independently
accepted the record and rejected a removed client with exit 3. An initial
reviewer probe on separate bind mounts refused its setup, so the corrected
run used one bind for the scratch directory and test copy.

The independent Chromium run served the same read-only packaged site.
Both sockets closed with 4003, all observed state reads returned 200, and
the app reopened after 15,014 ms and committed revision 3 on the same page.
Browser log SHA-256
`b933b37009c87543989d41d23150a87e91ae2ece45ad06a64e35cc93f2768077`.
The disposable test containers, networks and synthetic-data volumes were
removed. RP-785 is closed. `2f366a90` is the latest independently reviewed
local candidate. Supported-host R-006 and predecessor compatibility remain open.
