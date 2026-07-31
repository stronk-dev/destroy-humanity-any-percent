# Production Engine

The authoritative production engine turns the economy kernel's definitions and saved generator
counts into lazy, closed-form resource changes. It never ticks players in the background. A trusted
server caller supplies the current time and an internal `online` or `offline` mode; neither appears
in client intent data.

## Catalog and multiplier boundary

Catalog schema v3 adds manual actions, declared multiplier sources, T0–T3 progress coordinates,
manual-token policy, and offline/Compute Credit policy. The shipped Phase-0 artifact is
[`balance/catalogs/phase0.json`](../balance/catalogs/phase0.json). Historical v1/v2 catalogs remain
loadable for old `constants_hash` saves but cannot acquire v3 semantics silently.

The reviewed rate formula and slot order are published in
[`generated/production-formulas.json`](generated/production-formulas.json). Its SHA-256
`source_fingerprint` is generated from normalized, comment-free Go AST for the live `Rates`
function, multiplier ordering, and Commons formula/aggregate/smoothing authorities. The tool does not pretend to
infer algebra from arbitrary Go; it binds this human-readable formula to the exact executable
authorities that were reviewed. In prose:

```text
rate(resource) = sum over generators (
  exact_owned_count × base_rate × product(multiplier contributions)
)
```

Slots apply in this order: upgrades, milestones, faction, doctrine, commons, trust, event buffs,
prestige. Contributions within one slot apply by source id in raw-byte ascending order. Every
runtime contribution must match a catalog declaration's source, slot, and target and carry a
positive canonical Decimal factor. Commons and Trust are single-provider slots. Providers share
only the neutral `server/multiplier` package; production does not import feature packages.

`make formulas-check` regenerates the published artifact and fails on executable drift. Formatting
and comments do not change the fingerprint; changing a rate or ordering authority requires a
separate reviewed artifact-regeneration commit. The gate is part of the blocking server CI job.

## Time evaluation

`production.Evaluate` integrates all rates through the shared `AccrueConstant` primitive and
commits one positive-accrual ledger transaction:

- online: the complete non-negative elapsed interval at efficiency `1e0`;
- offline: at most `86,400,000` ms at efficiency `9e-1`;
- excess offline time: `floor(excess_ms × 1/2)` Compute Credit ms, capped at `259,200,000`;
- production passes raw non-negative deltas without calculating headroom; the ledger owns capped
  accrual rounding and saturates atomically at the exact declared hardcap;
- the authoritative accrual receipt carries a canonical applied delta that re-adds to its exact
  `after` value, including the one-ulp correction needed at some near-cap boundaries;
- production at an already reached cap succeeds with no ledger change and still advances the
  evaluation cursor, so it cannot block a following intent;
- a server-clock rollback or sub-millisecond interval advances nothing and grants nothing.

Production evaluation and manual-token refill first derive the same
`effective_now = CanonicalServerTime(now)`. They compute work from the exact integer-millisecond
difference and, when positive, set their cursor directly to `effective_now`; they never add a
separately floored duration back to a phase-bearing baseline. Calls within one millisecond and
clock rollback grant nothing. The caller's fractional current-millisecond remainder remains
naturally available to a later call, but it is never persisted as an independent cursor phase.
Compute Credits are integer time state, never a Decimal currency. Spending them belongs to a later
RFC.

## Intent API

The implemented command surface contains exactly eleven intents:

- `buy_generator`: exact positive safe-integer count or verified `max`;
- `perform_manual_batch`: positive safe-integer count and `window_ms` (audit/UX only; it grants no
  authority).
- `cross_gate`: mechanical gate id plus nullable route id; the server evaluates standard or route
  requirements and commits gate state and events atomically.
- `buy_route_hint`: Founder-scope purchase of permanent predicate disclosure with projected Route
  Knowledge; it never affects evaluation.
- `sign_compact`: company-scope membership at an exact catalog-bounded tithe.
- `leave_compact`: company-scope exit at the next accrual boundary; it clears Solidarity.
- `incorporate`: a Tier-2-or-later, once-per-run faction choice. Open Source atomically signs the
  Compact at its catalog tithe; its membership remains faction-bound until the run ends.
