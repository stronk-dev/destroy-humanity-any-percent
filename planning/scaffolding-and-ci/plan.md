# CI Baseline — Implementation Plan

- **RFC:** `rfc/scaffolding-and-ci.md`
- **Assignee:** Codex
- **Started:** 2026-07-28

## Work

1. Split the Makefile into server, client, browser, schema, and aggregate verification targets.
2. Add a pinned Ajv 2020-12 schema verifier plus positive and negative catalog fixtures.
3. Add the least-privilege four-job GitHub Actions workflow with frozen installs, dependency
   caches, matching Playwright container, and superseded-run cancellation.
4. Exercise every narrow Make target and the aggregate local gate.
5. Update README and canonical CI documentation.
6. Archive the completed RFC and planning record.

## Acceptance gates

- Clean, frozen dependency installation succeeds.
- Server, client, browser, and schema targets pass independently.
- A negative schema fixture is rejected while the schema command succeeds by observing that
  expected rejection.
- `make verify` passes, including all three browser engines.
- Workflow structure has read-only permissions, push/PR triggers, branch-scoped cancellation,
  supported action majors, exact Playwright image/package parity, and no browser-binary cache.
- Blocking elapsed time is measured on GitHub after the first push; the RFC cannot claim the
  hosted-runner five-minute gate from a local machine.
