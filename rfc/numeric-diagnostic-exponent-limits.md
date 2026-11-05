# RFC: Numeric Power Guards & Diagnostic Exponent Limits

- **Status:** draft — proposal, NOT implementation authority
- **Author:** Codex
- **Created:** 2026-10-15
- **Design refs:** `design/06-tech.md` (big numbers and server-authoritative numeric state)
- **Parent / amends:** [Numeric Core](archive/0001-numeric-core.md), diagnostic classification and operation boundary in [Boundary Hardening](archive/numeric-core-boundary-hardening.md)
- **Depends on:** existing state/wire guards and RP-663 conformance corrections
- **Planning:** existing `planning/clout-v1-and-pr-interns/` RP-663 evidence; implementation plan after acceptance

## Outcome and evidence

Guard power/exp results while the operation still has its operands. The native
Go boundary must refuse lower exponents it cannot represent. The pinned client
also loses some enormous nonzero results to canonical zero, in either direction;
its product callers therefore need an explicit checked-operation boundary.
Keep the library pinned, preserve existing valid results and base-zero
compatibility, and reject lost magnitudes before they become payments or wire
values. Neither this proposal nor its manual model is implementation authority.

The original RP-663 evidence uses `MAX=8999999999999999`:

- `1e-MAX ^ 2048`: native Go returns valid zero; pinned client retains finite
  raw(1,-18431999999999997952), invalid state/wire.
- `exp(-1e20)`: native Go returns valid zero; pinned client retains finite
  raw(1,-43429448190325178368), invalid state/wire.

Those exponents cannot fit Go's int64 field. Independent review then rejects
this draft's original Go-only solution: admitted finite `1e-MAX ^ 1e308` and
`exp(-1e309)` produce canonical valid zero in the pinned client too. Five
reviewed growing-magnitude counterparts also produce client zero, whereas Go
returns the proper infinity diagnostic. The [owning log](../planning/clout-v1-and-pr-interns/log.md)
records exact operations, controls and the original negative verdict.

Value-only validation cannot distinguish these lost results from legitimate0.
A made-up finite exponent or a larger representation is not an authorized fix
either. This revision explicitly proposes operand-aware guards; it no longer
claims unchanged client guards can discover provenance.

## Proposed contract

### N1 — Unchanged value domain; explicit finite-operation profile

Keep gameplay exponents exactly `[-MAX,MAX]`, strict constructors, state/wire
guards, canonical grammar and Quantize12. Canonical0 remains valid. No provenance
fields, expanded Decimal representation or changed library version.

The override profile is a lawful normalized finite source: admitted state or
an existing lower-domain finite diagnostic, with finite mantissa, finite integer
exponent, and `1 <= abs(mantissa) < 10` for nonzero values. Canonical zero is
also a valid source. NaN, infinity/sentinel and malformed representations are
outside this finite profile; preserve their existing diagnostic behavior and
all existing feature-level domain validation obligations. They do not become
legal gameplay operands. Existing cross-runtime obligations outside this change
remain obligations; a newly found discrepancy must remain a separately recorded
defect, not be blessed as compatibility by this draft.

For power the profile also requires a primitive finite floating argument and a
real operation: positive base, or negative base with an integer power. Base-zero
compatibility is preserved separately, including0^0 and0^negative. Introduce no
new finite-input or power ceiling. Wrong-type TypeScript power arguments return
diagnostic NaN without coercion; existing nonfinite numeric-argument diagnostic
controls remain required. Feature callers still validate operation domains.

### N2 — Evaluate first; guard before losing operand context

Native Go keeps its public `Pow`/`Exp` methods and adds these guards. TypeScript
adds `powChecked(base: Decimal, power: number)` and `expChecked(value: Decimal)`
to the existing numeric module. They call the unchanged pinned operation on a
clone, then apply N3. No monkey-patching, subclass, alternate exponentiation
engine or caller-selected policy flag. Raw pinned operations remain available
for explicit diagnostic characterization.

