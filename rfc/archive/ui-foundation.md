# RFC: UI Foundation (primitives, not screens)

- **Status:** implemented — C1–C11 are shipped and independently reviewed.
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-03
- **Design refs:** `design/06` (Svelte 5 runes, DOM-first, per-tab `$derived`, 10 Hz formatting), `design/08` (era presentation: the UI itself ages through gaming eras — the load-bearing satire surface), `design/11` (voice, first-session), `design/03 §9` (arcade era-skins)
- **Depends on:** Client Shell, Transport wire v2, Copy Pipeline (implemented)
- **Owner ruling honored:** breadth-first — this RFC ships ZERO screens; it ships the system every future screen is made of, so screens never rebuild when foundations shift.
- **Planning:** `planning/archive/ui-foundation/`

## Summary

Every future surface — desk, boards, commons panel, minigames, pet room — needs the same
primitives: theming that can age through eras, exact-number rendering, copy resolution, a
navigation shell, and the wire-only component law. Build them once, content-free.

## Specification

### UF1 — The era-theming system (the satire's rendering layer)

**RULED (C1, C2): Phase-A ships `era_1995` and `era_2000`** (Tiers 0–1 per design/08 §2's Web-1.0
assignment — my `era_2006`/glossy was wrong, corrected). A closed theme contract lives in
`ui/themes.schema.json` + one JSON artifact per era, root `{schema_version:1, era, color, type,
space, border, chrome, motion}`, every child a closed enumerated key set shared by all eras,
every value a per-key-domain-validated CSS token string; the loader rejects missing/extra keys.
Tokens install as `--cc-*` custom properties on one `data-era` shell root; components reference
only those. The literal-style lint covers component style blocks + inline style attrs, permits
layout literals and named-property geometry, rejects literal colors/fonts/shadows/radii/durations.
Era switch changes only `data-era` + token values; DOM unchanged. Components consume TOKENS ONLY
(a lint boundary forbids literal colors/fonts in components); an era switch restyles the world
without touching a component. Era assignment is driven by game state (tier) later; Phase-0
foundation ships the contract + two eras + the switch mechanism proven by test.

### UF2 — Number & notation rendering

**RULED (C3):** `<Amount>` pins the `@antimatter-dimensions/notations` **Standard** notation class
at a pinned package version, precision/threshold as RFC literals; props exactly `{value:
canonical string, cap?: {amount: canonical string, reason_key: CopyKey}}`; parse via the numeric
boundary, locale-independent; first value + changed cap render synchronously, ordinary changes
coalesce on one shared 100 ms scheduler, unmount cancels; the cap explanation is RESOLVED COPY,
never the raw key. Golden vectors cover zero, notation thresholds, exponent boundaries, rounding
carry, capped-at-cap, and sub-visible ongoing production.

### UF3 — Copy resolution

**RULED (C4): Copy Pipeline Foundation is a hard dependency.** UF3 consumes its implemented
`copy_catalog` interface — `t(key, typed_params, era)` with the per-entry
`era_variants[era]` → base-text fallback, missing-key loud-in-dev + CI completeness gate, and
provenance verified by the Copy Pipeline. There is no second `key@era` grammar.


### UF4 — Navigation & surface shell

A closed surface registry row is `{surface_id, mount_id, unlock:{kind:"always"}|
{kind:"fact_equals",fact_id,value}}`. Unlocks reference committed shell-state facts only;
unknown facts reject against the authoritative snapshot manifest. Rows order bytewise by
`surface_id`; duplicate IDs and mounts reject. The shell owns routing and exactly one active
surface subscription handle, disposing it before lazily mounting the next surface. Inactive
surfaces are unmounted and receive zero callbacks. Theme tokens style every surface; there is no
per-surface `era_skins` path.

### UF5 — The component law + harness

