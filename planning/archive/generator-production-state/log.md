# Generator Production State — Running Log

## 2026-07-28 — Start

- Archived the independently implemented constant-rate accrual primitive.
- Re-reviewed the parent production draft against the implemented catalog and save code.
- Confirmed the next dependency is structural: catalog v1 prices generators but cannot state output,
  while save v1 cannot represent owned counts or an authoritative evaluation cursor.
- Split those settled contracts from undefined intent, multiplier, event, and offline-credit
  mechanics. No gameplay values or flavor identifiers are introduced.
- RFC accepted and implementation started.
- No push performed.

## 2026-07-28 (claude — spec-compliance review of the archived RFC set, findings routed here)

A full spec-compliance review ran against all seven archived RFCs + the CI baseline
(all suites executed). **The archived implementation is unusually faithful** — quantize/grammar/
non-finite/postcondition contracts verified clean in both runtimes, ledger K1–K5 clean, save
D1–D8 clean against real Postgres, CI jobs match spec. Findings, ranked; F1 lands in THIS job:

- **F1 — DEFECT (high), `make verify` is RED at HEAD.** `d794f57` moved the Go loader +
  shared fixture + JSON schema to catalog v2 (`production` block) but
  `client/src/economy-kernel.ts:11` still pins `CATALOG_SCHEMA_VERSION = 1` →
  `client/test/economy-kernel.test.ts` fails at module load (verified). Violates archived
  RFC-0002 AC1/AC6 and CI-RFC AC1. Also: three validators now accept three different version
  sets (Go: {1,2}; TS: {1}; JSON Schema: {2}). **Process note, not just a bug: the server half
  of an RFC was committed with the cross-runtime suite red. If CI had been enforced this
  commit would have been blocked — do not land half of a cross-runtime contract, and align
  the three validators' accepted-version sets deliberately when the TS half lands.**
