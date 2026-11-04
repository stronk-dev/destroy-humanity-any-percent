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

## 2026-10-06 — passive observations retained; incomplete population refused

All handles terminal before this reconciliation. Native-worker 404 control 67300
fails all three selected cases / 66 selector skips, 8.11 s: actual 404, native
error while not terminated, zero predictions, unchanged output 100 and original
five-second assertion plus asynchronous guard red. Exact config/helper hashes
restored (same as previous entry), no production byte changed.

Continuously drained restored 93957 passes 300/300 / 22482 tests / six existing
skips, 49.32 s; separate performance 603 ms / 2.50 s. Capture 13 chunks, largest
13580 tokens, no truncation. HTTP 129/129 plus performance 1/1, zero pending.
Second restored 3595 does NOT complete: 299 file populations report completion,
only Firefox raw Garden-catalog population absent. Provider request to Vitest
runner fails NS_ERROR_FILE_NO_DEVICE_SPACE. Exact owned container 2bb1518c1355
has overlay zero available / 100% full, shared memory 7.8 GB free and host
workspace 197 GB free. Low current CPU/no completion is not proof of scanner
loop. Docker's 342 images / 145 volumes include substantial reclaimable storage
whose ownership is not established; no broad pruning or unrelated deletion.

Stopped only owned test after >=10m existing hosted ceiling; stop's ten-second
grace ends in exit 137, not natural completion. Capture 99 chunks / maximum
13761 tokens, no truncation, but no final denominator/performance/HTTP summary.
The predeclared two complete traced populations therefore did NOT both pass.
RP-236 records environment invalidity, independent of original RP-235 cause and
RP-218 recurrence. Do not infer that all failures share storage cause.

Dossier `platform-alignment/browser-module-research.md` and exact structured
JSONL retain HTTP/native/provider records, failing controls and capture/source
limits. Earlier 88496's truncated middle is explicitly excluded. Read-only
diagnostic output was also truncated once; complete retained collector data,
not that preview, supplies the artifact. Initial file inventory mistakenly
included screenshot directories; corrected isFile inventory is exactly 100
real files, with 100/99/100 reported. No invented suite denominator.

Review by: Codex, first-filter only. Recorded by: Codex. New designated range
starts `6ed42d45` exclusive and includes all predeclarations, instrument, artifact,
records and final pin. Claude remains required independently of every earlier
range. No test budget/skip/engine/concurrency/dependency/CI workflow or production
change, checkbox flip, acceptance, archival or push. Continue accepted API census
while Docker capacity needs a precisely scoped resolution.

Final first-filter handles 80700 and 30825 terminal: root typecheck has zero
TS/Svelte errors/warnings, shell/UI 14/8/22 boundaries and topology plus thirteen
negative controls pass; client 7366 pass / 134 existing Node skips. Independent
artifact read validates all 684 JSONL rows, each completed HTTP summary/count,
the three native failing controls and the incomplete run's missing summaries.
In-memory falsified complete/exit-zero labels are both rejected. git diff --check
passes. Initial read-only guessed Account test filenames were absent; rg --files
supplies the real names for next work. No further Docker test attempted with full
disk, and neither these fast passes nor the unchanged historical kernel guard
is reported as a green full CI.

Exact substantive new observation span: `6ed42d45..ee8b2356`, four commits /
eleven paths. This following pin-record edge also belongs to designated review;
the final handoff supplies its literal endpoint. Review by: Codex (first-filter
only). Recorded by: Codex. No Claude verdict, source correction, completed second
population or acceptance is implied. Earlier R-011/session/Garden ranges remain
independent. Docker capacity inquiry is nonblocking for non-Docker scoped work.

## 2026-10-08 — Keep Node fixtures in their actual CI runner (RP-396)

Accepted CI D2/D3; routine runner-discovery correction, not a new topology or policy.
The preceding full client88284 exited2 with two "No test suite found" errors:
Vitest collected the care/production proof files written for node:test. Their30
assertions passed under Node but could not register as Vitest suites.

Both files move byte-identically from `.test.mjs` to `.fixtures.mjs`, following
the existing refresh-observer convention; the existing test-game-ui-composed
invocation follows the names. Matching Git blobs: production
cf88a62e65d6e215816c45eb17267ba557536f96, care
028c4207869a6d558a0a42b4c9b85051a5e37293. No assertion, Vitest filter,
workflow/package, engine, timeout, retry, population or production byte changes.

Executed finished-batch checks (all handles terminal):

- 19241: whole test-client10017 pass/831 existing skips,112 passing files/23
 skipped; strict types/Svelte zero errors/warnings; build and topology with13
 negative controls pass.
- 9941: actual root test-game-ui-composed executes all30 Node checks/zero skips,
 then passes main built-client/Postgres/socket gameplay, default Cosmetic/care
 and axis-stack Clout/Cosmetic/care journeys. Standalone Node30 also passes.
- 39477: full native Chromium/WebKit functional population exits2:21635 pass,
 43 fail,10 existing skips;264 files pass/four fail. No Node-fixture collection
 errors.40 R9 host cases fail the stale `exit.disabled` pending assertion after
 focus-preserving transition handling (RP-397). Garden WebKit Enter/Space fail
 menu→Harvest Tab (RP-398). Typer WebKit Timed Enter fails keyboard provider's
 missing iframe before dispatch (RP-399). No rerun, deadline increase, assertion
 deletion or root-cause/reliability claim. Some passive trace output is truncated;
 this is executed regression evidence, not a complete module-research dataset.

Self-inspected rename/Make/docs/record diff; diff-check passes. No status or
acceptance checkbox flips, archival, history rewrite or push. Designated Claude
review remains required for the complete corrective commit after f423e704; this
is Codex implementation verification, not designated approval. RP-131 historical
guard and RP-307 arithmetic remain independently red, so neither this fix nor
the passing composed journey is a green verify-push/hosted/release claim. Next
accepted corrections are the named browser failures; author-held contracts and
the full nine-tier 1.0 floor remain unchanged.

## 2026-10-08 — Finish the current checkpoint; Firefox remains open

Under CI D2 and archived UI Foundation UF5/C7, inspected RP-451's existing native
Enter → era-change check and its plain-button fixture. No cause established and
no test, product, retry, timeout, engine or workflow change made. Root focused
command: `make test-browser-focused BROWSER_TEST_FLAGS='--project firefox test/shell-browser.test.ts'`.
Initial sandbox attempt cannot listen on localhost and runs no tests; permitted
run93099 starts Vitest, then fails to connect to the Firefox browser session
within60s, reporting no tests and one unhandled error. This is unavailable local
verification, not a reproduction or repair of the hosted era assertion. The
runner remained alive after reporting that error; interrupted only this owned
run during teardown, terminal Make exit130. No natural successful exit claimed.
The declared browser service's read-only capacity probe also reports overlay0free/
100%full; no install, cleanup or capacity workaround attempted.

Read-only GitHub reconciliation confirms37795560514 at cbd2069f is still the
latest CI run: schema/harness/composed pass; client/server/browser fail. Local
API repair and Account witness evidence stay in their owning logs, not a claim
this historical run became green. Current-state overview distinguishes those
repairs from open RP-131/RP-307/RP-450/RP-451 and the full unbuilt product scope.
This checkpoint changes records only; no acceptance, archival, release or push.
Next Firefox action: establish a usable native or declared Linux browser
environment, then capture actual focus/key/click/era behavior without replacing
native input or weakening its assertion. Safe accepted construction remains
possible independently; this verification limitation does not block all work.

## 2026-10-08 — Expose primitive input delivery without changing its oracle (RP-451)

Outcome under CI D2 / archived UI Foundation UF5/C7: the existing native Enter
and exact era_2000 assertion now retain passive before/after document/button
focus, actual key/click targets, trusted input and default prevention in the
failure message. No refocus, retry, wait, input substitution, product change or
CI-workflow change. The fixture/listeners now clean up in finally after failure.
Two native negative controls distinguish focus moved elsewhere from prevented
Enter; both keep era_1995 and produce no activation click on the expected button.

Environment diagnosis: initial minimal standalone Firefox launch13066 aborts
inside the sandbox. Permitted launch80331 hangs; verified only its own node
parent89187 and interrupted it, terminal130. Bounded stage diagnostic1113 exits1
at launch (no usable page), reporting macOS sandbox-extension denial for its
plugin-container and SWGL framebuffer failure. The diagnostic's5s launch limit
is not a changed test/acceptance timeout. No browser security setting, install,
Docker cleanup or hosted run was changed. These logs explain unavailable local
execution, not the hosted Enter assertion's cause. Local Firefox remains unproved.

Executed: initial focused run fails sandbox localhost admission before tests.
Permitted root `make test-browser-focused BROWSER_TEST_FLAGS='test/shell-browser.test.ts
--project=chromium --project=webkit'`59996 passes10/10, including existing era,
Amount, axe/focus/name checks and both diagnostic controls. Root `make typecheck`
13833 passes with zero Svelte errors/warnings. Final finished-source repeat8984
also passes10/10; all handles terminal and diff checked. No build, Postgres,
performance or whole-journey rerun for this isolated test/diagnostic change;
no full/hosted CI, Firefox or accessibility-completion claim.

Review by: Codex (implementer diff/first filter). Recorded by: Codex. Exact new
range starts after `c382a48a` through this containing test/docs/record commit;
designated cross-party review pending. RP-451 remains OPEN. Next: use a usable
Firefox environment to capture the original path with this diagnostic intact,
then repair the demonstrated cause. Other accepted construction may continue;
the Typer/Arcade public-wire question and Clout DG-D/D-021 remain unruled.

## 2026-10-08 — Revalidate RP-451 environment; continue independent construction

At `1e8a25bc`, read the actual failed hosted log for 37795560514: Firefox's native
Enter/era assertion remains the browser failure, not an Arcade failure. Current
source already contains the passive diagnostics from the preceding entry; do not
repeat that work or claim a cause from unavailable execution. The declared browser
service's fresh `df -h / /workspace /tmp` probe again reports overlay 0 free / 100% full,
while the source mount has 119 GB free. No install, test, cleanup or CI change ran
in that capacity probe. Unused build cache is 3.791 GB; narrowly scoped cleanup was
asked of the owner and remains unapproved. Images, volumes and services are intact.
Native Firefox's prior terminal session-timeout evidence is not a fresh test pass.
RP-451 remains open; its next step is the unchanged native-input test in a usable
Firefox environment. Independent accepted Combat C2 construction continues in
its own plan/log. No whole-CI, hosted repair, acceptance or archival claim.

## 2026-10-13 — Full job audit; keep composed proof oracles in Node (RP-525/RP-526)

Outcome under CI D2: the full Linux browser population exposed an actual collection
defect missed by focused selections. `opportunity-composed-proof.test.mjs` imports
the composed driver's `node:util` oracle; all three browser suites fail at import.
Exclude exactly that path, alongside the existing Node-only CSS fixture. Its 65
success/refusal cases remain mandatory in Node. No tests, assertions, player code,
copy, numeric/replay bytes, workflow topology or acceptance bounds changed.

Executed whole-job evidence at baseline `d618b0b4` and during this correction:

| Population | Actual result |
|---|---|
| `make verify-client` | Types/build pass; Node10,611 pass/1,012 existing skips. Job exits2 on historical RP-131 (`50a3a5141d5c4c1cd25cbf3c9c39d9e3cf8e2444`). Later client boundaries are not reached. |
| `make verify-server-ci` | Cold Linux-amd64/Postgres vet/package population completes; production's 27 RP-307 partition cases fail. Job exits2; later generation/boundary targets are not reached. Optional/manual prerequisites are not credited merely from package results. |
| Original `make test-browser-ci` | Three Node-only suite-import failures plus36 native-input failures;34,602 passes/24 existing skips,480.04s functional. Performance never runs. |
| Unchanged Typer/Garden files alone | 183/183 pass across the three Linux engines,10.68s; does not erase full-population faults. |
| Node exclusion + full `--maxWorkers=4` control | 423 suites collected without Node-import failures; two Firefox activation/focus failures,34,636 passes/24 skips,249.87s. Performance never runs. |
| Full `--no-file-parallelism` CLI control | All423 suites/34,638 functional cases pass,24 existing skips,290.81s; then all five performance cases pass,105 intentionally deselected cases,7.81s. This is an override control, not the ordinary job's verdict. |

The pinned browser image is **Linux ARM64**, Node24.18.0, and reports14 available
CPUs. Installed Vitest4.1.10 creates independent engine pools: default up to12
pages per engine, explicit four workers up to12 pages total. Sampled memory is
11.33GiB in the original run and3.62GiB in the serial control. These are diagnostic
samples, not a controlled benchmark: early runs overlap the server/client jobs,
so timings do not establish a hosted hardware/latency bound or each fault's cause.

A proposed config `fileParallelism:false` was then tested through **ordinary**
`make test-browser-ci`, not a substituted command. It stalled beyond the hosted
ten-minute ceiling without a terminal summary. The corrected raw-log census
reports141/141 files for Firefox/WebKit but137/141 for Chromium: missing
`axis-factor-format.test.ts`, `minigame-resolution.test.ts`, `pet-combat-inputs.test.ts`
and `ui/amount-format.test.ts`. A sandbox-denied first log pipe/directory-based
census and unavailable headless UI-RPC diagnostic are excluded, not interpreted
as product data. Actual container/process state remained live. After more than
twenty minutes, stop only owned `cloud-clicker-browser-run-784e8d68dbaa`; stop exits0,
Make returns2/Error137, and the container is absent. No other service or cache is
removed. Withdraw the concurrency config instead of committing it as verified.
RP-526 remains OPEN; the local wrapper has no equivalent of Actions' job deadline.

Finished-source verification for the retained **collection fix only**:

- `client/node_modules/.bin/vitest run --root client test/opportunity-composed-proof.test.mjs`:
  all65 cases pass,318ms, no skip.
- Declared frozen-install Linux browser selection requests both
  `test/opportunity-composed-proof.test.mjs test/generated-api.test.ts` under the
  ordinary browser config, without concurrency overrides: the Node fixture is
  not imported, all174 generated-client cases pass in Chromium/Firefox/WebKit,
  2.20s, terminal exit0. Use `compose.browser-test.yml`'s install recipe, replace
  the functional Vitest arguments with this selection and omit the unrelated
  performance command for this focused reproduction.
- `make typecheck verify-ci-topology`: zero errors/warnings; topology and all13
  existing negative controls pass. Final retained config differs from baseline
  only in the exact fixture exclusion/comment; no concurrency setting remains.
- Read-only `gh run list --workflow ci.yml --limit 3` still returns CI38222406320
  at `e0cb9187` as latest, completed/failure. Nothing here is hosted green.

Review by: Codex (implementer diff/first filter). Recorded by: Codex. Exact new
range starts **after `d618b0b4` through this containing config/docs/record commit**;
Claude's designated review remains pending, as do all earlier ranges. No feature
acceptance, archival, push or release. Next: reproduce/diagnose the four missing
Chromium file completions, enforce the existing local job deadline without raising
it, then establish a repeatable ordinary full job. Do not retry until a last green
conceals these failures; history/accrual contracts remain separate prerequisites.

## 2026-10-13 — Bound the local browser process; expose the missing-module boundary

Outcome under CI D2: `compose.browser-test.yml` now wraps the frozen install and
both unchanged browser lanes in `timeout --verbose --signal=KILL 10m`, matching
the existing hosted containment ceiling, not raising individual test/performance
or sub-five-minute acceptance bounds. Kill targets the disposable process group;
stderr distinguishes a fired deadline from other exit137 causes. This cannot
bound an unavailable Docker control plane before the container starts.

Reuse the existing passive Vite observer for the four RP-526 module URLs; no
request interception, retry, test exclusion, concurrency or player behavior change.
The raw trace differentiates HTTP loading from later test execution/reporting.

Executed after `e2efdebf`:

- Read the actual Compose-rendered argv and reuse its wrapper in the declared
  service. Healthy child exits0; deliberate child failure preserves23; a
  TERM-ignoring sleeping child with only the diagnostic duration shortened to
  0.2s exits137 with `timeout: sending signal KILL to command`. Normal configured duration
  stays10m. All controls finish; no host daemon/service is stopped by them.
- Frozen-install four-file selection (`axis-factor-format`, `minigame-resolution`,
  `pet-combat-inputs`, `ui/amount-format`) passes all123 cases in three engines,
  3.00s. Observer records12 starts/12 successful finishes, pending0, valid=true.
  These isolated cases do not dispose the previous full-job stall.
- Ordinary `make test-browser-ci` completes functional collection:423 files,
  34,623 passes/15 failures/24 existing skips,316.27s, terminal Make2. Performance
  never runs after red functional results. Chromium fetches all four suspect
  modules with200/text/javascript; no recurrence of the four-file stall in this
  run. Keep the red native-input result, not a full-CI or latency pass.
- Typer's whole native journey times out; subsequent composition checks receive
  `ls -lap` and `ls -lal` instead of `ls -la`. Post-timeout keyboard continuation
  is a concrete next hypothesis, not a proven product/cancellation diagnosis.
- `make typecheck verify-ci-topology`: zero errors/warnings; topology plus all13
  existing negative controls pass. Diff checked; no test/assertion bytes changed.

Then Docker stops answering stats/listing and a proposed four-CPU diagnostic.
Read-only host `df -h .` reports1.8GiB available/100%; do not infer its cause.
The CPU-quota proposal yields no result and is withdrawn. Exact task-owned client
commands/PIDs are verified via process metadata before terminating8132,10158,
17584/17586 and18774; these are probes/Compose clients, not Docker Desktop or
game services. Probe sessions terminate130/143. Server-side probe/container state
cannot be verified while the daemon is unavailable; no cleanup-success claim.
The first sandbox-denied process inspection is excluded, then repeated with
read-only approval. No daemon restart, image/volume deletion or broad cleanup.

Scoped disk inventory: `.cache/go-build`2.8GiB is regenerable compiler cache;
`.cache` also contains release evidence that must NOT be swept with it. Approval
to clear only that compiler cache is requested and remains ungranted at this
checkpoint. No cache bytes are removed. Release artifacts and research stay intact.

Review by: Codex (implementer diff/first filter). Recorded by: Codex. New range
starts **after `e2efdebf` through this containing Compose/observer/docs/record commit**;
designated Claude review remains pending and may consolidate with the prior CI
range after `d618b0b4`. No archive, publication or release. Resume Docker-based
diagnosis after capacity/control-plane recovery; investigate late input and the
original unfinished-file condition without raising bounds or retrying away red.

## 2026-10-13 — Stop expired Typer witness input contaminating its successor (RP-527)

Outcome under CI D2/evidence discipline: reproduce the actual framework lifetime
boundary before changing the driver. Vitest aborts the test signal on timeout but
does not cancel its async function; native RPC loops can continue. A deliberately
held producer fires the real50ms diagnostic deadline. The successor independently
checks that exact reason, focuses a real input and releases the old producer.
Old behavior writes `OLD` and completes instead of refusing: both Chromium/WebKit
fail the exact `{outcome:"refused",value:""}` assertion. Startup listen EPERM ran
no tests; repeated the root target with its narrowly approved localhost listener.

Correction is test-only: lifetime-bound keyboard actions preserve real native
delivery and fail on cancellation/closed or overlapping use. Track in-flight RPC;
cleanup drains it, and the next test fails before mounting if prior input has not
drained. The whole Typer journey uses that guard and idempotent finally/finish
cleanup, including its engine operation. No original assertion, input, prompt,
score/terminal outcome,15s journey deadline or hook timeout is removed/raised.
No product, kernel, schema, copy, CI topology or concurrency change.

Executed finished-source results:

- Focused timeout-isolation controls11610: native Chromium/WebKit, six positive
  passes/two deliberately expected timeout failures,1.63s. Actual expired-producer
  refusal leaves the successor empty. Active delivery completes into the old
  input before drain; another action cannot affect the successor. Pending-input
  refusal and callback failure/healthy completion also execute. The expected
  deadlines are negative controls, not independent acceptance passes.
- Entire affected file28203 through root `make test-browser-focused
  BROWSER_TEST_FLAGS='test/typer-table-browser.test.ts --project=chromium --project=webkit'`:
  54 passes/two expected diagnostic failures, no deselection,2.48s. All original24
  cases execute per engine, including the full real-pure-engine untimed keyboard
  journey, native focus, composition, terminal exit and axe checks.
- Final-source repeat30564 after adding cancellation checks before late mounting
  and direct focus: the same entire native population passes54/two expected
  diagnostic failures,2.46s. Final strict types63969 again report zero errors/
  warnings. No native input replacement or revised assertion.
- Root Node selector33704: one helper case passes/27 browser cases explicitly
  skip,581ms. This is not credited as native browser coverage.
- Root `make typecheck`57082: zero Svelte errors/warnings. Diff checked; no runtime
  build, Go, API, performance or database suite needed for this test-only change.

Current read-only checks: host free space is2.6GiB/100%; no cache deletion approval
arrived and no bytes are removed. Docker version probe15896 remains unresponsive;
cancel only its own TTY client with Ctrl-C, terminal130. Do not restart Docker or
substitute native-only checks for Linux/Firefox/ordinary full-CI acceptance.
The source regression proves the cancellation/isolation mechanism, not attribution
of every previous15 full-job failure or a repair of their workload timeouts.

Review by: Codex (implementer diff/first filter). Recorded by: Codex. Exact new
range starts **after `71682ad0` through this containing test/docs/record commit**;
designated Claude review pending, consolidatable with the preceding CI ranges.
RP-526/full CI, historical guard/accrual contracts and whole-product1.0 remain
open. Next: validate the guarded driver in the actual Linux full population after
capacity/daemon recovery, retaining failures; diagnose other input producers and
the unfinished-file condition rather than assuming they share this correction.

## 2026-10-13 — Await Garden's actual startup boundary and own its teardown (RP-528)

Under CI D2, investigate the full-job `urls=[]` failure at the fixture's initial
socket check. Existing helper waits five event-loop turns before asserting a
constructor dependent on asynchronous snapshot read/parse/bind. A held first
snapshot reproduces its premature rejection in both native engines92994: the
helper settles before release where the regression requires it to remain pending.
Keep that old failure, not attribution of every overloaded full-job result.

Test-only correction: wait for the actual socket factory; retain exact URL and
sent protocol frames, including the not-ready and recovering/resync cases. No
fake readiness, reactivation, input substitution, retries or larger deadline.
Use explicit caller `TestContext`, refuse canceled initialization and own one
idempotent cleanup on failure/finish. That cleanup rejects an unfinished startup
wait and unmounts the owned target; release/close assertions remain. All original
thirty test populations and their conditions/assertions stay intact.

Executed:

- Controlled delayed-read regression99195 passes2 native cases,1.91s. Before
  release it remains pending; after the real constructor it remains unready until
  the actual injected protocol handshake. No intent is manufactured.
- Startup-disposal/held-read controls61084 pass4 positive cases and two deliberately
  expected50ms framework deadlines,1.89s. Successor independently checks the exact
  timeout reason, disconnected owned target and explicit canceled-startup result.
  Expected deadlines are negative controls, not player acceptance passes.
- Initial complete corrected file41339 passes64/two expected deadlines,10.35s.
  Node88461 collects/skips all33 browser cases,0 executions: no positive evidence.
- Making cleanup ownership explicit briefly introduced a callback parameter
  `context` that shadowed the founder/constants loop variable. Strict types80309
  reject TS2367. Although native candidate96170 passes, that pass is EXCLUDED:
  its loop populations are wrong. Rename the callback/fixture parameter to
  `testContext`, preserving the domain loop and its real branches.
- Finished-source entire file75358: root `make test-browser-focused
  BROWSER_TEST_FLAGS='test/garden-host-browser.test.ts --project=chromium --project=webkit'`
  passes64/two expected diagnostic deadlines, no deselection,10.54s. Exact HTTP
  methods/auth, actual runtime/frame decoder, native recovery/intent/focus and
  all original lifecycle populations execute; transport/network and DB are fixture
  boundaries, not a real gameserver/Postgres journey. Strict types60382 pass with
  zero errors/warnings. Diff checked; no shipped build/Go/schema/performance change.

Read-only capacity check reports2.7GiB free/100%; cleanup permission has not arrived
and no cache/release/research byte is removed. Docker version79368 remains
unresponsive; interrupt only its own TTY client, terminal130. No daemon restart,
service/container cleanup or Linux/Firefox/full-CI substitute claim.

Review by: Codex (implementer diff/first filter). Recorded by: Codex. Range starts
**after `a181b935` through this containing test/docs/record commit**, designated
Claude review pending; recent related CI repairs may be consolidated. No feature
acceptance/archival or publication. Next: execute the corrected Garden/Typer drivers
in the declared Linux full population after recovery, retain any failures, and
separate genuine readiness/input faults from invalid fixture timing assumptions.

## 2026-10-13 — Isolate Reputation plan input after a test deadline (RP-530)

Under CI D2, extend the proved input-lifetime correction to the Reputation plan
fixture, preserving all original18 cases, native Tab/Shift-Tab/Enter/Space inputs,
budget/cascade/focus/callback/axe assertions and existing deadlines. Reuse Typer's
unchanged guard in a shared test-only helper rather than duplicate its implementation.
Bind every plan fixture to its actual TestContext, drain native RPC before
idempotent finish/finally disposal, and refuse expired selection/disclosure helpers
before DOM lookup/focus. Recheck cancellation after loading native input tooling.
No production, schema, kernel, authored copy, concurrency or budget change.

Regression: initial99229 resumes `choose` after disposal and fails on a stale-DOM
TypeError in both engines. That is not proof of keyboard contamination. Corrected
diagnostic58625 holds the actual native keyboard producer across a real50ms test
deadline, then releases it with a successor input focused. Unguarded delivery
completes and inserts a space in both native engines, failing the exact refused/
empty assertion. This proves the lifetime mechanism, not attribution of every
previous RP-526 failure. Retain the deliberately fired deadline as a negative
control; its successor independently checks the exact reason, canceled helper
calls, disconnected fixture, retained successor focus and empty input.

Executed final-source root `make test-browser-focused
BROWSER_TEST_FLAGS='test/reputation-plan-browser.test.ts test/typer-table-browser.test.ts
--project=chromium --project=webkit'`63018:92 passes/four expected diagnostic
timeouts, all96 cases/no deselection,2.66s. Earlier whole affected33852/18735 also
pass; final guards are covered by63018. Typer retains all original24 cases and
its existing drain/failure/healthy controls. `make typecheck`4835:zero errors/
warnings. Node60089 executes one helper test and explicitly skips47 browser
cases; no native acceptance credit. Diff checked; no shipped build/Go/API/DB or
performance boundary changed, so those suites are not rerun for this test fix.

Read-only Docker version17832 remains live/unresponsive on re-poll; stop only
that TTY client, terminal130. Host reports2.7GiB free/100%. No cleanup answer
arrived, no cache/evidence/container/service removed and no daemon restarted.
Linux/Firefox/ordinary full job and hosted CI remain unverified/red; RP-526 is
not closed and the Reputation host driver is not claimed repaired by this batch.

Review by: Codex (implementer first filter). Recorded by: Codex. Exact new range
starts **after `dcb561c5` through this containing test/docs/record commit**;
designated Claude review pending, consolidatable with related CI driver ranges.
Next: diagnose the host driver's lifetime/focus boundary and execute all corrected
drivers in the actual Linux full population once the declared service is usable.
Keep the whole nine-tier1.0 goal, existing contracts and release/review gates open.

## 2026-10-13 — Dispose Reputation host fixtures at the original test boundary (RP-531)

CI D2 outcome: a timed-out suspended host producer must not leave its mounted
UI/runtime/socket alive or navigate after cancellation. The original isolated
host file6189 passes176 native Chromium/WebKit checks,25.30s; that does not test
its timeout teardown. New real50ms framework-deadline regression75228 fails in
both engines because the old target is still connected. Combined observation
14129 confirms `{mounted:true,socketClosed:false}` instead of the required
detached/closed state. Release the held continuation in the diagnostic's finally
so even the old-failure run cleans up its owned subject.

Test-only correction: centralize the existing mount/cleanup in a small fixture
bound to the actual TestContext. One idempotent dispose runs in caller finally
and onTestFinished, drains the existing native-keyboard guard, releases controlled
network promises and unmounts/removes the owned UI. Expired navigation refuses
before polling/DOM focus; native input also rechecks its signal after tooling
loads. Every original88 host population and its native inputs, exact request/
receipt, rejection, pending/focus, subscription/recovery and remount checks remain.
No product, API, schema, copy, recovery policy, timeout or concurrency change.

Executed corrected controls39399:two positive native checks/two deliberate
deadline failures,1.69s; the successor independently observes the exact timeout
reason, detached target, closed controlled socket, canceled continuation and no
intents. Entire affected root `make test-browser-focused
BROWSER_TEST_FLAGS='test/reputation-host-browser.test.ts test/reputation-plan-browser.test.ts
test/typer-table-browser.test.ts --project=chromium --project=webkit'`21146:
270 passes/six expected diagnostic deadlines, all276 cases/no deselection,25.94s.
`make typecheck`80051:zero errors/warnings. Node56442 collects/skips all90 host
cases with no executed test; no browser credit. Diff checked. These use actual
native browser/host/runtime/parser with controlled network/storage, not a real
gameserver/Postgres/minted-content journey. No shipped build/Go/API/performance
change warrants those unrelated suites. A controlled cleanup failure does not
attribute every RP-526 focus/activation failure to this fixture.

Docker version72574 stays live/unresponsive on re-poll; cancel only that owned
read-only TTY client, terminal130. Host has2.6GiB free/100%; no cleanup approval,
cache/evidence deletion, container/service cleanup or daemon restart. Linux/
Firefox/ordinary full job and hosted CI remain unavailable/unverified/red.

Review by: Codex (implementer diff/first filter). Recorded by: Codex. Exact new
range starts **after `6a6a996d` through this containing test/docs/record commit**;
designated Claude review pending, consolidatable with related CI repairs.
Next: execute corrected drivers in the actual Linux population when usable and
retain remaining input/runner failures rather than declare RP-526 closed. Other
accepted product work and technical draft closure can proceed without Docker;
full nine-tier1.0 and review/release gates remain open.

## 2026-10-13 — Reconcile finished hosted CI and reproduce both blockers locally

Previous goal turn was progress: original GS0.1 disposition and the actual hoard
overflow refusal at61b145af. Read the confirmed live CI38381138075 handle rather
than restart when logs initially refuse access. It is now terminal failure on
published4eaed0c9; schema/harness/browser/composed pass, client/server fail.

Finished client job115241567501 executes10,627 Node passes/1,022 reported skips,
then the full-history gate rejects50a3a514 against0cf9f7a6, naming all six watched
minigame UI/adapter paths. The server job115241567727 fails exactly27 original
Clout partition subcases; the first reports1.00035846811e4 versus1.00035846812e4
with different full encoded states. No fixture weakening or ambiguity about cause.

Fresh local existing commands at61b145af reproduce both, terminal Make exit2:
92226 `make verify-kernel-version` rejects the exact same historical SHA/parent/
paths after its checkout/adversarial fixtures pass.79061 `make test-go
GO_PACKAGES='./production' GO_TEST_FLAGS='-count=1 -run TestAxisAccrualPartitionProperty -v'`
fails the same27 seeded cases. These are real known local failures, not a Linux-
only defect or a newly green tree. No DB operation claimed by the selected test.

Successful browser job115241568120 reports426 files/34,704 passes/twelve expected
deadline diagnostics/24 skips,305.18s, across Chromium/Firefox/WebKit. Isolated
performance reports5 passes/105 explicitly deselected,10.57s. Populated observation
exclusions remain simulated duration, manual4x profile, AT, SQL/default players,
other runtime ports and later tiers; no whole-AC7 claim. HTTP summary shows
1,998 finished/started requests and0 pending on this completed run. Browser logs
are very large and returned excerpt is truncated (539,996 estimated original
tokens); retain exact terminal job/status and complete final census, not a claim
to inspect every log line. Browser job6m38s, server6m40s and composed5m37s exceed
the strict five-minute workflow floor; do not relabel budget failure as green.

This supplies hosted ordinary three-engine coverage for the recent corrections,
not general reliability, attribution of every RP-526 local stall/input failure,
independent review or a release. Current local61b145af's hoard fix is not in the
published run. No code, timeout, retry, fixture, version, CI topology or publication
changes in this checkpoint. Reconcile the living plan/overview rather than keep
its obsolete pending-job list; older evidence stays append-only in the logs.

Owner question presented: adopt the draft exact-path/commit, independently proved
historical-scope exception policy without approving any candidate or waiving
review. No answer is recorded, so no implementation/exception row follows.
Conservation's accounting/origin contract and arithmetic remain separate work;
approval cannot repair the27 failures. Docker restart question is still unanswered,
no restart/cleanup performed. Next: explicit policy/contract closure, independent
historical candidate audit and measured latency work; accepted lanes continue.
Review by: Codex (executed evidence reconciliation). Recorded by: Codex.
Not designated implementation approval,
feature closeout or shrinking the complete nine-tier1.0 goal.

## 2026-10-13 — RP-131 distinguish completed correction evidence from requests

Outcome under existing KV-1/correction governance and independent-review rules:
metadata label names in a request/example cannot stand in for a completed verdict.
Old actual CLI42457 accepts all8 initial non-verdict controls. Reuse its isolated
Git-history fixture and replace only the review log, with healthy history before/
after; no new runner, framework or real-repository rewrite. Corrected55368 rejects
all12 instruction/fence/quote/comment/empty/pending/self-filter/conflict/indented/
nested-example cases at the specific review-binding failure. Completed approved,
negative and mechanical-remap records remain valid; the full existing miss/remap/
immutability/ancestry/archive/shallow/merge populations pass. Extra four negatives
are final controls, not a claim they were run in the initial old-tool baseline.

Recognize visible metadata paragraphs in the same heading section as the unchanged
exact offending anchor. Ignore example/code/comment text, reject missing/duplicate/
pending decisions and explicitly labelled first filters. Keep version bumps,
correction schemas/rows, archive fallback, ancestry and append-only rules unchanged.
This parses declared evidence; it does not authenticate execution or make any
historical-path exception proposal accepted.

