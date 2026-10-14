# Headcount Allocation implementation plan

Authority: accepted `rfc/headcount-allocation.md`; design/01 Tier 2.

## Current scope

2026-10-14 bounded S1/S2 correction (RP-571): reject non-string role IDs before sorting or
assignment lookup, preserving all valid literal names and arithmetic. Eight old-source
regressions fail; final Node56/native Chromium-WebKit112, cold Go/vet and types pass. The owning
log records the falsified line-ending hypothesis, exact source/test/record range after
`4c216b69` and pending Claude review. No live consumer, catalog, save, rate or output-model
selection. Hiring/output questions are re-presented to Marco, still unanswered/unadopted.

Begin with S2's pure effective-heads arithmetic, independently of the held OD-1
seat source. Both runtimes receive an explicit seat count, hardcap, complete
assignment and the arithmetic subset of already-resolved role definitions.
They return source production, diverted heads, idle converters and unassigned
seats. They acquire no seats and touch no catalog, save, clock, rate or intent.

Verified locally: literal shared vectors and exhaustive small-domain comparisons against
independent exact-integer test models, plus safe-integer boundaries and invalid
assignment/role definitions. Run Go tests/vet, Node and actual browser consumers,
and types. Reuse existing test collection; no new CI lane or measurement framework.
The final Go/Linux-amd64 package, Node42, Chromium/WebKit84 and type/vet checks pass;
native Firefox executes zero tests after launch/session timeouts. This is library construction, not a player
workflow or full AC3/AC5 completion. Cross-party review remains required.

## Remaining construction

### Current reconciliation boundary — 2026-10-13

There are two player-facing choices, not merely a missing version number:

- **Hiring:** a separate paid Hire action, or seats granted by purchased office generators.
- **Output:** staff directly produce resources, or staff multiply existing generator/manual
  output while Slide Decks remain a displayed allocation count rather than stored production.

Both questions have been presented to Marco; neither delivery nor a preselected answer is a
ruling. The recommended pair is separate hiring plus direct resource production, following
the standalone engine's division of responsibility. This is a proposal, not authority to
replace Tier 2's previously ruled role effects. A mixed choice needs its own explicit
reconciliation; answering only hiring does not select the output economy.

| Boundary | Headcount S0–S10 | Tier 2 H1–H5 | Required before integration |
|---|---|---|---|
| Seats | Purchased by `hire_headcount`; persisted per pool | Derived from purchased generator units; no hire action | Adopt the hiring model and reconcile both bodies |
| Roles | `produce` / `convert` / `pool_feed`; resource rates | Per-effective-head multipliers on generators/clicks | Adopt the output model; define any selected extension explicitly |
| Slide Decks | Converter resource output accumulates over time | Derived `slide_decks` count produces nothing | Fix its meaning in the chosen output model, not just its label |
| Catalog | `headcount_pools` with price/cap/window/effect rows | `headcount` plus `headcount_seats` generator roles | One canonical engine-owned schema; Tier 2 consumes it |
| Saved state | `headcount_seats` and `headcount_assigned` | `headcount_assignments` with derived budget | One activation/codec/migration contract, no duplicate state authorities |
| Commands/events | `hire_headcount`, `reallocate_headcount`; hired/reallocated events | `set_headcount_allocation`; allocated event | One closed wire and replay union, with matching rejection/receipt rules |
| Harness | Hiring relevance; role masks/reporting pending the allocation-aware persona | Generator-seat masks; allocation-aware P1/P5 gate | Reconcile the selected model without dropping either required measurement |

The ruling author must reconcile the accepted bodies before implementation. Provisional
prices/rates, copy adoption, measured pacing/relevance and the mint remain separate gates;
choosing a model does not ratify data or loosen acceptance. RP-413's invalid overflow fixture
and required-but-unproved mutant outcome also remain with the author.

Version evidence at `769b26a2`: economy schema5 already owns Clout `axis_stack`; Company19
already owns Clout attainment; Founder latest25 and replay-inputs latest12. The old v5/v19/v9
Headcount literals are not available reservations. Recheck and assign next-free identities at
the actual implementation boundary, with exact activation/history compatibility; do not
reinterpret existing versions or edit applied migrations.

After those bodies are reconciled and reviewed, build the complete selected chain: catalog →
saved seats/assignments → logged commands → production/replay → authoritative projection →
mounted allocation panel → real purchase/reallocation/Exit/reload proof. The existing S2
library is only one prerequisite, not a substitute for that chain.

- Resolve OD-1 with Tier 2's author/owner before seat-source-dependent work.
- Reconcile the siblings' conflicting catalog/state/intent/output contracts
  (RP-414), not only the seat source. No integration branch is chosen here.
- Reconcile S8/AC3's inadmissible overflow vector and unproved fault outcome
  (RP-413); `overflow-oracle-audit.md` now derives why TS's specified overflow
  output fault cannot occur in the admitted domain, and records actual Go overflow
  and TS ratio-control failures. The author must reconcile the criterion; do not
  manufacture a TS overflow failure, remove its guard or promote AC3/AC5.
- Reconcile occupied economy/save/replay version numbers in the accepted body
  before integration; do not reuse its historical v5/v19/v9 literals blindly.
- Catalog grammar/bindings, closed-form rates, state/migrations and logged intents.
- Replay, published formulas, harness and authoritative UI projection/panel.
- Full acceptance, adopted content, integrated player proof and designated review.

New arithmetic modules are not imported by the live engine/client. Kernel replay
semantics and watched paths remain unchanged here. Before authoritative wiring,
register both new module prefixes and perform the genuine integration version bump
with its replay/migration tests; this is not an exception to kernel identity.
