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

The embedded Centrifuge socket adapter, committed-receipt publisher, composed gameserver, drain
lifecycle, and soak tests remain implementing; this document does not claim those paths exist yet.
