# D-008/D-009/D-015 — current persisted-data inventory

Research checkpoint: 2026-09-23. Product source: `7e8aa708ff9e002e7ec4b46242d7d889f76b7fdd`;
the later commits through `70784af` are planning-only. Method and limits were predeclared in
[`data-rights-inventory-plan.md`](data-rights-inventory-plan.md). This is a source/runtime
inventory for decisions, **not** a retention policy, privacy notice, legal opinion, export
schema, or authorization to delete replay history. `[V]` means directly observed in current
source or an executed Postgres test; `[P]` means a configured operator mechanism not yet proved
on the supported clean host; `unknown` means no verified production bound.

## Denominator and controls

The ordered `server/save/migrations/00001`–`00074` Up chain creates 61 tables and drops the
historical `transport_receipt_outbox`, leaving **60 game tables**. A cold migration in the
declared Postgres integration service returned 61 public tables: these 60 plus Goose's own
`goose_db_version`. The 11 disjoint groups below contain exactly 60 names, with no duplicate,
missing or extra name when compared to the migration-derived set. Neither the dropped outbox nor
Goose's migration metadata is counted as a current game-data family. The short-lived credential
families and immutable history families are included, so the census cannot accidentally sample
only convenient-to-delete rows. [V]

Every group names its live tables. Writer/reader are the production package or workflow owning
the family, not a claim that every individual table has a player-facing reader. `No export`
means there is no mounted player export endpoint or Game UI action; it does **not** mean an
operator could not query the database. A player-deletion outcome is stated only where traced;
otherwise it is `not established`, never assumed to cascade.

| Family (tables) | Producer → consumer; link / deletion / retention at this source |
|---|---|
| **Account and credential (7):** `accounts`, `account_emails`, `account_founders`, `session_families`, `sessions`, `access_tokens`, `bootstrap_receipts` | Account/auth API and minute GC → login/bootstrap/authenticated routes. `account_id` links the family. Account deletion removes account/credential rows, unlinks and archives Founder mappings, and tombstones bootstrap ciphertext; the bootstrap row keeps `account_id`, digest and times permanently (confirmed below). Expired live credentials/ciphertext have a composed minute cleanup; no anonymous-account inactivity cleanup or tombstone expiry. No export. [V] |
| **Save and intent (3):** `save_streams`, `save_revisions`, `intent_records` | Save/Production → authoritative load, replay and idempotent intent response. Streams join by owner/Founder; deleting an account archives but keeps its streams and revisions. Ordinary revisions are bounded to the latest five; protected genesis references persist. The package-level `PruneIntentRecords` has **no non-test production caller**, so the documented 30-day intent policy is not current behavior. No export. [V] |
| **Run/Founder history (9):** `events`, `run_log`, `run_log_archive`, `run_genesis`, `founder_log`, `founder_genesis`, `run_epochs`, `run_frozen_contributions`, `run_version_drift` | Production/Save/Prestige/replay → authoritative replay, run identity and version evidence. Founder, stream and run identifiers can relate records to a person or archived Founder; account deletion does not erase this history. Active `run_log` rows may be deleted only after archive-backed verified compaction (some FK-referenced rows remain); the immutable archive retains full command/input/receipt/event content. Other immutability guarantees vary by table. No adopted expiry or player export. [V] |
| **Verification/boards (5):** `verification_queue`, `verification_dead_letters`, `verification_poison_dead_letters`, `verification_projection_events`, `verified_runs` | Verification projector → verified boards and operational retry/diagnostic flows. A cold joined Account API probe retained a seeded board→archived-Founder→archived-stream link after account unlinking (RP-130); it did not populate dead letters or use a public reader. Dead-letter/verified history has no adopted retention or player export; exact per-column deletion disposition remains to be ruled. [V] |
| **Catalog/epoch (4):** `catalog_sets`, `catalog_artifacts`, `epochs`, `epoch_hashes` | Catalog/epoch publisher → replay/restore authority. Primarily shared content/version records, not an account export by default; historical state can depend on these rows. Their preservation bound and inclusion in a portable import bundle are open. [V] |
| **Routes (5):** `route_projection_events`, `founder_route_executions`, `founder_route_state`, `route_hint_projection_events`, `registry_routes` | Route registry and execution → route state, hints and projection. Founder-linked rows remain a distinct export/deletion subject; registry definitions are shared authority. No adopted family-wide expiry/export. [V] |
| **Commons (7):** `commons_cohorts`, `founder_commons_assignments`, `company_compact_memberships`, `commons_projection_events`, `commons_member_samples`, `commons_health_scopes`, `commons_recruitment_offers` | Commons producer/projector → cohort and membership accounting. Founder/company IDs and samples are joinable; shared cohorts are not simply an account row. A cold joined Account API probe archived an active Founder's stream while the exact World member-count and health sample-selection predicates still returned one; `member=false` made both zero (RP-119). Full projection/recompute/player proof and deletion/export/expiry contract remain absent. [V for bounded SQL result] |
| **Guild (13):** `guilds`, `guild_members`, `guild_applications`, `guild_invitations`, `guild_account_revisions`, `guild_intent_records`, `guild_events`, `guild_health_inputs`, `guild_exchange_boundaries`, `guild_presence_outbox`, `guild_projection_events`, `guild_activity_windows`, `guild_clearing_results` | Guild intents/projector/presence relay → shared Guild history and membership. Deletion closes memberships and removes that account's applications/invitations; Guilds and event history remain. FK-linked account columns null or cascade by table. A cold diagnostic found the deleted UUID retained in `guild_events.payload` despite nulled actor FK (RP-120). `guild_presence_outbox.account_ref` can retain the UUID until successful publication, when it is nulled; no bounded failure/dead-letter retention has been shown. Shared-member disclosure/export needs a separate rule. [V] |
| **Minigame (5):** `minigame_sessions`, `minigame_session_commands`, `minigame_faucet_window`, `minigame_create_receipts`, `minigame_command_receipts` | Minigame API/tenant → session and idempotent command result. These rows refer to Founder/streams/sessions that account deletion retains, so their cascade FKs do not fire on account removal (RP-122, source-derived). Nested JSON/command bytes and expiry/export remain open. [V for source trace] |
| **Soul (1):** `soul_recovery_sessions` | Soul recovery API → resumable recovery state. Founder/stream-linked active or terminal session survives the account deletion path by source/FK trace (RP-122); nested receipt and expiry/export require decisions. [V for source trace] |
| **Transport (1):** `transport_player_outbox` | Transport producer → player socket/recovery delivery. The Founder stream remains after account deletion, so queued/published/dead-lettered rows remain by source/FK trace (RP-122). Publishing updates the row; it does not remove nested receipt/event payload or settle retention. [V for source trace] |

