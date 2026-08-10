# UI Foundation implementation log

## 2026-08-03 — post-ruling acceptance pass

C1–C8 were owner-ruled, and Copy Pipeline is now implemented. The follow-up found three literal
contracts still absent despite the draft describing them as closed/pinned: no theme token key or
era-value matrix, no notation package version/options/sign behavior/golden strings, and no named
a11y engine/version. C9–C11 record the narrow owner decisions needed; implementation does not
invent those presentation/dependency choices.

UF3 was reconciled directly to the implemented Copy Pipeline: callers pass an era to `t`, which
uses `era_variants[era]` and falls back to base text. The stale `key@era` grammar was removed rather
than creating a second copy system.

## 2026-08-04 — independent specification follow-up (`87f542d..24203ee`)

- **Review by:** Darwin
- **Recorded by:** Darwin
- **Decision:** **implementation remains correctly blocked, but the active RFC is not yet singular.**
  C9–C11 are real, explicitly open owner decisions and no UI implementation was improvised. The
  claimed C1–C8 reconciliation is incomplete, however.

**HIGH — stale normative C1–C8 prose still contradicts the ruled contract.** In the active
specification, UF3 repeats the withdrawn `key@era_2006` grammar immediately after saying no second
grammar exists; UF4 still declares `era_skins`/`unlock_condition_ref` after C5 removes `era_skins`
and defines the closed shell-fact condition; UF5 still allows Storybook after C7 rules it out; and
AC4 still requires an internal `$derived` count after C5 replaces that with observable subscription
counts. Reconcile those decision sites before ruling C9–C11 so resolving the three literal blockers
does not expose a second contradictory implementation path.

What held: the status, plan, and log correctly block all implementation on the missing theme token
matrix/era values, notation package version/options/golden strings, and accessibility engine/version.
No dependency or presentation choice was invented in this range.

## 2026-08-04 — ruled-text reconciliation

Removed the four stale normative paths identified by independent review: the withdrawn
`key@era_2006` grammar, `era_skins`/untyped unlock references, the Storybook option, and the
unobservable `$derived` acceptance assertion. UF3–UF5 and AC4 now state only the C1–C8 rulings.
C9–C11 remain the sole implementation blockers; no presentation literal or dependency choice was
invented while reconciling existing decisions.

## 2026-08-04 — independent ruled-text reconciliation review (`402ba20..85cbea6`)

- **Review by:** Darwin
- **Recorded by:** Darwin
- **Decision:** **approved as a specification-only reconciliation; implementation remains correctly
  blocked on C9–C11.**

UF3 now has only the Copy Pipeline's typed `t(key, params, era)` plus `era_variants` fallback; the
withdrawn `key@era_2006` grammar is absent from the active decision site. UF4 owns the ruled closed
surface/unlock row, byte ordering, one observable subscription handle, and token-only styling with
no `era_skins` implementation path. UF5 requires the repository-owned Vite fixture in the existing
browser suite and expressly forbids Storybook. AC4 asserts observable subscribe/unsubscribe and
inactive-callback behavior rather than compiler-internal `$derived` counts.

C9's literal token matrix/era values, C10's notation package pin/options/golden strings, and C11's
accessibility engine/version remain explicit unchecked plan items and the RFC status still blocks
implementation. No presentation constant, dependency, or playable screen was invented in this
range. Exact-range diff checking and the fresh root `make verify` both pass.

## 2026-08-07 — owner rulings C9-C11 (the last implementation blockers)
C9: the closed token key-set (color/type/space/border/chrome/motion, 41 keys) + literal era_1995 and
era_2000 matrices (values provisional/tunable; the KEY SET is the contract; the 1995 motion budget is
literally zero). C10: @antimatter-dimensions/notations Standard, 3 sig digits, plain ints <1000, no
scientific threshold in Phase A, negatives valid; exact version recorded in changelog+lockfile at the
implementing commit. C11: axe-core (same version-recording duty), WCAG 2.2 AA, zero serious|critical;
keyboard/focus/name/motion assertions stay separate. Status -> accepted; implementing. UI Foundation
is now the playability critical path together with the Minigame API & Surface amendment.

## 2026-08-10 — implementation handoff preparation

The C9–C11 literals now have executable owners: two strict 41-token theme artifacts and loader,
Standard notation 1.6.0 with canonical Decimal goldens and the shared 100-ms Amount scheduler, and
axe-core 4.12.1 in the existing three-browser fixture route. The surface registry rejects unknown
shell facts and proves unsubscribe-before-mount plus zero inactive callbacks. `App.svelte` is now a
content-free primitive fixture; it invents no screen or authoritative game value.

Focused typecheck, client tests, schema/boundary gates, and the full browser matrix are green. This
entry is a self-check record only: independent designated review and archival remain open.
