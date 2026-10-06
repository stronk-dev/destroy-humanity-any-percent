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
- [x] B3 (`fb0ab3b1`; TS activation witness in `cf57e956`) — Founder save v22 fields,
  boundary activation and codec/activation unit tests. **Not complete AC2/AC10:**
  RP-248's TS encoder correction needs designated review. RP-249's four required
  R7 shared corpus cases and baseline15 now execute locally with fired controls;
  their new Codex range also needs Claude. The original RT-DG-B unit substitution
  was not corpus completion. Codex's full original thirteen-path B3 review records
  CHANGES REQUIRED for those two gaps, not approval of its own later fixes.
  Seven earlier Go sources now pass both live-boundary and public Founder Exit
  activation with full encoded-state equality and four fired controls. New tests
  need Claude. A subsequent shared Go-authored corpus now supplies matching TS
  earlier-source replay, complete state/receipt/event/hash equality and three
  fired runtime controls; three corpus corruptions fail both lanes. This separate
  range also needs Claude. RP-250 subsequently closes a reproduced Go command
  admission gap locally (six corrupt-input cases and one output rollback fault,
  four positive controls, three fired guard removals), kernel0.3.156; Claude
  pending. RP-251's mirror retunes now execute and are locally corrected in Go
  live/Founder and TS Company/Founder (twelve cases, four fired removals),
  kernel0.3.157, designated review pending. Subsequent OD-7 admission rejects
  all 56 fixture removal cases; five independent guard omissions and the shared
  census corruption in both runtimes fail. Three legal starter boundary rows
  remain unchanged; retirement is separately refused. Kernel0.3.158 and the new
  complete span aftercf458b2b need Claude. Complete pinned R1/R7/AC11 remains;
  no removal/refund permission, real DB or full B4 verdict follows.
- [x] B4 (`541da96e`) — R5 `purchase_reputation_node` Founder intent (Go + TS replay parity, corpus). ACs 3, 4.
  **Not accepted closeout:** Codex's full original17-path designated review is
  CHANGES REQUIRED (RP-252 strict event admission, RP-253 persisted rejection
  population, RP-254 explicit result-hash evidence). The separate event repair
  underef37e83c rejects55malformed decisions, preserves six valid controls and
  eleven producer events, and fires three independent omissions; kernel0.3.159.
  Its new complete span after822774df needs Claude. Existing20purchase corpus
  rows/bundles are unchanged. Subsequent R8 tests undereddddc61 explicitly compare
  all20 purchase result pins and3 paired Founder Exit result pins in Go and TS.
  Four hash faults and four missing-observation controls fail; old direct
  comparisons alone survive their wrong-hash faults. Test-only, kernel159 and
  corpus unchanged; complete new span afterc355fb7d needs Claude. No fresh DB,
  full AC3/4/R8/B4 or approval of Codex's own fixes follows.
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
