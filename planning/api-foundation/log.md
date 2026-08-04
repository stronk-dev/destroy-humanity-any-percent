# API Foundation implementation log

## 2026-08-03 — accepted-contract reconciliation

C1–C17 are owner-ruled. A1–A8 now name the operation/schema single authority, literal public DTO
families, normalized board query, validate-before-parse HMAC cursor, and operational middleware
semantics. The two historical acceptance-blocker sections remain evidence, while the status and
changelog now truthfully mark implementation unblocked.

Formula fallback remains prohibited: hashes without stored formula artifact bytes return honest
unavailability until the future artifact-growth mint.

## 2026-08-03 — operation schema and cursor foundation

- Added an explicit Go schema DSL over exact object, array, string, bounded integer, boolean, null,
  ref, and oneOf. Named definitions validate at registry construction and the same values validate
  exact runtime response bytes.
- Added the sorted operation registry with method/path/surface/auth/public/schema validation,
  duplicate-route rejection, and status-specific response validation.
- Added validate-before-parse HMAC-SHA256 keyset cursors. Tokens bind operation + normalized filter,
  enforce canonical JSON, accept exactly current/previous 32+ byte deployment keys, reject padding,
  tampering, and cross-query reuse, and parse the key only after MAC verification.
- Added canonical board-variable encoding with explicit-null faction and integer booleans.
- Focused package and mandatory TypeScript gates pass. The `node:fs` corpus import was removed
  after independent review proved the prior typecheck claim false.
- C18/C19 retain exactness instead of weakening the DSL: heterogeneous artifact JSON needs
  per-owner discriminated schemas, and immutable gzip/raw evidence needs a raw response descriptor.

## 2026-08-04 — independent foundation review (`87f542d..24203ee`)

- **Review by:** Darwin
- **Recorded by:** Darwin
- **Decision:** **not approved; the schema/cursor foundation needs remediation before endpoint or
  generator work.** The HMAC primitive and canonical board-variable codec hold, but the operation,
  key-schema, reference-graph, and canonical-number boundaries are weaker than A5/A7 claim.

Findings, ordered:

1. **HIGH — cursors are not bound to the operation registry or an exact operation key schema.**
   `CursorCodec` receives no registry/key descriptor and accepts any operation matching the ID
   regex; `Encode("unregistered_operation", ...)` is valid despite A7 requiring unknown operations
   to reject. `key` is `any`, and Decode only uses `DisallowUnknownFields`: a signed canonical key
   missing a required struct field decodes to its zero value. A7 specifically requires exact decode
   and re-encode. The test's `get_board` operation is itself absent from the test registry, so it
   normalizes rather than catches the first defect. Register the closed key schema per paginated
   operation, reject unknown/non-paginated operations at encode/decode, and re-encode-compare the
   typed key after decode with missing/extra/wrong-arm fixtures.
2. **HIGH — named reference cycles pass startup validation and recurse without a runtime bound.**
   `validateSchema` checks that a `ref` target exists but never traverses the referenced definition;
   mutually recursive/self-referential named schemas therefore pass `NewRegistry`. `validateValue`
   follows refs with no visited/depth guard, so validating any value against such a schema recurses
   until stack exhaustion. Reject reference cycles (the public DTOs are finite) and add direct and
   indirect cycle fixtures.
3. **HIGH — the validated registry is mutable after construction.** `Schemas` and `Operations` are
   exported; schema pointers/field slices are retained from caller input without a deep clone, and
   the exported schema map is the same graph `ValidateResponse` reads. A caller can validate a
   closed registry, then mutate a field/ref/constraint or replace a map entry so runtime validation,
   future generation, and the startup verdict disagree. Deep-clone into private storage and expose
   defensive read APIs/snapshots; add caller-input and returned-value mutation tests.
4. **HIGH — `canonical-decimal` is not the RFC-0001 canonical grammar.** Its local regex accepts
   values such as `10e0`, `1.0e0`, and unbounded/noncanonical exponents, while rejecting valid
   negative canonical values. Big-number API fields would therefore admit bytes the numeric core
   rejects and reject bytes it accepts. Route the format through `decimal.ParseCanonical` (plus the
   field's signed/nonnegative semantic bound where needed) and promote numeric golden boundaries
   into schema tests instead of maintaining a second regex grammar.
5. **MEDIUM — the complete normalized board-filter hash has no authority yet.** The codec compares
   an arbitrary caller-supplied 64-hex string, and only the standalone variables object is
   canonicalized. No function/type composes and hashes C13's category, decoded variables, epoch,
   mandate, and limit, so two handlers can bind different filters while both using the codec. Land
   one exact normalized-filter encoder/hash with field-order and cross-query fixtures before board
   endpoint registration, or record it explicitly as pending rather than saying the foundation
   already binds the complete normalized query.

What held:

- Token size and unpadded-base64 bounds, 32-byte current/previous distinct keys, current-key
  signing, previous-key verification, HMAC-SHA256, constant-time signature comparison per attempted
  key, and MAC-before-JSON-parse order are correct. Tampering, cross-operation arguments, and
  cross-filter arguments reject after authentication.
- Board variables use the ruled key order, integer booleans, explicit-null faction, exact strict
  decoding, re-encode equality, and unpadded base64url. The operation registry correctly enforces
  sorted IDs, duplicate route rejection, surface/path/auth/public consistency, and known named JSON
  response descriptors for the currently representable arm.
- A fresh root `make verify` and focused `make test-go
  GO_PACKAGES='./meters ./achievements ./publicapi'` both exit 0; both exact-range diff checks pass.
  The green suite lacks fixtures for the findings above rather than contradicting them.
- C18 (artifact-specific JSON union), C19 (raw response descriptor), and C20 (literal operational
  policy/secrets lookup) are honestly unresolved in the RFC/plan. Public handlers, DTOs, raw
  evidence, OpenAPI/TS generation, compatibility pins, middleware, router composition, and privacy
  conformance remain unchecked—not silently credited. Historical formula fallback remains absent.

## 2026-08-03 — operational-policy gap retained

Source review confirmed C6/C16 enumerate cache ages but not the limiter capacity/refill, maximum IP
entries, trusted-proxy hops, cursor key IDs, or exact accepted request-ID grammar/bound. C20 carries
those owner/security literals. No production abuse-control value was improvised in middleware.
