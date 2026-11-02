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

### B — Tier-2 catalog content (engine-owned Headcount schema)

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
  "roles": [ { "kind": "synergy_feed", "pool_id": "pool.org_chart" } ] },
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
  "roles": [ { "kind": "synergy_feed", "pool_id": "pool.org_chart" } ] }
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

B5. **The cap-mints-a-currency law (`design/02 §2c` law 2).** The paid-hire pool has
the engine's required visible seat hardcap/reason (H1; Headcount OD-12). Tier-2
catalog/copy and the epoch changelog must name its parallel-currency response;
no office purchase silently raises the cap or grants free staff. That content
obligation is not completed by selecting the hiring model. The epoch-9 changelog
publishes the faucet/sink budget (law 1):

- **Sinks:** the gate burns its ratified amount; generator/upgrade purchases;
  paid hires at the pool's ratified cost curve. No payroll is introduced.
- **Faucets:** the three generators, the `pool.org_chart` factor and the canonical
  resource-producing/redirecting Headcount terms, each with its formula.

### H — Headcount allocation (paid hiring; canonical engine contract)

H0. **One engine authority.** Marco's 2026-11-02 ruling selects paid hiring with
producing/redirecting roles and delegates reconciliation of both specifications.
[Headcount Allocation](headcount-allocation.md) owns the single `headcount_pools`
grammar, exact effective-heads/rate algorithms, saved maps, logged commands,
events, replay and projection contracts. This RFC owns its actual Tier-2 content
and mounted player/harness consumers. There is no second generator-seat or
per-effective-head multiplier engine.

H1. **Tier-2 pool content.** The candidate must supply a complete engine S1 pool:
its mechanical ID, Company-resource hire price on an existing cost curve,
visible seat hardcap/reason, role rows and window. The Tier-2 pool opens at
`gate.t1_to_t2` and uses `to_gate: null` until the accepted Tier-3 automation
successor specifies closure. Office generator purchases do not grant staff.

- The four content roles remain Engineer, Sales, Support and Middle Manager.
- Engineer uses an executed `produce` effect for billable Company-resource output.
- Middle Manager uses `convert` from Engineer with `heads_per_converter: 1`.
  Diverted heads leave Engineer production and produce the declared deck resource;
  they are never duplicated. Extra managers are visibly idle.
- Sales/Support must use an executed effect from the engine's closed
  `produce | convert | pool_feed` vocabulary, subject to its source/binding rules.
  Their concrete effects/resources and rates remain HG-3 content work; the former
  generator/manual `per_effective_ppm` rows are rejected by this model ruling.
- HG-4 now owns the actual deck resource/cap/consumer content. A derived count
  named `slide_decks` that produces nothing is not this model. Do not invent a
  gate requirement, offer rule or new sink to fill the missing content.

The paid-hire price, role rates and complete resource/pool bindings require a
reviewed candidate and measured balance before M3 ratification. Engine fixture
IDs/prices/rates are not Tier-2 content. The ruling selects the model, not those
literals or a complete pool row.

H2. **Exact derived values and closed-form output.** Use engine S2's producing
heads, diverted heads, idle converters and unassigned seats, without another
formula. Use S3's resource-rate terms and normal ordered `all`/role-targeted
contributions. Generator-targeted contributions do not become staff production.
Headcount remains constant between actual commands; no tick, payroll, accrual
read payment or retroactive allocation is added. The existing 90%/24h offline
policy and resource hardcaps remain unchanged. Free, instant reallocation and
the 1:1 diversion model are preserved.

H3. **One state and activation owner.** Use engine S4's complete-key
`headcount_seats` and `headcount_assigned` maps and invariant
`0 <= sum(assigned in pool) <= seats[pool] <= seat_hardcap.count`. Derived
values are never saved. New runs start at zero; Exit resets the next Company and
retains the terminal history under its pins. Do not add the rejected alternative
`headcount_assignments` field or a second version floor. The engine/save owners
must reconcile their occupied historical schema/Company/replay version literals
before live integration; this model ruling allocates none.

