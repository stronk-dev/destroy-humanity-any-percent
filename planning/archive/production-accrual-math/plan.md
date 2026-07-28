# Production Accrual Math — Implementation Plan

- **Status:** completed 2026-07-28
- **RFC:** `rfc/archive/production-accrual-math.md`
- **Assignee:** Codex
- **Started:** 2026-07-28

## Work

1. [x] Define the shared vector corpus, including invalid and permutation cases.
2. [x] Implement deterministic constant-rate integration in Go.
3. [x] Implement the matching TypeScript primitive.
4. [x] Run Go, Node, and all-browser parity tests.
5. [x] Update canonical numeric docs and archive the RFC/planning record.

## Acceptance gates

- Both implementations consume one vector file and agree on every canonical result/error.
- Inputs are validated and source slices/arrays are not mutated.
- Summation is deterministic across permutations and quantizes only once.
- Full repository verification remains green.
