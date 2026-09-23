# D-008/D-009/D-015 — account/save field-disposition tranche

Research checkpoint: 2026-09-23. Product source `7e8aa70`; no subsequent product-path
changes at this checkpoint. The declared population/method and refusal controls are in
[`data-rights-fields-plan.md`](data-rights-fields-plan.md). This covers **10 of 60** current
game tables, not the remaining 50 or browser/operator stores. It records current behavior and
decision subjects, **not** a lawful-basis analysis, adopted export schema, retention schedule,
erasure promise or authority to modify immutable history.

## Exact current-schema field map

The Up DDL is in migrations `00001`, `00002`, `00010`, `00016`, `00019`, `00051` and `00073`.
`00019` makes `account_founders.account_id` nullable/`ON DELETE SET NULL`; `00051` adds a
uniqueness constraint to `save_streams` but no field. No later Up migration through `00074`
adds/removes a field in these ten tables. Every field in the ten tables is listed below;
parenthesized groups share the same current deletion result, not necessarily the same legal
classification.

| Table | Exact fields and current account-deletion result | Producer / current reader / export and retention question |
|---|---|---|
| `accounts` | `account_id`, `recovery_hash`, `created_at` — row **deleted**. | Account create/recovery → authentication. Account ID and creation time are candidate account-export metadata; the credential verifier must be treated as a secret, not ordinary portable content. No inactive-account cleanup is established. |
| `account_emails` | `account_id`, `email`, `verified_at` — row **cascades away**. | Optional account email record → account identity; current preview has no email recovery. If any email exists, D-008/D-009 must address it explicitly. No independent expiry established. |
| `account_founders` | `account_id` — **nulled** by FK; `founder_id`, `created_at`, `archived_at`, `imported` — **retained**; `archived_at` set if absent. | Founder identity/mapping → account active-Founder resolution and imported-history exclusion. The surviving UUID, times and imported flag are pseudonymous history, not proof of irreversible anonymity. No deletion/expiry for these rows established. |
| `session_families` | `family_id`, `account_id`, `created_at`, `revoked_at` — row **cascades away**. | Rotation/revocation serialization → token validation. Expired families are otherwise collected only after both token tables are empty. Family/revocation metadata is an export/retention decision; no player export exists. |
| `sessions` | `token_hash`, `family_id`, `account_id`, `created_at`, `expires_at`, `consumed_at`, `revoked_at` — row **cascades away**. | Opaque refresh-token verifier → refresh/replay detection. The token itself is not stored here; the hash and lifecycle times are security data. 30-day token expiry plus composed minute collector exists, but expiry is not proof the row vanished at the exact instant. |
| `access_tokens` | `jti`, `family_id`, `account_id`, `founder_id`, `expires_at`, `revoked_at`, `created_at` — row **cascades away**. | Issued JWT inventory → pre-expiry revocation. Fifteen-minute token expiry plus composed collector exists. `founder_id` is another linkable UUID while live; do not export token/security identifiers by default without a ruling. |
| `bootstrap_receipts` | `key_id`, `nonce`, `ciphertext` — **nulled** and `tombstoned_at` set; `request_digest`, `account_id`, `created_at`, `refresh_expires_at`, `tombstoned_at` — **retained permanently by trigger**. | Idempotent bootstrap/credential handoff → replay lookup and bounded ciphertext cleanup. This is **not** an erased account-linked row: `account_id` survives without an FK. D-009/D-015 must explicitly rule the UUID/digest/time tombstone or change its contract in a new migration; no agent may call secret destruction complete data erasure. |
| `save_streams` | `id`, `owner_kind`, `owner_id`, `scope`, `created_at` — **retained**; `archived_at` set for the departed Founder's streams. | Versioned save owner → load/replay/idempotency. Founder owner UUID remains joinable to `account_founders.founder_id`; guild/world streams are shared, not automatically one player's export. No archived-stream expiry established. |
| `save_revisions` | `stream_id`, `revision`, `version`, `state`, `constants_hash`, `created_at` — **retained subject to normal revision pruning**, not deleted as an account right. | Save writes → load/replay/genesis. Ordinary writes prune older revisions beyond the latest five unless protected as Founder genesis. `state` is JSON: field-by-field contents vary by save version and are **not** classified by this SQL-column table. D-008 must choose current-only versus historical revisions, and a separate JSON schema/payload audit is required before a completeness claim. |
| `intent_records` | `stream_id`, `intent_id`, `request_hash`, `outcome`, `receipt`, `created_at` — **retained**. | Idempotent intent writes → replayed receipt lookup. `PruneIntentRecords` exists and has an integration test, but no non-test production caller; **30 days is not a current bound**. `receipt` JSON needs a payload audit for export/deletion disclosure. |

All ten have no mounted player export action or route at this source. “Candidate export
metadata” above identifies what D-008 must decide; it is not a recommendation to export
security hashes, hidden state or another player's shared records.

## Executed and source controls

- Cold `make test-save-integration SAVE_TEST_PACKAGES='./account'
  SAVE_TEST_FLAGS='-run TestAccountSessionIntegration -count=1 -v'` passed on 2026-09-23
  against the declared Postgres service. The test checks zero account/email/session/access/
  family rows, zero **live bootstrap secrets**, retained archived save streams and five retained
  Founder rows with zero linked `account_id`. This is the positive deletion control; it does
  not assert zero bootstrap rows.
- The earlier temporary diagnostic at the same product source reported one retained
  `bootstrap_receipts` row still carrying the deleted account UUID. Migration `00073` forbids
  DELETE and account-ID change; `DeleteAccount` and the composed expiry pruner only null secret
  bytes. That is the first negative control and remains RP-113.
- `server/save/intent.go` defines `PruneIntentRecords`, but a production-caller search still
  finds only test use. The second negative control prevents a false 30-day retention claim.
  By contrast, `server/gameserver/composition.go` schedules credential cleanup every minute.
- Current JSON payloads, the other 50 tables and browser/backups/logs are **not** classified by
  this field map. The parent [`data-rights-inventory.md`](data-rights-inventory.md) retains
  their complete relation/non-DB denominator.

## Owner/legal handoff

1. **D-008:** choose whether the export covers active account metadata, current Founder/save
   state, archived Founders/versions/intents, and any security lifecycle metadata. Do not equate
   database dump with a safe portable import format. Classify `state`/`receipt` JSON and the
   remaining 50 tables before adopting a completeness claim.
2. **D-009:** author the exact deletion disclosure: live credentials/account/email removed;
   Founder/save/intent history retained; bootstrap UUID/digest/time tombstone permanent under
   current schema; browser keys/backups and shared records separate. Do not claim that the
   absence of direct `account_founders.account_id` makes every surviving record unjoinable.
3. **D-015:** choose purposes, triggers, durations and cleanup owners for retained account/save
   fields and later payloads. In particular, decide whether immutable bootstrap tombstones and
   unpruned intent receipts are acceptable; if not, require a new accepted contract/migration
   that preserves idempotency and replay invariants. Never edit applied migration `00073`.

R-003 remains downstream of those choices and a built player-facing prototype. This tranche
does not make account rights complete or change the preview's release status.
