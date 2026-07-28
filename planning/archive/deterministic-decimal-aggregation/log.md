# Deterministic Decimal Aggregation — Running Log

## 2026-07-28

- Review demonstrated ledger acceptance differed for permutations of the same legal multiset.
- Review also found locale-sensitive TypeScript ordering and a non-total canonical-string tie-break.
- Follow-up accepted and implementation started. No push.
- Added a shared Go n-ary sum that groups same-exponent terms before normalization and uses a raw
  numeric total order without mutating inputs.
- Ledger transactions and Go production accrual now use it. TypeScript production uses the
  equivalent helper and no longer calls locale-sensitive collation.
- Domain-edge cancellation permutations pass; genuine out-of-domain sums still reject.
- Focused Go and TypeScript suites passed. No push.
- Final `make verify` passed with Go vet/tests, strict TypeScript, schema validation, 6,347 Node
  tests, and 19,041 browser tests across Chromium, Firefox, and WebKit.
- Canonical numeric/economy docs updated; follow-up archived. No push.
