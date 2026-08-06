# Fiscal Quarters

Fiscal Quarters are a Founder-scoped, wall-clock progression mechanic. The implementation is
present in both runtimes, but no production epoch currently contains the optional `fiscal`
artifact, so the mechanic remains inactive until a later balance mint.

## Catalog and activation

The strict schema-v1 `fiscal` artifact owns the ripening windows, early-harvest probability,
credit and level hardcaps, hoard policy, generator-level rows, and unlock rows. Go and TypeScript
reject unknown, missing, duplicate, unsorted, unsafe-integer, and cross-artifact-invalid values.
The artifact bytes participate in the catalog-bundle hash and constants identity.

The artifact activates Founder save v19 only at a new-run boundary. It requires the pinned
minigame and pet artifacts that own Founder v17 and v18; the Company save axis remains unchanged.
New founders are initialized from the current pinned bundle through the same activation path.
Pre-v19 founders finish their current run without Fiscal state.

Founder v19 persists:

- integer `fiscal_credit`;
- the current period's opening wall timestamp and monotonic sequence;
- complete generator-level and unlock collections declared by the pinned artifact.

The codec rejects Fiscal state before v19 and incomplete, negative, over-cap, or unknown state at
v19.

The catalog loader also rejects a ripening horizon capable of exhausting
`fiscal_period_seq` within the safe wall-timestamp domain. The sequence never saturates because it
is part of the deterministic early-harvest draw identity.

## Wall-clock periods

Period ripeness is derived lazily from the server-authored wall timestamp and
`period_opened_wall_ms`; there is no background tick or stored ripeness counter. Before every
valid Founder command, the transition closes every complete `auto_ms` period in one phase-
preserving sweep, credits the aggregate mint, advances the period sequence, and emits one
`fiscal_period_harvested.v1` event with source `automatic` before the command event.

If the triggering command rejects, the sweep rejects with it and the entire Founder transition
rolls back. The next command deterministically retries the same sweep. This preserves the global
rule that rejected commands do not mutate state.

`harvest_fiscal_period` has three outcomes from the pinned windows:

- before `early_ms`, it rejects as `not_eligible` with detail `period_not_ripe`;
- from `early_ms` to `guaranteed_ms`, it uses SplitMix64 substream
  `fiscal.early_harvest.v1`, derived from the run-identity seed and period sequence;
- from `guaranteed_ms`, it succeeds without a draw.

Both early success and early failure consume the period and open the next one. Wall time determines
eligibility; it never supplies randomness. The resolved command logs the authoritative timestamp,
sequence, draw, swept-period count, and outcome, allowing replay without a live clock.

## Spending and frozen run effects

`spend_fiscal_credit` accepts no client-authored amount. The server resolves the exact cost from
the pinned artifact and current Founder state. Its closed targets are:

- `generator_level`, whose multi-level cost is the checked triangular sum and whose level hardcap
  has an announced reason key;
- `unlock`, whose eligibility, cost, and hardcap reason are catalog data.

Insufficient credit rejects as `unaffordable`; invalid or capped targets use the existing typed
rejection taxonomy. Applied spends debit exactly once and emit `fiscal_credit_spent.v1`.

Founder-derived production effects never read mutable Founder state during a Company command.
When a new Company run is created, the server materializes a complete immutable
`run_frozen_contributions` set containing:

- one hoard contribution derived from unspent credit, capped by the pinned hoard policy;
- one generator-level contribution for every declared generator row.

The rows are inserted in the same transaction as the run pin and genesis. A deferred database
constraint rejects missing or extra rows, retries must reproduce identical values, and committed
rows are immutable. Production resolves them only by Company stream and run sequence, then carries
them through the existing replay-input contribution path. Mid-run Fiscal harvests or spends affect
the next run only.

## Determinism and authority

Both intents mutate only through `ApplyFounderLogged`; payload, resolved inputs, receipt, and event
order are replay-owned. The Founder transition is the single authority for automatic sweeping.
Go and TypeScript share strict catalog fixtures and exercise period boundaries, seeded outcomes,
multi-period catch-up, spending, caps, rollback, save migration, and replay parity. Real-Postgres
tests cover frozen-row completeness, immutability, idempotent retries, Exit insertion, New-Founder
initialization, and import initialization.