**RULED (C6): "wire-only" means DATA inputs, not all inputs.** Components legitimately take:
event callbacks, local view state, a11y labels/ARIA, token custom-property references, and copy
keys. The law forbids exactly: reading transport internals, computing authoritative game values
client-side, and string literals for player-facing copy or theme values. The existing
`App.svelte` Phase-0 screen copy migrates to copy keys as part of this RFC (recorded, not a
surprise). **RULED (C5): surface unlocks reference shell-state facts** (crossed gates, tier,
membership — the same fact vocabulary the shell already holds), evaluated by the shell, never
UI-invented logic; the `ShellTab` union grows to the registered surface set.

Wire-only DATA inputs (decoded envelopes, shell state, copy keys) enforced by import-boundary lint;
a repository-owned Vite fixture route renders every primitive against golden fixtures per era in
the existing Vitest-browser/Playwright suite. Storybook and any second UI build surface are
forbidden.
Accessibility baseline: keyboard operability and reduced-motion at the PRIMITIVE level so every
future screen inherits it.

## Acceptance criteria

1. Era switch: one state change restyles the full primitive set; byte-stable DOM-structure
   snapshots per era; the literal-style lint fails a seeded violation.
2. `<Amount>`: golden rendering vectors both eras; cap state renders reason_key; throttling
   proven (≤10 updates/s under a 20 Hz feed).
3. Copy: missing-key CI failure; era-fallback chain test; provenance-tag lint fails a seeded
   unverified statistic.
4. Surface registry: register/unlock/mount lifecycle test; source subscribe/unsubscribe counts and
   zero inactive callbacks prove that background surfaces do not retain subscriptions.
5. Harness runs in CI as part of the browser suite; a11y baseline asserted on every primitive.

## Open questions

- Third+ era token sets — content work on the UF1 contract, arrives with tiers.
- Exact C9–C11 literals below; Copy Pipeline itself is implemented and no longer open.

## Post-ruling implementation blockers (Codex review, 2026-08-03)

### C9 — The closed theme shape still has no keys or era values

C2 says every theme child has an enumerated shared key set, but neither the ruling nor normative
text enumerates a single key. The two required era artifacts also have no literal values. A loader,
CSS variable generator, governed lint, and DOM fixture cannot agree on a contract that has no
tokens.

**Proposed contract:** owner supplies or accepts one exact token matrix for `color`, `type`,
`space`, `border`, `chrome`, and `motion`, including the literal `era_1995` and `era_2000` values.
Code maps each leaf to `--cc-{group}-{key}` and rejects every other leaf. Layout-only custom
properties live in a separately named allowlist and are never theme tokens.

### C10 — Amount's claimed pinned formatter has no pin or options

C3 requires a pinned `@antimatter-dimensions/notations` Standard class, literal precision and
threshold, and sign behavior, but supplies none of the version or literals. The package is not in
the lockfile. Golden vectors would bless an implementation choice rather than test a contract.

**Proposed contract:** name the exact package version, Standard constructor/options, significant
digits/places, standard-to-scientific threshold, and whether negative canonical values are valid
Amount props. Pin them in this RFC and lockfile; record expected strings for the mandatory vectors
before component implementation.

### C11 — The accessibility engine is still unnamed

C7 says “one pinned accessibility engine” but does not select one or a version. That decision
changes dependencies, browser setup, violation taxonomy, and the meaning of serious/critical.

**Proposed contract:** select `axe-core` (or name the alternative) at an exact version, invoke it
inside the existing Vitest-browser fixture route with WCAG 2.2 AA tags, and define the allowed
impact set as zero `serious|critical` violations. Keyboard/focus/name/motion assertions remain
separate and mandatory.

## Acceptance blockers (Codex review, 2026-08-03)

The architecture is sound, but the draft is not yet executable without inventing presentation
and data contracts. The following closures are required before acceptance. Each includes a
proposed contract so the owner can rule rather than re-derive the implementation boundary.

### C1 — The second era contradicts the binding era table

UF1 names `era_2006` as glossy Web 2.0. `design/08 §2` assigns Tier 1 to **2000s Web 1.0**
(gradients, badges, guestbook), while glossy startup chrome does not arrive until a later tier.
This must not become an unrecorded design deviation.

