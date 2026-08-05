# Doctrine & Compute Credit implementation plan

RFC: `rfc/doctrine-and-compute-credit.md`

- [x] Resolve acceptance blockers D1-D6 and reconcile their primary direction in the normative body.
- [ ] Resolve implementation blockers D7-D10: exact burst arithmetic, literal doctrine/gate data,
  legal Company-version activation, and enumerated wire grammar (including stale AC3).
- [ ] Implement the doctrine catalog/activation boundary and exact pick intent in Go and TS.
- [ ] Implement the ruled Compute Credit burst lifecycle in Go and TS.
- [ ] Add exact wire/event/replay fixtures and persistence migrations required by the rulings.
- [ ] Run normal root verification and obtain both required full-range review gates.
- [ ] Update canonical docs and archive only after all acceptance criteria and review ranges close.

Current blocker: D1-D6 choose the correct direction, but D7-D10 remain executable-contract gaps.
No burst economics, production doctrine gate, save activation history, or wire payload may be
invented in code.
