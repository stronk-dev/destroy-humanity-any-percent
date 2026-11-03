# Garage Player Surfaces implementation plan

RFC: `rfc/garage-player-surfaces.md` (accepted 2026-09-25; owner decisions at recommended defaults).
Scope of this plan: release-manifest Batch A slices over already-live backends only. The pet (GS4),
cosmetic, Clout, T2-era and ticker slices are out of scope until their producers land.

Constraint for this lane: no kernel-guarded path (`kernel/affecting-paths.json`) is edited. A slice
that would need one is recorded as a blocker in `log.md` instead.

## Current queue — 2026-11-03

Continue accepted construction and corrective review; do not repeat green populations
or expand an uninformative diagnosis into another research campaign.

1. Complete independent corrective review of Offer, purchase/conflict and queued
   transition fixes (`dc1a1adc`, `9b0c0a75`, `e55eb1e9`). Review is running;
   manual-input delivery policy RP-368 remains outside those scopes.
2. Then take the next still-unreviewed accepted scope; distinguish its producer,
   actual player outcome and remaining author/acceptance gates before changing code.
3. Keep RP-348/RP-351 causes open. R-013's report is reviewed, but the temporary
   profile is now absent and cannot be independently revalidated. No stall repair
   or cache/compiler/hosted-causality claim follows.

### Corrective closeout

These are independent **scoped** approvals, not whole-feature acceptance. Exact
base-exclusive/tip-inclusive ranges, reviewer provenance, executed regressions,
environment differences, exclusions and original rejected verdicts live in
[the owning log](log.md), not a second narrative here.

| Scope | Current disposition |
|---|---|
| Meter completeness/layout and contained errors (RP-329/RP-330/RP-332) | Approved; persisted workflow/AT gates separate |
| Source guards (RP-335/RP-337) | Approved; bounded source checks, not arbitrary runtime/dataflow proof |
| Native nav and shared states (RP-338–341) | Approved; native/Linux distinctions and controlled-race limits retained |
| Desk/shared/Fiscal visibility oracles (RP-707–709) | Original plus corrections approved; original CHANGES REQUIRED retained |
| Meters recovery/events (RP-343/RP-344) | Approved; runtime-double fixtures, not real transport recovery |
| Forced context/continuation (RP-082/RP-345–347) | Approved; not physical AT or natural ending reachability |
| Delayed Fiscal path (RP-349) | Approved; invalid/incomplete RP-350/RP-351 controls excluded |
| R-013 stall observation | Research-record closeout approved; cause unresolved and transient sample unavailable |
| Claim pending/removal and provisional contrast (RP-352–355) | Approved; isolated callback/actual axe/valid-credit controls executed |
| Claim connection state (RP-356/RP-357) | Approved through final record edge; regional disclosure and readiness independently discriminate |
| Desk announcement visibility (RP-358/RP-710) and upgrade reasons (RP-711) | Original plus complete correction approved; original REQUEST CHANGES retained, frozen Linux regressions discriminate |
| Four-effect presentation/unknown refusal (RP-359/RP-360) | Scoped approval69c4f972..bbee9805; actual unknown-copy/wrong-scope controls discriminate |
| Historical composed-execution record (RP-361) | Approved bbee9805..1d623dde; source correspondence checked, transient historical execution not independently replayed |
| Performance oracle/supplements/selector (RP-334/RP-484/RP-706) | Approved corrective scopes; full RP-333/AC7 remains open |

Last complete composed execution: `make test-game-ui-composed` passes at `7f540859`
(kernel 0.3.223), with real Postgres and production builds. Its controlled Garden/
Tier2 setup is not natural pacing, Headcount or exact-release proof. RP-132/RP-313/
RP-336 body holds, first-read/renewal policy, shared clock, actual AT, full AC7,
content, remaining integration/review and release-artifact gates remain separate.

## Historical checkpoints and implementation inventory

Older "next"/"pending" statements below describe their recorded coordinate, not
the current queue. Use the current queue above and exact verdicts in the log.

2026-10-15 fresh independent corrective review APPROVES complete RP-318/RP-419,
RP-440/RP-441 and RP-442 ranges. Full hashes and executed native/Linux/Node and real
Postgres/WebSocket journeys are in the owning log. Only these corrective scopes leave the
pending-review queue. Candidate copy still needs owner adoption; shared-clock repair,
manual AT, natural Garden growth/renewal and full Garage acceptance remain separate.

2026-10-15 fresh independent corrective review also APPROVES complete RP-407 host-disposal
range0fd146e0..b7f351cb. The owning log records full hashes, six mounted late-completion races,
actual native/Linux consumers and focused real-service driver execution. This closes that
correction's review, not first-read/renewal, full Garage or the separate new Pet focus finding.

2026-10-15 independent corrective review APPROVES the complete single-commit ranges
`04acc775^..04acc775` (RP-447), `c0809cab^..c0809cab` (RP-448),
`2db7b03d^..2db7b03d` (RP-498), `e8353eda^..e8353eda` (RP-499/RP-500)
and `aaf78306^..aaf78306` (RP-501/real-boundary queued-Claim regression).
The owning log records fresh Codex provenance, actual native/SQL journey and the initial
Firefox harness failure. Only these corrections leave the pending-review queue; original
feature acceptance, other review debt and held contracts remain separate.
Parent verification also executes the same two affected files in declared Linux Firefox:
528 pass/five performance exclusions. This supplements execution coverage, not the designated
review scope, native startup repair, intermittent-failure closure or full CI acceptance.

2026-10-15 GS0.2 correction (RP-667): the legacy intent response now uses the
generated transport's existing fatal UTF-8 JSON decoder. Twelve old-source cases
admit repaired applied/rejected replies or typed409 conflicts; those now refuse
before outcome mapping. Valid Unicode/BOM/JSON escapes, single request, credentials
and actual body-read/network error propagation remain. Receipt schemas, media
policy, retries and held generated-dispatch decisions are unchanged. The owning
log records real-byte tests, initial fixture assumptions and final checks. Fresh
independent corrective review APPROVES all ten paths of b3d1fb40..11478e9d;
its native Firefox collection failure remains explicit, separate from the parent's
successful Linux three-engine run. This is not whole GS0.2/Garage acceptance.

