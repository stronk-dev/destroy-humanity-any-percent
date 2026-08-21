# Research Completeness Sweep — the Gap Map

> **Frozen historical gap map — not the current roadmap.** This no-web synthesis describes the
> dossier/design population visible at commit
> `3375176f1eb78b397f9efb0a80b9214ceb470094` on 2026-08-06. Later research filled several named
> gaps, including absorption, cryptocurrency and labor. Every `GAP`, blocker, priority and
> “recommended next” statement below is preserved as the critic's historical output and is
> superseded for present authorization by
> [`planning/platform-alignment/research-queue.md`](../../planning/platform-alignment/research-queue.md)
> and [`planning/platform-alignment/execution-queue.md`](../../planning/platform-alignment/execution-queue.md).
> Do not use this file to start an RFC or implementation.
>
> **What this is.** A read-only survey of the ~51 dossiers in `design/research/` and design docs
> `00`–`13`, cataloguing coverage across five axes and flagging conspicuous gaps. This is a **gap
> map, not a deep dive** — each row is one line of rationale plus why it matters to Cloud Clicker.
> Produced by a completeness-critic pass; no web fetches, so nothing here is a sourced claim (`[M]`
> throughout). It supersedes nothing; it points at where the next research commissions should go.
>
> **Headline.** The corpus is genuinely deep and unusually well-cross-referenced. The gaps are not
> in the systems already built (economy/save/production/numeric/MMO-commons are over-served) but in
> **content-feeding genre and satire research for the roadmap's content phase** — and in a few
> genres that are *on-theme for a data-center game* yet unresearched. Two gaps are already flagged
> `❌ GAP` in `research/README.md` (absorption; Brignull provenance) and are restated here in context.

---

## Axis 1 — Game genres / mechanics we could mine for minigames or systems

**Well-covered (do not re-commission):** idle/incremental (`idle-landscape`, `cookie-clicker`,
`pacing-science`, `liveservice-idle-tier`, `kol-puzzle-pirates`, `gaia-hyperinflation`,
`neopets-systems`); clicker (`cookie-clicker`); **deckbuilder-roguelike + survivor-like +
auto-battler** (`roguelike-survivor-minigames` — StS/Balatro/Brotato/Super Auto Pets, with the
Balatro-maps-onto-our-core finding); rhythm (`rhythm-timing-games`); lane-pusher (`lane-pusher-design`);
**board/abstract + full card/parlor taxonomy** (`03 §5–5b` — trick-taking, melding, shedding,
racing, mancala, dice/bluff, trivia, social-deduction, all placed); tile-placement/spatial
(`tile-placement` — Dorfromantik/Carcassonne/Opus Magnum/Loop Hero); creature-battler/pet
(`creature-battler`, `cattery-reusables`); base-building **(rejected, with analysis preserved)**;
flash-era arcade (`flash-era-arcade`); social spaces (`social-spaces`).

