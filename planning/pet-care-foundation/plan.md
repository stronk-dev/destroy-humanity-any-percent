# Pet Care Foundation implementation plan

RFC: `rfc/pet-care-foundation.md`

- [x] Acceptance-review the draft and reconcile owner rulings C1-C8 into the normative body.
- [ ] Implement the reusable Founder-scoped persistence/replay boundary.
  - [x] Append immutable Founder command-log storage and typed replay command envelopes.
  - [ ] Add the pure `ApplyFounderLogged` transition and byte-parity replay fixtures.
- [ ] Implement strict fixture-only pet catalog/state schemas without production balance rows.
  - [x] Land C12a's literal stat/status/mood/behavior/event/rejection grammar, queue hardcap, and
    PRNG label with one shared Go/TypeScript parity fixture.
  - [ ] Land the exact pet catalog and Founder-state wire after their complete key sets are ruled.
- [ ] Implement attended-grid care, trust, mood, and bounded behavior FSM kernels in Go and TS.
- [ ] Add Founder activation/persistence, public status projection, and combat-input seam.
- [ ] Mint production pet content only after owner/harness-supplied catalog rows.
- [ ] Pass normal root gates, obtain full-range independent review, publish canonical docs, and
  archive.

Carried content dependency: starter species/temperament and all care/decay/trust/FSM numeric rows
are balance/content data. Structural boundaries and discriminating fixtures land first.

Blocked implementation contracts carried explicitly: C9 Founder replay segmentation, C10
attended-time authority, C11 Founder-only persistence/activation wire, and C12 closed mood/FSM/
status unions. The implemented persistence envelope remains valid independently of those rulings.

Post-C12a status: the literal enums, queue bound, and PRNG label are implemented. Pet catalog and
Founder-state serialization remain blocked because C3/C11 name their contents without enumerating
the exact action/decay rows, care-state keys, remainder maps, cooldown representation, behavior
queue entry shape, or bond-graph wire.
