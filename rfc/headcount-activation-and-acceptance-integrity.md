# RFC: Headcount Activation & Acceptance Integrity

- **Status:** draft — technical proposal, not implementation authority
- **Author:** Codex (for author/owner reconciliation)
- **Created:** 2026-11-02
- **Parent / amends:** Headcount Allocation S1/S4/S7/S8, AC2/AC3/AC7/AC8; Tier 2 H/E/P integration references
- **Design refs:** `design/06-tech.md` (versioned, server-authoritative state); unchanged Headcount design basis
- **Planning:** `planning/headcount-allocation/`

## Outcome and scope

Make the already-selected paid-hiring engine implementable without reusing
Clout's identities, silently activating Clout, changing old saves, or demanding
an impossible TypeScript overflow-output failure. This proposes technical
reconciliation only. It does not adopt new mechanics, rates, prices, copy,
full-policy behavior, balance or release scope. Parent bodies stay unreconciled
until the applicable author/owner direction; technical approval is not adoption.

## Current source evidence

At `59a56f16`, `server/economy/catalog.go` declares catalog schema5, including
Clout axis-stack semantics; Company19 in `server/save/state.go` owns attainment;
`server/save/runlog.go` declares replay inputs12. Headcount's old v5/v19/v9
literals therefore cannot be implemented as written.

There is a second, distinct integration constraint (RP-702):

- `CatalogBundle.versionFloors` in `server/production/replay.go` activates
  Company19 from `AxisStackDeclared`, not the economy version alone.
- `ValidateFoundationState` in `foundations.go` treats Company floors >=19 as
  attainment-bearing; new-run initialization creates attainment only for axis
  catalogs. `attainRun` executes whenever its map is non-nil.
- Go `wireSnapshot` in `intents.go` and TS `wireSnapshot` in `client/src/replay.ts`
  project the axis arm for Company versions >=19; both axis projectors refuse
  a catalog without that feature.

Thus simply replacing Headcount19 with20 would be insufficient. This is a
source-proven pre-integration constraint, not a current admitted-run exploit or
a claim that nonexistent Company20 already works.

RP-413's independent research approval is recorded in the owning log. S8's
original assignment exceeds the seat budget. A naive TS multiply/clamp has
correct integer output throughout the admitted domain despite an unsafe
intermediate. The actual-source intermediate-product witness has separate
independent corrective approval `ba1d7eaa..7b06dd5d`.

## Proposed contract

### I1 — Fresh identities, not reinterpretation

Use three distinct new identities: `H_catalog`, `H_company`, `H_replay`.
At the evidence coordinate, their next-free values are6,20,13 respectively.
These are proposals, not reservations or changes to current constants.
Before the first implementing commit, inspect both runtimes, schemas, retained
corpora and live writer/loader maxima. Record exact selected values in the
parent and plan together; if a value has become occupied, choose the next unused
value and obtain technical review of that mapping before landing.

`H_catalog` extends the complete current v5 grammar with required
`headcount_pools`; it does not replace axis-stack semantics. Catalogs1–5 retain
their exact admission/behavior and reject the new field. Empty pools are valid.
All S1 effect, binding, scope, integer and exact-key rules remain required.
Company and Founder versions remain scope-specific: this allocates no Founder
version and does not reinterpret existing Founder20/Soul.

### I2 — Independent feature activation and exact Company shape

Define headcount active as a pinned `H_catalog` artifact with a NONEMPTY
`headcount_pools` array. Define axis active from the existing pinned axis-stack
declaration, unchanged. Neither feature may be inferred solely from a larger
Company version or a gate crossing.

| Headcount | Axis | New-run Company floor / state, subject to I3's nonregression rule |
|---|---|---|
| absent/empty | absent | Existing bundle-derived floor; no Headcount fields |
| absent/empty | present | Existing Company19; unchanged attainment/axis behavior |
| nonempty | absent | `H_company`; complete Headcount maps; inactive attainment representation below |
| nonempty | present | `H_company`; complete Headcount maps and existing active attainment |

The parent's v16–v18 prerequisite artifacts remain required for an active
Headcount bundle. This proposal neither removes those nor adds a required Clout
catalog/upgrade, Founder progression or income source.

`H_company` has the existing v18 Company fields, the two attainment fields and
the parent's complete `headcount_seats`/`headcount_assigned` maps. Its exact shape
uses `achievements_attained_run: null` and `attainment_score_run: 0` when axis is
absent; when axis is present the attainment array is non-null, sorted and unique,
and all existing score/earned/attained derivations apply. No missing attainment
keys or non-null inactive map is allowed. This nullable arm is NEW-version-only;
Company19's non-null shape and all earlier codecs remain unchanged.

Bundle validation binds both feature states to these representations. When axis
is absent, no attainment hook, reattained event, axis contribution or axis snapshot
arm executes. When present, its existing mechanics execute identically. The
Company codec, genesis, Exit, clone/restore, validation, hooks and receipt
projectors must all consume this distinction in BOTH runtimes; changing only
`versionFloors` is not sufficient. Generic numeric `>=19` checks are not authority
to activate axis behavior for the new Headcount-only state.

All Headcount maps obey the parent's complete-key/integer/cap/assignment rules.
They are absent, not empty inserted fields, before `H_company`. Founder/Guild/
World refuse them. Active Headcount without its new Company identity and vice
versa refuse at the bundle boundary. No derived effective-heads values persist.

