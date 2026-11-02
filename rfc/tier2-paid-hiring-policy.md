# RFC: Tier 2 Paid-hiring Policy Binding

- **Status:** draft — technical review/adoption and actual capability binding pending; NOT implementation authority
- **Author:** Codex
- **Created:** 2026-11-02
- **Parent / amends:** Tier 2 Content P1/HG-10; Headcount Allocation S9's harness consumer
- **Design refs:** `design/01-tiers.md` Tier 2; `design/02-economy-balancing.md` §11/§11b; `design/06-tech.md` lazy authoritative math
- **Depends on:** Headcount activation reconciliation; adopted Tier 2 windows; Production Accrual Conservation and Accrual-aware Harness Search adoption/implementation
- **Planning:** `planning/tier2-content/`; engine prerequisites remain in `planning/headcount-allocation/`

## Outcome and evidence boundary

Specify HG-10's deterministic paid-hire/allocation decisions, not the remaining
content, accounting, loader, balance or player integration. Marco selected the
engine model, not these test algorithms. Review and adoption must precede parent
reconciliation or implementation.

Source `0e5cf832`: `first_hour_runner.go` enumerates single-generator purchases,
upgrades, manual and wait; Casual considers generator/upgrade purchases. Reference's
cost/intent/ranker paths do not implement Headcount and its first-hour adapter
stops ranking after the Garage gate. Tier2 currently runs an explicit subset.
A cash-priced unassigned hire would reduce cash without increasing rate in the
post-purchase ratio. Simply adding that arm cannot value its potential. This is
a source-derived constraint, not an executed hiring defect or balance result.

## Specification proposal

### HP1 — New policy authority, old policies unchanged

Use P1's future `balance/testdata/t2/it-company-policy-v1.json`, schema2, registry
ID `it_company_policy.v1`, version1; exactly `chaos.t2` v1, `casual.t2` v1 and
`reference.greedy` v2. No artifact is created here. Copy pinned first-hour registry
version3, SHA256 `e5e5de7051beb0340e54f7013ce7d4a48c35bfcc3343220310290478445d10c3`,
with P1's adopted windows, gate/exit/offer/objective replacements and these decision
extensions. Preserve complete Casual jitter/clock, cadence, run continuation,
first-action timing, seed material and wait legality. No old-ID fallback or relaxed v1.

Keep P1's six-class order; `within_class` becomes `raw_byte_command_key`.
Existing purchase keys remain item IDs; hire keys are `hire:<pool_id>:1`.
Move keys remain `move:<from>><to>`: the globally unique destination identifies
its pool, including moves from `unassigned`. Duplicate keys/incomplete enumeration
fail. In this policy context reserve the source sentinel `unassigned`: if an
actual role has that exact ID, refuse the WHOLE measurement during preflight,
before enumeration. Do not confuse that role with the bench or silently omit it.
This does not change the engine's generic ID grammar; the Tier2 `role.*` rows are
compatible. Each new row's `decision.legal_commands` contains all six classes in order.
Keep inherited decision kind/affordability/wait fields and add one required,
exact-key `headcount_decision` object:

| Row | Complete object; all keys/values required |
|---|---|
| Chaos | `{"kind":"seeded_unit_moves","hire_count":1}` |
| Casual | `{"kind":"unassigned_then_cheapest","hire_count":1,"allocation_metric":"company_cash_rate","purchase_order":"resource_then_cost_then_key"}` |
| Reference | `{"kind":"paid_action_then_allocation_closure","hire_count":1,"allocation_metric":"projected_objective_ms","per_pool_move_limit":"current_seats","on_exhaustion":"allocation_starved"}` |

These policy literals are not balance constants or callback escape hatches.
The COMPLETE v2 schema must also encode P1's gate/exit/offer/objective rules;
these objects or adopted windows alone do not constitute the complete registry.

### HP2 — Finite vocabulary and actual legality

Enumerate existing manual, single-generator and unowned-upgrade commands; one
`hire_headcount {pool_id,count:1}` per open pool; all legal unit moves; inherited
legal wait last. Hire price uses the engine bulk-cost query at ACTUAL hired seats,
not office counts. Count1 bounds the persona vocabulary, like single-generator
purchases; engine/UI exact-count/Max and their acceptance remain required.

A move takes one nonempty assigned or genuinely unassigned head to another role
in the SAME open pool. Materialize the complete absolute vector, preserving other
counts/pools. Destinations remain roles as P1 specifies; the player's engine still
supports benching through full vectors. No cross-pool transfer or partial map.
Chaos/Casual require current affordability and actual applied simulation decisions
on independent branches; Reference may search under HP5. Semantic rejection removes
an illegal choice; evaluation/codec/context/budget errors fail the measurement.
Hires leave new heads UNASSIGNED. Quotes/probes never alter the selected run.

