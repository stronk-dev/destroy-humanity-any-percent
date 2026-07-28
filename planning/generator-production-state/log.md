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
