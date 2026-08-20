# Repository inventory at product coordinate `190a4fa`

State: **Wave 1 structurally complete; row-level semantic depth remains in Waves 2–3.** This records
verified populations and reconciliation defects; it is not yet the completed per-capability trace.

## Counted populations

| Population | Count | What the count means | Reconciliation state |
|---|---:|---|---|
| Top-level design documents | 14 | `design/00` through `design/13` | Read in full; 121 preliminary stable outcome rows recorded, with Wave-2 splitting still required. |
| Tracked design research files | 1 | Only `provenance-extracts.md` survives the ignore policy | 33 distinct research dossiers are referenced from tracked design but absent from a fresh clone. |
| Active-directory RFC Markdown files | 24 | Process + index + template + 21 product/process RFC files | All 111 true acceptance rows are extracted and evidence-reconciled; five remain intentionally open for exact review/provenance. |
| Archived RFCs | 46 | Files in `rfc/archive/` | All 46 pair structurally; the predeclared 20-row/ten-domain deep replay is complete with 26 rows honestly structural-only. |
| Live top-level planning directories | 23 | Includes RFC plans and non-RFC maintenance/review threads | Every thread is mapped; four explicitly complete/historical threads remain live pending authorized closeout. |
| Tracked planning files after this checkpoint | 213 | Includes the alignment control plane and five tracked top-level records | Fresh-clone population is exact after the copy-consumption predeclaration lands. |
| Ignored local planning files | 25 | Seventeen coverage-map files, seven diagnostics, and one historical Codex fix record | Local-only records are not shared memory and receive no fresh-clone evidence credit. |
| Canonical docs | 38 | Includes three generated artifacts | Every file is classified against lifecycle/release evidence; only the numeric foundation is currently unqualified. |
| Go source files | 340 | All `server/**/*.go` | 189 production plus 151 test files; both exact ledgers reconcile. |
| Server package directories | 45 | Second-level directories under `server/` containing Go | All are mapped in `server-package-inventory.tsv`; Combat is uncomposed and several composed backends have no player consumer. |
| SQL migrations | 74 | Append-only files `00001` through `00074` | Sequence contiguous; each has one Up/Down marker and the cold migration lane passes. Per-migration semantic rollback remains unisolated. |
| Client source files | 82 | `client/src/**` | All rows mapped: 41 in the shipped entry/type graph and 41 outside; the build emits 38 authored runtime/style sources plus three type-erased contracts. |
| Tracked client test sources | 43 | `client/test/**/*.ts` | Two browser suites, one browser helper, one type fixture, and 39 unit/parity/candidate/measurement sources; every row names its production relationship. |
| Ignored local browser captures | 13 | `client/test/__screenshots__/**/*.png` | Generated local files, not tracked baselines or asserted evidence; absent from a fresh clone. |
| Game UI registered surfaces | 5 | Desk, Offer Sheet, Run End, Settings, Vision Slide | Real bootstrap reaches Desk; full workflow remains unproved. |
| Balance files | 91 | 19 epoch artifacts, four platform configs, 15 schemas, 53 fixtures | Every filename has a boundary/consumer row; live families have loader and server/client traces. |
| Copy files | 10 | Four source catalogs plus six policy/reference/report files | All four catalogs ship despite three `candidate` names; the 161-key orphan report includes live UI calls. |
| Make targets declared phony | 72 | Development, generation, verification, harness, browser, integration | Exact classification complete: only 26 currently have bounded-valid evidence; raw target count is not check count. |
| GitHub CI jobs | 7 | Six push/PR jobs plus scheduled numeric maintenance | Current-head workflow never completes because the harness lane is cancelled at 30 minutes. |
| Exposed HTTP/WebSocket operations | 24 | Three gameserver-hand, 11 account-hand, ten private-registry operations | Only bootstrap, snapshot, intents, and WebSocket have production browser consumers; see `route-operation-inventory.tsv`. |
| Server commands | 5 | Gameserver, balance harness, API generator, content-manifest generator, formula generator | Only gameserver is a deployable runtime; no production package exists. |
| Production deployment files | 1 | Content identity manifest only | Five Compose files are test harnesses; no production Docker/Caddy/Compose/runbook, metrics exporter, backup/restore tool, or sunset artifact. |
| Explicit deployed goroutine instances | 11 | Repository-authored long-lived launches in the gameserver path | HTTP listener, context monitor, relay, six jobs, job waiter, and transport world coalescer are exact; dependency-internal goroutines are excluded. |
| Attached background jobs | 6 | Jobs composed by `gameserver.Compose` and primed before readiness | World aggregation, replay verification, guild presence, guild clearing, disband sweep, credential GC. |
| Closed player event kinds | 48 | `save.AllEventKinds` persisted/replayed event authority | Backend boundary exact; production Game UI decodes six lifecycle wire kinds. |
| Closed transport envelope kinds | 5 | Receipt, snapshot, event, presence, system | Producer/channel rules exact; consumer payload subsets remain explicit. |

