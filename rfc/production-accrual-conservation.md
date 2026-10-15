# RFC: Production Accrual Conservation

- **Status:** draft — decision proposal, NOT implementation authority
- **Author:** Codex
- **Created:** 2026-10-08
- **Design refs:** `design/06-tech.md` lazy/server-authoritative idle math; `design/02-economy-balancing.md` production and hardcaps
- **Depends on:** Clout v1 CV3/AC6; Numeric Core; Production Engine; Save Layer
- **Parent / amends:** Production Engine & Intent API; Clout v1 CV3/AC6; prospective Company save/replay contract
- **Consumer follow-up:** [Accrual-aware Harness Search](accrual-aware-harness-search.md); required consumer proof remains part of full AC6 closeout
- **Planning:** existing `planning/clout-v1-and-pr-interns/` R-012 evidence; implementation plan after acceptance

## Summary

Extra internal evaluations must not change the resources earned over an otherwise
identical production interval. Current per-evaluation rounding violates that requirement.
Repair requires preserving enough accounting information across the authorized cuts, not a
tolerance, new balance numbers or a different rounding rule. The recommended proposal
keeps origins operation-local and retains existing saved-settlement/receipt semantics:
the current read path does not commit production, and logged transitions own mutation.
This is not yet adopted. The linked harness successor separately owns origin-aware
search/reuse; neither specification can claim the complete repair from primitive proof alone.

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

Marco explicitly delegated RP-308 wording reconciliation on 2026-10-15, preserving
the 24-hour offline cap, 90% rate, banking and separate-return-visit behavior.
Clout CV3/AC6 now distinguish internal chunks of one continuous absence, which share
one production allowance and banking calculation, from genuinely different return
histories. Display reads and internal calculations create no extra saved payments.
The 48-hour one-call/two-call observations cannot be declared equal by a rounding fix.
That normative wording amendment awaits designated review; this draft remains unaccepted.

The accepted contract must name the server-owned episode start/end and their persisted
replay evidence. The client supplies neither mode nor time. Equal-partition comparisons
keep real action chronology, content, producer inputs and episode identity unchanged;
they do not invent additional reconnects or intents. The delegation authorizes wording
reconciliation only, not an accounting algorithm, new episode state or origin lifetime.

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

The operation-local design is this draft's recommended proposal. The durable
alternative remains explicit research, not an implementer-selectable fallback:

- **Operation-local origin:** retain the origin through internal chunks of one logged
  operation; crash/retry replays from its committed pre-state. Do not invent partial commits.
  This is sufficient only if the adopted interval definition has no authorized commit
  inside an unsettled interval. Its complete action/hook and Go/TS proofs are still owed.
- **Durable origin:** if the adopted interval spans authorized saved heads, define the
  exact context, its consistency validation and versioned replay/restore semantics.
  A durable format cannot be justified merely by the scalar partition failure.

The reconciled Clout wording alone does not adopt that recommendation or specify the
settlement protocol. Neither choice may delete the original property, redefine cuts
as actions to avoid failures, or waive accepted behavior. The earlier assertion that durable state was necessarily
required was stronger than the evidence; it is corrected here, not implemented as fact.

### C1b — Concrete operation-local proposal (unadopted)

Engineering recommendation: keep the origin private to one existing logged operation
and keep its existing settlement stages. Do not introduce an additional durable origin,
periodic payment, public proof API or background player tick. This is a proposed
architecture, not an inference that Marco's wording delegation adopted it.

**Ownership and identity.** Bind an operation to the existing `ReplayCommand` tuple
`{company_stream_id, founder_id, revision, run_seq, run_log_seq, intent_id}`, the
pinned `ConstantsHash`, canonical payload and validated replay-input envelope.
The server resolves that envelope before applying the command; replay consumes the
recorded envelope, never a second live provider lookup. These are existing authorities,
not proposed client fields. A private stage discriminator distinguishes offline catchup
from ordinary accrual in the same command. Do not mint a new persisted episode ID:
the existing `offline_catchup.opened_at_ms` and `offline_span.{from_ms,to_ms}` remain
the catchup coordinates and must still match `buildOfflineCatchup` exactly.

**Origin and context.** Capture the complete admitted Company pre-stage state, not
just cash, using the existing versioned Company codec and its pinned economy. Retain
the derived `FactionStockResource` as `cloneReplayState` does; do not add it to saves.
Retain the complete already-validated replay envelope and pinned bundle, including
the external contribution rows, Commons weight, Guild batch, route context version,
Founder carry and active-play evidence described in C2. No raw computed-rate cache
or partial ad-hoc state schema replaces these inputs. Private snapshots must own
their mutable maps/slices; mutating or discarding a scratch result cannot alter the
origin, another result or the pending real command.

**Stage boundaries remain semantic, not arbitrary cuts.** The ordinary `ApplyLogged`
path currently applies Guild clearing, assembles catchup contributions, performs
the validated offline catchup and its hooks, applies the active-play schedule, then
assembles ordinary contributions and executes the intent's accrual/hooks/action.
Preserve that order. If internal projection is needed inside a stage, capture its
origin after that stage's preceding real mutations, not once before the whole
command and then replay Guild clearing on an already-cleared state. Offline and
ordinary stages are not merged into one new policy interval. An ordinary stage
at the catchup timestamp can have zero elapsed time; preserve its existing behavior.

