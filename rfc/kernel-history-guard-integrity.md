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
`client/src/minigame/` prefix without a kernel bump, then moved out in `79ff1aa1`. The original
guard also accepted requests for future independent review as completed evidence (RP-131);
the existing-rule conformance repair landed in `611273b5`. Subsequent RP-648 additionally
binds an explicitly declared Reviewed range instead of letting an incidental anchor override
it. Neither repair authorizes historical path exceptions. A pushed-history rewrite and a
cosmetic version bump are both prohibited. This RFC
proposes a narrow, append-only, independently reviewed **path-scope exception** for a proven
behavior-identical committed change, separate from the existing correction for a real missed
semantic bump. The proposed class covers old path-placement mistakes and newly prepared
runtime refactors only when their unchanged behavior is actually established. It does not
approve the measured rate-traversal candidate, any historical case or another draft.

## Motivation and scope

`kernel/affecting-paths.json` is deliberately conservative. A prefix can later contain UI-only
files; that fact does not make those files replay semantics. Conversely, dropping the prefix or
marking a whole commit harmless without examining every changed guarded path would hide future
semantic edits. The CI history walk must remain fail-closed and complete from the registry's
introduction. The existing `kernel/history-corrections.json` remains append-only and continues
to represent **real missed version signals**, including the now independently reviewed
`8add475`/`0.3.102` correction. This RFC does not authorize changing replay policy, version
numbers, API wire contracts, minigame behavior or historical Git objects.

The rate-traversal experiment at `9bdf9c29..d6e092e8` demonstrates why future runtime-neutral
work needs an explicit rule as well: two complete300-run reports match exactly, existing
raw-bit/order/refusal controls pass, and the candidate remains unapplied. Those finite results
support review but do not themselves prove identity for every admitted input or qualify an
exception. Preserving actual computation order, validation and side effects remains part of
the independent eligibility argument. This proposal does not turn sampled success into
automatic exemption or authorize implementation before acceptance.

## Proposed specification — subject to owner acceptance

1. Add an append-only `kernel/history-scope-exceptions.json` with schema version, sorted rows,
   exact 40-character offending commit SHA, sorted exact guarded path names, a substantive
   behavior-identity reason, and one review-log path. An exception applies only to the listed
   paths in that **one** committed change. The guard still demands a same-commit bump or a
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
3. Parse actual independent review evidence, not incidental keywords. Existing semantic
   corrections retain their completed negative/remap records: a reviewed missed version signal
   is not approval of the original feature. Current conformance checks already exclude
   examples/requests and honor explicit reviewed scope. A **new path-scope exception** instead
   requires a standalone `##` verdict section with anchored metadata for `Review by`,
   `Recorded by`, exact reviewed range and `Decision: APPROVED`; quoted/backticked labels,
   future-tense instructions, a rejected decision and a range belonging to another section
   cannot approve that exception. The reviewer must be
   the project's designated other party for the implementation range. A machine can check
   structure and exact hashes; cross-party identity remains a human review obligation and may
   not be inferred from the commit author or recorder.
   The same section must explicitly contain `Scope-exception targets:` followed by one JSON
   array of unique full target SHAs, all covered by its reviewed range. Empty/duplicate fields,
   a target absent from that list, or a generic feature approval without this eligibility
   declaration cannot qualify a row. This field does not change existing semantic-correction
   verdict grammar or permit a recorder to retrofit approval into another party's old record.
4. A scope exception needs a designated review verdict on the exact candidate commit
   and path set before it becomes valid. Do not repurpose the `8add475` semantic correction or
   cite a generic review of the surrounding feature. Exceptions and corrections are distinct
   append-only records, validated across every descendant commit in the history walk.
   One existing coherent-batch verdict may qualify multiple explicitly listed commit/path
   sets; do not require a second review, separate predeclaration commit or per-edit ceremony.
