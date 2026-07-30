# WebSocket Transport & Fan-out — append-only log

## 2026-07-29 — start

- Accepted through `planning/codex-batch-2026-07-29.md` after T1–T6 closed identity, inbound intent,
  wire, recovery, literal-limit, and runnable-lifecycle gaps.
- Verified the current official Go module releases before pinning: embedded Centrifuge v0.38.0 and
  coder/websocket v1.8.15. Centrifuge remains pre-v1, so its minor version is treated as an API
  boundary and pinned exactly.
- Implementation begins with strict data/config and recovery behavior before the HTTP/WebSocket
  composition, keeping protocol semantics testable without sockets.

## 2026-07-29 — policy, wire, recovery core

- Added the exact T5 Phase-0 literals as strict declarative data, validated in Go and JSON Schema.
  Origins must be distinct HTTP(S) authorities without path/query/fragment; every bound is positive
  and constrained to the accepted operational range.
- Added the v1 Go encoder and strict TypeScript decoder. Unknown top-level kinds are ignored before
  payload interpretation; known envelopes and transport-owned payloads reject unknown fields.
  Production receipts remain the exact C1 object rather than acquiring a transport copy schema.
- Implemented drop-stale world queues and bounded lossless private queues. Overflow is an explicit
  error mapped forward to close code 4000; no receipt-dropping method exists.
- Implemented channel-offset history with player count/TTL recovery and latest-only world behavior,
  plus closed server-side player/guild/cohort/match authorization. Tests demonstrate a truncated
  player history is unrecoverable, 300 queued world snapshots resume as exactly the newest one, and
  a non-participant cannot subscribe.
- Strict TS/Svelte checks, 6,426 client tests, schema verification, and the focused Go transport
  suite are green.

## 2026-07-29 — embedded node and actual socket path

- Embedded Centrifuge v0.38.0 behind its actual WebSocket handler and promoted both Centrifuge and
  coder/websocket from indirect pins to direct dependencies. The handler enforces the declarative
  origin allowlist and 64 KiB inbound cap.
- Connect authentication delegates to the Account repository's access-token authority. Centrifuge
  receives the account ID for connection limiting and only the Founder ID as presence-visible
  connection info. Subscription callbacks apply the closed channel authorization rules and client
  publish callbacks always reject.
- The node preserves the accepted typed lifecycle: access expiry/revocation closes with 4001, a
  fourth connection replaces the oldest with 4002, and drain publishes the courtesy system message
  before closing with 4003. Periodic token revalidation detects a revoked token before its JWT TTL.
- Public world publishing is now structurally coalesced: callers only replace a pending greatest
  revision and the node-owned ticker emits at the configured 4 Hz. Private, feed/social, world, and
  match publications attach their respective Centrifuge history contracts.
- Added actual coder/websocket tests covering allowed/denied origins, token authentication, private
  channel authorization, rejection of client publishing, and twenty world updates collapsing into
  one latest publication. A self-review also found and fixed a byte-cap bypass when replacing an
  already queued world snapshot.
- Remaining acceptance work is explicit: receipt/event mapping from committed intents, composed
  gameserver plus in-flight drain gate, mapping Centrifuge's internal slow-consumer close to typed
  code 4000, actual recovery/drain integration, and the 5k-connection soak.
