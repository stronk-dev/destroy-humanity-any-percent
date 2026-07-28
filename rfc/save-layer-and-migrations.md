# RFC: Save Layer & Migrations

- **Status:** draft
- **Author:** Marco (drafted by Claude, revised by Codex)
- **Created:** 2026-07-28
- **Design refs:** `design/06-tech.md §database, §anti-cheat`, `design/02-economy-balancing.md §1`, `design/07-roadmap.md` Phase 0
- **Research:** `design/research/tech-stack.md §1.7`, Profectus and Antimatter Dimensions source reading, `design/research/adaptive-balancing.md` (Balance Epoch)
- **Depends on:** RFC-0001 and RFC-0002 (implemented)
- **Planning:** `planning/save-layer-and-migrations/` (once implementing)

## Summary

RFC-0002 built an authoritative in-memory ledger and deliberately omitted persistence. This RFC
defines its validated restore path, owner-aware Postgres revision streams, and the migration chain
that must exist from the first save version.

It also makes an implicit coupling explicit: a saved Decimal uses the RFC-0001 canonical wire
grammar. A save is a wire payload at rest, not a second numeric format.

## Motivation

The Economy Kernel is complete but cannot survive a process restart. Save versioning cannot be
retrofitted safely after players already possess unversioned state, and poisoned large-number
values must never become the newest durable revision.

This RFC is bounded to persistence mechanics. It does not decide account, gameplay, or leaderboard
policy.

**Out of scope, with owners:** account and Founder lifecycle, including the free `New Founder`
action (a Founder-lifecycle follow-up; this storage shape supports multiple and archived founders);
the append-only gameplay event log (intent/API RFC); leaderboard storage and Balance Epoch
enforcement (leaderboard RFC); offline progression maths (production-engine RFC); client caching.

## Specification

### D1 — Canonical numeric grammar

**NORMATIVE: every Decimal in a save is an RFC-0001 canonical wire string.** No JSON numbers,
alternate encoding, or field-specific precision. Parse, quantize, and reserialize is exact and
idempotent, so an unchanged save round-trip is a byte-for-byte no-op for every numeric value.

Integer counts are exact JSON integers in the inclusive range `0..2^53 − 1`. Decimal fields are
always strings; integer fields are never Decimal strings.

### D2 — Owner-aware revision streams

```sql
save_streams (
  id          uuid primary key,
  owner_kind  text not null, -- 'founder' | 'guild' | 'world'
  owner_id    uuid not null,
  scope       text not null, -- 'company' | 'founder' | 'guild' | 'world'
  archived_at timestamptz null,
  created_at  timestamptz not null default now(),
  unique (owner_kind, owner_id, scope)
);

save_revisions (
  stream_id      uuid references save_streams(id),
  revision       bigint not null,
  version        int not null,
  state          jsonb not null,
  constants_hash text not null,
  created_at     timestamptz not null default now(),
  primary key (stream_id, revision)
);
```

Each lifetime is a separate revision stream. Company and Founder streams are owned by a `founder`
UUID, Guild by a `guild` UUID, and World by a `world` UUID. Any other owner/scope pairing is
rejected. A player's account-to-active-Founder relationship belongs to the future account/Founder
RFC; persistence does not collapse every founder into a player UUID.

Keep the latest five revisions per stream. Insert and pruning happen in the same transaction.
Archiving marks a stream read-only; the persistence API has no hard-delete operation.

Version 1 state has one exact envelope:

```json
{"balances":{"company.cash":"1.23e4"}}
```

`balances` contains exactly the resources declared for the stream's scope in the referenced
catalog: no missing IDs, extra IDs, JSON numbers, or non-canonical strings. Later state fields
require a save-version bump and migration.

### D3 — Non-finite values cannot cross the persistence boundary

A save containing any invalid Decimal is rejected before the database transaction mutates state.
Validation applies to known Decimal fields: `"NaN"` is syntactically valid JSON but invalid saved
state. The last good revision remains current.

The save command carries a mechanical `cause` string and optional intent ID. Rejection logs that
context plus the offending JSON field path. A non-finite value here is an upstream invariant
violation, not an ordinary user error.

