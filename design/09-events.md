# The Event Engine

> Three layers, one data format, one evaluator. Layer 1: personal narrative events (Paradox model). Layer 2: personal pressure meters (EU4 disaster model). Layer 3: server events — situations and Major Orders (CK3/Helldivers model). Plus the seasonal arc structure and GM operations. Full research: `research/events-playstyles.md`.

## 1. The data format

Paradox-shaped, hot-reloadable YAML/JSON; same schema family as balance data (`06-tech.md`). Core shape:

```yaml
id: outrage.10
namespace: outrage
type: personal            # personal | pressure | situation | major_order
scope: company            # company | founder | guild | world
theme: media_firestorm    # batches art/sound/voice
title_key: outrage.10.t
desc:                     # first_valid: parametric variation instead of new events
  - { trigger: { has_flag: denied_it }, key: outrage.10.d.denied }
  - { trigger: { pr_department: true }, key: outrage.10.d.managed }
  - { key: outrage.10.d.default }
trigger:
  all: [ { outrage_progress: ">= 40" }, { not: { had_flag: { flag: raided, days: 60 } } } ]
mtth:
  hours: 72
  modifiers:
    - { factor: 0.5, condition: { moderation_spend_ratio: "< 0.01" } }
    - { factor: 2.0, condition: { pr_department: true } }
cooldown: { days: 30 }
category_lock: { category: outrage, hours: 48 }     # global pacing lock
immediate: [ { save_scope_as: target_dc, pick: random_owned_datacenter } ]
options:
  - key: outrage.10.a
    flags: [dangerous]
    effects: [ { modify: { "trust.users.standing": -15 } }, { trigger_event: { id: outrage.11, delay: { hours: [12, 48] } } } ]
  - key: outrage.10.b
    trigger: { cash: ">= 50000" }
    show_as_unavailable: true
    effects: [ { spend: { cash: 50000 } }, { modify: { "trust.press.standing": +5 } } ]
  - key: outrage.10.c
    fallback: true
    effects: [ { modify: { uptime: -0.05, duration_hours: 24 } } ]
after: [ { clear_scope: target_dc } ]
```

