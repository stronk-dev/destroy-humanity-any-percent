# RFC-0001 Job Log

## 2026-07-27 (codex, session 1)

- RFC-0001 assigned by Marco and moved from `accepted` to `implementing` in the RFC and
  index.
- Read RFC-0000, RFC-0001, `design/00-vision.md`, `design/06-tech.md`, and the referenced
  generator-cost formula in `design/02-economy-balancing.md` section 2.1.
- Confirmed scope is rendering-independent. Browser-engine research and the eventual
  client-engine RFC do not block this implementation.
- Created the durable plan and acceptance gates. No scope changes or design deviations.
- Next: create minimal workspaces, pin dependencies, and interrogate the JavaScript
  reference behavior before porting it.

## 2026-07-27 (codex, session 1 — parity investigation)

- Added the minimal Go/TypeScript workspaces, pinned `break_eternity.js` 2.1.3 and
  Vitest, and generated a deterministic 6,055-case shared fixture. The JavaScript
  suite passes all vectors.
- Implemented an initial layer-0 Go port and used the shared fixture to compare actual
  wire strings. This exposed a blocking specification contradiction rather than a
  missing arithmetic branch.
- `DESIGN-GAP:` a normalized `{mantissa float64, exponent int64}` cannot preserve every
  exact layer-0 IEEE-754 value held internally by `break_eternity.js`. Concrete case:
  JS stores `9814222.82599102333` and renders `9814222.825991023`; its normalized
  mantissa is `9.81422282599102402`. Multiplying that mantissa back by `1e6` does not
  reproduce the stored value. There is no second field available to retain the lost
  layer-0 bits while also satisfying the specified normalized representation.
- `DESIGN-GAP:` exact cross-runtime strings for transcendental operations are not a
  portable contract. V8's `Math.log10`/`Math.pow` and Go's `math` can differ by an ULP;
  other browser engines may differ from V8 as well. The retained goja harness would not
  detect this class because goja implements JavaScript math through Go's math library,
  while the Node-generated vectors do detect it.
- Did not weaken, round, or curate away the failing cases. A spec decision is required:
  either change the Go representation/parity contract or explicitly define a shared
  canonical rounding/serialization layer and tolerances for transcendental results.
- Proposed RFC-0001 amendment: require exact canonical wire round-trips and exact basic
  arithmetic where representable; define a shared significant-digit canonicalizer for
  persisted/wire values; compare `pow`/`log10`/`ln`/`exp` numerically within a stated
  tolerance. Alternative: replace the specified representation and both math paths with
  one independently specified deterministic implementation, which is materially larger
  scope.
- BLOCKED: Marco must choose the parity contract before the numeric implementation can
  be completed without improvising behavior.

## 2026-07-27 (codex, session 1 — numeric research correction)

- Reopened the numeric architecture research after Marco correctly challenged whether the
  original pre-RFC research covered idle-game scale, precision, rounding, purchasing boundaries,
  and persistence deeply enough. It did not: `tech-stack.md §3` selected libraries and proposed
  golden vectors, but did not establish a viable cross-runtime contract.
- Landed the focused evidence report `design/research/numeric-core.md`. It distinguishes the
  enormous available range (`1e(9e15)`) from the approximately 15–17 significant digits of useful
  precision and documents why the former is sufficient for all currently designed magnitudes.
- Confirmed from primary sources that ECMAScript transcendental math is
  implementation-approximated and Go does not guarantee bit-identical `math` results across
  architectures. Confirmed that `break_infinity.js` deliberately rounds/ignores insignificant
  additions and that an established C# port uses relative-tolerance compatibility tests,
  including `1e-13` for `Exp`.
- Recommended contract for approval: keep layer-0 coefficient/exponent values; quantize
  authoritative state to 12 significant decimal digits at transition boundaries; use a
  project-defined canonical finite string; keep counts and other discrete facts as exact
  integers; reserve tolerances for raw cross-runtime arithmetic tests, never authorization;
  verify and correct every closed-form max-affordable result against the cost inequalities; and
  reject non-finite gameplay state and saves.
- RFC-0001 remains blocked pending approval of that contract. No RFC text or implementation was
  changed on the basis of the recommendation.

## 2026-07-27 (numeric contract adopted)

- RFC-0001 was amended in place while still implementing to adopt the researched
  server-authoritative numeric contract. The parity DESIGN-GAP is resolved and implementation is
  unblocked.
- Clarified the executable edges before resuming: `Quantize(12)` is a named operation with a
  coefficient-scaling, ties-to-even algorithm and edge vectors; cross-runtime integer counts are
  capped at JavaScript's exact `2^53 - 1`; economy helper return types are integer/Decimal rather
  than ambiguous; and tolerant vector comparisons use symmetric relative error for non-zero
  finite results with exact zero and domain classification.
- Updated the living plan and acceptance gates to match the amended RFC. The existing spike and
  vector corpus now require rework; they are evidence, not accepted implementation.
