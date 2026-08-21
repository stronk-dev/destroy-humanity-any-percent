# Harness Observability implementation log

Append-only session record. A fresh agent should be able to resume from this file plus the RFC.

## 2026-08-21 — implementation start

- Owner authority and the accepted RFC landed at `9c71562`; the implementation range begins after
  that commit.
- Read RFC-0000, the vision/tech/content versioning constraints, CI Baseline, the archived Balance
  Harness Foundation, current balance-harness docs, R-001 diagnosis, the CLI check path, relevance
  registry loader, and relevance report work-accounting seams.
- The implementation firewall is measurement-only. No population, governed bytes, budgets,
  timeout, dispatch, topology, balance, gameplay or release claim may move in this range.
- Planned the recorder at the CLI orchestration boundary and the selector through the registered
  loader. Claude is intentionally absent until the mandatory final exact-range review.

## 2026-08-21 — instrument and negative controls implemented

- Added an atomic non-authoritative observation recorder/validator, controlled error and signal
  closeout, repository-guard/standard-pacing/registered-row objectives, and completed-run-arm
  relevance progress. Ordinary relevance report bytes remain unchanged when observation is on.
- Added `relevance-registered` selection through `LoadRelevanceRegistry` and
  `LoadRegisteredRelevanceSuite`; raw/unregistered scenario paths fail before simulation. The
  selected report is byte-compared with its registered golden and active rows retain branch and
  epoch-constants validation.
- Added opt-in root Make targets only; neither is a CI/release dependency. No governed data,
  scenario, golden, baseline, epoch, budget, timeout, worker, dispatch, topology or gameplay path
  changed.
- Cold `./harness ./cmd/balance-harness -count=1` passed (about 30 seconds), as did focused `go vet`
  and schema verification. The exact registered fixture target produced 23/23 runs and 3,324/3,324
  transitions, byte-identical to its golden.
- Negative tests reject missing/running artifacts, controlled signal termination, fired guard,
  population exclusion, truncation, run/transition mismatch, severed completion and severed active
  constants identity. The full unchanged local observation remains to be run before closeout.

## 2026-08-21 — complete local R-001 observation

- Ran `make harness-observe HARNESS_WORKERS=12` without changing the population. The command exited
  0 after 1,004.859s and the separately invoked strict validator accepted the retained
  `local-r001-observation.v1.json`.
- Exact phases: repository/history guards 23.425s; standard pacing 767.618s (300/300 runs); active
  relevance 213.514s (107/107 runs, 1,968,171/1,968,171 transitions); fixture relevance 0.203s
  (23/23, 3,324/3,324). All objectives completed with clear guard/population-exclusion/truncation
  state and epoch 8's accepted constants identity.
- The local data refutes the working cost hypothesis: standard pacing is the dominant phase. This
  does not authorize a fix or predict hosted Linux. Hosted observation and CPU/resource comparison
  remain R-001 work after implementation review and explicit push/CI authority.
- Re-ran affected cold Go tests, focused vet, schema validation, the role activation control, and
  the independent artifact validator successfully. Tracked artifacts/goldens remain unchanged.

## 2026-08-21 — internal-review correction and regenerated evidence

- Found two pre-handoff defects: objective elapsed time was reconstructed from wall timestamps
  instead of retained monotonic starts, and the artifact did not predeclare the full objective
  list. The second defect meant a severed final registry row was not mechanically distinguishable
  from a shorter declared population.
- Added the ordered declaration before work, exact declaration/completion validation, monotonic
  objective timing, directory sync after atomic rename, and a negative case that appends an
  uncompleted declared row. The fixture selector and signal/error cases remain green.
- Regenerated the complete artifact from the corrected executable. It validates at 974.510s total:
  guards 21.735s; standard pacing 739.441s (300/300); active relevance 212.927s (107/107 and
  1,968,171/1,968,171); fixture relevance 0.316s (23/23 and 3,324/3,324). All four declared rows
  completed with clear guard/population-exclusion/truncation state.
- The earlier 1,004.859s artifact was replaced rather than edited and is superseded as admissible
  final-instrument evidence. No governed input or execution policy changed between the two runs.
