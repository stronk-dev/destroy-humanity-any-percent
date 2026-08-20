# Capability reality audit

Coordinate: `190a4fa`, 2026-08-20. “Proof” means a named executable witness, not a prose claim.

| Outcome | Intent | Producer / primitive | Consumer | Real content / workflow | Executable witness | Verdict |
|---|---|---|---|---|---|---|
| Start and play the live game | `design/11`, Game UI RFC | Bootstrap, snapshot v2, production intents, world channel | `GameUIApp.svelte` via `game-ui/runtime.ts` | Vision Slide -> Desk on epoch-7/8 catalogs | `make test-game-ui-composed`; current-head hosted job passed | **Proven to Desk only** |
| Recover realtime delivery after disconnect | Transport D4/T4, AC2/AC4/AC5 | Centrifuge positions/history, typed closes, full-state endpoint, per-scope cursor | Game UI opens one unpositioned socket; cursor is test-only | Initial live subscription only | Server protocol-driver recovery/overflow/drain tests; no browser recovery witness | **Proven server; consumer absent** |
| Complete the accepted Game UI transition path | T0–T1 RFC + Game UI C25–C28/AC1 | Live curriculum, gate/Exit events, run-end payload; v3 transition preview absent | Offer Sheet and Run End render; Gate/Wind Down/continue controls absent | Nine generators, ten upgrades, three first-ending branches | Server/Postgres first hour proven; browser stops at Desk and accepted control proof is absent | **Mechanical fragment; body-blocked** |
| Play with keyboard/assistive tech | Game UI accessibility clause | Semantic Svelte controls, focus styles, reduced-motion tokens | Browser DOM | Five Phase-A surfaces | Axe on all surfaces; Enter begins attempt | **Mechanical fragment** |
| Recover an anonymous account | Account RFC | One-time recovery code, session/refresh endpoints | Credential hidden with tokens in one localStorage document; refresh never consumed | Automatic account bootstrap only; missing/malformed storage creates replacement or generic offline state | Backend integration tests only | **No player recovery workflow** |
| Export/delete own data | Account/privacy intent | Delete endpoint exists; no export endpoint | No settings controls; destructive-confirmation helper is test-only | Settings displays status and one paragraph | Delete backend test only | **Mechanical deletion; export absent** |
| Continue locally during an outage and import later | `design/11 §1b`; older Account D4/`design/06` need reconciliation | Import endpoint exists | No local save runtime | Production startup creates the ruled anonymous server account | No fallback witness | **Claimed fallback only** |
| Receive offline progress | Design law 7 | Production accrual and session-boundary application | Snapshot/UI offline status | T0–T1 economy | Composed 48 h offline catch-up regression | **Proven integration** |
| Play The Pitch | `design/03`, minigame RFCs | Pitch tenant + minigame session/resolve service | Generated DTO metadata only; no HTTP client/component/table | Minted Pitch content | Cold real-Postgres/HTTP create→reconnect→commands→terminal/retries; resolver-sever negative | **Backend proven; surface absent/spec-blocked** |
| Recover Soul through a cozy activity | Soul RFCs | Recovery session coordinator | Framework scheduler only; no constructor/surface/toy | Three minted activities | Cold authenticated composed lifecycle including rotation/heartbeats/resolve/watchdog; no UI | **Backend proven; surface absent/spec-blocked** |
| Care for a pet | `design/04` | Pet state, decay, actions, replay | Catalog/TS helpers only | Minted policy, no acquired pet workflow | Unit/cross-runtime fixtures | **Mechanical fragment** |
| Fight a pet/lane battle | Combat RFC family | Proven dual-runtime arithmetic/RNG/chart kernel | No engine or surface | Shared arithmetic vectors only; no combat catalog | Cold Go/client vectors with independent stage/label/chart mutations; client slash gate | **Kernel proven; gameplay absent/spec-blocked** |
| Join and experience an MMO community | `design/05` | Commons/guild/presence/clearing | No game UI surface | Structural server data only | Server/Postgres integration tests | **Mechanical fragment** |
| View or submit leaderboards | `design/05 §6` | Run logs, replay verdicts, board projection | Public API/UI incomplete | Live categories exist | Projection tests | **Mechanical fragment** |
| Build against the documented API | API Foundation | Schema/operation DSL, generator, compatibility checker | 10 generated private operations; 12 hand-mounted routes; no public client/router | No public catalog/board/evidence workflow | Registered drift mutation fails; unregistered handler and raw-fetch-lint probes pass incorrectly | **Mechanical fragment; authority incomplete** |
| Experience events/world narrative | `design/09`, `design/13` | Meters/routes can feed future evaluators | None | No Layer-1/2/3 event content | None | **Absent** |
| Reach the designed ending | `design/01`, `design/11` | Early Exit machinery only | Early Run End component | T0–T1 only | First-company ending proof | **Absent as designed product** |
| Trust hosted CI at HEAD | CI RFC | Six blocking Actions jobs plus scheduled numeric maintenance | Branch/release gate | Current `main` | Runs 32009994004, 32096019304, 32212696707, 32328790752, 32404232364 | **Contradicted: harness cancellation** |
| Self-host/recover service | `design/06`, sunset research, Deployment draft | Buildable disk-dependent gameserver; migrations/readiness/drain; current-key refusal | No static client, production package, deploy owner, or recovery runbook | Test Compose only; public origin remains localhost and proxy/previous-key inputs are uncomposed | Binary build + missing-secret refusal only; no clean-host/backup/restore/rollback rehearsal | **Absent; useful primitives only** |

