# Millisecond Cursor Canonicalization — Running Log

Append-only implementation record. Resume from this file, `plan.md`, and the accepted RFC.

## 2026-07-29 — Implementation opened

- Re-read the accepted RFC, its archived parents, current save/production code, migration corpus,
  canonical docs, and the review acceptance at commit `787861a`.
- Confirmed the reproduced fault remains: production and manual refill add separately floored
  durations back to phase-bearing cursors, allowing the manual cursor to overtake production.
- No design gap: save v4, historical flooring, strict v4 validation, common advancement, and the
  required boundary corpus are fully specified.
- Corrected the RFC index from stale `draft` to `implementing` as planning began.

## 2026-07-29 — Save v4 and shared advancement implemented

- Added `save.CanonicalServerTime`: UTC truncated to an exact millisecond. Encoding v4 rejects
  zero, non-UTC, or sub-millisecond cursors rather than changing caller state.
- Incremented `save.CurrentVersion` to 4. V1–v3 restoration floors both cursor instants
  independently before validating their order; v4 restoration requires already-canonical values.
- Expanded the migration corpus with v1/v2 conversion, v3 phase-matched, demonstrated
  phase-mismatched, cross-boundary, and two lying-v4 cases.
- Production evaluation and manual-token refill now derive one canonical `effective_now`, compute
  exact integer milliseconds from it, and set their cursor directly to that shared instant.
- Focused save and production suites pass.

## 2026-07-29 — Reproducer and time properties green

- Added the demonstrated v3 `100.9 ms` / `100.1 ms` reproducer through migration and a real manual
  intent at `101.5 ms`. Both cursors commit at `101 ms`, receipt JSON agrees, and v4 encoding
  succeeds.
- Added 40,000 deterministic manual-intent steps over 200 independent cursor phases, forward
  intervals, same-millisecond calls, and clock regressions. The manual cursor never exceeds the
  production cursor and every resulting state encodes.
- Added explicit 0.999 ms, exact 1 ms, same-ms, rollback, and exact 86,400,000 ms production/refill
  boundaries without changing accrual or token results.
- Full save and production package suites pass.
