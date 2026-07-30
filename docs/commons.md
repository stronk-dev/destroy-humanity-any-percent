# Commons Compact

The Mutual Aid Compact is an optional, company-scoped production agreement. The server owns
membership, derives compliance from the already-validated multiplier stack, assigns a permanent
Founder cohort, and exports one neutral `commons.member` multiplier contribution. Production does
not import Commons code and Commons does not import production; the composition layer connects
their shared multiplier and post-accrual contracts.

This is the shipped server foundation. The incorporation contract line item, Open Source auto-sign,
cohort UI, guild term, and monthly direction vote are not yet shipped; their explicit successor is
`rfc/commons-onboarding-and-governance.md`, which depends on the client, account, faction, and guild
contracts rather than inventing them here.

## Catalog and arithmetic

[`balance/commons/phase0.json`](../balance/commons/phase0.json) is strict schema v1. Bounded ratios
are exact integer millionths (`ppm`); large production/capacity values and the exported modifier
are canonical RFC-0001 Decimals. The shipped starting values are:

Capacity is cumulative: every idempotently projected accrual adds its tithe to the member, cohort,
and server absolute totals. Replaying the same sample event adds nothing, and Capacity never scales
the production modifier.

- tithe default 10%, selectable from 5% through 15%;
- effective Health weights: guild 50%, cohort 30%, server 20%; guildless members substitute cohort
  Health for guild, making the effective split cohort 80% / server 20%;
- collapse below 0.35, healthy at 0.80, collective/personal modifier split 60%/40%, convex
  exponent 1.5, maximum bonus `5e0`;
- Health recovery 10%/hour and decay 4%/hour, applied lazily with the closed exponential form;
- 30-day Solidarity window, 150-Founder cohort target, merge floor 40;
- below 40 real samples, each missing member adds labeled NPC weight 0.5 at compliance 0.75.

Only cohort and server scopes exist today. With no shipped guild model, every member takes the
specified guildless fallback, so effective Health is cohort 80% / server 20%. The future governance
layer may add the guild 50% term without changing the current projection contract.

Both the live projection and population harness resolve the Commons catalog by the authoritative
revision's `constants_hash` and call the same `EffectiveHealthPPM` function. There is no hardcoded
80/20 runtime shortcut: changing the published weights changes both paths and the constants
identity together.

The authoritative formulas, exact Enclosure source-weight table, and every shipped Commons catalog
control (including labeled NPC weight/compliance and population floor) are generated into
[`generated/production-formulas.json`](generated/production-formulas.json). The executable
Enclosure, effective-Health blend, modifier, aggregate-Health, smoothing, production-rate, and ordering authorities all
participate in its source fingerprint.

Enclosure is derived only from active multiplier source IDs/factors. For each catalog-weighted
source, the accepted factor is raised to its slot weight; `d = clamp(1-clean/all,0,1)`. Route flags,
names, hints, votes, and player declarations are not inputs. The Phase-0 forsworn-source table is
empty because no shipped T0–T3 multiplier has yet been classified as a dark/externality source;
future balance catalogs add rows alongside their economy source declarations.

Compliance is `clamp(tithe/target,0,1) × (1-d)`. Solidarity is the sum of hourly
`compliance_ppm × covered_ms` over the fixed 2,592,000,000-ms window divided by that complete
window, so it genuinely rebuilds from zero rather than instantly treating a partial window as a
full history. Capacity is the deterministic sum of committed positive accrual deltas multiplied by
the tithe; it is projected for gates/caps and never enters the Health-driven rate formula.

The member modifier is:

```text
1 + 5 × (0.6 × max(0, ((H - 0.35) / 0.65)^1.5) + 0.4 × Solidarity)
```

At collapse a fully loyal member retains ×3.0 (base ×1 plus the ×2 personal term); no Commons
state can reduce the base game below ×1. A non-member supplies no Commons slot at all.

## Membership and accrual

The strict intents are `sign_compact {tithe_ppm}` and `leave_compact`. Both accrue at the old
membership state first and change membership on that exact authoritative boundary. Repeat sign
and non-member leave reject as `already_member` / `not_member`; out-of-band tithe is `invalid`.
Leaving is always allowed and clears the entire Solidarity window. Re-signing starts at zero.

Member accrual emits `compact_sampled` with founder/run identity, participation weight,
compliance, Enclosure, tithed Capacity, resulting Solidarity, and sampled milliseconds. The hook
runs only during an authoritative intent transition; there is no player or world tick loop.
The first member accrual uses the declared neutral NPC Health when no projection exists. Later
intents consume the latest projected effective Health and contribute exactly one
`commons.member` factor in the fixed Commons slot.

## Cohorts and projections

The first `compact_signed` event assigns the Founder to the oldest open cohort in the
server/activity bracket below its 150 target. A server/bracket advisory transaction lock makes
concurrent first-sign order an actual database decision. Assignment is non-elective and reused by
later runs and re-signs. Membership and sample projections are idempotent by event ID.
`compact_tithe_raised` preserves that assignment and membership while advancing the projected
tithe at the same Company revision as Open Source incorporation.
After strict payload validation, each projection transaction claims its event ID before resolving
current catalog, assignment, or membership state. An already-committed event returns successfully
without consulting those mutable dependencies, so replay after a later leave cannot wedge the
worker. A first delivery's claim remains in the same transaction as every derived write; any
validation or write failure rolls the claim back and remains retryable.
Membership rows retain the highest projected company-stream revision. For equal timestamps, an
older separately delivered leave therefore cannot overwrite a later re-sign; in-batch ordering
uses that same revision before kind/event tie-breakers. Revisions are deliberately not compared
across streams.

Collapse merge is explicit and one-way: membership below the configured floor is the only trigger;
Health is not. An additional below-floor cohort moves whole into the oldest compatible cohort only
when the result is at most `floor(1.5 × cohort_target_size)` (225 at the shipped target of 150).
A source at the floor or one that would exceed the ceiling stays intact—members are never split.
Successful source cohorts close, and standing is recomputed from member weight/compliance inputs.
Rounded cohort means are never averaged together.

Health is a weighted mean; Capacity is a Decimal sum. Cohort and server snapshots store both raw
and asymmetrically smoothed Health plus labeled NPC weight. Health band changes create immutable
server dispatch events (`compact_health_band_changed`, `compact_cascade_started`,
`compact_recovered`). The progress boundary can call the idempotent recruitment operator at
mid-T3; `commons_recruitment_offers` and `compact_recruitment_offered` guarantee at most one offer
per never-signed Founder. Transport and visible client panels remain owned by their active RFCs.

## Verification

Shared Go/TypeScript vectors cover Enclosure and the H × Solidarity modifier edges. Save migration
tests cover v1–v5 to v6. Real Postgres tests cover concurrent cohort assignment, full-history
replay after leave, failed-first-delivery rollback, sign/leave, sample projection, stable re-sign
identity, collapse merge, and recruitment uniqueness. The
SplitMix64 harness runs 128 seeds at 200 and 20,000 members; mean modifiers must stay within 100 ppm
and the 95% intervals must overlap. `make verify-commons-boundary`, `make commons-harness-check`,
formula drift, schema checks, and the complete `make verify` matrix are blocking gates.
