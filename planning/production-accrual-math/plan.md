# Production Accrual Math — Implementation Plan

- **RFC:** `rfc/production-accrual-math.md`
- **Assignee:** Codex
- **Started:** 2026-07-28

## Work

1. Define the shared vector corpus, including invalid and permutation cases.
2. Implement deterministic constant-rate integration in Go.
3. Implement the matching TypeScript primitive.
4. Run Go, Node, and all-browser parity tests.
5. Update canonical numeric docs and archive the RFC/planning record.

## Acceptance gates

- Both implementations consume one vector file and agree on every canonical result/error.
- Inputs are validated and source slices/arrays are not mutated.
- Summation is deterministic across permutations and quantizes only once.
- Full repository verification remains green.
