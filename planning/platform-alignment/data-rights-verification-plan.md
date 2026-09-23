# D-008/D-009/D-015 — verification/board payload tranche

Status: predeclared research only. Product source `7e8aa70`; planning after
`5714ee3`. No product, deletion or retention contract change is authorized.

## Question and population

Complete the content-bearing-field trace for the five current verification/
board tables already SQL-column-mapped: `verification_queue.last_error`,
`verification_dead_letters.detail`, `verification_poison_dead_letters.detail`,
`verified_runs.variables`, and the event/run identifiers joining
`verification_projection_events` to a verified row. Determine what diagnostic
text can contain, whether it is bounded/sanitized, how the values are written
and read, and what survives account unlinking.

This tranche excludes underlying run archive bytes (mapped separately), public
board UI/API delivery, historical migration Down shapes and operator log sinks.

## Method and controls

1. Trace the latest Up constraints, producer path and consumer path for each
   selected field. Distinguish immutable terminal/dead rows from mutable
   pending/claimed queue rows. Trace `boundedDetail` back to the actual error
   source; a 512-character cap is not a privacy sanitizer.
2. Inspect the exact `variables` key/value allowlist and source facts. A
   projection marker UUID is a join key, not a content-free proof of erasure.
3. Run cold real-Postgres verifier/projector tests. If they inject a failure,
   inspect the written detail and test's negative cases. Name any fixture
   limitation, including whether account deletion and a populated board/dead
   letter occur in the same test. Avoid displaying secrets or private rows.

## Exit and refusal

Produce a bounded dossier and reconcile the parent inventory, decision
evidence, backlog for new falsifiable findings, 1.0 board and append-only
logs. No player export, deletion copy, retention duration, legal conclusion,
product/test/migration edit, RFC state or release promotion is inferred.