## Initial boundary inventory

### Server producers

- `server-package-inventory.tsv` reconciles all 45 second-level package directories: 189 production
  files plus 151 tests equal the counted 340 Go files. Production-import consumers are named rather
  than inferred from directory names. `combat` has no production importer; `harness` and the
  grouped `internal` packages are tooling lanes, not player runtime composition.
- The account router declares 11 hand-wired account/session/founder/intent operations and mounts
  ten registry-owned Game UI, minigame, and Soul operations.
- The composed service exposes `/healthz`, `/readyz`, and `/connection/websocket`.
- Persistence spans 74 migrations and systems for accounts, saves, events, routes, Commons,
  guilds, leaderboards, transport, minigames, pets, Soul, attendance, and Exits.
- The production minigame tenant registry composes exactly `pitch.NewTenant()` at
  `server/gameserver/composition.go`; no combat tenant is registered.

### Persistence boundary

- `migration-inventory.tsv` enumerates the contiguous 00001–00074 population and assigns each row
  a domain owner. Every file has exactly one Goose Up and one Goose Down marker.
- `make validate-migrations` ran cold on 2026-08-21 and passed the Postgres integration/migration
  population; `make verify-schema` passed all current catalogs and schemas.
- Those aggregates prove current-chain execution and structural Down presence. They do **not**
  prove that each individual Down restores all prior data/constraints, so semantic rollback remains
  explicitly open rather than being promoted from marker presence.

### Client consumers

- `client-source-inventory.tsv` accounts for all 82 files exactly. Starting at `main.ts` and
  following static imports plus the Worker URL yields 41 source dependencies; 41 files remain
  outside that graph. Production source maps identify 37 authored JS/data/Svelte sources plus one
  CSS asset; three additional graph rows are compile-only contracts.
- The outside half is not dead code by definition: many rows own cross-runtime parity, replay,
  validators, or test fixtures. It is nevertheless not a mounted player consumer. Achievements,
  active play, Combat, Commons, curriculum, doctrines, faction, fiscal, guild, meters, minigames,
  pets, Pitch, routes, Soul, and the full replay engine receive no player-capability credit from
  their TypeScript module presence.
- `client-workflow-inventory.tsv` traces 25 bounded default, failure, and accessibility paths. The
  distribution is three proven-bounded, nine partial, one mechanical-only, eight backend-only, two
  absent, and two failed.
- The only production Svelte game surfaces are implemented by `GameUIApp.svelte` and
  `RunEndSurface.svelte`; UI Foundation has `Amount.svelte` and the test-only `FixtureHost.svelte`.
- `game-ui/runtime.ts` owns account bootstrap, HTTP intents/snapshots, WebSocket frames, and browser
  credential storage.
- Settings currently renders save status and an account note only. It has no recovery, import,
  export, deletion, email, or session-management controls.
- TypeScript modules exist for many server-side mechanics, but module parity is not a player
  consumer. Each still requires a mounted workflow trace.
- `make build-client` transformed 152 total dependency modules and emitted the main JS/CSS plus the
  prediction Worker. `make typecheck verify-client-boundary` passed with zero TypeScript/Svelte
  diagnostics and the declared 14 shell / eight UI / two Game UI component boundaries green. Those
  are valid build/structure witnesses, not proof of the 22 non-proven workflows.

### Executable evidence boundary

- `make-target-inventory.tsv` exactly equals the 72-name `.PHONY` population. It distinguishes 26
  bounded-valid targets from ten mutators, nine partial targets, five release-invalid aggregates,
  and setup/manual/historical/alias/parameterized/pending lanes. A generator or aggregate name is
  never counted as a passing check by existence.
- `ci-job-inventory.tsv` traces all seven jobs. In hosted run `32404232364`, server, client,
  browser, composed Game UI, and schema passed; harness was cancelled after 30m02s; numeric was
  skipped on the push. Scheduled/manual triggers run the ordinary blocking jobs too, and historical
  numeric failure fails the whole workflow.
- `client-test-artifact-inventory.tsv` exactly equals the current 56-file local test tree, but only
  43 TypeScript sources are tracked. The 13 PNGs under ignored `__screenshots__` are generated local
  captures with no screenshot assertion and no fresh-clone availability; they are not acceptance
  witnesses.
