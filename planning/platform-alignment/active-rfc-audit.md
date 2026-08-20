# Active RFC lifecycle audit at `190a4fa`

The status below is evidence-derived. It does not change an RFC's formal status by itself.

| RFC | File status | Reality at HEAD | Required next gate |
|---|---|---|---|
| RFC process | accepted | Active governance, not product work. | Keep. |
| Balance Harness Dispatch Integrity | withdrawn | Premise refuted; retained for history. | Decide whether withdrawn RFCs belong in active index or a non-implemented archive lane. |
| CI Baseline | implementing | Most lanes work, but no current-head workflow finishes because harness times out. Plan's “first hosted run” framing is stale. | R-001 diagnosis, discriminating repair under this RFC, current-head hosted success, full-range cross-party review. |
| Account & Session Bootstrap | implementing | Backend scope is extensively implemented and documented; plan/log stop before a modern full-range archival verdict and user account operations have no UI. D4's broad offline wording is unreconciled with `design/11 §1b`'s later server-default/local-fallback ruling. | Author reconciles the default/fallback body; re-walk RFC acceptance and range union; route UI/export/fallback gaps to a successor instead of silently widening this RFC. |
| WebSocket Transport & Fan-out | implementing | Gameserver composition and live browser consumption now exist, contradicting open plan item 4. | Reconcile plan/docs against successors; audit full implementation range and carried findings before archival. |
| Leaderboards & Balance Epochs | implementing | Core storage/projection exists; plan still leaves initial-state/archive compaction, docs/integration, and review open. | Finish accepted scope or explicitly split successors; no completion claim yet. |
| Prestige & Exits | implementing | Exit machinery and first-hour payoff exist; the content-dependent gate appears to have landed but the plan was not reconciled. | Re-run acceptance against live T0–T1 content and obtain complete review coverage. |
| API Foundation | accepted / implementing | Registry/generation/middleware exist. Public DTO/readers, evidence endpoints, router, privacy proof, docs, and review remain open. | Complete only the accepted public API scope. |
| Minigame Platform Foundation | accepted / implementing | Persistence, tenant boundary, payout/replay, composition, and The Pitch artifact were implemented across this and successors. However, the header/body/AC6 still claim a shipped combat-duel tenant while the later C9 text, draft combat RFC, absent catalog, and production registry prove otherwise; “body reconciled” is false. | Ruling author reconciles the normative body and AC6 first; then cross-RFC acceptance/range-union audit. Do not archive from the stale plan. |
| Combat Shared Data & Arithmetic | implementing | Arithmetic/RNG done; catalog, complete fixture, Trust/Soul vectors, lint, docs, and review open. | Complete parent before duel/lane engines. |
| Combat Duel Engine | draft | Acceptance review done; owner rulings and implementation absent. | Parent completion, owner rulings, acceptance. |
| Combat Lane Engine | draft | No implementing plan. | Parent contracts and owner acceptance. |
| Combat Bots & Integration | draft | No implementing plan. | Duel/lane engines plus account/minigame integration. |
| Game UI Screens | accepted / implementing | Phase-A screens and live transport exist; AC1 body contradicts the accepted GU-C25–C28 rulings and the full browser first-hour is absent. | Ruling author reconciles normative body; then implement and prove AC1; full-range designated review. |
| Commons Onboarding & Governance | draft | Backend Commons foundation exists; player-facing half absent. | Owner accepts reconciled draft before implementation. |
| World Layer Foundation | draft | No implementation. | Resolve release/content ordering and upstream Commons/production dependencies. |
| Feed & Dispatch Foundation | draft | No player feed implementation. | Transport/API/achievement dependencies and owner acceptance. |
| Events Engine Layer 1 | draft | No evaluator/content workflow. | Meter/achievement semantics and owner acceptance. |
| Minigame & Recovery API + Surface | accepted / implementing | Backend API lifecycle is composed and reviewed in slices; surface and complete range review are open. | Implement MA-C9 surface only after Game UI/UI dependencies; full-range review. |
| Permits & T3 Gate | promoted | Candidate and atomic epoch-6 activation ranges are designated-reviewed; AC1/AC2 are cold-proven. AC3 lacks the exact two-resource Go/TS replay row; P2/P4/AC4 contradict their rulings; canonical docs are pre-mint. | RP-027–RP-029: author body reconciliation, missing witness, docs correction, then a new complete designated range before archival. |
| First Content Epoch | implementing | Epoch 6 mint and repairs were designated-approved; AC1/AC3 are proven. AC2's fixed witness drifted to epoch 8, AC4 lacks exact range citations, AC5 body contradicts its range-head ruling, and epoch-7 harness scope was grafted into this bounded mint RFC. | RP-030–RP-033: reconcile authority/successor scope, restore exact witness and provenance, then transactional closeout. |
| Deployment Foundation | draft | Draft's central “local-only / THE PUSH” premise is false. Production packaging is still absent. | Owner rules release floor and repository disposition; author rewrites the body before acceptance. |

## Planning-record defects

- Multiple plans end in July/August blockers already resolved by later RFCs and commits. A stale
  checkbox is not evidence that work remains, and later code is not permission to flip it without
  an acceptance/range audit.
- `planning/run-genesis-archival-remediation/` remains outside `planning/archive/` despite its last
  blocker being recorded resolved.
- Completed maintenance threads (`ci-remediation-2026-08-10`, `harness-dispatch-cardinality`,
  `archived-four-review`, `production-review-round2`) occupy the live planning namespace without a
  declared lifecycle.
- The coverage map's generated-from-slices claim is not backed by a checked-in generator, and its
  six source slices have not been revalidated since 2026-08-05.
