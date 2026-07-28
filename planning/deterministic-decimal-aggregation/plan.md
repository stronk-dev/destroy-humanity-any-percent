# Deterministic Decimal Aggregation — Plan

- **Status:** implementing
- **RFC:** `rfc/deterministic-decimal-aggregation.md`

1. Implement deterministic n-ary sum in the Go numeric core.
2. Route ledger and production aggregation through it.
3. Remove locale-sensitive TypeScript production ordering.
4. Add permutation/overflow regressions, verify, document, and archive.
