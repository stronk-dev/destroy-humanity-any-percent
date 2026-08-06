# Fiscal Quarters Foundation implementation plan

RFC: `rfc/fiscal-quarters-foundation.md`

- [ ] Resolve F1-F8 and reconcile every ruling into the normative specification (rulings exist;
  implementation proof remains pending with F9-F15).
- [ ] Resolve implementation blockers F9-F15 and reconcile their exact persistence/wire contracts.
- [ ] Implement the exact fiscal artifact and Founder v19 activation/save chain.
- [ ] Implement lazy phase-preserving auto-report and deterministic manual harvest in Go and TS.
- [ ] Implement server-priced spend targets and the ruled Company-effect boundary.
- [ ] Add exact wire/event/replay vectors, migration corpus, cap and clock-regression boundaries.
- [ ] Run normal root gates and obtain both full-range review gates before docs/archive.

Current blocker: F9-F15. The gameplay shape is ruled, but the shipped Founder/store boundaries cannot
yet represent a sweep that commits beside a rejected action, and frozen Founder contributions have
no immutable run-owned storage. Exact artifact, seed, v19, wire, and Quarter-offer ownership also
remain owner work. No Fiscal mechanic code has started.
