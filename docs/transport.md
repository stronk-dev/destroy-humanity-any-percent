# WebSocket Transport

The transport foundation defines a strict version-2 outbound envelope and the bounded policies for
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
interpretation as the version-2 forward-compatibility rule requires.

The live Centrifuge writer is wrapped by an application-owned queue discipline. Its byte queue is
the first bound. A separate per-connection counter enforces the declared private-player message
bound; receipts remain ordered and lossless until either bound is reached, then the socket closes
with code 4000. Invalid publication framing closes with code 4004 instead of being mislabeled as
queue pressure. Reservations reject invalid revisions separately from overflow and are keyed by
authoritative revision, so command replies and rev-0 drain frames cannot decrement receipt
capacity they never reserved. `world` is a gauge: each flushed revision marks the newest value for every current
subscriber, and the transport-write hook discards older queued snapshots while preserving any
frame already in flight. The hook decodes both Centrifuge JSON and protobuf publication framing
before applying either discipline; malformed publication metadata fails closed with its typed
invalid-frame diagnosis. Centrifuge's own history is the sole history implementation and provides
monotonic per-channel offsets. Player recovery fails explicitly outside its count/TTL window,
causing the one full-state resync path; world recovery returns only the latest snapshot.

The production browser now owns one recovery controller. It persists each subscribed channel's
Centrifuge epoch/offset under the active Founder, reconnects after ordinary drops, and sends those
positions with `recover: true`. Recovered publications pass through the same strict decoder and
per-scope `PlayerRevisionCursor` as live publications. Duplicate event IDs at the current revision
are suppressed; historical compensation is emitted to the runtime consumer as structured audit
output without moving a cursor (the current Game UI has no owner-authored presentation for it); a
forward gap, an unrecoverable stream, `resync_required`, queue overflow (4000), or invalid frame
(4004) clears both positions and performs the existing authenticated `GET /api/v1/founder/state`
sync before a live resubscribe. The same snapshot resets Company and Founder cursors. A drain courtesy publication
has no stream offset, suppresses reconnect through `resume_after_ms`, and then uses that ordinary
recovery path. Auth expiry (4001) and replacement (4002) stop visibly: token rotation and
replacement arbitration belong to their owning successors and are not improvised by Transport.
Unknown future envelope kinds are not interpreted but their valid stream position is retained, so
forward compatibility cannot create an infinite replay loop.

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

World snapshot state is a closed version-1 integer schema: a world-owned monotonic revision;
planet depletion/Health; Commons server Health, active Founders, and members; online/total Founder
population; nullable milestone identity plus progress; and epoch identity/name. The envelope and
state revisions are identical. Go encoding and TypeScript decoding share the same corpus, ppm
fields are bounded to `[0,1,000,000]`, counts are safe integers, and an absent milestone requires
zero progress. Planet and milestone values are deliberately zero/null until those owners ship.
The gameserver aggregator samples at the catalog's 4 Hz cadence and advances its revision only
after the snapshot is accepted by the publisher.

Every Founder/Company event and every Company or Founder intent receipt enters one durable
`transport_player_outbox` in the transaction that commits it. A database trigger owns event
insertion for every event writer; intent and Exit transactions insert their exact normalized
receipts. An Exit therefore produces one transaction-ordered Founder sequence spanning its Founder
event, final-run Company events, next-run Company event, and receipt. Event payloads carry their
required `company|founder` scope, the revision in that scope, and a closed cursor effect. Ordinary
events are `advance`; `compensation` alone is `historical`. Historical compensation remains visible
audit output but does not advance, rewind, or gap-check the client cursor. Forward events may share
one revision, so the per-scope client cursor dedupes by event ID within its current revision. Those
per-scope revisions—not cross-scope arrival order—are the reconciliation authority when independent
transactions interleave.
Idempotent replay does not create a
second row. Receipts above 60 KiB are rejected at both the application and database boundaries.
The application insert measures PostgreSQL's exact `jsonb::text` representation before mutation,
so structural spacing cannot pass Go and then abort the surrounding intent on the database CHECK.
Authoritative events always commit, even when the resulting envelope is oversized. Such a row
reaches the relay, fails the 64-KiB envelope policy deterministically, dead-letters after five
attempts, and emits a small `resync_required` frame that invokes the ordinary full-sync path.
Transport can delay
event presentation; it cannot roll back game or projection history. Relay workers claim at most the oldest
pending row per Founder with expiring leases and `SKIP LOCKED`, sort the returned batch by outbox
identity, publish its event or receipt unchanged to `player:{founder_id}`, then acknowledge it. On publish or
acknowledgement failure every unprocessed claim is released. Deterministic envelope-policy failures
consume the failed head's attempt budget; five dead-letter the row and emit a
`player_message_dead_letter` invariant report. Publisher and acknowledgement infrastructure failures do
not consume that budget: the head retains its lease for a one-second backoff, records the last
error, and blocks newer rows for that Founder without blocking other Founders. A poison receipt
therefore cannot pin readiness forever, while a short outage cannot destroy a valid receipt. A
crash after publication but before acknowledgement may redeliver a message, which is safe because
events have immutable event IDs and receipts retain intent identity plus authoritative revision; a
crash before publication cannot lose it.

