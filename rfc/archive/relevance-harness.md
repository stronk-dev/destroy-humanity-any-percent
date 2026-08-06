# RFC: Relevance Harness

- **Status:** implemented. Archived against a test-only schema-v4 fixture; T0–T1 owns the first
  production relevance policy, scenario, golden report, and epoch mint.
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-01
- **Design refs:** `design/02 §11b` (tier-relevance doctrine — this RFC is its enforcement half)
- **Research:** `design/research/balance-enforcement.md` (restricted play, Demaine bounds, Riot ANY/ALL, GEEvo), `design/research/tier-relevance.md`
- **Depends on:** Balance Harness Foundation (implemented), Run Genesis (implemented — `ApplyLogged` is the transition the solver drives), **T0–T1 Playable Content (draft — the first real catalog this instrument measures; implement the harness first, gate on real content immediately after)**
- **Planning:** `planning/relevance-harness/` (once implementing)

## Summary

Make "every purchasable stays meaningful" a machine-checked law. Three deliverables: a
**reference near-greedy persona** (the approximately-optimal player), a **relevance report**
(per-purchasable counterfactual values, golden artifact), and a **CI relevance gate** (ANY/ALL
floors that fail a balance change shipping dead content). No shipped precedent exists for the
gate; the components have academic anchors (restricted play — Jaffe 2012; greedy bound — Demaine
et al. 2018; segment quantifiers — Riot's live framework).

## Specification

### V1 — The reference persona

A scripted persona `reference_greedy` driven by **payback period**: at each decision point, buy
the affordable purchasable minimizing `max(cost − bank, 0)/rate + cost/Δrate` (the Cookie Monster
metric), including indirect effects the harness can see (milestone crossings, synergy-pool
deltas). Justification: exact optimal buy-ordering is strongly NP-hard in the discrete-tick
regime; greedy achieves 1+O(1/log M) (Demaine et al., arXiv:1808.07540) — the reference is
declared *approximately optimal*, never optimal. Deterministic by construction (ties broken by
purchasable id). Divergence check (once, then per-epoch): a beam-search persona (width = config)
on the Phase-0 catalog measures the greedy gap; a gap > declared bound fails the harness — our
curve family must stay in greedy-friendly territory or the reference is re-justified.

### V2 — The relevance report (golden artifact)

Per epoch, per persona (existing chaos/casual + `reference_greedy`), the harness runs:

1. **Baseline routes** (existing runs, unchanged).
2. **LOO ablations:** for each purchasable p, re-run with p's effect zeroed (catalog-level mask —
   the purchase is still buyable and spends, its effect is null: this measures the EFFECT's value,
   not route perturbation; a second variant removes p from the action set for trap detection).
   Record `Δt(p, persona, milestone)` for every milestone declared by the selected relevance
   scenario (R3 — not a fixed "16"; HEAD has 4 milestones + 16 policy/statistic observations).
3. **Group ablations:** per tier and per catalog category, all-members-zeroed
   (substitute-redundancy detection — two near-substitutes each show ≈0 LOO while jointly
   essential; group relevance can rescue an individual member as `group_supported`, R4).
4. **Per-purchase instantaneous value:** payback period at the moment each persona bought it.
5. **Counterfactual tier contribution (R5 — renamed from "shares"):** per milestone segment, each
   tier's group-ablation delta (the measured impact of ablating that tier's effects — NOT required
   to sum to 100%), plus each tier's non-production role activations (capacity binding, minigame
   input use, stock rate, sacrifice) — "a tier fell off" = its counterfactual contribution below
   floor across a segment, a recorded fact.

Report format follows the existing golden discipline: byte-identical integers/canonical strings,
no floats, regenerated only via the BALANCE-CHANGE/CONSTANTS-IDENTITY protocols. Cost bound:
LOO count = |purchasables| × personas × 1 run each; group ablations = |tiers|+|categories| runs —
all deterministic and parallel; a `relevance_budget_max_runs` config fails the build if the
catalog outgrows the budget (explicit, never silently sampled).

### V3 — The CI gate (the Riot ANY/ALL shape)

For every purchasable p with declared availability window W (catalog field, new, required):

1. **Relevance floor (ANY):** ∃ persona with `Δt(p) ≥ epsilon_ms` for some milestone inside W —
   deleting p's effect must cost SOMEONE real time. `epsilon_ms` is catalog data (provisional
   30_000).
2. **Trap floor:** the reference persona purchases p at least once inside W (a purchasable the
   approximately-optimal route never touches is a trap; it may exist ONLY with an explicit
   catalog `trap_exempt: true` + changelog justification — satirical trap upgrades are on-theme
   and allowed, but always deliberate).
