# Generator Production State — Implementation Plan

- **RFC:** `rfc/generator-production-state.md`
- **Assignee:** Codex
- **Started:** 2026-07-28

## Work

1. Extend the strict schema and Go catalog with versioned production definitions.
2. Add exact per-scope generator ownership and canonical cursor to save version 2.
3. Migrate save version 1 deterministically from its revision timestamp.
4. Update repository operations and integration tests to carry complete state.
5. Run full verification, update canonical docs, and archive the RFC/planning record.

## Acceptance gates

- Catalog versions remain historically readable and v2 graph validation is strict.
- Save payloads contain exactly the scoped resources and generators declared by their catalog.
- No client time, float count, partial state, or implicit cross-scope production is accepted.
- Existing revision/concurrency/retention guarantees remain green.
