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
  - [x] Embed the already-ruled pet state in Founder v18 under the complete pinned artifact and
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

Blocked implementation contract carried explicitly: C17 must order pet/minigame activation on
the scalar Founder axis (or replace it with a feature vector) and enumerate the complete pet
artifact wire before the C14 map can be persisted.

Post-C17 status: [x] the scalar v17→v18 chain, complete pet-artifact loader in both runtimes,
replay-owned v18 map, artifact biconditional, and reachable mixed Founder/Company version path are
implemented. Production pet rows and care transitions remain deferred balance/content work.

Post-F1 acceptance pass: care-transition consumers are blocked on C18-C21. v18 lacks the per-pet
attended evaluation watermark required to prevent repeated decay; the ruled remainder fields have
no equation; care eligibility/diminishing order and mood/FSM scheduling are not exact; and the
Founder command/replay/event envelopes remain unnamed. The proposals are filed in the RFC rather
than inferred in code.
