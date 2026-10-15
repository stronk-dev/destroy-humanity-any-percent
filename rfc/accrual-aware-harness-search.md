# RFC: Accrual-aware Harness Search

- **Status:** draft — engineering closure and owner acceptance required; NOT implementation authority
- **Author:** Codex
- **Created:** 2026-10-15
- **Design refs:** `design/02-economy-balancing.md` §11/§11b; `design/06-tech.md` lazy authoritative math
- **Depends on:** [Production Accrual Conservation](production-accrual-conservation.md) C1b–C1e; existing harness foundations
- **Parent / amends:** [Relevance Harness](archive/relevance-harness.md) R10/R11; affected first-hour/content-dynamics consumers
- **Planning:** existing `planning/clout-v1-and-pr-interns/` evidence; dedicated implementation plan after acceptance
- **Split from:** Production Accrual Conservation C1c search reuse/C1f, under RFC-0000 scope discipline

## Summary and scope

Specify how the actual harness consumes retained calculation origins without
turning observations into settlements, merging future-inequivalent branches or
missing affordable windows at endpoint-context changes. This separates search
policy from the producer's private calculation/context/event contract. It does
not shrink Clout AC6, the full accounting repair or the nine-tier1.0 floor.

Both drafts require independent normative review and owner acceptance before their
implementation. Neither draft accepts the other. Primitive accounting work cannot
claim full AC6/harness closeout while affected consumers remain unmigrated/unproved.
Current scenario guards, reports, budgets, scores and tie-breaks stay authoritative
until an explicit accepted migration changes them.

## Specification proposal

### H1 — Real caller bindings

Use the parent's begin/fork/project/action contract, not a second arithmetic engine.

| Actual current caller | Required migration |
|---|---|
| `RelevanceSuite.advanceMode`, earliest-affordable probes and `finishToMilestone` | Project the retained origin at the query endpoint; do not start each probe from a previously rounded view. |
| `rankCandidate`, bootstrap/manual/gate actions | Execute the real semantic action from its retained pre-action origin. Do not pre-advance and thereby omit its accrual hooks. |
| Beam children, rollouts and clone helpers | Fork complete origin/context ownership, not merely the visible `save.State`; mutations and role collectors cannot leak to siblings. |
| `firstHourRuntime.step`, bank waits and reference-candidate adoption | Preserve origin across observations, actual session chronology and the lifetime adapter; applied actions/real policy boundaries close the old interval. |
| `contentDynamicsAccrueWindow` | Use owned projections with the parent's C1g diagnostic owner, preserving its window-start contribution sampling and complete original interval, not logged endpoint sampling. |

Parent C1g now proposes the last caller's explicit diagnostic-owner adapter; it must
not fabricate a harness scenario/policy identity to satisfy a constructor. Review
that binding and prove its actual factory/consumer during implementation. Preserve
old legacy-only constructors and the independent DG-D/D-021 axis holds; introducing
an origin is not permission to enable unsupported mechanics or remove their guards.

### H2 — All search reuse binds accounting identity

The parent owns canonical origin/context identity, state versions and private
events. This RFC owns reuse of those trusted identities:

- Beam equality includes origin/context identity, complete visible-state identity
  and virtual time. Path remains only the existing raw-byte tie-break.
- Ranked-completion caches also bind current revision and remaining decision
  allowance. Updating only beam deduplication leaves another unsafe reuse path.
- Dominance requires established future equivalence of accounting contexts before
  componentwise comparison. Better visible balances alone cannot justify pruning
  a branch with a different retained origin.
- Missing or unverified identity cannot fall back to visible-state-only reuse.
  Refuse the unsupported context; do not label that refusal an unreachable item.

RP-653 proves the relevant candidate constraint: two complete equal visible heads
at the same time can have different later payouts under different origins. This
does not show a current origin-free beam bug. Preserve existing ranking equations,
beam width, declared decision horizons and coverage. Any changed search population
must be measured under the adopted protocol, not hidden by relabelling old reports.

### H3 — Temporal affordability search (proposed)

An endpoint-selected buff can make a later origin-relative projection smaller.
Do not carry the current whole-horizon monotone-search assumption into the new
projection API. RP-656's bounded reference uses real schedule/application/Evaluate
at every integer endpoint0..1501ms: twelve online/offline profiles,18024 complete
evaluations, with every cash amount checked against independent literal fixture
factors. It covers production/building/combined expiry, ordinary reachability,
already-affordable time0 and unreachable targets. Six buffed profiles have early
qualifying windows but an unaffordable horizon. Unsplit monotone search misses them;
the boundary-aware reference matches the exhaustive earliest-time oracle exactly.
Targets are hypothetical fixed cash thresholds, not actual catalog purchases or
retuned prices; the reference does not execute the full hook-bearing parent C1c projector.
Original prospective ten-profile evidence and the two later declared unreachable
controls retain their separate provenance in the owning plan/log.

Proposed algorithm for the real caller:

