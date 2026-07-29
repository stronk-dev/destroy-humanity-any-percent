# Commons Projection Retry Idempotency — append-only log

## 2026-07-29 — start

- Independent archived-four review demonstrated D1 against real Postgres: a committed sample is
  checked against current membership before event-ID dedup, so replay after leave fails forever.
- Owner supplied the remediation order. Scope is the projection transaction boundary only; Health,
  Capacity, membership semantics, and event payloads do not change.
