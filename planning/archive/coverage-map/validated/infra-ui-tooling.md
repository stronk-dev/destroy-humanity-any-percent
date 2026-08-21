# Coverage Reconstruction — Domain: infra-ui-tooling

> **FROZEN HISTORICAL SNAPSHOT — NONCANONICAL.** Reconstructed on 2026-08-05; retained as
> evidence, not current status or execution authority. See `planning/platform-alignment/`.

Reconstructed independently from actual files on 2026-08-05. R = research dossier,
D = design section, F = RFC file + real status, I = implementation (archive + docs/).
Stage codes: R < D < F(draft) < F(accepted) < F(implementing) < I(implemented-unarchived) < I(archived).

## Summary table

| System | R (research) | D (design) | F (RFC + real status) | I (docs / archive) | Furthest | Tags |
|---|---|---|---|---|---|---|
| WebSocket transport & fan-out | `tech-stack.md §1` **specifies** (centrifuge-embedded, aggregate-then-broadcast, backpressure); `cicd-deploy.md §5` (drain handshake); `cattery-reusables.md` (rejected broadcast set). Stack-decision, not gameplay. | `design/06` stack table rows (fan-out, realtime) | `rfc/websocket-transport-and-fanout.md` — **implementing** | `docs/transport.md` (v2 envelope, present-tense = landed); `server/transport/`; planning F2/F3 remediation done, independent review pending | I (implemented, unarchived) | — |
| Account / session bootstrap | `compliance-2026-refresh.md` **specifies** data-minimisation posture; `tech-stack.md` mentions JWT accounts | `design/06 §backend` (JWT, anonymous-first) | `rfc/account-and-session-bootstrap.md` — **implementing** | `docs/accounts-and-sessions.md`; `server/account/`; round-3 LOW remediation reviewed/approved; session/token GC deferred to gameserver | I (implemented, unarchived) | — |
| API foundation | `neopets-social-history.md §5`, `wc3-custom-ecosystem.md §3`, `speedrun-governance.md` — **mention** (community-tooling motivation, not API mechanics) | `design/06` (chi REST), `design/05` (published formulas) | `rfc/api-foundation.md` — file: **accepted (C1–C17 ruled; implementing)** / README: **implementing** | `docs/api-foundation.md` (schema/op registry + keyset cursors landed); `server/publicapi/`; C18–C20 + endpoints/generation unbuilt | I (partial — foundation landed, surfaces open) | — |
| CI / scaffolding | `cicd-deploy.md` **primary/specifies** gate stack; `pacing-science.md`, `adaptive-balancing.md`, `compliance-2026-refresh.md` secondary | `design/07` Phase 0, `design/12 §7` | `rfc/scaffolding-and-ci.md` — **implementing** | `docs/ci.md` (`.github/workflows/ci.yml` exists); local green; **hosted <5min run unverified (never pushed)** | I (partial — local only) | STATUS-OFF (soft) |
| Gameserver composition | `tech-stack.md` (goroutine actors), `cattery-reusables.md` | `design/06` stack/architecture | `rfc/archive/gameserver-composition.md` — **implemented** | `docs/gameserver.md`, `docs/guilds.md`; `server/cmd/gameserver/`, `server/gameserver/` | I (archived) | — |
| Copy pipeline | `run-narrative-ux.md` **specifies** copy-system feed | `design/12` (content pipeline), `design/11 §6` | `rfc/archive/copy-pipeline-foundation.md` — **implemented** | `docs/copy-pipeline.md`; `server/copykeys/`, `client/src/copy/`, `cmd/verify-copy-sites` | I (archived) | — |
| Client shell & sim loop | `tech-stack.md §2`, `browser-rendering.md` **specify** (DOM-first, 20 Hz sim) | `design/06` (Svelte 5 runes, sim loop) | `rfc/archive/client-shell-and-sim-loop.md` — **implemented** | `docs/client-shell.md`; `client/src/`; D2/D3 reconciliation rulings owned | I (archived) | — |
| UI foundation | `tech-stack.md §2`, `browser-rendering.md`, `flash-era-arcade.md` (era texture) | `design/06`, `design/08` (era presentation), `design/11`, `design/03 §9` | `rfc/ui-foundation.md` — file: **accepted architecture (C1–C8 ruled; C9–C11 block impl)** / README consistent | none yet — **no implementation**; C9 (token matrix), C10 (notation pin), C11 (a11y engine) block impl; ruled-text reconciliation reviewed 2026-08-04 (Darwin) | F (accepted, impl-blocked) | — |
| Game-UI screens | `tech-stack.md §2`, `run-narrative-ux.md` | `design/11`, `design/08 §speedrun`, `design/06` | `rfc/game-ui-screens.md` — **draft** (Created 2026-08-03); no planning dir | none | F (draft) | STATUS-OFF (README index) |
| Deployment foundation | `cicd-deploy.md` **specifies** drain sequence §5, hosted-GHA topology | `design/06 §deploy` (single binary, Caddy, compose, PG16), `design/07` | `rfc/deployment-foundation.md` — **draft** (Created 2026-08-03); no planning dir | none — the push is Marco-gated | F (draft) | STATUS-OFF (README index) |
| Balance harness | `balance-enforcement.md`, `pacing-science.md`, `adaptive-balancing.md` | `design/02 §11` | `rfc/archive/balance-harness-foundation.md` + hardening: `balance-baseline-change-guard.md`, `phase0-pacing-observation-coverage.md`, `post-review-integrity-remediation.md` — all **implemented** | `docs/balance-harness.md`; `server/harness/`, `cmd/balance-harness` | I (archived) | — |
| Relevance harness | `balance-enforcement.md` **specifies** academic anchors (Demaine 2018, Jaffe 2012, Riot ANY/ALL); `tier-relevance.md` | `design/02 §11b` (tier-relevance doctrine) | `rfc/relevance-harness.md` — **draft** (Created 2026-08-01); `planning/relevance-harness/log.md` records a required-scope-split blocker (Purchasable Content owner prerequisite) | none | F (draft, blocked) | STATUS-OFF (README index) |
| Run genesis & replay | `speedrun-governance.md` (verification, timing, server-authority obligations) | `design/08 §speedrun`, `design/05` | `rfc/archive/run-genesis-and-replay.md` — **implemented** | `docs/leaderboards-and-epochs.md` (shared with Leaderboards); `server/runidentity/`, `replaycatalog/`, `replayverify/` | I (archived) | — |
| Gate predicates & routes | `speedrun-governance.md §3.5`, `tile-placement.md` (draft-picks as Route Knowledge) | `design/08 §6`, `design/02 §3`, `design/05 §6`, `design/11 §4`, `design/10 §3b` | `rfc/archive/gate-predicates-and-routes.md` + hardening: `route-registry-event-order-convergence.md`, `route-temporal-validity.md` — all **implemented** | `docs/routes.md`; `server/routes/`, `routeprojection/` | I (archived) | — |

