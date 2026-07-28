# Production Engine & Intent API — append-only log

## 2026-07-28 — acceptance and implementation start

- Re-reviewed the eight executable contracts added at `8e8938b` against the implemented catalog,
  ledger, save, and numeric boundaries.
- Closed remaining contract contradictions before acceptance: removed manual-click events, moved
  the large chaos test out of the local acceptance list, made the token bucket exact integer state,
  specified event payloads, separated terminal-rejection persistence from applied mutations, and
  made online/offline classification a trusted internal input rather than client data.
- Accepted locally at `d834793`; no push.
- Implementation starts with catalog v3 because save v3, progress evaluation, manual actions, and
  multiplier validation all depend on its closed data model.