H4. **One command/event/replay union.** The mounted controls and harness submit
engine S5's `hire_headcount {pool_id,count}` and
`reallocate_headcount {pool_id,assignment}`, with the normal logged envelope.
The latter is the complete absolute vector for that pool. Engine S5 owns exact
rejections, affordability/cap handling, pre-command accrual and idempotency;
engine S6 owns `headcount_hired` and `headcount_reallocated`. There is no
`set_headcount_allocation` alias or `headcount_allocated` event. Buying staff
does not increment generator purchase counters. No new behavior is inferred for
firing, payroll or Tier-3 automation.

H5. **Projection, formulas and mounted consumer.** Engine S3/S10 owns authoritative
per-pool seats/cap/unassigned/window and per-role assigned/producing/diverted/idle
heads, resource IDs and rates. The actual deck resource amount is resource state,
not the converter count. The mounting RFC carries that contract through the
registered API and receipt to the panel; no client-derived budget becomes
authority. Publish the canonical engine rate formula through its reviewed
regeneration commit, not a second per-head multiplier formula.

H6. **Simulation masks and measured roles.** Engine S7.4 masks a role's rate and
pool feed while retaining seats, assignment and hire cost. Generator masks no
longer remove a fictitious seat faucet or clamp a second assignment map. Paid
hiring is a purchasable for relevance; P5 still requires an allocation-aware
reference and meaningful role choices. The complete P1 hiring/allocation policy
must be reconciled to the actual engine before calibration; a window-only subset
is not full P1 and does not waive Headcount, accounting or release acceptance.

### E — Era presentation hooks (Tier 2 = `era_2010`)

E1. **Theme.** Add `ui/themes/era_2010.json`: the same 41 tokens and the same schema as
`era_2000`, in the `design/08 §2` "2010 flat startup" direction. Token values are
owner-reviewed design data. The following gain `era_2010`:

- `UI_ERAS` and `CopyEra`;
- `eraForSnapshot` (tier 2 → `era_2010`);
- `RunEndSurface` (tier 2 → `era_2010`).

Tier ≥ 3 remains a thrown `RangeError`. That fail-closed behavior is kept deliberately and is
tested.

E2. **Mounting/API-owned snapshot successor.** The mounting/API owners assign
the next actual version for H5/S10; this model ruling does not reuse an occupied
snapshot version. Existing receipts remain replayable under their minted versions.

- `transitions` projects the `gate.t1_to_t2` control. It is derived, like the first Gate, by
  invoking the production transition on a discarded clone.
- It projects an `incorporate` control listing the pinned faction catalog's rows. Their existing
  `incorporation_copy_key`s supply the labels, and the control is enabled only at Tier ≥ 2 with
  no faction.
- It carries the H5 `headcount` block.
- Any other gate or route stays fail-closed.

E3. **Desk surfaces.**

- A **Headcount** panel consumes engine S10: paid Hire/Max, seats/cap/reason,
  per-role absolute-vector editing, Unassigned, producing/diverted heads and idle
  converters. It submits `hire_headcount` and `reallocate_headcount`; only receipts
  change authoritative state. Show the actual deck resource amount/rate separately
  from the number of diverted heads; no derived count pretends to be stock.
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
  `[perform_manual_batch, buy_generator, buy_upgrade, hire_headcount, reallocate_headcount, wait]`.
- Allocation arm keys are `move:<from>><to>`:
  - `from` ∈ {`unassigned`} ∪ {roles with count > 0};
  - `to` ∈ roles;
  - `to ≠ from`;
  - raw-byte ascending;
  - materialize one complete absolute per-pool assignment for `reallocate_headcount`,
    never a partial map or delta command.
- The three policies:

