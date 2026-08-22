# Canonical Docs

This directory describes **what actually exists** — the living truth of the implemented system.

Implemented systems:

- [Numeric core](numeric-core.md) — cross-runtime large-number representation, canonical
  state/wire rules, economy helpers, and verification commands.
- [Economy kernel](economy-kernel.md) — strict shared resource/generator catalog, cross-runtime
  cost curves, scoped authoritative ledger transactions, and receipt boundary.
- [Continuous integration](ci.md) — public hosted workflow, blocking verification jobs, dependency
  cache boundaries, and balance-schema gate.
- [Deployment](deployment.md) — fail-closed production configuration and the manifest-derived,
  hash-bound runtime content staging boundary (Deployment Foundation remains in progress).
- [Save layer](save-layer.md) — owner-aware Postgres revision streams, canonical state format,
  migrations, optimistic concurrency, retention, and restoration.
- [Production engine](production-engine.md) — lazy authoritative accrual, multiplier slots,
  online/offline policy, exact manual-action clamp, idempotent intents, events, and progress.
- [Gate predicates and Route Registry](routes.md) — closed cross-runtime predicates, alternate
  gate costs, Depletion proof, first-executor naming, Route Knowledge, and idempotent projections.
- [Doctrine choice and Compute Credit](doctrine-and-compute-credit.md) — pinned transition choices,
  gate ordering, manual acceleration bursts, Company-v17 activation, and replay parity.
- [Fiscal Quarters](fiscal-quarters.md) — Founder-v19 wall-clock periods, deterministic harvest and
  spend transitions, and immutable per-run production contributions.
- [Active-play opportunities](active-play.md) — attended-time scheduling, deterministic claims,
  multiplicative buff windows, combo hardcaps, and schema-v2 cap receipts/events.
- [Soul foundation](soul.md) — Founder-v20 policy, debit and consumer gates, zero-output recovery
  persistence/replay, attended progress capabilities, and lazy watchdog cancellation.
- [Commons Compact](commons.md) — membership, source-derived Enclosure, Health/Capacity,
  persistent cohorts, AI fallback, production multiplier, dispatches, and population invariance.
- [Client shell](client-shell.md) — Svelte DOM routes, authoritative stream boundary, Worker
  prediction, reconciliation, visible caps, return flow, lifecycle, and client telemetry.
- [UI foundation](ui-foundation.md) — strict era tokens, exact-number presentation, Copy-backed
  primitives, surface lifecycle ownership, and the cross-browser accessibility fixture.
- [Game UI](game-ui.md) — the five Phase-A Svelte surfaces, bootstrap/runtime boundary, local
  display timing, accessibility and performance gates; two explicitly fail-closed arms remain
  under implementation-time rulings.
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
- [Gameserver composition](gameserver.md) — executable service graph, worker ownership,
  readiness, real settlement wiring, world snapshots, and bounded drain.
- [Minigame platform](minigame-platform.md) — frozen session genesis, Postgres claim ownership,
  and the pure tenant engine boundary (foundation implementation in progress).
- [The Pitch](minigame-the-pitch.md) — deterministic Decimal card scoring, pinned tenant content,
  Fiscal/Soul admission, certified payout, and fixture-first content verification.
- [Soul Recovery](soul-recovery.md) — zero-output recovery activities, authenticated coordinator
  lifecycle, heartbeat ceiling/reconnect behavior, and fixture-first activation boundary.
- [Balance harness](balance-harness.md) — scenario-driven pacing observation, baseline-change
  guard, and the relevance report.
- [Copy pipeline](copy-pipeline.md) — the shipped-string authority: catalog, references,
  generation, and the completeness gates.
- [Purchasable content](purchasable-content.md) — catalog-driven upgrade/purchase surfaces on the
  economy kernel.
- [T0–T1 playable content](t0-t1-playable-content.md) — the live first-hour catalogs, measured
  player policies, scripted first-company ending, branch-specific run-2 starters, and proof gates.
- [Guilds](guilds.md) — guild model, tithe, and reserved-credit clearing.
- [Combat](combat.md) — the shared combat data model and arithmetic kernel.
- [Meters](meters.md) — the live Company meter catalog: bands, decay, inputs, and Trust reseed.
- [Achievements](achievements.md) — live condition/proof grammar, score grants, settlement, and copy.
- [Pet care](pet-care.md) — live Founder-scoped care policy, decay/Trust arithmetic, and the FSM;
  pet acquisition and combat consumption remain successor work.
- [API foundation](api-foundation.md) — schema/cursor authority and the operation registry
  (foundation implementation in progress).
- [Founder transitions](founder-transitions.md) — Exit lifecycle and the attendance clock
  (canonical home for Founder Attendance per the 2026-08-07 ruling).

Rules (from RFC-0000):
- Organized by system, not history: `architecture.md`, `economy.md`, `data-formats.md`, `ops.md`, … created as systems land.
- Updated **in the same change** as any behavior change.
- When docs and code disagree, docs are the bug.

For intent see `design/`; for spec see `rfc/`; for in-flight work and job logs see `planning/`.
