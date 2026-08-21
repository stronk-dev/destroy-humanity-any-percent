# Deferred & Dropped Elements — explicit tracker

**Currentness check:** 2026-08-21. This is a maintained decision ledger, not an execution queue.
Only an owner decision plus the normal design/RFC route can revive a row.

Struck or deferred gameplay elements, kept here so they are **decisions we can revisit**, not holes
that quietly vanish. Each carries: why it was cut, what replaced it (if anything), and the
**revival condition** — the specific shape under which it could come back. A row here is NOT a gap
in the frozen historical `planning/archive/coverage-map/gap-backlog.md` (those were designed-and-
wanted-but-uncontracted at the snapshot date); a row here was deliberately
removed and needs a decision to return.

| Element | State | Why | Replaced by | Revival condition |
|---|---|---|---|---|
| **Async base building & raids** (Clash-style offline base defense) | 🟡 DEFERRED — *reclassified 2026-08-05 from REJECTED* | Originally struck 2026-07-28 (`ba1dbf0`) as a "no-FOMO" violation. The no-FOMO rule was **refined 2026-08-05**: the constraint is no net *loss* for being offline, not no offline interaction. The naive form (real defender loss) is still out; a **loss-decoupled** form is now viable. Secondary reasons stand: base *layout* is a solved puzzle players outsource (82% of outcomes, one optimum, wiki-copy +41pp), and it's a lot of system for a solo dev. | The **Lane pusher** + the **cosmetic house** cover the salvaged value today. | **Revivable in loss-decoupled form** — the Clash "Clash Anytime" clone-base model: an attacker raids a *clone* of your base and earns rewards, but you (the defender) lose **no** resources/standing (`lane-pusher-design.md §316` has the precedent). Pair with an *active* build/defend loop, not a persistent wiki-solved layout. Kept out of the foundation program; a candidate for the minigame content wave. |
| **Simplified base building + tower-defense minigame** (active) | 🟡 DEFERRED — kept out for now, tracked at owner's request (2026-08-05) | Owner intuition: at some point a *simplified* base-build + active-defense TD minigame may earn a slot. Held out of the current foundation program deliberately. | — (the "Desktop Tower Defense" slot below is its natural home) | Revisit as an **active** minigame (player present and defending in real time — no offline victim), on the **Minigame Platform** foundation, as ONE of the ~11 minigame content-RFCs. Cheap bot AI (TD, not Clash attack AI). Keep it simple: build + defend a wave, not a persistent raidable base. |
| **Desktop Tower Defense** (active base-defense minigame slot) | 🟡 HOMELESS — genuine gap, no owner | Was the *active* mode paired with base-raid defense (`design/04`); lost its home when raids were cut 2026-07-28. Research verified (HandDrawnGames, 2007; steal Bubble Tanks TD merge-to-grow). | — | This is the concrete home for the simplified-base-build/TD idea above. Draft as a Minigame-Platform content-RFC (T2–T3) whenever the minigame content wave starts. |
| **The punch-down multiplier** (200% payout punching up, 5% four tiers down) | ✅ HOMED 2026-08-06 | The single best anti-bully mechanism found in any researched game (pure incentive gradient, zero moderation cost). Its original home — async raid matchmaking — was rejected 2026-07-28. | A **shared PvP-payout rule** across all PvP surfaces (lane-pusher, M&A Arena acquisition payout, pet-battle ranked). | Resolved via `absorption-arena.md`: payout-by-relative-size beside antitrust (biggest-cell-shattered) + diseconomies-of-scale (mass decay). Bind it in whichever PvP RFC lands first — a reusable rule, not a per-RFC invention. |
| **Soul-as-a-meter** | ✅ RESOLVED → implemented/archived | Meters RFC had demoted Soul from a first-class meter to a read-only int64 carry. Owner ruling gave Soul its own foundation and recovery contracts. | Archived Soul Foundation plus Soul Recovery Activities and canonical docs. | Closed; any further change follows a successor RFC rather than reviving this row. |

## Status
- 2026-08-05: created at owner request. Base-building/raids reconfirmed rejected-as-offline-raiding
  but a **simplified active TD form is explicitly on the deferred list** with the Desktop-TD slot as
  its home. Revisit during the minigame content wave, not the foundation program.
- 2026-08-21: currentness sweep preserved the deferred gameplay calls, updated Soul to its archived
  implementation state and confirmed that this ledger authorizes no current batch.
