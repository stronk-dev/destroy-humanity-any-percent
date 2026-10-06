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
