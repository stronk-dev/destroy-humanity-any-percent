# Leaderboards & Balance Epochs lifecycle audit

Coordinate: product tree `190a4fa`; audit checkpoint after `ff2d51c`; 2026-08-20.

Post-coordinate reconciliation (2026-08-21): owner-delegated edits reconciled D2/D4/D5/D6/L1/L4,
the plan, and canonical docs to the six-verdict/five-category backend. A draft Leaderboard Readers
& Player Surface successor now owns the absent integration scope but is deliberately not accepted
implementation authority. The accepted foundation then repaired AC5 with an immutable-history,
three-arm real-Postgres witness. RP-042 is closed; RP-039 awaits designated review; RP-040/RP-041/
RP-043 and successor acceptance remain open.

This pass re-derived the active Leaderboards RFC from its specification, design references, plan,
append-only log, current implementation, the archived Run Genesis successor and remediation
records, canonical docs, tests, and historical review entries. It did not edit product code,
owner-authored RFC text, balance data, canonical product docs, or plan checkboxes.

## Bottom line

The repository has a strong server-side leaderboard foundation: immutable epoch/catalog identity,
atomic run pins and genesis, a six-verdict Go/TypeScript replay kernel, archive-backed compaction,
verification queue projection, exact SQL ranking, frozen historical storage, and guarded balance
history all execute successfully. The intended player capability does not yet exist, and one of the
six acceptance criteria is false in the current implementation:

- no runtime HTTP binding or client surface lets a player browse boards or epochs, invoke the
  transparency validator, or see the Route Registry alongside records;
- Compact assistance was derived from membership **at Exit** at this audit coordinate; the
  post-coordinate repair now derives immutable any-point membership and awaits designated review;
- AC1 and AC6 now have post-coordinate literal, mutation-proven witnesses; AC3 remains narrower
  than its literal cross-epoch replay→projection chain;
- player-authored/Exhibition categories, world-first dispatch emission, machine boards, abandoned
  run retention, and the rejected-log/supersession follow-ups are absent and now explicitly routed;
- the plan predates the implemented Run Genesis successor and is materially stale; and
- no tracked verdict supplies current-law explicit cross-party provenance and an exact range union
  for the complete Leaderboards foundation and later successor work.

This is therefore a mechanically substantial, unshipped capability—not an archival candidate.

## Current cold evidence

All commands ran from the repository root with cold Go counts where applicable:

- `make test-go GO_PACKAGES='./leaderboard ./production ./replaycatalog ./epochseed'
  GO_TEST_FLAGS='-count=1'` — green;
- `make test-client` — 39 files passed, two skipped; 6,655 tests passed, 15 skipped;
- `make test-save-integration SAVE_TEST_PACKAGES='./leaderboard ./production ./gameserver'` —
  green on real PostgreSQL;
- `make test-go GO_PACKAGES='./harness' GO_TEST_FLAGS="-count=1 -run 'TestEpochGuard'"` — green;
- `make harness-guard-check` — green against repository history.

The real-Postgres gameserver fixture proves Exit→stored replay verification→queue projection into
`verified_runs`. The focused board fixtures prove exact time/count/magnitude ordering, competition
ranks, keyset pagination, imported/drift exclusion, pinned category loading, and atomic world-first
arbitration. The shared Go-authored replay corpus and TypeScript suite exercise all six verdicts.
These are valid producer/storage proofs. They are not evidence of a player reader or complete
workflow because none is composed.

## Acceptance classification

| AC | Verdict | Evidence and limitation | Required closeout |
|---|---|---|---|
| AC1 | **Post-coordinate witness complete; review pending** | Two distinct sub-quantum Decimal sources converge independently through canonical wire strings and keys, then share rank 1 in real Postgres. Incrementing one derived mantissa produces ranks 1/2 and fails. | Include the exact witness range in the mandatory cross-party verdict. |
| AC2 | **Cold witness green; archival review missing** | The Go-authored corpus is replayed by Go and the shipped TypeScript module; gap, constants, engine, clock, state/event/receipt mutations produce the same closed verdict classes. The module is bundled but has no player-facing archive/catalog retrieval or invocation surface. | Preserve the parity corpus; decide whether L4 requires only bundled code or an actual player transparency workflow, then include the accepted scope in the full designated review. |
| AC3 | **Partial** | Prestige's real-Postgres fixture starts run 1 in epoch 1, mints genuinely changed bytes, finishes in epoch 2, and proves the ended event/pin retain epoch-1 hash while run 2 uses epoch 2. Separate replay and projection fixtures are green. No single fixture replays that crossing run against epoch N's archived catalog **and** proves its board row/query remains in N. | Add the literal start-N→mint→finish-N+1→replay-N→rank-N chain plus a current-catalog misclassification control. |
| AC4 | **Cold witness green; archival review missing** | Real temporary-Git fixtures reject an unregistered artifact change, a `BALANCE-CHANGE:` without exactly one mint, a mint missing its numbered changelog, and a hotfix hardcap reduction; valid hotfix/mint controls pass. The live history guard also completed green. | Retain the discriminating fixtures and include their post-rewrite ranges in the designated review union. |
| AC5 | **Post-coordinate repair complete; review pending** | `QueueProjector` now derives Commons from run-scoped `compact_signed` history. Real-Postgres join→Exit and join→leave→Exit positives plus a never-joined Solo control pass; severing the history signal fails both positives. | Include the exact implementation range in the mandatory cross-party verdict. |
| AC6 | **Post-coordinate witness complete; review pending** | A real board row is snapshotted in epoch N, two later distinct catalog identities are minted, and the complete N row/rank/key/world-first result remains equal. Querying the newest epoch at the post-mint seam fails the witness empty. | Include the exact witness range in the mandatory cross-party verdict. |

