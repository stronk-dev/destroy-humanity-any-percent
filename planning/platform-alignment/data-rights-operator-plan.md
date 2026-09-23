# D-008/D-009/D-015 — operator-persisted data plan

Status: predeclared research only; no host operation, secret read or policy adoption.
Coordinate: source at current HEAD, after core payload checkpoint `dfdd6f4`.

## Question and population

Trace the persisted operator surfaces named in `data-rights-inventory.md`:
encrypted off-host backup objects and manifests, application/security journal,
private metrics/alerts, release/rollback/rotation records and recovery identity
artifacts. For each, identify the production writer/config, data fields or
allowlist, sink, configured retention and whether account deletion propagates.
Distinguish a configured floor from a clean-host executed proof (R-006/R-007).

## Method and controls

1. Read the accepted Deployment contract and current code/config for backup,
   log/metrics and release artifact output. Do not access live credentials,
   private backups, journal or production services; source and synthetic tests
   are sufficient for this bounded research tranche.
2. Execute local no-secret validation/tests for one backup lifecycle path and
   one logging/metrics/manifest contract where relevant. State whether they
   run against the supported clean host and exact release manifest.
3. Positive control: retained backup snapshots can contain a deleted account
   even if live DB deletion succeeds. Negative control: a filename, private
   network or metric-label allowlist does not prove every log/artifact is free
   of player data. Never infer an account-erasure deadline from backup cadence
   or retention alone.

## Exit and refusal

Publish a bounded operator-surface map with source paths, configured versus
executed evidence and unresolved rights disclosure. Update parent inventory,
backlog if a new contradiction is found, D-008/D-009/D-015, 1.0 board and
append-only logs. This cannot set product-data retention, access policy, legal
basis, restore-time deletion replay or a public claim. No product, test,
deployment, migration, accepted RFC or authored-content edit in this wave.
