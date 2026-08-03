# Purchasable Content Foundation

The purchasable-content foundation is the implemented, catalog-driven layer above the numeric and
production kernels. It provides mechanics for upgrades, generator chains, milestone ladders,
synergy pools, and typed generator roles without introducing alternate arithmetic paths. Economy
schema version 4 owns these definitions. The active Phase-0 catalog remains schema version 3 until
the first T0–T1 content epoch is explicitly minted, so this document describes executable
capability rather than claiming unshipped balance content.

## Catalog and ownership

An upgrade has an immutable cost, a gate-bounded availability window, a closed Route-condition
predicate, and zero or more Decimal multiplier contributions. `buy_upgrade` evaluates accrued
state first, checks the window and predicate against that state, applies the exact ledger purchase,
and then records ownership plus one `upgrade_purchased` event. Rebuying an owned upgrade is a typed
`not_eligible/owned` rejection. Owned upgrades contribute only through the deterministic
multiplier stack.

Every schema-v4 generator class has at least one typed, executable role:

- `provision` binds exactly to its one-tier-down chain edge;
- `synergy_feed` binds exactly to a declared generator source in a named pool;
- `manual_output` binds to a real manual action and emits
  `1 + purchased_count × per_purchased_ppm / 1,000,000` through the upgrades slot;
- `stock_rate` scales faction-stock progress with the same exact integer-ppm factor and a
  persisted remainder.

Duplicate role bindings, roleless generator classes, dangling role targets, unmatched pool
sources, invalid tier topology, and undeclared multiplier sources reject the entire catalog. Roles
describe executed mechanics; they are not free-form labels.

## Purchased and provisioned counts

Save version 14 keeps purchased counts in `generators` and adds complete-key maps for
`generators_provisioned` and `provision_remainders_ppm`, plus the persisted stock-rate remainder.
Production uses purchased plus provisioned counts. Costs, purchase totals, ladders, manual roles,
stock roles, and synergy sources use purchased counts only.

Provisioning runs on an absolute 60-second grid aligned to `run_started_at`. At each boundary all
edges read the pre-boundary totals, compute their exact integer-ppm quotient and remainder, stage
their results, and commit simultaneously. New units therefore produce only in the following
bucket, and splitting the same interval across evaluations cannot change the result. Each target
declares an exact safe-integer cap and mechanical reason key. Provisioning alone may saturate at
that cap; the authoritative receipt snapshot exports the cap and reason so a future content UI can
explain a frozen counter.

## Ladders, pools, and formulas

Generator ladders are cumulative purchased-count thresholds. Every crossed rung emits its own
declared contribution. Synergy pools consume purchased generator counts and owned-upgrade flags,
then emit exactly one contribution in raw-byte source order:

```text
linear = 1 + sum_ppm / 1,000,000
log    = 1 + log10(1 + sum_ppm / 1,000,000)
```

Factors quantize once at emission. Pools cannot feed pools. The generated production-formula
artifact publishes the formulas, active tick, caps, and pool composition and fingerprints the
executable Go authorities that produce them.

## Simulation boundary

The balance harness has one simulation-only entrypoint accepting separate closed effect masks and
action-removal sets for generators, upgrades, and manual actions. A generator effect mask nulls
base production, ladder effects, pool feeds, manual factors, stock-rate effects, and outgoing
provisioning while preserving ownership and costs. An upgrade effect mask nulls its contribution
bundle. Action removal makes the matching purchase or manual action reject as unknown before
accrual. Simulation receives the same route catalog and accrual hook as the route under test, and
reports a typed role activation only when an applied transition executes a non-neutral role result.
The authoritative and replay entrypoints have no mask parameter, catalog field, save field, or
replay-input field; a Go source guard permits the simulation entrypoint only from the harness and
tests.

## Verification

The shared schema-v4 fixture is `testdata/economy-foundation-v4.json`. Go and TypeScript restore
the same save-v14 state and replay Go-authored upgrade, manual-role, provision-grid, and combined
content transitions byte-for-byte. Additional Go tests cover partition invariance, next-bucket
production, cap saturation, purchased-only ladders, simulation isolation, real-hook stock-rate
activation, exact linear/log factors, and deterministic two-pool ordering. The live schema-v3
balance artifact and pacing baselines do not change until a separately governed content mint.
