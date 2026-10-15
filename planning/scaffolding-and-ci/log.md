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
