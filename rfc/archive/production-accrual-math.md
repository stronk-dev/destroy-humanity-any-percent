# RFC: Production Accrual Math

- **Status:** implemented
- **Author:** Marco / Codex
- **Created:** 2026-07-28
- **Design refs:** `design/02-economy-balancing.md §2.4`, `design/06-tech.md §idle-math`
- **Depends on:** RFC-0001 Numeric Core
- **Parent:** Production Engine & Intent API draft
- **Planning:** `planning/archive/production-accrual-math/`

## Summary

Define and implement the shared constant-rate closed-form primitive used by later online and
offline production. This is pure arithmetic: no clock ownership, generator schema, save cursor,
intent, or gameplay policy.

## Motivation

The production draft correctly requires lazy closed-form evaluation and Go/TypeScript parity, but
its surrounding state and command schemas are incomplete. Constant-rate integration is already
fully determined and independently testable, so it should not be blocked or implemented ad hoc
inside the eventual actor.

## Specification

Both runtimes expose:

```text
accrueConstant(rates[], elapsedMilliseconds, efficiency) -> Decimal
```

- Every rate and `efficiency` is a finite authoritative Decimal state value and non-negative.
- `elapsedMilliseconds` is an exact integer in `0..2^53-1`.
- Rates are summed at full Decimal intermediate precision, then multiplied by
  `elapsedMilliseconds / 1000`, then by efficiency.
- The returned delta is quantized once to RFC-0001's 12 significant digits. Zero elapsed time,
  zero efficiency, or an empty/all-zero rate set returns canonical zero.
- Invalid input or a non-finite intermediate returns an error/diagnostic NaN and can never be
  committed.
- Source order cannot change the canonical result. Implementations must use a deterministic
  magnitude-aware summation order: ascending exponent, then canonical string as the tie
  breaker. Callers do not pre-aggregate or quantize individual sources.
- This primitive does not apply hardcaps. The future authoritative engine computes remaining
  headroom and commits through the ledger.

The shared vector file stores canonical rate strings, exact elapsed milliseconds, canonical
efficiency, and canonical expected delta. It includes the adopted offline baseline as input
(`9e-1`, 24 hours) without making policy part of this function.

## Deviations from design

None. This is the constant-rate case already required by the lazy closed-form architecture.

## Acceptance criteria

1. One shared JSON vector corpus passes Go, Node, Chromium, Firefox, and WebKit.
2. Vectors cover zeroes, multiple sources, sub-resolution sources, 90% × 24 hours, huge exponents,
   maximum exact elapsed milliseconds, invalid values, and permutation invariance.
3. Inputs are never mutated and individual sources are never boundary-quantized before summation.
4. Existing numeric/economy/save suites remain green.
5. Canonical numeric documentation describes the primitive and its boundary.

## Open questions

None.

## Changelog

- 2026-07-28: created and accepted as the settled mathematical slice of the reviewed production
  draft under owner direction to continue foundational work.
- 2026-07-28: implemented in Go and TypeScript, verified from one vector corpus in Node and all
  three browser engines, documented, and archived.
