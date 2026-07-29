# Commons Cohort Merge Capacity — append-only log

## 2026-07-29 — start

- Owner ruled that only member count below the merge floor triggers collapse handling. Health may
  recover independently and does not initiate a population move.
- The surviving cohort may reach but never exceed `floor(1.5 × cohort_target_size)`. Each source
  is atomic: merge all members or skip it; never split.

## 2026-07-29 — implementation

- `canMergeCohort` is the single eligibility decision: source count strictly below the floor and
  whole-source post-merge count at or below the integer `3*target/2` ceiling. Health is absent.
- Unit cases cover exact 1.5×, ceiling+1, floor equality, and an ordinary eligible source.
- The real-Postgres projection test verifies ceiling+1 moves no Founder assignment and does not
  close the source; reducing that same source to land exactly at 225 merges it whole. A 40-member
  source remains open with target capacity available.
- Canonical Commons documentation now states the floor-only trigger, 1.5× cap, and never-split
  behavior. Unit and integration tests are green.

## 2026-07-29 — local adversarial review

- Review found that spelling the ceiling as `target_size * 3 / 2` could overflow a platform-sized
  integer before applying the cap. The helper now computes `target + floor(target/2)`, guards the
  addition, and compares with subtraction so both policy and member totals fail closed on overflow.
- A 64-bit boundary regression proves the fail-closed behavior; 32-bit builds skip only that
  unreachable platform-sized fixture. Independent review remains the mandatory archival gate.
