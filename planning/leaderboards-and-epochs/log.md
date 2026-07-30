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

## 2026-07-30 — independent review: epoch history guard round (e2fce6c..2096916, L8 portion)

Full-diff review of 35aec16 (epochguard.go, seed, fixture, tests) plus the wiring commits.

**Verdict: approved with findings.** The guard is fail-closed everywhere I probed: shallow history
hard-fails, dirty seed/artifact worktree fails, artifact rename/delete without registration fails at
the HEAD hash check, mint requires exactly one appended immutable epoch + a changelog file new in
the same commit, hotfix is append-only on a sorted unique set, and the repository tests exercise
each rejection path with real git fixtures. Subject-prefix BALANCE-CHANGE detection matches
changeguard.go exactly.

Findings:

1. **MEDIUM — no single authority for the constants-hash artifact set.** The seed declares four
   artifacts (commons, economy, prestige, routes). `server/harness/harness.go:210` composes the
   suite ConstantsHash from three (no prestige); `server/production/prestige_integration_test.go:54`
   composes a different three (no commons). Each site is internally consistent, but the moment the
   composed server stamps live revisions, whichever set it picks will disagree with either the epoch
   accepted-hash set or the harness — and L2 verification would then report `constants_mismatch` for
   every honest run. Required: runtime and harness must derive the artifact list from the seed (one
   source), plus a parity test asserting seed set == every composition site. Fix-forward micro-RFC
   scope; not blocking this commit because no composed server exists yet.
2. **LOW — reverts to an already-accepted hash are inexpressible.** A hotfix whose resulting hash is
   already in the accepted set cannot be registered: the seed must change (artifact-changed rule) but
   `isAppendOnlySet` requires exactly one new hash. Fail-closed, so safe, but the escape hatch today
   is a mint. Ruling requested from no one — I rule it acceptable as-is; document in docs/ci.md that
   constants reverts mint.
3. **LOW/observation — cap-lowering detection is scoped to economy resource hardcaps.** Commons and
   routes catalogs carry balance values (capacity targets, tithe bands) that are not "hardcaps" in
   the L8 sense; lowering them registers as an ordinary hotfix/mint without the Cap-migration demand.
   Matches my L8 intent (the rule exists for player-held resource caps and migration clamping), so
   this is confirmation, not a defect.
4. **Observation — merge commits that combine multiple registrations fail closed** (first-parent
   diffing sees the combined delta and rejects it). Single-branch workflow unaffected; if the repo
   ever adopts merges, registrations must stay one-per-commit. No action.

## 2026-07-30 — independent review: leaderboards core (67e5a32, abcd110) — the review I owed

Adversarial two-lens review, all load-bearing findings re-verified against source by the reviewer.
**Verdict: the storage/projection layer is excellent — same-transaction run logs with fault-injected
rollback proof, race-free seq allocation, real immutability triggers, atomic pins serialized against
mint, concurrency-safe world-first, global ranks with stable keyset cursors, honest DESIGN-GAP
boundary (no replay pretense). But two HIGH findings are architectural and both are holes in MY RFC
contracts, not implementation errors. Rulings issued below and amended into the RFC.**

Findings:

1. **HIGH — a real epoch mint strands every in-flight run at Exit.** `exit.go:250` pins run N+1
   with the ENDED run's hash; `PinRunToCurrentEpochTx` requires that hash in the CURRENT epoch's
   accepted set — after any mint that changes bytes, every Exit fails `ErrEpochUnavailable`,
   forever (no constants-migration machinery exists). The integration test masks it by minting
   epoch 2 with identical artifacts. Root cause: my L5 never said which hash run N+1 starts under.
   **Ruling L5b:** the new run starts under the server's CURRENT constants hash — the Exit
   transaction assembles D6's new-run state from the current catalog, encodes the new company
   revision under the current hash, and pins with it; the ended run keeps its original pin.
2. **HIGH — a `kernel/VERSION` bump permanently strands every active run** (`runepoch.go:85`
   demands pin version == build version on every command; pins are trigger-immutable). Root cause:
   my L2/L3 defined replay identity but no run-completion story across a version bump. **Ruling
   L2b:** the play-time check becomes hash-only; on version drift the command executes and the run
   is appended to a new append-only `run_version_drift` table — the run stays playable, verification
   yields `engine_mismatch`, board projection excludes drifted runs. Runs never strand; ranked
   integrity is preserved structurally.
   **Ruling L5c (the deploy-ordering tail of both):** `cmd/gameserver` startup reconciles the DB
   epochs/hashes with the repo seed (idempotent insert of missing rows) BEFORE readiness, so the
   process's own hash is in the current epoch by construction; account creation can then never race
   a mint.
3. **MEDIUM — mint livelock on aborted first attempt**: `epochs.go:81-86` validates
   `changelog/epoch-<id>.md` AFTER the bigserial INSERT; an aborted mint burns a sequence value and
   every retry named for the expected id fails one-ahead forever. Fix: allocate the id first
   (`SELECT`+explicit id in one tx) or validate the ref against `currval` pattern-free.
4. **MEDIUM — no correction path for mis-projected verified rows** (trigger-immutable, `run_id` PK):
   the docs' append-a-new-record doctrine is unsatisfiable for the same run_id. Fix: supersession
   column (`superseded_by`) writable only via a dedicated audited insert — a follow-up contract, not
   silent mutability.
5. **MEDIUM — rejected intents consume unbounded permanent run_log storage** (full canonical payload
   per rejection, no revision advance needed): loopable for storage growth. Fix: per-run rejected-row
   cap (catalog value) with oldest-rejected pruning — rejections are not replay-load-bearing.
6. LOW — post-prune retry of an old rejected intent surfaces a raw unique-constraint error instead
   of a typed idempotency result. LOW — `epochs` table itself lacks an immutability trigger
   (closed-epoch `ended_at` rewritable when no current epoch exists). LOW — board rows accept
   malformed `run_id` (no format CHECK). OBSERVATIONS — world-first omits `mandate_level`
   (matches my L6 text; ruled deliberate for Phase 0 — mandates are one board family, world-first is
   per category/variables/epoch); import stamps current hash over run-1 pin (already filed under the
   account review as the import-wedge finding); unlogged `Store.Write`/`ApplyIntent` bypasses have
   zero non-test callers (grep-verified) — standing footgun, noted.

Planning-log claims all cross-checked and held; the only test-vs-reality gap is finding 1's
identical-artifact mint.
