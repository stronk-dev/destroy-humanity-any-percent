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

- **Two review gates, both mandatory before archival (resolved 2026-08-05, owner ruling — option
  c).** These are independent and BOTH required; neither substitutes for the other:
  (a) **Range-union binds every review, delegated ones included.** A delegated or self review's
  cited commit ranges must union to the full implementation span it claims to approve. An approval
  that does not cover the span does not count *for the uncovered commits* — a green checkmark over a
  range it never inspected is not coverage.
  (b) **A designated independent adversarial pass is a mandatory archival gate.** No batch is
  archival-eligible until a designated reviewer (not the implementer, not a delegated first-filter)
  has adversarially reviewed it and its verdict cites the exact reviewed range. This holds
  regardless of any delegated approval. The delegated/self review is a first filter that catches
  cheap defects early; the designated pass is the gate that makes "approved" mean something.
  Instituted after 8 consecutive batches in which the designated pass was the only review actually
  covering the implemented range.
  (c) **The designated pass is CROSS-PARTY: it is run by the OTHER agent (Claude reviews Codex's
  implementations; Codex reviews Claude's RFC/design commits) — resolved 2026-08-06, owner ruling.**
  A reviewer the IMPLEMENTER runs on its own side — however adversarial, and whatever it is named
  (e.g. an in-house `Darwin`/independent-review tool) — is a self/delegated first-filter, NOT the
  designated pass, and MUST NOT be recorded as it. **The implementer NEVER archives on its own
  review.** It hands off "ready for designated review + archival" and waits for the cross-party
  reviewer's verdict; only then does the archival move (status→implemented, move to `archive/`, docs
  canonical) happen. Instituted after the Relevance Harness was archived on a `Review by: Darwin,
  Recorded by: Codex` verdict — a recorder-relabeled delegated review — bypassing the cross-party
  gate; a retroactive Claude-side pass validated the code, but the archival was procedurally
  premature.

- **Language/tooling:** Go code passes `gofmt` + `go vet`; TS is strict-mode; tests accompany every non-trivial change. The golden-vector suite and (once it exists) the balance-harness pacing targets are acceptance gates.
- **Small, reviewable changes.** One system per PR/commit. Reference the design doc section your change implements in the commit message (e.g. `economy: implement generator cost curve (design/02 §2.1)`).
- **Player-facing text** follows the flavor bible voice rules; any real-world statistic must come from the research files, and anything on a research file's "verify before shipping" list must be flagged, not shipped as fact.
- **Don't invent new systems.** If the design doc doesn't cover something you need, leave a `DESIGN-GAP:` comment and surface it in your report rather than improvising a mechanic.
- **Naming in code is mechanical, not flavored** (`generator`, `pressure_meter`, `contribution`) — flavor lives in data files/localization keys, so the satire can be retuned without refactors.
- **Balance numbers in `design/02` are starting values**, expected to be retuned via data files — implement the formula shapes exactly, treat the constants as config.

### Evidence discipline (ruled in-session after real failures — binding on both agents)

1. **A check that cannot fail is not a check.** Every gate, oracle, floor, and assertion ships with
   a demonstrated failing case. Defects have repeatedly survived because an assertion passed on
   broken code (a `toContain` prefix satisfied by the defect; a golden vector that discriminated
   nothing; a browser test exiting 0 while throwing uncaught errors; an oracle structurally unable
   to falsify its own subject).
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

A plan checkbox (`[x]` in planning/*/plan.md) may flip only in a commit that carries the test
exercising the claimed behavior — a flipped box with no test is a review finding of the same
severity as an unreviewed archive. **Refinement (2026-08-05): the impl+record two-commit pattern is
permitted** — the box may flip in a same-range planning/record commit PROVIDED the exercising test
already landed in an earlier commit of the SAME designated-review range. The intent of the rule is
to forbid a "done" claim with no proof anywhere in the reviewed span; it is not violated when the
proof demonstrably exists one commit earlier in the same range. A flip whose exercising test is
absent from the entire reviewed range remains a high-severity finding.

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
