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

## 2026-07-29 — Completed and archived

- Closed the new-state boundary: `CreateStream` rejects a company whose two production cursors do
  not begin at the same canonical server instant. Historical mismatches remain accepted only by
  the v1–v3 migration path that repairs them.
- Updated `docs/save-layer.md` and `docs/production-engine.md` to describe save v4, strict UTC-ms
  persistence, historical flooring, shared advancement, and receipt serialization.
- Full `make verify` passed: Go vet/tests, formula drift, strict TypeScript, 6,354 Node tests,
  schema validation, and 19,062 Chromium/Firefox/WebKit tests.
- The real Postgres 16 save/intent integration suite passed against the disposable compose service;
  the service and network were removed after verification.
- All six RFC acceptance criteria are satisfied. Rotated the RFC and planning record into their
  archives. No push performed.

## 2026-07-29 (claude — per-change review of e22527b..a954448: APPROVED)

Full diff read, all four commits. The implementation matches the accepted RFC exactly, and the
two properties I derived from the RFC text at acceptance are both real and tested:

- **Root fix confirmed:** both `Evaluate` and `refillManualTokens` derive one
  `CanonicalServerTime(now)` and set their cursor **to that instant** — the floored-duration-on-
  phase-bearing-baseline pattern is gone from both sites. Ordering is preserved by construction
  (both cursors are truncations of the same monotone clock; `EncodeState` still validates R ≤ E
  as belt-and-braces).
- **Self-repair proven, not just predicted:** `TestManualIntentRepairsMigratedCursorPhaseMismatch`
  drives the *demonstrated* mismatched save (the literal `…00.1009` / `…00.1001` fixture) through
  a manual intent and asserts it heals and encodes.
- **Migration:** `restoreCursor` floors for `version < 4` and rejects non-canonical for v4 —
  exactly D2; v1's migration baseline is also canonicalized (a case the RFC didn't spell out,
  handled correctly). `formatCursor` now rejects non-canonical on encode, so a phase can't
  re-enter through any write path.
- **Corpus:** 7 cases — v1, v2, three v3 classes (phase-matched, the demonstrated mismatch,
  boundary-no-inversion), two lying-v4 rejections. Matches the RFC's required classes.
- **Property test:** 40k-step seeded ordering walk over interleaved evaluate/manual/encode ops.
- Suites green (save, production; integration per Codex's run).

**One cosmetic note, no action required:** `testdata/save-migrations.json`'s top-level
`"version": 3` is the *corpus file's* schema version sitting beside save-version-4 cases — the
name collision briefly misled this review. If the corpus schema ever revs, consider renaming the
field `corpus_version`.
