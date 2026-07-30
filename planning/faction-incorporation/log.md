# Faction & Incorporation — implementation log

## 2026-07-30 — accepted and started

- Owner answered the acceptance bounce with FA/FB/FC: stocks are run-scoped int64 Company fields,
  accrual is one unit per attended minute with carried milliseconds and accrual-only saturation,
  and all four Phase-0 faction objects are literal.
- Acceptance re-read confirmed these contracts close the previously recorded resource/rate/data
  gaps. RFC moved to implementing; no inferred balance mechanics are required.
- Implementation order follows the dependency surface: catalog/identity → save shape → intents and
  Commons binding → accrual → boards/reset/docs.

## 2026-07-30 — catalog and epoch-guard foundation

- Added the strict Go catalog and matching JSON Schema/client semantic validator. Both enforce the
  literal four-faction set, one producer/consumer per stock, the single Hamiltonian cycle, empty
  Phase-0 modifier hooks, mechanical copy keys, and the Open Source tithe against the Commons band.
- The accepted RFC grows the constants-artifact manifest, exposing an intentional limitation in
  the existing epoch guard: it rejected every artifact-set change, including successor-RFC growth.
  The guard now permits append-only artifact additions only in a `BALANCE-CHANGE` mint; hotfixes,
  removals, and rewrites still fail closed. Repository-history tests cover both acceptance and
  rejection paths, and the baseline guard recognizes faction inputs.
- Focused Go suites are green. `make verify-schema` reached the existing `pnpm --dir client`
  mirror stall after the validator had already passed earlier in this session; it was interrupted
  rather than misreported as green and will be rerun in the final verification batch.
