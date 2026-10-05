# Cloud Clicker 1.0 goal log

Append-only checkpoints for the strategic board in [`roadmap-1.0.md`](roadmap-1.0.md). Detailed
implementation history remains in each RFC's own plan/log; this record states only changes to the
long-term release picture.

## 2026-09-23 — goal established at `cf4ac25`

- **Product state:** development snapshot. The adopted next target is a bounded Phase-0 Playable
  Preview; the designed 1.0 remains Transcendence through Tier 8 and the three ending families.
- **Verified recent work:** Deployment Foundation is accepted and implementing. The local CI parity
  correction `e0e2201..cf4ac25` records a complete `make verify-push` pass and a Codex first-filter
  verdict in `planning/deployment-foundation/log.md`; no designated cross-party verdict is recorded
  for that range. This checkpoint did not rerun the gate or claim a hosted run.
- **Known limits:** the platform-alignment capability audit predates the Deployment implementation
  and is used as a bounded baseline. DP-F's exact clean-host rehearsal is unfinished; account
  rights/retention, task accessibility and the later gameplay phases remain open. The old
  executable-queue handoff described Deployment as a draft and was corrected in this planning
  checkpoint.
- **Next:** designated review of the correction range; DP-F completion and R-006; account-rights
  rulings and accessibility contract. The accepted per-RFC plan determines implementable tasks.
- **Provenance:** Codex established this planning rollup from `design/07-roadmap.md`, the accepted
  owner packet, active RFC index, Deployment plan/log and audited capability map. This is a
  planning first filter, not a designated implementation review or release verdict.

## 2026-09-23 — DP-F2 browser construction at local `44f5b3f`

- **Observed change:** the exact-manifest rehearsal browser driver now attempts the default
  bootstrap → gate → first Exit → next-run journey using enabled player controls, with strict
  version-2 evidence and a two-hour fail-closed execution guard. The prior single-click driver
  could not supply the populated recovery identity's verified-board domain.
- **Executed evidence:** cold focused browser/rehearsal tests, the declared broader rehearsal
  target and vet passed. Invalid-result mutations are covered by negative tests. The first broad
  target attempt was invalidated by sandbox Go module-cache denial; the unwrapped rerun passed.
- **Limits:** no real product browser run, DOM-action severing, asynchronous board observation,
  exact rebuilt bundle, clean-host R-006 or designated cross-party review has happened. DP-F's
  25/43 executed-probe count is unchanged. The branch is local; no push or release action occurred.
- **Next:** bounded board-arrival producer and its absence failure, followed by actual product
  execution and exact bundle rebuild. The per-RFC plan/log remains the implementation authority.
- **Provenance:** Codex first-filtered the construction range in the Deployment log; it did not
  substitute for the designated reviewer or mark AC1/AC4 complete.

## 2026-09-23 — verified-board recovery construction at local `6d89880`

- **Observed change:** the recovery identity now distinguishes a real `verified_runs` row from
  projection-event-only history. The rehearsal waits with a bounded guard for that asynchronous
  row, then checks identity again after backup before any destructive reset. Identity and private
  checkpoint schemas advanced to version 2.
- **Executed evidence:** cold focused/aggregate tests and vet passed. The declared backup lane
  passed against real Postgres 16 for empty/populated backup/restore and for an event-only board
  that the populated gate rejects. A first negative-fixture attempt tried to delete immutable
  board history and was rejected by the database; the corrected fixture never inserted the run.
- **Limits:** no actual fresh-player browser run, clean-host projection observation, exact rebuilt
  bundle, R-006 dossier or designated cross-party verdict. The 25/43 rehearsal probe count remains
  unchanged; no release status or plan checkbox moved.
- **Next:** execute the full player journey and populated recovery on the exact supported host,
  with named severing failures, then reconcile the release evidence. The Deployment per-RFC plan
  continues to own the detailed order.

## 2026-09-23 — corrected local release candidate at source `8621f33`

- **Observed change:** the validator now admits the exact retained pre-browser rollback schema
  only with a manifest lacking the whole browser closure. A browser-bearing bundle without its
  schema declaration and a historical bundle missing another required field still fail. The
  earlier candidate-v3 is diagnostic only because it embeds the old validator.
- **Executed evidence:** cold releasepackage/rehearsal/release tests, vet and the exact
  candidate-v3/previous probe passed after correction. Two independently rebuilt
  `0.1.0-preview.2` trees then matched across six static binaries, client, staged content,
  gameserver archive, seven normalized real image SBOMs and full 72-artifact bundles. Manifest
  SHA-256 is `521e5bca…334d36`; both candidate/previous probes, retained build-record validation,
  structured tracked-source/image secret scan and full local supply-chain check passed.
- **Limits:** raw Syft headers differ as expected; normalized release SBOMs match. The first
  Docker context and relative secret-scan output attempts failed and were corrected; RP-111
  tracks the latter tooling defect. No clean Linux host, actual player browser journey,
  populated restore, governed rollback, complete R-006, designated cross-party implementation
  verdict or release call has occurred. The 25/43 executed-probe count is unchanged.
- **Next:** designated review of the pending implementation ranges and exact-bundle clean-host
  execution/severing under the accepted Deployment plan. This is candidate construction, not
  supported self-hosting or 1.0 readiness.

## 2026-09-23 — rollback-version correction and final-source candidate `7e8aa70`

- **Observed change:** a future save-version bump no longer makes a byte-exact previous bundle
  fail manifest validation solely because it records the older company/founder versions. New
  bundles still record the current versions; zero and future versions refuse. The root secret-scan
  target now resolves a repository-relative result path and retains exclusive output creation.
- **Executed evidence:** the old equality-rule mutant failed the prior-version positive, and the
  corrected rule passed with zero/future negatives still red. The old relative scan invocation
  failed; the corrected one passed, then refused an overwrite without changing the output hash.
  Two fresh trees from `7e8aa70` match across binaries, client, image archive, seven normalized
  SBOMs and all 72 bundle artifacts. Both exact candidate/previous probes pass. The manifest hash
  is `677e94bb…47818e`; the retained record, zero-finding tracked-source/image scan and full local
  supply-chain result validate.
- **Limits:** source `8621f33` and its candidate-v4 pair are historical evidence, not the current
  candidate. None of the local passes prove a clean Linux host, player journey, populated restore,
  rollback, R-006 or designated cross-party verdict. The 25/43 probe count is unchanged. Codex
  did not push, archive, change release status or make an owner release call. During this checkpoint
  `origin/main` advanced externally to `7e8aa70`; hosted CI run `35865261957` was observed
  `in_progress`, not passed, at that exact source commit.
- **Next:** obtain the designated implementation-range review and an authorized clean Linux/amd64
  target for exact-bundle R-006 execution and severing. Continue the independent account-rights,
  accessibility and later-phase tracks under their own decisions and accepted RFCs.

## 2026-09-23 — hosted CI observed at candidate source `7e8aa70`

