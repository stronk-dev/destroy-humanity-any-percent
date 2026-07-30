# WebSocket Transport & Fan-out — implementation plan

- **Assignee:** Codex
- **RFC:** `rfc/websocket-transport-and-fanout.md`
- **Started:** 2026-07-29

1. [x] Add strict Phase-0 transport policy and closed Go/TypeScript wire envelopes.
2. [x] Embed Centrifuge with the implemented channel authorization, history, and backpressure policy.
3. [ ] Map committed production receipts/events into the private player channel. (receipts complete;
   event/snapshot relay remains)
4. [ ] Compose `cmd/gameserver`, readiness, HTTP intents, and bounded drain lifecycle. (lifecycle
   complete; binary blocked on the Commons owner resolver recorded in the log)
5. [x] Add recovery, overflow, authz, drain, and 5k-connection soak coverage.
6. [ ] Update canonical docs and run full verification.
7. [ ] Record independent review before archival.
