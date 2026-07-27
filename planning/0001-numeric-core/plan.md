# RFC-0001 Implementation Plan — Numeric Core

- **RFC:** `rfc/0001-numeric-core.md`
- **Assignee:** Codex
- **Status:** implementing
- **Started:** 2026-07-27

## Scope

Implement only RFC-0001: a layer-0 Go `Decimal` compatible with the corresponding
`break_eternity.js` behavior, TypeScript and Go geometric-series helpers, a seeded
shared golden-vector generator, exact shared-vector tests, and an off-CI Go fuzz
harness. Tetration/layers, player-facing formatting, runtime client UI, deploy
scaffolding, and CI configuration remain out of scope.

## Work breakdown

1. Create the minimal `server/`, `client/`, `testdata/`, and `tools/` structure plus
   root Make targets required by the RFC.
2. Pin the JavaScript numeric/test dependencies and implement a deterministic vector
   generator backed by the real `break_eternity.js` package.
3. Port the layer-0 decimal representation and operations to Go, including parsing,
   wire strings, comparisons, and non-finite propagation.
4. Implement geometric-series sum and max-affordable helpers in Go and TypeScript.
5. Load the committed vector file in both suites and assert exact result strings.
6. Add a Go fuzz harness that uses goja to execute the real JavaScript implementation;
   keep it out of normal CI/test execution.
7. Run formatting, static checks, vector reproducibility, both suites, and a documented
   confidence fuzz pass.
8. Document the shipped numeric core, update RFC/index status, and archive this planning
   directory when every gate is green.

## Acceptance gates

- `make vectors` produces a byte-for-byte deterministic `testdata/decimal-vectors.json`
  containing at least 5,000 broad-coverage cases and realistic economy cases.
- `go test ./...` passes from the repository root.
- `pnpm --dir client run test` passes against the same committed vector file.
- Go and TypeScript compare exact wire/result strings for every vector.
- The retained goja fuzz harness completes a documented confidence run with no
  divergence and does not run in the standard test gate.
- NaN and infinities never panic and match the JavaScript reference behavior.
- `gofmt` and `go vet ./...` pass.
- Canonical docs describe the implemented representation, supported operations, wire
  format, vector regeneration, testing, and deferred limits.

## Resume point

Read this file and `log.md`, then continue the first incomplete work item above. Do not
expand scope without a new RFC.
