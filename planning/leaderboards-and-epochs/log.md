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

## 2026-07-30 — HIGH remediation: epoch transitions and engine drift

- Exit decisions now declare the current hash for run N+1. The store validates the Founder and new
  Company under it, writes their revisions/events and receipt outbox under it, and pins the new run
  to the current epoch with it; the terminal Company revision/event remain on run N's hash.
- The regression fixture starts under epoch 1, mints epoch 2 with genuinely changed artifact bytes,
  exits, and asserts old/new run pins, Founder/Company revision hashes, and ended/started event
  hashes independently.
- Play-time run validation is hash-only. A kernel mismatch inserts one append-only
  `run_version_drift` fact and permits the command; a second command proves deduplication. Board
  projection rejects a run whose canonical Company-stream/run-seq identity has that fact.
- Focused Go tests/vet and the full Postgres integration target are green.

## 2026-07-30 — MEDIUM remediation: artifact authority and startup reconciliation

- Added `server/epochseed` as the single strict owner of the manifest schema, artifact names/paths,
  worktree bytes, current epoch, accepted hashes, and composed constants hash. The history guard now
  aliases that schema instead of maintaining a second decoder.
- Harness loading derives its identity from every manifest declaration and asserts its three
  executed catalog paths match the manifest. Prestige therefore participates in the harness hash
  even though pacing does not execute Prestige policy. Production/Prestige integration fixtures
  seed Postgres from the same four-artifact bundle; adding a fixture artifact to the manifest is
  proven to change composition without a code edit.
- Added transaction-serialized seed reconciliation. Before readiness, the gameserver's required
  synchronizer verifies its expected process hash, idempotently inserts exact current bytes, and
  either bootstraps epoch 1 or advances one epoch. Missing/skipped historical accepted sets fail
  closed because the current worktree cannot honestly recreate their immutable bytes.
- Epoch mint now computes and validates the next explicit ID before any insert, under the same
  advisory transaction lock used by reconcile/hotfix. The demonstrated invalid-first-attempt case
  no longer burns the bigserial sequence; a valid epoch-1 mint still receives ID 1.
- Added the narrow `CONSTANTS-IDENTITY:` baseline path required by this correction. Repository
  validation permits only the two harness artifacts, recomputes the manifest hash at that commit,
  blanks old/new hash fields, and then demands semantic equality of every pacing/golden field. The
  regenerated diff is exactly three hash replacements; no metric or final-state value changed.
- Post-commit self-review closed two boundary assumptions before handoff: exported/in-process Seed
  values now pass the same validator as decoded JSON before artifact reads or DB reconciliation,
  and identity-only baseline comparison strict-decodes exact schemas so unknown fields cannot be
  ignored by semantic comparison. Seeded negative tests cover both paths.

## 2026-07-30 — independent review: remediation round, epoch identity half (ac022cd, 8a291ae, 0e6cbe5, d7b4754)

**Verdict: L2a, L5c, and the mint-livelock fix are approved with evidence.** Single artifact
authority is real (epochseed bundle drives harness, guard, and every integration fixture; parity
test + seed-drives-composition-without-code-edit both proven); startup sync runs before readiness,
is idempotent, serializes all three mutators on one advisory lock, and fails closed on every
divergence probed (DB-ahead, hash conflict, forged bundle hash, unknown fields); mint allocates its
id explicitly under FOR UPDATE with the changelog check before any insert — an aborted attempt
burns nothing (regression-tested). The CONSTANTS-IDENTITY class is structurally checked, not
honor-system: artifact-only commit paths, recomputed-hash equality, and DeepEqual over fully
concrete structs with only hash fields blanked. d7b4754's `test:` subject is legal under the
current guard (golden-seed-only commits are outside the baseline walk — see finding 3).

Findings:

1. **MEDIUM — identity-class residual: a semantic change to a hashed-but-unexecuted artifact can
   ship as hotfix + CONSTANTS-IDENTITY with no mint.** Prestige bytes are in the constants hash but
   the pacing harness never executes them, so halving prestige costs leaves observations identical:
   hotfix (no subject demand) + identity-only baseline refresh = a balance change with history
   asserting "identity-only". Fix (structural, small): the identity guard additionally requires
   artifact BYTES unchanged between the previous baseline commit and this one for every seed
   artifact — identity refreshes may follow composition changes only.
2. **MEDIUM — ReconcileSeed cannot bootstrap a fresh database once real history exists** (empty DB
   + multi-epoch seed hits the fail-closed default; a past hotfix's multi-hash set fails the final
   equality even for single-epoch seeds). Fail-closed is right; unrecoverable-DR is not. Fix:
   bootstrap branch replays the FULL seed history (all epochs + accepted sets, closed ones ended)
   — deterministic, still fail-closed against divergence.
