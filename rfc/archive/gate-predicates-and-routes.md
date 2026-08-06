# RFC: Gate Predicates & the Route Registry

- **Status:** implemented
- **Author:** Marco (drafted by Claude; boundary split per Codex's 2026-07-28 review)
- **Created:** 2026-07-28
- **Design refs:** `design/08 §6` (route mechanics, as designed 2026-07-28), `design/02 §3` (Route Knowledge), `design/05 §6` (Registry + category model), `design/11 §4` (the Depletion gate), `design/10 §3b` (Doctrine exclusivity)
- **Depends on:** Save Layer (implemented — career ledger lives in Founder scope), Production Engine (implemented — gates consume its evaluated state read-only)
- **Parent / boundary split from:** `archive/production-engine-and-intents.md`
- **Planning:** `planning/archive/gate-predicates-and-routes/`

## Summary

Routes are **registered alternate preconditions for gates**. This RFC owns the predicate language, gate resolution, the Route Registry (detection, first-executor naming, the public ledger), the career execution history that gates Depletion, and Route Knowledge accounting. It exists as a separate RFC for one structural reason, stated by Codex's review and adopted: **the production engine must never interpret routes — routes touch gates only, and this boundary is a package boundary, not a convention.**

## Specification

### D1 — The predicate language is a closed union, not a formula language

A route predicate is data: a conjunction of typed conditions from a closed set — `resource_at_least` / `resource_at_most` (exact or Decimal, evaluated on committed state), `constituency_standing` / `constituency_grievance` bands (`02 §7`), `doctrine_is` / `doctrine_is_not` (`10 §3b`), `ledger_fact_present` (dated founder-ledger facts), `structure_is` (corporate wrapper), `region_trait` (`13 §5`). No arithmetic, no scripting (RFC-0002 K2 discipline). New condition kinds require an RFC amendment plus parity fixtures — predicates evaluate identically in Go and TS (client shows "route available" honestly; **the server's evaluation is the only one that counts**).

### D2 — Gate resolution

A gate (tier transition or subsystem unlock) is declared in the catalog with its **standard requirement** and zero or more **route alternatives**: `{route_id, predicate, effect}` where effect ∈ {`discount(fraction)` on the standard requirement, `substitute` (predicate satisfaction crosses the gate outright)} — **nothing else**. Effects apply at the moment of gate crossing, once, and are recorded as an event (immutable, revision-tied, per the Production RFC's event contract). Routes never write to any production slot; the routes package must not import the production package (compile-enforced, amplitude-lock pattern).

### D3 — Execution, detection, and the Registry

- **A route works whether or not anyone knows it**: the server checks all declared route alternatives at every gate crossing. Crossing via a route emits `route_executed {route_id, run_id, founder_id}`.
- **Discovery**: the first `route_executed` for a route not yet named (routes seeded by the house account carry provisional names) triggers the naming flow — **the first executor names it, permanently**, subject only to the standing name-moderation rules (`compliance.md`); the fallback name is the house seed name. The Registry is a public, dated, append-only ledger: route, predicate (revealed on registration), first executor, adoption curve.
- **Undeclared-route detection** (the Registry's second job, `05 §6`): gate crossings that satisfy no declared route but deviate from the standard requirement are impossible by construction in this model — *all* alternates are declared data. What the Registry detects instead is **novel predicate-satisfaction paths**: distinct minimal condition-sets players used to satisfy a declared predicate (e.g. two different ways to hold Regulator Standing high). These are tracked as route *variants* for the adoption ledger and Summoning-Salt retrospectives; they have no mechanical effect. **Truly emergent exploits (unintended state transitions) are bugs — they go to the audit log, and *if adopted* they are canonized as new declared routes by a balance-data change**, which is the game formalizing its own speedrun history and is the intended live-ops loop.

### D4 — Career execution history and the Depletion gate

- `route_executed` facts aggregate into the **founder-scope career ledger** (Save Layer, Founder stream): distinct routes executed, dates, runs.
- **The Depletion gate requires N distinct routes executed across the career** (N is balance data; provisional **5**, harness-gated). The catalog validator **must prove** that no single run can satisfy N routes — i.e. that Doctrine-exclusivity partitions the declared route set such that the maximum per-run subset < N. **This proof is a CI gate**: a balance change that accidentally makes the ending single-run-reachable fails the build.
- Glitchless is derived: zero `route_executed` events this run.

### D5 — Route Knowledge accounting

Earned: first-execution grant (large), repeat-execution grant (small), collapse-Exit bonus, region-draft picks (`13 §5`). Values are balance data. **Spent on exactly one thing: hints** — revealing a registered route's predicate in the map/UI. Hint state is founder-scoped and permanent once bought. Route Knowledge never enters any production or gate formula.

## Deviations from design

- `design/05 §6` and `design/08 §6` describe detecting undocumented state-transition
  sequences. This RFC does not make an unintended transition into a valid mechanic at runtime:
  undeclared transitions are audit-log bugs, and become routes only through a later declared
  catalog change. D3's variant ledger is the non-authoritative discovery surface in the interim.

## Executable contracts (answering the C1–C8 review, 2026-07-28)

### C1 — Schemas

Routes live in their own catalog family `balance/routes/*.json`, `schema_version: 1`, strict-loaded
under the economy loader's rules (unknown fields, dupes, dangling refs all fatal). **A condition
declares exactly one numeric representation — there is no runtime "either":**

```json
{"kind":"resource_at_least","resource_id":"company.cash","value":"1e9"}
{"kind":"meter_band","meter_id":"trust.regulators.standing","min":70,"max":100}
{"kind":"doctrine_is","transition":"t3_to_t4","doctrine_id":"doctrine.capture"}
{"kind":"structure_is","structure_id":"structure.nonprofit"}
{"kind":"ledger_fact_present","fact_kind":"exit.acquihire"}
{"kind":"region_trait","trait_id":"trait.lax_permits"}
```
Resource values are canonical Decimal strings; meter bands are inclusive integer bounds on 0–100
meters; everything else is mechanical IDs. Gate schema:

```json
{"gate_id":"gate.t3_to_t4","requirement":[{"resource_id":"company.permits","amount":"1e2"}],
 "routes":[{"route_id":"route.regulatory_capture","requires_context_version":2,
   "predicate":[…conditions…],"effect":{"kind":"discount","fraction":"4e-1"}}]}
```
`fraction` ∈ (0,1), canonical Decimal; the discounted amount is `Quantize12(amount × fraction)`
per requirement entry. `substitute` has no operand. Shared valid/invalid fixtures ship with the
catalog family, both runtimes.

### C2 — The `cross_gate` intent

New intent `cross_gate {gate_id, route_id|null}` under the Production C1/C3 contracts. Evaluation
order: accrue → (route path: evaluate predicate; standard path: check requirement) → **debit the
(possibly discounted) requirement through the ledger** → set `gates_crossed[gate_id]` in company
state → emit `gate_crossed` (+ `route_executed` when routed), one transaction. **A discount spends
the discounted requirement; a substitute spends nothing** — predicate satisfaction is the price,
and predicates cost elsewhere by construction. Standard requirements are resource spends *only* in
v1; anything non-resource is expressed as a predicate, so the "non-resource requirement" case
does not exist. New typed rejections (registry grows by RFC): `gate_already_crossed`,
`requirement_not_met`, `route_predicate_unmet`. **The pure evaluator is the `routes` package**
(imports: decimal + the context DTO only); the production handler imports the evaluator —
direction `production → routes`, never the reverse (the slot prohibition stands unchanged).

### C3 — The predicate context DTO

A closed, immutable struct assembled by the caller from committed state:
`{context_version, resources, doctrines_by_transition, structure_id, ledger_fact_kinds,
meter_bands?, region_traits?}` — serialized canonically for TS (all strings/ints already).
**Versioned availability instead of silent absence:** `context_version 1` carries resources,
doctrines, structure, and ledger facts (all exist in the implemented save model or land with
Prestige & Exits); meters and region traits arrive when their systems do (constituency meters →
the Trust implementation; traits → the region draft). **Catalog validation refuses any *active*
route whose conditions require a higher `context_version` than the build supports** — parity is
always tested against real committed state, never invented fixtures. A condition kind can
therefore never evaluate against a missing field: the route carrying it cannot activate.

### C4 — Cross-scope model: company event + idempotent projections

**The company stream is the sole source of truth.** `route_executed` and `gate_crossed` are
company events (atomic with the crossing, per C2). Founder career history and the public Registry
are **projections**: a projector consumes company events at-least-once and applies idempotent
upserts keyed by `event_id`. Projection runs synchronously post-commit with retry; on any read
miss, **read-repair recounts from events**. Consistency rule: a founder's own gate evaluations
(Depletion's career count) must observe all events from their own committed revisions —
read-your-writes for the founder, eventual for everyone else. `run_id = (company_stream_id,
run_seq)` with `run_seq` incremented by the Exit transaction (Prestige D3); `founder_id` is the
stream owner. Save migrations: `gates_crossed` and `run_seq` are a company save-version bump;
career tables are projections and need none.

### C5 — Global first-executor ordering (the review is right; the old rule was wrong)

Stream-local revisions cannot order a global race. **The registry decides by database
uniqueness:** `registry_routes(route_id pk, first_event_id, first_founder_id, occurred_at, name,
name_state)` — the projector's insert-if-absent wins the race; deterministic order for
simultaneous projection is `(occurred_at, event_id)` byte order. **What first execution reserves:
the naming right, for 72 h** — unexercised, the house seed name becomes permanent; exercised,
the name enters the standing moderation flow (`compliance.md`) as `pending` and publishes on
pass, house name on fail. **Executor credit is permanent regardless of naming.** The loser of the
race records as the second execution in adoption counts.

### C6 — Route Knowledge semantics (three different "firsts", named)

- **registry-first** (first execution ever): the naming right + `registry_first_bonus` (100).
- **founder-first** (this founder's first execution of this route): `founder_first_grant` (25),
  once per founder per route.
- **repeat** (subsequent executions): `repeat_grant` (5), **at most once per (run, route)** —
  `route_executed` is unique per run per route, so within-run farming is structurally impossible.
All values provisional balance data in the routes catalog. Founder save fields:
`route_knowledge_balance` (int), `hints_unlocked` (route ids); career executions are projection
data (C4). New intent `buy_route_hint {route_id}` (cost: flat `hint_cost = 50`, balance data);
rejections `insufficient_route_knowledge`, `already_unlocked`, `unknown_id`. Events:
`route_knowledge_granted {source: registry_first|founder_first|repeat|collapse_exit|region_draft}`,
`route_hint_purchased`.

### C7 — The shipped v1 route set and the proof's constraint model

**Exclusivity is declared, not inferred:** every route names an `exclusion_slot` —
`structure` (one per run, fixed at incorporation) or `doctrine:{transition}` (one per transition,
immutable once chosen, per `10 §3b`). The validator computes the exact maximum satisfiable
per-run subset by exhaustive search over slot assignments (route counts are small; exactness over
cleverness) and **fails the build if that maximum ≥ N whenever the Depletion gate is present in
the catalog**. v1 ships the three seeds expressible at `context_version 1` as *active*
(`Nonprofit Wrapper Zip` — structure substitute on the T5 gate; `IPO Sequence Break` —
doctrine+resource discount on T3; `Acquihire Out-of-Bounds` — ledger-fact discount on T3), and
the remaining four seeds as **declared-inactive** with their `requires_context_version` recorded —
in the Registry as "conjectured routes," which is itself flavor. `N = 5` stays declared; the proof
gate binds when Depletion's gate object lands. **Open dependency, recorded: BACKLOG hole #8
(doctrine coverage at three transitions) bounds the achievable exclusivity budget — those
transitions contribute none until designed.**

### C8 — Registry analytics: split out

Variants, adoption *curves*, time buckets, and the public read API move to a named follow-up,
**Registry Analytics**. This RFC keeps only: the `registry_routes` table (C5) and a plain
execution counter per route. Retrospectives and Summoning-Salt surfaces consume the follow-up.
The gate evaluator never grows analytics.

## Acceptance criteria

1. Predicate parity: shared fixtures evaluate identically in Go and TS across all condition kinds, including band edges.
2. A gate crossing via `discount`/`substitute` emits exactly one immutable `route_executed` event tied to the revision; replaying the intent does not duplicate it.
3. The routes package does not import the production package (build-enforced).
4. The Depletion-unreachability proof runs in CI: with the shipped catalog, max per-run route subset < N; a fixture catalog violating it fails.
5. First-executor race: two concurrent first executions from different company streams resolve by the registry's insert-if-absent (C5); the loser records as adoption; a replayed projection cannot double-insert.
6. Hints reveal predicates without altering evaluation anywhere.

## Open questions

- C1–C8 answered 2026-07-28 (above). Grant values, hint cost, and N remain provisional balance
  data, harness-gated — catalog locations now defined (C6/C7).
- Registry Analytics is the named follow-up carrying variants/curves/read-API (C8).
- BACKLOG hole #8 (doctrine coverage at three transitions) bounds C7's exclusivity budget.

## Changelog

- 2026-07-28: created (draft) from the route-mechanics design + Codex's boundary split.
- 2026-07-28: Codex acceptance review kept the RFC in draft and recorded eight executable-contract
  gaps: schemas, command semantics, predicate inputs, cross-scope persistence, global ordering,
  Route Knowledge, the Depletion proof, and Registry analytics.
- 2026-07-29: C1–C8 accepted by owner assignment; implementation started.
- 2026-07-29: cross-runtime predicates, exact executable Depletion proof, save v5 gate/founder
  state, authoritative intents/events, idempotent Registry projections, naming, Route Knowledge,
  canonical docs, and full verification completed; RFC archived.
- 2026-08-06: non-normative reference cleanup for publication; no spec change.
