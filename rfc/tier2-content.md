# RFC: Tier 2 Playable Content (T0–T2 completion)

- **Status:** accepted — owner batch acceptance 2026-09-25; implementing
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-09-25
- **Design refs:** `design/01 §Tier 2` (IT Company: headcount allocation, procurement, startup
  culture, FarmVille era, split `Incorporation`), `design/02 §2.1–2.2, §2c, §3.1, §11, §11b`
  (cost curves, production stack, cap-mints-currency law, prestige formula, pacing targets,
  tier-relevance doctrine), `design/07 §Phase 1` ("Tiers 0–2 complete: cost-curve economy,
  headcount allocation"), `design/10 §1, §3b, §4` (incorporation at Tier 2, doctrine per
  transition, `No Hires`), `design/08 §1–2` (voice rules; era-2 UI language),
  `design/11 §2, §5–6` (one hint per new system, copy keys, owner editorial cut),
  `design/03 §Tier table` (T2 minigames; out of scope here)
- **Depends on:** T0–T1 Playable Content (archived; epoch 8 `First-Hour Payoff` is the
  byte-preserved baseline), Purchasable Content Foundation and Relevance Harness (archived),
  Faction & Incorporation (archived; `incorporate` is Tier ≥ 2), Doctrine & Compute Credit
  (archived), Game-UI Screens and UI Foundation (archived), Copy Pipeline Foundation (archived),
  Prestige & Exits and Leaderboards & Balance Epochs (implementing; own the Exit path and the
  epoch mint protocol)
- **Parent / amends:** follow-up to archived `rfc/archive/t0-t1-playable-content.md`. It changes
  none of that RFC's text and does not edit epoch 7/8 bytes.
- **Supersedes / superseded by:** —
- **Planning:** `planning/tier2-content/` (once implementing)

## Summary

This RFC makes Tier 2 (IT Company, 2010s) a reachable, playable, measured tier. It adds the
missing `gate.t1_to_t2`, three Tier-2 generators, three Tier-2 upgrades, one synergy pool, and
Tier 2's defining grammar: **headcount allocation**. A finite seat budget is assigned across
Engineers, Sales, Support, and Middle Managers. The Middle Manager converts Engineers into Slide
Decks at 1:1. The RFC also adds the `era_2010` presentation hooks, a Tier-2 pacing scenario, and
a combined T1–T2 relevance scenario. All of it ships as one owner-gated epoch: **epoch 9, "IT
Company"**. Every literal here is a PROPOSED, retunable starting value. Exact bytes go through
the established draft → review → owner-ratify-by-SHA lane (the T01-C2 precedent) before any mint.

## Motivation

**Tier 2 is structurally unreachable today.** Routes declare `gate.t0_to_t1`, `gate.t2_to_t3`,
`gate.t3_to_t4`, and `gate.t4_to_t5`, and nothing else. `tierForGate` derives the tier from the
adjacent gate id, so crossing `gate.t2_to_t3` moves a Tier-1 company straight to tier 3. No path
produces `Tier == 2`. The consequences:

- `incorporate` (Tier ≥ 2, `server/production/intents.go`) is reachable only by skipping to tier 3;
- the Game UI throws `RangeError` for any tier other than 0 or 1
  (`client/src/game-ui/contracts.ts`, `RunEndSurface.svelte`);
- the `T-002.*` capability rows in `planning/platform-alignment/capability-outcome-ledger.tsv`
  record headcount allocation as `absent`.

Phase 1 of `design/07` cannot close without this tier.

**In scope:**

- (A) the T1→T2 gate and Tier-2 reachability;
- (B) Tier-2 catalog content on shipped mechanics;
- (H) the headcount allocation mechanic;
- (E) `era_2010` presentation hooks, including the Tier-2 gate and a minimal incorporation control;
- (P) harness pacing and relevance evidence;
- (M) the epoch-9 mint.

**Out of scope:** each item names its owner.

| Out of scope | Owner |
|---|---|
| Procurement automation of generator purchasing (`design/01` "Automates") | DESIGN-GAP HG-6 → successor mechanic RFC |
| Morale, crunch events, the Soul drains they cause, and VC term-sheet clause 7(c) | Events Engine Layer 1 (draft) + Soul debit sources (HG-5) |
| Layoffs event at record profit | Events Engine Layer 1 |
| Server Garden and board-game suite unlock at T2 | their minigame RFCs (Fiscal-purchased unlock per `design/03`) |
| Garage→IT Co doctrine (Dev Tools / Consumer App / Government Contractor) | HG-7 |
| Reputation payout at T2 Exits | HG-8 → Reputation tree v1 (`design/07` Phase 1 item) |
| Diegetic term-sheet presentation of the faction choice | successor UI RFC (this RFC ships only a minimal control) |
| Tiers ≥ 3 UI eras | later tier RFCs; tier ≥ 3 stays fail-closed in the UI |

## Specification

### A — The T1→T2 gate and Tier-2 reachability

A1. Add a new row to the Routes artifact `balance/routes/phase0.json`. It keeps schema v1 and
is edited in place, following the epoch-7 convention:

