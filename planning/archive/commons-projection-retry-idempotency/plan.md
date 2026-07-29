# Commons Projection Retry Idempotency — implementation plan

- **Assignee:** Codex
- **RFC:** `rfc/archive/commons-projection-retry-idempotency.md`
- **Started:** 2026-07-29

## Work breakdown

1. [x] Move event-ID dedup ahead of mutable projection dependencies in both Commons paths.
2. [x] Add replay-after-leave and failed-first-delivery Postgres regressions.
3. [x] Update canonical Commons docs and run focused/full verification.
4. [x] Record independent review and archive.
