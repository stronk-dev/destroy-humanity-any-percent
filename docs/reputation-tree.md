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

RP-253's additional `TestReputationPurchaseTaxonomyIntegration` now executes on
the declared disposable Postgres service. It covers all twenty pinned purchase cases
plus an overdue-Fiscal rejection, exact retry, and forty-two unrecorded
revision/idempotency conflicts. The first actual SQL execution exposed three
test-setup defects (RP-294–296): duplicate current epochs, an omitted service
current hash, and canonical replay payloads sent as wire requests without
`intent_id`. The fixture registers both source bundles under one epoch and
retains complete request envelopes separately; stored canonical-byte oracles
and production constraints are unchanged. All21 persisted profiles pass.
Omitting prerequisite refusal fails two profiles while19 controls pass;
omitting a source hash binding fails the fixture population check. This local
evidence needs designated review; it does not close all AC3 obligations.

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

`TestReputationExitPlanWriteFaultsIntegration` executes all14 logged-Exit write
fault stages against declared Postgres, using the live Exit callback and an
affordable prerequisite plan. Every stage must fire its exact injected error
and leave full persisted rows and both decoded stream heads unchanged. This
includes revision/event history, genesis/logs, run pins/frozen contributions,
receipts, player outbox and verification queue; sequence counters are not
claimed transactional state. Eight byte-identical diagnostic initial revisions
make retention delete real eligible rows in the final successful control.

Normal `Service.Handle` then commits all three purchases exactly once, freezes
the next bonus at1.003, prunes old revisions, passes Founder/Company replay and
returns an identical retry without new rows. A compiling source probe that
commits despite the injected retention error fails the full-row rollback
oracle in11 tables. Production source was restored exactly before final cold
tests. This is persistence-boundary evidence, not a new HTTP/socket/actor/guard
population, natural pacing, designated approval or full AC9/RFC acceptance.

```sh
make test-save-integration SAVE_TEST_PACKAGES='./production' SAVE_TEST_FLAGS='-v -run TestReputationExitPlanWriteFaultsIntegration'
```

`TestReputationPlanRequest*` separately checks the incoming request boundary
for both Wind Down and Accept Offer. It accepts absent/empty/ordered plans and
64 syntactically valid IDs; rejects65, duplicates, malformed IDs and non-string
arrays; and refuses plans on IPO, Decline Offer and a `scripted_first` request.
Literal canonical bytes and independent SHA-256 expectations preserve plan
order, presence, revisions and Offer ID. Omitted versus `[]` both parse to an
empty plan but have different request identity; whitespace/key order and intent
ID do not alter that identity. Other unknown fields and invalid Founder
revisions still reject.

Six compiling source probes separately break size, uniqueness, ID syntax,
canonical plan inclusion, parsed order and IPO exclusion; unchanged tests
reject each. Source is restored exactly and cold focused/SQL tests pass. This
is parser evidence:64 syntactic IDs are not64 registered/affordable nodes,
and Accept Offer parsing is not an offer-with-plan SQL or browser journey.

```sh
make test-go GO_PACKAGES='./production' GO_TEST_FLAGS='-count=1 -run TestReputationPlanRequest'
```

`TestReputationOfferPlanPayoutIntegration` separately drives a real gate-produced
Acquihire or Acquisition offer through `Service.Handle`. Its diagnostic run2
starts with zero Reputation. Independently expected payouts18/20 fund the
six-point prerequisite plan after Exit credit; an unaffordable last entry
instead preserves both full save heads, the pending offer and game/evidence
rows, while recording the rejection. Ordered purchase events, accepted offer
identity/resolution, committed receipt revisions, generated/cash starters,
next frozen1.009/1.01 and `run_started` v2 are checked against persistence.

