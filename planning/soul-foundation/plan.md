# Soul Foundation implementation plan

RFC: `rfc/soul-foundation.md`

- [x] Resolve SB1-SB9 and reconcile the normative body in the same ruling edit.
- [x] Resolve implementation blockers SB10-SB16, including once-only state and the Company
  suppression authority.
- [x] Wait for the exact Fiscal v19 artifact/activation chain before enabling Soul v20.
- [ ] Resolve implementation blockers SB17-SB23: literal artifact enums, debit errors/events,
  recovery persistence/commands, the suppressed transition, v20 activation bytes, consumer schema
  versions, and cross-stream event order.
- [ ] Implement the strict Soul artifact, dormant-to-active save transition, and pure band/gate API.
- [ ] Implement the ruled atomic debit component and real recovery activity coordinator.
- [ ] Add pet/minigame consumers, exact wire/event/replay vectors, and migration corpus.
- [ ] Route the Trust/Soul correlation and ending-copy obligations to named owners.
- [ ] Run normal root gates and obtain both mandatory full-range reviews before docs/archive.

Current blocker: Fiscal v19 is closed, but SB17-SB23 remain owner-ruling-required. SB10-SB16 settle
the architecture without enumerating the byte contracts their own acceptance text requires. No Soul
code or production drain/recovery rows may land until those contracts are reconciled.
