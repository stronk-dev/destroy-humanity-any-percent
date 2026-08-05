# Pet Care Foundation implementation log

## 2026-08-05 — owner rulings and implementation start

- Owner accepted C1-C8. The central deliverable is the reusable Founder-only transition and replay
  boundary; care must never mutate through Company `ApplyLogged` or spend Company state.
- Reconciled the normative body at the decision sites: fixed mechanical stat IDs, care through
  `ApplyFounderLogged`, and no Company resource/time/token cost.
- Implementation starts with persistence and replay plumbing. Production pet content and numeric
  policy rows remain owner/harness data and will not be invented.

## 2026-08-05 — Founder persistence/replay envelope

- Added `save.Store.ApplyFounderLogged`, a Founder-scope-only mutation path with a typed
  `FounderReplayCommand`. The Store assigns sequence and database server time under the locked
  stream; replay inputs must echo both exactly, so feature code cannot invent attended time or
  import live Company context.
- Appended migration 00054 for immutable `founder_log` rows. Applied and rejected commands are
  both recorded, only applied rows name a save revision, every row pins constants identity, and a
  database trigger rejects any non-Founder stream even if application validation is bypassed.
- Real-Postgres coverage proves applied state/revision, rejected no-op, byte-hash idempotent replay,
  Company-scope refusal before callback, exact row ordering, null-vs-applied revision semantics,
  and SQL immutability. Envelope tests reject mismatched server stamps/commands, unknown keys, and
  non-object resolved inputs.
- `make test-go GO_PACKAGES='./save'`, `make vet GO_PACKAGES='./save'`, and root
  `make test-save-integration` pass. Pet's pure care union and Founder replay verifier remain the
  next independently reviewable slice.

## 2026-08-05 — independent rejection of Founder boundary (`7620311^..7620311`)

- **Review by:** independent Codex subagent `/root/l7b_independent_review`
- **Recorded by:** Codex
- **Decision:** not approved; one HIGH and two MEDIUM findings require remediation.
- HIGH: exported legacy `Store.ApplyIntent` still accepted Founder streams, allowing an unlogged
  Founder mutation that bypassed `ApplyFounderLogged` and `founder_log` entirely.
- MEDIUM: receipt enqueueing hardcoded `scope=company`, so Founder receipts entered durable
  transport with false scope. MEDIUM: migration 00054's direct-insert archive check did not lock
  the parent stream and could race archival.
- The authoritative database timestamp, envelope closure, intended-path serialization,
  applied/rejected log semantics, transaction composition, immutability, retained archive history,
  migration direction, and kernel bump held. Coverage debt was named for concurrency, rollback,
  Founder event/outbox, and archive-race proofs.

## 2026-08-05 — Founder boundary remediation

- Founder scope now rejects the legacy unlogged mutation callback before invocation. Receipt
  enqueueing accepts and validates the actual stream scope, and transport documentation names
  Founder receipts.
- Forward-only migration 00055 replaces the already-applied 00054 trigger with a parent-row
  locking recheck, so archive and direct log insertion have a serial order.
- The real-Postgres fixture now covers the legacy bypass, Founder event and applied/rejected
  receipt scope/revisions, concurrent same-revision commands, a failure after the log insert that
  rolls back every write, and an archive-first/direct-insert race. Kernel semantics advance to
  0.3.26. Independent remediation review remains required.

## 2026-08-05 — remaining Pet implementation contracts

- Applying the boundary to a real pet transition exposed four gaps not resolved by C1-C8: replay
  needs segment genesis around externally-owned Exit mutations; wall-clock stamps cannot supply
  attended time; the Founder-only save/table/activation wire is unnamed; and the allegedly closed
  mood/FSM/status unions have no literal members.
- Filed C9-C12 with proposed executable contracts. No pet numeric balance row is requested; these
  are replay, persistence, time-authority, and wire-shape decisions that code must not invent.

## 2026-08-05 — independent remediation review (`cd60175^..cd60175`)

- **Review by:** independent Codex subagent `/root/l7b_independent_review`
- **Recorded by:** Codex
- **Decision:** not approved; one new MEDIUM in the migrated production consumer.
- Every original HIGH/MEDIUM finding closed. The reviewer verified legacy-bypass refusal,
  Founder-scoped event/receipt rows, forward-only archive serialization, same-revision command
  serialization, post-log rollback, route-hint migration inputs, and kernel 0.3.26.
- New MEDIUM: `buy_route_hint` requests carrying unknown/invalid fields entered the new Founder
  handler but skipped Company `ApplyLogged`'s deterministic invalid arm; a known route could still
  be purchased. The fix must log a closed invalid resolved variant before projector/catalog reads.

## 2026-08-05 — invalid-Founder-command remediation

