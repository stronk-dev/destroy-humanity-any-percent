# Founder Attendance Foundation implementation log

Append-only. A fresh agent must be able to resume from this file and the accepted RFC alone.

## 2026-08-05 — implementation opened

- Owner accepted A1-A5 in `rfc/founder-attendance-foundation.md`.
- Existing `age_ms` remains the only persisted completed-run attendance authority.
- First batch is the Pet-C9 prerequisite: immutable Founder genesis plus Exit participation in
  `founder_log`, followed by career-long Founder replay. Attendance consumers do not land before
  that replay boundary is proven.
- The malformed kernel registry history was repaired pre-push as explicitly approved; the
  corrected commits are `70d2d4a` and `374c7c0`. `node client/tools/verify-kernel-version.mjs`
  passes at kernel `0.3.27`.

## 2026-08-05 — immutable Founder genesis implemented

- Migration 00056 adds one immutable Founder genesis per Founder stream. Existing Founder logs
  backfill only from the exact pre-command revision named by their first replay envelope; missing
  retained state aborts migration rather than fabricating history.
- A deferred constraint makes every Founder-log insert uncommittable without genesis and binds
  sequence 1 to the genesis revision.
- `ApplyFounderLogged` creates genesis from the exact loaded pre-command bytes in the same
  transaction as the first applied or rejected log row. Mutation errors roll the genesis back.
- The real-Postgres fixture proves byte identity, immutability, and rejection of a direct
  genesis-less log insert. Root save integration is green.
- Kernel registry grows with `server/save/foundergenesis.go`; semantic version is `0.3.28`.

## 2026-08-05 — Exit-log/replay wire bounced (B1-B3)

- The archived Run Genesis contract explicitly kept cross-run Founder verification a non-goal
  because Exit's client receipt contains a next-Company snapshot. A4 reverses that decision but
  does not define a reproducible Founder-only receipt.
- B1 proposes one scope-local audit receipt plus a closed `exit.v1` fact arm; B2 makes the Company
  run-log identity a deferrable relational link rather than an unchecked JSON claim; B3 fixes the
  input/output constants-hash meaning and the Go/TypeScript replay surface.
- Exit logging/replay pauses at this exact wire decision. Immutable Founder genesis remains
  complete, tested, and independently reviewable.

## 2026-08-05 — Exit-as-Founder-log persistence implemented

- Migration 00057 adds the B2 source coordinates, a deferrable composite FK to the exact Company
  run-log command, the exit-kind/source biconditional, and the review-requested Founder-genesis
  revision FK.
- Applied and rejected logged Exits now create a Founder command under the existing
  Founder→Company locks, lazily pin genesis, and append a linked Founder row in the same
  transaction. The Company client receipt is unchanged; `founder_log.receipt` contains only the
  B1 audit receipt and is not enqueued a second time.
- `exit.v1` freezes the accepted Founder facts and B3 input/output hash split. The production path
  derives those facts from the shared terminal transition before copying the Founder result.
- The new FK exposed a real archive seam: verified-run compaction deleted the referenced terminal
  command, causing a deferred commit failure that left the queue claim stranded. Compaction now
  retains only Founder-referenced witness rows, and commit errors use the existing transient retry
  path. The composed Postgres exit→verify→board fixture proves the retained witness and is green.
- Fault coverage includes `founder_genesis` and `founder_log` boundaries, both-or-none rollback,
  genesis UPDATE/DELETE rejection, and relational source coordinates. Kernel is `0.3.29`.

## 2026-08-05 — shared Founder transition implemented

- Live `buy_route_hint` and Exit now execute through the projection-free `ApplyFounderLogged`
  boundary; Exit byte-compares the replayed Founder state, audit receipt, and ordered events before
  the multi-stream transaction can commit.
- TypeScript owns a strict Founder-state codec and the same closed
  `invalid | buy_route_hint | exit.v1` transition. A Go-authored shared corpus covers malformed
  commands, an applied route hint, a rejected Exit, and an applied Exit with attendance, facts,
  network slots, and its `founder_advanced` event.
- Both runtimes assert canonical state, receipt, ordered-event, and result-hash identity. Focused Go
  and all client tests are green. Kernel advances to `0.3.30` for the shared transition semantics.
- Persisted genesis-to-head verification remains the next slice; the plan checkbox stays open until
  that loader/verifier and its real-Postgres proof land.

## 2026-08-05 — career-long Founder verification implemented

- `save.Store.LoadFounderHistory` takes one repeatable-read snapshot of immutable genesis, every
  Founder-log row, Founder-scoped events, and the authoritative head. It validates JSON and the
  all-or-none relational Exit coordinates without reading Company state.
- Go and TypeScript now replay a complete three-command career (ordinary applied command, rejected
  Exit, applied Exit), enforce sequence/revision/hash transitions, and compare final state, every
  audit receipt, and ordered Founder events. Missing pinned artifacts never fall back to the
  deploy-current epoch.
- The real-Postgres elective-Exit fixture loads and verifies the committed history, then proves
  event tampering, a sequence gap, and an unavailable artifact set receive distinct fail-closed
  verdicts. The Go-authored cross-runtime corpus carries the same full career.
- The Founder replay plan item flips with these tests. Kernel advances to `0.3.31`; the next slice
  is the race-safe, offline-aware effective-attendance resolver.

## 2026-08-05 — shared attendance resolver and parity boundary implemented

- `ResolveFounderAttendance` now freezes the complete A2 tuple from the exact active Founder and
  Company streams. It reads Founder first, resolves the run's pinned bundle, classifies any
  unresolved gap on a clone using the pinned `catchup_ceiling_ms`, and computes the total partial
  with the existing prestige Attended-Time implementation. The clone is never persisted.
- `CompletedFounderAttendedMS`, `EffectiveFounderAttendedMS`, and
  `ValidateFounderAttendanceSample` make `age_ms` the only completed-run authority, reject exact-
  integer overflow, and fail stale when Exit advanced the Founder base or revision. TypeScript
  implements the same parsing and validation surface.
- Shared vectors cover zero, the MaxExactInteger boundary, equal effective time across a run
  transition, stale base/revision, and overflow. Unit coverage proves sub-ceiling reconnect time
  remains attended, a 25-hour dormant gap contributes zero, and deploy-current relabeling cannot
  replace the pinned bundle.
- The real-Postgres two-order fixture proves a pre-Exit accepted sample and an Exit-winning stale
  sample both converge to the same effective attendance, with the next run starting at partial
  zero. The resolver plan item flips with its proof; consumer-owned pet/faucet fixtures and the
  remaining rollback/retention coverage stay explicitly open. Kernel advances to `0.3.32`.