### D4 — Versioning and migration chain

- Version starts at `1` and increments on every state-shape or state-semantics change, including an
  optional field with a default.
- Migrations are an ordered append-only chain of pure functions
  `migrate_N_to_N+1(state) -> state`. A released migration is never edited; repairs append a new
  migration.
- Every load enters one migration dispatcher. It applies each step from stored version to current;
  a current-version save enters the same dispatcher and performs zero steps.
- A future version is refused explicitly, never best-effort parsed.
- A checked-in corpus contains at least one save from every historical version and only grows.

### D5 — Restore is a validated constructor

RFC-0002 forbids a public balance setter. Restore loads the catalog identified by
`constants_hash`, migrates state, validates the exact envelope and owner/scope pairing, then
constructs a ledger only if every balance satisfies the catalog's finite-state, minimum, hardcap,
and scope invariants.

One invalid balance rejects the whole save. A resource absent from the referenced catalog is a
migration error, not a silently dropped field. The public ledger mutation API remains unchanged.

### D6 — Balance artifact identity

Every save revision records the balance catalog under which it was written. `constants_hash` is
`sha256:` followed by 64 lowercase hexadecimal digits, computed over the exact catalog artifact
bytes accepted by the loader. Reformatting that artifact intentionally changes its identity.

This RFC stores and propagates the identity but defines no leaderboard policy. Historical
attribution cannot be reconstructed reliably when the exact artifact was not recorded at write
time.

### D7 — Atomic revisions and optimistic concurrency

Every write supplies its expected revision. In one database transaction the repository locks the
stream's latest revision, rejects a mismatch with a typed conflict, inserts exactly the next
revision, and prunes revisions older than the newest five. Validation failure, conflict, insert
failure, or pruning failure rolls back the entire transaction.

This rule also covers Guild streams. There is no guild-specific last-write-wins path.

## Deviations from design

- `design/06-tech.md` sketches `saves(player_id, version, state jsonb)` as one row. This RFC
  replaces it with owner-aware streams and revisions because `design/02 §1` defines different
  owners and lifetimes: World and Guild state are not player-owned, and multiple Founders cannot
  share a player key.
- The gameplay event table is deferred to the intent/API RFC so this bounded persistence kernel
  does not invent event payloads before commands exist.

## Acceptance criteria

1. Ledger to save to restore produces byte-identical canonical balance strings.
2. Invalid/non-finite Decimal state is rejected without changing the current revision; the field
   path and write context are logged.
3. Every historical-version fixture migrates to current in CI; the corpus only grows.
4. Future versions are refused clearly with no partial parse.
5. Missing, extra, cross-scope, or catalog-unknown resources reject the whole restore.
6. A Company write advances only its stream; Founder, World, and Guild revisions are unchanged.
7. Every revision has a syntactically valid, non-empty `constants_hash`.
8. Owner/scope mismatches are rejected, two Founder UUIDs have independent streams, and an
   archived stream is readable but rejects writes.
9. Two concurrent writes with the same expected revision produce one success and one typed
   conflict, never torn state.
10. Integration tests run against Postgres 16 and prove insert-plus-prune retains exactly the five
    newest complete revisions.

## Open questions

- **Migration tooling and integration environment:** choose before acceptance. It must support
  transactional Postgres migrations and clean-checkout tests without introducing an ORM.
- **Save cadence:** write-on-command versus scheduled coalescing belongs to the production-engine
  RFC. The repository exposes atomic writes without selecting the caller's cadence.
- **Founder lifecycle:** its RFC must provide a free, unlimited clean start and archive rather than
  delete old streams. The `design/10 §5` morality-persistence `DESIGN-GAP:` affects what Founder
  state contains, not storage identity.
- **Leaderboard policy:** behavior across `constants_hash` changes remains owned by that RFC.

## Changelog

- 2026-07-28: created (draft).
- 2026-07-28: replaced the player-keyed schema with owner-aware revision streams; specified the v1
  envelope, catalog hash, and concurrency transaction; split Founder lifecycle and events from the
  bounded persistence kernel.
