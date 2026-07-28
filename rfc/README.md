# RFC Index

Active implementation specifications. Process: `0000-rfc-process.md`. New RFCs start from
`template.md`; descriptive names are preferred over global sequence numbers.

## Active

| RFC | Status | Parent |
|---|---|---|
| [RFC-0000: The RFC Process](0000-rfc-process.md) | accepted | — |
| [CI Baseline](scaffolding-and-ci.md) | implementing | — |
| [Gate Predicates & the Route Registry](gate-predicates-and-routes.md) | draft | Production Engine & Intent API |
| [The Commons Compact](commons-compact.md) | draft | Production Engine & Intent API |
| [Combat Data Model](combat-data-model.md) | draft | — |
| [Client Shell & Sim Loop](client-shell-and-sim-loop.md) | draft | Production Engine & Intent API |

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

Planned next (not yet drafted): balance harness · Compute Credit spend · deployment and
draining.

### Deferred decisions register

RFC-0002's re-scope to the Economy Kernel correctly narrowed it to what was implementable. The
remaining deferred decisions retain named successors so defaults are not chosen silently during
later implementation:

| Deferred decision | Origin | Named owner | Why it matters |
|---|---|---|---|
| **Leaderboard/ranking order keys** — must be exact integers or times, never quantized `Decimal`; ties displayed as ties | RFC-0002 draft D4 | **balance harness / leaderboard RFC** | 12-digit quantization makes runs differing below the 12th digit indistinguishable, in a game framed around world records |
| **Minimum visible increment** — ✅ resolved in [Client Shell & Sim Loop](client-shell-and-sim-loop.md) D3 (interpolate at full precision, sub-unit accumulation, cap shows `reason_key`) | RFC-0002 draft D6 | owned | A frozen number with no explanation is indistinguishable from a bug, and `design/00` forbids unexplained caps |
| **Client reconciliation policy** — ✅ resolved in [Client Shell & Sim Loop](client-shell-and-sim-loop.md) D2 (bend continuous, snap discrete with receipts, story-not-error for gaps) | was a `design/06` table cell | owned | The most player-visible consequence of RFC-0001's whole numeric contract |
