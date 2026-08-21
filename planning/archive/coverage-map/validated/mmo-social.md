# Coverage slice — **mmo-social**

> **FROZEN HISTORICAL SNAPSHOT — NONCANONICAL.** Reconstructed on 2026-08-05; retained as
> evidence, not current status or execution authority. See `planning/platform-alignment/`.

Independently reconstructed 2026-08-05 from the actual `design/`, `design/research/`, `rfc/`,
`rfc/archive/`, and `docs/` files. Every stage cites file+section. Furthest stage is per the four-tier
protocol (R → D → F → I). Adversarial flags per row.

## Reconstruction table

| System | R (research) | D (design) | F (RFC + actual status) | Furthest | Flags |
|---|---|---|---|---|---|
| **Commons** (Mutual Aid Compact — Ostrom buff pool, Health/Capacity, cohort, tithe governance) | **`commons-game-theory.md §5`** is THE spec (§5a/b two-axis Health-mean/Capacity-sum; §5c cohort "load-bearing invention"; §5d full buff closed-form + votable tithe band; §5f graduated sanctions). `morality-systems.md` owns only the single-player half and cross-refs the pool to §5 (§7.5 revises pacing only) → cross-ref, not independent spec. **BACKED-SPEC**. | `design/05 §5` (two axes, buff `1+5·[0.6·f(H)+0.4·sᵢ]`, nested Health, 150-cohort, governance, front door); `design/13 §4` ("commons choice" ties to Planet); `design/10 §5` (Ethical% via commons) | **Server foundation: implemented** — `archive/commons-compact.md`, `archive/commons-projection-retry-idempotency.md`, `archive/commons-merge-capacity.md`, `archive/post-review-integrity-remediation.md`. **Player-facing half: draft (UNBLOCKED 2026-08-03)** — `commons-onboarding-and-governance.md` (cohort UI, incorporation line item, monthly tithe ballot, guild Health binding) | **I** (server foundation) | Clean. Onboarding/governance/cohort-UI/tithe-vote = F-draft, not yet implemented. |
| **Guilds** (tithe, leader upgrades, contribution grading, faction interdependence exchange) | Split: **BACKED-SPEC** for Break Room (`social-spaces.md §4.2`), faction-interdependence cycle + percentile-within-faction grading (`events-playstyles.md §2e`; `liveservice-idle-tier.md §4.2`), and the warm-channel layer (`liveservice-idle-tier.md §5.1/§5.3` — help taps, donations, ritual, roles). **BACKED-MENTION only** for the automatic tithe, leader-purchased upgrades, and contribution windows — `liveservice-idle-tier.md §5.1` attributes these to design `05 §3`, does not itself spec them. | `design/05 §3` (formation 2–50, automatic tithe, leader-purchased upgrades, contribution windows, faction interdependence cycle, Break Room, matchmaking grades, NPC fallback); `design/04 §4` (no-scarcity satire) | **implemented** — `archive/guild-model.md` (G1–G6 + GD1–GD6 + GD5a; structural guild, tithe, Health term, faction exchange clearing, NPC virtual-guild) + `archive/gameserver-composition.md` (GC3/GC3a composes `PendingSettlements`, clearing driver) | **I** | **UNBACKED (partial)** — the shipped automatic-tithe economics are design-originated, not research-specced (see F3). Large **deferred content** surface (NOT built): Break Room seats, guild events/bosses, contribution windows, matchmaking grades, leader-purchased upgrades, public guildless exchange, percentile display. |
| **World layer** (community milestones, planet depletion ratchet, Influence-by-rank) | Split: **BACKED-SPEC** for the Helldivers Galactic Impact Modifier / published contribution formula (`events-playstyles.md §1c/§1e`), the Jevons planet-depletion ratchet (`societal-satire.md §5`; `map-attraction.md Part 4c`), and the Influence field (`map-attraction.md Part 3`). **UNBACKED** for the **Elite-Dangerous dual reward axis** — no dossier specs or even sources it; it appears only as a back-reference to `design/05 §1`. | `design/02 §1` (Layer-3 World row: never-resets ratchet, milestone tiers, Influence) + `§4` (dual-axis milestone payout, Influence spent on world cosmetics/GM-shop/guild seeding); `design/05 §1` (global milestones, Elite-Dangerous dual axis, published contribution formula, impact modifier); `design/13 §4` (per-region Planet ratchet, epoch act-breaks, seasons) | **draft** — `world-layer-foundation.md` (WL1–WL5; created 2026-08-03). GC2 world-snapshot *schema* is implemented (`archive/gameserver-composition.md`, `docs/gameserver.md`, `docs/transport.md`) but publishes `planet`/`milestone` as **zero/null** "until those owners ship". | **F** (draft) | **STATUS-OFF** (README omits the draft) + **UNBACKED** (dual-axis reward, F4). |
| **Feed / dispatch / presence** | `events-playstyles.md §Ops lessons` (dispatches = highest-ROI narrative device; GW2 stuck-event watchdog) + `neopets-social-history.md §3` (Neopian-Times editorial-UGC gate; vote-based → social-capital-market hazard) + `social-spaces.md` (moderation model). Presence lineage: `cattery-reusables.md` (activity feed). **BACKED-SPEC**. | `design/05 §2` (live feed, global counters, ghosts, feed prominence ∝ Clout), `design/05 §Neopets adoptions` (amplification-gated curation), `design/09 §6` (dispatches = the narrative device), no-free-text law `design/00` | **draft** — `feed-and-dispatch-foundation.md` (FD1–FD4; 2026-08-03). Presence/channel **substrate implemented** — `docs/transport.md` (`feed`/`world`/`cohort`/`guild` channels, `presence` kind, aggregate join/leave, 50-item feed history) + guild presence relay (`docs/guilds.md`, `docs/gameserver.md`). | **F** (draft) for feed/dispatch content; **presence substrate = I** | **STATUS-OFF** — not in `rfc/README.md` Active index. `feed` channel exists but carries nothing until this RFC ships. |
| **Trading / player market** (server-global Market prices, cosmetic/seed gifting, anti-RMT non-transferable) | Split: **BACKED-SPEC** for the anti-RMT non-transferable crate flag (`compliance-2026-refresh.md §3.1/§8/§9.1` — the loot-box-ruling hinge, the sole code item in its P0) and for the *NPC stock-market minigame* (`neopets-systems.md §2` NEODAQ, `kol-puzzle-pirates.md §A2` KoL Mall — both explicitly for `design/03 §4`). **BACKED-MENTION** for gifting (`social-spaces.md §2`). The **player-to-player** market framing is only weakly backed — research specs an NPC/aggregate market, not player trade. | `design/03 §4` (The Market minigame — server-global prices lightly moved by aggregate trading + Situations); `design/05 §4` (The Market row; Trading/gifting: cosmetic+seed only, crate-derived non-transferable) | **none** | **D** | **GAP** — builds on: economy-kernel + transport + purchasable-content (crates/gifting) + world-layer (Situations move prices). "The Market" as a *minigame* also lands in the minigames slice. |
| **Server story arcs** (Layer-3: Situations · Major Orders · seasonal engine · GM ops/war log) | `events-playstyles.md §1` (CK3 Situations via on_actions; Helldivers Major Orders + impact modifier; GM war log; watchdogs) + `neopets-social-history.md §1` (plot-prize shop, communal puzzles, war waves, evergreen conversion, annual tournament) + `liveservice-idle-tier.md §3` (event-scrip exchange shop). **BACKED-SPEC**. | `design/09 §4` (Situations, Major Orders, rule-break events, one-time moments), `§5` (exchange shop), `§6` (seasonal arcs, dispatches, launch-year sketch), `§7` (GM ops, war log, watchdogs) | **none for Layer-3.** `events-engine-layer1.md` (**draft**) builds the shared evaluator but is **Layer-1 personal events only** — its §E3 explicitly defers "Layer 3 (server events)" to successor RFCs. | **D** | **GAP** — builds on: events-engine-layer1 (shared evaluator) + world-layer-foundation (milestones/depletion the arcs drive) + feed-and-dispatch (the dispatch surface). |
| **Lobbying / Influence-spend world surface** (billionaire-layer capture, Clout-gated regulatory outcomes) | **UNBACKED** — no dossier specs a doctrine or lobbying MMO mechanic; "lobbying" appears only as flavor (ticker lines) or as speedrun-skip names (`speedrun-governance.md §9` "Regulatory Capture Skip"). Adjacent-but-different: `map-attraction.md §4b/c` specs a world-*regulation* system (per-region Regulatory Capacity, pledges, coalitions), not lobbying. | `design/02 §6` (Clout gates lobbying/regulatory outcomes), `§4` (Influence spend targets); `design/13 §4` (influence field, attribution drama); `design/08` billionaire layer | **none** — `world-layer-foundation.md` WL5 declares the Influence *mint + ledger* but explicitly **defers spend surfaces** ("late-tier lobbying/capture events … later content"). | **D** | **GAP + UNBACKED** — builds on: world-layer-foundation (WL5 Influence ledger must land first) + events engine. |

