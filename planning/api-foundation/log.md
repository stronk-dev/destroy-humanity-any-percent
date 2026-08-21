# API Foundation implementation log

## 2026-08-03 — accepted-contract reconciliation

C1–C17 are owner-ruled. A1–A8 now name the operation/schema single authority, literal public DTO
families, normalized board query, validate-before-parse HMAC cursor, and operational middleware
semantics. The two historical acceptance-blocker sections remain evidence, while the status and
changelog now truthfully mark implementation unblocked.

Formula fallback remains prohibited: hashes without stored formula artifact bytes return honest
unavailability until the future artifact-growth mint.

## 2026-08-07 — C19/C20 response and operational-policy implementation

- Replaced the registry's JSON-only response map with C19's closed descriptor union. Schema
  responses name an exact JSON descriptor; raw responses admit only `application/json` or
  `application/gzip`, require a declared content-hash header, and validate the exact repository
  bytes against their SHA-256 without decoding.
- Added strict `balance/api/phase0.json` ownership for the ruled public limiter, trusted-proxy,
  cache, cursor-key-ID, and request-ID literals. The loader rejects unknown, missing, duplicate,
  trailing, or out-of-domain fields. Startup resolves both named 32-byte deployment secrets and
  permits the ruled same-secret first-deployment state without putting secrets in JSON.
- Extracted the account API's reviewed token-bucket and trusted-client-IP mechanics into the
  shared `httpapi` package, then routed the account API through that one authority. LRU bounds,
  refill behavior, clock-regression behavior, and proxy selection remain pinned by the original
  account tests plus shared-package fixtures.
- Added public request-ID resolution and exact cache serving. Valid caller IDs echo; invalid or
  overlong IDs become UUIDv7. Strong ETags hash served bytes, verification responses add
  `immutable`, and matching conditional requests return 304 before the public limiter, so repeated
  cache hits spend no token. Ordinary responses remain bounded per IP.
- The repository-root `make verify` gate passed in full: typecheck reported zero diagnostics,
  6,603 client tests passed (3 skipped), and all 19,818 browser assertions passed. C18's seven exact
  owner artifact arms and the public handlers/generator remain visibly open; this entry makes no
  endpoint, review, or archival claim.

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

## 2026-08-04 — schema and cursor authority remediation

- Cursor codecs now require the validated operation registry. Each paginated operation names one
  exact object schema for its key; unknown and non-paginated operations reject, raw key bytes must
  validate against that schema, and the decoded target must re-encode byte-exactly.
- Named schema reference graphs reject direct and indirect cycles before startup. Runtime
  validation retains a defensive depth bound.
- Registry construction deep-clones schemas and operation response maps into private storage.
  Enumeration APIs return defensive snapshots, so caller or generator mutation cannot change the
  runtime authority after validation.
- The `canonical-decimal` format delegates to `decimal.ParseCanonical`; the API no longer carries
  an incompatible second big-number grammar.
- Added the single normalized board-filter encoder/hash authority over category, variables, epoch,
  mandate, and limit. Every dimension has a discriminating hash-binding fixture.
- Focused `./publicapi` tests pass. C18–C20 remain honest endpoint/owner blockers; no handler,
  middleware, raw-body descriptor, or security literal was improvised.

## 2026-08-04 — expanded schema-depth closure

Independent review found that an acyclic named-reference chain could pass construction yet exceed
the runtime validator's defensive depth bound. Registry construction now walks the fully expanded
reference graph with the same limit before cloning it. The reviewer's 66-definition reproducer is
a permanent negative fixture, so startup cannot bless a schema runtime validation must reject.

## 2026-08-04 — independent schema/cursor remediation review (`402ba20..85cbea6` + `8697883^..8697883`)

- **Review by:** Darwin
- **Recorded by:** Darwin
- **Decision:** **approved for this implementation stage.** The five prior findings close, and the
  one new MEDIUM found during this review was fixed and re-reviewed in `8697883` before verdict.

Closure verified:

- Cursor construction is registry-bound. Paginated operations own one exact object key schema;
  unknown and non-paginated operation IDs reject on encode and decode; key bytes validate against
  that schema and typed decode must marshal back byte-identically. Operation response maps and the
  full nested schema graph are privately cloned, and all enumeration APIs return defensive clones.
