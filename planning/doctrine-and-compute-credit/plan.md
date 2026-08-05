# Doctrine & Compute Credit implementation plan

RFC: `rfc/doctrine-and-compute-credit.md`

- [x] Resolve acceptance blockers D1-D6 and reconcile their primary direction in the normative body.
- [x] Resolve implementation blockers D7-D10: exact burst arithmetic, literal doctrine/gate data,
  legal Company-version activation, and enumerated wire grammar (including stale AC3).
- [ ] Implement the doctrine catalog/activation boundary and exact pick intent in Go and TS.
- [ ] Implement the ruled Compute Credit burst lifecycle in Go and TS.
- [ ] Add exact wire/event/replay fixtures and persistence migrations required by the rulings.
- [ ] Run normal root verification and obtain both required full-range review gates.
- [ ] Update canonical docs and archive only after all acceptance criteria and review ranges close.

Current status: D1-D10 are ruled and the normative body is reconciled. Implementation is unblocked;
the production doctrine artifact and v17 activation remain deliberately unminted until the owner
supplies T3-to-T4 content and the paired Meters/Achievements production artifacts.
