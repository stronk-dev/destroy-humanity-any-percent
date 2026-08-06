# RFC: Balance Harness Foundation

- **Status:** implemented
- **Author:** Marco (drafted by Codex)
- **Created:** 2026-07-28
- **Design refs:** `design/07-roadmap.md` Phase 0 and sequencing principle 1;
  `design/02-economy-balancing.md §2, §3, §11`
- **Depends on:** Economy Kernel, Production Engine & Intent API, Geometric Affordability Fast
  Path (implemented); CI Baseline (implementing)
- **Planning:** `planning/archive/balance-harness-foundation/`

## Summary

Build the deterministic, headless simulation substrate that makes balance claims testable. This
bounded foundation runs the authoritative Phase-0 production code against versioned bot policies,
records exact distributional reports, checks a committed pacing baseline, and enters the blocking
CI budget. Later feature RFCs extend it with their actions, milestones, and scenario packs; they do
not build parallel simulators.

## Motivation

The roadmap says the harness gates every later phase, while provisional constants throughout the
implemented economy and production docs already name it as their owner. Today the closest thing is
`TestIntentPolicyPropertyTwentyFourHoursTwoHundredSeeds`: valuable, but embedded in a unit test,
limited to 24 hours, without a stable policy contract, report, baseline, or CI drift result.

The design's complete six-persona, six-month harness cannot be shipped honestly against a game
that has no Exit, tier-transition, event, route, or achievement state yet. This RFC therefore owns
the reusable runner and the executable T0 production slice. Each future mechanic adds a scenario
pack in the same RFC that adds the mechanic. The full-game pacing suite emerges by composition,
without this foundation inventing missing gameplay.

## Specification

### D1 — One authoritative transition core

- The harness is a Go package and CLI. It runs in memory: no HTTP, websocket, Postgres, wall-clock
  sleep, or goroutine-per-player actor is involved.
- Production intent validation and mutation are factored behind one exported deterministic
  transition function used by both the existing persisted `production.Service` and the harness.
  The harness may supply an in-memory revision/idempotency adapter, but **must not copy price,
  accrual, affordability, cap, receipt, or rejection logic**.
- The dependency direction is `harness → production/economy/save/multiplier`; production never
  imports harness. Harness transitions are calls into the authoritative core, not harness-owned
  arithmetic helpers.
- Policies receive a read-only observation derived from the committed canonical snapshot and the
  public catalog. They return intents, never state mutations or results. Every proposed intent is
  validated by the same authoritative transition function as a player intent.

### D2 — Run identity and deterministic execution

The complete run key is:

```text
(harness_schema_version, scenario_id, scenario_version,
 scenario_hash, policy_id, policy_version, seed, constants_hash)
```

- `seed` is an unsigned 64-bit integer serialized as a base-10 string. Random policies use
  SplitMix64 exactly: increment state by `0x9e3779b97f4a7c15`; set `z = state`;
  `z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9`;
  `z = (z ^ (z >> 27)) * 0x94d049bb133111eb`; return `z ^ (z >> 31)`. Every operation wraps as
  `uint64`. For a bound `n`, reject draws below `(-n) % n`, then return `draw % n`; an uncorrected
  modulo reduction is forbidden. No standard-library global PRNG is permitted.
- Virtual time is exact integer milliseconds from `2000-01-01T09:00:00Z`, so the Casual policy's
  first active session starts with the run. The runner advances
  only to a policy wake time, session boundary, or declared simulation boundary. Production still
  receives a `time.Time`, derived exactly from that cursor; no client timestamp enters an intent.
- Action candidates use the v1 kind order `perform_manual_batch`, then `buy_generator`, followed
  by raw-byte ascending mechanical ID. An RFC adding a candidate kind appends or amends this order
  explicitly. Any other reduction sorts by the complete run key. Parallelism is across runs only;
  a run is single-threaded. Collection writes into seed-indexed slots before ordered reduction.
