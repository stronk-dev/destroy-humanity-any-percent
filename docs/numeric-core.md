# Numeric Core

RFC-0001 implements the shared numeric foundation used by future economy systems. The client
uses pinned `break_infinity.js` 2.2.0; the server owns an operation-compatible Go `Decimal`.
Both are tested against the same deterministic golden-vector corpus.

## What the number represents

A non-zero value is a signed `float64` mantissa with absolute value in `[1, 10)` and an
integer base-10 exponent. For example, `6.25e123456` is stored as mantissa `6.25` and exponent
`123456`. Gameplay state accepts exponents from `-8,999,999,999,999,999` through
`8,999,999,999,999,999`.

This is enormous-range floating-point arithmetic, not arbitrary precision. Intermediate
calculations retain roughly float64 precision. Every value committed to authoritative state is
then deliberately rounded to 12 significant decimal digits. That stable boundary matters more
than retaining meaningless low digits at gigantic magnitudes.

`Decimal` is only for continuous magnitudes such as currency, prices, rates, and multipliers.
Counts, milestones, sequence numbers, and time units remain exact integers capped at
`9,007,199,254,740,991` when they cross the JavaScript boundary.

## Authoritative state boundary

Before a calculated value becomes player state, the server must:

1. Quantize it to 12 significant digits using round-half-to-even.
2. Reject NaN, either infinity, and values outside the exponent domain.
3. Serialize it to the canonical string grammar below.
4. Commit the state only if every validation succeeds.

The client may predict the result, but it never authorizes a purchase or threshold crossing.
It reconciles to the canonical value returned by the server.

Canonical zero is `0`. A non-zero value has the form `[-]d[.digits]e[-]digits`, contains at
most 12 significant digits, has no trailing coefficient zeros, and has no `+` or leading zeros
in the exponent. Examples:

```text
1e0
-4.25e-7
9.87654321012e123456
```

Canonical parsing is strict and idempotent. The more permissive constructors are for
configuration/import diagnostics and do not make a value safe to persist.

## Implemented operations

The Go package at `server/decimal` provides normalization, quantization, parsing and canonical
serialization; addition, subtraction, multiplication, division, powers, logarithms,
exponentials, floor, comparison, sign, absolute value, and min/max operations.

The TypeScript boundary helpers live in `client/src/numeric.ts`. Both languages also implement:

- `sumGeometricSeries`: exact closed-form cost for a non-negative integer purchase count.
- `affordGeometricSeries`: a closed-form estimate followed by correction and, when necessary,
  binary search until both affordability inequalities are proven.

The maximum-affordable result is always an exact integer. Floating tolerances are test metrics
for continuous calculations; they never decide gameplay authority.

## Verification and maintenance

Run the complete local gate from the repository root:

```sh
make verify
```

This runs Go tests and static analysis, strict TypeScript checking, the Node/V8 suite, and the
same TypeScript vectors in Chromium, Firefox, and WebKit. First-time setup is documented in the
root README.

The committed `testdata/decimal-vectors.json` contains more than 6,000 cases produced with a
seeded RNG and the real pinned JavaScript library. Regenerate it with:

```sh
make vectors
```

Any change to the client library, Go arithmetic, quantization, canonical grammar, or economy
helpers must regenerate the vectors when appropriate and keep both language suites and all
three browser engines green. Continuous cross-runtime results allow at most `1e-12` symmetric
relative error; canonical strings, ordering, state validity, committed transition fixtures,
and affordable counts are exact.

## Deliberate limits

Tetration and layered numbers are not implemented. They are unnecessary for the accepted game
design and require a new RFC if future balance work approaches the layer-0 exponent boundary.
Player-facing number notation is also separate from this storage and calculation contract.

## Known boundary gaps

The normal canonical gameplay path is verified, but a post-implementation audit found client
boundary cases involving non-normalized scientific coefficients, deliberately unsafe/mutated
`break_infinity.js` objects, and diagnostic infinity classification. Cross-runtime diagnostic
coverage also needs explicit zero/division/logarithm/overflow cases. These values are already
rejected from authoritative state, so existing canonical values are unaffected.

The active [Numeric Core Boundary Hardening RFC](../rfc/numeric-core-boundary-hardening.md)
specifies the required fixes and regression gates. This section must be removed or replaced by
the corrected canonical behavior when that follow-up is implemented.
