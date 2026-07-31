# RFC: Guild Model

- **Status:** implemented
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-07-30
- **Design refs:** `design/05 §3` (formation 2–50, automatic tithe, contribution windows, faction interdependence exchange, Break Room seats, small-guild mercy keyed on cohort, NPC fallback), `design/04 §4` (no scarcity economy — satire target)
- **Depends on:** Account & Session Bootstrap (implementing — account identity), Commons Compact (implemented — guild Health term seam), Faction & Incorporation (implemented — exchange cycle), Transport (implementing — `guild:{id}` channels + membership resolver)
- **Unblocks:** Commons Onboarding blockers #1/#5, transport guild resolver (replaces the deny-closed stub), Break Room / guild-events successor RFCs
- **Planning:** `planning/guild-model/`

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
3. Per producer in order (as amended by ruling GC-1, 2026-07-30, adopting the implemented
   kernel): `cap_i = min(intake headroom, stock_cap − received_i)` — a consumer with zero
   headroom on EITHER bound is excluded from `n`; `base = offered / n` (integer), `rem = offered
   mod n`; consumer `i` (in account_id order) is allocated `min(base + (1 if index(i) < rem else
   0), cap_i)`; one pass, **no intra-boundary redistribution** (undelivered units stay with the
   producer; next boundary retries). Zero eligible consumers → nothing clears. Nothing is ever
   destroyed at credit-saturation because saturation is an eligibility bound, not a clamp.
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

## Resolved implementation audit (2026-07-30)

The structural round proved six owner contracts were absent. They were not implementation freedom
because each changed persisted value semantics, production order, or account lifecycle; the
executable answers immediately below resolve them:

1. **GD1 — name validator:** G1 names an “existing” moderation charset, but no runtime name-policy
   owner exists. Specify normalization, accepted characters, and the injected validator contract.
2. **GD2 — tithe units:** G3 maps a percentage of arbitrary Decimal production deltas into int64
   `guild_xp` without naming the qualifying resource(s), magnitude normalization, rounding, or
   saturation/overflow behavior.
3. **GD3 — Guild Health:** G3 stores `(active_founders,tithed_xp)` but gives no formula/denominator
   mapping that pair to the required 0..1,000,000 `H_guild` input.
4. **GD4 — account deletion:** G1 covers New Founder, not deletion of a leader/member account.
   Specify deterministic succession/disband and how append-only history is anonymized while the
   account row is physically removed.
5. **GD5 — clearing activation:** GC defines the arithmetic but not the accrual-boundary identity,
   mixed-run-epoch catalog authority, or atomic multi-company-save mutation boundary. Those choices
   determine replay and lock ordering.
6. **GD6 — multiplier order:** GB names a `stock_consumption` production slot but does not place it
   in the canonical rounding-sensitive slot order or declare its economy source/provider row.

Composition remained fail-closed until the answers below were implemented. No placeholder math ran
in production during that interval.

## Executable contracts GD1–GD6 (owner answers, 2026-07-30)

### GD1 — Name validator

Normalize: NFKC → lowercase → trim → collapse internal whitespace runs to one space. Accept:
`[a-z0-9 _-]`, length 3–24 after normalization, no leading/trailing space/`-`/`_`, at least one
`[a-z0-9]`. Uniqueness compares the NORMALIZED form among non-disbanded guilds. Denylist: injected
`NameValidator` interface (fail-closed: nil validator ⇒ `create_guild` rejects `name_policy`);
baseline list committed at `moderation/guild-names.txt`; **matching (GD1-1, 2026-07-31) runs
against both the normalized form and the normalized form with `[ _-]` stripped** — separator
padding does not evade. Denylist entries validate under a laxer rule (≥2 chars, `[a-z0-9]` only).
Deployment may extend, never shrink below the committed file. Rejection category: `name_policy`.

### GD2 — Tithe units (exact, cross-tier fair)

Guild XP tithes **tier-local progress, not raw Decimals**: at each accrual evaluation of a member
company, `progress_delta_ppm` = the increase in the SHIPPED tier-local resource_log progress
coordinate (already integer ppm, already cross-runtime exact); `xp_delta = progress_delta_ppm *
guild_tithe_ppm / 1_000_000` (integer division, remainder carried per-company in
`guild_tithe_carry_ppm`, new company field). `guild_xp` is int64, **accrual-only saturating at
MaxExactInteger**. No Decimal ever crosses the boundary; a tier-N whale and a tier-1 newcomer
contribute comparable progress units by construction. Negative progress (cap-lowering migrations)
contributes zero — XP never decreases from production.

### GD3 — Guild Health

`H_guild_ppm = clamp(1_000_000 * window_xp / (active_founders * guild_xp_target_per_founder), 0,
1_000_000)` — per-capita, population-invariant by construction. `window_xp` = guild_xp accrued in
the trailing commons health window (the same window the cohort term uses);
`guild_xp_target_per_founder = 250_000` (new guilds-catalog literal, provisional, epoch protocol).
`active_founders` = accounts with an active membership AND ≥1 accrual EVALUATION in the window
(the cohort activity rule, reused) — **ruling GD3-1 (2026-07-31): evaluation, not XP production;
the projector records activity on every member evaluation regardless of progress** (the narrower
implemented definition inflated H_guild by excluding parked members from the denominator). Zero active founders ⇒ term contributes 0 (the compact's
guildless substitution never applies to members — a dead guild is a real signal).

### GD4 — Account deletion

