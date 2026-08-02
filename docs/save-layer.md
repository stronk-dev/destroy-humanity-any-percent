# Save Layer

The server persists authoritative economy state in Postgres 16 as owner-aware revision streams.
Company and Founder streams belong to a Founder UUID, Guild streams to a Guild UUID, and World
streams to a World UUID. Invalid owner/scope pairings are rejected in Go and by database checks.

`save_streams` stores identity and archive state. `save_revisions` stores immutable numbered
snapshots. Each write locks its stream, compares the caller's expected revision, inserts exactly
the next revision, and prunes within the same transaction so only the latest five remain. Two
writers with one expected revision produce one commit and one typed conflict.

`run_genesis` preserves the unprunable first revision of every pinned Company run as bytea plus its
save version and constants hash. Its bytes come from `INSERT ... RETURNING state::text`, so genesis
and the committed `jsonb` revision share PostgreSQL's actual representation rather than merely
decoding to equal objects. A deferred constraint requires one genesis per pin, and the standard
forensic trigger makes genesis update/delete impossible.

Archive uses the same locked-head transaction: it locks the stream row, reads and compares the
latest revision, then marks the stream archived. If a concurrent write advances the head first,
archive returns a conflict; if archive wins, the writer returns archived. It never relies on a
stale scalar subquery under PostgreSQL READ COMMITTED.

## State format

Version 12 is strict JSON. It contains the economy, production, Routes, Commons, Prestige, Faction, and Guild fields
described by their owning canonical docs; unknown or missing required fields remain invalid. A
representative prefix is:

```json
{
  "balances": {"company.cash":"1.23e4"},
  "generators": {"generator.example":7},
  "evaluated_through": "2026-07-28T08:00:00Z",
  "compute_credit_ms": 0,
  "manual_token_milli": 50000,
  "manual_token_refilled_at": "2026-07-28T08:00:00Z",
  "gates_crossed": {"gate.t2_to_t3":true},
  "run_seq": 1,
  "doctrines_by_transition": {},
  "structure_id": "",
  "ledger_fact_kinds": [],
  "meter_bands": {},
  "region_traits": [],
  "route_knowledge_balance": 0,
  "hints_unlocked": [],
  "compact_member": false,
  "compact_tithe_ppm": 0,
  "compact_solidarity_ppm": 0,
  "compact_solidarity_samples": [],
  "run_started_at_ms": 1785326400000,
  "run_pre_timer": false,
  "offline_spans": [],
  "collapsed_offline_ms": 0,
  "faction_id": null,
  "incorporated_at_ms": null,
  "stock_units": 0,
  "stock_progress_ms": 0,
  "consumed_stock_units": 0
}
```

Every in-scope catalog resource appears exactly once. Decimal values are RFC-0001 canonical
strings; JSON numbers, non-finite strings, unknown/missing resources, scope mismatches, minima and
hardcap violations, unknown fields, trailing JSON, and future save versions reject the whole load.

`generators` contains exactly the production-capable generator classes in the stream's catalog
scope. Counts are JSON integers from zero through `9,007,199,254,740,991`; they are not Decimals.
`evaluated_through` is a canonical UTC RFC3339 instant on an exact whole-millisecond boundary and
records the server-authored instant through which production state has been evaluated. The
repository accepts and returns one complete state object containing ledger, counts, and cursor.

`compute_credit_ms` and `manual_token_milli` are non-negative exact safe integers capped by the
state's immutable catalog policy. The manual refill cursor uses the same exact UTC-millisecond
domain and may not exceed `evaluated_through`. `save.CanonicalServerTime(t)`—UTC plus truncation to
a millisecond—is the shared constructor. Encoding rejects non-UTC or sub-millisecond caller state
rather than silently rewriting it. New company streams require both cursors to start at the same
canonical instant. These fields are company-scoped; another save scope cannot smuggle
production-policy state.

Route state is scope-checked. Company streams own crossed gates, a positive exact `run_seq`, and
the committed predicate context (doctrines, structure, ledger facts, meter bands, region traits).
Founder streams own the exact non-negative Route Knowledge balance and permanent unlocked-hint
set. Guild and World streams cannot carry either. Sets serialize as sorted mechanical-ID arrays;
maps reject invalid IDs, false membership entries, duplicate list entries, and meter values outside
0–100.

Compact membership is company-scoped. Tithe and Solidarity are integer ppm. A non-member must
have zero tithe, zero Solidarity, and no samples. Member samples are UTC whole-hour buckets with
bounded compliance and positive coverage no greater than one hour; they serialize in strictly
increasing order. Leaving clears the complete window.

Faction identity and interdependence stock are Company-only. An unincorporated run has null
faction/time and all three stock integers at zero. An incorporated run stores a mechanical
faction id, a canonical UTC whole-millisecond incorporation time no later than the evaluation
cursor, non-negative stock units, the carried millisecond remainder, and received stock units.
The stock resource is derived from the immutable faction catalog and is not persisted as a second
source of truth. Catalog-aware persistence enforces the catalog cap and interval.

`constants_hash` is `sha256:` plus the lowercase SHA-256 digest of a deterministically ordered,
named artifact bundle. Each artifact name and byte length is framed before its exact bytes, so
iteration and concatenation cannot make two bundles collide structurally. Phase-0 production
binds the Commons, economy, Prestige, Routes, and faction catalogs. Saves resolve that immutable
bundle before restoration; reformatting any catalog therefore changes its identity deliberately. Economy-only unit/store
fixtures retain the single-artifact helper where no Commons policy participates.

