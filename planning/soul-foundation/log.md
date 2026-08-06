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

## 2026-08-06 — owner rulings on SB17-SB23 (all accepted)
Artifact enums + copy-key registry (SB17, ending values are enum IDs never copy keys); ApplyDebit
typed errors + 3 exact event payloads (SB18); the 3 coordinator commands + UUIDv7 session row +
partial-unique active-session index (SB19); ApplySuppressedLogged — the shared zero-output boundary
that advances every watermark while asserting zero output (SB20); Founder v20 = v19 + exactly
soul_exhausted_source_ids, next_soul Exit arm recomputed-from-bytes (SB21); soul_gate via artifact
schema BUMPS, historical schemas valid only in Soul-less bundles (SB22); start/cancel/resolve event
ordering + Founder-then-Company single-transaction atomicity w/ fault injection (SB23).
Status -> SB1-SB23 complete; Soul implementable (Fiscal v19 archived).

## 2026-08-06 — implementation round: persistence/replay complete; recovery liveness DESIGN-GAP

Review by: pending cross-party designated review. Recorded by: Codex.

Implemented the strict Soul package and v20 activation, the pure debit component and exact events,
versioned pet/minigame gates, and the recovery persistence/replay skeleton. The recovery round adds:

- an immutable, claim-tokened `soul_recovery_sessions` lifecycle and one-active-session constraint;
- race-safe transaction guards for Company commands, Founder care, Exit, and minigame start;
- the shared Go/TypeScript `ApplySuppressedLogged` zero-output transition;
- Company-first then Founder-log resolution with one durable terminal receipt;
- exact event registry/database constraints, v20 migration/corpus, cross-runtime replay fixtures,
  full Founder-history verification, and real-Postgres fault injection at every persistence boundary.

The real database suite exposed and fixed three seam defects that unit tests did not cover: the
Founder-log multistream constraint initially excluded the new recovery arm; start and terminal events
share a session ID and therefore Founder-history event loading must scope by applied revision; and
retention attempted to delete immutable Founder genesis revision 1 after the sixth Founder command.
The shared retention helper now preserves the genesis revision in every write/coordinator path.

**DESIGN-GAP — recovery cannot accumulate attended time under the ruled exclusivity model.** SB14 says
ordinary Company commands reject while recovery is active, early resolve rejects without mutation,
and recovery eligibility advances by Founder attended time while pausing offline. The shipped
attendance resolver derives the current-run partial from Company state; that state advances only on a
successful Company transition, and any unresolved gap above the pinned catch-up ceiling is classified
offline. With every ordinary transition rejected, no legal writer advances `evaluated_through` or the
attended partial. A fixture can resolve at exactly the tolerance boundary, but one millisecond later
the gap becomes offline and the session can never become eligible. That exact-boundary fixture is not
accepted as proof of a live mechanic.

Required owner contract: name a server-authoritative way to advance recovery attendance while keeping
production at zero and preserving `rejected = no mutation`—for example, a claim-tokened recovery
presence/progress command or a composed socket-presence interval written through the suppressed
boundary. It must specify reconnect/offline classification, idempotency, replay bytes, and whether
progress commits before terminal resolution. No production `recovery_activities` row may mint until
that contract lands. Implementation remains active and is not ready for archival.

Verification completed from the repository root: replay fixture regeneration/check, focused Go and
TypeScript suites, `make test-save-integration` against the declared Postgres service, and the full
`make verify` gate. The full gate read through completion: Go vet/all packages, balance harness,
formula and schema drift, kernel history/parity at `0.3.73`, TypeScript/Svelte typecheck, production
client build, 6,582 unit assertions, and 19,755 browser assertions all passed. This is self-
verification only; it does not satisfy the cross-party designated review gate and does not authorize
archival.
