# RFC: CI Baseline

- **Status:** draft
- **Author:** Marco (drafted by Claude, revised by Codex)
- **Created:** 2026-07-28
- **Design refs:** `design/07-roadmap.md` Phase 0, `design/12-content-pipeline.md §7`
- **Research:** `design/research/cicd-deploy.md` (primary), `design/research/pacing-science.md`, `design/research/adaptive-balancing.md`, `design/research/compliance-2026-refresh.md`
- **Depends on:** RFC-0001 and RFC-0002 (implemented)
- **Planning:** `planning/scaffolding-and-ci/` (once implementing)

## Summary

The repo has no `.github/` workflow, while its Makefile and cross-runtime test suites only run
when a developer remembers to invoke them. This RFC establishes the smallest blocking CI baseline:
the gates that exist today, on hosted GitHub Actions, under a strict latency budget.

## Motivation

The RFC-0001 Go, Node, and browser vector suites already exist but never run remotely. This is the
cheapest closure of a foundational obligation. The research also corrected an unmeasured premise:
the future balance harness fits public-repository hosted Actions, so there is no capacity argument
for self-hosted runners.

This scope is intentionally executable now. Deployment, policy gates without source artifacts,
and the balance harness are not smuggled into the starter.

**Out of scope, with owners:** deployment, Compose, Caddy, draining, migrations, and reconnect
testing (a future deployment-and-draining RFC once those components exist); the balance harness
(its own RFC); policy gates for assets and player-facing content (added when those inputs first
land); production hot-reload semantics (`design/12`, currently carrying a `DESIGN-GAP:`); the save
layer (its own RFC); leaderboard integrity and Balance Epoch enforcement (leaderboard RFC).

## Specification

### D1 — Runner topology and repository visibility

**NORMATIVE: the repository is public and all CI runs on hosted GitHub Actions.** No Komodo and
no self-hosted runners. CD topology is outside this RFC.

`cicd-deploy.md §2` verified that Komodo is a CD orchestrator, not a CI replacement, and would add
MongoDB plus a low-bus-factor dependency to a one-node deployment.

Marco selected a public repository on 2026-07-28. The research cost and timing model therefore
uses public-repository hosted runners: four vCPUs and no billed Actions-minute quota. Repository
visibility does not substitute for an explicit license.

### D2 — Starter workflow

One workflow, `.github/workflows/ci.yml`, on push and pull request:

| Job | Runs | Blocking | Notes |
|---|---|---|---|
| `server` | `make verify-server` → Go tests and `go vet` | yes | Uses the checked-in Go toolchain declaration |
| `client` | `make verify-client` → strict TypeScript and Node Vitest | yes | Installs from the frozen pnpm lockfile |
| `browser` | `make test-browser` in the Playwright image matching `client/package.json`, across Chromium, Firefox, WebKit | yes | Image supplies deterministic browser binaries |
| `schema` | `make verify-schema` → validate schema documents and every checked-in balance catalog | yes | An explicit empty catalog set succeeds; malformed checked-in catalogs fail |

The existing aggregate `make verify` remains the local all-gates command. It composes the narrower
targets above; CI calls narrow targets so browser tests are not executed twice.

The workflow has least-privilege read-only repository permissions, cancels superseded runs on the
same branch, and caches Go and pnpm dependencies. Generated build outputs and browser binaries are
not cached.

### D3 — Gate tiers and latency budget

| Tier | Trigger | Blocking | Contents |
|---|---|---|---|
| **1 — Baseline** | every push/PR | yes | D2's four jobs |
| **2 — Policy** | future RFC | yes | Copy, tracker, asset-provenance, and formula-drift gates once inputs exist |
| **3 — Balance** | future harness RFC | warn/block | Harness tiers 1–3 on balance changes |
| **4 — Nightly** | future harness RFC | no | Harness tier 4 and pacing artifact |

**NORMATIVE: the complete blocking workflow must finish in under five minutes.** If it exceeds the
budget, slower work moves to a non-blocking tier; the blocking budget does not grow.

### D4 — Deliberately separate performance repair

The geometric `MaxAffordable` repair changes implemented RFC-0002 behavior and belongs to the
follow-up RFC `geometric-afford-fast-path.md`. That RFC owns the exact-integer cap, postcondition
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
- `design/research/README.md`'s “not a free-tier workload” claim was corrected by measurement and
  does not justify self-hosted hardware.

## Acceptance criteria

1. `make verify` passes locally and in CI from a clean checkout.
2. The RFC-0001 suite runs on Node and Chromium, Firefox, and WebKit in CI; a deliberately broken
   vector fails the relevant job.
3. The complete blocking workflow finishes in under five minutes on GitHub's public-repository
   hosted runners.
4. `make verify-schema` validates every schema and checked-in balance catalog; a deliberately
   malformed catalog makes that command fail.
5. Workflow permissions, cancellation, cache scope, and trigger behavior match D2.

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
