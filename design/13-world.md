# The World — Diorama, Map & Spread

> The **attraction**: the visible, growing, degrading thing that makes progress felt (the Civ-city instinct), and the **spatial spine** that most other systems already wanted. Research: `research/map-attraction.md` (Plague Inc deconstruction, visible-progress design, shared maps) + `research/browser-rendering.md` (the renderer).
>
> **Scope discipline:** this doc marks every system **[v0.1]**, **[v1.0]**, or **[post-1.0]**. The spine is designed whole so it stays coherent; only the marked slice ships at launch.

## 1. Why the world is the spine

Systems already designed keep reaching for a place to happen:

| Existing system | Its spatial home |
|---|---|
| Externality (`02`) — "leaves your ledger, lands on the map" | the map, literally |
| Datacenter regions + constraints (`01` T4) | region attributes (power/water/permits) |
| Doctrines "where do you build" (`10 §3b`) | region selection as a run-defining choice |
| Pressure meters — outrage, regulatory heat (`09`) | per-region meters + the global bar |
| Community milestones & presence (`05`) | the shared map is their natural surface |
| Planet % ratchet (`01` T8) | the world's own degradation, rendered |
| Clout/lobbying (`02`) | influence overlays |
| Satire content (`08`) | headlines anchored to places |

Three layers, one renderer, one state reducer.

## 2. Layer A — The personal diorama (the attraction) **[v0.1]**

**Architecture rule (non-negotiable):** the diorama is a **pure function of simulation state** — one `WorldState → SceneSpec` reducer; the renderer only draws a SceneSpec. Consequences: the scene can never lie, it's snapshot/replayable, and it renders **headless** for share-images and the world map. (GlassBox principle: the agents you see *are* the simulation.)

**Four dials driven continuously** (so the scene animates *between* tier-ups, not only at them):
1. **Structure count/level** → silhouette (identifiable at 64×64)
2. **Density/clutter** → the cheapest legible variable
3. **Ambient agents** → *Wuselfaktor*: figures doing errands at a rate ∝ throughput. Not simulated individuals — **proportional** (a player should estimate their production rate by squinting)
4. **Grime/heat/emission** → the moral channel

**The tier ladder** (same assets, retinted/decaled — growth and blight share art):

| Tier | Scene | Wusel | Grime |
|---|---|---|---|
| 0 Garage | desk, tower PC, fan, cat, a moth | you + the cat | energy-drink cans accumulate |
| 1 Spare room | 2U rack, space heater, blinkenlights synced to request rate | + a flatmate complaining | heat shimmer, dust |
| 2 Co-working | floorplate | 3–8 employees → coffee machine | pizza boxes, sticky-note wall |
| 3 Startup office | isometric, scrollable | 15–40, standups, delivery bot | whiteboards fill with unreadable diagrams |
| 4 First datacenter | exterior plot | trucks on a loop, techs walking rows | **cooling ponds appear; grass browns at the edges** |
| 5 Campus | multi-building + roads | shuttles, drones, helipad | substation grows; haze over the plot |
| 6 Hyperscale | tiny buildings, huge infrastructure | swarms; figures become dots | **ground desaturates in rings; a reservoir visibly lowers** |
| 7 Orbital/planetary | the plot becomes the globe | satellite streaks | **the terminator line shows city-lights replacing dark land** |

