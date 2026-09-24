# Deployment Foundation: Claude corrective range, designated-review handoff

**Status:** review input for **Codex** (the cross-party designated reviewer for Claude-authored
implementation under AGENTS.md (c)). This is not a verdict, and it does not make any archival or
release claim. **Prepared:** 2026-09-24.

On 2026-09-24 the owner directed Claude to implement the accepted-RFC corrections itself. Claude
recorded the designated verdicts on DP-A–DP-E and then implemented the fixes. Claude has not
reviewed and must not approve its own corrective work; only Codex's verdict counts.

## Exact range

| Range | Content |
|---|---|
| `67fd415..HEAD` at handoff update (`b706e4d`) | R1–R17 corrective implementation plus predeclaration and record commits, and the verification/DESIGN-GAP record. Each batch is logged in `log.md` from "DP-D corrective R1 predeclaration" onwards. |

Separately reviewable Claude research and planning commits from earlier the same day (no product
code):

- `8a02a3c` and `20991e7`: RP-122 probe and result;
- `08a8669` and `c08b592`: designated verdicts on DP-A–DP-E and the corrective range;
- `67fd415`: the DP-F advisory inspection.

## What each batch claims (details, evidence and severing probes are in `log.md`)

| Batch | Finding closed | Main files |
|---|---|---|
| R1 | Rollback validated its inputs only after destroying the database. Release ran preflight before loading the image. | `deploymentrelease/controller.go`, `docker.go` |
| R2 | Release, rollback and recovery never applied the previous-key rotation overlay. | `deploymentrelease/ledger.go`, `command.go`, `docker.go` |
| R3 | The image secret scan never opened compressed layers. | `releasepackage/secrets.go`, `deploymentrehearsal/probe.go` |
| R4 | Manifest image digests and epoch/constants/copy identity were not bound to Compose or content. | `releasepackage/manifest.go` |
| R5 | The cleanup alert could never fire. Delivery counted attempts, not successes. The Caddy admin API was open to peer containers. | `operations/*`, `deployment/Caddyfile`, alert rules, profile lane |
| R6 | The backup worker restarted endlessly on any foreign entry, taking a backup each time. Checksum checks had no witnesses. | `deploymentbackup/target.go`, `cmd/deployment-backup` |
| R7 | The production proxy-hop and origin wiring had no witness. The decoder had four unwitnessed rules. | `gameserver` integration test, `deploymentconfig` tests |
| R8 | Shipped transitive client packages lacked attribution. 0BSD support added (a licence-policy expansion). | `cmd/gen-release-metadata`, `releasepackage/metadata.go` |
| R9 | Drain and identity observers were unwitnessed. Readiness-down accepted any non-204. | `deploymentrelease/docker.go`, `smoke.go` |
| R10 | Rollback authority trusted the container's backup report instead of the host bytes. | `deploymentrelease/docker.go` |
| R11 | Alert tests did not discriminate thresholds, clauses, bounds or resolution. Stale host observations were served silently. | alert rules and promtool fixtures |
| R12 | Plan rows accepted constant commands. The negative exit code was ambiguous. | `deploymentrehearsal/plan.go`, CLI |
| R13 | Recovery smoke needed Alertmanager, which recovery did not start. A failed install could delete pre-existing volumes. | `deploymentrelease/docker.go` |
| R14 | Build records were bound only by manifest hash and source commit. | `deploymentrehearsal/supply_chain.go` |
| R15 | Bundle-mutation probes could pass on byte integrity alone. | `deploymentrehearsal/probe.go` |
| R16 | The committed backup envelope was never re-verified, and the rename was not synced. | `deploymentbackup/backup.go` |
| R17 | The browser-manifest binding was unwitnessed. | `deploymentrehearsal/run_test.go` |
| C2 | Consumed-offset replay suppression was unwitnessed. | `client/test/game-ui-runtime.test.ts`, `docs/transport.md` |
| R18 | The wrong-epoch, irreversible-migration and missing-rollback-input negatives had no producer. | `deploymentrehearsal/probe_release.go` |
| R19 | Lifecycle release and rollback had no producers. A successful rollback row cleared the fields the evidence validator requires, so real evidence could never validate. | `deploymentrehearsal/lifecycle.go`, `deploymentrelease/controller.go` |
| R20 | The non-clean restore negative had no producer, and its refusal was not attributable. | `deploymentbackup/postgres.go`, backup CLI, `deploymentrehearsal/nonclean.go` |
| R21 | Host clean-start checked only the Postgres volume. | `deploymentrehearsal/host.go` |
| R22 | The restart-during-admitted-work negative had no producer. | `deploymentrelease/docker.go`, `deploymentrehearsal/restart.go` |

Also in range, outside Deployment:
- `266b8d8`: API Foundation A6/AC4 body reconciliation (RFC text);
- `a965538`: Prestige plan item 6, the epoch-8 first-elective evidence, with its lower-bound caveat.

## Process deviations to examine (self-disclosed)

- R2, R7, R11, R15–R22 and C2 have no separate predeclaration commit. R2's predeclaration landed in
  its implementation commit.
- R4 committed a test file unformatted; it was fixed in R8.
- R13 changed a Codex test assertion: Alertmanager moved from the excluded list to the required
  list.
- R8 expands the licence allowlist by `0BSD`.
- R10 records an interpretation of "irreversible migration" for review.
- Two severing observations were corrected in the log after first being misread:
  - R13: a build failure hidden by an output filter;
  - R12: the real `removed_catalog → 3` against a pre-R5 bundle was vacuous.

## Cold gates run on this range

The last results for each gate:
- `make verify-push` (after R14);
- `make test-deployment-backup`, `-release` and `-rehearsal` at `ec5518b`;
- `make test-deployment-operations` after R11;
- promtool, with 13 mutations killed.

## Known consequences and open items

- **Every retained bundle (v1–v5 candidates and the previous bundle) is now invalid under the
  corrected rules.** R-006 needs freshly built candidate and previous bundles from reviewed
  source.
- Six DESIGN-GAPs are recorded in `log.md` for the RFC author: RPO semantics, per-family alert
  attribution, the `UpgradeResolved` lifecycle, overlay granularity, the previous-bundle security
  floor, and asserted host fields.
- No clean-host run has happened, and there is no release or archival claim.
