# RFC: CI Baseline

- **Status:** implementing
- **Author:** Marco (drafted by Claude, revised by Codex)
- **Created:** 2026-07-28
- **Design refs:** `design/07-roadmap.md` Phase 0, `design/12-content-pipeline.md §7`
- **Depends on:** RFC-0001 and RFC-0002 (implemented)
- **Planning:** `planning/scaffolding-and-ci/` (once implementing)

## Summary

This RFC establishes the smallest blocking CI baseline: fast deterministic gates on hosted GitHub
Actions under a strict latency budget, with exhaustive balance evidence isolated in a bounded
scheduled/manual workflow.

## Motivation

At adoption, the RFC-0001 Go, Node, and browser vector suites already existed but did not run
remotely. This was the cheapest closure of a foundational obligation. Later exhaustive balance evidence grew beyond the
blocking latency budget; D-014 retains hosted runners and separates that evidence from push/PR
gating instead of ratifying repeated 30-minute cancellations.

This scope is intentionally executable now. Deployment and policy gates without source artifacts
are not smuggled into the baseline. CI schedules the harness-owned exhaustive command but does not
change its scenarios, balance, bounds, or acceptance semantics.

**Out of scope, with owners:** deployment, Compose, Caddy, draining, migrations, and reconnect
testing (a future deployment-and-draining RFC once those components exist); balance-harness
semantics (its own RFCs); policy gates for assets and player-facing content (added when those inputs first
land); production hot-reload semantics (`design/12`, currently carrying a `DESIGN-GAP:`); the save
layer (its own RFC); leaderboard integrity and Balance Epoch enforcement (leaderboard RFC).

## Specification

### D1 — Runner topology and repository visibility

**NORMATIVE: the repository is public and all CI runs on hosted GitHub Actions.** No Komodo and
no self-hosted runners. CD topology is outside this RFC.

Komodo is a CD orchestrator, not a CI replacement, and would add MongoDB plus a low-bus-factor
dependency to a one-node deployment.

Marco selected a public repository on 2026-07-28. The cost and timing model therefore
uses public-repository hosted runners: four vCPUs and no billed Actions-minute quota. Repository
visibility does not substitute for an explicit license.

### D2 — Fast blocking workflow

`.github/workflows/ci.yml` runs only on push and pull request:

| Job | Runs | Blocking | Notes |
|---|---|---|---|
| `server` | `make verify-server-core` → cold Go tests outside the harness package, generation checks and `go vet` | yes | Uses the checked-in Go toolchain declaration and real Postgres |
| `harness` | `make verify-harness-fast` → cold package tests, role activation, Commons invariance and repository-history guards | yes | Excludes multi-million-transition pacing/relevance execution |
| `client` | `make verify-client` → strict TypeScript and Node Vitest | yes | Installs from the frozen pnpm lockfile |
| `browser` | `make test-browser` across Chromium, Firefox and WebKit, then isolated Chromium performance | yes | Installs the pinned Playwright browser versions explicitly |
| `game-ui-composed` | `make test-game-ui-composed` | yes | Real Chromium → Vite → gameserver → Postgres/WebSocket bootstrap-to-Desk witness |
| `schema` | `make verify-schema` → validate schema documents and every checked-in balance catalog | yes | An explicit empty catalog set succeeds; malformed checked-in catalogs fail |

Workflow actions use the current supported majors at acceptance: `actions/checkout@v6`,
`actions/setup-go@v6`, `actions/setup-node@v6`, `actions/cache@v5`, and
`pnpm/action-setup@v6`. Go reads `server/go.mod`; Node uses version 24; pnpm reads the exact
`packageManager` field. Browser jobs install the versions pinned by the Playwright package and
lockfile on the ordinary hosted runner.

Schema verification uses the pinned Ajv 2020-12 implementation in the client toolchain. Compiling
the schema validates it against its meta-schema. The command validates every production catalog,
requires all positive fixtures to pass, and requires all negative fixtures to fail.

The existing aggregate `make verify` remains the local exhaustive all-gates command. CI calls the
narrow fast targets so multi-million-transition evidence and browser tests are not duplicated in
the blocking path.

The workflow has least-privilege read-only repository permissions, cancels superseded runs on the
same branch, and caches Go and pnpm dependencies. Generated build outputs and browser binaries are
not cached.

### D3 — Gate tiers, maintenance workflow, and latency budget

