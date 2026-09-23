# D-008/D-009/D-015 field-disposition research — immutable history and boards

Status: research only; no adopted export/deletion/retention policy.
Coordinate: product source `7e8aa70`; confirm no later product-path change at execution.

## Question and population

Classify every SQL column in the nine current run/Founder-history tables and five
verification/board tables listed by `data-rights-inventory.md`. Together with the completed
account/save tranche this covers 24 of 60 game tables, not the other 36 or non-DB stores.
These fourteen tables are chosen because account deletion keeps replay/board-related history,
whose joinable identities, JSON payloads and dead letters can invalidate a blanket erasure or
current-state-only export claim.

## Method and controls

1. Build current schemas from original Up DDL and subsequent Up amendments through migration
   `00074`; enumerate exact fields, keys and any immutability/retention constraints. Do not
   infer current state from a Down body or archived RFC.
2. Trace production writes/readers, account-deletion effects, archival/projector cleanup and
   any composed scheduler. Distinguish immutable identity, payload, retry metadata and public
   board projection. Mark unknown or version-dependent JSON where source does not prove a bound.
3. Negative controls: a deleted account must not make Founder/run UUID history disappear by
   assumption; a dead-letter table must not be called short-lived merely because it is a queue.
   Positive control: verify the account deletion witness's retained streams/Founders and locate
   a production reader for replay/board authority.

## Exit and refusal

Produce a fourteen-table field map, cite source and executable evidence, update the parent
inventory/decision queue and record newly confirmed defects in the shared backlog. State
which JSON payloads and downstream tables are still unclassified. No product, migration,
assertion or player copy change; no invented duration, erasure, legal basis or export format.
