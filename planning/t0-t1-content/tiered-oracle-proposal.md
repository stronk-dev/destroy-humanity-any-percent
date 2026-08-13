# Proposed contract: hash-triggered tiered relevance oracle

- **Status:** draft for owner ruling; this document authorizes no implementation.
- **Author:** Codex, responding to the 2026-08-13 owner direction in `log.md`.
- **Scope:** supersedes only the verification cadence and oracle-report portions of R11, the
  R-block, T01-C19, and T01-C20. The reference arm, relevance/trap verdicts, exact production
  simulation, 5% bound, and content data are unchanged.

## Decision summary

Split the current oracle into two tools with different claims:

1. **`deviation.v1` — mandatory on every relevance run.** Replay the real reference trajectory and
   force deterministic legal alternatives at a declared sample of decision coordinates. Complete
   each altered trajectory using the same T01-C20 ranked policy. This is a cheap, content-specific
   falsifier; it is not an optimality proof.
2. **`beam.v2` — mandatory when its input identity changes and before a content mint/release.** Run
   the existing terminal-completion beam under a separately measured deep budget. Commit its result
   as an immutable attestation keyed by every input that can change its meaning. Ordinary runs check
   that attestation rather than re-executing the deep search.

This preserves one scoring semantics: reference, forced-deviation completion, and beam completion
all use the exact T01-C20 projected-time policy. Independence comes from forcing a different legal
choice, not from inventing another ranker.

## 1. Oracle identity and invalidation

The deep-attestation identity is the SHA-256 of canonical JSON with exactly these keys:

```json
{
  "schema_version": 1,
  "constants_hash": "sha256:…",
  "scenario_hash": "sha256:…",
  "relevance_policy_hash": "sha256:…",
  "kernel_version": "0.0.0",
  "oracle_implementation_hash": "sha256:…"
}
```

`kernel_version` invalidates the proof when canonical simulation semantics change.
`oracle_implementation_hash` is generated from the exact bytes of the relevance loader, report
validator, registry, reference solver, oracle solver, and their declared schema version. It is not
a manual version string. Formatting-only source changes may invalidate an attestation; false
freshness is worse than an occasional extra deep run.

Changing any identity member makes the prior attestation **stale**. Content identity alone is not
enough: a solver change can invalidate yesterday's proof while reading identical balance bytes.

Deep results live under an immutable, full-identity-keyed path:

```text
testdata/harness/relevance/deep/<sha256-identity>/report.v1.json
```

A sorted registry maps each scenario path to its current deep identity and report. Existing
baseline/changeguard history rules own both registry and reports. A prior identity remains in Git;
updating an entry never rewrites its old attestation.

## 2. Always-run deterministic deviation tier

### Trace

The reference run records, for each decision:

- `decision_ordinal`;
- pre-decision canonical state hash and virtual time;
- chosen arm (`bank` or a purchasable ID);
- the sorted legal alternative arms at that state.

This trace is evidence only and does not change the reference policy.

### Probe selection

The scenario declares `deviation_probe_count` and `deviation_alternatives_per_probe`. Their launch
literals must be derived from measured runtime and ratified; this proposal does not choose them.

Eligible coordinates have at least one legal arm different from the reference choice. Coordinates
are ordered by:

```text
SHA256(oracle_identity || "relevance.deviation.v1" || decision_ordinal || state_hash)
```

Take the first `min(deviation_probe_count, eligible_count)` coordinates. At each coordinate, order
non-reference legal arms by the same construction with the arm ID appended and take the declared
number. This is deterministic, seed-free from wall time, and deliberately uncorrelated with the
T01-C20 ranking.

For each selected arm, restore the immutable state/revision/time captured at that reference
coordinate, force that one legal choice, then complete from the resulting state through the same
T01-C20 ranked policy. The probe fails if any completion beats the reference by more than
`greedy_gap_maximum_ppm`. A failing run may stop after the first witness, but records the witness
coordinate, forced arm, reference time, alternate time, and exact gap.

### Falsifiability proof

The registered fixture must contain a declared-parameter witness where a forced alternative beats
the reference beyond the bound. Reverting forced-choice execution, correlating selection with the
reference choice, or suppressing its finding must fail a cold green-gate test. This is separate
from the fixture proving `beam.v2` can fire.

### Honest scope

A green deviation run means only: "none of the declared deterministic probes found a >bound
counterexample." It does not mean the reference is within the bound globally.

## 3. Hash-triggered deep tier

`beam.v2` retains the reviewed search semantics now shipped:

- deterministic pre-rollout child selection with at least two children;
- state/time identity dedup and componentwise dominance;
- declared width;
- terminal completion through the exact T01-C20 ranked policy;
- deterministic suffix memoization;
- the registered declared-parameter negative control;
- width-monotonicity proof;
- fail loud if beam is slower than reference.

Equality is healthy, not a failure. Deep outcomes are the closed union:

- `improvement_within_bound` — beam is faster, gap ≤ bound;
- `tied` — byte-exact equal completion time; no improvement found;
- `gap_exceeded` — beam is faster by more than the bound;
- `search_regressed` — beam is slower than reference;
- `milestone_unreached`;
- `budget_exhausted`.

