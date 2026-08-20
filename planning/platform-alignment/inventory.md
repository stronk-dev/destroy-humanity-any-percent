# Repository inventory at product coordinate `190a4fa`

State: **Wave 1 in progress.** This records verified populations and reconciliation defects; it is
not yet the completed per-capability trace.

## Counted populations

| Population | Count | What the count means | Reconciliation state |
|---|---:|---|---|
| Top-level design documents | 14 | `design/00` through `design/13` | Read in full; 121 preliminary stable outcome rows recorded, with Wave-2 splitting still required. |
| Tracked design research files | 1 | Only `provenance-extracts.md` survives the ignore policy | 33 distinct research dossiers are referenced from tracked design but absent from a fresh clone. |
| Active-directory RFC Markdown files | 24 | Process + index + template + 21 product/process RFC files | Initial lifecycle table exists; 111 true acceptance rows extracted after excluding a nested hardening list that a naive heading parser miscounted. Evidence reconciliation pending. |
| Archived RFCs | 46 | Files in `rfc/archive/` | File/index bidirectionality previously checked; risk-ranked acceptance re-execution pending. |
| Live top-level planning directories | 23 | Includes RFC plans and non-RFC maintenance/review threads | Several have no declared closeout location; exact mapping pending. |
| Planning files total | 190 | Includes ignored/internal coverage material visible in this checkout | Fresh-clone and generation provenance differ from local visibility. |
| Canonical docs | 38 | Includes three generated artifacts | Claims/index inventory begun; code/runtime agreement pending per system. |
| Go source files | 340 | All `server/**/*.go` | 151 are tests; package/producer ownership inventory begun. |
| Server package directories | 44 | Second-level directories under `server/` containing Go | Several are primitives with no player consumer. |
| SQL migrations | 74 | Append-only files `00001` through `00074` | Sequence contiguous; Up/Down and behavior ownership audit pending. |
| Client source files | 82 | `client/src/**` | 72 TS, five Svelte, three JSON, one CSS, one text declaration asset. |
| Client test artifacts | 56 | `client/test/**` | 43 TS and 13 screenshots; discrimination audit pending. |
| Game UI registered surfaces | 5 | Desk, Offer Sheet, Run End, Settings, Vision Slide | Real bootstrap reaches Desk; full workflow remains unproved. |
| Balance files | 91 | Live catalogs, schemas, candidates, and test fixtures | Catalog-family producer/consumer/epoch map pending. |
| Copy files | 10 | Live/candidate catalogs, provenance, references, generated reports | Shipped/candidate boundary and real surface consumption pending. |
| Make targets declared phony | 72 | Development, generation, verification, harness, browser, integration | Valid-objective and negative-control audit pending. |
| GitHub CI jobs | 7 | Six push/PR jobs plus scheduled numeric maintenance | Current-head workflow never completes because the harness lane is cancelled at 30 minutes. |
| Server commands | 4 | Gameserver, balance harness, API generator, content-manifest generator | Only gameserver is a deployable runtime; no production package exists. |
| Production deployment files | 1 | Content identity manifest only | Five Compose files are test harnesses; no production Docker/Caddy/Compose/runbook. |

## Initial boundary inventory

### Server producers

- The account router declares 11 hand-wired account/session/founder/intent operations and mounts
  ten registry-owned Game UI, minigame, and Soul operations.
- The composed service exposes `/healthz`, `/readyz`, and `/connection/websocket`.
- Persistence spans 74 migrations and systems for accounts, saves, events, routes, Commons,
  guilds, leaderboards, transport, minigames, pets, Soul, attendance, and Exits.
- The production minigame tenant registry composes exactly `pitch.NewTenant()` at
  `server/gameserver/composition.go`; no combat tenant is registered.

### Client consumers

- The only production Svelte game surfaces are implemented by `GameUIApp.svelte` and
  `RunEndSurface.svelte`; UI Foundation has `Amount.svelte` and the test-only `FixtureHost.svelte`.
- `game-ui/runtime.ts` owns account bootstrap, HTTP intents/snapshots, WebSocket frames, and browser
  credential storage.
- Settings currently renders save status and an account note only. It has no recovery, import,
  export, deletion, email, or session-management controls.
- TypeScript modules exist for many server-side mechanics, but module parity is not a player
  consumer. Each still requires a mounted workflow trace.

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
- Audit the 111-row `active-acceptance-ledger.tsv` against plan, implementation, docs, test,
  review range, and current verdict.
- Enumerate route/operation descriptors and match them to actual client/default consumers.
- Map all catalog rows and copy keys through loaders, epoch identity, and real workflows.
- Inventory every test/gate population, oracle, negative control, timeout, exclusion, and artifact.
- Reconcile all 23 live planning directories and 46 archived RFC ranges bidirectionally.
- Audit migration Up/Down contracts, account data ownership, export/deletion retention, and
  backup/restore implications.

The first section-level ID pass is recorded in `design-capability-ledger.tsv`. Rows deliberately
remain coarse until Wave 2 splits each outcome to one independently falsifiable workflow.