Preserve existing evaluation/normalization order. Do not pre-refuse an exponent
product merely because it leaves gameplay range: coefficient normalization or
later arithmetic can rescue a representable result. Do not pre-round arguments,
mutate inputs/provider results or invent a finite amount. The guard runs before
a product caller can consume a lost zero. Exp must retain its original Decimal
source even if an internal float conversion overflows to infinity.

### N3 — Operand-aware disposition of a lost nonzero result

A raw canonical zero from a finite-profile power with **nonzero base** has lost
magnitude; it is not a genuine zero. Compare `abs(base)` with1 using Decimal
ordering, without converting the whole value to an ordinary number or using an
approximate logarithm:

| Operand direction | Checked disposition |
|---|---|
| `abs(base)<1 && power>0`, or `abs(base)>1 && power<0` | NaN diagnostic |
| `abs(base)>1 && power>0`, or `abs(base)<1 && power<0` | Existing infinity diagnostic; negative only for negative base and odd integer power |
| Unit magnitude or power0, yet provider returns zero | Unexpected provider result; refuse as NaN, never fabricate a finite replacement |

For exp of a finite-profile source, raw zero is lost magnitude too: negative
source → NaN; positive source → positive infinity diagnostic. Source0 must still
produce1. A proposal/acceptance experiment observing unexpected zero at unit
magnitude, power0 or exp0 must fail as an unexpected execution, rather than
count it as an expected underflow/overflow case.

Base-zero powers and existing true-zero/domain controls keep their explicitly
accepted behavior. Division is unchanged. Proper Go upper infinity results stay
untouched, as do nonzero client diagnostics; this is not generic reclassification
of every invalid value. Client infinity constructors use the existing excluded
sentinel only diagnostically, never as a clamp, stored substitute or payment.

### N4 — Narrow native representation exception

For operations in N1's finite profile, when the selected lower path's
**normalized** exponent cannot fit int64, return
Go NaN, including negative-infinite overflow reached from finite operands.
Check bounds before conversion; never rely on wrap or architecture-dependent
float-to-int conversion. Merely leaving gameplay range is not exhaustion;
preserve finite lower diagnostics while int64 holds them.

Where the client retains a finite invalid diagnostic beyond native int64, its
checked result can remain finite while Go is NaN. Shared fixtures name both
classes separately; both must fail actual state/wire validation. The deliberately
irreversible Go result can prevent recovery in a longer arithmetic chain that
the client could retain. No full-chain equivalence is promised after native
representation exhaustion. The separate client-valid-zero population requires
N2/N3 and cannot be dismissed by citing this native-only exception.

### N5 — Bind the actual callers

After adoption, migrate product TypeScript Decimal pow/exp calls through the
static checked exports. At kernel229, direct scanning finds three pow calls in
`client/src/economy.ts`, two in `client/src/commons.ts`, and no direct product
exp call. This is a starting inventory, not proof of indirect coverage: rederive
method/static/aliased/helper callers at implementation and bind each explicitly.
Do not leave an unchecked product alias as an escape. Raw pinned calls remain
permitted in diagnostic characterization and the historical vector generator;
those are not product consumers.

Native callers continue through the now-guarded native methods. Actual native
Ledger.Apply and ApplyAccrual must refuse mixed invalid transactions atomically,
preserving complete prior snapshots/receipts. Feature handlers still validate
domains before calculation. Neither raw diagnostic compatibility nor a checked
helper authorizes an otherwise-invalid operation or a silent zero fallback.

### N6 — Identity and authority

Keep `break_infinity.js`2.2.0. No layers, global arithmetic switch, extra numeric
fields, save/wire format, changed prices/rates or hidden tolerance. This is an
intentional behavior/diagnostic contract change requiring the actual next free
kernel version, normative review/adoption and full-range implementation review.
No version is reserved here. A manual policy projection is unapplied research;
it cannot satisfy implementation acceptance or authorize a caller migration.

## Deviations from the old diagnostic contract

