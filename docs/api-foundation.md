# API Foundation

The API foundation has three implemented mechanical authorities: an explicit Go schema/operation
registry, generated API contracts, and authenticated keyset cursors.

Schema descriptors cover exact objects, arrays, strings with closed formats/enums, bounded int64
integers, booleans, null, named references, and one-of unions. Registry construction rejects an
invalid descriptor, duplicate operation ID, duplicate method/path, public/auth mismatch, or an
unknown request/response schema. Runtime response fixtures validate against the same descriptor
values consumed by runtime mounting, OpenAPI generation, TypeScript generation, and the committed
v1 compatibility pin. Authenticated Soul Recovery and minigame routes mount exclusively from this
registry; missing, extra, unsorted, or nil runtime bindings fail during router construction.

`ValidateJSON` checks raw object-member uniqueness before decoding to maps. Duplicate names
at any level, including different escape spellings of the same name, return `ErrInvalidSchema`
rather than silently keeping the final value. Valid escaped names, string contents resembling
JSON and repeated names in separate objects remain allowed. `ValidateRequest` and
`ValidateResponse` share this check; it does not add request-validation middleware to handlers
that do not already call the registry, or rewrite any successful handler response.

Operations may declare exact scalar query parameters (string, bounded integer, or boolean). They
are byte-sorted by name and never shadow a path parameter. The reserved `cursor` parameter is a
non-required string, and it is present exactly when the operation declares a cursor key.
`Registry.ParseQuery` decodes only declared parameters. Each parameter may appear at most once,
integers are strict base-10 (no sign, padding, or leading zero), and every value is checked by the
same descriptor validator as response bodies. A failure returns `InvalidQueryError` naming the
parameter; undeclared parameters are ignored.

Query parameters are generated as OpenAPI `in: query` entries, and TypeScript gets
`queryParameters` plus an optional-aware `query` type. Query-free operations generate
byte-identical output. The compatibility pin records the query set only for operations that have
one, and rejects any change to an existing operation's query set. That is stricter than C2 (which
would allow a new optional request field), but never looser.

A schema response may also declare a sorted immutable set of exact JSON wire bytes. This is a
runtime-validation narrowing layered over the generated DTO, not a generated-client or OpenAPI
shape change. Minigame error responses use it to bind each operation/status to the shipped
category/detail pairs and encoder newline, so a schema-valid cross-product or appended byte is not
accepted as contract evidence.

The shared error-detail enum includes `soul_recovery` (unavailable/internal recovery errors)
and `company_stream` (start without an active Company), matching existing handlers. These
are additive descriptor corrections, not handler or error-mapping changes. Real authenticated
Postgres-backed replies are checked against this same registry; invented details still fail.

`make api-schema` regenerates canonical OpenAPI 3.1 at `docs/generated/api.json` and exact client
DTO/operation metadata at `client/src/api/generated/types.ts`. `make api-check` regenerates and
byte-compares both outputs as part of `make verify-server`. The committed
`docs/generated/api-compat-v1.json` baseline enforces the v1 additive-only law: existing
operations, request unions, statuses, required fields, and bounds cannot narrow or disappear;
responses may add optional fields or widen an enum/union. Updating the compatibility baseline is
an explicit `make api-pin` operation, never an incidental effect of ordinary generation.
Existing optional fields cannot become required in requests, responses or shared definitions,
including through array/reference/union nesting; failures identify the affected field. Numeric
bounds compare as exact signed int64 values, not float64 projections, so a one-unit narrowing
above 2^53 or at an int64 boundary still fails. This checks schema metadata, not production math.
Schema names follow the current-plus-legacy convention: an unversioned name denotes the current
shape, while a retained historical shape uses an explicit `V<n>` suffix. A compatibility-pin
refresh must cite its authorizing ruling and be recorded in the owning planning log in the same
change; an otherwise valid widening is not permission for a silent re-baseline.