3. LOW — golden-seed-only commits bypass subject classification (pre-existing; operationally
   backstopped by check-mode regeneration). Extend the baseline walk to both artifacts when
   convenient. OBSERVATION — `advanceEpochSequence` setval is non-transactional (harmless: inserts
   use explicit ids); MintEpoch trusts its caller for artifact names until next reconcile;
   gameserver library still has no `cmd/` binary (known, queued).

## 2026-07-30 — round-2 MEDIUM remediation: identity artifact pin

- `CONSTANTS-IDENTITY:` now compares every seed-declared artifact byte-for-byte between the
  preceding pacing-baseline commit and the identity commit, in addition to its existing strict
  report comparison. A Prestige hotfix can no longer hide behind unchanged pacing observations.
- Artifact membership is pinned too. The only migration exception is the repository's first
  historical identity repair, whose previous baseline predates the seed manifest and therefore has
  no declared bundle to compare. Every later repair fails closed on missing, renamed, or changed
  artifacts.
- A real temporary-Git-repository regression registers changed Prestige bytes as a hotfix and
  proves the identity artifact check rejects them by name; the unchanged control passes.

## 2026-07-30 — round-2 MEDIUM remediation: full-history seed bootstrap

- Empty-database reconciliation now inserts the manifest's complete ordered epoch history and every
  accepted-hash identity in one advisory-locked transaction. Predecessor epochs are closed on a
  deterministic millisecond timeline ending at the supplied current start; only the declared
  current epoch remains open.
- The current worktree bundle still supplies the only artifact bytes reconciliation can prove.
  Historical hashes receive catalog-set identities to satisfy the immutable foreign-key graph, but
  never fabricated bytes; recovering old replay bytes remains a database-backup obligation.
- Existing databases retain the stricter prefix rule and one-epoch deployment advance. Every
  changelog in the seed is validated before mutation. A real-Postgres test bootstraps two epochs and
  three accepted memberships from empty state, proves closed/open status and idempotency, and checks
  the historical identity has no invented artifact rows.

## 2026-07-30 — governance-row integrity remediation

- Migration 00021 closes the remaining mutable-evidence path on `epochs`: the current row may gain
  exactly one `ended_at` while all identity metadata remains unchanged; a closed-row rewrite,
  current-row metadata rewrite, or delete raises. The existing mint path remains the only caller of
  the permitted transition.
- The same migration adds a database CHECK for canonical `company UUID:positive run_seq` board
  identities. Real-Postgres coverage attacks both epoch states and a malformed direct board insert.
- The verified-run supersession contract and rejected-log catalog value remain explicitly
  unimplemented: the review requires a follow-up contract for the former, and no accepted catalog
  owns the latter value. This batch does not hide either DESIGN-GAP behind a code constant.

## 2026-07-30 — independent review: 4115e78 (identity byte-pinning), cc2cb50 (seed bootstrap), 6f2047b (governance)

**4115e78: approved.** The identity guard now compares git blob bytes per seed artifact against the
previous baseline-walk commit with the artifact set DeepEqual-pinned; same-commit smuggling is
blocked upstream by the artifact-only path rule; negative test drives a prestige-byte hotfix +
identity commit through the full guard and demands rejection naming the artifact. The round-2
mint-free balance-change hole is closed structurally.

**cc2cb50: approved with one MEDIUM residual.** Empty-DB bootstrap now replays the full seed
history (deterministic timeline, all accepted-hash memberships, catalog-set identity rows for
historical hashes WITHOUT fabricated bytes — honest; docs assign historical bytes to backups).
**MEDIUM: the one-epoch ADVANCE path still wedges on a multi-hash current epoch** — mint N+1 plus
a hotfix on N+1 before redeploy leaves the DB at N; the advance inserts only the worktree hash and
then demands full equality with the declared set → `ErrInvalidEpoch`, server never ready, no
in-band recovery (AddHotfix needs a ready server). Fix: the advance path inserts the DECLARED
accepted set for the new epoch, exactly as bootstrap does — same code path, same honesty.

**6f2047b: approved as scoped.** The epochs trigger admits exactly one transition (NULL→NOT NULL
ended_at, all other columns byte-equal) — closed-epoch reopening is impossible regardless of
current-epoch existence; verified_runs gains the run_id format CHECK; live-Postgres attack tests.
Scope honesty: supersession and the rejected-log cap did NOT land and are recorded as open gaps
(correct); INSERT remains unguarded on epochs (load-bearing for bootstrap — a deliberate,
documented trade-off; a fabricated closed row is the residual surface, noted). Add a dedicated
reopen-attack test when convenient.