Literal inheritance below is pinned to the immutable `first_hour_policy.v1`
registry version 3, `balance/testdata/t0-t1/first-hour-policy-v1.json`, SHA256
`e5e5de7051beb0340e54f7013ce7d4a48c35bfcc3343220310290478445d10c3`.
The adopted [Tier 2 Pacing Activity Windows](tier2-pacing-activity-window.md)
amendment replaces only the new continuous personas' activity windows as below;
it does not alter the old registry or waive any allocation policy.

| Policy | Rules |
|---|---|
| `chaos.t2` v1 | The pinned `chaos.t0_t1` literals except the continuous-session replacement below; the complete reconciled policy must include paid-hire candidates and legal allocation moves in its seeded uniform draw (HG-10). |
| `casual.t2` v1 | The pinned `casual.t0_t1` activity/decision rules plus paid hiring and allocation to executed role effects. Its concrete hire/allocation ranking remains HG-10; the old `all`/`per_effective_ppm` rule does not apply to the selected engine. It must assign unassigned hires visibly and never manufacture a free allocation budget. |
| `reference.greedy` v2 | The T01-C20 projected-time ranker, preceded at each decision boundary by a strictly improving single-unit allocation hill-climb with raw-byte `move:` tie-break; at most `seats[pool]` moves per pool, loud `allocation_starved` on exhaustion. HG-10 must bind paid-hire costing and post-hire unassigned state without assuming newly hired heads already produce. Target sequence remains `gate.t0_to_t1` → `gate.t1_to_t2` → `1e9` cash, then banking. |

**HG-10 — Complete policy binding remains engineering work.** The model ruling
does not supply the missing deterministic hire quantity/key enumeration, Casual
effect ranking or Reference hire/allocate projection contract. Specify and
independently review those against the actual engine/candidate before implementing
the full P1 registry or claiming a calibration. Do not reuse the rejected
multiplier fields, synthesize a free hire, drop hiring from the workload, or call
the existing window-only subset complete. The adopted activity windows below and
all P2 populations/targets remain binding while this work is completed.

The NEW `chaos.t2` v1 and `reference.greedy` v2 each have one nonrepeating,
zero-jitter session starting at 0 and lasting 21,600,000ms. Chaos's cadence is
2,000ms; Reference's is 1,000ms. `casual.t2` v1 inherits the complete pinned
15-minute/30-minute session definition, independent per-session jitter and
5,000ms cadence; jittered durations are not forced to the nominal 15 minutes.
These are test personas, not six-hour player-session requirements.

Session coordinates are Founder-genesis wall time, span Company restarts and
remain half-open. Scripted/elective endings do not restart a window. P2's horizon
clips activity; it never extends or rewrites the session. Keep unreached seeds
visible. The new policy IDs/versions participate in the copied seed-derivation
grammar, so decision/jitter cohorts change: do not reuse old IDs to retain old
draws or call this a duration-only paired experiment. Old policy bytes/reports
remain historical; new observations identify their actual policy/scenario/source
bytes at new output paths. Strict v1 validation stays closed; these rows require
the complete reviewed schema-v2 implementation, not a relaxed or partial registry.

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
  - `milestone.first_headcount_allocation` — `intent_applied`, `reallocate_headcount`,
    `founder_attended_ms`, `must_reach`. It is an unbounded drift observation with no envelope.
- **Envelopes** (`design/02 §11`: "First minigame ~2–3 hours (Tier 2)"): the Chaos p50 and the
  Casual p50 of `milestone.it_company_gate` both lie in [7,200,000, 10,800,000] ms.
- Invariants: the nine T0–T1 invariants plus the engine-owned `headcount_bounds`
  (complete safe-integer keys and `Σ assigned ≤ seats ≤ seat_hardcap` per pool
  at every applied revision).
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

- `provision`: advance across the grid;
- `synergy_feed`: exercised pool target.

Paid hiring is additionally measured as a purchasable through engine S9's
ANY/ALL relevance path. Per-role masked controls use H6/engine S7.4, retaining
paid seats and costs while removing executed role output; no generator-seat
faucet or greedy assignment clamp exists.

