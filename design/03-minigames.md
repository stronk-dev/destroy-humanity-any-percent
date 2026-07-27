# Minigames Catalog

> Rules for every minigame: (1) it runs on a **distinct clock** (the Cookie Clicker lesson — garden/pantheon/grimoire/market each had a different time signature); (2) it **hooks into the main economy** in a stated way; (3) it has an **AI fallback** so it's fully playable with zero other players; (4) unlocks are **staggered across the tier arc** (fixing CC's all-at-once dump and the hour-5–40 sag); (5) state **persists across Exits** (fixing the Stock Market resentment) unless reset is explicitly the point.

Minigames unlock via Investor Confidence (Fiscal Quarters) at their host tier — the sugar-lump unlock pattern.

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
- **Sacrifice ritual:** delete the full collection for a permanent bonus + Confidence (CC's Seedless-to-nay).
- **AI fallback:** none needed (solo). **MMO hook:** rare strains are tradeable gifts; a community seed-census feeds a collection milestone.

## 2. Advisory Board (Tier 3, host: HQ)

The pantheon equivalent — **loadout with drawbacks + swap friction**.

- Three seats: **Boardroom (diamond) > Backchannel (ruby) > Group Chat (jade)**. Slot advisor archetypes, each with an upside *and a drawback* (the CC rule: no free spirits): e.g. `The Growth Guy` (+click power, Trust drain), `The Ascetic CTO` (+idle rate, unslots if you click a golden opportunity), `The Fixer` (sell-off combo enabler — Godzamok's role), `The Oscillator` (a Cyclius-style ±15% sine on 3/12/24 h real-time cycles you can time combos to), `The Safety Hire` (−p(doom), −speed).
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

- `$1 = 1 second of your highest unbuffed rate this run` — the CC trick kept verbatim: buy low, *raise your own rate*, sell — printing money off your own growth is the intended aha.
- Goods = fictional sector indices (GPUs, RAM, Compliance, Attention, Water futures — the satire writes itself). Resting-value drift + momentum + six market modes (CC's model).
- **Loans** for combo windows (+50–100% rate, penalty after).
- **Persists across Exits** (holdings survive; the fiction: your personal brokerage, not the company's).
- **MMO hook:** prices are *server-global* and lightly influenced by aggregate player trading + world events (a Situation like "Chip Shortage" moves the GPU index for everyone). Insider-satire events fire when you trade ahead of your own announcements (Trust drain).
- **AI fallback:** market makers are simulated; no humans required.

## 5. Board-Game Suite (Tier 2+, host: break room — "hobbies")

Chess puzzles, chess, tic-tac-toe, Connect-4, Othello, Gomoku, checkers; blackjack and poker at the casino tier (Tier 5's regulator-baiting `Loot Box Casino (Free)`).

- **Framing:** these are the founder's *hobbies* — playing them restores a trickle of **Soul** (and they lock at low Soul: "you no longer remember why this was fun"). That's the retention hook *and* the theme.
- **Clock:** turn-based matches; ranked queue with cooldowns; daily chess puzzle (wall-clock).
- **Economy hooks:** wins pay Clout + small cash scaled to tier; achievements (beat the top bot, win N ranked games) feed Clout; a daily-puzzle streak feeds a small multiplier. Bot matches pay reduced, non-ranked rewards (anti-farm).
- **AI fallback (the flagship of the vs-AI system, `06-tech.md §4`):**
  - Chess: Stockfish (skill level + shallow depth + top-k softmax + fake think time); Maia later if headline. Rules server-validated via notnil/chess.
  - Connect-4/Othello/Gomoku/checkers: ~150-line alpha-beta with iterative deepening + per-game eval; difficulty = depth + temperature.
  - Blackjack: basic-strategy table bot. Poker: rule-based hand-strength + pot odds + randomized bluffs.
  - Human-feel rules: plausible blunders (captures/checks, not random), fake think time, named bots with visible fake ratings that move, resign behavior, rubber-band toward ~55–60% player win rate, **bots never cheat** (same hidden-info boundaries as humans).
- **PvP:** ranked matchmaking with expanding Elo bands, bot backfill after 10–30 s (disclosed, consistent, reduced rewards, never hot-swapped mid-match), spectating, guild tournaments. Show queue depth to encourage waiting for humans.

## 6. Incident Response (Tier 3, host: on-call rotation)

An original — **the pager as a push-your-luck reflex minigame** on a regenerating clock.

- Incidents queue on a Poisson-ish timer (97% are noise — the 3%-Signal satire). Triaging is a fast pattern-matching mini-loop (match the error signature; wordle-ish deduction against a symptom grid). Real incidents pay big uptime buffs; false-alarm grinding drains a Burnout meter that debuffs if ignored.
- **Economy hook:** uptime multiplier; feeds the outrage pressure meter when fumbled.
- **AI fallback:** solo by design; a co-op raid variant during outage world-events (whole server triages `us-east-1` together — a GW2-style event, contribution-tiered).

## 7. Terminal Typer (Tier 1, host: the garage PC)

- Type-the-command skill runs (oregon-trail-flavored shell sessions; later: incantation-length k8s commands as the difficulty curve — the YAML joke as gameplay). Session-skill clock with daily streak.
- **Economy hook:** small click-power multiplier; achievements; purely optional skill expression for active players.
- **AI fallback:** solo; async ghost-races against other players' keystroke recordings (or bot ghosts).

## 8. Demo Disc Arcade (Tier 0–1, host: the CRT)

- Tiny era-authentic minigames (a Snake, a Minesweeper, a shareware-style platformer demo that ends with "ORDER NOW: $0.00"). Pure nostalgia texture with achievement hooks; the tier-0 tutorial hides here.
- **AI fallback:** solo.

## 9. Bakery, Inc. (Tier 3 easter egg, host: tenant VM)

- The game-within-the-game (`01-tiers.md Tier 3`): open a terminal on the tenant, click their cookie, buy their generators. Its production consumes your compute and pays your revenue, billed by the tick. Demand curve is literally `1.15^n`.
- **Economy hook:** an alternative active-play sink whose yield is tuned to be *almost* competitive — the joke is you keep checking on it anyway.

## 10. Pet battles & base raids

Specified in `04-pets.md` (turn-based battles vs AI + async PvP; Clash-style base attack/defense vs AI armies and other players' layouts). Listed here for clock coverage: turn-based match + async-defense clocks.

---

## Unlock stagger (fixing the sag)

| Tier | Unlocks |
|---|---|
| 0 | Demo Disc Arcade |
| 1 | Terminal Typer |
| 2 | Server Garden; Board-Game Suite (first games); pet adoption |
| 3 | Advisory Board; Ship-It Spellbook; Incident Response; Bakery, Inc. |
| 4 | The Market; base building + raids |
| 5 | Casino suite (blackjack/poker); pet battles ranked season |
| 6+ | Policy bench (automation-of-minigames as content: writing bots for your own hobbies — with the Soul question attached) |

Every tier from 0–6 introduces at least one new system; nothing dumps all at once.

## Consistency checklist (verification)

Every minigame above states: clock type ✓, economy hook ✓, AI fallback or solo-by-design ✓, persistence across Exits ✓ (all persist; the Market explicitly so; pet/base live on the founder layer), unlock tier ✓.