```json
{ "gate_id": "gate.t1_to_t2", "requirement": [{ "resource_id": "company.cash", "amount": "1e7" }], "routes": [] }
```

`1e7` is PROVISIONAL. The ratified literal comes from measurement against the P2 envelope (see
owner decision 7). The gate set is inserted in raw-byte order. The existing `cross_gate`
transition then yields `Tier = 2` through `setTierFromGate`. No engine change is needed for the
crossing itself.

A2. Update the Categories artifact `balance/categories/phase0.json`: `full_gate_set` gains
`gate.t1_to_t2`. The loader already requires set equality with the Routes gates, so omitting
it fails the load. `hundred_percent`'s `all_gates` then requires the new gate.

A3. Existing sequencing is unchanged. Out-of-order legal crossings remain monotonic
(`docs/prestige-and-exits.md`). A company that banks `1e9` at Tier 1 may still cross
`gate.t2_to_t3` and skip Tier 2. That is a sequence break, and it is recorded under Deviations.

A4. No doctrine row is added for `transition.t1_to_t2` (HG-7). `cross_gate` therefore does not
require `pick_doctrine` at this boundary.

A5. **The new gate is also an offer site.** Exit offers are drawn at `cross_gate`
(`afterPrestigeTransitionResolved`) once Founder Exit history is non-empty. The draw uses the
ratified `spawn_gate_ppm[2]` (400,000). Crossing `gate.t1_to_t2` is therefore the first Tier-2
offer opportunity. This matches `design/01` T2: "first prestige available shortly after". Offer
arithmetic is unchanged. What an accepted T2 offer pays is HG-8's question.

### B — Tier-2 catalog content (economy schema v5)

B0. **Price band (DESIGN-GAP HG-9).** `design/02 §2.1` asks for a ×12 cost / ×6.5 output
cadence. Continuing it from `generator.beige_tower_v2` (3.58e8) would put Tier-2 generators at
≥ 4.3e9. That is above the ratified `gate.t2_to_t3` requirement of `1e9`, so no Tier-2
generator could be bought inside its own tier. The shipped catalog already interleaves tiers:
`generator.legal_dept` (tier 3) costs 1e8. Tier-2 content therefore occupies **[1e7, 1e9)**,
interleaved with the upper Tier-1 band. Tier 2's growth engine is headcount (H), not a longer
generator ladder.

B1. **Generators.** Add these rows to `balance/catalogs/phase0.json`, edited in place. All
values are PROPOSED retunable balance data:

```json
{ "id": "generator.open_plan_floor", "tier": 2, "category": "category.office",
  "price": { "resource_id": "company.cash", "base": "1.2e7", "curve": { "kind": "geometric", "ratio": "1.13e0" } },
  "production": { "resource_id": "company.cash", "base_rate": "2e4" },
  "ladder": [ { "purchased_at": 25, "multiplier_ppm": 2000000 }, { "purchased_at": 50, "multiplier_ppm": 2000000 }, { "purchased_at": 100, "multiplier_ppm": 2000000 } ],
  "roles": [ { "kind": "headcount_seats", "seats_per_purchased": 5 }, { "kind": "synergy_feed", "pool_id": "pool.org_chart" } ] },
{ "id": "generator.managed_services_contract", "tier": 2, "category": "category.software",
  "price": { "resource_id": "company.cash", "base": "6e7", "curve": { "kind": "geometric", "ratio": "1.11e0" } },
  "production": { "resource_id": "company.cash", "base_rate": "1.3e5" },
  "provisions": { "generator_id": "generator.garage_rack", "rate_ppm": 100000 },
  "ladder": [ { "purchased_at": 20, "multiplier_ppm": 2000000 }, { "purchased_at": 50, "multiplier_ppm": 2000000 }, { "purchased_at": 100, "multiplier_ppm": 2000000 } ],
  "roles": [ { "kind": "provision", "generator_id": "generator.garage_rack" }, { "kind": "synergy_feed", "pool_id": "pool.org_chart" } ] },
{ "id": "generator.hot_desk_program", "tier": 2, "category": "category.office",
  "price": { "resource_id": "company.cash", "base": "1.44e8", "curve": { "kind": "geometric", "ratio": "1.12e0" } },
  "production": { "resource_id": "company.cash", "base_rate": "2.5e5" },
  "ladder": [ { "purchased_at": 25, "multiplier_ppm": 2000000 }, { "purchased_at": 55, "multiplier_ppm": 2000000 }, { "purchased_at": 100, "multiplier_ppm": 2000000 } ],
  "roles": [ { "kind": "headcount_seats", "seats_per_purchased": 12 }, { "kind": "synergy_feed", "pool_id": "pool.org_chart" } ] }
```

The provisioning edge follows the `§11b` kinetic chain: tier 2 provisions tier 1. It requires
one behavior-neutral addition to the existing `generator.garage_rack` row:
`"provisioned_hardcap": { "count": 9007199254740991, "reason_key": "generator.garage_rack.provisioned_cap" }`.
Without a provisioner that cap has no effect, so Tier-1 trajectories are unchanged.

