# RP-664: aligned compensated aggregation experiment

Status: candidate **rejected**, production unchanged. Source baseline `187cf426`.
This is numeric engineering evidence, not an adopted accounting algorithm.

## Question and declared boundary

Can aligning original Decimal terms to their largest exponent and applying
Neumaier compensation repair RP-664 without losing existing aggregation behavior?
The prototype sorts a copy by exponent/absolute/signed mantissa, scales each
original term once, compensates float additions, then normalizes once.
Executable counterparts: `server/decimal/aggregation_candidate_research_test.go`
and `client/test/numeric-aggregation-candidate-research.test.ts`.

Before measuring the candidate, declare the existing RP-664 pair, signed/domain
controls, existing accrual fixtures and seeded positive-input search. Search:
100,000 complete profiles; uint64 xorshift seed663; four positive 15-digit
mantissas at exponents−5..8; variable9.999999999999998→10. Count raw and Quantize12
reversals separately. Invalid outputs invalidate measurement; no exclusion,
truncation or changed operating limit. No sampled absence of reversal proves
global monotonicity. Add the tiny surviving cancellation control before executing
it, after inspection identifies scale-alignment underflow as a possible failure.

## Results

Independent exact rational sums of the **represented binary inputs** in the pair:

- lower: `2279856273713199179121 / 225179981368524800000`;
- upper: `2279856273713199579121 / 225179981368524800000`.

Both exceed decimal midpoint10.12459571165. Python Fraction/Decimal diagnosis
and Go big.Rat agree. This bounded mathematical diagnostic does not replace
RFC-0001's actual float-mantissa quantization contract.
The candidate returns1.01245957117e1 for both, unlike current upper1.01245957116e1.
All eleven valid existing accrual-fixture formula results and initial signed/
invalid/out-of-domain controls pass. The client formula check substitutes only
the sum; it is not actual client caller migration.

| Runtime | Current raw / committed reversals | Candidate raw / committed reversals |
|---|---|---|
| Native Go | 876 / 1 | 0 / 0 |
| Node, Linux Chromium, Firefox, WebKit | 875 / 1 each | 0 / 0 each |

An initial JS assertion incorrectly requires Go's raw count876 and fails with875.
Raw continuous arithmetic has a tolerance in RFC-0001; an observed Go census is
not an exact JS contract. Preserve/report each census instead. Committed counts
and the original literal pair remain exact; no ordering tolerance is added.
The raw-count discrepancy is not separately cause-attributed or hidden.

The decisive signed control rejects the candidate in both runtimes:

`1e300 + (-1e300) + 1e-300`

Current grouping cancels the large same-exponent terms and retains1e-300.
The candidate first scales the tiny input by10^-600; it becomes zero and cannot
be recovered after cancellation. Candidate result0 loses an existing successful
behavior. A compiler overlay replacing the actual exported Go reducer with the
prototype also fails the unchanged current-result assertion. Do not ship this
candidate, special-case the example, or restrict signed inputs to save it.

## Actual Go caller overlay, not active implementation

Compile the prototype as `decimal.SumDeterministic` through Go's source overlay;
its body is byte-identical to the committed test helper (name/comment excluded).
Local ignored source `.cache/aggregation-candidate.go` SHA256:
`4855c5b04a0fdbc3ef071f0960a06f98ea55563fadaa07b4a371ba7e65719402`.
Overlay maps only `server/decimal/aggregate.go`. The retained helper body specifies
the reproducible candidate; the ignored overlay is not a second production authority.

Cold actual Decimal/Economy packages pass under the overlay. Full Production is
red: all original27 partition cases remain, the44-profile policy observer reports
result drift, all three retained clock-reference pairs change, and the original
RP-664 actual Rates/Evaluate witness changes its upper payout. No source hash,
report or golden is refreshed. Fingerprinting reads active source originals while
compiled candidate behavior is deliberately different. This is compatibility-impact
evidence, not an accepted version or a repaired public replay.

Original-source policy observer, three clock-reference pairs and actual
Rates/Evaluate witness pass afterward, preserving previous outcomes. No live file
is patched/restored: only compiler input is substituted.

## Consequence and remaining work

Positive sampling plus correction of RP-664 is insufficient. A viable candidate
must preserve low-order contributions through cancellation over the admitted
exponent range, justify work bounds, refuse invalid/net-overflow outcomes, preserve
input/permutation safety and prove exact cross-runtime committed results. Do not
repeat this positive census instead of solving those requirements.

Deterministic Aggregation explicitly specifies exponent grouping and low-to-high
combination; replacing that recipe needs a reviewed/adopted follow-up under
RFC-0000. Numeric Core's monotonicity obligation does not silently authorize
changing another accepted algorithm or old replay outputs. First produce a viable
contract and compatibility/activation route. No live algorithm, kernel identity,
Clout settlement/search proposal, balance or content change is authorized here.
RP-664, RP-663 and the27 partition failures remain open.
