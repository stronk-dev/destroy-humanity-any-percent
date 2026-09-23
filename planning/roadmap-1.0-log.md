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
