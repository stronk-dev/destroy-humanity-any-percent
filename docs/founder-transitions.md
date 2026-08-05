# Founder-scoped transitions

`save.Store.ApplyFounderLogged` is the reusable authoritative mutation boundary for Founder-only
mechanics. It is deliberately separate from Company `production.ApplyLogged`: a Founder command
cannot read or spend live Company state, and a Company command cannot smuggle a Founder write.

The Store locks one active Founder-scope stream, assigns the next immutable Founder-log sequence,
and stamps the command with database server time. The callback receives only the restored Founder
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
inputs. Pet Care is the first consumer; Soul verbs, Founder ratings, and other Founder mechanics
must reuse this boundary rather than adding side writes or extending the Company transition.

The feature package still owns its closed canonical command, resolved-input, receipt, event, and
state-transition unions. The persistence layer validates the shared envelope and transaction; it
does not invent feature mechanics.

Legacy unlogged `Store.ApplyIntent` calls are rejected for Founder streams, so the boundary cannot
be bypassed by an older feature surface. The existing `buy_route_hint` Founder command is the first
production consumer: its immutable resolved inputs freeze the repaired Route Knowledge balance and
route-context version before the purchase. Founder event and receipt outbox rows retain Founder
scope and the exact applied or rejected revision.
