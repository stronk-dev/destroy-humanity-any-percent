# Headcount Allocation — implementation log

## 2026-10-08 — Source-independent S2 arithmetic

Baseline `dbd8e274`. Previous goal turn was progress (committed Clout help with
executed browser/real-service proof). This batch constructs accepted S2's pure
Go/TS arithmetic without choosing the explicitly held OD-1 seat source.

`server/headcount` and `client/src/headcount` project complete assignments into
producing/diverted heads, idle converters and unassigned seats. Safe-integer
bounds, complete keys and the converter graph fail closed; inputs and results
are independent. No clock, Decimal, catalog/save/version, rate, command or UI
integration is added. Only tests import the new modules. Before authoritative
wiring, register both prefixes and bump kernel identity with the actual semantic
integration; existing watched paths and kernel175 remain untouched here.

Evidence: twelve shared literal vectors, 7803 small-domain allocations per runtime
against independent arbitrary-precision models, maximum-domain cases, independent
conversion pairs, role order, fresh results and22 shared-condition refusals
(TS also tests fractions/NaN/Infinity). Final root commands/results:

- `make test-go GO_PACKAGES='./headcount' GO_TEST_FLAGS='-count=1'`: PASS,0.308s.
- `make test-go-ci CI_TEST_PACKAGES='./headcount' CI_TEST_FLAGS='-v'`: PASS (50359),
  actual Linux-amd64 container,0.102s. No DB operation is claimed by this pure package.
- `client/node_modules/.bin/vitest run --root client test/headcount-allocation.test.ts`:
  PASS42. Existing automatic Node/browser collection consumes the new test file.
- `make test-browser-focused BROWSER_TEST_FLAGS='test/headcount-allocation.test.ts
  --project=chromium --project=webkit'`: PASS84 (77778).
- `make typecheck vet GO_PACKAGES='./headcount'`: PASS (6621),zero Svelte errors/warnings.
- Firefox attempt48146: RED,zero tests;60s session-connect failure and180s launch
  timeout, with plugin-container sandbox-extension/graphics errors. Handle is terminal;
  no retry, relaxed budget, all-engine or hosted CI pass inferred.

RP-413: S8's3.1e15 converter/9e15 source assignment totals12.1e15, outside S4's
safe seat budget. It rejects in both kernels. A separate legal3.1e15/5.9e15 vector
passes. Direct Node observation confirms naive min still returns the exact9e15
despite an unsafe9.3e15 capacity; the mandated failing mutant is not established.
Do not claim AC3/AC5 complete or edit the author's accepted text around the result.

RP-414: Tier2 H1–H5 and Headcount S0–S5 disagree on catalog/state/commands and
output economy, not only seats. Both accepted bodies are unchanged. An async
seat-source question was sent to Marco; no answer/adoption is recorded. Reconcile
the whole boundary before live integration; do not treat a budget choice alone
as approval of every schema or output mechanic. Version literals are also occupied.

Review by: Codex (implementer first filter). Recorded by: Codex. Full new range
`dbd8e274` exclusive through this containing source/test/docs/record commit needs
designated cross-party review. No feature checkbox, archival, mint, push, complete
CI/release claim or narrowing of the nine-tier objective. Next: resolve the held
integration contract; other accepted lanes may continue in parallel with that hold.

## 2026-10-13 — Make the full Headcount integration choice explicit

At `769b26a2`, reread the complete Headcount contract and Tier 2 H/P boundaries. RP-414 is
not resolved by a hiring answer: the siblings also disagree on direct resource production
versus target multipliers, accumulating versus display-only Decks, catalog/save/command/event
ownership and harness masks/gates. The owning plan now pairs the two gameplay questions with
those exact reconciliation sites. Questions are presented to Marco, not answered/adopted.
Recommend the standalone engine's separate-hiring/direct-production pair; do not infer that
recommendation from the earlier batch acceptance or edit the ruling author's bodies.

Read-only source checks confirm current economy schema5/axis_stack, Company19/Clout attainment,
Founder25 and replay12. Old v5/v19/v9 numbers cannot be reused. Keep RP-413 and all pacing,
content, copy, full integration and review gates. No schema, mechanic, balance, version, test
or feature status changes in this editorial clarification; no software suite was run for it.
Links/sections/source claims and final diff are checked.

Separately, the actual declared integration environment is still unavailable: owned read-only
Docker probe29326 gives no daemon reply while live. `pgrep` identifies only its75273 client;
explicit SIGINT ends that probe with130 and client metadata only. No daemon restart, cleanup,
container mutation, secret access or SQL pass is inferred. Platform public solo/async fixture
construction also retains its already-recorded closed-wire hold; no test-only bypass is created.

