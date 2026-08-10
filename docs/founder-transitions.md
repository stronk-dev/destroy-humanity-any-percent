# Founder-scoped transitions

`save.Store.ApplyFounderLogged` is the reusable authoritative mutation boundary for Founder-only
mechanics. It is deliberately separate from Company `production.ApplyLogged`: a Founder command
cannot read or spend live Company state, and a Company command cannot smuggle a Founder write.

The Store locks one active Founder-scope stream, assigns the next immutable Founder-log sequence,
and stamps ordinary commands with database server time. Multi-stream coordinators freeze their
server-resolved timestamp in the internal request before taking the same locks. The callback receives only the restored Founder
state, its revision, and this authoritative command envelope. It returns the ordinary typed
decision plus a feature-owned resolved-input object inside the persistence-owned replay envelope.
The envelope must repeat the exact command and server timestamp; unknown fields, ambient state,
non-object resolved inputs, and mismatched coordinates fail closed.

Applied and rejected decisions both append to `founder_log` in the intent transaction. Applied
rows identify the new save revision; rejected rows carry no applied revision and cannot mutate
state or emit events. Existing intent-record hashing supplies byte-stable idempotency: retrying the
same intent returns its recorded receipt without invoking the mutation, while reusing an intent ID
for different bytes is rejected. Founder receipts use the normal player outbox.

`founder_log` accepts only active `owner_kind=founder, scope=founder` streams and is immutable at
the database boundary. Each row pins the constants hash and server timestamp used by its resolved
inputs. Pet Care, Fiscal Quarters, minigame ratings, and Soul recovery all consume this boundary.
New Founder mechanics must reuse it rather than adding side writes or extending the Company transition.

The first Founder-log activation atomically pins `founder_genesis` to the exact raw bytes, version,
hash, and revision loaded before that command. A deferred database constraint makes a Founder-log
row uncommittable without this immutable genesis; migration backfill fails rather than fabricating
a starting point whose retained revision is gone. Revision retention always preserves that one
genesis witness even after the ordinary five-revision window advances.

Logged Exits also append a Founder row. Its `constants_hash` is always the pre-command Founder
bundle, while the closed `exit.v1` resolved arm names the result bundle and the exact Founder facts
advanced by Exit. The row carries a Founder-only audit receipt; the existing Company receipt and
client outbox remain unchanged. Three nullable source coordinates are present exactly for
`exit.v1` and form a deferrable foreign key to the immutable Company run-log command authored in
the same transaction. Verified-run compaction retains that one referenced witness row after the
full run has been archived, preserving the relational proof without retaining the whole live log.

`production.ApplyFounderLogged` is the single projection-free transition used by the live Founder
path and by replay. Its closed Phase-A union includes invalid commands, Route hints, pet care,
Fiscal commands, minigame resolution, Soul recovery, and Exit; the Exit arm
updates only Founder-owned facts and generates a separate `founder_advanced` event and audit
receipt. The TypeScript port consumes the same Go-authored corpus and byte-compares the resulting
state, receipt, ordered events, and result constants hash. The production path also runs the same
transition before committing an Exit, so a live/replay semantic split fails the authoritative
transaction rather than entering the immutable log.

`save.Store.LoadFounderHistory` reads genesis, immutable commands, Founder-scoped events, and the
authoritative head in one repeatable-read transaction. Both Go and TypeScript replay each row from
genesis, require contiguous log and revision coordinates, resolve the row's input hash without a
deploy-current fallback, enforce relational Exit/minigame/Soul source coordinates, and compare the final
state to the saved head. Missing artifacts report `constants_mismatch`; sequence/revision gaps
report `log_gap`; malformed inputs or differing state, receipt, or ordered events report
`state_divergence`.

Founder attendance has one persisted authority: Founder `age_ms`, the exact sum of completed-run
Attended Time. `production.ResolveFounderAttendance` freezes the active Company stream identity,
run and revision, pinned constants hash, completed `age_ms`, current-run attended partial, and
their checked sum. It reads Founder before Company, classifies an unresolved Company gap against
the pinned prestige catch-up ceiling on a non-persisted clone, and rejects missing or ambiguous
contexts. A Founder-locked consumer must call `ValidateFounderAttendanceSample`; a concurrent Exit
changes the Founder revision and `age_ms`, making the old sample stale rather than double-counting
the completed run. No second Founder attendance cursor exists.

The Postgres boundary pins the lifecycle failure cases directly. A post-log failure after an
`age_ms` mutation rolls back the Founder revision, log, intent record, and player outbox together.
After enough Founder commands to cross the ordinary five-revision window, pruning removes only
superseded ordinary revisions: the immutable genesis revision remains alongside the latest five,
and `LoadFounderHistory` still returns the complete command career and authoritative head.

The feature package still owns its closed canonical command, resolved-input, receipt, event, and
state-transition unions. The persistence layer validates the shared envelope and transaction; it
does not invent feature mechanics.

Legacy unlogged `Store.ApplyIntent` calls are rejected for Founder streams, so the boundary cannot
be bypassed by an older feature surface. The existing `buy_route_hint` Founder command is the first
production consumer: its immutable resolved inputs freeze the repaired Route Knowledge balance and
route-context version before the purchase. Founder event and receipt outbox rows retain Founder
scope and the exact applied or rejected revision.
