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
