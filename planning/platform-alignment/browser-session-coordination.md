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

## Real Account/browser follow-up — 2026-10-12

Predeclared at `e497cae2` in the Account plan. Reproduce from the repository root
with `make diagnose-browser-session-coordination`. The existing Garden driver
builds the actual gameserver and initializes its declared disposable Postgres;
the diagnostic uses native Chromium, fresh contexts/accounts and a test-owned
loopback proxy. It does not run the production client or implement renewal.
Final helper SHA256 `e8a93b861e3ea71a660e431f9ebaa9cc5f8dc72fbb4fda6fc9b1fb1023a0fc71`;
Garden driver SHA256 `952a18291fb069e4588886c85d7802358747d420e477aae981fbb697809ac39e`.

Run10648 exits0 with all three cases, no exhausted guards or uncaught errors.
Each first refresh is an actual200 from Account, with the Postgres rotation
committed and the complete reply held before delivery to the owner page:

| Arm | Refresh requests / upstream statuses | Credential-family counts | Observed result |
| --- | --- | --- | --- |
| Healthy native queued overlap | 1 /200 | `[2,1,0,2,0,0]` | Both tabs retain the same new pair; authenticated read returns the same Founder. |
| Owner page closes after commit, before delivery | 1 /200 | `[2,1,0,2,0,0]` | Queued and reloaded successors stop on the retained marker; original stored pair remains; family is not revoked. |
| Identical closure without marker | 2 /200,401 | `[2,1,2,2,2,1]` | Successor receives exact `refresh_reused/session_family_revoked`; both old and held-descendant access are refused. |

Count order: sessions, consumed sessions, revoked sessions, access tokens,
revoked access tokens, revoked families. Each account starts `[1,0,0,1,0,0]`.
All ten Account/game tables listed by the instrument compare byte-exact privately
before/after refresh; no rows change. Tokens, recovery codes, reply bodies and
private snapshots are not serialized into the report. Thirty-second guards are
invalidity bounds, not new product deadlines.

Initial restricted attempt6579 cannot reach Docker (no cases). Permitted1957
fails bootstrap before any case because the instrument supplied a UUID-shaped
idempotency key instead of the existing64-lowercase-hex contract. The correction
matches the actual client's32-byte key construction; no production parser or
assertion changed. The final run above is the corrected instrument.

The existing `make test-garden-harvest-composed`9784 also exits0: the unchanged
normal branch exercises built client, actual server/Postgres/WebSocket, purchase,
plant/uproot, controlled maturity, single/all harvest, receipt retry/refusal and
reload. This checks the driver integration, not natural fifteen-minute growth.
Node syntax and unchanged CI topology plus its13 negative controls pass.

This establishes three bounded Chromium/real-Account committed-reply cases,
including a real revocation failure control. The interrupted arm's old access
still works before expiry; this is not successful recovery after expiry. The
test credential document/coordinator is not the shipped runtime. Supported-engine
completion, real credential persistence failures, background/process/OS lifecycle,
natural expiry and automatic reconnect remain open. The candidate marker and
renewal policy are not adopted. No feature acceptance or hosted-CI claim follows.
