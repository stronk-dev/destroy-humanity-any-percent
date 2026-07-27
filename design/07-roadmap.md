# Roadmap

> Build order, with the release schedule doubling as fiction: shipping cadence = in-game events. Every phase ends with something playable.

## Phase 0 — Foundations (engine core)

- Go server skeleton: accounts (anonymous + JWT upgrade), saves (versioned from commit #1), closed-form production engine, intent API, offline calculation.
- Go `Decimal` + TS `break_infinity` 2.2.0 golden-vector suite (the first tests in the repo). ✅ shipped — RFC-0001.
- Svelte 5 client shell: sim loop, tab rendering, number formatting, save sync.
- Balance data files + hot reload; the balance simulation harness (headless strategy runs — this gates every later phase).
- **Exit criteria:** a private build where Tier 0–1 plays end-to-end (click → generators → first Exit) and the harness reproduces the pacing targets.

## Phase 1 — v0.1 "The Garage" (first public)

- Tiers 0–2 complete: cost-curve economy, headcount allocation, first Exit loop, Reputation tree v1.
- Fiscal Quarters/Earnings Calls; Clout v1 (achievements + PR Interns); the cosmetic shop ($0.00, Horse Armor).
- Demo Disc Arcade, Terminal Typer, Server Garden; pet adoption + care (cattery port).
- Presence feed + global counters (the MMO's ambient layer, cheap and early).
- Era UI for tiers 0–2; ticker/dispatch system v1 with the launch corpus slice.
- **Fiction:** "a small business opens its doors."

## Phase 2 — v0.2 "Incorporation" (the MMO turns on)

- Factions (the four core), faction interdependence exchange, guilds + tithe + guild upgrades.
- First **community milestone** (Tier 3 content behind it — deliberately modest stakes to calibrate telemetry and the impact modifier in production).
- Tier 3: cloud market sim, enshittification slider, daemons, Bakery Inc. easter egg, Advisory Board, Ship-It Spellbook, Incident Response.
- Event engine Layers 1–2 (personal events + pressure meters: Outrage, Board Pressure, Burnout).
- **Fiction:** "Series A." First seasonal arc begins (`S1 Going Concern`).

## Phase 3 — v0.3 "Hyperscale"

- Tier 4 (community-unlocked): spatial datacenter play, constraint resources, Externality, the water counter, The Market.
- Base building + async raids; board-game suite ranked queue + matchmaking + bot backfill (the vs-AI engine's public debut).
- Event Layer 3 (Situations + Major Orders), GM dashboard + public war log.
- Speedrun chrome (timer, splits, categories v1, leaderboards).
- **Fiction:** `S2 The Race` — the community Major Order arc unlocks Tier 5.

## Phase 4 — v0.4 "Frontier"

- Tier 5: research tree, Safety/e/acc, corpus & model collapse, conspiracy layer + canonization events, casino suite, pet battles ranked season.
- The commons + Ethical% route; challenge runs v1; Compute Credits (time banking).
- Founder aging + longevity ladder + Soul-gated content complete.
- **Fiction:** `S3 Discovery` (antitrust arc).

## Phase 5 — v1.0 "Transcendence"

- Tiers 6–8: policies, TAS mode, the billionaire layer, Depletion, **all three endings + variants** (LAN Party, training-data, Ethical%).
- Category/challenge full set; run-end retrospectives; the Honesty appendix.
- **Fiction:** `S4 The Quiet Part` — first Ethical% world-first window; v1.0 is "the game now ends."

## Ongoing (post-1.0)

- Seasonal arcs on the ~3-month cadence with the PoE fold-in rule; the community-milestone pipeline as the content channel.
- Later factions (Crypto, Regulated Utility, Consultancy); minigame additions (one per season, aiming where Cookie Clicker never got: a minigame for *every* building eventually).
- Maia chess sidecar if chess is a hit; Discord embedded guild view (deliberate-open only); mobile polish pass.

## Sequencing principles

1. **The balance harness gates every phase** — pacing targets are acceptance tests.
2. **Multiplayer ambience early, multiplayer stakes late** (feed/counters in v0.1; milestones only after telemetry exists — the Clash of Clans rule).
3. **Every phase ships a satire beat** — the tone is load-bearing from v0.1, not a coat of paint at the end.
4. **Verify-before-ship:** each release's flavor content runs through the research files' confidence lists.
5. **No date promises in-fiction we can't keep out of fiction** — seasonal arcs are scoped to what one developer (plus the GM hat) sustains.
