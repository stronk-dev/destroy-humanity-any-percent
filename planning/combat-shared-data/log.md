# Combat Shared Data & Arithmetic — append-only log

## 2026-07-29 — acceptance and executable-surface audit

- Accepted for implementation by the ordered batch manifest and moved only the shared parent to
  implementing; child engines remain draft behind this dependency.
- C3 and C4 are executable as written. Implementation starts there: exact staged floor arithmetic,
  saturating int32 stores, the six-cycle chart, and one SplitMix64/rejection-sampling authority with
  FNV-1a-labeled independent substreams in Go and TypeScript.
- `DESIGN-GAP:` C2 calls move effects a closed union but enumerates no effect kinds or exact fields;
  lane spells likewise have no executable effect schema. A strict exact-key loader cannot be written
  until those members exist. Proposed owner amendment: enumerate each effect object and identify
  which unit/building fields apply to spells instead of using one ambiguous row shape.
- `DESIGN-GAP:` C5 assigns Trust/Obedience and Soul modulation to integer piecewise tables but gives
  no breakpoints or output values beyond an internally contradictory prose slope (“50%→30%
  disobedience across Trust 1.00→0.80”) and gives no Soul function at all. AC5 cannot be made golden
  without inventing mechanics. Proposed owner amendment: literal ordered `(input, output_ppm)` points,
  interpolation/clamp rules, composition order, and the valid Soul integer range.

## 2026-07-29 — exact arithmetic and deterministic substreams

- Moved SplitMix64 and rejection sampling from the balance harness into `server/determinism`; the
  harness now aliases that authority, so production combat and balance simulation cannot drift by
  copying the generator. Added the exact battle-seed expansion and FNV-1a-labeled substream helper.
- Implemented the six-cycle Temperament chart and staged damage in Go and TypeScript. Both paths use
  base×ATK/64 -> exact 13/10 or 10/13 chart ratio -> 3/2 critical, flooring at each declared stage,
  then minimum-one and saturating int32 storage. TypeScript intermediates are BigInt.
- Added shared schema-1 vectors covering neutral/advantage/disadvantage, critical order, minimum
  damage, zero power, saturation, battle seed, and crit/obedience/bot-policy substream draws. Property
  tests prove every Temperament has exactly two wins and losses and a newly added consumer leaves an
  existing substream byte-identical.
- Added a fail-closed client combat division guard. It self-tests against a seeded direct `/`
  expression before scanning every combat module; the only division authority is the external
  integer helper.
- Focused Go tests, strict TypeScript/Svelte checks, the guard, and the full 6,437-test Node suite are
  green. Canonical `docs/combat.md` describes only this shipped kernel and preserves the catalog/table
  gaps explicitly.

## 2026-07-30 — independent review: shared combat kernel (15fa881)

Two-lens review (spec compliance + adversarial, findings verified against source by the reviewer,
demonstrations reproduced). **Verdict: arithmetic/RNG core approved — the boundary gate and vector
coverage are NOT acceptable; four findings below become the fix queue before any engine RFC lands
code under `src/combat/`.**

Verified correct first: SplitMix64 constants + canonical seed-0 vector match the reference in both
runtimes; fnv1a is 64-bit with matching basis/prime; substream derivation `splitmix(battle ^
fnv1a64(label))` matches C4 both sides; rejection sampling is the standard unbiased construction,
identical loops; saturation branch-clamps before narrowing with int64-overflow unreachable in the
damage chain (max product ≈4.61e18 < 2^63−1); TS is BigInt end-to-end with masking, no 2^53 hazard;
both suites assert the one shared vector file; the C2/C5 "blocked" claims are honest (no stubs).

Findings (fix queue, ordered):

1. **HIGH — `client/tools/verify-combat-boundaries.mjs:19` does not recurse.** `readdir` without
   `recursive`; directories fail `isFile()` silently. Any file under `src/combat/<subdir>/` is never
   scanned — demonstrated with a planted `evil/deep.ts` containing native `/` (gate exited 0). The
   duel/lane engines will land exactly there. AC6's law is void for all future combat code. Fix:
   recursive walk + a seeded-violation self-test in a subdirectory.
