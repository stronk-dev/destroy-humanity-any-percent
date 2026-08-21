# Master Coverage Map — VALIDATED 2026-08-05

> **FROZEN HISTORICAL SNAPSHOT — NONCANONICAL.** Preserved for reconstruction evidence.
> Its status rows do not authorize work; current capability and queue truth lives in
> `planning/platform-alignment/`.

> **STALENESS STAMP (record audit, 2026-08-16).** The validation below is 11 days old and the
> counts have MOVED. Do not quote its headline numbers without reading this stamp. Landed since:
> the **Relevance Harness** and **T0–T1 Playable Content** archived (I: 14 → ~16); **epoch 7
> (`T0-T1 Playable Content`)** and **epoch 8 (`First-Hour Payoff`)** minted, so shipped content is
> now 9 generators / 10 upgrades across **tiers 0–1 of 9**, plus **1 of ~12 minigames** (The
> Pitch); AC0's session-boundary offline catchup built; the prestige loop given its first
> mechanical payoff (branched first-company endings + run-2 starter packages). The **~28
> design-only backlog is essentially unchanged** — that half is still the larger half and is what
> 1.0 turns on. A full re-validation sweep of the six domain slices is OWED; this stamp is a
> delta, not a re-validation.

Synthesized from the six independently-reconstructed domain slices in `validated/`. Each row's
evidence lives in its slice file; this table records the **furthest stage** and flags. Regenerate
from the slices — do not hand-edit rows.

Stages: **R** researched · **D** designed · **F** RFC contract exists (draft/implementing/accepted)
· **I** implemented + archived + `docs/`.

## Headline counts (distinct systems, deduped across domains)

| Furthest stage | Count | Was in my 2026-08-05 grep recap |
|---|---|---|
| **I — implemented & archived** | **14** | ~14 ✓ (accurate) |
| **F — has an RFC (implementing/draft)** | **~25** | ~23 (roughly right) |
| **D — designed, NO RFC (the gap backlog)** | **~28** | ~15 ✗ **undercounted** |
| Total distinct systems | ~67 | ~52 |

**The one big correction to my earlier answer:** the design-only backlog is ~**28**, not ~15. I
undercounted by lumping — "minigames" is really **~11 individual games**, and "narrative" is
**~9 distinct systems** (conspiracy, media canonization, speedrun chrome, endings, tiers 2–8,
Layer-2 disasters, Layer-3 arcs, GM ops, event-scrip shop). The foundation side was roughly right.
Net: **breadth-first foundations are ~55–60% contracted; the content that rides on them is the
larger, still-uncontracted half — exactly the intended sequencing.**

## I — Implemented & archived (14)

numeric core · economy kernel · production/idle + offline · save layer · factions · gate-predicates
/ Route Registry (+ the "skips as discovered content" backbone) · gameserver composition · copy
pipeline · client shell & sim loop · balance harness · run genesis & replay · **Commons**
(structural core only) · **Guilds** (structural core only) · **purchasable-content substrate**
(implements the §11b tier-relevance doctrine — NOT monetization content).

> Two "I" rows ship only their Phase-0 core: Commons' player-facing half (cohort UI, incorporation
> line-item, tithe ballot, guild-Health binding) is F-draft (`commons-onboarding-and-governance`);
> Guilds' headline social features (Break Room, events/windows, matchmaking grades, leader
> upgrades, public exchange) are a deferred content backlog. Purchasable-content ships 3 placeholder
> generators — **zero satire content loaded**.

## F — RFC exists (≈25)

**Implementing (12):** prestige/exits · meters · achievements · leaderboards & epochs · minigame
platform · combat data model · pet care · websocket transport · account/session bootstrap · API
foundation · CI · founder-attendance.
**Accepted, impl-blocked (1):** UI foundation (C9–C11).
**Draft (12):** doctrine + compute-credit · relevance harness · world layer · feed & dispatch ·
commons-onboarding · combat duel · combat lane · combat bots · events-engine Layer-1 · game-UI
screens · deployment · T0–T1 playable content.

> Cross-cutting: **4 "implementing" systems ship present-tense `docs/` while un-archived** (prestige,
> leaderboards, meters, achievements) — counted F, not I, because none has cleared the
> independent-review-before-archival gate. Transport & account bootstrap are effectively
> implemented-but-unarchived. **Everything's final gate converges on the owner-gated push**
> (deployment declares every archived guard "advisory until the push"; CI can't complete its gate
> without it).

## D — designed, no RFC (≈28) — see `gap-backlog.md` for dependency ordering

