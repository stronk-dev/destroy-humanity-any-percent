# Cloud Clicker 1.0 — long-term delivery board

**Goal:** ship a complete, supportable Cloud Clicker 1.0, **Transcendence**, in which a player can
reach and understand the designed endings through the full nine-tier game, while the service and
supported self-host package meet the same recovery, rights, accessibility, privacy, operations and
preservation obligations as the gameplay. This is the long-term goal, not a claim that 1.0 is near.

**Current checkpoint:** 2026-09-23, candidate source `7e8aa70`; see
[`roadmap-1.0-log.md`](roadmap-1.0-log.md). **Current public claim:** development snapshot.
**Next accepted release target:** the bounded Phase-0 Playable Preview, per D-001/D-007. The
long-term 1.0 goal does not promote the preview's scope or authorize public hosting.

## Authority and read order

This board is a rollup, not a second design or implementation specification:

1. [`design/07-roadmap.md`](../design/07-roadmap.md) defines the phase content and 1.0 ending.
2. [`planning/platform-alignment/owner-ruling-packet.md`](platform-alignment/owner-ruling-packet.md)
   and [`decision-queue.md`](platform-alignment/decision-queue.md) hold adopted and open choices.
3. [`rfc/README.md`](../rfc/README.md) and accepted RFC bodies authorize implementation.
4. [`planning/platform-alignment/execution-queue.md`](platform-alignment/execution-queue.md) names
   immediately executable work. A row here never changes an RFC's status or makes work `READY`.
5. Per-RFC plans/logs, [`docs/`](../docs/), and executed evidence establish current behavior.
   The [`capability map`](platform-alignment/capability-map.md) is an audited baseline, dated at its
   own coordinate; later changes must be rechecked at the current HEAD.

The sibling `chess-drills` roadmap supplied the useful pattern: one strategic rollup, explicit
capability dimensions, milestone dependencies, and evidence-linked checkpoints. Cloud Clicker
already has detailed ledgers and RFC registers, so this board links them instead of duplicating
their rows or inventing a percentage complete.

## What counts as done

A milestone is complete only when **all** of these are supported by current evidence:

| Dimension | Required proof |
|---|---|
| Intent and decisions | The included content and player promise are exact; owner choices and authored copy are adopted. |
| Rules and state | Server authority, versioned persistence, replay, migration and offline/return behavior cover the milestone. |
| Production API | The production route and generated client carry the contract, including errors and authorization. |
| Player experience | The default browser journey reaches the outcome and handles loading, absence, failure, reconnect and return. |
| Content and balance | Real, reviewed, licensed content exercises the journey; the harness validates pacing and failure cases. |
| Access and rights | Keyboard/assistive tasks, reflow, recovery, export, deletion and honest disclosure are tested as player workflows. |
| Operations | A clean supported host installs, backs up, restores, upgrades, rolls back and alerts within ruled objectives. |
| Release evidence | Current-head CI, negative/severing witnesses, exact artifact rehearsal, canonical docs and designated review all pass. |

`Partial` means at least one link exists but the whole outcome is unproven. `Blocked` names a
decision, dependency or unavailable witness. `Complete` requires the evidence above; an archived
foundation RFC, a green unit suite or a historical audit cannot make its parent milestone complete.
There is no percent-complete estimate. Counts of tests, files or RFCs are not player outcomes.

## Milestone path

These are cumulative release gates, not permission to weaken earlier obligations. Every public
milestone must satisfy the access, rights, safety and operations floor; those concerns cannot be
postponed by changing the release label.

| Gate | State at checkpoint | Product exit from `design/07` | Main dependency / proof still needed |
|---|---|---|---|
| Phase-0 Playable Preview | **In construction; unreleased** | Honest T0–T1 click → generator → first Exit → next Company browser journey, with the bounded preview manifest. | Finish and independently review Deployment Foundation; run R-006 on the exact clean-host bundle. Account rights, recovery, task accessibility and retention still need their own decisions/contracts and end-to-end proof. |
| v0.1 The Garage | **Not yet scoped in accepted release RFC** | T0–T2, first Exit/Reputation, Fiscal/Clout/shop, three named minigames, pet care, ambient presence/feed/counters, era UI and launch content. | Preview floor plus accepted producer→API→surface→content contracts and a tested default player journey for every named feature. |
| v0.2 Incorporation | **Design intent; blocked by upstream systems** | Factions, guild/exchange, Tier 3 and the first measured community milestone; Events Layers 1–2. | Rule and prove privacy-preserving milestone measurement before threshold selection; complete social/world/event dependencies and real content. |
| v0.3 Hyperscale | **Design intent; blocked by upstream systems** | Tier 4, The Lane, clocks/shop, ranked board-game queue with fair bot backfill, Event Layer 3, GM/war log and speedrun surfaces. | Accepted combat/match, feed/world, leaderboard readers, operator and content contracts with real multiplayer/bot fallback proof. |
| v0.4 Frontier | **Design intent; blocked by upstream systems** | Tier 5, research/canonization/casino, pet battles, Commons/Ethical%, challenge runs, Compute Credits and Soul-gated content. | Earlier world/social/combat paths and measured balance; accessible player surfaces and reviewed content for each mode. |
| v1.0 Transcendence | **Long-term goal; not release-ready** | Tiers 6–8, all three designed endings and variants, complete category/challenge set, run-end retrospectives and the Honesty appendix. | All prior gates plus an owner-adopted exact 1.0 content/release manifest, full-path ending proofs and preservation posture appropriate to the final release. |

