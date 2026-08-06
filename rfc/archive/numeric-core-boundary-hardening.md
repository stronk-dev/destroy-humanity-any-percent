# RFC: Numeric Core Boundary Hardening

- **Status:** implemented
- **Author:** Codex, from owner-directed foundation audit
- **Created:** 2026-07-27
- **Design refs:** `design/06-tech.md §3` (big numbers)
- **Depends on:** [Archived RFC-0001: Numeric Core](0001-numeric-core.md)
- **Parent / amends:** [Archived RFC-0001: Numeric Core](0001-numeric-core.md)
- **Supersedes / superseded by:** —
- **Planning:** `planning/archive/numeric-core-boundary-hardening/`

## Summary

Harden the implemented numeric core at its public boundaries before save, resource, and
production systems depend on it. Valid normalized gameplay values already satisfy RFC-0001 and
pass the cross-runtime corpus; this follow-up closes malformed-input, diagnostic-sentinel, and
domain-edge gaps discovered by a post-implementation audit.

## Motivation

RFC-0001's 6,278-vector corpus and Node/Chromium/Firefox/WebKit suites prove the intended
layer-0 arithmetic path. They do not yet prove every boundary that later infrastructure will
treat as a safety gate.

Concrete current behavior:

- TypeScript `quantize("12.345e2")` produces `1e3`, silently changing the magnitude instead of
  normalizing first and producing `1.2345e3`.
- A `Decimal.fromMantissaExponent_noNormalize(100, 0)` value passes TypeScript
  `isStateValue`; canonical serialization changes it to `1e1` rather than rejecting or
  correctly normalizing it to `1e2`.
- `break_infinity.js` encodes positive/negative infinity using exponent `9e15`; the project
  `classify` helper currently reports those sentinels as finite (while correctly rejecting them
  from state).
- The corpus includes NaN domain cases but no infinity-result cases and no explicit table for
  zero powers, division by zero, or logarithm domain edges. The Go and client implementations
  can therefore disagree on diagnostic results without failing the suite.

These are not range failures and do not invalidate existing canonical gameplay values. They
are boundary-contract holes that should be closed before those helpers become persistence and
transaction gates.

## Specification

### 1. Normalized-state invariant

`isStateValue` in both runtimes must accept a value only when all conditions hold:

- mantissa and exponent are finite;
- exponent is an exact integer inside RFC-0001's gameplay domain;
- zero is represented only as mantissa `0`, exponent `0`;
- non-zero has `1 <= abs(mantissa) < 10`.

The client must not trust `break_infinity.js` objects merely because their fields are finite;
the library exposes mutable fields and a deliberately unsafe no-normalize constructor.

### 2. Normalize before quantizing

TypeScript `quantize` and `canonicalString` must clone and normalize every `DecimalSource`
without mutating the caller's object. Rounding always operates on the normalized coefficient,
never on an arbitrary textual coefficient or unsafe internal representation.

Scientific input accepted by the permissive constructor may use any finite coefficient. Thus
`12.345e2`, `0.12345e4`, and `1.2345e3` must quantize to the same value and canonical string.
Strict `parseCanonical` continues to accept only RFC-0001's canonical grammar.

### 3. Diagnostic classification

The shared test classification must recognize the pinned client's exponent-`9e15` infinity
sentinel by mantissa sign. The excluded `-9e15` underflow boundary is invalid gameplay state but
is not mislabeled as infinity. NaN in either component remains `nan`.

Diagnostic classification never makes a non-finite value persistable: NaN, both infinities,
sentinels, and out-of-domain exponents remain invalid state and invalid canonical wire values.

### 4. Explicit domain-edge contract

The pinned `break_infinity.js` client remains the arithmetic oracle, including unconventional
diagnostic edge behavior. Add explicit shared vectors for at least:

- `0^0`, `0^positive`, and `0^negative`;
- negative base with integer and fractional exponents;
- finite division by zero and zero divided by zero;
- `log10`/`ln` of zero and negative values;
- `exp` results that remain finite, underflow, or reach the infinity sentinel;
- positive and negative infinity inputs where the library accepts them diagnostically;
- quantization carry at both gameplay exponent boundaries.

The Go port must match the expected value or exact diagnostic classification for every vector.
NaN and infinity diagnostic results may not pass `IsStateValue`, `isStateValue`, or canonical
parsing. The pinned library defines division by zero and `0` raised to a negative power as
finite zero; the Go port mirrors that result for compatibility. Feature-level handlers must
validate operation domains before calculation and must never treat those library edge results
as authorization for gameplay state transitions.

### 5. Corpus and property strengthening

The deterministic generator must report operation/classification counts and must commit at
least one vector for every domain-edge class above. Tests must fail if an expected class has
zero coverage. Add properties for normalized-state validation, equivalent scientific-source
normalization, no input-object mutation, and canonical idempotence.

The complete Node, Chromium, Firefox, WebKit, Go, vet, deterministic-generation, and fuzz gates
remain mandatory.

## Deviations from design

None. This follow-up enforces the representation and invalid-state rules already selected by
RFC-0001 and `design/06-tech.md`.

## Acceptance criteria

1. Equivalent finite scientific inputs with normalized and non-normalized coefficients
   quantize and serialize identically in both runtimes.
2. Unsafe/mutated client Decimal representations and malformed Go internal values fail state
   validation unless explicitly normalized first.
3. Infinity sentinels classify correctly and are rejected from state and canonical wire forms.
4. The shared corpus contains every specified domain edge; classification coverage cannot
   silently fall to zero.
5. Go and TypeScript agree on exact zero/domain classification and satisfy the existing
   tolerance for finite non-zero results.
6. Quantization does not mutate caller-owned Decimal objects.
7. `make verify`, deterministic regeneration, and the documented fuzz pass are green.
8. `docs/numeric-core.md` is updated, this RFC and its planning log are archived, and the
   shipped behavior is understandable without reading either RFC.

## Open questions

None blocking. The pinned client defines diagnostic arithmetic; server authority and state
validation prevent those diagnostics from becoming gameplay state.

## Changelog

- 2026-07-27: drafted from the post-RFC-0001 foundation audit at the owner's request.
- 2026-07-27: accepted by owner direction (perfect numeric foundation before downstream
  systems) and moved to implementing.
- 2026-07-27: clarified the executable domain-edge contract after vectors exposed the pinned
  library's finite-zero results for division by zero and `0` to a negative power. These remain
  compatibility behavior, not valid unguarded gameplay domains.
- 2026-07-27: implemented. Shipped normalized state guards, normalize-before-quantize behavior,
  infinity-sentinel classification, schema-3 coverage metadata, 20 mandatory edge vectors, and
  cross-runtime diagnostic parity; updated canonical docs and archived the work.
- 2026-08-06: non-normative reference cleanup for publication; no spec change.
