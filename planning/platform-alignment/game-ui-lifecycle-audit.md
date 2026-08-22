# Game UI Screens lifecycle audit

Coordinate: product tree `190a4fa`; audit checkpoint after `14282e1`; 2026-08-20.

## 2026-08-22 remediation update

The historical findings below remain the evidence at their audited coordinate. Queue row 4 has
since implemented their routed repair in `05acc65^..fa5fc42`: current snapshot v3 carries a
server-projected transition object backed by the existing production transition; v1/v2 receipts remain legal and inert; the three governed
controls traverse `runtime.ts`; both terminal variants and exact-next-run continuation pass in the
real composed browser/Postgres topology; and cap/drain/resync now have independent browser oracles.
Eleven targeted severings failed and were restored, including the two product defects the composed
test exposed (UUIDv4 intent IDs and eager-snapshot suppression of `run_ended`). The batch awaits the
mandatory Claude exact-range review and is not archival-eligible. AC5's manual 4×/frame observation
remains separately open exactly as this audit classified it.

This pass re-derived the active Game UI RFC from its full specification, blockers and owner
rulings, live plan/log, canonical doc, Svelte/runtime/event/timing/performance implementations,
generated snapshot contracts, component and composed-browser tests, successor content work, and
tracked cross-party verdicts. Temporary discrimination mutations were restored. No product code,
owner-authored RFC/design text, canonical product documentation, or implementation checkbox was
persistently changed.

## Bottom line

The current browser can bootstrap an anonymous account, persist credentials, fetch snapshot v2,
subscribe to initial world presence, render five Phase-A surfaces, buy/click through authoritative
intents, show offers and run-end payloads, preserve display-only timing, switch T0/T1 themes, and
render governed Copy/Presentation data. The structural component boundary and byte-only Run End
contract are genuinely discriminating.

The Game UI is not acceptance-complete:

- the accepted `game_ui_snapshot.v3.transitions` projection does not exist;
- the Desk has no server-eligible `cross_gate` or `wind_down` controls;
- Run End has no accepted “Start the Next Company” continuation control;
- the composed browser test ends at Desk, snapshot v2, WebSocket presence, and the visitor counter;
- the C25/C28 owner ruling explicitly narrowed AC1 to those three controls and rejected a full
  two-hour browser replay, but the normative RFC still requires “the full T2 script”; and
- the owner ruling lives only in the planning log. The RFC header still stops at GU-C24, U2 still
  fixes all Phase A to `era_1995` despite GU-C7's two-era ruling, and canonical docs still say the
  archival gate owns the full browser rendition.

The prior log's “AC2–AC5 MET” claim also overstates the evidence. AC2 and AC3 are proven. AC4's
oracle is vacuous: removing the cap explanation plus both system notices leaves all 20,007 browser
assertions green. AC5 proves only the ruled CI-observable update-count/long-task arm; its 4× CPU
throttle and dropped-frame fields are never consumed, and no manual reference-device result is
recorded.

## Current cold evidence

All valid commands ran from repository root:

- restored `make test-browser` — 123 files passed; 20,007 assertions passed, 3 skipped across
  Chromium/Firefox/WebKit; isolated performance lane 1 passed, 10 deliberately filtered/skipped;
- restored `make typecheck verify-client-boundary` — TypeScript and Svelte zero diagnostics;
  boundary gate green;
- the prior same-coordinate API wave ran `./publicapi ./account ./gameserver -count=1` green;
- the earlier audit batch ran the real composed Postgres/Chromium bootstrap, snapshot, and world
  handshake green. That bounded script is not restated as a transition or first-hour proof.

The performance-only browser body completed in about 290 ms while injecting 1,200 fixture
snapshots described as sixty simulated seconds. That is appropriate for an update-count fixture,
not a measurement of sixty wall-clock seconds, 4×-throttled frame delivery, or dropped frames.

## Restored discrimination probes

