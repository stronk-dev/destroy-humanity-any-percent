# D-008/D-009/D-015 field-disposition research — shared catalog, Routes, Commons

Status: research only; no export, erasure or retention policy adopted.
Coordinate: product source `7e8aa70`; confirm no later product-path change at execution.

## Question and population

Classify every current SQL column in the 4 catalog/epoch, 5 Routes and 7 Commons tables
named by `data-rights-inventory.md` (16 tables). These are the shared or Founder-linked
families most likely to be mishandled by a single-account export/deletion rule. Together with
the previous tranches, this will cover 40/60 game tables, while Guild (13), Minigame (5),
Soul (1), Transport (1), nested payloads and non-DB stores remain explicitly unclassified.

## Method and controls

1. Derive current fields/constraints from Up migrations through `00074`, including later
   ALTERs. Trace producer, consumer, joinable identities, current account-deletion effect,
   cleanup and any immutable/history contract. Do not infer a bound from a method name.
2. Negative control: epoch/catalog rows are shared replay authority, not automatically one
   player's export or deletion target. Positive control: Founder-linked route/Commons records
   must remain visible in the inventory even if account ownership is unlinked.
3. Run relevant real-Postgres integration witnesses cold where available, but check their
   fixture population before citing them for account-deletion or rights outcomes. Do not
   combine independent green tests into a joined proof they never executed.

## Exit and refusal

Produce a 16-table exact-column map and evidence/unknowns, update the parent 60-table
inventory, decision queue, shared backlog (for new contradictions) and append-only logs.
The map may inform D-008/D-009/D-015 but cannot make owner/legal choices, delete shared
history, choose a portable import schema or close player rights R-003. No product, migration,
test assertion or authored copy change in this wave.
