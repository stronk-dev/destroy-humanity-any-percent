# RFC: Explicit Integer Token Encoding in API Schemas

- **Status:** draft — RP-566; normative adoption and designated review pending. Not implementation authority.
- **Author:** Codex (engineering proposal for Marco)
- **Created:** 2026-10-14
- **Design refs:** `design/06-tech.md` (cross-runtime parity and versioned contracts)
- **Parent / amends:** API Foundation A5, C17 and C18; preserves A1/C2 and C19.
- **Planning:** existing `planning/api-foundation/` plan/log; no separate program.

## Summary

Make integer token grammar explicit in the closed Go schema descriptor. Keep every existing
integer's plain signed-int64 behavior by default; add one opt-in `exact-safe` encoding for the
relevance owner's already-admitted nonnegative safe integral decimal/exponent tokens. This is
an engineering representation proposal, not a request to choose new arithmetic or retune the
game. Adoption is still required before extending C17's descriptor contract.

## Motivation and executed evidence

At `bd0c9b97`, the actual relevance loader accepts `1.0`, `1e0` and
`9.007199254740991e15`, while publicapi's `SchemaInteger` refuses them through
`strconv.ParseInt`. Plain integers and the plain safe maximum pass both; fractional and unsafe
controls fail both. The API log and `probes/relevance-integer-probe.go.txt` retain that measured
seven-case boundary. RP-567 separately repairs quoted-number/array-axis admission under the
existing R12 contract at `f900340d`; quoted numbers are not valid alternate spellings.

The existing shared relevance mutation corpus additionally executes token lengths 64/65 and
exponents ±32/±33 in both runtimes. The boundary cases are mathematically equal to 1, so a domain
overflow cannot masquerade as a lexical-limit refusal. This proposal preserves those current
limits rather than choosing a new parser budget. The base corpus also covers integral decimal
versions, exponent versions/epsilons, the exact safe maximum and rounded fractions.

