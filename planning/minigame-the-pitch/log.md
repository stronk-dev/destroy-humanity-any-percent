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
