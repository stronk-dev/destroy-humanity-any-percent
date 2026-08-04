# Minigame Platform

The implemented foundation currently owns two boundaries: durable minigame sessions and pure
tenant engines. Production payout, faucet policy, registry balance rows, and gameserver routes are
not yet composed and are not claimed here.

## Authoritative sessions

`minigame_sessions` is the Postgres authority for Phase-A `solo` and `async_snapshot` sessions.
`live_pvp` is deliberately rejected until its separate service owns realtime lifecycle rules.
Every row freezes the Founder/Company run identity, pinned constants hash, minigame and engine
versions, unsigned seed, complete integer scaling-input object, mode, and genesis snapshot. A
database trigger rejects any later mutation of those fields.

Active sessions begin at revision 1. A server-side play or resolve command claims the row with a
database-generated UUID token after locking Founder then session. Concurrent workers cannot both
claim it; a crashed claim can be replaced after the same five-minute lease used by replay
verification. A completed play advances the revision exactly once and returns the row to active.
A resolved result advances once, clears the claim, records its completion time, and is immutable.

Resolution exposes a transaction-bound service write rather than a public client command. A
terminal play returns an opaque certification whose identity and result fields cannot be populated
outside the platform package. Resolution locks that exact Founder-owned Company/run/hash first,
then token-locks the matching session and records the terminal snapshot plus tenant-validated
result inside the same transaction. Rolling that transaction back leaves the claim and session
state unchanged, preserving the seam required for the later server-authored payout transition.

## Tenant boundary

A tenant registers one immutable descriptor: engine/version identity, command/snapshot/result
schema references, shipped modes, a sorted closed rejection taxonomy, and a complete map of frozen
scaling inputs to `power`, `breadth`, or `presentation` destinations. The registry rejects duplicate
engines, duplicate modes/errors, `live_pvp`, unknown destination classes, and incomplete scaling
maps. Creation, play, and terminal-result validation dispatch by the exact frozen
`(engine_ref, engine_version)` pair. If a deployment no longer carries that version, the session
defers unchanged; a newer engine under the same reference cannot execute it under the old label.

Tenant creation and transitions are pure calls. Their only inputs are mode, seed or revision,
canonical snapshot/command JSON, and a defensive copy of the frozen exact-integer scaling map.
They cannot emit economy deltas. A terminal result is limited to a mechanical outcome, sorted
typed integer score facts, and an optional exact-integer rating delta; payout remains platform
owned. Descriptor schema references are backed by tenant validators invoked before and after every
call. Noncanonical or wrong-schema snapshots, undeclared rejection codes, malformed results, and
unknown modes fail closed as rejection or tenant divergence. JSON numbers have one accepted
grammar: exact safe integers only; decimal/exponent aliases are rejected before the JSONB seam.

The current conformance tenant is test-only. The combat duel adapter will register when its engine
RFC supplies an implemented transition surface; the deferred lane engine is not fabricated by the
platform.
