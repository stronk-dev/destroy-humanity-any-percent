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
