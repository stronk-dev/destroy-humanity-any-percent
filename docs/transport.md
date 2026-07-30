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

Connection queues implement the two distinct loss rules: `world` is a gauge and replaces its stale
queued value, while private receipts remain ordered and lossless until the declared bound; overflow
returns the typed queue-overflow condition so the socket layer closes with code 4000. History uses
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
second row. Relay workers claim ordered batches with expiring leases and `SKIP LOCKED`, publish the
receipt unchanged to `player:{founder_id}`, then acknowledge it. Failed publication releases its
claim immediately. A crash after publication but before acknowledgement may redeliver the same
receipt, which is safe because intent identity and revision reconciliation are already idempotent;
a crash before publication cannot lose it.

The composed gameserver/in-flight transaction gate, event and snapshot relays, typed 4000 mapping
from Centrifuge's internal byte queue, and the 5,000-connection soak remain implementing; this
document does not claim those paths exist yet.
