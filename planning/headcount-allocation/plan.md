# Headcount Allocation implementation plan

Authority: accepted `rfc/headcount-allocation.md`; design/01 Tier 2.

## Current scope

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

- Resolve OD-1 with Tier 2's author/owner before seat-source-dependent work.
- Reconcile the siblings' conflicting catalog/state/intent/output contracts
  (RP-414), not only the seat source. No integration branch is chosen here.
- Reconcile S8/AC3's inadmissible overflow vector and unproved fault outcome
  (RP-413); retain the original rejection and legal boundary evidence separately.
- Reconcile occupied economy/save/replay version numbers in the accepted body
  before integration; do not reuse its historical v5/v19/v9 literals blindly.
- Catalog grammar/bindings, closed-form rates, state/migrations and logged intents.
- Replay, published formulas, harness and authoritative UI projection/panel.
- Full acceptance, adopted content, integrated player proof and designated review.

New arithmetic modules are not imported by the live engine/client. Kernel replay
semantics and watched paths remain unchanged here. Before authoritative wiring,
register both new module prefixes and perform the genuine integration version bump
with its replay/migration tests; this is not an exception to kernel identity.
