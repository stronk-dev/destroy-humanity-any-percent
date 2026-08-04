# Meters

The implemented Meters foundation provides the strict catalog boundary and a pure deterministic
transition kernel in Go and TypeScript; no production meter artifact has been minted yet.

Phase A recognizes exactly eleven run-scoped Company values in the integer range 0–100: Standing
and Grievance for users, employees, regulators, press, and investors, plus `doom.probability`.
Externality is not a meter—it remains addressed ledger facts for World Layer. Founder Soul remains
the existing separate carry value and is read-only to ordinary Company transitions.

Catalogs are schema v1 and declare initial values, floor-ordered bands, optional exact decay, and
two causal input kinds: newly emitted ledger facts and active contribution-slot sources. Meter IDs
are structurally disjoint from economy resource IDs and the meters package cannot import economy or
production owners, so no intent can use a meter as ledger payment.

The transition kernel validates complete state maps, advances decay first, then applies newly
caused ledger facts and active contribution sources, clamps the final value, and derives at most
one prior-to-final band change per meter. Decay and rate inputs preserve their exact millisecond
remainders, are invariant to splitting an interval across evaluations, and receive zero elapsed
time offline. A shared cross-runtime corpus covers partitioning, band changes, target-phase reset,
and saturation.

Go and TypeScript loaders enforce the same closed eleven-ID set, input shapes, numeric bounds,
unique source bindings, and band ordering. `balance/meters.schema.json` is wired into the root
schema gate with discriminating pre-mint fixtures. Literal production floors, initials, rates, and
bindings remain balance data that must be supplied before the epoch artifact is minted.
