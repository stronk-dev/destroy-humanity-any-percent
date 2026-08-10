# Client Shell and Simulation Loop

The browser client has a Svelte 5 presentation controller and a dedicated prediction Worker. The
runtime boundary remains available to future screens, while the production entry currently mounts
the content-free [UI Foundation](ui-foundation.md) fixture. Game UI owns the later composition of
that runtime with real screen surfaces and transport.

## Runtime boundary

`ShellRuntime` composes three independent pieces:

1. a `SnapshotStream` that supplies authoritative snapshots and intent receipts;
2. `PredictionWorkerClient`, which owns the dedicated module Worker; and
3. `ShellController`, which produces immutable presentation views for Svelte.

The adapter rejects unknown fields, non-canonical Decimal strings, invalid mechanical IDs,
unexplained caps, duplicate progress stages, malformed hashes and receipts, and unsafe revisions.
Snapshots older than the visible revision are ignored. Receipt and snapshot revisions must match.
The transport wire envelope and reconnect channel remain owned by the active transport RFC.

The runtime adapters validate the economy and client-shell catalogs before use. No resource,
generator, progress-stage, or cap ID is compiled into a UI primitive.

## Prediction

The Worker owns a monotonic 50-ms fixed-step accumulator and emits presentation snapshots every
100 ms. It anchors every resource to the last authoritative amount and evaluates total elapsed
milliseconds through the shared `accrueConstant` production primitive. Prediction therefore does
not change when the same elapsed time arrives as one pulse or several pulses. Declared hardcaps are
respected locally, but only the server commits state.

A clock gap above 5,000 ms performs no local catch-up and requests the authoritative offline path.
The same request occurs on visibility return. The lifecycle layer measures the actual hidden
interval independently of Worker throttling, so a Worker that continued ticking cannot bypass the
return-story rule.

Prediction is presentation only. `IntentDispatcher` accepts the closed ten-intent union currently
implemented by the production engine, validates exact fields at runtime, and always forwards the
intent to its authoritative adapter. It accepts no predicted balance or affordability result. A
source-boundary check also forbids prediction imports in the dispatcher and balance-mutation
imports anywhere in the shell.

## Reconciliation and display

The policy catalog currently sets the following reviewed values:

- divergence below 10,000 ppm (1%) bends toward authority over at most 400 ms;
- larger continuous divergence rebases immediately;
- discrete facts always snap;
- rejected receipts rebase and expose their typed rejection code;
- reduced-motion presentation samples at 500 ms and suppresses interpolation and pulses.

Counters retain unquantized presentation values. A producing resource increments a separate
activity indicator even when its visible digits cannot change at the current magnitude. A counter
at a declared cap exposes the cap's `reason_key`. Typed progress coordinates remain server data;
the shell renders the supplied current and target strings and derives only the visual fill ratio.
Formatting and Svelte view refresh are throttled to 100 ms.

Client telemetry is aggregate-only: epsilon breaches, discrete rebases, rejection categories,
Worker overrun count, and total overrun milliseconds. It retains no intent IDs or player data.

## Controller routes and lifecycle

The controller retains contract, main-panel, and run-end states for the future Game UI consumer.
The RTA clock begins only after the injected begin-attempt action succeeds. PB/WR comparison remains
unavailable until an Exit. The old hard-coded screen scaffold is no longer the production entry;
surface routing now belongs to the UI Foundation registry.

For an absence over 30 seconds, the authoritative snapshot loads behind one return recap instead
of visibly snapping the counters. The recap lasts at most five seconds and is skippable, gains stay
docked in the header, and the shell exposes one optional follow-up-modal flag without inventing
Fiscal Quarter mechanics. Other notifications remain badges owned by later feature surfaces.

`visibilitychange` measures the hidden gap and requests authority, `pagehide` flushes the stream,
and `freeze` disposes Worker and stream resources. There is no `unload` handler.

## Verification

Run:

```sh
make typecheck
make build-client
make test-client
make test-browser
make verify-client-boundary
make verify-schema
```

The unit suite pins controller state, continuous/discrete reconciliation, typed rejection handling,
and return behavior. The browser suite now exercises the content-free UI primitive fixture in
Chromium, Firefox, and WebKit. The production build retains the prediction Worker implementation
for the later Game UI composition.
