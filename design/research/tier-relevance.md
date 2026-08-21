# Tier-Relevance Mechanisms in Idle/Incremental Games — Raw Research Report

> Banked 2026-08-01. Companion: `balance-enforcement.md` (formal detection/enforcement, banked
> same day). Method note: gathered via direct fetches of game wikis (MediaWiki APIs), game source
> code, Pecorella's republished articles, Paper Pilot's guide, Steam's API, and archived community
> discussions; unverified items are explicitly flagged. Community anecdotes are context, not
> representative evidence. Pecorella's model supports the baseline that the newest affordable
> generator usually dominates in a standard uncoupled economy.

Research question: what mechanisms keep every tier/generator/purchase relevant across the whole
arc, vs. the Cookie Clicker anti-pattern (only the newest 1–3 buildings matter).

## 1. Production-chain architectures (tier N produces tier N−1)

### 1.1 Swarm Simulator — the canonical chain
- Structure (paraphrased from the community wiki): Drones consume meat and larvae and produce
  meat; Queens create Drones; successive tiers consume lower tiers and increase downstream
  production. Drone = 1 meat/sec base, costs 10 meat + 1 larva; ~10 tiers per resource track.
- **Formula shape** (Pecorella II, on the identical Derivative Clicker architecture): buying tier
  m yields cascading integrals — bottom-tier output grows as `1 + t + t²/2! + … + tᵐ/m!`; with m
  tiers currency grows **polynomially ~ tᵐ**, approaching eˣ as tiers accrue. Steady state: every
  tier's count enters the leading coefficient as a product term — **no tier's count ever stops
  mattering, structurally**.
- Gating currencies force breadth: everything costs **larvae** (Hatcheries cost `300 × 10^X` —
  deliberately brutal base-10 superexponential; Expansions +10% larvae, `10 × 2.45^X`); military
  units produce **Territory** (each tier costs ×450 the meat of the previous, produces only ×45
  the territory).
- **Verdict:** structurally sound; individual low-tier purchases become transient noise once
  higher tiers exist — the game leans on bulk-buy and larva scarcity for decisions. Community:
  respected pioneer, with anecdotal reports of long late-game check-in intervals. **Failure mode: the chain guarantees
  relevance of COUNTS, not interestingness of PURCHASES** — growth is still dominated by the tᵐ
  term.

### 1.2 Antimatter Dimensions — chain + purchased-vs-generated split
- 8 Dimensions; each higher produces the one below at 0.1/sec base; pre-Infinity growth
  describes polynomial growth whose degree is the highest Dimension unlocked.
- **The key trick: only PURCHASED counts have costs and grant multipliers** — each 10 purchased ×2
  that dimension and bumps price (base costs 10…1e24; per-10 multipliers 1e3→1e15; past Infinity
  the multiplier itself grows ×10/purchase). Generated dimensions are free and priceless.
  The purchase decision becomes *which tier's ×2-per-10 to advance* — an all-tiers decision —
  while the chain keeps all counts relevant. Then Max All automates it (see failure).
- Layer recursion: Infinity Dimensions repeat the pattern; ID1 produces Infinity Power boosting
  ALL normal dimensions by `max(1, n⁷)`; Time Dimensions again. Each chain terminates in a
  currency exponentiated into a multiplier on the
  previous layer.
- **Failure mode** (Paper Pilot): repeatable IP/EP purchases can decay into number inflation,
  punish forgetfulness, then return as automation. Max All erases
  per-tier decisions — relevance preserved numerically, not experientially.

### 1.3 CIFI — shipped modern chain
Mk2 generates Mk1, etc., 8 tiers, backbone of a 2022+ live mobile game. Its community formulas
page describes the large majority of bonuses as multiplicative `c = aᵇ` with b a
count-of-something-owned —
counts-as-exponents everywhere.

### 1.4 Derivative Clicker / Shark Game / Spaceplan
Derivative Clicker: chain + **self-count boost** (each purchased tier-1 building +0.05% to all
tier-1 production — purchased/generated split again). Shark Game: multi-currency converter web.
Spaceplan: NOT verified this session; prior knowledge says devices obsolete narratively — treat as
anti-example, flagged.

## 2. Cross-tier multiplicative coupling

### 2.1 Cookie Clicker's own escape hatches (verified in main.js + wiki)
- **Thousand fingers line:** +0.1 cookies per NON-CURSOR building, then ×5, ×10, ×20 at 150, and
  ×20 again at every 50 cursors to 550 — cumulative `0.1 × 5 × 10 × 20⁹ = 2.56e12` per non-cursor
  building. The community wiki reports that the full upgrade line can restore cursors to
  late-game relevance.
  Count-of-other-tiers-as-base: buying ANY building buffs cursors.
