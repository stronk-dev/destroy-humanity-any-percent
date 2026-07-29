# Phase-0 Pacing Observation Coverage — append-only log

## 2026-07-29 — start

- The post-remediation clean report completed 300/300 runs with zero invariant failures.
- Its committed aggregate has only two values because aggregation follows `envelopes`, while all
  runs already observe four milestones under Casual and Chaos. Current deterministic observations
  include Casual first generator at 10,000 ms and T0 progress at 337,000 ms; these are evidence,
  not newly declared targets.
- The existing unbounded envelope representation is sufficient. This RFC adds coverage and strict
  references without a second baseline grammar or any invented later-tier mechanic.

## 2026-07-29 — source implementation

- The Phase-0 scenario is version 4 and declares the complete 16-tuple matrix in policy,
  milestone, then p50/p95 order. The two existing target bounds are unchanged.
- Go loading and the repository schema command reject unknown policy/milestone references,
  duplicates, and incomplete matrices. Four shape-valid negative fixtures exercise those semantic
  failures, while a Go table test proves the same runtime boundary.
- A repository fixture test binds exact count and order so future scenario edits cannot silently
  drop an observed distribution. Harness unit and schema gates are green before artifact update.
