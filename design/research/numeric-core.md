# Research: Numeric Core for an Authoritative Idle Game

> Codex research pass, 2026-07-27. This report repairs the evidence gap exposed while
> implementing RFC-0001. It distinguishes numeric range, useful precision, cross-runtime
> prediction, authoritative economy decisions, and persistence. The RFC remains the
> implementation authority; the contract below is a recommendation until accepted there.

## Research question

Cloud Clicker moves from single units through familiar named magnitudes and potentially far
beyond them. What representation, rounding policy, wire format, and test contract let a Go
server and browser client handle that scale quickly without rounding deciding purchases,
unlocks, or saves incorrectly?

## Executive finding

The selected coefficient-plus-exponent representation has vastly more **range** than the
current design needs. Its approximately 15–17 significant decimal digits of **precision** are
also more than an idle-game economy can use. The original requirement for bit-for-bit result
strings from independent JavaScript and Go calculations is nevertheless unsound: the two
runtimes do not promise identical transcendental math, and `break_infinity.js` and
`break_eternity.js` do not use identical algorithms.

Cloud Clicker should keep the selected architecture, but make the server's canonical value
authoritative and treat the browser calculation as a prediction. Persisted resource values
should be quantized to 12 significant decimal digits at state-transition boundaries. Raw
cross-runtime arithmetic should be tested numerically; canonical parsing and gameplay
decisions should be tested exactly.

## 2026-07-27 follow-up: library fit at extreme exponents

Implementation against the pinned `break_eternity.js` 2.1.3 package exposed a second flaw in
the original analysis: the claim below that coefficient precision remains roughly constant as
the exponent grows is true for a coefficient-plus-exponent type such as `break_infinity.js`, but
not for `break_eternity.js` layer 1.

The distinction matters:

- `break_infinity.js` stores a normalized `float64` mantissa and a separate integer exponent;
  its documented range is approximately `1e(9e15)`.
- `break_eternity.js` stores `sign`, `layer`, and one `float64` magnitude. At layer 1 that
  magnitude is `exponent + log10(mantissa)`. The exponent and coefficient therefore compete for
  the same 53 binary precision bits. Its source explicitly says the mantissa eventually stops
  being relevant; the extra range is purchased for games using tetration and higher hyper
  operators.

An empirical round-trip through the pinned client library, using the input coefficient
`1.23456789012`, gives:

| Exponent | Recovered coefficient | Relative coefficient loss |
|---:|---:|---:|
| `21` | `1.2345678901200037` | `3.1e-15` |
| `300` | `1.2345678901200137` | `1.1e-14` |
| `1,000` | `1.2345678901198522` | `1.2e-13` |
| `4,096` | `1.2345678901208217` | `6.7e-13` |
| `1,000,000` | `1.2345678899760382` | `1.2e-10` |
| `1,000,000,000,000` | `1.2346752269846848` | `8.7e-5` |
| `8,999,999,999,999,999` | `1` | `0.19` |

This is graceful precision loss, not overflow: the order of magnitude remains representable and
useful for spectacle. It does, however, disprove RFC-0001's simultaneous promises of a
12-significant-digit canonical coefficient and `1e-12` relative client/server parity throughout
the entire exponent domain.

### Fit to Cloud Clicker

The accepted design currently requires familiar named magnitudes, geometric generator prices,
multiplicative buff windows, cube-root prestige, offline accrual, and a finite ending reached in
roughly two months. It contains no tetration, iterated logarithm, or mechanic approaching
`1e(9e15)`; RFC-0001 explicitly defers layers. A sextillion is only `1e21`, and even an
aggressively inflated endgame in the thousands or millions of decimal orders remains far inside
`break_infinity.js`'s range.

Accordingly, the best-fit layer-0 client for the currently specified game is
`break_infinity.js`, not `break_eternity.js`:

1. Its native mantissa/exponent representation matches the RFC and Go implementation.
2. It preserves approximately constant coefficient precision across the entire required
   exponent range instead of progressively spending those bits on the exponent.
