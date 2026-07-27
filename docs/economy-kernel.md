# Economy Kernel

The economy kernel is the implemented boundary between RFC-0001 large-number arithmetic and
future gameplay systems. It provides a strict shared catalog, matching Go/TypeScript cost queries,
and a scoped authoritative Go ledger. It contains no shipped balance or production mechanics yet.

## Catalog

The authoring schema is [`balance/economy.schema.json`](../balance/economy.schema.json). Runtime
validation is performed independently by `economy.LoadCatalog` on the server and `parseCatalog`
on the client. Both accept the same version-1 shape:

```json
{
  "schema_version": 1,
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
      }
    }
  ]
}
```

This example describes structure only. The repository does not yet contain a shipped game
catalog, generator price, or launch ratio.

Catalog rules:

- The only supported schema version is `1`.
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
after `owned`; geometric cost uses the RFC-0001 closed form. Max-affordable uses a bounded search
and guarantees that the returned count is affordable while the next count is not, unless the
exact-integer ceiling itself was reached.

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
- `Apply(Transaction)` as the only balance mutation;
- `Receipt` as the committed notification fact.

For every transaction, `Apply`:

1. validates resource existence, ledger scope, and every finite delta;
2. aggregates all entries per resource at full intermediate precision;
3. adds each aggregate and quantizes exactly once at the commit boundary;
4. validates all prospective balances against state, minimum, and hardcap invariants;
5. mutates nothing if any validation fails;
6. otherwise commits all changes and returns them sorted by resource ID.

A receipt contains canonical `before`, `delta`, and `after` strings. `delta` is the actual committed
difference after boundary quantization, not an uncommitted requested value. Resources unchanged by
quantization are omitted.

Hardcap overflow is rejected, never silently clamped. A future production engine may calculate
remaining headroom and accrue exactly to a cap, while purchases and conversions remain protected
from accidental value loss.

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
regression aggregates one million contributions individually too small to move a `1e100` bank;
the aggregate changes committed state while committing each contribution separately does not.

Run every gate from the repository root:

```sh
make verify
```

This runs Go tests/vet, strict TypeScript checking, Node tests, and the complete Chromium, Firefox,
and WebKit suites.

## Deliberate deferrals

The kernel does not yet implement save restoration, idempotency storage, production/time
integration, offline progress, multiplier stacks, purchases/actions, cross-scope coordination,
WebSocket publication, client interpolation, cap tooltips, leaderboard ordering, shipped balance,
or the balance harness. Each remains owned by a later bounded RFC.
