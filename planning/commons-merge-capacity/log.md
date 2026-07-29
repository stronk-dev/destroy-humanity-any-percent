# Commons Cohort Merge Capacity — append-only log

## 2026-07-29 — start

- Owner ruled that only member count below the merge floor triggers collapse handling. Health may
  recover independently and does not initiate a population move.
- The surviving cohort may reach but never exceed `floor(1.5 × cohort_target_size)`. Each source
  is atomic: merge all members or skip it; never split.
