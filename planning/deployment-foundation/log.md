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
