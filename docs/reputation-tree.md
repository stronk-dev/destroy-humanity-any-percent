# Reputation tree

Status: **implementing, fixture-first.** No epoch pins a `reputation_tree` artifact yet, so no live
Founder can spend Reputation. This page describes what exists in code. The normative contract is
`rfc/reputation-tree-v1.md`, and progress is tracked in `planning/reputation-tree-v1/`.

## Artifact and loaders

`server/reputation` (Go) and `client/src/reputation.ts` (TS) strictly load a `reputation_tree`
artifact and enforce R2 rules 1–8. Every rejection names its rule.

1. Schema version 1. The `bonus` block must match one economy `multiplier_sources` row with slot
   `prestige`, target `all` and provider `reputation_tree`.
2. The tree has between 1 and 64 nodes.
3. Node ids are mechanical, start with `reputation.`, and are unique.
4. Each cost is an integer in `1..MaxExactInteger`.
5. `requires` is byte-sorted and names only earlier rows. Array order is therefore both the
   topological order and the display order.
6. The `bonus_unlock` ladder strictly increases, each rung requires the previous one, and the
   final rung is `1_000_000`.
7. `starter` rows use the curriculum starter union. Each nested arm's key set is
   checked before decoding: keys belonging to another arm reject even when zero,
   empty or null. Semantic validation still uses `curriculum.ValidateStarter`;
   this Reputation-only check does not change the shared curriculum loader.
   They also pass an aggregate-headroom check: resource initial + the largest curriculum grant +
   every tree grant stays within the hardcap, and generated counts stay within the provisioned
   hardcap. A `preowned_upgrade` may appear only once.
8. Every title and body key is a declared copy key.

Registering those copy keys as copy-bearing paths in `copy/references.v1.json` lands with the
production artifact at mint.

The fixture tree is `balance/testdata/reputation-tree/fixture-v1.json`, which holds the RFC's
proposed R2 table. It loads against the live economy plus the one `reputation.founder_bonus`
declaration row. Its node copy lives in `copy/catalog/reputation-candidate.json` as explicit
`PENDING OWNER COPY` placeholders (OD-11).

## Accounting and bonus

- `available = level − spent`, with `0 ≤ spent ≤ level`.
- The unlock ppm is the largest `unlock_ppm` among owned `bonus_unlock` nodes known to the tree,
  or 0. Owned ids must be byte-sorted and unique, and unknown ids contribute nothing (OD-7).
- The Founder bonus factor is
  `1 + level × per_level_ppm × unlock_ppm / 1e12`, quantized once. It is computed from the earned
  **level**, never from the available balance.

Both runtimes consume the shared loader and bonus corpora:
- `testdata/reputation/tree-fixtures-v1.json`: one or more rejection fixtures per rule.
- `testdata/reputation/bonus-vectors-v1.json`: Go-authored vectors. Setting
  `REPUTATION_UPDATE_VECTORS=1` regenerates them from Go.

`testdata/reputation/starter-key-rejections-v1.json` additionally binds both
loaders to twenty cross-arm zero/empty/null refusals with a legal three-kind
control. RP-243 records the original Go admission, correction and exact review
scope; these loader tests do not prove a minted player career or full RFC acceptance.

`testdata/reputation/bonus-domain-v1.json` supplies a sampled R1/R3 matrix:
147 level/ppm/unlock triples, 462 distinct legal spend cases and eight invalid
bonus-domain tuples. Both pure implementations check that spending reduces
available Reputation but leaves the earned-level bonus unchanged; the existing
ten shared vectors retain canonical Go/TS byte comparisons. The matrix is not
an exhaustive integer-domain or all-case cross-runtime rounding proof. Independent
earned-to-available and overspend-guard removals fail in each runtime, including
existing TS replay consumers. These fixtures do not close codec, transaction,
minted-career, browser or whole-RFC acceptance; see the implementation log.

## Snapshot readers

Under R1/R3/R9, active server projections derive unlock ppm from
the pinned tree and require it to match the Founder mirror. Null, duplicate,
unsorted and nonmechanical ownership refuse rather than being normalized.
Unknown historical owned IDs still contribute nothing; this is not permission
to remove nodes between epochs. The client checks both factor fields as canonical
Decimals >=1, preserving a nullable current-run factor without recalculating
eligibility or the bonus formula. Regression tests exercise the isolated Go arm
and public transaction-local initial snapshot, plus both public TS parsers.
These are presentation admission checks, not fresh database, HTTP, browser or
default-player acceptance evidence (RP-257/RP-258).

