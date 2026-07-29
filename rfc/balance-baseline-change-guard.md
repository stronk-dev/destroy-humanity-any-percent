# RFC: Balance Baseline Change Guard Hardening

- **Status:** implementing
- **Author:** Codex, from independently verified findings A1 and D4
- **Created:** 2026-07-29
- **Design refs:** `design/07-roadmap.md` Phase 0; `design/02-economy-balancing.md §11`
- **Research / reproducer:** `planning/archived-four-review/log.md` A1 and D4
- **Amends:** `archive/balance-harness-foundation.md` D6-D7
- **Planning:** `planning/balance-baseline-change-guard/`

## Summary

Make the committed pacing baseline's review protocol an enforced repository invariant. The current
guard inspects only HEAD, silently accepts truncated history, and permits a `BALANCE-CHANGE:`
baseline commit to carry arbitrary code. CI compounds both defects with a two-commit checkout.

## Specification

### D1 — Full reachable-history validation

`ValidateRepositoryBaselineChange` must inspect every non-initial commit reachable from HEAD that
changes `testdata/harness/pacing-baseline.json`, regardless of later cover commits. A shallow
repository is an error, never an exemption. Missing parents, unreadable history, ambiguous git
output, or inability to inspect a baseline commit fail loudly.

The oldest commit that adds the initial baseline is the sole bootstrap exception because it
necessarily landed with the harness implementation. Every later baseline commit is validated.

### D2 — Generated-artifact-only commits

A later baseline commit:

1. has a subject beginning exactly `BALANCE-CHANGE:`;
2. changes `testdata/harness/pacing-baseline.json`;
3. may otherwise change only `testdata/harness/golden-seed.json`.

Catalogs, scenarios, docs, tests, source code, or any other path in that commit are rejected. The
inputs and the generated review artifacts therefore cannot hide in one diff, and the label cannot
authorize unrelated code.

### D3 — Changed inputs precede the artifact commit

For each later baseline commit, compare the previous baseline commit to the candidate's first
parent. That interval must change at least one accepted harness input:
`balance/catalogs/**` or `testdata/harness/scenarios/**`. The baseline artifact alone, or a second
rewrite without intervening input, fails. Input-domain expansion remains owned by its feature RFC;
this amendment does not silently absorb Commons into `constants_hash` (review finding A3).

### D4 — Clean artifact state and complete CI history

Tracked or untracked worktree changes at either generated artifact path make repository validation
fail because no commit subject or review boundary exists yet. The server CI checkout uses
`fetch-depth: 0`; local and hosted validation execute the same Go guard without GitHub metadata.

## Acceptance criteria

1. A bad baseline commit followed by an ordinary cover commit still fails.
2. A shallow clone fails explicitly instead of treating one visible baseline revision as initial.
3. A `BALANCE-CHANGE:` commit containing a code, catalog, scenario, docs, or test path fails.
4. An input commit followed by an artifact-only `BALANCE-CHANGE:` commit passes.
5. A baseline-only rewrite, same-commit input+baseline change, wrong subject, and dirty artifact
   worktree each fail; the repository's historical initial baseline remains valid.
6. CI fetches complete server history and the full verification matrix passes.

## Deviations from design

None. This enforces the already accepted separate-commit review protocol.

## Open questions

None. Commons hash/input coverage remains explicitly assigned to A3.

## Changelog

- 2026-07-29: drafted, accepted, and implementation started in the owner-directed A1+D4 order.
