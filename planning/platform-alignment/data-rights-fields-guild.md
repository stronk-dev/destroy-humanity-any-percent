# D-008/D-009/D-015 — Guild field-disposition tranche

Research checkpoint: 2026-09-23. Product source `7e8aa70`; no later product-path
change at execution. Method and controls: [`data-rights-fields-guild-plan.md`](data-rights-fields-guild-plan.md).
This maps every current SQL column in the 13 Guild tables, bringing the game-table
field map to **53/60**. Minigame (5), Soul (1), Transport (1), nested payloads beyond
the checked clearing case, and non-DB stores remain open. It is not an adopted
export, deletion, retention, or legal rule.

## Current schema and deletion disposition

Base Up DDL is in migrations `00025` and `00027`; `00026`, `00028`, `00029`,
`00044`, and `00046` change current fields/FKs. The account deletion path calls
`guild.PrepareAccountDeletion` before deleting `accounts`. That hook closes active
memberships, handles leader succession/disbanding, emits events/presence, and
removes the departing account's applications/invitations. FKs then null or
cascade as specified below. This is **not** a blanket Guild erasure: other
members' shared state and historical rows remain.

| Table | Exact current SQL fields; departing-account disposition | Producer → reader / rights boundary |
|---|---|---|
| `guilds` | `guild_id`, `name`, `created_at`, `founder_account`, `join_policy`, `revision`, `guild_xp`, `below_floor_since`, `disbanded_at`. Shared row remains; `founder_account` becomes null if that account is deleted. | Guild create/update → membership/health/settlement readers. `name` comes from the player's create intent (3–24 chars), so it is player-submitted text still present in a shared row. No automatic name erasure or public-UGC surface claim follows. |
| `guild_members` | `membership_id`, `guild_id`, `account_id`, `joined_at`, `left_at`, `role`. Active departing membership is closed first; `account_id` becomes null on deletion, including already-closed histories. Membership/time/role remain under the history guard. | Guild intents/deletion → authorization, leader selection, presence counts, clearing. Another member's membership is not the departing player's disposable row. |
| `guild_applications` | `guild_id`, `account_id`, `created_at`, `resolved_at`, `admitted`. Deletion hook deletes that account's rows before the account FK would block deletion. | Apply/resolution → Guild admission. Other accounts' requests remain; no free-text body field. |
| `guild_invitations` | `guild_id`, `account_id`, `created_at`, `resolved_at`, `accepted`. Deletion hook deletes that account's rows. | Invite/resolution → admission. Other accounts' invitations remain. |
| `guild_account_revisions` | `account_id`, `revision`. Account FK cascades row deletion. | Guild intent revision → stale-intent rejection. Account-scoped protocol state, not shared history. |
| `guild_intent_records` | `account_id`, `intent_id`, `request_hash`, `outcome`, `receipt`, `created_at`. Account FK cascades row deletion; `receipt` is JSON not exhaustively classified here. | Intent handler → idempotent response. `request_hash` and receipt are account-scoped, unlike retained Guild events. |
| `guild_events` | `event_seq`, `event_id`, `guild_id`, `revision`, `kind`, `actor_account`, `subject_account`, `intent_id`, `occurred_at`, `payload`. Row remains; actor/subject FKs become null. `payload` can still contain account UUIDs (RP-120). | Intents, deletion and clearing → shared event history/revision. FK anonymization is not JSON anonymization. Exact nested payload policy remains D-009/D-015. |
| `guild_health_inputs` | `guild_id`, `window_start`, `active_founders`, `tithed_xp`. Shared aggregate remains. | Projector → health/mercy calculation. No direct account column; source contributions are not thereby anonymous in every upstream table. |
| `guild_exchange_boundaries` | `guild_id`, `boundary_id`, `committed_at`. Shared row remains. | Clearing boundary writer → idempotency. Boundary history is not an account-owned row. |
| `guild_presence_outbox` | `outbox_id`, `guild_id`, `account_id`, `kind`, `guild_revision`, `occurred_at`, `published_at`, `active_count`, `claim_token`, `claimed_until`, `account_ref`. Row remains; account FK becomes null, but `account_ref` preserves the UUID until successful publication clears it. | Presence writer/relay → Guild socket message. A failed/unpublished row has no shown bounded retention; RP-114 already tracks this. |
| `guild_projection_events` | `event_id`, `event_kind`, `projected_at`. Shared idempotency row remains. | Tithe/activity projector → duplicate suppression. Underlying source event/Founder identity may be joinable; this row is not directly account-keyed. |
| `guild_activity_windows` | `guild_id`, `window_start`, `account_id`, `xp`. Account FK cascades that account's rows; other members' windows remain. | Tithe/activity projector → Guild active-founder and XP inputs. |
| `guild_clearing_results` | `guild_id`, `boundary_seq`, `account_id`, `debit_units`, `credit_units`, `allocations`, `committed_at`, `snapshot_hash`, `founder_id`, `company_stream_id`, `run_seq`, `membership_id`. Account FK cascades that account's own row; the later run-identity and membership fields are nullable for legacy rows. | Clearing writer → settlement reader/replay attribution. `allocations` is JSON: a surviving producer's row can name a departing consumer via `Allocation.account_id`. That residual is source-derived, **not** separately counted post-delete in this tranche. |