B2. **Upgrades.** Every upgrade uses window `{ "from_gate": "gate.t1_to_t2", "to_gate":
"gate.t2_to_t3" }`. `cost.amount` equals `requires[0].value` exactly, a pairing that changes
atomically (the T01-C24 rule). Each upgrade has `roles: ["synergy_feed"]` and a `copy_key`
equal to its id.

| id | cost | effect (`upgrades` slot, `factor: "2e0"`) |
|---|---|---|
| `upgrade.ping_pong_table` | `3e7` | `generator.open_plan_floor` |
| `upgrade.move_fast_break_things` | `1.5e8` | `generator.managed_services_contract` |
| `upgrade.nap_pod` | `3.6e8` | `generator.hot_desk_program` |

Each effect's `source_id` is `<upgrade id>.factor`, as in the T0–T1 rows. The names come from
the startup-culture seed bank. They are mechanical ids here, and their player-facing text is
owner-authored (E4). The research bank's side effects (incident rate, burnout, morale) are not
implemented (HG-5).

B3. **Synergy pool.** `pool.org_chart` uses `"slot": "upgrades"` and `"curve": "log"`. Its
sources are `generator.open_plan_floor` 3000, `generator.managed_services_contract` 3000,
`generator.hot_desk_program` 4000, and each B2 upgrade at 30000 `per_count_ppm`. B3 also adds
the `multiplier_sources` declaration
`{ "id": "pool.org_chart", "slot": "upgrades", "target": "all", "provider": "pool.org_chart" }`.

B4. **Unchanged:** every T0/T1 row other than the B1 hardcap addition, the progress coordinates,
the manual policy, the offline policy, and the curriculum. The prestige, factions, doctrines,
and curriculum artifacts are not edited (HG-7, HG-8).

B5. **The cap-mints-a-currency law (`design/02 §2c` law 2).** The headcount budget hardcap (H1)
mints **seats** as the parallel currency, because seat-bearing generators become the scarce
good. The catalog names it (`headcount.budget.cap`), and its presentation copy must satirize it
(hot-desking is the in-fiction answer to seat scarcity). The epoch-9 changelog publishes the
faucet/sink budget (law 1):

- **Sinks:** the gate burns `1e7`; purchases.
- **Faucets:** the three generators, the `pool.org_chart` factor, and the headcount role
  factors, each with its formula.

### H — Headcount allocation (DESIGN-GAPs HG-1–HG-4; recommended arms written normatively)

`design/01` fixes these facts:

- there is a finite hiring budget;
- there are four roles: Engineers, Sales, Support, and Middle Managers;
- Middle Managers convert Engineers into Slide Decks at 1:1;
- the player's decision is allocation ratios and re-allocation timing;
- `design/10`'s `No Hires` must be able to delete the system.

`design/01` does not fix the budget source, the reallocation cost, the Sales and Support effects,
or the use of Slide Decks. The text below writes each recommended arm normatively, so accepting
the defaults makes it implementable. Each open item is recorded as an HG gap with its
alternatives.

H1. **Catalog grammar (economy schema v5; v1–v4 still load byte-for-byte).**

- New generator role kind `{ "kind": "headcount_seats", "seats_per_purchased": <positive safe integer> }`.
  This extends the closed role vocabulary, which `§11b` says may grow only by RFC.
- New top-level block:

```json
"headcount": {
  "unlock_tier": 2,
  "budget_hardcap": { "count": 1000000, "reason_key": "headcount.budget.cap" },
  "roles": [
    { "id": "role.engineer", "copy_key": "role.engineer", "converts": null,
      "effects": [
        { "source_id": "headcount.role.engineer.open_plan_floor", "target": "generator.open_plan_floor", "per_effective_ppm": 50000 },
        { "source_id": "headcount.role.engineer.managed_services_contract", "target": "generator.managed_services_contract", "per_effective_ppm": 50000 },
        { "source_id": "headcount.role.engineer.hot_desk_program", "target": "generator.hot_desk_program", "per_effective_ppm": 50000 } ] },
    { "id": "role.middle_manager", "copy_key": "role.middle_manager",
      "converts": { "from_role": "role.engineer", "units_per_assigned": 1 }, "effects": [] },
    { "id": "role.sales", "copy_key": "role.sales", "converts": null,
      "effects": [ { "source_id": "headcount.role.sales.all", "target": "all", "per_effective_ppm": 10000 } ] },
    { "id": "role.support", "copy_key": "role.support", "converts": null,
      "effects": [
        { "source_id": "headcount.role.support.click", "target": "manual.click", "per_effective_ppm": 100000 },
        { "source_id": "headcount.role.support.answering_machine", "target": "generator.answering_machine", "per_effective_ppm": 100000 } ] }
  ]
}
```

- Every effect `source_id` has a `multiplier_sources` row with `"slot": "upgrades"` and
  `"provider": "headcount"`.
- The loaders reject a catalog, in both runtimes and in the JSON Schema, when any of these holds:
  - a `headcount` block exists with no `headcount_seats` role, or a seats role exists with no
    block;
  - a role id is duplicated or its ids are not raw-byte sorted;
  - an effect is undeclared or has a dangling target;
  - `converts.from_role` is unknown, is itself a converting role, or has no effects;
  - more than one role converts;
  - `units_per_assigned` or `per_effective_ppm` is not a positive safe integer;
  - `unlock_tier` falls outside `[0,3]`, the loader's tier domain.

