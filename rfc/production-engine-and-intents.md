# RFC: Production Engine & Intent API

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-07-28
- **Design refs:** `design/02 §2` (production stack, cost curves), `design/02 §10` (the daily clock), `design/06 §idle-math` (closed-form, lazy, server-authoritative), `design/00` law 7 (offline default)
- **Research:** `design/research/tech-stack.md §1` (swarmsim closed forms, intent validation), `design/research/pacing-science.md` (progress checker), `design/research/cookie-clicker.md` (multiplier stack order)
- **Depends on:** RFC-0002 (ledger), Save Layer & Migrations (accepted), Geometric Afford Fast Path (accepted)
- **Planning:** `planning/production-engine-and-intents/` (once implementing)

## Summary

The layer that makes the ledger a *game*: generator ownership, closed-form lazy production evaluated over `Δt` (never per-player server ticks), the intent API (clients send intents, never results), and offline progress. **This RFC also adopts the balance constants currently stranded in `AGENTS.md`** — the last items in the deferred-decisions register with no spec home.

## Motivation

RFC-0002 deliberately excluded "production sources, multiplier stacks, time integration, and offline progress." Nothing can accrue, and no client can buy anything, until this exists. Out of scope: WebSocket transport & fan-out (its own RFC), minigame matches, prestige (needs `design/03 §Exit` maths — a follow-up), the pacing *values* themselves (balance data, harness-gated).

## Specification

### D1 — Closed-form lazy evaluation

- Per player: store `last_evaluated_at`; on any read/intent, integrate production analytically over `Δt` (swarmsim model) at full intermediate precision, commit through the ledger **once** (RFC-0002 K3 quantization). **Never tick players server-side.**
- Threshold-crossing mechanics (caps, unlock predicates) use bucketed evaluation with capped iterations; crossing times are solved, not scanned.
- **Server clock only.** Client-supplied timestamps are never trusted (kills clock-rollback).

### D2 — The production stack (shape is code, values are data)

`rate(resource) = Σ_generators [count × base_rate × Π(multipliers)]` with the multiplier stack in a **documented, fixed order** (the Cookie Clicker lesson: order is observable; publish it). Multiplier sources register into named slots (upgrades, milestones at 25/50/100 owned, faction rules, commons buff, Trust modulation per `02 §7`). All declared in the balance catalog; **no formula strings in data** (RFC-0002 K2 rule extends here).

### D3 — The intent API

- Intents: `{buy_generator, buy_upgrade, collect, toggle}` + an idempotency key. The server validates affordability from **its own** evaluated state, executes through the ledger, and returns the mutation receipt + new canonical snapshot.
- Click/collect batches are rate-clamped (~20–25/s, silent clamp, per `design/06`).
- Invariant checks flag impossible jumps to the audit log (forensics, not auto-bans).
- **The append-only gameplay `events` table lands here** (deferred from the Save Layer RFC): purchases, prestiges, threshold crossings — not clicks.

### D4 — Offline progress (adopting the stranded constants)

- **Offline accrual defaults ON at 90% of online rate, capped at 24 h per absence** — moved here from `AGENTS.md` law 7; `AGENTS.md` now cites this RFC.
- Published in-game (it is already the answer to speedrun "attended time" — `05 §6`).
- Beyond the cap, time banks as **Compute Credits** (`design/02 §9`) at a declared ratio. Ratio is balance data.
- Offline evaluation is the same closed form as D1 — there is no separate offline code path to drift.

### D5 — Progress coordinate

Ship `subProgressValue(state) → 0..1` per stage (the AD progress-checker pattern) as part of this engine — the harness's y-axis and telemetry's core dimension. Stage definitions are balance data.

## Deviations from design

- `AGENTS.md` law 7's constants move into spec + balance data (the register's last stranded item). No other deviation.

## Acceptance criteria

1. A player absent `Δt` accrues exactly the closed-form amount (golden vectors incl. cap and credit-banking boundaries); online/offline paths produce identical results for identical `Δt`.
2. An unaffordable intent is rejected without mutation; an idempotent replay returns the original receipt.
3. Clock-rollback attempts produce no extra accrual (server-clock property test).
4. Multiplier stack order matches the published documentation (generated from source, per the CI formula-drift gate).
5. `subProgressValue` is monotonic under pure accrual within a stage.
6. 200-bot × 30-virtual-day chaos run: zero NaN/negative/soft-lock, ledger balances (extends the CI Baseline's tiers).

## Open questions

- Compute Credit spend UX (needs `design/02 §9` detail — non-blocking, spend path can land as a follow-up).
- Prestige/Exit maths — explicitly a follow-up RFC (needs `design/02 §3`'s cube-root formula plus the run-end sequence from `design/11`).

## Changelog

- 2026-07-28: created (draft).
