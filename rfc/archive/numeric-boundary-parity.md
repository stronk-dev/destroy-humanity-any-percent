# RFC: Numeric Boundary Parity

- **Status:** implemented
- **Author:** Codex
- **Created:** 2026-07-28
- **Research:** adversarial cross-runtime review recorded in
  `planning/archive/generator-production-state/log.md`
- **Depends on / amends:** RFC-0001 Numeric Core (implemented), Geometric Affordability Fast Path
  (implemented)
- **Planning:** `planning/archive/numeric-boundary-parity/`

## Summary

Repair demonstrated Go/JavaScript disagreement near the Decimal exponent boundary and close the
golden corpus blind spot that allowed it. Also make the exported geometric-series helpers treat a
ratio whose representable difference from one is zero as the constant-price case.

## Specification

- Division computes normalized mantissa division and exponent subtraction directly. It must not
  create a reciprocal whose exponent is invalid when the final quotient is valid. Existing
  diagnostic zero/infinity compatibility behavior remains unchanged.
- Multiplication allows the one-exponent lower-bound slack in which multiplying mantissas can carry
  the result back into the valid domain; normalization decides final validity.
- Geometric-series sum/afford treats a valid ratio as constant when the Decimal denominator
  `1 - ratio` is representationally zero. This prevents a zero cost and maximum-affordable result
  for ratios immediately above one.
- Golden random binary inputs span the complete legal exponent domain. Named division and
  multiplication boundary regressions are mandatory corpus edges.
- The vector generator remains a pinned-JavaScript compatibility oracle, not an independent
  mathematical oracle. Canonical/quantization helpers duplicated there must be audited whenever
  their client counterparts change.

## Deviations from design

None. These changes restore the cross-runtime and valid-state contracts.

## Acceptance criteria

1. The demonstrated valid division and multiplication pairs produce matching finite results in Go,
   Node, Chromium, Firefox, and WebKit.
2. Generated vectors contain operands above absolute exponent `4e15` and enforce both named edges.
3. Near-one geometric ratios never produce zero cost for a positive base/count and affordability
   satisfies both postconditions in both runtimes.
4. Existing 12-digit canonical, diagnostic-domain, economy, and save behavior remains green.

## Open questions

None.

## Changelog

- 2026-07-28: created, accepted, and implementation started from demonstrated adversarial-review
  reproducers.
- 2026-07-28: arithmetic, vectors, geometric limit, documentation, and all-runtime verification
  completed and archived.
