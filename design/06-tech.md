# Tech Stack & Architecture

> Decisions from `research/tech-stack.md`, restated as commitments with rationale. Solo-dev-shaped: one binary, one database, one compose file.

## The stack

| Concern | Choice | Rejected & why |
|---|---|---|
| Backend | **Go, single binary**: `coder/websocket` + **`centrifugal/centrifuge` embedded as a library** (channels, presence, history/recovery on reconnect, JWT); `chi` for REST | Phoenix (learning cost; BEAM slow at number math); Node/Bun (code-sharing win repaid via golden vectors); standalone Centrifugo (extra hop + deployable) |
| Concurrency model | Goroutine-per-player actor; one `World` goroutine (global counters); one `Match` goroutine per minigame; a `Matchmaker` goroutine | Actor frameworks (protoactor, pitaya) — overkill for one dev |
| Realtime fan-out | **Aggregate-then-broadcast**: coalesced global snapshot at 4–10 Hz; presence = join/leave events + periodic count, roster only on subscribe; per-conn buffered queue dropping stale snapshots | Per-click fan-out (the classic mistake) |
| Frontend | **Svelte 5 (runes) SPA**; game state in a plain TS object outside the framework; fixed-timestep 20 Hz sim loop (delta-time from `performance.now()`, catch-up capped — background tabs throttle timers); `$derived` bound only to the visible tab; number formatting throttled ~10 Hz; DOM for all panels; small canvas/Pixi layer only for particles + minigame boards | React (fighting reconciliation); Vue viable (Profectus lineage) but Svelte 5 chosen for fine-grained updates |
| Shell | **Astro** for the site around the game (landing, changelog, guide, leaderboards) with the game as one `client:only` island | SvelteKit-for-everything acceptable fallback if content pages shrink |
| Proxy | **Caddy** (auto-TLS; COOP/COEP headers if/when Stockfish NNUE) | nginx (fine, more config) |
| Database | **Postgres 16**: `saves(player_id, version, state jsonb, updated_at)` upsert keeping last ~5 versions; thin append-only `events` table (purchases, Exits, match results — **not clicks**) for forensics/rollback/analytics | SQLite+Litestream (single-writer lock vs concurrent autosaves) |
| Cache/boards | In-process first; **Redis later** for ZSET leaderboards, presence TTLs, rate buckets, pub/sub broker if multi-node | Starting with Redis (premature) |
| Big numbers | **`break_eternity.js`** client + **hand-written Go `Decimal`** (~400 lines: float64 mantissa + int64 exponent [+ layer when needed]; port from break_infinity.js / BreakInfinity.cs, replicating exact normalization order). **Golden-vector JSON file asserted by both Go and TS test suites**; one-time fuzz vs goja running the real JS. Wire format: **strings, never JSON numbers.** Display: `@antimatter-dimensions/notations` | math/big.Float (slow, base-2, won't agree bit-for-bit); BigInt (no fractional rates) |
| Balance & event data | Declarative, hot-reloadable files (JSON/YAML): generators, upgrades, achievements, events (`09-events.md` schema). `go:embed` in prod, disk-watch in dev | Constants in code (retuning = recompiling = misery) |
| Deploy | docker-compose: `game` (Go) + `postgres` + `caddy` (+ later `redis`, `stockfish`/`maia` sidecars). Single node to ~10k CCU | Kubernetes (the YAML joke stays in the game) |

## Idle math: closed-form, lazy, server-authoritative

- **Never tick players server-side.** Store `last_evaluated_at`; on any read/action, integrate production analytically over Δt (constant: `rate×Δt`; compounding: `a₀(e^{kΔt}−1)/k` — the swarmsim model). Bucketed simulation (1-h steps, capped iterations) only for threshold-crossing mechanics that resist closed form.
- **Clients send intents** (`{buy, generator, count}`), never results; the server validates affordability from its own state. Client-side prediction runs identical formulas for smooth counters and hard-reconciles to server snapshots.
- **Server clock only** (kills the clock-rollback exploit). Offline gains computed server-side on reconnect; the client renders the welcome-back modal from the response.
- **Sim parity:** the production formulas exist twice (Go + TS) — specified in one document, tested against shared golden vectors, like the Decimal core.

## Anti-cheat

1. Rate validation: click batches `{count, window_ms}` clamped at ~20–25/s, silently.
2. Idempotency keys on all mutating actions.
3. Invariant checks on save write: resource jumps beyond `max_theoretical_rate × Δt × slack` → flag to the events table; review, don't auto-ban.
4. Minigames: server owns board/deck/RNG seed; hidden information never sent to clients (the #1 hobby-multiplayer bug); bots subject to identical validation.
5. Leaderboards derive from server-observed events only; "verified" flag.
6. Save schema versioned from commit #1 (`{"v": N}` + migration chain). **NaN detection: refuse to persist a poisoned save, restore last good** (Profectus rule — NaN is the #1 idle-game bug).
7. Ship-It Spellbook RNG is server-seeded per cast, not predictable-deterministic (the FtHoF external-predictor lesson); the honest forecast panel is in-game instead.

## Minigame AI (the vs-AI fallback engine)

- **Chess:** Stockfish binary in the container driven via `notnil/chess`'s UCI subpackage (rules validated server-side for every move, human or bot). Difficulty = Skill Level + shallow `go depth` + **top-k softmax sampling** (temperature = the dial) + fake think time. **Maia** (lc0 sidecar) later if chess becomes headline — it makes *human* mistakes.
- **Connect-4 / Othello / Gomoku / checkers:** one generic ~150-line alpha-beta with iterative deepening + per-game eval functions (Pascal Pons' Connect-4 solver as the reference implementation). Difficulty = depth + temperature.
- **Blackjack:** basic-strategy table. **Poker:** rule-based hand-strength buckets + pot odds + randomized bluff frequency (not CFR).
- **Pet battles / raids:** small-state minimax + personality-flavored policies.
- **Human-feel layer (shared):** plausible-blunder filter (prefer attractive-looking mistakes), jittered think time, named bots with drifting fake ratings, resign behavior, rubber-band toward ~55–60% player win rate. **Bots never cheat.**
- **Matchmaker:** one goroutine; expanding Elo bands (±50, widen per 5 s); bot backfill after 10–30 s (disclosed, reduced rewards, no mid-match hot-swap); queue-depth shown.

## What we take from cattery, and what we don't

**Take:** elapsed-time offline simulation pattern · stat decay + care actions with diminishing returns · two-tier trust/mood · behavior state machines (activity → behavior chains) · CSS-only recolorable sprite system · ambient events · sound event bus · hint/overlay tutorial system · webhook deploy scripts · the `ensureColumn` no-dependency migration idea (translated to real migrations).
**Don't:** Directus/CMS (no CMS needed — content is data files in git) · the unsharded broadcast set (replaced by centrifuge channels) · the N+1 tick loop (replaced by closed-form laziness) · no-auth identity (JWT accounts; anonymous play supported with local-only saves and an upgrade path).

## Nakama decision

Not at the start. Reconsider **only** as a sidecar (matchmaking/leaderboards/social sharing Postgres + JWT) if social features outgrow the hand-rolled versions. Reasons against now: Go-plugin version fragility, opinionated platform shape, poor fit for always-on idle connections, and our matchmaker is ~200 lines.

## Read-before-coding list

[Profectus](https://github.com/profectus-engine/Profectus) (engine architecture, NaN handling, saves manager) · [Antimatter Dimensions source](https://github.com/IvarK/AntimatterDimensionsSourceCode) (GameDatabase data-driven balance; years of save migrations) · [swarmsim](https://github.com/erosson/swarmsim) (closed-form production) · [The Modding Tree](https://github.com/thepaperpilot/The-Modding-Tree) (the layer-definition DSL shape — copy the shape for our tier/generator files).
