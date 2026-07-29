# Economy Kernel

The economy kernel is the implemented boundary between RFC-0001 large-number arithmetic and
future gameplay systems. It provides a strict shared catalog, matching Go/TypeScript cost queries,
and a scoped authoritative Go ledger. The production engine now consumes this boundary for shipped
balance data, purchases, multiplier declarations, and lazy time integration.

## Catalog

The authoring schema is [`balance/economy.schema.json`](../balance/economy.schema.json). Runtime
validation is performed independently by `economy.LoadCatalog` on the server and `parseCatalog`
on the client. Current authoring uses version 3; both runtimes retain version-1 and version-2
loading for historical catalog hashes. The complete canonical example is
[`balance/catalogs/phase0.json`](../balance/catalogs/phase0.json); this excerpt shows its root shape:

```json
{
  "schema_version": 3,
  "resources": [
    {
      "id": "company.cash",
      "scope": "company",
      "numeric_kind": "decimal",
      "initial": "0",
      "minimum": "0",
      "hardcap": {
        "amount": "1e1000",
        "reason_key": "resource.company_cash.cap.depleted"
      }
    }
  ],
  "generator_classes": [
    {
      "id": "generator.example",
      "price": {
        "resource_id": "company.cash",
        "base": "1e2",
        "curve": { "kind": "geometric", "ratio": "1.13e0" }
      },
      "production": { "resource_id": "company.cash", "base_rate": "1e0" }
    }
  ],
  "manual_actions": [
    { "id": "manual.click", "output": { "resource_id": "company.cash", "amount_per_action": "1e0" } }
  ],
  "multiplier_sources": [],
  "progress_coordinates": ["four strict T0–T3 coordinate objects"],
  "manual_policy": { "refill_milli_per_ms": 25, "bucket_cap_milli": 50000 },
  "offline_policy": { "the strict time-policy fields": "see the canonical artifact" }
}
```

The shipped Phase-0 values are provisional balance data identified by the exact artifact-byte
`constants_hash`.

Catalog rules:

- Versions 2 and 3 require generator production metadata. Version 1 remains readable but its
  generators are explicitly not production-capable. Version 3 additionally requires production
  policies, actions, multiplier declarations, and progress coordinates.
- IDs are lowercase mechanical identifiers, optionally dot-namespaced. Flavor and visible text
  stay in later data/localization files.
- Resource scopes are `company`, `founder`, `world`, and `guild`.
- `numeric_kind` is explicit and currently only supports `decimal`. Owned counts remain exact safe
  integers rather than Decimal resources.
- Every numeric value is an RFC-0001 canonical string, never a JSON number.
- `hardcap` must be explicitly an `{amount, reason_key}` object or `null`. The technical Decimal
  limit is never a visible gameplay cap.
- Missing fields, unknown fields, duplicate IDs, dangling resource references, invalid bounds,
  unsupported tags, and malformed canonical values fail the entire load.
- A production `base_rate` is a positive canonical Decimal per second. Price and output resources
  must share a scope; cross-scope transfers require an explicit coordinator.
- Loading from bytes/objects is implemented. Disk watching, `go:embed`, and content hot-reload
  orchestration belong to later server/client-shell work.

## Cost curves

Curve types are code and their parameters are data. The initial closed set is:

```text
constant:  price(n) = base
linear:    price(n) = base + step × n       step ≥ 0
geometric: price(n) = base × ratio^n        ratio ≥ 1
```

`base` must be positive. `owned` and purchase `count` must be non-negative exact integers, and
their sum may not exceed JavaScript's maximum safe integer.

Both runtimes expose bulk-cost and max-affordable queries. Bulk cost sums the next `count` prices
after `owned`; geometric cost uses the RFC-0001 closed form. On the authoritative Go path,
geometric max-affordable uses RFC-0001's closed-form inverse, caps the candidate at
`MaxExactInteger - owned`, and verifies it through the public bulk-cost semantics. Local correction
handles rounding boundaries; a bounded search remains the safety fallback. Constant and linear
curves use the bounded search directly.

Every max-affordable result guarantees that its cost is affordable and that the next count is not,
unless `owned + count` has reached the exact-integer ceiling. The optimization therefore changes
runtime, not economy results or balance.