| Existing boundary | Proposed treatment |
|---|---|
| Ordinary intent and opportunity claim | One accrual settlement per existing stage, then existing hooks and action; a new factor only affects the next interval. |
| Exit and scripted-branch preview | Keep their respective existing pipelines; a preview owns a clone and cannot settle the real Company or consume the real origin. |
| Recovery suppression | Keep its existing consume-time/restore-output contract; do not route suppressed time through an earning origin. |
| Snapshot/rate read | No settlement or saved head; do not add a payment-capable origin to the read API. |
| Simulation wait edge/content diagnostics | Separate non-live ownership is required as described below; preserving live transactions does not repair these callers. |

**Internal projection and final settlement.** Every permitted fixed-context scratch
projection starts from a fresh copy of its stage origin and evaluates the total
origin-to-requested-time interval. It never starts from the previous rounded scratch
result. Returned projected state and `EvaluationResult` are diagnostic values, not
incremental payments or resource authority. Do not pass their receipts to the live
ledger, run external hooks for a cut, or pre-advance the real cursor. The existing
closed transition settles once from its actual pre-stage state and supplies that
single evaluation result to the existing hook chain. On refusal, discard all scratch
work and retain the existing complete rollback. SQL commits, identical retries and
crash recovery remain owned by the existing Store transaction/run-log boundary.
No operation-local object survives that operation or a process restart.

This deliberately recomputes from the origin rather than accumulating rounded
deltas. It addresses the lower-floor receipt constraint without admitting cumulative
receipts or promising monotone scratch targets. Endpoint-selected buffs remain
endpoint-selected: different scheduled endpoint contexts are not one fixed-context
projection session. The RP-652 decreasing targets cannot be installed as payments
or silently converted to prorated buff windows. Any more efficient accumulator must
prove equivalent full-state results before replacing this reference computation.

**The remaining call-contract gap must not be hidden.** The current
`TestAxisAccrualPartitionProperty` calls the mutating `Evaluate` twice on the same
state. Its first cut really advances the ledger and cursor. An operation-local
projection cannot make those unchanged calls equal while also declaring the first
call noncommitting. Retain that diagnostic and its 27 failures; neither clone-only
feasibility tests nor this proposal repairs the existing API. RFC acceptance must specify
the actual calculation/settlement API and its real caller, and exercise the same
128 seed/end/cut/mode profiles through it with complete exact state comparisons.
Changing that call boundary requires an explicit reviewed contract, not quietly
rewriting the test, excluding cases or calling the old failures legacy-green.
Do not add an unused production abstraction solely to obtain a passing test.

#### Real simulation callers are not covered by a command-only origin

Source tracing at `7eb64e37` identifies real non-test consumers, not just the
standalone AC6 diagnostic:

- `production.SimulateAdvance` mutates its supplied harness state, returns an
  evaluation and runs the supplied simulation hook. It refuses emitted events on
  that action-free edge; it does not promise that hook-owned state is unchanged.
- `harness.RelevanceSuite.advanceMode` uses that function for search wait edges.
  `firstHourRuntime.step` and its bank-wait branch use it too. Search branches and
  session boundaries must not be conflated with live persisted command identities.
- `contentDynamicsAccrueWindow` explicitly compares a one-step buff window with
  two successive mutating evaluations and refuses a differing result. The resulting
  check is a real content diagnostic, not an independent saved player payment.

These observations are static caller evidence, not executed new harness failures
or proof that any particular published report is invalid. The operation-local
proposal above must not claim to repair those boundaries by changing only
`ApplyLogged`, or quietly exempt them from conservation.

A simulation extension must define branch-owned origin/context lifetime across
action-free waits, clone/ablation ownership, changes to producer context, offline
session boundaries, and the complete hook-owned state. A later projected wait must
not add the same lifetime accrual twice, omit legitimate Guild/Commons updates or
consume a sibling search branch's origin. A real simulated action closes the old
interval according to the adopted chronology; a search observation does not invent
an action merely to reset rounding. This is still an engineering gap, not a new
product decision for the owner to approve away. Preserve current report bytes and
the original red population while specifying and proving the actual caller route.

There is also a concrete search-identity dependency: archived Relevance R11 keys a
node by `{canonical_state_hash, virtual_ms}`. If two equal visible projections retain
different origins that can produce different later payouts, that key is insufficient.
A candidate origin-carrying simulation design must either prove future equivalence
for every merged node or explicitly propose an R11 amendment binding origin and
producer-context identity in deduplication (with branch-private ablation state).
Changing the key can change search populations, budgets and reports; do not silently
alter the accepted search contract or regenerate its goldens as part of a ledger fix.

The bounded `TestAxisOriginProjectionSearchIdentityResearch` executes this constraint
with the original online3114ms/cut1553ms counterexample: one origin and a second
origin encoded at its cut project to byte-identical complete visible Company state
at that cut. Their future cash is respectively `1.00035846811e4` and
`1.00035846812e4`. The2000ms/cut1000ms control has the same visible-head equality
and equal future state/cash `1.00023023e4`. Both origins remain unchanged. This
verifies non-equivalence for an origin-carrying candidate (RP-653), not an actual
collision in today's origin-free beam or proof that its published reports are wrong.

