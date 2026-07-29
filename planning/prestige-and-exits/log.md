# Prestige & Exits — append-only log

## 2026-07-29 — start

- Accepted through `planning/codex-batch-2026-07-29.md` after P1–P8 answered the implementation
  review. The first timing envelope is elective; the scripted first Exit remains a fixed curriculum
  segment.
- Implementation follows the accepted seam: save v7 first, pure cross-runtime arithmetic second,
  multi-stream storage third, then Production intents and deterministic evaluation hooks.
- `founder seed` in P3 is not a P1 persisted field. The implementation will derive the stable
  unsigned seed from the immutable Founder UUID and `run_seq`, and records that literal mechanical
  interpretation here rather than adding an undeclared save field.
- Phase-0 has no harvested-Quarter intent, IPO chain, carried Network content, or Reputation tree.
  Those accepted fixture seams remain empty/gated; threshold-crossing offers, Wind Down, and the
  full transaction are implemented without inventing content.

## 2026-07-29 — save v7

- Added the exact P1 Company and Founder state fields plus P6's ordered offline spans. All times are
  integer milliseconds at rest and canonical UTC `time.Time` values in memory; prestige integers
  stay inside the exact-number domain.
- Validation rejects cross-scope leakage, malformed offers, duplicate/unsorted Network slots,
  non-append-only Exit histories, overlapping spans, unsafe timestamps, and non-finite or negative
  lifetime value. The list cap is 256 entries; the evaluation layer owns oldest-span collapse.
- Version 1–6 migrations receive explicit zero/null/empty defaults. The only-grows corpus now has a
  Founder-v6 fixture as well as the existing Company history. New account-created company runs set
  `run_started_at` from canonical server time.
- `go test -p 1 ./...` passes across every server package after the save bump.
