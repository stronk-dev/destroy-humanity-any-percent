# R-010 — native worker and refresh evidence

## Authority and baseline

Predeclaration `f4f62eac`, baseline `b299bbf5`. Accepted CI Baseline permits test evidence,
not changes to the archived shell. RP-218's original full-lane failure remains unexplained.
No production, clock, five-second witness bound, browser population or CI configuration changes.

## Executed initial population — 2026-10-06

The native `Worker.prototype.postMessage` observer forwards every original argument; actual
Worker message events supply predictions. No fake Worker or synthetic clock is used. The
controlled arms temporarily supply `document.visibilityState = visible`, dispatch the existing
visibility-return event every 250 ms and restore the descriptor, listener and native method.

- Cold isolated Linux Chromium/Firefox/WebKit: **9 pass**, 60 selector exclusions. Original
  witness succeeds in all three. Conflicting authority produces 15–16 real predictions yet text
  stays `100`; consistent authority produces 16 predictions and text `200`/`300`. Each controlled
  arm records 9–10 actual snapshot calls. No offline-required output observed.
- First complete cold `make test-browser-ci`: **22,068 pass / six intentional skips**, 276 file
  populations; 44.07 s. Separate Chromium performance **1 pass**, 2.05 s total invocation.
  Original witnesses observe native predictions, no refresh after injected fast authority and
  no offline-required output. The fixture confounder is not exercised by those original passes.
- Second complete cold lane: **22,067 pass / one fail / six intentional skips**, 41.64 s,
  exit 2. The new **WebKit conflicting-control** assertion sees zero native predictions while
  nine snapshot calls and low-rate replacements are delivered. The control is invalid under
  the predeclaration; it does not prove a healthy worker or a production defect. Separate
  performance is not reached. The original witness passes in all three, again with native
  prediction and no replacing refresh/offline output.
- `make typecheck`: zero errors/warnings. An initial invented `make check-client` alias fails
  immediately (no such target); only the actual successful `typecheck` is evidence.

## Findings and limits

Repeated low-rate authority can deterministically keep the DOM at `100` despite real worker
prediction. This is a demonstrated **fixture confounder**, not the cause of the earlier five-
second failure. Two original full-lane repeat populations are green, but one diagnostic population
is red. There is no general full-browser/hosted-green claim and no statistical flakiness closure.

The two-second control window mixed Worker startup with steady refresh observation. Its missing-
prediction arm fired the intended validity gate. Fix the instrument before remeasuring; do not
count that arm or retrospectively extend its window. A subsequent study must record startup and
bound readiness separately, then retain the same two-second steady-state population.

## Follow-up predeclaration — test-only correction

Baseline initial diagnostic `0a30f0d0`. Keep all commands native and all three browser engines.
The original witness will give the runtime and injected UI the **same** fast-growth snapshot;
assert actual native prediction as well as changed text within its unchanged five-second poll.
Record user agent, native first-prediction latency and low/high-rate input to make later failures
actionable. This removes the demonstrated confounder, not a claimed root cause.

Before each controlled two-second refresh arm, require the first actual native prediction within
five seconds; retain that startup latency explicitly. Then inject the fast snapshot, begin
visible-return refreshes and count **only new** native outputs during the original two-second
measurement. A startup timeout or absent steady-state output is invalid/red, never excluded.
Do not increase an assertion budget, retry until green, serialize/exclude jobs, or alter runtime.

Population: the three selected cases in all three Linux engines, two complete cold browser lanes
and separate performance invocations, root client/type/build checks. Deliberately suppress native
prediction publication once: original witness and both readiness controls must fail. Restore the
exact production file before final verification and record its hash. Test-only implementation and
diagnostic ranges require Claude's designated review. The original RP-218 cause remains open
unless a later original-failure trace demonstrates it; existing RP-131 still forbids whole-CI green.