### C1c — Proposed calculation/action API and branch lifecycle (unadopted)

The following names describe a proposed internal interface, not existing exported
functions or permission to replace their callers. Live logged operations retain
C1b's transaction ownership. Harness branches need a separate, immutable calculation
handle rather than an extra field quietly attached to a persisted `save.State`.

| Operation | Required input and result | Mutation and lifetime |
|---|---|---|
| `BeginCalculation` | Complete admitted base state, its source save version, pinned bundle, mode and validated producer/hook/ablation context → opaque origin. | Copies/owns mutable inputs; cannot begin from a projected view merely to discard rounding history. |
| `ForkCalculation` | Origin → independent branch handle with the same origin/context identity. | No accrual, hooks, action, clock change or revision; branch work cannot mutate siblings. |
| `ProjectCalculation` | Origin plus a canonical server/simulation endpoint → detached full state, evaluation, ordered private event values and role observations. | Fresh restoration from the origin, then actual evaluation and the permitted simulation hooks on scratch state; no new origin, saved payment or action. |
| `ApplyCalculationIntent` | Origin plus an actual intent, endpoint and expected revision → ordinary decision and, only if applied, the next origin. | Run the existing semantic transition from the origin, not from a prior projection. A refusal returns no child and leaves the parent unchanged. |

Projection endpoints must not precede the origin. Search may probe endpoints out of
order on independent scratch state; probing an earlier time after a later probe is
not a rollback of the origin. The detached result includes its origin/context
identity and endpoint. It is not an input to `BeginCalculation` or an incremental
receipt. Ordinary rate/quote reads may inspect its copy but cannot promote it into
settlement authority. Private implementation fields and caller boundaries must
enforce this distinction; a test-only convention is insufficient.

For a real intent at `t`, restore the origin and execute its normal accrual, hooks
and action at `t` once, even if search already inspected a projection at `t`.
Do not first install that projection and call `SimulateTransition` at the same
timestamp: its zero-elapsed evaluation would omit the interval's lifetime/stock
work. The resulting applied full state, including post-action attainment and debit,
becomes the next origin. Existing authoritative return/mode boundaries close their
old stage using the existing offline allowance/banking rules; arbitrary probe times,
provision ticks, rounding changes or failed purchases do not create such boundaries.
No partial-work handle survives a live Store commit or crash.

**Hook ownership is part of the API, not caller folklore.** For each projection,
restore all hook-owned state from the origin, allocate fresh role/mask collectors,
and run the permitted chain exactly once for the complete origin-to-endpoint
evaluation. Never append its cumulative lifetime/stock result onto an earlier view.
The existing wait API retains its event refusal; C1d separately proposes ownership
of returned private values for the new API. Returned role observations are cumulative
for that view, not repeated events to add every time a search probe runs. Actual
computation/probe counters still count every performed evaluation and refusal;
conservation must not hide search work from budgets or reporting.

The existing generic `AccrualHook` interface does not establish purity or identity
of arbitrary callbacks. Before this API is buildable, its production adapters must
enumerate their exact hook recipes, pinned catalog dependencies and captured inputs,
and establish that all gameplay mutations stay in the supplied scratch state.
Do not hash a function pointer, accept a caller-provided label as proof, share a
mutable closure between branches, or silently omit Guild/Commons/Prestige work.
Current `firstHourLifetimeHook` and the real stock-role counterfactual are concrete
adapter populations, not proof for every future hook. RP-654's source-version
correction is implemented independently; it does not implement this protocol.

The actual closed replay chain can return Guild/Commons events without a player
action. Today's `SimulateAdvance` refuses such events. The bounded online/offline
`TestAxisCalculationEventOwnershipResearch` executes both sides: a fresh private
Company19 evaluation plus the real chain yields ordered `guild_tithe_accrued` and
`compact_sampled` values, positive lifetime/progress/sample state and unchanged
origin bytes. Actual `SimulateAdvance` refuses; no-hook controls advance but lose
the hook effects. This is RP-655's two-profile API constraint, not a new production
API, event publisher, live bug or whole-domain proof. C1d proposes the distinct
ownership contract; the current refusal is not removed or filtered away.

**Search identity must cover every reuse path.** Proposed identity is the canonical
complete origin plus source save version, constants identity, evaluation mode and
the validated producer/hook/ablation context. Derived digests are private lookup
keys for trusted values, not proof that a producer is authentic. No rounded-rate
cache, mutable callback address or search path substitutes for that identity.

The canonical consumer rules for beam equality, completion-cache keys, dominance
and temporal search live in [Accrual-aware Harness Search](accrual-aware-harness-search.md).
That draft proposes the Relevance R10/R11 amendments; this primitive RFC does not
silently adopt or implement them. Keep current reports/guards and original failing
cases intact until the corresponding accepted contracts authorize their migration.

### C1d — Proposed closed hook adapters and private events (unadopted)

The new API takes a closed recipe selected by a production-owned factory, not an
arbitrary `AccrualHook` and a caller-asserted identity. Exact proposed recipe tags:

