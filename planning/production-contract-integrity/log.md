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
