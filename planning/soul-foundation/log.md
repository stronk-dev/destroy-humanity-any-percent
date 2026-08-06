# Soul Foundation implementation log

## 2026-08-05 — Codex acceptance review: blocked (SB1-SB9)

Review by: Codex. Recorded by: Codex.

The review confirms the intended scalar order (Fiscal v19, then Soul v20), runtime independence of
Trust and Soul, and the rule that this foundation may use only fixture source/activity rows. The
existing Founder save already has a dormant signed `soul` field, so v20 must activate and constrain
that field rather than claim to introduce it.

Nine blockers (SB1-SB9) are filed: legacy-field activation, debit eligibility/atomic benefit, floor-price
saturation, the missing opportunity-costed recovery coordinator, exact artifact grammar,
pet/minigame/UI gate ownership, ending-fact/copy-pipeline contradictions, the deferred correlation
gate, and exact wire/upstream sequencing. No unverified research numbers were promoted into data or
player-facing copy.

## 2026-08-06 — owner rulings on acceptance blockers SB1-SB9 (all resolved)
- SB1: v20 ACTIVATES the existing dormant `soul` field (not "adds"); pre-v20 soul==0/reject-nonzero;
  init to soul_initial (<=max) at first activation, recorded in Founder Exit arm. (S1/S6 reconciled.)
- SB2 (owner): Soul-debit is a COMPONENT inside the owner transition (Event/longevity/contract) so
  benefit+debit commit atomically; NO standalone pay_soul_price in production; fixture-only test command.
- SB3 (owner): full-debit-or-reject (unaffordable/soul if it crosses floor); may_exhaust:true source
  consumes to floor once (single-use). No cheap-at-floor exploit.
- SB4 (owner): touch_grass is a REAL activity lifecycle (start: attended start+duration + Company
  production-suppression interval; resolve: Company+Founder txn grants Soul + ZERO output; mutually
  exclusive with production). NOT an instant command. (S3 reconciled.)
- SB5: soul artifact schema-v1 (policy/bands/source rows/activity rows); fixture never epoch-seeded;
  band-threshold partition of full domain; v20 biconditional, requires v19 Fiscal artifact.
- SB6: EXACT bounded meter+band is VISIBLE/published via the pet panel (transparency law — fixes my
  "never a number" error), never a currency/shop wallet; human_content_locked predicate; recovery/
  essential-care available at floor. (S1/S4 reconciled.)
- SB7 (owner): ending keys on a DATED `soul.depleted` Founder fact (emitted first time a debit hits
  floor) + current band, NOT a live threshold (design/02 §7). Copy semantic-safety is a Copy-Pipeline
  successor; export the enum, don't claim the pipeline flags self-harm. (S5/AC5 reconciled.)
- SB8: runtime Trust/Soul independence normative; the CI correlation gate is a balance-harness successor
  (routed, not dropped).
- SB9: after FQ v19 closes, enumerate v20 codec/migration + wire; soul_band_changed.v1 on real band
  change only; DEPENDENCY-BLOCKED on Fiscal v19 implementation.
Status -> accepted; SB1-SB9 ruled; impl dependency-blocked on Fiscal v19. Body/README reconciled.

## 2026-08-06 — Codex implementation acceptance recheck: blocked (SB10-SB16)

Review by: Codex. Recorded by: Codex.

Checked SB1-SB9 against the shipped Founder save/replay, minigame session, pet action, Company accrual,
and multi-stream coordinator surfaces. Fiscal remains an upstream block, but seven independent Soul
contracts are also open:

- SB10: no exact artifact keys, interval convention, or ending/gate grammar;
- SB11: once-only exhaust has no persisted usage authority;
- SB12: the owner-invoked debit component and dated-fact event order are not byte contracts;
- SB13: touch-grass suppression has no persistent Company/replay owner;
- SB14: start/resolve/cancel/reconnect/Exit/concurrency outcomes are undecided;
- SB15: hobby/essential/recovery classifications and composed resolvers do not exist;
- SB16: v20 activation/wire is explicitly deferred rather than enumerated.

The stale acceptance criteria were reconciled to the already-ruled component/activity/dated-fact
model. Executable proposals preserve one replay-owned Soul balance, keep the forbidden public debit
rail absent, and make the real production opportunity cost auditable. No Soul code or data landed.

## 2026-08-06 — owner rulings on the 3rd-round blockers SB10-SB16 (all accepted)
- SB10: exact soul artifact {policy,bands(partition,only near_zero locked),debit_sources(may_exhaust==
  single_use),recovery_activities,ending_policy}.
- SB11 (owner): soul_exhausted_source_ids set in Founder v20 (may_exhaust rows only); reject
  soul_source_consumed after use.
- SB12: pure ApplyDebit(...) component; owner txn persists eligibility+benefit atomically; event order
  benefit->soul_price_paid->[band_changed]->[soul_depleted+LedgerFact]; package-private test entry only.
- SB13 (owner): touch-grass = soul_recovery_sessions table + suppression segment frozen into replay_inputs
  => NO Company bump (the Fiscal F12 pattern); exclusive_activity reject; resolve advances evaluated_through;
  Founder-then-Company lock order (C38).
- SB14: one active session/founder; attended pauses offline; resolve = zero-output segment + Soul + logs
  atomic (Founder-then-Company, corrected from SB14's wording); cancel=0 Soul; Exit rejects while active.
- SB15: minigame artifact soul_gate {human_hobby|unrelated} + pet actions {essential|recovery|ordinary};
  HumanContentLocked() + band projection; composed resolver; near_zero locks human_hobby+ordinary.
- SB16: Founder v20 retains soul key, revalidates [floor,max], appends eligibility/activity state;
  biconditional floor 20 (needs fiscal+minigames+pets); dependency-blocked on Fiscal v19 impl.
Status -> accepted; SB1-SB16 ruled; impl dependency-blocked on Fiscal v19.

## 2026-08-06 — Codex post-Fiscal implementation recheck: blocked (SB17-SB23)

Review by: Codex. Recorded by: Codex.

Fiscal v19 is now implemented, reviewed, and archived, so the named dependency is closed. Re-read the
full Soul RFC against the v19 Founder codec, Exit resolved arms, minigame session coordinator,
`ApplyLogged` accrual path, event registry, and the current pet/minigame artifact grammars.

Seven executable contracts remain absent: literal artifact enum/registry rules; debit errors and event
bytes; recovery-session schema/commands; the zero-output Company replay boundary; exact v20 extension
and activation evidence; versioned pet/minigame gating bytes; and the cross-stream event/log order.
SB16 explicitly asks a future edit to enumerate several of these and the SB10-SB16 ruling section does
not do so. Implementing now would invent public API and transactional semantics. SB17-SB23 are filed
with concrete proposals; no Soul code or data was introduced.
