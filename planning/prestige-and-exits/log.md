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

## 2026-07-29 — numeric policy core

- Added a strict declarative Prestige policy and JSON Schema. The provisional P8 fixture supplies
  the `1e12` threshold, typed payout modifiers, collapse Knowledge, offer lifetime/drift, tier spawn
  gates, and Advisor constants; unknown fields fail in both the Go loader and schema gate.
- Added the P2 integer binary-search cube root in Go and TypeScript. Both consume one shared corpus
  covering zero, threshold, `n³` below/exact/above boundaries, modifier flooring, a large exact
  cube, and the numeric-domain cap. Neither runtime calls floating `cbrt`.
- Added exact moral reseeding, bounded Advisor multiplication, and the mandated SplitMix64 stream
  with fixed output vectors. Go tests, all 6,422 client tests, strict TypeScript/Svelte checks, and
  the expanded schema verifier pass.

## 2026-07-29 — atomic Exit store

- Added `ApplyExitTransaction`: it discovers the active sibling Founder stream, locks Founder then
  Company, checks both expected revisions, and writes Founder advance, final old-Company revision,
  new-run Company revision, all events, and the Company-keyed idempotency record in one transaction.
- Added the six Prestige event kinds to both the Go closed registry and migration 00011's database
  constraint. Strict payload validators cover offer transitions, complete run-end rendering facts,
  structural Assisted variables, new-run identity, and Founder payout facts.
- Real-Postgres tests inject rollback failures after each of eight write boundaries and prove both
  streams remain at revision 1 every time. The successful path advances Founder `1→2` and Company
  `1→3`; retry returns the same receipt and the identical three ordered event records without
  rerunning mutation.
- Adversarial replay review found that database UUID/timestamp order could differ from first-delivery
  group order. Replay now applies an explicit Founder-advance → run-ended → run-started order and
  the integration test compares event IDs position by position.
