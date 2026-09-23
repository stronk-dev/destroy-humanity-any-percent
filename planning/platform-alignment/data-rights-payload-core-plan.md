# D-008/D-009/D-015 — retained core-payload lineage plan

Status: predeclared research only; no rights policy or product edit.
Coordinate: current product source `7e8aa70`, planning through `f09ba51`.

## Question and selected population

Trace the exact format, producer, reader, identity-bearing fields and deletion
disposition of the retained core payload lane: `save_revisions.state`,
`events.payload`, `intent_records.receipt`, `run_log.canonical_payload`/
`replay_inputs`/`receipt`, the corresponding `run_log_archive` encoding, and
`transport_player_outbox.payload`. These are the authoritative save/replay and
socket paths most likely to hide player-linked nested data after SQL-column
classification. This is a bounded tranche, not all JSON/bytea/text columns.

## Method and controls

1. Derive exact table field types and archive envelope from current Up
   migrations. Trace writers and readers through versioned save state, typed
   event payloads, command logging/compaction, idempotent receipts, and outbox
   event/receipt wrapping. Enumerate the schema/version or dynamic boundary;
   never pretend one sample represents all events or versions.
2. Run cold relevant non-DB tests and, if an existing real-Postgres witness
   covers archive/outbox, run it cold. Name exactly which records are populated
   and whether deletion is joined into that test.
3. Positive controls: retained Founder/stream/run identifiers, actor or subject
   IDs and nested account IDs must be surfaced. Negative controls: a hash is
   not itself the command payload; a transport `published_at` flag does not
   delete payload; `run_log` compaction must be traced into the archive rather
   than misclassified as erasure. Search for recovery credentials in these
   producer schemas without asserting absence beyond the inspected set.

## Exit and refusal

Produce a bounded payload-lineage dossier, update the parent inventory,
backlog for new contradictions, D-008/D-009/D-015 evidence, 1.0 board and
append-only logs. This may frame options but cannot adopt export scope,
deletion/retention semantics, player copy or a legal conclusion. No product,
test, migration, accepted RFC or authored-content change in this wave.