Also present in domain but tracked elsewhere: `balance-harness-dispatch-integrity.md` — **withdrawn (premise refuted 2026-08-03)**; parent = Balance Harness Foundation. Listed in README Active table as "withdrawn — premise refuted" (self-consistent; placement in the Active table is cosmetic).

## Systems accounted: 14 (+1 withdrawn amendment RFC)

Furthest-stage tally:
- **I — archived (implemented + docs canonical):** 6 — gameserver composition, copy pipeline, client shell & sim loop, balance harness, run genesis & replay, gate predicates & routes.
- **I — implemented but unarchived (RFC still `implementing`):** 2 — WebSocket transport & fan-out, account/session bootstrap.
- **I — partial (foundation landed, surfaces/verification open):** 2 — API foundation (C18–C20 + endpoints open), CI (local green, hosted run unverified).
- **F — accepted, implementation-blocked:** 1 — UI foundation (C9–C11).
- **F — draft:** 3 — game-UI screens, deployment foundation, relevance harness.

## Findings (adversarial, with evidence)

**STATUS-OFF #1 — README claims game-UI screens & deployment are "not yet drafted", but draft RFCs exist.**
`rfc/README.md:69`: "Remaining Phase-0 contracts (not yet drafted): … game-UI screens · deployment …". Yet `rfc/game-ui-screens.md` and `rfc/deployment-foundation.md` both exist with `Created: 2026-08-03` and `Status: draft`. The index text is stale relative to the drafted files.

**STATUS-OFF #2 — three domain RFCs are absent from the README Active table entirely.**
`rfc/game-ui-screens.md`, `rfc/deployment-foundation.md`, and `rfc/relevance-harness.md` are real draft RFC files (relevance-harness even has `planning/relevance-harness/log.md`), but none appear in the README Active table (lines 10–25). `grep -in relevance rfc/README.md` → no match. The Active index under-reports drafted work in this domain.

**STATUS-OFF #3 (soft) — CI acceptance gate open while status reads `implementing`.**
`planning/scaffolding-and-ci/log.md`: the workflow "cannot be executed on a hosted runner without pushing … hosted completion under five minutes is the only unverified acceptance gate." `docs/ci.md` describes the workflow as existing. Status `implementing` is technically correct, but the single blocking gate is verification-that-requires-a-push (owner-gated) — the same push `deployment-foundation.md` is built around. Worth a watch: this RFC cannot reach `implemented` without the deploy push.

**Observation — API foundation status wording differs between README and file (not a contradiction).**
File `rfc/api-foundation.md:2`: "accepted (C1–C17 ruled; implementing)"; README:23: "implementing — C1–C17 ruled". Both resolve to *implementing*; no reconciliation needed. Same pattern for UI foundation (file "accepted architecture (C1–C8 ruled; C9–C11 block implementation)" vs README "accepted architecture — C9–C11 implementation blockers") — consistent.

**Observation — no DRIFT between docs and RFCs for the unarchived `implementing` pair.**
`docs/transport.md` and `docs/accounts-and-sessions.md` describe present-tense landed behavior while their RFCs remain `implementing` and unarchived. This is expected for in-progress work (partial behavior documented as it lands), not doc/spec drift. Transport's independent diff review is still pending per its planning log; account bootstrap's is reviewed/approved with only LOW residuals (Argon2 param ceiling, trusted-proxy precondition doc, token GC deferred to gameserver).

**No UNBACKED, ORPHAN, or GAP tags in this domain.** Every system traces R→D→F; every named system has an RFC (no furthest=D-with-no-RFC gaps). No docs describe behavior lacking an RFC. Relevance harness has an internal blocker (planning log: needs a Purchasable Content Foundation owner RFC split before acceptance) — a scoping dependency, recorded, not an orphan.

## Cross-cutting note

The CI (#3) and deployment coupling is the load-bearing risk for the whole domain: `deployment-foundation.md` states every guard the archived infra built (epoch history, KV-1, balance harness, review ledger) is **ADVISORY** until the push lands, and CI cannot finish its acceptance gate without it. Two `implementing` RFCs and two drafts in this domain ultimately converge on that one owner-gated action.
