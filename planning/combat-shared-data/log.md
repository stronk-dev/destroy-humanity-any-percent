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
