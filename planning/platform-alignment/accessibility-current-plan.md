# Current accessibility baseline — predeclared check

2026-09-23. Product coordinate: `7e8aa708ff9e002e7ec4b46242d7d889f76b7fdd`.
This is a bounded research check before a draft cross-surface RFC, not implementation authority
or a claim of WCAG conformance. The dated `accessibility-release-audit.md` is the hypothesis,
not current proof.

## Question and population

Do the already-observed lifecycle focus loss, 320 CSS-pixel Desk overflow and incomplete reduced
motion wiring still exist at the current product source? Test the shipped Game UI five-surface
fixture in Chromium, Firefox and WebKit for the first two. Trace the production
Game UI → ShellController preference path for the third. Account rights, Minigame, Recovery and
later-tier screens are absent/deferred and must be recorded as missing, not counted as passing.

## Method, controls and exit

Temporarily add browser diagnostics beside the existing fixture so they consume the same
component and state. For focus: focus the manual action, inject an authoritative Offer, and
record active element and new heading focusability. For reflow: set the viewport to exactly
320×720 CSS pixels, render the complete Desk, and compare document/Desk `scrollWidth` to
`clientWidth`; do not crop a single card or hide scrollable preformatted content. Run each
diagnostic cold across all three engines, retain the output, then remove the temporary edits.
The existing broad suite passing is a control, not a substitute for these probes. Source trace
for reduced motion must distinguish CSS theme tokens, shell numeric interpolation and a
mid-session media-query change. If any test is skipped, does not reach the targeted state, or
silently fails, the measurement is invalid.

Success for the **audit** is an explicit pass/fail per property, with the exact observed
measure and cause separated. A negative result is completed research. The result may support a
draft successor RFC and an owner decision on release-task coverage; it cannot authorize a
product change, assert screen-reader success, or close R-005. Every new defect or contradiction
gets a backlog row or an exact existing row reference.