- The runner supplies the current revision as `expected_revision`. Intent IDs are deterministic
  UUIDv7 values: the high 48 bits are the virtual Unix millisecond modulo `2^48`; the remaining
  random bits come in order from a separate SplitMix64 stream initialized with
  `seed ^ 0xd1b54a32d192ed03`, with the UUID version/variant bits overwritten as required. IDs are
  unique within a run; generation fails on collision instead of retrying invisibly.
- Ordered paths iterate catalog slices, never Go maps. Report values are integer counts/times or
  canonical Decimal strings; the committed report contains no binary floating-point values.
- The catalog's existing `save.ConstantsHash` over the exact catalog bytes is the
  `constants_hash`. The scenario file has its own SHA-256 over its exact bytes in the report, so
  changing a target or policy schedule cannot masquerade as a balance change.

The canonical report encoding is UTF-8 JSON with fixed struct field order, two-space indentation,
LF line endings, one final LF, arrays sorted as above, and no maps in serialized report types.
Seed 0 is the golden run. Identical run keys must produce byte-identical golden reports on
linux/amd64 and darwin/arm64.

### D3 — Policy contract and Phase-0 policies

A versioned policy owns its cadence and decision rule; personas are code objects, not flags. The
foundation registers exactly these two policies:

1. `casual.phase0` v1 — three eight-minute active sessions per virtual day, beginning at 09:00,
   14:00, and 20:00 UTC. During a session it acts once per second. It buys one unit of the
   cheapest affordable generator (quoted authoritative cost; ties by raw-byte ID); if none is
   affordable it submits the catalog's first manual action with count 1. It accepts an offered
   Exit only after that intent exists in a later scenario extension. Time inside a session is
   evaluated as online; the gap before the next session is evaluated once as offline when that
   session begins, through the production engine's existing mode boundary.
2. `chaos.phase0` v1 — acts every five virtual minutes. It chooses uniformly from the sorted
   syntactically valid intent templates and then chooses any required positive count uniformly in
   `[1, 80]`; `buy_generator` chooses uniformly between exact count and `max`. Rejections are
   recorded and simulation continues. It is continuously online despite the five-minute decision
   cadence, has no privileged state, and gets no forced first action.

The candidate provider initially exposes the implemented `perform_manual_batch` actions and one
`buy_generator` template per catalog generator. An RFC that adds an intent must explicitly decide
whether and how it enters this provider. A new policy ID/version requires an RFC or an amendment to
an unimplemented RFC; changing behavior under an existing version is forbidden.

The Speedrunner, Optimizer, Idler, Lapser, and full Casual persona policies are named
successors, not aliases for the two policies above. They land when Exit/session/catch-up/route
state gives their distinctions executable meaning.

### D4 — Scenario and envelope schema

Checked-in scenarios live under `testdata/harness/scenarios/`; their schema is
`testdata/harness/scenario.schema.json`, the report schema is
`testdata/harness/report.schema.json`, and both are validated by `make verify-schema`. A v1
scenario contains:

```json
{
  "schema_version": 1,
  "id": "scenario.phase0_production",
  "version": 1,
  "catalog": "balance/catalogs/phase0.json",
  "runs": [
    {"policy_id": "chaos.phase0", "policy_version": 1,
     "seed_start": "0", "seed_count": 200, "horizon_ms": 2592000000},
    {"policy_id": "casual.phase0", "policy_version": 1,
     "seed_start": "0", "seed_count": 100, "horizon_ms": 1814400000}
  ],
  "milestones": [],
  "envelopes": [],
  "required_invariants": []
}
```

The checked-in object is populated during implementation; the empty arrays above illustrate the
shape, not accepted content. Closed milestone kinds in v1 are:

- `intent_applied {intent_kind, count}`;
- `event_seen {event_kind, count}`;
- `resource_at_least {resource_id, amount}`;
- `generator_count_at_least {generator_id, count}`;
- `progress_at_least {tier, amount}`.