3. **Role floor:** every generator class shows ≥1 non-production role activation in the report
   (the design/02 §11b law, measured end-to-end, not just loader-declared).

Gate failures name the purchasable, the failing floor, and the nearest passing epsilon — a
balance change that kills content fails CI with the evidence attached. The gate runs on every
BALANCE-CHANGE and every mint.

### V4 — Tuning loop (deliberately manual at Phase 0)

The report is the instrument; automated parameter search (CMA-ES over log-space catalog constants
with floor-violation penalties — the GEEvo shape, noise-free because we're deterministic) is a
NAMED FOLLOW-UP, not this RFC: the first epochs of relevance data must teach us what the floors
should be before an optimizer chases them (Goodhart risk is real and the satire game about metric
gaming should not itself ship a naively metric-gamed economy).

### V5 — Precisions (2026-08-03, answering the implementer's under-specification review)

1. **Ablation semantics, exactly two modes, both closed:** `effect_mask` (the purchase remains
   buyable and spends; its effect contributions — production, roles, synergy-pool feeds — are
   nulled at the contribution layer) and `action_removal` (the purchasable leaves the action
   set). LOO uses `effect_mask` (measures the EFFECT's value); trap detection uses
   `action_removal`. Group ablations use `effect_mask`. No third mode.
2. **Availability window:** a required field in the **`relevance_policy` artifact** (R6 — keyed by
   purchasable ID, NOT an economy-catalog field), `window: {from_gate,
   to_gate|null}` — gate-bounded, not time-bounded (windows are progression-relative; the
   pacing harness already maps gates to time distributions per persona, so time floors derive).
3. **Roles/categories:** the role vocabulary is design/02 §11b's closed list, read from the
   SAME catalog declarations the loader enforces — the report never invents a role taxonomy;
   category grouping = the catalog's existing generator-class/upgrade-family fields.
4. **Beam search:** width 8, depth = full run, config in the harness scenario file; divergence
   bound declared there (`greedy_gap_ppm`, provisional 50_000 = 5%); the check runs per epoch
   in `harness-check`, not per commit.
5. **Unreachable milestones:** Δt against a milestone the ablated route never reaches within
   horizon = `horizon − baseline_time` (a finite, conservative lower bound), flagged
   `unreached: true` in the report — never ∞, never dropped; a purchasable whose ablation makes
   a milestone unreachable is maximally relevant and the report says so legibly.
6. **CI artifact ownership:** the relevance report is a golden artifact under
   `testdata/harness/relevance/`, owned by the same `harness-check`/BALANCE-CHANGE machinery as
   the pacing baseline — same guard, same identity classes, no new protocol.

## Acceptance criteria

1. Reference persona: deterministic golden route on the Phase-0 catalog; beam-search divergence
   check within declared bound; both under the existing byte-identical report discipline.
2. Relevance report golden fixture: a seeded dead upgrade (effect zeroed in a fixture catalog) is
   flagged by floor 1; a seeded trap (never bought by reference) by floor 2; a seeded roleless
   generator by floor 3; a seeded near-substitute PAIR passes LOO but fails its group ablation.
3. Gate wiring: a BALANCE-CHANGE commit introducing the seeded dead upgrade fails CI with the
   named evidence; `trap_exempt` + changelog line passes.
4. Budget: exceeding `relevance_budget_max_runs` fails loudly.
5. Report determinism: byte-identical across two full regenerations.

## Open questions

- `epsilon_ms` per-tier scaling (a 30 s floor means different things at hour 1 vs hour 40) —
  first-epoch data decides; the field is per-window-declarable already.
- Whether MMO-side surfaces (commons/guild contributions) enter relevance windows — deferred to
  the follow-up after composition ships live data.

## Acceptance-review blockers (2026-08-05)

The stale Purchasable Content dependency is genuinely cleared: `SimulateTransition`, the two
ablation modes, typed role activations, and the authoritative harness runner exist. The remaining
gaps are in the solver and gate contracts, not the production engine. Codex must not choose these
definitions in implementation.

### R1 — `reference_greedy` is not an executable policy

V1 gives a formula but not the policy cadence, candidate set, buy count, wait rule, multi-resource
comparison, zero/negative-rate handling, or exact arithmetic/tie order. "Including indirect
effects" is especially ambiguous for fixed-grid provisioning, ladders, synergies, and gates: there
is no unique instantaneous `delta_rate` for an effect that materializes later.

**Proposed contract:** define the policy as event-driven over the scenario's authoritative
candidate provider. At each decision it evaluates one-unit generator buys and each unowned upgrade
by cloning committed state and running the real simulation boundary to the next declared decision
horizon; unaffordable candidates compute their earliest affordable time through the existing
closed-form quote. Rank an exact tuple (declared payback value, earliest benefit time, raw-byte ID),
with explicit sentinels for zero benefit/unreachable candidates. If no candidate is reachable,
advance to the next scenario boundary or fail the required milestone. The RFC must state the exact
comparison equation and Decimal/integer rounding; `max` buys are not implicit.

### R2 — The beam oracle and `greedy_gap_ppm` have no search contract

Width 8 and "depth = full run" do not define expansion cadence, node score, dominance pruning,
state deduplication, terminal objective, tie-breaking, or which milestone/time distribution the 5%
gap compares. Different reasonable implementations produce different golden routes.

**Owner ruling required:** enumerate the beam node key, action expansion order, wait edges, score,
pruning/dedup rule, terminal milestone vector, and integer gap equation. Recommended scope is a
small-catalog fixture oracle first, with the production per-epoch check enabled only once T0-T1
declares one primary optimization milestone. The 50,000 ppm value is provisional balance policy,
not a research-established constant.

### R3 — Run matrix and delta aggregation are unspecified

V2 says per epoch/per persona but does not say which seeds, scenario runs, or baseline routes are
paired with each ablation. V3's ANY quantifier cannot be evaluated until `delta_t` is reduced across
seeds. The sign is unstated, and the unreachable formula assumes a baseline time even when the
baseline itself is unreached. The text also repeats the withdrawn "16 milestones" claim; HEAD has
four milestones and sixteen policy/statistic observations.

**Proposed contract:** every ablation is paired one-to-one with the same baseline run key except
for an explicit mask dimension; define `delta_t = ablated_time - baseline_time`. Declare the seed
reducer used by the gate (recommended conservative `p05` or `worst`, not an average), and exact
rules for all four reached/unreached pairs. Replace "16 pacing milestones" with "every milestone
declared by the selected relevance scenario". ANY ranges over declared personas after seed
reduction, never over a lucky seed.

### R4 — Individual and group gates contradict each other

V2 says group ablations rescue the known LOO blind spot where two substitutes each have near-zero
individual delta but are jointly essential. V3 nevertheless requires every individual to pass its
own LOO floor. AC2 then says the pair "fails its group ablation," although a large group delta is
evidence the group matters. No group threshold or attribution back to members is defined.

**Owner ruling required:** decide whether group relevance can satisfy an individual member's floor.
Recommended exact shape: an item passes relevance if its individual LOO passes OR it belongs to one
declared group whose group delta passes; the report labels the latter `group_supported`, while the
reference-purchase trap floor remains individual. Define one membership per grouping axis, group
epsilon, and correct AC2 so the substitute pair fails individual LOO but passes group relevance.

### R5 — Production "shares" have no valid attribution formula

With multiplicative slots, synergy pools, provisioning, and cross-tier effects, total production
cannot be uniquely partitioned into additive tier shares. Raw base-rate shares ignore the effects
this report exists to measure; counterfactual marginal shares need not sum to 100%. "Nearest passing
epsilon" and gate-bounded milestone selection are likewise not defined mathematically.

**Owner ruling required:** either rename this output to counterfactual tier contribution and define
it as the exact group-ablation delta (not required to sum to one), or specify a real attribution
method. Also define how `{from_gate,to_gate}` selects milestone segments at inclusive/exclusive
boundaries, which epsilon applies when multiple milestones are inside the window, and the exact
nearest-passing diagnostic.

### R6 — Required catalog metadata does not exist

Schema-v4 generator classes have `tier` and `category` but no availability window, epsilon, or trap
exemption. Upgrades have a window but no family/category, epsilon, or trap exemption. The current
production Phase-0 catalog is schema v3 and has none of the Purchasable fields. Therefore V3 cannot
load its claimed required fields, and V5's "existing upgrade-family fields" claim is false.

**Owner ruling required:** choose the authority. Recommended: a hash-pinned relevance-policy
artifact keyed by economy purchasable ID, containing window, epsilon, trap exemption + justification
key, and declared group IDs. This avoids making harness-only policy part of the production economy
grammar while still making it epoch-owned and fail-closed. If the economy catalog owns the fields
instead, specify its next schema, Go/TS parity, artifact mint, activation boundary, and upgrade
family grammar explicitly.

### R7 — Report and budget wire formats are not enumerated

"Follows the existing golden discipline" does not define a relevance report schema. The run-budget
formula omits action-removal runs and does not say whether baselines are shared, how nested groups
count, or where `relevance_budget_max_runs` lives. A byte-identical artifact needs exact structs,
ordering, schema/version identity, and a total before work is dispatched.

**Proposed contract:** enumerate the report envelope and every ordered row family; prohibit maps and
floats as the base harness does; extend the run identity with ablation mode/target; sort all IDs by
raw bytes; declare the exact preflight cardinality equation including baseline, effect-mask,
action-removal, group, reference, and beam runs; fail before dispatch when it exceeds the scenario's
positive safe-integer budget. Add schema fixtures for duplicate/missing rows and a cardinality test.

### R8 — The real-content gate and archival boundary conflict

The RFC says implement now against Phase 0, but the checked-in Phase-0 economy catalog is schema v3:
one generator, no upgrades, roles, categories, or relevance metadata. It cannot satisfy the dead
upgrade, trap, roleless generator, substitute-group, or per-window acceptance cases. T0-T1 remains
a real dependency for the production relevance baseline even though the simulation seam itself is
implemented.

**Proposed contract:** split acceptance explicitly. This RFC may implement and archive the generic
solver/report/gate against a checked-in test-only schema-v4 fixture with discriminating failures.
The T0-T1 mint must, in the same change, add the first production relevance policy/scenario and
golden report and make `harness-check` fail closed if an active schema-v4+ production catalog lacks
them. Do not label the current schema-v3 catalog a real relevance baseline, and do not silently
skip an active content catalog.

## Changelog

- 2026-08-01: created (draft) from the banked tier-relevance + balance-enforcement research;
  enforcement half of design/02 §11b.
- 2026-08-05: Codex acceptance review — Purchasable substrate confirmed, but solver/gate contracts
  blocked on R1-R8; stale milestone and catalog-field claims identified and proposed resolutions
  recorded without inventing policy semantics.

## Owner rulings on R1-R8 (2026-08-05)

Accepting Codex's proposed contracts (sound) and ruling the four owner-calls (R2/R4/R5/R6) along the
recommended shapes. Full contracts, to avoid a re-bounce.

- **R1 — accepted.** `reference_greedy` is event-driven over the scenario's authoritative candidate
  provider: at each decision it evaluates one-unit generator buys + each unowned upgrade by CLONING
  committed state and running the real simulation boundary to the next declared decision horizon;
  unaffordable candidates compute earliest-affordable-time via the existing closed-form quote. Rank
  the exact tuple `(declared_payback_value, earliest_benefit_time, raw_byte_id)` with explicit
  sentinels for zero-benefit and unreachable candidates; if none reachable, advance to the next
  scenario boundary or fail the required milestone. The exact comparison equation + Decimal/integer
  rounding order is stated in the impl; `max` buys are NOT implicit. (This resolves "indirect
  effects": the payback is measured by real simulation to the horizon, not an instantaneous
  delta_rate.)
- **R2 — RULED (beam oracle).** Scope: a **small-catalog fixture oracle first**; the production
  per-epoch greedy-gap check activates ONLY once T0-T1 declares one primary optimization milestone.
  Beam contract: node key = the committed sim-state digest; expansion = the R1 candidate set
  (one-unit buys + unowned upgrades) in raw-byte-ID order PLUS a "wait to next horizon" edge; score =
  time to the scenario's declared terminal milestone vector (lower better); **Pareto dominance
  pruning** (a node is dominated if another reaches ≥ the same milestone progress in ≤ time with ≥
  resources); dedup by state digest keeping the best; width 8 keeps the 8 best-scored non-dominated
  nodes per expansion; tie-break by raw-byte path. `greedy_gap_ppm = ⌊(greedy_time − beam_time) *
  1e6 / beam_time⌋`; gap > `greedy_gap_max_ppm` fails the harness. The `50_000` ppm is PROVISIONAL
  balance policy (data), not a research constant.
- **R3 — accepted; the stale "16 milestones" claim is reconciled.** Every ablation is paired
  one-to-one with the same baseline run key except an explicit mask dimension; `delta_t =
  ablated_time − baseline_time`. The gate seed-reducer is **conservative (`p05` or `worst`, NOT
  average)**, declared per scenario; exact rules for all four reached/unreached pairs (both reached →
  signed delta; ablated-unreached-but-baseline-reached → +∞ / "regressed to unreachable"; baseline
  itself unreached → the pair is excluded and the milestone is flagged, never assumed a time). `ANY`
  ranges over declared personas AFTER seed reduction (never a lucky seed). The body's "16 pacing
  milestones" is replaced with "every milestone declared by the selected relevance scenario" (HEAD
  has 4 milestones + 16 policy/statistic observations).
- **R4 — RULED (individual vs group).** An item passes relevance if its **individual LOO passes OR it
  belongs to ONE declared group whose group delta passes** — the latter is labeled `group_supported`
  in the report. The **reference-purchase trap floor stays INDIVIDUAL** (a trap must not hide inside a
  group). One membership per grouping axis; group epsilon declared in the R6 artifact. AC2 corrected:
  the substitute pair fails individual LOO but PASSES group relevance (labeled `group_supported`).
- **R5 — RULED (rename + counterfactual definition).** The output is renamed **"counterfactual tier
  contribution"** and defined as the exact **group-ablation delta** per tier — the measured impact of
  ablating that tier's effects — explicitly **NOT required to sum to 100%** (multiplicative slots +
  synergy pools cannot be additively partitioned; a counterfactual marginal is the honest measure).
  `{from_gate, to_gate}` selects the milestone segment `[from_gate inclusive, to_gate exclusive)`;
  when multiple milestones fall in-window, each item uses its own R6-declared epsilon; the
  nearest-passing diagnostic reports the smallest epsilon change that would flip a failing item to
  passing.
