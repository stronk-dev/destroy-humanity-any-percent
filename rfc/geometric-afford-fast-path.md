# RFC: Geometric Afford Fast Path

- **Status:** accepted
- **Author:** Marco (drafted by Claude; scope per Codex's CI Baseline D4)
- **Created:** 2026-07-28
- **Design refs:** —
- **Research:** `design/research/cicd-deploy.md §9.3` (the measurement)
- **Depends on:** RFC-0001, RFC-0002 (implemented)
- **Planning:** `planning/geometric-afford-fast-path/` (once implementing)

## Summary

`server/economy/curves.go` `MaxAffordable` answers every curve kind by binary search over `[0, MaxExactInteger − owned]`. For `geometric` curves, RFC-0001 already shipped and tested a closed form — `decimal.AffordGeometricSeries` — which currently has **no non-test caller**. Measured: **20,486 ns/op vs 660 ns/op (~95×)**; on a 200-bot harness run, **3 min 01 s vs 1.91 s** — the difference between a blocking CI gate that is kept and one that is bypassed. This RFC routes the geometric case to the closed form.

## Specification

1. `MaxAffordable` dispatches on curve kind: `geometric` → `decimal.AffordGeometricSeries` (candidate) → **verify-and-correct** per RFC-0001 contract §7 (evaluate exact series cost; correct down/up until `sum(n) ≤ cash && sum(n+1) > cash`; bounded correction, else fallback + invariant report). `constant`/`linear` → the existing binary search, unchanged.
2. The result is capped at `MaxExactInteger − owned` exactly as today; all existing postconditions hold verbatim. **This is a performance change with zero semantic surface.**
3. A benchmark guard in CI asserts the geometric path stays within one order of magnitude of raw `AffordGeometricSeries`.

## Acceptance criteria

1. Every existing max-affordable golden vector and postcondition test passes unchanged.
2. A differential test runs both paths across the full vector corpus (r ∈ {1.07, 1.10, 1.13, 1.15} and edge exponents) and asserts **identical integer results**.
3. Benchmark: geometric `MaxAffordable` ≤ 10× `AffordGeometricSeries`; the 200-bot loop scenario from `cicd-deploy.md §9.3` completes in < 10 s.
4. `AffordGeometricSeries` has a non-test caller.

## Open questions

None. Deliberately the smallest possible RFC; blocks the balance-harness RFC.

## Changelog

- 2026-07-28: created; accepted (owner direction to keep the implementation queue full).
