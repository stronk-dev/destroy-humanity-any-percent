# Cosmetic Shop v1 implementation plan

RFC: `rfc/cosmetic-shop-v1.md` (accepted 2026-09-25, all ODs at recommended defaults).
Implemented by: Claude. Every range awaits Codex's designated cross-party review. There is no
self-approval, no archival and no production mint (fixture-first).

Numbering is landing-order. The RFC's "Founder v22" (OD-16: next free) becomes **Founder v24**,
because Reputation took v22 and Pet Adoption took v23. The replay-inputs carry and the events
migration take the next free numbers at landing.

- [ ] C1 Catalog grammar: `cosmetics` schema v1 loaders (Go `server/cosmetic`, TS
  `client/src/cosmetic`), shared reject fixtures, and the permanent-ID transition check (AC1, AC2).
- [ ] C2 Replay-bundle wiring: `cosmetics` joins the constants bundle (OD-10) and requires
  `pet_species` on the scalar Founder chain.
- [ ] C3 Founder v24 `cosmetics` state: codec, validation, activation at the new-run boundary,
  Exit carry and the replay-inputs carry, in both runtimes (AC3, AC4, AC8).
- [ ] C4 Intents `acquire_cosmetic` / `equip_cosmetic` / `unequip_cosmetic`, the three events and
  the migration, with Go/TS parity and Postgres integration (AC5, AC6, AC9).
- [ ] C5 Snapshot v4 optional `features.cosmetics` arm and its decoder (AC10).
- [ ] C6 Desk shelf, parody receipt, `CosmeticOverlay` and the curtain contract (AC11, AC12, AC15).
- [ ] C7 No real money: `verify-no-payment`, the copy currency lint, the Caddy
  `Permissions-Policy`/CSP, and the browser network trap (AC13, AC14).
- [ ] C8 Mechanical-isolation property test, package gates, docs (AC7, AC16).