The completed Company run (producer, rejection, applied Exit) and the complete
Founder history both verify. Copied plan-order/cost corruptions refuse without
SQL writes; exact rejected/applied retries and changed-plan same-ID conflict
also preserve complete table values. Compiling producer, payout-validation,
valid offer-ID and idempotency faults fail unchanged tests. An invalid
resolution literal is separately refused by the save validator before the
new event oracle. Source is restored exactly before broader cold tests.

This is locally executed diagnostic SQL evidence pending designated review,
not natural later-tier progression, browser/AT/minted content, full AC9/RFC/CI
or release acceptance. The existing shared Go/TS five-Exit corpus does not
carry this offered-plan population; the separate supplement below supplies
replay parity without rewriting that historical corpus.

```sh
make test-save-integration SAVE_TEST_PACKAGES='./production' SAVE_TEST_FLAGS='-v -run TestReputationOfferPlanPayoutIntegration'
```

### Offered-plan shared replay supplement (R8)

`testdata/replay/reputation-offer-plan-v1.json` is authored only by executed Go
transitions. Both runtimes compare ten Company and ten paired Founder arms:
Acquihire/Acquisition × payout-funded plan, last-entry unaffordable refusal,
absent plan, explicit empty plan, and preserved promised payout. Each stream's
population is eight applied/two rejected. Initial earned/spent Reputation is0;
fixed expected payouts18/20 fund the ordered six-point prerequisite plan.

Comparisons bind complete canonical receipts, final/new Company state,
Founder carry/state, ordered events and Founder result pin. Independent controls
check byte-sorted ownership versus purchase order, cash1e3/generated5/purchased0,
bonus1.009/1.01, accepted offer identity before run-end, and unchanged refused
state/no events. Absent/empty retain different canonical requests but equal
semantic outputs. Both runtimes refuse copied cost, promise and prerequisite-
order corruption, plus a rejected Founder arm falsely claiming credited delta.

Compiling source faults demonstrate the independent prospective-payout check,
exact TS purchase-event byte comparison and rejected-Founder delta refusal.
Generation refuses to replace expectations when independent controls fail.
The earlier twenty-purchase/five-Exit corpus remains byte-unchanged. These are
diagnostic stored-offer replay cases, not a second live producer, SQL retries,
natural progression, browser/AT, full R8/AC9/RFC/CI or release acceptance.
Designated review is pending.

```sh
make test-go GO_PACKAGES='./production' GO_TEST_FLAGS='-count=1 -run TestReputationOfferReplay'
make test-client
# Explicit fixture authoring only; ordinary tests never regenerate:
make reputation-offer-corpus
```

### Persisted activation and first-failure supplement (R6/R7)

`TestReputationExitBoundaryIntegration` executes24 profiles against declared
Postgres: WindDown, Acquihire and Acquisition each have three old-pin activation
controls and five whole-Exit refusal controls. Activation starts with the live
v21 Founder/tree absent and moves to a distinct diagnostic tree bundle only at
Exit. A direct purchase before Exit still refuses without changing either head,
even though the latest available bundle has the tree. Disposable epochs close
and append normally; no same-epoch tree-artifact addition or repository mint.

The nine activation cases check plan/absent/empty inputs, exact committed
receipt revisions, Founder v22, old run pin preservation and the next run pin,
spent6 or0, ordered purchase events, sorted ownership, starters, frozen1.003 or
unit factor, and the start summary. The completed old-pin Company run and whole
Founder history both verify; substituting the wrong hash refuses. Full replayed
next Company state matches its saved head.

The fifteen refusal cases independently name inactive/unknown/owned/missing-
prerequisite/unaffordable first failures after valid prefixes. Full heads,
pending offer and game/evidence rows stay unchanged; required rejection logs,
receipt and outbox are recorded. The refusal is replayed directly, not called
a completed verified run. Exact-number canonicalization compares every replayed
state value despite PostgreSQL jsonb key/spacing differences; copied tier changes
fail this comparison. Raw persisted before/after comparisons remain byte-exact.
All24 exact retries preserve twelve complete tables.

