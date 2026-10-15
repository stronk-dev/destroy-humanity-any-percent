# RFC: Exact Represented Aggregation

- **Status:** draft — proposed arithmetic/caller activation contract; NOT implementation authority
- **Author:** Codex
- **Created:** 2026-10-15
- **Design refs:** `design/06-tech.md` numeric parity/server authority; `design/02-economy-balancing.md` §11
- **Parent / amends:** [Deterministic Decimal Aggregation](archive/deterministic-decimal-aggregation.md); affected Numeric Core, Economy, Commons, replay and Harness report contracts
- **Related, not adopted by this draft:** [Production Accrual Conservation](production-accrual-conservation.md), [Accrual-aware Harness Search](accrual-aware-harness-search.md), [Numeric Diagnostic Exponent Limits](numeric-diagnostic-exponent-limits.md)
- **Evidence / planning:** `planning/clout-v1-and-pr-interns/sparse-aggregation-research.md` and owning plan/log; dedicated implementation plan after acceptance

## Summary

Add an exact represented-input sum, migrate its named live consumers together in Go/TS,
and preserve two pinned exceptions: existing catalog admission and Pitch engine1.0.0.
Use the existing kernel identity to distinguish changed harness producers rather than
retune inputs or silently replace goldens. Company drift follows its accepted policy;
Commons receives a capacity-only derived-cache cutover. No persistent arithmetic-mode
field, global runtime switch or new minigame version is proposed.

This is a separate numeric correction, not a solution to Clout partition conservation
or temporal search. Its proposal does not authorize either linked draft, report
regeneration, source activation, publication or deployment.

## Motivation and observed scope

RP-664's five admitted production rates yield a smaller Quantize12 payout after one raw
input increases across an exponent group. Aligned compensation loses valid cancellation
tails. The sparse prototype preserves those tails and remote ties; its2048 raw witness
digest agrees across the previously executed runtimes. Its separately measured Go
identity/scratch optimizations preserve that digest but do not supply full-domain review.

A global replacement also changes an admitted Pitch1.0.0 funding decision, catalog
weight validation, persisted Commons capacity and harness reports. A kernel bump alone
does not dispose of these independent contracts. The table below gives each a specific
proposed treatment; none may silently inherit whichever reducer is linked at build time.

## A1 — Two named operations, one owner for each recipe

Keep existing `SumDeterministic` / `sumDeterministic` as the grouped recipe specified by
the archived parent. Do not duplicate that recipe in a tenant or loader, rename old
goldens into a new authority, or change its outputs under this proposal.

Add `SumExact` / `sumExact` in the same numeric authority. Its inputs are the same
normalized finite state Decimals; no wire/precision/exponent-domain expansion:

The checked Go API returns `(Decimal, error)`; TS returns a Decimal and throws a named
invariant error only for an internal algorithm failure. Invalid/domain results below are
NaN with no invariant error, as opposed to broken-instrument/implementation execution.
Callers propagate internal errors through their existing failed-operation path. In
particular Harness resourceTotals must gain error propagation, not emit partial/empty
totals. The manual prototype's panic wrapper is NOT production authority or a crash policy.

1. Interpret every input's binary64 mantissa and decimal exponent exactly, not its
   12-digit rendering. Invalid inputs yield diagnostic NaN, never zero or exclusion.
2. Sum those represented values exactly, including signed cancellation across the full
   admitted exponent distance. Input order does not affect the result; inputs are untouched.
3. Exact zero is canonical positive zero. For nonzero magnitude, choose its decimal
   exponent and correctly round its normalized mantissa to binary64, ties to even.
   Normalize a rounding carry. An exact or rounded result outside the existing state
   domain yields diagnostic NaN; recoverable intermediate overflow is not a refusal.
4. Existing Quantize12 remains the caller's unchanged explicit state boundary. Do not
   pre-round inputs, add tolerance, change Add/Sub/Mul/Div or reinterpret receipts.

The proposed sparse implementation decodes q=abs(mantissa)*2^52 as an exact integer,
uses separate positive/negative radix1e9 sparse digits, skips empty exponent intervals,
represents long borrow intervals as runs, and uses leading chunks plus sticky information
for final binary rounding. The dossier supplies its derivation/work argument and independent
oracles. No dense exponent-sized allocation or per-exponent-gap loop is permitted.
Local zero/single-input identities and scratch reuse must preserve the same result.
All invariant/guard exhaustion is an explicit failed execution, not valid NaN data.

Neither operation accepts a caller-selectable policy flag. Static call sites name the
intended operation. No live account/run/session chooses its arithmetic from a save field,
report, environment variable, cache state or deploy-current content lookup.

## A2 — Complete current caller disposition

Baseline38cba707 has13 direct Go and7 direct TS product calls (excluding declarations/imports).
The landing review rederives this inventory; a newly discovered caller needs a disposition,
not default migration or silent exclusion.

