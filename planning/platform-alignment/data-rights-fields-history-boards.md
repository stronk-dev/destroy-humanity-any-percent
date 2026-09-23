# D-008/D-009/D-015 — history and board field-disposition tranche

Research checkpoint: 2026-09-23. Product source `7e8aa70`; no later product-path diff at
execution. The declared population/method is in
[`data-rights-fields-history-plan.md`](data-rights-fields-history-plan.md). This covers the
nine run/Founder-history and five verification/board tables, **14 of the 60** game tables.
Together with the account/save tranche, 24 have SQL-column maps; 36 game tables, nested
JSON/binary payloads and non-DB stores still lack field-level classification. This is not an
adopted export, erasure or retention policy.

## Current-schema field map

The table births and Up amendments are in migrations `00002`, `00012`–`00014`, `00017`,
`00020`, `00030`–`00038`, `00051`, `00054`–`00057`, `00062`, `00065`, `00069`, and `00074`.
Later event-kind/version checks and board-variable/key constraints change accepted values,
not the field list except as identified below. Current fields are enumerated completely;
“retained” means the current account-deletion transaction does not erase the row, **not** that
every row has a prescribed indefinite legal retention period.

| Table | Exact current fields and account-deletion effect | Writer → reader; export/retention issue |
|---|---|---|
| `events` | `event_id`, `stream_id`, `revision`, `schema_version`, `kind`, `intent_id`, `constants_hash`, `occurred_at`, `payload`, `event_seq` — retained. | Production/projectors → ordered replay, transport and verifier. No current production expiry found. `payload` varies by event kind/version and is not classified by the column map; unlike several other history tables, no general SQL update/delete rejection trigger was found. |
| `run_log` | `company_stream_id`, `run_seq`, generated `run_id`, `seq`, `intent_id`, `canonical_payload`, `receipt`, `applied_revision`, `server_ts_ms`, `created_at`, `replay_inputs` — retained while active; **archive-backed compaction deletes only unreferenced live rows** after verification. | `ApplyLogged` → verifier and immutable archive. `replay_inputs` is nullable for historical rows but required on new inserts. `canonical_payload`/`replay_inputs`/`receipt` need nested classification. Rows referenced by `founder_log` survive compaction. The SQL trigger rejects update and delete without an archive. |
| `run_log_archive` | `run_id`, `company_stream_id`, `run_seq`, `terminal_seq`, `encoding`, `bytes`, `sha256`, `archived_at` — retained, update/delete forbidden by trigger. | Verified-run archiver → forensic replay. `bytes` is deterministic `gzip+json.v1` carrying pin, genesis, commands, resolved inputs, receipts and events; it is **not** just a hash. No expiry or player export. |
| `run_genesis` | `company_stream_id`, `run_seq`, `state`, `version`, `constants_hash`, `created_at` — retained, update/delete forbidden. | Run pin → verifier/replay. `state` is serialized save JSON bytes and needs versioned payload classification; preserving the row is a replay invariant. |
| `founder_log` | `founder_stream_id`, `seq`, `intent_id`, `canonical_payload`, `replay_inputs`, `receipt`, `applied_revision`, `constants_hash`, `server_ts_ms`, `created_at`, `source_company_stream_id`, `source_run_seq`, `source_run_log_seq` — retained, update/delete forbidden. | Founder command logger → replay and live-log FK preservation. The three nullable source coordinates point to exact Company log rows for cross-stream events. JSON/input bytes remain unclassified. |
| `founder_genesis` | `founder_stream_id`, `revision`, `state`, `version`, `constants_hash`, `created_at` — retained, update/delete forbidden; revision FK protects its save revision. | Founder log genesis → replay. `state` bytes need versioned payload classification. No adopted expiry/export. |
| `run_epochs` | `company_stream_id`, `run_seq`, `epoch_id`, `constants_hash`, `engine_version`, `build_vcs_hash`, `seed`, `pinned_at` — retained, update/delete forbidden. | Epoch/run pin → deterministic replay and archive. Pin identity is account-adjacent through retained stream/run IDs, not generic shared catalog data. |
| `run_frozen_contributions` | `company_stream_id`, `run_seq`, `source_id`, `slot`, `target`, `factor`, `created_at` — retained, update/delete forbidden. | Fiscal contributor freezer → pinned replay. Exact row identity and factor are per-run history; no adopted expiry/export. |
| `run_version_drift` | `company_stream_id`, `run_seq`, `observed_version`, `first_seen` — retained, update/delete forbidden. | Version guard → replay/board exclusion. This is per-run forensic metadata, not an account-row cascade. |
| `verification_queue` | `company_stream_id`, `run_seq`, `status`, `attempts`, `available_at`, `claimed_at`, `completed_at`, `verdict`, `last_error`, `claim_token` — retained. | Exit enqueue → lease/retry/verifier. Pending/claimed rows mutate; verified/dead rows reject update/delete by trigger. `last_error` can carry bounded diagnostic text and needs disclosure/retention classification. A retry schedule is not an expiry schedule. |
| `verification_dead_letters` | `company_stream_id`, `run_seq`, `verdict`, `detail`, `failed_at` — retained, update/delete forbidden. | Deterministic replay failure → operator diagnostics. `detail` text and per-run link survive; no cleanup owner/bound proved. |
| `verification_poison_dead_letters` | `company_stream_id`, `run_seq`, `attempts`, `detail`, `failed_at` — retained, update/delete forbidden. | Repeated transient/expired-claim failure → operator diagnostics. A terminal dead letter is not short-lived by virtue of its name; no expiry proved. |
| `verification_projection_events` | `event_id`, `claimed_at` — retained, update/delete forbidden. | Board projector idempotency → duplicate suppression. Event UUID links to `verified_runs`; no expiry proved. |
| `verified_runs` | `run_id`, `event_id`, `founder_id`, `category_id`, `variables`, `epoch_id`, `mandate_level`, `key_ms`, `key_int`, `key_exponent`, `key_mantissa`, `verified_at`, `world_first` — retained, update/delete forbidden. | Verified projection → internal board queries/world-first arbitration. Current `variables` has bounded `commons`/`advisor`/`glitched`/`faction` keys; rank key uses exactly one supported representation. `founder_id` and `run_id` remain linkable after account unlinking. No mounted public board reader yet; storage is not a player-facing board/export. |

