# Server Garden implementation plan

RFC: `rfc/minigame-server-garden.md` (accepted 2026-09-25, every OD at its recommended default).
Fixture-first: no production mint (SG13). Numbering is landing-order: the RFC's "Founder v22"
means next-free, and at landing that is **Founder v25** (v22 Reputation, v23 Pet, v24 Cosmetics).
Replay inputs take the next free wire version, and the event migration takes the next free number.

- [x] G1 — `server/garden` + `client/src/garden`: SG1 loader (every rule, rejecting fixtures),
  SG2 state shape, SG3 advance, SG4 tick, SG5 pure commands, SG6 Founder-side harvest math and
  `harvest_hash`. Go-generated golden corpus replayed byte-for-byte by TS. AC1, AC3, AC4, AC5; pure
  half of AC7.
- [x] G2 — Pin `server_garden` in replay bundles (Go and TS), chain it on `cosmetics` (scalar Founder
  chain), and add the Fiscal cross-artifact rules (unlock row, host generator row). AC2 (bundle half).
- [x] G3 — Founder v25 `server_garden` save codec, activation at a new-run boundary and at
  New-Founder initialization, replay-inputs carry, Exit byte-identity. AC2, AC11.
- [ ] G4 — Founder intents `garden_plant`, `garden_uproot`, `garden_set_substrate`; the SG-P3
  advance pre-step on garden commands and `spend_fiscal_credit`; the frozen server-drawn salt;
  events and migration; TS replay; the AC6 trigger-set property test; Postgres witness. AC6, AC7,
  AC9 (Founder half), AC12.
- [ ] G5 — `garden_harvest` through `save.Store.ApplyGardenHarvestTransaction` (SG-P1/SG-P2):
  faucet window keyed `server_garden`, Company `credit_garden_harvest`, both logs binding
  `harvest_hash`, fault injection, idempotency. AC8, AC9.
- [ ] G6 — `GET /api/v1/garden/current` new operation (discarded-clone projection, no salt).
  AC10, AC15.
- [ ] G7 — Garden surface under `client/src/game-ui/` (grid, roving tabindex, non-colour stage,
  reduced motion, 320 px). AC13 stays blocked on the Accessibility RFC's acceptance; AC14 copy.
- [ ] Docs, then hand off for Codex's designated review. Never self-archive.
