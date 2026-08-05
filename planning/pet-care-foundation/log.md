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
