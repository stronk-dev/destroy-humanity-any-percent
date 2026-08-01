# Gameserver Composition — implementation plan

- **Assignee:** Codex
- **RFC:** `rfc/gameserver-composition.md`
- **Started:** 2026-08-02

1. [x] Implement the projection-owned Commons participation-weight resolver and prove the
   pre-first-sample/live/replay contract.
2. [x] Implement the closed Go/TypeScript world-snapshot schema and the 4 Hz aggregator with a
   strictly monotonic world-owned revision.
3. [x] Compose the production service with the real Guild settlement resolver and every owned
   account, projection, verification, relay, sweep, and GC driver.
4. [x] Compose Account/Commons/Guild transport authorization, leaving only `match:*` deny-closed.
5. [x] Add `cmd/gameserver`, readiness, signal handling, and bounded drain.
6. [x] Prove the composed binary against real Postgres, including the F4 settlement seam, world
   snapshots, real socket relay, worker ticks, and session GC.
7. [x] Update canonical docs and pass the repository verification gates.
8. [ ] Record implementer self-review and obtain a designated independent review before archival.

Checkboxes flip only in the same commit as their proof, per `AGENTS.md`.
