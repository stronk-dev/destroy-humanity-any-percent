# RFC: Client Shell & Sim Loop

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Design refs:** `design/06 §frontend` (Svelte 5, 20 Hz, DOM-first), `design/00` law 5 (visible caps), `design/11 §1–2` (contract screen, FTUE surfaces), `design/13 §2` (offline-return fast-forward)
- **Research:** `design/research/browser-rendering.md` (workers, wall-clock deltas, tab throttling), `design/research/tech-stack.md §2`, `design/research/adaptive-balancing.md` (the amplitude-lock boundary, client side)
- **Depends on:** Production Engine (implemented — this consumes its intents, receipts, snapshots, and progress coordinates)
- **Planning:** `planning/client-shell-and-sim-loop/` (once implementing)

## Summary

The Svelte 5 shell, the client-side prediction loop, and the display contract. This RFC owns the **two orphaned decisions** from the deferred-decisions register: **client reconciliation policy** (previously asserted as "hard-reconciles" in a `design/06` table cell, never decided) and **minimum visible increment** (RFC-0002 draft D6, deferred at re-scope). Out of scope: the Pixi spectacle layer and era design tokens (own RFCs), minigame boards, websocket transport specifics (transport RFC — this consumes an abstract snapshot/receipt stream).

## Specification

### D1 — The sim loop

- Game state is a plain TS object outside the framework; the sim is a **fixed-timestep 20 Hz loop driven by wall-clock deltas** (`performance.now()`), catch-up capped at 5 s of simulated time per frame burst — beyond that, the offline path takes over (it is the same closed form; there is no third code path).
- **The sim runs in a Worker** (browser-rendering research: rAF stops in hidden tabs, Workers don't); the main thread renders from a double-buffered snapshot. `$state`/`$derived` bind only to the visible tab's panels; number formatting throttles to ~10 Hz.
- The client predicts with **the same closed forms and the same catalog** the server evaluates (shared TS kernel, already shipped). Prediction is presentation: **no predicted value ever gates a client action** — affordability greying may predict; the buy intent is always sent and the server always decides.

### D2 — Reconciliation policy (owned decision, resolved)

On every authoritative snapshot/receipt, per resource:

| Divergence | Behavior |
|---|---|
| Exact match (canonical string equal) | nothing |
| Continuous drift `< ε_lerp` (default `1%` of displayed value, balance data) | **converge by interpolation over ≤ 400 ms** — the counter bends, never jumps |
| Drift `≥ ε_lerp`, or any **discrete** state difference (counts, unlocks, tier) | **rebase immediately** — snap state, then one subtle pulse animation on affected surfaces; if the cause is a rejected intent, the receipt's typed rejection renders as the explanation |
| Reconnect after gap > 30 s | the **offline-return path** (`design/13 §2` fast-forward), not a silent snap — a gap is a story, not an error |

Rationale: pure snap (the old table cell) makes every network hiccup a visible teleport on the most-watched numbers in the game; pure lerp lies about discrete facts. Continuous values bend, discrete facts snap with receipts. **Prediction windows are short by construction** (intents round-trip in one snapshot period), so `ε_lerp` breaches should be telemetry-rare; a counter of them is part of D4.

### D3 — Display contract (owned decision, resolved — RFC-0002 draft D6 lands here)

- Displayed counters **interpolate between committed states at render rate, at full client precision — never re-quantized for rendering.**
- **A counter must never appear frozen while its production rate > 0.** If the per-frame delta rounds below one displayed unit, the display accumulates sub-unit progress internally (the notation layer shows the next digit moving); the harness asserts this at pathological magnitude gaps.
- **A counter frozen at a declared hardcap is deliberately static and must show the cap and its `reason_key`** (catalog field, RFC-0002 K1) — a frozen number with no explanation is indistinguishable from a bug, and `design/00` law 5 forbids unexplained caps.
- Progress bars render the Production Engine's typed tier coordinates verbatim; the client never invents a progress formula.

### D4 — Shell responsibilities

- Route surface: contract screen → main shell (tabs: company, world, pet, minigame host) → run-end sequence. All DOM; panels lazy-mounted.
- **The reduced-motion token is honored end-to-end** (`prefers-reduced-motion` collapses D2's pulses and D3's interpolation to discrete steps at ≤ 2 Hz — still never frozen-while-producing).
- Telemetry counters (client-side, aggregate-only per compliance): `ε_lerp` breaches, rebase causes by rejection category, worker-tick overruns.
- **Amplitude lock, client side:** the shell package may not import balance-mutation paths; display and prediction read catalog + snapshots only (mirror of the server's compile-time boundary).

## Acceptance criteria

1. A hidden tab for 10 minutes shows, on return, state within one snapshot period of the server — via the catch-up path, without a visible teleport (recorded UI test).
2. Reconciliation table: fixtures for each row, including a rejected `buy_generator` rendering its typed rejection, and a discrete unlock snapping while a continuous counter bends.
3. Frozen-counter property: at a 10^13 magnitude gap between rate and bank, the display still visibly advances (D3 sub-unit accumulation), and at a declared cap it shows the `reason_key`.
4. No client code path gates an action on a predicted value (lint rule + review).
5. Reduced-motion mode passes the same fixtures with stepped rendering.
6. The shell builds and runs against the shipped Phase-0 catalog with zero catalog-specific code.

## Open questions

- `ε_lerp` default and pulse styling: balance/design data, harness- and taste-gated respectively.
- The offline-return modal composition (walkthrough hole #10 — splits vs fast-forward vs ripe-Quarter ordering) **lands in this RFC's implementation but is designed in `design/11`**; blocked on that design note, not on engineering.

## Changelog

- 2026-07-28: created (draft). Owns and resolves the two orphaned decisions from the deferred-decisions register.
