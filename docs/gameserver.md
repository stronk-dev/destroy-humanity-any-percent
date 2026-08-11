# Gameserver Composition

`server/cmd/gameserver` is the runnable Phase-0 process. It assembles the repository's reviewed
packages rather than owning gameplay rules: epoch and artifact loading, save/Production services,
accounts, Commons/Route/Guild projections, replay verification and leaderboard projection,
realtime relays, guild maintenance, world snapshots, and bounded shutdown.

## Startup and configuration

The binary requires:

- `DATABASE_URL`, a Postgres connection string;
- `CLOUD_CLICKER_SERVER_ID`, a canonical UUID identifying the Commons server shard;
- `CLOUD_CLICKER_JWT_KEY`, at least 32 key bytes encoded with standard base64;
- `CLOUD_CLICKER_BOOTSTRAP_KEY_ID`, the active bootstrap-receipt encryption key ID;
- `CLOUD_CLICKER_BOOTSTRAP_KEY`, exactly 32 AES-256-GCM key bytes encoded with standard base64.

`CLOUD_CLICKER_REPOSITORY_ROOT` defaults to the working directory,
`CLOUD_CLICKER_ACTIVITY_BRACKET` defaults to `activity.standard`, and `LISTEN_ADDR` defaults to
`:8080`. Build it with `make build-gameserver`.

Startup migrates Postgres, loads the repository epoch declaration, reconciles it before realtime
or readiness starts, and reconstructs every executable catalog bundle from immutable database
artifact bytes. Before readiness rises, every attached background worker completes one successful
prime pass, the world publishes its first snapshot, and the durable player relay completes one
flush. A failed prime aborts startup; a later worker failure or unexpected clean exit lowers
readiness and is returned by the process so its supervisor can restart it. Composition fails
closed if the current hash, artifact set, Commons/Economy boundary, moderation data, server
identity, or any owned service constructor disagrees.

The HTTP process exposes `/healthz`, `/readyz`, the account/Founder/intent API under `/api/v1`, and
the Centrifuge endpoint at `/connection/websocket`. Health means the process is alive. Readiness
also requires Postgres, a matching epoch, a running realtime node, and healthy background jobs.

## Runtime ownership

The composed Production service uses the real Guild `PendingSettlements` implementation. A
clearing driver reads active members' authoritative Company saves, commits content-addressed Guild
boundaries bound to each member's exact Founder, Company stream, and run, and the next player
intent applies that identity's exact slice through the replay-input contract. Later snapshots net
out committed-but-unapplied reservations, so scheduler ticks cannot allocate offline stock twice.
The driver pages through every active Guild, retries snapshots invalidated by ordinary
join/leave races, and restores historical saves with their revision timestamp and pinned Economy
and Faction catalogs. All active members of one boundary must resolve the same pinned stock cap;
a future cap-changing epoch is an explicit design gap and fails closed until its migration policy
is specified. No empty production resolver or duplicate clearing math exists in the binary.

Channel authorization resolves active Guild membership and projected Commons cohort membership
from Postgres. `match:*` remains deliberately deny-closed until its owner engine lands. The durable
player relay publishes transaction-owned events and receipts; the Guild presence relay publishes
claimed membership changes. The verification worker drives the shared replay verifier and
leaderboard projector. Guild clearing, below-floor sweeping, expired-session collection,
bootstrap-receipt tombstoning, and all relay/queue work run as owned jobs and are stopped before
socket drain. Session and bootstrap-receipt expiry share one ordered credential-GC job: a
transient session-prune failure prevents the receipt pass and fails the job instead of reporting a
partial success.

The world aggregator samples existing projection tables at the transport catalog's 4 Hz cadence.
It owns `world_rev`, publishes only closed integer version-1 snapshots, and increments the revision
only after publication succeeds. The counter is monotonic only for one gameserver process
lifetime; it restarts with the process, and clients must not compare world revisions across a
reconnect. Planet and milestone fields remain their declared zero values until those systems ship.

## Shutdown and proof

SIGINT and SIGTERM begin the transport drain contract: readiness falls, active channels receive
the restart courtesy frame, and new intents are rejected before shutdown waits on background jobs.
Admitted intents finish, jobs stop, the player outbox flushes, sockets close, and both Centrifuge
and HTTP stop under the catalog's 15-second bound. A stuck job therefore cannot consume the drain
deadline while admission remains open.

The real-Postgres composition integration test boots this exact graph, creates an anonymous
account and session, submits authoritative play, receives receipt and event envelopes over an
authenticated WebSocket, observes a world snapshot, authorizes Guild but denies Match channels,
authorizes both Guild and Commons cohort channels while Match remains denied, commits and consumes
three reserved clearing boundaries without over-allocation, proves a New Founder cannot inherit an
old settlement, restores a version-1 member at its authoritative revision timestamp,
and drains to non-readiness. A separate startup-barrier fixture proves the attached clearing and
credential-GC jobs themselves perform their prime passes, including destruction of expired
bootstrap ciphertext into its permanent tombstone; the socket/settlement fixture independently
proves pagination beyond the first Guild. `make test-save-integration` runs these inside the
repository's standard Postgres test topology.
The same composed graph also runs a progressed fixture through authoritative play and Exit, the
background shared-kernel verifier, immutable archive, and `any_percent` board projection. The
fixture seeds only the otherwise-unshipped T0→T2 content progression; every transition from the
first logged command through the board is the production path.
