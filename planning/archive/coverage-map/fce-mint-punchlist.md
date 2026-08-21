# First Content Epoch — mint punch-list (audit 2026-08-07)

> **FROZEN HISTORICAL SNAPSHOT — NONCANONICAL.** Preserved as evidence of the 2026-08-07
> mint sequence. It assigns no current work. Current authority lives in canonical production data,
> archived RFC records and `planning/platform-alignment/`.

Source: the six-foundation archival audit (designated-review coverage + plan-box state), run while
Codex executed The Pitch archival. Feeds FCE5.1 (reformulated same day — artifact-scoped, no longer
circular for mint-blocked foundations).

## The mint's actual critical path

| Foundation | Distance | What remains for the MINT (per reformulated FCE5.1) |
|---|---|---|
| Doctrine, Fiscal, Soul, (Pitch) | ✅ done / closing | Archived (Pitch archival executing now). |
| Soul Recovery | scheduler delta | F1/F2 remediation + narrow re-review (verdict 2026-08-07). Artifact rows already verified. |
| **Meters + Achievements** | **owner content** | Range-union sliver `04e1905` (designated review LAUNCHED 2026-08-07) + **owner-supplied literal rows**: meter band/initial/rate/input data, achievement rows. Their final AC IS the mint — they archive WITH epoch 6. |
| **Pet Care** | **owner content** | Owner-supplied species/temperament + care/decay/trust/FSM rows. AC3 (combat cross-verify) is combat-gated and EXCLUDED from the mint gate. Stale parent plan boxes need reconciling (sub-work complete + approved). |
| Minigame Platform | reviewed for mint | Artifact behavior fully designated-covered (chain through 06bf0f3 + the Pitch verdicts). AC1 (gameserver composition) is genuinely unbuilt — grep-verified zero minigame refs in server/gameserver/ — but it's the PLAYABILITY path (MA needs it), not the mint. AC6 combat-gated, excluded. |
| Founder Attendance | outside gate | Not an artifact contributor. Nearest to standalone archival: rollback/retention Postgres residue + docs decision (founder-transitions.md section vs dedicated page) + move. |
| API Foundation | outside gate | Not an artifact contributor. BUT: ~half unbuilt, owner rulings C18–C20 outstanding, and ZERO designated coverage (Darwin-only, self-recorded) — the largest review debt in the repo; on the MA/playability path, not the mint path. |

## Newly identified owner/Claude work (the real mint blockers)
1. Draft the production meter/achievement/pet content rows (balance data, design/02-grounded,
   provisional-bytes discipline same as Pitch) — Claude drafts, content gates + review, then FCE
   promotes. THIS is the long pole for the mint now that reviews are nearly closed.
2. Rule API Foundation C18–C20 (blocks Codex resuming API work; MA acceptance depends on API).

## Codex-side hygiene (bundle into next handoffs, not urgent)
- docs/README.md index missing 5 existing pages (meters, achievements, pet-care, api-foundation,
  founder-transitions); minigame-platform indexed with honest caveat.
- Stale plan boxes: minigame plan prose still says resolve composer "blocked on C37-C40" (ruled +
  approved 2026-08-05); pet-care parent boxes unchecked with all sub-work done.
- rfc/README.md Founder Attendance row understates ruled set (A1–A5 + B1–B3).
- rfc/README.md MA row parent cell still says "Minigame Platform (archived)" — stale.
