# Doctrine & Compute Credit implementation plan

RFC: `rfc/doctrine-and-compute-credit.md`

- [x] Resolve acceptance blockers D1-D6 and reconcile their primary direction in the normative body.
- [x] Resolve implementation blockers D7-D10: exact burst arithmetic, literal doctrine/gate data,
  legal Company-version activation, and enumerated wire grammar (including stale AC3).
- [x] Implement the doctrine catalog/activation boundary and exact pick intent in Go and TS.
- [x] Implement the ruled Compute Credit burst lifecycle in Go and TS.
- [x] Add exact wire/event/replay fixtures and persistence migrations required by the rulings.
- [x] Run normal root verification from the repository root.
- [ ] Obtain both required full-range review gates.
- [x] Update canonical docs to match the implemented pre-mint behavior.
- [ ] Archive only after all acceptance criteria and review ranges close.

Current status: implementation and canonical docs are complete and the normal root gate is green.
The production doctrine artifact and v17 activation remain deliberately unminted until the owner
supplies T3-to-T4 content and the paired Meters/Achievements production artifacts. Independent
full-range review remains open; this job is not archival-eligible yet.
