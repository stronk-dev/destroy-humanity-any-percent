# Client Shell & Sim Loop — implementation plan

- **Assignee:** Codex
- **RFC:** `rfc/archive/client-shell-and-sim-loop.md`
- **Started:** 2026-07-29

## Work breakdown

1. [x] Add the pinned Svelte/Vite shell and strict client-shell policy catalog.
2. [x] Implement the pure worker state machine and dedicated Worker entry point.
3. [x] Implement authoritative adaptation, reconciliation, display counters, lifecycle, and intent boundary.
4. [x] Build the DOM-first contract/main/run-end shell with reduced-motion and return-story states.
5. [x] Add Node/browser/property/boundary tests, canonical docs, and build/CI gates.
6. [x] Record the mandatory independent diff review, then archive.

## Acceptance gates

- Fixed-step/catch-up/offline handoff and every reconciliation row have deterministic fixtures.
- Frozen-while-producing and cap-reason properties pass at pathological magnitudes.
- No predicted-state action gate or balance-mutation import exists.
- Svelte production build and all three browser suites pass.
- Full `make verify` passes with Postgres.
