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
- Preserve the remaining five mechanically/cold-backed rows at their named exact-review routes;
  every active RFC lifecycle pass is now complete.
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
performance arm passed its bounded CI lane. Initial inspection left AC4 partial because the Game
UI fixture asserted only axe/mechanical-ID absence; Batch F's restored three-seam mutation later
demonstrated that all required outcomes can disappear while the complete browser population stays
green, so the final classification is contradicted. Filed RP-026.

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

## Batch D — CI hosted lifecycle and discrimination replay

| Evidence | Result | Interpretation |
|---|---|---|
| Public Actions run `32404232364` at remote `cb162a3` | server/client/browser/schema/game-ui-composed success; harness cancelled after 30m02s; workflow cancelled | Same product/workflow coordinate as the audit. Individual jobs are useful; AC1/AC3 are false. |
| Restored shared-vector mutation (`0^0` expectation `1`→`2`) | `make test-client` exit 2; one failed/6,654 passed/15 skipped | AC2's broken-vector arm discriminates in the actual client job population. |
| Restored malformed production catalog | `make verify-schema` exit 2 with required/additional-field diagnostics | AC4's command fails on a malformed production-path input, not only on an expected-invalid fixture. |
| Restored current populations | client 6,655 passed/15 skipped; schema full population exit 0 | Both temporary probes were removed; no probe byte remains in the worktree. |
| Scheduled Actions run `31562066740` | workflow conclusion failure when numeric-maintenance failed | The job is outside PR latency but not non-blocking in workflow-result semantics. |

Static population enumeration found 20 `*.schema.json` files and 20 explicit schema loads in the
verifier. AC2 and AC4 are proven. AC5 is contradicted: harness/numeric setup-go steps leave default
module-plus-build-output caching enabled, while D2/docs claim dependency-only; schedule/manual
runs the entire blocking workflow and does not tolerate numeric-maintenance failure. RP-056–RP-059
carry body/plan, workflow, review-union, and R-001 authority defects. Full reasoning is in
`ci-baseline-lifecycle-audit.md`.

## Batch E — API authority, generation, and client-boundary discrimination

| Evidence | Result | Interpretation |
|---|---|---|
| `make test-go GO_PACKAGES='./publicapi ./account ./gameserver' GO_TEST_FLAGS='-count=1'` | exit 0; publicapi/account/gameserver green | Current foundation, mounting, and bounded composition populations are cold-green. |
| `make api-check`; registered optional-field mutation | baseline exit 0; mutation exit 2 with exact OpenAPI/TS diffs; restored exit 0 | Generated drift discriminates for operations already inside the registry. |
| Removed registered nested response field | `make api-schema` exit 2, but diagnostic only `schema GameUISnapshot: invalid API schema` | Compatibility rejects removal but does not name the removed `GameUIResource.rate_per_second` required by AC2. |
| Added field to live hand-mounted Founder response | `make api-check` exit 0 | Eleven live v1 routes are outside generation/pins; AC1 does not cover the full handler surface. |
| Added direct `/api/v1` fetch to `game-ui/runtime.ts` | `make verify-client-boundary` exit 0 | The C9 lint scans only Game UI Svelte components and is blind to the production runtime. |
| `make typecheck` | exit 0, zero diagnostics | Current generated types compile; this does not create or prove a generated HTTP client. |

Generated OpenAPI has 10 private paths, zero public paths, and zero response-header descriptors.
Production code never constructs the public policy/cursor runtime. AC1/AC2 are partial, AC3/AC5
unmet, and AC4 contradicted under C9. Every temporary mutation was restored; RP-060–RP-065 and
`api-foundation-lifecycle-audit.md` carry the exact authority/body/composition/review defects.

## Batch F — Game UI lifecycle and acceptance discrimination

| Evidence | Result | Interpretation |
|---|---|---|
| `make test-browser` | exit 0; 123 files, 20,007 passed, 3 skipped; performance selector 1 passed/10 skipped | Current declared browser population is green, but its Game UI outcome coverage is not discriminating. |
| Direct `../transport` import in `GameUIApp.svelte` | `make verify-client-boundary` exit 2 naming the exact component/import | AC2's structural boundary fails on the prohibited dependency and is proven. |
| Optional `snapshot` input on `RunEndSurfaceProps` | `make typecheck` exit 2 because the negative `@ts-expect-error` became unused | AC3's byte-only component contract rejects the forbidden shell-state input and is proven with the rendered fixture. |
| Suppressed cap, drain, and resync output together | full `make test-browser` still exit 0 with 20,007 passed | AC4's claimed three browser tests do not observe any of the three required outcomes; the criterion is contradicted. |
| Performance scenario/source inspection | 1,200 update inputs and formatting/long-task bounds execute; CPU throttle and dropped-frame allowance are never applied | AC5 proves only a deterministic CI count/long-task arm; the ruled reference-device throttle/frame claim remains unproved. |
| Restored `make typecheck verify-client-boundary` and product diff | zero diagnostics; boundary green; no Game UI product diff | All temporary probes were removed before the planning checkpoint. |

