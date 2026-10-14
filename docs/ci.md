# Continuous Integration

The repository has two GitHub Actions workflows. `.github/workflows/ci.yml` is the blocking
push/pull-request workflow. `.github/workflows/maintenance.yml` runs exhaustive harness and numeric
evidence on a nightly schedule or manual dispatch. Both have read-only repository permissions and
cancel an older run when a new commit arrives on the same ref.

The repository is public. CI uses GitHub-hosted `ubuntu-latest` runners; there are no self-hosted
runners, deployment credentials, or deployment steps.

## Blocking jobs

| Job | Repository command | Coverage |
|---|---|---|
| `server` | `make verify-server-core` | Cold Go vet/tests outside the parallel harness lane, plus generated production-formula/API drift |
| `harness` | `make verify-harness-fast` | Cold harness tests, role proofs, Commons invariance, and complete balance/epoch history guards; no pacing/relevance simulation |
| `client` | `make verify-client` | strict TypeScript and Node/V8 tests; full Git history is required by KV-1 |
| `browser` | `make test-browser` | Chromium, Firefox, and WebKit functional suites, then isolated Chromium performance |
| `game-ui-composed` | `make test-game-ui-composed` | Production-built Chromium/gameserver/Postgres/WebSocket gameplay and recovery; default Cosmetic/care and unminted Clout journeys; two built-client no-payment negatives |
| `schema` | `make verify-schema` | schema compilation plus production and fixture catalogs |

The fast harness, client, and schema jobs have five-minute ceilings. Server, browser, and
composed-browser jobs have ten-minute ceilings. Jobs run in parallel and the complete blocking
workflow retains the RFC's sub-five-minute measured target; ceilings are failure containment, not
the latency objective.

## Scheduled maintenance evidence

Scheduled and manual runs execute a separate workflow outside the pull-request latency budget.
Its `harness-evidence` job first runs the fast gate, then runs the exhaustive pacing/relevance
population through `make harness-observe`. GNU `timeout` sends SIGINT at 50 minutes inside a
55-minute job ceiling, allowing the observation recorder to persist explicit incomplete state.
A completed artifact is validated; the observation uploads after success or failure for 30 days,
and a missing artifact fails loud.

The parallel `numeric-maintenance` job runs `make fuzz-ci` for bounded canonical-round-trip fuzzing
and `make vectors-check` for deterministic shared-corpus drift. Maintenance Go jobs cache only the
module download directory, never compiled or test outputs.

`make fuzz` remains the unbounded interactive command for deliberate local fuzzing. The bounded
target exists so automation always terminates.

## Reproducibility and caches

- Go reads its version from `server/go.mod`. Only the module-download directory is cached; the Go
  build cache is not.
- Node is version 24. pnpm reads the exact version from `client/package.json`, installs from the
  frozen lockfile, and caches only the pnpm dependency store.
- The browser job runs on the ordinary Ubuntu runner and installs Chromium, Firefox, and WebKit
  explicitly from the pinned Playwright package. This is the same runner pattern used by the
  composed browser job; browser executables are not cached separately.
- Workflow actions use supported major release tags. Updating Playwright requires changing the
  package, lockfile, and container tag together.

## Balance-schema gate

`make verify-schema` runs the pinned Ajv Draft 2020-12 validator in strict mode. It compiles
`balance/economy.schema.json`, which also validates the schema against its meta-schema, then checks:

- every JSON file under `balance/catalogs/` as production data, if that directory contains any;
- every `balance/testdata/valid/*.json` fixture must pass;
- every `balance/testdata/invalid/*.json` fixture must fail.

At least one positive and one negative fixture are required, so deleting the validator's proof
cases is itself a build failure. Runtime Go and TypeScript catalog validation remains independent;
JSON Schema is an early content-authoring gate, not the authoritative gameplay loader.

## Published-formula drift gate

`make formulas-check` regenerates `docs/generated/production-formulas.json` from the Go multiplier
boundary and fails if tracked bytes differ. It publishes the exact production-rate shape, slot
order, and within-slot order used by authoritative arithmetic. The target is part of
`make verify-server`, so changing code without updating the player-auditable artifact fails the
blocking server job.

## Balance-history and epoch gate