2026-10-14 bounded correction (RP-619): wide-layout native meters now reference their visible
constituency/axis headers or p(doom) heading. Mounted regressions identify all eleven missing
names on the old component; final affected Chromium/WebKit846, types/build/boundaries pass.
The owning log records fresh independent Codex APPROVED for the complete
`8f698279..33815f04` corrective range under Marco's explicit delegation.
Actual assistive-technology testing and full Garage acceptance remain separate.

Earlier checkpoint (2026-10-15): at `6201bef8` (kernel0.3.216), the complete
`make test-game-ui-composed` target passes after the newer replay/frozen-row corrections.
Actual disposable Postgres reports migration85 applied. Main, Cosmetic/Reputation, Clout,
payment negatives, Garden harvest and Tier-2 candidate journeys execute sequentially with
fresh builds, exact loaded assets/Worker, real sockets and committed-reply-loss recovery.
Both driver populations require and execute8/8 refresh cases, ten persisted parents and
public-SDK/TS replay checks; captured test headers have no skips or failures.
The owning log retains setup/architecture limits and previous-source evidence. This refresh
does not close independent review, natural pacing, Headcount, accessibility, public Arcade/
Typer integration, content/mint or exact-release proof. Continue accepted construction.

Preceding checkpoint (2026-10-12): the production-built real-service main journey now exercises
GS0.8's fresh Claim behind an actual purchase and held authoritative read, with exact consecutive
Postgres payloads/receipts/state and the observed building buff in its successor snapshot.
RP-501's stale Fiscal refusal selector is independently approved; both real refusal/recovery arms pass.
The latest log owns failed/final runs, boundaries and the consolidated Claim review range.
No full composed aggregate, hosted CI, actual-AT, all-effect or Garage/release promotion follows.
Continue accepted integration/acceptance work and consolidate outstanding designated review.

Preceding checkpoint (2026-10-12): RP-499's mounted Claim feedback is locally repaired under
GS0.6/GS5. The existing Desk status region owns pending and outcomes; Claim describes that
visible pending text without adding a child live region. RP-500 retains an explicit separator
between unchanged messages. The latest log owns old-source failures, final verification and
the now independently approved Claim review range. No actual-AT, whole Garage, CI or release promotion follows;
continue the owning plan's remaining integrated acceptance and consolidate outstanding reviews.

Preceding checkpoint (2026-10-12): RP-498's Claim guard now identifies only Claim's own flight,
not another purchase or a background read. Native input reaches the existing GS0.8 queue,
waits for the authoritative Company revision and stays single-flight; failed-read/unmount
controls prevent submission. The latest log owns verification and the independently approved range.
No new retry, Manual-input policy, server mechanics or whole-Garage acceptance. Its RP-499
follow-up is recorded in the newer checkpoint rather than reopening this input repair.

Preceding checkpoint (2026-10-08): RP-448 extends the existing applied-read handoff to Founder
actions under GS0.2/GS1-A6. A shared pre-harvest read must be refreshed against the actual
snapshot Founder revision before pending clears, including across Company changes. Failed
fresh reads preserve offline handling and never retry the intent. The owning log records
old-host failures, current-read/Company controls, final verification and independent approval.

Preceding checkpoint (2026-10-08): RP-447's ordinary Company purchase/read handoff is repaired
under GS0.2. A reused pre-commit read requires a post-response refresh before pending clears;
same-Founder/run comparison keeps successor-Company genesis separate. Deterministic native
purchase regressions fail the old host in Chromium/WebKit. Final affected and real-service
results, exact range and independent corrective approval belong in the owning log. No input retry,
new gameplay/schema/clock/limiter policy or whole Garage acceptance follows.

Preceding checkpoint (2026-10-08): RP-442's GS1/OD-3 ready-quarter navigation suffix is
locally implemented using the existing Fiscal phase and host estimate. No modal,
announcement, request or timer/payout policy changed. The harvested-event acknowledgement
remains separate; new copy is visibly pending owner adoption. Exact-edge native regressions
fail the old UI; affected browser, type/build/copy checks and the production-built
real-service target pass within their recorded populations. See the latest owning log for
commands, intermediate failures and the independently approved corrective range. This is not
whole Garage acceptance or a repair of the separate shared-clock defect.

Preceding checkpoint (2026-10-08): RP-440/RP-441 native pointer/read defects are locally corrected.
Care's command guard now identifies its own command, not every background read or
different-kind intent. Shared busy feedback remains in a reserved line so it cannot move
the button between pointerdown and click. The existing single-flight queue retains the
click and binds the refreshed Founder revision; failed-read/unmount and same-kind duplicate controls pass.
Affected-suite and real-service verification pass and belong in the latest owning log; no feature
acceptance or clock repair follows from this bounded correction.

Preceding checkpoint (2026-10-08): RP-318's missing pet status decoder/announcement is locally
implemented under GS0.3/GS4. Founder events announce only for an adopted pet while Pet Care is
mounted; replays/off-surface events do not repeat or queue speech, and values remain snapshot-owned.
The existing real-service care journey now exercises an actual low → high status transition,
stored event and socket delivery. See the latest owning log for final checks and exact review
range. New companion-tone copy remains pending owner adoption; RP-132/RP-365 and full acceptance
are separate obligations.

Preceding checkpoint (2026-10-08): OD-15 upgrade ineligibility reasons now reach the Desk from
the authoritative projection, with visible text and button descriptions. The production-built
real-service journey exercises window/requirement rows, and the complete composed target passes.
The latest owning log records affected checks and the pending exact-range designated review;
three new strings remain visibly pending owner copy. No Garage acceptance or release promotion.

