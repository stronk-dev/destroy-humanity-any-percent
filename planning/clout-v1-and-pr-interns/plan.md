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

- [x] P1 (`ace4e67e`, formulas `80f1bd47`) — CV1 economy schema v5 (Go `server/economy`, TS `client/src/economy-kernel.ts`):
  `axis_stack` block, the axis upgrade-effect arm, the upgrade-only `axis_at_least` requirement,
  the `axis_stack` multiplier slot (after `milestones`), provider/declaration pairing, and
  cross-artifact validation against `achievements`. Shared loader corpus. AC1.
- [x] P2 (`f256b235`, formulas `63aa0b62`) — CV3 formula: axis contributions from Company state, the clamp, `saturated`; shared
  Go-authored vectors reproduced by TS. AC2. Formulas artifact regenerated in its own commit
  (AC10).
- [x] P3 (`f256b235`) — implementation presence, not full acceptance: CV2/CV4/CV5 Company save v19 (next free) attainment set + derived score, second hook
  pass (Go + TS), `achievement_reattained.v1` event + migration, new-run activation, Exit discard,
  migration corpus. ACs 3, 5, 6, 7, 8.
  RP-307: AC6's seeded timing/partition property and retroactive-factor failing case
  are absent from the original population; formula/helper parity is not that proof.
- [x] P4 (`4c089f00`) — AC4 Gaia-law structural test (empty writer set under A; seeded writer fails).
- [x] P5 (`2f2c5a04`) — CV9 snapshot producer (optional v4 field) + Desk PR row progress: implementation presence, not full AC11/default composed journey acceptance. RP-303 locally corrects native progress associations; its Codex span independently needs Claude.
- [ ] P6 — PARTIAL (see log P6; blocked on harness attainment evaluation) — CV10 harness: scenario bundle rejection without achievements, relevance mask, dead-row
  fixture, observation + `axis_input_within_cap` invariant. AC9.
- [ ] P7 — Canon docs (AC12). Partial: `docs/axis-stack.md` and pointers landed; RFC index/manifest G09 rows and the final range are owed at completion.
- [ ] Designated Codex review, archival (Codex).

Kernel protocol: every commit touching a `kernel/affecting-paths.json` prefix bumps
`kernel/VERSION` (+ Go/TS constants) in the same commit. Save/snapshot/migration numbers are
assigned in landing order (the RFC's "v19" means next-free Company version).

## RP-303 — accessible PR progress, bounded consumer correction

Predeclared at b2b2d5a5, 2026-10-06. First reproduce the unnamed native progress
bar through a browser role/name query, without changing the production panel.
Test both unowned interns together, then a snapshot replacement with one owned:
exact name-to-row association, numeric value/maximum, description and remaining
bar must hold; no gameplay intent may be emitted. Negative controls remove the
label association and point it at the wrong row; both must fail after the fix.
Use existing copy keys only. No server/schema/arithmetic/kernel/epoch/CI changes.
Run cold Chromium/WebKit cases via root Make, type/client/build/boundary/copy
checks; report any Firefox launch failure separately, never as a passed case.
No checkboxes or full P5/AC11 promotion. New Codex implementation needs Claude.

Executed: unchanged-production failure and both compiling label controls
discriminate; restored Chromium/WebKit six cases plus the independent performance
lane pass. Types/build/client/boundary/copy/topology pass; Firefox zero executed
after session timeout, not waived. Full evidence and review range are in log.md.

## RP-304 — mandatory roles on axis upgrades

Predeclared at b30808e5, 2026-10-06. Review Claude ace4e67e^..ace4e67e's
Go/TS axis-role branches against CV1 item5. Existing branch loops check role
identity/duplicates but not cardinality. New separate shared cases must reproduce
empty roles on either/both PR rows; unchanged fixture, one declared role, and
legacy static-empty-role cases are controls. Original corpus stays unchanged.
If confirmed, reject empty roles only for axis effects, preserving static catalog
semantics. Watched loader changes require honest kernel161→162 in the same commit.
Independent compiling guard omissions in both runtimes must fail, then restore.
Run cold relevant packages/vet, full client/types/build and applicable root checks.
No schema/balance/copy/mint/role-vocabulary/pool retune or whole-P1 approval.
Role names alone do not prove executed pool binding; that question stays separate.
Every new Codex implementation/record edge requires Claude independently.

Executed: six-case paired corpus reproduces three illegal admissions per
runtime; both narrow guards fix them and independent omissions fail again.
Kernel162 accompanies the real acceptance-set change. Final cold relevant Go/
vet,8248units/types/build/formulas/boundaries/topology,80 native Chromium/WebKit
cases and real-Postgres production/gameui packages pass. Full historical kernel
guard is still red at50a3a514 (RP-131), not bypassed. Full new span afterb30808e5
including records requires Claude. Truthful pool binding remains D-022/RP-305;
no full P1/AC1, hosted CI or mint acceptance follows.

## RP-307 — actual timing and partition audit

Predeclared at07bb3d9d, 2026-10-07. Test-only first: actual ApplyLogged
generator-purchase transition on an admitted v19 fixture, with an owned PR Intern
and derivation-valid x=6. Check the interval preceding re-attainment uses x=6;
the purchase raises x to8 and only the subsequent interval uses x=8. Assert
actual ledger/debit, ordered achievement events and unchanged independent pins.
Use seeded nonzero intervals and online/offline accrual arms. Separately compare
full encoded Company states after one-shot versus seeded millisecond partitions
of the same axis-enabled interval, using actual Evaluate and pinned contributions.
Do not claim extra evaluations are extra player intents or full composed/SQL proof.

The oracle is exact, not a tolerance. Include whole-second and tiny-cut controls;
any numeric divergence is a finding, never justification to narrow AC6. Before
runtime faults, the unchanged source must execute; a retroactive-factor fault is
allowed only as a temporary compiling observation, with exact-byte restoration.
Do not change runtime/save/kernel/balance/copy/CI or owner-authored specifications
in this range. If proof requires a new persistence/numeric contract, route it to
the author/owner rather than inventing one. All new Codex test/record edges require
Claude; no checkbox flip, full P3/AC6, mint or archival approval follows.
