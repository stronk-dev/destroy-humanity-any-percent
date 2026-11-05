# RFC Index

Active implementation specifications. Process: `0000-rfc-process.md`. New RFCs start from
`template.md`; descriptive names are preferred over global sequence numbers.

## Active

| RFC | Status | Parent |
|---|---|---|
| [RFC-0000: The RFC Process](0000-rfc-process.md) | accepted | — |
| [Balance Harness Dispatch Integrity](balance-harness-dispatch-integrity.md) | withdrawn — premise refuted | Balance Harness Foundation |
| [CI Baseline](scaffolding-and-ci.md) | implementing | — |
| [Reputation Harness Observability](reputation-harness-observability.md) | draft — technical specification independently approved 2026-11-02; owner adoption pending, not implementation authority | Harness Observability / CI Baseline / Reputation Tree v1 |
| [Kernel History Guard Integrity](kernel-history-guard-integrity.md) | draft — technical specification independently approved 2026-11-02; owner adoption pending, not implementation authority | Run Genesis & Replay KV-1 / CI Baseline |
| [Numeric Diagnostic Exponent Limits](numeric-diagnostic-exponent-limits.md) | draft — RP-663 beyond-int64 diagnostic exception proposed; normative review/adoption pending; not implementation authority | Numeric Core / Numeric Core Boundary Hardening |
| [Exact Represented Aggregation](exact-represented-aggregation.md) | draft — RP-664 exact operation, pinned exceptions and producer/capacity cutover proposed; normative review/adoption and workload proof pending; not implementation authority | Deterministic Decimal Aggregation / Economy / Commons / Harness |
| [Account & Session Bootstrap](account-and-session-bootstrap.md) | implementing | Save Layer |
| [Account Recovery Player Flow](account-recovery-player-flow.md) | draft — complete corrected proposal independently technically approved 2026-11-04; owner flow/copy adoption requested, not implementation authority | Account / archived Game UI / API Foundation |
| [Browser Session Renewal & Recovery](browser-session-renewal.md) | draft — RP-048/RP-234 consumer diagnosed; policy/API/browser gates unaccepted | Account / Transport / API Foundation / archived Game UI |
| [Game UI First-read Failure & Recovery](game-ui-first-read-recovery.md) | draft — RP-342 measured; startup precedence/retry/copy unaccepted; not implementation authority | Garage Player Surfaces / archived Game UI / API Foundation / Account |
| [Monotonic Founder Attendance](founder-attendance-monotonicity.md) | draft — RP-365 reproduced; D-024 shared-clock boundary unruled | Founder Attendance / Pet Care & Adoption / Garden / Minigame Platform |
| [WebSocket Transport & Fan-out](websocket-transport-and-fanout.md) | implementing | Production Engine / Client Shell / Account Bootstrap |
| [Leaderboards & Balance Epochs](leaderboards-and-epochs.md) | implementing | Production Engine / Gate Predicates / Prestige |
| [Leaderboard Readers & Player Surface](leaderboard-readers-and-player-surface.md) | draft — not implementation authority | Leaderboards / API Foundation / Game UI / Route Registry |
| [Prestige & Exits](prestige-and-exits.md) | implementing | Production Engine / Save Layer / Account Bootstrap |
| [API Foundation](api-foundation.md) | accepted — C1–C20 ruled; implementing | Account API / Transport / Gameserver Composition |
| [Public Company Replay Evidence](public-company-replay-evidence.md) | draft — RP-606; independently technically approved 2026-11-05; owner adoption/body authority pending; not implementation authority | API C2/C14/C19 / Server Garden SG8/AC10/OD-12 / Run Genesis & Replay |
| [Artifact Schema Ownership & Kernel-neutral Delivery](api-artifact-schema-ownership.md) | draft — independently technically approved 2026-11-04; owner adoption/history-scope prerequisite remain; not implementation authority | API C17/C18 / Route C2 / Kernel History Guard Integrity |
| [Typed Dictionaries in API Schemas](api-typed-dictionaries.md) | draft — RP-391; independently technically approved 2026-11-04; owner choice/adoption pending; not implementation authority | API Foundation C17 / existing intent receipts |
| [Explicit Integer Token Encoding in API Schemas](api-integer-token-encoding.md) | draft — RP-566; independently technically approved 2026-11-04; owner adoption pending; not implementation authority | API Foundation C17/C18 / relevance owner grammar |
| [Original Artifact Bytes in Catalog Transport](api-artifact-byte-transport.md) | draft — RP-597/RP-598; independently technically approved 2026-11-04; owner adoption pending; not implementation authority | API Foundation C12/C17/C18 / artifact owners |
| [Bodiless Responses in the API Registry](api-bodiless-responses.md) | draft — RP-483; independently technically approved 2026-11-04 including BR2 correction; owner adoption pending; not implementation authority | API Foundation C19 / Account lifecycle / public caching |
| [Minigame Platform Foundation](minigame-platform-foundation.md) | accepted — C1–C40 ruled; implementing | Gameserver Composition / Founder Attendance / Combat |
| [Combat Shared Data & Arithmetic](combat-data-model.md) | implementing | — |
| [Combat — Duel Engine](combat-duel-engine.md) | draft | Combat Shared Data |
| [Combat — Lane Engine](combat-lane-engine.md) | draft | Combat Shared Data |
| [Combat — Bots & Integration](combat-bots-and-integration.md) | draft | Combat engines / Account Bootstrap |
| [Commons Onboarding & Governance](commons-onboarding-and-governance.md) | draft — dependency foundations available; ballot/cohort contracts still unclosed (RP-529; [readiness](../planning/commons-onboarding-and-governance/log.md)) | Commons Compact |
| [World Layer Foundation](world-layer-foundation.md) | draft | Commons / Production / Save |
| [Feed & Dispatch Foundation](feed-and-dispatch-foundation.md) | draft | Transport / Production / Clout |
| [Events Engine — Layer 1](events-engine-layer1.md) | draft | Production / Save / Meters |
| [Minigame & Recovery API + Surface](minigame-api-and-surface.md) | accepted — MA-C1–C15 ruled; implementing | API Foundation / Minigame Platform (accepted, implementing) / Soul / UI Foundation |
| [Minigame Tenant HTTP Bindings](minigame-tenant-http-bindings.md) | draft — technical specification independently approved 2026-11-02; adoption/parent bodies and Typer author holds remain; not implementation authority | API Foundation C2 / Minigame API / Terminal Typer TT-PA4 / Deployment DP5 |
| [Permits & the T3→T4 Gate](permits-and-t3-gate.md) | promoted in reviewed epoch 6; archival pending | Economy Kernel / Route Registry / First Content Epoch |
| [First Content Epoch](first-content-epoch.md) | epoch 6 minted and designated-approved; dependent archival in progress | ALL fixture-first content foundations |
| [Deployment Foundation](deployment-foundation.md) | implementing — accepted at `cd102d7` | Gameserver / Client / Account / API / CI / operations |
| [Current Release Authority & Interrupted Transitions](deployment-current-release-authority.md) | draft — complete linked adoption package independently technically approved throughed9a8f2a; owner adoption/native construction and acceptance remain | Deployment DP5 / operator ledger / R-006 |
| [Deployment Operation Evidence](deployment-operation-evidence.md) | draft — complete linked adoption package independently technically approved throughed9a8f2a; actual producers/owner adoption remain; not implementation authority | Current Release Authority CR3/CR6 / Deployment DP1/DP5 |
| [Native Deployment Supervisor Profile](deployment-native-supervisor-profile.md) | draft — prior direct-Compose/P01 reviews retained; complete linked adoption package independently technically approved throughed9a8f2a; owner adoption/native construction/conformance pending | Operation Evidence OP9 / Current Release Authority CR4/CR6 / Deployment |
| [Deployment Tool Artifact Inventory and Delivery](deployment-tool-artifact-inventory.md) | draft — complete single upstream-Compose proposal independently technically approved through70acee7f; tool-only MPL policy/artifact adoption pending; not implementation authority | Native Supervisor SP1/SP5/SP8 / Deployment DP1/AC8 |
| [Backup Protection Metadata Binding](backup-protection-metadata.md) | draft — complete proposal independently technically approved 2026-11-04; owner adoption pending, not implementation authority | Deployment DP6 / retention |
| [Accessibility of Player Workflows](accessibility-player-workflows.md) | draft — current Game UI failures measured; owner task/AT matrix pending | UI Foundation / Game UI Screens; later surface RFCs consume its acceptance floor |
| [v0.1 "The Garage" — Release Manifest & RFC DAG](v0.1-garage-release-manifest.md) | accepted 2026-09-25 (recommended defaults); implementing | design/07 Phase 1 |
| [Reputation Tree v1](reputation-tree-v1.md) | accepted 2026-09-25 (recommended defaults); implementing | Prestige & Exits / Leaderboards & Epochs |
| [Minigame — Terminal Typer](minigame-terminal-typer.md) | accepted 2026-09-25 (recommended defaults); implementing | Minigame Platform / Minigame API & Surface |
| [Minigame — Demo Disc Arcade](minigame-demo-disc-arcade.md) | accepted 2026-09-25 (recommended defaults); implementing | Minigame Platform / Minigame API & Surface |
| [Minigame — Arcade Pinball](minigame-arcade-pinball.md) | draft — not implementation authority; owner-requested addition (not in design/03), OD-P1–P15 | Demo Disc Arcade / Minigame Platform |
| [Tier 2 Content — IT Company](tier2-content.md) | accepted; paid-hire/output model ruled 2026-11-02; body translation independently approved; full content/policy/Headcount integration remains | Economy Kernel / Routes / Balance Harness |
| [Tier 2 Pacing Activity Windows](tier2-pacing-activity-window.md) | accepted 2026-11-02 — owner-adopted; parent-body translation independently approved; full P1/Headcount prerequisites remain | Tier 2 Content P1/P2 |
| [Tier 2 Paid-hiring Policy Binding](tier2-paid-hiring-policy.md) | draft — decisions and concrete modern Headcount binding independently technically approved; owner adoption pending; not implementation authority | Tier 2 Content P1 / Headcount S9 / accounting-search H1 |
| [Minigame — Server Garden](minigame-server-garden.md) | accepted 2026-09-25 (recommended defaults); implementing | Minigame Platform / Fiscal / Tier 2 |
| [Garage Player Surfaces](garage-player-surfaces.md) | accepted 2026-09-25 (recommended defaults); implementing | Game UI Screens / Fiscal / Achievements / Meters / Pet / Active Play / Minigame API |
| [Clout v1 & PR Interns](clout-v1-and-pr-interns.md) | accepted 2026-09-25 (recommended defaults); implementing | Achievements / Economy Kernel |
| [Production Accrual Conservation](production-accrual-conservation.md) | draft — technically reviewed operation-local proposal; owner adoption and implementation pending | Production Engine / Clout / Save Layer |
| [Accrual-aware Harness Search](accrual-aware-harness-search.md) | draft — technically reviewed caller/search/report contracts; owner adoption and implementation pending | Production Accrual Conservation / Relevance R10–R11 / first-hour and content diagnostics |
| [Cosmetic Shop v1 — Horse Armor](cosmetic-shop-v1.md) | accepted 2026-09-25 (recommended defaults); implementing | Game UI / Save Layer / Pet |
| [Pet Adoption v1](pet-adoption-v1.md) | accepted 2026-09-25 (recommended defaults); implementing | Pet Care / Save Layer / Game UI |
| [Ticker Launch Corpus](ticker-launch-corpus.md) | accepted 2026-09-25 (recommended defaults; blocked on Feed VQ-3); implementing | Feed & Dispatch / Copy Pipeline |
| [Headcount Allocation](headcount-allocation.md) | accepted; paid-hire/output model ruled 2026-11-02; translation independently approved; historical identity/AC-3 reconciliation remains | Economy Kernel / Tier 2 Content |
| [Headcount Activation & Acceptance Integrity](headcount-activation-and-acceptance-integrity.md) | draft — technical correction independently approved 2026-11-02; author/owner adoption pending, not implementation authority | Headcount / Tier 2 Content |

