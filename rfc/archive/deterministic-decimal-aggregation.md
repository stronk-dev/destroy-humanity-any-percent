# RFC: Deterministic Decimal Aggregation

- **Status:** implemented
- **Author:** Codex
- **Created:** 2026-07-28
- **Depends on / amends:** RFC-0001 Numeric Core, RFC-0002 Economy Kernel, Production Accrual Math
  (implemented)
- **Planning:** `planning/archive/deterministic-decimal-aggregation/`

## Summary

Define one order-independent n-ary Decimal sum for authoritative aggregation. Use it for ledger
transactions and production source totals, and remove locale-sensitive client ordering.

## Specification

`SumDeterministic(values)` validates finite state values, groups values by exponent, orders each
group by absolute mantissa then signed mantissa, sums each complete group before normalization,
and combines groups from lowest to highest exponent. The input is never mutated. An invalid input
or a mathematically out-of-domain result returns diagnostic NaN.

Grouping same-exponent terms before Decimal normalization prevents a valid cancellation such as
`[+B,+B,-B,-B]` from overflowing only because one request order encountered the positive prefix
first. The total ordering also distinguishes full-precision values that share a 12-digit canonical
string.

The Go ledger aggregates every resource through this helper before range validation. Go and
TypeScript production accrual use the equivalent total order. TypeScript compares numeric fields
directly and never uses locale-sensitive string collation.

## Deviations from design

- Production Accrual Math named canonical string as its final tie-break. That does not totally
  order distinct full-precision values sharing one 12-digit rendering. Raw mantissa is the final
  tie-break instead; canonical authoritative string inputs behave identically.

## Acceptance criteria

1. Every permutation of the demonstrated ledger multiset has the same acceptance and exact result.
2. Out-of-domain net results still reject atomically; inputs remain unmodified.
3. Production vectors pass identically in Go, Node, and all three browsers without locale collation.
4. Canonical numeric/economy docs describe deterministic n-ary aggregation.

## Open questions

None.

## Changelog

- 2026-07-28: created, accepted, and implementation started from verified review findings A5/F4.
- 2026-07-28: n-ary aggregation, ledger/accrual integration, permutation regressions, canonical
  docs, and all-runtime verification completed and archived.
