# Save Layer & Migrations — Implementation Plan

- **RFC:** `rfc/save-layer-and-migrations.md`
- **Assignee:** Codex
- **Started:** 2026-07-28

## Work

1. Add pinned Goose/pgx dependencies, embedded migrations, and Postgres 16 integration plumbing.
2. Add the economy ledger's validated restore constructor without exposing a balance setter.
3. Implement catalog artifacts/resolution, canonical version-1 state encoding, and migration
   dispatch.
4. Implement owner/scope validation and atomic stream create/write/load/archive operations.
5. Add unit and real-Postgres tests for corruption, versioning, revision retention, archive state,
   scope isolation, and concurrent conflicts.
6. Run the dedicated integration suite and complete cross-runtime verification.
7. Update canonical docs and archive the RFC/planning record.

## Acceptance gates

- All ten RFC acceptance criteria have direct automated coverage.
- SQL migrations are embedded, sequential, transactional, and idempotently reach version 1.
- Integration tests run against Postgres 16 rather than SQLite or a SQL mock.
- Failed validation and revision conflicts leave the current stored revision unchanged.
- Exactly five complete revisions survive pruning.
- `make verify` remains green across Go, Node, and three browser engines.
