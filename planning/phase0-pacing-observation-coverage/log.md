# Phase-0 Pacing Observation Coverage — append-only log

## 2026-07-29 — start

- The post-remediation clean report completed 300/300 runs with zero invariant failures.
- Its committed aggregate has only two values because aggregation follows `envelopes`, while all
  runs already observe four milestones under Casual and Chaos. Current deterministic observations
  include Casual first generator at 10,000 ms and T0 progress at 337,000 ms; these are evidence,
  not newly declared targets.
- The existing unbounded envelope representation is sufficient. This RFC adds coverage and strict
  references without a second baseline grammar or any invented later-tier mechanic.
