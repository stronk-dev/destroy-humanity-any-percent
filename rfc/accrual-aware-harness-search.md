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

The parent owns typed immutable origin/context state, opaque run-local identity,
source versions and private events. This RFC owns reuse of those identities:

- Beam equality includes origin/context identity, complete visible-state identity
  and virtual time. Path remains only the existing raw-byte tie-break.
- Ranked-completion caches also bind current revision and remaining decision
  allowance. Updating only beam deduplication leaves another unsafe reuse path.
- Dominance requires established future equivalence of accounting contexts before
  componentwise comparison. Better visible balances alone cannot justify pruning
  a branch with a different retained origin.
- Missing or unverified identity cannot fall back to visible-state-only reuse.
  Refuse the unsupported context; do not label that refusal an unreachable item.

Forks of one origin reuse its opaque identity. Distinct origins remain distinct
unless full future equivalence is separately proved. No origin JSON/hash protocol,
cross-run cache or identity field in reports is required. Opaque allocation order
must not affect ranking, tie-breaks or serialized output; identity participates in
equality only. Conservative separation still needs complete workload/budget proof,
not an allowance to exceed bounds or drop candidates.

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
Proposed protocol tags are exactly `legacy_mutating_wait_v1` and
`origin_projection_v1`. Legacy schemas retain their existing field set and dispatch
only to the legacy implementation: absence of a new field is not permission to
select modern behavior. New scenario schemas require `calculation_protocol` equal
to `origin_projection_v1` plus a required `calculation_profile`; new reports carry
the same protocol field and bind the
actual new scenario hash. Reject unknown/missing tags, new fields in legacy schemas
and any scenario/report protocol mismatch. No environment variable or warm cache
can change the protocol of the same declared input.

Exact new profile tags are `simulation_fixed_external_v1`,
`simulation_active_play_endpoint_v1`, and `content_dynamics_v1`. Relevance/first-hour
families accept only the first two, mapped by their factory to parent C1h's matching
simulation recipe. The endpoint profile cannot bypass current axis/attendance or
caller-support holds. Content dynamics accepts only the third: its active-play
window uses C1g, while existing Fiscal/Pitch/permit diagnostic methods retain their
own unchanged contracts. This profile does not claim those methods are active-play
calculations. The actual scenario hash binds the profile; reports cannot override
it and callers cannot choose a different private recipe at query time. Unknown or
wrong-family profiles refuse, never default to whichever implementation is available.

Observed source at `57906e5d` is the allocation baseline, not a reservation:

| Family | Existing formats and proposed dispatch |
|---|---|
| Relevance | Scenarios/policies1–2 remain legacy; current report writer6 remains legacy and existing readers retain their historical versions. The new scenario/report each take their own next-free schema, not an assumed shared number. |
| First hour | Existing scenario/experiment report1 and their ratified identities remain legacy. New schema/protocol gets its own scenario identity and report format; the old ratified SHA cannot authorize changed bytes. |
| Content dynamics | Existing scenario/report1 and immutable snapshots/goldens remain legacy. New protocol uses separately identified scenario/report bytes and the C1g window-start recipe, not generic endpoint sampling. |

The new Relevance scenario also requires `base_schema_version` equal to1 or2.
Validate its existing mechanics/segments using that explicit prior grammar and
require the loaded policy's schema to match that base, not the new outer scenario
schema. Current code couples scenario/policy schema numbers; adding accounting
metadata must not force a fake policy/balance mint or silently weaken its validation.
Reuse existing field validators without rewriting the source bytes to masquerade
as another version. No new scoring rule, window, policy value or budget is implied.

Allocate exact next-free family numbers and required ratified hashes in the reviewed
landing plan, checking contemporaneous source/lane ownership. These are not save,
replay-input or kernel versions. Existing serialized RunKeys already bind scenario
hash; keep their shape if unchanged rather than add another duplicate authority.
Current new-protocol inputs must actually execute the new path; a legacy green
report is not its acceptance evidence. Any first-hour hash adoption gate still
requires its actual authority, not a recorder-authored replacement constant.

For content registry migration, proposed registry schema2 retains the existing exact
entry fields. Scenario and report paths remain uniquely entry-owned. An immutable
bundle manifest becomes a registry-owned resource keyed once by its exact path;
multiple entries may reference it only when their epoch/seed coordinates agree and
the same validated manifest bytes/hash satisfy every reference. Generate/validate
that resource once, never overwrite it or manufacture duplicate owners. Scenario
decoders determine protocol; registry entries cannot override it. Schema1 keeps its
existing ownership/refusal rules. Test mismatched shared resources, duplicate report
paths and mixed protocol/report pairs before accepting this new registry path.

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

Before RFC acceptance, review parent C1g/C1h's owner/sampling bindings and H4's
explicit family/registry dispatch proposal, complete the actual temporal-boundary
algorithm/predicate justification, and reconcile the exact landing/version/hash
authority. The fixture reference is not a substitute. These are engineering gaps,
not owner choices to approve away.

Owner adoption of the approach and designated normative review remain separate
from the executable implementation acceptance above. Do not require an unbuilt
calculation API or migrated SQL workflow as proof this draft is implemented;
equally, do not claim those tests passed because the design was accepted.

## Deviations from existing contracts

Proposes origin-aware reuse and boundary-aware temporal search as explicit follow-ups
to archived Relevance R10/R11. Does not edit the archived body or silently change its
current implementation. No new balance values, caps, prices, grants, proration,
numeric tolerance, authored content or narrowed1.0 scope.

## Changelog

- 2026-10-15: align reuse with opaque typed run-local origins, not a new canonical
  origin codec/hash protocol. Preserve deterministic output, all reuse paths and
  full workload/budget gates; no implementation or adoption.
- 2026-10-15: specify proposed protocol/family dispatch, preserving legacy identities;
  separate new Relevance metadata from base policy grammar and make shared immutable
  content manifests single registry resources. No actual version/hash adoption or
  consumer implementation; complete boundary proof and review remain outstanding.
- 2026-10-15: split producer versus harness-search authority; retain RP-653/RP-656
  evidence, complete consumer obligations and unapplied status.