H2. **Derived quantities.** These are exact integers, computed identically in Go and TypeScript
and never persisted:

```text
budget            = min(budget_hardcap.count, Σ_g purchased(g) × seats_per_purchased(g))   // purchased only; provisioned units carry no seats
converted         = min(assigned[converting], assigned[from_role] / units_per_assigned)    // integer; v1 literal = 1
effective[from]   = assigned[from] − converted × units_per_assigned
effective[r]      = assigned[r]                  (every other role; the converting role has no effects)
slide_decks       = converted                    // displayed; produces nothing (HG-4)
factor(effect)    = 1 + effective[role] × per_effective_ppm / 1,000,000   // Decimal, quantized once at emission
```

The budget is computed with checked big-integer arithmetic before the hardcap. Overflow saturates
to the visible cap and never wraps. The purchased-only rule follows the `§11b`
purchased/generated split: generated units are free and priceless, so they grant no seats.

H3. **State.** The next free Company save version on the save layer's chain adds a complete-key
`headcount_assignments` map over the pinned catalog's role ids. The save layer assigns the
version number; this RFC does not pre-assign it.

- It activates only for a new run pinned to an epoch whose economy declares `headcount`. This is
  the new-run-bound activation used by the Soul and Doctrine foundations. A run pinned to epoch
  8 keeps its schema and finishes under its pins.
- Invariant: `Σ assigned ≤ budget`. Within a run, purchased counts never decrease, so the budget
  is monotone. Any violation is `ErrInvalidEngineState` at decode, apply, and replay.
- New-run assembly writes the all-zero complete map. Headcount is run-scoped.

H4. **Intent.**

```json
{ "kind": "set_headcount_allocation", "intent_id": "<uuidv7>", "expected_revision": 7,
  "assignments": { "role.engineer": 6, "role.middle_manager": 0, "role.sales": 3, "role.support": 1 } }
```

- `assignments` has exactly the pinned role ids, each a non-negative safe integer. The command
  goes through `ApplyLogged` like every other Company command.
- Order of operations:
  1. preflight, including the shipped lazy offline catchup (T01-C15);
  2. evaluate elapsed production under the **previous** allocation;
  3. run the accrual hook;
  4. validate;
  5. commit the new map;
  6. emit one `headcount_allocated` v1 event on the new revision.

  The allocation is piecewise-constant between commands, so evaluation stays closed-form. It can
  never apply retroactively to time that has already elapsed.
- Terminal rejections, using existing categories:

  | Condition | Rejection |
  |---|---|
  | Tier below `unlock_tier` | `not_eligible/tier` |
  | Map equal to the current one | `not_eligible/unchanged` |
  | Missing or extra key, negative, non-integer, or unsafe value | `invalid` |
  | `Σ > budget` | `cap_exceeded/headcount.budget.cap` |
  | Catalog has no `headcount` block | `unknown_id` |

- Reallocation is free, instant, and unlimited (HG-2). Timing matters because the budget grows
  with seat purchases and each role's marginal value shifts with the generator mix.
- `headcount_allocated` v1 payload, strict exact-object:
  `{founder_id, run_id:{company_stream_id,run_seq}, budget, previous:{…}, assignments:{…}}`.
  Replay-inputs gain no field, because every input is Company state or command payload. The event
  joins the event registry and the client exact-object decoder.

H5. **Production and projection.**

- Headcount contributions are assembled by the same private functions used for live evaluation,
  replay, the pure read projection, and simulation. They are ordered by `source_id` within the
  `upgrades` slot under the existing rule.
- The authoritative receipt snapshot adds
  `headcount: {budget, budget_hardcap:{count,reason_key}, assignments, effective, slide_decks}`,
  or `null` when the catalog has no block.
- `docs/generated/production-formulas.json` publishes H2 (law 9). Regenerating it is its own
  reviewed commit, and `make formulas-check` must fail on the tree before that commit.

H6. **Simulation-only masking** (for the relevance and role gates).

- A generator effect mask also nulls that generator's `headcount_seats`, which lowers the budget.
- Only under a simulation mask, `effective` is computed from assignments clamped greedily:
  `min(assigned[r], remaining_budget)` in raw-byte role order, so a masked budget can never fall
  below the assignment total.
- The authoritative and replay entrypoints still have no mask parameter. The existing source
  guard extends to cover this.

### E — Era presentation hooks (Tier 2 = `era_2010`)

E1. **Theme.** Add `ui/themes/era_2010.json`: the same 41 tokens and the same schema as
`era_2000`, in the `design/08 §2` "2010 flat startup" direction. Token values are
owner-reviewed design data. The following gain `era_2010`:

- `UI_ERAS` and `CopyEra`;
- `eraForSnapshot` (tier 2 → `era_2010`);
- `RunEndSurface` (tier 2 → `era_2010`).

Tier ≥ 3 remains a thrown `RangeError`. That fail-closed behavior is kept deliberately and is
tested.

