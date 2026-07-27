# Economy & Balancing

> The currency/layer architecture, the core math, and the pacing targets. Numbers here are **starting values for the balance spreadsheet**, not gospel — balance data lives in hot-reloadable data files (`06-tech.md`), and everything below will be retuned against telemetry.

## 1. The layer architecture

Progression layers are scoped by **what persists**:

| # | Layer | Scope | Persists? | Currencies / state |
|---|---|---|---|---|
| 1 | **Company** | the current run | resets on Exit (prestige) | Cash, Users, Compute/Tokens; constraints: Energy, Water, Permits; hidden: Trust, Externality |
| 2 | **Founder** | the player, forever | yes | Reputation, Network, Route Knowledge, Clout, Personal Wealth, **Soul**, founder age |
| 3 | **World** | the server, everyone | never resets (ratchet) | Planet %, community milestone tiers, seasonal war state, Influence (personal payout from world events) |
| 4 | **Guild** | the co-op | while the guild lives | Guild level (fed by automatic tithe), guild upgrades, commons pool |

Cross-cutting systems on their own clocks:

- **Fiscal Quarters / Earnings Calls** — the sugar-lump equivalent (real-time, production-immune; §5).
- **Clout** — the achievement/influence axis (milk equivalent + attention-economy gameplay; §6).
- **The moral axis** — Trust / Externality / p(doom): deliberately **not spendable** (§7).
- **Soul** — the founder's personal ledger (§8).
- **Banked time (Compute Credits)** — the idle build's spendable resource (§9).

**Presentation discipline:** each layer has a different decision grammar and clock; layers are introduced one at a time across the tier arc; the number of simultaneously visible currencies per tier stays small. The late-game conversion screen is *deliberately* confusing — as parody — and carries the giant **"Just Give Me The Thing"** button that does the right thing.

## 2. Company layer — core math

### 2.1 Generator cost curve

```
price(n) = base × r^n
```

- `r` is configured per generator class, never globally in engine code. **1.13 is the current
  provisional balance baseline** (research band 1.07–1.15; CC uses 1.15, AdCap 1.07), not a
  launch commitment; the balance harness selects shipped values per class.
- At 1.13, price doubles every ~5.7 purchases. This baseline is slightly gentler than Cookie
  Clicker because our multiplier stack is thinner early (we stagger systems across tiers).
