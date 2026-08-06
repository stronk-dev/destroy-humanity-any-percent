# RFC: World Layer Foundation (community milestones · depletion · contribution)

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-03
- **Design refs:** `design/02 §1` (World layer — server-scoped, never resets: planet depletion ratchet, community milestones, Influence by contribution rank), `design/05 §a` (global milestones, Elite-Dangerous dual reward axis, throughput telemetry), `design/09 §3` (Layer-3 server events)
- **Research:** `events-playstyles.md §Contribution accounting` (the Helldivers Galactic Impact Modifier — contribution = raw × normalizer(share of playerbase) × diminishing(player share); paces identically at any CCU; PUBLISH the formula), `societal-satire.md` (the Jevons engine, depletion as the ending's spine)
- **Depends on:** Gameserver Composition + Transport (implemented — the world aggregator GC2 already publishes world snapshots; this gives them their source systems); Production (implemented — contribution derives from committed production)
- **Owner ruling honored:** breadth-first — the world MECHANICS (milestone counters, depletion ratchet, the contribution formula), not the specific milestones or the ending content.
- **Planning:** `planning/world-layer-foundation/` (once implementing)

## Summary

GC2's world snapshot publishes `planet` and `milestones` fields that currently read zero — because
their source systems don't exist. This RFC builds them: the server-scoped never-resetting world
state, community milestone counters with the published contribution formula, and the depletion
ratchet that is the game's ending spine. This is the layer that makes it an MMO rather than
parallel solo games.

## Specification

### WL1 — World state (server-scoped, never resets)

A single world record (server singleton, its own stream): `{planet_depletion_ppm,
active_milestone, milestone_progress, epoch_beat, contribution_ledger_ref}`. Never resets across
any player's Exit or prestige (the §2 §1 law — the World layer is the one thing no reset touches).
Mutated by a server-side aggregator (the GC2 world aggregator, extended) consuming committed
production events — never by client report. The aggregator is the ONLY writer.

### WL2 — The contribution formula (published, the Helldivers law)

A player's contribution to a milestone = `raw_output × normalizer(active_playerbase_share) ×
diminishing(player_share)` — the Helldivers Galactic Impact Modifier, which **paces identically at
200 or 200,000 CCU** (the property that makes community milestones work at any population). Ranked
by PERCENTILE, not absolute (idle output spans orders of magnitude). **The formula is PUBLISHED**
(the transparency law — Helldivers' backlash was hiding it): it's a generated formula artifact, on
the public API (A3). Contribution is computed server-side from committed production, banked into
the world contribution ledger per Elite-Dangerous's dual axis: your personal rank-payout AND the
global tier unlock.

### WL3 — Milestones & the dual reward axis

A community milestone is a catalog object: `{id, target, contribution_metric, personal_rewards:
[by percentile band], global_unlock: {content_ref}}`. Reaching the target unlocks global content
for EVERYONE (the content-cadence mechanism — design/05's answer to "how do new tiers open");
personal payout scales with your contribution percentile (the dual axis — solves leeching and
free-riding at once). **Throughput telemetry is mandatory** (the CoC 48-hour lesson): the
milestone declares expected-completion-time bounds and the aggregator alarms if a milestone will
blow through its tiers — uncertainty designed into the final tier, never a hard ceiling that
embarrasses.

### WL4 — The depletion ratchet (the ending spine)

`planet_depletion_ppm` only ever increases (the Jevons engine: every efficiency upgrade increases
TOTAL consumption — `societal-satire.md`'s mechanically-novel-and-accurate satire). It's fed by
aggregate server production (the externality that leaves your ledger and lands on the world — the
Externality meter's server-scope sibling). It never decreases (accumulation is the point; it's the
spine of the three endings). Server-scoped, published, a world dial. Phase-0 foundation ships the
ratchet mechanic; the ending content (design/08's three endings) is a later RFC consuming it.

### WL5 — Influence (contribution-paid, not spendable-currency)

Influence is the World-layer reward currency, EARNED by contribution rank, spent on world-surface
participation (late-tier lobbying/capture events — design/08's billionaire layer). It's a real
spendable World currency (unlike the moral meters) but minted ONLY by contribution (a single-mint
discipline, the Clout/Gaia pattern applied to the World layer). Phase-0 declares the mint + the
ledger; spend surfaces are later content.

## Acceptance criteria

1. World singleton: never-resets across a founder's Exit (fixture: Exit leaves world state
   untouched); aggregator is the sole writer (grep-proven, no client path).
2. Contribution formula: the Helldivers modifier byte-exact, PUBLISHED as a formula artifact on
   the public API; percentile ranking; population-invariance property test (same pacing at 10×
   playerbase with proportional output).
3. Milestones: dual-axis payout (percentile personal + global unlock); throughput telemetry
   alarms on a seeded blow-through; global unlock gates content by fact.
4. Depletion ratchet: monotonic increase (never decreases across any interval); fed by aggregate
   production; published as a world dial.
5. Influence: single-mint (contribution only — seeded second source rejected); earned-by-rank;
   the spend ledger exists (surfaces deferred).

## Open questions

- Cohort vs global milestone scope (recommend: both, per design/05 — cohort milestones as the
  onboarding-scale version, global as the headline; the aggregator handles both).
- The GM/live-ops dial for Layer-3 server events (design/09 §3) — recommend a successor RFC; this
  foundation ships the milestone mechanic the GM would operate.

## Changelog

- 2026-08-03: created (draft) — the World layer mechanics; gives GC2's zero-valued snapshot fields
  their source systems; the published-contribution and depletion-ratchet laws structural.