1. Added a direct import from `../transport` to `GameUIApp.svelte`. `make
   verify-client-boundary` exited 2 naming the component and forbidden transport import. Restoring
   the import returned the boundary gate to green. AC2's direct component boundary discriminates.
2. Widened `RunEndSurfaceProps` with an optional `snapshot`. `make typecheck` exited 2 because the
   negative contract's `@ts-expect-error` became unused. Restoring the exact `{ended}` prop returned
   TypeScript/Svelte to green. AC3's compile-time isolation discriminates.
3. In one mutation, suppressed the drain notice and resync story-beat blocks and removed the
   `cap` prop from every rendered resource amount. Full `make test-browser` still exited 0 with
   20,007 passed/3 skipped, and the isolated performance lane also passed. The complete mutation
   was restored; the same full browser and type/boundary populations returned green.

The product worktree after restoration differs only by the user's pre-existing `AGENTS.md` edit;
the remaining changes are this audit's planning records.

## Acceptance classification

| AC | Verdict | Evidence and limitation | Required closeout |
|---|---|---|---|
| AC1 | **Unmet and body-blocked** | The composed browser stops at initial Desk/presence. Current snapshot is v2 and neither accepted transition projection, transition buttons, nor next-run continuation exists. C25/C28 narrows the browser obligation to visible/eligible controls, correct runtime intents/receipt handling, and Run End→next run, but the RFC still requires the rejected full T2 replay. | Ruling author reconciles AC1/U1/header to C25–C28. Then implement v3's one kernel-owned preview, the three adopted controls, and a composed proof whose server-side setup cannot submit the tested actions; disconnecting any control/surface must fail. |
| AC2 | **Proven structural boundary** | Both components compile over generated/decoded types. The boundary scans both Game UI Svelte files; a restored direct transport import makes it fail with the exact file. Runtime owns decoding outside the component boundary as ruled. | Preserve the negative import probe and include all current component/boundary ranges in the final review union. |
| AC3 | **Proven structural and rendered integration** | `RunEndSurfaceProps` is exactly `{ended}`; adding a snapshot makes the compile-time negative fail. The browser mounts the component with decoded `run_ended` bytes and renders terminal/curriculum/payout content without a snapshot prop or shell import. | Preserve both the type-level negative and byte-only browser rendering across the v3/continuation change; continuation stays in the parent. |
| AC4 | **Failed acceptance oracle** | The cap explanation and both system notices exist mechanically. The only combined fixture sets drain/resync and runs no content/transition assertion beyond axe/mechanical-ID absence. Suppressing all three outcomes simultaneously leaves every browser assertion green. | Add three independent browser cases asserting exact governed copy, state transition, action, and failure behavior. Each must fail when its one producer/render/action seam is removed. |
| AC5 | **Partial** | The scenario pins 1,200 inputs, ≤600 formatted commits, 200 ms long-task ceiling, and viewport. Unit validation rejects a combined over-budget object; the isolated Chromium count/observer fixture passes. `cpuThrottle` and `droppedFrameAllowancePPM` are referenced only by config equality, no CDP throttle/frame metric runs, and no manual mid-range-Android result is recorded. | Retain the fast deterministic CI arm, add separate per-bound negative cases, and execute/record the ruled 4× reference-device manual release check including actual dropped frames—or have the ruling author reconcile the literal AC to the CI-only claim. |

## AC1 producer-to-consumer trace

The missing user path is exact rather than aspirational:

1. Server production already accepts and replays `cross_gate` and `wind_down` intents.
2. Archived first-hour content proves the required milestones headlessly and through a composed
   Postgres server run. C28 intentionally says the browser need not replay those two hours.
3. The current Game UI snapshot schema has no `transitions` member. The generated TypeScript,
   strict client decoder, server projection, Bootstrap receipt, and API compatibility pin all stop
   at v2.
4. The Desk renders manual action, generator, and upgrade buttons only. Repository-wide UI/copy
   search finds none of the three adopted keys or texts outside the owner-ruling log.
5. `RunEndSurface` correctly remains payload-only, but its parent renders only `<RunEndSurface
   {ended}/>`; there is no continuation action to fetch and verify the next snapshot.
