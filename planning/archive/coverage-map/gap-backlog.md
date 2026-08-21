# RFC Gap Backlog — designed systems with no foundation (VALIDATED 2026-08-05)

> **FROZEN HISTORICAL SNAPSHOT — NONCANONICAL.** These dependency findings describe the
> 2026-08-05 repository and are not the current executable queue. Read
> `planning/platform-alignment/execution-queue.md` before acting.

The ~28 systems at furthest-stage **D**, ordered into waves by the foundation they build on, so an
RFC is drafted only once its substrate is landing. Generated from the validation sweep's GAP flags
(`../coverage-map/validated/*.md`). Waves are draft-order, not a schedule.

Legend: **[now]** substrate already implemented, draftable today · **[soon]** blocked on an
*implementing* foundation that will land shortly · **[later]** blocked on a *draft* foundation or
on upstream content.

---

## Wave A — draftable NOW (substrate implemented)

These need no foundation that isn't already archived or shipping. Highest leverage: they unblock
downstream content and were being wrongly treated as "not yet reachable."

| System | Builds on (all satisfied) | Status (2026-08-05) |
|---|---|---|
| **Fiscal Quarters** (sugar-lump wall-clock meta-currency) | Save + Run Genesis + Founder Attendance | ✅ **DRAFTED** — `rfc/fiscal-quarters-foundation.md`; queued for Codex acceptance review. Scope (Founder v19 vs Company) flagged as its first ruling. |
| **Active-play buff windows** (Lucky-bank formula, golden-opportunity, multiplicative combos) | Production + Save (archived) | ✅ **DRAFTED** — `rfc/active-play-buff-windows.md`; click clamp + timing-skill confirmed; daemon + rhythm-timing declared as successors. Queued for acceptance review. |
| **Doctrine + Compute-Credit spend** (already a draft RFC) | All substrate implemented | ✅ **dependency-ready** — status updated, queued for Codex acceptance review. |
| **Relevance Harness** (already a draft RFC) | Balance Harness + Run Genesis (archived) | ✅ **unparked** — stale Purchasable-Content blocker cleared in the RFC; queued for acceptance review. |
| **Soul** (dedicated RFC — ruled 2026-08-05) | Founder scope + events (drain triggers) | ✅ **DRAFTED** — `rfc/soul-foundation.md`; meter-shown/currency-triggered hybrid + opportunity-costed recovery + graduated gating; drain-source catalog gated on the events engine (successor). Queued for acceptance review. |

## Wave B — blocked on an *implementing* foundation (draft as it lands)

