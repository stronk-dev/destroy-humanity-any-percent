# Resource-Log Domain Parity — Implementation Plan

- **Status:** completed 2026-07-29
- **RFC:** `rfc/archive/resource-log-domain-parity.md`
- **Assignee:** Codex
- **Started:** 2026-07-29

## Sequence

1. [x] Add the `5e-15` semantic floor and positive-log defensive check to both catalog loaders.
2. [x] Route TypeScript resource-log evaluation through Decimal division and add a source assertion
   against native division.
3. [x] Extend shared catalog/progress fixtures at rejected, boundary, representative, and shipped
   target values; wire semantic schema verification.
4. [x] Update canonical economy and production documentation.
5. [x] Run focused and complete cross-runtime verification, archive the RFC/planning record, and make
   reviewable local commits.

## Acceptance gates

- Go, Node, Chromium, Firefox, and WebKit reject `1e-16`, `1e-15`, and `4e-15`, and accept `5e-15`,
  `9e-15`, and shipped targets for top-level and composite coordinates.
- Shared progress cases at `5e-15` and representative magnitudes produce identical canonical
  results in Go and TypeScript.
- TypeScript progress uses Decimal `.div`; an automated source assertion rejects native `/` in the
  resource-log evaluator.
- Existing `div-zero` and `zero-div-zero` mandatory vectors remain unchanged and green.
- `make verify-schema` combines JSON shape validation with the semantic target check.
- Canonical economy and production docs publish the target domain.
- `gofmt`, focused tests, and `make verify` pass.
