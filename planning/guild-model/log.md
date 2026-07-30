# Guild Model — implementation log

## 2026-07-30 — accepted and started

- The independent Faction remediation review approved FB-1/F2a and the LOW batch, so Guild's GA
  parent contract is implemented and archived. The owner explicitly placed Guild next.
- GB supplies the complete strict Phase-0 catalog; GC supplies a unique integer-only clearing
  answer, including the non-redistribution rule and labeled NPC counterparty. No balance mechanics
  need to be inferred.
- Implementation order is catalog/epoch → storage/lifecycle → tithe/Health/exchange → transport
  resolver → canonical docs and full verification. Per-change commits remain behind the mandatory
  independent review gate before archival.

## 2026-07-30 — catalog mint and owner-boundary audit

- Added the strict GB loader and matching client schema, then minted Epoch 3 with Guild as an
  append-only artifact and regenerated only the constants identity in the isolated harness commit.
- `DESIGN-GAP (name moderation)`: G1 refers to an existing name-moderation charset, but no runtime
  validator exists. The Guild service therefore requires an injected fail-closed `NameValidator`;
  it does not invent a charset. Composition remains blocked until the account/UGC owner supplies it.
- `DESIGN-GAP (tithe units)`: G3 declares integer `guild_xp` as a percentage of arbitrary Decimal
  production deltas but defines no Decimal→int64 normalization, resource basis, or overflow rule.
  Storage and the server-derived projection boundary can land; XP mutation cannot be improvised.
- `DESIGN-GAP (guild Health sample)`: `guild_health_inputs(active_founders,tithed_xp)` is declared,
  but no normative formula maps those two integers to the 0..1,000,000 guild Health term. The
  existing Commons weighted-Health function cannot derive that missing denominator.
- `DESIGN-GAP (account deletion)`: account deletion physically removes `accounts`, while Guild
  history and leadership reference account identity. The RFC defines New-Founder survival but not
  leader succession, disbanding, or anonymization on account deletion. Production composition is
  fail-closed rather than letting the new foreign keys brick the existing deletion contract.

## 2026-07-30 — structural implementation round complete; review gate open

- Commits `3c9f770..d519003` implement the strict catalog and Epoch-3 mint, isolated harness
  identity refresh, Postgres lifecycle model, closed account intents, byte-identical idempotency,
  append-only membership periods, cap/leader serialization, applications/invitations, role
  transfer, manual and seven-day automatic disband, and canonical docs.
- GC is a pure integer kernel with raw-account ordering, base/remainder allocation, per-boundary
  intake/headroom limits, deliberately no redistribution, absent-link inertia, and the reduced-rate
  labeled NPC path. Detached kernel values were renamed after the full writer-closure gate correctly
  rejected authoritative-looking `StockUnits` field names outside the Faction owner.
- Active membership implements Guild authorization. Join/leave commits also append a durable,
  leased presence-outbox row; the transport relay emits schema-validated `guild:{id}` presence
  envelopes and marks publication only after success. The authenticated account router exposes the
  separate exact-schema Guild intent surface.
- Real Postgres integration proves idempotent replay, concurrent cap admission, leave denial in
  authz, rejoin-as-second-history-row, deletion trigger refusal, and the exact grace boundary. The
  entire Compose integration suite is green.
- `make verify` is green: Go vet/tests, formula and epoch/baseline history guards, deterministic
  pacing harness, client typecheck/build, 6,452 client unit tests, schema parity including Guild,
  and 19,365 browser tests.
- Completion is intentionally blocked on GD1–GD6 now listed in the RFC. The most important newly
  surfaced gaps are Decimal production→int64 XP dimensionality, XP→Guild-Health normalization,
  mixed-epoch clearing authority/transaction ownership, and rounding-sensitive placement of the
  named `stock_consumption` slot. No runtime placeholder implements any of them.