- **R6 — RULED (a separate relevance-policy artifact, NOT economy grammar).** Add a **hash-pinned
  `relevance_policy` artifact keyed by economy purchasable ID**, carrying `{availability_window,
  epsilon, trap_exempt + justification_key, group_ids[]}`. This keeps harness-only policy OUT of the
  production economy catalog grammar while making it epoch-owned and fail-closed. Go/TS parity,
  activation-boundary law, hash-pinned. The economy catalog is NOT extended for harness policy.
- **R7 — accepted.** Enumerate the report envelope + every ordered row family; prohibit maps and
  floats (base-harness discipline); extend run identity with `ablation_mode` + `target`; sort all IDs
  by raw bytes; the exact preflight cardinality equation counts baseline + effect-mask +
  action-removal + group + reference + beam runs; FAIL before dispatch if it exceeds the scenario's
  positive-safe-integer `relevance_budget_max_runs`; schema fixtures for duplicate/missing rows + a
  cardinality test.
- **R8 — accepted (the archival split).** This RFC implements AND archives the generic
  solver/report/gate against a checked-in **test-only schema-v4 fixture** with discriminating
  failures (dead upgrade, trap, roleless generator, substitute-group, per-window). The **T0-T1 mint
  must, in the same change**, add the first production relevance policy/scenario + golden report and
  make `harness-check` fail closed if an active schema-v4+ production catalog lacks them. The
  schema-v3 Phase-0 catalog is NOT a real relevance baseline; an active content catalog is never
  silently skipped.

