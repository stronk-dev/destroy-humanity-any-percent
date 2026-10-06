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

An applied purchase adds the cost to `reputation_spent`, inserts the id in sorted order, and updates
the unlock mirror. It returns a receipt with `effective_from: "next_run"` (a Fiscal sweep, if any,
decorates it) and emits `reputation_node_purchased.v1` with `source: "direct"`. The event is
validated strictly in `save.validateEvent` and admitted to the database by migration 00075.

`testdata/replay/reputation-tree-v1.json` is the Go-authored cross-runtime corpus: the inactive,
invalid, unknown, requires, owned and unaffordable rows (including the `cost == available + 1`
boundary), plus a nine-node chain with an automatic Fiscal sweep. Set
`REPUTATION_UPDATE_FIXTURE=1` to regenerate it from Go.

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

`testdata/reputation/starter-boundaries-v1.json` additionally supplies four
fixture-only next-bundle cases: retiring the generated/upgrade nodes; granting
the same preowned upgrade from curriculum and the tree; landing exactly on the
provisioned cap; and landing exactly on the permit cap. Changed catalogs are
loaded strictly under hashes derived from their actual bytes. Removed ids stay
in Founder ownership but grant nothing in the next run. Both loaders refuse
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
