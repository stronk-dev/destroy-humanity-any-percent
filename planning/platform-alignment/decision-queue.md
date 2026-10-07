# Owner decision queue

Implementation agents may gather evidence and frame options; they may not infer these choices.

| ID | Owner choice | Evidence required first | Canonical home | Blocks |
|---|---|---|---|---|
| **D-001** | **RULED 2026-08-21:** current claim is repository development snapshot; next target is **Cloud Clicker Phase-0 Playable Preview**, not roadmap v0.1 or 1.0. Public hosting remains blocked on the complete release floor. | R-002 baseline and capability map complete; exact release manifest remains a release-RFC input. | `planning/platform-alignment/owner-ruling-packet.md`; later `design/07-roadmap.md`/release RFC reconciliation by the owner. | Exact preview manifest and every “release-ready” claim. |
| **D-002** | **COMPLETE 2026-08-21:** public repository and tracked public planning/research, after sensitive-material inspection and sanitization. | All 96 ignored artifacts are classified and every disposition is executed. The committed `e44e1a6` verifier rejects three forged states and passes from a fresh clone with 11 public/56 private Class-C dossiers, three duplicates and seven diagnostics. | Research disposition, ignore policy, and ops docs. | None. Future derivatives require their own bounded review; ignored raw inputs remain noncanonical. |
| **D-003** | **RULED 2026-08-22:** public MIT source remains current; target a supported self-host bundle only after R-006; make no covenant/notice/artifact/mirror/ratchet promise in the preview. | Public sunset brief and R-002 complete. | Owner packet; Deployment and later covenant RFC. | Deployment body reconciliation; later covenant remains deferred. |
| **D-005** | **RULED 2026-08-21:** one-time display plus copy/download and a recover-existing-account branch; no email recovery in the preview. Exact loss disclosure remains owner-authored copy. | R-003 validates the later prototype rather than choosing the posture. The present Game UI silently persists the code without a display/copy/download consumer (RP-123); this is a gap to implement under the ruling, not a reopened choice. | Account UX design and copy ruling. | Account recovery UI; export/deletion remain separately blocked. |
| **D-006** | **RULED 2026-08-22:** one Linux Compose node; Caddy-only ingress; one HTTPS origin/one proxy hop; encrypted off-host 6-hour/pre-upgrade backups; 7-day six-hourly + 30-day daily retention; RPO 6 h/RTO 4 h; previous-release rollback for 7 days; exact key overlaps. | Owner ruling complete; R-006 later validates the built package. | Deployment RFC and ops docs. | Deployment body reconciliation and implementation only. |
| **D-007** | **RULED IN PRINCIPLE 2026-08-21:** bounded audited T0–T1 vertical slice, not nine tiers. Exact included surfaces/content rows remain release-manifest work. | Capability map and pacing evidence plus the dated, non-adopted `phase0-manifest-proposal.md` P01–P07/F01–F05 source trace; Marco must accept or edit the exact task/content boundary. | `design/07` roadmap and release RFC. | Content RFC DAG, D-018 exact task population and honest release manifest. |
| **D-008** | Decide whether player data export includes only current state, all save revisions/events, leaderboard history, or a portable self-host import bundle. | Current 60-table/non-DB inventory and all 60 SQL-column maps; browser, core payload, selected core-event/receipt and verification/board diagnostic maps (`data-rights-verification.md`), plus bounded operator map. Other exact historic event/receipt variants, Minigame/Soul/Guild non-receipt fields and privacy/legal review remain. An encrypted whole-instance backup is not one player's export. R-003 validates the later prototype; it cannot be a prerequisite for this choice. | Account/data-portability design. | Export schema and UI. |
| **D-009** | Rule the account deletion disclosure: which anonymized gameplay records remain and for how long, including permanent bootstrap UUID tombstones and backup/device-local residuals. | Five field tranches, browser/core-payload/selected core-event/receipt/verification/operator maps. RP-130 retained one seeded board→archived-Founder→stream link; the current-HEAD composed producer→projector→delete witness now proves that a naturally verified Founder/run remains on the unauthenticated public board (RP-171). RP-118 still lacks a joined dead-letter/poison and backup-restore arm. A separate active-Founder probe found both Commons World-count and health sample-selection predicates still equal one after deletion (RP-119); full projection/recompute remains untested. A seeded joined probe kept active Minigame session/receipt/faucet, active Soul session with token, intent receipt and player outbox rows 1/1 after a 204 deletion (RP-122); composed-coordinator proof remains. Options are framed in `rights-decision-sheet.md`. Soul fixture confirms event→outbox copies and terminal progress-token retention (RP-127), not account deletion. Verification poison details pass bounded unredacted error text into immutable history (RP-129), not a proved secret leak. Archive fixture has no nontrivial payload/event (RP-124); Guild JSON retains a deleted UUID (RP-120), although per-account Guild receipts cascade; client storage lacks deletion (RP-123). Protected pre-deletion backups can outlive 30 days and full restore has no deletion replay (RP-125). Other nested fields, clean-host proof and legal review remain. | Account design/copy. | Player-facing deletion workflow. |
| **D-010** | Decide the lifecycle location for withdrawn RFCs and completed non-RFC maintenance threads. | Active RFC audit. | RFC-0000 amendment if needed. | A mechanically checkable active-work index. |
| **D-011** | **RULED FOR DEPLOYMENT 2026-08-22:** Marco owns official-instance ops/incidents/breaches; self-host operators own theirs; metrics private, rotated logs 14 d, raw IP maximum 7 d, seven blocking alert families, provider-off local alert manager permitted. | Owner ruling complete; R-007 later validates the built operations path. | Deployment RFC and ops docs. | Deployment body reconciliation/implementation; product data retention remains D-009/D-015. |
| **D-012** | **DEFERRED 2026-08-21:** Advisor Mode is not in the Phase-0 preview. | Owner ruling complete; Prestige/design ruling author must reconcile current claims. | `design/11`, Prestige ruling/body or successor RFC. | Prestige body reconciliation only. |
| **D-013** | **DEFERRED 2026-08-21:** composed async minigame lifecycle moves to the first real async-tenant successor. | Owner ruling complete; Minigame Platform author must reconcile current acceptance. | Minigame Platform ruling/body and roadmap. | Minigame body reconciliation only. |
| **D-014** | **RULED 2026-08-21:** retain the strict sub-five-minute push/PR target with a fast harness guard; move exhaustive harness and numeric evidence to a separate scheduled/manual workflow with bounded fail-loud artifacts. | Nine consecutive hosted cancellations plus the accepted local observation identify exhaustive pacing/relevance as incompatible with the blocking lane. | CI Baseline, `docs/ci.md`, workflow topology verifier. | Current-head hosted push/PR success and one complete maintenance observation remain evidence gates. |
| **D-015** | Adopt a complete data-retention policy across accounts, archives, receipts, events/logs, dead letters, projections, moderation records, and application/security logs. | Relation-complete 60-table/non-DB inventory, SQL-column maps for all 60, browser/core-payload maps, bounded core-event/selected receipt/verification maps and operator map. Exact historic event/receipt variants and other nested JSON/binary/text, host-observed logs/alerts and legal review remain. Backup exceptions (RP-125), operator identifiers (RP-126), terminal Soul tokens (RP-127), Minigame parent-delete trigger (RP-128), immutable unredacted poison diagnostics (RP-129), the seeded board→archived-Founder link (RP-130) and naturally projected public-board survival after deletion (RP-171) need explicit treatment. | Privacy policy, Account/Save/operations design. | RP-051/RP-086/RP-087/RP-113/RP-114/RP-118/RP-119/RP-120/RP-122/RP-123/RP-125–RP-130/RP-171, deletion disclosure, cleanup jobs. |
| **D-016** | **DEFERRED 2026-08-21:** no gameplay telemetry in the Phase-0 preview; operational metrics remain a separate unresolved contract. | Revisit only if a later milestone requires it, with privacy/legal evidence. | `design/00`, `design/02`, privacy/operations contract. | Community milestones and live balance calibration, not the preview. |
| **D-017** | **DEFERRED 2026-08-21:** no public UGC/social content in the Phase-0 preview. | Revisit only with compliance and moderation operations. | `design/05`, roadmap, ToS/moderation RFCs. | Later feed/UGC surfaces, not the preview. |
| **D-018** | For the exact tasks owned by D-007's release manifest, adopt the manual assistive-technology/browser/OS matrix, dynamic-announcement floor and whether independent in-game reduced-motion/visual-noise settings ship. D-018 does not choose or narrow the content manifest. | Current-source negatives in `accessibility-current-audit.md`, the bounded Phase-0 content manifest, supported-browser/device posture and the draft `rfc/accessibility-player-workflows.md`. | Accessibility design and accepted successor RFC; owner-authored setting copy if selected. | Accessibility RFC acceptance, R-005 population and any public task-accessibility claim. |
| **D-019** | Choose the proof required before each release to claim its new pre-upgrade backup is restorable: an isolated restore/decrypt with the off-host age identity before drain, a governed key-recipient attestation plus recurring clean-host restore drills, or another explicit contract. The present host-byte checksum and header check cannot establish decryptability; agents may not infer that it does. | RP-139's R10 source/test trace (positive backup creation with an invalid identity fixture), DP6's off-host-key constraint, measured backup size/restore time against the four-hour RTO, and R-006 empty/populated recovery evidence on the exact release manifest. | Deployment RFC DP5/DP6 and canonical operator runbook. | A truthful “working rollback before drain” claim and supported self-host/1.0 release floor; does not block the existing bounded host-byte validation. |
| **D-020** | Reconcile Arcade AC7's required Pitch-less start with API MA-C15's owner-ruled `minigame_api → pitch` dependency: explicitly amend to tenant-independent catalog activation, or retain the dependency and have the ruling author reconcile AC7. | RP-201: freshly hashed complete catalogs with only Pitch removed refuse in both actual Go/TS loaders; unchanged complete controls load. Arcade AR1.2 retains the earlier full chain and names no amendment to MA-C15. | `rfc/minigame-api-and-surface.md` MA-C15 and `rfc/minigame-demo-disc-arcade.md` AR1.2/AR-P3/AC7, authored body reconciliation. | Full Arcade AC7/A4 acceptance; internal starts in complete Pitch-containing bundles remain safe to verify. No independent permission to change public schema, mint or copy. |
| **D-021** | Resolve existing Clout DG-D: have the harness execute shared served foundation hooks, or accept an explicitly specified attainment observer with complete proof/transition parity. No branch selected; agents may not substitute approximate attainment for the actual rule. | Independent current-source probe on Claude 527246f1 removes only first-hour refusal and the unsupported fixture completes neutrally; restored guards/default harness suite pass. Scalar cap tests are not active PR scenarios. Accountable evaluation semantics, provenance/burn/Founder independence, pacing impacts and exact downstream measurement population must precede construction/ratchets. | Clout CV10/AC9 and accepted harness amendment as needed; `planning/clout-v1-and-pr-interns/log.md`, RP-241. | Remaining P6/AC9 PR relevance/dead-row/purchase observations/invariant scenario ratchet and honest Clout/T2 balance; not the approved bounded refusal primitive. |

