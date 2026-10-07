# Garage acceptance reconciliation

Source coordinate: `fc911784`, inspected again after predeclaration `124b086a` on
2026-10-07. This is an acceptance inventory, **not** an approval, checkbox flip,
RFC amendment or release certificate. The goal remains the complete nine-tier 1.0.

The accepted authority is [Garage Player Surfaces](../../rfc/garage-player-surfaces.md).
Later accepted producer RFCs do not silently amend its body. In particular,
RP-132 (pet contract) and RP-313 (opportunity sibling arm) still require their
ruling author's reconciliation. Existing implementation boxes do not prove the
acceptance gates below.

## Evidence coordinates and limits

- `G`: `server/gameui/features_test.go`, `projector_test.go`,
  `opportunity_test.go`, `opportunity_review_test.go`, `pet_care_fact_test.go`.
  Fresh root `make test-go GO_PACKAGES='./gameui ./fiscal ./achievements ./meters'
  GO_TEST_FLAGS='-count=1 -v'` exited 0 (handle 51228): 116 PASS lines including
  parent/subtest reports, **not** 116 independent top-level tests. The stored
  rate projector and Reputation public projection SQL tests explicitly skipped
  because `TEST_DATABASE_URL` was not set. No fresh persisted proof.
- `U`: `client/test/game-ui.test.ts`, `garage-surfaces.test.ts`,
  `game-ui-intent-outcome.test.ts`. Last full unit/types/build/boundary run at
  `2d3629ad` passed 9,758 client tests with 442 explicitly skipped browser
  declarations; those skips are not browser evidence. No fresh TS run in this audit.
- `B`: `client/test/garage-surfaces-browser.test.ts`, actual native Chromium
  and WebKit. Last full population at `a42406f0`: 238 functional executions
  and the separate performance observation. Runtime doubles and public snapshot
  fixtures, not real HTTP/server/DB/AT evidence. Compiler-only fault runs do not
  count as discriminating behavioral probes.
- `C`: `client/tools/test-game-ui-composed.mjs`, actual server/Postgres/socket/DOM
  driver. Latest successful execution predates production notice change
  `45bf2edb`; its producer/claim/Fiscal/Pitch observations remain historical,
  **not** a rerun of the changed host. The separate cosmetic composed driver
  recorded real adoption/feed/reload, not raw-care or all GS4 behavior.
- `F`: cold Linux three-engine browser lane at `36c28692` exited 2;
  Firefox imports/keyboard/worker failures and a WebKit worker observation remain
  RP-327/328. Subsequent passive diagnostic stopped with no final population
  verdict after measured Docker overlay 0 available / 100% full (RP-236).
  No third container population or unrelated cleanup is authorized by this file.
- `R`: exact review ranges and executed fault tables in [log.md](log.md) and
  [plan.md](plan.md). Codex's first filters cannot supply Claude's designated
  verdict on Codex work. Original Claude producer/consumer findings and all
  separately owed corrective/record ranges remain live.

## Named GS1–GS6 gates (29)

Each row names the actual producer/consumer and the strongest relevant evidence
found. “Partial” means the row's **whole accepted population is still open**;
it is not a percentage-complete claim. A seed listed by the RFC is not executed
evidence merely because its text exists.