### I3 — New runs, historical compatibility and replay

An existing run retains its pinned catalogs, save version and command history.
Loading never promotes it to `H_company`. New-Founder initialization or an actual
Exit into a Headcount-active bundle creates zero seats/assignments under the new
identity; terminal old-run state retains its original version. Exit chooses the
next axis state independently by I2 where the version transition is compatible.
Headcount resets at Exit as already ruled.

Preserve the existing nondecreasing per-scope Exit contract: the next Founder
and Company versions cannot regress below their respective current versions,
and terminal Company state retains its exact source version. Upward/equal
compatible transitions are allowed. A lower bundle-derived next floor is NOT
permission to downgrade: an `H_company` run cannot Exit into an empty/absent-
Headcount target whose derived Company identity is19 or lower. Refuse atomically
with no persisted state, payout, receipt, event, outbox or history changes;
preserve the existing invalid-transition error boundary, not a new public error
or an automatic fallback. Do not inherit/manufacture `H_company` or add empty
maps to a target that does not activate Headcount to bypass the rule.

Axis toggles between two genuinely Headcount-active bundles remain possible
under the same compatible `H_company` identity and I2's explicit nullable/active
representations. This does not authorize downward removal of Headcount. Operators
cannot treat an empty-pool bundle as automatic rollback/deactivation after higher
versions are active; such cutover needs its own compatible, accepted route.
This proposal neither implements No Hires nor waives Deployment's no-live-down-
migration and exact-predecessor/rehearsal obligations.

The parent's codec/migration-corpus obligation becomes a tested predecessor-to-
`H_company` construction boundary, not an in-place live-run upgrade. Keep exact
old source state, terminal result, new genesis and subsequent commands in both
runtimes. No applied SQL migration body or historical fixture is rewritten; append
the next unused SQL event-kind migration and new corpus rows/count ratchet.

`H_replay` extends the COMPLETE current v12 resolved-command union and metadata
with the two canonical Headcount commands. Existing versions1–12 keep their
original grammar and reject those arms. New Headcount state/commands require
`H_replay`, as does an Exit whose next bundle activates Headcount; do not silently
reinterpret an old replay envelope. New writer output for other commands retains
their complete existing resolved fields and behavior, not the historical v8
subset described by the parent's stale v9 proposal. Replay old and new rows under
their actual source/pins. Kernel-prefix registration and a genuine same-commit
kernel bump accompany authoritative wiring, never this proposal alone.

From compatible preactivation baselines, an empty-pool new catalog retains the
existing feature-derived Company floor, state/receipt/event/projection behavior
and formula outputs. Test old v4 and current v5/axis companions, not only a v4
baseline; these comparisons do not promise a successful downward postactivation
Exit. Recompute real catalog/bundle
digests normally and retain each observation's actual provenance; never rewrite
replay pins or normalize governed payloads to force equality. Artifact identity
is not expected to equal different artifact bytes.

### I4 — Honest overflow acceptance, unchanged guarded algorithm

Replace only S8/AC3's impossible overflow-output obligation:

1. Keep the original `3.1e15 + 9e15` example as an invalid-assignment refusal,
   with no mutation. It is not an admitted overflow witness.
2. Run the complete effective-heads corpus and independent exact-integer models
   in Go, Node, Chromium, Firefox and WebKit, including the admitted
   `C=3.1e15,E=5.9e15,k=3` and maximum-domain boundaries.
3. Demonstrate actual Go multiply-before-clamp overflow on the admitted
   `E=0,C=k=2^53-1` case and actual TS ignoring-`k` failure on the ratio vectors.
   They discriminate different faults; do not substitute one for the other.
4. Preserve S2's prohibition on unsafe intermediate multiplication. Execute the
   actual-source operator witness and its source-derived unsafe multiply/clamp
   control: ordinary clamped output remains correct, but observing the unsafe
   multiplication must fail. Do not call that a wrong-output result or a universal
   proof; retain exact-output tests and explicit instrumentation limits.

S2's algorithm, admission, safe-integer budget and guarded arithmetic do not
change. Full AC3/AC5 still require their entire named runtime population and
complete future corpus; no current Node-only witness closes either criterion.
All other acceptance criteria, including full closed-form/replay/SQL/UI evidence,
remain required and are updated only for the fresh identities/activation above.

## Acceptance of this correction and implementation

Before implementation: applicable author/owner adoption, independent technical
review and SAME-edit parent-body reconciliation (Headcount and Tier2). The
current delegated model ruling alone does not adopt this technical proposal.

Implementation must prove the full I2 matrix, strict missing/extra/null/state-pin
refusals, upward/equal compatible Exits, downward atomic refusals and unchanged
old-run reload. A forced-non-null
inactive attainment map, forced Clout hook/projection, global `>=19` activation,
old-envelope Headcount arm, load-time upgrade, permitted downward Exit and
empty-pool overlay must each be
caught by a relevant actual-boundary regression. Use source-derived mutations
only where ordinary refusals/regressions do not establish discrimination.

Run the affected existing Go/TS/vector/real-Postgres populations and parent
acceptance, including valid no-axis Headcount, axis-only and combined commands.
Preserve original accounting failures, full P1/content/mint gates and nine-tier
release obligations. Independent technical review neither adopts the correction
nor approves its future implementation or release.
