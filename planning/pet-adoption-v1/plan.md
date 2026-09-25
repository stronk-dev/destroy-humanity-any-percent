# Pet Adoption v1 implementation plan

RFC: `rfc/pet-adoption-v1.md` (accepted 2026-09-25; every OD at its recommended default).
Implemented by: Claude. Every batch awaits Codex's designated cross-party review. No
self-approval, no archival, and no production mint (fixture-first; OD-4/PA8.6 copy gates the mint).

Numbering, ruled by landing order: the RFC's "Founder v22" maps to the **next free Founder version
= v23**, because Reputation Tree v1 already took v22. The RFC's "replay-inputs v7" maps to the next
free replay-inputs version. Cosmetic Shop then takes v24.

- [ ] P1 Copy: the `companion` tone in both copy runtimes and the pipeline. The companion linter
  (PA8.2/AC13) rejects price, urgency, statistic and curtain tokens. A generated Go
  `copykeys.CompanionKeys()`. Candidate `pet.*` keys marked PENDING OWNER COPY.
- [ ] P2 `pet_species` loader (Go/TS parity) with shared fixtures (AC1), plus draw vectors for the
  pet ID, temperament and palette (AC3).
- [ ] P3 Bundle wiring: `pet_species` requires `pets` and `reputation_tree` (the scalar chain:
  Founder floor 23 implies v22's Reputation state).
- [ ] P4 Founder v23 codec and catalog-aware validation (AC2); activation at the new-run boundary
  and New-Founder-forward (AC14).
- [ ] P5 The `adopt_pet` intent: validation order, receipts, `pet_adopted.v1`, the migration, Go/TS
  Founder replay, idempotency (AC4, AC5, AC7, AC9, AC10), and the replay-inputs Founder carry (AC8).
- [ ] P6 The snapshot `features.pet` projection (AC11) and the adoption surface with its visual
  contract (AC12, AC13 DOM).
- [ ] P7 Economy isolation (AC15) and docs canon (AC17). AC16 is carried under the release floor
  (D-008/D-009/D-015 are unruled).
