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

The rate formula and slot order are generated from the Go source into
[`generated/production-formulas.json`](generated/production-formulas.json). In prose:

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

`make formulas-check` regenerates the published artifact and fails on drift. It is part of the
blocking server CI job.

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

The evaluation cursor advances by the exact whole milliseconds consumed, preserving any
sub-millisecond remainder for the next evaluation. Compute Credits are integer time state, never a
Decimal currency. Spending them belongs to a later RFC.

## Intent API

The implemented command surface contains exactly two intents:

- `buy_generator`: exact positive safe-integer count or verified `max`;
- `perform_manual_batch`: positive safe-integer count and `window_ms` (audit/UX only; it grants no
  authority).

Both use a lowercase UUIDv7 `intent_id` as the idempotency key and a positive safe-integer
`expected_revision`. The request hash is SHA-256 over deterministic JSON excluding `intent_id`.
Clients never submit prices, production deltas, elapsed server time, balances, or resulting state.

The manual action bucket is exact integer state: 25 milli-tokens per elapsed server millisecond,
50,000 cap, 1,000 per applied action. A full bucket therefore permits 50 immediate actions and
refills at 25 actions/s. Excess requested actions are silently discarded, while `applied_count`
makes the clamp visible. Manual batches do not create click events.

Applied receipts contain the net ledger changes, applied count, resulting revision, evaluation
cursor, and the complete canonical authoritative snapshot. Terminal deterministic rejections
(`unaffordable`, `unknown_id`, `invalid`, `cap_exceeded`) are idempotently recorded without a save
mutation. Revision and idempotency conflicts are typed but unrecorded so the same logical request
can be retried correctly.

## Persistence and events

`save.Store.ApplyIntent` locks the stream and performs idempotency lookup before revision checking
or accrual. An applied transition writes the next save revision, its immutable events, and the
normalized receipt record in one Postgres transaction. A terminal rejection writes only its
receipt record. JSON receipts normalize before first return and after JSONB load, so initial and
replayed bytes are identical.

Event registry v1 contains `generator_purchased`, `invariant_reported`, and `compensation`.
Purchases emit exactly one event; manual batches emit none. Events retain stream/revision and
`constants_hash` identity even after old snapshot rows are pruned. History has no update/delete
API; corrections are later compensation events.

Intent records are keyed by `(stream_id,intent_id)`. The store exposes cutoff-based pruning; the
future deployment scheduler owns calling it at the accepted 30-day retention boundary.

## Progress coordinate

Both Go and TypeScript evaluate the same closed catalog union:

- `resource_log`: `log10(1+x) / log10(1+target)`, clamped to `[0,1]`;
- `count_fraction`: total owned generator count divided by a required safe integer;
- `composite`: deterministic weighted sum of those two kinds.

The shared fixture [`testdata/production-engine.json`](../testdata/production-engine.json) covers
progress parity plus online/offline cap and credit-bank policy boundaries.

## Verification

`make verify` runs Go, Node, Chromium, Firefox, WebKit, schema, and formula-drift gates. The Go
suite includes a 24-simulated-hour property test over 200 seeded intent policies. With
`TEST_DATABASE_URL` set, integration tests prove atomic apply/replay, one purchase event,
hash/revision conflicts, terminal rejection non-mutation, pruning, exact buying, and the manual
clamp against real Postgres 16.
