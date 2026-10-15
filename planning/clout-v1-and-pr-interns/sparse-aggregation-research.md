# RP-664: sparse exact-input aggregation candidate

Status: unapplied arithmetic prototype, **not adoption-ready**. Baseline35011479.
Production still uses the accepted grouped algorithm. This does not repair AC6.

## Representation and proposed arithmetic

Every valid nonzero normalized binary64 mantissa has unbiased binary exponent0..3.
Therefore `q = abs(mantissa) * 2^52` is an exact integer, at most56 bits. The
prototype decodes those bits, rather than rounding a decimal rendering of the
mantissa. An input's exact represented value is `sign * q * 10^e / 2^52`.
This is private arithmetic, not a new save, wire, receipt or origin protocol.

Use radix `B = 10^9`. Euclidean division gives `e = 9k + r`, `0 <= r < 9`.
Split `q * 10^r` into at most three integer radix digits starting at indexk.
Keep positive and negative contributions in separate sparse maps, then normalize
their nonnegative carries in ascending occupied-index order. If carry becomes
zero, jump to the next occupied index; never walk an empty exponent gap.

Compare the normalized maps from their highest digit to determine sign. Subtract
the smaller from the larger exactly. A borrow through an empty interval yields
one run descriptor `{first,last,digit:B-1}`, not one allocation per missing digit.
This preserves low-order terms until cancellation is complete, including across
different exponent groups. Exact zero returns canonical zero.

### Final conversion, not intermediate float addition

Let `M` be the positive sparse magnitude before division by `D=2^52`, and let
`E_M` be its leading decimal exponent. The exact sum's normalized exponentE is
either `E_M-15` or `E_M-16`. Decide using the integer floor of `M/10^E` againstD.
The normalized mantissa's binary exponentk is0..3, determined by thresholds
`D * 2^k`. Neither decision needs a dense exponent-sized integer.

For rounding, extract `J=floor(M/10^(E-1))` and a sticky flag indicating a nonzero
remainder below that decimal cut. Each query intersects at most three radix
chunks: `E` differs from `E_M` by15 or16, so the final cut is at most17 decimal
places below the leading digit. DivideJ by `10 * 2^k`. Compare twice its integer
remainder with that denominator. Exact equality rounds upward when sticky is set;
otherwise round to an even quotient. Because the denominator is even and the
discarded fraction is less than1, no other comparison can change sign.

The resulting integer significand lies in `[2^52,2^53]`; converting it to binary64
and multiplying by `2^(k-52)` is exact. Normalize a possible decimal carry and
apply the existing state-domain check. Existing Quantize12 is unchanged.
No transcendental operation is used to choose the significand.

Proposed domain semantics: invalid inputs, an exact nonzero sum whose normalized
exponent is outside the state domain, or a final rounded representation leaving
that domain returns diagnostic NaN. Recoverable intermediate overflow is allowed
inside the integer calculation. This explicitly includes net-underflow refusal;
it does not change or resolve the separate RP-663 division policy by implication.
These candidate semantics and their compatibility still require normative review.

### Work argument to be reviewed

Forn inputs, initial digit contributions are at most3n. An accumulated digit plus
incoming carry is below `n*B`, so outgoing carry is at mostn. Each occupied key
therefore creates at most `digits_B(n)` extra adjacent carry positions before
either reaching another key or exhausting carry. Sorting costsO(n log n); carry
work and coefficient widths depend on input count, not exponent distance.
Subtraction visits only emitted digits and adds at most one borrow-run per gap.
The two or three rounding queries read at most three chunks each, found by binary
search over those pieces. These are representation-derived invariants, not new
operating budgets. An invariant failure invalidates the experiment; it is not an
unreachable result, silent exclusion or zero payout.

Exact accumulation is monotone in each input. Correctly rounding its normalized
mantissa is monotone within each exponent interval; a carry reaches the next
interval's lower endpoint rather than reversing order. Existing Quantize12 is
likewise monotone. This is the proposed mathematical argument, not independently
reviewed proof that the implementation covers every valid input or temporal
producer regime. H3c's actual fixed-origin caller proof remains separate.

