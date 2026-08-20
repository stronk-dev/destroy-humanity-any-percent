# Capability reality audit

Coordinate: `190a4fa`, 2026-08-20. “Proof” means a named executable witness, not a prose claim.

| Outcome | Intent | Producer / primitive | Consumer | Real content / workflow | Executable witness | Verdict |
|---|---|---|---|---|---|---|
| Start and play the live game | `design/11`, Game UI RFC | Bootstrap, snapshot v2, production intents, world channel | `GameUIApp.svelte` via `game-ui/runtime.ts` | Vision Slide -> Desk on epoch-7/8 catalogs | `make test-game-ui-composed`; current-head hosted job passed | **Proven to Desk only** |
| Recover realtime delivery after disconnect | Transport D4/T4, AC2/AC4/AC5 | Centrifuge positions/history, typed closes, full-state endpoint, per-scope cursor | Game UI opens one unpositioned socket; cursor is test-only | Initial live subscription only | Server protocol-driver recovery/overflow/drain tests; no browser recovery witness | **Proven server; consumer absent** |
| Complete the first hour in-browser | T0–T1 RFC + Game UI AC1 | Live curriculum, gate/Exit events, run-end payload | Offer Sheet and Run End components | Nine generators, ten upgrades, three first-ending branches | Server/Postgres proof exists; no full browser script | **Mechanical fragment** |
| Play with keyboard/assistive tech | Game UI accessibility clause | Semantic Svelte controls, focus styles, reduced-motion tokens | Browser DOM | Five Phase-A surfaces | Axe on all surfaces; Enter begins attempt | **Mechanical fragment** |
| Recover an anonymous account | Account RFC | One-time recovery code, session endpoint | Credential stored in browser localStorage | Automatic account bootstrap | Backend integration tests only | **No player workflow** |
| Export/delete own data | Account/privacy intent | Delete endpoint exists; no export endpoint | No settings controls | Settings displays one paragraph | Delete backend test only | **Mechanical deletion; export absent** |
| Continue locally during an outage and import later | `design/11 §1b`; older Account D4/`design/06` need reconciliation | Import endpoint exists | No local save runtime | Production startup creates the ruled anonymous server account | No fallback witness | **Claimed fallback only** |
| Receive offline progress | Design law 7 | Production accrual and session-boundary application | Snapshot/UI offline status | T0–T1 economy | Composed 48 h offline catch-up regression | **Proven integration** |
| Play The Pitch | `design/03`, minigame RFCs | Pitch tenant + minigame session/resolve service | API only | Minted Pitch content | Real-socket create -> command -> terminal resolve | **Backend proven; surface absent** |
| Recover Soul through a cozy activity | Soul RFCs | Recovery session coordinator | Scheduler/API only | Three minted activities | Authenticated composed lifecycle | **Backend proven; surface absent** |
| Care for a pet | `design/04` | Pet state, decay, actions, replay | Catalog/TS helpers only | Minted policy, no acquired pet workflow | Unit/cross-runtime fixtures | **Mechanical fragment** |
| Fight a pet/lane battle | Combat RFC family | Exact arithmetic and RNG | None | Fixture arithmetic only | Golden/unit tests | **Mechanical fragment** |
| Join and experience an MMO community | `design/05` | Commons/guild/presence/clearing | No game UI surface | Structural server data only | Server/Postgres integration tests | **Mechanical fragment** |
| View or submit leaderboards | `design/05 §6` | Run logs, replay verdicts, board projection | Public API/UI incomplete | Live categories exist | Projection tests | **Mechanical fragment** |
| Experience events/world narrative | `design/09`, `design/13` | Meters/routes can feed future evaluators | None | No Layer-1/2/3 event content | None | **Absent** |
| Reach the designed ending | `design/01`, `design/11` | Early Exit machinery only | Early Run End component | T0–T1 only | First-company ending proof | **Absent as designed product** |
| Trust hosted CI at HEAD | CI RFC | Six blocking Actions jobs plus scheduled numeric maintenance | Branch/release gate | Current `main` | Runs 32009994004, 32096019304, 32212696707, 32328790752, 32404232364 | **Contradicted: harness cancellation** |
| Self-host/recover service | `design/06`, sunset research | Gameserver binary; migrations | No production packaging/runbook | Test Compose only | No clean-host rehearsal | **Absent** |

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

## Genuine proofs worth preserving

- Numeric parity, canonical string wire values, save migration/replay, lazy production, offline
  catch-up, and the T0–T1 server path are unusually well evidenced.
- Current-head hosted server, client, schema, browser, and composed-browser jobs each pass.
- The real composed browser test proves bootstrap, Postgres, snapshot v2, and world WebSocket
  presence rather than a mocked transport.
- Backend account deletion and refresh-replay revocation have discriminating real-Postgres tests.

These proofs narrow the remediation program: do not rebuild the kernel. Bind the missing consumers,
release contracts, content, and lifecycle records to it.
