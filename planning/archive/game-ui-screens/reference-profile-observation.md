# Game UI AC5 manual reference-profile observation

Predeclared 2026-08-22 before measurement. The owner's standing direction to continue after the
designated Game UI approval is treated as delegation to execute this local, non-product release
check. It does not authorize a budget change, product optimization, CI change, push, or archive.

## Question and population

Does the mounted production Game UI hold its ruled performance budget for 60 real seconds in
pinned Chromium at 1280×720 under Chrome DevTools Protocol 4× CPU throttling?

The population is one authenticated Desk over the real Vite proxy, gameserver, Postgres, WebSocket,
production prediction worker, and shared Amount renderer. Ordinary server-side setup gives the
Company one producing Beige Tower so formatter observations are non-vacuous. No fixture clock,
manual scheduler flush, or simulated-duration shortcut is permitted.

## Measurement

1. Record host OS/architecture, CPU model, Node version, Playwright Chromium version, commit, and
   wall-clock start/end.
2. At 1× CPU, sample `requestAnimationFrame` for five seconds. The observation is invalid unless
   at least 120 intervals exist and their median is between 8 and 25 ms.
3. Apply `Emulation.setCPUThrottlingRate({rate: 4})` through a page-owned CDP session and confirm
   the requested rate in the result artifact.
4. For at least 60,000 monotonic milliseconds, observe:
   - actual `predicted_snapshot` messages from the production prediction worker;
   - character/child mutations of the visible cash Amount output;
   - all Long Tasks through `PerformanceObserver`;
   - every animation-frame timestamp.
5. For each animation-frame gap, estimate missed frames as
   `max(0, round(gap / baseline_median) - 1)`. Dropped-frame PPM is missed frames divided by
   observed-plus-missed frames, multiplied by 1,000,000.

## Predeclared gates

- viewport is exactly 1280×720;
- CDP throttle rate requested is exactly 4;
- measured duration is at least 60,000 ms;
- prediction-worker and Amount-mutation counts are present and nonzero;
- formatted Amount mutations are at most 600;
- Long Task instrumentation is supported and the longest task is at most 200 ms;
- dropped frames are at most 50,000 PPM (5%).

The existing deterministic CI lane continues to own its exact 1,200 injected inputs and 600
formatter windows. This manual observation supplements it with real wall time, actual worker
outputs, applied throttling, and frame delivery; it cannot loosen or replace the CI assertions.

Any missing instrument, guard exhaustion, page error, server/socket failure, out-of-range field,
or threshold violation makes the observation invalid or failed. A failed result is recorded as
completed evidence and blocks archival; thresholds are not adjusted after execution.

## Result

**PASS — 2026-08-22.** Measured from committed `0d53fd1`; the observation probe was temporary,
fully removed afterward, and made no product or test-harness change.

Environment:

- host: MacBook Pro `Mac15,10`, Apple M3 Max (14 cores), 36 GB, macOS 26.5 arm64;
- Node 26.7.0, Playwright 1.62.0, Chromium 151.0.7922.34;
- wall-clock valid-run window: 2026-08-22 11:12:33Z–11:13:55Z;
- real Vite/gameserver/Postgres/WebSocket composition; 100 producing Beige Towers;
- viewport 1280×720; requested CDP CPU throttle 4×.

Observed artifact:

| Field | Result | Gate |
|---|---:|---:|
| Baseline frame intervals | 601 | ≥120 |
| Baseline median frame interval | 8.3 ms | 8–25 ms |
| Measured duration | 60,001.9 ms | ≥60,000 ms |
| Published prediction-worker snapshots | 598 | nonzero |
| Visible Amount mutations | 355 | 1–600 |
| Long Task instrumentation | available | required |
| Longest Long Task | 0 ms | ≤200 ms |
| Observed frame intervals | 7,200 | nonzero |
| Estimated missed frames | 0 | reported |
| Dropped-frame rate | 0 PPM (0%) | ≤50,000 PPM (5%) |

The worker publishes predicted snapshots on the 100 ms snapshot cadence; its internal 50 ms
simulation steps remain covered by the deterministic lane's exact 1,200-input proof. This manual
result neither changes nor substitutes that bound.

One earlier attempt was invalid before the measurement window: after the profile reload it waited
for visitor Copy to re-emit and timed out. The socket itself was open. The rerun replaced that
presentation proxy with the direct open-WebSocket condition; population, instruments, algorithm,
thresholds, and product bytes were unchanged. No measurement from the invalid attempt is used.
