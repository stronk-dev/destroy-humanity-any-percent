# RFC: Doctrine Intents & Compute Credit Spend

- **Status:** accepted in principle; implementation blocked on D7-D10 (the D1-D6 rulings leave
  burst arithmetic, seed/gate data, activation sequencing, and the exact wire grammar unresolved).
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-03
- **Design refs:** `design/10 §doctrine` (doctrine trees — Age-Up branching, one building type nothing else gets, combos with faction), `design/02 §9` (Compute Credits — banked offline time as chosen-moment acceleration; the banker playstyle), `design/09 §doctrine events`
- **Research:** `events-playstyles.md §Time banking` (Compute Credits as "unused reserved capacity"; legibility warning — loud HUD affordance + optional auto-spend; every playstyle needs an anchor), routes docs (the doctrine-pick ordering the route registry already gates on)
- **Depends on:** Gate Predicates + Routes (implemented — routes already gate on `DoctrinesByTransition`; doctrine-pick is the missing intent), Production + Run Genesis (implemented — both are intents inside `ApplyLogged`)
- **Owner ruling honored:** breadth-first — the two named-but-undrafted intent contracts; the doctrine TREES and Compute-Credit spend TARGETS are content.
- **Planning:** `planning/doctrine-and-compute-credit/` (once implementing)

## Summary

Two named contracts the route/economy foundations already reference but never got: the
**doctrine-pick intent** (routes gate on `DoctrinesByTransition` — the picking has no intent) and
**Compute Credit spend** (offline time banks as Compute Credits — the spend has no mechanic). The
doctrine-pick intent is a closed addition inside the replay boundary; the Compute Credit spend adds a
persisted Company-run acceleration-burst state (D4) that requires a Company save-version bump. A
`doctrines` artifact is added (D1); doctrine EFFECTS and Compute-Credit auto-spend are named
successors (D3/D5).

## Specification

### DC1 — Doctrine-pick intent

