# D-008/D-009/D-015 — current data-rights inventory plan

## Restored credential lifetime supplement — 2026-11-05 at 0f3a5318

Extend the approved composed witness with two clock offsets,16 minutes and31
days after issuance, over BOTH B_pre and B_post. This isolates Account D2's
15-minute access/30-day refresh lifetimes from D5 deletion and DP6 full-dump
restore. All six credential families must still match the selected dump before
startup; then run actual composed startup/cleanup with the advanced application
clock. Obtain a fresh bystander session through its original recovery code and
require Founder200, so original-token expiry cannot make every result look denied.

Observe the original subject access, refresh, recovery and bootstrap retry. Old
access must refuse at both offsets; old refresh must refuse after31 days.
Recovery and the unexpired refresh may return200 or401, but every200 must yield
a fresh usable pair for the original Founder. A bootstrap201 must be the exact
original response; it may legitimately contain an already-expired access token,
so do not mislabel response replay as usable access. At31 days it must return
the existing expired-receipt conflict. B_post must deny all old credentials,
retain that conflict and create no replacement account. Reuse wrong-code/token/
malformed-key controls. Other statuses, unready service, missing populations or
incomplete execution invalidate measurement; keep guards unchanged. Log only
counts/statuses and coordinate, no secret/raw bodies. No key rotation, new
retention/expiry policy, browser renewal, wall-time31-day wait, release artifact,
off-host or R-006 proof. This answers whether credential expiration alone removes
the observed pre-deletion-restore access, without prescribing privacy policy.

## Composed bootstrap restore supplement — 2026-11-05 at 9a4bf86f

Question: does the actual gameserver, with repository epoch/catalogs, Founder
initialization and Guild deletion participation, expose the same credential
revival, and does its protected bootstrap retry survive the selected restore?
Account D5/AC6 and Deployment DP6/AC4 authorize observation, not a policy change.
Use a separate disposable Compose project with real Postgres16 tools and a
compiled gameserver test binary; mount only repository balance/changelog/
moderation content read-only. No product, migration, balance or CI changes.

Create subject and bystander through actual composed `/api/v1/bootstrap`,
retain each exact response privately, and prove Founder access and initial
Company state. Freeze application time within token validity, keep synthetic
signing/bootstrap/cursor material unchanged, and record actual epoch/migration.
Take encrypted B_pre, HTTP DELETE204, observe denial, then B_post. Restore both
into a fresh declared target. Before composition/HTTP can alter the population,
compare six SQL families exactly: accounts, Founder links, session families,
sessions, access tokens and bootstrap receipts. Start a new composed service,
require readiness and healthy bystander access, then observe original subject
access/refresh/recovery and original bootstrap retry. Any successful credential
must reach the original Founder; a successful retry must reproduce the exact
original response and reach that Founder. An absent account must deny all old
credentials and return the existing `bootstrap_expired` conflict for the
tombstoned retry, without creating another account. Wrong recovery/access and
malformed bootstrap keys supply negative controls. Accept only documented
success/unauthorized/conflict observations; transport, readiness, other status,
missing population or incomplete restore invalidates measurement.

Log counts/statuses and public content coordinate only, never credentials,
identifiers, SQL rows, raw responses, keys or database URLs. Use the existing
synthetic backup manifest builder with the actual epoch bytes/migration and
actual content hash substituted; do not call it a built release manifest.
This can close the constructor-only/bootstrap-retry evidence gap, not expiry,
rotation, gameplay actions, populated Guild membership, Caddy/off-host operation,
RPO/RTO, exact artifact/clean-host R-006 or D-009/D-015. Preserve failures as
findings; no assertion that revival is desirable, no deletion-policy repair.

## HTTP credential restore supplement — 2026-10-15 at 42ce00f0

Extend RP-125 with actual Account HTTP create/session/delete and real encrypted
pg_dump/pg_restore. Use the backup package's constructor-only catalog and fixture
manifest/epoch, synthetic unchanged signing keys and a frozen valid-token clock;
do not claim the gameserver's foundation initializer, Guild participation, real
gameplay, key rotation/expiry, Caddy or an exact clean-host release artifact.

Create a subject and healthy bystander through HTTP, retain credentials privately,
take B_pre, DELETE the subject through HTTP, verify its account is gone and old
access/refresh/recovery are refused, then take B_post. Restore each into the
declared fresh test target and rebuild its Account API. Observe old subject
access/refresh/recovery outcomes and account presence; preserve a successful
bystander read in both arms and reject wrong credentials. Any revived credential
must reach the original Founder, not merely return200. Admit only actual200/401
observations; other statuses invalidate the measurement. Report counts/statuses,
never IDs, credentials, raw bodies, URLs or keys. This records behavior, not a
requirement that old credentials should revive or a retention/deletion ruling.
Bind the restored account/Founder/session-family/session/access-token SQL rows
to the actual pre/post snapshot before HTTP observations can rotate/create tokens.

Predeclared 2026-09-23. This is desk research for owner decisions, not an accepted retention
policy, implementation authority, privacy notice or legal conclusion. The product coordinate and
exact migration endpoint must be recorded in the result before any current-state claim.

## Question and population

What player-, account-, operator- and service-linked data does the current production topology
actually persist, export, anonymize, delete or retain? Inventory every table created by the
current Postgres migration chain, plus client credential storage, release backups, application and
security logs, metrics, and any filesystem/object output the production composition creates.
Do not count test-only Compose services or proposed product features as current storage.

## Method

1. Derive the table/view denominator from the ordered migration chain and verify the chain's
   current endpoint with the migration test. Record renames/drops and group only after each live
   relation has a named row or explicit group membership.
2. For each family trace writer, joinable identifier, actual production reader/consumer,
   account-deletion behavior, export behavior, cleanup trigger/owner, and current retention bound.
   Mark `unknown` where source does not prove a bound; never infer a retention period from a
   method name, documentation promise or intended future job.
3. Trace non-Postgres storage separately: browser credentials, backup outputs/retention,
   journald/application/security logs and private metrics. Distinguish a configured operator
   mechanism from a completed rehearsal.
4. Reconcile the dated `account-rights-release-audit.md` and
   `operations-retention-preservation-audit.md` against HEAD; preserve historical negative findings
   only when their producer/consumer trace still holds.

## Controls and exit

The inventory must include the known short-lived access/refresh/bootstrap families and the
append-only run/founder/event families. It must explicitly test whether the existing
`PruneIntentRecords` primitive has a production scheduler, and whether any mounted player export
route/client action exists. Missing either family class or counting a package-only primitive as
a composed workflow invalidates the result. Every live relation must be counted exactly once or
declared as a named member of a grouped row; ambiguous historical DDL is a visible limitation.

The result may supply the evidence packet for D-008 (export scope), D-009 (deletion disclosure)
and D-015 (retention), and identify bounded RFC dependencies. It cannot choose legal basis,
retention durations, immutable-history disposition, player-facing copy or any cleanup mechanism.
R-003 and R-007 participant/operator rehearsal remain blocked on their named decisions and built
workflows; this inventory alone cannot close them.