- `decline_exit_offer`: clears one live, unexpired deterministic offer.
- `accept_exit_offer`: terminal two-stream commit against a live offer and both expected revisions.
- `wind_down`: terminal elective collapse from Tier 1 onward.
- `file_ipo`: reserved terminal intent, typed `not_eligible` until the S-1 content chain exists.

Both use a lowercase UUIDv7 `intent_id` as the idempotency key and a positive safe-integer
`expected_revision`. The request hash is SHA-256 over deterministic JSON excluding `intent_id`.
Clients never submit prices, production deltas, elapsed server time, balances, or resulting state.

The manual action bucket is exact integer state: 25 milli-tokens per elapsed server millisecond,
50,000 cap, 1,000 per applied action. A full bucket therefore permits 50 immediate actions and
refills at 25 actions/s. Excess requested actions are silently discarded, while `applied_count`
makes the clamp visible. Manual batches do not create click events.

Applied receipts contain the net ledger changes, applied count, resulting revision, evaluation
cursor, and the complete canonical authoritative snapshot. All cursor timestamps serialize the
same UTC millisecond instant used by save v4. Terminal deterministic rejections
(`unaffordable`, `unknown_id`, `invalid`, `cap_exceeded`) are idempotently recorded without a save
mutation. Gate-specific terminal categories are `gate_already_crossed`, `requirement_not_met`,
`route_predicate_unmet`, `insufficient_route_knowledge`, and `already_unlocked`. Revision and
idempotency conflicts are typed but unrecorded so the same logical request
can be retried correctly. A recorded terminal request is different: its UUIDv7 remains bound to
that request hash, so a corrected payload needs a new intent id and reusing the old id returns
`idempotency_conflict`.

Faction terminal categories add `not_eligible`, `already_incorporated`, and `faction_bound`.
Interdependence stock observes the same evaluated elapsed interval through the neutral accrual
hook: spans no longer than the configured catch-up ceiling accrue one unit per catalog interval,
carry the exact integer-millisecond remainder, and saturate at the visible catalog cap. Longer
catch-up spans accrue ordinary offline production but no faction stock. Stock never enters the
Decimal ledger or multiplier stack, and Phase 0 exposes no consumer path.

## Persistence and events

`save.Store.ApplyIntent` locks the stream and performs idempotency lookup before revision checking
or accrual. An applied transition writes the next save revision, its immutable events, and the
normalized receipt record in one Postgres transaction. A terminal rejection writes only its
receipt record. JSON receipts normalize before first return and after JSONB load, so initial and
replayed bytes are identical.

Event registry v1 additionally contains `compact_signed`, `compact_tithe_raised`, `compact_left`, `compact_sampled`,
`compact_health_band_changed`, `compact_cascade_started`, `compact_recovered`, and
`compact_recruitment_offered`; see [Commons](commons.md).
Faction incorporation adds `incorporated` and `faction_stock_saturated`; both are strict v1 event
payloads committed on the same revision as their state change.
Purchases emit exactly one event; manual batches emit none. Events retain stream/revision and
`constants_hash` identity even after old snapshot rows are pruned, but retention does not guarantee
that an old snapshot remains queryable. History has no update/delete API; corrections are later
compensation events.

Applied and replayed intent results also return their immutable event envelopes to internal
post-commit consumers. The Route Registry projector uses this to make projection retry part of the
same idempotency loop: if the gameplay transaction commits but projection fails, replaying the
identical intent re-delivers the original events without re-running gameplay mutation.

Numeric diagnostics flow through the exported `InvariantSink`. Applied fallback/clamp reports are
events on the gameplay revision and become audit/metric records only after that transaction
commits. A terminal rejection has no gameplay revision, so its collected reports become audit and
metrics only after the rejection receipt commits. `residual_abort` is emitted only after rollback
and never creates save or event history. Replays and conflicts do not duplicate observability.
`internal_invariant` remains an internal Go rejection plus audit/metric signal until the WebSocket
Transport RFC owns its wire mapping; no placeholder wire result is claimed here.

Receipt change construction parses every before/after value through the canonical Decimal parser.
A malformed internal snapshot returns Go `ErrInvalidEngineState` (the internal-invariant path); it
is never hidden by silently omitting a change. Transport mapping remains deferred as stated above.

