# Minigames Catalog

> Rules for every minigame: (1) it runs on a **distinct clock** (the Cookie Clicker lesson — garden/pantheon/grimoire/market each had a different time signature); (2) it **hooks into the main economy** in a stated way; (3) it has an **AI fallback** so it's fully playable with zero other players; (4) unlocks are **staggered across the tier arc** (fixing CC's all-at-once dump and the hour-5–40 sag); (5) state **persists across Exits** (fixing the Stock Market resentment) unless reset is explicitly the point.

Minigames unlock **at their host tier**; from Tier 2 onward the unlock is *purchased* with Investor Confidence (Fiscal Quarters — the sugar-lump pattern). **Tier 0–1 minigames (Arcade, Typer) are tutorial-tier and free** — the arcade hosts the tutorial, so it cannot sit behind a currency that first ripens on day 2.

## Clock taxonomy

| Clock type | Meaning | Games on it |
|---|---|---|
| Wall-clock | real time, production-immune | Server Garden, Pet care |
| Regenerating pool | resource refills on its own curve | Incident Response, Ship-It Spellbook |
| Swap budget | loadout changes cost a slow-regen token | Advisory Board |
| Rate-denominated | currency = seconds of your production | The Market |
| Turn-based match | opponent's move is the clock | Board-game suite, Pet battles |
| Session skill | short skill runs, cooldown-gated | Terminal Typer, Demo Disc arcade |

---

## 1. Server Garden (Tier 2, host: office/campus)

The garden equivalent — **crossbreeding discovery game on wall-clock ticks**.

- Grid of racks/planters (2×2 → 6×6 with building levels). You "plant" **server strains / software cultivars** (fictional distro-species: `Slackware Bonsai`, `Kube Vine`, `Legacy Perl Bramble`…). Mature units seed adjacent slots; ~30 species discovered via adjacency crossbreeding with sub-1% per-tick mutation odds; seeds permanent once harvested (Pokédex energy).
- Soils → **substrate choices** (Bare Metal = baseline; Containerized = fast ticks, −25% effect; Mainframe = slow, +25%; Chaos Monkey = ×3 mutation, reduced effect).
- Payloads touch the main economy: strains grant uptime buffs, golden-opportunity frequency, daemon spawn rate, Clout drops, one-shot cash harvests.
- **Sacrifice ritual:** delete the full collection for a permanent bonus + Confidence (the sacrifice-for-permanence pattern).
- **AI fallback:** none needed (solo). **MMO hook:** rare strains are tradeable gifts; a community seed-census feeds a collection milestone.

## 2. Advisory Board (Tier 3, host: HQ)

**Loadout with drawbacks + swap friction** (the advisor-loadout pattern).

