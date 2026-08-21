# D-002 Publication Disposition Execution — Bounded Revisions A

**Date:** 2026-08-21
**Authority:** D-002 owner ruling; `publication-rights-batch-02.md`
**Scope:** revise and expose three specifically approved Class-C research dossiers
**Product/design/RFC authority:** none

## Executed population

| File | Required correction | Result | SHA-256 after revision |
|---|---|---|---|
| `design/research/tech-stack.md` | Separate the July 2026 option study from current architecture; identify rejected numeric and bot recommendations. | Added a prominent historical/superseded banner, current authority links, and local labels for `break_eternity.js`, bit-for-bit parity and rubber-banding. | `1d2ec46cd95397d693e2fb953212ab9e129f5ef2487a3e3677690ebf23ae7d57` |
| `design/research/mobile-pwa.md` | Freeze repository claims to a commit/date and label copy/mechanics/routing as proposals. | Pinned repository observations to `ad06e03b7e84079e9d66c7a8385d3573c8692c8a` on 2026-08-08; labelled copy, constants, mechanics and routing non-adopted; retained the verify-before-shipping list. | `a6fdf12d3f4c9e9268e5aef0e8e1b34337772b4d02428061a58681ee61f9d8cc` |
| `design/research/tier-relevance.md` | Reduce quotation dependence, bind durable sources and distinguish anecdotes. | Replaced the substantive wiki/guide/Reddit excerpts with paraphrase, added direct Pecorella/Paper Pilot identifiers, and made the identifier-less archive anecdotes non-evidentiary. | `191c31393f1bcf1e0c69dc054ab9bab2e48547e8eda65705e6020b3cae538c1d` |

## Boundary and checks

- `.gitignore` gains only three file-specific exceptions; all other ignored research remains
  ignored.
- The revisions preserve research findings but do not adopt architecture, copy, mechanics,
  constants, legal claims, release scope or work authorization.
- `rg` finds no remaining verbatim excerpt in `tier-relevance.md`; quoted labels are ordinary
  terminology only.
- `git diff --check` passes for this execution scope.

Three bounded revisions remain: `absorption-arena.md`, `board-game-mechanics.md`, and
`_completeness-sweep.md`. No push, publication, deployment, product behavior, design intent or RFC
contract is authorized by this record.
