# RFC: Route Temporal Validity

- **Status:** implemented
- **Author:** Codex, implementing the owner's 2026-07-29 A7 ruling
- **Created:** 2026-07-29
- **Design refs:** `design/01-tiers.md`; `design/02-economy-balancing.md`; `design/08-satire-flavor.md §6`
- **Amends:** `archive/gate-predicates-and-routes.md`
- **Planning:** `planning/route-temporal-validity/`

## Summary

Reject route predicates that depend on a doctrine which cannot exist when their gate is crossed.
Repair the three Phase-0 seeds currently attached to `gate.t2_to_t3` despite depending on
`transition.t3_to_t4`.

## Specification

### D1 — Canonical chronology

Doctrine-bearing routes use canonical adjacent tier identifiers:
`gate.tN_to_tN+1` and `transition.tM_to_tM+1`. Every `doctrine_is` or `doctrine_is_not` condition
requires `N >= M`: the gate is crossed at or after the doctrine transition. An unparsable or
non-adjacent gate/transition on a doctrine-bearing route is invalid rather than unordered.

The Go loader, TypeScript loader, and repository schema-semantic command enforce the same rule.
Pure structure/resource/fact/meter/region routes retain mechanical gate IDs because they have no
temporal dependency.

### D2 — Phase-0 seed repair

`route.regulatory_capture`, `route.ipo_sequence_break`, and
`route.acquihire_out_of_bounds` move intact from `gate.t2_to_t3` to the already declared later
`gate.t4_to_t5`. Predicates, effects, names, activity flags, and balance numbers do not change.
The early standard gate remains declared with no routes.

This is a balance-catalog change. The three-family constants identity and pacing artifacts update
through a separate artifact-only `BALANCE-CHANGE:` commit.

## Acceptance criteria

1. Both runtime loaders and `make verify-schema` reject a shared route whose doctrine transition
   occurs after its gate.
2. A route on the same transition boundary and one on a later gate are accepted.
3. The three shipped doctrine routes resolve only under `gate.t4_to_t5`; requesting one at
   `gate.t2_to_t3` rejects as `route_predicate_unmet`.
4. The Depletion proof remains green after seed relocation.
5. Full verification and the separate baseline-regeneration protocol pass.

## Deviations from design

The original C1 example placed Regulatory Capture on `gate.t3_to_t4`. Phase-0 has no implemented
permit resource or standard T3→T4 gate, so the seed moves to the next already implemented gate
instead of inventing a resource or balance amount.

## Changelog

- 2026-07-29: owner ruled temporal impossibility invalid; accepted and implementation started.
- 2026-07-29: implemented across Go, TypeScript, and schema semantics; independent review
  approved the implementation. Same-boundary doctrine ordering remains assigned to the future
  doctrine-intent RFC before any such route may ship.
