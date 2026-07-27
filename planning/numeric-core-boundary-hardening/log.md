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
