# Phase-0 Playable Preview — exact-manifest proposal

**Status:** evidence-backed proposal for D-007; **not an adopted release manifest or RFC**.
**Coordinate:** 2026-09-23; product source `7e8aa70`, planning HEAD `e7d61f6` before this
record. The product-path diff between those coordinates is empty. Method and refusal controls
are in [`phase0-manifest-plan.md`](phase0-manifest-plan.md).

## Boundary already adopted

D-001/D-007 choose an honest T0–T1 preview, not the v0.1 Garage or nine-tier 1.0. Public hosting
is blocked until recovery, rights, accessibility, privacy and operations work is implemented and
rehearsed. The adopted deferrals exclude Advisor Mode, composed asynchronous minigames, gameplay
telemetry and public UGC/social content. These statements come from the adopted section of
[`owner-ruling-packet.md`](owner-ruling-packet.md), not from the proposal below.

Epoch 8 pins 19 artifact families at
`sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6`
([`balance/epochs/phase0.json`](../../balance/epochs/phase0.json)). Its economy includes one
manual action, nine generator classes and ten upgrades. Its T0–T1 curriculum has three
first-ending branches. **Pinning an artifact is not promising every backend family it contains
as a player-facing preview feature.** In particular, the pinned achievements and pets catalogs
have no mounted player task in the current Game UI; this is the negative control from the plan.

## Candidate player-task manifest

`P` means propose inclusion as a player task for owner adoption; it does not mean release-ready.
Evidence is at the named current paths, with the witness limit in the last column.

| ID | Candidate preview task | Current producer → consumer → content | Executable evidence and remaining gap |
|---|---|---|---|
| P01 | Start a new anonymous Founder and reach the Desk. | `/api/v1/bootstrap` → Vision Slide and credential storage → epoch-8 opening state. | `client/tools/test-game-ui-composed.mjs` performs real Postgres bootstrap/browser entry. One-time recovery credential is only stored, not handed to the player; D-005/R-003 still block release. |
| P02 | Take a manual action, observe cash/rate/cap, buy a generator and an upgrade. | Authoritative intents/snapshot → Desk controls → `balance/catalogs/phase0.json` (one action, nine generators, ten upgrades). | T0–T1 97-run harness and composed server/Postgres seed exercise first manual/generator/upgrade; the composed browser journey does **not** click all of them. Every included visible content row needs row-sensitive player-path or justified representative proof. |
| P03 | Cross the first Gate into the Garage when eligible. | Server-projected `gate.t0_to_t1` eligibility and `cross_gate` intent → Desk control → first route/gate. | Designated-approved composed browser journey clicks the enabled control through real API, with severing evidence. Other registered gates/routes are not implied. |
| P04 | Reach, understand and continue from the scripted first-company ending, including its branch/starter consequence. | Curriculum/`run_ended` v3 → Run End and **Start the Next Company** → acquihire, burnout or pivot starter. | The 97-run harness and separate branch proofs cover timing/starter behavior; the composed server/Postgres seed covers one ratified path, and the browser covers one scripted terminal and exact successor. Branch-specific browser understanding and nontechnical R-008 remain open. |
| P05 | In the next Company, return through the first Gate and take the later elective Exit; see terms/decline/accept when an Offer occurs. | Prestige and lifecycle events → Desk, Offer Sheet, Run End → epoch-8 prestige/curriculum. | Composed browser covers Gate, standard terminal, continuation and an Offer preemption path; governed 45–90 minute timing lives in the headless harness, not the browser. Offer focus/announcement is currently defective/unproved. |
| P06 | Close/reopen or reconnect and recover the authoritative state without fabricating progress. | Saves, subscriptions and snapshots → runtime recovery/resync/offline notice → current Founder state. | Composed browser severs WebSocket and checks positioned replay; Q-003 covers transport convergence. Expired credentials, second-device recovery and the player-facing recovery branch are not implemented. |
| P07 | See the bounded T0–T1 presentation and honest zero-price satire while playing. | Copy/Presentation catalogs → five mounted surfaces, local timer/splits, cosmetic and order-form controls. | `make copy-check`, Game UI/browser gates and player copy adoption cover current strings. R-008 must still check comprehension and misleading controls; no paid/shop transaction is promised. |

Five mounted surfaces are `vision_slide`, `desk`, `offer_sheet`, `run_end` and `settings`
([`client/src/game-ui/surfaces.json`](../../client/src/game-ui/surfaces.json)). Settings currently
shows save status and an account note, **not** recovery, export or deletion. The player tasks
above therefore describe the intended release obligation and disclose current gaps; they are
not a list of completed features.

## Mandatory public-preview floor, regardless of optional content

| ID | Release task or operator outcome | Current boundary / owner |
|---|---|
| F01 | Give the recovery code once with copy/download, then recover an existing account across a fresh browser/device; handle expiry, replacement and loss honestly. | D-005 posture adopted; account UI/copy and R-003 incomplete. |
| F02 | Export and delete the player's account through an actual UI/API workflow with a per-store disposition and truthful retention disclosure. | D-008/D-009/D-015 unresolved; 60-table inventory and retained bootstrap tombstone recorded, not a legal decision. |
| F03 | Complete every adopted P/F player task with keyboard/assistive technology, 320 px reflow, dynamic lifecycle focus/announcements and motion posture. | D-018 matrix open; current three-engine probes fail focus and reflow while ordinary browser CI passes; R-005 open. |
| F04 | Install, operate, back up, restore, upgrade, roll back and alert on the supported single-node bundle with the ruled RPO/RTO/privacy limits. | Deployment Foundation accepted and implementing; R-006/R-007 exact-artifact clean-host proof and designated review incomplete. |
| F05 | Meet current-head CI, negative/severing witnesses, source/artifact rights and canonical disclosure, then make an explicit owner release call. | Historical green code CI and local supply-chain checks are narrower than a hosted playable preview. |

The owner may **not** remove F01–F05 merely by calling the milestone a preview; they follow the
already adopted public-hosting floor. F03's exact manual device/AT matrix and optional in-game
settings are still D-018 choices, not inferred here.

## Explicitly outside this candidate promise

T2–T8; the three 1.0 terminal endings; minigame player surfaces (including Pitch and Soul
Recovery), pets/care, achievement/Clout readers, factions/guild governance, live feed/world/UGC,
ranked combat, community milestones, Advisor Mode and asynchronous minigames. Backend catalogs,
routes, validators or operator fixtures for these may remain in the pinned bundle without turning
them into promised preview tasks. The absent mounted pet/achievement paths satisfy the plan's
negative control. None of these deferrals overrides required privacy, preservation or security
handling for data the running service actually stores.

## Decisions and proof still required

1. **D-007 adoption:** Does Marco adopt P01–P07 and F01–F05 as the exact Phase-0 task boundary,
   with the current epoch-8 T0–T1 economy and three first-ending branches? If any visible
   generator/upgrade is excluded, specify how it is hidden or why the preview honestly exposes
   it without making a content claim. State any intended inclusion not listed here explicitly.
2. **D-018 dependency:** after D-007, bind the keyboard/AT/OS matrix and lifecycle announcement
   contract to these exact tasks; do not assume the present axe check proves them.
3. **Release RFC:** turn the adopted manifest into versioned in/out criteria and a task-level
   producer→API→surface→content→proof table. P02 and P04 currently lack full row/branch-sensitive
   browser proof; R-008 must test player comprehension. Keep R-006 tied to the exact candidate
   bundle, not a generic green checkout.

This is a static source/evidence synthesis, not a completed player study, legal analysis or
clean-host rehearsal. No product code or copy was changed to make the proposal appear true.
