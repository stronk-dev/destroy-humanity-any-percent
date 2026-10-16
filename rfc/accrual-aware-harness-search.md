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

For the proposed `simulation_active_play_endpoint_v1` profile, the post-purchase
rate used by the existing projected-milestone equation comes from the applied
child's actual current endpoint: pinned external inputs plus its still-active
buff contributions, assembled against its post-action state. Preserve the exact
numerator/denominator equation and wait-time addition. This is an explicit NEW-
profile sampling proposal, not proof that legacy scores are wrong. Fixed-external
and existing legacy profiles retain their own inputs; no report/dispatch default
changes before normative review and adoption. H3x supplies a bounded real-caller
comparison, not the missing full-workload or all-context justification.

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
   producer/schedule/absence inputs over the requested horizon. RP-656 originally
   declared only its fixture expiry offsets1000/1500ms; H3b adds a derived test-only
   clock/buff reference. Neither is the production factory or complete producer proof.
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
   RP-676 below now refutes generic monotonicity over these context-only regimes;
   this is a failed proposed proof route, not merely a missing larger sample.
4. Search regimes in order. Test the first endpoint, including the original time0.
   Only in a proven monotone regime may an unaffordable last endpoint exclude that
   regime; if its last endpoint qualifies, use integer lower-bound search only under
   that same proof. An unproved/nonmonotone regime cannot use this shortcut at all.
   The first
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
boundary enumeration, actual hook/resource predicate proof, work measurements and R10/R11 normative
review remain required; a per-ms exhaustive production scan is not authorized.

#### H3a — Actual predicate and clock bindings (RP-660)

Current `candidateCost` is one pinned `BulkCost(id, purchased,1)` for a generator
or the upgrade's pinned resource/amount. Current `gateRequirementsMet` is a
conjunction of resource-at-least checks. Waits do not buy generators or upgrades.
There is no timed purchase-discount contract in these callers; do not invent one
or require a proof over arbitrary future predicates. A new predicate family still
needs explicit support before adoption, not a default monotonicity assumption.

The proposed modern endpoint recipe must distinguish these existing mechanisms:

| Mechanism | Actual owner and boundary |
|---|---|
| Active buff expiry | `activePlayContributions` excludes a buff when its attended expiry is reached. Pending opportunity spawn/expiry alone does not claim an effect. |
| Attendance freeze | `prestige.RecordOfflineSpan` records the whole gap only when elapsed is strictly greater than `Prestige.CatchupCeilingMS` (currently5000ms). Resolve attended time through the actual history, not elapsed wall time alone. |
| Automatic online catchup | `buildOfflineCatchup` selects the offline stage only when online elapsed is strictly greater than `Economy.OfflinePolicy.AccrualCapMS` (currently86400000ms), not at the attendance ceiling. Catchup uses base contributions before the ordinary schedule/evaluation. |
| Production shape | `Evaluate` still owns the provision grid, burst duration, offline production cap and banking. Search ranges never reset these or turn their internal steps into payments. |

`TestAxisCalculationClockBoundariesResearch` executes actual logged transitions
at999/1000,4999/5000/5001 and24h−1/24h/24h+1/24h+2 milliseconds, both modes,
with the existing production buff and matched no-buff origins:36 cells. All exact
pre-action payouts agree with independent integer arithmetic followed by existing
state quantization; the forwarding hook observer preserves complete final
state/receipt/events. Attendance freezes at5001ms while automatic online catchup
remains absent through24h. At24h+1 the online pre-action payout is lower than at24h
even in the no-buff control; direct offline cash stays equal, banking1ms at24h+2.

These are independent operations from the same unchanged origin, not consecutive
saved payments or evidence of live balance loss. They require a separate cap+1
search boundary even without buffs. The original18 buffed cells passed before
the18 no-buff controls were declared; the owning plan/log retain that provenance.
This is bounded Go evidence and source binding, not an implemented enumerator,
whole-domain numeric/predicate proof, migrated harness, SQL/TS proof or AC6 repair.
The fixed-external and diagnostic profiles retain their distinct C1h policies;
do not add logged scheduling/catchup to those callers by association.

#### H3b — Concrete modern clock-regime candidate and rejected shortcut

For an immutable modern endpoint origin, measure offsets from its actual production
cursor, not the latest visible scratch time. Obtain original attended time `a` from
the actual history helper, attendance ceiling `C` from Prestige and production cap
`A` from Economy. For a requested horizon `h`, construct regime starts as follows:

1. Start with0. Add `C+1` when within the horizon, in either mode.
2. In online mode add `A+1` when within the horizon; offline mode has no automatic
   online-catchup switch. Do not confuse its ordinary production plateau with a
   new settlement or another allowance.
3. For every existing active buff, add `expires_attended_ms-a` when positive,
   at most `C` and within the horizon. Before the absence threshold attendance
   advances with elapsed time; beyond it the original attended coordinate is frozen.
   An unclaimed pending opportunity does not grant another buff or payout.
4. Sort and deduplicate starts. Each range ends at the next start minus1; the last
   ends at `h`, without computing `h+1`. For a later search lower bound, intersect
   these original-coordinate ranges with that bound; do not rebase the origin.

This candidate enumerates the identified clock/buff context switches. Production
still evaluates its whole original interval with actual provision, burst, caps,
hooks and quantization. The fixed-external profile does not gain logged scheduling;
the diagnostic retains its frozen window-start policy. Unsupported context, a
producer refusal or exhausted work is an error, never proof of unreachability.
The general producer/predicate and work obligations in H3 remain open.
RP-676's admitted provision-generated prefix defeats monotonicity within one of
these context-only regimes. This candidate enumerator is therefore insufficient
for the full admitted domain, even though its bounded clock/buff controls pass.
The offset formula is not permission to bypass source-state admission, actual
attendance/history validation, canonical clock checks or overflow handling. Its
bounded tests do not establish those contracts for every admitted timestamp/history.

The test-only enumerator now reproduces RP-656's independent literal cuts and
unchanged12/18024 population. A separate real-logged-probe reference uses horizon24h+2
and a hypothetical fixed100000-cash target. It finds the independently calculated
first qualifying millisecond11168956 (online buffed),78182687 (online plain) and
12409951 (offline buffed); plain offline is unreachable. Actual neighboring probes
confirm each found threshold. Whole-horizon search and omission of only `A+1` miss
both online windows; offline positive/unreachable controls remain correct.

That experiment also rejects a tempting arithmetic shortcut (RP-661). Its original
flat integer rate×duration cash oracle fails three actual probes by one canonical
unit. Direct existing `AccrueConstant`+ledger reproduces the ideal in two buffed
cases but not the actual `Evaluate`; in the plain case it agrees with `Evaluate`
and not the scaled reference. That reference converts integer nanocash through
`Decimal.New(...,-9).Quantize(12)`; it is not an exact-rational rounding oracle.
Provisioned counts stay unchanged. Preserve those exact negative
pairs; do not replace integration, change payouts or add tolerance. The prospectively
revised search check requires exact qualification/earliest/neighbor outcomes and
reports all cash-model disagreements separately. It does not claim the rejected
cash oracle passed or prove monotonicity of arbitrary unseen production states.

This executes a bounded reference over actual closed transitions, not a migrated
Relevance caller, new calculation factory, cross-runtime/SQL proof or repaired AC6.
All original partition, compatibility, full-workload and designated-review gates remain.

#### H3c — Numeric monotonicity must follow the actual varying inputs (RP-664)

Do not use monotonicity of arbitrary positive `SumDeterministic` inputs as a lemma:
the retained shared counterexample increases a raw term9.999999999999998→10 and
decreases the Quantize12 sum1.01245957117e1→1.01245957116e1. All terms can be
produced by canonical12-digit base rates and exact generator counts. An admitted
ephemeral catalog/save reproduces the one-second payout difference through actual
Go Rates/Evaluate; TS verifies the numeric producer prefix, not full replay.

That experiment changes purchased ownership, so its two projections have different
contexts. It does **not** establish a decreasing temporal predicate inside H3b's
fixed-origin regimes, require an invented temporal boundary, or show an actual
current harness miss. Within a provision bucket, raw rate aggregation is fixed;
elapsed time instead changes the integration/rounded-segment/ledger operations.
The actual caller proof must follow those varying inputs, including real provision
and burst transitions, rather than require or presume a false global lemma.
RP-662's eighteen passing pairs are bounded controls, not that lemma's proof.
No changed summation algorithm, payout tolerance, source pre-rounding or normative
adoption follows; the separate aggregation finding remains open for its own repair.

The current caller reduction is more specific than arbitrary ordered arithmetic:

- `candidateCost` fixes one generator/upgrade quote before the wait;
  `gateRequirementsMet` is a conjunction of fixed resource thresholds;
  `relevanceMilestoneReached` is one fixed resource threshold. Waits execute no
  purchase. Their predicates inspect the projected ledger, not hook event counts,
  lifetime value, stock progress or a newly invented timed discount.
- Inside one endpoint-selected context, `accrueContent` starts from the same
  provision/burst origin. Completed grid buckets have the same rate inputs,
  boosted milliseconds and Quantize12 deltas in every later projection. The
  current partial bucket has fixed raw rates; its elapsed and remaining boosted
  milliseconds vary. At the next grid edge a new nonnegative partial term enters.
  These are internal integration steps, not new settlements or context resets.
- The closed Prestige/Faction/Guild/resolved-Commons hooks change their owned
  lifetime/attendance/stock/tithe/sample state and private events, not the resource
  ledger after Evaluate. This source fact narrows these resource predicates; it
  does not waive those hooks' complete state/event equivalence obligations.

The remaining numerical argument must therefore cover the actual scalar
integration/Quantize12, addition of a varying **canonical** partial delta to its
fixed canonical bucket prefix, and the existing ledger quantization/cap/refusal
path. The original raw-rate counterexample is not itself that proof or its
negation. Exponent carries, sorted-group membership and normalization remain part
of this argument; do not assert that positive inputs alone prove them or introduce
extra search cuts without an actual algorithm/justification.

**Canonicalization alone does not discharge the group-carry step.** The shared
`numeric-canonical-group-carry-v1.json` retains65535 canonical3.003003003e1 terms,
one fixed3.003004057e1 term and a partial9.99999999999e0→1e1. Increasing that
canonical partial makes current grouped Q12 output fall1.96805804806e6→1.96805804805e6.
Go native/Linux-amd64, Node and all three Linux browsers reproduce the result,
including reversed source order. Independent exact decimal-literal sums increase
by1e-11. This refutes a general canonical-input summation lemma, not just a raw-rate one.
It does NOT establish that the unequal fixed-bucket prefix is reachable from one
admitted production origin or that an actual affordability predicate decreases.

One admitted v17 origin does reach the related equal-prefix RAW reversal: rate
3.003003003e-11, speed1e12,1000ms grid,65536 completed buckets and333 boosted
milliseconds in the partial bucket. Adjacent499/500ms produce exactly the canonical
partial pair above; actual Evaluate/ledger at both endpoints still commits the SAME
1.96805804804e6. These are unminted controlled inputs, not published balance or a new
loss finding. The remaining proof must use actual reachable prefix constraints and
the entire quantization/ledger path, or a separately adopted numeric contract with
its own proof; neither sampled equality nor canonical inputs imply monotonicity.

**A limited source-level sublemma, not the missing payout proof.** In both actual
`SumDeterministic` implementations, one fixed-exponent group's positive coefficients
are sorted ascending and accumulated with ordinary binary floating addition before
normalization. While the variable term stays in that same exponent group, that
coefficient fold is nondecreasing in its coefficient. This statement permits
rounding plateaus and assumes no invalid/non-finite intermediate; it is not exact
summation or a claim about the complete Decimal output.

Proof: rounding a nonnegative real sum to the ordered representable floating
values is nondecreasing, including ties, so a fixed sequence of additions is
nondecreasing in any one operand. Between adjacent fixed coefficients the
variable's sorted position is fixed. When it crosses a fixed coefficient `c`,
the two orderings coincide at the representable value `x=c`: swapping two equal
operands leaves the entire fold unchanged. Monotonicity on either side therefore
joins through that same value. Repeat for the finite list of crossed coefficients;
equal runs require no special ordering identity. All remaining fixed suffix
additions preserve the inequality. This argument does not require the empirical
287-bucket bound or turn passing profiles into a universal numeric lemma.

Consequently, moving within a coefficient group's sort order is not by itself
the unresolved step. The full caller argument still owes the actual scalar
integration and canonicalization, normalization of that group's total, subsequent
Decimal alignment/addition and ledger quantization/cap/refusal, and **changes of
exponent-group membership** (including grid/zero transitions). RP-664 crosses such
a group boundary and is not contradicted by this sublemma. No monotone regime,
additional search cut, alternative algorithm, work bound or RFC acceptance follows
until those remaining operations and actual caller boundaries are justified.

`TestCalculationBurstTemporalOrderResearch` attacks one difficult subfamily: a
large burst's fixed bonus makes the last bucket almost flat after its boosted
portion, allowing an adjacent-ms canonical delta exponent carry. All20000 seed666
profiles execute (tick1000..300000ms,1..287 fixed full buckets, burst portion between
one-quarter and three-quarters of a tick, factor1e10, derived canonical base rate).
Every profile has a crossing; none has a decreasing grouped/quantized total.
The first/last profiles also match four actual Evaluate/ledger projections from
the same admitted v17 origin and retuned v4 catalog, without consuming that origin.
This is native-Go-only bounded evidence, not a proof over all varying rates,
provision producers, burst policies, hooks, Go/TS factories or the modern caller.
No production balance is retuned and no per-ms production scan is introduced.