- The Founder route-hint handler now emits a deterministic rejected decision with resolved
  `{kind:"invalid",detail}` before any catalog or projection read. Its Postgres fixture proves an
  extra-field request remains at revision 1, unlocks nothing, records a rejected Founder-log row,
  then allows the canonical command to apply at the same expected revision.
- Kernel semantics advance to 0.3.27. Independent follow-up review remains required.

## 2026-08-05 — independent approval of invalid-command remediation (`17faed3^..17faed3`)

- **Review by:** independent Codex subagent `/root/l7b_independent_review`
- **Recorded by:** Codex
- **Decision:** approved; no findings.
- The reviewer verified the invalid arm executes before any catalog/projector access, persists the
  closed resolved variant, creates no revision/event or state mutation, preserves terminal
  idempotency, and leaves the same expected revision available to the canonical command.
- Independent production/save tests, vet, kernel-history verification, root Postgres integration,
  and exact-range diff check passed. Full root `make verify` also passes at kernel 0.3.27.

## 2026-08-05 — designated reviewer verdict: ApplyFounderLogged (7620311..1ba07fd) — APPROVE

Review by: the project's designated Claude reviewer. Recorded by: same.

**Approved.** The Founder mutation boundary is a faithful mirror of Company `ApplyLogged`, verified
at source: shared `applyIntent` core so it inherits FOR-UPDATE serialization (double-apply proven
impossible by the real-Postgres two-goroutine test), same-transaction idempotency with
byte-identical recorded receipts, applied AND rejected immutable history (reject_immutable_change
trigger), revision CAS, archive/genesis guards, and — critically — it commits the Founder stream
ONLY and rejects ambient company state in its closed replay envelope (an `ambient_company_state`
key is refused). All four prior findings fixed to the letter with regressions: the buy_route_hint
legacy bypass (the old RunLogSeq==0 unlogged founder path) closed at BOTH store guards and dispatch;
the founder-outbox scope corrected to `scope='founder'`; the archive race fixed forward-only with
FOR SHARE recheck; and malformed-command determinism fixed (invalid arm runs first — with the
honest nuance that for buy_route_hint no literal state-apply was actually reachable, so the real
defect was receipt-category/log-record determinism + defense-in-depth for future founder intents).

