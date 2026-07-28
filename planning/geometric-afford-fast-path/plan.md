# Geometric Affordability Fast Path — Implementation Plan

- **RFC:** `rfc/geometric-afford-fast-path.md`
- **Assignee:** Codex
- **Started:** 2026-07-28

## Work

1. Extract the existing generic bounded search behind an explicit helper.
2. Dispatch validated geometric curves to `decimal.AffordGeometricSeries`.
3. Enforce the economy cap `MaxExactInteger - owned` and verify both affordability postconditions
   through `economy.BulkCost`; fall back to the bounded search if verification fails.
4. Add deterministic boundary/property tests and a same-input benchmark comparison.
5. Run focused Go tests and benchmarks, then the complete `make verify` cross-runtime gate.
6. Update canonical economy documentation and archive the implemented RFC and planning record.

## Acceptance gates

- Existing Go, Node, Chromium, Firefox, and WebKit suites pass.
- Geometric boundary tests cover zero cash, ratio one, huge values, and an owned count adjacent to
  `MaxExactInteger`.
- Deterministic generated cases satisfy affordability, next-count, and exact-integer-cap
  postconditions.
- Constant and linear max-affordable behavior remains covered.
- The public geometric query/helper benchmark ratio is below 10× on the same run.
- `docs/economy-kernel.md` is canonical for the shipped behavior.