Structure + solver + gate + wire ruled; the epsilon/window/threshold NUMBERS and the production
scenario are balance data / T0-1 content.

## Implementation acceptance recheck (2026-08-06)

R1-R8 settle the intended direction, but the live harness exposes seven residual executable gaps.
No implementation started: selecting these bytes/arithmetic in Go would make a golden report and
future balance policy follow an implementation accident.

### R9 — Reached/unreached semantics still contradict the normative body

V5 requires an ablated-unreached delta to be the finite lower bound `horizon - baseline_time` and
forbids infinity. R3 instead rules that pair as `+∞`, while R7 independently prohibits floats.
R3 also permits either `p05` or `worst`, so identical run rows can yield two legal gate verdicts.

**Proposed contract:** keep the finite V5 rule and encode every pair as the exact closed row
`{status, baseline_ms, ablated_ms, delta_ms}`. Status is `both_reached`, `ablated_unreached`,
`baseline_unreached`, or `both_unreached`. `both_reached` carries all three signed integer fields;
`ablated_unreached` carries baseline, null ablated, and `horizon_ms-baseline_ms`; either
baseline-unreached status carries null delta and is excluded from that milestone's floor with a
named report failure when the baseline milestone is required. Each relevance scenario declares one
closed reducer, `p05` or `worst`; the test fixture uses `worst`. Reconcile V5 and R3 in the body.

