# MMO Systems

> Four models, all in: global shared milestones, visible presence, co-ops/guilds, direct interaction — plus the commons that makes Ethical% viable. **Every feature lists its AI/solo fallback**; the game must be complete with zero other players online. Server-authoritative throughout (`06-tech.md §1.8`), and **formulas are published** — the Helldivers transparency lesson is a design law here.

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

- **The live feed** (cattery's activity feed, grown up): a sidebar stream of real player actions — Exits ("<name> sold their company for $4.2B"), world-firsts, golden-opportunity chains, big daemon pops, pet adoptions, raid outcomes. Aggregate-then-broadcast (one coalesced snapshot at 4–10 Hz; never per-click fan-out).
- **Global counters:** players online, total production this hour, Planet %, current milestone bars — ambient MMO texture on every screen.
- **Ghosts:** anonymized presence in shared spaces (the world map shows clusters of activity; the leaderboard page shows live viewers).
- **Feed prominence** scales with Clout (the influence axis made visible).
- **AI fallback:** with no players online the feed runs on NPC companies (clearly styled as NPCs) so the world never reads dead.

## 3. Guilds (co-ops)

- **Formation:** 2–50 players; open/invite/apply; guild page with shared feed and pet-visit yard.
- **Automatic tithe (Idle Clans model):** a fixed small % of every member's output feeds guild XP passively — zero coordination required, every member visibly contributes, killing leech resentment structurally. Displayed per-member as percentile-within-faction (Egg Inc grading), not raw numbers.
- **Guild levels → leader-purchased upgrades** benefiting all members (production %, offline cap +, extra swap tokens, raid slots).
- **Contribution windows for guild events (Clicker Heroes model):** guild bosses/objectives cap any member's contribution per hour (e.g. one 30-s "surge" per hour) — headcount and diversity beat whales.
- **Faction interdependence (the Last Meadow rule):** each faction produces a resource **its own faction cannot consume** — Bootstrappers→Revenue→(VCs need it), VC→Hype→(Open Source needs it), Open Source→Libraries→(Enterprise needs it), Enterprise→Compliance→(Bootstrappers need it). A cycle, not a hierarchy: guilds genuinely want one of each, and recruiting a missing faction is a guild goal. Exchange is automatic within a guild (no trading UI needed) with a public exchange for guildless players.
- **Matchmaking brackets (Egg Inc v2):** guild-event matchmaking pools by measured 30-day activity grade (C→AAA), so casual guilds compete with casual guilds. Self-segregation by commitment is a supported feature, not an accident.
- **AI fallback:** solo players get an **NPC partner network** — a lightweight pseudo-guild of AI companies providing the faction-exchange cycle and a private tithe pool at reduced efficiency. All guild content is completable solo via NPCs (slower, clearly labeled).

## 4. Direct interaction / PvP

| Feature | Human version | AI fallback |
|---|---|---|
| Board-game matches | Ranked queue, expanding Elo bands, spectators, guild tournaments | Bot backfill after 10–30 s at matched rating; disclosed; reduced/non-ranked rewards; no mid-match hot-swap |
| Pet battles | Async vs snapshot pets (owner absent); live ranked in seasons | NPC trainers + bot policies at every tier |
| Base raids | Async vs layout snapshots | Procedural compounds; attacking armies always AI |
| The Market | Server-global prices moved lightly by aggregate trading | Simulated market makers |
| Trading/gifting | Cosmetic + seed gifting (no power trading — anti-RMT by construction) | NPC gift events |
| Sabotage-flavored PvP | **Deliberately soft**: "FUD campaigns" nudge a rival's outrage meter cosmetically during tournament seasons only; no destructive PvP on the main economy | n/a |
| Human-human only | Live tournament finals, guild-vs-guild raid leagues | Explicitly the one category allowed to require humans (per the design rule: human-only only where AI is more hassle than value) |

**PvP philosophy:** the main economy is PvE/coop; competitive expression lives in minigames, races (leaderboards, speedrun categories), and seasonal tournaments. Nothing another player does can meaningfully damage your progress — the MMO's teeth are in shared stakes (milestones can fail), not player-vs-player harm.

## 5. The commons (prisoner's dilemma — Ethical%'s engine)

- **The Mutual Aid Pool:** ethical-route players (any faction, flagged by route) may tithe into a server-wide commons that pays a **Trust-scaled production buff to all commons members** — the buff strength scales with total pool size and **degrades with the defection rate** (members who take the buff while running dark-pattern multipliers are defectors; the system detects it from route flags).
- Classic PD payoffs, live: defectors profit individually and weaken the commons; if defection stays under a threshold, the commons holds and Ethical% is genuinely viable; a defection cascade collapses the buff for everyone (with a dispatches post-mortem — tragedy of the commons as observable server history).
- **World-firsts:** the first player/guild to complete Ethical% is a broadcast, permanent, dated event (WoW race-to-world-first energy). Cheese routes count and are celebrated by category (`Ethical% (Glitched)` — "the only known route to being ethical is an exploit" is the intended joke).
- **AI fallback:** at low population, NPC co-ops participate in the commons at neutral rates so the pool exists; thresholds scale with the impact modifier.

## 6. Leaderboards & verification

- Boards per: speedrun category (per-tier splits, Sum of Best), playstyle (separate boards per faction/build — the Realm Grinder lesson: never one board for all verbs), minigame ratings, guild seasons, world-event contributions, collections.
- Speedrun culture cosplay, faithfully: verification queue, `[Retimed]` tags, a Load-Removed toggle, one permanently "disputed" WR with an in-fiction comment thread about whether externalities count toward IGT.
- **TAS board:** the AGI's runs live on a separate machine board (never the human one).
- Anti-cheat: boards derive from server-observed events only; flagged saves reviewed, not auto-banned (`06-tech.md §1.8`).

## 7. Anti-leech / anti-whale summary (the four tools, all used)

1. Percentile contribution within faction (Egg Inc grading).
2. Automatic tithe (Idle Clans) — passive players still contribute visibly.
3. Per-hour contribution windows on guild events (Clicker Heroes).
4. Dual-axis rewards (Elite Dangerous) — rank pays the individual, tier pays everyone.

## 8. Social platform embedding (later)

The Discord Embedded App SDK path (research §1d) is kept open as a post-launch channel: the game is a web app; an embedded guild-scoped view (guild dashboard + feed + quick actions) inside Discord is feasible without a separate codebase. **Hard rule from the Last Meadow backlash: nothing ambient/persistent in anyone's client chrome** — the embed is opened deliberately or not at all.
