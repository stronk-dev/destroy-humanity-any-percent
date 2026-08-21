# Prestige & Exits — implementation plan

- **Assignee:** Codex
- **RFC:** `rfc/prestige-and-exits.md`
- **Started:** 2026-07-29
- **Current closeout:** D2–D5/P3–P5 and canonical docs reconciled to current behavior and D-012;
  literal AC2–AC6 witness remediation is the next authorized implementation batch.

1. [x] Add save v7 Company/Founder state, validation, migrations, and corpus fixtures.
2. [x] Implement cross-runtime prestige arithmetic and declarative Phase-0 policy.
3. [x] Implement deterministic offers, timers, decline, wind-down, accept, and gated IPO intents.
4. [x] Implement atomic two-stream Exit persistence and exact event schemas.
5. [x] Add real-Postgres fault, replay, and concurrency tests plus arithmetic properties.
6. [ ] Reconcile the already-implemented first-elective-Exit harness evidence in the same
   designated-review range as this checkbox flip; do not infer completion from the older audit.
7. [x] Update canonical docs and run focused verification.
8. [x] Replace the provisional terminal revision with Leaderboards L1's atomic run-log sequence.
9. [ ] Record independent review before archival.
10. [x] Add the missing literal AC2–AC6 witnesses: offer-age/progress property, full-domain reseed
    and repeated non-empty ledger preservation, New-Founder lifecycle, eligible-state/event-chain
    matrix, and checked-in run-2 golden. Each witness needs a demonstrated failing mutation.
