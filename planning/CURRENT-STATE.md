# Current repository state

Last reconciled: 2026-08-21 through `bb615ed`; the local CI split described below is the pending
transactional batch.

This is a navigation brief, not a second execution queue. Current authorization lives in
[`platform-alignment/execution-queue.md`](platform-alignment/execution-queue.md), active RFC state
in [`../rfc/README.md`](../rfc/README.md), and implemented behavior in [`../docs/`](../docs/).

## Product status

Cloud Clicker is a development snapshot, not a 1.0 release. The owner-selected next milestone is a
bounded **Phase-0 Playable Preview** covering the audited T0–T1 vertical slice; its exact release
manifest is not yet accepted.

Substantial implemented foundations include:

- cross-runtime numeric, economy, save, production and offline-progress kernels;
- account/session, Postgres persistence, WebSocket transport and recovery;
- route, replay, epoch, Commons, guild, achievements, meters and pet-care foundations;
- one real minigame tenant (The Pitch), T0–T1 content and run-2 starter consequences; and
- a Svelte browser path that reaches the Desk through the real gameserver, Postgres and WebSocket.

This does not prove a complete first-hour browser workflow. Tiers 2–8, most minigames, combat
engines, player-facing social/world/event systems, production packaging and the designed endings
remain incomplete.

The evidence trace is in
[`platform-alignment/capability-map.md`](platform-alignment/capability-map.md) and
[`platform-alignment/capability-reality-audit.md`](platform-alignment/capability-reality-audit.md).

## Recently closed evidence work

- Q-001 Account witnesses are designated-approved at `34d04a5`.
- Q-002 Minigame API witnesses and the separate API registry tightening are designated-approved at
  `bfd9b65`.
- Q-003 Transport recovery is designated-approved at `249719c`.
- Harness Observability is implemented, designated-approved at `96a574d`, and archived at
  `edafd82`. Its complete local observation took 974.510 seconds, of which 739.441 seconds were
  standard pacing and 212.927 seconds active relevance.
- D-002 repository publication/shared-memory disposition is complete at `7ad5b54`. The
  `e44e1a6` verifier rejects forged publication states and passes from a fresh clone with 11 public
  and 56 private Class-C dossiers, three ignored duplicates and seven ignored diagnostics.
- Copy Amendment A1's owner record now reaches its exact cross-party closure (`3773510`), canonical
  Soul docs match the current epoch (`7d21484`), the root boundary gates honor the repository-local
  Go cache (`ab05f2d`), and four completed maintenance threads moved out of the live namespace
  (`b990b91`).

These closures do not make their broader parent RFCs archival-eligible; the exact remaining body,
consumer, lifecycle and range-union blockers are in the execution queue.

## Release blockers

The repository still lacks a coherent release floor:

- no supported production Docker/Caddy bundle, clean-host proof, rollback flow or backup/restore
  rehearsal;
- no player-facing account recovery, export or deletion workflow;
- incomplete accessibility evidence for keyboard, screen-reader, zoom/reflow and touch workflows;
- no accepted production observability, operator, incident or complete retention contract;
- no hosted verdict yet for the newly implemented fast/maintenance CI split; and
- several active RFC bodies and canonical docs still contradict accepted rulings or current HEAD.

The complete defect and decision populations are
[`platform-alignment/backlog.md`](platform-alignment/backlog.md) and
[`platform-alignment/decision-queue.md`](platform-alignment/decision-queue.md).

## Current execution posture

There is no unblocked product-implementation batch in the platform-alignment queue.

The next CI evidence action is a current-head push/PR run followed by one manual or scheduled
maintenance observation. The workflow split is implemented locally but cannot receive hosted
credit until an explicitly authorized push transfers it to GitHub. Remaining product lanes require
their named owner/ruling-author reconciliation or decisions before implementation.

Do not infer release readiness from green unit tests, backend primitives or this summary. A shipped
capability still requires its producer, consumer, current data/workflow and discriminating
integrated proof.