| Tag | Existing implementation and bound inputs | Scope |
|---|---|---|
| `none` | No hook. | Only a caller whose accepted simulation contract requires none; not a neutral substitute for required work. |
| `first_hour_lifetime` | Existing Prestige `AccumulateLifetimeValue` helper with the pinned value resource, matching `firstHourLifetimeHook`. | Preserve the first-hour runner's separate attendance/session handling; this is not the full Prestige hook. |
| `faction_stock` | Existing `faction.AccrualHook`; faction catalog and catchup ceiling resolved from the pinned bundle. | The stock-role simulation primitive, not full gameplay-hook parity. |
| `closed_replay` | Existing `closedReplayAccrualHook`: Prestige → Faction → Guild → resolved Commons. | Complete existing replay chain; freeze Commons weight and the replay-owned external context, never a live weight resolver. |

Reject unknown recipes, unavailable catalog dependencies or inconsistent captured
inputs. The factory owns dispatch to these exact implementations; a raw callback
cannot impersonate one by supplying a matching string. Existing legacy simulation
functions retain their current interface and failure semantics until their explicit
caller migration is accepted. Unit fault hooks on those functions are not a recipe
registration mechanism. Adding a future hook requires a reviewed recipe amendment.
Build catalog resolvers through the existing loaders from the bound artifact bytes;
do not accept an arbitrary caller-supplied implementation of the Commons policy
interface as proof of the recipe. Production can call the Prestige helper directly;
it must not import the harness package to construct the lifetime-only adapter.

For the new `ProjectCalculation` only, hook-returned `EventWrite` objects are ordered
private **values**, attached to that detached projection and its origin/context/
endpoint identity. Preserve kind, schema and complete payload without filtering.
They have no newly invented intent, event ID, revision or outbox write. Copy payload
buffers so a caller cannot alter the origin or a sibling result. The hook factories
must not carry a Store, publisher, network resolver or mutable external observer.
Never publish or persist a projection's events, including when its resources look
identical to a later settlement. Discarding a projection discards only these values,
not any real transaction's required events.

The real applied action recomputes from the origin through its existing closed
transition, binds that transition's actual intent identity and writes only its own
ordered events atomically with its receipt/state. Do not concatenate earlier
projection events or use their payloads as a substitute for the transition result.
Conflicting retry or refused action cannot leave an event fragment. Preserve current
service/Store idempotency; this proposal does not add a second settlement table.
Implementation acceptance must compare full projected/settled hook effects and event payloads,
and prove no projection delivery plus once-only actual commit/retry on real SQL.
The two-profile RP-655 test does not supply that Store/outbox proof.

### C1e — Proposed private context encoding and identity (unadopted)

Use a strict, versioned internal value, not a new Company/Founder field or public
artifact. Exact proposed root fields are `v`, `company_version`, `company`, `operation_company`,
`founder_version`, `founder`, `constants_hash`, `next_constants_hash`, `owner`,
`stage`, `mode`, `replay_inputs`, `simulation`, `diagnostic`. `v` is1. Company is the complete
typed state at its admitted source version. Founder and its version are both null
when not an input; otherwise retain its complete admitted typed state, not just a
partial counter list. Exit/continuation context binds the actual next bundle hash;
otherwise `next_constants_hash` is null. Each non-null hash must resolve to the
already validated pinned artifacts; a digest does not authenticate arbitrary bytes.

- `owner.kind == logged`: `owner` contains exactly `kind` and the existing complete
  `command` tuple; `replay_inputs` is the existing complete admitted versioned
  envelope, `operation_company` is the complete original command pre-state at the
  same Company version, and `simulation`/`diagnostic` are null. Reconstruct the requested stage
  entry from that pre-state through its actual preceding Guild/catchup/schedule
  steps and require exact equality with `company`. Command, mode, constants and
  recorded catchup coordinates must agree at their respective boundaries, not by
  incorrectly comparing a post-catchup cursor to the old absence start. Retain
  Founder carry and active-play evidence; no live lookups during reconstruction.
- `owner.kind == simulation`: `owner` contains exactly `kind`, `scenario_id`,
  `scenario_version`, `scenario_hash`, `policy_id`, `policy_version`, `policy_hash`,
  `seed`, `session_index`, `revision`. Bind the existing loaded scenario/policy
  identities and selected run seed, not IDs/versions alone. Seed is the existing
  canonical unsigned-decimal seed string; session index is null before assignment
  and otherwise an exact nonnegative integer. `replay_inputs` is
  null and `operation_company`/`diagnostic` are null. `simulation` contains exactly `external_contributions`, `commons_weight_ppm`,
  `guild_settlement_batch`, `route_context_version`, `absence_span`, `mask`, `hook_recipe`.
  Reuse the existing replay-accrual grammar/validation for the first four fields.
  `mask` contains exactly `generator_ids`, `upgrade_ids`, `removed_generator_ids`,
  `removed_upgrade_ids`, `removed_action_ids`: sorted unique valid IDs with the
  existing ablation semantics. `hook_recipe` is one C1d tag, not arbitrary code.
  `absence_span` is null online; offline it contains exactly `from_ms`, `to_ms`,
  the actual policy-owned absence start and return boundary. Its start matches the
  stage origin, endpoints lie within it, and internal probes cannot change the span
  or allocate another allowance. Validate it from the selected session history,
  not from a client-supplied mode or a convenient cut.
