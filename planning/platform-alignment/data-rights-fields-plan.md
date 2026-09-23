# D-008/D-009/D-015 field-disposition research — account/save tranche

Status: research only, not an adopted export/deletion/retention policy.
Coordinate: product source `7e8aa70`, with no later product-path change at plan start.

## Question and population

For the seven current account/credential tables and three save/intent tables named in
`data-rights-inventory.md`, what exact live fields can retain identity, credentials, gameplay
history or joinable links after account deletion? These ten tables are the first tranche of the
60-table + non-database census, selected because a real Postgres diagnostic already found a
surviving account-linked bootstrap tombstone. The other 50 tables are explicitly not covered
and must not inherit a disposition from this tranche.

## Method and controls

1. Read each table's original Up DDL and every later Up amendment through migration 00074.
   Classify field groups at current schema, not from an early snapshot or Down body.
2. Trace production insert/update/delete and export readers for each group. Separate account
   deletion now from proposed erasure or retention, and mark unknown when a bound is not proved.
3. Check two negative controls: `bootstrap_receipts` ciphertext must be distinguished from its
   retained `account_id`/digest/times; `intent_records` must not be called 30-day-expiring merely
   because `PruneIntentRecords` exists without a production caller.
4. Check a positive control: live session/access credentials and account rows actually removed
   by the current deletion flow, without claiming archived Founder/save history disappears.

## Exit and refusal

Write a field-level decision table with producer, current deletion result, possible export
subject and retention unknowns, citing DDL/source. Surface any newly found contradiction in the
shared backlog and reconcile D-008/D-009/D-015 evidence in the same research checkpoint. This
may inform owner/legal rulings but cannot make them, authorize a destructive cleanup, promise
portable import, or close player workflow R-003. Product code, test assertions and authored
player copy stay untouched in this wave.
