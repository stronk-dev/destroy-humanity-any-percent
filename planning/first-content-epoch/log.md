# First Content Epoch implementation log

## 2026-08-07 — achievement copy orphan stage

The FCE-C7/FCE-C9 owner-authored literals are assembled as one byte-sorted `copy.v1` source and
generated through the existing Copy Pipeline. The thirteen rows intentionally remain orphans until
the achievements candidate is staged and the epoch-6 mint atomically installs its reference rows.
No active balance artifact or epoch seed changed.

Candidate hashes awaiting owner ratification:

- `copy/catalog/achievements-candidate.json` —
  `0dd211486b3e988c0fffa5311ed95c216f5bc08b4f9fd6ef7068409c3a091cf3`
- `client/src/copy/generated/catalog.json` —
  `2f299e2a3babb8f04e02c8406a127f8d3fbab321689038fbb22614cc6dba166b`
- `copy/generated/orphans.v1.json` —
  `dd7da21767197d5cdc627a68e197d0f56e2aa847e26dd8b6be18a64f20472be2`

`make copy-check` exits 0 with 65 generated keys and 53 intentional orphan warnings.

## 2026-08-08 — literal candidate composition and cross-runtime gate

The owner-ratified meters, achievements, and pets documents were copied byte-identically out of
the ignored drafting bank into tracked candidate paths. Their SHA-256 values remain exactly the
FCE-C3 pins. Permits Routes, Minigame API, Minigames/Pitch, and every unchanged base artifact now
compose from literal files rather than decoded test-helper mutations.

The sixteen-artifact candidate loads through `replaycatalog.Load` in Go and
`loadReplayCatalogBundle` in TypeScript under the same computed label:
`sha256:1a4463bcf67440ce1ba01e6c6eb850c0614329cac63064ef07725d042c7cf21a`.
The machine-readable source/production/hash table is
`planning/first-content-epoch/promotion-manifest.candidate.v1.json`. Its status remains
`candidate_awaiting_owner_ratification`; no production artifact, epoch row, or active seed changed.

Composition exposed three previously hidden byte decisions, filed as FCE-C10–C12 in the RFC:

- Categories must add `gate.t3_to_t4` to `full_gate_set`; unchanged bytes fail both loaders.
- Economy and Fiscal tests had been appending reviewed cross-artifact rows only in memory.
- The Soul Recovery fixture's `owner_kind: fixture` debit source is intentionally rejected for an
  epoch-seeded catalog; the ruled recovery-only production artifact therefore has an empty
  `debit_sources` array.

One implementation defect was found and fixed during the parity pass: TypeScript already enforced
meter/economy resource separation, while the Go composition loader omitted its existing validator.
Go now calls `ValidateResourceSeparation` over the pinned Economy resources, a discriminating
collision test rejects the previously accepted bundle, and kernel semantics advance honestly from
0.3.85 to 0.3.86.

Focused evidence from the repository root:

- `make test-go GO_PACKAGES='./replaycatalog' GO_TEST_FLAGS='-run FirstContentCandidate -count=1 -v'`
  — three tests pass, including stale-category and meter/resource-collision rejection.
- `make test-client` — 6,607 passed, 3 skipped; the shared candidate loads and the stale category
  set rejects in TypeScript.
- `make verify` — exits 0 after the complete balance harness, Go vet/all-package tests, generated
  artifact drift checks, 0-error TypeScript/Svelte diagnostics, 6,607 client tests, kernel-history
  guard at 0.3.86, copy/schema/content-manifest gates, and 19,830 browser assertions.

Pending before the manifest can become owner-approved: C10–C12 rulings/ratification, the candidate
designated review (which replaces every `pending` verdict field), and the composed harness report.
