# D-008/D-009/D-015 — current data-rights inventory plan

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