The generated registry is not yet the complete runtime API. Current metadata covers twenty-two
operations, including the existing session creation and refresh routes and their actual error
alternatives. Both routes mount from the registry with their existing shared unauthenticated IP
limiter. Their successful two-string token pairs reuse the `BootstrapSession` owner descriptor;
creation takes `{account_id,recovery_code}`, refresh takes `{refresh_token}`. Registered errors
are exact operation/status-specific bytes: `400 invalid/body`, `429 rate_limited/ip`, creation's
`401 unauthorized/credential`, and refresh's `401 unauthorized/refresh_token` or
`401 refresh_reused/session_family_revoked`. The refresh parser's existing null/missing/case-
insensitive behavior is unchanged; documenting canonical typed callers does not introduce
request-schema middleware that tightens the live handler.

Explicit generated calls can now represent those routes/errors; they do not add automatic
renewal, credential replacement, recovery UX or retry policy. The deliberate additive pin refresh
under API C1/C2 also captures the already-registered Garden route and optional Garage response
shapes; it changes no original operation rows. Tests against the actual committed pin reject
removal of either session operation or its 401 alternative. Bootstrap, main state reads and the
Minigame/Soul/Garden ports use the generated client; intents still call `fetcher` outside the
generated directory. `api-check` passes
for the registered subset, not AC4 completion. Manual
`make research-refresh-generated-contract` now verifies represented refresh callers and
in-memory metadata-removal refusals outside CI/verify; its earlier retained report remains
historical evidence of the gap. It does not implement or authorize browser renewal. See the
[bounded census](../planning/platform-alignment/refresh-generated-contract.md).

The existing account lifecycle routes also mount exclusively from registered rows:
`create_account` (`POST /api/v1/account`, unauthenticated), `create_founder`
(`POST /api/v1/founder`, access token) and `get_founder` (`GET /api/v1/founder`, access token).
Account creation retains the shared IP limiter and `no-store` credential reply; both Founder
routes retain shared account limiting and failed-auth IP limiting. Their handlers and accepted
body forms are unchanged. Generated POST calls use the canonical empty object; the live parser
still admits an empty body as well and rejects internally spaced `{ }`.

`AccountCreated` is `{account_id,created_at,recovery_code}`. `FounderCreated` is
`{id,created_at,imported}`; `FounderProfile` is `{id,created_at,display:{}}`. Display is an exact
empty object, not permission to introduce arbitrary profile data. Unlike bootstrap's fixed
`.000Z` format, these existing handlers use `time.Time`'s UTC millisecond encoding with trailing
fractional zeros omitted. The separate `date-time-utc-ms` string descriptor validates those
bytes without relaxing any existing `date-time-ms` descriptor or re-encoding a reply.

Actual `400 invalid/body`, `401 unauthorized/access_token`, `404 unknown_id/account|founder`,
`429 rate_limited/ip|account` and `500 internal_invariant/account_create|founder_create` errors
are registered only on their applicable operations/statuses. The deliberate C1/C2 pin refresh
adds these three operations; no existing operation row changes. Real Postgres tests cover
creation → session → current Founder → New Founder → refreshed Founder binding, retained
archives, refusal nonmutation and shared limiting. This does not add recovery UX or deletion
policy. Bodiless deletion responses, intent/import routes, C9 enforcement and full API
acceptance remain unfinished.

Exact empty-object descriptors now generate `Record<string, never>`, not TypeScript `{}`
(which accepts extra fields and primitives). This corrects SDK typing for account creation,
New Founder, the empty display object and existing Minigame resolution requests; it changes
no OpenAPI shape, HTTP parser or wire acceptance set.

### Generated HTTP client

`createAPIClient(fetcher?, baseURL?)` is emitted inside the generated module from an embedded
generator template, with operation-specific method/path, auth, query and response metadata
from the registry. `call(operationID, input)` accepts generated path/request/query types,
explicit access tokens only for private operations, and optional request ID/abort signal.
Path components are URL-encoded; declared query values are serialized in registry order.
Unknown path/query fields, missing auth and inexact integer parameters fail before HTTP.
Content-Type is sent only for operations with a JSON request body, not bodiless reads.
The client never forwards an access token on a public operation. There is no token storage,
refresh, retry, timeout, or request-ID policy hidden in it.

`OperationResponses` preserves status/body associations. Schema responses are JSON-decoded;
these TypeScript types are not runtime schema validation. Wrappers retain their existing
application decoders. Raw successes are `Uint8Array`, not an omitted success type or parsed
JSON. They require the registered media type and content-hash header, checked against SHA256
over downloaded bytes. Error statuses on raw operations remain JSON. Undeclared status,
malformed JSON, raw type/hash failure and network/abort errors are surfaced, never retried.
Malformed JSON reports its HTTP status without echoing reply bytes. Body-read failures
propagate unchanged rather than being relabelled as JSON errors.
The undeclared 304 arm remains open; this client does not invent an alternative for it.

