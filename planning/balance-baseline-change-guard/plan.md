# Balance Baseline Change Guard Hardening — implementation plan

- **Assignee:** Codex
- **RFC:** `rfc/balance-baseline-change-guard.md`
- **Started:** 2026-07-29

## Work breakdown

1. [ ] Replace HEAD-only inspection with full baseline-history validation.
2. [ ] Enforce artifact-only commits, prior input changes, and clean generated artifacts.
3. [ ] Add real temporary-repository regressions for cover, shallow, smuggling, and valid flows.
4. [ ] Fetch complete server history in CI; update canonical docs; run full verification.
5. [ ] Record independent review and archive.
