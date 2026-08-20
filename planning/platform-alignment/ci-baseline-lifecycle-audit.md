# CI Baseline lifecycle audit

Coordinate: product tree `190a4fa`; hosted planning coordinate `cb162a3`; audit checkpoint after
`aa461c1`; 2026-08-20.

This pass re-derived the active CI RFC from its complete specification, plan/log, current workflow,
root Make graph, schema verifier and fixture populations, canonical CI docs, current public Actions
runs, R-001 diagnosis, successor harness/Game UI records, implementation history, and designated
review entries. It made only restored discrimination mutations while measuring and did not edit
product code, owner-authored RFC/design text, canonical product docs, or implementation-plan
checkboxes.

## Bottom line

The repository has a useful, mostly healthy set of individual hosted gates. At the latest remote
coordinate, server, client, browser, schema, and composed-browser jobs all pass in roughly 0.5–2
minutes. Shared Decimal corruption and malformed production catalogs demonstrably turn their
relevant gates red. Permissions are read-only and concurrency cancellation is branch/ref scoped.

The CI Baseline as a release guard is nevertheless false at HEAD:

- every current-product workflow ends cancelled because the blocking harness job is killed at its
  30-minute ceiling;
- the normative RFC requires the complete blocking workflow to finish in under five minutes and
  says the balance harness is out of scope, but the current workflow makes that harness a
  30-minute blocking job;
- the RFC's four-job/pinned-browser-container topology has become six blocking jobs on ordinary
  runners without body reconciliation;
- the implementation plan still says the “first hosted workflow” has not been observed, despite
  dozens of runs and repeated current cancellations;
- setup-go defaults make the harness and numeric jobs cache Go build outputs, contradicting the
  RFC/canonical dependency-only cache claim; and
- scheduled/manual events run the whole blocking workflow, while `numeric-maintenance` has no
  non-blocking failure policy. A historical numeric-maintenance failure produced an overall failed
  scheduled workflow, directly contradicting the “non-blocking nightly” label.

AC2 and AC4 are now proven. AC1, AC3, and AC5 are contradicted at the current coordinate. R-001's
first wave correctly refuses a timeout increase or workload reduction, but its required
instrumentation cannot silently land under an RFC that explicitly excludes the balance harness;
an accepted scope amendment or named harness owner is required first.

## Hosted evidence

The public repository's latest remote commit is `cb162a3`, seven planning-only audit commits behind
this local checkpoint and byte-identical in product/workflow behavior to `190a4fa` for these
criteria. GitHub Actions run `32404232364` started 2026-08-20 18:37:40Z and concluded `cancelled`:

| Job | Result | Execution window |
|---|---|---|
| server | success | 18:37:44–18:39:44Z; `verify-server-core` 1m33s |
| client | success | 18:37:44–18:39:17Z; `verify-client` 1m12s |
| browser | success | 18:37:43–18:39:32Z; install plus tests under 2m |
| schema | success | 18:37:43–18:38:14Z; verifier under 1s after setup |
| game-ui-composed | success | 18:37:44–18:39:31Z; composed gate 58s |
| harness | **cancelled** | 18:37:44–19:08:02Z; governed step killed after 30m02s |
| numeric-maintenance | skipped | push event only |

The harness setup completed at 18:37:57Z; `make verify-harness HARNESS_WORKERS=12` ran until
19:07:59Z without a verdict. Runs `32009994004`, `32096019304`, `32212696707`, `32328790752`, and
`32404232364` repeat the same current-product conclusion. R-001 establishes the last hosted green
at `146deded` and 6m16s, while the later `a6328df` capacity commit had no hosted run of its own.

Scheduled run `31562066740` is also discriminating lifecycle evidence: `numeric-maintenance`
failed and the overall scheduled workflow concluded `failure`. The job may be outside PR latency,
but it is not non-blocking within the scheduled workflow as D3/docs claim.

## Restored discrimination probes

Two temporary mutations were predeclared, executed, and restored with patches:

1. The first shared Decimal `0^0` expectation changed from `1` to `2`. `make test-client` exited 2
   with exactly one failed shared-vector test and 6,654 passing tests. Restoring `1` returned the
   full client population to 6,655 passed, 15 skipped.
2. A temporary `balance/catalogs/platform-alignment-invalid-probe.json` declared schema v4 plus an
   unknown field and omitted required catalog arms. `make verify-schema` exited 2 with exact missing
   and additional-property diagnostics. Deleting the probe returned the complete verifier to green.

The worktree after restoration differs only by the user's pre-existing `AGENTS.md` edit. These
probes are current executable failures, not historical prose about a mutation that once failed.

## Acceptance classification

| AC | Verdict | Evidence and limitation | Required closeout |
|---|---|---|---|
| AC1 | **Contradicted at current HEAD** | Individual local/hosted gates are substantially green, and a pre-current full local aggregate completed in the epoch-8 review range. No current hosted aggregate reaches a verdict: the harness job cancels every workflow. The initial audit's local `make verify -count=1` was also interrupted at the same silent harness phase and is explicitly invalid, not a local green. | Complete R-001 instrumentation and measured repair, then require one clean-checkout local aggregate and a current hosted workflow to reach actual exit. |
| AC2 | **Proven integration** | Current hosted `client` and Chromium/Firefox/WebKit `browser` jobs pass. The restored broken-vector probe makes the actual Node client population fail exactly at the corrupted shared vector; the same corpus is consumed by both Go and TypeScript. Historical cross-architecture fixes and current frozen installs preserve the relevant job routing. | Preserve the current mutation result in this tracked audit and include the vector/client/browser workflow ranges in final review union. |
| AC3 | **Contradicted at current HEAD** | The normative complete blocking workflow limit is five minutes. The current blocking harness alone is killed after 30 minutes; the last hosted green cited by R-001 was already 6m16s. The body says slower work moves non-blocking and the blocking budget does not grow, while later work did the opposite without reconciling this RFC. | Owner/ruling author chooses and writes the actual blocking topology/latency contract after R-001 cost evidence; do not raise the ceiling from killed runs. |
| AC4 | **Proven integration** | All 20 checked-in `*.schema.json` files are explicitly compiled/consumed by `verify-schema.mjs`; production and positive/negative populations are nonempty. The restored malformed production-catalog probe makes the command itself exit nonzero, then the full current population passes after removal. | Preserve population enumeration and the production-path failure probe in a checked-in non-destructive fixture harness or equivalent reviewed test. |
| AC5 | **Contradicted at current HEAD** | `contents: read` and ref-scoped cancellation match. Cache and trigger behavior do not: harness/numeric omit `cache: false`, and setup-go enables module plus build-output caching by default; canonical docs claim module-only. Schedule/manual triggers execute all ordinary blocking jobs, and numeric maintenance has no `continue-on-error`; run `31562066740` proves its failure fails the workflow. D2's four-job pinned-container topology also no longer matches the six-job ordinary-runner workflow. | Ruling author reconciles D2/D3 to the intended successor topology; explicitly disable build cache or change the contract; isolate/mark nightly behavior honestly; add a static workflow-contract test with failing fixtures. |

No plan checkbox was changed. Plan item 6 is not “hosted acceptance pending”; hosted acceptance has
repeatedly failed, and the item must be rewritten rather than merely left unchecked.

## Workflow-contract trace

| Contract | Current reality |
|---|---|
| Read-only permission | Exact `permissions: contents: read`; matches. |
| Push and pull request | Present and demonstrated by hosted runs; matches. |
| Superseded-run cancellation | `ci-${workflow}-${ref}` with `cancel-in-progress: true`; structurally matches. |
| Four blocking jobs | Six unconditional push/PR jobs now exist: server, harness, client, browser, game-ui-composed, schema. |
| Pinned Playwright container | Removed after a real environment failure; browser now uses ordinary Ubuntu plus explicit pinned browser install. This later topology is hosted-green but the RFC body was not reconciled. |
| Dependency-only Go cache | Server uses explicit module-only cache; harness and numeric use setup-go's default module-and-build-output cache. |
| Nightly non-blocking | Same workflow executes all blocking jobs; numeric failure affects workflow conclusion. |
| Under-five-minute blocking workflow | False by current hosted timestamps; harness exceeds 30 minutes. |

