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

#### External producer context is distinct from the Company origin

`TestAxisOriginDeferredExternalContextResearch` supplies a bounded additional finding:
three producer-context profiles (real Commons modifier, real frozen Fiscal modifiers,
and their combination with the declared Guild membership signal), each online/offline,
at2001ms with four internal scratch cuts. The existing exact replay-accrual codec retains
every resource's ordered raw rates, complete projections and subsequent logged manual
state/receipt/events. Dropping the external contributions changes projected cash in all
six cells. Pre-advancing the Company instead preserves cash but loses Prestige accrual
and the present Commons/Guild hooks. Company encoding alone is therefore insufficient
for these profiles; a rounded aggregate rate is still not a replacement.

Any eventual accounting context must bind the committed pre-state and pinned artifact
identity to these **existing** resolved inputs, rather than fetch their current values
at each internal cut:

- External `contributions`: exact ordered `{slot,source_id,target,factor}` rows. Their
  admitted canonical factors are inputs to reconstructing raw rates, not encoded raw rates.
  Preserve declaration/target validation and source/slot order.
- `commons_weight_ppm`: the nullable participation weight used by the Commons accrual
  event, distinct from the Commons production modifier and Company solidarity samples.
- `guild_settlement_batch`: the existing identity/base sequence/ordered debit-credit rows.
  A nonempty batch changes the pre-state before rate derivation and must apply once at its
  existing transition boundary. The external-context experiment uses an empty batch; the
  separate bounded settlement experiment below covers three nonempty batch profiles.
- `route_context_version`, Founder carry and active-play evidence: retain their current
  authorities in the closed transition. The external-context experiment does not prove
  that active buffs can be frozen for a whole interval, omit temporal schedule evidence,
  or resolve the shared attendance/absence policy.

This is a constraint on the pending design, not a new origin/save wire, adoption of an
origin lifetime or a production fix. The initial120001ms population refused before action
on the existing active-play due-transition ceiling; the successful shorter population
does not establish that longer boundary. Live provider/SQL state capture, general Guild
settlement workloads, general temporal workloads, Go/TS parity and full-domain conservation remain owed.

`TestAxisOriginDeferredTemporalContextResearch` adds42 admitted Go cells over
production frenzy, building special and their combination, at seven offsets in
both evaluation modes. Controlled existing histories use the unchanged pinned
durations. Actual schedule evidence survives the existing replay codec; fresh
scratch schedule/evaluation does not consume the committed origin. Its ledger,
plus the one accepted manual action, matches the actual closed transition ledger.
Retained replay reproduces whole state, exact receipt and ordered events; forged
attended coordinates or missing schedule evidence refuse without mutation.
Frozen-at-origin modifiers change cash in24 expired cases, while omitting all
modifiers changes cash in24 still-active cases. Legitimate equal controls remain.

This observes the **existing endpoint-selection semantics**: the schedule removes
expired buffs before the command evaluates its accrual interval. It does not prove
prorated payout through a buff boundary, nor authorize replacing endpoint behavior
with a segmented-window formula. Short ModeOffline intervals still advance the
attended coordinate; the6000ms case instead crosses the existing5000ms catchup
ceiling and records the whole absence span, preserving the origin's attended
coordinate in either mode. A future conservation implementation must retain that
context/compatibility explicitly, not infer absence solely from evaluation mode or
silently freeze active contributions. This is not adoption of the shared attendance
policy, durable origin, a new numeric algorithm, SQL/TS proof or AC6 completion.

`TestAxisOriginDeferredGuildSettlementResearch` adds six admitted Go cells: one clearing,
two ordered clearings and a nonzero existing watermark, each online/offline at2001ms.
The actual input resolver applies pending clearing before the actual stock-consumption
provider and restores the pre-state afterward. Fixed membership and batch lookup are test
inputs, not a live Guild/SQL claim. Existing replay encoding retains the exact batch and
contributions. Fresh scratch copies consume clearing independently; a subsequent logged
manual action on the unchanged committed Company matches the ordinary complete state,
receipt and ordered events, including positive Prestige/Guild hooks. Applying that same
batch again refuses without mutation. A known-generator unaffordable purchase, whose
refusal occurs after accrual/hooks, restores the whole original Company; the subsequent
valid manual action still reproduces the ordinary result.

The pinned Guild consumption bonus is zero. Omitting clearing therefore preserves cash
but loses stock/consumption/watermark state in all six cells. Cash equality cannot prove
clearing delivery, even with a membership contribution present. This constrains C2/C3's
once-only boundary and failure atomicity; it is not adoption of origin lifetime, an absence
policy, an arbitrary batch algorithm, temporal buff semantics or a conservation repair.

#### Bounded executable algorithm and write set

`server/production/axis_origin_projection_test.go` exercises the following candidate:

1. Restore the unchanged encoded origin and reconstruct its contributions using the
   actual pinned catalog. Evaluate it once from the original cursor to the requested time.
2. For each resource, attempt to form a canonical nonnegative difference between that projected
   balance and the currently visible balance. Submit those entries to `Ledger.ApplyAccrual`;
   require the committed balances to equal the projected targets exactly and every receipt
   delta to reproduce its target. A failed exact commit is a failure, not a direct overwrite.
   This step fails admitted lower-domain cases below; it is not a general production algorithm.
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
does not establish a production storage bound. The exact minimal context remains an
acceptance blocker, not an implicit conclusion of these 146 profiles.

#### Lower-domain counterexample: the candidate write set is insufficient

