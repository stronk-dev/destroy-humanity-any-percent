# RFC: Scaffolding & CI

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-07-28
- **Design refs:** `design/06-tech.md §7`, `design/07-roadmap.md` Phase 0, `design/12-content-pipeline.md`
- **Research:** `design/research/cicd-deploy.md` (primary), `design/research/pacing-science.md`, `design/research/adaptive-balancing.md`, `design/research/compliance-2026-refresh.md`
- **Depends on:** RFC-0001 (implemented), RFC-0002 Economy Kernel (implemented)
- **Planning:** `planning/scaffolding-and-ci/` (on implementing)

## Summary

The repo has **no `.github/`, no compose file, and no Makefile**, while six design and research documents have independently accumulated **seven distinct "enforced in CI" obligations** that nothing owns. This RFC establishes the CI/CD substrate and wires the gates that are *already built and merely unconnected*.

It is deliberately small. `cicd-deploy.md §11` measured the starter at **~2 hours**, and the largest risk to a gate stack is not that it is incomplete — it is that it is slow enough to be bypassed.

## Motivation

Two findings drive the scope.

**The cross-runtime parity suite already exists and does not run.** RFC-0001 shipped Go and TS vector suites plus a browser-matrix requirement. Nothing executes them on push. This is the cheapest possible closure of a stated obligation.

**A false premise was removed.** `design/research/README.md` previously asserted that the nightly N=1000 balance job "is not a free-tier GitHub Actions workload," which made runner choice look like a hard constraint. `cicd-deploy.md §3.1` benchmarked it against the shipped kernel: **N=1000 bots × 30 virtual days = 6.96 s on 14 cores**, with a pessimistic full-game/T5-horizon/4-core extrapolation of **~73 min** — inside the 6 h job limit. **There is no capacity argument for self-hosting.**

**Out of scope, with owners:** the balance harness itself (its own RFC — and see D4 below, it is *blocked*); production hot-reload semantics (`design/12`, currently carrying a `DESIGN-GAP:`); the save layer (its own RFC); leaderboard integrity and Balance Epoch enforcement (deferred to the leaderboard RFC, though D5 reserves the seam).

## Specification

### D1 — Runner topology: hosted Actions for CI, webhook for CD

**NORMATIVE: all CI runs on hosted GitHub Actions. CD is a cattery-style webhook script. No Komodo, no self-hosted runners, no preview environments.**

`cicd-deploy.md §2` verified against Komodo's docs *and its Rust source* that **Komodo is CD, not CI**: its `Build` runs `docker build` only — no test stages, no matrix, no fail→block, no rollback. Adopting it would add a MongoDB and a bus-factor-1 GPL dependency to replace a ~20-line script we already own in cattery.

