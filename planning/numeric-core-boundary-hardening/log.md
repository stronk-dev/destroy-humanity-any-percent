# Numeric Core Boundary Hardening — Job Log

## 2026-07-27 (codex, session 1)

- Owner prioritized a perfect numeric foundation before save, ledger, currency, production,
  and domain-access systems. Treated that direction as acceptance of the drafted follow-up.
- Created the implementation plan and moved the RFC to `implementing`.
- Confirmed RFC-0002 is independent economy policy work and will not be edited here.
- Next: implement normalized state/quantization guards and their direct property tests.

## 2026-07-27 (codex, session 1 — normalized boundaries)

- TypeScript now clones and normalizes every source before quantization; arbitrary equivalent
  scientific coefficients serialize identically without mutating caller-owned Decimal objects.
- Both runtimes now require canonical zero and normalized non-zero mantissas for state validity.
- Client classification now recognizes `break_infinity.js`'s exponent-`9e15` infinity sentinel.
- Added direct regression/property tests for all three defects found in the audit. Node 6,284
  tests, strict TypeScript, and Go tests pass.
- Next: add mandatory diagnostic/domain-edge vectors and bring any revealed Go behavior into
  parity with the pinned client.

## 2026-07-27 (codex, session 1 — diagnostic parity)

- Upgraded the deterministic fixture to schema 3 with operation/classification counts and 20
  named mandatory edge cases. Tests recompute the metadata so it cannot become a stale claim.
- Edge vectors exposed pinned-client behavior not covered by RFC-0001: division by zero,
  zero-divided-by-zero, `0^negative`, infinity cancellation, and infinity-times-zero all produce
  finite zero; log/ln of zero produce negative infinity; exponent overflow produces the signed
  infinity sentinel. Ported those exact classifications/results to Go.
- Clarified the active RFC: finite-zero library compatibility does not remove the obligation for
  feature handlers to validate arithmetic domains before authorizing state transitions.
- Node and Go suites pass the expanded corpus. Next: deterministic hash, browser matrix, vet,
  fuzz, canonical docs, and archive closure.
