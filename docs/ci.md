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
| `game-ui-composed` | `make test-game-ui-composed` | Real Chromium, Vite, gameserver, Postgres, and WebSocket bootstrap-to-Desk flow |
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

Run the exact aggregate gate from the repository root:

```sh
make verify
```

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
make verify-ci-topology
make fuzz-ci
make vectors-check
make vectors-check-ci
```

`make test-go-ci` reproduces the complete Go suite cold on Linux/amd64 with
Postgres, including packages whose integration tests are not selected by the
focused `test-save-integration` target.
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
enforced. `make test-browser` first runs the functional three-engine matrix with this one
wall-clock-sensitive test disabled, then invokes the focused target in a fresh Chromium process.
This keeps Firefox/WebKit contention and unrelated test files out of the measurement without
changing any budget. The real-time 4× CPU / dropped-frame profile remains the manual release check
described in the Game UI docs.

`make test-browser-ci` runs the same functional-matrix-then-isolated-performance sequence with
`CI=true` in the pinned Linux Playwright image, using fresh anonymous `node_modules` and
pnpm-store volumes on every invocation. Actions installs the same pinned browser versions
explicitly on its ordinary Ubuntu runner. `make verify-server-ci` runs the complete server gate on
linux/amd64 against the repository's Postgres test service. `test-go-core` and therefore
`verify-server-core` always pass an explicit `-count=$(CORE_TEST_COUNT)` (default `1`), so restored
compiler/module caches cannot restore test results; increase `CORE_TEST_COUNT` for focused stress
runs. Both harness lanes have the same cold package-test contract through `HARNESS_TEST_COUNT`
(default `1`). `make verify-ci-topology` rejects exhaustive work in push CI, push triggers in
maintenance, unbounded/missing observations, success-only uploads, build-cache restoration, and
missing observation validation. It also binds the exact job populations and observation path; the
cache check covers both workflows. Use these targets when host-platform success could mask
scheduling, architecture, or cold-run behavior.

`make test-game-ui-composed` starts its isolated repository Postgres service, the real composed gameserver,
and Vite, then drives Chromium through anonymous bootstrap, an authenticated live
`/api/v1/founder/state` v2 round trip, and the actual Centrifuge WebSocket subscription. The
snapshot assertion and visitor-counter assertion prove both HTTP synchronization and the socket
handshake; no browser runtime or server transport is stubbed. The dedicated
`game-ui-composed` Actions job runs this same Make target on every push and pull request.

Go commands invoked by the Makefile use the ignored repository-local `.cache/go-build` directory.
Focused tests can run without writing to a user-level cache or requiring sandbox permission:

```sh
make test-go GO_PACKAGES='./harness ./transport'
```

HTTP and WebSocket tests use a real `net/http` client/server exchange over in-memory `net.Pipe`
connections. They exercise upgrades, framing, and protocol recovery without binding a localhost
port, so ordinary test runs do not require network permission.

No CI job deploys anything. Compose, Caddy, migrations, websocket draining, and reconnect testing
belong to later RFCs.