E2. **Snapshot v4** (additive; v1–v3 receipts stay replayable under their minted version):

- `transitions` projects the `gate.t1_to_t2` control. It is derived, like the first Gate, by
  invoking the production transition on a discarded clone.
- It projects an `incorporate` control listing the pinned faction catalog's rows. Their existing
  `incorporation_copy_key`s supply the labels, and the control is enabled only at Tier ≥ 2 with
  no faction.
- It carries the H5 `headcount` block.
- Any other gate or route stays fail-closed.

E3. **Desk surfaces.**

- A **Headcount** panel shows the budget with its cap reason, per-role steppers, and Unassigned
  and Slide Decks counters. It submits one full-map `set_headcount_allocation`, and the receipt
  is authority.
- The existing Gate control is extended to the projected `gate.t1_to_t2`.
- The minimal **Incorporate** control submits `incorporate {faction_id}`.
- Per `design/11 §2`, each new system gets exactly one contextual hint key and one "why this
  matters" tooltip key.
- The FarmVille beat is a **non-stateful energy-bar chrome stub** under the Horse Armor
  precedent: always full, no intent, no state, with its curtain tooltip key. Its refill
  affordance is presentational and emits nothing.

E4. **Copy keys, all prose owner-authored/pending.**

- Families:
  - `generator.<id>.{title,description}` (presentation schema v3 binding for the three
    generators);
  - `upgrade.<id>.{title,description}`;
  - `role.<id>.{title,description}`;
  - `headcount.budget.cap`;
  - `headcount.slide_decks.curtain`;
  - `headcount.panel.hint`;
  - `generator.garage_rack.provisioned_cap`;
  - `split.gate.t1_to_t2` (split name `Incorporation`, per `design/01`);
  - `incorporate.panel.hint`;
  - `era_2010.energy_bar.{label,curtain}`;
  - `era_2010` variants for every existing Desk/Run-End key rendered at tier 2.
- The copy-bearing JSON paths (economy `headcount.roles[].copy_key`, the new reason keys) are
  registered in `copy/references.v1.json`.
- This RFC contains **no final copy**. Text comes from the owner's editorial round (the FCE-C7
  pattern) under `design/08 §1`: deadpan, curtain always pulled, never punching at
  employees-as-people. The Middle Manager and "We're a family"-register lines target the HR
  *system*.
- Any statistic from `research/societal-satire.md §1`, for example unlimited-PTO days taken,
  ships only with a verified provenance id. Otherwise it is flagged and cut.

### P — Harness evidence

P1. **Policy registry** `balance/testdata/t2/it-company-policy-v1.json` (schema v2 of the
`first_hour_policy` grammar; v1 bytes unchanged). It copies `seed_derivation` byte-for-byte and
extends it as follows:

- `enumeration_order.classes` becomes
  `[perform_manual_batch, buy_generator, buy_upgrade, set_headcount_allocation, wait]`.
- Allocation arm keys are `move:<from>><to>`:
  - `from` ∈ {`unassigned`} ∪ {roles with count > 0};
  - `to` ∈ roles;
  - `to ≠ from`;
  - raw-byte ascending.
- The three policies:

| Policy | Rules |
|---|---|
| `chaos.t2` v1 | The `chaos.t0_t1` literals, plus allocation moves in the seeded uniform draw |
| `casual.t2` v1 | The `casual.t0_t1` literals, plus one rule: whenever `unassigned > 0` at an action boundary, the action is one allocation assigning all unassigned seats to the role whose effects include target `all` with the highest `per_effective_ppm` (ties raw-byte). It never reallocates. |
| `reference.greedy` v2 | The T01-C20 projected-time ranker, preceded at every decision boundary by an allocation hill-climb: single-unit moves, strictly improving projected time only, `move:` key tie-break, run to a fixed point. Guard: at most `budget` moves per boundary; exhausting it fails loud as `allocation_starved`. Target sequence (T01-C39): `gate.t0_to_t1` requirement → `gate.t1_to_t2` requirement → `1e9` cash. It banks after the last target. |

- All three use `gate_crossing: first_boundary_requirements_met` over
  {`gate.t0_to_t1`, `gate.t1_to_t2`}.
- All three use `exit_rule: t01_c32_readiness_once`: the shipped T01-C32 predicate, taken only
  while Founder Exit history has exactly one entry. Without it, the shipped rule Exits every
  later run at the garage gate, and no policy could ever reach Tier 2.
- No policy incorporates (owner decision 10).
- Every policy sets `offer_rule: ignore`. It never submits `accept_exit_offer` or
  `decline_exit_offer`, so a spawned offer lapses through the shipped expiry at a later
  evaluation. Accepting would end the run before the milestone, and the offer timing is already
  owned by the ratified prestige policy.

P2. **Pacing scenario** `balance/testdata/t2/harness-scenario-v1.json`, scenario schema v2.

- The one grammar change: a `gate_crossed` milestone may set `"run_seq": null`, meaning the first
  crossing in any run.
