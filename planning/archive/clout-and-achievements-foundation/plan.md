# Achievements Foundation implementation plan

RFC: `rfc/archive/clout-and-achievements-foundation.md`

- [x] Reconcile the owner-approved C1–C10 rulings into one Clout-free contract.
- [x] Land the strict schema/loaders, bounded predicate/proof union, copy-key composition, and
  score derivation in Go and TypeScript.
- [x] Add save-v16 Company run and Founder lifetime state with migration corpus.
- [x] Evaluate achievements at the shared live/replay boundary and emit ordered earned events.
- [x] Settle run achievements atomically at Exit and reset the next run.
- [x] Join paired epoch identity and active Go/TypeScript replay fixtures without minting content.
- [x] Pass normal root verification and Postgres integration, obtain an independent full-range
  adversarial verdict, update canonical docs, and archive.

The owner-authored production achievement rows and their verified copy bindings shipped in epoch 6.

Activation ruling C11 is implemented: paired artifacts activate save v16 only for the next run;
pre-artifact runs finish under v14. Earning is live in the shared kernels and the production
artifact is pinned by epoch 6.