`make harness-check` validates both pacing output and the complete committed governance history.
The baseline guard enforces separate, artifact-only `BALANCE-CHANGE:` regeneration commits. The
epoch guard starts at [`balance/epochs/phase0.json`](../balance/epochs/phase0.json) and rejects any
later constants-artifact commit that does not register its exact resulting hash. Ordinary hotfixes
may only extend the current epoch; balance changes mint one successor with a same-commit public
changelog. A hardcap reduction additionally requires an explicit `Cap migration:` policy.

Both guards reject shallow history and uncommitted governed artifacts, so local and hosted checks
apply the same rules. CI checks out complete history for the server job.

## Local use

Run the exact blocking push aggregate from the repository root:

```sh
make verify-push
```

`make verify-push` is mechanically bound to the six workflow leaf commands in job order. The
broader developer aggregate, `make verify`, retains the exhaustive server harness but does not
stand in for the composed browser job and must not be cited as push-workflow parity.

### Kernel correction review evidence

Existing kernel history corrections must bind to a completed independent review
record, not instructions to obtain one. The guard reads visible reviewer/decision
metadata in the same section as the exact offending-commit anchor; examples in
code fences, blockquotes, indented code or comments do not count. Empty/pending
fields, conflicting decisions and explicitly labelled first filters refuse.
Correction citations must use canonical forward-slash `planning/.../log.md`
paths: empty, dot/parent segments, backslashes and NUL refuse before lookup.
Canonical active and archive citations, including the existing archive fallback,
remain valid. This is citation grammar, not filesystem identity or authentication.
Completed negative findings and approved mechanical hash-remap bindings remain
valid: correcting a missed version signal is not approval of the original feature.
Review execution and authorship still require the project's independent process;
parsing metadata is not authentication. No new historical exception class, version
waiver or history rewrite is authorized by this conformance repair (RP-131).

### Research provenance versus regression results

The Clout activation, action/buff/mode and Postgres anchor observers execute current code and
compare their complete results to the retained observations. Receipts, events, encoded states,
population counts, SQL outcomes, versions and acceptance labels remain byte-exact. The recorded
`source_sha256` values describe the historical measurement: a source-only edit is reported in
verbose test output, not treated as a changed gameplay result. Source-path populations and
digest encoding still validate. This does not verify the historical checkout at current HEAD.

Ordinary regression runs never rewrite these artifacts. Their explicit measurement switches
still write actual current source identities when deliberately requested. A changed result must
be investigated rather than restamped to get green. The independent Clout partition-invariance
property remains a release/CI failure until the underlying arithmetic contract is satisfied.

### Focused commands

The narrower commands are useful while iterating:

```sh
make test-go-ci
make verify-server
make verify-client
make verify-schema
make formulas-check
make test-browser
make test-browser-ci
make test-game-ui-performance
make verify-server-ci
make verify-harness-ci HARNESS_WORKERS=12
make verify-harness-fast
make verify-push
make verify-ci-topology
make fuzz-ci
make vectors-check
make vectors-check-ci
```

`make test-go-ci` reproduces the complete Go suite cold on Linux/amd64 with
Postgres, including packages whose integration tests are not selected by the
focused `test-save-integration` target.
On Apple Silicon, that Docker lane disables AVX2 inside the emulated amd64 test container because
Docker Desktop's QEMU CPU-feature combination corrupts the age authenticated-payload path. The
hosted Actions runner is native amd64 and retains ordinary CPU detection, so the optimized path is
still part of blocking CI; the local container proves the portable fallback rather than reporting
a false cryptographic failure from emulation.
The save-test compose stack publishes Postgres on `SAVE_TEST_HOST_PORT` (default
`55432`); pass a different Make value, for example
`make test-go-ci SAVE_TEST_HOST_PORT=55434`, when another isolated test stack is
already using the default host port. The test container always uses the service
name and port `5432`, so this flexibility does not change the CI database path.

`make vectors-check-ci` regenerates the shared Decimal corpus under Node 24 on
Linux/amd64 and rejects any byte drift. This complements the host
`make vectors-check` target: both must produce the same file, preventing
host-libm rounding differences from surfacing only in scheduled Actions runs.

