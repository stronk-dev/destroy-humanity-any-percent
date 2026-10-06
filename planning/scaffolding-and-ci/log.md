# CI Baseline — Running Log

## 2026-07-28 — Start

- Owner selected a public repository, resolving the draft's only blocking owner decision.
- Narrowed the inherited draft to executable CI behavior; deployment, reconnect, assets, copy,
  and formula policy gates have named future owners.
- Verified current primary documentation before acceptance: checkout/setup-go/setup-node and
  pnpm setup use supported v6 majors; Playwright requires its Docker image version to match the
  installed package, which is 1.62.0 here.
- Selected Node 24 and the Noble Playwright 1.62.0 image. Browser binaries come from the image and
  are deliberately not cached.
- Selected pinned Ajv 8.20.0 for Draft 2020-12 schema and fixture validation so local and CI use
  one repository-owned command rather than a globally installed utility.

## 2026-07-28 — Local implementation

- Split the Makefile into `verify-server`, `verify-client`, `verify-schema`, `test-browser`, and
  aggregate `verify` targets without executing the browser suite twice.
- Added Ajv 8.20.0, a repository-owned schema verifier, and one positive plus one negative catalog
  fixture. With a deliberately malformed file placed in `balance/catalogs/`, `make verify-schema`
  failed on the unknown field as required; removing it restored the green gate.
- Added the four-job GitHub workflow with read-only permissions, push/PR triggers, branch-scoped
  cancellation, frozen pnpm installs, Go-module and pnpm dependency caches, and no browser-binary
  cache.
- `pnpm install --frozen-lockfile`, every narrow non-browser target, and aggregate `make verify`
  pass locally. The aggregate run completed in 5.86 seconds and included 6,321 Node tests plus
  18,963 browser tests across Chromium, Firefox, and WebKit.
- The workflow file cannot be executed on a hosted runner without pushing. Owner previously said
  not to push, so the RFC remains `implementing`; hosted completion under five minutes is the only
  unverified acceptance gate.

## 2026-08-10 — designated cross-party verdict: CI-repair batch {4a45b8d, 0db9768} — BOTH APPROVED

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.
**The Decimal fix verified genuine at bit level:** the shipped Int64Exact classified via a lossy
float64 reconstruction whose snap-tolerance comparison the Go arm64 backend FMA-fuses (proven:
the darwin diff 7.46e-11 is not a ULP multiple — only possible with an unrounded intermediate),
while linux/amd64 rounds separately (exactly 1 ULP ≥ the 1e-10 snap) and REJECTED canonical
integers. Reproduced in both directions across real architectures. **Blast radius bounded:** the
sole consumer fails CLOSED (ErrInvalidEngineState — availability fault, never corruption); stored
mantissa/exponent bits identical on both platforms; no persisted state/receipt/replay/wire could
have diverged; unpushed repo, zero production exposure. The fix (round-trip through the
normalized representation) is architecture-independent, only tightens acceptance, and the 0.3.89
bump is honest and REQUIRED. No shared-vector op exists for this classification (no TS twin —
correctly); the Go-side pin + the linux/amd64 CI gate discriminate. Fixture repairs test-tier
(both genuinely stale, both reproduced); `make test-go-ci` faithfully mirrors the Actions job and
ran GREEN under emulation; the plan record faithful with no box flips.
**Routed follow-ups (non-blocking):** R-F1 (MEDIUM) — the root FMA sensitivity remains in
toFloat64's snap comparison, and Floor IS a shared-vector op: force materialization (explicit
float64 conversion) in a future kernel bump + a razor-edge floor golden vector. R-F2 (LOW) — the
regression test discriminates only on linux/amd64 (by nature; the CI gate carries it). R-F3
(LOW) — docs/ci.md owes the test-go-ci line. R-F4 — a Claude-side numbering slip, fixed same day.

## 2026-08-11 — Actions run 31486886470 browser failure reproduced and repaired

- **Implemented by:** Codex. **Recorded by:** Codex. This is an implementation record pending the
  ordinary designated review; it is not an approval.
- The public run metadata isolated the failure to the Playwright `browser` job at `9a97543`.
  Client and schema were green; server was cancelled by workflow fail-fast rather than failing.
  The local GitHub CLI credential was invalid, so the public Actions API supplied job metadata and
  the failure was reproduced in the exact `mcr.microsoft.com/playwright:v1.62.0-noble` image.