## Shared attendance continuity — D-024 / RP-365 (unruled)

Real adoption→pause→Feed and deterministic cutoff tests prove clone-only classification
can erase accepted attendance. Required outcome: one monotonic shared clock while genuine
absence remains offline. Proposed direction: durably record accepted Company activity before
Founder sampling; this changes archived A3's explicit non-persisted-clone boundary, so it is
not implementation authority inferred from A5's desired outcome. Author/owner must resolve
that boundary and its transaction/replay compatibility in `rfc/founder-attendance-monotonicity.md`.
Evidence/reproducer: latest Cosmetic log. No pet clamp, timer retune or retry chosen. Blocks
this shared-clock repair and complete care/platform acceptance, not other accepted work.

## Startup first-read failure — D-023 / RP-342 (unruled)

Predeclared ced3b5c7 actual browser runtime/decoder/native host measurement
confirms a completed failed first read remains visibly loading. Existing
visibilitychange can reread; there is no recovery control on the no-snapshot
page. Repeated401 does not renew or create a replacement account. Exact rows
and limits: `planning/garage-player-surfaces/first-read-observation.v1.json`.

Owner choice: show failure with an explicit read-only Retry (recommended in
the draft), or show failure while retaining only the existing lifecycle
reread. Neither permits credential clearing/bootstrap, token-policy changes,
hidden writes or misleading loading-after-completion. Adopt exact copy and
failed-read state precedence in `rfc/game-ui-first-read-recovery.md`; its
complete operation/refusal/identity boundaries require review before acceptance.
Garage's author reconciles any conflicting GS0.5 loading condition if adopted.
No ruling or implementation claimed here. D-005 recover-existing-account,
draft browser-session-renewal and privacy/retention stay separate. This blocks
the new startup repair, not already accepted Meters shared-state evidence.