- `verify-routes-boundary` and `verify-commons-boundary` override the mandated repository-local Go
  cache with task-named `/tmp` caches. The checks may still detect their import conditions, but the
  recipes contradict the repository's current routine-command/cache protocol.
- `server-test-file-inventory.tsv` exactly equals all 151 `*_test.go` files. Structurally they hold
  591 top-level `Test*` functions and one fuzz target: 25 filename-declared integration files, 124
  unit/package files, one corpus/fuzz file, and one fixture-helper file with no test function.
  Fifteen files directly show database/HTTP/WebSocket composition signals. These are source facts,
  not semantic oracle verdicts.
- `server-test-skip-inventory.tsv` records 40 explicit skip sites in 28 files. Thirty-nine sites in
  27 files skip when `TEST_DATABASE_URL` is absent; the remaining site is an architecture-size
  guard. Ordinary host `make test-go` can therefore exit green without the Postgres population,
  and Go's non-verbose output does not make that denominator visible. Docker/hosted Postgres lanes
  are required for any integration claim.

### Runtime concurrency, jobs, and events

- `runtime-concurrency-inventory.tsv` accounts for the deployed gameserver's 11 explicit long-lived
  goroutine instances: the HTTP listener, root-context monitor, player outbox relay, six attached
  background jobs, their waiter, and the transport world coalescer. Centrifuge and `net/http`
  internals are dependency-owned and deliberately excluded from the exact source count.
- Every attached job runs one synchronous prime pass before readiness, then gets its own goroutine.
  A job error cancels all jobs, lowers readiness, and reports to the main runtime-failure channel.
  The six jobs are world aggregation, replay verification, guild presence relay, deterministic
  guild clearing, guild disband sweep, and session/bootstrap credential GC.
- Binding `design/06` is not runtime truth: no player actor or Matchmaker launch/type exists, and
  world aggregation plus transport coalescing are two goroutines rather than one. Match actor and
  matchmaking lifecycle are explicitly deferred by the Minigame Platform RFC. Filed RP-103 for
  author reconciliation and accepted successor authority; transactional handlers are not silently
  relabeled actors.
- `event-family-inventory.tsv` maps the closed 48-kind player event registry, two player-outbox
  kinds, five transport envelopes, three system codes, five disconnect codes, six Game UI lifecycle
  wire kinds, seven guild-domain event kinds, two guild-presence kinds, six replay verdicts, five
  projection-idempotency ledgers, and three runtime invariant kinds. Backend registry breadth does
  not imply Game UI consumption: the shipped lifecycle decoder understands six event kinds.
- Two cleanup obligations have no runtime lane. The already-recorded 30-day intent-record pruner is
  absent (RP-086), and Route Registry's 72-hour naming expiration method has neither production nor
  test caller, cadence, batching, failure owner, or metrics. Filed RP-104; the audit does not invent
  a scheduler contract.

### Declarative data

- `balance-file-inventory.tsv` exactly equals the 91-file tree: 19 deploy-current epoch artifacts,
  four platform configs, 15 schemas, and 53 positive/negative/historical/candidate/measurement
  fixtures. A fixture is never counted as deploy-current content merely because a harness consumes
  it.
- `catalog-family-inventory.tsv` traces 23 live/platform families. The current epoch has 19
  artifacts and resolves to epoch 8 hash
  `sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6`;
  `make epoch-hash` reproduces the deployment manifest value. Client Shell, Transport, and the
  epoch seed are additional composed platform inputs. API policy is test-only and uncomposed.
- Live data exists for economy, routes, categories, epochs, Commons, factions, guilds, prestige,
  transport, achievements, doctrines, fiscal, meters, minigames, opportunities, pets, Pitch,
  Soul, curriculum, API policy, and client shell.
- No `balance/combat/` catalog exists even though Combat Shared Data AC1 requires a complete
  fixture and the Minigame Platform RFC top/body claim a combat duel tenant.
- `balance/minigames/first-content.json` contains only The Pitch.

### Copy boundary

- `copy-file-inventory.tsv` exactly equals the ten-file copy tree. The generator reads every JSON
  file in `copy/catalog/`, so `achievements-candidate.json`, `game-ui-candidate.json`, and
  `permits-candidate.json` are shipped sources, not held candidates. Their 13 + 143 + 1 rows join
  the 51-row `phase0.json` catalog in the 208-key client/server artifact.
- `make copy-check` is cold-green for drift, copy hash
  `sha256:9e816294f84e6c50e5050f59d33f49a1f09d42780abdb7f5b40bb5a5442c0e13`,
  and deployment-manifest identity. It emits 161 orphan warnings.
