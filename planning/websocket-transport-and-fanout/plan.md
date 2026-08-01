# WebSocket Transport & Fan-out — implementation plan

- **Assignee:** Codex
- **RFC:** `rfc/websocket-transport-and-fanout.md`
- **Started:** 2026-07-29

1. [x] Add strict Phase-0 transport policy and closed Go/TypeScript wire envelopes.
2. [x] Embed Centrifuge with the implemented channel authorization, history, and backpressure policy.
3. [x] Map committed production receipts/events into one transaction-owned private-player outbox;
   each scope is revision-ordered and snapshot recovery uses the authoritative full-state endpoint.
4. [ ] Compose `cmd/gameserver`, readiness, HTTP intents, and bounded drain lifecycle. (lifecycle
   complete; binary blocked on the Commons owner resolver recorded in the log). This item owns
   wiring the real `guild.Service.PendingSettlements`; account integration uses an explicitly
   empty constructor fixture and is not production composition.
5. [x] Add recovery, overflow, authz, drain, and 5k-connection soak coverage.
6. [x] Update canonical docs and run full verification.
7. [ ] Record independent review before archival.