`make test-game-ui-performance` owns the browser job's deterministic sixty-second Game UI
cadence simulation: 1,200 snapshot inputs (20 Hz) are grouped through 600 shared formatter flush
windows (10 Hz) in a 1280×720 Chromium viewport, with the commit count and 200 ms long-task ceiling
enforced. `make test-browser` first runs `make test-browser-functional` with this one
wall-clock-sensitive test disabled, then invokes the performance target in a fresh Chromium process.
This keeps Firefox/WebKit contention and unrelated test files out of the measurement without
changing any budget. The real-time 4× CPU / dropped-frame profile remains a manual release check;
the 2026-08-22 Game UI archival record preserves its latest passing environment and measurements.

The functional target runs the broad Chromium/Firefox/WebKit population excluding only
`test/typer-table-browser.test.ts`, then runs that complete file in a fresh three-engine
invocation. This keeps its native keyboard journeys out of the broad workload without changing
inputs, assertions, deadlines, worker defaults or engine coverage. The required Garage real-minute
idle observation remains in the broad blocking population. A broad-phase failure stops Make:
later phases are unexecuted, not passed. `make test-browser-focused` remains a single invocation
with the caller's original file/test selectors; it does not automatically append Typer.

For retained full-functional JSON results, set `BROWSER_REPORT_DIR=.cache/browser-results`.
The target writes separate `functional-broad.json` and `functional-native.json`; compare their
union when checking full coverage. A shared `--outputFile` in `BROWSER_TEST_FLAGS` is rejected
because the second phase would overwrite the first. Focused invocations still support their
ordinary Vitest reporter/output flags. Explicit project/file/test overrides narrow verification
and are not evidence of the default three-engine gate.

`make test-browser-ci` prepares GNU Make in the disposable Playwright container, installs
frozen dependencies and calls the same root `make test-browser`
target as Actions, rather than maintaining a second Vitest command in Compose. It runs the
same broad-functional → native-keyboard → isolated-performance sequence with
`CI=true` in the pinned Linux Playwright image, using fresh anonymous `node_modules` and
pnpm-store volumes on every invocation. Actions installs the same pinned browser versions
explicitly on its ordinary Ubuntu runner. Full-population concurrency controls and their failures
are recorded in the [CI log](../planning/scaffolding-and-ci/log.md); a focused or overridden run
does not establish success of this ordinary target, hosted reliability or the five-minute
acceptance floor.

The topology check also verifies that the Linux container delegates to the unmodified shared
target; copied, functional-only, focused or project-narrowed commands are rejected. This guards
execution parity, not browser behavior or hosted timing; the actual lane must still execute.

Reputation host tests emit `reputation-host-deadline` on an actual test abort, before teardown
removes the mounted evidence. The payload identifies the case/user agent, import/native-RPC/
drain/unmount stage, outstanding input, settled-command timing and controlled UI/network state.
These are native-command promise observations, not proof of delivered gameplay or a timeout
cause. The intentional 50 ms control uses `reputation-host-expected-deadline`; successful ordinary
cases emit neither marker. The observer does not cancel input or bypass the existing drain rule.

The normal container command wraps the frozen installation and both browser lanes in GNU
`timeout --verbose --signal=KILL 10m`, matching the hosted browser job's existing containment
ceiling. A stuck test process group cannot keep the local job alive indefinitely. A deadline
prints the timeout's signal message and fails the command; partial pass counts are not a verdict.
Exit137 without that message is not enough to diagnose a deadline rather than another kill.
This is containment, not a larger test timeout or permission to exceed the sub-five-minute
hosted acceptance target. Explicit Compose command overrides replace this wrapper too.
The timer starts when the container command starts; it does not bound a stuck Docker client
before startup. Require the CLI's terminal result as well as test results, and report an
unavailable daemon as unavailable verification, never as a passing job.

The Typer whole-game and Reputation plan/host native-input witnesses bind their keyboard producers
to Vitest's test signal through the shared test-only `native-keyboard-lifetime` helper.
A fired deadline remains a failed journey: future input is refused, an already-started
native RPC drains before owned-fixture teardown, and a successor fails its precondition if
input is still pending. The keyboard path and existing journey deadline are unchanged.
Their isolation controls deliberately fire a framework deadline and independently check the
reason and untouched successor; those expected failures are not player-acceptance passes.
Reputation also refuses expired disclosure/selection helpers before accessing or focusing
disposed DOM. Fixture teardown is idempotent and registered on the original test context.
The Reputation host registers the same ownership at mount: even a suspended timed-out
producer cannot keep its old UI/socket alive. Its expired navigation helper refuses before
polling or focusing disposed controls. Network/storage in these host tests remain controlled
fixtures, not live-server/Postgres proof.