The current composed browser proof still ends at bootstrap-to-Desk. Snapshot v3, Gate/Wind Down
eligibility and controls, and Run End→next-run continuation are absent. AC1 is also body-blocked:
its full first-hour obligation contradicts GU-C25–C28's later narrowed browser ruling. RP-008,
RP-026, RP-066, RP-067, and `game-ui-lifecycle-audit.md` carry the exact repair and review route.

## Batch G — Minigame API & Surface lifecycle discrimination

| Evidence | Result | Interpretation |
|---|---|---|
| Cold `./account ./minigame ./minigameapi ./gameserver -count=1` | all four packages green | Current unit/schema/coordinator populations execute without cache reuse. |
| Sequential Account and gameserver Compose/Postgres populations | both exit 0 | Mounted authenticated API, complete Pitch/Recovery lifecycle, limiter and repository paths are green against real Postgres. |
| Severed production tenant-content resolver | gameserver integration exit 2 at Pitch create with `500 internal_invariant/minigame_api` | AC1's composed producer/resolver/API seam is discriminating. |
| Bypassed Recovery progress limiter | Account integration exit 2 because seventh heartbeat returned 200 | Rejection is discriminating; the fixture is stateless and does not establish AC3's no-mutation clause. |
| Allowed `founder_id` in Recovery finish schema | complete cold Account unit population exit 0 | AC4's enumeration omits Recovery finish/progress families; four minigame operations are not “every” registered RFC operation. |
| Appended a byte to `minigame_revision` rejection | complete cold Account unit population exit 0 | AC2's alleged byte-match table uses substring containment; runtime is mechanically correct but the literal oracle is blind. |
| Restored `make api-check` and `make typecheck` | generation unchanged; zero diagnostics | Every temporary mutation was removed before this planning checkpoint. |

An attempted parallel launch of both Compose package targets is excluded: they share one declared
Postgres service and collided during destructive test setup, yielding deadlocks/duplicate seed
keys. Sequential cold reruns immediately passed. AC1 is promoted; AC2/AC3 remain partial, AC4 is a
failed enumeration oracle, and AC5 remains absent/body-blocked. RP-068–RP-071 and
`minigame-api-surface-lifecycle-audit.md` carry the repair and review route.

## Batch H — Combat Shared Data & Arithmetic discrimination

| Evidence | Result | Interpretation |
|---|---|---|
| Cold `./combat ./determinism -count=1` | Combat green; Determinism has no package-local tests | Go kernel is exercised through Combat; shared authority ownership remains cross-package. |
| Full cold client population | 39 files passed/2 skipped; 6,655 passed/15 skipped | TypeScript Combat plus every downstream shared-RNG consumer executed. |
| Removed ATK scaling independently in Go and TS | each population failed `atk_scaling_non_identity` with 100 versus 156 | AC2's non-identity arithmetic stage discriminates in both runtimes. |
| Collapsed every substream label to `crit` independently | Go failed the obedience vector; TS failed Combat plus Fiscal/Pitch/replay consumers | AC3's labeled shared authority and real consumers discriminate in both runtimes. |
| Removed one chart winning edge independently | Go and TS each failed with one win instead of two | AC4's exact cycle property discriminates in both runtimes. |
| Added nested client `combat/audit_nested/violation.mts` with `4 / 2` | boundary exit 2 naming the exact file | AC6's client half is recursive and discriminating on the real scanned tree. |
| Restored boundary/typecheck/product diff | gate exit 0; zero diagnostics; no product diff | Every mutation and temporary file was removed. |

AC2–AC4 are promoted. AC1/AC5 remain specification-blocked because their closed effect and literal
input-table contracts do not exist. AC6 is contradicted under the literal all-combat-path wording:
the proven gate scans only client modules while Go Combat contains native division. RP-072–RP-074
and `combat-shared-lifecycle-audit.md` carry body, implementation, and review-provenance repair.

## Batch I — Accessibility workflow discrimination

| Evidence | Result | Interpretation |
|---|---|---|
| Temporary desired lifecycle-focus test | failed in Chromium, Firefox, WebKit; focused Desk button removal leaves `<body>` active instead of the Offer heading | Axe/component semantics do not protect dynamic surface context. RP-082 is a reproduced implementation/oracle defect. |
| Temporary complete-Desk 320 px reflow test | failed in all three engines with `scrollWidth=647`, `clientWidth=320` | The fixed 1280×720 browser population cannot detect the real responsive failure. Source inspection identifies the unbreakable README row as the likely intrinsic-width source; that cause is an inference, not a mutation result. |
| Production reduced-motion source trace | CSS theme samples `matchMedia`; `GameUIShell` constructs `ShellController` with default false; no media-query change listener | Token reduction is real, but numeric interpolation/pulse and live preference changes are not integrated. |
| Restored declared Game UI browser target | functional 30 passed/3 skipped across engines; performance 1 passed/10 skipped | Both failed probes and screenshots were removed; the green baseline confirms current oracle blindness rather than residue. |

No screen-reader, 200%/400% zoom, coarse-pointer, or full-keyboard participant/manual record exists.
Missing Account, Minigame/Recovery, Gate/Wind Down, next-run, and later-tier surfaces cannot receive
accessibility credit from the five current fixture renders. RP-009/RP-082–RP-084 and
`accessibility-release-audit.md` carry the release contract and dependency route.
