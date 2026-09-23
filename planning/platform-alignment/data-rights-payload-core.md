# D-008/D-009/D-015 — retained core-payload lineage

Research checkpoint: 2026-09-23. Product source `7e8aa70`, planning after
`f09ba51`. Method and controls:
[`data-rights-payload-core-plan.md`](data-rights-payload-core-plan.md). This
traces the selected save/replay/transport payload lane, **not** all JSON,
bytea or free-text fields in the 60-table schema. It is not an export schema,
privacy notice, retention policy or authorization to rewrite history.

## Producer → representation → consumer

| Payload | Exact present format and current lineage | Identity, deletion and unknowns |
|---|---|---|
| `save_revisions.state` | `jsonb` object written by `save.EncodeStateVersion` and read by versioned restore. `save_revisions.version` is a separate SQL field; current latest Company/Founder encoders are 18/21, while historical versions remain readable. The wire carries balances, generators, run/Founder mechanics and later feature state according to scope/version. The latest five ordinary revisions are retained, with protected genesis history separate. | `stream_id` joins an archived Founder stream after account deletion. Embedded fields can include Guild, faction, route, pet, opportunity and run history. A single current-version sample cannot classify every historical version. No account recovery code or session token field was found in the inspected save-state encoder; that is a scoped source negative, not a whole-database secret audit. |
| `events.payload` | `jsonb` object attached to `kind`, `schema_version`, stream, revision and optional intent. The current SQL version constraint permits `run_ended` v1–v3, two opportunity/buff kinds v1–v2, and other kinds v1. Producers include ordinary intents and later projectors/coordinators. Event rows feed replay, projections, verification and a transport enqueue trigger. | Stream/Founder/run coordinates and kind-specific nested IDs can remain after account deletion. No universal payload shape is defined by `jsonb_typeof='object'`; a kind/version-specific inventory is still required. Events sharing a run command's intent are additionally copied into that run's archive. |
| `intent_records.receipt` | `jsonb` object from the accepted/rejected decision, normalized and written for idempotent replay. For Founder/Company intents the same receipt is enqueued as a transport `receipt`; logged commands also copy it into run/Founder log. `request_hash` is only a digest, not a substitute for these full persisted bytes. | The retained stream joins the archived Founder. Receipts vary by command and outcome; rejected receipts also persist. No composed production intent-pruner caller was found in the prior account/save tranche. A receipt-family schema inventory remains open. |
| `run_log.canonical_payload` and `replay_inputs` | `canonical_payload` is bytea containing validated canonical JSON command bytes (its SHA-256 must equal `request_hash`). `replay_inputs` is a `jsonb` v8 envelope: `{v,command,evaluated_at_ms,evaluation_mode,offline_catchup?,resolved}`. `command` holds `intent_id`, `company_stream_id`, `founder_id`, `revision`, `run_seq`, `run_log_seq`; `resolved` is a kernel-owned per-intent object, not globally one fixed schema. Old pre-v8 rows may have null `replay_inputs`; new inserts require it. | Both payloads retain exact player command/replay inputs. `founder_id` and stream ID are explicit nested identity links. There is no license to export arbitrary canonical commands without classifying the per-kind `resolved` union and any other referenced party. |
| `run_log.receipt` → `run_log_archive.bytes` | `receipt` is the same decision JSON. On verified compaction, `replayverify` builds an immutable `gzip+json.v1` archive with `schema_version`, `company_stream_id`, `run_seq`, epoch/catalog pin, genesis `{version,state}`, and ordered entries. Each entry includes `seq`, `intent_id`, `canonical_payload`, `replay_inputs`, `receipt`, `applied_revision`, `server_ts_ms`, and matching Company/Founder events with their full `payload`. A SHA-256 covers the compressed bytes. Unreferenced active `run_log` rows are then removed; Founder-log-referenced rows remain active as well. | Compaction preserves complete command, state, receipt and event content in archive bytes; it is **not** personal-data erasure. Account deletion archives the stream rather than deleting archive rows. Only events matched to archived commands by intent are included; this is not a claim that every event has an archive twin. |
| `transport_player_outbox.payload` | A `receipt` row holds the receipt JSON directly. An `event` row is a JSON wrapper with `event_id`, `kind`, `scope`, `rev`, `cursor_effect` and nested source `payload`. The event trigger enqueues from Founder-owned Company/Founder streams. Receipts have a 60 KiB bound; after migration `00042`, authoritative event payloads may exceed that cap and enter a bounded dead-letter path. `MarkPlayerPublished` only sets `published_at` and clears the claim fields. | Founder/stream/source identifiers remain as SQL columns and in event wrappers; nested event or receipt body remains even after publication/dead-letter. Account deletion retains the stream and therefore these rows (RP-122). A socket acknowledgement is not a retention rule or a deletion proof. |

The source search of the selected save/production/replay-verify producer files
found no `recovery_code`, `refresh_token` or `access_token` field in these
payload encoders. It does **not** establish that all current or future dynamic
intent/event payloads are secret-free, nor classify browser credential storage
(separately mapped in `data-rights-browser.md`).

## Executed evidence and limits

Cold real-Postgres `TestVerifiedArchiveCompactionIsDeterministicAtomicAndImmutableIntegration`
(`./replayverify`) and `TestPlayerOutboxOrderingDeadLetterAndSizeIntegration`
(`./save`) passed with `-count=1 -v`. The non-DB
`TestValidateReplayInputsPinsOfflineCatchupCoordinates` passed with `-count=1 -v`.
The archive witness checks deterministic compression, compaction and immutable
archive behavior; its fixture uses `{}` for the logged command, replay inputs
and receipt, and expects zero events, so it **cannot** prove preservation of
nontrivial payload fields or matched Company/Founder event content (RP-124).
That content lineage is source-derived and needs a populated, severable witness.
The outbox witness checks ordering/failure/size. Neither
deletes an account after populating all selected payloads. The source chain,
not a joined deletion witness, establishes which streams remain after account
removal. The ephemeral Postgres service/network was removed; named Go cache
volume preserved.

## Decision handoff

- **D-008:** current-state-only export omits full command/receipt/event and
  archive history. Decide whether any of these are included, how versioned
  replay/catalog dependencies travel, and how other-party fields are filtered
  without corrupting an integrity-preserving archive.
- **D-009:** explain retained versioned saves, full verified-run archive and
  published/dead-lettered event/receipt payload after account deletion. Do not
  equate live-row compaction or transport delivery with erasure.
- **D-015:** rule purpose, trigger, duration and cleanup owner for these
  linked payload lanes, respecting replay verification and immutable archive
  constraints. The uncalled intent-prune helper and unbounded outbox history
  need an implementation contract, not a label on a diagram.

Later bounded maps now cover the core-event registry/selected identity fields
in [`data-rights-events.md`](data-rights-events.md) and selected core,
Minigame, Guild and Soul receipt stores in
[`data-rights-receipts.md`](data-rights-receipts.md). Exact historic event
encoders, the full receipt union, Founder-log payload variants, non-receipt
Minigame/Soul/Guild JSON, verification dead letters and joined rights proofs
remain open. This original tranche does not close them.
