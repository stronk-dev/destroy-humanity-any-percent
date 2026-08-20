# Platform alignment — append-only log

## 2026-08-20 — scope correction after owner challenge

- Owner correctly challenged the elapsed time and depth of the first checkpoint. `cb162a3` is now
  explicitly classified as control-plane setup plus a principal-risk scan, not the completed
  sibling-repository-style audit.
- Replaced the lightweight plan with an eight-wave program: exhaustive inventory, per-outcome
  traces, acceptance/test discrimination, runtime/release research, owner rulings, dependency
  graph, executable queue, and contradiction/designated-review closure.
- Product evidence remains pinned to `190a4fa`. No verdict is promoted merely because the planning
  structure exists.
- Began Wave 1 with counted populations: 14 top-level design documents, 24 active-directory RFC
  files, 46 archived RFCs, 23 live planning directories, 38 docs, 340 Go files (151 tests), 74
  migrations, 82 client source files, 56 client test artifacts, 91 balance files, ten copy files,
  72 Make targets, seven CI jobs, and five registered Game UI surfaces.
- Filed RP-016 after a code/data/RFC trace disproved Minigame Platform's “normative body
  reconciled” claim. Production registers only The Pitch; the duel RFC is draft and `balance/combat`
  is absent, while the platform header, MP1, MP5, rulings, and AC6 still require or assume duel as
  tenant #1. This is routed to the ruling author rather than repaired by implementation inference.
- Full design-body reading corrected the initial audit's D-004 framing: silent server-anonymous is
  already the adopted default and local-only play is the outage fallback. Filed RP-017–RP-020 for
  the Soul-recovery, cold-open-WR, minigame-socket, and production-hot-reload contradictions. None
  was rewritten because owner-authored/ruling-body reconciliation belongs to the author.
- Completed the first full read of all 14 tracked top-level design documents and recorded 121
  preliminary stable outcome IDs in `design-capability-ledger.tsv`. These are section-level
  inventory rows, not yet Wave-2 proof: aggregate rows remain explicitly marked for splitting into
  independently falsifiable user workflows.
- Completed R-001's first diagnostic wave. A full current local
  `make harness-check HARNESS_WORKERS=12` reached exit 0 below the 30-minute ceiling, while hosted
  run `32404232364` spent 29m04s silently inside the same `mode=check` before job cancellation.
  Five consecutive current-product hosted runs are cancelled. The last hosted-green harness was
  6m16s at `146deded`; the later `a6328df` capacity commit has no Actions run of its own.
- Code tracing shows the 12-worker dispatcher covers the 300 standard pacing simulations only;
  registry rows and their relevance experiments are serial. The production catalog expanded from
  two generator classes to the T0-T1 schema-v4 catalog without changing the visible 300-run count.
  Filed RP-021–RP-023 for unproven hosted headroom, partial worker coverage, and silent termination.
- Corrected two tempting but invalid focused-run inferences: `make t1-relevance` uses fixture data,
  and a scenario-only active-path run does not receive the registry loader's accepted epoch hash.
  Neither is the active authoritative row. The full registry-aware local check matched its golden;
  R-001 now requires an authority-preserving selector before per-row cost claims.
- Extracted every true active-RFC acceptance criterion into `active-acceptance-ledger.tsv`: 111
  unique `(RFC, AC)` rows across 21 product/process RFCs. A first parser reported 115 because it
  swallowed Combat Duel's nested D4 hardening list; the heading-boundary control caught and
  removed those four false ACs before the inventory was accepted.
- Completed the first 111-row acceptance classification without promoting test names to proof:
  39 draft rows, 56 mechanically backed rows awaiting execution/range review, nine unmet/partial
  rows, three current contradictions, and four withdrawn rows. `acceptance-audit.md` records the
  exact CI, Minigame Platform, Game UI, API, Minigame Surface, and Combat downgrades plus the next
  cold-execution batches.
- Ran Account/Leaderboard/Transport/Gameserver Batch A cold. Account and Leaderboard real-Postgres
  packages, Account/Transport/Gameserver unit packages, and Integration-named Gameserver tests all
  reached exit 0. Kept Account AC2 partial because refresh-family revocation and connected-socket
  reauthentication have no composed witness. Kept WebSocket AC4/AC5 partial because overflow→resync
  and drain→reconnect are only separately tested. Filed RP-024.
- The first focused Account selector never launched: `SAVE_TEST_FLAGS` is expanded unquoted and a
  parenthesized Go regex is parsed by `/bin/sh`. The safe package-wide rerun passed; the invalid
  attempt was not counted. Filed RP-025 rather than hiding the selector failure.
