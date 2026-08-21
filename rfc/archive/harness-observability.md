# RFC: Harness Observability

- **Status:** implemented
- **Author:** Marco (scope ruled by Marco; drafted by Codex)
- **Created:** 2026-08-21
- **Design refs:** `design/00-vision.md` §Design pillars 7; `design/06-tech.md` §Balance & event data;
  `design/12-content-pipeline.md` §7
- **Depends on:** Balance Harness Foundation (implemented), Relevance Harness (implemented),
  CI Baseline (implementing)
- **Parent / amends:** `archive/balance-harness-foundation.md` D7, observability only
- **Supersedes / superseded by:** —
- **Planning:** `planning/archive/harness-observability/`

## Summary

Add non-authoritative progress and termination evidence to the complete balance-harness check, plus
an authority-preserving selector for one registered relevance row. This makes R-001's local/hosted
cost experiment valid without changing the population, gameplay, balance, acceptance result, work
budget, or CI contract being measured.

## Motivation

The current hosted `make verify-harness HARNESS_WORKERS=12` repeatedly reaches the 30-minute kill
without a harness verdict. The same complete `harness-check` has completed locally in roughly
18 minutes, but the CLI emits no internal phase/row progress. A killed run therefore leaves no
machine-readable identity, current objective, completed work, guard state, or valid denominator.

The standard pacing phase parallelizes across its 300 runs; registered relevance rows then execute
serially. A scenario-only relevance command is not an admissible substitute because it bypasses
`LoadRegisteredRelevanceSuite` and therefore can have a different active epoch constants identity
and deterministic population. Measurement must observe the governed command or select through the
same registry authority.

This RFC owns instrumentation only. It does not authorize an optimization disguised as a
measurement repair.

## Specification

### H1 — Explicit non-authoritative observation artifact

The balance-harness CLI accepts an optional explicit observation path. When present, it writes a
canonical UTF-8 JSON `harness_observation.v1` checkpoint at startup and atomically replaces it at
every objective start, progress checkpoint, completion, or controlled termination. The artifact is
diagnostic and contains `authoritative: false`; it is never a golden, baseline, epoch input, or
acceptance substitute.

The top-level shape is fixed-struct JSON with this information:

- schema version, kind, `authoritative: false`, and the requested CLI mode;
- process start/update/finish timestamps and elapsed milliseconds from Go's monotonic clock;
- terminal state: `running`, `complete`, or `incomplete`;
- termination kind: null while running, otherwise `objective`, `error`, `signal`, or
  `interrupted`; external loss inferred from a surviving running checkpoint is `interrupted`;
- current objective ID, the complete ordered objective-ID declaration persisted before work, and
  an ordered list of realized objective records that must match that declaration exactly;
- standard pacing scenario/constants identity and active epoch ID/constants identity when loaded;
- registry identity for each relevance row: registry index, scenario/catalog/policy/golden paths,
  scenario hash, relevance-policy hash, accepted constants hash, and active flag;
- declared/executed runs and transitions when the underlying governed report exposes them;
- explicit guard, exclusion, and truncation state; absent evidence is represented as `unknown`,
  never silently as zero/false;
- ordered errors, if any.

Objective IDs are mechanical and stable: repository guards, standard pacing, and one
`relevance:<registry-index>:<scenario-path>` objective per registered row. Later objectives require
an RFC amendment; player-facing flavor never enters this artifact.

Successful process exit is allowed only after the final checkpoint has `state: complete`,
`termination: objective`, no errors, every declared objective complete, no fired guard, no
truncation, and every known declared/executed cardinality relation valid. The ordinary harness
acceptance checks remain independently binding.

### H2 — Fail-loud interruption semantics

The recorder writes checkpoints through create-in-directory plus rename so readers see either the
prior complete JSON or the next complete JSON, never a torn file. `SIGINT` and `SIGTERM` are
recorded as `state: incomplete`, `termination: signal` before nonzero exit. Controlled errors are
recorded as `termination: error` before nonzero exit.

An uncatchable termination cannot execute a finalizer. To keep that case honest, the most recent
atomic checkpoint remains `state: running`; the observation validator classifies such a surviving
artifact as incomplete/interrupted and rejects it. No artifact or artifact with `running` state is
never interpreted as successful measurement evidence.

