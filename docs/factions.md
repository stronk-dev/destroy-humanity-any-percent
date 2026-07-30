# Factions and Incorporation

Phase 0 defines four run-scoped factions in the strict
[`balance/factions/phase0.json`](../balance/factions/phase0.json) catalog. Each faction produces
one interdependence stock and consumes another; the loader proves that the four producer/consumer
edges form one Hamiltonian cycle with no self-consumption and exactly one producer and consumer
per stock:

| Faction | Produces | Consumes | Compact binding |
|---|---|---|---|
| `bootstrapper` | `revenue` | `compliance` | none |
| `vc_funded` | `hype` | `revenue` | none |
| `open_source` | `libraries` | `hype` | auto-sign at 130,000 ppm |
| `enterprise` | `compliance` | `libraries` | none |

The catalog also owns the 100,000-unit visible stock cap and 60,000-ms accrual interval. Its
modifier lists are empty, so choosing a faction does not silently alter the production stack in
this epoch. The exact artifact is part of the balance-epoch constants identity; its introduction
minted Epoch 2 and regenerated the harness baseline in a separate guarded commit.

## Incorporation

`incorporate {faction_id}` is an exact-schema, idempotent Company intent. It is available once per
run at Tier 2 or later. Unknown ids, early use, and a second choice return typed terminal
rejections without mutating the revision. Success records the server's canonical evaluation
instant, emits `incorporated`, and exposes the derived stock resource in the authoritative receipt.

Open Source incorporation additionally emits `compact_signed` and establishes Compact membership
at 130,000 ppm in the same save/event/receipt transaction. An already-signed company must leave
before choosing Open Source; after incorporation, `leave_compact` returns `faction_bound` until an
Exit ends the run. The other three choices do not modify existing Compact state.

Prestige new-run assembly clears faction identity, incorporation time, stock, and remainder. The
Founder stream never contains faction state, so a later run may choose again independently.

## Interdependence stock

Stock is exact integer Company state, not a Decimal ledger currency. After every ordinary lazy
production evaluation, the faction accrual hook receives the same authoritative elapsed result.
An elapsed span at or below the run's hash-pinned Prestige `catchup_ceiling_ms` is attended:
elapsed milliseconds join the saved remainder, each complete catalog interval earns one unit, and the remainder carries. A longer
catch-up span advances ordinary offline production but earns no faction stock.

Accrual saturates at the catalog cap rather than overflowing. Crossing the cap emits one
`faction_stock_saturated` event; subsequent accrual retains the time remainder but forfeits units
while full. Phase 0 intentionally has no consume or exchange operation. `consumed_stock_units` is
reserved persisted state for the Guild exchange RFC, and no current transition changes it.

Leaderboard rows include a nullable `faction` structural variable. Database storage and query
types enforce that shape now; the Run Genesis verifier will derive the value from the immutable
`incorporated` event when replay verification ships.
