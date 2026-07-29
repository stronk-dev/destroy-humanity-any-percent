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

## 2026-07-29 — verification and local review

- The artifact-only `BALANCE-CHANGE:` commit records the relocated route catalog. The constants
  identity changed; pacing metrics did not.
- Full `make verify`, including real-Postgres integration, the balance guard, 6,412 client tests,
  and 19,245 browser cases, passed.
- A local spec/adversarial pass found no additional route defect. Independent review remains the
  mandatory archival gate.

## 2026-07-29 (claude — independent review of 0df79d9..baf4d43: APPROVED)

Both rulings implemented faithfully; parity verified in Go + TS + schema; suites green.

- **Chronology:** `gateTier < transitionTier → reject`, identical integer rule in both runtimes,
  with non-canonical gate/transition IDs on doctrine-bearing routes refused outright. The three
  seeds moved to `gate.t4_to_t5` — strictly after their doctrine's transition, sidestepping the
  equality case. **Routed forward, not a blocker: the `gateTier == transitionTier` binding
  (allowed by this rule, and C1's original example) depends on whether the doctrine pick
  evaluates before the same crossing's gate predicate — pin that ordering in the future
  doctrine-intent RFC before any same-gate route ships.** The TS loader's
  `maxRoutesPerRun() >= depletion → throw` confirms the proof still binds after the moves.
- **Merges:** floor-triggered only, whole-cohort moves across assignments/memberships/samples in
  one transaction, source zeroed and closed, cap enforced. `floor(1.5×target)` vs the ruling's
  ceiling: accepted deviation — conservative direction, differs only on odd halves.
- **baf4d43:** the merge-cap overflow was again self-found by Codex's adversarial pass before
  reaching this gate.
- The `BALANCE-CHANGE:` artifact regeneration flowed through the hardened guard: constants
  identity moved, pacing metrics unchanged — exactly the shape the guard exists to certify.

Both ruling RFCs clear to archive.
