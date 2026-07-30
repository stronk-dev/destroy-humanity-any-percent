# RFC: Guild Model

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-07-30
- **Design refs:** `design/05 §3` (formation 2–50, automatic tithe, contribution windows, faction interdependence exchange, Break Room seats, small-guild mercy keyed on cohort, NPC fallback), `design/04 §4` (no scarcity economy — satire target)
- **Depends on:** Account & Session Bootstrap (implementing — account identity), Commons Compact (implemented — guild Health term seam), Faction & Incorporation (draft — exchange cycle), Transport (implementing — `guild:{id}` channels + membership resolver)
- **Unblocks:** Commons Onboarding blockers #1/#5, transport guild resolver (replaces the deny-closed stub), Break Room / guild-events successor RFCs
- **Planning:** `planning/guild-model/` (once implementing)

## Summary

The second missing owner contract: guilds as data — formation, membership lifecycle, the automatic
tithe, the Commons Health input, and the transport membership seam. Phase 0 is deliberately the
*structural* guild (identity, membership, tithe, Health, exchange); guild events, bosses, Break
Room seats, and matchmaking grades are successor RFCs on these tables.

## Specification

### G1 — Tables (all mutations evented, projections idempotent by event_id)

- `guilds(guild_id uuidv7 PK, name, created_at, founder_account uuid, join_policy 'open'|'invite'|'apply', disbanded_at nullable)` — name: 3–24 chars from the existing name-moderation charset; uniqueness case-insensitive among non-disbanded.
- `guild_members(guild_id, account_id, joined_at, left_at nullable, role 'leader'|'officer'|'member')` — **membership history is append-only**: leaving sets `left_at`, rejoining inserts a new row. Partial unique: one active membership per account (an account is in ≤1 guild — the design's guilds are homes, not tags); exactly one active `leader` per guild (partial unique index).
- **Membership is account-scoped, not founder-scoped** (guilds survive `New Founder`; the transport authz already keys `guild:*` on AccountID).
- Size: active members ≤ 50, enforced in the join transaction (`FOR UPDATE` on the guild row); floor of 2 within 7 days is a lifecycle rule for the disband sweep, not a hard insert constraint.

### G2 — Lifecycle intents (closed set)

`create_guild {name, join_policy}` · `join_guild {guild_id}` (open) / `apply_guild` + `admit_member` (apply) / `invite_member` + `accept_invite` (invite) · `leave_guild {}` · `set_role {account, role}` (leader; leadership transfer = `set_role leader`, demoting self atomically) · `disband_guild {}` (leader, requires sole remaining active member). All are account-level intents on a new `guild` intent surface following C1 exactly (idempotency by intent_id, typed rejections: `guild_full`, `already_member`, `not_leader`, `name_taken`, `not_applicant`…). Events: `guild_created/member_joined/member_left/role_changed/guild_disbanded`, feed-eligible.

### G3 — The automatic tithe and the Health term

- Tithe is **structural, not a transfer**: a catalog band `guild_tithe_ppm` (Phase-0 literal, joins the commons catalog artifact) of each member's production accrues to `guild_xp` (exact integer, guild row) at accrual evaluation — computed server-side from the member's committed production deltas, never client-reported. Members see percentile-within-faction, not raw numbers (a projection, successor UI).
- **The guild Health term** (the 0.5·guild slot in EffectiveHealthPPM): the guild's contribution snapshot is `guild_health_inputs(guild_id, window_start, active_founders, tithed_xp)` written by the same projection cadence the cohort term already uses; population-normalized exactly like the cohort formula — **small-guild mercy may change access/progress, never this formula** (the Commons Compact law, restated as this table's contract). Guildless founders keep the shipped cohort substitution, unchanged.
- Guild membership sets the run's `commons` variable only via the existing compact rules (guild ≠ compact; a guild of non-signatories contributes a Health term of zero and that is legal).

### G4 — Faction exchange (activates F2's inert stock)

Within a guild, the interdependence cycle clears **automatically at accrual boundaries**: each member's `produces` stock offers to guildmates whose faction `consumes` it, pro-rata, at a catalog-declared clearing rate — no trading UI, no orders, no prices (the design's "exchange is automatic"). All clearing is evented per member (`exchange_cleared`) and deterministic given the member set and stocks (ordered by account_id — no RNG). The guildless public exchange stays out (successor RFC).

### G5 — Transport resolver (replaces the deny-closed stub)

`GuildMemberships` implements the transport `Memberships` interface from `guild_members` active rows (account-scoped, matching the existing authz key). Membership changes publish `presence` to `guild:{id}`. The resolver is the ONLY consumer-facing query surface Phase 0 ships; guild pages/panels are Commons-Onboarding + successor client work.