| Gap | Rationale | Why it matters | Priority |
|---|---|---|---|
| **Absorption / viral (agar.io, Katamari, Osmos, slither, Hole.io)** | Already `❌ GAP` in `README`; the **M&A Arena** is called the "standout" of the viral sweep and slotted **v1.0** (`BACKLOG` "Viral/absorption sweep"), yet has zero dossier. Process rule: *no RFC drafts from a GAP row.* | Blocks a v1.0 minigame RFC; "acquisition as literal consumption" is the thesis in miniature. Netcode + bot-fallback design for a real-time shared arena is non-trivial and unspecified. | **HIGH** |
| **Factory / automation (Factorio, Satisfactory, shapez, Dyson Sphere Program)** | The single most **on-theme unresearched genre** for a data-center game — belts/ratios/throughput-balancing/"build a machine that builds" is what a hyperscaler *is*. Currently only brushed: `tile-placement` name-drops Factorio once; the Server Garden and the T7–8 "allocate autonomous extractors" endgame are automation-shaped but designed without a genre study. | Feeds the **endgame grammar** (T7–8 4X consumption / Paperclips Act 3) which is under-researched as *gameplay*, and the rack/compute-allocation layer. The genre's core loop (optimize a production graph you keep rebuilding) is exactly the fantasy we're selling and we have no design reference for it. | **HIGH** |
| **4X / grand-strategy as gameplay (not just event chains)** | EU4/CK3 are cited only for **event-chain architecture** (`events-playstyles`). The *four X's*, tech trees, map-painting, snowball/victory-condition design, and 4X-AI opponents are unstudied — but `01 §T7` literally names the endgame "4X consumption endgame." | The designed ending's terminal grammar (extractor allocation, value-drift-in-your-own-swarm) is a 4X/RTS-lite loop with no research behind it; the depletion-tier splits depend on it pacing well. | **MED** |
| **Sports / franchise-management sim (Football Manager, OOTP, GM modes)** | **Zero coverage.** This is arguably the closest AAA genre-DNA to what Cloud Clicker *is*: a spreadsheet-depth people-and-numbers management sim. `liveservice-idle-tier` covers gacha *roster* meta but not the management-sim tradition (scouting, tactics-as-sliders, morale, the "manage humans as stats" satire). | The Advisory Board, hiring, faction-roster, and the whole "you manage people as numbers" satire could borrow depth here; it's also a direct satirical target (Taylorism-with-a-dashboard). | **MED-LOW** |
| **Tower-defense as a designed system (static defense, creep economy)** | TD-proper is a distinct grammar from the lane-pusher (which owns the *offense/push* side). Multiple candidates want it — Desktop Tower Defense, Missile-Command-for-Incident-Response, Bubble Tanks merge-TD — all `💡` in `BACKLOG`, none researched. `Incident Response` (`03 §6`) is TD-adjacent and un-referenced to any TD study. | Several backlog minigames converge on TD; wave/tower/pathing/DPS economy is unspecified. Lower than automation because the lane engine already covers adjacent combat. | **LOW-MED** |
| **Programming/automation puzzle micro-genre (Zachtronics: SpaceChem, Opus Magnum, Human Resource Machine, TIS-100)** | Only Opus Magnum, once, in `tile-placement`. **Human Resource Machine is dead-on-theme** (programming = corporate-labor satire; "assembly language as a soul-crushing office job"). | Cheap, deterministic (trivially server-verifiable — pure `(program, input)` function), and the most literal "your job is now a puzzle" satire available. Natural Terminal Typer / T3 engineering-org cabinet. | **LOW** |

*Intentionally out of scope (noted, not gaps):* metroidvania / exploration-gated progression (no
diorama fit); fighting-game frame-data; open-world sandbox. Skip unless a design need appears.

---

## Axis 2 — Cultural / satirical touchstones the flavor may be missing

**Well-covered:** enshittification (`gaming-enshittification`), billionaires + post-war decay
(`billionaires-decay`), conspiracy/media (`conspiracy-media`), societal satire (`societal-satire`),
attention economy + Zuboff chain (`08 §3`), AI/datacenter/scaling-laws era, six-generation culture
dossiers (`culture-pre-boomer`…`culture-genalpha`), internet-platform arc + meme canon
(`internet-culture`), media-format nostalgia (`formats-nostalgia`), regulatory capture
(`regulatory-capture`).