- Runs:
  - `chaos.t2`: 64 seeds, horizon 21,600,000 ms;
  - `casual.t2`: 32 seeds, horizon 43,200,000 ms wall;
  - `reference.greedy` v2: 1 seed, horizon 21,600,000 ms.
- Milestones:
  - `milestone.it_company_gate` — `gate_crossed`, `gate.t1_to_t2`, `run_seq: null`,
    `founder_attended_ms`, `must_reach`;
  - `milestone.first_headcount_allocation` — `intent_applied`, `set_headcount_allocation`,
    `founder_attended_ms`, `must_reach`. It is an unbounded drift observation with no envelope.
- **Envelopes** (`design/02 §11`: "First minigame ~2–3 hours (Tier 2)"): the Chaos p50 and the
  Casual p50 of `milestone.it_company_gate` both lie in [7,200,000, 10,800,000] ms.
- Invariants: the nine T0–T1 invariants plus `headcount_budget_respected` (`Σ assigned ≤ budget`
  at every revision).
- `transition_budget` = measured worst-case work × 2, rounded up to a round literal (T01-C17
  branch B). The measurement is recorded in the planning log before ratification, never set
  blind.

P3. **Relevance.**

- **T0 scenario:** re-run on the epoch-9 bundle. T2 rows are unaffordable below its `1e5`
  target, so its report must be **byte-identical to the epoch-8 golden except identity fields**.
- **T1-only scenario:** retired for epoch 9 (its golden stays as history). It is replaced by
  `scenario.t1_t2_relevance`, which targets `1e9` cash with three segments:
  `[null→gate.t0_to_t1]`, `[gate.t0_to_t1→gate.t1_to_t2]`, `[gate.t1_to_t2→gate.t2_to_t3]`.
  T1 windows close at `gate.t2_to_t3`, so T1 and T2 rows are judged in the same scenario. That is
  the honest test of whether Tier-2 content kills Tier-1 content.
- **Policy artifact:** gains the six T2 rows (three generators, three upgrades) at `epsilon_ms`
  1000, with no trap exemptions.
- **Reachability** is proven by measurement and loader-checked (T01-C18).
- **Budgets** follow T01-C17.

P4. **Role gate** (`make t2-role-check`, the analog of `t0-t1-role-check`). Each T2
generator-role row runs its minimum honest context, paired with a masked control:

- `headcount_seats`: one seat source, all seats assigned to `role.sales`; the mask removes the
  budget, and the H6 clamp neutralizes the factor;
- `provision`: advance across the grid;
- `synergy_feed`: exercised pool target.

P5. **Allocation-decision gate.** This is the relevance analog for roles, which are not
purchasables. On the P3 combined scenario's reference trajectory, `role.engineer` and
`role.sales` must each hold a nonzero effective allocation at some decision boundary. That
proves the allocation decision is not dominated by a single role. `role.support` and
`role.middle_manager` shares are reported as observations only (HG-3, HG-4).

### M — The epoch-9 mint

M1. **Artifacts edited in place** (the epoch-7 convention):

- economy (v4 → v5);
- routes (A1);
- categories (A2);
- relevance (P3).

Every other artifact in the current nineteen-artifact bundle is carried byte-for-byte.
`balance/epochs/phase0.json` appends epoch 9 `IT Company` with `changelog/epoch-9.md`. Earlier
epochs are immutable; their identities are enforced by the existing epoch-history guards.

M2. **Commit protocol**, per `docs/balance-harness.md`:

1. inputs (artifacts, scenarios, policy registry) land first;
2. the formulas artifact regenerates in its own reviewed commit;
3. each generated pacing, relevance, and role golden lands in an isolated `BALANCE-CHANGE:`
   commit containing only generated artifacts;
4. `deployment/content-manifest.v1.json` regenerates through `copy-generate`.

A combined commit fails the guard. Published history is never rewritten.

M3. **Pre-mint ratification** (owner, by SHA): the economy, routes, categories, relevance
policy, both scenarios, the P1 registry, the presentation binding, and `era_2010.json`. After
ratification, any edit records a replacement hash in this RFC.

M4. **Activation.**

- Epoch-8 runs finish under their pins.
- An Exit or fresh genesis after the mint receives epoch 9 and the new Company save version (H3).
- Tier 2 renders `era_2010` only from the authoritative tier fact.

### Implementation sequence (for `planning/tier2-content/plan.md`)

1. **H1/H2 loaders.** Go, TypeScript, and JSON Schema, plus the mutation corpus.
2. **H3–H5 runtime.** State, intent, event, production and projection, Go-authored replay corpus
   consumed by TypeScript.
3. **Formulas regeneration commit.**
4. **Harness grammar.** P1 registry schema v2, scenario `run_seq: null`, the reference
   allocation hill-climb, the H6 masks, and P4/P5 gates, each with its negative control.
5. **Candidate documents.** Codex drafts the literal candidates from §A/§B/§H; Claude reviews.
6. **Calibration measurement.** It sets the gate literal and budgets, then the owner ratifies by
   SHA.
7. **UI.** E1–E3.
8. **Owner copy round (E4).**
9. **Mint (M1–M4).** Includes the goldens and the composed Postgres proof.
10. **Docs.** Listed under "Done" below.

