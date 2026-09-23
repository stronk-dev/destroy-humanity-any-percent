# Current-source accessibility check — Game UI release floor

2026-09-23. Product source `7e8aa708ff9e002e7ec4b46242d7d889f76b7fdd`;
research plan committed at `dc77031`. This checks three previously reported defects against
current source. It is not R-005, a WCAG conformance claim, a manual screen-reader study or an
accepted implementation contract. The older `accessibility-release-audit.md` is dated at
`190a4fa`; this checkpoint revalidates only the facts below.

## Executed probes

Two temporary desired-behavior tests were placed beside the committed five-surface fixture in
`client/test/game-ui-screens-browser.test.ts`, run via the root `make test-browser` selector,
and removed afterward. The tests asserted target state before measuring; they did not use a
clipped card, a programmatic replacement of the UI, or a passing assertion that cannot detect
the defect. The deliberately failing run was stopped after Vitest printed its verdict; the
Make target's second performance command did not run and is not claimed. [V]

| Property | Current observation | Verdict |
|---|---|---|
| Authoritative Offer preempts a focused Desk manual button | Chromium and WebKit: Offer heading exists but has `tabIndex=-1` (not programmatically focusable); `document.activeElement` becomes `BODY`. | **Fails** meaningful focus transfer in both executed engines. RP-082 persists. |
| Full Desk at 320×720 CSS-pixel viewport | Chromium and WebKit: `main.scrollWidth=647`, `main.clientWidth=320`; document `scrollWidth=647`, `clientWidth=320`. | **Fails** horizontal reflow in both executed engines. RP-083 persists. |
| Firefox arm | Multi-engine run and a separate `--browser.name=firefox` retry both failed to connect to the Firefox browser session within 60 seconds, before setup/import/test execution. | **Invalid/unexecuted**, not a product pass or fail. Three-engine acceptance remains open. |

The diagnostic assertions were *expected to fail* on broken behavior and did. In both executed
engines the emitted exception included the observed focus element or width pair. The temporary
test file was restored byte-identically. The browser runner wrote ignored failure screenshots;
they are diagnostics, not tracked or cited as acceptance baselines.

## Current production path trace

`GameUIApp.svelte` installs CSS theme tokens from a one-time/effect-time
`matchMedia('(prefers-reduced-motion: reduce)').matches` sample. There is no media-query
`change` listener. `GameUIShell` constructs `ShellController(policy)` without a reduced-motion
argument; the controller's default is `false` and passes that value to each `DisplayCounter`.
`DisplayCounter` therefore permits interpolation and pulse even if the theme's durations are
zero, and its reduced-motion field cannot change mid-session. This is a source-level failure of
the OS-preference-to-numeric-shell path, not a measured user/AT animation study. RP-084 persists.

The committed axe fixture, native controls and headings remain useful primitive evidence, but
its default 1280×720 viewport and programmatic Begin focus cannot detect the observed failures.
The existing composed Game UI journey proves visible gameplay controls, not keyboard traversal,
assistive announcements, zoom, touch or device preference changes. Account recovery/export/delete
and Minigame/Recovery surfaces are not mounted player tasks; they must not be marked accessible
by this check.

## Consequence and next proof

The Phase-0 release floor cannot cite the current Game UI as task-accessible. A cross-surface
successor contract must own focus/context on lifecycle replacement, full-surface 320-pixel and
200%/400% reflow, live OS reduced motion including numeric shell, keyboard traversal,
coarse-pointer and non-color state, plus manual screen-reader records. It must name seeded
failures and require a real default player workflow in addition to fixture checks. The owner
must accept the task and assistive-technology matrix; future account/minigame/later-tier
surfaces join the same release floor rather than inheriting a pass from current fixtures.

No product, theme, copy, test or accepted RFC behavior was changed by this research.