Preceding checkpoint (2026-10-08): the locked-Pitch witness now handles post-response offer
preemption through visible Decline/navigation and fresh DOM Start, but still requires an exact
409 and a visible mounted-launcher notice to pass. Native observer/host controls24 and full
three-journey root23009 pass with exit0. That real run has no offer decline, so post-response
re-consent remains controlled-boundary evidence, not a uniquely traced70861 fix. RP-389 remains
open on attribution/reliability; designated review remains. Next: consolidated review and
accepted integration work, preserving recurrence diagnostics rather than rerunning for a seed.

Preceding checkpoint: all three expanded composed populations execute and pass
locally56007, including the unminted Clout purchase/SQL variant. Earlier locked-Pitch notice
failure RP-389 remains open: a controlled native host test proves that an Exit offer correctly
unmounts the launcher before or after the refusal, while an unconditional notice wait assumes
it stays mounted. Failure-only public lifecycle diagnostics now retain that context. Affected
Chromium/WebKit files184 and types pass; next repair the oracle's lifecycle handling without
dropping the mounted-launcher notice requirement. No reliability/hosted/full-Garage acceptance.

Preceding checkpoint: the main composed player journey now uses the production client
build, with exact browser-loaded HTML/JS/CSS/worker bytes and a started bundled worker verified.
Its real Postgres/socket/gameplay population passes; the aggregate then fails at Cosmetic equip
(RP-385), so whole-lane green is not claimed. Test-only; no gameplay/CI budget changes. Next:
consolidate Garage review/remaining acceptance and diagnose the separate equip failure in its
owning lane. Latest logs own the evidence; no full Garage/CI/1.0 acceptance.

Preceding checkpoint: purchases and Gate/Incorporate/Wind Down retain native pending focus under
GS0.8. Removed transitions recover focus within their region without stealing newer choices;
queued Wind Down now binds both refreshed revisions (RP-380). Affected functional browser,
type/build/boundary and final real-service journeys pass; designated review remains.
RP-381's hosted WebKit error-constructor assertion is locally corrected: Node27/native54 and
types pass without changing product handling. Next: consolidate Garage review/remaining
acceptance. Manual's pending-input choice is unruled; Clout and historical kernel CI remain red.
The latest `log.md` owns results; do not infer full Garage/CI/1.0 acceptance.

Preceding checkpoint: ordinary whole composed68777 passes required Postgres, Fiscal refusals,
both terminal states/continuation/recovery and Cosmetics. RP-378's earlier Fiscal response
timeout remains open; retained input/request diagnostics discriminate a controlled blocked
Enter from a server response failure. Do not rerun until green or guess a gameplay/clock fix.
Next: use the trace on recurrence or construct a bounded reproduction; other accepted work
continues while this diagnosis and designated review remain open. `log.md` owns details.

Preceding checkpoint: RP-371 stale-state controls are locally corrected under GS0.5. Focused
old-behavior regressions fail; corrected native host checks, affected surfaces, type/build and
the real composed journey pass. Sustained native input produces no stale submissions but still
hits rate limits and misses activations: RP-368 is OPEN, not repaired by disabling controls.
Next: resolve the action/request-throughput contract without guessing a larger limit or replaying
commands. The latest `log.md` owns commands, results and review range. Designated review and full
Garage acceptance remain; RP-365/D-024, GS3-A4/RP-363 and the original RP-364 timeout remain open.
RP-319's Offer pending-focus remainder is now locally corrected under GS0.8; Manual and other
Desk pending controls remain separate work. Manual click retention is an unanswered owner choice
in the decision queue, not permission to change GS0.8's pending-drop rule.

Historical checkpoint before the GS2 supplement: RP-362/363 fired a real initial-meter authority conflict;
GS3-A4 asks fresh Standing50, actual bootstrap assembles90/high via published
Notoriety0 reseed. Grievance0/doom50 match. Candidate fails before DOM/five
planned faults; retained replayable initial-meter-dom-probe.patch, original
registered driver restored exactly. No assertion changed to90 or formula/body
edit. Author reconciles fresh-Founder/run/Notoriety/seed contract before repair.
Subsequent existing root9ed7c3→3a5178 is RED: main fully passes, Cosmetic Buy
response times out at420 (RP-364), unlike earlier two whole passes on identical
sources. No cause/reliability claim; owning Cosmetic lane must first predeclare
boundary diagnostics. Whole1d623dde-exclusive through records needs Claude,
all earlier ranges independent. Next: RP-364 diagnosis, then separate GS2-A4
purchase/announcement/row proof. No checkbox/CI/budget/product/owner/body/
balance/lifecycle/release promotion; full nine-tier1.0 and all remaining holds.

Preceding bounded work: RP-361 real-service composed evidence under f46298e7.
Unchanged full root target passes twice: current host/server/Postgres/socket/
DOM, both endings/continuation/recovery, observed Production/Click buff claims,
Fiscal/Pitch and compiled cosmetic/adoption/care/reload. Mounted Claim no-op
fails specifically with no emitted intent; exact source restored and entire
root rerun green. Post-run source/driver/service hashes unchanged, target ports
absent, named test DB other sessions0. No committed product/test change.
Whole bbee9805-exclusive through records needs Claude, older ranges independent.
Host-native route is usable; Docker image/tmp overlay0/100% still holds
disk-dependent Linux/image/SQL-container/release work, not this route.
Next: inspect/predeclare real initial-meter DOM and first-purchase→earned
announcement→row-refresh gaps (RP-331). No Lucky/all-effect/production-epoch/
release-artifact/Firefox/AT/400%/whole Garage or full-nine-tier1.0 promotion.
RP-132/313/318 author-body and all owner/review/platform holds remain.

