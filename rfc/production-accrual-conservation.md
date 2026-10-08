# RFC: Production Accrual Conservation

- **Status:** draft — decision proposal, NOT implementation authority
- **Author:** Codex
- **Created:** 2026-10-08
- **Design refs:** `design/06-tech.md` lazy/server-authoritative idle math; `design/02-economy-balancing.md` production and hardcaps
- **Depends on:** Clout v1 CV3/AC6; Numeric Core; Production Engine; Save Layer
- **Parent / amends:** Production Engine & Intent API; Clout v1 CV3/AC6; prospective Company save/replay contract
- **Planning:** existing `planning/clout-v1-and-pr-interns/` R-012 evidence; implementation plan after acceptance

## Summary

Extra internal evaluations must not change the resources earned over an otherwise
identical production interval. Current per-evaluation rounding violates that requirement.
Repair requires preserving enough accounting information across the authorized cuts, not a
tolerance, new balance numbers or a different rounding rule. Whether that origin must
be persisted is unresolved: the current read path does not commit production, and logged
transitions are the mutation boundary. This draft proposes an origin-based direction and
identifies the decisions and exact contracts still needed before it is buildable.

## Evidence and scope

At `10d94e70`, cold `TestAxisAccrualPartitionProperty` still fails 27/128 seeded
online/offline cases. The separate 64-arm `TestAxisAttainmentChangesOnlyFutureAccrual`
passes: prior-interval factor timing is not the defect to repair.

Completed R-012 research establishes these bounded facts:

- The conserved-carry experiment matches its bounded population, but materializing
  enormous rationals is not a solution for the admitted exponent range.
- The frozen-anchor experiment preserves the existing one-shot arithmetic in its
  sampled domain, including restoration. It is not a production Company implementation.
- Serializing raw computed rates to canonical 12-digit strings changes five of 64
  tested payouts. Do not use rounded rate strings as a substitute for producer context.
- Existing Company context survives the 64-profile real-jsonb experiment with exact
  reconstructed raw rates. This does not prove a new frozen-context schema or codec.
- Actual logged Go/TS actions, buffs, modes, nonzero permits, provisioning, Exit and
  transactional sequences have bounded evidence. None implements conservation.
- A test-only whole-Company origin now projects all 128 original profiles through
  the actual evaluator and ledger, preserving the original one-shot final state.
  Twelve multi-resource provider/burst/mode intersections, four single-episode
  offline cases and two cap/debit cases also pass. This is bounded Go feasibility,
  not the proposed production codec or a repaired `ApplyLogged` path.

Sources: [anchor](../planning/clout-v1-and-pr-interns/anchor-research.md),
[rates and SQL](../planning/clout-v1-and-pr-interns/rate-and-sql-research.md),
[policy boundaries](../planning/clout-v1-and-pr-interns/policy-boundary-research.md),
[continuous replay](../planning/clout-v1-and-pr-interns/sequence-research.md) and
[persistence](../planning/clout-v1-and-pr-interns/sequence-persistence.md).

Scope: production accounting, its durable state, replay and compatibility. No change
to PR factors, achievement grants, offline efficiency/cap/banking, burst policy,
provider formulas, hardcaps, authored copy or the full nine-tier release objective.

## Proposed contract direction

### C1 — Define the interval before changing arithmetic

Recommended, pending RP-308 owner/author reconciliation: internal chunks of one
continuous absence share one production allowance and banking calculation. Genuinely
different reconnect histories remain different absence episodes under existing policy.
The 48-hour one-call/two-call observations cannot be declared equal by a rounding fix.

The accepted contract must name the server-owned episode start/end and their persisted
replay evidence. The client supplies neither mode nor time. Equal-partition comparisons
keep real action chronology, content, producer inputs and episode identity unchanged;
they do not invent additional reconnects or intents. This question has been delivered
to Marco; delivery is not adoption or permission to edit another author's ruled body.

### C1a — Decide origin lifetime before adding save fields

The current call paths do not establish a need for a new persisted origin:

| Boundary | Actual current behavior | Consequence for this proposal |
|---|---|---|
| `gameui.Projector.GameUISnapshot` | Loads the committed Company/Founder; projects rates and display clocks. | A display read is not an accrual settlement or a new saved head. |
| `production.Service.Handle` | Builds replay inputs and invokes ApplyLogged inside the Store mutation callback. | Applied transitions own persistence; internal work does not independently commit. |
| `ApplyLogged` / `ApplyLoggedExit` | Evaluate, run the closed hooks and apply the semantic transition. | The complete operation is the existing replay/rollback unit. |
| `applyReplayOfflineCatchup` | Runs within the logged transition before its ordinary action. | The offline and ordinary arms are distinct policy work, not arbitrary equivalent cuts. |
| `ApplySuppressedLogged` | Consumes recovery time while restoring production-bearing outputs. | Recovery suppression is a semantic transition, not an ordinary earning cut. |
| Simulation/content diagnostics | Evaluate hypothetical states. | Their work is not authority to write live Company revisions. |

