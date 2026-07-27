# Numeric Normalization Carry

- **Status:** implementing
- **Author:** Codex
- **Created:** 2026-07-28
- **Design refs:** `design/research/numeric-core.md`
- **Depends on:** Numeric Core Boundary Hardening (implemented)
- **Parent / amends:** `rfc/archive/0001-numeric-core.md`
- **Supersedes / superseded by:** —
- **Planning:** `planning/numeric-normalization-carry/`

## Summary

Repair the Go Decimal normalizer when floating-point logarithm/power rounding leaves an exact
mantissa of `10` after its first scaling pass. The value is mathematically finite but violates the
numeric core's normalized-representation and state invariants.

## Motivation

RFC-0002's million-source aggregation test discovered that adding ten equal `1e87` values can
produce an internal Go value `{mantissa: 10, exponent: 87}`. Its canonical string is `1e88`, but
`IsStateValue` correctly rejects the malformed internal representation. Valid arithmetic must not
produce invalid representation.

This is a narrow conformance repair to implemented behavior, not a new numeric feature.

## Specification

After logarithmic scaling, `Normalize` must perform a final carry/borrow correction so every
finite non-zero result has `1 <= abs(mantissa) < 10`. Adjusting across either exponent boundary
must preserve the existing non-finite diagnostic behavior. Canonical strings, 12-digit
quantization, arithmetic domains, and wire grammar do not change.

## Deviations from design

None. This restores the documented invariant.

## Acceptance criteria

1. A regression test adds ten `1e87` values and obtains normalized, state-valid `1e88`.
2. Positive and negative direct normalization around a scaling carry remain normalized.
3. The existing deterministic vector file is unchanged and all Go, TypeScript, and browser
   suites pass.
4. `docs/numeric-core.md` records the repaired normalization guarantee; RFC and planning are
   archived when implemented.

## Open questions

None.

## Changelog

- 2026-07-28: created and implementation started after RFC-0002 exposed the invariant violation.