⚠️ **OWNER DECISION, blocking: repository visibility.** Public repos get 4 vCPU and unlimited free Actions minutes; private repos get 2 vCPU and a monthly quota. **This is the input to every cost number in `cicd-deploy.md`.** It also interacts with the "published community formulas" design law (#9) and with `compliance.md`'s licensing section. Decide explicitly; do not let it default.

### D2 — The starter pipeline (this is the whole first deliverable)

One workflow, `ci.yml`, on push and PR:

| Job | Runs | Blocking | Notes |
|---|---|---|---|
| `verify` | `make verify` → `go test ./...` + `vitest run` + `go vet` + lint | **yes** | The Makefile is part of this RFC |
| `browser` | The RFC-0001 TS vector suite under a **Playwright container** across Chromium, Firefox, WebKit | **yes** | Container, not `setup-*` actions — deterministic browser versions |
| `schema` | `check-jsonschema` over `balance/` | **yes** | ~3 lines; catches malformed catalogs before they reach a loader |

**Acceptance:** `verify` + `browser` together close the RFC-0001 cross-runtime obligation, which is currently the single largest unwired gate in the repo.

**NORMATIVE budget: the blocking set must complete in under 5 minutes.** `cicd-deploy.md §9` is explicit that this is the line between a gate that is kept and a gate that is worked around. If the set exceeds the budget, work moves to a non-blocking tier — it does not get a longer budget.

### D3 — Gate tiers

| Tier | Trigger | Blocking | Contents |
|---|---|---|---|
| **1 — Fast** | every push/PR | yes | D2's three jobs |
| **2 — Policy** | every PR | yes | Vale copy lint; no-tracker/no-consent-banner check; asset-manifest attestation (D6) |
| **3 — Balance** | PR touching `balance/**` | warn/block | Harness tiers 1–3 — **deferred to the harness RFC**; this RFC only reserves the trigger path |
| **4 — Nightly** | schedule | no | Harness tier 4; publishes the pacing-curve artifact |

### D4 — Blocking defect: `MaxAffordable`

**`server/economy/curves.go:59` binary-searches to `MaxExactInteger`** (`high := decimal.MaxExactInteger - owned`) instead of delegating the `geometric` case to **`decimal.AffordGeometricSeries`** — the closed form RFC-0001 already shipped and tested, which currently has **no non-test caller**.

Measured (`cicd-deploy.md §9.3`): **20,486 ns/op vs 660 ns/op, ~95×.** On a 200-bot harness run that is **3 min 01 s vs 1.91 s** — straddling D2's budget by itself.

**NORMATIVE:** the generic binary search remains correct and stays as the fallback for `constant` and `linear`. **`geometric` must take the closed form.** A benchmark guard asserts the geometric path stays within one order of magnitude of `AffordGeometricSeries`.

**This blocks the balance-harness RFC**, and is small enough to land independently of the rest of this document.

### D5 — Deploy: drain, do not blue/green

**NORMATIVE: single-container rolling replacement with an explicit drain. No blue/green.**

`cicd-deploy.md §5` establishes that Go's `Shutdown` explicitly does not wait for hijacked/websocket connections — so a second container does not solve what it appears to solve. And **design law 7 already makes a brief disconnect a first-class game state**: the client is server-authoritative and must reconcile on reconnect regardless.

Drain sequence: (1) stop accepting new websocket upgrades; (2) broadcast a `server_restarting` frame so the client can show it diegetically; (3) flush pending state commits; (4) `Shutdown` with a bounded timeout, then replace.

**Client reconnect is mandatory and is a dependency, not an optimisation.** Migrations run before the new binary accepts traffic; rollback is redeploying the previous image tag.

### D6 — Policy gates, and what they cannot prove

Three obligations are *design laws expressed as CI*. **State their limits honestly rather than implying more assurance than exists.**

| Law | Gate | What it actually proves |
|---|---|---|
| No genAI assets | Every file in `assets/` has a manifest row with a human attestation and a source | **It cannot prove an asset was not AI-generated.** It proves someone signed a claim, and makes an unattested asset a build failure. That is the whole of the guarantee — say so in the manifest header |
| No trackers / no consent banner | Banned-pattern lint over client source and built output | Strong: this is a real property of the bundle. Note `compliance-2026-refresh.md` found web push does **not** violate this rule — a browser permission prompt is not a consent banner |
| Published formulas | Community-facing formula docs regenerate from source; CI fails on drift | Strong, and it makes law #9 mechanical rather than aspirational |

### D7 — Repo layout this RFC creates

```
.github/workflows/ci.yml
Makefile                  # verify, test, vectors, lint, run
docker-compose.yml        # postgres + server + caddy
Caddyfile
scripts/deploy.sh         # webhook target, ported from cattery
.vale.ini + styles/
```

## Deviations from design

- **`design/06 §7`** anticipated a deploy story only; this RFC adds the CI substrate the rest of the corpus turned out to assume.
- **`design/research/README.md`'s** "not a free-tier workload" claim was wrong by ~2 orders of magnitude and has been corrected in place. Recorded here because it is the reason this RFC does *not* provision hardware.

## Acceptance criteria

1. `make verify` passes locally and in CI from a clean checkout.
2. The RFC-0001 vector suite runs on Node **and** Chromium, Firefox, WebKit in CI; a deliberately broken vector fails the build.
3. Tier 1 + Tier 2 complete in **under 5 minutes** on the chosen runner class.
4. `check-jsonschema` rejects a deliberately malformed balance catalog.
5. `MaxAffordable`'s geometric path delegates to the closed form; the benchmark guard passes; **existing max-affordable postconditions from RFC-0001 still hold** (this is a performance change, not a semantic one).
6. An unattested file in `assets/` fails the build.
7. A deploy to staging drains websockets, runs migrations, and serves the new binary; the client reconnects without losing committed state.

## Open questions

- ⚠️ **Repository visibility (D1).** Blocks `accepted`.
- **Cross-architecture float determinism.** `cicd-deploy.md` flags this as its riskiest item: the byte-identical golden-seed gate runs on amd64 CI and arm64 locally. **Resolve before the harness RFC depends on byte-identity** — if it does not hold, the gate becomes tolerance-based, which is a real weakening and should be an explicit decision rather than a discovery.
- **Deferred:** hot-reload's gate bypass (`design/12` `DESIGN-GAP:`) — the scaffolding must not be read as blessing an ungated production mutation path.

## Changelog

- 2026-07-28: created (draft).
