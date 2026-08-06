# Fiscal Quarters Foundation implementation log

## 2026-08-05 — Codex acceptance review: blocked (F1-F8)

Review by: Codex. Recorded by: Codex.

Verified that Founder scope, scalar version 19 after minigames 17/pets 18, `ApplyFounderLogged`, and
server-stamped replay inputs are the correct implemented substrate. The early-harvest proposal is
coherent and the lazy auto-report avoids a background tick loop. The draft still leaves eight
load-bearing contracts undecided: multi-period catch-up/order, a nonexistent claimed Founder RNG,
server-derived spend pricing, cross-stream production effects, exact artifact grammar, announced
hardcaps, closed wire/rejection/event objects, and activation/clock-regression behavior.

No implementation or balance literals were introduced. The design's approximate 20h/23h/24h,
50%, +1%, and cap-100 values remain balance data rather than acceptance-review facts.

## 2026-08-06 — owner rulings on acceptance blockers F1-F8 (all resolved)
- F1: no bg job; n=floor((now-opened)/auto_ms) before the action, mint n*credit, opened+=n*auto_ms
  (phase preserved); compound transition, aggregated auto-report event first; harvest consumes auto.
- F2: FIXES my nonexistent-RNG error — early-harvest draw = pure SplitMix64 over (founder_id,
  fiscal_period_seq), substream fiscal.early_harvest.v1; seq++ on every close so no draw reuse; never
  now_wall_ms. (FQ3 body reconciled.)
- F3: remove client `amount`; server derives cost; triangular k*(2L+k+1)/2 wide-int; unknown/dup/zero
  use closed rejections. (FQ4 body reconciled.)
- F4 (owner): FREEZE-AT-GENESIS — levels+hoard frozen into Company genesis for the run (materialized
  genesis rows, no ambient Founder read); mid-run changes affect NEXT run only. (FQ4/FQ5 reconciled.)
- F5: `fiscal` artifact schema-v1 {clock,credit,hoard,generator_level rows,unlock rows}; cross-check
  gens vs economy artifact; floor>=19 biconditional, requires minigames+pets.
- F6 (owner): VISIBLE hardcaps+reason keys for credit + level rows; accrual saturation; prove no
  overflow; fiscal_period_seq does NOT saturate (reject horizons that could exhaust it).
- F7: reuse closed taxonomy (not_eligible/period_not_ripe|already_unlocked, unaffordable/fiscal_credit,
  cap_exceeded/target) — my invented not_ripe/insufficient_credit removed; manual vs auto via closed
  `source` field. (FQ6 + acceptance reconciled.)
- F8: init period_opened from the v19 genesis/Exit-activation server timestamp; now_wall>=opened
  required, equality ok, regression => internal_invariant no-op; enforce timestamp safe-int domain.
Status -> accepted; F1-F8 ruled; implementing. Body/README reconciled.

## 2026-08-06 — Codex implementation acceptance recheck: blocked (F9-F15)

Review by: Codex. Recorded by: Codex.

Read the fully ruled RFC against the shipped Founder replay/store, save-version, run-genesis,
contribution-provider, and Prestige offer surfaces. F1-F8 settle the gameplay but do not yet select
one executable persistence/wire result:

- F9: the artifact's nested rows and multiplier identities are not exact;
- F10: `(founder_id,seq)` does not select a byte-level seed derivation;
- F11: rejected Founder decisions cannot persist the mandatory pre-action auto-sweep;
- F12: frozen Founder effects have no immutable run-owned storage without changing Company state;
- F13: v19 fields/key completeness and timestamp injection are not enumerated;
- F14: Fiscal target/resolved/receipt/event objects are not byte-exact;
- F15: Prestige requires a Quarter-harvest spawn site, but no transition owns both streams.

Executable proposals reuse the existing runidentity/determinism functions, introduce immutable
run-keyed contribution rows (live provider → existing replay-input contributions), and recommend a
v19-only compound Founder receipt rather than weakening rejected-intent immutability. No Fiscal code
or balance numbers were introduced. Implementation remains blocked pending owner rulings and body
reconciliation.
