# RFC: Gate Predicates & the Route Registry

- **Status:** draft
- **Author:** Marco (drafted by Claude; boundary split per Codex's 2026-07-28 review)
- **Created:** 2026-07-28
- **Design refs:** `design/08 §6` (route mechanics, as designed 2026-07-28), `design/02 §3` (Route Knowledge), `design/05 §6` (Registry + category model), `design/11 §4` (the Depletion gate), `design/10 §3b` (Doctrine exclusivity)
- **Research:** `design/research/speedrun-governance.md §3.5`, `design/research/tile-placement.md` (draft picks as Route Knowledge)
- **Depends on:** Save Layer (implemented — career ledger lives in Founder scope), Production Engine (implemented — gates consume its evaluated state read-only)
- **Parent / boundary split from:** `archive/production-engine-and-intents.md`
- **Planning:** `planning/gate-predicates-and-routes/` (once implementing)

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

## DESIGN-GAPs blocking acceptance (Codex review, 2026-07-28)

The system boundary is sound, but the draft does not yet authorize implementation. These are
contracts the RFC owner must resolve; an implementation must not choose defaults for them.

### C1 — Catalog and wire schemas

D1 names condition kinds but does not define their tagged JSON shapes, identifiers, numeric kinds,
band-edge encoding, or invalid-value rules. D2 likewise has no typed schema for a gate's standard
requirement, so `discount(fraction)` has no specified operand, rounding rule, or valid range.
Define the catalog-version bump, complete condition/effect/gate schemas, uniqueness and reference
checks, canonical ordering, and shared valid/invalid fixtures. Clarify whether "exact or Decimal"
means a condition chooses one declared numeric kind or accepts either representation at runtime.

### C2 — The authoritative gate-crossing command

The implemented Production API intentionally contains only `buy_generator` and
`perform_manual_batch`; there is no tier/unlock state or crossing command. Specify the new intent
envelope, receipt, closed rejection extension, evaluation order relative to lazy accrual, standard
requirement spend, the exact state mutation that marks a gate crossed, and the new event payloads.
In particular, define whether a discount spends the discounted requirement and how a substitute
interacts with non-resource requirements. If transition state belongs to another RFC, split the
pure predicate evaluator from its gate integration explicitly.

### C3 — Predicate input state

Most D1 inputs do not exist in the implemented save model: constituency bands, doctrines,
founder-ledger facts, structure, and region traits. Define a closed, immutable predicate-context
DTO (including scope and numeric representation), the authoritative source of every field, how a
missing/unavailable field evaluates, and the exact projection sent to TypeScript. Otherwise Go/TS
parity can be tested only against invented fixture state, not the committed game state.

### C4 — Cross-scope atomicity and persistence

One execution currently needs to mutate or project across company state, founder career history,
and the public Registry, while `save.Store.ApplyIntent` locks and revises exactly one stream.
Choose the source of truth and transaction shape: atomic multi-stream mutation, immutable company
event plus idempotent projections, or another explicit model. Define `run_id`/`founder_id`
ownership, projection retry and deduplication, save-version migrations, and the consistency rule
used when Depletion reads career history.

### C5 — Global first-executor ordering and naming

"Earliest committed revision" cannot order executions from different company streams because
their revisions are local. Define the globally comparable database order and uniqueness rule,
what the first execution actually reserves, the naming deadline/fallback, moderation state, and
the relation between a house provisional name and the executor's permanent name. Concurrent
execution and concurrent naming fixtures must follow from that model.

### C6 — Route Knowledge and hints

D5 lacks executable currency semantics. Define whether first/repeat is per founder, per run, or
global; the catalog objects and provisional values; the hint price rule; founder save fields; the
purchase intent/receipt/rejections; and grant/spend event payloads. A route execution, a Registry
first execution, and a founder's first execution are different facts and must not share the word
"first" without qualification.

### C7 — The Depletion proof and shipped route set

No concrete gate/route catalog exists from which acceptance criterion 4 can be proved, and
`design/BACKLOG.md` still records missing Doctrine coverage at three transitions. Check in the
seed routes, their gate alternatives, Doctrine/structure constraints, and N, then define the
constraint model and validator algorithm that computes a maximum satisfiable per-run subset.
Doctrine labels alone are insufficient if a run may choose a different Doctrine at each
transition or if non-Doctrine predicates conflict.

### C8 — Registry variants and adoption read model

D3 asks for "distinct minimal condition-sets" although D1 predicates are conjunctions and no
provenance model says which underlying actions satisfied a condition. The clustering granularity
is also left open. Either move variants, adoption curves, public querying, and retrospectives to a
named Registry Analytics follow-up, or define the canonical signature, storage schema, time
buckets, and read API here. This non-mechanical surface must not silently grow the gate evaluator
into an analytics system.

## Acceptance criteria

1. Predicate parity: shared fixtures evaluate identically in Go and TS across all condition kinds, including band edges.
2. A gate crossing via `discount`/`substitute` emits exactly one immutable `route_executed` event tied to the revision; replaying the intent does not duplicate it.
3. The routes package does not import the production package (build-enforced).
4. The Depletion-unreachability proof runs in CI: with the shipped catalog, max per-run route subset < N; a fixture catalog violating it fails.
5. First-executor naming: concurrent first executions resolve deterministically (earliest committed revision wins); the loser's execution still records.
6. Hints reveal predicates without altering evaluation anywhere.

## Open questions

- C1–C8 above block acceptance.
- N=5 and all grant values may remain provisional and harness-gated once C6/C7 define their
  catalog locations and checked-in fixtures.
- Route-variant clustering must be specified or split per C8 before acceptance; it is not
  implementation freedom while the Registry remains in this RFC's normative scope.

## Changelog

- 2026-07-28: created (draft) from the route-mechanics design + Codex's boundary split.
- 2026-07-28: Codex acceptance review kept the RFC in draft and recorded eight executable-contract
  gaps: schemas, command semantics, predicate inputs, cross-scope persistence, global ordering,
  Route Knowledge, the Depletion proof, and Registry analytics.
