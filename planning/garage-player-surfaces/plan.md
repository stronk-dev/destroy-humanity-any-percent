# Garage Player Surfaces implementation plan

RFC: `rfc/garage-player-surfaces.md` (accepted 2026-09-25; owner decisions at recommended defaults).
Scope of this plan: release-manifest Batch A slices over already-live backends only. The pet (GS4),
cosmetic, Clout, T2-era and ticker slices are out of scope until their producers land.

Constraint for this lane: no kernel-guarded path (`kernel/affecting-paths.json`) is edited. A slice
that would need one is recorded as a blocker in `log.md` instead.

Latest bounded repair: `ce5c4ea8` (RP-329 / GS3-A1), predeclared bff97b20 /
c5d5220c, failed-first 43ef0a21. Non-null meter data must contain exactly
the existing eleven-ID contract; no second list/catalog/producer/kernel/
balance/schema change. Baseline: 27 admission failures. Five independent
client faults fail 27/11/4/3/23 assertions; both actual producer-control
faults fail; all transient source restored. Clean client 9,814 passes /
446 explicit browser skips, types/build/boundaries/cold Go/vet and native
Garage Chromium/WebKit 246 pass. Two SQL tests explicitly skip; isolated
performance 1 passes but its null-feature fixture is not populated AC7.
Copy/content-manifest/topology and 13 negative controls pass. Complete
range after 5fbf4cff through this record needs Claude; RP-332 separately
cc62cea8..5fbf4cff, audit fc911784..cc62cea8, runtime a42406f0..fc911784
and earlier ranges remain owed. Inventory covers all 29/eight gates without
promoting acceptance. Next accepted work: RP-330 / GS3-A3 semantic narrow
layout, then populated AC7 / exact guards. RP-331 real persisted/DOM proof
waits for capacity repair/recheck; last measured Docker overlay is full,
no new run/cleanup. Author/body/Firefox/AT/full 1.0 gates remain; no box or
lifecycle change.

RP-330 / GS3-A3 is now in progress under `d9a8ca27`: three new native
declarations, six Chromium/WebKit executions. Corrected typed-clean baseline
50408 has six semantic failures; the wide control passes and every scenario
fails on the remaining narrow table. Initial native-value attribute oracle
error disclosed and replaced by exact native properties before production.
No MetersSurface change yet. Next: separately record bounded semantic markup /
live-breakpoint repair and negative probes, then full gates. This new range
starts after bddfc58e and remains independent of RP-329 5fbf4cff..bddfc58e.

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