- Ran the complete browser suite cold (123 files; 20,007 passed; three visible skips), its isolated
  Game UI performance arm, the real Postgres/gameserver/Chromium composed bootstrap, and client
  type/boundary gates. The composed script explicitly ends at Desk/snapshot/presence, confirming it
  cannot prove AC1. Downgraded AC4 and filed RP-026 because drain/resync are co-injected with only
  an axe/mechanical-ID assertion, not the required story-beat or recovery behavior.

## 2026-08-20 — Codex truth audit and program foundation

- Audited local and remote `HEAD` at `190a4fa`; `origin/main` is identical. Preserved the user's
  sole pre-existing worktree edit in `AGENTS.md` and made no product, schema, balance, RFC, or
  player-copy change.
- Read RFC-0000, the active index, vision/architecture/voice rules, current state, backlog,
  research matrix, coverage map, every active RFC status, every live plan, representative logs,
  canonical docs, client/server boundaries, Make gates, and deployment artifacts.
- Remote verification found the repository public and every current-head Actions run cancelled.
  Push run `32009994004` proves server/client/schema/browser/composed-browser green but cancels the
  harness after 30 minutes inside `balance-harness -mode=check`; the following three scheduled runs
  repeat the cancelled conclusion.
- Reclassified outcomes at user-workflow granularity. The numeric/save/production and bounded
  T0–T1 server paths are genuinely proven. Accounts, UI, minigames, pets, MMO, leaderboards, and
  accessibility contain strong mechanical fragments without their complete user consumers. The
  nine-tier product, release packaging, backup/restore, export, and self-host/sunset deliverables
  are absent.
- Filed the release audit and alignment artifacts. The execution hold permits only R-001 diagnosis
  and lifecycle record reconciliation until owner choices and authored body repairs land.
- Verification note: a cold local `make verify GO_TEST_FLAGS='-count=1'` was launched. All Go
  packages and pre-harness generators passed before it entered the same long-running
  `balance-harness -mode=check` step. After continued silence it was manually interrupted; the
  aggregate exited 130 at `harness-check`. This is **not** a green or completed measurement and is
  not used to set a larger budget. It independently reproduces the hosted stall location only.
- Validation exposed RP-015: the mandated internal backlog, research matrix/dossiers, and coverage
  map are gitignored from the public repository. The internal ledgers were updated locally, while
  `backlog.md` and `release-platform-audit.md` carry the same program evidence in tracked form.
  Nothing was force-added or published; D-002 owns the durable public/private model.

## 2026-08-20 — Permits and First Content lifecycle reconciliation

- Read both active RFCs and their complete plan/log authority chains, inspected the exact candidate,
  mint, repair, baseline, archival, and cold-cache commits, and reconciled their cross-party verdict
  ranges. The Permits candidate set `{7d9cb37, 90633a6, d30ab9e}` and the FCE mint set
  `{3ff34bf, 84cf570, c41b388, 08c995e}` are genuinely designated-review-covered; later F1 and
  cold-cache repairs also have exact designated verdicts.
- Promoted only the criteria that survived fixture inspection and range review. Permits AC1/AC2
  are proven; FCE AC1/AC3 are proven; FCE AC2 is historically proven but its HEAD witness drifted;
  FCE AC5 is proven only under the accepted range-head ruling. Kept Permits AC3 and FCE AC4 partial.
- Filed RP-027–RP-033. Permits never landed the required exact two-resource Go/TypeScript replay
  row; the existing doctrine replay fixture constructs a cash-only T3→T4 gate. The Permits body
  still contradicts PT-C1/PT-C2/PT-C4. FCE's changelog does not resolve every consumed artifact to
  an exact reviewed range, and FCE5.5/AC5 still contradict the range-head ruling.
- Confirmed current canonical-doc drift: Economy/Purchasable docs call the active catalog pre-mint
  schema v3, Routes says production has no T3→T4 gate, and Pitch still says no production epoch pins
  its content. No canonical docs were silently rewritten because closeout also needs missing proof,
  authority reconciliation, and a new designated range.
- Current cold evidence passed: `./production ./replaycatalog ./harness -count=1` (4.758 s,
  5.759 s, 35.667 s); client (39 files passed, 2 skipped; 6,655 tests passed, 15 skipped);
  real-Postgres Integration tests for `./production ./gameserver`; `./economy ./routes ./doctrine
  ./epochseed -count=1`; schema, copy, and deployment-content-manifest checks.
- Wrote `permits-first-content-lifecycle-audit.md` with criterion-level verdicts and the smallest
  honest closeout order. Did not flip either implementation plan, edit owner-authored RFC text,
  archive, change product bytes, or touch the user's `AGENTS.md` edit.