## 2026-07-30 — round-3 epoch remediation

- Reconciliation now inserts every declared accepted-hash membership before validating equality,
  for both an already-current epoch and a one-step advance. Only the worktree hash receives local
  artifact bytes; historical hotfix identities remain honest placeholders.
- The real-Postgres advance fixture declares a second accepted hash on epoch N+1 before first
  deployment and proves startup succeeds. Governance coverage explicitly attacks reopening a
  closed epoch with `ended_at = NULL`.

## 2026-08-01 — L7b canonical categories and exact Valuation boards

- Replaced the synthetic L7a fixture with `balance/categories/phase0.json`: five literal category
  rows, an intentionally empty completion set, and prefix-matched `darkpattern.*` /
  `externality.*` exclusions. The Go loader and JSON Schema both pin those Phase-0 literals; the
  evaluator rejects unregistered terminal-fact namespaces.
- Registered categories as the seventh constants artifact in epoch 5. Go replay loading validates
  category bytes against the pinned route gate set; TypeScript hashes the same exact seven-artifact
  set. The queue projector resolves route/category bytes from `catalog_artifacts` by the run's
  immutable constants hash rather than from process-current data.
- Added the Valuation magnitude key as two exact SQL columns: canonical exponent and a padded
  12-digit quantized mantissa. SQL competition ranking and keyset pagination operate on the tuple;
  no binary float or packed surrogate participates. Pre-timer runs project only Valuation, while
  imported and version-drifted runs still project nowhere.
- Proof shipped with the checkbox changes: loader/prefix/namespace tests, exact magnitude parser
  vectors at extreme exponents, shared-rank and pagination integration coverage, pinned artifact
  validation, five-category projection, and pre-timer-only Valuation projection.
- Post-commit adversarial review caught canonical zero's tuple `(0,0)` sorting above positive
  sub-unit values under a naïve exponent-first order. Migration 00039 adds an indexed zero-order
  component, and the integration board proves `1.25e15, 1.25e15, 9.99e14, 9e-1, 0` ranks
  `1,1,3,4,5` with keyset pagination intact.

## 2026-08-01 — L7b independent review follow-up

The category evaluator/projector remained approved. The review found one cross-runtime loader gap:
Go validated pinned category bytes against the route gate set while TypeScript only included those
bytes in the constants hash. The TS replay catalog loader now parses the closed predicate union,
canonical Phase-0 shapes, fact-set namespaces/prefixes, and requires `full_gate_set` to equal the
pinned routes artifact. A malformed category artifact with a recomputed valid hash is rejected, so
hash correctness can no longer disguise invalid category semantics.

## 2026-08-21 — authority and capability-boundary reconciliation

- Owner direction delegated the blocked ruling-body reconciliation. D2/L4 now describe the actual
  six-verdict Go/TypeScript parity foundation and machine-only operational queue without claiming a
  player archive/validator workflow.
- D4 now names the five epoch-owned Phase-0 categories and separates the absent player-authored,
  Exhibition, Route Registry and records surfaces. D5 now claims only bounded mandate-key plumbing;
  no catalog, intent, modifiers or consumer are inferred. D6 now distinguishes stored world-first
  arbitration from absent dispatch and machine-board consumers.
- L1 no longer invents a live `run_ttl_days` cleanup contract. Verified compaction remains current;
  abandoned-run deletion is routed to D-015 and the future accepted retention/operations contract.
- Added the draft `leaderboard-readers-and-player-surface.md` as a concrete home for generated
  readers, browser browsing/history, archive verification, Route Registry composition, dispatch
  and machine boards. Draft status is explicit and authorizes no implementation; D-017 still gates
  player-authored/Exhibition scope and mandate mechanics remain a separate gameplay successor.
- Canonical docs now state the capability boundary and the live terminal-only Commons defect. No
  product code, schema, balance, copy, plan checkbox, RFC acceptance/status promotion, push,
  publication, review verdict or archival changed.

## 2026-08-21 — AC1/AC3/AC5/AC6 witness predeclaration

- Predeclared source→key→board, epoch-crossing replay/projection, Compact any-membership and
  two-later-epoch populations in `witness-manifest.md`, including explicit negative mutations.
- Scope is the accepted backend only. The draft reader/player successor, API operations, UI, copy,
  retention, mandates, dispatch and machine boards remain out of range.
