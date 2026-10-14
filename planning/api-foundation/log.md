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

## 2026-08-21 — exact schema-response byte implementation and Codex first-filter

- Implementation commit: `0331444`, after predeclaration `a9fdb23`. `Response.ExactJSON` is a
  validation-only, schema-valid, sorted literal-byte narrowing. Registry construction rejects bad
  literal sets; internal rows and returned operation snapshots own deep-cloned bytes; raw responses
  cannot carry the schema-only arm.
- The four Minigame operations now attach their existing error bytes to the owning status and
  action. No handler mapping or emitted byte changed. Recovery and Game UI continue using ordinary
  schema validation; generated OpenAPI, TypeScript, and compatibility-pin bytes remain unchanged.
- API Foundation tests cover empty, duplicate, unsorted, schema-invalid, and raw-response literal
  declarations plus caller/snapshot mutation. The Q-002 exact handler table additionally proves all
  13 shipped deterministic outputs are accepted by their operation/status row while a schema-valid
  illegal cross-product and an appended byte reject.

Demonstrated failing mutations, all restored before `0331444`:

- bypassed `ExactJSON` matching after schema validation: Public API and Account tests accepted the
  wrong pair, and both populations failed at their literal oracle;
- replaced response deep-cloning with a shallow slice copy: mutating caller-owned bytes corrupted
  the registry and the immutability test failed;
- bypassed literal-set construction validation: empty, duplicate, schema-invalid, and unsorted
  registrations all became accepted and failed their four named subtests.

Cold restored-tree evidence:

- `make test-go GO_PACKAGES='./publicapi ./account ./gameserver' GO_TEST_FLAGS=-count=1` — green.
- `make test-go-core CORE_TEST_COUNT=1` — every non-harness Go package green cold.
- `make vet`, `make api-check`, and `make typecheck` — green; generated API files remained
  byte-identical and Svelte reported zero diagnostics.
- Sequential real-Postgres populations passed cold: Account in 1.353 s and Gameserver in 25.941 s
  through the root `make test-save-integration` target with `SAVE_TEST_COUNT=1`.
- `git diff --check` is clean and only the user-owned `AGENTS.md` remains dirty outside the lane.

Codex first-filter verdict: **APPROVED** for the predeclared API Foundation corrective scope.

- **Review by:** Codex.
- **Recorded by:** Codex.
- **Reviewed range:** `a9fdb23..0331444` (predeclaration plus implementation/tests/docs).

The production tightening is now outside Q-002's test-only boundary and has its own authority,
negative controls, gates, and provenance. Claude must designated-review the API Foundation range
including this record commit. This first filter authorizes no public endpoints, surface claim,
archival, publication, push, or Q-003 start.

## 2026-08-21 — Claude designated cross-party review of the exact-response lane `b9ebab7..a854e46` — APPROVED

- **Review by:** Claude (designated cross-party). **Recorded by:** Claude.
- **Range:** `b9ebab7..a854e46` = `{a9fdb23, 0331444, a854e46}`.
- **Authority verified:** the predeclaration cites accepted A4/A5, and `rfc/api-foundation.md`
  carries both (operation rows owning typed error/status alternatives; one operation/schema
  authority). `registry.go` is this lane's home. The scope-split remedy from `ba8ca65` is
  satisfied exactly: the re-landed registry diff is **byte-identical** to the content I already
  reviewed in the rejected mixed range, now under its own predeclaration, first filter, and
  canonical-doc update.
- **Process correction, mine:** my `ba8ca65` verdict offered a history-rewrite split as one lawful
  option. It was not — that verdict itself cited the affected hashes, which is precisely what
  closes the rewrite carve-out ("once a planning-log verdict cites a hash, that history is
  append-only"). Codex declined the unlawful option for the right stated reason and
  forward-reverted instead. The erroneous offer is mine and is corrected here.
- **Executed:** `./account ./publicapi` green cold `-count=1`; `make api-check` 0 (generated
  contract byte-unchanged, as the predeclaration requires); Account + Gameserver Postgres
  populations green on an isolated port. **Severing probe (mine):** gutting the `ExactJSON`
  narrowing inside `ValidateResponse` in a scratch worktree fails `./publicapi` immediately — the
  lane's own negative controls discriminate against the exact mutation that would make it
  decorative.
- **No findings. APPROVED.**

## 2026-09-24 — Drafter body reconciliation (Claude): A6 and AC4

Queue item 3f listed API Foundation body text that contradicted owner rulings. As drafter of the
RFC, Claude reconciled only **body** sections, per AGENTS.md "Rulings reconcile the body":

- **A6:** "catalog artifacts embed sorted named JSON" contradicted C18, which rules no free-form
  JSON arm. It now states C18's closed `oneOf` discriminated by artifact `name`, with
  owner-exported exact descriptors and one arm per served artifact name.
- **AC4:** "hand-written API layer is replaced … diff shows deletion" contradicted C9. It now
  carries C9's ruled criterion: the generated client is the only HTTP-calling code, and a lint
  forbids raw `/api/` fetch outside `client/src/api/generated/`.

**Not changed:** the C18 ruling's sequencing note ("until then the union carries only the 7 base
arms"). It is owner-ruled text; under evidence-discipline rule 5 its author edits it. It is not
false as a conditional, but it is historical: epochs 6–8 are minted and epoch 8 pins 19 artifact
families. The changelog records that 19 arms are required before the catalogs reader is composed.
This is **Claude-authored RFC text requiring Codex cross-party review**. It changes no mechanic, and
no implementation starts from it until reviewed.

## 2026-09-25 — Public epochs reader batch predeclared (Claude)

**Implemented by:** Claude, as implementer on the owner's 2026-09-24 direction. Awaiting Codex
designated review; not self-approved.

Authority is the accepted A3/A5/A6/A7/A8 text plus rulings C2, C3, C6, C12 (literal `EpochPage`),
C15 and C16. Only ruled text is used; the pending 2026-09-24 A6/AC4 body reconciliation is not
relied on.

Scope of this batch:
1. **Registry query parameters (A5).** Paged public operations need `limit` and `cursor`, but the
   registry currently models only path parameters. The work adds exact scalar query-parameter
   descriptors to operation rows, OpenAPI `in: query` generation, TS `query` fields, and
   compatibility pins. Any change to an existing operation's query set is rejected, which is
   stricter than C2 and never looser.
2. **`GET /api/public/v1/epochs`.** C12 fixes the page shape as `EpochPage {items:[{epoch_id,name,
   started_at,ended_at,changelog_ref,changelog_markdown,accepted_hashes}],next_cursor}`:
   - newest-first;
   - UTC RFC3339 milliseconds, with `ended_at` explicitly null while open;
   - byte-sorted hashes;
   - `limit` defaults to 50, bounded 1..100;
   - C15 keyset cursor bound to the normalized filter;
   - C16 cache class `catalogs_epochs`.

   It is served through a public registry that is generated alongside the private registry.

Predeclared evidence:
- failing-first unit and Postgres integration tests;
- a severing probe for each gate (query validation, the compatibility rule, ordering, cursor
  binding, the null rule, changelog binding);
- cold `-count=1` runs;
- `make api-check`.

**Not in this batch (C18 descriptor gap):** 19 artifact families need owner-exported exact
descriptors. The catalogs reader stays closed until every owner exports one, and no free-form arm
is added.

## 2026-09-25 — Public epochs reader implemented (Claude)

**Implemented by:** Claude. **Review:** awaiting Codex designated cross-party review. This is not
self-approved; no box is flipped and nothing is archived.

**Range:** the predeclaration commit `cd5ba06d`, the query-parameter commit `32560c4b`, and the
reader commit that follows this entry.

- **Query descriptors (`32560c4b`):**
  - adds exact scalar `QueryParameter` rows: byte-sorted, no path-name shadowing, and a `cursor`
    that is present exactly when there is a cursor key and is never required;
  - `Registry.ParseQuery` accepts strict base-10 integers only, validated through the shared
    descriptor validator;
  - generated OpenAPI gets `in: query` entries and TypeScript gets `queryParameters`/`query`;
  - pins carry `query` only for operations that have one, so every prior generated byte is
    unchanged (`make api-check` was clean at that commit);
  - any change to an existing operation's query set is rejected, which is stricter than C2.
  - **Severings (each run red):**
    - Q1 sort check off;
    - Q2 cursor↔key binding off;
    - Q3 canonical integer off (a first attempt survived because the JSON re-decode also rejects
      `010`/`+5`; a `-0`/`00` case now witnesses the pattern on its own);
    - Q4 query dropped from the compatibility check.

    The first Q1/Q2 runs were compile failures and are not counted; they were redone with
    compiling mutations.
- **Reader:**
  - `leaderboard.PublicEpochPage` (Postgres integration test);
  - `publicread` schemas, operation and handler (unit tests with a fake reader);
  - `publicapi.MergeRegistries`, plus gen-api generating both surfaces.
  - **Severings (each run red):**
    - P2 ASC order;
    - P3 `<=` keyset bound;
    - P4 changelog fail-closed removed;
    - P1′ Go hash sort removed;
    - H1 filter ignores limit;
    - H2 `ended_at` dropped;
    - H3 cursor taken from the first row;
    - H4 limit→cursor error mapping;
    - M1 merge conflict check off.
  - **Honest non-discrimination:** P1, removing the SQL `ORDER BY` inside `string_agg`, SURVIVED
    the integration test, because Postgres's primary-key scan already yields sorted hashes. The
    byte order is therefore enforced by an explicit Go sort, which the unit test witnesses. The SQL
    ordering is not credited as a check.
- **Compatibility pin refresh:** `make api-pin` is authorized by C2 (a new operation is additive)
  and C12 (the ruled public DTO). The pin gains `list_public_epochs` and its schemas. The
  `APIError` detail enum widens by `cursor`, `limit` and `public_api`, which is a response-enum
  widening C2 permits. The ordinary compatibility check accepted it before the refresh.
- **Cold evidence:**
  - `make test-go GO_PACKAGES='./publicapi ./publicread ./leaderboard ./account ./gameserver
    ./cmd/gen-api' GO_TEST_FLAGS='-count=1'` passed;
  - the Docker Postgres `-run Integration` run for leaderboard, account, gameserver, publicread
    and publicapi passed;
  - gofmt and go vet are clean;
  - client `tsc --noEmit` is clean against the regenerated types.

**DESIGN-GAPs (for the RFC author; this batch implements the most literal reading and flags it):**
1. **`invalid/limit` detail.** C12 bounds the limit (1..100) but names no rejection detail. The
   handler uses `invalid/limit`, following the ruled per-parameter pattern (`invalid/cursor`,
   `invalid/variables`). This needs ratification.
2. **`internal_invariant/public_api` detail.** No ruled detail exists for a public-read server
   failure. It mirrors `internal_invariant/minigame_api` and needs ratification.
3. **304 has no C19 response arm.** The C19 union is schema|raw, so the bodiless 304 that C16
   requires cannot be declared, and OpenAPI omits it. A third, bodiless arm (or an explicit ruling
   that 304 is transport-level and undeclared) is needed before the conformance test can enumerate
   every status.
4. **Undeclared query parameters are ignored, not rejected.** No ruled error exists for them, and
   the ETag is computed over the served bytes, so a cache split cannot mislabel content.
5. **The C18 catalog union** still needs an exact exported descriptor from each of the 19 pinned
   artifact owners. No free-form arm was added, and the catalogs reader remains closed.

**Not done (still open plan items):**
- composing the public router into the gameserver: policy load, cursor secrets from deployment
  config via `ResolveCursorCodec`, and mounting;
- the boards, verification and registry readers;
- the thin generated-client transport;
- the full public privacy enumeration (AC5).

## 2026-09-25 — Predeclaration: compose the public router (Claude)

**Implemented by:** Claude; awaiting Codex designated review.

The batch is authorized by accepted A5, A7, A8, C10 and C20, together with Deployment Foundation's
table row "CLOUD_CLICKER_CURSOR_* … required when public cursor readers are composed". Scope:

1. Gameserver composition loads `balance/api/phase0.json` through `publicapi.LoadPolicy`. It builds
   the public registry, `ResolveCursorCodec` and a `publicapi.Runtime`, and mounts every public
   operation from the registry under `/api/public/v1/`, beside the account routes. Startup
   **fails** if the policy or cursor secrets are missing or invalid. There is no
   restart-generated fallback.
2. **Cursor-secret resolution (literal reading, flagged for review):** the C20 key ID `k1` resolves
   to the deployment cursor pair entry whose ID is `k1`, and `k0` to the entry whose ID is `k0`.
   When the deployment supplies no previous pair, `k0` resolves to the current secret. That is the
   C20 clause "at first deployment both names MAY resolve to the same secret value", and the
   deployment decoder forbids a previous value equal to the current one. Any other ID set fails
   startup.
3. The production profile makes the cursor pair **required** (Deployment Foundation row). The
   compose template, config schema, `.env.example` and release-template validation provision it
   the way they provision JWT. The development profile gains a required inline `CLOUD_CLICKER_CURSOR_KEY`.
4. **Tests:**
   - composition fails without keys;
   - a real composed `GET /api/public/v1/epochs` returns the ruled page with request-ID,
     ETag/Cache-Control and limiter behaviour;
   - a cursor minted under `k1` still decodes after rotating `k1→k0`;
   - the composed Game UI lane sets the dev key.

   Each check gets a severing probe.

Out of this batch: boards, routes, verification readers, AC5 enumeration (next batches), and C18
(still a gap).

## 2026-09-25 — Public router composed (Claude)

**Implemented by:** Claude. **Review:** awaiting Codex designated cross-party review. Not
self-approved; nothing flipped or archived. Predeclared in `5825ad24`; the implementation is the
commit that follows this entry.

**Delivered:**
- `publicread.NewRouter` composes the public surface. It uses the strict policy,
  `ResolveCursorCodec`, request IDs, the runtime and limiter, and `Registry.Mount`. Unknown paths
  and methods return `404 unknown_id/route` (the account router's existing pair).
- `CursorSecretResolver` binds the C20 names `k1`/`k0` to the deployment pair by ID. With no
  previous pair, `k0` resolves to the current secret (the C20 first-deployment clause).
- The gameserver mounts the router at `/api/public/v1/` beside the account routes. Composition
  **fails closed** without a valid policy or cursor pair.
- The production cursor pair is required (Deployment Foundation row). The compose template pins
  `CLOUD_CLICKER_CURSOR_CURRENT_ID=k1` and mounts the `cursor-current` secret. The config schema,
  `.env.example`, template validation and the manifest/schema required lists all carry the new
  secret-file input.
- The development profile requires `CLOUD_CLICKER_CURSOR_KEY`, bound to `k1`, which is rejected in
  production.
- The runtime content closure stages `balance/api/phase0.json`.

**Evidence (cold):**
- `make test-go GO_PACKAGES='./publicapi ./publicread ./leaderboard ./gameserver ./deploymentconfig
  ./cmd/gameserver ./releasepackage' GO_TEST_FLAGS='-count=1'`, and the Docker Postgres
  `./gameserver -run Integration` run, both pass.
  - The new composed witness covers the fail-closed no-keys composition, the epochs page,
    request-ID echo, cache headers, a 304, and the unknown public route.
- The deployment lanes are all green:
  - `make test-deployment-release`, including the real Caddy integration;
  - `make test-deployment-rehearsal` (both retained build records still validate);
  - `make test-deployment-operations` (promtool SUCCESS and the private profile integration).
- `make test-game-ui-composed` passes, and now fetches `/api/public/v1/epochs` through the Vite
  proxy from the built binary.
- `make api-check` shows no diff, and client `tsc` and go vet/gofmt are clean.
- `make deployment-config-check` validates the ambient environment. With an explicit dev
  environment it passes when `CLOUD_CLICKER_CURSOR_KEY` is set and fails without it.

**Severings (each run red; code restored):**
- S1: remove the gameserver mount. The composed witness fails with 404.
- S2: drop the first-deployment `k0` rule. Three `publicread` tests fail.
- S3: production cursor pair not required. Four config tests fail.
- S4: silently fall back to account-only routes when the public router errors. The composed witness
  fails ("composition without cursor keys: <nil>").
- S5: omit the API policy from the closure. `TestRepositoryRuntimeClosureIsManifestDrivenAndExact`
  fails. My first S5 run used a `-run` filter that selected no closure test and printed `ok`; it is
  not counted and was redone on the whole package.
- S6: remove the dev cursor key from the composed lane. The gameserver exits 1.

**Interference, logged:** the first composed-lane run failed to build because the concurrent
Reputation lane was mid-edit in `server/production` (`undefined: reputation`). It passed once that
package compiled again. None of my files are involved.

**DESIGN-GAPs (for the ruling authors):**
6. **Cursor rotation vs the fixed C20 names.** `validPolicy` pins `k1`/`k0`, and C20 says rotation
   changes only deployment config. Deployment Foundation's rotation ledger (`rotation-activate` /
   `rotation-remove`) records a new ID per rotation and forbids a previous ID equal to the current
   one. A fixed-name scheme instead swaps the secret files behind `k1`/`k0`, which the ID-based
   ledger cannot express. Cursor rotation therefore stays inactive (the compose rotation overlay has
   no cursor entry). The API and Deployment authors need to reconcile: either the policy stops
   pinning names, or the ledger learns fixed-name secret swaps.
7. **Retained release bundles:** candidates built before this commit lack the cursor secret mount.
   `ValidateComposeTemplate` now rejects them, adding to the existing R5 invalidation. R-006 needs
   fresh bundles anyway.

**Next in this lane:** the boards reader (C12/C13), the routes reader (C12 RoutePage), and the AC5
privacy enumeration.

## 2026-09-25 — Public boards reader (Claude)

**Implemented by:** Claude. **Review:** awaiting Codex designated cross-party review; not
self-approved. Authority: accepted A6, C12 (BoardPage literal), C13 (normalized query, ranking
kind from the pinned catalog) and C15 (cursor MAC over the full filter). This is a new operation,
so the widening is additive under C2. The pin was refreshed (`make api-pin`) after `gen-api`
accepted the change against the prior pin.

**Delivered:**
- `list_public_board` at `GET /api/public/v1/boards/{category}`.
- `leaderboard.PublicBoardRankingKind`, with the pure `resolveRankingKind`.
- `leaderboard.PublicBoardPage`, which fetches `limit+1` through the existing board SQL, now split
  into a bounded exported wrapper and an internal query.
- `rowQuerier` lets the projector's pinned-catalog loader be shared read-only.
- The shared `APIError` detail enum is widened by `category`, `epoch`, `mandate` and `variables`
  (a C2 response-enum widening).
- No kernel-guarded path is touched: `leaderboard/categories.go` is unchanged.

**Evidence (cold):**
- Unit and integration runs:
  - `make test-go GO_PACKAGES='./publicapi ./publicread ./leaderboard ./gameserver ./account
    ./cmd/gen-api' GO_TEST_FLAGS='-count=1'` passes.
  - Docker Postgres: `./leaderboard` passes, covering the ranking kinds from the real seeded epoch,
    unknown category/epoch, paging with ties, variables partitioning, magnitude order, no cursor on
    an exactly-full page, and a stored unloadable catalog failing loudly. `./account` passes.
  - `./gameserver` passes; the composed witness now also serves an empty valuation board with exact
    bytes and a `404 unknown_id/category`.
- **One full `./gameserver` Postgres run failed** in
  `TestComposedAccountFamilyRevocationRevalidatesSocketsIntegration` (24 s). The test passed alone
  and the whole package passed on rerun. I can't attribute it: no other test container was running
  at the rerun. It is logged as an unexplained single failure, not as a pass.
- `make api-generate` then `make api-pin`: the diff contains the new operation, its schemas and
  the widened detail enum. Client `tsc` is clean.

**Severings (each run red; restored):**
- B1: fetch `limit` instead of `limit+1`. The Postgres witness fails (`more=false`).
- B2: cursor filter drops `limit`. The limit-mutated cursor is accepted, and the unit test fails.
- B3a/B3b: the cursor-arm checks. My first B3 mutation SURVIVED because `false && …` still left the
  other checks active. I added a both-arms cursor case and severed exactly the `Key != nil` and
  magnitude-field checks, and both then failed.
- B4: advisor flag dropped before the reader. The unit test fails.
- B5: timer disagreement ignored. The resolver unit test fails.

**Honest non-discrimination, corrected:** my first integration "disagreement" case was VACUOUS. The
canonical-shape category loader rejects any non-phase-0 timer, so the rewritten catalog failed to
load, and `ErrInvalidEpoch` came from the loader rather than from the disagreement branch. The
integration step now asserts what it actually shows: an unloadable stored catalog fails loudly for
every category. Disagreement is tested on the pure `resolveRankingKind`. With the current loader,
disagreement between loadable catalogs is unreachable, so the branch is defensive.

**DESIGN-GAPs:**
8. **Unruled details.** `invalid/{epoch,mandate,variables}` follow the ruled per-parameter pattern.
   `unknown_id/{category,epoch}` follow C12's `unknown_id/constants_hash` pattern. A malformed path
   category is treated as unknown (404), not as 400. All of this needs ratification.
9. **Required query parameters.** `epoch`, `mandate` and `variables` are required because C13
   lists them as the normalized filter and names no defaults. A "current epoch" default would need
   a ruling.
10. **Mandate rows.** The projector only writes `mandate_level` 0 today, so `mandate>0` boards are
    always empty until a mandate producer exists.

## 2026-09-25 — Route Registry reader and AC5 privacy enumeration (Claude)

**Implemented by:** Claude. **Review:** awaiting Codex designated cross-party review; not
self-approved.

**Authority:**
- A3's `GET registry/routes`.
- C12's RoutePage: "route ID, public name, first-executor founder ID (nullable after
  anonymization), credited-at, naming status/deadline, and adoption count … Define its exact DB
  source and ordering".