Preceding bounded work: RP-359 four-effect/refusal proof, test-only undereb6e7e8d/
7bf9d286/a531bdeb.20 native executions at320/1280 across four registered and
one unknown pending row; actual public parser, native Tab/Enter-Space, exact
Company2/opportunity-only requests, three distinct live buffs/times, combo cap,
Lucky curtain and unknown offer/description withholding+diagnostic. Twelve
valid source faults discriminate/restored; RP-360 initial cap-space instrument
mismatch corrected/excluded. Final native818/four isolated-performance skips
plus performance2/client9816/711 browser skips/types0/unchanged build/copy/
boundaries/topology/no-payment pass. No product defect or product bytes changed.
Whole69c4f972-exclusive through final records needs Claude; earlier ranges
independent. Next: separately predeclare unchanged host-native composed root
target after verified existing tmpfs DB7.7G free/no sessions and host180Gi free;
Docker image/build overlay still0/100%, not blanket proof this route is blocked.
No reset of operator data, Linux/release-artifact substitution, RP-313 author
body/copy/CI/cleanup/startup ruling or lifecycle promotion. Real-service rerun,
SQL/Firefox/AT/400%/all-era/state/platform/privacy/numeric/endings/owner/review
and full-nine-tier1.0 holds stay explicit until actually proven.

Preceding bounded work: RP-358 announcement authority/replay proof, test-only under
639bf480/90b3cc7b with supplemental value controls70bbe82c. Twelve declarations/
24 native executions cover six paths/two widths/two engines: exact registered
spawn/buff text, off-Desk withholding, ID/cursor replay, native focus and event
versus snapshot value binding. Nine actual source faults fail/restored; full
native798/four isolated-performance skips plus performance2/client9816/701
browser skips/types0/unchanged build/copy/boundaries/topology pass. No product
defect fired; all production bytes unchanged. Entire10e2d656-exclusive through
final records needs Claude, all older spans independent. Next safe accepted
work: inspect and separately predeclare GS5 four-effect presentation/unknown-
row fail-closed proof; current native projection populations cover Lucky and
Production, not every registered effect. No author RP-313/public-wire repair,
real acquisition/payout/expiry/SQL/Firefox/AT/400%/all-state/era/full1.0 or
lifecycle promotion; capacity/platform/numeric/review/owner holds remain.

Preceding bounded work: RP-356 connection-state correction40934e15 after separately
predeclared52becf28/8e834bd5 typed red40. Five paths ×320/1280 ×independent
availability/disclosure ×two engines =40 native cases. Present projected values
remain committed, stale explanation visible and Claim disabled through unready
snapshot reception; recovered signal clears reason/enables, native next Company2
Claim uses new opportunity ID, not Founder8. Nine complete faults discriminate/
restore; full native774/four isolated-performance skips plus performance2,
client9816/689 browser-only skips/types0/build214/copy/boundaries/topology pass.
RP-357 initial bad test copy key corrected/excluded. Whole85b8e826-exclusive
new span through records requires Claude independently of prior spans. Next
accepted work: predeclare GS5/GS0.3 on/off-Desk spawn/buff announcement and replay
proof; source guards present, no defect inferred yet. Real service/SQL/Firefox/
AT/400%/all-state/era/body/platform/numeric/full1.0/review/owner holds remain.

Preceding bounded work: RP-352/353 Claim pending/removal repair5064245f after
typed-clean test-first red09b665; separate RP-354 C9 provisional two-token
tuning under48f6d63e. Twenty declarations/40 Chromium-WebKit executions at
320/1280 with native Enter/Space: held refusal/applied/read, exact Company
revision/receipt, duplicate guard, surviving/newer-selection focus and removed
Claim handoff. Fourteen valid source faults discriminate/restored; RP-355 invalid
Decimal control discarded and replaced by a valid credit-misbinding control.
Full native734/four isolated-performance skips plus performance two; client9816/
669 browser-only skips/types0/build214/copy/boundaries/topology/no-payment pass.
Whole f5be35f6-exclusive new span through final records needs Claude; all prior
spans independently owed. No boxes/lifecycle/whole-GS5/full1.0 promotion.
Its next GS5-inherited GS0.5 native connection-state proof and RP-356 correction
now execute above; no original red coordinate is retroactively made green.
RP-313 author wire-body, Docker capacity, real service/SQL/Firefox/AT/400%/
populated performance/full-nine-tier/owner release holds remain.

Preceding bounded work: RP-349 native Fiscal pending/receipt focus under aa296eba,
test-only. Three actions ×320/1280 ×Enter/Space ×Chromium/WebKit:24 executions;
held refusal/applied/read, duplicate guard, native Settings/newer focus and native
return with new Founder revision. Eight valid source faults discriminate and
restore; contaminated RP-350 probe discarded, two incomplete runner probes at
RP-351 retained and routed to R-013. Full browser694/four isolated-performance
skips plus performance two; client9816/649 browser skips/types/unchanged build/
copy/boundaries/topology/no-payment pass. No product bytes changed or whole-GS1
acceptance. Entire c24b8f6d-exclusive new range requires Claude, all prior spans
independent. R-013's separately declared first diagnostic wave completes without
reproducing either stall or attributing its phase; RP-351 remains OPEN. Its next
RP-352 native held Opportunity Claim proof under GS0.6/0.8/GS5 now runs above,
without resolving author-owned RP-313 compatibility intent or guessing mechanics.
Linux/SQL capacity and all full-nine-tier 1.0/owner/review holds remain.

RP-352's original test-first predeclaration and red remain in the append-only log;
the separate repair/tuning and current full proof do not retroactively make that
baseline green. Current public wire inputs remain preserved. Every earlier
review obligation remains independent.

Preceding cross-lane checkpoint: RP-348 observation work moved to accepted Garden
SG1/AC1 undera9fadedd. Complete child diagnostics and real startup/hang controls
locally pass; historical failure cause remains open. No admission/deadline/CI or
product change. Final client9816/637 skips, native raw74/six Node-only skips,
types/unchanged build/boundaries/topology/performance pass. New4b44a563-exclusive
range and every earlier Garage range independently need Claude. Next accepted
Garage work: inspect then predeclare native Fiscal pending/receipt-focus audit,
retaining existing copy/revision/refusal states. No startup Retry authority.