Only the first two are passing attestations.

The deep run may stop early on `gap_exceeded` because the oracle question has been answered. A pass
must complete the declared search; no truncated pass is legal.

## 4. Measurement and budgets

The cheap and deep tiers have separate transition budgets. Neither inherits the other's number.

To establish a deep budget:

1. Run a non-authoritative measurement with an explicitly supplied generous **measurement
   ceiling**. It may emit diagnostics but cannot create an attestation.
2. If it finishes, record actual uncached work, cache hits, elapsed time, and search result in the
   planning evidence. Elapsed wall time is never part of canonical attestation bytes.
3. Pin `deep_budget_max_transitions` to measured worst-case work ×2, rounded up to a reviewed round
   literal, per T01-C17 branch B.
4. Run again under the pinned budget. Only this authoritative run may create the attestation.

An incomplete measurement cannot set a budget. Because executed transitions count uncached work,
the attestation records `oracle_implementation_hash`, cache hits, and uncached transitions; a cache
implementation change invalidates the identity and requires remeasurement.

## 5. Report grammar and gate semantics

Replace the single ambiguous `greedy_oracle` status with:

```json
{
  "oracle_verification": {
    "schema_version": 1,
    "effective_status": "verified|stale|failed",
    "identity": "sha256:…",
    "deviation": {
      "kind": "deviation.v1",
      "status": "passed|counterexample|budget_exhausted|milestone_unreached",
      "eligible_coordinates": 0,
      "declared_probes": 0,
      "executed_probes": 0,
      "unprobed_coordinates": 0,
      "deep_search_executed": false,
      "best_alternate_ms": null,
      "gap_ppm": null,
      "witness": null
    },
    "deep": {
      "kind": "beam.v2",
      "status": "current|stale|missing",
      "attestation_identity": null,
      "attestation_report_hash": null,
      "result": null
    }
  }
}
```

`declared_probes`, `executed_probes`, and `unprobed_coordinates` make skipped coverage visible. The
cheap command always writes `deep_search_executed:false`; it cannot impersonate a deep run.

Effective status:

- deviation failure → `failed` and a named `greedy_oracle:deviation_*` finding;
- matching passing deep attestation + passing deviation tier → `verified`;
- missing/mismatched attestation → `stale` and `greedy_oracle:deep_stale`;
- failing attestation → `failed` with its exact deep outcome.

Therefore a balance change cannot get a verified green from the cheap tier. It becomes visibly
stale until the deep run is completed, reviewed, and committed.

## 6. Commands and cadence

- `make relevance-check`: full reference/content matrix + deviation tier + deep-attestation
  freshness check. Runs in ordinary verification. Never executes beam search.
- `make relevance-deep-measure`: explicit non-authoritative measurement; never writes the registry
  or an attestation.
- `make relevance-deep`: runs the complete beam under the pinned budget and writes the immutable
  candidate attestation. Used when identity changes and before mint/release sign-off.
- `make relevance-deep-check`: explicitly re-executes and byte-compares the current attestation.
  Runs when a new attestation is reviewed and as an optional explicit audit, not every ordinary CI
  or unchanged-content release.

No calendar trigger is normative. A scheduled run may call `relevance-deep-check`, but freshness is
defined exclusively by identity.

## 7. Guard and review requirements

- An identity change without a new deep attestation makes `relevance-check` fail stale.
- An attestation commit is governed like a golden: isolated subject class, only immutable report
  bytes plus its registry pointer, and preceding governed input changes required.
- The guard recomputes identity and the report hash; retargeting a registry entry to unrelated bytes
  fails.
- The designated cross-party review of a new attestation reruns `relevance-deep-check`. A
  self-generated attestation is not sufficient for mint/release eligibility. A later release with
  unchanged identity checks freshness and cites that verdict; it need not repeat the deep search.
- The inactive fixture retains both negative controls: deviation can fire and deep beam can fire.

## 8. Explicit non-goals and rejected shortcuts

- A prior deep pass does not authorize changed content, changed scenarios, changed policy, or
  changed oracle code.
- The T01-C20 closed form is a ranking key, not an admissible branch-and-bound lower bound.
- A cheap pass is never described as deep-verified.
- Calendar age alone neither invalidates nor renews an attestation.
- Raising the current T1 ceiling from the incomplete 5,000,000 run remains forbidden.

## Owner decisions requested

1. Accept `deviation.v1` as the always-run falsifier and `beam.v2` as the hash-triggered deep tier.
2. Accept the full oracle identity, including exact implementation bytes, rather than balance hash
   alone.
3. Accept `tied` as healthy and `search_regressed` as the strict beam-worse failure.
4. Accept stale deep identity as a blocking finding for content mint/release, while ordinary
   code-only CI remains cheap when identity is unchanged.
5. Authorize a one-time generous T1 **measurement-only** run after implementation; its result, not
   this proposal, determines the deep budget and deviation-probe literals.