The actual Game UI bootstrap and main state reads now use this transport without changing
the bootstrap journal, credential storage, snapshot decoder or revision reconciliation.
Lost replies and HTTP/parser failures retain the persisted bootstrap key for an explicit
retry; no client-level retry is added. Remaining HTTP callers and the C9 raw-fetch lint are
still unfinished, so this is not AC4 completion. Unit/browser tests use controlled fetches;
the existing composed journey separately exercises these migrated calls against real HTTP,
Postgres and WebSocket services. Raw-client controlled-byte tests are not yet a public
TypeScript archive-reverification journey.

The Minigame, Soul Recovery and Garden ports dispatch by generated operation ID with typed
path/request inputs. Their shared adapter reads the current access token once per explicit
call and preserves the existing MinigameAPIError/MinigameTransportError classes and surface
rejection mappings. Error categories come from the generated registry, not a second list.
Undeclared HTTP statuses are transport failures; declared errors retain the previous exact
two-field, known-category/nonempty-detail guard. This is not per-operation error-literal or
success-body schema validation. Existing wire-binding/error tests cover all nine port calls;
the real composed journey covers Pitch and default locked-Garden reads, not a complete
persisted Soul Recovery or active-Garden workflow.

GU-C26 authorized the Game UI schema v3 compatibility-pin baseline. Accepted Garage Player
Surfaces GS0.1 (2026-09-25) authorizes the v4 re-baseline. The unversioned `GameUISnapshot` is the
current v4 shape. `GameUISnapshotV1`, `GameUISnapshotV2` and `GameUISnapshotV3` retain exact
stored-bootstrap receipt validation, and the live Founder-state operation returns only v4.

Public pagination cursors contain canonical `{filter_sha256,key,op,v}` JSON followed by an
HMAC-SHA256 signature, encoded as unpadded base64url. The codec verifies the signature with the
current or previous deployment key before parsing JSON, binds the cursor to operation and complete
normalized query, and rejects noncanonical or oversized input. Board variables use canonical
exact-key JSON with integer booleans and explicit-null faction.

## Public reads

`server/publicread` owns the unauthenticated `/api/public/v1/` registry. `publicapi.APIErrorSchema()`
is the single `APIError` definition that both registries reference. `publicapi.MergeRegistries`
unions the private and public registries into one generation authority. A schema name may be
shared only if the definition is byte-identical, and operation IDs must not collide. As a result,
`docs/generated/api.json`, the TypeScript module and the compatibility pin cover both surfaces,
while each surface still mounts from its own registry.

The shared error schema belongs to `publicapi`, not the private Account repository. Public
readers no longer import Account to construct their registry. A production-source import guard
checks both `publicread` and `publicapi`, rejecting direct Account/Session/Save/Production owner
imports (including aliases and subpackages). It ignores comments and test fixtures and refuses
unparseable source. This is a direct package-boundary check, not proof of every transitive
repository dependency or runtime disclosure; those remain separate privacy acceptance gates.
The move preserves the error fields/enums, all operation/status-specific exact replies, and
byte-identical generated OpenAPI, TypeScript and compatibility pins.

`GET /api/public/v1/epochs` (`list_public_epochs`) returns the C12 `PublicEpochPage`:

- `items` are newest-first. Each has `accepted_hashes` (byte-sorted in Go, an empty list when
  there are none), the exact UTF-8 `changelog_markdown` read from the staged `changelog_ref`, and
  UTC millisecond `started_at`/`ended_at` (`ended_at` is explicit `null` while open).
- `next_cursor` is a C15 keyset cursor on `epoch_id`, bound to the normalized filter
  `{"limit":N}`. `limit` defaults to 50 and is bounded 1..100.
- **Rejections:** a malformed limit returns exactly `400 {"category":"invalid","detail":"limit"}`.
  A malformed, tampered or filter-mismatched cursor returns `400 invalid/cursor`.
