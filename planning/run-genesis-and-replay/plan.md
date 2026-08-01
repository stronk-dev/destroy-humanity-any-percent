# Run Genesis & Replay Verification — implementation plan

- **Assignee:** Codex
- **RFC:** `rfc/run-genesis-and-replay.md`
- **Started:** 2026-07-31

1. [x] Accept an executable RA/RB contract: close every live transition input, terminal Exit
   dependency, hook-order rule, and legacy-row migration semantic before creating the immutable
   replay schema.
2. [x] Add the versioned replay-input wire object and run-log persistence, then refactor the live
   Go engine so every logged mutation uses `ApplyLogged` and persists the exact consumed inputs in
   the same transaction.
3. [x] Port `ApplyLogged` to the TypeScript verification kernel and land shared full-transition
   parity fixtures.
4. [x] Add immutable run genesis storage at account creation, import, and Exit run-start sites,
   with pin/genesis atomicity and byte-identity proofs.
5. [ ] Implement the six-verdict verifier and shared failure corpus, including pre-timer board
   behavior and terminal-fact checks.
6. [ ] Add the verification queue/dead-letter/project-only path, archive compaction, canonical
   docs, full verification, and independent review before archival.

Acceptance gates are the RFC's six criteria. No migration or transition refactor lands while the
closed RA/RB shape cannot represent inputs the current live code demonstrably consumes.