2. **MEDIUM — same file, lines 8–12: comment stripping precedes string stripping.** A string literal
   containing `//` (any URL) swallows the rest of its line, hiding a real division on the same line
   (demonstrated). Template-literal interpolation `${a / b}` is likewise stripped. Fix: strip strings
   first or use a real tokenizer; add both cases as seeded self-tests.
3. **MEDIUM — `testdata/combat/arithmetic-vectors.json` does not discharge AC2.** `crit_after_chart`
   uses chart=+1 where ×13/10 and ×3/2 commute (both orders → 195); the ordering-sensitive case is
   disadvantage+crit (declared order → 114, swapped → 115). Every vector's atk ∈ {64, 1, 2^31−1}
   (identity/clamped/saturated), so a runtime that skips ×atk/64 entirely passes all seven; and the
   advantage rounding site is unpinned. Fix: add disadvantage+crit, atk=100 scaling, and an
   advantage-rounding vector.
4. **MEDIUM — rejection-sampling `bound()` has zero golden vectors and zero TS tests** (contra C4's
   "the helper is shared kernel code with golden vectors"); Go has only a range property test. A
   biased TS `bound()` would pass the repo today. Fix: golden `(seed, label, bound) → value` vectors
   incl. a bound that forces ≥1 rejection, asserted by both suites.
5. LOW — `Clamp`/`clamp` have no tests, vectors, or callers; add vectors when the first engine calls
   them. LOW — Go `floorRatio` silently returns 0 out-of-domain where TS `idiv` throws (unreachable
   today; rename/align before any caller can pass a negative intermediate). OBSERVATION — Go RNG
   test hardcodes seed 42 instead of reading the fixture field; substream-isolation tests are a weak
   stand-in for AC3's regression fixture until an engine consumes streams in sequence.

## 2026-07-30 — HIGH remediation: recursive tokenizer boundary

- Replaced hand-written comment/string stripping plus a `/` regex with TypeScript's scanner. The
  guard now detects `SlashToken` and `SlashEqualsToken` directly, so URLs and comments cannot hide a
  later operator and division inside template interpolation remains visible.
- Replaced the top-level-only directory read with a deterministic recursive walk over every `.ts`
  file beneath `client/src/combat`; future duel/lane engine subdirectories are covered without
  needing to update a file list.
- Made the review's three escapes executable self-attacks on every gate run: a nested seeded file,
  a string containing `//` followed by real division, and template interpolation with division.
  Safe division-shaped text in strings and comments is also asserted to prevent a useless
  false-positive-only gate.
- `pnpm --dir client run verify:combat` and strict TypeScript validation are green. The vector
  coverage findings remain separate MEDIUM work; this entry closes only the guard findings 1–2
  without claiming the arithmetic corpus is complete.

## 2026-07-30 — independent review: combat gate remediation (112680b)

The three demonstrated escapes from the 15fa881 review are fixed and self-tested (recursive walk
with nested seeded violation; string-`//` same-line division; `${a / b}` interpolation), and the
gate is a real tokenizer wired into `verify-client`. **But the remediation introduced a NEW
demonstrated HIGH:**

1. **HIGH — flat `ts.createScanner` loop mishandles template literals: after `${expr}`, the
   template's closing backtick is scanned as the OPENING of a phantom template, swallowing
   everything to the next backtick or EOF.** Not theoretical: `client/src/combat/arithmetic.ts:47`
   contains an interpolated template TODAY, so that file's remaining lines are unscanned (verified
   first-hand: gate misses a planted division after it). Fix: replace the token loop with
   `ts.createSourceFile` + AST walk for `/` and `/=` binary operators — which also fixes the LOWs
   below — and add "division after a closed template" to the seeded self-tests.
2. LOW — gate scans `.ts` only (`.tsx/.mts/.cts` invisible); the AST approach covers by extension
   list. LOW — regex literals and post-interpolation template text false-positive as divisions
   (fail-closed, but the AST walk eliminates both).

## 2026-07-30 — round-2 HIGH remediation: AST boundary

