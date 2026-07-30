# Agent Onboarding — Cloud Clicker

You are working on **Cloud Clicker**: a free, browser-based MMO idle game — *"speedrun any% destroy humanity"* — climbing from a 1995 garage to a world-consuming AI megacorp. Satirical, Cookie-Clicker-lineage, with an MMO layer and a designed ending.

## Repo structure — four tiers (defined in `rfc/0000-rfc-process.md`, read it)

| Tier | Dir | Role |
|---|---|---|
| Design | `design/` | Intent + research. Where ideas come from. NOT the implementation spec. |
| **RFC** | `rfc/` | **Active implementation specs. You implement RFCs, nothing else.** Implemented RFCs move to `rfc/archive/`. |
| Planning | `planning/<rfc-slug>/` | Your `plan.md` + append-only `log.md` per RFC. The long-term job log — write it so a fresh agent can resume from it alone. |
| Docs | `docs/` | Canonical description of what exists. Update in the same change as any behavior change. |

**Current state:** design complete; the numeric core, economy kernel, save layer, production
engine, balance-harness foundation, gate/Route Registry foundation, and Commons server foundation
and the Client Shell/Worker foundation are implemented and archived. See `rfc/README.md` for
active work.

## Your workflow

1. You are assigned an **active RFC** (check `rfc/README.md` for status).
2. Create/resume `planning/<rfc-slug>/`: write `plan.md` (task breakdown + acceptance gates from the RFC), append to `log.md` every session.
3. Implement. Missing spec = `DESIGN-GAP` in the log + propose a draft RFC; never improvise mechanics.
4. Done = acceptance criteria green + `docs/` updated + RFC status set to `implemented` +
   RFC and planning directories moved to their archives — all in the final change. After that,
   `docs/` is canonical; archived RFCs are history.

## Reading order (minimum to be productive)

1. `rfc/0000-rfc-process.md` — the process (5 min).
2. Your assigned RFC, fully, including its design refs.
3. `design/00-vision.md` — pitch, pillars, anti-goals.
4. `design/06-tech.md` — stack and architecture decisions. **Binding.**
5. Skim `design/08-satire-flavor.md` §1 (voice rules) before writing ANY player-facing text.

## Non-negotiable design laws

These are settled. Do not "improve" them without explicit sign-off from Marco:

1. **No real money, ever, in any direction.** No IAP, no ads, no telemetry beyond gameplay. Free-ness is the satire's foundation.
2. **Server-authoritative.** Clients send intents, never results. Production math is closed-form and lazy (`06-tech.md §idle-math`) — never a per-player server tick loop.
3. **Big numbers:** `break_infinity.js` 2.2.0 (client) + the hand-written Go `Decimal` (server). Wire format is **strings**. Any change to either implementation must keep the shared golden-vector tests green in BOTH test suites. `break_eternity.js` is deferred until an accepted design actually requires tetration/layers.
4. **Balance data is declarative** (JSON/YAML data files, hot-reloadable), never constants in code.
5. **Hardcaps, never softcaps.** Every cap is a visible number.
6. **Every multiplayer feature has an AI/bot fallback**; bots never cheat (same server validation and hidden-info boundaries as humans).
7. **Offline progress is default** (90% rate, 24 h cap — canonical behavior:
   `docs/production-engine.md`) — never a purchased privilege.
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

- **Per-change review is mandatory, both directions.** Every implementation batch gets a recorded
  diff review in its planning log *before* the next RFC acceptance — Claude reviews Codex's
  batches, Codex reviews Claude's RFC/design commits (it already does). Milestone agent-audits
  complement this; they do not replace it. A batch whose review found nothing still gets a
  recorded "approved" line — the absence of a review entry means the review didn't happen, not
  that it passed. (Instituted 2026-07-29 after the R1 batch initially received only a
  spot-check; the review it then got found a latent cap-lowering policy gap the spot-check
  missed.)

- **Language/tooling:** Go code passes `gofmt` + `go vet`; TS is strict-mode; tests accompany every non-trivial change. The golden-vector suite and (once it exists) the balance-harness pacing targets are acceptance gates.
- **Small, reviewable changes.** One system per PR/commit. Reference the design doc section your change implements in the commit message (e.g. `economy: implement generator cost curve (design/02 §2.1)`).
- **Player-facing text** follows the flavor bible voice rules; any real-world statistic must come from the research files, and anything on a research file's "verify before shipping" list must be flagged, not shipped as fact.
- **Don't invent new systems.** If the design doc doesn't cover something you need, leave a `DESIGN-GAP:` comment and surface it in your report rather than improvising a mechanic.
- **Naming in code is mechanical, not flavored** (`generator`, `pressure_meter`, `contribution`) — flavor lives in data files/localization keys, so the satire can be retuned without refactors.
- **Balance numbers in `design/02` are starting values**, expected to be retuned via data files — implement the formula shapes exactly, treat the constants as config.

### Routine command authority

Routine local development commands are pre-authorized. Agents should run them directly and in
useful batches instead of asking for permission one invocation at a time:

- formatting and generation (`gofmt -w`, repository format/generation targets);
- local verification (`go test`, `go vet`, `pnpm` checks/tests/builds, and `make` verification
  targets such as `make verify`);
- non-destructive Git bookkeeping (`git status`, `git diff`, `git add`, and intentional
  intermediate `git commit`s).

History rewriting (rebase/amend of committed work) is permitted for exactly one purpose: correcting
a protocol-violating commit subject (`BALANCE-CHANGE:`/`CONSTANTS-IDENTITY:` classes) — and only
while no review verdict references the affected hashes and nothing has been pushed. Once a
planning-log verdict cites a hash, that history is append-only; a wrong-subject commit discovered
after that point gets a follow-up correction commit and a planning-log ruling, never a rewrite.

Use stable, narrowly scoped command prefixes so the execution environment can remember approval.
Run every routine format, build, test, and Git command with the repository root as its working
directory. Do not create task-specific cache directories or move into package subdirectories just
to run tooling; use the root Make targets and their package selectors.
Do not wrap an otherwise approved command in a custom shell, environment assignment, or compound
command unless the wrapper is actually required; wrappers often defeat persistent prefix approval.
Group related checks into the repository's existing `make`/package targets where practical.
The Makefile exports a repository-local ignored `GOCACHE`; use `make test-go` or
`make test-go GO_PACKAGES='./harness ./transport'` for Go tests so normal compilation never needs
permission to write into a user-level cache. Postgres integration runs through the declared Docker
service: `docker compose -f compose.save-test.yml run --rm test` (the Make target is the human-facing
alias). It owns its test URL, network, and Go caches, so agents never wrap tests in environment
assignments, run them from `server/`, or request host-network approval. Focus ordinary non-Postgres
runs through `GO_PACKAGES`/`GO_TEST_FLAGS` on the root Make target.

If sandbox or network access genuinely requires escalation, request a reusable narrow prefix for
that class of command once, then continue. Approval is still required for destructive operations,
external publication or deployment, secret access, and commands outside the repository's normal
development surface. **Never push, publish, deploy, or open a PR unless the user explicitly asks.**

## Where to start

The numeric, economy, save, production, harness, route, Commons, and client-shell foundations are
implemented and archived. Choose work only from the active index in `rfc/README.md`; draft and
accept missing Phase-0 contracts before starting from the roadmap.
