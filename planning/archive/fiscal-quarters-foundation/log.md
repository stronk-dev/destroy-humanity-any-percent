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

## 2026-08-06 — owner rulings on the 3rd-round blockers F9-F15 (all accepted)
- F9: exact artifact rows (clock/credit/hoard/generator/unlock) w/ multiplier slot/source/target,
  Decimal factors, shared mutation corpus.
- F10: seed = runidentity.Seed(founder_id, seq) (SHA-256 first-8 big-endian xor seq) →
  determinism.Substream(base,"fiscal.early_harvest.v1").Bound(1e6); byte-identical TS port + vectors.
- F11 (owner): compound wrapper {intent_id,outcome,founder_revision,fiscal_sweep,action} on EVERY
  Founder command under v19; non-null sweep => outer applied + revision commits even if action rejected
  (no rejected-with-mutation added to the store); harvest post-sweep = consumed_by_auto.
- F12 (accept): immutable run_frozen_contributions rows keyed (company_stream_id,run_seq,source_id)
  {slot,target,factor}, written at genesis/Exit, immutability trigger, ContributionProvider reads
  run-keyed rows, replay-input array carries bytes -> NO Company save bump, no ambient Founder read.
- F13: Founder v19 keys fiscal_credit/opened_wall_ms/seq/generator_levels(complete)/unlocks; activation
  inits from the logged server ts; pre-v19 none, v19 all.
- F14: byte-enumerated harvest/spend requests + resolved arms + outcome enum inside the F11 wrapper.
- F15 (owner): emit fiscal_period_harvested.v1 FACT only; Quarter->offer spawn routed to a successor
  Founder->Company consumer; prestige-and-exits.md line 81 amended with the dependency note.
Status -> accepted; F1-F15 ruled; implementing. Prestige cross-RFC reconciled.

## 2026-08-06 — designated INDEPENDENT verdict: Fiscal implementation (3930b6a^..e959f7e) — APPROVE (ready for archival)
Review by: the designated Claude reviewer (cross-party). Recorded by: same. Rule-(c) archival gate.
Verified in an isolated worktree at e959f7e; all gates re-run independently (make verify, typecheck,
client tests, real-Postgres integration) — GREEN. Range-union complete (3930b6a^ = 68c16dd chains from
the last reviewed point c4efe5f through only rfc-doc commits; no unreviewed CODE gap; no prior verdict).

**The four save-corruption/replay-determinism critical paths VERIFIED CORRECT:**
- F2/F10 SplitMix64 early-harvest draw: Go runidentity.Seed + determinism.Substream and the TS port
  reduce to identical SplitMix64(sha8 ^ seq ^ fnv1a64(label)); byte-identical constants + Bound
  threshold; pure fn of (founder_id, fiscal_period_seq); shared vectors (seq 0/1/high) asserted BOTH
  runtimes; now_wall never feeds randomness; seq increments on every close (no reuse).
- F1/F11-revised ROLLBACK sweep: clone-before + defer-restore on any non-applied outcome (rejected
  command rolls its sweep back, re-triggers next applied; phase-preserving opened+=n*auto never =now);
  no compound wrapper; consumed_by_auto prevents double-harvest. Tested both runtimes.
- F4/F12 freeze-at-genesis: FrozenContributionProvider reads ONLY run-keyed rows (no ambient Founder
  read grep-proven); replay uses the logged array; immutable table (trigger + deferred complete-set
  constraint); Company axis UNTOUCHED (LatestCompanyVersion stays 17); fault-injection covered.
- F8/F13 Founder v19 + clock guard: exact v19 keys (complete generator levels), pre-v19 rejects/v19
  requires all, biconditional floor>=19; activation from logged genesis ts; now<opened => ErrClockRegression
  no-op; validatedState prevents any invalid/NaN v19 persist.
Secondary F3/F5/F6/F7/F9/F14/F15 verified; kernel 0.3.60->0.3.63 lockstep, KV-1 covered.

**Two LOW findings (neither touches save integrity or replay):**
- F-1 (LOW, recommended): the F6 loader does NOT reject artifacts whose auto_ms horizon could exhaust
  fiscal_period_seq (only runtime prevents saturation). No realistic hazard (~24h => ~1e8 << 1e15).
  Deviation from the ruling's letter; add the loader horizon check.
- F-2 (LOW, REQUIRED before archival, ruling-mandated): the F15 rfc/prestige-and-exits.md amendment is
  present in the WORKING TREE but NOT committed at e959f7e — so the committed prestige RFC still lists
  Quarter harvests as an offer-spawn site with no deferral note. Root cause: ALL owner-ruling RFC edits
  are uncommitted (see the repo-hygiene escalation). Must be committed into/with the archival.
- INFO: cross-epoch existing-founder exit-activation of fiscal fields would ERROR (never corrupt),
  mirroring the archived pet/minigame pattern; owner-awareness only.

**Verdict: APPROVE. Ready for archival once F-2 (the Prestige amendment) is committed.**

## 2026-08-06 — post-filter archival closure

Review by: the designated Claude reviewer (cross-party). Recorded by: Codex.

The owner-approved unpublication filter remapped the reviewed implementation range to
`53d1b4a^..5347a4d`; the committed rewrite map proves content identity. F-2 is present in the
tracked Prestige RFC. Recommended F-1 landed in `4711f15`: both loaders reject a clock whose
earliest-close horizon could exhaust `fiscal_period_seq` within the safe timestamp domain, with a
shared negative fixture. Fiscal is implemented and archived; the optional artifact remains
unminted.
