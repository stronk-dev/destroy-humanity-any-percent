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
