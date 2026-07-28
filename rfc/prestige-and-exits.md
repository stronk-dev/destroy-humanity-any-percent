# RFC: Prestige & Exits

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Design refs:** `design/02 §3` (formula, sub-currencies, **Exit offers — designed 2026-07-28**), `design/11 §3–4` (run-end sequence, **scripted first failure, Advisor Mode**), `design/02 §6` (the Clout carry rule), `design/10 §5` (ledger persists, scores reseed)
- **Research:** `design/research/pacing-science.md` (first-Exit pacing), `design/research/morality-systems.md` (the reseed), `design/research/run-narrative-ux.md §6b` (as adopted)
- **Depends on:** Production Engine (implemented), Save Layer (implemented — Company/Founder stream split is the whole trick), Gate Predicates (draft — collapse-Exit Route Knowledge bonus; can land with a stub grant)
- **Planning:** `planning/prestige-and-exits/` (once implementing)

## Summary

The loop's hinge: ending a run. This RFC specifies Exit offers and Wind Down as intents, the atomic Company-reset/Founder-advance transaction, the prestige formula's evaluation, the morality reseed, the scripted first failure, and the new event kinds all of it requires. It is the first RFC whose *primary deliverable is a feeling* — the design docs own the feeling; this owns the transaction under it.

## Specification

### D1 — New intents

- `accept_exit_offer {offer_id}` — valid only against a live, unexpired offer event; terminal.
- `wind_down {}` — always valid from Tier ≥ 1 (the always-open door, `02 §3.3`); collapse-typed.
- `file_ipo {}` — valid only under the qualifying doctrine + filing-window state; opens the S-1 chain whose completion event carries the Exit.
- All follow the C1/C3 contracts (idempotent, evented, typed rejections — `offer_expired`, `not_eligible`, `window_closed` join the closed taxonomy by this RFC, which is how the registry grows).

### D2 — Offers are events with terms

Offer events (Layer-1, `09 §2`) carry a **server-computed terms object**: `{offer_id, exit_type, expires_at, payout_preview: {reputation_levels, network_slot_unlocks, route_knowledge, clout_reach_note}}` — computed at spawn from the run's live state by the same formulas acceptance will use, so **the preview is a promise, not an estimate**; acceptance recomputes against commit-time state and pays `max(preview, recomputed)` (the player is never punished for progressing between offer and click). Spawn probability scales on tier progress and harvested Quarters (balance data). Expiry and drift per `02 §3.3`.

### D3 — The Exit transaction (atomic, one commit)

1. Evaluate accrual to the commit instant (Production D1).
2. Compute `ReputationLevel = ⌊(lifetimeValue/T)^(1/3)⌋` delta; apply exit-type payout modifiers (data).
3. Write **Founder stream** revision: Reputation, Network grants, Route Knowledge grants (incl. collapse bonus), Clout lifetime total, Soul delta, founder age advance, **the run's ledger facts** (morality entries, executed routes, exit record).
4. **Reseed** company-facing moral scores for the next run: `clamp(90 − 0.35·Notoriety, 55, 90)` (`10 §5`).
5. Write **Company stream**: archived-final revision (the obituary's source data), then the new run's initial revision (catalog initials + Network-carried items + Reputation starter effects).
6. Emit `run_ended` and `run_started` events (new kinds, registered here) with the full terms object — **the run-end screen renders from the `run_ended` event alone**, so replays/verification see exactly what the player saw.

All six steps are one transaction across both streams (the Save Layer's multi-stream write; revisions advance together or not at all). A crash mid-Exit leaves the old run intact.

### D4 — The scripted first failure (`11 §3`, executable)

A founder with zero career Exits reaches a **scripted collapse trigger** at run-minute ~15 (exact trigger: first threshold crossing after 15:00 of attended time — deterministic, not wall-clock-fragile). It fires `wind_down` server-side with a distinct `exit_type: scripted_first` paying full first-Exit Reputation. It is **in every category's route** (verification treats it as a fixed segment; it cannot be skipped and advantages nobody). One per founder, ever — `New Founder` archives include it.

### D5 — Advisor Mode

Founder-scoped toggle; while enabled, a `prestige`-slot contribution of `+2% × completed_runs`, capped `+50%` (balance data), and **every run it touches carries the `Assisted` variable** (Leaderboards D4 — structural, set at timer start or on first enable mid-run). The label text is normative UI copy (`11 §3`); no other mechanical effect, no nag.

### D6 — What run 2 opens with

Deterministic assembly, in order: catalog initials → Network-carried items (designated upgrades exist, owned) → Reputation starter effects (tree purchases) → reseeded moral scores → this-run Clout = 0 (the carry rule: lifetime Clout affects reach surfaces only, never the stack). The **first 10 minutes of run 2 differ from run 1 by exactly these** — the walkthrough's "same game faster" observation is accepted for T0–T1 *by design* (`10 §3b`: decisions start at Tier 2); this RFC makes the carried differences it does promise real and enumerable.

## Acceptance criteria

1. Exit atomicity: a fault injected between any two D3 steps leaves both streams at their pre-Exit revisions (integration test, real Postgres).
2. Preview-is-a-promise: accept after further progress pays ≥ the previewed terms (property test across offer ages).
3. The reseed matches `10 §5` exactly across the Notoriety range; ledger facts survive every Exit un-mutated.
4. Scripted first: fires once per founder at the deterministic trigger, full payout, `scripted_first` typed; a second founder (via `New Founder`) gets their own.
5. Wind Down is accepted from any eligible state, including mid-event-chain (no state can trap a player in a run).
6. Run 2 assembly is byte-deterministic given the Founder stream (golden fixture: same founder state → identical opening company state).
7. `run_ended` contains everything `design/11 §3`'s screen renders (schema-checked against the copy system's template slots).
8. Harness hook: the Casual persona's first-Exit time lands in the [45, 90] min envelope with the scripted failure included (H2).

## Open questions

- Offer spawn-rate curve and exit-type payout modifiers: balance data, harness-gated.
- The S-1/IPO event chain content: `design/09` Layer-1 authoring, not blocking (the `file_ipo` intent can ship gated off until the chain exists).
- Founder-age advancement constants: balance data; the actuarial-wall interactions live with the Tier-6 content RFC.

## Changelog

- 2026-07-28: created (draft), immediately after the run-end design sitting it depends on.
