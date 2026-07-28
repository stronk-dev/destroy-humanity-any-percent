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

## 2026-07-28 — catalog contract v3

- Added catalog schema v3 across JSON Schema, Go, and TypeScript while retaining runtime loading of
  historical v1/v2 catalogs.
- The v3 root now requires manual actions, multiplier-source declarations, four explicit T0–T3
  progress coordinates, exact manual-bucket policy, and exact offline/Compute Credit policy.
- Added `balance/catalogs/phase0.json`, the first shipped mechanical catalog; schema verification now
  reports one production catalog instead of zero.
- Multiplier declarations authorize a provider/source/target/slot but do not activate it. Runtime
  factors remain state-owner outputs and must validate against those declarations.
- Go catalog tests, TypeScript strict checking/tests, and JSON Schema verification are green.

## 2026-07-28 — save state v3

- Save state v3 persists `compute_credit_ms`, `manual_token_milli`, and the canonical server-authored
  refill cursor beside balances, generator counts, and the production evaluation cursor.
- V1 and v2 migration paths initialize credits to zero, fill the manual bucket from the immutable
  catalog policy, and use the prior evaluation cursor as the refill baseline.
- Extended the checked-in migration corpus with independent v1→v3 and v2→v3 cases.
- Restore rejects credits/tokens above catalog caps, unsafe integers, company-only state in another
  scope, and refill cursors later than evaluated production state.
- Save unit and migration tests are green; the Postgres lifecycle suite will run with the later
  intent/event migration integration pass.
