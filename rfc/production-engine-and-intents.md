# RFC: Production Engine & Intent API

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-07-28
- **Design refs:** `design/02 §2` (production stack, cost curves), `design/02 §10` (the daily clock), `design/06 §idle-math` (closed-form, lazy, server-authoritative), `design/00` law 7 (offline default)
- **Research:** `design/research/tech-stack.md §1` (swarmsim closed forms, intent validation), `design/research/pacing-science.md` (progress checker), `design/research/cookie-clicker.md` (multiplier stack order)
- **Depends on:** RFC-0002 (ledger), Save Layer & Migrations (implemented), Geometric Afford Fast Path (implemented)
- **Split follow-up:** `archive/production-accrual-math.md` (implemented closed-form numeric primitive)
- **Split follow-up:** `archive/generator-production-state.md` (implemented catalog output, owned counts, save cursor)
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

1. A player absent `Δt` accrues exactly the closed-form amount (golden vectors incl. cap and
   credit-banking boundaries). Online and offline evaluation use the same primitive and agree
   when given the same efficiency; the default offline policy supplies `9e-1`, so it intentionally
   yields 90% of the online delta.
2. An unaffordable intent is rejected without mutation; an idempotent replay returns the original receipt.
3. Clock-rollback attempts produce no extra accrual (server-clock property test).
4. Multiplier stack order matches the published documentation (generated from source, per the CI formula-drift gate).
5. `subProgressValue` is monotonic under pure accrual within a stage.
6. 200-bot × 30-virtual-day chaos run: zero NaN/negative/soft-lock, ledger balances (extends the CI Baseline's tiers).

## Open questions

- **Generator persistence and production data:** implemented and archived in
  `archive/generator-production-state.md`.
- **Intent semantics:** `buy_upgrade`, `collect`, and `toggle` have no data model or authoritative
  state transition yet. They cannot be implemented from names alone.
- **Idempotency retention:** storage shape, retention window, and whether failures are replayed are
  unspecified.
- **Compute Credits:** `design/02 §9` requires a bank cap while D4 names only a ratio. Both must be
  concrete balance fields before offline time beyond 24 hours can be persisted.
- **Events:** event payload identity and rollback semantics must be specified before creating the
  deferred append-only table.
- **Progress coordinate:** stage boundaries and behavior at prestige/reset are unspecified.
- **Numeric invariant reporting:** RFC-0001 requires reporting when affordability correction falls
  back or clamps a residual. The numeric helpers are transport-free and currently silent; define
  the actor/audit sink before purchase handlers make those paths live.
- Compute Credit spend UX (needs `design/02 §9` detail — non-blocking, spend path can land as a follow-up).
- Prestige/Exit maths — explicitly a follow-up RFC (needs `design/02 §3`'s cube-root formula plus the run-end sequence from `design/11`).

## Changelog

- 2026-07-28: created (draft).
- 2026-07-28: reviewed by Codex; recorded blocking schemas instead of improvising them and split
  the settled cross-runtime constant-rate accrual primitive into
  `archive/production-accrual-math.md`.
- 2026-07-28: corrected the draft acceptance criterion that had required default 90%-efficient
  offline accrual to equal 100%-efficient online accrual.
- 2026-07-28: split the settled generator output, ownership, and save-cursor contract into
  `archive/generator-production-state.md`; undefined intent and policy mechanics remain gaps here.
- 2026-07-28: generator production state follow-up implemented and archived; remaining open
  questions still block the parent intent engine.
- 2026-07-28: adversarial review routed the currently-latent numeric fallback-reporting contract
  here, where the first authoritative purchase handler can provide its audit sink.