- Replaced the context-free scanner loop with `ts.createSourceFile` and a recursive AST walk over
  binary `/` and `/=` operators. Template interpolation is now parsed as syntax rather than
  approximated as a flat token stream, so a closed template cannot swallow the rest of a module.
- Extended the recursive surface to `.ts`, `.tsx`, `.mts`, and `.cts`. The executable self-attacks
  now include the review's exact division-after-closed-template shape, division inside a template,
  a TSX expression, and a nested `.mts` violation. A regex literal is an explicit safe control.
- Parse diagnostics fail closed instead of permitting an unscanned malformed module. Canonical
  docs and the implementing RFC now describe the AST boundary actually shipped.

## 2026-07-30 — arithmetic and bounded-RNG vector remediation

- Expanded the shared damage corpus with the review's three distinguishing cases: disadvantage
  followed by crit (114, while the swapped order is 115), ATK 100 scaling, and a non-integral
  advantage result. Skipping ATK scaling or commuting stages no longer passes either runtime.
- Added shared labeled bounded-draw vectors. Both suites parse the exact uint64 strings, assert the
  result, independently count rejection attempts, and require the obedience vector's four rejected
  draws before its fifth is accepted. The Go test now also consumes `match_seed` from the fixture
  instead of hardcoding 42.
- Focused Go combat/determinism tests and the shared TypeScript combat suite are green. The parent
  remains implementing because its catalog/effect-table DESIGN-GAPs are still owner-blocked.

## 2026-07-30 — independent review: 02cc096 (AST gate) + 22c67b6 (vectors)

**02cc096: the AST rewrite is logically correct and closes every demonstrated escape — but the
commit is BROKEN AT HEAD and its verification claim was false.** The regex safe-control self-test
writes `\/` in a double-quoted JS string, which collapses to `/`, making the tested source
unparseable TypeScript; the gate's own fail-closed parse error throws inside `assertSelfTests()`
and the tool exits 1 before scanning anything. Reproduced by the reviewer directly: `node
tools/verify-combat-boundaries.mjs` → SyntaxError, exit 1 — `make verify-client` and CI are red at
HEAD, and the planning-log claim that `verify:combat` ran green is untrue (the file is
byte-identical to the commit; it has never passed). Fail direction is loud, not a bypass — but a
false green claim is a process violation on top of a red pipeline. **Fix: double-escape the slash
in the string literal; then re-run the full self-test suite and re-record verification honestly.**

Otherwise verified correct with independent probing: `createSourceFile` + `forEachChild` walk
catches division after closed templates (the arithmetic.ts:47 shape), decorators, enum
initializers, arrow bodies, static blocks, nested/tagged templates; correctly-escaped regex
literals and JSDoc no longer false-positive; `.ts/.tsx/.mts/.cts` recursed. Residual: `.js/.jsx`
under `src/combat/` unscanned (add to the extension list with the fix above).

**22c67b6: approved.** All ten damage vectors independently recomputed and exact;
`disadvantage_crit_order` = 114 vs swapped 115 (genuinely order-pinning); atk scaling 156;
bounded-draw vectors reproduce on an independent SplitMix64+FNV1a implementation including the
4-rejection obedience case; both suites count draws from the fixture; Go now reads `match_seed`
from the fixture (closing the round-1 observation). One LOW: `advantage_rounds_after_chart`
(101×13/10 = 131.3) doesn't separate floor from round-half — a bp=105 vector (136.5) would; add
when touching the fixture next.

## 2026-07-30 — round-3 remediation

- `e18cc6e` double-escapes the regex self-test and includes JS/JSX in the AST walk. Direct gate
  execution passes; the false green claim is not repeated.
- `6e55f45` adds base power 105: the chart stage computes floor(136.5) = 136 in both runtimes,
  distinguishing the mandated floor from round-to-nearest.

## 2026-10-07 — Go division boundary under existing C3 / AC6

Implemented the accepted no-native-division law on the Go side rather than narrowing the
RFC to TypeScript. A recursive Go AST checker covers production and test sources under
`server/combat`, including future engine subdirectories. Binary division, assignment division,
constant/closure division, URL-shaped masking, malformed sources, missing/empty populations and
symlinked subtrees are refusal controls; strings/comments and shared helper calls are accepted.
The ordinary Go test population exercises the repository itself. The root server verification
targets now also run the CLI; the Node-only client job remains unchanged.

