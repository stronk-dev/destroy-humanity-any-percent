# Achievements

The Achievements foundation owns permanent achievement IDs and exact integer achievement score.
It does not mint Clout, import lifetime Clout into production, or expose achievement score as an
economy resource.

Catalog schema v1 defines byte-sorted achievement rows with run- or career-scoped conditions, a
bounded predicate tree, one proof discipline, a positive exact score grant, and a Copy Pipeline
key. The Phase-A condition union is fact presence, registered counters, exit count, generator
ownership, and bounded conjunction. Meters predicates are intentionally absent until the Meters
artifact is pinned.

Proof rows are closed:

- provenance names sorted immutable event kinds and may not justify current possession;
- burn names the event, economy resource, and positive canonical amount consumed by the earning
  transition;
- possession requires a generator-ownership predicate and an explicit justification copy key.

Go and TypeScript loaders enforce the same shape and registry bindings. Both runtimes expose pure
condition evaluation, lifetime/run latching, and score derivation. A package gate prevents the
foundation from importing ledger spending, production/persistence owners, or lifetime-Clout
symbols. Production content, save v16, Exit settlement, and the shared live/replay hook remain the
next implementation batches.