Actual root initially refuses `09d5027`: its exact anchor cites the implementer
first filter. Source inspection finds Claude's completed2026-09-24 designated
review of `3f58fea..65099e7`; actual `git rev-list` contains09d5027 and the two
surrounding commits. Append only a binding transcription in Deployment's log,
retaining Claude's attributed subagent inspection/parent recording and original
CHANGES REQUIRED decision. Codex is only recorder, never a new reviewer. Neither
the original verdict nor correction row is rewritten/promoted. Actual94658 then
passes existing correction bindings and fails at the original `50a3a514` miss.

Both tool syntax checks and `make verify-ci-topology` pass with its13 negative
controls; all named handles terminal and final diff checked. Kernel0.3.188,
runtime, correction/path registries, workflows, job populations and bounds stay
byte-unchanged. No Go/browser/DB/hosted or whole-CI pass is claimed for this CLI-only
change; existing RP-307 and five-minute failures remain. RP-540 records the source-
observed noncanonical review-path admission for the next isolated CLI experiment,
not an executed external-file read or permission to touch secrets.

Review by: Codex (implementer first filter). Recorded by: Codex. Full range after
`a7e33b73` through this containing tools/tests/docs/binding/record commit needs
designated Claude review. The transcribed old DP-E verdict approves none of it.
Continue safe accepted engineering and pending policy/contract closure; no push,
archival, release or reduction of the full nine-tier1.0 objective.

## 2026-10-13 — RP-540 canonical correction citations before lookup

Outcome under existing repository-citation authority: prefix/suffix matching is
not enough to admit a review path. The actual old CLI79706 admits four dot,
parent, empty-segment and backslash aliases, using genuine new-correction rows
before commit. Healthy canonical review/version controls pass before/after; the
later immutable-row comparison therefore cannot mask this admission experiment.
Refuse those segments, backslashes and NUL before filesystem lookup. Preserve
canonical active/archive paths, normal archive fallback and all immutable rows.

Retained historical tool611273b5 executes in the same isolated Git fixture with
`planning/../../external-review/log.md`: a valid sibling review passes, but
replacing that sibling's content removes binding. The corrected tool refuses
the path at entry validation regardless of its valid content. Both repo and
sibling live beneath one newly created, owned temporary directory; no host-private
file is read. Only the fixture's guard copy is switched to historical bytes;
real source/history is never rewritten. This is an exterior-citation proof,
not a live correction exploit, symlink/inode guarantee or review authentication.

Final actual fixture1448 passes aliases/exterior/NUL refusals, both canonical
controls and the full prior review/miss/remap/archive/ancestry/merge population.
Earlier23258 covers aliases/NUL/archive but precedes the final exterior control,
so it is not the full final result. Both syntax checks and topology pass; actual
root32541 still fails50a3a514 after its checkout controls pass. All handles terminal,
diff checked. Current rows/path registry, kernel0.3.188, runtime, workflows and
bounds stay byte-unchanged. No Go/browser/DB/hosted or whole-CI pass is claimed.

Review by: Codex (implementer first filter). Recorded by: Codex. Full corrective
range after `611273b5` through this containing tools/tests/docs/record commit
needs designated Claude review, consolidatable with RP-131 after `a7e33b73`.
No history-exception adoption, new correction, false bump, archival or push.
Next: genuine accepted engineering and pending contract/policy closure, without
repeating these completed fixture populations or reducing the full1.0 goal.

## 2026-10-13 — Current-client full population exposes another RP-526 deadline

At `b8eca33c`, `make test-client`19155 passes10,817 with1,038 explicitly skipped.
Complete native functional `make test-browser-focused BROWSER_TEST_FLAGS='--project=chromium
--project=webkit'`76642 fails Typer's whole-game native journey at its unchanged15s deadline:
23,547 pass, one failure, eight expected diagnostic failures and16 skips;165.14s, Make exit2.
The runner subsequently terminates; no timeout is treated as process death. Its failure
screenshot is blank after fixture disposal, so neither the blocked stage nor latency cause is
known. Large passive HTTP output is truncated by tool delivery; the final verdict/counts and
specific failure are retained, not claims of a complete raw first-run recording.

Add only passive test observation before deadline teardown: stage/elapsed time, engine phase,
prompt/revision/command count, pending/input-length/focus and native-RPC flight state. Preserve
every original key sequence, operation, assertion, timeout, cleanup/drain and concurrency setting.
The existing real50ms deadline control now requires the observer to sample a still-connected
input; healthy/disposed controls require silence. It records no typed text or credentials.
Affected file68457 passes54 plus two explicitly expected deadlines in Chromium/WebKit; types
45861 pass with zero diagnostics. Node executes its one pure helper test and skips27 browser
cases, not player evidence.

One instrumented complete diagnostic population79633 terminates exit0:286 files/23,548 pass,
eight expected deadlines/16 skips,147.22s. No actual whole-game deadline occurs, so it yields
no original-failure phase attribution and is NOT a latency/reliability repair or erased red run.
Full output remains ignored at `.cache/typer-native-observation.log`, SHA-256
`0de5759bbd572cc31c9c221cd67e42856e1c228f5e3459daa890a0bda7751251`.
HTTP observation finishes1,361/1,361 requests, none pending; that is its module population,
not all HTTP, SQL, player journeys or a performance measurement.

Actual remaining client-guard union7037 uses `make -k` to retain all outcomes: history still
fails50a3a514, while topology/13 negatives, combat/meter/achievement/cosmetic/no-payment
boundaries and copy/content generation pass. Aggregate exit2 remains failure. No Firefox,
Linux/SQL/composed, isolated performance, current hosted run or `verify-push` pass is claimed.
Existing RP-307, strict hosted latency and full nine-tier release requirements remain open.

Review by: Codex (test implementer first filter). Recorded by: Codex. New diagnostic
tests/docs/records range starts after `b8eca33c` through this containing commit and needs Claude
designated review, consolidatable with the existing native-input containment corrections.
All handles terminal; diff/whitespace checked. No product, schema, kernel, balance, CI workflow,
acceptance, archival, publication or deployment change. Next: retain the phase capture on an
actual recurrence; continue authorized construction/review without claiming the timeout solved.

## 2026-10-13 — Fresh current-source blockers and published-state reconciliation

At clean local `a713137e`, main has28 commits after cached origin/main `4eaed0c9`; no agent push.
Read-only `gh run list` and `gh run view` still return completed failed CI38381138075 on that
published source. Client/server fail; schema/browser/composed/harness pass. Observed job times:
server6m40s, browser6m38s, composed5m37s; all three exceed the existing five-minute floor. This
refreshes metadata for the same run, not a new hosted execution or new cause-log analysis.

Fresh root `make test-go GO_PACKAGES='./production' GO_TEST_FLAGS='-count=1'`8341 exits2:
the full package executes37.942s and TestAxisAccrualPartitionProperty reports27 divergent encoded
state subcases (RP-307). Root `make verify-kernel-version`46629 exits2 after both checkout
controls pass, rejecting historical50a3a514 (RP-131). No bound/history/source/version alteration.
The two failures are still locally reproducible; this is not `verify-push` or full server/CI proof.

Direct read-only Docker socket curl53483 is bounded to5s and exits28 after5006ms with zero bytes.
Fresh `df` reports124GiB available/86% used. Do not rerun a hanging integration target, restart or
clean up the daemon without authority, or substitute host Postgres. RP-551's declared integration
extension remains compiled but unexecuted. No Docker service/process mutation or secret read.

Updated the existing whole-product board and stale October8 queue head, preserving historical
evidence and feature-owned review ranges. The recent Account/Combat/Deployment repairs improve
foundations and falsifiers, not another tier or an accepted release. Review by: Codex (editorial
fact-check, not designated implementation approval). Recorded by: Codex. No new specification,
code, acceptance/archival waiver or publication authority. All named handles are terminal.
Next: native RP-526 first-failure diagnosis/remaining accepted reviews and construction; explicit
held-contract closure, then declared Linux/Postgres and exact-artifact proof when authorized.

## 2026-10-13 — RP-553 capture Typer deadlines before fixture initialization

At `ed3d802e`, unchanged focused Typer30417 passes54 plus two expected deadline controls.
One unchanged full native Chromium/WebKit population39990 passes288 files/23,704 checks,
eight expected deadlines/16 skips in152.46s; no RP-526 recurrence or cause attribution.
Its final observed module summaries have0/0 and1,364/1,364 terminal requests. Tool delivery
truncates the large passive output; these are observed final counts/exit, not a complete transcript.

Source exposes a distinct diagnostic gap: the observer is armed after async content identity,
engine creation and mount. Reuse the actual journey as a function with a test-only optional
pre-setup gate; its normal invocation has no gate or added await. Controlled50ms framework
deadline/successor tests require a phase record and the exact expired-continuation refusal.
Initial58233 and refined21895 both fail the two native observation assertions with zero records;
54 existing checks and four expected deadlines remain. Refinement puts the controlled barrier
before driver import so dependency loading cannot delay barrier creation. It is not a claim that
the original15s failure happened there, and the negative deadlines are not player acceptance.

Arm the same observer before the first await, record driver/content/engine/mount phases, and
use safe `fixture_ready:false` context until the existing fixture is assigned. Register disposal
at entry as well as ordinary teardown. Existing mounted-input sampling, native-input drain,
seed/content, server-time fixtures, gameplay sequence/assertions,15s budget and concurrency stay.
Only diagnostic/test code changes; no component, engine, wire, balance, kernel or CI topology.

Corrected focused47474 passes56 plus four expected deadlines in3.01s. Typecheck10214 reports
zero errors/warnings. Node executes one pure helper check and skips29 browser cases, not native
evidence. Final unchanged full native population39806 passes288 files/23,706 checks, ten expected
deadlines/16 skips in149.29s; observed module summaries0/0 and1,362/1,362 have no pending requests.
Its raw passive output is also tool-truncated. No ordinary deadline fires, so RP-526 remains
unattributed/open; these passes are neither a reliability fix nor erased older failures.

Review by: Codex (diagnostic implementer first filter). Recorded by: Codex. Exact test/docs/
records range starts after `ed3d802e` through this containing commit; designated Claude review
pending, consolidatable with the earlier native-input/phase-observer range. Formatting/diff
checked; all named handles terminal. No full `verify-push`, Firefox/Linux, performance, SQL,
composed/player or release claim; known RP-131/RP-307, environment and full nine-tier1.0 gates
remain. No push/deploy, timeout/retry changes or production content. Retain the complete phase
capture on a real recurrence and continue accepted construction/review, not reruns for green.

## 2026-10-14 — RP-526 real native deadline localization and prior-stage timings

Resume clean `213ad5f6`; the preceding goal turn made progress with a measured representation
proposal, not gameplay. Continue the live queue's CI D2 native-input diagnosis. One unchanged
complete Chromium/WebKit functional population26841 exits2:292 files,24,427 passes, one real
WebKit Typer15s deadline, ten explicitly expected deadline controls and16 skips,175.54s. All
selected functional files finish. Passive module records finish1,359/1,359 with none pending;
the separate Chromium summary is0/0. No Firefox, isolated performance, SQL/composed or whole CI.

Failure-only capture before teardown: `submit-prompt-4`, elapsed15001ms/current-stage1040ms,
typing/prompt4/revision7/commands6, last command submit_line; input length12, focused input true,
fixture connected true, pending false, native-input RPC outstanding. This localizes the new
failure beyond initialization and leaves roughly13,961ms in earlier stages; it does not prove
which earlier phase consumed that time, that the native provider itself caused it, or that any
particular older failure shares this cause. No typed text/credentials are recorded. Actual
installed provider source dispatches the sequence through serial key down/up/insert calls;
source alone is not a latency attribution or permission to substitute fill/synthetic events.

Retain complete ignored output `.cache/native-functional-213ad5f6.log`, SHA256
`e4c8b6e7cba6f8bb643f0f2a91875c7a530fad27b3062f5f952fd96322a6c75d`.
Read-only process inspection distinguishes this owned Make/Vitest tree from a concurrently
running sibling-repository browser job. No process is killed or reconfigured, and the machine
is not represented as isolated. The owned handle terminates normally; no observation timeout
is treated as process death. A fresh bounded Docker socket ping88587 independently exits28
after5007ms/zero bytes; no restart authority, service mutation or substitute DB is assumed.

Extend only the existing failure observer with every completed stage's duration in order.
History has no truncation and is emitted only with the cancellation capture; normal input/callbacks/assertions,
fixture content, concurrency and ordinary15s budget remain unchanged. Existing real50ms setup
and connected-input deadlines now require that ordered history and nonnegative finite timing.
Old observer99061 fails all four new native observation assertions (52 existing checks pass,
four expected deadlines); corrected22099 passes56 plus those same four expected deadlines,
2.96s. Healthy/disposed observations stay silent. Types95221 have zero errors/warnings; Node
80540 executes one pure helper check and explicitly skips29 browser cases, not player evidence.

The full ordinary red run remains red; no subsequent full-run green replaces it. This is a
diagnostic correction, not a repaired input latency, new acceptance bound or1.0 milestone.
Diff checked; all named handles terminal. Docs/owning plan/RP-526 now carry the actionable
boundary. Review by: Codex (test implementer first filter). Recorded by: Codex. Exact complete
diagnostic/tests/docs/records range after `213ad5f6` through this containing commit needs Claude
designated review, consolidatable with the existing native-lifetime/observer range. No product,
schema, kernel, balance, copy, CI workflow, archival, push or deployment change. Next: use the
completed-stage capture on recurrence to distinguish setup, native dispatch and engine/render
cost before proposing any optimization; other accepted construction continues independently.

## 2026-10-14 — RP-526 accumulated input waits and overlapping callback measurement

Resume clean `3c6d5acc`; the preceding goal turn made progress with actual deadline capture and
a discriminating stage observer. One complete unchanged native Chromium/WebKit population91456
again exits2:292 files/24,427 passes/one WebKit Typer15s deadline/ten expected deadlines/16 skips,
157.90s. All selected files finish; passive module summaries0/0 and1,363/1,363 have none pending.
The real capture at prompt6 is retained verbatim in `typer-deadline-3c6d5acc.json`, verified
against both equal console emissions and the original raw digest. Complete ignored raw record:
`.cache/native-stage-history-3c6d5acc.log`, SHA256
`ee0d3b579e9e478853347f7231ba5d22b52593123594b20dffa626a84e8249d6`.

All15,000ms are accounted: completed13,337 plus active1,663. Initialization/initial render4ms;
completed input-command promise intervals13,332ms, including submit3 at4,294 and submit4 at3,480.
The input stays focused/connected, pending false, revision9/commands8 and an outstanding native
RPC. These intervals INCLUDE callbacks triggered by Enter: the pure-engine/snapshot work can
finish before the native promise resolves, so later0ms response waits do not prove0ms engine cost.
This changes the next measurement, not the attribution. Machine isolation, CPU cost, current
hosted CI and the cause of every earlier failure remain unproved. No dependency/source optimization.

Extend the existing fixture's failure-only context with completed command-kind/duration records
and any active callback phase. Measure pending-state render, input/engine call+wait and result
decode/render inside the actual callback; these are elapsed intervals including scheduling, not
isolated CPU. Native input, seed, snapshot semantics, assertions and ordinary15s deadline remain.
The ordinary journey now also requires ten records. Old producer84543 completes actual gameplay
then fails both new record assertions (54 other checks/four expected deadlines pass). Corrected
89762 passes56/four expected deadlines; type7277 and pure-helper/29-skip Node33215 pass.

One full instrumented population73374 exits0:292 files/24,428 passes/ten expected deadlines/16
skips,156.21s, with no ordinary failure capture. Module summaries0/0 and1,362/1,362 complete.
Retain `.cache/native-callback-timings-3c6d5acc.log`, SHA256
`cc1806ccecf3f4af4da5369661bbe9f727419bc41c1ebf723dd48f4489de7adf`.
At that execution, the only source delta is test producer SHA256
`b0fc6351cd93bef5a199e23ba9414954475f8c01006447f41d0da0f06050b6ee`.
This pass yields no callback profile on recurrence, is not a latency repair, and does not erase
the two preceding real full-run failures. No further full rerun for a green status.

Strengthen the new measurement with one diagnostic-only apply seam: normal calls still use the
real engine directly; the control adds an asynchronous hold then invokes that same engine and
all original native gameplay. Independently measure the hold and require each recorded apply
interval to include it, without assuming25ms timers fire exactly on schedule. Unconnected hook
33095 fails both controls with zero invocations. Connected full-file32562 passes58/four expected
deadlines; types43176 are clean and Node96869 executes one helper/30 explicit browser skips.
Because this timing could inform a future operating change, explicitly force reported apply
durations to0 once:96575 fails both native controls (held26.400000095ms/27ms), then restore exact
source hash `a3d81bcbb9a7ab1974b27f3cb6adea713e754dd701e90616f8f89731af40a182`.
Final restored affected-file97461 passes58/four expected deadlines,2.73s. The added controlled
case/optional hook are not in the earlier full73374 population; affected final file/types are the
verification for that refinement. Controlled delays are explicitly diagnostic, not real failure
attribution; ordinary journey input/deadline stays unchanged. No product byte was mutated.

Diff/record checked; all handles terminal. Review by: Codex (diagnostic implementer first filter).
Recorded by: Codex. Exact entire test/capsule/docs/records range after `3c6d5acc` through this
containing commit needs designated Claude review, consolidatable with existing lifetime/observer
corrections. No policy/budget, concurrency, framework upgrade, CI workflow, kernel, balance, copy,
SQL/Firefox/performance/hosted proof, archival, publication or nine-tier milestone change. Next:
separate callback and dispatch intervals on recurrence, or predeclare a bounded workload control;
do not optimize based on the zero post-input response waits or count this local pass as CI repair.

## 2026-10-14 — Bounded worker comparison and real-minute critical path

Resume clean `8352f1c2`; previous goal turn made progress with failure timing/callback controls.
Predeclare exactly two diagnostic arms in the existing plan, not a new program: complete native
Chromium/WebKit functional population, default workers then four per engine, same default+JSON
reporters, no retry. All input, deadlines, assertions, selected files and runtime/test source stay
fixed. Source inspection: installed Vitest4.1.10 creates independent pools capped at12 each on
this14-CPU Mac; the override means4 each,8 total. Hosted4-CPU defaults differ; forcing4 globally
would not even be the same direction of change. No config, environment budget or workflow edit.

Default15139 exits0,292 files/24,430 ordinary passes/ten expected deadlines/16 skips,156.85s.
Four-worker40219 exits0 with the same counts,159.62s. Both passive module populations finish
with no pending/early-closed requests (1,362 and1,361 respectively; other summaries0/0). No
ordinary/controlled Typer deadline fires, so exclusive cause and reliability remain unresolved.
The JSON reporter groups expected-deadline controls into24,440 passed; preserve the separate
ordinary/negative counts rather than inflate positive acceptance. Both reports have exactly
24,456 matching file/title/skip tuples; source hashes remain unchanged. JSON omits engine IDs,
so this is a complete duplicated-case multiset comparison, NOT per-engine identity proof.

Retain concise summary/digests in `native-concurrency-8352f1c2.json`; large raw JSON/logs stay
ignored under the four named `.cache/concurrency-*-8352f1c2.*` paths. Case identity/outcomes
are compared, not merely totals. The retained fingerprint is repository-relative; the initial
absolute-path diagnostic fingerprint is not a second corpus. Single ordered pair/cache/host
load are uncontrolled: shorter Typer durations (2.676/5.629s versus0.672/0.695s) support no
default change or causal/hosted-floor claim, and whole duration does not improve.

The actual cost census changes the next action: Garage's424-case file spans150.546/151.318s
in the default arm and151.621/156.935s in the control, close to both complete runs. Inspect the
actual longest case and accepted GS5-A2: one full native minute without any submitted intent,
checking every250ms, with its unchanged70s test deadline. The observed case lasts60.141–60.204s
across these populations. Do not shorten it, use fake time, move it out of blocking collection
or call4 workers a release fix. Safely separate that independent minute case from the interaction
file, reuse the same fixtures and prove all mapped case identities plus the full native lane.
That is accepted CI D2 test organization, not a changed player or acceptance contract.

All handles terminal; diff checked. Review by: Codex (experiment recorder/first filter).
Recorded by: Codex. No product/test source, CI config, kernel, balance, copy, acceptance,
archival, push or deployment change. These planning/evidence records after `8352f1c2` do not
approve its diagnostic implementation; its full designated review debt remains. No Firefox,
Linux/SQL, isolated performance, full `verify-push`, current hosted or1.0 milestone claim.

## 2026-10-14 — Preserve and independently schedule the real-minute GS5-A2 observation

Resume clean `93b94980`; previous goal turn completes the bounded worker comparison and names
the actual serialized minute in the critical-path Garage file. Under CI D2 and accepted Garage
GS5-A2, move that exact case to `garage-idle-opportunity-browser.test.ts`. Extract its existing
Runtime/base data/mount/settle/opportunity helpers into one `garage-surfaces-fixture.ts` used
by both files; no duplicate fixture, changed callback, player/data/copy or test assertion.
The original60s real-timer loop/250ms zero-intent checks/70s deadline remain byte-identical,
body SHA256 `bea3e1d6d8e846a258b25f11d0ecb202982e64185a0f6ca1d9c9f04ecbf80a59`.

Mechanical source checks: remaining interaction source equals the original with only declared
extractions/import edits; shared bodies equal the originals except export declarations. Initial
raw TypeScript-token check falsely reports drift because it lacks template-literal rescanning
and treats following declarations as template text. Inspect the first mismatch and replace that
invalid check with exact source-prefix normalization; do not credit its failed token comparison.
Both actual typechecks40047/55589 have zero errors/warnings. Node63412 collects and explicitly
skips all424 browser cases, not positive execution proof.

Whole native31866 terminates exit2:294 files,24,428 ordinary passes/two unexpected Typer
deadlines/ten expected deadline controls/16 skips,103.77s. Both independent real-minute cases
pass at60.145/60.177s; all423 remaining Garage cases pass in each engine (spans94.121/94.608s).
Map ONLY the named GS5-A2 file relocation back to its old path: all24,456 file/title/skip tuples
exactly match the complete baseline; no case dropped, added or reclassified. JSON omits engine
IDs as before. Modules finish1,362/1,362 with none pending; other summary0/0. The single elapsed
observation is not a causal/hosted-floor benchmark or a completed browser/CI gate.

The unchanged RP-526 class recurs with useful callback evidence: ordinary WebKit timeout at
submit6/15,002ms, eight completed commands total36ms, no active command, focused connected input,
pending false and native RPC outstanding. Chromium's separately labelled controlled-delay
population also times out at submit4; do not relabel it ordinary player latency. These are
aborted prefixes, not finished-journey measurements. Input/event/automation waits and per-key
handling remain; the complete engine/render callbacks are not the dominant elapsed cost in
this ordinary failure. No default concurrency or timeout/input substitution is made.

Retain concise `garage-idle-shard-93b94980.json` with mapped fingerprint, outcomes and raw
digests; large JSON/logs stay ignored in its named `.cache` paths. All affected Garage
populations pass, but the whole gate stays RED. No rerun to conceal it, new metadata validator,
kernel bump, product/schema/CI-workflow change, archive, push or deployment. Firefox/Linux/SQL,
isolated performance, current hosted and proper nine-tier1.0 remain unproved.
Review by: Codex (test implementer first filter). Recorded by: Codex. Exact complete extraction/
case/capsule/docs/records range after `93b94980` through this containing commit needs designated
Claude review, consolidatable with the existing CI diagnostics. Next accepted work: arrange
native journey execution to test the input-path workload hypothesis while preserving the full
functional population and all original bounds; do not call this test organization CI completion.

## 2026-10-14 — Run the full native keyboard population after broad collection

Resume clean `f6ab6ab6`; previous goal turn clarifies owner choices but makes no implementation
progress. Under accepted CI D2, add `test-browser-functional`: broad collection excludes only
the complete Typer file, which runs second; `test-browser` then retains isolated performance.
Focused selection is byte-unchanged. No test/input/assertion/timeout/worker/browser config,
product/schema/kernel/balance or workflow-job changes. A broad failure stops later phases;
unexecuted is not passed. Separate optional report paths avoid overwriting first-phase evidence.

Actual functional13101 terminates exit0: broad292 files/24,372 ordinary passes/six expected
deadlines/16 skips in93.93s; native two files/58 ordinary passes/four expected deadlines in2.40s.
Compare the union against the retained preceding RED full-run JSON: all24,456 repository-relative
file/fullName/skip tuples match, no omissions/additions/reclassification. JSON still omits engine
IDs; no per-engine ordering inference. Normal native journeys finish178/191ms, controlled
engine-wait cases399.7/400ms. Typer source SHA remains
`a3d81bcbb9a7ab1974b27f3cb6adea713e754dd701e90616f8f89731af40a182`.
Unchanged isolated Chromium performance45491 terminates exit0: five passes/105 selector skips,
5.55s. These separate executions cover the changed functional aggregate and performance lane;
not a default three-engine `make test-browser` or full CI run.

Topology/fixture gate passes16 invalid-state controls and actual Make dry-run selectors/report
safety/full ordering. Execute the original topology function from `f6ab6ab6` against genuinely
removed broad/native/performance commands: all three are admitted before, rejected by the new
function. Node syntax and diff checks pass. No software type/build run for unchanged TS/product.

Raw retained evidence (ignored, not duplicated into a new dossier):
- `.cache/browser-phases-f6ab6ab6.log` SHA256 `1bfefcdbff50da1397e3de340a9350fc819c5c6b3ae3492e9efa2872065886ee`.
- `.cache/browser-phases-f6ab6ab6/functional-broad.json` SHA256 `76861d1f54cca8a95d2af014df0d63e3ff61c9ee2b4299f659d22c5f7c7eb951`.
- `.cache/browser-phases-f6ab6ab6/functional-native.json` SHA256 `9dde8851ab2a24949256ae2c196d7a37239c1e0884e6608b6252ba23e26935b0`.
- `.cache/browser-phases-performance-f6ab6ab6.log` SHA256 `886053efe9b559b05a97124f451c5d6c433ace37da250ab12738a72a067c8dba`.

Review by: Codex (implementer first filter). Recorded by: Codex. Complete range after
`f6ab6ab6` through this containing commit needs designated Claude review, consolidatable.
RP-526 stays open: one native pass is not intermittent/hosted reliability or exclusive cause
proof. Prior red populations remain evidence. Firefox/Linux/SQL/hosted sub-five-minute,
historical guard, conservation, feature/release acceptance and nine-tier1.0 remain open.
No push, deployment or archival. Next: declared Linux job when usable, actual hosted reconciliation
when authorized; do not rerun native for green or infer new owner approval.

## 2026-10-14 — Hosted maintenance timeout and whole-product status reconciliation (RP-596)

Resume clean product `8a5dcebf` (kernel0.3.208/epoch8),75 commits after origin/main4eaed0c9.
Previous goal turn committed Garden correction. Fresh `gh run list` plus both run/job queries
confirm latest CI38381138075 still failed at4eaed0c9: client/server fail, four other jobs pass.
Finished times retain browser6m38s/server6m40s/composed5m37s, not the five-minute floor.
No hosted verdict exists for the75 local commits. Maintenance38507071157 at the same source
finishes failure: numeric passes; Reputation career332.70s and relevance1471.92s fail report
comparison. These log excerpts do not establish a new unique cause or authorize regeneration.

Actual failed logs show the unchanged50-minute harness wrapper exits124. Artifact11869564730
has state incomplete/termination signal, repository guards19,759ms, standard pacing300/300
in2,711,416ms, and first relevance85/107 with1,271,801 transitions in267,687ms. Second declared
relevance never starts; partial guard/exclusion/truncation fields remain unknown, not clear.
Root elapsed2,998,875ms. Completed phase timing is useful cost evidence, not complete harness
acceptance or a measured new full-run ceiling. Next: profile the real standard-pacing population
and distinguish simulator/observer work before a separately authorized optimization proposal.

Read public artifact with `gh run download ... --name harness-observation-38507071157` into
owned temporary storage. Preserve the small exact JSON at
`planning/scaffolding-and-ci/observations/38507071157-harness.json`; `cmp` confirms identical
bytes. SHA256 `5b733fa065bd512d3b4c975081a90bf4c8e1c667ebba6b2fb038777ee6ed6a19`.
Actual `make harness-observation-check HARNESS_OBSERVATION=/private/tmp/cloud-clicker-maintenance.C4Juns/harness-observation.json`
refuses with "incomplete harness observation" (CLI1/Make2), as required. Initial un-escalated
log retrieval had a connection error; subsequent read-only retrieval succeeds. No job rerun,
timeout/retry/population/golden/source change or product telemetry is performed.

Fresh `df -h .` reads71GiB available/92% used; last bounded Docker ping remains unresponsive
at5008ms/zero bytes, not proof the daemon stopped. No cleanup/restart or integration proof.
Update current-state/roadmap/queue using actual source and distinguish fifteen C18 exports,
fixture-first player work, substantial unbuilt scope, technical amendments and owner policy.
Historical evidence is retained; no every-feature acceptance rerun or new milestone inferred.

Editorial/provenance diff, links and artifact identity checked; no unrelated software suite for
Markdown. Review by: Codex (evidence/status reconciliation). Recorded by: Codex. This editorial
checkpoint does not independently approve any outstanding implementation range. All existing
reviews, draft adoption, arithmetic/privacy/accessibility/deployment and full-nine-tier1.0
obligations remain. Continue accepted work; no push, archival or release promotion.
## 2026-10-14 — RP-596 standard-phase CPU profile started, not yet completed

Prior goal turn made source/test progress at `0411ef34`; resume clean main,78 commits ahead.
The current CI plan declares the measurement before execution. Add a manual-only benchmark
that loads the exact standard suite and executes all300 keys through the same
`RunAllWithWorkers(12)` as maintenance; no horizon/seed/dispatch/algorithm/observer or CI change.
It checks complete unique keys, completed outcomes/invariants/aggregate and logs a full-phase
canonical digest. It does not run ordinary tests automatically or replace the complete check.

Executed root command (still live at this checkpoint):
`make test-go GO_PACKAGES='./harness' GO_TEST_FLAGS='-count=1 -run NoTests -bench BenchmarkStandardPacingProfile -benchtime=1x -timeout=60m -cpuprofile=/private/tmp/cloud-clicker-standard-profile.Y4bce9/cpu.pprof -o=/private/tmp/cloud-clicker-standard-profile.Y4bce9/harness.test -v'`.
Actual runtime identifies Go1.27.2, darwin/arm64, Apple M3 Max,14 logical CPUs; hardware read
confirms Mac15,10,14 physical/logical CPUs and38,654,705,664 memory bytes. Initial sandbox
hardware read refused; permitted read succeeded. Scenario SHA10aa5c8a2dd8a5c79a5902e74f4e55a1ca3ab46a2d35c9bf3dd25fc729dbb07a
and constants SHAbaa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6
match the published observation's standard phase. Other heavy tests are not started concurrently.

**No completed measurement, pass, profile digest or hotspot conclusion yet.** Execution session
`9106` is confirmed live by its latest poll; resume that handle, do not restart because a poll
has no new text. Raw profile/test binary stay in the named owned temporary directory. On actual
completion inspect the exit, complete-population checks, reported elapsed/digest and CPU samples;
record failure honestly if it fails. Source inspection confirms standard-phase recorder calls
bracket RunAll rather than every transition, but does not quantify their I/O overhead or prove
the hosted cause. Native samples/timing are not Linux/hosted timing, a full-check verdict or
optimization/CI/timeout authority. Future completed batch still requires designated Claude
review; no self-acceptance, cleanup, push or release claim. Current benchmark/docs/plan are
uncommitted work in progress pending that executed result, not a finished delivery.

## 2026-10-14 — RP-596 complete native standard-phase profile and reproducible reader

Resume the same live9106 session from the preceding turn; do not restart. It terminates exit0:
all300 unique governed keys complete, aggregate/individual invariant checks pass. Benchmark wall
783,821,386,125ns (13m3.821s); whole Go command784.387s. Canonical complete phase report SHA256
`1ffa47cfd4f85f6daf1b340cf6e43cfa7d534cb39308cdf80e955c9d28be826d`.
The previously declared native hardware, Go1.27.2,12 workers and input identities apply; no
production/balance/epoch/golden/scenario/horizon/workflow/dispatch/acceptance bytes change.

Raw files remain `/private/tmp/cloud-clicker-standard-profile.Y4bce9/`:
CPU profile SHA256`aebcd0a56767a41a7066bfc2cbc1d1b70a218e1fe09a801e140c31efd3e86219`;
compiled test binary SHA256`d7594c1006507c66dd40ef05dd7ee652f4c2611b5d3ec38d95dcabe5acce1486`;
executed benchmark source SHA256`218991f980444a817a94cbc24adc24e5be4dfc61f208accfe79ab8c669f2d83c`.
No binary/raw profile is added to Git. The shell's prebuilt `go tool pprof` is absent; source is
present in the same toolchain. Add manual root `harness-profile-report` using `go run cmd/pprof`
and the existing Make-owned cache, not an installation/custom cache/new simulation instrument.

Executed the reader on the actual profile with default top25, `-top -cum -nodecount=30`,
`-list ratesWithProvisionedAndPolicy -nodefraction=0` and focused cumulative view. Profile spans
784.02s,5854.30s aggregate sampled CPU across threads. Production rate calculation contributes
2285.94s/39.05% cumulative sampled CPU. Source attribution at engine.go333's repeated bySource
map iteration is1531.81s cumulative; this is inside the former subtotal, not an additional
disjoint saving. Map iteration/write-barrier/runtime allocation paths are prominent. Flat runtime
madvise23.04% and usleep15.42% are also reported; do not infer a Linux cause from Darwin samples.
These CPU attributions overlap, are not wall-time percentages and predict no guaranteed speedup.
Source inspection shows standard observation writes only before/after RunAll; no timed observed
versus unobserved pair or quantitative recorder I/O cost is measured by this profile.

