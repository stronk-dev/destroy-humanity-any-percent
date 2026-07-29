# Production Contract Assertions & Integrity — running log

Append-only. A fresh agent resumes from this file and `plan.md`.

## 2026-07-29 (Codex — start)

- Owner explicitly directed implementation of the RFC queue; treated that direction as acceptance
  of `rfc/production-contract-integrity.md`.
- Marked the RFC `implementing` and opened this planning record.
- Work order is D1/D2, D3/D4, D5, full verification/archive, then the accepted Balance Harness
  Foundation. No push is authorized.

## 2026-07-29 (Codex — D1/D2 implementation)

- Added the neutral raw-byte source-order helper and made `production.Rates` use it. Direct package
  tests lock the closed slot order, uniqueness, unknown-slot rejection, raw-byte punctuation order,
  and input immutability.
- Added shared multiplier catalog cases consumed by Go and TypeScript: valid multi-provider slots,
  duplicate sources, single-provider commons/trust, unknown slot/target, and malformed provider.
  Go runtime tests additionally cover declaration mismatches, duplicates, non-positive factors,
  and all six permutations of a three-source slot.
- Exported `InvariantKind`, `InvariantReport`, and `InvariantSink`; replaced the private report slice
  contract with a transaction-local collector. Applied reports become validated events, committed
  rejection reports are audit/metrics-only, abort reporting admits only `residual_abort`, and
  replays do not double-count.
- Added unit tests for the exported sink, event payloads, report/request identity, audit/metric
  routing, and replay suppression. Extended the Postgres integration test with atomic applied-event,
  terminal-rejection, rollback, and replay cases.
- Focused verification green: `go test ./multiplier ./economy ./production`; client Vitest 6,372
  tests green.

## 2026-07-29 (Codex — D3/D4 implementation)

- Upgraded the published formula artifact to schema v2. The generator hashes normalized,
  comment-free AST for `production.Rates`, `multiplier.Order`, and the ordering helper actually
  called by Rates. The published order token now comes from the multiplier package.
- Generator tests prove comment/format changes preserve the fingerprint, executable rate changes
  alter it, and a renamed/missing authority fails closed. Canonical docs now say source-bound and
  review-gated rather than claiming algebra is automatically inferred.
- Serialized both the explicit integration target and `test-go` package execution with `-p 1`;
  CI supplies `TEST_DATABASE_URL` to `test-go`, so leaving that target parallel would retain the
  same first-run migration race under another name.
- Ran the explicit integration target ten consecutive times against Postgres 16: all ten passed,
  including the new invariant transaction cases.

## 2026-07-29 (Codex — D5 implementation)

- Renamed save-migration fixture metadata to `corpus_version` and added a checked-in baseline
  manifest. Server verification now rejects a smaller corpus, a missing required case, duplicate
  names, or an unexpected corpus/baseline schema.
- Made receipt change construction fallible. It parses every before/after canonical value, rejects
  changed resource sets, and propagates `ErrInvalidEngineState` rather than omitting bad data.
  Regression tests cover malformed equal and unequal values plus a missing resource.
- Added the service-level Postgres regression for a corrected payload reusing a terminal intent id:
  it returns `idempotency_conflict`, does not mutate state, and requires a new UUIDv7.
- Canonical save/production docs now distinguish immutable event origin identity from retained
  snapshot availability, publish sticky idempotency, describe every invariant outcome path, and
  keep `internal_invariant` transport mapping explicitly deferred.
- Focused save/production suites and the updated serialized Postgres integration target are green.
