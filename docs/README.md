# Canonical Docs

This directory describes **what actually exists** — the living truth of the implemented system.

Implemented systems:

- [Numeric core](numeric-core.md) — cross-runtime large-number representation, canonical
  state/wire rules, economy helpers, and verification commands.
- [Economy kernel](economy-kernel.md) — strict shared resource/generator catalog, cross-runtime
  cost curves, scoped authoritative ledger transactions, and receipt boundary.
- [Continuous integration](ci.md) — public hosted workflow, blocking verification jobs, dependency
  cache boundaries, and balance-schema gate.
- [Save layer](save-layer.md) — owner-aware Postgres revision streams, canonical state format,
  migrations, optimistic concurrency, retention, and restoration.
- [Production engine](production-engine.md) — lazy authoritative accrual, multiplier slots,
  online/offline policy, exact manual-action clamp, idempotent intents, events, and progress.

Rules (from RFC-0000):
- Organized by system, not history: `architecture.md`, `economy.md`, `data-formats.md`, `ops.md`, … created as systems land.
- Updated **in the same change** as any behavior change.
- When docs and code disagree, docs are the bug.

For intent see `design/`; for spec see `rfc/`; for in-flight work and job logs see `planning/`.