The existing persisted public-projection integration test now samples four additional
display times before the same later Exit. The display clock advances, while complete
committed resource/generator views and all twelve watched table populations stay unchanged.
This is bounded evidence for that real Postgres projector path, not every HTTP/read caller.

Two lifetime designs therefore remain explicit:

- **Operation-local origin:** retain the origin through internal chunks of one logged
  operation; crash/retry replays from its committed pre-state. Do not invent partial commits.
  This is sufficient only if the adopted interval definition has no authorized commit
  inside an unsettled interval. Its complete action/hook and Go/TS proofs are still owed.
- **Durable origin:** if the adopted interval spans authorized saved heads, define the
  exact context, its consistency validation and versioned replay/restore semantics.
  A durable format cannot be justified merely by the scalar partition failure.

The Clout author must resolve the interval/settlement meaning before selecting either.
Neither choice may delete the original property, redefine cuts as actions to avoid failures,
or waive accepted behavior. The earlier assertion that durable state was necessarily
required was stronger than the evidence; it is corrected here, not implemented as fact.

### C2 — Preserve a producer origin, not a rounded-rate cache

Candidate direction, with lifetime subject to C1/C1a: retain a canonical initial balance,
the origin time and the semantic context needed to reconstruct the original production calculation. Later
internal evaluations project from that origin, not from the last rounded balance.
Visible balances still cross the existing 12-digit state boundary.

Before acceptance, enumerate the exact context fields and their authority: purchased
and provisioned counts/remainders, pinned content, ordered multiplier inputs, active
buff/burst timing and any external frozen contribution inputs. Do not assume a scalar
rate list or a copied whole Company is sufficient. Restored context must reconstruct
every original raw rate without a new raw-float wire format or ambient reads.

Automatic provision/burst boundaries require an explicit accounting algorithm. Resetting
the ledger at every minute would introduce extra quantization compared with the existing
one-shot evaluation; simply resetting at every producer change is not yet a solution.
Multi-resource accounting and the catalog's original source/slot order remain mandatory.

#### Bounded executable algorithm and write set

`server/production/axis_origin_projection_test.go` exercises the following candidate:

1. Restore the unchanged encoded origin and reconstruct its contributions using the
   actual pinned catalog. Evaluate it once from the original cursor to the requested time.
2. For each resource, form a canonical nonnegative difference between that projected
   balance and the currently visible balance. Submit those entries to `Ledger.ApplyAccrual`;
   require the committed balances to equal the projected targets exactly and every receipt
   delta to reproduce its target. A failed exact commit is a failure, not a direct overwrite.
3. Transfer only Evaluate's non-ledger outputs: `ComputeCreditMS`,
   `ComputeBurstRemainingMS`, `GeneratorProvisioned`, `ProvisionRemaindersPPM` and
   `EvaluatedThrough`. Require complete encoded Company equality with the one-shot projection.
4. Restore the visible state after every cut, retaining the distinct unchanged origin.
   At a real debit boundary, settle first and create a new origin from the post-debit state.

The original branch that instead rebases at each cut retains all 27 discrepancies.
The additional intersection checks require positive permits, expired burst, and literal
provision counts/remainders; the bank cases check floor/saturation, and the debit cases
require literal post-debit payouts rather than recovery of earlier capped overflow.

This experiment freezes the entire existing Company only to avoid guessing a partial
schema. It uses no external/active-buff contributions, SQL or new versioned fields. It
does not establish a production storage bound or prove that all admitted numeric targets
can be reached through a quantized difference. The exact minimal context and that domain
obligation remain acceptance blockers, not implicit conclusions of these 146 profiles.

### C3 — Settlement, caps and receipts are part of the repair

Specify which applied actions settle the old interval before a debit or input change,
and which origin is used afterward. A new PR factor cannot reprice the prior interval.
Rejected actions must preserve the complete state; internal evaluations cannot silently
become settlement actions. Mode/episode changes follow C1, not a convenient reset rule.

