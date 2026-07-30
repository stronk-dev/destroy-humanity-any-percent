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
- [Gate predicates and Route Registry](routes.md) — closed cross-runtime predicates, alternate
  gate costs, Depletion proof, first-executor naming, Route Knowledge, and idempotent projections.
- [Commons Compact](commons.md) — membership, source-derived Enclosure, Health/Capacity,
  persistent cohorts, AI fallback, production multiplier, dispatches, and population invariance.
- [Client shell](client-shell.md) — Svelte DOM routes, authoritative stream boundary, Worker
  prediction, reconciliation, visible caps, return flow, lifecycle, and client telemetry.
- [Accounts and sessions](accounts-and-sessions.md) — anonymous recovery credentials, exact JWTs,
  refresh-family rotation, Founder ownership/import, deletion, rate limits, and the HTTP intent
  boundary.
- [Prestige and Exits](prestige-and-exits.md) — exact prestige arithmetic, deterministic offers,
  atomic Founder/Company Exit commits, server-derived timers, and new-run assembly.
- [Leaderboards and Balance Epochs](leaderboards-and-epochs.md) — atomic run logs, immutable catalog
  artifacts, run epoch pinning, exact competition ranks, world-first arbitration, and replay status.
- [WebSocket transport](transport.md) — strict outbound envelopes, literal limits, recovery history,
  drop-stale versus lossless queues, and server-side channel authorization.
- [Factions and incorporation](factions.md) — four run-scoped faction identities, Tier-2
  incorporation, Open Source Compact binding, and attended-time interdependence stock.

Rules (from RFC-0000):
- Organized by system, not history: `architecture.md`, `economy.md`, `data-formats.md`, `ops.md`, … created as systems land.
- Updated **in the same change** as any behavior change.
- When docs and code disagree, docs are the bug.

For intent see `design/`; for spec see `rfc/`; for in-flight work and job logs see `planning/`.