**Current handoff:** [`planning/platform-alignment/execution-queue.md`](../planning/platform-alignment/execution-queue.md)
— the evidence-derived READY/BLOCKED queue at HEAD. The older July batch manifest is history,
not current authorization.
**Coverage map:** [`planning/archive/coverage-map/`](../planning/archive/coverage-map/) preserves the
2026-08-05 historical sweep as a frozen, noncanonical snapshot. Use the platform-alignment
capability and active-RFC audits for current status.

## Archive

Implemented behavior lives in `docs/`; these frozen RFCs are historical specifications.

| RFC | Status | Canonical docs |
|---|---|---|
| [Copy Pipeline Foundation](archive/copy-pipeline-foundation.md) | implemented | [Copy pipeline](../docs/copy-pipeline.md) |
| [Purchasable Content Foundation](archive/purchasable-content-foundation.md) | implemented | [Purchasable content](../docs/purchasable-content.md), [Production engine](../docs/production-engine.md), [Economy kernel](../docs/economy-kernel.md) |
| [Gameserver Composition](archive/gameserver-composition.md) | implemented | [Gameserver composition](../docs/gameserver.md), [Guilds](../docs/guilds.md) |
| [Run Genesis & Replay](archive/run-genesis-and-replay.md) | implemented | [Leaderboards & epochs](../docs/leaderboards-and-epochs.md) |
| [RFC-0001: The Numeric Core](archive/0001-numeric-core.md) | implemented | [Numeric core](../docs/numeric-core.md) |
| [Numeric Core Boundary Hardening](archive/numeric-core-boundary-hardening.md) | implemented | [Numeric core](../docs/numeric-core.md) |
| [Numeric Normalization Carry](archive/numeric-normalization-carry.md) | implemented | [Numeric core](../docs/numeric-core.md) |
| [RFC-0002: Economy Kernel](archive/0002-economy-constants-and-ceilings.md) | implemented | [Economy kernel](../docs/economy-kernel.md) |
| [Geometric Affordability Fast Path](archive/geometric-afford-fast-path.md) | implemented | [Economy kernel](../docs/economy-kernel.md) |
| [Save Layer & Migrations](archive/save-layer-and-migrations.md) | implemented | [Save layer](../docs/save-layer.md) |
| [Production Accrual Math](archive/production-accrual-math.md) | implemented | [Numeric core](../docs/numeric-core.md) |
| [Generator Production State](archive/generator-production-state.md) | implemented | [Economy kernel](../docs/economy-kernel.md), [save layer](../docs/save-layer.md) |
| [Save Archive Compare-and-Swap](archive/save-archive-cas.md) | implemented | [Save layer](../docs/save-layer.md) |
| [Numeric Boundary Parity](archive/numeric-boundary-parity.md) | implemented | [Numeric core](../docs/numeric-core.md) |
| [Deterministic Decimal Aggregation](archive/deterministic-decimal-aggregation.md) | implemented | [Numeric core](../docs/numeric-core.md), [economy kernel](../docs/economy-kernel.md) |
| [Production Engine & Intent API](archive/production-engine-and-intents.md) | implemented | [Production engine](../docs/production-engine.md), [economy kernel](../docs/economy-kernel.md), [save layer](../docs/save-layer.md) |
| [Production Hardcap Saturation](archive/production-hardcap-saturation.md) | implemented | [Production engine](../docs/production-engine.md), [economy kernel](../docs/economy-kernel.md) |
| [Millisecond Cursor Canonicalization](archive/millisecond-cursor-canonicalization.md) | implemented | [Save layer](../docs/save-layer.md), [Production engine](../docs/production-engine.md) |
| [Resource-Log Domain Parity](archive/resource-log-domain-parity.md) | implemented | [Economy kernel](../docs/economy-kernel.md), [Production engine](../docs/production-engine.md) |
| [Production Contract Assertions & Integrity](archive/production-contract-integrity.md) | implemented | [Production engine](../docs/production-engine.md), [Save layer](../docs/save-layer.md) |
| [Balance Harness Foundation](archive/balance-harness-foundation.md) | implemented | [Balance harness](../docs/balance-harness.md), [Production engine](../docs/production-engine.md) |
| [Harness Observability](archive/harness-observability.md) | implemented | [Balance harness](../docs/balance-harness.md) |
| [Gate Predicates & the Route Registry](archive/gate-predicates-and-routes.md) | implemented | [Routes](../docs/routes.md), [Production engine](../docs/production-engine.md), [Save layer](../docs/save-layer.md) |
| [The Commons Compact](archive/commons-compact.md) | implemented | [Commons](../docs/commons.md), [Production engine](../docs/production-engine.md), [Save layer](../docs/save-layer.md) |
| [Client Shell & Sim Loop](archive/client-shell-and-sim-loop.md) | implemented | [Client shell](../docs/client-shell.md) |
| [Commons Projection Retry Idempotency](archive/commons-projection-retry-idempotency.md) | implemented | [Commons](../docs/commons.md) |
| [Route Registry Event-Order Convergence](archive/route-registry-event-order-convergence.md) | implemented | [Routes](../docs/routes.md) |
| [Balance Baseline Change Guard Hardening](archive/balance-baseline-change-guard.md) | implemented | [Balance harness](../docs/balance-harness.md) |
| [Post-Review Integrity Remediation](archive/post-review-integrity-remediation.md) | implemented | [Balance harness](../docs/balance-harness.md), [Routes](../docs/routes.md), [Commons](../docs/commons.md), [Save layer](../docs/save-layer.md) |
| [Route Temporal Validity](archive/route-temporal-validity.md) | implemented | [Routes](../docs/routes.md) |
| [Commons Cohort Merge Capacity](archive/commons-merge-capacity.md) | implemented | [Commons](../docs/commons.md) |
| [Phase-0 Pacing Observation Coverage](archive/phase0-pacing-observation-coverage.md) | implemented | [Balance harness](../docs/balance-harness.md) |
| [Faction & Incorporation](archive/faction-incorporation.md) | implemented | [Factions](../docs/factions.md), [Production engine](../docs/production-engine.md), [Save layer](../docs/save-layer.md) |
| [Guild Model](archive/guild-model.md) | implemented | [Guilds](../docs/guilds.md), [Production engine](../docs/production-engine.md), [Save layer](../docs/save-layer.md) |
| [Doctrine & Compute Credit](archive/doctrine-and-compute-credit.md) | implemented | [Doctrine and Compute Credit](../docs/doctrine-and-compute-credit.md), [Production engine](../docs/production-engine.md), [Save layer](../docs/save-layer.md), [Routes](../docs/routes.md) |
| [Relevance Harness](archive/relevance-harness.md) | implemented | [Balance harness](../docs/balance-harness.md), [Production engine](../docs/production-engine.md) |
| [Fiscal Quarters Foundation](archive/fiscal-quarters-foundation.md) | implemented | [Fiscal Quarters](../docs/fiscal-quarters.md), [Production engine](../docs/production-engine.md), [Save layer](../docs/save-layer.md) |
| [Active-Play Buff Windows](archive/active-play-buff-windows.md) | implemented | [Active-play opportunities](../docs/active-play.md), [Production engine](../docs/production-engine.md), [Save layer](../docs/save-layer.md) |
| [Soul Foundation](archive/soul-foundation.md) | implemented | [Soul foundation](../docs/soul.md), [Production engine](../docs/production-engine.md), [Save layer](../docs/save-layer.md) |
| [The Pitch](archive/minigame-the-pitch.md) | implemented | [The Pitch](../docs/minigame-the-pitch.md), [Minigame platform](../docs/minigame-platform.md) |
| [Soul Recovery Activities](archive/soul-recovery-activities.md) | implemented fixture-first | [Soul Recovery](../docs/soul-recovery.md), [Soul foundation](../docs/soul.md) |
| [Founder Attendance Foundation](archive/founder-attendance-foundation.md) | implemented | [Founder transitions](../docs/founder-transitions.md), [Save layer](../docs/save-layer.md) |
| [Meters Foundation](archive/meters-foundation.md) | implemented | [Meters](../docs/meters.md), [Production engine](../docs/production-engine.md), [Save layer](../docs/save-layer.md) |
| [Achievements Foundation](archive/clout-and-achievements-foundation.md) | implemented | [Achievements](../docs/achievements.md), [Production engine](../docs/production-engine.md), [Save layer](../docs/save-layer.md) |
| [Pet Care Foundation](archive/pet-care-foundation.md) | implemented; combat AC3 carried | [Pet care](../docs/pet-care.md), [Founder transitions](../docs/founder-transitions.md), [Save layer](../docs/save-layer.md) |
| [UI Foundation](archive/ui-foundation.md) | implemented | [UI foundation](../docs/ui-foundation.md) |
| [Game-UI Screens](archive/game-ui-screens.md) | implemented | [Game UI](../docs/game-ui.md) |
| [T0–T1 Playable Content](archive/t0-t1-playable-content.md) | implemented | [T0–T1 playable content](../docs/t0-t1-playable-content.md), [Balance harness](../docs/balance-harness.md), [Prestige and Exits](../docs/prestige-and-exits.md) |

