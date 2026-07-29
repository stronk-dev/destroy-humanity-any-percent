# Gate Predicates & Route Registry — implementation plan

- **Assignee:** Codex
- **RFC:** `rfc/archive/gate-predicates-and-routes.md`
- **Started:** 2026-07-29

## Work breakdown

1. Add the strict routes catalog family, Go/TypeScript loaders, closed predicate DTOs, shared parity fixtures, and exact Depletion reachability proof.
2. Add save v5 company gate/run state and migrations, plus `cross_gate` and `buy_route_hint` intent schemas and rejection/event contracts.
3. Implement the production-to-routes boundary: pure predicate/gate evaluation, ledger debit, immutable gate/route events, and compile-time import prohibition.
4. Add founder and registry projection persistence with at-least-once idempotency, read repair, global first-executor race semantics, naming reservation, execution counters, Route Knowledge grants, and hint purchases.
5. Add tests for route parity, discount/substitute crossing, replay, projection concurrency, proof rejection, and hint non-interference; wire schema/proof gates into CI.
6. Update canonical docs, run the complete verification suite, record the per-change review, then archive the RFC and planning record.

## Completion

All six steps and all six RFC acceptance criteria completed on 2026-07-29.

## Acceptance gates

- Shared Go/TypeScript predicate fixtures agree for every active condition kind and band edge.
- Discount and substitute crossings emit one revision-tied `route_executed`; intent replay emits no duplicate.
- `server/routes` has no production import and CI enforces that boundary.
- Exact route-set proof passes the shipped catalog and rejects a single-run-reachable fixture.
- Concurrent projectors select one registry first executor and replay cannot double count.
- Route hints affect disclosure state only, never predicate evaluation.
- `make verify` passes with Postgres integration enabled.
