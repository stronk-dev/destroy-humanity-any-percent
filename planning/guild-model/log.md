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
