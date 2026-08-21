# Cloud Clicker

[![CI](https://github.com/stronk-dev/destroy-humanity-any-percent/actions/workflows/ci.yml/badge.svg)](https://github.com/stronk-dev/destroy-humanity-any-percent/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

> Speedrun any% destroy humanity.

Cloud Clicker is a free, browser-based MMO idle game in development. You begin in a 1995 garage
and build an AI company toward a deliberately finite ending while the interface and institution
around it gradually enshittify.

There are no ads, purchases, or real-money mechanics. The satire only works if the game is not
selling the systems it mocks.

## What we're building

- An idle game with a designed ending, where prestige loops are routes through the story rather
  than a substitute for finishing it.
- Nine technology eras whose mechanics and interface change as the company grows from honest
  shareware to a world-consuming megacorp.
- Active and idle play as different builds, with generous offline progress available to everyone.
- Server-wide milestones, guilds, shared events, leaderboards, and minigames, each with an honest
  AI fallback when no other player is available.
- Server-authoritative simulation with published community formulas, visible hardcaps, and
  declarative balance data.
- Parodies of modern dark patterns with the curtain pulled all the way back—and a sincere pet who
  is never part of the joke.

## Project status

Cloud Clicker is pre-1.0 and under active development. The repository contains a tested T0–T1
vertical slice and substantial server foundations, including:

- deterministic big-number arithmetic shared between Go and TypeScript;
- server-authoritative economy, production, offline progress, saves, and replay;
- account/session and WebSocket transport backends;
- declarative content, routes, achievements, guild, pet-care, and minigame infrastructure; and
- a Svelte client that completes a real Chromium bootstrap through PostgreSQL and WebSocket into
  the Desk screen.

The complete first-hour browser workflow, tiers 2–8, most minigames, player-facing MMO and world
systems, production packaging, backup and restore, and the designed endings are not complete.
APIs, schemas, and content may change before 1.0.

Current status is maintained in the checked project artifacts rather than copied into this README:

- [Current-state brief](planning/CURRENT-STATE.md)
- [Capability map](planning/platform-alignment/capability-map.md)
- [Execution queue](planning/platform-alignment/execution-queue.md)
- [Active RFC register](rfc/README.md)
- [Implemented-system documentation](docs/README.md)

## Development setup

Requirements:

- Go 1.26 or newer
- Node.js 22 or newer
- pnpm 11.15.1
- GNU Make
- Docker with Compose support for PostgreSQL and composed browser tests

Clone the repository and install the pinned client dependencies and browser versions:

```sh
git clone https://github.com/stronk-dev/destroy-humanity-any-percent.git cloud-clicker
cd cloud-clicker
make setup
```

Build the backend and client:

```sh
make build-gameserver
make build-client
```

Build artifacts are written beneath `.cache/` and `client/dist/`.

The repository does not yet provide a supported production Docker/Caddy bundle or a one-command
self-host deployment. The Compose files at the repository root are test infrastructure. The
current backend's required environment, startup, readiness, and shutdown behavior is documented in
[Gameserver Composition](docs/gameserver.md).

## Verification

Run the standard local aggregate from the repository root:

```sh
make verify
```

Use the Docker-backed lanes when claiming PostgreSQL or complete browser-to-server evidence:

```sh
make test-go-ci
make test-game-ui-composed
```

Some balance-harness checks are intentionally long-running. A green unit package or ordinary host
run does not prove that dependency-conditional and composed workflows executed. See
[Continuous Integration](docs/ci.md) for the complete verification contract and focused commands.

## Architecture

| Layer | Stack |
|---|---|
| Server | Go, Chi, embedded Centrifuge, `coder/websocket` |
| Client | Svelte 5, TypeScript, Vite, Web Worker prediction loop |
| Persistence | PostgreSQL 16, Goose migrations |
| Numeric model | `break_infinity.js` 2.2.0 and a matching Go Decimal implementation |
| Testing | Go test, Vitest, Playwright, Docker Compose |

The server is authoritative. Clients submit intents and predict presentation state; they do not
commit gameplay results. Production is calculated lazily from closed forms rather than by running
a permanent server tick for every player.

| Path | Responsibility |
|---|---|
| [`server/`](server/) | Gameplay services, persistence, migrations, and commands |
| [`client/`](client/) | Svelte application, prediction worker, generated API types, and browser tests |
| [`balance/`](balance/) | Declarative gameplay content, schemas, fixtures, and harness inputs |
| [`copy/`](copy/) | Governed player-facing copy catalogs and generated artifacts |
| [`testdata/`](testdata/) | Shared cross-runtime vectors and fixtures |
| [`docs/`](docs/) | Canonical documentation for implemented behavior |
| [`design/`](design/) | Product intent and research-backed design |
| [`rfc/`](rfc/) | Active implementation specifications and archived RFC history |
| [`planning/`](planning/) | Per-RFC plans, append-only logs, and repository alignment records |

## Design boundaries

Cloud Clicker is not a conventional endless mobile tycoon with its purchases removed. It is a
finite, replayable satire built around several non-negotiable boundaries:

- no real money, advertising, NFTs, or paid offline progress;
- no hidden softcaps or client-authored gameplay outcomes;
- no multiplayer system without a non-cheating, clearly identified bot fallback;
- no punishment for leaving an idle game idle; and
- no dark-pattern parody that conceals what it is doing.

The complete product thesis is documented in [the vision](design/00-vision.md).

## Contributing

The project uses an evidence-first, RFC-driven workflow. Before a substantial change, read
[AGENTS.md](AGENTS.md), [the RFC process](rfc/0000-rfc-process.md), and the relevant design and
canonical documentation. Product implementation requires an accepted RFC; missing specification
is recorded as a design gap rather than improvised in code.

Changes should include the smallest discriminating test, update canonical documentation when
behavior changes, and pass the relevant verification gates. See the
[active RFC register](rfc/README.md) for current work.

## License

Cloud Clicker is available under the [MIT License](LICENSE).
