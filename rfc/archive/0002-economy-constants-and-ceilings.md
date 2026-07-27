# RFC-0002: Economy Kernel

- **Status:** implemented
- **Author:** Marco (initial draft by Claude; re-scoped by Codex)
- **Created:** 2026-07-27
- **Revised:** 2026-07-28
- **Design refs:** `design/02-economy-balancing.md §1, §2.1–2.2`, `design/00-vision.md`
  (hardcaps), `design/06-tech.md §idle-math`, `design/07-roadmap.md`
- **Research:** `design/research/economy-kernel.md`, `design/research/numeric-core.md`,
  `design/research/pacing-science.md`
- **Depends on:** RFC-0001 and Numeric Core Boundary Hardening (implemented)
- **Supersedes / superseded by:** —
- **Planning:** `planning/archive/0002-economy-kernel/`

## Summary

RFC-0001 defines how enormous continuous values are calculated and committed. This RFC defines
the first reusable economy boundary around those values: a strict data catalog, cross-runtime
cost curves, and an atomic server ledger.

The engine does not choose whether a generator uses `1.10`, `1.13`, or another supported value.
It evaluates the curve declared for that generator class. Launch tuning remains balance data and
is gated by the future balance harness.

## Scope

**In scope:**

- a versioned JSON catalog for resources and generator-class prices;
- strict Go and TypeScript catalog loaders;
- `constant`, `linear`, and `geometric` cost curves with bulk-cost and max-affordable queries;
- an authoritative Go ledger with read queries and atomic transaction commands;
- deterministic mutation receipts for later publication by actors/transports;
- one-quantization-boundary accrual semantics and cap/minimum enforcement;
- shared cross-runtime fixtures and parity tests.

**Out of scope:**

- shipping game resources, generators, prices, ratios, or cap amounts;
- production sources, multiplier stacks, time integration, and offline progress;
- save persistence, migrations, idempotency storage, actors, REST, WebSockets, or GraphQL;
- client animation/interpolation and player-facing cap rendering;
- leaderboards and ranking keys;
- upgrades, unlock predicates, conversions, prestige, and arbitrary formula/rule languages;
- hot-reload file watching and production `go:embed` wiring.

Those systems consume this kernel through later RFCs. Test fixtures are examples, not shipped
balance.

## Decisions

### K1 — Catalog shape is shared, versioned, and strict

The catalog root has exactly these fields:

```json
{
  "schema_version": 1,
  "resources": [],
  "generator_classes": []
}
```

Resource definition:

```json
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
```

`hardcap` must be explicitly either the object above or `null`. A declared cap must exceed or
equal the minimum and include a non-empty localization key. Technical RFC-0001 limits are never
game-visible caps. A later production definition that can grow indefinitely must reference a
capped resource; that graph validation belongs to the production RFC because this catalog alone
cannot infer how a resource is produced.

Generator-class price definition:

```json
{
  "id": "generator.example",
  "price": {
    "resource_id": "company.cash",
    "base": "1e2",
    "curve": { "kind": "geometric", "ratio": "1.13e0" }
  }
}
```

All identifiers are stable mechanical strings. Every decimal is an RFC-0001 canonical string.
The catalog rejects unknown fields, unsupported versions/kinds/scopes, invalid identifiers,
duplicate IDs, dangling resource references, non-state decimals, invalid bounds, and invalid
curve parameters.

The initial supported scopes are `company`, `founder`, `world`, and `guild`. `numeric_kind` is
explicit and currently only accepts `decimal`; exact owned counts remain safe integers outside
the ledger. Adding another numeric kind or scope changes behavior and requires a follow-up RFC.

### K2 — Curve types are code; parameters are data

Supported cost curves are a closed tagged union:

- `constant`: `price(n) = base`
- `linear`: `price(n) = base + step × n`, with `step ≥ 0`
- `geometric`: `price(n) = base × ratio^n`, with `ratio ≥ 1`

`base` must be positive. Owned and purchase counts are exact non-negative integers no greater than
RFC-0001's `MaxExactInteger`. Bulk cost is the sum of the next `count` prices after `owned`.
Max-affordable must satisfy the exact postcondition that `n` is affordable and `n+1` is not,
unless `n == MaxExactInteger`.

Go and TypeScript evaluate the same tagged definitions. There is no global ratio and no launch
default in source code. A new curve kind requires a follow-up RFC and shared parity fixtures; data
cannot contain executable formulas or callbacks.

