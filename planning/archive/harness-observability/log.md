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

## 2026-08-21 — guard-state review correction

- Internal range review found that a returned `reference_decision_starved` or incomplete deviation
  report would fail the overall command but could close its row with `guard_state: clear` before
  final validation. Narrowly classify those two execution-guard outcomes as fired; ordinary
  relevance/content findings and a real deviation counterexample remain non-guard findings.
- Added a negative classification test and re-ran cold affected tests, vet, and strict validation
  of the retained complete artifact. Normal governed output and the retained observation are
  unchanged.

## 2026-08-21 — final validator hardening

- The same internal range review tightened the final validator around the evidence envelope: only
  complete-check and registered-row modes are accepted; timestamps must parse; elapsed values
  cannot be negative; instrument-exclusion arrays must be present and sorted/unique; complete-check
  registry indices must be contiguous and exactly one row must bind the active epoch.
- Added failing unknown-mode, negative-elapsed and missing-exclusion fixtures, quoted every Make
  path argument, and re-ran affected cold tests, vet, and validation of the retained full artifact.
  No domain report or measurement value changed.

## 2026-08-21 — signal negative-fixture race

- A later cold run fired the signal subprocess control: the helper exposed a running objective
  before installing its handler, so the parent could deliver `SIGTERM` in that gap and leave the
  last artifact `running`. The real CLI already installs the handler before declaration/work.
- Moved the helper's handler installation before its declaration so the test exercises production
  order. The fired artifact remains evidence that the old fixture was nondeterministic, not a
  product-path failure or a reason to weaken the assertion.

## 2026-08-21 — Codex first-filter review — APPROVED

- **Review by:** Codex. **Recorded by:** Codex. This is the implementer's mandatory first filter,
  not the cross-party designated archival verdict.
- **Reviewed range:** `9c71562..a9d3612` (implementation start through the last test correction).
  Range paths are limited to the Harness Observability RFC/index/plan/log/evidence, platform R-001
  records, canonical harness docs, root opt-in Make targets, and `server/harness` plus
  `server/cmd/balance-harness`. No `AGENTS.md`, balance/catalog/scenario/golden/baseline/epoch,
  workflow, deployment, gameplay, API, persistence, timeout, worker, dispatch or topology path is
  in range.
- Adversarial findings were fixed in-range: declared objective equality + monotonic timing
  (`616c58a`), fired guard classification (`b5807bf`), strict envelope validation (`9986dbe`), and
  the signal negative-fixture race (`a9d3612`). No unresolved first-filter finding remains.
- Final cold evidence: affected Go packages pass; focused signal termination passes 20/20;
  affected vet, schema, role activation, registered fixture byte equality, and the strict retained
  artifact validator pass. The corrected complete local population reached all four predeclared
  objectives at 974.510s with clear guard/population-exclusion/truncation state.
- **Verdict:** APPROVED as the first filter. The implementation remains `implementing`; it is ready
  for Claude's one mandatory cross-party review over the exact post-`9c71562` range ending at the
  handoff commit that carries this record. No archival, hosted run, D-014 ruling, push or CI change
  is authorized by this verdict.

## 2026-08-21 — Claude designated cross-party review of `9c71562..afd4fb2` — APPROVED

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.
- **Range:** `9c71562..afd4fb2`, eight commits: the owner-authorized RFC (scope ruled by Marco,
  drafted by Codex per queue row 1 — role deviation sanctioned and disclosed in the header) plus
  the implementation. This verdict covers contract text and code together.
- **H5 firewall verified:** no catalog, testdata, balance, or CI-workflow path in the range; the
  Makefile adds three measurement-only lanes explicitly absent from CI/release dependencies; every
  `relevance.go`/`relevance_solver.go` line is observation plumbing; zero test assertions removed.
- **Executed:** `./harness ./cmd/balance-harness` cold `-count=1` GREEN; the registered-row
  selector lane end-to-end against the fixture row — `authoritative: false`, full declared
  objective population, terminal `termination: objective`.
- **Forgery probe (mine):** relabelling the single-row artifact as `mode: check` is REJECTED by
  `observation-validate` (exit 1, "invalid complete-check objective order"); the honest artifact
  validates. H4's "a selected row cannot claim the complete lane" holds against an actual attack.
- **The instrument already answers R-001's question:** the recorded local observation
  (`local-r001-observation.v1.json`, 16.2 min, complete) attributes **12.3 min to
  `standard-pacing`** and only 3.5 min to the largest relevance row — the opposite of the standing
  assumption that relevance recomputation dominated. Measure-first vindicated before any hosted
  run.
- **Observation (non-blocking):** H2's torn-file/SIGTERM semantics are covered by the new test
  population; I did not SIGKILL a live run. R-001's hosted runs will exercise interruption in
  anger; a torn artifact there is an H2 finding.

**Verdict: APPROVED.** The hosted half of R-001's measurement may proceed. Per H5's own text,
neither this range nor this verdict is authority for any optimization, sharding, budget, or
CI-contract change — only measurement evidence, in its own future RFC.

## 2026-08-21 — transactional closeout

- Consumed Claude's designated cross-party approval at `96a574d` for the exact implementation
  range `9c71562..afd4fb2`; every RFC acceptance criterion and plan item is complete.
- Set Harness Observability to implemented and moved its frozen RFC and planning record into their
  archives. `docs/balance-harness.md` remains the canonical current description.
- Kept hosted Linux execution in R-001. The reviewed code is still local-only, so hosted execution
  remains blocked on explicit publication authority; no push, workflow, CI contract, timeout,
  optimization, sharding, budget, or governed input changed during closeout.
