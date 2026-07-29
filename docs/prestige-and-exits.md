# Prestige and Exits

Prestige ends one Company run and advances the owning Founder without changing either stream's
identity. The server commits the old run's terminal state, Founder rewards, and the next run's
initial state in one PostgreSQL transaction. A fault at any write boundary leaves both scopes at
their pre-Exit revisions; an idempotent retry returns the original receipt and events.

## Persisted state

Save version 7 carries the run lifecycle explicitly. Company state records tier, lifetime value,
the optional live offer, run start time, and server-derived offline spans. Founder state records
Reputation, Reputation unlock strength, Network slots, lifetime Clout, Soul, age, Notoriety,
Advisor Mode, and append-only Exit history. Older saves migrate with zero, empty, or null defaults;
v7 encoding refuses non-canonical cursor times or invalid cross-scope state.

The Phase-0 Prestige policy is declarative in
[`balance/prestige/phase0.json`](../balance/prestige/phase0.json). Its `value_resource_id` names the
authoritative Company resource whose positive accrual advances lifetime value. Offer duration,
spawn gates, decline drift, payout modifiers, collapse Route Knowledge, and Advisor constants are
data, never code constants.

## Exact arithmetic

The new Reputation level is `floor(cuberoot(lifetime_value / threshold))`. Both Go and TypeScript
find the largest exact integer whose Decimal cube does not exceed the ratio using integer binary
search. Neither runtime calls a floating-point cube-root function. Exit modifiers are integer ppm
applied to the positive level delta and then floored. The shared Prestige vector corpus asserts
cube boundaries, modifier rounding, zero, threshold equality, and the exact-number cap in both
runtimes.

## Offers and intents

The production intent surface adds:

- `accept_exit_offer`, with the live offer id and both expected stream revisions;
- `decline_exit_offer`, which clears the live offer and advances its deterministic drift walk;
- `wind_down`, the always-open elective collapse from Tier 1 onward;
- `file_ipo`, currently returning typed `not_eligible` until the S-1 content chain exists.

Offer checks happen only at deterministic evaluation sites. Their SplitMix64 stream is seeded from
the immutable Founder id and run sequence. A spawned offer persists its server-computed terms and
market modifier. Acceptance recomputes against commit-time state, reapplies that same modifier,
then takes the field-wise maximum for integer rewards and the set union for Network slots. The
preview therefore remains a promise as the run advances. Expiry and decline are events; there are
no background timers.

## Exit transaction and run facts

`save.Store.ApplyExitTransaction` locks Founder then Company, validates both expected revisions,
and commits Founder revision `+1` plus Company terminal and new-run revisions `+2`. The idempotency
record belongs to the Company stream. The `run_ended` event is self-contained for the obituary:
run identity, exit type, server start/end times, RTA, Attended Time, payout, tier, lifetime value,
ledger facts, revision-bounded executed routes, and separate Commons/Advisor assisted variables.
`run_started` is the next timer's `[BEGIN ATTEMPT]` fact.

Attended Time is RTA less recorded offline spans. An accrual gap larger than the production
catch-up ceiling is recorded using canonical integer milliseconds; the bounded span list collapses
its oldest entries without changing the total offline duration.

The first Founder run has one scripted curriculum Exit: the first threshold crossing at or after
900,000 attended milliseconds ends as `scripted_first`. The Founder Exit history makes it
once-per-Founder, while creating a New Founder provides a genuinely fresh lifecycle. Elective
Wind Down remains separate and is the event measured by the future T0–T1 first-Exit pacing gate.

## New-run assembly

The next Company state is deterministic: catalog initials, carried Network items, Reputation
starter effects, moral-score reseed, and zero this-run Clout. The current fixture has no Network or
Reputation-tree content, so those declared seams are empty. Trust-facing meter bands reseed from
Founder Notoriety using the published clamp, while lifetime Founder facts and the old run's ledger
facts remain intact.

## Verification status

Unit and shared-vector tests cover arithmetic, offer monotonicity, deterministic run assembly,
offline accounting, and exact parsing. Real-Postgres tests inject failure at every Exit write
boundary, assert byte-identical replay, execute elective and scripted Exits end to end, preserve
executed-route facts, and prove a progressed offer never pays below its preview. The first elective
Exit pacing envelope remains gated on the explicitly missing T0–T1 playable-content contract; no
fixture timing is presented as shipped balance.
