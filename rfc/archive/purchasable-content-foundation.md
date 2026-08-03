# RFC: Purchasable Content Foundation

- **Status:** implemented
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-03
- **Design refs:** `design/02 §2` (tiered upgrades, milestone multipliers), `design/02 §11b` (the tier-relevance doctrine — THIS RFC implements its mechanics: role law, chain split, synergy schema, staggered ladders), `design/03 §12b` (scaling-seam consumers)
- **Research:** `tier-relevance.md` (all five families), `cookie-clicker.md` (tiered-upgrade grammar), `balance-enforcement.md` (the ablation boundary this RFC must expose)
- **Depends on:** Economy Kernel + Production + Run Genesis (implemented — everything lands inside `ApplyLogged`; kernel bump per KV-1)
- **Unblocks:** Relevance Harness → T0–T1 Content → Game UI (the corrected Phase-A order)
- **Planning:** `planning/purchasable-content-foundation/` (once implementing)

## Summary

The doctrine (§11b) shipped as design; its mechanics never shipped as code — the engine has
generators and manual actions, nothing else purchasable. This RFC adds the missing layer as five
catalog-driven mechanics, one save-state extension, and one instrumentation seam, all inside the
replay boundary. It corrects two false claims of mine in the T0–T1 draft ("no new engine code";
"16 milestones" — HEAD has 4 milestones × 4 persona-runs = 16 observations).

## Specification

### P1 — Upgrades as first-class catalog objects

`upgrades` catalog family: `{id, cost: {resource, amount}, window: {from_gate, to_gate|null},
requires: closed predicate (the routes/categories union, reused), effects: [closed effect union],
roles: [0+ from the §11b closed vocabulary — optional on upgrades per ruling C4], copy_key}`. Purchase via the existing intent surface
(`buy_upgrade {upgrade_id}` — C1 envelope; owned set is a save-state field `upgrades_owned`
sorted list, next save version). Effects execute ONLY through the existing contribution system:
an owned upgrade contributes declared `(slot, source_id, factor)` rows (Decimal factors only, per ruling C4) into the
deterministic-aggregation stack — no bespoke math paths; an upgrade IS a permanent contribution
bundle. **The loader rejects a roleless GENERATOR CLASS** (the §11b law, scope per ruling C4 — upgrades
declare roles optionally).

### P2 — The purchased/generated split (chain generators)

Generator classes gain optional `provisions: {generator_id, rate_ppm}` — tier-N units generate
tier-N−1 units at the declared rate during accrual (integer ppm accumulation, remainder carried,
the faction-stock arithmetic pattern). Save state splits counts: `generators_purchased` (existing
counts — pay costs, earn milestones and the per-count multipliers) vs `generators_provisioned`
(new field — free, priceless, produce only). Production reads the SUM; milestones, ladders, and
the purchase accumulator read PURCHASED only (the AD split, verbatim from §11b).

### P3 — Synergy pools

`synergy_pools` catalog family: `{id, sources: [{kind: generator|upgrade, id_or_class, per_count_ppm}],
slot, curve: linear|log}` — every declared count feeds named pools; each pool contributes one
multiplicative row into its slot (pooled-stat coupling, the Synergism architecture). Pool
composition is exported into the generated formula artifacts (the legibility law: renderable, no
opaque nesting).

### P4 — Staggered milestone ladders