Negative control: account-scoped applications, invitations, revisions, intent
records, activity windows and own clearing result do not have the same deletion
behavior as shared Guild rows. Positive control: retained `guild_events.payload`
contains account IDs even when relational actor/subject FKs become null. Neither
control chooses whether other members can export shared history or how long it
may be kept.

## Executed evidence and limits

The cold real-Postgres `make test-save-integration SAVE_TEST_PACKAGES='./guild'
SAVE_TEST_FLAGS='-run Integration -count=1 -v'` run passed all three Guild
integration tests. The lifecycle fixture covers clearing, account deletion,
leader succession, membership FK anonymization and pending presence reference;
it does **not** assert payload UUID erasure or eventual relay cleanup.

A predeclared, temporary diagnostic after that fixture's committed deletion
counted `guild_events` rows for its Guild with `kind='exchange_cleared'`,
`actor_account IS NULL`, and `payload->>'producer_account_id'` equal to the
deleted account. It found **one** on real Postgres while the test passed. The
diagnostic was removed byte-identically; no permanent test/product byte moved.
This is RP-120, not an assertion that all nested JSON has been classified.

The same unmodified suite failed when rerun against the **same populated test
database**: the lifecycle test returned two invalid Guild intents; the other
two tests hit duplicate `accounts_pkey` rows. After resetting only the
ephemeral Postgres service and running cold, all three passed. A second warm
rerun reproduced those failures. Fixed fixture IDs and no per-run database
reset explain the contamination (RP-121). Therefore a green cold run is valid
for its stated properties, but repeatability against a reused service is not.
The test container and ephemeral network were removed after this research;
the named Go cache volume was preserved.

No joined deletion fixture populates every Guild table, proves outbox failure
aging, or checks every nested receipt/event/settlement payload. No player export
or deletion control exists. Source traces and this one diagnostic cannot
substitute for a ruled policy and accepted, discriminating end-to-end rights
witness.

## Decision handoff

- **D-008:** choose export treatment for own Guild membership, Guild name,
  shared event/clearing history, other members' IDs, and settlement attribution.
  A single-account dump is not automatically a lawful or portable shared-Guild
  export.
- **D-009:** author the deletion preview for account-scoped removals versus
  retained shared Guild name/history, JSON UUIDs, and queued presence. Decide
  how identity references in shared settlements/events should behave without
  inventing an erasure that breaks other members' history.
- **D-015:** rule purpose, trigger, duration and cleanup owner for retained
  Guild rows/payloads and outbox failure paths. RP-120 requires nested-payload
  treatment; RP-121 separately needs an accepted test-isolation repair, not a
  privacy-policy ruling.

Seven game tables plus nested JSON/binary/text and browser/operator stores
remain outside the completed SQL-column map. Legal review remains necessary
before a public rights claim.