**An admitted temporal counterexample, RP-676 (2026-10-16).** The separate
[manual provision diagnostic](../planning/clout-v1-and-pr-interns/probes/provision-carry-probe.go.txt)
closes the specific reachability gap with a different unequal prefix, not by
declaring the earlier synthetic list reachable. Tick1000ms, origin1ms after run
start, first999ms and65535 full buckets. Base3.003003003e-23 with999999352334
purchased units; one provision boundary adds647666 units and hits its explicit
provisioned hardcap. A1e12 burst covers the first bucket, every full bucket and
333ms of the last bucket. At adjacent final499/500ms the canonical partial crosses
9.99999999999→10. The first canonical bucket is29.99998057, later full buckets
30.03003003. Both total durations remain below the existing24h cap.

From one encoded/restored v17 origin, TWO actual Evaluate/receipt projections
commit1.968058018e6→1.96805801799e6. The later queried payout is smaller. Independent
exact decimal-literal sums196805801799661999/1e11→196805801799662000/1e11 increase;
they are an order control, not the engine's arithmetic authority. The source origin
stays unchanged. This is no longer merely a raw-sum reversal or an arbitrary list.

Direct Evaluate receives assembled contributions, so it alone cannot claim the
harness wrapper. A separate admitted catalog clone removes only the active target's
ladder/manual/feed roles and feed membership, retaining its stock role and all
numeric/provision inputs. Actual contribution assembly is empty; TWO actual
SimulateAdvance projections with nil hook/default mask reproduce the same decline,
receipt/provisioned cap and provision-role activation. A fixed threshold equal to
the earlier cash returns true then false. This Go diagnostic does not execute the
complete modern factory, hook-bearing logged replay or search. Inputs are deliberately
unminted research content; no published-report defect, natural acquisition or
saved-payment loss is inferred.

Separate TypeScript public-replay confirmation uses the same neutral economic recipe
and a properly initialized v17 foundation state. A freshly hashed10-artifact bundle
passes its actual loader, and TWO independent ApplyLogged calls spend1 already-banked
compute millisecond after accrual without changing cash. Both apply; final cash and
receipts retain1.968058018e6→1.96805801799e6, provision cap647666, credits4999 and new
burst1ms. No private evaluator export or source mutation is used. It executes the
public v17 logged path, not SQL persistence, the modern v19 factory/search or complete
Go/TS state/event equivalence. The original Go primitive origin is not relabelled
as that complete foundation state. The owning log preserves failed setup attempts.

Consequently, the generic actual-wait monotonicity proposition over H3b's existing
context regimes is FALSE for the admitted domain. Do not keep extending positive
censuses or ask for a proof of that false proposition. The implementation gate now
requires a correct alternative search/partition with a complete proof, or a separately
adopted numeric contract with its actual caller proof and compatibility/workload gates.
Neither hypothetical additional cuts nor an existing draft is that solution.

The separately bound UNAPPLIED exact represented-input sum prototype eliminates this
specific decrease: same inputs, two Evaluate and two neutral SimulateAdvance projections
both pay1.968058018e6. This bounded counterfactual identifies a relevant correction
candidate; it is not whole-domain monotonicity, unchanged old payouts, acceptable cost,
independent review or adoption of either numeric or accounting proposal. Its already
recorded Pitch/catalog/Company/Commons/report compatibility obligations remain.

**Actual standalone search/ranking consumer, RP-680 (2026-10-16).** After RP-678's
independently approved source-version clone correction, the separate
[manual consumer diagnostic](../planning/clout-v1-and-pr-interns/probes/relevance-provision-window.go.txt)
loads the real scenario/routes/policy/catalog/hash through `LoadRelevanceSuite`.
It reuses RP-676's admitted neutral recipe and v17 late origin, retuning the existing
`upgrade.beige_tower_cache` cost and matching resource requirement to `1.968058018e6`.
`candidateCost` obtains that real quote; no hypothetical predicate replaces it.

Actual `earliestAffordable` on horizon65536499 finds affordable65536333 with27
charged transitions. Extending the horizon to65536500 returns unreachable, nil error,
after1 transition, although the earlier witness remains in range. Direct actual
SimulateAdvance/receipts at the two horizons retain the known cash decrease and
provision cap. Actual `rankCandidate` applies the upgrade at65536333/revision2 using29
transitions for the shorter horizon, but suppresses that candidate for the longer one.
The encoded origin is unchanged; work never exhausts the admitted scenario budget.

This is a current private consumer counterexample, not merely a proposed search over
literal thresholds. The catalog is deliberately retuned/unminted and the valid late
origin is seeded, not acquired through a natural run. The standalone factory has nil
hook/frozen inputs; this is not full RunRelevance, modern axis/attainment integration,
historical report drift, SQL or player loss. It does not prove the short method result
globally earliest over every preceding millisecond. It does prove the longer method's
unreachable result wrong because it contains the independently checked valid witness.
No algorithm adoption follows. A correct search must preserve such in-range candidates
without assuming whole-horizon affordability is monotone or hiding work exhaustion.

#### H3d — Positive aggregation upper enclosure, research component

An alternative to the disproved generic monotonicity lemma is chronological interval
search with **certified infeasibility**: exclude a range only when a sound upper bound
for every payout in that range is below the frozen quote. A failed endpoint does not
supply such a certificate. Search the earlier child first; exact point checks still use
the real producer. Budget exhaustion remains an error, never an unreachable result.
Without useful certificates this degenerates to a forbidden per-ms production scan;
the component below does NOT establish a complete or feasible algorithm.

For current `decimal.SumDeterministic`, consider nonnegative valid represented inputs
with a successful valid result. Let S be their exact mathematical sum, interpreting
each binary64 mantissa exactly and multiplying by exact 10^exponent, NOT parsing its
Q12 String. Let N bound the number of inputs and G bound distinct positive-input
exponent groups. Use uniform ceilings for an entire queried range, not an endpoint's
group count assumed to apply before a carry. This research component restricts
0<=G<=N<=86400002. That is an experimental domain, not a new content/runtime cap.

Candidate source-derived enclosures, using exact rational arithmetic:

```
epsilon = 2^-52
s_raw   = (N + 16*G)*epsilon + G*(5/10^15)
s_Q12   = s_raw + 2*epsilon + 5/10^12
U_raw   = S / (1 - s_raw)
U_Q12   = S / (1 - s_Q12)
```

Require the denominator positive. The maximum declared Q12 slack is about7.582e-7;
this is an analytical rounding allowance, not a changed equality tolerance, work
budget or acceptance criterion. It does not replace actual payouts with S or U.

Derivation for64-bit Go and positive finite-normal arithmetic:

1. Each group's coefficient fold has at most N normal floating additions in total.
   Each rounds upward by at most a factor1+epsilon. The group fold remains below1e9
   under this N ceiling. Positivity is essential; signed cancellation is excluded.
2. A group's `arithmeticResult` contributes at most one /10 rounding, plus normalization.
   A nonzero valid normalization has coefficient-scaling shift in[-2,16]: its input is
   in[1,4e15], and Pow10 outside that band leaves the coefficient outside[1,10) even
   after the one correction. The finite power range checks below establish this
   without relying on Log10 choosing the mathematically exact exponent. Saturation
   that yields a valid zero only drops positive value; invalid results are refusals.
   In-band Pow10 relative error is <=epsilon. Its reciprocal costs at most a
   factor1+2*epsilon; division and the possible /10 or *10 correction each cost
   another1+epsilon. Including arithmeticResult is at most five epsilon units/group.
3. A nontrivial positive `Add` uses at most five epsilon units for the two scaled
   operands, included Pow10 error and their sum. The aligned value is >=1e14.
   `Floor(aligned+0.5)` is no greater than
   aligned*(1+epsilon)*(1+5e-15): one addition rounding plus the explicit half-unit
   allowance. Normalization costs at most four further epsilon units. Gaps>17
   discard positive value and cannot violate an upper enclosure. Together with
   step2 this is at most15 epsilon units/group;16 is a conservative integer ceiling.
4. For Q12, exact factor1e11, positive normalized mantissa>=1 and nearest-integer
   rounding give two normal multiply/divide roundings and relative allowance5e-12.
   A carry to10 replaces it with exact1 at the next exponent. It adds no error term.