## Clout executed role binding — D-022 / RP-305

CV1 item5 makes truthful executed roles an explicit pre-mint condition. The two
proposed PR upgrades declare synergy_feed, but the read-only fixture census
c99da6 finds no synergy-pool source for either. Existing upgrade-role validation
checks names/duplicates, not execution binding. A nonempty label is not proof.

Owner/RFC author must either provide a real declared pool binding, with its
values measured/ratified under CV8/OD-5, or reconcile the role contract/vocabulary
under OD-3 with an explicit executable meaning. Canonical intent/spec homes:
Clout CV1 item5/OD-3 and, if changed, the purchasable-content role law. No choice
made here; no pool/rate/role restriction or owner ruling fabricated. Blocks
truthful CV8 content and full CV1/P1 approval, not RP-304's independently defined
nonempty-axis-role correction. D-021 harness authority remains separately open.

## Garden unnamed resource event — RP-229 (author reconciliation pending)

Accepted Garden SG8 calls for `garden_harvest_credited.v1` "plus the ordinary resource event";
Minigame Platform C21 uses the same unnamed premise. The closed save event registry has no
generic resource-credit event kind, and actual Garden live/replay emits its named credited
event only. This is a specification-author clarification route, not an owner decision
silently delegated to the implementer or a request to invent a new resource event.

