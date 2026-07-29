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

## 2026-07-29 (claude — adversarial lens; both HIGHs verified at source before filing)

Reproducers ran against real Postgres; each finding below quoted actual output. I re-verified the
two HIGH code sites directly (membership-check-before-dedup at `commonsprojection/projector.go:214`;
first-executor `ON CONFLICT DO NOTHING` with no occurred_at comparison at
`routeprojection/projector.go:104`). **Combined priority for remediation now leads with these.**

### Demonstrated

- **D1 — HIGH: commons projection is not retry-idempotent; a sampled event permanently poisons
  the worker after the leave lands.** `projectSample` checks active membership **before** the
  event-id dedup insert — the mirror image of `project`, which correctly dedups first. Retry of
  an already-committed `[signed, sampled, left]` batch (at-least-once delivery, or crash between
  commit and cursor advance) errors `sample without active membership`; `Project` aborts the
  whole batch; **the worker wedges forever, and a full-history rebuild fails the same way.** The
  existing integration test replays samples only *before* the leave — exactly the gap. Fix: move
  the dedup insert/early-commit above the membership lookup, mirroring `project`.
- **D2 — HIGH: Registry first-executor is delivery-order, not event-order; replay-from-scratch
  disagrees with the live projection.** The `(occurred_at, event_id)` sort orders only *within*
  a batch; different company streams arrive in different batches, and `ON CONFLICT DO NOTHING`
  awards whoever *projects* first. Demonstrated: B (15:00, projected first) beats A (14:00) —
  and the irreversible 100-knowledge `registry_first` grant follows the wrong founder. A
  from-scratch rebuild yields A: **the projection is non-convergent**, the one property a
  projection must have. Fix: on conflict compare `(occurred_at, first_event_id)` and displace
  (compensating the grant per C4's compensation events), or funnel all `route_executed` through
  one ordered cursor.
- **D3 — MEDIUM: independently re-demonstrates A2** (the dead catalog weights) with a mutation:
  retuned weights move `EffectiveHealthPPM` (940000) but not `Snapshot` (920000). One fix
  covers A2+D3: `Snapshot` calls `EffectiveHealthPPM` with the event's resolved catalog — note
  it currently doesn't even take a constants hash, so the fix has a plumbing component.
- **D4 — MEDIUM: sharpens A1 — a `BALANCE-CHANGE:` commit can smuggle arbitrary code.** The
  guard never inspects what else the baseline commit touches; demonstrated a baseline rewrite +
  `server/code.go` change passing with verdict nil. Fix folds into A1: the baseline commit must
  touch *only* the baseline, and the guard must fail loudly on truncated history.

### Suspected (seams to close, no exploit today)

- **S1:** the Depletion proof's soundness rests on exclusion-slot values being **immutable
  within a run** — nothing writes `StructureID`/doctrines outside save-decode *today* (verified
  by grep), but any future intent that does silently unsounds the load-time proof. Assert
  immutability in the engine (reject writes once any gate is crossed) or record slot values at
  first crossing. **Route this to the Prestige/Exits and future doctrine-intent RFCs as a
  standing constraint.**
- **S2:** same-millisecond samples never move smoothed Health (converges next spaced sample;
  cosmetic). **S3:** solidarity window over-counts ≤1h of coverage (≤0.14% at the 30-day window;
  a real 1.5× inflation if ever retuned to ~2h — add the guard when windows become tunable).
  **S4:** `MergeCollapsed` merges on member-count floor only and can overfill past target —
  needs an intent ruling (pairs with spec-lens A-tier).

### Held under attack (recorded)

Harness byte-identical across full runs; drift gate fails at 26%/passes at 9% exactly; **no
commons-buff circularity — `EnclosureIndex` skips `SlotCommons`**; `CompliancePPM` matched
exact big.Rat across ~6000 points; the Modifier is continuous at the H=0.35 knee; the 2ebea38
in-batch reorder fix genuinely closes its case; capacity retry doesn't double-count; discount
fractions can't zero a requirement; the client-shell 5000ms boundary drains in exactly 100
steps with no prediction wedge.

**Remediation order for Codex: D1, D2 (new HIGHs — projection convergence is the contract),
then A1+D4 as one guard fix, A2+D3 as one blend fix, A3, then the second tier.**

## 2026-07-29 (claude — independent review of e0dcbcb..796ac55: APPROVED; two design rulings)

Full diff of the nine-commit round. Every finding remediated at root:

- **A1+D4 (the guard):** all four holes closed — every reachable baseline revision validated,
  shallow history a HARD failure (`fetch-depth: 0` in CI; ambiguity also fails), the artifact
  commit restricted to baseline+golden paths only (smuggling closed), changed-inputs-before-
  artifact enforcing the separate-commit protocol, dirty-artifact refusal. Fail-closed
  throughout; provider-metadata-free so local and CI enforce identically.
- **A2+D3 (the blend):** both live paths now call `commons.EffectiveHealthPPM(catalog, …)`;
  the hardcoded 0.8/0.2 is gone from both sites; the published formula is the executed one.
- **A3 (+ Codex's own catch):** `ConstantsHashArtifacts` extends the identity over economy,
  Commons, **and Routes** catalogs — the Routes binding was found missing by Codex's own
  adversarial pass and fixed (b7c838f), with the two `BALANCE-CHANGE:` rebaselines flowing
  through the newly hardened guard as their own first customers.
- **Self-found fail-open:** the first route-allowlist recipe had `|| true` spanning the whole
  import pipeline — a failed `go list` would have passed. Now enumeration is captured first and
  fails closed. The log correctly states self-review does not satisfy the independent gate.
- **A8:** persisted stream revisions in projection rows (migration 00009) + `GREATEST` on
  assignment timestamps — stale delivery cannot move the read model backward; reproducer tested.
- Second tier landed: runtime milestone validation, full run-key diagnostics, negative-factor /
  wrong-gate / ratchet / real-report-float / catalog-mutation tests. Suites green.

**The two deferred items are genuine design decisions — ruled here per the role division:**

1. **A7 (doctrine/gate alignment):** a `doctrine_is(transition=X)` condition evaluated at a gate
   crossed *before* X is chosen is unsatisfiable-by-construction — the three seeds as placed are
   dead routes. **Ruling: catalog validation rule — a doctrine-keyed condition may only appear on
   gates whose crossing occurs at or after transition X completes; move the three seeds to their
   doctrine's own-or-later gates.** (Belongs with the Depletion validator, which should also
   reject temporally-unsatisfiable predicates outright — a dead route silently weakens the
   exclusivity budget the proof depends on.)
2. **S4 (cohort merge policy):** **Ruling: merge triggers on the member floor only — rename
   `MergeCollapsed` to say so (`MergeBelowFloor`); Health is recoverable, membership is not.
   Overfill on merge is permitted to at most ⌈1.5 × CohortTargetSize⌉, preferring the emptiest
   target; never split.** (Lands in the active `commons-onboarding-and-governance` RFC.)

Both RFCs in this round clear to archive. Remaining from the archived-four review: nothing —
the board is clean except the two rulings' implementation.
