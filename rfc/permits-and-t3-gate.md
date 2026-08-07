# RFC: Permits & the T3→T4 Gate (pre-mint)

- **Status:** draft — the narrow pre-mint contract commissioned by FCE-C1 (owner ruling
  2026-08-07: introduce `company.permits` now rather than interpolate a cash-only gate).
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-07
- **Design refs:** `design/02 §1` (Permits = a Company-run constraint resource alongside
  Energy/Water), `design/02 §role-law` (capacity role: "Energy/Water/Permits caps — the Kittens
  model"), `design/01 §T4` (permits as data-center-era constraint), `design/02 §Externality`
  (constraint-dodging — "unpermitted turbines" — feeds Externality; "Legal% = every permit first").
- **Depends on:** Economy Kernel + Production Engine (archived — this is two rows in the existing
  grammar), Route Registry (archived — one gate row), First Content Epoch (draft — consumes these
  bytes; FCE-C1).
- **Planning:** `planning/permits-and-t3-gate/` (once implementing)

## Summary

The epoch-6 bundle cannot load without `gate.t3_to_t4` (FCE-B1), and the owner ruled the gate is
permits-backed, not cash-only. This RFC introduces the game's SECOND economy resource —
`company.permits` — using ONLY existing grammar shapes (one resource row, one generator class, one
gate row), and defines the T3→T4 gate over it. No new engine mechanics, no new schema fields, no
save-schema bump expected.

## Specification

All amounts are PROVISIONAL BYTES (constants are config; the composed harness and epoch-7 retunes
own tuning). Shapes are exact.

### P1 — The resource row (economy artifact, additive)

```json
{"id": "company.permits", "scope": "company", "numeric_kind": "decimal",
 "initial": "0", "minimum": "0",
 "hardcap": {"amount": "24", "reason_key": "resource.company_permits.cap.phase0"}}
```

A small, VISIBLE hardcap (the hardcaps-never-softcaps law; the Kittens constraint model). Raising
the cap via capacity-role generators is the design's stated destination but is NOT in this
contract — the capacity role is not yet loader-implemented; a later epoch owns it.

### P2 — The faucet: one T3-era generator class (economy artifact, additive)

```json
{"id": "generator.legal_dept",
 "price": {"resource_id": "company.cash", "base": "1e8",
           "curve": {"kind": "geometric", "ratio": "1.15e0"}},
 "production": {"resource_id": "company.permits", "base_rate": "1e-3"}}
```

The idle-honest faucet: permits accrue lazily from a purchased legal department (closed-form
compatible; offline-progress law applies unchanged). `base_rate 1e-3`/s ≈ 1 permit per ~17 min per
department, saturating at the cap. **DESIGN-GAP (flagged, owner-confirmed direction):** design
names permits as a constraint resource and the Kittens cap model but never names the faucet
mechanism; the generator faucet is the chosen minimal shape (over a manual filing action, which
would violate the idle-default law as sole faucet). Confirmed by the FCE-C1 owner ruling round.

### P3 — The gate row (routes artifact, additive — closes FCE-B1)

```json
{"gate_id": "gate.t3_to_t4",
 "requirement": [{"resource_id": "company.cash",    "amount": "1e12"},
                 {"resource_id": "company.permits", "amount": "12"}],
 "routes": []}
```

Cash keeps the shipped ladder (1e9 → **1e12** → 1e15, and the Tier-3 1e12 pacing coordinate);
permits are the constraint the era is ABOUT. Both debit at crossing (the established gate
semantics). `routes: []` in v1 — permit-dodging skips (the Externality-priced
`Non-Road Engine Clip` class) are later route content on this gate, exactly as design/02
§Externality frames them.

### P4 — Copy & integration

- Copy rows for the cap reason key and the generator's player-facing family, through the copy
  pipeline in the same change (names per the existing resource/generator copy conventions).
- The doctrine fixture's `gate.t3_to_t4` reference becomes loadable UNCHANGED (FCE-B1's goal).
- No save-schema bump expected: resource balances are id-keyed. If any code path hardcodes
  `company.cash` as "the" resource, that is an implementation finding to surface, not a contract
  change here.

## Acceptance criteria

1. Both loaders (Go + TS) accept the extended economy and routes artifacts; doctrine
   `ValidateRoutes` passes against the extended routes bytes; composed bundle parity fixture green.
2. Production accrual: a company with N legal departments accrues permits at the closed-form lazy
   rate, saturating at the visible cap with the reason key; offline accrual honors the standard
   90%/24h policy.
3. Gate crossing debits BOTH requirements exactly; insufficient permits rejects with the standard
   typed rejection; the crossing replays byte-identically.
4. Chronology/depletion and route-registry gates green; the change lands as `BALANCE-CHANGE:`
   with its own designated review BEFORE the First Content Epoch consumes the bytes (FCE-C1).

## Changelog

- 2026-08-07: created (draft) — commissioned by the FCE-C1 owner ruling (permits now); minimal
  shapes over existing grammar; faucet DESIGN-GAP flagged and direction owner-confirmed.