GitHub Actions [run 35865261957](https://github.com/stronk-dev/destroy-humanity-any-percent/actions/runs/35865261957)
completed `success` for exact head `7e8aa708ff9e002e7ec4b46242d7d889f76b7fdd`.
The six jobs — server, harness, schema, client, browser and game-ui-composed — each completed
successfully. This corrects the prior in-progress observation; it is hosted code CI, not a
clean-host deployment rehearsal, R-006 evidence or designated cross-party review. Codex did not
initiate the push or change the branch's publication state in this checkpoint.

## 2026-09-23 — data-rights inventory against product `7e8aa70`

- **Observed change:** a predeclared relation-complete census now names all 60 live game tables
  in 11 disjoint families, plus browser, backup, journal, metrics and operator outputs. D-008,
  D-009 and D-015 link to this evidence, but remain owner/legal decisions.
- **Executed evidence:** the current migrations applied on real Postgres and showed 60 game
  tables plus `goose_db_version`; migration Up create/drop accounting matched. The existing
  Account integration test passed cold with a temporary diagnostic that reported **one retained
  account-linked bootstrap tombstone after deletion**. The diagnostic was removed, leaving no
  product or test change. The dropped historical outbox and metadata table were exclusion
  controls; absence of a production intent-pruner caller and player export route was rechecked.
- **Limits:** this is a relation-complete family map, not a per-column legal classification,
  retention ruling, privacy notice, rights workflow or R-003/R-007 completion. The prior
  no-backup/no-metrics audit claims are dated; current Deployment mechanisms still need R-006.
  No release status, designated implementation review or clean-host result changed.
- **Next:** owner/legal rulings for export, deletion and retention; accepted implementation
  contracts and real player/operator rehearsals. In parallel, finish Deployment's exact-host
  R-006 and obtain its required cross-party review.

## 2026-09-23 — accessibility current-source negatives and draft contract

- **Observed change:** a predeclared check at product `7e8aa70` reconfirmed lifecycle focus loss
  and 320-pixel Desk overflow, then a draft cross-surface RFC was added. D-018 names the owner
  task/assistive-technology matrix; nothing was accepted for implementation.
- **Executed evidence:** temporary fixture-adjacent browser probes failed meaningfully in
  Chromium and WebKit: Offer focus became `BODY`, and the full Desk/document width was 647/320.
  A Firefox-only retry, like the multi-engine attempt, timed out before test execution. The
  probe file was restored byte-identically. Source trace shows CSS theme motion is sampled but
  the numeric shell receives the default non-reduced mode and has no live preference listener.
- **Limits:** no full keyboard, screen-reader, zoom, coarse-pointer, color-vision or real player
  task claim follows; R-005 is open. The draft RFC and D-018 need owner acceptance, task scope
  and manual test environment. No product change, release status or designated verdict moved.
- **Next:** rule D-018 and the exact release-task manifest, accept or revise the successor, then
  implement with failing controls and composed/manual task proof. The independent Deployment
  R-006/review path remains required.

## 2026-09-23 — three-engine accessibility negative and green ordinary control

- **Observed change:** the Mac Firefox startup gap no longer limits the current-source product
  finding. The declared Linux browser lane executed the same temporary probes in Chromium,
  Firefox and WebKit; focus and reflow failed identically in all three.
- **Executed evidence:** six deliberate failures named `BODY` focus and 647/320 Desk/document
  width, alongside 20,049 ordinary passes and three skips. After removing the probes, cold
  `make test-browser-ci` passed 123 files/20,049 tests (three skips), then its separate
  Chromium simulated-60-second budget test. The normal gate can be green while these two
  player-task properties fail.
- **Limits:** the Mac Firefox launcher remains unexplained; the declared Linux lane works.
  No accessibility implementation, real keyboard/screen-reader task, R-005 completion, owner
  D-018 ruling, designated review or release status follows from negative reproduction.
- **Next:** accept or revise the exact task contract and repair RP-082/RP-083/RP-084 under
  authority with permanent negative/positive witnesses; retain full manual and composed
  release-task proof.

## 2026-09-23 — D-007 exact-preview-manifest proposal prepared

- **Coordinate and scope:** product source `7e8aa70`; planning HEAD `e7d61f6` before this record;
  no product-path diff between them. Predeclared the Phase-0 source/evidence population and
  controls in `phase0-manifest-plan.md`. No code, authored copy or release authority changed.
- **Observed result:** `phase0-manifest-proposal.md` names seven candidate player tasks and five
  mandatory public-preview floor outcomes. Epoch 8 pins 19 artifact families, but only five
  Game UI surfaces mount; registered achievements and pets are the negative control against
  treating every pinned artifact as a shipped player task. Gate/Exit/next-run is the positive
  control, with designated-approved composed-browser evidence. The first-hour 97-run harness,
  branch proofs and composed server/Postgres seed are kept distinct from the narrower browser
  journey, which does not click every generator/upgrade or prove player comprehension.
- **Authority and limit:** D-001/D-007 adopted the bounded T0–T1 direction and public-hosting
  floor, not this exact P01–P07 manifest. D-007 still needs Marco's adoption/edit and a release
  RFC; D-018 and R-008 then consume the adopted task list. Static tracing is not a player study,
  clean-host R-006 or legal review. Deployment designated review and account-rights decisions
  remain independent blockers.
- **Next:** seek the exact D-007 boundary ruling; meanwhile continue authorized Deployment
  and relation/retention research without pretending this proposal authorizes implementation.

## 2026-09-23 — account/save fields classified, rights decisions still open

- **Coordinate:** product source `7e8aa70`, planning after `4492424`; no product-path change.
  The predeclared `data-rights-fields-plan.md` selects ten of the 60 current game tables and
  refuses to extrapolate to the other 50 or nested JSON.
- **Evidence:** every SQL column in those ten tables is mapped in
  `data-rights-fields-account-save.md`. The cold Postgres account-deletion integration witness
  passed with `-count=1`; it proves removal of live account/credentials and Founder unlink,
  not erasure of retained saves or permanent bootstrap UUID tombstones. Source search still
  finds no production intent-pruner caller. No new owner/legal conclusion follows.
- **Next:** D-008/D-009/D-015 choose export/deletion/retention semantics with legal review;
  continue the remaining 50-table/payload/non-DB classification before a complete rights
  claim. D-007 exact preview tasks, D-018, Deployment designated review and R-006 remain open.

## 2026-09-23 — history/board field tranche and cold archive/board witnesses

- **Coordinate:** product `7e8aa70`, planning after `624a2cb`; no product-path change.
  Predeclared fourteen history/verification tables and mapped their current SQL columns in
  `data-rights-fields-history-boards.md`, making 24/60 field-mapped game tables.
- **Evidence and correction:** archive-compaction and board-projection integration tests
  passed separately on real Postgres with `-count=1 -v`. A combined regex command had first
  failed because Make passed `|` to the shell; it was discarded as invalid evidence. RP-117
  corrects the parent inventory's blanket immutable wording for active `run_log` rows.
  RP-118 names the missing joined account-deletion-with-populated-board witness; source
  retention is not promoted to executed rights proof.
- **Limit/next:** 36 game tables, nested payloads and non-DB fields remain; owner/legal
  D-008/D-009/D-015 and D-007/D-018 remain open. No product/release status changed. Continue
  classification and require an accepted rights contract before a joined deletion witness.

## 2026-09-23 — shared data field tranche and Commons deletion risk

- **Coordinate:** product `7e8aa70`, planning after `345a3dc`; no product-path change.
  The predeclared catalog/Routes/Commons tranche maps 16 more SQL schemas, making 40/60
  game tables with field-level disposition. Twenty Guild/Minigame/Soul/Transport tables,
  nested payloads and non-DB stores remain.
- **Executed evidence:** seven Route/Commons/Leaderboard integration tests passed cold on
  real Postgres with `-count=1 -v`. They establish projector/epoch behavior, not deletion
  after those projections exist. The ephemeral DB container/network was removed.
- **New boundary:** RP-119 records the source-derived possibility that a deleted account's
  archived Founder remains `member=true` and therefore contributes a stored sample to
  later Commons health recomputation. No synthetic leave or cleanup policy was inferred;
  it requires D-009/D-015 semantics, accepted contract and joined witness.
- **Next:** finish the 20 remaining game tables and payload/non-DB classification, while
  independent Deployment review/R-006, D-007 exact tasks and D-018 accessibility remain open.

## 2026-09-23 — Guild field tranche, retained JSON UUID and reused-DB test failure

- **Coordinate:** product `7e8aa70`, planning after `0b4b53f`; no product/test-path
  change. The predeclared Guild tranche maps all 13 current Guild schemas in
  `data-rights-fields-guild.md`, making 53/60 game tables field-mapped.
- **Executed evidence:** a temporary diagnostic in the existing real-Postgres
  Guild deletion fixture found one retained `exchange_cleared` event whose
  `actor_account` was null but `payload.producer_account_id` was the deleted
  UUID. The diagnostic edit was removed byte-identically. The unmodified
  three-test Guild suite passed cold. It failed on a warm rerun (invalid
  intents/duplicate account IDs), passed after resetting the ephemeral
  Postgres database, then failed again on a second warm rerun. RP-120 records
  the data residual; RP-121 records test isolation. The ephemeral test
  container/network was removed, preserving the named Go cache volume.
- **Limit/next:** neither the Guild result nor FK rules classify all nested
  payloads or adopt deletion semantics. Seven game tables, nested payloads and
  non-DB stores remain, alongside D-008/D-009/D-015 and legal review. D-007,
  D-018 and the independent Deployment/R-006 path remain unchanged.

## 2026-09-23 — all game-table SQL columns mapped, payload decisions still open

- **Coordinate:** product `7e8aa70`, planning after `22967d1`; no product/test-path
  change. The predeclared final tranche maps five Minigame, one Soul and one
  Transport table in `data-rights-fields-final.md`, reaching **60/60** current
  game tables at SQL-column level.
- **Evidence:** selected Minigame, Soul and player-outbox tests passed on real
  Postgres with `-count=1 -v`. Source/FK trace shows account deletion retains
  Founder/streams, so those families' cascades do not fire on account removal
  (RP-122). No joined populated-family deletion outcome is claimed. The
  ephemeral Postgres service/network was removed; named Go cache retained.
- **Limit/next:** 60/60 is not privacy-policy completeness: nested JSON,
  command bytes, device-local credentials/timing, backups, logs, metrics and
  operator artifacts need field/disposition work. D-008/D-009/D-015 and legal
  review must rule export/deletion/retention before implementation. D-007,
  D-018, Deployment review and R-006 remain independent release gates.

## 2026-09-23 — browser persistence and recovery-code posture

- **Coordinate:** source `a6956de`, planning-only research. The predeclared
  `data-rights-browser-plan.md` led to a four-key current browser storage map
  in `data-rights-browser.md`: bootstrap retry key, credential document,
  per-Founder transport positions and per-Founder timing/splits. Production
  client search found no other client-managed storage API.
- **Evidence:** source traces exact writers/readers/clear paths; `make test-client`
  passed cold (39 files/6,662 tests passed, 2 files/22 tests skipped). Existing
  tests verify bootstrap credential storage and transport/timing mechanics,
  not a one-time recovery display or deletion of populated local state. RP-123
  records that the recovery code is silently kept in `localStorage` while the
  current Game UI lacks the owner-ruled display/copy/download consumer.
- **Limit/next:** D-005's recovery posture is settled but unimplemented;
  D-008/D-009/D-015 still choose export/deletion/retention details. Continue
  nested game-data payload and backup/log/metric/operator-field classification.
  No product/test byte, ruling, RFC status or release claim changed.

## 2026-09-23 — retained save/replay/transport payload lineage

- **Coordinate:** product `7e8aa70`, planning after `f09ba51`. The predeclared
  `data-rights-payload-core-plan.md` produced a bounded lineage dossier for
  versioned save state, event/intent payloads, command/replay logs, immutable
  gzip run archives and player transport rows.
- **Evidence:** cold real-Postgres archive-compaction and outbox integration
  tests passed with `-count=1 -v`; replay-input envelope validation passed in
  the non-DB Go lane. Source tracing confirms archive compaction copies full
  command, receipt, replay-input, genesis and matched-event bytes before
  removing unreferenced active rows, and transport publication retains payload.
  The ephemeral test Postgres service/network was removed; named cache kept.
- **Evidence limit:** the archive fixture uses `{}` command/replay/receipt
  payloads and expects no events. RP-124 records that its green result cannot
  prove nontrivial payload or matched-event preservation; those claims remain
  source-derived pending a populated, severable witness.
- **Limit/next:** these tests do not delete an account with all payloads
  populated. The dossier does not classify every event kind/version, receipt
  family, Founder log, Minigame/Soul/Guild JSON or operator artifact. D-008/
  D-009/D-015 and legal review remain open; no product/release status changed.

## 2026-09-23 — operator-held data and restore/deletion collision

- **Coordinate:** source after `dfdd6f4`, planning-only research. The
  predeclared `data-rights-operator-plan.md` produced a bounded map of full-DB
  encrypted backups, journald, private metrics/alerts, release/rotation
  ledgers and rehearsal/recovery identity artifacts.
- **Evidence:** `make test-go` for deploymentbackup/operations/deploymentrelease
  passed cold. `make test-deployment-backup` passed four real-Postgres 16
  populations, including populated encrypted backup→restore and non-clean
  refusal; its dedicated temporary Compose project was removed. These are
  local tests, not R-006/R-007 on the supported clean host.
- **New boundaries:** RP-125 records that newest/unresolved backup protection
  makes 30 days a non-hard maximum and that restoring a pre-deletion full dump
  can recreate a deleted account without a deletion replay. RP-126 records
  caller-supplied operator identifiers in append-only release/rotation ledgers
  without an adopted lifecycle. The restore/deletion collision is source-
  derived, not executed by the current backup test.
- **Next:** finish event/receipt and other nested-payload schemas, seek owner/
  legal D-008/D-009/D-015 rulings, then accepted rights/restore workflows and
  exact-host proof. No product, deployment or release status changed.

## 2026-09-23 — retained core-event identity checkpoint

- **Coordinate:** product `7e8aa70`, planning after `eeb2a3f`; predeclared
  `platform-alignment/data-rights-event-plan.md` and bounded result
  `platform-alignment/data-rights-events.md`. No later server diff was present.
- **Evidence:** SQL and Go have the same 48 event kinds. SQL permits 52
  kind/version pairs; the Go write validator admits 51, excluding historical
  `run_ended` v1. Ten selected validator tests and three separate real-Postgres Account,
  Soul and Minigame integration tests passed cold. A read-only fixture query
  matched all 28 Soul recovery events to outbox copies with equal nested JSON.
  The temporary Postgres service/network was removed; cache retained.
- **Limit/next:** none of the tests deletes an account after populating these
  Soul/Minigame event bodies. This is selected identity mapping, not every
  event key/version or an export schema. D-008/D-009/D-015, legal review and
  accepted joined-rights work remain; no product/release state changed.

## 2026-09-23 — selected receipt and response checkpoint

- **Coordinate:** product `7e8aa70`, planning after `14884fa`; predeclared
  `platform-alignment/data-rights-receipts-plan.md` and bounded result
  `platform-alignment/data-rights-receipts.md`.
- **Evidence:** five focused real-Postgres tests passed cold. Production
  Minigame API/session and core resolution receipt pairs matched in their
  separate fixtures (1/1 each). Soul's 14 terminal sessions held non-null
  progress tokens and matching core terminal receipt JSON (14/14). An exact
  Minigame parent-delete probe with API receipt children failed from the
  immutable-child trigger and rolled back. RP-127/RP-128 record the distinct
  token-retention and cleanup-migration boundaries. Temporary database
  service/network was removed; named cache retained.
- **Limit/next:** Guild intent receipts cascade with account rows, but shared
  Guild event history remains separate. None of the Minigame/Soul fixtures
  joins account deletion; other receipt variants and nested records remain.
  D-008/D-009/D-015 and legal review still own policy; no product, RFC or
  release state changed.

## 2026-09-23 — verification/board diagnostic checkpoint

- **Coordinate:** product `7e8aa70`, planning after `5714ee3`; predeclared
  `platform-alignment/data-rights-verification-plan.md` and bounded result
  `platform-alignment/data-rights-verification.md`.
- **Evidence:** three verifier and one leaderboard real-Postgres tests passed
  cold. A synthetic transient failure was retained exactly in queue/poison
  text; deterministic failure stored only its fixed verdict phrase. The board
  fixture had five rows, four event markers and exactly four allowed variable
  keys. The temporary Postgres service/network was removed, cache retained.
- **New boundary:** RP-129 covers unsanitized upstream error text in immutable
  poison history, not a proved secret leak. Neither these tests nor the prior
  rights tests delete an account after populating boards/dead letters. D-008/
  D-009/D-015, legal review, accepted retention contract and joined proof
  remain; no product, RFC or release state changed.

## 2026-09-23 — populated-board deletion link checkpoint

- **Coordinate:** product `7e8aa70`, predeclaration `d753360`; bounded dossier
  `platform-alignment/data-rights-board-delete.md`.
- **Evidence:** a temporary addition to the Account API integration test
  inserted one schema-valid board row joined to its account before the real
  DELETE. The cold declared Postgres lane passed without skip and observed
  `1` pre-delete join, then `1` retained board row linked to `1`
  archived/unlinked Founder and `2` archived streams after the account was
  removed. The diagnostic edit was removed exactly; no product/test diff
  remains.
- **New boundary:** RP-130 is an executed relational-retention fact, not a
  retention-policy ruling. RP-118 still needs the full projector, historic
  log/dead-letter and public-reader joined population. D-008/D-009/D-015,
  legal review, player workflow and clean-host evidence remain open; no RFC
  or release state changed.

## 2026-09-23 — Commons active-count deletion checkpoint

- **Coordinate:** product `7e8aa70`, predeclaration `f050057`; bounded dossier
  `platform-alignment/data-rights-commons-delete.md`.
- **Evidence:** the first temporary probe did not compile due to a test-driver
  field-name error. The corrected cold Account API/Postgres test passed without
  skip. An active Founder and its sample counted 1/1 before deletion; after
  account deletion archived its stream, the exact World-count and
  `refreshScope` sample-selection SQL still counted 1/1. A rollback-only
  `member=false` control made both zero. The temporary test change was removed
  exactly and the Postgres service/network shut down.
- **New boundary:** RP-119 is now verified at the deletion/predicate boundary,
  not as a full World/health/player workflow. Owner D-009/D-015 must rule active
  versus historical Commons semantics before a repair; a complete signed-event
  and recomputation witness remains. No product, RFC or release state changed.

## 2026-09-24 — Deployment designated review and RP-122 checkpoint

- **Coordinate:** HEAD after `c08b592`; product source still `7e8aa70`. Review and probe by Claude.
- **Deployment evidence:** Claude's designated cross-party pass executed the root lanes cold at
  each range tip and ran severing probes per range. The lanes were
  `test-deployment-backup/-release/-operations`, `verify-kernel-version`, `verify-ci-topology`
  and `verify-push`. Results:
  - DP-A–DP-E are CHANGES REQUIRED.
  - The corrective range `7b510df..cf4ac25` is APPROVED.
  - Verified defects persisting at HEAD make the `7e8aa70` candidate an unsuitable R-006 input:
    - the image-layer secret scan checks nothing inside the image;
    - release preflights before `docker load`;
    - rollback deletes the database volume before checking its inputs;
    - the rotation overlay is never applied;
    - the cleanup alert cannot fire;
    - receiver delivery counts attempts, not successes;
    - the Caddy admin API is reachable from peer containers.
  Details are in `deployment-foundation/log.md`.
- **Rights evidence:** RP-122 was predeclared at `8a02a3c`. A seeded cold joined Account API
  deletion returned 204 and kept eight Minigame/Soul/intent/outbox families 1/1, including the
  Soul progress token. The controls fired, and RP-128 still blocks the parent delete. The
  temporary test was removed exactly.
- **Decision support:** `platform-alignment/rights-decision-sheet.md` frames the D-008/D-009/D-015
  options and a ruling template. It adopts nothing.
- **Next boundary:** Codex corrective ranges for DP-A–DP-E, then fresh designated passes and a
  rebuilt candidate. DP-F review remains outstanding. Owner rulings are needed for D-007, D-008,
  D-009, D-015 and D-018, and clean-host authority for R-006. No RFC status, release or archival
  change.

## 2026-09-24 — Claude corrective implementation checkpoint

- **Coordinate:** `ec5518b`. Owner direction: Claude implements accepted-RFC work, and Codex is
  the cross-party reviewer.
- **Change:** R1–R17 implement every DP-A–DP-E blocking finding and the DP-F evidence-integrity
  items. Each batch has failing-first tests and severing probes (see
  `deployment-foundation/log.md`).
- **Evidence:** cold `make verify-push` passed. The backup, release, rehearsal and operations lanes
  passed. promtool killed 13/13 mutations.
- **Limits:**
  - Every retained bundle is invalid under the corrected rules, so rebuilds are needed.
  - There is no clean-host run and no Codex verdict yet.
  - Six DESIGN-GAPs are open for the RFC author.
- **Next:** Codex designated review of `67fd415..ec5518b`, then a rebuilt candidate/previous and
  owner authority for R-006.

## 2026-09-30 — resume audit after Claude's v0.1 wave

- **Coordinate:** product source `fc4191fed050f4adeadf3da333e7c1a3aa54d87b`, equal to
  `origin/main` before this record-only checkpoint. Worktree was clean. Claude's
  `planning/v0.1-claude-implementation-handoff.md` names implementation ranges; it is a
  handoff, not Codex's designated verdict. No release artifact was tested here.
- **Observed state:** the accepted v0.1 RFC index now lists implementing feature contracts and
  substantial code, API, client, content and test landings. This supersedes the prior board's
  “no accepted v0.1 RFC” assertion, but none of those new ranges has a complete Codex
  cross-party approval/archival union. Phase-0 and 1.0 remain unreleased.
- **Executed CI evidence:** cold `make verify-kernel-version` exited 2 at historical commit
  `50a3a514` (guarded `client/src/minigame/` paths without a same-commit bump). Thus the
  dependent `verify-client`/`verify-push` gate is red at current source; neither was claimed
  green. Cold `make verify-ci-topology` exited 0 and its 13 negative controls rejected. The
  passing topology check is only that check, not the whole CI lane.
- **Provenance finding RP-131:** the current kernel validator searches a `##` log section for
  literal `**Review by:**`, `**Decision:**` and the offending commit range. The `8add475`
  correction's log contains precisely those strings in an instruction saying the review is
  *still awaited*. The guard reaches the later `50a3a514` failure instead of rejecting that
  correction, so the provenance oracle is falsely satisfied. This is a source-and-execution
  inference, not yet a dedicated forged-review fixture. The correction's semantic code and
  tests have not received a Codex verdict in this checkpoint.
- **Targeted cross-party finding RP-132:** inspection of accepted Pet Adoption PA7/AC11,
  `c380896a`, current server schema/projector, generated client and docs found a different
  snapshot contract. Cold `make test-go GO_PACKAGES='./account'
  GO_TEST_FLAGS='-run TestGameUISnapshotAPIRegistryPinsTheProjectionEnvelope -count=1 -v'`
  exited 0 while the v4 fixture omitted `pet_adoption`; it cannot enforce PA7's sibling-field
  presence/shape. Codex recorded a CHANGES REQUIRED finding for that one commit/criterion in
  `pet-adoption-v1/log.md`, **not** a full-range Pet Adoption verdict. API C2's additive-only
  law is a genuine authority conflict requiring the ruling author's choice.
- **Limits and next:** no product code, RFC body, copy, release bundle or deployment changed.
  RP-131/RP-132 entered the shared ledger and the executable overlay was updated. Next is a
  ruled kernel-guard/PA7 reconciliation, then bounded independent range reviews (including
  Deployment R1–R17) with cold/severing evidence. Do not use the historical green CI claims or
  Claude's handoff as current-head release proof.

## 2026-09-30 — RP-133 rollback safety review and Codex correction

- **Coordinate:** targeted source review of Claude's R1 commit `1978660^..1978660` at product
  `fc4191fe`; Codex predeclaration `5f14a808` and correction `d19b5d8a` followed. No release
  artifact was rebuilt or deployed.
- **Cross-party verdict boundary:** Codex recorded **CHANGES REQUIRED** on R1's
  `VerifyRestoreInputs` safety claim. A second valid age key, mode 0600, passed host preflight
  for a backup encrypted to the first key. Temporary cold probe failed with the actual nil
  result; it was removed and the original focused test rerun green. Thus R1 could erase the
  database volume before the containerized restore rejected a wrong operator identity.
- **Correction and discrimination:** under accepted DP5/AC5, Codex added exactly-one-key parsing
  and a full authenticated decrypt to a discard sink before `StopFailed`/`ResetDatabase`. The
  permanent wrong-key test failed first on the old code, then passed with the correction;
  malformed/multiple identities and ciphertext whose outer hash was recomputed after mutation
  also refuse. Canonical Deployment docs and the owning log changed with code/tests. This is
  Codex-authored implementation, not Codex approval of itself.
- **Executed gates:** cold focused and full `deploymentrelease` Go suites passed; real
  `make test-deployment-release` passed through Compose/Postgres/Caddy. `make vet` failed before
  package analysis because the sandbox denied Go module-cache creation for `chromedp`; the
  focused `go test` ran its default package vet. RP-131's independent kernel-history failure
  remains, so current-head `verify-push` is not green. No RTO claim follows from the tiny test
  backup; a production-sized exact-bundle clean-host rehearsal remains required.
- **Next boundary:** Claude designated review of exact Codex correction range
  `5f14a808^..d19b5d8a`; Codex continues the independent review of the remaining Claude
  Deployment ranges. No RFC archival, candidate validity or release status changed.

## 2026-09-30 — zero-credit minigame and kernel correction designated review

- **Coordinate:** exact Claude implementation/correction range `8add475^..0cf9f7a`; product
  source for checks `d19b5d8a`. Codex's first actual independent verdict is in
  `minigame-platform-foundation/log.md` and approves only this two-commit scope.
- **Executed evidence:** cold Go zero-credit replay passed; a real Postgres session-resolution
  integration test passed without skip; the full client suite passed 6905 tests (76 skipped),
  including TS replay; the cross-runtime fixture passed cold; 6296 numeric vectors regenerated
  byte-identically. A temporary mutation changing the empty ledger receipt from `"0"` to `"1"`
  made the focused Go replay test fail; it was restored exactly. The dedicated Postgres service
  was shut down.
- **Boundary:** the genuine verdict now gives the `0.3.102` correction real provenance, rather
  than the validator's false-positive reading of a future-review instruction. RP-131's parser
  defect remains. A post-verdict cold `make verify-kernel-version` exited 2 at pushed
  `50a3a514` after both CI kernel-history fixture checks passed; the current-head full
  client/push gate therefore remains red.
  No wider Minigame, Deployment, CI or release claim changed.

## 2026-09-30 — RP-131 kernel-history repair drafted, not accepted

- **Coordinate:** after the genuine zero-credit correction review; the blocking history gate
  still stops at pushed `50a3a514` and no product/kernel code was changed for this draft.
- **Draft:** `rfc/kernel-history-guard-integrity.md` proposes a separate append-only,
  exact-commit/exact-path exception for independently proved presentation-only historical
  changes, plus a structured independent-verdict parser. It preserves KV-1's same-commit bump
  for semantic bytes and the existing correction record for real misses. The three known
  minigame/soul path-placement commits are candidates for audit, not pre-approved waivers.
- **Authority:** draft only; Marco must accept or narrow the proposed exception class and the
  whole historical population must be audited before implementation. No current CI, version,
  RFC implementation or release status is promoted by writing the proposal.

## 2026-09-30 — full Go vet rerun after cache permission denial

The prior `make vet` attempt failed before package analysis because the sandbox denied Go's
module-cache creation for `chromedp`. A retry of the exact root `make vet` target with narrowly
scoped module-cache access downloaded the declared dependencies and exited 0 (`go vet ./...`).
This closes the environmental verification gap only; `make verify-kernel-version` remains red
on pushed `50a3a514`, and no full `verify-push` success or release proof is inferred.

## 2026-09-30 — Deployment R2 targeted review finds post-stop rotation race

- **Coordinate:** Claude's R2 commit `92fab70^..92fab70`; predeclared probe `fe4fae24` on
  clean product source `d19b5d8a` (later commits to planning/RFC/review only). Codex's
  targeted CHANGES REQUIRED verdict is in `deployment-foundation/log.md` as RP-134.
- **Evidence:** the current R2 static overlay test passed cold and failed when Codex
  temporarily severed overlay insertion. The independent interleaving probe held the
  release lock, recorded a both-open rollback `down`, then successfully removed previous JWT
  through the different rotation lock after its governed minimum. Bootstrap stayed open;
  `ResetDatabase` refused before its volume operation. With no rotation the same steps
  proceeded; holding the rotation lock blocked removal. Both temporary probe and mutation
  were restored exactly.
- **Limit and route:** this proves a reachable recording-runner/control-flow interleaving,
  not a measured live-host outage. It is a new safety defect distinct from the already
  recorded single-family-overlay owner decision. The accepted DP4/DP5 repair must prevent
  rotation from changing the governed overlay mid-operation and retain a permanent negative
  witness. No Deployment range is approved or archived by this finding; R-006 remains open.

## 2026-09-30 — RP-134 shared operator lock correction

- **Coordinate:** Codex predeclaration `8e9b46d0`, implementation `2d8ca3f1`; exact range
  `8e9b46d0^..2d8ca3f1` awaits Claude's designated review. The source is not an R-006
  release artifact.
- **Implementation/evidence:** install, release, rollback, both recovery producers and
  rotation mutation now share the canonical release-ledger lock for each entire operation;
  rotation retains its own ledger lock. Permanent rollback and recovery interleaving tests
  failed on old code, passed after the correction, and both failed again when the new
  `RemovePrevious` shared-lock call was temporarily severed. The mutation was restored.
  Cold full `deploymentrelease`/`deploymentrehearsal` packages, real Compose/Postgres/Caddy
  release lane, `make vet`, and the root rehearsal target passed. The rehearsal target's first
  sandboxed attempt was denied an `httptest` loopback socket before assertions; its exact
  rerun with local socket access passed, including the old build-record validators.
- **Boundary:** this prevents supported helper commands from changing rotation state
  mid-operation; it does not decide the owner-held per-family overlay question, prevent
  manual operator-state edits, validate historical bundles as current candidates, prove
  clean-host rollback or clear RP-131's red kernel-history gate. No archival or release claim.

## 2026-09-30 — R3 scanner review and legacy-tar correction

- **Coordinate:** Claude R3 targeted review `2b6c0cb4^..2b6c0cb4` is CHANGES REQUIRED on
  RP-135. Codex predeclared and implemented a bounded scanner correction over
  `d871339d^..a7ab2640`, requiring Claude's designated review.
- **Evidence:** R3's original gzip/plain-layer and seeded-image tests passed cold and failed
  when recursive scanning was temporarily severed. A separate predeclared V7 layer with a
  nested gzip sentinel was initially an invalid fixture because default gzip exposed the
  raw sentinel; best compression and a raw-byte absence assertion corrected the driver.
  The USTAR control produced a finding, Go's reader accepted the checksummed V7 tar and
  payload, but the old scanner returned no finding and no error. The permanent V7 test
  failed first, passed after recognition by Go's tar reader, failed again with the old
  USTAR-only recognizer, and passed after restoration.
- **Cold gates/limits:** full `releasepackage` and `deploymentrehearsal` packages, tracked
  source scan (1776 files, no findings), and `make vet` passed. No current gameserver image
  archive, rebuilt candidate or clean-host R-006 artifact was scanned. This does not claim
  Docker currently emits V7 layers, only that the stated recursive-tar scanner contract
  was previously false for a Go-readable format. No Deployment archival or release claim.

## 2026-10-04 — Bounded Deployment R4 designated review

Codex designated-reviewed Claude's manifest cross-binding commit `751717f4^..751717f4`
and recorded **APPROVED** in the Deployment log. Cold forged-claim tests passed; removing
both new validation calls made all seven forged claims pass validation and failed the test,
then the source was restored and the package passed cold. This closes only R4's review
boundary. R1–R3 findings, the remaining Claude implementation range, the three Codex
corrections awaiting Claude review, RP-131's red CI gate and clean-host R-006 remain open.

## 2026-10-04 — R5 host gate and R6 checksum witness repair

R5's cold Go population passed, but the manual composed operations lane stopped before
`promtool`: the pinned `amd64` image returned `exec format error` on this `aarch64` Docker
host. R5 therefore has no designated verdict yet. Codex's targeted R6 review found RP-136:
the new checksum fixture changed payload length and stayed green when `ReadHeader` lost only
its SHA comparison. A predeclared test-only correction now uses different equal-length
ciphertext and fails under separate ReadHeader/Restore SHA severing probes; cold backup Go
packages and vet pass. Claude review of the correction and R6 closeout remain pending, as do
all clean-host/release gates.

## 2026-10-04 — R7 proxy/origin witness approved in isolation

Codex designated-reviewed Claude's R7 `63bed7dd^..63bed7dd`. Native Postgres test-image
refresh made the declared integration lane runnable on this arm64 host without changing repo
or release artifacts. The real-Postgres composed-server test passed cold; forcing zero trusted
proxy hops failed its second-client limit assertion, and replacing the allowed origin failed
the configured-origin WebSocket upgrade. Both mutations were restored and the test passed
again. The Deployment log records **APPROVED** for R7 only. The Caddy/clean-host deployment
path, R5, R6's cross-party correction review, RP-131 and R-006 remain separate gates.

## 2026-10-04 — R8 source-map completeness review and correction

A fresh client build and metadata run produced 53 dependencies, including the transitive
`pad-end` and `tslib` notices. Codex's predeclared second-JavaScript-asset fixture then showed
the R8 extractor silently accepted an emitted chunk without its map (RP-137). The accepted-DP2
correction requires linked maps for every JavaScript output and rejects orphan/wrong-linked maps;
the new negative fails if the gate is severed, while cold Go/vet and the fresh generator pass.
The Deployment log records a targeted CHANGES REQUIRED verdict for Claude's R8 and routes the
Codex correction to Claude. CSS-only dependency attribution, full packaging-rights review,
R-006 and the 1.0 release claim remain open.

## 2026-10-04 — R9 readiness-origin evidence repaired locally

R9's Caddy-facing down poll accepted any 503, although the gameserver also returns 503 for
non-drain failures. Codex's predeclared unmarked-503 case failed first (RP-138). A drain-only
response marker now distinguishes the gameserver's ordered withdrawal; the observer requires
both status and marker. Producer and consumer severing probes fail independently, and cold
gameserver/release/config tests plus vet pass. Claude cross-party review and the exact amd64
Caddy/Postgres composed run remain pending; this is not R9 or release approval.

## 2026-10-04 — R10 host-byte proof narrowed; D-019 queued

Codex executed the cold R10 release/rehearsal packages and severed the new host-envelope
re-read: the differing-header witness failed, so that integrity check is real. The same
positive fixture succeeds with an invalid restore identity, however, because DP6 keeps the
identity off-host during backup creation. This cannot prove a working rollback before drain.
RP-139 records a targeted CHANGES REQUIRED verdict; canonical Deployment docs now limit the
claim to host-byte checksum/header agreement. D-019 asks the owner what per-release
decrypt/restore proof is required. No new key-mounting mechanic was inferred, and R-006 remains
the exact clean-host recovery gate.

## 2026-10-04 — R12 plan producer identity is still forgeable

The cold R12 rehearsal/CLI suites pass, including a plan-executor test that runs every canonical
row through a generated shell script named `deployment-rehearsal` and hard-codes the expected
0/3 exits. The new basename/subcommand gate rejects obvious `/usr/bin/true` substitutions but
does not bind the executable bytes to the candidate bundle. Codex recorded targeted CHANGES
REQUIRED (RP-140) and DP-F DESIGN-GAP 8 for the RFC author: define candidate manifest and
helper identity authority, then require a same-basename fake-binary negative. No R-006 plan
proof is promoted from the current test.

## 2026-10-04 — R13 recovery/install safety approved in isolation

Codex designated-reviewed Claude's R13 `0ce7508e^..0ce7508e`. Cold release/rehearsal
packages passed. Removing Alertmanager from recovery startup failed the dependency witness;
narrowing initial-install volume inspection to Postgres failed the retained-certificate case.
Both were restored and tests reran green. The exact Docker prefix filter also returned this
project's existing volumes on the live daemon. The Deployment log records **APPROVED** for
R13 only; no actual ACME-bearing volume was deleted, and clean-host R-006, the wider review
union and the unresolved R12 producer-identity gate remain open.

## 2026-10-04 — R14–R18 bounded review; two proof gaps corrected, one remains

Codex reviewed Claude's R14–R18 commits in sequence. R14's build-record-to-bundle
comparison is sound in the bounded unit lane, but its original tests did not exercise the
previous-release branch: removing that branch left the suite green. A Codex-added
self-consistent previous-record corruption now fails on the severing and passes restored.
R15's probe test likewise passed without its manifest-rebind call; a Codex-added artifact
and SBOM hash assertion now fails on the severing. Both Codex test corrections await Claude
cross-party review. The R16 envelope-commit and R17 browser-manifest batches received
bounded designated approvals with executed severing checks; neither establishes R-006.

R18's migration and missing-rollback-input negative producers have no invoking tests.
Replacing both dispatch arms with unconditional `ProbeRejected, nil` left the cold rehearsal
and CLI packages green. RP-141 records CHANGES REQUIRED: add reproducible positive/negative
fixtures and producer severing before these rows count. Earlier diagnostic bundles were not
retained, so their recorded CLI exits cannot substitute for release-artifact evidence. The
Deployment review union, R5 composed amd64 lane, D-019 recovery-proof choice, kernel CI
repair, owner/privacy/accessibility decisions and real clean-host R-006 all remain open.

## 2026-10-04 — R19 lifecycle contract approved in unit scope

Codex designated-approved Claude's R19 lifecycle release/rollback producer commit
`1a8375af` in isolation. Reinstating the old rollback-row field clearing failed both
controller and lifecycle tests; omitting the release transition failed the lifecycle
two-row predicate. All four affected Go packages passed cold when restored, and CI
topology validation passed. The fixture uses a fake runtime, so real Caddy drain,
Postgres restoration, timing and host artifact evidence remain R-006 obligations; the
R12 and R18 proof gaps are unchanged.

## 2026-10-04 — R20 dirty-target refusal attribution corrected locally

Codex's predeclared false-positive test showed the R20 rehearsal classifier accepted an
unrelated error that merely echoed `"error_class":"non_clean_target"` (RP-142). The
correction carries typed command/exit/stderr provenance and parses the exact structured
backup restore refusal. Cold Go suites and vet pass; the old substring path fails the
new test. The pinned amd64 Postgres target exited before testing on this arm64 host, so
Claude review and a working amd64 composed/host run remain necessary before R20 or
R-006 can count. No overall release status changed.

## 2026-10-04 — R21 clean-start volume boundary approved in isolation

Codex designated-approved Claude's R21 `a1da95c5`: the cold host-observer test fails if
the check is narrowed back to Postgres while a retained certificate volume exists. The
live Docker filter also returned three project cache volumes, confirming it observes
non-Postgres state on this machine. That demonstrates the rule, not a clean host;
R-006 and the remaining evidence/decision gates are unchanged.

## 2026-10-04 — R22 restart row does not exercise admitted work

Codex's targeted R22 review found RP-143. The new producer proves SIGKILL is not a
graceful drain and that service can restart, but it never admits or holds a write and
never checks that write after restart. Its cold Go suites pass without any in-flight
operation. Canonical Deployment docs now state that limitation; DP-F must define and
implement the exact mid-write/restart witness before this named R-006 row can count.

The R-006 plan now fails closed: a failing-first plan test showed that the crash-only
helper was accepted as the admitted-work producer; Codex removed that binding, and
reinserting it fails the test. The crash command remains diagnostic, while the named
row routes to an unsupported exit-2 probe until its real fixture exists. Claude review
of `95d1bb60^..81a970ce` and the DP-F author contract remain open.

## 2026-10-04 — F9 active Pitch preview proven in the composed player path

Codex designated-approved Claude's bounded F9 Wind Down preview/composition commit.
The prior composed Pitch test ran at Tier 0 and could not discriminate the active
session rule (RP-144). A new Tier-1 Game UI/browser/Postgres path observes Wind Down
eligible before Pitch, ineligible during an active session, and eligible again after
the terminal receipt. Removing the `!minigameActive` predicate fails the middle
assertion; the restored full composed target passes. Claude must review the Codex
test-only range `f773cf07^..5ad457ce`. This is one player-facing seam, not closure
of MA AC1–AC5, the v0.1 handoff, accessibility or 1.0.

## 2026-10-04 — Soul surface lifecycle corrected; MA review still open

Product source `47562850`. Codex targeted Claude's MA3 Soul surface and found two
player-lifecycle defects: an already-hidden document still sent recovery beats
(RP-145), and deferred start/reconnect responses restarted the heartbeat
scheduler after unmount (RP-146). Predeclared browser negatives fired before
each correction. Removing the initial visibility callback after the fix
produced two hidden-tab beats; the deferred-response tests observed three
post-unmount beats after start and one after reconnect on the original code.

Codex's bounded corrections `9ceed1cf^..47562850` now pass cold typecheck,
6,905 client tests, boundary/copy checks, client build, and full Chromium and
WebKit browser suites (6,983 tests each, plus the performance lane). The
composed Postgres/Pitch path passed before the second correction and does not
exercise Soul recovery. Local Playwright Firefox aborts before test import, so
the three-browser gate remains unverified. A cold `make verify-kernel-version`
at this checkpoint again exited 2 at historical commit `50a3a514` (the
original presentation files were under a kernel-watched path without a real
version bump); `verify-push` therefore remains red. **Review by:** Codex on Claude's
targeted Soul code; **recorded by:** Codex. The Codex-authored correction awaits
Claude's exact-range designated cross-party review. The larger MA acceptance
set, release-artifact workflow, historical kernel CI failure and 1.0 rights,
access, operations and content floors remain open; no release claim changed.

## 2026-10-04 — Fiscal→Pitch default player journey closes a real sequencing gap

Product source `efe68333`. The existing composed Pitch witness used direct API
mutations for Fiscal harvest and unlock even though the Earnings Calls surface
was implemented. Replacing that setup with rendered player controls caused
the immediately following unlock to return `revision_conflict` against a stale
Founder revision (RP-147). This violated accepted Garage Surfaces GS0.2's
post-apply authoritative-refresh rule. A browser runtime-double test failed
first because no refresh was requested; after the host held pending through
refresh, the next DOM unlock used the new revision. The composed real
Postgres/Vite/gameserver journey now applies both DOM-issued Fiscal intents,
starts Pitch, sends five keyboard-driven commands, reaches an applied terminal
receipt and observes the refreshed Company snapshot. Disconnecting either
Fiscal button fails the composed test before Pitch can start.

Cold typecheck, 6,905 client tests, boundary/copy checks, client build, full
Chromium and WebKit suites (6,984 tests each plus the performance lane), and
the unmodified composed target pass. The local Firefox launch failure and
historical kernel-history CI failure persist; neither is counted green.
**Review by:** Codex on the targeted Claude implementation and old witness;
**recorded by:** Codex. Codex's corrective `efe68333` range awaits Claude's
cross-party designated review. Neither complete RFC range, a release-artifact
journey, nor the 1.0 rights/access/operations/content floors are closed.

## 2026-10-04 — Soul recovery keyboard component witness

Evidence source `33ced43d`; product source remains `efe68333`. A new real-browser
MA3 case activates Soul recovery Begin and Stop early with Enter only, observes
the exact start/cancel calls and terminal focus, and runs axe in both active
and terminal states. Disconnecting the visible Begin control fails at zero
starts; restored Soul suites pass 9/9 in Chromium/WebKit. Cold
`make test-client` passes 6,905 tests after a browser-only import was moved
inside the skipped test body; full Chromium/WebKit populations pass 6,985
tests each plus the performance lane. The composed Minigame API lifecycle
passes cold against real Postgres, and `make api-check` is byte-clean.

**Review by:** Codex. **Recorded by:** Codex. Claude's exact-range review of
this test-only addition remains required. This is a component keyboard proof,
not Firefox, assistive-technology task, real-server Soul UI or wider MA
range-union acceptance evidence. The current default Firefox launch failure,
historical kernel CI failure and 1.0 floor remain open.

## 2026-10-04 — Typer prompt announcement and keyboard-evidence correction

Product/evidence source `9b7137c2`. A targeted Codex review of Claude's Typer child
`ade1083b` found that TT8.3's new-prompt announcement was absent and the browser
"begins by keyboard" test used `.click()`. RP-149 and the Typer log record the
CHANGES REQUIRED slice, predeclaration and original failing browser output. A
test-only harness advances prompts in one mounted instance; the initial check
failed with zero prompt live regions. The correction maintains one initially
empty polite region and updates it on prompt-ID changes while focus stays in
the input. Removing its assignment made the same check fail again. The ready
and line-submit controls now have actual Enter-driven browser witnesses.

Cold full Chromium and WebKit browser suites each pass 6,986 tests plus the
performance lane; client suite passes 6,905; typecheck has zero diagnostics;
boundary, copy and build checks pass. Firefox remains unavailable on this
host, and no manual assistive-technology task was performed. The Typer child
is still unregistered, no public v1 route carries its commands/snapshot, and
content is provisional; the TT-PA4/C2 owner contract and broader B1–B7
designated review remain open. **Review by:** Codex of the targeted Claude
slice; **recorded by:** Codex. Claude must cross-party review Codex's
`9b7137c2` corrective range before it is accepted as an implementation batch.

## 2026-10-04 — Typer valid-UTF-8 differential and historical CI boundary

Product/evidence source `c9040bbe`. Codex's targeted review of Claude's B1
found that Go and TS rejected valid U+FFFD even though TT4.4 excludes only
malformed UTF-8 and C0/DEL. Failing-first Go/TS tests and a raw-byte probe
confirmed the valid/malformed distinction. The correction checks raw JSON
before Go decoding, accepts valid raw/escaped U+FFFD as a scored miss, keeps
malformed bytes and lone surrogate escapes rejected, adds a shared corpus
vector and bumps the Go/TS kernel identity to 0.3.137 in the same commit.
Severing the raw guard makes the malformed arm fail. Cold Go/vet, client,
typecheck, corpus, full Chromium/WebKit (6,987 tests each) and the real
Postgres Typer composed test pass. An initial Docker selector ran zero tests
and was discarded; the corrected `./production` selector executed and passed.

`make verify-kernel-version` still exits 2 at the earlier pushed
`50a3a514` against `0cf9f7a6` (six `client/src/minigame/` guarded paths
changed without a same-commit version bump). The new `c9040bbe` commit
contains its own same-commit version bump, but the script stops at that
historical violation before it can certify the whole range. No green CI
claim is made. **Review by:** Codex on the targeted Claude B1 slice;
**recorded by:** Codex. Claude's cross-party review of `c9040bbe` is still
required; wider Typer acceptance, public wire, owner content, Firefox and
the complete 1.0 floor remain open.

## 2026-10-04 — Typer B2 time witness gains a real injected DB clock

Evidence source `589fc06c`; product source remains `c9040bbe`. Codex's
targeted review of Claude B2 `3eb7e401^..3eb7e401` ran the cold real-Postgres
AC3 witness and independently severed nonterminal persistence, terminal
persistence and replay-time propagation; each failed, then the restored
witness passed. RP-151 recorded that the test still lacked AC3's named
injected-clock population. A temporary production query seam and exact DB
sequence worked but was withdrawn: it changed a guarded kernel file without
production behavior, which cannot honestly force a version bump.

A test-only Postgres schema now shadows the unchanged unqualified DB clock
query on one connection, returning one fixed DB-derived time a day ahead.
The new composed population proves that value reaches all three tenant
commands, their persisted stamps and verification replay. Bypassing the
shadow via `pg_catalog.clock_timestamp()` fails its preflight. The original
real-clock/stall population remains, because a fixed injected clock alone
would not detect insert-time resampling. Both named populations, the full
`./minigame` real-Postgres integration subset, non-Postgres package and vet
pass cold. No production, schema, kernel, API or copy byte changed.

**Review by:** Codex on the bounded Claude B2 mechanism; **recorded by:**
Codex. Claude's cross-party review of the Codex test-only `589fc06c` range
is required before B2/AC3 closes. Typer public wire, owner content, Firefox,
the historical kernel-history CI failure and the whole 1.0 floor remain open.

## 2026-10-04 — Typer B3 exact unlock parity and API-path proof

Product source `a33d4d6a`, evidence through `76f8cc1d`. Codex found that Claude's B3 Go loader treated explicit `exit_history_at_least:null` as an absent clause, while the TS loader rejected it. A Go negative failed first; the bounded correction rejects null, adds both-language negatives, updates canonical docs and advances the watched kernel identity to 0.3.138. Cold `./kernel ./minigame ./production`, client units, typecheck and vet pass. The whole kernel-history gate remains red at historical `50a3a514`, before judging the new commit.

The earlier real-Postgres Typer AC9 test used the legacy start path, not the B3 API resolver. A new test drives `StartMinigameAPISession` with pinned Typer content and Founder v21: Tier 0 and no-Exit starts reject without a receipt, session or Founder advance; Tier 1 with one Exit starts. The original legacy v20 population remains intact. Both populations pass cold on real Postgres, and independent Tier and Exit severing mutations each fail the new test before restoration. This proves the internal API-start gate, not a public route or player workflow.

RP-153 separately records that the Game UI availability projector still treats a non-Fiscal Typer row as unlocked, contrary to the server gate. The accepted Garage GS7 surface is Pitch-only; a reconciled surface contract and same-row parity proof are required before Typer's player-facing unlock can be claimed. Claude must cross-party review both new Codex ranges; wider Typer review, public wire/content mint, Firefox, CI-history repair and the full 1.0 floor remain open. No push or release action occurred.

## 2026-10-04 — Typer B4 loader chain parity

Product/evidence source `19f71825`. Codex's targeted review of Claude's B4 loader chain found that Go required the `typer` definition to name engine `typer@1.0.0`, while TS accepted a hash-consistent `typer` ID with the wrong engine ref or version. The wrong-ref TS negative failed first; the exact-identity correction and matching Go/TS negatives pass cold, and removing the TS version check makes its own negative fail. Client tests, typecheck, Go replaycatalog/kernel and vet pass. Kernel identity advances to 0.3.139 because the TS replay loader is watched. RP-155 records the correction; Claude's cross-party review of this Codex range remains required. No production Typer epoch, public route, player workflow or 1.0 release claim is inferred.

## 2026-10-04 — Typer B6 child keyboard completion proof

Evidence source `dfcc6270`; product source remains `19f71825`. The Typer child browser suite previously used real Enter for Begin/Submit but did not test native keyboard activation of End run or Exit to host (RP-156). A test-only Chromium/WebKit case now uses Enter and Space on those focused native buttons. Both callbacks were separately severed in temporary product mutations and each made the new case fail; both mutations were restored. Targeted 7/7 browser cases plus the separate performance lane pass. The first test draft used the wrong visible label, failed, and was corrected before the passing run. This is child wiring evidence only; public Typer routing, host registration, owner-authored copy/content, full accessibility task evidence and Claude's cross-party review remain open.

## 2026-10-04 — Typer B7 real Exit before/after

Evidence source `ec2dfbf9`; product source remains `19f71825`. The old service-level Typer AC8 test only inspected repository activity and constructed production without the activity resolver. A real `wind_down` request failed first as resolver-unavailable. The corrected fixture binds the repository, uses a second-run Company for the synthetic prior Exit, and keeps the staged Founder/clock at a valid v21 current coordinate. The live Exit now rejects while Typer is active without Company/Founder advancement; after resolved `end_run`, a second Exit applies and appends one Founder record. A temporary false activity result made the first Exit apply and failed the new assertion; the production mutation was restored. Both `TestTyper` populations pass cold on real Postgres, along with non-Postgres production tests and vet. This is not the full AC8 claim: faucet-cap forfeit, offline-quality charge and public MA endpoint journey remain open. Claude's cross-party review of the test-only Codex range is still required.

## 2026-10-04 — Typer B7 daily cap and offline-quality evidence

Evidence source `1260be38`; product source remains `19f71825`. Six same-day Typer sessions on one Founder now exercise the real Postgres faucet: the first five credit cash, the sixth credits zero and reports positive configured forfeit with `cap.minigame_faucet`. Each run's receipt is compared with the Company cash delta; the first clean-seven result charges the Founder's Typer offline-quality grade to 800,000 ppm. Severing the live faucet boundary (`<` to `<=`) fails the sixth-send assertion; shifting the quality grade by one fails the grade assertion. Both mutations were restored. Cold `TestTyper` integration, non-Postgres production tests and vet pass. RP-158 tracks the bounded witness; it does not prove the public MA endpoint journey, settle the Typer wire contract, or waive Claude's designated cross-party review. B7 and 1.0 remain open.

## 2026-10-04 — Typer B7 replay quota-forgery negative

Evidence source `11c7af4b`; product source remains `19f71825`. A separate fixture-derived validation negative now accepts the honest exhausted-quota forfeit and rejects a forged sixth credit. Broadening the replay comparison from `<` to `<=` makes the new test fail; the mutation was restored. This resolves RP-159's narrow test gap, not a full forged-log end-to-end or public MA journey. Cold production tests, real-Postgres `TestTyper`, and vet pass. Claude's designated cross-party review of this Codex range and the full 1.0 release floor remain open.

## 2026-10-04 — Typer B5 partial cross-party verdict

Codex designated-approved Claude's `d7c1ce6e^..d7c1ce6e` error-detail increment after cold account/public API tests, unchanged generated API artifacts and v1 pin, and a tier-detail mutation that failed its exact 409-body assertion. The precise verdict is in `planning/minigame-terminal-typer/log.md`. The approval is only for B5's response-enum and handler mapping subset; the TT-PA4 versus API C2 public-wire conflict still blocks request/snapshot arms, generated client and the MA-endpoint acceptance path. No B5 checkbox or Typer lifecycle status changed.

## 2026-10-04 — Typer B1 corpus-budget review finding

Evidence source `64467fd8`; product source remains `19f71825`. Codex's targeted review of Claude B1 found AC12's transition budget violated its stated population: 67 corpus command attempts but only 62 applied transitions counted. A Go equality assertion failed first; the bounded test/corpus correction sets the budget to 67 and counts all attempts in TS. Restoring the old TS counter fails its gate 62/67; no mechanics or content scenarios changed. Cold Go Typer, full client, typecheck and vet pass; the separate corpus-check target passed with cache reuse. RP-160 records the finding, B1 is now marked partial, and Claude must cross-party review this Codex correction. This is not a full B1 verdict or 1.0 proof; the public wire, content mint, accessibility and CI-history issues remain open.

## 2026-10-04 — Typer B1 pinned tier-clamp correction

Product/evidence source `cdc6c20c`. TT1's scaler clamps `typer.era_tier` to 1..9, but a valid tier-zero-era content fixture let both pure engines create a tier-zero session. Matching Go/TS negatives failed first. The bounded correction enforces the named minimum in both engines and Go's standalone snapshot validator; restoring that snapshot validator's old comparison fails the separate negative. Kernel identity moves to 0.3.140 with canonical Typer docs. Cold Go decimal/Typer/kernel/replay/minigame, full client tests, typecheck, build, vet and decimal vectors pass. Full `make verify-kernel-version` still stops at the older pushed `50a3a514` violation (RP-131); it is not a green gate for this commit. RP-161 and the Typer plan keep B1 partial pending Claude's cross-party review of this correction and Codex's review of the remaining B1 range. No public-route or release claim follows.

## 2026-10-04 — Typer B1 result-bound authority gap

Codex's B1 review found RP-162: TT4.7 says `ValidateResult` checks `lines_cleared ≤ run_length`, while the shipped `Tenant.ValidateResult(*Result)` interface carries no pinned run length. A temporary cold diagnostic showed a nine-line result accepted against the fixture's eight-line run, then was removed with the tree clean. The normal engine still stops at eight and the platform does not accept client-submitted scores, so this is a false independent-validator claim, not a demonstrated payout exploit. The Typer/Minigame Platform ruling authors must choose a context-bearing validation route or reconcile an engine/replay-owned bound and its negative witness. B1/AC7 stays partial; no spec or product change was made for this gap.

## 2026-10-04 — Typer AC2 seeded-order evidence

Codex independently severed `typer.prompts.v1` in Go and TS one at a time. The Go corpus gate failed stale; TS failed both the first replayed terminal state and its explicit dealt-order check. After restoration, cold Go content-gate and full client tests passed. The detailed targeted evidence is in the Typer log. This verifies AC2's current mechanism only, not a designated verdict for all of Claude's B1 commit or an archival gate. B1 remains partial on RP-162, cross-party review of Codex corrections, and the rest of the range.

## 2026-10-04 — Typer AC4/AC5 timing evidence

Codex separately severed the exact-deadline comparison and backwards-clock max in Go and TS. Each of the four mutations failed the relevant unit or cross-runtime corpus witness; restored cold Go Typer and full client suites pass. The Typer log has the specific failure coordinates. This is bounded evidence for AC4/AC5 mechanics, not a full B1 verdict or an answer to the server-authored DB clock (AC3). RP-162's content-free result-validator gap and the public route remain open.

## 2026-10-04 — Typer TT4.5 replay-state parity

Product/evidence source `505c4e3b`. Codex found RP-163: TS `applyTyper` accepted malformed snapshots that Go refused, including negative counters, invalid assist/submission shapes, and counters above the pinned catalog cap; a negative-miss state resolved to `typer.misses:-1` before correction. Matching Go/TS negatives and a bounded TS decoder/catalog validation correction now refuse these states. Removing the new TS cap guard and submission guard separately makes their negative rows fail; both mutations were restored. Kernel identity advances to 0.3.141 with canonical docs. Cold Go decimal/Typer/kernel/replay/minigame, full client tests, typecheck, build, vet, vectors and client-boundary check pass. Whole-history kernel verification remains red at pushed `50a3a514` (RP-131). Claude must cross-party review the Codex range, and B1/AC7 still waits on RP-162's contract decision. No public Typer route, content mint or 1.0 claim follows.

## 2026-10-04 — Deployment packaging-rights CSS provenance

Codex checked the accepted Deployment DP1/AC8 rights boundary and found RP-164:
the current client emits CSS without a source map, and the release-metadata
generator accepted an added unmapped stylesheet. A bounded build/metadata
correction now records final CSS asset hashes and package stylesheets reached
through both JavaScript imports and nested CSS `@import`, then includes a
CSS-only package's license in the delivered notices. The nested-import test
first failed on the initial hook, and separate Vite/Go severing probes failed
after correction. Cold client, typecheck, focused Go, vet, real build and
metadata generation pass at 53 dependencies. This is locally implemented,
pending Claude's designated review; URL/worker CSS, final bundle/image rights
and clean-host R-006 remain open. The unrelated pushed kernel-history gate
still fails at `50a3a514`, so no green full CI or 1.0 claim follows.

## 2026-10-04 — Unmerged Claude worktree reconciliation

Main had no intervening Claude commit, but three clean Claude worktrees from
2026-09-24 were found. Two hold bounded RP-118/RP-125 data-rights diagnostics
with real Postgres observations on an arm64 override and no retained executable
test; their old baselines and limits are recorded, not promoted to current
rights/release proof. The third holds an unaccepted R-006 options draft.
Codex reviewed its exact commit `cd3235a7` cross-party and recorded CHANGES
REQUIRED: it omits the current producer-identity, admitted-work, restore-
authority and negative-producer gaps while claiming complete producer
coverage. All three worktrees remain untouched and unmerged. The exact
inventory and routes are in `planning/platform-alignment/log.md`; no owner
policy, RFC acceptance or product behavior was inferred from them.

## 2026-10-04 — CSS URL rights follow-up

The RP-164 follow-up found RP-165: a Vite-built CSS `url()` inlined an npm
SVG into shipped CSS while the v1 package graph omitted the asset-only
package. A base64-decoded shipped-byte control made the omission test valid.
The v2 graph and Go metadata reader now include such package resources and
reject a missing file or stale v1 record; separate build-hook and Go-consumer
severing probes fail. A worker-CSS arm did not fire: its JavaScript source map
already named the package stylesheet, and a Go fixture confirms that map
source reaches a notice. Current metadata remains at 53 dependencies with no
package CSS URL resource. Cross-party review, final bundle/image rights and
clean-host AC8 remain open; this does not advance a release claim.

## 2026-10-04 — Worker CSS source-map inference retracted

The prior CSS URL checkpoint inferred that worker CSS was covered because
the raw source-map file contained the package path. Parsing its `sources`
array disproved that: only the worker's import text carried the path, while
the package CSS sentinel was actually shipped. RP-166 adds a shared Vite
worker-build provenance hook and a Go notice fixture with no package source
in the worker map. The raw-string inference is withdrawn, not silently
overwritten. This remains locally implemented pending cross-party review and
the exact release-bundle rights gate.

## 2026-10-04 — Cosmetic Shop keyboard evidence correction

Codex's bounded C6/AC11 review found RP-167: the claimed keyboard Buy browser
witness focused the button but used `.click()`. A test-only Enter/Space
supplement now passes in Chromium and WebKit and fails both cases when the
native Buy handler is severed. Firefox could not connect to Vitest on this
host, even when run alone, so the three-engine acceptance gate remains open.
Claude's designated review of the Codex test range, the real-server AC14
Buy→reload flow, owner copy adoption and live pet-panel overlay are still
required; this is not a Cosmetics or 1.0 completion claim.

## 2026-10-04 — Cosmetic Shop network-witness scope corrected

RP-168 found that the C7 browser trap saw `fetch` and WebSocket but not XHR,
beacon or resource requests despite its all-HTTP claim. A browser-page
observer and a four-transport disallowed-path negative now pass in Chromium
and WebKit and fail when the observer is severed. The first unreachable-port
negative did not produce WebKit request events; off-origin attempts blocked
before network dispatch, Firefox and the real-server Buy→reload flow remain
unproven. This moves N5 evidence forward without claiming AC13, Cosmetics or
release completion.

## 2026-10-04 — Cold Linux browser-lane correction

Product source `2ea9f316`; test/evidence source `c3311dbd` (coordinate record follows). No newer Claude commit is present on main: the three old Claude worktrees are already inventoried in the platform-alignment log, and the main checkout was clean at `6e6b1786` before this batch. Native Firefox launch timed out both sandboxed and unsandboxed on this macOS 27 host; `~/Library/Application Support/Firefox` is OS-denied to the launching process. The cached ARM64 Linux Playwright image provided a safe three-engine runner without touching user browser data.

The first cold `make test-browser-ci` found two new evidence defects: a Node-only CSS provenance fixture imported and failed before assertions in all three browsers (RP-169), and the Cosmetics four-transport negative failed in full-suite Chromium/Firefox despite passing alone because its 100 ms stop did not await actual request observation (RP-170). An exact browser-config exclusion preserves the CSS fixture's four passing Node cases. A nonce-tagged, bounded request-audit wait now requires all four URLs. Removing the page request listener made the focused test fail in Chromium, Firefox and WebKit with an empty observed set; it was restored. The corrected cold Linux lane passed 20,976 browser tests with three skips plus the separate performance case; `make test-client` passed 6912, typecheck had zero errors/warnings. This is local Linux evidence, not hosted CI, Cosmetics AC14, final packaging rights, or 1.0 release readiness. Claude's designated review of the Codex test correction remains due, and the historical kernel-version gate remains red.

## 2026-10-04 — current-HEAD public-board deletion boundary

At product source `f9ab1037`, the retained Account AC6 supplement runs the composed first-hour
script against real Postgres, naturally verifies/archives two runs, projects board rows, then
deletes the authenticated account. The account vanishes and its Founder/streams archive, but the
same Founder/run remains on the unauthenticated public HTTP board (RP-171). A temporary reader
severing makes the post-delete assertion fail; different board variables exclude the row. Cold
gameserver/leaderboard/account Postgres controls, focused Go tests and vet pass. The exact census
and controls are in `platform-alignment/data-rights-public-board-delete.md`. This replaces the
old “no public-reader proof” caveat but does not close RP-118's dead-letter/poison or backup-restore
arms. D-009/D-015 and legal review remain owner choices, and Claude's designated cross-party
review of the Codex range remains required. No 1.0 or Account archival status changed.

## 2026-10-04 — current-HEAD pre-deletion backup resurrection boundary

The retained RP-125 real-Postgres test now brackets Account repository deletion with age-encrypted
B_pre/B_post and restores both through the product backup package. Source deletion removes the
account/active Founder link; B_pre restore re-creates them and live streams, whereas B_post keeps
the deleted/archived state. A bystander and verified board row persist in both. The complete
five-test integration population passed without skip on an isolated ARM64 Postgres project;
temporarily severing the account-row deletion made the test fail at the source census, then
production bytes were restored and the suite passed again. This is not the old Claude branch's
credential-revival proof at current HEAD, nor supported-host R-006 or a D-009/D-015 ruling.
The exact counts/limits live in `platform-alignment/data-rights-restore-resurrection.md`.
Cross-party review of the Codex range remains open; no release status changed.

## 2026-10-04 — Cosmetic AC14 composed witness and Pitch CI reliability boundary

The accepted Cosmetic Shop AC14 now has a retained built-client/Chromium → composed gameserver
and Postgres witness on a synthetic current-epoch artifact set: live WebSocket visitor presence,
T0 locked shelf, visible T1 cross-gate and Buy, one applied `acquire_cosmetic` intent and event,
then server-owned state after a full reload under the strict N5 request/payment audit. The
synthetic Cosmetics pin is fixture-only, not a production content mint. Temporarily severing
`server/gameui/features.go` Cosmetics projection fails the first arm check; production bytes were
restored. The exact `make test-game-ui-composed` target with a committed ARM64 image override
passed both drivers in 17 seconds, and `make verify-ci-topology` plus typecheck passed. This is
local first-filter evidence; Claude's designated cross-party review is still due.

The older composed Game UI driver independently failed on intermittent no-request Pitch Start
and Fiscal Unlock clicks in multiple cold runs, even though it also passed several runs. In one
failure, the Fiscal unlock remained enabled with `main[aria-busy=false]` after its 30-second
wait. An explicit reset of the named ephemeral test DB prevents the synthetic Cosmetics epoch
from contaminating the older driver but did not remove the intermittent failure. Temporary
`act` logging produced only passing runs and was restored byte-exact, so handler entry/drop was
not established. RP-172 records the unresolved CI reliability boundary; no retry or relaxed
assertion was added. Separately, the historical kernel-version gate remains red, and no hosted
CI success or release status is inferred from this local target pass.

## 2026-10-05 — RP-172 composed Pitch click classification and correction

The test's capture-phase trace caught the alleged Fiscal Unlock “click” as only pointerdown/up
while the button was disabled and `main[aria-busy=true]`; there was no click event. The button
was enabled after the 30-second timeout, which made the previous post-hoc diagnostic misleading.
A separate cold run found the legitimate Exit offer preempting Pitch before Start. The test-only
Garage/Minigame correction now dispatches one exact DOM click only when the host is idle and the
control enabled, and declines any pre-click offer through its visible UI intent before returning
to the surface. It never retries after a click was dispatched. Three independent cold runs plus
the combined Cosmetic/older composed Make target passed; disconnecting Fiscal Unlock's handler
then produced an enabled click event but no request and failed at the intended oracle. The handler
was restored byte-exact. The final combined Make run naturally exercised one visible offer decline
and still passed Pitch and Cosmetics. A forced-offer negative has not run, and hosted CI has not
run. This is a narrowed test timing defect, not evidence of a product `act`
drop or a complete CI reliability verdict; Claude's designated review is still due.

## 2026-10-05 — Cosmetic C5 producer/reader contract correction

Codex's cross-party review of Claude C5 found three public-wire forms that the TypeScript
decoder accepted despite the accepted contract: reversed and duplicate per-item `worn_by` pet
IDs, and lock tier 9. Three new tests failed first; the bounded decoder correction, shared
Go-projected two-wearer fixture and canonical docs now agree. Separate order and tier severing
probes fail their targeted cases. Cold client, typecheck, focused Go and API generation/check
plus production client build, Go vet, cosmetic-boundary and no-payment checks pass, but this
Codex correction still needs Claude's designated review. RP-174 separately
records the ruling-author RFC §7.1 body conflict; no full Cosmetics or 1.0 promotion follows.

The post-commit ARM64 Compose Game UI target also passed against real Postgres/WebSocket:
both standard composed drivers and Cosmetics AC14's built-client Buy→reload journey. It
exercised a synthetic Cosmetics test epoch and did not run malformed-arm negatives through
the browser. The named test-only database service was stopped after the run; review and
production pinning remain separate.

## 2026-10-05 — Current-HEAD CI gate check after Cosmetic C5 correction

`make verify-ci-topology` passed its 13 negative controls at `7cc8463a`. A separate cold
`make verify-kernel-version` rerun still exits 2 at the historical pushed commit
`50a3a5141d5c4c1cd25cbf3c9c39d9e3cf8e2444`, reporting six minigame-client files
changed without a real kernel/VERSION bump. The kernel-history checkout contract and its
adversarial fixtures passed before that failure. This confirms the local Cosmetic correction
did not make the broader CI gate green; RP-131 and the owner/accepted-RFC route remain open.

## 2026-10-05 — Cosmetic C1 raw catalog number parity

Codex's designated C1 review found Go/TS disagreement on JSON integer spelling: TS accepted
catalog tier `1.0`/`1e0` and schema version `1.0` after `JSON.parse` normalized them; Go rejected
all three. New independently named shared negatives failed first on TS and passed Go. The
bounded raw TS loader now rejects those lexemes, with a targeted severing failure; the guarded
change bumps kernel 0.3.141→0.3.142 in the same commit. Cold client, focused Go/kernel,
typecheck, client build, cosmetic-boundary and no-payment gates pass. Claude must review the Codex
correction. This does not close C1, Cosmetics, the red historical kernel CI gate or 1.0.

Post-commit `make verify-kernel-version` reached the same historical `50a3a514` failure and
exited 2 before inspecting the new C1 commit. Direct parent/commit inspection confirms the
new guarded loader change and the 0.3.141→0.3.142 source/Go/TS mirrors landed together in
`76a9fe04`. This is a local version-bump fact, not a green full CI gate.

## 2026-10-05 — Cosmetic C2 replay-bundle designated review

Codex designated-approved Claude's C2 `8e315569^..8e315569` on current HEAD after cold Go
and client populations and restored severing of Go/TS pet-species dependency, Go/TS constants
identity, and permanent-ID settlement. The exact evidence and verdict live in the Cosmetic
planning log. It proves C2's bundle wiring in isolation, not a production Cosmetics epoch,
C1/C3–C8 review, archive eligibility, a green kernel-history CI gate or 1.0 readiness.

## 2026-10-05 — Cosmetic C3 save-corpus contract gap

Codex's targeted C3/AC4 review found RP-176: all five §6-mandated save-migration corpus
cases are absent (under both literal v21/v22 and OD-16-adjusted v23/v24 names), while the
11-case baseline still ends at v9. Cold Go/client tests pass because none asserts that
population. Direct v24 codec and Exit replay witnesses exist, so this is not a demonstrated
product activation failure. The named corpus harness lacks a pinned bundle and cannot
represent the RFC's new-run-only activation without a contract change. C3 stays CHANGES
REQUIRED pending ruling-author body reconciliation, then a firing current-coordinate test.

## 2026-10-05 — Cosmetic C4 AC9 database boundary correction

Codex's designated targeted review found RP-177: Postgres admitted the three Cosmetic event
kinds but allowed extra payload keys, contrary to accepted AC9. The retained rollback-only
real-Postgres test failed first for acquired, equipped and unequipped. Append-only migration
00084 now constrains their exact keys; all valid arms, extra/missing negatives and an unrelated
event-kind control pass. A temporary transactional DROP of the new constraint makes the extra
key witness fail, then rollback restores it. The first local CI server-core run found a stale
release-package migration pin (83), corrected to 84; its second cold run passed against real
Postgres. This is local implementation/first-filter evidence, not Claude's designated review,
hosted CI, full C4 approval or Cosmetics archival. C3/AC4's RP-176 ruling-author blocker
remains separate; the historical kernel-version CI gate remains red.

## 2026-10-05 — Cosmetic C4 Soul-recovery dispatch correction

Codex reproduced RP-178 through production's service option and real Postgres: active Soul
recovery rejected a valid Tier-1 cosmetic acquisition as `exclusive_activity`, contrary to
accepted §4.5. The retained test failed first. The bounded dispatch exemption now permits
acquisition, equip and unequip while ordinary gameplay stays blocked; the same recovery then
resolves and its Founder history verifies with cosmetic ownership preserved. Selectively
removing each exemption arm makes the workflow fail. A fixed-date test-clock failure was
disclosed and corrected by anchoring the fixture to the database clock. Kernel 0.3.143 lands
with the guarded routing change. The cold Postgres CI server-core target and client/type
checks pass locally. Claude must independently review this Codex correction before closure;
the original C4 range still carries its targeted CHANGES REQUIRED verdicts, and Cosmetics
and the broader 1.0 remain unfinished.

## 2026-10-05 — Cosmetic C4 exact review spans and remaining AC6 proof

The RP-177 SQL correction `fca686c5^..ce7688a7` and RP-178 recovery correction
`380d854b^..646daa08` are recorded for Claude's designated review. Both include their
tests, code, docs and records. RP-179 adds a distinct AC6 evidence gap: the shared corpus
contains 20 Founder cases but reaches only one pet; the two-wearer shape/projection tests
do not exercise a second equip transition in both runtimes. A test-only Go-authored corpus
supplement and selective Go/TS severing are predeclared. C4 and Cosmetics remain open.

## 2026-10-05 — Cosmetic C4 AC6 proof now reaches two wearers

RP-179's test-only correction adds a shared five-transition population: one acquisition,
two actual adoptions and two equips preserving both wearers. The missing required row failed
first; single-wearer mutations independently fail Go and TS. The regenerated corpus has
25 Founder rows plus the unchanged Exit pair. Focused cold Go, 6,950 client tests and
typecheck pass, with both mutations restored. No product or live content change was needed.
Claude's designated review is still required; this does not close C4, Cosmetics or 1.0.

## 2026-10-05 — Cosmetic C7 package-gate bypass corrected locally

RP-180's temporary actual dynamic economy import passed the old isolation gate. A bounded
tooling correction now parses syntax, normalizes package-local paths and scans nested script
files. Retained dynamic/path-escape negatives failed first; actual top-level and nested probes
now fail and were restored. The gate passes 22 negatives, 6,950 client tests and typecheck
pass locally. Existing client CI invokes the gate; no workflow changed and no fresh hosted
result is claimed. The independent AC7 contribution-producer mutation also fails the honest
population, but the Company comparison itself does not consume that provider and must not be
cited as integrated proof. Claude review and the broader Cosmetic review remain open.

The fresh `make verify-client` at `ad789790` passes typecheck/build/client/shell checks,
then fails at historical `50a3a514` in the kernel-history gate (RP-131, exit 2). The complete
client CI target remains red; separately executed Cosmetic/no-payment gates pass. No new
hosted result or historical exception is claimed.

## 2026-10-05 — Cosmetic AC8 proof covers actual owned/equipped Exits

RP-181 found that AC8's two named nonempty carry paths had no shared or persisted witness.
The retained required-row guard failed first. The test-only supplement adds wind-down and
accept-offer Exit pairs, with independent Company-output and Founder-audit comparisons.
Real Postgres creates ownership/equip through actual service commands, persists each Exit,
reloads/retries and verifies both history axes. The first invalid composition's missing
minigame activity resolver was disclosed and corrected with the real repository, not a
guard bypass. Resetting cosmetics in the actual Exit breaks both Go cases, both PG subcases
and both TS cases; all mutations are restored. Cold Postgres server-core, 6,953 client tests,
typecheck and the package/no-payment gates pass locally. No live loss bug or product change
was established. Claude review, RP-176 body reconciliation and broader release work remain.

## 2026-10-05 — Cosmetic AC7 asserts receipts and consumes real frozen bonuses

RP-182's receipt-only probe exposed a discarded production receipt; both Company arms also
used nil contributions. The test-only correction retains 200 × 288 steps, compares each
outcome/receipt/full state, freezes real non-unit bonuses once per run and exercises a separate
next-run purchase/accrual consumer. Disabling the receipt oracle and disconnecting that consumer
each fail their named retained negative. Two invalid first fixture assumptions are disclosed
in the RFC log; no failing setup was counted as discrimination. Cold full-policy and real-PG
server-core checks pass, as do 6,953 client tests and type/package/no-payment checks. No product
or content change and no live bonus bug established. Claude's designated review and the
broader release work remain; the historical client CI gate is still red (RP-131).

## 2026-10-05 — Cosmetic wearer controls now reference their visible disclosures

RP-183's retained all-controls assertion failed first on Equip in Chromium, Firefox and
WebKit. The bounded link-only correction uses the existing bound disclosure IDs on Equip
and Unequip, including pending states. Removing Unequip's link independently failed all
three equipped populations; the probe is restored. The full cold `make test-browser-ci`
passes 21,099 tests plus its separate performance population. Type/unit/boundary, production
build and append-only copy/content-manifest checks pass. No new copy or AT study is claimed.
Claude must designated-review `e5c63de5^..e02f6560`, as well as the preceding AC7 test-only
span `52bd6963^..7c16d890`. The fresh kernel-history gate still fails at `50a3a514` (RP-131);
its repair RFC remains draft. No hosted-green, Cosmetics closure, archive or push is claimed.

## 2026-10-05 — Existing live pet overlay gains persisted player-workflow proof

RP-184 reconciles a false absence claim: Garage already mounts the armor on the live pet,
but Cosmetics docs described that consumer as missing. The retained real-server driver now
continues Buy/reload through actual DOM adoption/equip, the live annoyed/no-text pet panel,
worn reload, actual browser reduced-motion preference and unequip/unworn reload. Every new
intent is one exact Founder-scoped POST and pet identity comes from real adoption, not a seed.
The consumer-severing and combined motion-defense probes each fail their named outcome and
are restored. Full composed CI passes both drivers, including Fiscal/Pitch; full cold
three-engine browser CI passes 21,099 tests plus the separate performance case. Unit/type/
boundaries also pass. No product, catalog, owner copy or workflow changed. The fixture epoch
is not a production mint, Claude review is pending, GS4×PA7 remains unresolved, and G10/1.0
release gates remain open. The historical kernel-history gate remains red (RP-131).

## 2026-10-05 — Actual motion preference exposes CSS defect; full browser CI reveals two more failures

RP-185's original prop-only AC15 case survives a severed media rule in all three engines.
The retained actual-preference test fails first on unchanged component CSS (RP-186): the
static selector loses to the animation selector. The bounded precedence correction passes
12/12 shelf cases; media-rule and caption probes independently fail all three and are restored.
Typecheck, 6,953 unit tests, production build and boundaries pass, as do both built-client
real-Postgres/WebSocket composed journeys. No owner copy, catalog, kernel or CI topology change.

The full cold browser target is red, not green: 21,097 passed, 2 failed, 3 skipped. Chromium
Typer's update-depth failure and WebKit Snake's fixed-delay batching failure are separately
tracked as RP-187/RP-188 for deterministic diagnosis in their accepted lanes. The separate
performance case was not reached. Fresh kernel-history verification still fails at pushed
`50a3a514` (RP-131). Claude review of the Cosmetic correction and wider product/rights/content/
deployment gates remain; no archive, release claim or push.

## 2026-10-05 — Typer display loop reproduced and corrected; Snake still needs diagnosis

RP-187's increasing injected clock reproduces the cold update-depth error on unchanged Typer
in all three engines. A display-only local-sample correction removes the self-read dependency
without changing time authority/payout/content. All 24 child cases pass; reinstating only the
reactive read-back fails all three, then is restored. Unit/type/build/boundaries pass. A cold
full browser CI run now passes 21,102 tests plus the separate performance case.

This is not a retry-only fix or complete CI green. The first red run remains on record, and
Snake passed unchanged, so RP-188's fixed-delay/callback assumption is the next accepted-lane
diagnosis. RP-131's historical kernel-history failure, cross-party correction reviews, public
wire/content/AT, rights and deployment/release obligations remain open. No push or archive.

## 2026-10-05 — Arcade fixed-delay oracle corrected without changing gameplay

RP-188's real-child/TS-engine population demonstrates that 190 ms of wall-clock movement alone
delivers no ticks. Controlled callbacks certify the exact fourth-tick batch, blur stopping both
movement and dispatch, rejected tick-8 recovery to server tick 4, and the unacknowledged lead
limit. The native-timer D-pad/terminal path remains. Batch/blur/resync/lead mutations independently
fail all three engines; probes are restored with no production diff. No higher sleep, retry,
weaker command, content/copy/kernel or CI change.

Root type/unit/build/boundaries pass (6,953 unit tests); full cold Linux browser CI passes 252
populations / 21,102 tests with three intentional performance skips, then its separate performance
case. The prior red run stays on record. This establishes an invalid timing oracle and a test-only
correction, not a sole-cause or hosted reliability verdict. Claude review is still required.
Next: bounded review of Arcade A6's native keyboard completion/quit evidence. RP-131's draft-only
kernel-history repair, broader review/ruling/content/rights/clean-host release gates remain open.

## 2026-10-05 — Native-key Arcade completion and quit now execute against real engines

RP-189 supplements the missing A6 keyboard proof without changing product bytes. Native Enter,
arrows and Space choose and clear the existing seed-202 Mine Grid scenario in one mounted
component; the full terminal equals the Go corpus and hidden mine positions remain absent until
terminal. Native Enter on Snake Quit yields its real engine terminal, revision and status.
The old 12-case suite survives a severed Quit; the new case fails all three engines. Separately
disconnecting Mine Grid's primary action fails its named case in all three. Probes restored.

Root type/unit/build/boundaries pass (6,953 unit tests, 88 browser-only skips); full cold Linux
browser target passes 21,108 tests plus the separate performance case. Claude designated review
remains required, and this is controlled-focus component/engine evidence, not full Tab/AT or
public Arcade acceptance. RP-190 is newly recorded: 43 declared transitions vs 60 attempted
commands; Go/TS bounded AR7 budget audit/correction is next. Complete CI remains blocked on
RP-131; no blocked wire/mint/content, archival, public claim or push is authorized.

## 2026-10-05 — Arcade attempted-command denominator corrected and reproducible

RP-190's independent Go/TS checks fail first on the 43-applied-only budget. The generator and
execution consumer now count all 60 attempts, including 17 rejections. Regeneration changes
only `transition_budget`; structural comparison keeps all scenarios/commands/results/content
identical and a second generation is byte-identical. Go's old producer and TS's old execution
counter independently fail their respective gates; both probes are restored.

Cold Go Arcade (`-count=1`), root type/unit/build/boundaries/vet pass (6,954 client tests). Full
cold Linux browser CI passes 252 populations / 21,111 tests, then its separate performance case.
This is test/fixture metadata, not gameplay/content/balance/kernel change, and still requires
Claude cross-party review. Next is RP-191's bounded TS hidden-state witness: the named AC4 test
currently reads terminal metadata only, although actual Go/browser positive controls exist.
Complete CI's RP-131 history gate, broader reviews/rulings/rights/content and clean-host release
proof remain open. No archival, deployment, public release claim or push.
