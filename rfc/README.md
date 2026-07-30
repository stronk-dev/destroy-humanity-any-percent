# RFC Index

Active implementation specifications. Process: `0000-rfc-process.md`. New RFCs start from
`template.md`; descriptive names are preferred over global sequence numbers.

## Active

| RFC | Status | Parent |
|---|---|---|
| [RFC-0000: The RFC Process](0000-rfc-process.md) | accepted | — |
| [CI Baseline](scaffolding-and-ci.md) | implementing | — |
| [Account & Session Bootstrap](account-and-session-bootstrap.md) | implementing | Save Layer |
| [Commons Onboarding & Governance](commons-onboarding-and-governance.md) | blocked-on-owner (faction+guild drafts now exist — unblocks when they accept) | Commons Compact |
| [Combat Shared Data & Arithmetic](combat-data-model.md) | implementing | — |
| [Combat — Duel Engine](combat-duel-engine.md) | draft | Combat Shared Data |
| [Combat — Lane Engine](combat-lane-engine.md) | draft | Combat Shared Data |
| [Combat — Bots & Integration](combat-bots-and-integration.md) | draft | Combat engines / Account Bootstrap |
| [WebSocket Transport & Fan-out](websocket-transport-and-fanout.md) | implementing | Production Engine / Client Shell / Account Bootstrap |
| [Leaderboards & Balance Epochs](leaderboards-and-epochs.md) | implementing | Production Engine / Gate Predicates / Prestige |
| [Prestige & Exits](prestige-and-exits.md) | implementing | Production Engine / Save Layer / Account Bootstrap |
| [Faction & Incorporation](faction-incorporation.md) | implementing | Production Engine / Commons / Prestige |
| [Guild Model](guild-model.md) | draft — executable contracts complete; waits on Faction review | Account / Commons / Transport |
| [Run Genesis & Replay](run-genesis-and-replay.md) | draft — DESIGN-GAPs | Leaderboards / Prestige / Account |

**Current handoff:** `planning/codex-batch-2026-07-29.md` — the ordered implementable-now queue.

## Archive

Implemented behavior lives in `docs/`; these frozen RFCs are historical specifications.

| RFC | Status | Canonical docs |
|---|---|---|
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

Remaining Phase-0 contracts (not yet drafted): T0–T1 playable content
(Prestige & Exits owns the reset) · production Balance Epoch artifact/hot-reload semantics
(Leaderboards owns board binding) · doctrine intents (must define doctrine-pick ordering before
same-boundary doctrine routes can ship). Later named work: Compute Credit spend · deployment and
draining.

### Deferred decisions register

RFC-0002's re-scope to the Economy Kernel correctly narrowed it to what was implementable. The
remaining deferred decisions retain named successors so defaults are not chosen silently during
later implementation:

| Deferred decision | Origin | Named owner | Why it matters |
|---|---|---|---|
| **Leaderboard/ranking order keys** — ✅ resolved in [Leaderboards & Balance Epochs](leaderboards-and-epochs.md) D1 (exact integer/time keys; magnitude ties remain ties) | RFC-0002 draft D4 | owned | 12-digit quantization makes runs differing below the 12th digit indistinguishable, in a game framed around world records |
| **Minimum visible increment** — ✅ resolved in [Client Shell & Sim Loop](archive/client-shell-and-sim-loop.md) D3 (interpolate at full precision, sub-unit accumulation, cap shows `reason_key`) | RFC-0002 draft D6 | owned | A frozen number with no explanation is indistinguishable from a bug, and `design/00` forbids unexplained caps |
| **Client reconciliation policy** — ✅ resolved in [Client Shell & Sim Loop](archive/client-shell-and-sim-loop.md) D2 (bend continuous, snap discrete with receipts, story-not-error for gaps) | was a `design/06` table cell | owned | The most player-visible consequence of RFC-0001's whole numeric contract |