| Gap | Rationale | Why it matters | Priority |
|---|---|---|---|
| **Crypto / web3 depth** | Thin: crypto/web3/NFT/blockchain appear only **1–3 times each** across the three main satire files, mostly as one-off cards ("the 96%-dislike NFT cutscene", "no NFTs even ironically"). No dedicated dossier. | A **massive, on-theme satire vein** left on the table: ICOs, rug pulls, NFT mania, DAOs, exchange collapses (FTX/Celsius), pump-and-dump, and above all **crypto mining's energy draw**, which ties straight into the game's datacenter/energy/depletion thesis. This is the highest-leverage *content* gap. | **HIGH** (satire) |
| **Labor / unionization / tech-worker organizing** | Present only as scattered cards ("Union Hall — a building you demolish", Taylorism in `culture-pre-boomer`, layoffs event). No dossier. The satire is **executive/billionaire-heavy and worker-light.** | The commons/Ethical% counter-thread wants a real labor-solidarity vocabulary (Alphabet Workers Union, game-industry/QA unions, gig-economy, RTO fights, strike history). Pairs with the "melancholy is real data" pillar; balances a satire that currently punches up but rarely stands *with* the ground floor. | **MED** |
| **Open-source / hacker / demoscene culture** | "Open Source" is a **faction** (`10 §1`) and the commons is Ostrom-modeled, but open-source *culture* is unresearched: FOSS history, GPL/license wars, corporate OSS capture, maintainer burnout, Log4j/xz-backdoor, the demoscene/warez/hacker aesthetic. | Directly feeds the Open Source faction's voice and the commons satire; the demoscene/cracktro aesthetic is also free era-flavor for T0–1. | **MED** |
| **Non-Western / national tech scenes** | `enclave-atlas` covers Asian MMO giants; culture dossiers are Western + generational. China's superapp/social-credit/gacha-industrial-complex and Korea's PC-bang/RMT-origin scenes as *satire material* are thin. | Lower now (i18n is deferred), but the AI/surveillance endgame has obvious non-Western referents the corpus doesn't reach. | **LOW-MED** |

---

## Axis 3 — Idle / incremental design patterns

**Verdict: over-served, minimal gaps.** `idle-landscape` + `cookie-clicker` + `pacing-science` +
`kol-puzzle-pirates` + `liveservice-idle-tier` + `gaia-hyperinflation` + `neopets-systems` +
`adaptive-balancing` + `balance-enforcement` + `tier-relevance` collectively cover prestige layering
(AD), pacing science, faucet/sink economies, dailies, offline-quality grades, DDA, and
anti-dead-content architecture. `break_eternity`/tetration-layer games are **deliberately deferred**
by design law 3. No new commission recommended; the one thin spot (multi-layer prestige-tree /
Profectus-lineage as a distinct study) is adequately proxied by the AD coverage. **No priority.**

---

## Axis 4 — MMO / social systems

**Well-covered:** global milestones + dual-axis contribution (`05 §1`), presence/feed, guilds +
tithe + interdependence, the **commons/cohort** (Ostrom, `commons-game-theory` + `05 §5`),
player-run markets (`player-markets`, `gaia`, `neopets`), guild identity surfaces (Neopets adoptions),
governance (cohort tithe votes, graduated sanctions), moderation (largely dissolved by the
no-free-text law + `compliance` name-moderation), spectating (Break Room, leaderboard live viewers,
board-game spectate).

| Gap | Rationale | Why it matters | Priority |
|---|---|---|---|
| **Mentorship / veteran-newbie onboarding** | Not covered. No study of RuneScape mentor, FFXIV sprout/mentor-roulette, WoW Recruit-A-Friend, EVE corp onboarding. | The game has a **designed ending** and a churn-risk first-prestige (`00` anti-goals). A veteran-guided onboarding social layer is a natural retention + commons-culture-transmission system and currently has no design reference. | **MED-LOW** |
| **Streaming / spectating integration + speedrun-marathon culture** | Spectating-*within-game* is covered; **streaming-out** is not (Twitch extensions, streamer mode, race-to-WR broadcast, the GDQ/marathon angle). The whole game is framed as a *speedrun* — the most streaming-native framing there is. | `speedrun-governance` covers verification/categories (the rules) but not the *broadcast culture* the framing invites. Partly marketing-adjacent (deferred), but the streamer-mode + race-broadcast hooks are product features. | **MED-LOW** |
| Reputation / social-status beyond commons standing | The Button flair-caste lesson + Clout + current-state-only standing already cover the design risk. | Adequately handled; no commission. | LOW |

---

## Axis 5 — Monetization-satire patterns

**Verdict: deeply covered.** `gaming-enshittification` (crate/pity/50-50/kompu-gacha, the
achievement-quote museum, Horse Armor, subscription stacking, the NFT cutscene), `billionaires-decay`,
the full dark-pattern parody suite in `08 §3`, lootbox/gacha/battle-pass parody in `04`, and the
compliance corpus. Two small gaps:

| Gap | Rationale | Why it matters | Priority |
|---|---|---|---|
| **Dark-pattern taxonomy provenance (Harry Brignull, "dark pattern", 28 Jul 2010)** | Already `❌ GAP` in `README` — the coiner appears nowhere in the corpus despite the term being used throughout. | A one-page citation/lore fix, not a system; but the honesty-appendix voice (`08 §7`) wants the receipt. | LOW |
| Ad-tech mechanics (RTB / programmatic / MFA / ad-fraud) | Attention economy is covered *thematically* (Zuboff) but not as *mechanics* (real-time bidding auctions, made-for-advertising sites, the ad-fraud ecosystem). | A possible minigame/economy-satire seam (an ad-auction as a market minigame), but low-leverage vs. the crypto gap. | LOW |

---

## Historical 2026-08-06 research queue (superseded)

The priorities below explain what the frozen sweep recommended at its coordinate. They have not
been refreshed into a new roadmap; consult the current queues linked above.

1. **Absorption / viral game design** *(the M&A Arena dossier)* — **HIGH.** A named v1.0 minigame is
   sitting on a `❌ GAP` row and the process rule forbids drafting its RFC until researched. Cover
   agar.io netcode + bot-fallback, Katamari scale progression, Osmos mass-as-fuel; deliver a
   shippable spec for a real-time shared arena. *Unblocks content-phase work directly.*

2. **Factory / automation genre study** — **HIGH.** The most on-theme unresearched genre for a
   data-center game, and the **endgame grammar (T7–8 extractor allocation) currently has no gameplay
   reference.** Factorio/Satisfactory/shapez/DSP: production-graph optimization, ratio/throughput
   design, and how the "build a machine that builds" loop paces. Feeds the ending *and* the
   rack/compute layer.

3. **Crypto / web3 satire dossier** — **HIGH (satire content).** The largest untapped satire vein,
   and uniquely **energy-thesis-adjacent** (mining draw → the datacenter/depletion arc). ICOs, rug
   pulls, NFT mania, DAOs, FTX-class collapses, pump-and-dump. Feeds `08` content banks and possibly
   a faction/era beat. Currently 1–3 scattered mentions.

4. **Labor / unionization + tech-worker organizing** — **MED.** Rebalances an executive-heavy satire
   and gives the commons/Ethical% counter-thread a real solidarity vocabulary. Pairs with the
   "melancholy is real data" pillar; the worker ground-floor is currently under-written.

5. **4X / grand-strategy as *gameplay*** — **MED.** Not the event-chain architecture (already done)
   but the four X's, tech trees, snowball/victory-condition pacing, and 4X-AI. The designed ending's
   terminal loop depends on this pacing well and it has no study behind it. *Could merge with #2 as
   one "endgame grammar" commission.*

6. **Open-source / hacker / demoscene culture** — **MED.** Feeds the Open Source faction voice and
   the commons satire; the demoscene/cracktro aesthetic doubles as free T0–1 era flavor. FOSS
   history, license wars, corporate capture, maintainer burnout, xz/Log4j.

7. **Sports / franchise-management sim** — **MED-LOW.** Closest genre-DNA to the whole game
   (people-and-numbers management). Depth for the Advisory Board / hiring / roster layer and a direct
   "manage humans as stats" satirical target. Lower only because gacha-roster meta is partly covered.

8. **Mentorship + streaming/spectator integration** *(pair as one "social retention surfaces"
   commission)* — **MED-LOW.** Veteran-guided onboarding (RuneScape/FFXIV/EVE) for the churn-risk
   first-prestige, plus streamer-mode / race-broadcast hooks native to the speedrun framing. Partly
   marketing-adjacent, so gated behind foundation completion.

*Lower-tier, batchable when convenient:* tower-defense design study (LOW-MED), Zachtronics
programming-puzzle micro-genre incl. Human Resource Machine (LOW), Brignull dark-pattern provenance
one-pager (LOW), ad-tech mechanics (LOW).
