# Save Archive Compare-and-Swap — Running Log

## 2026-07-28

- Adversarial review demonstrated that the prior scalar-subquery UPDATE can archive a stale head
  after waiting on a concurrent writer under PostgreSQL READ COMMITTED.
- Follow-up accepted and implementation started. No push.