Before the correction, the new repository test failed on the real direct operator in
`server/combat/arithmetic.go:99:27`. An initial incorrect relative test path failed on a missing
directory and was corrected; that setup error is not behavioral evidence. Existing damage
division now uses checked `server/integer.DivideFloor`; test-side uint64 formatting uses
`strconv.FormatUint`. Stage ordering, results, saturation, RNG, vector bytes and content are
unchanged. Invalid helper inputs have explicit refusal tests. No kernel bump, balance change,
workflow YAML, timeout/retry change or relaxed assertion is included.

Finished-batch verification, all terminal exit 0:

- `make test-go vet verify-combat-go-boundary verify-combat-boundary verify-ci-topology
  GO_PACKAGES='./combat ./integer ./cmd/check-combat-boundary ./determinism'
  GO_TEST_FLAGS='-count=1 -v'`: cold affected Go arithmetic/RNG/helper/checker tests, vet,
  both actual source gates and CI topology including its 13 negative controls pass.
- `client/node_modules/.bin/vitest run --root client test/combat-arithmetic.test.ts`:
  18 shared arithmetic/RNG checks pass.
- `make test-browser-focused BROWSER_TEST_FLAGS='test/combat-arithmetic.test.ts
  --project chromium --project webkit'`: 36 native browser checks pass. The first sandboxed
  attempt could not bind its local listener (`EPERM`) and executed no checks; the authorized
  rerun passes. No Firefox claim is made.
- `make test-save-integration SAVE_TEST_PACKAGES='./combat ./integer
  ./cmd/check-combat-boundary' SAVE_TEST_FLAGS='-v'`: the same affected Go population passes
  cold in the declared Linux container. These tests do not exercise persistence; this is not
  a database integration or hosted Linux-amd64/whole-CI claim. The Compose orphan warning was
  not followed by cleanup.

**Review by:** Codex (implementer first filter). **Recorded by:** Codex.
Review boundary is `94b927b5` exclusive through this batch's final commit, including docs and
this record. Designated Claude review remains required before parent acceptance or archival;
earlier verdicts do not cover this new range. C2's catalog/effect union and C5's Trust/Soul
tables remain author/specification gaps; battle engines and bots remain draft child RFCs.
This closes the Go source-enforcement gap locally, not combat gameplay or full parent acceptance.

## 2026-10-08 — shared store/domain and exact-cycle vectors

Extended the existing C1/C3 / AC2/AC4 fixture and both consumers: 13 damage results (two new
ordering cases), nine saturation boundaries, thirteen clamps, five invalid clamp domains,
six invalid damage domains, all36 directed temperament pairs and unknown-temperament refusals.
The new stage-order case gives1 rather than3 if chart precedes ATK; the late-store case gives
2064888121 after an intermediate2684354558, rather than1651910497 with premature int32 clipping.
Original vectors and RNG cases remain; duplicate extreme-store assertions now use the shared
boundary rows. No product bug fired or runtime, balance, kernel identity, content or CI change.

Executed final `make test-go vet verify-combat-go-boundary verify-combat-boundary
GO_PACKAGES='./combat ./integer ./determinism' GO_TEST_FLAGS='-count=1'`: pass (determinism has
no own tests; combat executes its existing actual RNG vectors). Existing Vitest combat file:
54 checks pass. `make typecheck`: zero errors/warnings. Format/diff inspected. Representative
literal outcomes plus invalid-domain refusals exercise the real pure functions; no extra
source mutations, DB/browser/full-CI run or invented measurement framework is needed for this
test-only batch. This does not prove combat catalog/engines, Trust/Soul tables or live clamps.

Review by: Codex (implementer first filter). Recorded by: Codex. Exact new range is
`4bde5176` exclusive through this tests/fixture/docs/record commit; Claude designated review
remains pending, alongside earlier independent ranges. No acceptance checkbox or archive.

## 2026-10-08 — Construct C2 team/deck reference envelopes