| Gate | Producer → player consumer | Evidence present | Remaining acceptance / route |
|---|---|---|---|
| GS1-A1 projection | `features.go:projectFiscal` → `contracts.ts:parseFeatures` → `FiscalSurface.svelte` | G `TestFiscalArmSweepPreviewMatchesTheNextHarvest`; U v4 Fiscal decode | **Partial:** preview/next harvest comparison is pure kernel/clone, not the specified real Postgres projection/receipt witness. Shared v4 fixture and relevant decoder fault population also need exact proof. Add persisted witness after RP-236 capacity repair. |
| GS1-A2 revision | `GameUIApp.svelte:act` → Fiscal callbacks → runtime | B phase/revision and held-refresh supplement, plus RP-349 native host paths using Founder7→8 versus Company1→2; actual C Fiscal commands before host change | **Partial:** new native exact-payload/refusal/held-reply/read/context-survival proof has eight complete restored faults. No fresh changed-host composed run or designated corrective approval. |
| GS1-A3 phases | public v4 fixture → actual `fiscal-phase.ts`/FiscalSurface → native exact text/readiness | Unit exact 99/100/199/200 and retained host margin samples; new decoder-admitted component fixtures at 99,999/100,000/199,999/200,000 under unchanged 100k/200k/300k period. Eight source faults fail phase/text/visibility/risk/countdown/native readiness/callback checks in both engines, restored exactly. | **Locally witnessed fixed-time rendering; designated review pending** for full range after 57efdc39. Exact registered copy and actual visible DOM, not color/data-phase/prefix only. No fake timers or live-clock/server/pacing claim. RP-336 props/body correspondence and Fiscal mint/owner/real-service/Firefox/AT gates stay open. |
| GS1-A4 refusals | receipt + mapper → Fiscal own polite notice | B ordinary/cap/unknown/exclusive/HTTP populations and RP-324/326 corrections; RP-349 delayed ordinary refusals/native retries/exact Fiscal-only notices | **Partial:** native two-engine fixture proof and deliberate notice/role/context-leak faults. Exact review ranges and all-engine/all-state proof open. Real HTTP service failures are not supplied by thrown runtime-double errors. |
| GS1-A5 accessibility | Fiscal DOM + host chrome/nav | B axe, native Enter/Space harvest/level/unlock, pending/stale/focus/320 px supplements; RP-34924 complete keyboard paths with delayed reply/read and newer native Settings selection | **Partial:** whole-page native reflow/axe/newer-focus boundaries locally witnessed, not all-state/three-engine/400% zoom/manual assistive evidence. RP-350 contaminated probe and RP-351 incomplete runs excluded; exact new range needs Claude. F is RED, not waived. |
| GS1-A6 composed | real Fiscal harvest/spend → projected Pitch fact/nav | C actual DOM harvest/unlock/Pitch sequence; recorded receipt sequencing repair | Historical source proof only; changed host must rerun after capacity repair. No synthetic clock allowed; short live clock is not approval of final pacing. |
| GS2-A1 state text | `projectAchievements` → `AchievementsSurface.svelte` | Original earned-run/lifetime/locked assertion plus visible state/scope/grant/title/possession/public-score fixtures under6c578197; G live arms/overlap refusal | **Bounded fixture proof; review pending:** actual state-text omission fails original and new gates; CSS-hidden state survives the original raw text test but fails new native visibility. RP-332 missing-copy error state is now explicitly checked here. No server acquisition/all-engine/AT claim. |
| GS2-A2 once | `events.ts` cursor → host chrome announcement | B simulated reconnect/dedupe; RP-312 repair and executed omission probes | Bounded announcement proof; not actual server acquisition, full reconnect/engine population or designated repair approval. |
| GS2-A3 not Clout | achievement score projection → score copy/component | Existing DOM assertion plus RP-337 existing-tool extension under 305b8e8b/a32b8358: actual parsed Trophy Case + canonical source achievement copy/era variants; seven component/four copy negatives | **Locally corrected source guard; designated review pending:** actual identifier/computed-member/base-label/era-label/checker-bypass faults fail semantically and restore exactly. Comment/score/separate-Clout controls pass. Not whole-program/computed-dataflow or real persisted-player proof; all prior review/GS2-A4/full acceptance gates remain. |
| GS2-A4 composed | first real generator purchase → earned event + refreshed row | C buys generators; separately reads achievement rows exist | **Missing named witness:** driver does not assert that purchase earns this ID, emits its announcement and changes the displayed row. Add DOM/event/refresh binding and independently sever decoder/arm after RP-236. |
| GS2-A5 accessibility | Trophy Case DOM + chrome/nav | Visible populated/empty/error fixtures and RP-338 native nav repair retained. RP-339–341 red baseline then separate88153822 host scope: forty held-read/reconnect/not-ready/null-arm/cancel native executions at320/1280, nine real source faults | **Locally repaired bounded populations; review pending:** full Garage326 and performance two pass. No forced Tab focus/provider waiver. Other forced contexts, first-read failure/retry/all conditional nav states, Firefox, actual400% zoom/manual AT and real acquisition remain; not complete GS2-A5/full platform acceptance. |
| GS3-A1 decoder | v4 meter rows → `contracts.ts:parseFeatures` | Audit found incomplete-set admission; subsequent RP-329 failed-first population covers all eleven missing IDs and same-count substitutions, extra/subset rows, order, bounds and exact keys | **Locally repaired; designated review pending:** existing REQUIRED_METER_IDS reused, 27 baseline failures and five independent client faults demonstrated; two actual producer-control faults also demonstrated. Clean client/types/build/boundaries/cold Go/vet and 246 native Garage executions pass. SQL explicitly skipped; full shared-v4 fixture and all-engine/persisted proof are not inferred. See separate follow-up below. |
| GS3-A2 non-color | committed values → `MetersSurface.svelte` numeric/band text | Original B cell checks; later RP-330 exact per-ID value/band assertions in both layouts, with omitted-band/native-value faults caught; RP-332 presentation-error population retained | Bounded public-fixture proof in Chromium/WebKit; designated review and whole-state/all-engine/real persisted populations remain separate. |
| GS3-A3 reflow | Meters arm → responsive semantic list / retained wide table | RP-330 failed-first six-case baseline; distinct decoder-admitted values, exact eleven dt/dd/native-label associations, live resize and refresh, full page/nav/focus/axe; six independent faults demonstrated | **Locally repaired; designated review pending:** current full Garage Chromium/WebKit 252 passes, no assertion/budget loosened. Source/body mismatch corrected without author amendment. Viewport/DOM associations are not actual 400% zoom, Firefox or AT evidence. See separate follow-up. |
| GS3-A4 composed | real initial pinned meter values → rendered DOM | C authentic public read asserts 11 rows and doom=50 | **Missing named witness:** not DOM text for each Standing=50/Grievance=0/doom=50; add that and arm-severing case after RP-236. |
| GS3-A5 accessibility | meter DOM/header semantics → reading/navigation | B static axe plus semantic narrow/wide populations. RP-343: recovering/closed/resync/server-restart/not-ready-only ×320/1280 ×Chromium/WebKit, exact eleven retained values/bands, visible stale note, recovery without invented changes, explicitly newer snapshot, native focus/layout/axe/read-only controls; six actual faults | **Locally repaired bounded shared-state proof; designated review pending:** red checkpoint 0a191ebd followed separate bfd7b4ff product scope; same 20 cases pass, full Garage 398 and performance two pass. Whole 4b2fd904-exclusive span requires Claude. Three-engine/all-state/manual assistive workflow, real service, actual zoom and full acceptance remain. Native associations and axe are not screen-reader evidence. |
| GS4-A1 care states | PA7 public identity/band/eligibility → `pet/PetCareSurface.svelte` | B availability/refusal/pending/keyboard/captured notices; actual decoder-accepted PA7 fixtures | **Author-blocked RP-132:** accepted GS4 asks raw stats/mood/behavior/cooldowns while PA7 exposes another arm. Current tests must not invent raw private fields or claim that contract covered. |
| GS4-A2 absent pet | `pet_care_fact_test.go` adoption count → nav unlock | G `TestPetCareFactRequiresAnAdoptedPet`; B petless nav absent | Bounded Go/fixture proof; accepted historical “epoch-8 fresh Founder” is not a current release pin. Preserve source/epoch identity and separate current real-server absence assertion. |
| GS4-A3 revision | Founder `care_action` → runtime/refresh | B Founder7/8 command and held refresh; separate cosmetic composed adoption/feed/reload | Bounded public-arm proof, not full raw-care contract. Corrective ranges still need Claude; changed host real service rerun remains. |
| GS4-A4 tone | owner pet copy → Copy Pipeline/lint | `client/tools/copy-pipeline.mjs:334–345` companion rule bans price/urgency/curtain/satire tokens; `verify-copy.mjs:74–81` executed failing fixtures belong to the existing copy gate | Dedicated gate **exists**, not merely general lint. Last recorded copy pass predates this audit; no fresh run claimed. PA8.2 requires `companion`, whereas this Garage body still says `diegetic`: author-owned reconciliation remains, not implementer copy edits. |
| GS4-A5 composed | adoption producer → real care/receipt/state/reload | Historical cosmetic composed positive feed witness | DG-4 producer now mechanically exists, but RP-132 body reconciliation and RP-318 missing pet-status decoder/copy remain. **Not silently cleared:** full GS4 population is blocked, never a skipped success. |
| GS5-A1 stable order | optional opportunity arm → persistent Desk region | B stable region/index/focus native test | Bounded proof; original wire mismatch RP-313 remains. Scope review does not follow from local tests. |
| GS5-A2 no synthetic commands | host cadence → no idle intent | B actual 60 s native wait; RP-315 timer correction/probe | Bounded two-engine actual-timer evidence, not a shortened virtual run. Full F/hosted population still RED/held. |
| GS5-A3 cap | Lucky/buff receipt → Desk notice + pinned reason | B saturated Lucky/capped-buff/projected cap; RP-314 compiling omissions; RP-352 exact complete credited/cap text on held reply/read | Bounded receipt fixture proof, not real Lucky payout. Valid wrong credit fails24 cases; invalid RP-355 numeric fault discarded. Two old palette leaves each independently fail actual capped-bank axe. Corrective ranges awaiting designated review; RP-313 authority hold independent. |
| GS5-A4 composed | real DOM clicks/claim → receipt-bound successor | C actual click buff at revision26; shared driver's oracle unit/native tests and six faults | Actual buff branch only, **not Lucky integration**. Changed-host rerun held; old vacuous shortcut survivor and corrected fixture disclosed under RP-316. |
| GS5-A5 accessibility | opportunity region → native claim/text | B Tab + Enter/Space, axe and 320 px; RP-352/35320 declarations/40 actual two-engine executions | Locally repaired pending and removed-control focus under separate scopes;14 valid source faults/restorations and full native734/performance2 green. Not all-state/engine/400%/AT proof; connection states (RP-356), announcement timing/off-surface withholding need their own populations. Designated review still owed for the entire f5be35f6-exclusive new span and prior spans. |
| GS6-A1 cap reason | generator `provision_cap` → Desk card | B provisioned=cap fixture checks exact reason; G generator projection | Bounded reason-text proof; executed omission seed and exact review coverage must be pinned for this gate. No whole Desk acceptance inferred. |
| GS6-A2 no fake purchase | production Go predicate/decoder → real component AST → existing root boundary | RP-335: a valid registered-copy fake legacy-shelf button passed the old boundary; new verification-only guard rejects it, a replaced/dynamic real callback kind, a missing production decoder case and a checker bypass; exact sources restored | **Locally witnessed source guard; designated review pending.** Ten Go/eleven Svelte negatives, legitimate acquire/equip/unequip and current source pass. Explicit envelopes/current wrappers only, not whole-program analysis, eligibility, persistence or complete Shop acceptance. Later accepted Shop commands are not banned; any stale Garage body reconciliation remains author-owned. |
| GS6-A3 no wider Desk | paired public v4 fixture → current full 1995 Desk/chrome → native 320 px measurements | Test-only supplement predeclared 757f1ca2: provisioned=0/owned=false vs provisioned=cap=3/owned=true, identical other content and cap. Chromium/WebKit document/main 320/320, Desk/chrome 284/284, extents 0–320; 90→93 measured nodes. Six real faults fail both engines; sources restored. | **Locally witnessed fixture comparison; designated review pending.** Absence/presence in current source, NOT historical executable, full later Desk, actual 400% zoom, Firefox, AT or real service. Accessibility retains ownership of its old 647 px record; no acceptance/lifecycle promotion. Complete range after 54c8c8da needs Claude separately from source guard d3ce0f76..54c8c8da and every preceding span. |