### R10 — The reference policy needs a wait boundary and an exact value equation

`production.SimulateTransition` can apply a real action to a clone, but it cannot advance a clone
without an action. Therefore earliest-affordable search and R2's wait edge do not exist. R1 also
delegates `declared_payback_value` and its rounding to "the impl" without defining what quantity is
compared when a candidate has delayed provisioning, synergy, ladder, or a different cost resource.

**Proposed contract:** add one harness-only `production.SimulateAdvance` beside
`SimulateTransition`. It accepts the same committed state/catalog/dependencies/mode/time/mask,
executes the same accrual and hook chain, returns role activations, and emits no intent, receipt, or
revision. The source-boundary guard permits only `server/harness` and tests. Earliest affordability
is the first canonical integer millisecond in `[now,next_horizon]` whose advanced clone can pay the
real one-unit quote, found by lower-bound binary search; unreachable uses a tagged sentinel.

The owner must still choose the exact candidate value. Recommended executable narrowing for the
first fixture: the relevance scenario declares one `resource_at_least` optimization milestone and
all candidate prices use that resource. For each candidate, advance baseline and candidate clones
to its earliest affordable time, buy exactly one on the candidate clone, then advance both to the
next horizon. Define gross marginal output as
`candidate_balance + paid_cost - baseline_balance`; if non-positive, value is unreachable.
`payback_ms = wait_ms + ceil(paid_cost * (horizon_ms-earliest_ms) / gross_marginal_output)` using
exact Decimal arithmetic and an exact-integer ceiling. Rank `(payback_ms, earliest_positive_delta_ms,
raw_byte_id)`. General multi-resource/objective scoring remains a successor unless the owner supplies
a conversion rule now.