The groups are a **relation-complete census**, not a field-by-field personal-data classification.
For all families, a future contract must specify nested payloads' inclusion and
erasure disposition. The selected save/replay/transport lineage is traced in
[`data-rights-payload-core.md`](data-rights-payload-core.md): verified-run
compaction retains full commands, receipts, replay inputs, genesis and matched
events inside an immutable gzip archive, while transport publication leaves its
payload row intact. The bounded core-event registry and selected run/Founder,
Minigame and Soul identity fields are traced in
[`data-rights-events.md`](data-rights-events.md): 48 kinds, 52 SQL-accepted
kind/version pairs, 51 Go-write-validator-admissible pairs, and tested Soul
event→outbox copying. This is not a complete event-key or historic-encoder
classification. The selected core, Minigame, Guild and Soul receipt/response
stores are mapped in [`data-rights-receipts.md`](data-rights-receipts.md):
Guild per-account receipts cascade on account deletion, whereas retained
Minigame/Soul receipts have distinct copies and no adopted expiry; terminal
Soul rows retain progress tokens (RP-127), and Minigame API receipt triggers
block parent-session cascade cleanup (RP-128). Other payload families remain
unclassified. The five verification/board tables' selected diagnostics,
variables and marker joins are traced in
[`data-rights-verification.md`](data-rights-verification.md): deterministic
dead-letter text is fixed, but transient/poison details pass bounded,
unredacted upstream error text into immutable history (RP-129). No group gets
a made-up duration from its package name or apparent lifecycle.

The account/save, history/board, shared catalog/Routes/Commons, Guild and final
Minigame/Soul/Transport SQL-column tranches are in
[`data-rights-fields-account-save.md`](data-rights-fields-account-save.md),
[`data-rights-fields-history-boards.md`](data-rights-fields-history-boards.md) and
[`data-rights-fields-shared.md`](data-rights-fields-shared.md), and
[`data-rights-fields-guild.md`](data-rights-fields-guild.md), and
[`data-rights-fields-final.md`](data-rights-fields-final.md). Together they cover **60 of 60**
game tables' SQL columns. Nested JSON/binary/text payloads and non-DB stores are
**not** fully field-classified; this relation-complete census remains the parent inventory.

## Other persisted surfaces

