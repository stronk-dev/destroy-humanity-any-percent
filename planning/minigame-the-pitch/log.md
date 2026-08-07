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
