# RFC-0002: Economy Constants, Ceilings & Accrual Policy

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-07-27
- **Design refs:** `design/02-economy-balancing.md §2.1, §3` (cost curves, hardcaps), `design/00-vision.md` (anti-goals: no softcaps), `design/01-tiers.md` (the ending), `design/07-roadmap.md`
- **Depends on:** RFC-0001 (numeric core)
- **Supersedes / superseded by:** —
- **Planning:** — (not yet implementing)

## Summary

RFC-0001 specified *how numbers behave*. It did not specify *which numbers the game uses*, and several gameplay-consequential constants are currently either unstated, stated inconsistently across documents, or being chosen implicitly by whoever writes the code first. This RFC collects those decisions into one owned place, resolves the ones that have a defensible answer today, and marks the rest as explicit owner decisions with proposed defaults.

It also fixes one **latent correctness bug** in the interaction between RFC-0001's quantization rule and a stated design pillar (§D6) — that item alone justifies the RFC.

## Motivation

A coverage audit found six decisions being made silently. They share a shape: *a number or policy with real gameplay consequences that engineering will otherwise pick by default, because a reasonable default exists and therefore nothing ever blocks.* This is precisely the class the `DESIGN-GAP:` convention cannot catch — an agent only escalates when it is stuck, and none of these make anyone stuck.

**In scope:** cost-curve ratio(s); the visible hardcap; ceiling behaviour; leaderboard ordering keys; production accrual precision policy.

**Out of scope, with owners named:**
- **Client reconciliation feel** (snap vs. lerp vs. rebase). `design/06` says "hard-reconciles to server snapshots", which is a decision made in a table cell; it belongs to the client-engine RFC and should be restated there normatively.
- **Offline progression defaults**, currently a balance number living in `AGENTS.md`. Belongs to the production-engine RFC.
- **Per-tier pacing targets.** Owned by the balance-harness RFC; this RFC must not pre-empt them.

## Specification

### D1 — Cost-curve ratio `r` — ⚠️ OWNER DECISION

**Current state, conflicting:**

| Source | Says |
|---|---|
| `design/02 §2.1` | `price(n) = base × r^n`, **r = 1.13** (band 1.07–1.15) |
| `design/research/pacing-science.md` | **r ≈ 1.07–1.10** — "gentler than CC's 1.15; longer smooth runs suit a 2-month arc"; explicitly notes *"design/02 currently says 1.13 — revisit against the harness"* |
| Shipped reference points | Cookie Clicker **1.15**, AdVenture Capitalist **1.07** |
| `tools/gen-vectors.mjs` | generates economy vectors for **all three** of 1.07 / 1.13 / 1.15 |

The golden-vector coverage of three ratios is **correct engineering** — the library must handle the whole band — and is not itself the problem. The problem is that *the game value* is unresolved while documents assert it as settled.

**Two decisions, not one:**

1. **Is `r` global or per-generator-class?** `design/02` presents a single `r`; the original design intent recorded in planning was "r per generator class". **NORMATIVE: `r` is a per-generator-class field in balance data, never a global constant in code.** Rationale: it costs nothing to make it per-class, it is the strictly more general form, and it is required by `design/02`'s own tier-pacing table where late-tier generators break cadence deliberately. Nothing downstream may assume a single global `r`.
2. **What is the launch default?** ⚠️ **Owner decision.** Proposed default: **1.10** — the top of the pacing research's band, splitting the difference against `design/02`'s 1.13, and defensible because pacing-science derived its band from *our stated two-month arc* whereas 1.13 predates that research. **This default is provisional and the balance harness overrides it.** RFC-0002 must not be read as freezing `r`.

**Required follow-up:** whichever value is chosen, `design/02 §2.1` and `pacing-science.md`'s note must be reconciled in the same commit. Two docs disagreeing about the game's most consequential number is the actual defect.

### D2 — The visible hardcap — ⚠️ OWNER DECISION

`design/00` and `design/02 §3` make **"hardcaps, never softcaps — any cap is a visible number with a tooltip"** a design law. **No such number exists anywhere in the repo.** RFC-0001 defines only a *technical* cap (`2^53 − 1` for exact counts, exponent domain ±9e15) and explicitly defers: *"later system RFCs may set lower visible hardcaps."* This one is that RFC, for the economy layer.

**NORMATIVE:**
- Hardcaps are **per-resource**, declared in balance data, never inferred from the numeric type's limits. A resource with no declared cap is uncapped; this must be explicit, not a missing field.
- Every declared cap ships with a **`cap_reason` string** rendered in the tooltip. The design law requires a visible number *and an explanation*; a number alone does not satisfy it.
- **The technical limits of RFC-0001 are never a game-visible cap.** If a player can reach `2^53 − 1` of anything, that is a balance defect, not a cap to be surfaced.

⚠️ **Owner decision:** whether a **global ceiling** exists at all, distinct from per-resource caps. See D3 — the two questions are coupled, and there is a strong narrative answer.

### D3 — Ceiling behaviour — RESOLVED (proposed)

**Current behaviour, by accident:** `server/decimal/decimal.go` sets `maxExponent = 8_999_999_999_999_999`. `Quantize` returns `NaN` on carry at that boundary. RFC-0001 §5 requires that a transition producing a non-finite value *"fails without mutating state."* Composed, that means: **at the ceiling, purchases silently fail.** No message, no explanation, no diegesis.

That is the worst available outcome, and it is nobody's decision — it is three correct local choices producing a bad global one.