- `owner.kind == diagnostic`: the exact C1g profile requires a non-null diagnostic payload;
  `operation_company`, `founder_version`, `founder`, `next_constants_hash`,
  `replay_inputs` and `simulation` are null. This is not a synthetic harness run.
- `stage` is exactly `offline_catchup`, `ordinary`, `simulation_wait` or
  `diagnostic_window`; `mode` uses
  the existing online/offline union. Simulation owners use `simulation_wait`; logged
  owners use one of the first two under the existing replay envelope rules.
  Diagnostic owners use only `diagnostic_window`/online under C1g. A stage
  origin is captured after preceding semantic mutations as
  C1b requires: retained clearing evidence does not mean applying that batch again
  to an already-cleared origin. A caller cannot change stage/mode to renew a cap.

Unknown/extra fields, invalid null combinations, unsupported versions, invalid
state/context or disagreement with pinned inputs refuse before constructing a
handle. Byte decoders also refuse duplicate keys and trailing/multiple JSON values
before canonicalization; do not let last-key-wins parsing choose the identity.
No raw computed rate, function pointer or ambient clock is encoded. Restore
the original source versions; rederive catalog-owned runtime facts such as
`FactionStockResource` from state plus pinned catalogs and reject disagreement,
rather than persisting a second authority. Artifact/Company/Founder maps and input
buffers must not alias caller-owned mutable data.

Encode the validated logical value to match `canonicalJSONString` in
`client/src/replay.ts`: canonical state Decimal strings, exact integer coordinates,
byte-sorted object keys and semantically ordered arrays, without report indentation,
HTML-escape differences or a trailing newline. Hash the bytes with SHA-256 prefixed by the
UTF-8 domain separator `cloud-clicker/accrual-origin/v1` followed by one zero byte.
That private origin/context digest is independent of a projection's endpoint; reuse
keys add endpoint/visible-state/revision/budget coordinates as C1c specifies. Input
spelling or JSONB key order is not identity; an actual field/order change is.
This memory-only encoding is not approved for saving, public export or publication
of private state. Runtime/size bounds and strict codec/refusal/parity tests remain
mandatory for implementation acceptance; no new limit is invented from this schema proposal.

### C1f — Harness search follow-up (split; unadopted)

