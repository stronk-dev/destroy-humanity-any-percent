# Client Shell & Sim Loop — append-only log

## 2026-07-29 — acceptance and start

- Owner explicitly directed implementation of the RFC queue. Reviewed the RFC against the binding stack, browser-rendering research, and adopted `design/11 §1` return sequence.
- The draft's offline-return blocker was stale: design already resolves the sequence. Added C1-C6 for the abstract authoritative stream, worker messages/clock, strict policy data, reconciliation/display state, lifecycle behavior, and action/amplitude boundaries.
- Pinned current registry releases Svelte 5.56.8, Vite 8.1.5, and `@sveltejs/vite-plugin-svelte` 7.2.0. No push will be performed.