Historical implementation checkpoint: RP-346 native tab stop / RP-347 pending focus repair.
2e5ba552 → red a3dab732 → separatefd56b866 → retry diagnosis → separate
38d8d650. Current read control explicit tabindex0/pending aria-disabled;
existing guard actually suppresses duplicated native activation. Standard96
across four labels/three eras + stronger scripted56 cases pass. Eleven real
faults discriminate/restored; traversal bounds unchanged. Full native670/four
isolated-performance skips, separate performance two; types/build/copy/
boundaries/static topology/no-payment pass. First full client RED at Garden
child entry; unchanged later full run9,814/637 skips passes. RP-348 retains
that failure/unknown cause, not hidden by a repeat. Next safe work: separately
predeclare child-control startup/scanner observability and budget discrimination,
read its accepted Garden authority before any test edit; no ceiling/CI change
inferred. Full0e7910ed-exclusive through repair/records needs Claude, prior
0e548992..0e7910ed and earlier spans independent. Fresh Docker0 available/100%
still blocks Linux/SQL. No cleanup/startup ruling, Retry, auth/copy/read-policy
change, boxes/body/status/archive/push/release/full1.0 promotion.

Preceding bounded work: RP-345 / RP-082 exact-next native continuation repair.
Test scope19cc5609 → red94c88b72 → separate accepted scope160eeb0d. Refuse
wrong Founder before binding, retain exact-next/failure behavior; removed focused
continuation exposes Desk heading without taking surviving newer focus. Same56
native cases pass; seven real faults discriminate/restored. Combined Garage/
Game UI574 passes/four isolated-performance skips, separate performance two;
client9,814/589 explicit skips, types/build/copy/boundaries/static topology/
no-payment pass. Entire0e548992-exclusive through containing repair/records
requires Claude; all preceding spans independent. Next accepted work: predeclare
standard-terminal native continuation/refusal/recovery population. New decoded
helper currently emits scripted_first only; old standard click-only test is
not native focus proof. Real composed/HTTP/auth/SQL/Firefox/AT/full shared-state/
full1.0 remain open. No new startup Retry/copy/credential/navigation-cancellation
policy, cleanup/ruling/boxes/lifecycle/archive/push/release inferred.

Preceding bounded work: RP-082 forced Offer/Run-End context focus. Predeclare
3451fdb0 → actual 52-failure/8-control red checkpoint 73a5ccce → separate
accepted product scope ac6a5788. Existing headings now focus after render;
newer choices cancel pending focus, cursor ordering/replay preserved. Same
60 native cases pass; eight real faults discriminate and restore. Combined
Garage/Game UI 518 passes / 4 isolated-performance skips; separate performance
two, client 9,814 / 561 explicit skips, types/build/copy/boundaries/no-payment/
static topology pass. Full fbea2158-exclusive range through repair/records
needs Claude; 9f485993..fbea2158 and earlier exact spans independently owed.
Next accepted work: predeclare native continuation context/focus truth audit,
including exact-next-Company and failed-read controls, before any repair.
Current .click/surface-only test does not prove keyboard/focus. Not yet a new
executed continuation defect. Capacity/full Linux/SQL/Firefox/AT/body/privacy/
platform/numeric/full-nine-tier 1.0/review holds remain; cleanup/startup draft
unanswered, no deletion/Retry/checkbox/lifecycle/push authority inferred.

Preceding bounded work: RP-344 GS3 event/badge/value separation, test-only under
96439cfe. Existing combined oracle survives actual forbidden global-announcement
fault; sixteen new native executions catch it and eight other real source
faults. Up/down p(doom)/users Standing at 320/1280: exact polite text/badge/
visit/old-cursor/unknown-ID controls; all eleven values change only on explicit
newer snapshot. Source restored exactly. Final Garage 414 / performance two,
client 9,814 / 531 explicit skips, types/unchanged build/copy/boundaries/static
topology pass. Whole 9f485993-exclusive range through this proof/record commit
requires Claude; prior repair 4b2fd904..9f485993 and earlier spans remain owed.
Next accepted work: predeclare native GS0.4/0.6 lifecycle forced-change focus
truth audit from mounted Garage surfaces. Current source is a candidate only,
not yet an executed defect; do not treat nav state alone as focus evidence.
Full Linux/SQL capacity freshly still zero free/100%; cleanup unanswered,
no deletion or container population. Startup draft, Firefox/AT/body/privacy/
platform/numeric/full-nine-tier 1.0/review obligations remain; no boxes/lifecycle.

Preceding bounded work: RP-343 Meters connection disclosure. Test scope f3b2f421
reproduced 20 missing-note failures at red checkpoint 0a191ebd; separate
product scope bfd7b4ff extends the existing host stale note to Meters only.
The same 20 native cases pass; six actual faults discriminate and restore.
Eleven committed values/bands remain exact until a newer snapshot, including
through recovery. Final Garage 398 / performance two, client 9,814 / 523
explicit skips, types/build/copy/boundaries/static topology pass. No new copy,
clock, value math, transport, balance or CI behavior. Whole 4b2fd904-exclusive
span through this repair/record commit needs Claude; all earlier spans owed.
Next accepted work: inspect GS3 on-surface announcements/off-surface badge
evidence and predeclare any missing discrimination. Do not invent startup
Retry from the draft. Full Linux/SQL capacity and Firefox/AT/body/privacy/
platform/numeric/full-nine-tier 1.0/review obligations remain; no cleanup
permission, deletion, acceptance checkbox or lifecycle promotion.

Preceding bounded work: RP-342 completed-first-read diagnosis, predeclaredced3b5c7.
Actual runtime/Response.json/decoder/native host28 cases; all24 failures retain
visible loading after completion, four healthy controls render Desk. Existing
visibilitychange rereads; valid same-Founder replies recover, repeated401 does
not. Exact metadata retained, four faults discriminate/restored. Final Garage
378 /performance two, client9,814 /513 explicit skips/types/unchanged build/
copy/boundaries/static topology pass. No product fix: draft first-read recovery
RFC/D-023 resolve precedence/Retry/copy before implementation; not renewal.
Whole a551d3c2-exclusive span through records needs Claude, earlier spans owed.
NEXT safe accepted work: separately predeclare GS3 Meters reconnect/not-ready
last-value/stale-disclosure proof under GS0.5, failed-first if required. No new
owner mechanics. Linux/SQL capacity and full-nine-tier1.0/body/privacy/AT/
Firefox/review holds remain; no cleanup permission or lifecycle promotion.