- **Economy (4):** Soul drain/recover mechanic · quarters/earnings calls · ideology/enshittification
  toggles · active-play buff windows (Lucky bank / golden-opportunity / daemons).
- **MMO (3):** trading / player market · Layer-3 server story arcs · lobbying / Influence-spend.
- **Minigames (~11, all on the platform):** Server Garden · chess · board suite
  (connect-4/othello/gomoku/tic-tac-toe/checkers) · parlor-taxonomy expansion · the Market ·
  Ship-It Spellbook · Advisory Board · Incident Response · Terminal Typer · Demo Disc Arcade ·
  Shipping Wars (also on lane engine).
- **Collection / monetization satire (4):** loot boxes / free gacha · hats / cosmetics / collection
  book · breeding & rarity · the full dark-pattern parody suite. (Horse Armor exists only as a
  hardcoded stub shelf item.)
- **Narrative (6+):** conspiracy system · media canonization / Clout-milestone events · speedrun
  chrome (title bar / splits / PB / TAS) · the endings · tiers 2–8 content ladder · Bakery Inc.
  easter egg · (+ Layer-2 disasters, GM ops, event-scrip shop, seasonal-arc content).

## Validation findings (contradictions the sweep surfaced)

### Corrections to my earlier recap
- **`base building & raids` is DROPPED, not pending** (struck 2026-07-28; replaced by Lane engine +
  cosmetic house). Not a gap. Do not reintroduce.
- **Design-only backlog is ~28, not ~15** (see above).

### STATUS-OFF — `rfc/README.md` index is substantially stale → FIXED this pass
- Active table omitted **11** existing RFC files (founder-attendance, minigame-platform, pet-care,
  doctrine+compute-credit, events-engine-layer1, feed-and-dispatch, world-layer, game-ui-screens,
  deployment, relevance-harness, t0-t1-content).
- Line 69's "not yet drafted" list named 5 systems (Layer-1 events, doctrine, Compute Credit,
  game-UI, deployment) that **all now have draft files** (2026-08-03).
- Reconciled in the README rebuild committed alongside this map.

### DRIFT — design/RFC bodies contradicting accepted rulings (reconcile pending)
1. **`design/02 §6` line 148 says "every achievement grants +4 Clout"**, contradicting the same
   section's §6b Gaia one-mint law (line 160: "Clout has exactly ONE mint… nothing emits Clout
   outside") — and the Achievements RFC correctly mints NO Clout. The design body still describes
   the pre-ruling model. **Reconcile once the Achievements RFC's replacement axis (what achievements
   actually feed) is confirmed — a design-body edit, not a word swap; routed, not yet applied.**
2. **Meters RFC still frames meters as Save v15** (M5/C13), while the shipped Company axis rests at
   v14/v16. Needs verification: is v15 a real transient in the atomic meters+achievements→v16
   co-activation, or stale? **Verify against the meters impl before editing** (do not blind-swap
   v15→v16 — v15 may be a correct intermediate).
3. **Soul is orphaned:** the Meters RFC demoted Soul from a meter (design 02 §7) to a read-only
   int64 carry, and the Trust↔Soul CI correlation gate is unbuilt. Soul's drain/recover mechanic has
   no home — a real GAP (in the backlog), plus a recorded design-vs-RFC divergence.

### UNBACKED — design sub-mechanics with no research dossier (flag before shipping)
- World layer's **Elite-Dangerous dual reward axis** cites only the design doc, no dossier (the
  Helldivers modifier + Jevons ratchet it builds on ARE backed).
- **Lobbying / Influence-spend** is entirely unbacked — no dossier specs a doctrine/lobbying
  mechanic (`map-attraction §4b/c` specs world-*regulation*, not lobbying).
- Soft: Guilds' auto-tithe / leader-upgrades / contribution-windows are research-*mentioned*, not
  dossier-specced. Trading's player-to-player leg is unbacked (research backs an NPC/aggregate
  market + a strongly-backed anti-RMT non-transferable flag).

### Naming fix for the backlog
- Monetization mechanics source from **`gaming-enshittification.md §6`** at all tiers;
  `billionaires-decay.md` carries only late-tier *era flavor*, not monetization mechanics.

### Clean bills
No ORPHAN (all research reached ≥ D) except the §6 parody suite stranded at the F boundary. No
UNBACKED/DRIFT in infra or collection domains beyond the above. R→D fidelity is otherwise honest;
the conspiracy/canonization designs map near line-for-line onto their research.

## Status
- 2026-08-05: validated synthesis complete; README index rebuilt; gap-backlog generated. Two
  design-body reconciliations (Clout §6, meters v15) routed with a verification gate, NOT yet
  applied — next action.
