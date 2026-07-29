# Leaderboards & Balance Epochs — append-only log

## 2026-07-29 — start

- Accepted through `planning/codex-batch-2026-07-29.md` after L1–L8 answered the implementation
  review. Implementation starts with L1 because Prestige P7 deliberately emits a provisional
  terminal revision until the transaction-local run-log sequence exists.
- The run log is persistence infrastructure, not a derived analytics feed: canonical request bytes,
  normalized receipts, applied revision, and server time commit with the gameplay mutation.
- Replay/version/catalog storage follows before board projection so no unverified row can enter a
  ranking table merely because a terminal event exists.