The Phase-0 label and scope are adopted; the exact later release manifests have not been adopted.
Rows for v0.1–1.0 are design intent and must be converted into decisions, research and accepted
RFCs before implementation. No date or team-capacity promise is inferred.

## Cross-cutting workstreams at this checkpoint

| Workstream | Current evidence boundary | Next authority or proof |
|---|---|---|
| Gameplay foundation and T0–T1 | Strong server foundations and a bounded, designated-approved Game UI browser flow; tiers 2–8 and terminal endings absent. | Complete Phase-0 manifest and then phase-specific content/RFC DAG; preserve fixed historical and player workflow witnesses. |
| API, transport and minigames | Transport recovery and exact API witnesses are approved; generated client, public readers and player minigame surfaces remain incomplete. | Reconcile ruling bodies, accept the missing surface contracts, then prove full production routes and consumers. |
| Account, rights and privacy | Backend security/deletion tests exist; player recovery/export/delete and retention disclosure do not. All 60 game tables have SQL-column maps, the four browser-local key families are traced, and the core save/replay/transport payload lineage is mapped; other nested payloads and operator fields remain unclassified. Source shows verified-run compaction preserving complete commands/receipts/genesis/matched events in immutable archive bytes, but the current archive fixture has empty payloads and no events (RP-124); socket publication retains its payload row. The browser silently persists the one-time recovery code without a player display/copy/download path (RP-123). Guild deletion leaves a UUID in event JSON despite a nulled FK (RP-120), and the reused-DB Guild suite is not isolated (RP-121). Populated-board deletion remains untested; Commons post-delete health (RP-119) and Minigame/Soul/Transport residuals (RP-122) are still source-derived. | Owner/legal D-008/D-009/D-015 decisions, remaining nested/operator-field classification and joined rights/Commons witnesses, accepted Account/UI/retention contracts, R-003 on the actual workflow. |
| Accessibility and usability | Current-source probes fail lifecycle focus and 320 px reflow in Chromium, Firefox and WebKit; the unmodified Linux browser lane passes, proving its fixture oracle misses both. Numeric shell motion still ignores OS preference. A cross-surface RFC is drafted but not accepted. | Owner task/AT matrix and accepted contract, RP-082–RP-084 repairs under that authority, then R-005/R-008 on built player tasks. |
| Deployment and operations | Deployment Foundation is accepted and implementing. Two independent 72-artifact candidate bundles at source `7e8aa70` match byte-for-byte; both pass the exact previous-bundle probe, and the retained build record, source/image secret scan and full local supply-chain check pass. Historical save-version rollback validation now has a discriminating fixture. Hosted code CI at that source passed all six jobs. The browser and bounded verified-board recovery gates are constructed and locally tested, including a real-Postgres event-only negative. No real clean-host player run or projection arrival is proved. CI parity correction at `e0e2201..cf4ac25` has only Codex first-filter review. | Designated cross-party review of the implementation ranges; execute/sever the browser path and populated recovery on the exact clean-host bundle, then finish R-006 and R-007 with the operator path. |
| Multiplayer, world and later content | Server Commons/guild/faction primitives exist; real social/world/feed and later-tier player journeys are incomplete. | Decide later public social/telemetry scope, then dependency-ordered RFCs, content and integrated proofs. |
| Release governance | Dated audits, defect ledger, research/decision queues, RFC graph and per-RFC logs exist. The old executable-queue handoff was stale at this HEAD. | Reconcile the live queue and this checkpoint whenever a reviewed batch, ruling or release witness changes the critical path. |

## Current next moves

1. Obtain the designated cross-party verdict for `e0e2201..cf4ac25`; its local
   `make verify-push` pass is a first filter, not the required independent review.
2. Continue the accepted Deployment Foundation DP-F plan: take the byte-matched local candidate
   through designated review and an authorized clean Linux host, then execute and sever the browser
   journey, board-arrival gate, recovery and rollback on those exact candidate/previous bundles.
   Record R-006 and its negative cases; local supply-chain success is not product proof.
3. Bring the source-traced [`Phase-0 exact-manifest proposal`](platform-alignment/phase0-manifest-proposal.md)
   to Marco for D-007 adoption or edits; it names P01–P07 player tasks and F01–F05 public-release
   floor obligations without calling them complete. Then bind D-018 and R-008 to the adopted tasks.
   Finish nested-payload and remaining operator-data classification, then
   prepare the separate account export/deletion/retention decisions. RP-119 also needs active
   Commons semantics and a joined witness; none of this is optional preview polish.
4. Once the preview floor is proved, reconcile the v0.1–1.0 design into bounded milestone manifests
   and dependency-ordered RFCs. Do not implement later content directly from this board.

## Update protocol

At every material checkpoint, append to [`roadmap-1.0-log.md`](roadmap-1.0-log.md): date, exact
HEAD/release artifact, observed change, executed evidence and limits, decision/review provenance,
and the next blocked/ready boundary. In the same reviewed range, update this board's state and
current next moves, the relevant RFC plan/log, and the executable queue if authorization changed.
If a source is stale, mark its audit coordinate; do not copy its old count as current truth.
The goal closes only after the complete 1.0 manifest and all eight dimensions pass on the release
artifact and Marco makes the release call. A Phase-0 or v0.x release advances the goal but cannot
close it.