| Surface | Current behavior and evidence boundary |
|---|---|
| Browser `localStorage` | Four exact key families are mapped in [`data-rights-browser.md`](data-rights-browser.md): bootstrap retry key, credential document, per-Founder transport positions and per-Founder personal-best/split history. `GameUIRuntime` persists the one-time recovery code alongside access/refresh tokens without a current display/copy/download consumer (RP-123); no player delete/export action or local account-deletion cleanup exists. No other production client storage API was found in source. This is device-local state, not a Postgres cascade. [V for source/unit test] |
| Encrypted off-host backups | The exact fields/sinks are in [`data-rights-operator.md`](data-rights-operator.md). Six-hour and pre-upgrade encrypted full-DB backups retain every six-hour object seven days and one daily through day 30, but newest valid/unresolved pre-upgrade copies are protected beyond that; 30 days is not a hard maximum. Full pre-deletion restore can recreate an account without a deletion replay (RP-125, source-derived). Local real-Postgres empty/populated backup→restore passes, not a deletion/restore or supported-host R-006 witness. [V for code/tests, P for host] |
| Application/security journal | All seven services use tagged journald; gameserver structured calls inspected are bounded, with a 14-day measurement/budget contract that must alert before pressure can evict evidence early. No raw-IP producer is shipped. Actual host retention/purge and every container's message fields remain unobserved; incident extracts need an access/disclosure rule. [V for source, P for host] |
| Private metrics/alerts | Prometheus is private with explicit 30-day TSDB retention; Alertmanager's persistent volume has no explicit retention flag in the template. Inspected gameserver/host metric labels are bounded, not a proof about every scraped collector or receiver. Alert/metrics volumes remain operator data and clean-host validation is pending. [V for source, P for host] |
| Release/operator artifacts | Release/rotation JSONL ledgers store identifiers, digests, times, outcomes and a caller-supplied `operator` identifier; R-006/recovery artifacts contain host/tool/version/timing and population hashes rather than full rows. The operator field lacks an adopted lifecycle (RP-126). Exact retention/access/disclosure and clean-host R-006 evidence remain open. [V for code, P for host] |

The earlier `operations-retention-preservation-audit.md` is **dated at product `190a4fa`**:
its “no metrics surface”, “no backup”, and “no configured journal sink” findings are superseded
by the constructed Deployment package, subject to R-006/R-007 host proof. Its orphaned intent
pruner, absent inactive-account cleanup and incomplete product retention findings still hold at
this source. The earlier account-rights audit's missing export/player deletion surface also holds.

## Executed deletion diagnostic and blind spot

The existing `TestAccountSessionIntegration` inserts a live bootstrap receipt, deletes the
account through `DELETE /api/v1/account`, and asserts **zero live bootstrap secrets**. I added a
temporary post-delete count and ran
`make test-save-integration SAVE_TEST_PACKAGES='./account' SAVE_TEST_FLAGS='-run TestAccountSessionIntegration -count=1 -v'`
cold in the declared Postgres service. It reported `retained_bootstrap_tombstones=1` and passed.
I then removed the
temporary diagnostic edit. Source agrees: `DeleteAccount` only nulls receipt secrets, migration
`00073` forbids row deletion and changing its `account_id`, and the pruner also only tombstones.
Thus the permanent account-linked UUID is a verified remaining row, not an inference from the
method name. The current AC6 witness is valid for its secret-removal claim, but **not** for a
claim that no account-linked data remains. [V]

The trace controls also fired as intended: the migration-derived set excluded the dropped
`transport_receipt_outbox`; the production search found only the declaration of
`PruneIntentRecords`; the mounted Account router has `DELETE /account` but no export route; the
Game UI Settings has no account export/deletion action. These negatives prevent promoting a
backend primitive into a player outcome.

## Decision handoff — not decisions made here

1. **D-008 export:** choose current state versus revisions/events/verified and shared histories,
   browser-local timing, and portable import. An export must state treatment of other members'
   Guild/Commons data and retained backups; then an accepted schema/API/UI RFC and R-003 can test
   the actual player task. R-003 is downstream validation, not evidence required to choose D-008.
2. **D-009 deletion:** choose and author the exact preview of removed credentials/account rows,
   archived Founder/save/history/boards, Guild/shared records, bootstrap UUID tombstones,
   device-local keys and backup lag. The current backend endpoint is not a complete player right.
3. **D-015 retention:** supply purpose/basis and trigger/duration/cleanup owner for each of the
   11 groups plus browser, backups, journals, metrics and operator artifacts. Address the orphaned
   intent pruner, anonymous-account growth, bootstrap tombstone's immutable UUID and shared-data
   relay failure paths. Preserve replay integrity and migration append-only rules; do not simply
   schedule generic row deletion. Legal review is still required before a policy or public claim.

R-003 and R-007 remain open. This inventory supplies their denominator and a real negative
control, but cannot replace participant tasks, an accepted operations/retention contract, an
operator rehearsal, or owner-authored disclosure.