Reader missing-input and malformed-text controls both refuse (Make2); real profile reads pass.
Cold dispatch tests execute three cases, including mismatched-key refusal and worker-count byte
stability; affected `make vet GO_PACKAGES='./harness'` passes. `make verify-ci-topology` passes
its actual topology and16 negatives; new reader has no CI/release dependency. Formatting/diff
checked. Do not run unrelated broad suites to decorate an instrumentation result. This is a full
standard population measurement, NOT the complete check: guards, relevance and golden/baseline
comparison stay outside it. No Linux/SQL/hosted-green, performance improvement or1.0 claim.

Review by: Codex (diagnostic implementer first filter). Recorded by: Codex. Complete benchmark/
Make/docs/plan/queue/ledger/start-and-result-log range after `0411ef34` through this containing
commit needs designated Claude review, consolidatable with CI diagnostics. Next: propose a
separately authorized rate-traversal optimization that preserves exact multiplier order,
validation/activation and report identity; prove actual gain and full affected runtime parity
before adoption/closeout. Existing neutral-delivery/kernel authority, Linux/hosted/full-maintenance
and unrelated numeric/history/release holds are not waived. No cleanup, publication or archival.

## 2026-10-14 — Scoped rate-traversal proposal and executed exactness controls (RP-596/RP-599)

Resume clean `a114b972`; preceding turn completed the full native CPU measurement. Translate
the measured repeated-map scan into the existing plan's Production C2 refactor proposal: validate the whole set
first, prepare sorted per-slot buckets once per call, filter matching targets without changing
factor sequence, and activate only at existing application points. No cross-call cache,
reassociation, arithmetic/quantization/seed/worker/budget/CI change. Runtime-neutral eligibility
and the exact-path history-scope mechanism are separately unadopted; the descriptor-only lane
cannot authorize this runtime refactor. No production algorithm is implemented in this batch.
Initially created an extra performance RFC, then removed that uncommitted agent-authored draft
and its new index row: RFC-0000/Marco's current procedure authorizes equivalent routine refactors
under the existing owner contract, not another specification gate. Keep the substantive scope
and tests in the existing plan; no historical/user work is removed. The discarded text remains
in this task's edit history if needed. Measurement-only authority is not repurposed to change
Production C2; the actual unresolved guard-policy/identity-review boundary remains explicit.

New test-only fixture first failed legitimately: epoch5's v3 cannot declare an axis_stack source.
Do not loosen the loader. Use the actual existing v5 axis fixture, retain its real rows/effect
bindings and reuse its Axis/Commons sources; add only controlled declarations for other factors.
This is admitted synthetic data, not minted content, natural producer factors or player pacing.
Literal expected traversal covers all nine slots and interleaved global/specific upgrade IDs.
Five purchased/provisioned/zero/masked states ×three input rotations check exact mantissa bits/
exponents, contribution order/resource slices, activation set and caller-input immutability.
Rejected reversed multiplication and global-first-within-slot alternatives really differ;
neither can be waved through by a generic multiplication/product argument. Twelve invalid
unused-factor populations (six corruptions ×zero/masked) refuse before any activation/output.

Executed cold TestRateTraversal, existing TestRates, both simulation synergy-application tests,
frozen-rate valid/refusal population and the complete multiplier package; all pass. The first
mixed-package TestRates selector ran no multiplier tests, so it was not claimed as that proof;
the later complete multiplier run executes its two named tests. Affected vet passes. Root
`make formulas-check` is byte-unchanged. New snapshot assertions are included in the final
TestRateTraversal rerun. No unrelated full production/harness/browser/SQL/hosted gate rerun for
test-only proof; known Clout partitions/history/service/full-maintenance/release holds remain.

RP-599 corrects canonical Production prose: its rate subsection omitted axis_stack, while its
opening section, actual multiplier registry and accepted Clout CV3/OD-6 include that slot.
This is editorial reconciliation, not an accepted-body/ruling change, retune or formula rewrite.
The current CI plan/queue/backlog now agree on the next authority/proof boundary.

Review by: Codex (test implementer first filter and refactor-proposal author). Recorded by: Codex. Full
test/docs/plan/queue/ledger/log range after `a114b972` through this containing commit
needs designated Claude review, consolidatable with CI diagnostics. Existing guard-policy
adoption and exact-path identity review precede runtime landing; full affected parity and
measured gain precede closeout. No claimed
speedup, kernel bump/exception, archival, mint, push or narrowed nine-tier1.0 floor.

## 2026-10-14 — Published CI38560135778, not the old remote-source result

Read-only gh run list/view identifies completed failure at published `be66902b`.
Finished failure markers confirm client RP-131 at50a3a514 and all27 server
TestAxisAccrualPartitionProperty failures (RP-307). Schema/harness/browser/
game-ui-composed jobs succeed. Job metadata gives browser5m21s/composed6m07s/
server8m11s, each above D-014's five-minute floor. Current local changes are
not included. CI38381138075 and maintenance38507071157 remain earlier evidence.
No workflow, budget, assertion, report, guard or authority changed; these are
executed external-state reads, not a new diagnosis or repaired CI claim.

Docker subsequently responds and the declared local Linux-amd64/Postgres backup
lane passes; the Deployment log owns that separate result. Do not generalize
daemon recovery or scoped backup success to whole server/browser/hosted CI.

## 2026-10-14 — Linux browser command parity and retained full-population timeout

Resume clean `81107554` after completed player/SQL verification. Under CI D2,
correct RP-613: Compose's copied Vitest command bypassed the shared broad/native/
performance phases. Delegate to root `make test-browser`; no workflow, test,
assertion, project, worker, timeout, gameplay or kernel change. Prepare GNU Make
in the disposable pinned Playwright image, preserving frozen pnpm installation.

Regression: new negative fixture against the original verifier exits1 for its
admitted copied browser command. The corrected verifier also rejects the actual
unchanged old Compose command. After correction, `make verify-ci-topology`
passes the original16 negatives plus four copied/incomplete/narrowed container
controls and actual Make selector/report checks. Node syntax and diff checks pass.

Actual `make test-browser-ci`67706 exits2/Error127 after frozen install: the
image lacks Make, no browser test ran. After adding that prerequisite, full
52310 is terminal Make2/child137 with GNU timeout's explicit KILL message at
the unchanged10-minute ceiling. The broad phase runs; native and performance
never execute. Do not call partial file results a completed gate. Image metadata
is Linux/ARM64, not hosted AMD64. A live resource sample at14:42:44 UTC reads
CPU105.22%, memory11.36GiB/15.6GiB and1610 PIDs; not a cause or hardware benchmark.

Read-only live-log comparison with retained
`.cache/browser-phases-f6ab6ab6/functional-broad.json` (146 distinct paths):
all reference paths have terminal file lines in Chromium/WebKit; Firefox lacks
`test/reputation-host-browser.test.ts`. Current completed-file counts158/157/158
also include newer files; this is not a current title/skip/coverage-union proof.
No output for the missing file is inferred from truncated tool chunks.

Focused diagnostic71583: declared Compose run with the same preparation and
containment, overriding only the leaf to `make test-browser-focused
BROWSER_TEST_FLAGS="test/reputation-host-browser.test.ts --project firefox"`.
Terminal exit0: one file,89 ordinary passes/one deliberate expected failure,
23.757s tests/26.55s total. This changed diagnostic population is not an unchanged
full-run retry, full success, hosted evidence or explanation of the timeout.
RP-526 remains open; next instrument that file's full-population input/cleanup
completion, preserving the original failure and all semantic/latency bounds.

Review by: Codex (implementer first filter). Recorded by: Codex. Complete range
after `81107554` through this containing commit needs designated Claude review,
consolidatable with prior CI work. Both runs use automatically removed disposable
containers/anonymous dependency volumes; final declared inventory is empty. No
push, archival, changed guard authority or full-nine-tier acceptance follows.

## 2026-10-14 — RP-526 Reputation host deadline observation

At clean `a57b7fd8`, the prior default Linux timeout remains red; isolated
success does not explain it. Under CI D2/RFC-0000, add test-only failure
observation to the mounted host: case/user-agent, elapsed stage, active input,
native calls started/settled and timing, focus/mount/busy, snapshot/intent and
controlled socket state. Capture before teardown; preserve native RPC, lifetime
guards and drain order. No player, worker, assertion, timeout, balance or CI
population removal. One additional ordinary control exercises actual native
navigation and silent success; the existing deliberate deadline stays50ms.

First candidate13697 fails two observer assertions because Vitest's reason
includes help text. Next12451 fails two stage assertions because the deadline
fires while subscription readiness is awaited, before the producer gate exists.
Compare the actual reason and independently captured continuation-at-deadline;
do not lengthen50ms or falsely label initialization as a producer wait.

Final full mounted population36323:180 ordinary/2 expected failures in native
Chromium/WebKit,24.97s. Actual Linux/ARM64 Firefox94144:90 ordinary/1 expected,
26.03s, frozen disposable Compose preparation and root focused leaf. Neither
is a whole Linux/hosted gate. Disconnected-observer41455 fails both negative
controls with empty evidence; restore registration exactly. Final selected
controls91264 pass4 ordinary/2 expected (176 explicitly deselected),1.54s.
Both expected payloads reach actual Vite error output with user-agent/stage and
still-mounted state. Initial console.info delivery yielded no visible marker;
use the demonstrated error channel, clearly label deliberate failures. Types/
Svelte27556 pass with zero errors/warnings; diff checked. No client rebuild for
unchanged shipped code. All named checks are terminal; disposable inventory empty.

Next: run the ordinary full Linux job with these passive observations; no
unchanged green-seeking retry or cause/reliability claim. Review by: Codex
(implementer first filter). Recorded by: Codex. Exact range after `a57b7fd8`
through this containing commit needs designated Claude review, consolidatable
with CI work. RP-526, hosted latency, other CI failures and full1.0 remain open.

## 2026-10-14 — Complete default Linux browser gate at 8e4d0877

Continue the same live `make test-browser-ci`88376, not a restarted job.
At committed `8e4d0877`, terminal exit0; Linux/ARM64 pinned Playwright image,
fresh frozen dependencies and shared Make target, no selection override.
Broad474 files:39,615 ordinary passes/9 expected failures/24 reported skips,
149.32s. Separate native-keyboard3 files:87 ordinary/6 expected,3.81s.
Isolated Chromium performance:5 passes/105 pattern skips,6.58s. All three
required phases actually execute. No aggregate count substitutes for their
individual outcomes or claims every acceptance population ran.

Expected Reputation deadline records from Chromium/Firefox/WebKit reach the
live container log with still-mounted context and zero native calls during
subscription readiness. No ordinary Reputation deadline record appears in the
filtered log before completion. The earlier Firefox missing-file timeout does
not recur here; neither its cause nor elimination is demonstrated. Keep52310's
full failure and isolated71583/94144 outcomes. No assertion, input, worker,
deadline, population or performance floor was loosened for this pass.

Large module traces truncate tool chunks; read the live container's specific
deadline lines before removal, then recover final summaries/exit from the same
session. No reconstructed raw transcript or title-union claim. Automatic
disposable cleanup completes; final declared inventory empty. This closes the
previously unexecuted current default Linux gate, not native-amd64 hosted parity,
five-minute Actions timing, other CI failures, independent review or full1.0.
Verification by: Codex; Recorded by: Codex. Editorial evidence checkpoint;
underlying source/review ranges remain open and may include this closeout.

## 2026-10-14 — RP-623 cold defaults reject real corpus drift

Under CI D2/RFC-0000, make Replay/Pitch/Typer corpus checks explicitly count1. Generic
`test-go` places its cold default before GO_TEST_FLAGS, preserving focused selection and an
explicit developer repetition count. Compiler/module caches remain; no job topology, test
population, retry, deadline, numeric/balance/kernel/product behavior or golden is changed.

Actual old failure: healthy `make replay-fixture-check` returns cached success. Temporarily
change only the committed fixture's root version1→2, rerun the identical old target, and it
still exits0 with `(cached)` (ee0b8c). Restore in finally; Git confirms exact original bytes.
This demonstrates a local false green, not the cause of every hosted failure. Inspection also
finds the same omitted cold flag in Pitch/Typer checks and the generic target; their old corrupt
behavior is not inferred from the one executed replay case.

Actual corrected negative controls: mutate each original corpus's version separately and run
its real Make checker. All three exit2 with their own generator/byte-comparison assertion:
shared ApplyLogged fixture stale; Pitch content corpus drifted; checked-in Typer corpus stale.
Each finally restores the original bytes; `git diff --exit-code` across all three is clean.
No generated replacement or intentionally bad corpus remains in the tree.

Finished verification by: Codex. Recorded by: Codex.

- Native healthy three-target batch93872 passes0.515s/.217s/.100s. Declared Linux-amd64
  execution98524 runs those exact root Make targets inside the existing test container,
  passes1.008s/.459s/.123s, exit0. The standard emulation CPU setting remains. Its disposable
  Postgres dependency starts, but these three gates are pure corpus checks, not SQL journeys.
- Generic focused root execution33887 runs the actual test once, passes. Explicit count2
  execution42948 visibly runs it twice and passes; neither result is cached.
- Existing topology fixtures initially fail because the old validator admits cached test-go
  (2c617c). The corrected verifier rejects all20 mutations, including four dropped defaults;
  browser controls and actual Make expansion/selector/repetition controls remain green.
  These recipe checks supplement the real corrupted-corpus proof, never substitute for it.
- Diff/whitespace checks pass. No software build/whole browser/full server job is claimed for
  this Make/diagnostic-only change. RP-131, RP-307, D-014 latency and other CI holds remain.

Designated Claude review is pending for the single commit containing this entry, parent
`1f1a5481`, all seven Make/tool/docs/plan/ledger/log paths. No self-approval, archival, push,
deployment or full release promotion. After terminal checks, remove only the fresh declared
postgres test service; inventory is empty. Disposable tmpfs is discarded; no durable volume,
compiler/module cache, network, unrelated service or user data is removed.

## 2026-10-14 — RP-633 Reputation maintenance recipe guard

Under CI D2/D3/AC5 and RFC-0000, bind the actual `reputation-harness-check`
recipe, not just the workflow's invocation. The existing target still opts in
to both named studies with count1 and its unchanged60-minute timeout. No Make,
workflow, study, balance/report, kernel or production bytes change.

Regression executed against real current workflow/Make/browser sources: load
the old verifier from `c05a196a` without modifying either checkout. Three
independent string corruptions (remove opt-in, select only one study, rename
the target) each pass the old guard and fail the correction. Both verifiers
accept the healthy sources. The missing opt-in could skip both tests while Go
returns success; this is a validation blind spot, not an observed hosted skip
or attribution of an existing CI failure.

Finished checks: `make verify-ci-topology` exits0, rejecting26 permanent
mutations, six of them covering this recipe. Existing Linux-browser and actual
Make selector/repetition/report controls also pass. Both changed scripts pass
`node --check`; diff/whitespace checked. The separate historical complete
Reputation run remains active and is not a passing result for this checkpoint.
No unrelated package/build/browser/SQL suite or fresh hosted result is claimed.

Review by: Codex (implementer first filter). Recorded by: Codex. The coherent
six-path corrective range after `c05a196a` through the containing commit needs
designated Claude review, consolidatable with prior CI corrections. Independent
original maintenance-range review and study results belong to the Reputation
log; this fix does not approve that range or complete CI/feature/1.0 acceptance.
No new owner decision, RFC, deadline, archival, push or deployment is needed or
performed for this conformance repair.

## 2026-10-15 — RP-635 observation IDs bind to their existing identity

Verification began2026-10-14; this closeout crosses midnight. Under archived
Harness Observability H1 and RFC-0000, require the existing stable
names for repository guards, standard pacing and registered relevance's row
index/scenario path. Declaration equality alone was not enough: six executable
old-validator cases rename both copies and still validate, including invented
Reputation names and wrong row coordinates. Two healthy controls pass at that
same old revision. The correction rejects those six cases; a valid inactive
selected row retains its noncurrent constants identity.

Verification by: Codex. Recorded by: Codex. Focused recorder/identity/registered
selector/signal tests pass, including an actual registered23-run row with exact
golden bytes and a real SIGTERM subprocess leaving incomplete evidence. Selected
research-observation tests explicitly skip; they are not fresh Reputation studies.
Cold full affected packages pass: harness73.091s and balance-harness CLI0.596s,
then same-package vet passes. After adding the inactive-row positive control,
the final focused nine-case identity population and harness vet pass0.288s;
unchanged production/other cases retain the preceding full-package result.
`make verify-ci-topology` passes26 negative controls and existing Make/browser
controls. Formatting/diff checks pass. No exhaustive study, SQL/player journey,
full hosted workflow, artifact authenticity or new objective acceptance claimed.

This changes only observation validation, not its shape/producer, allowed modes,
simulation, balance/report bytes, CI population/budget or kernel identity. Canonical
docs describe the self-consistency boundary. RP-634 still requires a bounded
objective amendment and the missing D3 hosted interruption/upload proof; do not
silently extend/relabel the v1 schema. No owner decision or new RFC was required
for this existing-name conformance correction.

Review by: Codex (implementer first filter). Recorded by: Codex. Complete six-path
range after `a201a7e8` through the containing commit requires designated Claude
review, consolidatable with the prior CI corrections. No self-approval, archival,
push, deployment or release promotion. Continue accepted construction and the
bounded RP-634 proposal without narrowing the complete nine-tier1.0 objective.

## 2026-10-15 — RP-634 bounded observation amendment drafted

`rfc/reputation-harness-observability.md` proposes a typed fixture profile using
the shared recorder, leaving v1 untouched and applying CI D3's existing50/55-minute
floor. Complete measurement is distinct from H4/H5 acceptance; failed criteria,
reasoned domain exclusions, report drift and all admitted arms remain visible.
There is no new gameplay, balance, optimization or population choice.

Checked against recorder/validator, actual test producers, source/lineage helpers,
report paths and maintenance workflow. A read-only current-fixture census confirms
64/32/1 seeds and9 nodes:194 H4 arms,970 H5 arms. Signal propagation through the
actual root target, dirty/source/report refusals and unchanged-domain comparisons
are explicit future acceptance checks, not executed results. Draft/index links
and diff/whitespace checked; no software suite for this proposal-only batch.

Review by: Codex (author first filter). Recorded by: Codex. Five-path draft/records
range after `ab3442dc` requires adoption and designated normative review before
implementation. The archived parent and all product/tool/workflow/data bytes are
unchanged. RP-634 remains open; prior corrective reviews remain separate.
Scope adoption requested2026-10-15; no answer or independent approval is inferred.

## 2026-10-15 — Current full Linux server gate remains red

Run `make verify-server-ci`8001 at product source `3bde51b5` (kernel0.3.216),
using the existing Linux-amd64/Postgres composition and its unchanged emulation
CPU setting. Vet succeeds;64 test packages pass. Production alone fails107.387s
with exactly27 `TestAxisAccrualPartitionProperty` children (RP-307); all named
one/split decimal differences are captured, not a truncated census. Overall
Make exits2. The subsequent Pitch/formula/API generation and boundary leaves
do not run after `test-go-core` fails. Quiet package output does not enumerate
skips, and a passing package is not full feature acceptance.

Gameserver88.533s, Arcade48.594s, Save1.996s and Transport13.964s pass. This
refreshes current-source Linux/SQL consumer evidence, not RP-614 cause/reliability,
hosted native-amd64/five-minute timing, the full push aggregate, natural balance
or 1.0 acceptance. No source/test/limit/population changes, assertion relaxation,
retry-until-green or golden rewrite. An editorial Transport reconciliation lands
while this run lives; server/client/kernel/balance diff from its start is empty.

Verification by: Codex. Recorded by: Codex. The fresh declared test runner
terminates and auto-removes; remove only its owned postgres container afterward,
discarding regenerable tmpfs fixtures. Final composition inventory is empty;
no durable/user data, cache, volume or network removed. Record-only checkpoint,
diff/links checked; prior reviews and named conservation/author holds remain.

## 2026-10-15 — Current browser gate exposes RP-637

