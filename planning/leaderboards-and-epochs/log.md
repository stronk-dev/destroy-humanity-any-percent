# Leaderboards & Balance Epochs — append-only log

## 2026-07-29 — start

- Accepted through `planning/codex-batch-2026-07-29.md` after L1–L8 answered the implementation
  review. Implementation starts with L1 because Prestige P7 deliberately emits a provisional
  terminal revision until the transaction-local run-log sequence exists.
- The run log is persistence infrastructure, not a derived analytics feed: canonical request bytes,
  normalized receipts, applied revision, and server time commit with the gameplay mutation.
- Replay/version/catalog storage follows before board projection so no unverified row can enter a
  ranking table merely because a terminal event exists.

## 2026-07-29 — atomic run log

- Migration 00012 adds one immutable sequence per `(Company stream, run_seq)`. Applied and terminal
  rejected Company intents persist the exact canonical bytes covered by `request_hash`, normalized
  receipt, nullable applied revision, and database server milliseconds in the gameplay transaction.
  Revision conflicts and idempotent replays do not create sequence gaps or duplicate rows.
- `ApplyIntentLogged` and `ApplyExitTransactionLogged` preserve the existing unlogged store API for
  non-gameplay tests/callers while making production's logging obligation explicit and hash-checked.
  Founder-scope hint purchases are career commands rather than Company-run transitions and do not
  enter a run replay.
- The transaction allocates the Exit intent's log sequence before mutation and exposes it only on
  the locked Company revision. Prestige now writes that exact value into `run_ended.terminal_seq`;
  the old provisional save revision is gone.
- Real-Postgres coverage proves ordered applied/rejected logging, byte-identical canonical payloads,
  replay non-duplication, terminal-sequence equality, and rollback when a fault is injected after
  the run-log insert but before the rest of the Exit commit.