Per-generator-class `ladder: [{purchased_at, multiplier_ppm}]` (catalog), applied as contribution
rows when purchased-count crosses each rung — thresholds deliberately staggered across classes
(Pecorella's rotation; values are balance data the Relevance Harness will tune).

### P5 — The ablation seam (what Relevance consumes)

One instrumentation boundary, replay-safe, AS RULED IN C10 (which supersedes this draft's
original CatalogBundle-field design): masks live outside `CatalogBundle` and `replay_inputs`
entirely — one private transition implementation takes a closed simulation policy; exported
`ApplyLogged` always supplies nil; a separate simulation entrypoint (import-guarded to
harness+tests) supplies the mask. A masked generator nulls ALL its output effects (production,
ladders, pool feeds, provisioning) while units and costs remain; a masked upgrade nulls its
contribution bundle; removed actions submitted anyway reject `unknown_id`. Real-run replay has
no mask-shaped input — leakage is structurally impossible, not asserted. This is Relevance
V5.1's `effect_mask`/`action_removal` made concrete.

### P6 — Scenario milestones

The harness scenario grows from 4 to the T0–T1 milestone set when the content RFC lands (this
RFC only ensures milestone definitions are scenario data, which they already are — recorded here
to correct my "16 milestones" conflation: 4 milestones × 4 persona-runs = 16 observations at
HEAD).

## Owner rulings on C1–C11 (2026-08-03)

**C1, C3, C6, C8, C9, C11 — ACCEPTED as proposed** (economy schema v4 owns everything, no eighth
artifact; the exact `buy_upgrade` lifecycle; per-edge remainder wire shapes with `generators`
remaining purchased-counts; cumulative strictly-increasing purchased-only ladders; state-derived
contributions computed once inside `ApplyLogged` and combined with recorded externals; the full
C11 closure enumeration binds the implementation. The schema-v4 + content change is a **MINT**.)

**C2 — accepted, with the rejection ruling:** route `Condition` union only, as an implicit
all-of; window `{from_gate: null|id, to_gate: null|id}` with the proposed semantics. **No new
terminal categories**: unmet window and unmet predicate both reject `not_eligible` with `detail`
= `"window"` or `"requires"` — the closed category registry is untouched; the detail field is
where this distinction lives, as it already does elsewhere.

**C4 — the deviations ruled:**
- Phase-0 effect union = generator-target and manual-action-target Decimal multiplier
  contributions, EXACTLY (capacity deferred to its ledger/hardcap RFC — accepted).
- **Role-law scope corrected: the §11b law binds GENERATOR CLASSES only** (design-exact). My
  draft's extension to every upgrade is withdrawn; upgrades declare roles optionally.
- **Role activation requires an executed mechanic — accepted — and therefore the Phase-0 role
  vocabulary is exactly the mechanics this RFC ships:** `provision` (chain edge), `synergy_feed`
  (pool source), `manual_output` (manual-target contribution), `stock_rate` (the existing
  faction hook). `capacity` and `minigame_input` join the vocabulary WHEN their owners land —
  the role law stays honest by construction.

**C5 — ruled: fixed-grid bucketing, not call-relative boundaries.** Provisioning quantizes to an
absolute grid aligned to `run_started_at` (`provision_tick_ms` catalog literal, 60_000 — the
stock-interval precedent): provisioned units materialize at grid boundaries and produce from the
NEXT bucket. Partition invariance holds because boundaries are absolute:
`advance(a+b) == advance(advance(a), b)` for any split, proven by golden vectors in both
runtimes. Rates read PURCHASED+PROVISIONED (the chain compounds); evaluation is bounded exact
per-bucket iteration (the existing offline catchup ceiling already caps bucket count); topology:
`tier` field required, edges point strictly one tier down, acyclic by construction, at most one
provider edge per target (Phase 0). **Provisioned counts are int64 with a declared per-class
hardcap and accrual-only saturation + reason key** — counts are counts, not Decimals; the
ceiling is announced, never discovered.

**C7 — ruled: the example contracts ARE the Phase-0 formulas.** `linear`: factor =
`1 + sum_ppm/1e6`; `log`: factor = `1 + log10(1 + sum_ppm/1e6)` through the kernel's existing
log-domain helpers, quantize-once at factor emission. Purchased counts only feed pools; pools
may NOT feed pools; every pool maps to exactly one declared multiplier source (loader-enforced).
Golden vectors include raw-byte source ordering and boundary carry, per the demand.

**C10 — accepted, with the mask semantics ruled:** masks live outside `CatalogBundle` and
`replay_inputs`; one private transition takes a closed simulation policy; exported `ApplyLogged`
supplies nil; the simulation entrypoint is import-guarded to harness+tests. **A masked generator
nulls ALL its output effects — base production, ladder factors, pool feeds, AND provisioning
edges — while its units and costs remain**: the mask measures "what does this generator's
existence contribute," whole. A masked upgrade nulls its contribution bundle. Removed actions
submitted anyway reject `unknown_id`.

**C12–C14 rulings (2026-08-03):**

- **C12 — accepted**; contradictions fixed in place (status, AC4, open questions, Decimal-only
  effect rows).
- **C13 — the proposed typed role bindings are ACCEPTED AS THE CONTRACT, formulas included:**
  roles are typed objects — `{kind:"provision", generator_id}` and `{kind:"synergy_feed",
  pool_id}` must match their executable declarations exactly; `{kind:"manual_output", action_id,
  per_purchased_ppm}` emits manual factor `1 + purchased×ppm/1e6`, quantized once, through the
  upgrades slot; `{kind:"stock_rate", per_purchased_ppm}` scales faction stock progress by the
  same factor with exact integer-ppm multiplication and a persisted remainder. Purchased counts
  only; duplicate `(kind, target)` rejects; the loader proves every role object binds to a real
  declaration; **the harness records activation only on a non-neutral executed result** — the
  self-confirming-label hole is closed.
- **C14 — accepted as proposed:** `provisioned_hardcap: {count, reason_key}` per target class;
  save v14 adds `generators_provisioned: {id: int64}` and `provision_remainders_ppm:
  {target_id: int64}` with complete key sets including zeros; per-tick semantics exactly as
  written (all edges read pre-boundary totals, overflow-safe `source_total×rate_ppm +
  remainder`, staged floor/mod commits applied simultaneously with accrual-only saturation, new
  units produce next bucket); the cap reason exports to the snapshot/formula artifacts and the
  client renders it for frozen counters (the reason-key law, applied).

## Acceptance criteria

1. Upgrade purchase round-trip: buy → contribution rows appear in the deterministic stack →
   receipt/replay byte-parity in both kernels (corpus rows); window/`requires` rejections typed.
2. Chain accrual: provisioned counts accumulate with remainder carry, byte-identical Go/TS;
   milestones and costs read purchased-only (fixture: provisioned units cross no ladder rung).
3. Synergy pools: golden vectors per curve; formula-artifact export renders pool composition;
   a two-pool fixture proves deterministic aggregation order.
4. Role law: loader rejects a seeded roleless GENERATOR CLASS (upgrades optional, ruling C4);
   loader proves every typed role object has its corresponding executable declaration (C13).
5. Ablation seam: a masked fixture produces the ablated route deterministically; production
   composition asserts nil mask; replay of a real run is bit-identical with or without the mask
   type present (proving it can't leak).
6. Save-version migration with corpus fixtures; kernel bump; KV-1 registry covers the new files.

## Open questions

None. Capacity is deferred and manual targets are included by ruling C4.

## Changelog

- 2026-08-03: created (draft) — the mechanics §11b promised; corrects the T0–T1 draft's false
  "no new engine code" claim and the milestone conflation. Implementer's order accepted:
  Foundation → Relevance → T0–T1 content → UI.
- 2026-08-03: C12–C14 ruled (typed role bindings with formulas accepted; provision wire shapes and simultaneous-commit tick semantics accepted; contradictions reconciled). IMPLEMENTATION UNBLOCKED.
- 2026-08-03: C1–C11 ruled — fixed-grid partition-invariant chains, role law re-scoped to generator classes with an honest activatable vocabulary, example synergy formulas adopted as Phase-0 contracts, import-guarded simulation entrypoint, schema-v4 mint. Implementation unblocked.
- 2026-08-03: implemented and independently approved across the complete implementation range; canonical behavior is in `docs/purchasable-content.md`.