## False or stale claims found

1. The pre-audit `planning/CURRENT-STATE.md` said the branch was 39 commits ahead of `origin/main`;
   it is 0 ahead and 0 behind. The current brief is reconciled.
2. The pre-audit `README.md` said the shell had no live transport adapter and the game was not
   playable end to end; the real bootstrap/snapshot/WebSocket path now exists, while the stronger
   full-first-hour claim remains unproven. The README is reconciled in this batch.
3. `design/research/README.md` says the repository is private; GitHub reports it is public, and the
   CI log records the owner's public-repository choice.
4. `rfc/deployment-foundation.md` repeatedly says the repository is local-only and “the push” has
   not happened; `origin/main` equals `HEAD` and current Actions runs exist.
5. `planning/scaffolding-and-ci/plan.md` says hosted acceptance is merely unobserved. It has now
   been observed failing to reach a verdict at current HEAD.
6. The pre-audit `rfc/README.md` pointed to a 2026-07-29 batch as the “current handoff.” The index
   now routes to the platform-alignment execution queue.
7. The 2026-08-05 coverage map still labels several now-archived systems as implementing/draft and
   explicitly says it must not be quoted without revalidation.
8. `design/BACKLOG.md` still marks pacing and compliance research “in flight” while the research
   matrix calls both covered.
9. The mandated design backlog, research matrix/dossiers, and coverage map are gitignored, so they
   are not repository-shared memory for a fresh clone. `backlog.md` is the tracked interim ledger.
10. Transport T3/T4 and canonical docs call the per-scope cursor/recovery path client authority,
    but the production Game UI neither imports that cursor nor stores positions/reconnects.
11. CI's body/plan still describe a four-job, under-five-minute baseline awaiting its first hosted
    run; current topology has six blocking jobs, repeated hosted cancellations, and cache/nightly
    behavior that contradicts canonical claims.
12. API A5/C1 claim one operation authority, but only 10 of 21 live v1 routes are registered; the
    generated document has zero public routes and the accepted C9 client-only rule is neither
    implemented nor enforced.
13. Game UI's RFC and canonical doc still require the full browser first-hour even though C25/C28
    rejected that obligation; the accepted v3 controls are absent, AC4's three outcomes can all be
    removed without a red test, and AC5's throttle/drop-frame profile is unexecuted.
14. Minigame API's body still claims no persistence, a catalog command budget, `session_expired`,
    and stale dependencies despite opposite rulings; AC4 omits Recovery enumeration and AC5 has no
    exact surface contract, HTTP client, components, or player workflow.
15. Combat's ledger says chart/lint are unimplemented although the client gate and both-runtime
    chart properties exist; the RFC still lacks effect/table contracts, and its all-path native
    division ban is enforced only for client modules while Go Combat divides natively.
16. Deployment says catalogs are embedded and origin/proxy/rotating secrets are host-injected;
    composition reads the repository tree, pins WebSocket origin to localhost, hardcodes Account
    proxy depth to zero, and exposes no previous-key configuration.
17. Deployment and the RFC index still center a future “THE PUSH,” even though the public remote
    and hosted Actions prove that external phase transition already occurred.
18. Account D1 says the recovery code is shown once; production hides it in localStorage. Missing
    or malformed credentials have no recovery branch, and the refresh token is never used.
19. Settings says offline progress is parked locally even though the client has no local gameplay
    save, durable intent queue, or reconnect/import flush path.

## Genuine proofs worth preserving

- Numeric parity, canonical string wire values, save migration/replay, lazy production, offline
  catch-up, and the T0–T1 server path are unusually well evidenced.
- Current-head hosted server, client, schema, browser, and composed-browser jobs each pass.
- The real composed browser test proves bootstrap, Postgres, snapshot v2, and world WebSocket
  presence rather than a mocked transport.
- Backend account deletion and refresh-replay revocation have discriminating real-Postgres tests.

These proofs narrow the remediation program: do not rebuild the kernel. Bind the missing consumers,
release contracts, content, and lifecycle records to it.
