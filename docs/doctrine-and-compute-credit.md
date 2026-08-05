# Doctrine Choice and Compute Credit

This foundation closes two replay-owned mechanics without shipping doctrine effects or new tier
content.

## Doctrine choice

The optional hash-pinned `doctrines` artifact is strict schema v1. Each row declares an adjacent
tier transition, its exact source tier and gate, and at least two sorted mechanical doctrine IDs.
Go and TypeScript validate identical shared vectors and close every doctrine reference against the
pinned Routes artifact.

`pick_doctrine` is a Company command with exact fields `transition_id` and `doctrine_id`. It is
legal only at the declared source tier, before the declared gate is crossed, and while that
transition has no choice. A new command for an already-chosen transition rejects; ordinary intent-
ID replay still returns its original receipt. At the exact declared boundary, `cross_gate` rejects
until the choice is committed. The applied event is `doctrine_picked`.

Doctrine effects are not part of this implementation. Choices immediately satisfy existing Route
predicates, while production contributions, tree unlocks, and faction combinations require a
successor content RFC. The production epoch has no doctrine artifact yet; the shared fixture uses
`transition.t3_to_t4` solely to prove the mechanism.

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
