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
7. `starter` rows use the curriculum starter union, validated by `curriculum.ValidateStarter`.
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

Both runtimes are bound by two shared corpora:
- `testdata/reputation/tree-fixtures-v1.json`: one or more rejection fixtures per rule.
- `testdata/reputation/bonus-vectors-v1.json`: Go-authored vectors. Setting
  `REPUTATION_UPDATE_VECTORS=1` regenerates them from Go.

## Bundle wiring

`reputation_tree` is an optional epoch artifact, loaded by `server/replaycatalog` and
`client/src/replay.ts` into `CatalogBundle.ReputationTree` / `reputationTree`.

- **Chain.** It requires `minigame_api`, the artifact that owns Founder v21. Its own Founder save
  version lands in B3, so no version floor is raised yet.
- **Pairing (R2).** The artifact is present exactly when the economy declares a multiplier source
  with provider `reputation_tree`. If either is present without the other, the bundle is rejected.
  Go checks this again in `CatalogBundle.valid`.
- **Frozen contributions.** `production.ResolveFrozenContributions` now accepts that provider as
  well as `fiscal`, ready for the run-frozen `reputation.founder_bonus` row (B5).

## Founder save v22

`save.LatestFounderVersion` is 22. Founder v22 adds the required fields `reputation_spent` and
`reputation_nodes_owned` (byte-sorted and unique). The pinned bundle owns v22 when it carries
`reputation_tree`, so its `versionFloors` founder floor is 22.

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

The TypeScript `restoreFounderReplayState` / `encodeFounderReplayState` enforce the same rules and
now carry the unlock mirror; before v22 the mirror is always 0.

**Temporarily fail-closed:** the Company-side Founder carry in replay inputs has no tree fields
yet. Reconstructing a v22 Founder from a carry therefore fails closed, in Go and in TS, until the
next replay-inputs version (R6) adds them.

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