Three inactive-tree cases then take an actual plan-free WindDown on those same
heads, with no reseeding. Both two-entry histories verify and exact retries
preserve complete rows. No Reputation purchases or frozen bonus row appear.
These are three sequential fallback controls, not three extra matrix profiles.

Compiling source probes discriminate current-versus-next validation, exact
refusal detail, independent persisted spent accounting and the plan-free door.
Skipping plan application is caught earlier by the existing live/replay parity
guard; the separate shared-helper spent reset reaches and fails the new oracle.
Every production source is restored byte-identically before broader cold tests.

This is diagnostic initial earned6/run2/tier3/stored-offer SQL evidence, pending
designated review. It does not establish naturally earned progression, live
offer generation (proved separately above), Go/TS cross-pin corpus parity,
default browser/AT behavior, minted release data, full AC9/RFC/CI or1.0 readiness.

```sh
make test-save-integration SAVE_TEST_PACKAGES='./production' SAVE_TEST_FLAGS='-v -run TestReputationExitBoundaryIntegration'
```

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

## Composed real-Postgres career evidence

`TestReputationCareerTwoExitsIntegration` supplies the separate R8/AC15 SQL
population: one scripted burnout Exit, a direct unlock purchase with exact
retry, ordinary accrual and Garage gate, an elective Exit with the cash/tower
prerequisite plan, then run3 production. After the initial fixture genesis,
all gameplay transitions go through `Service.Handle`; later runs, Exits and
frozen rows are not independently seeded. The initial earned level6 is an
explicit diagnostic budget, not naturally earned pacing or OD-2 acceptance.

Both completed Company runs replay from their stored genesis/log/events, and
the complete three-entry Founder history verifies. Current-run frozen bonus
stays unit after the direct purchase; run3 freezes1.003 and consumes cash1e3
plus five generated towers with zero purchased. Its first command's replay
matches the full stored head and independently expected production. Run3 is
unfinished, not labelled a completed Company-verifier population. Company
replay excludes automatic Founder Fiscal prefixes, which the full Founder
history verifies separately; terminal plan events remain in its population.

Changing one existing copied frozen-factor byte returns `state_divergence`;
so does a copied Founder-head corruption. A live bonus calculation from
available instead of earned Reputation fails the run3 factor assertion.
No immutable database evidence is edited. The whole production Integration
population passes cold on declared Postgres after exact source restoration.
This is locally executed diagnostic evidence pending designated review, not
minted/default-browser gameplay, complete RFC/CI or release acceptance.

```sh
make test-save-integration SAVE_TEST_PACKAGES='./production' SAVE_TEST_FLAGS='-v -run TestReputationCareerTwoExitsIntegration'
```

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
recalculation, not fresh calibration. The separately declared dated companion
below now supplies fresh H1/H2 provenance and reproduction; H4/H5 provenance
remains open. Neither the local observer correction nor the new measurement
ratifies a threshold or replaces a historical report.

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
27 malformed populations reject, a valid changed clock is accepted, unchanged
sensitivity rejects, and career/Tier2 input overrides reject. Compiling input,
wrong-factor, census and sensitivity omissions make these controls fail. That
fast control is not the full study and does not make H3 green.

Historical H1/H2 recalculation is distinct from the fresh dated H1/H2 companion
below. That measurement does not accept this fired H3 criterion. Existing H4
ties, H5 epsilon/run4 and report drift remain; no historical-report replacement,
threshold/policy change or epoch mint follows.

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
H4/H5 or AC13 acceptance. The RFC's absent named career-data artifact and H4/H5
report-envelope provenance remain open; fresh H1/H2 evidence is separate below.

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
H4/H5 lineage remains open; dated H1/H2 lineage is separately declared below.
The H3 aggregate-coordinate gap is locally corrected as described below.
The original six omission probes were executed
at the H2 repair checkpoint, before adding the shared H3 coordinate guard;
the H2 guard is now an independent defense, not proof the shared one is present.
No old report regeneration or H2/AC13 completion follows from this local repair.