- Generator ladder steps ~×12 cost / ~×6.5 output between generator types (CC's proven cadence), with deliberate cadence breaks at tier boundaries (the last generator of each tier is a gated splurge).
- **Milestone multipliers at 25 / 50 / 100 / 150 / 200… owned** (Math of Idle Games) so old generators never die; Cursor-style exception generators (one per tier scales per-other-building, a different curve in the same slot).
- Implement geometric-series bulk-buy and max-affordable inverse in the Decimal type from day one:
  `bulk(k,n) = base·r^k·(r^n−1)/(r−1)`; `maxAffordable = ⌊log_r(cash·(r−1)/(base·r^k) + 1)⌋`.

### 2.2 The production stack

Documented, ordered, and capped in depth (Paper Pilot: limit nested multiplier chains):

```
Output = Σ_generators [ base × count × tierUpgrades(×2 each) × synergies × buildingLevel(+1%/lvl) ]
       × CloutStack (multiplicative kittens-equivalent)
       × FounderBonus (+1% per Reputation level, unlock-gated 5→25→50→75→100%)
       × FactionModifiers (rule-changing, see 10-playstyles.md)
       × BuffWindows (golden-opportunity events, multiplicative, short)
       × QuarterBonus (+1% per unspent Investor Confidence, cap 100)
       × WorldModifiers (seasonal situations, guild upgrades)
       × moral-route modifiers (Ethical% forgoes several of the above; Trust-based alternative stack)
```

Hard rule: **hardcaps, never softcaps.** Any cap is a visible number with a tooltip explaining it.

### 2.3 Active play: buff windows and the bank formula

- **Golden opportunities** (our golden cookies): spawn on Cookie Clicker's shaped distribution (t⁵·exp — suppresses instant re-spawns and long tails), base 300–900 s, upgradeable toward ~120 s. Effects mirror the proven set: Frenzy-type (×7 production, ~77 s), Click-Frenzy-type (×777 per click, ~13 s), Building-Special-type (+10%/owned of one generator, 30 s), instant payouts.
- **The Lucky formula, kept verbatim:** `payout = min(0.15 × bank, 900 × rate) + ε` → creates the 6,000×rate bank-management target and hoard-vs-spend tension.
- **Buffs multiply.** Combo ceiling target: ~10⁵–10⁶× rate for ~15 s windows with full setup (loans + spell-equivalent + faction tools). Active play is a separate skill discipline, not a nerf target.
- **Clicking is designed, not left to autoclickers:** click batches are rate-validated server-side (~20/s clamp, silent); above the clamp, *timing* not rate is the skill (combo windows, buff sequencing). An in-game "macro bench" (unlocked at AGI tier as policies) makes automation an explicit, sanctioned progression instead of an external cheat.
- **Wrinkler equivalent:** daemon processes (Tier 3+) attach to tenants, each draining 5% of visible rate; popping returns ×1.1 of drained total — quadratic patience bonus, net ×6 at 10 daemons. Idle build's crown jewel, gated behind the enshittification slider (the "make it worse for a multiplier" opt-in).

### 2.4 Offline / idle

- **Offline progress is default** at 90% efficiency, cap 24 h (Melvor's goodwill number), extendable by upgrades and the Enterprise faction (contracts fulfil fully offline).
- Server computes offline gains closed-form on reconnect; client renders the welcome-back modal from the server result (`06-tech.md §idle-math`).

## 3. Founder layer (prestige)

### 3.1 The prestige formula

```
ReputationLevel = ⌊ (lifetimeValue / T)^(1/3) ⌋      // cube root, lifetime-based
```

- **Cube root** (CC's choice): ~8× lifetime value to double Reputation — generous early (first Exits minutes-to-hours apart, dopamine front-loaded), naturally decelerating.
- **Lifetime-based** (not per-run): forces forward progress, prevents same-point farming.
- Each Reputation level = +1% production, additive — but the benefit is **unlock-gated** (5% → 25% → 50% → 75% → 100% of your Reputation bonus, bought with Reputation): the first several Exits are about buying the right to benefit from Exiting.
- **First-Exit pacing target: the first Exit pays off visibly within the same session** (the Unnamed Space Idle churn lesson). Aim: first Exit available ~45–90 min in; second run reaches the same point in <15 min.

### 3.2 Founder sub-currencies

- **Reputation** — heavenly-chips equivalent; spent in a permanent tree (offline extension, starter packages, synergy unlocks, golden-opportunity upgrades, permanent-slot purchases).
- **Network** — permanent slots with fiction: "your old CTO joins the next venture" — designate an owned upgrade/person to carry into every future run. Slots I–V bought with Reputation.
- **Route Knowledge** — the speedrun meta: discovered skips (`Regulatory Capture Skip`, `Nonprofit Wrapper Zip`, `IPO Sequence Break`…) persist as route options; categories completed unlock modifiers.
- **Clout** — see §6. Persists.
- **Personal Wealth** — a *separate number from company cash*, extracted from the company via late-tier actions (dividends, buybacks, Strip-Mining). Grows even when the company dies — "the founder always lands softly." Buys the billionaire-layer content (yachts, bunkers, longevity rungs, philanthropy parody). Deliberately cannot be reinvested into company production at fair value — the satire is that it's *extractive*.
- **Founder age** — advances in run-time and with certain choices; crunch ages you. The endgame barrier (`01-tiers.md Tier 6`): longevity rungs push the actuarial wall back and **cost Soul**.

### 3.3 Exit fiction

Exit types are tier-flavored (acquihire / acquisition / IPO / collapse) with slightly different payouts: e.g. collapse pays less Reputation but more Route Knowledge ("every death teaches you the map" — Increlution). Strip-Mining (convert a fully-built branch into instant currency; branch unbuildable this run) is the LBO-flavored pre-Exit squeeze.

## 4. World & guild layers

Summarized here; full design in `05-mmo.md`:

- **Planet %** — the shared ratchet; drained by everyone's late-tier consumption; never resets. The server's long-term story.
- **Community milestone tiers** — dual-axis payouts (personal Influence by contribution rank + global unlock).
- **Guild tithe** — a fixed small % of every member's output automatically feeds the guild level (Idle Clans model); leaders spend guild levels on member-wide upgrades. Contribution is percentile-normalized per faction.
- **Influence** — the personal payout currency from world events; spent on world-layer cosmetics, GM-shop items, and guild seeding.

## 5. Fiscal Quarters (the CpS-immune clock)

The sugar-lump equivalent, diegetic:

- An **Earnings Call** ripens every 24 real hours (harvest window mechanics mirror sugar lumps: harvestable early at 20 h with a 50% "missed estimates" fail chance; guaranteed at 23 h; auto-reports at 24 h).
- Yields **Investor Confidence**: spent on building levels (+1% that generator per level, level N costs N), minigame unlocks (staggered across tiers), and special quarter types (Golden Quarter doubles banked cash capped at 24 h of rate; Caramelized-equivalent refills pantheon-style swap cooldowns).
- **Sugar Baking equivalent:** +1% production per unspent Confidence, cap 100 — hoard-vs-spend at the meta level.
- **Immune to production rate by design.** No amount of output accelerates the clock. This is the long-tail pacing device and it is never for sale.

## 6. Clout (achievements + influence)

Two feeds, one axis:

1. **Achievements** (the milk model): every achievement grants +4 Clout. Achievements are deliberately weird and load-bearing (own 300 of a generator, win a chess minigame against the hardest bot, keep the pet at max trust for a week, trigger `AGI In Two Years` five times). Target ~600 at launch across all systems.
2. **Attention-economy play**: posting (a lightweight timed action), viral events, the podcast circuit, thought-leadership upgrades. Fed by and feeding the events system.

Conversion: **PR Interns** (kitten equivalents) each multiply production by `(1 + Clout × factor)`, stacking multiplicatively — the single biggest multiplier family in the game, making optional weirdness load-bearing (CC's highest-leverage idea).

Downstream: Clout also gates **regulatory outcomes** (lobbying events check Clout), **MMO visibility** (feed prominence, leaderboard flair), and **canonization events** (the South Park episode fires at Clout thresholds — see `08-satire-flavor.md`). Clout persists with the founder: the personal brand survives the company. (Satire: it survives *better than the employees do* — a lore card says so.)

## 7. The moral axis (not spendable)

- **Trust** (0–100, starts high): drained by enshittification stages, dark-pattern parody purchases, externality events; raised slowly by genuinely good choices (which cost growth). **Cannot be bought.** Gates: event branches, the Ethical% route's production stack (Trust becomes the multiplier), certain endings.

  > `DESIGN-GAP:` **four contradictions from `design/research/morality-systems.md §7.6`, owner decision required.**
  > 1. **Trust is currently shaped like a Paragon bar** — one signed 0–100 that gates. Mixed play fails every check. Proposed: **five constituencies × two independent bars** (Standing/Grievance), the New Vegas factional architecture rather than the Fallout 3 global-karma one. *We already have the factions.*
  > 2. **"Gates" is the optimisation pathology.** The "20 more evil points" failure comes from *published thresholds*, not from visible values — and Dishonored proves hiding the number doesn't fix it, it just moves the optimisation to a wiki (importing two of our own anti-goals). Proposed rule: **meters modulate, facts gate.** Publish everything, have no thresholds, key endings off dated ledger facts.
  > 3. **`§1` lists Trust as hidden** while this line makes it load-bearing. Pick one.
  > 4. **Four moral quantities is too many to stay legible.** Proposed: cut **p(doom)** from the moral axis (it is already a pressure meter in `09 §3`) and demote **Externality** to a ledger → **two meters + one ledger**. **Soul survives the legibility test** (self-directed, distinct named zero state, both off-diagonal corners already written into `§8`) — keep it, and enforce decorrelation with a CI correlation gate or content authors will silently merge it back into Trust.
  >
  > Cross-cutting principle: **derive morality from the production stack, never award it for acts** — the difference between farmable Fallout 3 karma and the commons dossier's Enclosure index.
- **Externality**: accumulates from constraint-dodging (unpermitted turbines, e-waste export, offsets). Leaves your ledger, lands on the world map; revealed in late-tier audit events; feeds the Planet drain and conspiracy/outrage pressure meters.
- **p(doom)** (Tier 5+): raised by e/acc tree, capability rushing; lowered by safety spending (which slows you). Gates AGI-tier event chains and one ending variant.
- **Bidirectional routes (Undertale model):** the canon path is enshittification → oligarch — the game tilts you toward it. **Ethical%** forgoes the dark multipliers entirely (no dark patterns, no externality dumping, no VC — "no XP") in exchange for a Trust-based slow-compounding grammar, viable mainly via the MMO commons (prisoner's-dilemma mechanics, `05-mmo.md §commons`) or celebrated cheese. The game remembers your morality across runs (founder-scoped; events reference past companies).

## 8. Soul (the personal ledger)

Distinct from morals: Trust is what you do to others; Soul is what's left of *you*.

- **Drains:** Faustian contracts (VC term sheet clause 7(c): "…and one (1) soul"), crunch/founder-mode events, longevity rungs, skipping the recital.
- **Gates the human content:** as Soul drains, the pet stops recognizing you and its UI literally greys out (the cattery trust system is the soul proxy); hobby minigames lock behind *"you no longer remember why this was fun."*
- **Recovery:** touch-grass activities that cost time and produce nothing (a genuine idle-build synergy; an active-play tax).
- **Endgame:** the Soul balance answers "what exactly ascends?" at Transcendence. Zero Soul → training-data ending.
- Both corners are reachable and both are satire: the ethical workaholic (high Trust, no Soul) and the well-rested oligarch (low Trust, full Soul, does yoga).

## 9. Banked time (Compute Credits)

- Offline hours accrue as **Compute Credits** ("unused reserved capacity") up to a banked cap; spend them as chosen-moment acceleration bursts (e.g. 1 credit-hour = 60× speed for 1 min, tunable).
- Makes absence a *decision* and creates the banker playstyle (`10-playstyles.md`).
- **Legibility rules** (the Idle Spiral lesson): a loud, unmissable banked-time affordance in the primary HUD, and an optional auto-spend toggle so casual players aren't punished for not engaging.

## 10. Pacing targets (the spreadsheet's acceptance tests)

| Milestone | Target |
|---|---|
| First generator | < 60 seconds |
| Tier 1 | ~15 minutes |
| First minigame | ~2–3 hours (Tier 2) |
| First Exit available | 45–90 minutes; obviously worth it same-session |
| Tier 3 (cloud + easter egg + daemons) | day 2–4 |
| First Earnings Call ritual established | day 2 |
| Tier 4 (community-gated at launch) | week 1–2 |
| Tier 5–6 | weeks 3–6, interleaved with seasonal arcs |
| First ending reachable | ~2 months of mixed play (faster for optimizers — that's what the leaderboard is for) |
| Long tail | Quarters, seasonal arcs, categories/challenge runs, Ethical%, world-firsts — clock- and community-driven, not grind-driven |

Verification approach: a balance simulation harness (headless run of the production stack against scripted strategies — pure-idle, check-in, optimizer) must reproduce these targets within tolerance before launch and after every balance patch. Community-milestone thresholds are set from live telemetry with uncertainty margins (the Clash of Clans 48-hour lesson), never guessed.

## 11. Big numbers

`break_infinity.js` 2.2.0 client / Go `Decimal` port server, golden-vector tested; wire format strings; display via notations library. Full detail in `06-tech.md §3`. NaN detection refuses to persist poisoned saves (Profectus rule).
