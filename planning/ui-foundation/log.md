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
