# RFC: Prestige & Exits

- **Status:** implementing
- **Author:** Marco (drafted by Claude)
- **Design refs:** `design/02 §3` (formula, sub-currencies, **Exit offers — designed 2026-07-28**), `design/11 §3–4` (run-end sequence, **scripted first failure, Advisor Mode**), `design/02 §6` (the Clout carry rule), `design/10 §5` (ledger persists, scores reseed)
- **Research:** `design/research/pacing-science.md` (first-Exit pacing), `design/research/morality-systems.md` (the reseed), `design/research/run-narrative-ux.md §6b` (as adopted)
- **Depends on:** Production Engine (implemented), Save Layer (implemented — Company/Founder stream
  split is the whole trick), Gate Predicates (implemented — collapse-Exit Route Knowledge bonus),
  T0–T1 playable content, account/session bootstrap
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
8. Harness hook: the Casual persona's **first elective Exit** (first non-scripted Exit) lands in the [45, 90] min envelope; the scripted first failure is a fixed ~15-min segment excluded from this envelope and included in total run time (ruling 2026-07-29, resolving the D4/AC8 contradiction found by Codex's review).

## Open questions

- Offer spawn-rate curve and exit-type payout modifiers: balance data, harness-gated.
- The S-1/IPO event chain content: `design/09` Layer-1 authoring, not blocking (the `file_ipo` intent can ship gated off until the chain exists).
- Founder-age advancement constants: balance data; the actuarial-wall interactions live with the Tier-6 content RFC.

## Executable contracts (answering the 2026-07-29 review)

### P1 — Persisted state (save-version bump, both scopes)

**Company v(next):** adds `tier int`, `lifetime_value canonical`, `offer_state {offer_id, exit_type, terms_json, spawned_at, expires_at} | null` (at most one live offer), `run_started_at ms`. **Founder v(next):** adds `reputation_level int`, `reputation_unlock_ppm int`, `network_slots [{slot, carried_ref}]`, `clout_lifetime int`, `soul int`, `age_ms int`, `notoriety int`, `advisor_mode bool`, `exit_history [{run_id, exit_type, occurred_at, reputation_delta}]` (append-only), plus the existing route/knowledge fields. Migrations default everything to zero/null/empty; corpus gains one fixture per scope. **No field is inferred from flavor names; this list is the closed set.**

### P2 — Prestige arithmetic (exact, both runtimes)

`T` is balance data (provisional `1e12`). Algorithm: `ratio = lifetimeValue.Div(T)` (Decimal); `level = floor(cbrt(ratio))` computed as **integer binary search over n: largest n with `Decimal(n)^3 ≤ ratio`**, n capped at MaxExactInteger — no floating cbrt anywhere, identical by construction in Go/TS. Exit-type modifiers are integer ppm multipliers on the **delta** (`new_level − old_level`, never total), applied then floored; collapse's Route-Knowledge bonus is a flat grant (routes catalog). Golden vectors: cube boundaries n³±1ulp, zero, T-exact, modifier rounding.

### P3 — Offer state machine

