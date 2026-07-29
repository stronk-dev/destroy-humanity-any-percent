# Balance Harness Foundation — implementation plan

- **RFC:** `rfc/balance-harness-foundation.md`
- **Assignee:** Codex
- **Started:** 2026-07-29

## Sequence

1. Extract the exported deterministic production transition shared by service and harness.
2. Implement deterministic run identity, SplitMix64, UUIDv7 generation, policies, runner, and
   invariant/milestone reporting.
3. Add scenario/report schemas, Phase-0 scenario, CLI, golden report, and pacing baseline.
4. Add drift/baseline guards and Make/CI integration.
5. Verify determinism, pacing, chaos, read-only checks, and full suite; review and archive.

## Completion

All five steps and all seven RFC acceptance criteria completed on 2026-07-29.