## Executed evidence

Declare before measurement: original RP-664 pair; fourteen empty/signed/full-range/
tie/domain controls;2048 uint64-xorshift seed664 profiles,0..16 terms at exponents
−32..32 with signed15-digit mantissas, adding a cancelling first-term partner
every fourth nonempty profile. Compare raw output exactly, reverse inputs and
verify input ownership. Check the derived digit/carry/query invariants; no sampled
success substitutes for their general argument. No production input domain is narrowed.

Independent oracles:

- Go: dense big.Rat from the binary64 inputs, exact decimal normalization and
  native correctly rounded Float64 conversion; no sparse algorithm is reused.
- TS: bounded dense integer sum, multiplied by5^52 to form an exact finite decimal
  expansion, then native Number parses its exact normalized text. It reuses no
  sparse borrow, sticky-bit or significand-rounding code.

Both original pair outputs become1.01245957117e1 and match the independent models.
The candidate retains the rejected aligned reducer's tiny cancellation tail.
Literal full-domain cases also cover a cross-exponent cancellation and long borrow,
even/odd ties with positive/negative far-away tails, recoverable overflow and exact
net overflow/underflow. The remote-above-half case would round differently if its
sticky information were lost; these checks compare raw mantissa bits, not only
canonical strings that would hide the difference.

All2048 complete input/output witnesses have SHA256:
`ef6cdc7626f91ff237961488f84f07a022b5ac746542d53dfb68fb258768047f`.
Native Darwin/arm64 Go1.27.2, declared Linux-amd64 Go, Node, and Linux Chromium,
Firefox and WebKit agree on that digest. Each language also passes its own dense
oracle. Full-domain borrow uses five digit events, one run and nine rounding reads;
the largest normalized-work sampled row has34 input chunks/20 normalized positions.
These observed counts confirm the examples, not a new measured production budget.

Cold active Decimal/vet passes0.449s. Full declared Linux-amd64 Decimal package
passes0.751s, including named prototype populations; verbose old vectors truncate
in tool output, so no complete per-vector skip census is claimed. Its accompanying
Postgres exists only because the standard test target starts it; no SQL behavior
is tested here. Node affected6412 and Linux browsers19236/15 files pass, as do
strict types/Svelte diagnostics. Native and Linux original grouped raw census
differs876 versus875 in the preceding aligned study; this candidate's full witness
digest agrees. No cause is assigned to the old raw-count difference.

## Actual Go caller impact, not a global patch

An ignored compiler overlay substitutes only `server/decimal/aggregate.go`.
All helper declarations match the retained test prototype after nominal
Research→Overlay suffix substitution; its wrapper panics on internal invariant
failure so measurement cannot silently relabel a bug as ordinary invalid data.
Overlay SHA256:
`f83a37cfbf80d52ad22a20db64f062b53be5a40c1c27e3b845bdc8787feba9fb`.

Actual Decimal/Economy packages pass0.835/9.655s. Full Production45.140s is red:
27 partition cases remain, the44-profile policy observer reports result drift,
three retained clock-reference payouts change and RP-664's actual upper evaluator
payout changes. No panic or rounding-width guard fires. No report/golden/hash is
refreshed. Those timings are not a controlled performance comparison. This proves
impact beyond the new helper tests, not an accepted kernel or end-to-end repair.

## Current caller boundary — engineering closure still needed

At35011479 there are13 direct Go and7 direct TS product calls, excluding declarations
and imports. A global replacement reaches more than accrual:

| Consumer | Current direct owner/counterpart | Compatibility question |
|---|---|---|
| Signed ledger net | `server/economy/ledger.go`; client replay applyLedger | Atomic refusal/receipts and restored ledgers; constructors currently have no engine/profile parameter |
| Constant accrual | `server/production/accrual.go`; `client/src/production.ts` | Published primitive results and actual producer arithmetic |
| Segment rate/delta | `server/production/engine.go`; client replay advance | One-shot/segmented receipts and saved transitions |
| Lucky payout | `server/production/active_play.go`; client replay claim | Rate-derived request/cap outcomes |
| Progress coordinates | `server/production/progress.go`; `client/src/economy-kernel.ts` | Gate/progress predicates and rounding |
| Rate reads/simulation | `server/production/rate_projection.go`, `simulation.go` | Display/context/simulation policy consistency |
| Catalog weight validation | `server/economy/catalog.go` | Existing stored/retuned artifact admission, not just saved balances |
| Commons capacity | `server/commons/health.go` | Shared world aggregate and its own identity/compatibility boundary |
| Pitch metric | `server/pitch/engine.go`; `client/src/pitch/engine.ts` | Separately pinned tenant engine1.0.0, including hack-produced fractional terms |
| Harness reports | `server/harness/harness.go` | Report identity, retained workloads and baseline protocol |

Do not invent a global compatibility fork merely because old outputs change.
Existing Company policy already allows continuing play on a new kernel while
recording immutable run_version_drift and excluding ranking; unmatched pinned
verification defers. See `server/save/runepoch.go`,
`server/replayverify/repository.go` and Run Genesis R2/R3. A genuine new kernel
must use that actual policy, not a silent version signal or rewritten records.
Conversely, that Company policy does not automatically dispose of separately
pinned minigame engines, catalog admission or harness reports. Each needs its
named compatibility disposition/proof; a kernel bump alone is not evidence.

Next: close this caller/version/activation packet, measure actual workload cost,
obtain the explicit aggregation follow-up and independent review, then implement
and prove the affected actual Go/TS, SQL/retry and player/replay boundaries. The
current accepted grouping recipe and production source remain unchanged. No
new save field, domain restriction, content mint, accounting/search adoption,
RP-663/AC6 repair, feature completion or release approval follows from this study.

## Caller compatibility findings after f3280ff2

### Pitch: unchanged-engine equivalence is disproved

A bounded search seeks one four-term committed disagreement among at most20000
seed665 profiles: four positive canonical12-digit base metrics, one positive
canonical factor, exponents−3..3, actual Decimal multiplication then each reducer.
It reaches its declared counterexample objective at profile384, not an exhausted
guard or a claimed complete census. The candidate result also agrees with the
independent rational model of those four represented terms.

`testdata/pitch/aggregation-compatibility-research-v1.json` retains the literal
bases/factor. Both actual catalog loaders admit the retuned content. Actual Go
registry Create/Apply and TS createPitch/applyPitch keep the seed2 hand; only hack
ownership is synthetic. Under unchanged engine1.0.0, the existing result is
5.97255552034e3. At that exact valid funding threshold the transition enters shop.
The actual Go compiler overlay instead yields5.97255552033e3 and stays playing;
the named current-engine witness fails with both differences. No tracked source
is mutated. This is an admitted transition, not evidence of reachable hack
acquisition, a published-content defect, SQL persistence or natural player loss.
TS checks the existing transition, not the candidate through the TS tenant.

The earlier ordinary Pitch package pass was therefore insufficient evidence for
unchanged semantics. Any replacement must either preserve engine1.0.0's grouped
scoring explicitly, or introduce a new engine identity with lawful pinned-session
handling. A Company kernel bump alone cannot rename this contract. The registry
currently keys one tenant per engine_ref and rejects duplicate refs; merely passing
two versions of pitch to NewTenantRegistry is not a coexistence implementation.
No new minigame version, content epoch or session migration is authorized here.

### Other activation boundaries

- Company: `requireRunEpochTx` already records drift while permitting commands;
  `VerifyStoredRun` returns engine_mismatch for actual drift and defers an
  unavailable pinned engine. Reuse this accepted policy, with the existing real-DB
  drift/deferral tests at activation; no perpetual per-run calculation fork or
  persisted arithmetic-mode flag is justified by this study.
- Catalog admission: composite weights use raw equality to One after the current
  grouped reducer. A new mathematical sum cannot silently redefine acceptance
  of old stored/retuned artifacts. The follow-up must explicitly preserve that
  grammar or specify and prove its compatibility; do not adjust weights for green.