- C10 and AC5 for privacy.
- `list_public_routes` is a new operation, so the widening is additive under C2. The pin was
  refreshed after `gen-api` accepted the change against the prior pin.

**Decisions within the C12 delegation, recorded for the reviewer:**
- **Source:** `registry_routes` LEFT JOIN `account_founders`.
- **Order:** by the immutable `route_id`, because `occurred_at` is rewritten when an earlier
  execution re-credits a route.
- **`public_name`:** the player name only when `published`, otherwise the house name. Pending names
  are unmoderated, and exposing them would publish unreviewed text.
- **Executor:** withheld when the ownership row's `account_id` is NULL (anonymized) and also when
  no ownership row exists (privacy-conservative).
- **Field names:** snake_case of the C12 phrases (`first_executor_founder_id`, `credited_at`,
  `naming_status`, `naming_deadline`, `adoption_count`, `public_name`, `route_id`).

**Evidence (cold):**
- `make test-go GO_PACKAGES='./publicapi ./publicread ./leaderboard ./routeprojection ./gameserver
  ./account ./deploymentconfig ./releasepackage ./cmd/gameserver' GO_TEST_FLAGS='-count=1'` passes
  apart from `releasepackage` (see interference).
- Docker Postgres `./routeprojection ./leaderboard ./gameserver ./account -run Integration` passes.
  The new route witness covers route_id order, keyset paging, a pending name hidden behind the
  house name, the anonymized and orphaned executors withheld, a published name with its live
  executor shown, and no cursor on an exactly-full page. The AC5 enumeration runs inside the
  composed-server test.
- `make test-game-ui-composed` passes.
- `make api-generate` / `make api-pin` add only the routes schemas and operation.

**Severings (each run red; restored):**
- R1: leak the pending name. The Postgres witness fails.
- R2: leak the anonymized executor. The Postgres witness fails.
- R3: the route cursor ignores `limit`. The unit test fails.
- P1: add `PublicRoute.account_id`. The structural privacy test fails.
- P2: an unlisted founder field. The structural privacy test fails.
- A1: drop one operation's enumeration request. The composed test fails ("has no privacy
  enumeration request").
- A2: a canary secret known to appear (`ranking_kind`). The composed scan fails, which shows the
  value scan can detect presence.

**Honest limit of the value scan:** the seeded founder has no verified run and no credited route,
so the scan proves that none of the account's private values leak through any public operation.
It cannot exercise the allowed public-identity positions with that founder. Those positions are
covered structurally (P1/P2) and by the route witness (R2).

**Interference, not mine (logged, not counted as a pass):** while this batch ran, the concurrent
Reputation lane had uncommitted work in progress. `server/save/migrations/00075_reputation_node_purchased.sql`
is untracked, which fails `releasepackage.TestCurrentMigrationIsContiguous` (migration 75 against
the manifest), and `client/src/replay.ts` is mid-edit, so client `tsc` reports undefined names
there. Neither involves my files. The same releasepackage test passed at `731bb3e7`.

**API Foundation status after this lane:**
- Implemented: the epochs, boards and routes readers are composed and mounted, and AC5 is
  enforced.
- Still open:
  - the verification endpoints (C14 raw-bytes manifest);
  - the catalogs reader (C18 descriptors, gap 5);
  - the thin generated-client transport and the AC4 raw-fetch lint;
  - the conformance test's 304 enumeration (gap 3).

## 2026-10-06 — predeclare refresh generated-contract evidence, not renewal code

Accepted A2/A4/A5, C1/C7/C9 authorize inspecting the existing registry/generated
contract. Current source has fourteen generated operations, no refresh path;
APIError lacks refresh_reused/session_family_revoked and unauthorized/refresh_token.
Actual Account parser census already emits the latter. RP-240 records this seam,
not a new policy or chosen future operation ID. Renewal RFC remains draft.

Bounded manual compiler probe outside CI/verify: use repository TypeScript against
actual generated types and virtual callers (no files written by compiler). Positive
registered-bootstrap path and invalid/body error controls must compile; exact
refresh path and both existing refresh error pairs must be rejected with semantic
diagnostics on the virtual caller, never incidental module/import errors. In-memory
counterfactual path/error widening must remove only the corresponding refusals;
counterfactual bytes are synthetic, never written or adopted as a contract.
Run actual api-check drift/compat generation and cold publicapi tests. Census
operation counts/paths and AST-confirm actual fetcher call sites; no claim of
runtime HTTP, successful rotation, browser coordination, generated dispatch or
complete request/status schema. Source hashes and limits retained explicitly.
Instrument goes in *.fixtures.mjs, not Vitest-discovered *.test.mjs (RP-239 lesson).
All handles drain before edits; no production/generated/schema/pin/auth/copy/
kernel/CI membership or checkbox changes. Review by: Codex (predeclaration only).
Recorded by: Codex. New range starts 29e1ff02 exclusive, Claude independently
required. Docker capacity approval, prepared DB populations and all prior spans
remain separate, no shortcut/archival/push or 1.0 promotion.

## 2026-10-06 — compiler executes existing refresh representation gap

Under 6b279ef6 predeclaration, api-check 60224 passes generation/drift/compat,
no artifact/pin diff. Actual TypeScript 5.9.2 probe 19884: three positive callers
compile; refresh path and both error pairs fail exactly TS2322 on virtual callers.
Path-only/error-only synthetic compiler-memory counterfactuals remove only the
corresponding refusals, twelve total compiler arms, 3815 ms. No contract bytes
written/adopted, incidental import/type errors refused. Fourteen OpenAPI/TS
metadata rows match ID/method/path, but actual mounted refresh path is absent.
AST finds three literal runtime fetcher calls plus one dynamic minigame-port
call outside generated; only those two files censused, not a full AC4 lint.

Retained refresh-generated-contract.v1.json and dossier pin actual counts/limits,
source hashes/HEAD and diagnostics. No private tokens/data, HTTP/DB/browser or
rotation evidence. Cold publicapi/vet 75783 passes. Root type/client/topology
52260 passes zero errors/warnings, 7366 tests/134 existing skips, 84/17 files,
thirteen topology negatives. RP-239 fixture-name lesson holds, no extra Vitest
population. Full composite remains prior RP-131 RED, not rerun or relabelled.
All handles terminal before records; no Docker workload/deletion, fresh database
or renewal-policy implementation. No generated/pin/production/auth/kernel/copy/
CI membership/checkbox/archival/push bytes changed.

RP-240 supplies draft S-A1 evidence and accepted C1/C7/C9 remaining work. Next
real DB outcomes/descriptor binding still awaits capacity; do not silently tighten
parser or turn the synthetic global error widening into an exact production
descriptor. Review by: Codex (first-filter only). Recorded by: Codex. New span
29e1ff02 exclusive includes predeclaration/instrument/Make/report/dossier/docs/
ledger/reconciliation and following pin; Claude required independently of prior
Account/CI/browser ranges. Proper full nine-tier 1.0 unchanged, not completed.

## 2026-10-06 — pin generated-contract census separately

Substantive 29e1ff02..79960054, two commits/eleven paths: predeclaration, manual
compiler census, evidence and reconciliation only. This following record edge
also belongs in the complete designated range; closing relay names literal tip.
No generated/compatibility pin/production/auth/schema/copy/kernel/workflow changes.
Twelve compiler arms/four scoped calls/fourteen matching operations are bounded
evidence of incompleteness, not accepted dispatch/rotation/policy. All handles
terminal, clean checkpoint sought, no Docker deletion or push. Review by: Codex
(first-filter only). Recorded by: Codex. Claude independently pending, prior
Account observer 3121a376..29e1ff02 and all earlier ranges remain separate.
- Gaps 6 to 10 are in the entries above.

## 2026-10-07 — preserve interrupted public-board reads; composed gate stays red

Accepted A3/A6/C13. RP-367: PublicBoardRankingKind checked QueryContext and Close, not Rows.Err.
Driver-level regression (78849) reproduces two errors falsely returned as unknown category:
Rows.Next failure and injected context.Canceled. A completed-empty control passes. These are
injected database/sql row errors, not an actual network outage or real cancellation study.
Four-line correction preserves the original error before loading/resolving any catalogs, so
the existing BoardHandler error branch returns server failure instead of category-not-found.
No schema/metadata/pin, balance, kernel or auth policy change.

Cold leaderboard/publicread/publicapi (9787) and selected vet PASS. Added the existing
TestPublicBoardRankingAndPagesIntegration to the composed preflight with mandatory PASS marker.
Actual Postgres case passes (0.16 s): pinned ranking kinds, unknown controls, filtered/tied pages,
magnitude ordering, exactly-full page and unloadable-catalog refusal. Other database cases and
eight refresh populations pass too. The whole make test-game-ui-composed (37452) then FAILS at
Pitch command HTTP429 rate_limited/account. No cosmetic population executes in this run and no
whole-green claim is made; RP-368 retains the observed separate failure. Prior source's full
passes are not substituted. No retries, delays, limit increases or assertions were changed.

Node syntax and diff checks PASS; all handles terminal before records. Review by: Codex
(first-filter only). Recorded by: Codex. New repair/test/binding/docs/records range starts after
`a4bae5af`; designated Claude review remains, independently of older ranges. No archive or push.
Next: measure the actual authenticated request pattern behind RP-368; keep its root gate red
until explained and corrected under accepted authority. Catalog owner descriptors, raw-evidence
privacy/reader work, draft browser policy and full CI/nine-tier 1.0 remain separate obligations.

## 2026-10-07 — RP-368 request diagnostics, not a repair

Test-only browser/driver HTTP metadata in the existing composed journey; no tokens, IDs,
queries or bodies printed, no new artifact pipeline. Run8700 passes the entire root target
(eight refresh cases, Fiscal/epoch/board Postgres cases, main DOM/Pitch and cosmetic journey).
Main account issues98 requests over14.476s, peak14 in one second; GS2 has10 intent submissions
and12 browser state reads, GS5 has10 submissions and14 reads. One interrupted navigation read
remains pending: issued counts are not proof of server arrival. GS5 takes9 manual clicks here,
versus16 in failed37452; do not replace the earlier red with this green or close RP-368.
No behavior, rate, retry or deadline change. Syntax/diff checks pass; fixture connections and
owned listener ports are clear after completion. Review by: Codex (first-filter); Recorded by:
Codex. Range starts after21ff5578; designated review pending. Next: isolate redundant
receipt-triggered reads and the shared request budget without weakening either contract.

## 2026-10-07 — mounted public-cache and limiter coverage

Accepted A8/C16/C20; test-only batch after `5f856f97`. `publicread/cache_test.go` enumerates
the mounted registry (missing fixture fails), uses the actual operational policy, router,
handlers and middleware, and controlled reader data/time. All three operations verify SHA256
ETags over served bytes, exact cache ages/request-ID echo, bodiless 304, zero token charge
before/after exhaustion, exactly 60 uncached successes, typed/non-cacheable 429, changed-data
invalidation, separate client budgets and a one-token refill. A fourth case spends one shared
IP budget across all three paths. Existing production bytes/registry/contracts/CI unchanged.

Focused test (73070) and cold httpapi/publicapi/publicread/leaderboard packages (5313) PASS;
selected vet PASS. Six leaderboard Postgres tests explicitly SKIP without TEST_DATABASE_URL;
no DB, deployed proxy, generated 304-arm, hosted CI or whole-AC3 claim. Existing push CI runs
these tests through verify-server-core's package population. Negative cases are ordinary
cache/budget/content cases, not a new mutation or measurement framework. All runs terminal.

Review by: Codex (diff/first-filter only); Recorded by: Codex. New test/plan/log range requires
designated review with related API work; no archival or push. Next: implement C14's accepted
raw verification readers; catalog owner descriptors, generated dispatcher and 304 metadata
resolution remain separate open obligations. Full nine-tier 1.0 remains incomplete.

## 2026-10-07 — C14 immutable public-evidence repository source

After `834369d7`, implement the accepted C14 producer in leaderboard/public_evidence.go,
with its Postgres test and a mandatory execution marker in the existing composed preflight.
One query authorizes via verified_runs (not queue status), returns pinned metadata and exact
stored genesis/archive bytes, checks genesis/constants and archive SHA256 consistency, and
distinguishes unknown/private runs from corrupt public evidence and operational DB errors.
No wire schema, endpoint, migration, simulation/kernel, policy, copy or workflow changed.

Cold leaderboard/publicread packages (39837), selected vet, Node syntax and diff checks PASS.
Native package run alone skips DB tests; the whole root composed target (76163) then PASSes
on real Postgres: 8/8 refresh cases, existing Fiscal/epoch/board cases and the new evidence
population (12 subcases), main DOM/Pitch/transitions/endings/continuation/recovery and Cosmetic
journey. New evidence cases cover exact bytes including distinct gzip metadata, caller-owned
buffers, multiple board categories, queue-only/archive-only refusal, absent/corrupt data,
pinned-hash mismatch, invalid JSON/identities and cancellation. Fixtures are synthetic evidence,
not a shipped verifier verdict or HTTP download. Post-run fixture DB sessions/owned listeners
are absent. Docker's non-PG overlay remains full; no deletion or Linux-image/full-CI claim.

Canonical docs and plan updated. Review by: Codex (diff/first-filter); Recorded by: Codex.
The complete producer/test/driver/docs/record range after `834369d7` requires designated review;
no archival or push. Next: C14 HTTP/manifest binding and actual verified-run download proof,
then continue the catalogs/generated-client/304 contract obligations. Full 1.0 remains active.

## 2026-10-07 — C14 public downloads and real verified-run replay

Batch after `bb84e355`, accepted C2/C4/C14/C19 and A8/C16. Register/mount genesis, replay-log
and nine-field verdict; preserve stored bytes, pinned references and content-hash headers.
Only board-authorized evidence is public. Private/unknown runs refuse identically; corrupt
evidence fails without caching. Existing shared limiter/cache tests now enumerate all six
public operations. Add raw-header collision refusal, real-account private-run refusal, and
HTTP downloads from the existing actual queue-verified Exit/board witness. No kernel, archive
producer, migration, balance, account policy, CI workflow or player copy change.

Verification: full `make test-game-ui-composed` (41927) PASS: 8/8 refresh cases, six mandatory
Postgres parent tests (including downloads/privacy), main gameplay/Pitch/recovery and Cosmetic.
Downloaded genesis/archive equal storage bytes and manifest hashes; replay passes with pinned
DB catalogs; corrupted receipt refuses. Initial implementation run rejected reversed path-param
order. Download checks then exposed two test-adapter errors: archive embeds compact JSON rather
than PG's spaced genesis, and full Founder archive history includes Fiscal events outside the
database verifier's Company/founder_advanced projection. Correct only those comparisons/adapters;
retain exact download/storage bytes and corruption rejection. Runs 52239/31541/23348/90340
remain recorded failures, not production fixes or green evidence. No retries or bounds loosened.
Cold seven affected Go packages (71228), selected vet, typecheck (47720), and diff check PASS.
Final publicread rerun (8906) PASS after preserving exact per-operation media-type assertions.

Regenerate OpenAPI/TS (17 operations). Old compatibility pin passes before refresh. `api-pin`
also included earlier Garden/UI additions; retain only its generated three evidence rows,
PublicRunVerdict and APIError response detail `run`, under C2's additive law and C4/C14/C19.
Previous baseline rows/schemas remain unchanged; unrelated additions are not re-baselined here.
Final staged `make api-check` PASS; fixture DB sessions and owned listener ports absent.
All verification handles terminal before edits. Cleanup check's first invocation used the
wrong fixture role and failed; the corrected declared user/database check succeeds.

Review by: Codex (diff/first-filter only); Recorded by: Codex. Designated review of the entire
range after `bb84e355` remains required; no archival, push or hosted/full-CI claim. Public
catalogs, TypeScript raw transport/reverification, cross-epoch retrieval and bodiless-304
registry metadata remain open. This is a same-epoch Go download proof, not the whole public
verification loop or full nine-tier 1.0. Next: catalog-owner descriptors and public catalog
serving under C18; no current-byte fallback for missing historical artifacts.

## 2026-10-07 — generated HTTP transport and default bootstrap/state consumers

Batch after `f5450259`, accepted A2/A5/C9/C19. RP-373 reproduces omitted raw-success types
(68082 red); generate Uint8Array successes, status/body associations and response metadata.
Embed one transport template in the generator: registered paths/method/auth/query/body,
JSON decoding or byte-preserving raw decoding with media/hash checks, explicit signal/request
ID, no token storage/refresh/retry/timeout policy. Bind actual Game UI bootstrap and state
reads; preserve existing snapshot parser, journal, credential ownership and revision handling.
No HTTP wire/pin/OpenAPI, kernel, balance, gameplay, copy or CI-workflow change.

Cold publicapi/publicread/account/gen-api (8739), selected vet, typecheck (62910,18852) and
client build PASS. Client tests (35884):132 pass across generated client/runtime/run-started/
intent outcomes. Native Chromium/WebKit (57736):88 pass, including browser raw SHA checks.
Wrong/missing hash, changed bytes, wrong media, malformed JSON, undeclared status and bad
runtime inputs refuse; network/abort propagates once. Bootstrap lost-reply and HTTP/parser
negatives retain the journal, and explicit retry reuses its key. Raw fixtures are controlled
fetch bytes, not an actual public TypeScript archive-verification journey.
Full real-Postgres/browser `make test-game-ui-composed` (75802) PASS: eight refresh cases,
six persisted parent cases (including public downloads/privacy), main gameplay/Pitch/endings/
recovery and Cosmetic. This does not close the earlier RP-368 limiter/reliability finding.
First draft runs85903/76732/7051 failed at the old whole-source query-word assertion,
two mistaken operation IDs and a compiler-too-complex cast. Correct IDs from registry, inspect
actual query declarations, and cast the unvalidated JSON result through unknown; no assertion,
request/status rule or threshold weakened to conceal a product failure. All handles terminal.

Regeneration preserves OpenAPI/compatibility-pin bytes; staged `make api-check` PASS.
Fixture DB sessions and owned composed/browser listeners absent. Review by: Codex
(diff/first-filter only); Recorded by: Codex. Entire range after `f5450259` needs designated
review, independently of preceding work; no archive/push/hosted-CI/full-AC4 or 1.0 claim.
Next: migrate registered Garden/Minigame/Soul callers through this generated boundary.
Missing route contracts, C9's raw-fetch lint, C18 owner descriptors/historical catalogs,
304 metadata and full public TypeScript re-verification remain open.

## 2026-10-07 — registered component ports use generated dispatch

Batch after `ec74899a`, accepted A2/A5/C9/C19: migrate nine Minigame/Soul/Garden calls;
preserve wire inputs, token ownership, UI error classes/copy/cadence and no-retry behavior.
Generate error categories from the existing descriptor. Bodiless reads omit Content-Type;
JSON parse errors omit reply bytes, while body-read failures propagate. No server/pin/OpenAPI,
gameplay, balance or CI-workflow change. RP-374/375 regressions failed before correction.

