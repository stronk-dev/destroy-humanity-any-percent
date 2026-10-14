# Doctrine Choice and Compute Credit

This foundation closes two replay-owned mechanics without shipping doctrine effects or new tier
content.

## Doctrine choice

The optional hash-pinned `doctrines` artifact is strict schema v1. Each row declares an adjacent
tier transition, its exact source tier and gate, and at least two sorted mechanical doctrine IDs.
Go and TypeScript validate identical shared vectors and close every doctrine reference against the
pinned Routes artifact.

Go checks the exact original root and row names before struct decoding; case aliases refuse.
`schema_version` and `source_tier` must be explicit, non-null integers. An explicit source tier
zero remains valid in a matching structural fixture; omission/null does not silently become zero,
and a final null version does not inherit an earlier duplicate's value. These admission corrections
advance the shared kernel to `0.3.194`, without changing choice eligibility or burst arithmetic.
Escaped canonical names and complete last-value controls remain; no global duplicate-name
rejection is introduced.

`doctrine.DoctrineSchemas()` exports the closed v1 structural descriptor for API Foundation C18.
Current and stored catalogs validate with their corresponding Routes artifacts without byte changes.
Adjacent transition/gate/tier agreement, branching/order and cross-artifact Route bindings still
require the loader and `ValidateRoutes`. The public catalog response union is not activated here.

`pick_doctrine` is a Company command with exact fields `transition_id` and `doctrine_id`. It is
legal only at the declared source tier, before the declared gate is crossed, and while that
transition has no choice. A new command for an already-chosen transition rejects; ordinary intent-
ID replay still returns its original receipt. At the exact declared boundary, `cross_gate` rejects
until the choice is committed. The applied event is `doctrine_picked`.

Doctrine effects are not part of this implementation. Choices immediately satisfy existing Route
predicates, while production contributions, tree unlocks, and faction combinations require a
successor content RFC. Epoch 6 already minted the identity-only Doctrine artifact alongside the
Meters/Achievements chain and the T3→T4 permit gate; the current epoch-8 manifest still pins it.
The shared diagnostic fixture is not a new content mint or proof of natural player progression.

## Compute Credit burst

`spend_compute_credit` accepts a positive safe-integer `amount_ms` and the closed target
`accelerate`. The amount is both the exact Compute Credit debit and the boosted wall duration.
Spending rejects rather than clamping when the amount exceeds the pinned cap, the balance is too
small, or another burst is active.

The Company save persists only `compute_burst_remaining_ms`; speed always resolves from the run's
pinned economy artifact. Evaluation consumes `min(elapsed_ms, remaining_ms)` in both online and
offline modes. Its production-eligible part receives a second rate segment at
`burst_speed - 1`, integrated with the ordinary rate inside the same fixed provision buckets and
one explicit quantization boundary. Offline elapsed time past the production cap can expire the
burst but cannot create bonus production. Exit resets the remainder to zero. The applied event is
`compute_credit_spent` and records the exact debit, duration, target, and pinned speed.

Company save v17 owns the burst field. It activates only for a new run pinned to an epoch that
contains Meters, Achievements, and Doctrines together; the codec and replay path exist before that
content mint. There is no auto-spend setting or additional spend target.

The Go-authored sequential replay corpus is consumed byte-for-byte by TypeScript and covers the
choice requirement, write-once choice, activation, partial and complete consumption, every
semantic spend rejection, and malformed input. The engine suite separately covers partitioned
evaluation, a provision-boundary crossing, a 25-hour offline return, and Exit reset.