The shared H3 population oracle now also checks aggregate schema, scenario id,
scenario hash and constants hash against its declared suite. Its synthetic
positive fixture carries valid aggregate coordinates before each field is
corrupted independently; a blank-header fixture is not a valid positive source.
The compiling guard omission fails all four H3 cases while the H2 boundary stays
green through its separate admission. Aggregate value arrays are not re-derived.
The complete current five-arm study rerun observes 485 runs/3,395 clocks and
passes source admission and neutrality, but tiny sensitivity remains 0/679
against strong 289/679. This does not make H3 or AC13 green.

The separate missing H4 career-data artifact cannot silently freeze the current
test policy: R10 says purchases at each Exit, whereas R6 forbids a plan on the
scripted first Exit. The harness currently purchases after run2's collapse.
That coverage/intent ambiguity needs RFC-author body reconciliation; no new
scripted-first behavior or owner choice is inferred by the implementer.

## Fresh first-hour and threshold evidence (2026-10-06)

The dated [H1 report](../planning/reputation-tree-v1/first-hour-reputation.2026-10-06.v1.json),
[H2 report](../planning/reputation-tree-v1/threshold-measurement.2026-10-06.v1.json)
and [lineage companion](../planning/reputation-tree-v1/measurement-lineage.2026-10-06.v1.json)
are a new, fully executed measurement, not replacements for the historical v1 files.
Both record and independent full replay complete all 97 ratified seeds and 679
milestone clocks, using the unchanged epoch-8 catalog and experiment. The entire
first-hour aggregate is recomputed and must have no fired criteria before recording.
H2 derives all 30 thresholds using that suite's exact Prestige catalog; raw H1/H2
SHA-256, committed server/balance trees, kernel and runtime identity are retained.
These are provenance/reproduction facts, not a cryptographic authenticity claim.

Only Reference's two lifetime observations differ from historical H1; its clocks,
ending and all other runs/aggregate are unchanged. Reference H2 payout changes
at thresholds `1e4` (0→1), `2e4` and `5e4` (1→0). Casual/Chaos statistics and the
satisfying proposal set `2e4`, `5e4`, `1e5`, `2e5` are unchanged. The live `1e12`
row remains zero for every persona and fails the envelope. **No threshold is adopted.**

The fast artifact test resolves recorded Git identities, checks hashes/census,
recomputes aggregate and H2, and explicitly does not claim a fresh producer run.
The fresh test is opt-in (`off`, `record`, `verify`; unknown values fail). Record
requires committed clean inputs and exclusively creates the declared paths; an
attempt to repeat it fails before production, rather than overwriting evidence.
Verify requires the same committed server/balance trees and kernel, reruns the
whole population and demands byte-identical H1/H2. A future source-tree change
requires a separately declared new measurement, not relabelling this one as current.

Reproduce this recorded producer while those trees still match:

```sh
make test-go GO_PACKAGES='./harness -args -reputation-current-measurement=verify' \
  GO_TEST_FLAGS='-run "^TestReputationCurrentMeasurement\z" -count=1 -v -timeout 5m'
```

Twenty-four refusal cases and four compiling guard omissions discriminate raw
binding, semantic aggregate/result/grid corruption and source-tree drift. The
recorded producer is `da512cd39b1e49e5fc734cd0e05be6a11a73fb2e`; its local full
record/replay take 68.91/62.30 seconds. Cold local fast harness and core gates pass.
This does not claim hosted CI, SQL, browser, H3/H4/H5 or AC13 acceptance. RP-263's
H4/H5 report-envelope provenance and all designated review ranges remain open.

### Career experiment-source observations

`RunReputationCareer` now returns `measurement_source` in addition to the unchanged
catalog RunKey. It records that complete paired key, first-hour policy hash, hash
of the canonical serialized **effective loaded Prestige policy**, experiment tuple,
effective horizon, purchase policy and excluded node. In particular, a fixture
threshold override is not hidden behind a catalog hash containing the live policy.
This identifies observed inputs, not software provenance or policy adoption.