- Linux Firefox exposed a test-clock race: the real Worker/render integration was required to
  change the visible amount within a fixed 750 ms while all three browser projects ran
  concurrently. The Worker remained correct, but Firefox was scheduled after that arbitrary
  deadline and the assertion observed `100 == 100`.
- Commit `61da160` retains the same discriminating observable contract but waits up to five seconds
  for the amount to change, polling every 50 ms. It does not stub the Worker, skip Firefox, or
  weaken the expected output. The exact CI image then passed all 120 browser files and 19,972 tests
  with two declared skips; the focused ordinary Make target passed all three engines locally.

## 2026-08-11 — Actions run 31486886470 server timeout repaired

- **Implemented by:** Codex. **Recorded by:** Codex. Pending the ordinary designated review.
- The server job began at 11:30:16Z and GitHub cancelled it at 11:35:32Z while `make
  verify-server` was still running: the five-minute job budget, not a test assertion, terminated
  the gate. A cold linux/amd64 reproduction took 4m30s merely to finish the Go/Postgres suite and
  generation checks before entering the composed balance harness, proving the old timeout
  structurally insufficient.
- The server job now has a ten-minute ceiling without removing, splitting, caching away, or
  skipping any verification. `make verify-server-ci` reproduces the entire gate on linux/amd64
  with real Postgres and completed green. `make test-browser-ci` similarly owns the exact pinned
  Playwright-image reproduction and completed all 19,972 browser tests green. Both normal targets
  are documented in `docs/ci.md`; the existing host targets remain unchanged.

## 2026-08-11 — Actions run 31504640635 browser environment repaired

- **Implemented by:** Codex. **Recorded by:** Codex. This is an implementation record pending the
  ordinary designated review; it is not an approval.
- Public Actions metadata isolates the new failure to the `browser` job at `ebf3ddf`: dependency
  installation succeeded, but `make test-browser` exited 2 after roughly two seconds. The server,
  client, schema, and composed-browser jobs all passed; notably, the composed browser job already
  used the ordinary Ubuntu runner plus an explicit Playwright install.
- The full browser suite was reproduced with `CI=true`, fresh dependencies, and the pinned
  `mcr.microsoft.com/playwright:v1.62.0-noble` image. A clean rerun completed all 19,993 assertions
  with two declared skips, so no product assertion or browser implementation was weakened.
- The Actions browser job now uses the same ordinary-runner setup that passed the composed-browser
  job and installs all three pinned browser engines explicitly. The local `make test-browser-ci`
  path now sets `CI=true` and uses fresh anonymous dependency volumes, eliminating the warm named
  volumes that made its prior “exact” claim inaccurate and polluted the tracked pnpm-store index.

## 2026-08-11 — Actions run 31506864417 minigame-start identity failure repaired

- **Implemented by:** Codex. **Recorded by:** Codex. This is an implementation record pending the
  ordinary designated cross-party review; it is not an approval.
- The failing server job was read directly through `gh run view --log-failed`. Its sole assertion
  failure was `TestComposedMinigameAPILifecycleUsesPinnedTenantResolverIntegration`: session create
  returned `500 internal_invariant/minigame_api`. Repeating that test against cold Postgres exposed
  the intermittent underlying error: `event intent does not match mutation`.
- Root cause: the minigame-start coordinator validated Founder events against the public session
  ID even though `ApplyFounderLogged` correctly attributes its lazy Fiscal sweep to the distinct,
  server-authored Founder command intent ID. The defect surfaced only when the 300 ms Fiscal auto
  boundary elapsed between fixture creation and the transaction timestamp.
- The coordinator now validates the decision against the Founder intent ID. Its Postgres fixture
  makes the Fiscal sweep unconditional, keeps session/intent IDs distinct, and asserts the stored
  event identity. Renaming the fixture into the `Integration` test namespace also closes the gate
  hole that previously kept this file-level integration test out of `make test-save-integration`.
- Discrimination proof: restoring the old session-ID validation makes the ordinary Postgres target
  fail deterministically at the original error; restoring the fix makes the same complete target
  green. The exact cold linux/amd64 `make test-go-ci` job then passed every package. Kernel 0.3.93
  is an honest semantic bump. The Make targets now expose repeat-count selectors alongside their
  existing package/regex selectors so focused stress runs remain ordinary repository commands.