3. It includes the incremental-game arithmetic and geometric-series helpers relevant here.
4. The selected `@antimatter-dimensions/notations` package documents `break_infinity.js` as its
   companion numeric type.
5. Its `1e(9e15)` range is already so far beyond the current design that switching to a layered
   representation now buys no player-visible capability.

`break_eternity.js` remains the appropriate migration target if a later accepted mechanic uses
tetration or outgrows a safe-integer exponent. That migration should be a new RFC with a
`sign/layer/magnitude` server representation and layer-aware wire grammar, rather than silently
pretending it is the same numeric contract.

**Decision approved by Marco:** amend the binding stack and RFC-0001 client from
`break_eternity.js` to pinned `break_infinity.js` 2.2.0. Keep 12-significant-digit authoritative
quantization, canonical strings, exact discrete counts, server-only authorization, and verified
bulk-buy postconditions. Add a balance-harness acceptance gate that records the maximum exponent
reached by every scripted route; draft the layered-number RFC only if an accepted route consumes
a meaningful fraction of the `9e15` exponent budget.

Primary evidence: [`break_infinity.js` representation and range](https://github.com/Patashu/break_infinity.js/blob/master/src/decimal.ts),
[`break_infinity.js` project rationale](https://github.com/Patashu/break_infinity.js),
[`break_eternity.js` layer representation and mantissa limits](https://github.com/Patashu/break_eternity.js/blob/v2.1.3/src/index.ts),
and [`@antimatter-dimensions/notations` setup](https://github.com/antimatter-dimensions/notations).

## Range is not precision

A number such as `6.02e23` has two useful parts:

- `6.02` is the coefficient (also called the mantissa or significand);
- `23` says where the decimal point belongs.

Keeping those parts separately means the program never has to store all of the zeros. A
sextillion is only around `1e21`. A layer-0 coefficient plus a safe JavaScript integer exponent
can reach approximately `1e(9e15)`: nine quadrillion decimal orders of magnitude. That is not
merely larger than sextillions; it leaves an enormous margin over every value currently in the
design.

With a coefficient-plus-exponent type, the coefficient uses a binary 64-bit floating-point value
and retains roughly 15–17 useful decimal digits while the exponent grows. The follow-up above
clarifies why this statement does not apply to `break_eternity.js` layer 1. At `1e30`, adding `1`
disappearing is expected. The game cannot display or balance around a one-unit difference at
that scale anyway. What must not disappear is a discrete fact such as “the player can afford 7
generators but not 8.”

`break_eternity.js` itself declares 17 as the maximum significant digits it assumes and uses
`9e15` as its layer transition limit. [Source](https://github.com/Patashu/break_eternity.js/blob/v2.1.3/src/index.ts)
`break_infinity.js`, which is designed to prefer speed over arbitrary precision, deliberately
rounds during addition and ignores an addend more than 17 decimal orders below the larger
value. [Source](https://github.com/Patashu/break_infinity.js/blob/master/src/decimal.ts)

## Why exact Go/browser operation strings are not a valid contract

IEEE-754 binary64 gives both languages a common shape for ordinary floating-point numbers,
but it does not make every higher mathematical function identical. ECMAScript specifies
`Math.exp`, `Math.log`, and `Math.log10` as implementation-approximated results, so different
browser engines may choose different algorithms. [ECMAScript 2025](https://tc39.es/ecma262/2025/multipage/numbers-and-dates.html)
Go's `math` package also explicitly declines to guarantee bit-identical results across
architectures. [Go math documentation](https://go.dev/pkg/math/?m=old)

The implementation spike found this difference in practice: V8 and Go can produce adjacent
last-place results for `log10`, `pow`, and `exp`. It also found a representation mismatch in
ordinary layer-0 `break_eternity.js` values: normalizing an already-rounded binary value into a
separate coefficient and exponent is not reversible for every possible bit pattern.

Established ports account for this rather than requiring strings to match. `BreakInfinity.cs`
defines relative-tolerance equality, and its compatibility tests allow a `1e-13` tolerance for
functions including `Exp`. [Implementation](https://github.com/Razenpok/BreakInfinity.cs/blob/master/BreakInfinity/BigDouble.cs),
[compatibility tests](https://github.com/Razenpok/BreakInfinity.cs/blob/master/BreakInfinity.Tests/DoubleCompatibilityTests.cs)

A JavaScript interpreter embedded in Go is not a sufficient cross-runtime oracle when it calls
Go's own math functions underneath. It can validate control flow, but it can mask exactly the
browser-versus-Go difference the test is meant to find.

## Recommended numeric contract

### 1. Separate continuous magnitudes from exact discrete state

Use `Decimal` only for values that scale across orders of magnitude: currencies, production
rates, prices, multipliers, and accumulated production.

Use integer types for generator ownership, purchase counts, milestones, sequence numbers, and
time units. These values are facts, not approximations. Counts crossing a client representation
limit must have an explicit, visible hardcap or a string/BigInt wire representation; they must
not silently become floating-point counters.

### 2. Keep the layer-0 coefficient/exponent representation

The Go representation remains a signed normalized `float64` coefficient plus an integer base-10
exponent. The interoperable gameplay domain is `|e| < 9e15`, excluding `break_infinity.js`'s
±`9e15` normalization/sentinel boundary; JavaScript represents every exponent inside that domain
exactly.

Tetration/layers remain deferred. Add them only if an accepted design actually approaches the
layer-0 exponent boundary; familiar named magnitudes do not come remotely close.

### 3. Quantize authoritative resource state to 12 significant decimal digits

Use round-to-nearest, ties-to-even, to 12 significant decimal digits whenever a calculation is
committed as authoritative player state: production accrual, a purchase or reward, prestige,
offline progress, migration, or save import.

Intermediate steps within one formula retain the runtime's full precision. Rounding every
primitive operation would make formula rearrangement change balance and would accumulate more
error; rounding only at a state-transition boundary makes the rule observable and testable.

Twelve digits are intentionally conservative. They leave several guard digits below the
underlying libraries' working precision and far exceed player-facing formatting, which normally
shows only a few. The resulting maximum relative quantization is about `5e-12`. This choice is a
game-state stability policy, not a claim that independent transcendental calculations can never
land on opposite sides of a decimal rounding midpoint.

Derived values such as a displayed rate may retain full precision because they are recomputed;
the resource balance produced from that rate is quantized when committed. Balance/config inputs
remain decimal strings rather than binary literals where their written decimal value matters.

### 4. Define a project wire format instead of inheriting `toString()`

The server emits the only authoritative resource string. Canonical finite values use normalized
scientific notation independent of JavaScript's display rules:

- zero is `0` (negative zero is canonicalized away);
- non-zero is `[-]d[.digits]e[-]digits`;
- the coefficient is in `[1, 10)`, has at most 12 significant digits, and has no trailing zeros;
- the exponent has no leading `+` or leading zeros.

Examples: `1e0`, `-4.25e-7`, `9.87654321012e123456`.

Parsing, quantization, and re-serialization of a canonical value must be exact and idempotent in
both suites. Player-facing notation is a separate client-only concern. NaN and infinities may
exist as diagnostic arithmetic results but are invalid in gameplay state, wire payloads, and
saves. A transition producing one fails without mutating state; persistence rejects it.

### 5. Authority and comparisons

The browser predicts for smooth presentation. The server evaluates every intent from its own
state, returns the canonical resulting strings, and the client reconciles to them. A tolerance is
never used to authorize a purchase, cross an unlock threshold, rank a score, or compare a hardcap.
Those decisions use the server's ordered values and exact integer state.

Approximate equality is only for tests of independently computed continuous results and, if
useful, deciding whether a visual reconciliation can animate instead of snap.

### 6. Verify closed-form bulk buying

The logarithmic inverse for max-affordable is an estimate, not the final authority:

1. Compute a candidate count from the closed form.
2. Evaluate the geometric-series cost for that integer candidate.
3. Decrement until `cost(candidate) <= cash`.
4. Increment while `cost(candidate + 1) <= cash`.
5. If local correction exceeds a small bound, fall back to monotonic binary search and report an
   invariant failure for diagnosis.

The required postcondition is exact in server semantics:

```text
sum(n) <= cash  and  sum(n + 1) > cash
```

Only then may the transaction subtract the cost. A tiny negative residual caused solely by final
quantization may clamp to zero within one canonical rounding unit; a larger negative aborts the
transaction and records an invariant failure. This makes rounding unable to grant an extra item
or consume money the player did not have.

## Test contract

The shared suite should contain distinct categories instead of asserting one kind of equality
for every operation:

| Category | Required assertion |
|---|---|
| Canonical parse, normalize, quantize, and wire round-trip | Exact string equality |
| Ordering, signs, zero, exponent boundaries, invalid-value rejection | Exact result |
| `add`, `sub`, `mul`, `div` cross-runtime result | Relative error within `1e-12`, with an absolute rule around zero |
| `pow`, `log10`, `ln`, `exp` cross-runtime result | Relative error within `1e-12`; exact domain/error classification |
| Authoritative state-transition fixtures | Exact canonical server string |
| Max-affordable and bulk purchase | Exact integer count plus both affordability postconditions |
| Property/fuzz tests | Normalization, round-trip idempotence, monotonicity, no panic, no non-finite result from valid positive-domain cases |

The committed corpus should still span ordinary values, `1e-300...1e300`, very large positive
and negative exponents, near cancellation, and realistic ratios. Cross-engine confidence should
exercise Node/V8 and at least the supported Chromium, Firefox, and WebKit browser engines. The
server's Go tests remain the gameplay authority; browser discrepancies outside the tolerance are
prediction defects, not disputes over player state.

The `1e-12` arithmetic threshold aligns with the 12-digit state contract and is deliberately
looser than the `1e-13` tolerance already needed in an established port's transcendental
compatibility tests. It is a gate to validate empirically across the full corpus, not permission
to curate away failures. A failure requires fixing the port, tightening the operation domain, or
revisiting this contract in the RFC.

## Alternatives considered

### Arbitrary-precision decimal

This stores more digits but does not solve cross-language transcendental semantics by itself. It
adds CPU, allocation, payload, and implementation complexity for precision the economy cannot
meaningfully use. It remains inappropriate unless a future mechanic genuinely depends on many
exact significant digits rather than many orders of magnitude.

### Fixed decimal coefficient with integer-only arithmetic

A fixed significant-digit coefficient plus exponent and specified integer rounding could make
basic operations deterministic in both languages. `pow`, `log`, and `exp` would still need a
shared approximation specification. This is warranted only if exact deterministic replays across
runtimes become a product requirement; the current server-authoritative design does not require
it.

### One shared WebAssembly implementation

One numeric implementation could remove duplication, but it would complicate the Go server and
the browser build and would make a core server dependency less idiomatic. It is disproportionate
while client calculations are explicitly predictions.

## RFC-0001 implications

RFC-0001 should be amended before implementation resumes:

- remove “bit-for-bit agreeable” and exact raw-operation string requirements;
- define 12-significant-digit state quantization and the canonical wire grammar;
- distinguish exact discrete integers from approximate continuous magnitudes;
- make non-finite gameplay values invalid instead of round-trippable save values;
- require verified/corrected max-affordable decisions;
- replace the goja-only confidence claim with numerical and actual-browser-engine coverage;
- state that the server's canonical result is authoritative and client prediction reconciles.

This is a correction to the numeric contract, not a change to the game's economy or rendering
architecture.
