# Save Archive Compare-and-Swap — Running Log

## 2026-07-28

- Adversarial review demonstrated that the prior scalar-subquery UPDATE can archive a stale head
  after waiting on a concurrent writer under PostgreSQL READ COMMITTED.
- Follow-up accepted and implementation started. No push.
- Replaced Archive with the same `SELECT ... FOR UPDATE` locked-head transaction shape as Write.
- Added a real-Postgres regression that waits until Archive is blocked on the writer's row lock,
  advances the head, then proves Archive returns `ErrConflict` and leaves the stream active.
- Integration suite passed. No push.
