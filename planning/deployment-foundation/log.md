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

