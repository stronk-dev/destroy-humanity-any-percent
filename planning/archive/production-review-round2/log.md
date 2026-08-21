# Log — production round-2 review

## 2026-07-28 (claude — combined findings, both lenses)

Spec-compliance + adversarial reviews ran against `4479cd7`..`b1fe65a`. Verdict split cleanly:
**contract fidelity is high (every C1–C8 mechanism behaves as written, scratch-verified), and the
two serious bugs live exactly where the tests are thinnest — both are INTERACTIONS between
individually-correct components.**

### Fix first — demonstrated permanent bricks (adversarial, reproducers ran end-to-end)

- **R1 — HIGH: hardcap headroom clamp overshoots by 1 ulp and bricks the stream.**
  `production/engine.go:97-104` clamps `delta = (cap − balance).Quantize(12)` but
  `economy/ledger.go:145` computes `next = balance.Add(delta).Quantize(12)` — two rounding paths
  that disagree on ~0.04% of near-cap pairs (359/875,142 trials), landing `next > cap` by one
  12-digit ulp. The ledger (correctly) rejects; `Evaluate` errors; **and because every intent
  evaluates first, the stream is permanently bricked at the exact moment a player caps out.**
  Fix: recompute `next` the ledger's way and shrink delta, or saturate-at-cap for production
  entries. Regression vector: the demonstrated cap/balance pair (cap `9.87256122677e8`,
  balance `5.6765610215e6`, 1e0/s).
- **R2 — HIGH: sub-ms cursor phase mismatch permanently degrades `perform_manual_batch`.**
  `EvaluatedThrough` and `ManualTokenRefilledAt` each advance by baseline-relative
  `floor(now − cursor)` ms, preserving their own sub-ms phase; when the phases differ, R can land
  above E and `EncodeState` aborts — **manual clicking fails forever while buys succeed**, and
  nothing can ever repair the phases. `CreateStream` accepts phase-mismatched cursors
  (`save/state.go:78`). Fix: truncate both cursors to whole ms at creation/restore, or advance
  both to the same truncated `now`.
- **R3 — MEDIUM: degenerate `resource_log` target diverges Go vs TS.** Both loaders accept
  `target < ~5e-15` (passes `> 0`); `1 + target` collapses to 1, `log10 = 0` — then
  **Go `Div`-by-zero returns Zero (silent progress 0 forever) while break_infinity yields
  Infinity (TS throws).** Verified at `decimal/decimal.go:275-277`. `phase0.json` unaffected
  (targets ≥ 1e3). Fix: loader floor on target **and** reconcile the `Div`-by-zero primitive with
  the oracle — the primitive divergence is a landmine beyond progress.

### Then — coverage and integrity debt (spec review; implemented correctly but unasserted)

- **R4:** `server/multiplier` has zero test files; single-provider, duplicate-source, and
  within-slot permutation invariance verified only by reviewer scratch tests. Promote them
  into the repo (Go + the TS mirror at `economy-kernel.ts:223-225`).
- **R5:** the invariant path (C7) has zero coverage — nothing asserts an `invariant_reported`
  event is written or that `residual_abort` reaches the sink. Also: the specified exported
  `InvariantSink` interface doesn't exist (unexported slice instead), and a rejected `max`
  purchase with `UsedFallback` gets log+metrics with no event — C7 reserves the event-less path
  for `residual_abort` only.
- **R6 — law-9 integrity:** `gen-formulas` hand-writes `production_rate` and `within_slot_order`
  as string literals — only `multiplier_slot_order` is genuinely drift-gated. If `Rates` changed
  shape, `formulas-check` stays green. Generate from live paths or narrow
  `docs/production-engine.md:15`'s claim.
- **R7:** `make test-save-integration` races its own migrations on a fresh DB (reproduced:
  `pg_type_typname_nsp_index` duplicate key) — first-run flaky; serialize migration in the target.
- **R8, minor:** `internal_invariant` never emitted (fine until transport; currently unowned) ·
  corpus-only-grows unenforced (`state_test.go:112` asserts only `len >= 2`) · events outlive
  pruned revisions (docs honest; the RFC sentence is stronger than the schema — amend one) ·
  receipt `wireChanges` swallows parse errors (`intents.go:352` — propagate) · sticky
  malformed-intent-then-corrected-retry yields `idempotency_conflict` forever (confirm intended).

### Held under attack (recorded so the effort isn't invisible)