### HP3 — Chaos: one draw, no helpful assignment

Enumerate all legal HP2 candidates in class/key order. Use inherited SHA256 seed
material, NEW identity and modulo reduction once per ordinary decision boundary.
Apply only that command/wait. No automatic post-hire allocation or friendlier
retry. Expose unassigned/idle heads. An empty population fails, never coasts.

### HP4 — Casual: allocation consumes an ordinary action

After inherited ending/gate precedence:

1. If a legal move FROM `unassigned` exists, choose the greatest increase in the
   COMPLETE engine `company.cash` rate, including actual conversion/pool feed.
   Compare this one resource only; ties use raw move keys. Zero gain is allowed
   so legal deck-only assignment is visible. Rate errors fail rather than become
   zero. Execute ONE full-vector command and end the decision. This persona does
   not optimize already assigned heads.
2. Otherwise sort affordable generator/upgrade/count1-hire purchases by
   `(price resource ID, exact cost within that resource, raw command key)` and buy
   the first. Raw resource order is an explicit deterministic convention, NOT a
   comparison of different currency units. Cash-only cases retain cheapest-then-ID.
3. Otherwise use the inherited manual fallback, not a second strategy.

A hire at t cannot produce until a later ordinary allocation boundary, normally
t+5000ms. Session/ending/horizon clipping remains real: no offline/after-final
allocation. Closed-pool unassigned heads remain visible. This is an unhurried
persona, not an optimal-player claim.

### HP5 — Reference: paid action plus actual allocation closure

Keep P1's first unmet gates, then `company.cash >= 1e9`, then bank; preserve its
multi-resource ordering and exit-once precedence. No Garage-only early return or
ranking toward a met objective.

An allocation closure at one fixed endpoint considers every legal unit move
against the ACTUAL current objective. Use unchanged T01-C20 projected-time ratios
(deficit / canonical resource rate, including milliseconds), not resource weights.
Choose the smallest strictly improving score, raw move key on equal improvements,
apply that full-vector command on the branch and repeat. Finite improves
unreachable; two unreachable scores do not improve. Zero deficit remains zero.

At each ordinary Reference decision:

1. Perform P1's pre-decision allocation closure on the selected run.
2. Rank bank and all eligible generator/upgrade/count1-hire purchases. Use accepted
   origin-preserving search for earliest affordability in the ACTUAL price resource,
   with engine window/cap/requires checks. Preserve the decision horizon, clipped
   by scenario horizon and half-open active session; no action at/after session end.
   A later failed endpoint cannot suppress an earlier affordable result.
3. On an isolated branch apply the paid command at that endpoint: real debit,
   revision and (for hires) an unassigned seat. Run the SAME allocation closure
   for EVERY purchase class, not just hires.
4. Score its resulting state with the original projected-milestone equation plus
   actual affordability wait. Purchases beat bank on exact ties, then raw key.
5. For the winner retain/execute exactly that paid command and ordered closure,
   matching its COMPLETE scored result. Every actual command has a separate
   expected revision, capture/log row, result and event. Never install a scored
   allocation without its commands, grant production before payment, or choose
   another allocation after charging. A selected action error fails, not retries.

Only legal purchase sequences whose objective projection is reachable enter the
ranked list. A genuinely unreachable projection is not a finite score; arithmetic,
codec, hook, context and budget errors fail the measurement, never become
unreachable. If NO ranked purchase and NO reachable bank remain, apply exactly
the inherited one-batch manual bootstrap at the current ordinary boundary,
after the pre-closure. Speculative future probes do not move that endpoint.
An actual rejected/failed bootstrap fails the run rather than waiting, adding
free output or silently trying another action. A met terminal objective instead
banks to the next ordinary boundary; it does not trigger bootstrap or buy away
the already-met objective. All actions retain normal charging/capture.

Use at most `seats[pool]` applied moves across ALL closures of one ordinary
Reference decision, including its deferred purchase endpoint.
Carry pre-closure used counts into each purchase branch/selected continuation;
after a hire, use actual new seats MINUS used moves, not a renewed full allowance.
At the limit check for another improving move; if present fail `allocation_starved`.
Charge all trial/fixed-point checks. Forking never renews allowance. This bound is
neither a global-optimality proof nor a scenario-budget exemption.

Free instant Reference reallocations may share a timestamp but consume actual
transitions/revisions. Decision ordinal advances once per ordinary boundary.
Commands cannot precede affordability or the preceding action. Preserve P1's
existing newly eligible gate/ending precedence, not additional action windows.

Illustration ONLY: cash10, target100, rate1/s, hire cost5. Bank90s; unassigned
hire95s; if a subsequent LEGAL allocation makes rate2/s, paid sequence47.5s,
not cost-free45s. This is arithmetic, NOT admitted catalog/engine experiment proof.