- Three seats in descending influence: **Boardroom > Backchannel > Group Chat** (seat strength scales the advisor's effect). Slot advisor archetypes, each with an upside *and a drawback* (house rule: no free advisors): e.g. `The Growth Guy` (+click power, Trust drain), `The Ascetic CTO` (+idle rate, unslots if you click a golden opportunity), `The Fixer` (sell-off combo enabler), `The Oscillator` (a ±15% sine on 3/12/24 h real-time cycles you can time combos to), `The Safety Hire` (−p(doom), −speed).
- **Swap budget:** 3 banked swaps, regen 16 h / 4 h / 1 h; Caramelized-quarter refills. Loadout changes are decisions, not menus.
- **No universal optimal setup** is a design requirement — verified per patch in the balance harness.
- **AI fallback:** none needed (solo).

## 3. Ship-It Spellbook (Tier 3, host: engineering org)

The grimoire equivalent — **regenerating pool + push-your-luck**.

- **Velocity** pool (max scales with engineers, regen deliberately slows as the pool grows — the CC anti-scaling brake). Spend on "ships": `Hotfix` (instant cash = 30 min of rate; backfire: incident), `Force The Demo` (summon a golden opportunity; backfire: summon an outage), `Crunch Sprint` (+buff durations; ages the founder, drains Soul), `Spin Up Side Project` (free generator; backfire: destroys one), `Blameless Postmortem` (backfire chance ÷10 for 5 min).
- Base backfire 15%. **RNG is seeded server-side per cast and NOT deterministic-predictable** — the FtHoF lesson: we don't ship an external-tool meta. Instead, a **built-in forecast panel** (unlocked late) shows the next cast's odds honestly — the predictor is in the game.
- **AI fallback:** none needed (solo).

## 4. The Market (Tier 4, host: finance dept)

The stock-market equivalent, **rate-denominated and persistent**.

- `$1 = 1 second of your highest unbuffed rate this run` — the rate-indexed pricing trick: buy low, *raise your own rate*, sell — printing money off your own growth is the intended aha.
- Goods = fictional sector indices (GPUs, RAM, Compliance, Attention, Water futures — the satire writes itself). Resting-value drift + momentum + six market modes.
- **Loans** for combo windows (+50–100% rate, penalty after).
- **Persists across Exits** (holdings survive; the fiction: your personal brokerage, not the company's).
- **MMO hook:** prices are *server-global* and lightly influenced by aggregate player trading + world events (a Situation like "Chip Shortage" moves the GPU index for everyone). Insider-satire events fire when you trade ahead of your own announcements (Trust drain).
- **AI fallback:** market makers are simulated; no humans required.
- **NEODAQ lessons (2026-08-01, `research/neopets-systems.md §2`):** the beloved 26-year shape is
  daily-capped purchases (the faucet governor), a visible buy floor + price floor creating a
  one-sentence teachable meta, flat fees, and returns priced in calendar time (idle-native). Two
  laws adopted: **the daily purchase cap is load-bearing** (an uncapped bounded-downside market
  is a faucet exploit), and **positions are never manually touched** (their Nov-2004 manual
  bankruptcies are remembered as the original sin 20+ years later — our epoch protocol already
  forbids it; now it's also fiction: "the SEC does not exist here, but we do"). Decide the joke
  deliberately: bounded-downside = index-fund satire; lossy = WSB satire. Ours is bounded
  (design/00's kindness) with fake tickers whose news may or may not correlate with the walk.

## 5. Board-Game Suite (Tier 2+, host: break room — "hobbies")

Chess puzzles, chess, tic-tac-toe, Connect-4, Othello, Gomoku, checkers; blackjack and poker at the casino tier (Tier 5's regulator-baiting `Loot Box Casino (Free)`).

- **Framing:** these are the founder's *hobbies* — they **lock** at low Soul ("you no longer remember why this was fun"), the retention hook *and* the theme. **They do NOT *restore* Soul** (ruled 2026-08-06): recovery is the cozy/touch-grass category ONLY — a rewarding, *paying* activity isn't restful (the Stardew trap). Board games are active content that pays Clout+cash; the zero-reward cozy minigames (`§5c`) are the sole Soul-recovery source.
- **Clock:** turn-based matches; ranked queue with cooldowns; daily chess puzzle (wall-clock).
- **Economy hooks:** wins pay Clout + small cash scaled to tier; achievements (beat the top bot, win N ranked games) feed Clout; a daily-puzzle streak feeds a small multiplier. Bot matches pay reduced, non-ranked rewards (anti-farm).

  > **Resolved 2026-07-28:** streak bonuses here and in §7 key on **cumulative days played (count-up, never reset)** per `02 §10` — the daily/weekly session scaffold now owns this pattern. Absence is never priced (`00 §6`).
- **Design identity (added 2026-08-01): the classics are the game's FLAT layer.** Under the
  Fairness Law (`§12b`) nothing here scales with tiers — no building count, upgrade, or prestige
  level moves a ranked board-game outcome, ever. That makes this suite the one surface idle
  progression cannot buy: **an eternal skill ladder — the only numbers in the game money can't
  inflate** — which is simultaneously the anti-pay-to-win statement, the satire's control group,
  and evergreen content that never needs rebalancing. Scaling inputs allowed: breadth (variants,
  board sizes) and era-authentic cosmetics only.
- **Three surfaces per game:** the daily puzzle (solo, wall-clock, streak-fed — tactics/tsume/
  life-and-death); the ranked queue (real ratings, bot backfill per the combat-bots contract);
  and **casual boards as furniture in the Guild Break Room** (`05 §3`) — a board is a seat-pair
  object, spectate-able, the co-presence ritual's natural anchor. Hobbies are social fixtures
  first, competitions second.
- **The ilk arrives with the world (internationalization beat):** T0–1 lunch-break games
  (tic-tac-toe, checkers), chess from T2. **Regional classics unlock with Tier-4 global
  expansion — opening a region office opens its parlor:** xiangqi and Go with APAC regions, shogi
  with Japan. Localization satire included ("our Global Expansion strategy deck lists 'board game
  parlors' under synergies").
- **Go is the AI tier's narrative beat.** From Tier 6 your own AI *offers* to play Go — and always
  wins (the Move 37 moment, played straight). Achievement: `Move 37`. And the Soul rule sharpens:
  **letting your AI play your hobbies FOR you drains Soul** — automation of the one thing that was
  restoring you is the whole thesis in a single mechanic.
- **AI fallback (the flagship of the vs-AI system, `06-tech.md §vs-AI`):**
  - Chess: Stockfish (skill level + shallow depth + top-k softmax + fake think time); Maia later if headline. Rules server-validated via notnil/chess.
  - **Shogi/xiangqi (and chess variants): Fairy-Stockfish** — one engine binary covers chess,
    shogi, xiangqi and dozens of variants `[M — verify engine + license fit before the minigame
    RFC]`. Go: 9×9 boards with a light MCTS at parlor ranks; KataGo only if Go becomes headline
    `[M]`. The AI-tier "your AI plays Go" beat needs strength ABOVE the player, which small-board
    engines deliver cheaply — full 19×19 strength is deliberately not promised.
  - Connect-4/Othello/Gomoku/checkers: ~150-line alpha-beta with iterative deepening + per-game eval; difficulty = depth + temperature.
  - Blackjack: basic-strategy table bot. Poker: rule-based hand-strength + pot odds + randomized bluffs.
  - Human-feel rules: plausible blunders (captures/checks, not random), fake think time, resign behavior — and **real published ratings with rating-based matchmaking** (decided 2026-07-28 per `research/adaptive-balancing.md §2`): a bot genuinely rated 1400 plays like 1400 and needs no rubber band to give a 1400 player ~55% — matchmaking produces the target win rate honestly. No hidden outcome adjustment, ever; **bots never cheat** (same server validation and hidden-info boundaries as humans).
- **PvP:** ranked matchmaking with expanding Elo bands, bot backfill after 10–30 s (disclosed, consistent, reduced rewards, never hot-swapped mid-match), spectating, guild tournaments. Show queue depth to encourage waiting for humans.

## 5b. The parlor taxonomy (added 2026-08-01)

> The classics are not two categories (chess-likes + casino) but a full space. Mapped against our
> four placement axes: **social geometry** (solo / 1v1 / 4-seat table / house / guild-scale),
> **chance spectrum** (pure skill → mixed → pure chance), **era/region arrival** (which tier or
> region office opens the parlor), and **the no-free-text law** (menu-safe or excluded).

| Family | Examples | Geometry | Chance | Arrival | Notes |
|---|---|---|---|---|---|
| Abstract strategy | chess, checkers, Go, shogi, xiangqi, Othello, Gomoku | 1v1 ranked + daily puzzle | none | T2 + region offices (§5) | The flat layer; already specced |
| Trick-taking | Hearts, Spades, Euchre, klaverjas, Bridge | **4-seat table, partnerships** | mixed | T2–3 (MSN-Zone/internet-café era skin) | THE Break Room furniture — partnerships create the social bonds the seat-room exists for; regional variants with region offices (klaverjas = NL office) |
| Melding/tile | Gin Rummy, Canasta, mahjong (real 4-player), dominoes, Rummikub | 4-seat table | mixed | mahjong/dominoes with T4 region offices | Dominoes hall + mahjong parlor culture = social-space texture |
| Shedding | President/Asshole, Crazy Eights, Durak | 3–6 table | mixed | T2 | **President is mandatory content: the card game that IS corporate hierarchy** (winner president, loser trades their best cards up — the satire writes itself) |
| Racing | backgammon, Ludo/Parcheesi, Snakes & Ladders | 1v1 / 4-seat | mixed (dice) | backgammon T2; Ludo with region offices (huge in South Asia) | Backgammon's doubling cube is a stakes mechanic worth stealing for Clout wagers |
| Mancala family | oware, bao | 1v1 | none | region offices (Africa/ME) | Flat-layer member; ancient-games texture |
| Solo dailies | Sudoku, picross/nonogram, Wordle-shape daily, crossword-lite | solo, wall-clock | none | T0–1 onward, one unlocking per era | Extends the daily-puzzle streak scaffold; cheap evergreen content |
| Paper-and-pencil duels | Battleship, dots-and-boxes | 1v1, async-friendly | low | T0–1 (office-memo skin) | Async = idle-friendly PvP precedent (play a move per visit) |
| Dice/bluff | Yahtzee, Farkle, liar's dice | table | high | T3+ | Liar's dice = bluffing without free text (bids are menu picks) |
| Casino (house) | blackjack, poker, plinko, craps | vs house / table | high → satire | T5 casino tier ONLY | Already specced as regulator-bait; never ranked, pays satire not power |
| Trivia/quiz | pub-quiz nights | **guild-scale party** | none | any; guild feature | Multiple-choice = menu-safe; **the natural weekly Break Room ritual content** (`05 §3`'s booked 30-minute co-presence slot wants exactly this) |
| Social deduction | Werewolf-likes | 6–12 party | none (human-only carve-out) | **HUMAN-ONLY** | **Ruled 2026-08-06: the designated human-only feature** — AI fallback is genuinely too much hassle AND would have to cheat on hidden info (no honest same-rules bot exists), so it takes the plan's "human-only where AI is too much hassle" carve-out. **STRUCTURED communication only** (preset claims/accusations/votes — never free-text, so no moderation / no-free-text-law problem). No bot backfill: an empty lobby makes it *unavailable* (acceptable for a human-only mode). Ship only once the structured-menu deduction proves richer than a coin flip; Break Room v2 |
| Drawing/charades | Pictionary-likes | party | none | **excluded** | Free-form expression violates the no-free-text law; no exception |

**Two placement rules the table implies:**

1. **Ranked chance games use duplicate scoring.** A card/dice game can never enter a real ladder
   on raw outcomes — luck drowns skill at our match counts. The fix is the bridge world's:
   **duplicate deals** — the same seeded shuffle (SplitMix64, naturally) dealt to every
   table/pair in a rated event, scored on relative performance. Our determinism infrastructure
   makes this nearly free, it extends the flat-layer fairness guarantee into the mixed-chance
   families, and it is exactly the speedrun sensibility (same seed, compare execution) applied to
   cards. Casual tables stay raw-shuffle.
2. **4-seat table games are Break Room furniture first** (the trick-taking/mahjong/dominoes
   families exist to fill seats and create partnerships), solo dailies extend the streak scaffold,
   and guild-scale content (trivia) anchors the weekly ritual. Every family lands on an existing
   social surface — no new venue needed.

## 6. Incident Response (Tier 3, host: on-call rotation)

An original — **the pager as a push-your-luck reflex minigame** on a regenerating clock.

- Incidents queue on a Poisson-ish timer (97% are noise — the 3%-Signal satire). Triaging is a fast pattern-matching mini-loop (match the error signature; wordle-ish deduction against a symptom grid). Real incidents pay big uptime buffs; false-alarm grinding drains a Burnout meter that debuffs if ignored.
- **Economy hook:** uptime multiplier; feeds the outrage pressure meter when fumbled.
- **AI fallback:** solo by design; a co-op raid variant during outage world-events (whole server triages `us-east-1` together — a GW2-style event, contribution-tiered).

## 7. Terminal Typer (Tier 1, host: the garage PC)

- Type-the-command skill runs (oregon-trail-flavored shell sessions; later: incantation-length k8s commands as the difficulty curve — the YAML joke as gameplay). Session-skill clock with daily streak.
- **Economy hook:** small click-power multiplier; achievements; purely optional skill expression for active players.
- **AI fallback:** solo; async ghost-races against other players' keystroke recordings (or bot ghosts).

## 8. Demo Disc Arcade (Tier 0–1, host: the CRT) — the evolving nostalgia container

- Tiny era-authentic minigames (a Snake, a Minesweeper, a shareware-style platformer demo that ends with "ORDER NOW: $0.00"). Pure nostalgia texture with achievement hooks; the tier-0 tutorial hides here.
- **The arcade evolves per era:** cover disc (T0) → **Flash portal** (T2–3, Newgrounds/Miniclip energy) → **app store parody** (T5, everything free with IAP-screen jokes). Each stage hosts era-appropriate additions from the nostalgia sweep (`BACKLOG.md`): Solitaire/FreeCell **with the Boss Key**, "Defrag" match-3, a distinct-expression block stacker, tower defense, plinko at the casino tier, a 2048-lineage daily slot. Legal rule: mechanics yes, trade dress and names never.
- Hobby framing applies (Soul restoration) across the arcade.
- **AI fallback:** solo; daily-slot ghost scores.

## 9. Bakery, Inc. (Tier 3 easter egg, host: tenant VM)

- The game-within-the-game (`01-tiers.md Tier 3`): open a terminal on the tenant, click their cookie, buy their generators. Its production consumes your compute and pays your revenue, billed by the tick. Demand curve is literally `1.15^n`.
- **Economy hook:** an alternative active-play sink whose yield is tuned to be *almost* competitive — the joke is you keep checking on it anyway.

## 10. The Lane ("Push to Prod") — Tier 3+, host: the deploy pipeline

> Adopted by owner 2026-07-28 in place of Clash-style base raids; spec from `research/lane-pusher-design.md` (simulation-backed). The lane is a deployment pipeline: units are workloads pushed toward the rival's core switch; towers are load balancers, rate limiters, WAFs.

- **Live player, async opponent.** The simulation result is binding: auto-resolved loadout-vs-loadout collapses to a single dominant deck (Nash support 1, 9/9 configurations), so **the player plays their side in real time**; the opponent is a snapshot deck driven by a bot. Depth lives in tempo play — elixir-style trades, overloads, cycling — not in the deck.
- **One lane ships only with bypass verbs** (RFC acceptance criterion, not a nice-to-have): a buildings-only unit, a cheap cycle unit, and at least one spell. Without them: 1% decisive, 52% draws. With them one lane reaches 44% decisive; a two-lane variant (78%) is the tournament-season stretch goal.
- **Roster = the shared combat data model:** workloads typed by the **six Temperaments** on the 6-cycle chart, 1.3×/0.77× multipliers + stamina on advantage — one type chart, one stat schema, one balance harness, one sprite family shared with pet battles (§below). **Two engines** (turn-based duel vs continuous lane are different physics); **int32 only in both arenas** — replays bit-exact across Go and TS.
- **The pet is the On-Call Leader** — never a unit (a lane grinds bodies; the pet layer's tone guard holds). One Trust-scaled leader ability per match; Soul drain degrades obedience visibly.
- **Matchmaking payouts carry the punch-down multiplier** salvaged from the rejected raid design: **200% for punching up, 5% four tiers down** — the cheapest anti-bully mechanism found in any researched game.
- **Economy hooks:** wins pay Clout + tier-scaled cash; seasonal ladder titles; deck slots unlock by play, never by purchase (there is no purchase).
- **AI fallback:** the 40-line greedy policy family (beats random 83%, ~13,000× real time on the server) with published behaviour-flag difficulty manifests — transparently weaker bots, never secretly throttled ones.

## 10b. Pet battles

Specified in `04-pets.md §2` (turn-based battles vs AI + async PvP vs snapshot pets; shared combat data model with the Lane — `rfc/combat-data-model.md`). Listed here for clock coverage: turn-based match clock.

---

## Unlock stagger (fixing the sag)

| Tier | Unlocks |
|---|---|
| 0 | Demo Disc Arcade |
| 1 | Terminal Typer |
| 2 | Server Garden; Board-Game Suite (first games); additional pet variants adoptable (the companion itself arrives at Tier 0 — `04 §1`, `13 §2`) |
| 3 | Advisory Board; Ship-It Spellbook; Incident Response; Bakery, Inc. |
| 4 | The Market; the Lane (push to prod, §10); house decor expansion |
| 5 | Casino suite (blackjack/poker); pet battles ranked season |
| 6+ | Policy bench (automation-of-minigames as content: writing bots for your own hobbies — with the Soul question attached) |

Every tier from 0–6 introduces at least one new system; nothing dumps all at once.

## Consistency checklist (verification)

## 12a. Dailies doctrine (adopted 2026-08-01, `research/neopets-systems.md §6`)

Neopets proves the daily loop retains for decades AND that its cost is chore-ification plus
streak-FOMO (their 2015 streak machine sells insurance against its own loss-aversion design —
our satire target, not our template). Ours: a **bounded 5-minute "morning standup"** that
VISIBLY clocks out ("that's everything — go build"), streaks that count UP and never reset
(already law, `02 §10`), cooldowns fixed to calendar days (never drifting odd-hour timers), the
checklist shipped IN-GAME (their community needed third-party chore boards — ours is the quest
log), and dailies deliberately feeding every subsystem (pet blessing, market pull, minigame
token) because cross-linking is why the loop touches the whole game. Parody content: a streak
machine that pays identically with or without the streak, with increasingly desperate
insurance-upsell copy.

## 12b. Tier→minigame scaling (adopted 2026-08-01)

> The inverse hook: minigames already pay OUT tier-scaled rewards; this section is how tier
> progress and building counts flow IN. It is the minigame half of `02 §11b`'s role-
> differentiation law — a tier whose base production has faded stays alive by powering minigame
> surfaces, and the relevance harness measures that as a role activation.

**The seam contract (generalizing combat C5):** every minigame declares closed, named
`scaling_inputs` in its catalog entry — integer functions over (company state, founder state) —
computed server-side at match/session creation and frozen into the match snapshot. Nothing inside
a minigame reads live company state; the seam is always a handful of integers, replay-safe by the
RA rule (server-resolved inputs). Precedents: combat's `(trust_ppm, soul)`; Velocity's
engineer-scaled pool max (§4, retroactively an instance of this contract).

**The Fairness Law (extends the 2026-07-28 hardcap decision from pets to ALL ranked play):**
tier/building counts may scale PvE difficulty, rewards, resource pools, and BREADTH — **never
power in ranked PvP.** Ranked arenas keep hardcapped identical ceilings and real ratings; tier
scaling enters ranked play only as breadth (roster/deck/species options unlocked, every option
inside the same ceiling — the deck-slots-by-play pattern). A scaling input that touches a ranked
power stat is a loader error.

**Scaling axes (closed vocabulary, grows by RFC):** `reward` (exists) · `resource_pool` (Velocity
pattern) · `challenge` (PvE difficulty/boss tiers unlocked by counts) · `breadth` (options, not
power) · `era` (which content set is active — see below) · **`offline_quality` (adopted
2026-08-01 from Puzzle Pirates' labor grades, `research/kol-puzzle-pirates.md §B1/B3`): recent
minigame PERFORMANCE charges the quality tier of related automated output while away — their
booched→incredible ladder mapped puzzle skill to basic/skilled/expert offline labor, the
cleanest active/idle bridge ever shipped.** Our version: a 6-word in-fiction ladder ("SLA
Breached → … → Five Nines"), newbie-masked at the bottom (never show a new player the failure
word), decaying over ~31 days of absence exactly like theirs — skill buys quality, absence
degrades gracefully, nothing is ever zeroed (no-FOMO law intact). Two flagged tensions RULED:
(1) PP's score formula was deliberately opaque and community-spaded — ours is PUBLISHED (the
Helldivers transparency law stands; community science can happen on top of published formulas,
as our formula-view artifacts already prove); (2) PP sold the right to work (labor badges) —
NEVER here (free-forever wins; the mechanic survives only as satire copy, "Contractor
Onboarding Fee: $0.00").

### The second lane: "Shipping Wars" (Age of War pattern) — Tier 2+, host: the roadmap board

The lane ENGINE already ships two games for the price of one. Beside ranked Push to Prod (§10,
fairness-law-clean, snapshot PvP), a **solo/PvE lane variant where your company IS the age**:

- **Era = your tier.** Your unit roster is your tier era's content set (T1 garage: interns,
  beige-box servers, a LAN cable · T3 cloud: containers, autoscalers, a YAML golem · T5:
  ad-bots, dark-pattern popups · T6+: agents, aligned and otherwise). Age of War's mid-match
  age-advance becomes: your LIVE progression selects the era; prestige resets it — a run's Shipping
  Wars campaign replays the tier ladder as a combat arc.
- **Old tiers power their era's units:** scaling inputs map each tier's building counts to its
  era-unit stats/spawn rates (integer ppm tables, catalog). Buying T1 buildings at Tier 6 makes
  your T1-era units better — a measurable minigame job for every faded tier, and the harness's
  role floor sees it.
- **Opponent = bot difficulty manifests** (existing combat-bots contract); campaign ladder =
  challenge-axis scaling; rewards pay the standard tier-scaled Clout/cash OUT hook. Unranked by
  definition — this is where tier power is ALLOWED to flex, which is exactly why it can exist
  beside the ranked lane without touching the Fairness Law.
- Engine cost: near zero — same pure lane function, era decks are catalog decks selected by tier,
  counts arrive as C5-pattern scaling inputs. Content cost: era unit sets (per-tier art/copy,
  the satire writes itself).

**Quantify-first stance (restating `02 §11b`):** none of this FORCES balance — the relevance
harness quantifies usefulness (LOO deltas, role activations, tier shares) and the gate flags
deadness with evidence; floors are tweaked from epoch data, and deliberate exceptions are
declared (`trap_exempt`), never silent.

Every minigame above states: clock type ✓, economy hook ✓, AI fallback or solo-by-design ✓, persistence across Exits ✓ (all persist; the Market explicitly so; pet/base live on the founder layer), unlock tier ✓.