- The orphan output is not a consumption oracle: it marks direct production calls such as
  `chrome.run_title.company_fallback`, `desk.buy_one`, and `screen.run_end.founder_note` orphaned.
  The pipeline enumerates selected artifact references and two explicit Go sites, but not client
  `t()` call sites. True unused copy and live UI copy are therefore mixed in one green report.

### Planning and canonical-record boundary

- `planning-thread-inventory.tsv` exactly equals the 23 top-level planning directories and records
  tracked/local file counts, authority, current state, and closeout gap for each. Twelve are active
  RFC planning threads, one is a blocked draft child, and the others are archive, provenance,
  maintenance, historical measurement/review, or the alignment control plane.
- Four threads explicitly describe their work as complete, withdrawn-history-only, or superseded
  while remaining live: `archived-four-review`, `harness-dispatch-cardinality`,
  `production-review-round2`, and `run-genesis-archival-remediation`. The last already requests a
  future move to `planning/archive/`. The audit records that lifecycle defect but does not infer an
  archival move or rewrite review provenance.
- At the current copy-consumption-plan checkpoint, planning contains 213 tracked
  files. Another 25 files are ignored and local-only: 17 coverage-map records, six archived T0-T1
  diagnostics, one platform-alignment diagnostic, and one historical Codex fix record. Their local
  presence cannot support a fresh-clone/shared-memory claim.
- `docs-file-inventory.tsv` exactly equals all 38 files under `docs/`. Three generated artifacts
  are drift-checked, the numeric contract is a proven foundation, and every other current-state,
  backend-only, partial, stale, or contradicted claim is tied to an existing RP repair/authority
  route. Classifying a doc does not silently repair author-owned text or promote backend mechanics
  to a player workflow.

## Confirmed reconciliation defect: Minigame Platform

The RFC header says “normative body reconciled,” its dependency line says the Combat Shared Kernel
and duel engine tenant are implemented, MP1 says two combat modes are already replay-proven, MP5
names duel as tenant #1, and AC6 requires its adapter. Later C9 blocker/ruling text explicitly says
the combat children are not implemented and moves duel/lane registration to their child RFCs.

At `HEAD`, Combat Shared Data remains implementing, the duel RFC is draft and records the missing
catalog, `balance/combat/` is absent, and the gameserver tenant registry contains only The Pitch.
This is a normative body/ruling contradiction, not merely a stale plan. Per `AGENTS.md`, the ruling
author must reconcile it; implementation and archival cannot infer the intended contract.

## Confirmed design-body contradictions

- **Account default/fallback:** `design/11 §1b` already chooses silent server-anonymous as the
  default and local-only play as the labeled outage fallback. `design/06` and Account D4 retain a
  broader fully-offline formulation. This is author reconciliation plus missing fallback
  implementation, not a new owner choice.
- **Soul recovery:** `design/03 §5` says rewarding hobbies never restore Soul and cites missing
  §5c as the sole zero-reward recovery source. The Go paragraph later calls hobbies restorative,
  and §8 explicitly grants Arcade Soul restoration.
- **Cold-open record:** `design/11 §1` displays a WR in the specimen; §1b's later owner ruling says
  no WR line may appear before validated boards exist.
- **Minigame architecture:** `design/12 §3` specifies a ticking `MatchActor`; the accepted platform
  RFC/runtime use pure tenants and DB-authoritative session state.
- **Production hot reload:** `design/12 §5` says “pack push + hot reload” while the immediately
  following adopted ruling says hot reload is dev-only and production requires an epoch-stamped
  deploy.

These are recorded rather than edited because the relevant text is owner-authored or an accepted
RFC body. The body-reconciliation rule assigns the correction to its author.

## Remaining semantic-depth work

- Split the 121 section-level design IDs until each names one independently falsifiable workflow.
- Extend the completed file/family map into row-level gameplay-content and all 208 copy-key call
  sites; replace the non-discriminating orphan report before using it for cleanup/release claims.
- Complete semantic row-level fixture/oracle/negative-control maps now that every executable file,
  target, job, and skip population is structurally bounded.
- Reconcile the four routed historical/complete live planning threads transactionally after their
  exact review/provenance dependencies are satisfied. The 46-row archive structure and 20-row
  deep replay are complete; the remaining 26 archives retain structural-only status rather than
  inferred semantic approval.
- Continue migration semantic rollback discrimination beyond the now-proven contiguous/marker/cold-
  chain baseline; account ownership/export/deletion/retention and backup/restore implications now
  have release dossiers and exact blocked routes.

The first section-level ID pass is recorded in `design-capability-ledger.tsv`. Rows deliberately
remain coarse until Wave 2 splits each outcome to one independently falsifiable workflow.