Every milestone has a mechanical `id` and `must_reach` boolean. Its first satisfied virtual
millisecond is recorded; never reached is JSON `null`. Envelope statistics are `best`, `p05`,
`p50`, `p95`, or `worst`, with optional inclusive `minimum_ms`/`maximum_ms`. Percentiles use the
nearest-rank definition over reached integer times after sorting; an unreached `must_reach`
milestone fails before percentile evaluation.

The shipped Phase-0 scenario declares at least: first applied manual action, first generator
purchase, generator count 1, and T0 progress 1. Its Casual envelope requires first generator
purchase below 60 seconds and generator count 1 below five minutes, matching the roadmap's
first-hook/automation targets. Its Chaos run must complete the 30-day horizon with every invariant
green. Future tier RFCs own the target values they append; harness code never switches on a
particular generator, tier name, or flavor string.

### D5 — Exact reports and correctness invariants

Each run report contains its run key, scenario hash, outcome, first-milestone times, final virtual
time, final canonical state hash, applied/rejected counts by kind/category, exact source and sink
totals by resource, final balances, maximum gap between declared progression events, and an
ordered invariant-failure list.

The v1 invariant registry is closed:

- `state_encodes` — every applied transition passes `save.EncodeState`;
- `numeric_domain` — every persisted Decimal is canonical, finite, and within the state domain;
- `resource_bounds` — no resource is below its minimum or above its hardcap;
- `ledger_reconciles` — every receipt's canonical `before` matches the preceding committed balance,
  its canonical `after` matches the resulting committed balance, and every changed resource appears
  exactly once. The 12-digit explanatory `delta` is domain-valid and signed for source/sink totals,
  but is not required to re-add exactly across lossy cancellation where no 12-digit delta exists;
- `revision_monotone` — applied transitions advance once, terminal rejections do not;
- `must_reach` — every required milestone is reached within the scenario horizon.

Invariant failure stops that run, but the suite completes and reports all failed run keys. A
panic, timeout, unknown policy, unknown milestone kind, missing catalog, or malformed report is a
harness failure, never a skipped run. `soft_lock` is not guessed by heuristics in v1: it is the
explicit failure of a scenario's `must_reach` milestone. Feature RFCs add more precise liveness
milestones as their recovery paths become real.

### D6 — Baseline and drift

`testdata/harness/golden-seed.json` is seed 0's canonical run report and
`testdata/harness/pacing-baseline.json` is the canonical aggregate report; both are committed.
Aggregate distributions are reported, never arithmetic means. Drift compares matching
`(scenario, policy, milestone, statistic)` integer times:

```text
abs(current_ms - baseline_ms) / max(1, baseline_ms)
```

The ratio is compared by cross-multiplied integers, not float. Drift over 10% is printed as a
warning; drift over 25% fails. An intentional change is made by regenerating the baseline in a
separate, reviewable commit whose subject begins `BALANCE-CHANGE:`; CI verifies that any baseline
change has a changed catalog/scenario hash and refuses a baseline-only rewrite. The baseline is a
source artifact, never a CI cache.

### D7 — CI tiers and artifacts

- `make harness HARNESS_OUTPUT=/path/report.json` writes a report to that required explicit path.
- `make harness-check` runs the Phase-0 suite, compares the committed golden/baseline files, and
  changes no tracked files.
- `make harness-update` regenerates those files for deliberate review; it is never called in CI.
- `harness-check` joins the blocking server job and has a 60-second local/hosted target. The whole
  blocking workflow retains CI Baseline's normative five-minute ceiling.
- The blocking run is T1 correctness (Chaos 200 × 30 virtual days) plus the current near-horizon
  envelopes. Far-horizon T3/T5 envelopes run on `main`; N=1000, search, sensitivity, and diversity
  are nightly successors and never block a pull request.
- The CLI always emits raw canonical JSON. A graph renderer/PR comment is a presentation
  follow-up; failure to render a picture must never hide a failed numeric gate.

### D8 — Objective-function and extension boundary

