# CI Baseline — Implementation Plan

- **RFC:** `rfc/scaffolding-and-ci.md`
- **Assignee:** Codex
- **Started:** 2026-07-28
- **Status:** fast/maintenance split implemented locally; hosted acceptance pending

## Current checkpoint — 2026-10-14

RP-596 standard-phase measurement declared at clean `0411ef34` and now completed: all300 keys,
12 workers, native macOS/ARM64/M3 Max,783.821s wall time. Full population/outcome/invariant checks
pass. The manual-only `BenchmarkStandardPacingProfile` loads the same suite and calls the
same `RunAllWithWorkers` as the CLI; it refuses missing/repeated/failed keys and aggregate
failures, reporting source identities and canonical report digest. Run with `-benchtime=1x`.
No population/algorithm/observer/CI/budget changes. The measured phase excludes CLI guards and
registered relevance and cannot be labelled a complete check. CPU samples include process
setup; the benchmark's wall metric brackets only RunAll. No fixed improvement threshold or
optimization adoption is inferred. Compare hotspots with CLI observation placement, not native
wall time as proof of hosted timing. Raw profile/binary stay in an owned temporary directory;
the owning log retains exact commands, identities, report/profile/binary digests and limits.
The executed profile attributes39.05% cumulative sampled CPU to production rate calculation;
source attribution points to its repeated contribution-map iteration. Runtime memory-management
samples are also substantial. These are overlapping sampled stacks, not a predicted wall-time
saving or proved hosted cause. The root manual `harness-profile-report` reader executes and
rejects missing/malformed inputs; ordinary16 topology negatives and dispatch controls still pass.
Next: scope a separately authorized rate-traversal performance proposal, preserving every
multiplier's canonical order, validation and activation semantics; retain whole300-run report
identity/acceptance and representative negative/runtime parity evidence. Do not change production
under measurement authority. Linux/hosted attribution and full maintenance remain open.

Fresh hosted reconciliation at local product source `8a5dcebf`: remote CI remains38381138075
on `4eaed0c9`, failed;75 local commits are outside it. Maintenance38507071157 also completes
failure. Numeric maintenance passes; Reputation career/relevance report comparisons drift.
RP-596's complete harness observation exits124 at the unchanged50-minute limit: standard
pacing completes300/300 in45.19 minutes, first relevance row stops at85/107 and second never
starts. The small exact artifact is retained in `observations/38507071157-harness.json`; the
actual root observation validator rejects it. Log owns hash, timings and commands.

The native profile above now completes the standard-pacing measurement step. CLI inspection
confirms only start/finish recorder calls bracket RunAll, not per-action writes; no quantitative
observer-overhead timing comparison was performed. This
incomplete full run cannot set a larger budget or authorize dropping rows, retries, golden
regeneration, CI changes or optimization under the archived measurement-only RFC. Linux/SQL
execution remains unavailable with the unresponsive declared Docker service. Latest host read
has71GiB available/92% used; no cleanup/restart or publication authority is inferred.

At baseline `f6ab6ab6`, the full functional target now runs broad collection then the complete
Typer native-input file in fresh browser processes. The default Chromium/Firefox/WebKit
population, inputs, assertions, deadlines and workers are unchanged; focused selection remains
one invocation. Isolated performance remains after functional success. Separate JSON reports
prevent phase overwrite; a first-phase failure leaves later phases unexecuted.

Executed native Chromium/WebKit gate: broad93.93s plus native2.40s, all24,456 case/title/skip
tuples identical to the preceding full run.24,430 ordinary passes, ten expected deadlines,
16 skips, no unexpected failures; both normal Typer journeys finish178/191ms. The unchanged
isolated Chromium performance target passes five cases in5.55s (105 selector skips).
Topology rejects16 invalid states; actual Make expansion verifies focused selection,
two-phase/report separation and performance ordering. The log owns raw paths/digests.

This is a locally verified execution-organization change, not proof the intermittent RP-526
failure is eliminated or the hosted five-minute gate is met. Firefox/Linux/SQL/current hosted
remain unverified. The preceding red full run stays evidence; no test source, product, schema,
kernel, balance, workflow job, timeout or acceptance criterion changes. Complete range after
`f6ab6ab6` through the containing commit needs designated Claude review, consolidatable
with existing CI diagnostics. Next: execute the unchanged Linux three-engine job when its
declared Docker service is usable; then reconcile actual hosted results when publication is
authorized. Continue other accepted implementation meanwhile; do not repeatedly rerun native
green, loosen deadlines or treat review/owner holds as a global development stop.

Earlier completed diagnostic/extraction/concurrency checkpoints and failed controls remain
in the append-only owning log and their named evidence capsules, not stale next-action plans.

Published [CI38381138075](https://github.com/stronk-dev/destroy-humanity-any-percent/actions/runs/38381138075)
at `4eaed0c9` is completed **failure**: browser/composed/schema/harness pass;
client fails the history guard at `50a3a514` (RP-131); server fails 27 Clout
partition cases (RP-307). Both reproduce locally at `61b145af` through existing
root selectors. The later local hoard fix is not in this run.

Hosted three-engine browser coverage executes 34,704 passes, twelve deliberately
expected timeout controls and 24 skips; five Chromium performance cases pass.
Functional duration is 305.18s and total browser job 6m38s; server is 6m40s.
The five-minute blocking floor remains failed. One complete run does not prove
reliability, attribute every earlier stall/input failure or provide designated
approval. RP-525/527/528/530/531 corrective ranges still need review; RP-526
retains the ordinary local failure and investigation in the [owning log](log.md).

Independent holds: obtain explicit adoption/review of the history-guard policy, close the
conservation accounting contract and repair its arithmetic without tolerances,
and measure latency before proposing changes. Preserve all functional coverage
and budgets. Host space is externally freed, but Docker restart permission is
pending; no agent cleanup/restart or CI topology change. Full workflow, maintenance,
feature and release acceptance remain open. Historical details stay in the log,
not a duplicated stale execution queue.

## Work

1. [x] Split the Makefile into server, client, browser, schema, and aggregate verification targets.
2. [x] Add a pinned Ajv 2020-12 schema verifier plus positive and negative catalog fixtures.
3. [x] Add the least-privilege hosted workflow with frozen installs, dependency caches, pinned
   browser installs, and superseded-run cancellation.
4. [x] Exercise every narrow Make target and the aggregate local gate.
5. [x] Update README and canonical CI documentation.
6. [ ] Observe the first hosted workflow under five minutes, then archive the completed RFC and
   planning record.
7. [x] Under D-014, split push/PR harness smoke from scheduled/manual exhaustive harness and
   numeric evidence; enforce the topology with seeded failures.
8. [ ] Observe a green current-head push/PR workflow and one complete scheduled/manual harness
   observation before claiming hosted closeout.

## Acceptance gates

- Clean, frozen dependency installation succeeds.
- Server, client, browser, and schema targets pass independently.
- A negative schema fixture is rejected while the schema command succeeds by observing that
  expected rejection.
- `make verify` passes, including all three browser engines.
- Workflow structure has read-only permissions, push/PR triggers, branch-scoped cancellation,
  supported action majors, browser/package version parity, and no browser-binary cache.
- Blocking elapsed time is measured on GitHub after the amended workflow is pushed; the RFC cannot
  claim the hosted-runner five-minute gate from a local machine.
- Scheduled/manual exhaustive execution must upload a complete validated observation, or an
  explicit incomplete artifact when the 50-minute command ceiling fires.
