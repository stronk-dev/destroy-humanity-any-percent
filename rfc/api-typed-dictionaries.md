# RFC: Typed Dictionaries in API Schemas

- **Status:** draft — RP-391; owner choice and designated review pending. Not implementation authority.
- **Author:** Codex (proposal for Marco)
- **Created:** 2026-10-08
- **Design refs:** `design/06-tech.md` (server authority, versioned saves and string numbers)
- **Depends on / amends:** API Foundation A2/A5, C7/C9/C17; preserves A1/C2 and C18.
- **Planning:** existing `planning/api-foundation/` plan/log; no separate execution program.

## Summary and decision

Add one key-constrained, value-typed dictionary arm to the existing Go schema language.
This is the recommended RP-391 option: describe the actual saved intent receipts rather
than introduce another HTTP projection or change replayed bytes. Marco must choose this
option or the versioned-projection alternative before acceptance; drafting is not adoption.

`production.wireSnapshot` includes resource/generator/transition/meter dictionaries, and
`founderCosmeticReceipt.cosmetics.equipped` is keyed by runtime pet UUIDs. Enumerating today's
fixture keys would misdescribe valid future/historical data. An arbitrary JSON arm would
discard the schema guarantee. Both are excluded.

## Scope

This amendment owns the descriptor, validator, generator, compatibility and immutability
support. It does not register only a subset of `/api/v1/intents`, change its handler, rewrite
stored receipts, add intent kinds, change errors, or alter gameplay/save/replay semantics.
The complete intent union and caller migration remain API Foundation work after this
prerequisite and any remaining contract/guard issues are resolved.

C18's public catalog response remains a name-discriminated, owner-described union. This is
not permission to replace its artifact descriptors with a generic dictionary. Public/private
flags, authorization, C2's request-enum law and original response bytes remain unchanged.

## Specification

### TD1 — Closed descriptor

Add `SchemaDictionary = "dictionary"`, `KeyFormat string` and `Values *Schema` to the current
schema DSL. A dictionary requires exactly one nonempty key format and a valid value descriptor.
The initial key-format set is exactly `mechanical-id` and `uuid`, using the existing string
format grammars in `publicapi/schema.go` without widening or narrowing either. There is no
arbitrary-regex, unrestricted-key, reflection, `any` or implicit-value arm.

For a dictionary, `Fields`, `Items`, `Enum`, `Minimum`, `Maximum`, `Format`, `Ref` and
`Alternates` must be absent. Every other schema kind rejects `KeyFormat`/`Values` when supplied.
`Values` may use the existing closed kinds or another dictionary; references, cycle detection,
depth checks, cloning and immutable registry exports traverse it exactly like existing children.
Invalid definitions fail construction, not only first use.

For example, a resource-balance field is:

```go
&Schema{Kind: SchemaDictionary, KeyFormat: "mechanical-id",
    Values: &Schema{Kind: SchemaString, Format: "canonical-decimal"}}
```

The actual receipt owner supplies each value descriptor. A Go `map[string]any` used internally
is not itself a schema, and fixed heterogeneous objects still need exact named fields/unions.
Key syntax does not establish catalog membership, pet ownership or permission; those remain
the existing server-owned semantic checks. No current fixture becomes an ID allowlist.

### TD2 — Runtime validation without rewriting

Accept JSON objects only. Empty dictionaries are valid. Validate every decoded own member's
whole key against the chosen grammar and every value against `Values`; invalid keys/values
refuse rather than disappear. Arrays, scalars and null refuse unless null is explicitly an
alternative in the surrounding descriptor. A dictionary does not require any particular key.

Registry HTTP admission retains its original-byte duplicate-member checks, including escaped
duplicates in dictionary members and nested values. Shared stored-artifact validation retains
its existing decoder contract; this amendment does not reverse RP-449. No serializer sorts,
renames, strips, reconstructs or otherwise rewrites a stored receipt to satisfy its descriptor.
Existing numeric formats/bounds, JSON limits and depth checks apply; no new member-count cap
or timestamp/numeric-domain policy is invented.