- **Grandma coupling:** One Mind/Communal Brainsweep +0.02 CpS per grandma each (self-count
  quadratic); Elder Pact +0.05/Portal; dragon aura Elder Battalion +1% grandma CpS per non-grandma
  building; 18 grandma-type upgrades double grandmas AND boost the paired building per grandma
  (paired-% from prior knowledge, flagged).
- **Synergies Vol. I/II** (need 15/75 of both): **older building +5% CpS per newer; newer +0.1%
  per older** — verified across all pairs. Deliberately asymmetric toward the older building.
- **Verdict:** rescues cursors and grandmas specifically; mid-tiers stay dead except as
  synergy-pair donors and achievement fodder. A retrofit patch, not an architecture.

### 2.2 Synergism — dense cross-system multiplication
Coin producers → Diamond buildings (produce Crystals multiplying coin production) → Mythos →
Particle buildings; each layer's buildings produce a currency multiplying layers below. Coupling
routes through two global pools — **Accelerators and Multipliers** — fed by counts of everything
(verified upgrade texts: free Multiplier per 160 Coin producers; free Accelerator per 80; per
2,000; cross-feeds between Accelerators and Multipliers; crystal producers +0.1%/purchased).
**Architecture: every count feeds shared global stats.** Failure mode: opacity — Paper Pilot's
analysis warns that nested multiplier chains can obscure player agency.

### 2.3 Generalized pattern
Four verified forms: (a) pairwise synergies (CC, +5%/+0.1%), (b) count-as-base (CC fingers),
(c) pooled stats (Synergism), (d) inter-layer exponentiation (AD n⁷). Realm Grinder's Fairy
Heritage is (b) at faction scale: **+0.075% ALL production per Farm/Inn/Blacksmith owned** —
tier-1 buildings stay the faction's scaling engine.

## 3. Count-as-input architectures

- **Staggered milestone multipliers** — Pecorella recommends varying per-generator thresholds so
  the best next purchase rotates across tiers. Shipped: AdCap (25/50/100…), CC tiered upgrades
  (1/5/25/50/100/150/…/550 owned), AD (per-10).
- **CC achievement→milk→kittens** (verified in source): milk = 4%/achievement (622, max 2,488%);
  kittens multiply total CpS by `(1 + milk × f)`, f = 0.1, 0.125, 0.15, 0.175, 0.2×5, 0.05,
  0.105/0.11/0.115. ~¾ of achievements are own-N/bake-N-with-X → dead buildings convert to
  permanent global CpS. Confirmed as CC's only early-tier redemption, with numbers.
- **Realm Grinder:** 900 trophies, **566 building trophies** (1→20,000 of each of 22 types +
  total-buildings to 300k); Secret trophies grant upgrades. Faction upgrades re-weight which tiers
  matter per run (Fairy: T1 +600 base, mana regen `+0.3·x^0.3` over T1 counts, `+250·x^0.9%`
  production) — tier relevance as **strategic choice per prestige**.
- **Gnorp Apologue** (no wiki; Steam + community verified): shards generated then physically
  collected; ~15 unit types with hard interactions; Talent Stone; prestige = full re-spec.
  Over-investing in one unit backfires; you deliberately don't buy everything; finishable 30–40 h
  with several effective strategies; community contrasts it favorably against CC. Mechanism:
  **build identity** — each purchase is a slot in a composition, not a ladder rung.

## 4. Role-differentiation (production stops, job remains)

- **Kittens Game** (verified in buildings.js): Barn priceRatio **1.75** grants storage caps
  (catnip 5000, wood 200, minerals 250, coal 60, iron 50, titanium 2, gold 10); Warehouse 1.15
  storage-only; Library/Academy/Observatory science+culture caps; Huts manpower+kittens.
  Production buildings ratio 1.10–1.15, storage up to 1.75. Research/crafting are capacity-gated →
  **storage buildings are mandatory forever; every building carries a distinct effect bundle —
  near-zero dead content by role design.**
- **Evolve** (verified in src/resources.js:2644-2671): crateValue 350→500→+, containerValue
  800→1200; crates/containers player-allocated per resource; dedicated buildings exist only to add
  slots. Storage allocation is an active decision across the whole multi-prestige arc.
- **Sacrifice/consumption roles:** AD Dimensional Sacrifice (consumes non-bought D1–D7; D8
  multiplier ×= `max(log₁₀(n)/10,1)^m`, m=2 base; post-IC2 `n^m`, m=1/120) — a lower-tier
  population as FUEL. CC Krumblor (sacrifice 100 grandmas for Elder Battalion); garden sacrifice
  (yield from memory, flagged). **CC sugar-lump building levels give buildings second jobs**:
  cursor level = stock-market capacity + loan slots; grandma level/count = cheaper stockbrokers —
  buildings as minigame inputs after CpS irrelevance (verified).
