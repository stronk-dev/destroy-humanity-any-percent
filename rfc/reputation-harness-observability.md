# RFC: Reputation Harness Observability

- **Status:** draft — not implementation authority
- **Author:** Codex
- **Created:** 2026-10-15
- **Design refs:** `design/00-vision.md` pillar 7; `design/06-tech.md` Balance & event data;
  `design/12-content-pipeline.md` §7
- **Depends on:** Harness Observability (implemented), Reputation Tree v1 (implementing),
  CI Baseline (implementing)
- **Parent / amends:** `archive/harness-observability.md` H1–H3; applies CI Baseline D3
- **Supersedes / superseded by:** —
- **Planning:** `planning/scaffolding-and-ci/` for delivery; Reputation plan owns study acceptance

## Summary

Give the existing Reputation fixture checks interruption-safe evidence through the shared
harness recorder. A versioned fixture profile identifies the full population and preserves
negative results. It is not a standard harness observation, a minted epoch, or Reputation
acceptance. Bring the additional maintenance job under CI D3's existing timeout and upload floor.

## Motivation

RP-634 identifies a 60-minute test/job pair with no observation validation, interruption margin
or always-uploaded artifact. A completed local run does not prove that a killed hosted run
leaves usable evidence. The archived observation contract explicitly requires an amendment
for additional objectives. Its v1 active-epoch identity also cannot truthfully describe the
Reputation fixture bundle and effective Prestige override.

## Specification

### O1 Existing studies and one recorder

The observed target runs exactly `TestReputationCareerStartersShortenRunThree` followed by
`TestReputationTreeRelevance`, cold and explicitly opted in. Reuse their shared producers,
projectors, source admission and report comparisons. Observation-disabled behavior stays
unchanged. Do not change scenarios, fixture threshold, policies, seeds, masks, horizons,
parallelism, report bytes or assertions; do not reuse one study's simulations to shorten another.

Add `make reputation-harness-observe REPUTATION_OBSERVATION=<explicit path>` and an independent
`make reputation-observation-check REPUTATION_OBSERVATION=<same path>`. Reuse the existing
recorder's atomic writer, clocks, lifecycle and interruption machinery, with a typed fixture
profile; do not add a second recorder or parse verbose test logs as a measurement authority.
The ordinary `reputation-harness-check` remains available unchanged.

### O2 Versioned fixture profile

Use `schema_version:2`, `kind:"harness_observation.v2"`, `authoritative:false` and
`mode:"reputation-fixture-check"`. Existing v1 check/registered-row artifacts, API signatures
and validation remain unchanged. A fixture artifact cannot validate as either v1 mode.

Retain the common timestamp, monotonic elapsed-time, state, termination, ordered declaration,
objective, work, condition and error fields. Replace v1's active-epoch fields with `producer`
and each objective's registry `identity` with `fixture_identity`; the v1 identity fields are
absent in this profile. These replacement fields have the following fixed shapes:

- `producer`: `commit`, `server_tree`, `balance_tree`, `kernel_version`, `go_version`, `goos`
  and `goarch`, using the existing Reputation measurement identity and runtime. Inputs must be committed
  and clean before and after work; use the existing identity/clean-input checks.
- Per-objective `fixture_identity`: `sources_sha256`, `retained_report_path` and
  `expected_report_sha256`. The source hash covers the canonical ordered
  `ReputationCareerMeasurementSource` array; the report hash covers its exact retained bytes.
  Each source already binds run key, effective bundle, first-hour/effective Prestige policy,
  experiment, horizon, purchase policy and mask. Bind expected report bytes to the producer's
  committed report blob, not an unchecked working-tree replacement.
- Per-objective `work_unit:"career_arm"` and nullable `observed_report_sha256`.
  Hashes here use the existing measurement helper's 64 lowercase hexadecimal representation.
  A report hash is populated only from fresh producer bytes, before comparison with the pin.

Do not populate v1's active-epoch fields with the fixture's mixed constants identity. Fixture
identity is explicit; the complete ordered source declarations are included in the diagnostic
artifact bundle so their digest and population can be recomputed independently.
Companions use the explicit checkpoint path plus `.sources.json`, `.h4.json` and `.h5.json`.
The source companion is a closed object with `schema_version:1`, `h4` and `h5` source arrays;
H4 orders treatment/control within each pair, H5 orders baseline/masks within each seed.
The validator reads these exact companions, not arbitrary paths supplied by an artifact.
Retained paths are the existing `planning/reputation-tree-v1/career-h4.v1.json` and
`planning/reputation-tree-v1/relevance-h5.v1.json`, resolved at the producer commit.
Observed mode refuses either report-update environment flag: it cannot rewrite those pins.