Both report consumers admit the result against their declared experiment before
projection, including its actual run key/policy and completed outcome. H4 retains
treated and no-purchase control sources on every row; H5 retains every baseline
and leave-one-out arm's source in its report header. Four actual Chaos seed0
controls prove the catalog-key collision while their complete gameplay-result
fingerprints remain unchanged after adding metadata. Thirty-six corrupt-source
profiles reject through H4 treated/control and H5 projection. Independent input,
caller-admission and source-retention omissions make the corresponding tests fail.

The historical H4/H5 v1 files are not rewritten: the corrected producer must
reject their old projection shape as well as previously observed clock drift.
The complete local study now finishes all 97 treated/control pairs and 970 H5
arms: source admission passes for 194/970 arms, with all 970 H5 sources retained.
It takes 636.768 seconds and exits red on both preserved historical comparisons;
H4 still observes 93 timing comparisons, three valid exclusions and six ties.
These observed counts do not turn the focused corruption cases into naturally
earned populations or make the acceptance gates green. H4's strict-sooner failure,
H5's unruled epsilon/run4, missing career-data authority and the each-Exit intent
ambiguity remain. H5 originally omitted its finite-pair denominator (RP-272);
the population observations below expose it but do not supply a censoring rule
or certify the median as a
full-persona result. The earlier dated H1/H2 evidence remains pinned to its recorded
producer tree, not silently promoted to this changed server tree.

Additional report-coverage findings remain explicit. H4's savings map currently
includes only strictly-faster finite pairs: historical Casual data have 28 finite
eligible pairs (six ties), but its published min/p50/max use 22 strictly-faster
pairs. All-finite minimum is 0 ms; the published minimum is 15,000 ms. That map
is a pass-only summary, not the full finite distribution (RP-273). H4 also lacks
the per-node aggregate required by R10 (RP-274); shared package savings must not
be labelled isolated node effects. Neither gap is repaired by source metadata.

The test-side report producers now expose population counts (RP-272/273).
H4's `population_by_policy` covers every admitted row, distinguishes rows
without starters, and partitions starter pairs into finite faster/tied/slower,
treated-only, control-only and both-unreached observations. The existing
`saved_ms_min_p50_max_by_policy` remains **strictly-faster finite pairs only**;
`finite_starter_pair_saved_ms_min_p50_max_by_policy` separately describes all
finite starter pairs, including zero and negative savings. Neither assigns a
saving to an unreached clock or attributes a compound package effect per node.

H5's node-level `population_by_policy` exposes baseline careers, not-purchased
careers and the same clock-state partition for bought pairs, including zero-
bought/zero-finite populations. Its existing median is conditional on **bought
and both clocks finite**, not all bought careers. The estimator, epsilon,
classifier, horizon and purchase policy are unchanged; the absence of a median
still does not mean a zero effect. Historical report files remain untouched.
Synthetic controls and eleven compiling omissions verify counts, conservation,
finite statistics and JSON retention. The full current study at `b22ce51e`
finishes 97 H4 pairs and 970 H5 arms in 547.497 seconds. Casual H4 has 32 starter
pairs: 28 finite (22 faster, six tied), one treated-only and three both-unreached.
Its all-finite savings are `[0, 80000, 350000]` ms; Chaos has 64 finite faster
pairs with `[38000, 118000, 212000]` ms. Reference has no starter, not a zero
starter effect.

For Casual H5, cash-small's 70,000 ms median uses **28/32** bought pairs; one
is treated-only and three both-unreached. The tower's 15,000 ms median uses
**20/23** bought pairs, with three both-unreached and nine other careers that
never bought it. All nine nodes and all three persona groups are observed,
including zero-purchase groups; the exact census is in the implementation log.
Cold fast harness and vet pass. Both historical comparisons and H4's six-tie
criterion remain RED. This is no censoring adoption, H4/H5 acceptance,
current-artifact software-provenance closure or RP-274 attribution resolution.

