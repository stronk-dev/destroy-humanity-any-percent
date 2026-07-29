# Balance Harness Foundation — running log

Append-only. Resume from this file and `plan.md`.

## 2026-07-29 (Codex — start)

- Production integrity follow-up is implemented, fully verified, reviewed, and archived.
- Started the already-accepted Balance Harness Foundation immediately per owner direction.
- No push is authorized; intermediate commits remain local.

## 2026-07-29 (Codex — reconciliation finding)

- The first small deterministic run failed `ledger_reconciles`: a combined intent receipt delta
  re-applied one canonical ulp below its authoritative `after` value. State and component ledgers
  were correct; `production.wireChanges` recomputed a naive quantized subtraction.
- Direct probing established that no 12-digit canonical delta exists for some valid combined
  before/after pairs; expanding the accrual-only ±2-ulp search cannot solve a representational
  impossibility. Amended the active RFC's invariant to reconcile the authoritative canonical
  before/after chain while retaining signed canonical deltas for aggregate reporting. The shipped
  receipt and accrual semantics remain unchanged.

## 2026-07-29 (Codex — harness foundation implementation)

- Exported `production.ParseIntent` and the pure `production.Transition`; persisted service and
  harness now call the same mutation core. A parity fixture locks identical receipt and state.
- Implemented pinned SplitMix64 with rejection sampling, a separate deterministic UUIDv7 stream,
  exact virtual time, fixed action ordering, four-worker seed-indexed execution, and complete-key
  reduction. Serialized reports contain structs/slices of integers and canonical strings only.
- Implemented Casual v1 and Chaos v1, the full 300-run Phase-0 scenario, four milestone kinds in
  use, exact before/after receipt-chain validation, state/domain/bounds/revision/must-reach
  invariants, ordered reports, nearest-rank envelopes, and integer-only drift comparison.
- Added scenario/report schemas plus two negative fixtures, complete-run CLI output, read-only check
  and deliberate update targets, initial seed-0 golden and pacing baseline, and the baseline commit
  guard. CI fetches two commits so the guard cannot be bypassed by a shallow checkout.
- `make harness-check` is green and completes in about 47 seconds locally, below the normative
  60-second blocking target. The 300-run aggregate has zero failures; Casual p50/p95 automation is
  10 seconds.