At each transition where doctrines branch (design/10: doctrines pick at tier transitions),
`pick_doctrine {transition_id, doctrine_id}` — C1 envelope, evented (`doctrine_picked`),
replay-logged. Valid exactly once per transition (idempotent by transition), the doctrine catalog-
declared for that transition, mutually-exclusive within a transition (the branching — picking one
forecloses the siblings for the run). Writes `DoctrinesByTransition` (the save field routes
already read — closing the loop the route registry left open). **The doctrine-pick ORDERING
constraint** (the routes RFC's deferred item: same-boundary doctrine routes need pick-ordering) is
resolved: a doctrine is picked BEFORE any route gated on it can execute (loader-validated route
predicate ordering: `gateTier ≥ doctrine-pick tier`). Eligibility (D2): a pick is legal only while
Company tier == the transition's `source_tier`, before its gate is crossed, and while the transition
has no recorded choice; a repeated intent ID replays, a new intent ID for an already-picked
transition rejects `doctrine_already_picked`. **Doctrine EFFECTS (contribution bundles, tree
unlocks) are NOT in this RFC (D3) — this RFC persists and exposes the CHOICE only; a successor
content RFC owns effects and mints them onto new runs. Route predicates consume the persisted choice
immediately.**

### DC2 — Compute Credit spend

Compute Credits already accrue (`ComputeCreditMS` is save state — integer ms of banked offline
time). This RFC adds the spend: `spend_compute_credit {amount_ms, target}` — C1, evented,
replay-logged. `target` (closed union): `accelerate` — spend `amount_ms` of banked credit to activate
a **persisted acceleration BURST** (D4-ruled: NOT an instant time-warp): the burst multiplies
production rate by the catalog `burst_speed` for a `boosted_duration` (derived from `amount_ms` +
catalog, capped by `burst_max_duration_ms`), consumed as production accrues; **Exit resets any active
burst** (the banked `ComputeCreditMS` persists per its existing scope; only the in-run burst is
consumable). Later targets by RFC. Spend debits `ComputeCreditMS` (it IS a spendable currency —
unlike the moral meters; the banked-time currency), integer-exact, cannot go negative (typed
rejection). **Legibility (the Idle Spiral warning):** the spend is a loud, explicit affordance and
the wire snapshot exposes the balance. **Auto-spend is deferred to a successor settings RFC (D5) —
this intent is manual-only; the HUD does not claim an auto-spend policy exists.**

### DC3 — Replay & determinism

Both intents are state mutations inside `ApplyLogged`. Doctrine pick writes a save field. Credit
spend debits `ComputeCreditMS` and activates a **persisted Company-run burst state** (D4) — so this
RFC requires a **Company save-version bump** for that state (correcting the earlier "no save bump"
claim; this is a second Company-version claimant alongside Active-Play Buff Windows — sequence or
co-activate per the meters+achievements pattern at implementation). The resolved arm records the
burst activation; the burst multiplies rate over the SAME bucketed production accrual (the
provision-grid partition-invariance applies — a boosted interval is the ordinary accrual with a
rate multiplier, capped by the existing horizon rules). The F1 online-horizon precondition applies —
a large accrual must not brick the stream. Nothing new in `replay_inputs` (burst state is
save-derived; `now`/attendance already flow through the existing resolved arm).

## Acceptance criteria

1. Doctrine pick: once-per-transition, mutually-exclusive within transition, catalog-validated,
   writes `DoctrinesByTransition`; a route gated on an unpicked doctrine rejects; pick-ordering
   (`gateTier ≥ doctrine tier`) loader-validated — closing the routes RFC's deferred item.
2. Compute Credit spend: debits exactly, cannot go negative (typed rejection), acceleration
   applies the warped interval byte-parity Go/TS; per-spend cap enforced; a warp exceeding the
   online horizon is handled per the F1 precondition (no brick).
3. Legibility: the wire snapshot exposes the credit balance + auto-spend policy; auto-spend as a
   founder setting round-trips.
4. Both intents evented + replay-logged; sequential corpus rows in both kernels.

## Open questions

- Doctrine-tree content (the actual doctrines, their unlocks, faction combos) — design/10 content,
  later RFC; this ships the pick intent + ordering law.
- Additional Compute Credit spend targets (beyond `accelerate`) — later RFCs on the closed union.

## Acceptance-review blockers (2026-08-05)

The implemented fields and replay boundary are suitable substrate, but the two additions are not
yet executable contracts. Codex must not choose the following mechanics in code.

### D1 — No doctrine catalog or activation authority is specified

`pick_doctrine` must validate a doctrine against something immutable, but the RFC names neither an
artifact nor an exact schema, resolver, activation boundary, or cross-artifact check against Routes.
The current Route catalog is not a complete doctrine catalog; it mentions only doctrines used by
routes. "Doctrine-tree content later" therefore leaves the intent with no authoritative choice set.

**Proposed contract:** add a hash-pinned `doctrines` artifact whose first schema contains only the
mechanical identity and ordering rows needed by this RFC: transition ID, source tier, gate ID, and
the closed doctrine IDs available at that transition. Adding the artifact is an epoch mint and
activates only for new runs under the existing activation-boundary law. Both bundle loaders require
every doctrine-bearing Route predicate/exclusion value to resolve to a doctrine row. Contribution
bundles and tree unlocks remain absent from schema v1 and arrive only through a successor RFC.

### D2 — Pick eligibility and same-boundary gate ordering are ambiguous

`gateTier >= doctrine-pick tier` proves only temporal possibility. It does not say when a pick is
legal, whether every declared transition blocks its gate, or how the separate `pick_doctrine` and
`cross_gate` commands order at the same boundary. "Idempotent by transition" also conflicts with
the existing intent-ID idempotency contract: a second command with a new intent ID must have a
defined rejection, not silently replay the first command.

**Proposed contract:** a pick is legal only while Company tier equals the transition's source tier,
before its gate is crossed, and while that transition has no recorded choice. `cross_gate` rejects
`not_eligible/doctrine_required` when its active doctrine row has no pick. The player submits the
pick as a prior committed intent; route predicates then evaluate the committed choice. A repeated
intent ID replays normally; a new intent ID for an already-picked transition rejects
`not_eligible/doctrine_already_picked`.

### D3 — Doctrine effects are simultaneously promised and deferred

DC1 says a pick applies contribution bundles and doctrine-tree unlocks, while the open-questions
section defers the actual doctrines, unlocks, and faction combinations. No effect union, slot
binding, lifetime, or formula-artifact representation exists, so implementing that sentence would
require invented mechanics.

**Proposed contract:** this RFC persists and exposes the choice only. Remove the effect sentence
from DC1. A successor content RFC expands the pinned artifact and owns contributions/unlocks; its
mint activates those semantics on new runs. Existing route predicates may consume the persisted
choice immediately.

### D4 — `accelerate` has no exact time/cost equation or cursor semantics

The existing catalog has `burst_speed` and `burst_max_duration_ms`, but DC2 never defines whether
`amount_ms` is credit consumed, wall duration, or simulated duration; how `burst_speed` converts it;
or whether acceleration is instant or persists across future wall time. An instant synthetic
interval cannot both use absolute provision buckets and leave `evaluated_through` honest without a
separate cursor. A persistent burst requires state and a save migration, contradicting the claimed
"no save-version bump" scope.

**Owner ruling required:** choose the precise equation and lifecycle. Recommended shape is a
persisted Company-run burst state activated by the intent and consumed by later real attended
intervals: credit cost, boosted wall duration, bonus production, cap, partial consumption, Exit
reset, and offline behavior must all be stated with integer/Decimal rounding order. If instant
conversion is preferred, the RFC must instead define the independent virtual-time cursor and its
fixed-grid semantics explicitly.

### D5 — Auto-spend has no owner, persistence, command, trigger, or replay contract

There is no existing founder-settings gameplay store. A Founder-scoped toggle that changes Company
production also crosses the Founder/Company replay boundary; a client-only toggle cannot be
authoritative. The RFC requires round-trip storage while also claiming no save bump.

**Owner ruling required:** either (a) defer auto-spend to a successor settings RFC and keep this
intent manual-only, or (b) specify its exact scope, mutation command, persisted shape/version,
resolved replay input, trigger rule, and race behavior with manual spend and Exit. The HUD may
expose the balance now without claiming the policy exists.

### D6 — The supposedly closed wire additions are not enumerated

The server event registry, Go parser, TS parser, receipt encoder, and replay fixture all require
exact grammars. The RFC names only the intent kinds and event kind; it does not enumerate exact
keys, safe-integer bounds, event payloads, receipts, or the existing closed rejection taxonomy
details.

**Proposed contract:** enumerate both intent objects, applied receipts, and event payloads
byte-for-byte; use positive safe-integer `amount_ms`; reuse only existing rejection reasons with
closed detail values; add both event kinds to the Go/DB registries; and add shared sequential
Go/TypeScript fixtures covering applied, repeated-ID replay, malformed exact-key input, every
semantic rejection, and the chosen burst-boundary cases. No new replay input is needed only if D4's
chosen semantics truly read solely from the request, pinned bundle, and committed state.

## Changelog

- 2026-08-03: created (draft) — the two deferred intent contracts; doctrine-pick closes the
  routes RFC's ordering item; Compute Credit spend gives banked time its mechanic.
- 2026-08-05: Codex acceptance review — blocked on D1-D6; proposed contracts recorded without
  implementing doctrine content, burst lifecycle, or settings semantics.

## Owner rulings on D1-D6 (2026-08-05)

- **D1 — accepted.** Add a hash-pinned `doctrines` artifact, schema v1 carrying only the mechanical
  identity/ordering rows: `{transition_id, source_tier, gate_id, doctrine_ids[]}` (closed IDs per
  transition), sorted. Epoch mint, activation-boundary law (new runs only). BOTH bundle loaders
  require every doctrine-bearing Route predicate/exclusion value to resolve to a `doctrines` row
  (the cross-artifact integrity check). Contribution bundles + tree unlocks are NOT in v1 — a
  successor content RFC expands the artifact.
- **D2 — accepted.** A `pick_doctrine` is legal only while Company tier == the transition's
  `source_tier`, BEFORE its gate is crossed, and while that transition has no recorded choice.
  `cross_gate` rejects `not_eligible/doctrine_required` when its active doctrine row has no pick. The
  pick is a PRIOR committed intent; route predicates evaluate the committed choice. Idempotency: a
  REPEATED intent ID replays normally; a NEW intent ID for an already-picked transition rejects
  `not_eligible/doctrine_already_picked` (this resolves the "idempotent by transition" vs intent-ID
  conflict — the transition is write-once, the intent-ID contract is unchanged).
- **D3 — accepted; DC1 reconciled.** This RFC persists and exposes the CHOICE only; the effect
  sentence is removed from DC1. Route predicates consume the persisted choice immediately; a
  successor content RFC owns contributions/tree-unlocks and mints them onto new runs.
- **D4 — RULED: persistent acceleration-burst (NOT instant), and it DOES need a Company save bump.**
  The existing catalog already carries `burst_speed` + `burst_max_duration_ms`, and the design's
  language is "acceleration burst" — so the model is a persisted Company-run **burst state** activated
  by the intent and consumed by later production accrual, NOT an instant synthetic interval (which
  would need a separate virtual-time cursor and dishonest `evaluated_through`). Lifecycle:
  `spend_compute_credit` debits `ComputeCreditMS`, sets a burst `{activated, boosted_duration derived
  from amount_ms + catalog, burst_speed}` capped by `burst_max_duration_ms`; production accrual
  multiplies rate by `burst_speed` while the burst has remaining duration, consuming duration as
  production is computed (attended AND offline both consume it); **Exit resets any active burst** (the
  banked `ComputeCreditMS` persists per its existing scope; only the in-run burst is consumable);
  integer/Decimal rounding order stated in the impl; the F1 online-horizon precondition applies (a
  large accrual must not brick). **This corrects the draft's "no save-version bump" claim: the burst
  state requires a Company save-version bump. This is a SECOND Company-save-version claimant alongside
  Active-Play Buff Windows (A-blockers) — the two must be sequenced or atomically co-activated
  (meters v15 + achievements v16 pattern); resolve the exact Company version at implementation.**
- **D5 — RULED (option a): defer auto-spend to a successor settings RFC; this intent is MANUAL-ONLY.**
  There is no founder-settings gameplay store, and a Founder toggle that changes Company production
  crosses the replay boundary — too much for this foundation. The HUD exposes the credit BALANCE now;
  it does NOT claim an auto-spend policy exists. DC2 reconciled (the auto-spend affordance removed).
- **D6 — accepted.** Enumerate both intent objects, applied receipts, and event payloads byte-for-
  byte; `amount_ms` positive safe-integer; reuse only existing rejection reasons with closed detail
  values; register both event kinds in the Go/DB registries; shared sequential Go/TS fixtures cover
  applied, repeated-ID replay, malformed exact-key input, every semantic rejection, and the burst
  boundary cases (activate, partial-consume, exhaust, Exit-reset).

Structure ruled; the doctrine catalog, burst numbers (`burst_speed`, cap), and spend targets beyond
`accelerate` are content/data. Compute Credit auto-spend and doctrine effects are named successors.

## Implementation acceptance recheck (2026-08-05)

Reading the reconciled body against the live seven-artifact epoch, v14/v16 Company save chain, and
closed Go/TypeScript replay grammars found four residual executable gaps. The D1-D6 direction is
sound, but code cannot select these answers without creating mechanics or replay history.

### D7 — The persistent burst still has no executable cost/time equation

D4 chooses persistence, but `boosted_duration derived from amount_ms + catalog` does not define the
derivation. It also leaves active-burst stacking/replacement, cap behavior, offline-cap interaction,
and the rounding boundary unspecified. Those choices change how much production one credit buys.

**Proposed contract:** `amount_ms` is both the exact credit debit and boosted wall duration, 1:1.
Reject (do not clamp or waste credit) when `amount_ms > burst_max_duration_ms`, the balance is too
small, or a burst is already active. Persist only `compute_burst_remaining_ms`; speed is resolved
from the run-pinned economy artifact. During evaluation let
`boosted_wall_ms = min(elapsed_ms, remaining_ms)` and consume that wall duration in online and
offline modes. Base accrual remains unchanged; a second bonus segment accrues the eligible boosted
portion at factor `(burst_speed - 1)` through the same provision-boundary segmentation and
deterministic Decimal sum, quantized once with the ordinary accrual. Offline production remains
bounded by `accrual_cap_ms`; wall time beyond that cap may expire the burst but produces no bonus.
Require `burst_speed > 1`. Exit writes the terminal state with zero remaining and starts the next run
at zero. Partition and 25-hour-return fixtures must cover a burst spanning a provision boundary and
the offline cap.

### D8 — No literal doctrine row can satisfy the pick/gate contract at Phase 0

D1 defines a row shape but supplies no row. The only committed doctrine IDs are
`doctrine.capture` and `doctrine.ethical` on `transition.t3_to_t4`; Phase 0 deliberately has no
`gate.t3_to_t4` or permit resource, and the temporal-validity RFC moved those routes to
`gate.t4_to_t5` for exactly that reason. The implementation cannot choose a row's `gate_id`, gate
requirement, or whether the pick is currently reachable. Nor does D2 state whether a declared
doctrine row makes its gate require a pick only at the row's source tier or on every later
out-of-order crossing (the current gate engine permits legal monotonic out-of-order crossings).

**Proposed contract:** ship the schema and cross-artifact validator now, but keep the production
`doctrines` artifact unminted until the T3-to-T4 content owner supplies the missing gate and literal
row. Test with a shared fixture row
`transition.t3_to_t4 / source_tier 3 / gate.t3_to_t4 / [doctrine.capture, doctrine.ethical]`.
When active, `cross_gate` requires the pick only for the exact declared gate and only when the
Company tier equals `source_tier`; any other tier is rejected by the gate-sequence owner rather than
silently satisfying or bypassing doctrine ordering. If the production artifact must mint in this
RFC, the owner must instead supply the literal Phase-0 row and any new gate requirement.

### D9 — “Company save bump” has no legal activation path from the current epoch

The live epoch has seven artifacts and Company v14. Company v15/v16 are an atomic pair activated
only by the still-unminted `meters` + `achievements` artifacts; standalone v15 is forbidden. Adding
only `doctrines` and calling the result Company v17 would skip the required lower artifact pair.
Company v17 is also already a different Founder-axis schema, so the scope-specific wire keys must
be explicit. Active-Play is a second unresolved Company-version claimant.

**Proposed contract:** implement the Company-v17 codec and doctrine activation logic now, but do not
mint/activate it. A doctrine-bearing production epoch must also contain the paired `meters` and
`achievements` artifacts and activates Company v14 -> v17 only at Exit; `versionFloors` requires all
three artifacts and forbids disappearance. Company-v17 exact keys are v16 Company keys plus
`compute_burst_remaining_ms`; Founder-v17 keeps its existing minigame schema and must reject burst
state. Active-Play therefore sequences at Company v18 unless its owner explicitly amends this RFC
before either production mint. If Doctrine is intended to activate before Meters/Achievements,
owner must define a non-linear Company feature-version model instead.

### D10 — D6 says “enumerate,” but the wire remains unenumerated and AC3 is stale

No exact objects were added for either intent, event payload, snapshot field, or rejection detail.
The current acceptance criterion 3 still requires an auto-spend policy and Founder setting even
though D5 explicitly defers both. The event registry and TypeScript parser cannot be implemented
from “byte-for-byte” as an instruction.

**Proposed contract:** exact requests are
`{intent_id,kind:"pick_doctrine",expected_revision,transition_id,doctrine_id}` and
`{intent_id,kind:"spend_compute_credit",expected_revision,amount_ms,target:"accelerate"}`.
Both use the existing generic applied receipt. Add snapshot fields `compute_credit_ms` and
`compute_burst_remaining_ms` only; no auto-spend field. `doctrine_picked` payload is
`{founder_id,run_id:{company_stream_id,run_seq},transition_id,doctrine_id}`.
`compute_credit_spent` payload is
`{founder_id,run_id:{company_stream_id,run_seq},amount_ms,target,burst_duration_ms,burst_speed}`.
Malformed details are `<kind>.fields`, `transition_id`, `doctrine_id`, `amount_ms`, or `target`.
Semantic failures use existing category `unknown_id` for unknown catalog IDs and `not_eligible`
with closed details `tier`, `gate_crossed`, `doctrine_already_picked`, `doctrine_required`,
`compute_credit_balance`, `burst_duration`, or `burst_active`. Replace AC3 with: “the authoritative
snapshot exposes credit balance and remaining manual burst duration; no auto-spend policy exists.”