**Two residuals actioned/recorded:**
1. **KV-1 registry gap — FIXED directly by the reviewer:** `server/save/founderlog.go` (the exact
   Founder analog of the registered `runlog.go`, owning ValidateFounderReplayInputs / insertFounderLog
   / determinism logic) was NOT in `kernel/affecting-paths.json` — a future founder-ONLY edit would
   slip the kernel-drift gate (the span passed only because it co-touched registered intent.go).
   Added as a surgical append-only registry line (guard-integrity fix, no replay-byte change, no
   version bump — exactly the growth the guard's append-only rule accepts). Codex to confirm it
   lands green.
2. **Range-union gap — RECORDED, blocks any future archival of this span:** the cited Darwin ranges
   union to {7620311, cd60175, 17faed3}; **a2696b5 is UNREVIEWED** — and it is NOT docs-trivia, it
   added acceptance blockers C9–C12 to pet-care AND +63 normative lines to minigame-platform. Per
   the range-union rule, this span is NOT archival-eligible until a2696b5's RFC changes are
   reviewed. (Those changes are the blockers I ruled this turn — so they ARE now owner-reviewed via
   the rulings; recording that the review-of-record for a2696b5 is this turn's ruling pass.) Also
   noted: no single verdict approves the cumulative HEAD; the boundary rests on
   rejection-then-remediation per-commit, acceptable for a live slice but the archival gate must
   cite the remediated end-state.

Dormant, tied to future work: `scopeMatchesChannel` needs a `founder` arm when founder snapshots
land (Pet C7/C8); the 7620311 transient bisect-break is intermediate-only.

**Verdict: the boundary is sound and the buy_route_hint migration is clean. Founder-replay
verification is correctly deferred (Pet C9). Proceed.**

## 2026-08-05 — owner ruling: APPROVE rewrite of 80456c1 (my error)

Approved: rewrite unpushed `80456c1` to place `server/save/founderlog.go` in strict lexical
position (before `genesis.go`), replay `4005f00` on top. Preconditions verified in-repo:
unpushed (no remote contains it), referenced by NO verdict hash (my prior verdict referenced the
CHANGE — "add founderlog.go" — never the commit `80456c1`, which Codex created after), and
protocol-violating with NO forward remedy — `verify-kernel-version.mjs`'s `validateGuard` enforces
strict ascending sort (`paths[i-1] >= entry → throw`) at EVERY historical commit, and the
`history-corrections.json` mechanism forgives only missing-version-bumps (`assertBump`), never a
malformed/unsorted registry. So a forward correction cannot help; the rewrite is the only fix.

**This is my mistake.** Last turn I fixed a real KV-1 coverage gap (founderlog.go absent from the
registry) but did the insert POSITIONALLY (after runlog.go, next to its analog) instead of in the
registry's required strict lexical order (founderlog < genesis < runlog). I didn't check
`validateGuard`'s own sort invariant before editing its data. The working tree is already
corrected (founderlog.go now precedes genesis.go); the rebase carries that correct state into the
commit. **Lesson recorded: editing kernel-guarded config (`affecting-paths.json`) means respecting
the guard's structural invariants — strict lexical sort, append-only-per-commit — not just adding
the right path.** Direct edits to guard data get a `validateGuard`-style sanity check before commit.

## 2026-08-05 — designated reviewer verdict: Founder genesis batch (70d2d4a..7af623c) — APPROVE

Review by: the project's designated Claude reviewer. Recorded by: same. **This is the review of
record for the batch** — the prior ApplyFounderLogged approval covered 7620311..1ba07fd, and
1ba07fd is the PARENT of this batch's base; F2 (range-union) correctly flagged the batch was
otherwise unreviewed. Now covered.

**Approved — sound and complete for what it claims (the immutable-Founder-genesis HALF of A4).**
No correctness defects. Verified independently at source:
- **The poisoned-history repair (70d2d4a) is clean AND proven over the WHOLE rewritten span** —
  the reviewer RAN the guard (exit 0, non-shallow so the full guardIntroduction..HEAD walk),
  which means every commit in the rewrite has a sorted registry and a satisfied bump rule.
  founderlog.go is in strict lexical order at every descendant; the subsequent commits replayed
  intact. Repairing a fail-closed guard correctly (no NEW unsorted/unbumped commit anywhere in
  history) was the thing that most needed independent proof — confirmed.
- **Genesis mirrors run-genesis faithfully:** byte-identical pre-command Founder state via the raw
  jsonb bytes (the `state::text` equivalent, NOT a Go re-encode); genesis + first founder_log in
  one transaction; the deferrable AFTER-INSERT constraint binds seq=1 to the genesis revision so a
  founder command CANNOT run without genesis; ON CONFLICT + FOR SHARE recheck + the seq-1 PK force
  exactly one concurrent-first winner; immutable (UPDATE rejected, tested); fail-closed backfill
  RAISEs if a pruned revision made genesis unrecoverable (never fabricates history); Founder-only
  (no Company read). Kernel 0.3.28 justified, both new paths registered.
- **B1-B3 honesty confirmed:** grep proves Exit writes no founder_log/genesis — Exit-as-founder-log
  is honestly deferred behind the B1-B3 rulings, genesis claims only its half.

Findings (none blocking): **F1 LOW — no real-Postgres fault-injection test for the founder
both-or-none property** (run-genesis has runExitFault; applyIntent has no fault hook). The property
holds structurally (single-tx rollback + the genesis-less deferred-constraint rejection IS tested —
the stronger guarantee), but add the fault-injection parity test when the Exit-founder-log slice
lands (it'll need the hook anyway). F3 INFO — an available `(founder_stream_id, revision) →
save_revisions` FK is unused hardening (parity with run-genesis, not a regression); DELETE-
immutability untested though the same trigger covers it — add the one-line negative.

**Verdict: proceed with Exit-as-founder-log against the B1-B3 rulings; the genesis foundation is
solid.**

## 2026-08-05 — C12a cross-runtime Pet grammar

- Implemented the owner-ruled Phase-A stat IDs, status bands, moods, behavior states/events, care
  rejection details, queue hardcap, and PRNG label in isolated Go and TypeScript modules.
- One shared fixture asserts literal member order and queue boundaries in both suites. Go returns
  defensive vocabulary copies so a caller cannot mutate protocol authority at runtime.
- Added canonical docs scoped honestly to the grammar; no pet identity, state, action, decay,
  threshold, species, or temperament row is claimed or synthesized.
- `DESIGN-GAP:` C3/C11 still do not enumerate the exact pet catalog and Founder-state wire keys,
  including action/decay rows, remainder and cooldown maps, behavior queue entries, and the bond
  graph. Those mechanics remain blocked rather than inferred from prose.

## 2026-08-05 — remaining Pet wire bounce C13-C14

- Filed an exact fixture-catalog proposal for stat decay, actions, trust, mood thresholds, and
  behavior candidates; all numeric values remain balance data.
- Filed one replay-owned Founder pet-state proposal that resolves the mutable-table/snapshot and
  stored/derived-mood contradictions. Immutable identity stays relational; mutable mechanics stay
  inside the Founder transition boundary; writable bonds remain deferred.

## 2026-08-05 — protocol authority hardening

- Self-review confirmed Go vocabulary accessors return defensive copies, then found the
  TypeScript `as const` arrays were compile-time readonly but still mutable at runtime through a
  cast. Froze every exported array and added a mutation regression so protocol vocabulary cannot
  drift inside a running client.