| Consumer | Proposed operation / behavior |
|---|---|
| Economy ledger and TS replay ledger net | SumExact, including signed batches; retain atomic range refusal, saturating-credit and reproducible receipt rules. |
| Constant accrual in Go/TS | SumExact for raw rates; retain efficiency/duration formula and explicit quantization. |
| Production segment rates and final delta totals; TS replay advance | SumExact at both corresponding boundaries; no rounding of source terms or change of provision/burst timing. |
| Active-play lucky rate and TS replay claim | SumExact; retain actual request/cap/credited-delta and receipt semantics. |
| Go/TS composite progress evaluation | SumExact for weighted values, with existing clamp/quote/gate definitions; catalog admission remains the exception below. |
| Go simulation resource totals and ProjectRates resource/isolated-generator totals | SumExact, matching authoritative production; reads still cannot settle, write or emit events. |
| Commons AggregateHealth capacity | SumExact; integer health arithmetic unchanged; follow A5 before serving derived rows. |
| Harness resourceTotals and production consumers | SumExact under the new producer identity/report contract in A6; preserve all experiments and populations. |
| Existing catalog composite-weight admission | KEEP grouped SumDeterministic's exact raw equality to One. No stored/retuned catalog acquires a different grammar because runtime arithmetic changed. |
| Pitch engine1.0.0 scoring, both runtimes | KEEP grouped SumDeterministic. No tenant engine identity/session migration/coexistence change. |

The last two are explicit pinned consumers, not fallback for invalid exact sums. Adding
another grouped consumer requires its own accepted contract/review. Do not route ledger
refusals through the grouped operator to make an exact result acceptable.

## A3 — Catalog and Pitch compatibility proof

Both catalog loaders must keep their existing weight grammar and byte/hash identity.
Compare old/new admission on all retained content plus accepted and refused boundary
fixtures, including the exact equality calculation; no normalized/reweighted substitute.
Progress execution may change under the new kernel while admission remains the same law.

Pitch's admitted retuned seed2 hand remains5.97255552034e3 and enters shop at that funding
threshold under engine1.0.0. Test actual Go registry Create/Apply, TS tenant and persisted
resume/payout, not only the arithmetic helper. Its grouped recipe is unchanged even when
the surrounding Company uses the new kernel. Existing minigame authorization, payout
atomicity, certification and replay-input boundaries remain required.

The retained global-overlay negative must still demonstrate that substituting exact sum
inside Pitch changes its phase; the adopted implementation must not do that. No new
engine version is needed only if complete tenant behavior is actually preserved.

## A4 — Company, replay and receipts

Activation genuinely changes shared-kernel behavior and requires the actual next kernel
version in Go/client, path registration where needed, and current full-history guards.
Do not reserve a version number in this draft or use it to waive existing history errors.

Reuse accepted run-version drift: old Companies may continue on the new kernel with
immutable drift recorded and time-board exclusion; drifted verification reports
engine_mismatch, and unavailable pinned verification remains deferred. Do not rewrite
old state, receipts, logs, genesis, catalogs, terminal facts or verification verdicts.
No perpetual per-Company arithmetic branch or schema migration is introduced.

Paired Go/TS replay must use SumExact at the table's corresponding live boundaries,
through the same closed transition as the server. Actual Store/Postgres proves commit,
rollback, identical retry and drift/deferral. Existing lower-domain nonrepresentable
deltas still refuse atomically; this proposal does not authorize direct balance writes.

Clout conservation's proposed compatibility criterion compares accounting changes against
the SAME adopted numeric kernel. Only separate acceptance of this numeric correction can
establish its new-kernel reference payouts; old outputs remain historical evidence, not
silently edited expectations. Neither draft accepts the other, changes today's reference
or closes the27 partition failures. The accepted Clout/offline contract stays untouched.

## A5 — Commons capacity-only cutover

Existing member samples/history remain exact. Recompute derived cohort/server capacity
from the same admitted membership joins and stored sample capacities with SumExact.
Before the new single-node binary admits gameplay/reads, update every existing derived
scope's capacity in one database transaction, under the deployment's exclusive writer
boundary. Failure rolls back and fails readiness; no partial cache or silent skipped scope.
Missing/new scopes retain their ordinary creation path. Measure complete-scope work before
activation; do not invent a cap, truncate samples or raise startup/recovery bounds.

Do NOT simply call current refreshScope for this cutover: it also smooths health and writes
evaluated_at. The capacity-only operation must preserve raw_health_ppm, health_ppm,
real_members, npc_weight_ppm, evaluated_at, samples, memberships, outbox and events byte-for-
byte. Share the actual member-selection authority, not a divergent SQL approximation.
Ordinary later refreshes use the same new capacity operator and retain health chronology.
Replay or rebuild must not relabel already processed old events as new deliveries.

Real Postgres controls include exact before/after retained fields, two scopes, duplicate
delivery, malformed capacity, failed cutover rollback and threshold-sensitive downstream
consumers. Do not count pure AggregateHealth vectors as persisted cutover proof.

## A6 — Honest harness producer identity and baseline migration

