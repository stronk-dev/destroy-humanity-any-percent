# RFC: Geometric Affordability Fast Path

- **Status:** implemented
- **Author:** Marco / Codex
- **Created:** 2026-07-28
- **Design refs:** `design/06-tech.md §idle-math`, `design/07-roadmap.md` Phase 0
- **Research:** `design/research/cicd-deploy.md §3.1c, §9.3`
- **Depends on:** RFC-0001 and RFC-0002 (implemented)
- **Parent / amends:** RFC-0002 Economy Kernel
- **Supersedes / superseded by:** —
- **Planning:** `planning/geometric-afford-fast-path/` (once implementing)

## Summary

Route geometric economy quotes through RFC-0001's closed-form affordability inverse while
preserving every RFC-0002 semantic boundary. Constant and linear curves keep the generic bounded
search. This removes a measured harness blocker without coupling the repair to CI scaffolding.

## Motivation

`economy.MaxAffordable` currently binary-searches every curve kind across as many as
`2^53 - 1 - owned` purchases. `decimal.AffordGeometricSeries` already implements an estimated
closed-form inverse, local correction, postcondition verification, and bounded-search fallback,
but has no production caller.

The research measured roughly 20,486 ns/op for the generic economy path and 660 ns/op for the
standalone Decimal helper (about 31×). Its wider 200-bot harness measurements were 3 min 01 s and
1.91 s (about 95×); these are different measurement scopes, not interchangeable speedup claims.

## Specification

For a validated `geometric` `PriceDefinition`, `economy.MaxAffordable`:

1. calls `decimal.AffordGeometricSeries(cash, base, ratio, owned)`;
2. caps its result at `decimal.MaxExactInteger - owned`;
3. verifies the capped result through `economy.BulkCost` so economy-level validity and
   quantization semantics remain authoritative;
4. corrects downward or upward when needed until both postconditions hold inside that cap:
   `cost(count) <= cash`, and either `count == cap` or `cost(count+1) > cash`;
5. uses the existing bounded search inside `0..cap` if local verification cannot establish both
   postconditions.

Constant and linear curves continue through the generic bounded search unchanged. Invalid inputs
return the same error category and zero count as before. This is a performance repair, not a new
curve or balance mechanic.

A benchmark compares the public geometric economy query with the Decimal helper using the same
inputs. The public path must remain below ten times the helper's runtime. The benchmark reports a
ratio; it is a diagnostic guard and does not use a wall-clock pass/fail threshold.

## Deviations from design

None. This applies the already selected closed-form/lazy-math architecture at a missed call site.

## Acceptance criteria

1. Existing Go, TypeScript, and browser suites remain green.
2. Geometric queries cover zero cash, ratio one, huge exponents, and `owned` adjacent to
   `MaxExactInteger`.
3. Property tests assert both affordability postconditions and `owned + count <= MaxExactInteger`
   across deterministic generated cases.
4. Constant and linear regression tests demonstrate equivalent generic-search semantics.
5. The public-path/helper benchmark ratio is below 10× on the same benchmark run.
6. Canonical economy documentation describes the geometric fast path and exact cap.

## Open questions

None.

## Changelog

- 2026-07-28: created and accepted by owner direction to proceed after selecting a public
  repository; split from the draft CI baseline because it amends implemented economy behavior.
- 2026-07-28: implemented, verified in both runtime suites and all browser engines, documented,
  and archived.
