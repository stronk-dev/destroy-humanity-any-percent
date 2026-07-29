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

## 2026-07-29 (claude — independent review of 37f8ba4..7013b5d: APPROVED)

Full diff read. The fix is the prescribed shape and correct in the direction that matters:

- **Claim-then-validate, one transaction, both functions.** The shared `claimEvent` helper
  performs the dedup insert before *any* state-dependent validation; `!claimed → tx.Commit()`
  makes a retried batch a no-op. Crucially the claim sits **inside the same transaction as the
  validation** — a genuinely invalid first delivery rolls back the claim too, so the event is
  not falsely marked processed. Retry-idempotency and first-delivery strictness both hold, and
  the regression tests assert both directions (replay-after-leave succeeds; first-delivery
  sample-after-leave still errors).
- **The divergence that caused D1 cannot recur by drift:** `project` and `projectSample` now
  share one claim path instead of hand-mirroring each other — the bug was precisely that the
  mirror was imperfect.
- Post-claim resolution failures (assignment/catalog) roll back the claim — transient failures
  remain retryable. The `result` shadowing cleanup in the leave branch is correct.
- Suite green.

**Correction accepted against my own client-shell review:** the `window_ms` finding was FALSE.
`server/production/intents.go:938` requires `window_ms` in the exact-keys envelope and the docs
record it ("audit/UX only; grants no privilege") — I reviewed the archived RFC text instead of
the shipped decoder, the same spec-not-source misattribution the R3 reviewer made. Codex's
source-evidence correction stands; the Transport RFC open item is void. Review rule sharpened:
**verify findings against the implemented surface, not the document that preceded it.**

D1 clear to archive. D2 (Registry event-order convergence) may proceed.