In the account-deletion transaction (A-D5a's), for each ACTIVE membership: stamp `left_at`,
`account_id` → NULL (FK ON DELETE SET NULL — history rows survive anonymized; the append-only
trigger admits exactly this transition). If the deleted account held `leader`: succession in the
same transaction — oldest active `officer` by `(joined_at, row id)`, else oldest active `member`,
promoted with `role_changed`; no remaining active members ⇒ `guild_disbanded`. All evented,
deterministic, single transaction with the deletion.

### GD5 — Clearing activation (the commons-projection precedent, replay-safe via RA)

Clearing is **projection-side, never a multi-company transaction**: a boundary job (composed
gameserver, claim/outbox discipline) runs per guild every `clearing_interval_ms` (new catalog
literal, 300_000) computing GC's allocations from members' latest COMMITTED stock snapshots,
writing idempotent `guild_clearing_results(guild_id, boundary_seq, allocations jsonb)` rows +
`exchange_cleared` events. Each member company then applies ITS OWN slice at its next accrual
evaluation — debit and credit both — reading the projection exactly as `commons_modifier` is read
today, watermarked by `guild_boundary_seq` (new company field, monotonic). Double-spend is
impossible: stocks have no other consumer, and a company applies each boundary exactly once by
watermark. **Catalog authority: the CURRENT epoch's guilds artifact** (guilds are server-shared,
not run-pinned); replay stays exact because applied slices are server-resolved inputs and land in
`replay_inputs` (RA) — a replayer never recomputes clearing. Lock ordering: none needed — every
company transaction locks only itself; the projection is the sole coordination point.

### GD6 — Multiplier slot

Slot key `guild.stock_consumption`, Company scope, multiplicative, value `1 + consumed_this_window
* consumption_bonus_ppm_per_unit / 1e6` (ppm arithmetic, catalog rate currently 0). Provider: the
faction accrual hook (it already owns `consumed_stock_units`). Ordering: the slot enters the
SHIPPED deterministic-aggregation order (contributions sorted canonically by slot key — the
existing kernel rule); no bespoke position exists or is needed. The provider row is registered in
the production stack's slot registry by this contract; a second provider for the key is a loader
error.

### DESIGN-GAP GD5a — watermark identity (implementation re-review, 2026-07-31)

`guild_boundary_seq` is scoped by `guild_id`, but the v11 company field stores only the sequence.
After moving from a mature guild at boundary 10,000 to a new guild at boundary 5, the monotonic
watermark would reject every new-guild slice forever. The owner must choose one executable shape:
persist `(guild_id,boundary_seq)` in the company save (reset-to-current on membership change), or
make clearing sequence globally monotonic. The shipped resolver is intentionally uncomposed until
that identity is specified; the clearing writer/results and pure application kernel remain usable
and tested. This is not delegated to Run Genesis: RA records resolved inputs but cannot repair an
ambiguous authoritative watermark.

### GD5a — Ruling (2026-07-31)

The company watermark is the PAIR `(guild_id, boundary_seq)` (save-version bump; migration maps
the existing bare seq to `(current guild at migration, seq)` — correct because pre-v-next states
can only have earned watermarks in their current guild). Application rule: a clearing result is
applied iff `result.guild_id == watermark.guild_id AND result.boundary_seq > watermark.seq`. On
membership change (join, or first evaluation where the active guild differs from the watermark's),
the watermark RESETS FORWARD to `(new_guild_id, latest computed boundary_seq for that guild at
that instant)` — never backward, never cross-guild. This is safe by construction: GC allocations
only ever include members present in the boundary's committed snapshot, so a joiner appears first
in post-join boundaries; forfeiting pre-join boundaries is not a loss, it is the definition.
Leaving a guild leaves the watermark in place (inert until the next membership). AC: the
guild-switch fixture from the gap report (boundary 10,000 → other guild boundary 5) applies the
new guild's boundaries 6.. and never re-applies or cross-applies.

## Changelog

- 2026-07-30: created (draft) — the guild owner contract Commons Onboarding blockers #1/#5 named; unblocks the transport guild resolver.
- 2026-07-30: Codex bounce answered — GA (parent resolved), GB (complete literal catalog incl. the 0-ppm consumption hook), GC (total deterministic clearing: ordered one-pass allocation, no intra-boundary redistribution, NPC = one capped virtual counterparty).
- 2026-07-30: implementation audit added GD1–GD6 after the structural implementation exposed missing owner contracts; no mechanic was inferred.
- 2026-07-30: GD1–GD6 answered with executable contracts (name policy; tier-progress tithe units; per-capita Health; deletion succession; projection-side clearing under the RA replay-input rule; canonical slot ordering).
- 2026-07-30: complete-diff review APPROVED with findings (see planning log); ruling GC-1 adopts the implemented headroom-eligibility clearing semantics; AC6 explicitly parked under composition by name.
- 2026-07-31: implementation re-review found GD5a: a per-guild sequence needs its guild identity in the company watermark (or a global sequence); composition remains fail-closed pending owner ruling.
- 2026-07-31: GD5a ruled — watermark is the (guild_id, boundary_seq) pair with forward-only reset on membership change; migration maps bare seqs to the current guild.
- 2026-07-31: runtime-round review — F1 HIGH blocks archival (closed-row deletion trigger); rulings GD3-1 (activity = evaluation) and GD1-1 (separator-stripped denylist matching) recorded.
- 2026-07-31: remediation round approved — all five ruled contracts verified in source; archive unblocked.
- 2026-07-31: runtime remediation implemented — closed-history deletion anonymization, evaluation-only activity events, separator-resistant moderation, content-addressed clearing retries, save-v12 guild-scoped watermarks, lifecycle cleanup, and the deferred rejection/concurrency regressions.