- **Failures:** repository or changelog failures, including a missing or non-UTF-8 staged
  changelog, return `500 internal_invariant/public_api`.
- **Caching and validation:** successful bytes are validated against the registry row. They are
  then served through the shared cache/limiter runtime with the `catalogs_epochs` class
  (`public,max-age=3600`, strong ETag, and a 304 that spends no rate token).
- **Page source:** `leaderboard.Repository.PublicEpochPage`, which reads the page from a single SQL
  statement.

`GET /api/public/v1/boards/{category}` (`list_public_board`) returns the C12 `BoardPage`
`{category_id, epoch_id, items, mandate_level, next_cursor, ranking_kind, variables}`.

- **Query:** C13's normalized query. `variables` is the base64url canonical JSON
  `{advisor,commons,faction,glitched}` (sorted keys, 0/1 flags, explicit-null faction). `epoch`,
  `mandate` (0..20) and `variables` are required; `limit` (1..100, default 50) and `cursor` are
  optional.
- **Ranking kind:** resolved from the pinned categories artifact of every stored constants hash
  accepted into the epoch. Timed categories (`rta`/`attended`) are `time_ms` and the untimed
  valuation category is `magnitude`. `count` is declared for the union, but no category produces
  it yet. Catalogs that disagree, or a stored catalog that cannot load, are internal invariants.
  The accepted-hash result must complete successfully: a row-iteration failure is propagated
  as a server failure, never resolved as an unknown category from an incomplete result.
- **Items:** `{founder_id, key, rank, run_id, verified_at, world_first}`, where `key` is the closed
  union `{kind:"time_ms"|"count",value}` / `{kind:"magnitude",exponent,quantized_mantissa}`.
  `rank` is the competition rank over the whole board.
- **Cursor:** the keyset cursor (`{key|exponent+quantized_mantissa, run_id}`) is MAC-bound to the
  complete normalized filter (category, variables, epoch, mandate, limit). A cursor whose arm
  doesn't match the ranking kind is rejected.
- **Paging:** the reader fetches `limit+1` rows, so a page that is exactly full carries no cursor.
- **Rejections:**
  - `400 invalid/{cursor,epoch,limit,mandate,variables}`;
  - `404 unknown_id/{category,epoch}`;
  - `429 rate_limited/ip`;
  - `500 internal_invariant/public_api`.
- **Caching:** the `boards` class (`public,max-age=60`).

`GET /api/public/v1/registry/routes` (`list_public_routes`) returns the C12 `RoutePage`
`{items, next_cursor}`.

- **Source:** `registry_routes` joined with `account_founders`, through
  `routeprojection.Projector.PublicRoutePage`.
- **Items:** `{adoption_count, credited_at, first_executor_founder_id, naming_deadline,
  naming_status, public_name, route_id}`.
  - `public_name` is the approved player name only when `published`. Reserved, pending
    (unmoderated) and expired names show the house name.
  - `first_executor_founder_id` is `null` once that Founder's account is anonymized (or no
    ownership row exists).
  - Moderation state beyond the status enum and all account identity are absent.
- **Order and paging:** ordered by the immutable `route_id`; `credited_at` can move when an
  earlier execution re-credits a route. The keyset cursor is bound to `{"limit":N}`, and `limit`
  is 1..100 (default 50).
- **Rejections:** `400 invalid/{cursor,limit}`, `429 rate_limited/ip`,
  `500 internal_invariant/public_api`.
- **Caching:** the `registry` class (`public,max-age=300`).

**Privacy (AC5)** is enforced two ways:

- **Structural:** `publicread/privacy_test.go` requires every public operation to be an
  unauthenticated GET under `/api/public/v1/` with no request body. It walks every public schema
  field and rejects account, email, token, session, recovery, password, secret, stream, save, IP,
  device and presence fields. It permits founder identity only at `PublicBoardItem.founder_id` and
  `PublicRoute.first_executor_founder_id`.
- **Composed:** the composed-server witness seeds a real private account. It then requests every public
  registry operation (a missing request builder fails the test) and asserts that no response
  body or header contains the seeded account ID, recovery code, tokens, founder ID or company
  stream ID. Evidence endpoints return identical, non-cacheable 404s for private and unknown
  runs. A separate actual verified-run download check permits C4's public Founder/Company
  history but rejects account IDs, recovery codes and session tokens in the manifest, genesis
  and decompressed replay archive. Raw-response privacy requires this content check; the
  schema walker cannot inspect a gzip body.

