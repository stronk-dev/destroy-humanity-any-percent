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

## 2026-07-29 (claude — independent review of 9d9cc5a..61c7582: APPROVED)

- **The matrix is strict in every direction:** unknown policy/milestone/statistic rejected,
  duplicate and incomplete matrices rejected (six negative fixtures), bounds validated,
  baseline flowed through the hardened guard as an isolated `BALANCE-CHANGE:` commit with
  constants identity moving and a T0 drift regression pinned. Suites green.
- **One observation, not a finding:** `casual.phase0`'s p50 ≡ p95 on every milestone — the
  policy is seed-independent, so 300 runs of it are 300 identical runs; the run count buys
  distribution coverage for `chaos` only. Correct behavior, worth remembering when reading the
  baseline: casual columns are point values, not distributions, until a seed-sensitive persona
  (Speedrunner/Idler) joins the matrix.
- These are toy-catalog scaffolding numbers, not game pacing — their job is drift detection
  until real content arrives, and they now do it with zero warnings across 16 observations.

Clear to archive.

**Ruling on the Prestige contradiction Codex found (D4 ~15-min scripted Exit vs AC8's 45–90-min
envelope — both mine, mutually exclusive as written):** the envelope's referent is **the first
ELECTIVE Exit** — the first non-scripted Exit chosen by the player. The scripted first failure
is a fixed ~15-minute segment present in every route (per D4) and is excluded from the
first-Exit envelope while remaining inside total run time. AC8 amended accordingly; the same
clarification lands in `design/02 §3.1` ("first Exit available ~45–90 min" now reads "first
*elective* Exit"). The pacing gate measures what the player *decides*; the scripted beat is
curriculum, not pacing.
