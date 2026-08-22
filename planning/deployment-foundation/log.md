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
