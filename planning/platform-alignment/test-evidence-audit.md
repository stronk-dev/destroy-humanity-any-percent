# Test and acceptance evidence audit

Coordinate: product tree `190a4fa`; 2026-08-20. State: **in progress**.

This log records tests actually executed in the audit and whether they witness an entire RFC
criterion. A green package is not automatically a green criterion; separated halves remain
partial until one integrated fixture proves the required event sequence.

## Batch A — Account AC1–AC7 and transport boundary

### Commands read through final exit

| Command | Result | Population actually executed |
|---|---|---|
| `make test-save-integration SAVE_TEST_PACKAGES=./account` | exit 0, Go package 1.279s | Every `Integration`-named Account test against Compose-owned Postgres, cold `-count=1`. |
| `make test-go GO_PACKAGES=./account GO_TEST_FLAGS='-count=1'` | exit 0 | Account unit/schema/token/bootstrap tests, cold. Postgres tests skip in this arm by design. |
| `make test-save-integration SAVE_TEST_PACKAGES=./leaderboard` | exit 0, Go package 0.389s | Current imported-founder exclusion and leaderboard epoch/projector integration. |
| `make test-go GO_PACKAGES=./transport GO_TEST_FLAGS='-count=1'` | exit 0, Go package 4.217s | Actual WebSocket recovery/drop-stale tests, 5k soak, queue/authz/unit transport population. |
| `make test-go GO_PACKAGES=./gameserver GO_TEST_FLAGS='-count=1'` | exit 0 | Drain/readiness/job unit population; Integration-named tests skip without Postgres. |
| `make test-save-integration SAVE_TEST_PACKAGES=./gameserver` | exit 0, Go package 2.161s | Integration-named composed gameserver tests against Postgres. |

The first attempted focused Account command used a parenthesized Go `-run` regex through
`SAVE_TEST_FLAGS`. The Make target expands the variable unquoted, so `/bin/sh` rejected `(` before
Docker or Go ran. The attempt is an invalid measurement, not a test failure. The safe package-wide
rerun above reached its objective. RP-025 owns selector safety.

### Account criterion findings

- **AC1:** the real Account test performs create → session → production intent → receipt against
  Postgres. Registry/schema tests mechanically cover exact wire. Current witnesses are green;
  negative exact-key discrimination and review-range union still need explicit audit.
- **AC2 partial:** refresh reuse and concurrent replay revoke the complete family in real
  Postgres. Transport authenticates at connect and re-authenticates from `OnAlive`, but no test
  composes a real refresh-family revocation with an already-connected socket and then proves the
  required subscribe/disconnect behavior. Separate green tests do not prove the clause.
- **AC3:** New Founder archives the prior stream, creates a fresh run genesis, and leaves the old
  stream loadable in the Account integration fixture. Current witness green; mutation/range audit
  pending.
- **AC4:** exact five-claim JWT round-trip and an extra-claim rejection ran cold in the unit arm.
- **AC5:** import/restore and the imported flag run in Account Postgres; the separate current
  Leaderboard Postgres fixture rejects imported projection. Both halves are green, with exact
  range/discrimination review pending.
- **AC6:** deletion removes accounts, email/session rows and account links while retaining archived
  anonymized save/founder streams in Postgres. This proves the RFC's backend criterion, not the
  absent player deletion UI/disclosure tracked by RP-005.
- **AC7:** the unauthenticated limiter returns the typed `rate_limited` response in real Postgres.

### WebSocket criterion findings

- **AC1, AC2, AC3, AC6:** current cold package witnesses completed, including the 5k simulated
  connection population, actual private-receipt/world recovery, stalled-world coalescing, and
  subscription authorization. Oracle/mutation and range review remain.
- **AC4 partial:** queue overflow/typed close and client resynchronization exist in separate test
  layers; no single fixture proves overflow → close → reconnect → committed revision.
- **AC5 partial:** drain ordering/flush/close and reconnect recovery are separately tested. No
  integrated drain → reconnect fixture proves zero receipt loss across that exact boundary.

## Pending evidence work

- Demonstrate a discriminating failure or relevant mutation for every remaining row promoted
  beyond mechanical presence.
- Reconcile the remaining 19 mechanical/cold-witness rows, beginning with CI AC2/AC4/AC5.
- Construct exact current-history review unions before any active RFC archive; the lifecycle
  audits name the missing Account and Transport ranges rather than treating old green prose as a
  verdict.
- Record skips, exclusions, guards, caches, and architecture differences as visible fields rather
  than silently narrowing the population.

## Batch B — Game UI browser and boundary evidence

| Command | Result | Interpretation |
|---|---|---|
| `make test-browser` | exit 0; 123 files, 20,007 passed, 3 skipped; isolated performance arm 1 passed/10 skipped | Complete declared browser suite and governed performance selector ran cold. Skips remain visible. |
| `make test-game-ui-composed` | exit 0 | Real Postgres + gameserver + Chromium proved only bootstrap, v2 snapshot, stored credentials, and WebSocket presence; the script's own PASS label states that bounded scope. |
| `make typecheck verify-client-boundary` | exit 0; Svelte 0 errors/0 warnings | AC2's compile/boundary mechanics are green. |

The composed script stops immediately after Desk and presence. It does not execute the first-hour
script, offer/Exit, run-end rendering, or run-2 state, so it cannot witness Game UI AC1.

Browser AC3's run-end component receives only the decoded `run_ended` payload and passed. The
performance arm passed. AC4 remains partial: the cap explanation primitive is asserted in the UI
Foundation browser fixture, while the Game UI test simultaneously forces drain and resync and then
asserts only axe and absence of mechanical IDs. It has no assertion for restart timing/copy,
resync copy/button, refresh invocation, failure behavior, or return to a synchronized surface.
Filed RP-026.

## Batch C — Transport production-consumer and lifecycle replay

| Command | Result | Interpretation |
|---|---|---|
| `make test-go GO_PACKAGES='./transport ./gameserver' GO_TEST_FLAGS='-count=1'` | exit 0; Transport 4.104 s, gameserver 0.486 s | Cold actual-socket, 5k soak, queue/history/authz, and server drain populations. |
| `make test-client` | exit 0; 39 files passed/2 skipped, 6,655 tests passed/15 skipped | Complete Node client population, including isolated cursor and Game UI runtime tests. |
| root `pnpm --filter ...` attempt | did not start; no root package manifest | Invalid command, excluded from evidence; the root Make target supplied the client result. |

Fixture and consumer inspection promotes only Transport AC1. The 5,000-socket oracle validates
each subscriber's trace across ten world ticks, requires a strictly increasing subsequence ending
at the terminal revision, rejects public receipt-shaped publication, and has a
malformed/private-push negative.

AC2–AC6 remain partial. Actual server recovery is real, but `game-ui/runtime.ts` sends unpositioned
subscribe commands, discards offsets, never reconnects, and does not import the per-scope cursor.
Typed overflow and bounded drain are also server-only halves: the close listener discards the code,
and `resume_after_ms` has no scheduling consumer. AC3's connected stall is under one second and
allows in-flight plus newest; AC6 has no non-member Guild negative. RP-052–RP-055 and
`websocket-transport-lifecycle-audit.md` carry the exact remediation and review-range requirements.
