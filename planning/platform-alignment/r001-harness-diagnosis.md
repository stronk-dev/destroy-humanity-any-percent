# R-001 harness diagnosis — evidence checkpoint

Coordinate: product tree `190a4fa`; planning-only reproduction `cb162a3`; 2026-08-20.

This is a completed first diagnostic wave, not closure of R-001. It establishes the regression
window, dispatch shape, local/hosted divergence, and the instrumentation missing for a valid cost
study. It does not authorize a larger timeout.

## Question and predeclared boundaries

Question: why does the governed `balance-harness -mode=check` fail to reach a verdict inside the
30-minute hosted job budget when archived records call the lane measured and green?

Population: the complete `make verify-harness HARNESS_WORKERS=12` lane, including the harness Go
tests, the T0-T1 role proof, the 300-run production pacing suite, and both relevance-registry
entries. A killed process is an invalid measurement, never a green partial population.

Required controls:

- a current full local run must reach its actual exit;
- hosted job timestamps and the last emitted command must be read directly;
- worker-control scope must be traced in code;
- scenario-only relevance output must not be treated as registered evidence unless the active
  epoch authority binding is identical;
- no ceiling may be raised from an incomplete run.

## Verified observations

### Hosted lane

The latest run at planning commit `cb162a3` (same product tree) is Actions run `32404232364`.
Setup completed at 18:37:57Z. The harness package finished at 18:38:52Z, the role proof at
18:38:54Z, and the Commons control at 18:38:55Z. `balance-harness -mode=check -workers=12` then
emitted nothing and was killed at 19:07:59Z. The process therefore consumed 29m04s without
reaching a report or verdict before the 30-minute job ceiling.

Runs `32009994004`, `32096019304`, `32212696707`, `32328790752`, and `32404232364` repeat the
cancelled current-product conclusion. All other current-head CI jobs are green. The harness lane
alone prevents a complete current-head workflow verdict.

The last known successful hosted harness run is `31862160012` at `146deded`. Its complete
`make verify-harness` step took 6m16s. There is no Actions run for the later capacity commit
`a6328df`; its claim that the hosted lane now had sufficient headroom was never exercised on
GitHub before T0-T1 archival. Subsequent hosted runs disprove the claimed headroom.

### Current local lane

`make harness-check HARNESS_WORKERS=12` reached exit 0 locally at the current product coordinate.
The interactive elapsed time was approximately 18 minutes. The CLI has no phase timestamps, so
that elapsed observation is useful only to prove local completion below 30 minutes, not as a
phase-cost artifact or a portable budget.

This establishes an architecture/runner divergence: identical governed bytes and worker count
complete locally but exceed the hosted ceiling. It does not establish which internal phase owns
the extra hosted time.

### Workload and dispatch trace

`testdata/harness/scenarios/phase0-production.json` still declares the same 300 simulations as at
the last hosted-green coordinate: 200 Chaos runs over 30 days and 100 Casual runs over 21 days.
The workload behind each simulation changed materially. Between `146deded` and the current tree,
the production economy moved from schema v3 with two generator classes to schema v4 with the
T0-T1 catalog: nine generator classes, ladders, roles, upgrades, new routes, opportunities, and
relevance data. The principal mint is `38eaffd`.

`Suite.RunAllWithWorkers` parallelizes only the 300 standard pacing tasks. After that returns,
`cmd/balance-harness` iterates relevance-registry entries in a serial `for` loop and calls
`RunRelevance` without a worker parameter. Baseline, ablation, removal, group, and deviation work
inside a relevance row is also serial. Consequently `HARNESS_WORKERS=12` does not parallelize the
relevance phase despite governing the command as a whole.

The active relevance golden declares 107 runs and 1,968,171 transitions. The CLI exposes neither
registry-row timing nor transition progress. A job-level kill leaves no authoritative partial
artifact, exclusion count, guard state, current row, or last completed objective.

### Measurement correction caught by the controls

An initial focused `make t1-relevance` run was not the active registry row: it uses
`balance/testdata/t0-t1/relevance-scenario-t1-v2.json`, a fixture catalog/policy rather than the
production epoch inputs.

A second scenario-only run pointed at the active scenario and produced 106 runs / 1,980,281
transitions, apparently drifting from the golden. That inference is also non-authoritative. The
standalone loader derives a catalog hash, while `LoadRegisteredRelevanceSuite` replaces it with the
accepted epoch bundle hash. The hash participates in deterministic run identity, so the two loader
paths do not execute identical populations. The full registry-aware check matched its golden and
passed locally.

This is why a scenario path alone is not a valid registry-row selector. R-001 must add or use an
authority-preserving selector before recording per-row measurements.

## Findings

1. The current hosted release guard is red-by-cancellation, while repository records still
   describe the governed lane as green and adequately budgeted.
2. The 30-minute budget was justified from native/Docker evidence, not a completed hosted run.
   The capacity commit itself has no Actions verdict.
3. Production simulation cost expanded sharply without changing the top-level 300-run cardinality,
   so run count alone concealed the regression.
4. Worker control covers standard pacing but not relevance; the expensive command is only partly
   parallel.
5. The check is operationally opaque and fails the repository's fail-loud measurement law when
   externally killed.
6. Existing standalone relevance targets cannot isolate the active registered row with identical
   authority binding; using them as substitutes produces misleading identities.

## Next experiment required before a fix

The CI lifecycle audit resolved the parenthetical authority question: the active CI Baseline RFC
explicitly declares the balance harness out of scope, while the relevant harness/content RFCs are
archived. RP-059 therefore blocks instrumentation until an accepted CI amendment or active
harness-observability RFC owns it. After that authority exists, add diagnostic-only phase and
registry-row timing with:

- phase/row start and completion records;
- declared and executed runs/transitions;
- guard, exclusion, and truncation state;
- the active epoch constants identity;
- a registry-aware row selector that uses `LoadRegisteredRelevanceSuite`;
- an always-written non-authoritative termination artifact;
- tests proving a killed/guard-fired population cannot be accepted as complete.

Then run the same complete population locally and on hosted Linux. Only those measurements may
authorize algorithmic optimization, authority-preserving parallel dispatch, CI sharding, or a
budget change. Scenario removal, reduced seeds/horizons, and a timeout increase remain forbidden
responses to the present evidence.
