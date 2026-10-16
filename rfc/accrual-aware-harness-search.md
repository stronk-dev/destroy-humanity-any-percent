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
