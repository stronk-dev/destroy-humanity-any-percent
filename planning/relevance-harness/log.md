# Relevance Harness — running log

## 2026-08-03 — pre-acceptance implementation review

- **Review by:** Codex
- **Recorded by:** Codex
- **Reviewed surface:** uncommitted draft `rfc/relevance-harness.md` V1–V5,
  `rfc/t0-t1-playable-content.md`, and the shipped economy/save/production/harness surfaces at
  `8e6029c`.
- **Verdict:** not implementable yet; implementation did not start.

The counterfactual report and ANY/ALL gate are sound goals, but V5 refers to authoritative
content and transition seams that do not exist at HEAD. Implementing them inside the harness
would improvise a second economy engine and violate RFC-0000.

### Acceptance blockers

1. **Purchasable model.** The economy catalog has generator classes only. It has no upgrade
   definitions, purchasable union, category/upgrade-family field, availability window,
   `trap_exempt`, role declaration, synergy declaration, or staggered milestone ladder.
   Save v13 persists generator counts only; the closed intent union has no upgrade purchase.
2. **Role execution.** The design's closed role names describe six different owner systems, but
   no catalog schema or runtime executes those bindings. Loader validation cannot be added until
   each accepted role variant defines its inputs, effect, activation record, and owning package.
   Counting a declaration as an activation would be placeholder math and would make the role
   floor self-confirming.
3. **Purchased/generated split.** State has one purchased generator count. No generated count,
   chain-provisioning formula, accrual rule, save migration, receipt, event, or Go/TS parity
   contract exists. T0-T1 requires this while simultaneously claiming it needs no engine code.
4. **Ablation boundary.** Production has no contribution-layer `effect_mask`. The harness cannot
   mutate authoritative catalog effects ad hoc: live and replay must continue to share one
   `ApplyLogged` transition, with the mask and trace supplied through a closed, harness-only
   evaluation policy that production cannot obtain from client input.
5. **Harness boundary mismatch.** The current harness calls `production.Transition`, not
   `ApplyLogged`. V5 must specify the deterministic `ReplayInputs` and full catalog bundle used by
   the solver, plus how rejected transitions and emitted events participate in the report.
6. **Solver definition.** Width eight and full-run depth do not determine a beam search. The RFC
   must define action expansion (including wait), wait-event boundaries, objective milestone,
   state canonicalization/deduplication, dominance pruning, termination, and the exact integer
   greedy-gap calculation. Payback also needs zero-rate, unaffordable, indirect-effect, and
   multi-resource semantics.
7. **Report/gate arithmetic.** The shipped scenario has four milestones and sixteen
   persona/statistic observations, not sixteen milestones. The report must bind to milestone IDs,
   define which milestone inside a window supplies `delta_ms`, and specify comparisons for an
   unreached baseline as well as an unreached ablation.
8. **Group and exemption gates.** AC2 says a near-substitute pair must fail group ablation, but V3
   gives no group floor. `trap_exempt` also needs an exact changelog/guard rule; prose that a line
   exists is not enforceable by the existing balance guard.
9. **Run budget.** V2 says one run per purchasable/persona while the shipped chaos/casual
   definitions contain 200/100 seeds. The RFC must say whether relevance uses one canonical seed,
   the full seed sets, or an aggregation, then count baseline, LOO, removal, group, and beam runs
   with one formula enforced by `relevance_budget_max_runs`.
10. **Epoch cadence.** `harness-check` runs per commit. “Per epoch, not per commit” needs a
    machine-detectable skip condition tied to unchanged artifact identities, or the wording should
    require deterministic re-execution on every check and restrict only baseline regeneration to
    epoch changes.

### Required scope split

Before this RFC can be accepted, split and accept a **Purchasable Content Foundation** owner RFC
for the upgrade lifecycle, purchased/generated counts, synergy/role schemas and execution, and
their Go/TypeScript/save/replay contracts. The Relevance RFC then owns only the harness evaluation
policy, solver, report, and gates. T0-T1 becomes declarative content on those implemented schemas;
its “no new engine code” claim is true only after that prerequisite lands.

The owner may instead explicitly narrow the first relevance release to generator production
only, defer upgrades/roles/chains, and weaken the T0-T1 acceptance criteria. That is a material
design choice, not an implementation default.

## 2026-08-05 — Codex acceptance review: blocked (R1-R8)

Review by: Codex. Recorded by: Codex.

Verified at HEAD: Purchasable Content Foundation is implemented; `SimulateTransition` owns the
closed effect-mask/action-removal seam; typed role activations are emitted only for executed
non-neutral mechanics; and the base harness already supplies deterministic run identity, dispatch,
reports, and balance-change guarding. The earlier foundation blocker is therefore stale as claimed.

The draft still cannot produce one canonical report or gate verdict. R1-R8 cover the reference
policy, beam oracle, seed/delta reduction, group-vs-individual contradiction, tier attribution,
missing relevance metadata, exact report/budget wire, and the generic-vs-real-content acceptance
split. The review also confirms that the current production Phase-0 catalog remains schema v3 and
cannot honestly be called the first relevance baseline. No implementation started pending rulings.

## 2026-08-05 — owner rulings on acceptance blockers R1-R8 (all resolved)
- R1: reference_greedy = event-driven clone-and-simulate to the next horizon; rank
  (payback, earliest_benefit_time, raw_byte_id) w/ sentinels; resolves "indirect effects" via real
  simulation, not instantaneous delta_rate.
