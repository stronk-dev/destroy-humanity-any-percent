# Commons Projection Retry Idempotency — append-only log

## 2026-07-29 — start

- Independent archived-four review demonstrated D1 against real Postgres: a committed sample is
  checked against current membership before event-ID dedup, so replay after leave fails forever.
- Owner supplied the remediation order. Scope is the projection transaction boundary only; Health,
  Capacity, membership semantics, and event payloads do not change.

## 2026-07-29 — implementation

- Added one `claimEvent` transaction helper and moved it ahead of mutable assignment, catalog, and
  membership reads in both membership and sample paths. Strict payload/numeric validation remains
  before the claim. First-delivery failures roll back the claim with the transaction.
- Extended the real-Postgres integration test with the demonstrated signed→sampled→left replay
  after leave. It asserts unchanged cumulative Capacity and exactly five projected event IDs. A new
  post-leave sample still rejects and leaves zero dedup rows.
- Focused Commons package and complete `make test-save-integration` matrix pass.
- Canonical `docs/commons.md` now states claim-before-mutable-state and transactional rollback.
  Full `TEST_DATABASE_URL=… make verify` passes, including the harness baseline, 6,412 Node tests,
  and 19,245 Chromium/Firefox/WebKit tests. Ready for the mandatory independent diff review.
