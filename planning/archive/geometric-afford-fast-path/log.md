# Geometric Affordability Fast Path — Running Log

## 2026-07-28 — Start

- Owner directed work to proceed and selected a public repository for future hosted CI.
- Reviewed RFC-0000, the implemented numeric/economy docs and code, both new drafts, and the CI
  research findings.
- Split this repair from the CI draft because it amends implemented RFC-0002 behavior.
- Corrected the research summary: the isolated microbenchmarks imply about 31×, while the broader
  harness measurements imply about 95×; they are different scopes.
- Identified the critical semantic boundary: `decimal.AffordGeometricSeries` can return up to
  `MaxExactInteger`, while the public economy query must cap purchases at
  `MaxExactInteger - owned`.
- No design gaps remain for this bounded repair.

## 2026-07-28 — Implementation and focused verification

- `economy.MaxAffordable` now dispatches geometric curves to the Decimal closed-form helper,
  clamps to remaining safe-integer capacity, verifies through public `BulkCost`, and falls back to
  the extracted bounded search if local correction cannot establish both postconditions.
- Constant and linear curves retain the generic search.
- Added zero-cash, ratio-one, huge-exponent, ceiling-adjacent, deterministic generated, and
  constant/linear regression coverage.
- Focused economy tests pass. The in-test benchmark measured 3,895 ns/op public versus
  2,504 ns/op helper (1.56×) on one calibration run.
- Three explicit benchmark runs measured the public path at 3,979–4,207 ns/op and the helper at
  2,244–2,325 ns/op, with zero allocations. The RFC's same-run ratio guard is below 10×.

## 2026-07-28 — Acceptance and handoff

- Full `make verify` passed: Go vet and tests, strict TypeScript, 6,321 Node tests, and 18,963
  browser tests across Chromium, Firefox, and WebKit.
- The first sandboxed browser attempt could not bind an IPv6 localhost port (`EPERM`); the same
  command passed with permission to run the local Playwright server. This was an environment
  restriction, not a product failure.
- Updated `docs/economy-kernel.md` with the implemented fast path, exact ceiling, verification, and
  fallback behavior.
- All RFC acceptance criteria are satisfied. RFC and planning records archived.
