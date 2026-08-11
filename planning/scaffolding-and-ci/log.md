# CI Baseline — Running Log

## 2026-07-28 — Start

- Owner selected a public repository, resolving the draft's only blocking owner decision.
- Narrowed the inherited draft to executable CI behavior; deployment, reconnect, assets, copy,
  and formula policy gates have named future owners.
- Verified current primary documentation before acceptance: checkout/setup-go/setup-node and
  pnpm setup use supported v6 majors; Playwright requires its Docker image version to match the
  installed package, which is 1.62.0 here.
- Selected Node 24 and the Noble Playwright 1.62.0 image. Browser binaries come from the image and
  are deliberately not cached.
- Selected pinned Ajv 8.20.0 for Draft 2020-12 schema and fixture validation so local and CI use
  one repository-owned command rather than a globally installed utility.

## 2026-07-28 — Local implementation

- Split the Makefile into `verify-server`, `verify-client`, `verify-schema`, `test-browser`, and
  aggregate `verify` targets without executing the browser suite twice.
- Added Ajv 8.20.0, a repository-owned schema verifier, and one positive plus one negative catalog
  fixture. With a deliberately malformed file placed in `balance/catalogs/`, `make verify-schema`
  failed on the unknown field as required; removing it restored the green gate.
- Added the four-job GitHub workflow with read-only permissions, push/PR triggers, branch-scoped
  cancellation, frozen pnpm installs, Go-module and pnpm dependency caches, and no browser-binary
  cache.
- `pnpm install --frozen-lockfile`, every narrow non-browser target, and aggregate `make verify`
  pass locally. The aggregate run completed in 5.86 seconds and included 6,321 Node tests plus
  18,963 browser tests across Chromium, Firefox, and WebKit.
- The workflow file cannot be executed on a hosted runner without pushing. Owner previously said
  not to push, so the RFC remains `implementing`; hosted completion under five minutes is the only
  unverified acceptance gate.

## 2026-08-10 — designated cross-party verdict: CI-repair batch {4a45b8d, 0db9768} — BOTH APPROVED

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.
**The Decimal fix verified genuine at bit level:** the shipped Int64Exact classified via a lossy
float64 reconstruction whose snap-tolerance comparison the Go arm64 backend FMA-fuses (proven:
the darwin diff 7.46e-11 is not a ULP multiple — only possible with an unrounded intermediate),
while linux/amd64 rounds separately (exactly 1 ULP ≥ the 1e-10 snap) and REJECTED canonical
integers. Reproduced in both directions across real architectures. **Blast radius bounded:** the
sole consumer fails CLOSED (ErrInvalidEngineState — availability fault, never corruption); stored
mantissa/exponent bits identical on both platforms; no persisted state/receipt/replay/wire could
have diverged; unpushed repo, zero production exposure. The fix (round-trip through the
normalized representation) is architecture-independent, only tightens acceptance, and the 0.3.89
bump is honest and REQUIRED. No shared-vector op exists for this classification (no TS twin —
correctly); the Go-side pin + the linux/amd64 CI gate discriminate. Fixture repairs test-tier
(both genuinely stale, both reproduced); `make test-go-ci` faithfully mirrors the Actions job and
ran GREEN under emulation; the plan record faithful with no box flips.
**Routed follow-ups (non-blocking):** R-F1 (MEDIUM) — the root FMA sensitivity remains in
toFloat64's snap comparison, and Floor IS a shared-vector op: force materialization (explicit
float64 conversion) in a future kernel bump + a razor-edge floor golden vector. R-F2 (LOW) — the
regression test discriminates only on linux/amd64 (by nature; the CI gate carries it). R-F3
(LOW) — docs/ci.md owes the test-go-ci line. R-F4 — a Claude-side numbering slip, fixed same day.

## 2026-08-11 — Actions run 31486886470 browser failure reproduced and repaired

- **Implemented by:** Codex. **Recorded by:** Codex. This is an implementation record pending the
  ordinary designated review; it is not an approval.
- The public run metadata isolated the failure to the Playwright `browser` job at `9a97543`.
  Client and schema were green; server was cancelled by workflow fail-fast rather than failing.
  The local GitHub CLI credential was invalid, so the public Actions API supplied job metadata and
  the failure was reproduced in the exact `mcr.microsoft.com/playwright:v1.62.0-noble` image.
- Linux Firefox exposed a test-clock race: the real Worker/render integration was required to
  change the visible amount within a fixed 750 ms while all three browser projects ran
  concurrently. The Worker remained correct, but Firefox was scheduled after that arbitrary
  deadline and the assertion observed `100 == 100`.
- Commit `61da160` retains the same discriminating observable contract but waits up to five seconds
  for the amount to change, polling every 50 ms. It does not stub the Worker, skip Firefox, or
  weaken the expected output. The exact CI image then passed all 120 browser files and 19,972 tests
  with two declared skips; the focused ordinary Make target passed all three engines locally.

## 2026-08-11 — Actions run 31486886470 server timeout repaired

- **Implemented by:** Codex. **Recorded by:** Codex. Pending the ordinary designated review.
- The server job began at 11:30:16Z and GitHub cancelled it at 11:35:32Z while `make
  verify-server` was still running: the five-minute job budget, not a test assertion, terminated
  the gate. A cold linux/amd64 reproduction took 4m30s merely to finish the Go/Postgres suite and
  generation checks before entering the composed balance harness, proving the old timeout
  structurally insufficient.
- The server job now has a ten-minute ceiling without removing, splitting, caching away, or
  skipping any verification. `make verify-server-ci` reproduces the entire gate on linux/amd64
  with real Postgres and completed green. `make test-browser-ci` similarly owns the exact pinned
  Playwright-image reproduction and completed all 19,972 browser tests green. Both normal targets
  are documented in `docs/ci.md`; the existing host targets remain unchanged.
