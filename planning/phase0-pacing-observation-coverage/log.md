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

## 2026-07-29 — artifact and local review

- The standalone `BALANCE-CHANGE:` regeneration produced 16 aggregate values and changed only the
  golden-seed and pacing-baseline artifacts. The hardened history guard accepts the source-first
  commit followed by the artifact-only commit.
- A named regression proves T0-progress p50 now reaches the existing integer 10% warning and 25%
  failure paths; this was not present in the two-value baseline.
- Local spec/adversarial review found no untracked milestone tuple or second aggregation authority.
  Independent diff review remains the archival gate.

## 2026-07-29 — full verification

- The clean 300-run check passed with all 16 aggregate values, no drift, and zero invariant
  failures. `make harness-check` also validated the complete commit history through the hardened
  baseline guard.
- Full real-Postgres `make verify` passed: Go/vet/integration, formula generation, Commons
  population invariance, harness, TypeScript diagnostics/build, schema semantics, 6,412 client
  tests, and 19,245 browser tests.
- Implementation is complete. The RFC remains `implementing` solely for its mandatory independent
  diff review.