`publicread.NewRouter` composes the surface. It loads the strict policy, resolves the named cursor
secrets (`CursorSecretResolver`, see `docs/gameserver.md`), builds the request-ID runtime and
limiter, and mounts every public registry operation (and only those) through `Registry.Mount`.
Unknown public paths and non-GET methods return `404 unknown_id/route` and never fall through to
the account router. The gameserver mounts it at `/api/public/v1/`, and composition fails closed
without a valid policy or cursor pair. The composed-server Postgres witness checks all of this:
the served epoch page, request-ID echo, cache headers, a 304 on a matching ETag, the unknown-route
404, and fail-closed composition. The composed Game UI lane also fetches the page through the Vite
proxy.

### Public verification evidence

The registry mounts three unauthenticated GETs at
`/api/public/v1/runs/{stream}/{seq}/{genesis,replay-log,verdict}`. The stream is a canonical
Company UUID; the sequence is an exact positive integer without signs or leading zeros.
Private/unverified, unknown and invalid run identities return exactly
`404 {"category":"unknown_id","detail":"run"}`; corrupt authorized evidence or operational
failures return `500 internal_invariant/public_api`. Neither response exposes evidence or
carries cache/content-hash metadata.

`verdict` returns the exact nine-field `PublicRunVerdict` descriptor:
`catalog_url`, `constants_hash`, `engine_version`, `genesis_sha256`, `genesis_url`,
`replay_log_sha256`, `replay_log_url`, `run_id`, `verdict`. The only verdict is `verified`.
URLs are relative public API paths; hashes use `sha256:` prefixes. The catalogs URL names
the run's pinned constants, never current constants. Its HTTP reader is still unavailable.
Genesis version is present in the archive rather than an extra manifest field.

`genesis` serves the original stored JSON as `application/json`; `replay-log` serves the
original gzip bytes as `application/gzip`. Both raw registry arms require `X-Content-SHA256`
(unprefixed hex over served bytes). All three successful responses have strong SHA256 ETags
and `public,max-age=31536000,immutable`. Matching conditional reads return bodiless 304s
before charging the shared IP bucket, retaining cache/ETag and raw hash headers. Exhausted
uncached reads return non-cacheable `429 rate_limited/ip` without a raw hash header.

`leaderboard.Repository.PublicRunEvidence` supplies C14's stored bytes and pinned metadata in
one statement snapshot. Only an actual `verified_runs` record authorizes retrieval; a queue
verdict, run pin or archive alone does not. Invalid/unknown run identities return
`ErrUnknownPublicRun` without evidence. Missing or corrupt evidence for an authorized public
run returns `ErrInvalidPublicRunEvidence`; database errors propagate as operational errors.

The source returns the original genesis JSON bytes/version, the original `gzip+json.v1` replay
archive, engine version and constants hash. It verifies that genesis uses the pinned constants
hash and that archive bytes match the stored SHA256, and computes the genesis SHA256 over the
stored bytes. It does not re-encode JSON, recompress gzip, substitute current engine/constants,
or expose account/session metadata. Caller mutation does not rewrite stored evidence.

The repository's real-Postgres test uses synthetic evidence to verify retrieval and refusal,
including queue-only verification, missing bytes, corrupt hashes and cancellation. The normal
composed lane requires this test to execute. It also executes the existing composed Exit/board
witness: a real queue-verified run is downloaded over unauthenticated HTTP, every served byte
and digest is checked against immutable storage, and its Company/`founder_advanced` event
projection reverifies with the pinned DB catalog bundle. This matches the database verifier's
selection; the archive retains additional Founder history (including Fiscal events), which
that kernel projection does not reverify. A deliberately corrupted downloaded receipt is
rejected. Embedded genesis JSON is compared with whitespace-only compaction because the
existing archive encoder embeds the stored JSON value compactly; endpoint/storage equality
remains byte-exact. This does not prove cross-epoch catalog retrieval, a public TypeScript
verification journey, or the complete third-party loop without database-supplied catalogs.

### Historical catalog database source