### HP6 — Accounting, actual capabilities and full workload

Fork complete admitted state AND retained accounting context. Apply actions from
real origins, not installed rounded views. Preserve offline stages, lifetime,
role observations and rollback; discarded/refused branches supply no selected
evidence. Accounting/search drafts remain dependencies, not authority adopted here.

Explicit remaining prerequisite: Accrual-aware Harness Search H1 currently binds
first-hour/Tier2 to Company14, not Headcount. Before runtime work, specify,
independently review and adopt the ACTUAL candidate's source / sampling / hook /
completion binding. It must preserve complete new-version maps and independent
axis activation. NO downcast to14, numeric-floor widening, partial hook or constructor
bypass. Existing axis/attendance holds remain. HP1–HP5 alone are NOT a complete
implementation-ready P1 contract.

Use engine S7.4 role masks: retain seats/costs/assignments and remove executed
rate/pool feed. Every candidate/closure sees the same mask. No masked output
selects a branch or rescues a floor. Preserve P4/P5's Engineer/Sales gates and
Support/Manager observations. Every probe/search/action/closure/exact fallback
consumes the declared budget; setup is charged under the accepted accounting
protocol. Guard exhaustion is incomplete, never coasted success or a basis to
raise a ceiling. P2's complete measured/ratified budget remains; no number here.

### HP7 — Full P2 runs to its horizon, not the first gate

For the NEW full P2 population, process all declared policy activity boundaries
through its declared horizon, subject to the adopted half-open session rules.
Reaching the gate, first allocation, or both milestones does NOT end the run.
Keep the first gate's Founder-attended timestamp/run independently; keep the
first actual applied `reallocate_headcount` timestamp independently. Each remains
its first observation while later actions continue. Do not replace gate time
with allocation time or manufacture an allocation when the pool opens.

Current `tier2Runtime.done()` returns on the gate alone; the complete new runner
must replace that subset termination rule. Its pool opens at that same gate,
so retaining the early return would suppress the required later milestone.
Keep the existing subset/reports historically identified, not relabelled full P2.

No extra action/evaluation is forced at a clipped horizon or excluded session-end
timestamp. Existing scripted/elective run continuation and Founder-global windows
remain unchanged. After the last admitted boundary, finalize the measurement
without a saved payment. Missing gate/allocation remains a `must_reach` failure,
with unreached values retained in every seed/aggregate, never dropped. Record
horizon completion distinctly from early error/guard/unsupported interruption;
an early interruption is not a completed full run or a budget measurement.
Continuing to the horizon is a NEW protocol choice requiring adoption, not a claim
that the current subset already did so. Full-workload budget proof stays required.

## Future acceptance — not achieved here

1. Complete strict v2 registry/all three identities; malformed/missing/extra/null
   literals, quantities, classes and keys refuse. Preserve v1 and verify new-ID draws.
2. Independent multi-pool enumeration: complete keys, costs, vectors and actual
   production decisions; closed/capped/unaffordable/non-cash/idle-converter/pool-feed
   controls. Source/destination membership corruption fails.
3. Actual paid-then-allocated Reference sequence beats bank when appropriate.
   Free/preallocated hires, missing debit or intermediate capture/revision/event
   fail; ordinary purchases share the closure; no selected branch/evidence leak.
4. Public Casual runner pays then allocates at a later legal boundary; no offline /
   after-horizon or already-assigned optimization. Chaos has no helpful closure.
5. Closure controls cover strict/no improvement, unreachable/finite, ties, shared
   allowance and charged trials. A too-long improving chain fails loudly. Empty
   ranked options with unreachable bank execute the real manual bootstrap;
   projection errors cannot be hidden as unreachable/bootstrap success.
6. Full captures replay against actual Headcount source/pins across returns/both
   endings: full state/lifetime/caps/receipts/events/unassigned/idle in Go/TS and
   actual Postgres where owned, not a manually installed winning assignment.
7. FULL P2/P3/P4/P5, adopted windows, unchanged first-hour controls and content /
   copy ratification/mint. No subset or this draft repairs RP-433's population conflict.
8. Real full-P2 public runner reaches the gate, later pays/allocates and continues
   through all admitted boundaries to the declared horizon, retaining first times
   and every seed. Reinstating gate-only OR all-milestones early stopping must
   fail. A no-allocation fixture stays failed/unreached; after-session/horizon
   actions and early guard completion must fail too. Pure boundary/model checks
   do not replace this future actual engine journey.

No software pass, production behavior or feasibility is claimed for this prose.
Keep six-hour windows, 2h–3h target, 90%/24h offline, all populations and full
nine-tier1.0 scope. Technical review first; resolve findings and HP6 before
requesting adoption, reconciling parents or claiming ready to implement.
