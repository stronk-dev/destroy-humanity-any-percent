# Meters Foundation implementation plan

RFC: `rfc/archive/meters-foundation.md`

- [x] Reconcile owner-approved C1–C12 into one implementable Company-scope contract.
- [x] Land strict meter schema/loaders and discriminating test catalogs in Go and TypeScript.
- [x] Land save v15 maps, migration corpus, canonical wire closure, and reset assembly.
- [x] Implement exact attended-time decay and the closed causal input union inside `ApplyLogged`.
  - [x] Land the pure Go/TypeScript transition kernel and shared parity corpus.
  - [x] Bind paired artifact/state validation and new-run activation to Go/TypeScript replay.
  - [x] Bind the kernel to run-pinned catalogs at the live/replay boundary.
- [x] Add band events, transport/schema registries, formula output, replay bundle identity, and
  cross-runtime sequential fixtures.
- [x] Mint the production meter artifact after owner-supplied literal band/initial/rate/input data.
- [x] Update canonical docs, pass normal root verification and Postgres integration, obtain an
  independent full-range adversarial verdict, and archive.

The literal production balance rows shipped in owner-authorized epoch 6 and are pinned by the
approved First Content mint verdict.

Activation ruling C13 is implemented: Meters and Achievements activate atomically as save v16 at
the first new-run boundary pinned to both artifacts. Legacy runs remain v14. The live/replay hook,
events, formulas, identity, reset assembly, production content, and final acceptance are closed.