## Bundle wiring

`reputation_tree` is an optional epoch artifact, loaded by `server/replaycatalog` and
`client/src/replay.ts` into `CatalogBundle.ReputationTree` / `reputationTree`.

- **Chain.** It requires `minigame_api`, the artifact that owns Founder v21.
  A tree raises the bundle's Founder floor to at least v22; later Pet Adoption,
  Cosmetic Shop and Server Garden artifacts raise it to v23, v24 and v25.
- **Pairing (R2).** The artifact is present exactly when the economy declares a multiplier source
  with provider `reputation_tree`. If either is present without the other, the bundle is rejected.
  Go checks this again in `CatalogBundle.valid`.
- **Frozen contributions.** `production.ResolveFrozenContributions` now accepts that provider as
  well as `fiscal`, ready for the run-frozen `reputation.founder_bonus` row (B5).

## Founder save v22

`save.LatestFounderVersion` is 25, including later Founder features. Reputation
activates at Founder v22, which adds the required fields `reputation_spent` and
`reputation_nodes_owned` (byte-sorted and unique). A pinned bundle with
`reputation_tree` requires at least v22; later active feature artifacts can
raise its `versionFloors` Founder floor further.

**Codec (`server/save`):**
- v22 rejects `spent > level` and any unsorted or duplicate owned set, on encode and on load.
- Any state before v22 must carry no tree state and must have `reputation_unlock_ppm == 0`.
  A non-zero mirror is corruption, rejected rather than repaired (R7).

**Pinned-tree validation (`CatalogBundle.ValidateFoundationState`):** the owned set must derive
exactly the persisted `reputation_unlock_ppm`, and the available balance must be derivable.
Without a tree, all tree state must be empty.

The live purchase resolver checks this pinned invariant before resolving cost.
For activated v22+ state, the shared Go `ApplyFounderLogged` boundary also
checks it before any command, including a recorded-invalid arm, and before
returning successful output. Corrupt mirrors are errors, never repaired by a
purchase. Output validation uses the resulting epoch's bundle when Exit changes
the pin; failure restores the complete pre-command state and emits no receipt
or event. Legacy inactive purchases retain their ordinary R5 rejection.
Bare codecs remain artifact-free structural checks; this does not claim that
every repository reader/writer or real database commit has been audited.

At each new-run boundary, the mirror is re-derived from owned IDs against the
next epoch's tree before next-pinned admission and frozen bonus assembly. This
applies in the Go live/Founder paths and TS Company/Founder replay. A legal
same-ID unlock-effect retune therefore updates the next mirror without changing
earned Reputation, historical spend or owned IDs. It does not repair invalid
input or mutate the just-ended run's frozen contribution. Higher/lower fixture
retunes and unchanged controls exercise these paths; each rebinding removal
fails its affected retunes. Separately, OD-7 transition admission compares an
actual previous/next tree pair: every previously defined ID must remain, and an
active tree cannot disappear. Go linked-bundle admission and the direct live
boundary enforce this; TS linking and both direct Exit replay entries enforce
the same rule. A standalone historical artifact remains loadable, and initial
activation, legal effect/price retunes and appended IDs remain valid. Removing
defined IDs needs a successor RFC with an explicit refund, not unknown-ID
fallback. Nine independently removed IDs in three arms per runtime and a
whole-tree linked withdrawal each refuse; five guard omissions and an incomplete
shared census fail. This is fixture evidence, not global epoch-history or real
database admission proof (RP-251).

**Activation.** At a new-run boundary, and on the Founder-log Exit replay arm, a Founder moving
onto a tree bundle starts with `spent = 0` and `owned = []`. Earned `reputation_level` carries
over in full.

The TypeScript `encodeFounderReplayState` rejects invalid earned/spent/unlock
domains, unsorted/duplicate/nonmechanical owned ids and pre-v22 tree-state
leakage before serializing; it preserves valid ids and order rather than repairing
them. `restoreFounderReplayState` also checks the unlock mirror against the
pinned tree. Like the Go structural codec, the bare TS encoder has no tree
artifact and cannot establish that derived mirror by itself. Catalog-bound
restoration/validation is mandatory for that half of admission. Before v22 the
mirror is always 0. These rejection checks do not change valid replay bytes.

