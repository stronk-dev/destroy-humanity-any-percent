# RFC Index

Active implementation specifications. Process: `0000-rfc-process.md`. New RFCs start from
`template.md`; descriptive names are preferred over global sequence numbers.

## Active

| RFC | Status | Parent |
|---|---|---|
| [RFC-0000: The RFC Process](0000-rfc-process.md) | accepted | — |
| [CI Baseline](scaffolding-and-ci.md) | implementing | — |
| [Production Engine & Intent API](production-engine-and-intents.md) | draft | — |
| [Production Accrual Math](production-accrual-math.md) | implementing | [Production Engine & Intent API](production-engine-and-intents.md) |

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

Planned next (not yet drafted — carve from `design/07-roadmap.md` Phase 0): client shell & sim loop
· balance harness · deployment and draining.

### Deferred decisions register

RFC-0002's re-scope to the Economy Kernel correctly narrowed it to what is implementable — but two
decisions it originally owned are now listed only as "out of scope" with **no named successor**.
That is precisely the failure mode RFC-0002 was written to catch: *a decision with a reasonable
default and no owner gets made silently by whoever writes the code first.* Assigning owners:

| Deferred decision | Origin | Named owner | Why it matters |
|---|---|---|---|
| **Leaderboard/ranking order keys** — must be exact integers or times, never quantized `Decimal`; ties displayed as ties | RFC-0002 draft D4 | **balance harness / leaderboard RFC** | 12-digit quantization makes runs differing below the 12th digit indistinguishable, in a game framed around world records |
| **Minimum visible increment** — a counter must never appear frozen while production > 0; a counter frozen *at a cap* must show the cap and its `reason_key` | RFC-0002 draft D6 | **client shell & sim loop RFC** | A frozen number with no explanation is indistinguishable from a bug, and `design/00` forbids unexplained caps |
| **Client reconciliation policy** — snap vs. lerp vs. rebase | never in an RFC; asserted in a `design/06` table cell | **client shell & sim loop RFC** | The most player-visible consequence of RFC-0001's whole numeric contract |
| **Offline progression constants** | stranded in `AGENTS.md` | **production engine RFC** | A balance number living in an onboarding brief is binding on agents but untestable and unversioned |