### Out-of-domain (checked, reclassified)

- **Doctrine** (`design/10 §3b`) is **run-scoped** ("Doctrines are run-scoped (reset on Exit)"), not
  MMO-scoped. Its foundation `rfc/doctrine-and-compute-credit.md` (draft, 2026-08-03) belongs to the
  **economy-progression** slice. Noted here only because the domain prompt listed it conditionally;
  it fails the "if MMO-scoped" condition. (Same README STATUS-OFF applies — see Findings.)

## Furthest-stage tally

- **I (Implemented):** Commons (server foundation), Guilds — **2**
- **F (Foundation/draft):** World layer, Feed/dispatch — **2**
- **D (Designed, no RFC → GAP):** Trading/market, Server story arcs (Layer-3), Lobbying/Influence-spend — **3**

7 in-domain systems (+ Doctrine reclassified out-of-domain).

## Findings (evidence)

### F1 — STALE RFC INDEX: four 2026-08-03 draft RFCs are absent from `rfc/README.md` (STATUS-OFF, cross-cutting)
`rfc/README.md` line 69 states: *"Remaining Phase-0 contracts (not yet drafted): Layer-1 events
engine · doctrine intents · Compute Credit spend · game-UI screens · deployment."* But draft RFC
files dated 2026-08-03 already exist for all three named systems, plus two more this domain owns:
- `rfc/world-layer-foundation.md` (draft) — **not in the Active table, not mentioned anywhere**.
- `rfc/feed-and-dispatch-foundation.md` (draft) — **not in the Active table, not mentioned**.
- `rfc/events-engine-layer1.md` (draft) — README calls "Layer-1 events engine" *not yet drafted*.
- `rfc/doctrine-and-compute-credit.md` (draft) — README calls "doctrine intents / Compute Credit spend" *not yet drafted*.

