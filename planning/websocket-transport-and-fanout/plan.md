# WebSocket Transport & Fan-out — implementation plan

- **Assignee:** Codex
- **RFC:** `rfc/websocket-transport-and-fanout.md`
- **Started:** 2026-07-29

1. [x] Add strict Phase-0 transport policy and closed Go/TypeScript wire envelopes.
2. [ ] Embed Centrifuge with the implemented channel authorization, history, and backpressure policy.
3. [ ] Map committed production receipts/events into the private player channel. (receipts complete;
   event/snapshot relay remains)
4. [ ] Compose `cmd/gameserver`, readiness, HTTP intents, and bounded drain lifecycle. (lifecycle
   complete; binary blocked on the Commons owner resolver recorded in the log)
5. [ ] Add recovery, overflow, authz, drain, and 5k-connection soak coverage. (actual private/world
   recovery, overflow semantics, authz, and drain complete; 5k soak and application close-code
   adapter remain)
6. [ ] Update canonical docs and run full verification.
7. [ ] Record independent review before archival.
