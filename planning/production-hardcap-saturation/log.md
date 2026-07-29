# Production Hardcap Saturation — Running Log

Append-only implementation record. A fresh agent should read this file with `plan.md` and the RFC.

## 2026-07-29 — Implementation opened

- Re-read `AGENTS.md`, RFC-0000, the accepted RFC, its design references, and the binding technical
  architecture.
- Confirmed the demonstrated fault still exists in the implementation shape: production computes
  quantized hardcap headroom independently, then ordinary `Ledger.Apply` recomputes the final
  balance through a different rounded addition.
- Opened the implementation with the RFC's prescribed ownership boundary: a dedicated
  positive-accrual ledger operation will own saturation; ordinary transactions remain strict.
- Corrected the RFC index from stale `draft` to `implementing` as planning began.

## 2026-07-29 — Ledger ownership implemented

- Added `Ledger.ApplyAccrual`, sharing validation and atomic aggregation with ordinary `Apply` but
  rejecting negative entries and saturating positive production at declared hardcaps.
- The ledger now selects saturated receipt deltas in the specified nominal, −1, +1, −2, +2 ulp
  order and refuses the complete transaction if no canonical delta reproduces the committed cap.
- Removed production's independent hardcap-headroom calculation; `Evaluate` now passes raw accrued
  deltas to the positive-accrual boundary.
- A focused test run exposed why the new receipt rule must stay scoped to accrual: a strict
  purchase can cancel most of a larger balance, producing an `after` with more significant
  resolution than any 12-digit serialization of the aggregate purchase delta. Retaining ordinary
  `Apply` receipt behavior preserves D3; positive accrual has no such cancellation and enforces the
  new re-applicability invariant.
- Existing economy and production suites pass after the ownership change.
