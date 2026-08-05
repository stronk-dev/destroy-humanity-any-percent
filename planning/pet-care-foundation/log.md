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
