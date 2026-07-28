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

## 2026-07-28 — production evaluation and progress parity

- Added the neutral `server/multiplier` boundary with the closed slot union, canonical slot order,
  and mechanical runtime contribution type. Catalog declarations authorize contributions; runtime
  source/slot/target/factor mismatches reject before arithmetic.
- Generator rate evaluation applies base rate, exact owned count, slots in published order, and
  source ids in deterministic raw-byte order within each slot.
- Online evaluation uses the full trusted interval at 100%; offline evaluation applies the same
  constant-rate primitive at 90% for 24 h, banks half the excess as exact ms, and advances the
  persisted cursor without losing sub-millisecond remainder. Resource and credit hardcaps clamp to
  their published values; clock rollback produces no state change.
- Implemented `subProgressValue` over the catalog's closed coordinate union in Go and TypeScript.
  Both runtimes consume the new shared `testdata/production-engine.json` fixture.
- Production/economy/save Go suites and strict TypeScript/unit suites are green.

## 2026-07-28 — atomic intent and event persistence

- Added embedded Postgres migration 00002 for per-stream intent records and immutable v1 events.
  Events keep their originating revision number without an FK to prunable snapshot rows.
- Added `Store.ApplyIntent`: lock stream, replay by `(stream_id,intent_id,request_hash)`, verify
  expected revision, restore working state, and atomically commit applied save/event/receipt state.
  Terminal rejection receipts commit alone and never create a save revision or event.
- Added strict UUIDv7, hash, outcome, event-kind, and per-kind payload validation plus an explicit
  pruning API for a deployment-owned 30-day cutoff.
- Real Postgres testing exposed that JSONB rewrites whitespace, making first and replayed receipt
  bytes differ. Receipts now normalize through one deterministic JSON boundary before return and
  after load; the corrected integration suite proves identical replay, one event, one save mutation,
  hash conflict, rejection non-mutation, and pruning.
- Disposable Postgres container was stopped and removed after the green integration run.

## 2026-07-28 — authoritative intents

- Added strict parsing and canonical request hashing for the two accepted wire intents. UUIDv7,
  safe-integer fields, exact/max purchase modes, mechanical ids, root/nested field sets, trailing
  JSON, and semantic-invalid recording are explicit.
- `buy_generator` evaluates trusted elapsed time on working state, verifies exact or max cost,
  applies one ledger transaction, updates the exact owned count, emits one purchase event, and
  returns the net receipt plus canonical authoritative snapshot.
- `perform_manual_batch` refills/spends integer milli-tokens from server time, silently clamps the
  requested batch, commits no click event, and still reports `applied_count` including zero.
- Added the dormant numeric fallback diagnostic without changing affordability results. Successful
  fallback/clamp reports join the gameplay transaction; abort-only reports go to structured audit
  and metrics because there is deliberately no save revision to attach an event to.
- Real Postgres integration proves exact purchase, byte-identical replay, one purchase event,
  manual 50-action clamp, zero-action follow-up, terminal unaffordable non-mutation, and typed stale
  revision response. The Make target now runs all package integration tests.
- The retained tier-1 property gate drives both intents for 24 simulated hours across 200 seeded
  policies; every state remains finite/non-negative/encodable and every policy acquires a generator.
