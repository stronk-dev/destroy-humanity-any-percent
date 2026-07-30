# Guilds

Guilds are account-scoped homes; changing Founder does not change membership. The server stores
guild identity, append-only membership periods, roles, applications/invitations, account-level
intent receipts, guild events, Health inputs, and durable presence publications in Postgres.

The active-membership invariants are database-enforced: an account belongs to at most one guild,
a guild has at most one leader, names are case-insensitively unique among active guilds, and joins
serialize on the guild row before applying the 50-member cap. Leaving closes a membership period;
rejoining inserts another row. A membership identity cannot be rewritten or deleted.

## Lifecycle surface

Authenticated account requests use `POST /api/v1/guild/intents`. The exact-schema closed kinds are
`create_guild`, `join_guild`, `apply_guild`, `admit_member`, `invite_member`, `accept_invite`,
`leave_guild`, `set_role`, and `disband_guild`. Each carries a UUIDv7 `intent_id` and positive
`expected_revision`. Receipts and typed rejections are persisted by `(account_id,intent_id)`;
retries return byte-identical JSON and a changed payload under the same id is rejected.

Open joins, application admission, and invitation acceptance all take the same row lock and cap
check. Officers may admit and invite. Only the leader changes roles; transferring leadership
demotes the old leader to officer in the same transaction. Manual disbanding requires the leader
to be the sole member. A guild below two members records the first deficient instant and a sweep
disbands it after seven complete days. Membership changes enter a claimed outbox and publish strict
`presence` envelopes on `guild:{guild_id}`; authorization queries the active account row directly.

Production composition remains fail-closed until the active RFC's owner gaps are resolved: the
repository has no concrete name-moderation validator; account deletion has no declared guild
leadership/anonymization policy; and the tithe/Health contracts do not define the Decimal-to-XP or
XP-to-Health normalization. The HTTP, persistence, exchange arithmetic, resolver, and relay seams
exist without silently choosing those mechanics.

## Balance and deterministic exchange

[`balance/guilds/phase0.json`](../balance/guilds/phase0.json) is part of Balance Epoch 3:

- tithe: 20,000 ppm;
- member clearing: 500,000 ppm; NPC clearing: 250,000 ppm;
- per-consumer boundary intake: 120 units;
- stock-consumption production bonus: 0 ppm per unit;
- membership limits: 2–50, with a seven-day below-floor grace period.

The pure clearing kernel sorts producers and consumers by raw `account_id`. A producer offers
`floor(stock × rate / 1,000,000)`. The offer is divided into integer base and remainder shares;
each consumer receives at most its share, remaining boundary intake, and persisted stock headroom.
Unused capacity is not redistributed in that boundary and stays with the producer. Missing-faction
links therefore remain inert. The solo path runs the same cap rules against one labeled NPC
counterparty at the lower rate and credits the founder's consumed-stock counter.