- Direct and indirect named-reference cycles reject at construction. Inline pointer cycles also
  reject through the construction depth bound, while runtime validation retains a fail-closed
  recursion bound. `canonical-decimal` delegates directly to `decimal.ParseCanonical`, including
  signed values and the numeric core's exponent boundary.
- `EncodeBoardFilter`/`BoardFilterSHA256` are the single exact authority over category, decoded
  variables (including explicit-null faction), epoch, mandate, and limit; discriminating fixtures
  bind every dimension. C18–C20 remain accurately open and no endpoint surface is credited.

Finding found and closed in-range:

1. **MEDIUM (closed by `8697883`) — construction accepted acyclic schemas that runtime validation
   could never accept.** A sorted 66-definition chain `S00 -> S01 -> ... -> S65`, with `S65` a
   string schema, passed `ValidateSchemaDefinitions`; validating the otherwise-valid JSON string
   `"ok"` at `S00` then deterministically returned `ErrInvalidSchema` because `validateValue`
   rejected depth 65. The cycle defense was fail-closed, but construction and runtime disagreed
   about the accepted graph domain. The follow-up introduces one `maximumRuntimeSchemaDepth` used
   by both paths, walks the fully expanded inline/reference graph at construction, and retains the
   exact 66-definition reproducer as a negative fixture. Mixed inline/reference paths traverse the
   same edge increments in both checks, so there is no second depth interpretation.

Independent verification: exact-range `git diff --check`, an uncached focused `./publicapi` test
after `8697883`, the adversarial chain reproducer above, and a fresh repository-root `make verify`.
The full gate exits 0 with 6,517 client assertions and 19,560 browser assertions. C18–C20 and the
later endpoint/generation/composition work remain the only recorded API blockers; approval here is
for the implemented schema/cursor foundation, not those unbuilt surfaces.

## 2026-08-07 — authenticated registry generation and compatibility gate

- Added canonical OpenAPI 3.1 and TypeScript DTO/operation generation from the immutable Go
  registry. The same operation rows now mount the authenticated Soul Recovery and minigame
  handlers, so runtime routing and generated contracts cannot drift into parallel authorities.
- Added exact path-parameter descriptors and the ruled UUIDv7, opaque-ID, mechanical-ID, semver,
  and prefixed-SHA formats. Registry construction rejects any path-template/descriptor mismatch;
  runtime request and response fixtures reject unknown fields and private coordinator state.
- Committed an additive-only v1 compatibility pin. Ordinary generation checks the prior pin before
  writing outputs; only the explicit `make api-pin` target can replace the baseline. Negative tests
  cover operation removal/change, response-field removal, request-union growth, response-status
  removal, and constraint narrowing; optional response growth remains permitted.
- `make api-schema` is the RFC-named generator target and `make api-check` is part of
  `verify-server`. This slice generates the authenticated registry only. Public readers, the thin
  generated-client transport, full public privacy enumeration, and the combined MA real-socket
  lifecycle remain open and are not credited here.
- This entry records implementation/self-review evidence only. It does not satisfy the designated
  cross-party gate and authorizes neither archival nor publication.

## 2026-08-21 — exact schema-response byte lane predeclared

- Claude's designated Q-002 review at `ba8ca65` found that the implementation correctly placed
  exact error bytes in the API registry but violated Q-002's backend-test-only boundary. Q-002 has
  been returned to a net test-only tree by `b9ebab7`; history was not rewritten because the verdict
  cites the rejected hashes.
- This separate API Foundation lane is authorized by accepted A4/A5: operation rows own typed
  error/status alternatives and runtime fixture validation. It may add a validation-only literal
  byte narrowing to schema response descriptors and populate it for the four existing Minigame
  operations.
- It must not change emitted handler bytes, status/category/detail mappings, JSON schema shapes,
  OpenAPI, TypeScript output, compatibility pins, Recovery/Game UI response authority, public
  endpoints, mechanics, surface components, or player copy.
- Predeclared negatives: a schema-valid illegal category/detail cross-product; one appended byte;
  empty/unsorted/duplicate/schema-invalid registered literals; and mutation of both caller-owned and
  enumerated snapshot bytes after registry construction. Every mutation must be restored.
- Required gates: cold focused and root Go tests, vet, generated API drift check, strict client
  typecheck, and sequential Account and Gameserver Postgres populations. The range receives a
  Codex first-filter and mandatory Claude exact-range designated review; it authorizes no archival,
  publication, push, or Q-003 start.