Required action: identify the existing event/consumer contract that the phrase intends and
reconcile its body, or explicitly route genuinely new event/schema behavior through its own
accepted authority. Canonical home: `rfc/minigame-server-garden.md` SG8 and any parent clause
it relies on; source/diagnostic evidence in `planning/minigame-server-garden/log.md`, RP-229.
No new event, migration, replay version or ordinary player-mechanics choice is authorized.
The bounded RP-228 saved-head/hash proof does not resolve or bypass this question.

## Garden starter-set evolution — RP-222 (unruled)

The 2026-10-06 G2/G3 diagnostic under `d6c2b6a2` / `1aa23496` admits an epoch retune
that promotes uncollected `strain_c` to starter. Both Go/TS loaders and transition validation
admit it, but unchanged SG2 state fails the next catalog's starter-superset invariant; the
actual Go foundation boundary rejects it. The ordinary harvest-value retune succeeds.

The author/owner must reconcile SG1 rule 11's value retunes with SG2/AC11's byte-identical
permanent carry and starter-superset rule. Options require an explicit ruling: make starter
membership an epoch invariant, or define/version a specific carried-state grant/migration
exception. Neither is inferred here. Changing ordinary balance values remains safe to verify;
publishing a starter-changing Garden epoch and full G2/G3 acceptance remain blocked on this
named contract question. Garden is not publicly minted; no current-player failure is claimed.
Canonical home: `rfc/minigame-server-garden.md` SG1/SG2/AC11, reconciled by its author;
evidence in `planning/minigame-server-garden/log.md` and `design/BACKLOG.md` RP-222.

The formerly blocked GU-C25–GU-C28 authored action is complete; Game UI is archived and its
archival transaction is designated-approved at `f199f9a`.

Resolved input, not an owner question: `design/11 §1b` adopted silent server-anonymous as the
default and local-only play as the labeled outage fallback. `design/06` and Account D4 still use a
broader “may run fully offline” formulation. Their authors must reconcile the bodies and specify
the fallback contract; an implementation agent may not re-open the already chosen default.

## Reputation Run End data boundary — RP-284 (author reconciliation pending)