5. Positive weighted sums are bounded by the product of their path factors; the
   product of all factors overcounts every path. For nonnegative e_i with sum<1,
   product(1+e_i)<=1/(1-sum(e_i)). This gives the displayed enclosures. Removing
   roundings via fusion cannot exceed this positive-operation upper bound; do not
   assume every source expression is separately rounded. Go explicitly permits
   fusion ([language specification](https://go.dev/ref/spec#Floating-point_operators)).

The [manual component test](../planning/clout-v1-and-pr-interns/probes/positive-aggregation-enclosure.go.txt)
checks all40 Pow10 values from-17 through22 against exact rationals (nonnegative
powers there are exact), all632 finite power arguments' normalization range bounds,
and the two documented saturation edges
([Pow10 contract](https://pkg.go.dev/math#Pow10)). It compares actual raw/Q12 sums
with the candidate enclosure for declared edge cases and RP-676's two65537-term
lists. Both fit while actual Q12 decreases; the SAME N=65537/G=2 envelope increases.
Zero-slack and omitted-Q12-allowance controls actually underbound their named examples.
This manual exact-rational reference explicitly refuses exponents outside[-32,32]
and invalid/negative/special inputs or inadequate/unsupported ceilings.

This source argument and bounded challenge await research/normative review, not
corrective-code approval. They do not prove an interval enclosure for actual production
rates, partial bucket generation, modifiers, ledger addition, caps or hooks. Even if
all input terms increase, a **uniform** bound for their full range still needs proof;
the observed pair is not such a proof. Complete producer enclosure, chronological
search completeness, all caller refusal semantics, compatibility and unchanged work
budgets remain open before implementation/adoption. RP-680 and the27 original AC6
failures remain unresolved. No payout, runtime, report or ordinary CI change follows.

#### H3e — Fixed-prefix partial-production cell, bounded composition

The next [manual diagnostic](../planning/clout-v1-and-pr-interns/probes/production-cell-enclosure.go.txt)
composes H3d with actual `accrueBoostedConstant`, final aggregation/Q12 and actual
`Ledger.ApplyAccrual`. This is a research-only chronological search over ONE fixed
1000ms partial bucket, not a replacement for `earliestAffordable`/`rankCandidate`.
RP-680 separately binds the threshold to the real catalog quote; this diagnostic
reuses that value and does not invoke those private harness methods or buy anything.

Within this cell, the completed canonical prefix, raw rate vector, efficiency,
remaining burst and starting balance are fixed. Let P be the exact represented sum
of those completed deltas, R the exact represented value of the actual constant
`SumDeterministic(rates)`, E efficiency, V bonus factor and B nonnegative starting
cash. For integer d in[0,1000], prepare F[d] as the maximum exact represented value
of the ACTUAL `FromFloat64(ms).Div(FromFloat64(1000))` over ms0..d. This explicitly
costs1001 scalar preparations once; it is not an unproved endpoint shortcut or a
general solution for arbitrary large grids. Any unsupported query fails.

For epsilon=2^-52 define positive-operation factors from H3d's operation counts:

```
M = 1/(1 - 6*epsilon)                  // Mul + normalization
A = 1/(1 - 10*epsilon - 5/10^15)       // Add + normalization
Q = 1/(1 - 2*epsilon - 5/10^12)        // Q12
U_partial(D) = R * (F[D] + F[min(D,remaining_burst)]*V*M) * A*M^2*E*Q
U_payment(D) = (P + U_partial(D)) / (1 - s_raw) * Q
U_cash(D)    = (B + U_payment(D)) * A*Q
```

N is prefix term count+1; G is distinct positive prefix exponents+1, so a partial
exponent carry cannot invalidate the ceiling. s_raw is H3d's raw aggregation slack
with these **uniform** ceilings. Nonnegative multiplication and a monotone prefix
maximum F make this a range upper function without claiming actual payouts increase.
Every d<=D has both scalar seconds bounded by the corresponding F entries. M bounds
the bonus multiply and the two output multiplies, A the scalar Add, Q partial
quantization; H3d bounds final aggregation. The ledger's one-entry aggregation is
an identity, then its Add/Q12 is bounded by A*Q. Hardcap saturation cannot increase
the balance. Zero operations also fit. The rational arithmetic is certificate
research, not a replacement persisted calculation or changed equality tolerance.

**Definedness is separate from an upper bound.** A bound on successful outputs alone
cannot authorize pruning a general range that might contain producer/ledger errors.
All4004 declared pipeline endpoints execute successfully in the diagnostic; a future
real certificate factory must establish the validity domain too, or propagate an
error rather than silently prune it. This test does not establish that general proof.

Four complete0..1000 populations check actual cash against the proposed bound:
the retained65536-term provision/burst prefix; ordinary90% efficiency with nonzero
starting cash; zero rate; zero efficiency. The retained helper also executes the
two real Evaluate and two neutral SimulateAdvance anchor projections. The exhaustive
pipeline population is NOT full Evaluate at every point, a natural acquisition run,
SQL, modern hook/axis integration or full-state/receipt equivalence over all times.

The research search visits intervals chronologically, pruning only from U_cash(hi),
then tests fresh actual pipeline output at a leaf. The exhaustive reference establishes
earliest333ms within the retained cell. Horizons499/500/1000 all find333, including
the unaffordable500 endpoint:14/15/16 certificate nodes, one actual point each.
The full table/prefix/reference work is separately declared, not included in those
query counts or relabelled as actual harness transition accounting. Already-affordable,
unreachable, omitted-burst, endpoint-only, exhaustion and refusal controls remain
part of the diagnostic. No general workload feasibility follows from these counts.

Next: construct actual cells/prefixes from the full producer/context factory, prove
coverage/definedness, account for scalar/prefix/certificate construction work, preserve
complete caller semantics and validate full workloads under unchanged limits. No
per-ms live scan, numerical adoption, search-policy acceptance or defect closure is
authorized by this bounded composition. Research/normative review remains required.

#### H3f — Source-driven whole-wait cells, research constructor

The [next manual constructor](../planning/clout-v1-and-pr-interns/probes/production-wait-cells.go.txt)
removes H3e's hand-supplied prefix. Its explicit domain is the current action-free
`SimulateAdvance` path with nil hook/external inputs and Company versions below18.
It rejects hook-bearing or modern contexts rather than infer their behavior. This
is not an implemented API or the parent's unbuilt origin/context factory.

Construction follows the actual producer, not a separate balance model:

1. Assemble masked contributions once from the unchanged origin, as current
   `SimulateAdvance` does. Clone provision/remainder maps only; no settlement,
   origin reset, action or live-state mutation occurs.
2. Use the catalog's grid and original run start. For a positive tick, the first
   relative boundary is tick-((evaluated_cursor-run_start) mod tick), not tick from
   the query's origin. Accrue a segment using actual `ratesWithProvisionedAndPolicy`
   and `accrueBoostedConstant`; materialize its ending provision boundary using
   the same source function before obtaining the next segment's rates.
3. Retain the actual completed canonical cash deltas in an append-only shared
   prefix, plus running exact represented sum/count/positive exponent set. Each
   cell captures an immutable prefix view and its actual raw rate vector. No
   quadratic copy of every earlier prefix is needed. No authoritative ledger
   payment is made while constructing this diagnostic data.
4. Assign endpoints exactly once: first cell includes0 and its ending boundary;
   each later cell begins at preceding_end+1. Its partial duration is query minus
   the preceding boundary. Cash at a boundary uses the just-completed segment;
   newly materialized generators produce in the next segment, as in `accrueContent`.
5. Offline production stops at the catalog cap. If wall time extends further,
   append a constant-cash plateau from cap+1 onward, using the complete accrued
   prefix. Do not pay a second cap-sized interval or reinterpret it as banking.
   Real leaf evaluation still owns credit banking and wall-time burst consumption.
6. Build the explicit scalar prefix-maximum table for the largest actual cell
   width, then apply H3e per cell. Search cells chronologically; within a cell,
   prune only from its bound and check a fresh full `SimulateAdvance` leaf.

The manual experiment refuses more than100000 cells or scalar widths above60000;
these are research instrumentation limits, NOT new catalog caps or accepted work
budgets. Its diagnostic search uses the finite binary-tree node ceiling2*(horizon+1)
and explicitly errors on injected exhaustion. It does not declare a truncated or
unsupported search unreachable. Construction, scalar preparation, reference
reconstruction, certificates and real leaf calls are separately visible work.

Six populations compare reconstructed cash AND its complete cash receipt change
with fresh, independently restored actual waits: the unmodified60000ms grid at
twelve boundary/burst neighbors; complete0..3001 endpoints on1000ms online,
offline-capped and provider-masked variants; complete0..3001 no-tick Phase0; and
eight named references on RP-680's retained long wait. All12028 reference calls
execute. Every observed cash value fits its owning cell bound; all four exhaustive
small populations' search results equal their independent earliest-point scan.

The retained long constructor builds65537 cells from the source, not literal
repeated terms. Its search finds65536332ms elapsed (absolute65536333 from the1ms
cursor) despite the unaffordable final endpoint, using65551 certificate checks and
one fresh actual wait. This is not a separate exhaustive65M-endpoint oracle or
general proof that pruned ranges contain no errors. It is not a catalog purchase;
RP-680 separately owns the real quoted candidate/ranking evidence.

Negative controls detect lost provision and a grid wrongly reset to the cursor
at59984ms on the default cadence and984ms on the short grid. Missing offline cap
differs at1501ms. Its negative-control horizon is2000ms, because an uncapped3001ms
projection correctly trips the existing provision-horizon guard before a numerical
comparison. The honest offline population remains complete through3001ms. Hook,
domain, scalar/cell guards and explicit search exhaustion refuse rather than hide.

Each of the six profiles also executes whole-wait already-affordable0 and
unreachable1e12 controls. Main query counters are snapshotted before controls;
total work includes their additional certificate/point calls rather than conceal them.

This closes a bounded constructor experiment, not this draft's engineering gates.
The certificate remains an upper bound on successful outputs: all-resource domain,
ledger saturation/receipt refusals, modern endpoint contexts/hooks and general
definedness still require proof before pruning can be adopted. The actual full
scenario/beam/first-hour workloads and unchanged accepted budgets are unmeasured
for this prototype. One ordered local setup/search timing is not comparative
performance evidence. Current runtime/search/arithmetic/report identities and all
acceptance/review gates remain unchanged; RP-680 and Clout AC6 stay open.

#### H3g — All-resource partial arithmetic certificate (RP-681)

H3f's missing definedness premise has an executed counterexample, not merely a
warning. In an admitted v17/tick1000 origin, a secondary permits rate at the minimum
valid exponent succeeds for a full1000ms segment, yet actual `SimulateAdvance`
refuses every1..999ms. Cash stays zero. The old research constructor accepts its
full segments and cash-only search returns unreachable/nil without a point call.
That is a research-prototype defect, not newly executed private harness evidence.

The [partial-domain diagnostic](../planning/clout-v1-and-pr-interns/probes/wait-cell-partial-domain.go.txt)
executes all0..1000 actual waits for this input and three companions: more exponent
headroom, provider masking and zero offline efficiency. The bad input gives999 typed
engine/accrual refusals and two successes; all3003 companion points succeed. Failed
Company save encodings remain byte-identical, not a claim about every transient
in-memory field. Zero efficiency is a valid companion under the actual helper:
do not replace its behavior with an invented raw-exponent rejection policy.

The manual constructor now retains every produced resource's rate vector and
preflights the actual `accrueBoostedConstant` for integer partial durations. A
failure names cell/resource/duration and preserves the helper's error identity;
the bad permits context fails at1ms before cash pruning can hide it. The old
upper-only constructor is retained explicitly for the negative control. No active
production code, arithmetic, harness policy or ordinary CI population changes.

Reuse has a narrow source-derived justification. The helper validates all input
rates, then depends on them only through its exact raw deterministic sum. Validate
each new vector before reuse. Cache identity contains that sum's binary mantissa
bits AND exponent, efficiency/bonus bits and exponents, and burst schedule. Separate
no-burst, full-cell burst and partial burst; a partial schedule binds its exact
remaining milliseconds. For each key retain the greatest completely checked
duration, extend missing durations explicitly and reuse only covered subranges.
The cache is constructor-local, not a new persisted or cross-run identity protocol.

String identity is unsafe even for this validity-only cache: a safe headroom rate
and a raw coefficient just below a decimal carry share a Q12 String, but the latter
fails at1ms. Same-mantissa/different-exponent and one-versus-two-ms burst aliases
also hide real helper failures if merged. Executed semantic controls expose those
refusals, alongside raw identity, equivalent validated vectors, coverage extension
and burst-class controls. They are pure-helper inputs, not asserted natural saves.

All12028 prior actual-wait references/controls still pass. The long retained
constructor uses five partial-arithmetic contexts and4004 actual checks, reusing
131068 resource-cell ranges after131074 rate sums. Default60000ms-grid preparation
needs120008 partial checks across three contexts. This explicit setup cost is NOT
hidden in the one-point query count or proof of full-workload feasibility.

This certifies only the pure partial arithmetic over the manual constructor's
finite domain. It does not certify final multi-segment aggregation, ledger caps,
receipt reproduction, progress, hooks or modern contexts. Nor does an eager failed
preflight adopt an error-precedence policy over a valid earlier search result
(including already affordable time0). The real caller must preserve its accepted
semantics and distinguish invalid/unproved ranges from certified absence; this
prototype is not authority to shrink its supported domain. Complete composite
definedness/error handling, real workloads, compatibility and normative adoption
remain open. RP-680 and the27 original accounting failures are not fixed here.

#### H3h — Final positive aggregation domain (RP-682)

Partial arithmetic validity alone is insufficient. The next
[manual diagnostic](../planning/clout-v1-and-pr-interns/probes/wait-cell-aggregate-domain.go.txt)
admits a v17/tick1000 Company with a secondary permits rate6eMAX and two seconds
of wall time. All pure partials up to1000ms succeed. Actual complete waits succeed
through1666ms and refuse1667..2000 when aggregation is no longer a state value,
before permits'24 cap can apply. The partial-only prototype reports cash1 unreachable
with no point call. This is another prototype gap, not an executed private harness
failure or incorrect production arithmetic: the correct aggregate is out of domain.

The constructor now retains immutable canonical prefixes for EVERY produced resource,
not cash alone. Prefix maxima come from those actual nonzero deltas. H3g records
the maximum output exponent of each fully checked partial context; using a larger
checked context for a subrange is a conservative ceiling, never a tighter guess.
Each cell's possible final aggregate is its real prefix plus one such partial.

**Proposed sufficient interior lemma.** For at most100000 nonnegative state terms,
let E be the maximum positive-input exponent. If E<=MAX-8, actual positive
SumDeterministic followed by Q12 cannot leave the state exponent domain. An all-zero
sum is valid separately. This is NOT a new content cap or a monotonicity assertion.
Source-derived argument for the pinned64-bit Go normalization:

1. Coefficients lie in[1,10). Positive group folds have at most100000 additions,
   so coefficient intermediates stay below2e6: no binary overflow/subnormal path.
   Their normalization can raise an exponent by at most six, never below the
   group's minimum input exponent. There is no signed cancellation.
2. H3d's positive-operation error budget at N=G=100000 is less than1e-8. Hence a
   prospective prefix has represented magnitude below2e6*10^E and normalized
   exponent at most E+6. Establish this inductively: group normalization is already
   interior; the next Add's prospective carry is at most E+7<=MAX-1, so its state
   guard cannot fail, then the magnitude bound tightens its output back to E+6.
   This does not assume a successful final result in order to prove success.
3. Positive Add alignment has coefficients in[1e14,2e15], with finite normal
   operations. Its temporary exponent-minus14 fits int64 and is normalized before
   state admission; it cannot turn a positive input into a lower-domain result.
   Q12 can add at most one exponent, leaving at most E+7<=MAX-1.
4. This uses the current small-coefficient normalization contract, including its
   one-step decade correction, and H3d's finite power premises. The pinned Go
   [Log error analysis](https://go.dev/src/math/log.go) and
   [amd64 implementation](https://go.dev/src/math/log_amd64.s) support the needed
   near-decade logarithm accuracy; do not generalize to arbitrary math libraries
   or substitute the finite stress population for review of this argument.

This sufficient bound need not decide every admitted case. If inconclusive, the
prototype executes actual SumDeterministic/Q12 at every partial duration belonging
to that cell, using its COMPLETE real prefix, not a previously rounded aggregate.
The first invalid value returns cell/resource/duration and an engine-class error.
Valid extreme values are not rejected merely because the cheap proof fails.
Exact fallback checks/terms are visible preparation cost; this can be expensive
and is not authority for a live per-ms player simulation or increased work budget.

Five complete0..2000 actual-wait populations execute10005 references. The bad case
has1667 successes/334 expected typed engine/non-state-entry refusals with unchanged
encoded failed saves. Headroom4eMAX, near-boundary4.99999999999eMAX, provider masking
and zero-efficiency companions all succeed. Both valid extreme companions actually
execute2001 fallback checks/3000 terms. The bad constructor reports second-cell
duration667; the retained raw+partial-only path still demonstrates the missed error.

All16032 prior actual-wait references/cache/fault/guard controls are retained.
Five interior challenge shapes cover100000-term upper/lower/mixed/zero populations
and empty input. They challenge the lemma, not prove it by sampling. The retained
long wait uses131074 interior certificates and no exact fallback. No general cost
or complete workload conclusion follows from that particular case.

This closes aggregate-domain evidence within the manual constructor's scope,
pending research/normative review. Ledger Add/Q12, minimum/hardcap/receipt refusal,
progress, modern contexts and hooks are NOT thereby certified. The earlier-result
versus error-precedence caveat from H3g still applies. No runtime or accepted search
policy changes; RP-680,27 accounting failures and full1.0 requirements remain open.

#### H3i — Ledger Add, caps and receipt validity (RP-683)

H3h's aggregate certificate still does not certify the ledger. The
[next manual diagnostic](../planning/clout-v1-and-pr-interns/probes/wait-cell-ledger-domain.go.txt)
executes actual waits with valid nominal payouts but failing ledger addition or
receipt reproduction. Adding to6eMAX overflows at667ms; adding to an already-capped
9eMAX overflows at167ms, BEFORE capping can make it a no-op. Separately, a tiny cap
just above1eMIN requires an out-of-domain receipt delta and refuses every positive
endpoint. These are correct production refusals hidden by the previous prototype,
not permission to cap earlier or change arithmetic.

The manual constructor now validates the complete initial ledger once. Each
resource/cell needs its H3h aggregate certificate, including whether a positive
delta is possible and a conservative maximum exponent. The source-derived
sufficient ledger cases below rely on admitted canonical starting balances and
valid canonical nonnegative aggregate deltas:

1. No possible nonzero delta: the engine omits this resource's entry, so it remains
   untouched. Do not manufacture an Add merely because a raw rate was positive.
2. Starting balance zero: Add copies the delta; Q12 is idempotent. If capping binds,
   the cap is canonical, subtracting zero preserves it, and receipt reproduction's
   first candidate replays it. Zero cap/no change also succeeds. This proof works
   at the exponent edges and does not invent an interior-only supported domain.
3. For nonnegative nonzero before and minimum<=0, let E bound both before and delta
   exponents. If E<=MAX-2, positive Add and Q12 stay finite; the aligned coefficient
   is bounded as in H3h, and a possible two-decade carry still fits. The result is
   nonnegative, so it cannot violate that minimum. Require this BEFORE treating an
   already-capped balance as safe.
4. Under step3, an absent cap or cap exponent>=E+2 cannot bind: the rounded sum is
   below2.1*10^(E+1), while the cap is at least10^(E+2). Preferred delta is the
   original canonical input, so receipt candidate0 runs exactly the same Add/Q12
   and reproduces after. If before equals the cap, either no capping occurs and
   this argument applies, or capping returns before and no receipt is required.

Q12 idempotence here is a numeric argument, not a sampling claim: its coefficient
is the rounded binary64 value of k/1e11, with integer k<=1e12. Multiplication back
by1e11 has error far below half an integer, so nearest-integer rounding recovers k;
division repeats the same mantissa. A normalized carry to1 is stable as well.
Catalog caps and restored balances use the corresponding canonical grammar.

If none of these sufficient cases applies, the prototype computes the actual full
prefix-plus-partial delta and calls public `Ledger.ApplyAccrual` on a restored copy.
It preserves the actual ledger error, with resource/cell/duration, rather than
reject a value for being near a bound. Only successful checks are cached, by
resource ID AND raw delta mantissa bits/exponent, constructor-local to the fixed
original balances/catalog. A shared nominal amount is not a shared ledger context.
The source ledger validates resources independently before its commit loop, so
these per-resource validity checks compose for this engine's one-entry-per-resource
transaction. This does not certify later progress or hooks.

Twelve full populations execute13012 actual waits:2168 expected ledger refusals,
10844 successes with every emitted receipt strictly replayed through `Ledger.Apply`
from the original scoped balances. Valid headroom, zero starts (including tiny
caps), already-capped safe values, zero efficiency and sub-resolution no-change
cases remain supported. The existing R1 numeric recipe requires the actual adjusted
delta9.81579561655e8, not its nominal9.81579561656e8; its offline plateau repeats
the full receipt and exercises one exact-input cache reuse. A separate component
control proves the same delta can succeed for cash yet fail for permits, so the
cache may not erase resource identity. It is not an asserted acquired Company.

All26037 prior actual references/controls are retained. The long retained factory
uses131074 ledger fast certificates, one initial restore and no exact ledger calls
or fallback aggregate work. Other starting balances/caps may require expensive
prefix recomputation: counters expose it, and full-workload feasibility is still
unproved. No accepted budget is raised and no live per-ms loop is authorized.

This is ledger validity within the research constructor, pending independent
research/normative review. It neither settles error precedence over earlier valid
results nor proves the entire wait, progress, modern contexts, hooks, caller
compatibility or workloads. No production/core numeric/CI/balance/accepted-contract
change; RP-680,27 AC6 failures and full1.0 scope remain open.

#### H3j — Progress definedness after ledger commit (RP-684)

The [manual progress probe](../planning/clout-v1-and-pr-interns/probes/wait-cell-progress-domain.go.txt)
reaches another actual refusal after H3g–H3i all succeed. Its admitted v17 origin
has one provider paying permits at1/second and cash0. Tier0's composite combines
resource_log(permits,target1e6) weighted1eMIN with count_fraction(total_owned/1)
weighted1. Both positive canonical weights are admitted: their deterministic sum
is1. Initial progress succeeds, but each positive0..1000ms projection multiplies
the small nonzero log coordinate by the minimum weight outside the Decimal domain.
The real post-ledger tierProgressPPM refuses. The old ledger-certified prototype
returns cash-unreachable/nil without a point call and hides all1000 refusals.

The actual failed projection retains its changed permits and evaluated cursor,
and its encoded save differs from the origin, although its returned evaluation
is empty. This is **not** a claim that current production persisted a failed
transition: the tested real harness wait evaluates a clone. The research origin
stays byte-identical. Do not generalize earlier atomic-ledger error observations
to every later stage of evaluation.

The proposed manual guard calls actual tierProgressPPM on the origin once. After
H3i, the final ledger contains valid state balances; each resource referenced by
progress that was nonnegative initially remains nonnegative (all entries are
nonnegative, and an admitted cap cannot be below its starting balance). For an
admitted catalog, the following are sufficient cases, not new catalog restrictions:

1. A direct count_fraction stays defined: the actual initial check validated
   the sum/required count, and action-free accrual never changes GeneratorCounts.
   Provisioned ownership is a separate map and is not read by countFraction.
2. A direct resource_log stays defined. One.Add(nonnegative state value) is
   finite and at least1: potential upper-domain inputs are separated from1 by
   more than Add's17-decade alignment cutoff; possible carries occur far below
   the exponent ceiling. Its Log10 is nonnegative finite. The catalog already
   validated the positive finite denominator One.Add(target).Log10(). Both
   logarithms are converted from float64 by actual Decimal.FromFloat64. Any
   nonzero finite float64 has normalized decimal exponent between-324 and308;
   their direct quotient therefore has exponent no lower than-633, including a
   one-decade coefficient normalization. Zero stays zero. No tight transcendental
   approximation or monotonicity claim is needed. Clamping/Q12 gives valid[0,1].
3. For a composite with at most100000 terms, require each weight's exponent
   >=MIN+700 for this sufficient proof only. Count coordinates have much more
   headroom (safe integers divided by a positive safe integer); resource-log
   coordinates have the preceding loose bound. The margin keeps every nonzero
   weighted product within the lower domain. Products are nonnegative and bounded
   near1, so the positive aggregate interior argument of H3h applies. Clamp/Q12
   and the final ×1e6/Floor/Int64Exact conversion stay in their finite/safe range.
   The initial count checks cover any count terms; their inputs do not change.

No coordinate means actual tierProgressPPM returns0. These cases certify each
cell without projecting every endpoint. They do not replace actual final progress
values or receipts with a surrogate. The pinned Decimal/production sources own
the argument; the generous700 margin is a proof case, NOT an accepted weight floor
or a measured runtime budget. It needs independent research/normative review.

When this argument is inconclusive, the manual fallback walks the assigned
positive endpoints, reconstructs every actual final resource aggregate, applies
the complete real Ledger.ApplyAccrual transaction to a fresh ledger and calls
actual tierProgressPPM on an origin copy with that ledger. Progress reads no
other changed field. Time0 is skipped because the actual wait returns before
progress; this does not settle the constructor's separate eager-error ordering.
The fallback changes no origin map or live state and caches no partial-coordinate
result. Every aggregate/term/partial, ledger and progress call is separately counted.

Seven complete populations execute7007 actual waits/1000 expected refusals, with
exact successful receipts and progress-result checks. Headroom weight, masked
producer, zero efficiency, direct resource_log and direct count_fraction pass.
Crucially the SAME minimum weight also passes with target5e-15, where each positive
resource coordinate clamps to1 before weighting; it is not rejected for missing
the sufficient-case margin. All three valid tiny-weight companions execute1000
fallback ledgers/progress checks. The count-only population creates a generated
tower at1000ms but retains500000 progress/delta0, testing purchased-count semantics.
Native and declared emulated Linux AMD64 retain all earlier39049 references too.

This guards the demonstrated manual progress gap, not complete caller semantics.
All-range eager validation may still reject a valid earlier result; original
already-affordable/no-op ordering must be preserved rather than assumed. Hooks,
modern endpoint contexts, global error precedence, complete workloads, adoption
and full acceptance remain open. No production search/math/schema/CI is changed.

#### H3k — Chronological result/error ordering, research query (RP-685)

The [manual query](../planning/clout-v1-and-pr-interns/probes/wait-query-order.go.txt)
and its [actual-reference tests](../planning/clout-v1-and-pr-interns/probes/wait-query-order-test.go.txt)
separate a successful whole-domain certificate from the query's outcome. Eager
H3g–H3j validation can itself suppress a valid earlier result. With admitted
cash1/second origins, a ledger refusal at667 hides a valid500ms result, and an
aggregate refusal at1667 hides1500ms. Raw producer preparation can also fail
later than a valid result. The eager composition additionally rejects available
time0 results or consumes an unnecessary wait. This is a manual composition
finding, not a separately executed new live-harness defect or a policy adoption.

The proposed outcome is the first qualifying result OR actual error in canonical
endpoint order. It must not redefine any point's payout, silently skip a refused
point, or use a later failure to discard an earlier result. Initial balance >=
quote returns0 before preparing/evaluating a wait, matching the current private
caller's source contract. If it does not qualify and horizon=0, there is no future
endpoint: return unreachable without a budget, callback or constructor. Positive
results still require an actual full SimulateAdvance leaf, not a balance surrogate.
Caller compatibility and normative adoption remain separate, particularly the
current future-search method's horizon-first error ordering.

For the explicit H3f–H3j legacy/nil-hook/nil-external research domain:

1. Prepare actual source cells once, retaining producer failure as an explicit
   stop, not an omitted tail. A segment's full-width arithmetic failure stops at
   that endpoint; an ending provision failure also belongs to that endpoint.
   A new rate-vector failure after a completed boundary first affects start+1.
   Keep all earlier cells, clipping the failed endpoint from normal cell coverage.
   Their immutable prefixes predate the failed source operation. Initial rate
   failure leaves no cells and stop1, not a vacuous completed factory.
2. Translate the online provision-horizon refusal to its first invalid duration:
   (floor(catalog_cap/tick)+2)*tick. Prepare through stop-1 and retain that stop.
   This follows the actual guard's strict integer comparison. The offline capped
   producer does not gain an extra allowance. An earlier producer failure replaces
   the later horizon stop. Existing eager construction retains its original guard.
3. Try the complete H3g–H3j certificate once for all prepared cells. If it succeeds,
   reuse it throughout the search; do not repeat every partial/domain check for
   each node. If it fails, that is an inconclusive range proof—not the query's
   semantic error. Each visited subrange gets its own fresh partial/aggregate/
   ledger/progress certificate, retaining all failed/repeated preparation work.
4. Only a defined range AND a sound cash upper bound below the quote may be
   pruned. Otherwise split in integer order, left before right. A singleton
   executes the full actual wait and returns its real error or qualification.
   A failed/unsupported upper-bound calculation cannot prune either. Never use
   a successful end sample as proof that earlier points are defined or affordable.
5. After all earlier cells are exhausted, an explicit source stop MUST execute
   its actual endpoint. Return that real error. If it unexpectedly succeeds,
   fail the source-stop witness loudly; do not bless the missing tail. An earlier
   error discovered by refinement precedes a later source stop.

Conditional on the component certificates, the ordering argument is structural:
each earlier exhausted range was certified defined and unaffordable or exhausted
as actual nonqualifying leaves. Every reported positive leaf is therefore earliest;
an error leaf has no earlier hidden error/result. Unreachable requires complete
coverage and no outstanding stop. A certificate refusal alone establishes none of
those outcomes. The manual wrapper does not invent validity of arbitrary global
preconditions, hook/modern contexts or unsupported resource/quote families.

QueryNodes counts each visited node and mandatory stop, with the diagnostic finite
tree ceiling2*(horizon+1). Exhaustion is an error. QueryRanges/Refinements/Stops and
existing partial/aggregate/ledger/progress counters expose actual repeated setup;
global certification is setup, not a concealed free per-node operation. A difficult
range may require very expensive exact certification or singleton evaluation.
This is NOT permission for a per-ms production scan, a new runtime work ceiling or
an adopted optimization; complete workloads under accepted budgets remain gates.

Eight complete profiles enumerate14008 actual references/5003 expected refusals:
ledger, aggregate, raw partial, progress, future and initial rate failures, online
horizon guard and a masked valid companion. Four quotes each distinguish time0,
earlier result, error before later qualification and genuinely unreachable.
Empty-future and budget-exhaustion controls retain their distinct meanings. The
same six original constructor profiles/12028 references are reused by a shared
runner to bind the new query too, without duplicating their reference census.
The retained long nonmonotone wait still finds65536332ms with65551 search nodes,
one actual leaf and one global domain certificate, despite its unaffordable horizon.
All original source-fault/refusal/cache/receipt controls remain selected. Deliberate
definedness bypass and dropped-stop variants fail the first-error oracles.

This establishes a bounded proposed query mechanism, not a migrated real caller
or complete wait-domain proof. Modern contexts/hooks, full caller/state/resource
contracts, compatibility/error-policy adoption, complete workloads and designated
research/normative review remain open. RP-680 and Clout AC6 are not closed. No
live search, math, schema, constants, normal CI or player behavior is changed.

The [tracked overlay manifest](../planning/clout-v1-and-pr-interns/probes/wait-query-overlay.json)
maps the manual Go sources using paths relative to the server working
directory used by root Make and the declared Docker service. Reproduce the full
population with root `make test-go GO_PACKAGES='./production' GO_TEST_FLAGS='-count=1 -overlay=../planning/clout-v1-and-pr-interns/probes/wait-query-overlay.json -run TestWait -v'`.
It requires no machine-specific ignored overlay or change to ordinary CI. The
portable manifest's actual new-query consumer also passes on native/Linux;
this is not an executed clean-host/release rehearsal or a full CI claim.

#### H3l — Concrete faction-stock hook range component (RP-686)

The [stock-hook probe](../planning/clout-v1-and-pr-interns/probes/wait-stock-hook-domain.go.txt)
uses admitted Company19/axis inputs and the actual faction.AccrualHook through
simulationHook, including its real masks, version-preserving counterfactuals and
event refusal. It does not remove H3f's legacy/no-hook constructor guard, accept
arbitrary callbacks or implement a modern context factory. Its operation is the
existing fixed-context SimulateAdvance stock primitive, not logged scheduling.

Hook-domain validity itself is not monotone over a whole horizon. With one stock-
role tower, original stock cap-1 and progress interval-1500, actual waits saturate
from1499ms, emit faction_stock_saturated and are refused by SimulateAdvance. From
5001ms they succeed again because faction accrual skips elapsed time strictly
above its pinned5000ms attendance ceiling. Masking the tower retains the base
stock rate, shifting the first refusal to1500ms rather than removing all stock.
A proposed endpoint-only hook certificate therefore incorrectly blesses ranges
ending at6001ms. This is a failed proof shortcut, not a live event-policy defect.

The source-bound component uses the concrete hook's read set: pinned faction
catalog/resolver/catchup policy, elapsed time, fixed purchased counts/masks and
original stock units/progress/remainder. It reads no cash, production receipt,
ProductionMS, ProgressDeltaPPM or lifetime value. It writes only the three stock
fields. Full-wrapper probes normalize only those fields and require identical
encoded origin, guarding masks, source version, attainment, ledger and other
replay-owned state. Role observations come from the actual counterfactual wrapper;
they are not reconstructed from guessed stock-rate multipliers.

Conditional on an admitted pre-hook state and the independent producer/ledger/
progress/encoding proof, range certification follows this specific source:

1. On the attended branch1..C, the hook's factor is the positive exact integer
   1e6 + sum(purchased_count*stock_role_ppm), after the actual mask. Big-integer
   arithmetic gives floor((elapsed*factor+original_remainder)/1e6), nondecreasing
   in elapsed. The overflow test against MaxExactInteger-original_progress is
   therefore monotone. From the original state, earned stock is another integer
   quotient of original_progress+scaledMS. If initially below cap, saturation-
   event eligibility is monotone too; if initially capped, no new event occurs.
   Modulo progress/remainder and role labels need NOT be monotone for this proof.
2. Each stock-role counterfactual only removes a nonnegative purchased factor;
   it cannot overflow when the actual factor fits. The source wrapper retains
   actual events; its counterfactual events are not promoted into actual output.
   Source-version/state-admissibility obligations are not waived by this argument.
3. Above C the fixed policy/state checks precede the early stock skip. Their
   inputs do not depend on the query endpoint. No stock accrual/event then occurs.
   Zero elapsed bypasses the hook in the actual wait, and the none recipe has no
   hook-domain work. These are not generic assertions that every wait is valid.
4. Split a crossing range at the actual strict cutoff C/C+1. Probe the end of
   each contained branch using the real concrete simulation hook. A successful
   attended end excludes earlier integer overflow/events in that branch. A
   successful absent end covers its fixed checks. Never certify only the absent
   end while ignoring the attended branch. Refusal means the range proof is
   inconclusive: H3k would refine it, not label the range unreachable or infer
   that every point fails. Instrumentation/write-set errors remain loud errors.

Eight complete online/offline populations enumerate48016 full waits/14006 expected
refusals, with no-hook and already-capped companions. All full waits match their
fresh-origin concrete-hook stock/role probes; actual no-hook cash controls match.
At64 named anchors, an additional no-hook full wait agrees on evaluation when
returned and on the entire encoded state after normalizing only the three stock
fields—even where the event-bearing wait returns an error after mutation.
All104 named range results equal independently counted complete-reference validity;
24 invalid-range controls refuse. The old endpoint-only method fails all eight
mixed-cutoff oracles, while valid crossing companions remain certifiable.

Reference probes and range probes are separate work: each hook-bearing population
has6001 reference wrapper probes; its13 range checks need12 more wrapper probes
when saturating or14 when already capped, with two actual-cutoff splits. The
`none` recipe uses zero wrapper probes. These counts do not pretend the internal stock-
role counterfactual is free. This is not a production performance/workload result.

The range certificate is hook-only and must not independently authorize cash
pruning. It is not integrated into the modern query, a proof of complete logged
stage/absence ownership, full callback purity, or a replacement for other hook
families. The new private-event recipe proposed by parent C1d has DIFFERENT event
ownership; do not carry legacy event refusal into it by association. Bind each
operation's adopted contract, preserve complete scratch state/events and obtain
research/normative review/adoption before live use. Existing constructor/axis
holds, global/caller/workload gates, RP-680 and the27 Clout failures remain open.

#### H3m — Owned fixed-v19 producer/query composition (RP-687)

The [manual modern binding](../planning/clout-v1-and-pr-interns/probes/wait-modern-binding.go.txt)
composes H3f–H3k's real producer/range search with H3l's concrete faction-stock
domain. The unbound integration first reproduces hidden/late hook failures even
though its v19 cash producer is correct. It is a separate research entrypoint,
NOT permission to remove the real harness's legacy/axis holds or implement this
draft before adoption. Ordinary production and its accepted contracts are unchanged.

The manual fixture owner:

- Takes an already-admitted pinned test bundle and Company19, explicit online/
  offline mode, one closed recipe (`none` or `faction_stock`), a validated ablation
  mask and frozen external contribution rows. It refuses other versions/recipes,
  including `closed_replay` and `logged_endpoint`; there is no arbitrary hook
  callback input or silent fallback to none.
- Deep-copies artifact bytes and reloads their objects with the existing complete
  replay test loader. It owns the encoded origin, copied external value rows and
  all five mask slices. Actual codec, foundation, faction, mask and contribution
  checks run on its owned inputs. This is a trusted-fixture research owner, NOT
  a general untrusted artifact-admission API or completed BeginCalculation contract.
- Uses the SAME source-construction body as the legacy experiment, with an explicit
  fixed-modern context. It never changes WireVersion to bypass guards. Actual
  assembleContributionsWithPolicy derives the Company19 axis input/upgrade factors
  and combines the fixed external rows; actual rates/provision/accrual helpers
  build the cells. The old entrypoint supplies the legacy context and still rejects
  modern saves/hooks. No live harness constructor changes.

The shared query core now distinguishes globally proven producer definedness
from globally proven hook definedness. A failed stock-domain proof cannot become
unreachable or a surrogate error. When the producer is already certified, refining
only the hook does not repeat producer/partial certification at every node. When
producer proof is inconclusive, H3k's full range certification still runs. Before
pruning, BOTH components must be defined and the actual cash upper bound must
exclude the quote. Guard instrumentation/invariant errors remain loud. The range
guard interface is an internal proof seam; the fixture owner selects the concrete
recipe rather than infer purity of a caller-supplied callback.

Preparing producer cells without a hook is not omission of required hook work:
H3l proves the concrete stock hook's conditional cash read/write boundary, its
range certificate checks the existing wait's refusals, and every selected positive
leaf executes a fresh FULL SimulateAdvance with the exact real recipe, external
rows, mask, source19 and pinned dependencies. The returned query selection owns
that detached complete state and actual evaluation/role observations. Failure,
unreachable and work exhaustion expose no successful projected state. Already-
affordable time0 returns a detached original with no wait; an empty future has no
projection. Internal observations do not consume the origin or create payments.
This does not waive global producer/pre-hook encoding/timestamp obligations.

Twelve complete0..6001 populations enumerate72024 fresh actual references and
14006 expected stock-event refusals: none, saturating/masked/capped stock, fixed
external factor2, and an unminted1500ms offline-cap fixture, in both modes. The
external row is validated by the actual simulation primitive; this does NOT
prove full logged Guild settlement/provenance. Four quotes per population give48
queries, checked against exhaustive chronological outcome/error references.
Prices sampled from actual no-hook750/2000 projections are fixed thresholds,
not catalog purchases, timed discounts or an adopted balance change.

Selected positive results are compared against independently restored DIRECT
SimulateAdvance calls using caller-bundle objects and separately constructed
concrete faction hooks—not against the fixture owner's projection helper. Entire
encoded Company state, evaluation and every role observation must match. The
initial draft's selected-state oracle reused that helper; it was corrected before
claiming this proof. The subsequent compiling fault that disconnects only the
returned projection's hook, leaving domain safety intact, fires six full-state
mismatches and eight first-error mismatches. Cash agreement alone is insufficient.

Ownership controls mutate the original state, mask/external slices, artifact
bytes, faction/policy objects, and an earlier returned state/role buffer. The next
query retains its original full state/result and pinned hash, with the owned
origin unchanged. Zero-time, empty-future and node exhaustion stay distinct.
Seven invalid version/recipe/sampling/mask/contribution/hash/mode cases refuse.
These are one-owner repeated-query isolation checks, NOT full fork/action/cache
identity, all artifact-corruption admission or lifecycle proof.

The complete affected run retains earlier60064 legacy and48016 stock references
too:180104 enumerated references plus the separate anchor/query/control calls.
Both native and declared emulated Linux AMD64 pass. Setup is not free: each owner
reloads one artifact bundle (49498 bytes, or49494 for the cap fixture), retains
2816/2817 origin bytes and reports begin time. Query restores/real leaves, all
producer setup and hook range/probe/split work are separately visible; contributions
are assembled during owner validation and again by actual source construction.
The diagnostic node ceiling is unchanged. No runtime budget or performance claim
comes from these single overlapping verification runs.

This closes the bounded fixed-modern binding experiment, not modern gameplay or
the parent accounting API. Remaining concrete hook families, actual logged stage/
clock/absence ownership, global timestamp/encoding/state domains, complete caller
and resource contracts, error-policy compatibility, full workloads under accepted
budgets and research/normative review/adoption remain mandatory. No generic hook
purity, complete modern factory, natural acquisition, report/candidate migration,
Go/TS/SQL parity or repaired AC6 follows. RP-680 and the27 partitions remain open.

#### H3n — Actual first-hour lifetime inputs and bounded query binding

The manual [harness capture](../planning/clout-v1-and-pr-interns/probes/wait-first-hour-capture.go.txt)
and [production consumer](../planning/clout-v1-and-pr-interns/probes/wait-first-hour-lifetime.go.txt)
add a separately identified `first_hour_lifetime` recipe. These are research
overlays, NOT live caller migration or authority to implement this draft.

The producer runs the actual ratified reference.greedy seed0 scenario with the
already-tested experiment tuple (200,2e0,50,1e4,10). It separately drives the real
runtime.step over real boundaries and requires its complete result AND19-command
script to equal public RunExperimentScript (11916 transitions). Three distinct
post-step Company14 origins are captured: first generator, tenth generator and
second-run start. Actual first-hour admission and the real reference ranker/hook
run; no observer, production source rewrite or forced version is involved.

Each origin supplies1001 fresh actual advances over the real1000ms cadence, cash
and lifetime, four full encoded-state/evaluation/role anchors and two actual
candidateCost/earliestAffordable queries. Cursor time is the captured state's
actual EvaluatedThrough, including a previously adopted future projection. These
are diagnostic calls AFTER the step, not a claim they occurred inside that step.
Ordinary nil external context is required; career/diagnostic contexts are absent,
not silently serialized away. Missing populations, failed references or source
mutation fail before the generated ignored packet is written.

An instrument initially refused a real generator price because String rounds its
internal mantissa. The corrected packet stores exact IEEE mantissa bits plus
exponent and asserts reconstruction; its canonical string is a display label
only. It does not change game wire/save precision. All six actual quotes are
unreachable in this short window. Two extra explicitly diagnostic thresholds per
origin (original and500ms cash) provide two positive500ms selections plus zero-time
controls; they are not falsely labelled catalog purchases.

The consumer reloads the identified pinned artifact bytes and restores source14
unchanged, never weakening H3m's separate strict19 owner. Its closed hook calls
the real prestige.AccumulateLifetimeValue, matching the private first-hour hook
without copying arithmetic or adding the served hook's offline-span work. The
independent oracle is the separately executed harness packet: all3003 cash/lifetime
observations,12 complete anchors and12 queries must agree. Selected positive state,
evaluation and every role match those harness anchors, not the consumer's own
projection helper. A compiling point-hook disconnection keeps cash/search/guard
intact yet causes2000 point mismatches and eight full-anchor/selection mismatches;
the zero-production second-run companion remains green.

The sufficient lifetime guard is conditional on producer/ledger definedness. For
nonnegative canonical before and receipt delta, receipt reproduction implies the
delta exponent cannot exceed after's exponent: positive Add's aligned coefficient
is at least1e14, so normalization/Q12 cannot lower the greatest input exponent.
If cash has a nonnegative cap, after<=cap bounds delta's exponent. Requiring both
cap and original canonical nonnegative lifetime exponents<=MAX-3 leaves room for
positive Add and Q12's at-most-two-decade carry. Cash before must be nonnegative.
Missing/extreme premises return INCONCLUSIVE, not invalid/unreachable; H3k then
refines to actual leaves. Six direct guard controls include a valid ignored delta
at lifetime9eMAX that is deliberately uncertified. This is not a complete lifetime
domain, generic hook purity or global pre-hook encoding/timestamp proof.

The generated228362-byte packet is native/Linux byte-identical. Reproduce the
harness capture first, then the production consumer through the portable overlay;
missing packets fail loudly. Its artifact copy/restore,3003 direct references,
cell/scalar/partial preparation and query work are visible costs, not free setup
or a proposed runtime allowance. The experimental early-run tuple and bounded
one-second populations do not prove natural/default play, all masked/ranked
queries, live report compatibility, full workloads or accepted budgets. Other
hook recipes/logged stages and all remaining H3m global/adoption gates stay open;
RP-680 and the27 AC6 failures remain unrepaired.

#### H3o — Fixed closed-replay private-event reference (range proof still open)

The manual [closed-replay reference](../planning/clout-v1-and-pr-interns/probes/wait-closed-replay.go.txt)
binds the actual Prestige → Faction → Guild → resolved Commons chain to the
proposed chronological query. This is a RESEARCH reference, not implementation
authority, a completed factory, or permission to change legacy wait semantics.
It owns source19 state/artifact bytes, mode, all mask slices, external value rows,
revision and a copied Commons weight. There is no pending Guild settlement batch,
live weight resolver, Store, publisher or active-play schedule. Fixed contribution
sampling is not the logged endpoint/catchup recipe, even with the complete hook.

Projection uses actual simulationPolicyFor, assembleContributionsWithPolicy,
evaluateWithSimulationPolicy and simulationHook(closedReplayAccrualHook). Its ONE
deliberately proposed difference from SimulateAdvance is retaining ordered detached
EventWrite values rather than refusing the legacy action-free envelope. It invents
no intent, log/revision, saved payment or event delivery. Payload buffers are copied.
Actual generator masks and stock-role counterfactuals still execute.

The independent reference forwards the real chain through the real SimulateAdvance
using a test tap that suppresses only the envelope's event rejection. It captures
events ONLY for the main scratch state, never the stock-role counterfactual. Every
real hook and error still executes. Whole encoded state, evaluation, roles, event
kind/schema/payload/order and errors are compared, not only cash or event counts.
The ordinary unwrapped wait still refuses event-bearing results and is checked.
This is execution of an existing primitive, not independent implementation of
numeric arithmetic or proof of Store/outbox non-delivery/once-only settlement.

Eight complete0..6001 populations (both modes) cover full membership, generator
mask, plain no-membership effects and a missing-weight dependency fault injected
AFTER valid construction. Origins start at12:59:59.500, stock cap-1/progress
interval-1500, so stock saturation1499/1500ms, attendance5000/5001ms and hour
sampling boundaries execute. The binding constructor itself refuses missing or
out-of-range Compact weight; the fault population does not authorize accepting it.
All48016 paired points include12002 expected actual errors and24004 successful
event-bearing references. Four diagnostic thresholds per population give32
queries with complete selected state/events, true zero-time, unreachable and
first-error outcomes. These are not catalog purchases or natural acquisitions.

The original reference atc61c4eb2 has no whole-range certificate for this chain.
Its explicit INCONCLUSIVE guard forces chronological real projections; producer cash upper bounds alone
cannot prune required hook work. An unreachable6001ms query therefore executes
6002 points/12003 nodes. Positive750/2000ms queries need751/2001 points, not one.
Missing-weight queries encounter the actual error at1ms after evaluating0ms.
Source preparation, scalars/partials, query/guard/refinement counts, owner restores
and actual main/counterfactual hook calls are visible. This is deliberately NOT
workload feasibility, an adopted budget or a release-ready search algorithm.

Input/returned-result isolation mutates caller weight, state, external factor,
mask, artifact/policy objects and earlier returned sample/role/payload buffers;
the next query must preserve full original state/evaluation/events. Failed,
unreachable, empty-future and exhausted queries expose no successful selection.
The separate compiling control disconnects only the projection's Commons stage,
leaving the real reference and producer preparation intact. Missing Commons state,
events and errors must fail even when cash and nominal purchase times agree.

H3p below adds a conditional closed-chain range certificate, preserving this
original unproved reference's results rather than its deliberately linear work.
Complete global domain proof and exact logged stages/clock/real settlement remain.
All global domain,
caller identity/lifecycle, full workload/compatibility, Go/TS/SQL, independent
research/normative review and owner-adoption gates remain. RP-680 and27 Clout AC6
failures are still open; no current payout, kernel, save, CI or live caller changes.

#### H3p — Conditional closed-chain range certificate and receipt domain

The [manual closed-domain component](../planning/clout-v1-and-pr-interns/probes/wait-closed-domain.go.txt)
composes source-derived hook obligations with the producer certificate. It is
NOT a whole-domain, global encoding/timestamp, adopted API or live-harness proof.
Its internal bind receives the existing producer and whether that producer's
complete global certificate passed. Only then may it use its per-resource bounds.
If producer proof was inconclusive, the guard stays inconclusive even when a
later local producer range succeeds; actual chronological leaves remain available.
Preparation failures internal to the instrument stay loud; a possible semantic
hook error is not substituted for the real endpoint's error.

Positive nonzero minimum exponents now accompany the existing maximums in raw-
identity partial caches, immutable complete prefixes and aggregate certificates.
Positive SumDeterministic and Q12 cannot reduce the smallest positive input's
exponent. Exact aggregate fallbacks record actual minimums. Reused wider partial
contexts can make the bound looser, never unsafely stronger. Existing arithmetic,
failure guards and catalog/version holds are unchanged.

For a successful canonical nonnegative resource receipt:

- Before>0: a delta below before.exponent-17 would be discarded by actual Add's
  gap rule; Q12 of the unchanged canonical before cannot reproduce a DIFFERENT
  after. Therefore delta.exponent>=before.exponent-17.
- Before=0: candidate0 already reproduces canonical after. Its exponent is at
  least the producer aggregate minimum, or the positive cap exponent when capping
  binds. A balance already equal to cap emits no receipt after successful positive
  accrual; zero-cap resources therefore need no invented zero payment.
- Nonnegative after<=cap and receipt reproduction bound delta.exponent above by
  cap.exponent. Missing caps or negative before are inconclusive, not unsupported.

Commons multiplies EVERY resource delta by the captured tithe. For positive
factor f, require lowerDelta+f.exponent>=MIN. Its product exponent is at most
upperDelta+f.exponent+1. Apply the existing sufficient100000-term/MAX-8 interior
margin to sequential capacity Add: its positive14-digit aligned rounding grows
by less than1+2^-44 per step, so at100000 steps total<1.001 times the exact sum;
the exponent is at most maxTerm+6 and finalQ12 adds at most one. Zero tithe yields
zero capacity without manufacturing receipts. This is a sufficient numeric
proof domain, NOT a new resource, value, work or acceptance limit. Inconclusive
cases refine and execute actual hooks, including valid lower diagnostics absorbed
by a larger cash term. Treating every out-of-domain intermediate as a new runtime
refusal would change existing behavior and is not proposed here.

The remaining fixed-chain obligations compose as follows:

- Reuse H3n's conditional lifetime certificate. For absence, call the actual
  RecordOfflineSpan on fresh scratch at the largest duration; from is fixed at
  the origin, so merge/collapse eligibility and any old collapsed duration do not
  vary with endpoint. Boundary timestamp premises remain explicit and conservative.
- Reuse H3l's closed stock read-set/cutoff proof, permitting its validated saturation
  event ONLY as a private value in this recipe. Ordinary stock waits still refuse
  those events. Actual full selected projections retain every hook/counterfactual.
- Use certified progress0..1e6, actual fixed membership contributions and the pinned
  Guild policy/carry. Its numerator is bounded by1e12+1e6. Duplicate/missing or
  inconsistent dependencies are inconclusive, never silent no-membership fallback.
- Run actual frozen Commons enclosure/compliance once per proof preparation; keep
  weight validity and original sample/hour-coverage premises. All-resource capacity
  requires the receipt argument above, not a cash-only bound.

The measured regression has valid modern19 producer/ledger output at EVERY point,
but actual Commons capacity fails at1..8ms online and1..9ms offline, then succeeds
again. It is an unminted secondary permits rate1e(MIN+3), real1.15115 axis input,
default.1 tithe and cash already capped at1e6 so no cash receipt masks the diagnostic.
Initial uncapped-cash attempts did NOT fail and remain recorded. Headroom atMIN+6
and full-prefix/offline-tail companions remain valid; actual receipt bounds and
complete state/evaluation/roles/events/errors are checked against the real primitive.
Additional uncapped-cash companions preserve valid absorption despite inconclusive
proof. These are admitted seeded diagnostics, not natural acquisition or lost saves.

The original48016-point closed reference and32 queries keep their complete results.
For certified ordinary ranges, unreachable needs zero actual leaves and one node;
750/2000ms selections each execute one FULL leaf. Missing-weight and the17 real
capacity refusals still reach the actual first error rather than become unreachable.
Producer scalar/partial work, extra contribution assembly/catalog load, stock proof
probes, span restores and receipt bounds are visible costs, not free preparation or
a full-workload speed claim. Removing only the lower-exponent safety causes the
new admitted-refusal oracle to fail while headroom/prefix companions remain valid.

Complete global pre-hook encoding, timestamp/state/caller identity/lifecycle,
logged endpoint stage/clock, real clearing/SQL/event delivery/idempotency, full
workload/compatibility and research/normative review/adoption remain mandatory.
No active arithmetic, save/kernel/data/CI/legacy contract changes. RP-680 and27
Clout AC6 failures remain unrepaired; this does not close the full accounting goal.

#### H3q — Pre-hook cursor/codec range obligation (RP-689)

The [manual temporal component](../planning/clout-v1-and-pr-interns/probes/wait-state-time.go.txt)
adds an independent range obligation before pruning a hook-bearing wait. Arithmetic
and hook-specific certificates do not prove that simulationHook's initial
Save.EncodeState and any stock-role counterfactual RestoreState can complete.
The typed research recipe owners select this obligation: lifetime and closed
replay require it, stock does when installed, none does not. A raw callback or
boolean is not a proposed public factory/admission API.

The failure is executable on existing loader/codec-admitted Company14 inputs:

| Seeded primitive clock | Actual behavior at/after1000ms | Required distinction |
|---|---|---|
| v4 catalog,9999-12-31T23:59:59Z, purchased beige stock role, lifetime hook | Encoder formats year10000, but the real stock-role counterfactual cannot restore that RFC3339 cursor. | First actual error is1000ms, not a later probe or unreachable. Masking that role removes this restore call and remains successful. |
| Existing v3 catalog,0000-12-31T23:59:59Z, permitted zero RunStartedAt, lifetime hook | At1000ms the new cursor equals Go's zero time and the pre-hook encoder refuses it;1001ms can succeed again. | A later successful wait cannot erase the earlier real error; masking production does not remove the required encoding. |
| Same origins, no hook | The primitive does not invoke the pre-hook encoder/counterfactual. | Preserve its actual success even when a separate encode/restore observation fails. Do not invent a new primitive error. |

The proposed temporal certificate is CONDITIONAL on an admitted restored origin,
its actual source version/pinned catalog and the earlier complete producer proof.
Within the existing86_460_000ms research domain, canonical UTC whole-ms addition
is ordered, exact and cannot overflow the absolute seconds/ms representation of
a parsed RFC3339 year0..9999 origin. Require exact relative endpoint offsets,
endpoint format/parse equality and the closed year interval. Exclude Go's zero
sentinel ANYWHERE inside the interval, not only at its endpoints. There are no
other calendar grammar holes between ordered UTC endpoints. Thirteen direct
controls cover the two boundaries, valid neighbors and invalid proof requests.

An inconclusive temporal certificate refines chronologically to actual primitive
leaves; it does NOT reject a date, add a server clock limit or manufacture an
error. Producer and hook certificates retain their independent global reuse, so
only temporal proof is repeated when it is the unresolved component. Time0/
empty-future shortcuts remain before future preparation. QueryTimeGuards counts
the extra checks separately; accepted node/work budgets are unchanged. This is
research, not permission to land this draft or enable held gameplay contexts.

The twelve complete1502-point populations cover both modes and none/lifetime/
masked-lifetime recipes. All18024 actual primitive outcomes include1008 actual
errors. The initial composition loses TEN first-error query results; the corrected
48 queries preserve earlier500ms success, first1000ms errors, genuine unreachable
and already-affordable0. Successful masked/no-hook companions are retained, not
discarded to make a broad refusal rule pass. A compiling bypass of ONLY the time
obligation must reproduce those misses with all actual references unchanged.

Full selected state/evaluation/roles are compared with exact cursor equality.
For observation ONLY, an encoded shallow view substitutes the original cursor
while the real cursor is checked separately, permitting complete comparison at
the actual zero-cursor failure. The real projected object and its encoder/error
observations are untouched. Six encode and3012 restore failures are recorded
separately from1008 primitive errors, not relabelled as additional failures.
These are seeded clock-domain diagnostics, not natural player dates or save loss.

The pre-hook write-set audit also identifies the remaining admission premises:

| Evaluate writes | Conditional preservation argument | Required source premise |
|---|---|---|
| Ledger | Successful all-resource ledger/receipt certificate retains canonical finite balances within catalog bounds. | Same admitted pinned catalog and original ledger; not a caller assertion of hash equality. |
| Compute credits/burst | Banking is limited by remaining safe-integer cap; burst subtracts min(original,elapsed). | Restored original bounds and actual validated policy; no new allowance per probe. |
| Provisioned/remainder maps | Actual materialization writes existing declared targets, clamps to a safe-integer hardcap and uses nonnegative integer remainder modulo1e6. | Admitted original map membership, valid mask and successful actual source boundaries. |
| Cursor | The temporal certificate above preserves the required encoding/roundtrip path, or actual leaves decide. | Canonical admitted restored origin and the actual hook recipe; not every hypothetical primitive requires a codec call. |
| Other Company fields | Evaluate leaves them unchanged. Monotone cursor advance preserves original incorporation/refill/run/offline ordering constraints. | Full original codec admission and source-version fidelity, not merely a valid ledger. Post-hook output validation is a separate obligation. |

This is not a complete arbitrary-input admission/identity, post-hook persistence
or all-caller proof. Remaining mask/map membership, actual stage/clock and owned
context/lifecycle obligations must be completed before global soundness is claimed.
Keep logged clearing, SQL/event delivery/idempotency, full workload/compatibility,
Go/TS parity, independent normative review and owner adoption mandatory. Existing
runtime/codec/kernel/data/CI and accepted contracts stay unchanged. RP-680 and27
Clout AC6 failures remain OPEN.

#### H3r — Primitive success and post-hook codec completion are distinct (RP-691)

The [manual codec-boundary reference](../planning/clout-v1-and-pr-interns/probes/wait-post-hook-codec.go.txt)
executes the complete real fixed hook chain and, separately, actual Company
EncodeState/RestoreState. It does not change SimulateAdvance's error contract or
declare that every observation must be a persistable settlement.

Codec/foundation admission of the origin is not alone a post-hook shape proof:

| Seeded admitted source | Actual primitive | Actual post-hook codec |
|---|---|---|
| Commons sample in a future hour | Succeeds and emits private Guild/Commons events. | Earlier appended hour breaks ordered samples. |
| Current Commons hour already fully covered despite a cursor500ms into it | Succeeds and emits the same event kinds. | Appended coverage exceeds the existing one-hour bound. |
| A500-year run with collapsed offline duration equal to actual saturated time.Sub | Succeeds, including RecordOfflineSpan after5000ms. | At5001ms, offline total exceeds the still-saturated run-duration check. |

These are input-domain/stage diagnostics, not natural player provenance, a live
SQL failure or evidence that ordinary gameplay lost saves. The time diagnostic
uses2526/2026, so both timestamps actually pass the nonnegative exact-ms codec.
An initial1526 run start did not pass admission and is excluded, not counted.

The previous primitive-only query remains correct for its named primitive. It
cannot be reused as proof of the distinct completed-codec boundary: the first
two cases produce eight different query outcomes; the span case lets its existing
successful guard prune two first codec errors as unreachable. The reference now
keeps the codec call in that diagnostic leaf and composes an additional sufficient
span obligation. If an interval crosses the catchup threshold, require nonzero
admitted run start and `original_run_ms + hi <= floor(MaxInt64_ns / 1ms)` before
using the run-duration argument. Otherwise refine to actual codec leaves. A
saturated run with EMPTY history remains valid; an unsaturated full-history
neighbor also remains valid. Inconclusive is not an input rejection or new date cap.

The source write-set audit names15 fields written by Evaluate plus this concrete
chain: ledger, provision/remainder maps, cursor, credit/burst, lifetime/offline
history, stock counters/remainder, Guild carry and Commons samples/solidarity.
Every other State field, including purchased membership, source version, axis and
scheduler state, is checked unchanged on the selected references. Full state,
evaluation, roles and private events also match a separate actual SimulateAdvance
tap even where the resulting state cannot be encoded. Allowing a field to change
is not a proof that its changed value passes the codec.

The bounded population is72 selected endpoint pairs,24 actual encoder refusals
and30 paired primitive/completed-codec queries. It is not an exhaustive time census,
whole-input-domain proof or full workload measurement. Extra encoding/restoration
and guard work is real work, not a proposed free preparation step or budget change.

The owning caller must specify its actual completion stage. Preserve primitive
success where that is its contract; preserve an actual downstream codec error
where that caller executes the codec. Do not silently add errors to all reads,
discard admitted states, or infer live settlement/rollback behavior from this
manual adapter. Next bind the real logged catchup/ordinary stages and endpoint
sampling against ApplyLogged, including its refusal path. Global input/lifecycle,
caller/workload/compatibility and independent normative adoption remain open;
RP-680 and27 Clout AC6 failures remain unrepaired. Kernel223/epoch8 and active
math, saves, gameplay, CI and accepted contracts are unchanged.

#### H3s — Actual logged-stage compatibility reference (RP-692)

The [manual logged reference](../planning/clout-v1-and-pr-interns/probes/wait-logged-stages.go.txt)
now observes actual ApplyLogged, not a second accounting engine. Its
[generator](../planning/clout-v1-and-pr-interns/probes/prepare-logged-stage-observation.mjs)
inserts deferred evaluation/hook/schedule observations only in ignored source
copies. Removing those insertions/named return slots recovers every original
byte. Observer-off/on commands retain identical full state, receipts, events and
invariants. No production file or default CI population changes.

The admitted unminted Company19 bundle has offline cap1500ms, so the actual
catchup branch is exercised cheaply rather than assumed from the5000ms prestige
attendance boundary. Four none/production/building/combined buff profiles, both
modes, eight endpoints0..6001ms and applied manual/rejected purchase arms produce
128 commands, including32 catchups and64 exact rejected rollbacks. Each carries
a nonempty captured Guild batch and resolved Commons weight. Buff histories are
controlled, not naturally acquired; the retuned cap is not released balance.

The initial manual prefix omits catchup and fails exactly32 cases while96 survive.
The corrected reference clears the captured batch on a fresh copy, assembles base
contributions, performs required offline evaluation/closed hooks BEFORE scheduling,
applies the actual endpoint schedule, then assembles endpoint contributions and
performs ordinary evaluation/hooks. No arithmetic is copied. Complete pre/post
states, source version, contribution rows, evaluations and private events match
actual command traces. Applied event prefixes retain their actual order.

Catchup paths contain a real ordinary Evaluate call with elapsed0. Its
runAccrualHook wrapper is invoked but correctly skips the INNER hook; recording
that zero stage is not another payment. Literal attendance/expiry, positive
lifetime/Guild/Commons output and exactly-once clearing anchor the observations.
All64 duplicate captured batches refuse unchanged. Six malformed clock/catchup/
span/opened/attendance/weight cases retain real refusal/rollback behavior; the
attendance failure occurs after catchup but before ordinary evaluation.

Native ARM64 and declared Linux AMD64 generate the same128-command/six-fault
packet byte-identically. The [manual client replay](../planning/clout-v1-and-pr-interns/probes/wait-logged-replay.test.ts.txt)
uses actual unmodified TypeScript ApplyLogged:203 checks pass, comparing full
state/receipt/events/restoration, duplicate and malformed-context refusals,
population/source identity and forged expectations. The Vite overlay replaces
one test slot only. This is not inclusion in ordinary CI or TS typecheck.

This supplies a bounded compatibility reference, not a new validated origin/
factory, arbitrary-context admission, masks in a modern recipe, full search,
SQL/live-provider settlement, natural/default workflow or repaired accounting.
Next bind an owned modern endpoint factory and chronological query to this exact
recipe, including retained private events and source/clock/clearing identity;
complete workload/compatibility and independent normative adoption remain gates.
RP-680,27 AC6 failures, kernel223/epoch8 and all live contracts remain unchanged.

#### H3t — Owned modern endpoint component and full-state reference (RP-693)

The [manual endpoint owner](../planning/clout-v1-and-pr-interns/probes/wait-modern-endpoint.go.txt)
now binds a Company19 source from H3s's real logged packet to owned artifact bytes,
revision coordinates, mode, all five mask slices, external rows, Commons weight
and a captured nonempty Guild batch. Existing loaders/validators check the fixture
context; copies isolate caller mutation. This is not a public production constructor,
DB-owner authorization, scenario/session provenance or permission to begin from a
previous projected head. The typed upstream factory requirements remain open.

Each independent endpoint starts from the SAME source version/origin. It clears
the captured batch once on scratch state, resolves actual catchup/schedule,
executes masked production plus the closed simulation hook, and retains ordered
`offline_catchup`/`ordinary` evaluations, private event values and role observations.
An ordinary elapsed0 stage remains. No intent is invented; error results expose no
successful state or private event fragment. A private nonzero-size context pointer
is run-local identity: forks share it, distinct begins do not. It is not serialized.

Complete0..2001ms references plus5000/5001/6001 anchors cover none/combined buffs,
online/offline and generator mask/no-mask:16040 full state/stage/event/role pairs.
The independent oracle invokes actual SimulateAdvance separately at the real
stages, intercepting only its existing event-envelope refusal. An upgrade-mask
ownership control additionally mutates caller state, masks, rows, weight, batch,
artifact/catalog views and nonempty returned receipt/event/role/buff-pointer data.
Parent and sibling projections remain unchanged. Twelve invalid contexts refuse.

The first owner roundtrips external values through the replay serializer and
silently rounds an admitted raw factor immediately above1 back to1. The raw-bit
control first executes the real primitive successfully, then fails that owner.
The correction retains existing grammar checks but OWNS the actual typed values
without another persistence quantization. Existing saved/replay wire precision
is unchanged; this is a prototype input-identity correction, not a payout finding.

The attached chronological query is explicitly an EXACT REFERENCE, not the
optimized candidate or live per-ms search. Each actual projection, including0,
consumes its finite point budget; exhaustion is an error, never unreachable.
Thirty-two reference queries match complete earlier scans and preserve full
selected results/identity. Two genuine early-affordable/late-unaffordable windows
discriminate endpoint-only search. A separately admitted stock-domain control
retains the first actual1000ms error after1001 charged points, with no success state.

Constructor/project durations include their complete calls; named counters expose
bundle bytes/load, explicit restores, clearing, schedule resolution, evaluations
and actual/counterfactual hook calls. Internal simulation-wrapper codec work is
included in elapsed time but not fully itemized. Reference and fixture preparation
are not free, adopted harness transition accounting or workload feasibility proof.

Next compose semantic endpoint regimes and source-cell/bound queries with this
owned actual projector, preserving schedule, hook, codec and first-error semantics.
Simulation-wrapper versus logged error-policy compatibility is still not proved
over all admitted inputs. Full context/lifecycle/mask-family/caller/workload and
normative adoption requirements remain; no default scenario guard, accepted binary
search, live arithmetic/save/data/CI contract or27 Clout failures are changed.

#### H3u — Endpoint-regime bound integration and query floors (RP-694)

The [manual bounded adapter](../planning/clout-v1-and-pr-interns/probes/wait-endpoint-bound.go.txt)
connects H3t's owned real projector to the existing source-cell/producer/closed-hook/
temporal certificates. It does not replace the accepted live binary search.

Derive chronological endpoint regimes from actual buff expiries minus original
attended time, attendance-cutoff+1 and (online) offline-cap+1. A boundary belongs
to the right regime. The clock argument requires an admitted original run whose
full horizon does not saturate time.Sub. Otherwise execute actual chronological
leaves; a valid saturated-clock control remains valid, not a fabricated refusal.

The actual0 projection is charged and may perform captured clearing/scheduling,
but has zero production, unchanged ledger/cursor/allowance and no inner accrual
hook. Its validated deterministic schedule prefix supplies the no-FURTHER-due
certificate. This matters: the actual source already has an opportunity due at0;
the first adapter falls back everywhere and fails the pruning check. If no later
transition is due, the full original schedule consumes exactly that same prefix
and its same transition budget. This is not a new origin, reset timer or free work.

Every scalar model still reconstructs the COMPLETE original interval. Catchup
uses base external context/offline mode; ordinary uses the actual endpoint buff
context. Post-catchup endpoint assembly is checked separately: its purchased/
upgrade/axis read set is unchanged by the closed hook, not a newly sampled payout.
The complete real projector remains the leaf and supplies selected state/stages/
events/roles. Unknown due-schedule ranges refine or use actual leaves. No generic
schedule-validity or unmeasured full-domain completion claim is made.

The shared manual query gains an optional endpoint floor, default0. It clips
visitation and the mandatory source-stop witness, and disables the time0 shortcut
for a later regime. It NEVER changes the production origin, prefixes or allowance.
The SHA-exact old helper returns0 for a1001ms floor; the regression fails. Default-
floor error-ordering and whole-wait constructor populations remain required checks.

The declared37 paired queries cover32 short cases, four late unreachable cases
and a selected window first affordable at5001ms. They preserve complete results
and first errors while using3301 full projections against49336 exact-reference
points. Setup/model/certificate work is exposed separately; this is not a matched
latency result or adopted harness budget. The node guard is the existing finite
full-tree2*(horizon+1) diagnostic bound, never a raised release ceiling.

Actual stock failure at1000ms and schedule failure at1001ms retain later-valid
companions. Bypassing ONLY the no-due guard moves the latter error to1501ms and
fails the oracle. The stock control deliberately retains1001 actual point calls
because its conservative upper bound does not use the resource cap. No savings
is inferred for that control. Exhaustion remains an error with no success state.

Next measure complete candidate/consumer work and error/scoring compatibility,
including general scheduling, all mask families, actual TS owner/query, lifecycle
and full workloads before normative adoption. No live math/save/schema/data/CI
change, RP-680 repair,27-AC6 closure or 1.0 milestone follows from this component.

#### H3v — Actual first-hour candidate boundary (bounded compatibility)

The existing [manual harness source](../planning/clout-v1-and-pr-interns/probes/wait-first-hour-capture.go.txt)
also executes the CURRENT `rankCandidate` against an independent chronological
`SimulateAdvance` reference and one original-state `SimulateTransition`. It reuses
the existing legal run3 fixture with none/unit/tree frozen factors, two actual
catalog quotes, all five mask fields plus unmasked, and immediate/delayed cash
controls. These72 cases are controlled ledger diagnostics, not natural acquisition.

All selected full states, revisions, earliest times and exact score components
agree. Twenty-four delayed purchases (12 generator/12 upgrade),12 rejected
actions, six unreachable quotes and six applied-but-zero-rate cases execute.
The latter is correctly not a ranked milestone candidate. Budget exhaustion and
an injected failure after the actual lifetime hook propagate without a child or
origin mutation.
Role lists agree but are EMPTY in this fixture: no nonempty-role claim follows.

The current first-hour wait-then-buy path does NOT lose lifetime state in this
population. That does not license installing modern projection views as action
origins or discarding their required private events. This test is not execution
of H3u's bounded query in the harness, a closed-replay action adapter, generic
schedule/score/order proof or complete workload/dispatch acceptance. The owning
log separates21062 reference advances from412 actual ranker transitions; neither
is total work or a matched performance claim. Existing score/policy/budgets and
all original RP-680/Clout AC6 failures remain unchanged.

#### H3w — Owned modern action boundary (manual component)

The [manual action component](../planning/clout-v1-and-pr-interns/probes/wait-modern-action.go.txt)
binds H3t's immutable Company calculation to an owned copy of the existing frozen
Founder carry, validated under the CURRENT replay-input version. It is not an
ApplyLogged wrapper, new wire format, live Service or general replay decoder.

Each action restores the original Company, derives stock, clears its captured
batch once, executes actual catchup/schedule/arithmetic/closed hooks and the real
semantic transition, then runs actual Prestige/foundation hooks and receipt refresh.
Prior projections are discarded, not installed as the action's origin. Applied
actions alone return a detached child with new identity and incremented simulation
coordinates. Its successor batch is empty at the consumed boundary: it is not a
live provider or permission to deliver the captured settlement twice.

All128 H3s command records compare complete actual ApplyLogged stage traces and
decision/state/receipt/events/invariants, including64 rejected actions with no
child or staged output. Eight additional actual generator/upgrade purchases cover
ordinary/catchup paths; upgrade cash is controlled, not naturally earned. A fork,
nonempty Founder-map/event/receipt/buff mutations, a second child command, invalid
bindings and an admitted actual stock-hook failure exercise ownership/refusal.
Severing only post-action foundation work fails all eight purchase controls even
though the pre-action trace structure still agrees.

This component supports only manual/generator/upgrade actions and the current
captured-input grammar. It does not prove every mask, nonempty role result, raw
input/error/codec domain, scenario/session factory, context change, gate/terminal
operation, Store retry/outbox or complete work budget. Actual harness integration,
TS ownership/query/action, workloads and normative review/adoption remain required.
No live accounting/search repair, RP-680/Clout AC6 closure or release follows.

#### H3x — Actual single-candidate caller connection (manual overlay)

The [manual generator](../planning/clout-v1-and-pr-interns/probes/prepare-owned-candidate-observation.mjs)
reversibly changes four exact binding sites in the CURRENT harness solver:
earliest-affordable query, skip of pre-advance for an owned original-state action,
semantic action call and post-purchase rate input. All other ranking/score code
remains byte-recoverable to its original source. Test-only production exports
and an external test join the owned implementations to actual `rankCandidate`;
the callback plumbing is not a proposed runtime plugin or trusted origin factory.

Twenty-eight controlled real-quote cases cover none/combined buffs, both modes,
generator/upgrade purchases and available/deficit/zero cash. Four extra windows
derive a controlled starting balance from actual500/2001ms payouts, keeping the
catalog quote unchanged. The ranker finds those genuine earlier qualifying
windows at338/348ms despite an unaffordable2001ms tail. It matches the exact
full-projection query and original-state ApplyLogged action, including complete
state/receipt/events and unchanged source. Post-action score components match
the proposed H1 sampling above. Twenty candidates apply,12 after waiting; eight
queries are truly unreachable. Zero/one-call budgets return errors, not success
or unreachable results. The returned role list is checked against the adapter,
not an independent proof of every role/mask family.

Restoring ONLY whole-horizon rejection loses all four windows. Omitting ONLY
active contributions in the actual rate callback changes12 scores, while the
independent reference remains unchanged. Legacy Reference lifetime/frozen-input
checks pass through the same generated solver's default path on native/Linux.
The separate caller overlay removes H3s's three normal-file observation taps,
which depend on test-only declarations; it does not change that producer or its
reference packet. Mixed-package runs must not inherit those taps accidentally.

The28 queries use40 full projections against21556 reference points and60 caller
transitions;12 fixture projections plus constructor/model/certificate/codec/role
work are separate, NOT free. The finite-tree-plus-one-action diagnostic ceiling
is not an adopted R11 preflight or matched speed result. This is ONE candidate
boundary: retaining its returned child through actual choice, beam/cache/fork/
rollout/context reuse, all masks, full error domains, TS/SQL, dispatch/workload
proof and normative review/adoption remain. No live repair or AC6 closure follows.

#### H3y — Selected child and real clone lifetime (manual component)

The [scoped choice bridge](../planning/clout-v1-and-pr-interns/probes/owned-choice-bridge.go.txt)
binds complete heads to opaque owned action contexts while the ACTUAL
`rankDecisionOptions` selects and the REAL `cloneState` restores its chosen state.
The [test factory](../planning/clout-v1-and-pr-interns/probes/owned-choice-export.go.txt)
supplies independent branch callbacks/collectors; forks share immutable identity,
real applied actions create new identities, and copied head buffers reject missing,
typed-nil, mutated or unbound contexts. No identity is serialized, hashed into a
report, used as an ordering key or attached to a save. The maps are scoped manual
plumbing, not a trusted production constructor or long-lived cache.

Four two-choice runs (combined buffs, both fixed modes, available/deficit cash)
exercise a DECLARED TWO-ITEM diagnostic policy. All eight selected purchases match
independent earliest/full original-state ApplyLogged references, including four
positive waits, receipts/events, increasing revisions and unchanged parents.
The chosen child's context survives its actual clone. A separately ranked sibling
has the same result; mutating its returned generator map cannot change the selected
branch's next purchase. Root/malformed-clone controls refuse. Severing ONLY the
real clone's context propagation makes all four purchase populations fail with
an unbound-head error; both bank refusal controls remain correctly refused.

This bridge accepts TRUE ACTION ORIGINS only. Two bank-only cases explicitly
refuse; they do not rebase a projected view, renew an allowance, select the legacy
wait or claim unreachability. Retained bank views need an origin-relative query
floor before integration. Fixed-mode diagnostic actions are not natural return/
session behavior. Actual beam/dedup/dominance/completion-cache/rollout reuse and
full policy/workloads remain unproved; the two-item population is NOT a narrowed
beam width or replacement for them. H3x/legacy checks still execute unchanged.
No live API/math/schema/report/CI change or normative acceptance follows.

### H4 — Compatibility, work and evidence

The separate proposed [Exact Represented Aggregation](exact-represented-aggregation.md) A6
binds numeric producers through existing kernel identity in NEXT report formats, without
pretending a numeric correction is a changed experiment input. It owns its explicit baseline-
guard amendment; this draft does not adopt it. Here legacy/origin tags describe accounting/
wait semantics, not a license to disguise a new numeric build as an old producer. If that
numeric follow-up is separately accepted, legacy wait scenarios may run under a new bound
kernel/report format; reproduction of old reports still requires their evidenced historical
source. Old formats keep their exact fields/readers. Allocate new formats in actual landing
order, including the separate later origin-protocol fields, without reusing a schema number
for different grammars. Until separate adoption, current numeric/report behavior remains.

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

- 2026-10-15: supply the fixed-exponent coefficient-fold sublemma by composition
  and a sort-position crossing argument. Keep normalization, group reclassification,
  ledger and temporal caller obligations open; no new census, algorithm or adoption.

- 2026-10-15: reduce the actual resource predicate to frozen quotes, completed
  bucket prefixes, the varying canonical partial delta and ledger behavior.
  Execute the bounded near-flat burst carry attack with four real-evaluator
  controls; no reversal is found, not a general proof or algorithm adoption.
- 2026-10-15: retain RP-664's positive aggregation/actual evaluator counterexample
  while distinguishing changed ownership from fixed-origin time. H3c rejects a
  false universal numeric lemma, not the unproved temporal algorithm. No repair,
  payout change, relaxed criterion or adoption follows.
- 2026-10-15: supply the concrete clock/buff start-set candidate and execute actual
  logged long-horizon probes, retaining the original expiry census. Reject the flat
  payout shortcut explicitly; exact search results are not exact-integer payout proof.
  Full producer/predicate/work and adoption gates remain open; no production change.
- 2026-10-15: bind actual fixed-price/resource predicates and distinguish the5000ms
  attendance edge from the24h automatic catchup edge. Execute36 exact logged profiles,
  including declared no-buff controls; retain general proof/work/adoption gates.
  No production, policy, search algorithm, report or authority change.
- 2026-10-15: align reuse with opaque typed run-local origins, not a new canonical
  origin codec/hash protocol. Preserve deterministic output, all reuse paths and
  full workload/budget gates; no implementation or adoption.
- 2026-10-15: specify proposed protocol/family dispatch, preserving legacy identities;
  separate new Relevance metadata from base policy grammar and make shared immutable
  content manifests single registry resources. No actual version/hash adoption or
  consumer implementation; complete boundary proof and review remain outstanding.
- 2026-10-15: split producer versus harness-search authority; retain RP-653/RP-656
  evidence, complete consumer obligations and unapplied status.
- 2026-10-15: distinguish accounting/wait protocol from the separately proposed numeric
  producer identity. Preserve exact old formats and historical-source reproduction; new
  kernel-aware reports/guard rules require their own acceptance. No numeric adoption.
- 2026-10-15: retain the canonical-input group-carry counterexample and distinguish
  its synthetic unequal prefix from an admitted equal-prefix raw reversal with equal
  ledger payments. Canonicalization does not supply the missing temporal proof;
  no new algorithm, bound, arithmetic or RFC acceptance follows.
- 2026-10-16: RP-676 reaches an actual committed-time decline via a provision-generated
  unequal prefix, including the neutral actual wait wrapper. Reject generic H3b-regime
  monotonicity; the unapplied exact-sum counterfactual fixes only this bounded case.
  Correct complete search/caller proof, compatibility, cost and adoption remain gates.
- 2026-10-16: RP-680 binds the admitted temporal decrease to an actual loaded standalone
  suite, catalog quote, affordability search and applied candidate-ranking arm. Extending
  the horizon by1ms suppresses a verified in-range purchase. No natural/report/full modern
  factory defect, globally earliest oracle, implementation or adoption is inferred.
- 2026-10-16: derive a bounded positive-aggregation upper enclosure and execute its
  finite power premises, retained decrease and meaningful omitted-margin controls.
  Keep actual temporal producer, complete search, work and adoption obligations open;
  this is not an algorithm implementation or a changed payout/equality tolerance.
- 2026-10-16: compose the enclosure with actual partial production and ledger in one
  fixed cell; compare chronological search with its complete exhaustive earliest
  reference. Keep whole-horizon cell construction, definedness, actual caller/work,
  compatibility and adoption gates explicit; no runtime change or RP-680 closure.
- 2026-10-16: construct diagnostic prefixes/cells from real contribution, rate and
  provision producers; compare12028 actual waits and retain control/guard failures.
  Keep general definedness, modern contexts, real workloads and adoption open;
  no active factory/search, budget or payout behavior changes.
- 2026-10-16: reproduce RP-681's hidden secondary-resource partial refusals, add
  constructor-local raw-identity partial certification and execute semantic cache
  alias controls. Keep final aggregation/ledger/error precedence, modern contexts,
  complete workloads and adoption separate; no production search or numeric change.
- 2026-10-16: reproduce RP-682's combined-payout overflow after valid partials;
  add all-resource prefix/interior-domain proof with exact inconclusive-case checks.
  Retain valid extreme inputs and explicit costs; ledger/modern/workload/adoption
  gates remain, with no runtime, arithmetic, budget or ordinary CI change.
- 2026-10-16: reproduce RP-683's ledger Add/receipt refusals after valid aggregates;
  add sufficient-case/exact-ledger certification and resource-bound successful
  reuse, checked by strict receipt replay. Progress/modern/hooks/workload and
  adoption gates remain; no active ledger, search, arithmetic or policy change.
- 2026-10-16: reproduce RP-684's post-ledger progress refusal and distinguish
  failed cloned-state mutation from persisted behavior. Add sufficient-case and
  actual progress fallback without excluding valid tiny weights; retain purchased
  versus provisioned count semantics. Whole caller/error ordering, modern/hooks,
  workloads and research/normative adoption remain open; no live behavior change.
- 2026-10-16: reproduce RP-685's eager-validation loss of earlier results; add a
  proposed chronological query with range certification/refinement, mandatory
  source stops, zero-time/empty-future semantics and explicit repeated work.
  Retain original nonmonotone populations. Caller compatibility, global/modern/
  hook domains, complete workloads and normative adoption still gate live use.
- 2026-10-16: reproduce RP-686's nonmonotone faction-stock hook validity on
  admitted Company19 and derive/test a concrete attendance-split hook component.
  Preserve event policy, complete encoded-state/cash boundaries and legacy guards.
  No modern factory/query integration, full logged-hook proof or adoption follows.
- 2026-10-16: bind an owned fixed-Company19 producer to chronological search and
  the real stock-domain/leaf recipe; reproduce missing binding first, then retain
  independent full-state/role and mutable-input isolation checks. Existing legacy
  guards remain; logged/other-hook/global/caller/workload and adoption gates stay open.
- 2026-10-16: H3r/RP-691 distinguishes actual primitive success from post-hook
  codec completion on admitted diagnostic inputs. Retain full-state/write-set
  observations, actual codec errors, valid neighbors and a conditional span
  obligation; no generic primitive refusal, live fix or normative adoption.
- 2026-10-16: H3s/RP-692 binds a manual prefix to observed actual logged
  catchup/schedule/ordinary stages and full Go/TS outcomes. Retain zero-elapsed
  ordinary evaluation, real clearing/rollback and malformed-context controls;
  no modern factory/query, production accounting change or normative adoption.
- 2026-10-16: H3t/RP-693 constructs a manual owned modern endpoint component,
  exposes/fixes raw-input rounding, and binds full-state exact reference queries.
  Optimized bounds, complete factory/caller/error policy, workload and normative
  adoption remain open; no live accounting or search change.
- 2026-10-16: H3u/RP-694 connects source-cell bounds to owned endpoint regimes,
  with validated zero schedule prefix, origin-preserving floors, actual first-error
  controls and exposed setup/fallback work. Full consumer/workload and normative
  adoption remain; accepted live binary search is unchanged.
- 2026-10-16: H3v grounds the actual current delayed candidate/action/score
  boundary in72 controlled first-hour cases. Preserve its bounded positive
  compatibility result without claiming modern-query integration or lost lifetime
  in this existing caller; full workload/dispatch and normative adoption remain.
- 2026-10-16: H3w adds the owned manual modern action component and actual
  full logged comparisons, preserving Founder input, post-action hooks and
  consumed-batch child semantics. Full caller/domain/workload and adoption remain;
  no live API or accepted contract change.
- 2026-10-16: H3x connects actual rankCandidate to owned query/action and
  proposed post-action endpoint-rate sampling in a reversible manual overlay.
  Real-quote windows, score/input faults and legacy fallthrough discriminate;
  retained child/beam/context reuse, complete work/dispatch and adoption remain.
- 2026-10-16: H3y carries selected child context through actual choice and
  clone boundaries, with independent two-action references and scoped failure
  controls. Bank views/beam/cache/rollout/full workloads remain; no live dispatch
  or normative adoption.
