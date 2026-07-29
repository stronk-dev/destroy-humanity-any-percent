# RFC: Route Registry Event-Order Convergence

- **Status:** implemented
- **Author:** Codex, from independently demonstrated review finding D2
- **Created:** 2026-07-29
- **Design refs:** `design/08-satire-flavor.md §6`; `design/06-tech.md §persistence`
- **Research / reproducer:** `planning/archived-four-review/log.md` D2
- **Amends:** `archive/gate-predicates-and-routes.md` C4-C6
- **Depends on:** `archive/commons-projection-retry-idempotency.md` (shared review lesson only)
- **Planning:** `planning/route-registry-event-order-convergence/`

## Summary

Make the public Registry converge on the earliest immutable `route_executed` event by
`(occurred_at,event_id)`, independent of projection batch or delivery order. The current
insert-if-absent awards the Registry and its 100-Knowledge bonus to whichever company stream is
projected first, so live projection and event-order rebuild can disagree permanently.

## Specification

### D1 — One serialized Registry decision

Every first-delivery `route_executed` projection takes a transaction advisory lock keyed by route
ID, then locks/reads `registry_routes`. Event order is ascending `(occurred_at,event_id)`; UUID
comparison uses PostgreSQL/Go byte-equivalent canonical ordering.

- No row: insert the incoming event as winner with execution count one.
- Existing row and incoming event is later: increment execution count only.
- Existing row and incoming event is earlier: increment execution count and replace
  `first_event_id`, `first_founder_id`, `occurred_at`, house name, naming reservation, and naming
  state with the incoming event's values.

Resetting the naming state is required for convergence: a name submitted by a provisional loser
could not exist in a from-scratch event-order projection. The incoming winner receives a fresh
`reserved` state whose deadline remains exactly incoming `occurred_at + 72h`; normal expiry can
immediately house-name an already-expired reservation.

### D2 — Immutable grant compensation

On displacement, locate the active `route_knowledge_granted` event with source `registry_first`
derived from the old winning execution. Append one standard `compensation` event referencing that
grant with reason `route.registry_first_reordered`; never delete or rewrite history. Apply the exact
amount encoded in the old grant as a negative projection delta, then grant the incoming winner the
bonus from its own constants epoch.

The derived projection table adds non-negative `route_knowledge_debt`. A negative delta consumes
available cached balance first and records any remainder as debt. Positive grants repay debt before
becoming spendable. Founder save state remains non-negative and `RepairFounder` exposes only the
spendable balance. If a provisional winner already bought a permanent hint, the purchase remains
but its cost is carried as debt against future grants; displacement therefore mints no net
Knowledge and never creates an invalid save.

Read-repair reconstructs signed projection state from immutable history: uncompensated grants minus
hint costs. A negative result becomes debt with zero spendable balance. Compensation matching is by
the referenced grant event ID, not timing or mutable Registry state.

### D3 — Idempotency and transactionality

The existing route-event dedup claim, Registry decision, compensation, cached balance/debt update,
new winner grant, founder execution projection, and execution count commit in one transaction.
Retry of either event is a no-op. Any missing/malformed old grant or arithmetic overflow aborts the
incoming projection and leaves the old Registry decision intact.

## Acceptance criteria

1. Real Postgres projects B at 15:00 first and A at 14:00 second in separate batches; final winner
   is A, execution count is two, and replay changes nothing.
2. Exactly one active Registry-first grant remains in immutable-history accounting: B's grant has
   one compensation, A receives its epoch-correct grant, and cached balances match read-repair.
3. A provisional winner that spends 50 of 100 before displacement ends with zero spendable balance
   and 25 debt after retaining its founder-first 25; later grants repay debt first.
4. Naming reservation and house name move to A. Equal timestamps resolve by raw event-ID order.
5. The existing concurrent first-delivery race still produces one winner, two executions, and no
   duplicate grants.
6. Migration up/down and the full Postgres/Go/client/harness verification matrix pass.

## Deviations from design

None. This makes “first executor” a convergent event-ledger fact and uses the existing immutable
compensation contract for corrections.

## Open questions

None. Registry Analytics remains separately deferred.

## Changelog

- 2026-07-29: drafted, accepted, and implementation started after D1 independent approval and the
  owner-directed D2 remediation order.
- 2026-07-29: event-order Registry decisions, immutable grant compensation, projection debt,
  canonical docs, Postgres regressions, full verification, and independent review completed;
  archived with no findings.