P5. **Allocation-decision gate.** This is the relevance analog for roles, which are not
purchasables. On the P3 combined scenario's reference trajectory, `role.engineer` and
`role.sales` must each hold a nonzero effective allocation at some decision boundary. That
proves the allocation decision is not dominated by a single role. `role.support` and
`role.middle_manager` shares are reported as observations only (HG-3, HG-4).

### M — The epoch-9 mint

M1. **Artifacts edited in place** (the epoch-7 convention):

- economy (the Headcount engine's reviewed schema/activation version; its old v5
  draft literal is occupied and is not a reservation);
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

1. **Canonical engine contracts/loaders.** Headcount owns Go/TypeScript/JSON Schema
   and its mutation corpus. Close its historical identity/AC-3 technical issues;
   Tier 2 consumes that grammar rather than implementing a second engine.
2. **H3–H5 integration.** Engine-owned state, commands, events, rates, projection
   and shared replay corpus; actual consumer wiring, not duplicate definitions.
3. **Formulas regeneration commit.**
4. **Harness grammar.** Complete HG-10's paid-hire/allocation policy contract and
   reviewed full P1 registry, scenario `run_seq: null`, H6 masks and P4/P5 gates,
   each with its negative control. Preserve adopted activity windows and old v1.
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
7. **Headcount model is ruled:** paid hires and canonical producing/redirecting
   roles, with free reallocation. Concrete Tier-2 role/resource/rate/cap content
   still requires the candidate/measurement/ratification lane (HG-3/HG-4). Middle
   Manager is a role, not the research bank's "Building"; decks are real resource
   output, not a nonproducing derived allocation count.
8. **The FarmVille beat is a non-stateful chrome stub.** No energy or timer mechanic exists.
9. **The T2 minigames are not included.** They belong to their own RFCs.

## DESIGN-GAPs

| Gap | Question | Options | Recommended |
|---|---|---|---|
| HG-1 | Budget source | Paid hires under the canonical Headcount engine versus office-granted seats | RULED2026-11-02: paid `hire_headcount` with cost curve and visible cap; office-granted seats rejected. Price/cap literals require candidate ratification. |
| HG-2 | Reallocation cost/timing | (a) free, instant, unlimited; (b) cooldown per move; (c) onboarding ramp (new effective counts phase in) | (a): timing emerges from budget growth; (b)/(c) add hidden clocks with no design text |
| HG-3 | Concrete role content | Executed `produce`, `convert`, `pool_feed` under engine S1/S3 | Producing/redirecting model ruled; complete Sales/Support effect/resource/rate rows still require reviewed candidate content and P5 proof. No legacy per-head multiplier authority. |
| HG-4 | Deck resource content | Actual converter output resource/cap/consumer versus the rejected derived count | Actual resource output ruled; candidate content must define the resource and its meaning. No new T3 gate/offer/sink is inferred. |
| HG-5 | Morale, crunch → Soul, clause 7(c) | own this RFC / Events Layer 1 + Soul debit rows | defer; an events-engine consumer is required anyway |
| HG-6 | Procurement auto-buy | own this RFC / successor mechanic RFC | successor RFC |
| HG-7 | T1→T2 doctrine | ship a choice-only row now / defer until doctrine effects exist | defer: an effectless choice would be hollow |
| HG-8 | Reputation at T2 Exits | lower the threshold now / defer to Reputation tree v1 | defer: Reputation has no consumer yet, so a nonzero number would do nothing |
| HG-9 | T2 band | [1e7,1e9) interleave / move ratified `gate.t2_to_t3` (cascades into Permits/T3) | interleave |
| HG-10 | Full paid-hire/allocation policy | Bind existing engine/candidate to deterministic complete P1 choices, quantity keys and post-hire projection | Technical specification/review required before full P1 implementation/calibration; no change to adopted windows or pacing targets. |

## Acceptance criteria

Each criterion ships with its demonstrated failing case, which is recorded in the planning log
before the green run is claimed. All gate claims run with `-count=1`.

