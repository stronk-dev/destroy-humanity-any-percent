# Clout v1 + PR Interns implementation plan

RFC: `rfc/clout-v1-and-pr-interns.md` (accepted 2026-09-25; every owner decision at its
recommended default: OD-1 = A, OD-2 = `achievement_attainment_run`, OD-3 = one-time cash upgrades
through `buy_upgrade`, OD-4 = social mint deferred, OD-6 = slot after `milestones`, OD-9 = Clout line
withheld). Fixture-first: no production epoch is minted by this plan; the CV8 numbers are proposals
ratified later by owner SHA (OD-5).

Under A, CV6 (the Clout ledger) and CV7 (the social seam) are **not** implemented: the RFC's CV0
table marks the ledger "no" for A, and the acceptance rejects B/C. The Gaia-law test (AC4) still
lands, with the allowed writer set empty.

## Batches

- [ ] P1 — CV1 economy schema v5 (Go `server/economy`, TS `client/src/economy-kernel.ts`):
  `axis_stack` block, the axis upgrade-effect arm, the upgrade-only `axis_at_least` requirement,
  the `axis_stack` multiplier slot (after `milestones`), provider/declaration pairing, and
  cross-artifact validation against `achievements`. Shared loader corpus. AC1.
- [ ] P2 — CV3 formula: axis contributions from Company state, the clamp, `saturated`; shared
  Go-authored vectors reproduced by TS. AC2. Formulas artifact regenerated in its own commit
  (AC10).
- [ ] P3 — CV2/CV4/CV5: Company save v19 (next free) attainment set + derived score, second hook
  pass (Go + TS), `achievement_reattained.v1` event + migration, new-run activation, Exit discard,
  migration corpus. ACs 3, 5, 6, 7, 8.
- [ ] P4 — AC4 Gaia-law structural test (empty writer set under A; seeded writer fails).
- [ ] P5 — CV9 snapshot producer (optional v4 field) + Desk PR row progress. AC11.
- [ ] P6 — CV10 harness: scenario bundle rejection without achievements, relevance mask, dead-row
  fixture, observation + `axis_input_within_cap` invariant. AC9.
- [ ] P7 — Canon docs (AC12).
- [ ] Designated Codex review, archival (Codex).

Kernel protocol: every commit touching a `kernel/affecting-paths.json` prefix bumps
`kernel/VERSION` (+ Go/TS constants) in the same commit. Save/snapshot/migration numbers are
assigned in landing order (the RFC's "v19" means next-free Company version).