## 2026-08-11 — Actions run 31527960279 browser-performance topology repaired

- **Implemented by:** Codex. **Recorded by:** Codex. This is an implementation record pending the
  ordinary designated cross-party review; it is not an approval.
- The public Actions check annotations identify the only failure as the Chromium Game UI
  performance assertion in `validatePerformanceObservation`; every setup step succeeded. The
  assertion previously reported no measured values, hiding which budget arm fired.
- The wall-clock-sensitive Chromium case ran inside the complete three-engine suite while
  Firefox, WebKit, and unrelated high-volume test files competed for a hosted runner's CPU. That
  topology makes shared-runner scheduling part of the long-task observation. The same code passed
  both the host browser target and the pinned Linux Playwright image, confirming the failure was
  not a deterministic formatted-commit or input-count regression.
- `make test-browser` now runs the complete functional Chromium/Firefox/WebKit matrix first and
  then runs the unchanged 200 ms/600-commit performance budget in a fresh Chromium-only process.
  `make test-browser-ci` mirrors that two-stage flow with cold anonymous dependency volumes. The
  validator now includes every observed and allowed value in its error, so a future hosted failure
  is actionable from the annotation alone. No product behavior or performance literal changed.
## 2026-08-12 — scheduled run 31562066740 repair

- **Implemented/inspected by:** Codex. **Recorded by:** Codex. Ready for designated cross-party
  review; this entry is not an approval and nothing was pushed or rerun remotely.
- The public Actions metadata identifies two failures at published `d05dc13`: `server` exits 2 on
  the already-recorded mixed relevance-golden commit, and `numeric-maintenance` exits 2 at
  `make vectors-check`. The append-only baseline correction `{759e369, 0151ddc, 3a3f4cd}` closes
  the server failure; the exact Linux/Postgres `make verify-server` job passes at this range head.
- Linux/amd64 Node 24 reproduction found three last-digit differences in approximate `pow`, `exp`,
  and `sum` expectations. The generator had serialized the full host-libm result even though both
  runtime assertions allow `1e-12` relative error. Approximate expectations now normalize to 15
  significant digits: materially tighter than the assertion while byte-identical on Darwin/arm64
  and Linux/amd64. Both hosts generated SHA-256
  `879ce1306cfd01bc7cfc493a98441280f35f7686a1b9e8801fcc7c55011a0f22`.
- Added `make vectors-check-ci` and its pinned-major Linux/amd64 Node compose service so scheduled
  numeric drift is reproducible locally. The scheduled setup-go step now names `server/go.sum`,
  closing its root-`go.mod` cache warning. Go and TypeScript shared-vector suites pass.

## 2026-08-14 — Claude designated cross-party review of `{02d00d7}` — APPROVED

- **Review by:** Claude. **Recorded by:** Claude. **Range:** `{02d00d7}` — the browser-gate
  environment repair that landed during the 8cfa00b..31a49cd Game-UI review and was explicitly
  named uncovered by that verdict. This closes that gap; the Game-UI range-union note is
  discharged.