### G6 — NPC fallback (declared, minimal)

A solo founder's "partner network" is a **virtual guild**: the exchange clears against catalog-declared NPC counterparties at a reduced rate (`npc_exchange_ppm < guild clearing rate`, catalog). No guild row, no membership — a pure function in the accrual path, clearly labeled in receipts. This satisfies the design law (all guild content completable solo, slower) for the only guild mechanic Phase 0 ships.

## Executable contracts (answering the 2026-07-30 bounce)

### GA — Parent resolved

Faction FA/FB/FC (amended 2026-07-30) declare the stocks (int64 Company fields), the accrual
formula, and the literal catalog. G4's ledger input exists.

### GB — The complete catalog object

`balance/guilds/phase0.json` (strict loader, joins the epoch seed — mint):
```json
{"schema_version": 1, "guild_tithe_ppm": 20000, "clearing_rate_ppm": 500000,
 "npc_exchange_ppm": 250000, "stock_intake_cap": 120, "consumption_bonus_ppm_per_unit": 0,
 "max_members": 50, "min_members": 2, "grace_days": 7}
```
`consumption_bonus_ppm_per_unit` is the modifier hook consumed units feed (a `stock_consumption`
slot in the production stack's registry): **literally 0 at Phase 0** — the flow is real and
observable, the buff is balance work. All provisional; epoch protocol applies.

### GC — Deterministic clearing (total specification)

At each accrual boundary, per guild, in one transaction:
1. **Producers** = active members with `stock_units > 0`, ascending `account_id`. For each:
   `offered = stock_units * clearing_rate_ppm / 1_000_000` (integer division).
2. **Consumers** of resource R = active members whose faction `consumes` R, ascending
   `account_id`, each with remaining intake `cap_i = stock_intake_cap − consumed_this_boundary_i`.
3. Per producer in order: `n` = consumers with `cap_i > 0`; `base = offered / n` (integer),
   `rem = offered mod n`; consumer `i` (in account_id order) is allocated
   `min(base + (1 if index(i) < rem else 0), cap_i)`; allocation runs one pass in order —
   **units an over-cap consumer cannot take are NOT redistributed within the boundary** (they
   stay with the producer; next boundary retries). Zero consumers → nothing clears.
4. Debit `producer.stock_units`, credit `consumer.consumed_stock_units` (saturating at
   `stock_cap`), emit one `exchange_cleared {producer, allocations: [{account, units}]}` event
   per producer with any allocation. No RNG, no clocks: the golden answer is unique given
   member set + stocks.
NPC fallback (G6): a solo founder's clearing uses the same algorithm with one virtual consumer of
infinite… no — **one NPC consumer with `cap = stock_intake_cap`**, clearing at
`offered_npc = stock_units * npc_exchange_ppm / 1_000_000`, credited to nothing (units leave the
stock; the founder's `consumed_stock_units` gains `offered_npc` — the NPC network both buys and
supplies, labeled `npc: true` in the receipt).

## Acceptance criteria

1. Lifecycle intents: C1 conformance suite (idempotency, typed rejections per arm); join races at the 50 cap and leader-uniqueness proven under concurrent transactions (real Postgres).
2. Membership history append-only (trigger); rejoin creates a second row; `New Founder` leaves membership intact.
3. Tithe: production round-trip accrues `guild_xp` exactly `tithe_ppm` of committed deltas (golden fixture); no client-supplied path exists.
4. Health inputs: population-invariance property test (doubling members with identical per-capita tithe leaves the term fixed); guildless substitution untouched (regression).
5. Exchange: deterministic clearing fixture (3 factions present, 1 absent — the absent link's stock stays inert); NPC fallback clears at the reduced rate with labeled receipts; replay-idempotent.
6. Transport: guild channel subscribe allowed for active member, denied after `leave_guild` within one revision cycle; presence events observed on the real socket path.
7. Catalog additions land via epoch-seed mint (guard-proven).

## Open questions

- Disband sweep cadence + sub-2-member grace (7 days, lifecycle job) — implementation freedom within the stated rule.
- Officer permission matrix beyond admit/invite — successor RFC with guild events.
- Name moderation shares the existing rules; display surfaces are client work.

## Changelog

- 2026-07-30: created (draft) — the guild owner contract Commons Onboarding blockers #1/#5 named; unblocks the transport guild resolver.
- 2026-07-30: Codex bounce answered — GA (parent resolved), GB (complete literal catalog incl. the 0-ppm consumption hook), GC (total deterministic clearing: ordered one-pass allocation, no intra-boundary redistribution, NPC = one capped virtual counterparty).
