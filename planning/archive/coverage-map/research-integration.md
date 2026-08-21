# Research Integration Map — 2026-08-06 dossier batch → design + RFCs

> **FROZEN HISTORICAL SNAPSHOT — NONCANONICAL.** The routing language below records the
> 2026-08-06 research wave and assigns no present work. Use the platform-alignment research,
> decision and execution queues for current routing.

Maps the new research (2026-08-05/06 batch) to its design-doc targets, new gap-backlog entries, and
the DESIGN-GAPs that need an owner ruling. Nothing here is auto-applied to the design docs — this is
the routing sheet; the design/RFC edits it schedules are tracked as their own follow-through.

## New dossiers this batch

| Dossier | Feeds | Status |
|---|---|---|
| `player-markets.md` | Trading (NPC/aggregate, P2P-ready) | banked (Wave-B/C) |
| `regulatory-capture.md` | Lobbying/Influence-spend | banked; naming collision flagged |
| `rhythm-timing-games.md` | Active-Play click successor | folded into `active-play-buff-windows.md` AB4 |
| `roguelike-survivor-minigames.md` | Minigame content (The Pitch/Balatro first) | in gap-backlog |
| `soul-mechanic.md` | Soul RFC (hybrid) | folded into `soul-foundation.md` |
| **`cozy-relaxing-games.md`** | Soul-recovery / cozy minigames | **routing below** |
| **`believable-pet-personality.md`** | Pet layer depth | **routing below** |
| **`board-game-mechanics.md`** | Minigame content | **routing below** |
| **`_completeness-sweep.md`** | Research roadmap | **commissions below** |

## Routing the four new dossiers

### cozy-relaxing-games → `design/03` + `design/02 §8` + gap-backlog
- **New design content:** a **cozy / touch-grass minigame category** — the far corner of "no loss AND
  no gain." Reward-that-replaces-the-resource is one of: *completion / aesthetic authorship / rhythm*.
  Key property: **produces-nothing == zero-cheat == trivial server + trivial AI-fallback** (a
  zero-reward toy has nothing to validate). All solo → law-6 fallback trivially met.
- **New gap-backlog minigames (3):** *Defragment the Disk* (zen tile-layer), *The Server Room*
  (Townscaper-style toybox), *Repot the Server-Room Plants* (daily ritual) — each a **Soul-recovery /
  touch-grass** activity that costs deliberate time and grants zero resource.
- **⚠ DESIGN-GAP (owner ruling — README rule 5):** `design/03 §5` Board-Game Suite claims it *restores
  Soul* AND pays *Clout + cash* — this contradicts `design/02 §8`'s "touch-grass produces nothing"
  (Stardew is the cautionary proof: a rewarding activity is not restful). **Recommendation:** a
  distinct touch-grass category pays literally zero and is the ONLY Soul-recovery source; strike the
  board-games' Soul-restore claim (board games are *active* content that pays, not recovery). Routed
  to the owner (below).

### believable-pet-personality → `design/04` + a candidate RFC
- **The extracted rule:** `PERSONALITY = trait vector (bias) × memory (decaying weighted event log) ×
  aggregation (fixed rule)` — every part DETERMINISTIC. Believability = *legible, consistent,
  well-timed reaction*, not simulation depth or randomness. **Our deterministic FSM is the right tool,
  not a tax** — RNG-on-reactions actively breaks consistency-of-character.
