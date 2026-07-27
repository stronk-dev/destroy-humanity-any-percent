# RFC-0001: The Numeric Core

- **Status:** implemented
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-07-27
- **Design refs:** `design/06-tech.md §3` (big numbers), `design/02-economy-balancing.md §2.1` (cost curves)
- **Depends on:** —
- **Supersedes / superseded by:** —
- **Planning:** `planning/archive/0001-numeric-core/`

## Summary

The dual big-number implementation the whole game rests on: pinned `break_infinity.js` 2.2.0 on the client, a hand-written `Decimal` in Go on the server, governed by a **server-authoritative numeric contract** (canonical wire strings, state quantization, and category-appropriate test assertions — see `design/research/numeric-core.md`). Includes the geometric-series helpers the economy needs.

> **Amended 2026-07-27:** the original "bit-for-bit agreeable" requirement was unsound — ECMAScript and Go both specify transcendental functions as implementation-approximated, so exact cross-runtime string equality of raw operations is not achievable (the C# port's own compatibility tests use 1e-13 tolerances). Replaced by the contract below.

## Motivation

Idle-game values exceed float64's native range fast. Client and server must implement the same numeric contract and reach the same discrete gameplay decisions; continuous client predictions may differ within the bounded tolerance below and reconcile to authoritative server state. There is no Go port of break_infinity/break_eternity, so we write one and prove contract compliance by test, not by hope. Out of scope: tetration/`layer` support (add via a follow-up RFC when any design number approaches `1e(9e15)`), display formatting (notations — client-only concern, separate RFC if needed).

## Specification

### Go `Decimal`

```go
type Decimal struct {
    mantissa float64 // sign carried here; |m| in [1,10) or 0
    exponent int64
}
```

- Operations: `Normalize`, `Quantize`, `Add`, `Sub`, `Mul`, `Div`, `Pow` (Decimal^float64), `Log10`, `Exp`, `Ln`, `Cmp`, `Eq/Lt/Lte/Gt/Gte`, `Neg`, `Abs`, `Max/Min`, `Floor`, `FromFloat64`, `FromString`, `String`.
- **Semantics: port `break_infinity.js` operation-for-operation as closely as practical — but cross-runtime agreement is governed by the numeric contract below, not by bit-for-bit equality** (which is unachievable: ECMAScript and Go both specify transcendentals as implementation-approximated).

### The numeric contract (normative — from `design/research/numeric-core.md`)

1. **Discrete vs continuous:** `Decimal` is only for continuous magnitudes (currencies, rates, prices, multipliers). Counts, milestones, sequence numbers, and time units are integers — exact facts, never floats. Cross-runtime counts use Go `int64` / TypeScript `number` and have a technical hardcap of `9,007,199,254,740,991` (`2^53 - 1`, JavaScript's largest exactly representable integer); later system RFCs may set lower visible hardcaps. Integer wire fields at risk of generic JSON-number coercion are strings.
2. **Representation:** signed normalized float64 coefficient + integer base-10 exponent; interoperable exponent domain `|e| < 9e15` (exactly `[-8,999,999,999,999,999, +8,999,999,999,999,999]`). The library's ±`9e15` normalization/sentinel boundary is deliberately excluded from gameplay state. Tetration/layers are deferred (follow-up RFC if any accepted design approaches the boundary).
3. **State quantization:** whenever a calculation is committed as authoritative player state (accrual, purchase, prestige, offline progress, migration, import), `Quantize(12)` rounds to **12 significant decimal digits**, round-half-to-even. For normalized `|m|` in `[1,10)`, round `|m| * 10^11` to the nearest integer with ties to even, divide by `10^11`, reapply the sign, and carry into the exponent if the rounded coefficient is `10`. Intermediates keep full precision; the rule applies only at state-transition boundaries. Golden vectors include positive/negative midpoint, carry-to-next-exponent, zero, and exponent-boundary cases.
4. **Canonical wire grammar (server-emitted, authoritative):** zero is `0` (negative zero canonicalized away); non-zero is `[-]d[.digits]e[-]digits` with coefficient in `[1,10)`, ≤12 significant digits, no trailing zeros; exponent without leading `+`/zeros. Examples: `1e0`, `-4.25e-7`, `9.87654321012e123456`. Parse → quantize → re-serialize is exact and idempotent in both suites. Display notation is a separate client concern.
5. **Non-finite values are invalid gameplay state:** NaN/±Infinity may occur as diagnostic arithmetic results but never in state, wire payloads, or saves — a transition producing one fails without mutating state; persistence rejects it.
6. **Authority:** the client predicts; the server evaluates every intent from its own state and returns canonical strings; the client reconciles. Tolerances are never used to authorize purchases, cross thresholds, rank scores, or compare hardcaps — those use server-ordered values and exact integers.
7. **Verified max-affordable:** the logarithmic closed form is an estimate. Compute candidate → evaluate exact series cost → correct down/up until `sum(n) <= cash && sum(n+1) > cash` (fallback binary search + invariant report if correction exceeds a small bound). Only then transact. A tiny negative residual from final quantization clamps to zero within one canonical rounding unit; larger aborts and records an invariant failure.

### Economy helpers (both sides)

- `AffordGeometricSeries(cash, base, r, owned) → n` (max affordable exact integer: Go `int64`, TypeScript `number`, capped as contract §1)
- `SumGeometricSeries(n, base, r, owned) → cost` (`n` and `owned` are exact capped integers; result is `Decimal`)
- Both implemented in Go and TS from the same closed forms (`design/02 §2.1`), included in the golden vectors.

### Golden vectors

- A Node script (`tools/gen-vectors.mjs`) uses the real pinned `break_infinity.js` package to emit `testdata/decimal-vectors.json`.
- Coverage: ≥5,000 cases spanning 1e-300…1e300 and the ±9e15 exponent boundaries, zero, negatives, near cancellation, realistic ratios, and a dedicated economy-helper block (r ∈ {1.07, 1.13, 1.15}).
- **Assertion categories (per the test contract in `design/research/numeric-core.md`):**

| Category | Assertion |
|---|---|
| Canonical parse / normalize / quantize / wire round-trip | exact string equality |
| Ordering, signs, zero, exponent boundaries, invalid-value rejection | exact |
| `add/sub/mul/div` cross-runtime | symmetric relative error ≤ 1e-12 for non-zero finite results; expected zero is exact |
| `pow/log10/ln/exp` cross-runtime | symmetric relative error ≤ 1e-12 for non-zero finite results; exact zero/domain/error classification |
| Authoritative state-transition fixtures | exact canonical server string |
| Max-affordable / bulk purchase | exact integer count + both affordability postconditions |
| Property/fuzz | normalization & round-trip idempotence, monotonicity, no panic, no non-finite from valid domains |

- The 1e-12 threshold aligns with 12-digit quantization and is looser than the 1e-13 the C# port's transcendental tests already need. It is a gate to validate across the corpus, not permission to curate failures.
- Cross-engine confidence: run the TS suite on Node/V8 **and** the supported browser engines (Chromium, Firefox, WebKit — e.g. via playwright). A JS interpreter embedded in Go (goja) is NOT a sufficient oracle (it calls Go's own math underneath, masking exactly the difference under test).

### Repo placement

This RFC includes the minimal scaffolding it needs: Go module (`server/`), TS workspace (`client/`), shared `testdata/`, Makefile targets (`make test`, `make vectors`). Anything beyond that (compose, CI config) belongs to the next RFC.

## Deviations from design

- `design/06-tech.md §3` and its source research assumed bit-for-bit cross-runtime results and a goja oracle. The implementation spike disproved that assumption. This RFC retains the selected Go/client stack and shared vectors but replaces exact raw-result parity with the server-authoritative, category-tested contract above, per `design/research/numeric-core.md`.
- `design/06-tech.md` originally selected `break_eternity.js`. The second implementation/research pass proved that its layer-1 logarithmic representation progressively spends coefficient precision on the exponent, contradicting this RFC's layer-0 mantissa/exponent contract. Marco approved `break_infinity.js` 2.2.0 instead; its `1e(9e15)` range already covers the full interoperable domain. Tetration and `break_eternity` remain deferred to a future RFC.

## Acceptance criteria

1. `make vectors` regenerates the vector file deterministically (seeded RNG).
2. `go test ./...` and `vitest run` both pass the full vector suite under the category assertions above (exact where exact, ≤1e-12 where tolerant).
3. TS suite green on Node + Chromium + Firefox + WebKit engines.
4. Property/fuzz harness in-tree; a documented run with zero contract violations noted in the planning log.
5. No panics on any input; non-finite values rejected from state/wire/saves per contract §5.
6. Max-affordable postconditions (`sum(n) <= cash && sum(n+1) > cash`) hold on every economy-helper vector.

## Open questions

- None blocking. Deferred: `layer` field (tetration) → future RFC.

## Changelog

- 2026-07-27: created; accepted.
- 2026-07-27: **amended** per `design/research/numeric-core.md` (Codex implementation spike): removed the unsound bit-for-bit contract; added the numeric contract (12-digit state quantization, canonical wire grammar, discrete/continuous split, non-finite invalidity, verified max-affordable, server authority); replaced goja-oracle claim with category assertions + real browser-engine coverage.
- 2026-07-27: **amended** after extreme-scale research and owner approval: replaced `break_eternity.js` with pinned `break_infinity.js` 2.2.0 for the layer-0 client. Deferred layered/tetration math remains unchanged.
- 2026-07-27: implemented. Shipped the pinned client and operation-compatible Go core,
  deterministic 6,278-vector corpus, canonical 12-digit state boundary, verified geometric
  helpers, property/fuzz coverage, and green Node/Chromium/Firefox/WebKit suites.