Clean `b248d26a` full Linux/ARM64 browser execution exits2 at Firefox Codex Close;
diagnostic execution fails the same case at Open. Later instrumented full execution
passes all phases without a production fix. Detailed commands/counts, retained failure
text, final81-case diagnostic control and exact review boundary belong to the
[Clout log](../clout-v1-and-pr-interns/log.md#2026-10-15--rp-637-full-linux-native-disclosure-failure-and-diagnostics).
Current plan/status link there; previous successful `8e4d0877` is historical, not
the current-source verdict. No CI topology, budget, assertion weakening, hosted
success, reliability closure or release claim. RP-637 remains agent diagnosis,
not a pending owner decision; the full server/history failures remain independent.

## 2026-10-15 — Complete browser gate at fc398997

Clean `fc398997` `make test-browser-ci`75582 terminates exit0 with the committed
final diagnostic controls: broad477 files/39804 ordinary passes,9 expected failures,
24 skips in184.54s; native3 files/87 ordinary and6 expected failures in3.93s;
Chromium performance5 pass/105 selector skips in6.66s. Same declared cold install,
root Make phases, Linux/ARM64 image and unchanged ten-minute containment. Source
stays unchanged during execution. Full stdout is partly truncated; terminal summary
and these phase results are retained, not a complete individual-case transcript.

The attempted parent-focus probe was fully withdrawn before this run; its negative
result belongs to the Clout log. RP-637 does not recur here and remains unattributed,
not repaired. This is current local browser evidence, not hosted AMD64/runner timing,
full push aggregate, feature/release acceptance or independent approval. Previous
browser failures and server/history failures remain. The runner auto-removes; final
service inventory is empty. Record-only checkpoint; prior corrective review debt
persists, with no product change, archival or push.

## 2026-10-15 — RP-645 deadline populations and RP-637 cancellation evidence

Under CI D2/RFC-0000, extend the existing Codex failure capture with post-dispatch
defaultPrevented, key repeat/modifiers, focus at delivery, keypress and toggle.
Retain Event references until capture so a later target listener's cancellation
is visible; the capture phase alone is too early. Existing blocked Close now
requires uncancelled first Enter/cancelled second Enter, without repeat/modifiers.
Same-image focused55483 passes all81 cases; types78820 pass. No product repair or
RP-637 causation inferred; its original two failures remain open.

Full `make test-browser-ci`99155 then fails a different test: Firefox Reputation
deadline cleanup expects `sockets:[{closed:false,commands:3}]`, observes `[]` at
the actual50ms deadline. Broad477 files:476 pass/one fail,39815 ordinary passes,
9 expected failures/24 skips,148.99s; native/performance do not run. The final
assertion at old line240 is retained, although bulk stdout is truncated. Previous
stage handling allowed early initialization but still asserted a subscribed socket.

Correct the fixture, not the UI: ordinary bounded beforeEach establishes either
a subscribed controlled socket or a held initial HTTP read, before the unchanged
50ms test-body deadline. The installed runner confirms hooks precede that timer.
Each real deadline must capture its exact population, unmount/close every created
socket, then reject a released late continuation with the original abort reason,
no native action/intent, no new socket or remount. Ready-arm socket requirements
remain exact; the separate pending-read arm does not pretend a socket existed.
Focused61447 passes9 ordinary/6 expected timeouts in all three engines (264
explicitly deselected),2.92s. Final types14449 report zero errors/warnings.

Final full76960 uses the declared Compose image/timeout/frozen preparation and
root `make test-browser BROWSER_REPORT_DIR=client/test-results/rp645-20261015-728508de`
(reporting-only command override). Exit0: broad477 files/39819 ordinary passes,
12 expected failures/24 skips in138.77s; native3 files/87 ordinary/6 expected
in3.61s; Chromium performance5 pass/105 pattern skips in6.16s. Source stays fixed:
axis test Git blob `eb779688f9876cd786f36d4a0c15f425c6cbbbf5`, Reputation
`5ac2cef441f81903c12852bd3871a1673c4eacab`, above product `728508de`.
Both JSON reports validate success/zero failures and include all15 Reputation
deadline controls plus three blocked Codex controls. They classify expected
failures as passed; preserve the CLI split above. Ignored raw reports remain local:
broad12269534bytes SHA256 `1420e3888d298dfe353ea455f8b656a22f2e8e4b39470b0d3749eec140895854`;
native26173bytes SHA256 `39efa3118876af34157ab5d038c23e077ff2e70d8fc0506452bb019a83b203dc`.

Image inspection confirms Linux/ARM64, not hosted AMD64/runner timing. RP-637 does
not recur and remains unattributed; no repeated green closes it. No production,
timeout, retry, budget, workflow, kernel, schema, copy or content change. Full
server/history failures, current-source hosted proof and all release gates remain.
Diff checked; both runners auto-remove and service inventory is empty, no cache/
network/durable data cleanup. Review by: Codex (test implementer first filter).
Recorded by: Codex. Complete five-path test/records range after `728508de` through
this containing commit needs designated Claude review; no acceptance, archival,
push or deployment. Next use actual recurrence evidence for RP-637 and continue
accepted construction, not another unchanged full run for reassurance.

## 2026-10-15 — RP-121 Guild fixture isolation on a reused database

Under RFC-0000's ordinary fixture-repair procedure/CI D2 and the unchanged Guild
contract, reproduce old `make test-save-integration SAVE_TEST_PACKAGES='./guild'
SAVE_TEST_FLAGS='-run Integration -v' SAVE_TEST_COUNT=2`82393: first three cases
pass, second lifecycle gets two invalid intents and the other two hit accounts PK
duplicates. No resetting the database/service between those repetitions.

Factor setup/cleanup for the three original fixtures. Before any migration or
TRUNCATE, query the actual database name and require the declared cloud_clicker_test
target. Clear guilds/accounts/save-stream roots and their test dependencies before
and after each case; cleanup precedes connection close. No t.Parallel is permitted
on this shared fixture; all existing transaction races, idempotency, history-trigger,
clearing, health and leadership assertions remain byte-unchanged. The negative
connects read-only to maintenance postgres and requires the exact refusal sentinel,
so an unrelated SQL/connection error cannot satisfy it.

First corrected16491 passes all four cases twice1.779s. Tighten the negative's
error identity; final1622 passes the separate process with -shuffle=12345 and
-count=2 on the same still-running Postgres service1.842s, no selected skips.
Final cold native Guild/vet77829 pass0.230s. No production, schema, migrations,
kernel, balance, privacy, CI topology/budget or archived-contract change. These are
fixture repeatability checks, not a new Guild/player/erasure or full-CI acceptance.

Diff/whitespace checked; all handles terminal. Remove only this turn's inspected
owned tmpfs Postgres, preserving caches/network/durable data. Refresh RP-121 and the
current Guild research pointer without rewriting its historical failures. Review
by: Codex (test implementer first filter). Recorded by: Codex. Whole five-path
test/ledger/research-pointer/plan/log range after `604f2a82` through this containing
commit needs designated Claude review. No acceptance, archival, push or deployment;
numeric conservation/history, owner/author and release gates remain independently open.

## 2026-10-15 — RP-648 correction review scope must contain its target

Do not repeat RP-131's completed future-request repair. Re-read the current tool:
611273b5 already recognizes completed visible metadata. A different conformance
gap remained: Reviewed range was ignored while any exact offending anchor in that
same section could bind the correction. New actual isolated-history CLI2389 admits
all four initial wrong/empty/invalid/duplicate scope examples and exits1 at the
fixture's explicit admitted-examples assertion. No real repository history changed.

Retain the declared range metadata, refuse duplicate declarations, and resolve its
actual Git commit set before accepting an incidental offending anchor. Honor both
exclusive-start and caret-inclusive syntax; unresolvable ranges refuse. Do not
require a new field in existing legacy records, change a negative completed verdict
to approval, or infer reviewer identity from the syntax. Existing semantic-correction
and mechanical-remap classes, append-only/path/version rules are unchanged.

First corrected fixture12848 passes. Final3338 includes six explicit-scope negatives
and six completed-record positives, alongside all existing miss/remap/immutability/
ancestry/archive/shallow/merge controls, and passes. Two extra exclusive-start/
unresolvable negatives and the enclosing positive were added after the four-case
old baseline; they are final controls, not additional recorded old-tool failures.
Both Node syntax checks and make verify-ci-topology pass (26 negative controls,
existing Go/Make/browser-container dispatch checks). No unrelated software suites
for this CLI-only behavior are needed or claimed.

Actual make verify-kernel-version14274 passes its CI-checkout contract and negative
fixtures, admits all five current correction bindings, then exits2 at unchanged
50a3a5141d5c4c1cd25cbf3c9c39d9e3cf8e2444 versus parent0cf9f7a on the same six
registered minigame paths. Its final fixture recipe is not reached after that failure;
the separately executed full fixture above supplies that evidence. This is not a
client/push CI pass, scope exception, false bump or authorization to rewrite history.

Canonical CI docs/live plan and RP-648 reconcile. Update my own history-exception
draft's stale motivation and split its proposed exception approval from completed
negative semantic-correction evidence; do not represent this existing-rule repair
as acceptance of that proposal. Product, kernel0.3.216, path/correction registers,
workflow/populations/budgets and real cited verdicts remain untouched.

Diff/whitespace checked; all handles terminal. The existing fixture owns and removes
only its temporary Git repository; no development services or user data touched.
Review by: Codex (implementer/draft-author first filter). Recorded by: Codex. Entire
seven-path CLI/tests/docs/plan/log/ledger/draft range after `691e9bac` through this
containing commit needs designated Claude review. No feature acceptance, archival,
push/deployment or release; Headcount answers and the other named holds remain open.

## 2026-10-15 — RP-596 isolated rate-traversal candidate; complete measurement running

Outcome at clean baseline `9bdf9c29`: construct an experimental implementation of the
existing Production C2 proposal without landing a watched runtime edit or inventing a version
signal. `rate-traversal-candidate.patch` applies cleanly to the active engine, but remains
unapplied. A temporary copy at `/private/tmp/cloud-clicker-rate-candidate.2kNxJt/engine.go`
is compiled only through that directory's explicit Go overlay. Full contribution validation
stays first. For the first active/unmasked generator, construct per-call buckets indexed by
the existing multiplier order, sort each by the unchanged full source/target identity and
filter its targets without reassociating any multiplication. No shared cache or changed
generator/resource order, count arithmetic, activation point or persisted/wire behavior.

Source SHA256s: active engine `b872d4746779a097f68f88475cef66214a2a7f8970817298f788ffad515ae875`;
final candidate `d9b789badba9585010ecd332170ffe559f77c9ceb3665984fc81e1236bbb52ea`;
retained patch `300f496d23148c542f4972878b592fe292145cec44d1bc718c8f0f0c74a49025`;
new benchmark source `8be15a70af591a9eb085b95ff07d58e3c0ba00dfee023be4ff54de8f41d04849`.
The patch's original/new Git blobs also match the actual two engine files. `git apply --check`
passes; no patch is applied to the checkout. An overlay is experimental source replacement,
not a governed artifact or a kernel/history exception. Historical reports stay untouched.

Manual benchmark added to the existing ordering-fixture test file; no new measurement framework
or CI lane. Root `make test-go GO_PACKAGES=./production
GO_TEST_FLAGS='-run NoTests -bench BenchmarkRateTraversal -benchtime=200ms -count=3'`
and the same command with `-overlay=/private/tmp/cloud-clicker-rate-candidate.2kNxJt/overlay.json`
run sequentially on Go1.27.2/darwin/arm64/Apple M3 Max. Three-sample medians below are
microbench observations, not statistical or whole-harness/hosted improvement claims:

| Controlled profile | Active ns/op | Final candidate ns/op | Active→candidate allocations/op |
|---|---:|---:|---:|
| no contributions | 4,016 | 3,169 | 58→58 |
| all nine slots | 20,567 | 6,535 | 341→84 |
| inactive | 2,554 | 2,774 | 49→49 |
| masked | 2,592 | 2,836 | 49→49 |

The final empty-input samples include a5,154ns outlier; inactive/masked timing is still worse
in this small sample, despite equal allocations. Do not erase either. The first eager-map
prototype used63 allocations and roughly3.8–4.4µs for inactive/masked versus49 and roughly2.6µs
active; avoid that unused preparation in the final candidate. An intermediate lazy-map edit
failed compilation on an unused loop variable; corrected before successful execution. Neither
that build failure nor discarded prototypes are credited as completed candidate evidence.

Final overlay: cold focused `TestRate` and `TestSimulation` pass, including exact mantissa bits/
exponents, all nine slots, interleaved global/target order, activation/input preservation,
12 unused invalid-factor refusals, projection and frozen-rate controls. Cold full native
Production executes in37.813s and fails exactly the original27 partition cases; no new failure
or arithmetic repair is claimed. Database-dependent tests have no real Postgres prerequisite
here and receive no integration credit. Active-source Production/Multiplier vet passes;
candidate-specific full vet remains owed. No independent verdict or full-domain equivalence.

**Live full-population measurement:** session51264, same existing benchmark, exact root command:
`make test-go GO_PACKAGES=./harness
GO_TEST_FLAGS='-overlay=/private/tmp/cloud-clicker-rate-candidate.2kNxJt/overlay.json -run NoTests -bench BenchmarkStandardPacingProfile -benchtime=1x -timeout=60m -cpuprofile=/private/tmp/cloud-clicker-rate-candidate.2kNxJt/cpu.pprof -o=/private/tmp/cloud-clicker-rate-candidate.2kNxJt/harness.test -v'`.
Its startup identifies12 workers,14 logical CPUs, scenario `10aa5c8a…bb07a` and epoch8
`baa89050…128f6`, matching the earlier complete standard-phase profile. The last poll confirms
the same handle live, with no result yet. Resume it; do not restart on silence. No concurrent
heavy benchmark/test is launched after it. On completion inspect all300 governed keys,
failures, wall time and canonical report digest against historical `1ffa47cf…e826d`; disclose
any mismatch. This phase excludes registered relevance, CLI guards and full maintenance.
The earlier783.821s profile is historical, not a matched current-source timing control.

Review by: Codex (candidate/test implementer first filter). Recorded by: Codex. The new
benchmark/patch/docs/plan/ledger/log range after `9bdf9c29` needs designated Claude review;
this checkpoint is not an optimization adoption or a landing-ready full-runtime verdict.
Existing exact-path guard authority/independent eligibility and remaining parity/replay/player
proofs precede active runtime landing/closeout. No kernel/balance/golden/workflow/budget change,
new mechanics, shortened population, publication, deployment, archive or1.0 promotion.

Packaging follow-up: `git diff --cached --check` flags unified-patch context lines' required
space before existing Go tabs. The checkpoint commit was made despite that failed packaging
check; do not record it green. A file-exact `.gitattributes` rule identifies this patch-data
format only; source whitespace and other patch checks retain their defaults. This does not
alter the candidate or runtime, and belongs to the same pending review range. Recheck the
whole range and the patch's applicability before handoff; do not rewrite the checkpoint.

## 2026-10-15 — RP-596 complete candidate population passes; current-source control running

Resume session51264, not a new candidate run. It terminates exit0: all300 governed keys
complete, no run/aggregate invariant failures, wall396,571,846,833ns (6m36.572s), whole Go
command397.087s. Canonical complete-phase report SHA256 is exactly
`1ffa47cfd4f85f6daf1b340cf6e43cfa7d534cb39308cdf80e955c9d28be826d`, matching the historical
active result with no golden/report regeneration. The earlier source/scenario/policy/constants/
worker and full-population controls apply. This is standard pacing only, not complete maintenance.

Candidate CPU profile SHA256 `7dadb25e49a116d9ef4ac557ea5a98069c2e0a8dc87995a052fb16befedc24ef`;
compiled binary `c0d25a91203beab830568057ef1edbe8c24171cd5a17758338a05e8c4f34a705`.
Both remain in `/private/tmp/cloud-clicker-rate-candidate.2kNxJt/` with the unchanged hashed
candidate/overlay. The root profile reader executes ordinary/focused cumulative views:
3163.77s total sampled CPU; ratesWithProvisionedAndPolicy408.83s cumulative (12.92% of samples).
The historical profile was2285.94s/39.05%, but those overlapping sampled stacks are not
disjoint savings, wall-time percentages or a matched current-source performance result.

**Source attribution correction:** the first ordinary `-list ratesWithProvisionedAndPolicy`
read the unchanged active engine at the overlay's recorded filename. Its displayed line text
is wrong for this binary; discard that line-level interpretation. Executed corrected reader:
`make harness-profile-report
HARNESS_CPU_PROFILE=/private/tmp/cloud-clicker-rate-candidate.2kNxJt/cpu.pprof
HARNESS_PROFILE_FLAGS='-list ratesWithProvisionedAndPolicy -nodefraction=0 -trim_path=/Users/stronk/repos/cloud-clicker/server/production -source_path=/private/tmp/cloud-clicker-rate-candidate.2kNxJt'`.
It now displays the hash-checked candidate buckets/target-filter code at the matching lines.
Function-level sample totals are unchanged. Canonical docs explain this overlay-specific trap;
do not use unchanged checkout text as a witness of experimental compiled source.

Actual candidate-specific full vet passes through the root target:
`make vet GO_PACKAGES='-overlay=/private/tmp/cloud-clicker-rate-candidate.2kNxJt/overlay.json ./production ./multiplier'`.
Active engine, retained patch and both benchmark-source hashes remain unchanged. Full-range
whitespace and patch-applicability checks pass after the narrow patch-data attribute correction.
No real Postgres, Go/TS full release-artifact or player proof is newly claimed by this phase.

**Live current-source control:** session20829, started after candidate/vet/profile work finished:
`make test-go GO_PACKAGES=./harness
GO_TEST_FLAGS='-run NoTests -bench BenchmarkStandardPacingProfile -benchtime=1x -timeout=60m -cpuprofile=/private/tmp/cloud-clicker-rate-candidate.2kNxJt/control-cpu.pprof -o=/private/tmp/cloud-clicker-rate-candidate.2kNxJt/control-harness.test -v'`.
It uses no overlay and reports the same Go1.27.2/darwin/arm64/M3 Max,14 logical CPUs,12 workers,
scenario and constants. This is one sequential candidate→control pair, not randomized repeated
performance sampling. No concurrent heavy workload is deliberately launched. The last poll
confirms the same handle live with no final result; resume it rather than restart on silence.
Require all300 keys and complete report identity, then report the actual paired timings and
sampling/environment limits. The historical783.821s alone is not that current control.

Current plan/RP-596 now route to completing that control and the existing landing/review gates,
not rerunning the successful candidate or broadening a standard phase into full maintenance.
Review by: Codex (candidate/measurement first filter). Recorded by: Codex. This result, earlier
candidate and packaging correction share the pending designated-review range after `9bdf9c29`.
No active optimization, kernel/history exception or owner/draft adoption, CI success, review
approval, archive, release, push or deployment. The same27 arithmetic failures and full1.0 floor
remain unchanged. A process-list read was sandbox-denied; the actual execution handle, not an
inferred PID or state file, supplies live/terminal authority throughout.

## 2026-10-15 — RP-596 matched pair and Linux/Postgres candidate checks complete

Resume control session20829; it exits0 with all300 governed keys, no run/aggregate failures
and the same full report SHA256 `1ffa47cfd4f85f6daf1b340cf6e43cfa7d534cb39308cdf80e955c9d28be826d`.
Active control wall735,604,548,625ns (whole Go736.065s), candidate396,571,846,833ns:
46.088989% less elapsed time /1.85491× observed ratio in this single sequential native pair.
Both use the same scenario/constants,12 workers, Go1.27.2/M3 Max and profiling options;
only the candidate uses the compiler overlay. No deliberate concurrent heavy run. This is not
randomized/repeated sampling, control of other host activity, hosted timing, registered
relevance or a completed full-maintenance observation. Retain the worse inactive/masked
microbench samples too. Do not promote the historical783.821s into this current control.

Control profile SHA256 `f49b4cf2a1779b16b439082abaed3dd1ee335b3b0b3bdcc0e7d87323066b6ed6`;
control binary `a8c7643378768ca020cdac2f0472be565fe37b397a43bee2ece737ff63716cdb`.
The existing root profile reader executes: total6069.22s sampled CPU, rate calculation
1412.48s cumulative (23.27%), versus candidate408.83s/12.92%. These overlapping sampled
stacks are not disjoint or wall-time savings. All four raw profile/binary files and the
candidate source remain in the already-recorded owned temporary directory for review.

After timing finishes, copy the identical candidate into the existing ignored `.cache/`
file `rate-traversal-candidate-2kNxJt-engine.go`; the Linux overlay there maps
`/workspace/server/production/engine.go` to that file. Its SHA256 remains
`d9b789badba9585010ecd332170ffe559f77c9ceb3665984fc81e1236bbb52ea`.
No new cache directory, active source replacement, fingerprint rewrite or hidden version bump.

Actual root `make test-go-ci CI_TEST_PACKAGES=./production
CI_TEST_FLAGS='-overlay=/workspace/.cache/rate-traversal-candidate-2kNxJt-overlay.json'`
executes the complete affected package cold on declared Linux/amd64 with real Postgres.
Session17438 exits2:106.616s, exactly the same27 named partition divergences, no additional
failure. This is a red package result, not full CI success or an arithmetic repair. No skip
census is claimed from non-verbose output; separately opted-in runtime populations remain
outside the claim. The local Go1.26-family image is linux/amd64 digest
`golang@sha256:eb36c1664dd974cde625f736e02c204383deebe03977365caaec5bf49f794348`.
This declared Docker Desktop lane uses the existing AVX2 workaround under emulation, not
native hosted-amd64 crypto/performance evidence.

Then actual root `make test-go-ci CI_TEST_PACKAGES=./gameserver
CI_TEST_FLAGS='-overlay=/workspace/.cache/rate-traversal-candidate-2kNxJt-overlay.json -run TestComposedGameserverReplaysRatifiedFirstHourAtPinnedSeed -v'`
session71342 PASSes (24.92s): real bootstrap, same script/milestones, both terminal transitions,
ordering and persisted Company/Founder replay. This is one pinned backend journey, not browser
or97/300 persisted journeys. Its existing post-delete census still exposes retained/public
leaderboard data (1155 events in this run); no retention-policy or data-rights approval follows.

All handles terminal. Inspect this turn's disposable tmpfs Postgres service (no durable mounts),
then stop/remove only it; final Compose project is empty. Preserve the shared Go cache/network,
ignored candidate copies and raw profile evidence. Active engine SHA256 remains
`b872d4746779a097f68f88475cef66214a2a7f8970817298f788ffad515ae875`; kernel/history/workflow/
Makefile unchanged. Full-range whitespace and unapplied-patch applicability checks pass.

Update the existing plan, RP-596 and execution queue to completed bounded measurement and
the actual next review/landing boundary; do not rerun these populations for reassurance.
Review by: Codex (implementer/measurement first filter). Recorded by: Codex. Consolidated
candidate/test/measurement/docs/records/patch-data attribute range after `9bdf9c29` through
this containing commit still needs designated Claude review and independent identity eligibility.
No guard-policy adoption, active optimization, false version, semantic change, archival,
publication/deployment, hosted-green or1.0 claim. Remaining accepted work/owner choices and
the full nine-tier/platform release floor are unchanged; this completed experiment is not1.0.

## 2026-10-15 — Clarify the next versioning-authority decision, not another experiment

At `d6e092e8`, all candidate/control/Linux handles are terminal and the tree is clean.
Rechecked the live roadmap, accepted plans, outstanding review provenance, KV-1 and actual
guard: later C14/consumer implementations are Codex-authored and cannot receive Codex's
designated approval. Held gameplay/public-wire/rights/author contracts are not implied
authority, and the successful rate experiment is complete. Do not restart it or claim all
remaining engineering is finished.

Amend my own existing history-guard draft with explicit proposed runtime-neutral eligibility
and the descendant-row flow needed to avoid a self-referential commit hash. Preserve real
semantic bumps, complete path/history inspection, exact other-party eligibility and fail-closed
publication. Reuse one coherent-batch verdict instead of per-edit predeclaration, duplicated
review or a fresh owner approval per eligible routine change. Finite matching reports are not
general identity proof; a revert does not erase a semantic miss. No accepted body is amended.

Deliver one owner direction question (general narrow class / historical UI-only / keep held).
No response/adoption is recorded; the tool's accepted delivery is not a ruling. Update the
existing decision queue. Normative cross-party review remains before acceptance; no guard,
exception, runtime, version, historical object, active queue authorization or other draft
is changed by this proposal. Diff/authority/links checked; no software suite for a draft-only
edit. Review by: Codex (draft author first filter). Recorded by: Codex. The new draft/decision/
record range after `d6e092e8` needs designated normative review; earlier implementation reviews
remain owed. No live process, self-approval, archive, push/deployment or1.0 completion.

Separate next safe check: actual CLI source still parses Git's display-quoted path output
and porcelain arrow text. Record RP-649 as unverified before probing the existing isolated
fixture. A confirmed filename blind spot would be ordinary KV-1 conformance work; it neither
needs this proposed exception authority nor establishes a current unversioned product change.

## 2026-10-15 — RP-649 preserve actual Git filenames in the existing version guard

Outcome under existing KV-1 at `f667bce0`: replace display-quoted/newline-split path handling
with NUL-delimited Git names. One shared diff helper disables rename collapsing for both the
history walk and actual-correction-target check; working changes union explicit unstaged,
staged and nonignored untracked lists. A move out of a watched prefix retains its old path.
Error diagnostics JSON-escape names rather than emitting literal control characters. No new
scope policy, semantic exemption, registry/correction mutation or version change.

Regression10309 runs the new isolated real-Git fixture against the unchanged old CLI and
fails:42/54 working-state cases are incorrectly accepted. Eight literal names ×prefix/exact
registration ×unstaged/staged/untracked, plus Unicode with quotePath disabled, retain plain
and unquoted controls. Quoted spaces, Unicode, quotes, tabs, newline, backslash and arrow
names account for the admissions; no actual unversioned product change is inferred.

Corrected full fixture35749 passes60 refusal cases, legitimate matching-version controls,
quoted historical miss→real forward correction, staged/committed move-out refusals and a
versioned sibling-branch move. It then passes the complete prior review-provenance, remap,
ancestry, archive, immutable-row, shallow and merge controls. The first corrected attempt99183
legitimately refused my fixture's plain review labels; use the existing bold metadata grammar
in that fixture, never relax the parser. No original-history-bypass execution is claimed from
the initial working-state failure; the later history/correction/move checks are new controls.

Both Node syntax checks pass. Root `make verify-ci-topology` passes26 negatives plus real
Make cold/focused/repetition/browser-dispatch checks. Actual `make verify-kernel-version`67846
passes checkout contract/fixtures, accepts the existing correction bindings, then still fails
the real50a3a514 six-path miss. Its trailing fixture leaf does not execute; the separate full
fixture result above covers that population, not a green aggregate. No full client/server,
Node24/Linux/hosted or release claim for this tool-only fix. The separate27 arithmetic failures
and unadopted history policy remain unchanged.

All handles terminal. The fixture owns/removes only its temporary Git histories; no main
history rewrite, services, user data or cleanup outside those existing fixture boundaries.
Update canonical CI behavior, the existing plan and RP-649; diff/whitespace checked.
Review by: Codex (implementer first filter). Recorded by: Codex. Entire source/test/docs/
plan/ledger/log range after `f667bce0` through this containing commit needs designated Claude
review, consolidatable with existing CI conformance repairs. The preceding draft proposal
keeps its separate normative/adoption gate; no self-approval, archive, push or deployment.

## 2026-10-15 — RP-650 literal filenames in baseline and epoch history

Outcome under existing Harness D6 / Leaderboards L8: use NUL-delimited, non-renamed Git
file lists, literal exact-file selectors and no filename trimming. Registered artifact names
retain their existing grammar; governance cannot interpret their quotes/whitespace/brackets.
No artifact, baseline, epoch, version, history correction, workflow or runtime change.

Old-code real-Git regressions: epoch session9127 misses six transient unregistered revisions
whose bytes are subsequently restored, and refuses six valid registrations (Unicode, quote,
tab, newline, backslash, trailing space); plain/interior-space/bracket controls pass. Baseline
session35034 refuses six corresponding legitimate reports, includes an unrelated bracket-pattern
match in report history, and admits three outside-path whitespace aliases plus a false input
alias. The alias test's later mixed-classifier assertion was not reached on old code. A dirty
actual bracket-named artifact already refuses on old code; the separate untracked bystander
test reproduces a false refusal, not a bypass. No actual repository artifact miss is inferred.

Corrected `make test-go GO_PACKAGES=./harness GO_TEST_FLAGS='-run TestEpochGuard -v'`
session83595 PASS7.041s; corresponding `-run TestRepositoryGuard -v` session31742 PASS7.661s.
Additional `TestRepositoryGuardLiteralGitProtocol` PASS0.406s: exact newline/trailing-space
names, empty diff, refusal of non-NUL output and failed Git query. Full cold
`make verify-harness-fast` PASS143.567s: entire harness78.737s, T0/T1 role0.314s,
Commons invariance0.382s and actual committed repository baseline/epoch guards. Final
`make vet GO_PACKAGES=./harness` PASS. The initially lost terminal response was recovered from
this thread's completed command record; the live full gate was not restarted. No unrelated
browser/server/full pacing study, Linux/hosted-green or release claim. The separate kernel
history failure at50a3a514 and27 numeric partition failures remain open.

Only isolated tests' disposable Git fixtures were created/removed; no services or main history
rewrites. Update canonical docs, plan and RP-650 once. Review by: Codex (implementer first
filter). Recorded by: Codex. Entire source/test/docs/plan/ledger/log range after `e10d24f4`
through this containing commit needs designated Claude review, consolidatable with existing
CI conformance repairs. No self-approval, archival, policy adoption, push or deployment.

## 2026-10-15 — Published maintenance38677066823 and current rollup reconciliation

At clean implementation source `7ad33db8`,89 commits after published be66902b,
read-only `gh api .../branches/main`, `gh run list` and run/job queries confirm
the unchanged remote SHA/failed CI38560135778 and a newer completed failed
maintenance38677066823 on that SHA. No remote mutation, rerun or push. Inspect
job116232970495's log: the unchanged50-minute wrapper interrupts harness-observe
(exit124). Job116232970706 finishes red on career/relevance report comparisons;
its logs retain97 careers per arm, the six casual starter ties and970
admitted relevance arms. Numeric job116232970746 passes. Do not regenerate reports
or label the unapplied local candidate's timing a hosted improvement.

Artifact11948536992 is1196 ZIP bytes; its upload log reports ZIP SHA256
`a93a66527b3d645f1ee4f2ee48f072f8dfc92be64c31fd9a1e1aa4d37a0db2ec`.
Retain its exact JSON as `observations/38677066823-harness.json`, confirmed with
cmp against the download; JSON SHA256
`d2fa29b77fc69aac5bb9293b516e4d5d55d8ea045bae5f46279bec9425fab6de`.
It records incomplete/signal,2998910ms overall; repository guards complete20963ms;
standard pacing300/300 complete2750518ms; relevance0 incomplete80/107 with1083433
executed transitions in227414ms; declared relevance1 never starts. Unknown guard/
exclusion/truncation fields remain unknown, not clear. Actual root
`make harness-observation-check HARNESS_OBSERVATION=planning/scaffolding-and-ci/observations/38677066823-harness.json`
rejects it: incomplete harness observation, CLI1/Make2. The first sandboxed download
cannot connect to GitHub; one narrow approved read-only download succeeds. No
study is rerun locally for this evidence update.

Reconcile current-state/roadmap/queue coordinates and superseded next actions:
kernel0.3.217/epoch8, delegated Clout wording, actual later native/client/composed
checks, source-version/authorization/expiry/query repairs and unaccepted calculation/
search drafts. No new playable tier, full CI, draft adoption, designated approval
or release. The state remains incomplete Garage hardening plus substantial later
game/platform construction. Earlier checkpoints/results and all review debt remain.
Editorial/source checks and the real artifact consumer above are the verification;
no unrelated software suite, services or durable-resource removal. Existing owning
implementation ranges remain pending designated review. Review by: Codex (evidence/
editorial first filter). Recorded by: Codex. Goal remains full nine-tier1.0.

## 2026-10-15 — Shared full/focused Linux browser execution

At `72fb7849`, native Firefox cannot connect for the new numeric tests; the existing
Linux browser command has no focused selector. Under RFC-0000's affected-boundary
procedure and CI D2, extract its unchanged apt/Make/Corepack/pnpm-frozen prelude into
`client/tools/test-browser-ci.sh`. Keep the same image, CI environment, fresh anonymous
dependency volumes and ten-minute KILL containment. The default command still invokes
the complete root `test-browser`; new `test-browser-ci-focused` forwards one Make
selector assignment to existing `test-browser-focused`. No workflow, browser version,
population, worker, assertion, deadline or performance budget changes.

Execute the real commands against Docker28.4.0 linux/arm64, Compose2.39.4-desktop.1,
cached Playwright1.62.0 image config5e63ffc997394026acb443c255703c7278f97d43cdc3d49cf51bc8fdce7d5383:

- Focused numeric vectors + ordered transforms: all three engines,19212 tests/6 files,
  PASS7.36s after cold setup. This supplies RP-662's missing Firefox execution, not
  a fix of the native launch or RP-663's retained underflow discrepancy.
- Deliberate nonexistent file selector: actual Vitest reports no files and exits1;
  inner/outer Make return nonzero (outer2). No pass-with-no-tests behavior.
- Full `make test-browser-ci`: PASS exit0. Broad480 files pass40119 ordinary tests,
  with12 expected failures/24 skips,196.06s; native Typer3 files pass87 ordinary/
  6 expected,3.89s; isolated Chromium performance passes5 with105 selector skips,
  6.24s. No retry, dropped case or changed deadline. These phase times exclude setup
  and are not a hosted/five-minute estimate or proof RP-637's recurrence is repaired.
- `make verify-ci-topology`: PASS, including existing26 negatives, four wrong full
  defaults, five setup/containment/drift controls, actual script execution with owned
  tool doubles (full/focused argv, invalid/missing targets, setup/test error propagation),
  and actual Make selector/report expansion. `bash -n` and diff checks pass.

The topology consumer now reads the shared script and keeps the full default bound;
a focused command cannot silently replace CI. Test doubles prove command forwarding,
not browser execution; the real runs above supply that evidence. Root Docker socket
access is separately approved; no daemon mutation, publication, deployment or secrets.
Compose confirms no browser containers remain. Only owned disposable fixture files
are removed. No Go/product/kernel/balance bytes change or unrelated SQL/server check.

Canonical CI docs, current plan and the overview's local-browser row reconcile once.
All eleven changed paths after `72fb7849` through this containing commit need designated
review, consolidatable with the CI lane. Review by: Codex (implementer first filter).
Recorded by: Codex. No acceptance/archival or full-CI/1.0 claim. Accounting/search,
historical kernel failure, rights/accessibility and the full nine-tier build remain.

## 2026-10-15 — Shared browser runner designated review: CHANGES REQUIRED

Review by: fresh independent Codex sub-agent `/root/linux_browser_runner_corrective_review`.
Recorded by: Codex parent, transcribing its verdict under Marco's corrective-review delegation.
Reviewed complete eleven-path range:
`72fb7849bf6e5630d4b76bba2dfbb7199d9cb663..d81713b5ff4f589e3413a07daab9c0bee2e14871`
(exclusive/inclusive). Verdict: **CHANGES REQUIRED**, RP-665/P2.

The reviewer executes the existing `make research-browser-coordination`: its shell
override now reaches the persistent restricted test entrypoint and exits2 before
research, printing "expected test-browser or test-browser-focused". The docs also
incorrectly claim command overrides replace the timeout wrapper. This real consumer
was absent from the previously passing topology fixtures.

Independent topology/setup controls PASS; actual selected Linux numeric vectors/
ordered tests PASS19212/6 engine-file executions,4.69s; nonexistent-file selection
correctly fails Vitest1/outerMake2. Bash syntax, Compose config, range whitespace
and actual full Make routing PASS. Reviewed implementation bytes remain unchanged.
No full browser rerun or hosted/timing/feature acceptance claim. Reviewer changes
no repository files; its owned temporary fixtures and browser containers are gone,
and it preserves the other reviewer's Postgres resources. The repair below is not
silently absorbed into this verdict.

## 2026-10-15 — RP-665 restore the manual browser research consumer

Under existing CI D2/RFC-0000, explicitly select the shell entrypoint in the existing
manual Make target, preserving its pre-extraction non-CI setup/timeout semantics.
The full/focused CI entrypoint, default population and containment remain unchanged.
Add a topology negative restoring the broken command and reconcile the docs. Move
new instrument output to ignored `.cache/browser-coordination-observation.v1.json`
so this real rerun cannot replace the tracked historical observation; population,
assertions, guards and source/instrument identity fields stay unchanged.

Initial sandbox Docker access refuses before execution. The narrowly approved actual
root target then completes exit0: Chromium/Firefox/WebKit each40 interval cases and
10 actual page terminations, total120/30; no runtime errors, exclusions or guard firing.
This is Linux/arm64 primitive research, not renewal/product/timing acceptance. Source
coordinate is `6d2f42eb` plus this working diff; instrument SHA256
`ce78dad29a60a1fa1beb1341a8dc75227886c7b37b85354b61edd44662478791`.
Ignored report SHA256 is `07585030a6234f41c7cceabcdebbe585c4cddef9abe26ab4f58a68b250a1e221`.
Historical tracked report is byte-unchanged. Actual topology/negative/setup/Make
fixtures PASS (27 topology negatives), instrument syntax and diff checks PASS.

Eight-path Make/tool/test/docs/plan/ledger/log supplement after `6d2f42eb` through
this containing commit needs designated corrective review, together with the original
runner range above. Review by: Codex (repair implementer first filter). Recorded by:
Codex. No product, workflow, arithmetic/kernel or acceptance changes; no whole-CI,
hosted, renewal, release, archival or publication claim. Browser container is disposable;
the other reviewer's Postgres service is untouched. Existing red gates remain.

## 2026-10-15 — RP-665 and shared browser runner re-review: APPROVED

Review by: independent Codex sub-agent `/root/linux_browser_runner_corrective_review`.
Recorded by: Codex parent, transcribing its returned re-review under Marco's delegation.
Reviewed complete ranges (base exclusive, tip inclusive): original eleven-path
`72fb7849bf6e5630d4b76bba2dfbb7199d9cb663..d81713b5ff4f589e3413a07daab9c0bee2e14871`
and eight-path supplement
`6d2f42ebbf5db83f40e429aa600b79791a4884fc..3f8ed26e80d09e114a58f08fd8d7ffb619ab75a1`.
Verdict: **APPROVED**, RP-665 resolved, no remaining actionable findings in these
ranges. The preceding negative verdict remains the record of the caught regression.

Independent actual repaired manual target PASS on Linux/arm64: all three engines,
120 interval cases/30 page terminations, no exclusions/runtime errors/guard firing;
historical tracked report unchanged. Ignored report SHA256
`feba8ab5187525f1549e111c6ba7ddefe0448a3a0532d5daacdc0e12675ddcc4`, source `3f8ed26e`.
Actual topology PASS27 negatives including restored broken command, executable
setup/error controls and Make selectors. Instrument syntax, range diff and actual
Make routing PASS; full/default/shared-runner bytes unchanged. Earlier independent
19212-test Linux focused pass and missing-file failure remain applicable, not
relabelled as a second/full execution. No full-browser repeat for this caller repair.

Reviewer changes no source/tests/records, cleans its fixtures, verifies no browser
containers remain and touches no other services. Later unrelated review records
are outside the verdict. No hosted timing, renewal, full CI, feature, archival or
release acceptance; unrelated review debt and historical/arithmetic red gates remain.

## 2026-10-15 — RP-649/RP-650 independent corrective review: APPROVED

Review by: fresh independent Codex sub-agent `/root/literal_git_paths_corrective_review`.
Recorded by: Codex parent, transcribing its returned verdict under Marco's delegation.
Reviewed complete eleven-path range (base exclusive, tip inclusive)
`f667bce01d8eebf7b1db77786a8d0b513e7bcf95..54f352f62759a4ac7aa49fbf48e9f18e1c242b60`,
exactly commits `e10d24f41222678b6b1b4feaff95eed18989762a` and
`54f352f62759a4ac7aa49fbf48e9f18e1c242b60`. Verdict: **APPROVED**,
no actionable findings; corrective literal-path work only.

Independent old JS CLI reproduces42/54 wrong admissions. Corrected fixtures pass
60 refusals plus valid version/correction/history/rename controls. Old Go overlays
reproduce six transient-epoch misses, six valid-registration refusals, six literal-
report refusals, bracket-pattern false matches, three whitespace aliases and one
false input alias. Corrected focused Go passes uncached9.399s. Additional review-
only overlay tests pass0.515s: exact Unicode/bracket/tab/newline/trailing-space
rename endpoints and rename-out/restore cannot conceal unregistered epoch history.
Full cold `make verify-harness-fast` passes actual history guards (harness82.920s,
role0.272s,Commons0.376s). Harness vet, Node syntax, Go formatting and range diff pass.
At verification HEAD264842b4, the six reviewed guard/test files are byte-unchanged
from54f352f6 in both HEAD and worktree.

Actual `make verify-kernel-version` still fails historical
`50a3a5141d5c4c1cd25cbf3c9c39d9e3cf8e2444` versus parent
`0cf9f7a6aba4038fadcdf35e5b94f56986164af7`, same six minigame paths.
Its trailing fixture leaf is not reached; separately executed fixtures pass.
No historical waiver, false bump, rewrite, RP-648 citation approval or adoption of
the unaccepted kernel-history policy. No DB/container/Linux/hosted, full CI, feature
or release acceptance. Reviewer makes no repository edits/commits/worktrees or
publication. Review-only files remain in private temporary storage; no cleanup is
claimed for them. Parent records this verdict without rerunning unrelated software
for the Markdown-only transcription; factual ranges, diff and whitespace checked.

## 2026-10-15 — Four verification corrections independently approved

Review by: fresh independent Codex `/root/ci_verification_corrective_review`.
Recorded by: Codex parent under Marco's corrective-code delegation. All **APPROVED**,
complete source/test/docs/planning/ledger ranges (7/6/6/7 paths respectively):

- RP-623: `1f1a54818a07f4da5382a1568617ce6d7cbcda8c..ac88de36f6db36d78eae4e0c739350c6f13069a4`.
- RP-633: `c05a196af1a9f10d318da828ec3dd75fbbe8441d..15e82e384b85960c2c6d06a79bb139bac02ca1e5`.
- RP-635: `a201a7e85e5cfbde50f9060fe11d27f105f402d7..ab3442dc31328557b924a104578d3e332e5856b1`.
- RP-648: `691e9bac2b82215259bfd4a4f2e48eff5ff8dc20..59787fc34f3ac41778db9d5fe0df08255ede663f`.

At descendant a45f5478/product312f6e02/kernel221 actual topology27 permanent negatives and
existing Linux/Make controls pass. Actual cold Replay/Pitch/Typer comparisons pass (.517/.329/
.190s); explicit count2 really executes replay twice (.899s). In-memory exact historical guard/
subject comparisons reproduce old admission of4 missing-cold flags and3 invalid maintenance
recipes; corrected tips refuse all7, healthy controls pass. Initial comparison wrongly paired
old guard with modern browser Compose and fails its existing browser contract; corrected
historical subjects supply evidence. No source mutation or changed corpus/golden regeneration.

Cold observation Harness/CLI pass (.214/.241s), all9 identity cases including6 forged refusals;
2 opt-in research studies explicitly skip. Complete CLI6 tests (.490s) execute real23-run
registered-row/golden and SIGTERM→invalid-incomplete-artifact checks. Affected vet passes.
Full kernel adversarial fixtures pass, including6 range negatives/6 genuine legacy/completed/
enclosing/remap positives and60 literal-path controls. Actual root kernel gate passes checkout
contract/adversarials then exits2 at unchanged50a3a5141d5c4c1cd25cbf3c9c39d9e3cf8e2444
versus0cf9f7a6aba4038fadcdf35e5b94f56986164af7, same6 paths. Trailing fixture recipe is not
reached; separate full fixture execution supplies that evidence. Syntax/4 range diff-checks pass.
Old corrupted-corpus and Go-validator failures inspected, not rerun; observation consumers
unchanged since their tip. No edits/services/worktrees/overlays/browser/Docker/full-study,
hosted/performance/full-CI or feature/release approval. All reviewer handles terminal.

RP-648's draft author reconciliation is inspected, NOT normative adoption or scope-exception
eligibility. Parent independently queries GitHub: published main stillbe66902b659d59a5572037a9b6431d2ac02fb731;
latest CI38560135778 and maintenance38677066823 remain completed failures. CI browser/composed/
schema/harness succeeded; client/server failed. New local work is not hosted evidence. Parent
asks Marco about the proposed exact-commit/path exception class while accepted work continues;
no answer or adoption inferred, no exception implemented, guard weakened, rewrite or push.

## 2026-10-15 — RP-121 reused-database fixture correction approved

Review by: fresh independent Codex `/root/ci_guild_fixture_corrective_review`.
Recorded by: Codex parent under Marco's corrective-code delegation. **APPROVED**, all5 paths
of `604f2a826b652a39f2fa6d96d7a3fae19f4af40a..b976110b751969702a17b4fddaebeb21cc0b47af` inspected.
The fixture requires actual current_database()=cloud_clicker_test before migration/TRUNCATE,
and registered cleanup precedes connection closure. Existing lifecycle/concurrency/history/
deletion/Health/clearing assertions are unchanged, with no shared-service t.Parallel. The
maintenance-database negative only connects/pings/SELECTs and requires the exact refusal
sentinel, not an arbitrary connection/query failure. No runtime/retention-policy change.

Actual descendant16219ad9/product874d8d17/kernel221: declared Linux-amd64/Postgres four
integration parents pass count2 (2.091s), then a separate shuffled12345/count2 process on
the SAME live service passes (1.975s). Ordering actually changes; all3 original populations
execute4 times, plus4 exact-sentinel refusals, no skips or service/DB reset. Container ID
7e4d5f61067a remains identical across both invocations. Cold native Guild .274s and vet pass;
native DB skips excluded. Range diff-check passes; test blob776e06f00463d4ab5058624d48e00a64625e2545
matches its reviewed tip. Old failure inspected, not re-manufactured; no edits/mutations.
Initial denied Docker inventory runs no tests. All processes terminal; only created Compose-
labelled no-volume/tmpfs service removed after inspection. Network/caches/volumes/BuildKit
preserved, final declared inventory empty. No browser/player/full-server/hosted-CI, erasure,
normative, feature, archival or release approval; RP-120 stays separate.

## 2026-10-15 — RP-645 deadline fixture independently approved

Review by: fresh independent Codex `/root/ci_deadline_fixture_corrective_review`.
Recorded by: Codex parent under Marco's corrective-code delegation. **APPROVED** complete
five-path `728508de50671a77603d891adb1a95282ee4fbdd..b45dcd1d9faadaa325f9a92ac5f2388eba326d78`.
Both test files, backlog and CI plan/log inspected; no production/workflow/retry/timeout/
kernel/copy/normative changes. Files remain byte-identical through executedcc37b6d3.

Setup establishes subscribed-socket or deliberately held-initial-read before the unchanged
actual50ms test-body deadline. Successors retain exact abort reason, phase observations,
teardown/closed sockets and refusal of late navigation without intents/native actions/new
sockets/remount. Blocked Codex Close retains its original assertion/cause after cancellation.
Actual typecheck68217 passes0/0; native Chromium/WebKit52253 has20 ordinary passes plus4
deliberately expected deadline failures,216 deselected (2.24s). Declared Linux ARM64
Playwright12433 has30 ordinary passes plus6 expected failures,324 deselected (3.87s).
These expected failures test cleanup, not gameplay success. Independently parsed retained
full-run JSON byte counts/SHA256 match the record: success/zero failures, all15 Reputation
and3 blocked-Codex controls present. Their JSON labels expected failures as passed; not
ordinary successes. Complete-range whitespace checks pass.

Historical failure inspected, not source-mutated. Initial native listen denial and first
malformed nested Linux selector execute no tests; corrected commands supply evidence. Linux
stdout partly truncated; terminal counts retained. All handles terminal, --rm container absent,
declared browser Compose empty, no other resources removed or source edited. RP-637 cause/
reliability remains OPEN. No new full-suite/performance/hosted-AMD64/timing/real-SQL/feature/
policy/archival/release acceptance; original failed evidence remains intact.

## 2026-10-15 — Original Reputation host lifetime correction independently approved (RP-531)

Review by: /root/reputation_host_lifetime_corrective_review (fresh independent Codex under
Marco's corrective-code delegation). Recorded by: Codex root, transcription. APPROVED complete
6a6a996d782778dc64dd3ded4085cea6d0e0fc19-exclusive through
7f8e3126ae835ce31510208032985eb192bfd64a-inclusive, all5 paths. Accepted CI D2 only;
no product/workflow/retry/timeout/concurrency/copy/normative change. All original88 ordinary
host cases retain input/outcome assertions. Disposal binds original TestContext, is idempotent,
drains native input before teardown and refuses expired navigation before polling/focus.
Existing shared-helper controls check drain/delivery errors and untouched successor input.

Historical qualification: original50ms fixture assumed socket readiness without establishing
it; its earlier native pass was not universal proof. Separately approved complete RP-645
amendment (preceding verdict) establishes socket-ready/held-initial-read populations BEFORE
the unchanged actual50ms body deadline, then checks exact abort, detached UI/closed created
sockets, no intents/late navigation/remount. Current executions use that corrected population;
this approval does not assert the original fixture was sound or erase its failed evidence.

Independent full host/plan/Typer native Chromium/WebKit31823 passes278 ordinary plus10
explicitly expected diagnostic deadlines,288total/no deselection,26.56s. Declared cold Linux
ARM64 three-engine service54780 passes417 ordinary plus15 expected deadlines,432total/no
deselection,31.50s. Root make test-browser-focused/test-browser-ci-focused use the same
three test files; make typecheck77218 passes0errors/0warnings. Complete range whitespace passes.
Initial sandbox listen denial executes no tests. Partial stdout retains terminal censuses,
not a full-trace claim. Source unchanged through23fcf989; host blob
5ac2cef441f81903c12852bd3871a1673c4eacab. No source mutation/new old-source reproduction.

Real browser/DOM/runtime/parser with controlled network/storage, NOT real server/Postgres/
socket service/minted content. No full suite/performance/hostedamd64/latency/whole-CI,
RP-637 reliability, feature/adoption/archival/release approval. All handles terminal;
--rm container absent, declared browser Compose inventory empty; no manual resource removal,
tracked edits or commits by reviewer. Complete RP-531 range leaves corrective review debt;
RP-526/RP-637 original causes, other ranges and release/whole-product gates remain.

## 2026-10-15 — Retain independent Firefox screens setup failure (RP-673)

During fresh /root/clout_cap_presentation_corrective_review's RP-536 check, declared
Linux ARM64 three-engine axis/screens54143 finishes RED:395pass/1fail/15 existing performance
exclusions. Axis file's81 executions pass. Firefox screens `late http-conflict completion
cannot refresh an unmounted host`, line235, observes runtime.requests length0 instead of1
immediately after native Enter/setup, BEFORE the late response/unmount property. Root reads
the actual fixture: it has no axis_stack, so the changed cap preflight is not invoked. This
separates the observed failing boundary; it does not establish why dispatch failed or approve
the screens file/late-completion claim. Product remains137f7a8d; no source mutation/rerun/fix
or input/deadline substitution by root. Original red is retained, not erased by a later pass.
CI D2 next action: inspect the actual native dispatch/readiness/lifetime evidence, then add
a meaningful regression only once the mechanism is known. Full CI and earlier failures remain
open. The RP-536 review retains its own exact scoped verdict and other evidence in Clout log.

## 2026-10-15 — Observe native late-completion setup without concealing RP-673

CI D2 diagnostic only: preserve all four original native actions/request/unmount/late-response
assertions and deadlines. On failure retain stage, pre/post input/settle readiness, focus,
subscription and actual native events, reading defaultPrevented AFTER dispatch. Wrap the
original assertion as cause; no retry, refocus, alternate input or extra settle on failure.
One separately named canceled-Enter control uses the actual mounted enabled button and
trusted native input, requires zero requests at the original assertion, exact cancellation,
intact focus/readiness and preserved cause. That control is NOT a late-completion success.

Actual root make test-browser-ci-focused with screens file and -t 'late .*completion cannot
refresh an unmounted host',50271: PASS2.96s,12 ordinary+3 canceled-input controls,318 explicitly
deselected. Initial sandbox socket refusal executes no tests; narrow permitted command runs.
Complete affected axis/screens with same root target38651 PASS21.61s,393 ordinary+6 diagnostic
controls (3 canceled Enter/3 existing blocked Codex Close),15 existing performance exclusions.
No new diagnostic deadline tests. make typecheck9040 passes0errors/0warnings. Tool truncates
bulk module/worker stdout; terminal census retained, not a complete trace claim. Original
RP-673 does not recur, so its cause/reliability remains OPEN; no repaired product or CI claim.

Only one test file/docs/ledger/owning plan/log changed. No shipped code/build/performance,
schema/math/copy/workflow or deadline change; unrelated lanes not rerun. Review by: Codex
(diagnostic implementer first filter). Recorded by: Codex. Complete range begins
5afbedcd-exclusive through this containing commit; independent corrective review pending.
Next consume genuine recurrence evidence, not unchanged green repetitions; continue other
accepted construction. All handles terminal; Docker is declared ARM64 Desktop, not hosted
nativeamd64. No real SQL/service/clean-host/full-CI/feature/archival/publication proof.

## 2026-10-15 — RP-673 diagnostic fidelity independently approved; incident remains open

Review by: /root/late_completion_diagnostic_corrective_review (fresh independent Codex under
Marco's corrective-code delegation). Recorded by: Codex root, transcription. APPROVED complete
5afbedcdc99b71a3e964c4fb3d54239f7d95a351-exclusive through
9a551e61113ccbb8cecbf1fe97966ccb854fad62-inclusive, one commit/all5 paths. Original four
native journeys/actions/settles/assertions/deadlines preserved. Canceled actual native Enter
on an enabled focused control reaches original request assertion with zero requests, observes
trusted delivered cancellation after dispatch and preserves AssertionError cause. Ordinary
failures rethrow with diagnostics; listeners and original fixture lifetime clean up in finally.

Independent make typecheck90117 passes0/0. Declared make test-browser-ci-focused with both
screens/axis files14041 exits0,20.12s: all6 engine/file populations,393 ordinary+6 diagnostic
controls (3 canceled Enter/3 blocked Codex Close),15 existing performance exclusions. Complete
range whitespace passes. No retries/source edits/mutations/commits by reviewer; both handles
terminal, disposable service absent and clean source at9a551e61. Docker LinuxARM64 Desktop,
not hostedAMD64; runtime doubles, not SQL/real WebSocket gameplay. Original Firefox failure
does not recur; RP-673 stays unexplained and OPEN. Approval covers diagnostic fidelity only,
not reliability/full CI/new policy/feature/archival/release acceptance.

## 2026-10-15 — Rate-traversal candidate review CHANGES REQUIRED (RP-674)

Review by: /root/rate_traversal_corrective_candidate_review (fresh independent Codex under
Marco's corrective-code delegation, bounded conformance candidate only). Recorded by: Codex
root, transcription. CHANGES REQUIRED complete4-commit/8-path range
9bdf9c29f119e57366fd2f55dacfbe05f89a2127-exclusive through
d6e092e8d14ebb15d0aa905071ad2e87ae02fb65-inclusive. All patch/tests/docs/records/attribute
paths inspected. Candidate sort.Strings bypasses multiplier.OrderedSourceIDs despite matching
today's order: implemented Production Contract Integrity D1 requires that neutral helper in
production; D3 binds it to published formula authority. This is an unapplied-candidate defect,
not a new active-runtime failure. Keep helper ownership; no specification amendment needed.

Independent selected candidate exact-bit/order/refusal/input/activation/simulation/evaluation/
accrual/replay controls pass0.843s; full Decimal/Multiplier0.564/0.169s pass, candidate-specific
Production/Multiplier/Gameserver vet passes. Actual declared Linux-amd64/Postgres composed
first-hour18499 passes21.81s, persisted replay/terminal transitions. Post-delete retained public
data diagnostic still fires: reviewer observed1149 events,
6boards/6markers/2queue/2archives; no privacy closure. Original source/patch/both candidate
copies/all4 profile/binary hashes match; binary metadata is Darwinarm64/Go1.27.2. Complete
range whitespace and patch applicability pass. Filtering sorted buckets otherwise preserves
the original multiply/order/count/validation/activation sequence, without cross-call cache.

Historical396.572/735.605s full300 observations inspected, NOT independently rerun. No raw
report file available to independently rederive its digest without another full experiment.
No broad-domain equivalence, new guard policy/exception eligibility, active landing, accounting,
CI/feature/archival/release approval. All handles terminal; only owned twice-inspected tmpfs
SQL3a8f3938… removed; original2BuildKit preserved. Reviewer edits/commits no tracked bytes.

Corrective scope under the SAME D1/D3/C2 (before correction checks): retained patch now calls
the neutral helper once per slot bucket. Active engine stays unchanged; preserve original
ignored candidate/profile evidence. New candidate source1bb2a62d5c4e42595e89e8b0ddab5565f7572830137679c5141359e9bea377f6,
Git blob645b203492ac887ad5f7dac93ed0e199c13fa2e1. Use separate ignored helper-bound and
helper-sever overlays, never mutate main runtime. With the neutral helper returning empty,
old candidate should ignore the sever and pass; corrected candidate must fail the existing
literal exact-order oracle, then pass intact. Execute affected checks/actual SQL journey and
candidate-specific vet. Full300 source-bound measurement is required for the changed candidate;
prior timing/report identities cannot be relabelled. No new tolerance, version, formula artifact,
budget, production or workflow change. Root first-filter/re-review still pending.

Root correction checks terminal: old-helper-sever overlay40126 leaves the old candidate's
existing exact-order oracle green; corrected-helper-sever96577 fails4 active cases with
raw-bit differences, all-zero control passes. No compile fault or main-tree mutation; this
directly witnesses the previously missing neutral-helper dependency. Intact corrected
rate/refusal/activation/input/simulation/evaluation/accrual/replay controls14101 pass;
full Decimal/Multiplier pass0.403/0.063s, candidate Production/Multiplier/Gameserver vet35320
passes. Corrected actual declared Linux-amd64/Postgres first-hour17548 passes23.85s with
persisted replay/terminals; retained public-row diagnostic has1152events, no privacy closure.
Only root-owned twice-inspected tmpfs7dd35240… is stopped/removed; its disposable SQL data
is unrecoverable, persistent volumes/caches/builders preserved.

New complete source-bound measurement4365 is LIVE, not yet a result. Root command:
make test-go GO_PACKAGES='./harness'
GO_TEST_FLAGS='-overlay=/Users/stronk/repos/cloud-clicker/.cache/rate-traversal-helper-bound-overlay.json
-run NoTests -bench BenchmarkStandardPacingProfile -benchtime=1x -timeout=60m
-cpuprofile=/Users/stronk/repos/cloud-clicker/.cache/rate-traversal-helper-bound-cpu.pprof
-o=/Users/stronk/repos/cloud-clicker/.cache/rate-traversal-helper-bound-harness.test -v'.
Startup confirms12workers/Go1.27.2/Darwinarm64/M3Max/14logicalCPUs and original scenario
10aa5c8a…bb07a/constantsbaa89050…128f6. No deliberately competing heavy work starts.
Resume that exact handle. Require all300 keys/no failures and compare full report digest
against original1ffa47cfd4f85f6daf1b340cf6e43cfa7d534cb39308cdf80e955c9d28be826d;
this is standard pacing only, not complete maintenance/registered relevance/hosted timing.
Patch19d781d9f79f10c8c428a7a46ba2197a1af7d14ff308b15ffde66ae400a1efb0 applies cleanly;
new candidate1bb2a62d…77f6 and active engineb872d474…ae875 remain unchanged.

Read-only GitHub API independently reconfirms mainbe66902b659d59a5572037a9b6431d2ac02fb731.
CI38560135778 stays completed failure (client/server fail; browser/composed/schema/harness
pass); maintenance38677066823 completed failure (harness/reputation fail, numeric passes).
SSH ls-remote cannot resolve github.com; successful API branch/run/job replies supply
the actual metadata. No fetch, push, dispatch, new hosted execution or local-CI-green claim.

## 2026-10-16 — Corrected helper-bound candidate: complete standard pacing, not active landing

Measurement4365 above is now terminal exit0: all300 distinct standard-pacing keys complete,
nonempty outcomes, no invariant/aggregate failure. Benchmark449209272000ns/op (449.209272s),
package449.680s; full canonical report SHA256
1ffa47cfd4f85f6daf1b340cf6e43cfa7d534cb39308cdf80e955c9d28be826d exactly matches the original.
The benchmark's actual completeness/assertion path executes; no population/threshold changes.
Scenario10aa5c8a2dd8a5c79a5902e74f4e55a1ca3ab46a2d35c9bf3dd25fc729dbb07a,
constantsbaa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6.

Corrected candidate1bb2a62d5c4e42595e89e8b0ddab5565f7572830137679c5141359e9bea377f6;
patch19d781d9f79f10c8c428a7a46ba2197a1af7d14ff308b15ffde66ae400a1efb0.
Retained ignored .cache/rate-traversal-helper-bound-cpu.pprof SHA256
ebace88af3a17ddab7517ab4657f133b11b73fbd549edf677f4d5e4c6cbf8369 and
.cache/rate-traversal-helper-bound-harness.test SHA256
b9ca218aad67f6dbabaf609110de4e7e77c1a21dbf297796eff3200e70c4fd13.
Binary metadata confirms Go1.27.2/Darwinarm64; no embedded VCS revision claim.
An early binary read found no file before terminal publication; optional pgrep was denied
by system-monitor permission. Neither is liveness/termination evidence; the actual4365
handle stayed authoritative and completed without restart. No competing heavy check was started.

This is standalone current-source evidence, NOT a matched improvement versus735.605s:
Decimal/replay/intent/simulation dependencies changed since the old pair. Historical
396.572/735.605s belongs only to the original candidate. No raw canonical report file was
persisted, so digest equality is the executed benchmark output, not an independent file rehash.
No candidate-source line-level pprof claim, complete harness/relevance/maintenance, hosted
timing or CI pass. Active engineb872d474…ae875/runtime137f7a8d/kernel222/epoch8 unchanged.

Independent corrective re-review is executing for original9bdf9c29-exclusive..d6e092e8-inclusive
plus complete correctiona11446b4-exclusive through this record commit (including e973fcb1 and
c1545119). Original CHANGES REQUIRED remains historical; no self-approval, guard-policy
adoption, exception eligibility, active landing, archival or release promotion. Review of the
terminal record is still required; the reviewer owns any new temporary SQL fixture cleanup.

## 2026-10-16 — Helper-bound candidate corrective re-review: APPROVED, unapplied

Review by: /root/rate_traversal_corrective_candidate_review (independent Codex under Marco's
corrective-code delegation). Recorded by: Codex root, transcribing its terminal verdict.
Decision: APPROVED — bounded unapplied corrective candidate only; no actionable findings.
Complete reviewed union, base exclusive/tip inclusive:

- original9bdf9c29f119e57366fd2f55dacfbe05f89a2127..d6e092e8d14ebb15d0aa905071ad2e87ae02fb65 (4commits/8paths);
- correctiona11446b40fe9d4bd450d309c23b0c6e8fb9424ee..78631c312dac24629dd1cf2fd64e66db99b3e1f2 (3commits/9paths).

All11 unique paths inspected, including patch, benchmark test, canonical docs and complete
terminal/board records. Original CHANGES REQUIRED remains historical. RP-674's helper-ownership
defect is corrected under accepted D1/D3. Sorting each canonical slot bucket then filtering
preserves the old filter-then-sort multiplication sequence; validation/count arithmetic,
activation, generator/resource order and state ownership remain, without a cross-call cache.

Independent empty-helper overlay leaves original candidate green; corrected candidate fails
4 active raw-rate cases while all-zero control passes. Intact affected rate/order/refusal/input/
activation/simulation/evaluation/accrual/replay pass0.758s; full Decimal/Multiplier0.472/0.171s
and candidate Production/Multiplier/Gameserver vet pass. Actual declared Linux-amd64/Postgres
first-hour90390 passes28.10s with persisted replay/terminals. Retained-public-data diagnostic
still fires1149events; no privacy closure. Source/patch/profile/binary hashes independently
match the terminal record; active engine unchanged. Patch applicability/range whitespace pass.

Root's full300/449.209272s/digest output inspected, not independently repeated or reconstructed
from a raw report file. Actual benchmark enforces governed distinct keys, complete outcomes,
invariants and aggregate failures. Changed dependencies preclude a matched speedup claim.
No active landing, guard-policy adoption, scope-exception eligibility, formula-regeneration
approval, complete CI/maintenance, hosted timing, accounting repair, feature/archival/release
approval. Desktop emulation is not native hosted hardware. Reviewer handles terminal; only its
owned twice-inspected tmpfsd4360f17… removed, both preexisting builders preserved; no tracked
edits/commits. RP-674 candidate defect leaves corrective debt; remaining authority gates stay.

## 2026-10-16 — Observe completed archive phase in composed verification (RP-677)

Independent cursor reviewer /root/api_cursor_pairs_corrective_review discovers an actual
RED in current Linux-amd64/Postgres population81956: composed Exit/board's duplicate
VerifyStoredRun call returns log_gap. That method reads live rows; the real100ms worker
already verifies, archives and compacts in one transaction. A test trying to beat that
commit is not a valid phase boundary. This is a fixture race, not a proved lost archive
or a cursor-admission defect.

Reviewer-controlled ignored overlay waits for the REAL archive before the unchanged old
assertion: current resolver40146 fails1.046s, pre-RP-496 resolver61634 fails0.856s, each
archive_rows1/live_rows1/null_inputs0/live min=max seq3. An isolated old-resolver pass does
not prove stability; initial trailing-$ Make selector runs zero and is excluded. Diagnostic
source.cache/cursor-pair-review-exit-archive-diag_test.go SHA256
82fe32b94310ded45be2100029f7c2ff665aa3905c2fe8dbad219d7d87b39c78;
overlay JSON a6fe8df89d6f520ed88f77f2439f1788e9644a9a230ff4b21e7ec35f33539e94.
Exact root reproduction:
`make test-go-ci CI_TEST_PACKAGES='./gameserver' CI_TEST_FLAGS='-overlay=/workspace/.cache/cursor-pair-review-exit-archive-diag.json -run ^TestComposedGameserverExitVerificationAndBoardIntegration -v'`.
The old resolver control uses.cache/cursor-pair-review-exit-archive-old.json; both are
function/source overlays against current dependencies, not complete historical checkouts.

Under CI D2 and Leaderboards' transactional archive contract, change only the integration
test: require queue verified/verdict verified, exact archive terminal3 (ranked) or2 (imported),
and zero unreferenced live rows in ONE SQL snapshot before continuing. Use the existing
three-second readiness bound/25ms observation interval, never an arbitrary delay or enlarged
budget. Remove the two duplicate live-log replays after that phase. The actual worker's
verification remains mandatory; existing served-archive/storage/hash checks, actual Go replay
and corrupt-receipt refusal remain unchanged, as do imported-board exclusion, historical
epoch, real socket/privacy and projection assertions. No replacement of replay with a mere
status assertion, production/schema/policy/kernel changes or assertion-to-wrong-result fix.

Root reuses the independently executed forced old failures above, then runs the exact
previously red selected population as81902: terminal0, publicread0.168s/gameserver3.780s/
deploymentconfig0.056s. Both actual queue/compaction checkpoints execute (terminal3/2,
unreferenced_live0); public and historical archives reverify and corrupt receipts refuse.
cmd/gameserver selects no tests, not credited. TS seam explicitly not requested, no TS/browser
claim. Full affected Linux-amd64/Postgres gameserver/replayverify plus native vet65352
terminates0:83.660/1.629s. Both changed checkpoints/public replay/corruption and all replay
queue/claim/retry/version/atomic-compaction/corrupted-archive populations actually execute.
The opt-in one-minute credential-cleanup timer is explicitly skipped, not credited. Existing
retained-rights observations are not privacy success or changed by this fixture repair.
No full hosted-CI, root history/arithmetic repair, feature acceptance or release claim.
Root-created disposable Postgres service remains available for the ensuing independent
review; no background test workload remains. Review by: Codex root (repair implementer/first
filter). Recorded by: Codex root. Complete12769925-exclusive through the containing repair
commit requires fresh independent corrective review; cursor approval is not this repair's
approval. No unrelated dirty work existed or was changed.

## 2026-10-16 — Archive-phase fixture repair independently approved (RP-677)

Review by: /root/composed_archive_phase_corrective_review (fresh independent Codex under
Marco's corrective-code delegation). Recorded by: Codex root, transcribing terminal verdict.
Decision: APPROVED, no findings; complete four-path single-commit range
127699253f13a99896f0a5a9681e9b1be6e244de-exclusive..aed11047c200fe2fab1692543698b8d7abe36f77-inclusive.
All changed integration-test/CI plan/log/backlog paths inspected against CI D2 and the
existing transactional verification/archive contract. Committed SQL phase is observed in
one snapshot; malformed verified/dead states fail, existing3s/25ms readiness values remain.
Served-byte/hash binding, real Go replay/corrupt receipt, historical epoch, Founder witness,
socket/privacy and imported-board exclusion assertions remain, not a tracker-only substitute.

Independent old diagnostic93622 fails as required: real archive1/live1/non-null seq3 then
log_gap (test0.920s/package0.999s). Diagnostic hashes match; only the real-archive handshake/
scalar logging precede the unchanged old assertion. Full actual declared LinuxAMD64/Postgres
gameserver/replayverify60907 passes89.860/1.668s; both terminal3/2 checkpoints/unreferenced0,
public/historical replays, corrupt receipt and queue failure/rollback/compaction/immutable/
corrupt-content/omitted-field negatives execute. Native affected vet, complete-range diff
and changed Go formatting pass. Opt-in one-minute cleanup skips and unrequested TS seam
are explicitly not credited; retained-rights observations remain negative evidence.

Execution code atb58473af matches the reviewed endpoint; intervening changes are Markdown
only. All handles terminal/slot released; reviewer inspects then removes only disposable
tmpfs Postgres4ea2b0326940 (ephemeral test data gone), preserving both BuildKit builders and
cache volumes. No tracked reviewer edits/commits/main-tree mutation/publication. Docker
Desktop LinuxAMD64 emulation is not native hosted hardware. Only this fixture correction
is approved; no neighboring cursor, full CI/history/arithmetic, feature/archival/release or
1.0 promotion. Root records the bounded debt closed; remaining whole-product gates stand.

## 2026-11-01 — Browser fixture corrections approved; current Linux selection red

APPROVED corrective scopes only. Review by: fresh independent Codex
`/root/browser_fixture_isolation_corrective_review`. Recorded by: Codex root.
All changed source/test/docs/record edges and current descendants inspected:

- RP-525 `d618b0b4..e2efdebf`, six paths: retain composed proof in mandatory Node,
  exclude only that Node-only file from browser collection.
- RP-527 `71682ad0..a181b935`, five paths: expired Typer producer containment.
- RP-528 `a181b935..0073094e`, five paths: actual Garden startup and owned teardown.
- RP-530 `dcb561c5..6a6a996d`, seven paths: Reputation plan input/DOM lifetime.

At clean source `9feb826a`, independent complete three-file native Chromium/WebKit
passes 160 ordinary cases plus eight expected diagnostic deadlines (14.19s).
Node passes all65 proof controls plus one helper;83 browser cases skip. Types pass.
The declared frozen-install Linux three-engine selection also requests generated
API/Node proof paths: generated132×3 and Reputation20×3 pass; no Node-only import.
Garden Chromium/WebKit33 each pass. Firefox reports two failed cases: harvest-all
focus recovery and host harvest receipt. Typer Chromium/Firefox ordinary journeys
hit the unchanged15s deadline; their isolation/drain controls pass. WebKit Typer
never reports completion. Only11/12 files finish; the unchanged10m guard kills
the container, outer exit2/Error137. No final aggregate or assertion expected/
actual values exists. No attribution or retry; RP-526 remains OPEN.

Complete retained output `.cache/browser-fixture-review-linux-final.log`,507lines,
92361bytes, SHA256 `03cba9b8eb61e2c08e417f375da64bff9d6bf93037974de4910b4adbe540040d`.
LinuxARM64 Playwright1.62.0/Node24/Vitest4.1.10, not hosted amd64; other services
were present, so timings are not a controlled benchmark. Initial native sandbox
listener denial ran no tests; permitted execution passed. Owned browser container
is absent. No reviewer tracked mutations, deadline/retry/coverage changes or
intervening topology/RP-531/full-feature/release approval.

### Follow-up — preserve Garden failure context before disposal

Root changes only the two relevant test families and one mounted control: capture
original assertion/error or signal abort plus live focus/grid/busy and bounded
request/intent metadata BEFORE disposal, then preserve the original failure.
Control requires connected pending harvest data, detached evidence after teardown,
abort capture and removed-observer silence; a broken diagnostic sink cannot replace
the assertion. Original gameplay assertions, native inputs and deadlines unchanged.
No natural failure cause or production fix is claimed.

Root final-source complete native file passes66 ordinary/two expected deadlines,
12.27s; declared Linux file passes99/three expected deadlines,13.67s; types0/0.
Earlier pre-abort-observer checks passed the same populations but do not stand in
for this final-source execution. Initial default native selection instead exits2:
Chromium/WebKit controls pass, Firefox session fails at60s and its launch/close
reports sandbox/framebuffer failure at180s. It is not green. Initial sandbox
listener denial ran no tests; process inspection was also denied, not evidence.
Focused working-engine and declared Linux selections have explicit narrower scope.

Code/docs/plan/ledger/record batch after `834cbb20` needs independent corrective
review. Review by: Codex root (implementer/first filter). Recorded by: Codex root.
This is CI D2 diagnostic repair, not ordinary timeout attribution or RP-526/RP-430
closure, public/SQL gameplay, hosted/full-CI, content/kernel/feature/release or
accounting/history approval. No workflow, product code, retry or budget changed.

### Independent approval of Garden failure observation

APPROVED complete five-path `834cbb20..b7045cbd`, no findings. Review by:
independent Codex `/root/browser_fixture_isolation_corrective_review`.
Recorded by: Codex root, transcribing its verdict under Marco's delegation.
All test/docs/ledger/plan/log edges and actual callback/cleanup wiring inspected.

Independent native instrumented families/control pass24 cases,44 selector skips,
8.88s. A temporary ignored real-TestContext probe establishes a ready pending-
harvest fixture in beforeEach, then fires the actual50ms framework deadline;
its successor requires connected/busy evidence before original-context disposal,
the exact timeout reason and subsequently disconnected target. Native passes
two positive/two expected deadlines,68 selector skips,2.92s. Declared LinuxARM64
three-engine selection passes nine positive/three expected deadlines,198 selector
skips,4.93s. These prove observer timing, not the ordinary failure's cause.
Types0/0 and exact range whitespace pass. Initial nested-selector Linux shell
failure exits2/Error127 before tests, no credit; its owned container was already
absent when exact stop was attempted. Corrected bounded selector passes, not a
retry of the ordinary failed journey.

Retained ignored executed probe `.cache/garden-host-review.executed.ts`, SHA256
`25ba7fb5b6d4a2342827a3df19400228a10f3dffffdee135909b454114efe8af`;
Linux terminal-after-setup capture `.cache/garden-observer-review-linux-terminal.log`,
SHA256 `061ba8de07837f0a0dd95628ffa635cc923433eb3ac648843d969946e4ed033b`.
The latter is not a complete raw transcript. Reviewed actual Garden source SHA256
`c356827031c2b6fa1a30c5f0a0b8d22325cc514c84160561239582f8b828ed42`.
Temporary test copy removed from collection; owned browser probes absent; reviewer
made no tracked edits/commits. Root separately rechecks the original larger red
artifact digest and confirms only unrelated containers/builders remain running.

This closes diagnostic corrective review, not RP-526/RP-430, whole hosted/local
CI, accounting/history, feature/archival/release or full1.0. No thresholds, inputs,
assertions or retries changed. The two accounting/search drafts and separate Tier2
activity-window amendment still await their explicit adoption answers; no push.

## 2026-11-01 — ordinary Linux browser job and bounded worker controls

At source `89cee373`, run the actual root target used by Actions: broad functional
three engines excluding only Typer, then fresh three-engine Typer, then isolated
Chromium performance. The earlier five-file focused co-run bypassed that sequence;
its failures remain evidence but were not an ordinary-job verdict. Only this
plan's diagnostic notes change during the controls; runtime/test/config bytes do
not. Each control is predeclared above and runs once, not retried until green.

| Command through declared service | Terminal outcome | Completed evidence / missing phases |
|---|---|---|
| `make test-browser-ci` | exit2/Error137; unchanged guard kills at22:51:22 UTC | Broad reports492 unique file rows/39,201 declared cases in those rows,136 failures/24 skips. Garage423-case file unreported in all three engines. No final summary, separate Typer or performance. |
| `docker compose -f compose.browser-test.yml run --rm browser test-browser BROWSER_TEST_FLAGS=--maxWorkers=1 BROWSER_REPORT_DIR=.cache/browser-worker1-89cee373` | exit137 at23:01:46 UTC | Broad495 files:40,434 ordinary passes/12 expected diagnostic failures/24 skips,545.38s. Separate Typer3 files:87 ordinary passes/6 expected failures,14.12s. Performance starts23:01:29 but has no result before guard. Full target FAILS. |
| Same service/target with `BROWSER_TEST_FLAGS=--maxWorkers=2 BROWSER_REPORT_DIR=.cache/browser-worker2-89cee373` | exit137 at23:17:10 UTC | Broad494 unique rows/40,469 declared cases: three Firefox failures/24 skips. Firefox real-idle file unreported. No final summary/JSON, separate Typer or performance. Full target FAILS. |

Worker2's reported failures: Reputation host `era_2000, Enter, desk-remount`;
Garden codex native input active/error at1280px. Their final expected/actual
assertion detail is unavailable because the runner never finalizes. These are not
deadline-observer failures and are not attributed to product or fixture code.
An initial quick raw-text filter missed the colored failed header; ANSI-stripped
census corrected that report while the run was live. No green claim survives it.
The failure/abort-only Garden-host observer does not cover the Garden-surface file.

Worker1 JSON and stdout cross-check the population: JSON's passed counts INCLUDE
the expected diagnostic failures, which are separated above. Its24 skips are
five performance-only cases and three Node child-process controls in each engine,
not silently missing gameplay. JSON contains495 file results, not its nested
suite-count field; repeated paths do not expose an engine identity. Aggregate
file spans across the three copies identify Garage423 cases at449.65s summed,
real60s idle at180.89s, screen111 cases at114.15s and Reputation host93 at111.34s.
These sums are not elapsed time or a hosted benchmark. All cases remain required.

Full Docker logs were replayed/followed to ignored cache before container removal.
Tool-output chunks themselves truncate; they are NOT complete transcripts.
SHA256 identities (raw bulky artifacts remain ignored):

- `.cache/browser-ci-89cee373-container.log`,1,644,153bytes:
  `6774196ec22a513d15ae41002da8f3649b564f1d37480801f52ae78e4baaedac`.
- `.cache/browser-worker1-89cee373-container.log`,1,673,376bytes:
  `0fa5281b8337beae467cbf20a026f2264e48986e8e8a468d6ae0422c62c08b87`.
- `.cache/browser-worker1-89cee373/functional-broad.json`,12,468,363bytes:
  `12435223f8c3883282c0ab10173422decdced3f321fd128e94a6ec8f58864f6b`.
- `.cache/browser-worker1-89cee373/functional-native.json`,26,221bytes:
  `1896621e7a9527d80de71d0adf3a066c773ff1f8a37692abcadfe7efc9d39d12`.
- `.cache/browser-worker2-89cee373-container.log`,1,804,648bytes:
  `6004876d47f457356a9193782797a70cd4493540c7160f4a3127e8ea29b526a8`.

LinuxARM64 Playwright1.62.0/Node24/Vitest4.1.10;14-CPU Docker VM. Other services
and host load are uncontrolled; optional JSON reporting also differs from default.
This is a bounded scheduling diagnosis, not a matched timing/causality experiment.
Actual installed Vitest uses maxWorkers per browser project; one/two means up to
three/six active files across engines. All owned browser containers are absent;
no unrelated service is stopped or removed.

No software fix or worker cap lands. RP-526/RP-637/RP-430 and strict hosted
sub-five-minute/full-CI acceptance stay OPEN; ten-minute containment is not the
accepted latency floor. Next: explain a concrete native-input failure and account
for measured Garage/real-idle cost without reducing cases, replacing native
inputs, increasing deadlines or guessing another worker count. Accounting/search
and Tier2 amendment adoption remain separately pending, not silently authorized.
Recorded by: Codex root (executed diagnosis, not independent approval). No push.

## 2026-11-02 — RP-699 browser screenshot and attachment isolation

Outcome under CI D2: preserve each engine's own failure evidence without concurrent
overwrites. Source `31ba8dfe` narrowed to Garden surface/Reputation host through
`make test-browser-ci-focused` with `--maxWorkers=2` and separate JSON output.
It ends exit2/Error137: five of six files report pass, WebKit Reputation never
reports and no final JSON exists. Sandbox-denied first invocation executes no tests.
Full Docker log `.cache/browser-input-localization-31ba8dfe-container.log`,167240bytes,
SHA256 `7abb3a6d918c5722263e66117773a9d3e85b59d2793af44f777b389491a63d0e`.

Live read-only process/resource inspection plus a five-second inspector profile
finds JavaScript idle (2107 of2108 samples), one92.1%-CPU libuv worker, and one
pending FSReqPromise. Open descriptors hold the7346-byte initial-read-pending
intentional-deadline PNG and its attachment. Installed Vitest's real source hashes
only the screenshot path: all engines share source and destination; the observed
destination hash `e9172ef3349a0af853d2fb4c85f0bddcfb483054` recomputes exactly.
The native worker's stack is access-denied. Separate ordinary copyFile3.78ms and
buffered-copy12.77ms controls pass, so no general filesystem failure, precise
native-loop mechanism or cause of all historical failures is claimed. Inspector
attachment changes this diagnostic run's observation conditions, not its deadline.
Ignored profile `.cache/browser-input-node-31ba8dfe.cpuprofile`, SHA256
`6bca7b908b4531c6228593028bbb089ad9762aeb680882d70fe8803b4396f436`.

Correction changes only browser-instance screenshot directories, including the
performance instance. Assertions, population, native input, workers, retries and
deadlines remain. New actual PNG/annotation regression fails old config in Linux
Chromium (exit2,177ms assertion/4.56s run), before annotation. Its first annotation
call shape is rejected by types and corrected to the installed public API; that
unreached branch is not the old-config failure. Corrected types0 errors/0 warnings.
Root CI topology and all27 mutation controls, shared Linux entrypoint controls,
Go cold-selector controls and separate browser-report controls pass.

Corrected three-file/all-engine selection with the same two-worker selector and
verbose/JSON reporters finishes53.04s, exit2: eight files pass/one fails;305 ordinary
passes, six expected diagnostic failures, one ordinary failure (312 total).
All three artifact cases pass. Source PNGs differ by engine; actual annotation
copies exist at three distinct path hashes and compare byte-equal to their sources
(Chromium6469, Firefox8176, WebKit8614bytes). JSON `.cache/browser-artifact-isolation-corrected.json`,
91478bytes, SHA256 `c1719578c58adf8b570776af7fa1ae03c1f53c5f437bc530af710a8ffcab7f09`.
The attempted corrected Docker-log follower starts too late and reports missing
container; tool output truncates. Neither is claimed as a full raw transcript;
the completed JSON and returned terminal summary are the retained outcome evidence.

Actual remaining Firefox failure: `R9 host plan matches visible selection:
era_1995, Enter, desk-empty`, line373, fresh consent after rejection expects two
requests but sees one. This is a concrete next debugging point, not resolved by
artifact isolation. Garden surface and WebKit Reputation complete in this run;
earlier failures are retained. LinuxARM64, not hosted/amd64/full CI or reliability
proof; no performance timing claim. Review by: Codex root (implementer first filter).
Recorded by: Codex root. Complete batch after `31ba8dfe` needs independent corrective
review; full1.0, RP-526/RP-637/RP-430 and adoption/release gates remain open. No push.

### Independent corrective approval and traced follow-up

APPROVED artifact isolation only, complete six-path
`31ba8dfe261d9c9d821452d78dba7564d1c63309..22c50316e9f9d4c71be1fec054171ce1fcfb58e9`.
Review by: fresh independent Codex `/root/browser_artifact_corrective_review`.
Recorded by: Codex root, transcribing its verdict under Marco's delegation.
Reviewer inspects every edge and independently runs artifact/Reputation host/
Garden HOST (not surface), all three engines with the same2worker selector:
exit0, nine files,375 ordinary passes/nine expected deadlines,52.03s.
It independently recomputes the three annotation and nine automatic-deadline
screenshot attachment identities; each distinct1280×720 source byte-matches its
copy. Types0/0 and exact-range whitespace pass. Ignored old-config Chromium
control fails the intended namespace assertion,115ms/1.11s, exit1.
Earlier shell/sandbox, duplicate-config and alternate-location filesystem-access
misses ran no intended negative; none is credited. Temporary config removed;
reviewer makes no tracked edits. Positive JSON104386bytes SHA256
`c2e96af6d047cfd644a9043441de20d5e0ca0107fecf887daa2b4d8142b5eebb`;
valid negative JSON1366bytes SHA256
`7b4d5e3113079eaace4d5c02adb1b2cfbc80992c400470299cb7998df8de7dd4`.
Ignored report names are `.cache/browser-artifact-independent-review.json` and
`.cache/browser-artifact-old-independent-review.json`, respectively.
Tool output truncates; no complete transcript is claimed. Direct copy-byte
identity is independently checked, not an assertion inside the new test.

Root then runs ONE built-in `--browser.trace=retain-on-failure` diagnostic of the
complete Garden SURFACE/Reputation host pair, same2worker selector, verbose/JSON.
Exit0: six files,303 ordinary/six expected deadlines,60.90s. No failure recurs,
so no failure trace survives retention and no cause/fix is inferred. No custom
observer, assertion change or further unchanged retry. Complete container log
`.cache/browser-input-traced-22c50316-container.log`,284704bytes, SHA256
`07ec20234c1a86f02a2767b949f037b83dcd4b043c81b3493cc74bf67f7ba5ea`;
JSON84659bytes SHA256 `0f743b05b415786874e0218808bbb60e8e51116d34147aec6d18e0538f3476f8`.
Original Firefox fresh-consent red, native-loop causation, full/hosted CI,
performance/five-minute floor and full1.0 remain open. Next is the ordinary full
Linux target once at the corrected revision, not another worker-count experiment.

### Ordinary full-job result at reviewed correction

At clean `72656dd0`, exact `make test-browser-ci` (no selector/worker/reporter/
trace overrides) is terminal RED: exit2/Error137, unchanged guard at
2026-11-02T00:05:50.569051180Z. Broad begins23:56:00.575042879Z. No final
aggregate, separate Typer or performance invocation occurs. ANSI-stripped partial
reporter census:440 unique project-file rows/37,945 declared cases in those rows,
86 failures/19 skips. This is NOT a full passing-case count. Compared with the
previous complete broad population plus the three added artifact cases,58 of498
project-file rows remain unreported: Chromium1, Firefox55, WebKit2. All three
Garage interaction files, Firefox Game UI screens and Firefox Reputation host
are among them. Original native-input assertions/deadlines still fail; sampled
Reputation aborts retain outstanding native RPC and accumulated command timing.
Missing rows/phases are not skipped or passed.

Complete retained Docker log `.cache/browser-full-72656dd0-container.log`,1241403bytes,
SHA256 `db08bc766fe57d5e7122b6fe9d5e8de0fcf355cff47c89a9415082bb0530698c`.
Tool chunks truncate; full Docker follower ends exit0; owned container absent.
No source changes during this run. No further unchanged full-load retry.

Actual container Nodev24.18.0 reports availableParallelism14/visible CPUs14.
Installed Vitest `createBrowserPool` computes max(1,min(12,CPUs-1)) per project,
then starts all three project pools concurrently: up to36 active files here,
not a global12-worker ceiling. A single Docker stats observation reads878.81%CPU,
10.68GiB/15.6GiB. The RFC's four-CPU runner model instead implies nine; contemporary
hosted CPU visibility is NOT measured here. Other services/load remain uncontrolled.
These facts identify an environment/concurrency mismatch, not a proved universal
cause, safe capacity, new limit or authority to choose another convenient count.
The prior one/two-worker controls and isolated fresh-consent red remain evidence.

RP-699 stays independently approved only for artifact identity. Full CI, resource/
input reliability, strict hosted latency, historical guard/accounting and full1.0
remain open. Next diagnosis must account for actual resource conditions and the
native input/runner boundary, using existing verbose/trace support rather than
another custom framework or unmeasured worker cap. Recorded by: Codex root;
this execution record is not a new designated feature/release approval. No push.

### RP-700 — actual iframe activation failure; no safe upstream upgrade assumed

Source156fa626, four-CPU trace control uses ignored Compose cpuset0-3; actual
Node availableParallelism4 gives the installed default three files per engine.
Full broad → Typer → performance target is terminal137 at00:22:57.735354210UTC,
after broad start00:13:07.379625132. Seven explicit verbose failures (not eight:
default reporter repeats one) include Firefox fresh consent, four Garage engine/
input variants plus two further Firefox Garage cases. No broad aggregate/JSON,
Typer or performance completion. Tracing overhead and other load prohibit timing
comparisons. Full Docker log2032122bytes SHA256
`03197239b241d4131937f2968f45c3026ddccd08a23ae9a9d59c1d10c59ca64f`,
`.cache/browser-fourcpu-156fa626-container.log`; follower exits0, container removed.
Raw traces are recoverably moved to ignored `.cache/browser-fourcpu-156fa626-traces/`.
The Firefox Reputation era2000/Enter/offer-selected ZIP1340239bytes SHA256
`0ad8a1b208ef3a5355981cfc0e428fb9ccdd3605f58fe8f7d15e74b8b6f21be2`.
Immediately before the fresh keyboard command the real iframe DOM is offer_sheet,
aria-busy=false, buttons not disabled. App rejection precedes this command. Native
Enter down/up execute, but the second request never arrives. Trace action markers
are NOT observed DOM events and do not prove handler entry or causation alone.

Independent synthetic probe executes via declared Playwright service, frozen
dependencies, same CPU override,120s process containment. Eighteen cases finish
exit0: all three engines × top-level/iframe × healthy/other-page-pointer/explicit
frame-focus restoration. Initial native button focus is asserted; after another
page's mouse click and200ms upstream settling interval, the installed Vitest
conditional-focus logic precedes real Enter down/up. Firefox153.0 iframe alone
loses the click: trusted keydown/keypress/keyup target the active button, repeat
false, activation0, blur1. The explicit window.focus arm restores trusted click/
activation1 despite hasFocus still false; all other17 cases activate once.
This demonstrates a concrete automation defect without game code; it does NOT
attribute every historical failure or prove refocus is race-safe in parallel.
The committed `firefox-focus-probe.mjs` preserves executed source with only its
relative Playwright import adjusted for the committed directory. Executed ignored
source SHA256 `f0075120b76061dacaf7c85b857e482671d8c24a5cddfafb7a4e47e71086a028`.
Complete captured stdout exports18 observations to
`.cache/browser-iframe-focus-probe-156fa626.json`, SHA256
`45f0b685bbf91ca0d593fec185413a104ad46d49adb6a0e883a00a0906cd479c`.
The late Docker follower missed the removed container; it is NOT a raw transcript.
Probe exit0 means diagnostic completion, NOT an activation/reliability gate pass.

Primary upstream evidence checked2026-11-02:
[42933](https://github.com/microsoft/playwright/issues/42933) describes this exact
shared-browser iframe focus defect. Its [42969](https://github.com/microsoft/playwright/pull/42969)
fix was rolled inr1553, then reverted after
[43280](https://github.com/microsoft/playwright/issues/43280) broke iframe→parent
focus.42933 is reopened and reports the defect still present inr1569; do not
blindly bump the pinned toolchain. The supported per-instance serial-file option
is documented in [Vitest7916](https://github.com/vitest-dev/vitest/issues/7916).

Next bounded candidate is Firefox-only fileParallelism:false, not a global worker
cap, refocus retry, synthetic activation or deadline relaxation. Other engines
retain default scheduling. Execute the complete ordinary Linux target; all phases
and terminal0 required. No guard increase or full-CI/hosted/1.0 promotion from
partial evidence. Recorded by: Codex root. No product/kernel/balance change or push.

### Independent review — RP-700 diagnostic, not the candidate

APPROVED, complete four-path
`156fa626dc8b3f3857a33638f841bebbe5944c71..a2a93d642b4f543cce19e4aa96fd0bd0c68dd542`.
Review by: fresh independent Codex `/root/firefox_focus_diagnostic_review`.
Recorded by: Codex root, transcribing the verdict under Marco's delegation.
Reviewer executes committed probe once through declared Docker service/frozen
deps/four-CPU override/120s containment, terminal exit0,18 cases. It independently
reproduces exactly the Firefox iframe/other-page-pointer trusted-Enter/no-click
miss and explicit-frame-focus restoration; the other17 activate once. It checks
raw artifact sizes/digests, seven verbose failure rows, real enabled/nonbusy DOM
and upstream reopened/rollback evidence. Exact diff/whitespace pass. No tracked
edits, product mutation or interference with root's full-job container. This
approves the diagnostic only; no candidate-config, causal universality, reliability,
whole-CI or release approval. Diagnostic exit0 is not an activation gate pass.

### Firefox-only serial-file candidate — terminal RED, not installed

At a2a93d64 plus exactly four config lines, browserInstance adds only Firefox
fileParallelism:false. Chromium/WebKit defaults, all engines/populations/native
inputs/assertions/deadlines and10m containment remain. No trace/retry/worker/CPU
override. Exact complete target runs via declared browser service with separate
functional JSON directory. Executed config SHA256
`762d6a2539011e43b80fe007907672959e33d1c0311c6432dde8761b0a8f81f2`, preserved
ignored `.cache/browser-firefox-serial-a2a93d64.config.ts`.

Terminal exit2, Make Error1, NOT guard termination. Broad finishes507.85s:
497/498 project-file rows pass;40436 ordinary passes,12 expected diagnostic
failures,24 declared skips and one ordinary failure (40473 total). JSON counts
the12 expected failures among40448 passes; do not double-count. All166 Firefox
rows pass, including423-case Garage, Reputation host/surface and real60s idle.
No inference of reliability or universal cause from one population. One Chromium
failure remains: `R9 native Tab confirmation and cancellation: era_2000,
two-buyable, Enter`, unchanged15000ms deadline, reporter duration21763.2ms.
No assertion mismatch/cause is inferred from timeout. The complete job's Typer
and performance phases do not execute because the broad failure correctly stops
Make. The candidate's predeclared complete-target criterion therefore FAILS.
All earlier red/control results remain evidence; no retry to substitute a green.

Complete Docker log1802422bytes SHA256
`7a5f608b400595563bb081fb15ef2d3619e1d533f89a43285f94911d24afb0c3`,
`.cache/browser-firefox-serial-a2a93d64-container.log`; follower terminal0,
owned container absent. Completed broad JSON12498359bytes SHA256
`cf0b61f820649d9abccf91a8a02ae7f6329a7214214fdbc2b7e1a0b0d26e03e9`,
`.cache/browser-firefox-serial-a2a93d64/functional-broad.json`. Tool chunks truncate;
these retained full artifacts own the counts/failure. Concurrent review probe and
brief host type/topology checks mean no clean-load timing benchmark is claimed.

`make typecheck` exits0, Svelte0 errors/0 warnings. `make verify-ci-topology` exits0,
including27 negative fixtures and shared Linux setup/Go selector/report controls.
They do NOT make the browser job pass. Candidate lines are removed with apply_patch;
active config diff is empty, candidate remains recoverable ignored, not installed.

Next diagnosis is the actual failed Chromium native confirmation journey at
reputation-surface-browser.test.ts:310: retain native command/stage and abort
timing before teardown using existing lifetime/observation helpers and a forced
cancellation control. No assertion/input/deadline weakening, arbitrary worker
count, unchanged full-target retry or assumed toolchain upgrade. RP-700, broader
CI/resource/latency/hosted gates, accounting/adoption and full1.0 stay open.
Recorded by: Codex root; this execution checkpoint is not designated corrective
or feature acceptance. No product/kernel/balance/workflow change, archive or push.

### Chromium confirmation timeout — native lifetime and abort-only context

Outcome: capture the exact native command/stage and completed timings before
the original failing Reputation confirmation fixture is removed (CI D2/RP-637).
The nine original keyboard calls, assertions, test identity and15000ms deadline
are unchanged. Existing lifetimeKeyboard now owns that journey's native promises:
refuse expired continuations, drain a started RPC before unmount, idempotent
body/finished-hook cleanup and a fail-loud pending-input guard. Test/diagnostic
scope only; no game, worker/config, workflow, dependency, bound or copy change.

Abort-only `reputation-confirmation-deadline` records active key/duration, ordered
completed timings, mounted/focus/confirmation state and original abort reason.
Healthy input emits no failure record. A native trusted Enter listener deliberately
aborts during the real awaited RPC (without canceling its default action); control
asserts mounted capture, active/unsettled command, later drain and refusal of Tab.
A real Vitest50ms deadline is established after native fixture setup in beforeEach;
its follow-up verifies original timeout text, capture before removal, cleanup and
late-input refusal. These control manual cancellation AND the framework deadline,
not fake an outstanding RPC with a held test promise. Timings are promise lifetimes,
not observed key delivery or a causal explanation for the original timeout.

Final `make test-browser-ci-focused` complete affected file/all three engines,
verbose/JSON: terminal0,243 ordinary passes plus three expected50ms deadline
failures (246 total),8.18s. `make typecheck` terminal0, Svelte0 errors/0 warnings.
Earlier development populations240 pass before the real-deadline pair; not the
final acceptance population. Initial quoted regex fails at the Make/shell boundary
and Docker access before tests start; corrected ordinary file selector executes.
First typecheck catches an implicit test-callback type, corrected before final
checks. First late Docker follower misses the removed container and is NOT a raw
transcript. Final follower is complete/terminal0: log68032bytes SHA256
`6d92b94b4294834658df7b890e6c70dda53c42dfdb2ef283cd54cc60afaac76c`,
`.cache/reputation-confirmation-deadline-907f9021-container.log`; JSON71008bytes
SHA256 `d79afd531753df3ae6fcda3b5aeace66f6ea76eccd3f748b67dd7b43a2b1953e`,
`.cache/reputation-confirmation-deadline-907f9021.json`. Final test source SHA256
`4a3ae00db1f6d6db9380766580bd5c14c20f6b383881cd5030bac651d2d3d3b9`.

Next: one ordinary full target at this instrumented revision, no trace/concurrency/
CPU overrides, retaining new real abort context and every failure/incomplete phase.
This focused pass does not explain/correct the earlier Chromium timeout or repair
RP-700/whole CI, hosted latency or release acceptance. Complete batch after907f9021
requires fresh independent review; Review by: Codex root (implementer first filter).
Recorded by: Codex root. No archive or push.

### Independent confirmation-diagnostic approval; genuine accumulated RPC delay

APPROVED, complete four-path
`907f902133e54fa1172a2bfeb48ce3fd43ce63e1..5db340da7bea6cf4198ec9a73809e1c7a6b8b361`.
Review by: fresh independent Codex `/root/confirmation_abort_corrective_review`.
Recorded by: Codex root, transcribing under Marco's delegation. Reviewer executes
the entire affected Linux three-engine file:243 ordinary passes/three expected
deadlines,246total,8.23s; types0/0 and exact diff pass. It independently checks
all nine original keys/assertions/identities, actual trusted-input cancellation,
framework50ms capture followed by ordinary cleanup/refusal assertions, and raw
source/log/JSON digests. Initial sandbox denial runs no tests; authorized run does.
No tracked reviewer edits or mutations. Scope is lifetime/diagnostic behavior,
NOT latency cause/correction, whole CI, hosted reliability or feature/release.
Routing correction to the preceding entry: this broader native timeout belongs
to RP-526, not RP-637 (the separate Clout Open/Close defect). Historical entry
remains intact; the mistaken ID did not authorize or change any implementation.

One ordinary full target executes at clean5db340da, default scheduling and all
three engines, separate JSON directory, no trace/CPU/concurrency override. Terminal
137 at01:15:44.666061634UTC unchanged10m guard. ANSI-stripped reporter census has
494/498 broad project-file rows: Chromium165,Firefox164,WebKit165,39215 declared
cases in those rows,96 reported failures/24 skips. This is NOT a completed passing
case count. Missing are three423-case Garage files plus Firefox's real-minute
idle case (1270 cases); no final aggregate/JSON, Typer or performance phase runs.
Full Docker log1472754bytes SHA256
`2d693382426c17eb14d539f5dbd7785aee1848b2814f00d685e6159a034d4812`,
`.cache/browser-native-observed-5db340da-container.log`; follower terminal0,
owned container absent. Tool output truncates; retained full log owns this result.

The new marker captures six genuine ordinary aborts, two Chromium/four Firefox.
Exact earlier Chromium era2000/two-buyable/Enter recurrence:14893.4ms observed
lifetime, eight settled native-command promises totaling14598.6ms (98.02%),
ninth Enter pending35.3ms. Completed commands include2613.6ms Enter,2252.4ms Tab
and4685.4ms Tab. Connected/focused/enabled confirmation remains present. Chromium
Space has14381.3ms in five settled commands plus526ms active Escape; Firefox
cases likewise show accumulated completed+active waits, not a sole15s stuck call.
These are native-promise lifetimes, NOT attribution to exclusive provider, native
event delivery, game handler or a universal cause. Original failure is retained.
Actual Nodev24.18.0 available/visible CPUs14; one Docker sample1308.22%CPU,
9.646GiB/15.6GiB. Other host load is uncontrolled; no matched timing benchmark.

Next bounded candidate: use a shared capacity plan across the three project pools,
derived from the installed runner's max(1,min(12,CPUs-1)) heuristic and measured
per-project multiplication (36 here); reserve Firefox one file for independently
reproduced RP-700. Do not claim the inherited heuristic is a measured safe global
limit. Keep all original populations/native inputs/assertions and bounds; verify
the complete target once. Terminal0/all phases required before a passing candidate
claim; guard/incomplete results refuse adoption. No arbitrary worker sweep,
assumed upgrade or CI/five-minute/1.0 promotion. No archive or push.

### Shared-capacity trial — project-group correction, terminal negative, withdrawn

Candidate after6adb0128 shares the inherited CPU/~12 heuristic across engines,
reserves one Firefox browser file, and retains all original cases/native inputs/
assertions/deadlines. It is NOT installed and establishes no safe operating limit.
Node plan checks initially22 and finally29 pass; types0/0 and existing topology/
27 invalid-fixture/shared-entrypoint controls pass. No balance/game/workflow,
dependency or performance-bound change. No meaningful full-target pass follows.

Initial prototype explicitly gives Firefox root maxWorkers1 and Chrome/WebKit5.
Pinned Vitest groupSpecs places the isolated maxWorkers1 project in an appended
group. Actual log agrees: last WebKit/Chromium Garage rows01:27:36.819193255/
01:27:42.549358091UTC precede first Firefox module observation01:27:48.408568719.
Root stops only owned container847c95362633 after recognizing this mismatch to
the intended concurrent browser pools; terminal143, NOT the ten-minute guard.
No aggregate/phase pass is claimed. Full log1298968bytes SHA256
`6d724ae7cfc24ed1b7521400496f9edca561d4807935464ab33f5707f0d31520`,
`.cache/browser-shared-capacity-6adb0128-container.log`; follower terminal0.
Initial source copies with matching original hashes are preserved ignored under
that prefix: config.ts `6d2ac747fd88709c45d014056b3a148ff8dd3ab0a195bc270d214e95d1d37893`,
worker-budget.ts `0cdd20316e2fb2d771d2ba07e3de69999d69ca8e3d3402cc3793b4e71e131b18`,
worker-budget.test.ts `6481d658de6e2367da7e2a1962c4255650de04cfff4e6584f3dabed2d3fc09ee`.

Correction uses equal root counts for all projects, Firefox browser-only serial
file switch, and avoids explicit root1 when the calculated parallel count is1.
Real startup now shows Firefox alongside five Chrome renderers/five WebKit pages.
Pure controls cover capacity arithmetic/minimum floors/invalid input, common root
counts and the browser-only one-file option. This confirms expression of the
candidate, not a reliable capacity or the claimed latency outcome.

Clarification of root's overly broad intermediate chat inference: mixed-count
deferral is established, but the old all-engine --maxWorkers1 timing is NOT thereby
invalidated as a per-engine serial control. Source3628–3711 collects all browser
specifications within each group, then its browser pool can run projects together.
No earlier outcome or timing is rewritten or reclassified from that inference.

Corrected full target, no trace/CPU/CLI worker override, exits137 at unchanged
guard01:40:07.271591922UTC. Partial reporter rows: Chromium167, WebKit167,
Firefox42 (376/501);30317 declared cases in those rows, nine reported failures,
21 skips. These are NOT complete passing counts. One WebKit reduced-motion/
prediction assertion fails (4428ms); Firefox Garage has five failures/434860ms,
plus three further Garage failures in the other engines. No final assertion
summary/JSON, Typer or isolated performance completion. The candidate fails its
complete-target criterion, even though no ordinary confirmation abort marker
appears (Firefox's confirmation file remains unreported). Do not infer cause,
successful integration or performance from those absences.
Complete log1862969bytes SHA256
`25f2cfe91b2323f31c74fb75840d5964afffb0e4787949029377748af67c8ac7`,
`.cache/browser-shared-pools-6adb0128-container.log`; follower terminal0,
owned container absent. One late sample115.56%CPU/6.897GiB; main JS waits in epoll,
libuv workers idle, no screenshot file descriptors. This sample does not repeat
the earlier RP-699 copy-spin evidence or prove its permanent absence.

Corrected source copies preserved ignored with matching hashes under that prefix:
config.ts `a73e11bb01af8ad131e9240e4d463131731585e6c68b58c950ebc4e7b90e8f5d`,
worker-budget.ts `a3ea5d0569e5921363f28221dfe805a8c6fbee210cabc4ae5d8f9a963e48a2dc`,
worker-budget.test.ts `588de4263b6229aefa5bb61befeddce13e804abfbb18a9b98f0650bb158026fd`.
Candidate helper/test are removed with apply_patch after preservation; active
browser config diff is empty. None of the candidate is promoted or sent for
corrective approval as a passing repair. Independent approval of5db340da's
diagnostic/lifetime batch remains bounded and valid.

Next: retain actual ordinary failure error/stack at test completion, before a
guard can prevent final aggregate/JSON. Built-in GithubActionsReporter is inspected
and also emits errors only in onTestRunEnd, so merely adding it cannot close that
gap. Use the existing test-failure hook/reporting boundary and verify an actual
ordinary failure versus expected-failure/healthy controls before any new full
run. No new worker sweep or inferred reduced-motion cause; all bounds, RP-526,
RP-700, hosted CI/five-minute and full1.0 obligations remain open. No push/archive.

### 2026-11-02 — preserve completed failures before aggregate/guard

Root Make focused, both functional phases and isolated performance commands now
add a small Vitest onTestCaseResult reporter alongside default/requested JSON
reporters. Only final failed verdicts emit browser-test-failure with engine/file/
full name and error name/message/stack. Correct expected failures remain quiet;
unexpected .fails passes are failures. No exit, bound, scheduling, input, assertion,
engine or dependency changes. Collection/global errors and unfinished tests are
not covered by this hook. Existing topology fixtures bind all Make lanes; a
non-collected probe template retains the explicit real control without adding a
permanently failing or held test to ordinary discovery.

Real Linux three-engine probe runs an ordinary failure, expected failure, healthy
case, unexpectedly passing .fails case, then a deliberately held test. Current
Make emits six failed-case records with messages/stacks before all three holds;
no aggregate yet. Exact prior fc7300e6 Makefile, same source/population, emits zero
failure records before its three holds and no aggregate. Root explicitly stops
each owned container after observing that boundary: positive wrapper exit2 /
container143; old direct container143. Neither is a passing suite or guard expiry.
Temporary collected probe is removed. Original executed source and prior Makefile
are retained ignored under .cache/browser-failure-reporting-probe-fc7300e6.test.ts
and .cache/browser-early-errors-old.Makefile. Tracked template's explanatory
comment is clarified afterward; its executed test bodies are unchanged.

Completed affected command: make test-browser-ci-focused with
BROWSER_TEST_FLAGS='test/browser-failure-reporter.test.ts test/reputation-surface-browser.test.ts --reporter=json --outputFile.json=../.cache/browser-early-errors-affected-fc7300e6.json'.
Terminal0, six project-files,258 ordinary passes plus three expected failures
(261 total),8.50s, zero spurious failure markers. Five Node reporter cases also
pass; make typecheck terminal0/Svelte0 errors/0 warnings. Existing topology fixture
command terminal0, including27 rejected mutations/shared Linux entrypoint/cold
Go selectors and added reporter presence across Make lanes.

Retained .cache/browser-early-errors-{positive,negative,affected}-fc7300e6-container.log:
positive22044bytes SHA256 c54eb7b942a88c2e8298dc4aaa5ae1daffd04940419d23e6375170d0dca22765;
negative7860bytes SHA256 82d5a418a4f1b62c81960c343aa2ccfcf1231a9bcfc36508faf7315dd43620f8;
affected12172bytes SHA256 4de22f03c98cd7735135dca6428a4a20aa3fc2beb3c88f8a42335676e34b6c36.
Affected JSON74665bytes SHA256 ca4452fa40cd90bcda60934d45dfa933a02952615f968b3964881ce39933f6f9.
JSON's261 passed includes its three expected failures; do not relabel all261
ordinary passes. Final source comment-only adjustment needs no new browser run.

Review by: Codex root, implementer first filter. Recorded by: Codex root.
Complete batch after fc7300e6 requires fresh independent corrective review.
Next: one ordinary full target with this new signal, retaining every failed/
unexecuted phase and actual error details. Full CI/hosted latency/RP-526/RP-700/
1.0 remain open. No worker sweep, reduced-motion inference, archive or push.

### Independent completed-failure reporter approval

APPROVED, complete eight-path
`fc7300e684daff7346169cb768bc48df7b9aea88..fa5faff59dc02cbd9135fed92d56f0fc94fc5159`.
Review by: fresh independent Codex /root/browser_failure_reporting_review.
Recorded by: Codex root, transcribing under Marco's delegation. No findings.
Reviewer independently executes Node5/193ms, real Linux reporter selection across
three engines15/2.89s, typecheck0/0, topology including27 rejected mutations and
Make lane wiring, and exact-range diff check. Initial Docker sandbox denial runs
no tests; authorized declared target passes. Its browser container is terminal
before root begins the next full job, avoiding overlapping reviewer test load.

Reviewer checks pinned .fails inversion occurs before test-finished dispatch to
onTestCaseResult and failed results expose an error array. It independently
checks held-control log digests/six versus zero early records, no aggregates,
byte-identical old Makefile, matching probe bodies excluding comments, and the
affected JSON's261 verdicts including three expected failures. No files modified
by reviewer. Approval covers observation tooling, NOT scheduling/native timeout
cause or repair, entire CI, hosted latency, RP-526/RP-700 or release acceptance.

### Ordinary full job with early errors — complete broad red, not a guard kill

At fa5faff5, docker compose -f compose.browser-test.yml run --rm browser
test-browser BROWSER_REPORT_DIR=.cache/browser-completed-failures-fa5faff5 executes
the ordinary frozen Linux target with no trace, engine, CPU or worker override.
Root terminal2/Make Error1; Docker follower terminal0; owned c42317ff8d08 absent.
Initial raw follower lacks socket permission, runs no capture; immediately
restarted authorized follower retains the complete container log from startup.
No test rerun or overwritten failure. Broad finishes364.81s:12 failed/489 passed
project-files (501),40431 ordinary passes,15 expected failures,30 failed and24
skipped cases (40500). JSON's40446 passed includes15 expected failures. Native
Typer and isolated performance phases do not execute after broad failure.

All30 early records match the completed report: Chromium7,Firefox20,WebKit3;
24 native-input deadlines and six assertions (Firefox Codex activation/two Garden
focus checks; Chromium two Fiscal navigation states; WebKit Fiscal focus).
The Firefox Codex record retains trusted nonrepeat Enter events on its connected
focused SUMMARY, document_focused:false and no open state after settling, matching
the independently reproduced RP-700 boundary. This does not attribute the other
29 failures or prove their independence from earlier timeouts. Existing lifetime
markers again show accumulated native RPC waits. The prior candidate's WebKit
reduced-motion failure does NOT recur; no cause or correction is inferred.
Module HTTP summary2028 started/finished,zero closed early/pending. One late
resource sample388.73%CPU/11.53GiB, uncontrolled other host services; not a benchmark.

Raw log .cache/browser-completed-failures-fa5faff5-container.log2068884bytes SHA256
aff41ac79f456a4618013db753ac2537f4dbad519f37849db525830c7d0bbb15;
functional-broad.json12534910bytes SHA256
49980412d1bfbfdd86cac8f619834c8ec2a4a2aee0aa7856540540d1022f1b53.
No guard signal occurs. Tool display truncates; raw artifacts own the result.
An initial read-only metadata one-liner has a syntax error; corrected reads above
succeed without changing artifacts. Full job is RED, also above the hosted
five-minute floor already in its broad phase; no deadline or floor is raised.

Next correction targets the pinned provider's focusIframe predicate, which skips
window.focus when an active button remains named despite the document losing
native focus. Reuse the real two-page negative/control population, but invoke the
installed provider keyboard command rather than only copying its predicate.
Any candidate must preserve genuine native input, target selection, all assertions
and bounds; no synthetic click, destination focus or dependency-upgrade assumption.
The reporter's independent approval stands; no broad-CI/1.0/archival/push claim.

### 2026-11-02 — real-provider focus regression and version-pinned correction

Existing18-case focus probe now registers the installed PlaywrightBrowserProvider's
actual __vitest_keyboard command through a minimal project binding and supplies
real pages/frames. No copied focus/key implementation. Old cold Linux command
fails terminal1: Firefox153 interrupted iframe retains focused subject/trusted
nonrepeat Enter down/press/up but zero clicks; other17 controls activate once.
The assertion now requires every case's one native activation, not only printing
the miss. Prior diagnostic18-case observation remains historical evidence.

One-line version-specific @vitest/browser-playwright4.1.10 pnpm patch adds
!document.hasFocus() to focusIframe's restoration predicate. Package versions,
target selection, real keyboard delivery, assertions, scheduling, retries and
deadlines unchanged. Patch uses pnpm's declared patchedDependencies mechanism;
pinned11.15.1 generates the lockfile in an isolated Linux dependency volume.
Only patch hash/consumer references change, no resolved package version. Generic
peer warning from lockfile generation is disclosed; frozen installation succeeds.
Host has pnpm12.16.1 and no corepack, so neither substitutes for pinned generation.

Real old/new commands use the declared image/anonymous dependency volumes and
shared test-browser-ci.sh frozen setup, then execute the diagnostic in the same
container under the unchanged10m guard. Explicit manual entrypoint appends the
probe after a small reporter-file setup check; it is NOT the full CI target.
Exact command before the move: docker compose -f compose.browser-test.yml run
--rm --entrypoint timeout browser --verbose --signal=KILL 10m bash -lc
'bash client/tools/test-browser-ci.sh test-browser-focused BROWSER_TEST_FLAGS=test/browser-failure-reporter.test.ts && node planning/scaffolding-and-ci/firefox-focus-probe.mjs'.
Old terminal1/new terminal0; both raw followers terminal0. New frozen install
passes18/18 native controls; the setup selection's15 tests also pass,8.11s.
Old log22620bytes SHA256 dea1afee894ad4df107f4b7ba5cef0a8163ffe9c712fda118a8525134dd42471,
.cache/provider-focus-old-d7518dea-container.log; new21728bytes SHA256
9b74e4ea24a757c13cedfcabaec87708e0c260bb8dc683f478b1d8f0fb0fb90b,
.cache/provider-focus-new-d7518dea-container.log. No failure erased/retried.

After execution, move the same regression into client/tools/verify-browser-provider-focus.mjs
(relative imports/comment adjusted). Root Make test-browser-provider-focus is a
full-functional prerequisite, sequential before broad/native/performance; focused
selection stays focused. Existing topology fixtures verify this ordering and
preserve all prior controls; terminal0 including27 negative fixtures. Typecheck
terminal0/Svelte0 errors0 warnings. No product or authored-copy bytes change.

One ordinary full target is now executing against this candidate, with separate
.cache/provider-focus-full-d7518dea reports and no CPU/worker/trace override.
Its result is NOT yet known. Specific native-regression correction is locally
verified; whole CI/RP-526/hosted-five-minute/1.0 remain open. Complete batch after
d7518dea requires fresh independent corrective review. Review by: Codex root,
implementer first filter. Recorded by: Codex root. No archive/push.

### Independent provider correction approval; ordinary full broad has two deadlines

APPROVED complete nine-path
`d7518dea40ae069c87b07854f32262adbcfe1aad..ae367f8bfd5fe11d79399a841f384abb82e4d0b6`.
Review by: fresh independent Codex /root/provider_frame_focus_corrective_review.
Recorded by: Codex root, transcribing under Marco's delegation. Initial reviewer
turn fails at model capacity before execution; same authorized reviewer retries.
It independently checks scope, patch/lock hash, actual installed command binding,
real focus/activation, Make ordering/failure propagation and all retained digests.
Independent declared cold Linux run terminal0: setup15/15 tests9.74s, then actual
Make provider regression18/18. Firefox interrupted iframe retains active subject,
loses document focus, then gets trusted Enter and exactly one trusted click.
make typecheck verify-ci-topology terminal0, Svelte0/0,27 rejected controls and
ordering/setup/selector propagation. No blocking findings. Range diff-check flags
only space-before-tab in the standard unified dependency patch's context lines,
not implementation whitespace; reviewer explicitly discloses rather than calling
that check clean. No reviewer edits. Approval is specific, not whole CI/RP-526.

Root ordinary full target begins before commit at the same behavior source later
committed in ae367f8b; only documentation/staging changes while it runs. Exact
command: docker compose -f compose.browser-test.yml run --rm browser test-browser
BROWSER_REPORT_DIR=.cache/provider-focus-full-d7518dea. Default scheduling/image/
install/engines/limits, no CPU/worker/trace override. Root terminal2, follower0,
owned ec7ba04bf054 absent. New normal Make native-provider prerequisite passes18/18.
Broad completes250.46s,499 passed/two failed project-files (501),40459 ordinary
passes,15 expected failures,24 skips,two failed (40500). JSON's40474 passed includes
the15 expected failures. Two15s native sequential-plan deadlines: Chromium
era2000 Enter and Firefox era1995 Enter. All other ordinary cases pass THIS run;
not proof every earlier issue is fixed. Typer/performance never run after red.
Module HTTP2041 started/finished,zero closed early/pending; no guard expiry.
No matched/controlled host-load timing comparison, so30→2 is not a universal
causal repair or reliability estimate. The strict hosted five-minute floor remains.

Raw .cache/provider-focus-full-d7518dea-container.log1835865bytes SHA256
50156c41c24886a7801049ba5ed1fe201d0c31611334425559ad0d7118650a22;
functional-broad.json12498761bytes SHA256
342ea15ef747ce5aa98d24106ff4ef9f71af9230b10f5e3eb3dbf80ac7d5b20b.
Reviewer independently checks these digests; tool display truncation is not missing
raw evidence. Initial Git staging sandbox denial changes nothing; ordinary scoped
authorized add/commit succeeds without push. Tree clean at reviewer completion.

Next bounded scheduling candidate: separate the unchanged whole Reputation-plan
file into its own fresh three-engine phase, as for existing Typer isolation. Its
original population/key/assertion/deadline and total report union must remain.
No worker-count sweep or retry. Complete ordinary target terminal0/all phases
required before a passing candidate claim; an isolated pass is insufficient.
No whole-CI/RP-526/hosted-five-minute/release promotion, archive or push.

### Fresh Reputation-plan phase experiment — failed and withdrawn

At6498c70e, the candidate excludes the unchanged whole plan file from broad and
adds it as its own fresh three-engine invocation before Typer, with separate JSON.
No original case/key/assertion/deadline/engine or worker change. Existing topology
consumer initially rejects the new sequence; explicit three-phase validation and
two missing/duplicated-population controls then pass29 negatives and all existing
setup/selector/report checks. That structural pass is not functional evidence.

One ordinary declared full job runs with this candidate and report directory
.cache/reputation-plan-phase-6498c70e. Provider regression18/18; broad completes
270.35s,496 passed/two failed project-files (498),40401 ordinary passes/12 expected/
24skipped/three failed (40440). Firefox Garden seed-Tab and Reputation host
offer-selected/era2000/Space time out; following Garden mature-menu Enter asserts
its action is undefined. That sequence suggests a lifetime check, not proof of
causation. Reputation-plan, Typer and performance phases never execute: no pass
or union-coverage claim. Goal continuation resumes after an external gap; original
wrapper/follower handles are then missing and owned1d464ae67dae absent. Complete
raw log ends at Make Error1 with JSON/aggregate, not guard expiry; wrapper numeric
exit is not retained. Do NOT restart because handles expired.
Log1855463bytes SHA256 c54e6d17463020d34563d9b2452472cf2fda479a5caeca9a4ad40d7ba430d3aa;
functional-broad.json12481342bytes SHA256
91fa4ad55575f22bd6e9587cc2f24f2d3a93c79daff8ab94cf76cbc9eb341a83.

Candidate fails its complete-target criterion and is NOT installed. Exact original
bytes verified against ignored .cache/reputation-plan-phase-6498c70e-* copies:
Makefile08611fe5ca9896e80106d33306dcf82a2500467ce7664ed001f21b99d71b948d;
ci-topology.mjs e7a340648ab9aef12492c73e33aea0aeb214b1a20636e46dc3ecdd8c8fcb7a5d;
verify-ci-topology-fixtures.mjs2a17c317e22ee5779c9fec33c0b628b47ea1016102637add2779ce8940d8ff32.
All three active files restored via apply_patch to6498c70e. Initial inverse-diff
conversion rejects its reversed path header before editing; corrected conversion
applies and git status confirms zero residual scheduling/validator diff. Keep
the failed evidence; no further arbitrary phase/worker sweep or acceptance claim.

### RP-701 — expired Garden fixture cleanup and native-input containment

Regression first: beforeEach mounts a real Garden and opens its menu through native
Enter, then a50ms framework deadline fires in the test body. The ordinary successor
requires the old fixture disconnected and preserves original timeout reason.
Old fixture fails that assertion in all three engines: terminal2,three failed/
three expected deadline cases,4.42s;123 other selected-out cases are not passes.
Successor finally cleans the old subject to avoid contaminating later tests.
Old raw .cache/garden-lifetime-old-6498c70e-container.log41398bytes SHA256
7062c50e9357477b45ceb22f123f98ac77d63562fa1697ad44277b111e346164;
executed source copy .cache/garden-lifetime-old-6498c70e.test.ts SHA256
81cdb90d1e3e329c893a9a9e4e65012c27df58882e60aef619dff4cbd7f228fc.

Correction binds every mounted component fixture to its explicit TestContext.
Existing lifetimeKeyboard is the sole real native producer, checks cancellation
around import/delivery, drains before unmount/removal, and refuses subsequent input.
Idempotent disposal is shared by body-finally and onTestFinished; beforeEach checks
for undrained native input. NativeWalk retains the exact batched key string and
every intermediate focus/tabstop assertion; all41 original populations/key inputs/
behavior assertions remain. Ordered-read/unmount cases use the same disposal.
No product, copy, deadline, retry, scheduling or provider changes in this batch.

Two additional controls use the actual trusted Enter RPC: healthy input opens the
menu normally; cancellation calls dispose from keydown without preventing default.
Keyup must still see the connected fixture, then disposal completes; another space
must not type into a focused successor. Closed healthy fixtures also refuse input;
cached disposal is idempotent. These controls do not hold a fake RPC or claim to
reproduce the initial15s timeout's cause.

Complete affected command: make test-browser-ci-focused
BROWSER_TEST_FLAGS='test/garden-surface-witnesses-browser.test.ts --reporter=json --outputFile.json=../.cache/garden-lifetime-new-6498c70e.json'.
Terminal0,132 ordinary passes/three expected deadlines (135total),6.73s,all three
engines; follower terminal0. make typecheck terminal0,Svelte0 errors0 warnings.
Corrected log34217bytes SHA256
4c553b7b2faefe79b1a61d5b5cee10ba3c62e1070d8541626f0935605c0662f6;
JSON35922bytes SHA256 0bf00149274576166f6eb9cbc2a68859148a113bfb0579c3cb5974206a2b2efc.
Source SHA256341ec1a8dabed509838a3468c878a022310723f7563242c2c2263b1b7b4ede27.
JSON135 passed includes three expected deadlines. Raw files own truncated tool
display. Review by: Codex root, implementer first filter. Recorded by: Codex root.
Fresh independent corrective review required for complete batch after6498c70e;
then one ordinary full target may use the new containment signal. No claim that
this fixes original latency, RP-526, whole CI or1.0. No archive/push.

### Independent Garden target/input approval and adjacent-node follow-up

APPROVED complete five-path
`6498c70ea90055c2045bb9cf5507aad384f46a9b..bc6d3fcbaa421cb8bfba34d4c4db0684b9986f23`.
Review by: fresh independent Codex /root/garden_fixture_lifetime_corrective_review.
Recorded by: Codex root, transcribing under Marco's delegation. No findings.
Reviewer independently executes complete corrected Linux file132 ordinary/three
expected deadlines,6.27s, types0/0 and diff-check clean. It temporarily materializes
the exact retained old source in client/test and reproduces three successor failures/
three expected deadlines,2.60s,123 deselected; verifies source outside added regression
equals baseline, removes its temporary file with apply_patch and confirms clean tree.
All reviewer browser workloads terminal. Resource metadata missed the removed
containers and is not fabricated or obtained by restarting. Specific containment
approval only; no original-timeout cause/whole-CI/latency/release approval.

Root then identifies an uncovered neighboring-node lifetime: two original newer-
focus controls and the Refresh sentinel still rely on body finally. Extend the same
ready-fixture framework-deadline control with an adjacent external button. At the
approved target/input code, the Garden is gone but the adjacent control survives;
ordinary successor fails all three engines, terminal2,three expected/three failures,
129 deselected,2.81s. Old source .cache/garden-adjacent-old-bc6d3fcb.test.ts SHA256
39ee11567ef15b17375763ae75bb47521d4161f84540f13c2d6e0688203d2a27;
old log42372bytes SHA256 ab0c574936c8247ea67860643fa6fbfd594161751a0f927d6d2e6dae697f1b3b.

Small fixture-local ownership list removes those registered neighbors only after
input drain/unmount. Registration checks cancellation/closure and happens before
insertion. The native-control successor input deliberately remains a separate live
subject, with its own test-completion cleanup; removing it with the expired host
would make the late-input check vacuous. No original key/assertion/deadline changes.
Complete affected file132 ordinary/three expected deadlines,135total,6.43s,
terminal0; follower0. make typecheck terminal0/Svelte0/0. Corrected log33780bytes
SHA2565460f3ea3def33b758250c71c218433e865fd55fc61a21a7eee6f0389cd31280;
JSON35962bytes SHA2563bb7e9456ab842793c7d64454f8472cd55dd69525abf0b026434a04536140078.
Paths use .cache/garden-adjacent-{old,new}-bc6d3fcb-container.log and the corresponding
new-bc6d3fcb.json. Initial docs patch context mismatch edits nothing; corrected edit
and diff-check pass. Earlier approval does NOT cover this follow-up afterbc6d3fcb:
independent review is pending. No full run restarted across goal-resume gaps.
Next finish that review, then one ordinary full target with actual containment
changes. RP-526, hosted latency and proper full1.0 remain open. No archive/push.

### Adjacent-node independent approval; ordinary full run remains red

APPROVED complete five-path follow-up
`bc6d3fcbaa421cb8bfba34d4c4db0684b9986f23..b91790efac2ad496d2c4008211c47e18758c4304`.
Review by: independent Codex /root/garden_fixture_lifetime_corrective_review.
Recorded by: Codex root under Marco's delegation. No findings. Reviewer executes
full affected Linux file132 ordinary/three expected deadlines,8.94s, types0/0 and
clean diff-check. Independent exact old-neighbor source fails all3 successors on
remaining adjacent controls,three expected/129 deselected,3.06s/Make2; verifies
source outside regression equals baseline. Temporary test removed via apply_patch,
tree clean; all reviewer browser work terminal before root full job begins.
This extends, not retroactively enlarges, the earlier6498c70e..bc6d3fcb verdict.

One ordinary full job at cleanb91790ef: docker compose -f compose.browser-test.yml
run --rm browser test-browser BROWSER_REPORT_DIR=.cache/garden-containment-full-b91790ef.
No trace/CPU/worker/engine override. Provider regression18/18; broad completes
470.84s,488 passed/13 failed project-files (501),40446 ordinary passes/18 expected/
24 skips/24 failed (40512). JSON40464 passed includes18 expected. Root terminal2,
follower0,ownedbeecbb7fb2ed absent; no guard signal. Typer/performance never execute.
All new Garden containment controls pass; four original Firefox Garden native
cases time out. The other failures remain retained, not excused by scoped approval.
Early records: Firefox18,Chromium3,WebKit3;22 test deadlines,one Fiscal navigation
assertion and one WebKit Worker readiness failure. Not a reliability/speed comparison.

Full mapped error identifies game-ui-screens-browser.test.ts:442, polling for first
predicted_snapshot before the cadence window. Thus the earlier reduced-motion
failure is NOT evidence its numeric500ms spacing or pulse assertions failed.
The actual poll bound remains unchanged; no cause inferred. Route under RP-218's
existing native-prediction/startup investigation, not a newly invented product defect.
Existing observeNativePrediction already has a detailed report; next reuse it at
this readiness failure before changing any timeout or behavior. No unchanged full
retry, worker-count grid or phase extraction from this run.

Owned-container read-only metadata: Node24.18.0,availableParallelism14,
cpu.max='max 100000',cpuset.cpus.effective='0-13',memory.max='max'. This rules out a
hidden4CPU cgroup quota, not VM/host load or resource pressure. Read-only fetch at
normalized stack port63315 is refused; actual logged Vite port51204 returns source,
but its SSR transform offsets do not directly map the browser stack. No fetched
code executes and no test/source changes; final mapped report owns the line442
finding. Module HTTP2031 started/finished,zero closed early/pending.

Raw .cache/garden-containment-full-b91790ef-container.log2069742bytes SHA256
963afd328e3db0ba3e9f18f18a06c754b65abdf73b0a331c5fa4ebe2a4b92333;
functional-broad.json12522077bytes SHA256
65ef55291a59cf11da65ef0d615c7f313db6b5340713a0139e57d76f0fd4d49b.
Tool display truncation does not replace these complete artifacts. Scoped repairs
and their review gates are closed; original latency/whole CI/hosted-five-minute/
proper full1.0 remain open. No feature archive, publication or push.

### First-prediction readiness — reuse native observer at the actual failed poll

At57cc8217 the reduced-motion witness's mapped failed predicate is first native
predicted_snapshot readiness. A small helper preserves that exact predicate and
Vitest poll defaults, but on failure prints the existing observer report before
fixture cleanup and rethrows the same error. Context includes real lifecycle/
commands/outputs/gaps/visibility, explicit null absent first-prediction time,
requested/rendered preference, counter connectivity and actual wait duration.
No production, native input, media/cadence assertion, timer, bound, provider or
concurrency change. This is poll-error observation, not all test-abort coverage.

Negative control creates a real native Worker, posts its valid dispose command
and immediately terminates before the first event turn. The unchanged poll fails;
ordinary assertions check original error/stack, zero predictions, command/terminate
history, timing and context. No fake Worker/output/clock or suppressed listener.
Cleanup is idempotent, owned by TestContext and body-finally, including construction
failure; healthy readiness remains quiet. All prior Game UI populations remain.

Complete affected Linux three-engine file at the initial diagnostic source:
terminal0,321 passed/15 existing exclusions (336total),27.70s. Afterwards final
failure-payload/cleanup refinements execute the same real control in all engines:
terminal0,three passes/333 deselected,8.25s; make typecheck terminal0/Svelte0/0.
Deselected cases are not passes. Both Docker followers terminal0. Final control
does not replace the broader earlier pass; independent review must execute the
finished affected file. No claim the original readiness failure is reproduced or
fixed by an isolated healthy run. Reporter markers remain absent from healthy logs.

Ignored .cache/worker-readiness-report-57cc8217-container.log263537bytes SHA256
063605678b9eb4fa812d450fec533912b82f9bb4550a9f8d353145423fb234fe;
JSON93808bytes SHA256 d8e8d432c69570b4f693893f457e50abd2e953de0a8265aa4ed5c4c03cad9501.
Final .cache/worker-readiness-control-final-57cc8217-container.log11047bytes SHA256
41f710a5810263860f94dbe9c2c52ed94fdc0b8651ed43cf6f915ffd8bf0147d;
JSON88248bytes SHA2569836f85a3fb5fa2f2dbefd5870b98eac8f642fec6872dcd09c073c2c0f103c1a.

Owning plan now separates the actual current queue from preserved historical
checkpoints; old “next” instructions no longer masquerade as current authorization.
Complete batch after57cc8217 requires independent corrective review. Review by:
Codex root, implementer first filter. Recorded by: Codex root. After review, one
ordinary full run with this new signal is justified, not an unchanged retry.
RP-218/RP-526, full CI/hosted latency and proper1.0 remain open. No archive/push.

### Independent approval — first-prediction readiness context

APPROVED, no findings, complete five-path range
57cc82171d1c653b826cad495a8f86c78d72bcaf..f862b80716be16a8fadbe8bdb848bfd2459c12be.
Review by: fresh independent Codex /root/worker_readiness_context_corrective_review.
Recorded by: Codex root under Marco's delegation. Reviewer independently executes
the finished complete affected Linux file: terminal0,321 passes/15 existing
performance exclusions (336total),28.76s, all three engines. Types terminal0,
Svelte0/0; exact-range diff-check clean. Actual terminated-Worker controls pass;
predicate, poll defaults, cadence assertions and15s deadline remain unchanged.
Historical plan text is preserved; log append-only; retained artifact identities
match. Reviewer JSON .cache/worker-readiness-review-f862b807.json SHA256
9bc3fd8b80d5b87050de8afc0f0a2e417ba529a8cd604cea5f4fa4f58a773eb3.
Initial sandbox denial executes no tests; declared authorized run passes. All
reviewer browser workloads terminal before root full run. Approval is diagnostic
only: original startup/native deadlines, whole CI, hosted latency and release
remain open. No specification adoption, feature archival or push.

### Complete ordinary job after readiness observation — still red

At f862b807, the unchanged declared full Linux target completes its broad phase in
470.10s: 23 failed/478 passed project-files (501),40410 ordinary passes,18 expected
failures,24 skips,63 failures (40515). JSON's40428 passes include the18 expected.
Provider regression18/18. Root terminal2/follower0; owned29f946334058 absent, no
guard termination. Make correctly leaves Typer/performance unexecuted.

Early final-verdict records agree with the aggregate: Chromium15,Firefox38,
WebKit10 failures. Sixty-one are15s test deadlines. The other two are Chromium/
Firefox's first-prediction poll, not numeric cadence. Existing real observer logs
one initialize, visible document, connected cash output, correct rendered preference,
no prediction/error/gap, and1008.2/1066ms readiness waits respectively. It does not
show whether the terminated-on-cleanup Worker would have started later. Do not
infer a product crash or lengthen the poll from this evidence.

Read-only resource samples: actual availableParallelism14; at10:36:48 UTC load57.48,
cgroup CPU some avg10=53.22/full10.53, free memory623271936bytes; memory OOM/kill
events0 and no CPU quota throttling. Existing unrelated database containers consume
about3.2 CPU cores in a separate sample; root neither stops nor modifies them.
At10:39:35 load31.86, CPU some avg10=20.50, OOM events still0. These are uncontrolled
load observations, not a matched benchmark or attribution of every failure.
All2036 observed module HTTP requests finish; zero early closes/pending.

Raw .cache/worker-readiness-full-f862b807-container.log2201612bytes SHA256
e919f5e734ade6e9fc11fc41d6eead1558299537725810dbb27a3b7c3cea977e;
functional-broad.json12559030bytes SHA256
273eff377fe59775c2c8fed470f618862160c8aa2448a030b3c71f43ad07f884.
Tool-output truncation does not replace those complete artifacts. Original failure
causes, full CI/hosted latency and full1.0 remain open.

Next control reuses the existing ignored four-CPU affinity override at this source,
now AFTER the independently approved provider repair. The prior four-CPU traced
control predates that correction. No trace this time; unchanged complete target,
default scheduling, all engines/populations/limits and distinct reports. This tests
the RFC's four-CPU model, not actual contemporary hosted hardware. All phases and
terminal0 are required; no override adoption or timing comparison from partial data.
Actual initial metadata verifies availableParallelism4, affinity0-3, no CPU quota.

### Four-CPU model after provider correction — complete local pass, latency open

The separate declared full target with only .cache/browser-fourcpu-control.yml
finishes terminal0, follower0; no guard. Product/test source remains f862b807;
only the owning records change during execution. Provider18/18; broad501/501,
40473 ordinary passes/18 expected/24 skips,402.31s; native3/3,87 ordinary/6 expected,
6.89s; isolated Chromium performance5/5,8.87s (107 unrelated cases deselected).
Both functional JSON populations match the ordinary run's declared totals.
No assertion/input/deadline change or missing phase. This is an overridden LOCAL
full pass, not default14CPU reliability, hosted proof or five-minute acceptance.

The control's actual visible CPUs4/affinity0-3 are verified before tests. A later
sample observes6.191GiB owned-container memory, versus9–10GiB in the ordinary run;
resource/host conditions are uncontrolled, not a matched speed comparison. CPU
waiting remains measurable; no universal CPU-only cause is inferred. Garage's
unchanged423-case file alone takes338.43/382.04/348.03s across engines. That serial
file exceeds the entire workflow budget, even with every case green. Next repair
must preserve its populations while distributing coherent test families; no bound
increase, deletion, engine omission or permanent affinity adoption follows here.

Raw .cache/provider-fixed-fourcpu-f862b807-container.log1822860bytes SHA256
2832ab90be12c62adecfd853f1a9213429480f5ffa84b9aaf4bd231d90d8f80e;
functional-broad.json12493417bytes SHA256
034c69d6cbd0332644fe5653fc7d42ef0b381bd7c99f5ba1b6efe29794ac962a;
functional-native.json26243bytes SHA256
722e388e254993da50e64f0954396489098ca7c87cf2ce0b3ecea5e64c9e942d.
Full CI/hosted latency, original default-load failures and proper1.0 remain open.

### Garage-family candidate withdrawn; native interleaving reproduced and corrected

The semantic split preserves all98 original non-import top-level statements byte-
identically apart from helper exports. Types0/0. Its actual full four-CPU run,
however, is RED:519 project-files,517 pass/two fail;40471 ordinary passes,18 expected,
24 skips,two failures,433.83s. Root terminal2/follower0; no guard; Typer/performance
unexecuted. Actual old/new JSON case multisets match all40515 declarations after
normalizing moved paths, including all1269 Garage cases/21 family rows. One Firefox
Trophy Case Space re-entry fails at the unchanged surface assertion (line146 in
the candidate); one unchanged WebKit Reputation input-shape case hits15s. No speed
comparison is inferred from uncontrolled runs, nor is the candidate a successful
latency repair. Original tracked source/docs restored exactly; candidate files
remain under .cache/garage-families-candidate-26d617c0/. Sorted path+NUL+content
SHA256 3e251447feb0c81c52d17c16fe31e55d32479c31475c50dce0b82c42b878505f.
Raw .cache/garage-families-full-26d617c0-container.log1936603bytes SHA256
c840448d57d763b25ad02ab77254669deafc2c001d3d80e841b8f9b90b402672;
functional-broad.json12516662bytes SHA256
f9f16c8bd29346b23b7dc2b1fdef9b0b6e748c2cd26912deb02ab63f33233c92.

A separate native controlled interleaving uses the actual installed provider's
registered keyboard/click commands and real pages/frames. It holds A only after
the real frame-focus evaluation, executes B's registered native pointer, then
releases A's native key. All six before-command interruption controls pass. Both
Firefox interleavings lose activation despite trusted keydown/keypress/keyup and
the same active button; Chromium/WebKit retain activation. Thus the earlier focus
predicate repair is insufficient for this reproduced interleaving, not invalidated
for its original static-interruption scope. No universal failure cause is claimed.
The12-case diagnostic exits0 because its reproduction criterion is met, NOT because
the provider is healthy. Raw Docker follow misses the already-removed fast container;
complete retained tool output owns the observations, not the failed follower file.
Probe .cache/provider-interleaving-26d617c0.mjs SHA256
4bf2893a46353c7502ab2a683dfef546a2b841405b2634c99454878059c7c3c3;
tool-output.log7448bytes SHA256
61d75a9c43f892204321eafb1adba2159ea7d2dc43bbcbe91e09a1d8f96c1cdc.

Correction under accepted CI D2: the pinned provider patch now orders Firefox's
built-in native-input commands FIFO per provider. Frame restoration plus delivery
cannot overlap another registered native pointer. Files/non-input commands and
other engines remain parallel; rejected work rejects its caller unchanged and
releases the tail. No forced destination focus, synthetic activation, assertion,
input, deadline, worker-count, gameplay or package-version change. Lock regenerated
with pinned pnpm11.15.1 in isolated Linux volumes; generic peer warning disclosed.
Frozen installation succeeds; only patch hash references change in the lockfile.

The new real installed-provider regression catches old behavior: terminal1, Firefox
pointer enters before A releases and A has zero activations. Corrected regression
passes9/9, including all-engine original-error/recovery controls; prior focus18/18
and actual Trophy Case36/36 pass (1233 deselected, not passed), frozen Linux setup
terminal0. Types0/0 and topology27 negative fixtures pass. Unified patch context
lines trigger space-before-tab in diff-check, as in the earlier approved patch;
not an implementation whitespace claim. Complete four-CPU-model target is running;
finished batch and full evidence still need independent corrective review.

Old regression .cache/provider-native-input-old-26d617c0.log3164bytes SHA256
8079e2adf6762106d965411d1122eac2ae8cddf9ddb3dcf03685fc5260b541bc;
corrected log44380bytes SHA256
f69c8e59896c23d039c714fe84356da1d02542cfb41fa32fc2a66ee3c5f0cc71;
consumer JSON291764bytes SHA256
589b9ca7b62ae78a097c39e2e0e5890956d3be57c2350c13bd01b6c8986d76fc.
Review by: Codex root, implementer first filter. Recorded by: Codex root. No
designated approval, archive, hosted measurement, release promotion or push.

The finished correction's complete four-CPU-model job is terminal0. Provider27/27;
broad501/501,40473 ordinary/18 expected/24 skips,331.03s; native3/3,87 ordinary/
six expected,5.83s; isolated Chromium performance5/5,8.13s (107 unrelated cases
deselected). No guard or absent phase. Actual executed case multisets match the
prior four-CPU baseline exactly:40515 broad plus93 native. This retains the entire
population, not just the Trophy Case. Standard14CPU/default-load reliability and
hosted proof remain unestablished; broad331s alone still misses the five-minute
workflow criterion. No matched speed comparison or whole-CI closure is claimed.
Unchanged original Garage layout remains; failed split is not installed.

Raw .cache/provider-input-full-26d617c0.log1566823bytes SHA256
186434dfdb1eb20873030027777e3471205978e1b38d2723054209d3ef06f869;
functional-broad.json12486500bytes SHA256
3a7c5ea617b8f921b42b7fee7ef3873483ca9af6b60fae3d2202775fdc93085a;
functional-native.json26235bytes SHA256
7f6664243156a250ef2190553c67c9fb4d9eb2b95d6e53a25c49825499ebf014.
All non-patch tracked diffs pass diff-check; unified-patch context warning remains
explicit. The exact finished range now goes to fresh independent corrective review.

### Independent approval — Firefox native-input ordering

APPROVED, no findings, complete nine-path range
26d617c016b37c4eba6e5c755dca27fc84fad60e..b5ea329f25253595aac6092f1fdfa742b1734c15.
Review by: fresh independent Codex /root/firefox_native_input_corrective_review.
Recorded by: Codex root under Marco's delegation. Reviewer independently executes
cold frozen Linux Trophy Case36, focus18/18 and ordering/rejection9/9, terminal0;
types0/0 and topology27 controls. Disposable exact old archive plus final regression
exits1 with trusted Firefox keys but zero activation after competing pointer. Main
tree untouched. Retained reproduction/full-run hashes and identical40515+93 case
populations independently verified; full job inspected, not independently rerun.
Non-patch whitespace clean; unified-patch context warning disclosed. All reviewer
browser workloads terminal. Approval covers this correction, not default-load/
hosted/five-minute acceptance or the accounting/version/adoption/full1.0 gates.

### Distribute existing Garage cases and overlap the measured waits

After the independent Firefox correction, reuse the preserved semantic split,
not its earlier failed-result claim. The423 existing cases now occupy seven
coherent family files; shared fixtures/assertions register no tests. All98 original
non-import statements compare byte-identically apart from helper exports. No test
body, native input, data, assertion or deadline changes. Vitest's sequencer hook
front-loads the measured event family and required real-minute idle wait within
each project. It preserves the base sequencer's exact project positions, stable
ordering of every other file and inherited sharding. No filtering, deduplication,
worker-count change, stage extraction, fake clock or shorter idle acceptance.

Four Node controls execute the actual BaseSequencer/subclass, checking priority,
project/order/membership preservation, duplicates, unknown files and real inherited
disjoint shard union. Types0/0. Complete Node client suite terminal0:147 passed/30
skipped files,12503 passed/1106 skipped cases (13609),9.73s. The new sequencer unit
fixture uses Vitest's Node API and is excluded by exact path only from browsers;
it remains in the mandatory client Node population. Canonical focused command
now selects every Garage family plus the unchanged real-minute observation.

Finished cold four-CPU-model full target terminal0, not guard: provider27/27;
broad519/519,40473 ordinary/18 expected/24 skips,342.47s; native3/3,87 ordinary/
six expected,5.78s; isolated Chromium performance5/5,8.17s (107 unrelated cases
deselected). Actual old/new path-normalized case-name multisets are identical:
40515 broad plus93 native, including all1269 Garage executions. Final JSON proves
both priority files start5–6s into each engine; real idle durations60.20/60.29/60.30s.
The measured intended ordering works. Full cold /usr/bin/time wall391.49s still
misses five minutes: no overall speedup, default14CPU reliability, hosted result,
CI acceptance or1.0 promotion is claimed. This is not a matched timing benchmark.

Raw .cache/browser-wait-order-full-70737410.log1553146bytes SHA256
257d12149d0b700006434dffe45c462421b9cdc96df7a5862e28a57fb4bd79e1;
functional-broad.json12480235bytes SHA256
dce7f03e426856d945928003e48108adde607d6c3cc2ae5e21fe8907eac42bc3;
functional-native.json26221bytes SHA256
8062d018c8a3fdcffd38360d9bb3fe7185c77284a033eeec9f76f5e7d01efe6e.
Diff-check clean. Review by: Codex root, implementer first filter. Recorded by:
Codex root. Complete batch after70737410 still needs fresh independent corrective
review. No archive, specification adoption, release approval or push.

### Independent approval — Garage families and wait ordering

APPROVED, no findings, complete15-path range
70737410df1a69fadff58c192277b89f08491f47..5a509e46e1f092566e336be06667251e37905fdf.
Review by: fresh independent Codex /root/browser_family_ordering_corrective_review.
Recorded by: Codex root under Marco's delegation. Independently verifies98 unchanged
statements/import bindings, ordering/project/shard properties, full-run hashes and
identical40515+93 case populations. Executes full Node12503 passes/1106 skips,
22.77s; types0/0; cold Linux-arm64 with verified4CPUs/affinity0-3 whole Garage
selection27 files/1281 cases passes172.71s. This includes1269 moved interactions,
three real-minute idle cases and nine existing garage-surfaces.test.ts cases.
Idle durations60.25–60.98s. Reviewer JSON SHA256
40dece37a10977f8825abdbfbd580f8ef9a2257b9506f6c89719bd0dc48faa47.
Full target inspected, not independently rerun; exact-range diff-check clean,
tracked tree unchanged and all reviewer browser workloads terminal. No speedup,
default-load/hosted/five-minute/full1.0 acceptance or specification adoption.

### Bounded coordinator profile — valid sample, full run red

Atb6818e60, run the unchanged complete target under the existing ignored4CPU
override, with durable shell redirection and /usr/bin/time. Attach the built-in
Node inspector to the single exact owned broad Vitest argv, verify process.pid3031
over the inspector, then capture one60s CPU window. No test/source/configuration,
worker, budget or population edits. Script guards fail loudly and would preserve
an invalid artifact. Actual window12:08:18.896–12:09:19.103UTC on2026-11-02 is valid:
60.205s elapsed,22463 samples,9888 nodes; Node24.18.0/Linux arm64, available CPUs4,
affinity0-3. Of60.044 sampled seconds,29.810 idle,1.958 GC and7.426 native writes:
4.268 through Playwright,1.809 Vite HTTP,1.311 WebSocket,0.039 other. These are not
simply console writes. This partial Node window neither measures browser CPU nor
proves a root bottleneck, full-run timing comparison or corrective optimization.

The same full job completes terminal2, not the guard. Provider27/27 passes; broad
517/519 project-files pass,40471 ordinary passes/18 expected failures/24 skips/two
failures,420.84s. Both fail the unchanged15s deadline: Firefox Pitch native-keyboard
launcher-to-return journey (reported18.782s), WebKit canonical partial-term group-
carry research (23.494s). They occur after the sampled window. No assertion/value
mismatch or native activation loss is established by those deadline reports. The
failure screenshot is blank after cleanup, not evidence of the interrupted stage.
Typer/performance are unexecuted after broad failure. Cold wall462.39s includes
profiling overhead; not a matched speed benchmark. Same40515 broad case population
collected. All owned execution handles terminal; no workload left running.

Ignored evidence SHA256, paths under.cache/:
- profile-browser-coordinator-b6818e60.mjs5384bytes:
  8fe7c62b7b9636f9a5fcc17e7b6db7e17cdf239f0cd048b145bd87176da3b4da
- browser-coordinator-profile-b6818e60.json2651503bytes:
  dfbc02ce388a18eebaf61bd4358ce9621c7b994f2422f92cc69c07da436750a7
- browser-coordinator-profile-observation-b6818e60.log1287bytes:
  f03afeb9cd664855d00a14f6ba2895ea2fcfad672b43482fba2060970c116eec
- browser-coordinator-profile-b6818e60.log1621809bytes:
  d56fde9d5363187b9a466224543603920e3aae737364831d5d967460a1057925
- browser-coordinator-profile-b6818e60/functional-broad.json12508510bytes:
  b3765cf69e555a24475b73fd7cfc86c5b45967c3a1fee76a19d4911d599a5c12

Review by: Codex root, observation first filter. Recorded by: Codex root. This
editorial measurement checkpoint supplies no independent implementation approval,
hosted/default-load reliability, acceptance promotion, archive, adoption or push.

### Full-load resource control — green, pressure not a sufficient explanation

At8f808c63, both retained deadline files execute unchanged through declared cold
Linux4CPU setup:33/33 three-engine cases,6.24s, terminal0. Canonical group-carry
case591/606/920ms; Pitch native journey743/689/601ms. This does not reproduce the
full-load failure. Initial sandbox-denied Docker launch executes no tests and is
retained separately; the authorized invocation supplies the actual result.
Its read-only1Hz observer is killed137 by disposable-container teardown; seven
samples retained, NO terminal record. That is partial resource evidence, not a
valid complete measurement or an unexplained passing observer.

Then execute ONE unchanged full target with the same4CPU override and passive
resource sampling, no source/config/worker/input/deadline edits. Terminal0, all
phases: provider27/27; broad519 files,40473 ordinary/18 expected/24 skips (40515),
252.20s; native3 files,87 ordinary/six expected (93),5.89s; isolated performance
five passes/107 unrelated deselections,7.51s. Actual full case-name multisets equal
the prior complete pass40515+93. Previous red run stays evidence; this is not
reliability proof, a matched speed benchmark or a cold-wall/hosted timing result.

Observer identifies only this container's processes and labels /proc pressure/
meminfo VM-wide. It retains265 samples13:56:17.085–14:00:45.167UTC,268.082s; then
container teardown kills it137 before a terminal record. Explicitly PARTIAL even
though the independently observed test command completes. During the sampled
interval cgroup deltas are979.943 CPU seconds,191.783 CPU some-pressure seconds,
0.062 memory-pressure seconds and0.229 I/O-pressure seconds. Peak cgroup memory
11327012864bytes; minimum VM available5183268KB; no cgroup OOM/throttling events.
These counters include the small observer, are not isolated browser CPU profiles,
and do not attribute earlier unobserved failures. High CPU pressure also occurs in
this green population, so pressure alone is insufficient as a failure discriminator
or authority for an arbitrary capacity sweep. No repeat of unchanged isolated
green selectors is the next diagnostic. All owned handles terminal.

Ignored evidence SHA256:
- .cache/observe-browser-resources-8f808c63.mjs3858bytes:
  31585764d1f8724bdb2c92f873aef367ac2fb70c658e9bb2b614a2466b8480b0
- .cache/browser-deadline-resource-samples-8f808c63.jsonl43007bytes:
  1f230af7ec80ded7250c93809be9861c418a4e429093df417be580a0a780153d
- .cache/browser-deadline-resources-authorized-8f808c63.log7196bytes:
  d93f2b4a245101dce80d92c25aa20f90277dd814a678cc8b0af7e0ff141291e7
- client/.cache/browser-deadline-resources-8f808c63.json10596bytes:
  1ef3bc7e9ad47618072418ac056f8a35f20e75ebae5c5e0ecf123edd58294cb3
- .cache/browser-full-resource-samples-8f808c63.jsonl1654201bytes:
  2d4a2293fd75cbcf02f74a2ab67c8c0e8b99d2408146f7ae63005c1ea964c178
- .cache/browser-full-resources-8f808c63.log1539026bytes:
  9dcd7734718cf8021dc35e0182a4e2c9e864db6ec7f38fce2e3edcaaaafc8c47
- .cache/browser-full-resources-8f808c63/functional-broad.json12482902bytes:
  2769726dd609b53dc5f340594efaa34c75329ec816563b3bdf5758a5ea693a34
- .cache/browser-full-resources-8f808c63/functional-native.json26215bytes:
  c6b7e7b3b50fb64b7ae75b32f1c0876b09279fc5b407ed64d392dd330a6b823d

Review by: Codex root, observation first filter. Recorded by: Codex root. No new
implementation approval, whole-CI/hosted/release promotion, archive, adoption or push.

### Pitch deadline-stage context — finished diagnostic batch

After7c674b96, add passive failure-only observation to the existing Pitch native
launcher-to-return journey. Arm before its first import; framework abort captures
current/completed stages, elapsed times, fixture readiness/connection, issued
command kinds, visibility and focus identity before screenshot cleanup. Remove
the listener on test finish. No production, provider, fixture behavior, native
input, assertion, deadline, concurrency, budget or test exclusion change.
AST/source comparison retains all19 existing matcher assertions and nine native
keyboard call sites character-identically, including loop key sequences.

Three new browser controls per engine: deliberate real50ms framework deadline
while a controlled gate waits; successor requires exact timeout reason, one
pre-cleanup connected-fixture capture and actual disconnected fixture; healthy/
disposed observer remains quiet. The control does not pretend its gate is a real
native command. Original nine component cases remain. Final cold declared Linux
entire file terminal0:three project files,33 ordinary passes/three expected
deadlines (36 total),8.01s. Types/Svelte0 errors0 warnings; diff-check clean.
Neither failure-context label appears during the successful ordinary journeys.
Earlier pre-label check also passes33/three expected in35.07s; not a matched
benchmark. One attempted native check is startup-blocked by sandbox listen EPERM
before collection and supplies no passing evidence. Final Linux result owns this
revision. No full browser rerun is claimed for this bounded diagnostic change.

Final ignored raw .cache/pitch-deadline-stages-final-7c674b96.log6475bytes SHA256
97a57c90028c740bb7d6bb1c8c074c7366e2169540c48048cc8be60885b8ef78;
JSON11186bytes SHA256
0238e448dfb9e78e97e995c2e267a39713cb58f5bee92d001c30877489ac055e.
Review by: Codex root, implementer first filter. Recorded by: Codex root. Complete
four-path finished range now requires fresh independent corrective review. This
does not repair the original deadline, close RP-526/whole CI, adopt a contract,
archive, approve release or authorize publication.

### Independent approval — Pitch deadline-stage context

APPROVED, no findings, complete four-path base-exclusive/tip-inclusive range
7c674b96d878fdce6421f787254e100b52c9bf72..2a469cd059477d4c022f79245288cb41a3f595cd.
Review by: fresh independent Codex /root/pitch_deadline_context_review.
Recorded by: Codex root under Marco's delegation. Independently executes declared
cold Linux whole Pitch file:33 ordinary/three expected real deadlines, all three
engines,4.42s, terminal0; types/Svelte0/0, complete-range diff-check clean. Actual
deadline successors require exact reason, one connected-fixture capture before
cleanup and disconnected fixture afterward; healthy/disposed controls quiet.
Independently confirms original assertions, nine native keyboard sites, replies
and inherited deadline unchanged. JSON11206bytes SHA256
f2152ff3bdeccad433ffe7623d70e115789017215928ad7df1bca284e259400c.
All reviewer handles terminal, main untouched. Diagnostic approval only: no cause,
full-suite/default-load/hosted/latency acceptance, adoption, archival or release.

### Ordinary default startup failure and one model-width candidate

At8179d43a, ONE ordinary default full container run exits2, not the guard.
The browser process reports an unhandled Firefox session-connection timeout,
then closed browser/createTesters RPCs; broad stops60.69s with collection and
tests still running. One project-file completes;278 provisional passes/one
provisional failure/ten skips among1612 registered cases are NOT the full
40524-case population or a stable completed-test distribution. The provisional
failure is an existing expected50ms Garden control during interrupted execution;
no new Garden product/diagnostic defect is inferred. Typer/performance unexecuted.
No Pitch deadline marker is emitted: its instrumented journey is not reached.

All process CPU affinities0-13. Resource observer retains52 samples,15:28:53.420–
15:29:47.096UTC; killed137 at container teardown, no terminal record, PARTIAL.
Peak cgroup7982288896bytes, minimum VM available8210828KB; no OOM/throttling
events. Last memory some-pressure counter2.731s. This does not establish OOM,
memory exhaustion, universal load attribution or a causal startup explanation.
Raw .cache/browser-default-pitch-context-2a469cd0.log55050bytes SHA256
812696b062b572353893cd78e8b1bd37e8e663eb20fe1cd9dc904261dfbced39;
partial broad JSON529022bytes SHA256
57dac2b6abb1a342349a2d0b7a0817c9a4f8e1ed59305f0524d5c286e508b0ca;
resource JSONL695243bytes SHA256
84aee415815c885ee0df0f17c47cb890a00bfade581ca2c9ded3191b6278ddda.

Next single candidate uses the width from D1's fourCPU model: at most three
files PER engine, fewer on smaller hosts, equal root maxWorkers for all projects.
This avoids the14CPU VM's inherited12-per-engine multiplication without changing
the4CPU model's existing width or adding serial phases. Multiple4CPU complete
passes and one red observation inform that control; they do not prove a safe
capacity. All engines, cases, inputs, assertions, deadlines and phases remain.
Only complete target terminal0/all phases qualifies as a passing candidate;
red/incomplete means withdrawal, not retries or a worker-count sweep. Fresh
independent corrective review remains required if it passes.

Finished single candidate terminal0 through the ordinary default14CPU container,
no affinity or CLI worker override. Provider27/27; broad522 files,40485 ordinary/
21 expected/24 skips (40530),269.99s; native3 files,87 ordinary/six expected (93),
6.38s; isolated Chromium performance five/107 unrelated deselections,11.38s.
Exact case-name multiset comparison against the prior complete green population
has NO missing40515 broad cases. Added15 executions are solely nine previously
approved Pitch controls and six new width controls. Native93-case multiset equal.
No new serial phase, omission, budget, native input, assertion or dependency edit.
This one passing control is not proof of universal capacity, a causal explanation,
reliability, matched speedup, complete cold-wall or hosted timing acceptance.

Pure width controls cover1/2/3/4/5/14/64/MAX_SAFE_INTEGER CPU observations and six
invalid inputs. Types/Svelte0/0 and topology27 negative controls pass. The first
mistyped `make test-client -- --run ...` actually runs the full Node12505/1109
population green, then Make exits2 for the unsupported target; NOT a green command.
A subsequent host pnpm selector fails resolving its pinned manager before tests
(DNS restriction), not source failure. Final valid `make test-client` exits0,
same12505 passes/1109 skips (13614),8.89s. No test run is credited to those wrappers.
Full resource observer is again killed137 by disposable-container teardown;
retained raw samples are explicitly partial, not a complete passing instrument.

Ignored evidence SHA256:
- .cache/browser-model-width-8179d43a.log1531294bytes:
  fb90e46d5c031675cddb73983430c7fa1dc5d3fd4d42017b2c326b5a73459e11
- .cache/browser-model-width-8179d43a/functional-broad.json12512323bytes:
  6467778ad5326074b1891c95fe0029810ae61d2f80ad48bad5c34c003c8a763c
- .cache/browser-model-width-8179d43a/functional-native.json26237bytes:
  da3068ce1a35373238aba4889a9b1cb9379a739c4b620ba2503dc3b1ac121fd7
- .cache/browser-model-width-resources-8179d43a.jsonl1566281bytes:
  f42980cc1027a92fbb5f9151f139d5105050492c00f27efecdde79b93216094c
- .cache/browser-model-width-client-8179d43a.log352bytes:
  e3985a1559dac3394fd9124dc8d70972a0dbdd804b87593bd37bda4e14285ff0

Review by: Codex root, implementer first filter. Recorded by: Codex root. Complete
six-path source/test/config/docs/records range after8179d43a now needs independent
corrective review. No whole-CI acceptance, archival, specification adoption,
release approval or push. Tier2's separately received owner adoption is another lane.

### Independent approval — measured-model browser width

APPROVED, no findings, complete six-path base-exclusive/tip-inclusive range
8179d43a69450d7d37caf35c5111b4424c59ac9d..7cc5f4b88597820750ccc03461e2526d325bb95e.
Review by: fresh independent Codex /root/browser_model_width_review.
Recorded by: Codex root under Marco's delegation. Independently runs ordinary
declared Linux complete target, terminal0: provider18+9, broad522 files/40485
ordinary+21 expected+24 skips (40530),181.86s; native87+six expected (93),6.26s;
isolated Chromium performance five,7.69s/107 unrelated deselections. Actual
installed resolveConfig reports14 CPUs/affinity0-13/maxWorkers3, all engines,
headless/fileParallelism true; no CLI/affinity override. Installed source verifies
per-project semantics. Independent Node12505/1109skips terminal0, types0/0,
topology27 controls and exact-range diff-check pass. All old40515 broad cases
remain; added15/natives93 and final reviewed multisets match root exactly.
All five retained candidate digests independently match.

Reviewer raw .cache/browser-model-width-independent-7cc5f4b8.log1485506bytes SHA256
4b4fb223d0b412498014d1cdfca6acecc46a9a45b82df10388d3d8af0df41e68;
its functional-broad.json12513842bytes SHA256
c7a099eabbf4acc79effacc72220da4e44e58bc97c601c82f156770f5fb5ee28;
functional-native.json26223bytes SHA256
67ca8bbf40553cad40970a215381b96f58610ee88ea76a8a2764f2c0f3bf2ad2.
Initial sandbox denial executes no tests; authorized invocation then passes.
All handles terminal; owned container removed; main untouched. Later Tier2 range
excluded. This closes corrective config review, not universal safe capacity,
causation/reliability, matched speedup, cold-wall/hosted five-minute acceptance,
spec adoption, archival or release. No publication.

## 2026-11-02 — Historical scope policy technically reviewed, not adopted

APPROVED, technical specification only. Reviewed single document
rfc/kernel-history-guard-integrity.md at7b06dd5d0af5675cb69f560b2f58751d98835707,
Git blob667e9c93af0931b0ee24506a2c0b1f569c8e147d.
Review by: fresh independent Codex /root/kernel_scope_policy_technical_review.
Recorded by: Codex root. No prior full normative approval located. Independently
inspects parent KV-1, accepted CI D1–D3, current guard/registers/fixtures and
Make/checkout wiring. Exact ancestor/path sets, independent trace-preserving
eligibility, N2 descendant registration and separate semantic corrections are
implementable without false versioning, rewriting or automatic amnesty.

Actual kernel fixtures pass:60 literal-path refusals plus legitimate version/
correction/rename/full adversarial controls; checkout-history fixtures exit0.
Document/whitespace and final status clean. Implementation must use historical
parent/child path registries and reject actual review-log resolution escapes,
not just lexical traversal. No target eligibility, implementation, owner adoption,
full CI or release is approved. No tracked reviewer edits/services/publication.

Root's new actual make verify-client atba1d7eaa plus Headcount witness passes
types/build/Node/boundaries, then reproduces unchanged50a3a514 failure (Make2);
the Headcount log owns exact evidence. The new witness has separate full-range
corrective approval. Owner adoption/implementation delegation for the reviewed
kernel policy is now requested through the async question interface, not inferred.
No policy/exception row or watched prefix changes. Continue accepted work;
do not rerun unchanged green browser populations or claim this fixes history.

## 2026-11-02 — Reconcile newer hosted CI and validate completed maintenance artifact

Read-only GitHub branch/run/job/log queries now confirm main at
d6b1c6689574a4823473a44c9fe51d4659d5fd09 and completed-failure CI41476621196,
not the older be66902b coordinate still named by the overview. No workflow rerun,
push or deployment performed here. Later local Headcount changes are not in this run.

| Job ID / job | Result | API job elapsed |
|---|---|---:|
| 125383759822 / server | FAIL: TestAxisAccrualPartitionProperty,27 subcases | 524s |
| 125383760225 / client | FAIL: historical kernel guard50a3a514 | 103s |
| 125383760426 / harness | PASS | 154s |
| 125383760441 / game-ui-composed | PASS | 375s |
| 125383760563 / browser | PASS, all declared phases | 436s |
| 125383760722 / schema | PASS | 15s |

These are actual started_at→completed_at job intervals, not fixture durations.
Browser's make step is376s; broad522 files:40485 ordinary+21 expected failures+
24 existing skips=40530,329.81s. Provider18 focus cases/zero misses and9 native
input controls execute. Native Typer87 ordinary+six expected=93,6.58s; isolated
Chromium performance five pass/107 unrelated deselections,12.59s. This is a
genuine hosted complete pass for the reviewed provider/worker-model code, not
proof of universal reliability, failure causation or the strict five-minute floor.
Browser/composed/server jobs still exceed that floor. Client types/build/Node
12505pass/1109skips pass before the same historical guard; server logs contain
the parent partition test plus exactly27 failed children, not28 independent cases.
Expected database-negative ERROR lines are not additional failing tests.

Latest inspected maintenance41383212848 is a DIFFERENT source,
efa9223cf11462829560455f4418edceaeb67bb3. Harness-evidence125078293010 and numeric
maintenance125078293205 PASS; Reputation125078293376 FAILS career and relevance
report comparisons. Its logs retain six Casual starter ties and970 admitted/
retained relevance arms. No report is regenerated or failure erased.

Unlike the older38677066823 artifact, this harness observation actually completes
all four objectives. Download artifact13116276146 using the declared GitHub read
surface; initial sandbox gh download cannot connect and executes no measurement,
authorized read succeeds. ZIP1446bytes SHA256
c053b9ca0b59e8bc519c9b6b25e8ed31007b279ff0515b3f3fc1b8ecc01c3be4
matches GitHub's artifact digest. Archive contains only harness-observation.json,
5694bytes; retained ignored .cache/hosted-harness-41383212848/harness-observation.json
SHA256d8c2eee9851f04eb8737eb3d8ed48ec7ebc6fad8132ef0f4130e00595cf95692.
Actual root make harness-observation-check on those extracted bytes exits0.

The artifact is complete/objective,2979091ms: guards24447ms, standard pacing
300/300 in2579394ms, active relevance107/107 and1968171/1968171 transitions in
374962ms, fixture relevance23/23 and3324/3324 in260ms. Guard/exclusion/truncation
statuses clear; declared instrument-excluded IDs remain visible (three active,
one fixture), not evidence those roles were measured. Actual epoch8/hash and
scenario/policy/report identities are retained. This validates completeness on
its recorded published source, not current-source equivalence, full1.0 balance,
Reputation success or spare operating capacity: only20.909s separates its recorded
elapsed time from the existing50-minute wrapper. No ceiling increase or active
landing of the unapplied rate candidate follows. Later first-hour/Tier2 code
differs; no matched speedup or historical-source identity is inferred.

Source links: CI https://github.com/stronk-dev/destroy-humanity-any-percent/actions/runs/41476621196
and maintenance https://github.com/stronk-dev/destroy-humanity-any-percent/actions/runs/41383212848.
Review by: Codex root, evidence reconciliation/first filter. Recorded by: Codex root.
These editorial evidence updates do not adopt pending contracts, supply designated
feature approval, change workflows/populations/budgets or publish anything. Next
is reviewed contract adoption/implementation and actual corrective work, not
repeating this completed measurement or treating an old failed artifact as latest.

## 2026-11-02 — Reputation observation contract technically APPROVED

Review by: fresh independent Codex `/root/reputation_observability_contract_review`.
Recorded by: Codex root, transcribing its actual verdict under Marco's technical-
specification review delegation. **APPROVED**, technical specification only.
Exact inspected source `aef938c3d465e996f304d90bc3406e7319be10d8`, RFC blob
`97c3fb4333852b933e72f685900679bfd0546f6c`; complete original five-path proposal
`ab3442dc31328557b924a104578d3e332e5856b1` exclusive through
`5fb680096bdd409c795166d6f9c87275a64355d9` inclusive, no intervening RFC amendments.

Reviewer checks actual recorder/validator/signals, CI D3, R10 producers, source
admission/declarations, report serialization/recomposition, Make and workflow.
194 H4 and970 H5 arms are correctly specified. The v2 fixture profile must extend
shared internals to retain H5 progress after failed H4 while preserving v1; O3/AC3
already require it. Fixed companions, committed pins, actual source arrays, Make
signal propagation and missing-checkpoint refusal have executable acceptance
requirements. Current legacy pins drift: complete measurement/reproducibility is
not H4 balance acceptance; no pin rewrite or new H5 first-elective claim is allowed.

Cold focused recorder/registered-row/SIGTERM and actual career-source/admission/
declaration/recomposition/population controls PASS Harness9.409s and CLI0.604s;
affected vet, topology27 negatives and proposal whitespace pass. First selector
fails shell parsing before execution; corrected invocation supplies actual proof.
No exhaustive study, DB/hosted/root-wrapper SIGINT or new v2 runtime is claimed.
Initial tree clean; concurrent parent drift-diagnostic edits seen at finish are
explicitly OUTSIDE this verdict, not silently included in its approval.

Only status/index metadata is reconciled here. Adoption, v2 implementation and
its independent review, RP-634 closure, hosted evidence and full feature/release
gates remain open. Existing workflow, budgets, assertions and reports unchanged.

## 2026-11-03 — RP-706: isolated Linux performance selector

Under CI D2 and RFC-0000's routine tooling procedure, add
`make test-game-ui-performance-ci`. It dispatches the existing performance target
through the shared browser entrypoint, pinned image, cold dependency volumes,
frozen install and ten-minute KILL containment. No workflow, Compose configuration,
test body, performance threshold or default full browser command changes.

Test first: add the real-script tool-double case for the performance selector to
the existing topology fixtures. Against the old entrypoint it exits1 from the
fixture with script status2, no setup calls and `expected test-browser or
test-browser-focused` (session2660). The correction explicitly allows only the
existing third target. Fixture controls execute all three routes, exact setup and
argument boundaries; apt-get, pnpm and Make failures37 propagate on every route;
missing/unknown/near-match targets refuse before host tools. Narrowed defaults or
performance dispatch fail the topology check. `make verify-ci-topology` PASS
(17431), including existing negative controls. Bash/Node syntax and whitespace
checks PASS.

Actual subject: `make test-game-ui-performance-ci` PASS (99422 terminal0), cold
Linux container on this ARM64 Docker host. Five unchanged Chromium scenarios pass,
107 functional cases unselected by the existing performance target; test durations
413/703/1166/1151/2066ms, total7.48s. Those are test durations, not new runtime
budgets or performance annotation counters. Initial restricted invocation denies
Docker access before tests and supplies no evidence; the permitted invocation is
the result above. Existing route control:
`make test-browser-ci-focused BROWSER_TEST_FLAGS='test/numeric-monotonicity.test.ts'`
PASS (34719 terminal0), 300 tests across Chromium/Firefox/WebKit,7.79s. Both actual
runs install frozen dependencies into fresh volumes; no repeat-to-green or
assertion relaxation. Containers are --rm. Full browser population, hosted runner
hardware/CI latency, manual performance floor and whole-feature acceptance are not
claimed.

Review by: Codex root, implementer first filter. Recorded by: Codex root.
The complete eight-path tooling/docs/record batch after
`58c0264acc1c06f78f2caf809566e47433ad89ef` through this commit awaits fresh independent
corrective review. RP-706 is corrected locally, not self-approved. Next: independent
review of this exact range; preserve the separate original performance-oracle
approval and broader open performance/CI/release gates.

## 2026-11-03 — RP-706 corrective review APPROVED

Review by: fresh independent Codex `/root/linux_performance_selector_review`.
Recorded by: Codex root, transcribing its actual APPROVED verdict, no findings.
Complete eight-path range
`58c0264acc1c06f78f2caf809566e47433ad89ef..7b74bc74ff43066062c605eb372535fe1a621be2`
(base exclusive, tip inclusive), including the final batch records.

Reviewer independently executes the old entrypoint extracted from the base:
performance target exits2 before setup. `make verify-ci-topology` passes27
primary negatives plus default/dispatch and real-script success/failure controls.
Actual `make test-game-ui-performance-ci` exits0:five Chromium cases pass,
107 unselected,7.55s. Existing `make test-browser-ci-focused
BROWSER_TEST_FLAGS=test/numeric-monotonicity.test.ts` exits0:300 tests across all
three engines,2.39s. Both container runs use fresh frozen installs. Initial
restricted Docker refusal executes no tests; permitted runs supply the evidence.
Bash/Node syntax and exact-range whitespace pass. No tracked reviewer edits;
clean at reviewed tip, both sessions terminal, browser inventory empty.

This closes RP-706's corrective review only. Linux/ARM64 is not hosted AMD64 or
hosted timing. Full CI, AC7/manual performance, broader feature/archival/release
gates remain open. Root updates current statuses; no suite for these editorial
records, no push or product/specification change. Full1.0 goal remains active.

## 2026-11-04 — Node fixture discovery correction independently approved

Review by: fresh independent Codex `/root/node_fixture_collection_corrective_review`.
Recorded by: Codex root, transcribing its APPROVED verdict, no findings. Complete
seven-path `f423e704dfd0cdaaeeaac9dcc51891b82380021a..663976f1d88f23b83c421a47084ac16f4b1f70c3`
(base exclusive, tip inclusive), including all original records. Both100% renames
and Make references are the entire executable change: no assertions, exclusions,
workflow topology, retry, deadline or production change. Original fixture blobs
remain care028c4207869a6d558a0a42b4c9b85051a5e37293 and
productioncf88a62e65d6e215816c45eb17267ba557536f96.

Fresh at6076a1c4: actual current owning Make-lane Node command35/35 passes,
zero skips (original30 unchanged plus five separately added N5 controls).
Private exact-old reconstruction under actual current Vitest configuration
fails BOTH with "No test suite found", exit1, despite separate Node execution.
Changing only those private suffixes selects neither in Vitest and retains
Node30/30. Current discovery still selects legitimate manual-budget-observation
`.test.mjs`; this is suffix/runner ownership, not a blanket mjs exclusion.
`make verify-ci-topology`47669 terminal0 includes27 negative controls and shared
Linux/Make-selector controls; exact-range whitespace passes. No tracked mutation.

Affected app/server/main driver/proof helpers/fixtures/Make/workflow/declared
Compose diff from independently executeddf531b9b is empty. Its full composed
86769 is REUSED with GarageNov3 reviewer provenance, not freshly executed here.
Original full-client19241 and broader RED native39477 remain historical; no
fresh whole-suite, full CI/hosted-latency/feature/archival/release proof follows.
Reviewer starts no services, leaves no active handles or tracked edits/commits;
private fixture retained at `/private/tmp/cloud-clicker-rp396-review.uogbttob`.
Root reconciles current statuses only; no software suite for these editorial
records. Full nine-tier goal remains active; no push or adoption.

## 2026-11-04 — Linux browser command parity independently approved

Review by: fresh independent Codex `/root/linux_browser_parity_corrective_review`.
Recorded by: Codex root. APPROVED, no findings, complete nine-path range
`8110755490b3596b73d4330eaa4f5815d29027c4..a57b7fd862949c4b4b61a005fd9c74a32d3235ef`
(base exclusive, tip inclusive), including all original records. Missing Make is
installed in the disposable image; root browser phases share Actions' authority.
Image/frozen install/population/assertions/workers/deadlines/containment unchanged.

Fresh topology49224 terminal0:27 primary negatives, copied/incomplete/narrowed
container controls, actual shared-script success/failure and Make/report selectors.
Private exact-Git controls30770 terminal0: old verifier admits old copied command;
corrected verifier rejects it and four changed commands, honest configuration
passes. Actual historical Make source, with ONLY Vitest replaced by a disposable
executable double, runs broad→native→performance and stops/propagates each leaf's
injected failure (four controls). This proves orchestration, NOT browser behavior.
Bash/Node syntax and complete-range whitespace pass; private fixtures removed.

Actual Linux evidence is REUSED with independent provenance: API component reviewer
78930 at6076a1c4 executes837 across all three engines; affected executable diff is
empty. Linux performance-selector reviewer at7b74bc74 executes performance5 and
numeric300/three engines; relevant runtime/selected tests/Make/Compose/shared setup/
topology/config/package/lock/patches match. No fresh full-browser run here. Original
52310 timeout and8e4d0877 completion retain implementer provenance. Later shared-
setup/RP-706 repairs remain separate reviewed scopes. Both handles terminal;
reviewer creates no services/Docker resources/tracked edits/commits. Root's concurrent
e853255b is Markdown-only. RP-613 review closes, not RP-526 reliability, default/
hosted latency, full CI, feature/archival/release acceptance or the nine-tier goal.

## 2026-11-04 — Citation paths approved; original review parser requires changes

Review by: independent Codex `/root/kernel_review_citation_corrective_review`.
Recorded by: Codex root. Complete original contexts, all8/7 paths inspected:

- RP-131 CHANGES REQUIRED:
  `a7e33b73a39a435157c7049a8d13abdd68a63c4d..611273b5f6c8ce15126f163ff473d48bb2239ed2`.
- RP-540 APPROVED:
  `611273b5f6c8ce15126f163ff473d48bb2239ed2..fce6c3d94a120b1d7e1188ddb7998ac6d5fd480e`.

At a6fa32ac, private actualCLI11164/37946 admits lazy quoted Review by/Decision/
anchor (RP-716), while honest visible review passes and explicit quote refuses.
[CommonMark0.31.2§5.1](https://spec.commonmark.org/0.31.2/#example-247) confirms
markerless paragraph continuation remains quoted; no new governance is needed.
Fresh37946 also runs exact old611273b5/current citation checks: four aliases
admitted old/refused current before lookup. Full52939 fixtures include NUL/
exterior/canonical active/archive/fallback and60 literal-path refusals; topology
99483 passes27 negatives/current Make controls; syntax/range checks pass.
Root90711 remains expected Make2 at50a3a514. DP-E transcription correctly keeps
Claude reviewer/recorder and CHANGES REQUIRED; actual3f58fea..65099e7 contains
09d5027 and the original cold missing-bump finding. No new review is inferred.
No tracked reviewer edits/services; all handles terminal. Owned fixtures retained
under `/private/tmp/cloud-clicker-citation-review-gFk675` and `...-AzlyUe`;
exterior files remain inside owned storage. Later scope/literal-path fixes stay
separate; no symlink/inode/authentication, history exception or full-CI claim.

## 2026-11-04 — Exclude lazy quote review metadata (first filter)

Existing KV-1's quoted-example exclusion now keeps quote-paragraph continuation
hidden after an omitted marker. Track quote fences/paragraphs and actual blank/
heading/list/fence/thematic boundaries in the bounded ledger visibility filter;
tabs use Markdown indentation. Noninterrupting ordered text and an invalid
backtick fence cannot expose later quoted metadata. No renderer/dependency,
correction-row/path registry, kernel/math/version, workflow, exception or author
attribution change. This does not authenticate review execution.

Root actual old-code17424 exits1, admitting the three initial lazy/nested/mixed
quote shapes. Initial corrected59836 passes. Expanded final91228 fixtures pass
eight quote negatives, six legitimate visibility exits and all existing literal/
correction/miss/remap/immutability/ancestry/archive cases. Topology47931 passes27
negatives/shared scripts/Make controls; both Node syntax checks and diff pass.
Actual root `make verify-kernel-version`91964 passes checkout controls then exits
expected Make2 at original50a3a514, not a newly refused legacy review binding.
No Go/browser/DB/hosted run for this CLI-only change. Operations captures are
concurrent on unchanged server/Make/Compose bytes; root never rebuilds their binary.
Complete six-path follow-up is a6fa32ac-exclusive through this containing tools/
tests/docs/ledger/plan/log commit; designated independent union review remains.
Review by: Codex root, implementer first filter only. Recorded by: Codex root.
Original CHANGES REQUIRED, unadopted history proposal and full nine-tier goal stay;
no archival, push, release or self-approval.