If Typer's complete native journey reaches its deadline, its test-only observer prints
`typer-native-journey-deadline` before teardown: current stage/elapsed time, engine phase,
prompt/revision/command count, pending state, input length and native-RPC flight state.
It also retains every completed stage's elapsed time in order, without truncation, so a
deadline during a short final stage does not conceal where earlier time was spent. This
history is emitted only with the cancellation capture; healthy/disposed observers remain silent. It contains
stage names and durations, not key sequences, prompt text or credentials.
Command callbacks additionally record pending-state render, pure-engine call/wait and
result-decode/snapshot-render elapsed intervals, plus any active callback phase. These are
elapsed wall-clock intervals, including scheduling waits—not isolated CPU measurements.
They can run inside a native keyboard promise, so a later zero-length response wait does
not prove that engine/render work cost zero. Completed records contain command kinds and
durations only. The ordinary successful journey verifies all ten callback records without
printing a healthy observation; no native action or deadline is replaced.
A separate diagnostic invokes the same actual pure engine behind a timed asynchronous
hold. Its independently observed hold duration must appear in the apply interval; deliberately
reporting zero fails in both browsers. The normal invocation still calls the engine directly.
It arms when the journey body starts, before its first await, so driver loading, content identity,
engine creation and fixture mount are also distinguishable. Until the fixture exists, context
explicitly reports `fixture_ready: false` rather than reading uninitialized engine/DOM variables;
elapsed time includes initialization. An early-disposal hook covers setup failures as well.
It does not record typed text or change input delivery, assertions, timers or concurrency.
One real-timeout control requires a connected fixture's state to be sampled before removal;
another pauses the same journey before initialization and requires an early phase record plus
refusal/drain of the expired continuation. Both use controlled50ms deadlines, not the ordinary
journey's15s bound. The ordinary sequence is unchanged and has no setup gate.
Healthy/disposed observers stay silent. This is diagnostic context, not a latency correction:
a later passing population does not resolve an earlier unexplained deadline.

The Garden host fixture waits for its actual socket-factory callback before asserting the exact
URL; elapsed event-loop turns do not establish startup completion. Construction does not imply
transport readiness: the existing protocol-frame checks remain. Owned-fixture cleanup belongs
to the caller's explicit test context and runs on failed initialization or test completion, with
an independent timed-out-startup teardown control. No startup retry policy or deadline is changed.

Garage's mandatory GS5-A2 real-minute idle observation lives in
`test/garage-idle-opportunity-browser.test.ts`, alongside the interaction file
`test/garage-surfaces-browser.test.ts`. Both consume `garage-surfaces-fixture.ts` and stay
in ordinary blocking browser collection. The observation still runs real native timers for
at least60s, checks every250ms for zero intents and keeps its70s deadline. Selecting only
the interaction file no longer covers that acceptance criterion; select both for a focused
whole-Garage check. No maintenance-only move, fake clock or shortened acceptance is introduced.

Two Node-side fixtures, `css-dependency-graph.test.mjs` and
`opportunity-composed-proof.test.mjs`, are excluded by exact path from browser collection because
their helpers require Node APIs. Both remain in the mandatory client Node suite; other test
formats are collected by default. This is execution in the correct runtime, not an exemption
from testing the composed driver or its proof oracles.

`make verify-server-ci` runs the complete server gate on
linux/amd64 against the repository's Postgres test service. `test-go-core` and therefore
`verify-server-core` always pass an explicit `-count=$(CORE_TEST_COUNT)` (default `1`), so restored
compiler/module caches cannot restore test results; increase `CORE_TEST_COUNT` for focused stress
runs. Both harness lanes have the same cold package-test contract through `HARNESS_TEST_COUNT`
(default `1`). `make verify-ci-topology` rejects exhaustive work in push CI, push triggers in
maintenance, unbounded/missing observations, success-only uploads, build-cache restoration, and
missing observation validation. It also binds the exact job populations and observation path; the
cache check covers both workflows. The same guard binds every blocking job to its one Make leaf and
binds those six leaves, in order, to `make verify-push`; removing, adding, or substituting either
side fails a permanent negative fixture. Use these targets when host-platform success could mask
scheduling, architecture, or cold-run behavior.