**Phase-0 contract status (reconciled 2026-08-05 against the coverage-map sweep):** the contracts
previously listed here as "not yet drafted" — Layer-1 events engine, doctrine intents, Compute
Credit spend, game-UI screens, deployment — **now all exist as draft RFCs** (in the Active table
above; drafted 2026-08-03). Remaining structural notes: Prestige & Exits owns the reset; Leaderboards
owns board binding; doctrine intents must define doctrine-pick ordering before same-boundary doctrine
routes can ship. Later named work: an outcome-sensitive near-cap Guild reserved-credit regression
before either a nonzero stock-consumption modifier or multi-worker clearing topology ships.

The **still-uncontracted (design-only) systems** were counted as ~28 in the 2026-08-05 coverage
sweep (individual minigames, monetization-satire content, narrative, tiers 2–8). Its internal
[`planning/archive/coverage-map/gap-backlog.md`](../planning/archive/coverage-map/gap-backlog.md)
retains the historical dependency analysis as a tracked, frozen snapshot. Do not treat either its
count or queue as current authority.

### Deferred decisions register

RFC-0002's re-scope to the Economy Kernel correctly narrowed it to what was implementable. The
remaining deferred decisions retain named successors so defaults are not chosen silently during
later implementation:

| Deferred decision | Origin | Named owner | Why it matters |
|---|---|---|---|
| **Leaderboard/ranking order keys** — ✅ resolved in [Leaderboards & Balance Epochs](leaderboards-and-epochs.md) D1 (exact integer/time keys; magnitude ties remain ties) | RFC-0002 draft D4 | owned | 12-digit quantization makes runs differing below the 12th digit indistinguishable, in a game framed around world records |
| **Minimum visible increment** — ✅ resolved in [Client Shell & Sim Loop](archive/client-shell-and-sim-loop.md) D3 (interpolate at full precision, sub-unit accumulation, cap shows `reason_key`) | RFC-0002 draft D6 | owned | A frozen number with no explanation is indistinguishable from a bug, and `design/00` forbids unexplained caps |
| **Client reconciliation policy** — ✅ resolved in [Client Shell & Sim Loop](archive/client-shell-and-sim-loop.md) D2 (bend continuous, snap discrete with receipts, story-not-error for gaps) | was a `design/06` table cell | owned | The most player-visible consequence of RFC-0001's whole numeric contract |