## Deviations from design

1. **Procurement automation is not shipped.** `design/01` T2 "Automates: generator purchasing"
   is deferred (HG-6). Auto-purchase inside lazy evaluation breaks closed-form segment
   integration and needs its own replay contract. T0–T1 likewise shipped no autoclicker.
2. **Morale, crunch, Soul drains, clause 7(c), and the layoffs event are deferred** (HG-5).
   Tier 2 introduces no Soul debit source.
3. **No Garage→IT Co doctrine** (`design/10 §3b` requires one per transition) (HG-7).
4. **T2 Exits pay no Reputation.** The ratified threshold is `1e12`, and T2 lifetime value sits
   around 1e9–1e10, so `⌊∛(value/1e12)⌋ = 0`. The payoff is Route Knowledge until Reputation
   tree v1 (HG-8). `design/01` says the first T2 prestige "must feel great". Also retained:
   Exit is already available from Tier 1, which is a T0–T1 deviation.
5. **Cost cadence break.** Tier-2 generators interleave with the upper Tier-1 band in
   [1e7, 1e9) instead of continuing ×12/×6.5 (HG-9).
6. **Tier 2 is skippable** by crossing `gate.t2_to_t3` directly (A3). This is the existing
   monotonic-gate rule, kept as a sequence break.
7. **Headcount specifics are proposals where the design is silent:** the seat budget source,
   free reallocation, the Sales, Support, and Engineer targets, and Slide Decks producing nothing
   (HG-1–HG-4). Middle Manager is a role, as in `design/01`, not the research bank's "Building".
8. **The FarmVille beat is a non-stateful chrome stub.** No energy or timer mechanic exists.
9. **The T2 minigames are not included.** They belong to their own RFCs.

## DESIGN-GAPs

| Gap | Question | Options | Recommended |
|---|---|---|---|
| HG-1 | Budget source | (a) seats from T2 generators' `headcount_seats` role (Kittens housing); (b) each hire is its own purchasable per-role generator (no allocation, just more generators); (c) budget derived from Reputation/Network | (a): the only option that makes allocation a *finite-budget* decision on shipped cost curves |
| HG-2 | Reallocation cost/timing | (a) free, instant, unlimited; (b) cooldown per move; (c) onboarding ramp (new effective counts phase in) | (a): timing emerges from budget growth; (b)/(c) add hidden clocks with no design text |
| HG-3 | Role effects | Engineer: T2 generators / all generators; Sales: `all` / cash-only; Support: `manual.click` + answering machine / a Trust meter (unshipped) | H1 literals; Support's weakness is reported (P5), not hidden |
| HG-4 | What Slide Decks do | (a) nothing: curtain-pulled satire trap, reported; (b) an input to `gate.t2_to_t3` (edits ratified T3 content); (c) exit-offer "investor interest" (edits the prestige offer model) | (a) now; (b)/(c) belong to a later RFC with their owners |
| HG-5 | Morale, crunch → Soul, clause 7(c) | own this RFC / Events Layer 1 + Soul debit rows | defer; an events-engine consumer is required anyway |
| HG-6 | Procurement auto-buy | own this RFC / successor mechanic RFC | successor RFC |
| HG-7 | T1→T2 doctrine | ship a choice-only row now / defer until doctrine effects exist | defer: an effectless choice would be hollow |
| HG-8 | Reputation at T2 Exits | lower the threshold now / defer to Reputation tree v1 | defer: Reputation has no consumer yet, so a nonzero number would do nothing |
| HG-9 | T2 band | [1e7,1e9) interleave / move ratified `gate.t2_to_t3` (cascades into Permits/T3) | interleave |

## Acceptance criteria

Each criterion ships with its demonstrated failing case, which is recorded in the planning log
before the green run is claimed. All gate claims run with `-count=1`.

0. **Reachability.** Under the epoch-9 candidate bundle, a Tier-1 company crosses `gate.t1_to_t2`
   and gets `Tier == 2`, and `incorporate` then applies.
   *Failing case:* the same test under the epoch-8 bundle rejects `unknown_id` and
   `not_eligible/tier`.
1. **Strict loaders.** Economy v5, Routes, and Categories load identically in Go and TypeScript.
   Every H1 rejection, plus a categories set missing `gate.t1_to_t2`, fails in both runtimes and
   in the JSON Schema. v1–v4 catalogs re-load byte-identically.
   *Failing case:* each mutation fixture must be observed red.
2. **Headcount transition parity.** A Go-authored sequential corpus is replayed byte-for-byte by
   TypeScript. It covers every H4 outcome, conversion with E=3/M=5 (effective 0, decks 3),
   hardcap saturation, and the Exit reset.
   *Failing case:* a mutant that omits the conversion subtraction breaks the vectors.
3. **Closed-form partition invariance.** With an allocation change at t1, evaluation over
   [t0,t2] equals the sum of the segment evaluations, as a property test.
   *Failing case:* a mutant that applies the new allocation to the pre-command span fails.
4. **Role gate (P4) green for every T2 role row.**
   *Failing case:* the control with its mask removed still shows activation, and the gate must
   go red.
