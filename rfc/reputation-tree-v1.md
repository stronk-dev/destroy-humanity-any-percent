# RFC: Reputation Tree v1

- **Status:** draft — not implementation authority
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-09-25
- **Design refs:** `design/02 §2.2` (production stack: `FounderBonus (+1% per Reputation level,
  unlock-gated 5→25→50→75→100%)`), `design/02 §3.1` (prestige formula; "the benefit is
  unlock-gated … bought with Reputation: the first several Exits are about buying the right to
  benefit from Exiting"; first-elective-Exit pacing), `design/02 §3.2` (Reputation is the
  heavenly-chips equivalent, "spent in a permanent tree (offline extension, starter packages,
  synergy unlocks, golden-opportunity upgrades, permanent-slot purchases)"; Network slots I–V
  bought with Reputation), `design/02 §2c` (every cap mints a currency), `design/02 §11`/`§11b`
  (pacing targets; every purchasable meaningful and measured), `design/07` Phase 1 ("first Exit
  loop, Reputation tree v1"), `design/11 §3–4` (Founder card, `[NEW ROUTE]` carry-over summary;
  "each Exit must advance something visible (Reputation tree, …)"), `design/11 §7` (tabs, depth ≤ 2,
  keyboard operability, curtain rule), `design/10 §4` (challenges are gated in the Reputation tree —
  deferred here), `design/08 §1` (voice rules — no copy is authored by this RFC)
- **Depends on:** Prestige & Exits (implementing — owns the Exit transaction, `reputation_level`,
  `reputation_unlock_ppm`, and the new-run assembly order D6), Founder Attendance / Founder-scoped
  transitions (implemented — `ApplyFounderLogged`, `founder_log`), Fiscal Quarters Foundation
  (implemented — the `run_frozen_contributions` precedent this RFC reuses), T0–T1 Playable Content
  (implemented — the curriculum `starter_package` vocabulary this RFC reuses), Purchasable Content
  Foundation (implemented), Game-UI Screens (implemented — snapshot v3, surfaces), Copy Pipeline
  (implemented), Balance Harness / Relevance Harness (implemented), Leaderboards & Balance Epochs
  (implementing — epoch mint protocol)
- **Parent / amends:** amends Prestige & Exits D1/D6 append-only (optional Exit-attached purchase
  plan on `accept_exit_offer`/`wind_down`, conditional on OD-3; Reputation starter effects and the
  Founder bonus fill the declared-but-empty D6 seam). No archived behavior changes.
- **Supersedes / superseded by:** —
- **Planning:** `planning/reputation-tree-v1/` (once implementing)

## Summary

Reputation is earned today and does nothing. `reputation_level` accumulates at every Exit, the
Founder save carries a `reputation_unlock_ppm` field that no path writes, the production stack has a
`prestige` slot whose only occupant is Fiscal, and `prestige.NewRunState` applies "Reputation
starter effects" from a seam that is empty. This RFC makes Reputation spendable in a small,
data-declared permanent tree whose v1 content is exactly the two node families the design names and
the kernel can already execute: the **unlock ladder** that turns on `FounderBonus` (+1% production
per Reputation level × 5/25/50/75/100% unlock), and **starter packages** that pre-seed each new
Company run through the existing curriculum starter vocabulary. It specifies the artifact, one
Founder intent (`purchase_reputation_node`), an optional Exit-attached purchase plan so spending
pays off in the same session, Founder save v22, the frozen per-run contribution, the Game-UI
surface, harness gates, and the owner decisions — including a measured finding that under the
live `threshold = 1e12` the Phase-1 window earns **zero** Reputation, so a tree with no retune would
be unreachable content.

## Motivation