The raw pinned client ceases to be the unchecked oracle specifically for finite
lost-magnitude zeros: the project checked boundary intentionally refuses them.
Some formerly accepted zero results therefore change. Separately, Go may return
NaN where the client retains an already-invalid finite exponent beyond int64.
The state domain and rounding stay unchanged; canonical0 is not outlawed.
Archived parents remain frozen; canonical docs change with implementation only.
This draft does not waive any other numeric or release requirement.

## Acceptance criteria after adoption

1. Permanent failing-first regressions reproduce both original native zero cases
   and all seven reviewed client lost-zero cases. Preserve raw-provider evidence
   separately from expected project behavior. Shared fixtures name native/client
   checked classes, actual state/wire refusal, proper signed upper infinities
   and the finite-client/native-NaN exception explicitly. Put these new cases in
   a separately versioned supplement; do not rewrite the historical raw-provider
   oracle or reuse its unchecked result as the expected checked classification.
2. Exercise int64's edge/floating neighbors, integer/fractional/negative bases,
   exp, finite-source float overflow, coefficient-normalization rescue and lower
   diagnostic recovery before exhaustion. Preserve0^0/positive/negative, division
   compatibility and all unchanged results; no unsafe early exponent test or
   blanket classification-only weakening. Cover wrong-type TS power arguments
   without coercion and retained nonfinite numeric-argument diagnostics.
   Keep all6296 old golden rows and their assertion/tolerance semantics unchanged;
   this does not demand new cross-runtime bit equality for transcendental cases.
3. Execute actual checked exports AND real Economy/Commons callers. Replacing a
   test helper alone cannot satisfy N5. Preserve representative admitted caller
   outcomes; distinguish direct numeric extreme-input evidence from naturally
   reachable gameplay. Actual native mixed Ledger/ApplyAccrual refusals preserve
   state/receipts. No SQL guarantee follows from a native ledger check.
4. Cold Go/Node shared numeric suites, Chromium/Firefox/WebKit, vet, strict types,
   deterministic golden regeneration and the documented fuzz lane pass. Execute
   Linux-amd64 overflow/conversion cases and affected consumers. Preserve/disclose
   existing red lanes; do not regenerate numerical research or relax bounds.
5. Update canonical docs, RP-663 and owning records in the behavior batch. Obtain
   normative review/adoption first, then independent implementation review.
   Review delegation does not adopt either override or the checked client API.
   AC6/partitions, RP-664, hosted CI, full1.0 and release gates remain separate.

## Bounded pre-adoption model

`planning/clout-v1-and-pr-interns/probes/checked-power-policy.mjs` evaluates the
actual pinned provider and project state/wire helpers, then separately projects
N3's proposed disposition. The source-bound observation uses953 existing
pow/exp golden inputs,13 retained lower-domain cases and24 explicit extreme/compatibility
cases. It preserves966 actual raw results and refuses seven reproduced provider-valid
zeros (two shrinking, five growing), with no unknown execution or input mutation.
This is a **client-only policy projection**, not implemented checked exports,
native code, indirect-caller proof or a full-domain theorem. N4's native change
and all implementation acceptance above remain required.

Reproduce without overwriting the retained artifact:

```sh
node --experimental-strip-types planning/clout-v1-and-pr-interns/probes/checked-power-policy.mjs
```

The small observation artifact is
`planning/clout-v1-and-pr-interns/checked-power-policy-observation.json`.
For exact byte reproduction, use source `06da156d` and locked dependencies
matching the recorded hashes. On a later runtime, retain this historical
observation and write any newly bound result separately; metadata may differ.
Passing this model does not adopt the proposal or repair today's runtime.

## Open decision before acceptance

Adopt N1–N6's checked-operation boundary plus native NaN exception, or require an
alternative complete representation/compatibility design? Any alternative must
also fix the demonstrated client lost-zero population; widening Go alone cannot.
The checked boundary is recommended but unadopted. Keeping false zero as a valid
payment or weakening a release criterion is not an alternative.

## Changelog

- 2026-10-15: initial native representation-limit proposal; no implementation.
- 2026-11-05: forward-correct after independent CHANGES REQUIRED. Distinguish
  client lost-zero underflow/overflow from native representation exhaustion;
  specify operand-aware checked exports and actual caller binding. No implementation.
