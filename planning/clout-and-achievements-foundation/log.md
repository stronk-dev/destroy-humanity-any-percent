# Achievements Foundation implementation log

## 2026-08-03 — accepted-contract reconciliation

The owner accepted C1–C10, narrowing the RFC from Clout + achievements to Achievements only. The
file still contained the rejected Clout mint/stack in its summary, normative sections, acceptance
criteria, open questions, README status, and last changelog entry. Reconciliation makes one rule:

- achievement IDs and exact achievement score are permanent and non-spendable;
- ordinary Company transitions latch run IDs; the existing multi-stream Exit settles Founder
  lifetime state idempotently and resets the next run;
- Clout has no writer or production factor in this foundation; future social activity owns its
  sole mint and PR-Intern content owns any run-local multiplier;
- Phase-A predicates use only implemented state/event boundaries; Meters predicates append later;
- production definitions are content and copy-key data, not invented foundation mechanics.

`make verify` passed at `61b6392` immediately before this implementation round.

## 2026-08-03 — catalog, proof, and evaluation foundation

- Added strict schema-v1 Go/TypeScript loaders. Achievement IDs are byte-sorted; score grants are
  positive exact integers; every copy, generator, event, resource, and counter binding resolves
  through an explicit composition registry.
- Added the bounded condition union and the three proof arms. Possession requires an ownership
  predicate plus justification copy, provenance refuses current-possession predicates, and burn
  requires a registered event/resource plus positive canonical Decimal minimum.
- Added pure evaluation against one observation snapshot, lifetime/run latch exclusion, and score
  derivation from the earned-ID set as the single authority.
- Added JSON Schema fixtures and a fail-closed package/source gate proving this foundation imports
  no economy spending, production/save owner, or lifetime-Clout symbol.
- Focused Go, client, TypeScript type-check, schema, and boundary gates pass. No production
  achievement or copy was invented; fixtures are explicitly pre-mint.