**Shared migration corpus.** `testdata/save-migrations.json` corpus v9 now
contains R7's four named Founder cases alongside eleven unchanged legacy cases;
the baseline requires all fifteen names. Its modern arm references the existing
Reputation replay fixture by source SHA, with shared patches/expectations.
Go and TS execute the actual Company Exit and Founder replay for v21→v22,
preserving earned4 with zero spent/owned/unlock and original canonical Founder
state, Company genesis and receipt bytes. Loading alone stays v21. The three corrupt inputs reject at structural
or pinned-mirror admission. Each runtime guard/default removal fails its case;
missing-case, source-SHA and closed-row-shape controls also fail. This fulfils
the previously absent corpus/ratchet locally (RP-249), not all earlier-version
activation chains, real database/default-player or whole B3/RFC acceptance.

**Earlier Founder activation (shared Go/TS).** `TestReputationEarlierFounderActivationChain`
executes seven legal writable sources (v14, v16–v21). Loading under each matching
bundle retains its version; the live new-run kernel and public Founder Exit
replay both produce v22 with earned11, spent0, owned[] and unlock0. Both outputs
encode/restore under the next bundle, pass pinned validation and agree in full
encoded Founder bytes after the expected Exit history append. Existing age and
route knowledge survive. Four independent Go runtime corruptions make this test
fail. Older fixture targets keep minigame/Fiscal keys stable and use the existing
unlock chain without starters; v21 uses epoch8 and the full tree fixture. This
is not a newly minted epoch, historical v15 writable path, Company-log/DB commit
proof or complete AC11.

`testdata/replay/reputation-earlier-activation-v1.json` stores the exact seven
Go-authored source/command/result rows with nine deduplicated artifact bundles.
Normal Go runs require byte-exact regeneration equality; authoring is explicit
via `make reputation-activation-corpus` after all transition assertions pass.
TS loads each source/target through the actual strict artifact loader and its
public Founder replay must match Go's complete state, receipt, events and result
hash. The pre-state also encode-roundtrips without activation. TS spent/default,
minigame and Soul initialization removals fail the affected rows. Removing a
case, forging a receipt or deleting an expected event fails both lanes. Each
row contains one event: this is not multi-event permutation/transaction proof.
Existing migration and Reputation replay corpora are unchanged.

The Company-side Founder carry includes tree fields for pinned Founder floors
of v22 or higher. Both runtimes reconstruct them; the former pre-R6 carry gap
is no longer the current behavior. See the replay-input contract below.

## `purchase_reputation_node` (R5)

This is a Founder-scope intent on `POST /api/v1/intents`:
`{intent_id, kind: "purchase_reputation_node", expected_revision, node_id}`.
`Service.handleFounderReputation` routes it through `ApplyFounderLogged`, and the Go and TS replay
arms share one contract. Like the Fiscal intents, it is exempt from the Soul-recovery exclusivity
gate.

Validation runs in this order, and the first failure wins. Every rejection is a recorded
`founder_log` row that emits no event and changes no state.

| Step | Condition | Rejection |
|---|---|---|
| 0 | Wrong fields, or a non-mechanical `node_id` | `invalid / purchase_reputation_node.fields` or `invalid / node_id` |
| 3 | Inactive: the bundle has no tree, or the Founder is below v22 | `not_eligible / reputation_tree_inactive` |
| 4 | Unknown node | `unknown_id / <node_id>` |
| 5 | Node already owned | `not_eligible / owned` |
| 6 | Prerequisite missing | `not_eligible / requires` |
| 7 | Cost exceeds available | `unaffordable / reputation` |

The resolved inputs are `{kind, node_id, resolved_cost, reputation_level, reputation_spent_before,
owned_before}`. `resolved_cost` is 0 unless the purchase applies. Replay recomputes every field and
refuses tampered inputs.

The Go purchase arm checks the original resolved JSON for exactly those six
required, non-null, case-exact fields and rejects duplicate keys before struct
decoding. Invalid frozen inputs return `ErrInvalidReplayInputs` with no receipt
or events and restore the complete pre-command state, including any lazy Fiscal
sweep. This does not alter the envelope decoder, other intent arms or valid bytes.
TS checks the exact parsed object and typed values; its public replay API does
not receive raw JSON and cannot recover duplicate keys erased during parsing.

`testdata/reputation/purchase-input-shape-v1.json` binds nineteen valid controls
to the existing purchase corpus SHA: eighteen purchase-arm cases plus a zero-
earned unaffordable case. It records six fields × five single-field mutations
per control. Go refuses all 570 raw mutations; TS refuses 456 malformed objects
and reproduces 114 honestly normalized duplicate controls. All nineteen valid
results byte-match state, receipt, ordered events and result pin. Removing each
runtime's key gate fails its affected population; source-hash, missing-row and
altered-row controls fail both consumers. Authoring is explicit through
`make reputation-input-shape-corpus`; ordinary tests never regenerate it.
This RP-255 supplement needs designated review and actual database coverage
remains separate. Native Chromium/WebKit execute all 590 new TS tests, but
Firefox launch is blocked on this Mac (RP-256); no three-engine/CI pass is claimed.

