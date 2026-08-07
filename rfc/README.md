# RFC Index

Active implementation specifications. Process: `0000-rfc-process.md`. New RFCs start from
`template.md`; descriptive names are preferred over global sequence numbers.

## Active

| RFC | Status | Parent |
|---|---|---|
| [RFC-0000: The RFC Process](0000-rfc-process.md) | accepted | — |
| [Balance Harness Dispatch Integrity](balance-harness-dispatch-integrity.md) | withdrawn — premise refuted | Balance Harness Foundation |
| [CI Baseline](scaffolding-and-ci.md) | implementing | — |
| [Account & Session Bootstrap](account-and-session-bootstrap.md) | implementing | Save Layer |
| [WebSocket Transport & Fan-out](websocket-transport-and-fanout.md) | implementing | Production Engine / Client Shell / Account Bootstrap |
| [Leaderboards & Balance Epochs](leaderboards-and-epochs.md) | implementing | Production Engine / Gate Predicates / Prestige |
| [Prestige & Exits](prestige-and-exits.md) | implementing | Production Engine / Save Layer / Account Bootstrap |
| [API Foundation](api-foundation.md) | accepted — C1–C17 ruled; implementing | Account API / Transport / Gameserver Composition |
| [Meters Foundation](meters-foundation.md) | accepted — C1–C12 ruled; implementing | Production / Run Genesis / Purchasable Content; unblocks Achievements and Pet Care |
| [Achievements Foundation](clout-and-achievements-foundation.md) | accepted — C1–C10 ruled; scope narrowed to Achievements; implementing | Meters / Copy Pipeline / Production / Run Genesis |
| [Founder Attendance Foundation](founder-attendance-foundation.md) | accepted — A1–A5 ruled; implementing | Save / Run Genesis / Prestige |
| [Minigame Platform Foundation](minigame-platform-foundation.md) | accepted — C1–C40 ruled; implementing | Gameserver Composition / Founder Attendance / Combat |
| [Pet Care Foundation](pet-care-foundation.md) | accepted — C1–C21 ruled; introduces `ApplyFounderLogged`; implementing | Save / Run Genesis / Combat Shared Kernel |
| [Combat Shared Data & Arithmetic](combat-data-model.md) | implementing | — |
| [Combat — Duel Engine](combat-duel-engine.md) | draft | Combat Shared Data |
| [Combat — Lane Engine](combat-lane-engine.md) | draft | Combat Shared Data |
| [Combat — Bots & Integration](combat-bots-and-integration.md) | draft | Combat engines / Account Bootstrap |
| [UI Foundation](ui-foundation.md) | accepted — C1–C11 ruled; implementing | Client Shell / Transport / Copy Pipeline |
| [Game-UI Screens](game-ui-screens.md) | draft | UI Foundation / Client Shell |
| [Commons Onboarding & Governance](commons-onboarding-and-governance.md) | draft — UNBLOCKED, blockers answered | Commons Compact |
| [World Layer Foundation](world-layer-foundation.md) | draft | Commons / Production / Save |
| [Feed & Dispatch Foundation](feed-and-dispatch-foundation.md) | draft | Transport / Production / Clout |
| [Events Engine — Layer 1](events-engine-layer1.md) | draft | Production / Save / Meters |
| [Minigame & Recovery API + Surface](minigame-api-and-surface.md) | accepted — MA-C1–C9 ruled; implementation-ready | API Foundation / Minigame Platform (accepted, implementing) / Soul / UI Foundation |
| [Permits & the T3→T4 Gate](permits-and-t3-gate.md) | draft — pre-mint contract (FCE-C1 ruling) | Economy Kernel / Route Registry / First Content Epoch |
| [First Content Epoch](first-content-epoch.md) | draft — owner-gated mint (successor of TP-C18/SR-C13) | ALL fixture-first content foundations |
| [Deployment Foundation](deployment-foundation.md) | draft — THE PUSH | Gameserver / CI / all archived guards |
| [T0–T1 Playable Content](t0-t1-playable-content.md) | draft | Production / Purchasable Content / Copy Pipeline |

**Current handoff:** `planning/codex-batch-2026-07-29.md` — the ordered implementable-now queue.
**Coverage map:** `planning/coverage-map/` (internal, unpublished) — the validated research→design→RFC→impl tracker and the
dependency-ordered gap backlog for the still-uncontracted (design-only) systems.

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

**Phase-0 contract status (reconciled 2026-08-05 against the coverage-map sweep):** the contracts
previously listed here as "not yet drafted" — Layer-1 events engine, doctrine intents, Compute
Credit spend, game-UI screens, deployment — **now all exist as draft RFCs** (in the Active table
above; drafted 2026-08-03). Remaining structural notes: Prestige & Exits owns the reset; Leaderboards
owns board binding; doctrine intents must define doctrine-pick ordering before same-boundary doctrine
routes can ship. Later named work: an outcome-sensitive near-cap Guild reserved-credit regression
before either a nonzero stock-consumption modifier or multi-worker clearing topology ships.

The **still-uncontracted (design-only) systems** — the ~28 gameplay elements that are researched
and designed but have no RFC yet (individual minigames, the monetization-satire content, the
narrative layer, tiers 2–8) — are tracked with their build-on dependencies and draft-order waves in
`planning/coverage-map/gap-backlog.md`. That backlog, not this paragraph, is the source of truth for
what remains to be drafted.

### Deferred decisions register

RFC-0002's re-scope to the Economy Kernel correctly narrowed it to what was implementable. The
remaining deferred decisions retain named successors so defaults are not chosen silently during
later implementation:

| Deferred decision | Origin | Named owner | Why it matters |
|---|---|---|---|
| **Leaderboard/ranking order keys** — ✅ resolved in [Leaderboards & Balance Epochs](leaderboards-and-epochs.md) D1 (exact integer/time keys; magnitude ties remain ties) | RFC-0002 draft D4 | owned | 12-digit quantization makes runs differing below the 12th digit indistinguishable, in a game framed around world records |
| **Minimum visible increment** — ✅ resolved in [Client Shell & Sim Loop](archive/client-shell-and-sim-loop.md) D3 (interpolate at full precision, sub-unit accumulation, cap shows `reason_key`) | RFC-0002 draft D6 | owned | A frozen number with no explanation is indistinguishable from a bug, and `design/00` forbids unexplained caps |
| **Client reconciliation policy** — ✅ resolved in [Client Shell & Sim Loop](archive/client-shell-and-sim-loop.md) D2 (bend continuous, snap discrete with receipts, story-not-error for gaps) | was a `design/06` table cell | owned | The most player-visible consequence of RFC-0001's whole numeric contract |
