# Client Shell & Sim Loop — implementation plan

- **Assignee:** Codex
- **RFC:** `rfc/client-shell-and-sim-loop.md`
- **Started:** 2026-07-29

## Work breakdown

1. Add the pinned Svelte/Vite shell and strict client-shell policy catalog.
2. Implement the pure worker state machine and dedicated Worker entry point.
3. Implement authoritative adaptation, reconciliation, display counters, lifecycle, and intent boundary.
4. Build the DOM-first contract/main shell with reduced-motion and return-story states.
5. Add Node/browser/property/boundary tests, canonical docs, and build/CI gates.
6. Review the complete diff, run full verification, and archive.

## Acceptance gates

- Fixed-step/catch-up/offline handoff and every reconciliation row have deterministic fixtures.
- Frozen-while-producing and cap-reason properties pass at pathological magnitudes.
- No predicted-state action gate or balance-mutation import exists.
- Svelte production build and all three browser suites pass.
- Full `make verify` passes with Postgres.
