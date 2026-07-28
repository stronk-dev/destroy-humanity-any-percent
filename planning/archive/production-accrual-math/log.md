# Production Accrual Math — Running Log

## 2026-07-28 — Start

- Found and restored the committed Production Engine & Intent API draft to the RFC index.
- Review found seven blocking schema gaps. Recorded them in the parent draft rather than inventing
  generator persistence, undefined intent behavior, idempotency retention, events, or Compute
  Credit limits.
- Split the already-settled pure constant-rate primitive into an accepted bounded follow-up.
- Corrected the draft summation wording before implementation: magnitude-aware ordering is by
  ascending exponent, not absolute exponent.
- No push. Concurrent uncommitted research edits remain untouched.

## 2026-07-28 — Implemented

- Added `testdata/production-accrual.json` schema 1 with 16 shared valid/invalid vectors.
- Implemented deterministic non-mutating constant-rate accrual in Go and TypeScript. Sources are
  sorted on a copy, aggregated at intermediate Decimal precision, and boundary-quantized once.
- Focused Go and TypeScript suites passed.
- `make verify` passed in 7.14 seconds: Go vet/tests, strict TypeScript, 6,338 Node tests, schema
  validation, and 19,014 browser tests across Chromium, Firefox, and WebKit.
- Updated canonical numeric documentation. The parent production RFC remains draft because its
  generator/save schema and several gameplay intent contracts are still explicit design gaps.
- No push performed.