`leaderboard.Repository.PublicCatalog` now supplies the exact stored artifact set for an
epoch-accepted constants hash, including historical epochs. It reads acceptance and bytes in
one statement snapshot; a stored but unaccepted set is indistinguishable from an unknown hash
(`ErrUnknownPublicCatalog`). Multiple epochs accepting the same hash do not duplicate artifacts.

The internal bundle carries the constants hash and artifacts sorted by name, each with its
original bytes and `sha256:` digest. It recomputes the bundle identity using the existing
length-framed constants-hash authority. Missing, additional or altered bytes, invalid names,
duplicate names, invalid UTF-8 or malformed JSON fail as `ErrInvalidPublicCatalog`; database
and interrupted-row errors propagate without returning partial evidence. Returned buffers do
not alias the source. The reader neither consults current filesystem catalogs nor generates
formulas. A historical set lacking formulas remains exactly that set.

These are internal byte sources, not registered wire DTOs. The real-Postgres test retrieves
the committed artifact set, mints a test-only newer formula-bearing set through the epoch
repository, and retrieves both identities without substituting newer bytes into the older one.
It also checks unaccepted sets, corrupt/missing accepted evidence and cancellation. The test
mint does not modify the product manifest or authorize a release mint; it does not establish
the complete public HTTP/TypeScript verification loop. The separate replay integration below
now covers loading that kind of stored set.

`server/formulas` owns the extracted version-14 production-formula model and exports its
closed `ProductionFormulasV14` descriptor through `Schemas()`. The actual generator consumes
that model and validates its output before writing; the published artifact remains byte-identical.
`Validate` checks stored bytes without rewriting or regenerating them, rejecting unsupported
versions, undeclared/missing fields, wrong nested shapes/enums, malformed UTF-8/JSON, invalid
digests/Decimal strings and integers outside signed int64. The integer domain mirrors the
existing model, not newly chosen operating limits. This is a grammar check, not proof that
arbitrary prose is mathematically true or belongs to a particular epoch.

The descriptor covers the current formula version only, including a null axis pin and all
three declared populated input variants. Historical version support must be explicit; no
open-JSON arm or current regeneration fallback is introduced. This owner export is not yet
registered in the catalog response union or product epoch manifest.

Both replay loaders now accept an optional stored `formulas` artifact, validate its version-14
grammar, and include its original bytes in the full constants identity. Go's subsequent catalog
resolution rechecks that grammar and identity, including caller mutation. Older sets without
formulas remain loadable and never gain current formulas. `make formulas` derives the client's
closed descriptor metadata from the Go owner; signed int64 bounds are decimal strings, and
the client preserves raw number tokens before validating them with BigInt. No JavaScript Number
rounding, formula regeneration, new command or simulation arithmetic is introduced.
The acceptance-set extension advances the mirrored replay kernel to `0.3.168`; formula validator
sources and generated metadata now belong to its affecting-path guard.

Shared controls exercise required shapes, null/populated branches, malformed values, exact
int64 limits/overflow, JSON last-value semantics and string formats in both runtimes. A real
Postgres fixture mint loads and resolves alongside the older epoch, and its public repository
bytes also load without rewriting; a hash-matching unsupported formula version refuses.
These are test-only epochs, not adopted balance or proof of the HTTP/third-party verifier loop.

`leaderboard.CategorySchemas()` now exports the closed `CategoryCatalogV1` owner descriptor.
Its five existing category arms bind each ID to its name key, timer and predicate shape;
fact-set keys and nested predicate fields are closed. Current, historical and candidate catalog
bytes validate without normalization. Unknown categories/fields, mismatched row bindings,
unsupported versions, wrong types and invalid int64 literals refuse. The descriptor does not
replace `LoadCategoryCatalog`: array cardinality/order, gate correspondence, fact membership
and canonical policy values remain loader checks. A grammar-valid count of 41 is still refused
by that loader's existing Phase-0 policy. No balance rule or historical bytes changed.

The catalog HTTP reader, remaining generated-client caller migration and full public
TypeScript verification loop remain open. Categories is the first owner descriptor for the
nineteen currently pinned artifact families; the other eighteen and the composed C18 catalog
union remain unfinished. The separately exported formula artifact needs a
protocol-compliant product mint and public registration. No open JSON wire arm or
current-formula fallback is introduced.
