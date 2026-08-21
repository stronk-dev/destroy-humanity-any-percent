# Active RFC lifecycle audit at `190a4fa`

The status below is evidence-derived. It does not change an RFC's formal status by itself.

| RFC | File status | Reality at HEAD | Required next gate |
|---|---|---|---|
| RFC process | accepted | Active governance, not product work. | Keep. |
| Balance Harness Dispatch Integrity | withdrawn | Premise refuted; retained for history. | Decide whether withdrawn RFCs belong in active index or a non-implemented archive lane. |
| CI Baseline | implementing | D-014's six fast push/PR jobs passed hosted at `aa27705`; AC2–AC4 are proven, while AC1 retains its exact local aggregate and AC5 retains maintenance execution. Ten topology negatives pass. | Complete those two archival checks and the exact cross-party review union; none blocks unrelated product work. |
| Account & Session Bootstrap | implementing | Backend AC1–AC7 are cold-proven; Q-001's literal AC2/AC3/AC5–AC7 repairs are designated-approved at `34d04a5`. The Game UI still never refreshes its 15-minute token; account/recovery/export/delete/fallback surfaces, retention ruling, body and archival union remain open. | RP-004–RP-007/RP-024/RP-048/RP-050–RP-051: owner/body reconciliation, accepted player-session/rights/fallback successor, retention decision and exact full-history review. |
| WebSocket Transport & Fan-out | implementing | Server transport and production browser position/cursor recovery, reconnect/full-sync, typed closes, drain delay, owner-reconciled ten-second convergence and Guild denial are designated-approved at `249719c`. | RP-055: separately reconcile T6/plan history, Account token-rotation dependency and complete archival range union. |
| Leaderboards & Balance Epochs | implementing | Epoch/replay/projection mechanics are cold-green. Body/docs now match the six-verdict/five-category backend and route absent integration to an unaccepted draft successor. AC5 remains false; AC1/AC3/AC6 lack literal witnesses; no runtime reader/client exists; plan evidence and review provenance remain stale. | RP-039–RP-041/RP-043: accept/narrow reader successor, then AC5 fix, witnesses, plan/thread reconciliation and full-range cross-party review. |
| Prestige & Exits | implementing | Atomic Exit machinery and current first-hour payoff are cold-green. Body/docs match shipped behavior and D-012. AC2–AC6 now have cold-green, severing-proven literal witnesses; AC8 remains proven under the governed availability policy. The archival review gate is absent. | RP-038 plus AC8 plan-evidence reconciliation: obtain a tracked exact full-range designated verdict before archival. |
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
| Minigame & Recovery API + Surface | accepted / implementing | AC1's composed backend and Q-002's AC2–AC4 exact errors, stateful limiter nonmutation and all-eight-operation privacy contracts are designated-approved at `bfd9b65`. AC5 still has no exact surface contract/client/components; Summary/MA1–MA3/header/plan remain stale and the final union is unassembled. | RP-071: ruling-author body/exact-surface repair, API client, both visible workflows, then exact current-history cross-party review. |
| Permits & T3 Gate | promoted | Candidate and atomic epoch-6 activation ranges are designated-reviewed; AC1/AC2 are cold-proven. AC3 lacks the exact two-resource Go/TS replay row; P2/P4/AC4 contradict their rulings; canonical docs are pre-mint. | RP-027–RP-029: author body reconciliation, missing witness, docs correction, then a new complete designated range before archival. |
| First Content Epoch | implementing | Epoch 6 mint and repairs were designated-approved; AC1/AC3 are proven. AC2's fixed witness drifted to epoch 8, AC4 lacks exact range citations, AC5 body contradicts its range-head ruling, and epoch-7 harness scope was grafted into this bounded mint RFC. | RP-030–RP-033: reconcile authority/successor scope, restore exact witness and provenance, then transactional closeout. |
| Deployment Foundation | draft | Draft's central “local-only / THE PUSH” premise is false. Production packaging is still absent. | Owner rules release floor and repository disposition; author rewrites the body before acceptance. |

## Planning-record defects

- Multiple plans end in July/August blockers already resolved by later RFCs and commits. A stale
  checkbox is not evidence that work remains, and later code is not permission to flip it without
  an acceptance/range audit.
- RP-102 moved the completed `run-genesis-archival-remediation`, `harness-dispatch-cardinality`,
  `archived-four-review`, and `production-review-round2` threads under `planning/archive/` on
  2026-08-21. `ci-remediation-2026-08-10` remains live because it was outside that bounded route.
- The coverage map's generated-from-slices claim is not backed by a checked-in generator, and its
  six source slices have not been revalidated since 2026-08-05.