5. **Relevance.** `scenario.t1_t2_relevance` has zero dead purchasables among the T1 and T2 rows.
   The T0 scenario is byte-identical to epoch 8 except identity fields. The P5 gate is green.
   *Failing cases:* the registry's dead-upgrade fixture fires; a P5 fixture with Sales ppm set to
   0 fails; a deliberately perturbed T0 row breaks the identity comparison.
6. **Pacing.** The P2 envelopes are green, must-reach holds at every seed, and
   `headcount_budget_respected` holds. The first-hour scenario, re-run on epoch 9, keeps all
   seven of its ratified envelopes and the same-seed relation green, with drift reported under
   the 10%/25% rule.
   *Failing case:* a candidate with the gate at `1e5` must violate the lower envelope bound.
7. **Formulas and replay.** The formulas artifact regenerates in its own commit, and
   `formulas-check` is red on the parent. Replay verification passes for all new commands and
   events.
8. **UI.** Tier 2 renders `era_2010`. The Gate, Incorporate, and Headcount controls operate
   through intents in the composed browser path. Tier ≥ 3 still throws. axe reports no serious or
   critical WCAG 2.2 AA violations in Chromium, Firefox, and WebKit.
   *Failing case:* removing the tier-2 mapping reproduces the `RangeError`.
9. **Copy.** `make copy-check` is green with no orphan keys, no known-red names, and no
   unverified statistic. Every E4 key is owner-adopted.
   *Failing case:* one deleted key fails the check.
10. **Mint and composed proof.** Epoch 9 is appended with the M2 commit protocol green. A
    real-Postgres composed test replays the pinned-seed `casual.t2` path. It reaches
    `gate.t1_to_t2`, allocates headcount, verifies durable replay, and shows an epoch-8 run
    finishing under its pin.
    *Failing case:* bypassing the composed catalog must make the test fail.

**Done** means all of the following, in the final change:

- AC0–AC10 are green;
- cross-party designated review has cited the full implementation range;
- `docs/` is updated:
  - new `docs/tier2-content.md`;
  - `purchasable-content.md`, `economy-kernel.md`, `production-engine.md`, `save-layer.md`,
    `routes.md`, `balance-harness.md`, `game-ui.md`, `ui-foundation.md`, and
    `leaderboards-and-epochs.md`;
- the RFC and planning directories are archived.

## Owner decisions needed

1. **HG-1 budget source.** Recommended: seats from `headcount_seats` on T2 generators.
2. **HG-2 reallocation.** Recommended: free, instant, unlimited.
3. **HG-3 role effects.** Recommended: the H1 literals (Engineer +5% per effective engineer on
   T2 generators; Sales +1% per head on `all`; Support +10% on `manual.click` and the answering
   machine).
4. **HG-4 Slide Decks.** Recommended: they produce nothing and the curtain is pulled; the Middle
   Manager is the reported satire trap.
5. **HG-5/HG-6/HG-7 deferrals** (morale/crunch/Soul, procurement, T1→T2 doctrine). Recommended:
   defer all three, with the named owners above.
6. **HG-8 Reputation at T2 Exits.** Recommended: keep threshold `1e12` and defer to Reputation
   tree v1.
7. **HG-9 band and gate literal.** Recommended: T2 content in [1e7, 1e9) and `gate.t1_to_t2`
   provisionally `1e7`. The ratified literal is the measured value meeting the [2h, 3h] envelope.
8. **Pacing envelope reading of "~2–3 hours".** Recommended: Chaos p50 **and** Casual p50 in
   [2h, 3h] of Founder-attended time.
9. **One epoch or two.** Recommended: a single epoch 9. Headcount is T2's defining grammar, and
   pacing measured without it would need a second calibration round.
10. **Faction choice in T2.** Recommended: ship the minimal Incorporate control and defer the
    term-sheet presentation. The harness policies do not incorporate; faction balance is measured
    per faction by its own RFC.

## Open questions

- The Tier-1 progress coordinate (composite, cash target `1e6`) reaches 100% before a `1e7` gate.
  Recommended: leave it unchanged this epoch. The coordinate yields the evaluation's
  `ProgressDeltaPPM`, which the Guild accrual hook and receipts consume. Retargeting it would
  change Tier-1 Guild contribution and receipts, which is out of this RFC's scope. Revisit once
  the gate literal is ratified.

## Changelog

- 2026-09-25: created (draft) by Claude for Marco's batch acceptance.

## Owner acceptance (2026-09-25)

Marco accepted this RFC in the 2026-09-25 batch (the answer to a direct batch question in-session,
recorded by Claude). Every owner decision this RFC lists is **ruled at its stated recommended
default**. Where the text offers alternatives, the recommended option is the normative one and the
alternatives are rejected. Provisional numbers stay provisional and are ratified by harness
measurement and SHA, as the RFC already requires. Owner-authored copy stays owner-authored:
implementation ships copy keys, with any placeholder or candidate text clearly marked.
**Exception:** OD-1 (source of the hiring budget) is NOT ruled: it conflicts with Headcount Allocation OD-1 and is held for an owner explanation; implement nothing that depends on the seat source until it is ruled.
