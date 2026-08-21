# Cloud Clicker

> Speedrun any% destroy humanity.

Cloud Clicker is a free, browser-based MMO idle game in development. Starting in a 1995 garage,
the player builds an AI company toward a deliberately finite ending while the interface and
institution around it gradually enshittify.

There are no ads, purchases, or real-money mechanics.

## Project status

Cloud Clicker is not a 1.0 release. The repository currently contains a tested T0–T1 vertical
slice and substantial server foundations.

Implemented foundations include:

- deterministic big-number arithmetic shared between Go and TypeScript;
- server-authoritative economy, production, offline progress, saves, and replay;
- account/session and WebSocket transport backends;
- declarative content, routes, achievements, guild, and minigame infrastructure; and
- a Svelte client that completes a real Chromium bootstrap through Postgres and WebSocket into the
  Desk screen.

The complete first-hour browser workflow, tiers 2–8, most minigames, player-facing MMO/world
systems, production packaging, backup and restore, and the designed endings are not complete.

For exact current evidence and blockers, see the [current-state brief](planning/CURRENT-STATE.md),
[capability map](planning/platform-alignment/capability-map.md), and
[execution queue](planning/platform-alignment/execution-queue.md).

## Technology

| Layer | Stack |
|---|---|
| Server | Go, Chi, Centrifuge, `coder/websocket` |
| Client | Svelte 5, TypeScript, Vite, Web Worker prediction loop |
| Persistence | PostgreSQL 16, Goose migrations |
| Numeric model | `break_infinity.js` 2.2.0 and a matching Go Decimal implementation |
| Testing | Go test, Vitest, Playwright, Docker Compose |

The server is authoritative. Clients submit intents and predict presentation state; they do not
commit gameplay results.

## Requirements

- Go 1.26+
- Node.js 22+
- pnpm 11.15.1
- GNU Make
- Docker with Compose support for PostgreSQL and composed browser tests

## Setup

Install the pinned client dependencies and the Chromium, Firefox, and WebKit versions used by the
browser suite:

```sh
make setup
```

Build both applications:

```sh
make build-gameserver
make build-client
```

Build artifacts are written beneath the repository-local `.cache/` and `client/dist/` directories.

## Verification

Run the standard local aggregate:

```sh
make verify
```

An ordinary host Go run can skip tests that require PostgreSQL. Use the declared Docker lanes when
claiming database or complete composed evidence.

| Command | Purpose |
|---|---|
| `make test` | Run the default Go, client, and cross-browser suites |
| `make test-go-ci` | Run the complete Go population against PostgreSQL in Docker |
| `make test-save-integration` | Run focused PostgreSQL integration tests |
| `make test-game-ui-composed` | Run Chromium through the composed Vite, gameserver, and PostgreSQL stack |
| `make typecheck` | Run strict TypeScript and Svelte checks |
| `make verify-schema` | Validate production catalogs and negative fixtures |
| `make verify-client-boundary` | Enforce client prediction and mutation boundaries |
| `make api-check` | Regenerate and diff API artifacts and compatibility pins |
| `make formulas-check` | Regenerate and diff the published formula artifact |
| `make vectors-check` | Regenerate and diff shared numeric vectors |
| `make publication-authority-check` | Verify the public/private research boundary and its negative controls |

Some harness checks are intentionally long-running. A green unit package or default host run is
not evidence that dependency-conditional or integrated workflows executed.

## Running the server

`server/cmd/gameserver` is the current runnable backend. It requires PostgreSQL and cryptographic
configuration; startup migrates the database and validates the active content epoch before raising
readiness.

Build it with `make build-gameserver`. Required environment variables, endpoints, startup behavior,
and shutdown semantics are documented in [Gameserver Composition](docs/gameserver.md).

The repository does not yet provide a supported production Docker/Caddy bundle or one-command
self-host deployment. The Compose files at the repository root are test infrastructure.

## Repository layout

| Path | Purpose |
|---|---|
| [`server/`](server/) | Go services, gameplay engine, persistence, migrations, and commands |
| [`client/`](client/) | Svelte application, prediction worker, generated API types, and browser tests |
| [`balance/`](balance/) | Declarative gameplay content, schemas, fixtures, and harness inputs |
| [`copy/`](copy/) | Governed player-facing copy catalogs and generated artifacts |
| [`testdata/`](testdata/) | Shared cross-runtime vectors and fixtures |
| [`design/`](design/) | Product intent and research-backed design |
| [`rfc/`](rfc/) | Active implementation specifications and archived RFC history |
| [`planning/`](planning/) | Per-RFC plans, append-only logs, and repository alignment records |
| [`docs/`](docs/) | Canonical documentation for implemented behavior |

## Development process

Behavior changes are made from accepted RFCs, with a plan, executable acceptance evidence,
canonical documentation, and cross-party review. Start with:

1. [AGENTS.md](AGENTS.md)
2. [RFC-0000](rfc/0000-rfc-process.md)
3. [the active RFC index](rfc/README.md)
4. [the current execution queue](planning/platform-alignment/execution-queue.md)

Do not infer feature completion from a package, route, component, or green test alone. The required
trace is intent → producer → consumer → current data/workflow → discriminating executable proof.

## License

Cloud Clicker source code is available under the [MIT License](LICENSE).
