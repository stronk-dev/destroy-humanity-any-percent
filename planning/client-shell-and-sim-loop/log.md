# Client Shell & Sim Loop — append-only log

## 2026-07-29 — acceptance and start

- Owner explicitly directed implementation of the RFC queue. Reviewed the RFC against the binding stack, browser-rendering research, and adopted `design/11 §1` return sequence.
- The draft's offline-return blocker was stale: design already resolves the sequence. Added C1-C6 for the abstract authoritative stream, worker messages/clock, strict policy data, reconciliation/display state, lifecycle behavior, and action/amplitude boundaries.
- Pinned current registry releases Svelte 5.56.8, Vite 8.1.5, and `@sveltejs/vite-plugin-svelte` 7.2.0. No push will be performed.

## 2026-07-29 — foundation and acceptance audit

- Landed the strict policy catalog, Svelte/Vite shell, abstract snapshot/receipt boundary, dedicated
  prediction Worker, reconciliation/display primitives, lifecycle hooks, intent dispatcher, and
  Node/browser/boundary gates in `5bcc00f`.
- Full-diff self-review found two correctness gaps before archive. First, prediction incremented
  each 50-ms tick independently, making rounding depend on pulse partitioning. Prediction now
  anchors to the last authoritative amount and calls the shipped `accrueConstant` closed form over
  total elapsed milliseconds; a 50+50 versus 100-ms regression fixture is exact. Second, the
  ten-minute return path depended on the Worker stalling, but browsers may keep a throttled Worker
  alive. Lifecycle now measures hidden duration directly, opens the return recap, and immediately
  requests authority on visibility return.
- Closed the remaining executable-contract surfaces: exact snapshot/receipt adaptation, closed
  six-intent dispatch union with runtime exact-field validation, aggregate-only reconciliation and
  Worker telemetry, server-coordinate progress rendering, generic discrete-state rendering,
  reduced-motion steps, contract/main/run-end routes, RTA start semantics, and the single
  fast-forward/gains/optional-modal return state.
- Focused verification: typecheck and Svelte diagnostics clean; 6,412 Node tests pass; production
  build emits a separate prediction Worker chunk; 19,245 tests pass across Chromium, Firefox, and
  WebKit; schema and source-boundary checks pass.
