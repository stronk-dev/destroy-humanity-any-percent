# Production Hardcap Saturation — Implementation Plan

- **Status:** completed 2026-07-29
- **RFC:** `rfc/archive/production-hardcap-saturation.md`
- **Assignee:** Codex
- **Started:** 2026-07-29

## Sequence

1. [x] Add a positive-only saturating accrual operation to the economy ledger while retaining strict
   `Ledger.Apply` behavior.
2. [x] Move production accrual to the ledger-owned saturation path and remove production-side
   hardcap headroom arithmetic.
3. [x] Add the demonstrated R1 regression, deterministic 2,000,000-case near-cap coverage, and
   atomicity/strictness regressions.
4. [x] Update canonical economy and production documentation.
5. [x] Run focused tests, the complete verification gate, archive the implemented RFC and planning
   record, and commit each reviewable stage locally.

## Acceptance gates

- The demonstrated cap/balance/rate vector commits the exact cap with a re-applicable receipt
  delta, and a following purchase succeeds.
- At least 2,000,000 deterministic cases across exponent boundaries complete without rounding
  rejection; every receipt delta reproduces its authoritative `after` value.
- Already-capped accrual succeeds with no change while production time advances.
- Ordinary hardcap overflow remains `ErrAboveHardcap`; negative accrual and invalid starting
  state fail atomically.
- Multi-resource accrual independently saturates and advances resources in deterministic order.
- `docs/economy-kernel.md` and `docs/production-engine.md` state the strict-versus-saturating
  boundary.
- `gofmt`, focused Go tests, and `make verify` pass.