| System | Blocked on (implementing) | Notes |
|---|---|---|
| **The individual minigames** (Server Garden, chess, board suite, the Market, Ship-It Spellbook, Advisory Board, Incident Response, Terminal Typer, Demo Disc Arcade, Shipping Wars, parlor-taxonomy expansion, + new candidates below) | Minigame Platform | Platform ships a **test-only fixture tenant** — zero real games. Each is a content-RFC on the platform. **Wave-B drafting started 2026-08-07: “The Pitch” + Soul Recovery Activities are DRAFTED (rfc/). Research-backed prototype order (2026-08-05):** ① **"The Pitch"** (Balatro-like score=chips×mult deckbuilder — turn-based, deterministic, PvE, and it *is* our multiplicative buff-stack in the big-number regime → closest to skinning the shipped engine); ② **"The Tech Stack"** (Slay-the-Spire-like deckbuild-a-run); ③ rhythm minigames (Build Cadence, etc.) built on the active-play 20 Hz-tick timing model; real-time survivor-likes are the expensive trap (if any, Brotato's wave-boundaries give certification checkpoints). See `roguelike-survivor-minigames.md`, `rhythm-timing-games.md`. |
| **Pet-battle content** (rosters, seasons) | Combat duel/lane/bots engines (+ pet care) | Mechanics are the combat engines' job; this is the content that rides them. |
| **Media canonization / Clout-milestone events** | Clout mint (Achievements/Feed) — **not minted yet** | `docs/achievements.md` confirms Clout isn't minted. South-Park/Onion/Streisand ignore-embrace-sue branches. Also needs the events engine. |
| **Ideology / enshittification toggles** | Meters artifact | `design/10 §3`. The opt-in visible-degradation-for-multiplier playstyle switch. |
| **Soul** (dedicated RFC — ruled 2026-08-05) | Founder scope (`ApplyFounderLogged`) + events (drain triggers) | `design/02 §8`. **Its own foundation RFC**, not a meters sub-field: drain verbs (crunch, Faustian VC term-sheets), recover verbs (touch-grass, cost-time-produce-nothing), gates human content (pet recognition, hobby locks) + the transcendence ending. Un-orphaned. |
| **Loot boxes / gacha · hats/cosmetics · breeding & rarity** | Purchasable-content (done) + pet-care + game-UI | The monetization-satire content on the substrate. game-UI U3 explicitly excludes any cosmetic system beyond the Horse Armor stub. |
| **Trading — NPC/aggregate market** (ruled 2026-08-05) | Economy + transport + purchasable-content + world-layer | Scoped to an order-book vs an aggregate/NPC counterparty + the anti-RMT non-transferable flag (research-backed). **Architect P2P-ready** (order/settlement model + non-transferable flag + anti-cheat boundary built so a P2P layer needs no rewrite). P2P itself = a later expansion, gated on its research dossier (dispatched 2026-08-05). |

## Wave C — blocked on a *draft* foundation, or late-arc content

| System | Blocked on | Notes |
|---|---|---|
| **Layer-2 pressure-meter disasters** | Meters + Layer-1 events evaluator (draft) | EU4-disaster model; the evaluator explicitly scopes disasters OUT today. |
| **Layer-3 server story arcs** (Helldivers Major Orders) — *highest narrative leverage* | Layer-1 evaluator + world-layer + feed/dispatch + leaderboards | The content-cadence engine. Layer-1 RFC explicitly defers Layer-3. |
| **Lobbying / Influence-spend** | world-layer WL5 (Influence ledger) + **its research dossier** (ruled 2026-08-05: research-first, dispatched) | Entirely unbacked; no RFC until the dossier lands. |
| **Conspiracy system** | Meters registry growth (its Pressure meter is NOT in the 11-row catalog) + events + Clout | Design maps line-for-line onto `conspiracy-media.md`. |
| **Speedrun chrome** (run-title bar, splits/PB/gold-splits, TAS mode) | Leaderboards + UI + Prestige (+ Tier 7 for TAS) | The numeric spine (Leaderboards) exists; this is the presentation layer over it. |
| **Tiers 2–8 content ladder** | Each an epoch on purchasable-content + doctrines/meters/minigames | Only T0–T1 is drafted (garage). **The largest single content program** — one content-RFC per tier/epoch. |
| **The endings** (three-endings / both-datasets) | Full tier ladder through Tier 8 + Prestige + world-first + Ethical% | Ending C is strictly last. |
| **Bakery Inc. easter egg** | Tier-3 market/tenant sim + Minigame Platform + Tier 7 | The Cookie-Clicker homage-as-tenant. |
| **GM operations · event-scrip shop · seasonal-arc content** | Events engine (all layers) + world-layer | The live-ops surface over the events engine. |

---

## New minigame candidates from the 2026-08-06 research batch (see `research-integration.md`)
- **Cozy / touch-grass (Soul-recovery, zero-reward):** Defragment the Disk, The Server Room, Repot
  the Plants — produces-nothing = zero-cheat = trivial server/AI-fallback.
- **Board games:** Vertical Integration (engine-builder — maps onto the buff-stack core), The Draft
  (best multiplayer fit), Capacity Planning (Azul — sharpest satire). Social deduction is
  NON-COMPLIANT (do-not-build the hidden-role form; PvE puzzle only).
- **Pet depth:** a care-history memory ledger (one small Founder-save field, reuses the C18 primitive)
  → candidate RFC; deepens the "real fake personality" the Soul-greying leans on.
- Full routing + the 5 recommended next research commissions (factory/automation, M&A absorption,
  crypto/web3, labor, 4X) are in `research-integration.md`.

## Draft priority (what to write first)

1. **Wave A, all four** — pure upside, substrate already done; two are just status promotions
   (doctrine+compute-credit, relevance-harness) and two are clean new content-foundations
   (quarters/earnings, active-play windows).
2. **Minigame content-RFC template** — because ~11 GAP rows share one shape, draft ONE exemplar
   minigame RFC against the platform (chess or Server Garden) as the pattern, then the rest
   replicate. Biggest content unlock per unit of design effort.
3. **Clout mint** — it gates media canonization AND conspiracy AND lobbying reach; minting it
   converts three Wave-B/C narrative systems from blocked to draftable.
4. **Per-tier content-RFC template** (Tiers 2–8) — the largest program; needs one exemplar
   (Tier 2) to establish the epoch-content shape before scaling.

## Open scope questions — status (see `decisions-log.md`)
- **Trading:** ✅ RESOLVED — NPC/aggregate market first, architected P2P-ready; P2P deferred to its
  research dossier (dispatched).
- **Lobbying/Influence-spend:** ✅ RESOLVED — research-first; dossier dispatched, no RFC until it
  lands.
- **Soul:** ✅ RESOLVED — its own dedicated foundation RFC (drain/recover verbs, Founder scope).
- **World-layer dual-reward axis:** ✅ RESOLVED 2026-08-07 (Claude) — CONFIRMED, not down-scoped.
  The mechanic is directly research-backed: Elite Dangerous Community Goals (personal payout by
  contributor rank + global tier unlock) is the named model, and GW2's percentile medal tiers +
  the per-faction-percentile contribution note back the payout shape
  (`events-playstyles.md:219,338-339,542`). WL3 stands as specced. One research refinement
  recorded for the Layer-3 events successor (not the milestone foundation): failable objectives
  pay ~half on failure ("everyone who touched it gets something") — foundation milestones have no
  deadline/failure state, so the clause attaches to Layer-3 server events when they're drafted.