5. `make verify-kernel-version` stays in the blocking client/`verify-push` lane. Its output
   identifies the first unexcused commit and affected paths; it must never silently skip a
   shallow history, missing review log, malformed row or incomplete path list. Generated Go and
   TypeScript version constants must still equal `kernel/VERSION`.

### Exact scope-register wire

The proposed file is one exact object with `schema_version: 1` and `exceptions: []`.
Each row has exactly `target_commit`, `paths`, `reason`, `review_log`; no optional fields.
Rows sort by unique `target_commit` (40 lowercase hexadecimal characters). `paths` is a
nonempty sorted unique array equal to the complete guarded-path set derived by the existing
walk for that target. Paths are exact repository-relative names, not prefixes/globs, absolute
paths, backslashes or dot/traversal segments. The target must actually have a guarded change
without a same-commit version signal; unknown objects, targets outside the walk and already
versioned commits refuse. A semantic miss cannot become neutral merely by fitting this shape.

`reason` is a non-whitespace string. `review_log` uses the existing safe planning Markdown
resolver and explicit archive mapping, never an exterior file/URL. Reject unknown/missing/null
fields, wrong types, duplicate raw object members, unsorted/duplicate targets or paths and
malformed/trailing JSON before using any row. Empty `exceptions` is valid and excuses nothing.
This register has no corrected-version field and cannot replace `history-corrections.json`.
Once introduced, every row is immutable and cannot be removed in any descendant, even if the
source edit is later reverted. New rows must pass the same target/eligibility checks.

### N1 — Qualification of a runtime-neutral refactor

Proposed eligibility, not accepted authority: for the same complete inputs, pinned artifacts,
clock/mode and frozen provider context, including valid and invalid cases, the refactor must
preserve KV-1-governed output bytes, admission/refusal behavior, arithmetic and observable
callback order, persisted effects and rollback.
Only computational cost/allocation may change. The review must account for every changed
guarded path on every parent edge inspected by the existing walk. Reordered floating-point
products, weaker validation, new data bounds, different errors or missing hooks are semantic
changes, not neutral performance work. Uncertainty means hold, not a false version bump.

Provide an exact old/new trace-preservation argument and execute the affected existing
tests, including meaningful invalid cases. A benchmark, canonical-string equality, a green
fixture subset or an implementer's claim alone cannot establish eligibility. Use additional
discrimination only where the critical oracle needs it, under RFC-0000's risk-based procedure.
Do not repurpose a descriptor-only approval for runtime code. API/owner-schema architecture,
other pending contracts and full feature/release gates keep their separate authority.

### N2 — Append-only delivery, without a self-referential commit hash

This flow becomes lawful only after this RFC's owner direction and normative review are
complete and its mechanism is implemented. Until then use an unapplied patch/overlay.

1. Prepare and test the exact candidate in a coherent batch. Preliminary patch review is
   useful, but cannot supply the committed target hash or an automatically valid exception.
2. Record the neutral candidate as an intentional local intermediate commit without a false
   version. Its guard remains red until it is qualified; it is not publishable/mergeable on
   that pending claim. Do not claim approval or rewrite it after a verdict cites its hash.
3. Obtain the designated other party's exact committed-diff eligibility verdict. That may be
   part of the normal batch review, with the exact guarded path set explicit. If identity is
   not established, keep it held or forward-revert; a real semantic change needs its actual
   accepted contract and versioned implementation, not a neutral label.
   A forward revert does not erase an earlier historical miss; an actual missed semantic
   bump still uses the existing reviewed version-correction mechanism.
4. In a descendant commit, append the approved target SHA/path set and verdict citation.
   Validate that the target is an ancestor, is a genuine unversioned guarded change, and
   that the verdict covers the complete guarded set. Never put a commit's own unknown hash
   inside itself. Uncommitted-row checks may target current HEAD; once committed, the
   introduction commit must strictly descend from that target.
5. Apply the current validated exception map to the whole existing history walk, including
   its earlier target; do not require the later row to exist inside that target's tree.
   Validate append-only row history from introduction onward. Surface every unrelated miss
   and require the finished affected gates before publication or merge.