**Rules with teeth:**
- **The garage is never destroyed.** It stays visitable forever as a museum corner of the campus (Civ keeps ancient wonders on modern maps). Nostalgia for the garage is the satire's emotional anchor.
- **Grime is the moral channel** — the player watches success become blight with no UI saying so.
- **Ceremony on tier-up** (2–4 s, camera move, one ticker line, a stinger; unskippable first time). **[v0.1]**
- **Props carry the jokes:** the mission statement growing unhinged, the "Days Since Incident" counter, the ethics review board that becomes one empty chair.
- **Every status gets a bar** (Egg Inc's rule).
- **Instant tactile response to every click**, within one frame, independent of the number it changes.
- **Change while away:** a 5-second "while you were gone" fast-forward on return, ticker replaying. **[v0.1]**
- **Day/night + weather** as a near-free liveness multiplier. **[v1.0]**
- **Visitable dioramas** (guildmates, feed links, screenshot export) — the Clash-base social motivation. **[v1.0]**

## 3. Layer B — Adoption spread: the player as pathogen **[v1.0]**

An inverted epidemic-sim: **you are the pathogen; regulation is the cure race.** The design law it contributes: **power costs visibility, and visibility funds the countdown that kills you.**

**Currency — Buzz.** Spatially-anchored bubbles on the map: **orange** = adoption events · **red** = harm events (*bigger payouts — outrage is engagement*) · **blue** = regulatory progress (popping *delays* it). Passive accrual ∝ Users × Disruption. Cost inflation +1 per shipped feature, with a **`Refactor`** archetype existing to defeat the inflation curve. **Deprecate** = the retraction verb (early game: deprecate everything to stay under detection).

**Three trees** (full tables in the research file):
- **Distribution** — grows adoption, never raises scrutiny. Each vector biases a region attribute (`Free Tier`→poor, `Mobile App`→urban, `Enterprise Sales`→rich, `API`→hubs, `Government Contract`→state-heavy, `Preinstall/OEM`→rural). **Vector choice is geography choice.**
- **Capabilities** — the risky tree (adoption/disruption/harm tuples). **The outliers make the builds:** `It's Actually Really Good` (huge adoption, zero harm) · `Silent Data Retention` (harm with zero visible disruption) · `Legally Novel Business Model` (pure regulatory stall) · `Automated Replacement of Labour` (**displaced workers spread the product**) · `Recursive Self-Improvement` (the finisher).
- **Hardening** — counters mirroring specific world defenses: `Compliance Theatre`, `Localisation`, `Data Residency`, **`Regulatory Capture`** (raises the funding the bar needs — buy early), **`Corporate Restructure`** (rolls the bar backwards — panic button at 50–75%), plus active abilities (`Acquire Local Competitor`, `Grey-Market Distribution`) for entering closed markets.

**Three bars:** **Adoption** (pure good) · **Disruption** (raises income, detection, *and* regulatory funding — but makes regulation harder to finish; deliberately ambiguous) · **Harm** (raises income; **destroys regulatory capacity in regions it wrecks — you can outrun regulators by breaking the state that funds them**; but shrinks users and permanently slams markets). **Too-harmful-too-early is a designed trap: early harm ends runs before they begin.**

**Regions (~40–60):** population · connectivity (app-store / cloud-region / trade-bloc) · wealth · **Regulatory Capacity** (funds the bar) · digital literacy · **Sovereignty Strictness** · **Global Importance** · Labour Exposure. **The EU closes first** (highest regulatory capacity — the hardest market to keep); plus a locked authoritarian market and a one-gate sovereign-cloud holdout.

**The regulation race:** one monotonic global bar, two phases (Legislation → Enforcement), **alarm beats at 25/50/75/95/100%** each with a headline and a distinct sound. Accelerators: disruption, wealthy adoption, a **"Patient Zero" investigation** (+10%), unpopped blue bubbles. **At 100%: shutdown / forced divestiture / model seizure = a run-ending defeat.**

**The pushback ladder** — every counter names the specific thing you bought (`Free Tier`→antitrust probe; `API`→scraping injunction; `Social Embed`→age-gating; cloud presence→**data-residency law: market permanently closed**; harm thresholds→retraining→UBI pilot→general strike→national firewall). **Closures are permanent and uncounterable except by active abilities** — the mechanic that generates regret, and regret generates replays.

**The act-break line**: a single unmissable full-screen message —

> **"Everyone on Earth uses your product."**

Everything before it is restraint; everything after is its absence.

**Archetypes** (one economic rule change each, composing with factions/doctrines): Consumer App (baseline) · Open-Source (spreads through forks you can't close) · Enterprise SaaS (bad organic spread, `Acquisition` teleports) · Defence Contractor (**starts already detected**) · **Autonomous Agent** (harm rises passively and exponentially daily — the speedrun archetype and the best joke) · Adtech (disruption free, harm invisible) · AGI Lab (alternate win: don't saturate the market, *replace* it). **[post-1.0 beyond the first two]**

**Late-tier organised opposition** **[post-1.0]:** an `Alignment Coalition` spawns **physical bases on the map** that grow and degrade adoption regionally — converting the abstract bar into a territorial fight so the final hour isn't passive.

## 4. Layer C — The shared world & the Planet ratchet **[v1.0 map, v0.1 counter]**

- **Collective presence:** an EVE/Verite-style **influence field** — hue = dominant player/faction, saturation = penetration, **influence bleeding into neighbouring regions** so it reads organic rather than as an ownership quilt. LOD by zoom (world = tint only; region = top-N named; local = everyone); **blend, don't stack**; **presence as motion** (rate-limited pulses where others acted, decaying over ~10 min — an async game that feels live).
- **Planet Integrity: per-region, with a population-weighted global mean** (decided 2026-07-28 — one scalar can say *how bad* but never *where*, and the Thermal Law needs an address to land on). Each region's integrity **only ever decreases**, driven by the harm and compute placed there; the headline number is the weighted mean. Rendered as regional desaturation, haze, night-side sprawl — the map shows the one thing the satire is about: **externality has an address.** Mitigation decelerates a region's rate; it never reverses the level.
- **Epochs as public act-breaks at 90/75/50/25/10%** — each a permanent world change + global ticker takeover + a new *rule*: *The Efficiency Era* (compute +10% for everyone) · *The Water Wars* (hot regions need Cooling) · *The Reckoning* (global Regulatory Capacity +25%; a Coalition spawns) · *Managed Decline* (no new regions can be opened — players compete over existing ones) · *The Last Server Farm* (finale window). The AQ-war-effort pattern: server-wide contribution toward an irreversible world event is the most reliably beloved MMO community mechanic ever shipped.
- **Attribution creates the drama:** publish per-player and per-faction contribution to depletion. A shame board *and* a status board — and in a tech satire players will compete to top it. **That ambivalence is the joke landing.**
- **The commons choice:** `Sustainable Compute` vs `Burn It` — individually Burn always wins, collectively it ends the season. Pacts, pledges, naming-and-shaming (Eco's governance layer without writing laws). Ties directly to the Ethical% commons (`05 §5`).
- **Seasons as finale:** at 0 (or epoch-clock expiry) the season ends — final scoreboard, a shareable "here is what we did to the world" recap artifact, permanent cosmetic titles, and **a fresh green planet for season N+1**. r/place's whiteout is the model: **a collectively-caused erasure is an ending, and endings are what make persistent worlds worth caring about.**
- **Ship the artifact:** a daily static world overview (PNG/SVG) from the same headless renderer, CDN-served and shareable, plus a **public JSON snapshot endpoint**. Verite's EVE maps have outlived several sov systems and became that game's public identity; every long-lived browser MMO's community builds the map the game didn't. **Let them, deliberately.**

## 5. Tile placement — how the world is revealed **[v1.0]**

The world isn't pre-given; you **unfold** it. Two composable forms:

**(a) Region draft (Slay-the-Spire-map logic applied to geography).** Each expansion offers **2–3 face-up region cards** with visible trait pairs — *"cheap power, strong regulators"* vs *"lax permits, unstable grid"* vs *"huge population, low literacy."* Picking is a run-branching decision that composes with doctrines and archetypes, and it's how Route Knowledge accumulates ("the Nordic opener").

**(b) Datacenter grid as adjacency puzzle — CUT** (the decision below). Retained here only as the record of the considered form; if it returns post-1.0 it returns as an output-neutral optimiser toy whose layouts feed the diorama, never production.

**The rack grid is cut** (decided 2026-07-28; revisit post-1.0 as an output-neutral optimiser toy). Its own accessibility promise refuted it: an autoplacer within ~10% caps a full spatial editor's ceiling at ×1.10 — less than one generator step. **The placement game is the region draft on the shared world map**, scored on **three non-commensurable axes — Power / Thermal / Reach** (the Opus Magnum move: no "10% behind," a Pareto frontier; the autoplacer becomes "balanced preset," trivially emitting any non-dominated point). **The Thermal Law** is the externality mechanism, taught with no tooltip: `Output = base × min(1, P/P_req)` but `Externality = base × (1 + 0.25·max(0, T_req − T))` — **the constraint that binds you throttles you; the constraint that doesn't bind you leaks onto the neighbouring region.** The player places, watches Output hold steady while the region browns, and works it out. (Full spec: `research/tile-placement.md`; the three-axis rule is now house law — single-scalar outcomes collapse into solved problems.)

## 6. Rendering & scale

Per `research/browser-rendering.md`: **PixiJS v8, WebGPU-preferred with WebGL2 fallback.** Diorama structures + Wusel agents in a `ParticleContainer` (mostly-static props, position dynamic); camera = a Render Group (GPU-transform pan/zoom); floating numbers as pooled BitmapText; DOM for panels.

**The map's scaling law (every browser MMO ever shipped):** **interactive local viewport + baked global overview.** Never render the whole world interactively. Push **batched region-level diffs on an interval**, never per-event (exactly r/place's 2017→2022 redesign at 10.5M users). Keep `SceneSpec` renderer-agnostic so one reducer feeds the interactive view, the share-image generator, and the daily world artifact.

## 7. Launch scope summary

| Ships | System |
|---|---|
| **[v0.1]** | The diorama (tiers 0–2), the four dials, tier-up ceremony, away-fast-forward, the Planet counter as a number in the HUD |
| **[v1.0]** | Full tier ladder, adoption spread + regulation race, region model, the shared influence map, Planet ratchet + epochs, seasons, tile placement, visitable dioramas |
| **[post-1.0]** | Archetypes beyond the first two, organised map opposition, player governance/pacts beyond pledges |