Preceding bounded work: test-only remaining shared-effect focus consumers,
predeclaredfb2a9488. Fiscal/Meters/Reputation ×320/1280 ×retained/removed
feature facts:24 native passes. Each omitted predicate fails its eight cases;
no-op focus/removed Desk tabindex each fail24. Exact host restored, no
production byte changes. Final full Garage350 /performance two, client9,814
/499 explicit skips/types/unchanged build/copy/boundaries/static topology pass.
Whole82614848-exclusive span through records needs Claude independently of
0fe2f57c..82614848 and earlier ranges. No checkbox/lifecycle/archive/push.
Next: RP-342 failed-first-read truth audit. Source candidate only: refresh
catches offline but credentialled no-snapshot Desk still renders loading.
Test actual failure/malformed/healthy recovery before choosing any new retry
UI; do not infer a missing recovery contract. Linux/SQL capacity hold and all
full-nine-tier1.0/review/body/privacy/platform/Firefox/AT holds remain.

Preceding bounded work: RP-339–341 accepted GS0.5/0.6 host repair after actual
red test checkpoint6d36a9d3 and separate product predeclaration88153822.
Held Desk read now has heading/status, Trophy stale values are disclosed,
forced arm-loss Desk focus is cancellable by newer selection/snapshot.
Forty selected native executions pass; nine real faults discriminate and
restore exactly. Final full Garage326 /performance two, client9,814 /487
explicit skips/types/build/copy/source/static topology pass. Compilation
collision in first attempt disclosed and fixed before behavioral evidence.
Entire0fe2f57c-exclusive span through this repair/record commit needs Claude,
including both predeclarations and failed-test checkpoint; earlier exact
ranges remain owed. No checkbox/lifecycle/archive/push/release promotion.
Next: predeclare remaining forced-context focus populations (Fiscal/Meters/
Reputation) under GS0.5/0.6; first-read failure/retry requires a separate truth
audit, not invented UI. Real GS2-A4/Linux/SQL remain capacity-held, cleanup
permission unanswered, no deletion/full-disk rerun. Firefox/manual AT/zoom/
author/body/GS4/full AC7/default-player/privacy/platform/numeric/full-nine-
tier1.0 obligations remain live.

Preceding bounded work: GS2-A1/A5 visible Trophy Case fixtures and RP-338 native
nav repair. Test-only6c578197..ffabda04 records actual WebKit Tab failure;
separate e30b5e00 predeclares explicit tabindex0 on ten existing nav buttons,
unchanged order/handlers/copy. Twelve public fixture cases×two engines cover
populated/empty/error at320/1280 with Enter/Space, exact visible text and
unmasked geometry/native traversal/axe. Ten actual faults discriminate;
sources restore exactly. Full native Garage286 /isolated performance two,
types/client9,814 /467 explicit browser skips/build/copy/source/static topology
pass. UI JS changes as expected; worker/CSS unchanged. Driver annotation
error and disconfirmed axe hypothesis disclosed, no assertion weakened.
Complete new span after85fdcfdf through records needs Claude independently
of source guard3609e776..85fdcfdf and every earlier span. No acceptance box/
lifecycle/archive/push/mint/release change. Next accepted work: GS2 shared
loading/reconnect/null-arm context-change fixtures under GS0.5, predeclared
and discriminating; persisted GS2-A4 waits on capacity. Cleanup permission
still unanswered, no deletion/full-disk run. Firefox/all-state/actual zoom/
AT/author/body/GS4/full AC7/default-player/privacy/platform/numeric/full-nine-
tier 1.0 holds remain. No shortened-preview substitution.

Preceding bounded work: RP-337 GS2-A3 verification-only source guard under
305b8e8b/a32b8358. Old gate survives real binding/label faults; existing-tool
extension checks actual Svelte AST and canonical source achievement copy/era
variants, not generated output. Seven component/four copy negatives and five
actual semantic faults fail; production/copy restored exactly. Healthy types,
client9,814 /455 explicit browser skips/unchanged build/copy/boundaries/static
topology pass; focused native GS2-A1 two passes /260 explicitly unselected,
chained performance two /22 unselected. Not full native/all-engine/live SQL/
arbitrary computed-dataflow/player acceptance. Complete new span after
3609e776 through records needs Claude independently of all earlier spans.
No box/lifecycle/archive/mint/push/deploy/release/preview promotion.
Next: capacity restoration only if human authorizes exact scoped cleanup;
otherwise predeclare GS2-A1 actual non-color state-text fault discrimination
and GS2-A5 fixture/error/empty narrow-readability populations, preserving
manual AT/real-service/full Linux gates. Docker authority still unanswered;
no deletion/full-disk run. All author/body/GS4/full AC7/default-player/privacy/
platform/numeric/full-nine-tier 1.0 holds stay open.

Preceding bounded work: test-only GS1-A3 exact rendered Fiscal edge fixtures,
predeclared 2ff59fbe/f1914ace. Four decoded public component fixtures at
early-1/early/guaranteed-1/guaranteed; actual visible exact copy/risk/countdown,
native readiness/curtain/callbacks checked. Eight source faults discriminate,
restored exactly; setup type narrowing mistakes disclosed. Final full native
Garage Chromium/WebKit262 passes, isolated performance two passes, client
9,814 passes /455 browser skips, types/unchanged build/boundary/topology pass.
No product/clock/copy/balance/CI change or live service/pacing/Firefox/AT claim.
New RP-336 props/body finding remains author-owned. Complete new range after
57efdc39 through final records needs Claude separately from Desk
54c8c8da..57efdc39, source guard d3ce0f76..54c8c8da and every earlier span.
Fresh read-only Docker still 0 free/100%; no deletion/container rerun. Cache
ownership inspection ongoing, scoped cleanup authority requested but not
assumed. If authorized, recover measured capacity then re-run declared cold
Linux/SQL gates; otherwise accepted GS2-A3 score-vs-Clout source guard is
available for predeclared work. RP-331/author/body/GS4/full AC7/AT/default-
player/privacy/platform/numeric/review/full-nine-tier 1.0 holds stay open.
No checkbox/lifecycle/archive/mint/push/release/preview promotion.

