# Reputation Tree v1 implementation plan

RFC: `rfc/reputation-tree-v1.md` (accepted 2026-09-25; every owner decision at its recommended
default). Fixture-first: no production epoch is minted by this plan (R11 is owner-gated; OD-2's
threshold retune is measured and reported, then ratified by owner SHA).

## Batches

- [x] B1 (`a522fdf1`) — R2 artifact + loaders (Go `server/reputation`, TS `client/src/reputation.ts`), R1
  accounting helpers, R3 bonus arithmetic; shared rejection-fixture corpus and bonus vectors.
  ACs 1 (loader half), 5.
- [x] B2 (`c62afe47`) — Bundle wiring: `reputation_tree` optional artifact in the Go/TS bundle loaders, economy
  declaration pairing rule (tree ⇔ `reputation.founder_bonus` row), frozen-contribution resolver
  accepts provider `reputation_tree`. AC1 (bundle half).
- [x] B3 (`fb0ab3b1`; TS activation witness in `cf57e956`) — Founder save v22 (R1/R7): fields, codec invariants, v21→v22 migration, migration corpus
  cases + baseline ratchet. ACs 2, 10.
- [x] B4 (`541da96e`) — R5 `purchase_reputation_node` Founder intent (Go + TS replay parity, corpus). ACs 3, 4.
- [x] B5 (`7ea6b942`, `078b205e`, `7d130b89`) — R3/R4 new-run assembly: frozen `reputation.founder_bonus` row (all run-creation paths),
  starter application, DB migrations (completeness function, event kinds, founder_log arms),
  `run_started` v2. ACs 6, 7, 8, 11.
- [x] B6 (`ac8a9a65`, `cf57e956`) — R6 Exit-attached plan (`exit.v2`, replay-inputs bump). AC9.
- [x] B7 (`1e9b3a9b`, `1edc1c37`; composed purchase-through-UI awaits a tree-pinning epoch) — R9 UI: snapshot reputation block, `reputation_tree` surface, plan panel. AC12.
- [x] B8 (`b522c577`, `293c3ac7`, `6d821bab`; H4 verdict FAIL recorded as owner input, not loosened) — R10 harness H1–H5 and the OD-2 threshold measurement report. AC13.
- [ ] B9 — BLOCKED on the owner-gated mint (no epoch pins a tree; formulas-check clean) — formulas regeneration (separate commit), composed verification career. ACs 14, 15.
- [ ] Canonical docs, designated Codex review, archival (Codex).

Kernel protocol: every commit touching a `kernel/affecting-paths.json` prefix bumps
`kernel/VERSION` (+ Go/TS constants) in the same commit; `server/reputation/` and
`client/src/reputation.ts` are registered in the commit that creates them. Save/snapshot versions
are assigned in landing order (the RFC's "v22" means next-free).
