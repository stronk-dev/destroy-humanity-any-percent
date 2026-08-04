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
condition evaluation, lifetime/run latching, and score derivation. Save v16 persists run IDs/score
on Company and lifetime IDs/score on Founder. Exit settles the run set exactly once, re-derives the
lifetime score under the next pinned catalog, and resets the new run.

Replay-inputs v3 freezes the Founder lifetime set and score. Active runs require v3 globally;
historical seven-artifact runs retain v2 compatibility. Both replay loaders accept only the exact
legacy seven-artifact bundle or the paired nine-artifact bundle containing both Meters and
Achievements. The shared Go-authored corpus covers non-empty carry, unknown-ID and score-mismatch
rejection, active-state preservation, active Exit, and a next-catalog score-grant retune.

A package gate prevents the foundation from importing ledger spending, production/persistence
owners, or lifetime-Clout symbols. Production achievement content, ordered earning events, and
the live achievement-evaluation hook remain pending; no artifact has been minted.
