# RFC: Accessibility of Player Workflows

- **Status:** draft — not implementation authority
- **Author:** Codex (draft for Marco's acceptance)
- **Created:** 2026-09-23
- **Design refs:** `design/00-vision.md` (browser-first, mobile-usable; honest play),
  `design/06-tech.md` (Svelte/DOM-first UI), `design/11-ux-writing.md §7` (core-loop keyboard,
  reduced motion, non-color states, primary-economy labels)
- **Depends on:** archived UI Foundation and Game UI Screens; the current release-task manifest
  must be exact before final acceptance; Account and Minigame/Recovery successors own their
  unbuilt surfaces
- **Parent / amends:** follow-up to archived UI Foundation and Game UI Screens; does not edit
  either archive
- **Supersedes / superseded by:** —
- **Planning:** `planning/accessibility-player-workflows/` after acceptance

## Summary

The existing fixture-level axe and primitive keyboard gates are useful but do not establish an
accessible player journey. At source `7e8aa70`, the Offer drops focus to `<body>`, the Desk is
647 pixels wide in a 320-pixel viewport, and OS reduced motion never reaches the numeric shell.
This successor specifies the shared player-workflow behavior and evidence floor. It begins with
the mounted Phase-0 Game UI and applies as an acceptance dependency to every later release
surface; it cannot declare an unbuilt account, minigame or later-tier workflow accessible.

## Scope and authority

The initial implementation surface is the five mounted Game UI surfaces (`vision_slide`,
`desk`, `offer_sheet`, `run_end`, `settings`), persistent chrome, drain/resync notices, and their
`GameUIShell`/`ShellController` display path. A release-task register, owned by the exact release
manifest, enumerates the default tasks and each surface/action that a player must use. The
initial rows include Begin → Desk, manual/buy/cross-gate/Wind Down, Offer accept/decline,
Run-End → next Company, Settings and system recovery. No fixture-only navigation counts as a
default-task row. When Account rights, Minigame/Recovery, or later phases add a surface, their
accepted RFCs must add its tasks and produce the same evidence before its release. This RFC
does not authorize those products, choose their copy, or shorten 1.0 to the Phase-0 subset.

This is accessibility of game interaction, not a replacement for the separate legal/privacy,
content or clean-host release gates. No authored text is changed here; screen-reader names and
announcements resolve through the existing copy pipeline where player-facing wording is needed.

## Proposed specification — pending owner acceptance

### A1 — focus and dynamic context

1. Every active surface has one semantic heading. On a player-selected tab change, keep focus
   on the still-mounted navigation control. On an authoritative Offer/Run-End preemption, move
   focus after mount to the new surface heading using a programmatically focusable,
   non-tab-stop heading (`tabindex=-1`), even if the old focused nav element still exists:
   the context change itself must be exposed. Do not leave focus on `<body>` or force an
   unrelated visible control. Non-preemptive snapshot/restyle updates preserve a still-mounted
   focus target. The archived Game UI rule that an open destructive confirmation is not
   preempted remains binding.
2. Offer, Run-End, continuation and resync transitions expose the new heading, purpose and
   available choices to assistive technology exactly once. Focus placement supplies context;
   status/alert regions retain their distinct roles for asynchronous system messages. Avoid
   duplicated simultaneous heading and live-region announcements. When continuation removes
   its own button, focus the newly mounted Desk heading; when a recoverable resync preserves
   its triggering control, retain that focus and use the status/alert channel.
3. After continuation or a recoverable error, keyboard users can reach the next actionable
   control in DOM order. A disconnected, stale or expired action neither hides the current
   focus target silently nor strands focus on a disabled/removed control.

### A2 — keyboard and pointer

1. Every release-task control is reachable in a logical Tab/Shift-Tab order, visibly focused,
   operable by standard native Enter/Space behavior, and usable without hover, drag or a
   pointer-only gesture. A visible disabled state must have an available reason or a nearby
   status that does not depend solely on color.
2. No surface traps Tab, intercepts browser shortcuts, or requires a timed key sequence.
   Authoritative offer expiry remains an honest time rule, but must provide a readable
   remaining-time state and a recoverable expired-action result; accessibility may not change
   payout or server authority.
3. On a coarse pointer, every actionable target has at least a 24×24 CSS-pixel hit area or
   the explicitly measured separation allowed by the accepted target-size criterion. The
   task is tested by touch/pointer input, not inferred from a CSS rule alone.

### A3 — reflow and non-color information

1. At a 320 CSS-pixel viewport, with normal text and at 200%/400% browser zoom on the
   declared desktop reference viewport, every included task remains reachable without
   horizontal page scrolling or clipped controls. A deliberately two-dimensional data view
   may have its own named scroll container only if its contents genuinely require two axes;
   ordinary copy, the README skin, controls, headings and terms do not qualify. Test the
   complete mounted surface including persistent chrome, not a cropped component.
2. Every essential state—affordability, cap, selection, warning, success, error, Guild/mode
   status when later shipped—has text, shape or other non-hue distinction. A color-vision
   simulation/manual review complements the structural assertion; a palette screenshot alone
   cannot prove comprehension.

### A4 — reduced motion

1. Respect `prefers-reduced-motion: reduce` on initial mount and when the preference changes
   during play. Attach one `MediaQueryList` change listener and dispose it on unmount. The
   resulting setting drives both UI theme durations and `GameUIShell`/`ShellController`/
   `DisplayCounter`; CSS-only reduction is insufficient.
2. Entering reduced motion during an in-flight counter interpolation ends that interpolation
   at the latest authoritative target, suppresses the pulse, and keeps the displayed number
   valid. Future authoritative changes snap without pulse. Leaving reduced motion affects
   only future transitions; it must not replay missed effects. The setting never alters
   server state, sim rate, timer truth or intent semantics.
3. Later canvas, pet, minigame and world surfaces must expose an equivalent control path and
   task proof before inclusion. This RFC does not invent their renderers or approve a generic
   animation exemption.

### A5 — evidence record

For every release-task row, record the artifact/source coordinate, browser/OS, viewport and
zoom, input mode, assistive technology and version where used, start state, player actions,
observed result and failure state. Automated axe results retain pinned WCAG 2.2 AA tags and
zero serious/critical threshold, **in addition to** task completion. Manual records distinguish
what was directly observed from an inference. The register is revised when a new surface,
dynamic state or input mode ships; old release observations do not certify new tasks.

## Deviations from design

None proposed. The research suggestion of an independent in-game reduced-motion toggle is not
silently adopted: its setting ownership and any player-facing wording need an owner decision
before acceptance if it belongs in this release. The OS preference path is required regardless.

The proposed 320 CSS-pixel/400% reflow and 24 CSS-pixel target floors follow the current
[W3C Reflow explanation](https://www.w3.org/WAI/WCAG21/Understanding/reflow) and
[W3C Target Size (Minimum) explanation](https://www.w3.org/WAI/WCAG22/Understanding/target-size-minimum).
The live reduced-motion path is a project requirement; W3C's
[reduced-motion technique](https://www.w3.org/WAI/WCAG22/Techniques/css/C39) is informative and
does not imply CSS alone covers this game's numeric shell. These references support the
measurement choices, not a conformance claim.

## Acceptance criteria — every gate must discriminate

1. **Focus/context:** Chromium, Firefox and WebKit browser tests place keyboard focus on a
   Desk action and separately on persistent nav, inject an authoritative Offer/Run-End, and
   assert the correct new heading is focused and context exposed. A seeded removal of the focus
   manager reproduces `<body>` or leaves nav oblivious to the preemption and fails. Persistent
   nav focus under a non-preemptive tier restyle remains unchanged.
2. **Reflow/zoom:** full five-surface/chrome states at 320 CSS pixels, plus manual 200%/400%
   zoom, have no horizontal page overflow or lost action. Restoring the current 647-pixel Desk
   defect must fail the automated oracle. Tests check visible usable controls as well as widths.
3. **Keyboard/pointer:** from a clean browser profile, Tab/Shift-Tab and Enter/Space complete
   every built release-task row; coarse-pointer execution hits the intended controls. A seeded
   trap, pointer-only action or hidden focus indicator fails. Fixture clicks alone do not count.
4. **Motion:** initial and live-changed OS preference reaches CSS and numeric shell, stops an
   in-flight interpolation/pulse without changing authoritative state, and survives era changes.
   A CSS-only mutant and a severed media-query listener both fail the dedicated test.
5. **Assistive/non-color:** axe runs on each included dynamic surface/state with zero
   serious/critical findings; manual screen-reader task records prove heading/choice/error
   comprehension, and color-vision checks verify every essential state has non-hue meaning.
   Seeded unlabeled-status and hue-only fixtures must fail.
6. **Composed proof:** the actual release browser → real API/WebSocket → Postgres path completes
   the mounted default tasks under keyboard-only input, including an Offer/Run-End transition
   and recovery. Any absent account/minigame/later-tier task remains an explicit blocker, not a
   skipped or vacuously passing test. Server-side test setup may establish preconditions, but
   all player actions originate from enabled DOM controls. R-005 runs only after all
   release-floor tasks exist.
7. **Closeout:** canonical `docs/` behavior, release-task register, backlog, per-RFC log and
   exact cross-party review range reconcile in the same closeout. No claim of full 1.0 task
   accessibility follows from the Phase-0 implementation range.

## Open questions before acceptance

1. Adopt the exact Phase-0 release-task manifest; the current D-001/D-007 ruling fixes the
   bounded T0–T1 direction but not every included task/surface.
2. Rule D-018's supported manual assistive-technology/browser/OS matrix and who executes/releases
   its records. The current Mac run could not start Firefox twice; that is a test-environment
   gap, not permission to drop Firefox or screen-reader evidence.
3. Decide under D-018 whether the independent in-game reduced-motion and visual-noise settings recommended
   by design research belong to the next release, and adopt any required copy separately.
4. Resolve under D-018 whether the current Offer countdown's authoritative expiry needs an additional
   non-visual announcement cadence; no cadence or gameplay extension is invented here.

## Changelog

- 2026-09-23: draft from the current-source focus/reflow probes and reduced-motion call trace
  (`planning/platform-alignment/accessibility-current-audit.md`). No implementation authorized.
