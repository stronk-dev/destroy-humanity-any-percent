# RFC: Resource-Log Domain Parity

- **Status:** accepted
- **Author:** Marco (drafted by Codex from the round-2 review)
- **Created:** 2026-07-28
- **Design refs:** `design/07-roadmap.md` Phase 0; `AGENTS.md` law 3
- **Research:** `planning/production-review-round2/log.md` R3, corrected by direct runtime probe
- **Amends:** `archive/0002-economy-constants-and-ceilings.md` and
  `archive/production-engine-and-intents.md`
- **Planning:** `planning/resource-log-domain-parity/` (once implementing)

## Summary

Reject `resource_log` targets too small for `log10(1 + target)` to remain positive in the shared
numeric domain, and make the TypeScript evaluator use Decimal division rather than native
JavaScript number division. This fixes the demonstrated Go/client progress divergence without
changing the already-correct `break_infinity.js` division-by-zero contract.

## Motivation

Both catalog loaders currently accept every canonical positive target. In the numeric core's
aligned-add algorithm, a target below `5e-15` can collapse `1 + target` to exactly 1, making the
resource-log denominator zero. Go then follows `break_infinity.js` Decimal semantics and returns
zero from Decimal division by zero; the client progress function instead divides two native
JavaScript numbers and obtains `Infinity`, which it rejects.

The review classified this as a primitive Go/JS divergence. Direct verification corrected that
diagnosis: installed `break_infinity.js` 2.2.0 returns zero for `1 / 0` and `0 / 0`, exactly like
Go, and the shared golden corpus already enforces `div-zero` and `zero-div-zero`. Changing Go's
primitive to Infinity would itself violate RFC-0001.

## Specification

### D1 — The Decimal primitive does not change

Go `Decimal.Div` and the TypeScript `Decimal.div` oracle retain their existing behavior for a zero
divisor. The mandatory golden-vector edges `div-zero` and `zero-div-zero` remain corpus-growth
guards. Native JavaScript `/` is not a numeric-core oracle and may not be used for progress
division.

### D2 — Minimum resource-log target

Every `resource_log` target—top-level coordinate and composite term—must be a canonical Decimal
greater than or equal to `5e-15`. This is the exact half-unit threshold induced by the shared
aligned-add calculation (`1e14 * target` rounded to the nearest integer); at `5e-15`,
`log10(1 + target)` is positive, while `4e-15` collapses to zero.

The Go loader, TypeScript loader, and schema-verification command all enforce the semantic bound.
Because JSON Schema cannot numerically compare a canonical scientific-notation string, the Ajv
shape check is followed by the same explicit semantic target check; a regex is not used as a fake
numeric comparison.

After parsing, both runtimes also verify defensively that `log10(1 + target)` is finite and
strictly positive. Failure rejects the catalog—it never becomes a runtime progress value.

### D3 — Matched evaluation shape

TypeScript evaluates `resource_log` using Decimal objects for both logarithms and calls
`numerator.div(denominator)` before the shared progress clamp. It must not use native number `/`.
Go retains its current `Decimal.Log10().Div(...)` shape. Both sides quantize only at the existing
progress-result boundary.

Catalog acceptance and progress results are covered by one shared fixture family. No runtime is
allowed to accept a catalog that its peer rejects.

### D4 — Existing catalogs

All shipped Phase-0 targets (`1e3` through `1e12`) are unchanged. This is a domain guard and
operator-parity correction, not a balance retune; it does not mint a Balance Epoch or require a
`BALANCE-CHANGE:` baseline update.

## Deviations from design

None. The correction preserves RFC-0001's selected library semantics and Production's intended
cross-runtime progress parity.

## Acceptance criteria

1. Shared loader fixtures reject `1e-16`, `1e-15`, and `4e-15`; accept `5e-15`, `9e-15`, and the
   shipped targets; Go, Node, Chromium, Firefox, and WebKit agree.
2. Shared progress fixtures at the `5e-15` boundary and across representative values produce the
   same canonical result in Go and TypeScript.
3. A source/lint assertion prevents native `/` in the TypeScript `resource_log` evaluator, and a
   targeted unit test would fail if it returned to number division.
4. The existing `div-zero` and `zero-div-zero` golden vectors remain unchanged and green in both
   runtimes, proving the primitive contract was not rewritten to fix a caller bug.
5. `make verify-schema` performs both JSON shape validation and the semantic target-floor check;
   an invalid checked-in catalog fails CI.
6. `docs/economy-kernel.md` and `docs/production-engine.md` publish the resource-log target domain.

## Open questions

None. The numeric threshold and the owning layer are explicit.

## Changelog

- 2026-07-28: drafted from round-2 finding R3, correcting its primitive-level diagnosis after
  direct Go and installed-library verification.
