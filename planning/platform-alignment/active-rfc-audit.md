# Active RFC lifecycle audit at `190a4fa`

The status below is evidence-derived. It does not change an RFC's formal status by itself.

| RFC | File status | Reality at HEAD | Required next gate |
|---|---|---|---|
| RFC process | accepted | Active governance, not product work. | Keep. |
| Balance Harness Dispatch Integrity | withdrawn | Premise refuted; retained for history. | Decide whether withdrawn RFCs belong in active index or a non-implemented archive lane. |
| CI Baseline | implementing | AC2/AC4 are proven and five individual hosted jobs are green, but AC1/AC3 fail because harness cancels beyond 30 minutes; AC5 is false on build-cache/nightly topology. The four-job/sub-five-minute/harness-excluded body, plan, docs, and review union are stale. | RP-021–RP-023/RP-056–RP-059: author scope/latency ruling, accepted R-001 instrumentation, measured repair, workflow-contract fix, current hosted success, exact cross-party union. |
| Account & Session Bootstrap | implementing | Backend account/security/composition is cold-green; AC1/AC4 are proven. AC2/AC3/AC5–AC7 retain literal cross-system/repeated/all-stream/all-route gaps. The Game UI never refreshes its 15-minute access token; account/recovery/import/delete/fallback surfaces are absent; D4 and review history are stale. | RP-004–RP-007/RP-024/RP-048–RP-051: body/owner reconciliation, witness-only backend closeout, accepted player-session/rights/fallback successor, exact current-history cross-party review. |
| WebSocket Transport & Fan-out | implementing | Server fan-out/recovery/drain is cold-green and AC1 is proven, but the production browser has no position persistence, cursor binding, recovery, reconnect, typed-close handling, or drain delay. AC3/AC6 also lack literal witnesses; body/plan/review history is stale. | RP-052–RP-055: implement the accepted recovery consumer, close exact witnesses/body-plan truth, translate and union current ranges, then obtain cross-party review. |
| Leaderboards & Balance Epochs | implementing | Epoch/replay/projection mechanics are cold-green and Run Genesis later completed the stale plan's replay/compaction item. AC5 is false (join→leave→Exit ranks Solo); AC1/AC3/AC6 lack literal witnesses; no runtime board/epoch reader or client exists; body/scope and review provenance remain stale. | RP-039–RP-043: author reconciliation, AC5 fix, witness-only remediation, accepted reader/surface successor, plan/thread reconciliation, then full-range cross-party review. |
| Prestige & Exits | implementing | Atomic Exit machinery and current first-hour payoff are cold-green. AC8 is proven under the governed availability policy, but AC2–AC6 have narrower witnesses than their literals; Advisor has no player control; body/docs are stale; and the current archival review gate is absent. | RP-034–RP-038: author/owner reconciliation, witness-only remediation, plan/docs reconciliation, then a tracked full-range designated verdict. |
| API Foundation | accepted / implementing | Registry/generation/cursor/middleware mechanics exist, but only 10 of 21 live routes are governed; public paths/readers/evidence/runtime are absent; generated client/header/query contracts are incomplete; AC4 contradicts C9; docs/review union are stale. | RP-060–RP-065: ruling-author AC4 reconciliation, complete A4–A8 authority/migration/public composition, literal proofs, exact cross-party union. |
| Minigame Platform Foundation | accepted / implementing | Durable sessions, tenant/policy boundaries, atomic payout/replay, production Pitch, and authenticated API are cold-green. AC2 is false because offline quality never decays observably or reaches an automated-output consumer; AC1 lacks async composition; AC3 lacks a bot match; AC6 falsely claims a duel tenant. RFC/schema, plan, docs, successor status, and review union are stale. | RP-016/RP-044–RP-047: ruling-author body/schema reconciliation, accepted decay fix, async scope ruling, combat-dependent bot/duel work, plan/docs reconciliation, then exact cross-RFC designated review. |
| Combat Shared Data & Arithmetic | implementing | AC2–AC4 are proven in both runtimes. AC1/AC5 are blocked by missing exact effect/table design. AC6's client gate is proven but contradicts C3's all-path wording because Go Combat divides natively. Plan/ledger/review hashes are stale and no current designated union exists. | RP-072–RP-074: ruling-author schemas/tables/slash scope, remaining parent implementation, current-history cross-party review before children. |
| Combat Duel Engine | draft | Acceptance review done; owner rulings and implementation absent. | Parent completion, owner rulings, acceptance. |
| Combat Lane Engine | draft | No implementing plan. | Parent contracts and owner acceptance. |
| Combat Bots & Integration | draft | No implementing plan. | Duel/lane engines plus account/minigame integration. |
| Game UI Screens | accepted / implementing | Phase-A/v2 screens exist; AC2/AC3 are proven. Accepted v3 Gate/Wind Down/next-run controls are absent; AC1/U2/header/docs contradict C7/C25–C28; AC4's oracle passes with all three outcomes removed; AC5 lacks throttle/frame/manual evidence. | RP-008/RP-026/RP-066/RP-067: ruling-author reconciliation, narrowed control proof, AC4/AC5 evidence repair, full-range designated review. |
| Commons Onboarding & Governance | draft | Backend Commons foundation exists; player-facing half absent. | Owner accepts reconciled draft before implementation. |
| World Layer Foundation | draft | No implementation. | Resolve release/content ordering and upstream Commons/production dependencies. |
| Feed & Dispatch Foundation | draft | No player feed implementation. | Transport/API/achievement dependencies and owner acceptance. |
| Events Engine Layer 1 | draft | No evaluator/content workflow. | Meter/achievement semantics and owner acceptance. |
| Minigame & Recovery API + Surface | accepted / implementing | AC1's composed backend is proven. AC2/AC3 are partial, AC4's four-operation enumeration is false for the Recovery family, and AC5 has no exact surface contract/client/components. Summary/MA1/MA2/MA3/header/plan contradict rulings/current dependencies; final union is unassembled. | RP-068–RP-071: ruling-author body/spec repair, literal backend witnesses, API client, both visible workflows, exact current-history cross-party review. |
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
