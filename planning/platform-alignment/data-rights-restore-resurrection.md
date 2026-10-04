# RP-125 — retained current-HEAD backup/delete/restore witness

**Coordinate:** product source `54ba0124`, 2026-10-04. **Status:** executed repository-level
behavior observation, not a retention ruling or clean-host release proof. The retained test is
`TestPreDeletionBackupRestoresDeletedAccountIntegration` in
`server/deploymentbackup/postgres_integration_test.go`.

## Population and result

The current backup integration fixture migrates real Postgres 16 and seeds an account, active
Founder link, two save streams, one verified board row, epoch and a separate bystander account.
Using the real backup functions, the test creates encrypted **B_pre**, calls
`account.Repository.DeleteAccount`, creates encrypted **B_post**, and restores each backup into a
freshly recreated target database with real `pg_restore` and the age identity. It checks the
source after deletion and both restored targets, not merely the backup header.

| State | Account | Active/archived Founder link | Live/archived streams | Board row | Bystander |
|---|---:|---:|---:|---:|---:|
| Source before deletion | 1 | 1 / 0 | 2 / 0 | 1 | 1 |
| Source after deletion | 0 | 0 / 1 | 0 / 2 | 1 | 1 |
| Restore B_pre | **1** | **1 / 0** | **2 / 0** | 1 | 1 |
| Restore B_post | 0 | 0 / 1 | 0 / 2 | 1 | 1 |

Both restore arms and the four pre-existing backup integration tests passed cold with
`-test.count=1 -test.v`, without skips. The B_post arm rules out an unconditional restore-side
account creation: resurrection is attributable to restoring an older complete dump. A temporary
mutation changed Account's `DELETE FROM accounts` to a no-op update. The new test failed at the
source post-delete census (`account:1`, archived/unlinked Founder link absent); the production
byte was restored, verified by `git diff --exit-code`, and the full suite passed again.

The test used a uniquely named Docker Compose project and an out-of-tree override for the cached
ARM64 Postgres image and ARM64 test binary. The repository's pinned release/backup test platform
is linux/amd64, which this host cannot execute; this run is **not** supported-host R-006 or R-007
evidence. The ARM64 service emitted the harmless inherited `GODEBUG: unknown cpu feature "avx2"`
line, then executed every declared test. No shared Docker project or production database was used.

## Relationship to Claude's older diagnostic and limits

Claude's unmerged `b27ca321` diagnostic executed a richer repository-level population at product
`a64ed0e`: real bootstrap and session rows, original access-token authentication, recovery-code
session creation and byte-identical bootstrap receipt replay returned after B_pre restore. It
removed its temporary test. Those are **historical executed findings**, not results of this
retained current-HEAD test. This test uses seeded account/board data, not bootstrap/session
creation or the HTTP deletion route; it does not prove credentials revive today. It does not
compose Guild deletion participation, a production Founder initializer, live gameplay, a real
off-host backup target, or a clean supported release host. It does prove the current full-dump
restore can re-create an account row that the real Account repository deleted. There is no
post-restore deletion replay in this exercised path.

D-009/D-015 must decide the backup lag, protected-backup exception, post-restore reconciliation
and disclosure before a release claim. A later accepted Account/Deployment rights contract and
clean-host witness must cover the chosen policy; this test does not choose it.
