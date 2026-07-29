# Log — archived-four review

## 2026-07-29 (claude — spec-compliance lens, findings verified independently)

Gates all green including 3× fresh-DB integration (empirically closing R7) and the Depletion
proof executed in both runtimes (max reachable 4 < N=5). **Process fact confirmed: all four
planning logs contain only Codex self-review entries — this round is the missing independent
direction.** Verdict: contract-faithful to an unusual degree; three findings would have blocked
archival. I re-verified A1 and A2 directly before filing.

### Would have blocked archival

- **A1 — DEFECT: the BALANCE-CHANGE baseline guard is honor-system in CI.** Two failures, both
  verified: (a) `ValidateRepositoryBaselineChange` (`server/harness/changeguard.go:54-56`) runs
  only when **HEAD itself** touches the baseline — a baseline rewrite followed by any cover
  commit is never inspected; (b) CI checks out `fetch-depth: 2` (`ci.yml:40`), so the guard's
  `git log -- baseline` sees ≤1 commit and hits `len(commits) < 2 → return nil`
  (`changeguard.go:65-67`) — **a silent no-op in the only place it's enforced**, directly
  contradicting `planning/archive/balance-harness-foundation/log.md:34`'s claim. The
  stale-baseline hash check in `balance-harness/main.go` is sound but enforces freshness, not
  the review protocol. **Fix: deepen the server job's checkout (fetch-depth: 0) and fail LOUDLY
  on truncated history; validate every pushed commit touching the baseline, not just HEAD.**
  Related spec bug: `beb996a` (which fixed same-commit semantics to separate-commit) landed
  inside the commons range — the harness RFC was archived with contradictory guard semantics.
- **A2 — DEFECT: the Commons H-blend the server executes is not the published one.**
  `balance/commons/phase0.json` declares 0.5/0.3/0.2 and `commons.EffectiveHealthPPM`
  (`formula.go:79-89`) implements the RFC's catalog-driven blend with guildless substitution —
  **zero non-test callers** (verified). The live paths hardcode `(cohort*800_000 +
  server*200_000)/PPM` at `commonsprojection/projector.go:381` and
  `harness/commons_population.go:41`. Numerically identical today (0.5+0.3=0.8 under
  substitution), but editing the published balance data changes the generated formula artifact
  and NOT the runtime — violating law 4 (declarative balance data) and law 9 (published
  formulas), and half-failing commons AC7. `gen-formulas` prose also hardcodes "0.5/0.3/0.2".
  **Fix: route both live paths through `EffectiveHealthPPM`; delete the hardcoded blends;
  mutation-test that a weights edit moves the runtime.**
- **A3 — GAP: `balance/commons/` sits outside the balance-change machinery.** `ConstantsHash`
  covers economy-catalog bytes only; the guard recognizes only `balance/catalogs/` and harness
  scenarios. A commons-values edit shifts neither the hash nor the guard. **Fix: fold
  `balance/commons/` into both the hash domain and the guard's input set.**

### Second tier

- **A4:** routes boundary grep (`Makefile:71-72`) checks only `production`; the C2 claim is
  "decimal + context DTO only" — assert the full import set via `go list -deps` in CI.
- **A5:** harness loader silently ignores unknown milestone kinds (`harness.go:409-426`);
  RFC D5 demands a harness failure — only the JSON-schema gate catches it today.
- **A6 — UNTESTED cluster:** harness AC3/AC5 fixtures uncommitted (manual runs only); aggregate
  failures report policyID/seed, not the full run key; production R5's collector→outcome wiring
  in `Service.Handle` never driven end-to-end; `cross_gate` wrong-gate route rejection category
  unpinned; the no-floats guard marshals a toy struct; non-positive factor tested only at zero.
- **A7 — DEVIATION:** the three doctrine-keyed routes sit on `gate.t2_to_t3` while C1's example
  binds the t3_to_t4 doctrine to its own gate — crossed before the doctrine is chosen;
  unobservable until doctrine intents exist; needs a decision record.
- **A8 (theoretical):** commons projection tie-break `(occurred_at, kind-priority, event_id)`
  mis-orders a cross-transaction leave→re-sign on an exact timestamp tie; `Revision` would
  disambiguate. **A9 (documented narrowing):** the corpus gate doesn't ratchet on additions.

### Verified faithful (so the effort is visible)

R4 multiplier tests promoted both runtimes · R5 exported `InvariantSink` verbatim + atomic
event proof · R6 fingerprint-gated formulas (fail-closed, empirically tested) · R7 `-p 1`, 3/3
fresh green · harness single-source (`production.Transition`, zero duplicated math) · SplitMix64
constants exact · golden reports float-free and parallel-deterministic · casual.phase0 as spec ·
C1–C8 all faithful including the Depletion proof genuinely bound to runtime predicates and
first-executor races proven on real Postgres · commons D1–D5 including s_i-zeroed leave, d_i
from accepted contributions only, absent (not 1.0) non-member slot, additive capacity under
advisory lock, and a compliant BALANCE-CHANGE commit.