The Game UI worker witness uses the real native Worker and consistent initial/refresh fixture
authority. Its five-second visible-growth bound is unchanged; observation records native input,
prediction, first-output latency, visibility and refresh count. Paired visible-return controls
separately require bounded native readiness before their two-second measurement, so startup
cannot silently invalidate the comparison. R-010's dossier retains the fired initial control and
the original unresolved full-lane failure; local repeat passes are not hosted reliability proof.

The browser configuration also emits passive `browser-module-http` JSON for Route fixtures,
native prediction-worker modules and RP-526's four unreported test modules: request start,
HTTP finish/premature close, and a server
close-summary including pending requests. Native witness logs now identify per-test worker
command/output/error/termination ordering before and after cleanup. Neither observer retries,
intercepts or suppresses failures. IDs are scoped to each server/test observer, not global.
The optional existing `VITEST_PW_DEBUG=1` provider logging reports request failures; teardown
cancellations also occur in passing runs. See the retained
[cold-run dossier](../planning/platform-alignment/browser-module-research.md).
Latest local repetition is incomplete: Firefox reports no device space and Docker's writable
filesystem is full, despite free shared memory and host space. Passing individual repetitions
do not establish reliable CI. Broad Docker pruning is not a repository verification command.

The primitive keyboard/era assertion includes passive focus and native key/click diagnostics
in its failure message: frame focus, active button, before/after era, delivery target, trusted
input and default prevention. Its Enter input and exact expected era are unchanged; it does
not retry or refocus to obtain a pass. Native moved-focus and prevented-Enter controls check
the diagnostic distinctions. The fixture and listeners are removed even after an assertion
fails. This observation is not a correction of the open hosted Firefox failure (RP-451).

`make test-refresh-census` is a manual Account refresh population observer, outside
CI and `verify`. It runs the declared Postgres integration service cold and requires
all four named tests/seven cases to execute and pass. A package pass with missing,
skipped, failed or duplicated cases exits nonzero. It retains counts and listed
source hashes, never private test output. `make test-refresh-observer` checks its
synthetic controls; these are not database evidence. The actual host control
`node client/tools/observe-refresh-population.mjs --missing-db-control` rejects
Go exit zero with four dependency skips (zero/seven cases). Normal Postgres mode
is still unexecuted pending Docker capacity; see the
[census and retained negative](../planning/platform-alignment/session-refresh-contract-census.md).
Node controls use `observe-refresh-population.fixtures.mjs`, explicitly invoked
by Make, so they are not misclassified as Vitest client tests. The initial
filename collision (RP-239) was caught by `verify-client` and corrected. That
composite still fails the existing RP-131 history guard; passing its client
population and separately run remaining gates is not whole-CI success.

`make test-game-ui-composed` first runs the five production-asset oracle checks, twenty-five
care-projection controls and five no-payment fixture/classifier checks with Node's native test
runner. Those files use the `.fixtures.mjs`
suffix (`production-client-proof.fixtures.mjs`, `care-snapshot-proof.fixtures.mjs`,
`cosmetic-n5-fixture.fixtures.mjs`), not
`.test.mjs`: Vitest cannot register `node:test` suites. They remain mandatory in this blocking
Make/Actions lane rather than being silently excluded from all automation. The client/browser
Vitest configurations and their legitimate `.test.mjs` population are unchanged.

