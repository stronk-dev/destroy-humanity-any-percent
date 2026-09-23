# D-008/D-009/D-015 — retained core event payloads

Research checkpoint: 2026-09-23. Product source `7e8aa70` (no later `server/`
diff at this checkpoint). Method and controls:
[`data-rights-event-plan.md`](data-rights-event-plan.md). This classifies the
core `events` table's registry and selected high-linkage payload families. It
does **not** classify all event keys, `guild_events`, receipts or every historic
encoder, and is not an adopted export/deletion policy.

## Closed denominator, but not one uniform payload

The latest `events_kind_check` Up constraint (migration `00072`) and the 48
`EventKind` literal values in `save/intent.go` have identical sorted sets; the
comparison has no diff. The live Postgres `pg_get_constraintdef` agrees. The
latest schema-version constraint (`00074`) permits 52 kind/version pairs:
`run_ended` v1/v2/v3, `opportunity_claimed.v1` and `buff_started.v1` each
at SQL schema v1/v2, and the other 45 kinds at v1. The Go *write* validator
admits only `run_ended` v2/v3, so it admits 51 pairs. SQL's `run_ended` v1
is a historical persistence allowance, not a pair accepted by the current Go
write validator. The 51 count is validator admissibility, not proof that
production has a live producer for every pair.

`validateEventPayload` switches on all 48 kinds, with strict JSON decoding
for the selected families. Strict syntax and kind-specific checks prevent
arbitrary extra keys at current writes; they do not make a retained UUID
anonymous, define a player export, or prove every historic payload's meaning.
The sorted 48-name validator reference set matches the 48-name literal set
after filtering helper type names, so no current kind was silently omitted
from this registry control.
There are unions within a version (for example opportunity-claim and fiscal
harvest), and `run_ended` v3 adds branch/starter content. Treating the 48
kinds as 48 identical JSON objects would lose the actual denominator.

## Selected identity-bearing families

| Core event family | Nested linkage and current producer | Consumer / post-delete boundary |
|---|---|---|
| Run/Founder progression (`run_ended`, `run_started`, `founder_advanced`, Gate/Route/Compact/Doctrine/Guild-accrual events) | `founder_id` and/or `run_id: {company_stream_id, run_seq}` appear in the validator and production/Guild writers. `run_ended` also carries tier, timed run, route/gate, payout, assisted and v3 branch/starter facts. These are exact Founder/run links even if the account FK is later nulled. | Save replay, projections, verification and matching run archive entries consume the history; the event trigger copies Founder-owned event payload into `transport_player_outbox`. Account deletion archives rather than deletes Founder streams and does not clear these event bodies. Guild-accrual core events are distinct from `guild_events`, whose separate account-ID retention is RP-120. |
| Minigame (`minigame_resolved.v1`, `minigame_rating_changed.v1`) | Both include `session_id`, `minigame_id` and certified-result hash. Resolution includes payout and Founder revision; rating includes old/new Elo, season member and nested quality states. Production writes one Company and one Founder event on resolution. Session linkage to other participants and export redaction is not settled by this core-event schema. | The minigame integration path persists/replays them; the shared trigger produces event outbox copies. Account deletion retains the referenced Founder/Company streams and does not edit event JSON. |
| Soul recovery (`soul_recovery_started.v1`, `soul_recovery_cancelled.v1`, `soul_recovered.v1`) | Each carries `session_id`, `activity_id`, `company_stream_id` and `run_seq`; terminal variants add attended-time and soul values, recovered additionally adds amount/bands/reason. The Founder replay producer writes exact-key JSON checked by `validateEventPayload`. | Source events feed replay and the outbox. The session and Company-stream IDs remain joinable after account unlinking. A read-only check of the cold Soul integration fixture found 28 core recovery event rows, and all 28 matching outbox rows held JSON-equal nested `payload` and the same `kind`. This is a producer→outbox observation, **not** an after-deletion result. |

For Founder-owned Company/Founder streams, migration `00042`'s `AFTER INSERT`
trigger creates an outbox JSON wrapper with the full source `payload`, event
ID/kind, scope, revision and cursor effect. `DeleteAccount` archives
`account_founders` and `save_streams`, tombstones bootstrap secrets, then
removes `accounts`; it neither rewrites these `events.payload` objects nor
deletes the retained stream. Thus a detached account link is not an event
payload anonymization step. The separate core-payload dossier tracks archive
copies and outbox retention.

## Executed checks and what they cannot claim

- Cold `make test-go GO_PACKAGES='./save'` with the event-validator filter and
  `-count=1 -v`: ten selected registry/typed-payload tests passed. This
  confirms the current validator's tested examples and negatives, not all 52
  persisted historic pair schemas.
- Cold real-Postgres `TestAccountSessionIntegration`,
  `TestSoulRecoveryIntegrationAtomicSuppressionReplayAndExclusivity`, and
  `TestResolveMinigameSessionIntegrationAtomicReplayAndFaults` each passed
  separately with `-count=1 -v`. The account test deletes an account and checks
  archived streams, but does not join that deletion to populated Soul/Minigame
  event payloads. The latter tests populate those event families but do not
  delete the account. No joined deletion claim follows from three green tests.
- An initial attempted combined Make filter used unescaped shell `|` tokens;
  the shell treated them as pipelines and that invocation exited 2. It is not
  a test result. The three separate, successful commands above are the
  admissible runs.
- The read-only fixture query above also observed the selected Soul payload
  key sets and 28/28 event→outbox equality. Its denominator is only the
  records left by that particular integration test, not all kinds or production
  data. The ephemeral Postgres container/network was removed; the named Go
  cache volume was retained.

No new ledger row duplicates RP-118 (joined populated-history deletion),
RP-122 (retained Minigame/Soul/Transport family), or RP-124 (archive fixture
with empty payloads/no matched events). Those existing findings remain open.

## Decision handoff and remaining research

- D-008 needs a per-kind/version export rule and explicit treatment of shared
  Minigame/session fields, replay dependencies and archive copies. A current
  state-only export plainly omits these retained records, but whether that is
  acceptable is the owner's decision with legal review.
- D-009/D-015 need a disclosure, purpose, cleanup trigger/duration and
  preservation exception for nested Founder/run/session links and their
  event→outbox/archive copies. Account unlinking alone is insufficient to
  describe the remaining bytes as erased.
- Selected core, Minigame, Guild and Soul receipt stores now have a bounded
  follow-up map in [`data-rights-receipts.md`](data-rights-receipts.md). The
  full receipt union, every core event's exact keys/historic encoders,
  separate non-receipt Guild/Minigame/Soul JSON, verification dead letters,
  and joined populated deletion/restore witnesses remain open. No rights
  implementation or release status is authorized by this research.