Accepted Reputation R9 requires available Reputation after Exit and R7's
run_started v2 carry-over summary. Archived Game UI GU-C3 instead requires
RunEndSurface to accept only decoded run_ended; its compile-time negative
explicitly forbids a snapshot prop. The current ended payload has payout delta,
not level/spent/available. Founder-advanced has delta too; the receipt's Company
state is not the Founder feature. A latest v4 snapshot supplies current available
balance, but is not an immutable balance captured by that particular Exit.

Required author action: reconcile the normative display/data boundary and name
the authoritative balance read, its timing/identity, permitted component inputs,
and the run_started/retained-v1/null handling without auto-dismissing Run End.
Either explicitly permit a guarded current-balance companion read, or specify
an Exit-bound event/receipt addition under accepted schema/version authority;
neither contract is inferred by the implementer. Clarify whether the displayed
balance is current or captured-at-Exit rather than silently choosing between them.
Canonical home: Reputation R9 and its explicit GU-C3 amendment/reference, with
the ruling author's body reconciliation. This is not a reopened product choice
to have Reputation or permission to edit another author's ruled text.

Safe reader supplement in the range after0ae1fa11 does not resolve this gap:
it decodes/delivers the existing immutable summary only, with executed positive,
malformed, race and duplicate controls. Next recovered-publication reader checks
may proceed independently. No UI, payout formula, balance math, new wire or
whole AC12 is authorized by this finding. Evidence: Reputation implementation
log and design/BACKLOG.md RP-284. RP-283's eligibility-only Wind Down preview
bridge is a separate author question, not cured by this event reader.

## Reputation plan reconciliation on authoritative update — RP-292

R9 specifies advisory selections, projected available and server revalidation,
but not what happens to an already selected plan when authoritative props change.
Both-era native component observation7076ef/9a1aab shows a selected node becoming
owned: its checkbox disappears, yet the callback still retains the id and the
formula subtracts its cost again. A budget reduction retains enabled checked
selections with projected balance below zero. Native Clear restores a coherent
empty plan; valid formula/ownership update controls pass. This is controlled
public-prop evidence, not proof of a real concurrent player/host/SQL occurrence.

Required author action: specify what invalidates a plan and reconcile the R9
body with the chosen callback/visible-selection behavior. Choose and specify
reset-to-empty, deterministic pruning, or an explicit invalid-plan/refusal state,
including how users learn of the change and whether valid selections survive.
An implementer cannot silently choose a pruning order or discard user intent.
Any new player-facing text requires the existing owner-copy mechanism. Canonical
home: Reputation R9's plan panel contract, with the author's body reconciliation.
No change to R6's atomic server revalidation, Wind Down's open door, payout math,
epoch/price, or RP-283's separate authoritative-preview bridge is inferred.
Tests named CHARACTERIZATION ONLY record the current limitation; their green
result is not AC12 or approval of this UX. Evidence: Reputation implementation
plan/log and design/BACKLOG.md RP-292. Other accepted work remains available.

## Reputation published formula units and parameter domain — RP-293

The accepted R9 copy contract names integer `per_level_percent` and
`unlock_percent`; actual candidate/generated metadata uses integer `perlevel`
and `unlock`, and the component supplies raw ppm. The reproducible
`node planning/reputation-tree-v1/header-copy-contract-check.mjs` exits1:
five balance/factor controls match, formula metadata does not. Its self-test
uses synthetic metadata only and rejects four corruptions; not prose approval.

R2 permits all integer ppm values1..1,000,000. An actual TS-loader-admitted
1ppm/1ppm tree reproduces the type conflict: each percentage is0.0001, and
the existing integer formatter rejects it. Other header accounting/frozen-next
bindings pass native tests; current raw-placeholder rendering is specifically
not proof of the required published arithmetic. Owner prose remains pending.

Required author action: reconcile R9's representation, parameter names/types,
units and null/current-run handling with R2's existing exact input domain and
R3's `1 + level × per_level_ppm × unlock_ppm / 1e12` arithmetic. Specify an
exact percentage representation or an explicitly unit-labelled ppm formula;
do not quietly round percentages or narrow admitted R2 data to convenient
multiples. The resulting prose still follows OD-11's owner-adoption mechanism.
Canonical home: Reputation R9 copy/header contract and its author-reconciled
body; any generated metadata/component repair follows that accepted contract.
No global integer-validation weakening, new numeric kernel, retune/epoch mint,
report restamp or implementer-authored player text is authorized. RP-283 preview,
RP-284 balance/display and RP-292 plan policy remain separate questions.