Shared-context supplement underfb2a9488: Fiscal/Meters/Reputation each mounts
an exact-decoder admitted arm; null with retained/removed feature fact at
320/1280 returns/focuses Desk; subsequent native Settings choice survives
repeat null. Twenty-four native Chromium/WebKit passes and five valid source
faults discriminate (three individual branch omissions, no-op focus, heading
tabindex removal). Host restored byte-identically. Full Garage350/performance
two and client/types/build/copy/boundaries/topology pass. Local first filter,
not lifecycle/all-context/first-read rejection/retry/Firefox/AT/live SQL proof.
Separate entire82614848-exclusive range requires Claude.

## RP-352–355 native Claim correction and contrast evidence

Twenty declarations execute40 times across Chromium/WebKit. Three held-host
paths (retained offer, removed Claim, newer native Settings selection) ×widths
320/1280 ×Enter/Space =24; independent immediate-success/removal eight and
isolated pending/unavailable callback eight. Actual public decoder and Company
1→2 versus Founder7→8 distinguish scope; exact requests/notices/receipt credit,
read counts, delayed pending/native duplicates, native return, focus, reflow and
axe asserted. Only initial manual focus is seeded; Tab reaches host Claim and
later navigation without direct action focus. Isolated callback cases prevent
the existing host guard from hiding a removed component guard.

