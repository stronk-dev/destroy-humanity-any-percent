# Faction & Incorporation — implementation plan

- **Assignee:** Codex
- **RFC:** `rfc/archive/faction-incorporation.md`
- **Started:** 2026-07-30

1. [x] Add the strict four-faction catalog, Commons cross-validation, epoch identity, schema gate,
   and Hamiltonian-cycle negative fixtures.
2. [x] Add the next save version with run-scoped faction identity, incorporation time, stock
   counters/remainder, migration corpus, and wire snapshot fields.
3. [x] Add the `incorporate` intent and `incorporated` event with C1 parsing, typed rejections,
   idempotent persistence, Open Source compact binding, and bound-leave rejection.
4. [x] Integrate attended-time stock accrual and saturation into the existing evaluation path,
   with exact remainder/offline/cap properties.
5. [x] Extend leaderboard structural variables and run reset/terminal facts.
6. [x] Update canonical docs and generated/schema artifacts; run focused Go/client suites, full
   Compose/Postgres integration, vet, formula/epoch guards, and independent review before archive.
