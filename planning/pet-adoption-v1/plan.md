# Pet Adoption v1 implementation plan

RFC: `rfc/pet-adoption-v1.md` (accepted 2026-09-25; every OD at its recommended default).
Original implementation by: Claude. P1–P7 now have complete original-batch Codex review
dispositions: CHANGES REQUIRED. Corrective descendants need their own designated review;
the original verdicts do not approve them. No archival or production mint (OD-4/PA8.6).

2026-10-15 fresh independent Codex corrective review APPROVES RP-403
(`9c3b49c1^..9c3b49c1`), RP-404 (`9ba37f9b^..9ba37f9b`), RP-406
(`7b4bc001^..7b4bc001`) and RP-463 (`ff5275fc^..ff5275fc`) under Marco's
delegation. The owning log records complete exact ranges, actual cross-runtime/SQL
checks, provenance and limits. Other corrections and feature acceptance remain separate.

2026-10-15 further fresh independent corrective review APPROVES complete RP-405 history,
RP-408 paired economy-isolation and RP-428 receipt-bound welcome ranges. The owning log
records exact hashes and actual Postgres/history, all97 paired seeds, native/Linux UI and
both focused real-service driver executions. RP-392 is CHANGES REQUIRED: review finds
RP-669, collapse/reopen can reclaim a newer focus choice that disappears before tick.
The added regression reproduces4 old native failures; the bounded listener correction at874d8d17
now has fresh independent APPROVED for completea45f5478..874d8d17. Reviewer actually executes
244 native/366 Linux affected host cases with explicit performance exclusions, types/build/
boundaries. RP-669 closes; RP-392's historical verdict remains, not original feature/copy/AT acceptance.

2026-10-08 review checkpoint: complete original P1 (RP-403) and P2/P3 (RP-404) reviews
retain CHANGES REQUIRED for phrase and raw species admission. Codex's
corrections now have independent approval and executed evidence in the latest log entry.
Complete original P5 review is still CHANGES REQUIRED for the missing final-history
AC7 refusal checks (RP-405). Those are locally corrected; the selected real-Postgres
corruption population passes, and its complete correction now has independent approval.
Existing checkboxes describe
implementation presence, not full-feature acceptance.

P6's adoption welcome also has a bounded local correction (RP-428): saved pet reads no longer
celebrate a new adoption, and receipt completion respects newer keyboard focus. The welcome
requires an actual local applied receipt and its matching authoritative pet. The latest owning
log records native regressions, the real-service journey and independent RP-428 approval; this
does not resolve PA7, shared attendance, owner copy or the full accessibility floor.
Both complete original P6 commits now have CHANGES REQUIRED dispositions: PA7/C2's
author-owned wire conflict (RP-132), and the keyboard/welcome defects (RP-392/RP-428).
Complete original P7 review also retains CHANGES REQUIRED for RP-408's missing paired
economy comparison. Its later test supplement now has independent approval, including all97 pairs.

P4's complete original range `b29e70c0^..b29e70c0` now has a designated Codex review:
CHANGES REQUIRED. In addition to RP-406's missing/null Company carry coordinates, RP-463
finds case-aliased saved identity keys accepted by Go but refused by TypeScript. Both are
locally corrected with refusal and valid-encoding controls; affected real-Postgres checks
pass. Both corrections now have independent exact-range approval in the latest log. This original verdict
does not approve corrected descendants, other original batches or full Pet acceptance.

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

Implementation-presence tests originally recorded across `e5b7541a..235b888a`:
this table is provenance, not designated approval of that range or its descendants.

| Box | Commit | Tests |
|---|---|---|
| P1 | `bcccca24` | `verify-copy` fixtures |
| P2 | `9d01eef2` | `species_test.go`, `pet-species.test.ts` |
| P3 | `2f9878bc` | `pet_species_test.go`, `pet-species-bundle.test.ts` |
| P4 | `b29e70c0` | `pet_identity_state_test.go`, `pet_adoption_activation_test.go`, `pet-founder-state.test.ts` |
| P5 | `5ff37cc1` | `pet_adoption_test.go`, `pet_adoption_integration_test.go`, `pet-adoption-replay.test.ts` |
| P6 | `c380896a` + `21707c2b` | `gameui/pet_adoption_test.go`, `pet-adoption-arm.test.ts`, `pet-visual.test.ts`, `pet-adoption-browser.test.ts` |
