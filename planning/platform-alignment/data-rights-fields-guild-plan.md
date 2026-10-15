# D-008/D-009/D-015 field-disposition research — Guild

## Current shared-consumer deletion supplement — 2026-10-15 at 0c54a6fa

Use the existing composed gameserver's real Guild create/join and clearing worker.
Its stock/faction preparation is controlled fixture setup, not earned-stock evidence.
After every original clearing/New-Founder/GC assertion, issue a fresh session for
the consumer and use the authenticated Account DELETE. Select the surviving
producer's real clearing rows and exchange events that name this consumer before
deletion; capture their immutable numeric keys. Re-query by those captured keys
afterwards, not by the reference that deletion might correctly erase. Clearing
identity includes guild, boundary and producer account; boundary alone is not
unique. Require the captured-key and before-row counts to agree, and exercise
that guard against the earlier boundary-only read.

Observe counts/content equality and nested allocation-reference counts, without
requiring retention as the desired outcome. Before populations/references must be
nonzero; nonexistent keys must produce zero. The producer's own empty-allocation
synthetic reservation row is excluded by the preselection. Observe the consumer's
own settlement rows separately, since their FK can cascade even when another
account's payload retains its UUID. Verify live access before DELETE and401 after.
No raw account/token/payload/digest output. Run the owning composed parent on real
Postgres, retaining its original assertions. No product, policy, schema, migration,
timing, content or CI change; no legal classification or whole-Guild/export/restore
claim. This supplements RP-120's old producer-deletion result, not a new ruling.

## Historical SQL-column tranche

Status: research only; no export, deletion or retention policy adopted.
Coordinate: product source `7e8aa70`; confirm no later product-path change at execution.

## Question and population

Classify all current SQL columns and account-deletion effects in the 13 Guild tables named
by `data-rights-inventory.md`. These tables mix one account's membership/applications with
other members' shared history, privacy-sensitive presence and retry/outbox data. With the
previous three tranches, completing this map will cover 53/60 game tables. Minigame (5),
Soul (1), Transport (1), nested payloads and non-DB stores remain separate.

## Method and controls

1. Derive current fields, FKs and constraints from every relevant Up migration through
   `00074`, including later ALTERs. Trace Guild writers/readers, deletion hook, event and
   presence relay, and any composed cleanup. Do not infer deletion from an FK name alone.
2. Run the real-Postgres Guild/account deletion witnesses cold where available; state their
   fixture population and refusal boundary. A passed live-membership test does not imply
   outbox failure-path or shared-history cleanup.
3. Negative controls: a departed account's `guild_presence_outbox.account_ref` and shared
   guild events cannot be omitted; account-scoped applications/invitations must be kept
   distinct from shared Guild history. Check whether any field holds free-form player text.
   A source trace found `guild_events.payload` serializes `Clearing` with producer and consumer
   account UUIDs. Before claiming relational FK anonymization is complete data unlinking,
   temporarily inspect the existing real-Postgres Guild deletion fixture after it commits a
   clearing and deletes that producer account. Count surviving JSON UUIDs and remove the
   diagnostic edit byte-identically; do not change permanent assertions or deletion behavior.

## Exit and refusal

Produce an exact 13-table SQL-column map, current deletion outcome and unknowns; update the
parent inventory, decision queue, shared backlog for new contradictions, and append-only
logs. This work may inform D-008/D-009/D-015 but cannot choose legal basis, shared-member
export, moderation policy, retention duration, synthetic Guild behavior or player copy.
No product, migration, test assertion or authored content changes in this wave.
