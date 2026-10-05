# Server Garden implementation plan

RFC: `rfc/minigame-server-garden.md` (accepted 2026-09-25, every OD at its recommended default).
Fixture-first: no production mint (SG13). Numbering is landing-order: the RFC's "Founder v22"
means next-free, and at landing that is **Founder v25** (v22 Reputation, v23 Pet, v24 Cosmetics).
Replay inputs take the next free wire version, and the event migration takes the next free number.

**Review checkpoint, 2026-10-06:** the checked G1–G7 rows below record implementation presence,
not designated acceptance. Codex's bounded SG1 review found RP-212–RP-214 in the original loader;
the separately predeclared catalog correction and 36-case raw-byte population are locally green
and await Claude's designated review. Kernel 0.3.149 covers this actual admission change, not a
repair of historical RP-131. No Garden archive, SG13 mint or whole-G1/Garden approval is claimed.
Next review lane: the remaining G1 state/clock/engine/commands/corpus, then G2–G7 under bounded
predeclarations. Retain SG2 nullable fields and all original engine/replay/activation evidence.

The separate SG2 codec review under `e8d3f2df` confirms RP-215 (null integer fields default to
zero). Its local correction uses scoped nullable-key lists and kernel 0.3.150, with original
save/replay versions unchanged. Actual codec provenance is G3 `415bea4d`, not G1: the shared
state shape was G1, but its Go strict decoder landed in G3. This is a bounded codec finding,
not completed G2/G3 or whole-G1 review. Eight valid / 21 typed refusal cases run in both
runtimes; two additional raw duplicate refusals are Go-only with visible TS parsing-loss
controls. Both positive/negative probes discriminate and restore. Claude review is required.

Separate test-only SG6 supplement under `3fbf39af`: real Go/TS harvests match 18 independently
derived boundary result/hash/post-state bytes, including the large-product floor and 36-plot
sum. Four restored TS mutations prove arithmetic, seed, hash and post-state checks can fail.
No production/balance/kernel change; kernel stays 0.3.150. This is not AC8's coordinator or
whole-G1/Garden acceptance, and its test range awaits Claude. Next: SG3/SG4 clock/tick review.

Separate RP-217 correction under `14ad4287`: actual Postgres diagnosis shows handler-time
harvest can reject on lag or persist future growth and block later Fiscal commands. SG-P2's
Garden-specific request now takes no timestamp; the DB sample and applicability probe run
under the canonical stream locks. Six skew arms plus retries/hash conflicts/history, two
restored probes and twenty cold repetitions pass. Kernel 0.3.151; cross-party review required,
not full-G5/SG3/SG4 or Garden acceptance. Continue remaining pure clock/tick and G2–G7 review.

- [x] G1 — `server/garden` + `client/src/garden`: SG1 loader (every rule, rejecting fixtures),
  SG2 state shape, SG3 advance, SG4 tick, SG5 pure commands, SG6 Founder-side harvest math and
  `harvest_hash`. Go-generated golden corpus replayed byte-for-byte by TS. AC1, AC3, AC4, AC5; pure
  half of AC7.
- [x] G2 — Pin `server_garden` in replay bundles (Go and TS), chain it on `cosmetics` (scalar Founder
  chain), and add the Fiscal cross-artifact rules (unlock row, host generator row). AC2 (bundle half).
- [x] G3 — Founder v25 `server_garden` save codec, activation at a new-run boundary and at
  New-Founder initialization, replay-inputs carry, Exit byte-identity. AC2, AC11.
- [x] G4 — Founder intents `garden_plant`, `garden_uproot`, `garden_set_substrate`; the SG-P3
  advance pre-step on garden commands and `spend_fiscal_credit`; the frozen server-drawn salt;
  events and migration; TS replay; the AC6 trigger-set property test; Postgres witness. AC6, AC7,
  AC9 (Founder half), AC12.
- [x] G5 — `garden_harvest` through `save.Store.ApplyGardenHarvestTransaction` (SG-P1/SG-P2):
  faucet window keyed `server_garden`, Company `credit_garden_harvest`, both logs binding
  `harvest_hash`, fault injection, idempotency. AC8, AC9.
- [x] G6 — `GET /api/v1/garden/current` new operation (discarded-clone projection, no salt).
  AC10, AC15.
- [x] G7 — Garden surface under `client/src/game-ui/` (grid, roving tabindex, non-colour stage,
  reduced motion, 320 px). AC13 stays blocked on the Accessibility RFC's acceptance; AC14 copy.
- [ ] Docs, then hand off for Codex's designated review. Never self-archive.
