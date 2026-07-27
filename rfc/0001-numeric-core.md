# RFC-0001: The Numeric Core

- **Status:** implementing
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-07-27
- **Design refs:** `design/06-tech.md §3` (big numbers), `design/02-economy-balancing.md §2.1` (cost curves)
- **Depends on:** —
- **Supersedes / superseded by:** —
- **Planning:** `planning/0001-numeric-core/` (to be created when implementation starts)

## Summary

The dual big-number implementation the whole game rests on: `break_eternity.js` on the client, a hand-written `Decimal` in Go on the server, kept bit-for-bit agreeable via a shared golden-vector test file. Includes the geometric-series helpers the economy needs.

## Motivation

Idle-game values exceed float64 fast. Client and server must compute the *same* results (client predicts, server is authoritative; divergence = visible desync and exploitable disputes). There is no Go port of break_infinity/break_eternity, so we write one and prove agreement by test, not by hope. Out of scope: tetration/`layer` support (add via a follow-up RFC when any design number approaches `1e(9e15)`), display formatting (notations — client-only concern, separate RFC if needed).

## Specification

### Go `Decimal`

```go
type Decimal struct {
    mantissa float64 // sign carried here; |m| in [1,10) or 0
    exponent int64
}
```

- Operations: `Normalize`, `Add`, `Sub`, `Mul`, `Div`, `Pow` (Decimal^float64), `Log10`, `Exp`, `Ln`, `Cmp`, `Eq/Lt/Lte/Gt/Gte`, `Neg`, `Abs`, `Max/Min`, `Floor`, `FromFloat64`, `FromString`, `String`.
- **Semantics: port `break_infinity.js` operation-for-operation, replicating its exact normalization and rounding order.** Where break_infinity has documented precision quirks, we replicate the quirk. Agreement > elegance.
- Special values: zero, negatives, ±Infinity, NaN must round-trip `String`/`FromString` identically to the JS library. NaN propagates; nothing panics.
- `String()` emits the wire format: shortest form matching JS `Decimal.toString()` (plain notation below 1e21, `M.MMMe+E` style above, matching JS exactly — this string IS the wire/save format per design law #3).

### Economy helpers (both sides)

- `AffordGeometricSeries(cash, base, r, owned) → n` (max affordable count)
- `SumGeometricSeries(n, base, r, owned) → cost` (bulk price)
- Both implemented in Go and TS from the same closed forms (`design/02 §2.1`), included in the golden vectors.

### Golden vectors

- A Node script (`tools/gen-vectors.mjs`) uses the real `break_eternity.js` (configured to break_infinity-compatible range) to emit `testdata/decimal-vectors.json`:
  `[{ "a": "...", "b": "...", "op": "add|sub|mul|div|pow|log10|exp|cmp|afford|sum", "expect": "..." }, ...]`
- Coverage: ≥5,000 cases spanning magnitudes 1e-300…1e300 and beyond to 1e±9e15 boundaries, zero, negatives, near-equal subtraction, ±Infinity, NaN, and a dedicated block of economy-helper cases with realistic constants (r ∈ {1.07, 1.13, 1.15}).
- The file is committed. Both test suites (`go test`, `vitest`) load the same file and assert **exact string equality** of results.
- One-time confidence pass: Go fuzz (`go test -fuzz`) against goja running the actual JS for a few million ops; keep the fuzz harness in-tree but off CI.

### Repo placement

This RFC includes the minimal scaffolding it needs: Go module (`server/`), TS workspace (`client/`), shared `testdata/`, Makefile targets (`make test`, `make vectors`). Anything beyond that (compose, CI config) belongs to the next RFC.

## Deviations from design

None.

## Acceptance criteria

1. `make vectors` regenerates the vector file deterministically (seeded RNG).
2. `go test ./...` and `vitest run` both pass the full vector file with exact string equality.
3. Fuzz harness exists and a documented run found zero divergences (note the run in the planning log).
4. No panics on any input; NaN/Infinity behavior matches JS.

## Open questions

- None blocking. Deferred: `layer` field (tetration) → future RFC.

## Changelog

- 2026-07-27: created; accepted.