Typed baseline36fail/four unavailable controls, separate component scope then
32 contrast failures/eight controls, separate two-leaf C9 provisional palette
scope then40pass. Fourteen valid faults all fail their intended properties and
restore exactly; RP-355's non-canonical0e0 arm discarded, valid1e0 replacement
fails24 actual whole-credit comparisons. Component guard omission caught only
by four isolated cases, explicitly not attributed to the36 passing host cases.

Final full native734/four isolated-performance skips plus performance two;
client9816/669 explicit browser skips/types0/build214/copy/boundaries/topology/
no-payment pass. Real60s idle objective retained. Same current wire fixture,
not acquisition/expiry/payout/real HTTP/SQL/Firefox/AT/400%/all-era/whole-GS5.
RP-313 authority and all full-release holds remain. Entire f5be35f6-exclusive
new span through final records requires Claude; earlier spans independently owed.
No acceptance checkbox/status/archive/mint/push/release promotion.

## RP-349 native Fiscal delayed completion (test-only)

Under aa296eba, twelve test declarations execute24 times in Chromium/WebKit.
Only initial Desk nav focus is seeded; every later selection/action/return uses
actual Tab and native Enter/Space. Three actions at320/1280 have held ordinary
refusals, then applied replies and held reads. Pending remains focusable and
guards duplicate input; Fiscal owns notices. Newer native Settings selection
retains focus/context across read completion; returning and retrying binds the
new Founder revision8, not prior7 or Company2. Exact bodies/IDs/status/read counts,
full-page reflow and axe asserted. Public decoder fixtures retain eligibility;
not a natural quarter, persisted purchase, removal policy, auth or AT session.

