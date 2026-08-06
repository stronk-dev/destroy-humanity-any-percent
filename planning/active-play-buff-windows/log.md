# Active-Play Buff Windows implementation log

## 2026-08-05 — Codex acceptance review: blocked (A1-A8)

Review by: Codex. Recorded by: Codex.

Checked the active index and save codecs: no other pending RFC claims Company v17. The accepted
active-play direction is attended-only, so opportunity/buff clocks must pause across classified
offline spans. The current draft instead uses wall timestamps and leaves the lazy scheduler absent.

Eight blockers are filed: duplicate/client-authorized click batching, attended scheduler semantics,
cross-runtime RNG/sampling, missing claim intent, incomplete artifact/effect grammar, buff authority
and contribution order, payout saturation/rate ownership, and exact v17 wire/event state. No balance
numbers or rhythm/daemon successor mechanics were implemented.

## 2026-08-06 — owner rulings on acceptance blockers A1-A8 (all resolved)
- A1: NO click_batch (duplicated the shipped perform_manual_batch + trusted client window). click_frenzy
  modifies perform_manual_batch's per-click value; its token bucket is the sole clamp. (AB4 reconciled.)
- A2: all opportunity/buff coords in Company ATTENDED ms (not wall); single lazy scheduler, phase-preserving,
  bounded due transitions; compound-transition/rejection rule like Fiscal F1. (AB1/AB3 reconciled.)
- A3 (owner): server-only t^5*exp float sampler over SplitMix64 substream active_play.spawn.v1 (run seed +
  spawn_seq); RESULT logged, replay READS the schedule (no cross-runtime float recompute). Effect+generator
  selection = integer weight tables over SplitMix64 (byte-identical). (AB1 reconciled.)
- A4: claim_opportunity {intent_id,kind,expected_revision,opportunity_id}; scheduler+accrual first, then
  validate pending+attended-expiry, atomically clear+apply; at most ONE pending, no queue; idempotent retry.
- A5: opportunities artifact schema-v1 (schedule/effect rows/effect keys/combo/click bindings); building
  factor 1+owned*ppm/1e6 (owned=purchased, generator logged); instant_payout REMOVED from v1; cross-validate
  IDs vs economy catalog. Artifact = Company **v18** mint (NOT v17 — Doctrine D9).
- A6 (owner): minimal buff state {buff_instance_id,effect_row_id,selected_target?,activated_attended_ms,
  expires_attended_ms}; factor DERIVED from artifact (not stored); event_buffs source IDs; multiply raw-byte
  instance order, quantize at contribution boundary, clamp product to combo_cap before other slots; expired
  filtered immediately. (AB3 reconciled.)
- A7: Lucky rate = authoritative rate path with all buffs active BEFORE the claim, excluding the pending
  effect; faucet-saturation to cash hardcap (never overflow-reject); zero-rate/zero-bank + quantize vectors.
- A8: Company v18 save keys enumerated (scheduler cursor+spawn_seq, pending opp, active-buff array, clamp
  accounting) + null/empty; strict migration + new-run activation (v18 requires v17 doctrines pinned); exact
  Go/TS wire; raw-byte buff order; big sequential fixture. (AB5 reconciled.)
Status -> accepted; A1-A8 ruled; implementing (Company v18). Body/README reconciled.

## 2026-08-06 — Codex implementation acceptance recheck: blocked (A9-A16)

Review by: Codex. Recorded by: Codex.

Read A1-A8 against the shipped Company save v17 codec, `ApplyLogged` rollback behavior, frozen
contribution inputs, event registry, replay catalog, and Exit activation path. The gameplay shape is
settled, but eight implementation contracts remain open:

- A9: the artifact has no exact nested schema or validation identity;
- A10: seed/addressing, sampler rounding, draw order, and stable IDs are unspecified;
- A11: a lazy scheduler commit beside a rejected command is impossible under the current store;
- A12: state-derived event-buff contributions have no owner inside replay;
- A13: Lucky's saturation method and quantize order are not exact;
- A14: Company v18's keys, nulls, and invariants are not enumerated;
- A15: replay/receipt/event bytes and rejection details are absent;
- A16: artifact activation, first schedule, and Exit reset are not a closed invariant.

The obvious stale v17 references in the Summary and AC5 were reconciled to the already-ruled v18.
Executable proposals preserve rejected-intent immutability, derive active contributions inside the
shared transition, and use the existing runidentity/determinism primitives. No Active-Play code or
balance literals were introduced. Implementation remains blocked pending owner rulings and body
reconciliation.

## 2026-08-06 — owner rulings on the 3rd-round blockers A9-A16 (all accepted)
- A9: exact opportunities artifact {schedule_policy, effects tagged-union, combo_policy}; per-effect keys.
- A10: base=runidentity.Seed(founder,run_seq), Substream(base^spawn_seq,"active_play.spawn.v1"); one Go
  sampler (interval = trusted logged input for TS); Bound(weight)+Bound(gen_weight); UUIDv7 IDs; spawn_seq++.
- A11 (owner): ROLLBACK lazy scheduler on rejection (keeps global invariant, no compound receipt) — and
  Fiscal F11 REVISED to match (both rollback; the wrapper is dropped).
- A12: internal activePlayContributions() owns event_buffs (not in replay_inputs); production_frenzy->
  generators only, click_frenzy->declared actions, building_special->logged generator; Lucky rate excludes
  the pending effect.
- A13: Lucky = Quantize(min(Quantize(frac*bank),Quantize(cap*rate))+eps); accrual-only ledger returns
  (delta,saturated), never rejects; claim always consumes; requested+actual+saturated+cap-reason vectors.
- A14: exact v18 state bytes; A1's manual-token reuse => no clamp-accounting state added.
- A15: replay-inputs bump w/ active_play arm on every applied command; reuse rejection details; exact
  5 event payloads registered Go+DB; event ordering fixed.
- A16: opportunities <=> Company floor 18 (needs meters+achievements+doctrines), forbidden on Founder;
  codec max->18; activation inits empty + schedule from coord 0; Exit discards + re-inits next run.
Status -> accepted; A1-A16 ruled; implementing (Company v18). Fiscal F11 alignment recorded.