- **Prestige bases** (all verified, Pecorella III): RG `p=(√(1+8·c_max/10¹²)−1)/2` (~4× to
  double); AdCap `150·√(c_life/10¹⁵)` (~3–4×); CC `∛(c_life/10¹²)` (~8×); Egg Inc
  `Δp=(c_run/10⁶)^0.14` (~128×, run-only → forces active rebuilds); Clicker Heroes per-purchase.
  Lifetime-based prestige keeps ALL historical production relevant; run-based makes early tiers
  relevant again every run.
- **Melvor mastery** (verified): per-item mastery; MXP/action formula over unlocked actions +
  item level; total mastery feeds drops and per-recipe bonuses — count-as-input applied to
  ACTIONS; old recipes permanently relevant as mastery targets.

## 5. Normalization / percentage schemes

**No shipped game found computing total output as weighted geometric mean, softmax share, or
log-domain sum.** What ships instead: (1) chains — the geometric-mean property achieved
kinetically (product over tiers; no factor can vanish); (2) counts-as-exponents (CIFI ~90%);
(3) pooled-stat coupling. Paper Pilot argues that multiplicative bonuses can extend an existing
progression line without erasing earlier mechanics; inflation
controls are growth asymmetry, hard caps, softcaps (softcaps flagged controversial for
readability). **A true normalized/share-of-total production function is open design territory** —
with the warning that it can break per-purchase legibility and therefore needs an unusually
strong clarity justification.

## 6. Community & design literature

- **Pecorella corpus:** [Part I](https://www.gamedeveloper.com/design/the-math-of-idle-games-part-i)
  (cost
  `base×rate^owned`; AdCap lemonade base 4 rate 1.07; bulk-buy and max-affordable closed forms;
  newest-generator dominance; staggered milestones as the fix), Part II (chain math),
  [Part III](https://www.gamedeveloper.com/design/the-math-of-idle-games-part-iii) (prestige
  formulas), and [Quest for Progress](https://www.gdcvault.com/play/1023876/Quest-for-Progress-The-Math)
  (GDC Europe 2016).
- **Paper Pilot:** [A Guide to Incrementals](https://paperpilot.dev/garden/guide-to-incrementals/defining-the-genre)
  treats engagement as the test for content, warns that repeatables can decay into chores before
  automation is sold back, and contrasts replacement of old mechanics with building on a core
  mechanic.
- **Archived Reddit discussions:** anecdotal complaints described buildings becoming useless and
  asked how to preserve tier choices while keeping effects legible. The pullpush retrieval did not
  retain durable thread identifiers, so these anecdotes are not evidence for a product claim and
  must be re-sourced before citation.
- **NOT verified:** any Orteil or Hevipelle interview quote — do not cite without re-verification.

## 7. Novel/obscure answers

- **Progress Knight** (verified): nothing is a generator; jobs/skills level with use
  (`maxXp = baseXP×(level+1)×1.01^level`, income `1+log₁₀(level+1)`); rebirth replays ALL content;
  essence milestones convert grind into automation. Relevance model: **content is re-run, not
  outgrown.** Idle Loops / Increlution same family (prior knowledge, flagged).
- **CIFI**: chain + flat construction milestones (e.g. #14 Cells ×1e44) as a live game's pacing
  spine. **Magic Research / FE000000**: unverified, flagged. **Melvor mastery** (§4): strongest
  obscure mechanism.
- **Universal Paperclips** — anti-example: hard stage replacement, celebrated anyway because the
  arc is ~hours. **Lesson: tier-relevance engineering matters in proportion to intended play
  length.**

## Cross-cutting synthesis

Five orthogonal, stackable families: **(1) kinetic chains** (counts in a product; relevance
guaranteed; decisions still collapse to top tier without a purchased/generated split — AD's split
+ per-10 milestones is the best-shipped fix); **(2) count-as-multiplier coupling** (pairwise ≪
pooled ≪ count-as-base, in increasing power; retrofit-friendly); **(3) count milestones /
achievement passives** (cheap, shipped everywhere; makes buys eventually worthwhile, not
interesting); **(4) role differentiation** (Kittens/Evolve capacity-gating is the ONLY family
where early buildings stay mandatory through endgame for non-numeric reasons; sacrifice and
minigame-input roles extend it); **(5) build re-authoring** (RG factions, Gnorp talents —
strongest community verdicts in the survey). Normalized production is unshipped, open territory.