### R11 — The beam oracle has no bounded executable score

"Time to the terminal milestone vector" is unavailable for a nonterminal node without a rollout;
the ruling does not choose a vector order, rollout policy, maximum decisions, or what a wait edge
waits to. Deduplication by state digest also omits whether virtual time/path participates, and the
run budget cannot bound an unspecified number of node simulations.

**Proposed contract:** the fixture oracle uses the single R10 optimization milestone. Node identity
is `{canonical_state_hash, virtual_ms}`; path is not identity and is only the final tie-break. Each
node expands the raw-byte candidate list plus exactly one edge to the next strictly greater declared
decision horizon. Score each child by a deterministic reference-greedy completion to the milestone
or scenario horizon; reached sorts before unreached, then lower completion time, then raw-byte path.
Dominance compares equal virtual-ms nodes componentwise over milestone reached/progress, every
canonical resource balance, and purchased counts. The scenario declares positive safe integers
`max_decisions` and `beam_width` (fixture requires width 8); reaching either terminal milestone or
max decisions stops expansion. Internal child/rollout simulations are counted in a separate
`relevance_budget_max_transitions`, preflighted from the declared maxima.

### R12 — `relevance_policy` is named but has no canonical grammar

R6 does not define schema version, top-level keys, row ordering/completeness, epsilon units,
trap-exemption nullability, group declarations, group epsilon ownership, or how tier/category group
IDs bind to the economy catalog. "Go/TS parity" has no vector format to consume.

**Proposed contract:** schema v1 exact root
`{schema_version:1, items:[...], groups:[...]}`. Items are raw-byte sorted and complete for the
economy catalog's generator+upgrade purchasable union, with exact keys
`{purchasable_id, availability_window:{from_gate,to_gate}, epsilon_ms, trap_exempt,
justification_key, group_ids}`. `to_gate` and `justification_key` are nullable; exemption iff
justification is non-null. Epsilon is a positive safe integer. Group IDs are sorted unique. Groups
are raw-byte sorted exact rows `{group_id, axis, member_ids, epsilon_ms}` where axis is
`tier|category|declared`, members are sorted unique and complete for the derived tier/category when
those axes are used, and an item has at most one group per axis. Require shared Go/TS mutation
vectors for missing/duplicate/unsorted/dangling rows and the exemption biconditional.

### R13 — The report remains unenumerated

R7 says to enumerate an envelope but supplies no fields or ordered row families. In particular it
does not place baseline/ablation pairing, purchase traces, group support, tier contributions, role
activations, beam comparison, budget totals, or failures into canonical bytes.

