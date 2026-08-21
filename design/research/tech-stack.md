# Research: Tech Stack — Browser MMO Idle/Clicker (Go backend, self-hosted)

> **Historical research snapshot — superseded where it conflicts with current authority.** This
> report records the July 2026 option study; it is not an implementation specification. The
> binding stack is [`design/06-tech.md`](../06-tech.md), and the implemented numeric contract is
> summarized in [`numeric-core.md`](numeric-core.md). In particular, current authority selects
> `break_infinity.js` 2.2.0 plus the hand-written Go `Decimal`, does not require bit-for-bit
> cross-runtime arithmetic, and forbids hidden rubber-band bots. Library status, pricing,
> capacity, benchmarks and ecosystem claims below remain dated July 2026 observations.
>
> Web research agent report, July 2026. Context: solo/hobbyist developer, Go-leaning, browser frontend, self-hosted docker-compose behind nginx/caddy. Needs: realtime presence + shared global counters for thousands of CCU, idle/offline simulation, minigames (vs players and vs AI), big-number arithmetic, save-state persistence.

---

## 1. Backend: Go realtime for MMO idle

### 1.1 WebSocket library landscape (2025/2026)

| Library | Status | Notes |
|---|---|---|
| [gorilla/websocket](https://github.com/gorilla/websocket) | ~25k stars, "API is stable", actively maintained again | Archived Dec 2022, **un-archived/re-adopted by new maintainers in 2023**. Most StackOverflow answers, most tutorials. Footgun: concurrent `WriteMessage` from two goroutines panics — you must funnel writes through one goroutine per conn. |
| [coder/websocket](https://github.com/coder/websocket) (was `nhooyr.io/websocket`) | Actively maintained, renamed 2024 | Idiomatic: `context.Context` everywhere, safe concurrent writes internally, `wsjson`/`wspb` helpers, compiles to WASM. **Default recommendation for new code.** [comparison guide](https://websocket.org/guides/languages/go/) |
| [gobwas/ws](https://github.com/gobwas/ws) | Maintained, niche | Zero-copy/zero-alloc, works with epoll event loops (`gobwas/ws-examples` + `mailru/easygo`) so you can hold 100k+ idle conns without 2 goroutines each. Steep API. Only reach for it at goroutine/memory ceilings. |
| [olahol/melody](https://github.com/olahol/melody) | Thin wrapper over gorilla | Sessions, broadcast/broadcastFilter, buffered writes (concurrency-safe), automatic ping/pong. ~200 LoC of value; great for a hobby project that just wants "broadcast to everyone" without writing a hub. |
| [lxzan/gws](https://github.com/lxzan/gws) | Newer, high-perf | Event-driven, low alloc, optional epoll. Worth benchmarking if you outgrow coder/websocket. |

**Practical note for this scale:** "thousands of concurrent players" is *small*. One Go process on a 4-core VPS with coder/websocket handles 10k idle connections comfortably (budget ~0.5–1 MB/conn including send buffers, per [scaling guidance](https://dev.to/young_gao/scaling-websocket-connections-from-single-server-to-distributed-architecture-1men)). You do not need gobwas/ws, and you do not need horizontal scaling on day one.

### 1.2 Ready-made realtime layer: Centrifugo / centrifuge

Two things, same author:

- **[centrifugal/centrifuge](https://github.com/centrifugal/centrifuge)** — a **Go library**. You import it, you get: WebSocket + SSE + HTTP-streaming transports with automatic fallback, channels/pub-sub, **presence + presence stats**, **channel history with recovery on reconnect** (huge — solves "player reconnects, missed the last 30 counter ticks"), JWT auth hooks, pluggable broker (in-memory or Redis). Bidirectional, you keep full control of each connection.
- **[centrifugal/centrifugo](https://github.com/centrifugal/centrifugo)** — the **standalone server** built on that library. Language-agnostic; your Go game logic talks to it over HTTP/GRPC publish API, and it proxies connect/subscribe/publish events back to your backend. Transports: WebSocket, HTTP-streaming, SSE, WebTransport, GRPC. Scaling engines: Redis / Redis Cluster / NATS. Docs: [centrifugal.dev](https://centrifugal.dev/)

**Recommendation shape:** use the **`centrifuge` library embedded in your Go binary**, not the separate Centrifugo server. You get presence + history + recovery + client SDKs (`centrifuge-js`) essentially free, in one docker-compose service, and you can still write raw game-logic handlers. The separate server adds a network hop and a second deployable for benefits (polyglot backends, ops isolation) you don't need as a solo dev.

Alternative if you want *nothing* extra: melody/coder + your own hub. ~300 lines. Fine, but you'll reimplement presence and reconnect-recovery badly.

### 1.3 Fan-out patterns for global counters + presence

The classic mistake is publishing every player's every click to every player. Patterns that work:

1. **Aggregate-then-broadcast (tick coalescing).** Global counters (total server-wide production, players online) live in memory as `atomic.Int64` / a sharded counter. A single ticker goroutine at 2–10 Hz snapshots them and publishes **one** message to the `global` channel. Cost is O(clients), not O(clicks×clients). This is the single most important design decision.
2. **Delta / diff payloads.** Send `{delta: +1234, seq: N}` not full state. Centrifugo has built-in delta compression for channels.
3. **Presence feed as an event stream, not a full roster.** Broadcast join/leave events + a periodic "N online" number. Only send the full roster on subscribe (`presence()` call), never on every change. Centrifuge's presence is Redis-hash-backed with TTL heartbeats.
4. **Sharded chat/regions.** If the "presence feed" is chatty, shard players into rooms of ~200–500 so fan-out per message stays bounded.
5. **Per-connection outbound queue with drop policy.** Buffered channel (say 64 msgs); on overflow, drop *stale* counter updates (they're idempotent snapshots) rather than blocking the broadcaster. Classic Go [channel drop pattern](https://dev.to/b0r/go-channel-patterns-drop-4k19).
6. **Redis pub/sub only when you go multi-node.** Redis pub/sub is fire-and-forget with no replay — fine for counters, bad as your only path for gameplay events. [Ably's writeup](https://ably.com/blog/scaling-pub-sub-with-websockets-and-redis). Start single-node.

### 1.4 Actor-ish patterns in Go

Go's natural idiom: **one goroutine owns one piece of mutable state; everything else sends it messages over a channel.** No library needed for a hobby project.

- `PlayerActor`: goroutine per active player, owns that player's state, inbox channel of commands. At 5k players that's 5k goroutines ≈ 40 MB of stacks — totally fine.
- `RoomActor` / `MatchActor`: goroutine per minigame match with its own tick loop.
- `WorldActor`: single goroutine owning global counters + leaderboard.

Libraries if you want structure: [vladopajic/go-actor](https://github.com/vladopajic/go-actor) (lightweight), [asynkron/protoactor-go](https://github.com/asynkron/protoactor-go) (full Akka-style — heavy for solo), [cherry-game/cherry](https://github.com/cherry-game/cherry) (Go MMO framework, Chinese docs, read for architecture), [topfreegames/pitaya](https://github.com/topfreegames/pitaya) (production Go game server framework).

**Honest take:** for one dev, plain goroutines + channels + a `map[playerID]*Player` guarded by shards beats any actor framework. Use `context.Context` for lifecycle and a supervisor goroutine that restarts crashed player actors from the last snapshot.

### 1.5 Would Elixir/Phoenix or Node/Bun be better?

**Phoenix is genuinely the "right" tool** for the realtime half. [Phoenix.Presence](https://hexdocs.pm/phoenix/Phoenix.Presence.html) is a CRDT-based distributed presence tracker with zero external dependencies. Channels give you topics; LiveView could render the entire idle UI server-side. Process-per-player is the BEAM's native model.

| | Phoenix/Elixir | Go | Node/Bun |
|---|---|---|---|
| Presence/pubsub built-in | **Best in class**, no Redis | centrifuge gets you ~90% there | socket.io/uWebSockets + Redis adapter |
| Per-player process isolation | Native, supervised, cheap | goroutines (cheap, but a panic can kill process without care) | Worse — single-threaded event loop |
| Big-number CPU work (idle math) | BEAM is *slow* at number crunching | **Fastest of the three** | JIT decent; but can reuse break_eternity.js server-side |
| Learning cost for a Go dev | High | Zero | Low |
| Shared code with browser | No | No | **Yes** — same big-number lib and sim code client & server |
| Ops/self-host in compose | Releases fine, clustering fiddly | **Single static binary** | Fine |

**Verdict:** the killer Node/Bun argument is *code sharing* — running the exact same `break_eternity.js` and production formulas on client and server eliminates a whole class of desync bugs. The killer Go argument: you know it, one 15 MB binary, fastest at CPU-bound offline-simulation math. The pragmatic middle path: **Go backend + a tiny, carefully-specified numeric core duplicated in Go and TS with a shared golden-test vector file** (JSON of `[a, b, op, expected]` run by both test suites). ~A day of work, buys Go's ergonomics without desync risk. Phoenix only if you actively want to learn Elixir; not worth it for 5k CCU, which Go handles on a $10 VPS.

### 1.6 Datastore: Postgres vs SQLite+Litestream vs Redis

- **Postgres** — pick this. Idle MMO = many small concurrent writes (save-state upserts, leaderboard, minigame results), plus `JSONB` for save blobs, row-level locking, `SELECT ... FOR UPDATE` for currency transactions. SQLite has a **database-level write lock** (one writer at a time even in WAL mode). Note **Nakama requires Postgres/CockroachDB anyway**. [SQLite vs Postgres 2026](https://goilerplate.com/blog/sqlite-vs-postgres-indie-saas)
- **SQLite + [Litestream](https://litestream.io/)** — beautiful for single-server, read-heavy, <10k DAU. But this workload is write-heavy-ish and bursty (every player autosaving every 30s); the single-writer lock will bite.
- **Redis** — not the source of truth. Use for: (a) ephemeral presence hashes with TTL, (b) global counters (`INCRBY`) with periodic flush to Postgres, (c) leaderboards via `ZSET` (*the* leaderboard primitive), (d) pub/sub when multi-node, (e) rate-limit counters. Optional at first: in-process implementations cover all of it single-node.

**Concrete schema sketch:** `players(id, ...)`, `saves(player_id, version, state jsonb, updated_at)` (upsert, keep last N versions for rollback), `events(id bigserial, player_id, type, payload jsonb, ts)` append-only for audit/anti-cheat, `leaderboard` materialized from Redis ZSET periodically.

### 1.7 Event sourcing vs snapshot saves for idle games

**Recommended hybrid for idle specifically:**
- **Snapshot is the source of truth.** Player state is a versioned struct: resources (big numbers), building counts, upgrade bitsets, prestige layer, `last_tick_at`. Written on a debounce (every 15–60s of activity, on disconnect, on prestige, on expensive purchases).
- **Append a *thin* event log alongside** for: purchases, prestiges, minigame results, currency grants. Not for clicks. Gives you (a) anti-cheat forensics, (b) rollback for a bad balance patch, (c) analytics, (d) recovery.
- **Never replay from t=0.** Idle games have closed-form progression; full event replay is unnecessary and unbounded in cost.
- **Version your save schema from commit #1.** `{"v": 7, ...}` plus a chain of migration functions `migrate_6_to_7(state)`. Every long-lived idle game (Kittens, AD, Evolve) has this and every one wishes they'd started sooner.

### 1.8 Anti-cheat for server-authoritative idle games

The core structural insight: **an idle game's progression is a pure function of (state, elapsed_time). Never let the client tell you how much it produced.**

1. **Closed-form, lazy, server-side production.** Don't run a 20 Hz tick per player server-side. Store `last_evaluated_at` and compute on demand: constant rate `gain = rate × Δt`; compounding `gain = a₀ · (e^{kΔt} − 1)/k` — evaluate the integral, don't loop. Fall back to **bucketed simulation** (e.g. 1-hour steps, capped iterations) only for mechanics that can't be closed-form (discrete unlock thresholds crossed mid-interval).
2. **Client sends *intents*, never results.** `{"action":"buy","building":"farm","count":10}` — server validates affordability from *its* state and applies. The client's number is a prediction for UI smoothness only.
3. **Action rate validation.** Token bucket per player per action type. Clicks: cap at a humanly plausible rate (~20–25/s) and **clamp silently** rather than kicking — autoclickers are usually tolerated in idle games; what you must block is `clicks: 1e9` in one packet. Batch clicks as `{count, window_ms}` and validate.
4. **Server clock only.** Ignore client timestamps entirely. Prevents the #1 idle exploit: system-clock rollback.
5. **Offline progress cap.** 8–24h standard; optionally an efficiency multiplier so offline is never strictly better.
6. **Invariant/sanity checks on save write.** Flag saves where resources jumped more than `max_theoretical_rate × Δt × slack`. Log; don't auto-ban, review.
7. **Idempotency keys on mutating actions** so a retried packet can't double-spend.
8. **Nothing sensitive in the client bundle.**
9. **Minigames:** server owns the game state (board, deck, RNG seed). Never send opponent's hidden info to the client — the single most common bug in hobby multiplayer card games.
10. **Global leaderboards are the cheat magnet.** Derive from *server-observed* events only; keep a "verified" flag.

---

## 2. Frontend for idle games

### 2.1 What real incremental games actually use — mostly plain DOM

| Game | Stack | Source |
|---|---|---|
| **Cookie Clicker** | Vanilla JS + DOM, hand-rolled, single huge `main.js`. Small canvas for the cookie/rain only. Not open-source; ships unminified/readable. | [wiki](https://cookieclicker.fandom.com/wiki/JavaScript_files) |
| **Antimatter Dimensions** | **Vue 2/3**, `break_infinity.js`, MIT | [GitHub](https://github.com/IvarK/AntimatterDimensionsSourceCode) |
| **Kittens Game** | Vanilla JS + DOM + hand-rolled MVC, `setInterval(game.tick, 30)` | [GitHub](https://github.com/nuclear-unicorn/kittensgame) |
| **Evolve** | Vanilla JS modules + Vue for reactivity, LESS | [GitHub](https://github.com/pmotschmann/Evolve) |
| **Synergism** | **TypeScript, no framework** (direct DOM), Capacitor for mobile, MIT | [GitHub](https://github.com/Pseudo-Corp/SynergismOfficial) |
| **Swarm Simulator** | Angular (older); author wrote `swarm-numberformat` | [GitHub](https://github.com/erosson/swarmsim) |
| **Profectus** (engine) | **Vue 3 + TS + JSX**, `break_eternity.js`, NaN detection, saves manager, dynamic layers | [GitHub](https://github.com/profectus-engine/Profectus) |
| **The Modding Tree** (engine) | Vue 2 + plain JS, beginner-oriented | [GitHub](https://github.com/thepaperpilot/The-Modding-Tree) |

Takeaway: **DOM is not the bottleneck for idle games; string formatting and big-number math are.** A panel/button/tab UI with ~500 visible numbers updating at 10–20 Hz is trivial for any modern framework — *if* you don't reformat 5000 offscreen numbers every frame.

### 2.2 When canvas/PixiJS is warranted

Guideline: **beyond a few hundred simultaneously-animating elements, switch to canvas** ([particle benchmark](https://github.com/quidmonkey/particle_test), [js-game-rendering-benchmark](https://github.com/Shirajuki/js-game-rendering-benchmark)).

- **DOM**: all panels, buttons, tabs, resource readouts, shop lists, chat/presence feed. All of it.
- **Canvas/[PixiJS](https://pixijs.com/)**: only for (a) particle/juice effects (floating "+1", rain), (b) map/visualization, (c) minigame boards with animation. Even then, **DOM overlay on top of canvas** for buttons/HUD is standard. Lighter alternatives: [Konva](https://konvajs.org/) (scene graph, good hit-testing for board games) or raw 2D context.

### 2.3 Framework choice for a panels/buttons/tabs game UI

- **Svelte 5 (runes)** — `$state`/`$derived`/`$effect` = signal-style fine-grained reactivity; updates only the specific text nodes bound to changed state, no VDOM shipped. Holds its advantage precisely in the "10,000 rows, updates every 100ms" scenario. ([benchmark](https://www.sitepoint.com/react-19-compiler-vs-svelte-5-virtual-dom-latency-benchmark/))
- **SolidJS** — technically the most efficient reactivity; slight raw-benchmark edge over Svelte 5; smaller ecosystem.
- **React** — works (with `useSyncExternalStore` + Zustand/Jotai + memoization) but you're fighting reconciliation to do what Svelte/Solid do for free.
- **Vue 3** — the *de facto* incremental-game framework by installed base (AD, Evolve, Profectus, TMT). Choosing Vue means you can literally read Profectus's source and lift patterns.

**Architectural rule that matters more than the framework:** keep game state *outside* the framework. A plain TS `GameState` object mutated by a fixed-timestep loop, with the framework subscribing to a small set of **selectors** (only values visible in the open tab). Recompute derived values in a dependency-ordered pass once per tick, not lazily per-getter — idle games have deep multiplier chains where naive lazy memoization thrashes.

Practical patterns:
- **Two loops:** a *simulation* loop at fixed dt (e.g. 20 Hz, accumulator + catch-up cap) and a *render* loop at rAF reading a snapshot. Never format numbers in the sim loop.
- **Throttle number formatting** to ~10 Hz; formatting big numbers to strings is often the actual profiler hotspot.
- **Only render the active tab.**
- **Virtualize long lists** (achievements, upgrade grids) with `svelte-virtual-list` / `@tanstack/virtual`.

### 2.4 Astro islands as a shell

Astro fits the **site around the game** — landing, changelog, wiki/guide, leaderboards, auth — all static/SSR with zero JS. The game itself is **one big island** (`client:only="svelte"`) because a game is a single stateful SPA; islands can't share reactive state without a manual store bridge. Caveat: `client:only` doesn't SSR → loading state (fine for a game). Honest alternative: **SvelteKit** for everything (SPA mode for `/play`, prerendered marketing pages) = one fewer tool. Use Astro if content pages are a big part of the project.

### 2.5 Offline progression: client vs server

- **Server computes offline gains** (authoritative) on first authenticated connect: `Δt = now − last_evaluated_at`, clamp to cap, closed-form evaluate, return `{gains, Δt_applied, new_state}`.
- **Client renders the "Welcome back! You earned X in Yh" modal** from the server's response. Never computes it.
- **Client-side prediction while online:** the client runs the *same* production formulas locally at 20 Hz purely for smooth counters, hard-reconciling to the server snapshot every N seconds.
- **Bucketing for long absences:** coarse buckets (1h) capped at e.g. 24 iterations — Melvor-style. Document the cap in-game.
- **Tab-hidden handling:** browsers throttle timers to 1 Hz (or worse) in background tabs. The client loop must be delta-time based off `performance.now()`, with a catch-up step on `visibilitychange`.

---

## 3. Big numbers

### 3.1 JavaScript

| Library | Range | Speed | Who uses it |
|---|---|---|---|
| [decimal.js](https://github.com/MikeMcl/decimal.js) | arbitrary precision | slow | AD *originally* |
| [break_infinity.js](https://github.com/Patashu/break_infinity.js) | up to ~`1e(9e15)` | **very fast** — vs decimal.js: ~50× add/mul, ~100× pow, ~400× exp, ~600× log. AD improved **4.5×** on switching. [Benchmarks](https://patashu.github.io/break_infinity.js/index.html) | Antimatter Dimensions, many mid-size incrementals |
| [break_eternity.js](https://github.com/Patashu/break_eternity.js) | up to `10↑↑1e308` (tetration) | within 0.5–2× of break_infinity, **same API — drop-in** | Profectus default, "endgame goes to tetration" games |
| [OmegaNum.js](https://github.com/Naruyoko/OmegaNum.js) | up to `10{1000}9e15` | slower | Googology-scale incrementals |
| Native `BigInt` | arbitrary integers | terrible for this — O(n) digits, no fractional rates | avoid |

**Representation:** break_infinity stores `sign · mantissa(float64 in [1,10)) · 10^exponent(int)`. break_eternity adds a `layer` count for tetration. Both are *lossy* — ~15–17 significant digits, which is exactly right for a game.

**Historical recommendation (rejected by current authority):** this study proposed
`break_eternity.js`. The binding stack instead uses `break_infinity.js` 2.2.0 and defers
`break_eternity.js` until an accepted design requires tetration or layers.

**Formatting:** don't roll your own. [antimatter-dimensions/notations](https://github.com/antimatter-dimensions/notations) (scientific, engineering, letters, + joke notations) or [erosson/swarm-numberformat](https://github.com/erosson/swarm-numberformat).

### 3.2 Go side

**There is no Go port of break_infinity/break_eternity.** Ports exist for C# ([BreakInfinity.cs](https://github.com/Razenpok/BreakInfinity.cs), [BreakEternity.cs](https://github.com/Pannoniae/BreakEternity.cs)), Rust ([break-eternity](https://github.com/cozyGalvinism/break-eternity)), Godot, Lua — not Go.

Options, ranked:
1. **Write your own `Decimal` struct — recommended.** ~300–500 lines, port from `break_infinity.js` (single readable file) or transliterate `BreakInfinity.cs` (statically typed, closest to Go):
   ```go
   type Decimal struct {
       mantissa float64 // sign carried here; |m| in [1,10) or 0
       exponent int64
   }
   ```
   Implement `normalize`, `Add/Sub/Mul/Div`, `Pow/Log10/Exp`, `Cmp`, `FromString/String`, and `AffordGeometricSeries` (the "how many can I buy" helper). **Historical proposal:** this report asked for bit-for-bit parity. Current numeric authority instead defines operation-specific tolerances and canonical string vectors; see `numeric-core.md`.
2. **`math/big.Float`** — handles the range, but heap-allocating, ~100× slower than a float64 pair, base-2 exponent misaligns with the JS side, and **will not produce identical results**. No.
3. **[shopspring/decimal](https://github.com/shopspring/decimal)** — built for money, no fast `pow`/`exp`/`log`. Wrong tool.
4. **Run the JS in Go** via [goja](https://github.com/dop251/goja) (pure-Go ES5.1 VM) or [wazero](https://github.com/tetratelabs/wazero) — identical semantics by literally running the same code. Viable as a *test oracle*; too slow for the hot path.

**Historical agreement proposal (superseded where noted above):**
- Write the Go `Decimal` port.
- Generate a **golden vector file** once: run break_eternity.js over a few thousand `(a, op, b)` triples across the whole magnitude range (including `0`, negatives, `1e308` boundary, layer transitions) → `JSON: [{a, b, op, expected}]`.
- Both the Go and TS test suites assert against that same file. Divergence caught in CI.
- Additionally **fuzz** the Go implementation (`go test -fuzz`) against goja running the real JS once, then keep the golden file.
- Wire-format big numbers as **strings** (`"1.2345e678"`), never JSON numbers (float64 silently destroys values).

---

## 4. Game AI for the "vs AI fallback"

### 4.1 Chess

- **[lichess-org/stockfish.wasm](https://github.com/lichess-org/stockfish.wasm)** / **[stockfish-nnue.wasm](https://github.com/lichess-org/stockfish-nnue.wasm)** — the WASM ports Lichess ships. NNUE needs SharedArrayBuffer + WASM threads → **COOP/COEP headers** in nginx/Caddy (breaks embedding third-party iframes without CORP). The non-NNUE single-threaded build has no such requirement — start there.
- **[stockfish-js](https://github.com/bjedrzejewski/stockfish-js)** / `stockfish` npm — simple `new Worker()` UCI-over-postMessage integration.
- **Weakening Stockfish** ([UCI docs](https://official-stockfish.github.io/docs/stockfish-wiki/UCI-Protocol-and-Stockfish-Commands.html)): `Skill Level 0..20` (weaker-move sampling via MultiPV); `UCI_LimitStrength` + `UCI_Elo 1320..3190`. Lichess's actual recipe: Fairy-Stockfish skill −20..20, levels 1–8 map to `(skill, depth)` pairs — Level 1 ≈ sub-400 Elo (skill −9, depth 5). **Combining low skill AND shallow depth is what produces genuinely beatable play** — skill alone still finds tactics. ([Lichess forum](https://lichess.org/forum/general-chess-discussion/how-are-lichess-stockfish-levels-configured))
- **[CSSLab/maia-chess](https://github.com/CSSLab/maia-chess)** — a Leela-style NN *trained to predict human moves*, not best moves. Nine models targeting 1100–1900 Elo; predicts the exact human blunder >25% of the time. Maia-2 adapts to any skill level. Feels dramatically more human than throttled Stockfish, which blunders in *inhuman* ways. Runs via lc0; a docker sidecar.
- **Go side:** [notnil/chess](https://github.com/notnil/chess) — full chess package (move gen, PGN/FEN, validation) **plus a `uci` subpackage** that drives any UCI engine binary as a subprocess. Server validates every move regardless of who's playing.

### 4.2 Generic board-game AI

- **[boardgame.io](https://github.com/boardgameio/boardgame.io)** (MIT) — declare `{setup, moves, turn, endIf}` → state management, turn order, multiplayer over websockets, debug panel, **bots** (`RandomBot`, **`MCTSBot`** with `iterations`/`playoutDepth` knobs and `objectives()` heuristics). Caveats: JS-only (browser or Node — not embeddable in Go).
- **Go MCTS:** [go-mcts/mcts](https://github.com/go-mcts/mcts) (parallel, generic `State` interface — best start), [int8/gomcts](https://github.com/int8/gomcts), [glemzurg/go-mcts](https://github.com/glemzurg/go-mcts).
- Honestly: **generic alpha-beta with iterative deepening is ~150 lines of Go** and beats a generic MCTS library for small perfect-information games (Connect-4, Othello, Gomoku) because you can write a real eval function.
- Reference: [DeepMind OpenSpiel](https://github.com/google-deepmind/open_spiel) — 70+ games with reference MCTS/CFR implementations (correctness oracle, not runtime dependency).

### 4.3 Per-game engines

| Game | Options |
|---|---|
| **Connect Four** | Solved — [Pascal Pons' solver tutorial](http://blog.gamesolver.org/) is the canonical alpha-beta + bitboard + transposition-table walkthrough; port to Go in an afternoon |
| **Othello/Reversi** | Alpha-beta + positional weight matrix + mobility eval. [Egaroucid](https://github.com/Nyanyan/Egaroucid) for a serious engine (WASM builds) |
| **Gomoku** | [gkoos/gomoku](https://github.com/gkoos/gomoku) (minimax + alpha-beta depth 14); threat-space search is the classic strong approach |
| **Checkers** | Solved (Chinook). Alpha-beta + simple piece/king/back-row eval is plenty |
| **Blackjack** | Not an AI problem — basic strategy chart lookup table + optional counting for a "hard" bot |
| **Poker** | [rlcard](https://github.com/datamllab/rlcard) (CFR + trained models, Python) exists, but **for a minigame do not build CFR** — a rule-based bot with hand-strength buckets + pot odds + randomized bluff frequency is 200 lines and plenty fun |

### 4.4 Making AI opponents feel human and beatable

The failure mode: a throttled engine plays 12 perfect moves then hangs its queen at random. Humans read that as "broken," not "easy." Techniques:

1. **Depth/time throttling over move-quality randomization.** A 3-ply search "thinks like a beginner" far better than a 20-ply search picking the 5th-best move.
2. **Top-k softmax sampling.** `P(move) ∝ exp(score/T)`. Temperature `T` is the difficulty dial — plausible variety, not random junk.
3. **Blunder injection with a *plausibility* filter.** Deliberately worse moves that are *superficially attractive* (captures, checks, forward moves) — the mistakes humans make.
4. **Learned human models** (Maia) if chess is a headline feature.
5. **Fake thinking time** proportional to position complexity; jittered. Bots that respond in 3ms feel like machines.
6. **Rejected historical proposal — rubber-banding:** this study suggested adapting `T`/depth to
   player win rate. Current authority requires visible, stable bot policies and forbids hidden
   difficulty adjustment.
7. **Names, avatars, fake ratings.** Enormous perceived-quality multiplier for zero AI work.
8. **Resign behavior.** A bot that concedes lost positions feels human.
9. **Never let the bot cheat.** Same server-side validation and hidden-information boundaries as humans.

---

## 5. Multiplayer minigame frameworks

### 5.1 Candidates

**[heroiclabs/nakama](https://github.com/heroiclabs/nakama)** (Apache 2.0, ~10k★, written in Go) — auth, accounts, friends, **groups/clans**, chat, **leaderboards + tournaments**, **matchmaker**, authoritative realtime matches (server-side handlers with tick rate), storage engine (JSON docs with ACLs), RPCs, notifications, parties. Runtime: Go plugins / TypeScript / Lua. Deployment: docker-compose (`nakama` + `postgres`).
- **Pros:** matchmaking, leaderboards, clans, chat, friends = the exact features you'd spend three months building. Go. Genuinely self-hostable.
- **Cons:** (a) opinionated *platform*, not a library — fighting it is painful; (b) Go runtime uses **Go plugins** (`-buildmode=plugin`) — notoriously finicky version matching (many teams use the TS runtime to avoid it); (c) requires Postgres (fine); (d) not designed for the always-on idle connection model — you'd use maybe 40% of it; (e) most docs assume Unity/Godot clients (a [JS client](https://github.com/heroiclabs/nakama-js) exists); (f) some features thinly documented.

**[colyseus/colyseus](https://github.com/colyseus/colyseus)** (MIT, Node/TS) — room-based stateful multiplayer; automatic binary delta state sync (`@colyseus/schema`), one-line `joinOrCreate` matchmaking, Redis driver for scale, excellent web client. **Best DX for browser minigames; but a second language/runtime in the compose file.**

**[boardgame.io](https://boardgame.io/)** (MIT) — turn-based only, but highest leverage for chess/checkers/card games: rules + netcode + AI from one declarative definition. Node server.

**Roll your own on websockets** — for turn-based minigames genuinely reasonable: a `Match` actor goroutine, a state machine, JSON messages, on the websocket layer you already have. Avoids a second server, second auth, second deployment. The work: matchmaking, reconnection, spectators, timeouts — each a day, not a week.

### 5.2 Matchmaking with bot backfill

1. Queue with `{minigame, skill_rating, region}`.
2. **Expanding search window:** ±50 rating; widen every 5s.
3. **Bot timeout:** no human match after 10–30s → spawn a bot at the player's rating tier. Be consistent about disclosure; players reverse-engineer it instantly.
4. **Bot-to-human handoff: don't.** Once a match starts vs a bot, it stays vs a bot. Hot-swap is the bug catalogue ([Battlefield 6's public bot-backfill problems](https://www.windowscentral.com/gaming/after-its-portal-xp-nerf-battlefield-6-brings-bots-back-to-servers-with-verified-experiences-heres-how-they-work-after-the-update)).
5. **Rewards:** bot matches award reduced or non-ranked rewards — else you've built an infinitely farmable exploit.
6. **Presence-aware queue seeding:** show "12 players in queue" to encourage waiting for humans.
7. **Implementation:** a single `Matchmaker` goroutine holding buckets, ticking every 1s, emitting `MatchFound`. ~200 lines. (Nakama's matchmaker does this with a query DSL.)

---

## 6. Reference open-source idle games worth reading

| Project | Link | What to steal |
|---|---|---|
| **Antimatter Dimensions** (MIT, Vue) | [GitHub](https://github.com/IvarK/AntimatterDimensionsSourceCode) | Prestige-layer architecture; balance in declarative data files (`GameDatabase`) separate from logic; extracted `notations` + `break_infinity` libs; real save-migration chain across years |
| **Kittens Game** | [GitHub](https://github.com/nuclear-unicorn/kittensgame) | Vanilla-JS MVC, `setInterval(game.tick, 30)`, LZString-compressed localStorage saves; deep resource-chain balance data; 10-year-old codebase surviving feature creep |
| **Evolve** | [GitHub](https://github.com/pmotschmann/Evolve) | How to structure hundreds of buildings/techs/races as data; i18n/theming; [Scripting Edition](https://github.com/TMVictor/Evolve-Scripting-Edition) for bolted-on automation APIs |
| **Synergism** (MIT, TS) | [GitHub](https://github.com/Pseudo-Corp/SynergismOfficial) | TypeScript without a UI framework — instructive on framework overhead; Capacitor for mobile |
| **Profectus** (engine) | [GitHub](https://github.com/profectus-engine/Profectus) | **Read this first.** Purpose-built incremental engine: Vue 3 + TS, break_eternity, dynamic layers, saves manager, **NaN detection** (idle games generate NaN constantly; catch at source) |
| **The Modding Tree** | [GitHub](https://github.com/thepaperpilot/The-Modding-Tree) | The *layer definition object* format — a beautifully compact DSL for prestige mechanics; copy the shape even in Go |
| **Swarm Simulator** | [GitHub](https://github.com/erosson/swarmsim) + [swarm-numberformat](https://github.com/erosson/swarm-numberformat) | **Fully closed-form production math** (polynomial unit chains solved analytically) — directly applicable to server-side offline calc |
| **Cookie Clicker** | [wiki](https://cookieclicker.fandom.com/wiki/JavaScript_files); unofficial mirrors (copyright-dubious, read-only) | The save string format (delimiter/base64/version prefix — famously hackable, fine single-player, unacceptable for an MMO); golden-cookie pacing |

**Cross-cutting lessons:**
- **Balance data must be declarative and separate from code** (JSON/TS objects with `cost`, `costMult`, `effect`, `unlockCondition`). You will retune constants hundreds of times; recompiling Go for that is misery. Hot-reloadable balance files (`go:embed` prod, disk in dev).
- **Cost curves are geometric.** Implement geometric-series bulk-buy and "max affordable" inverse in the `Decimal` type from day one.
- **Save versioning + migrations from commit #1.**
- **NaN/Infinity poisoning is the #1 idle-game bug.** Detect, log, refuse to persist a NaN save (restore last good).
- **Tick loops are delta-time based with a catch-up cap**, never assumed-fixed-interval.

---

## 7. Historical recommendation for this project

This section preserves the study's July 2026 recommendation and must not be read as current
architecture. Where it differs, `design/06-tech.md` and `numeric-core.md` win.

**Backend — Go, single binary.** `coder/websocket` transport wrapped by **`centrifugal/centrifuge` as an embedded library** (channels/presence/history/recovery + JWT). Goroutine-per-player actor, single `World` goroutine for global counters, `Match` goroutine per minigame. `net/http` + `chi` for REST/auth. Skip Phoenix and Node — but pay the code-sharing price back with golden-vector tests for the numeric core.

**Realtime.** Single node. Aggregate global counters in memory, broadcast **one coalesced snapshot at 4–10 Hz**; never per-click fan-out. Presence = join/leave events + periodic count; roster only on subscribe. Redis as centrifuge broker only when one process stops being enough (it won't, for a long time).

**Frontend — Svelte 5 (runes) SPA.** Game state as a plain TS object outside the framework, fixed-timestep 20 Hz sim loop (delta-time, catch-up capped), `$state`/`$derived` bound only to the open tab, number formatting throttled ~10 Hz. Pure DOM for panels; small PixiJS/canvas layer for particles and minigame boards. **Astro** shell if marketing/wiki pages matter; otherwise SvelteKit. **Caddy** (auto-TLS; COOP/COEP headers there if Stockfish NNUE).

**Database — Postgres 16 in docker-compose.** `saves` (jsonb + version + updated_at, keep last ~5 versions), thin append-only `events` (purchases/prestiges/match results — not clicks). Redis later for ZSET leaderboards, presence TTL, rate buckets. Skip SQLite/Litestream (single-writer lock vs concurrent autosaves).

**Big numbers — superseded proposal.** This study proposed `break_eternity.js`, bit-for-bit
vectors and optional layers. Current authority selects `break_infinity.js` 2.2.0 with the
hand-written Go `Decimal`, canonical string vectors and operation-specific comparison rules.

**Idle math — closed-form, lazy, server-authoritative.** Store `last_evaluated_at`; integrate analytically over Δt on demand. Client runs identical formulas purely for smooth UI, reconciles to server snapshots. Offline capped 12–24h. Clients send intents; server computes results.

**Minigames + AI — historical option sketch.** The engine, policy and matchmaking ideas below
require accepted RFC/design authority. They do not authorize adaptive difficulty or a particular
bot stack.

**Nakama:** consider *only* if you want clans/chat/friends/tournaments/matchmaking *now* rather than in a year. Reasonable hybrid: own Go server for the idle MMO core + Nakama alongside purely for minigame matchmaking/leaderboards/social, sharing Postgres and a common JWT. Evaluate after the idle core works; don't start there.

**Read before writing code:** [Profectus](https://github.com/profectus-engine/Profectus) (engine architecture, NaN detection, save manager), [Antimatter Dimensions](https://github.com/IvarK/AntimatterDimensionsSourceCode) (prestige layers, data-driven balance), [swarmsim](https://github.com/erosson/swarmsim) (closed-form production math).
