# Cloud Clicker — Vision

> **Working title:** Cloud Clicker. **One-line pitch:** *Speedrun any% destroy humanity.*
>
> A free, browser-based MMO idle game in which you climb from a 1995 garage to a world-consuming AI megacorp — a satire of 25 years of tech, 25 years of gaming, and 70 years of everything else, that plays it completely straight.

## The elevator pitch

You start as a sole proprietor in the last good year of the internet. You own your tools, your customers like you, cheat codes work. Nine tiers later you are restarting nuclear plants to feed token factories, your pet no longer recognizes you, 4% of your users believe you are a lizard (the board calls this "within tolerance"), and a speedrun timer in the corner has been counting the whole time. The game is Cookie Clicker's spiritual successor with the two things Cookie Clicker never had: **other people** and **an ending**.

## Why this game

- **The genre gap is real.** The "garage → singularity" theme is crowded with shallow mobile tycoons and served earnestly by Cell to Singularity, but the **deep / satirical / MMO** corner is completely empty (see `research/idle-landscape.md §6`). Cookie Clicker's own community names the missing pieces: an endgame with a shape, minigames for every building, meaningful build divergence, faster content cadence. We are building exactly those.
- **The MMO solves the content-cadence problem.** Cookie Clicker shipped one major system every 2–3 years and survived on being first. Our content cadence *is a game mechanic*: community milestones and seasonal story arcs unlock new tiers, minigames, and events for everyone — shipping is diegetic.
- **The satire has teeth because the game is free.** AdVenture Capitalist's critical failure was selling IAPs inside a capitalism satire — "you feel like part of the joke, rather than in on it." Orteil's rule ("idle + microtransactions is a questionable, if not exploitative combination") is our license: **no real money, ever, in any direction.** Free-ness is not a business decision; it is the game's central satirical asset and it is stated in-game, repeatedly, in the corporate voice of the systems it mocks.

## The satire thesis

