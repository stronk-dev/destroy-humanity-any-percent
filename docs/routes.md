# Gate Predicates and Route Registry

Routes are declarative alternate preconditions for gates. They can reduce a gate's resource
requirement or substitute for it; they cannot change a production rate, multiplier contribution,
or slot. The pure Go evaluator lives in `server/routes`, the TypeScript mirror in
`client/src/routes.ts`, and `make verify-routes-boundary` fails if the Go package acquires a
transitive dependency on `server/production`.

## Catalog and predicate context

[`balance/routes/phase0.json`](../balance/routes/phase0.json) is strict routes catalog schema v1.
It declares the supported predicate-context version, Route Knowledge grants and hint cost,
Depletion's distinct-route count, gate requirements, route activation, exclusion assignments,
predicates, and effects. JSON Schema is published at
[`balance/routes.schema.json`](../balance/routes.schema.json); both runtime loaders additionally
enforce canonical Decimal strings, unique IDs, active context availability, effect ranges, and the
Depletion proof. Bundle validation also rejects a gate requirement or resource condition whose ID
does not resolve to a company resource in the economy catalog.

The closed predicate union is:

- `resource_at_least` and `resource_at_most`, against canonical committed Decimal balances;
- `meter_band`, with inclusive integer boundaries from 0 through 100;
- `doctrine_is` and `doctrine_is_not`, keyed by transition;
- `structure_is`;
- `ledger_fact_present`;
- `region_trait`.

A context is an immutable snapshot of resources, doctrines, structure, ledger facts, meter bands,
and region traits plus `context_version`. Missing values do not silently acquire defaults: a
condition is unmet, and an active route may not require a context version newer than the build.
Shared fixtures in `testdata/routes/predicate-vectors.json` exercise every condition kind and both
meter boundaries in Go and TypeScript. Client evaluation is disclosure/presentation only; the
server result is authoritative.

The shipped catalog contains all seven house-seeded routes. `Nonprofit Wrapper Zip`,
`IPO Sequence Break`, and `Acquihire Out-of-Bounds` are active at context v1. The other four are
declared inactive, context-v2 conjectures until constituency meters and region traits exist.

## Gate crossing

`cross_gate` carries a mechanical `gate_id` and a nullable `route_id`. In one authoritative
transition it:

1. evaluates elapsed production;
2. rejects an already-crossed gate;
3. resolves the standard requirement or checks the selected active route predicate;
4. debits every required resource atomically through the company ledger;
5. records `gates_crossed[gate_id]` in save v5;
6. emits one `gate_crossed`, and for a route path one `route_executed`, on the new revision.

A discount debits `Quantize12(requirement × fraction)`. A substitute debits nothing: satisfying
its predicate is the price. Typed terminal rejections are `gate_already_crossed`,
`requirement_not_met`, and `route_predicate_unmet`; malformed/unknown identifiers retain the
production API's `invalid`/`unknown_id` behavior. Intent replay returns the original receipt and
events without crossing or projecting twice.

Company save v5 also carries `run_seq` and the committed context fields. A run identity is the
tuple `(company_stream_id, run_seq)`. The Exit system owns incrementing `run_seq`; this RFC starts
legacy and new company state at 1.

## Registry and founder projections

The company event stream is the source of truth. After commit—and again on replay after a lost
response—the route projector consumes event records in `(occurred_at, event_id)` byte order.
Projection writes are idempotent by event ID, and `(founder, company stream, run, route)` is unique,
so a route cannot be farmed repeatedly in one run. Registry decisions are serialized per route and
compare this immutable order across separate deliveries; whichever batch happens to arrive first
does not define the winner.

The earliest event receives permanent executor credit and a 72-hour naming reservation. If a
still-earlier event is delivered later, it replaces the provisional winner and receives a fresh
reservation from its own occurrence time. Any provisional name is reset to the catalog house name:
it could not exist in an event-order rebuild. A submitted name otherwise enters `pending`;
approval publishes it, failure restores the house name, and an unused reservation expires to that
same name. The Registry stores a plain execution count; variants, time buckets, adoption curves,
and its public read API belong to Registry Analytics.

Founder career executions and Route Knowledge are idempotent projections. Grants are:

- registry first: 100, in addition to the founder-first grant;
- that founder's first execution of the route: 25;
- a later run by the same founder: 5.

All values are catalog data. Each derived grant is an immutable `route_knowledge_granted` event.
When event ordering replaces a provisional Registry winner, the projector appends a standard
`compensation` for that exact grant and awards the true winner from its own catalog epoch. A
correction consumes cached spendable Knowledge first; any amount already spent becomes
non-spendable projection debt that later grants repay before increasing the balance. Purchased
hints remain permanent and founder saves never become negative. Read-repair rebuilds the same
balance and debt from uncompensated grants minus purchases. `buy_route_hint` runs on the Founder
stream, costs the catalog's 50 Route Knowledge, and permanently adds the route ID to
`hints_unlocked`. Hints reveal catalog predicates only; the predicate context and evaluator do not
contain hint state.

## Depletion proof and verification

Every route declares one immutable exclusion slot/value (`structure` or a doctrine transition),
and both loaders require that declaration to match an explicit `structure_is` or `doctrine_is`
condition in the executable predicate.
Both loaders exhaustively enumerate all declared slot assignments and calculate the exact maximum
route subset possible in one run. The shipped maximum is 4 and Depletion requires 5. A catalog
with maximum greater than or equal to N fails loading and CI; the shared negative fixture proves
the gate fails closed.

`make verify` covers schema semantics, cross-runtime predicate parity, import boundaries, save-v5
migration, gate debit/event behavior, hint non-interference, and Postgres integration. The database
suite races first executions, reverses delivery order across batches, exercises equal-time UUID
ordering and an already-spent provisional award, rebuilds the cache from immutable history, and
asserts one active Registry grant with no duplicated adoption.
