# Meters Foundation implementation plan

RFC: `rfc/meters-foundation.md`

- [x] Reconcile owner-approved C1–C12 into one implementable Company-scope contract.
- [x] Land strict meter schema/loaders and discriminating test catalogs in Go and TypeScript.
- [ ] Land save v15 maps, migration corpus, canonical wire closure, and reset assembly.
- [ ] Implement exact attended-time decay and the closed causal input union inside `ApplyLogged`.
  - [x] Land the pure Go/TypeScript transition kernel and shared parity corpus.
  - [ ] Bind the kernel to run-pinned catalogs at the live/replay boundary.
- [ ] Add band events, transport/schema registries, formula output, replay bundle identity, and
  cross-runtime sequential fixtures.
- [ ] Mint the production meter artifact after owner-supplied literal band/initial/rate/input data.
- [ ] Update canonical docs, pass normal root verification and Postgres integration, obtain an
  independent full-range adversarial verdict, and archive.

Carried acceptance debt: the literal production balance rows are a DESIGN-GAP and block only the
epoch mint/final archive, not schema, state, engine, or parity implementation.

Activation blocker found during save-v15 implementation: C13 in the RFC. Complete-key v15 state
cannot be written for a run pinned to an epoch without a meter artifact. Save activation/live hook
binding wait for the owner to rule new-run-only activation versus a legacy-value preservation
source. Pure catalog/transition work remains valid and independently approved.
