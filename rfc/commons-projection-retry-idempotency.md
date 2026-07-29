# RFC: Commons Projection Retry Idempotency

- **Status:** implementing
- **Author:** Codex, from independently demonstrated review finding D1
- **Created:** 2026-07-29
- **Design refs:** `design/05-mmo.md §5`; `design/06-tech.md §persistence`
- **Research / reproducer:** `planning/archived-four-review/log.md` D1
- **Amends:** `archive/commons-compact.md` C7
- **Planning:** `planning/commons-projection-retry-idempotency/`

## Summary

Make every already-committed Commons projection event a successful no-op before the projector
consults mutable membership, assignment, or catalog state. The sample path currently validates
active membership before event-ID deduplication, so replaying a committed sample after its company
leaves wedges at-least-once delivery and full-history rebuilds forever.

## Specification

After strict payload and numeric validation, both membership and sample projection paths begin a
transaction and insert the immutable event ID into `commons_projection_events` with
`ON CONFLICT DO NOTHING`. If the row already exists, the transaction commits immediately. It does
not resolve current assignment/catalog state or query current membership.

For a first delivery, the insert and all derived writes remain in the same transaction. Any later
validation or write failure rolls the dedup row back with the rest of the projection. Moving dedup
earlier therefore cannot mark a failed first delivery as complete.

## Acceptance criteria

1. Real Postgres projects `compact_signed → compact_sampled → compact_left`, then replays that full
   history after membership is inactive; replay succeeds.
2. Replay changes neither cumulative Capacity, membership, nor projected-event cardinality.
3. A first-delivery sample without active membership still fails and leaves no dedup row.
4. Existing assignment concurrency, sample replay, leave replay, health, and harness tests pass.

## Deviations from design

None. This restores the at-least-once guarantee already required by Commons C7.

## Open questions

None.

## Changelog

- 2026-07-29: drafted, accepted, and implementation started by owner-directed remediation order
  after D1 was independently reproduced against Postgres.
