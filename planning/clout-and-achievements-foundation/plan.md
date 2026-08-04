# Achievements Foundation implementation plan

RFC: `rfc/clout-and-achievements-foundation.md`

- [x] Reconcile the owner-approved C1–C10 rulings into one Clout-free contract.
- [x] Land the strict schema/loaders, bounded predicate/proof union, copy-key composition, and
  score derivation in Go and TypeScript.
- [ ] Add save-v16 Company run and Founder lifetime state with migration corpus.
- [ ] Evaluate achievements at the shared live/replay boundary and emit ordered earned events.
- [ ] Settle run achievements atomically at Exit and reset the next run.
- [ ] Join epoch identity/replay fixtures when literal production content is minted.
- [ ] Pass normal root verification and Postgres integration, obtain an independent full-range
  adversarial verdict, update canonical docs, and archive.

Carried content dependency: literal production achievement rows are owner-authored T0–T1 content;
the strict engine uses discriminating fixtures and does not invent achievements.
