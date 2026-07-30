# WebSocket Transport

The transport foundation defines a strict version-1 outbound envelope and the bounded policies for
realtime delivery. HTTP remains the only inbound intent path; WebSocket clients never publish game
commands or results.

Phase-0 limits live in [`balance/transport/phase0.json`](../balance/transport/phase0.json): world
snapshots at 4 Hz, feed history 50, player history 512 messages for 10 minutes, player queues at 256
messages/1 MiB, individual messages at 64 KiB, 16 subscriptions per connection, three connections
per account, and a 15-second drain timeout. Origins are an explicit HTTP(S) allowlist. Go and JSON
Schema reject unknown fields and unsafe/out-of-range policy values.

Outbound envelopes carry version, channel, kind, authoritative revision, constants hash, server
timestamp, and payload. Known kinds are receipt, snapshot, event, presence, and system. The
TypeScript decoder ignores unknown kinds for forward compatibility, but exact-key validates every
known envelope and owned payload. Production receipts pass through unchanged because their closed
schema belongs to the Production Engine.

Go publication and TypeScript decoding enforce the same shared wire corpus. Snapshot and event
payload revisions must equal the envelope revision, snapshot scope must match its channel family,
and nested state/event payloads must be objects. A closed channel-kind matrix permits receipts only
on private player channels; `world` accepts gauges/presence/system messages and `feed` accepts
curated events/presence/system messages. Consequently a per-click receipt cannot reach a public
channel through the generic publisher. Unknown future kinds remain client-ignored before payload
interpretation as the version-1 forward-compatibility rule requires.

The live Centrifuge writer is wrapped by an application-owned queue discipline. Its byte queue is
the first bound. A separate per-connection counter enforces the declared private-player message
bound; receipts remain ordered and lossless until either bound is reached, then the socket closes
with code 4000. Reservations are keyed by authoritative revision, so command replies and rev-0
drain frames cannot decrement receipt capacity they never reserved. `world` is a gauge: each flushed revision marks the newest value for every current
subscriber, and the transport-write hook discards older queued snapshots while preserving any
frame already in flight. The hook decodes both Centrifuge JSON and protobuf publication framing
before applying either discipline; malformed publication metadata fails closed. Centrifuge's own history is the sole history implementation and provides
monotonic per-channel offsets. Player recovery fails explicitly outside its count/TTL window,
causing the one full-state resync path; world recovery returns only the latest snapshot.

Channel authorization is closed: world/feed are public to authenticated sessions, a player channel
must match the Founder identity, and guild/cohort/match channels delegate to server-side membership
lookups. Identity never grants membership through token claims.

The server embeds Centrifuge behind an origin-checked WebSocket handler. Its connect command consumes
the account service's access token, uses the opaque account ID for the per-account connection limit,
and retains only the Founder ID as connection information. Subscription authorization is evaluated
server-side for every channel; client publication is always rejected. Three concurrent connections
are permitted per account and the oldest is closed with code 4002 when a fourth connects. Access
tokens carry their existing expiry into the socket and expire with code 4001; periodic liveness
checks also detect server-side revocation.

Publishing `world` state stores only the greatest pending revision and a node-owned ticker emits at
the configured rate, so call sites cannot accidentally turn player actions into public per-click
traffic. Centrifuge history is configured at publication time: 512/10 minutes for player streams,
50 for feed and social streams, and one latest snapshot for world and match state. The node's drain
operation publishes the courtesy `server_restarting` system envelope to active channels, closes
clients with code 4003, and shuts down under the caller's context.

Every new Company intent record inserts its exact normalized receipt into
`transport_receipt_outbox` in the same database transaction as the rejection or new save revision;
an Exit inserts against the new run's authoritative revision. Idempotent replay does not create a
second row. Receipts above 60 KiB are rejected at both the application and database boundaries.
The application insert measures PostgreSQL's exact `jsonb::text` representation before mutation,
so structural spacing cannot pass Go and then abort the surrounding intent on the database CHECK.
This leaves room for the closed 64-KiB transport envelope. Relay workers claim at most the oldest
pending row per Founder with expiring leases and `SKIP LOCKED`, sort the returned batch by outbox
identity, publish the receipt unchanged to `player:{founder_id}`, then acknowledge it. On publish or
acknowledgement failure the failed row records an attempt and every unprocessed claim is released;
newer rows for that Founder remain ineligible until the head is published or dead-lettered. Five
deterministic failures dead-letter the row and emit a `receipt_dead_letter` invariant report, so a
poison receipt cannot pin readiness forever. A crash after publication but before acknowledgement
may redeliver the same receipt, which is safe because intent identity and revision reconciliation
are already idempotent; a crash before publication cannot lose it.

The embedded node sets Centrifuge's process-wide slow-writer policy once, before any node runs, so
its bounded byte queue closes stalled clients with application code 4000 rather than the library's
generic code. Channel namespace metrics use only the bounded labels `world`, `feed`, `player`,
`guild`, `cohort`, `match`, and `other`; enabling that classifier also makes Centrifuge retain the
channel metadata required by the transport-write guard. The acceptance soak holds 5,000
authenticated in-memory WebSocket connections on one node at 10 Hz. Every subscriber must observe
a strictly increasing subsequence ending at the final world revision; skipped intermediate gauges
are valid under drop-stale, while any wrong channel/kind, duplicate/regressing revision, missing
final state, or click-shaped publication fails the soak.
Drain courtesy messages deliberately bypass recoverable history because their revision is zero.
The gameserver broadcasts first, then closes intent admission; every exit branch—including a failed
broadcast—closes sockets and shuts down under the same bounded context.

The runnable `cmd/gameserver` wiring and event/snapshot relays remain implementing; this document
does not claim those paths exist yet.

The in-process gameserver lifecycle now owns `/healthz`, `/readyz`, WebSocket mounting, and exact
intent admission during shutdown. Draining irreversibly marks readiness false before publishing
the courtesy message, then closes admission immediately after that broadcast. A relay tick cannot
raise readiness during the broadcast window. It then waits for every already-admitted HTTP intent,
flushes the receipt outbox to empty, stops the relay, closes sockets with code 4003, and shuts down
Centrifuge under the configured 15-second bound. New intents during that interval receive HTTP 503
with `server_draining/retry_same_intent_id`; health remains process-liveness while readiness also
checks Postgres.
