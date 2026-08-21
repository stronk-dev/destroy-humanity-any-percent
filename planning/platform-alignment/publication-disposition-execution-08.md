# D-002 Publication Disposition Execution — Fresh-Clone Authority Proof

**Date:** 2026-08-21
**Authority:** D-002 owner ruling and `publication-sensitivity-audit.md`
**Implementation commit:** `e44e1a6`
**Product/design/RFC authority:** none

## Implemented gate

- `publication-authority-manifest.json` names the complete 67-file Class-C population: 11 public
  dossiers and 56 ignored private-source dossiers.
- It separately binds the two private Class-B raw sources to their tracked public derivatives and
  review contract, the three ignored duplicate drafts to canonical production artifacts, six
  moved historical references to their tracked archive paths, and seven ignored diagnostics to
  tracked canonical records.
- `tools/verify-publication-authority.mjs` validates counts, disjointness, tracking, ignore rules,
  per-file review contracts, derivatives/canonical records, the exact tracked research population,
  and ignored-artifact links from the named authority files.
- `make publication-authority-check` runs three negative controls and the honest manifest. The
  gate is intentionally outside CI/release dependencies; D-002 authorized a local proof, not a
  permanent workflow change.

## Demonstrated discrimination

The gate rejected all three forged states before accepting the honest one:

1. a private/missing path promoted into the public Class-C set;
2. a diagnostic pointed at itself instead of a tracked canonical record;
3. one private Class-C row removed from the denominator.

Each failure named its exact violated invariant. The honest result was:

```text
publication authority: PASS (11 public Class-C, 56 private Class-C, 3 duplicates, 7 diagnostics)
```

## Fresh-clone proof

After `e44e1a6` committed the gate, `make publication-authority-fresh-clone-check` cloned committed
HEAD into a new temporary directory with no ignored working-tree files, reran all three negative
controls, and returned:

```text
fresh-clone publication authority: PASS
```

D-002's executable exit is complete. This proof does not publish, push or delete anything and does
not convert the 56 ignored raw dossiers or seven diagnostics into authority.