Intent records are keyed by `(stream_id,intent_id)`. The store exposes cutoff-based pruning; the
future deployment scheduler owns calling it at the accepted 30-day retention boundary.

Every Company-run command also appends one `run_log` row in that same transaction. The row contains
the canonical request bytes covered by the idempotency hash, normalized receipt, nullable applied
revision, database server time, and a normalized `replay_inputs` object. Persistence supplies and
validates the authoritative command envelope (intent, Company stream, Founder, revision, run, and
run-log sequence); production owns the closed resolved-input union. Decimal contributions are
canonical strings and are sorted before persistence. Evaluation mode/time, Commons participation
weight, explicit ordered Guild settlement batches (currently produced empty until the scheduler is
composed), Route context version, offer inputs, and terminal Founder carry/route/next-hash inputs
are frozen rather than re-read from mutable projections. Non-empty settlement batches already have
closed UUID/safe-integer/order validation, so composition does not require a replay schema change.

Rows written before replay-input migration deliberately retain SQL NULL and will become `log_gap`
when verification lands. Database triggers reject NULL for every new run-log insert and reject all
updates or deletes of run-log evidence. Terminal
rejections are replay facts and therefore consume a sequence; revision conflicts and idempotent
retries do not. Founder-career commands are outside the Company run log. An Exit allocates its
sequence before building `run_ended`, making the event's `terminal_seq` a transaction-local
completeness proof.

All replayable Company commands execute through `ApplyLogged`; the live service is a producer and
consumer of the same object rather than a separate oracle. The boundary receives Company state,
intent-less canonical payload, one six-artifact catalog bundle, and replay inputs, then returns the
new state, receipt, and ordered events. Terminal Exits use its terminal arm and include the next
catalog bundle selected by the frozen next hash. Founder mutation itself remains a separate stream
audit concern, but every Founder-derived value in the Exit receipt and next Company snapshot is
computed from the frozen carry view. The live service rejects mixed Founder/Company catalog hashes
before freezing that carry, and `ApplyLogged` independently asserts its recorded
`founder_constants_hash` against the bundle. The bundle likewise recomputes its constants identity
from the exact six artifact byte strings; a caller cannot relabel parsed catalogs under another
hash. Company services cannot be constructed without the Prestige/faction runtime.

Faction stock-resource identity is derived inside `ApplyLogged` from the restored faction ID and
the pinned faction artifact. It is not an ambient live-service repair, so restored revisions and
live transitions produce the same receipt field.

The replayable post-accrual registry is closed and ordered: Prestige, faction, Guild, Commons.
Commons receives its projection-derived participation weight as a resolved input. Runtime service
options cannot append hooks to this chain; a new hook requires a new replay-input/RFC version.

## Progress coordinate

Both Go and TypeScript evaluate the same closed catalog union:

- `resource_log`: `Decimal(log10(1+x)).div(Decimal(log10(1+target)))`, clamped to `[0,1]`;
- `count_fraction`: total owned generator count divided by a required safe integer;
- `composite`: deterministic weighted sum of those two kinds.

Resource-log targets are canonical Decimals greater than or equal to `5e-15`; loaders reject any
target whose post-parse denominator is not finite and strictly positive. Go and TypeScript both
divide through their Decimal implementations. Native JavaScript `/` is forbidden in this
evaluator and the schema/source verification gate asserts the `.div` shape. Decimal division by
zero retains the numeric core's existing zero result; the caller prevents a zero denominator.

The shared fixture [`testdata/production-engine.json`](../testdata/production-engine.json) covers
the exact target floor, representative progress parity, and online/offline cap and credit-bank
policy boundaries.

## Verification

`make verify` runs Go, Node, Chromium, Firefox, WebKit, schema, and formula-drift gates. The Go
suite includes a 24-simulated-hour property test over 200 seeded intent policies. With
`TEST_DATABASE_URL` set, integration tests prove atomic apply/replay, one purchase event,
hash/revision conflicts, terminal rejection non-mutation, pruning, exact buying, and the manual
clamp against real Postgres 16.
