# Agent Onboarding — Cloud Clicker

You are working on **Cloud Clicker**: a free, browser-based MMO idle game — *"speedrun any% destroy humanity"* — climbing from a 1995 garage to a world-consuming AI megacorp. Satirical, Cookie-Clicker-lineage, with an MMO layer and a designed ending.

## Current state

**Design phase is complete; implementation has not started.** The repo contains:

- `design/00-vision.md` … `design/10-playstyles.md` — eleven design docs. These are the spec.
- `design/research/` — nine research reports the design cites. Reference material; don't re-litigate settled decisions from them.
- No engine code yet. Phase 0 (see `design/07-roadmap.md`) is next: Go server skeleton, numeric core, Svelte client shell, balance harness.

## Reading order (minimum to be productive)

1. `design/00-vision.md` — pitch, pillars, anti-goals (5 min; read fully).
2. `design/07-roadmap.md` — build order; find your task's phase.
3. `design/06-tech.md` — the stack and architecture decisions. **Binding.**
4. Whichever design doc covers your task's system (`02` economy, `03` minigames, `09` events, etc.).
5. Skim `design/08-satire-flavor.md` §1 (voice rules) before writing ANY player-facing text.

## Non-negotiable design laws

These are settled. Do not "improve" them without explicit sign-off from Marco:

1. **No real money, ever, in any direction.** No IAP, no ads, no telemetry beyond gameplay. Free-ness is the satire's foundation.
2. **Server-authoritative.** Clients send intents, never results. Production math is closed-form and lazy (`06-tech.md §idle-math`) — never a per-player server tick loop.
3. **Big numbers:** `break_eternity.js` (client) + the hand-written Go `Decimal` (server). Wire format is **strings**. Any change to either implementation must keep the shared golden-vector tests green in BOTH test suites.
4. **Balance data is declarative** (JSON/YAML data files, hot-reloadable), never constants in code.
5. **Hardcaps, never softcaps.** Every cap is a visible number.
6. **Every multiplayer feature has an AI/bot fallback**; bots never cheat (same server validation and hidden-info boundaries as humans).
7. **Offline progress is default** (90% rate, 24 h cap baseline) — never a purchased privilege.
8. **Save schema versioned from commit #1** with a migration chain; refuse to persist NaN saves.
9. **Published formulas** for all community/contribution math (the Helldivers transparency rule).
10. **Parodied dark patterns always pull the curtain** (tooltip states what it is). The pet layer is sincere, never a joke (`design/04-pets.md §tone`).

## Stack (decided — see `design/06-tech.md` for rationale)

- **Server:** Go, single binary. `coder/websocket` + embedded `centrifugal/centrifuge`; `chi` for REST. Goroutine-per-player actors; one World goroutine; Match goroutine per minigame.
- **Client:** Svelte 5 (runes) SPA; game state in a plain TS object outside the framework; fixed-timestep 20 Hz sim loop; DOM-first (canvas only for particles/minigame boards). Astro shell for site pages.
- **DB:** Postgres 16 (`saves` jsonb + versions; thin append-only `events`). No Redis until needed.
- **Deploy:** docker-compose (game + postgres + caddy).
- Rejected (don't reintroduce): Nakama at the start, Colyseus, SQLite for saves, math/big.Float, kubernetes.

## Working conventions

- **Language/tooling:** Go code passes `gofmt` + `go vet`; TS is strict-mode; tests accompany every non-trivial change. The golden-vector suite and (once it exists) the balance-harness pacing targets are acceptance gates.
- **Small, reviewable changes.** One system per PR/commit. Reference the design doc section your change implements in the commit message (e.g. `economy: implement generator cost curve (design/02 §2.1)`).
- **Player-facing text** follows the flavor bible voice rules; any real-world statistic must come from the research files, and anything on a research file's "verify before shipping" list must be flagged, not shipped as fact.
- **Don't invent new systems.** If the design doc doesn't cover something you need, leave a `DESIGN-GAP:` comment and surface it in your report rather than improvising a mechanic.
- **Naming in code is mechanical, not flavored** (`generator`, `pressure_meter`, `contribution`) — flavor lives in data files/localization keys, so the satire can be retuned without refactors.
- **Balance numbers in `design/02` are starting values**, expected to be retuned via data files — implement the formula shapes exactly, treat the constants as config.

## Suggested first tasks (Phase 0, in dependency order)

1. Repo scaffolding: Go module + Svelte/Vite workspace + docker-compose + Makefile.
2. **The numeric core**: Go `Decimal` (mantissa/exponent, normalize/Add/Mul/Pow/Log10/Cmp/String + geometric-series bulk-buy & max-affordable) + the golden-vector generator script (Node, break_eternity) + both test suites consuming the same JSON vectors.
3. Save layer: Postgres schema, versioned save struct, migration chain, NaN guard.
4. Production engine: data-file loader (generators/upgrades), closed-form evaluation over Δt, intent API (`buy`, `click-batch`) with validation.
5. Client shell: sim loop, server sync/reconcile, one playable tab (Tier 0).
6. Balance harness: headless strategy runner asserting the pacing targets in `design/02 §10`.

Each of these is independently reviewable and ordered so nothing blocks.
