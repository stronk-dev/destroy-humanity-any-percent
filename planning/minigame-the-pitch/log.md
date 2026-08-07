# The Pitch implementation log

## 2026-08-07 — Codex acceptance review: blocked (TP-C1–TP-C10)

Review by: Codex. Recorded by: Codex.

The concept is a strong exemplar, but the shipped platform cannot carry its Decimal certified
result, cannot resolve a hot-reloadable tenant content artifact, does not enforce the declared
Fiscal unlock, and exposes no playable client/API surface. The draft also lacks the exact engine,
effect, tenant-row, and launch-content byte contracts. Ten blockers with proposed resolutions are
filed in the RFC; no code or balance data was introduced.

## 2026-08-07 — Codex implementation-readiness review: blocked (TP-C11–TP-C18)

Review by: Codex. Recorded by: Codex.

The owner rulings settle the product decisions, but the normative body still contains the original
placeholders and no exact engine snapshot, command, effect, content, tenant, artifact-identity,
unlock, exponent, or mint bytes. Eight narrower blockers with proposed contracts are filed. This is
not a re-litigation of TP-C1–TP-C10; it is the executable layer those rulings require.

## 2026-08-07 — Codex source-grounded implementation review: blocked (TP-C19–TP-C25)

Review by: Codex. Recorded by: Codex.

TP-C11–TP-C18 settle the launch rows and visible product shape, but a walk against the actual tenant
service proves the engine still cannot receive its pinned content or immutable seed during Apply.
The exact snapshot contradicts its required genesis identity; duplicate-free base card IDs make the
pair hack unreachable; shop identity, currency income, two effect predicates, terminal facts, six
required tenant-policy literals, and the content-gate corpus remain undefined. Seven narrow
blockers with executable proposals are filed. Soul Recovery implementation continued; no Pitch
mechanic or balance byte was improvised.

## 2026-08-07 — Codex implementation: ready for designated review

Review by: pending designated cross-party reviewer. Recorded by: Codex.

Implemented TP-C1–TP-C25 fixture-first without minting a production artifact: exact Go/TypeScript
catalog loaders and tenant engines; pinned tenant-content resolution; Fiscal unlock composition;
the complete 12-card/8-hack launch fixture; shared content-gate corpus; generated copy bindings; and
a real Postgres create→play→resolve→payout→retry→Founder-replay path. The composed test exposed a
pre-existing coordinator seam: live minigame resolution manually mutated Founder rating/quality
while replay used `ApplyFounderLogged`, so a Fiscal auto-sweep could make an honest log diverge.
Live resolution now invokes the same Founder boundary and the integration test proves the combined
Fiscal + Pitch history verifies. Kernel semantics advance to 0.3.77 and register both Pitch paths.

This is a handoff, not a verdict. No archive or production mint is authorized by this entry.

Implementation range: `d1cd39cffbbb3d2b3e1af174420bebb137d2776d..0eb3772`.

Self-check evidence read to terminal completion before handoff:

- `make verify` — green, including all Go packages, Pitch content-corpus drift, formula/harness
  checks, TypeScript + Svelte typecheck, production client build, 6,594 client tests, kernel-history
  guard 0.3.77, copy/schema/boundary gates, and 19,791 browser assertions.
- `make test-save-integration` — green against real Postgres, including the composed Pitch path and
  Founder replay with the Fiscal sweep active.
- `git diff --check` — green before commit.

Codex self-review additionally closed Go/TypeScript parity gaps for undeclared copy keys, negative
snapshot Decimals, unknown slotted hacks, malformed command rejection, and canonical nested offer
bytes. This self-review does not satisfy the mandatory designated cross-party review.

## 2026-08-07 — designated cross-party verdict: The Pitch (0eb3772 + 853ef93) — NOT APPROVED (narrow)

- **Review by:** the designated Claude reviewer (independent; make verify + Postgres suite re-run,
  discriminating probes executed and reverted). **Recorded by:** same.

**Substantively excellent — every TP-C19–C25 byte contract verified at source:** the
TenantContentResolver identity chain (cloning proven against aliasing), the 13-key snapshot +
instance IDs (dark_pattern's pair reachability test-proven), per-round Fisher–Yates with rejection
sampling and the fixed deal slices, the income/shop contracts, the one-quantize byte equation
(order-pinned by test), terminal semantics + integer-facts-only Result, the complete loadable v3
definition in BOTH runtimes with the fiscal_unlock arm and reject-before/succeed-after proof, the
12-scenario content-gate corpus (transition budget 108 = exact command sum, drift-checked),
containment (no production artifacts), kernel 0.3.75→77 lockstep.

**The two highest-risk items both verified clean:**
- The approved d4c2312 rewrite was executed EXACTLY as approved (only the sentinel-refactor hunks +
  their one dependent test assertion dropped; lifecycle tests/corpus/evidence intact; no bump owed).
- The ApplyFounderLogged composer change is CORRECT and empirically pinned: live resolution
  previously hand-mutated Founder state while replay ran the full boundary (Fiscal sweep +
  receipt decoration) → honest logs failed VerifyFounderHistory. The fix routes live through the
  same boundary; atomicity/lock-order/idempotency untouched; reverting the hunk makes the
  integration test fail with state_divergence — a real discriminating regression.

**BLOCKING (route to Codex):**
- **F1 (MEDIUM): AC1's soul-gate arm has ZERO test coverage** — deleting the human_hobby
  enforcement block leaves the whole suite green. Close with one composed test: Soul ≤ near-zero →
  Pitch start rejects `not_eligible/human_content_locked`.
- **F4 (LOW-MED): AC2's "big-number-regime hand" vector is absent** (corpus max ≈2.5e3; the TP-C17
  vectors exercise exponent projection only, never score()). Close with one synthetic big-regime
  scoring vector (catalog-independent, both runtimes).