Checks PASS: 125 focused Node cases; 202 native Chromium/WebKit cases including actual Garden
refresh component; cold `make test-go GO_PACKAGES='./publicapi ./publicread ./account ./cmd/gen-api'
GO_TEST_FLAGS='-count=1'`, selected vet, typecheck, build and staged `make api-check`.
`make test-game-ui-composed` PASS
against real Postgres/HTTP/WebSocket: Pitch play, default locked Garden, gameplay/recovery and
Cosmetic. This does not prove full persisted Soul Recovery, active Garden or hosted CI.
First browser run failed because Chromium normalizes an errored Response stream; replace that
fixture with an explicitly injected body-read failure and assert exact propagation/no retry.
Correct a too-wide generic-body assignment and a misspelled Soul test selector; final named
populations execute all intended files. No production behavior or assertion was loosened.
Post-run DB sessions/listeners absent. All verification handles terminal before edits.

Review by: Codex (diff/first-filter); Recorded by: Codex. Entire range after `ec74899a` requires
designated review; no archival/push. Next: remaining route contracts/C9 lint, C18 catalog owner
descriptors and the undeclared 304 arm. Full API acceptance and nine-tier 1.0 remain incomplete.

## 2026-10-07 — RP-376 existing Soul error-descriptor correction

Batch after `27ffb9c7`, accepted A4/A5/C2: add only `soul_recovery` and `company_stream` to
APIError's detail enum; regenerate OpenAPI/TS. Handlers, statuses, copy and compatibility pin
unchanged. New tests compare exact real reply bytes with the same operation registry and
reject an invented detail. Existing composed preflight now requires the new Account test.

Old descriptor: unit32789 fails all eight unavailable/internal-mapper cases; composed15184
fails all five real authenticated/Postgres cases, including missing Company. Corrected cold
Account/publicapi/publicread/gen-api (37706), selected vet, typecheck19774 and 85 focused TS
cases65096 PASS. Real composed89770 executes/passes all five new cases, eight refresh cases,
remaining persisted parents and main gameplay; staged `make api-check` also PASSes without a
compatibility-pin change. Its Cosmetic phase then fails at the immediate
post-equip live-overlay assertion (RP-377). Full target remains RED. Persisted wearing had
passed; source inspection identifies an immediate DOM sample, not a proved cause or correction.
No retries, relaxed bounds or repeat-to-green. Owned listeners/fixture DB sessions absent.

Review by: Codex (diff/first-filter); Recorded by: Codex. Entire range after `27ffb9c7` needs
designated review; no archival/push/whole-CI or API-completion claim. Next: diagnose RP-377 in
the Cosmetic owner lane; missing route contracts/catalog descriptors/304 remain open.

## 2026-10-07 — exact v1 compatibility guard (RP-383/RP-384)

Outcome: accepted A1/A5/C2 requires rejecting optional-to-required changes and constraint
narrowing, with the affected field identified (AC2). Keep actual endpoints, schemas, generator
outputs and compatibility baseline unchanged; repair the comparison, not a re-baseline.

Regression77773 on old code: three response contexts admit requiredness promotion; request/
shared contexts reject but fail the field-name assertion. Four narrowing cases also admit a
one-unit change above2^53 or at signed int64 limits; both widening controls pass. UseNumber/
Int64 removes lossy metadata decoding; the same requiredness rule now applies to both uses.
Preserve nested field/union error context. First correction73853 still loses the union field
name, fixed by propagating the existing error; that red observation is not a passing gate.

Coverage: cold89847 `make test-go GO_PACKAGES='./publicapi ./publicread ./cmd/gen-api'
GO_TEST_FLAGS='-count=1'` PASS; vet71314 same packages PASS; actual `make api-check`16256 PASS,
all generated/pin bytes unchanged. This exercises the comparison and its real generator
consumer. No DB/browser/Decimal suite is substituted or claimed: no handler, persistence,
shipped TS, transport or production-arithmetic byte changes. Existing C18/C9/304/third-party
verification, owner/author and whole-1.0 gates remain. Current hosted CI is not rerun.

Review by: Codex (implementer first filter); Recorded by: Codex. Entire range after `73e1c9da`
through this guard/test/docs/record commit needs designated review; earlier ranges remain
separate. Diff inspected, handles terminal, no acceptance checkbox, pin refresh, archive,
push or release promotion. Next: remaining accepted API DTO/route integration and coherent
review ranges, not reinterpretation of the open numeric or owner decisions.

## 2026-10-07 — historical catalog immutable-byte source

Outcome: implement the C3/C11/C12 database source for the missing catalogs path. Only
epoch-accepted hashes are public; return exact stored artifacts in name order with per-artifact
digests, and validate the full length-framed bundle identity. One statement binds authorization
and bytes. No filesystem fallback, new wire schema, runtime HTTP route, kernel/balance change
or product mint. Formula availability and owner schema validation remain the later consumer's
responsibility, not grounds to fabricate current bytes under an old hash.

Coverage: focused native6201 passes the source controls (DB test explicitly skips). Real
Postgres61786 executes the new catalog integration and controls. Final cold95228
`make test-save-integration SAVE_TEST_PACKAGES='./leaderboard ./publicread'
SAVE_TEST_FLAGS='-v'` PASSes both complete affected packages, including mint/reconciliation,
board/projector/evidence and public HTTP regressions. Selected vet27212 PASS. Exact byte
and digest/order checks cover the committed set, a test-only newer formula-bearing epoch,
historical retrieval after that mint, repeat reads after caller mutation, and one hash accepted
by multiple epochs. Unknown/unaccepted/missing/corrupt sets, hash-matching invalid JSON/UTF-8,
extra/duplicate artifacts, cancellation and interrupted rows refuse without partial evidence.
No severing ceremony or unrelated browser/production suite needed for this new read-only
boundary; no public HTTP/TypeScript verifier or hosted CI result claimed. The fixture mint
neither edits release data nor proves replay-loader support for formulas.

Review by: Codex (implementer diff/first filter); Recorded by: Codex. Entire range after
`d2cc6709` through this source/tests/docs/record commit needs designated review. Prior review
debt remains separate. All handles terminal before records/commit; no archival/push. Next:
artifact-owner descriptors and the catalog HTTP binding, then the real public third-party
verification loop; C9/304 and full API/1.0 acceptance stay open.

## 2026-10-07 — shared formula model and closed owner descriptor

Outcome: C5/C11/C17/C18 needs an importable owner for the generated formula artifact, rather
than a second schema built by the API beside an executable-only model. Extract the existing
version-14 model into `server/formulas`; the generator's aliases bind to those same types.
Export closed named descriptors and a stored-byte validator, and have the actual generator
validate its output before writing. Required exact objects, arrays, nullable pinning, existing
owner enums, signed int64 parameters and existing canonical string formats are preserved.
No reflection, open-JSON descriptor, formula-content rewrite, catalog fallback or product mint.

Coverage: cold54302 `make test-go GO_PACKAGES='./formulas ./cmd/gen-formulas ./publicapi'
GO_TEST_FLAGS='-count=1 -v'` PASS; selected vet PASS. `make formulas-check`47472 executes the
real generator and proves published bytes unchanged; `make api-check`52548 also PASSes with
OpenAPI/TS/pin bytes unchanged. The extracted typed model round-trips the published bytes and
field order exactly. Tests exercise the current null pin, all three declared populated inputs,
populated source weights, fresh exports, 28 malformed structured controls and three malformed
byte populations without rewriting input. Existing source-fingerprint authority mutations and
catalog-composition tests still execute. This proves grammar/generation, not formula truth,
historical versions, an epoch mint, public HTTP or a third-party replay journey. No DB/browser/
numeric simulation changes require their unrelated suites; no hosted CI claim.

Review by: Codex (implementer diff/first filter); Recorded by: Codex. Entire range after
`9578aba3` through this model/descriptor/generator/tests/docs/record commit needs designated
review; preceding ranges stay separate. All handles terminal before records/commit. No archive,
push or kernel version signal: no watched simulation path or simulation output changed.
The nineteen current artifact owners still need their C18 exports; API-only edits under their
watched prefixes must not cause a false replay bump. Sequence those exports with the genuine
accepted formula-loader/artifact-set extension, or obtain an explicit path-scope resolution;
do not cite the draft Kernel History Guard Integrity RFC as authority. Next: complete the
catalog owner/loader binding, closed union and real public verification loop, not a partial
schema claim over the whole API.

## 2026-10-07 — optional stored formula replay binding

Outcome: implement the C5/C11 prerequisite for formula-bearing epoch bundles. Go/TypeScript
loaders accept optional stored version-14 formulas, validate the closed owner grammar and hash
the original bytes; Go resolution rechecks them. Historical sets without formulas remain valid.
Generate client metadata from the Go descriptor with decimal-string int64 bounds and preserve
number tokens during client validation. JSON last-value behavior and the Go decoder's nesting
limit are retained, including overwritten fractions/deep values; no rounded Number oracle.
This genuinely widens replay-input acceptance, so all three mirrors advance 0.3.167→0.3.168
and the new validator/metadata paths join the guard. No simulation arithmetic, balance bytes,
product epoch, public HTTP registration, compatibility pin or workflow changes.

Coverage: cold61414 affected formula/replaycatalog/kernel/generator packages PASS. Node57496
formula + full replay files:121 PASS. Native Chromium/WebKit35902:242 PASS. Final typecheck/
build28274 and selected vet PASS. Real Postgres82252 executes the entire replaycatalog package,
including stored formula fixture mint→DB load→resolution and public repository source→load,
old/new coexistence and hash-correct unsupported-version refusal. Shared26 controls cover
exact signed bounds/overflow, >2^53, required shapes, nullable/populated arms, malformed JSON,
last-value duplicates and strict string formats; additional nesting limits execute in both
runtimes. Historical caller aliases and wrong-identity cases refuse. `formulas-check`/`api-check`
52339 PASS, published formula/OpenAPI/TS/pin bytes unchanged; generated client formula metadata
has its own executed drift test. These are test-only mints, not a public downloaded replay
verification loop or formula-truth/balance proof.

The selected production replay run22306 initially fails two remaining strict-source research
comparisons. Apply RP-382's existing comparator to logged-policy/terminal tests: every recorded
outcome/receipt/state/population/acceptance byte still compares exactly, only historical source
digests differ visibly. No research artifact is restamped. Corrected run49949 PASSes the affected
replay population plus comparator negative controls. This does not repair the separate Clout
partition failures or claim the complete production suite green.

Limits: initial browser startup hit sandbox localhost denial. Permitted three-engine3170 records
230 pre-refinement passes but Firefox never connects; teardown exposes macOS sandbox/framebuffer
launch failure and is stopped by its exact identified PID. A mistaken repeated `--browser.name`
selector also exits before tests; final existing `--project chromium --project webkit` selection
is the 242-pass evidence above, not Firefox/CI. `verify-kernel-version`13767 passes its checkout
and adversarial fixtures then remains RED at historical `50a3a514` (RP-131). No bypass, history
rewrite, draft authority or hosted-green claim. Full 1.0 scope and prior review debt remain.

Review by: Codex (implementer diff/first filter); Recorded by: Codex. Exact range after
`556b4c99` through this implementation/tests/docs/record commit needs designated cross-party
review. All handles terminal before final records/commit; no archival or push. Next: complete
the catalog owner descriptors/closed response union and HTTP binding, then the public third-party
verification loop and protocol-compliant product mint; C9/304 and whole API acceptance stay open.

## 2026-10-07 — existing session operations join the generated authority

Outcome under accepted API A1/A2/A5/C1/C2/C7/C9 and Account D2/D3: register existing
`create_session` and `refresh_session`, reuse the account's `BootstrapSession` token-pair schema,
and mount those handlers only from their rows. Existing shared unauthenticated IP limiting,
body parsing, credential handling, token policy, response bytes and statuses are unchanged.
Add the real refresh/credential error literals to the shared response enum and bind these
operations' errors to exact status-specific bytes. Generated calls are explicit: no automatic
renewal, retry, credential storage/replacement or recovery copy. Other unregistered routes,
intent dispatch/C9 lint, catalog HTTP/owners, 304 and the full API remain unfinished.

Regression91009 fails both missing operations and response validation on old metadata. New
descriptor tests reject extra/private/missing/mistyped fields and cross-operation error pairs.
Existing parser controls now validate actual error replies; all persisted refresh controls
validate real replies against the registry. A new Postgres session-create→rotate→reuse case
checks genuine issued credentials, family row effects and no gameplay/failed-request mutation.
The shared-limiter regression proves creation and refresh still consume the same IP bucket.

Executed finished-batch evidence:

- Cold affected Account/publicapi/publicread/generator tests and vet pass (90224, then27946).
  Native tests skip DB cases; the actual DB results below are separate.
- Real Postgres54344 executes the entire Account package, including session creation, refresh
  expiry/reuse/New-Founder/limiter/closed-DB/lost-reply populations and the new registry case.
  All pass; no package PASS or skipped test is substituted for execution.
- Node46788:117 affected generated-client/operation-port/Game-UI-runtime checks pass.
  Native Chromium/WebKit2575:234 pass. Types19364 report zero errors/warnings. Actual production
  client build and shell/topology controls32181 pass. No Firefox or whole hosted-CI claim.
- Whole `make test-game-ui-composed`38984 passes: mandatory persisted parents, main player
  journey, both early endings/continuation/recovery, Fiscal/Pitch and Cosmetic adoption/care/
  reload. Source inspection corrects earlier “built-client” wording: this runner builds the Go
  server but serves the client through Vite; the separate successful client build is not a
  packaged-client browser rehearsal. The saved whole-product overview is corrected.
- Manual compiler census15712 passes all six actual callers and synthetic path/error removal
  refusals, schema version2; final71369 repeats successfully after the pin refresh. Its retained
  v1 evidence is unchanged, not restamped as current.

Pin authority and execution: API C1/C2 requires generated operation truth and additive v1
compatibility. The ordinary generator checked the old pin successfully before any refresh.
Initial combined90224 exits2 only because generated outputs were not yet staged; after staging,
actual `api-check`32181 passes. Regression3965 then demonstrates that the old committed pin
does not guard removal of either new operation or its 401 status. Explicit `make api-pin`
(existing accepted A1/A5/C1/C2 lane, no compatibility waiver) captures the current registry.
The semantic comparison finds three added rows: these two plus already-registered
`get_current_garden`; zero original operation-row changes/removals or removed schemas.
It also captures existing optional Garage shapes and prior additive error members. Those
preexisting feature ranges retain their own pending reviews; this pin does not approve them.
Cold27946 then passes actual committed-pin removal/status controls, affected packages/vet and
staged byte-identical `api-check`. No original pin is rewritten to conceal a rejected change.

**Review by:** Codex (implementer diff/first filter). **Recorded by:** Codex. Exact new range
is `e1580fdb` exclusive through this batch's final implementation/test/generated/docs/record
commit. Full designated cross-party review remains required; prior ranges remain separate.
No kernel/balance/migration/product mint, workflow or timeout/retry change, archival, publication
or push. Full nine-tier 1.0 remains active. Next: remaining operation/consumer migration and
catalog owner/HTTP/public verification integration, with explicit decisions for the held lanes.

## 2026-10-08 — existing account lifecycle joins the generated authority

Outcome under API A1/A2/A5/C1/C2/C7/C9/C17 and Account D1/D2/D3: register and exclusively mount
the existing account-create, current-Founder and New-Founder handlers. Shared unauthenticated/
failed-auth IP buckets and authenticated account buckets, no-store recovery response, body
parser, response bytes and archive/session behavior are unchanged. Explicit generated calls do
not add recovery UX, retry/renewal or deletion policy. Empty display stays an exact empty object.

Regression51119 fails all three absent operations on the old registry. Initial registration9396
also rejects actual timestamp fixtures: unlike bootstrap's fixed milliseconds, these handlers
use time.Time's zero-trimmed UTC millisecond encoding. Separate AccountCreated/Founder
descriptors and the date-time-utc-ms scalar format preserve that wire shape; zero/.1/.12/.123
codec cases pass and unsupported precision, offsets and noncanonical fractional forms refuse.
Existing fixed-width date-time-ms remains unchanged.

RP-386: compiler88005 exposes generated TypeScript `{}` admitting undeclared creation fields
(two unused @ts-expect-error negatives). Empty objects now generate Record<string, never>;
compiler44767 passes the negatives and actual valid callers, with zero Svelte errors/warnings.
This also corrects existing MinigameEmptyRequest typing, not its wire/parser semantics.

Finished-batch verification, all processes terminal before records:

- Cold39185 Account/publicapi/publicread/gen-api/Gameserver packages and vet78658 PASS. Native
  DB skips are not proof; full declared Postgres26572 executes both Account/Gameserver packages
  (4.898/34.289s), including creation → credential login → current Founder → replacement →
  refresh binding, actual archived Founder/stream rows, entropy-failure and refusal nonmutation,
  account/IP limits, and existing composed socket/verification/Garden/minigame populations.
- Node70987:122 affected SDK/port/runtime cases PASS. Chromium/WebKit41612:244 PASS, including
  no credential forwarding on account creation, exact method/body/auth and no creation retry.
  Initial sandbox browser command cannot listen and executes no tests; permitted run above
  completes. No Firefox, full browser/CI or assistive-technology result substituted.
- Normal generation79488/26854 accepts the old compatibility pin. New operation/status-removal
  controls261570 then fail all six new cases against that old pin. Explicit make api-pin6445da,
  authorized by A1/A5/C1/C2, adds exactly create_account/create_founder/get_founder (19→22 rows);
  semantic comparison882927 finds no old operation row change/removal. Actual pin controls
  pass in cold39185. Staged api-checkce45f7 is byte-identical; build-client9c99ec PASS214 modules.

Review by: Codex (implementer diff/first filter). Recorded by: Codex. Exact new range is
`17c552db` exclusive through this batch's implementation/test/generated/docs/record commit;
designated cross-party review remains pending and prior ranges remain separate. No kernel,
balance, migration, content mint, Actions, retry/timeout, owner-copy, archive or push change.
Known Clout/kernel-history CI failures remain; no whole-API/Account/1.0 acceptance claim.
Next: remaining HTTP caller/route migration and catalog owner/public verification integration.
The closed C19 schema|raw response model still cannot honestly register existing bodiless
DELETE successes or conditional304; amend that contract before describing those as JSON.

## 2026-10-08 — category artifact owner descriptor

Outcome under API A5/C17/C18: `leaderboard.CategorySchemas()` exports the existing v1
category grammar, binding each of the five IDs to its name/timer/predicate and closing nested
objects. No loader, balance, historical artifact, mounted route or generated contract changes.
Array cardinality/order, gate correspondence and canonical policy values remain loader checks;
the count-41 control explicitly proves schema validation alone is not policy validation.

Executed: `make test-go GO_PACKAGES='./leaderboard ./publicapi' GO_TEST_FLAGS='-count=1'`
PASS; `make vet GO_PACKAGES='./leaderboard ./publicapi'` PASS; `make api-check` PASS with
byte-unchanged generated outputs. A separate cold verbose `TestCategoryOwner` run executes
five actual catalog fixtures, all five row arms, 26 negative cases, semantic separation and
fresh deterministic exports. No DB boundary changed; native DB-dependent tests skip, and no
Postgres/HTTP/TypeScript third-party proof is inferred from these checks.

Review by: Codex (implementer diff/first filter). Recorded by: Codex. New review range is
`a7f8f54a` exclusive through this coherent source/test/docs/record commit. Designated review
remains pending. Categories is one of nineteen served artifact families, not a completed
catalog union. Next: remaining eighteen owner descriptors, then complete HTTP composition
and actual public-verifier journey. No content mint, archive, CI/workflow change or push.

## 2026-10-08 — public error authority no longer imports Account

Outcome under A4/A5/C10: move the one API-wide error descriptor from Account to `publicapi`;
both registries consume it directly. RP-390's production-import regression15902 fails at the
actual public→Account edge before the extraction. No handler, middleware, status, enum, exact
response bytes, simulation or mounted operation changed. This removes one concrete private
dependency; it does not declare full privacy acceptance.

