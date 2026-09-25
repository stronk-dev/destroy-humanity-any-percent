# Terminal Typer log

## 2026-09-25 — Predeclaration (Claude)

**Implemented by:** Claude, per the owner's 2026-09-25 acceptance and implement direction.
**Review:** Codex designated cross-party review required before any archival. Nothing here is
self-approved.

Batches B1–B7 as in `plan.md`. Each batch lands with failing-first tests, a severing probe per
gate, and cold `-count=1` runs. Copy: mechanical keys ship with implementer-drafted candidate text
in `copy/catalog/typer-candidate.json`. That text is owner-adoption material, not ruled copy.
Prompt `text` rows in the fixture are PROVISIONAL placeholders (OD-9), not for production mint.

## 2026-09-25 — B1: content + pure engine (Claude)

**Implemented by:** Claude. **Review:** awaiting Codex designated review.

Landed:
- `server/typer` and `client/src/typer`, registered in `kernel/affecting-paths.json`; kernel bumped
  0.3.102 → 0.3.103 in this commit.
- `ApplyInput.ServerTimeMs` (TT-PA1 item 2), which Pitch ignores.
- Fixture `balance/testdata/typer-v1.json` with 15 placeholder prompts.
- Candidate copy `copy/catalog/typer-candidate.json`, which is owner-adoption material.
- The shared corpus, `make typer-corpus`/`typer-corpus-check`, and `docs/minigame-terminal-typer.md`.

Evidence (cold, `-count=1`):
- `./typer ./minigame` pass. `client/test/typer-content-gate.test.ts` passes 6/6.
- Severing probes, each turned red (Go/TS):
  - no `max` on the clock (TestTyperClockNeverRewinds);
  - deadline `>=` (content gate);
  - the run substream used directly, as a compiling mutant; a first, non-compiling attempt is
    discarded (Go content gate; TS 2 failed);
  - clears counted as clean (TestTyperCleanLineAccounting and the gate);
  - unsorted nested `last_submission` fields (TestTyperSnapshotsAreRegistryCanonical);
  - TS case folding;
  - TS truncation instead of `line_too_long`;
  - **TS NFKC.** It first SURVIVED, because no vector used a character that NFKC folds and OD-7
    does not map. A fullwidth look-alike vector was added to Go, TS and the corpus. Rerun: TS
    2 failed.

Implementation decisions, all within ruled text:
- **Validation order.** `invalid_text` is validated at command decode, which is phase-independent
  like Pitch's `hand_too_large`. `illegal_phase` precedes `line_too_long`, because the length
  bound needs the content.
- **Terminal handling.** `end_run` and `timed_out` retain `last_submission`.
- **`ValidateResult`** is content-free and checks structural bounds: facts ≥ 0,
  `clean ≤ cleared`, `assisted ∈ {0,1}`. The engine enforces `run_length` and the hardcaps.

**DESIGN-GAP (minor, recorded):** the RFC does not say whether U+FFFD submitted legitimately is a
miss or `invalid_text`. Both runtimes treat it as `invalid_text`, because Go's decoder cannot tell
it apart from invalid UTF-8.
