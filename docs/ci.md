# Continuous Integration

The repository has one GitHub Actions workflow, `.github/workflows/ci.yml`, triggered by every
push and pull request, a nightly schedule, and manual dispatch. It has read-only repository
permissions and cancels an older run when a new commit arrives on the same ref.

The repository is public. CI uses GitHub-hosted `ubuntu-latest` runners; there are no self-hosted
runners, deployment credentials, or deployment steps.

## Blocking jobs

| Job | Repository command | Coverage |
|---|---|---|
| `server` | `make verify-server` | Go vet/tests plus generated production-formula drift |
| `client` | `make verify-client` | strict TypeScript and Node/V8 tests |
| `browser` | `make test-browser` | Chromium, Firefox, and WebKit suites |
| `schema` | `make verify-schema` | schema compilation plus production and fixture catalogs |

Every job has a five-minute timeout. The complete blocking workflow has a normative five-minute
elapsed-time budget; the first hosted measurement is pending the initial push. Until that run is
observed, the CI Baseline RFC remains implementing.

## Nightly numeric maintenance

Scheduled and manual runs add a non-blocking `numeric-maintenance` job outside the pull-request
latency budget. `make fuzz-ci` fuzzes canonical round trips for 30 seconds; `make vectors-check`
regenerates the deterministic shared corpus and fails if tracked bytes change. The job installs
the same pinned Go, Node, and pnpm toolchains as the blocking jobs and has a ten-minute timeout.

`make fuzz` remains the unbounded interactive command for deliberate local fuzzing. The bounded
target exists so automation always terminates.

## Reproducibility and caches

- Go reads its version from `server/go.mod`. Only the module-download directory is cached; the Go
  build cache is not.
- Node is version 24. pnpm reads the exact version from `client/package.json`, installs from the
  frozen lockfile, and caches only the pnpm dependency store.
- The browser job runs inside `mcr.microsoft.com/playwright:v1.62.0-noble`, exactly matching the
  `playwright` package. Browser executables come from the image and are not cached separately.
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

## Local use

Run the exact aggregate gate from the repository root:

```sh
make verify
```

The narrower commands are useful while iterating:

```sh
make verify-server
make verify-client
make verify-schema
make formulas-check
make test-browser
make fuzz-ci
make vectors-check
```

No CI job deploys anything. Compose, Caddy, migrations, websocket draining, and reconnect testing
belong to later RFCs.