`design/07` Phase 1 lists "first Exit loop, Reputation tree v1" as a v0.1 deliverable, and
`design/11 §4` requires every Exit to advance something visible. Today the only thing a first
elective Exit visibly advances is Route Knowledge (collapse's flat 25), because:

1. **No consumer exists.** The archived first-hour payoff design recorded it at source:
   "Reputation and Route Knowledge appear only in exit bookkeeping and feed no multiplier"
   (`planning/archive/t0-t1-content/first-hour-payoff-design.md`). The epoch-8 curriculum
   starter is a one-time run-2 payoff, not a Reputation mechanic.
2. **No Reputation is earned in the window.** Level = `floor(cbrt(lifetime_value / 1e12))` against
   the **per-run** `lifetime_value` (reset to zero by `prestige.NewRunState`). The T0–T1 economy
   reaches roughly `1e4` cash at the scripted first ending (archived experiment
   `first-hour-experiment-proposed.v1.json`, `runs[].ending.cash`) and its top Phase-1 generator
   produces `4.9e5/s` base; a first elective Exit at 45–90 min cannot approach `1e12`. The level —
   and therefore the payout — is 0. (Exact first-elective lifetime values are **not** in any
   committed report; H1 below makes them a recorded field. The conclusion "level 0 under 1e12" is
   an order-of-magnitude bound, not a measurement, and is labelled as such.)

**In scope:** Reputation accounting (earned / spent / available); the `reputation_tree` artifact and
loaders (Go + TS); the FounderBonus provider in the `prestige` slot; starter-node application at
new-run assembly; `purchase_reputation_node`; the Exit-attached plan (OD-3); Founder save v22 and
DB migrations; replay/verification parity; the Game-UI tree surface and plan panel; harness
observation and gates; the mint of one new epoch.

**Out of scope (named successors, not silently dropped):** offline-cap extension, synergy unlocks,
golden-opportunity upgrades, Network slots I–V and their designation intent, challenge-run gating,
Consultancy unlock, respec/refund, Tier-2 era UI (`era_2010`). Each is a DESIGN-GAP below or an
owner decision; none is improvised here.

## Grounding — what exists at HEAD (`c2d9bbc`)

| Fact | Where |
|---|---|
| Founder `reputation_level` (int, exact domain) credited `+= terms.reputation_delta` at Exit | `server/production/prestige.go` `finishExitResolved` |
| Founder `reputation_unlock_ppm` (0..1,000,000), persisted, never written, codec-bounded | `server/save/state.go` (`ReputationUnlockPPM`) |
| Reputation delta = `level(run lifetime) − founder.reputation_level`, × exit modifier ppm, floored; saturating | `server/prestige/math.go` `ReputationDelta`; `balance/prestige/phase0.json` (`collapse` 750,000 ppm) |
| `prestige` multiplier slot, applied last; raw-byte source order within slot | `docs/production-engine.md` |
| Founder-derived production is frozen per run in `run_frozen_contributions`, completeness enforced by a deferred trigger | `docs/fiscal-quarters.md`; migration `00065_run_frozen_contributions.sql` |
| Fiscal's `ppmFactor`: `1 + FromString(count×ppm)/1e6`, quantized once | `server/fiscal/catalog.go` |
| Starter vocabulary `resource_grant` / `generated_generators` / `preowned_upgrade` | `balance/curriculum/t0-t1.json`, `server/curriculum/catalog.go` |
| Founder commands mutate only through `ApplyFounderLogged`; closed Phase-A union | `docs/founder-transitions.md` |
| Founder save chain v17–v21 (minigames, pets, Fiscal, Soul, session seq); activation only at a new-run boundary under a pinned artifact | `docs/save-layer.md`, `docs/fiscal-quarters.md`, `docs/soul.md` |
| `LatestFounderVersion = 21` | `server/save/state.go` |
| Game UI snapshot v3, five surfaces, intents via `POST /api/v1/intents` | `docs/game-ui.md`, `client/src/game-ui/surfaces.json`, `server/gameserver/server.go` |
| An artifact cannot be added to a minted epoch | `TestEpochGuardRejectsArtifactAdditionAsHotfix` (per archived payoff design) |

## DESIGN-GAPs found while drafting

Each gap lists options and the recommended resolution; the binding choice is the matching owner
decision.

- **DG-1 — Is Reputation a level or a balance?** `design/02` uses both ("+1% per Reputation level";
  "bought with Reputation"; "heavenly-chips equivalent"). In the Cookie Clicker lineage the prestige
  level drives the bonus and chips are spent separately without lowering it. The code makes this
  more than taste: `ReputationDelta` pays `level − founder.reputation_level`, so **if spending
  lowered `reputation_level`, the next Exit would refund the spend.** Options: (a) keep
  `reputation_level` as the earned high-water total and add `reputation_spent`; available =
  level − spent; (b) spending debits the level. → OD-1, recommend (a).
- **DG-2 — When does a purchase take effect?** The Exit transaction creates run N+1 atomically
  (Prestige P4), and Fiscal ruled that Founder-derived production never reads mutable Founder state
  during a Company command. A purchase made on the run-end screen therefore could only apply to run
  N+2 — a full run of delay, contradicting "obviously worth it same-session". Options: (a)
  next-run-only (Fiscal precedent); (b) (a) plus an optional purchase plan carried on the terminal
  Exit intent, executed inside the Exit transaction before new-run assembly; (c) re-materialize
  run N+1 while it has no commands — rejected: violates immutable `run_genesis` and immutable
  frozen rows. → OD-3, recommend (b).
- **DG-3 — The threshold earns nothing in Phase 1.** See Motivation §2. Additionally the collapse
  modifier floors small deltas: a level-1 Wind Down pays `floor(1 × 0.75) = 0`. Any envelope must
  therefore be stated on **paid** Reputation, not level. → OD-2.
- **DG-4 — Network slots need a designation contract.** "Designate an owned upgrade/person to carry
  into every future run" has no intent, carry-eligibility rule, or conflict rule with starters. Out
  of scope; successor RFC.
- **DG-5 — Offline extension interacts with the Compute Credit bank.** Excess offline time mints
  Compute Credits (`bank_ratio 1/2`); extending `accrual_cap_ms` silently converts bank into
  production. The design does not say which wins. Out of scope; successor RFC.
- **DG-6 — Tree changes across epochs.** Owned nodes are permanent Founder facts, but node costs,
  effects, or existence may change in a later epoch. The design is silent. → OD-7.
- **DG-7 — Relevance across runs.** `§11b` requires every purchasable to be measurably meaningful,
  but the relevance harness measures within one run and tree nodes act on the next. Additionally
  the design's own formula makes the bonus tiny at low levels (level 4 × 5% unlock = +0.2%), so the
  first unlock node is expected to fail a strict per-node ε gate in the Phase-1 horizon. → OD-5.
- **DG-8 — Respec.** The design neither offers nor forbids refunds. v1 ships none (no mechanic
  invented). → OD-8.
- **Observation (not a gap in this RFC):** `design/02 §3.1` says "lifetime-based (not per-run)";
  the shipped Prestige contract evaluates the per-run `lifetime_value` against the Founder's
  high-water level ("beat your record"). This RFC consumes that behavior unchanged and flags it for
  the owner because it shapes every Reputation number below.

## Specification

### R1 — Reputation accounting

Founder state gains `reputation_spent` (exact non-negative safe integer) and
`reputation_nodes_owned` (sorted set of node ids). Definitions, identical in Go and TypeScript:

```text
reputation_level      earned high-water total (existing; only Exit writes it)
reputation_spent      sum of the recorded cost of every purchase (only purchases write it)
reputation_available  = reputation_level − reputation_spent            (derived; never persisted)
reputation_unlock_ppm = max(unlock_ppm of owned bonus_unlock nodes), or 0 (derived AND persisted)
```

Invariants enforced by the v22 codec (load **and** encode): `0 ≤ reputation_spent ≤
reputation_level`; every owned id is unique and sorted; `reputation_unlock_ppm` equals the derived
value against the pinned tree artifact. `reputation_unlock_ppm` stays persisted because Prestige P1
declared it; it is a checked mirror, never an independent authority. Spending never changes
`reputation_level`, so `ReputationDelta` and Exit previews are unaffected by purchases (DG-1).

### R2 — The `reputation_tree` artifact

Path `balance/reputation-tree/phase1.json`, schema `balance/reputation-tree.schema.json`, artifact
name `reputation_tree`, optional in the bundle (Fiscal precedent), bytes in the constants hash.

**Exact shape (closed; unknown or missing keys reject):**

```json
{
  "schema_version": 1,
  "bonus": {
    "per_level_ppm": 10000,
    "source_id": "reputation.founder_bonus",
    "slot": "prestige",
    "target": "all"
  },
  "nodes": [ <node row>, ... ]
}
```

Node rows are a closed union on `kind`:

```json
{ "node_id": "reputation.unlock.p05", "kind": "bonus_unlock", "cost": 1, "requires": [],
  "unlock_ppm": 50000,
  "title_key": "reputation_tree.node.unlock_p05.title",
  "body_key":  "reputation_tree.node.unlock_p05.body" }

{ "node_id": "reputation.starter.cash_small", "kind": "starter", "cost": 2, "requires": [],
  "starter": { "kind": "resource_grant", "resource_id": "company.cash", "amount": "1e3" },
  "title_key": "reputation_tree.node.starter_cash_small.title",
  "body_key":  "reputation_tree.node.starter_cash_small.body" }
```

`starter` is byte-for-byte the curriculum `starter_package` union (`resource_grant {resource_id,
amount}`, `generated_generators {generator_id, count}`, `preowned_upgrade {upgrade_id}`); no new
starter kind is introduced.

**Loader rules (Go and TS; each is an AC1 rejection fixture):**

1. `schema_version == 1`; `per_level_ppm` in `1..1_000_000`; `(source_id, slot, target)` must match
   one economy `multiplier_sources` row with `slot = prestige`, `target = all`,
   `provider = reputation_tree` (the Fiscal `validDeclaration` rule).
2. `1 ≤ len(nodes) ≤ 64` (structural bound on plan payload size; not balance).
3. `node_id` matches the mechanical-ID grammar, begins with `reputation.`, and is unique.
4. `cost` is an integer in `1..9_007_199_254_740_991`.
5. `requires` is byte-sorted, duplicate-free, excludes the row itself, and names only rows **earlier
   in the array** (array order is a topological order, which makes cycles unrepresentable; array
   order is also UI display order).
6. `bonus_unlock` rows: at least one; `unlock_ppm` in `1..1_000_000`, strictly increasing in array
   order; each row after the first `requires` its predecessor bonus_unlock; the last row's
   `unlock_ppm` is exactly `1_000_000` (the design's 100%).
7. `starter` rows: validated by the curriculum `validateStarter` rules against the pinned economy
   catalog, plus aggregate headroom so application can never fail at Exit:
   - for every resource: `initial + max(curriculum resource_grant for it, 0) + Σ tree grants ≤
     hardcap`;
   - for every generator: `max(curriculum generated count for it, 0) + Σ tree generated counts ≤
     provisioned_hardcap.count` (or `MaxExactInteger` when undeclared);
   - `preowned_upgrade` ids are unique within the tree.
8. `title_key`/`body_key` are registered copy-bearing paths in `copy/references.v1.json` and
   resolve in the copy catalog (`make copy-check`).

The economy artifact gains exactly one declaration row:
`{"id": "reputation.founder_bonus", "slot": "prestige", "target": "all", "provider":
"reputation_tree"}`. A tree artifact without it, or it without a tree artifact in the same bundle,
rejects the bundle.

**PROPOSED starting data (retunable; ratified only with OD-2/OD-10 measurement):**

| # | node_id | kind | cost | requires | effect |
|---|---|---|---|---|---|
| 1 | `reputation.unlock.p05` | bonus_unlock | 1 | — | unlock 50,000 ppm (5%) |
| 2 | `reputation.starter.cash_small` | starter | 2 | — | `resource_grant company.cash 1e3` |
| 3 | `reputation.starter.generated_beige_tower` | starter | 3 | #2 | `generated_generators generator.beige_tower 5` |
| 4 | `reputation.unlock.p25` | bonus_unlock | 5 | #1 | unlock 250,000 ppm |
| 5 | `reputation.starter.upgrade_continuous_feed_paper` | starter | 4 | #3 | `preowned_upgrade upgrade.continuous_feed_paper` |
| 6 | `reputation.starter.cash_large` | starter | 12 | #2, #4 | `resource_grant company.cash 1e5` |
| 7 | `reputation.unlock.p50` | bonus_unlock | 25 | #4 | unlock 500,000 ppm |
| 8 | `reputation.unlock.p75` | bonus_unlock | 100 | #7 | unlock 750,000 ppm |
| 9 | `reputation.unlock.p100` | bonus_unlock | 400 | #8 | unlock 1,000,000 ppm |

`bonus.per_level_ppm = 10000` is the design's literal +1%/level. Starter choices avoid
`upgrade.reply_all_macro` (the curriculum pivot branch's grant) so no node is dead for a class of
Founders. `requires` arrays are shown by row number here; the file carries sorted node ids. Total
cost of the full tree: 552.

### R3 — FounderBonus arithmetic and the frozen contribution

For the Founder state **after** the Exit's Reputation credit and any Exit-plan purchases (R6):

```text
numerator = reputation_level × per_level_ppm × reputation_unlock_ppm      (big.Int / BigInt)
factor    = FromString(numerator).Div(1e12).Add(1).Quantize(CanonicalSignificantDigits)
```

This is the Fiscal `ppmFactor` construction with a `1e12` denominator (two ppm scales). The factor
must be state-valid and `≥ 1`. No hardcap is introduced (OD-12): the input domain is already bounded
by `reputation_level ≤ MaxExactInteger` and `unlock_ppm ≤ 1e6`.

When the **next run's** pinned bundle contains `reputation_tree`, the new-run materialization
inserts exactly one `run_frozen_contributions` row
`(company_stream_id, run_seq, "reputation.founder_bonus", "prestige", "all", factor)` in the same
transaction as the run pin and genesis — including `factor = "1e0"` for a Founder with no unlock
node. The same applies to every path that creates a run: Exit, New-Founder initialization, and
import initialization. Production resolves the row only by Company stream and run sequence through
the existing frozen-contribution replay path; it never reads Founder state during a Company
command. **Mid-run purchases change the next run only.**

### R4 — Starter application at new-run assembly

Prestige D6's order becomes, exactly:

1. catalog initials (`prestige.NewRunState`);
2. `settleAndActivateFoundations` (unchanged);
3. curriculum starter, when the Exit is `scripted_first` (unchanged; retains its assignment
   semantics on a fresh run);
4. **Reputation starters:** for each owned `starter` node in **tree array order**, apply
   additively — `resource_grant` credits the ledger; `generated_generators` does
   `provisioned[g] += count`; `preowned_upgrade` sets ownership (idempotent if already set);
5. active-play initialization and the rest of D6 (unchanged).

Generated units remain unpurchased (the `§11b` purchased/generated split). R2 rule 7 guarantees no
step can exceed a cap; an out-of-cap result is therefore `ErrInvalidEngineState`, never a clamp.
Owned `bonus_unlock` nodes contribute only through R3. Only nodes known to the **next** bundle's
tree apply (OD-7).

### R5 — Intent `purchase_reputation_node`

Founder-stream command on `POST /api/v1/intents`, dispatched to `ApplyFounderLogged`, added to the
closed Founder Phase-A union.

**Request (exact keys):**

```json
{ "intent_id": "<uuidv7>", "kind": "purchase_reputation_node",
  "expected_revision": <Founder revision, positive safe int>, "node_id": "<mechanical id>" }
```

**Validation order** (first failure wins; every rejection is a recorded terminal decision with no
state change and no event):

| # | Check | Result |
|---|---|---|
| 0 | exact keys, UUIDv7, positive safe revision, mechanical `node_id` | `invalid / purchase_reputation_node.fields` (or the failing field name) |
| 1 | revision CAS; idempotency hash | unrecorded `revision_conflict` / `idempotency_conflict` (existing) |
| 2 | Fiscal automatic sweep, when Fiscal is active (existing Founder-command rule; rolls back with a rejection) | — |
| 3 | pinned Founder bundle lacks `reputation_tree`, or Founder version < 22 | `not_eligible / reputation_tree_inactive` |
| 4 | `node_id` not in the pinned tree | `unknown_id / <node_id>` |
| 5 | node already owned | `not_eligible / owned` |
| 6 | any `requires` id not owned | `not_eligible / requires` |
| 7 | `cost > reputation_available` | `unaffordable / reputation` |
| 8 | apply | `reputation_spent += cost`; insert id; recompute `reputation_unlock_ppm` |

Categories and details mirror `buy_upgrade` (`not_eligible/owned`, `not_eligible/requires`) and
Fiscal (`unaffordable/<currency>`); no new category joins the taxonomy.

**Resolved inputs** (frozen in `founder_log`; replay recomputes and compares):

```json
{ "kind": "purchase_reputation_node", "node_id": "...", "resolved_cost": <int>,
  "reputation_level": <int>, "reputation_spent_before": <int>, "owned_before": ["..."] }
```

`resolved_cost` is `0` for a rejection at step ≥ 3 (Fiscal precedent).

**Applied receipt:**

```json
{ "intent_id": "...", "outcome": "applied", "founder_revision": <n+1>, "fiscal_sweep": <existing shape or null>,
  "node_id": "...", "resolved_cost": <int>,
  "reputation_available_before": <int>, "reputation_available_after": <int>,
  "reputation_unlock_ppm_after": <int>, "effective_from": "next_run" }
```

**Event** `reputation_node_purchased.v1` (Founder scope, one per applied purchase):

```json
{ "node_id": "...", "node_kind": "bonus_unlock" | "starter", "cost": <int>,
  "reputation_level": <int>, "reputation_spent_before": <int>, "reputation_spent_after": <int>,
  "unlock_ppm_after": <int>, "source": "direct" | "exit_plan" }
```

Strict validator: `spent_after = spent_before + cost ≤ reputation_level`, `cost ≥ 1`,
`0 ≤ unlock_ppm_after ≤ 1e6`, closed enums.

### R6 — Exit-attached purchase plan (conditional on OD-3 = b)

`accept_exit_offer` and `wind_down` gain one **optional** key `reputation_plan`: an array of
0–64 unique node ids, in purchase order. Absent and `[]` are the same request semantically but
hash differently (the key's presence is part of the canonical bytes); the client omits the key for
an empty plan so every existing request remains byte-identical. `file_ipo` and the command-replaced
`scripted_first` never carry a plan.

Inside the Exit transaction, after Prestige D3 step 3 credits `reputation_level` (commit-time
payout, which is ≥ preview):

1. If the plan is non-empty and the **next** bundle lacks `reputation_tree` → reject the Exit
   `not_eligible / reputation_plan.tree_inactive`.
2. Apply R5 steps 4–8 to each id in order against the evolving Founder state, so a node may require
   one earlier in the same plan. The first failure rejects the **whole Exit** as
   `<category> / reputation_plan.<detail>` (e.g. `unaffordable / reputation_plan.reputation`),
   recorded like any terminal rejection; nothing commits. `wind_down` without a plan stays
   available, so the always-open door is preserved.
3. Emit one `reputation_node_purchased.v1` (`source: "exit_plan"`) per applied id on the Founder
   revision, ordered after `founder_advanced`.
4. Then R3 and R4 run against the post-plan Founder state.

Replay: the plan is part of the canonical request; the Founder log's Exit arm widens append-only to
`exit.v2` carrying `reputation_purchases: [{node_id, resolved_cost}]`; terminal replay inputs gain
the Founder's `reputation_spent` and `reputation_nodes_owned` in the frozen carry (the next free
replay-inputs version). Historical `exit.v1` rows and earlier replay-input versions stay readable
and byte-identical.

### R7 — Save, events, and migrations

- **Founder v22** (next free Founder version at HEAD; renumber if another RFC claims 22 first) adds
  required `reputation_spent` and `reputation_nodes_owned`. It activates only at a new-run boundary
  when the pinned bundle contains `reputation_tree` and the artifacts owning v17–v21. Pre-v22 codec
  rejects the new fields. Company scope never carries them.
- **Migration v21→v22** (and every earlier chain step into v22): `reputation_spent = 0`,
  `reputation_nodes_owned = []`. A pre-v22 save with `reputation_unlock_ppm ≠ 0` is **rejected**, not
  repaired: no shipped path can write it, so a non-zero value is corruption. Reputation earned before
  activation remains in `reputation_level` and is fully spendable after activation.
- **Corpus:** `testdata/save-migrations.json` gains `founder-v21-to-v22`, `founder-v22-spent-over-level`
  (reject), `founder-v22-unlock-mismatch` (reject), `founder-v21-nonzero-unlock` (reject); the
  baseline manifest ratchets its exact case count in the same change.
- **DB migrations (next free numbers; applied migrations are append-only):**
  1. extend the closed event-kind constraint with `reputation_node_purchased` (schema v1) and
     `run_started` schema v2;
  2. `CREATE OR REPLACE` the frozen-contribution completeness function so the expected count is
     `fiscal_rows + (1 if the pinned bundle has a reputation_tree artifact else 0)`;
  3. extend the `founder_log` resolved-arm check for `purchase_reputation_node` and `exit.v2`.
- **`run_started` v2** (append-only widening; v1 stays accepted): adds
  `"reputation_tree": null | {"bonus_factor": "<canonical decimal>", "applied_starter_node_ids":
  ["..."]}` — the `[NEW ROUTE]` carry-over summary's source (`design/11 §3` step 5). `null` exactly
  when the run's bundle has no tree artifact.

### R8 — Replay and verification

`production.ApplyFounderLogged` (Go) and its TS port gain the purchase arm; both consume one
Go-authored corpus (`testdata/replay/reputation-tree-v1.json`) and byte-compare state, receipt,
ordered events, and result constants hash for: every rejection row of R5, a purchase chain across
all nine nodes, a Fiscal-sweep-then-purchase, and each R6 plan case. `LoadFounderHistory` replay
covers a career of purchases across two Exits. The run verifier replays a Company run whose genesis
includes Reputation starters and a non-unit frozen bonus. `make formulas-check` regenerates
`docs/generated/production-formulas.json` for the new prestige-slot provider in a separate
reviewed regeneration commit.

### R9 — Game-UI surface contract

**Snapshot v4** (`GameUISnapshot` becomes v4; `GameUISnapshotV3` retained for stored-bootstrap
receipts per GU-C26 convention; compatibility pin refresh cites this RFC's ruling). Adds one
required object:

```json
"reputation": {
  "tree_active": <bool>,
  "level": <int>, "spent": <int>, "available": <int>, "unlock_ppm": <int>,
  "bonus_factor_this_run": "<canonical decimal>" | null,
  "bonus_factor_next_run": "<canonical decimal>",
  "nodes": [ { "node_id": "...", "kind": "bonus_unlock" | "starter", "cost": <int>,
               "requires": ["..."], "state": "owned" | "available" | "unaffordable" | "locked",
               "title_key": "...", "body_key": "..." } ]
}
```

`tree_active: false` ⇒ `nodes: []` and zero spent. `bonus_factor_this_run` reads the current run's
frozen row (`null` if the run has none). `state` is server-derived (`locked` = some prerequisite
unowned; `unaffordable` = prerequisites met, cost > available). The client never recomputes
eligibility; the receipt is authority. A new shell fact `founder.reputation_tree_active` mirrors
`tree_active`.

**Surface `reputation_tree`** (new row in `surfaces.json`, unlock `fact_equals
founder.reputation_tree_active = true`; one tab, depth ≤ 2). Component
`ReputationTreeSurface.svelte` props:

```ts
{ reputation: ParsedReputation; era: "era_1995" | "era_2000";
  purchase(nodeId: string): Promise<IntentOutcome>; }   // runtime.ts owns transport
```

- **States:** `inactive` (fact false — surface not mounted), `ready`, `purchasing` (one in-flight
  intent; all buy controls disabled, row `aria-busy="true"`), `refreshing` (after an applied receipt
  the runtime fetches the authoritative snapshot before re-enabling controls — the existing
  decline-offer boundary), `offline` (transport not recovered: controls disabled, existing offline
  copy).
- **Rows** render in artifact order: title, body, cost via `Amount`, requirement list (titles, never
  ids), state label, and for `available` a Buy button. `owned`/`locked`/`unaffordable` rows have no
  enabled control; the reason is text, not color alone.
- **Header:** available / level / spent, and both bonus factors with the published formula line
  (curtain rule: the +1%/level × unlock% arithmetic is shown, not implied), and the standing note
  that purchases apply from the next company.
- **Irreversible spend:** Buy reveals an inline Confirm/Cancel pair; focus moves to Confirm.
- **Errors:** a rejected receipt renders inline on the row from `reputation_tree.error.<category>`
  (`not_eligible`, `unknown_id`, `unaffordable`, `invalid`, `revision_conflict`); a revision
  conflict triggers the existing resync. Mechanical ids and details are never rendered.
- **Keyboard:** Tab reaches the header then each row's single control in display order; Enter/Space
  on Buy opens Confirm; Enter/Space on Confirm submits; Escape cancels and returns focus to Buy;
  after the receipt, focus stays on the row and a polite live region announces the result. No
  pointer-only affordance. Reduced motion honored.

**Plan panel (OD-3 = b)** inside the Offer Sheet's accept step and the Desk's Wind Down control: a
collapsed disclosure listing nodes with checkboxes, showing projected available =
`available + preview reputation_delta − Σ selected cost`, disabling selections that are locked
(given earlier selections) or unaffordable. Default is empty, so the existing one-action Exit flow
is unchanged. The plan is advisory UI; the server re-validates (R6).

**Run End:** the Founder card adds available Reputation after the Exit; the `[NEW ROUTE]` summary
renders `run_started` v2's `reputation_tree` block.

**Copy keys** (keys only; **all prose owner-authored, pending** — no text is proposed here, per
`design/08 §1` and evidence rule 6): `reputation_tree.tab`, `.title`, `.intro`,
`.balance.available {amount:integer}`, `.balance.level {amount:integer}`, `.balance.spent
{amount:integer}`, `.bonus.this_run {factor:canonical_decimal}`, `.bonus.next_run
{factor:canonical_decimal}`, `.bonus.formula {per_level_percent:integer, unlock_percent:integer}`,
`.applies_next_run`, `.state.owned`, `.state.available`, `.state.unaffordable`, `.state.locked`,
`.requires {list:string}`, `.action.buy {cost:integer}`, `.action.confirm`, `.action.cancel`,
`.result.applied`, `.error.not_eligible`, `.error.unknown_id`, `.error.unaffordable`,
`.error.invalid`, `.error.revision_conflict`, `.plan.heading`, `.plan.projected {amount:integer}`,
`.plan.clear`, `.run_end.available {amount:integer}`, `.new_route.bonus {factor:canonical_decimal}`,
`.new_route.starters`, and per node `reputation_tree.node.<suffix>.{title,body}` (18 keys for the
proposed data).

### R10 — Harness impact

- **H1 — observation.** The first-hour report records, per Exit, `payout.reputation_delta`,
  `reputation_level`, and `reputation_available` as first-class fields. A run whose Exit terms
  could not be computed is an invalid measurement and fails (never omitted).
- **H2 — Phase-1 Reputation gate.** Chaos p50 and Casual p50 **paid** Reputation at the first
  elective Exit fall inside the OD-2 envelope. Demonstrated failing case: the same scenario under
  the current `threshold = 1e12` must fail H2.
- **H3 — first-hour non-regression.** With empty plans and no purchases (a first-hour horizon earns
  nothing before its first elective Exit), all seven existing milestone distributions are identical
  to the epoch-8 evidence; only hash fields and H1's added fields differ. A multiplier of `1e0` in
  the prestige slot is required to be a no-op; failing case: a factor of `1.000001e0` must move a
  milestone.
- **H4 — career scenario** `balance/testdata/reputation-tree/career-scenario-v1.json`: runs 1–3 at
  the ruled seeds; purchase policies are declared data: Reference = at each Exit, plan the cheapest
  available node repeatedly (ties by array order); Casual = same; Chaos = seeded uniform choice among
  available nodes. Control = same seeds, empty plans. Gate: at every seed where run 3 starts with ≥1
  starter node, run-3 `gate.t0_to_t1` is strictly sooner than control; the report records the
  per-node savings distribution (min/p50/max).
- **H5 — tree relevance** (OD-5): per node × persona, leave-one-out Δ on the following run's Garage
  gate and first elective Exit; each node must show Δ ≥ ε for some persona **or** appear in a
  visible `excluded` list with a reason (`unreachable_in_horizon`, or `owner_exempt:<OD-5>`). An
  exclusion without a reason fails the report.

### R11 — Epoch

One new epoch ("Reputation Tree v1") pins: the new `reputation_tree` artifact, economy bytes with
the one new declaration row, and prestige bytes with the OD-2 threshold. Commits touching balance
data use the `BALANCE-CHANGE:` subject class; the epoch-history guard, `make epoch-hash`, and all
existing content harnesses re-run. Runs pinned to earlier epochs finish unchanged; their Founders
activate v22 at their next Exit.

## Deviations from design

1. **Partial tree.** `design/02 §3.2` lists five node families plus Network slots; v1 ships two
   (unlock ladder, starter packages). The rest are deferred with named gaps (DG-4, DG-5) or OD-6.
2. **Level vs. balance** (OD-1): Reputation spending draws on a derived balance and never lowers the
   level that drives `FounderBonus` and the Exit delta.
3. **Deferred effect timing** (OD-3): purchases apply from the next Company run; the design implies
   but does not specify immediacy.
4. **The 5% tier is purchased** (OD-4): the Founder bonus is 0% until the first node.
5. **"Tree" is a prerequisite DAG** in array/topological order.
6. **Threshold retune** (OD-2) changes a Prestige literal (data, not formula).
7. **Relevance exemption** for low-level bonus-unlock nodes, if OD-5 is taken as recommended.

## Acceptance criteria

Each criterion ships with its demonstrated failing case (evidence rule 1). `-count=1` for every gate.

1. **Loader strictness.** Go and TS reject one fixture per R2 rule (duplicate id, forward/cyclic
   `requires`, non-monotonic unlock, missing final 1e6, starter over hardcap in aggregate,
   undeclared bonus source, unregistered copy key, 65 nodes). *Failing case:* deleting any single
   rule's check makes its fixture load, turning the test red.
2. **Accounting invariants.** Codec rejects `spent > level`, unsorted/duplicate owned ids, and a
   mismatched `reputation_unlock_ppm`. *Failing case:* a v22 fixture with `spent = level + 1` must be
   rejected on load and on encode.
3. **Intent taxonomy.** A table test drives every R5 row and asserts category, detail, no mutation,
   no event, and a recorded `founder_log` row. *Failing case:* removing the `requires` check lets the
   locked-node case apply.
4. **Replay parity.** Go and TS byte-match the R8 corpus; identical retry returns the original
   receipt. *Failing case:* a TS cost off by one fails the byte comparison.
5. **Bonus vectors.** Shared vectors for level 0, unlock 0, level 1 × 50,000 ppm, the proposed full
   tree at level 552, and `MaxExactInteger × 1e6`. *Failing case:* computing from
   `reputation_available` instead of `reputation_level` fails at least one vector.
6. **Frozen-row completeness (real Postgres).** A run pin without the reputation row, or with an
   extra row, fails commit; an idempotent retry reproduces the identical factor. *Failing case:*
   deleting the insertion raises the completeness exception.
7. **Next-run-only.** A mid-run purchase leaves the current run's rate projection byte-identical and
   changes the next run's frozen factor. *Failing case:* a live-Founder-read implementation changes
   the current projection.
8. **Starter assembly.** Run-2 golden extended: burnout curriculum (10 generated) plus
   `generated_beige_tower` (5) yields exactly 15 generated, 0 purchased; order and determinism
   byte-stable. *Failing case:* assignment instead of addition yields 5 and fails the golden.
9. **Exit plan (if OD-3 = b).** Valid plan with an in-plan prerequisite commits atomically; an
   unaffordable last entry rejects the entire Exit with both streams at pre-Exit revisions (fault
   injection at each write boundary as Prestige AC1); an Exit without the key is byte-identical to
   the historical corpus. *Failing case:* partial application of the plan's prefix.
10. **Migration.** New corpus cases pass; the baseline manifest ratchet is in the same change.
    *Failing case:* defaulting `reputation_spent` to `reputation_level` fails `founder-v21-to-v22`.
11. **Activation.** v22 activates only at a new-run boundary with the artifact pinned; a purchase
    on an unactivated Founder rejects `reputation_tree_inactive`; adding the artifact to a minted
    epoch is still refused by the epoch guard. *Failing case:* activating mid-run fails the test.
12. **UI.** Surface and plan panel render every state in Chromium, Firefox, and WebKit; a
    keyboard-only buy → confirm → applied flow and an Escape-cancel flow pass; axe WCAG 2.2 AA; client
    boundary scan passes; no mechanical id is rendered; the Accessibility of Player Workflows floor
    applies once accepted. *Failing case:* a missing copy key throws in tests; an `unaffordable` row
    with an enabled button fails.
13. **Harness.** H1–H5 pass as specified, each with its stated failing case demonstrated in the
    planning log.
14. **Formulas and copy.** `make formulas-check` and `make copy-check` pass; the regenerated formula
    artifact names the new provider. *Failing case:* an unregenerated artifact fails
    `formulas-check`.
15. **Verification.** A composed real-Postgres career (scripted first → elective Exit with a plan →
    run 3) replays through the Founder-history and run verifiers with no `state_divergence`.
    *Failing case:* mutating one frozen factor byte produces `state_divergence`.

## Owner decisions (recommended defaults in bold)

1. **OD-1 Accounting model.** (a) **Earned high-water level plus `reputation_spent`; available =
   level − spent; spending never lowers the bonus base or the Exit delta** / (b) spending debits the
   level (the next Exit would refund it).
2. **OD-2 Threshold and Phase-1 envelope.** **Retune `threshold` by measurement so that paid
   Reputation at the first elective Exit is in [3, 10] at Chaos p50 and Casual p50** (enough for
   nodes #1+#2). `1e8` is an unmeasured placeholder for the first experiment, not a proposal to
   ratify. Alternatives: leave `1e12` (the tree is unreachable in v0.1) or change the collapse
   modifier (a Prestige change).
3. **OD-3 Effect timing.** (a) next-run only / (b) **next-run for direct purchases, plus the optional
   Exit-attached plan (R6)** / (c) re-materialize the fresh run (rejected — breaks immutability).
4. **OD-4 First unlock tier.** **5% is the first purchasable node (cost 1)** — the design's "buying
   the right to benefit" / 5% is a free baseline.
5. **OD-5 Relevance of low-level unlock nodes.** **Report their Δ and list them as
   `owner_exempt:OD-5` in the Phase-1 horizon; revisit at the Tier-2 content epoch** / raise
   `per_level_ppm` above the design's 1% (a design deviation) / accept a failing gate (not
   permitted).
6. **OD-6 v1 scope.** **Unlock ladder plus four starter nodes; defer offline extension (DG-5),
   golden-opportunity, synergy unlocks, Network slots (DG-4), challenge gating** / add offline
   extension now (requires a DG-5 ruling first).
7. **OD-7 Cross-epoch policy.** **Node ids are append-only across epochs (loader guard against the
   previous epoch's tree); owned nodes apply the current epoch's effect; removing a node requires a
   successor RFC with an explicit refund** / freeze effects at purchase time.
8. **OD-8 Respec.** **None in v1** / a refund intent (new mechanic; needs design).
9. **OD-9 Leaderboards.** **No new board variable; tree state is Founder progression, like the
   curriculum starter** / add a "tree-assisted" variable.
10. **OD-10 Proposed literals.** **Ratify the R2 table only after H2/H4/H5 measurement**, with
    Codex returning the tuple plus evidence (the epoch-8 payoff protocol).
11. **OD-11 Copy.** **Owner authors all prose for the R9 key list**; Claude may draft candidates
    only on request, for explicit adoption.
12. **OD-12 Bonus cap.** **No hardcap** (none is in the design; a cap would mint a currency per
    `§2c`) / a visible cap with a reason key.

## Open questions

- Should the Offer Sheet's plan panel show the offer's preview delta or the always-higher
  commit-time payout as the budget? Recommended: the preview (a promise), noting the server may pay
  more.
- The successor RFCs for DG-4 (Network) and DG-5 (offline extension) are the next Reputation-tree
  work; both depend on this RFC's accounting and artifact.

## Changelog

- 2026-09-25: created (draft — not implementation authority).