- Commons: capacity is not just a pure diagnostic. `commonsprojection.refreshScope`
  parses persisted member samples and writes derived capacity to
  commons_health_scopes. A changed reducer needs real projection/replay evidence
  for that boundary; the integer-ppm health calculation is separate.
- Harness: RunKey contains scenario/policy/constants identities, not kernel version.
  Identical keys promise identical reports. Kernel bumps alone cannot justify
  restamping old reports; baseline updates require the existing input-first,
  separate-artifact protocol. A changed scenario must describe a genuinely changed
  experiment, not dummy hash churn. If code-only corrections cannot fit that
  protocol, the explicit follow-up must resolve it before artifacts are refreshed.
- Ledger/progress/rate/replay callers remain in the original table. A future shared
  arithmetic change must update actual Go/TS consumers together, retain atomic
  refusal and receipt reproduction, and prove the affected SQL/retry boundaries.

These are concrete engineering constraints, not extra owner product choices.
The activation packet remains incomplete until each disposition is specified and
reviewed. They do not block H3c's separate fixed-origin temporal proof by requiring
a false universal grouping lemma, nor do they adopt either accounting draft.

### Exploratory actual-caller cost

The retained manual BenchmarkAggregationCallerResearch executes AccrueConstant on
actual Rates output and Evaluate on a freshly restored save each iteration.
Epoch5 has one purchased class/one cash source. The existing controlled axis
fixture has nine purchased classes, twelve contributions across slots and eight
cash rates. Catalog parsing/rate construction is excluded; the Evaluate rows
include restoration plus the actual full interval/receipt credit. One-second and
24-hour offline-cap evaluations use the one-source catalog, not the axis workload.

Run each source with the same selector,100ms per sample,three samples per row:

```sh
make test-go GO_PACKAGES='./production' GO_TEST_FLAGS='-run=^$$ -bench=BenchmarkAggregationCallerResearch -benchtime=100ms -count=3 -benchmem'
make test-go GO_PACKAGES='./production' GOFLAGS='-overlay=/Users/stronk/repos/cloud-clicker/.cache/sparse-aggregation-go-overlay.json' GO_TEST_FLAGS='-run=^$$ -bench=BenchmarkAggregationCallerResearch -benchtime=100ms -count=3 -benchmem'
```

The double dollar is Make escaping, not a broadened test selector. The overlay is
the ignored research artifact described above, not a committed build input.
Two clean active/overlay rounds on Darwin arm64/Apple M3 Max produce these ranges
of per-round medians (not confidence intervals):

| Actual call | Active ns/op | Candidate ns/op | Active→candidate allocs/op |
|---|---:|---:|---:|
| Eight-source accrual |345–353|2961–2973|4→157|
| One-source accrual,1s |91–94|948–967|2→48|
| One-source accrual,24h |94–95|961–997|2→48|
| Restore + Evaluate,1s |4807–5213|7420–8396|82→220|
| Restore + Evaluate,24h |4788–5222|8168–9000|82→242|

An intervening active round overlapped package tests/vet and is excluded from
these comparisons, not silently treated as timing evidence. An earlier smoke run
predated the eight-source arm and is also not a matched comparison. All five
rows execute successfully in the retained comparison rounds. No pacing threshold,
CI timeout, workload population or gameplay acceptance bound is loosened.

The prototype has a substantial allocation/constant-cost penalty even though work
is independent of exponent distance. Consider proven-equivalent zero/single-input
fast paths and allocation reduction before adoption, then run the actual complete
producer/search workloads. These short native measurements are neither a production
budget, hosted-runner benchmark, full offline-axis test nor final performance proof.

## Bounded Go identity-path optimization — 2026-10-15

