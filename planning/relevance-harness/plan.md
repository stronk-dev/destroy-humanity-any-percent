# Relevance Harness implementation plan

RFC: `rfc/relevance-harness.md`

- [x] Resolve acceptance blockers R1-R8 and reconcile their primary direction.
- [ ] Resolve implementation blockers R9-R15 and reconcile the normative body/wire in the ruling
  edit.
- [ ] Extend the authoritative harness with the ruled reference and beam policies.
- [ ] Implement paired ablation runs, group attribution, exact report schema, and budget preflight.
- [ ] Wire the fail-closed relevance gate and its BALANCE-CHANGE artifact discipline.
- [ ] Prove the generic layer with a discriminating test-only content fixture.
- [ ] Carry the first production scenario/baseline gate explicitly to the T0-T1 content mint.
- [ ] Run normal root verification and obtain both mandatory full-range reviews before archival.

Current blocker: R9-R15. The implemented ablation seam is sufficient, but the authoritative wait
edge does not exist and the reached-state, solver arithmetic, policy/report bytes, budget, and
window binding remain non-unique. They are not to be inferred in code.