6. `test-game-ui-composed.mjs` clicks Begin, waits for Desk and visitor presence, directly probes
   snapshot v2, then exits. It performs no post-bootstrap UI action.

This is why server first-hour proofs and the browser bootstrap proof cannot be added together and
called AC1. The accepted missing consumer is small, but it is still missing.

## AC4 and AC5 oracle anatomy

The browser test named for axe visits all five surfaces, calls both `fixtureSystem` arms, then runs
`assertNoMechanicalPresentation` and axe. Neither helper requires drain title/body, resync
title/body/button, nor a state change from pressing the resync action. No test asserts the resource
cap reason text. The restored three-seam deletion passing the full population proves RP-026 rather
than merely suggesting weak assertions.

The performance test does execute the UI and observe formatted DOM commits. It manually flushes
the shared formatter every second injected input, so its 600-commit bound is useful. Its other
claims must stay bounded:

- `durationMS` contributes to the declared `inputCount`, not wall-clock duration;
- `cpuThrottle:4` is not passed to Playwright/CDP or any runtime;
- `droppedFrameAllowancePPM` is never read by the validator or browser test;
- no frame timestamps or dropped-frame count are observed; and
- when long-task observation is unavailable, the fixture substitutes zero rather than failing
  loud or recording the missing instrument.

Canonical `docs/game-ui.md` correctly calls throttle/drop a manual release profile. The planning
log's unconditional “AC5 MET” and active acceptance ledger did not preserve that qualification.

## Normative, canonical, plan, and review drift

1. The RFC header says GU-C1–GU-C24 ruled even though C25–C28 were owner-ruled later in the planning
   log. Those rulings were never incorporated into the RFC.
2. AC1 still requires the full T2 browser script, directly contradicting C25/C28's narrowed
   controls-only browser responsibility. The existing implementation blocker correctly refuses to
   choose for the owner.
3. U2 still says `era_1995 for Phase A`, although GU-C7 says tier 0→1995 and tier 1→2000 and calls
   that original sentence superseded. The implementation and canonical doc follow the ruling, but
   the plan checkbox claiming C1–C8 body reconciliation is false.
4. Canonical `docs/game-ui.md` says the archival gate owns “the full browser rendition of the live
   sequence,” which is the C28-rejected obligation. It also describes snapshot v2 accurately and
   therefore cannot document the not-yet-built v3 controls as shipped.
5. Cross-party verdicts validly cover the major Phase-A/remediation slices, `02d00d7`, performance
   isolation, and later curriculum/Run End changes in their respective logs. The active Game UI
   log has not assembled those current-history ranges into one final criterion union, and future
   C26/C27 implementation necessarily remains uncovered. The plan correctly leaves designated
   review and archival open.

## Smallest honest closeout order

1. Ruling author edits the RFC header, U1/U2, and AC1 so C7 and C25–C28 are normative in the body;
   canonical docs adopt the narrowed browser responsibility only when that behavior lands.
2. Implement snapshot v3 through the existing generated API/Bootstrap compatibility lane using
   the one pure kernel preview; preserve v1/v2 receipt replay and fail-closed controls.
3. Add the adopted gate, Wind Down, and parent-owned next-run controls. Every non-wait action must
   originate from an eligible visible DOM control and traverse `runtime.ts`.
4. Build the narrowed composed test with ordinary server-side precondition setup, then prove each
   control's visibility, exact intent, receipt/snapshot result, and Run End→next-run transition.
   Retain the ruled disconnect/suppression negative mutations.
5. Replace AC4's axe-only fixture with independent cap/drain/resync behavior tests and their
   one-seam mutations.
6. Keep AC5's deterministic count/long-task arm, make missing long-task instrumentation visible,
   add independent bound negatives, and record the ruled manual throttle/frame result.
7. Reconcile plan/log/docs and assemble all successor verdict ranges plus the new implementation
   into one exact current-history union; obtain the designated cross-party verdict before archival.