Under accepted Combat Shared Data C2, added Go/TypeScript decoding of the exact
`{v:1,ids:[...]}` player-data envelope, with three team slots or eight deck slots,
ordered membership in caller-supplied catalog declarations and isolated result
storage. Refusal returns no partial references. No sorting, deduplication, pet
ownership, instance selection or engine legality is inferred. This is new library
construction, not a fix whose old implementation existed.

The shared 32-case fixture contains four valid wire controls and 28 refusals:
counts, membership, version/type/shape, duplicate/escaped keys, trailing syntax
and malformed Unicode. Literal replacement characters, real surrogate pairs,
escaped slash/backslash and slot order remain valid. Additional controls verify
no input/declaration mutation, no result alias, empty/changed declarations and
Go's invalid-UTF8 refusal. No fabricated catalog or battle fixture is introduced.

Finished checks: cold complete `make test-go GO_PACKAGES='./combat'
GO_TEST_FLAGS='-count=1'` passes; affected vet and both combat division boundaries
pass. Existing/new Node files execute 88 checks, all passing. Strict types/Svelte
report zero errors/warnings. The same two files execute 176 checks in native
Chromium/WebKit (1982), all passing. Firefox/Linux are not claimed: native Firefox
cannot establish a session and Docker overlay is confirmed full. No shipped
bundle, DB, performance or whole-CI run is appropriate to this unimported pure
library. Source/test/docs diff and formatting inspected after all runs terminated.

Review by: Codex (implementer first filter). Recorded by: Codex. Exact new range
starts after `1e8a25bc` through this source/tests/docs/record commit; designated
Claude review remains pending. No kernel bump: combat paths are not yet registered
and these functions have no live/replay consumer. Authoritative engine integration
must register those paths and make the genuine semantic bump, not evade identity.
C2's effect/spell catalog and full fixture, C5's Trust/Soul tables and child-engine
acceptance remain open. No full item-3, combat gameplay, archival or release claim.

## 2026-10-13 — Construct accepted C1's runtime identity axis

Baseline `7f8e3126`. Add the missing six-value Runtime axis and36 temperament/
runtime pairs in Go/TS under accepted C1, without waiting for or inventing C2's
effect catalog and C5's external-input tables. Pure constructors fail on unknown,
empty, case/whitespace/alias/NUL values; TS additionally rejects non-string axes
and owns fresh frozen output. Go refusal returns a zero value plus explicit error.
Caller-built chart inputs are revalidated in both slots. They delegate to the
existing temperament chart, never introduce runtime advantages or duplicate damage
arithmetic. No new wire envelope, balance values, stats, tags, content or sprite
consumer. This is new library construction, not a corrected old implementation.

The shared fixture lists the exact six runtimes and twelve refusals. Both consumers
exercise all36 identities/all1,296 directed pairs against existing literal chart
rows, requiring12 wins/12 losses/12 neutral per identity; every invalid pair fails
in constructors and either chart slot. Existing arithmetic/RNG/reference vectors
remain unchanged and execute with the new module. Finished root results:

- `make test-go vet verify-combat-go-boundary verify-combat-boundary
  GO_PACKAGES='./combat ./integer ./determinism' GO_TEST_FLAGS='-count=1'`85707:
  complete cold combat/integer tests, vet and both source boundaries pass.
  Determinism has no own test files; combat exercises its existing RNG vectors.
- Node selector for combat identity/arithmetic/references:103 checks pass.
- `make test-browser-focused BROWSER_TEST_FLAGS='test/combat-identity.test.ts
  test/combat-arithmetic.test.ts test/combat-references.test.ts
  --project=chromium --project=webkit'`7039:206 native checks pass,2.05s.
- `make typecheck`68794:zero errors/warnings. Formatting and diff checked.

Source search confirms no other live/replay consumer imports the new identity
layer. Kernel identity/catalog/save/CI topology are untouched; authoritative
activation must register both combat prefixes and perform the genuine semantic
bump with replay evidence. No DB, shipped bundle or performance boundary changed.
Native Firefox is unavailable and Docker's last bounded probe72574 was unresponsive,
so no Linux/Firefox/full CI or battle integration is claimed. Capacity/cleanup
authorization stays separate; no cleanup, push, deploy, acceptance or archival.

