# D-008/D-009/D-015 — selected receipt and response payloads

Research checkpoint: 2026-09-23. Product source `7e8aa70`, with no later
`server/` diff. Method and controls:
[`data-rights-receipts-plan.md`](data-rights-receipts-plan.md). This is a
bounded source-and-fixture classification, not a complete receipt union,
player export schema, deletion policy or legal conclusion.

## Four stores are not one receipt schema

| Store / producer → consumer | Current fields and copies | Account deletion and retention boundary |
|---|---|---|
| `intent_records.receipt` — Save `ApplyIntent` and cross-stream Minigame/Soul coordinators → idempotent retry. | SQL requires a JSON object, and `validateIntentDecision` checks object syntax/outcome but does not impose one universal receipt schema. Ordinary applied Company intents include `intent_id`, outcome, applied count, evaluated time, resource changes, new revision and a full `snapshot`; rejected decisions have current revision and category/detail. Founder care, exit, Minigame and Soul have different typed bodies. Founder/Company receipts are also copied to player outbox; logged commands copy receipts to run/Founder logs and later archive. A request hash is only an identity digest, not the JSON. | The account deletion path archives Founder streams, so this stream-FK row remains. `PruneIntentRecords` exists but has no production caller. A future export/deletion rule must address copies even if intent records are pruned. No joined populated-rights witness exists. |
| `minigame_create_receipts.response` and `minigame_command_receipts.response` — production Minigame API coordinator → exact idempotent response replay. | `00071` requires JSON objects and makes child receipt rows immutable. The production create response has session ID, Minigame/engine identity, mode, constants hash, revision, status and full `snapshot`. Nonterminal command response uses that session shape. Terminal command response adds `resolution_receipt` containing session/result hash, Company/Founder revisions, credited resource/delta, cap forfeit, rating change and nested quality change. `minigame_sessions.resolution_receipt` separately stores the inner resolution object; `intent_records`/run and Founder logs can hold related but not always byte-identical API-vs-kernel receipts. | Founder/session parents are archived/retained, not deleted, by account deletion (RP-122). Direct child update/delete is rejected. A parent-session delete with both child receipt kinds also fails: `00071`'s unconditional `BEFORE DELETE` trigger fires during FK cascade (RP-128). Thus a proposed parent-expiry job cannot work under the current migration without an accepted repair. Legacy resolved sessions can have a null resolution tuple (`00060`). |
| `guild_intent_records.receipt` — Guild `finish` → idempotent retry. | Applied body has `intent_id`, `outcome`, `new_revision`, `guild_id`; rejected body has `intent_id`, `outcome`, `current_revision`, and nested category/detail. SQL requires an object. This per-account receipt is separate from retained shared `guild_events.payload`. | `guild_intent_records.account_id` references `accounts` with `ON DELETE CASCADE`, so deleting the account removes the per-account receipt row. A cold Guild fixture left seven live-account receipt rows and zero orphan rows after its deletion paths; that fixture does not assert a pre/post receipt count for the deleted account. Guild shared history can still retain account IDs in JSON (RP-120). |
| `soul_recovery_sessions.terminal_receipt` and `progress_token` — Soul start/progress/finish → recovery and idempotent terminal retry. | Start API response includes a UUID `progress_token`; Progress requires that value with an authenticated active Founder and compares it in constant time. A repeated start rotates it for an active session. Terminal receipt is a distinct JSON object with session/intent/activity, action, revisions, Soul before/after and bands, and optional `cancelled_by`; it does **not** include the progress token. The session row stores the token separately as non-null UUID. The terminal receipt is also written into `intent_records` and the Company player outbox; the Company/Founder replay logs store their separate transition receipts. | `FinishTx` writes terminal status/receipt but does not clear or rotate `progress_token`; `00070` requires it non-null, and terminal rows are immutable. In the cold Soul fixture all 14 terminal rows (10 cancelled, four resolved) still held a token and terminal receipt; all 14 matched `intent_records.receipt` JSON by session ID. The token is no longer sufficient to progress a terminal session, and the API also requires authentication; this is retained historical capability material, **not** proof of an active post-delete authorization bypass (RP-127). Account deletion retains the Founder/session path (RP-122). |

## Executed evidence and limits

Cold real-Postgres `-count=1 -v` runs passed for:

- `./minigame` `TestAPIReceiptAndCurrentSessionIntegration` — idempotent
  create/command receipt storage and immutable-child checks. Its fixture
  response bodies are intentionally small (`{kind,session_id}` and
  `{revision,session_id}`), so they do not prove the production API shape.
- `./production` `TestStartMinigameAPISessionAtomicSequenceIdempotencyAndReplayIntegration`
  and `TestResolveMinigameSessionIntegrationAtomicReplayAndFaults` — production
  create/terminal coordinator paths. Read-only SQL observed one create response
  with all nine production top-level fields, one API command response with a
  nested resolution receipt, and one core Minigame resolution receipt. In the
  matching fixture rows, the API command's nested resolution equaled its
  session receipt (1/1), and the separate core receipt equaled its session
  receipt (1/1). These are two session populations, not one cross-population
  equality claim.
- `./production` `TestSoulRecoveryIntegrationAtomicSuppressionReplayAndExclusivity`
  — 14 terminal session rows held non-null progress tokens and terminal
  receipts; all 14 joined core intent receipts equaled the terminal JSON.
  The receipt key sets exclude `progress_token`; the SQL column holds it.
- `./guild` `TestGuildLifecycleConcurrencyAndHistoryIntegration` on a fresh
  temporary database — seven Guild receipt rows remained for live accounts,
  zero orphans, with the applied/rejected keys above. The FK cascade is the
  deletion authority; this observation alone is not a counted deletion proof.

In a fresh Minigame fixture, an exact session had one create and one command
receipt. A transaction-scoped `DELETE FROM minigame_sessions WHERE session_id=
'018f0000-0000-7000-8000-000000000104'` exited 1 with `minigame API receipts
are immutable`, naming the create-receipt cascade in the error context. The
connection rolled back and a read-only count confirmed the session still
existed. This is an executable negative for the currently blocked parent
cleanup path, not authorization to remove the immutability guard.

None of the Minigame or Soul fixtures deletes an account *after* populating
these exact records. The Guild fixture includes account deletion but does not
assert the per-deleted-account receipt count. None proves a player export,
disclosure, retention duration or clean-host restore behavior. Each temporary
Postgres service/network was removed after use; the named Go cache volume was
retained.

## Decision handoff

- **D-008:** decide whether receipts, API snapshots, nested Minigame result
  facts, Soul session status and retained history copies enter player export,
  and how cross-party/session data is filtered without breaking replay.
- **D-009/D-015:** decide disclosure, purpose, access, cleanup trigger and
  duration for retained core/Minigame/Soul receipts and the terminal Soul
  progress token. RP-128 requires an implementation contract if Minigame
  parent-session expiry is chosen. Do not call Guild per-account receipt
  cascade a deletion of its shared event history.
- Remaining receipt variants, historical encoders, Guild/Minigame/Soul
  non-receipt JSON, verification dead letters and joined populated deletion/
  restore proof remain open. No implementation or release promotion follows
  from this map.
