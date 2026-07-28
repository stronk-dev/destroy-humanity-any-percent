# RFC: The Balance Harness

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Design refs:** `design/02 §11` (pacing targets), `design/00` (a designed ending; hardcaps), `design/07` (phase gates)
- **Research:** `design/research/pacing-science.md` (the four-tier gate stack, personas, envelopes — the primary source), `design/research/cicd-deploy.md §3` (measured feasibility: N=1000×30d = 6.96 s on 14 cores), `design/research/adaptive-balancing.md` (simulator-as-inference, persona-mixture ABC)
- **Depends on:** Production Engine (implemented — the sim kernel *is* the engine), Geometric Fast Path (implemented — the 95× fix that makes gates affordable), CI Baseline (implementing — tiers 3/4 land in its reserved trigger paths)
- **Planning:** `planning/balance-harness/` (once implementing)

## Summary

The genre has no prior art here — Antimatter Dimensions ships separated balance data and zero tests. This RFC builds the thing the whole corpus keeps promising: **bot personas driving the real engine through the real intent API**, gate tiers wired into CI, and the pacing curve as a reviewable artifact on every balance PR. It is also the **aggregation point for the acceptance criteria other RFCs have already lodged here by name**: the 200-bot × 30-day chaos gate (Production C8), the anti-Nash-1 combat gate (Combat AC3), the Depletion-unreachability proof (Gate Predicates D4), and commons population-invariance (Commons AC6).

## Specification

### D1 — Personas are policies over the intent API, nothing else

Six personas, each a deterministic seeded policy emitting only the public intents (`buy_generator`, `perform_manual_batch`, and future intents as their RFCs land): **Chaos** (fuzz-weighted random) · **Casual** (2 sessions/day, 20 min, buys greedily) · **Speedrunner** (optimal-known-route, max-affordable spam, active windows) · **Idler** (one action/day, banks credits) · **Optimizer** (marginal-efficiency purchasing) · **Ethical** (refuses dark-multiplier sources; commons-dependent once the compact ships). Personas share the server's closed forms via the engine itself — **the harness never reimplements production math**; it drives it. Persona definitions are versioned data; changing one is a `BALANCE-CHANGE:`-class event because it silently moves every envelope.

### D2 — The gate tiers (pacing-science's stack, now normative)

| Tier | Trigger | Blocking | Contents |
|---|---|---|---|
| **H1 Correctness** | every PR (CI tier 1 budget) | **yes** | Chaos × 200 seeds × 30 simulated days: zero NaN/negative/soft-lock, ledger balances, every declared gate reachable, golden-seed determinism |
| **H2 Envelopes** | PR touching `balance/**` | **yes** | Persona completion times inside declared envelopes (provisional, from research: Casual T1 ∈ [45, 90] min · T2 ∈ [7, 14] d · T3 ∈ [50, 70] d; Speedrunner-best T3 ∈ [1.5, 3] h; Idler ≤ 4× Casual); per-stage wall budgets; **gate on best-of-N, not means** (Roohi) |
| **H3 Drift** | PR touching `balance/**` | warn/block | Diff all persona times vs the checked-in baseline: **warn > 10%, block > 25% unless the PR declares `BALANCE-CHANGE:`** — which mints/amends an epoch (Leaderboard RFC) and regenerates the baseline |
| **H4 Nightly** | schedule | no | N = 1000 mixed-persona runs; Bayesian search over declared knobs against declared targets (**human approval applies changes — the search proposes, it never commits**); strategy-diversity audit (≥ 2 materially different routes reach T3 within 20%); contributed gates: anti-Nash-1, Depletion-unreachability, commons invariance |

**The artifact rule:** every H2/H3 run publishes **the pacing-curve picture** (all personas' progress coordinates over simulated time, overlaid on the baseline) as a PR artifact. A reviewer sees the shape, not a number table.

### D3 — Determinism & the arch question

Runs are `(catalog constants_hash, persona set version, seed)` → byte-identical event streams — this is what makes H3's diff meaningful. **The open cross-architecture float question (CI RFC) is resolved *here*, empirically, as the first implementation task:** run the golden-seed suite on amd64 and arm64; if byte-identity fails, H1's determinism gate downgrades to per-arch baselines **by explicit amendment**, never silently.

### D4 — Envelope data & the inference seam

Envelopes, wall budgets, and knob declarations are balance data (schema-validated like everything else). The simulator-as-inference-engine work (`adaptive-balancing §5`: ABC over constants *and persona-mixture weights*) is a **named follow-up** once live telemetry exists — the harness ships the forward simulator it will need, and the `subProgressValue` coordinate is already its y-axis.

## Acceptance criteria

1. H1 runs inside the CI tier-1 budget on hosted runners (the measured 6.96 s kernel figure, with full-engine overhead, stays under 60 s for the PR-blocking slice).
2. A deliberately broken catalog (soft-lock at T1) fails H1 with the seed and event-stream tail in the failure output.
3. A 3× cost-curve change without `BALANCE-CHANGE:` blocks at H3; with the tag, it regenerates the baseline and the epoch hook fires.
4. The pacing-curve artifact renders on a real PR.
5. Golden-seed byte-identity holds on the CI arch; the amd64/arm64 comparison has a recorded verdict either way.
6. The three contributed gates (anti-Nash-1, Depletion-unreachability, commons invariance) run in H4 and fail on fixture catalogs constructed to violate each.

## Open questions

- Envelope values: provisional, explicitly expected to move when H2 first runs against the real Phase-0 catalog — the first honest measurement replaces the research estimate.
- Persona policy sophistication (Speedrunner's "optimal known route") grows with the Route Registry; v1 uses scripted routes.

## Changelog

- 2026-07-28: created (draft).
