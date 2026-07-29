# Route Temporal Validity — append-only log

## 2026-07-29 — start

- Owner ruled that a doctrine condition may bind only a gate crossed at or after its transition.
  Temporal impossibility is a catalog error because dead routes weaken the Depletion proof's
  effective route budget.
- The three affected seeds move intact to the existing later T4→T5 gate. Creating the RFC's
  illustrative T3→T4 permits gate would require an unimplemented resource and invented balance
  value, so it is explicitly not done here.

## 2026-07-29 — implementation

- Go and TypeScript parse adjacent tier boundaries and reject every doctrine condition whose gate
  source tier precedes its transition source tier. Unknown/non-adjacent chronology fails closed.
- The repository schema command carries the same semantic check, and the shared
  `temporal-impossibility.json` fixture must be rejected by both runtimes and CI. Same-boundary and
  later-boundary cases pass.
- The three doctrine seeds moved intact from T2→T3 to T4→T5. Intent tests now prove the active
  discount at its new gate and the old gate returns `route_predicate_unmet`; projection fixtures
  use the repaired event identity.
- Go routes/production/projection tests, 6,412 client tests, TypeScript diagnostics, and schema
  semantics are green. The Routes input commit must land before the artifact-only pacing update.