Eight complete faults fail24/24/24/24/24/24/8/24; all production restored exactly.
RP-350's contaminated run and both RP-351 incomplete/stopped runs explicitly
excluded. Full current browser694/four isolated-performance skips plus separate
performance two, client9816/649 explicit browser skips/types/unchanged build/
copy/boundaries/topology/no-payment pass. No Linux/SQL/Firefox/hosted or whole-GS1
acceptance; c24b8f6d-exclusive new span through containing records needs Claude,
every prior range independently owed. Original RP-348 remains unexplained.

## RP-346 / RP-347 standard-terminal keyboard recovery

2e5ba552 test scope → a3dab732 typed red (48 WebKit/48 Chromium controls) →
fd56b866 separate tab-stop scope → partial retry red24df13 → diagnostic
fec61e (BODY loses document focus) →38d8d650 separately declared pending-focus
scope. Stronger pending/focus/duplicate oracles in both populations fail all152
before aria repair. Current control is explicit tabindex0, pending aria-disabled,
with existing actual pending guard. No read/intent/auth/copy/navigation policy
change, extra keyboard handler or raised traversal bound.

96 native executions = four registered standard labels ×tiers0/1/2 ×widths
320/1280 ×Enter/Space ×Chromium/WebKit. Every case is non-first run2 and performs
four refused native read stages then healthy same-Founder run3/Garage. Actual
public snapshot/terminal decoders; native Tab reachability without focusing
Continue in script; exact complete terminal paragraphs/payout/era/title, retained
terminal/focus on refusal, no subscription change/intents, Desk heading focus,
then native Settings saved-state reader. No production issuance/natural ending
reachability/physical AT/actual HTTP/auth/SQL proof. New96 + stronger old56 pass.
Eleven actual faults fail48/96/72/104/112/104/128/96/152/152/152 with stated
positive controls and exact restoration. Omitting the function pending guard
now actually yields duplicated reads; prior HTML disable did not prove it.

Full native670/four isolated-performance skips; separate performance two;
types/build/copy/boundaries/static topology/no-payment pass. First full client
fails Garden child-control entry (RP-348); unchanged rerun after native/copy
terminal passes9,814/637 explicit skips. This does not resolve that timing cause
or establish hosted/global CI. Fresh Docker read still0 available/100%; no
Linux/Postgres rerun or cleanup. Entire0e7910ed-exclusive through containing
repair/records requires Claude including all predeclarations/red/refinement.
Prior0e548992..0e7910ed and every earlier range remain independent. No boxes/
lifecycle/RFC body/release promotion; full real workflows/1.0 stay open.

## RP-345 / RP-082 bounded native continuation

19cc5609 test-only scope → red94c88b72 (16 failures/32 controls) → separate
accepted repair160eeb0d. Eight other-Founder cases wrongly bound Desk; eight
healthy continuation-focus cases left BODY. Existing same/skipped sequence and
rejected reads refused. Initial spy annotation error disclosed; typed red repeat
unchanged. Host now checks terminal Founder+exact-next sequence before binding
and hands off removed native origin to Desk heading after rendering without
displacing a surviving newer focus. Navigation during held read remains existing
Desk-on-completion policy; no new cancellation/auth/credential/category rule.

56 native executions: seven arms ×320/1280 ×Enter/Space ×Chromium/WebKit.
Actual public envelope/event and snapshot parsers admit fixtures; held runtime
double read, no intents, duplicate native activation stays disabled, stable
terminal on refusal, unchanged Founder subscription, exact heading focus,
surviving Settings focus, reflow and axe. Seven compiling faults fail8/16/8/8/
56/32/8 cases respectively; source restored. Pending-control fault proves native
disable, not isolated entry-guard necessity. Broader focus fault does not prove
isolated scheduling-generation guard necessity.

Combined Garage/Game UI574 passes/four isolated-performance skips; separate
performance two. Types/client9,814/589 explicit skips/build/copy/boundaries/
static topology/no-payment pass. Entire0e548992..0e7910ed
requires Claude, including both scopes/red/control
refinement. All earlier spans independently owed. Only scripted terminal here;
old standard-terminal click test is not native/focus proof. Real composed/HTTP/
auth/SQL/Firefox/manual AT/full1.0 remain open; no lifecycle promotion.

## RP-082 bounded forced lifecycle context focus