| Tier | Trigger | Blocking | Contents |
|---|---|---|---|
| **1 — Baseline** | every push/PR | yes | D2's six jobs, including the fast harness guard |
| **2 — Policy** | every push/PR as inputs exist | yes | Current copy/formula/schema/boundary guards; later asset-provenance gates land with their inputs |
| **3 — Balance evidence** | schedule/manual | no PR block | Exhaustive pacing/relevance check with a validated, always-uploaded observation artifact |
| **4 — Numeric maintenance** | schedule/manual | no PR block | 30-second numeric fuzzing and deterministic vector-regeneration diff |

**NORMATIVE: the complete blocking workflow must finish in under five minutes.** If it exceeds the
budget, slower work moves to a non-blocking tier; the blocking budget does not grow.

`.github/workflows/maintenance.yml` owns tiers 3–4. The exhaustive harness command receives SIGINT
at 50 minutes inside a 55-minute job so the observation recorder can persist an explicit incomplete
artifact before Actions kills the runner. A successful run must validate the completed artifact;
success or failure must upload it, and a missing artifact is itself a failure. Maintenance Go jobs
disable setup-go's build/test cache and cache only downloaded modules.
The observation upload uses the current supported `actions/upload-artifact@v7` major.

### D4 — Deliberately separate performance repair

The geometric `MaxAffordable` repair changes implemented RFC-0002 behavior and belongs to the
implemented follow-up RFC `archive/geometric-afford-fast-path.md`. That RFC owns the exact-integer cap, postcondition
tests, and benchmark. This CI RFC only benefits from the repaired runtime when the future harness
arrives.

### D5 — Policy seams, not empty theatre

This RFC does not create placeholder asset manifests, Vale rules, deploy files, or formula
generators before their source material exists. Each gate lands with the first artifact it can
truthfully validate. In particular, future asset CI can prove attestation coverage, not that a
human-origin claim is true.

## Deviations from design

- `design/06-tech.md` anticipated deployment but not the CI substrate the corpus now assumes.
- Deployment is deferred until a server, websocket lifecycle, client reconnect path, and database
  migrations exist, so its acceptance tests can exercise real behavior.
- An earlier “not a free-tier workload” assumption was corrected by measurement and does not
  justify self-hosted hardware.

## Acceptance criteria

1. `make verify` passes locally; the six D2 jobs pass in CI from a clean checkout.
2. The RFC-0001 suite runs on Node and Chromium, Firefox, and WebKit in CI; a deliberately broken
   vector fails the relevant job.
3. The complete blocking workflow finishes in under five minutes on GitHub's public-repository
   hosted runners.
4. `make verify-schema` validates every schema and checked-in balance catalog; a deliberately
   malformed catalog makes that command fail.
5. Workflow permissions, cancellation, cache scope, trigger separation and observation failure
   behavior match D2/D3; seeded topology mutations are rejected.

## Open questions

- **Cross-architecture float determinism.** Resolve before the balance harness depends on
  byte-identical amd64/arm64 traces. A tolerance-based gate would be a real weakening and requires
  an explicit harness-RFC decision.
- **Deferred:** reconcile `design/12`'s production hot-reload promise with build-time gates before
  a production content mutation path is implemented.

## Changelog

- 2026-07-28: created (draft).
- 2026-07-28: recorded the public-repository decision; corrected the existing-Makefile premise;
  narrowed scope to executable CI gates; split geometric affordability into its own follow-up and
  deferred deploy/policy infrastructure until their runtime inputs exist.
- 2026-07-28: accepted by owner direction after verifying current supported action majors and the
  Playwright 1.62.0 container guidance against their official documentation.
- 2026-07-28: review recorded that fuzzing and vector-regeneration drift are not yet CI-enforced;
  assigned them to the nightly tier instead of leaving “mandatory” as a memory-based convention.
- 2026-07-28: added the scheduled/manual non-blocking numeric-maintenance job with a 30-second fuzz
  budget and deterministic vector-regeneration diff. Hosted execution still awaits the first push.
- 2026-08-06: non-normative reference cleanup for publication; no spec change.
- 2026-08-21: owner ruled D-014 for a strict sub-five-minute push/PR lane plus separate
  scheduled/manual exhaustive evidence. Reconciled the six-job current baseline, moved the full
  harness and numeric work into maintenance, and required a bounded fail-loud observation artifact.
