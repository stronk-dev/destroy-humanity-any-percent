# CI Baseline — Implementation Plan

- **RFC:** `rfc/scaffolding-and-ci.md`
- **Assignee:** Codex
- **Started:** 2026-07-28
- **Status:** fast/maintenance split implemented locally; hosted acceptance pending

## Current checkpoint — 2026-10-13

RP-525's exact Node-only fixture exclusion is locally verified in its actual Node/browser
consumers and awaits designated review. The complete job audit still fails: native client
hits RP-131; Linux-amd64/Postgres server hits RP-307; full browser input/runner evidence is
RP-526. Its proposed single-file config was withdrawn after the ordinary target stalled,
despite an earlier successful CLI control. Do not claim browser or whole CI green.

The local container now enforces the existing ten-minute ceiling and passively observes
the four unreported module URLs. The subsequent ordinary job completes in316.27s but still
has15 native-input failures; those modules fetch successfully and pass123 isolated cases.
This diagnoses a boundary, not the original stall or input defects. Next: investigate
post-timeout native-input continuation and make the full job complete repeatably. Docker
subsequently becomes unresponsive with only1.8GiB host space free; compiler-cache cleanup
approval is requested, not assumed. The owning log preserves commands, capacity and outcomes.
RP-527 now reproduces and locally corrects Typer's post-timeout input-producer lifetime
in native Chromium/WebKit, including the unchanged complete native journey. The CI log owns
the controlled old failure and corrected passes; this does not close RP-526 or prove Linux/
Firefox/full-job reliability. Resume that population when Docker/capacity is available;
cleanup approval remains pending. Historical guard/accrual contracts, hosted acceptance and
reviews remain separate; no deadline/budget increase or CPU/concurrency setting retained.

RP-528 also locally corrects the Garden host's premature socket precondition and initialization
cleanup in native Chromium/WebKit. The original full file still executes its handshake, recovery,
intent and focus population; an intermediate shadowed-context candidate is excluded by strict
types. This does not close RP-526. The owning log records exact results and next Linux verification.

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
