# Route Registry Event-Order Convergence — append-only log

## 2026-07-29 — start

- D2 was independently demonstrated on Postgres: B's 15:00 event projected first beats A's 14:00
  event delivered later, while event-order rebuild chooses A. The 100-Knowledge Registry-first
  grant follows the non-convergent row.
- The follow-up makes `(occurred_at,event_id)` authoritative across batches. It explicitly covers
  already-spent provisional grants with projection debt; silently preserving the false bonus or
  allowing a negative founder save would violate the existing currency/save contracts.
