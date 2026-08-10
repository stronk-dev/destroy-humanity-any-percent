# Meters

The Meters foundation provides the strict catalog boundary and a deterministic transition hook in
the shared Go/TypeScript replay kernels. Epoch 6 pins the production schema-v1 meter artifact and
new runs activate it through the paired Meters/Achievements foundation boundary.

Phase A recognizes exactly eleven run-scoped Company values in the integer range 0–100: Standing
and Grievance for users, employees, regulators, press, and investors, plus `doom.probability`.
Externality is not a meter—it remains addressed ledger facts for World Layer. Founder Soul remains
the existing separate carry value and is read-only to ordinary Company transitions.

Catalogs are schema v1 and declare initial values, floor-ordered bands, optional exact decay, and
two causal input kinds: newly emitted ledger facts and active contribution-slot sources. Meter IDs
are structurally disjoint from economy resource IDs. Both composed replay-catalog loaders enforce
that separation against the pinned Economy artifact; the meters package itself remains free of
economy and production imports, so no intent can use a meter as ledger payment.

The transition kernel validates complete state maps, advances decay first, then applies newly
caused ledger facts and active contribution sources, clamps the final value, and derives at most
one prior-to-final band change per meter. Decay and rate inputs preserve their exact millisecond
remainders, are invariant to splitting an interval across evaluations, and receive zero elapsed
time offline. The production hook derives each step from the delta of the canonical attended-time
ledger before and after the transition; the HTTP transport's `online` evaluation mode is not an
attendance claim. Shared cross-runtime cases prove that gaps of 5,001 ms and 25 hours are recorded
as offline, leave Meters stable, and still advance the authoritative stream. The corpus also covers
partitioning, band changes, target-phase reset, and saturation.

Go and TypeScript loaders enforce the same closed eleven-ID set, input shapes, numeric bounds,
unique source bindings, and band ordering. `balance/meters.schema.json` is wired into the root
schema gate. The epoch-6 artifact starts Standing at 50, Grievance at 0, and p(doom) at 50; every
row exposes `low` at 0 and `high` at 70. Standing and p(doom) decay toward 50, while Grievance
decays toward 0. The live causal bindings are `externality.emitted` → p(doom) +3 and active
`commons.member` → users' Standing +1 per attended hour. These remain declarative balance rows,
not code constants.

Save v16 activates Meters and Achievements together only on a new-run boundary whose pinned epoch
contains both artifacts. Active Company state carries complete meter value and remainder maps;
legacy runs keep the v14 placeholder bands and never read deploy-current catalogs. Go and
TypeScript replay loaders validate the paired artifact identity, exact state maps, Notoriety
reseed, active route context, and byte-identical ordinary and terminal transitions. Active v4
commands execute decay and causal inputs after the existing production hooks, emit exact
`meter_band_changed.v1` events in meter-ID order, and expose authoritative meter state in receipts.
The formula artifact publishes the transition order and carried arithmetic. Historical v3 rows
retain their pre-hook semantics. Epochs 1–5 preserve their earlier artifact sets and do not
retroactively activate Meters.