The proposed temporal-search algorithm and RP-656 evidence now live in
[Accrual-aware Harness Search](accrual-aware-harness-search.md#h3--temporal-affordability-search-proposed).
That successor owns Relevance R10/R11 reuse/search and affected consumer migration.
Both drafts remain unaccepted; the full AC6/accounting closeout still requires the
applicable consumer proof. This split does not waive any original criterion.

### C1g — Proposed content-dynamics diagnostic owner (unadopted)

Do not force this real caller into a fake scenario/policy owner. Exact proposed
owner is `{kind: diagnostic, profile: content_dynamics_active_play,
profile_version: 1}`. Its `diagnostic` object contains exactly `input`, `branch`,
`window_start_attended_ms`, `duration_ms`, `target_resource_id`,
`contribution_sampling`, `contributions`, `hook_recipe`.

- `input` has exactly `founder_id`, `run_seq`, `effect_row_id`, `target_generator_id`,
  matching the existing `ContentDynamicsActivePlayInput` and its actual validation.
- `branch` is `claimed` or `control`. `contribution_sampling` is exactly
  `window_start`; `hook_recipe` is exactly `none`, matching the current diagnostic.
  Neither tag permits omission of hooks in a different caller's contract.
- Reconstruct the actual `newContentDynamicsCompany` v18 seed, one owned target,
  natural first draw and real scheduled pending opportunity from input plus pinned
  artifacts. Refuse a changed expected effect/selected target as the current caller
  does. The control is its pre-claim copy; the claimed branch executes the existing
  opportunity claim with the caller's fixed diagnostic intent
  `018f6b7c-9abc-7def-8abc-0123456789ab` and expected revision1.
  Extract/reuse that production-owned prefix during implementation rather than
  maintain a second constructor or hand-author a pending opportunity.
- Require exact complete `company` equality with the selected reconstructed branch.
  Window start is the actual spawn's attended coordinate; duration is the pinned
  effect's existing duration; target resource is the actual generator production
  resource. `contributions` uses the existing ordered replay-contribution grammar
  and must exactly reproduce `activePlayContributions` at that window start followed
  by ordinary contribution assembly. A forged factor, missing row or later sampled
  context cannot pass simply because a digest was recomputed.
- Projection endpoints lie within the original diagnostic window. Recompute the
  whole origin-to-endpoint interval using this fixed validated contribution context;
  do not sample the expired buff at the endpoint, reset the origin at a cut or
  prorate a new window. Only scratch state changes. Profile1 preserves the actual
  Company18 diagnostic; it does not enable held axis mechanics or a new schema floor.

This preserves a measured distinction, not a gameplay exception. RP-657's registered
epoch8 natural production draw/claim yields the existing5000ms diagnostic amounts:
claimed35.035, control5.005, bonus30.03. Running the real endpoint schedule/evaluator
on fresh copies instead yields5.005 in both branches because the buff expires at
that endpoint. Both origins remain unchanged. The current diagnostic intentionally
samples its window start; replacing it with logged endpoint selection would erase
the measured bonus, not fix rounding. Logged callers retain their existing endpoint
policy. Neither policy is silently changed or generalized by this proposal.

The bounded test reconstructs the current prefix and checks actual caller outputs;
it is not an implemented private owner factory/codec. Implementation acceptance
must include both branches, all admitted window-effect kinds, wrong
input/branch/window/contribution refusals and actual registered
content-dynamics comparison without restamping its historical golden. Keep profile
construction private; no public schema, replay event or saved payment is introduced.

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
settlement workloads, general temporal workloads, general Go/TS parity and full-domain conservation remain owed.

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

The subsequent shared `temporal-replay-research-v1.json` observation extends
these same42 cells to the actual TypeScript `applyLogged`, not a scalar prototype.
The hash-bound bundle and admitted pre-state reproduce complete canonical Company,
receipt and ordered-event bytes in Go/TS; restoring each result preserves it.
Positive lifetime accrual, exact surviving buff identities and the existing long-
absence span are retained. Both runtimes refuse the42 forged attended coordinates
and42 omitted schedules without state mutation. This closes bounded temporal replay
compatibility for this population only. General workloads, general external/Guild context
parity, SQL/live capture, origin lifetime and a viable conservation algorithm remain
unproved; the original27 partition failures are unchanged. No runtime semantics or
new accounting format is adopted by this observation.

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

The shared `external-replay-research-v1.json` observation now carries all twelve
external/Guild cells above through actual Go and TypeScript logged replay. Complete
canonical Company, receipt and ordered events match, and restored results round-trip
in Node, Chromium and WebKit. The six nonempty settlement cells additionally retain
omission with equal cash but different state, duplicate refusal without mutation,
and late unaffordable-purchase rollback. Three forged expected outputs are rejected
by the consumer comparison. This closes that bounded compatibility question only;
fixed membership/batch lookup is not live provider/SQL evidence. General workloads,
origin lifetime, absence semantics and the viable accounting algorithm remain open;
the original27 partition failures are not repaired or redefined.

#### Live Guild capture complements the fixed-input populations

The initial2026-10-15 `b2d77add` `stored_guild_settlement_replay` supplement in the existing
composed Postgres/socket/clearing test captures one actual online transition.
HTTP account/Guild creation and membership feed three worker clearings; an
authenticated manual action consumes the real nonempty batch. The fixture's
faction/stock arrangement remains controlled, not naturally acquired gameplay.
The committed pre-head, original canonical command, replay inputs, receipt and
ordered Company events are read from SQL. With only those frozen inputs and the
original pinned catalog, actual Go and TypeScript logged replay reproduce the
complete saved Company, receipt and events. TS also restores its complete result.
Omitting the batch loses settlement state; duplicating a clearing refuses without
mutation in both runtimes. Private captures travel in memory, not public artifacts.

This closes live-provider/SQL capture and cross-runtime fidelity for that one
current-content online Guild profile only. It does not cover the twelve earlier
profiles or positive idle-production fidelity (this consumer buys no generators),
Clout-v19 activation, arbitrary external/temporal producers, independent
process/host restart, or the accounting algorithm. The Clout owning log records
execution and review; C1/C1a and the lower-domain write-set conflict remain open.

The following positive-producer extension retains that same current-content
online profile but adds exactly one controlled owned beige tower in the existing
setup write. The old zero-generator population fails the new idle/lifetime check.
The actual post-clearing HTTP action now settles idle production through the
Prestige hook; complete SQL state, receipt and ordered events match Go/TS replay.
Removing only that producer in a fresh replay origin preserves clearing consumption
but removes the lifetime increase in both runtimes. Omission/duplicate controls
remain. This proves positive idle/hook fidelity for this one admitted profile,
not purchase/earning provenance, Clout-v19 behavior, a general context algorithm
or conservation. The owning log retains the old population and final gates.

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
does not establish a production storage bound. The complete-context reconstruction
and cost proof remain implementation gates. These146 profiles do not authorize an
optimized minimal-context schema in place of the explicit C1e proposal.

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
admitted domain. A proposal committing those intermediate heads needs a viable
accounting/ledger algorithm before RFC acceptance; an origin lifetime alone does
not resolve it. The recommended operation-local route instead makes no such
intermediate commits, retaining the counterexample and final-settlement proofs.
Do not silently restrict the domain,
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
publishing hook outputs from every internal projection would turn cuts into additional
semantic events. C1d's unadopted private-value computation is distinct from publication.
The accepted repair must bind elapsed/production/banking/progress and resource
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

The ledger remains the atomic resource authority. Under the unchanged receipt
contract, every committed balance must have a canonical nonnegative accrual receipt
whose delta re-adds to that exact balance from its immediately previous committed head.
Detached views are not commits and their origin-relative diagnostics cannot be
applied as incremental grants to the live ledger.
No direct balance overwrite may bypass minimum/hardcap/transaction checks. Saturation
must not bank hidden overflow that becomes spendable after a later debit. Below-cap
quantities that round visibly to the cap require an explicit, compatible rule.
The optional cumulative representation below would explicitly amend that incremental
receipt requirement, not claim to satisfy it simultaneously.

#### A cumulative receipt is an explicit alternative, not an additive-delta fix

`axis_cumulative_receipt_research_test.go` tests a distinct representation against
the same20 lower-domain cells and both cap/debit modes. Restore the trusted encoded
origin, reconstruct contributions with the pinned catalog and call actual Evaluate
from that origin to each requested time. Its existing ledger receipt is relative
to the **origin**, not the preceding visible projection. All20 cumulative receipts
reproduce their one-shot targets, including the nine cells whose previous-head
additive deltas remain inadmissible. Neither the floor nor those refusals changed.

The reference object binds origin/catalog identity, previous/full projected Company
identities, expected server time/mode and the complete EvaluationResult. Verification
reconstructs both prior and requested states from the origin; equality of balances
alone is insufficient. Fourteen altered-field cases plus a non-forward interval
return no candidate. Both cap/debit profiles reject an old-origin proof even when
its previous-head hash is recomputed to match the post-debit state. Starting a new
origin after the debit gives the existing literal payouts, not reclaimed overflow.
The hashes identify already trusted inputs; they do not authenticate a producer or
establish persisted provenance. There is no proposed public proof endpoint or wire
codec in this experiment, and no candidate is installed into a live caller state.

This result narrows the engineering options:

- If the ruled cuts are internal work only, defer settlement to the existing closed
  transition. Do not introduce saved intermediate heads or new receipts just because
  a reference projection exists. C1/C1a and full closed-boundary proof remain required.
- If the ruled cuts must be saved, the demonstrated pair cannot retain the requirement
  that every receipt's Delta adds to its immediately previous head. A separately
  adopted **origin-relative cumulative receipt** is a candidate amendment: keep
  ordinary action/debit receipts unchanged; bind the origin, before image, projected
  result and once-only settlement context explicitly. Never label its cumulative
  Delta an incremental grant or apply it to the current visible balance.

For that second option, the production ledger must own an atomic, expected-head-bound
projection operation which independently verifies the trusted origin calculation and
all resource scope/minimum/hardcap constraints before installing any result. A caller
assigning a projected ledger or abusing RestoreLedger as a mutation shortcut would
not satisfy it. The complete new receipt grammar, origin/context codec, replay identity,
reset points, hook protocol, corruption refusal and historical activation still need
specification and actual Go/TS/SQL proof. This Go-only reference verifies expressive
sufficiency for its22 profiles, not that production mutation protocol or a general
conservation algorithm. It does not resolve arbitrary origins/external/temporal contexts,
propose raw-float persistence, authorize a receipt change or repair the original27 cases.

The scalar prototype's 512-byte envelope is not a production storage/CPU bound. Establish
bounds from the actual proposed field set and measured full-state workloads; guard
exhaustion must invalidate the operation, not truncate history or reset rounding.

#### Temporal projections need not be monotone — RP-652

`TestAxisCumulativeTemporalHeadMonotonicityResearch` executes eight pairs around
the unchanged production/building buff expiry boundaries, online/offline, using
the real active-play schedule and evaluator. Every origin-relative receipt is a
valid positive accrual from its own origin. Nevertheless every later target is
lower than the preceding target: for the online production case, cash projects
from `1.0008049992e4` at999ms to `1.000115115e4` at1000ms. All eight matched no-buff
controls increase, independent literal payout calculations agree, and the actual
ledger refuses each representable negative difference without mutation.

This follows the existing endpoint-selection behavior described in C2: when a
buff expires, the endpoint no longer supplies it to that one-shot calculation.
The producer's temporal context therefore changes. This is not a counterexample
to conservation under an unchanged producer context, a defect in those one-shot
payouts, or evidence that live players received and then lost these two payments.
The test projects fresh copies; it neither persists cuts nor invokes the earlier
constant-context reference verifier as a temporal mutation API.

It does rule out naively promising all three of the following for this population:
every intermediate **saved payment** matches its independent endpoint one-shot
projection; subsequent positive accrual never debits that balance; the existing
endpoint payout semantics remain unchanged. A cumulative receipt's positive Delta
from the origin does not imply a positive increment from the previous saved head.
This is separate from the lower-domain unrepresentable-positive-increment finding.

The selected protocol must define semantic context/settlement boundaries before
installing cumulative targets. Automatically settling/reanchoring at buff expiry
would change existing one-shot payout behavior; it is not an authorized rounding
repair. Noncommitting internal projections do not make two payments and are not
excluded by this result, but their complete closed-boundary contract/proof and
C1/C1a adoption remain owed. No ledger overwrite, proration, negative production,
criterion weakening, origin lifetime or production implementation is adopted here.

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

Under the recommended operation-local route, no new Company/Founder field, save
migration or replay-envelope field is proposed. Existing persisted `Evaluate`,
`ApplyLogged` and Exit histories retain their pinned one-shot behavior; the private
C1e encoding must not be slipped into those schemas. A new calculation interface
and changed consumer dispatch need their own reviewed caller bindings. The harness
successor owns new scenario/policy/report dispatch, not a hidden switch on an old
identity. Actual runtime changes still obey kernel-version rules; do not invent a
false version bump or assign future version numbers before the landing order exists.

## Proposed implementation acceptance criteria

These prove feature completion after an accepted specification is implemented;
they are not claims of existing passes or requirements to pretend the code already
exists before RFC acceptance.

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
8. All affected harness consumers complete the accepted linked successor's criteria.
   Splitting specifications does not allow primitive closeout to masquerade as full AC6.

## RFC readiness versus implementation gates

Before RFC acceptance, agree the operation-local recommendation, complete exact
production caller bindings/context compatibility, review C1g's diagnostic binding, and obtain
designated normative review plus Marco's acceptance. The wording delegation alone
does not do that. The failed difference-write/durable/cumulative alternatives are
not silently adopted; no reviewer or owner can approve away their counterexamples.

The linked harness draft separately owns its complete boundary/predicate algorithm,
diagnostic owner and versioned caller/report dispatch. Those engineering gaps must
close before that specification is accepted; accepting this primitive would not
accept the successor, remove current constructor holds or finish AC6.

Strict codec/adapter reconstruction, Go/TS parity, full admitted-domain contexts,
real Store commit/rollback/retry, actual workload/storage measurements and affected
CI are executable implementation gates above. Do not demand a completed new codec
or SQL path as proof an unimplemented RFC is already shipped. Equally, do not defer
an unspecified algorithm to implementation or substitute owner approval for any
required proof. Keep full-state bounds measurable and fail-loud; no ceiling increase
or population reduction is authorized here. No implementation begins from draft status.

## Deviations from design

No reward or numeric-law deviation is adopted or implemented. Explicit interval/origin lifetime
extends the implementation contract. A durable conservation format, if selected, is not an
already implied save field. The optional cumulative-receipt direction would amend the existing
incremental receipt law; it is not behavior-compatible with that law and requires a complete
contract, explicit adoption and independent normative review before production work.

## Changelog

- 2026-10-15: split harness search/reuse/caller migration into its own linked draft
  under RFC-0000. Recommend the operation-local route without adopting it; distinguish
  RFC readiness from executable implementation gates and retain full AC6 dependencies.
- 2026-10-15: execute18024 actual temporal endpoints over twelve profiles; a boundary-
  aware reference agrees with exhaustive search while unsplit search misses six
  early windows. Specify the proposed chronological-regime algorithm and outstanding
  completeness/predicate/work proof. No runtime, report, timing or payout change.
- 2026-10-15: execute the two-mode event-bearing chain/wait constraint; specify
  proposed closed adapter recipes, private event values versus once-only real
  settlement, and a complete versioned private context/identity shape. Legacy APIs,
  events, persistence and reports stay unchanged; draft is not implementation authority.
- 2026-10-15: specify the proposed calculation/fork/projection/action lifecycle and
  complete-state hook ownership. Include completion-cache and dominance reuse, not
  beam deduplication alone; name exact adapter/encoding/temporal-search work still
  needed. Draft only; no existing API, R10/R11 rule, test or report is changed.
- 2026-10-15: propose concrete operation-local command/context ownership, existing
  catchup/action stage ordering and settlement/rollback boundaries. Explicitly retain
  the mismatch between the original mutating-`Evaluate` partition diagnostic and a
  noncommitting projection API. No test rewrite, receipt change, lifetime adoption
  or production implementation; acceptance and designated review remain outstanding.
- 2026-10-15: record Marco's explicit narrow wording delegation and reconcile C1/C1a
  with the Clout body. Accounting, exact identity/settlement, origin lifetime, acceptance
  and independent review remain required; no production authority follows.
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
- 2026-10-13: add actual42-cell Go/TypeScript temporal replay compatibility and84 refusal
  controls per runtime. Still no conservation repair, interval/lifetime adoption or SQL proof.
- 2026-10-15: record actual twelve-cell external/Guild Go/TypeScript replay compatibility
  and six omission/duplicate/rollback control sets. Draft evidence only; no accepted
  contract, accounting implementation, origin lifetime or SQL/live-provider claim changed.
- 2026-10-15: add one real composed Guild/SQL input capture with complete Go/TS replay
  and omission/duplicate controls. This complements rather than relabels the fixed-input
  populations; no accounting/lifetime adoption or conservation completion follows.
- 2026-10-15: extend the live profile with one controlled producer and an executed
  zero-producer counterfactual to distinguish idle lifetime accrual from clearing.
  Original zero-generator evidence stays bounded; no production accounting change.
- 2026-10-15: distinguish an executed origin-relative cumulative receipt alternative
  from the impossible lower-floor incremental delta. All20 retained cells and both
  cap/debit modes pass the reference proof checks, including alteration/stale-origin
  refusals. This is an unadopted receipt-contract alternative, not a production
  ledger mutation, full conservation implementation or permission to alter AC6.
- 2026-10-15: retain eight temporal decreasing-target counterexamples plus eight
  no-buff controls. Separate independent endpoint projections from monotone saved
  payments and fixed-context conservation; no live-loss claim or policy adoption.