The economy package exposes `RestoreLedger` as a validated constructor, not a balance setter.
The save package's `RestoreState` adds exact generator/cursor validation around it.

## Migrations and verification

Sequential SQL migrations are embedded in the Go package and applied transactionally through
Goose 3.27.1 using pgx/v5's `database/sql` driver. There is no ORM, SQLite substitute, runtime
migration directory, or separate migration artifact.

Save versions 1 through 7 remain readable. V1 initializes every in-scope production-capable
generator to zero and uses the revision's database-authored `created_at` as the evaluation cursor.
Versions 1 and 2 initialize Compute Credits to zero, fill the manual bucket from the resolved
catalog, and use the evaluation cursor as the refill baseline. During v1–v3 restoration, both
cursor instants are independently floored to UTC whole milliseconds before their ordering is
validated; no whole millisecond of work is invented. A claimed v4 snapshot with either
sub-millisecond cursor is rejected. Pre-v5 company saves receive empty gate/context state and
`run_seq = 1`; pre-v5 Founder saves receive a zero balance and no hints. Pre-v6 saves receive
empty non-member Compact state. Pre-v7 Company saves backfill `run_started_at_ms` from their
authoritative `evaluated_through` cursor and set `run_pre_timer=true`; the marker preserves
playability while excluding those historical runs from time ranking. V8 saves initialize
`collapsed_offline_ms` to zero; V9 moves evicted offline-span durations into that exact integer
accumulator so bounded history does not alter Attended Time. V10 adds null faction identity and
zero stock state in both Company and Founder migration fixtures; Founder scope rejects any later
faction leakage. V11 adds the exact Guild tithe carry, clearing sequence, and consumed-window
counter. V12 pairs that clearing sequence with its Guild UUID; a restored v11 bare sequence stays
explicitly legacy until the Guild resolver binds it to the account's active membership, and cannot
be re-encoded as a current save while unpaired. The checked-in
`testdata/save-migrations.json` corpus fixes v1/v2 upgrades plus phase-matched, phase-mismatched,
boundary, route-default, lying-v4, founder-v6, and company-v6 pre-timer cases; migrations never
read the wall clock implicitly. Its `corpus_version` is metadata, not a save version. A separate
baseline manifest makes required case names and the exact case count a server-test gate, so an
addition or removal requires an explicit reviewed baseline ratchet.

## Intent and event transaction

`intent_records` keys normalized receipts by `(stream_id,intent_id)` with a SHA-256 canonical
request hash. `events` stores the closed v1 event registry with stream, originating revision,
schema version, intent, `constants_hash`, timestamp, database-authored `event_seq`, and strict
payload. The sequence is the cross-stream committed order used for idempotent Exit replay. It
deliberately does not
foreign-key the revision to a snapshot row because snapshot retention prunes old rows while event
history remains append-only. The origin revision number and constants hash remain immutable event
identity; retention does not promise that the corresponding historical snapshot is still
queryable.

The routes migration expands the closed event-kind constraint and adds idempotent projection
tables for per-event execution, founder career counts, the cached Route Knowledge balance, hint
debits, and the public Registry's first-executor/name state. Projection rows are rebuildable; the
company events remain authoritative.

Commons migrations add membership/cohort/sample/Health projections, once-per-Founder recruitment,
and the closed Commons event family. These rows are rebuildable from immutable company events;
save v6 remains the authority for current company membership and Solidarity. Migration 00045
invalidates stale pre-membership sample labels and refuses a rollback that would invent their lost
run identity. Guild clearing results similarly leave legacy ownership nullable: migration 00046
adds immutable membership-period identity without guessing it from millisecond timestamps.

`Store.ApplyIntent` locks the stream before replay/revision decisions. Applied state, next revision,
events, and receipt commit together. A deterministic terminal rejection stores only its receipt;
revision conflicts and internal failures store nothing. The store exposes time-cutoff pruning for
the deployment scheduler's 30-day idempotency retention policy.

That same transaction inserts the normalized receipt into the transport outbox. Persistence
enforces a 60-KiB receipt ceiling, per-Founder head ordering, leased claims, an attempt counter, and
a durable dead-letter timestamp after the relay's fifth deterministic failure. Dead letters stay
queryable as operational evidence but no longer count as pending readiness work; successful rows
retain their publication timestamp.

An intent id is bound to the canonical hash of its first recorded terminal or applied request for
that retention window. Replaying the identical request returns the original normalized receipt;
correcting a payload under the same id returns `idempotency_conflict`. A corrected payload is a
different logical request and must use a new UUIDv7.

Run the disposable Postgres integration suite locally. The test service owns its network,
database URL, and Go cache, and starts its declared Postgres dependency:

```sh
docker compose -f compose.save-test.yml run --rm test
```

The CI server job supplies the same variable through a Postgres 16 service container. Normal unit
tests skip database integration when the variable is absent; `make test-save-integration` fails
immediately if it is absent.

Archiving is read-only and reversible by future account policy; the persistence API exposes no
hard delete. New-Founder gameplay, save cadence, event consumers, Compute Credit spending, and
leaderboard Balance Epoch policy remain separate RFC responsibilities.