Recorded by: Codex. Editorial decision preparation, not an owner ruling, normative amendment,
designated approval or release call. Next: record actual gameplay answers and obtain the
author's coherent body reconciliation before the complete Headcount integration chain.

## 2026-10-13 — Resolve RP-413's empirical question, not its ruling

At `bd46e71e`, the previous goal work made concrete Deployment authority/native-process
progress; the intervening answer clarified owner versus engineering responsibility but did
not itself change product state. Continue the already-started Headcount oracle investigation.

`overflow-oracle-audit.md` records the bounded question, exact mutation boundaries, primary
Number semantics, derivation and reproduction. The illegal S8 example remains rejected.
Within admitted integer counts, TS multiply-before-clamp cannot yield the specified overflow
output fault; samples alone were not used as the universal proof. Signed-zero representations
are outside that overflow claim. The existing independent exact-model tests discriminate
other real faults: ignoring `k` produces eight Node and sixteen Chromium/WebKit failures,
including the actual k3 literal. Go's analogous int64 capacity mutation fails cold on an
admitted maximum-domain assignment with negative diversion. No assertions were weakened.

Executed: naive TS comparison Node84/Chromium-WebKit168 pass; wrong-ratio comparison exits
1/Make2 with8/16 mutant failures and all actual implementation checks passing; Go overlay
`TestExactMaximumDomain -count=1` exitsMake2. Actual full Headcount Go passes0.121s/vet0;
after deleting only our temporary clone files, Node42 and native Chromium/WebKit84 pass
(0.247s/1.43s). Root commands, outcomes and limitations are in the audit. Initial browser
EPERM ran no tests; lost tool results are excluded, and matching test processes were absent
before replacement runs. No production file was mutated or leftover mutant added to CI.

RP-413 stays OPEN for its author's S8/AC3 reconciliation; its next action is now explicit,
not another hunt for an impossible overflow output failure. Accepted body, guarded algorithm,
fixtures, versions, balance, hiring/output decisions and feature checkboxes are untouched.
No Firefox/Linux/hosted/full CI, live integration, AC3/AC5 completion or release claim.

Review by: Codex (research/implementer first filter). Recorded by: Codex. This research-only
range after `bd46e71e` through the containing checkpoint needs independent review before it
can support a normative amendment; earlier implementation review debt remains separate.
Next: author reconciles the criterion against this evidence; continue accepted independent
construction without inferring the held gameplay model or making a release call. No push.

## 2026-10-14 — Refuse non-string IDs before allocation ordering (RP-571)

Resume clean `4c216b69`; previous turn completes the API reference-metadata repair. Re-read
Headcount S1–S3 and the owning plan: S2 is already overflow-safe; RP-413 concerns the author's
invalid required test, not an unfixed multiply-before-clamp path. Re-present the separate hiring
and output choices through the async question interface; delivery/preselection is not adoption.

Under S1's mechanical string IDs/S2's validated inputs, actual TypeScript Project admits false,
true, null, undefined, single/nested arrays and boxed strings as coerced names. A two-role input
also executes an object's coercion hook in sorting. Eight old-source refusals fail; 48 existing/
literal-string controls pass. File RP-571 immediately. The trailing-line-ending hypothesis is
falsified: all five native regex probes reject; leave that regex unchanged. Prevalidate primitive
string/grammar before ordering/lookup, keep duplicate/key checks and all arithmetic unchanged.

Final Node passes56; native91498 executes112 in Chromium/WebKit, terminal0/1.56s. Cold Go
`make test-go GO_PACKAGES='./headcount' GO_TEST_FLAGS='-count=1'` passes0.182s; same-package vet
passes. Typecheck66597 reports zero errors/warnings. Input-preservation, literal primitive-name/
constructor IDs, no coercion before rejection and all existing exact-model/boundary cases pass.
Source import search finds no live/replay consumer; no watched path or kernel identity change.
No shipped rendering, rate, save, catalog, intent, seat/output mechanics or player copy change.

Raw old `.cache/headcount-id-old-4c216b69.log` SHA256
`6e498fdef94edb3f099e606a7df62060979b9432851a069dbff3b74a74cedf1a`;
native `.cache/headcount-id-native-4c216b69.log` SHA256
`d90f79ac6a6087232ea4852451544341bcb33884003b5dc152e1bac1fa7de6d9`.
Diff checked. Review by: Codex (implementer first filter). Recorded by: Codex. Whole source/test/
docs/records range after `4c216b69` through this containing commit needs designated Claude review,
consolidatable with existing Headcount construction. RP-571 is locally corrected, not approved.
No live player/SQL/Linux/Firefox/hosted/AC3–AC5 or1.0 milestone claim, archival, push or deployment.
Next: actual hiring/output ruling and author reconciliation before the complete integration
chain; RP-413, version ownership, data/pacing/copy and independent review remain separate gates.