Accepted GS0.4/0.6 own floor, not acceptance of the broader Accessibility draft.
3451fdb0 test predeclaration → 52 native focus failures / 8 controls at 73a5ccce
→ separate product scope ac6a5788. Existing cursor-ordered host navigation now
focuses the actual Offer/Run-End heading after DOM update; both headings are
non-tab-stop focus targets. Newer player choice cancels pending focus. Sixty
executions cover Fiscal/Trophy Case/Meters ×nav/removed-heading ×destination
×320/1280 ×two engines plus Settings cancellation/ordered lifecycle/replay.
Real decoders admit public fixtures; runtime-double delivery, explicit heading
focus and synchronous race controls are not a physical keyboard/AT/server session.
Eight compiling actual faults fail 52/24/28/24/28/8/4/12 cases respectively;
remaining controls pass and every source fault is restored. No assertions,
budgets, copy, lifecycle precedence, payload-only Run-End or clock changed.

Combined Garage/Game UI 518 native passes / 4 explicitly isolated performance
skips; separate performance two pass. Client/types/build/copy/boundaries/
no-payment/static topology pass. Full fbea2158-exclusive through repair/record
commit needs Claude separately from prior spans. Firefox/manual AT/real service/
continuation/all contexts/full shared accessibility remain open. Changed host
still needs real composed rerun after capacity repair; no lifecycle/acceptance
promotion from the native fixture proof.

## RP-344 GS3 event/badge/value separation (test-only)

96439cfe predeclared; existing combined oracle survives an actual off-surface
global-announcement fault in both engines. New sixteen native executions:
p(doom)/users Standing ×up/down ×320/1280 ×Chromium/WebKit. Production
envelope/announcement decoders admit fixtures before runtime-double delivery.
Off-surface badge leaves existing global text unchanged; visiting clears badge
without new announcement; on-surface exact polite text; a distinguishable
opposite event followed by consumed-cursor replay does not overwrite it.
All eleven values stay committed through events and only change on explicitly
newer snapshots. Unknown presentation withheld/diagnostic; native focus/axe/
layout/read-only checks. Nine compiling source faults each fail sixteen cases,
including the exact forbidden global fault the old test survived. Production
restored byte-identically. Initial invented separator-space assertion failed;
corrected fixture expectation, not a product/copy defect or weakened semantic
boundary. Whole 9f485993-exclusive proof/record span requires Claude.

Full Garage 414 / performance two, client/types/unchanged build/copy/boundaries/
static topology pass. Not all IDs/eras/states, real socket/server/reconnect,
manual AT/Firefox/actual zoom/SQL/full GS3 or full CI proof. RP-343 review and
GS3-A4 real-service value binding remain separate; Docker overlay freshly
reports zero available/100%, no cleanup authorized or new container run.

## RP-342 first-read measurement, not acceptance of failed display

Predeclared ced3b5c7; actual runtime/Response.json/decoder feeding native host,
injected responses and socket. Both selected engines at320/1280 agree:

| First reply | Settled read display | Explicit visibilitychange + next reply |
|---|---|---|
| Healthy v4 | loading absent,13 controls,one socket | Not needed |
| Network rejection /503 /malformed JSON /malformed v4 arm /valid legacy v3 | loading visible,aria-busy=false,zero controls/socket/failure text/alerts/diagnostics | Valid same-Founder v4: second read,loading removed,13 controls,one socket |
|401 | Same failed loading display,credentials unchanged | Repeated401: second read,loading remains,no socket/renewal/bootstrap |

28 native executions,52 recorded settled/lifecycle rows; no hidden retry in
350ms real observation, not indefinite absence proof. Actual legacy v3 is
decoder-valid after removing v4-only feature/provision fields. Initial wrong
legacy fixture and TaskMeta typing/report-output mistakes disclosed in log.
No desired-bug assertion: rows are observations; rejection/healthy/read-count/
identity controls validate measurement. Four faults fail24/28/24/28 cases.
Exact source restored. Full Garage378 and performance two/client/types/build/
copy/boundaries/topology pass; not real server/auth/network/physical hide/
all-engine/manual AT/release proof. The retained JSON is the bounded metadata
projection of a complete parsed passing runner report, not a new acceptance
gate. Rows do not independently label engine; root invocation names both.

DESIGN-GAP: failure precedence/retry/wording before a first snapshot requires
D-023 and acceptance of the draft first-read recovery RFC. Instrument green
does not close RP-342. Entire a551d3c2-exclusive span requires Claude.

## Overall criteria (8)

