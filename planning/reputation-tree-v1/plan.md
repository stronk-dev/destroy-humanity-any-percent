# Reputation Tree v1 implementation plan

RFC: `rfc/reputation-tree-v1.md` (accepted 2026-09-25; every owner decision at its recommended
default). Fixture-first: no production epoch is minted by this plan (R11 is owner-gated; OD-2's
threshold retune is measured and reported, then ratified by owner SHA).

Current tooling checkpoint: RP-281 paired template/generated Go repair under
da7cfa56 passes five generation goldens/collision/six corruptions, fired template
severings and actual drift; output equals gofmt(old Go). Other artifacts/copy658/
hash/manifest unchanged. Cold client/vet/core package union pass with initial
sandbox socket failure and narrow rerun disclosed. New Go producer tree
1c54c2c3d0139b50e97937bfe88c65f339cab6b5; dated career reports not restamped.
Full span after3df3ff32 needs Claude; no acceptance checkbox changes. Next ground
RP-284 consumers, RP-283 author contract gap; all full1.0 holds remain.

## RP-284 bounded event-reader supplement — 2026-10-06

Implement only accepted R7's existing `run_started` v1/v2 payload reader and
its runtime delivery, including the HTTP-snapshot-first race. No new wire,
RunEndSurface prop, snapshot-derived terminal rendering, automatic navigation,
copy, payout calculation, balance mutation, cursor policy widening for other
events, or acceptance checkbox. The older Game UI GU-C3 payload-only boundary
and Reputation R9's post-Exit balance requirement need author reconciliation
before the display bridge changes. This supplement is not the RP-284 UI fix.