0. **Reachability.** Under the epoch-9 candidate bundle, a Tier-1 company crosses `gate.t1_to_t2`
   and gets `Tier == 2`, and `incorporate` then applies.
   *Failing case:* the same test under the epoch-8 bundle rejects `unknown_id` and
   `not_eligible/tier`.
1. **Strict loaders.** The canonical Headcount economy version, Routes and Categories load identically in Go and TypeScript.
   Every engine S1 rejection, plus a categories set missing `gate.t1_to_t2`, fails in both runtimes and
   in the JSON Schema. v1–v4 catalogs re-load byte-identically.
   *Failing case:* each mutation fixture must be observed red.
2. **Headcount transition parity.** A Go-authored sequential corpus is replayed byte-for-byte by
   TypeScript. It covers every H4 outcome, conversion with E=3/M=5 (producing
   Engineer heads 0, diverted heads 3, idle managers 2—not a stock amount),
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
   *Failing cases:* the registry's dead-upgrade fixture fires; a P5 control with
   Sales' actual executed effect severed after valid catalog admission fails;
   a deliberately perturbed T0 row breaks the identity comparison.
6. **Pacing.** The P2 envelopes are green, must-reach holds at every seed, and
   the canonical `headcount_bounds` invariant holds. The first-hour scenario, re-run on epoch 9, keeps all
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
    `gate.t1_to_t2`, pays for actual hires, reallocates headcount through the canonical
    commands, verifies durable replay, and shows an epoch-8 run
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

1. **HG-1 budget source.** RULED2026-11-02: paid hires under Headcount S5.1,
   with its cost curve and visible pool cap. No seats from office purchases.
2. **HG-2 reallocation.** Recommended: free, instant, unlimited.
3. **HG-3 role effects.** RULED2026-11-02: producing/redirecting roles under
   engine S1/S3, not the rejected generator/manual per-head multipliers. Actual
   Tier-2 role content and balance literals retain their candidate/adoption gates.
4. **HG-4 Slide Decks.** The selected converter produces the actual deck resource;
   it is not a derived count that produces nothing. Concrete resource/consumer
   content remains H1/HG-4 work; no new gate, offer or sink is adopted here.
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
- 2026-11-02: Marco adopts the independently reviewed new-persona activity windows
  and delegates P1 reconciliation; retain P2 targets/horizons, old identities and
  all Headcount, full-policy, balance, content and release gates.
- 2026-11-02: Marco selects paid hiring with producing/redirecting roles and
  delegates both-body reconciliation. Replace the competing Tier-2 engine with
  canonical Headcount ownership; retain explicit content/full-policy/identity gates.

## Owner acceptance (2026-09-25)

Marco accepted this RFC in the 2026-09-25 batch (the answer to a direct batch question in-session,
recorded by Claude). Every owner decision this RFC lists is **ruled at its stated recommended
default**. Where the text offers alternatives, the recommended option is the normative one and the
alternatives are rejected. Provisional numbers stay provisional and are ratified by harness
measurement and SHA, as the RFC already requires. Owner-authored copy stays owner-authored:
implementation ships copy keys, with any placeholder or candidate text clearly marked.
**OD-1 resolution, 2026-11-02:** Marco explicitly selects paid hiring with
producing/redirecting roles and delegates reconciliation of both specifications.
The former office-seat/per-head-multiplier model is rejected; H0–H6 and the
dependent body sections above consume the canonical Headcount engine. This does
not ratify prices/rates, copy, content, calibration or a release. Complete role
content, policy binding and engine identity/acceptance issues remain explicit.

## Activity-window adoption (2026-11-02)

Marco's explicit answer: “Adopt the reviewed windows and delegate parent-text
reconciliation.” This adopts W1–W4 of `tier2-pacing-activity-window.md`, technically
approved in `059e8706..8aa53527`, and authorizes the matching P1 body edit above.
The translation requires independent technical review. It does not adopt Headcount
choices, the accounting/search drafts, authored copy, balance literals or a release;
it neither implements the new registry nor completes any P2 acceptance gate.
