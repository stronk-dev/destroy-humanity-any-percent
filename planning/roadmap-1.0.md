# Cloud Clicker 1.0 — long-term delivery board

**Goal:** ship a complete, supportable Cloud Clicker 1.0, **Transcendence**, in which a player can
reach and understand the designed endings through the full nine-tier game, while the service and
supported self-host package meet the same recovery, rights, accessibility, privacy, operations and
preservation obligations as the gameplay. This is the long-term goal, not a claim that 1.0 is near.

**Current checkpoint:** 2026-10-04, product source `efe68333` (RP-145–RP-147
corrections await cross-party review; the broader v0.1 and Deployment ranges remain unapproved,
and the kernel-version CI gate remains red). Other workstream rows below retain their dated
2026-09-30 audit population unless a later checkpoint explicitly updates them; see
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
| Phase-0 Playable Preview | **In construction; unreleased** | Honest T0–T1 click → generator → first Exit → next Company browser journey, with the bounded preview manifest. | Independently review Deployment Foundation's corrective range, rebuild both bundles, and run R-006 on the exact clean-host artifact. Account rights, recovery, task accessibility and retention still need their own decisions/contracts and end-to-end proof. |
| v0.1 The Garage | **Accepted feature RFCs implementing; release unproved** | T0–T2, first Exit/Reputation, Fiscal/Clout/shop, three named minigames, pet care, ambient presence/feed/counters, era UI and launch content. | Claude landed implementation ranges for the accepted v0.1 RFCs, but Codex has not approved their complete range union. RP-132 blocks Pet Adoption PA7; public Typer/Arcade wire, adopted copy/content mint and integrated player journeys remain open. The Phase-0 floor also applies. |
| v0.2 Incorporation | **Design intent; blocked by upstream systems** | Factions, guild/exchange, Tier 3 and the first measured community milestone; Events Layers 1–2. | Rule and prove privacy-preserving milestone measurement before threshold selection; complete social/world/event dependencies and real content. |
| v0.3 Hyperscale | **Design intent; blocked by upstream systems** | Tier 4, The Lane, clocks/shop, ranked board-game queue with fair bot backfill, Event Layer 3, GM/war log and speedrun surfaces. | Accepted combat/match, feed/world, leaderboard readers, operator and content contracts with real multiplayer/bot fallback proof. |
| v0.4 Frontier | **Design intent; blocked by upstream systems** | Tier 5, research/canonization/casino, pet battles, Commons/Ethical%, challenge runs, Compute Credits and Soul-gated content. | Earlier world/social/combat paths and measured balance; accessible player surfaces and reviewed content for each mode. |
| v1.0 Transcendence | **Long-term goal; not release-ready** | Tiers 6–8, all three designed endings and variants, complete category/challenge set, run-end retrospectives and the Honesty appendix. | All prior gates plus an owner-adopted exact 1.0 content/release manifest, full-path ending proofs and preservation posture appropriate to the final release. |

The Phase-0 label and scope are adopted; the exact later release manifests have not been adopted.
Several v0.1 feature RFCs have been accepted and implemented locally, but that does not adopt a
v0.1 release manifest or complete a milestone. Rows for v0.2–1.0 remain design intent requiring
decisions, research and accepted RFCs before implementation. No date or team-capacity promise is inferred.

## Cross-cutting workstreams at this checkpoint

Rows retaining the older `7e8aa70` audit population are leads, not assertions that every
underlying table or browser failure was re-executed at `a7ab2640`; this checkpoint independently
rechecked the CI/kernel, PA7 and RP-133–RP-135 Deployment boundaries recorded below.

