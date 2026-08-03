# Purchasable Content Foundation — implementation plan

- **RFC:** `rfc/purchasable-content-foundation.md`
- **Assignee:** Codex
- **Status:** implementing

1. [ ] Land economy schema v4: typed upgrades, roles, provision edges/caps, ladders, synergy
   pools, strict cross-reference validation, and Go loader tests.
2. [ ] Land save v14: owned upgrades, provisioned counts, per-edge provision remainders, stock-rate
   remainder, migration corpus, canonical wire closure, and reset/import/genesis coverage.
3. [ ] Implement state-derived contributions, upgrade purchase, manual targets, stock-rate roles,
   fixed-grid provision accrual, events/receipts, and Go replay parity.
4. [ ] Port catalog/save/transition behavior to TypeScript and extend the shared sequential replay
   corpus with byte-identical state, receipt, and event expectations.
5. [ ] Add the simulation-only ablation entrypoint and import/source guard; integrate the harness
   without adding mask state to authoritative bundles or replay inputs.
6. [ ] Close transport/client/schema/event registries, formula generation, KV-1, epoch mint,
   canonical docs, and full repository verification.
7. [ ] Obtain an independent full-span review, remediate all findings, then archive the RFC and
   planning record with exact range-union provenance.

Acceptance checkboxes flip only in commits containing the corresponding proof, per `AGENTS.md`.
