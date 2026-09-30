# RFC: Kernel History Guard Integrity

- **Status:** draft — not implementation authority
- **Author:** Codex (proposal for Marco's acceptance)
- **Created:** 2026-09-30
- **Design refs:** `design/06-tech.md` §Idle math (server-authoritative, replayable state)
- **Depends on:** [CI Baseline](scaffolding-and-ci.md), [Run Genesis & Replay](archive/run-genesis-and-replay.md) KV-1
- **Parent / amends:** Run Genesis & Replay KV-1's history-guard implementation; does not weaken its semantic version rule
- **Supersedes / superseded by:** —
- **Planning:** `planning/kernel-history-guard-integrity/` once accepted

## Summary

Keep KV-1's same-commit version bump for every change to transition, receipt, event, snapshot or
persisted-state bytes, while making its historical CI witness tell the truth. The current guard
stops at pushed `50a3a514`: presentation code was created under the registered
`client/src/minigame/` prefix without a kernel bump, then moved out in `79ff1aa1`. Separately,
the guard accepts text *requesting* a future independent review as if it were that review
(RP-131). A pushed-history rewrite and a cosmetic version bump are both prohibited. This RFC
proposes a narrow, append-only, independently reviewed **path-scope exception** for a proven
behavior-identical historical commit, separate from the existing correction for a real missed
semantic bump. No candidate commit is pre-approved by this draft.

## Motivation and scope

`kernel/affecting-paths.json` is deliberately conservative. A prefix can later contain UI-only
files; that fact does not make those files replay semantics. Conversely, dropping the prefix or
marking a whole commit harmless without examining every changed guarded path would hide future
semantic edits. The CI history walk must remain fail-closed and complete from the registry's
introduction. The existing `kernel/history-corrections.json` remains append-only and continues
to represent **real missed version signals**, including the now independently reviewed
`8add475`/`0.3.102` correction. This RFC does not authorize changing replay policy, version
numbers, API wire contracts, minigame behavior or historical Git objects.

## Proposed specification — subject to owner acceptance

1. Add an append-only `kernel/history-scope-exceptions.json` with schema version, sorted rows,
   exact 40-character offending commit SHA, sorted exact guarded path names, a substantive
   behavior-identity reason, and one review-log path. An exception applies only to the listed
   paths in that **one** historical commit. The guard still demands a same-commit bump or a
   separately valid semantic correction for every other affected path. No prefix/glob, commit
   range or release-wide waiver is allowed. A SHA binds the historical diff; an exception row
   cannot be mutated or removed later.
2. Before proposing an exception row, independently inspect the entire candidate commit and
   each registered changed path against KV-1's semantic categories. The review record must say
   why none of those paths changes transition, receipt, event, snapshot or persisted-state bytes,
   cite exact commit/diff range, and name executed relevant tests. The known `50a3a514`,
   `b92a05de` and `79ff1aa1` path-placement cases are audit **candidates**, not automatically
   qualifying exceptions. After each allowed entry, walk all remaining history and surface the
   next failure; do not claim the three candidates are exhaustive merely because the current
   guard stops at the first.
3. For both semantic corrections and path-scope exceptions, parse an actual independent
   review verdict, not a keyword search through a section. Require a standalone `##` verdict
   section with anchored metadata lines for `Review by`, `Recorded by`, exact reviewed range
   and `Decision: APPROVED`; quoted/backticked labels, future-tense instructions, a rejected
   decision and a range belonging to another section cannot satisfy it. The reviewer must be
   the project's designated other party for the implementation range. A machine can check
   structure and exact hashes; cross-party identity remains a human review obligation and may
   not be inferred from the commit author or recorder.
4. A scope exception needs its own designated review verdict on the exact historical commit
   and path set before it becomes valid. Do not repurpose the `8add475` semantic correction or
   cite a generic review of the surrounding feature. Exceptions and corrections are distinct
   append-only records, validated across every descendant commit in the history walk.
5. `make verify-kernel-version` stays in the blocking client/`verify-push` lane. Its output
   identifies the first unexcused commit and affected paths; it must never silently skip a
   shallow history, missing review log, malformed row or incomplete path list. Generated Go and
   TypeScript version constants must still equal `kernel/VERSION`.

## Deviations from design

None to game design or KV-1's semantic version rule. This adds a narrowly audited way to say a
registered historical *path* change was not a semantic change. It does not excuse a real
transition/encoding change from a kernel bump.

## Acceptance criteria

1. A seeded semantic edit in any registered path without a same-commit bump still fails the
   history gate; a scope exception omitting one changed guarded path also fails.
2. A fully reviewed, exact-path presentation-only historical exception passes; changing its
   commit, path set, reason or reviewer citation later fails. An unknown/unreviewed row fails.
3. A forged log section containing the literal tokens `**Review by:**`, `**Decision:**` and a
   target range as *instructions to review later* fails. Wrong range, `CHANGES REQUIRED` and
   verdict text outside the cited `##` section fail programmatically. The designated reviewer
   also rejects self/delegated-only review at the human cross-party gate; a syntactically valid
   label is not proof of reviewer identity. The genuine `8add475` verdict remains valid.
4. The full-history walk from a non-shallow checkout, all existing kernel negative fixtures,
   cold Go/TS replay/vector gates and the blocking `verify-client`/`verify-push` lanes pass on
   the actual corrected HEAD. Each proposed historical exception has its own executed
   cold/severing evidence and designated verdict; the implementation's own range also gets
   cross-party review. Passing the guard is not a product or release acceptance claim.

## Open owner choices before acceptance

- Authorize (or reject) an append-only **path-scope exception** class for proved
  behavior-identical pushed history. If rejected, name a different lawful solution that does
  not rewrite pushed objects, falsely bump replay version or remove the registered prefix.
- Rule whether a scope exception may cover a commit that also changes unguarded UI/API files
  when **all guarded paths** are proved presentation-only; the recommendation is yes, because
  KV-1 governs the semantic bytes, but the reviewed diff must still include the whole commit.
- Confirm the exact candidates only after the complete historical audit and designated review;
  this draft does not decide `50a3a514`, `b92a05de` or `79ff1aa1` on their author's claim.

## Changelog

- 2026-09-30: draft from RP-131 and the current red kernel-history gate; no implementation authority.