Each file's own status line reads `draft`; the README index contradicts them. This is a
documentation-truth defect the coverage map exists to catch: three of this slice's furthest-stage
determinations (World layer = F, Feed/dispatch = F, and the Server-arc GAP's substrate) turn on
these files existing. **The README's Active table and its "not yet drafted" line must be reconciled
to the draft files.**

### F2 — Commons: furthest stage is I, but the *social* half is still F-draft (not drift)
`docs/commons.md` + `archive/commons-compact.md` (and the three hardening RFCs) implement the
server foundation: membership, derived Enclosure/compliance, the `commons.member` modifier,
150-Founder cohort assignment, collapse-merge, projection idempotency, NPC fallback, generated
formula artifact. **Explicitly NOT shipped** (per `docs/commons.md` ¶2 and the compact's own body):
the incorporation-contract line item, Open-Source auto-sign UI, cohort panel, the guild Health
term binding, and the monthly direction-only tithe vote. Their successor
`commons-onboarding-and-governance.md` is `draft — UNBLOCKED` (2026-08-03): all six original
blockers answered against now-implemented owner contracts (account, faction, guild, transport).
Status is internally consistent (README, RFC status line, and the RFC's own 2026-08-03 rulings
agree) — **no STATUS-OFF**. Note the guild Health term (`0.5·H_guild`) is specified but inert in
production today: `docs/commons.md` says every member currently takes the guildless 80/20
substitution.

### F3 — Guilds: implemented structurally; a research-depth caveat and a deferred content backlog
`archive/guild-model.md` (implemented) + `docs/guilds.md` ship identity/membership lifecycle, the
automatic tithe (tier-progress-normalized, `guild_tithe_ppm`), the population-invariant guild
Health input, deterministic faction-exchange clearing (activating F2's inert stock), the NPC
virtual-guild fallback, and the transport resolver. `archive/gameserver-composition.md` composes
the real clearing driver.
**Research-depth caveat (soft UNBACKED):** the research survey found that the *shipped* economics —
the automatic tithe, leader-purchased upgrades, and per-hour contribution windows — are only
NAMED in research (`liveservice-idle-tier.md §5.1` attributes them to `design/05 §3`), not
independently specified there. What the dossiers actually spec is the *social* layer
(Break Room `social-spaces.md §4.2`, warm-channel reciprocity `liveservice-idle-tier.md §5.1/§5.3`)
and the faction/grading arithmetic (`events-playstyles.md §2e`). The tithe design is sound and
implemented — but its research grounding is a mention, not a dossier spec. Not disqualifying (the
mechanic is an idle-genre convention), flagged for honesty.
**Deferred to successor RFCs** (design/05 §3 features with no contract yet): the Break Room
seat-room, guild events/bosses + per-hour contribution windows, matchmaking grades (C→AAA),
leader-purchased upgrades, the public guildless exchange, and percentile-within-faction display.
These share the guild foundation (not independent GAP rows) but are the guild content backlog.

### F4 — World layer: mechanics are F-draft; only the transport envelope is implemented (and honestly zeroed)
The `world` snapshot schema (GC2) is implemented and validated in both runtimes
(`docs/transport.md` ¶ "World snapshot state is a closed version-1 integer schema", `docs/gameserver.md`),
but `docs/transport.md` states plainly: *"Planet and milestone values are deliberately zero/null
until those owners ship."* The owning RFC `world-layer-foundation.md` (WL1 world singleton, WL2
published Helldivers contribution formula, WL3 dual-axis milestones, WL4 depletion ratchet, WL5
Influence mint) is **draft**. So the *mechanic* is furthest-stage F; the implemented schema is a
transport envelope, not the system. D→F fidelity is good (RFC WL2–WL4 track design/05 §1 and
design/13 §4 closely — Helldivers modifier, dual axis, monotonic ratchet, published formulas). No
DRIFT. **UNBACKED sub-mechanic:** the "Elite-Dangerous dual reward axis" that design/05 §1 leans on
has **no research dossier** — the survey found it only as a back-reference to the design doc itself
(`commons-game-theory.md` cites `05-mmo.md §1`, not a source). The dual-axis *idea* is carried
operationally by the (backed) Helldivers impact modifier and the commons dual-axis, so this is a
citation-hygiene flag, not a mechanic at risk — but design/05 §1 should not present "Elite
Dangerous" as researched precedent.

### F5 — Server story arcs (Layer-3) are a real GAP masked by an adjacent draft
`events-engine-layer1.md` exists and is draft — but it is deliberately **Layer-1 personal events
only**. Its §E3: *"Layer 2 (meters) and Layer 3 (server events) are successor RFCs on the same
evaluator."* Design/09 §4–§7 fully specify the Layer-3 surface (Situations, Major Orders with the
published impact modifier and canonical failure, the ~3-month seasonal arc engine, GM war log,
watchdogs), and research backs it richly (`events-playstyles.md §1`, `neopets-social-history.md §1`).
**No RFC drafts the Layer-3 engine.** This is the highest-leverage mmo-social GAP: it is the
"MMO layer" the world-layer RFC calls the thing that "makes it an MMO rather than parallel solo
games", and it depends on world-layer + feed/dispatch + the events evaluator all landing first.

### F6 — Trading/market is backed research, not carried to a foundation (GAP, not ORPHAN)
The Market has genuine research depth (`neopets-systems.md §2` NEODAQ, `kol-puzzle-pirates.md`
Mall/labor markets) and the anti-RMT non-transferable-crate rule is a dated 2026-07-28 decision in
`design/05 §4` citing `compliance-2026-refresh.md`. Design exists (`design/03 §4` minigame + `05 §4`
trading table). **No RFC.** Furthest = D → GAP. Lowest draft priority of the three GAPs (it is
launch-window content — design/09 §6 sketches "The Market opens" only in season S3).

### F7 — ORPHAN check: five research-survey candidates all refuted (none is a true orphan)
The research survey flagged five "researched-but-maybe-uncarried" candidates. Verified against
design — all are carried (so **no ORPHAN in-domain**):
- `map-attraction.md §4b` adoption-spread map (Plague-Inc regions, 3-bar Adoption/Disruption/Harm,
  coalition bases) → carried into **`design/13-world.md §3` (Layer B)** and §4. [v1.0]/[post-1.0], no RFC.
- `map-attraction.md §4c` season-finale + per-player depletion shame board, Sustainable-vs-Burn
  pacts, r/place whiteout reset → carried into **`design/13-world.md §4`**.
- Eco-style player-authored-and-voted laws → deliberately **narrowed** ("pacts/pledges, not writing
  laws", `design/13 §4`; commons polls direction only, `design/05 §5`) — a conscious rejection, not
  an orphan.
- `events-playstyles.md §2e` faction-as-verb rulesets → carried into **`design/10 §1`** (factions
  as distinct verbs); the full rule *packages* are deferred content (faction-incorporation ships
  empty modifier lists — economy-progression slice).
- NEODAQ/KoL-Mall market source → `design/03 §4` (The Market minigame) exists; see F6.

### Adversarial-check summary
- **UNBACKED:** three real sub-mechanic findings — (a) World layer's **Elite-Dangerous dual reward
  axis** (no dossier, F4); (b) **Lobbying/Influence-spend** entirely (no doctrine/lobbying dossier);
  (c) **soft** — Guilds' automatic-tithe/upgrades/contribution-windows are research-*mentioned*, not
  -specced (F3). The core buff/formula/depletion/dispatch/arc mechanics ARE dossier-specced.
- **ORPHAN:** none — all five survey candidates carried into design/10, /13, or /03 (F7).
- **DRIFT:** none — the four draft RFCs are faithful boundary/mechanic splits (world-layer,
  feed/dispatch, commons-onboarding, guild-model restate design laws as contracts).
- **STATUS-OFF:** F1 — the README index is stale against four 2026-08-03 draft RFCs (World layer,
  Feed/dispatch, Events Layer-1, Doctrine/Compute-Credit). The dominant contradiction.
- **GAP:** Trading/market, Server story arcs (Layer-3), Lobbying/Influence-spend.
</content>
</invoke>
