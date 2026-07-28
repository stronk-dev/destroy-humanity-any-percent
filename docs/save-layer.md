# Save Layer

The server persists authoritative economy state in Postgres 16 as owner-aware revision streams.
Company and Founder streams belong to a Founder UUID, Guild streams to a Guild UUID, and World
streams to a World UUID. Invalid owner/scope pairings are rejected in Go and by database checks.

`save_streams` stores identity and archive state. `save_revisions` stores immutable numbered
snapshots. Each write locks its stream, compares the caller's expected revision, inserts exactly
the next revision, and prunes within the same transaction so only the latest five remain. Two
writers with one expected revision produce one commit and one typed conflict.

## State format

Version 1 is strict JSON:

```json
{"balances":{"company.cash":"1.23e4"}}
```

Every in-scope catalog resource appears exactly once. Decimal values are RFC-0001 canonical
strings; JSON numbers, non-finite strings, unknown/missing resources, scope mismatches, minima and
hardcap violations, unknown fields, trailing JSON, and future save versions reject the whole load.

`constants_hash` is `sha256:` plus the lowercase SHA-256 digest of the exact catalog artifact
bytes. Saves resolve that immutable catalog before restoration. Reformatting a catalog therefore
changes its identity deliberately.

The economy package exposes `RestoreLedger` as a validated constructor, not a balance setter.
Restoration creates a ledger only after every persisted balance passes the catalog invariants.

## Migrations and verification

Sequential SQL migrations are embedded in the Go package and applied transactionally through
Goose 3.27.1 using pgx/v5's `database/sql` driver. There is no ORM, SQLite substitute, runtime
migration directory, or separate migration artifact.

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