1. Enumerate all relevant semantic endpoint-context boundaries from the actual pinned
   producer/schedule/absence inputs over the requested horizon. The reference knows
   only its two fixture expiry offsets1000/1500ms; it is not a production enumerator
   or proof of completeness for Guild, attendance, provision or future content.
2. Partition the canonical integer-ms endpoint domain into chronological regimes.
   A context change at `b` belongs to the right regime; the left ends at `b-1`.
   These are **search ranges**, not payout segments, settlements or origin resets.
   Every predicate still reconstructs the complete origin-to-endpoint projection
   with the correct context and original offline allowance.
3. Establish monotonicity of the actual affordability predicate inside each regime.
   A constant label, a single end sample or sampled increasing values are not a
   whole-domain proof. Account for numeric quantization, caps, provision, burst,
   hook-owned state and the actual resource/quote contract. If that proof does not
   hold, specify a correct alternative; do not skip the interval or declare no result.
4. Search regimes in order. Test the first endpoint, including the original time0.
   Only in a proven monotone regime may an unaffordable last endpoint exclude that
   regime. Otherwise use integer lower-bound search within that regime. The first
   qualifying result is the earliest over all earlier exhausted regimes.
5. Return a tagged unreachable result only after every required regime is exhausted.
   A valid solution at0 is not absence. Charge actual boundary work and predicate
   evaluations to the existing counters; exhausted work/size bounds are invalid
   measurements, not an unreachable item or an excluded population.

The reference finds production497ms, building1422ms, combined452ms and unbuffed869ms
in both modes, plus the two already-affordable0ms and two unreachable controls.
All known regimes are exhaustively checked monotone in this bounded population;
removing the expiry cuts gives the six missed-window controls. No origin is consumed.
This is a Go-only algorithm feasibility result over observed endpoint values, not
execution of `RelevanceSuite.earliestAffordable`, a migrated harness, an adopted
search-policy amendment or proof that current reports are incorrect. The current
harness does not use this proposed temporal projection path. Full real-caller
boundary enumeration, hook/dynamic-quote predicate proof, work measurements and R10/R11 normative
review remain required; a per-ms exhaustive production scan is not authorized.


### H4 — Compatibility, work and evidence

Keep old pinned scenarios/reports reproducible through their existing semantics.
An accepted new calculation/search protocol needs an explicit scenario/policy
version and caller dispatch, not an environment variable that silently changes
the same report identity. Do not assign those versions or regenerate artifacts
before the contracts and landing order are fixed.

Count all actual boundary, predicate, child and rollout work. Existing CI and
search bounds are not raised by this draft. If a correct candidate exceeds them,
retain the failed measurement and solve the engineering issue; do not narrow the
population, coast to a result, drop branches or restamp a golden. Production
per-player ticks or per-ms scans are not authorized by the exhaustive test oracle.

## Implementation acceptance criteria

1. The original128 partition profiles retain their exact final-state comparisons
   through the accepted calculation boundary; real caller journeys exercise it.
   Historical raw-evaluator failure evidence and relevant negative controls remain.
2. RP-653's equal-head/different-origin pair cannot collide in beam or completion
   reuse. Sibling mutation, ablation/context mismatch and invalid dominance controls
   discriminate; unchanged legitimate contexts still reuse correctly.
3. RP-656's complete twelve-profile/18024-endpoint population passes; the real
   boundary enumerator/predicate algorithm has its own completeness and failure
   proof over every admitted current producer family, not only two fixture edges.
4. Real first-hour/relevance/content-dynamics populations preserve complete hook,
   state, receipt/event and policy semantics. No lost/duplicated lifetime work,
   extra saved payment or event delivery from a discarded view.
5. Actual complete pacing/relevance/beam workloads and declared counters/bounds are
   verified; any expected report changes are produced honestly under the adopted
   protocol with exact source/content identity. No marketing-sized test counts.
6. Parent and consumer docs/plans, full reviewed implementation ranges and affected
   CI reconcile before feature closeout. This is not a release or1.0 approval.

## Open engineering and authority gates

Before RFC acceptance, review parent C1g's diagnostic-owner binding and complete the
actual temporal-boundary algorithm/predicate justification and exact versioned caller/report
dispatch. The fixture reference is not a substitute. These are engineering gaps,
not owner choices to approve away.

Owner adoption of the approach and designated normative review remain separate
from the executable implementation acceptance above. Do not require an unbuilt
production codec or migrated SQL workflow as proof this draft is implemented;
equally, do not claim those tests passed because the design was accepted.

## Deviations from existing contracts

Proposes origin-aware reuse and boundary-aware temporal search as explicit follow-ups
to archived Relevance R10/R11. Does not edit the archived body or silently change its
current implementation. No new balance values, caps, prices, grants, proration,
numeric tolerance, authored content or narrowed1.0 scope.

## Changelog

- 2026-10-15: split producer versus harness-search authority; retain RP-653/RP-656
  evidence, complete consumer obligations and unapplied status.