| Workstream | Current evidence boundary | Next authority or proof |
|---|---|---|
| Gameplay foundation and T0–T1 | Strong server foundations and a bounded, designated-approved Game UI browser flow. Claude landed Tier 2 feature/content code, not yet designated-reviewed or proved as a default journey; higher tiers and terminal endings remain unproved. | Complete Phase-0 manifest and then phase-specific content/RFC DAG; preserve fixed historical and player workflow witnesses. |
| API, transport and minigames | Transport recovery and exact API witnesses were approved in earlier ranges. Claude has since landed new API, minigame and Game UI surfaces, but the new ranges are not designated-approved. Codex's 2026-10-04 targeted MA3 review found and locally corrected hidden-at-start and post-unmount Soul heartbeat leaks (RP-145/RP-146), with browser negatives firing first. A DOM-only Fiscal→Pitch composed path then exposed GS0.2's stale Founder revision after applied harvest (RP-147); the accepted-contract correction now holds pending through refresh, and the real Postgres browser journey plays Pitch to an applied terminal receipt. Full Chromium/WebKit suites pass; the three-browser gate is unverified because local Firefox aborts before test import. The PA7/C2 conflict remains an observed contract mismatch (RP-132). | Claude must cross-party review the bounded Codex Soul correction `9ceed1cf^..47562850` and GS0.2/MA composed correction `efe68333`; Codex must finish the wider MA and Garage reviews. Settle PA7/C2 plus Typer/Arcade public-wire conflicts, then prove production routes, generated consumers and current data in default player workflows. |
| Account, rights and privacy | Backend security/deletion tests exist; player recovery/export/delete and retention disclosure do not. All 60 game tables have SQL-column maps; browser, core-payload, selected core-event/receipt, verification/board and operator surfaces have bounded traces, while other nested payloads and clean-host behavior remain open. The event registry has 48 kinds/52 SQL kind-version pairs, and a cold fixture showed 28/28 Soul events copied into the player outbox; that fixture did not delete an account. Selected receipt fixtures exposed raw progress tokens on 14/14 terminal Soul rows (RP-127) and a Minigame API receipt trigger that blocks parent-session cleanup (RP-128). Verification's immutable poison path retains bounded but unredacted upstream error text (RP-129), not a proved secret leak. A separate cold joined Account API probe retained one seeded verified board row linked to an archived Founder and two archived streams after deletion (RP-130); full projector/dead-letter and public-reader deletion proof remains absent (RP-118). Source shows verified-run compaction preserving full content, but its fixture has empty payloads/no events (RP-124). The browser silently persists the one-time recovery code without a player display/copy/download path (RP-123). Guild deletion leaves a UUID in event JSON (RP-120); the reused-DB Guild suite is not isolated (RP-121). Another joined deletion probe showed the exact Commons World-count and health sample-selection predicates still counting an archived Founder; the full projection/recompute/player flow remains absent (RP-119). A seeded joined probe kept Minigame session/receipt/faucet, active Soul session with token, intent receipt and player-outbox rows 1/1 after a 204 deletion (RP-122, composed coordinators unproved). Backup restoration of a deleted account (RP-125) remains a source-derived risk. The ledgers' operator identity lacks retention policy (RP-126). | Owner/legal D-008/D-009/D-015 decisions (options framed in `platform-alignment/rights-decision-sheet.md`), remaining nested-field/clean-host classification and full rights/Commons witnesses, accepted Account/UI/retention contracts, R-003 on the actual workflow. |
| Accessibility and usability | Current-source probes fail lifecycle focus and 320 px reflow in Chromium, Firefox and WebKit; the unmodified Linux browser lane passes, proving its fixture oracle misses both. Numeric shell motion still ignores OS preference. A cross-surface RFC is drafted but not accepted. | Owner task/AT matrix and accepted contract, RP-082–RP-084 repairs under that authority, then R-005/R-008 on built player tasks. |
| Deployment and operations | Deployment Foundation remains accepted and implementing. The `7e8aa70` candidate and previous bundle are historical and invalid as current R-006 inputs. Claude landed R1–R17 through `ec5518b`; Codex's targeted R1/R2/R3 reviews found a destructive wrong-key rollback gap (RP-133), a post-stop rotation race (RP-134), and a valid legacy tar/gzip layer invisible to the secret scan (RP-135). Codex landed bounded corrections at `d19b5d8a`, `2d8ca3f1` and `a7ab2640`, with firing negative tests and relevant cold Go/Compose/vet gates, but none has Claude's designated verdict. R4, R7 and R13 have bounded Codex approvals; R5's composed lane is unexecuted on this arm64 host, and R6/R8/R9/R10/R12 have targeted CHANGES REQUIRED findings (RP-136–RP-140), with Codex corrections awaiting Claude review where implemented. R10 proves host-byte integrity, not decryptable rollback; D-019 asks for an owner recovery-proof contract. R12's plan still permits a same-basename fake producer; DP-F DESIGN-GAP 8 is with the RFC author. Neither party has approved the other's complete relevant ranges, and no rebuilt clean-host player/recovery/supply-chain run exists. Cold `make verify-kernel-version` fails on pushed history; full `verify-client`/`verify-push` cannot be called green despite a passing CI-topology fixture gate. RP-131 also exposes false-positive review provenance in the kernel correction validator. | Cross-party review the exact Claude and Codex Deployment ranges; rule D-019, the per-family overlay and DP-F producer identity under accepted authority, resolve the real red version boundary without rewriting pushed history or a false bump, rebuild both bundles, then execute/sever R-006 and R-007 on the exact clean-host artifacts. |
| Multiplayer, world and later content | Server Commons/guild/faction primitives exist; real social/world/feed and later-tier player journeys are incomplete. | Decide later public social/telemetry scope, then dependency-ordered RFCs, content and integrated proofs. |
| Release governance | Dated audits, defect ledger, research/decision queues, RFC graph and per-RFC logs exist. The executable queue now has a `fc4191fe` overlay, but the underlying earlier routing table remains a dated snapshot. | Reconcile the live queue and this checkpoint whenever a reviewed batch, ruling or release witness changes the critical path. |

