# D-008/D-009/D-015 — catalog, Routes and Commons field tranche

Research checkpoint: 2026-09-23. Product source `7e8aa70`; no later product-path change at
execution. Method and controls: [`data-rights-fields-shared-plan.md`](data-rights-fields-shared-plan.md).
This maps every current SQL column in 16 shared/Founder-linked tables. The three field
tranches together cover **40 of 60** game tables. Guild (13), Minigame (5), Soul (1), Transport
(1), nested payloads and non-DB stores remain outside this tranche. This is not an adopted
export/deletion/retention rule.

## Exact current-schema field map

Original Up DDL is in migrations `00003`, `00005`–`00007`, `00013`; later Up changes in
`00008`, `00009`, `00021`, `00024`, `00043`, `00045` affect current fields/constraints.
`account.DeleteAccount` archives Founder save streams and deletes the account row; it does not
delete any of these sixteen tables. That conclusion comes from DDL and the deletion source,
not a joined account-deletion test populated with all these families.

| Table | Exact current SQL fields and deletion disposition | Producer → reader / unresolved rights issue |
|---|---|---|
| `catalog_sets` | `constants_hash`, `created_at` — retained; update/delete rejected by trigger. | Epoch/catalog publisher → replay identity. Shared bundle identity, not one account's row. Export as portable import dependency remains D-008. |
| `catalog_artifacts` | `constants_hash`, `artifact_name`, `bytes`, `created_at` — retained; update/delete rejected. | Publisher → pinned runtime/replay. `bytes` are shared authored balance/content artifact bytes, not automatically player data; historical availability is a preservation question. |
| `epochs` | `epoch_id`, `name`, `started_at`, `ended_at`, `changelog_ref` — retained. Only the current row's `ended_at` may be set once; other update/delete rejected. | Epoch mint/reconcile → pinned run interpretation. Shared governance record. |
| `epoch_hashes` | `epoch_id`, `constants_hash`, `accepted_at` — retained; update/delete rejected. | Epoch mint/hotfix → accepted replay hashes. Shared historical authority. |
| `route_projection_events` | `event_id`, `route_id`, `founder_id`, `company_stream_id`, `run_seq`, `occurred_at` — retained. | Route-event projector → idempotency/route history. Founder and stream IDs remain linkable after account unlinking; no adopted cleanup/export. |
| `founder_route_executions` | `founder_id`, `route_id`, `first_event_id`, `first_occurred_at`, `last_occurred_at`, `execution_count` — retained. | Projector → Founder distinct-route reader and registry history. First/last timestamps and count are player-history candidates under D-008/D-009. |
| `founder_route_state` | `founder_id`, `route_knowledge_balance`, `route_knowledge_debt` — retained. | Route/hint/knowledge projector → authoritative balance/debt reader. `route_knowledge_debt` was added in `00008`; an export of only positive balance would omit current state. |
| `route_hint_projection_events` | `event_id`, `founder_id`, `route_id`, `cost` — retained. | Hint projector → idempotency and debited Route Knowledge. Founder-linked purchase history; no adopted expiry/export. |
| `registry_routes` | `route_id`, `first_event_id`, `first_founder_id`, `occurred_at`, `house_name`, `name`, `name_state`, `naming_reserved_until`, `execution_count` — retained. | Global first-executor registry → route/record lookup. `first_founder_id` is a retained player link in a shared row. Current producer seeds `name` from `house_name`; package methods can accept a proposed name and resolve it, but no non-test production caller or mounted player naming route was found. `ExpireNames` also lacks a production caller (RP-104). Do not claim live public UGC or bounded naming expiry. |
| `commons_cohorts` | `cohort_id`, `server_id`, `activity_bracket`, `cohort_seq`, `member_count`, `standing_numerator`, `standing_denominator`, `closed_at`, `created_at` — retained. | Commons projector → cohort placement/health/merge. Shared aggregate; deleting one account does not imply deleting the cohort or resetting historic assignment count. No adopted expiry. |
| `founder_commons_assignments` | `founder_id`, `server_id`, `activity_bracket`, `cohort_id`, `first_signed_at`, `last_signed_at` — retained. | Commons signer → stable Founder/cohort mapping. Pseudonymous per-Founder times and membership link remain after account unlinking. |
| `company_compact_memberships` | `company_stream_id`, `founder_id`, `run_seq`, `cohort_id`, `member`, `tithe_ppm`, `updated_at`, `projected_revision` — retained. | Signed/left compact projector → current company membership and replay-safe revision. Account deletion does not synthesize a leave event or clear `member`; the health reader currently uses `member=true` without checking account or stream archival (RP-119). The desired transition is a product/retention decision, not invented here. |
| `commons_projection_events` | `event_id`, `kind`, `founder_id`, `company_stream_id`, `run_seq`, `occurred_at` — retained. | Compact event projector → idempotent sign/leave/sample/tithe history. `kind` accepted values were widened through `00024`; event-row retention is not a player-surface proof. |
| `commons_member_samples` | `company_stream_id`, `founder_id`, `cohort_id`, `weight_ppm`, `compliance_ppm`, `solidarity_ppm`, `enclosure`, `capacity`, `sampled_ms`, `updated_at`, `run_seq` — retained. | Sampling projector → Commons health/board variables. `run_seq` is **nullable** after `00045`: stale backfill labels were invalidated rather than fabricated. This is a per-company sample with an aggregate consumer, not purely anonymous metric data. |
| `commons_health_scopes` | `scope_kind`, `scope_id`, `raw_health_ppm`, `health_ppm`, `capacity`, `real_members`, `npc_weight_ppm`, `evaluated_at` — retained. | Health projector → cohort/server aggregate reader. `scope_id` can refer to a cohort; it is shared output, though underlying samples remain Founder-linked. No adopted expiry. |
| `commons_recruitment_offers` | `founder_id`, `event_id`, `offered_at` — retained. | Recruitment event producer → Founder offer/idempotency record. Founder UUID remains; no adopted expiry/export. |