| Criterion | Current evidence / remaining gate |
|---|---|
| 1. All per-surface gates discriminate | The 29-row inventory above is not a pass. Fault tables prove particular repairs, not every listed population. GS7's adopted API lane is separate and is not re-certified here. |
| 2. Shared projection parity | Go/TS positive/negative tests exist, but inspected Garage fixtures are independently constructed, not one identified shared v4 artifact. Five required seeded classes and producer/decoder same-fixture binding need explicit inventory/proof. |
| 3. Runtime receipts | `a42406f0..fc911784` adds 21 actual-runtime/Response.json cases (injected fetch, not live service); existing outcome tests retained. Typed outcomes/errors/unknown outcome/exact rejection, one unchanged request and four behavioral faults are recorded. No generated-client or renewal conformance inferred. |
| 4. Boundary | Current-source baseline and restored root boundary/types pass after d3ce0f76. Four independently executed valid-Svelte OpportunityRegion faults (literal text, raw fetch, real shell/runtime import, literal color) fail named policies. RP-335 additionally binds cosmetic source kinds to the production Go AST/decoder with five real faults caught and restored. **Locally witnessed source policies**, not runtime behavior or designated approval. Complete new range after d3ce0f76 requires Claude separately from the performance range 127eb052..d3ce0f76. |
| 5. Accessibility | Native Chromium/WebKit subsets pass. Linux three-engine lane is RED and diagnostic environment invalid; manual AT, every state, full-page 400% zoom/coarse-pointer target measurements remain open. |
| 6. Composed | Historical Fiscal→Pitch and buff claim paths are real; GS2-A4 and GS3-A4 DOM assertions missing, GS4 accepted-body hold persists, changed host not rerun. RP-236 must be repaired/rechecked without unowned cleanup before Docker evidence. |
| 7. Performance | Original guard retained byte-identically but RP-334 native probes expose omitted-input and final-task survivors. New isolated GS5/GS6 populated Desk supplement counts actual delivery, verifies all named regions per input, requires visible activity/exact terminal cash/native observers and drains final records; six real faults fail. **Partial, not promoted:** GS4/pet/later Desk/full release population, manual 4× profile, actual default player and designated review remain. See the RP-333/334 follow-up below. |
| 8. Closeout | Canonical docs describe implementation/limits, but producer body holds and separately owed Claude ranges remain. No RFC status change, checkbox flip, archive or release promotion authorized. |

## Shared error-state investigation and next bounded work

GS0.5 requires a contained `common.surface_error` alert, no controls, one
invariant and other surfaces still operable. Source inspection found
`AchievementsSurface.title()` and `MetersSurface.band()` throw for missing
copy/presentation; `GameUIApp` has no render boundary. Their decoder admits
an unknown mechanical achievement copy key and a meter band declared in the
wire but absent from presentation. **Candidate defect, not yet a fresh executed
browser failure.** This is distinct from malformed wire, module-import failure
or changing authoritative data.

Next: separately predeclare decoder-admission/healthy controls and native
mount/navigation cases for those two populations; record actual crashes and
contract assertions. No production repair inside this audit range. If reproduced,
use a separate failed-first GS0.5 repair range with existing copy, error isolation,
one diagnostic per failing surface and retained navigation. Do not broaden it
into an invented renderer, wire contract or owner-copy rewrite.

Following safe accepted work: GS3-A1 complete-ID decoder; GS3-A3 semantic narrow
layout; exact source guards; missing real DOM acquisition/meter
witnesses after capacity restoration. RP-132/313/318 author obligations remain
separate. No additional empirical or owner decisions are answered by this file.

## Follow-up: RP-332 / RP-333 (separate range after cc62cea8)

The candidate above was reproduced after this audit: predeclared `4c9e06b2`,
typed-clean failed-first `532855fe` (eight native rendering/assertion failures),
bounded repair step `81b4f51d`. The two components now contain unavailable
achievement-copy / declared meter-band presentation with an own alert and
once-per-error-episode invariant; healthy authoritative refresh recovers, other
navigation remains operable, no synthetic intent. Five compiling faults each
fail four of eight focused executions, exact source restored. Fresh full Garage
Chromium/WebKit passes 246 executions; client/types/build/boundaries/copy/
topology pass. This is a local correction, **not** designated approval or full
GS0.5/GS2/GS3/accessibility/composed acceptance. Exact evidence in log.md.

Further performance inspection identifies RP-333: the unchanged screen fixture
in `client/test/game-ui-screens-browser.test.ts` has every feature arm null.
Its isolated passing 1,200-update scenario covers sixty **simulated** seconds,
not the populated Garage Desk regions required by overall AC7. Retain that
existing guard and add the missing population separately; do not loosen its
budget or infer a performance regression from this source gap.

Audit range `fc911784..cc62cea8`, the subsequent whole RP-332 range through its
record edge, and all prior ranges need their own Claude verdicts. RP-329/330/331,
Docker capacity (fresh read-only check still 0 available / 100%), Firefox/AT/
author/body/default-workflow/full nine-tier 1.0 holds remain independently open.

## Follow-up: RP-329 / GS3-A1 (separate range after 5fbf4cff)