No plan box was changed. The successor Run Genesis implementation makes plan item 3 stale, but the
checkbox law requires its proof and record reconciliation inside one future designated-review
range; audit inference is not such a range.

## Producer→consumer reality

| Layer | Reality at HEAD |
|---|---|
| Intent/run evidence | Proven: atomic per-run logs, immutable genesis, terminal sequence, pinned catalogs. |
| Verification | Proven mechanically: Go authority, TS parity module, queue/retry/dead-letter/archive paths. |
| Projection/storage | Proven mechanically: categories, exact keys, structural variable JSON, epoch keys, imported/drift exclusion, world-first bit. |
| Public contract | Fragment only: `publicapi` owns board filter/cursor policy primitives and test-only example operations, but the production Account registry exposes no board/epoch operations and gameserver mounts no board handlers. |
| Client consumer | Absent: no leaderboard data model, transport reader, page, browse controls, archive verifier workflow, historical-epoch UI, or live-viewer surface. Current chrome shows a local PB only. |
| Route/record surface | Absent: the Route Registry projection exists elsewhere, but no composed page renders it with boards as D4/design require. |
| Executable user proof | Absent: no HTTP/browser fixture can retrieve or browse a board because no runtime reader exists. |

The design outcome—other players' records, historical epochs, categories, routes, world-firsts,
and transparent verification—is therefore **claimed intent over a working backend fragment**, not a
shipped user capability.

## Normative, canonical, and scope drift

1. D2 still declares four machine causes while L2 adds a fifth and the implemented/archived replay
   contract has six (`verified` plus five failure causes). Its “no queue” wording is also
   unreconciled with the operational verification queue, even though there is correctly no human
   judgment queue.
2. D4 says four canonical categories; L7b and current data ship five by adding Valuation. The
   original body also promises player-authored predicates, Exhibition, and a combined Route
   Registry/records surface; none has a runtime consumer.
3. D4/L7 require Commons to mean membership at any point. Canonical docs instead say the projector
   takes Commons from the terminal assisted record, accurately documenting the implementation but
   not the accepted behavior.
4. D5 has only numeric range/key plumbing. No mandate catalog objects, opt-in intent, gameplay
   modifier application, or consumer exists. Content is explicitly out of scope, but the body
   overstates “each a declared catalog object” and “validation” relative to the shipped range check.
5. D6's world-first database arbitration is live, but no feed/dispatch event is emitted and no
   distinct `machine` board class exists.
6. L1's `run_ttl_days` abandoned-run cleanup is absent. The core review also recorded missing
   correction/supersession and rejected-intent retention contracts; no accepted successor owns
   them at HEAD.
7. Canonical docs comprehensively describe backend mechanics but never state that board/epoch API
   readers and player surfaces are absent. `docs/README.md` calls the foundation implemented, which
   is defensible only if “foundation” is kept distinct from the designed capability.

## Plan and review truth

Plan item 3—immutable initial state, archive compaction, replay verification, shared Go/TS
fixtures—was implemented and archived under Run Genesis & Replay. Item 6's canonical backend docs
and composed server verification are also largely present, but the literal AC gaps and absent user
integration prevent a safe checkbox flip. `planning/archive/run-genesis-archival-remediation/` records its
last blocker resolved on 2026-08-16 yet remains in the live planning namespace.

The early Leaderboards log contains technically detailed “independent review” entries for the
epoch guard, core, identity/bootstrap/governance remediation, and L7b. They predate today's
mandatory `Review by:`/`Recorded by:` format, cite pre-history-rewrite hashes, and do not provide a
post-rewrite range union. The archived Run Genesis designated review covers
`2599a11..d3620cb`, and the later remediation designated review covers `6141a0f..8578fb1`; those
ranges validate substantial successor replay/category work but do not retroactively cover the
earlier Leaderboards commits such as current-history `ce0a5ec`, `c864b3f`, `b7ab748`, and their
remediations. The tracked Leaderboards plan correctly leaves independent review open.

## Smallest honest closeout order

1. **Body reconciliation complete after this coordinate; successor acceptance remains open.**
2. **AC5 implementation/proof complete after this coordinate; designated review remains.**
3. **AC1/AC6 complete after this coordinate; add AC3's composed crossing witness.**
4. Review and explicitly accept (or narrow) the draft successor for board/epoch readers, player
   browsing, validator delivery, Route Registry integration, world-first dispatch, and machine
   surfaces. D-017 separately gates player-authored/Exhibition categories.
5. Reconcile the stale plan and archive the completed Run Genesis remediation thread in a reviewed
   record-only range.
6. Obtain the mandatory tracked cross-party verdict whose exact post-rewrite ranges union every
   Leaderboards implementation/remediation commit it claims to approve.
7. Archive Leaderboards only after the accepted current scope, docs, plan, log, and RFC status close
   transactionally in that reviewed span.