The v1 harness has no optimizer. Its schema accepts pacing-envelope targets only and has no field
for retention, session length, return probability, monetization, or engagement. Any future search
RFC must preserve that closed target vector and unit-test that its loss contains only declared
pacing-envelope error, explicitly naming the rejected EOMM objective.

Each gameplay RFC that changes pacing must provide, in the same implementing change:

1. any new deterministic action-candidate adapter;
2. milestone or progression-event registrations;
3. at least one scenario fixture exercising the mechanic;
4. an intentional baseline update if its accepted target moves.

The harness owns orchestration and measurement, not mechanics or target taste.

## Deviations from design

- The pacing design proposes six full-game personas and T0–T7 gates in one harness. This foundation
  ships two meaningful policies and the implemented T0 slice; policies and far-horizon packs land
  with the mechanics they need. This is a scope split, not a rejection of the six-persona target.
- The far T3/T5 envelopes move from every pull request to `main`, adopting the measured
  CI-horizon correction. Near envelopes still block every pull request.
- The design allows a drift over 25% to pass when a PR declares `BALANCE-CHANGE:`. This RFC makes the
  declaration a committed baseline update with a matching catalog/scenario hash, so the same rule
  works locally and does not depend on GitHub metadata.

## Acceptance criteria

1. The harness and persisted service call the same authoritative production transition; a mutation
   fixture produces identical receipt and canonical state in both adapters.
2. Seed-0 golden output is byte-identical on darwin/arm64 and linux/amd64; repeated and parallel
   runs produce the same bytes and aggregate ordering.
3. Chaos 200 × 30 virtual days has zero invariant failures, and a seeded invalid-state fixture is
   reported with the exact failing run key.
4. The shipped Casual distribution reaches its first generator below 60 seconds at p50 and count 1
   below five minutes at p95; all shipped `must_reach` milestones succeed.
5. A fixture above 10% drift warns without failing, one above 25% fails, an unreachable milestone
   fails, and a baseline-only rewrite fails validation.
6. Scenario and report schemas reject unknown fields/kinds and unsafe integer seeds; all checked-in
   scenarios and the baseline pass `make verify-schema`.
7. `make harness-check` is read-only, completes within 60 seconds on the hosted runner, and keeps
   the complete blocking workflow below five minutes.

## Named follow-ups

- T0–T1 content + Exit scenario pack (with the first-prestige 45–90 minute gate).
- Full persona pack (Speedrunner, Optimizer, Casual, Idler, Lapser, Chaos) as the required action
  surface lands.
- Balance Epoch artifact + development hot reload/production immutability.
- Far-horizon pacing, strategy-diversity, optimizer, Sobol/identifiability, and PR visualization.
- Player-facing golden-run verification and leaderboard epoch binding.

**Absorbed from the parallel draft `balance-harness.md` (deleted 2026-07-28 — drafted concurrently, this foundation landed first and is execution-grade):** the **contributed-gates registry** — feature RFCs lodge named scenario packs here (anti-Nash-1 from Combat AC3, Depletion-unreachability from Gate Predicates, commons population-invariance from Commons AC6) and the harness runs them in the nightly tier; **best-of-N gating** for skill-sensitive envelopes (Roohi et al. — gate on best-of-N runs, never persona means); and the explicit rule that **changing a persona policy is itself a `BALANCE-CHANGE:`-class event**, because it silently moves every envelope.

## Open questions

None for this bounded foundation. Later target values are not silently selected here; they belong
to the named scenario-pack follow-ups above.

## Changelog

- 2026-07-28: created as the executable Phase-0 harness foundation; split unavailable full-game
  personas and far-horizon systems into named extensions.
- 2026-07-29: implementation started after Production Contract Assertions & Integrity archived.
- 2026-07-29: deterministic transition, policies, reports, schemas, baseline/drift gates, CI wiring,
  cross-architecture golden proof, and canonical docs completed; RFC archived.
- 2026-08-06: non-normative reference cleanup for publication; no spec change.