### K3 — The ledger is the authoritative mutation boundary

A ledger is constructed from a catalog for exactly one declared scope and begins with every
resource in that scope at its declared `initial` value. Company, Founder, World, and Guild
balances therefore live in separate ledgers owned by their respective future actors. A primitive
transaction cannot cross scopes; later cross-scope actions require an explicit coordinating
command that validates all participating ledgers. Its public responsibilities are:

- **queries:** look up definitions, read one balance, or take a canonical-string snapshot;
- **commands:** apply a transaction containing one or more `(resource_id, delta)` entries;
- **events:** return a deterministic receipt containing changed resource IDs and canonical before,
  net-delta, and after values.

There is no public direct balance setter. Loading persisted balances is deferred to the save-layer
RFC and must enter through a separately validated constructor or restore path.

`Apply` performs these steps:

1. validate all entries, resource references, and the ledger scope;
2. sum all entries for each resource at full intermediate precision;
3. add each net delta to its prior balance and quantize exactly once to 12 significant digits;
4. validate every prospective result against finite-state, minimum, and hardcap invariants;
5. if any result fails, return an error and mutate nothing;
6. otherwise commit every result atomically and return changes sorted by resource ID.

The ledger rejects cap overflow; it does not silently clamp. A future production engine may
explicitly calculate headroom and accrue exactly to a cap, but conversions and purchases must
never accidentally destroy value because an output was silently truncated.

### K4 — Receipts are the subscription seam

The kernel does not implement callback registration, sockets, channels, or GraphQL. The successful
mutation receipt is the sole notification fact. The future player actor may persist it, append an
audit event, and publish it to subscribers after commit. This keeps transaction correctness
independent of delivery failures and prevents UI or network code from mutating the ledger.

The TypeScript client may query the shared catalog, quote costs, and apply authoritative snapshots
or receipts in a future client RFC. It does not receive a balance-commit API in this RFC.

### K5 — Caps are per resource; tuning remains provisional

There is no global gameplay ceiling. Each resource explicitly declares a hardcap or `null`; future
systems must justify indefinitely growing uncapped resources during their own RFC review. Numeric
overflow is always an invariant violation, never an ending or a visible wall.

`1.13` remains the current design and research baseline only. `1.10`, `1.13`, and any other valid
ratio are ordinary data values. The balance harness decides shipped values per generator class.

## Deviations from the previous draft and design

- The previous draft combined tuning, production, client rendering, and leaderboard policy. Those
  are split out because they cannot be implemented or verified in this bounded kernel.
- `design/02 §2.1` previously presented `r = 1.13` as a single global constant. It is now a
  provisional per-generator-class balance value, not an engine constant.
- The previous draft proposed `1.10` as a launch default. This RFC deliberately chooses no launch
  value before the balance harness exists.
- There is no global gameplay ceiling. Hardcaps are explicit per-resource data.

## Acceptance criteria

1. One shared catalog fixture is accepted by strict Go and TypeScript loaders; malformed fixtures
   cover every required rejection class, including missing/unknown fields and dangling IDs.
2. Shared fixtures prove bulk-cost and max-affordable parity for constant, linear, and geometric
   curves, including at least two different geometric ratios and huge exponents.
3. No Go or TypeScript engine source contains a global generator ratio or launch tuning default.
4. The Go ledger exposes queries plus atomic `Apply`; tests prove insufficient funds/minimum,
   hardcap overflow, invalid numeric results, unknown resources, deterministic receipt ordering,
   and all-or-nothing multi-resource failure.
5. A regression test aggregates at least one million sub-resolution source entries before the
   commit boundary and moves state, while the per-entry-commit negative control does not.
6. `go test ./...`, `go vet ./...`, strict TypeScript typecheck, Node tests, and Chromium/Firefox/
   WebKit suites pass.
7. Canonical documentation describes the catalog, curve contract, ledger boundary, receipts, and
   deliberate deferrals; the implemented RFC and planning directory are archived.

## Changelog

- 2026-07-27: created as a draft covering economy constants, ceilings, accrual, presentation, and
  leaderboard policy.
- 2026-07-28: re-scoped after architecture research to the implementable economy kernel; tuning,
  production, presentation, and ranking concerns deferred; accepted by owner direction to make
  the engine configuration-driven and proceed with implementation.
- 2026-07-28: implemented, verified across Go, Node, Chromium, Firefox, and WebKit, documented,
  and archived.
