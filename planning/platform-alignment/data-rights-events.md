# D-008/D-009/D-015 — retained core event payloads

Original research checkpoint: 2026-09-23. Product source `7e8aa70` (no later `server/`
diff at this checkpoint). Method and controls:
[`data-rights-event-plan.md`](data-rights-event-plan.md). This classifies the
core `events` table's registry and selected high-linkage payload families. It
does **not** classify all event keys, `guild_events`, receipts or every historic
encoder, and is not an adopted export/deletion policy.

## Current source/native refresh — 2026-10-14 at 089849ad

The original 48-kind census below is historical, not today's complete registry.
Ordered Up migrations through `00084_cosmetic_event_payload_keys.sql` leave the same
60 game tables (61 creates, one drop). Their exact set still matches the parent
inventory's eleven groups. Migrations 00075–00084 add no tables or SQL columns;
they do add event kinds, version allowances and history constraints. The old
60/60 SQL-column map therefore does **not** cover the new nested payload variants.

The latest kind constraint, 00082, contains **60 distinct kinds**, exactly matching
the executed Go `AllEventKinds` population. The latest version constraint, 00077,
permits **65 SQL-source kind/version pairs**: `run_ended` v1/v2/v3;
`run_started`, `opportunity_claimed.v1` and `buff_started.v1` each v1/v2;
the other 56 kinds v1. Executing `validEventSchemaVersion` across the registered
population gives **64 admissible write-version pairs**, excluding historical
`run_ended` v1. This is version admission, not payload validity or proof of a
current producer for every pair. The version predicate alone is not the unknown-kind
guard. No current installed Postgres constraint was inspected in this refresh.

### Twelve added kinds and their retained links

All twelve added kinds use schema version 1. Payloads without an embedded account
UUID still belong to identified event streams; the absence of that key is not
an anonymization result.

| Added family | Actual payload / producer | Retained linkage and research limit |
|---|---|---|
| `reputation_node_purchased.v1` | `production/reputation_intent.go` emits `node_id`, `node_kind`, `cost`, `reputation_level`, `reputation_spent_before`, `reputation_spent_after`, `unlock_ppm_after`, `source`; both direct purchases and Exit-plan purchases use it. | Founder stream and event intent/revision identify purchase history. `source` distinguishes direct versus Exit-plan execution. Native admission/producer checks do not establish deletion or an export policy. |
| `pet_adopted.v1` | `production/pet_adoption_intent.go` emits `pet_id`, `species_id`, `temperament`, `palette_id`, `name_key`. Saved identity and the receipt additionally carry adoption times. | Founder stream plus pet ID link adoption, care and equipment history. This event map does not classify all saved identity/receipt fields or infer that `name_key` is arbitrary player text. |
| `cosmetic_acquired.v1`, `cosmetic_equipped.v1`, `cosmetic_unequipped.v1` | `production/cosmetic_intent.go`: acquire carries `cosmetic_id`, `order_number`; equip carries `cosmetic_id`, `pet_id`, nullable `replaced_cosmetic_id`; unequip carries `cosmetic_id`, `pet_id`. | Founder event ownership and pet ID link purchase/equipment history. Strict key tests protect the current grammar, not historical removal or retention duration. |
| `achievement_reattained.v1` | `production/axis_stack.go` emits `run_id: {company_stream_id, run_seq}`, `achievement_id`, `score_grant` for already-earned run achievements attained again. | Exact Company/run coordinate survives inside JSON, distinct from Founder lifetime earning. The old earned-event map is not a substitute for this run-local variant. |
| `garden_advanced.v1`, `garden_planted.v1`, `garden_uprooted.v1`, `garden_substrate_set.v1`, `garden_harvested.v1`, `garden_harvest_credited.v1` | `save/garden_events.go` defines exact key sets. `production/garden_intent.go` emits advance/plant/uproot/substrate; `production/garden_harvest.go` emits Founder harvest and Company credit. Fields include grid/species/recipe facts, attended tick/catchup facts, substrate changes, harvest plots/units/discovered seeds/hash, and credited/forfeited/faucet/cap facts. | Founder and Company histories share `harvest_hash`; the coordinator also persists source stream/log coordinates and receipts/replay inputs. The six event bodies are not a complete Garden state, faucet, command or receipt export map. |

