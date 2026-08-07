# RFC: The Pitch (minigame content — THE TEMPLATE)

- **Status:** draft — Wave-B content; queued for Codex acceptance review. **This is the exemplar
  minigame-content RFC**: its structure (tenant row → engine contract → certified result → economy
  hooks → content-as-data) is the template the other minigame content RFCs replicate.
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-07
- **Design refs:** `design/03 §8/§9` (casino tier + Demo Disc Arcade — the two homes), `design/03 §12b`
  (the Fairness Law — ranked outcomes never scale with tier power), `design/02 §2.2` (the
  multiplicative production stack this game deliberately mirrors)
- **Depends on:** Minigame Platform Foundation (implemented + archived — C1–C40; this registers as a
  tenant), Fiscal Quarters (archived — the unlock sink), Soul Foundation (`soul_gate` field,
  archival-eligible). No new platform mechanics: this RFC adds a TENANT, an ENGINE, and CONTENT.
- **Planning:** `planning/minigame-the-pitch/` (once implementing)

## Summary

A solo score-multiplier card game: draft **metric cards**, play scoring hands, and stack
**growth-hack modifiers** whose effects multiply — `valuation = metrics × hype`, computed
**Decimal-exact** and deliberately pushed into the big-number regime the numeric core already
governs. Rounds are **funding-round targets**: beat the round's valuation threshold or the run ends.
It is turn-based, deterministic, PvE-against-a-threshold — the AI-fallback law is satisfied by
construction (solo by design), and the certified result is a pure recomputation of
`(seed, ordered choices)`. Mechanically this is the game's own multiplicative buff-stack exposed as
a card game — and computing it in exact `Decimal` where the inspiration genre lives with float
overflow is itself the satire.

## Motivation

The minigame platform ships with a test-only fixture tenant — zero real games. The research ranked
this shape first on every axis that matters (server-validatable, bottable, economy-hook, clock) and
uniquely on core-alignment: the score engine IS our production formula. It is the cheapest real
minigame we can ship and the best template, because it exercises every platform seam (tenant row,
certified result, faucet, offline-quality, unlock, soul_gate) with the least new engine surface.

Out of scope: multiplayer/ranked modes (solo only in v1; an async daily-seed board is a declared
successor), the card/hack CATALOGS beyond the launch set (content/data), and all numbers.

## Specification

### TP1 — The tenant row (the platform's C37 grammar, filled in)

One row in the pinned `minigames` artifact:
`{minigame_id: "pitch", engine_ref: "pitch", engine_version: 1, modes: ["solo"],
result_score_fact_ids: [...], scaling, payout, fallback, offline_quality, rating_policy,
unlock_condition}` — every sub-object per the already-ruled closed grammars. Specifics:
- **`scaling`:** the Fairness Law applies in full — NOTHING here scales with tier power. The only
  scaling input is `breadth` (unlocked card-set variants), never a power stat. Loader-rejected
  otherwise (the platform already enforces this).
- **`fallback`:** `solo` (the closed fallback row for solo-by-design games — no bot exists because
  no opponent exists; the threshold is the opponent).
- **`soul_gate`: `human_hobby`** — the founder's hobby; locks at near-zero Soul per SB15.
- **`unlock_condition`:** the Tier-5 casino home unlocks via the standard catalog fact; the Demo
  Disc Arcade variant is tutorial-tier per `design/03 §5`. The purchased unlock uses the **Fiscal
  Investor-Confidence unlock sink** (`fiscal_unlocks` — the cross-foundation tie this template
  demonstrates).
- **`rating_policy`:** none in v1 (solo, unrated — `rating_delta: null` per C40; Elo/count unchanged,
  resolution + offline-quality still commit).

### TP2 — The engine contract (deterministic, Decimal-exact)

A new engine package (`server/pitch` + `client/src/pitch`) behind the platform's engine seam:
- **A run** = `(seed, ordered choices)`. The seed derives from the platform's session identity via
  the established substream discipline (a named `pitch.run.v1` substream); a **daily-seed variant**
  uses the calendar day as the published seed input (the built-in-predictor stance: the seed is
  SHOWN, the skill is the drafting).
