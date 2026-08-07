# Permits & T3 gate implementation log

## 2026-08-07 — candidate assembly by Codex

The candidate economy/routes documents were assembled from the active Phase-0 bytes using the
owner-ruled insertion order. The RFC's human-form amounts were normalized to RFC-0001 canonical
Decimal bytes (`24` → `2.4e1`, `12` → `1.2e1`) after the real loader rejected the noncanonical
spellings; values and mechanics are unchanged.

Evidence shipped in the same range:

- shared Go and TypeScript candidate loading;
- route-resource and doctrine-route composition;
- neutral and non-neutral all-target multiplier behavior;
- 24-hour offline accrual and near-cap saturation;
- generated copy outputs and orphan record for the inactive cap-reason binding.

`make copy-check` and `make test` passed. This is a self-review record only; it does not satisfy
the designated independent review or authorize activation, archival, or the epoch-6 mint.

## 2026-08-07 — designated cross-party verdict: candidate implementation {7d9cb37} — NOT APPROVED (narrow)

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.
- **Reviewed range:** `7d9cb37^..7d9cb37` (the RFC's full implementation span to date; b14f82e is
  RFC-text-tier, not consumed).

**The candidate bytes are excellent — all verified clean:** all four manifest SHAs recomputed and
matching; P1/P2/P3 rows exact with the ruled insertion order (pure additive hunks, surrounding
bytes byte-identical to active); `2.4e1`/`1.2e1` confirmed pure RFC-0001 canonical normalization;
the PT-C4 owner copy literal byte-for-byte on the bound key; the PT-C1 no-active-artifact law HELD
(subject correctly not BALANCE-CHANGE:); orphan-stage copy landing consistent with FCE-C8 stage 1;
kernel honestly unchanged (guards green).

**BLOCKING:**
- **F1 (HIGH — repo gate red from this commit forward):** `verify-schema` (a `make verify`
  component) semantically validates every `balance/routes-testdata/valid/` fixture against the
  ACTIVE company-resource universe — the routes candidate's `company.permits` requirement fails
  (`/gates/1/requirement/1 references unknown company resource`). Green at 7d9cb37^, red at
  7d9cb37; every tracked commit until the mint fails the repository's own gate. The log claimed
  only copy-check + `make test` (which excludes verify-schema) — undisclosed by gap, not intent.
  Fix: pair candidate routes with their candidate economy in the sweep, or move the candidate out
  of the semantically-swept `valid/` directory (content unchanged → manifest HASHES survive; the
  manifest PATH cells get a reconciling edit).
- **F2 (MEDIUM, probe-proven):** the "near-cap saturation" case is non-discriminating — it lands
  exactly at the cap without needing the clamp (clamp neutered → case still passes; only the
  offline case catches it). One-line fix: start at 2.39995e1 or run 2 s so the unclamped result
  overshoots.
- **F3 (MEDIUM):** the PT-C3 "shared parity fixture" is not shared — the four behavior cases are
  Go-only; the TS client has NO hardcap clamping in its prediction paths and permits will be the
  first reachable cap. Cross-runtime saturation parity must be proven (or the changelog's "shared
  … hardcap proofs" claim retracted and the TS gap filed as its own finding).
- **F4 (MEDIUM):** the PT-C5 activation fixtures (cross-epoch Exit init, fresh-genesis, old-run
  replay) are ABSENT and untracked in plan.md; and this is the repo's FIRST two-requirement gate —
  multi-requirement crossing debit/rejection/replay is untested anywhere. Even if deferred to the
  mint range by design, plan.md must carry the unchecked items; AC1/AC3 were owner-amended to
  include them.

**Verdict: NOT APPROVED pending F1 (urgent — the repo gate) + F2–F4; re-review will be narrow.
Owner ratification of the manifest hashes is HELD until the fix lands (hashes need not change if
files move rather than change content).**
