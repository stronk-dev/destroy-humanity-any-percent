# Achievements

> **Run-local attainment (Clout v1, fixture-first):** under an economy declaring `axis_stack`, the
> hook's second pass attains run-scoped achievements per run whether or not the Founder owns them for
> life. This emits `achievement_reattained.v1` and feeds the PR Intern axis. It mints no Clout. See
> [Axis stack](axis-stack.md).

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

Go and TypeScript loaders enforce the same shape and registry bindings.
Root and row member names are exact in both loaders; Go checks the original row object before
case-insensitive struct decoding. Escaped canonical names and existing last-value duplicate
semantics are preserved. No TypeScript numeric-token policy was changed by this correction.
Kernel 0.3.191 records the genuine admission change, not an achievement grant or balance retune.

`achievements.AchievementSchemas()` exports API C18's structural v1 owner descriptor, with a
closed proof union and finite depth-specific condition definitions. It validates current and
stored artifact bytes without normalization. Registry/copy/proof bindings, row order, array
cardinality and the 64-node condition bound remain mandatory semantic-loader checks. Exported
definitions are fresh; this does not register public catalogs or replace the complete replay loop.

Both runtimes expose pure
condition evaluation, lifetime/run latching, and score derivation. Save v16 persists run IDs/score
on Company and lifetime IDs/score on Founder. Exit settles the run set exactly once, re-derives the
lifetime score under the next pinned catalog, and resets the new run.

Replay-inputs v4 freezes the Founder lifetime set and score on every active-run command, so career
conditions never read mutable Founder state during replay. Historical active v3 rows remain
replayable under their pinned semantics, and seven-artifact runs retain v2 compatibility. The
foundation activation requires the paired Meters/Achievements artifacts; later accepted
foundations extend the complete bundle rather than permitting one half of that pair.
The shared Go-authored corpus covers non-empty carry, unknown-ID and score-mismatch
rejection, active-state preservation, active Exit, and a next-catalog score-grant retune.

A package gate prevents the foundation from importing ledger spending, production/persistence
owners, or lifetime-Clout symbols. Active v4 commands evaluate once after Meters from one
post-action snapshot, stage simultaneous earns in byte order, and emit exact
`achievement_earned.v1` events. Burn proofs additionally require the declared same-batch event and
an actual minimum action debit measured from the post-accrual balance boundary, so ordinary idle
income cannot hide or fabricate a sink. Terminal evaluation includes the settling Exit in career
age, fact, and exit-count observations before the run set is unioned into Founder lifetime state;
both kernels reject any run/lifetime ownership overlap before settlement.
Historical v3 rows retain their pre-hook semantics. Epoch 6 pins twelve production definitions
with exact score grants of 2, 4, or 8 and verified Copy Pipeline bindings; epochs 1–5 remain on
their earlier semantics. The production rows cover attended-time and Exit career milestones plus
run-scoped gate and generator milestones.
