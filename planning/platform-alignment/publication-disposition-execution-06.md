# D-002 Publication Disposition Execution — Bounded Revisions B

**Date:** 2026-08-21
**Authority:** D-002 owner ruling; `publication-rights-batch-03.md`
**Scope:** revise and expose two specifically approved Class-C minigame research dossiers
**Product/design/RFC authority:** none

## Executed population

| File | Required correction | Result | SHA-256 after revision |
|---|---|---|---|
| `design/research/absorption-arena.md` | Separate evidence from arena/economy/tier/multiplier/name/copy proposals; soften cost/bot certainty; narrow legal claims and stale release placement. | Added an authority banner, reduced source quotations, reframed the single-human shape as an unmeasured prototype, removed 1.0 placement, labelled constants/copy/names non-adopted, and replaced the clearance-style legal matrix with issue-spotting. | `6c8978d3fc01b8d62aa7028492af473abda50d12c9dbc151f85fde37868b85f7` |
| `design/research/board-game-mechanics.md` | Make rankings hypotheses, remove unfetched specifics, label roster/copy/economy calls as proposals, correct stale shipped claims and narrow legal certainty. | Added a dossier-wide authority boundary; removed unfetched exemplar specifics from the factual body; rewrote mappings as non-adopted; labelled the ranking unmeasured; corrected the Ship-It HEAD claim; and narrowed legal language. | `7a51e941a9143c18bdf8458137ceef096d5677bea7a82139bb894c358a785e84` |

## Boundary and checks

- `.gitignore` gains only two file-specific exceptions; all other ignored research remains
  ignored.
- The taxonomy and comparative findings remain research. No minigame, bot, economy hook, formula,
  tier, name, copy, release placement or legal conclusion is adopted.
- Unfetched named examples remain visible only as omitted/verification records or taxonomy labels,
  not as factual mechanical support.
- `git diff --check` passes for this execution scope.

One bounded revision remains: `_completeness-sweep.md`. No push, publication, deployment, product
behavior, design intent or RFC contract is authorized by this record.
