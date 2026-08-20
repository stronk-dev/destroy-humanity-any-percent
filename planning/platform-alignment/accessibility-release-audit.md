# Accessibility release audit

Coordinate: product tree `190a4fa`; audit checkpoint after `2b9a568`; 2026-08-20.

This pass traced the accessibility promises through UI Foundation, Game UI, theme/shell runtime,
all Svelte surfaces, browser configuration, automated tests, responsive layout, and the actual
release workflows. Two temporary desired-behavior browser probes ran in Chromium, Firefox, and
WebKit and were then removed with their screenshots. The restored declared Game UI browser and
performance populations passed. No product, test, design/RFC, copy, canonical-product-doc, or
theme byte was persistently changed.

## Bottom line

The repository has a useful semantic baseline:

- the complete five-surface fixture passes pinned axe-core WCAG 2.2 AA tags with zero
  serious/critical findings in Chromium, Firefox, and WebKit;
- controls are predominantly native buttons/details/meter/output, surfaces have labeled headings,
  system resync/drain uses alert/status roles, focus outlines are visible, and cap state is textual;
- Begin activates with Enter, Amount exposes a composed cap label, focus survives an era restyle,
  the 1995 theme has zero motion-duration tokens, and a primitive fixture proves forced reduced
  theme durations become zero; and
- current button line-height plus padding/borders mechanically exceeds the 24 CSS-pixel minimum.

That baseline is not a release workflow proof, and two core behaviors are now directly failed:

1. **Lifecycle focus/context fails in every browser.** With the Desk manual button focused, an
   authoritative offer preempts the Desk. The focused node is removed and `document.activeElement`
   becomes `<body>`; the new Offer heading is neither focusable nor focused. There is no surface
   live-region announcement or focus manager.
2. **320 CSS-pixel reflow fails in every browser.** The mounted Desk reports `scrollWidth=647`
   for `clientWidth=320`. Source inspection identifies the long unbreakable README separator inside
   a grid item as the likely intrinsic-width source; the layout has no `min-width:0`/breaking rule
   to contain it. That cause is an inference; the failed width itself is directly measured.

Reduced motion is also only partially wired. `GameUIApp` samples the media query for theme tokens,
but does not listen for preference changes. More importantly, `GameUIShell` constructs
`ShellController` without the preference, leaving its numeric interpolation/pulse mode at the
default `false` even when CSS durations are zero.

## Current evidence and its limits

| Evidence | What it proves | What it does not prove |
|---|---|---|
| Game UI axe pass over five fixture surfaces and two system beats, three browser engines | Current rendered fixture has no axe serious/critical result under selected WCAG tags. | Keyboard completion, screen-reader comprehension, zoom/reflow, coarse pointer, dynamic focus, account rights, minigames, or future screens. |
| Begin button is directly focused, outline checked, Enter pressed | One native activation and visible-focus style work. | Tab reach/order, all controls, first-hour completion, traps, shortcut conflicts, or error recovery. The test calls `.focus()` rather than reaching Begin by Tab. |
| Focus survives tier/era snapshot update | A persistent nav node remains focused during token restyle. | Surface replacement. The failed probe proves lifecycle preemption loses focus. |
| Primitive reduced-motion setter zeroes duration tokens | Theme installer can produce zero CSS duration values. | Production shell motion, OS preference changes after mount, non-CSS animation, future Pixi/canvas/minigame motion. |
| Native DOM and textual cap/system states | Strong semantic starting point and non-color-only examples. | Manual assistive-technology reading order, announcement timing, or task success. |
| One 1280×720 browser configuration | Desktop fixture behavior at that viewport. | Mobile/coarse pointer, 200%/400% zoom, portrait, reflow. The restored 320 px probe directly fails. |

The restored command after both failed probes was:

- `make test-browser BROWSER_TEST_FLAGS='test/game-ui-screens-browser.test.ts'` — functional
  Game UI population: 30 passed/3 skipped across three engines; isolated performance: 1 passed/10
  skipped. This confirms no temporary probe residue and also demonstrates why the green declared
  suite cannot currently detect the failed focus/reflow behaviors.

## Release-task coverage

| Task | Keyboard | Screen reader | Reflow/zoom | Coarse pointer | Reduced motion | Verdict |
|---|---|---|---|---|---|---|
| Bootstrap to Desk | Enter after programmatic focus | Axe only | 320 px not tested in suite | Not tested | CSS token baseline | **Partial** |
| Buy/manual core loop | Direct `.click()` after bootstrap | Axe fixture only | Desk fails at 320 px | Not tested | Shell preference not wired | **Failed/partial** |
| Offer/Run End lifecycle | Native buttons exist | New context not announced; focused node is removed | Not tested per surface | Not tested | CSS only | **Failed focus/context** |
| Account recovery/export/delete | Controls absent | Workflow absent | Workflow absent | Workflow absent | N/A | **Absent** |
| Pitch/Soul Recovery | Surfaces absent | Surfaces absent | Surfaces absent | Surfaces absent | Scheduler/toy unmounted | **Absent** |
| Full first hour/next run | Accepted controls/path incomplete | Not tested | Not tested | Not tested | Not tested | **Absent/body-blocked** |

## Contract and record drift

1. Game UI U2 requires a keyboard-completable first hour in addition to axe. No such path exists;
   the only keyboard event covers Begin and the RFC's full-browser criterion is body-blocked.
2. UI Foundation docs say the fixture exercises focus, names, and reduced motion. True for the
   primitive fixture, but insufficient for later screens and production shell motion.
3. Canonical Game UI documentation accurately says axe runs on five surfaces, but a reader can
   mistake that for WCAG/task conformance without the missing workflow qualification.
4. The internal mobile/PWA research already identifies absent breakpoint, target-size, safe-area,
   and viewport contracts. The production browser gate nevertheless fixes all engines to 1280×720.
5. Design research requires listening to mid-session reduced-motion changes and encoding status in
   shape/text, but the production media query is sampled only as an effect dependency unrelated to
   media changes.

## Smallest honest next order

1. D-001 defines which workflows are in the release floor. R-005 uses exactly those tasks; axe
   remains an automated arm, not the result.
2. Accept a cross-surface accessibility contract owning lifecycle focus/announcement, 320 px
   reflow plus 200%/400% zoom, keyboard traversal/activation, production reduced-motion wiring and
   live changes, coarse-pointer targets, non-color state, and assistive-technology manual records.
3. Repair RP-082–RP-084 under that accepted owner with checked-in failing fixtures. Do not bury the
   responsive fix in copy or shorten the README to satisfy the current row.
4. Only test workflows that exist. Account, Minigame/Recovery, Gate/Wind Down, and next-run
   accessibility remain dependency-blocked until their exact surfaces are implemented.
5. Run R-005 with the seeded focus-loss and unlabeled-status controls plus the now-observed reflow
   negative. Preserve browser/AT/version/viewport/preference evidence and obtain designated review.
