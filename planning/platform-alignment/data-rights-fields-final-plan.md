# D-008/D-009/D-015 field-disposition research — final seven game tables

Status: research only; no export, deletion or retention policy adopted.
Coordinate: product source `7e8aa70`; confirm no later product-path change at execution.

## Question and population

Map every current SQL column and account-deletion effect in Minigame (5), Soul (1)
and Transport (1): `minigame_sessions`, `minigame_session_commands`,
`minigame_faucet_window`, `minigame_create_receipts`, `minigame_command_receipts`,
`soul_recovery_sessions`, `transport_player_outbox`. These are the seven remaining
tables in the migration-derived 60-table denominator. Completing this tranche
does not classify nested JSON/binary/text or non-DB stores.

## Method and controls

1. Trace base and subsequent Up migrations through `00074`, current writers/readers,
   expiry/prune production callers and deletion FKs/hooks. Record nullable and
   legacy fields; do not equate an account FK with a complete rights outcome.
2. Run one cold real-Postgres integration population per family if available.
   A green producer/transport test is not proof of account deletion after
   populated minigame/recovery/outbox data.
3. Negative controls: distinguish account-scoped receipts and commands from
   shared match/session data, Soul recovery state from permanent entitlement,
   and transport delivery acknowledgement from data removal. Inspect payload
   shape and mark unknowns; do not silently claim nested payload classification.

## Exit and refusal

Publish the exact seven-table map, update the parent inventory, decision queue,
backlog for new contradictions, 1.0 board and append-only logs. This may inform
D-008/D-009/D-015 but cannot select owner/legal policy, export format, shared
session rights, retention duration or player copy. No product, migration, test
assertion, accepted RFC or authored content changes in this wave.
