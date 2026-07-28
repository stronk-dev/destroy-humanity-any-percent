# Save Layer

The server persists authoritative economy state in Postgres 16 as owner-aware revision streams.
Company and Founder streams belong to a Founder UUID, Guild streams to a Guild UUID, and World
streams to a World UUID. Invalid owner/scope pairings are rejected in Go and by database checks.

`save_streams` stores identity and archive state. `save_revisions` stores immutable numbered
snapshots. Each write locks its stream, compares the caller's expected revision, inserts exactly
the next revision, and prunes within the same transaction so only the latest five remain. Two
writers with one expected revision produce one commit and one typed conflict.

Archive uses the same locked-head transaction: it locks the stream row, reads and compares the
latest revision, then marks the stream archived. If a concurrent write advances the head first,
archive returns a conflict; if archive wins, the writer returns archived. It never relies on a
stale scalar subquery under PostgreSQL READ COMMITTED.

## State format

Version 2 is strict JSON:

```json
{
  "balances": {"company.cash":"1.23e4"},
  "generators": {"generator.example":7},
  "evaluated_through": "2026-07-28T08:00:00Z"
}
```

Every in-scope catalog resource appears exactly once. Decimal values are RFC-0001 canonical
strings; JSON numbers, non-finite strings, unknown/missing resources, scope mismatches, minima and
hardcap violations, unknown fields, trailing JSON, and future save versions reject the whole load.

`generators` contains exactly the production-capable generator classes in the stream's catalog
scope. Counts are JSON integers from zero through `9,007,199,254,740,991`; they are not Decimals.
`evaluated_through` is canonical UTC RFC3339Nano and records the server-authored instant through
which production state has been evaluated. The repository accepts and returns one complete state
object containing ledger, counts, and cursor.

`constants_hash` is `sha256:` plus the lowercase SHA-256 digest of the exact catalog artifact
bytes. Saves resolve that immutable catalog before restoration. Reformatting a catalog therefore
changes its identity deliberately.

The economy package exposes `RestoreLedger` as a validated constructor, not a balance setter.
The save package's `RestoreState` adds exact generator/cursor validation around it.

## Migrations and verification

Sequential SQL migrations are embedded in the Go package and applied transactionally through
Goose 3.27.1 using pgx/v5's `database/sql` driver. There is no ORM, SQLite substitute, runtime
migration directory, or separate migration artifact.

Save version 1 remains readable. Its first real migration initializes every in-scope,
production-capable generator to zero and uses the revision's database-authored `created_at` as the
cursor. The checked-in `testdata/save-migrations.json` corpus fixes that result and grows with every
future save version; migrations never read the wall clock implicitly.

Run the disposable Postgres integration suite locally:

```sh
docker compose -f compose.save-test.yml up -d --wait
TEST_DATABASE_URL='postgres://cloud_clicker:cloud_clicker_test@127.0.0.1:55432/cloud_clicker_test?sslmode=disable' make test-save-integration
docker compose -f compose.save-test.yml down
```

The CI server job supplies the same variable through a Postgres 16 service container. Normal unit
tests skip database integration when the variable is absent; `make test-save-integration` fails
immediately if it is absent.

Archiving is read-only and reversible by future account policy; the persistence API exposes no
hard delete. New-Founder gameplay, save cadence, event auditing, offline progress, and leaderboard
Balance Epoch policy remain separate RFC responsibilities.