The target then starts its isolated repository Postgres service. Its first driver
starts the real composed gameserver and serves production-built client assets, then drives Chromium through anonymous bootstrap, an authenticated live
`/api/v1/founder/state` v4 round trip, and the actual Centrifuge WebSocket subscription. The
snapshot assertion and visitor-counter assertion prove both HTTP synchronization and the socket
handshake. The same witness uses server-side precondition setup, then requires visible enabled
Gate/Wind Down controls to traverse `runtime.ts`, renders both terminal variants, and continues to
the exact next Company. No browser runtime, server transport, fixture clock, or direct browser
intent call is substituted. The dedicated
`game-ui-composed` Actions job runs this same Make target on every push and pull request.
The Fiscal refusal population waits for Harvest's actual actionability before focusing it and
sending native Enter. A preceding harvest resets the quarter to ripening; focusing its disabled
control would leave focus on navigation and emit no intent. A trial click only waits, never
activates. The exact invalid/stale responses, request counts, persisted-state checks and fresh
consent remain mandatory, without retries or longer response deadlines.
The witness first requires exclusive ownership of `127.0.0.1:18081`, builds an ignored
repository-local gameserver binary, and owns that exact child through teardown. A stale listener is
a loud setup failure and cannot satisfy `/readyz` for a newly failed child.
The second driver resets only that named ephemeral test database, pins Cosmetics in a synthetic
fixture epoch, serves the compiled client through a same-origin static/API/WebSocket proxy, and
requires live visitor presence before T0 locked → T1 Buy → server-owned reload. It audits browser
requests and refuses payment APIs or unexpected destinations. The fixture does not mint production
content. The third invocation selects `--axis-stack` in the same driver: only test cash is seeded;
native generator/PR purchases earn attainment and persist ownership, while rendered factors,
reload and the SQL head/event are checked. `make test-clout-composed` selects that variant alone.
It does not mint content or claim natural pacing. The target then invokes `make test-cosmetic-n5`:
two fixture-only Vite builds inject checkout-fetch and PaymentRequest attempts into the actual
native Buy callback. Both must reach the real applied Buy receipt, then fail specifically at
the primary N5 trap before egress, complete cleanup and exit 1. An unrelated failure, a green
child or the independent route-abort safety net cannot satisfy that gate. All Cosmetic builds
use isolated temporary output directories; negative bytes never replace `client/dist`.
All earlier assertions and the six-job topology
remain; no timeout grows. Recorded local expanded-run failure RP-389 stops in the first driver
after a correct locked-Pitch409 but before its visible notice, so that run does not execute
the later variants. A subsequent expanded run executes all three populations successfully;
it does not close RP-389 or establish reliability/hosted latency/green. Controlled native host
evidence demonstrates offer preemption removing Pitch's notice. The driver now handles that
interruption through visible Decline/navigation and fresh Start consent, bounded to three
submissions. Each must return the exact409 pair; an offer alone cannot pass the required visible
mounted-launcher notice. Individual30s waits remain. The corrected target passes locally with
no declines, so its post-response re-consent branch is not real-service proof yet. Failure
diagnostics retain mounted public context; original-run attribution/reliability remain open.
On local ARM64 hosts, use
`make test-game-ui-composed GAME_UI_COMPOSE_FILES='-f compose.game-ui-test.yml -f compose.game-ui-arm64.yml'`;
hosted x86 CI uses the default Compose file. RP-172's test-only actionability correction
addresses observed intermittent Pitch no-request timeouts; an independent hosted run and
cross-party review are still needed before calling that lane reliable.

Go commands invoked by the Makefile use the ignored repository-local `.cache/go-build` directory.
Focused tests can run without writing to a user-level cache or requiring sandbox permission:

```sh
make test-go GO_PACKAGES='./harness ./transport'
```

HTTP and WebSocket tests use a real `net/http` client/server exchange over in-memory `net.Pipe`
connections. They exercise upgrades, framing, and protocol recovery without binding a localhost
port, so ordinary test runs do not require network permission.

Deployment has two explicit cold Linux/amd64 component lanes outside hosted push CI.
`make test-deployment-backup` runs the Postgres 16 encrypted dump/clean-restore identity population.
`make test-deployment-release` uses an isolated Compose project with real Postgres, a production-
boundary gameserver and real Caddy internal TLS. It holds one admitted intent at a database row
lock, proves readiness-down plus exact drain refusal/courtesy/closure, releases and completes the
admitted request, restarts through Caddy, then restores the exact encrypted pre-upgrade backup into
a newly created database and authenticates through the same Caddy HTTP/WebSocket path. Successful
runs remove their isolated test containers, network and temporary volume; failures retain them for
diagnosis. Neither lane claims the final exact-bundle clean-host rehearsal.

No CI job deploys anything. Destructive restore, release/rollback, alert delivery and clean-host
rehearsals remain explicit manual release lanes; blocking CI proves their static and component
contracts without receiving deployment secrets.