**Proposed contract:** report schema v1 exact root
`{schema_version,scenario_id,scenario_hash,constants_hash,relevance_policy_hash,run_budget,
greedy_oracle,items,groups,tier_contributions,role_activations,failures}`; arrays only, no maps or
binary floats. `run_budget` carries `{declared_runs,executed_runs,declared_transitions,
executed_transitions}`. `greedy_oracle` is nullable or exact
`{milestone_id,greedy_ms,beam_ms,gap_ppm,maximum_ppm,passed}`. Item rows are raw-byte sorted and carry
identity, window, epsilon, trap fields, baseline-purchase count, individual reduced deltas, support
`individual|group_supported|failed`, supporting group or null, nearest-passing epsilon, and passed
booleans. Group/tier delta rows use the R9 tagged delta row. Role rows reuse the existing ordered
`{generator_id,kind,target_id,count}` shape. Failures are sorted unique mechanical evidence strings.
Add a JSON schema plus positive fixture and mutations for every missing row family, duplicate ID,
unsorted ID, illegal null, and float.

### R14 — The run-budget equation is still not exact

R7 names six run classes but no formula. It is unclear whether `reference_greedy` is inside the
persona set, whether action removal runs for every persona or only the reference, and whether beam
node rollouts count as runs.

**Proposed contract:** let `E` be total seeds across non-reference personas, `R` reference seeds,
`I` items, and `G` groups. Full-run preflight is
`(E+R) * (1 + I + G) + R * I + (beam_enabled ? R : 0)`: baseline, effect-mask LOO, group masks,
reference-only action removal, and one beam-oracle invocation per reference seed. All multiplication
uses checked exact integers; fail before dispatch when above `relevance_budget_max_runs`.
Beam-internal work is not mislabeled as a full run and is instead bounded by R11's exact transition
budget. The report records declared and executed totals and requires equality.

### R15 — Windows cannot select milestones and trap/role evidence is ambiguous

Current harness milestones have no gate segment, so `[from_gate,to_gate)` cannot select them.
Action-removal runs are required even though the trap law is phrased only as a baseline purchase,
and the role floor does not say which run/mask supplies activation evidence. T0-T1's fail-closed
handoff has no mechanical registry path.

**Proposed contract:** relevance scenarios add a raw-byte-sorted `segments` array of exact
`{milestone_id,from_gate,to_gate}` rows, complete for relevance milestones; loader validation proves
gate IDs/order against Routes and uses the same inclusive/exclusive window rule. Trap evidence is
the reference baseline purchase count; the action-removal delta is reported diagnostic evidence but
cannot rescue a trap. Role evidence is the sum of unmasked baseline activations across all declared
personas after seed reduction; masked runs never satisfy it. The scenario registry is a checked-in
exact mapping from active schema-v4+ economy artifact path to relevance scenario, policy artifact,
and golden report. Schema-v3 is explicitly absent; on the T0-T1 mint, a missing mapping or artifact
is a `harness-check` error before dispatch.

Owner rulings must reconcile V1-V5/R1-R8, enumerate the selected wire, and either accept R10's
single-resource first-fixture narrowing or provide the missing multi-resource conversion. Until
then the solver/report implementation is blocked; the shipped ablation seam remains unchanged.

## Owner rulings on R9-R15 (2026-08-06) — the exact arithmetic/encoding/schema under R1-R8