## Clout offline-episode meaning — RP-308

Actual R-012 policy research predeclareda25c6d5e executes44paired full-state
profiles with independent integer policy bindings. One48h offline Evaluate
produces24h and banks43,200,000ms; two24h calls produce48h and bank0. The
current per-call cap/floor/bank saturation is canonical behavior, not a defect
this evidence authorizes changing. A pure computational split of one absence
and multiple capped catch-up histories are distinct questions; clients never
select server time or evaluation mode.

Required author action: reconcile CV3/AC6's unqualified "any interval" with
the intended catch-up episode boundary while preserving the existing24h cap,
90% efficiency, banking and separate-reconnect behavior. The owner was asked
whether to delegate that wording reconciliation; no response or authority is
inferred from preselection/automatic goal continuation. If the owner chooses
different behavior, it requires its own explicit design/contract change, not
an implementer convenience. Canonical home:Clout CV3/AC6 and its author-reconciled
body, with linkage to any later accumulation/persistence RFC. This cannot waive
RP-307's numeric failures, invent episode state, restamp evidence or retune data.
Evidence:planning/clout-v1-and-pr-interns/policy-boundary-research.md and log.

## Garage active-play projection body reconciliation — RP-313

Original producer `cdb8fe61^..cdb8fe61`, all twelve paths, designated-reviewed
by Codex: CHANGES REQUIRED. GS0.1 requires exactly six feature keys and nullable
`active_play`; current code keeps that key null and uses optional `opportunity`,
whose presence drives `feature.active_play`. The implementation log cites API
C2's type-change refusal, but no author body reconciliation adopts the sibling.
Pure existing-kernel projection is read-only in executed tests, yet directly
reads the source instead of the body-promised discarded clone.

Required author action: reconcile GS0.1/GS5 arm location, absent/null contract,
fact binding and read-only/clone boundary with C2 and the retained v4 baseline.
Preserve the accepted gameplay and compatibility obligations; no implementer
re-pin, v2/v5 choice, gratuitous cloning or silent accepted-body amendment.
If a changed runtime contract is selected it needs its own bounded accepted
authority and discrimination tests. RP-132/RP-174 are analogous, not permission
to infer this ruling. Eight executed faults and the cold local/real-Postgres
evidence are in the Garage log. The separate Desk consumer can be inspected
diagnostically meanwhile; no approval, archival or player-content adoption is
inferred. New Codex tests/records require Claude separately.

## Garage initial Standing criterion — RP-363 (author reconciliation pending)

Accepted GS3-A4 names fresh Standing50/Grievance0/doom50. Predeclared807e0b6e
actual fresh DOM bootstrap/public read, before any gameplay or server setup,
fails f3b5ed→b42d86: all five Standing axes are90/high; Grievance0/low and
doom50/low match. Replayable candidate and exact source pins:
planning/garage-player-surfaces/initial-meter-dom-probe.patch and log.md.
No DOM or source-severing population passed; the original driver is restored.

Actual chain: gameserver attaches production.FounderInitializer; Account
initialStates initializes zero Notoriety and invokes it; shared foundation
assembly calls meters.NewRunState, which applies the pinned trust_reseed
base90 at Notoriety0. Meter artifact literal standing initial50 is distinct.
Archived Meters M5/C9/C13 prescribe Standing reseed at new-run assembly.
Those clauses do not silently amend Garage GS3-A4 or independently prove a
first-Founder exception was accepted. This is a conflict requiring exact
author reconciliation, not permission to decide a balance fix from a red test.

Required author action: reconcile Garage GS3-A4 with the published assembly
rule and explicitly name its fresh-Founder/run identity, Notoriety and expected
initial values. If fresh bootstrap is meant to use literal50 instead, specify
that exception under accepted shared live/replay/account authority before
implementation; do not alter the artifact, existing histories or replay/migrations
to satisfy one browser check. No choice/body edit/owner ruling claimed here.
Blocks RP-362/all GS3-A4 acceptance, not independent GS2-A4 or investigation
of the separately reproduced Cosmetic Buy timeout RP-364.
