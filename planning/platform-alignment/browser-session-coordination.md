# R-011 — browser coordination primitive observation

Observed 2026-10-06. Bounded research, NOT automatic renewal, accepted policy,
a release gate or completion of R-011.

## Predeclaration and reproduction

Research queue R-011 and Account log at `d023dc26`: four arms times ten
repetitions per declared engine, plus ten holder-page closures. Run root
`make research-browser-coordination`: existing Playwright 1.62.0 Linux Compose
service, test-owned loopback origin, synthetic values. No game server, credential,
Account call or replacement session manager. Manual target outside CI/verify.

The [artifact](browser-coordination-observation.v1.json) pins predeclaration HEAD
`d023dc265fda75549b47b13027eff55f4776f898` and actual then-uncommitted script SHA
`9dbad4855302ab7cefee97b79c51a0753b5fb4515987eb1094dbff363034a583`.
Start `2026-10-06T03:17:12.907Z`; not a claim the script already existed at HEAD.
Linux arm64 / 6.10.14-linuxkit; actual Chromium 151.0.7922.34, Firefox 153.0,
WebKit 26.5. Actual UA/secure-context/native API fields are retained.

## Direct observations [V]

| Population | Per engine | Observed result |
| --- | ---: | --- |
| Same-context, same-name exclusive Web Lock | 10 | Peak one; native pending contender then completion after release |
| No-lock control | 10 | Peak two overlapping holders |
| Different-name native locks | 10 | Peak two overlapping holders |
| Same name in isolated contexts | 10 | Peak two; separate storage buckets |
| Actual holder page closure | 10 | Successor excluded before close, acquired after close start, completed; owner call interrupted |

All three complete **120 interval cases / 30 page terminations**. Same-context
synthetic storage is shared, isolated context sees no value, replacement/reload
sees next generation. No exclusions, exhausted guards or uncaught browser errors.
Separate Node check recomputes cardinality from every ordered entry/exit trace,
exact per-arm repetitions and termination ordering; actual script SHA matches.

## Failure controls and attempts [V]

- Initial script `23b9fe24…`, run 48229 completes all. Exclusive bypass 6764
  fails first Chromium case with both entered; original script restored exactly.
- Refinement adds time/source/hash to failure output and stops polling on watchdog
  exit; no population/threshold/oracle weakened. Full 35833 completes. Bypass
  20140 fails `exclusive lock admitted overlapping holders`, SHA `e086016d…`.
- Forced one-millisecond observation watchdog 1306 fails with
  `guard_exhausted:true`, zero completions, `guard fired before first holder
  acquisition`; SHA `1baf9a8c…`. Terminates, not a partial pass.
- Final exact SHA restored; full **23130** completes. Retained artifact is this
  final run. Failed invocations do not overwrite an earlier completed artifact:
  compare invocation status/source/instrument identity, not file existence alone.

Normal 30-second watchdog is an observation invalidity guard, NOT a measured
production deadline. Synthetic timings cannot authorize a product budget.

## Limits and next evidence

These headless Linux profiles support further investigation of Web Locks, not
adoption. No full browser-process crash, OS suspension/background, mobile/native
host, storage-denial, real HTTP rotation/lost-commit reply, durable marker safety,
unsupported fallback, natural fifteen-minute Garden or player recovery is proven.
No coordinator/marker/credential/auth policy/copy ships. Later real-Account arms
need separate predeclaration; API/owner/recovery dependencies remain unaccepted.

Root checks 9136: 7366 pass / 134 browser-only Node skips; zero TS/Svelte
diagnostics; boundary counts 14/8/22; all 13 topology controls rejected.
Node syntax and artifact trace checks pass.

**Full unchanged browser CI 54700 is RED:** Firefox fails dynamic import of
`balance/routes-testdata/invalid/dangling-resource.json?import`. 299/300
populations pass, 22466 tests / six existing skips, 51.98 seconds. Sixteen Route
tests never import; performance command not reached. RP-235 is not a skip,
proved root cause or hosted green. Historical RP-131/worker/native caveats remain.

Codex implementer first-filter only. New span begins after `39f95329`, includes
`d023dc26`, instrument/artifact/dossier and records; literal tip pinned after
commit. Claude's cross-party verdict required; no prior span, owner decision,
checkbox or archive consumed.

## Native interrupted-dispatch follow-up — 2026-10-12

Question/population predeclared at baseline `a6fc7217` in the Account plan. Reproduce
with `make research-browser-refresh-dispatch RESEARCH_BROWSER_ENGINES='chromium webkit'`.
The default names all three engines; explicit subsets report unexecuted engines and
cannot claim full supported-browser completion. The instrument is test-only and
prints its complete selected-case report rather than replacing the historical artifact.
Final source SHA256 `86b1b77e32128cbf517c6e53b57f73fb305a786f5dc2ae438d3b1c29fa52abd8`;
source HEAD identifies the pre-experiment baseline, not a committed instrument at that SHA.

Native macOS27.0/arm64 Chromium151.0.7922.34 and WebKit26.5 execute five repetitions
of each arm,40 total, with no exhausted guards or uncaught page errors:

| Arm | Actual boundary/outcome per case |
| --- | --- |
| Healthy overlap | Native queued lock; one HTTP dispatch, generation2 stored, successor rereads it without sending. |
| Owner closes after dispatch, before reply | Actual interrupted page; marker survives. Queued and reloaded successors stop with generation1 intact; one HTTP dispatch. |
| Same closure without marker | Successor sends again; two HTTP dispatches and the synthetic receiver's reuse refusal. This is the failure control, not successful renewal. |
| Native quota denial | Actual `QuotaExceededError`; marker cannot be stored and no HTTP request is sent. Original generation remains. |

The final run13295 exits0. Earlier10160 passes the same40 before tightening launch
cancellation/reporting. Initial restricted run cannot listen (zero cases, invalid).
Separate Firefox19452 exits2 after its30-second launch guard: sandbox-extension/SWGL
failure, zero of20 cases. No Firefox or full60-case result is credited.

This bridges native exclusion/storage to a real HTTP dispatch and page-loss boundary.
The HTTP receiver and generations are synthetic; real Account rotation/revocation,
post-reply credential storage failure, permissions denial, identity replacement,
browser-process/OS crash and disk durability remain separate. Quota-fill limits and
observation guards are invalidity bounds, not production capacity/deadline choices.
No marker schema/lifetime, coordinator, recovery copy or automatic retry is adopted.
The Account log owns exact review/procedure evidence; R-011 and renewal stay open.
