# RFC: Commons Cohort Merge Capacity

- **Status:** implemented
- **Author:** Codex, implementing the owner's 2026-07-29 S4 ruling
- **Created:** 2026-07-29
- **Design refs:** `design/05-mmo.md`; `design/02-economy-balancing.md`
- **Amends:** `archive/commons-compact.md`
- **Planning:** `planning/commons-merge-capacity/`

## Summary

Make collapse merges population-driven and bounded: low membership triggers a whole-cohort merge,
Health does not, and no merge may fill the surviving cohort beyond 1.5 times its target.

## Specification

### D1 — Trigger and target

Only an additional open cohort with `member_count < cohort_merge_floor` is a merge source. Health
and Health band are not inputs. The oldest compatible open cohort remains the target.

### D2 — Whole-cohort capacity bound

The target's post-merge count must be at most `floor(3 * cohort_target_size / 2)`. A source either
moves in full or is skipped; members are never split and a cohort at the floor is never a source.
The operation remains one-way and idempotent under the existing advisory transaction lock.

## Acceptance criteria

1. A below-floor source that lands exactly on the 1.5× ceiling merges.
2. A below-floor source that would exceed it does not merge or move any assignment/sample.
3. A source exactly at the floor does not merge even when capacity is available.
4. Existing recomputation, replay, and population-invariance gates remain green.
5. Canonical Commons docs state the trigger, ceiling, and never-split rule.

## Deviations from design

None. This resolves the previously unspecified overfill boundary without introducing a Health
trigger or cohort splitting.

## Changelog

- 2026-07-29: owner ruled floor-only trigger, 1.5× ceiling, and never split; accepted and
  implementation started.
- 2026-07-29: implemented with overflow-safe capacity arithmetic and independently reviewed.
