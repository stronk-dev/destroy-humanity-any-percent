# RFC: The Commons Compact

- **Status:** draft
- **Author:** Marco (drafted by Claude; boundary split per Codex's 2026-07-28 review)
- **Design refs:** `design/05 §5` (the commons + the front door, as designed 2026-07-28), `design/02 §7` (Trust constituencies; derivation rule), `design/10 §1` (Open Source = heavier participation)
- **Research:** `design/research/commons-game-theory.md` (the spec), `design/research/morality-systems.md` (solo floor 1.8–2.5×, corrections)
- **Depends on:** Save Layer (implemented — compact state is stream data), Production Engine (draft — consumes one named slot)
- **Planning:** `planning/commons-compact/` (once implementing)

## Summary

The Mutual Aid Compact: membership, the Health/Capacity computation, the Enclosure index, cohort assignment, and the **single named production slot** through which all of it reaches the economy. Codex's boundary, adopted: **production consumes the computed modifier through a generic slot and knows nothing else** — the commons package computes, the production package multiplies.

## Specification

### D1 — Membership

- `sign_compact` / `leave_compact` are intents (Production RFC contract: idempotent, evented, revision-tied). Signing is offered as the incorporation contract line-item every run (`05 §5`); Open Source incorporation auto-signs with the larger faction tithe.
- Membership is company-scoped state (resets at Exit, like the contract it lives in); the founder ledger records signature history as dated facts.
- Leaving is always allowed, takes effect at the next accrual boundary, and zeroes `sᵢ` (Solidarity rebuilds from scratch on re-signing — the TPP no-exit lesson requires a real exit, the Ostrom lesson prices it).

### D2 — Health, Capacity, and the Enclosure index

Per `research/commons-game-theory.md`, normative here:

- **Enclosure index `dᵢ ∈ [0,1]`** — derived **entirely from the member's own production stack**: the fraction of their multiplier stack (weight-normalized) coming from dark-pattern stages, externality-ledger-generating sources, and route-of-harm slots, evaluated at each accrual boundary. **No reports, no votes, no declared route flags.** The exact slot-weight table is balance data; the formula is published in-game.
- **Health** = weighted mean of member compliance `(1 − dᵢ)`, computed at three scales and blended `H = 0.5·H_guild + 0.3·H_cohort + 0.2·H_server`. Members without a guild take the cohort value in the guild term.
- **Capacity** = absolute sum of tithes; drives caps and content gates only, never the buff rate.
- **The buff**: `M = 1 + 5·[0.6·f(H) + 0.4·sᵢ]`, `f(H) = ((H−0.35)/0.65)^1.5` clamped ≥ 0. `sᵢ` (personal Solidarity, `[0,1]`) accrues with tithe-in-good-standing time, decays under `dᵢ` above threshold; parameters are balance data. All published.

### D3 — The slot boundary

The commons package exposes exactly one value to production: `commons_modifier(member) → Decimal`, populated into the fixed named slot `commons` (Production RFC D2). Non-members: slot absent (not 1.0 — absent; the panel shows no line). **The production package must not import the commons package or vice versa** — both depend on a shared slot-contract package only (compile-enforced).

### D4 — Cohorts

- **Server-assigned, non-elective, persistent, target size ~150** (`05 §5`). Assignment on first sign; rebalancing only on population collapse (merge, never split below floor 40) — cohort identity is the shadow of the future and must not churn.
- Cohort surface: the panel (named neighbors, co-ops, current standing — **current state only, never a permanent badge**, The Button rule), the cohort-scale Health term, the tithe-dial vote (monthly, direction-not-implementation, within a server band), and mercy scaling for small guilds (`05 §3`).
- Alt-resistance: assignment is non-elective and account-age-weighted; a founder's cohort persists across runs (founder-scoped).

### D5 — Ambient surfaces (the front door's server half)

Dispatch events for commons state transitions (Health band crossings, cascade onset, recovery) are Layer-3 server events (`09 §4`), visible to non-members. The one NPC recruiting event fires per founder per career at mid-T3 if never-signed. NPC co-ops hold Health near neutral below population floor (labeled in the formula panel).

## Acceptance criteria

1. `dᵢ` derivation: golden fixtures over representative production stacks (canon-heavy, ethical, mixed) produce the specified indices; **no code path reads route flags or player declarations.**
2. The buff formula matches spec across the H × sᵢ grid, including the clamp and the 40% floor property (total collapse costs a loyal member ≤ 47% of their commons contribution, never the base game).
3. Slot boundary: commons and production packages share no imports beyond the slot contract (build-enforced); a non-member has no `commons` slot.
4. Sign/leave/re-sign: evented, idempotent, `sᵢ` zeroed on leave; the incorporation line-item appears every run.
5. Cohort assignment is deterministic given server state, non-elective, and stable across runs; merge preserves standing.
6. Population invariance: Health-driven buff rates are statistically indistinguishable at simulated 200 vs 20,000 CCU (harness scenario).
7. All formulas render in the in-game formula panel (law 9), including `dᵢ`'s slot-weight table.

## Open questions

- Slot-weight table for `dᵢ`, sᵢ accrual/decay constants, tithe band: balance data, provisional per the research, harness-gated.
- Guild-term fallback for guildless members (cohort substitution) is specified here; revisit if the harness shows it double-weights cohorts.

## Changelog

- 2026-07-28: created (draft) from the commons front-door design + Codex's boundary split.
