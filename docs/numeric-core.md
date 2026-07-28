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
configuration/import diagnostics and do not make a value safe to persist. State validation
also enforces the internal representation: zero must use exponent `0`, and every non-zero
mantissa must have absolute value in `[1, 10)`.

The TypeScript boundary clones and normalizes inputs before rounding, so equivalent inputs such
as `12.345e2`, `0.12345e4`, and `1.2345e3` all serialize as `1.2345e3`. It never mutates a
caller-owned `Decimal`, including objects created through the library's unsafe no-normalize
constructor.

The Go normalizer also performs a final carry/borrow correction after logarithmic scaling.
IEEE-754 rounding can otherwise leave an exact boundary mantissa such as `10e87`; the correction
guarantees that valid arithmetic results retain the normalized `[1, 10)` representation required
by authoritative state.

## Implemented operations

The Go package at `server/decimal` provides normalization, quantization, parsing and canonical
serialization; addition, subtraction, multiplication, division, powers, logarithms,
exponentials, floor, comparison, sign, absolute value, and min/max operations.

The TypeScript boundary helpers live in `client/src/numeric.ts`. Both languages also implement:

- `sumGeometricSeries`: exact closed-form cost for a non-negative integer purchase count.
- `affordGeometricSeries`: a closed-form estimate followed by correction and, when necessary,
  binary search until both affordability inequalities are proven.
- `SumDeterministic`: an order-independent n-ary sum that groups same-exponent terms before
  normalization and orders ties by absolute then signed mantissa.
- `accrueConstant`: deterministic closed-form production from non-negative per-second rate
  sources, exact elapsed milliseconds, and a non-negative efficiency multiplier.

The maximum-affordable result is always an exact integer. Floating tolerances are test metrics
for continuous calculations; they never decide gameplay authority.

The Go server additionally exposes detailed affordability queries that report whether verification
left the bounded local-correction path for binary-search fallback. Count and error semantics are
identical to the original API. The production purchase handler turns that diagnostic into its
transaction-local invariant event/audit contract; numeric helpers remain transport-free.

### Constant-rate production accrual

The Go primitive in `server/production` and TypeScript primitive in `client/src/production.ts`
share this boundary:

```text
accrueConstant(rates[], elapsedMilliseconds, efficiency) -> Decimal
```

Each source is an authoritative per-second Decimal state value. Elapsed time is an exact integer
from zero through `9,007,199,254,740,991` milliseconds. Both implementations validate
non-negative inputs, sum a copy through the deterministic n-ary aggregation boundary, multiply by
elapsed seconds and efficiency, and quantize the final delta once. They never mutate or
individually quantize the caller's sources.

This function is deliberately policy-free. A 90% efficiency and a 24-hour interval appear in the
shared vectors; the implemented [production engine](production-engine.md) owns offline efficiency,
the server clock, caps, and ledger commit. Invalid inputs or non-finite intermediates fail and
cannot become state.

### Arithmetic domain edges

The Go port mirrors pinned `break_infinity.js` behavior, including its unconventional finite
zero results for division by zero, zero divided by zero, `0` raised to a negative power,
opposite-infinity addition, and infinity multiplied by zero. Logarithm of zero produces
negative infinity; negative logarithm inputs produce NaN; positive exponent overflow produces
the signed infinity diagnostic.

Multiplication, division, and integer powers normalize their final arithmetic result at the
exponent edge. This preserves a mantissa carry that rescues a valid lower-bound value, avoids a
transient reciprocal overflow during division, and matches the pinned JavaScript diagnostic for
true overflow/underflow. A geometric ratio whose representable difference from one is zero uses
the constant-price limit rather than dividing by zero.

NaN and infinity are rejected from state and canonical wire values. The finite-zero compatibility
cases cannot carry their origin, so feature handlers must validate domains before calculation:
a purchase, conversion, rate, or modifier may never rely on the library result of an invalid
division or power operation.

## Verification and maintenance

Run the complete local gate from the repository root:

```sh
make verify
```

This runs Go tests and static analysis, strict TypeScript checking, the Node/V8 suite, and the
same TypeScript vectors in Chromium, Firefox, and WebKit. First-time setup is documented in the
root README.

The committed `testdata/decimal-vectors.json` uses schema 3 and contains 6,295 cases produced
with a seeded RNG and the real pinned JavaScript library. It includes recomputed
operation/classification counts, 22 mandatory named domain edges, and an assertion that random
binary inputs reach above absolute exponent `4e15`, so upper-domain coverage cannot silently
disappear. Regenerate it with:

```sh
make vectors
```

The generator is a pinned-runtime compatibility oracle, not a wholly independent mathematical
oracle: its canonical and quantization helpers intentionally mirror the client boundary. Changes
to those helpers therefore require direct review in addition to regenerated agreement.

The hand-reviewed `testdata/production-accrual.json` schema 1 corpus covers the production
primitive's exact results and rejected domains, including permutation invariance, sub-resolution
sources, a 24-hour 90% example, huge exponents, and the maximum exact elapsed time. Go, Node, and
all three browser engines consume the same file.

Any change to the client library, Go arithmetic, quantization, canonical grammar, or economy
helpers must regenerate the vectors when appropriate and keep both language suites and all
three browser engines green. Continuous cross-runtime results allow at most `1e-12` symmetric
relative error; canonical strings, ordering, state validity, committed transition fixtures,
and affordable counts are exact.

## Deliberate limits

Tetration and layered numbers are not implemented. They are unnecessary for the accepted game
design and require a new RFC if future balance work approaches the layer-0 exponent boundary.
Player-facing number notation is also separate from this storage and calculation contract.
