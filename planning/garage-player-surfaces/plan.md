# Garage Player Surfaces implementation plan

RFC: `rfc/garage-player-surfaces.md` (accepted 2026-09-25; owner decisions at recommended defaults).
Scope of this plan: release-manifest Batch A slices over already-live backends only. The pet (GS4),
cosmetic, Clout, T2-era and ticker slices are out of scope until their producers land.

Constraint for this lane: no kernel-guarded path (`kernel/affecting-paths.json`) is edited. A slice
that would need one is recorded as a blocker in `log.md` instead.

Latest bounded work: RP-337 GS2-A3 verification-only source guard under
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

Latest bounded GS0.6 supplement (not a new acceptance checkbox):45bf2edb,
predeclareda8f2a527/test-first7cac400d. Fiscal/care own mapped notice regions,
origin captured before act awaits.24 baseline DOM failures; five independent
compiling faults fire/restored exactly. Full native two-engine Garage174/
performance1, client9,737/410browser skips/types/build/boundaries/copy/
manifest/topology pass. No previous assertion relaxed. Changed-tree composed
not rerun under RP-236; no earlier-source substitution. Whole span after
e950216a through tracking needs Claude separately from browser4dcd9969..
e950216a/Fiscalb64a91af..4dcd9969 and prior ranges. HTTP-failure/queued-origin/
all-surface/AT/Firefox/fullGS0.6/full-Garage/release remain unproved. Next
accepted native HTTP-error origin/refresh census; Docker capacity preflight
must succeed before another container population. No lifecycle promotion.

- [x] GS0.2 runtime: `intent()` returns the typed outcome; non-2xx throws `GameUIRequestError`;
  `act()` scopes `expected_revision` (Company vs Founder); rejected intents render their reason.
- [x] GS0.1 projection: Game UI snapshot v4 (`features` arms + `generators[].provision_cap` +
  feature facts), registered schema, generated types, client parser; live sync requires v4.
- [x] GS3 Meters arm + surface.
- [x] GS2 Achievements arm + surface.
- [x] GS1 Fiscal arm + surface.
  Narrow pending/readiness/row-reason correction `891aee10` is locally verified,
  not designated-approved: failing-first33 native assertions; final full Garage
  Chromium/WebKit150/client9,737/398 skips/types/boundaries pass. Independent
  faults and refined former shortcuts discriminate; invalid instrumentation
  runs disclosed. Same production tree build/copy/manifest/topology/actual
  composed pass. Whole new span after `b64a91af` through final tracking requires
  Claude, including predeclared tests and RP-325 refinements. RP-326 per-surface
  outcome ownership, Firefox/AT/full-lane/release gates remain; this existing
  checkbox still records implementation only, never archival acceptance.
- [x] GS7 minigames availability arm wired into the existing Pitch surface/nav.
- [x] GS5 active-play arm (`features.opportunity`, kernel export `ProjectActiveCombo`) + Desk opportunity region + composed claim witness (`cdb8fe61`, `f32f6175`).
  Original producer `cdb8fe61^..cdb8fe61`, all twelve paths: designated Codex
  CHANGES REQUIRED, RP-313 wire/clone body reconciliation. Green primitive
  tests do not resolve this. Original consumer `f32f6175^..f32f6175`, all
  fourteen paths, is separately CHANGES REQUIRED for RP-314/315/316. Narrow
  cap/keyboard/timer correction through `d9e4e274` passes 40 native cases;
  exact Codex correction `d90aded7..0f3a1a7a` needs Claude. RP-316's shared
  oracle through `0befc01c` is locally corrected (native86, six faults fire,
  actual click-buff composed pass, NOT Lucky); entire Codex span after
  `c0eb3dc5` through tracking needs Claude. RP-317 diagnostic mapping/host repair
  through `092ef3eb` passes 50 native cases; separate entire Codex span after
  `0f3a1a7a`, including its tracking edge, needs Claude. Producer supplement
  `87fd23d4..d90aded7` requires
  Claude independently.
  This existing checkbox records implementation, not acceptance or archival.
- [x] GS6 provisioning caps + owned-upgrade text on the Desk.
- Shared GS0.2 supplement (not an acceptance checkbox): RP-322 through
  `efa557b8` locally restores rate-limit/exclusive refresh-before-reactivation.
  Baseline mapper/native failures and separate arm severing discriminate;
  full Garage Chromium/WebKit100/client9,737/373 skips/types/build/boundaries/
  actual composed pass. Runtime-double refused intents, not real429/Soul
  sessions. Entire new span after `3f867956` through final tracking requires
  Claude separately from care `c7d8f815..3f867956` and all earlier ranges.
- [x] GS0.3 event decoders (announcements only): achievement and meter-band announcements.
  Review note: original `301728c8^..301728c8` is **CHANGES REQUIRED** for RP-312;
  the scoped direction repair and simulated-reconnect supplement are locally verified,
  not designated-approved. Claude must review exact `7aab0e2e..900f409e`,
  including planning/record edges. The existing box is not an archival or full-GS0.3 gate.
- [x] GS0.3 remainder: `fiscal_period_harvested.v1` nav badge (OD-3) and `buff_started.v1` announcement (`6594b646`).
  Bounded designated Codex verdict: `6594b646^..6594b646`, all five paths,
  APPROVED in the 2026-10-07 Garage log; not acceptance of the full lane,
  later records, RP-312 correction, Firefox or manual accessibility gates.
- [x] GS4 pet care surface over the PA7 arm + cosmetic overlay (G10) (`7a61e4b6`); raw-care fields blocked by DESIGN-GAP GS4×PA7.
  Designated Codex review of original `7a61e4b6^..7a61e4b6`, all eight paths:
  CHANGES REQUIRED for existing RP-132/GS4×PA7 author contract hold. Native4/
  cold Go/five actual fault probes verify mechanics only. Care supplement
  `ea06183b` with predeclaration `970d47a5`/test-first `0784f126` locally
  corrects RP-319/320/321: full Garage Chromium/WebKit82/performance1 and
  actual DOM-adopted care feed→bound receipt→public state/reload pass. Eight
  native faults and real callback severing fail and restore exactly. Test-only
  bundle, no status crossing/raw-care/full GS4-A5 acceptance. Firefox zero
  executed before connection timeout; RP-256 remains. Entire new span after
  `c7d8f815` through its tracking edge requires Claude independently. RP-318
  records the unimplemented pet-status decoder/
  announcement and missing owner-copy key. No acceptance/archival credit.
- [x] 320 px reflow measurement across Desk/Fiscal/Meters/Trophy Case/pet (`6594b646`).
- [ ] Canonical docs, cold gates, hand-off for Codex designated review (never self-archive).