At baseline38dfc002, add only the exact length0/1 identities to the Go test prototype.
No inputs gives canonical positive zero. One input is first checked with IsStateValue;
invalid/diagnostic/unnormalized values still give NaN, either signed zero becomes canonical
positive zero, and other valid inputs return their unchanged represented value. The exact
represented sum of one such input is the input, so no rounding/normalization is needed.
Two or more inputs still use the unchanged full sparse algorithm, including zero padding
and invalid inputs. This is not a one-nonzero-term filter, changed quantizer or domain restriction.

The new raw-bit identity controls compare against a deliberately zero-padded general sum,
literal input identities and invalid/signed-zero/full-exponent boundaries. Cold full native
Decimal/vet pass. Existing14 edge controls,2048 independent-rational profiles, original pair
and retained Pitch disagreement pass in native Go and declared Linux-amd64 Go. Witness SHA256
remains `ef6cdc7626f91ff237961488f84f07a022b5ac746542d53dfb68fb258768047f`.
TS prototype is untouched; this is Go execution/equivalence evidence, not new browser proof.

Preserve the original overlay SHA f83a37cf… as the old arm. The separate optimized ignored
overlay `.cache/sparse-aggregation-single-candidate.go` mechanically copies all current helper
declarations with Research→Overlay suffixes and the same fail-loud wrapper. Its SHA256 is
`60dbfd2966571baa51a38054160c278fe24e494d05104849eb42d84b6679842a`.
The corresponding `.cache/sparse-aggregation-single-go-overlay.json` replaces only aggregate.go
for manual research builds, never tracked production. Actual Company/Pitch compatibility
tests fail identically on BOTH overlays: Company upper payout1.01245957117e1 instead of the
current1.01245957116e1; Pitch5.97255552033e3/playing instead of5.97255552034e3/shop. The same
two tests pass on active source. Cost reduction does not repair these incompatibilities.

Predeclared measurement: existing five actual-caller rows,100ms per sample,count3; two
sequential old→optimized→active rounds. All owned reviewer Go/browser/SQL work is terminal
before timing starts. All six arms/90 result rows complete on Darwin arm64, Go1.27.2,
Apple M3 Max. Ranges below are two per-round medians, NOT confidence intervals:

| Actual call | Old candidate ns/op | Optimized candidate ns/op | Active ns/op | Old→optimized allocations/op |
|---|---:|---:|---:|---:|
| Eight-source accrual |3034–3159|3048–3199|356.9–360.9|157→157|
| One-source accrual,1s |1021–1028|57.47–57.97|97.87–98.10|48→0|
| One-source accrual,24h |1030–1041|62.01–62.54|99.44–101.6|48→0|
| Restore + Evaluate,1s |8343–8441|4853–5004|5081–5216|220→76|
| Restore + Evaluate,24h |9137–9341|4912–4975|5035–5146|242→76|

Single-source accumulator cost is removed, not merely hidden by shorter work. Full interval/
receipt credit and fresh restoration checks remain in every benchmark iteration. Eight-source
cost/allocation remains substantial; the one3888ns sample remains in the raw record, not dropped.
Raw commands/outputs are ignored `.cache/sparse-aggregation-identity-bench-38dfc002.json`, SHA256
`19026d95da52dfd9a21e14deb0ba847f111a28745b0b9ead63f37eec588a868d`.
Run the existing benchmark command above with that second overlay for the optimized arm;
the tracked helper and baseline38dfc002 retain both code versions independently of cached files.

Disclosures: an initial overlay selection wrongly ran the current-versus-candidate search with
both arms replaced, so its no-disagreement refusal is preserved, not correctness evidence.
A first narrowed regex loses its dollar through Make and runs no tests; corrected selector
passes. Hardware sysctl is sandbox-refused; actual Go benchmark output supplies CPU identity.
The declared Go service starts an otherwise unused disposable Postgres; no SQL is tested here.
No thresholds, population, golden, report, CI, active arithmetic or kernel version changes.
This is author-verified prototype work, NOT designated review, adoption or AC6 repair.
Next is proven-equivalent multi-input allocation reduction and full affected workloads, plus
the still-required caller/version/baseline dispositions and normative review. H3c's fixed-origin
proof,27 accounting failures, RP-663 and full1.0 obligations remain separate and unresolved.