When receipts are later registered, explicitly describe any historically valid nullable maps
or versioned field sets. Do not normalize null to `{}` or omit older variants to make registration
pass. That later registration must cover the complete served request/receipt population.

### TD3 — OpenAPI and generated TypeScript

Generate a JSON Schema object with exactly one anchored `patternProperties` arm for the
existing key grammar, the generated value schema, and `additionalProperties: false`:

```json
{
  "additionalProperties": false,
  "patternProperties": {
    "^[a-z][a-z0-9_]*(?:\\.[a-z][a-z0-9_]*)*$": {
      "format": "canonical-decimal",
      "type": "string"
    }
  },
  "type": "object"
}
```

The generator derives the pattern from the same format authority as runtime validation;
there is no second regex table. Anchoring and refusal of unmatched properties follow the
[JSON Schema object rules](https://json-schema.org/understanding-json-schema/reference/object#patternProperties).
Only the two named formats are legal descriptor inputs; emitting a regex is not accepting
user-defined patterns.

Generate `Partial<Record<string, V>>` for the value type `V`. This preserves typed values and
honestly models missing-key reads under the repository's current strict TS configuration,
which does not enable `noUncheckedIndexedAccess`. Regex key constraints remain runtime/schema
constraints, not a promise TypeScript can enforce arbitrary strings. No `any`/`unknown` value
fallback, branded-key assertion or guarantee that every possible ID exists.

Existing descriptors/operations that do not use dictionaries must generate byte-identical
OpenAPI, TypeScript and compatibility pins. Existing fixed objects remain closed.

### TD4 — Compatibility and ownership

Pins retain the complete generated dictionary schema, not just `type: object`. Compare the
single key pattern and `additionalProperties: false` explicitly, then compare its value
descriptor recursively under the existing request/response C2 rules. Key-format changes,
dictionary/fixed-object conversions, extra patterns and unrestricted-property replacements
are incompatible. Existing response widening rules and request-enum restrictions apply to
values; bounds must retain exact int64 comparison, including above 2^53.

Report the dictionary field/value path on failure. Nested value schemas must be deeply cloned
on registry construction/export; caller mutations cannot change validation, generation or pins.

Adding this unconsumed DSL arm does not change gameplay replay identity. Later receipt-owner
exports inside kernel-watched packages still require lawful scope/version authority; this
proposal supplies no history exception, false kernel bump or permission to relocate owner
descriptors merely to evade the guard.

## Acceptance criteria

1. Descriptor tests cover both formats, empty/populated maps, typed/ref/nested values, missing
   values, unknown formats, mixed schema fields, dangling/cyclic references, depth and deep-copy
   mutation controls. Invalid definitions fail before registration.
2. Executed runtime tests accept valid unseen IDs and reject partial/uppercase/invalid keys,
   wrong values, extra nested fields, arrays/null/scalars and raw duplicate/escaped members.
   Existing artifact-decoder controls remain green; values are not coerced or dropped.
3. Generation is deterministic, preserves unrelated bytes, and emits one anchored pattern,
   closed additional properties and the actual value schema. Generated TS compiles with positive
   cases and compiler-negative wrong-value/unguarded-missing-key cases, including union values.
4. Executed pin tests reject each TD4 incompatibility and one-unit integer narrowing above
   2^53, while retaining legal response widening and the existing request restrictions.
5. Affected `publicapi` tests/vet, generation/type checks and existing API consumer tests pass.
   Test the changed schema outcome; do not run unrelated gameplay studies for this DSL-only
   batch. Designated review covers the complete implementation and records before archival.

These criteria finish the schema amendment, not API Foundation AC4, the intent migration,
public privacy, whole CI or a release. Actual handler/SQL/browser checks belong to the later
receipt registration and consumer change and cannot be replaced by schema tests.

## Deviations and remaining decision

Explicitly amend C17's unsupported-map restriction only for the typed dictionary arm above.
No deviation from gameplay or C18's no-free-form-artifact law. The one adoption question is
typed dictionaries preserving current receipts versus a separately designed versioned HTTP
projection. Other existing owner/author decisions are not inferred from either choice.

## Changelog

- 2026-10-08: source-grounded RP-391 proposal; owner choice requested, not received or adopted.
