# Dependency and shared-resource graph

Coordinate: product tree `190a4fa`; evidence through the Wave-5 checkpoint; 2026-08-20.

This is a dependency DAG, not a feature list. `dependency-resource-ledger.tsv` is its machine-
readable producer -> transformation -> consumer -> refusal -> witness inventory. A dependency is
not satisfied because a type or RFC with a similar noun exists.

## Release authority spine

```text
R-002 baseline ──> D-001 milestone floor ──> D-007 content scope
      │                         │
      │                         ├──> exact release tasks ──> accessibility owner ──> R-005
      │                         ├──> complete first session ───────────────────────> R-008
      │                         └──> content DAG / honest label
      │
      ├──> D-002 repository disposition
      ├──> D-003 sunset posture ──> refreshed research ──> sunset contract
      └──> D-005/D-008/D-009 account posture ──> prototype ──> R-003

accepted R-001 instrument ──> local/hosted measurement ──> D-014 CI contract
      └──> CI repair ──> current-head hosted verdict ──────────────────────────────┐
                                                                                  │
D-006 topology + D-011 observability + D-015 retention                            │
      ├──> Deployment package/config/backup ──> R-006                             │
      └──> Operations/retention contract ─────> R-007                             │
                                                                                  │
D-016 telemetry ──> aggregate instrument ──> R-009 ──> first real milestone       │
                                                                                  v
                                                         integrated release acceptance
```

The release is the intersection of those lanes. Deployment cannot promote missing account rights,
accessibility, content, or CI. Conversely, a complete browser loop cannot prove backup, retention,
or incident response.

## Runtime and player-surface spine

```text
Account backend ──> recovery/export/deletion/session successor ───────────────┐
       │                                                                       │
       └──> token rotation ─┐                                                   │
                            v                                                   │
Transport history/cursors ──> production reconnect/drain controller ───────────┤
                                                                                v
API one-registry authority ──> generated client ──> Game UI snapshot-v3 controls
                                      │                    │
                                      │                    ├──> first-hour / R-004
                                      │                    └──> account Settings / R-003
                                      v
Minigame Platform backend ──> Minigame/Recovery API ──> visible tenant surfaces
```

Required order:

1. ruling authors reconcile API, Game UI, Minigame Platform/API, Account fallback, and Transport
   bodies where later rulings contradict normative text;
2. API owns all v1 operations and generates the only browser client;
3. Account and Transport bind token rotation and per-scope positions at the production runtime;
4. Game UI consumes server eligibility and the generated client; and
5. account/minigame/accessibility research runs only on those real workflows.

## Gameplay/content spine

```text
Combat Shared exact catalogs/tables/slash scope
    ├──> Duel Engine ─┐
    └──> Lane Engine ─┴──> Bots & Match Integration ──> pet/minigame surfaces

Leaderboards/replay ──> API public readers/evidence ──> speedrun/spectator/Game UI

Commons/Guild backend ──> Onboarding/Governance ─┐
                                                   ├──> World Layer ──┐
Telemetry/R-009 ──────────────────────────────────┘                  │
Meters/routes/achievements ──> Events Layer 1 ───────────────────────┤
                                                                      v
Transport feed channel ───────────────────────────────> Feed/Dispatch/UI

reviewed mechanics + real content + migration + fixed historical witness
    ──> content epoch mint ──> later tier/content RFCs ──> designed endings
```

World thresholds cannot precede D-016/R-009: `design/00` and `design/07` explicitly forbid a
global milestone without measured throughput. Feed depends on real World/Event/Commons producers
and D-017's UGC/moderation posture; a transport channel is not a feed feature.

## Active RFC disposition against the graph

