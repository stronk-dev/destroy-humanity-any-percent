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