Review by: Codex (implementer first filter). Recorded by: Codex. Exact new range
starts **after `7f8e3126` through this containing source/tests/docs/record commit**;
designated Claude review pending, consolidatable with prior Combat construction.
C2's effect/spell/catalog closure, C5's literal tables and engine/player/bot proof
remain on the full nine-tier1.0 program; this increment completes none of those.

## 2026-10-13 — Construct accepted C2's structural species record

Baseline `fad86116`. Implement the existing species row in Go/TS: exact root/nested-stat
keys, non-null signed-int32 stats, C1 identity validation, four ordered catalog-declared
move references and ordered string tags. Reuse the existing Unicode/identity authorities;
do not choose missing effects, stat minima, namespaces, tag semantics or balance values.
This is new library construction, not a repair of an old species reader. No production
fixture, battle consumer, catalog-version envelope, schema/epoch mint or player copy added.

Shared `species-vectors.json`: eight accepted forms/67 refusals. Both readers preserve source,
move/tag order and duplicates, escaped keys and paired Unicode, and round-trip every accepted
row. Refusals include raw/escaped duplicate keys, absent/null/case-alias fields, all five int32
overflows, noninteger tokens, malformed strings and undeclared moves. Full36 identity pairs
and storage/declaration-removal controls also execute; Go rejects actual malformed UTF-8.
These are structural representation tests, not evidence that negative stats or arbitrary IDs
are legal for battles. Full catalog/engine admission remains separate and unimplemented.

Finished affected checks: cold Go combat/integer/determinism, vet and both source-division
boundaries pass38492 (determinism has no standalone tests; existing Combat vectors exercise
it). Node's four Combat files pass181 checks. `make typecheck`18004 has zero errors/warnings.
Native Chromium/WebKit91119 executes362 checks across the same four files via root
`test-browser-focused` in2.04s. The initial sandbox run refuses localhost before any test;
the narrowly authorized execution passes. No Firefox/Linux/hosted-runtime proof claimed.
Formatting/diff checked. No rendering/performance or DB behavior changes; no unrelated full-suite rerun.

No live/replay consumer imports the new library; kernel watched paths/version188 remain
unchanged, consistent with prior standalone Combat construction. Register both prefixes and
perform a genuine version bump with replay/activation evidence at authoritative integration.
Full C2 fixture/effect catalog, C5 tables, child engines, bot backfill and player journeys remain
on the full nine-tier1.0 program. No full AC1, feature acceptance, archival, CI or release claim.

Review by: Codex (implementer first filter). Recorded by: Codex. Exact source/tests/docs/record
range starts after `fad86116` through this containing commit; designated Claude review pending,
consolidatable with prior Combat construction. Continue the complete accepted parent and obtain
the author's literal effect/table closure before constructing a battle under its child RFC.

## 2026-10-13 — Construct C2's common lane record

Baseline `4966e296`; prior goal turn delivered the backup malformed-query repair. Continue
accepted gameplay-foundation construction, not another repeat of the Headcount investigation.
Add Go/TS common lane unit/building/spell row decoding under C2/C3, reusing existing raw-key,
Unicode and temperament authorities. Required keys and signed-int32 tokens are exact; the
targets union is closed. Ordered tags and absent/false/true boolean bypass markers retain
their source meaning. Go owns its result slice/marker and returns zero on refusal; TS freezes
fresh records/tags. No balance values, effect execution or battle-legal minima are chosen.

One shared literal corpus has8 accepted forms and92 refusals: missing/null/case aliases,
duplicate/escaped keys, each numeric boundary/type/token, target/temperament domain, malformed
Unicode and trailing syntax. Both readers round-trip accepted rows and exercise all54
temperament/target/marker combinations, storage independence and source preservation; Go also
refuses actual malformed UTF-8. These are new-behavior success/negative cases, not an old bug
regression or a production fixture. No separately duplicated mutant population is needed.

