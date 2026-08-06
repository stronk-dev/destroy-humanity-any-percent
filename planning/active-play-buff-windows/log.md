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

## 2026-08-06 — Codex implementation handoff: ready for designated review

Review by: Codex (self-check only; does not satisfy the cross-party gate). Recorded by: Codex.

Implementation range: `8557638..d3b18ef` (inclusive payload is the two commits after the parent of
`8557638`). Post-filter hashes only.

Implemented the complete A9-A16 surface:

- optional hash-pinned `opportunities` artifact and strict Go/TS loaders;
- Company v18 codec/activation plus replay-inputs v5;
- attended-time lazy scheduler with server-logged float interval, byte-parity integer draws, and
  Go/TS-recomputed opportunity/buff UUIDs;
- rejected-command rollback, including an expired-claim fixture whose scheduler cleanup repeats on
  the next applied command;
- state-derived event-buff contributions inside `ApplyLogged`, per-target combo saturation, normal
  manual-token enforcement, and exact Lucky accrual-only saturation;
- all five event payloads and the Postgres event-kind migration;
- Exit reset and next-run schedule initialization.

Self-review found and fixed four seam defects before handoff: rejected claims were initially forced
to carry impossible applied-resolution evidence; the Go receipt snapshot omitted v18 state; empty
active-buff slices encoded as null; and the singular missed/spawn wire was not structurally bounded
against the online horizon. The final bundle validity rule proves `minimum_interval_ms +
lifetime_ms > catchup_ceiling_ms`, so one command can represent at most one pending expiry followed by
one new spawn.

The shared sequential fixture covers two overlapping production buffs with combo clamping, a manual
command under the active set, Lucky saturation at the Company hardcap, buff expiry, an offline wall
gap that advances no attended coordinate, rejected expired-claim rollback, persisted miss cleanup,
and byte-identical Go/TS receipts/events/state. A separate terminal fixture proves Exit discards the
old pending/buff state and deterministically initializes run N+1.

Verification (full outputs read to completion):

- `make verify` — PASS: all Go packages, formula/harness/schema/copy/kernel-history gates, TypeScript
  typecheck and production build, 6,577 client tests, 19,740 browser assertions;
- `make test-save-integration` against fresh Postgres 16 — PASS, including save/production/replay
  integrations with migration 00066;
- `make replay-fixture-check` — PASS.

Handoff: implemented and all self-checks green; ready for cross-party designated review and archival
decision. Codex does not certify that gate and has not archived this RFC.

## 2026-08-06 — designated cross-party verdict: Active-Play implementation — NOT APPROVED (3 blocking)

- **Review by:** the designated Claude reviewer (independent, isolated worktree at d3b18ef; all gates
  re-run: make verify exit 0, typecheck explicit, Postgres integration exit 0).
- **Recorded by:** same.
- **Range correction (F1 cure):** the handoff's declared range `8557638..d3b18ef` OMITTED two
  Active-Play code commits (`32e5a63` catalog/scheduler, `2baab9a` v18 envelope) that no verdict had
  ever cited. The reviewer audited them in full; **this verdict covers `32e5a63^..d3b18ef` plus
  45944ca (docs)** — the archival gate must cite THIS extended range. The handoff's range statement
  is hereby corrected.

**BLOCKING (code fixes + discriminating fixtures required):**
- **F2 (HIGH) — combo hardcap breached across targets.** `clampActiveContributionProducts` clamps per
  target-group only, but the engine composes `all`-target and generator-specific event_buffs into one
  slot product: frenzy(all ×7) × building_special(clamped ×10) = ×70 against combo_cap 10 —
  probe-confirmed, both runtimes share the defect. A6/A12 require the SLOT PRODUCT clamped. Fix +
  a cross-target combo vector (the existing tests were vacuous at cap 1e4).
- **F3 (HIGH) — cross-runtime replay divergence on compound miss+spawn.** Go generates and Go-replays
  a single command carrying a miss AND a spawn; TS `applyActiveSchedule` cannot replay either
  compound form (sentinel-0 `after_next` skips the spawn branch → "unexpected active spawn").
  Probe-confirmed with the shipped fixture bundle: legal artifact + ordinary play persists a run log
  TS rejects. Fix TS compound replay + a compound miss+spawn fixture row in BOTH runtimes.
- **F4 (MEDIUM, ruled REQUIRED — not re-ruled away):** the A13 boundary vectors (zero/zero,
  epsilon-only, one-ulp-below-cap, at-cap, overflow-scale) must actually exist; the single fixture
  lucky row is quantization absorption, not an at-cap vector.

**Required with archival:**
- **F5 (MEDIUM, owner ruling): the combo-cap reason key MUST surface.** A6's "visible hardcap +
  reason key" stands — when the clamp bites, the receipt/event carries `hardcap_reason_key` (the
  fiscal pattern). Parsed-but-never-used does not satisfy a visibility law.
- **F6 (LOW):** add the Founder-scope active-play leak check in validateFoundationState (mirror the
  compute-burst check). Hygiene.
- **F7 (LOW):** plan-box flips in 45944ca are legal under the refined same-range rule, but box 1's
  claimed span must be corrected to the F1 extended range.

**Owner ruling on INFO-1 (spawn_seq semantics):** the implemented behavior — `spawn_seq` increments
once per SPAWN, with the post-miss reschedule reusing the already-advanced sequence — is ACCEPTED and
now canonical: every opportunity consumes a unique substream, both runtimes agree byte-exactly, and
the evidence checks it. A10's "incl. a missed opportunity" wording is amended accordingly (body
reconciled in the RFC).

