# Millisecond Cursor Canonicalization — Implementation Plan

- **RFC:** `rfc/millisecond-cursor-canonicalization.md`
- **Assignee:** Codex
- **Started:** 2026-07-29

## Sequence

1. Introduce the canonical UTC whole-millisecond time constructor and save-v4 validation/migration.
2. Change production and manual-token cursor advancement to the shared canonical server instant.
3. Expand the migration corpus and add the demonstrated dead-click regression, phase properties,
   and time-boundary tests.
4. Update canonical save and production documentation.
5. Run focused and complete verification, archive the RFC and planning record, and commit each
   reviewable stage locally.

## Acceptance gates

- Historical `100.9 ms` / `100.1 ms` cursors migrate to `100 ms`; a manual intent at `101.5 ms`
  applies and the resulting v4 state encodes.
- V4 round-trips exact-ms UTC cursors and rejects either sub-millisecond cursor; v1–v3 floor both
  cursors independently before order validation.
- Deterministic phase/interval properties prove `manual_token_refilled_at <= evaluated_through`
  after every legal intent sequence.
- Same-ms, 0.999 ms, exact 1 ms, rollback, and 86,400,000 ms boundaries retain specified results.
- State, receipt JSON, and migration fixtures serialize the same canonical millisecond instant.
- `docs/save-layer.md` and `docs/production-engine.md` describe save v4 and shared cursor semantics.
- `gofmt`, focused tests, and `make verify` pass.

