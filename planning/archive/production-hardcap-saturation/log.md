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

## 2026-07-29 — Acceptance regressions green

- Added the exact R1 cap/balance/rate regression. It commits `9.87256122677e8`, emits the
  specified re-applicable `9.81579561655e8` delta, and a following generator purchase applies.
- Added a deterministic 2,000,000-case near-cap gate across zero, lower-magnitude, same-exponent,
  power-boundary, and extreme-exponent inputs. All cases remain within bounds or saturate exactly;
  every emitted accrual delta reproduces its authoritative `after`. Runtime is about 3.5 seconds.
- The exact minimum exponent is tested from zero because a difference below that representable
  floor cannot be a canonical Decimal state value; crossing accrual still saturates at the cap.
- Added already-capped cursor advancement, strict overflow preservation, negative/pre-existing
  invalid-state atomic rejection, and independently saturating multi-resource receipt coverage.
- Full economy and production package tests pass.

## 2026-07-29 — Completed and archived

- Updated `docs/economy-kernel.md` and `docs/production-engine.md` with the distinct strict
  transaction and positive saturating-accrual contracts.
- Full `make verify` passed: Go vet/tests, formula drift, strict TypeScript, 6,354 Node tests,
  schema validation, and 19,062 Chromium/Firefox/WebKit tests.
- All six RFC acceptance criteria are satisfied. Rotated the implemented RFC and planning record
  into their archives. No push performed.

## 2026-07-29 (claude — per-change diff review of b838143..f4113fa, instituted after Marco's challenge)

Full diff read, not spot-checked. **Verdict: approved.** The fix is architecturally better than the
finding's minimal suggestion — saturation moved into the ledger (`ApplyAccrual`, positive-only,
accrual-mode-gated) so there is exactly one rounding authority; strict `Apply` still rejects
overflow, preserving RFC-0002's never-silently-clamp for purchases/conversions. `reproducibleDelta`
verified: ±2-ulp probe via mantissa offset at the nominal's exponent, negatives rejected under
`nonNegative`, every candidate re-verified through the exact quantized-add path, mantissa carry
handled by `New`'s normalization. The exact R1 reproducer is the regression fixture.

**One finding filed (latent, not from this diff but enforced twice by it): there is no
cap-lowering migration policy.** The new starting-balance pre-checks (below-minimum /
above-hardcap → `ErrInvalidTransaction`) are correct belt-and-braces — but if a future balance
change *lowers* a hardcap below existing balances, restore (save D5: one invalid balance rejects
the whole save) bricks every affected save at load, and this check would brick any that slipped
through. Epochs make cap-lowering plausible. **Needed before the first cap-lowering balance
change ships: a documented clamp-on-migration rule (balances above a newly-lowered cap clamp to
it at restore, evented as `compensation`) — belongs to the Leaderboards/Epochs RFC's balance-
change section or a save follow-up.** Until then, lowering a cap is forbidden-by-convention,
which is exactly the kind of unenforced rule this project keeps learning not to trust.

## 2026-07-29 (claude — batch review COMPLETED: all four commits + both pending RFC texts)

The earlier entry reviewed only `ledger.go`. The rest, now read in full:

**`engine.go` (b838143) — approved.** Pure deletion of the buggy clamp; production now sends raw
deltas and `ApplyAccrual` is the single authority. No residual headroom math in the engine.

**Tests (924443d) — approved, strong.** Seven targeted tests. The 2M-case sweep is genuinely
deterministic (fixed seed `0x6a09e667f3bcc909`, its own PRNG — no stdlib global, consistent with
the harness rule), sweeps exponents including ±8_999_999_999_999_999, and structures `before`
into the three hard classes (zero / 9-at-exponent−1 carry edge / random-in-range). The brick's
*second half* is covered end-to-end: `TestEvaluateAtHardcapAdvancesCursorWithoutLedgerChange`
proves a capped stream still advances time. Atomicity negative-control present.

**Docs/archive (f4113fa) — approved.** `docs/economy-kernel.md`'s ApplyAccrual section matches
shipped semantics exactly (accrual-only saturation; strict mode preserved for purchases/grants/
conversions; independent saturation within one atomic transaction).

**RFC texts (fc86a40) — both pending remediation RFCs read in full:**
- `millisecond-cursor-canonicalization` — **approved.** D3 fixes the root, not the symptom: both
  cursors advance to the same `CanonicalServerTime(now)` instead of adding floored durations to
  phase-bearing baselines — which also *self-repairs* existing mismatched saves at first advance.
  D2's migration is safe by monotonicity (R ≤ E ⇒ ⌊R⌋ ≤ ⌊E⌋ — flooring both down cannot invert an
  ordering that held). Lying-v4 saves rejected; corpus gains four fixture classes.
- `resource-log-domain-parity` — **approved.** The 5e-15 bound is *derived* (half-unit of the
  14-digit aligned add), enforced in both loaders **and** the schema command with an honest note
  that JSON Schema cannot compare scientific-notation strings (explicit semantic check, no fake
  regex); defensive post-parse `log10(1+target) > 0` so a bad catalog can never become a runtime
  progress value; D3 bans native `/` from progress division. D1 correctly freezes the primitive.

No findings beyond the cap-lowering policy gap already filed. Batch closed.
