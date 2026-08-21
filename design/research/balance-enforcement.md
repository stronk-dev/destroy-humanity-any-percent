# Formal detection and enforcement of purchasable relevance — research report

> Banked 2026-08-01. Provenance axis: **[V]** verified this session (bibliographic record and/or
> content fetched) · **[M]** model knowledge, high confidence, not re-verified · **[S]**
> speculation/synthesis. Discovery ran through bibliographic APIs (arXiv, Crossref, DBLP, Semantic
> Scholar, GitHub) plus targeted fetches; absence claims are correspondingly weak.
> Companion: `tier-relevance.md` (mechanism survey, banked separately).

## 0. The correct theoretical frame

An idle game under a scripted or optimal policy is **not** an adversarial game — it is a
deterministic single-agent **optimal-control / shortest-path problem**: state = (bank, rates,
owned multiset, unlocks), actions = {buy p, wait}, objective = minimize time-to-milestone.
"Purchasable p is meaningful" then has a precise definition: *p lies on optimal (or ε-optimal, or
persona-typical) paths, and deleting p from the action set worsens the optimal value V\* by at
least ε during p's intended window.* Alpha-beta / Nash tooling is the wrong import; the right
imports are restricted-play evaluation, DP/greedy solver theory, and simulation-driven parameter
search.

## 1. Marginal-value / counterfactual attribution

- **Restricted play is the direct prior art for counterfactual deletion.** Jaffe, Miller,
  Andersen, Liu, Karlin, Popović, "Evaluating Competitive Game Balance with Restricted Play"
  (AIIDE 2012, doi:10.1609/aiide.v8i1.12513) **[V]**: measure an element's impact by forbidding it
  and measuring the outcome delta. Single-player transplant: forbid purchasable p, re-solve/
  replay, measure Δ(time-to-milestone). Jaffe's UW dissertation expands this into a metric family
  **[M]**.
- **Per-action value attribution (VAEP analog).** Decroos, Bransen, Van Haaren, Davis, "Actions
  Speak Louder than Goals" (KDD 2019, doi:10.1145/3292500.3330758) **[V]**: value each atomic
  action as the change in expected outcome. Idle analog: each purchase valued as ΔV(state) under a
  value estimate — O(1) per purchase; a cheap tier-1 attribution.
- **Shapley values: no published application to game-content attribution found** (searches
  surfaced team/player contribution and SHAP churn explanation only). **[S]** Exact Shapley over
  N≈hundreds of purchasables is 2^N runs — infeasible; Monte-Carlo permutation sampling (Castro et
  al. 2009 **[M]**) needs ~10³–10⁴ full runs per stable estimate — feasible only because the
  harness is deterministic and parallel. **Why ever pay it:** leave-one-out (LOO) systematically
  undervalues *redundant* content — two near-substitutes each show ≈0 LOO delta while jointly
  essential. Practical middle ground: LOO per purchasable + **group ablations** (delete a whole
  tier/category) to expose redundancy; sampled Shapley only for flagged clusters.
- **Greedy marginal metrics are community-standard.** The Cookie Monster addon
  (github.com/CookieMonsterTeam/CookieMonster) **[V]** ranks purchases by **Payback Period**:
  `max(cost − bank, 0)/cps + cost/Δcps`, including indirect effects, color-coded. PP-at-purchase
  is a cheap, well-understood instantaneous marginal value to log per purchase.

## 2. Optimal-play solvers for incrementals

- **The academic anchor: Demaine, Ito, Langerman, Lynch, Rudoy, Xiao, "Cookie Clicker"
  (arXiv:1808.07540, 2018, cs.GT) [V, content read].** Results: (1) every optimal strategy is
  **buy-phase-then-wait-phase** (never delay an intended purchase); (2) with fixed costs, optimal
  order is monotone blocks with an efficiency metric and a **threshold T where item preference
  switches**; (3) exact optima via DP in k-item cases; (4) **greedy achieves approximation ratio
  1+O(1/log M)** for large targets; (5) hardness: rate-goal with growing costs and nonzero
  starting bank are **weakly NP-hard** (Partition); the **discrete-timestep version is strongly
  NP-hard** (3-Partition). Implication: the tick-based sim is the strongly-NP-hard regime — an
  exact optimal-buy-order oracle is off the table in general; the approximation bound justifies a
  **greedy/PP-driven reference persona as the "approximately optimal" baseline**. (The name
  "Almanza et al." is a misattribution; no such clicker paper exists on arXiv [V, absence].)
