# Production Contract Assertions & Integrity — implementation plan

- **RFC:** `rfc/production-contract-integrity.md`
- **Assignee:** Codex
- **Started:** 2026-07-29

## Sequence

1. Add the neutral multiplier ordering authority and Go/TypeScript declaration/runtime contract
   tests (D1).
2. Export the invariant types/sink and assert applied, rejected, aborted, and replay paths in unit
   and Postgres tests (D2).
3. Upgrade the formula artifact to schema v2 with normalized-AST provenance and update its drift
   tests/documentation (D3).
4. Serialize the integration target and prove repeatability on fresh Postgres (D4).
5. Enforce migration corpus growth, propagate receipt parse failures, document event retention and
   sticky idempotency, and add regressions (D5).
6. Run `make verify` and Postgres integration gates, review the complete diff, update canonical
   docs, and archive the RFC/planning record.

## Acceptance gates

The seven acceptance criteria in the RFC are binding. No Balance Harness baseline is generated
until this RFC is implemented, archived, and green.
