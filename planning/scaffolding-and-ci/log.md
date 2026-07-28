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