RP-253's additional `TestReputationPurchaseTaxonomyIntegration` is retained but
has not yet executed on Postgres. It covers all twenty pinned purchase cases
plus an overdue-Fiscal rejection, exact retry, and forty-two unrecorded
revision/idempotency conflicts. Its local profile preparation and population
controls pass; they are not persistence proof. Actual SQL execution and fired
persistence controls remain required before AC3 closeout.

An applied purchase adds the cost to `reputation_spent`, inserts the id in sorted order, and updates
the unlock mirror. It returns a receipt with `effective_from: "next_run"` (a Fiscal sweep, if any,
decorates it) and emits `reputation_node_purchased.v1` with `source: "direct"`. The event is
validated by `save.validateEventPayload` through `validateIntentDecision`, before ordinary or
Exit event writes, and admitted to the database by migration 00075.

The eight event fields are required, non-null and exact-case; duplicate/extra
keys and trailing JSON reject. Cost is in `1..MaxExactInteger`; earned level and
both spend totals are in `0..MaxExactInteger`; unlock is in `0..1_000_000`.
These bounds are checked before addition, so a signed overflow cannot satisfy
`spent_after == spent_before + cost <= level`. Kind and source stay closed enums.
An isolated R5 supplement exercises 55 malformed decisions, six valid boundaries
and all eleven original direct-purchase producer events. Independently omitting
shape, domain or relationship admission fails 20, seven or two assertions.
It does not change producer bytes or the applied migration, and is not a fresh
Postgres run or proof that every direct rejection has a persisted log row
(RP-252/RP-253).

`testdata/replay/reputation-tree-v1.json` is the Go-authored cross-runtime corpus: the inactive,
invalid, unknown, requires, owned and unaffordable rows (including the `cost == available + 1`
boundary), plus a nine-node chain with an automatic Fiscal sweep. Set
`REPUTATION_UPDATE_FIXTURE=1` to regenerate it from Go.

The Go generator and TS consumer now explicitly compare the result constants
hash for all twenty direct purchases (eleven applied, nine rejected) against
the shared bundle pin. They also compare the three paired Founder Exit results
against their recorded next pin; two rejected Company Exit cases have no Founder
result. Population assertions prevent silently dropping these observations.
Wrong direct-result hashes fail both consumers while the older state/receipt/
event comparisons alone still pass. Wrong old-pin Exit results and missing-row
controls fail too. The Go Exit fault is injected after the runtime call: it proves
the assertion, not a production defect. Corpus and runtime bytes are unchanged.
This RP-254 supplement needs Claude's designated review and does not close the
remaining R8 history/verifier work or RP-253's actual Postgres population.

## Frozen Founder bonus (R3)

`production.FrozenFounderContributions(bundle, founder)` produces the complete run-frozen Founder
contribution set. It holds the Fiscal rows, plus exactly one `reputation.founder_bonus` row
(prestige slot, target `all`) whenever the pinned bundle carries `reputation_tree`. That includes
a `1e0` factor for a Founder with no unlock node.

The factor is `1 + level × per_level_ppm × unlock_ppm / 1e12`, computed from the Founder state
after the Exit's Reputation credit. The earned level is used, not the available balance, and a
mismatched unlock mirror is refused. Exit (`prestige.go`) and New-Founder initialization
(`FounderInitializer`, which account import also uses) both call it.

Migration 00076 replaces `require_fiscal_frozen_contributions()`. A run pin now expects the Fiscal
rows plus one extra row when the pin's bundle has a `reputation_tree` artifact. A missing or extra
row fails the commit.

Production reads a run's bonus only from the stored rows, so a mid-run purchase leaves the current
run's contributions byte-identical and changes only the next run's frozen factor.

## Starters, run_started v2, and the v9 carry (R4, R6 carry half, R7)

**Replay inputs.** Reputation introduced the v9 carry; the current
`save.ReplayInputsVersion` is 12 for later features. Legacy v9 remains readable,
with an explicit Go replay and TS AC8 consumer at both v9 and v12. The Founder
carry extension adds `reputation_spent`, `reputation_unlock_ppm` and
`reputation_nodes_owned`. These fields are present exactly when the pinned bundle's Founder floor is
22 or higher. They are rejected before v9 and absent below floor 22. A v22 Founder therefore now
reconstructs from a carry in both runtimes, replacing the earlier fail-closed behavior.