- **The loop (structure ruled; all magnitudes are catalog data):** each round presents a drafted
  hand of `metric_card`s (each carries a base `metric` value) and the player's slotted
  `growth_hack`s (each a modifier in a **closed effect union**: flat metric adds, per-card
  multipliers, hand-shape multipliers, and hack-order interactions). A played hand scores
  `valuation = Σ(metrics after adds) × Π(hack multipliers)` — **computed in `Decimal`, never
  float** (the big-number law; scores are EXPECTED to enter the `break_infinity` regime — that is
  the point and the joke). Round N's `funding_target` (a Decimal threshold from the catalog curve)
  must be met within the round's hand budget or the run ends. Between rounds: a shop/draft step
  (spend run-local score-currency on hacks/cards — run-local, NEVER company resources).
- **Hardcaps, visibly:** hack slots, hand size, and rounds are visible hardcaps with reason keys.
  The score itself is uncapped (the numeric core governs it — watching the number leave the solar
  system is the reward).
- **The certified result** = `{final_round, best_hand_valuation (canonical Decimal string),
  outcome}` — recomputable byte-for-byte from `(seed, choices)`; `engine_version` pins the calc;
  the platform's resolve composer (C37–C40) owns everything after certification. Go/TS byte-parity
  golden vectors over the scoring math, including hack-order interactions and the
  big-number boundary.

### TP3 — Economy hooks (all via the platform, nothing new)

Payout = the certified score through the platform's grade curve + **faucet governor** (daily cap,
forfeit-with-reason above it — already shipped). `offline_quality` charges/decays per the C34
ladder. No Clout (the platform's ruled no-Clout law). Numbers (grade thresholds, payout rates,
funding-target curve, card/hack magnitudes) are balance data.

### TP4 — Content scope (the deckbuilder discipline)

Launch catalog: **small and flat** — the research's content-obligation warning is adopted as a
constraint: ≥~12 `metric_card`s and ≥~8 `growth_hack`s, each individually flat and legible, rather
than a large pool (depth comes from hack INTERACTIONS, which the closed effect union makes
enumerable and testable). Card/hack rows are exact-key catalog data with copy keys through the copy
pipeline (flavor lives in data; `pitch`/`metric_card`/`growth_hack`/`valuation` are the mechanical
names). Every card/hack must pass the relevance harness's floors once T0-1's production baseline
activates (dead content fails CI — the instrument exists; use it).

### TP5 — The template notes (for the ~10 RFCs that replicate this)

The reusable skeleton this RFC establishes: ① one tenant row filling the C37 grammar (never new
platform mechanics); ② one engine package with a versioned pure certified-result function over
`(seed, choices)`; ③ economy hooks exclusively through the shipped platform seams; ④ content as
exact-key catalog data with copy-pipeline keys; ⑤ Fairness-Law scaling declaration; ⑥ soul_gate
classification; ⑦ numbers-as-data. A content RFC that needs anything outside this skeleton is
touching platform territory and must say so explicitly.

## Deviations from design

None — `design/03`'s casino/arcade homes, clock taxonomy (session-skill), and the Fairness Law are
followed; mechanical naming per the naming law.

## Acceptance criteria

1. Tenant row loads under the pinned artifact in both runtimes; Fairness-Law loader rejection
   proven (a power-stat scaling input fails); soul_gate honored (near-zero Soul rejects start).
2. Engine determinism: `(seed, choices)` recomputes the certified result byte-identically Go/TS;
   golden vectors cover scoring incl. hack interactions and a big-number-regime hand; daily seed
   published in the wire.
3. Full platform path: create→play→resolve→payout for a real run through the shipped composer;
   faucet cap forfeits with reason; offline_quality charges; unlock via the Fiscal sink proven.
4. Content: launch catalog rows exact-key validated; copy keys resolve; relevance floors green once
   the production baseline exists.
5. No Company resource enters the run; no run-local currency leaves it (grep + test).

## Open questions

- The async daily-seed leaderboard variant (same seed, compare certified scores) — a declared
  successor once the platform's async_snapshot mode has a first consumer.
- Which of the four researched turn-based/snapshot surfaces ship (the redundancy watch) — a
  `design/03` roster decision owed before the NEXT combat-shaped minigame RFC; The Pitch is not a
  combat surface and does not wait on it.

## Changelog

- 2026-08-07: created (draft) — Wave-B opener; the exemplar minigame-content RFC.
