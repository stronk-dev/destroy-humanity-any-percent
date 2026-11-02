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

## 2026-10-15 — RP-571 independent corrective review: APPROVED

Review by: fresh independent Codex sub-agent `/root/headcount_ids_corrective_review`.
Recorded by: Codex parent, transcribing its verdict under Marco's corrective-review
delegation. Complete six-path range (base exclusive, tip inclusive)
`4c216b690fc32cb03f784446377c69f069979da7..10ca8fbfac95494aad58562ce8e1908dd2b19bf4`.
Verdict: **APPROVED**, no actionable findings; implementation, tests, canonical docs,
backlog and owning plan/log inspected. At14761616 the five other files and the
reviewed backlog row remain unchanged.

Independent Node56, cold headcount Go (count1,0.204s), matching vet, typecheck
(zero errors/warnings) and range diff checks pass. In-memory transpilation of exact
old/current source reproduces all eight old failed refusal oracles and corrected
passes; caller coercions fall7→0. Six valid literal-name controls and twelve shared
arithmetic vectors pass both sources. Initial probe stops on an erroneous auxiliary
object-order assertion; the successful deep-equality run omits only that invalid
auxiliary assertion, not any refusal oracle. No repository source edits.

The primitive string/grammar check precedes sorting and lookup without changing
allocation math. No live/replay consumer or watched kernel change; no browser,
DB, build or integration run needed/claimed for the unused pure library. Historical
native-browser checks remain author evidence, not independent executions. Prior
construction/research debt, full feature acceptance, hiring/output/seat/version
choices, normative adoption, kernel identity, archival and release remain open.
Reviewer creates no files, commits, services or worktrees; no cleanup/publication.

## 2026-11-02 — Paid-hiring/output model ruled and bodies reconciled

Owner ruling by: Marco. Recorded by: Codex root. His direct selected answer is
“Paid hiring + producing/redirecting roles; reconcile both specifications
(recommended).” The question explicitly retains separate prices/rates, copy and
balance ratification. This resolves both actual model questions, not merely
the hiring verb. No preselection or technical-review verdict is relabelled as
owner authority.

Headcount keeps its canonical pool/produce-convert-pool_feed/rate/state/command/
event/projection/mask contracts. Tier 2's body now consumes them: remove office
seat faucets, second catalog/map/intent/event authorities, per-head multiplier
formulas and fake nonproducing Deck counts. Reconcile B/H/E/P/M, owner rows,
acceptance references and plans/index/ledger/current routing in the same batch.
Paid hiring has real cost/unassigned state; role masking keeps those seats/costs.
Actual Deck stock/rate is distinct from diverted heads. No generator-purchase
counter change, payroll, firing or new gate/offer/sink is inferred.

The translation does not fabricate the remaining complete Tier-2 role/resource/
price/rate/cap/consumer content. HG-10 names the necessary deterministic full
paid-hire/allocation policy contract rather than reusing incompatible ppm fields
or pretending a window-only subset is full P1. The adopted activity windows,
seeds/horizons/envelopes, must-reach, role/relevance and full release floor remain.
Historical engine v5/v19/v9 proposals are NOT assigned: current schema5 owns
axis_stack, Company19 attainment, replay inputs12; RP-413's impossible required
TS overflow outcome also remains explicit technical reconciliation. This ruling
does not silently waive either gate or turn them into new gameplay questions.

No product, data, kernel, migration, schema allocation, authored copy or old
artifact/report change. Root reviews the complete normative diff/remaining
references and checks whitespace; no software-suite pass is claimed for Markdown.
Review by: Codex root, author/first filter. Complete translation afterc6f11121
requires fresh independent technical review before implementation. Nothing pushed.

## 2026-11-02 — Model translation review finding and forward correction

Review by: fresh independent Codex /root/headcount_model_ruling_translation_review.
Recorded by: Codex root. CHANGES REQUIRED on complete eleven-path range
c6f11121..bf7396f92a1a4dc340cc6c5cb67ded86a2148d8c. One P2 finding:
Tier 2 E4 still required registering rejected `headcount.roles[].copy_key`;
canonical S1 defines `headcount_pools` and no role copy field. The reviewer
independently checked all paths, preserved logs/windows and the selected model;
no other blocking finding. Whitespace passed; no software suite for Markdown.

Correct E4 forward: use actual seat/resource/generator cap-reason paths and
explicitly require the real mounting/content presentation binding before
implementation/mint. Do not invent a role copy field, schema or authored text.
Root checks the correction against S1/S10 and E4; complete original-plus-correction
range requires independent re-review. No runtime/data/kernel/copy or gate change.

## 2026-11-02 — Complete model translation independently approved

APPROVED, no remaining finding, complete eleven-path base-exclusive/tip-inclusive
range c6f11121..8a00ca57b1a86b0de76010d7d9ca87bd848bce78.
Review by: fresh independent Codex /root/headcount_model_ruling_translation_review.
Recorded by: Codex root, transcribing the verdict. The reviewer rechecks the full
combined range: E4 uses actual cap-reason paths and requires the real presentation
binding without inventing role fields; eight other paths match the original tip;
all three logs remain append-only. Adopted windows and P2 populations/targets are
unchanged except the intended canonical allocation-command substitution. All paths
are Markdown; runtime/data/kernel/migrations/copy/deployment artifacts unchanged.
Whitespace passes. No software suites run for specification-only edits or reviewer
source edits/services/publication. This closes technical model translation only;
Marco's direct answer supplies adoption. Content/policy, historical identity,
RP-413, full implementation/acceptance, mint and release remain separate gates.