Declare exactly `reputation-fixture:h4` and `reputation-fixture:h5`, in that order, with
matching kinds `reputation_fixture_h4` and `reputation_fixture_h5`. Require declaration,
realized IDs, kinds and fixture bindings to agree; changing both copies of a name is insufficient.
Persist the declaration before the first simulation. No row selector, cached artifact replay,
empty test selector, missing opt-in or test repetition can produce a complete fixture observation.

### O3 Population and results

Derive declarations from the same suite/tree/configuration as the current producers. The
existing cohort is 97 seeds (64 Chaos, 32 Casual, 1 Reference), with nine node masks. H4 declares
194 career arms, treatment and control for each seed. H5 declares 970, one baseline and nine
masks per seed. These are whole career attempts, not individual Companies or reached gates.
Require exact source order, uniqueness and cardinality; no new population is selected here.

Record progress after completed admitted arms without altering result order or simulation.
Executed counts are observed work, not inferred from declarations. Transition counts remain
unknown unless the existing producer exposes them. Guard/exclusion/truncation conditions remain
unknown until witnessed; guard exhaustion, missing arms, skipped tests and dropped work prevent
completion. A domain row with a null clock or a declared exclusion is still measured work,
not permission to remove an arm from the denominator.
Clear work conditions require the existing result's completed outcome and empty invariant-failure
list, along with successful source admission; absence of a logged error alone is insufficient.

Retain fresh H4/H5 report bytes alongside the checkpoint and source declarations, and validate
them with the existing projectors/recomposition/admission checks. Preserve H4's failed gate,
violations and all-finite populations, and H5's conditional distributions and reasoned exclusions.
No new H5 pass flag may imply that its currently missing first-elective-Exit dimension was tested.

`state:complete` means both named checks actually ran, every declared arm was recorded,
reports were validated and both existing pin comparisons passed. It does not mean H4/H5 balance
criteria passed. A valid negative finding stays negative; changing criteria or approving
Reputation/content remains with the owning RFC. Report drift, source mismatch or a test failure
closes incomplete and exits nonzero. If Go continues to the second check after a first-check
failure, retain that check's observed progress too; never mark unobserved work complete.

### O4 Interruption and CI

Write an atomic startup checkpoint before simulation. Controlled failures and SIGINT/SIGTERM
leave parseable incomplete evidence and nonzero exit. An uncatchable kill leaves a running
checkpoint that validation rejects as interrupted. Missing files, unknown fields, trailing JSON,
partial work and forged profile/identity/report bindings cannot validate as complete.

The root observed target must deliver the real wrapper's signal to the recorder-owning process;
test this through the actual Make/process boundary, not only a direct signal to a helper.
Compilation/setup failure with no artifact is a failure, never a successful measurement.

Keep the job scheduled/manual and nonblocking. Apply CI D3's 55-minute job and
`timeout --signal=INT --kill-after=30s 50m` command, validate successful completion and always
upload the checkpoint plus any generated diagnostic companions. A missing checkpoint is an
upload failure; an interrupted bundle is retained but never promoted. Retain the existing
read-only permissions, module-only cache and 30-day artifact retention. Extend the existing
topology guard for the real target, exact population, bounded command, validator and upload.
Neither the blocking five-minute floor nor any study work limit grows.

## Deviations from design

None. This amends the observation contract only for fixture evidence and implements the
already-ruled CI floor. It does not resolve R10's career-data/purchase-policy, metric,
epsilon, attribution or content-mint requirements.

## Acceptance criteria

1. Run the full unchanged two-check population cold with observation disabled and enabled.
   Report bytes and domain verdicts agree, including red/drift results; do not regenerate a
   pin to obtain a pass. Complete evidence accounts for 194 H4 and 970 H5 admitted arms.
2. The validator rejects missing/duplicated/reordered arms, partial or skipped phases, wrong
   producer/source/report identity, forged objective/profile names, changed report bytes,
   fired guards, truncation and an invented transition total. Valid negative H4 evidence
   and reasoned domain exclusions retain their exact populations.
3. Execute controlled failure and both signal cases, including the actual root Make wrapper.
   Atomic incomplete evidence survives and validation exits nonzero. Kill a subprocess without
   a finalizer; its surviving running checkpoint cannot validate as complete. A first-phase
   failure does not conceal execution of a second phase that the unchanged test runner performs.
4. Current real-source CI topology and invalid recipe/timeout/validator/upload controls execute;
   record one hosted run's complete or explicit incomplete bundle when publication is authorized.
   Local passing checks are not that hosted result.
5. Existing v1 recorder/registered-row/signal tests and strict validation remain green, as do
   affected cold Go/vet checks. Update canonical docs and the owning CI log; complete designated
   cross-party review before acceptance/archival. No balance/report rewrite or kernel bump.

## Open questions

Adoption and designated normative review remain required. No new gameplay or balance choice
is proposed; unresolved Reputation acceptance questions remain in its own plan.

## Changelog

- 2026-10-15: draft for RP-634, preserving the shared recorder, full fixture population,
  existing failed criteria and CI D3 limits. Not accepted or implemented.
