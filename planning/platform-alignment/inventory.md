# Repository inventory at product coordinate `190a4fa`

State: **Wave 1 in progress.** This records verified populations and reconciliation defects; it is
not yet the completed per-capability trace.

## Counted populations

| Population | Count | What the count means | Reconciliation state |
|---|---:|---|---|
| Top-level design documents | 14 | `design/00` through `design/13` | Read in full; 121 preliminary stable outcome rows recorded, with Wave-2 splitting still required. |
| Tracked design research files | 1 | Only `provenance-extracts.md` survives the ignore policy | 33 distinct research dossiers are referenced from tracked design but absent from a fresh clone. |
| Active-directory RFC Markdown files | 24 | Process + index + template + 21 product/process RFC files | All 111 true acceptance rows are extracted and evidence-reconciled; five remain intentionally open for exact review/provenance. |
| Archived RFCs | 46 | Files in `rfc/archive/` | File/index bidirectionality previously checked; risk-ranked acceptance re-execution pending. |
| Live top-level planning directories | 23 | Includes RFC plans and non-RFC maintenance/review threads | Several have no declared closeout location; exact mapping pending. |
| Planning files total | 190 | Includes ignored/internal coverage material visible in this checkout | Fresh-clone and generation provenance differ from local visibility. |
| Canonical docs | 38 | Includes three generated artifacts | Claims/index inventory begun; code/runtime agreement pending per system. |
| Go source files | 340 | All `server/**/*.go` | 151 are tests; package/producer ownership inventory begun. |
| Server package directories | 45 | Second-level directories under `server/` containing Go | All are mapped in `server-package-inventory.tsv`; Combat is uncomposed and several composed backends have no player consumer. |
| SQL migrations | 74 | Append-only files `00001` through `00074` | Sequence contiguous; each has one Up/Down marker and the cold migration lane passes. Per-migration semantic rollback remains unisolated. |
| Client source files | 82 | `client/src/**` | All rows mapped: 41 in the shipped entry/type graph and 41 outside; the build emits 38 authored runtime/style sources plus three type-erased contracts. |
| Client test artifacts | 56 | `client/test/**` | 43 TS and 13 screenshots; discrimination audit pending. |
| Game UI registered surfaces | 5 | Desk, Offer Sheet, Run End, Settings, Vision Slide | Real bootstrap reaches Desk; full workflow remains unproved. |
| Balance files | 91 | Live catalogs, schemas, candidates, and test fixtures | Catalog-family producer/consumer/epoch map pending. |
| Copy files | 10 | Live/candidate catalogs, provenance, references, generated reports | Shipped/candidate boundary and real surface consumption pending. |
| Make targets declared phony | 72 | Development, generation, verification, harness, browser, integration | Valid-objective and negative-control audit pending. |
| GitHub CI jobs | 7 | Six push/PR jobs plus scheduled numeric maintenance | Current-head workflow never completes because the harness lane is cancelled at 30 minutes. |
| Exposed HTTP/WebSocket operations | 24 | Three gameserver-hand, 11 account-hand, ten private-registry operations | Only bootstrap, snapshot, intents, and WebSocket have production browser consumers; see `route-operation-inventory.tsv`. |
| Server commands | 5 | Gameserver, balance harness, API generator, content-manifest generator, formula generator | Only gameserver is a deployable runtime; no production package exists. |
| Production deployment files | 1 | Content identity manifest only | Five Compose files are test harnesses; no production Docker/Caddy/Compose/runbook, metrics exporter, backup/restore tool, or sunset artifact. |

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

### Declarative data

- Live data exists for economy, routes, categories, epochs, Commons, factions, guilds, prestige,
  transport, achievements, doctrines, fiscal, meters, minigames, opportunities, pets, Pitch,
  Soul, curriculum, API policy, and client shell.
- No `balance/combat/` catalog exists even though Combat Shared Data AC1 requires a complete
  fixture and the Minigame Platform RFC top/body claim a combat duel tenant.
- `balance/minigames/first-content.json` contains only The Pitch.

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

## Remaining Wave 1 work

- Split the 121 section-level design IDs until each names one independently falsifiable workflow.
- Extend the completed 111-row active acceptance audit into a risk-ranked archived-RFC sample;
  preserve the five exact review/provenance rows as open rather than unaudited.
- Extend the completed 24-operation route/consumer map into actor, worker, event, and background-job
  ownership; retain the 17 backend-only and one unimplemented route verdicts.
- Map all catalog rows and copy keys through loaders, epoch identity, and real workflows.
- Inventory all 56 client test artifacts and every remaining test/gate population, oracle, negative
  control, timeout, exclusion, and artifact.
- Reconcile all 23 live planning directories and 46 archived RFC ranges bidirectionally.
- Continue migration semantic rollback discrimination beyond the now-proven contiguous/marker/cold-
  chain baseline; account ownership/export/deletion/retention and backup/restore implications now
  have release dossiers and exact blocked routes.

The first section-level ID pass is recorded in `design-capability-ledger.tsv`. Rows deliberately
remain coarse until Wave 2 splits each outcome to one independently falsifiable workflow.
