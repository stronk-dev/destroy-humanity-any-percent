# Numeric Core Boundary Hardening — Implementation Plan

- **RFC:** `rfc/numeric-core-boundary-hardening.md`
- **Assignee:** Codex
- **Status:** implementing
- **Started:** 2026-07-27

## Scope

Close only the public-boundary and diagnostic-coverage gaps specified by the numeric-core
hardening follow-up. Do not add resource, ledger, save, production, transport, or display
notation abstractions here.

## Work breakdown

1. Enforce normalized state invariants in Go and TypeScript.
2. Normalize cloned TypeScript inputs before quantization/canonical serialization and prove
   equivalent scientific coefficients plus no caller mutation.
3. Correct infinity-sentinel classification and define explicit arithmetic domain edges.
4. Extend deterministic vectors with mandatory per-class coverage and bring the Go port into
   exact diagnostic parity.
5. Run Node, browser, Go, vet, deterministic-generation, and fuzz gates.
6. Update canonical docs and archive the follow-up RFC and planning record.

## Acceptance gates

- Every criterion in `rfc/numeric-core-boundary-hardening.md` is represented by a regression
  test.
- The vector generator remains byte-for-byte deterministic and reports non-zero mandatory
  diagnostic coverage.
- `make verify`, `git diff --check`, and a documented Go fuzz pass are green.
- No unrelated RFC-0002 or design work is changed.

## Resume point

Read this plan and `log.md`, then continue the first incomplete work item. Keep implementation
in small local commits and do not push.