## Current next moves

1. Keep the newly landed v0.1 implementation **unapproved** until Codex reviews exact commit
   ranges. MA3's RP-145/RP-146 corrections await Claude's designated review of
   `9ceed1cf^..47562850`; RP-147's GS0.2/Fiscal→Pitch correction at `efe68333` also awaits
   cross-party review. None approves the wider Minigame API/Surface or Garage Surfaces ranges. The
   earlier targeted finding RP-132 requires Pet Adoption PA7/C2 author reconciliation;
   Typer/Arcade public wire and adopted content/copy remain separate open boundaries. Do not
   archive by counting implemented RFCs.
2. Resolve RP-131's kernel-history/provenance gate under explicit process authority. The
   [Kernel History Guard Integrity draft](../rfc/kernel-history-guard-integrity.md) frames the
   path-scope exception and forged-review controls; it is not implementation authority. At the
   current HEAD, the red `verify-kernel-version` result blocks a green `verify-push` claim;
   `origin/main` already contains the offending commits, so no unpushed-history rewrite route exists.
3. Codex designated review of Claude's Deployment corrective range `67fd415..ec5518b`
   (R1–R17, `planning/deployment-foundation/claude-corrective-handoff.md`) and remaining DP-F
   ranges. Codex's targeted R1 RP-133 and R2 RP-134 findings are CHANGES REQUIRED; Claude must
   separately designated-review Codex's R1 correction `5f14a808^..d19b5d8a`, R2 correction
   `8e9b46d0^..2d8ca3f1`, and R3 correction `d871339d^..a7ab2640`. The RFC author must rule on six
   recorded DESIGN-GAPs: RPO semantics, per-family alert
   attribution, `UpgradeResolved`, overlay granularity, the previous-bundle security floor and
   asserted host fields. Rebuild candidate and previous bundles after review, then take them
   through authorized clean-host browser, board-arrival, recovery and rollback witnesses (R-006/R-007).
4. Bring the source-traced [`Phase-0 exact-manifest proposal`](platform-alignment/phase0-manifest-proposal.md)
   to Marco for D-007 adoption or edits; it names P01–P07 player tasks and F01–F05 public-release
   floor obligations without calling them complete. Then bind D-018 and R-008 to the adopted tasks.
   Finish remaining nested-payload classification and clean-host operator proof, then
   prepare the separate account export/deletion/retention decisions. RP-119's predicate-level
   joined probe now confirms stale active counts; the full projection/recompute/player witness
   and owner semantics remain. None of this is optional preview polish.
5. Reconcile v0.1's accepted feature set to its still-unadopted exact release manifest, then
   advance v0.2–1.0 through bounded manifests and dependency-ordered RFCs. Do not implement later
   content directly from this board; the preview floor never replaces the 1.0 obligations.

## Update protocol

At every material checkpoint, append to [`roadmap-1.0-log.md`](roadmap-1.0-log.md): date, exact
HEAD/release artifact, observed change, executed evidence and limits, decision/review provenance,
and the next blocked/ready boundary. In the same reviewed range, update this board's state and
current next moves, the relevant RFC plan/log, and the executable queue if authorization changed.
If a source is stale, mark its audit coordinate; do not copy its old count as current truth.
The goal closes only after the complete 1.0 manifest and all eight dimensions pass on the release
artifact and Marco makes the release call. A Phase-0 or v0.x release advances the goal but cannot
close it.
