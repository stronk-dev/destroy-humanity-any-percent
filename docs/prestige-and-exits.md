# Prestige and Exits

Prestige ends one Company run and advances the owning Founder without changing either stream's
identity. The server commits the old run's terminal state, Founder rewards, and the next run's
initial state in one PostgreSQL transaction. A fault at any write boundary leaves both scopes at
their pre-Exit revisions; an idempotent retry returns the original receipt and events.

## Persisted state

Save version 9 carries the run lifecycle explicitly. Company state records tier, lifetime value,
the optional live offer, run start time, server-derived offline spans, and the exact duration of
older spans collapsed out of the bounded list. Founder state records
Reputation, Reputation unlock strength, Network slots, lifetime Clout, Soul, age, Notoriety,
Advisor Mode, and append-only Exit history. A pre-v7 Company backfills its missing run start from
`evaluated_through` and persists `run_pre_timer=true`, so it can Exit but cannot claim a time-ranked
record for a run that predates timer semantics. V9 encoding refuses non-canonical cursor times or
invalid cross-scope state.

The Phase-0 Prestige policy is declarative in
[`balance/prestige/phase0.json`](../balance/prestige/phase0.json). Its `value_resource_id` names the
authoritative Company resource whose positive accrual advances lifetime value. Its
`catchup_ceiling_ms` is the sole server-side attended/offline boundary used by both Prestige span
accounting and faction stock accrual. Offer duration, spawn gates, decline drift, payout modifiers,
collapse Route Knowledge, and Advisor constants are data, never code constants.

## Exact arithmetic

The new Reputation level is `floor(cuberoot(lifetime_value / threshold))`. Both Go and TypeScript
find the largest exact integer whose Decimal cube does not exceed the ratio using integer binary
search. TypeScript constructs the ratio directly from mantissa division and exponent subtraction,
matching the Go Decimal path without reciprocal double-rounding. Neither runtime calls a floating-
point cube-root function. Exit modifiers are integer ppm applied to the positive level delta and
then floored using exact integer products (`big.Int` in Go and `BigInt` in TypeScript); a result
beyond the exact integer domain saturates at `9,007,199,254,740,991` so an
extreme valid run cannot become un-exitable. The shared Prestige vector corpus asserts cube
boundaries including non-unit thresholds, modifier rounding, saturation, zero, threshold equality,
and the exact-number cap in both runtimes.

Server-computed terms also clamp Reputation delta to the Founder's remaining exact-domain
headroom. Offer previews therefore never promise a payout that the Founder hardcap cannot accept.

## Offers and intents

The production intent surface adds:

- `accept_exit_offer`, with the live offer id and both expected stream revisions;
- `decline_exit_offer`, which clears the live offer and advances its deterministic drift walk;
- `wind_down`, the always-open elective collapse from Tier 1 onward;
- `file_ipo`, currently returning typed `not_eligible` until the S-1 content chain exists.

Offer checks happen only at deterministic evaluation sites. Their SplitMix64 stream is seeded from
the immutable Founder id and run sequence. Offers cannot spawn until Founder Exit history is
non-empty, so the market path cannot bypass the scripted first-collapse curriculum. A spawned offer persists its server-computed terms and
market modifier. Acceptance recomputes against commit-time state, reapplies that same modifier,
then takes the field-wise maximum for integer rewards and the set union for Network slots. The
preview therefore remains a promise as the run advances. Expiry and decline are events; there are
no background timers. Decline drift counts only declines from the current `run_seq`, so a new run
begins with a clean offer walk.

## Exit transaction and run facts

`save.Store.ApplyExitTransaction` locks Founder then Company, validates both expected revisions,
and commits Founder revision `+1` plus Company terminal and new-run revisions `+2`. The idempotency
record belongs to the Company stream. The `run_ended` event is self-contained for the obituary:
run identity, exit type, server start/end times, RTA, Attended Time, payout, tier, lifetime value,
ledger facts, revision-bounded executed routes, its pre-timer status, and separate Commons/Advisor
assisted variables plus the nullable run faction.
`run_started` is the next timer's `[BEGIN ATTEMPT]` fact.
The terminal sequence is the Exit command's atomically committed per-run intent-log sequence, not a
save revision or an eventually projected counter. Exit events also receive a database-authored
commit sequence; idempotent replay returns every event for the intent in that recorded order,
including evaluation events that precede the terminal Exit events.

Attended Time is RTA less recorded offline spans. An accrual gap larger than the run-hash-pinned
Prestige policy's catch-up ceiling is recorded using canonical integer milliseconds. When the bounded 256-span list
fills, the oldest exact duration moves into `collapsed_offline_ms` before the span is removed;
online gaps are never absorbed and total offline duration is invariant.

The first Founder run has one scripted curriculum Exit: the first threshold crossing at or after
900,000 attended milliseconds ends as `scripted_first`. Choosing Wind Down before that trigger is
also typed `scripted_first`; the curriculum cannot be skipped through the always-open Exit. The
Founder Exit history makes it once-per-Founder, while creating a New Founder provides a genuinely
fresh lifecycle. The pacing gate measures the first later, genuinely elective Exit.

## New-run assembly

The next Company state is deterministic: catalog initials, carried Network items, Reputation
starter effects, moral-score reseed, and zero this-run Clout. The current fixture has no Network or
Reputation-tree content, so those declared seams are empty. Trust-facing meter bands reseed from
Founder Notoriety using the published clamp, while lifetime Founder facts and the old run's ledger
facts remain intact.

An Exit may cross a balance-epoch boundary. The terminal Company revision and `run_ended` event
retain the ended run's pinned hash. The Founder transition, new Company revision, `run_started`
event, receipt snapshot, and run-N+1 epoch pin all use the process's current constants hash in the
same transaction. The next run is therefore assembled from current catalog initials without
rewriting the forensic identity of the run that just ended.

Gate delivery is monotonic: crossing a catalog-legal gate records the fact, while Company tier is
updated to `max(current_tier, gate_tier)`. A delayed lower-tier gate can never brick an otherwise
valid run.

## Verification status

Unit and shared-vector tests cover arithmetic, offer monotonicity, deterministic run assembly,
offline accounting, and exact parsing. Real-Postgres tests inject failure at every Exit write
boundary, assert byte-identical replay, execute elective and scripted Exits end to end, preserve
executed-route facts, and prove a progressed offer never pays below its preview. The suite also
mints an epoch with changed artifact bytes while a run is active and exits it across the boundary,
and migrates a literal v6 Company through a successful pre-timer scripted Exit. The first elective
Exit pacing envelope remains gated on the explicitly missing T0–T1 playable-content contract; no
fixture timing is presented as shipped balance.
