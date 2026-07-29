# Gate Predicates & Route Registry — append-only log

## 2026-07-29 — acceptance and start

- Re-read RFC-0000, AGENTS.md, the complete Gate Predicates RFC, its design references, and the active index.
- Owner assignment to implement the RFC queue is explicit acceptance. C1–C8 answer the previous review blockers; provisional numeric values are correctly isolated as balance data and Registry Analytics is a named follow-up.
- Marked the RFC `implementing`. No push will be performed.

## 2026-07-29 — pure catalog, evaluator, and proof boundary

- Added the standalone Go `routes` package and TypeScript mirror. Both strict-load the versioned catalog, implement the closed condition union, inclusive meter bands, canonical Decimal comparisons, discount quantization, and substitute semantics.
- Shipped the seven named seed routes: three context-v1 active routes and four context-v2 conjectured routes. Grant values, hint cost, requirements, activation, and Depletion N are data.
- Implemented the exact exclusion-slot search. The shipped maximum is 4 routes per run against N=5; shared invalid fixtures prove a reachable catalog is rejected.
- Added shared predicate vectors for every condition kind and boundary behavior. Go and TypeScript suites pass all vectors.
- Added JSON Schema validation and a build gate that rejects any transitive `routes -> production` dependency.
- Batch review: checked strict unknown-field handling, canonical numeric parsing, raw-byte deterministic ordering, inactive-context behavior, exact proof enumeration, and client/server edge parity. No unresolved finding; the local Go cache path in the boundary target was made explicit so the check is hermetic under restricted runners.

## 2026-07-29 — save v5, authoritative intents, and projections

- Added save v5 with company `gates_crossed`, `run_seq`, and the committed predicate-context fields; founder Route Knowledge and permanent hint state; strict scope validation; deterministic set encoding; and an updated migration corpus from v1 onward.
- Added `cross_gate` and `buy_route_hint` to the canonical intent parser/hash path. Gate crossing accrues first, checks the selected standard/route path, debits requirements atomically through the ledger, sets one-shot gate state, and emits validated `gate_crossed`/`route_executed` events. Discount and substitute tests cover both debit semantics and typed rejections.
- Event records now return from first application and replay, allowing the same post-commit projection path to run after a lost response. The projector is idempotent by event ID and `(founder, company, run, route)` uniqueness.
- Added the Registry/founder projection migration, insert-if-absent first-executor resolution, 72-hour naming reservation and moderation lifecycle, permanent executor credit, exact registry/founder/repeat Route Knowledge grants, founder read repair, execution counts, and hint-cost projection.
- A real-Postgres concurrency test races two company streams, proves one registry winner and two adoptions, replays both events without double counting, exercises hint repair, and completes the naming flow. The production integration test proves a route crossing replay emits no duplicate and a founder can immediately spend the projected grant.
- Review finding fixed before commit: route projection tables initially survived test `TRUNCATE` because the balance table intentionally has no save-stream FK. Integration setup now names every projection table, preventing state leakage between packages. Hint intents also skip company contribution providers.
- Verification: `go test ./...`, `go vet ./...`, and the full Postgres integration target pass.

## 2026-07-29 — canonical docs and harness artifact review

- Added `docs/routes.md` and updated the canonical production/save docs and repository status for the four-intent surface, save v5, route events, replay delivery, and projection tables.
- The first full `make verify` correctly rejected the harness golden: save v5 changes the encoded terminal state even though pacing is unchanged. Regeneration changed only the two deterministic `final_state_hash` values; `pacing-baseline.json`, milestones, elapsed times, outcomes, and invariants are byte-identical. Accepted as a state-envelope artifact update, not a balance change.
- Full acceptance rerun with Postgres passed: Go/vet/integration, harness golden and pacing baseline, strict TypeScript with 6,388 Node tests, routes/economy/harness schemas, and 19,164 tests across Chromium, Firefox, and WebKit.