JSON Schema's integer type describes the mathematical value, not whether a decimal point was
written. Its [numeric reference](https://json-schema.org/understanding-json-schema/reference/numeric)
explicitly treats 1 and 1.0 alike. Therefore the existing plain-token restriction and the proposed
bounded token grammar require explicit repository metadata; neither can be represented by
ordinary numeric bounds alone. Globally widening all current integer validators is excluded:
the existing formula owner deliberately refuses decimal integer tokens and uses signed int64.

## Specification

### IE1 — Descriptor and construction

Add `type IntegerEncoding string` and `IntegerEncoding IntegerEncoding` to `publicapi.Schema`.
The only nonempty legal value is `IntegerExactSafe = "exact-safe"`. An empty field preserves
the existing plain JSON integer token plus signed-int64 domain, including its existing `-0`
behavior. There is no explicit second spelling for the default, custom regex or generic
JSON-number arm.

Only `SchemaInteger` permits the field. All other kinds refuse it, including ref/oneOf
wrappers: annotate the actual integer child, not its container. Existing integer field
exclusions remain. `exact-safe` additionally requires both bounds, with
`0 <= Minimum <= Maximum <= 9007199254740991`. Invalid modes, absent bounds and out-of-domain
bounds fail schema construction, not first use. Existing default integers retain optional
bounds and the full int64 domain.

Clone/export the encoding with the existing immutable descriptor tree. Changes to an input
descriptor or an exported nested descriptor must not mutate an already-built registry.
Reference/depth/cycle laws remain unchanged; resolving a ref does not discard its target's mode.

Path/query parameter registration refuses `exact-safe`, including a reachable integer behind
a permitted reference. Their current URI grammar is unchanged. This does not reinterpret
epoch IDs, cursors or route components as decimal/exponent strings.

### IE2 — Exact validation, not floating-point rounding

For an `exact-safe` child, require a JSON number value and retain its original lexical token.
Strings, booleans, null, arrays and objects refuse. Parse JSON syntax first; no alternate radix,
leading plus, invalid leading zero, NaN, Infinity or trailing syntax is accepted.

Apply the following current relevance grammar before any binary floating-point conversion:

1. Number token is at most 64 ASCII bytes, excluding surrounding JSON whitespace.
2. If present, exponent has at most two digits after its optional sign and value in [-32,32].
   Thus `1e00` is allowed and `1e000` refuses even though both mean 1.
3. Its exact decimal value must be integral, nonnegative and at most 9007199254740991 (S).
4. Apply the descriptor's inclusive integer bounds to that exact value.

For bounds [1,S], `1`, `1.0`, `1e0` and the exponent-form safe maximum pass. Quoted equivalents,
`1.0000000000000001`, the value S+1 and mathematically integral but over-limit lexical forms
refuse. A negative zero is the mathematical value 0 and passes only if the declared bounds
include 0; this does not affect the relevance version/epsilon fields, whose lower bound is 1.

Use exact integer/rational or coefficient-and-scale arithmetic, not float64 followed by an
integrality test. Validation never rewrites tokens, source bytes, stored artifacts or hashes.
Default integer validation remains its current raw plain-token/int64 path.

`ValidateJSON` retains its existing artifact last-value member semantics. Registered HTTP
request/response/literal admission retains its additional raw duplicate-key check. This scalar
amendment does not change either layer's object grammar or imply that a descriptor alone proves
the owner's semantic policy.

### IE3 — Generated representation and compatibility

Emit `type: "integer"` and the existing inclusive bounds. For `exact-safe` only, also emit
`"x-cloud-clicker-integer-encoding": "exact-safe"`. The extension names the closed IE2 grammar;
it is not a free-form metadata input. OpenAPI permits
[specification extensions](https://spec.openapis.org/oas/v3.1.1.html#specification-extensions).
Generic JSON Schema validators need not enforce this raw lexical restriction; the repository
validator does. Do not claim that OpenAPI alone preserves original numeric token spelling.

Generate TypeScript `number`, since every admitted value is exactly representable. A TS type
does not enforce bounds or prove raw-token validation. No new client metadata exporter or
runtime decoder is introduced by this scalar amendment. A later client descriptor exporter
that consumes annotated schemas must carry the mode; a validator consuming such metadata must
check the raw token before JSON.parse can erase an invalid fractional spelling. No handwritten
per-field allowlist or Number.isSafeInteger-only substitute is allowed.

This amendment does not add the general generated success-body guard proposed by the separate
tenant-binding draft, or relabel the existing SDK's JSON decoder as such a guard. Later C18
consumer integration must bind actual owner validation to the downloaded artifact representation
and preserve source identity; static DTO generation is not its acceptance proof.

`CanonicalOperationPins` already serializes generated schemas; it must retain the extension.
Compatibility compares effective mode (absent means current default) at every nested integer,
including refs, arrays, objects and union arms. Any mode change in an existing used schema
refuses for both request and response use, even with unchanged bounds. A present null, unknown
or nonstring extension is invalid, not silently treated as a default. Ordinary new schemas/operations remain
subject to the existing C2 rules; pin comparison must not ignore the new keyword merely because
the mathematical type is unchanged.

Default descriptors must produce byte-identical OpenAPI, TS and compatibility pins. Do not
reset pins, increment the generator version to bypass comparisons or annotate every old
integer gratuitously. No existing operation opts in as part of this scalar amendment.

### IE4 — C18 consumption and rollout boundary

After this amendment is accepted/implemented, the relevance owner can describe its version
and item/group epsilon fields using `exact-safe`: version bounds 1..1 or 2..2 in disjoint v1/v2
arms; positive epsilon bounds 1..S. The v1/v2 availability/null differences and all other exact
fields still require complete owner descriptors. Cardinality, ordering, catalog membership,
gate/group symmetry and other owner semantics remain loader checks, not implicit DSL features.

This proposal does not export that descriptor, mount a partial catalog response, mint an
artifact, normalize historical JSON or widen a request/tenant union. All remaining artifact
owners are kernel-watched; their API-only descriptor additions still need the separately
recorded scope resolution. This amendment is not permission for a false kernel bump or moving
owner code out of a watched path. Scalar support in publicapi does not itself change replay.

## Deviations from design

None. Extends C17's explicit typed descriptor, preserving existing API and owner behavior.
No gameplay, numerical production formula, balance, retention, copy, deployment or CI change.
The proposed field is unadopted normative authority, not a routine bug-fix shortcut.

## Acceptance criteria

1. Construction rejects unknown/noninteger modes, missing/out-of-domain/reversed bounds and
   use in path/query parameters. Existing default definitions remain valid. Input/export
   mutation controls prove fresh immutable encoding and nested descriptor storage.
2. Actual registry/direct validation executes representative roots, nested fields, refs,
   arrays and oneOf branches with plain, decimal, exponent, safe-limit, rounded-fraction,
   quoted, malformed, overlength and exponent-boundary inputs. Default int64/formula behavior
   remains unchanged; raw duplicate-key HTTP refusal remains unchanged.
3. Reuse the existing shared numeric controls for annotated scalar descriptors and the actual
   Go/TS relevance loaders. Do not treat unrelated semantic-rejection cases as scalar proof.
   Both sides must agree on numeric admission, with original bytes/hashes unchanged.
4. Actual-registry generation emits the annotation only where declared and TS number types.
   Default fixture generation and current committed API files remain byte-identical. A client
   metadata consumer cannot silently omit or ignore a declared mode.
5. Pin tests reject mode changes in both directions, through every referenced/nested shape and
   on request-only, response-only and shared schemas. Unknown/nonstring annotations refuse;
   unchanged-mode and ordinary C2 additive controls pass without a pin reset.
6. Execute affected cold Go/vet, generated/client checks when their boundary changes and
   `make api-check`. Record existing unrelated red lanes rather than concealing them. Canonical
   docs and the owning log reconcile the complete implementation range. Specification review
   precedes acceptance; designated cross-party approval of the complete implementation precedes
   closeout. No full C18 or1.0 claim.

## Open questions

No gameplay choice is requested. Normative adoption and independent review of this recommended
representation remain open. Owner-descriptor kernel scope, the complete C18 public-byte/client
loop and other API amendments remain named separate prerequisites, not silent defaults here.

## Changelog

- 2026-10-14: source-backed proposal for RP-566, preserving default integer behavior and measured
  relevance limits. Six additional shared boundary controls execute before proposing the limits;
  this document and those test-only additions do not implement or adopt the new encoding.
