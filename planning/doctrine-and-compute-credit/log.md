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

## 2026-08-05 — owner rulings on the second-round blockers D7-D10 (all accepted)
The D1-D6 rulings chose direction but left immutable mechanics open; Codex bounced D7-D10 with exact
proposed contracts. All accepted:
- D7: burst equation — amount_ms is the credit debit AND 1:1 boosted duration; reject (no clamp) on
  over-cap/low-balance/burst-active (no stacking); persist compute_burst_remaining_ms; burst_speed>1
  from the pinned economy artifact; bonus = a second segment at (burst_speed-1) through the shared
  fixedgrid, quantized once; offline bounded by accrual_cap_ms; Exit resets to zero; partition +
  25h-return fixtures.
- D8: ship schema+validator + a FIXTURE doctrine row (transition.t3_to_t4/source_tier 3/gate.t3_to_t4/
  [capture,ethical]); production doctrines artifact UNMINTED until the T3->T4 content owner supplies
  the real gate+rows; cross_gate requires the pick only at the declared gate when tier==source_tier.
- D9: Company v14->v17 linear activation at Exit, ONLY in a meters+achievements+doctrines epoch;
  versionFloors requires all three; v17 keys = v16 keys + compute_burst_remaining_ms; codec built now,
  NOT minted. **Active-Play sequences at Company v18** — settles the cross-RFC coordination; I carry
  v18 into the A-blocker rulings.
- D10: exact request/event/snapshot/rejection grammar enumerated; AC3 replaced (no auto-spend);
  AC2 reconciled to the burst model.
Lesson carried forward: for "owner-ruling-required" blockers, provide the FULL contract in one pass
(not just the direction) to avoid a second bounce — the remaining Wave-A RFCs get exact rulings.
Status -> accepted; D1-D10 ruled; implementing. Body/README reconciled.

## 2026-08-05 — doctrine artifact and activation-boundary landing

- Added the strict schema-v1 doctrine-choice catalog in Go and TypeScript with one shared parity
  corpus containing the ruled T3-to-T4 fixture row. The loader enforces adjacent transition/gate
  identity, source-tier equality, sorted unique branching choices, and exact JSON keys.
- Added cross-artifact validation: every doctrine-bearing route predicate resolves to the pinned
  doctrine catalog, and every doctrine choice names a gate present in the pinned routes catalog.
- Replay bundles now admit `doctrines` only above the paired Meters/Achievements artifacts. The
  optional doctrine axis composes independently with Founder-only minigame/pet artifacts, and
  derives Company floor v17 without changing the Founder floor.
- No production doctrine artifact or epoch was added. D8/D9 remain honored: the mechanism and
  codec surface ship pre-mint; T3-to-T4 content owns activation.
- Kernel semantics advanced to 0.3.57 and the new Go/TypeScript doctrine authorities joined the
  fail-closed watched-path registry in lexical order.

Focused evidence from repository root: `make test-go GO_PACKAGES='./doctrine ./replaycatalog
./production'`, `make verify-schema`, `pnpm --dir client typecheck`, and the doctrine/replay Vitest
files all pass.

## 2026-08-05 — Company save v17 burst state

- Added the scope-specific Company-v17 envelope: v16 foundation fields plus the required exact
  integer `compute_burst_remaining_ms`. Founder v17 remains the minigame envelope; the two scopes
  cannot decode or encode each other's fields.
- Preserved `LatestSupportedVersion=16` as the historical foundation alias used by existing tests
  and Founder carry. `LatestCompanyVersion=17` is the new Company-axis authority; genesis accepts
  it independently from the Founder-v18 ceiling.
- Go and TypeScript reject missing burst state, state above the pinned economy duration cap on
  restore, burst state before v17, and Company/Founder feature leakage. New-run activation derives
  v17 only from a pinned doctrines artifact and initializes the remaining duration to zero.
- Kernel semantics advanced to 0.3.58. Focused root evidence: `make test-go
  GO_PACKAGES='./save ./production ./replaycatalog'`, `pnpm --dir client typecheck`, and
  `pnpm --dir client test -- replay.test.ts` pass.

## 2026-08-05 — Doctrine intents and Compute Credit burst implementation

- Commit `277b43b` adds both exact intent grammars to the Go/TypeScript replay boundary, strict
  `doctrine_picked` and `compute_credit_spent` payload validation, and the Postgres event-kind
  expansion. Doctrine choice is write-once at the declared source tier and blocks its exact gate
  until committed.
- Compute Credit spending is manual-only and exact: `amount_ms` is the debit and duration; active,
  insufficient, and over-cap requests reject without clamping. Evaluation consumes wall duration
  online or offline and integrates base plus `(burst_speed-1)` bonus inside each fixed provision
  segment before one explicit quantization boundary. Offline bonus stops at the ordinary accrual
  cap while later wall time may expire the burst. Exit clears the active remainder.
- The Go-authored shared replay artifact now contains one sequential eleven-command doctrine/burst
  run consumed by TypeScript. It covers the unpicked gate, applied and duplicate choice, activation,
  active/balance/duration rejection, partial and exhausted consumption, malformed input, and the
  final legal gate. Regeneration exposed and fixed a fixture-bundle seam: the fixture's added gate
  also had to join the pinned category gate set before the TypeScript loader would accept it.
- The first arithmetic draft quantized ordinary and bonus deltas separately. Pre-commit diff review
  tightened this to the ruled single bucket boundary in both runtimes; the shared final-state bytes
  now pin that corrected equation.
- Postgres initially rejected the migration set because the already-committed Founder link owned
  sequence 62. Commit `1b80275` renames the new, never-applied Doctrine migration to 63 with its SQL
  body byte-identical. A later SQL migration could not repair a duplicate that Goose refuses to
  enumerate. The full disposable-Postgres suite then passed.

Focused evidence from the repository root: `make test-go GO_PACKAGES='./production ./save
./doctrine ./replaycatalog'`, `make test-client` (6,566 passed, 3 skipped), `make typecheck`,
`make verify-schema`, `make verify-kernel-version` (0.3.59), and `make test-save-integration` all
pass. Canonical docs were updated in the following record commit; no production doctrine artifact,
epoch mint, archival, push, or deployment occurred.

## 2026-08-05 — full root verification and Codex full-span self-review

Review by: Codex (self-review; not an independent archival gate). Recorded by: Codex.

Reviewed range: `2262eab..5e279bd`.

- Added a direct terminal-transition regression proving Exit clears the active burst in both the
  ending Company state and its new-run state (`345b913`).
- The first full `make verify` correctly failed because the generated production-formula artifact
  still carried its pre-burst source fingerprint. The generator changed only that fingerprint; no
  formula or balance literal changed. Commit `3c9b818` records the generated artifact.
- Full-span self-review found one acceptance-coverage gap after the green gate: burst replay parity
  and fixed-grid provisioning were each tested, but no case combined a burst with a real provision
  boundary. Commit `5e279bd` adds a 90-second one-shot-versus-split regression across the absolute
  60-second grid and pins identical cash, provisioned units, and exhausted burst state.
- Re-ran the complete root `make verify` at `5e279bd`: green across Go vet/tests, deterministic
  harness checks, TypeScript/Svelte typecheck, production build, 6,566 client tests, 19,707 browser
  tests, formula/schema/boundary generation checks, and kernel-history fixtures.

No correctness finding remains from the self-review. Its one coverage finding was fixed inside the
reviewed range rather than carried forward.

No production artifact, epoch mint, archive, push, or deployment occurred. The implementation is
ready for the required independent full-range reviews; this self-review does not satisfy them.