Cold affected `./publicapi ./publicread ./account ./cmd/gen-api ./gameserver` tests41716 and
vet83539 PASS. `make api-check`79629 regenerates byte-identical OpenAPI/TS/pin outputs. Focused
verbose66671 passes actual public-package imports and schema enumeration, six forbidden-import
cases (aliases/dot/blank/subpackages included), four allowed/comment controls, parse refusal,
four existing error bodies, seven malformed error bodies and fresh-export checks. Native DB
cases skip; no DB behavior changed and no persisted/hosted/full-API result is substituted.

Catalog sequencing revalidated at7226ec7b: faction/guild C18-only exports would hit the guard's
whole-package prefixes despite changing no replay behavior. Seventeen remaining owner packages
are watched; moving schemas out to dodge that or adding a false version bump is not lawful.
This was already identified in the formula-model checkpoint above, not a newly implemented
exception. The historical Kernel History Guard Integrity RFC remains draft; prospective scope
authority is not inferred from it. Asked Marco for an exact-file independent-review exception
choice; no answer is inferred from preselection. Relevance's existing rational integer spelling
also needs reconciliation with the descriptor's strict int64 grammar before an exact export.

Review by: Codex (implementer diff/first filter). Recorded by: Codex. New range is `7226ec7b`
exclusive through this source/test/docs/record commit. Designated review pending; prior ranges
remain separate. No false replay version, history rewrite, CI/workflow edit, archive or push.
Next: resolve the catalog scope boundary; continue accepted remaining API integration meanwhile.

## 2026-10-08 — remaining intent caller has a schema-contract gap (RP-391)

Source inspection atce626440 traces the remaining handwritten `runtime.ts` intent fetch to
`account.submitIntent`, which returns the original persisted receipt bytes. Company's
`production.wireSnapshot` embeds typed maps keyed by catalog resource/generator/transition/meter
IDs; `founderCosmeticReceipt` embeds `cosmetics.equipped` keyed by runtime-created pet UUIDs.
These are dictionaries, not finite field sets derived from one current fixture. The current
C17 union has no dictionary arm; C19's raw response also requires headers this route does not
emit. No honest full registration follows from the present contract.

This is a source-confirmed architecture gap, not a measured gameplay failure or runtime schema
validation result. The actual Cosmetic/native purchase journey separately passes under its
existing contract. Required API author/owner reconciliation: explicitly key-constrained typed
dictionaries, or a versioned HTTP projection preserving original storage/replay receipts.
Neither is adopted here; do not introduce arbitrary JSON, silently re-encode old receipts,
register only the convenient intent kinds or evade the replay guard.

Editorial finding/route only; checked cited sources and diff, no software suite for this record.
No product, RFC body, acceptance, archive or push change. Next: obtain that contract choice;
safe work under other already-accepted contracts continues.

## 2026-10-08 — exact-schema validation retains duplicate-member evidence (RP-439)

Under A5/C17, reuse the package's existing raw duplicate-key checker before `ValidateJSON`
decodes objects to maps. Cold regression82742 fails on all seven duplicate response arms
and the first duplicate request: final-value map decoding erased invalid/ambiguous members.
Corrected regressions pass nine duplicate cases, including escaped names and nested refs/
unions; valid escapes, JSON-looking string contents and separate-object key scope still pass.
No new schema arm, middleware, handler output, stored receipt, kernel or generated byte changes.

Executed: focused `make test-go GO_PACKAGES='./publicapi'
GO_TEST_FLAGS='-count=1 -run TestSchemaRegistryRejectsDuplicate -v'` passes. Final cold
`make test-go vet GO_PACKAGES='./publicapi ./publicread ./account ./gameui ./gameserver
./cmd/gen-api' GO_TEST_FLAGS='-count=1'` passes all six packages and vet. `make api-check`
regenerates byte-identical OpenAPI/TS/pin files; gofmt and diff checks pass. These are validator,
consumer-package and local router checks, not Postgres integration, a new browser journey,
every handler's request admission, hosted CI or full API acceptance. Ordinary bug regression
already proves discrimination; no extra source mutations or unrelated full suites were run.

Review by: Codex (implementer first filter); Recorded by: Codex. Exact new range begins after
`84acb392` through this coherent repair/test/docs/record commit; designated review remains
pending, without relabelling prior ranges. Bodyless204/304, typed intent dictionaries, catalog
owner descriptors and the public third-party verifier remain separate contract/construction
gates. No checkbox, archive, release or push promotion. All process handles are terminal.

## 2026-10-08 — enforce existing activity/bootstrap request descriptors (RP-444)

API A5/C17 and Minigame API MA1/MA2 already own these exact request shapes. Bootstrap,
Minigame create/play/resolve and Soul start/progress/resolve/cancel now reuse original-byte
bounded validation before struct decoding/coordinator calls. Thirty controls extend the
existing composed HTTP lifecycle, comparing all rows in21 named credential/activity/save/
history tables before/after refusal. Ordinary Pitch payout/reconnect/retry and recovery
progress/resolve/watchdog still execute; valid bootstrap authenticates and retries byte-exactly.
The normal composed target now explicitly requires this lifecycle's PASS marker.

Old-source77832 demonstrates real malformed successes and stored mutations, including nested
Pitch duplicates and null resolution. Initial bootstrap placement interfered with the separate
three-member journey's shared IP budget. Relocation uses the lifecycle's existing24h watchdog
boundary; no limits/clock policy changed. Second old-source77199 confirms Minigame/Soul
failures but stops at the watchdog assertion before relocated bootstrap; it is not a second
bootstrap result. Focused corrected20545 passes all30 refusal cases and valid lifecycle.
Initial native24989 exposed a unit API literal missing its now-used registry; the fixture now
installs the real registry, with no production nil fallback or weakened assertion.

Final cold native62354 passes Account/publicapi/gameserver plus vet. `make api-check` passes
with unchanged generated bytes. Full real-Postgres37656 (`./account ./gameserver`, Integration
population) passes, including the30 controls and real socket reauthentication, no skipped gate
claimed. Whole `make test-game-ui-composed`46706 is RED: these persisted populations and
preceding browser journeys pass, but Garden's observer retains an incomplete upstream read
cancelled by reload. Later Tier2 is not reached; no whole-lane success is claimed. Follow-up
will diagnose that observer independently, rather than retrying this run into a green claim.

No schema/pin/command-enum/payout/rate/kernel/CI-workflow change; held Typer/Arcade public
wire choices, handwritten import/intents and historical empty-body routes are not adopted.
Review by: Codex (implementer first filter). Recorded by: Codex. Exact new range begins
after `13acd872` through this coherent repair/test/docs/record commit; designated review
pending, not conflated with preceding ranges. No checkbox/archive/release/push promotion.
Both full process handles are terminal. Next: repair the demonstrated Garden observer failure.

## 2026-10-08 — Keep HTTP admission out of historical artifact decoding (RP-449)

Outcome: API A5/C17/C18. RP-439 added HTTP duplicate rejection to the shared descriptor
validator, also used by formula artifacts. That unintentionally narrowed the previously
implemented Go/TypeScript last-value decoding contract. Its earlier affected-package selection
missed formulas/replaycatalog; the published failure reproduces locally. Move original-byte
uniqueness into registry request/response admission and exact-literal construction. Generic
descriptor validation again preserves the established artifact grammar. No formula fixture,
stored/generated artifact, schema pin, arithmetic, watched source, kernel version or policy
change; raw historical bytes remain identity-bearing and unchanged.

Regression: the prior cold formulas/replaycatalog run fails existing duplicate/overwritten
nesting controls. New shared-versus-HTTP regression42901 also fails the old code with
"shared decoder contract narrowed". Corrected Go formula/replay controls and unchanged
TypeScript controls pass. Existing escaped/nested duplicate HTTP refusal tests remain green;
an additional exact-literal member-repeat case prevents the construction path losing its check.

Coverage: cold99380 tests/vet pass publicapi, formulas, replaycatalog, both generators,
publicread, Account, Game UI and gameserver;60044 additionally passes leaderboard/vet.
Final publicapi tests also pass after adding the literal control. Node99072 passes all121
formula/replay tests. `make formulas-check api-check`10255 preserves generated formula,
client metadata, OpenAPI, TypeScript and compatibility-pin bytes. Real Postgres52249 executes
session request admission, composed Minigame/Soul request refusal/lifecycle and stored formula
replay; all pass. Its selected publicread package has no matching tests and contributes no DB
proof. Separate leaderboard91150 executes all nine historical catalog cases, including exact
stored formula bytes and corrupt/refused sets; all pass. First malformed shell selector exits2
before running tests, then is corrected. Gofmt/diff checks pass; all sessions are terminal.

Review by: Codex (implementer diff/first filter). Recorded by: Codex. Exact new range begins
after `c0809cab` through this coherent source/tests/docs/plan/ledger/checkpoint/log commit;
designated review pending. No whole-browser/whole-CI, public catalog HTTP completion, mint,
acceptance, archive, push or release claim. RP-131/RP-307/RP-450/RP-451 remain open; next
diagnose socket-revocation and Firefox failures independently under their existing contracts.

## 2026-10-08 — RP-391 typed-dictionary proposal, not implementation

Rechecked the remaining handwritten player-intent caller against `account.submitIntent`,
`production.wireSnapshot` and the founder cosmetic receipt. Existing saved response bytes
contain typed resource/generator/transition/meter and runtime pet-ID dictionaries, which
accepted C17 cannot describe. The draft `rfc/api-typed-dictionaries.md` specifies a closed
dictionary arm with existing key grammars, typed values, runtime admission, generated schema/
TypeScript, compatibility and immutability tests. It preserves saved receipts and the separate
stored-artifact decoder contract. Existing plans/logs own the work; no new execution program.

Owner choice between that preserve-receipt option and a separately designed versioned HTTP
projection was requested, not received or adopted. Complete intent registration, lawful
watched-owner exports and the other API gates remain separate; no open JSON, partial union,
history exception or implementation is authorized by the draft. Index, existing RP-391 row
and decision routing now link the proposal without changing their unresolved status.

Verification: descriptor/generator/compatibility source and current TS configuration inspected;
draft diff, local links, JSON example and whitespace checked. No software suite or implemented
schema support claimed for this proposal-only change. Review by: Codex (draft author/first
filter). Recorded by: Codex. New range begins after `75f5830b` through this checkpoint; owner
adoption and designated normative review remain pending. No acceptance, archival or push.

## 2026-10-08 — Refuse incomplete/case-aliased operational policy (RP-473)

Under API C20/A8, require every exact non-null root/nested member before Go struct
decoding. Missing/null trusted_proxy_hops previously became zero; case replacement
and collision were accepted across the policy. Existing values, integer bounds,
secret resolution, proxy topology, rate budget, duplicate-key rule and error remain.
Explicit zero and escaped canonical names still admit; public router construction
now refuses malformed policy before returning a handler, with a valid actual GET control.

Old-code focused regression fails on missing/null proxy and root/nested aliases.
Initial99789 overpopulated nested cases by reusing a nonnil map; corrected the test
before changing production, and the corrected old-source run still fails. Final
68-case matrix and positive controls pass. No source mutation ceremony was needed.

Cold root tests/vet24399 pass publicapi/publicread/gameserver/cmd/gameserver.
API drift64296 passes with byte-unchanged generated files. Declared Linux-amd64/
Postgres88884 passes publicapi/publicread and all but one gameserver case: live
revocation receives3012 no-pong rather than4001 (RP-474); aggregate exits2. Real
first-hour/verified evidence/proxy boundary/Garden/Minigame/Soul cases execute/pass,
not skipped. This red is retained, not replaced with the earlier native green.
The existing raw peer ignores application pings; repair that separately, without
weakening authentication assertions. Original RP-450 drain failure remains distinct.

Review by: Codex (implementer diff/first filter); Recorded by: Codex. Exact new
range starts after53f3f663 through this source/tests/docs/plan/ledger/log commit;
designated review pending. No browser, full hosted CI, clean-host, complete API,
new player feature or release claim. All handles terminal; no acceptance, archival,
kernel bump, config/content change, cleanup or push. Next: fix the demonstrated
raw-peer defect, then continue accepted integration with existing contract holds.

## 2026-10-08 — Select the declared SDK representation (RP-476)

Under A2/A5/C19, generated dispatch now selects by media type when one status has
multiple declared representations, before choosing JSON or opaque-byte handling.
Missing/undeclared media refuses before body reads. Single-schema decoding, raw
SHA256 requirements, application decoders, mounted endpoints and retries are unchanged.

The old generated client fails both new raw/raw and raw/schema cases with a gzip
type mismatch. Actual Go registry/generation accepts both populations and retains
their metadata. Tests also retain missing/wrong raw hash and missing/unknown-media
refusals; temporary synthetic operation metadata is restored in finally. This is
SDK conformance, not a new server representation or full runtime schema validation.

Final affected checks: Go publicapi/cmd/gen-api tests at -count=1 and vet6878 pass;
Node42573 passes160; native Chromium/WebKit50848 passes346; strict types, client
build and shell boundaries10480 pass. api-check passes after staging the intended
generated TS; OpenAPI and compatibility-pin bytes are unchanged. Real main journey
33328 passes with the production client bytes, Go server, Postgres and WebSocket:
both early terminal states/next Company, native purchases/Fiscal/Pitch, committed
reply-loss and actual server-process restart recovery, attended Soul and cancellation.
The driver also executes its eight refresh and ten named DB boundary populations,
with no skips. Expected proxy ECONNREFUSED during deliberate process death is
followed by successful recovery. This is the main driver, not the aggregate's
separate Cosmetic/Clout/Reputation/Garden/Tier-2 journeys or hosted CI.

Diff inspection additionally found RP-477's same-status OpenAPI header aggregation
problem. It is recorded as source analysis only, not repaired or executed evidence.
The complete catalog reader remains held on lawful artifact-owner schema exports;
no guard exception or unanswered owner choice is assumed.

Review by: Codex (implementer diff/first filter). Recorded by: Codex. New range
begins after59e5cc49 through this coherent source/tests/docs/plan/ledger/log change;
designated review pending. All check handles terminal. No acceptance, archival,
kernel/balance change, push, deployment, whole-CI green or release claim.

## 2026-10-08 — Preserve representation-specific OpenAPI headers (RP-477)

A2/A5/C19 generation now aggregates every raw hash header at a status, coalesces
case-insensitive names and marks a header unconditionally required only if every
alternative declares it. Otherwise generated prose names the requiring media type;
the runtime/SDK still require the selected raw descriptor's hash. This follows the
status-level response/header structure in OpenAPI3.1:
https://spec.openapis.org/oas/v3.1.0.html#response-object and #header-object.

Actual NewRegistry→GenerateOpenAPI regression on old source fails distinct-header,
case-alias and raw/schema populations; single/shared raw and schema-only controls
pass, with a separate404 proving per-status isolation. Initial21495's lowercase
first character violates the existing descriptor grammar; the corrected legal
X-content-sha256 population reproduces the overwrite before any production change.
Final full publicapi/cmd/gen-api -count=1 tests26196 and vet6699 pass, including all
six populations. api-check preserves mounted OpenAPI/TS/compatibility bytes exactly.
Gofmt/diff checks pass. No browser/DB rerun: only generation changes; no mounted
operation gains alternatives, and no decoder, HTTP response or runtime hash check moves.

Review by: Codex (implementer diff/first filter). Recorded by: Codex. New range
begins after1b563d68 through this source/tests/docs/plan/ledger/log commit; designated
review pending and consolidatable with the preceding representation-dispatch fix.
No whole-API acceptance, hosted-CI green, archival, kernel bump, publication or push.

## 2026-10-09 — Executed bodiless-response gap; bounded amendment drafted (RP-483)

At source `b9613d1b`, rechecked remaining accepted API work rather than infer approval of
unanswered dictionary/kernel proposals. Existing C19 cannot represent actual204/304 without
lying about body/media type. `api.go` still writes bare204 for both DELETE lifecycle handlers;
`Runtime.WriteCached` writes304 before rate limiting, without a content body/media type.

Executed `make test-go GO_PACKAGES='./publicread' GO_TEST_FLAGS='-count=1 -run
TestMountedPublicReadsCacheWithoutSpendingTokensAndInvalidateChangedBytes -v'`76208:
all six mounted operations pass cold, including budget-exhausted304, exact metadata and
changed-byte invalidation. Controlled readers, not DB/proxy/hosted integration. A direct
Node26 import of the actual generated `types.ts` then calls `list_public_epochs` with a
fetcher returning `new Response(null,{status:304})`: exact rejection is `undeclared API
response status (304)`, one fetch. `operations` has neither DELETE account nor DELETE session.
This is an executed missing declaration/dispatch, not evidence of a newly broken server.

Draft `rfc/api-bodiless-responses.md` defines the closed empty arm, strict byte validation,
OpenAPI/status/undefined typing, explicit If-None-Match input, additive pins and unchanged
handler registration. Checked RFC9110 sections15.3.5/15.4.5 and OpenAPI3.1's no-return-value
response example against the current writer/generator; primary links are in the draft.
Reuse API's existing plan/log; no new program, test framework, acceptance flip or product change.
The existing logged gap now has ledger/decision/index links. Owner question delivered; no answer
or authority is inferred. Full proposed acceptance is not executed before adoption/build.

Review by: Codex (proposal author, actual-diff first filter). Recorded by: Codex. Exact draft/
record range begins after `b9613d1b` through this containing commit, pending owner adoption
and designated normative review. No generator/SDK/handler/pin/replay bytes changed; no full
API, current-head CI, rights completion, archival, push or release claim. All check handles
terminal before edits. Next: the specified implementation after those authority gates; continue
other safe accepted work without treating this draft as a contract or shrinking the1.0 goal.

## 2026-10-12 — Reject raw hash headers conflicting with protocol metadata at construction

Outcome: RP-485 conformance fix under accepted A5/A8/C19. Registry construction admitted
digest headers named ETag, Cache-Control, Content-Type, Content-Length or X-Request-ID,
while the existing cached writer already rejected those names case-insensitively. Share
that existing grammar/reservation rule; invalid operation definitions now fail before
mounting or generation. Writer admission and current emitted bytes are unchanged.

Corrected old-source baseline78464 fails all11 reserved/case-variant definitions while
six valid-header/JSON-or-gzip controls pass. The first run9250 also had a test-authoring
error (nil required request-ID dependency); it is not a healthy control baseline. The
retained mounted tests use actual registry/chi/middleware/writer paths and require exact
raw bytes/digests, content type/length, quoted ETag, cache policy and caller request ID.
Existing304 controls retain metadata and no body; no304 registry arm/SDK policy is added.

Final root `make test-go GO_PACKAGES='./publicapi ./publicread' GO_TEST_FLAGS='-count=1 -v'`
26059 passes both full affected packages, including existing mounted cache/evidence/privacy
checks. Neither package has a skip path. `make vet GO_PACKAGES='./publicapi ./publicread'`
51206 and `make api-check`9129 pass; OpenAPI, compatibility pin and generated client remain
byte-unchanged. No DB/private gameplay/browser/balance boundary changed or claimed tested.
No observed metadata overwrite/leak in current endpoints is alleged; this is startup
admission of definitions the runtime already refuses. All handles terminal, diff checked.

Review by: Codex (implementer first filter). Recorded by: Codex. Exact new implementation/
test/docs/records range is `da3a922d` exclusive through this containing commit; designated
cross-party review pending. No draft adoption, whole API/hosted CI, archival, push or release
claim. The typed-dictionary/bodiless/history proposals and earlier corrective review debt
remain separate; continue safe accepted integration work.

## 2026-10-12 — Preflight the whole route binding set before mounting (RP-486)

Outcome: accepted A5 conformance repair. `Registry.Mount` now checks every binding and
required auth middleware before wrapping handlers or registering routes. Valid mounts
retain their order/auth; invalid bindings cannot leave a partial route set in the caller's
router. This covers validation errors, not transactional rollback of middleware/router panics.

Old-source focused run15453 fails all three late-invalid configurations: the early route
still serves200 after Mount returns an error; valid control passes. The retained real-chi
regression also preserves a preexisting route and checks handler/wrapper counts. Final
focused run passes all four cases. Full root `make test-go GO_PACKAGES='./publicapi ./publicread'
GO_TEST_FLAGS='-count=1 -v'`83341 passes both affected packages, including existing authenticated
mount/cache/refusal tests. `make vet GO_PACKAGES='./publicapi ./publicread'` and `make api-check`
pass; generated OpenAPI, pin and client bytes are unchanged. No DB/gameplay/browser boundary
changes or full CI claim; current composition callers' error handling is not alleged broken.

