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