- **Design content for `design/04`:** reaction vocabulary (presentation-only), temperament-biased
  behavior chains (activates the reserved **C17 `from_state`-qualifier hook**), recognition/greeting
  keyed on trust+Soul, and **staging the Soul-greying as reversible bereavement** (the withdrawal that
  lands only because everything else was the deposit — no-death's substitute for permanent grief).
- **One structural add → a candidate RFC (`pet-personality-memory` or a pet-care amendment):** a
  **care-history memory ledger** — a small Founder-save field, reusing the **C18 partition-invariant
  decay primitive** (`fixedgrid`), giving the pet a remembered past. This is the single mechanical
  addition; everything else is presentation. No RNG, replay-safe, server-authoritative.

### board-game-mechanics → `design/03` + gap-backlog + a flag upgrade
- **New gap-backlog minigames (3, ranked):** ① *Vertical Integration* (engine-builder/Splendor —
  `base × Π(discount)` **maps onto our production buff-stack**, the tabletop echo of Balatro); ②
  *The Draft* (7-Wonders-style simultaneous pick-and-pass — **best multiplayer fit**, zero turn
  latency, invisible bot backfill, a Break-Room social surface); ③ *Capacity Planning* (Azul-like —
  over-provisioning-is-billed is the sharpest satire beat).
- **⚠ Flag upgrade (design/03 §5b):** **social deduction is NON-COMPLIANT** — it violates TWO binding
  laws (needs a free-text medium + no honest same-rules bot fallback exists). Upgrade §5b's flag from
  "design-risk" to **"do NOT build the human-vs-human hidden-role form"**; the only compliant
  expression is a **PvE Clue-style deduction puzzle**. (Routed to the owner as a confirm.)
- Engine-builders carry the same **Fairness-Law hazard** as Balatro: must be sandboxed to a seeded
  economy; tier counts enter only as `breadth`, never as ranked power.

### _completeness-sweep → research roadmap (next commissions)
**PHASE SEQUENCE (owner-ruled 2026-08-06): Wave-A (foundations) → THESE RESEARCH COMMISSIONS → Wave-B
(content).** The commissions run as their own phase AFTER Wave-A completes and BEFORE any Wave-B
content RFC. The corpus is deep; gaps are content-feeding research, not built systems. **Recommended
commissions** (do NOT auto-launch — owner paces):
1. **Factory / automation** (Factorio/Satisfactory/shapez) — HIGH. The **T7–T8 endgame grammar**
   ("allocate autonomous extractors", `01 §T7`) has NO gameplay reference today. Merge with 4X.
2. **M&A Arena / absorption** (agar.io/Katamari/Osmos) — HIGH. **Process-blocked:** a named v1.0
   minigame is a `❌ GAP` row in `research/README.md`; its RFC cannot be drafted until researched.
3. **Crypto / web3 satire** — HIGH. Largest untapped satire vein; energy-thesis-adjacent (feeds the
   depletion arc).
4. **Labor / unionization + tech-worker organizing** — MED. Rebalances executive-heavy satire; gives
   the commons / Ethical% counter-thread a solidarity vocabulary.
5. **4X / grand-strategy as gameplay** — MED. The designed ending's terminal 4X loop has no pacing
   study (merge with #1 as one "endgame grammar" commission).
Lower-priority batchables: open-source culture, sports-management, mentorship/streaming, tower-defense,
Zachtronics puzzles, the Brignull dark-pattern provenance one-pager.

## Owner rulings — RESOLVED 2026-08-06 (both applied to design/03)
1. **Board games vs Soul-recovery → RESOLVED:** cozy/touch-grass is the ONLY zero-reward
   Soul-recovery source; board games pay-but-don't-restore. §5 struck the Soul claim (kept the lock).
2. **Social deduction → RESOLVED (owner chose human-only, not my PvE recommendation):** the
   designated HUMAN-ONLY feature (AI-too-much-hassle carve-out), STRUCTURED comms only, no bot
   fallback (empty lobby = unavailable). §5b flag upgraded.

## Follow-through (scheduled, not yet done — candidate "content wave")
- Fold cozy category + 3 minigames into `design/03`; add the touch-grass category to `design/02 §8`.
- Deepen `design/04` with the pet reaction/memory/bereavement content.
- Add the 3 board minigames + the §5b social-deduction flag to `design/03`.
- Record the 5 commissions in `design/research/README.md`.
- Candidate new RFCs (Wave-B content): pet-personality memory ledger; the first minigame content RFC
  (The Pitch/Balatro); the cozy/touch-grass minigames.

## Research commissions — status & the future roadmap

### NOW RUNNING (the 3 HIGH gaps, launched 2026-08-06, this commission phase)
- `endgame-grammar.md` — factory/automation + 4X (the T7–T8 endgame + terminal consumption loop; the biggest gap).
- `absorption-arena.md` — agar.io/Katamari/Osmos (unblocks the process-gated M&A Arena minigame).
- `crypto-web3-satire.md` — the crypto grift + energy thesis (feeds the depletion arc).

### QUEUED for this phase (MED, from the completeness sweep — pace as desired)
- Labor / unionization + tech-worker organizing (rebalances executive-heavy satire; commons/Ethical% solidarity).
- (open-source culture, sports-management, mentorship/streaming, tower-defense, Zachtronics, Brignull one-pager — low-priority batchables.)

### FUTURE commissions — the SHIP/POLISH wave (post-Wave-B; craft/liveops/finale, NOT content mechanics)
These are new angles beyond the completeness sweep — needed to make the content shippable and good,
not to create it. Identified 2026-08-06.
1. **Designed-ending craft** — how games *execute* a designed ending well (Universal Paperclips, Spaceplan,
   Outer Wilds, NieR:Automata, A Dark Room, Gris). We have the ending CONCEPT ("both datasets real"); this is
   the emotional-payoff/execution craft. HIGH for the finale.
2. **Onboarding & first-session retention** for idle/MMO — the critical first 5 minutes; tutorial-as-content;
   what hooks vs. loses an idle player. HIGH (churn is the #1 idle failure).
3. **Localizing satire/humor at scale** — heavily satirical text; localizing humor is notoriously hard; how
   games do it (and whether to). MED–HIGH given the flavor-heavy design.
4. **Live-ops & seasonal cadence playbook** — the OPERATIONAL side of running seasons/events (GM tooling,
   content pipeline, community mgmt) — we have the events ENGINE design, not the liveops playbook. MED.
5. **Speedrun community & verification mechanics** — our whole framing is speedrun; how communities/
   categories/leaderboards/verification/anti-cheat work as SOCIAL systems. MED (we have the framing, not the
   community mechanics).
6. **Accessibility for click-heavy/idle games** — colorblind, screen-reader, motor (clicking!), reduced-motion.
   MED, ethically load-bearing for a clicker.
7. **Audio & music design** for idle/cozy/rhythm — the soundscape; the cattery sound-event-bus; satirical audio.
   MED (the rhythm minigames need it; cozy leans on it).
8. **MMO anti-cheat / security red-team** — a deep security pass for the live game beyond the server-authoritative
   design already in place. MED–HIGH before the public launch.
9. **Moderation & community governance** — commons/trading/guilds/the human-only social-deduction; how player
   communities are actually moderated (we designed AROUND moderation cost; this studies the real thing). MED.
10. **Narrative delivery in idle games** — telling stories with minimal UI (A Dark Room, Paperclips, Cookie
    Clicker lore); the delivery craft for our conspiracy/media/tier narrative. MED.

## Commission-phase dossiers (2026-08-06) → routing

### endgame-grammar → `design/01 §T7-T8` + the ending
- **The T7–T8 grammar is now named (and we already ship it):** the endgame is a **closed-form
  `min()`-over-rates** abstract logistics layer (the shipped Thermal Law IS this grammar — no per-tick
  belt sim). We are already a 4X with eXterminate pointed at the biosphere. "Ends right" = automate
  everything below to a formality (T7) + introduce EXACTLY ONE new decision + consume a FINITE resource
  whose exhaustion IS the credits roll (T8 Planet → both-graphs ending). **T7 decision = triage under
  irreversibility (what ascends); T8 decision = tragedy-of-the-commons rate choice (Sustainable vs
  Burn).** Universal Paperclips Act 3 is our T7→T8 almost line-for-line (fixed-size allocation +
  closed-form attrition, not an RTS).
- **⚠ DESIGN-GAPs to route (design/01):** the exact T7/T8 split of the "autonomous extractors"
  mechanic, and the mechanical form of "value drift" (fighting your own swarm). Both to a future
  endgame content RFC.

### absorption-arena → `design/03` (M&A Arena) + homes the punch-down multiplier + `design/13`
- **The M&A Arena is tractable (Form B):** single-human, server-authoritative real-time arena with
  viewport-bounded bots + read-only company snapshots — **same cost as the already-greenlit lane
  pusher** (agar.io's cost is humans-per-arena, not real-time). Many-human = a later tournament stretch.
  No offline victim. Distinct axis (size/position, no type chart) → not "the same game thrice".
- **The punch-down multiplier (200%/5%) is HOMED:** it lands on acquisition payout by relative size,
  beside two genre-native anti-snowball mechanics that ARE the satire — **antitrust = agar.io's virus**
  (biggest cell shattered) and **diseconomies of scale = mass decay**. It applies to ALL PvP payout
  surfaces (lane-pusher + arena). (deferred tracker updated.)
- **⚠ Owner-note:** proposed host Tier 4+, session-skill clock — `design/03 §5`'s unlock stagger doesn't
  list the arena yet; place it at the content RFC. Katamari = the endgame roll-up ceremony (reuse the
  `13` diorama); Osmos = a separate tiny Soul-safe toy.

### crypto-web3-satire → `design/08` + `design/10 §1` (crypto faction) + the depletion arc
- **The energy thesis is the load-bearing bridge:** proof-of-work mining = the Tier 3→4 historical
  dress-rehearsal for the AI-datacenter draw the depletion arc already runs on — same Externality
  ledger, same Jevons Engine, reuse the two-number water counter as an energy counter.
- **The crypto faction is fleshed out** (against `10 §1`'s upgradeable-RNG-volatility seed): anchor =
  Reserves, failure state = **Depeg** (Terra's death spiral as a mechanic), negative space = stability
  is anti-synergy; curtain-pulled upgrades (Leverage, Algorithmic Stablecoin, Diamond Hands, Yield
  Farm, Rug Pull) + a `Depeg Warp` route candidate.
- Content banks (buildings/achievements/tickers) in the flavor-bible voice; **never punch down** on
  Filipino scholars / retail bagholders (victims, not butts); cross-refs defer Axie/Ronin/Quartz to
  `gaming-enshittification.md` and the base energy/Jevons to `societal-satire.md`.

### New gap-backlog / candidate RFCs from this phase
- **M&A Arena minigame** (Form B; now research-unblocked) — Minigame Platform + the punch-down payout.
- **Crypto faction** — extends the Faction system (`faction-incorporation` is archived) → a faction
  content RFC.
- **T7–T8 endgame content** — the closed-form extractor/rate grammar → a late endgame RFC (far future).

_All three integrated into the routing map; the deeper `design/01/03/08/10` prose edits happen at each
content RFC's drafting (the design-doc-updated-with-the-change workflow), tracked here._

## 2026-08-07 batch → routing
- `onboarding-retention.md` → design/11 §1b (five adoptions RULED: account A, Vision Slide required,
  WR/presence defaults, satire beat); feeds Game-UI screens + the API/Surface RFC directly.
- `labor-organizing.md` → design/08 (employer-side temptation shelf + Ethical% union-recognition
  fact), design/09 (strike as pressure-meter disaster; Burnout "fires" column), Soul drain content
  (sourced crunch beats: EA Spouse / Rockstar / CDPR / Raven). Fills gaming-enshittification §223's
  unionization stub. Excluded-content lines absolute (no Foxconn material, employer is the verb's
  subject). Content lands with the events-engine content RFCs.
- `era-1995-satire.md` → design/11 §1b (beat RESOLVED: ORDER NOW $0.00 + count-up nag + README.TXT),
  T0/T1 copy corpus (~40 lines) for the T0-T1 content RFC's copy pass; "seventeen Super Bowl ads"
  BANNED (verified: 12 of 61).

## 2026-08-08 — the DISCIPLINE sweep (owner: "no stone left unturned")

Prior sweeps were genre/content-shaped; this batch sweeps by discipline. **NOW RUNNING (all 8
dispatched 2026-08-08):**
1. `healthy-engagement.md` — anti-addiction / design-for-stopping (thesis-completing).
2. `designed-sunset.md` — MMO end-of-life as day-one content; save portability (thesis-completing).
3. `ai-authorship-meta.md` — disclosure norms + 2026 AI-slop landscape + our own provenance.
4. `launch-distribution.md` — how free browser games find players; feeds THE PUSH.
5. `mobile-pwa.md` — phones/PWA/notifications vs our actual stack (tech gap).
6. `spectator-racing.md` — broadcast-native speedrun design; races, the eval-bar equivalent.
7. `arg-mechanics.md` — ARG craft + the ethics boundary for the conspiracy layer.
8. `gamification-satire.md` — the sins escaped gaming (fintech/fitness/work); reverse-satire corpus.

## THE STANDING NOT-YET-RESEARCHED REGISTER (owner rule 2026-08-08: everything unresearched gets
## noted here; nothing drops silently)

Ship/polish wave (identified 2026-08-06, still unrun): designed-ending craft · localizing satire ·
live-ops/seasonal playbook · speedrun community & verification · accessibility (incl. the
number-comprehension/dyscalculia slice from the discipline sweep) · audio & music design ·
MMO anti-cheat red team · moderation & community governance · narrative delivery in idle games.
Low-priority batchables (2026-08-06): open-source culture · sports-management · mentorship/
streaming · tower-defense · Zachtronics puzzles · Brignull dark-pattern provenance one-pager.
Noted 2026-08-08 (identified in the discipline sweep, deliberately not dispatched): game
preservation/museum-exhibit craft (feeds the "games we killed" memorial) · the phenomenology of
waiting (academic; fold into pacing-science if ever needed).

## 2026-08-08 — discipline sweep COMPLETE (8/8 landed) + consolidation
All eight dossiers delivered (healthy-engagement, designed-sunset, ai-authorship-meta,
launch-distribution, mobile-pwa, spectator-racing, arg-mechanics, gamification-satire); coverage
matrix rows added by each agent. Owner rulings recorded in decisions-log (license, Law 10 kept,
multitouch, provenance-standard-practice + AGI-breaks-rules content direction, rested-bonus
candidate → design/02 §9, open beta + EARLY ACCESS™ framing, all three ARG surfaces).

### REGISTER ADDITIONS (not yet done, noted per the standing rule)
- **Verification sweep over the 2026-08-08 batch** — MANDATORY before any of it ships as fact:
  the session's search budget was exhausted, so [M]-density is above house standard across all
  eight (heaviest: gamification §10, launch's subreddit-norms section, sunset's post-cutoff 2026
  claims, AI-meta's V11/V12 incl. the "No Gen AI" seal question, spectator's 16 items, mobile's
  on-device Worker/socket test matrix). Run with a fresh search budget.
- **Dependency license audit** (owner license ruling: Unlicense-or-comply) — before THE PUSH.
- **DESIGN-GAP routing owed from the batch:** design/06 + UI Foundation mobile provisions
  (breakpoints/touch/safe-area/perf budget); transport background-disconnect policy + shell
  lifecycle triggers; design/08 EARLY ACCESS™ + honest-queue beats; design/05 §5 watch-party vs
  non-elective cohorts; spectator anonymous read path (public API); Sunset Covenant RFC (new);
  launch-sequence gates → Deployment Foundation RFC; ARG surface trio → future RFCs; AGI-tier
  breaks-rules beats → AGI-tier content RFC.

### Register addition 2026-08-09
- **New-content harness scenarios** (FCE5.3 review Finding 1): pacing simulation exercising buff
  windows, fiscal harvests, pitch payouts, permits accrual — the epoch-7 retune lane's evidence
  base. Registered at mint sign-off.

### Register addition 2026-08-10
- **R-F1 (numeric core):** force float64 materialization in toFloat64's snap comparison (kills
  the remaining arm64 FMA sensitivity on Floor — a shared-vector op) + a razor-edge floor golden
  vector. Next kernel-bump commit. R-F3: docs/ci.md owes the test-go-ci reproduction line.
