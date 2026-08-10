# Founder Attendance Foundation implementation log

Append-only. A fresh agent must be able to resume from this file and the accepted RFC alone.

## 2026-08-05 — implementation opened

- Owner accepted A1-A5 in `rfc/archive/founder-attendance-foundation.md`.
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

## 2026-08-05 — designated reviewer verdict: Founder Attendance impl (7b1d356..c2be82c) — APPROVE

Review by: the project's designated Claude reviewer. Recorded by: same. Range-union: NO prior
verdict covered this span (the plan lists independent review as OPEN) — this pass is the review of
record. That's the FOURTH consecutive batch where the delegated approvals stopped short and the
designated pass is the actual coverage; recorded as a standing pattern for the archival gate.

**Approved — no blocking defects. A1-A5 + B1-B3 implemented as ruled, verified at source:**
- **A1:** grep confirms NO persisted founder_attended_ms field anywhere; age_ms is the sole stored
  authority (finishExitResolved the only additive writer); the effective value is a pure computed
  accessor. My A1-duplicate error is fully corrected in code.
- **A2 (the critical item):** ResolveFounderAttendance freezes the 7-tuple; ValidateFounderAttendance
  Sample rejects with ErrFounderAttendanceStale unless persisted age_ms == completed_attended_ms AND
  the Founder revision matches. Every interleaving traced: an Exit between resolve and Founder-lock
  advances age_ms + bumps the revision, so the stale sample FAILS CLOSED — no schedule combines a
  post-Exit age with a pre-Exit partial; nothing double-counts or loses the interval; the transition
  reads only Founder state. The real two-order Postgres fixture proves exactly-once in BOTH
  command-wins and Exit-wins schedules. This was the one place a silent attended-time double-count
  could have hidden — it can't.
- **A3:** the resolver clones, RecordOfflineSpan with the run's PINNED catchup_ceiling_ms BEFORE
  AttendedMS, never persists the clone; the 25h-dormant→zero-attended and first-back hazard are
  closed; wrong/unresolved bundle fails closed typed.
- **A4:** the founder verifier + loader read ONLY founder-scoped rows (genesis, founder save_revisions,
  founder_log, founder events) — grep confirms no run_log / Company-stream read; Exit-as-founder-log
  now writes the relational-source-coord row for applied AND rejected Exits in the multi-stream tx.
- **B1/B3:** exit.v1 audit receipt regenerates from Founder state alone (live byte-parity checked);
  constants_hash column = input always, result_constants_hash fact = output (equal on rejection).
- **B2:** migration 00057 — nullable source columns, UNIQUE on run_log coords, deferrable composite
  FK, all-or-none + exit-shape CHECKs; both rows in one Exit tx.
- **A5:** overflow rejected before mutation in the new code (EffectiveFounderAttendedMS,
  applyFounderExitResolved); shared Go/TS vectors cover zero/carry/transition/stale/overflow.
- Kernel 0.3.32, all four new production files under the registered prefix, founderhistory.go added.

Low/non-blocking (queue with the consumers): N1 — the resolver has no LIVE consumer yet (proven at
fixture level; the "inside a Founder-locked command" placement is contract-by-convention until pet
decay / faucet wire it); N2 — the pinned-vs-deploy-current ceiling difference is covered only
indirectly (add the explicit differing-ceiling fixture); N3 — finishExitResolved mutates-then-checks
age_ms (pre-existing Exit code, harmless: overflow aborts the whole tx); N4 — the exit-shape CHECK
is SQL-NULL-permissive, backed by app-side validation (defense-in-depth).

**Verdict: the Founder cross-stream chain (ApplyFounderLogged → genesis → Exit-as-founder-log →
age_ms-sampled race-safe clock) is complete and sound. Proceed to the pet-decay and faucet
consumers.**

## 2026-08-07 — owner-side docs ruling (Claude): canonical home = founder-transitions.md

The archival audit (2026-08-07) flagged an open docs decision: dedicated page vs the existing
attendance section in docs/founder-transitions.md. RULED: the founder-transitions.md attendance
section IS the canonical home — attendance is a facet of founder transitions, not a standalone
system, and a dedicated page would split one lifecycle across two documents. The archival move
cites this ruling; the remaining pre-archival items are unchanged (rollback/retention Postgres
coverage residue + the final verify + move). docs/README.md must index founder-transitions.md
when the archival lands (it is currently unindexed — flagged in the 2026-08-07 hygiene list).

## 2026-08-10 — rollback/retention closure handoff

- Audited the carried coverage debt against the live suite. The rollback arm was already real and
  attendance-specific: `ApplyFounderLogged` increments `age_ms`, an oversized authoritative
  outbox receipt fails after the log write, and Postgres proves the Founder revision, log, intent
  record, and outbox all remain unchanged.
- Added the missing retention arm. Five more applied Founder commands advance the stream beyond
  the ordinary five-revision window; the exact retained set is genesis revision 1 plus revisions
  4–8. Revisions 2–3 are pruned, while `LoadFounderHistory` still returns genesis 1, head 8, and all
  eight immutable command rows.
- The canonical attendance section in `docs/founder-transitions.md` now records both guarantees,
  and `docs/README.md` already indexes that ruled canonical home.
- `make test-save-integration` and the complete root `make verify` gate are green with the new
  real-Postgres proof (6,623 client tests, 19,881 browser tests, zero type diagnostics, and kernel
  history green at the unchanged 0.3.88).
- This closes the remaining test debt only. It is ready for designated cross-party review after
  the root gate; no self-review authorizes archival.

## 2026-08-10 — designated cross-party verdict: rollback/retention closure {ab08ee8} — APPROVED; ARCHIVAL AUTHORIZED

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.
The retention test's exact-set SQL assertion ({1,4,5,6,7,8}: genesis witness + last five) is
self-evidently discriminating; the archive-race seq fix necessary; plan flips honest; state
honestly pre-archival; docs updated in-change with the ruled canonical home indexed. With the
prior FA designated verdict + the canonical-docs ruling + this closure, **the archival move is
AUTHORIZED** (provenance note: the prior verdict's pre-repair hashes dangle; coverage is closed
via the mint verdict's range-union declaration — recorded here for future readers).

## 2026-08-10 — archival execution

Founder Attendance is archived under the authorization immediately above. The frozen RFC lives at
`rfc/archive/founder-attendance-foundation.md`, its evidence at
`planning/archive/founder-attendance/`, and its canonical behavior remains the attendance section
of `docs/founder-transitions.md` (indexed from `docs/README.md`). This move consumes the original
designated implementation verdict, the 2026-08-07 canonical-docs ruling, the mint verdict's
range-union provenance closure, and the designated rollback/retention verdict for `ab08ee8`.

## 2026-08-10 — designated cross-party verdict: the archival execution {fc1440d} — APPROVED

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.
Archival procedurally valid: the authorization verdict precedes the execution entry; every
citation resolves (implementation verdict, canonical-docs ruling, mint range-union provenance
closure, the ab08ee8 closure verdict); the dangling-hash provenance note carried verbatim; moves
byte-intact apart from sanctioned edits; index rows correct; zero dangling path references; no
self-approval language. Founder Attendance is the eleventh archived foundation.
