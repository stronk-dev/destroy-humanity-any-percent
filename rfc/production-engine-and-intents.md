# RFC: Production Engine & Intent API

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-07-28
- **Design refs:** `design/02 §2` (production stack, cost curves), `design/02 §10` (the daily clock), `design/06 §idle-math` (closed-form, lazy, server-authoritative), `AGENTS.md` law 7 (offline default)
- **Research:** `design/research/tech-stack.md §1` (swarmsim closed forms, intent validation), `design/research/pacing-science.md` (progress checker), `design/research/cookie-clicker.md` (multiplier stack order)
- **Depends on:** RFC-0002 (ledger), Save Layer & Migrations (implemented), Geometric Afford Fast Path (implemented)
- **Split follow-up:** `archive/production-accrual-math.md` (implemented closed-form numeric primitive)
- **Split follow-up:** `archive/generator-production-state.md` (implemented catalog output, owned counts, save cursor)
- **Boundary follow-up:** `gate-predicates-and-routes.md` (gate alternatives and routes; production is read-only input)
- **Boundary follow-up:** `commons-compact.md` (commons computes one generic multiplier-slot contribution)
- **Planning:** `planning/production-engine-and-intents/` (once implementing)

## Summary

The layer that makes the ledger a *game*: generator ownership, closed-form lazy production evaluated over `Δt` (never per-player server ticks), the intent API (clients send intents, never results), and offline progress. **This RFC also adopts the balance constants currently stranded in `AGENTS.md`** — the last items in the deferred-decisions register with no spec home.

## Motivation

RFC-0002 deliberately excluded "production sources, multiplier stacks, time integration, and offline progress." Nothing can accrue, and no client can buy anything, until this exists. Out of scope: WebSocket transport & fan-out (its own RFC), minigame matches, prestige (needs `design/02 §3` Exit maths — a follow-up), the pacing *values* themselves (balance data, harness-gated).

## Specification

### D1 — Closed-form lazy evaluation

- Per player: use the implemented save cursor `evaluated_through`; on any read/intent, integrate production analytically over `Δt` (swarmsim model) at full intermediate precision, commit through the ledger **once** (RFC-0002 K3 quantization). **Never tick players server-side.**
- Threshold-crossing mechanics (caps, unlock predicates) use bucketed evaluation with capped iterations; crossing times are solved, not scanned.
- **Server clock only.** Client-supplied timestamps are never trusted (kills clock-rollback).

### D2 — The production stack (shape is code, values are data)

`rate(resource) = Σ_generators [count × base_rate × Π(multipliers)]` with the multiplier stack in a **documented, fixed order** (the Cookie Clicker lesson: order is observable; publish it). Multiplier sources register into named slots (upgrades, milestones at 25/50/100 owned, faction rules, commons buff, Trust modulation per `02 §7`). All declared in the balance catalog; **no formula strings in data** (RFC-0002 K2 rule extends here). The exact closed slot union, its order, and its contribution-combination rule are acceptance-blocking schema work below; this draft does not authorize an implementation to invent them.