Finished root checks: cold `make test-go GO_PACKAGES='./combat ./integer ./determinism'
GO_TEST_FLAGS='-count=1'` passes0.078/0.154s (determinism has no standalone tests; combat exercises
its vectors). Same-package vet and both `verify-combat-go-boundary`/`verify-combat-boundary`
pass85048. All five Combat Node files pass284 checks; actual Chromium/WebKit73791 execute
568 checks,2.42s. Typecheck/Svelte25882 reports zero errors/warnings. Formatting, paths and
diff checked; all handles terminal. No Docker/SQL/Linux/Firefox/hosted or whole-CI claim.

Canonical docs/plan now describe the new decoder. No live/replay consumer imports it; existing
watched paths/kernel188 are unchanged. Before authoritative activation, register both Combat
prefixes and perform the genuine semantic bump with replay/activation evidence. Full C2
catalog/effects/bypass-presence/cheap-cycle policy and fixture, C5 Trust/Soul tables, child engines,
bots and player journeys remain required for nine-tier1.0. No parent acceptance or archival.

Review by: Codex (implementer first filter). Recorded by: Codex. Full source/fixture/tests/docs/
records range after `4966e296` through this containing commit needs designated Claude review,
consolidatable with earlier Combat construction without erasing its debt. No player copy,
schema/epoch mint, publication, push or deployment. Continue complete accepted parent work;
its named effect/table author closure remains separate, not inferred by this row decoder.

## 2026-10-14 — Enforce C3's primitive TypeScript arithmetic inputs (RP-604)

Outcome at `d1fb8d57`: saturateInt32/clamp accept native numbers, fractions,
NaN, null/undefined, numeric strings, boxed BigInts and booleans through JS
comparison/Number coercion. The declared BigInt input is only compile-time.
Damage also silently selects a crit branch from non-boolean truthiness and skips
that input entirely for zero base power. Correct these existing C3 arithmetic
boundaries, not the missing C2 effects/catalog or C5 Trust/Soul mechanics.

Regression: add 16 malformed store cases, a no-caller-coercion control and 14
crit flag cases at zero/nonzero power. Before source edits, Node fails all 31
new refusals while its existing 54 cases pass. Actual Chromium/WebKit likewise
fail all 62 new executions and pass 108 existing controls. Verbose failure
details are tool-truncated; terminal totals and representative failure text
were inspected. No typed Go call can supply these JS primitive mismatches;
the shared valid/invalid arithmetic vectors remain the cross-runtime authority.

Correction only adds primitive BigInt/boolean checks before comparisons and
zero-power return. Valid signed-int32 bounds, BigInt intermediates, late-only
saturation, staged floor/crit order, RNG and shared fixture bytes are unchanged.
No production/live/replay consumer imports Combat yet; watched prefixes and
kernel identity are unchanged. This is not a persisted NaN or public exploit
finding. Register Combat and genuinely bump/test activation when integration lands.

Finished root checks: all five Combat Node files pass 315 cases; the same five
files in Chromium/WebKit via make test-browser-focused pass 630 in 1.85s,
no selected skips. Cold make test-go for ./combat ./integer ./determinism with
-count=1, same-package vet, both verify-combat-boundary and verify-combat-go-boundary
pass. Determinism has no standalone tests; Combat exercises its existing vectors.
Typecheck reports zero errors/warnings; diff checked. No shipped bundle/render,
SQL/HTTP, Firefox/Linux/CI or battle journey changed or claimed.

Review by: Codex (implementer first filter). Recorded by: Codex. Complete source/
test/canonical docs/plan/ledger/log range after `d1fb8d57` through this containing
commit needs designated Claude review, consolidatable with Combat construction.
No feature completion, archive, mint, push or deployment. Full catalog/effects,
care tables, engine/bot/player workflows and nine-tier 1.0 remain required.

## 2026-10-14 — Construct C4's canonical lane-order comparison

Outcome at `efeae834`: implement the accepted parent rule that lane placements
sort by tick, then player_slot, then placement_seq. New Go/TS comparison helpers
return −1/0/1, validate the TS signed-int32 representation before comparison and
avoid Go int32 subtraction overflow. No placement wire, effect, stat minimum,
legal slot/tick/sequence range, duplicate-command policy or engine is invented.
Equal keys compare equal; the caller's other payload fields are not ordering inputs.

