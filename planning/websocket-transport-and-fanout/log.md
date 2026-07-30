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

## 2026-07-29 — crash-safe receipt relay

- Rejected the tempting post-HTTP callback design: a crash after commit but before callback would
  permanently lose the receipt even though the RFC names the player channel as authority.
- Added migration 00015 and an intent-transaction outbox. Applied and rejected Company intents,
  including the two-stream Exit transaction, insert the exact normalized Production receipt with
  Founder identity, authoritative Company revision, constants hash, and database timestamp in the
  same transaction as the intent record and state transition. Idempotent replay does not duplicate
  the unique `(company_stream_id,intent_id)` row.
- Added ordered lease claims using `FOR UPDATE SKIP LOCKED`, explicit release on publish failure,
  claim-token-checked acknowledgement, and a transport relay that maps the stored row to the v1
  player receipt envelope without modifying the payload. Delivery is at least once; duplicate
  delivery after publish-before-ack is intentionally absorbed by existing intent/revision
  idempotency rather than weakened into at-most-once loss.
- Real Postgres tests prove ordinary and Exit intent atomicity, replay deduplication, claim expiry
  ownership, acknowledgement, and rollback at the existing injected Exit fault boundaries. Relay
  unit tests prove exact payload mapping and immediate failure release.

## 2026-07-29 — composed in-process drain lifecycle

- Added the composed server boundary around the account HTTP router, embedded WebSocket handler,
  Postgres readiness, and receipt relay. Health means process-up; readiness additionally requires a
  successful database ping, a healthy relay, and non-draining state.
- Intent admission is an explicit gate around only `POST /api/v1/intents`. Beginning drain closes
  admission and returns typed HTTP 503 responses for new work while retaining an exact count of
  already-admitted transactions. No `sync.WaitGroup.Add` races with `Wait`; the gate's zero channel
  is changed under the same mutex as admission.
- Drain order is now executable: mark not-ready/close admission -> broadcast courtesy -> await
  admitted intents -> flush the transactional receipt outbox to empty -> stop relay -> close sockets
  with 4003 -> shut down Centrifuge, all under the catalog's 15-second bound. Unit coverage blocks an
  intent mid-handler and proves sockets remain open until it commits and its relay flush completes.
- `cmd/gameserver` remains a DESIGN-GAP rather than a fake binary: production composition needs a
  concrete Founder-to-server/activity-bracket resolver and participation-weight resolver for the
  already-shipped Commons projector. The active Commons Onboarding RFC explicitly assigns those
  owners to the still-undrafted faction/incorporation and guild contracts. The generic lifecycle is
  complete, but hard-coding a deployment-wide cohort owner here would improvise that blocked model.
- The pinned Centrifuge API closes its internal byte queue with library code 3008 and exposes no
  per-node override for that disconnect. T5 requires application code 4000. Mutating Centrifuge's
  exported package-global `DisconnectSlow` would make multiple nodes/tests race and is rejected.
  This exact adapter/library mismatch remains an implementation blocker to resolve by an accepted
  dependency patch/fork decision; the separate bounded queue continues to prove the required loss
  semantics without falsely claiming the actual socket emits 4000.

## 2026-07-29 — complete-diff self-review correction

- Found a bounded-drain defect in the lifecycle diff: if an already-admitted intent outlived the
  15-second budget, the timeout branch returned before closing sockets. The ordinary ordering test
  passed because its intent completed, so the missing exceptional cleanup needed a dedicated stalled
  transaction case.
- Every failure/timeout branch now cancels the relay, closes sockets with the typed drain code, and
  invokes Centrifuge shutdown before returning the joined cause. The node also initiates Centrifuge
  shutdown when its caller context is already expired instead of returning before the library sees
  shutdown. A blocked-intent regression proves `broadcast -> close -> shutdown` still occurs at the
  deadline and the caller receives `context.DeadlineExceeded`.

## 2026-07-30 — actual recovery and sandbox-safe network tests

- Added an actual Centrifuge JSON-protocol recovery test over coder/websocket. It records the live
  stream epoch/offset, disconnects after revision 1, publishes private revisions 2 and 3 while the
  client is absent, and proves both return in order with their consecutive offsets on recovery.
- The same test disconnects a world subscriber, submits revisions 2–20 inside one coalescing
  interval, and proves recovery returns exactly one cached publication containing revision 20.
  This exercises the embedded node's real history/recovery options rather than the separate
  in-memory history model, closing Transport AC2 and AC3 at the socket boundary.
- Replaced OS-bound `httptest.NewServer` use with a shared in-memory `net.Pipe` HTTP server. The
  actual HTTP upgrade, WebSocket framing, origin checks, and account API client/server exchange all
  remain exercised, but routine Go tests no longer need permission to bind localhost. Transport and
  Account focused suites now run inside the repository sandbox with the local Go build cache.
