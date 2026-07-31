# Guild Model — implementation plan

- **Assignee:** Codex
- **RFC:** `rfc/guild-model.md`
- **Started:** 2026-07-30

1. [x] Add the strict Guild catalog and schema, register it as an append-only epoch artifact,
   mint the required balance epoch, and prove the epoch/baseline protocol.
2. [x] Add guild, membership-history, application/invitation, event, projection, XP, Health-input,
   and exchange persistence with migrations and database invariants.
3. [x] Implement the closed account-scoped lifecycle intent surface with C1-compatible exact
   schemas, idempotency, typed rejections, transactional member/leader/cap enforcement, and events.
4. [x] Integrate server-derived production tithe, deterministic GC clearing, the 0-ppm consumption
   hook, NPC fallback, receipts, replay idempotency, and the canonical accrual boundary order.
5. [x] Replace the deny-all `guild:*` transport resolver with active-membership authorization and
   relay membership presence changes through the existing transactional outbox path.
6. [x] Publish the shipped lifecycle/tithe/Health/exchange formulas and operations in canonical
   docs; run focused unit/schema/boundary tests, real-Postgres concurrency/integration, the full
   verification suite, and an independent diff review before archival.

Acceptance gates are the RFC's seven criteria. Open-question work stays limited to the declared
seven-day disband sweep; officer permissions and client surfaces remain successor RFC scope.