Preceding bounded work: test-only GS6-A3 paired whole-Desk 320 px measurement,
predeclared 757f1ca2. Exact absent/present provision/cap/owned census,
visible unchanged regions, focus/no-intents and whole-page/descendant/own-
scroll/no-masking measurements; six real faults fail both engines and restore
exactly. Native widths document/main 320/320, Desk/chrome 284/284, extents
0–320, measured nodes 90→93. Not a historical executable, accessibility's
647 px closure, actual zoom/AT/Firefox/real-service/later Desk proof.
Final full Garage Chromium/WebKit 254 passes, isolated performance two passes,
types/client 9,814 passes /451 browser skips/unchanged build/boundaries/
topology pass. No production/CI byte moved; selector/reporting mistakes
disclosed. Complete new span after 54c8c8da through final records requires
Claude independently of d3ce0f76..54c8c8da and all preceding exact spans.
Next accepted native work: predeclare exact rendered Fiscal phase edges
GS1-A3 with explicit fixture time, not synthetic live-clock proof. RP-236/
RP-331/full Linux/review/author/body/GS4/full AC7/AT/default-player/privacy/
platform/numeric/full-nine-tier 1.0 gates remain; no lifecycle promotion.

Preceding bounded work: AC4 current-source discrimination and RP-335 GS6-A2
verification-only repair, predeclared 2a5deeee/da2bc019. Four actual component
policy faults fail; the old boundary's fake cosmetic button survivor is
recorded before correction. Production Go AST/decoder now owns the command
set; Svelte AST checks explicit envelopes/current wrappers, not arbitrary
whole-program/runtime/eligibility proof. Five real faults and ten Go/eleven
Svelte negatives fail; sources restored exactly. Final boundary/types/client
9,814 passes /450 explicit browser skips/build/isolation/payment/topology/
isolated performance pass. Cold cosmetic parser/transition passes; two SQL
tests explicitly skip. Instrument wrapper-binding mistakes are disclosed.
No production/API/copy/balance/CI/Make byte moved. Exact range
d3ce0f76..54c8c8da needs Claude separately from performance
127eb052..d3ce0f76, layout bddfc58e..127eb052 and all earlier exact spans.
Next accepted work at that checkpoint: predeclare GS6-A3 whole-Desk before/after
comparison, not Firefox/actual 400% zoom/AT substitution. RP-331 SQL/DOM
proofs await repaired/rechecked Docker capacity; no run/cleanup here. GS4/
RP-333 full performance, author/body/platform/privacy/numeric/default-player/
review/full-nine-tier 1.0 gates remain. No box/lifecycle/archive/push/release
promotion; earlier slices retain their independent proof/review boundaries.

GS0.6 corrective review, 2026-11-02 (not a new acceptance checkbox): RP-326's
captured-origin/own-panel repair is independently APPROVED for the complete
`e950216a..6d700838` span, separately from earlier browser/Fiscal ranges.
The reviewer executes the original24 ownership regressions against actual old
and corrected snapshots:24 fail, then24 pass. RP-487's later one-stable-region
repair is separately APPROVED for complete `b3c8f178..237c2f17`; its descendants
run in the current438/438 Linux three-engine Fiscal/care/feedback population.
Types/build/client boundaries pass. The owning log records exact hashes,
provenance, retained versus newly executed evidence and exclusions. RP-336's
author-owned props discrepancy, RP-132's wire reconciliation, all-surface/
HTTP-failure/queued-origin acceptance, actual AT, full Garage/CI and release
remain separate. No lifecycle promotion or new specification adoption.

- [x] GS0.2 runtime: `intent()` returns the typed outcome; non-2xx throws `GameUIRequestError`;
  `act()` scopes `expected_revision` (Company vs Founder); rejected intents render their reason.
  Original `66ae01f9^..66ae01f9`, all fourteen changed paths, now has a designated Codex
  CHANGES REQUIRED disposition for RP-147/317/322 in the 2026-10-13 owning log.
  This is not approval of Codex's later corrections or full GS0.2 acceptance.
- [x] GS0.1 projection: Game UI snapshot v4 (`features` arms + `generators[].provision_cap` +
  feature facts), registered schema, generated types, client parser; live sync requires v4.
  Original `5165abcf^..5165abcf`, all21 paths, now has designated Codex CHANGES REQUIRED
  (2026-10-13): RP-329 partial-meter admission, RP-532 hoard overflow and RP-533 kernel
  derivation/representation gap. RP-329's complete `5fbf4cff..bddfc58e` correction
  and RP-532's complete `4eaed0c9..61b145af` guard are independently corrective-approved;
  exact provenance and fresh executed checks are in the 2026-11-03 owning log.
  The three SQL tests explicitly skipped in those native reviews. A separate
  test-only supplement now runs all three against declared Postgres, including
  eleven Fiscal cases (seven persisted hoard boundaries). Its private guard-removal
  overlay fails exactly three refusals while preserving four controls. The new
  complete five-path `1667a10f..dff167a6` supplement is independently corrective-approved
  with fresh real-Postgres execution and the three-failure/four-control overlay result.
  It is not a larger-value representation or browser/HTTP timing proof. RP-533's kernel-owner closure and
  full GS0.1/feature/CI/release acceptance stay open.
