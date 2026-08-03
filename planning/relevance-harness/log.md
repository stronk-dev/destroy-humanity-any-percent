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

