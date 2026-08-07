# RFC: T0–T1 Playable Content (the first hour)

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-03
- **Design refs:** `design/01 §Tier 0–1` (Sole Proprietor 1995 / Garage 2000s, era beats), `design/02 §2, §11` (curves, pacing targets: scripted failure ~15 min, first elective Exit [45,90] min), `design/03` (T0–1 arcade toys free, staggered unlocks), `design/08` (voice rules, era presentation), UX docs (first-session narrative)
- **Depends on:** **Purchasable Content Foundation (draft — the upgrade/role/chain/synergy mechanics this content declares; my original "no new engine code" claim was FALSE and is withdrawn)**; Relevance Harness (draft — lands between Foundation and this)
- **Closes:** the oldest line in `rfc/README.md`'s remaining-contracts list
- **Planning:** `planning/t0-t1-content/` (once implementing)

## Summary

The Phase-0 catalog has 3 placeholder generators. This RFC turns tiers 0–1 into a playable,
funny, correctly-paced first hour: the full T0–T1 catalog (generators, upgrades, gates, manual
actions), the first-session script, and the copy — all as data on existing schemas, gated by the
existing harness. One new catalog content epoch (mint) on the Foundation's mechanics. (Correction 2026-08-03: the original draft claimed "no new engine code" and "16 milestones" — both false; the Foundation RFC owns the engine work, and HEAD has 4 milestones × 4 persona-runs = 16 observations. The milestone set grows with this RFC's content.)

## Specification

### T0 — BLOCKING PRECONDITION (F1 from the Foundation review, 2026-08-03)

The mint activates `provision_tick_ms`, and online-mode `Evaluate` hard-fails past the offline-cap
horizon (`engine.go:153`) while the live service evaluates ONLY `ModeOnline`. A founder idle
> `accrual_cap_ms` (24h) bricks their stream permanently the moment provisioning is live. **This
RFC may not mint until online evaluation drives offline catchup at session boundaries OR clamps
the online horizon** — the fix lands here (or in session-bootstrap) as AC0.

### T1 — Catalog content (the mint)

- **Tier 0 (Sole Proprietor, 1995):** manual action `manual.click` re-skinned ("Reply to a
  Customer"); generators: `beige_tower` (exists — becomes real), `dot_matrix_queue`,
  `answering_machine`, `nephew_intern` (4 generator classes, cost curves in the 1.07–1.12 band
  per design/02, staggered milestone ladders per §11b). Upgrades: 8–12, each with ≥1 non-production-rate
  role (loader-enforced role law) — capacity on Permits, minigame token yield, stock-rate.
- **Tier 1 (Garage, 2000s):** `gate.t0_to_t1` gains real requirements; generators:
  `garage_rack`, `crt_wall`, `first_hire`, `beige_tower_v2` (chain-provisioning per the §11b
  purchased/generated split); the cosmetic shop appears with `Horse Armor (Free)` as its first
  item (design/01's beat, cosmetic-only, satire copy).
- Era-authentic **event copy** for the existing event kinds (threshold crossings, gate
  crossings, offers) — Layer-1 authored events are OUT of scope (the events-engine RFC owns the
  evaluator; this RFC writes only copy for events the engine already emits).
- All values provisional balance data; the epoch mint carries the first REAL relevance report if
  the Relevance Harness lands first (recommended order).

### T2 — The first session script

Deterministic sequence on existing machinery: boot → one manual click pays → first generator
affordable ≤ 30 s → first upgrade ≤ 2 min → `gate.t0_to_t1` crossable at ~8–10 min (chaos
persona) → **the scripted first failure fires ~15 min** (implemented; this RFC writes its
narrative copy — the wind-down screen, the "run 2 opens with" beat) → run 2 reaches Tier 1
faster (D6 assembly working as designed) → first elective Exit lands in [45,90] min (the
existing harness gate now measures REAL content instead of placeholder values — AC1).

### T3 — Copy discipline

Every player-facing string through the flavor bible voice rules; every real-world statistic
carries verified provenance in the claim registry; anything not yet verified is flagged, not
shipped. The copy lands as catalog/copy-system data (the content pipeline design), reviewable
like any diff.

**Carried debt (PT-C4, 2026-08-07):** generator presentation copy has NO binding surface in the
economy grammar (rows carry no copy field, and the Copy Pipeline forbids deriving display text
from mechanical IDs) — this RFC owns defining how generators become player-facing (either the
generator-copy grammar extension PT-C4 declined to add, or a presentation-layer binding). The
first two consumers waiting on it: `generator.beige_tower` (shipped, presentation-less) and
`generator.legal_dept` (Permits RFC; its title/description explicitly carried here).

## Acceptance criteria

0. **F1 precondition:** a founder idle beyond the offline cap under a provisioning catalog
   evaluates successfully (no online-horizon brick) — the mint-blocking regression.
1. The pacing harness passes on REAL content: the grown T0–T1 milestone set's distributions within design/02 §11
   targets; the [45,90] elective-Exit gate green with the T0–T1 catalog (not fixture values).
2. Relevance report (if harness landed): zero dead purchasables in the T0–T1 window; every
   generator class shows a non-production role activation.
3. Every upgrade/generator/gate loads through strict loaders; the mint follows the epoch
   protocol; golden reports regenerate.
4. First-session fixture: a scripted persona completes the T2 sequence against the composed
   gameserver (the composition integration harness gains one content-driven run).
5. Copy audit: no 🔴-flagged names, no unverified statistics (grep-able provenance tags in the
   copy source).

## Open questions

- Exact generator/upgrade names and joke density — content review with the owner (the fun part).
- Whether `Horse Armor (Free)` needs the cosmetic-slot system or ships as a stub shelf item
  (recommend: stub shelf — the cosmetics system is its own later RFC; the joke can't wait).

## Changelog

- 2026-08-03: created (draft) — the oldest remaining contract, now the critical path to a game.
- 2026-08-06: non-normative reference cleanup for publication; no spec change.