- **The Math of Idle Games I–III**, Pecorella (Kongregate 2016–17, republished on Game Developer)
  **[V parts I and III]**: exponential cost vs polynomial production, walls, optimal play as best
  income-per-cost purchasing; **staggered multiplier thresholds so older generators keep becoming
  dominant again** — a design-side relevance mechanism; Part III surveys prestige formulas with
  companion spreadsheets and warns models are rough-feel tools.
- **LLM-era playtesting:** Mu, Cai, Lu, Zhang, Tei, Li, "Knowledge Graph-enhanced LLM for
  Incremental Game PlayTesting" (arXiv:2511.02534, 2025) **[V]** — KG-modeled dependencies driving
  LLM playtesting; directly on-genre.
- **Community route optimizers** are per-game and mostly small (skyllic/TheorySim,
  Belvenix/IdleonCogOptimizer genetic algorithm, Levijom/CookieClicker — all **[V]** via GitHub;
  Frozen Cookies efficiency metric **[M]**). Communities converge on greedy payback + restricted
  brute-force/DP — corroborating the scripted-greedy reference policy as practical state of the art.
- **When is greedy optimal? [S, grounded in Demaine]:** near-optimal iff (i) single currency,
  (ii) purchases only add rate (no synergies/unlocks), (iii) continuous time, (iv)
  immediate-purchase holds. Every feature breaking these (tick discreteness, cross-tier
  multipliers, unlock gates, timed events, multi-currency) is a divergence site — worth an
  explicit harness experiment (greedy vs beam-search/DP persona on small catalogs), not an
  assumption.

## 3. Automated / simulation-based balance tuning

- **Machinations** (machinations.io) **[V]** — visual internal-economy graphs with batch/
  Monte-Carlo simulation; academic root Dormans, "Simulating Mechanics to Study Emergence in
  Games" (AIIDE workshop 2011, doi:10.1609/aiide.v7i3.12477) **[V]**; Adams & Dormans 2012 **[M]**.
  The Go harness is a higher-fidelity Machinations.
- **Search-based auto-balancing:** Volz, Rudolph, Naujoks (GECCO 2016,
  doi:10.1145/2908812.2908913) **[V]** — surrogate-assisted evolutionary balancing of Top Trumps
  against simulated play. Isaksen, Gopstein, Togelius, Nealen (IEEE ToG 2018,
  doi:10.1109/tciaig.2017.2750181) **[V]** — parameter sweeps + survival analysis over simulated
  players (maps onto persona time-to-wall curves).
- **Closest to catalog-constant tuning: Rupp & Eckert, "GEEvo: Game Economy Generation and
  Balancing with Evolutionary Algorithms" (IEEE CEC 2024, doi:10.1109/cec60901.2024.10612054;
  arXiv:2404.18574) [V]** — graph-based economies re-balanced by EA against simulation-derived
  fitness, game-logic-agnostic. Companions: RL level balancing (CoG 2023
  doi:10.1109/cog57401.2023.10333248; IEEE ToG 2024 doi:10.1109/tg.2024.3399536), "Level the
  Level" per-archetype balancing (FDG 2025, doi:10.1145/3723498.3723747) **[all V]**.
- **Persona-driven playtesting precedent:** Holmgård, Green, Liapis, Togelius, procedural
  personas via MCTS+evolved heuristics (IEEE ToG 2019) **[V]**; Gudmundsson et al. (King),
  "Human-Like Playtesting with Deep Learning" (CIG 2018, doi:10.1109/cig.2018.8490442) **[V]** —
  production pipeline predicting Candy Crush difficulty pre-release from agent rollouts (the
  strongest industry precedent for simulation-gated content shipping); Pfau et al., "Dungeons &
  Replicants" (CoG 2020 doi:10.1109/cog47356.2020.9231958; II — IEEE ToG 2023
  doi:10.1109/tg.2022.3167728) **[V]** — behavior clones auto-balancing MMORPG dimensions; Zhao et
  al. (EA), "Winning Is Not Everything" (IEEE ToG 2020, doi:10.1109/tg.2020.2990865) **[V]**;
  Pfau, "Progression Balancing × Baldur's Gate 3" (CHI 2025, doi:10.1145/3706598.3713162) **[V]**;
  González-Ortega et al. (SCCC 2025, doi:10.1109/SCCC67219.2025.11420422) **[V, listing only]**.
