# CI Baseline — Implementation Plan

- **RFC:** `rfc/scaffolding-and-ci.md`
- **Assignee:** Codex
- **Started:** 2026-07-28
- **Status:** fast/maintenance split implemented locally; hosted acceptance pending

## Current checkpoint — 2026-10-14

New native recurrence, 2026-10-14 at `213ad5f6`: complete Chromium/WebKit functional execution
fails the unchanged Typer15s deadline in `submit-prompt-4`, with a focused connected input,
pending false and native delivery in flight. The capture records only1.04s in that current
stage, so earlier-stage costs need the existing observer's ordered timing history, now verified
with real framework-deadline controls. This is
new failure localization, not a proved player bug or an input/deadline/concurrency correction.
The owning log retains the complete red result, raw artifact digest and diagnostic verification.

Preceding checkpoint (2026-10-13): RP-553 closes a diagnostic gap: Typer's phase observer arms before the first journey await,
not after content/engine setup. A real pre-initialization deadline in the same journey fails
the old placement in both native projects, then passes after the correction. Full final native
Chromium/WebKit functional population passes23,706 with ten expected deadlines/16 skips; this
is not cause attribution or an RP-526 repair. The log owns exact results/range and raw-output
truncation limits. Keep input, ordinary15s bound, concurrency, data and assertions unchanged.

Current native cross-surface verification retains an RP-526 recurrence: at `b8eca33c`,
the complete Chromium/WebKit functional population fails Typer's unchanged15s whole-game
deadline. A new failure-only phase observer is locally verified, including a real deadline
sampling DOM state before teardown. One instrumented complete population passes afterward;
that is not attribution or repair of the original failure. The log owns both outcomes,
source/range and limits. Preserve the observer on recurrence; do not extend the timeout,
substitute synthetic input, rerun for green or credit native results as hosted/Linux proof.

RP-540's citation boundary is locally corrected: old genuine-new-row CLI controls
admit four aliases, and a controlled exterior-review read is reproduced without
leaving the owned temporary fixture. Canonical active/archive citations and all
existing guard populations pass after entry-level refusal; the real root still
fails `50a3a514`. The log owns verification and pending consolidated review.
Do not repeat this completed population or credit it as authentication/whole CI.

RP-131's non-verdict parser loophole is locally corrected independently of the
unadopted historical-exception proposal. Actual CLI regressions retain the old
eight admissions; twelve invalid evidence forms now refuse while completed
findings/remap records pass. The existing `09d5027` binding is transcribed from
Claude's completed September24 enclosing-range verdict, not a new review or
approval. The real root gate still fails `50a3a514`; topology/fixtures pass.
The owning log records the exact pending designated ranges. The citation follow-up
is complete locally as above; continue genuine accepted work and policy closure,
not more repetitions of these completed guard populations.

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
