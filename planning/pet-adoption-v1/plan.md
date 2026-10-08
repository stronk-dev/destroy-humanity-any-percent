# Pet Adoption v1 implementation plan

RFC: `rfc/pet-adoption-v1.md` (accepted 2026-09-25; every OD at its recommended default).
Implemented by: Claude. Every batch awaits Codex's designated cross-party review. No
self-approval, no archival, and no production mint (fixture-first; OD-4/PA8.6 copy gates the mint).

2026-10-08 review checkpoint: bounded P1 phrase admission (RP-403) and P2/P3 raw species
admission (RP-404) are CHANGES REQUIRED in the original Claude implementation. Codex's
corrections and executed evidence are in the latest log entries and need Claude's exact-range
review. P5's missing final-history AC7 refusal checks (RP-405) are also locally corrected;
the selected real-Postgres corruption population now passes, and designated review remains.
Existing checkboxes describe
implementation presence, not full-feature acceptance.

P6's adoption welcome also has a bounded local correction (RP-428): saved pet reads no longer
celebrate a new adoption, and receipt completion respects newer keyboard focus. The welcome
requires an actual local applied receipt and its matching authoritative pet. The latest owning
log records native regressions, the real-service journey and pending designated review; this
does not resolve PA7, shared attendance, owner copy or the full accessibility floor.

P4's Company replay identity decoder also has a bounded CHANGES REQUIRED finding (RP-406):
missing/null adoption coordinates were accepted as zero. The local correction, paired refusal/
explicit-zero controls and pending independent review are recorded in the latest log; this
does not approve P4's complete range or close its full acceptance population.

Numbering, ruled by landing order: the RFC's "Founder v22" maps to the **next free Founder version
= v23**, because Reputation Tree v1 already took v22. The RFC's "replay-inputs v7" maps to the next
free replay-inputs version. Cosmetic Shop then takes v24.

- [x] P1 Copy: the `companion` tone in both copy runtimes and the pipeline. The companion linter
  (PA8.2/AC13) rejects price, urgency, statistic and curtain tokens. A generated Go
  `copykeys.CompanionKeys()`. Candidate `pet.*` keys marked PENDING OWNER COPY.
- [x] P2 `pet_species` loader (Go/TS parity) with shared fixtures (AC1), plus draw vectors for the
  pet ID, temperament and palette (AC3).
- [x] P3 Bundle wiring: `pet_species` requires `pets` and `reputation_tree` (the scalar chain:
  Founder floor 23 implies v22's Reputation state).
- [x] P4 Founder v23 codec and catalog-aware validation (AC2); activation at the new-run boundary
  and New-Founder-forward (AC14).
- [x] P5 The `adopt_pet` intent: validation order, receipts, `pet_adopted.v1`, the migration, Go/TS
  Founder replay, idempotency (AC4, AC5, AC7, AC9, AC10), and the replay-inputs Founder carry (AC8).
- [x] P6 The snapshot `features.pet` projection (AC11) and the adoption surface with its visual
  contract (AC12, AC13 DOM).
- [x] P7 Economy isolation (AC15) and docs canon (AC17) implementation presence, not acceptance.
  RP-408 replaces the original no-adoption-in-the-harness argument with an actual paired pacing
  comparison and both runtimes' Company Exit vectors; executed results and review are in the log.
  AC16 is carried under the release floor
  (D-008/D-009/D-015 are unruled).

Tests for each flipped box, all inside the reviewed range `e5b7541a..HEAD`:

| Box | Commit | Tests |
|---|---|---|
| P1 | `bcccca24` | `verify-copy` fixtures |
| P2 | `9d01eef2` | `species_test.go`, `pet-species.test.ts` |
| P3 | `2f9878bc` | `pet_species_test.go`, `pet-species-bundle.test.ts` |
| P4 | `b29e70c0` | `pet_identity_state_test.go`, `pet_adoption_activation_test.go`, `pet-founder-state.test.ts` |
| P5 | `5ff37cc1` | `pet_adoption_test.go`, `pet_adoption_integration_test.go`, `pet-adoption-replay.test.ts` |
| P6 | `c380896a` + `21707c2b` | `gameui/pet_adoption_test.go`, `pet-adoption-arm.test.ts`, `pet-visual.test.ts`, `pet-adoption-browser.test.ts` |
