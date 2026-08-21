# Predeclared row-level test oracle audit

Predeclared: 2026-08-21. Product coordinate: `190a4fa`. Frozen source populations:

- all 592 top-level Go `Test*`/`Fuzz*` functions in the 151 files reconciled by
  `server-test-file-inventory.tsv`; and
- all 43 tracked TypeScript sources in `client-test-artifact-inventory.tsv`, expanded to every
  static `it`/`test` declaration. A tracked helper/type-contract file with no declaration receives
  one explicit helper unit rather than disappearing.

Ignored screenshots, generated local captures, production source, external dependencies, and
candidate tests absent from these frozen inventories are excluded. Make/CI command topology remains
owned by its existing 72-target/seven-job ledgers; this pass inspects the row-level test oracles
those lanes invoke.

## Question

For each declared test/fuzz oracle, what subject and boundary can it falsify, what fixture/data it
actually consumes, which dependency/skip/cache/guard conditions can remove it from a green run,
what assertion and negative-control signals exist, whether prior cold execution reached it, and
what exact larger capability or acceptance claim it cannot support?

## Oracle units

1. One server unit per top-level Go `Test*` or `Fuzz*` function. Nested `t.Run` cases, table rows,
   loop expansions, corpus seeds, assertions, and dependency skips are recorded as signals on that
   owner rather than silently multiplied into independent top-level witnesses.
2. One client unit per static Vitest/Playwright `it` or `test` declaration, including `.each`,
   `.skip`, `.only`, `.fails`, and conditional declarations. A parameterized declaration is one
   source oracle with its expansion signal and cannot claim a case denominator unless the executed
   runner output exposes it.
3. One `helper_not_oracle` unit for each frozen tracked client source with no `it`/`test`
   declaration. Helpers, compile-time negative fixtures, and browser guards remain visible but do
   not become passing tests by existence.

Stable identity is runtime + repository-relative file + declaration name + source byte offset. The
ledger also records a hash of the exact function/callback body so semantic changes cannot retain a
stale row unnoticed.

## Output schema and verdicts

`test-oracle-row-ledger.tsv` contains:

```text
oracle_id, runtime, file, declaration, source_line, body_sha256, subject,
execution_lane, fixture_or_data, dependency_skip_guard, assertion_oracle,
negative_control, verdict, authority_route, evidence_limit
```

Closed verdict vocabulary:

- `integrated_discriminating` — exact real dependency/current-data/default workflow plus an oracle
  demonstrated to fail when the claimed integrated behavior is severed;
- `bounded_discriminating` — a primitive/package/protocol boundary has assertions and a demonstrated
  relevant failure/negative case, with no larger integration credit;
- `positive_only` — success behavior is asserted but no relevant negative/mutation evidence proves
  the oracle can reject the claimed defect;
- `fixture_or_mock_only` — the oracle may discriminate its fixture/helper but does not consume the
  real dependency/current data/player workflow required by the larger claim;
- `dependency_conditional` — a real dependency test exists but ordinary default execution may skip
  it; only the named Compose/hosted lane can claim it;
- `non_discriminating` — a relevant subject/outcome can be removed or broken while this oracle
  stays green, or its assertion structurally cannot distinguish the defect;
- `invalid_or_guarded` — timeout, guard exhaustion, truncation, architecture exclusion, invalid
  selector, or similar condition prevents a valid result;
- `helper_not_oracle` — source supports tests/typechecking/error capture but is not an independently
  executed oracle; and
- `review_unresolved` — source structure is bounded but semantic discrimination cannot yet be
  established. This is an explicit incomplete result, never a pass.

Test names, assertion count, coverage, fixture size, package green status, and archived RFC state do
not promote a row. A single top-level function may exercise many cases while still proving only one
bounded subject.

## Evidence lanes

For every unit record:

1. exact body hash, declaration line/name, subtest/parameterization count signals, and assertion
   mechanisms (`t.Fatal*`, equality/matcher, thrown-error, snapshot, property/fuzz invariant);
2. production current artifact, candidate/historical/testdata fixture, mock/fake, real Postgres,
   actual socket/browser, or pure in-process input;
3. dependency skip/feature guard, environment requirement, architecture condition, timeout/deadline,
   early return, truncation/iteration cap, and cache-sensitive lane;
4. success oracle and exact relevant rejection/mutation/failure signal, distinguishing generic
   malformed-input tests from a negative capable of firing the acceptance claim;
5. latest cold/package/Compose/browser/hosted execution evidence already recorded at `190a4fa`,
   including invalid runs and visible skipped counts; and
6. capability/RFC criterion/RP/research/accepted successor route for every limitation.

## Controls and failure conditions

- The Go files/functions must reconcile exactly to 151/592 and the client source files exactly to
  43; no zero-declaration helper may disappear.
- Seeded dropped, duplicate, helper-omission, body-hash drift, and missing-route rows must fail.
- A server integration function with `TEST_DATABASE_URL` skip cannot be called unconditional from
  the ordinary host lane.
- A mock/fixture-only row cannot be promoted to integrated merely because production types or
  current IDs appear in its body.
- A positive-only row cannot be promoted without a demonstrated relevant failing case.
- The already fired Game UI cap/drain/resync oracle, API unregistered-route/client-lint oracles,
  warm-cache/harness invalidity, 320 px/focus gaps, and dependency-skipped Postgres population are
  mandatory controls; a classification that turns those green is invalid.
- Aggregate runner totals do not replace static declaration reconciliation; static declarations do
  not fabricate dynamic case counts the runner did not expose.

Manual second pass: every integrated/bounded promotion, every browser/composed/Postgres/socket test,
every conditional/invalid/non-discriminating row, every helper, every test named by an active or
archived acceptance criterion, every row using deploy-current content, and every family with mixed
verdicts. Final summaries must reconcile with `test-evidence-audit.md`, lifecycle dossiers,
`active-acceptance-ledger.tsv`, the capability/content ledgers, queues, RP ledgers, and review draft.

## Authority limit

This wave may add audit-only extractors/ledgers, execute existing tests cold, use temporary restored
negative probes already authorized by accepted scopes, file/refine RP routes, and reconcile planning
summaries. It may not change product tests or acceptance bounds, fix discovered defects, edit
design/RFC/canonical product text, flip implementation-plan boxes, archive work, or treat this
self-audit as the designated cross-party review.
