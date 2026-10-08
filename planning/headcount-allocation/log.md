# Headcount Allocation — implementation log

## 2026-10-08 — Source-independent S2 arithmetic

Baseline `dbd8e274`. Previous goal turn was progress (committed Clout help with
executed browser/real-service proof). This batch constructs accepted S2's pure
Go/TS arithmetic without choosing the explicitly held OD-1 seat source.

`server/headcount` and `client/src/headcount` project complete assignments into
producing/diverted heads, idle converters and unassigned seats. Safe-integer
bounds, complete keys and the converter graph fail closed; inputs and results
are independent. No clock, Decimal, catalog/save/version, rate, command or UI
integration is added. Only tests import the new modules. Before authoritative
wiring, register both prefixes and bump kernel identity with the actual semantic
integration; existing watched paths and kernel175 remain untouched here.

Evidence: twelve shared literal vectors, 7803 small-domain allocations per runtime
against independent arbitrary-precision models, maximum-domain cases, independent
conversion pairs, role order, fresh results and22 shared-condition refusals
(TS also tests fractions/NaN/Infinity). Final root commands/results:

- `make test-go GO_PACKAGES='./headcount' GO_TEST_FLAGS='-count=1'`: PASS,0.308s.
- `make test-go-ci CI_TEST_PACKAGES='./headcount' CI_TEST_FLAGS='-v'`: PASS (50359),
  actual Linux-amd64 container,0.102s. No DB operation is claimed by this pure package.
- `client/node_modules/.bin/vitest run --root client test/headcount-allocation.test.ts`:
  PASS42. Existing automatic Node/browser collection consumes the new test file.
- `make test-browser-focused BROWSER_TEST_FLAGS='test/headcount-allocation.test.ts
  --project=chromium --project=webkit'`: PASS84 (77778).
- `make typecheck vet GO_PACKAGES='./headcount'`: PASS (6621),zero Svelte errors/warnings.
- Firefox attempt48146: RED,zero tests;60s session-connect failure and180s launch
  timeout, with plugin-container sandbox-extension/graphics errors. Handle is terminal;
  no retry, relaxed budget, all-engine or hosted CI pass inferred.

RP-413: S8's3.1e15 converter/9e15 source assignment totals12.1e15, outside S4's
safe seat budget. It rejects in both kernels. A separate legal3.1e15/5.9e15 vector
passes. Direct Node observation confirms naive min still returns the exact9e15
despite an unsafe9.3e15 capacity; the mandated failing mutant is not established.
Do not claim AC3/AC5 complete or edit the author's accepted text around the result.

RP-414: Tier2 H1–H5 and Headcount S0–S5 disagree on catalog/state/commands and
output economy, not only seats. Both accepted bodies are unchanged. An async
seat-source question was sent to Marco; no answer/adoption is recorded. Reconcile
the whole boundary before live integration; do not treat a budget choice alone
as approval of every schema or output mechanic. Version literals are also occupied.

Review by: Codex (implementer first filter). Recorded by: Codex. Full new range
`dbd8e274` exclusive through this containing source/test/docs/record commit needs
designated cross-party review. No feature checkbox, archival, mint, push, complete
CI/release claim or narrowing of the nine-tier objective. Next: resolve the held
integration contract; other accepted lanes may continue in parallel with that hold.
