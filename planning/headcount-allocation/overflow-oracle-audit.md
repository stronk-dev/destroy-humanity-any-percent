# RP-413 — overflow oracle finding

2026-10-13; source/test baseline `bd46e71e`. Research, not an adopted amendment.
Owning authority: Headcount S2/S4/S8 and AC3. Production source, accepted body,
shared vectors and assertions stayed unchanged throughout.

## Question and method

Before measurement: determine whether replacing only the TS guarded capacity
expression with `Math.min(converters * k, heads)` can cause a wrong overflow
result for an admitted assignment. Compare the existing literal vectors and
independent BigInt model in Node, Chromium and WebKit; retain invalid-input
checks. A separate mutation ignoring `k` must fail the existing ratio vectors,
otherwise the instrument has not demonstrated discrimination. Compare the Go
multiply-before-clamp mutation with its existing arbitrary-precision oracle.
This cannot authorize a seat source, output economy, integration or acceptance
promotion. Samples alone cannot establish full-domain equivalence.

## Result and derivation

S8's required `C=3.1e15`, `E=9e15`, `k=3` assigns `12.1e15` seats, above
`S=9007199254740991`. Both actual kernels correctly reject it. The admitted
`C=3.1e15`, `E=5.9e15`, `k=3` vector diverts exactly `5.9e15` heads; the naive
TS expression does too.

Let `0 <= E,C <= S`, `1 <= k <= S`, and exact integer `P=C*k`. ECMAScript
Number multiplication rounds the exact product to binary64; integers through
`2^53` are representable. [Primary specification, Number and Number::multiply](https://tc39.es/ecma262/multipage/ecmascript-data-types-and-values.html#sec-numeric-types-number-multiply).
The following is our derivation, not a sampled universal claim:

- If `P <= S`, the product is representable exactly, so clamping gives the exact
  integer diversion.
- Otherwise `P >= 2^53`. Since `2^53` itself is representable, rounding cannot
  produce a value below it. But `E <= S < 2^53`, so clamping returns exactly `E`.
  Also `P <= S^2 < 2^106`, excluding floating-point infinity.

Thus an unsafe intermediate product cannot create AC3's specified TS overflow
output failure within this domain. This concerns integer diversion, not
bitwise equivalence of all JavaScript representations: `-0` is accepted by the
current predicate, and signed-zero cases are not an overflow witness. It does
not authorize removing the guarded algorithm or changing admission.

## Executed evidence

Temporary TS source clone was first byte-identical, then differed only at the
capacity expression. The test clone differed only in its module import; it
retained all existing assertions and the shared JSON. Go used a compiler overlay
with only the capacity block changed. Production files were never mutated.

| Population | Result |
|---|---|
| TS naive multiply/clamp, existing tests plus actual implementation, Node | PASS 84 (42 per implementation) |
| Same comparison, Chromium/WebKit | PASS 168, session56711, 1.61s |
| TS ignore-`k` control, same Node population | Exit1: 8 mutant failures, 76 passes; actual implementation42 passes |
| Same control, Chromium/WebKit | Make exit2: 16 mutant failures, 152 passes, session81683; actual implementation84 passes |
| Go naive `min(converters*k, headCount)`, cold `TestExactMaximumDomain` | Make exit2; `E=0`, `C=k=S` produces diversion `-18014398509481983`, producing heads `18014398509481983`, idle converters `9007199254740992` |
| Actual Go, cold complete Headcount package; vet | PASS0.121s; vet exit0 |
| Actual Node after probe removal | PASS42, 0.247s |
| Actual Chromium/WebKit after probe removal | PASS84, session33398, 1.43s |

The wrong-ratio control's first `k=3` literal expected diversion6, idle0 and
producing2; received diversion2, idle1 and producing6. Eight failures include
five literal vectors, both exact-model populations and independent conversions.
The Go overflow failure occurs on an admitted assignment and is not an invalid
fixture rejection. These are distinct faults, not interchangeable proofs.

Reproduce with clones of `client/src/headcount/allocation.ts` and
`client/test/headcount-allocation.test.ts`, changing only the test import and the
one capacity expression to each mutation above. Run root Vitest on original plus
clone and `make test-browser-focused` selecting both files and Chromium/WebKit.
For Go, copy `server/headcount/allocation.go`, replace its guarded capacity block
with the expression above, and supply an absolute-path Go overlay to root
`make test-go GO_PACKAGES='./headcount' GO_TEST_FLAGS='-count=1 -overlay=PATH -run TestExactMaximumDomain -v'`.
Check each clone diff before running. Never edit production for these probes;
remove the temporary test files after each terminal run. No duplicate mutant
population is added to normal CI.

The first native browser attempt stopped at localhost `EPERM` before tests.
Some subsequent tool outputs were lost during context recovery and are not
counted as evidence. Read-only process inspection confirmed no matching test
processes remained before the recorded replacement runs; this was not a restart
on an observation timeout. All recorded runs above are terminal.

## Required next action

The ruling/specification author must reconcile S8/AC3: retain the inadmissible
example as a refusal, use admitted boundaries for exact-output checks, and
distinguish the demonstrated Go overflow fault from the demonstrated TS ratio
fault. If the TS algorithmic ban on unsafe intermediate products is retained,
give it an explicit mechanical witness rather than an impossible overflow output
failure. This is a proposed repair route, not an adopted replacement criterion.

AC3/AC5 stay open. Firefox, Linux/hosted CI, live production/replay, catalog/save,
UI and pacing are not proven here. Earlier cross-party review remains outstanding.