This does not grant pending rows, aliases, whole-range exemptions, an implementer-controlled
approval, automatic historical amnesty or permission to push. Adoption delegates individual
qualification to the existing designated-review process; routine eligible batches do not
need another owner decision. A review cannot waive actual changed semantics.

## Deviations from design

None to game design or KV-1's semantic version rule. This adds a narrowly audited way to say a
registered committed *path* change was not a semantic change. It does not excuse a real
transition/encoding change from a kernel bump.

## Acceptance criteria

1. A seeded semantic edit in any registered path without a same-commit bump still fails the
   history gate; a scope exception omitting one changed guarded path also fails.
2. A fully reviewed, exact-path behavior-identical exception passes; changing its
   commit, path set, reason or reviewer citation later fails. An unknown/unreviewed row fails.
3. A forged log section containing the literal tokens `**Review by:**`, `**Decision:**` and a
   target range as *instructions to review later* fails. Wrong range and verdict text outside
   the cited `##` section fail programmatically. `CHANGES REQUIRED` cannot approve a new scope
   exception; it remains admissible completed defect evidence for the existing semantic
   correction class. The designated reviewer
   also rejects self/delegated-only review at the human cross-party gate; a syntactically valid
   label is not proof of reviewer identity. The genuine `8add475` verdict remains valid.
4. The full-history walk from a non-shallow checkout, all existing kernel negative fixtures,
   cold Go/TS replay/vector gates and the blocking `verify-client`/`verify-push` lanes pass on
   the actual corrected HEAD. Each proposed exception has affected cold checks, meaningful
   invalid-case evidence and designated eligibility; the mechanism's implementation range also gets
   cross-party review. Passing the guard is not a product or release acceptance claim.
5. A newly committed runtime-neutral candidate initially fails without its approved row;
   the descendant append-only row qualifies only its exact ancestor/path set, without a
   fake version or self-hash. Pending/rejected review, omitted guarded paths, a different
   target and mutation/removal of the row refuse. The independent reviewer must also
   reject a purported neutral refactor that changes multiplier order or invalid-input
   behavior; a syntactically valid reviewer name is not proof that such review happened.
6. Exact register-shape/duplicate/null/order/path/target negatives refuse; an empty register
   leaves all existing failures intact. A generic feature verdict without its explicit target
   eligibility list, wrong target list, or a conflicting/duplicate metadata field refuses.

## Open owner choices before acceptance

- Authorize (or reject) the append-only **path-scope exception** class for proved
  behavior-identical committed changes, including historical placement mistakes and new
  runtime refactors under N1/N2. The recommendation is yes, with exact other-party eligibility,
  not automatic approval of the current candidate. If rejected, choose a lawful alternative that does
  not rewrite pushed objects, falsely bump replay version or remove the registered prefix.
- Rule whether a scope exception may cover a commit that also changes unguarded UI/API files
  when **all guarded paths** are proved behavior-identical; the recommendation is yes, because
  KV-1 governs the semantic bytes, but the reviewed diff must still include the whole commit.
- Delegate per-candidate eligibility to the existing designated review under the accepted
  class, without a fresh owner ruling per routine batch. This draft does not decide
  `50a3a514`, `b92a05de`, `79ff1aa1` or the rate-traversal patch on their author's claim.

## Changelog

- 2026-09-30: draft from RP-131 and the current red kernel-history gate; no implementation authority.
- 2026-10-15: distinguish already-implemented review-evidence conformance from the unadopted
  exception policy. Preserve existing negative semantic-correction verdicts; proposed scope
  exceptions still require their own exact-range approval. No exception or policy adopted.
- 2026-10-15: specify proposed runtime-neutral eligibility and the two-commit descendant-row
  flow, distinguish finite experiment results from general identity, and reuse coherent-batch
  review rather than add per-edit owner approvals. Direction/adoption and normative review
  remain pending; no runtime code, exception or guard policy is implemented by this edit.
