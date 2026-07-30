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
