# MMO Systems

> Four models, all in: global shared milestones, visible presence, co-ops/guilds, direct interaction — plus the commons that makes Ethical% viable. **Every feature lists its AI/solo fallback**; the game must be complete with zero other players online. Server-authoritative throughout (`06-tech.md §anti-cheat`), and **formulas are published** — the Helldivers transparency lesson is a design law here.

## 1. Global community milestones

**The model: Elite Dangerous dual reward axis.**

- A milestone is a **tiered global bar** ("the community provisions 10^X exaFLOPs", "collectively resolve N incidents during the outage"). Everyone's qualifying output contributes automatically — no accept/turn-in (GW2 rule).
- **Axis 1 — global tier unlock:** the tier reached determines the world reward — a new tier of the ladder (Tiers 4 and 5 were milestone-gated at launch), a new minigame, a permanent world-state change (EVE's Pochven precedent: some changes are forever). This **is our content cadence**: shipping is diegetic.
- **Axis 2 — personal rank payout:** your **Influence** payout scales with your contribution *percentile* (not absolute output — idle games span orders of magnitude), in GW2-style medal tiers: Gold (top 10%) 100%, Silver (top 40%) 85%, Bronze (participated) 75%; roughly half on failure. Only success grants the permanent unlock.
- **Contribution formula (published in-game):**
  `contribution = qualifying_output × impactModifier(active_population) × perFactionPercentile`
  The **impact modifier** (Helldivers' Galactic Impact Modifier) normalizes by the share of the playerbase currently active, so a milestone paces the same at 200 CCU or 20,000. Displayed live, with the current decay/regen dials.
- **Throughput discipline (the Clash of Clans lesson):** thresholds set from live telemetry with uncertainty margins; final tiers designed to be genuinely uncertain; a GM can tune dials mid-event — and every GM intervention is logged to a public **war log** (opacity → meta-narrative).
- **Per-faction asks:** milestones request different things from different factions (Enterprise's banked compute, VC burst windows, Open Source pool contributions, Bootstrapper steady output), measured per-faction-percentile so no faction outclicks another on a shared number.
- **Failure is allowed and canonical.** A failed milestone becomes lore (dispatches narrate it) and re-routes the season (`09-events.md`).
- **AI fallback:** at very low population the impact modifier scales until a lone player's contribution is meaningful; NPC "partner companies" visibly contribute the fictional remainder so bars always move (disclosed in the formula panel — the NPCs are labeled).

## 2. Presence & feed

- **The live feed** (cattery's activity feed, grown up): a sidebar stream of real player actions — Exits ("<name> sold their company for $4.2B"), world-firsts, golden-opportunity chains, big daemon pops, pet adoptions, lane results. Aggregate-then-broadcast (one coalesced snapshot at 4–10 Hz; never per-click fan-out).
- **Global counters:** players online, total production this hour, Planet %, current milestone bars — ambient MMO texture on every screen.
- **Ghosts:** anonymized presence in shared spaces (the world map shows clusters of activity; the leaderboard page shows live viewers).
- **Feed prominence** scales with Clout (the influence axis made visible).
- **AI fallback:** with no players online the feed runs on NPC companies (clearly styled as NPCs) so the world never reads dead.

## 3. Guilds (co-ops)

- **Formation:** 2–50 players; open/invite/apply; guild page with shared feed and pet-visit yard.
- **Automatic tithe (Idle Clans model):** a fixed small % of every member's output feeds guild XP passively — zero coordination required, every member visibly contributes, killing leech resentment structurally. Displayed per-member as percentile-within-faction (Egg Inc grading), not raw numbers.
- **Guild levels → leader-purchased upgrades** benefiting all members (production %, offline cap +, extra swap tokens, lane deck cosmetics).
- **Contribution windows for guild events (Clicker Heroes model):** guild bosses/objectives cap any member's contribution per hour (e.g. one 30-s "surge" per hour) — headcount and diversity beat whales.
- **Faction interdependence (the Last Meadow rule):** each faction produces a resource **its own faction cannot consume** — Bootstrappers→Revenue→(VCs need it), VC→Hype→(Open Source needs it), Open Source→Libraries→(Enterprise needs it), Enterprise→Compliance→(Bootstrappers need it). A cycle, not a hierarchy: guilds genuinely want one of each, and recruiting a missing faction is a guild goal. Exchange is automatic within a guild (no trading UI needed) with a public exchange for guildless players.
- **Micro-reciprocity (the warm channel — `research/liveservice-idle-tier.md §5`):** help taps (one tap shaves a visible sliver off a guildmate's timer, attributed), directed donations with names on them, and **one scheduled 30-minute guild co-presence ritual per week** (booked, not ambient — the calendar creates the appointment, the tithe does the economics). The tithe kills leech resentment; this is what makes a guild feel *inhabited*.
- **Small-guild mercy scaling keys on the cohort object (§5), not only guild size** — a 3-player guild in a healthy cohort is not the same as a 3-player guild alone at 4 a.m.
- **Matchmaking brackets (Egg Inc v2):** guild-event matchmaking pools by measured 30-day activity grade (C→AAA), so casual guilds compete with casual guilds. Self-segregation by commitment is a supported feature, not an accident.
- **AI fallback:** solo players get an **NPC partner network** — a lightweight pseudo-guild of AI companies providing the faction-exchange cycle and a private tithe pool at reduced efficiency. All guild content is completable solo via NPCs (slower, clearly labeled).

## 4. Direct interaction / PvP

| Feature | Human version | AI fallback |
|---|---|---|
| Board-game matches | Ranked queue, expanding Elo bands, spectators, guild tournaments | Bot backfill after 10–30 s at matched rating; disclosed; reduced/non-ranked rewards; no mid-match hot-swap |
| Pet battles | Async vs snapshot pets (owner absent); live ranked in seasons | NPC trainers + bot policies at every tier |
| The Lane (push to prod) | Live player vs async snapshot decks (`03 §10`) | Bot-driven decks at published manifest levels; always available |
| The Market | Server-global prices moved lightly by aggregate trading | Simulated market makers |
| Trading/gifting | Cosmetic + seed gifting (no power trading — anti-RMT by construction). **Crate-derived items carry a non-transferable flag** (decided 2026-07-28 per `research/compliance-2026-refresh.md`: transferability is the hinge every adverse loot-box ruling turned on — zero price kills the *stake*, non-transferability kills the *prize*); hand-authored cosmetics and seeds stay giftable. | NPC gift events |
| Sabotage-flavored PvP | **Deliberately soft**: "FUD campaigns" nudge a rival's outrage meter cosmetically during tournament seasons only; no destructive PvP on the main economy | n/a |
| Human-human only | Live tournament finals, guild lane leagues | Explicitly the one category allowed to require humans (per the design rule: human-only only where AI is more hassle than value) |

**PvP philosophy:** the main economy is PvE/coop; competitive expression lives in minigames, races (leaderboards, speedrun categories), and seasonal tournaments. Nothing another player does can meaningfully damage your progress — the MMO's teeth are in shared stakes (milestones can fail), not player-vs-player harm.

## 5. The commons (Ostrom's engine — Ethical%'s backbone)

> Synthesized 2026-07-28 from `research/commons-game-theory.md` (the spec) and `research/morality-systems.md` (the corrections). The earlier route-flag version of this section is superseded: **membership is open to all, and defection is *derived*, not declared.**

- **Two axes, published:** **Health** = the *weighted mean* of member compliance (population-invariant by construction — drives the buff rate) · **Capacity** = the *absolute sum* of tithes (drives caps and content gates). The Helldivers impact modifier already normalizes contribution; Health is what prices *selfishness*, which Helldivers never had.
- **Defection is read off the production stack** (the Enclosure index): the same multiplier terms that pay you mark you. No reports, no votes, no moderation queue — Ostrom's monitoring principle for free, and it can't be gamed by claiming a route you don't run.
- **The buff:** `1 + 5·[0.6·f(H) + 0.4·sᵢ]`, `f(H) = ((H−0.35)/0.65)^1.5`, clamped ≥ 0. **40% is personal Solidarity (sᵢ) nobody else can touch** — a total server collapse costs a loyal member ~47%, never everything. Fragility lives in the payoff curve, not the clock: collapse is convex, recovery is forgiving, and the meter is always visible (the TPP lesson: trap people with no exit and you get organised sabotage).
- **Nested Health — and this is where the *cohort* lives:** `H = 0.5·H_guild + 0.3·H_cohort + 0.2·H_server`. A **cohort** is a **server-assigned, non-elective, persistent group of ~150 players** (Dunbar-scaled). It exists because at 20,000 players server-mean defection is free and the dilemma evaporates; the cohort restores a shadow of the future (you meet these people again), defeats alt-stacking (non-elective), and gives dispatches a cast of recurring named neighbors. *The cohort is a first-class object: leaderboard slices, dispatch targeting, and mercy scaling all key on it.*
- **Governance (Ostrom, published):** cohorts vote the tithe dial inside a server band (poll direction, not implementation — the OSRS rule); sanctions are graduated (buff-share decay before exclusion, never bans); standing is **current state only, never a permanent badge** (The Button's flair-caste lesson: a public score becomes a caste and turns toxic).
- **Solo viability:** solo Ethical% targets a **1.8–2.5×** compensating grammar (the morality dossier's correction — at 7–10 months nobody travels far enough to discover the commons is the answer, so the solo floor must hold without it); the commons lifts that toward parity with canon, it is never the sole gate.
- **World-firsts:** unchanged — first Ethical% completion is a broadcast, permanent, dated event. Cheese counts, as the `Glitched` *variable* (per `research/speedrun-governance.md`: it's a variable, not a category).
- **AI fallback:** NPC co-ops hold Health near neutral at low population (labeled in the formula panel); thresholds scale with the impact modifier.

## 6. Leaderboards & verification

- Boards per: speedrun category (per-tier splits, Sum of Best), playstyle (separate boards per faction/build — the Realm Grinder lesson: never one board for all verbs), minigame ratings, guild seasons, world-event contributions, collections.
- **Verification is replaying the intent log** (`research/speedrun-governance.md`): no video, no human queue — the server re-simulates every submitted run; four machine-generated rejection causes; the validator ships to players; **boards are derived, never authored.** Boards key on **`epoch_id`** (a run is pinned to the epoch live at its timer start, for its whole duration), with `constants_hash` as the within-epoch invariant.
- **Categories:** 4 canonical (terminal conditions in code: `Any%`, `100%`, `Ethical%`, `Low%`) + a player-authored predicate surface the server enforces automatically (our `Legal%`/`Net Zero%`/`Glitchless`/`Pacifist` seeds live here, under a house account) + an Exhibition tier. Promotion is threshold-based, never curated.
- **The Route Registry** (`08 §6`): undocumented state-transition sequences are detected server-side; **the first player to execute one names it, permanently** — a dated public ledger of every named route, its first executor, and its adoption curve.
- Speedrun culture *chrome* kept as fiction where it's honest: `[Retimed]` tags, an Attended-Time toggle (our load-removed analogue), and one permanently "disputed" WR thread about **what counts as attended time** (the honest version of the joke — closed-form production makes IGT exact, so the clock dispute would be fake).
- **TAS board:** the AGI's runs live on a separate machine board (never the human one).
- Anti-cheat: boards derive from server-observed events only; flagged saves reviewed, not auto-banned (`06-tech.md §anti-cheat`).

## 7. Anti-leech / anti-whale summary (the four tools, all used)

1. Percentile contribution within faction (Egg Inc grading).
2. Automatic tithe (Idle Clans) — passive players still contribute visibly.
3. Per-hour contribution windows on guild events (Clicker Heroes).
4. Dual-axis rewards (Elite Dangerous) — rank pays the individual, tier pays everyone.

## 8. Social platform embedding (later)

The Discord Embedded App SDK path (`research/idle-landscape.md`, Discord Embedded App SDK) is kept open as a post-launch channel: the game is a web app; an embedded guild-scoped view (guild dashboard + feed + quick actions) inside Discord is feasible without a separate codebase. **Hard rule from the Last Meadow backlash: nothing ambient/persistent in anyone's client chrome** — the embed is opened deliberately or not at all.