**NORMATIVE:** the economy layer must not be able to reach the numeric ceiling. A declared per-resource hardcap (D2) must exist below it for every unbounded-growth resource, such that **`Quantize` overflow is unreachable through legitimate play and remains strictly an invariant-violation signal.** Reaching it is a bug report, not a game state.

**Design opportunity, and the reason this is worth doing properly:** our arc *ends* with the world stripped of resources. The numeric ceiling and the narrative ceiling should be the same event. A resource approaching its declared cap is the Depletion tier's most honest possible signal — the number stops going up because *there is no more of the thing*, which is the entire thesis of the game. Recommend `design/01` and `design/13` own that beat and this RFC merely guarantees the engine reaches it cleanly rather than erroring.

### D4 — Leaderboard ordering keys — RESOLVED

RFC-0001 §3 quantizes committed state to 12 significant digits. Two runs differing below the 12th digit are indistinguishable. In a game framed around **speedrun world records**, silent ties are a credibility problem.

**NORMATIVE: leaderboard and ranking order keys are exact integers or exact times, never quantized `Decimal` values.** Run comparison ranks on RTA/IGT (integer milliseconds) and on exact integer counts. `Decimal` magnitudes may be *displayed* alongside a ranking but must never be the sort key. Where a magnitude genuinely must be ranked (e.g. a "largest bank" board), rank on `(exponent, quantized_mantissa)` and **display ties as ties** rather than resolving them arbitrarily.

This dissolves D5-as-originally-stated: 12 digits is the right precision for state; it was only ever a problem because nobody had said what ranks.

### D5 — Accrual precision — RESOLVED (**latent bug fix**)

RFC-0001 §3 lists **accrual** as a state transition, and therefore quantizes it to 12 significant digits. RFC-0001 §3 also says intermediates keep full precision. **These compose badly and nobody has specified the boundary.**

The failure: with 12 significant digits, `bank + gain` loses `gain` entirely once `gain` is more than ~12 orders of magnitude below `bank`. If accrual is quantized **per production source**, every early-tier generator's contribution rounds to exactly zero in the late game.

That silently destroys a stated design pillar. `design/02` requires **milestone multipliers at 25/50/100 that keep old generators alive**, and `design/research/cookie-clicker.md` identifies exactly this — tiered upgrades keeping early buildings relevant — as one of the mechanics worth stealing. The numeric spec as written would quietly delete it, and the symptom (an upgrade that visibly does nothing) would be extremely hard to trace back to a rounding rule.

**NORMATIVE:**
- **Production is summed across all sources at full intermediate precision and quantized exactly once, at the commit boundary.** Per-source quantization is forbidden.
- The same rule applies to any multiplicative stack: apply the full stack at intermediate precision, quantize the result once.
- **Test gate:** a golden vector in which a single source's contribution is 13+ orders of magnitude below the bank, asserting that a large number of such sources summed together **does** move committed state. This is the regression test for the bug this section prevents.

### D6 — Minimum visible increment — RESOLVED

Given D5, per-tick accrual reaches committed state correctly, so the remaining concern is presentational.

**NORMATIVE:** a displayed counter must never appear frozen while the player's total production rate is greater than zero. Display interpolates between committed states at full client precision; it does not re-quantize for rendering. If a resource is at a declared hardcap (D2), the counter is *deliberately* static and **must show the cap and its `cap_reason`** rather than simply stopping — a frozen number with no explanation is indistinguishable from a bug, and this game has a design law against unexplained caps.

## Deviations from design

- **`design/02 §2.1`** states `r = 1.13` as settled. This RFC reopens it (D1) and requires the value to live in balance data rather than prose. `design/02` must be amended in the same commit as the decision.
- No other deviation. D2/D3/D6 *implement* existing design laws that were previously unimplementable for want of a number.

## Acceptance criteria

1. `r` is declared per generator class in balance data; no global `r` constant exists in Go or TS source.
2. Every resource in balance data has either a declared `hardcap` + `cap_reason`, or an explicit `hardcap: null`. Schema validation fails on a missing field.
3. A test asserts the economy layer cannot reach RFC-0001's `maxExponent` through any legitimate purchase or accrual path at any declared balance configuration.
4. Golden vector for D5 passes: ≥10⁶ sources each 13+ orders of magnitude below the bank, summed, **do** move committed state; the same sources quantized per-source demonstrably do not (kept as a negative control).
5. No leaderboard or ranking code path takes a quantized `Decimal` as a sort key (enforced by review + a type-level distinction if practical).
6. `design/02 §2.1` and `pacing-science.md`'s conflicting-`r` note are reconciled.

## Open questions

- ⚠️ **D1(2): launch default for `r`.** Proposed 1.10, provisional pending the balance harness. **Blocks `accepted`.**
- ⚠️ **D2: does a global ceiling exist distinct from per-resource caps?** Recommend "no — per-resource only, with the Depletion arc providing the narrative ceiling." **Blocks `accepted`.**
- **Deferred to the client-engine RFC:** reconciliation policy (snap/lerp/rebase) and its feel implications.
- **Deferred to the production-engine RFC:** offline progression constants, currently stranded in `AGENTS.md`.
- **Deferred to the balance-harness RFC:** all per-tier pacing targets. This RFC deliberately sets no pacing numbers.

## Changelog

- 2026-07-27: created (draft). Origin: coverage audit of silent decision points across RFC-0001 and the design corpus.
