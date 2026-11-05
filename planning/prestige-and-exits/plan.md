# Prestige & Exits — implementation plan

- **Assignee:** Codex
- **RFC:** `rfc/prestige-and-exits.md`
- **Started:** 2026-07-29
- **Current closeout:** D2–D5/P3–P5 and canonical docs reconciled to current behavior and D-012;
  AC2–AC6 witnesses have Claude's designated approval (`fbc2452^..c50744d`), and item6's
  epoch-8 pacing evidence has Codex's designated approval (`a965538e^..a965538e`; see log).
  The final P5c/P2d corrective four-path range `d9d8ad59..4589a642` now also has
  fresh independent Codex approval under Marco's corrective-review delegation.
  The20-path pre-timer/epoch correction `6f4704d9..ec2460b1` and complete
  Exit-credit correction `408517bf..1ccdb7a0` have bounded independent approval too;
  historical kernel identity, RP-771 and original feature acceptance remain open.
  P6c's16-path policy-pin correction `498f8ae5..e5ad0bac` now also has fresh
  independent corrective approval, with old/new persisted-pin discrimination.
  Next: reconcile complete foundational implementation/review coverage for item9;
  these scoped verdicts do not authorize archival or approve adjacent changes.

**AC7 run-end data defect,2026-11-05 (RP-773):** actual current-epoch8
Go/TS replay pairs end at distinct Founder ages but emit identical `run_ended`
events; the event-only screen cannot display the required correct Founder age.
The owning log retains the cold failing requirement regression, actual UI decoder/
term-consumer observation and exact manual probes. Independent evidence/authority
review approves `f2a2c82f..c195c0ae` and confirms D3/AC7 authorizes the bounded
versioned correction. Corrected input v13 / ending v4 freezes
the exact committed age, old v2–v12 outputs remain unchanged, and the UI decoder
preserves age without a live-state read. New/old Go/TS, real Postgres insertion/
refusal/current curriculum, worker/archive and board checks pass; full native
Production still has the known Clout partition failures. Complete44-path
`c195c0ae..a12218ab` has independent corrective APPROVAL, with fresh Go/TS/current
Postgres/archive/board/migration checks and disclosed nonblocking migration EOF
whitespace. This closes the data defect, not full AC7. New player copy remains
owner-authored; the age label and full obituary rendering remain open.
The kernel227/migration87 runtime now postdates the kernel226 `c5130acf` local
candidate; rebuild only after this coherent correction's independent review.

**Closed correction,2026-11-05 (RP-772):** the old actual epoch8 curriculum
producer saved75 earned Knowledge, but a real projector made the affordable50
hint reject: saved75/projection0/hint false (private real-PG68322). Complete
`408517bf..1ccdb7a0` corrects the Exit-credit → projection → hint-consumer binding
under accepted Routes D5/C6 and Prestige D3 using transactional Exit/hint event claims
and one Founder-row lock. Fresh independent Codex review approves all12 paths;
its private pending-debit severing fails the second-purchase assertion.
The current-content regression now passes with the real projector, held hint delivery,
unaffordable second purchase and exact receipt retry. Projector checks cover late/duplicate
delivery, existing/missing rows, rollback, invalid ownership, concurrent repair and Registry
debt repayment. Migration00086 adds the Exit claim table; canonical docs and rights inventory
are updated. Full affected production/server SQL and client curriculum/replay checks pass;
retain exact commands/results/limits in the log. Preserve current numeric capacity and source/replay wire.
This ordinary-value correction is independent of RP-771's pending capacity choice.

**Current AC5 defect,2026-11-05 (RP-771):** actual Postgres proves eligible
Wind Down fails at the admitted Route Knowledge maximum, including a companion
owning every hint. Full state/revision/history rollback holds; the original six
matrix rows still pass. [Retained red regression and contract boundary](route-knowledge-capacity.md).
Resolve counter capacity/preview/grant semantics before repair; do not infer
Route Knowledge saturation from Reputation P2c, narrow valid states or claim
universal AC5 from the older bounded matrix. No code change is adopted here.

1. [x] Add save v7 Company/Founder state, validation, migrations, and corpus fixtures.
2. [x] Implement cross-runtime prestige arithmetic and declarative Phase-0 policy.
3. [x] Implement deterministic offers, timers, decline, wind-down, accept, and gated IPO intents.
4. [x] Implement atomic two-stream Exit persistence and exact event schemas.
5. [x] Add real-Postgres fault, replay, and concurrency tests plus arithmetic properties.
6. [x] Reconcile the already-implemented first-elective-Exit harness evidence in the same
   designated-review range as this checkbox flip; do not infer completion from the older audit.
   Evidence: `first-hour-epoch8-report.v1.json` (current epoch 8, 97 runs, 0 failures) and the
   2026-09-24 log entry, including the lower-bound caveat; designated Codex review recorded
   2026-10-15, with the headless/composed and branch-population limits explicit.