- [x] GS3 Meters arm + surface.
  Original surface batch `acf43132^..acf43132`, all22 paths, has designated Codex
  CHANGES REQUIRED (2026-10-13): RP-323/324/330/332 and catalog correspondence RP-534.
  This covers the GS1/GS2/GS3/GS6/GS7 changes below, not approval of subsequent Codex
  corrections. RP-330's complete `bddfc58e..127eb052` semantic layout correction
  and RP-332's separate complete `cc62cea8..5fbf4cff` contained presentation errors
  are now independently corrective-approved. Current three-engine Meters51/51,
  types/build/boundary checks pass; original feature/actual zoom/AT gates remain.
  The owning log retains exact provenance, historical evidence limits and ranges.
- [x] GS2 Achievements arm + surface.
- [x] GS1 Fiscal arm + surface.
  RP-323/324/325 pending/readiness/row-reason/focus correction is independently
  APPROVED for complete `b64a91af..4dcd9969`, including tests and final records.
  Fresh private historical checks fail8/8 on the old source and pass12/12 on
  the corrected tip, including alternative parsed cap reasons and multiple
  unequal-distance survivors. Current Fiscal/care/feedback descendants pass
  438/438 across Linux Chromium/Firefox/WebKit; types/build/boundaries pass.
  RP-326 outcome ownership and RP-487 stable feedback regions have their own
  approvals above. Exact evidence/limits are in the 2026-11-02 owning log.
  Author/wire, actual AT, full-feature/CI/release gates remain; this existing
  checkbox records implementation only, never archival acceptance.
- [x] GS7 minigames availability arm wired into the existing Pitch surface/nav.
- [x] GS5 active-play arm (`features.opportunity`, kernel export `ProjectActiveCombo`) + Desk opportunity region + composed claim witness (`cdb8fe61`, `f32f6175`).
  Original producer `cdb8fe61^..cdb8fe61`, all twelve paths: designated Codex
  CHANGES REQUIRED, RP-313 wire/clone body reconciliation. Green primitive
  tests do not resolve this. Original consumer `f32f6175^..f32f6175`, all
  fourteen paths, is separately CHANGES REQUIRED for RP-314/315/316. Narrow
  cap/keyboard/timer correction has independent corrective approval for complete
  `d90aded7..0f3a1a7a`. RP-316's shared oracle has its own complete
  `c0eb3dc5..876b331c` approval. The 2026-11-02 log records actual-minute/native
  input checks and an unforced Lucky claim through the real composed driver at
  revision42/run3, with exact successor/SQL proof and no account selection.
  This is not natural whole-game pacing, standalone cross-account identity or
  full GS5 acceptance; earlier click-buff-only runs remain historical.
  A later root execution supplement passes80/80 affected Linux Firefox cases,
  including actual60054ms idle; native macOS startup remains RP-256, not repaired.
  RP-317 diagnostic mapping/host repair
  has independent corrective approval for complete `0f3a1a7a..c0eb3dc5`, including
  its final record edge (2026-11-02 owning log). Producer supplement
  `87fd23d4..d90aded7` requires
  Claude independently.
  This existing checkbox records implementation, not acceptance or archival.
- [x] GS6 provisioning caps + owned-upgrade text on the Desk.
  OD-15's missing upgrade reasons now also have a server → generated API → decoder → Desk
  implementation, with snapshot-only explanation and an accessible button description.
  See the latest owning log for executed checks and pending designated review. New reason
  copy remains visibly pending owner adoption; this does not complete Garage acceptance.
- Shared GS0.2 supplement (not an acceptance checkbox): RP-322 through
  `efa557b8` locally restores rate-limit/exclusive refresh-before-reactivation.
  Baseline mapper/native failures and separate arm severing discriminate;
  full Garage Chromium/WebKit100/client9,737/373 skips/types/build/boundaries/
  actual composed pass. Runtime-double refused intents, not real429/Soul
  sessions. Complete `3f867956..b64a91af`, including final tracking, is independently
  corrective-approved; the 2026-11-02 log records historical-function discrimination
  and current three-engine mounted checks. Care and other ranges remain separate.
- [x] GS0.3 event decoders (announcements only): achievement and meter-band announcements.
  Review note: original `301728c8^..301728c8` is **CHANGES REQUIRED** for RP-312;
  the scoped direction repair/simulated reconnect has independent corrective approval
  for complete `7aab0e2e..900f409e`, including planning/record edges (2026-11-02 log).
  Actual historical function comparisons, current runtime and mounted three-engine
  checks pass. The existing box is not an archival or full-GS0.3 gate.
- [x] GS0.3 remainder: `fiscal_period_harvested.v1` nav badge (OD-3) and `buff_started.v1` announcement (`6594b646`).
  Bounded designated Codex verdict: `6594b646^..6594b646`, all five paths,
  APPROVED in the 2026-10-07 Garage log; not acceptance of the full lane,
  later records, RP-312 correction, Firefox or manual accessibility gates.
- [x] GS4 pet care surface over the PA7 arm + cosmetic overlay (G10) (`7a61e4b6`); raw-care fields blocked by DESIGN-GAP GS4×PA7.
  Designated Codex review of original `7a61e4b6^..7a61e4b6`, all eight paths:
  CHANGES REQUIRED for existing RP-132/GS4×PA7 author contract hold. Native4/
  cold Go/five actual fault probes verify mechanics only. Care's narrow
  RP-319/320/321 correction is independently APPROVED for complete
  `c7d8f815..3f867956`, including its final record edge. Reviewer438/438 current
  Linux three-engine Fiscal/care/feedback cases and types/build/boundaries pass;
  historical failing-first, eight faults and real care/receipt/reload evidence
  are inspected, not rerun. Its fresh real-service Cosmetic driver fails before
  native Buy/care at RP-705; the 2026-11-02 log preserves that red and limits.
  No current composed-pass or full GS4 credit follows. RP-318's separate status
  correction already has its exact independent approval above; owner copy,
  raw-care/wire, actual AT and full-feature/archival gates remain.
- [x] 320 px reflow measurement across Desk/Fiscal/Meters/Trophy Case/pet (`6594b646`).
- [ ] Canonical docs, cold gates, hand-off for Codex designated review (never self-archive).