For RP-275, H5 now retains `arm_observations`: each admitted source, observed
run-3 gate (including null/unreached), and purchased-node list. Mutable clocks
and lists are copied at retention. The ordinary H4/H5 checks use shared test-
side measurement/report builders; later dated evidence must use those same
builders, not a second estimator. No new dated artifact is generated here.

H5 recomposition refuses missing/duplicated baseline or mask groups, unknown
masks/nodes, masked-node purchases, negative clocks and mismatched paired
sources. It rebuilds medians, populations and classifications from retained
arms and compares the whole report, including sources and header fields. A
bought starter with no qualifying effect or exclusion still fails the existing
H5 rule, even if its report is internally consistent. Synthetic controls and
twelve compiling omissions exercise these boundaries. Full current caller
coverage at `d96bfaf1` completes 97 H4 pairs and 970 H5 arms in 641.626 seconds;
all 970 raw observations are retained and successfully recomposed. Observed
medians and populations match the preceding study, including the unchanged
six H4 ties. Cold fast harness and vet pass; both historical comparisons remain
RED. No new dated artifact is generated in this preparation stage.
Recomposition is internal consistency, not
proof that a synthetic cohort was earned or that a producer is authentic;
exact committed software/input identity and full replay remain separate.

The private opt-in `TestReputationCareerCurrentMeasurement` lane now uses those
same builders. `-reputation-career-measurement=record` refuses existing outputs
and uncommitted inputs; `verify` requires the recorded source/data/kernel trees
and native runtime before executing the entire population again. Both modes
admit the ordered 97 H4 pairs and 970 H5 baseline/mask arms against the pinned
declaration, recompose complete reports, and cross-bind H4 treatment to H5's
baseline. Raw hashes, committed producer identity, runtime and census live in
`career-measurement-lineage.2026-10-06.v1.json`. Reproduction requires identical
H4/H5 bytes. The retained-artifact check is explicitly not fresh execution;
before the recording commit it explicitly reports absence, never evidence.

Synthetic complete-cohort controls include internally consistent forgeries;
eight compiling guard omissions discriminate. They test admission, not earned
player progress or payout mechanics. A truthful H4 FAIL remains a valid negative
observation, not acceptance. At the instrumentation checkpoint no dated career
artifact had yet been recorded or reproduced. Original reports, gameplay,
statistics, threshold, horizon, epsilon, classifier and strict gates are unchanged.

Subsequently the dated career reports were recorded and completely replayed
from `6bd24e8d9e211b68aa0bb04dd18f4f02653616a8` (server tree
`c3d6616e9b33ff94e71ea452e0c1d5ffddebbef5`, unchanged balance tree/kernel
0.3.161, Go 1.27.1/darwin/arm64). Each execution covers 97 treated/control
career pairs and all 970 baseline/mask arms. Record completes in 676.102 seconds;
verify completes in 469.043 seconds with byte-identical H4/H5 reports. These
are execution durations, not a performance comparison or acceptance budget.

The dated H4 SHA is `29d0e6d2758502db3e82578db349cc1e080535afb74be0ae77d5df0e0ca2f3f7`;
H5 is `8044667158f80ccd920cce5861bd029fa3d46339c77fdb4a7230b6d3e9b404fe`.
All populations/medians and the six strict H4 failures remain unchanged. A real
overwrite attempt refuses before production. Retained-artifact validation and
the cold fast harness lane pass with these files present. The companion binds
both files to the producer; replay authenticates this declared observation,
not player/default-host/SQL integration, H5's missing elective-Exit dimension,
an unruled epsilon/censoring policy, literal adoption or release readiness.
Earlier dated H1/H2 retain their earlier producer; they are not restamped as
fresh at this changed server tree. All ranges still need designated review.