**The slot boundary (structural, per Codex's review):** multiplier providers emit mechanical contributions into **fixed named slots**. The commons compact populates its slot through the Commons Compact RFC's computed modifier — production consumes the number and knows nothing else. **Route predicates are structurally prohibited from contributing to any production slot** — the Gate Predicates RFC owns them, and its effects touch gates only. Enforce as a compile-time package boundary (the amplitude-lock pattern from `research/adaptive-balancing.md`), not review discipline.

### D3 — The intent API (contract per Codex's 2026-07-28 review, adopted)

- **Two intents only in this RFC: `buy_generator` and `perform_manual_batch`.** `buy_upgrade`, `toggle`, and feature-specific collection are deferred until their state models exist — an intent without a data model is a name, not a contract.
- The server validates affordability from **its own** evaluated state, executes through the ledger, and returns the mutation receipt + new canonical snapshot.
- **Idempotency is per save stream: `(key, request_hash)`.** Replaying a key returns the original success or the original terminal rejection; **reusing a key with a different request hash is a typed conflict.** Retention: **30 days** (provisional — comfortably beyond any reconnect scenario, bounded for storage; owner may tune).
- `perform_manual_batch` is rate-clamped (target range ~20–25/s, silent clamp, per `design/06`). The exact request grammar, action catalog, and clamp algorithm are acceptance-blocking schema work below.
- Invariant checks flag impossible jumps to the audit log (forensics, not auto-bans). **The numeric fallback-reporting contract (RFC-0001 §7, routed here by the adversarial review) lands in this handler's audit sink.**
- **Events are immutable and atomically tied to the resulting save revision.** Corrections are compensating events on later revisions — **history is never deleted.** Purchases, prestiges, threshold crossings; never clicks.

### D4 — Offline progress (adopting the stranded constants)

- **Offline accrual defaults ON at 90% of online rate, capped at 24 h per absence** — moved here from `AGENTS.md` law 7; `AGENTS.md` now cites this RFC.
- Published in-game (it is already the answer to speedrun "attended time" — `05 §6`).
- Beyond the cap, time banks as **Compute Credits** — **exact integer milliseconds, never `Decimal` currency** (Codex's review; time is a count, per RFC-0001 contract §1). Required balance fields, all four: `bank_ratio` (banked ms per excess offline ms), `bank_cap_ms`, `burst_speed` (rate multiplier while spending), `burst_max_duration_ms`. **Provisional launch values, harness-gated: ratio 0.5, cap 72 h, burst ×2, max burst 4 h per activation.**
- Offline evaluation is the same closed form as D1 — there is no separate offline code path to drift.

### D5 — Progress coordinate

Ship `subProgressValue(state) → 0..1` per stage (the AD progress-checker pattern) — the harness's y-axis and telemetry's core dimension. **Coordinates are typed, tier-local definitions, never arbitrary formulas** (Codex's review): the catalog declares one of a closed kind-union per tier — `resource_log` (log-progress toward a resource threshold), `count_fraction` (owned/required exact counts), `composite` (fixed weighted sum of the former two). **Every tier requires an explicit monotonic coordinate before this section is accepted.** Candidate T0–T3 shapes are T0 `resource_log` on cash toward the first-generator ladder, T1 `composite` counts+cash, and T2/T3 `resource_log` on the tier-gate resource. They remain proposals until concrete catalog entries define their thresholds and weights; T4+ land with their tier content.

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

## DESIGN-GAPs blocking acceptance

These are contracts, not implementation details. Per RFC-0000 and `AGENTS.md`, the RFC remains
draft until each is specified here or deliberately split to a named follow-up with a clean boundary.

- **Intent wire contracts:** define the request and receipt schemas for `buy_generator` and
  `perform_manual_batch`, including generator/action identifiers, count semantics, the authoritative
  evaluation order, and typed rejection categories. The manual action also needs catalog data for
  its output and an exact server-time clamp algorithm; “~20–25/s” is not executable.
- **Production slot contract:** define the closed slot union, canonical order, how multiple
  contributions within one slot combine, the shared package interface, and the balance-catalog
  schema. The current catalog contains generator price/output data but no multiplier-slot objects.
- **Idempotency persistence:** D3 settles the 30-day policy and replay behavior, but not key grammar,
  canonical request hashing, which rejections are terminal, the persistence schema, expiry, or the
  transaction boundary tying idempotency records to ledger/save/event commits.
- **Event envelope:** define event identity, version, kind registry, stream/save revision, intent
  reference, constants hash, timestamp, payload schemas, and atomic persistence. This RFC may emit
  purchase events; future prestige and threshold event kinds must be extensible without pretending
  their absent state models are implemented. Compensation semantics need a typed contract too.
- **Compute Credit state transition:** define the save field and migration, exact integer rounding
  for `bank_ratio = 0.5`, cap behavior, elapsed-time partition at the 24 h boundary, and the eventual
  spend transition. If spending remains a follow-up, name that RFC and keep only accrual/persistence
  in this one.
- **Progress coordinate data:** add concrete T0–T3 catalog entries with stage boundaries,
  thresholds, weights, and reset behavior. D5 itself requires these before acceptance; prose labels
  for proposed shapes do not satisfy that requirement.
- **Numeric invariant reporting:** RFC-0001 requires reporting when affordability correction falls
  back or clamps a residual. Define the actor/audit interface and sink before purchase handlers make
  those paths live.
- **Chaos-harness ownership:** acceptance criterion 6 depends on bot inputs and a harness that do not
  yet exist. Define its fixtures here or move that criterion to the named Balance Harness RFC while
  retaining smaller production-engine safety tests locally.

## Named follow-ups

- Compute Credit spending and UX, if split from the accrual/persistence scope above (needs
  `design/02 §9` detail).
- Prestige/Exit maths (needs `design/02 §3`'s cube-root formula plus the run-end sequence from
  `design/11`).
- `buy_upgrade`, feature collection, and toggles return only with their authoritative state models.

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
- 2026-07-28: Codex acceptance review rejected premature acceptance: reconciled adopted decisions
  with the open-question section, corrected the implemented save-cursor name, and recorded the
  remaining executable contracts as explicit DESIGN-GAPs.
