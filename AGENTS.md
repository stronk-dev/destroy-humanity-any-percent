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
2. Resume the owning `planning/<rfc-slug>/plan.md`. Choose a coherent change and the smallest
   meaningful verification for its risk under RFC-0000's delivery procedure. A routine fix does
   not need a new RFC, research program or separately committed predeclaration.
   Use its “Choosing the actual checks” table for existing commands; test the changed outcome,
   not the tracker. Iterate focused, then verify the finished batch's affected boundaries.
3. Implement. Missing spec = `DESIGN-GAP` in the log + propose a draft RFC; never improvise mechanics.
4. Record one concise batch outcome in the owning log. Update canonical docs when behavior
   changes; update other tracking only when its status or next action actually changes.
   Feature completion requires its acceptance criteria and designated review before archival.
   Passing a focused check or updating a tracker does not complete a feature.

## Reading order (minimum to be productive)

1. `rfc/0000-rfc-process.md` — the process (5 min).
2. Your assigned RFC, fully when taking up a new feature; the relevant contract sections when
   resuming a bounded fix. Read its relevant design references, not every research dossier.
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

- **Review coherent batches, not paperwork commits.** Behavior-changing batches and normative
  specification amendments require designated cross-party review before acceptance or archival.
  Claude reviews Codex implementations; Codex reviews Claude implementations/specifications.
  Related fixes, tests and records may share one exact review range. Editorial docs and routine
  log updates do not each require a separate review. See RFC-0000 for risk-based verification.

- **Rulings reconcile the body, not just append.** When an RFC's rulings block resolves a blocker
  that contradicts the specification body, the SAME edit must fix the body text — a normative
  section left contradicting its own accepted ruling blocks implementation and reads as an
  unresolved conflict. A status line may claim "body reconciled" only when no normative section
  contradicts a ruling (blocker-record text quoting the original defect is exempt and expected).

- **Review provenance is explicit.** Every verdict entry names both `Review by:` (the person or
  agent that actually inspected the diff) and `Recorded by:` (when someone else transcribed or
  summarized it). A recorder may not relabel a delegated or self-review as the project's
  designated independent review. An archival gate cites the exact verdict entry and reviewed
  commit range it consumed. Those cited ranges must union to the full implementation span being
  archived; uncovered edge commits remain unreviewed even when later dependent commits passed.

- **Independent review remains an archival gate.** The implementer checks its own diff and runs
  relevant tests, but cannot call that designated approval or archive on it. The other party's
  verdict must cover the complete implementation range, including tests and closeout records;
  multiple verdicts may cover that range together. No uncovered commits or recorder-relabeled
  self-reviews. Existing outstanding reviews may be consolidated, never silently marked approved.

- **Language/tooling:** Go code passes `gofmt` + `go vet`; TS is strict-mode; tests accompany every non-trivial change. The golden-vector suite and (once it exists) the balance-harness pacing targets are acceptance gates.
- **Small, reviewable changes.** One system per PR/commit. Reference the design doc section your change implements in the commit message (e.g. `economy: implement generator cost curve (design/02 §2.1)`).
- **Player-facing text** follows the flavor bible voice rules; any real-world statistic must come from the research files, and anything on a research file's "verify before shipping" list must be flagged, not shipped as fact.
- **Don't invent new systems.** If the design doc doesn't cover something you need, leave a `DESIGN-GAP:` comment and surface it in your report rather than improvising a mechanic.
- **Naming in code is mechanical, not flavored** (`generator`, `pressure_meter`, `contribution`) — flavor lives in data files/localization keys, so the satire can be retuned without refactors.
- **Balance numbers in `design/02` are starting values**, expected to be retuned via data files — implement the formula shapes exactly, treat the constants as config.

### Evidence discipline (ruled in-session after real failures — binding on both agents)

1. **Test the changed behavior.** A bug fix needs a regression that reproduces the bug on the
   old behavior and passes on the correction. New behavior needs representative success and
   failure cases. Use deliberate severing for a new critical acceptance oracle when ordinary
   regression/negative cases do not establish discrimination, not for every assertion or edit.
2. **Run it, don't read it.** Reviews that bypassed a check and executed the thing caught an
   architecture-dependent numeric divergence, a cache-masked red baseline, and a vacuous
   acceptance oracle. Use `-count=1` for any gate claim; warm caches have hidden a red tree.
3. **Fail loud; never degrade quietly.** A measurement that silently truncates, coasts, or
   excludes is worse than no measurement because it looks like data. Instrument artifacts
   (exclusions, guard exhaustion, truncation) are first-class visible fields, and a run that
   terminates on a guard rather than its objective is an invalid measurement that must fail.
4. **Budgets and bounds come from measurement, never from convenience.** Never raise a ceiling
   from an incomplete run; never loosen an acceptance bound to make a gate pass.
5. **The ruling author reconciles their own stale text.** When a review finds an owner ruling's
   normative text stale or self-contradicting, the implementer files the finding and waits — the
   author edits. Appending "this is superseded" beside unchanged body text does not satisfy the
   body-reconciliation rule.
6. **Owner-authored content is owner-authored.** Implementers may not edit ruled copy text;
   detector-forced rewrites come back for explicit adoption.

### Routine command authority

Routine local development commands are pre-authorized. Agents should run them directly and in
useful batches instead of asking for permission one invocation at a time:

- formatting and generation (`gofmt -w`, repository format/generation targets);
- local verification (`go test`, `go vet`, `pnpm` checks/tests/builds, and `make` verification
  targets such as `make verify`);
- non-destructive Git bookkeeping (`git status`, `git diff`, `git add`, and intentional
  intermediate `git commit`s).

A behavior-completion checkbox needs a named executable test and an actual passing result for
the reviewed revision. Existing tests may be reused; do not add a duplicate test merely to put
test bytes in the checkbox commit. New or corrected tests belong in the reviewed batch. Task
completion is not independent review, release approval or proof of a larger workflow.

History rewriting (rebase/amend of committed work) is permitted for exactly one purpose: correcting
a protocol-violating commit — a wrong subject (`BALANCE-CHANGE:`/`CONSTANTS-IDENTITY:`
classes), OR an unpushed commit that forces a false version signal (a behavior-identical change to
a kernel-watched file, which can neither honestly bump nor pass the guard) — and only
while no review verdict references the affected hashes and nothing has been pushed. Once a
planning-log verdict cites a hash, that history is append-only; a wrong-subject commit discovered
after that point gets a follow-up correction commit and a planning-log ruling, never a rewrite.

Applied migrations are append-only. Once a migration has landed in a commit, corrections use a
new migration; do not edit its Up or Down body in place, even before publication.

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

For the long-term 1.0 objective, read `planning/roadmap-1.0.md` and its append-only checkpoint log.
That board tracks the whole product through Transcendence; it never substitutes for an accepted
RFC or the current per-RFC implementation plan.

## Procedure authority

Marco's 2026-10-07 direction replaces the old per-edit audit/predeclaration/severing and mirrored
record routine with RFC-0000's risk-based delivery procedure. It applies to both agents
(`CLAUDE.md` links here). Keep historical evidence intact; do not waive a feature's explicit
acceptance criteria, author/owner decisions, existing review debt or the full nine-tier 1.0 floor.
