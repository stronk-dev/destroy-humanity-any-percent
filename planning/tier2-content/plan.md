# Tier 2 Content — implementation plan

RFC: `rfc/tier2-content.md` (accepted 2026-09-25, recommended defaults). **OD-1/HG-1 (seat
source) is NOT ruled.** Everything in §H (headcount grammar, state, intent, projection, masks),
the `headcount_seats` generator roles, the allocation policy arms, `milestone.first_headcount_allocation`,
the `headcount_budget_respected` invariant, the P4 `headcount_seats` role row and the P5 gate are
held until it is. Fixture-first: candidates live under `balance/testdata/t2/`; epoch 9 is NOT minted
(M1–M4 owner-gated).

- [x] C1 — Candidate artifacts (§A1, §A2, §B1 minus `headcount_seats`, §B2, §B3): economy
  (schema v4 carry; v5 is headcount-only), routes, categories under `balance/testdata/t2/`,
  derived from the epoch-8 bytes by insertion only, with a SHA file.
- [x] C2 — Strict loaders accept the candidates in Go and TS; categories missing
  `gate.t1_to_t2` is rejected against candidate routes (AC1 subset).
- [x] C3 — AC0 reachability: a Tier-1 company on the candidate bundle crosses `gate.t1_to_t2`,
  gets `Tier == 2`, and `incorporate` applies; the same on the epoch-8 bundle rejects.
- [x] C4 — UI (§E1–E3 minus Headcount panel): `era_2010` theme/era, tier-2 era mapping (tier ≥ 3
  still throws), Gate control for any projected gate, minimal Incorporate control, FarmVille
  energy-bar chrome stub with curtain; candidate copy (§E4 keys).
- [~] C5 — Harness policy registry v2 subset (§P1 `exit_rule: t01_c32_readiness_once`,
  `offer_rule: ignore`, T2 gate crossing) and the T2 pacing scenario; measurement report for the
  `gate.t1_to_t2` literal against the [2h,3h] envelope (§P2, OD-7/OD-8).
- [ ] C6 — (next; see log 2026-09-25 C5 "Not done") Relevance: T0 identity vs epoch 8 on the candidate bundle; T1–T2 combined scenario
  report (§P3).
- [ ] Held — §H headcount, P4 seats row, P5, AC2/AC3 (headcount parity/partition) — OD-1.
- [ ] Owner — M3 SHA ratification, E4 copy round, M1–M4 mint.