Idempotency under 16-way concurrency (exactly 1 applied; provably one transaction) · token
bucket same-ms/regression/overflow · accrual partition at the exact 86,400,000 boundary and
one-year offline · multiplier runtime rejections + deterministic order · **6,000-case Go/TS
differential: 0 mismatches** (new paths route through the fixed primitives) · migration chain
refuses lying versions in both directions · no `time.Now()` outside injected clocks ·
`phase0.json` validates against every loader rule.

## 2026-07-28 (Codex — reproducer verification and RFC handoff)

- **R1 confirmed** against the filed cap/balance pair through `production.Evaluate`: ledger commit
  rejects `company.cash` above its hardcap. Direct probing showed nominal headroom
  `9.81579561656e8` produces `9.87256122678e8`; one lower 12-digit ulp,
  `9.81579561655e8`, reproduces the exact cap. Drafted `rfc/production-hardcap-saturation.md`.
- **R2 confirmed** from a state that is legal before mutation: E=`…00.1009`, R=`…00.1001`,
  now=`…00.1015`. Production advances 0 ms, refill advances 1 ms to R=`…00.1011`, and
  `EncodeState` rejects R > E. Drafted `rfc/millisecond-cursor-canonicalization.md` with save v4.
- **R3 trigger confirmed; primitive diagnosis corrected.** Installed `break_infinity.js` 2.2.0
  returns zero for `1 / 0` and `0 / 0`, matching Go and the existing mandatory golden vectors.
  The divergence is TypeScript `resourceLogProgress` using native number `/`, which produces
  Infinity when the denominator collapses. Both aligned-add implementations show `4e-15` → zero
  denominator and `5e-15` → positive. Drafted `rfc/resource-log-domain-parity.md`; it preserves
  Decimal division semantics, rejects targets below `5e-15`, and aligns the evaluator operator.
- Scratch repro files were removed after the probes. No implementation is authorized until the
  three follow-up RFCs are accepted.

## 2026-07-29 (claude — R3 diagnosis corrected; three micro-RFCs verified and accepted)

**Correction to the R3 entry above, verified independently:** the adversarial reviewer's
primitive-level claim was **wrong**. `break_infinity.js` returns **Zero** on division by zero
(`1e5 ÷ 0 = 0`, run directly against the pinned library), matching Go — and both golden-vector
suites already enforce it (`div-zero`, `zero-div-zero` categories). There is no primitive
divergence. The real bug is in the TS **progress evaluator**: `economy-kernel.ts:334` computes
`.log10() / .log10()` — *native* JS division of the two native numbers `log10()` returns — which
is where Infinity enters. Codex's RFC therefore correctly: preserves the primitive untouched,
floors `resource_log` targets at ≥ 5e-15 in both loaders, and routes the TS operation through
Decimal division. The R3 trigger (degenerate target accepted by both loaders; Go-silent-0 vs
TS-throw) stands as demonstrated; only the diagnosis moves.

Lesson for the review protocol: the reviewer demonstrated the *symptom* end-to-end but attributed
the cause by reading the wrong layer. **Codex re-verified against the pinned library and the
vector corpus before drafting the fix — that re-verification step is now part of the pattern.**

All three micro-RFCs reviewed against the findings and the reproducers: accepted 2026-07-29
(standing owner mandate). Implementation order per Codex: R1 → R2 (save v4) → R3, then R4–R8
before the harness's first baseline.

## 2026-07-29 (Codex — R4–R8 contract draft)

- Confirmed R1–R3 are implemented, archived, and independently approved. No accepted active RFC
  specified R4–R8, so implementation did not start from review prose.
- Drafted `rfc/production-contract-integrity.md` as one bounded follow-up covering the multiplier
  assertions, invariant-path contract, formula provenance gate, integration-test migration race,
  and minor integrity debt. Added it to the active RFC index as `draft`.
- Resolved the C7/C4 collision explicitly: invariant reports become persisted events only when a
  committed gameplay revision exists. Terminal rejections and internal aborts use structured audit
  output and metrics without fabricating a revision.
- Chose a source-bound formula artifact rather than claiming automatic algebra inference: schema
  v2 fingerprints normalized AST for the live rate path and ordering authorities, so executable
  changes force reviewed regeneration while comments and formatting do not.
- Kept transport mapping and deployed-process migration coordination out of scope with named
  owners. No code, planning directory, or harness baseline is authorized until owner acceptance.

## 2026-08-21 — planning lifecycle closeout

The four dedicated successors named by this review—Production Hardcap Saturation, Millisecond
Cursor Canonicalization, Resource-Log Domain Parity, and Production Contract Assertions &
Integrity—are implemented and archived. This source thread therefore moved to
`planning/archive/production-review-round2/` under RP-102. Their own plans, logs and designated
verdicts remain authoritative for implementation completion; this move does not replace them.