- R2 (owner): beam oracle — node=sim-state digest, expand R1 candidates + wait-edge, score=time to
  terminal milestone, Pareto pruning, dedup, width 8; greedy_gap_ppm=floor((greedy-beam)*1e6/beam);
  fixture oracle first, production check gated on T0-1's primary milestone; 50000 ppm is data.
- R3: pair each ablation to one baseline key; delta_t=ablated-baseline; conservative seed reducer
  (p05/worst); 4 reached/unreached rules; ANY over personas post-reduction. Stale "16 milestones"
  reconciled -> "every milestone the scenario declares".
- R4 (owner): pass if individual LOO OR one declared group's delta passes (label group_supported);
  trap floor stays INDIVIDUAL. AC2 corrected in body.
- R5 (owner): rename "shares" -> "counterfactual tier contribution" = group-ablation delta, NOT
  required to sum to 100%; {from_gate,to_gate}=[from inclusive,to exclusive); nearest-passing
  diagnostic defined. Body reconciled.
- R6 (owner): a hash-pinned `relevance_policy` artifact keyed by purchasable ID
  {window,epsilon,trap_exempt+justification,group_ids} — NOT economy-catalog grammar. Body V2
  reconciled (window is a policy-artifact field, not a catalog field).
- R7: enumerate report envelope/rows, no maps/floats, run identity += ablation_mode/target, raw-byte
  sort, exact preflight cardinality (baseline+effect-mask+action-removal+group+reference+beam),
  fail-before-dispatch on budget; dup/missing + cardinality fixtures.
- R8: implement+archive the generic solver/report/gate vs a test-only schema-v4 fixture; T0-1 mint
  adds the first production relevance policy/scenario+golden report AND makes harness-check fail
  closed if an active schema-v4+ catalog lacks them; schema-v3 is NOT a real baseline.
Status -> accepted; implementing. Body/README reconciled.

## 2026-08-06 — implementation acceptance recheck: blocked (R9-R15)

Review by: Codex. Recorded by: Codex.

Re-read the ruled RFC against `server/harness/harness.go`, the v1 scenario/report schemas, and
`production.SimulateTransition`. The ablation and typed-role seams are real, but R1's
earliest-affordable search and R2's wait edge have no authoritative transition: the only simulation
entrypoint applies an intent. Seven residual blockers are filed in the RFC with proposed contracts:

- R9: V5 finite-unreached vs R3 `+infinity` contradicts R7's no-float wire; reducer still has two
  legal choices.
- R10: no harness-only advance boundary and no exact payback quantity/rounding for indirect or
  multi-resource effects.
- R11: beam score/rollout/termination and internal-work bound remain non-unique.
- R12: the hash-pinned relevance-policy artifact has no exact grammar or group authority.
- R13: the report envelope and ordered row families were never actually enumerated.
- R14: the promised exact run-budget formula is absent.
- R15: current milestones cannot bind gate windows; trap/role evidence and the T0-T1 fail-closed
  registry are not mechanically owned.

No code was started. Implementing past these gaps would invent a new production advance boundary,
immutable golden bytes, and balance-gate arithmetic. Proposed contracts reuse the existing harness
discipline, add a guarded `SimulateAdvance` sibling, and narrow the first greedy/beam oracle to a
single-resource fixture unless the owner supplies a multi-resource conversion.

## 2026-08-06 — owner rulings on the second-round blockers R9-R15 (all accepted)
R1-R8 set the shapes but left exact arithmetic/encoding/schemas open; Codex bounced R9-R15 with
executable contracts. All accepted:
- R9: reached/unreached = closed row {status, baseline_ms, ablated_ms, delta_ms}; FINITE
  ablated_unreached delta = horizon-baseline (SUPERSEDES my R3 +inf error, which also broke R7
  no-floats); ONE declared reducer per scenario (p05|worst), fixture=worst (R3's "either" corrected).
- R10: add harness-only production.SimulateAdvance (advance clone w/o action, no intent/receipt/rev,
  source-guarded); earliest-affordable by lower-bound binary search; single-resource value:
  gross_marginal = candidate+cost-baseline; payback_ms = wait + ceil(cost*(horizon-earliest)/gross)
  exact Decimal; rank (payback, earliest_positive_delta, raw_byte_id); multi-resource = successor.
- R11: bounded beam — node={state_hash, virtual_ms}, expand candidates + next-horizon edge, score by
  reference-greedy rollout (reached<unreached, then time, then path), componentwise dominance at equal
  virtual_ms; max_decisions + beam_width(8); internal sims => separate relevance_budget_max_transitions.
- R12: relevance_policy schema v1 {schema_version, items[], groups[]}; item keys incl. window/epsilon/
  trap_exempt/justification/group_ids; trap_exempt IFF justification non-null; groups {group_id, axis
  in tier|category|declared, member_ids, epsilon_ms}; Go/TS mutation vectors.
- R13: report schema v1 root enumerated (arrays only, no maps/floats); item rows w/ support
  individual|group_supported|failed + nearest-passing; R9 tagged delta rows; role rows reuse existing
  shape; JSON schema + fixtures + mutations.
- R14: run budget = (E+R)*(1+I+G) + R*I + (beam?R:0); checked ints; fail-before-dispatch; beam-internal
  bounded by R11 transitions; declared==executed required.
- R15: scenarios add segments[{milestone_id,from_gate,to_gate}] (Routes-validated); trap evidence =
  baseline purchase count (action-removal is diagnostic, cannot rescue); role evidence = unmasked
  baseline activations post-reduction; scenario registry checked-in, T0-1 extends fail-closed.
Status -> accepted; R1-R15 ruled; implementing. Body (R3) reconciled via R9 supersession. README updated.