1. **Play it straight (the Paperclips rule).** No winking narrator. The player *becomes* the misaligned optimizer; the game never announces the moment you became the villain — you just notice the Trust stat stopped mattering.
2. **Reproduce every dark pattern faithfully, price it at zero, set disclosure to maximum (the Almanac rule).** Loot boxes with 100.00% drop rates ("odds sum to 4,700.00%. This is normal"), a battle pass with two identical free tracks whose timer counts *up*, an energy bar that refills faster than you can spend it. The joke is the unbearable sincerity of a system that gives you everything and still performs scarcity.
3. **The game enshittifies itself as you climb (the Doctorow rule).** The act structure is literally his three stages — Act I "Good To Your Users", Act II "Good To Your Business Customers", Act III "Claw Back All The Value" — and the prestige screen is titled **"Then, They Die."** Presentation decays with you: the 1995 tier is honest shareware UI; the megacorp tier is a parody of modern dark-pattern chrome that the game visibly inflicts on *itself*.
4. **Mock the machinery, never the joy (the Bogost rule).** Cow Clicker's lesson: satire of a compulsion loop that *is* a compulsion loop gets played sincerely. So make the sincere version the good one. Hats are genuinely great. Gacha art is genuinely gorgeous. The pet genuinely loves you. The satire lives in tooltips, disclosures, achievement names, and the corporate voice — never in making the game worse to play.
5. **The melancholy is real data (the Both-Graphs rule).** The background radiation is the post-WW2 decay arc — 1971 wage decoupling, trust **77% (1964) → 17% (2025)** `[V]` ([Pew](https://www.pewresearch.org/politics/2024/06/24/public-trust-in-government-1958-2024/)), Bowling Alone — and the counter-thread is equally real over the same post-war span: global under-5 mortality **22.3% (1955) → 3.74% (2024)** `[V]` ([OWID](https://ourworldindata.org/grapher/child-mortality), Gapminder + UN IGME) and global extreme poverty **49.1% (1955) → 10.0% (2018)** `[V]` ([OWID, Moatsos 2021 cost-of-basic-needs](https://ourworldindata.org/grapher/share-of-population-living-in-extreme-poverty-cost-of-basic-needs)). Neither cancels the other. The ending shows both graphs and asks "Play again?"

   > **Resolved (2026-07-28).** The figures previously cited here (*"child mortality 50%→4.3%, extreme poverty 75%→10%"*) were 1800/1820 baselines presented as ~70-year records; they were withheld 2026-07-27 and are now replaced with verified ~1955-baseline OWID series, baseline years stated inline above. One caveat survives: **the World Bank's own poverty series cannot be anchored at 1955** — its aggregate starts decades later (at the $3.00/day 2021-PPP line: **44% (1990) → 10% (2025)** `[V]` [OWID](https://ourworldindata.org/extreme-poverty)); the 1955 anchor is the Moatsos (2021) academic reconstruction, which ends in 2018 — if the poverty chart ships, name the source on-screen. The under-5 mortality series needs no caveat and is the recommended improvement series for Ending C. Full sourcing and candidate pairings: `research/billionaires-decay.md §2F`.

## Design pillars

1. **Idle and active are different builds, not different speeds.** Cookie Clicker's active meta (multiplicative buff windows, 10⁵–10⁷× combos) and its idle meta (wrinklers) are both first-class here — via factions whose *verbs* differ (see `10-playstyles.md`). Offline progress is default and generous, never a purchased privilege.
2. **MMO everywhere, AI fallback everywhere.** Global milestones, presence, guilds, PvP — and every multiplayer feature has a bot/AI substitute so the game is complete with zero other players online. Bots never cheat and are labeled honestly.
3. **A designed ending.** The genre's defining failure is that players lapse instead of finishing. This game ends — with a choice, three endings, and a leaderboard. Prestige loops live *inside* the arc ("New Route"), not instead of one.
4. **Each tier is a paradigm shift.** Per Paperclips/Antimatter Dimensions: a new tier is a new mechanic grammar and a new decision type, and it automates the tier below. Never the same shop with more zeroes.
5. **Everything on different clocks.** Cookie Clicker's deepest lesson: garden on wall-clock, pantheon on swap budgets, market on CpS-denominated dollars. Our systems deliberately occupy different time signatures — including one hard real-time currency no production rate can accelerate (Fiscal Quarters).
6. **Free forever, honest forever.** No microtransactions, no ads, no FOMO with real costs. Parodied dark patterns are always curtain-pulled. Data collection is minimal and disclosed (we are not building the thing we mock).
7. **Server-authoritative, transparently.** Production is computed server-side from closed forms; community-event formulas are *published* (the Helldivers lesson: opacity, not mechanics, caused the backlash).

## What we steal, from whom

| Source | The steal |
|---|---|
| Cookie Clicker | Achievements-as-currency, the Lucky bank formula, buff multiplicativity, sugar-lump-style real-time gate, wrinkler-style disguised buffs, one-toggle-two-games |
| Universal Paperclips | Paradigm shift per act; the silent moral escalation; an actual ending |
| Antimatter Dimensions | Layered prestige where each layer automates the one below; challenges as rule-modified runs |
| Realm Grinder | Factions as verbs; balance per engagement pattern, not per output axis |
| Last Meadow Online | Profession interdependence: you cannot produce what your own class consumes |
| Elite Dangerous | Dual reward axis on community goals (personal rank + global tier unlock) |
| Helldivers 2 | Server-orchestrated war with a GM, dispatches, canonical failure — plus the transparency lesson |
| EU4 / CK3 | Data-driven event chains, disaster progress bars, situations with phases and catalysts |
| Progress Knight | The founder ages; a lifetime is the run unit |
| Undertale | The canon route is the dark one; Ethical% is the brutal, true-pacifist path |
| GW2 | Contribution without accept/turn-in; medal tiers; hearts as the content floor |
| cattery (our own) | Pet stat/decay/trust systems, behavior state machines, CSS-only sprites |

## Anti-goals (the trap list)

- No IAPs, no ads, no NFTs — not even ironically functional ones.
- No anti-idle hazards (The Last Meadow's sin) — inattention is a build, not a failure. **Refined 2026-08-05: the constraint is on *consequences*, not on interaction. What is barred is a net *loss* inflicted on a player for being offline — not offline interaction itself. An asymmetric mechanic where an active player acts against your absent snapshot is permitted as long as your side loses nothing: loss-decoupled defense, the Clash "Clash Anytime" clone-base model (attacker earns rewards from a clone; the defender loses no resources/standing). This is why offline raiding was rejected in its naive form (real defender loss) but a loss-decoupled form is on the deferred list, not banned.** "No FOMO with real costs" (§pillar 6) is thus a balance target — zero net cost to the absent player — not a prohibition on the mechanic.
- No wiki-dependence by design (IdleOn's sin) — if optimal play needs an external tool, build the tool into the game (the Cookie Clicker FtHoF-predictor lesson).
- No softcaps — hardcaps with visible numbers (Paper Pilot).
- No global milestones without throughput telemetry (Clash of Clans burned all five tiers in 48 hours).
- No first prestige that doesn't obviously pay off within one session (Unnamed Space Idle's churn wall).
- No desktop-only: browser-first, mobile-usable.
- No unearned edge: bots play by the same server-validated rules as humans.

## Document map

| Doc | Contents |
|---|---|
| `01-tiers.md` | The nine-tier ladder, era mirroring, endings |
| `02-economy-balancing.md` | Currency/layer architecture, core math, pacing |
| `03-minigames.md` | Minigame catalog: clocks, economy hooks, AI fallbacks |
| `04-pets.md` | Tamagotchi layer, battles, cosmetic house, cosmetics/lootboxes |
| `05-mmo.md` | Milestones, presence, guilds, PvP, anti-leech systems |
| `06-tech.md` | Stack, architecture, anti-cheat |
| `07-roadmap.md` | Build order and release-as-fiction schedule |
| `08-satire-flavor.md` | The flavor bible: eras, content banks, voice, conspiracy layer, canonization |
| `09-events.md` | Three-layer event engine, seasonal arcs, GM ops |
| `10-playstyles.md` | Factions, challenge runs, moral routes, time banking |
| `11-ux-writing` | Surfaces, FTUE, run-end, copy production |
| `12-content-pipeline` | Packs, extension points, effects vocabulary |
| `13-world` | Diorama, region draft, shared map, Planet ratchet |
| `BACKLOG.md` | The topic ledger — every idea gets a row |
| `research/` | The research corpus (coverage matrix: `research/README.md`) |
