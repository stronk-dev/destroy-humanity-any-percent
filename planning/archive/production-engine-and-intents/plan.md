# Production Engine & Intent API — implementation plan

- **RFC:** `rfc/archive/production-engine-and-intents.md`
- **Status:** implemented
- **Assignee:** Codex
- **Started:** 2026-07-28

## Work breakdown

1. **Catalog contract v3**
   - Add strict Go, TypeScript, and JSON Schema support for manual actions, multiplier-source
     declarations, T0–T3 progress coordinates, and offline/Compute Credit policy.
   - Preserve historical catalog v1/v2 loading and shared cross-runtime validation fixtures.
   - Ship one mechanical Phase-0 production catalog; values remain provisional balance data.
2. **Save state v3**
   - Persist `compute_credit_ms`, exact manual milli-tokens, and the server-authored refill cursor.
   - Extend the migration corpus from v1 and v2; reject unsafe integers, malformed cursors, unknown
     fields, and scope-invalid production state.
3. **Production evaluator**
   - Compute generator rates through the fixed slot order and validated runtime contributions.
   - Evaluate trusted online/offline intervals, hardcap headroom, credit banking, and clock rollback.
   - Implement T0–T3 progress coordinates in Go and TypeScript with shared fixtures.
4. **Atomic persistence envelope**
   - Add Postgres migrations for intent records and immutable events.
   - Add a locked-stream mutation primitive that replays idempotent outcomes and atomically writes
     applied state, revision, events, and intent receipt; terminal rejection records mutate no save.
   - Expose bounded 30-day intent-record pruning for the future deployment scheduler.
5. **Authoritative intent handlers**
   - Strictly validate and canonically hash `buy_generator` and `perform_manual_batch` requests.
   - Apply working-state accrual plus the requested action, produce typed receipts/rejections, and
     route affordability diagnostics through the transaction-local invariant sink.
   - Cover exact/max buying, silent manual clamp, replay/conflict behavior, hardcaps, and rollback.
6. **Acceptance and close-out**
   - Run the 24-hour × 200-seed local property test and the complete `make verify` gate.
   - Run the real Postgres integration suite when the disposable database is available.
   - Update canonical docs, mark the RFC implemented, and archive both RFC and planning records.

## Acceptance gates

- All six RFC acceptance criteria pass, with the large 30-day harness explicitly deferred.
- Go and TypeScript agree on catalog/progress fixtures and every browser suite remains green.
- Save migrations cover every prior version and no non-finite or unsafe numeric state persists.
- Postgres proves applied mutation atomicity, deterministic rejection replay, key/hash conflict, and
  no duplicate purchase event under concurrent replay.
- `make verify`, `git diff --check`, and `make validate-migrations` (if present) are green.

## Commit boundaries

1. RFC acceptance and planning start.
2. Catalog v3 + shared fixtures.
3. Save v3 + migration corpus.
4. Production evaluation + progress parity.
5. Atomic intent/event persistence.
6. Intent handlers + property/integration tests.
7. Canonical docs + archive.

## Completion

All work items and acceptance gates completed on 2026-07-28. The large 200-bot × 30-day balance
harness remains deliberately owned by the future Balance Harness RFC; this RFC's 24-hour × 200-seed
gate is implemented and blocking.
