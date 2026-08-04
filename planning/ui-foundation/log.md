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