The predeclared 2026-10-12 extension executes20 cells through actual Company admission,
Evaluate, codec restoration and `Ledger.ApplyAccrual`: five admitted in-memory generator
rate exponents, both modes and1000→1001ms /2000→2001ms. The fixture retains its existing
1.001 frozen multiplier. Its modified catalog is hash-bound but not adopted game content.
Nine increasing projected pairs produce a non-state difference and the ledger correctly
refuses them without mutation; ten commit exactly and one is unchanged. There are no
wrong-target commits. The retained census checks positive controls as well as refusals.

For example, both `1.001e-8999999999999999` and `1.002001e-8999999999999999` are admitted
canonical projected balances. Their required increment lies below the existing exponent
floor, so quantized subtraction is not an admitted ledger delta. Five literal client
arithmetic controls independently retain this refusal, the near-floor offline refusal,
and near-floor/ordinary successes; they are not TS Company or SQL integration evidence.
See `planning/clout-v1-and-pr-interns/log.md` and RP-488 for commands and limits.

Consequently step2 cannot satisfy the unchanged numeric/receipt contract across the
admitted domain. A viable accounting/ledger algorithm is required before acceptance;
choosing an origin lifetime alone does not resolve it. Do not silently restrict the domain,
change rounding/receipt semantics, overwrite balances or substitute a tolerance. This
finding concerns a test-only candidate; it neither proves current served content reaches
these synthetic rates nor attributes the original27 partition failures to this edge.

#### Intermediate commits and deferred settlement are different contracts

The subsequent predeclared boundary experiment retains the actual floor pair above.
Zero accrual leaves `1.001e-8999999999999999`; the minimum positive admitted delta
`1e-8999999999999999` produces `2.001e-8999999999999999`, already above the desired
`1.002001e-8999999999999999`. Actual ledger receipts reproduce those literal results;
three larger positive controls also overshoot. Go rejects sub-floor construction;
client arithmetic independently confirms zero/minimum sums and canonical sub-floor
refusal. This is more specific than an unsuccessful subtraction implementation.

The normalized-domain bound gives the analytical constraint: a positive state delta
has mantissa at least1 and exponent at least the floor. For this pair, its minimum
increment already exceeds the desired difference by orders of magnitude, well beyond
the quantizer's rounding width. There is no canonical nonnegative accrual increment
that reaches that intermediate target under the current ledger addition rule.
This argument is not a finite-search proof over the whole numeric domain; it concerns
the retained pair. It rules out committing every projected cut in that population
while preserving both the target and the existing receipt rule.

A separate test-only operation-local experiment restores/projects scratch copies at
two intermediate times, leaves the original committed Company untouched, then makes
one real ledger settlement from that original state. All20 lower-domain cells reach
the exact full one-shot Company state with reproducing receipts. These origins have
zero cash; no claim about arbitrary origins, saved intermediate heads, complete
ApplyLogged/hooks, SQL or a production representation follows. Earlier candidate
refusals and all27 original partition failures remain unchanged.

Therefore C1/C1a must explicitly distinguish internal work/display projections from
authorized saved settlements. Deferred settlement is a bounded candidate for the
former, not permission to drop a required saved head, hide a production failure or
redefine AC6. If the adopted interval requires these intermediate commits, the
conflicting contract must be reconciled explicitly; a better subtraction cannot
satisfy an unrepresentable increment. Neither a domain restriction nor a receipt
change is adopted here.

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

The test-only `TestAxisOriginDeferredLoggedPurchaseResearch` now executes all128 original
seed307/end/cut/mode profiles with nonzero initial cash through actual `ApplyLogged`
purchases. Scratch projections leave the committed pre-state unchanged; subsequent logged
settlement reproduces the ordinary full Company, receipt and ordered events, including
positive prestige lifetime accrual and purchase reattainment. All128 deliberately
pre-advanced controls still apply and produce the same cash, but lose prestige lifetime
accrual. Resource equality alone therefore cannot validate this integration. This is
bounded Go evidence for leaving the closed transition in charge, not a new accounting
implementation, all-hook/Go–TS/SQL proof, lower-floor purchase proof or AC6 repair.

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
- **C2/C3:** complete field set, boundary integration and a viable ledger/receipt algorithm,
  demonstrated with actual Company producers and the retained lower-domain counterexample.
  The current difference-write candidate fails; this is technical work, not a choice for the
  owner to approve away. Operation-local deferred settlement now has bounded lower-domain
  evidence, not full integration. If intermediate saved heads are required, reconcile the
  demonstrated unrepresentable-increment conflict rather than demand another subtraction
  algorithm or waive the original failures. This draft does not invent those missing details.
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
- 2026-10-12: retain actual lower-domain ledger counterexamples and paired client arithmetic
  controls. Mark the candidate difference-write algorithm insufficient rather than generalize
  its146 earlier successes. Still draft; no numeric, receipt, policy or implementation authority.
- 2026-10-12: distinguish the floor pair's unrepresentable positive increment from a subtraction
  defect. Record20 bounded deferred-settlement successes and the unresolved internal-cut/saved-
  settlement contract. No production implementation, criterion weakening or adoption.
- 2026-10-12: add bounded closed-purchase evidence with nonzero origins and full state/receipt/
  event comparisons; all128 pre-advanced controls expose lost prestige accrual despite matching
  cash. Preserve the unresolved lifetime/compatibility and production accounting obligations.
- 2026-10-12: record bounded temporal producer/replay evidence and distinguish current endpoint
  selection from prorated window integration. No runtime semantics or pending owner choice changed.