Predeclared population: absent/null/object tree arms; canonical >=1 factor;
unique starter IDs in producer array order (not lexical order); exact base,
assisted, run-ID and tree fields. Refuse malformed fields/types/coordinates,
noncanonical/below-one/out-of-state-range factors and duplicate/invalid IDs.
Runtime controls cover event-first, snapshot-first, wrong Founder, old/future
run, duplicate channel offset, and malformed summary taking existing resync.
Demonstrate baseline failures, then compiling reader omission, summary
validation omission, and snapshot-first delivery omission. Restore exact bytes
after each terminal result. Root client types/build/unit/boundary/copy gates;
record inherited history/Firefox/SQL/mint holds separately. Full range after
`0ae1fa11` requires Claude; no archival or whole AC12 claim.

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
  RP-285/RP-286 under2b2536a9 adds explicit native header/button Tab stops.
  Original44 failures/10 controls pass; three compiling probes fail36/52/12;
  final248 selected Chromium/WebKit pass after exact restoration. Types/build/
  units8106/boundaries/copy/manifest pass;255 Node skips/RP-131 RED/Firefox
  unexecuted remain. Controlled native component evidence, not manual AT/
  SQL/mint/full AC12. Whole span after8a3bfb70 needs Claude. Next RP-281
  tooling or RP-284 producer ground; RP-283 author gap remains. No checkbox
  flip, inherited B7 closure or new purchase/browser policy.
  RP-282 under673c09a6/916abc50/c12e985f now synchronizes existing fresh
  panel selection with host, preventing invisible old-plan submission. Refined
  baseline16 fail/24 controls green; two compiling mutations fail16/24 with
  exact restoration; final194 native Chromium/WebKit pass. Initial incomplete
  Offer fixture is disclosed, not product revision evidence. Types/build8106
  units/boundaries/copy/manifest pass;228 Node skips/RP-131 RED/Firefox remain.
  Actual host/runtime over controlled HTTP/socket, not server/SQL/mint/full AC12.
  Whole span aftereb258a7a needs Claude. Next RP-285 Tab header/row diagnostic;
  RP-283 authoritative preview DESIGN-GAP/RP-284 Run End consumer stay separate.
  No checkbox flipped, full B7 closure or new persistence semantics.
  RP-279 underd12fdcff/20c44899/285a247e now persists actual Amount cost
  outside all row-control conditions. Baseline40 fail; seven compiling mutations
  discriminate and restore exact source; final154 native Chromium/WebKit pass.
  Four states/both eras/notation/native confirmation and held task/replacement
  execute. Types/build8106 units/boundaries/copy/manifest pass;208 Node skips,
  RP-131 RED/Firefox unexecuted remain. New diagnostic is component-only, not
  host/SQL/mint/full AC12. Whole span after22e03946 needs Claude. Next ground
  R9 plan-panel/host lifecycle RP-282; RP-283 preview and RP-284 Run End producer
  gaps remain separately routed. No checkbox flipped or inherited B7 closure.
  RP-278 host correction underaf744e65/defa9679/441013a5 first fails48
  rejection cases while eight applied controls pass. Restored host+child114
  and shared GS0.2 two cases pass; seven compiling omissions discriminate and
  restore exactly. Actual runtime/parser/host over controlled HTTP/socket input,
  not live network/SQL/mint/full AC12. Both conflicts retain ordered refresh;
  correct row/clearance/Founder revision/native keyboard execute. RP-280 adds
  required conflict key with marked placeholder only; copy658/new producer,
  prior career reports historical. Types/build8106 units/boundaries/copy/manifest
  pass; RP-131/Firefox remain. Whole span afterb8ee639f needs Claude. Next RP-279
  all-state Amount cost diagnosis. No box flipped or full B7/AC12 closure.
  RP-277 row-busy correction under367fe467/27ff4cd5/294a1b08 first fails32
  native cases, then restored Chromium/WebKit pass58. Four compiling omissions
  discriminate and restore exactly. Existing returned host task/shared pending
  retain only the submitted busy row and disable controls until both settle;
  unrelated refresh does not revive the old row. Types/build8106 unit tests/
  boundaries/copy pass; RP-131 and unexecuted Firefox remain. Controlled tasks/
  arms, not host/SQL/mint/full AC12. Whole span after5bf6017f needs Claude.
  Next ground RP-278 host rejection/resync/row path; RP-279 costs remain separate.
  No checkbox flipped or old B7 acceptance implied.
  RP-276 focus correction under770a97bb/930b69d9/31795a1e reproduces native
  loss first and after compiling focus-call removal; restored Chromium/WebKit
  pass26 cases. Types/build8106 unit tests/boundaries/copy pass; RP-131 history
  stays RED, Firefox executes zero even outside sandbox. Component prop injection,
  not host/SQL/mint or whole AC12. Full new span after020a25c6 needs Claude.
  RP-277 now has the separate bounded correction above; RP-278 row rejection
  and RP-279 persistent cost remain source-contract findings. No checkbox flipped.
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
  RP-267's subsequent observer correction under821b867b now binds Reference
  candidates and waiting. Six reproduced omissions are corrected;18lifetime/
  16frozen-input observations and six compiling hook/mask omissions discriminate.
  Actual seed0 lifetime rises at both Exits; paid2/0, purchases and clocks stay
  unchanged. Cold core/vet, fast harness and client8106/134/types/build plus
  separate boundaries pass; composite remains RP-131 RED. Full H4/H5 remain
  report-drift RED in836.130s, retaining the same six Casual ties. Harness-only,
  kernel161/live math/balance/reports/corpus unchanged; span after973d983c
  needs Claude. Subsequent RP-262 under64badd7b/3d7c93ff executes all485 H3
  runs/3395 clocks: epoch8/no-row/unit neutrality passes, tiny factor changes
  0/679 clocks, strong control289. Tiny input changes lifetime, not milestones;
  exact H3 criterion fires. Four actual Reference controls and27 oracle/refusal
  children pass; five compiling probes discriminate with exact restoration.
  Cold core/vet/fast pass. Harness-only, kernel161 and old reports unchanged;
  complete new span afteree1064a0 needs Claude. H3 author reconciliation is
  required before closeout. RP-264 undera0aab5a6/5e3064fb/8ef50ccb now repairs
  false-exclusion/saving admission and stale census accounting. All28 synthetic
  profiles plus the legacy gate pass; nine compiling omissions discriminate
  with exact restoration, two as caught panic faults. Cold fast/core/vet pass.
  Complete97-pair H4/970-arm H5 finish787.881s, still six-tie/report-drift RED;
  separate current97-pair census confirms93 counted/3 valid excluded rows.
  Test-only, kernel161/old reports unchanged; complete span afterdb8398a3 needs
  Claude. Next RP-269 source/Exit admission diagnosis and RP-268 artifact
  authority grounding, not threshold minting or criterion changes. Subsequent
  RP-269 under87d526ed/e3620cda admits the ordered first-hour Exit pair and
  full H2 study source/aggregate coordinates, retaining generic subsets.
  Forty-five diagnostic children plus legacy threshold/H1/H3 tests pass;
  six compiling omissions discriminate with exact restoration. Cold fast/core/
  vet pass and retained30-point measurement remains byte-identical. Harness-
  only, kernel161 unchanged; whole span after9f5b81a0 needs Claude. Next RP-270's
  shared H3 aggregate-source gap, then fresh H1/H2/lineage and RP-268 grounding.
  RP-270 under3d117b39/9ced05cf now corrects that coordinate admission:31 H3/
  45 H2/28 H4 diagnostic children and legacy checks pass; compiling omission
  fires four H3 corruptions while H2 remains defended. Exact source restoration,
  cold fast/core/vet pass. Full485-run/3395-clock study completes313.732s:
  neutrality passes, tiny0/679 versus strong289; H3 criterion remains RED.
  Test-only, kernel161/old reports unchanged; whole span after0a486d79 needs
  Claude. Next fresh H1/H2/lineage. RP-271 requires author reconciliation of
  each-Exit purchase wording vs scripted-first plan prohibition before RP-268's
  missing data artifact can fix policy authority. No criterion/box promotion.
  RP-263 underc095e2b5 now records fresh dated H1/H2/lineage artifacts from
  committed producerda512cd3. Full97 runs/679 clocks complete twice with exact
  byte-identical reproduction; all30 candidates use pinned suite Prestige.
  Proposal set unchanged; three Reference rows differ, no adoption. Twenty-four
  refusals/four compiling omissions discriminate; cold local fast/core pass.
  Old v1 files/kernel161/live math/balance/CI/corpus unchanged; whole span after
  c352370a needs Claude. H4/H5 report-envelope provenance remains separate,
  alongside fired H3/H4, H5 gaps, author boundaries and all prior reviews.
  Subsequent RP-263 underec71a33c now records effective career inputs and binds
  both report consumers, retaining both H4 sources and all H5 arm sources.
  Four real careers preserve whole gameplay fingerprints;36 refusals/nine
  compiling omissions discriminate. Full97-pair/970-arm study636.768s passes
  194/970 source admissions, retaining970 H5 sources; H4 six ties and both
  historical comparisons remain RED. Cold local fast/core pass. Whole span
  afterdfc2fb8c needs Claude. RP-272/273 denominator diagnosis and RP-274 per-node
  attribution remain separate; no software-provenance artifact refresh or
  author-policy/epsilon adoption. Docker still100%/39784KiB free; no cleanup.
  RP-272/273 under69c79d28/b22ce51e subsequently expose report populations and
  all-finite H4 descriptive savings. Synthetic controls/eleven compiling omissions
  discriminate; complete97-pair/970-arm census finishes547.497s: Casual H4
  28 finite/six ties/one treated-only/three both-unreached, H5 cash/tower finite
  medians28/32 and20/23 bought pairs. Cold fast/vet pass; gates stay RED.
  Full new span aftera3e36f94 needs Claude. Dated H4/H5 provenance remains next,
  separately predeclared, not adoption of censoring or unresolved author intent.
  RP-275 under1f6a09ca/d96bfaf1 subsequently retains raw observations and shares
  composition with ordinary H4/H5 checks. Nineteen corrupt/invalid controls and
  twelve compiling omissions discriminate. Full97-pair/970-arm study641.626s
  retains/recomposes970 raw H5 observations, medians/populations unchanged;
  strict ties and both old comparisons stay RED. Cold fast/vet pass. Whole
  span after5e083766 needs Claude. Dated producer identity/full replay remains
  separately predeclared next; no new dated artifact or acceptance here.
  RP-263 dated career observation instrument is subsequently predeclared at
  80cc4015: exact97-pair/970-arm source/cohort admission, raw recomposition,
  H4/H5 baseline binding and committed provenance. Synthetic controls/eight
  compiling omissions and cold fast/vet pass. Subsequent dated record/replay at
  6bd24e8d completes676.102/469.043s:97 pairs/194 sources and970 H5 arms each,
  byte-identical reports, six H4 failures unchanged. Actual overwrite refuses
  before production; retained-artifact admission and post-artifact fast pass.
  Whole new span after5c93ff6f needs Claude. This resolves local dated career
  observation/reproduction, not AC13 or author policy/attribution/epsilon/
  career-data/mint questions. Original files preserved; H1/H2 keep their earlier
  producer, not freshness at this changed server tree. Next ground remaining
  R9/AC12 client-consumer coverage under B7 without claiming minted/real-DB
  integration, changing ruled copy or treating absent native Firefox as a pass.
  No AC13/box/mint promotion. At the earlier checkpoint below,
  fresh H1/H2/report provenance/reproduction remained separate; no retuning or
  report-regeneration authority here. No checkbox flipped.
  Complete new span after7da200f0 needs Claude. No checkbox flipped here.
- [ ] B9 — BLOCKED on the owner-gated mint (no epoch pins a tree; formulas-check clean) — formulas regeneration (separate commit), composed verification career. ACs 14, 15.
- [ ] Canonical docs, designated Codex review, archival (Codex).

Kernel protocol: every commit touching a `kernel/affecting-paths.json` prefix bumps
`kernel/VERSION` (+ Go/TS constants) in the same commit; `server/reputation/` and
`client/src/reputation.ts` are registered in the commit that creates them. Save/snapshot versions
are assigned in landing order (the RFC's "v22" means next-free).