- **Diff review:** the browser job moves off the Playwright container onto the ordinary Ubuntu
  runner with an explicit pinned `playwright install --with-deps` — the exact pattern the
  composed-browser job already used successfully, eliminating the divergence that made one job
  red and its twin green. The local compose reproduction switches to anonymous volumes + `CI=true`,
  making every invocation a cold install — this also retires the "warm caches masked the hosted
  environment" failure mode this thread's evidence rules warn about, and fixes the tracked
  `.pnpm-store` index mutation noted INFO in the prior Game-UI verdict. Docs updated in the same
  change; the planning record is honest ("pending the ordinary designated review; not an
  approval").
- **Verified in the target environment, not just locally:** the commit is published, and the
  NEWEST hosted Actions run at `c718d6a` shows `browser: success` alongside all sibling jobs —
  the job this commit exists to repair is green on the infrastructure it was red on. No product
  assertion count was weakened (19,993 with two declared skips, matching the record).
- **No findings.** APPROVED.

## 2026-08-14 — Codex repair of Actions run 31797558199 `verify-server-core` soak failure — READY FOR DESIGNATED REVIEW

- **Implemented by:** Codex. **Recorded by:** Codex. This is an implementation record, not a
  designated verdict. The supplied public job was inspected directly with `gh run view`; it ran at
  `ebcfc15` and its sole failing assertion was
  `TestFiveThousandConnectionWorldFanoutSoak`: subscriber 0 received an empty JSON application
  frame, which the test tried to decode as a world publication (`unexpected end of JSON input`).
- **Root cause:** Centrifuge's batch writer calls its transport `WriteMany` after the application's
  `OnTransportWrite` filter. When every queued stale/coalesced world item in one batch is filtered,
  the resulting JSON batch is empty and the WebSocket transport writes an empty application frame.
  The real browser runtime already splits newline-delimited protocol frames and ignores empty
  lines, keepalives, and command replies. The soak incorrectly required every WebSocket frame to
  contain exactly one publication, so it could fail under legitimate batching/coalescing even
  though the player runtime remained connected.
- **Repair:** the soak now uses a strict JSON-stream decoder matching the runtime boundary. It
  ignores only empty/whitespace frames and valid replies without a push; it decodes every
  newline-delimited push and still rejects malformed JSON, non-world pushes, private receipt kinds,
  invalid envelopes, and non-monotone/out-of-range revisions. A focused test pins empty filtered
  batches, `{}` keepalives, command replies, multi-publication frames, and the receipt-leak negative
  case.
- **Gate hardening:** `test-go-core` now passes explicit `-count=$(CORE_TEST_COUNT)`, default `1`.
  Therefore the hosted `make verify-server-core` job cannot satisfy its package tests from Go's
  result cache; callers may raise the count without bypassing the repository Make target. Canonical
  CI docs now name the actual hosted command and its cold-run contract.
- **Verification:** the original soak passed **50 consecutive Linux/amd64 runs** (250,000 total
  WebSocket connections) after the repair. The exact `make verify-server-ci` Linux/Postgres wrapper
  then ran the complete hosted `verify-server-core` command successfully. No product transport
  behavior, wire byte, balance artifact, or kernel version changed.

Nothing was pushed or rerun remotely. This test/CI batch must join the current designated-review
range before any completion claim.

## 2026-08-14 — Codex cold-harness follow-up — READY FOR DESIGNATED REVIEW

- The hosted `harness` job previously called `go test ./harness` without an explicit count, leaving
  it eligible for Go result-cache restoration even after `verify-server-core` was hardened.
- `test-harness` now passes `-count=$(HARNESS_TEST_COUNT)`, default `1`, and the Linux/Postgres
  `make verify-harness-ci` wrapper exercises that exact path. The workflow comment is reconciled to
  the shipped cheap-probe design; the manual beam is not relabelled as a CI gate.
- No production path, balance artifact, or kernel byte changed. This record does not substitute
  for the designated cross-party review.

## 2026-08-21 — amended blocking workflow hosted-green

- Push run `32518522514` executed exact HEAD `aa27705` and passed all six governed jobs.
- Hosted durations were server 2m57s, harness 1m43s, client 1m35s, browser 1m37s,
  game-ui-composed 2m08s, and schema 23s. The normative sub-five-minute blocking target is met.
- This closes the hosted push/PR acceptance gap. The separate maintenance observation and exact
  designated review union remain archival requirements, not blockers on unrelated product work.

## 2026-10-06 — R-010 / RP-218 diagnosis predeclaration

- **Work by:** Codex. **Recorded by:** Codex. Baseline `b299bbf5`, clean tree. The full Linux
  browser run's WebKit worker failure is retained in RP-218; its isolated pass is not closure.
- Research population, controls and authorization limits are predeclared in research-queue R-010.
  Only browser test observation/controls and research records are in this range. Native worker
  commands and clocks, production bytes, CI configuration, five-second assertion and browser
  population stay unchanged. No claim about the original cause until executed trace evidence.
- Designated Claude review remains required for any retained test change. This is neither an
  approval nor a current-head hosted CI claim.

## 2026-10-06 — R-010 initial experiment, control fired

- **Work by:** Codex. **Recorded by:** Codex. Native Worker observation, cleanup on witness failure
  and paired visible-return controls are test-only. Executed results and limitations are retained
  in `planning/platform-alignment/worker-prediction-research.md`.
- All nine isolated browser cases pass. First complete cold lane passes 22,068 / six intentional
  skips plus separate performance. Second is red: new WebKit conflicting-control observes nine
  refresh calls but zero native predictions in its two-second window. Original witnesses pass
  across both lanes. Do not call the control valid or claim the original failure fixed.
- This commit intentionally retains a labelled fired diagnostic pending instrument correction;
  no archived production byte or CI policy changed. No designated approval or archival.

## 2026-10-06 — R-010 instrument/fixture correction predeclaration

- Baseline `0a30f0d0`; scope and complete final population are predeclared in the R-010 dossier.
  Keep actual startup bounded at five seconds, then begin the existing two-second refresh arm.
  Native first-output latency remains visible, not absorbed into the measurement or excluded.
- Make original fixture authority consistent and require native evidence as well as visible
  change. No product source change except one temporary, exactly restored severing probe.
  No timeout, runtime, balance, CI, archive, hosted or 1.0 status promotion.

## 2026-10-06 — R-010 bounded test correction, READY FOR DESIGNATED REVIEW

- **Implemented/first-filter by:** Codex. **Recorded by:** Codex. No designated verdict.
  Dossier retains both initial populations (including the red WebKit validity control), separate
  corrective predeclaration `85ae9552`, native traces and precise limitations.
- Consistent fixture authority removes the demonstrated confounder; original cause remains
  unproven. Real first-output readiness is independently bounded at five seconds; unchanged
  two-second controls count fresh prediction after readiness. No runtime, CI or assertion-budget
  change. Suppressing actual worker publication fails all nine cases in three engines; production
  file restores byte-exactly before gates.
- Two complete cold Linux lanes pass 22,068 / six intentional skips and separate performance
  each. Root Node tests pass 7,256 / 106 deliberate DOM-only skips; type/build pass. Native first
  output ranges up to 4049 ms, explicitly visible. These are local bounded proofs, not reliability,
  current-head hosted CI, archival or 1.0 completion. Exact reviewed range must include initial
  diagnostic and correction, not only the final net test diff.
- Fresh whole-history gate passes checkout/adversarial controls, then fails the unchanged RP-131
  commit `50a3a5141d5c4c1cd25cbf3c9c39d9e3cf8e2444` against
  `0cf9f7a6aba4038fadcdf35e5b94f56986164af7`. Neither browser green nor this range closes it.
- Codex first-filter inspects the complete net span from `b299bbf5`: exactly one browser test
  plus canonical CI documentation and research/backlog/queue/roadmap records, no production or
  CI configuration diff. Native method forwarding, independent message observation, fresh-output
  controls and cleanup are inspected; executed publication severing rejects all cases. This is
  self-review, not Claude's designated verdict. CI topology and 13 negative controls, client
  shell/UI boundaries and `git diff --check` pass.

## 2026-10-06 — exact R-010 handoff pin

- **Ready for Claude designated review, not approved:** complete span
  `f4f62eac^..3c96ade8` (`b299bbf5..3c96ade8`). Initial diagnosis
  `f4f62eac^..0a30f0d0` and separately predeclared correction `85ae9552^..3c96ade8`
  union to that complete span. Inspect both the initial invalid/red control and final correction.
- This pin records implementation `3c96ade8`; it is not a verdict or archival action. Earlier
  Garden SG1/SG2/SG6/clock requests remain separate and pending. Nothing pushed.

## 2026-10-06 — RP-235 module-import diagnosis predeclaration

Baseline `6ed42d45`, clean main fourteen ahead of observed remote. R-011 full
research/pin range `39f95329..6ed42d45` remains pending Claude independently;
its 150-case primitive pass does not waive full-browser CI 54700's red result.
Firefox Route suite failed before execution, naming the dangling-resource JSON
module. 299/300 populations, 22466 pass / six existing skips, sixteen Route tests
never imported; performance not reached. Original HTTP/network cause unmeasured.

**Authority:** accepted CI Baseline D2/AC1/AC2, observation of existing declared
browser population. This new range is diagnosis/test observation and records;
no product/fixture/test assertion, dependency, browser removal, retries, timeout,
parallelism, workflow, balance/kernel, owner copy or archival change authorized.

Predeclare two complete cold Compose populations with the installed Playwright
provider's existing `VITEST_PW_DEBUG=1` request-failure observation enabled. Source
inspection of pinned `@vitest/browser-playwright` confirms it only attaches a
`requestfailed` listener before navigation; it does not intercept/retry requests.
Run unchanged default service command, all three engines and existing separate
performance command. Capture exact failing URL/resource type/error, suite/test
denominators, uncaught errors and terminal status. No production service/credentials
in this population. Logging a network cancellation is not by itself its cause.

If full cold observations cannot reproduce the import failure, retain that result
without calling the defect fixed; predeclare further targeted request/status
instrumentation rather than a configuration guess. If it recurs, inspect the named
request/provider/Vite path and add only passive HTTP/module observation as needed,
with an intentional fixture-request failure proving the observer sees a real red
suite. Do not edit while handles live, and do not replace full population with a
selector. A later correction requires its own evidence/authority predeclaration,
demonstrated failing control and full restored gates. Precise limits, designated
Claude review and historical RP-131 remain mandatory. Proper full 1.0 stays active.

## 2026-10-06 — two cold observations; next passive instrument predeclared

Both original diagnostic handles terminal. 74841 completes 300/300 populations,
22482 tests / six existing skips, 53.02 s; separate performance 525 ms / 2.49 s.
2773 is red: 299/300 populations, 22481 pass / one failed / six existing skips,
43.40 s; performance not reached. Firefox original worker witness reports
initialize/authoritative_snapshot at the same 1e3 rate, zero predictions, unchanged
output 100, five-second assertion failure and native `prediction worker failed`.
Existing provider also reports canceled worker/dependency requests. Their browser/
worker identity and timing are not sufficient to attribute those to the active
failure: passing populations also produce canceled requests during teardown.
Route's original import failure does not recur in either run; RP-235 stays open,
and this new RP-218 recurrence invalidates any local reliability closure.

Next bounded instrument, still CI D2 observation only: retain passive Vite HTTP
start/finish/premature-close observations for Route JSON fixtures and prediction
worker modules, exact UA/path/status/content-type and completed/incomplete counts,
without interception, retry or body/header secrets. Extend the existing R-010
test helper to record per-native-Worker command/output/error/termination ordering,
forwarding original methods byte-for-byte and never suppressing errors. Record
before and after cleanup; unchanged assertions, five-second bound and complete
matrix remain. No production worker/UI/fixture bytes or dependency/CI policy change.

Predeclared populations: two additional complete cold instrumented Linux lanes,
all 300 populations and separate performance on success. First demonstrate a real
test-owned module-request failure (temporary 404 for the already-named fixture),
requiring HTTP observer evidence AND failed Route suites. Restore exact instrument
bytes before populations. A failure that cannot be reproduced remains open;
do not infer a common cause from two different symptoms or grow bounds to hide it.
Further correction only after attributable evidence and separate authority.

## 2026-10-06 — module control discriminates; trace retrieval refined

Passive config/helper typecheck 11850 passes. Temporary targeted fixture 404
82623 logs exact 404/text-plain in each engine and fails all three Route imports,
no Route tests executed, 4.88 s. Observer counts 18 started / 18 finished / zero
premature/pending; runner API server has a separate truthful zero-count summary.
Config restored exact SHA `ada11b645b62ab03ebb539ec4bd49b05ff2625435f18a19f00461f6d310481a0`;
test helper SHA `a585474a5e441ed4030776b08ba686b2ee26100b1d30cc0739376b77cf5f6466`.
Full 88496 passes 300 populations / 22482 tests / six existing skips, 43.26 s,
separate performance 487 ms / 2.71 s. Observed summary 135/135 and 1/1 requests,
but tool retrieval accumulated 56,774 tokens and truncated its middle at 20,000.
That is a valid executed CI pass, NOT a retained complete per-worker/HTTP dataset.
Do not silently label the lost traces complete or use it to prove root cause.

Before further measurement, refine capture only: continuously drain the same live
handle in bounded chunks, retain all chunks, reject any indicated truncation,
then materialize only structured module/worker records plus denominators/source
hashes into a bounded research artifact. No test source or budget change for this.
Two full cold restored populations with complete captured traces are required;
88496 is a disclosed earlier execution, not a substitute for either dataset.
Additional failing control: temporary 404 of the native prediction-worker module,
selected original worker case in each engine. Require actual HTTP 404, native
Worker error lifecycle while not terminated, zero predictions, unchanged output,
and the original assertion/async guard red. Restore exact config SHA before the
two complete populations. No fixture/product/clock/assertion/engine/policy changes.
