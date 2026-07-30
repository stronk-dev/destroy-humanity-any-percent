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

## 2026-07-29 — epoch identity and board storage

- Added `kernel/VERSION` as the transition-semantics source of truth with fail-closed Go/TypeScript
  parity checks. Run pins record this semver, build VCS detail, the accepted constants hash, and the
  deterministic unsigned Founder/run seed.
- Catalog sets/artifacts, accepted hash sets, closed/current epochs, and run pins are append-only
  evidence. Database triggers reject historical updates/deletes. Mint closes the current epoch and
  inserts the next epoch plus artifacts atomically; hotfix adds an accepted immutable hash.
- Account creation pins run 1 inside its stream-creation transaction. Logged Company commands refuse
  an absent/mismatched pin. Prestige pins run N+1 under a shared current-epoch lock in the same Exit
  transaction as `run_started`, so a concurrent mint produces exactly one epoch assignment.
- Implemented immutable verified-run projection, imported-Founder rejection before claim, separate
  Commons/Advisor/Glitched variables, atomic partial-unique world-first arbitration, time and count
  boards, competition ties, and keyset cursors. Real PostgreSQL asserts `1,1,3`, page continuity,
  old-epoch queryability, and artifact immutability.
- Migration review caught an attempted edit to already-applied migration 00013 while board tables
  were still uncommitted. Board storage moved to append-only migration 00014; upgraded and fresh
  databases now converge through the same history.
- **DESIGN-GAP (blocks L2/L4 and final L1 archival):** `run_log` preserves canonical commands and
  receipts, but neither L1 nor `run_started` preserves the immutable initial Company state (or a
  sufficient Founder snapshot/reference). Run 2 includes Network-carried items and Reputation
  starter effects, while historical save revisions are pruned. A verifier cannot reconstruct its
  initial state from catalog bytes and seed alone. Do not invent an oracle or claim replay parity;
  the RFC needs an executable initial-state/archive contract.
- **DESIGN-GAP (blocks L7 catalog completion):** the RFC requires four literal canonical category
  rows, but only names Any%/100%/Ethical%/Low%; the exact terminal predicates for the latter three
  are not enumerated in accepted design. The closed predicate loader can land after those content
  contracts are supplied; no plausible-sounding terminal rules were improvised.
- Full `make verify` is green with the local PostgreSQL integration path: Go vet/tests, formula and
  harness gates, strict TS/Svelte, production build, 6,422 client tests, schema/boundary checks, and
  19,275 Chromium/Firefox/WebKit cases. The preceding save-v7 state-hash drift was regenerated and
  committed separately under `BALANCE-CHANGE:`; pacing observations and balances did not move.

## 2026-07-30 — repository epoch history guard (L8)

- Added `balance/epochs/phase0.json` as the repository source for the current epoch, its immutable
  predecessors, the named exact artifact bundle, and each epoch's sorted accepted-hash set. The
  first committed seed is the enforcement boundary; later changes are walked across complete Git
  history rather than trusting only `HEAD` or CI metadata.
- Any later commit touching a declared constants artifact must update the seed in the same commit.
  An ordinary commit may append only the exact resulting hash to the current epoch. A
  `BALANCE-CHANGE:` commit must append exactly one epoch and add its numbered changelog in the same
  commit. Seed-only registration and artifact-identity changes fail closed.
- The guard loads the parent and resulting economy catalogs through the production parser and
  compares every declared hardcap using the numeric core. A lower cap is rejected as a hotfix; a
  mint additionally requires a literal `Cap migration:` policy in the new changelog. Tests exercise
  unregistered changes, valid hotfixes, forbidden hotfix cap reductions, valid cap-lowering mints,
  missing changelogs, and seed-only edits in temporary real Git repositories.
- `make epoch-hash` prints the exact worktree bundle identity for review without running the pacing
  simulation. `make harness-check` owns both the earlier baseline-history guard and this epoch
  registration guard.
