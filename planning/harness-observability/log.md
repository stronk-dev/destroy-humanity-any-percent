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
