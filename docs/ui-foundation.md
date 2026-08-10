# UI Foundation

The client exports one content-free UI boundary from `client/src/ui/`. It owns theme loading,
large-number presentation, Copy-backed primitive text, and surface subscription ownership. It
does not own a game screen, transport connection, authoritative calculation, or balance mutation.
Game UI and minigame surfaces build on this boundary.

## Era themes

[`ui/themes.schema.json`](../ui/themes.schema.json) is the closed theme wire. Phase A ships exactly
`era_1995` and `era_2000`, each with the same 41 tokens across color, type, space, border, chrome,
and motion groups. The runtime loader rejects missing, additional, and out-of-domain leaves before
installing them as `--cc-{group}-{key}` properties on a single `data-era` root. An era change
replaces only that attribute and those tokens; primitive DOM structure remains unchanged.

The 1995 theme deliberately has a zero-motion budget. A reduced-motion preference also installs
zero duration values without changing the catalog artifact. Component styles may use layout
geometry directly, but governed colors, fonts, shadows, radii, and durations must reference
`--cc-*` tokens. The source gate parses Svelte style ASTs and rejects seeded static, inline, and
directive violations.

## Amount

`Amount` accepts a canonical Decimal string and an optional `{amount, reason_key}` cap. It parses
through the numeric boundary, formats with `@antimatter-dimensions/notations` 1.6.0 Standard
notation, uses three significant mantissa digits, and renders values below 1,000 as integers.
Negative values use the Unicode minus sign. Cap explanations resolve through the Copy catalog;
the mechanical reason key is never rendered.

The first render and cap changes are synchronous. Ordinary value changes share one 100-ms
scheduler, so a 20-Hz source produces no more than ten visible updates per second. The scheduler
coalesces by component instance and cancels pending work at unmount.

## Surface ownership

Surface rows are byte-sorted `{surface_id, mount_id, unlock}` records. Unlocks are either `always`
or `fact_equals` against a declared authoritative shell fact. Unknown facts, duplicate IDs, and
duplicate mount targets fail load. The host owns exactly one active subscription: it unsubscribes
and unmounts the old surface before mounting the next, and inactive surfaces receive no callbacks.

## Fixture and gates

The production entry currently mounts the content-free primitive fixture. That fixture is the one
Vite/browser test route; there is no Storybook or second UI build. It exercises both eras, Amount,
Copy, focus, names, cap presentation, reduced motion, and `axe-core` 4.12.1 under WCAG 2.2 AA tags.
Serious or critical violations fail the browser suite in Chromium, Firefox, and WebKit.

Run:

```sh
make verify-client-boundary
make verify-schema
make typecheck
make test-client
make test-browser
make build-client
```
