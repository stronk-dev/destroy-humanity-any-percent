# Doctrine & Compute Credit implementation log

## 2026-08-05 — Codex acceptance review: blocked (D1-D6)

Review by: Codex. Recorded by: Codex.

The implemented substrate was checked against the draft: `DoctrinesByTransition` and
`ComputeCreditMS` exist in the Company save; Route predicates consume committed doctrine choices;
the pinned economy catalog already carries the provisional burst parameters; and both new commands
can ultimately live inside `ApplyLogged`. That does not close the behavioral contracts.

Six blockers are filed in the RFC: doctrine artifact/activation authority (D1), pick/gate ordering
(D2), promised-but-deferred doctrine effects (D3), exact acceleration equation and cursor lifecycle
(D4), auto-spend ownership/persistence/replay (D5), and exact wire/event/rejection grammar (D6).
Implementation is intentionally not started. Owner rulings must reconcile the normative body, not
merely append answers.

## 2026-08-05 — owner rulings on acceptance blockers D1-D6 (all resolved)
- D1: hash-pinned `doctrines` artifact v1 {transition_id, source_tier, gate_id, doctrine_ids[]};
  epoch mint, activation-boundary; both loaders require doctrine-bearing Route predicates to resolve
  to a doctrine row. Effects/unlocks deferred to a successor.
- D2: pick legal only while tier==source_tier, before gate crossed, no recorded choice; cross_gate
  rejects doctrine_required; pick is a prior committed intent; repeated intent-ID replays, new
  intent-ID on an already-picked transition rejects doctrine_already_picked.
- D3: persist the CHOICE only; DC1 effect sentence removed (body reconciled); effects → successor.
- D4 (owner ruling): PERSISTENT acceleration-burst, NOT instant — matches the existing
  burst_speed/burst_max_duration_ms catalog fields. Burst state is Company-scoped => REQUIRES a
  Company save-version bump (corrects the draft's "no save bump" claim). SECOND Company-version
  claimant alongside Active-Play Buff Windows — sequence or atomically co-activate (meters v15 +
  achievements v16 pattern) at implementation. Rate×burst_speed over accrual, Exit resets active
  burst, F1 horizon precondition applies. DC2/DC3 reconciled.
- D5 (owner ruling, option a): defer auto-spend to a successor settings RFC; this intent is
  MANUAL-ONLY; HUD shows balance only. DC2 reconciled.
- D6: enumerate intents/receipts/events byte-for-byte; positive safe-int amount_ms; reuse existing
  rejection reasons; register both event kinds; shared Go/TS fixtures incl. burst-boundary cases.
Status → accepted; implementing. RFC body + README reconciled.

## 2026-08-05 — implementation acceptance recheck: blocked (D7-D10)

Review by: Codex. Recorded by: Codex.

The D1-D6 rulings were re-read against the live engine before implementation. Their direction is
sound, but four executable gaps remain and are filed in the RFC with proposed contracts:

- D7: persistent burst selected, but credit-to-duration arithmetic, stacking, cap behavior,
  offline interaction, and rounding remain unspecified;
- D8: the production doctrine row/gate cannot be authored from Phase-0 data — the committed route
  history explicitly says the T3-to-T4 gate and permit resource do not exist;
- D9: current production is a seven-artifact Company-v14 epoch, while v15/v16 may activate only as
  the unminted Meters/Achievements pair; a doctrine-only v17 mint would skip that binding chain;
- D10: D6 says the wire is enumerated but supplies no exact objects, and AC3 still requires the
  auto-spend setting D5 deferred.

No implementation commit was started. This is the required DESIGN-GAP bounce: choosing any of
these in Go first would make the TypeScript verifier and immutable replay history follow an
implementation accident rather than an owner-ratified contract.