The outbox migration is a single-process, migrate-before-readiness operation. It intentionally
does not replay historical events that predate installation; live delivery begins at installation
and full sync remains authoritative. Applied migration files are immutable—later corrections use
new forward migrations. A dead-lettered event is not skipped invisibly: the relay explicitly
requests full sync even when a same-revision receipt would otherwise conceal the missing event.

The embedded node sets Centrifuge's process-wide slow-writer policy once, before any node runs, so
its bounded byte queue closes stalled clients with application code 4000 rather than the library's
generic code. Channel namespace metrics use only the bounded labels `world`, `feed`, `player`,
`guild`, `cohort`, `match`, and `other`; enabling that classifier also makes Centrifuge retain the
channel metadata required by the transport-write guard. The acceptance soak holds 5,000
authenticated in-memory WebSocket connections on one node at 10 Hz. Every subscriber must observe
a strictly increasing subsequence ending at the final world revision; skipped intermediate gauges
are valid under drop-stale, while any wrong channel/kind, duplicate/regressing revision, missing
final state, or click-shaped publication fails the soak.

The active RFC's literal ten-second connected-stall criterion is not yet met. A cold 11.29-second
actual-socket probe on 2026-08-21 filled Centrifuge's byte queue and closed the world-only subscriber
before resume. The current transport-write hook can discard all queued stale frames except the
already in-flight frame, but it runs after queue admission and therefore cannot enforce an exact
one-frame bound for a genuinely blocked writer. The browser's typed 4000 path now recovers the user
through full sync, but that does not redefine AC3; the RFC author must either provide a pre-queue
coalescing authority or reconcile the criterion to the implementable in-flight-plus-newest
contract.
`world_rev` is a process-lifetime ordering key, not persisted world history. It may restart when the
gameserver restarts, so reconnecting clients treat the recovered latest snapshot as a new baseline
and never compare its revision with the prior connection.
Drain courtesy messages deliberately bypass recoverable history because their revision is zero.
The gameserver broadcasts first, then closes intent admission; every exit branch—including a failed
broadcast—closes sockets and shuts down under the same bounded context.

The runnable `cmd/gameserver` composes the event/receipt and Guild-presence relays, the 4 Hz world
snapshot driver, Postgres-backed Guild and Commons membership resolvers, and a deny-closed Match
resolver. Its real-socket integration proof uses the same authenticated handler mounted in the
production binary.

The in-process gameserver lifecycle now owns `/healthz`, `/readyz`, WebSocket mounting, and exact
intent admission during shutdown. Draining irreversibly marks readiness false before publishing
the courtesy message, then closes admission immediately after that broadcast. A relay tick cannot
raise readiness during the broadcast window. It then waits for every already-admitted HTTP intent,
flushes the player-message outbox to empty, stops the relay, closes sockets with code 4003, and
shuts down Centrifuge under the configured 15-second bound. New intents during that interval receive HTTP 503
with `server_draining/retry_same_intent_id`; health remains process-liveness while readiness also
checks Postgres.
