# Route Registry Event-Order Convergence — append-only log

## 2026-07-29 — start

- D2 was independently demonstrated on Postgres: B's 15:00 event projected first beats A's 14:00
  event delivered later, while event-order rebuild chooses A. The 100-Knowledge Registry-first
  grant follows the non-convergent row.
- The follow-up makes `(occurred_at,event_id)` authoritative across batches. It explicitly covers
  already-spent provisional grants with projection debt; silently preserving the false bonus or
  allowing a negative founder save would violate the existing currency/save contracts.

## 2026-07-29 — implementation

- Registry decisions now take transaction advisory locks in sorted route-ID order, avoiding the
  absent-row race and keeping multi-route advisory-lock acquisition globally ordered.
- An earlier delivery compensates the exact active Registry grant, resets naming to the incoming
  event's reservation, and grants the incoming winner under its own constants epoch.
- `founder_route_state.route_knowledge_debt` preserves non-negative saves when a provisional award
  was already spent. Positive grants repay debt first; read-repair reconstructs the signed net from
  uncompensated grants and permanent hint purchases.
- Postgres coverage now includes reverse delivery, a published provisional name, a spent award,
  replay, immutable-history repair, later debt repayment, equal-timestamp event-ID ordering, and
  rollback when the old grant is missing.
- Verification is green: the focused package and full Postgres integration suites pass, followed by
  `make verify` (Go tests/vet/formula and harness gates; 6,412 client tests; 19,245 browser tests).
  The batch now waits at the mandatory independent diff-review gate before archival.

## 2026-07-29 (claude — independent review of 638844d..ae41b5e: APPROVED)

Full diff of both commits. The hardest remediation so far, and it holds:

- **Convergence is decided in exactly one place, in Go.** Ordered advisory locks (sorted route
  IDs, salted key — deadlock-free by construction), `FOR UPDATE` read, then the three-way
  decision with `eventBefore(occurred_at, event_id)` — the comparison never delegates to
  Postgres text collation, so live projection and from-scratch rebuild share one comparator.
  `TestProjectorIntegrationConvergesAcrossDeliveryOrder` asserts the D2 reproducer directly.
- **Compensation is immutable-history accounting done right:** the displaced grant is located
  by its event ID with a windowed count guarding duplicate/missing matches (abort → old decision
  stands, per D3); one `compensation` event appended; the debt model keeps founder saves
  non-negative while making displacement mint no net Knowledge — the spend-then-displace test
  asserts the exact balance/debt split and that later grants repay debt first.
- **The naming reset is deliberate and spec'd:** a provisional loser's name cannot exist in a
  from-scratch projection, so convergence *requires* the reset; the visible-name-flip window is
  projection-lag-scale and the RFC owns the tradeoff explicitly.
- Retry idempotency preserved through the existing claim (D1's lesson applied); suite green.

No findings. D2 clear to archive; A1+D4 (the guard) next.
