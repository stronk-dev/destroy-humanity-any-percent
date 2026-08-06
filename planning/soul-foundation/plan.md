# Soul Foundation implementation plan

RFC: `rfc/soul-foundation.md`

- [x] Resolve SB1-SB9 and reconcile the normative body in the same ruling edit.
- [x] Resolve implementation blockers SB10-SB16, including once-only state and the Company
  suppression authority.
- [x] Wait for the exact Fiscal v19 artifact/activation chain before enabling Soul v20.
- [x] Resolve implementation blockers SB17-SB23: literal artifact enums, debit errors/events,
  recovery persistence/commands, the suppressed transition, v20 activation bytes, consumer schema
  versions, and cross-stream event order.
- [x] Implement the strict Soul artifact, dormant-to-active save transition, and pure band/gate API.
- [ ] Implement the ruled atomic debit component and real recovery activity coordinator. The debit
  component is complete; recovery persistence/replay/atomicity are complete, but eligible recovery
  cannot progress beyond the online catch-up tolerance without a ruled attendance-advancement seam.
- [x] Add pet/minigame consumers, exact wire/event/replay vectors, and migration corpus.
- [ ] Route the Trust/Soul correlation and ending-copy obligations to named owners.
- [ ] Run normal root gates and obtain both mandatory full-range reviews before docs/archive.

Current blocker: the SB14 lifecycle requires an exclusive recovery to become eligible after a
Founder-attended duration while ordinary Company commands reject without mutation. The only shipped
mid-run attendance authority advances through successful Company evaluations and classifies a gap
larger than `catchup_ceiling_ms` as offline. Consequently an exclusive recovery has no legal writer
that can advance its attendance/evaluated cursor; after the tolerance elapses, every future resolve
remains ineligible. A new ruled presence/heartbeat or suppressed-progress boundary is required before
the recovery coordinator can be called complete or a production recovery row can mint.
