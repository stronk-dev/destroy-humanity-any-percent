# CI remediation log — 2026-08-10

## GitHub Actions run 31400646088 — implementation record

- **Implemented by:** Codex.
- The browser job exited before testing because the official Playwright container did not contain
  `make`; the workflow now installs it and continues through the repository-owned
  `make test-browser` entrypoint.
- The server job exposed an architecture-dependent `Decimal.Int64Exact` rejection: a canonical
  integer produced by `Floor` reconstructed one ULP below its integer value on linux/amd64.
  Classification now requires equality with the normalized Decimal representation of the nearest
  integer, preserving fractional rejection without depending on a second float reconstruction.
  Kernel version `0.3.89` records the semantic correction.
- Two DB-backed tests had stale inputs hidden by local runs without `TEST_DATABASE_URL`: the
  minigame tenant omitted its required `trust_ppm` scaling destination, and the Exit fixture emitted
  a schema-v1 `run_ended` envelope while expecting logged genesis behavior. They now use the shipped
  tenant grammar and the real logged Exit path with the schema-v2 terminal payload.
- `make test-go-ci` is the normal reproducer: it runs every Go package uncached in a linux/amd64
  container against the declared Postgres service. `CI_TEST_PACKAGES` and `CI_TEST_FLAGS` retain
  focused-debug flexibility without bypassing Make.

Verification completed so far: focused Decimal regression, kernel-history guard, `make
test-browser` (114 files / 19,932 assertions), focused logged-Exit Postgres regression, and the full
`make test-go-ci` package graph. The complete `make verify` output was read to exit 0: Go vet and
all packages, content/formula/API drift guards, balance harness, TypeScript/Svelte diagnostics,
client build, 6,640 unit assertions (4 skipped), boundary/history/schema/copy guards, and the full
114-file / 19,932-assertion browser suite all passed.
