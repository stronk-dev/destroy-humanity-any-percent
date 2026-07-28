# RFC: Save Archive Compare-and-Swap

- **Status:** implementing
- **Author:** Codex
- **Created:** 2026-07-28
- **Design refs:** `design/06-tech.md §persistence`
- **Depends on / amends:** Save Layer & Migrations (implemented)
- **Planning:** `planning/save-archive-cas/`

## Summary

Make archive and write use the same locked-head compare-and-swap transaction so Archive cannot
successfully act on a stale revision while a concurrent Write advances the stream.

## Specification

`Archive(streamID, expectedRevision)` begins a transaction, locks the stream row with
`SELECT ... FOR UPDATE`, rejects missing/already-archived streams, reads the latest revision while
holding that lock, compares it with `expectedRevision`, updates `archived_at`, and commits. It must
not encode the head check as a scalar subquery inside one `UPDATE`: under PostgreSQL READ COMMITTED,
that subquery can retain a stale snapshot after waiting for a concurrent writer.

Archive/Write races have one serial outcome. If Write obtains the lock first and advances the head,
Archive returns `ErrConflict`. If Archive obtains it first, Write later returns `ErrArchived`.

## Deviations from design

None. This repairs the optimistic-concurrency guarantee already documented by the save layer.

## Acceptance criteria

1. A deterministic real-Postgres regression forces Write to hold the row lock while Archive begins;
   after Write commits, Archive returns `ErrConflict` and the stream remains active.
2. Existing write/write, retention, archive, and load integration tests remain green.
3. Canonical save documentation states that archive uses the locked-head transaction.

## Open questions

None.

## Changelog

- 2026-07-28: created, accepted, and implementation started from a demonstrated concurrency
  reproducer found during adversarial review.
