# Soul foundation

Soul is a Founder-scoped exact-integer axis that remains mechanically independent from Trust. The
runtime foundation is implemented, but no production epoch contains a `soul` artifact and no
production debit source or recovery activity is enabled. The complete recovery lifecycle is ready
for content activation once its implementation passes the closing independent review.

## Artifact, state, and activation

The strict schema-v1 artifact owns the bounded policy, heartbeat and session-wall ceilings, complete
ordered bands, debit sources, recovery activities, and ending variants. Go and TypeScript enforce the same exact keys, closed
enums, ordering, interval coverage, safe-integer domains, copy-key references, and fixture-versus-
epoch row rules. The artifact participates in constants identity and requires the bumped minigame
and pet artifacts plus Fiscal v19.

Founder save v20 activates only at a new-run boundary under a pinned Soul artifact. It retains the
existing `soul` field, bounds it by the pinned policy, and adds exactly one persisted collection:
the sorted `soul_exhausted_source_ids` eligibility set. The Exit replay arm freezes and recomputes
the next initial Soul and band from the exact result artifact. Pre-v20 saves require dormant Soul
state; Company codecs reject Founder-only versions.

## Debit and consumer gates

The store-free debit component is callable only inside an owning transaction. It proves source
identity, owner type, eligibility, full affordability, and single-use exhaustion before returning
the exact before/debit/after values and bands. Owner transitions emit their benefit first, then
`soul_price_paid.v1`, optional `soul_band_changed.v1`, and the first exact-floor
`soul_depleted.v1` plus dated `soul.depleted` fact. There is no standalone public payment intent.

The Soul package exports a pure band projection and `human_content_locked` predicate. New pinned
minigame rows classify `human_hobby|unrelated`; new pet action rows classify
`essential|recovery|ordinary`. The composed server resolves Founder state and the exact pinned
bundle. Near-zero Soul blocks human-hobby minigames and ordinary pet care while preserving
unrelated, essential, and recovery paths.

## Recovery persistence and replay

`soul_recovery_sessions` is the authoritative UUIDv7 lifecycle, with one active-or-claimed row per
Founder, immutable identity, request hashes, a recoverable claim lease, and durable terminal
receipt. Start serializes against other exclusive sessions and writes its Founder event. Cancel or
resolve use the established Founder-then-Company lock order and one transaction for the terminal
session, Company suppression revision/log, Founder revision/log, events, receipt, and outbox.
Literal retries return the committed receipt; mismatched reuse fails closed.

Start returns a distinct opaque progress capability. Calling start while a recovery is already
active reconnects to that session and rotates the capability, invalidating stale clients. A
server-stamped progress command mutates only the session row: gaps at or below the pinned beat
ceiling add attended milliseconds; longer gaps add zero and merely reset the watermark. Duplicate
beats are harmless, absence pauses progress, and beats produce no resources, Soul, events, or replay
bytes. Resolve becomes eligible when the accumulated progress reaches the activity duration.

Sessions beyond the pinned wall-age ceiling are cancelled lazily during the next recovery
coordinator preflight. The watchdog uses the same atomic zero-Soul cancel path, ends suppression at
the last progress coordinate, and records `cancelled_by: watchdog`; player cancellation records
`cancelled_by: player`. Ordinary Company commands never execute the watchdog from inside their
Company lock. They remain read-only rejections and add `session_expired: true` when coordinator
cleanup is due.

The shared Go/TypeScript `ApplySuppressedLogged` boundary advances Company time and hook watermarks
through the frozen suppression interval while restoring every output-bearing authority. It asserts
zero ledger, provision, stock, guild, meter, achievement, and lifetime-value output. Founder replay
applies the exact saturation and event ordering. Founder history links the audit row to the Company
run-log witness and loads events by applied revision, so the separate start event cannot contaminate
terminal replay parity. Real-Postgres fault injection covers every write boundary.

Heartbeat cadence is intentionally absent from replay. Terminal Founder inputs freeze the accumulated
attended total, and Company replay freezes the suppressed wall interval; live and replay validate the
same totals without turning coordinator presence traffic into immutable gameplay commands.
