# Board-Game Mechanics — Minigame Inspiration Dossier

> **Comparative research, not a roster or implementation authority.** Every ranking, bot/server-
> cost label and product-fit verdict below is a prototype hypothesis, not a measurement. Names,
> copy, economy hooks, roster additions and “do not build” calls are non-adopted proposals routed
> to current owner/design/RFC authority. Unfetched exemplar specifics have been removed from the
> body; `[M]` family-level context remains non-authoritative. The legal note is issue-spotting, not
> copyright, trademark or patent clearance. Statements about what “ships” describe the dated
> design coordinate, not HEAD.
>
> **Feeds:** `design/03-minigames.md` (the minigame roster + clock taxonomy, esp. `§5`/`§5b` the parlor suite) · `design/08-satire-flavor.md` (voice) · any future board-game-suite or minigame-platform RFC.
>
> **Builds on, does not redo:** `research/lane-pusher-design.md` — **read its §4 (async + AI fallback) and §5 (MVP/determinism) before this file.** That dossier owns the async-snapshot PvP pattern, the behaviour-flag bot-difficulty ladder, and the determinism-as-verification law; everything here inherits it. · `research/roguelike-survivor-minigames.md` — **owns the deckbuilder-roguelike / auto-battler / drafting-a-run family (Slay the Spire, Balatro, Super Auto Pets). This file deliberately does NOT re-cover the deckbuilder angle** — where a board-game drafting family (7 Wonders) touches it, I cross-reference rather than re-derive. · `research/creature-battler.md §3.4` (auto-resolved combat converts small edges into near-certainties — the law that governs any bot-vs-bot mode). · `design/03-minigames.md §5b` already catalogued the *parlor taxonomy* (trick-taking, melding, racing, dice/bluff, mancala, etc.); **this dossier is the mechanic-family layer beneath that taxonomy** — it explains why each family is fun and how hard it is for us to certify and bot, and it does not re-list §5b's placement table.
>
> **Scope boundary (owner's brief):** chess/abstract-strategy and deckbuilder-roguelikes are covered elsewhere and are **out of scope here** except as reference points. This file covers the rest of the tabletop design space: tile-laying, worker placement, engine-building/tableau, drafting, area-control/majority, auction/bidding, route/network building, set collection, push-your-luck, and social deduction.
>
> **Provenance glyphs** (README house conventions): `[V]` verified against a fetched URL this pass · `[P]` plausible/secondary · `[M]` model knowledge, unverified this pass. **No `[M]` constant should be printed as fact in shipped copy without re-checking §Verify first.**
>
> **Tooling note:** WebSearch budget was exhausted before this pass; verification is via direct Wikipedia fetches (which succeeded for the seven flagship exemplars) plus strong model knowledge of extremely well-documented games. Every fetched claim carries `[V]`; every unfetched specific carries `[M]` and appears on the §Verify list.

---

## 0. The one-paragraph answer

Turn-based, discrete-input games are promising candidates for deterministic server replay, but
their runtime cost and bot quality are not established by this desk study. Three prototype
hypotheses emerge: a sandboxed engine-builder may reuse the shape of the idle economy; simultaneous
drafting may reduce multiplayer wait chains; and free-form social deduction conflicts with the
current no-free-text and AI-fallback constraints. The remaining families are comparative options,
not an approved roster. Their qualitative ordering in §12 must be tested against measured server
cost, bot behavior, accessibility, economy isolation and differentiation from surfaces that
actually exist at HEAD.

---

## 1. The axis that decides tractability (and why board games win it)

Design law 2: clients send **intents**, never results; the server must **certify/replay** any outcome. For board games this is almost free, and it is worth stating precisely why, because it is the reason the owner can treat this whole space as "easy enough":

| Property | Board-game families (this dossier) | Why it matters to us |
|---|---|---|
| **Input alphabet** | small enums — *place tile at (x,y,rot)*, *put worker on action space k*, *pick card i and pass*, *bid n*, *claim route r*, *roll-again / stop* | Certification = replay the choice log against a seeded RNG. No floats, no continuous sim, no per-player tick loop (law 2 satisfied by construction). |
| **Physics** | turn-based / simultaneous-commit; no real-time control | Match cost is microseconds in Go; a `Match` goroutine per board is affordable at any plausible CCU (`06-tech.md`). |
| **RNG surface** | tile bag / deck shuffle / dice — all seedable | Same seeded, re-derivable, **daily-seed and duplicate-scoring** property `03 §5b` already wants. Determinism is a *feature request already granted*, not a cost. |
| **Bot** | a heuristic evaluator over a discrete choice set | The board-game bots are the tractable end even of `03 §5`'s existing vs-AI system (alpha-beta / eval tables), not the pathfinding-heavy end the lane dossier warned about. |

**The one structural caveat, inherited from `creature-battler.md §3.4` and re-measured by the lane dossier:** if a match *auto-resolves with no live input*, small persistent edges compound to near-certainty and the game collapses to a dominant strategy. Board games mostly dodge this because **the human plays every turn** — but it flags any variant where we'd let bots play bots for stakes, and it flags **chance-heavy families** (dice, some card games) whose raw-outcome variance drowns skill at our match counts. `03 §5b` already has the fix on record: **duplicate scoring** (same seeded shuffle to every seat, score on relative performance) for any chance game entering a real ladder. That rule does a lot of work below and I lean on it rather than re-deriving it.

---

## 2. Tile-laying — Carcassonne, Azul, Patchwork

**The verified exemplars.** **Carcassonne** (Klaus-Jürgen Wrede, Hans im Glück, **2000**): draw a terrain tile, place it adjacent so roads/cities/fields match across edges, optionally station a **meeple** to claim the feature; completed features score by tile count (cities 2/tile, roads 1/tile), fields score at game end; **72 base tiles** `[V]`. **Azul** (Michael Kiesling, Plan B, **2017**): draft **all tiles of one colour** from a factory display or the centre, fill pattern rows, then move one tile per filled row onto your 5×5 wall for adjacency-scored points and completion bonuses `[V]`. The exact overflow penalty was not verified and is not used as a factual premise here.

**Why it's fun.** Two loops: a **spatial-optimisation puzzle** (where does this piece go to score most / waste least) and a **drafting-under-scarcity** loop (Azul's genius is that taking what *you* want often hands your opponent exactly what *they* want, and over-taking is punished — every draft is a small prisoner's dilemma). The dopamine is a board filling up *neatly*; the skill is denial and efficiency.

**Tractability — server-cheap, bot-easy.** Input alphabet is *(tile id, position, rotation)* or *(colour, source, destination row)* — small enums, trivially replayed. RNG is the tile bag/factory refill, seeded. Match cost negligible. **Bot difficulty: easy.** A greedy "maximise immediate placement score, break ties by denying the opponent's best follow-up" heuristic is one screen; a 1-ply lookahead is strong in Azul because the branching factor is tiny (≤~30 draft options). Difficulty scales cleanly by lookahead depth + an ε-greedy blunder rate — the same knob `03 §5` uses for Connect-4/Othello. **This family is squarely in the tractable band.**

**AI-fallback difficulty: LOW.** Solo-playable against the heuristic with no quality loss; the bot is characterful (a "hoarder" bot that over-takes and eats the penalty reads as *greedy*, not *broken* — the same personality-via-policy trick the lane dossier used).

**Non-adopted tech-satire mapping:** a tile-laying prototype could explore datacenter layout or
capacity planning. Exact names, copy and an overflow-billing rule require design authority and a
verified mechanic source.

---

## 3. Worker placement — Agricola, Lords of Waterdeep, Everdell

**The verified exemplar.** **Agricola** (Uwe Rosenberg, Lookout, **2007**): players place family members on exclusive action spaces, creating denial pressure; harvests require feeding the family or losing victory points `[V]`. Other worker-placement hybrids were not fetched and are omitted as factual exemplars.

**Why it's fun.** **Action-selection under scarcity is a pure tension engine.** Every turn is "which of the things I need can I still get, and what am I forcing my opponent to give up?" The blocking is the game — placement is simultaneous in effect even when sequential in turn, because the *board state after each placement* is the shared resource. Agricola's feeding clock layers a **push-your-economy pressure** on top: expand too fast and you starve.

**Tractability — server-cheap, bot MODERATE.** Input alphabet is *(worker, action space, optional sub-choice)* — clean enums, replay-trivial. But the **bot is the hardest of the "easy" families here**: worker placement rewards *lookahead and denial*, so a naive greedy bot (take my best action now) is exploitable and reads as dim. A competent bot wants a shallow search with an opponent model ("what will they be denied if I take this?") — still tractable (the branching factor per turn is the number of open spaces, small), but it is real eval-function work, closer to `03 §5`'s alpha-beta games than to a one-line heuristic. Difficulty ladder = search depth + how many behaviours (denial, feeding-safety, engine-priority) are enabled — the **behaviour-flag ladder** the lane dossier established transfers verbatim, and it produces characterful weak bots (a bot with `PLANS_FEEDING` off literally starves — a legible mistake).

**AI-fallback difficulty: LOW–MEDIUM.** Fully solo-playable; the only cost is that a *good* bot needs a real eval function, not a heuristic. This is the family where the **economy-hook fit is strongest of the classic Euros**, which offsets the bot cost (see §12b synergy note below).

**Non-adopted tech-satire mapping:** workers could represent engineers competing for sprint actions,
with a payroll/burn pressure inspired by the verified Agricola feeding cycle. Exact names, copy,
resources and scoring require design authority.

---

## 4. Engine-building / tableau — Wingspan, Terraforming Mars, Splendor

> **This is the family that overlaps our economy core — the brief asked for the synergy called out explicitly, so here it is up front.**

**The verified exemplars.** **Splendor** (Marc André, Space Cowboys, **2014**): collect gem chips, buy development cards that permanently discount future cards, and chain discounts into an engine; first to 15 prestige points wins `[V]`. **Wingspan** (Elizabeth Hargrave, Stonemaier, **2019**): play bird cards into three habitats whose actions improve as cards are added, with some chained abilities `[V]`. Unfetched Terraforming Mars specifics are omitted.

**Why it's fun — and why it IS our game.** The loop is **"spend now to make everything cheaper/faster later,"** i.e. investment compounding into acceleration. That is the idle genre's central pleasure rendered as a bounded card game with a finish line. Splendor's discount-chaining is the cleanest instance: your fifth card costs almost nothing because your first four discount it. **This is structurally identical to Cloud Clicker's production formula** (`base_rate × Π(buff_i)`, generators that lower the effective cost of the next tier) — the same observation the roguelike dossier made about Balatro's `chips × mult`, arriving here from the board-game side. **Engine-building is the board-game family that maps onto our numeric core, full stop.**

**The synergy, stated as both an opportunity and a hazard:**
- **Opportunity:** the fit is *instant* — the tutorial for an engine-builder minigame is the tutorial for the whole game, and the big-number `Decimal` core and buff-stack resolver are already built. It's the same "expose the engine we shipped as a card game" argument the roguelike dossier makes for Balatro, and it applies to Splendor/Wingspan with equal force.
- **Hazard — `03 §12b` Fairness Law:** because the math is the *same* math, the temptation to let the minigame **read live company state** (your real production buffs seeding your engine) is exactly the pay-to-win loader error the Fairness Law forbids for ranked play. **An engine-builder minigame must run in its own seeded, sandboxed economy** — tier/building counts may enter only as `breadth` (which cards are in your pool) via the C5 `scaling_inputs` seam, **never as power** inside a rated board. Get this boundary right and it's the best-fitting minigame in the space; get it wrong and it's the thing that breaks the game's whole no-pay-to-win promise.

**Tractability — server-cheap, bot-easy.** Input alphabet: *(take chips / buy card i / reserve card)* or *(play card, assign resource)* — small enums, replay-trivial, seeded deck. **Bot: easy-to-moderate.** A greedy "buy the highest points-per-turn card I can afford, else accumulate toward the best reachable card" heuristic plays Splendor respectably; Wingspan/TM reward light lookahead over your own engine (no opponent modelling needed — engine-builders are often **multiplayer solitaire**, which makes the bot *easier* because it barely needs to react to opponents). Difficulty = greediness horizon + card-valuation noise.

**AI-fallback difficulty: LOW.** The multiplayer-solitaire property means a bot that just plays its own engine well is a *fully adequate* opponent — the AI-fallback law is nearly free here, like the PvE roguelikes.

**Non-adopted tech-satire mapping:** a sandboxed compute-stack engine could test whether permanent
discounts and chained actions teach the core economy. Names, resources, formulas and copy require
design authority.

---

## 5. Drafting — 7 Wonders, Sushi Go

**The verified exemplar.** **7 Wonders** (Antoine Bauza, Repos, **2010**): 2–7 players play three ages / 18 turns; each turn all players simultaneously choose a card and pass the remaining hand, building a tableau `[V]`. Unfetched Sushi Go specifics are omitted.

**Why it's fun — and why it's the best multiplayer fit in this whole dossier.** **Simultaneous pick-and-pass means there is no turn order** — every player commits at once, then hands rotate. That single property is enormous for us: an N-player round is *one committed choice per player per pass*, with **zero turn-order latency**, which is the friendliest possible shape for **async multiplayer** (you don't wait for a chain of opponents; everyone's move for pass *k* resolves together). The skill is **reading the wheel** — the hand you pass becomes an opponent's options, and the hand coming back to you is depleted by everyone between — so it's social and tense without any real-time demand.

**Tractability — server-cheap, bot-easy, ASYNC-native.** Input alphabet: *(pick card i from current hand)* — the smallest alphabet in the dossier. Certification = replay picks against the seeded initial deal and deterministic hand-rotation. **Bot: easy.** A card-evaluator ("value each card for my tableau + a hate-draft term for denying the neighbour's obvious build") is one screen and plays well; difficulty = weight on the hate-draft/lookahead term + a valuation-noise blunder rate. **Simultaneous commit also means the bot backfill is invisible** — a bot filling an empty seat commits at the same instant as everyone else, so `03 §5`'s "bot backfill after 10–30 s, disclosed, never hot-swapped" pattern is trivially satisfied.

**AI-fallback difficulty: LOW.** N-player games are fully playable with any mix of humans and bots because everyone acts simultaneously — this is the cleanest AI-fallback story of any *multiplayer* family here (engine-builders are easier but are effectively solitaire; drafting is easy *and* genuinely multiplayer).

**Non-adopted tech-satire mapping:** a simultaneous hiring or standards draft could test the
pick-and-pass interaction. Exact seating, tracks, names, content and copy require current
design/RFC authority.

---

## 6. Area control / majority — El Grande, Blood Rage

**The verified exemplar.** **El Grande** (Wolfgang Kramer & Richard Ulrich, Hans im Glück,
**1995**): players place caballeros into regions; majority positions score at three rounds, cards
set turn order/action power, and the Castillo supports a hidden reveal `[V]`. Unfetched Blood Rage
specifics are omitted.

**Why it's fun.** Majority scoring creates indirect conflict through commitment and denial.
El Grande's verified hidden Castillo reveal also supplies a menu-safe commitment layer.

**Tractability — server-cheap, bot MODERATE.** Input alphabet: *(place k caballeros in region r)* / *(bid card n)* / *(commit combat card face-down)* — clean enums, replay-trivial. **Bot: moderate**, for the same reason as worker placement — majorities reward *evaluating the whole board and denying*, so a competent bot needs a board-value function and light opponent modelling ("can I flip this region cheaply, or should I concede it and lock another?"). Still tractable; the region count is small and scoring is a simple sort. Difficulty ladder = evaluation depth + willingness to contest vs. hoard (a legible personality axis: an *aggressive* bot over-commits and gets out-tempo'd; a *passive* bot never contests and loses on totals). The **face-down commitment** sub-games are pure hidden-info-at-reveal, which is fine — the bot commits under the same information a human has (law 6), no cheating.

**AI-fallback difficulty: LOW–MEDIUM.** Fully solo-playable; the bot cost is the board-eval function, same as worker placement.

**Non-adopted tech-satire mapping:** market-share regions and a concealed launch commitment are
possible metaphors. Exact names, regions, scoring and copy require design authority.

---

## 7. Auction / bidding — Modern Art, Power Grid

**The verified exemplar.** **Power Grid** (Friedemann Friese, Rio Grande, **2004** English 2nd
ed.): an open power-plant auction combines with a commodity market, network building and reverse
turn order that disadvantages the leader during resource buying `[V]`. Unfetched Modern Art
specifics are omitted.

**Why it's fun.** Auctions turn price discovery into player interaction: the central tension is
estimating private value while deciding whether to raise a rival's price. Power Grid also uses
turn-order/resource-market feedback to constrain a leader.

**Tractability — server-cheap, bot TRACTABLE-but-subtle.** Input alphabet: *(bid n / pass)* — tiny. Certification: replay the bid sequence; sealed bids are committed-then-revealed (hash-commit if we want strict anti-peek, but server-side secrecy suffices under law 2). **Bot: tractable with a real valuation model.** An auction bot needs (a) a private-value estimate per item and (b) a bid policy (bid up to value − margin; drop when price exceeds value). That's standard and cheap, and it makes a **naturally strong, honest opponent** — no rubber-banding, the bot simply values items and bids its value, which is exactly the "genuinely-rated bot" honesty `03 §5` mandates. Difficulty ladder = valuation accuracy + bluff/bid-up aggression flags (a weak bot mis-values items and over- or under-pays legibly). **One caution:** in a table of *all* bots the auction can degenerate (everyone values identically → predictable), the `creature-battler.md §3.4` echo — so, per the async law, keep **≥1 human at the table** and use bots for backfill, not as a self-contained bot-vs-bot economy.

**AI-fallback difficulty: LOW–MEDIUM.** Genuinely multiplayer but the bot is honest and one-screen; backfill works because bidding is per-round-simultaneous-ish (Power Grid's plant auction is sequential, but each decision is small).

**Non-adopted tech-satire mapping:** a capacity auction could use compute supply and rising spot
prices as its metaphor. Exact names, copy, market rules and links to another minigame require
owner/design work.

---

## 8. Route / network building — Ticket to Ride

**The exemplar.** **Ticket to Ride** (Alan R. Moon, Days of Wonder, **2004**): each turn you **draw coloured train-car cards, claim a route between two cities by spending matching cards, or draw destination tickets**; **destination tickets are secret goals** (connect city A to city B) that score if completed and **penalise if not**; bonuses for longest continuous route; only one player can claim each route segment, so the map is a **contested shared network** `[V]`. **18 million copies sold** by 2024 `[V]` — the genre's mass-market anchor. (Set collection lives inside it — the coloured cards — and `set collection` as a standalone family is covered in §9.)

**Why it's fun.** **A hidden-goal network race with blocking.** The public tension is the map filling up (claim the route now or risk it being taken); the private tension is your **secret tickets** — you're building toward goals opponents can't see, and *they're doing the same*, so the map is a partial-information puzzle where a claimed route might be someone's critical link. The push-your-luck garnish: **take more tickets for more points, risking the penalty if you can't connect them.**

**Tractability — server-cheap, bot-easy.** Input alphabet: *(draw card / claim route r with card set / draw tickets)* — clean enums, seeded card deck and ticket deck, replay-trivial. **Bot: easy.** Route-building reduces to a **shortest-path / set-cover problem over a small graph** — a bot that plans the cheapest route set to connect its tickets, then claims greedily while blocking an opponent's obvious link, is standard graph code and plays well. Difficulty = planning horizon + how aggressively it blocks + ticket-count risk appetite. Hidden tickets are hidden-info-symmetric (the bot's tickets are hidden from you too, law 6). **This is one of the most bot-friendly families here** — the domain is literally a graph algorithm.

**AI-fallback difficulty: LOW.** Fully solo-playable; the bot is a graph planner, honest and strong.

**Tech-satire mapping: lay fiber / peering agreements — the brief's own example.** **"Peering"** / **"Lay Fiber"**: routes are **network links between datacenters/IXPs/regions**; coloured cards are **bandwidth/permits of a given carrier**; **destination tickets are SLAs** (secretly commit to connect `us-east` to `ap-south`; score on delivery, pay a penalty on breach); longest-route bonus = **backbone**; a claimed link is **capacity a rival can no longer lease.** The internet-infrastructure-as-monopoly satire is right on the megacorp arc ("we don't own the internet, we own the *cables*"). Alternatively a **supply-chain / dependency-graph** skin (npm-install-the-world) fits the same engine.

---

## 9. Set collection & push-your-luck

These two are less "standalone genres" than **verbs that appear inside the families above**, but the brief asks for them and they each anchor a distinct minigame feel.

### 9.1 Set collection

**What it is** `[M]`: score by assembling matched groups; the family appears within drafting,
route-building and traditional melding games. Its hypothesized appeal combines collection with the
tension of holding versus cashing.

**Tractability: LOW / trivial.** Input alphabet is *(take card / play set)*; scoring is a table lookup over your groups; the bot is "collect toward the highest-value reachable set, cash when threatened" — one screen. This is the **easiest family in the dossier** and it's already latent in our roster (Server Garden's strain-collection, the Surprise Crate workloads of the lane). It rarely stands alone; it's the **scoring layer** to bolt onto a draft or a market.

**Tech-satire mapping: collect the full stack / the compliance-badge set.** Assemble matched sets of **certifications** (`SOC2 + ISO + HIPAA` = a compliance combo), **a full microservice suite**, or **the six Temperament workloads** (ties to the shared combat data model). "Gotta ship 'em all."

### 9.2 Push-your-luck — Can't Stop and the dice family

**The exemplar.** **Can't Stop** (Sid Sackson, Parker Brothers, **1980**): roll **four dice, split into two pairs**, advance markers up the columns those pair-totals name; **after each roll choose to roll again or bank your progress** — but if a fresh roll **can't be assigned to any of your three active columns, you bust and lose all un-banked progress** this turn; first to top out three columns wins `[V]`. The family also includes Yahtzee, Farkle, and liar's dice (`03 §5b` dice/bluff row).

**Why it's fun.** Escalating commitment against rising bust probability creates a legible
push-your-luck tension. It resembles the Ship-It Spellbook backfire concept described at the dated
design coordinate; this dossier does not claim that player surface exists at HEAD.

**Tractability: LOW to certify, but note the chance-variance caveat.** Input alphabet: *(roll / stop)* + *(assign dice to columns)* — tiny, and the RNG is **seeded dice** (re-derivable, daily-seed-friendly). Certification trivial. **Bot: trivial** — the optimal-ish policy is a closed-form EV threshold ("stop when marginal bust risk × stake > expected gain"), literally a formula; difficulty = deviating from EV by a tunable temperature (a *reckless* bot pushes past EV and busts — legible personality). **The real caveat is law-shaped, not tech-shaped:** dice variance drowns skill at low match counts, so per `03 §5b` rule 1, **a push-your-luck game can only enter a *ranked* ladder via duplicate scoring** (same seeded dice to every seat, score relative). Casual/solo play is raw-roll and fine.

**AI-fallback difficulty: TRIVIAL.** Solo by nature; the bot is a formula.

**Tech-satire mapping: ship-to-prod-on-Friday roulette / the funding-runway gamble.** **"Friday Deploy"**: each "roll again" is **one more change shipped without a full test pass** — push your luck for more features live before the weekend, but a bad roll is a **prod incident that wipes the sprint's un-shipped work**. Or **"Extend The Runway"**: keep raising at a higher valuation (push) vs. take the safe term sheet (bank), bust = **down-round wipes your paper gains.**

---

## 10. Social deduction — Werewolf, Avalon, The Resistance — **the hard "no," confirmed**

**The exemplars.** **The Resistance** (Don Eskridge, **2010**): a hidden-role game — **one-third of players are secretly Spies (who know each other); the rest are Resistance (who know only the spy *count*)**; a rotating leader **proposes a mission team, everyone votes simultaneously to approve/reject, and approved team members secretly submit success/fail cards** — one fail sinks the mission; the game is won by deducing identities across five missions, and it is explicitly built on **"extensive discussion — negotiate, argue, and deduce identities through debate"** `[V]`. **Werewolf/Mafia** (the ancestor): night-kill + day-lynch with hidden wolves; **Avalon** (a Resistance variant) adds Merlin/assassin role asymmetry.

**Why it's fun — and why that fun is exactly what we cannot ship.** The entire mechanism *is* **humans reading humans through free-form talk.** The votes and mission cards are a thin structured skeleton; the game lives in the **accusation, defense, bluff, and tell-reading** that happens between them. Strip the talking and you have a near-random voting minigame.

**This violates TWO binding laws simultaneously — it is the one family in this dossier I flag as non-compliant, not merely hard:**

1. **The no-free-text law (`03 §5b`, `08` voice rules).** Social deduction's medium *is* unstructured communication. `03 §5b` already excludes free-form expression games (Pictionary/charades) with "no exception," and social deduction is the same class one layer up. A **fully menu-driven** variant ("who leaked the roadmap?" with structured accusations from a fixed list) is *conceivable* — `03 §5b` flags exactly this — but `03 §5b`'s own bar stands: **do not ship until the menu-chat design proves richer than a coin flip**, and my read is that it rarely will, because the skill you're digitising is *reading a human*, and a dropdown of canned accusations reads a human about as well as a horoscope.

2. **The AI-fallback law (law 6 + `06-tech.md §vs-AI`).** *Every* multiplayer feature needs a bot that plays by the same rules and doesn't cheat. Social deduction's bot must **bluff convincingly and detect bluffing** through the communication channel — and if that channel is free text, an honest bot (no hidden info, law 6) is either trivially readable or requires an LLM roleplaying deception, which (a) is not "the same server validation and hidden-info boundaries as humans" in any clean sense, (b) is exactly the kind of ungoverned LLM surface the project has no appetite for, and (c) would be a *different game* than the human one, failing the "fallback" premise. If the channel is menu-only, the bot is honest but the game is a coin flip (see law 1 above). **There is no version that satisfies both the no-free-text law and a non-cheating, same-rules bot fallback while remaining fun.**

**Server-authority itself is fine** (votes and face-down mission cards are trivially certifiable enums) — that is *not* the blocker. The blockers are the **communication medium** and the **bot**, and they are structural, not tuning.

**Disposition (matches `03 §5b`'s existing flag exactly):** **do not build.** Keep it on the flagged/backlog shelf; revisit only if a Break Room v2 delivers a **rich structured-accusation grammar** that provably beats a coin flip in playtest, and even then never as a headline or a ranked ladder. If a social-deduction *flavor* is wanted cheaply, express it as **PvE narrative** ("find the mole" as a single-player logic-deduction puzzle against a scripted scenario — which is a *different, compliant* genre: a Clue-style deduction puzzle, seeded and bot-free-by-construction).

**Tech-satire mapping (for the record, if a compliant PvE-deduction version is ever built):** **"Find The Mole"** / **"Who Leaked The Roadmap"** — a structured-accusation whodunit among NPC coworkers; a **security-incident post-mortem** where you deduce which service/insider caused the breach from menu-selected evidence. Ship it as a **solo logic puzzle**, never as human-vs-human hidden-role.

---

## 11. Verify before ship

| # | Claim | Status |
|---|---|---|
| 1 | **Carcassonne:** Wrede, Hans im Glück, 2000; draw-and-place matching-edge tiles; meeples claim features; cities 2/tile + 2/pennant, roads 1/tile, fields score at end; 72 base tiles | `[V]` [Wikipedia](https://en.wikipedia.org/wiki/Carcassonne_(board_game)) |
| 2 | **Azul:** Kiesling, Plan B, 2017; draft all tiles of one colour from a factory/centre; fill pattern rows → 5×5 wall; adjacency + row/column/colour bonuses | `[V]` [Wikipedia](https://en.wikipedia.org/wiki/Azul_(board_game)) |
| 3 | **Azul overflow penalty specifics** | Not established by the fetched source and removed from the factual body; verify before reintroduction. |
| 4 | **Patchwork specifics** | Not fetched and removed from the factual body. |
| 5 | **Agricola:** Rosenberg, Lookout, 2007; one worker per action space per round; feed family at harvests or lose VP; 14 rounds | `[V]` [Wikipedia](https://en.wikipedia.org/wiki/Agricola_(board_game)) |
| 6 | **Lords of Waterdeep / Everdell specifics** | Not fetched and removed from the factual body. |
| 7 | **Splendor:** André, Space Cowboys, 2014; gem chips → development cards with permanent discounts; 15 prestige to win; 90 dev cards / 10 nobles | `[V]` [Wikipedia](https://en.wikipedia.org/wiki/Splendor_(game)) |
| 8 | **Wingspan:** Hargrave, Stonemaier, 2019; card-driven engine-builder; three habitats each tied to an action; birds chain abilities | `[V]` [Wikipedia](https://en.wikipedia.org/wiki/Wingspan_(board_game)) |
| 9 | **Terraforming Mars specifics** | Not fetched and removed from the factual body. |
| 10 | **7 Wonders:** Bauza, Repos, 2010; 2–7 players; 3 ages/18 turns; simultaneous pick-one-and-pass; tableau of resource/military/science/civic/commerce | `[V]` [Wikipedia](https://en.wikipedia.org/wiki/7_Wonders_(board_game)) |
| 11 | **Sushi Go specifics** | Not fetched and removed from the factual body. |
| 12 | **Power Grid:** Friese, Rio Grande 2004 (Eng 2nd ed); open plant auction; escalating resource market; reverse-turn-order resource buying; network build | `[V]` [Wikipedia](https://en.wikipedia.org/wiki/Power_Grid) |
| 13 | **Modern Art specifics** | Not fetched and removed from the factual body; verify from a durable source before reintroduction. |
| 14 | **El Grande:** Kramer & Ulrich, Hans im Glück 1995; caballeros → region majorities (1st/2nd/3rd score); card bidding 1–13 sets turn order/power; Castillo hidden reveal; scoring in 3 of 9 rounds | `[V]` [Wikipedia](https://en.wikipedia.org/wiki/El_Grande) |
| 15 | **Blood Rage specifics** | Not fetched and removed from the factual body; verify from a durable source before reintroduction. |
| 16 | **Ticket to Ride:** Moon, Days of Wonder 2004; draw cards / claim routes / draw tickets; secret destination tickets score-or-penalise; longest-route bonus; 18M copies by 2024 | `[V]` [Wikipedia](https://en.wikipedia.org/wiki/Ticket_to_Ride_(board_game)) |
| 17 | **Can't Stop:** Sackson, Parker Brothers 1980; 4 dice → 2 pairs → columns 2–12; bust if a roll fits no active column; win by topping 3 columns | `[V]` [Wikipedia](https://en.wikipedia.org/wiki/Can%27t_Stop_(board_game)) |
| 18 | **The Resistance:** Eskridge, 2010; ~1/3 spies (know each other), resistance know only spy count; leader proposes team, simultaneous approve/reject vote, secret mission success/fail cards; built on discussion | `[V]` [Wikipedia](https://en.wikipedia.org/wiki/The_Resistance_(game)) |
| 19 | **The tractability thesis** (board games are the cheap-to-certify end; bots range trivial→moderate; social deduction is the sole non-compliant family) | **Analysis, not measurement** — grounded in binding laws 2/6 + `lane-pusher-design.md §4` bot-cost findings + `03 §5b`. Pressure-test any family in a prototype before an RFC leans on the bot-difficulty claims. |
| 20 | **Engine-building ≈ our production buff-stack** (`base × Π(buff)` / discount-chaining maps onto our core; Fairness-Law sandboxing required) | `[M]`/analysis — the strongest core-fit claim here; confirm the buff-stack resolver exposes the ops a card game needs and that the C5 `scaling_inputs` seam can gate it to `breadth` before scoping an RFC. |
| 21 | Patent exposure for any mechanic here | ❌ **Not searched.** Standing caveat from `lane-pusher-design.md §7`/`roguelike §7`: mechanics are copyright-safe (Tetris v. Xio) but not automatically patent-safe. Run a real search before an RFC. Exposure probably nil (free, EU-ish, non-commercial), but "probably" is doing work. |

---

## 12. Design implications for Cloud Clicker

### 12.1 Prototype-hypothesis ranking: tractability, economy hook and differentiation

This qualitative ordering is a hypothesis for prototypes, not a roadmap. Its server/bot labels are
unmeasured, and its dated “already ship” comparisons are not claims about HEAD.

| Rank | Family | Tractability | Bot / AI-fallback | Economy-hook fit | New feel vs. our roster? | Verdict |
|---|---|---|---|---|---|---|
| **1** | **Engine-building / tableau** (Splendor, Wingspan) | 🟢 turn-based, replay `(seed, choices)` | 🟢 easy — multiplayer-solitaire, bot plays its own engine | 🟢🟢 **IS our buff-stack core** (with §4 sandbox discipline) | 🟡 same *math* as the economy, but a bounded card-game *feel* | **Prototype-worthy.** Best core-fit in tabletop; the one to build if we build one. Guard the Fairness-Law boundary. |
| **2** | **Drafting** (7 Wonders, Sushi Go) | 🟢🟢 smallest input alphabet | 🟢🟢 **best multiplayer fit** — simultaneous commit, invisible backfill | 🟢 Clout/cash OUT hook; daily-seed friendly | 🟢 genuinely new: N-player async social, unlike our 1v1 duels | **Prototype-worthy.** The best *multiplayer* board game for us; Break-Room furniture (`§5b`). |
| **3** | **Route / network building** (Ticket to Ride) | 🟢 turn-based, seeded decks | 🟢 easy — bot is a graph planner (strong, honest) | 🟢 secret-SLA goals + longest-backbone bonus | 🟢 new feel: hidden-goal graph race | **Strong single-player-vs-bots pick;** on-theme (lay fiber). Content-light. |
| **4** | **Tile-laying** (Carcassonne, Azul) | 🟢 turn-based, seeded bag | 🟢 easy — 1-ply heuristic strong | 🟡 the Azul overflow-penalty = cloud-bill satire is a *great* hook | 🟢 new feel: spatial optimisation | **Good.** Azul's "provisioning waste is billed" is the sharpest single satire beat in the dossier. |
| **5** | **Worker placement** (Agricola) | 🟢 turn-based | 🟡 **moderate** — needs eval function + denial modelling | 🟢🟢 "allocate eng across the sprint" is a bullseye theme; payroll = burn | 🟡 action-selection is fresh but adjacent to drafting | **Good, higher bot cost.** Strong theme; budget for the eval function. |
| **6** | **Area control / majority** (El Grande) | 🟢 turn-based | 🟡 moderate — board-eval + denial | 🟢 market-share satire; menu-safe face-down bluff | 🟡 adjacent to worker placement in feel | **Fine, mid-priority.** Good if we want a *contest-the-map* feel; watch bot cost. |
| **7** | **Auction / bidding** (Power Grid, Modern Art) | 🟢 turn-based | 🟡 tractable, honest bot; degenerates if all-bot | 🟢 "bid for GPU capacity" is *the* 2026 theme; rhymes with The Market | 🟢 new feel: price-discovery-as-weapon | **Fine, keep ≥1 human.** Great theme; the all-bot degeneration caveat caps it. |
| **8** | **Push-your-luck** (Can't Stop) | Hypothesis: small input surface | Hypothesis: EV policy may suffice | Dated design overlaps Ship-It Spellbook; HEAD status not inferred | Potentially redundant | Prototype only if a distinct theme survives current design review. |
| **9** | **Set collection** (standalone) | 🟢🟢 trivial | 🟢🟢 trivial | 🟡 collector's itch, already latent (Server Garden, crates) | 🔴 rarely stands alone; it's a scoring *layer* | **Not a standalone build.** Bolt onto a draft/market as the scoring rule. |
| **—** | **Social deduction** (Werewolf, Avalon, Resistance) | 🟢 votes/cards certifiable | 🔴🔴 **NON-COMPLIANT** — no honest, same-rules, fun bot exists | — | — | **DO NOT BUILD.** Violates the no-free-text law *and* the AI-fallback law simultaneously (§10). Confirms `03 §5b`'s existing flag. |

### 12.2 Non-adopted candidates for owner/design consideration

1. **Engine-building / tableau ("Vertical Integration" — a Splendor/Wingspan-like).** It is the tabletop family that *is* our economy — discount-chaining and production-ramp cards are our buff-stack, so the build is closer to "expose the engine as a bounded card game" than to a new system, and the theme (own your compute stack → everything downstream gets cheaper) is the megacorp arc in miniature. **The one hard requirement: sandbox it in its own seeded economy** and let tier/building counts enter only as `breadth` via the `03 §12b` C5 seam — never as power in a rated board (Fairness Law). Pairs naturally with the roguelike dossier's Balatro recommendation: both are "our multiplicative core wearing a hat," and shipping one card-engine game may satisfy both niches — **decide deliberately whether Balatro-poker or Splendor-tableau is the single engine-building minigame, rather than shipping both.**

2. **Drafting ("The Draft" — a 7 Wonders/Sushi Go-like).** The best *multiplayer* board game for us: simultaneous pick-and-pass means zero turn-order latency, invisible bot backfill, a one-screen card-evaluator bot, and a natural home on the 4-seat Break Room furniture (`03 §5b`). It fills the "genuinely social, genuinely async, genuinely fair" slot that our 1v1 duels and the lane don't — and unlike a fourth snapshot-battler it is a *new loop*, not the same fight re-skinned. Highest multiplayer value per unit of build cost.

3. **Route/network building ("Lay Fiber" — a Ticket to Ride-like) OR tile-laying ("Capacity Planning" — an Azul-like).** Pick one as the solo-vs-bots evergreen. Route-building's bot is a strong honest graph planner and the secret-SLA hidden-goal race is a fresh feel; Azul's edge is the **single sharpest satire beat in the dossier** — over-provisioning is *billed as waste*, the cloud-bill joke rendered as a core rule. If forced to one, I lean **Azul-like** for the satire density and lower content weight; **Ticket-like** if we want the map/network fantasy that ties to the "consume the internet's infrastructure" arc.

### 12.3 The three cross-cutting rulings

- **Engine-building maps onto our core — flag to the owner as the tabletop counterpart of the Balatro finding.** `base × Π(buff)` is our production formula and Splendor's discount-chain and Wingspan's habitat-actions are the same shape. This is an *opportunity* (instant fit, engine already built) and a *hazard* (`03 §12b` Fairness Law — sandbox it or it becomes pay-to-win). Both card-engine dossiers now point at the same conclusion: **the one place a minigame should touch the numeric core is a sandboxed, seeded engine-builder — and probably only one of {Balatro-poker, Splendor-tableau} should ship.**

- **Redundancy is the real constraint, not tractability.** *Everything here except social deduction is cheap to certify and bot.* The discriminator is feel: **drafting (N-player async social), engine-building (economy-bound), and route/tile-laying (spatial) each add something our current roster lacks**, whereas push-your-luck duplicates the Spellbook and set-collection is a scoring layer, not a game. Build for *new feel*, not for *another certifiable turn-loop*. `DESIGN-GAP:` before any of these gets an RFC, `03 §5b`'s parlor taxonomy should record *which specific tabletop mechanic-games we intend to ship*, so we don't build three adjacent "commit-and-resolve" Euros that feel the same.

- **Social deduction is the one law-violating family — and it violates two laws, not one.** Server-authority is *not* the problem (votes are certifiable); the problem is that its medium is free-form human communication (no-free-text law) and its bot must bluff/detect-bluff through that medium (AI-fallback law), and **no design satisfies both while staying fun.** `03 §5b` already flags it; this dossier upgrades the flag from "design-risk" to **"non-compliant with two binding laws — do not build the hidden-role human-vs-human form."** The only compliant expression is a **PvE structured-deduction puzzle** ("find the mole," seeded, bot-free by construction), which is a different genre (Clue-like) and should be evaluated on its own merits, not as social deduction.

---

## Sources

**Fetched this pass (Wikipedia):** [Carcassonne](https://en.wikipedia.org/wiki/Carcassonne_(board_game)) · [Azul](https://en.wikipedia.org/wiki/Azul_(board_game)) · [Agricola](https://en.wikipedia.org/wiki/Agricola_(board_game)) · [Splendor](https://en.wikipedia.org/wiki/Splendor_(game)) · [Wingspan](https://en.wikipedia.org/wiki/Wingspan_(board_game)) · [7 Wonders](https://en.wikipedia.org/wiki/7_Wonders_(board_game)) · [Power Grid](https://en.wikipedia.org/wiki/Power_Grid) · [El Grande](https://en.wikipedia.org/wiki/El_Grande) · [Ticket to Ride](https://en.wikipedia.org/wiki/Ticket_to_Ride_(board_game)) · [Can't Stop](https://en.wikipedia.org/wiki/Can%27t_Stop_(board_game)) · [The Resistance](https://en.wikipedia.org/wiki/The_Resistance_(game))

**Internal cross-references:** `research/lane-pusher-design.md` (§4 async/AI, §5 determinism/MVP, §6 redundancy warning) · `research/roguelike-survivor-minigames.md` (the deckbuilder/Balatro/auto-battler family — deliberately not re-covered here; §5 the buff-stack-maps-onto-core finding this dossier's §4 echoes) · `research/creature-battler.md §3.4` (auto-resolve law) · `design/03-minigames.md` (roster, clock taxonomy, §5 board-game suite, §5b parlor taxonomy, §12b Fairness Law + C5 scaling-inputs seam) · `design/06-tech.md` (`§vs-AI`, fixed-timestep sim, Match goroutine) · design laws 2/3/5/6 (server-authority, big numbers, hardcaps, AI-fallback) and the no-free-text law (`03 §5b`).

**Legal issue-spotting:** *Tetris Holding, LLC v. Xio Interactive, Inc.* illustrates that copying
expression can create liability even when general mechanics are discussed. One US case is not
project-wide clearance; no patent/trademark search was performed (§11 item 21).