**Required primitives** (priority order): boolean trigger DSL over game state · MTTH with multiplicative modifiers · `immediate`/`after` · options with per-option triggers, `show_as_unavailable`, `dangerous`/`exclusive`/`fallback` flags · `trigger_event` chains with randomized delay · `hidden` logic-node events (invisible dispatchers — branching logic in the same DSL as content) · `cooldown` + `fire_only_once` · weighted `random_events` pools **with a `0` = nothing bucket** · **category-level pacing locks** ("no event from category X in N hours" — the cheapest anti-fatigue device, from EU4 Shinto's 30-year gate).

**Dispatch model (CK3 lesson):** no free-floating scans — events attach to **on_actions**: pulses (`hourly_pulse`, `daily_pulse`, `quarterly_pulse`) and lifecycle hooks (`on_exit`, `on_tier_up`, `on_first_datacenter`, `on_milestone_reached`, `on_soul_below_20`, `on_founder_age_60`…). Content packs **append** to on_action lists, never overwrite — seasonal content is additive by construction.

**Anti-fatigue rules (from CK3's community postmortem):**
1. Text is SHORT. Two sentences beats two paragraphs on the ninth viewing.
2. Variation via `first_valid` desc blocks + scope substitution (which datacenter, which rival, which employee *by name*) — templates with slots, not more unique events.
3. Frequency budgeted conservatively; density complaints > quality complaints.
4. Category locks everywhere.

## 2. Layer 1 — personal events

The satire beats as choices: "a journalist wants a quote" · "your CTO is leaving to start a rival" · "the VC wants a board seat" · Faustian contracts (Soul) · the Sarcastaball event ("your sarcastic pitch gets funded") · media canonization arrivals (South Park episode, the Onion item, "Senator, We Run Ads", the John Oliver segment) with the **Ignore / Embrace / Sue** branch table and Streisand math (`08-satire-flavor.md §canonization`) · immortality-quest chain (the Sleep Test fires when you've been idle 6d — the Gilgamesh joke as an actual idle check) · conspiracy events fed by Layer 2.

Long personal arcs (CK3 story-cycle analogue): each run carries 1–2 **story cycles** — multi-event narratives attached to the founder (the rival's arc, the regulator's arc, the pet's arc) that survive across sessions and reference past runs (the game remembers your morality).

## 3. Layer 2 — pressure meters

EU4 disasters, reskinned. Visible, forecastable bars with contributing factors listed in the UI; the disaster is always a **consequence of an optimization the player chose** — this is where the satire bites.

Launch set:

| Meter | Fed by | Fires | While active | Ends |
|---|---|---|---|---|
| **Public Outrage** | moderation cuts, engagement optimization, scandals, Externality reveals | **The Raid** (the QAnon-raids-your-datacenter chain: RSVP counter 2,140,000 → 151 show up; someone "self-investigates" the lobby) | hiring +25%, ad revenue −15% | PR staffing + 90 quiet days |
| **Conspiracy Pressure** | growth, secrecy purchases (barges!), outrage spillover | theory-spawn events (5G/microchips/lizard tier scales with tier); mast-arson; Waymo-coning | ticker floods, small Press-Standing drain (`02 §7` constituencies) | monetize it (Gargoyle Pivot), make it true (late-game), or wait it out — **fighting it raises it** (suppression/debunking/silence all feed belief; published rule) |
| **Board Pressure** (VC faction) | missed milestones, idle runway burn | board coup event chain | equity squeeze | hit growth targets |
| **Regulatory Heat** | dark-pattern stages, egress fees, lobbying failures | inquiry → hearing ("Senator, We Run Ads") → consent decree | compliance tax | settlements, Clout-gated lobbying |
| **Burnout** (org) | crunch, on-call noise | attrition wave / unionization arc | output −%, Soul drain | rest policies, touch-grass |
| **p(doom)** (Tier 5+) | capability rushing, e/acc tree | the AGI-tier crisis chain | — | safety spending (slows you) |
| **Litigation** | Pirate 7 Million Books etc. | the $1.5B bill | — | pay it ($3,000/book, itemized) |

## 4. Layer 3 — server events

**Situations** (CK3 model): long-running, phased, world- or region-scoped modifier containers with their own event pools and end conditions. Phases **swap rules, not numbers** (Chip Shortage: GPU index ×3, training runs +50% duration; AI Winter: capability progress halved, safety cheap; The Antitrust Era: megacorp-tier taxes, breakup events). Advanced by **catalysts** — aggregated player actions worth ±points, published live. Participation tiers: Involved (full modifiers + unique options) vs Interloper (weaker effects) by engagement.

**Major Orders** (Helldivers model): time-boxed collective objectives with a deadline, a shared bar, the published impact-modifier formula, and **a real possibility of failure — failure is canon** and re-routes the season. Objective types: produce/provision X, hold a state until expiry, collectively resolve N incidents, unlock-race objectives (the milestone-gated tiers), defend objectives (a 24–48 h no-decay defense window with a **gambit** alternative — solve the source instead of the symptom).

**GW2 hearts principle:** alongside server events, every player always has a personal objective board (rotating always-available goals contributing to the current arc) — the content floor when no world event is live.

**Rule-break events (rare, memorable):** once-per-season WoW-Zombie-style 72-hour rule inversions (e.g. "The Rollback" — everyone's UI reverts to Tier-0 1995 chrome for a weekend; "Leap Second" — all clocks run 2× for 24 h). Divisive by design, remembered forever; always opt-out-able for accessibility.

**One-time moments (Fortnite The-End model):** season finales are synchronized, non-repeatable, timestamped. Absence-as-content is allowed (the servers "go dark" for the finale's final minute, once).

## 5. The exchange shop (the reward interface)

> From `research/liveservice-idle-tier.md §3`: the live-service tier's own anti-lootbox valve, adopted as our **only** event-reward primitive.

- Every event pays **event scrip**; the event's shop lists rewards at **posted prices**. No RNG between effort and reward, anywhere in the event system.
- **Leftover scrip auto-converts** to a standard currency at event end at a published rate. Nothing expires worthless; there is no "you were 40 short" ending.
- Curtain-pull line ships with it: *"Company scrip has historically been a scam. Ours converts. We checked with history."*
- Cadence primitive for solo-dev sustainability: the **weekly mirror calendar** (`02 §10`) runs on this shop with rotating stock — rhythm without new content.

## 6. Seasonal arcs

- **Cadence:** ~3 months. Each arc = one new mechanic (trial) + one Situation (the narrative container) + 3–6 Major Orders (the story beats) + a finale.
- **Fold-in rule (PoE):** at season end, the trial mechanic is folded into the core permanently or retired. A failure-tolerant content pipeline; the community sees the decision made (and post-OSRS, we **poll direction, not implementations** — plurality, in-client, no supermajority veto).
- **Dispatches:** the narrative device — short in-fiction bulletins (HN front pages, TechCrunch headlines, leaked Slack screenshots) pushed server-side. Cheap, on-theme, tone-carrying. The news ticker is the dispatch surface; the flavor bible (`08-satire-flavor.md`) is its corpus.
- **Launch-year sketch:** S1 "Going Concern" (community unlocks Tier 4; the first Raid wave) → S2 "The Race" (Tier 5 unlock; Safety-vs-e/acc server tug-of-war — a Situation whose ending the players decide) → S3 "Discovery" (the antitrust arc; The Market opens) → S4 "The Quiet Part" (conspiracy inversion begins; first Ethical% world-first window).

## 7. GM operations

- **A human game master** (us) with a dashboard: decay/impact dials, dispatch composer, incident spawner, manual awards for edge cases. Budgeted as an ops role, not just tooling.
- **The war log:** every GM intervention is publicly logged in-fiction ("Head Office adjusted regional demand"). Opacity → meta-narrative; the Helldivers "the war is a lie" backlash is designed out by radical disclosure.
- **Watchdogs:** every event/situation has a max lifetime, a forced-resolution path, and an alarm (GW2's stuck-event bug is the #1 operational risk for this genre of system).
- **Telemetry-first thresholds:** all collective-objective numbers are set from measured throughput with uncertainty margins, tunable mid-event via the (logged) dials.


## Neopets plot adoptions (2026-08-01, `research/neopets-social-history.md §1`)

- **The plot-point prize shop** joins the contribution-reward toolbox alongside GW2 medals and
  the impact modifier: scored actions fill a plot wallet; a one-time shop opens at arc's end;
  players CHOOSE prizes (collectors/battlers/merchants buy differently); ops tunes economy
  injection by pricing after seeing participation. Adopted as the default reward surface for
  Layer-3 arcs.
- **Individual completion, communal solving**: at least one puzzle per major arc hard enough to
  REQUIRE community collaboration (their Lost Desert model) — the third contribution mode beside
  aggregate meters and personal rank; shared bewilderment is the social engine, and our
  no-free-text boards get structured "theory" posting surfaces for it.
- **Tiered war waves** (weak/medium/strong brackets) so every build contributes to fight-shaped
  arcs — already consonant with our contribution-window law.
- **The evergreen conversion** (their Altador model): each retired seasonal arc converts to a
  permanent self-paced quest — event content becomes onboarding content, the anti-FOMO answer to
  "I missed it."
- **One annual immovable ritual tournament** (their Cup ran 20 years including through the
  drought): identity-based team join, low-floor participation, support games for non-twitch
  players — and their 2026 flat-threshold reform (250/500/1,000 points, no All-Star grind
  ladder) is the version we adopt from day one.
- **The plot drought is the satire**: the in-fiction event pipeline gets "defunded" while the
  in-fiction cash shop ships weekly — played straight, per the enshittification arc.


## The Black Sunday doctrine (adopted 2026-08-01, `research/kol-puzzle-pirates.md §A2`)

KoL's 2004 hyperinflation exploit was fixed WITHOUT a rollback: an in-fiction collections agency
(the Penguin Mafia) confiscated bug-currency, and pre-authored vanity sinks — including a
1-billion-meat item that was "a completely useless accessory meant only to be a symbol of
shameful prestige" — drained the rest over months. Adopted as our economy-incident doctrine: a
pre-authored **"Billing Anomaly" event kit** (in-fiction collections NPC + shameful-prestige
sink SKUs) ships WITH the economy, so the answer to a live exploit is narrative + sinks, with
rollback as the last resort the deterministic ledger makes possible but the community never
prefers. Gaia's counter-example (sink events as theater while faucets kept selling) defines the
failure mode: our incident sinks must never coexist with an open faucet.
