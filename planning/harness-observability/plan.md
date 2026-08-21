# Harness Observability implementation plan

Status: implementing. Authority checkpoint: `9c71562`.

## Scope firewall

- [x] No catalog, scenario, policy, seed, horizon, milestone, golden, baseline, epoch, work budget,
  timeout, worker, dispatch, parallelism, sharding, algorithm, gameplay, deployment, or CI topology
  change appears in the implementation range.
- [x] Observation remains explicitly non-authoritative and cannot satisfy a harness acceptance gate.

## Implementation

- [x] Add the canonical `harness_observation.v1` model, atomic checkpoint writer, strict validator,
  elapsed-time source, and incomplete/interrupted interpretation.
- [x] Add CLI observation output and controlled signal/error finalization without changing ordinary
  no-observation exits or domain report bytes.
- [x] Instrument repository guards, standard pacing, and each registered relevance row with exact
  loaded identity, start/completion, known work counts, and explicit unknown/fired states.
- [x] Add a registry-aware single-row mode that selects only a unique registered scenario through
  `LoadRelevanceRegistry` + `LoadRegisteredRelevanceSuite` and validates the same row evidence as
  complete check.
- [x] Expose a root Make measurement target with explicit observation path/selector inputs; do not
  add it to CI or a release gate in this RFC.

## Discriminating acceptance

- [x] Cold affected Go tests prove complete artifact acceptance and reject missing/running/error/
  signal, fired guard, truncation, exclusion, identity mismatch, and cardinality mismatch cases.
- [x] A CLI subprocess `SIGTERM` leaves parseable incomplete evidence and exits nonzero.
- [x] A completion-severing negative fixture retains successful domain work but the observation
  validator fails.
- [x] An active constants-binding mutation makes the registry-aware selector fail.
- [x] Observation-on and observation-off domain outputs are byte-identical for a bounded governed
  fixture; existing tracked harness artifacts remain byte-identical.
- [ ] Cold `make test-go GO_PACKAGES='./harness ./cmd/balance-harness' GO_TEST_FLAGS='-count=1'`,
  `go vet` through the root Make target, schema validation, and the relevant existing harness gates
  pass. Long complete measurement is recorded separately and never called green if interrupted.

## Closeout

- [ ] Update `docs/balance-harness.md`, R-001 diagnosis/research/queue/backlog records, this log, and
  the RFC/index without changing D-014.
- [ ] Commit a small exact review range, run an internal diff/protocol check, and hand only that
  range to Claude for the mandatory cross-party designated verdict.
- [ ] Archive only after that exact range is approved and all acceptance criteria actually pass.
