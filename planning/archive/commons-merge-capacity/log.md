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

## 2026-07-29 (claude — independent review of 0df79d9..baf4d43: APPROVED)

Both rulings implemented faithfully; parity verified in Go + TS + schema; suites green.

- **Chronology:** `gateTier < transitionTier → reject`, identical integer rule in both runtimes,
  with non-canonical gate/transition IDs on doctrine-bearing routes refused outright. The three
  seeds moved to `gate.t4_to_t5` — strictly after their doctrine's transition, sidestepping the
  equality case. **Routed forward, not a blocker: the `gateTier == transitionTier` binding
  (allowed by this rule, and C1's original example) depends on whether the doctrine pick
  evaluates before the same crossing's gate predicate — pin that ordering in the future
  doctrine-intent RFC before any same-gate route ships.** The TS loader's
  `maxRoutesPerRun() >= depletion → throw` confirms the proof still binds after the moves.
- **Merges:** floor-triggered only, whole-cohort moves across assignments/memberships/samples in
  one transaction, source zeroed and closed, cap enforced. `floor(1.5×target)` vs the ruling's
  ceiling: accepted deviation — conservative direction, differs only on odd halves.
- **baf4d43:** the merge-cap overflow was again self-found by Codex's adversarial pass before
  reaching this gate.
- The `BALANCE-CHANGE:` artifact regeneration flowed through the hardened guard: constants
  identity moved, pacing metrics unchanged — exactly the shape the guard exists to certify.

Both ruling RFCs clear to archive.