| Active RFC | Actual graph position | Current gate |
|---|---|---|
| CI Baseline | release evidence root | RP-059 accepted instrumentation owner, R-001, then D-014/body reconciliation |
| Account & Session | backend producer with Q-001 witnesses designated-approved | broader player capability needs D-005/D-008/D-009 successor; body/range closeout separate |
| WebSocket Transport | server, production recovery consumer and ruled AC3 witness designated-approved | body/range closeout separate |
| Leaderboards & Epochs | backend producer for public evidence | ruling-author body repair, exact witnesses, reader/UI successor, current review union |
| Prestige & Exits | backend transition producer | ruling-author body repair and D-012 Advisor choice before witness closeout |
| API Foundation | shared wire/client authority | ruling-author repair, then registry/generator/public/runtime work |
| Minigame Platform | tenant/session producer | body must stop claiming absent Combat; repair offline-quality output and acceptance scope |
| Combat Shared | shared child contract | RP-072 tables/effects and slash-scope ruling, then exact fixtures/current review |
| Game UI Screens | main player consumer | ruling-author repair, snapshot-v3 controls, recovery/client dependencies, R-004/R-005 |
| Minigame API + Surface | designated-approved backend API witnesses plus absent player consumer | body/exact surface/client dependencies block visible work |
| Permits | live content contract awaiting closeout | exact replay/body/docs/provenance repair only |
| First Content Epoch | historical mint authority awaiting closeout | fixed epoch-6 witness/body/changelog/range reconciliation only |
| Deployment | release package owner, currently draft | D-001/D-002/D-003/D-006/D-011/D-015 and complete body rewrite before acceptance |
| Combat Duel / Lane | mode engines, drafts | exact Combat Shared parent first |
| Combat Bots & Integration | downstream match/fallback, draft | both engines plus Minigame/Account integration and blinded bot acceptance |
| Commons Onboarding | player consumer, draft | exact Account/API/Game UI dependencies and milestone scope; “unblocked” does not mean accepted |
| World Layer | shared world producer, draft | D-016/R-009, exact API/UI/content contracts and accepted body |
| Events Layer 1 | narrative producer, draft | dependency/status reconciliation, accepted DSL/effect/refusal/content contract |
| Feed & Dispatch | downstream social consumer, draft | World + Events + Commons producers, D-017, API/UI/moderation contract |
| Balance Harness Dispatch Integrity | withdrawn | no downstream edge; D-010 chooses its durable lifecycle location |

## Confirmed dependency defects

1. **Status-as-dependency:** Minigame Platform calls the Combat kernel/duel tenant implemented;
   they are not. Feed calls Transport implemented while its active closeout and client recovery are
   open. Draft headers must name exact bounded resources, not optimistic RFC status.
2. **Missing browser authority:** Game UI and Minigame surfaces consume hand-written fetch paths
   while API Foundation owns a generated-only client rule. UI work cannot proceed in parallel with
   a second wire authority.
3. **Missing release owners:** cross-surface accessibility, account rights/recovery, operations,
   retention, and sunset have no accepted active implementation RFC. Deployment cannot absorb
   their product decisions implicitly.
4. **Missing cross-system edges:** Commons Onboarding needs Account/API/Game UI; World needs ruled
   telemetry and public/player consumers; Combat Bots needs the Minigame match platform; Deployment
   needs client/license/Account/API/operations contracts. Their headers under-declare these edges.
5. **Content vs mechanism:** First Content Epoch/Permits closeout does not authorize tier 2–8
   content. Conversely, future content may not mint against mechanical packages lacking a default
   player consumer and fixed historical witness.
6. **Research after build:** R-003/R-004/R-005/R-006/R-007/R-008/R-009 validate concrete ruled
   workflows. None may be used to choose the contract its population requires.

RP-090 records the header/graph drift. Existing RP-016/RP-019/RP-046/RP-052/RP-060/RP-063/
RP-071/RP-075–RP-078/RP-085–RP-089 carry the specific underlying defects.

## Shared-resource ownership laws

- One resource has one canonical contract owner. Other RFCs amend or consume it explicitly.
- API Foundation owns wire/client generation; feature RFCs own operation semantics, not alternate
  clients or envelopes.
- Account owns credential/data-right semantics; Transport owns delivery position/recovery; Game UI
  owns presentation. Their composed acceptance is shared, their authorities are not duplicated.
- Release package, backups, licenses, migration set, config, export import, and sunset artifact bind
  through one versioned release manifest.
- Accessibility follows release tasks and is consumed by every surface. Component axe is only a
  lower-level producer.
- Operations metrics and gameplay telemetry are separate resources with separate privacy and
  acceptance contracts.
- Every content epoch binds accepted mechanics, real data/copy, migrations, a default consumer,
  refusal behavior, fixed historical replay, and exact review provenance.