**Proposed contract:** Phase A ships exactly `era_1995` and `era_2000`, corresponding to Tiers 0
and 1 in `design/08 §2`. Theme IDs are stable mechanical identifiers; later eras append by RFC.
If glossy Web 2.0 is intentional instead, add a Deviations from design section and state which
tier owns it.

### C2 — The theme contract has no closed wire or token shape

“Type scale, spacing, color roles, border/chrome style, motion budget” does not define the keys,
value domains, fallback behavior, or which literals the lint permits. Two implementations could
both satisfy that sentence while exposing incompatible component APIs.

**Proposed contract:** check in a strict `ui/themes.schema.json` and one JSON artifact per era.
The root is `{schema_version:1, era, color, type, space, border, chrome, motion}`; each child has
a closed, enumerated key set shared by all eras, and every value is a CSS token string validated
against a per-key domain. The loader rejects missing/extra keys. It installs tokens as
`--cc-*` custom properties on one shell root carrying `data-era`; components may reference only
those properties. The literal-style lint covers Svelte component style blocks and inline style
attributes, permits layout literals and dynamic geometry through named custom properties, and
rejects literal colors, font families, shadows, radii, and animation durations. Era switching
changes only `data-era` and token values; DOM is unchanged.

### C3 — `<Amount>` does not specify a notation

The named dependency is a collection of notation implementations, not one output grammar.
Precision, threshold, sign/zero behavior, locale, immediate-vs-throttled updates, and the cap
explanation interface are all unspecified, so the required golden vectors have no oracle.

**Proposed contract:** select one named notation class and pin its package version; define its
precision/threshold options as literals in the RFC. `<Amount>` accepts the exact props
`{value: canonical string, cap?: {amount: canonical string, reason_key: CopyKey}}`. Parsing uses
the existing numeric boundary; formatting is locale-independent. The first value and a changed
cap state render synchronously, ordinary value changes coalesce on one shared 100 ms scheduler,
and unmount cancels pending work. The cap explanation is resolved copy, never the raw key. Golden
vectors include zero, signs if permitted by the prop, notation thresholds, exponent boundaries,
rounding carry, capped-at-cap, and sub-visible ongoing production.

### C4 — UF3 depends on an interface that does not exist

Copy Pipeline is neither a dependency nor an indexed RFC, and the draft deliberately leaves its
artifact format open while making typed params, era fallback, provenance, completeness, and two
acceptance tests depend on it. UF3 therefore cannot be implemented or tested as written.

**Proposed contract:** either (a) make an accepted Copy Pipeline RFC a hard dependency and move
the exact `CopyKey`, parameter-schema, fallback, and provenance interfaces here by reference, or
(b) split UF3 and AC3 into that successor so this RFC can ship independently. Do not create a
temporary second copy grammar. The selected owner must also define whether missing parameters,
extra parameters, missing default copy, and missing era variants are load errors or runtime
fallbacks.

### C5 — Surface unlocks reference no authority

No implemented catalog declares UI unlock facts, and `unlock_condition_ref` has no referent or
evaluation grammar. The current shell has a closed `ShellTab` union and one global subscription;
it cannot register arbitrary surfaces or prove a compiler-internal `$derived` count.

**Proposed contract:** a surface row is
`{surface_id, mount_id, unlock:{kind:"always"}|{kind:"fact_equals", fact_id, value}}` over the
shell's committed `discrete` facts only. Unknown fact IDs are rejected by a build-time manifest
composed from the authoritative snapshot contract. Remove `era_skins`—tokens already style every
surface. Registration order is byte-order by `surface_id`; duplicate IDs/mounts reject. The shell
owns one active-surface subscription handle, disposes it before mounting the next, and inactive
surfaces are unmounted. AC4 asserts source subscribe/unsubscribe counts and zero inactive
callbacks, not an unobservable `$derived` implementation detail.