**New-run assembly.** After the curriculum starter, every owned `starter` node known to the next
bundle's tree is applied additively, in tree array order:
- `resource_grant` credits the ledger;
- `generated_generators` adds to provisioned units, and exceeding the provisioned hardcap is an
  engine error, never a clamp;
- `preowned_upgrade` sets ownership.

**run_started v2.** When the new run's bundle has a tree, `run_started` is emitted at schema 2 with
`reputation_tree: {bonus_factor, applied_starter_node_ids}`; otherwise v1 stays byte-identical.
`save.validateEvent` checks the arm, and migration 00077 admits schema 2.

**Executed fixture evidence.** The original AC8 corpus proves ten curriculum
units plus five Reputation units, with zero purchased units and canonical
receipt/state/event comparisons. Assignment-instead-of-addition fails in Go and
TS. `testdata/reputation/starter-effects-v1.json` adds three shared semantic
expectations on the real Exit replay: all known nodes plus a retired unknown id,
retired id only, and no owned nodes. The full case exercises all four starter
nodes, cumulative cash grants, the preowned upgrade and tree-ordered applied ids;
unknown ids remain owned but contribute nothing. Independently dropping upgrade
ownership or sorting emitted ids bytewise fails each runtime's new test.
These three semantic cases are not new all-case canonical byte corpora.

`testdata/reputation/starter-boundaries-v1.json` (table version 2) supplies three
legal fixture-only next-bundle cases: granting the same preowned upgrade from
curriculum and the tree; landing exactly on the provisioned cap; and landing
exactly on the permit cap. Their expectation rows are unchanged. The former
retirement positive is now a separately consumed forbidden transition removing
the generated/upgrade IDs: both runtime consumers assert refusal without state
mutation. Changed catalogs load independently under hashes derived from their
actual bytes, but the previous/next pair must obey OD-7. The earlier unknown-ID
fallback cases remain independent and do not authorize defined-ID removal.
Both loaders refuse
over-cap raw variants. Go's served helper rejects bad headroom with
`ErrInvalidEngineState`; TS checks the defensive guards through explicitly
fault-injected copies of parsed bundles, which are NOT admitted artifacts or
persistence evidence. Independent cap-guard, saturation, wrong-current-tree
and upgrade-toggle mutations fail in each runtime. No minted tree, real DB/
default player journey or full Reputation acceptance follows; RP-245/RP-247
and the implementation log retain the precise boundaries.

## Exit-attached purchase plan (R6)

`accept_exit_offer` and `wind_down` accept an optional `reputation_plan` key: 0–64 unique
mechanical node ids, listed in purchase order. The key belongs to the canonical request and its
hash. Requests without it are byte-identical to before. `/api/v1/intents` is not an operation in
the generated API registry, so API Foundation C2's rule that v1 request unions must not widen does
not apply.

Inside the Exit, after the Exit's own Reputation credit, the plan is dry-run first:
- If the next bundle has no tree, the whole Exit is rejected with
  `not_eligible / reputation_plan.tree_inactive`.
- Otherwise the purchase rules run in order, each against the state left by the previous entry, so
  a node may require one bought earlier in the same plan. The first failure rejects the whole Exit
  before any mutation, as `<category> / reputation_plan.<detail>`.

The purchases are then applied after v22 activation and before starter assembly and the frozen
bonus row. Each emits `reputation_node_purchased.v1` with `source: "exit_plan"`, ordered after
`founder_advanced`.

The Founder log records an Exit with an applied non-empty plan under the append-only `exit.v2` arm:
`exit.v1` plus `reputation_purchases: [{node_id, resolved_cost}]`. Founder replay re-derives those
purchases and refuses any mismatch. Migration 00078 admits `exit.v2` as a Company-linked Founder-log
arm.

When the live Exit is checked against its replay (parity), the expected Founder state now copies
the v22 tree fields. Without that, a live Exit with a plan, or one activating v22, diverged.

## Portable Founder-history evidence

`server/production/reputation_history_test.go` drives the public
`VerifyFounderHistory` consumer with the unchanged purchase corpus: twenty
individual commands, three paired plan/activation Exits and the complete
nine-purchase chain. Single-command wrappers start at their original revision
and rebase only the local log sequence; the chain preserves its original
coordinates and requires adjacent full states and pins to match. Expected
heads, receipts and events come from the source, not from the replay under test.

