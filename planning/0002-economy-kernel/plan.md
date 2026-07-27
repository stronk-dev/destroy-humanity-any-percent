# RFC-0002 Economy Kernel — Implementation Plan

- **RFC:** `rfc/0002-economy-constants-and-ceilings.md`
- **Assignee:** Codex
- **Status:** implementing

## Work breakdown

1. Define the shared catalog JSON Schema and representative valid/invalid fixtures.
2. Implement strict Go catalog loading, typed definitions, and cost-curve evaluation.
3. Implement strict TypeScript catalog loading and matching cost-curve evaluation.
4. Implement the authoritative Go ledger query/transaction/receipt boundary.
5. Add shared parity tests and ledger invariant/regression tests.
6. Run all local gates; correct drift without loosening invariants.
7. Distill canonical behavior into `docs/economy-kernel.md`, update indexes/README, mark the
   RFC implemented, and archive both RFC and planning directory.

## Acceptance gates

- Both loaders accept the shared valid fixture and reject missing fields, unknown fields,
  duplicate IDs, dangling references, invalid decimals/bounds, invalid scopes/kinds, and bad
  curve parameters.
- Constant, linear, and geometric bulk quotes and max-affordable results agree across Go and
  TypeScript for normal and enormous values.
- The Go ledger is atomic, enforces minimums/hardcaps/state validity, aggregates before one
  quantization, and returns deterministic receipts.
- The million-source sub-resolution regression and its per-entry negative control pass.
- `make verify` passes without changing RFC-0001 numeric semantics.
- Canonical docs and archives are complete; worktree is clean after milestone commits.

## Scope guard

Do not add shipped balance values, production/time integration, save persistence, network APIs,
UI state, arbitrary formulas, or callbacks. Any need for them is a `DESIGN-GAP` and a separate RFC.