Do not implement this by pre-projecting the live state before calling the unchanged
`ApplyLogged`. The existing transition evaluates and then passes its `EvaluationResult`
to the closed Prestige → Faction → Guild → Commons hook chain. Pre-advancing the cursor
would make that evaluation report zero elapsed time and omit legitimate hooks. Conversely,
running those hooks for every internal projection would turn cuts into additional semantic
events. The accepted repair must bind elapsed/production/banking/progress and resource
receipts to the logical settlement interval, preserving event order and once-only hooks.
Internal work cuts do not authorize unlogged persisted Company writes, new commands or
extra revisions. Any durable partial-work protocol must be specified explicitly rather
than added beside the existing closed mutation boundary.

The ledger remains the atomic resource authority. A projected balance must have a
canonical nonnegative accrual receipt whose delta re-adds to that exact balance.
No direct balance overwrite may bypass minimum/hardcap/transaction checks. Saturation
must not bank hidden overflow that becomes spendable after a later debit. Below-cap
quantities that round visibly to the cap require an explicit, compatible rule.

The scalar prototype's 512-byte envelope is not a production storage/CPU bound. Establish
bounds from the actual proposed field set and measured full-state workloads; guard
exhaustion must invalidate the operation, not truncate history or reset rounding.

### C4 — Compatibility, and versioned persistence if required

Do not assign save/replay/kernel version numbers until the contract and landing order
are known. If new persisted context is adopted, specify new-run activation versus migration,
missing/extra-field refusal and old-run dispatch. An operation-local representation still
requires an explicit arithmetic/replay compatibility route; absence of new wire fields does
not authorize silently changing historical pinned outcomes or losing replay evidence.

JSON transport validation and logical Postgres jsonb restoration are separate boundaries:
SQL can reorder fields and normalize whitespace. Validate complete typed context,
identity and reconstructed values, not transport byte order inside jsonb.

Keep the original 128 seeds/cuts and failure evidence. Any versioned fixture evolution
must be explicit and retain old-semantics compatibility tests; it cannot silently replace
the red population with a convenient new one or alter its monetary expectations.

## Proposed acceptance criteria

1. Original 128 partition cases pass exactly on the accepted new semantics, with complete
   final-state equality and unchanged prior-interval timing proof. No tolerance or skips.
2. Paired Go/TS full-state sequences cover multiple internal cuts, SQL restart at authorized
   persist points, purchases, re-attainment, buff/burst expiry, provision boundaries,
   multiple resources and modes.
   Rebase-at-each-cut and future-factor controls must fail the relevant requirements.
   If persisted internal cuts are adopted, restoration at those cuts is mandatory too;
   this draft does not authorize introducing such commits beside the closed boundary.
3. Existing numeric goldens and legitimate one-shot payouts remain exact. Include the five
   known rounded-rate differences, exponent edges, near-cap/debit and no-overflow-bank cases.
4. New codec/version/identity/reconstruction refusals execute in both runtimes. Real
   Service/Store/Postgres tests prove commit, rollback, identical retry and retained replay.
5. Old pinned histories retain their exact behavior; new-run/migration/Exit behavior follows
   the adopted activation contract. Missing or forged accrual context cannot be accepted.
6. Full-state storage and execution measurements satisfy the adopted operating bounds without
   hidden truncation. Actual affected CI lanes pass; research artifacts are not CI success.
7. Canonical numeric/production/save/replay docs, exact-range designated review and all feature
   gates reconcile before archival. This repair alone does not establish 1.0 readiness.

## Open questions before acceptance

- **C1/C1a:** owner/author adoption of absence-episode/settlement meaning, exact identity
  boundaries and whether an unsettled interval can span authorized committed heads.
- **C2/C3:** complete field set, boundary integration and ledger/receipt algorithm, demonstrated
  with actual Company producers. This draft deliberately does not invent those missing details.
- **C4:** activation/migration compatibility and assigned versions after the representation is fixed.
- **Operating limits:** actual full-state size/work evidence, not the scalar prototype's bound.

No implementation is authorized by choosing the candidate direction alone. Resolve these
questions in the body and obtain acceptance/designated review before production changes.

## Deviations from design

No reward or numeric-law deviation proposed. Explicit interval/origin lifetime extends the
implementation contract. A durable conservation format, if selected, is not an already
implied save field and needs its own complete contract.

## Changelog

- 2026-10-08: decision draft from completed R-012 evidence and current cold regressions.
- 2026-10-08: add actual evaluator/ledger origin projection feasibility and the explicit
  closed-hook/persistence integration boundary. Still draft; no runtime authority.
- 2026-10-08: correct the unsupported assertion that new durable state is necessarily
  required. Trace the actual closed mutation/read paths; distinguish operation-local and
  durable origins pending author resolution. No accepted criterion or runtime changed.
