# Relevance Harness implementation plan

RFC: `rfc/relevance-harness.md`

- [x] Resolve acceptance blockers R1-R8 and reconcile their primary direction.
- [x] Resolve implementation blockers R9-R15 and reconcile the normative body/wire in the ruling
  edit.
- [ ] Extend the authoritative harness with the ruled reference and beam policies.
- [ ] Implement paired ablation runs, group attribution, exact report schema, and budget preflight.
- [ ] Wire the fail-closed relevance gate and its BALANCE-CHANGE artifact discipline.
- [ ] Prove the generic layer with a discriminating test-only content fixture.
- [ ] Carry the first production scenario/baseline gate explicitly to the T0-T1 content mint.
- [ ] Run normal root verification and obtain both mandatory full-range reviews before archival.

Current work: implement the now-ruled R1-R15 contract. The production wait seam, solver, canonical
policy/report bytes, budget preflight, and fixture gate must land together before any production
schema-v4 catalog can opt in.
