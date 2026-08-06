# RFC: Client Shell & Sim Loop

- **Status:** implemented
- **Author:** Marco (drafted by Claude)
- **Design refs:** `design/06 §frontend` (Svelte 5, 20 Hz, DOM-first), `design/00` law 5 (visible caps), `design/11 §1–2` (contract screen, FTUE surfaces), `design/13 §2` (offline-return fast-forward)
- **Depends on:** Production Engine (implemented — this consumes its intents, receipts, snapshots, and progress coordinates)
- **Planning:** `planning/archive/client-shell-and-sim-loop/`

## Summary

The Svelte 5 shell, the client-side prediction loop, and the display contract. This RFC owns the **two orphaned decisions** from the deferred-decisions register: **client reconciliation policy** (previously asserted as "hard-reconciles" in a `design/06` table cell, never decided) and **minimum visible increment** (RFC-0002 draft D6, deferred at re-scope). Out of scope: the Pixi spectacle layer and era design tokens (own RFCs), minigame boards, websocket transport specifics (transport RFC — this consumes an abstract snapshot/receipt stream).

## Specification

### D1 — The sim loop

- Game state is a plain TS object outside the framework; the sim is a **fixed-timestep 20 Hz loop driven by wall-clock deltas** (`performance.now()`), catch-up capped at 5 s of simulated time per frame burst — beyond that, the offline path takes over (it is the same closed form; there is no third code path).
- **The sim runs in a Worker** (rAF stops in hidden tabs, Workers don't); the main thread renders from a double-buffered snapshot. `$state`/`$derived` bind only to the visible tab's panels; number formatting throttles to ~10 Hz.
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

- Pulse styling remains presentation data and may change without changing reconciliation semantics.
- The offline-return composition is resolved in `design/11 §1`: ≤5 s skippable diorama
  fast-forward, gains docked in the header, then at most one ripe-Quarter modal.

## Executable contracts

### C1 — Abstract authoritative stream

The shell consumes an injected `SnapshotStream`; it does not own WebSocket framing. A snapshot has
an exact non-negative safe-integer revision and evaluated-through millisecond, a canonical
`constants_hash`, resource records `{amount, rate_per_second, cap?}`, discrete facts, and typed
progress coordinates. A receipt has revision, intent ID, applied/rejected status, and the Production
Engine rejection code. Resource amounts/rates/caps are canonical Decimal strings. Unknown or
non-canonical state fails at the adapter boundary; revisions below the last committed revision are
ignored. The transport RFC maps its wire envelope into this interface.

### C2 — Worker and clock protocol

The dedicated module Worker owns prediction state. Exact message kinds are `initialize`,
`authoritative_snapshot`, `clock_pulse`, and `dispose`; outputs are `predicted_snapshot`,
`offline_required`, and `worker_metric`. Every pulse carries `performance.now()` from the injected
monotonic clock. The worker uses a 50-ms accumulator, executes at most 100 fixed steps per pulse,
and sends a 100-ms snapshot cadence. A gap beyond 5,000 ms emits `offline_required` and performs no
local catch-up. Worker messages contain structured-clone data only; no `SharedArrayBuffer`, WASM,
or client persistence authority is introduced.

### C3 — Reconciliation policy data

Strict client-shell catalog schema v1 owns `tick_ms=50`, `snapshot_ms=100`,
`catchup_ceiling_ms=5000`, `epsilon_lerp_ppm=10000`, `lerp_duration_ms=400`,
`reconnect_story_threshold_ms=30000`, and `reduced_motion_render_ms=500`. Continuous relative
divergence is `abs(predicted-authoritative)/max(abs(authoritative),1)` through Decimal operations.
Below epsilon, a resource converges linearly over the configured duration; at/above epsilon it
rebases. Discrete differences always rebase. Rejected receipts always rebase and expose their typed
code as the explanation. Gaps over 30 s select the design/11 return story.

### C4 — Display state

Display arithmetic never calls canonical quantization. Each counter retains committed Decimal,
predicted Decimal, interpolation endpoints, and an unquantized sub-unit accumulator. The view exposes
the live Decimal plus a monotonic `activity_ppm` indicator, so a producing counter cannot render as
unchanged even when notation digits cannot move. A counter equal to its catalog cap exposes the cap's
`reason_key`; missing reasons reject catalog/snapshot adaptation. Counts/unlocks/tier are discrete and
never interpolated. Reduced motion uses 500-ms discrete presentation samples with identical values and
receipts.

### C5 — Lifecycle and return story

`visibilitychange` requests an authoritative snapshot; `pagehide` requests a flush; `freeze`
disposes worker/channel resources. No `unload` handler exists. A reconnect gap over 30 s yields the
single return sequence from design/11: skippable fast-forward capped at 5 s, gains header, optional
ripe-Quarter modal, everything else badges. The shell exposes this state but does not invent Fiscal
Quarter mechanics.

### C6 — Action and package boundaries

`IntentDispatcher.send(intent)` accepts no predicted state or affordability result and always invokes
the injected authoritative request adapter. A source-boundary test forbids the dispatcher from
importing prediction modules and forbids shell code from importing balance mutation paths. Shell code
may read validated catalogs, snapshots, receipts, and the existing numeric/production kernels only.

## Deviations from design

None. The worker uses structured-clone snapshots initially; a transferable-ArrayBuffer
optimization remains optional because canonical Decimal strings cannot be represented losslessly as
binary floats.

## Changelog

- 2026-07-28: created (draft). Owns and resolves the two orphaned decisions from the deferred-decisions register.
- 2026-07-29: accepted for implementation by owner direction; C1-C6 close the abstract stream,
  worker, catalog, reconciliation, display, lifecycle, and action-boundary contracts. The stale
  offline-return blocker was resolved by the already-adopted `design/11 §1` sequence.
- 2026-07-29: implemented, independently reviewed, documented in `docs/client-shell.md`, and
  archived. The review's sole routed finding was subsequently disproved by Production C1 and the
  live exact-schema parser, both of which require `window_ms`; the correction is preserved in the
  append-only planning log.
- 2026-08-06: non-normative reference cleanup for publication; no spec change.