No field above is automatically removed by `DeleteAccount`: the account transaction archives
Founder-owned streams and nulls `account_founders.account_id`, but does not traverse these
history/board rows. A later cold joined probe now confirms this for **one directly seeded board
row**, not for the full history/dead-letter population or public reader.

## Executed controls and limitations

- Cold real-Postgres `TestVerifiedArchiveCompactionIsDeterministicAtomicAndImmutableIntegration`
  passed with `-count=1 -v`. It checks deterministic bytes/hash, compaction, rollback and
  archive immutability. `TestQueueProjectorCategoriesVariablesPreTimerAndRetryIntegration`
  separately passed cold and exercises board projection/retry. These are separate properties;
  neither original test deletes an account with populated history and boards.
- The first combined Make invocation used a `|` alternation in `SAVE_TEST_FLAGS`; the Make
  recipe expanded it as a shell pipe, yielded `command not found`/broken pipe and exited 2.
  It is **invalid evidence**. Two separate simple-name invocations then passed. This is a
  command-shape limitation, not a claim that the tests failed.
- A later temporary extension of the cold Account API test inserted one schema-valid board row,
  then executed the real deletion route. The row remained linked to an archived/unlinked Founder
  and two archived streams. See [`data-rights-board-delete.md`](data-rights-board-delete.md).
  It was a relational seed, not a projector-produced run; the full fourteen-table history and
  dead-letter population still needs a joined end-to-end rights witness.
- Archive bytes, `run_genesis.state`, `founder_genesis.state`, event/log JSON and diagnostic
  `detail`/`last_error` text are **content-bearing**. SQL-column classification does not imply
  their nested contents are understood or safe to export/publish.

A later bounded producer/consumer trace of the five verification/board
tables' content-bearing fields is in
[`data-rights-verification.md`](data-rights-verification.md). It distinguishes
fixed deterministic verdict text from unredacted transient error text and
classifies the current `variables` keys without claiming a joined deletion
witness or public board. The later joined *relational* deletion result is
separately bounded in `data-rights-board-delete.md`.

## Owner/legal handoff

**D-008:** choose whether all these histories, archives, board rows and diagnostics belong in
player export, in what machine-readable form, and with what shared/forensic exclusions. A
current-state-only export may be selected, but its omission of retained history must be
explicit. **D-009:** disclose precisely what persists after account deletion, including
immutable archive/board links and possible operator diagnostics; a populated-delete witness
and nested payload audit are required before claiming completeness. **D-015:** choose purpose,
duration and cleanup/legal disposition for every retained family. Existing SQL immutability
cannot be silently waived or rewritten in an applied migration; a conflicting ruling would
need a separately accepted forward migration and replay-preservation design.

The release remains blocked on these choices, the other 36 tables/payloads, built player
export/deletion workflows and R-003. No player rights outcome is completed by this research.
