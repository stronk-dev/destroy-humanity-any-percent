# Post-Review Integrity Remediation — append-only log

## 2026-07-29 — start

- Owner explicitly requested a large implementation/review cadence rather than serial micro-RFC
  pauses. The independently verified executable findings A2-A6, A8, A9, and D3 are grouped here;
  intermediate commits remain system-coherent and one combined independent review follows the
  full verification pass.
- DESIGN-GAP A7 (doctrine route/gate timing) and S4 (cohort merge capacity policy) alter mechanics.
  They are recorded in the RFC and excluded rather than improvised.
- The already implemented Balance Baseline Change Guard Hardening remains at its independent-review
  gate and will be reviewed together with this batch, not archived early.

## 2026-07-29 — implementation input commit ready

- Projector snapshots and the population harness now call `EffectiveHealthPPM`; snapshot lookup is
  keyed by the revision constants hash. A real-Postgres mutation regression proves retuning the
  weights changes the projector result, and a seeded population regression proves the harness
  path moves too. The generated formula names the ppm arithmetic and fingerprints the function.
- `ConstantsHashArtifacts` hashes sorted, named, length-framed exact bytes. The Phase-0 scenario is
  v2 and binds economy plus Commons artifacts; either byte stream moves the hash. The baseline
  guard accepts Commons as an input but retains artifact-only output commits.
- Harness loading rejects its closed milestone registry at runtime. Actual run output and the full
  report type are checked for JSON floats, and invariant failures print every run-key coordinate.
- The routes boundary now checks every direct internal import against its allowlist. Regressions
  pin negative multipliers, wrong-gate routes, and exact migration-corpus counts.
- The initial in-batch revision comparator was insufficient for the finding's cross-transaction
  case. Migration 00009 persists the highest company revision on membership rows; an older leave
  delivered after a same-time re-sign becomes an idempotent stale event. The Postgres reproducer
  passes.
- Go unit packages, schema semantics, import boundaries, and the complete serial Postgres
  integration suite are green. Harness artifacts intentionally remain stale until this input
  commit lands, after which `make harness-update` creates the required separate balance commit.

## 2026-07-29 — local adversarial pass

- The first route allowlist recipe had `|| true` over the complete import pipeline. That correctly
  tolerated an empty grep result but would also have converted a failed `go list` into success.
  The recipe now captures import enumeration first and fails closed before filtering the allowlist.
- Re-checked the cross-transaction membership reproducer: revision is persisted in the projection
  row, stale events are still claimed exactly once, and founder assignment timestamps use
  `GREATEST` so old delivery cannot move the read model backward.
- No other implementation finding emerged from the spec/adversarial pass. This is a self-review,
  not the mandatory independent approval; neither active RFC is archived on its strength.

## 2026-07-29 — review correction: Routes join the identity bundle

- The first D2 implementation fixed the filed Commons omission but repeated its shape for Routes:
  the production resolver selected a Routes catalog by `constants_hash` even though those bytes
  did not create the hash. A route retune could therefore reuse an old save/run identity.
- D2 is amended before implementation completes. Phase-0 identity now binds economy, Routes, and
  Commons, the scenario is v3 and strict-loads all three, and the guard accepts both feature
  catalog families. Exact-byte tests independently mutate every member of the bundle.

## 2026-07-29 — implementation complete, independent review pending

- Local commits are deliberately layered: `60d5271` is the main implementation input,
  `5c98632` its first artifact-only identity update, `0c6c075` the fail-closed review fix,
  `b7c838f` the Routes identity correction, and `6f8517f` the final artifact-only update.
- The final `make verify` passes with real Postgres: Go vet and every Go/integration package,
  generated-formula drift, Commons population invariance, full-history balance guard, TypeScript
  typecheck/build, 6,412 client tests, strict schemas/boundaries, and 19,245 Chromium/Firefox/WebKit
  tests. The v3 baseline retains identical pacing values; only scenario and constants identities
  changed.
- Worktree is clean. Per the review convention, this RFC and Balance Baseline Change Guard
  Hardening remain `implementing` until a separate reviewer records approval of the combined diff.

## 2026-07-29 — independent review approved

- Independent review approved both this RFC and Balance Baseline Change Guard Hardening with no
  new findings. It specifically verified the fail-closed guard, catalog-driven live formula, all
  three catalog families in constants identity, and the two self-review corrections.
- Owner rulings close the two excluded design gaps: doctrine conditions may bind only gates at or
  after their transition; collapsed-cohort merges remain floor-triggered, never split, and may
  fill a target only through 1.5× its configured size. Those rules are assigned to follow-up RFCs.
- The archived-four remediation board is clean. This RFC is implemented and ready for archival.