There is no global price ratio. `1.10`, `1.13`, and other valid ratios are ordinary per-generator
data. Shared fixtures currently exercise both values and enormous exponents; they are tests, not
launch tuning. Adding another curve kind requires a follow-up RFC plus Go, TypeScript, and shared
parity coverage. Balance data cannot contain executable formulas or callbacks.

## Scoped authoritative ledger

`economy.NewLedger(catalog, scope)` creates one server-side bank for exactly one catalog scope.
Company, Founder, World, and Guild values cannot be read or mutated through one another's ledger.
The future actors own ledger identity—for example, which company or guild a particular ledger
belongs to. Cross-scope actions such as an Exit require an explicit higher-level coordinator and
cannot be smuggled through a primitive transaction.

The public primitive boundary is:

- `Balance(resourceID)` and `Snapshot()` for read-only queries;
- `Apply(Transaction)` for strict ordinary mutations such as purchases and grants;
- `ApplyAccrual(Transaction)` exclusively for non-negative production accrual;
- `Receipt` as the committed notification fact.

Both mutation operations:

1. validates resource existence, ledger scope, and every finite delta;
2. aggregates all entries per resource with the deterministic n-ary Decimal sum;
3. adds each aggregate and quantizes exactly once at the commit boundary;
4. validates all starting and prospective balances against state, minimum, and hardcap invariants;
5. mutates nothing if any validation fails;
6. otherwise commits all changes and returns them sorted by resource ID.

`Apply` rejects any result above a hardcap with `ErrAboveHardcap`; it never clamps purchases,
grants, migrations, or conversions. `ApplyAccrual` additionally rejects every negative entry. If a
positive aggregate would exceed a cap, it commits the declared cap exactly. Multiple resources
saturate independently inside the same all-or-nothing transaction.

A receipt contains canonical `before`, `delta`, and `after` strings. For positive accrual, `delta`
is the actual aggregate applied and is guaranteed to reproduce the authoritative result through
`Quantize12(before + delta) == after`. Saturation selects that delta by probing the quantized
headroom at nominal, −1, +1, −2, and +2 12-digit ulps. An unreproducible result is an invariant
failure and commits nothing. Resources unchanged by quantization—including further production at
an already reached cap—are omitted.

The n-ary sum groups same-exponent terms before normalization and orders full-precision mantissas
numerically. Transaction acceptance is therefore invariant under entry permutation, including
domain-edge cancellations; a genuinely out-of-domain net result still rejects atomically.

The distinction is deliberate: positive time accrual saturates at its authoritative ledger
boundary, while every ordinary hardcap overflow rejects. No above-cap state is committed or
exposed, and rewards cannot opt themselves into silent value loss.

## Query, command, and subscription responsibilities

“Query / mutation / subscription” is a domain separation, not a GraphQL commitment:

- callers query catalogs and ledger snapshots;
- server-owned commands are the only mutation path;
- successful receipts can later be persisted and published by the player/world/guild actor.

The kernel owns no sockets, callbacks, channels, goroutines, REST routes, or UI stores. A transport
failure therefore cannot roll back or partially apply economy state. The TypeScript implementation
loads definitions and predicts costs but has no authoritative balance-commit API.

## Verification

The shared fixture is [`testdata/economy-kernel.json`](../testdata/economy-kernel.json). It covers
strict rejection categories and cross-runtime constant, linear, geometric, multi-ratio, and huge
exponent queries.

Go ledger tests cover scope isolation, deterministic receipts, below-minimum and above-hardcap
rejection, non-finite/unknown inputs, numeric overflow, and all-or-nothing behavior. The accrual
suite includes the demonstrated one-ulp hardcap regression, already-capped and multi-resource
cases, and 2,000,000 deterministic near-cap inputs across exponent boundaries. Every emitted
accrual receipt delta must re-apply to its exact `after`. The aggregation regression separately
combines one million contributions individually too small to move a `1e100` bank; the aggregate
changes committed state while committing each contribution separately does not.

Run every gate from the repository root:

```sh
make verify
```

This runs Go tests/vet, strict TypeScript checking, Node tests, and the complete Chromium, Firefox,
and WebKit suites.

## Deliberate deferrals

Idempotency storage, production/time integration, offline progress, multiplier slots, purchases,
manual actions, and shipped Phase-0 balance now exist; see [Production engine](production-engine.md).
The kernel still does not own cross-scope coordination, WebSocket publication, client
interpolation, cap tooltips, leaderboard ordering, Compute Credit spending, or the large balance
harness. Each remains owned by a later bounded RFC.
