# D-008/D-009/D-015 — Minigame, Soul and Transport field tranche

Research checkpoint: 2026-09-23. Product source `7e8aa70`; no later
product-path change at execution. Method and controls:
[`data-rights-fields-final-plan.md`](data-rights-fields-final-plan.md).
This maps the last seven current game tables, bringing the SQL-column
denominator to **60/60**. It does **not** classify every nested JSON/bytea/text
payload, browser storage, backups or operator artifacts, and it does not adopt
any export/deletion/retention policy.

## Exact current SQL-column map

Base Up DDL: migrations `00040`, `00049`, `00052`, `00058`, `00068`, `00071`;
later current-schema changes: `00041`, `00042`, `00051`, `00060`, `00070`.
Account deletion archives Founder mappings and save streams, then removes the
`accounts` row. `account_founders.account_id` becomes null but the Founder row
remains; archived save streams remain. None of these seven tables has an FK
directly to the deleted account. Their `ON DELETE CASCADE` references are to
Founder, save-stream or session rows that this account path does **not**
delete. Thus a cascade clause is not proof of deletion on account removal.

| Table | Exact current SQL fields; account-deletion disposition | Producer → reader / unresolved rights issue |
|---|---|---|
| `minigame_sessions` | `session_id`, `minigame_id`, `founder_id`, `company_stream_id`, `run_seq`, `engine_ref`, `engine_version`, `constants_hash`, `scaling_inputs`, `seed`, `mode`, `status`, `revision`, `genesis`, `state`, `result`, `claim_token`, `claimed_at`, `created_at`, `updated_at`, `resolved_at`, `resolution_receipt`, `resolution_company_revision`, `resolution_founder_revision`. Founder/stream FKs do not fire on account deletion; row remains, whether active or resolved. Legacy resolved rows may have the all-null resolution tuple; later resolutions require all three values. | Minigame API/repository → active/current session, replay, resolution. `scaling_inputs`, `genesis`, `state`, `result` and `resolution_receipt` are JSON requiring nested-field classification. `mode` allows solo/async-snapshot, not proof a current multiplayer player surface exists. |
| `minigame_session_commands` | `session_id`, `seq`, `command`, `applied_revision`, `server_ts_ms`, `created_at`. Immutable command bytes remain while parent session remains; FK cascades only if that session is actually deleted. | Minigame play writer → deterministic replay. `command` bytea may encode player actions; do not infer privacy from binary storage. |
| `minigame_faucet_window` | `founder_id`, `minigame_id`, `attended_day`, `quota_used`, `conversion_remainder_ppm`, `updated_at`. Founder remains after account deletion, so quota history remains. | Faucet writer → attended-day quota and remainder. This is per-Founder accounting, not anonymous aggregate data. |
| `minigame_create_receipts` | `founder_id`, `idempotency_key`, `request_hash`, `session_id`, `response`, `created_at`. Founder and session remain after account deletion; immutable receipt row remains. | API create → idempotent response. `response` JSON needs nested classification; `idempotency_key` is a caller-supplied identifier, not automatically harmless text. |
| `minigame_command_receipts` | `session_id`, `command_id`, `request_hash`, `response`, `created_at`. Session remains, so immutable receipt remains. | API command → replay-safe response. `response` JSON and caller-supplied `command_id` remain to classify. |
| `soul_recovery_sessions` | `session_id`, `founder_id`, `founder_stream_id`, `company_stream_id`, `run_seq`, `constants_hash`, `activity_id`, `founder_attended_start_ms`, `required_duration_ms`, `status`, `start_request_hash`, `terminal_request_hash`, `claim_token`, `claimed_at`, `terminal_receipt`, `created_at`, `updated_at`, `terminal_at`, `progress_token`, `attended_progress_ms`, `last_progress_server_ms`. Founder/streams remain after account deletion, so active or terminal recovery row remains. Progress fields were added without fabricated legacy backfill in `00070`. | Soul recovery API/coordinator → resumable progress and terminal replay. `terminal_receipt` JSON and progress/claim tokens need purpose/retention classification; a session row is not a permanent Soul entitlement. |
| `transport_player_outbox` | `outbox_id`, `founder_id`, `stream_id`, `message_kind`, `source_id`, `scope`, `revision`, `constants_hash`, `payload`, `occurred_at`, `claim_token`, `claimed_until`, `published_at`, `attempt_count`, `last_error`, `dead_lettered_at`. Stream remains after account deletion, so queued, published or dead-lettered row remains. Publication marks `published_at`; it does not delete the row or clear payload. | Save event/intent enqueue → player relay/recovery. `payload` nests receipts or source events and can include player state; event payloads are not bounded by the receipt's 60 KiB check after `00042`. `last_error` is capped diagnostic text, not proven secret-free by that cap. No production outbox-prune caller was found. |

Negative controls: Minigame command/create receipts are not the same thing as a
shared match session; Soul recovery progress is not a permanent entitlement;
transport publication is an update to a delivery row, **not** erasure. The
account-delete path retaining Founder and streams is the positive control
against assuming these seven `ON DELETE CASCADE` clauses fire (RP-122).

## Executed evidence and limits

Three selected real-Postgres integration populations passed with `-count=1 -v`
through the declared root Make target:

- `./minigame`, `TestSessionClaimIntegration`: session create/claim state.
- `./production`, `TestSoulRecoveryIntegrationAtomicSuppressionReplayAndExclusivity`:
  recovery coordination and replay behavior.
- `./save`, `TestPlayerOutboxOrderingDeadLetterAndSizeIntegration`:
  outbox ordering, delivery failure/dead-letter and sizing behavior.

They verify producer/consumer mechanics on current schema, not deletion of an
account after all three families have populated rows. RP-122 is a source/FK
trace; it still needs a joined real-Postgres deletion witness under accepted
rights semantics. There is no demonstrated retention/expiry worker for these
seven tables. The ephemeral test Postgres container/network was removed after
execution; the named Go cache volume was preserved.

## Decision handoff

- **D-008:** decide whether export includes Minigame inputs/commands/results,
  Soul recovery history, and transport receipts/events, including any shared
  session boundary and portable replay dependencies. A current-state export
  would omit retained histories unless disclosed.
- **D-009:** explicitly say that today's backend account deletion archives
  Founder/streams and leaves these rows. Choose desired current and historical
  outcomes before changing cascades, sessions, outbox or player-facing copy.
- **D-015:** choose purpose, trigger, duration and cleanup owner for active and
  terminal sessions, commands, idempotency receipts, faucet windows, recovery
  tokens, published/dead-lettered transport rows and their nested payloads.
  No general delete job is authorized by this map.

The **60/60 number is SQL-column coverage only**. Nested JSON/bytea/text,
device-local state, backups, journals, metrics and operator artifacts remain
separate field-level work, as do legal review and joined player rights proof.
