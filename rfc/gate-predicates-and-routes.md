# RFC: Gate Predicates & the Route Registry

- **Status:** draft
- **Author:** Marco (drafted by Claude; boundary split per Codex's 2026-07-28 review)
- **Design refs:** `design/08 §6` (route mechanics, as designed 2026-07-28), `design/02 §3` (Route Knowledge), `design/05 §6` (Registry + category model), `design/11 §4` (the Depletion gate), `design/10 §3b` (Doctrine exclusivity)
- **Research:** `design/research/speedrun-governance.md §3.5`, `design/research/tile-placement.md` (draft picks as Route Knowledge)
- **Depends on:** Save Layer (implemented — career ledger lives in Founder scope), Production Engine (draft — gates consume its evaluated state read-only)
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

## Acceptance criteria

1. Predicate parity: shared fixtures evaluate identically in Go and TS across all condition kinds, including band edges.
2. A gate crossing via `discount`/`substitute` emits exactly one immutable `route_executed` event tied to the revision; replaying the intent does not duplicate it.
3. The routes package does not import the production package (build-enforced).
4. The Depletion-unreachability proof runs in CI: with the shipped catalog, max per-run route subset < N; a fixture catalog violating it fails.
5. First-executor naming: concurrent first executions resolve deterministically (earliest committed revision wins); the loser's execution still records.
6. Hints reveal predicates without altering evaluation anywhere.

## Open questions

- N=5 and all grant values: provisional, harness-gated.
- Route-variant clustering granularity (D3): implementation freedom, no gameplay effect.

## Changelog

- 2026-07-28: created (draft) from the route-mechanics design + Codex's boundary split.
