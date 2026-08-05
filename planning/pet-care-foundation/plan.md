# Pet Care Foundation implementation plan

RFC: `rfc/pet-care-foundation.md`

- [x] Acceptance-review the draft and reconcile owner rulings C1-C8 into the normative body.
- [ ] Implement the reusable Founder-scoped persistence/replay boundary.
  - [x] Append immutable Founder command-log storage and typed replay command envelopes.
  - [ ] Add the pure `ApplyFounderLogged` transition and byte-parity replay fixtures.
- [ ] Implement strict fixture-only pet catalog/state schemas without production balance rows.
  - [x] Land C12a's literal stat/status/mood/behavior/event/rejection grammar, queue hardcap, and
    PRNG label with one shared Go/TypeScript parity fixture.
  - [x] Land C15's exact mood-threshold and behavior-candidate rows in Go and TypeScript against
    one shared fixture, without production balance rows.
  - [ ] Embed the already-ruled pet state in the Founder wire under C16's pinned artifact and
    independent Founder version axis.
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

Post-C14 status: ownership and the top-level state keys are ruled. C15 carries the two nested row
families whose exact keys are still absent; C16 carries the pinned-artifact/Founder-only version
transition required to activate that state without advancing or bricking Company saves.

Post-C16 infrastructure status: Exit now validates independently supplied Founder/Company floors
and accepts mixed-axis tuples without relaxing monotonicity or the decode-only v15 ban. Assigning
the first pet/minigame Founder versions and accepting their artifact bytes remains owner/mint
work; no code labels the partial fixture grammar as a production pet artifact.