**Owner rulings on F2/F3 (recorded now, reconcile in the remediation edit):**
- **F2 RULED: the IMPLEMENTED interleaved single raw-byte pass is canonical** — shape_factor and
  chain_factor apply in one pass over hacks in raw-byte hack_id order (simpler, corpus-pinned, and
  identical in both runtimes). TP-C22's staged wording (shapes-then-chains) is SUPERSEDED; the RFC
  body gets the one-line reconciliation in the remediation edit.
- **F3 RULED: the shipped substream binding is canonical** —
  `Substream(Substream(seed, "pitch.run.v1").Next() XOR round, label)` is the recorded meaning of
  the ruling's `Substream(seed, label, round)` notation; record the binding in the same edit.

F5/F6 observations: no action required (cosmetic parity nuance; unreachable guard asymmetry).

**Range-union:** this verdict consumes exactly {0eb3772, 853ef93}. Soul Recovery's code
(4973c8e, ab9d15e, f04c2f3, d1cd39c) awaits its own designated review; 506f12d/f88f178 belong to
the completed Soul Foundation gate; all other span commits are docs-tier. No code commit between
3ff2082 and 853ef93 lacks a named review thread.

**Verdict: NOT APPROVED pending F1 + F4 (one test + one vector); re-review is narrow. Archival will
cite {0eb3772, 853ef93} + the remediation range.**

## 2026-08-07 — Codex narrow remediation: ready for designated re-review

Review by: pending designated cross-party reviewer. Recorded by: Codex.

- **F1 closed:** the composed real-Postgres test now creates a separately unlocked Founder at Soul
  0 and proves Pitch start rejects `ErrInvalidIntent` with `human_content_locked` before the tenant
  or session row is reached.
- **F4 closed:** `testdata/pitch/big-number-v1.json` is one shared Go/TypeScript vector. Both
  runtimes execute the actual scoring path and produce the exact Decimal `2e400` from two `1e300`
  cards under a `1e100` card factor—well beyond binary-float range.
- **F2/F3 reconciled:** the normative RFC now records the interleaved raw-byte predicate-factor pass
  and the exact nested `pitch.run.v1`/per-round substream construction ruled by the designated
  reviewer.

Evidence read to completion: `make test-save-integration`, focused Go tests, TypeScript/Svelte
typecheck, 6,595 client tests, and full `make verify` including 19,794 browser assertions. The diff
contains tests, one shared test vector, RFC text, and this planning evidence only; kernel remains
0.3.77. This entry is a handoff, not a verdict or archival authorization.

## 2026-08-07 — designated cross-party re-review: remediation 02ccc4c..2a55e12 — APPROVED

- **Review by:** the designated Claude reviewer (independent; both gate suites re-run at 2a55e12;
  destructive discrimination probes run in a throwaway worktree and removed). **Recorded by:** same.

Narrow scope verified — the range contains exactly one commit (`2a55e12`, tests + planning + RFC
reconciliation only, zero production code, no scope creep):
- **F1 CLOSED, discrimination proven:** composed Postgres test
  (`server/production/pitch_integration_test.go:130-173`) isolates the soul gate (fiscal unlock
  pre-granted), Soul=0 → `StartMinigameSession` rejects `ErrInvalidIntent` +
  `human_content_locked`. Probe: deleting the `human_hobby` enforcement block makes the suite fail
  at exactly this assertion.
- **F4 CLOSED, discrimination proven in BOTH runtimes:** shared vector
  `testdata/pitch/big-number-v1.json` (2×(1e300×1e100)=2e400 — past float64 range, hack
  interaction included, byte-pinned literal, hand-verifiable). Go consumes it through the real
  `score()`; TS through the full `applyPitch` path against the same bytes. Tamper probe: mutating
  the expected value fails both suites independently.
- **F2/F3 reconciled:** TP-C22 body+ruling now state the interleaved single raw-byte pass; all
  five substream references record the shipped nested binding; zero stale notation remains; no new
  contradictions.
- Gates: `make verify` green (6,595 client tests, 19,794 browser assertions, kernel parity+history
  guard 0.3.77); Postgres suite green; no-bump claim verified mechanically (`_test.go`/unwatched
  residue only).

**Follow-up editorial (non-blocking, pre-existing):** TP-C4's accepted ruling (~rfc lines 291-294)
still carries the coarse staged "hand-shape factors → ordered hack interactions" wording predating
the F2 ruling; TP-C22's byte equation governs. Reconcile in a future RFC edit (e.g. the archival
move's edit).

**Verdict: APPROVED for the combined set {0eb3772, 853ef93, 2a55e12}.** Range-union: prior verdict
consumed {0eb3772, 853ef93}; this re-review consumed 02ccc4c..2a55e12 = {2a55e12}; the union covers
every implementation commit of The Pitch span (intervening 7688c59 = the verdict filing itself,
02ccc4c = a Claude-side RFC draft outside this span, in the Codex-reviews-Claude channel).

**The Pitch is ARCHIVAL-ELIGIBLE on this verdict.** The archival move (status→implemented, RFC +
planning to archives, canonical docs page) is Codex's to execute, citing this entry and the exact
consumed set — recommended to fold the TP-C4 editorial line into the same edit.
