# Leaderboards & Balance Epochs — implementation plan

- **Assignee:** Codex
- **RFC:** `rfc/leaderboards-and-epochs.md`
- **Started:** 2026-07-29

1. [x] Add the atomic per-run intent log and bind Prestige terminal sequences.
2. [x] Add immutable catalog artifacts, kernel identity, epoch mint/hotfix storage, and run pinning.
3. [ ] Define the missing immutable initial-state input, then implement archive compaction, replay verification, and shared Go/TypeScript verdict fixtures.
4. [x] Finish category catalog content; verified-run projection, competition ranking, cursors, imported exclusion, and world-first arbitration are implemented.
5. [x] Extend the hardened balance-history guard for epoch/hotfix and cap-lowering rules.
6. [ ] Complete canonical docs and full integration/verification coverage.
7. [ ] Record independent review before archival.