The population includes 144 corrupt-history refusals: head mirror, head pin,
receipt outcome, extra event, sequence and linked-source presence. Independent
comparison omissions discriminate; either sequence defense alone still rejects
bad ordering. This is portable replay evidence, not execution of SQL
`LoadFounderHistory`, a two-Exit purchase career, the Company-run verifier or
the real-Postgres composed acceptance gate (RP-259).

## Portable Company-run evidence

`testdata/reputation/company-run-v1.json` retains two complete runs consumed by
`VerifyReplayRun`: the original corpus's exact run-2 genesis (cash1e3, fifteen
generated towers, zero purchased) with its frozen bonus1.003, and a synthetic
unit-factor control with identical starter state. Both accrue, cross the Garage
gate through the ordinary command, then Wind Down. No direct cash/tier/gate
setup replaces those commands. The source Exit's receipt, ordered events and
full Company outputs are checked before fixture generation.

`server/production/reputation_run_verifier_test.go` normally reads retained
commands, inputs, receipts, events and terminal states; it does not generate
its own expected verifier answers. A separate byte-regeneration check detects
drift and writes only under an explicit test-generation flag. Fixture presentation
whitespace is normalized only for canonical command bytes. Independent first-
action arithmetic distinguishes the frozen bonus's generator production from
the manual action's unmodified base grant.

Eighteen corrupted runs refuse, including removed starter assets, changed
first/terminal factors, broken logs/pins and forged receipts/events. Four
population controls and two false expected-head controls reject. Omitting the
bonus consumer breaks the non-unit run while preserving the unit control;
receipt/event/terminal/head/population omissions also discriminate. This is
portable pinned replay proof (RP-260), not SQL provenance, a stored two-Exit
purchase career, the default browser workflow or AC15's real-Postgres gate.

Run from the repository root:

```sh
make test-go GO_PACKAGES=./production GO_TEST_FLAGS='-run TestReputationCompanyRun -count=1'
```

Explicit regeneration uses the same root target with
`GO_PACKAGES='./production -args -update-reputation-company-run-fixture'` and
`GO_TEST_FLAGS='-run TestReputationCompanyRunFixture -count=1'`. Its changed
expected bytes require review; normal verification never enables that flag.

## Career measurement inputs

The headless fixture composer copies the complete base artifact set, retains
the exact `fixture-v1.json` tree bytes and pairs them with the economy's
`reputation.founder_bonus` declaration. It computes the complete catalog hash
and loads the result through `replaycatalog.Load`; it no longer changes parsed
catalogs while retaining the tree-less epoch's artifacts and hash. The career's
run key uses the supplied fixture bundle's hash. Threshold, purchase policy and
exclusion remain separate experiment inputs, not part of that catalog hash.

`TestReputationCareerFixtureIdentity` checks retained tree/economy sources,
public-loader roundtrip and a completed Chaos seed0 no-purchase career, while
requiring the original suite to remain untouched. Removed-tree, undeclared-source
and false-hash copies refuse at the public loader. Independently reverting the
fixture identity, omitting the run-key assignment or mutating the base artifact
map makes the test fail. This corrects RP-263's helper/run-key portion; retained
H4/H5 report envelopes still lack complete source/configuration provenance.

The composed Reference planner now carries the run's frozen contributions
through candidate transitions, action-free advances and masked rate projection.
Actual Reference waiting uses the same input. It previously omitted that bonus
in these consumers (RP-265); this correction does not change live production
arithmetic, tree literals, purchase policies or thresholds.

`TestReputationReferenceFrozenInputDiagnostic` compares no-row, synthetic unit
and actual tree-derived `1.003e0` inputs against canonical production. Its legal
isolated run3 starts with two served starters, cash1e3 and five generated/zero
purchased Beige Towers. Projection must be5/5.015; full encoded advance,
purchase and bank states must match their canonical counterparts. Effect masks
and zero production remain controls. The bank branch explicitly uses a zero-cash
counterfactual; this is not a naturally earned or persisted career.
The instrument uses legacy v14 Company states without modern active-play
simulation. Its rate oracle uses the same economy-only projection contract.

The producer tests retain nine valid rate profiles and eleven invalid-input
refusals, including malformed or duplicate contributions and invalid masks.
Direct/alias/quoted-decoy controls cover both rate entrypoints (RP-266).
Independent compiling input and mask omissions make the named tests fail;
the candidate's projection is checked independently of candidate-state equality.