Use the existing kernel/VERSION authority, not a second arithmetic-protocol registry.
For new report formats, the next-free RunKey schema adds required kernel_version where a RunKey exists;
reports without RunKeys carry required kernel_version at their existing identity root.
Allocate next-free family report schemas at landing. Writers use the actual loaded build's
kernel identity, never a report/scenario override. Validation requires one matching producer
per report. Old report formats retain their exact old fields and readers; missing identity
in a new format refuses, not an assumed legacy/current producer.

Bind that value in actual key serialization, equality, ordering, reuse/cache identity and
baseline comparisons, not only a displayed metadata field. Identity-root-only reports bind
it in their complete report identity/comparison too. A version label is not proof of executing
the operator: actual caller/counterexample tests must reject a label-only migration.

The existing scenario/policy/catalog hashes and populations remain unchanged when only
the producer changes. Do not edit a scenario, retune a constant or mint a balance epoch
merely to manufacture an input-first baseline signal. The separate accrual-search draft's
future calculation_protocol/profile remains genuine experiment input and independent
authority; it is not adopted or replaced by kernel_version.

This requires an EXPLICIT amendment to the baseline guard, not a claim its current rule
already allows code-only restamping. For the new kernel-aware formats only, a later artifact
may follow either a real accepted input change OR a prior genuine accepted kernel-behavior
change with a different bound producer version. Preserve artifact-only BALANCE-CHANGE commits,
complete reachable-history validation, full population and dirty/malformed/shallow refusals.
Legacy-format changes retain their current input-first rule. The artifact's producer must
match kernel/VERSION at its source parent; a claimed different version without that actual
source change refuses. No dummy bump or historical guard exception is authorized.

Before the first new-format baseline, bind each old artifact's exact SHA to its evidenced
historical producer revision and reproduce it with that pinned source/toolchain. Missing
provenance blocks migration; do not infer an old kernel from today's tree. Retain old bytes
and reference results in history. Run the entire SAME input/seed/policy/window population
with the new producer, retain complete before/after differences and failures, then submit
the separate generated-artifact range for review. New identities distinguish changed math;
they do not make failed floors pass or authorize golden replacement automatically.

Future current-source CI runs the complete population against its bound current producer;
historical-source reproduction is distinct from claiming old reports were generated by the
new kernel. Coordinate the new family formats with accrual-search H4 so neither reader
silently accepts another producer or reuses the same schema number for different fields.

## A7 — Proof, cost and acceptance

1. Independent review of exact representation, sparse carry/borrow, full-range rounding,
   overflow/underflow and work arguments; passing sampled profiles alone is insufficient.
2. Retain original RP-664 and cancellation failures; shared Go/TS raw-output vectors cover
   signs, permutation, empty/zero/single inputs, remote sticky ties, recoverable overflow,
   net refusal, ownership and all existing numeric operators' unchanged goldens.
3. Complete caller inventory/migration, paired transitions, negative fixtures and A3/A4/A5
   real boundaries; prove static exceptions cannot bleed into live exact consumers.
4. A6 readers/writers and real-Git seeded negative/positive migration controls execute;
   all historical guard failures stay visible until lawfully repaired independently.
5. Actual complete affected pacing/relevance/search/content/Commons workloads execute with
   unchanged bounds/counters and no dropped populations, coast, timeout inflation or hidden
   guard exhaustion. The41-allocation multi-source prototype costs about5x the grouped
   reducer in short samples. A later unapplied global-overlay standard-phase native pair finishes300/300
   under each implementation, candidate770.708s versus active948.034s timed. Its matching
   aggregate hides36 changed final cash balances/state hashes; it is not behavior-preserving.
   The owning RP-664 dossier/log retain full artifact identities and limits. One phase/pair
   proves neither statistical/hosted speedup nor complete affected-workload feasibility;
   other workloads and this draft's compatibility/producer/cutover gates remain mandatory.
6. Canonical numeric/economy/production/replay/Commons/Harness docs, owning logs, exact-range
   designated review and all accepted feature gates reconcile before archival. A prototype,
   green fixture or kernel bump is not acceptance, release readiness or Clout AC6 completion.

## Deviations, open gates and authority

This intentionally changes the represented-input aggregation recipe for the named live
consumers and proposes kernel-aware producer identity plus a bounded baseline-protocol
amendment. It preserves state precision, pinned tenant/catalog behavior, resource formulas,
caps, offline policy, real-money prohibition and owner copy. It is not an editorial repair.

RFC acceptance requires agreement on this exact algorithm, static exceptions, error/transaction
and report/guard contracts, designated normative review and owner adoption. Full workload
cost, evidenced old-report producer bindings and capacity-only cutover proof are mandatory
implementation/activation gates, not a demand that an unimplemented RFC already be shipped
before it can be accepted. Linked draft compatibility text explicitly names the separate
reference/producer boundary; that reconciliation is not either adoption.
Implementation does not begin from this draft. The current active grouping, CI failures,
report bytes and all nine-tier1.0 obligations remain unchanged.

## Changelog

- 2026-10-15: propose one explicit numeric/caller/activation follow-up from RP-664's
  executed counterexamples and cost studies; no source, artifact or policy activation.