The negative control holds: `catalog_sets`/`catalog_artifacts`/`epochs`/`epoch_hashes` are
shared replay authority and cannot be deleted as one player's account cleanup. The positive
control also holds: Founder/stream IDs occur in Routes and Commons tables even though
`account_founders.account_id` becomes null after account deletion. Neither fact selects an
export format, retention duration or lawful deletion disclosure.

## Executed evidence and limits

Cold `make test-save-integration SAVE_TEST_PACKAGES='./routeprojection ./commonsprojection
./leaderboard' SAVE_TEST_FLAGS='-run Integration -count=1 -v'` passed: two Route projector,
one Commons projector, and four Leaderboard/epoch integration tests on real Postgres. These
exercise producer/idempotency/convergence, cohort assignment and epoch/board use, including
Commons membership variables. Their fixtures **do not delete an account with those rows
populated**; no joined rights outcome is claimed. The test container and its ephemeral network
were removed after execution, leaving the named Go cache volume intact.

The source search found no non-test production calls to `SubmitName`, `ResolveName` or
`ExpireNames`, nor any product delete/expiry job for the sixteen tables. Lack of a caller is
not proof a future interface cannot write a name; it bounds the current preview claim.
No row-level payload classification is inferred from an identifier or numeric field name.

**Source-derived Commons hazard (RP-119):** `account.DeleteAccount` archives the Founder's
`save_streams` and deletes the account but touches no Commons membership. A signed
`company_compact_memberships.member=true` row therefore stays true; `refreshScope` selects
`commons_member_samples` joined to those membership rows with `m.member=true`, without a
`save_streams.archived_at` or account check. A future recomputation can continue counting the
departed Founder's stored sample in cohort/server health. This is a concrete producer→reader
trace, **not yet** a cold joined account-delete→refresh witness or a ruled policy to subtract
historical contributions. The rights/Commons successor must decide active-vs-historical
semantics, implement the chosen transition without corrupting replay, and prove it on real
Postgres.

## Decision handoff

- **D-008:** choose if per-Founder Routes/Commons history and global first-executor credit
  belong in export, and how a portable self-host import carries shared epoch/catalog identity
  without pretending the player owns those rows. A current-state-only export must disclose
  omitted retained history.
- **D-009:** specify what account deletion does and says about Founder-linked projections,
  registry first-credit/name and shared cohorts. RP-119 requires a rule for active Commons
  health after account deletion; no cleanup or synthetic leave mechanic is authorized here.
- **D-015:** rule purpose, trigger, duration and cleanup owner for Route/Commons projections
  and shared aggregates, alongside immutable catalog preservation. Do not schedule generic
  table deletion across FK-linked replay authority.

The other twenty game tables, nested JSON/binary/text payloads, browser storage, backups,
logs and operator artifacts still require separate evidence before a complete rights claim.
