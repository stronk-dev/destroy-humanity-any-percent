# RFC-0001 Implementation Plan — Numeric Core

- **RFC:** `rfc/0001-numeric-core.md`
- **Assignee:** Codex
- **Status:** implemented
- **Started:** 2026-07-27

## Scope

Implement only RFC-0001: a layer-0 Go `Decimal` compatible with pinned
`break_infinity.js` 2.2.0 behavior under the amended server-authoritative numeric contract,
TypeScript and Go geometric-series helpers, a seeded shared golden-vector generator,
category-appropriate exact/tolerant tests, property/fuzz coverage, and browser-engine
confidence tests. Tetration/layers, player-facing formatting, runtime client UI, deploy
scaffolding, and CI configuration remain out of scope.

## Work breakdown

1. Create the minimal `server/`, `client/`, `testdata/`, and `tools/` structure plus
   root Make targets required by the RFC.
2. Pin the JavaScript numeric/test dependencies and implement a deterministic vector
   generator backed by the real `break_infinity.js` package.
3. Port the layer-0 decimal representation and operations to Go, including 12-digit
   quantization, canonical wire strings, comparisons, and diagnostic non-finite propagation.
4. Implement geometric-series sum and verified max-affordable helpers with exact capped
   integer counts in Go and TypeScript.
5. Load the committed vector file in both suites and apply the RFC's exact or tolerant
   assertion category to each case.
6. Add property/fuzz coverage and run the TypeScript suite on Node, Chromium, Firefox,
   and WebKit; do not use goja as the cross-runtime oracle.
7. Run formatting, static checks, vector reproducibility, both language suites, browser
   suites, and a documented confidence fuzz pass.
8. Document the shipped numeric core, update RFC/index status, and archive this planning
   directory when every gate is green.

## Acceptance gates

- `make vectors` produces a byte-for-byte deterministic `testdata/decimal-vectors.json`
  containing at least 5,000 broad-coverage cases and realistic economy cases.
- `go test ./...` passes from the repository root.
- `pnpm --dir client run test` passes against the same committed vector file.
- Canonical/state/decision vectors match exactly; continuous arithmetic satisfies the
  symmetric `1e-12` relative-error contract and exact domain classification.
- Node, Chromium, Firefox, and WebKit suites pass; property/fuzz coverage completes a
  documented run with no contract violations.
- NaN and infinities never panic and are rejected from gameplay state, wire, and saves.
- Max-affordable returns an exact capped integer and proves both affordability postconditions.
- `gofmt` and `go vet ./...` pass.
- Canonical docs describe the implemented representation, supported operations, wire
  format, vector regeneration, testing, and deferred limits.

## Resume point

Implementation is complete and this directory is archived. Read `docs/numeric-core.md` for
canonical behavior; use this plan and `log.md` only as the historical implementation record.