### H3 — Progress boundary and work accounting

The CLI records start and completion for the standard pacing phase and every registered relevance
row. Relevance execution additionally reports bounded progress at completed governed run-arm
boundaries, including executed runs and transitions, without reading or mutating simulation state
outside the existing runner seams. Progress emission must be cheap and deterministic with respect
to *where* it occurs; wall durations are diagnostic and do not enter report bytes or run identity.

The observation records the scenario-declared run/transition ceilings and the final report's
declared/executed counts. It does not invent a total where the current contract exposes only a
maximum. Guard exhaustion, exclusion, or truncation is explicit and makes the observation
incomplete even if a partial domain report exists.

Instrumentation must not alter canonical harness/relevance/branch report bytes. With observation
disabled, the CLI's outputs and acceptance behavior remain byte-for-byte and exit-for-exit
compatible.

### H4 — Registry-aware row selector

The CLI accepts an optional relevance registry selector for diagnostic measurement. The selector
identifies exactly one entry from `LoadRelevanceRegistry`; it then loads that entry exclusively via
`LoadRegisteredRelevanceSuite`. Raw scenario paths are not accepted by this mode. Unknown,
ambiguous, duplicate, or non-registered selectors fail before simulation.

The selector writes the same non-authoritative observation shape and executes the exact registered
row population and active epoch binding used by `mode=check`. It validates the row's report and,
where applicable, its registered branch evidence exactly as the check path does. A selected-row
success proves only that row; it cannot produce or validate a complete-lane observation.

### H5 — Scope firewall

This RFC changes none of the following:

- catalogs, scenarios, policies, seeds, horizons, milestones, golden reports, baselines, or epoch
  bytes;
- work/run/transition budgets, guards, acceptance bounds, timeouts, worker counts, dispatch,
  parallelism, sharding, algorithms, or CI job topology;
- gameplay, API, persistence, player copy, deployment, or release status.

R-001 first measures the complete unchanged population locally and on hosted Linux. Only the
resulting evidence may support a separate optimization, dispatch, sharding, budget, or CI-contract
RFC. This RFC cannot be cited as authority for one.

## Deviations from design

None. The artifact is operational evidence for the existing deterministic harness; it adds no
content, balance, or player mechanic.

## Acceptance criteria

1. A cold complete check with observation enabled reaches the same exit and produces the same
   canonical harness/relevance bytes as observation disabled; the artifact closes complete with
   exact loaded identities and valid known cardinalities.
2. A registry-aware selected active row has the same scenario, policy, accepted epoch constants
   identity, canonical report bytes, and acceptance result as that row inside complete check.
   A scenario-only or unregistered selector fails before simulation.
3. Injected controlled failure and `SIGTERM` cases leave atomic parseable incomplete artifacts and
   nonzero exits. A seeded surviving `running` checkpoint, missing artifact, fired guard,
   truncation, exclusion, mismatched identity, and declared/executed cardinality mismatch are each
   rejected by the observation validator.
4. A test severs progress completion while retaining successful domain work; validation still
   fails. Another test severs the active registry constants binding; selected-row validation fails.
5. `gofmt`, cold `go test`/`go vet` for affected packages, schema validation, and the existing
   harness guard/check populations pass. No governed data/report artifact changes.
6. `docs/balance-harness.md`, this RFC's plan/log, the platform R-001 records, and the active RFC
   index describe the same instrument and limitations. The complete implementation range receives
   the mandatory cross-party designated review before archival.

## Open questions

None for instrumentation. D-014's CI latency/topology choice remains explicitly unresolved until
R-001 obtains complete local and hosted measurements.

## Changelog

- 2026-08-21: created and accepted from Marco's bounded R-001 instrumentation ruling; no CI or
  harness-budget decision adopted.
- 2026-08-21: implementation started after authority checkpoint `9c71562`.
- 2026-08-21: internal review added pre-work objective declaration/equality and retained monotonic
  objective starts before the implementation handoff.
- 2026-08-21: implemented and archived after Claude's designated cross-party approval of
  `9c71562..afd4fb2`, recorded at `96a574d`.
