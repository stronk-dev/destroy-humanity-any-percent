# Pet Care Foundation implementation plan

RFC: `rfc/pet-care-foundation.md`

- [x] Acceptance-review the draft and reconcile owner rulings C1-C8 into the normative body.
- [ ] Implement the reusable Founder-scoped persistence/replay boundary.
  - [x] Append immutable Founder command-log storage and typed replay command envelopes.
  - [ ] Add the pure `ApplyFounderLogged` transition and byte-parity replay fixtures.
- [ ] Implement strict fixture-only pet catalog/state schemas without production balance rows.
- [ ] Implement attended-grid care, trust, mood, and bounded behavior FSM kernels in Go and TS.
- [ ] Add Founder activation/persistence, public status projection, and combat-input seam.
- [ ] Mint production pet content only after owner/harness-supplied catalog rows.
- [ ] Pass normal root gates, obtain full-range independent review, publish canonical docs, and
  archive.

Carried content dependency: starter species/temperament and all care/decay/trust/FSM numeric rows
are balance/content data. Structural boundaries and discriminating fixtures land first.
