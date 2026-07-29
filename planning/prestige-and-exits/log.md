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
- `lifetimeValue` is named but its resource valuation is not defined in the RFC. The Phase-0 policy
  therefore declares `value_resource_id=company.cash`; authoritative positive accrual in that
  resource advances lifetime value. Future multi-resource valuation is balance data, not a hidden
  conversion invented in engine code.

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

## 2026-07-29 — lifecycle runtime

- Implemented deterministic threshold-site offer generation, persisted preview terms and market
  modifier, typed expiry/decline, promise-preserving acceptance, elective Wind Down, and a gated
  IPO seam. Offer identity and PRNG draws derive only from immutable Founder/run inputs.
- Production accrual now advances the declared lifetime-value resource and records server-derived
  offline spans. The scripted-first path fires on the first eligible threshold crossing, commits
  through the same atomic Exit operation, and is prevented by append-only Founder history from
  repeating.
- New-run assembly is byte-deterministic and carries only declared Network/Reputation seams before
  reseeding moral bands. No Phase-0 fixture content was invented for those still-empty seams.
- `run_ended` now reads executed routes from canonical Company events through the request's expected
  revision. If another route commits concurrently, the Exit CAS fails; facts from the wrong
  revision cannot enter the obituary. The real-Postgres test asserts preservation.
- The event currently carries the terminal Company revision as `terminal_seq`. This is deliberately
  provisional under P7 and is replaced immediately by Leaderboards L1's transaction-local run-log
  sequence; Prestige cannot archive before that integration lands.
- **DESIGN-GAP:** D5 names Advisor Mode as a Founder-scoped toggle, but D1's closed intent union does
  not provide a toggle intent. The persisted field, multiplier semantics, and structural assisted
  fact exist; no undeclared player command was invented. The independent review/owner must either
  add a closed intent or confirm that another accepted RFC owns the control surface.
- AC8 remains blocked by this RFC's explicit T0–T1 playable-content dependency. The current harness
  has no eligible elective Exit policy and cannot honestly measure a 45–90 minute first elective
  Exit. The transaction/runtime hook is implemented; no synthetic balance milestone was added and
  presented as evidence.
- Focused verification is green: all Go packages; strict TS/Svelte checks; 6,422 client tests;
  client shell boundary lint; real-Postgres elective, scripted, preview-promise, route-fact, eight
  fault-boundary, and replay integration cases.

## 2026-07-29 — Leaderboards L1 integration

- The provisional terminal Company revision has been replaced by the Exit intent's transaction-local
  run-log sequence. `run_ended.terminal_seq` and the committed `run_log.seq` are asserted equal in
  the real-Postgres lifecycle test, closing P7 without a second write path.