- **Constraint/declarative:** Smith & Mateas, ASP for PCG (TCIAIG 2011,
  doi:10.1109/tciaig.2011.2158545) **[V]** — invariants as logic constraints; Butler, Andersen,
  Smith, Gulwani, Popović, "Automatic Game Progression Design through Analysis of Solution
  Features" (CHI 2015, doi:10.1145/2702123.2702330) **[V]** — prior art for tier/curriculum
  ordering. **No published LP/SMT solve of idle-curve constants against marginal-value invariants
  found [V, absence]** — a genuine gap. **[S]** For geometric cost/production families, "PP(p) is
  within top-k at some reachable state in p's window" is piecewise-analytic in the exponents; a
  direct solver or CMA-ES-with-penalties over log-parameters is plausible and likely novel.
- **Field definition:** Becker & Görlich, "What is Game Balancing?" (ParadigmPlus 2020,
  doi:10.55969/paradigmplus.v1n1a2) **[V]** — "balance" lacks a shared formal definition;
  motivates defining relevance formally.

## 4. Dominance / dead-strategy detection

- Framed as optimal control, a **dominated purchasable** = LOO delta ≈ 0 for *every* persona
  including the near-optimal reference — restricted play evaluated over a policy portfolio, not
  adversaries. EGTA/best-response sweeps (Wellman **[M]**) are the multi-agent analog; in
  single-player they collapse to the counterfactual machinery. RL action-elimination (Even-Dar et
  al. **[M]**) is the MDP-side analog.
- **[S] Two-axis relevance (per the Riot precedent):** (a) *prescriptive* — the near-optimal
  reference buys p and loses ≥ε without it; (b) *descriptive* — scripted personas buy p and
  benefit. Content can pass one and fail the other (trap upgrade vs crutch); gate on the
  combination.

## 5. Metrics precedents from live games

- **Riot's League balance framework** (2020 dev post) **[V]**: 148 champions across four segments
  (average/skilled/elite/pro), win-rate-conditioned-on-ban-rate bands; **nerf if over-threshold in
  ANY segment, buff only if under-threshold in ALL segments** — a shipped, public example of
  per-content statistical floors with segment quantifiers; personas stand in for skill segments.
- Hearthstone deck-inclusion rates, MOBA pick/ban dashboards as content-utilization instruments
  **[M]**.
- **No verified public case of an idle game shipping telemetry-driven relevance floors** [V,
  absence — weak]. A deterministic-harness CI relevance gate would be without direct precedent.

## 6. The enforcement shape for Cloud Clicker (synthesis, [S] except cited anchors)

1. **Per-purchasable relevance report** (per persona × window): LOO counterfactual
   Δtime-to-milestone (restricted play); purchase share and time-of-purchase across personas;
   PP-at-purchase as instantaneous marginal value; **tier/category group-ablation deltas** for
   substitute-redundancy; sampled Shapley only on flagged clusters. Determinism makes every number
   byte-reproducible — the report is a golden artifact.
2. **CI gate**: for every purchasable p with window W: (i) ∃ persona with LOO delta ≥ ε within W
   (matters to someone), and (ii) the reference near-greedy persona buys p at all (not a trap) —
   the ANY/ALL asymmetry borrowed from Riot. Precedents for simulation-gated shipping exist
   (King, Pfau); a CI-enforced version is novel.
3. **Parameter search to satisfy floors**: GEEvo is the closest published pipeline — EA with
   constraint-violation penalties over relevance floors, CMA-ES over log-space catalog constants.
   Determinism removes fitness noise — the hardest part of every published setup. Demaine's
   hardness says don't chase the exact oracle; his bound says greedy reference play is defensible.

**Honest gaps:** no published Shapley-for-content; no LP/SMT curve solving against marginal-value
invariants; no idle game with published relevance floors. Those three are where Cloud Clicker
would be stating new practice rather than citing it.