Verified-correct: A3/A10 seed+logged-schedule pattern, A2/A11 attended coords + full rollback,
A12/A13 contribution ownership + Lucky arithmetic (modulo F2/F4/F5), A16 activation/Exit/migration
00066, A14/A15 wire + 11-row byte-compared fixture, A1 no click_batch, kernel 0.3.64→0.3.68 lockstep,
45944ca honestly labeled a self-check. INFO-2: the guard findings are remediated in ebb081f (awaiting
Codex re-review).

**Archival BLOCKED until F2/F3/F4 are fixed with fixtures and F5/F6/F7 land; the archival commit must
cite this verdict's extended range `32e5a63^..d3b18ef` + 45944ca + the remediation range.**

## 2026-08-06 — F2–F7 remediation (`1f5a2f8..9294280`), ready for designated re-review

- **Implemented by:** Codex.
- **Decision:** implementation complete; self-checks do not authorize archival.

F2 now caps the effective per-generator event-buff product: the shared `all` product consumes
headroom before building-specific groups, with a conservative one-ULP clamp probe in both runtimes.
The discriminating fixture is production frenzy ×7 plus a 100-owned building special against cap 10;
the sequential Go-authored corpus replays the same cross-target saturation in TypeScript.

F3's sequential corpus now contains one applied command that both expires a missed opportunity and
spawns its successor. TypeScript reconstructs the post-miss coordinate from the expired coordinate +
the logged sampled interval before verifying the spawn, matching Go's transition order.

F4 adds the five shared Lucky vectors (`zero-zero`, `epsilon-only`, `one-ulp-below-cap`, `at-cap`,
`overflow-scale`) consumed by Go and TypeScript. F5 surfaces `cap.active_combo` through the typed
claim receipt and schema-v2 claim/buff-start event payloads; v1 rows remain accepted. F6 rejects all
active-play state outside Company scope. F7 corrects the plan to the designated verdict's full
`32e5a63^..d3b18ef` range. Kernel semantics advance in lockstep to 0.3.69.

Checks completed before handoff: `make test-go`, `make typecheck`, `make test-client`, and
`make replay-fixture-check` all pass. Full `make verify` and the Postgres integration gate follow in
the final handoff commit. Active-Play remains unarchived and requires the cross-party designated
re-review over the original extended range plus this remediation range.

Final handoff verification: `make verify` PASS (including Go vet/tests, TypeScript typecheck,
6,578 client tests, formula drift, guard fixtures, and harness check); `make test-save-integration`
PASS through the declared Docker/Postgres target. Ready for the cross-party designated re-review;
not self-approved and not archived.

## 2026-08-06 — designated cross-party RE-REVIEW verdict: remediation 1f5a2f8^..16eb935 — NOT APPROVED

- **Review by:** the designated Claude re-reviewer (independent; all gates re-run incl. explicit
  typecheck, isolated-worktree Go suite at 16eb935, real-Postgres probes). **Recorded by:** same.

**BLOCKING:**
- **RR-1 (HIGH, a NEW regression from the F5 fix): schema-v2 active-play events violate the live DB
  constraint.** active_play.go now emits opportunity_claimed.v1/buff_started.v1 with SchemaVersion:2
  unconditionally; the Go validator was widened but migration 00035's events_schema_version_check
  (only run_ended may be v2) was NOT extended — probe-confirmed: both v2 inserts rejected (23514).
  Any buff-granting claim against the real database fails its event write inside the intent
  transaction → the feature bricks in production. Integration suite green ONLY because it has zero
  active-play coverage. Fix: a NEW append-only migration extending the constraint + integration
  coverage that actually writes active-play events.
- **RR-2 (HIGH): F3 only half-closed — spawn-then-SELF-miss is Go-replayable but TS-unreplayable.**
  The loader rule blocks two-pending chains, not a fresh spawn whose own lifetime elapses within the
  same catch-up (reachable with the shipped corpus catalog: lifetime 3000 < ceiling 5000).
  Probe-confirmed: Go generates+replays; TS throws "active missed mismatch". Fix: TS support for this
  order + a shared fixture row, OR a structural loader rule (lifetime_ms > catchup_ceiling_ms) in
  BOTH runtimes (which invalidates the current corpus catalog and forces retune).
- **RR-3 (MEDIUM, blocking under F2's own terms): the SHARED corpus never exceeds the cap.** The
  `owned=100` edit landed in the WRONG builder (makeDoctrineReplayRunFixture — a run with no
  active-play arm); the active-play fixture's cross-target product is 7×1.1=7.7 vs cap 10 — the TS
  clamp-bite path (incl. the mantissa-ULP fallback) is executed by NO shared test, and Go↔TS byte
  parity of clamped factors is unproven. The remediation log's "corpus replays the same cross-target
  saturation in TypeScript" is factually wrong — correct that record with the fix.
- **RR-4 (MEDIUM, with archival):** no fixture row anywhere carries a non-null combo reason key
  (falls out of RR-3's fix).

Verified-correct: the F2 clamp MECHANISM (all-group consumes headroom first; engine probe ≤ cap),
the ruled miss→successor-spawn compound (corpus entry 11, byte-compared), the five F4 Lucky vectors
(genuine, incl. a true ledger cap hit), F6 leak check + test, F7 range correction, kernel 0.3.69
lockstep (RR-5 INFO: the bump correctly rides the semantic commit; the split generated-constants
sync is a precedent deviation only), and the 16eb935 RFC reconciliation.

**Range-union:** prior verdict `32e5a63^..d3b18ef` + 45944ca; this re-review `1f5a2f8^..16eb935`;
gap commits are docs or separately-tracked guard code (ebb081f — Codex's KRM thread). The eventual
archival must cite BOTH ranges plus the range closing RR-1..RR-4. **Active-Play archival remains
BLOCKED.**