### New variants inside existing history stores

- `run_started` v2 retains the same Founder/run/time/assistance links as v1 and
  adds `reputation_tree`: null, or `bonus_factor` plus `applied_starter_node_ids`.
  `save/intent.go` owns its grammar; `production/prestige.go` selects v2 when
  a tree carry-over summary is present. This is a new pair, not a new event kind.
- Migration 00078 admits Company-linked Founder `exit.v2`; the producer's
  `reputation_purchases` retain node IDs and resolved costs in the existing
  `founder_log.replay_inputs` arm. Migration 00076 requires the extra frozen
  Reputation contribution for a tree-pinned run in existing
  `run_frozen_contributions` rows. Neither adds a SQL column.
- Migration 00083 admits the Company-linked Founder `garden_harvest_credited`
  resolved arm. The coordinator's Company `credit_garden_harvest` inputs carry
  a Founder log coordinate and harvest/policy/faucet evidence. The source
  coordinate is separate from the shorter core event payload.

The current `account/Repository.DeleteAccount` still archives Founder mappings
and Founder-owned streams, tombstones bootstrap secrets and deletes the account;
it does not rewrite core event JSON. Migration 00042's Founder-owned event trigger
copies the complete `NEW.payload` into `transport_player_outbox`. These are
current source traces, **not** fresh joined post-delete observations for the
twelve added kinds. The old 28/28 Soul fixture below remains dated and Soul-only.

### Executed refresh and limitations

- An inline read-only Node census folds all 84 ordered Up bodies, refusing duplicate
  creates/unknown drops and unclassified table renames/views. It confirms the
  exact 60-name inventory relation set, no post-00074 table/column additions,
  the latest kind/version constraint endpoints and 65 source pairs. Its first
  version-grammar extraction refused with `unknown fallback` because it expected
  spaces around `=`; the corrected extraction matches the actual source and
  checks the exceptional-kind set. The failed attempt yielded no accepted census.
- Temporary `TestRightsEventCensusObservation`, run through
  `make test-go GO_PACKAGES='./save' GO_TEST_FLAGS='-count=1 -run TestRightsEventCensusObservation -v'`,
  passes: registry 60, native write-version pairs 64, exact sorted SQL-source/Go
  kind equality. It checks version values −1 through 4; the actual predicate was
  also inspected. The temporary test is removed after measurement; no production
  or permanent test bytes change in this research checkpoint.
- Existing cold `./save` checks pass using root Make and separate `-run` prefixes:
  `TestReputationPurchase`, `TestCosmeticEvent`, `TestGarden`, `TestValidate`,
  `TestFounderReputationMigrationCorpus`, `TestCompanyAxisMigrationCorpus`.
  Reputation/Garden producer controls run actual pure producers and their validator;
  codec migration corpora are **not** live DDL migrations. This selected population
  does not validate every current/historic payload or prove account deletion.
- The declared Docker Unix socket's read-only `/_ping` returned exit 28 after
  5.002526s, zero bytes/HTTP 000. Docker is unresponsive, not proved stopped.
  A repeat after checking the configured context also timed out: exit 28,
  5.003793s, zero bytes/HTTP 000.
  No restart, cleanup, new Postgres migration, joined deletion or restore run
  was performed. Earlier database results are retained with their original dates.

RP-602 records this evidence-coverage correction. D-008/D-009/D-015 still need
explicit policy and fuller nested-state/receipt/history classification plus
joined populated deletion/restore proof. No legal conclusion, retention duration,
schema change or feature/release approval follows.

## Historical 2026-09-23 denominator, not one uniform payload

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
