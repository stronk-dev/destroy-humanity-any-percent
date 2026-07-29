# Commons Compact — implementation plan

- **Assignee:** Codex
- **RFC:** `rfc/commons-compact.md`
- **Started:** 2026-07-29

## Work breakdown

1. Bind the strict Commons catalog, fixed-point/Decimal formula boundary, Go/TypeScript parity fixtures, and production import guard.
2. Add save v6 membership/Solidarity state and authoritative sign/leave intent/event contracts.
3. Add idempotent Postgres membership, cohort, Health/Capacity, merge, recruitment, and query projections.
4. Wire the neutral Commons contribution into the existing multiplier slot and publish generated formulas.
5. Add the deterministic 200-vs-20,000 population harness gate and full integration coverage.
6. Update canonical docs, run complete verification, review the entire diff, and archive the RFC/planning record.

## Acceptance gates

- Shared Go/TypeScript Enclosure and buff vectors agree at every boundary.
- Sign/leave/re-sign are idempotent and leaving clears Solidarity.
- Cohort assignment is deterministic, concurrency-safe, persistent, and merge-safe.
- Non-members supply no Commons slot; import boundaries are build-enforced.
- Population invariance passes at 200 and 20,000 simulated members.
- Formula artifact and source-weight table are generated from shipped data.
- `make verify` passes with Postgres integration enabled.