The setup-go cache behavior was rechecked against the
[v6 action's primary README](https://github.com/actions/setup-go/tree/v6#caching-dependency-files-and-build-outputs):
caching is enabled by default and covers “Go modules and build outputs”;
`cache-dependency-path` changes dependency identity, not the cached output classes. This makes the
cache mismatch executable workflow truth, not an assumption about action internals.

## R-001 status and authority gap

R-001 first-wave diagnosis is complete and strong:

- five current hosted cancellations and one current local 12-worker completion are recorded;
- workload growth from two to nine generator classes and the T0–T1 systems is traced;
- worker parallelism is proven to cover the 300 pacing tasks but not serial relevance rows/arms;
- standalone scenario runs are rejected as authority-equivalent registry selectors; and
- killed runs expose no row, transition, exclusion, guard, or truncation artifact.

R-001 itself remains open. Its next experiment requires phase/row timing, cardinality, invalid
termination output, and an authority-preserving registered-row selector before any optimization,
shard, or budget decision. The active CI RFC explicitly says the balance harness is out of scope;
the archived harness/content RFCs cannot authorize new implementation. Therefore the diagnostic
instrument needs an accepted CI amendment or a new accepted harness-observability RFC. Treating
“diagnostic-only” as permission would repeat the exact scope-smuggling this program exists to stop.

## Normative, canonical, plan, and review drift

1. Summary/Motivation/D2/D3 call this a four-job, under-five-minute starter and explicitly exclude
   the balance harness. Current topology and hosted behavior contradict all three claims.
2. D2 requires the Playwright container; the browser repair correctly moved to ordinary Ubuntu,
   was designated-reviewed, and is currently hosted-green, but the ruling author never reconciled
   the normative body.
3. D3 calls numeric maintenance non-blocking, while the workflow and a historical scheduled
   failure prove otherwise. Scheduled/manual events also run the complete blocking population.
4. Canonical `docs/ci.md` omits `game-ui-composed` from its “Blocking jobs” table even though later
   prose admits it runs on every push/PR. It calls Go caches module-only despite two jobs caching
   build outputs, and calls numeric maintenance non-blocking without a workflow mechanism.
5. The plan/log begin with “first hosted run pending,” then accumulate numerous hosted failures and
   repairs without reconciling the plan. Its last two CI entries are marked ready for designated
   review; successor logs later approve some exact commits, but this record does not link them.
6. Designated verdicts exist for several later slices (`02d00d7`, `928126b`, `e42a506`, the
   epoch-8 range through `a6328df`, and other successor batches). They are scattered across Game UI,
   Minigame, and archived T0–T1 logs. No CI closeout record declares ranges whose union covers the
   initial baseline plus every current workflow/Make/schema repair, and the active RFC is false on
   three criteria anyway.

## Smallest honest closeout order

1. Owner/ruling author reconciles the CI RFC's scope and latency law: decide which jobs are
   blocking on push/PR, which are nightly/manual, and whether the five-minute rule remains binding.
2. Accept the R-001 observability contract under a named active RFC; implement only phase/row
   timing, authority-preserving selection, and fail-loud termination artifacts with discriminating
   complete/incomplete fixtures.
3. Measure identical complete populations locally and on hosted Linux. Only then choose
   algorithmic optimization, safe parallel dispatch, or sharding; do not reduce scenarios or raise
   timeouts from current killed runs.
4. Reconcile setup-go cache policy and make nightly workflow-conclusion behavior explicitly match
   the adopted “blocking” meaning. Add a repository-owned workflow structure test whose seeded
   cache/trigger/permission/cancellation mutations fail.
5. Reconcile `docs/ci.md`, the implementation plan, and successor review links in the same range.
6. Construct the exact current-history implementation union and obtain the mandatory tracked
   cross-party verdict.
7. Run one clean-checkout local aggregate and one hosted workflow through final exit. Archive only
   if every retained criterion is true at that coordinate.