Predeclared bff97b20/c5d5220c; failed-first 43ef0a21. The original doom-only
positive fixture is now complete, generated from the existing public meter
artifact, with its original domain/undeclared-band checks retained. Fifty-six
new declarations prove complete/null/unchanged-input controls and every
missing ID, same-count unknown replacement, extra/subset, ordering, value-bound
and exact-field refusal. The typed-clean baseline has 27 actual admission
failures. The repair imports the existing immutable REQUIRED_METER_IDS without
editing the guarded catalog or inventing a second list/count.

Five independent client faults fail 27/11/4/3/23 assertions respectively.
Cold actual projector tests also catch an omitted emitted row and each of
eleven missing saved values when the guard is severed. All transient changes
are restored exactly; the server production diff is empty. Root clean types,
9,814 client passes / 446 explicit browser skips, build/boundaries/cold Go/vet
and full native Chromium/WebKit Garage 246 pass. The existing isolated
performance test passes, with RP-333's null-feature limitation unchanged.

This closes the *local decoder defect*, not designated review, whole GS3,
overall shared-v4 fixture parity, all-engine accessibility, real persisted
player workflows or release readiness. Both SQL tests explicitly skip outside
the declared Postgres lane. Docker capacity and RP-330/331/333 remain open.
Complete new range after 5fbf4cff through its final record edge needs Claude,
independently of RP-332 cc62cea8..5fbf4cff, audit fc911784..cc62cea8,
runtime a42406f0..fc911784 and every earlier range.

## Follow-up: RP-330 / GS3-A3 (separate range after bddfc58e)

Predeclared d9a8ca27/639add6e; test-first 09655d1e. Corrected baseline
50408 reproduces six real narrow semantic failures with healthy wide controls.
Initial raw-HTML-value instrumentation and a later native-media-event timing
observation error are disclosed in log.md; neither is a softened assertion.
The final driver reads exact native properties after an actual rendering frame.

MetersSurface now renders either the retained wide table or actual narrow
definition rows. All eleven public values have exact text/bands; narrow
terms retain constituency/axis (separate p(doom)) and name their own native
meters. One rendered population, live 320/479/480/1280 resizing, refreshed
values, retained navigation focus, zero intents, full-page geometry and axe
are covered in Chromium/WebKit. Six independent label/name/value/breakpoint/
listener/band faults fail 6/6/6/6/4/6 executions, with exact source restoration.

Full current native Garage 252, client 9,814 passes / 449 explicit browser
skips, types/build/boundaries/copy/manifest/topology pass. Unchanged isolated
performance 1 remains RP-333's null-feature guard, not populated AC7. No
host/runtime/wire/catalog/kernel/producer/balance/copy/CI change. Documentation
now also refuses a current-source three-engine/actual zoom/AT claim.

This is local semantic-layout repair, not whole GS3/accessibility/real player
workflow or release acceptance. Complete range after bddfc58e through its
final records needs Claude separately from RP-329 5fbf4cff..bddfc58e,
RP-332 cc62cea8..5fbf4cff, audit fc911784..cc62cea8 and all earlier spans.
RP-331/333/Docker/Firefox/AT/author/body/full-nine-tier 1.0 holds remain.

## Follow-up: RP-333 / RP-334 (separate range after 127eb052)

Original null-feature scenario and budget unchanged. Native Chromium probes
prove the old instrument can pass with no delivery and a final actual 400 ms
task. New test-only supplement verifies pending Lucky, three windowed buffs/
combo (Lucky is not a buff), provisioning cap, owned upgrade, free shelf and
achievements/Fiscal/Meters nav through every measured input. The cash output
is specifically selected, not the earlier static combo output. Public decoder,
coherent runtime samples, mandatory native observation, nonzero visible
rendering, exact terminal cash, real final rendering/task turns and cleanup
provide a bounded measurement, not a new performance budget.

Six faults execute and fail: no delivery, fixed cash, final 400 ms task,
unavailable support sensor, missing building buff, wrong final cash with prior
rendering intact. Old guard survives, as recorded. Healthy diagnostic reports
1,200 actual inputs and region checks, 217 commits, longest task 0, terminal
1.30 K and zero intents. Its reporting-only exit is not counted as a gate pass;
restored native lane exits 0, two passes /22 unselected. Types/client/build/
boundaries pass, 9,814 unit passes /450 explicit browser skips; production build
unchanged. Fixture/type mistakes and local annotation-attachment limitations
are disclosed in log.md. Original body byte comparison and all restoration
hashes executed; no transient fault remains.

RP-334 is locally addressed by the new driver, not by rewriting old evidence.
RP-333/all AC7 remains partial: GS4/pet, later additions/full current release
population, real service/default-player and manual reference profile are not
this fixture population. Full Linux CI/Firefox/AT/SQL holds unchanged. Entire
new range after 127eb052 through final records requires Claude, separately
from exact layout bddfc58e..127eb052 and all preceding spans. No acceptance,
checkbox, lifecycle, archive, push or release promotion.