Review by: Codex (implementer first filter). Recorded by: Codex. Exact new source/tests/docs/
records range is `3b5ce376` exclusive through this containing commit; designated review pending,
consolidatable with RP-485 as `da3a922d` exclusive through this commit. Earlier review debt and
the three draft-contract adoption gates remain open. All test handles terminal; diff checked.
No acceptance, archival, push, deployment or release claim.

## 2026-10-12 — Designated review: original query descriptors (RP-492)

Review by: Codex (designated independent reviewer of Claude's implementation).
Recorded by: Codex.
Reviewed range: `32560c4b^..32560c4b` (exact seven changed paths).
Decision: CHANGES REQUIRED.

Inspected the descriptor/registry, validator, OpenAPI/TS generation, pins, tests and docs.
The decoder uses `strconv.Quote`, whose Go-only escapes are invalid JSON. The retained
`TestQueryStringEncodingPreservesDescriptorValues` first runs against the unmodified decoder
at `6d2ffb99`; that file is byte-identical to the original query implementation, verified
against `e63796d5`. Session61158 fails six descriptor-admitted strings and an explicitly
declared escaped enum member; ten other string controls and four refusal controls pass.
Each positive control first satisfies the actual JSON descriptor and URL round-trips exactly.
This is execution of the unchanged original decoder, not a whole historical-checkout run.

This verdict covers only the original query commit, not `cd5ba06d`, epoch-reader composition
`e63796d5`, interleaved F9 or later corrective batches. No current mounted ASCII filter/cursor
failure is alleged. The finding needs a conformance repair, not an invented string restriction
or adoption of any draft. Original-query approval remains open until corrective review.

## 2026-10-12 — Preserve exact descriptor-admitted query strings (RP-492)

Outcome under A5/C17: JSON-encode string values before the existing shared validator, with
explicit valid-UTF-8 admission so the encoder cannot silently repair malformed bytes. Required,
empty, duplicate, integer/boolean and descriptor enum/format behavior otherwise remains.
No new schema kind, status, handler, query metadata, saved receipt or compatibility pin changes.

Final focused regression12971 passes all16 string cases, four refusal cases and enum positive/
negative controls. Full `make test-go GO_PACKAGES='./publicapi ./publicread ./cmd/gen-api'
GO_TEST_FLAGS='-count=1'`79284 passes; affected `make vet` passes. `make api-check` passes with
byte-unchanged OpenAPI/TypeScript/pins. Existing `client/node_modules/.bin/vitest run --root
client test/generated-api.test.ts` passes28, including URL/query serialization and refusal.
No DB, browser input, gameplay or hosted-CI boundary changed or claimed tested.

Review by: Codex (implementer first filter, not designated approval of this correction).
Recorded by: Codex. Exact correction range begins after `6d2ffb99` through this containing
source/test/docs/plan/ledger/log commit; Claude's designated review remains pending and may
consolidate related API corrections. All handles terminal; diff checked. Full API acceptance,
remaining original-reader review, prior review debt and draft adoption stay open. No archival,
push, publication or release claim; continue accepted work without shrinking the1.0 objective.

## 2026-10-12 — Designated review: original public epoch-reader commit

Review by: Codex (designated independent reviewer of Claude's implementation).
Recorded by: Codex.
Reviewed range: `e63796d5^..e63796d5` (exact fourteen changed paths).
Decision: CHANGES REQUIRED.

Inspected original SQL/page/HTTP descriptors and tests, registry merge, shared error extraction,
generation entry point, all three generated diffs and docs/log. SQL, its integration/unit tests
and merge implementation remain byte-identical to this commit. Real declared Postgres run29236
executes `TestPublicEpochPageIntegration` with no skip: newest-first pagination, strict keyset
continuation, exhausted non-nil list, timestamps/null, byte-sorted accepted hashes, exact staged
changelog bytes, bounds and missing-changelog refusal pass. Native hash test65741, full cold
leaderboard88550 and vet pass; native DB tests in that full package are not counted as DB proof.
The current API suites/generation checks are recorded in RP-492 above; current handler tests
94782 additionally pass all five epoch response/cursor/limit/failure/cache/registry cases.

Two original-batch blockers are not disguised by those passes. RP-390: the original public
reader imports private Account for the error descriptor; the later shared-publicapi extraction
already repairs that edge but still needs its own corrective review. RP-483: the reader serves
the C16-required304 while its operation/OpenAPI omit that status. Fresh direct invocation of
the current actual generated client still rejects a controlled304 as `undeclared API response
status (304)`, exactly one fetch; this is dispatch evidence, not an HTTP/DB journey. The bodiless
amendment remains draft, not authority to bypass conformance or fabricate a JSON body.

No new defect row duplicates these existing findings. The original predeclaration `cd5ba06d`
was inspected as context, not promoted to implementation approval; query `32560c4b` has its own
verdict above. Interleaved F9 and later router/boards/routes/privacy commits are outside this
review. This records a completed review with required corrections, not an archival approval,
full public privacy, historical-checkout execution, hosted CI or a release claim. Next: review
the existing RP-390 correction and the adopted304 amendment when authority permits; safe
accepted work continues. No product bytes, owner rulings or authored copy changed here.

## 2026-10-12 — Targeted board cursor evidence review and repair (RP-493)

Review by: Codex (designated reviewer of Claude's original acceptance evidence).
Recorded by: Codex.
Reviewed range: `4c316c4c^..4c316c4c`, scoped to the public board handler/cursor test;
the helper composition change in `09d12987` was inspected as context.
Decision: CHANGES REQUIRED (targeted C13/C15 evidence, not the full eighteen-path batch).

The changed-epoch arm expects404 because its fake only admits epoch8. It therefore cannot
prove a cursor binds the epoch. With only `EncodeBoardFilter`'s epoch field replaced by a
constant, the old focused test91236 still passes. The corrected fixture admits epoch9 and
proves each changed filter works without a cursor before requiring exact invalid/cursor
refusal with the old cursor; none of these refused requests may reach page retrieval.
The same compiling fault21227 fails with `epoch-mutated filter accepted the cursor: 200`.
Both probes restore the exact production source in a finally block; no production diff remains.
Healthy corrected focused59846 passes. Initial22153 used Values.Clone, unavailable to the
repo's Go1.26 language version; replaced with supported maps.Clone, not counted as a gate run.

Full cold `make test-go GO_PACKAGES='./publicapi ./publicread ./leaderboard ./routeprojection
./gameserver' GO_TEST_FLAGS='-count=1'`80934 and affected vet pass; native DB skips are not DB
evidence. `make api-check` preserves generated/pin bytes. Real declared Postgres56275 executes
the public board/route/epoch/catalog/evidence reader populations without skips. Initial regex
selection failed at the Make recipe's shell before any test; the corrected `-run TestPublic -v`
selector is the credited run. Separate `make test-go-ci CI_TEST_PACKAGES='./publicread ./gameserver'
CI_TEST_FLAGS='-run "TestBoardReader|TestComposedGameserverPostgresSocketClearingAndGCIntegration" -v'`
62894 passes all three board-handler functions and the actual composed Postgres/socket/privacy/
clearing/GC function on the declared Linux-amd64 Go1.26 service. This approximates CI, not a
new hosted verdict, whole server job or full public/privacy/release acceptance.

The review range inventory contains interleaved Reputation commits; these are excluded, never
silently included in an API verdict. Original `731bb3e7`, `4c316c4c` and `09d12987` full-range
reviews remain incomplete; this finding approves none of them. RP-367's earlier row-error repair,
RP-390's shared error extraction,304/dictionary/rotation questions and prior review debt remain
separate. No new test framework, product behavior, generated wire, policy or authored copy change.

Review by: Codex (test correction implementer first filter, not independent approval).
Recorded by: Codex. The correction's exact test/plan/ledger/log range starts after `737e0b2e`
through this containing commit and awaits Claude review. All handles terminal; diff checked.
Next: continue the remaining original public-router/board/route review and accepted construction;
no archival, publication, push or narrowing of the full nine-tier1.0 goal.

## 2026-10-12 — Designated review: Route Registry reader and structural privacy

Review by: Codex (designated independent reviewer of Claude's implementation).
Recorded by: Codex.
Reviewed range: `09d12987^..09d12987` (exact sixteen changed paths).
Decision: CHANGES REQUIRED (RP-494).

Inspected source, handler/SQL tests, registry wiring, composed enumeration, all generated diffs
and docs/log against A3/C10/C12/AC5. The four route reader/test files remain byte-identical to
this commit. The prior real-Postgres56275 result applies to that unchanged source: exact-page
continuation, approved-name control, pending-name withholding and deleted/orphan executor
withholding pass. Execution count is the existing Registry aggregate, not a newly invented
unique-player count or adoption curve. No new reader/privacy-policy defect is alleged.

The structural privacy witness does not enforce its claimed publication boundary: it accepts
innocent-looking extra state, skips APIError, and discards parent paths during recursion.
RP-494's actual valid-registry fixtures reproduce three false admissions on the unchanged
admission rules, extracted into an error-returning helper for testing; the existing forbidden
account_id control refuses. This executes the original checker logic on the current registry,
not a whole historical checkout. The composed value scan for a private, unverified Founder
does not make an independent structural field gate unnecessary. Full API/304/rights and prior
corrective review remain separate; router `731bb3e7` and boards `4c316c4c` are not approved here.

## 2026-10-12 — Structural privacy field gate corrected (RP-494)

Outcome under A3/C10/AC5: independent C12/C14 publication expectations for all current named
public schemas, including APIError, plus full nested paths. Existing private-name exclusions
and public-identity checks remain. These are test expectations, not another runtime registry
or permission to widen publication. No handler, descriptor, generated or persistence byte changes.

Old gate66101 admits route cash, error cash and nested reuse of an allowed name. Final five
negative fixtures refuse while the actual healthy registry passes before and after independent
exports. The optional additions also pass the actual compatibility check: compatibility alone
is not privacy authority; the nested type change separately fails compatibility. Focused40145,
full cold publicapi/publicread/gameserver38457, affected vet and byte-unchanged api-check pass.
Selected Linux-amd64 Go1.26/Postgres final25296 passes the import/structural/refusal tests and
`TestComposedGameserverPostgresSocketClearingAndGCIntegration`, executing the current six-operation
private-read enumeration and real socket/clearing/GC path without skips. Earlier52670 also passes.
No raw-body/gzip content, public verified-history policy, full CI or release acceptance follows.

Review by: Codex (correction implementer first filter, not designated approval).
Recorded by: Codex. Exact test/docs/plan/ledger/log correction range starts after `6af2830e`
through this containing commit, pending Claude review and consolidatable with RP-493. All check
handles terminal; diff checked. Next: remaining original router/board review and accepted build
work. No owner-copy change, draft adoption, archival, publication, push or reduced1.0 scope.

## 2026-10-12 — Designated review: original public board reader

Review by: Codex (designated independent reviewer of Claude's implementation).
Recorded by: Codex.
Reviewed range: `4c316c4c^..4c316c4c` (all eighteen changed paths; supersedes the earlier targeted-only review scope).
Decision: CHANGES REQUIRED.

Inspected the SQL query extraction/shared pinned-category loader, page/filter/key DTOs, handler,
registry/composition changes, all tests, generated OpenAPI/TS/pin and docs/log against A6/C12/C13/C15.
The original omits iteration-error handling (RP-367), imports Account for its error descriptor
(RP-390), omits the served304 status (RP-483), and has the vacuous epoch-cursor assertion (RP-493).
Later local corrections do not convert this original range into approval; the304 contract remains draft.

Cold affected Go32942, vet and api-check pass on current source. Real Postgres83923 executes
`TestPublicBoardRankingAndPagesIntegration`: pinned ranking kind, ties/keyset continuation,
variable partitioning, magnitude order, exactly-full exhaustion and invalid stored catalog refusal.
SQL helpers and that test are byte-identical to the original; ranking lookup differs only by
RP-367's error propagation. Linux-amd64/Postgres54941 also executes that population and the
existing epoch/legacy-board witness (ties, sub-quantum magnitude, negative exponent and zero).
Linux19895 passes all three handler functions and the actual composed Exit→verified board/evidence
round trip; it is current integration, not a historical checkout or the missing third-party TS loop.

Independent JSON comparisons preserve every existing OpenAPI operation/schema except the four
new response-error values and add exactly one board operation/five schemas. My first pin assertion
failed because it wrongly assumed only those four values widened: the refresh also captures five
already-served Typer values from `d7c1ce6e`. Comparing against the predecessor's actual OpenAPI
confirms that provenance; all other existing pin operations/schemas are unchanged. No breaking
pin reset, new count-category producer or mandate producer is inferred. The original log's exact
error/path ratification questions remain visible. Reputation commits and other API ranges are excluded.

## 2026-10-12 — Designated review: original public router and deployment wiring

Review by: Codex (designated independent reviewer of Claude's implementation).
Recorded by: Codex.
Reviewed range: `731bb3e7^..731bb3e7` (all twenty-three changed paths).
Decision: CHANGES REQUIRED (RP-495/RP-496).

Inspected resolver/router, gameserver entry/composition, production/development config, secret
schema/template/closure, Caddy fixture, composed/browser-driver changes and docs/log. Native
affected packages1366 and vet pass; declared Linux-amd647452 runs config/closure/template/bundle
and entry-point tests. Linux/Postgres12186 executes router/rotation controls, missing-key refusal,
public read/cache/request-ID/privacy and real socket/clearing/GC plus the one-proxy boundary.
`make test-deployment-release`2347 passes the actual Caddy/Postgres drain, release and exact backup
rollback fixture; its isolated temporary test services are removed by that target. This is not
R-006, rebuilt release bundles, the retained R5 range's approval or an official-host deployment.
The changed composed browser driver was inspected, not rerun or claimed green in this review.

RP-495 promotes the original disclosed DESIGN-GAP6 to the shared defect ledger: fixed runtime
aliases conflict with the continuous never-reused rotation IDs and the overlay omits cursors.
Existing native rotation/overlap tests32152 pass, proving their own mechanism, not compatibility
with the runtime aliases. The ruling authors must reconcile their contract; this review invents
no new IDs, alias swaps or shorter overlap. RP-496 is separately reproduced and corrected below.
Later readers, middleware corrections, publication/rights contracts and archival remain separate.

## 2026-10-12 — Reject incomplete direct-composition cursor pairs (RP-496)

Outcome under A7/C20 and Deployment's complete-pair contract: only a wholly absent previous
pair may use the first-deployment alias. The existing resolver now refuses an orphaned ID or
material; production environment admission, wire bytes, policy names and key overlap are unchanged.
On old source1734, resolver and real-router assertions fail: orphaned material resolves both
names and mounts with nil error. The missing-material control already refuses. No secret contents
are printed. Full corrected cold publicread/publicapi/gameserver/deploymentconfig/entry51881,
vet and byte-unchanged api-check pass. Final Linux-amd64/Postgres72921 runs both half-pair
composition refusals and valid real HTTP/socket/privacy/clearing/GC, proxy and verified-run paths
without skips. Existing first-deployment, current/previous and retired-key controls stay green.

Review by: Codex (correction implementer first filter, not designated approval).
Recorded by: Codex. Exact correction/records range starts after `36f81dae` through this containing
commit and awaits Claude review. Original independent verdicts above do not approve this repair.
All handles terminal; no full hosted CI, browser, clean-host, draft adoption, archive, push or
release promotion. Next: remaining accepted API construction/later review; full nine-tier1.0 stays active.

## 2026-10-12 — Refuse explicit duplicate cursor secrets in direct composition (RP-497)

Outcome under API A7/C20 and Deployment DP3: an explicitly supplied previous cursor pair
must differ from current; only a wholly absent previous pair may use the first-deployment
alias. The environment decoder already enforces this, but direct Go composition bypassed it.
Old-source focused32212 fails both the resolver and actual-router duplicate-value assertions
(`err=<nil>` mounted); duplicate-ID and existing half-pair controls already refuse. Reject
duplicates before resolving either name, without changing policy IDs, key bytes or overlap.

Full cold `make test-go GO_PACKAGES='./publicread ./publicapi ./gameserver ./deploymentconfig
./cmd/gameserver' GO_TEST_FLAGS='-count=1'`73232, affected vet and unchanged api-check pass.
Declared Linux-amd64/Postgres1248 (`make test-go-ci`, publicread/gameserver, selected cursor,
router, evidence HTTP and composed socket/clearing/proxy/verified-run tests) passes with no
skips. Actual composition checks both duplicate cases, then the valid real HTTP/socket path;
the existing first-deployment and distinct-key rotation witnesses pass. Download evidence
remains byte-identical to storage and the corrupted receipt refuses. The same-epoch Go
re-verification uses pinned catalogs from DB, not the missing public catalog/TypeScript loop.

Review by: Codex (implementer first filter, not independent approval). Recorded by: Codex.
Exact supplement range starts after `1ea43c67` through this containing commit; Claude may
consolidate it with the preceding `36f81dae..2fbd246b` cursor correction. No self-review
promotion, RP-495 reconciliation, browser/full hosted CI, clean-host rehearsal, draft adoption,
archive, publication or release claim. All verification handles completed before this record.

## 2026-10-12 — Designated review: Claude's A6/AC4 body reconciliation

Review by: Codex
Recorded by: Codex
Reviewed range: `266b8d8c^..266b8d8c`
Decision: APPROVED

Inspected both changed files, the complete original diff and the affected A2/A5/A6,
AC4, C9/C12/C17/C18 and formula rulings. A6 now requires exactly the already-ruled
artifact-owner descriptors in a name-discriminated closed union; AC4 repeats C9's
generated-only HTTP caller/lint criterion instead of demanding deletion of an
absent client layer. No owner ruling bytes, schema DSL, HTTP bytes or implementation
are changed. The active RFC remains byte-identical to the reviewed revision.

Re-derived the changelog's nineteen-family claim directly from
`balance/epochs/phase0.json` at the predecessor and current `8f737115`: both name
current epoch8 and the same nineteen artifacts, with no `formulas` artifact. The
older C18 seven-arm sequencing note is explicitly historical/conditional, not
authority to compose a partial current union or pretend a formula mint happened.
Current served names still need complete owner exports; acceptance of this record
does not approve missing implementations, generic JSON or a new mint.

`git diff 266b8d8c^ 266b8d8c --check` passes; compared the ruling/body text and
manifest sources rather than running unrelated software suites for Markdown.
This closes only the outstanding normative correction review. Generated-caller
migration/lint, catalogs and public TypeScript re-verification, missing204/304
contracts, RP-495 and later corrective review ranges remain separate. No draft
adoption, feature acceptance, archival, push, deployment or 1.0 promotion.

## 2026-10-12 — HTTP conditional-cache conformance (RP-502)

Outcome under accepted A8/C6/C16: valid weak tags, tag lists, repeated field lines and
wildcard conditional GETs retain the existing bodiless304/cache/hash/request-ID behavior
without spending an IP token. Parse the complete combined field, preserving quoted commas
and valid empty list members; malformed/nonmatching conditions take the ordinary bucket path.
Authorization and evidence validation remain before caching. No policy values, receipt/replay
bytes, endpoint DTO, generation pin, retry, clock, kernel, schema or CI workflow changes.
This is ordinary HTTP conformance, not adoption of RP-483's empty-response schema amendment.
Primary semantics checked against RFC9110 §§8.8.3,13.1.2 and recipient list grammar §5.6.1.

Extend the existing policy and six-operation cache tests, not a new framework. Initial62559
also exhausts the fixture's finite request-ID entropy; later500s are instrumentation failures,
not caching findings. Use explicit valid request IDs and retain exact echo assertions.
Definitive unchanged-source64592 fails nine policy cases (eight valid conditional forms plus
a malformed second field ignored after a matching first) while fourteen controls pass; all
six mounted-operation cases fail on weak revalidation returning200 instead of304.

Corrected cold `make test-go GO_PACKAGES='./publicapi ./publicread' GO_TEST_FLAGS='-count=1'`
26763 passes. Final root `make test-go vet api-check GO_PACKAGES='./publicapi ./publicread
./gameserver' GO_TEST_FLAGS='-count=1'`6788 passes all affected native non-DB checks/vet and
byte-unchanged generated contract/pin/types. Native database tests skip without their URL;
they are not cited as database proof. Declared `make test-go-ci CI_TEST_PACKAGES='./publicapi
./publicread ./gameserver' CI_TEST_FLAGS='-v -run TestCachedMiddlewareSkipsLimiterFor304AndEchoesRequestID\|TestMountedPublicReadsCacheWithoutSpendingTokensAndInvalidateChangedBytes\|TestComposedGameserverExitVerificationAndBoardIntegration\|TestComposedGameserverPostgresSocketClearingAndGCIntegration'`
81527 passes all four named functions, without skips, on Linux-amd64 with real Postgres.
The evidence helper exercises five conditional forms on all three real queue-verified
downloads, exact stored bytes/hashes and replay/corrupt-receipt refusal. Actual private/unknown
run requests with wildcard still refuse identically with404 and no cache/hash. Mounted tests
exhaust/refill the actual configured bucket and retain changed-byte invalidation. The declared
emulation CPU setting is unchanged; this is not native hosted hardware or whole Actions CI.
Catalogs still come from the pinned DB bundle, not the missing public TypeScript/catalog loop.
The Compose orphan warning caused no cleanup. All handles terminal; format/diff checked.

Review by: Codex (implementer first filter). Recorded by: Codex. Exact new range starts after
`aaf78306` through this containing commit and needs Claude designated review. Earlier review
debt stays separate. No draft adoption, full API/1.0 acceptance, archive, push or deploy.
Next: continue accepted API/player construction and consolidated corrective review; held
owner/author choices and release-artifact gates remain explicit, not a blanket work stop.

## 2026-10-12 — Preserve generated operation paths across URL parsing (RP-506)

Under A5/C1/C9, generated dispatch must retain its registered operation path. Exact
`.`/`..` parameters survive encodeURIComponent but the Request URL parser rewrites
them as navigation. Old-source Node b34abd fails four cases: public-board requests
become `/boards/` or `/api/public/v1/`, and authenticated minigame creation becomes
`/api/v1/minigames/sessions` or `/api/v1/sessions`. Six ordinary dotted/percent/slash/
query controls and the existing28 checks pass. Browser19606 independently reproduces
all four in Chromium and WebKit. WHATWG's URL path parsing confirms why percent-encoding
the dots again is not a repair. No cross-origin credential leak or server exploit is alleged.

Fix only the embedded client template and regenerate its consumer: reject those exact
path components before fetch/auth dispatch. Keep query dots and ordinary encoded values,
requests/response decoding, errors, token ownership and no-retry behavior unchanged.
No schema, mounted handler, pin, replay bytes, policy or kernel version changes. Existing
regressions suffice; no separate RFC or repeated source-severing ceremony.

Final57038 passes156 affected Node tests across generated client, operation ports and
Game UI runtime.20239 passes206 complete affected Chromium/WebKit checks.34455 cold
`make test-go vet GO_PACKAGES='./publicapi ./cmd/gen-api' GO_TEST_FLAGS='-count=1'`
passes generator/registry packages/vet;90194 types/Svelte/build pass. `make api-check`
c3b87a regenerates deterministically after staging the intended generated module;
OpenAPI and compatibility-pin bytes are unchanged.

51746 executes309 complete affected cases in all three declared Linux browser engines,
including Firefox, with a cold frozen-lockfile install. Command uses the existing
`compose.browser-test.yml` service/image/anonymous dependency volumes and replaces only
its full-suite command with generated-api/generated-operation-ports file selection;
no CI/config/retry/timeout exclusion changed. Native19606's Firefox failed session connection
at60s (existing RP-116 environment note) and then hung in cleanup; initial sandbox listen
denial ran no tests. Confirmed the exact task-owned runner PID7322 before TERM, retained
its partial/red outcome and left unrelated user browsers alone. Linux success is distinct
evidence, not a rewrite of that failure. No orphan services/volumes were pruned.

These checks exercise the actual generated module, browser Request parsing and controlled
fetch boundaries, not real API mutation/SQL, a full player journey or hosted Actions.
All handles terminal; diff/whitespace checked. Review by: Codex (implementer first filter).
Recorded by: Codex. Exact range starts after `ce61b82a` through this containing commit;
Claude designated review pending, consolidatable with earlier generated-client corrections.
No independent approval of Codex's earlier download/caller batches, draft adoption,
full API/CI/1.0 acceptance, archival, push or deployment. Continue accepted work; the
catalog/dictionary/bodiless/rotation and release-artifact holds remain explicit.

## 2026-10-12 — Consume actual public downloads with the TypeScript verifier

C14 requires downloaded evidence to work with the shipped verifier, not merely two
independent fixture suites. Extend the existing Postgres Exit/board witness while its
real HTTP server is alive: Node loads the actual generated SDK/replay modules through
Vite SSR, fetches manifest/genesis/archive without credentials, checks exact bytes against
the Go witness's storage-checked evidence, binds hashes/engine/URLs, and reverifies the
Company/`founder_advanced` projection. No fake fetch response or synthetic verified row.
Receipt corruption, event removal, a missing first entry and wrong constants each return
the specific non-verified verdict. These invalid cases establish discrimination without
another production-source mutation. Raw archive/storage equality remains separate from
the decoded kernel projection; broader Founder events are not reverified by that projection.

Pinned catalog bytes are supplied from the database bundle over the child process's stdin.
The loader validates their exact hash; it never uses current files as a historical fallback.
This does not complete the public catalog route, third-party download loop, browser-mounted
verification UI, cross-epoch proof or release artifact. The original DOM-ending journey has
explicit cash setup, so it is not relabelled replayable evidence. No runtime, kernel, schema,
generated artifact, copy, balance, configuration or CI workflow bytes change.

71272 `make test-game-ui-composed` completes exit0: both parent integration populations
require the new PASS marker as well as all named Go parents, and the main built-client
journey, actual process-crash/lost-reply recovery, Cosmetic/Clout/Reputation, N5 refusals,
Garden and Tier-2 fixture populations complete. Existing cash/maturity/budget setups keep
their declared limitations.71679 first stopped at my helper's undefined `httpServer` name;
corrected to its actual `server` parameter before the passing run, not hidden as a green
attempt.58443 selected cold Linux-amd64/Postgres Exit test passes, explicitly reporting
that Node is not requested in the ordinary Go lane.92385's initial Make selector lost its
literal dollar anchor and ran zero tests; not evidence, corrected before58443.13647 cold
native gameserver tests and vet pass (native DB exclusions remain; not SQL proof); Node
syntax and diff/whitespace checks pass. Actions already invokes the same root composed
target with Node/dependencies, but no new hosted run or exact hosted-environment claim.

Review by: Codex (implementer first filter). Recorded by: Codex. Exact test/docs range
starts after `492f37ca` through this containing commit; Claude designated review remains
pending and may consolidate with the earlier C14/generated-consumer ranges. No acceptance,
archival, draft adoption, push or deployment. Next: exact owner catalog descriptors/public
catalog composition and remaining generated consumers under their accepted authority;
dictionary/bodiless prerequisites and the full nine-tier release floor remain open.

## 2026-10-13 — Single-schema generated replies enforce their declared media (RP-524)

Under accepted A5/C9/C19, add the existing schema descriptor's media check before
body reading in the embedded generator template and regenerate its client. Raw
and multi-format handling stay unchanged. Missing/plain/gzip media must refuse
even when bytes are JSON; declared JSON and charset controls retain exact status/
body. No HTTP handler, operation/schema/pin, auth, retry, balance, kernel or copy
changes. The earlier verification/download implementation is Codex-owned, so
this is conformance work/first filter, never designated independent review.

Old source `9b48c9c6` fails all12 new single-schema negatives across public verdict
and refresh success/error;8 positive JSON controls pass. Their first rejection
assertions fail, so subsequent spy checks are not claimed old execution. Final
negatives also require no body read and exactly one request. Existing reply doubles
implicitly used text/plain: switch their JSON values to Response.json and give
malformed/read-error fixtures declared JSON media. Keep their original assertions,
error mapping, cancellation, credential rotation and private-diagnostic checks.
Intermediate21264 fails five such doubles; full55052 finds the last Soul double;
both failures are retained, not fixed by allowing undeclared media.

Finished checks: full `make test-client`34759 passes10611, with1012 existing skips
reported, not credit. Native affected Chromium/WebKit20067 passes246; cold declared
Linux browser1712 passes369 in Chromium/Firefox/WebKit with frozen-lockfile install.
Types/Svelte/build33130 pass with zero diagnostics. Cold generator/registry64620
and vet pass; `make api-check` passes after staging the intended generated module,
with OpenAPI/compatibility-pin bytes unchanged. Whole `make test-game-ui-composed`
33334 exits0: actual refresh/schema/SQL parents and public SDK→TypeScript replay,
main built-client/Pitch/process-crash/recovery, Cosmetic/Reputation/Clout, Garden
and Tier2 populations complete. Their existing cash/maturity/budget controls and
DB-supplied catalog limitation remain; not natural pacing or public catalog serving.

All handles terminal; diff checked. These checks do not prove actual AT, full API
acceptance, hosted Actions, clean-host release or the unresolved rights/clock/numeric
contracts. Review by: Codex (implementer first filter). Recorded by: Codex. Entire
new range after `9b48c9c6` through this containing commit needs Claude designated
review, consolidatable with earlier generated-client corrections without erasing
their debt. No draft adoption, archival, push, deployment or1.0 promotion. Continue
remaining accepted construction; catalog/dictionary/bodiless prerequisites stay open.

## 2026-10-13 — Declared response media identity (RP-535)

Routine A5/C9/C19 conformance repair, baseline `15bf18b2`. The generated client
rejects valid mixed-case JSON/gzip headers in all three selection paths. HTTP
[RFC9110 §8.3.1](https://www.rfc-editor.org/rfc/rfc9110.html#section-8.3.1)
defines type/subtype identity as case-insensitive; no new media arm is proposed.
Actual old-client Node20244 fails16/80 cases (64 controls pass); typed-clean
native67965 fails the same32/160 in Chromium/WebKit (128 controls pass).

Normalize only the observed type/subtype in the embedded template, then regenerate
the client. Do not fold parameters, payloads or hashes. Retain strict wrong/absent
media refusal before reading, exact raw bytes/digest checking, status associations,
body-free errors and once-only dispatch. Existing same-status cases now use explicit
expected representation flags instead of a case-sensitive expected-value shortcut.
Nine operation ports also prove ordinary401/429 error classification survives the
equivalent header without extra requests or token reads. No gameplay, handler,
schema/operation/pin, balance, kernel, copy, limiter or CI-workflow changes.

Finished root checks: Node69839 passes217 SDK/runtime/port cases; native3615
passes434 corresponding Chromium/WebKit executions,2.34s. Cold publicapi tests/vet
24806 pass. Types/Svelte/build/boundary88319 pass (217 modules; zero diagnostics);
final types59245 pass after the port supplement. Full `make test-client`95924
passes10668, with1022 explicit existing skips,12.75s. All handles terminal.
`make api-check` passes against the intentionally staged new generated client;
OpenAPI/pin remain byte-unchanged, and a second generation preserves all three
observed hashes. The generated client changes only its normalization/comment lines.

These are SDK/native-Response and injected-port checks, not a new HTTP/Postgres
journey, Linux/Firefox/hosted-CI or release-artifact claim. The Docker lane remains
unavailable without restart authority; prior real-service evidence is not relabelled
as a run on these bytes. The known historical guard and Clout CI failures remain.
Review by: Codex (implementer first filter). Recorded by: Codex. Exact complete
range starts after `15bf18b2` through this containing commit and needs designated
Claude review, consolidatable with the preceding SDK corrections. No acceptance,
archival, draft adoption, publication or nine-tier milestone promotion.

## 2026-10-14 — RP-564 registry-bound typed error classification

Resume at clean `5cd7119f`. Previous goal turn made progress: corrected actual availability
and specified, but did not adopt/implement, the tenant HTTP transaction seam. Continue the
generated boundary under A5/C9/C17 and MA-C7, not the unaccepted successor. The SDK explicitly
JSON-decodes rather than schema-validates success bodies. Its shared operation port additionally
accepts any known category/nonempty detail, so wrong status/family pairs become typed gameplay
refusals; a controlled unknown detail also enters that diagnostic. No real secret leak claimed.

Regression first: generated-operation-ports Node run exits1 with16 failures across four
minigame operations;87 controls pass. Actual Go generator test64524 independently exits2 for
missing registered literal metadata. Emit each existing Response.ExactJSON string array into
generated response metadata using JSON encoding; the port selects the operation/status JSON
descriptor and compares semantic category/detail pairs before constructing MinigameAPIError.
Wrong pairs become the existing body-free transport error, without retry or another auth read.
No manual error table, new server refusal, SDK success guard or raw duplicate/whitespace check.
Garden/Soul descriptors have no literal narrowing and retain their existing guard.

Corrected focused Node187 pass, including cross-operation idempotency controls. Full client
26829 exits0:11197 pass/1040 explicit skips,12.52s. Cold Go/vet92586 exits0 for publicapi0.219s,
Account0.518s and Gameserver0.385s. Typecheck/build27639 exits0, zero Svelte diagnostics;
client boundary check passes. Generator regression checks byte preservation of two literals
including newline/Unicode/HTML characters and no leakage to another operation/status.

Native browser startup20116 first fails sandbox listen with zero executions. Permitted run65075
executes374 selected Chromium/WebKit checks successfully, but Firefox never connects and runs
zero tests; the aggregate is NOT green. Retain its60s session error and180s launch/teardown
timeout with sandbox-extension/framebuffer diagnostics. Send Ctrl-C only after its failed
summary; the same handle eventually terminates130. Process inspection performs no kill or
cleanup of other runners. No timeout/bound/CI change or replacement green claim.

`make api-check` passes against the intentionally staged generated client; OpenAPI and v1
compatibility pin are byte-unchanged. Format/diff and all changed paths checked. These are
generator/native-Response/injected-fetch consumer checks, not current Go HTTP/Postgres,
Linux/Firefox, hosted CI or a release artifact. No Docker restart or publication. Canonical
docs/owning plan/RP-564 reconcile this correction; full API, public Typer, outstanding author
decisions and full nine-tier1.0 remain open. Review by: Codex (implementer first filter).
Recorded by: Codex. Exact entire correction/generated/tests/docs/records range after `5cd7119f`
through this containing commit needs Claude designated review, consolidatable with related
SDK repairs. Next: finish the successor's descriptors/generated guard and metadata compatibility
contracts without pretending this pair check implements them.

## 2026-10-14 — C18 numeric readiness and relevance type conformance (RP-566/RP-567)

Resume at `bd0c9b97`; the preceding answer clarified decisions but made no implementation
progress. Actual C18 readiness probe first: `TestPolicyIntegerDescriptorReadinessProbe`
executes the owner loader with real economy/routes fixtures and the current publicapi integer
validator. Plain1 and maximum-safe plain tokens pass both;1.0,1e0 and maximum-safe exponent
pass the owner but fail the descriptor; fractional and unsafe controls fail both. Root command
exits0 (0.179s). This is a measured representation prerequisite, not authority to narrow the
owner or invent a completed descriptor. Retain reproducible source outside CI at
`probes/relevance-integer-probe.go.txt`; its comments identify the pre-correction coordinate.

The same probe observes Go admitting quoted version/epsilon fields. Under archived Relevance
Harness R12's accepted numeric and closed-string grammar, add four shared refusal cases before
repair: Go cold39370 exits2 with all3 quoted fields admitted; Node exits1 with the array-valued
category axis admitted (21 other tests pass). Parameterize the existing client corpus so every
case executes independently. Add integral-exponent controls, including the exact safe maximum;
do not change existing decimal/fraction/exponent expectations. Preserve raw numeric kind before
Go json.Number conversion and require a real TS axis string; no formula, epsilon value,
scenario helper or historical artifact is edited. Genuine replay-admission change advances the
mirrored kernel189→190 and registers both actual relevance owners as affecting paths, not an
API-only false bump or a path evasion.

Finished shared corpus: all20 cases pass in Go and TS. Actual replay loaders additionally
load freshly hashed complete19-artifact populations: quoted version/epsilon refuse, decimal
version and exponent epsilon retain parsed values, full constants identity and original bytes.
Go70911 executes all4 named controls, exits0 (0.174s); historical/current/candidate replay
checks remain in the cold package population. These are real parser/replay boundaries, not
HTTP/SQL, a product mint, a browser player journey or a full C18 descriptor.

Root package28871 passes relevancepolicy (no standalone tests), replaycatalog0.619s and
kernel0.142s; focused harness policy checks and affected vet pass. Final Node80729 passes162
selected tests; native Chromium/WebKit14564 passes324. Initial native sandbox startup runs
zero tests; permitted localhost execution succeeds. Initial typecheck10366 catches an undefined
artifact assumption in the new test; explicit fixture guard corrects it, with final types/Svelte/
build96472 green (zero diagnostics). Full client74349 passes11250 with1040 existing explicit
skips,7.09s; shell boundary passes. No Firefox/Linux/Docker/DB run is claimed.

Root `make verify-kernel-version`62637 remains RED, terminal exit2 at historical50a3a514;
its checkout/adversarial CI topology checks pass, but the historical failure stops the larger
guard before current-worktree validation. No history exception, rewritten commit, relaxed
assertion, whole-CI or release-readiness claim. All handles terminal; format/diff checked.
Canonical docs/plan and RP-566/RP-567 reconcile this batch. Review by: Codex (implementer first
filter). Recorded by: Codex. Exact entire code/tests/kernel/probe/docs/records range after
`bd0c9b97` through this containing commit needs designated Claude review, consolidatable with
related catalog corrections. No archival, push, deployment, draft adoption or tier promotion.
Next: specify the complete C18 integer representation without redefining historical admission;
all other accepted integration work and outstanding owner/author/review gates remain.

## 2026-10-14 — Buildable integer-token representation proposal (RP-566)

Resume clean `f900340d`; the previous goal turn made progress by correcting actual relevance
field kinds and preserving historical admission. Source-check current schema validation,
cloning, OpenAPI/TS generation, URI scalar admission and recursive pins before proposing any
extension. Formula owner tests explicitly refuse1.0 and preserve signed int64: globally widening
SchemaInteger would be wrong. Official JSON Schema numeric documentation confirms integer is a
mathematical value; OpenAPI's extension mechanism can carry the separate lexical contract.

Draft `rfc/api-integer-token-encoding.md` specifies one opt-in exact-safe mode with explicit
nonnegative safe bounds, unchanged default/URI behavior, original-token validation, immutable
exports, generated annotation and recursive mode-sensitive compatibility. It does not introduce
a new SDK decoder or claim that static TS types validate raw tokens. C18 owner descriptors,
API-only kernel scope and the complete public-byte/consumer loop stay explicit prerequisites.
No new gameplay choice, arbitrary parser budget or production implementation is made.

Before specifying limits, extend the existing shared corpus with six boundary cases. Each
represents exactly1: lengths64/65 and exponents±32/±33 isolate lexical limits from numeric
overflow. Actual Go60892 executes all26 shared cases cold, exits0 (0.286s). Node66967 passes28
tests; native Chromium/WebKit7263 passes56, terminal exit0 (3.19s). No source/parser change, full
client/CI, SQL, Firefox, generated DSL/compatibility or hypothetical-mode execution is claimed.
The earlier seven-case owner-versus-descriptor measurement remains the actual gap evidence.

Diff/paths checked; index, canonical limitation note, owning plan and RP-566 link the proposal.
All handles terminal. Review by: Codex (author first filter). Recorded by: Codex. Exact complete
draft/test-only/docs/records range after `f900340d` through this containing commit needs Claude
designated review and normative adoption; it does not approve the preceding behavior range.
No archival, kernel bump, push, deployment or milestone promotion. Next: continue accepted
construction while this recommendation and the other named authority gates await resolution.

## 2026-10-14 — Refuse malformed UTF-8 in generated schema replies (RP-568)

Resume clean `5872b66e`; previous goal turn completes the browser execution-organization batch.
Under accepted A5/C9/C19 JSON decoding, the generated client's `Response.text()` silently
repairs malformed UTF-8 inside JSON strings. Execute36 old-client refusals (nine byte faults
across public/private success/error operations): all fail because corrupt replies are admitted;
92 existing/positive/raw controls pass. Record RP-568 immediately. JSON network encoding and
BOM tolerance are checked against RFC8259§8.1; do not implement the unadopted Typer draft's
stricter BOM/DTO policy.

Change only the generator template and regenerated transport: read bytes outside the error
sanitizer, fatal UTF-8 decode then parse JSON inside it. Invalid encoding uses the existing
body-free status diagnostic; body-read failures still propagate unchanged. Raw evidence/hash
handling remains byte-identical. Controls preserve valid2/3/4-byte characters, Unicode maximum,
literal U+FFFD, existing JSON-escaped surrogate handling and optional BOM. Nine deliberately
corrupt raw bodies still return exact digest-verified bytes; this is identity, not valid genesis.

Final Node88456 passes289 affected SDK/operation-port/runtime checks; actual Chromium/WebKit
83713 passes578. Cold Go15864 passes `./publicapi ./cmd/gen-api`; affected vet passes.
Final typecheck89673 has zero errors/warnings; build41820 succeeds. Initial helper annotation
`Uint8Array<ArrayBufferLike>` fails two Response BodyInit type checks despite runtime passes;
correct it to its actual ArrayBuffer-backed return and execute final checks again. An earlier
incorrect CLIENT_BIN override fails before Vitest starts and supplies no test evidence.
Stage the intentionally changed generated transport, then `make api-check` re-emits identical
candidate bytes; OpenAPI/compatibility pins remain unchanged against HEAD. Diff checks pass.

Retained old failure `.cache/api-utf8-old-5872b66e.log` SHA256
`2d2d77c327fd92e0259904ca913afe715b993620b6b9505adb7148a4dfedb10a`;
final native `.cache/api-utf8-native-final-5872b66e.log` SHA256
`dd536d8996ce9d744cb1709c51912421c5352a1a62cd8dc1e0e359d7a05656f2`.
No new evidence framework, gameplay/schema/kernel/balance/copy, retries, routes or workflow
change. Tests exercise actual Response decoding with injected fetchers, not a composed server,
Postgres, Firefox/Linux, hosted CI or full public verification/download loop.

Review by: Codex (implementer first filter). Recorded by: Codex. Complete implementation,
test/docs/records range after `5872b66e` through this containing commit needs designated Claude
review, consolidatable with related API corrections. RP-568 is locally corrected, not designated
approved. Existing draft/owner/author/review/CI and full nine-tier1.0 gates remain. No archival,
push or deployment. Next: accepted API construction and outstanding consumer/owner contracts;
do not repeat these completed focused checks as progress or infer proposal adoption.

## 2026-10-14 — Bind typed operation errors to the existing detail enum (RP-569)

Resume clean `e56da0ec`; previous turn completes the encoding correction. Read the accepted
Garage/minigame plans, then trace their shared generated-operation boundary. Docker21312 still
returns no bytes and terminal28 after5002ms; no daemon restart or substitute DB. The adapter
checks category but accepts any nonempty detail when ExactJSON metadata is absent. Actual old
port regressions admit a fictional unknown detail for Garden and four Soul operations (five
failures/111 passes); the existing four minigame literal bindings reject it. File RP-569.

Under API A4/A5/C9/C17, generate detail metadata from the existing descriptor, not a handwritten
consumer table; the adapter checks it before constructing MinigameAPIError. Four old-generator
checks fail: bounded enum, changed owner values, enum-free string and Unicode enum. Corrected
metadata tracks all four cases; enum-free is null, not an empty allowed set. No descriptor,
server handler/error bytes, status, retry/recovery policy, schema/pin, kernel or balance change.
Existing minigame per-operation/status literal checks remain. This does not close other
operations' stronger pair-conformance gaps or provide runtime successful-DTO validation.

Cold Go13165 passes `./publicapi ./cmd/gen-api`; affected vet passes. Node74290 passes312 SDK,
operation/runtime and surface checks. Native42823 terminates0 with702 checks/14 files across
Chromium/WebKit, including mounted Soul Recovery and Garden refresh populations. Types82047
have zero errors/warnings; build60248 succeeds. OpenAPI and compatibility pins are unchanged.
Stage the new generated metadata and run `make api-check` against the candidate index to prove
deterministic regeneration. These are controlled Response/port/mounted-runtime tests, not
real Go HTTP/Postgres or public-server recovery, Firefox/Linux or hosted CI evidence.

Ignored raw evidence: `.cache/port-detail-old-e56da0ec.log` SHA256
`119e52190002e1d4a2b9a44b8d9fed436c8ebc2e7df44f85668924faa1c0a52e`;
`.cache/port-detail-generator-old-e56da0ec.log` SHA256
`0a77fcdae9b520e82254f256cef4a13c7a006b28156e32744998d6371d1358f7`;
`.cache/port-detail-native-e56da0ec.log` SHA256
`f69817d84754e9cec1df2b3c1ca9078d8e2326e55a235ca54d9bc91894eea5d1`.

Review by: Codex (implementer first filter). Recorded by: Codex. Whole implementation/test/
docs/records range after `e56da0ec` through this containing commit needs designated Claude
review, consolidatable with the API corrections. RP-569 is locally corrected, not approved.
The marker is fictional; no real secret leak claim. Existing owner/author, Docker, CI, review,
feature and full nine-tier1.0 gates remain. No archival, push or deployment. Next: remaining
accepted construction and real integration when its declared services are usable, not reruns
of these completed focused populations or inferred adoption of the pending drafts.

## 2026-10-14 — Preserve error enum metadata through named references (RP-570)

Resume clean `180f25e2`; previous turn completes the shared error-detail repair. Re-read Combat's
accepted parent, child ruling, implemented records and design references: RP-072's effect/input
tables still need author closure and DU-C1 forbids battle construction before them. No guessed
mechanics or new battle work. During the generator follow-up, valid C17 reference chains reveal
lost category/detail exports, including an aliased error envelope.

First fixture is unsorted and correctly fails registry admission; exclude it as generator evidence.
Sort the declared schemas and rerun unchanged production: both inline/aliased envelope cases
admit the declared error, reject the unknown detail through actual ValidateResponse, then fail
the generated metadata oracle. File RP-570. Resolve the envelope and fields through the registry's
already-validated acyclic reference graph; do not rewrite aliases, schemas, pins or caller input.

Final cold Go78452 passes `./publicapi ./cmd/gen-api` with all new/existing tests; affected vet
passes. `make api-check`15102 terminates0: OpenAPI, compatibility pins and production generated
TypeScript are byte-identical to HEAD. Existing inline/unbounded/Unicode metadata controls pass.
No changed shipped TS/browser code or DB behavior: no unrelated client/browser/composed rerun.
Formatting/diff checked. Valid old regression output `.cache/error-enum-alias-old-valid-180f25e2.log`
SHA256 `a63d0abd505ba52a6aed934cb77656700731bad4c2548c56516b66801dd51f81`.

Review by: Codex (implementer first filter). Recorded by: Codex. Whole generator/test/docs/records
range after `180f25e2` through this containing commit needs designated Claude review, consolidatable
with RP-569. RP-570 is locally corrected, not a current default-handler defect or whole-DSL/runtime
DTO validation claim. No owner/draft adoption, kernel/balance, archival, push or deployment.
Combat, other authored/owner contracts, actual Linux/SQL/hosted, review and nine-tier1.0 obligations
remain. Next: close the named foundational contracts and continue accepted construction, not
repeat these finished generator checks as progress.

## 2026-10-14 — Construct the Achievements artifact-owner schema with a real admission correction

Resume clean `10ca8fbf`; previous turn completes the Headcount ID repair. Read API C18,
the archived Achievements CA1/C6 contract, both owners and existing category-export pattern.
Go root/row struct decoding admits16 case aliases; TypeScript's exact objects already reject
them. Actual cold old-loader cases fail16/16, not a hypothetical metadata-only version premise.
RP-572 names the discrepancy. Preserve raw row bytes long enough to check exact member names
before struct decoding; reuse the owner's existing key semantics. Escaped canonical names and
last-value duplicate behavior remain; TS numeric-token behavior is explicitly unchanged.

`AchievementSchemas` exports seven fresh named definitions: root/row, proof oneOf and four
finite condition-depth definitions. All five condition and three proof arms are closed; no
recursive/open JSON escape. The current catalog and two checked-in stored copies validate
byte-unchanged. Existing fixture plus run/career counter, Exit, burn, possession, bounded allOf,
sixteen malformed shapes, depth4/5 boundary, semantic-copy refusal and fresh-export controls
pass. This is structural metadata; loader bindings, order, cardinality and total-node bounds
remain mandatory. No union/HTTP registration or new content is claimed.

Kernel0.3.190→0.3.191 records the actual Go admission change in the same range as the owner
export; no false API-only bump or namespace escape. Final cold full achievements/publicapi/kernel
packages pass; affected production selection passes. The initial combined selector also
matches existing SQL tests without a DB, so it is not integration evidence. Separate verbose
36233 executes all nine pure TestFoundation top-level cases (including settlement/next-catalog,
hook order, debit/event failure paths, offline ledger and carry/clone), no skips, terminal0.
Affected vet and achievement source/transitive/copy boundaries pass. Node90929 passes158
Achievement/replay checks; native88642 passes316 in Chromium/WebKit, terminal0/2.45s. Types90529
have zero errors/warnings, production client build and deterministic `make api-check` pass.
Generated API/pins, current catalog and materialized stored bytes remain unchanged.

Full kernel guard46946 remains RED at historical50a3a514; it does not reach a current-tree green
verdict. Keep RP-131 and known production/hosted failures open. No Docker/SQL/Firefox/Linux,
clean-host, public catalog loop or complete API/1.0 acceptance claim. Raw evidence stays ignored:
old `.cache/achievements-alias-old-10ca8fbf.log` SHA256
`765fd8dab589d4dfc60a346fec9d0d7492f4c0979b6c8907eacda9a1357ad0a9`;
native `.cache/achievements-owner-native-10ca8fbf.log` SHA256
`ae33fb28a3a8d3a540829d8aff4f30415043b60f687739f1445e3ebe7a474da2`;
pure production `.cache/achievements-owner-foundations-10ca8fbf.log` SHA256
`1d35890d6e07e2e5364901855f29281fd5c740005886ac16125d2c77e41faeae`.

Review by: Codex (implementer first filter). Recorded by: Codex. Full owner-code/schema/tests/
version/docs/records range after `10ca8fbf` through this containing commit needs designated
Claude review, consolidatable with API/owner corrections. RP-572 is locally corrected, not
approved. Two of nineteen pinned-family descriptors exist; seventeen and complete C18/third-party
verification remain. No archival, push, deployment or owner/draft adoption. Next: other owner
descriptors within lawful scope, then the complete union/HTTP loop; do not ratify the unadopted
integer representation mode or substitute metadata for semantic loading/release proof.

## 2026-10-14 — Construct the Meters owner descriptor and repair existing admission contracts

Resume `c7c4e561` with the Meters batch in progress. Under API C18, Meters M1/C6 and Clout
CV3, export `MeterSchemas` from the owner: closed root/reseed/row/band/input/decay shapes,
both input arms, bounded signed deltas, null/active decay, registered meter IDs and slots.
Current catalog, two stored artifacts and existing fixtures validate byte-unchanged. Thirty-six
wrong-shape controls refuse in descriptor and loader; all nine slots × four signed boundary
deltas × two decay forms pass. Order/cardinality/source uniqueness/reseed relationships remain
mandatory loader checks, demonstrated by four structurally valid but semantically invalid cases.
Fresh-export and escaped-name/last-value controls pass; no union/public HTTP activation or mint.

Actual old Go loader admits38 case-alias controls; actual old TS refuses the registered axis_stack
candidate (one failure/seven passes), while Go accepts it. RP-573/RP-574 name these defects.
Check original object keys before typed Go decoding, including reseed/bands/decay; restore TS's
existing slot domain and use a type-only canonical slot alias. No runtime economy dependency.
Exact final reseed objects must be complete, not inherit missing fields from an earlier duplicate;
no global duplicate-name rejection is introduced. Existing numeric-token behavior, balances,
decay/reseed arithmetic and replay encodings remain. Genuine kernel0.3.191→0.3.192 accompanies
these admission corrections, not a metadata-only bump.

Final cold full meters/publicapi/kernel packages pass; affected vet, Meters source/transitive
boundary, typecheck (zero errors/warnings), production build and unchanged `make api-check` pass.
Node78551 passes157; actual Chromium/WebKit28004 passes314, terminal0. Separate verbose production
13923 executes all nine pure TestFoundation cases and relevant negative subcases, no skips.
These prove owner/replay and pure composed-hook behavior, not Postgres/HTTP player workflows,
Firefox/Linux, hosted CI, full C18 or1.0. Historical kernel guard remains known-red at50a3a514
from the preceding checkpoint; no rerun or current-tree/full-CI green claim. Docker and other
owner/author/review/release gates remain unchanged.

Ignored raw old-Go `.cache/meters-alias-old-c7c4e561.log` SHA256
`59c12d6ec3f9da3d594b882e8582627f3302d0b2e846faa09020dc89c1a3d524`;
old-TS `.cache/meters-slot-old-c7c4e561.log` SHA256
`9f9634ba51da8f9ad3190733b4054210e0fcecf4a8a3c38fb7ee1d59afe57c8d`;
final-native `.cache/meters-owner-native-final-c7c4e561.log` SHA256
`dda91cd143a0bf8a8e70066b85111da08c690f391eb44ca3e4e8303156d3f8cc`;
pure-production `.cache/meters-owner-foundations-c7c4e561.log` SHA256
`4ac4e6815bab134c7d77f84924f8d8a27a0ee6659315934a99b66dd7c292ac18`.

Review by: Codex (implementer first filter). Recorded by: Codex. Complete source/schema/tests/
version/docs/records range after `c7c4e561` through this containing commit needs designated Claude
review, consolidatable with related C18 corrections. Three of nineteen pinned-family descriptors
exist; sixteen, complete catalog union and public re-verification remain unfinished. No archival,
push/deployment, new content or draft/owner adoption. Continue accepted construction without
silently choosing Headcount's conflicting hiring/output models or treating proposal delivery as
owner approval. Those decisions do not block this completed Meters batch.

## 2026-10-14 — Construct Opportunities' owner descriptor with the A9 admission repair (RP-575)

Resume clean `11efb691`; preceding Meters construction is concrete progress. C18's remaining
owners require lawful scope, not a metadata-only kernel bump. Read the complete archived
Active-Play contract and current readers; A9 already requires exact keys/closed arms and safe
integer durations. Actual old Go admits38 root/policy/effect case aliases, four wrong-arm null
members and three unsafe duration candidates; unchanged TS rejects them (five tests pass).
File RP-575, check original object shapes before struct decoding, and bound all three duration
arms by the existing shared maximum. No mechanics, cadence, payouts, buffs, click clamp or data
retune. Genuine kernel0.3.192→0.3.193 accompanies the owner addition.

`ActivePlaySchemas` exports closed root/policy/four-effect shapes, existing sampler/substream
literals, canonical Decimal strings and numeric/ID bounds. Current and two stored catalogs
validate byte-unchanged and load with their corresponding Economy catalogs. The shared fixture
contains all four arms; duration1/S controls, twenty-one structural refusals, old discrepancy
populations and freshness/escaped-name/duplicate controls pass. Four semantic counterexamples
still require the loader (bindings, factor relationships, nonempty/order). No union/public HTTP
activation, fixture rewritten into adopted content, open JSON or metadata-only version signal.

Final cold activeplay/publicapi/kernel packages and affected vet pass. Node28453 passes146 and
actual Chromium/WebKit61263 passes292 across active-play, opportunity-arm and replay files.
Typecheck41918 has zero errors/warnings; production build and unchanged `make api-check` pass.
Cold replaycatalog59130 also passes the affected catalog-consumer package; no SQL evidence is
inferred from its package exit.
Verbose pure production executes five existing top-level active-play cases and all five Lucky
boundary vectors, no skips. The first broad selector also selects the SQL event test, which
skips without Docker/Postgres; it supplies no DB evidence. An initially unquoted parenthesized
Make regex fails at the shell before Go runs; quote it and execute the final pure selector.
No native Firefox/Linux/SQL, hosted, complete API/CI or release claim. Historical RP-131 guard
remains known-red from prior execution; no rerun/current-tree green claim or history rewrite.

Ignored raw old `.cache/opportunities-admission-old-11efb691.log` SHA256
`79516685e02aa30e037348dda619d2d7654c848d4846404ad70b0761273fcedc`;
native `.cache/opportunities-owner-native-11efb691.log` SHA256
`eb7ce073a09af1777931ec02466141f607c86c25e55ce6c0dc4080bc30a63266`;
final pure `.cache/opportunities-owner-pure-production-11efb691.log` SHA256
`83e1a68863a559156d487ebf82cc188c8c25dbcc87c193c54234aea67d6decfe`.

Review by: Codex (implementer first filter). Recorded by: Codex. Complete owner/schema/tests/
version/docs/records range after `11efb691` through this containing commit needs designated
Claude review, consolidatable with related C18 corrections. Four of nineteen pinned-family
descriptors exist; fifteen, full response union/catalog HTTP and public re-verification remain.
No archival/push/deployment, owner/draft adoption or full nine-tier1.0 promotion. Continue accepted
construction; live intent/dictionary, bodiless replies and other authored contracts remain held.

## 2026-10-14 — Construct Doctrine's owner schema and reject implicit catalog values (RP-576)

Resume clean `e670cd1a`; Opportunities is the previous concrete progress. Read the complete
archived Doctrine D1–D10 contract, current owner/TS parser and C18 plan. The actual old Go reader
admits12 case aliases, absent/null source_tier as zero on a matching boundary, and a final null
schema_version inheriting the earlier value. TS rejects all those candidates; explicit zero
loads in both. First unquoted Make alternation is a shell/tool selection failure, not evidence;
execute the entire cold package to retain the valid old failures. File RP-576.

Preserve original transition object bytes for exact names before typed decoding; pointer-backed
numeric fields require explicit non-null values. `DoctrineSchemas` exports D1's closed root/row
grammar, mechanical IDs and0..8 bounds. Current and two stored catalogs validate/load and bind
to their corresponding Routes artifacts byte-unchanged. Sixteen structural refusals, zero/eight
boundaries, four schema-valid semantic refusals, fresh exports, escaped keys and complete
last-value controls pass. No doctrine effects/choices, Compute Credit arithmetic, mint, replay
encoding, route policy or global duplicate rejection is introduced. Genuine kernel193→194
records the actual admission correction, not a metadata-only version signal.

Final cold doctrine/publicapi/kernel/replaycatalog packages29115 pass; affected vet passes.
The initial new schema test file has an unused import, so that compilation/vet run is invalid;
remove it and execute final checks. Node77450 passes142; Chromium/WebKit61341 passes284.
Verbose production97041 executes four top-level cases: intent grammar, pick/spend transitions,
partition/offline burst bounds and Permits/Doctrine composition; no skips. Types65528 report
zero errors/warnings, production build and unchanged `make api-check` pass. No SQL/player
HTTP, native Firefox/Linux, hosted CI, complete C18 or release claim. Historical RP-131 guard
remains known-red from previous execution; no rerun/current-tree green verdict.

Canonical Doctrine docs also carried RP-029's stale pre-mint premise. Current epoch8 manifest
pins doctrines; the existing First Content record and actual catalog/Route pair show epoch6
minted the identity-only row with its gate/activation chain. Correct docs, not frozen historical
RFC text or content. Consolidate the plan's three duplicated recent C18 narratives into one
current5/19 checkpoint; detailed evidence/history remain in this append-only log.

Ignored raw old `.cache/doctrines-admission-old-final-e670cd1a.log` SHA256
`0cc074f22ceb62215a416a282ff747394a9023b6c91c6f6f08b00de8790a2758`;
native `.cache/doctrines-owner-native-e670cd1a.log` SHA256
`615ac4f25cd0476215996a7942d72a6a065b21eeb9d5ff24565e27d628548ca2`;
production `.cache/doctrines-owner-production-e670cd1a.log` SHA256
`d6e2a03e0fc77ba08bcbe6d460c481e85867bceaec406d43fcbf8ce5790a5561`.

Review by: Codex (implementer first filter). Recorded by: Codex. Complete owner/schema/tests/
version/docs/records range after `e670cd1a` through this containing commit needs designated
Claude review, consolidatable with related C18 corrections. Five of nineteen pinned-family
descriptors exist; fourteen, the complete union/HTTP and public re-verification remain. No
archival/push/deployment, owner/draft adoption or full nine-tier1.0 promotion. Continue accepted
construction without inventing missing contracts or calling these scoped checks full CI.

## 2026-10-14 — Construct Fiscal's owner descriptor and enforce exact final policies (RP-577)

Resume clean `108706a6`; Doctrine is preceding concrete progress. Read the full archived Fiscal
F5/F9 contract and both current readers. Actual old Go admits26 case aliases across root,
policies and both row families, six final partial/null policy objects and one final null reason
key. Unchanged TS rejects the same candidates (six tests pass). File RP-577. Check final original
object shapes and non-null members before typed decoding can fold names or inherit earlier fields.
Complete last-value controls and escaped canonical keys remain valid; no global duplicate-member
rejection. Genuine kernel0.3.194→0.3.195 accompanies the real admission correction.

`FiscalSchemas` exports F9's closed root/policy/row shapes, safe-integer/PPM bounds, IDs and slots
derived from multiplier.Order. Current catalog, two stored artifacts and the shared fixture pass
byte-unchanged; every stored catalog also loads with its corresponding Economy artifact. Twenty-eight
structural refusals, probability/credit endpoints, all nine descriptor slots and matched/mismatched
pinned declarations, five schema-valid semantic refusals and fresh-export/member controls pass.
Clock/credit relationships, sequence horizon, row cardinality/order and declarations remain loader
checks. No thresholds, harvest risk, costs, factors, frozen contributions, state/replay encoding,
mint, public union or HTTP activation changes. Clarify canonical docs' ambiguous “duplicate” claim
as duplicate row IDs, not a falsely claimed global JSON-member policy.

Final cold fiscal/publicapi/kernel/replaycatalog19542 passes; affected vet passes. Node85799
passes144; actual Chromium/WebKit93389 passes288. Verbose production12642 executes five pure
top-level Fiscal cases and the Founder-clock offset controls, no skips: activation, frozen
contributions, provisional sweep rollback, clock regression and independent save axes. Types72929
have zero errors/warnings, production build and byte-unchanged `make api-check` pass. Revalidate
Docker with one bounded read-only ping: curl42258 returns no bytes, terminal28 after5010ms.
This proves only an unresponsive endpoint, not a stopped daemon; no restart/substitute DB or SQL
claim. No native Firefox/Linux/hosted, complete API/CI or release proof. Historical RP-131 guard
remains known-red from preceding execution; no rerun/current-tree green verdict.

Ignored old `.cache/fiscal-admission-old-108706a6.log` SHA256
`eddc0ad7c1aea38885d338bb77ff0f6c830f32db5a945c70cef98a2f845f6b4b`;
native `.cache/fiscal-owner-native-108706a6.log` SHA256
`1fe97c083e072b87f8cf0a8d1d95c63877742ddbad16d959441536625fe77e03`;
production `.cache/fiscal-owner-production-108706a6.log` SHA256
`d357fcdac372933ebc419363b6d64f22813d43d1dd713b318dd23051c31223b9`.

Review by: Codex (implementer first filter). Recorded by: Codex. Complete source/schema/tests/
version/docs/records range after `108706a6` through this containing commit needs designated Claude
review, consolidatable with related C18 corrections. Six of nineteen pinned-family descriptors
exist; thirteen, full response union/catalog HTTP and public re-verification remain. No archival,
push/deployment, new content or owner/draft adoption. Continue accepted construction; review,
services, author/owner contracts, full CI and the entire nine-tier1.0 release remain open.

## 2026-10-14 — Construct Soul's owner descriptor and refuse null defaults (RP-578)

Resume clean `d87bffdc`; Fiscal is preceding concrete progress. Read the accepted SB10/SB17/
SB27 and SR-C2 catalog/enum/presentation contracts and actual Go/TS readers. Old Go admits
eight null policy/band/flag candidates, including initial Soul silently becoming zero; TS refuses
all eight. Explicit zero/false remains legal. File RP-578 and add non-null checking to the
catalog-only exact-key helper (all call sites checked). Existing uniqueJSONKeys stays unchanged.
Genuine kernel0.3.195→0.3.196 records this correction; no arithmetic/cadence/content retune.

`SoulSchemas` exports closed v1 policy/band/debit/activity/ending shapes, exact enums and numeric
domains. Current and two stored artifacts plus recovery fixture validate byte-unchanged and load
against the tracked generated copy registry. Twenty-five shape/refusal controls, all four owner
kinds × three toy kinds, six schema-valid semantic failures, empty production-versus-fixture
rules, fresh exports and escaped/duplicate controls pass. Descriptor map validation admits a
duplicate control that Soul's original-byte uniqueness correctly refuses; shared artifact
validator semantics are not changed. Bindings, partition, ordering, context and value relations
still require the loader. No source mint, catalog union or public HTTP activation.

Final cold soul/publicapi/kernel/replaycatalog26554 and affected vet pass. Node18174 passes149;
actual Chromium/WebKit37966 passes298 catalog/replay/scheduler cases, plus mounted Recovery80346
passes52. Verbose production92509 executes all three activation/suppression/coordinate-refusal
cases, no skips. Types1538 have zero errors/warnings; build and byte-unchanged `make api-check`
pass. No SQL/real-server recovery, native Firefox/Linux/hosted, accessibility or whole API/CI
claim. Docker's preceding bounded ping remains unresponsive; no restart/host-DB substitution.
Historical RP-131 guard remains known-red; no rerun/current-tree green verdict.

Soul overview also had a stale “no mounted surface” claim. Actual GameUIApp host mounts
SoulRecoverySurface; the component renders RecoveryToy and tracked disclosure keys. Reconcile
canonical docs to that construction, keeping the successor's player/review/AT gates separate.
No frozen RFC, player-facing copy, schema/body ruling or acceptance status is changed.

Ignored old `.cache/soul-null-old-d87bffdc.log` SHA256
`89df364849b28a99d423aff1fd92d741768eba539e98fc0805c68e728708ed2e`;
native `.cache/soul-owner-native-d87bffdc.log` SHA256
`8fff577b7ad97bafaee37fc42faeea4092fbe21b5ac815bab018b5e1589f56e6`;
mounted `.cache/soul-owner-mounted-d87bffdc.log` SHA256
`ed20547ad34bf2ed4aa17e3f4957bff5311ccaee3c1cab572961874fa7159862`;
production `.cache/soul-owner-production-d87bffdc.log` SHA256
`ca1438d9b9f0fdf10a08d1b68524b748dfaa1e6718c5ee00271254c74a247916`.

Review by: Codex (implementer first filter). Recorded by: Codex. Complete source/schema/tests/
version/docs/records range after `d87bffdc` through this containing commit needs designated Claude
review, consolidatable with related C18 corrections. Seven of nineteen pinned-family descriptors
exist; twelve, complete response union/catalog HTTP and public re-verification remain. No archival,
push/deployment, new content or owner/draft adoption. Continue accepted construction; services,
review/author/owner, full CI and the entire nine-tier1.0 release remain open.

## 2026-10-14 — Construct Curriculum's owner descriptor and bind the ruled starter branches (RP-579)

Resume clean `4c068540`; Soul is preceding concrete progress. Read the current curriculum
owner/TS parser and ruled First-Hour Payoff mechanism resolving T01-C31/C35/C36. Starter TYPE
is fixed by ending; the generic starter union reused by Reputation is a different authority.
Actual old Go admits37 case aliases, six missing/null/wrong-arm default fields and three valid
generic starters swapped onto the wrong branches. TS refuses all46 candidates; explicit zeros
remain valid. File RP-579, check original exact non-null shapes and enforce branch/starter kinds
in the curriculum-only validator. Generic ValidateStarter stays byte-unchanged. Genuine kernel
0.3.196→0.3.197 records the correction; no payoff/copy/mint or terminal-behavior policy change.

`CurriculumSchemas` exports the closed root/failure and three branch/starter pairs. Current,
epoch8 stored and candidate-v2 bytes validate/load with their corresponding Economy/Routes and
tracked copy declarations. Epoch7 has no curriculum artifact; do not fabricate one. Eighteen
shape refusals, all three wrong-pair refusals, five schema-valid semantic failures, all generic
starters, valid A/B/R/S/G retunes and fresh/escaped/complete-duplicate controls pass. Loader
binding, positive Decimal relations and branch order/cardinality stay mandatory. No public
catalog union/HTTP activation or hardcoded candidate balance literals.

Final cold curriculum/publicapi/kernel/replaycatalog95781 and affected vet pass. Paired-route
fixture first uses the wire field name as a Go field and fails compilation/vet; inspect Gate.ID,
correct it and execute final checks. Node31880 passes142; Chromium/WebKit94622 passes284.
Verbose production63019 executes six pure/replay cases with all branch/gate/session controls:
due cross-gate, pre-gate refusal, neutral Reputation plus all starters, accrued branch preview,
legacy Reputation starter replay and generic starter effects; no skips. Types24775 report zero
errors/warnings, build82956 and byte-unchanged `make api-check` pass. No SQL/composed player,
natural pacing, native Firefox/Linux/hosted or whole API/CI claim. Docker remains unavailable
under the preceding verified ping; no restart/substitute DB. Historical RP-131 remains known-red,
not a current-tree green guard verdict. Balance, stored/testdata and generated API files unchanged.

Ignored old `.cache/curriculum-admission-old-4c068540.log` SHA256
`adf13340f0f48bf73da340ce18e2a8fecd4d9b2aef403544f46f88fce7337e82`;
native `.cache/curriculum-owner-native-4c068540.log` SHA256
`e609575b919ac8ea9435e92e3f31bca2e7f2aa7d2a492ec82574611ad79b856c`;
production `.cache/curriculum-owner-production-4c068540.log` SHA256
`ff84426893400c35b7c0abc675eb2d9efdc3d992e36da704f75138c13abddba6`.

Review by: Codex (implementer first filter). Recorded by: Codex. Whole owner/schema/tests/
version/docs/records range after `4c068540` through this containing commit needs designated
Claude review, consolidatable with C18 corrections. Eight of nineteen pinned-family descriptors
exist; eleven, complete response union/catalog HTTP and public re-verification remain. No archival,
push/deployment, owner/draft adoption or full nine-tier1.0 promotion. Continue accepted construction;
service, review/author/owner, full CI, content/pacing and the entire release remain open.

## 2026-10-14 — Construct Prestige's owner policy schema and align admission/expiry bounds (RP-580)

Resume clean `65ee6a8e`; Curriculum is preceding concrete progress. Read the complete accepted
Prestige contract and actual Go/TS policy/expiry paths. Old Go admits35 malformed candidates:
eleven case aliases, eight absent/null zero fields, five null payouts, ten null spawn entries
and one final incomplete modifier map merged with earlier keys. TS already refuses those but
admits the first duration above Go's existing int64-nanosecond representation. Actual old TS
duration regression fails (15 other checks pass); Go's maximum/first-invalid control passes.
File RP-580. Enforce exact original non-null shapes; name the unchanged Go duration bound and
mirror its arithmetic in TS. No convenience limit, raised ceiling or new gameplay policy.
Genuine kernel0.3.197→0.3.198 accompanies the admission correction.

`PrestigeSchemas` exports closed policy fields, a finite five-key payout object, numeric bounds
and existing duration representation. Current and two stored policies validate/load byte-
unchanged. Eighteen structural refusals, threshold/table semantic counterexamples, maximum-
domain/zero/retune and fresh/escaped/complete-duplicate controls pass. Positive state threshold,
table cardinality and composition's Economy resource binding remain mandatory. No formula,
payout/threshold data, offer/preview/expiry mechanics, Advisor activation, mint or public union/HTTP
change. Generated API/pins and balance/testdata bytes stay unchanged.

Final cold prestige/publicapi/kernel/replaycatalog13226 and affected vet pass. Node35480
passes154; actual Chromium/WebKit78842 passes308 including shared Prestige vectors/replay.
Verbose production9575 executes the offer replay corpus and copied-evidence refusals with
all payout/promise controls, no skips. Types53653 report zero errors/warnings, build and
byte-unchanged `make api-check` pass. No SQL/composed player, native Firefox/Linux/hosted,
full API/CI or release claim. Docker remains unresponsive under the preceding verified ping,
no restart/substitute DB; historical RP-131 guard remains known-red, not current-tree green.

Ignored old-Go `.cache/prestige-admission-old-65ee6a8e.log` SHA256
`2e64b27d2fb24784b446da7eefcd28e8ccae69a558ee9ebc055c6c0bf5adb424`;
old-TS `.cache/prestige-duration-old-65ee6a8e.log` SHA256
`6d956ff67c0077b0bb3e9a91aaacb9671388bb620add589390abb56819ca78bc`;
native `.cache/prestige-owner-native-65ee6a8e.log` SHA256
`f11789d33f05daabc73b7eec3c437fc47cf4369d8739053a9572ec999f4036b3`;
production `.cache/prestige-owner-production-65ee6a8e.log` SHA256
`60236362302cae4fa169555d38af80706645a80272c3767cb0fcdbc755ef79a5`.

Review by: Codex (implementer first filter). Recorded by: Codex. Complete code/schema/tests/
version/docs/records range after `65ee6a8e` through this containing commit needs designated
Claude review, consolidatable with C18 corrections. Nine of nineteen pinned-family descriptors
exist; ten, complete union/catalog HTTP and public re-verification remain. No archival/push/
deployment, new content or owner/draft adoption. Continue accepted construction; services,
review/author/owner, full CI, content/pacing and full nine-tier1.0 remain open.

## 2026-10-14 — Pitch catalog descriptor and required-price admission (RP-581)

Resume clean `307725e4`. Under API C17/C18 and Pitch TP-C13/TP-C21/TP-C22, export
`PitchCatalogV1`: exact root/policy/card/hack/funding objects, four closed effect arms,
existing literal policy and safe-integer/Decimal domains. Current epoch8, fixture and two
stored artifacts validate/load byte-unchanged. Twenty-one structural refusals and nine
semantic counterexamples demonstrate the boundary: row count/order, positive values,
copy declarations/bindings and chain partners still require owner validation. Escaped
canonical members pass; raw duplicates still fail Pitch's existing unique-member gate,
not the shared map-based descriptor validator. Fresh exports cannot mutate the owner.

The actual Go reader admitted all eight null hack prices as zero; the old-source regression
fails eight refusal arms while eight explicit-zero controls pass. TS's sixteen corresponding
checks already pass. Require a non-null price before typed admission; preserve legal zero.
Final regressions also exercise tenant creation with each malformed/legal candidate.
Genuine kernel0.3.198→0.3.199; no content, shop/scoring/payout or duplicate-policy change.
Correct docs' stale pre-mint claim against the current epoch manifest. Complete schema
union/HTTP serving and public re-verification remain unfinished, now ten of nineteen owners.

Final cold pitch/publicapi/kernel/replaycatalog/decimal packages and affected vet (including
production) pass. Node Pitch/economy169 and numeric vectors6304; native Chromium/WebKit338
and numeric vectors12608 pass. Pure production `TestPitchBundleResolvesImmutableTenantContent`
executes, with no skips. Types report zero errors/warnings; build and byte-unchanged
`make api-check` pass. Initial fresh-export test dereferenced a bound on the wrong field and
panicked; replace positional assumption with named schema_version selection and rerun the
finished packages successfully. No DB/composed player, Firefox/Linux/hosted or full CI claim;
known historical guard/Clout failures are not waived. Balance/testdata/generated API bytes
remain unchanged. No new owner decision, draft adoption, publication or release promotion.

Review by: Codex (implementer first filter). Recorded by: Codex. The complete source/schema/
tests/version/docs/records range after `307725e4` through this containing commit requires
designated Claude review, consolidatable with the related C18 batches. Continue the nine
remaining owners and full catalog integration within accepted authority; no archival claim.

## 2026-10-14 — Pet care catalog schemas and required numeric admission (RP-582)

Resume clean `293a05e4`. Pet Care C13/C17 and Soul SB22 already define the complete policy
and historical-v1/current-v2 action grammars. Old Go accepts all31 null numeric candidates
under a legal explicit-zero control; actual refusal regressions fail31 times, while the same
TS population/control passes32 checks. File RP-582. Add non-null checks only to catalog
objects/rows, including the standalone mood grammar; leave the mutable-state key helper and
unique-member gate unchanged. Genuine kernel0.3.199→0.3.200 accompanies this admission fix.

`PetSchemas` exports `PetCatalog` as the closed named v1/v2 union, with version-bound Soul
gate presence, existing owner enums and numeric domains. Historical v1 fixture, current and
two stored v2 artifacts, legal-zero/maximum-domain controls and all three gates pass unchanged.
All48 from-state/event/to-state combinations validate as single-row deterministic fixtures;
no policy content is minted. Twenty-two malformed shapes refuse; seven schema-valid relational
counterexamples still fail the loader. Exact stat/mood cardinality, ordering, Trust/floor and
recovery relations, deterministic transition uniqueness and Soul context remain owner checks.
Escaped members pass, raw duplicates still refuse at the owner, and fresh exports are isolated.

Final `make test-go GO_PACKAGES='./pet ./publicapi ./kernel ./replaycatalog ./decimal'
GO_TEST_FLAGS='-count=1'` and affected vet (also production) pass. Node affected catalog/care/
Founder/Exit and numeric population passes6360; Chromium/WebKit affected population passes12718,
then final null/standalone-grammar population passes66 after adding that extra test. Actual
`TestPetIdentityAndCareSurviveBothExitPaths` executes both nonempty Exit preservation arms,
no skips. Types report zero errors/warnings; build and byte-unchanged `make api-check` pass.
Balance/testdata/generated bytes, care math, clocks, state codecs and version activation remain
unchanged. No DB/composed player, Firefox/Linux/hosted or full CI/release claim; the known
history-guard/Clout failures are not waived and service access remains separately unresolved.

Review by: Codex (implementer first filter). Recorded by: Codex. Complete catalog/schema/
tests/version/docs/records range after `293a05e4` through this containing commit requires
designated Claude review, consolidatable with C18 work. Eleven of nineteen pinned owners now
export; eight, the complete union/catalog HTTP and public re-verification remain. No draft/
owner adoption, archival, publication or narrower1.0 floor. Continue accepted construction.

## 2026-10-14 — Faction catalog schema and original-member admission (RP-583/584)

Resume clean `3d9f621f`; read the complete archived Faction contract and actual Go/TS loaders.
F1/FC require literal root/row/Compact members, including present nullable `compact` on all
four rows. Old Go admits30 aliases and three missing null bindings: all33 new refusals fail
against the old reader while TS34 and canonical null controls pass. Enforce exact original
decoded shapes before struct decoding; require non-null fields except explicitly nullable
Compact. Preserve escaped names and complete last-value controls, not a new duplicate ban.
Genuine kernel0.3.200→0.3.201; no stock/incorporation/tithe or artifact changes.

`FactionSchemas` exports the closed F1/FC root, faction/resource enums, stock domains, typed
nullable Compact and reserved modifier-row grammar over the existing slot registry. Current
and two stored catalogs load/schema-validate against their matching Commons bands, without
byte changes. Valid maximum-domain retunes,18 structural refusals and ten semantic
counterexamples pass. Hamiltonian cycle/cardinality, copy/Compact relationships, declared
tithe band and empty modifier lists remain mandatory owner/composition checks. Fresh bounds
and enum mutations do not affect the owner. Full catalog union/HTTP/public verification remain.

Final cold faction/publicapi/kernel/replaycatalog/decimal packages and affected vet (including
production) pass. Node catalog/replay/numeric6476 and native Chromium/WebKit12952 pass; types
report zero errors/warnings, build and byte-unchanged `make api-check` pass. Two pure production
tests execute: `TestIncorporateAndOpenSourceBinding` and
`TestApplyLoggedDerivesFactionStockResourceInsideBoundary`, no skips. The first attempted
combined selector failed shell parsing before any tests; separate selectors actually ran.
RP-584 reconciles stale canonical docs against existing Guild settlement/provider and
leaderboard projector code, not new runtime implementation or acceptance. Balance/testdata/
generated API bytes remain unchanged. No DB/composed player, Firefox/Linux/hosted/full CI or
release claim; known history-guard/Clout/service limitations are not waived.

Review by: Codex (implementer first filter). Recorded by: Codex. Complete code/schema/tests/
version/docs/records range after `3d9f621f` through this containing commit needs designated
Claude review, consolidatable with C18. Twelve of nineteen pinned owners export; seven,
complete union/catalog serving and public re-verification remain. No owner/draft adoption,
archival/publication or narrower1.0 floor. Continue the accepted construction.