- **F2 — GAP:** no checked-in save-migration corpus (save-RFC D4/AC3 demand one "from the
  first save version", growing, CI-enforced). v1 saves exist only as inline literals in
  `state_test.go`. Worth landing alongside this job's v2 envelope — the corpus's whole value
  starts at the first real migration, which is this one.
- **F3 — GAP:** RFC-0001 §7's invariant *reporting* (fallback-exceeded + residual clamp)
  is normative text with no code anywhere (`decimal/economy.go:70-73`,
  `economy/curves.go:82-87` fall back silently). Latent until purchase commands — i.e. the
  production-engine RFC — so it belongs on that RFC's list, not as an emergency.
- **F4 — DEVIATION + SPEC BUG:** accrual tie-break — TS uses `localeCompare`
  (`client/src/production.ts:26`, ICU-sensitive) vs Go byte-compare; should be code-point
  comparison. And the spec's canonical-string tie-breaker cannot totally order two state
  values sharing one 12-digit string + non-stable `sort.Slice` → permutation invariance not
  fully guaranteed at full precision. Fix the TS comparator; amend the accrual RFC's
  deviations section for the ordering caveat.
- **F5 — GAP:** `make fuzz` and vectors-regeneration-diff are in no CI job and not in
  `make verify` — "remain mandatory" currently means "remembered manually." Cheap to add as
  a nightly job.
- **F6 — minor:** save AC6 (cross-scope isolation) untested; D3 field-path logging never
  asserted; bare `.toThrow()` in the TS negative tests will pass vacuously for the four new
  production invalid cases — match on the error message; Archive/CreateStream conflict
  granularity; no dedicated near-cancellation vector block.

An adversarial correctness review (parser-grammar desyncs, float edges, store concurrency)
is still in flight; its findings will be appended when verified.

## 2026-07-28 (claude — adversarial correctness review, all findings demonstrated with reproducers)

Second review pass: attack what spec and tests both missed. Everything below was DEMONSTRATED
with a runnable reproducer (I independently re-reproduced A4 before filing). Ranked:

- **A1 — HIGH, `Store.Archive` CAS races `Write`** (`server/save/store.go:191-200`). Archive's
  single `UPDATE … WHERE (SELECT max(revision))=$2` blocks on the writer's row lock, then under
  READ COMMITTED the subquery is NOT re-evaluated against the new data (EvalPlanQual) — the
  stale CAS succeeds and archives a stream whose head it never saw. Demonstrated against real
  Postgres. Write-vs-Write is safe; the integration test only covers that case. **Fix: give
  Archive the same transaction shape as Write** (`SELECT … FOR UPDATE` → check → `UPDATE`).
- **A2/A3 — HIGH, server/client desync at the exponent edges.** 49,877-case differ found
  **502 legal-state input pairs where Go returns NaN and the client returns a valid state
  value**: `Div` via `reciprocal()` overflows transiently (`1e8999999999999999 ÷ 2.5e…` → JS
  `4e-1`, Go NaN; 480 cases, `decimal.go:282`) and `Mul`'s underflow pre-check fires before
  the mantissa carry (22 cases, `decimal.go:266`). Both fixes are local: compute
  `New(m1/m2, e1-e2)` directly; allow one exponent of slack and let `Normalize` decide.
- **A4 — MEDIUM, everything is free at ratio ∈ (1, 1+5e-15)** (`decimal/economy.go:26`,
  mirrored in `client/src/economy.ts:43`): `One.Sub(ratio)` collapses to Zero → cost 0 →
  afford(cash=1, base=100) = 9007199254740991, err=nil. **Re-reproduced independently.**
  Unreachable via canonical catalogs (ratio−1 ≥ 1e-11) but it's an exported API violating its
  own documented contract. Fix: branch to the ratio==1 arithmetic when `One.Sub(ratio).Eq(Zero)`.
- **A5 — MEDIUM, `Ledger.Apply` acceptance is entry-ORDER-dependent for the same multiset**
  (`ledger.go:128-131`): `[+B,+B,−B,−B]` at the domain edge rejects while `[+B,−B,+B,−B]`
  accepts — same net delta (zero). Either document per-prefix validation as normative or
  aggregate per-resource before range-checking.
- **A6 — MEDIUM, the vector generator never tests the top half of the exponent domain**
  (`tools/gen-vectors.mjs:98-107`): `randomExponent` clamps the boundaryExponents list to
  4e15 — dead code for every op except log10; max |exponent| in the shipped corpus is exactly
  4e15. **A2/A3 live precisely in the unexercised region.** Also: the generator's
  quantize/canonical/sumGeometric are verbatim copies of the client code — for the TS side
  those vectors are a self-confirming oracle (only verifiedAffordable is independent). Fix:
  un-clamp the boundary list; consider an independent-oracle note in the vectors RFC.
- **A7/A8 — LOW, latent:** permissive-parser divergences (`FromString` vs `new Decimal`:
  "0x1p4", "1_000", "1e5x", "INF") — safe today because the wire path uses ParseCanonical
  exclusively; 7,622 class-only differences on invalid results (both sides still reject as
  state). Suspected-only: unstable `sort.Slice` ulp nondeterminism (unreachable with
  canonical rates); `AffordGeometricSeries` can return owned+count > MaxExactInteger
  (caller clamps); `store.go:196` discards RowsAffected() error.

**What HELD under attack (worth recording):** quantize idempotence (2M cases, both runtimes,
0 failures) · canonical wire grammar parity (brute force + crafted edges, 0 diffs — the wire
really is desync-proof) · canonical formatting parity (~41k mutually-valid results, 0
mismatches) · negative zero, subnormals, Sub self-cancellation, half-tie rounding · fast-path
postconditions + termination · Write-vs-Write and prune-vs-read concurrency.

**Combined-review priority for this job:** (1) F1 client catalog v2 (HEAD is red) ·
(2) A1 Archive transaction shape · (3) A2/A3 Div/Mul edge fixes + A6 generator un-clamp in
the same change (the tests that would have caught them) · (4) A4/A5 with doc-or-fix decisions
recorded in the RFC's deviations section.

## 2026-07-28 — Review fixes and save v2 implementation

- Restored TypeScript catalog parity: versions 1 and 2 now match Go, v2 production definitions
  receive the same reference/scope validation, and the shared suite is green again.
- Added save payload v2 with exact generator counts and canonical UTC evaluation cursor.
- Added the first checked-in save-migration corpus; v1 migration is deterministic from the
  revision's server-authored timestamp.
- Updated repository create/write/load to validate and return the complete state.
- Real Postgres integration passed, including v1 migration (with PostgreSQL's microsecond timestamp
  precision) and the forced Archive/Write interleaving from the adversarial review.
- No push performed.
- Final `make verify` passed: 6,347 Node tests and 19,041 browser tests across Chromium, Firefox,
  and WebKit. Canonical economy/save docs updated; RFC and planning archived.