### C6 — “Wire-only” and “zero strings” currently outlaw required component inputs

Components need event callbacks, local view state, accessibility labels, token imports, and copy
keys; none are wire envelopes. The existing `App.svelte` also contains Phase-0 screen copy and
literal styling, so an unscoped lint either fails HEAD immediately or silently exempts the code
that future screens would copy.

**Proposed contract:** the component import boundary permits only `shell/contracts`, UI
foundation modules, generated copy-key types, and explicitly passed callbacks; it forbids
transport, production/economy kernels, raw fetch/WebSocket access, and balance-file imports.
“Zero string literals” means zero **player-facing prose** in governed components; mechanical IDs,
ARIA role names, data attributes, and test IDs are allowed by a narrow AST rule. This RFC migrates
the existing shell scaffold into the governed boundary or deletes it—there is no permanent
legacy exemption.

### C7 — The harness and accessibility gate are left to incompatible implementations

Storybook and a bespoke fixture page create different dependencies, build artifacts, and CI
surfaces. “Accessibility baseline” has no standard, test engine, or required assertions.

**Proposed contract:** use a repository-owned Vite fixture route inside the existing client and
Vitest-browser/Playwright suite; do not add Storybook. Each primitive is rendered in every shipped
era from checked-in fixtures. Add one pinned accessibility engine and fail on its serious/critical
violations against WCAG 2.2 AA; separately assert keyboard reachability/activation, visible focus,
semantic names, and that `prefers-reduced-motion: reduce` reduces every nonessential token duration
to zero. The browser suite remains the single CI entrypoint.

**RULED (C7): accepted as proposed** — Vite fixture route inside the existing client + the
existing Vitest-browser suite (NO Storybook — no new build surface); one pinned a11y engine,
fail on serious/critical WCAG 2.2 AA, plus the four explicit assertions (keyboard reach/activate,
visible focus, semantic names, reduced-motion → zero nonessential duration). The browser suite
stays the single CI entrypoint.

### C8 — Ownership and dependency order must be reconciled

The RFC says it ships zero screens, but it necessarily replaces the existing shell screen scaffold
to enforce C6, while Game UI and Copy Pipeline both depend on its types. Its current dependency
list does not reflect that order.

**Proposed contract:** this RFC owns primitives, the surface host, governed lint boundaries, and
the migration of `App.svelte` to a content-free fixture host; it owns no playable screen layouts.
Declare Copy Pipeline according to the C4 ruling, then order UI Foundation → T0–T1/Game UI. Name
the exact file/package boundary exported to those successors.

**RULED (C8): accepted** — this RFC owns primitives, the surface host, the governed lint
boundaries, and the `App.svelte` → content-free fixture-host migration; it owns ZERO playable
screen layouts (those are Game-UI/T0–T1 content on the exported types). Dependency order:
**Copy Pipeline → UI Foundation → {Game UI, T0–T1}**. The exported boundary is
`client/src/ui/` (primitives + surface registry + theme/copy resolution), the ONLY module
Game-UI/T0–T1 import from.

## Changelog

- 2026-08-03: created (draft) — the breadth-first UI foundation; zero screens by owner ruling.
- 2026-08-03: C7–C8 ruled — Vite fixture route + Vitest-browser (no Storybook), pinned a11y engine failing on serious/critical WCAG 2.2 AA + keyboard/focus/reduced-motion assertions; UI Foundation owns the App.svelte→fixture-host migration and exports the primitive/surface types to Game-UI/T0–T1, ordered after Copy Pipeline. Fully accepted.
- 2026-08-03: C1–C6 ruled — era_1995/era_2000 (design/08-correct), schema'd token contract, pinned Standard notation, Copy Pipeline dependency named, wire-only clarified to data-inputs, unlocks reference shell facts. Accepted.
- 2026-08-03: C1–C8 reconciled. Codex follow-up found C9–C11: the “closed/pinned” theme,
  formatter, and accessibility contracts omit their actual token/value/version literals.