An explicit Reference seed0 career observation at threshold1e5 now reaches
the run3 gate at357000ms in both treated and control arms. Before correction,
the treated arm was355000ms and control357000ms. Its actual tree-derived factor
is1.001 with no starter purchase. This is instrument drift, not a balance ruling
or the missing H3 `1.000001e0` milestone proof. Run the observation from root:

```sh
make test-go GO_PACKAGES='./harness -args -reputation-reference-observe=current' GO_TEST_FLAGS='-run ^TestReputationReferenceCareerObservation -count=1 -v'
```

The observation logs complete source coordinates and both outcomes without
writing reports. Its default is an explicit skip. `baseline` mode asserts the
retained row; it does not switch to old code and is expected to reject that
stale row after the correction.

Reference's candidate adoption and actual waiting now also bind the same
`firstHourLifetimeHook` as ordinary first-hour intents and advances (RP-267).
That observer delegates to `prestigecore.AccumulateLifetimeValue` with the
canonical accrual receipt. The candidate/advance simulation carries it as an
internal optional dependency; standalone relevance suites remain observer-free.
No payout formula, offline-session accounting or live-game behavior changes.

`TestReputationReferenceLifetimeAccounting` has18observations: for each no-row,
unit and tree input, actual candidate adoption, actual bank dispatch and direct
ranker advance must credit5/5.015 and match complete canonical states. Ordinary
intents are healthy controls. Zero elapsed time and effect-masked production
credit0; starter grants are not fabricated as paid production. Hook-binding,
shared-hook-no-op and advance-mask omissions make the tests fail. Known-value
assertions catch a broken shared observer even when consumer and oracle agree.
The existing16frozen-input tests retain complete state equality, with their
three canonical state-producing calls reconciled to the now-required observer.
Their earlier Routes-only scope did not prove lifetime accounting.

At the same actual Reference seed0 career/threshold1e5, corrected scripted
Exit lifetime is`1.4605083614e6` instead of`1.26604673417e6`, and elective lifetime
is`3.54431965065e6` instead of`3.34889972124e6`. Both treated/control careers
complete with two Exits. Paid deltas remain2/0, ownership/starters/factors stay
unchanged and both run3 gates remain357000ms. All seven preceding milestone
clocks are unchanged for this seed; this is not the full H3 population or its
tiny non-unit sensitivity proof. Full H4/H5 reproduction still rejects the
retained snapshots; the same six Casual ties remain.

The threshold reproduction test reads retained first-hour H1 samples rather
than rerunning the corrected producer. Its green result is historical-source
recalculation, not fresh calibration. Complete H1/H2/report provenance and
current-producer reproduction remain separate work; no report refresh or
threshold ratification follows from the local observer correction.

### H3 current-producer sensitivity study

The full H3 instrument runs all 97 ratified first-hour seeds in five arms:
fresh epoch-8, complete paired Reputation fixture with no contribution,
explicit prestige factor `1e0`, factor `1.000001e0`, and a separate strong
control of `2e0`. All use the original two-hour horizon, empty purchase plans,
live `1e12` threshold and unchanged first-hour policy/experiment. These are
synthetic input counterfactuals, not earned Founder bonuses or SQL careers.

An internal optional input is copied into the ordinary runtime. Nil preserves
existing callers; a diagnostic input combined with career or Tier2 mode is
refused before execution. Frozen contributions still pass canonical production
admission. No live multiplier authority or report-key/schema field is added.
The diagnostic logs its factor separately: the existing RunKey alone does not
identify this synthetic extra input.

The 2026-10-06 run completed all 485 runs and 3,395 milestone observations.
Every run completed with seven clocks, two Exit samples and zero paid
Reputation. Fresh epoch-8 and paired/no-row reproduce the retained milestone
clocks and ending samples. Explicit unit matches no-row clocks, endings,
Exit samples and transition counts. All per-policy distributions are compared.
The tiny factor moves **zero of 679 clocks**; the strong control moves **289**.
The tiny factor does reach the consumer: Reference lifetime production changes
even though its milestone clocks do not. A transition-count change is not
relabelled as a milestone. The exact RFC H3 tiny-factor criterion is therefore
**RED**, not waived or replaced by the strong control. Its author must reconcile
the fired criterion before H3/AC13 closeout.

Run the full observation from the repository root:

```sh
make test-go GO_PACKAGES='./harness -args -reputation-first-hour-observe=all' GO_TEST_FLAGS='-run ^TestReputationFirstHourSensitivityObservation -count=1 -v -timeout 10m'
```

