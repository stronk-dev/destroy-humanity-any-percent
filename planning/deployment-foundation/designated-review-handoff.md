# Deployment Foundation designated-review handoff

**Status:** review input, not a verdict or release claim. **Prepared:** 2026-09-23.
The accepted authority is `rfc/deployment-foundation.md`; implementation gates and batch
boundaries are in `planning/deployment-foundation/plan.md` and the append-only `log.md`.
Codex's first-filter verdicts in that log are not the required cross-party review.

## Bounded review now available

Claude can inspect the following contiguous, disjoint ranges and record a separate exact-range
adversarial verdict for each batch. Git ranges below are `base..tip` (base excluded, tip included).
The final record commit is included in each range, so scope and claimed evidence are reviewable
alongside product bytes. A single review session may cover the six ranges; the log must still name
which exact range each verdict approves or rejects.

| Batch | Exact range | Claimed boundary; inspect the log before testing |
|---|---|---|
| DP-A | `cd102d7..d7d443f` | Production config, secret files, one-hop origin/proxy, key composition; no rotation or release claim. |
| DP-B | `d7d443f..9b916c2` | Self-contained bundle and digest/SBOM/attribution checks; no deployed runtime claim. |
| DP-C | `9b916c2..ab70327` | Encrypted backup/restore and retention, with real Postgres negatives; no RPO/RTO observation. |
| DP-D | `ab70327..3f58fea` | Release/drain/rollback state machine and real Caddy/WebSocket witnesses; no final clean-host claim. |
| DP-E | `3f58fea..65099e7` | Private operations profile and seven alert paths; see the later corrective range before a final DP-E conclusion. |
| DP-E/DP8 corrective | `7b510df..cf4ac25` | WebSocket instrumentation repair, pushed kernel-history correction, and local/hosted push-CI topology parity. |

The corrective range is deliberately separate: DP-F construction commits lie between DP-E and
the corrective range, so a broad `65099e7..cf4ac25` citation would claim uninspected DP-F work.
The DP-F construction, later follow-ups, and exact R-006 run need their own review coverage; this
handoff does not approve them.

## Adversarial protocol

1. Inspect `git diff --stat` and `git diff` for each exact range, the accepted RFC's corresponding
   criteria, and the batch's predeclaration/first-filter section in `log.md`. Do not infer scope
   from this short table alone.
2. Run the relevant repository-root gates cold. The available lanes are
   `make test-deployment-backup`, `make test-deployment-release`,
   `make test-deployment-operations`, `make test-deployment-rehearsal`, and
   `make verify-push`. Postgres runs use the repository's declared Docker service.
   Report any gate not run; a cited old pass is not a fresh execution.
3. For at least one load-bearing outcome per range, independently check that its witness fails
   when the required behavior is severed. Use an isolated worktree or an exactly restored
   scratch mutation; never leave a probe in the implementation tree. Check that a passing test
   did not skip its required dependency.
4. Record `Review by: Claude`, `Recorded by: ...`, exact range, tests/probes actually executed,
   findings, and APPROVED or CHANGES REQUIRED in the append-only log. A recorder must not turn
   Codex's first filter into the designated verdict. Do not archive or promote 1.0 status here.

## Unmet gates after these reviews

DP-F3 still requires an authorized clean Linux/amd64 host and a run against the exact
candidate/previous bundle pair. The local DP-F2 record reports byte-matched builds and 25/43
real rehearsal probes, not a clean-host R-006 result. DP-F4, full implementation-range review,
RFC archival, and the owner's release call follow only after the remaining AC1–AC10 evidence.
