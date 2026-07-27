# RFC-0002 Economy Kernel — Running Log

Append-only. A fresh agent should be able to resume from this file and `plan.md` alone.

## 2026-07-28 — Research and acceptance

- Audited the original RFC against the implemented numeric core and actual repository state.
- Found that the draft bundled five absent systems and two unresolved tuning decisions.
- Researched Antimatter Dimensions, The Modding Tree, Swarm Simulator, JSON Schema 2020-12,
  and the Unlimited Rulebook economy reference architecture. Findings are recorded in
  `design/research/economy-kernel.md`.
- Owner direction: make the engine abstract enough that it does not care about `1.10` vs `1.13`,
  then research, adjust the RFC, and implement.
- Re-scoped RFC-0002 to a strict resource/generator catalog, three closed cost-curve types, an
  atomic server ledger, deterministic receipts, and cross-runtime parity tests.
- Resolved policy: no global gameplay ceiling; no launch ratio in code; per-resource caps and
  per-generator-class curve parameters are declarative data.
- RFC accepted in commit `9076057`; implementation started immediately afterward.
- No `DESIGN-GAP` currently blocks the bounded kernel.

## 2026-07-28 — Numeric dependency defect

- Implemented the first catalog, curve, ledger, schema, fixture, and test pass locally.
- The million-source aggregation gate exposed a numeric-core invariant bug after ten equal
  `1e87` additions: the Go normalizer returned mantissa `10`, exponent `87`.
- Split the conformance repair into `rfc/numeric-normalization-carry.md`; RFC-0002 remains bounded
  and resumes after that dependency is green.