Spawn check runs **inside accrual evaluation** (Production D1) at threshold crossings and Quarter harvests only (deterministic sites, no timers): if no live offer and `spawn_gate(tier_progress, harvested_quarters)`
<!-- Fiscal F15 (2026-08-06): the **Quarter-harvest spawn site** is a cross-stream trigger — Fiscal is Founder-scoped and emits the immutable `fiscal_period_harvested.v1` fact, but offer state is Company-owned. The Founder→Company bridge that converts a Quarter harvest into a Company-side offer-spawn is DEFERRED to a successor multi-stream/event-consumer RFC; until it lands, only the threshold-crossing spawn site is live. `harvested_quarters` in the gate reads the frozen/committed Founder count via that consumer, never an ambient Founder read. -->
 (balance data, integer ppm table) exceeds a draw from **the save-seeded SplitMix64 stream** (seed = founder seed ⊕ run_seq — replayable), emit `exit_offer_spawned` with the full terms object (P1 shape). Expiry is checked at evaluation sites (an expired offer nulls state and emits `exit_offer_expired` — no background jobs). **Decline = `decline_exit_offer` intent** (evented; next spawn's terms drift by the declared ppm walk). `max(preview, recomputed)` is **field-wise on integer fields** (reputation_delta, route_knowledge); Network slot unlocks are **set-union** (preview slots ∪ recomputed slots); nothing non-monotonic exists in terms.

### P4 — Multi-stream atomicity (the store extension this RFC owns)

New store op `ApplyExitTransaction(founderStream, companyStream, newRunInit)`: **lock order is founder-then-company (lexicographic stream-id tiebreak), both `FOR UPDATE`, one Postgres transaction**; expected-revisions for both streams in the intent envelope; idempotency record written under the **company** stream (the intent's origin); events: `run_ended` on the old company revision, `run_started` on the new, founder events on the founder revision — all in the same commit. The "new run" is **the same company stream, `run_seq+1`** (established by Gate Predicates C4) — no new stream identity, archives are revision-history. Retry = standard idempotent replay; compensation is never needed because nothing partial can commit.

### P5 — Scripted-first (contradiction resolved by ruling, 2026-07-29)

The pacing envelope measures the **first elective Exit** (AC8 as amended). The scripted trigger is: first `threshold_crossed` event with `attended_ms ≥ 900_000` on a founder with `exit_history == []` — evaluated server-side at event emission, firing `wind_down(exit_type=scripted_first)` in the same evaluation. Attended-ms derivation is P6's.

**P5b (review ruling, 2026-07-30):** an elective `wind_down` from a founder with empty
`exit_history` IS the scripted first failure — typed `scripted_first`, full first-exit payout,
regardless of attended time. The always-open door (D1) and the unskippable curriculum (D4) are both
preserved; AC4 gains this case.

**P5c (review ruling, 2026-07-30):** Exit offers do not spawn while `exit_history` is empty — the
scripted first failure precedes the market noticing you. Closes the offer-path curriculum skip the
P5b implementation review found.

**P2d (review ruling, 2026-07-30):** the delta × modifier ÷ 1e6 product is exact integer
arithmetic in BOTH runtimes (Go big.Int, TS BigInt); no Decimal float path. The reproduced ±1
divergence point is a corpus vector.

**P2b/P2c (review rulings, 2026-07-30):** the TS kernel wraps prestige division as
mantissa/mantissa to match Go's `Div` (reciprocal-multiply double-rounds; a non-unit-mantissa
threshold can flip a cube boundary by 1 ulp); golden vectors gain non-unit-mantissa thresholds at
cube boundaries. `ReputationDelta` SATURATES at MaxExactInteger in both runtimes — never an error
on the exit path.

### P6 — Timer facts (recorded here, consumed by Leaderboards)

`run_started_at` (server ms, P1) starts RTA. **Attended Time = RTA − Σ offline spans**, where an offline span is any accrual evaluation whose elapsed exceeded the online session gap (`catchup_ceiling_ms`, already catalog data) — the span is recorded as `{from, to}` on the company state's append-only `offline_spans` (capped list, oldest-collapsed). All integer ms, all server-derived; the client clock contributes nothing.

**P6a/P6b (review rulings, 2026-07-30):** the next save version backfills `RunStartedAt :=
EvaluatedThrough` where zero (pre-v7 migrated runs — currently trapped un-exitable) and flags such
runs pre-timer (time-board excluded); span-list overflow drops the oldest span into a
`collapsed_offline_ms` accumulator so total offline duration is invariant (the shipped collapse
absorbed online time). Corpus gains the company v6→current case.

**P6c (Faction review ruling, 2026-07-30):** `catchup_ceiling_ms` is owned by the immutable
Prestige policy artifact. Prestige offline-span accounting and faction attended-stock accrual
resolve the same value from the run's pinned `constants_hash`; no process-configured gameplay
ceiling remains. The client-shell copy is presentation policy only.

**P4b (spec correction, 2026-07-30):** the `run_ended` event + run log are the obituary's source of
record (AC7 already guarantees this); revision retention stays at 5 and P4's "archives are
revision-history" phrase is corrected accordingly. Idempotent replay returns events in RECORDED
order (by committed seq), not kind-mapped order. Decline drift (P3) is scoped per run: `declined`
counts reset with `run_seq`.

### P7 — Log retention for replay (shared answer with Leaderboards #1)

This RFC emits every event replay needs; **the run log itself is the Leaderboards RFC's `run_log` table** (canonical intent payloads, receipts, sequence — written at intent commit, retained per-run until the run is verified+archived or abandoned+expired, exempt from the 30-day `intent_records` prune). Prestige's only obligation: `run_ended` carries the log's terminal sequence number so verification knows completeness.

### P8 — Unblocked-by

Account & Session Bootstrap (new draft) supplies founder/session identity; T0–T1 playable content remains the only *content* dependency and gates only the *feel*, not this RFC's transaction — implementable against the Phase-0 catalog with fixture content.
## Changelog

- 2026-07-28: created (draft), immediately after the run-end design sitting it depends on.
- 2026-07-29: updated implemented dependencies; Codex acceptance review found the scripted-first
  timing contradiction and nine additional executable-contract gaps.
- 2026-07-29: accepted for implementation by the ordered Codex batch manifest; implementation
  started after Account & Session Bootstrap supplied the required Founder owner.
- 2026-07-30: implemented P2b/P2c, P4b, P5b, and P6b remediation: save v9 preserves collapsed
  offline duration, prestige math is cross-runtime exact and saturating, replay follows committed
  event order, decline drift is run-scoped, tier advancement is monotonic, and the first Wind Down
  remains the scripted curriculum Exit.
