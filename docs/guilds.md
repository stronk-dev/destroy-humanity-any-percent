# Guilds

Guilds are account-scoped homes; changing Founder does not change membership. The server stores
guild identity, append-only membership periods, roles, applications/invitations, account-level
intent receipts, guild events, Health inputs, and durable presence publications in Postgres.

The active-membership invariants are database-enforced: an account belongs to at most one guild,
a guild has at most one leader, normalized names are unique among active guilds, and joins
serialize on the guild row before applying the 50-member cap. Leaving closes a membership period;
rejoining inserts another row. A membership identity cannot be rewritten or deleted.

## Lifecycle surface

Authenticated account requests use `POST /api/v1/guild/intents`. The exact-schema closed kinds are
`create_guild`, `join_guild`, `apply_guild`, `admit_member`, `invite_member`, `accept_invite`,
`leave_guild`, `set_role`, and `disband_guild`. Each carries a UUIDv7 `intent_id` and positive
`expected_revision`. Receipts and typed rejections are persisted by `(account_id,intent_id)`;
retries return byte-identical JSON and a changed payload under the same id is rejected.
Guild IDs are generated independently by the server; a client intent ID is never an object ID.
Names are NFKC-normalized, lowercased, whitespace-collapsed, restricted to the published ASCII
mechanical charset, and checked against the committed denylist plus deployment-only additions.

Open joins, application admission, and invitation acceptance all take the same row lock and cap
check. Officers may admit and invite. Only the leader changes roles; transferring leadership
demotes the old leader to officer in the same transaction, with one event for each role mutation.
Every authority check is repeated after the guild lock; the lock order is guild then membership.
Manual disbanding requires the leader
to be the sole member. A guild below two members records the first deficient instant and a sweep
disbands it after seven complete days. Membership changes enter a claimed outbox and publish strict
`presence` envelopes on `guild:{guild_id}`; authorization queries the active account row directly.

Account deletion closes and anonymizes membership history. A departing leader deterministically
hands leadership to the oldest officer, then oldest member; an empty guild disbands. The mutation
and account removal share one transaction.

## Balance and deterministic exchange

[`balance/guilds/phase0.json`](../balance/guilds/phase0.json) is part of Balance Epoch 4:

- tithe: 20,000 ppm;
- member clearing: 500,000 ppm; NPC clearing: 250,000 ppm;
- per-consumer boundary intake: 120 units;
- stock-consumption production bonus: 0 ppm per unit;
- membership limits: 2–50, with a seven-day below-floor grace period.
- full per-capita Guild Health target: 250,000 tier-progress XP per active founder;
- clearing interval: 300,000 ms.

Production converts the increase in the current tier's shipped progress coordinate to integer ppm.
The two-percent tithe uses integer division with a persisted per-company remainder; only active
members contribute. A company-stream event projects XP into the account's guild, saturating only
at the shared exact-integer ceiling. The trailing Commons window counts founders with at least one
evaluation, and Health is `clamp(1,000,000 × window_xp / (active × 250,000), 0, 1,000,000)`.

The pure clearing kernel sorts producers and consumers by raw `account_id`. A producer offers
`floor(stock × rate / 1,000,000)`. The offer is divided into integer base and remainder shares;
each consumer receives at most its share, remaining boundary intake, and persisted stock headroom.
Unused capacity is not redistributed in that boundary and stays with the producer. Missing-faction
links therefore remain inert. The solo path runs the same cap rules against one labeled NPC
counterparty at the lower rate and credits the founder's consumed-stock counter.

Committed boundaries are idempotent rows keyed by guild and sequence. Each company applies its own
debit/credit slice exactly once using save-v11's monotonic watermark; projection never locks more
than one company. Consumed units declare `guild.stock_consumption` in the existing faction slot;
the Epoch-4 rate is zero, so the structural path is live without a hidden output buff. The scheduler,
replay-input capture, and real-socket driver remain explicit composition work owned by the queued
Run Genesis and gameserver rounds; no alternate clearing math runs meanwhile. Composition also
waits on GD5a: the persisted watermark must include its guild identity (or clearing must receive a
global sequence) before switching guilds can be ordered safely.