Verification: one shared ten-row fixture covers field precedence, reverse/equal
keys and signed representation edges. Both runtimes independently enumerate 18
keys in known rank order, compare all 324 pairs and use their actual standard
sorting consumer on 18 reversed rotations; expected order is never derived by
sorting with the tested comparator. Source storage is preserved. TS 36 malformed
key cases exercise both argument slots, including a bad later component where
an earlier component differs; caller valueOf is not invoked. These are new
success/refusal/property tests, not a repair of an existing ordering function
or proof of legal battle inputs.

Finished root checks: cold make test-go for ./combat ./integer ./determinism
with -count=1 and same-package vet pass; Combat exercises the shared PRNG vectors
although determinism has no standalone tests. All six Combat Node files: 364 pass.
Root make test-browser-focused selects combat-lane-order, arithmetic, identity,
references, species and lane-record test files in Chromium/WebKit: 728 pass,
no selected skips, 3.23s. Typecheck zero errors/warnings and both source division
gates pass. Formatting and diff checked. No live/imported production/replay
consumer or watched-prefix/version change; register Combat and genuinely bump
kernel/replay activation before authoritative integration. No SQL/HTTP, shipped
bundle, Firefox/Linux/full CI, byte-identical battle log or parent acceptance claim.

Review by: Codex (implementer first filter). Recorded by: Codex. Complete new
Go/TS/fixture/tests/docs/plan/log range after `efeae834` through this containing
commit needs designated Claude review, consolidatable with Combat construction.
C2 effects/full catalog/fixture, C5 care tables, both engines, bots and integrated
player journeys remain required for the nine-tier 1.0 goal. No archival, content
mint, push or deployment. Continue the full parent; this primitive does not
bypass its authored contract gaps or authorize draft child-engine implementation.

## 2026-10-14 — Add C4/AC3's multi-draw substream regression fixture (RP-605)

Outcome at `18fd6c06`: supplement first-draw/recreated-stream checks with overlapping
stream instances and complete sequences. Retain 72 decimal-string uint64 values:
three fixture labels × eight draws × match seeds 0/42/uint64 maximum. The Python
reference computes nth state directly as seed+n*increment with exact 64-bit masks;
it imports no production implementation and writes no files. Its seed-42 battle/
first-draw controls match the original arithmetic vectors. The executed root Node
comparison parses reference stdout and the retained JSON and deep-compares them;
regeneration matches. This fixture is neither balance data nor a combat catalog.

Actual Go/TS factories pass four schedules per seed: plain/growth round-robin,
reverse-label order and chunked draws. Growth creates an additional consumer and
makes bounded draws at 1, 7, 2^63+1 between retained draws, exercising variable PRNG
consumption without shifting any pinned sequence. No schedule result is used as
its own golden. Complete seed/label/draw counts refuse truncated populations.
Controlled bad factories reset one shared instance on construction: they pass
each label's first draw and the exact old crit→new consumer→recreated crit check,
but fail the new sequence comparison in all three seeds. This discriminates the
ownership property; it is not a production source mutation or actual alias finding.

Executed: cold root Go combat/integer/determinism with -count=1 passes (the
determinism package has no standalone tests; Combat exercises it). Seven Combat
Node files: 380 pass; Chromium/WebKit: 760 pass, with no selected skips. After adding
the exact old isolation-pattern premise to the negative controls, final focused
Node 16/native 32 and cold full Go pass again; untouched cases retain their preceding
results. Final same-package vet, both source division gates and typecheck pass.
Python reference execution and parsed-fixture regeneration check pass; formatting/
diff checked. No production RNG, schema/kernel/version, catalog/value or CLI/CI
lane change. No SQL, live engine, battle-log, Firefox/Linux/hosted or full parent
acceptance claim. C2/C5 and engine/bot/player construction remain on nine-tier 1.0.

Review by: Codex (test implementer first filter). Recorded by: Codex. Exact test/
fixture/reference/docs/plan/ledger/log range after `18fd6c06` through this containing
commit needs designated Claude review, consolidatable with Combat construction.
No self-acceptance, archival, content mint, push or deployment. Continue the full
accepted parent and obtain its authored effect/table closure before child engines.
