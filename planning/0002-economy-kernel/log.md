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

## 2026-07-28 — Kernel implementation checkpoint

- Added the version-1 JSON Schema at `balance/economy.schema.json` and shared catalog/curve
  fixtures at `testdata/economy-kernel.json`. The fixture is deliberately test data, not shipped
  balance.
- Implemented strict Go and TypeScript catalog loaders with exact field sets, canonical Decimal
  strings, stable IDs, duplicate/reference checks, explicit caps, and closed curve tags.
- Implemented constant, linear, and geometric bulk quotes plus verified max-affordable search in
  both runtimes. Shared vectors cover ratios `1.10` and `1.13` and an exponent above 100,000.
- Implemented the authoritative Go ledger with read queries, aggregate-before-quantize atomic
  transactions, invariant validation, and deterministic receipts.
- The normalization follow-up is implemented and archived in commit `708d6a1`.
- `make verify` passes with 6,321 Node tests and 18,963 browser tests across Chromium, Firefox,
  and WebKit. Go economy tests include the million-source regression and negative control.
- Pre-archive review tightened ledger ownership: each ledger now has exactly one catalog scope.
  It cannot expose or mutate Company/Founder/World/Guild balances across that boundary; later
  cross-scope actions must coordinate explicit ledgers above the kernel.
