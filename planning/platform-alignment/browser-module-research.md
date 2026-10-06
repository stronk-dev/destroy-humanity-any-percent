# Cold browser module and worker observation — 2026-10-06

Authority: accepted CI Baseline D2/AC1/AC2; predeclarations `c8ff3139`,
`6b119e76`, `e6ea05bf`. This is observation, not a production correction.
Full proper nine-tier 1.0 remains the objective. No tests, engines, bounds,
dependencies, production worker behavior or workflow requirements were removed.

## Executed populations

| Handle | Population | Result |
| --- | --- | --- |
| 54700 | Original complete cold matrix | RED: Firefox Route JSON import fails; 299/300 files, 22466 tests pass / six existing skips; sixteen Route tests never imported. Original HTTP cause unknown. |
| 74841 | Unchanged matrix with existing provider request-failure logs | PASS: 300/300, 22482 / six; 53.02 s, separate performance 525 ms / 2.49 s. |
| 2773 | Same unchanged population | RED: original Firefox worker witness emits no prediction; unchanged output 100, original five-second assertion and native asynchronous error fail; 22481 pass / one failed / six skips. Route import succeeds. |
| 82623 | Intentional Route fixture HTTP 404, three engines | RED: all three imports fail, no Route tests executed; observer 18 starts / 18 finishes. Firefox reports NS_ERROR_CORRUPTED_CONTENT. |
| 88496 | Restored full instrumented population | PASS: 300/300, 22482 / six; 43.26 s, separate performance 487 ms / 2.71 s. Tool retrieval truncates the middle of 56,774 tokens; **invalid as complete retained traces**, not included in the artifact. |
| 67300 | Intentional native-worker module HTTP 404, original case in three engines | RED: three cases fail / 66 selector skips; real HTTP 404, native error while worker not terminated, zero output predictions, original assertion and asynchronous error fail. Observer 3 starts / 3 finishes. |
| 93957 | First complete restored population with continuously drained capture | PASS: 300/300, 22482 / six; 49.32 s, separate performance 603 ms / 2.50 s. HTTP summaries 129/129 plus separate performance 1/1, zero pending. |
| 3595 | Second complete-population invocation, same source | **INCOMPLETE**: 100 Chromium / 100 WebKit / 99 Firefox files report completion; Firefox Garden raw-catalog file missing. Provider reports NS_ERROR_FILE_NO_DEVICE_SPACE loading Vitest runner. Owned container stopped after existing ten-minute CI ceiling, exits 137. No final test denominator, HTTP close-summary or performance. |

Two predeclared full traced populations did **not** both complete. An untruncated
capture of an incomplete invocation is not a passing or complete measurement.
No retry-to-green, flakiness closure or hosted/whole-CI claim follows.

## Retained evidence and source

[Structured JSONL artifact](browser-module-observations.v1.jsonl) retains every
captured HTTP record, native observation and provider request-failure line for
82623, 67300, 93957 and 3595, plus exact summaries and capture metadata. It is not
the full passing-test transcript. Source HEAD at measurement: `e6ea05bf`; these
two instrument files were then uncommitted, identified by actual SHA-256:

- browser config: `ada11b645b62ab03ebb539ec4bd49b05ff2625435f18a19f00461f6d310481a0`;
- browser witness: `a585474a5e441ed4030776b08ba686b2ee26100b1d30cc0739376b77cf5f6466`.

Both intentional HTTP interceptions were temporary and restored to those exact
hashes before restored populations. No production byte changed. The native
helper forwards original postMessage/terminate arguments and never prevents an
error, retries a message or suppresses the existing uncaught-error guard.

HTTP observation covers **only** Route JSON and prediction-worker module paths,
not all imports or Garden raw-catalog dependencies. HTTP IDs belong to each
server instance; worker IDs belong to each test observer. Multiple truthful
zero-count server summaries are distinct from the actual browser server.
HTTP 304 and teardown request cancellations occur in passing populations too;
neither is by itself a defect or its cause. A 404 control proves observation of
that fault, not that an original failure was a 404.

## Storage finding (RP-236)

Read-only diagnosis of owned container `2bb1518c1355`, before termination:

- Firefox requests Vitest runner index.js and reports
  `NS_ERROR_FILE_NO_DEVICE_SPACE`.
- Container `/` overlay: 126 GB total, 120 GB used, zero available, **100%**.
- `/dev/shm`: 7.8 GB available, zero used; this is **not measured shared-memory exhaustion**.
- Host-mounted workspace: 197 GB available, 79% used.
- Docker inventory: 342 images (78.3 GB), 145 volumes (39.58 GB), 2.634 GB build
  cache; substantial reclaimable storage exists, but its ownership was not
  established. No broad prune, unrelated volume/image deletion or user-service
  stop was performed.

This is concrete environmental invalidity for the stalled invocation. It does
not retroactively establish RP-235 or RP-218's causes, nor a scanner loop or
late-disposed-worker error. Garden scanner source is unchanged; lack of a file
completion is not a test-body failure. Whole invocation remains incomplete.

## Next authorized work

Resolve Docker disk capacity by a precisely owned-cache cleanup or an owner
choice; never use a broad prune. Before another expensive browser invocation,
predeclare a capacity/containment observation and retain its failing case under
CI D2. Existing ten-minute hosted containment is not currently a local command
timeout. Do not expand budgets or remove a test to conceal the hang.

Meanwhile, accepted Account/API contracts permit an exact refresh-response
census without implementing the unaccepted browser-renewal draft. R-011 browser
primitive success remains separate from real refresh ambiguity, owner policy,
native fifteen-minute Garden, full release proof and designated review.

Review by: Codex (first-filter only). Recorded by: Codex. Complete new range
begins `6ed42d45` exclusive and includes all three predeclarations, instrument,
artifact, reconciliation and following pin. Claude's cross-party verdict remains
required; no approval, checkbox promotion, archival, push or publication occurs.
