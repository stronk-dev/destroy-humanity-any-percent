# Economy Kernel Architecture Research

> Researched 2026-07-28 for RFC-0002. This report concerns the reusable economy boundary,
> not launch tuning.

## Question

How much should the first economy layer abstract so that later resources, currencies,
generators, actions, banks, and transports can be added without baking balance choices into
code or creating an untestable second programming language?

## Evidence reviewed

### Mature incremental games separate data from runtime state—but not always cleanly

Antimatter Dimensions collects static definitions behind a `GameDatabase`, then maps those
definitions into runtime mechanic objects. Its `Currency` facade centralizes common operations
such as addition, subtraction, and purchasing. That is useful precedent for stable resource
access, but individual currency setters also perform unrelated record and notification side
effects. The lesson for a new server-authoritative game is to keep the primitive ledger free of
feature-specific side effects and return an explicit change receipt instead.

- [Antimatter Dimensions `GameDatabase`](https://github.com/IvarK/AntimatterDimensionsSourceCode/blob/master/src/core/secret-formula/game-database.js)
- [Antimatter Dimensions currency facade](https://github.com/IvarK/AntimatterDimensionsSourceCode/blob/master/src/core/currency.js)

The Modding Tree demonstrates why data-shaped content is attractive: a layer definition groups
identity, resources, requirements, and formulas in one place. It also allows arbitrary JavaScript
callbacks throughout that data. That flexibility is appropriate for a modding toolkit, but it
would defeat our Go/TypeScript parity tests, allow balance files to mutate runtime state, and turn
hot reload into code reload. We should copy the declarative shape, not executable callbacks.

- [The Modding Tree layer definition](https://github.com/Acamaeda/The-Modding-Tree/blob/master/js/layers.js)

### Economy subsystems benefit from query/request/event boundaries

The Unlimited Rulebook reference architecture identifies three services at an economy boundary:
queries about state, requests to change state, and notifications about changes. It also argues
for separating records from computations and keeping external systems from mutating the record
directly. That maps well to our server actor model without requiring its full dynamic rule engine.
The same paper notes the learning and up-front implementation cost of the general rule system;
our first kernel should therefore use typed commands and receipts rather than predicate dispatch
or a general formula DSL.

- [Unlimited Rulebook paper](https://www.ime.usp.br/~kazuo/media/ArtigoWilsonICSA2020.pdf)

Here, “query / mutation / subscription” describes domain responsibilities, not a commitment to
GraphQL. REST handlers, WebSocket actors, save code, simulations, and tests should all call the
same query and command ports. A successful command returns a receipt; a later actor or transport
may publish that receipt. The ledger itself must not own sockets, callbacks, or goroutines.

### Incremental simulation should materialize only at discontinuities

Swarm Simulator represents continuous production analytically and “reifies” immediately before
buying units, casting abilities, or other discontinuities. This reinforces the existing project
decision: production integration belongs in a later production engine, while the economy kernel
owns the atomic discontinuity where calculated gains or costs become committed state.

- [Math of Swarm Simulator](https://math2.swarmsim.com/)

### Strict data catches tuning mistakes earlier

JSON Schema 2020-12 distinguishes required properties from object-property validation and allows
unknown properties to be forbidden. Both runtime loaders should likewise reject missing fields,
unknown fields, duplicate IDs, dangling references, malformed canonical decimals, and unsupported
curve kinds. Silent defaults are especially dangerous for balance data because a misspelled field
still looks like a plausible game value.

- [JSON Schema 2020-12 specification](https://json-schema.org/specification)
- [JSON Schema object validation](https://json-schema.org/understanding-json-schema/reference/object)

## Selected architecture

The first kernel should abstract the parts already known to vary:

1. A versioned catalog defines resources and generator-class prices using stable mechanical IDs.
2. Resource values, curve parameters, and prices use RFC-0001 canonical decimal strings.
3. Cost curves use a closed tagged union: `constant`, `linear`, and `geometric` initially.
4. Generator counts remain exact integers; continuous balances remain `Decimal` values.
5. A server ledger is the only primitive balance mutator. It aggregates all entries per resource,
   quantizes once, validates every result, then commits all-or-nothing.
6. Queries return definitions, balances, and snapshots without mutable references.
7. Mutations return deterministic receipts. Subscription delivery is an adapter concern later.
8. The TypeScript side parses the same catalog and evaluates the same cost curves for prediction;
   it does not gain authority to commit balances.

## Explicit non-goals

- No arbitrary expressions, JavaScript callbacks, embedded scripting, or general rule engine in
  balance data.
- No universal “entity/component” model before concrete mechanics require it.
- No final `1.10`, `1.13`, cap, price, or production-rate decision in engine code.
- No production integration, offline simulation, persistence, networking, or UI subscription
  implementation in this RFC.
- No event callback registry inside the ledger. Receipts are safer to test and can be published by
  the future player actor without coupling domain state to a transport.

This is intentionally an extensible seam, not a promise that every future mechanic can be added
without writing code. New formula shapes require a follow-up RFC, cross-runtime implementation,
and shared parity tests.
