# RFC: Post-Review Integrity Remediation

- **Status:** implementing
- **Author:** Codex, from independently verified findings A2-A9 and D3
- **Created:** 2026-07-29
- **Design refs:** `design/02-economy-balancing.md §11`; `design/05-mmo.md`; `design/06-tech.md §idle-math`
- **Research / findings:** `planning/archived-four-review/log.md`
- **Amends:** `archive/balance-harness-foundation.md`; `archive/gate-predicates-and-routes.md`; `archive/commons-compact.md`; `archive/save-layer-and-migrations.md`
- **Planning:** `planning/post-review-integrity-remediation/`

## Summary

Close the executable integrity gaps found by the independent review of the harness, routes, and
Commons implementations. This follow-up makes the published Commons formula the only runtime
formula, includes Commons balance data in the constants identity and baseline review protocol,
and turns the review's unasserted contracts into failing gates.

## Specification

### D1 — One catalog-driven Commons health blend

Every live server and harness calculation of effective Commons Health must call
`commons.EffectiveHealthPPM`. Snapshot resolution is keyed by the committed revision's
`constants_hash`; no current-weight fallback exists. The generated formula artifact derives its
shape and source fingerprint from the same executable authority. A test must mutate the catalog
weights and prove both live paths move to the exact result returned by that authority.

### D2 — Balance identity is an artifact bundle

`constants_hash` is the SHA-256 identity of a named, deterministically ordered bundle of exact
artifact bytes. The Phase-0 production bundle contains the economy catalog and Commons catalog.
Names and byte lengths are framed before bytes so concatenation is unambiguous. Reordering bundle
inputs must not change the hash; changing either artifact must.

Harness scenarios name both artifacts. The baseline change guard treats `balance/commons/**` as
an accepted input. Adding Commons to an existing scenario is a reviewed balance-input change and
requires a separate artifact-only `BALANCE-CHANGE:` baseline regeneration commit.

### D3 — Closed harness semantics and attributable failures

The runtime loader rejects unknown milestone kinds before any run. Aggregate invariant failures
carry the complete run key: harness schema, scenario identity/version/hash, policy identity/version,
seed, and constants hash. Float-freedom is asserted across the actual report type and serialized
real-run output, not a toy shape.

### D4 — Compile-enforced route dependency boundary

CI enumerates the routes package's direct internal imports and rejects every package except the
explicit DTO/numeric allowlist. A grep for one forbidden package is not sufficient.

### D5 — Seam regressions become repository tests

Tests pin wrong-gate route rejection, negative multiplier rejection, the full run-key failure
diagnostic, and unknown milestone rejection. Existing invariant collector/outcome tests remain
the authoritative coverage unless a demonstrated missing path requires a test seam.

### D6 — Same-stream event order and corpus ratchets

For events with the same timestamp and stream, Commons projection orders by authoritative stream
revision before kind priority and event id. Cross-stream ties retain the deterministic kind/event
ordering because revisions are not comparable across streams.

The save migration corpus count must equal its committed baseline count. Adding or removing cases
therefore requires an explicit baseline ratchet; required case names remain enforced.

## Design gaps deliberately excluded

- **DESIGN-GAP A7:** three doctrine-keyed routes currently target a gate crossed before doctrine
  selection. Choosing a new gate or moving doctrine selection changes route mechanics and belongs
  in a doctrine/gate alignment RFC.
- **DESIGN-GAP S4:** whether collapsed-cohort merges may exceed target size is a population policy,
  not a correctness default. It belongs in Commons governance.

These exclusions do not block implementation or verification of D1-D6.

## Acceptance criteria

1. Retuning Commons weights changes projector and harness output identically and updates the
   published artifact.
2. Economy or Commons byte changes alter `constants_hash`; bundle iteration order does not.
3. Commons changes are accepted guard inputs and the current harness artifacts are regenerated in
   a separate `BALANCE-CHANGE:` commit.
4. Unknown milestone kinds and disallowed route imports fail runtime/CI respectively.
5. Failure diagnostics identify the exact run, and actual report output contains no JSON floats.
6. Same-time leave then re-sign in one stream follows revision order.
7. Migration corpus additions fail until the count baseline is updated.
8. Full Go, Postgres, schema, formula, harness, client, and browser verification is green.

## Deviations from design

None. The RFC makes already published data and accepted boundaries authoritative.

## Changelog

- 2026-07-29: drafted, accepted, and implementation started as the owner-directed large
  remediation round; mechanic-changing A7/S4 findings explicitly split as design gaps.
