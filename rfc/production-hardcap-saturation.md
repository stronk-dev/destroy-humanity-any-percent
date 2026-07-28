# RFC: Production Hardcap Saturation

- **Status:** draft
- **Author:** Marco (drafted by Codex from the round-2 review)
- **Created:** 2026-07-28
- **Design refs:** `design/00-vision.md` law 5; `design/02-economy-balancing.md §2.2`
- **Research:** `planning/production-review-round2/log.md` R1
- **Amends:** `archive/0002-economy-constants-and-ceilings.md` and
  `archive/production-engine-and-intents.md`
- **Planning:** `planning/production-hardcap-saturation/` (once implementing)

## Summary

Make positive production accrual saturate atomically at a declared hardcap without weakening the
ledger's reject-on-overflow rule for ordinary transactions. The current production headroom and
ledger addition paths can disagree by one 12-digit ulp, causing every later intent to fail while
accrual tries to cross the cap again.

## Motivation

The round-2 review demonstrated the failure with cap `9.87256122677e8`, balance
`5.6765610215e6`, and rate `1e0/s`. Production computes headroom `9.81579561656e8`; applying that
through the ledger produces `9.87256122678e8`, one ulp above the cap, and the ledger correctly
rejects. Because intent evaluation accrues before every command, the company stream is then
permanently unusable. Both components honor their existing contracts in isolation; the missing
contract is who owns capped accrual rounding.

## Specification

### D1 — The ledger owns capped accrual

Add a distinct ledger operation for **positive accrual transactions**. Production passes its raw,
non-negative accrued deltas to this operation and removes its independent headroom subtraction.
The ledger validates scope, aggregates entries with `SumDeterministic`, and computes the committed
result using the same quantized-add path as ordinary transactions.

For each resource:

1. Reject an invalid/non-state/negative accrual delta or a starting balance already outside its
   declared bounds.
2. Compute `unbounded = Quantize12(before + aggregate_delta)`.
3. If no hardcap exists or `unbounded <= hardcap`, commit `unbounded` normally.
4. If `unbounded > hardcap`, commit the declared hardcap exactly and report a corrected applied
   delta whose quantized addition to `before` reproduces that hardcap.

This is saturation at the authoritative transaction boundary, not overflow followed by repair:
no above-cap value is ever committed or exposed.

### D2 — Reproducible receipt deltas

Receipts report the **actual aggregate delta applied**, not a fresh `after - before` subtraction.
For a saturated entry, the ledger starts from `Quantize12(hardcap - before)` and probes neighboring
12-digit ulps in deterministic order (nominal, −1, +1, −2, +2), discarding negative candidates.
It selects the first candidate for which `Quantize12(before + candidate) == hardcap`. Failure to
find one is an internal invariant error and commits nothing.

For the demonstrated vector, the applied delta is `9.81579561655e8`, which re-applies to the exact
cap. Every emitted change must satisfy both:

```text
Quantize12(before + delta) == after
minimum <= after <= hardcap (when present)
```

The receipt's `before` and `after` remain authoritative snapshots. This contract also makes the
Balance Harness Foundation's ledger-reconciliation invariant executable.

### D3 — Ordinary transactions remain strict

`Ledger.Apply` retains its current semantics: a purchase, grant, migration, or other ordinary
transaction that would exceed a hardcap returns `ErrAboveHardcap` and mutates nothing. Only the
new positive-accrual operation may saturate, and it rejects negative entries. Call sites cannot
opt arbitrary rewards into clamping.

### D4 — Production at and near a cap

- Accrual that reaches or crosses a cap succeeds, leaves the balance exactly at the catalog cap,
  and advances the evaluation cursor normally.
- Further positive accrual at the cap is a successful zero-change evaluation; it does not emit a
  ledger change and does not block the following intent.
- Multiple produced resources saturate independently in the same atomic transaction. An invalid
  resource still aborts the complete transaction.
- Client prediction may temporarily display a value above the cap internally, but reconciliation
  uses the authoritative capped `after`; the Client Shell RFC's visible-cap rule remains binding.

## Deviations from design

None. This makes the existing hardcap law executable while retaining ordinary overflow rejection.

## Acceptance criteria

1. The demonstrated cap/balance/rate vector evaluates successfully to `9.87256122677e8`; its
   receipt delta is re-applicable and a following purchase intent can execute.
2. At least 2,000,000 deterministic near-cap cases across exponent boundaries either remain below
   the cap or saturate exactly; zero evaluations fail from rounding and every receipt re-applies to
   its `after` value.
3. Accrual at an already reached cap succeeds with no balance change and advances time.
4. A normal `Ledger.Apply` overflow still returns `ErrAboveHardcap`; a negative accrual entry and a
   pre-existing above-cap balance are rejected atomically.
5. Multi-resource tests prove one resource can saturate while another advances normally in one
   receipt, with deterministic resource ordering.
6. `docs/economy-kernel.md` and `docs/production-engine.md` describe the distinct strict-transaction
   and saturating-accrual contracts.

## Open questions

None. The saturation boundary and correction order are fully specified above.

## Changelog

- 2026-07-28: drafted from demonstrated round-2 finding R1.