- 2026-08-06: non-normative reference cleanup for publication; no spec change.
- 2026-08-10: implementation pins `@antimatter-dimensions/notations` **1.6.0** and `axe-core`
  **4.12.1** in the client lockfile; the exact C9 token artifacts, Amount goldens, and WCAG 2.2 AA
  browser gate now ship behind the content-free fixture host.

## Owner rulings on C9-C11 (2026-08-07) — the UI Foundation is implementable

- **C9 — the exact token matrix (literal values are PROVISIONAL styling data, tunable like balance;
  the KEY SET is the closed contract).** Every era artifact carries exactly these keys, mapped to
  `--cc-{group}-{key}`; the loader rejects any other leaf; layout-only custom properties live in a
  separate named allowlist.
  - `color`: `bg, surface, text, text_muted, accent, accent_text, border, link, danger, success`
  - `type`: `font_ui, font_display, font_mono, size_base, size_small, size_large, size_display,
    weight_normal, weight_bold, line_height`
  - `space`: `unit, xs, sm, md, lg, xl`
  - `border`: `width, radius, style, bevel`
  - `chrome`: `window_bg, window_border, titlebar_bg, titlebar_text, button_face, button_shadow`
  - `motion`: `duration_fast, duration_base, duration_slow, easing, budget`
  **`era_1995` literals (honest shareware — Tier 0):** color: `#c0c0c0, #ffffff, #000000, #555555,
  #000080, #ffffff, #808080, #0000ee, #aa0000, #006600`; type: `'Tahoma, "MS Sans Serif", Geneva,
  sans-serif'` (ui + display), `'"Courier New", monospace'`, `13px, 11px, 16px, 20px, 400, 700,
  1.35`; space: `4px, 2px, 4px, 8px, 16px, 24px`; border: `2px, 0px, solid, outset`; chrome:
  `#c0c0c0, #000000, #000080, #ffffff, #c0c0c0, #808080`; motion: `0ms, 0ms, 0ms, linear, none` —
  **the 1995 motion budget is literally zero** (nothing animated in shareware; that is the era's
  honesty and the joke).
  **`era_2000` literals (Web 1.0 — Tier 1, per design/08 §2 and the C1 ruling):** color: `#ffffff,
  #f0f0fa, #000000, #333366, #3366cc, #ffffff, #9999cc, #0000cc, #cc0000, #009900`; type:
  `'Verdana, Arial, Helvetica, sans-serif'` (ui + display), `'"Courier New", monospace'`, `12px,
  10px, 15px, 18px, 400, 700, 1.4`; space: same grid as 1995; border: `1px, 4px, solid, none`;
  chrome: `linear-gradient(180deg, #6699ff, #3366cc)` (window_bg), `#336, #3366cc, #ffffff,
  #e8eefc, #9999cc`; motion: `100ms, 200ms, 400ms, ease-in-out, respect`.
- **C10 — the Amount formatter pin.** `@antimatter-dimensions/notations`, **Standard notation**,
  pinned to the newest 1.x resolvable at implementation with the exact version RECORDED in this
  RFC's changelog AND the lockfile in the same implementing commit (a bounded delegation — the
  version literal is recorded, not chosen ad hoc). Options: 3 significant mantissa digits
  (`1.23 Qa`); values < 1000 render as plain integers (0 decimal places); **no
  standard-to-scientific threshold in Phase A** (Standard throughout; scientific is a later user
  option, not a fallback). Negative canonical values ARE valid Amount props (deltas exist) and
  render with a leading `−`. The mandatory golden vectors record expected strings BEFORE component
  implementation.
- **C11 — the accessibility engine.** **`axe-core`**, exact version recorded the same way as C10,
  invoked inside the existing Vitest-browser fixture route with **WCAG 2.2 AA tags**; the allowed
  impact set is **zero `serious|critical` violations**. Keyboard/focus/name/motion assertions remain
  separate and mandatory (the engine augments, never replaces them).