All accepted (Codex's proposed contracts are executable and sound). Full contracts, to converge.

- **R9 — accepted; SUPERSEDES my R3 on two points I got wrong.** Every reached/unreached pair is the
  exact closed row `{status, baseline_ms, ablated_ms, delta_ms}`, status ∈ `both_reached |
  ablated_unreached | baseline_unreached | both_unreached`. `both_reached` = all three signed ints;
  `ablated_unreached` = baseline, null ablated, `delta = horizon_ms − baseline_ms` (the FINITE V5
  rule — my R3's `+∞` was wrong and violated R7's no-floats; **the finite encoding governs**);
  baseline-unreached = null delta, excluded from that milestone's floor with a NAMED report failure
  when the milestone is required. Each scenario declares exactly ONE closed reducer (`p05` OR `worst`
  — my R3's "either" allowed two legal verdicts on identical rows and is corrected); the fixture uses
  `worst`. (V5 body already carries the finite rule — consistent.)
- **R10 — accepted (the wait seam + the exact value equation).** Add harness-only
  `production.SimulateAdvance` beside `SimulateTransition`: advances a clone with NO action (same
  accrual + hook chain, returns role activations, emits no intent/receipt/revision), source-boundary
  guarded to `server/harness` + tests. Earliest affordability = the first canonical integer ms in
  `[now, next_horizon]` whose advanced clone can pay the real one-unit quote, by lower-bound binary
  search; unreachable → tagged sentinel. Value (first fixture, single-resource): the scenario
  declares one `resource_at_least` milestone and all candidate prices use that resource; advance
  baseline + candidate clones to earliest-affordable, buy one on the candidate clone, advance both to
  the next horizon; `gross_marginal_output = candidate_balance + paid_cost − baseline_balance`
  (≤ 0 → unreachable); `payback_ms = wait_ms + ceil(paid_cost * (horizon_ms − earliest_ms) /
  gross_marginal_output)` in exact Decimal with exact-integer ceiling; rank `(payback_ms,
  earliest_positive_delta_ms, raw_byte_id)`. Multi-resource/objective scoring is a named successor.
- **R11 — accepted (the bounded beam).** Fixture oracle uses the single R10 milestone. Node identity
  = `{canonical_state_hash, virtual_ms}` (path is NOT identity — final tie-break only). Expand the
  raw-byte candidate list + one edge to the next strictly-greater declared decision horizon. Score
  each child by a deterministic reference-greedy completion to milestone-or-horizon: reached sorts
  before unreached, then lower completion time, then raw-byte path. Dominance compares equal-
  virtual-ms nodes componentwise over milestone reached/progress + every canonical resource balance +
  purchased counts. Scenario declares positive-safe-int `max_decisions` + `beam_width` (fixture 8);
  stop at the terminal milestone or `max_decisions`. Internal child/rollout sims count against a
  SEPARATE `relevance_budget_max_transitions` (not the run budget).
- **R12 — accepted (`relevance_policy` grammar).** Schema v1 root `{schema_version:1, items[], groups[]}`.
  Items raw-byte sorted, COMPLETE for the generator+upgrade purchasable union, exact keys
  `{purchasable_id, availability_window:{from_gate, to_gate}, epsilon_ms, trap_exempt,
  justification_key, group_ids}`; `to_gate` + `justification_key` nullable; **`trap_exempt` iff
  `justification_key` non-null** (biconditional); `epsilon_ms` positive safe int; `group_ids` sorted
  unique. Groups raw-byte sorted `{group_id, axis, member_ids, epsilon_ms}`, axis ∈ `tier|category|
  declared`, members sorted-unique + complete for the derived tier/category when those axes are used,
  ≤ one group per axis per item. Shared Go/TS mutation vectors for missing/duplicate/unsorted/dangling
  rows + the exemption biconditional.
- **R13 — accepted (report schema v1).** Root `{schema_version, scenario_id, scenario_hash,
  constants_hash, relevance_policy_hash, run_budget, greedy_oracle, items, groups,
  tier_contributions, role_activations, failures}`; arrays only, no maps/floats. `run_budget =
  {declared_runs, executed_runs, declared_transitions, executed_transitions}`. `greedy_oracle` null
  or `{milestone_id, greedy_ms, beam_ms, gap_ppm, maximum_ppm, passed}`. Item rows raw-byte sorted:
  identity, window, epsilon, trap fields, baseline-purchase count, individual reduced deltas, support
  `individual|group_supported|failed`, supporting group or null, nearest-passing epsilon, passed
  booleans. Group/tier delta rows use the R9 tagged row. Role rows reuse `{generator_id, kind,
  target_id, count}`. Failures = sorted-unique mechanical evidence strings. JSON schema + positive
  fixture + mutations for every missing family/dup/unsorted/illegal-null/float.
- **R14 — accepted (exact run budget).** With `E` = seeds across non-reference personas, `R` =
  reference seeds, `I` = items, `G` = groups: preflight `= (E+R)*(1 + I + G) + R*I +
  (beam_enabled ? R : 0)` (baseline, effect-mask LOO, group masks, reference-only action-removal, one
  beam per reference seed). Checked exact integers; fail before dispatch above
  `relevance_budget_max_runs`. Beam-internal work is bounded by R11's transition budget, never
  mislabeled a run. The report records declared + executed totals and requires equality.
- **R15 — accepted (window→milestone + evidence ownership).** Relevance scenarios add a raw-byte-
  sorted `segments` array `{milestone_id, from_gate, to_gate}`, complete for relevance milestones;
  loader validates gate IDs/order against Routes with the same inclusive/exclusive window rule. Trap
  evidence = the reference BASELINE purchase count (the action-removal delta is reported DIAGNOSTIC
  only and can never rescue a trap). Role evidence = the sum of UNMASKED baseline activations across
  all declared personas after seed reduction (masked runs never satisfy the role floor). The scenario
  registry is checked-in; T0-1 extends it fail-closed (R8).

R1-R15 now fully ruled — solver arithmetic, encodings, both artifact schemas, budgets, and evidence
ownership are executable. Numbers (epsilon/window/thresholds) and the production scenario remain
data / T0-1 content.