Default execution explicitly skips this full study; an unknown selector fails.
It never updates retained reports. The ordinary fast lane instead executes a
small, explicitly single-Reference-row consumer control and an oracle population:
23 malformed populations reject, a valid changed clock is accepted, unchanged
sensitivity rejects, and career/Tier2 input overrides reject. Compiling input,
wrong-factor, census and sensitivity omissions make these controls fail. That
fast control is not the full study and does not make H3 green.

Historical H1/H2 recalculation, source/report provenance and full fresh
calibration remain separate. Existing H4 ties, H5 epsilon/run4 and report drift
remain; no report refresh, threshold/policy change or epoch mint follows.

The R10 headless career runner accepts exactly `cheapest`, `seeded_uniform`
and `none`. An optional leave-one-out exclusion must name a node in the loaded
fixture tree. Both are checked before simulation: a misspelled policy must not
silently become a no-purchase control when the seed earns no spendable Reputation,
and a misspelled exclusion must not silently reproduce the baseline arm.
Invalid inputs return `ErrReputationCareer`.

`TestReputationCareerRefusesInvalidInputs` covers empty/unknown policies and
unknown exclusions at both fixture and live thresholds. Four legal controls
exercise all three policies and a real-node exclusion. Independent guard
removals make the tests fail. These unconditional tests run in the fast harness.
The exhaustive report reproduction remains in `make reputation-harness-check`.
After the Reference correction, strict H4 and H5 reproduction both reject
report drift; their v1 bytes remain historical snapshots, not current-instrument
acceptance. The full H4 run still records the same six Casual ties. H5's
epsilon/run-4 questions, report provenance and H3's fired tiny-factor criterion
remain unresolved. No report refresh, retune or release claim follows.

The test-side H4 report gate validates each row before starter eligibility.
Only `run3_gate_beyond_ratified_horizon_in_both_arms` with both clocks absent
is an exclusion; observed clocks must be nonnegative. A finite pair must carry
exactly control minus treated as its saving. Any unreached clock forbids a
finite saving. No-starter rows remain outside timing comparisons, but cannot
publish malformed evidence. Derived gated/excluded counts reset on each call.
Strictly sooner remains the requirement: ties and missing treatment fail.

`TestReputationCareerGateExclusionDiagnostic` exercises 28 synthetic profiles,
including false exclusions, missing/wrong savings and repeated census calls.
These prove report admission, not naturally earned careers or SQL integration.
Nine independent compiling guard/caller/counter/strict-comparison probes fail;
the missing-saving guard and ignored-admission probes expose caught panics,
not semantic refusals. The current complete 97-pair producer independently
logs 93 counted comparisons and three valid both-unreached exclusions
(Casual seeds 4, 14 and 21). The same six genuine Casual ties remain failures.
The separate full 970-arm H5 reproduction still rejects retained report drift.
Old report bytes, horizon, policy, balance and kernel161 remain unchanged.
This is locally verified instrument repair, pending designated review, not
H4/H5 or AC13 acceptance. The RFC's absent named career-data artifact and fresh
H1/H2 source admission/lineage remain separate work.

The H2 payout reader now requires exactly two ordered first-hour Exit samples:
run1 `scripted_first`, then run2 `collapse`. It does not search through duplicate,
misordered or later-run samples. Canonical lifetime parsing and the existing
prestige arithmetic still determine the payout; valid zero lifetimes remain
zero. This is a first-hour measurement helper, not a general career-history API.

The pinned threshold study additionally admits its complete source against the
loaded ratified scenario, policy, constants, experiment and run coordinates,
including required observations/outcomes and aggregate source/count coordinates.
It uses the existing full-population oracle plus a threshold-only aggregate
coordinate check. The generic `MeasureReputationThresholds` calculator remains
usable with legal smaller populations; it does not itself certify a full study.
Aggregate value arrays are not re-derived by this admission wrapper.

Forty-five diagnostic children cover arithmetic controls, malformed Exit pairs,
24 malformed study sources, full/reordered historical sources and a legitimate
two-row generic subset. Six independent compiling omissions fail, including
dropping the previous Founder level, then restore exactly. The retained 30-point
measurement reproduces byte-identically. This is historical-source recalculation,
not fresh H1 production, current threshold calibration or owner ratification.
Full-source lineage and the separate H3 aggregate-coordinate gap remain open;
no old report regeneration or H2/AC13 completion follows from this local repair.
