# RFC: Balance Harness Dispatch Integrity

- **Status:** draft
- **Author:** Codex
- **Created:** 2026-08-03
- **Design refs:** `design/02 §11` (balance simulation as an acceptance test), `design/02 §11b`
  (relevance enforcement)
- **Depends on:** Balance Harness Foundation (implemented)
- **Parent / amends:** `rfc/archive/balance-harness-foundation.md` D2 and D7
- **Supersedes / superseded by:** —
- **Planning:** `planning/balance-harness-dispatch-integrity/` (once implementing)

## Summary

Restore the Balance Harness Foundation's one-task/one-slot execution contract before the
Relevance Harness multiplies the run set. The current `RunAll` producer submits every task index
twice. Two workers can therefore execute the same run concurrently and write the same report slot,
while the suite silently performs twice the declared work.

## Motivation

The archived foundation requires parallelism across runs, one single-threaded execution per run,
and collection into seed-indexed slots. Duplicate dispatch violates all three operational claims:
it creates an unsynchronized same-slot write, doubles the cost budget, and makes declared run count
different from executed run count. Deterministic runs often write equal bytes, which masks the
defect in ordinary baseline comparisons.

This follow-up changes harness orchestration only. It does not alter gameplay transitions,
personas, catalogs, pacing targets, report schemas, run keys, or committed balance artifacts.

## Specification

### D1 — Exact task cardinality

For the ordered task list derived from scenario run specifications, `RunAll` dispatches each task
index exactly once. Every index has exactly one writer for its preallocated report slot. Worker
count remains bounded and results remain reduced by the complete run-key order.

The implementation exposes a package-private execution seam used by the runner and tests. The seam
may wrap `suite.run`, but it may not permit policies or callers to mutate state or bypass the
authoritative Production transition.

### D2 — Fail-closed collection proof

A test executor records invocation count by complete run key. A duplicate key, missing key,
out-of-range slot, or more/fewer executions than the scenario's declared total is a test failure.
The production runner continues to return one report per declared task; no deduplication pass may
hide duplicate execution after the fact.

The focused harness suite runs under Go's race detector. A test must exercise at least two workers
and more tasks than workers so the worker-pool handoff is real.

### D3 — Artifact and budget invariance

The Phase-0 golden report and pacing baseline remain byte-identical. No `BALANCE-CHANGE:` or kernel
version bump is warranted because authoritative transition semantics and report values do not
change. `harness-check` must execute the declared number of runs and remain inside the existing
60-second harness budget; the repair should reduce elapsed work, not move the limit.

## Deviations from design

None. This restores the already-accepted harness execution contract.

## Acceptance criteria

1. A multi-worker cardinality fixture proves each complete run key executes exactly once, with no
   missing or duplicate task and one result per declared task.
2. The focused harness tests pass under `-race`.
3. `make harness-check` passes without changing either committed harness artifact.
4. `make verify` passes and the harness remains below its existing 60-second budget.

## Open questions

None.

## Changelog

- 2026-08-03: created after the Relevance Harness acceptance pass found duplicate dispatch in the
  implemented foundation.
