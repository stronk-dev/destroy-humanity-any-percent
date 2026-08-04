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

Replay-inputs v4 freezes the Founder lifetime set and score on every active-run command, so career
conditions never read mutable Founder state during replay. Historical active v3 rows remain
replayable under their pinned semantics, and seven-artifact runs retain v2 compatibility. Both
replay loaders accept only the exact
legacy seven-artifact bundle or the paired nine-artifact bundle containing both Meters and
Achievements. The shared Go-authored corpus covers non-empty carry, unknown-ID and score-mismatch
rejection, active-state preservation, active Exit, and a next-catalog score-grant retune.

A package gate prevents the foundation from importing ledger spending, production/persistence
owners, or lifetime-Clout symbols. Active v4 commands evaluate once after Meters from one
post-action snapshot, stage simultaneous earns in byte order, and emit exact
`achievement_earned.v1` events. Burn proofs additionally require the declared same-batch event and
an actual minimum resource debit. Terminal evaluation includes the settling Exit in career age,
fact, and exit-count observations before the run set is unioned into Founder lifetime state.
Historical v3 rows retain their pre-hook semantics. Production definitions and the epoch mint
remain pending.
