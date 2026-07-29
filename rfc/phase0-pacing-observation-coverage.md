# RFC: Phase-0 Pacing Observation Coverage

- **Status:** implementing
- **Author:** Codex
- **Created:** 2026-07-29
- **Design refs:** `design/02-economy-balancing.md §11`; `design/research/pacing-science.md §7`
- **Depends on:** Balance Harness Foundation (implemented)
- **Parent / amends:** `archive/balance-harness-foundation.md`
- **Supersedes / superseded by:** —
- **Planning:** `planning/phase0-pacing-observation-coverage/`

## Summary

Make the first Phase-0 pacing baseline representative of everything the implemented scenario can
observe. The current aggregate records only two bounded Casual statistics even though every run
already measures four milestones under two policies. This follow-up records p50 and p95 for every
declared policy/milestone pair without inventing Tier-1 or Exit behavior.

## Motivation

A clean 300-run report completed with zero invariant failures, but its committed aggregate contains
only first-purchase p50 and generator-count p95 because the harness aggregates only entries in the
scenario's `envelopes` array. T0 progress and the Chaos distribution can drift without changing the
baseline. The existing envelope shape already permits an entry with no minimum or maximum, so the
coverage can expand without a second report grammar.

This RFC is observation coverage, not retuning. Existing inclusive bounds remain unchanged and no
new pass/fail pacing target is inferred from current output.

## Specification

### D1 — Complete observation matrix

For every distinct policy ID declared in `runs` and every milestone ID declared in `milestones`, the
scenario contains exactly one p50 and one p95 envelope tuple. An unbounded tuple contributes an
aggregate/baseline value and drift comparison but no target failure. A bounded tuple performs both
roles.

Phase-0 therefore produces exactly 16 ordered values: two policies × four milestones × two
statistics. Ordering is scenario declaration order: policy order, then milestone order, then p50
before p95. The existing Casual first-purchase p50 ≤60,000 ms and generator-count p95 ≤300,000 ms
bounds remain on their matching tuples.

### D2 — Runtime and schema validation

Both the Go loader and repository schema-semantic command reject:

- an envelope whose policy is absent from `runs`;
- an envelope whose milestone is absent from `milestones`;
- a duplicate `(policy_id, milestone_id, statistic)` tuple;
- incomplete p50/p95 coverage of any declared policy/milestone pair.

The JSON Schema continues validating field shapes; cross-reference and matrix checks are semantic.
Scenario content version increments because its checked aggregate surface changes. Harness/report
schema versions do not change.

### D3 — Baseline update discipline

The scenario change lands first. `golden-seed.json` and `pacing-baseline.json` regenerate in a
separate artifact-only commit beginning `BALANCE-CHANGE:`. The hardened history guard must validate
that commit. The canonical docs state the 16-value coverage but do not copy observed values that are
already authoritative in the generated baseline.

### D4 — Honest scope boundary

This remains a T0 production baseline. It does not claim Tier 1, first Exit, wall, strategy-diversity,
or long-arc coverage. T0–T1 playable content and Prestige & Exits own those milestones and envelopes
when their state transitions exist.

## Deviations from design

The pacing research names full-game personas and milestones. This follow-up deliberately measures
only implemented mechanics and strengthens regression coverage without fabricating later tiers.

## Acceptance criteria

1. The Phase-0 aggregate contains exactly 16 unique values in the normative order.
2. Go and `make verify-schema` reject unknown policy/milestone references, duplicate tuples, and one
   fixture missing a required matrix cell.
3. A changed T0-progress observation participates in the existing 10% warning/25% failure comparison.
4. The scenario and artifact commits satisfy the hardened `BALANCE-CHANGE:` history guard.
5. `make verify` passes; the clean report completes all 300 runs with zero invariant failures.
6. Independent diff review is recorded before archival.

## Open questions

None. Later feature RFCs own new milestones and their target values.

## Changelog

- 2026-07-29: created and implementation started by owner direction to establish the first real
  pacing baseline after remediation completed.
