# RP-125 — retained current-HEAD backup/delete/restore witness

## Restored credential lifetime observation — 2026-11-05

Baseline `0f3a5318`; method in the predeclared lifetime supplement. The same
actual composed/bootstrap/HTTP-delete/encrypted-restore witness now restores each
selected dump at application-clock offsets0,16 minutes and31 days. Six credential
families match before each startup; the aged bystander logs in afresh with its
recovery code and must reach its original Founder. This prevents universal
expired-token denial from masquerading as deletion protection.

| Snapshot | Time after issue | Original access | Original refresh | Original recovery | Bootstrap retry | Bystander |
|---|---|---:|---:|---:|---:|---:|
| B_pre | 0 | 200 | 200 | 200 | 201 | 200 |
| B_pre | 16 minutes | 401 | **200** | **200** | 201 | 200 |
| B_pre | 31 days | 401 | 401 | **200** | 409 | 200 |
| B_post | 0 | 401 | 401 | 401 | 409 | 200 |
| B_post | 16 minutes | 401 | 401 | 401 | 409 | 200 |
| B_post | 31 days | 401 | 401 | 401 | 409 | 200 |

Every successful refresh/recovery returns a fresh pair reaching the original
Founder, with no extra account. At16 minutes, bootstrap201 is still the exact
original response but its original access token is expired401: replayed bytes
are not proof of usable access. At31 days, the receipt expires409 while the
original recovery code still opens the restored pre-deletion account. Thus
waiting out access/refresh expiry alone does not prevent the observed recovery.
This proves that population, not permanent validity or a retention policy.

Cold `make test-account-restore`20771 passes6.68s, all seven observed states
(deleted source plus six restores), zero skips. Wrong recovery/access and
malformed bootstrap controls remain. Native93433/vet pass with manual/DB skips
excluded. Same actual epoch8/migration85/content hash as the initial composed
study below. Linux-amd64 runs locally through Docker Desktop; application clock
offsets are not a31-day wall-time or native clean-host rehearsal. Keys are
unchanged, manifest metadata remains synthetic, and key rotation, other offsets,
populated Guild membership, browser renewal, off-host/RPO/RTO/R-006 and policy
remain separate. No production, balance, migration, schema, deadline or CI change.
The owning Deployment log records review status and cleanup; RP-125 remains open.

Independent corrective review APPROVES the complete six-path
`0f3a5318..6df67ab2` supplement. Fresh real run85435 passes7.31s with all seven
states/zero skips and reproduces the full matrix; native6871/vet/format pass,
with15 top-level native manual/DB skips excluded. Exact provenance and cleanup
are in the owning log; no policy, feature or release acceptance follows.

## Initial composed bootstrap observation — 2026-11-05

Baseline `9a4bf86f`; predeclared composed supplement in the inventory plan.
`make test-account-restore` executes
`TestComposedBootstrapDeletionRestoreIntegration` through actual gameserver
composition, repository epoch/catalogs, Founder initialization and the attached
Guild deletion participant. Both accounts originate in `/api/v1/bootstrap`;
their original credentials reach their Founder and the initial Company state
is readable under the actual content hash. Six selected SQL table populations,
now including bootstrap receipts, match before the restored service starts.

| Observation | Subject account | Original access | Original refresh | Original recovery | Original bootstrap retry | Bystander |
|---|---:|---:|---:|---:|---:|---:|
| Deleted source | 0 | 401 | 401 | 401 | 409 `bootstrap_expired` | 200 |
| Restore B_pre | 1 | **200** | **200** | **200** | **201, exact original response** | 200 |
| Restore B_post | 0 | 401 | 401 | 401 | 409 `bootstrap_expired` | 200 |

The successful retry's original access token and every returned credential pair
reach the original Founder. Wrong code/token and malformed bootstrap keys refuse;
no observation creates an extra account. The observer permits future credential/
retry refusal rather than prescribing resurrection. Only counts/statuses and the
public content coordinate are logged. Final cold Linux-amd64/Postgres16 run42376
passes all three populations in3.64s, no skips. Epoch8, migration85, actual hash
`sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6`.
Missing source URL with the manual lane enabled fails immediately (8873, exit1),
rather than skipping. Native package42103/vet pass; DB/manual populations skip
natively and receive no DB credit. The owning Deployment log records setup
limitations, cleanup and independent review status.

Independent corrective review APPROVES the complete eight-path
`9a4bf86f..8c7ccf17` range. Fresh real run75053 reproduces every arm in4.61s,
zero skips; missing-URL control fails1, native87025/vet pass with manual/DB
skips excluded. The owning log holds exact provenance, scope and cleanup.

This removes the constructor-only and unretained bootstrap-retry evidence gaps.
The backup manifest is still synthetic metadata with actual content/epoch/migration
substituted; this is not the built release artifact or R-006. Synthetic keys stay
unchanged and time stays within token validity. No expiry/rotation, gameplay action,
populated Guild membership, Caddy, off-host storage or RPO/RTO proof. No product/
balance/migration/schema/policy change; D-009/D-015 and RP-125 remain open.

## Current HTTP credential observation — 2026-10-15

Baseline `42ce00f0`; method in the HTTP credential supplement of
[`data-rights-inventory-plan.md`](data-rights-inventory-plan.md).
`TestHTTPDeletionBackupRestoresCredentialAccessIntegration` creates the subject,
bystander and both credential pairs through actual Account HTTP routes. It takes
encrypted B_pre, deletes the subject through HTTP, verifies account removal and
credential denial, then takes B_post. Real pg_dump/age/pg_restore restores each
into the declared fresh test target; a newly constructed Account API reads it.

| Observation | Subject account | Original access | Original refresh | Original recovery code | Bystander access |
|---|---:|---:|---:|---:|---:|
| Source after HTTP DELETE204 | 0 | 401 | 401 | 401 | 200 |
| Restore B_pre | 1 | **200** | **200** | **200** | 200 |
| Restore B_post | 0 | 401 | 401 | 401 | 200 |

Every revived credential reaches the original Founder; a200 alone is insufficient.
Wrong recovery/access credentials refuse in every observed state. Before restored
HTTP can rotate/create tokens, five ordered SQL populations (accounts, Founder
links, session families, sessions and access tokens) must match the actual selected
pre/post snapshot exactly. Private rows/keys/codes/tokens stay in memory or owned
secret fixtures; logs contain counts/statuses only. The observation permits future
denial of restored credentials rather than requiring revival as desired policy.

Final `make test-deployment-backup`95335 passes all eight integration functions
cold on the declared Linux-amd64/Postgres16 service, no skips; new HTTP observation
2.50s. Initial36645 also passes before adding exact credential-row binding. Final
native package/vet78814 pass; the owning Deployment log holds review/cleanup.
The seven original integration functions remain intact.

The catalog is the backup fixture's constructor-only resolver, the epoch/manifest
are fixtures, signing/bootstrap keys are synthetic and unchanged, and the clock is
frozen within token validity. This does not prove revival after expiry/key rotation,
the real gameserver foundation/Guild initializer, gameplay, Caddy, off-host storage
or a supported clean-host release. No production/policy/migration/schema changes.
It does establish current HTTP credential revival after restoring an older full
dump, not merely account-row resurrection. Deletion reconciliation after restore
remains absent and D-009/D-015 remain unruled; no legal conclusion is inferred.

## Historical repository-level observation — 2026-10-04

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