7. [x] Update canonical docs and run focused verification.
8. [x] Replace the provisional terminal revision with Leaderboards L1's atomic run-log sequence.
9. [ ] Record independent review before archival.
10. [x] Add the missing literal AC2–AC6 witnesses: offer-age/progress property, full-domain reseed
    and repeated non-empty ledger preservation, New-Founder lifecycle, eligible-state/event-chain
    matrix, and checked-in run-2 golden. Each witness needs a demonstrated failing mutation.

## Concrete remaining implementation-review request — 2026-11-05

Authority is the accepted Prestige RFC, not a new mechanic or specification
adoption. This identifies original review work; it is **not** an exhaustive
archival coverage manifest or an assertion that all later integration is approved.
Use post-rewrite coordinates; anonymous legacy prose cannot supply reviewer identity.

| Batch | Exact base-exclusive/tip-inclusive range | Present review disposition |
|---|---|---|
| Original save, arithmetic, atomic store and lifecycle | `9e2ccc6f5dbb8660a5cad32f85865c1b87a3c200..c3c58b950dd8908a6ba09a0b185a98dce930a134` | Four construction commits/35 paths. Historical findings are retained; new explicit designated implementation review needed. |
| Original run-log integration | `cbe0ad968c11faa9e8e3d04309907e5d0ba8e987..ce0a5ece584daace9b91f4ba14d12796698df928` | One construction commit/16 paths; explicit designated coverage needed. |
| Pre-timer/cross-epoch correction | `6f4704d9b2824100ca61910fd0b7d310f2a7916b..ec2460b115602720ead0150aca53667367a1ef31` | Fresh independent corrective APPROVAL2026-11-05: all20paths inspected, current1ccdb7a0 executed with actual legacy-v6/cross-epoch/refusal/board checks. Original kernel identity remains unresolved; no adjacent/original-feature approval. |
| Lifecycle review remediation | `2359c4f7668828964002a628b12a3cf7edfce6e1..0a7f129280f5307a2add8b3cdcc676a7fd5f27be` | Fresh independent corrective APPROVAL only as the exact corrected union with d9d8ad59..4589a642; historical0a7f alone/intervening commits are not approved. Current RP-771 remains a separate AC5 finding. |
| Hash-pinned attended ceiling | `498f8ae51fda9d0e71470ac3ce3c2da6f98f55a1..e5ad0bacd056dce983376a5b7d97256877a13f1e` | Fresh independent corrective/technical APPROVAL2026-11-05, all16paths inspected and current21a889b7 executed. Actual old/new pin and constant-override controls discriminate; anonymous legacy prose stays unattributed and historical kernel identity remains unresolved. No adjacent/original-feature approval. |
| Final P5c/P2d correction | `d9d8ad590b1fade049e3a804f114673d2ec92373..4589a642cc1c951fd73ec0d50f37d7c9471a119b` | Fresh four-path corrective APPROVAL,2026-11-05; do not repeat. |

The already-approved `fbc2452^..c50744d` witness/body range and
`a965538e^..a965538e` pacing-evidence range remain separately attributed in the
log. Later replay/genesis, active-play/Fiscal/Minigame, Reputation/Pet/Cosmetic/
Clout/Garden and ending-preflight changes must be joined to their own exact
verdicts before any full archival union claim. Approval of original code cannot
approve those descendants by implication.

The designated reviewer must inspect original diffs AND current consumers,
distinguishing historical inspection from current execution. Start with cold
Go Prestige/Save unit checks, client Prestige/shared-vector and relevant replay
checks, and affected vet/types. Use actual Postgres for
`TestApplyExitTransactionAtomicFaultsAndReplayIntegration`,
`TestPrestigeWindDownAndScriptedExitIntegration` and
`TestPrestigeWindDownEligibleStateMatrixIntegration`; check populated Founder
ledger carry and checked-in run-2 golden within those populations. The real
gameserver New-Founder path and composed first-hour proof, plus the governed
97-run pacing report, own AC4/AC8's additional boundaries. Inspect/reuse their
source-bound approved evidence when unchanged; rerun affected boundaries when
necessary, not all unrelated suites. RP-770's retained race failure and the
known27 accounting/history-guard failures cannot disappear behind focused green.
AC7 additionally requires current ended-event/template-slot binding, not merely
an event existing. Name any missing criterion as a finding; do not weaken it.

Current delegation permits independent Codex corrective/technical-spec reviews,
not the original feature-construction acceptance review above. That review must
use the existing designated party unless Marco explicitly extends delegation.
No new product/copy/specification adoption or release permission is requested
by preparing this review scope; item9 and RP-038 remain open.
