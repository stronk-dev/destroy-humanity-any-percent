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
  RP-255's subsequent frozen-input supplement under761462fc/4ca6dc7d refuses
  all570raw malformed fields after376baseline admissions, preserving19valid
  controls. TS already refuses456objects;114parsed duplicates are honest
  positive controls, not wire refusals. Two runtime omissions and six fixture
  corruptions fire, restoring exact SHAs; kernel0.3.160. Cold core/vet and
  client/type/build plus separate boundaries pass8047/134. Full client remains
  RP-131 RED; native Chromium/WebKit590each pass but Firefox executes no tests,
  including a narrowly escalated run (RP-256). Complete span after4d690f29
  needs Claude. RP-253 actual persisted taxonomy, history/verifier and remaining
  pinned readers still required; no new closeout or checkbox follows.
  RP-253's test-only preparation under5425e684/373a6e1a/2e9cc4c4 now retains
  all20source cases plus overdue-Fiscal rejection. All21local profiles pass;
  three source/population corruptions reject and fire on validator omission.
  The SQL lane prepares42unrecorded conflicts, exact retry and complete stored
  state/log/event/outbox assertions, but explicitly skips without Postgres.
  Cold core/vet pass; actual DB and four persistence controls remain mandatory.
  The invalid v21/tree fixture is a separate pinned-state refusal, not an
  invented valid persisted profile. Complete new span after86c96035 needs
  Claude; no AC3 closeout or reversal of the original CHANGES REQUIRED verdict.
  RP-259 under e3378e96 adds portable public history verification for 24
  source-derived profiles, including the complete nine-node chain and three
  paired Exits. All 144 corruptions and four population controls reject.
  Six compiling runtime omissions fire; two single sequence omissions survive
  through the other defense. Under c7c7a0fb, the missing-paired-Exit control
  fires on paired-count omission; whole-validator omission fires all four.
  Historical three-control runs stay recorded. Exact sources restore; cold
  core/vet pass. Test-only, kernel160
  unchanged; new span after b6c7ebca needs Claude. Actual SQL/two-Exit career
  and whole R8 remain; the separate Company-run supplement is RP-260 below.
  RP-260 under4d4846f2/f7822250/d23da963 retains two full Company runs from
  the exact source starter genesis, with frozen1.003 and synthetic unit control.
  Public verdict/full states pass; 18 corruptions, four population controls and
  two false heads reject. Six compiling omissions fire; all exact sources
  restore, cold core/vet pass. No runtime/kernel/old-corpus/schema/CI change.
  New span after721c0ee1 needs Claude. Actual SQL/two-Exit/AC15 and remaining
  harness consumers still mandatory; no box flipped here.
- [x] B5 (`7ea6b942`, `078b205e`, `7d130b89`) — R3/R4 new-run assembly: frozen `reputation.founder_bonus` row (all run-creation paths),
  starter application, DB migrations (completeness function, event kinds, founder_log arms),
  `run_started` v2. ACs 6, 7, 8, 11.
- [x] B6 (`ac8a9a65`, `cf57e956`) — R6 Exit-attached plan (`exit.v2`, replay-inputs bump). AC9.
- [x] B7 (`1e9b3a9b`, `1edc1c37`; composed purchase-through-UI awaits a tree-pinning epoch) — R9 UI: snapshot reputation block, `reputation_tree` surface, plan panel. AC12.
  RP-257/RP-258 reader corrections under336503b3 now refuse14corrupt Go
  projections and48invalid TS factors, preserving valid controls. Three compiling
  guard omissions fire2/12/48 failures and restore exact sources. Cold core/vet,
  client8106/134, types/build and separate boundaries pass; composite stays
  RP-131 RED. Presentation-only, kernel160/corpus/schema unchanged. Full new
  span after4030c895 needs Claude. Original B7, author body reconciliation,
  minted purchase-through-UI and whole AC12 remain; no box flipped here.
- [x] B8 (`b522c577`, `293c3ac7`, `6d821bab`; H4 verdict FAIL recorded as owner input, not loosened) — R10 harness H1–H5 and the OD-2 threshold measurement report. AC13.
  RP-261's bounded input correction under9337f4b8 now refuses six malformed
  career configurations after four baseline admissions. Four legal careers
  pass; two guard omissions discriminate and restore exactly. Cold fast/core/
  vet/topology pass; exhaustive H4/H5 reports reproduce unchanged. This does
  not close AC13: H4 still fails, H5's epsilon/run-4 questions remain, and
  RP-263's helper/run-key portion is now locally corrected under26e97e4b:
  seven controls and three compiling omission probes discriminate; cold
  fast/core/vet/topology pass and full H4/H5 reproduce unchanged in840.123s.
  Report-envelope provenance remains open. New span after6963692b needs Claude.
  RP-265/266's subsequent frozen-input/boundary correction underaa4c241b passes
  all16Reference observations,20producer profiles and six boundary controls;
  nine independent compiling omissions discriminate with exact restoration.
  Cold core/vet, fast harness and client8106/134/types/build plus separate
  boundaries pass; composite stays RP-131 RED. Kernel0.3.161, no balance/live
  formula change. Seed0 treated gate moves355000→357000ms; control stays357000ms.
  Full H4/H5 now reject untouched retained report drift in786.459s; the same
  six Casual ties remain. Complete new span afterd18d4e09 needs Claude.
  Next predeclare/measure RP-267's Reference lifetime-hook lead before RP-262 H3,
  then RP-264's separate exclusion oracle. Report provenance/
  refresh remains separate; no retuning or report-regeneration authority here.
  Complete new span after7da200f0 needs Claude. No checkbox flipped here.
- [ ] B9 — BLOCKED on the owner-gated mint (no epoch pins a tree; formulas-check clean) — formulas regeneration (separate commit), composed verification career. ACs 14, 15.
- [ ] Canonical docs, designated Codex review, archival (Codex).

Kernel protocol: every commit touching a `kernel/affecting-paths.json` prefix bumps
`kernel/VERSION` (+ Go/TS constants) in the same commit; `server/reputation/` and
`client/src/reputation.ts` are registered in the commit that creates them. Save/snapshot versions
are assigned in landing order (the RFC's "v22" means next-free).
