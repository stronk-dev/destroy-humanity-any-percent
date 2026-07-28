# Playstyles

> Playstyle identity comes from **rule variance, not stat variance**. The five working rule-change categories (from `research/events-playstyles.md §2c`): delete a system · invert a valence · rewire the production graph · change the constraining resource · change what action generates value. Every faction and challenge below is built from those moves, and each build has **an anchor that needs support and something actively bad under it** — the negative space defines the identity.

## 1. Factions

Chosen per run at Tier 2 (the incorporation moment); switching is *content*, not commitment (Realm Grinder's abdicate model — meta-progression persists across factions). **The balance law: never balance factions against each other on one output axis; balance each to be optimal for a distinct real-world engagement pattern.** Then there is no global optimum, only a local one per player.

| Faction | Verb (what you do) | Constraint | Actively bad | Real-world fit | Produces for others (cannot self-consume) |
|---|---|---|---|---|---|
| **Bootstrapper** | Reinvest revenue; all growth self-funded; tight buy-order optimization | Cash-on-hand; idle capital decays via `Opportunity Cost` | Hoarding; hypergrowth; debt | Steady grinders | **Revenue** → VC (feeds their milestones) |
| **VC-Funded** | Raise rounds: burst capital injections costing permanent **Equity** (a % tax on all future output) + a **Board Pressure** meter | **Runway** — a literal countdown; milestones or death; runway burns while you're away | Slow steady play; idling | Burst-active players, timed pushes, risk | **Hype** → Open Source (infrastructure grants) |
| **Open Source** | Contribute output to the server-wide **commons pool** you also draw from; strong when the community is strong | Community goodwill / contributor count (a social resource) | Playing alone; monetizing; hoarding private upgrades | Social/guild players; the natural Ethical% faction | **Libraries** → Enterprise (contract inputs) |
| **Enterprise** | Sign **contracts**: N-hour deliverables that fulfil while you're away; SLA-bound | Real elapsed time; SLA compliance | Fiddling — frequent changes break SLAs and void contracts | Pure idle, once-a-day players | **Compliance** → Bootstrapper (unlocks customer tiers) |

The fourth column's cycle is the **Last Meadow rule** (you cannot produce what your own class consumes) — the engine of guild interdependence (`05-mmo.md §3`).

**Later factions:** **Crypto** (a per-tick resampled volatility multiplier; the RNG itself is upgradeable; extreme variance as identity) · **Regulated Utility** (low ceiling, immune to disasters/pressure meters — the sanctuary build) · **Consultancy** (the Mercenary: pick 12 upgrades + 1 tool from any faction; unlocked at high Reputation — build-your-own as endgame).

**Faction-inverted upgrades** (the cheapest identity tool): the same tree slot reads `+50% output per unspent dollar` for Bootstrappers and `+50% output per dollar spent this hour` for VC. One inverted heuristic per tier minimum.

## 2. The four engagement builds

Orthogonal to faction (though correlated by design):

| Build | Loop | Key mechanics | Leaderboard |
|---|---|---|---|
| **Active** | presence, timing, combos | golden-opportunity windows, bank management (the Lucky formula), Advisory Board choreography, minigame skill | combo records, speedrun categories |
| **Idle** | setup quality, then absence | daemons (the wrinkler economy), Enterprise contracts, offline cap upgrades, "Asceticism" advisors | days-idle efficiency boards |
| **Check-in** | many short sessions | Earnings Calls, swap-budget timing, garden ticks, pet care, dailies with streaks | streak/ritual boards |
| **Banker** | bank absence, spend it deliberately | Compute Credits (banked offline time → burst acceleration), loud HUD affordance, auto-spend fallback | biggest-burst boards |

**Separate leaderboards per build are mandatory** — else active wins all social comparison and the other identities are second-class regardless of math.

## 3. Ideology toggles (one toggle, two games)

The Grandmapocalypse pattern: opt-in state changes with a cheap reversible valve and an expensive permanent one, **both sides with real adherents** (a binary with a dominant option is just a delayed unlock — verified per patch in the balance harness).

- **The Enshittification slider** (per product, Tier 3+): stages raise revenue, drain Trust, spawn daemons (the idle economy's crown jewel lives here — idle builds *want* it maxed; active builds want it low because outrage events wreck combo windows). Valve: `Quiet Period` (temporary calm, cost doubles per use — the Elder Pledge); permanent off: `Take It Private` at a flat visible −5% output (the Covenant).
- **Go Public** (Tier 4+): quarterly-earnings pressure events + a shareholder meter + a much higher ceiling and shorter, riskier windows. Reverse: `Go Private Again` — expensive, flat tax, revocable.
- **The Golden Switch equivalent:** `Autopilot Mode` — golden opportunities stop spawning, +50% flat rate. The explicit active↔idle lever.

## 3b. Doctrines — the Age-Up choice (build-to-something)

Runs branch **inside** the ladder, Age-of-Mythology style: **every tier transition demands an irreversible 1-of-3 commitment** — the company's defining bet for that era. You don't pick perks; you *build toward an identity* that gates content (EU4 mission-tree logic):

| Transition | The question | Example doctrines |
|---|---|---|
| Garage → IT Co | What did you build? | Dev Tools · Consumer App · Government Contractor |
| IT → Cloud | Platform religion? | Open Ecosystem · Walled Garden · Defense Cloud |
| Cloud → Data Centers | Where do you build? | Frontier Regions (cheap, dirty) · Fortress Regions (compliant, slow) · Orbital/Offshore (absurd, late-payoff) |
| Data Centers → AI | Lab identity? | Product Lab · Frontier Lab · Open-Weights |
| AI → AGI | Alignment posture? | Ship It · Contain It · Merge With It |

Rules:
- Each doctrine opens a **doctrine tree**: unique upgrades, Layer-1 event chains, minigame modifiers, and one building type nothing else gets. Doctrines **combo** with faction and with each other (a Bootstrapper Dev-Tools → Open-Ecosystem → Open-Weights line is a coherent, distinct game from a VC Consumer → Walled-Garden → Product-Lab line).
- **Content gating is real:** some events, achievements, cosmetics, and at least one ending variant are only reachable via specific doctrine paths — the reason run #7 differs from run #2, and the engine behind "numerous prestiges to the first ending" feeling like *routing*, not repetition.
- Doctrines are **run-scoped** (reset on Exit); discovering a doctrine's content contributes Route Knowledge; named route combos surface on leaderboards ("Walled-Garden rush").
- Choice presentation is diegetic: a board meeting, three term sheets on the table, the timer running.
- Balance law extends: doctrines are balanced to be *interesting*, not equal — but every doctrine must anchor at least one viable route to the ending (verified in the balance harness).

## 4. Challenge runs (rule-modified runs as mainline content)

The AD Eternity Challenge + Trimps Challenge² model: each challenge is a **rule deletion or inversion**, gated in the Reputation tree (unlocking one is a progression reward), completable ~5× with escalating goals, paying **permanent** bonuses on **independent tracks** (deeper = more; repeating the same depth pays nothing — breadth is mandatory).

Launch set (one per rule-change category):

| Challenge | Rule change | Category |
|---|---|---|
| `No Hires` | headcount system deleted; automation only | delete a system |
| `Degrowth` | production *above a cap* is wasted; score = efficiency | invert a valence |
| `Vertical Integration` | generators produce the *next* generator, not cash (graph rewire) | rewire the graph |
| `Dial-Up` | game runs 10× slower; hard wall-clock limit | change the constraint (time) |
| `Founder Mode` | only clicks and policies generate value; generators produce 0 passively | change the verb |
| `Hardware Lottery` | random 50% of upgrades locked per run (roguelite variance) | run variance |
| `Legacy Systems` | Melvor-Adventure-style: no system's level may exceed your lowest system | constraint coupling |

Speedrun categories (`08-satire-flavor.md §speedrun`) double as challenge modifiers: `Glitchless` (no regulatory-arbitrage skips), `Low%`, `Legal%` ("nobody runs this category"), `Net Zero%`, `Pacifist/No Externality`.

## 5. Moral routes (the Undertale structure)

- **Canon route** — enshittification → oligarch: the path of least resistance; the game visibly tilts you toward it (defaults, tooltips, the board's suggestions). Full multiplier access. Ends in the oligarch arc and Endings A/B.
- **Ethical%** — the hard, true route: forgo dark-pattern stages, externality dumping, and VC money ("no XP"). A different grammar: **Trust becomes the production stat** (a Trust-scaled stack replacing the forgone multipliers), mutual aid via the **commons** (`05-mmo.md §5` — the prisoner's dilemma that makes it viable), Open Source synergies, and slow compounding. Brutally slow solo; genuinely viable with cooperation; **technically completable by cheese** — and cheese is celebrated (`Glitched` is a run **variable** on Ethical% per `05 §6`'s category model; the joke is the point).
- **The game remembers — as a ledger, not a debuff** (decided 2026-07-28 per `research/morality-systems.md §7.6`): **dated facts persist** across runs — what you did, which NPCs remember it, what the pet saw; NPCs, story cycles and the pet reference past companies, and there is no memory-wipe purchase. **Scores reset each run**, reseeded as `clamp(90 − 0.35·Notoriety, 55, 90)` — **Notoriety** = the count-weighted sum of harm-category entries in the founder's permanent ledger across all runs (the Externality ledger and dark-pattern stage-3 events weigh heaviest); **formula published in-game, per law 9** — so a canon-route veteran can genuinely turn and attempt Ethical% (a persistent score made that arithmetically unreachable by ~run 15), and **nobody ever starts unplayable**. Memory becomes narrative; redemption stays possible; the record stays permanent.
- **World-first broadcasting:** first Ethical% completion (per season, and all-time) is a permanent dated broadcast event.

## 6. Policies (the AGI-tier meta-playstyle)

At Tier 6, playstyle itself becomes authorable: **policies** are saved, named, shareable scripts (Idle Loops / Stuck in Time lineage) that automate lower-tier verbs — including minigames (writing a bot for your own hobbies triggers the Soul question by design). Policy libraries are community content: shareable, forkable, with a policy-exchange board. The active/idle/check-in/banker distinction ascends one level: now it describes how you *manage policies*.
