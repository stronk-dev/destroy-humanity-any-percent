# RFC: API Foundation (versioning, schemas, the public read surface)

- **Status:** accepted (C1–C17 ruled; implementing)
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-03
- **Design refs:** `design/06` (chi REST, closed surfaces), `design/05` (published formulas — the transparency law), the fan-tooling stance (`design/05 §Neopets adoptions`: "a documented read-API stance so the JellyNeo-equivalent builds WITH us from day one")
- **Research:** `neopets-social-history.md §5` (the shadow-infrastructure lesson), `wc3-custom-ecosystem.md §3` (community-built platform gaps), `speedrun-governance.md` (validator transparency)
- **Depends on:** Account API + Transport + Gameserver Composition (implemented)
- **Owner ruling honored:** breadth-first — this is the contract every future endpoint lands on, so the API never reshapes under consumers.
- **Planning:** `planning/api-foundation/`

## Summary

The write surface (intents) is closed and disciplined; the API as a PLATFORM is not: no
versioning policy beyond `/v1/`, no generated schemas, no public read surface, no SDK boundary.
Every enclave we studied teaches the same thing — the community builds the database, tooling,
and institutional memory. This RFC decides we build WITH them from day one.

## Specification

### A1 — Versioning & compatibility policy

`/api/v1/` is additive-only: new endpoints and new OPTIONAL response fields are allowed; removals,
renames, type changes, and new required request fields require `/api/v2/` (which may be mounted
alongside indefinitely). The wire envelope's own `v` fields (transport v2, replay-inputs v2,
schema versions) are independent of the HTTP version — this RFC governs HTTP paths and JSON
response shapes only. A compatibility test pins every v1 response schema; breaking a pin fails
CI with the field named (the golden-fixture discipline applied to the API).

### A2 — Generated schemas as artifacts

OpenAPI 3.1 document GENERATED from the router + typed handlers (not hand-written; the
formula-artifact pattern — one authority), committed as `docs/generated/api.json`, drift-gated
in CI (regenerate + byte-compare). Client request/response types in TS generated from the same
source — the SDK boundary: `client/src/api/` becomes generated code, hand-written wrappers only
above it.

### A3 — The public read API (the fan-tooling stance made real)

New closed read-only surface `/api/public/v1/`, no authentication, aggressively cacheable
(ETag + max-age), rate-limited per-IP by the existing limiter:

- `GET catalogs/{epoch_id}` — the constants artifacts (already public-by-design: published
  formulas are law) + the generated formula artifacts.
- `GET epochs` — the epoch registry with changelogs.
- `GET boards/{category}/{epoch}?cursor=` — verified board pages (public data by definition).
- `GET runs/{stream}/{seq}/verification` — verdict + the replayable evidence bundle references
  (the shipped-validator transparency loop: a third party can fetch and re-verify).
- `GET registry/routes` — the Route Registry's public ledger (design/05 §6 already declares it
  public).

Explicitly ABSENT: anything founder-private (states, saves, presence), anything mutable.
Additions by RFC; the surface is closed like every other. Per-founder opt-outs are not needed
because nothing personal is exposed (board rows are already the public record; display
identity policy stays with the profile system when it lands).

### A4 — Error taxonomy & platform conventions

The typed-rejection shape becomes the documented API-wide contract (it already is in practice):
`{category, detail}` with the closed category registry exported into the OpenAPI doc. Plus:
request-ID header echo for support, standardized pagination envelope (keyset cursors as the
only pagination — already the house pattern), and a documented deprecation-header mechanism
for the day v2 exists.

### A5 — One operation and schema authority

A closed Go operation registry owns operation ID, method, path, stability surface, auth mode,
public flag, named request/success descriptors, and typed error/status alternatives. The router,
OpenAPI 3.1, generated TypeScript, compatibility pins, and runtime fixture validation consume
those rows. Schema descriptors are explicit Go values over the closed C17 union; reflection over
handler structs and handwritten OpenAPI fragments are forbidden. Startup rejects invalid schemas,
duplicate IDs, and duplicate method/path pairs.

### A6 — Public v1 DTOs and normalized board query

C12's DTO rows are the literal proposed contracts: catalog artifacts embed sorted named JSON;
epochs include ordered accepted hashes and changelog bytes; board items use the closed time/count/
magnitude key union; Route Registry exposes public credit/naming/adoption fields only; optional
values are explicit null and empty lists are `[]`. List limit defaults to 50 and is bounded 1..100.
Boards use `GET boards/{category}?variables=&epoch=&mandate=&cursor=`. Variables decode from the
canonical exact object `{advisor:0|1,commons:0|1,faction:string|null,glitched:0|1}` and every cursor
is bound to the complete normalized filter.

### A7 — Signed keyset cursor

Cursor bytes are `base64url(canonical_json(key_fields) || HMAC-SHA256)`. Authenticated key fields
are exact `{v:1,op,filter_sha256,key}`. The token is size-bounded; the codec verifies its MAC
against the current+previous deployment keys before parsing any JSON, then exact-decodes and
re-encodes the key fields. Secrets are stable deployment configuration and at least 32 bytes. Bad
MAC, malformed canonical bytes, unknown operation, or filter mismatch maps to
`invalid/cursor`. Cursors do not expire; key rotation retains the prior verifier through the
maximum cache lifetime.

### A8 — Public operational middleware

Strict `balance/api/phase0.json` is operational API configuration, loaded only by the API policy
owner and excluded from simulation artifacts. Request ID and trusted-client-IP resolution are
shared middleware. Conditional 304 is evaluated before the public limiter and therefore spends no
rate token per C16; it still echoes request ID, ETag, and Cache-Control. Strong ETag is SHA256 of
served bytes. Verification responses add `immutable`; other response classes use C6's literal
ages. Deployment secrets never enter JSON configuration or epoch identity.

## Endpoint-layer design gaps found during implementation

### C18 — Heterogeneous embedded artifact JSON conflicts with exact DTO schemas

C12 says `CatalogBundle.artifacts[].json` embeds arbitrary catalog JSON. C17 simultaneously
forbids unsupported maps/interfaces and requires exact descriptor validation. A single open-JSON
field cannot satisfy both, and silently adding an `any` schema arm would puncture the public DTO
law.

**Proposed contract:** `CatalogArtifact` is a discriminated `oneOf` keyed by literal artifact
`name`; each artifact owner exports its exact schema descriptor, and artifact-set growth adds a
response-union arm (allowed response widening under C2) in the same RFC/mint. Common fields remain
`{name,sha256,json}`. Do not add a free-form JSON/map descriptor.

### C19 — Raw verification bodies need an operation response descriptor

C14 requires exact `application/json` genesis bytes and `application/gzip` replay archive bytes,
while C17 currently describes every success as a named JSON DTO. Forcing raw bytes through a JSON
schema would re-encode the evidence or lie in OpenAPI.

**Proposed contract:** an operation response is the closed union
`schema {status,content_type:"application/json",schema_ref}` or
`raw {status,content_type:"application/json"|"application/gzip",content_hash_header}`. Raw handlers
write repository bytes without decoding; generation documents binary/string format and the
conformance test hashes exact bytes. No generic media type is accepted.

## Acceptance criteria

1. Schema generation: `make api-schema` regenerates `docs/generated/api.json`; CI drift gate;
   a seeded handler change without regeneration fails.
2. Compatibility pins: v1 response-schema fixtures; a seeded field removal fails CI naming the
   field; a seeded optional addition passes.
3. Public reads: each endpoint golden-tested (content + cache headers + limiter application);
   the verification endpoint round-trips a real verified run's evidence references; auth
   endpoints reject nothing new (no regression).
4. Generated TS types compile and the client's hand-written API layer is replaced by wrappers
   over them (diff shows deletion, not addition).
5. The privacy assertion: an integration test enumerates `/api/public/v1/` responses against a
   seeded founder and proves no founder-identifying field beyond public board identity appears.

## Open questions

- Both namespaces are additive-only for v1's lifetime per C2; this is no longer open.
- Display-name/identity policy for boards — the profile RFC owns it; until then boards expose
  the existing founder public id only.

## Owner rulings on C1–C10 (2026-08-03)

**C1 — accepted: a closed Go operation registry is the single source.** Each operation owns
method, path template, stability surface, auth mode, named request/success DTO(s), typed
error/status set; the chi router mounts FROM the registry; OpenAPI + TS types generate from the
same rows; startup rejects duplicate method/path or operation IDs; existing handlers adapt behind
registered operations (no reflection over bodies, no hand-written OpenAPI). Generator + registry
version pinned, output byte-canonical.

**C2 — accepted: both `/api/v1/` and `/api/public/v1/` are additive-only for v1's lifetime**, with
the exact comparison algorithm as proposed (allow new operations / new optional response props /
response-enum widening; reject removal, method/path change, prop removal/rename, optional→required,
type/format change, constraint narrowing, status removal, new required request prop; **request
enums do NOT grow inside v1** — generated exhaustive clients would silently break). Exceptions are
`/v2`, never a bypass tag.

**C3 — accepted, keys corrected:** `GET catalogs/{constants_hash}` (not epoch — an epoch holds
multiple hashes); boards keyed by the full tuple `boards/{category}/{variables}/{epoch}/{mandate}`
with the ranking-key family implicit in the category; Route Registry and epoch rows get declared
public DTOs generated from their operation rows.

**C4 — accepted: the verification endpoints must be DEREFERENCEABLE for the transparency loop to
be real** — `/api/public/v1/runs/{stream}/{seq}/genesis`, `/replay-log`, `/verdict` return the
immutable genesis, ordered replay inputs/receipts/events, and verdict. **This is compatible with
the privacy assertion because a VERIFIED run is already public record** (it's on a board); the
privacy test asserts only that UNVERIFIED/private founder state is never reachable. Imported and
drifted runs (never boardable) are not exposed.

**C5 — accepted: formula artifacts become epoch artifacts.** The generated formula document joins
`catalog_artifacts` keyed by constants hash at mint (an eighth artifact — the schema-v4/content
mint is the natural place), so `GET catalogs/{hash}` serves historically-exact formulas. Until a
mint carries it, the public endpoint serves the current-epoch formula document with its hash
labeled.

**C6 — accepted: cache/limiter literals fixed.** `max-age` per class (catalogs/epochs: 3600;
boards: 60; verification: immutable → 31536000; registry: 300), ETag = the served bytes' hash,
`/api/public/v1/` gets its own per-IP limiter instance (config literals in `balance/api/phase0.json`)
sharing the trusted-proxy parsing already ruled for the account limiter — extracted to a shared
middleware, not duplicated.

**C7 — accepted: two distinct shapes, documented as such.** Intent REJECTIONS are persisted 200
responses carrying the receipt envelope (application outcomes); HTTP/router/auth/limiter/server
failures are top-level `{category, detail}` errors. The OpenAPI doc models them as different
response types per operation; A4 is corrected to stop conflating them.

**C8 — accepted: one opaque signed cursor.** `{key_fields, tamper_mac}` base64url, MAC'd with a
deployment secret so a forged cursor rejects; per-operation the key_fields differ (time/count/
magnitude boards, registry, epoch) but the envelope and tamper handling are shared. Keyset only,
no offsets (the house rule).

**C9 — accepted, AC4 corrected:** there is no hand-written client HTTP layer to delete; AC4
becomes "the generated TS client is the ONLY HTTP-calling code; a lint forbids raw `fetch` to
`/api/` outside `client/src/api/generated/`."

**C10 — accepted: composition and privacy are structural** — the public router mounts only
registry operations flagged `public: true` (a field, loader-checked), and the privacy integration
test enumerates every public operation against a seeded private founder proving no private field
leaks (not a spot check — the full registry).

## Closure blockers after owner rulings (Codex re-review, 2026-08-03)

C1–C10 settle the architecture. The accepted text still cannot be implemented as a stable public
wire contract until the following details are ruled and reconciled into A1–A4. These are bounded
closures, not a redesign.

### C11 — The pre-mint formula fallback violates C4/C5

Serving the current formula document under a constants hash whose `catalog_artifacts` do not
contain those bytes is neither historical nor dereferenceable. It also makes a supposedly
immutable catalog response change after deployment while retaining the same resource identity and
ETag premise.

**Proposed contract:** remove the fallback. API Foundation itself performs the protocol-compliant
artifact-set-growth mint that adds `formulas` before enabling `GET catalogs/{constants_hash}`.
Generate canonical bytes under `balance/generated/production-formulas.json` (the epoch seed only
admits `balance/` paths), make the docs copy derive from the same generation, and register the
artifact through the single epoch authority. A requested accepted hash missing `formulas` returns
typed `not_available/historical_formulas`—never current bytes mislabeled as history. The API can
therefore ship before T0–T1 and does not inherit its F1 gate.

### C12 — The public DTOs remain undeclared

C3 says DTOs “get declared” but does not declare them. Epoch changelog representation, catalog
artifact encoding/order, board filters/items/ranking-key union, Route Registry fields, null rules,
limits, and unknown-ID behavior are all permanent additive-only API decisions.

**Proposed contract:** add literal v1 DTO tables to A3:

- `CatalogBundle {constants_hash, artifacts:[{name,sha256,json}]}` sorted by name; artifacts must
  be valid JSON and are embedded as JSON values, not base64. Unknown/unaccepted hash → 404
  `unknown_id/constants_hash`.
- `EpochPage {items:[{epoch_id,name,started_at,ended_at,changelog_ref,changelog_markdown,
  accepted_hashes}],next_cursor}` ordered newest-first; timestamps are UTC RFC3339 milliseconds,
  hashes byte-sorted.
- `BoardPage {category_id,variables,epoch_id,mandate_level,ranking_kind,
  items:[{run_id,founder_id,rank,key,verified_at,world_first}],next_cursor}` where `key` is the
  closed union `{kind:"time_ms"|"count",value:int64}` or
  `{kind:"magnitude",exponent:int64,quantized_mantissa:int64}`.
- `RoutePage` exposes only permanent/public registry facts: route ID, public name, first-executor
  founder ID (nullable after anonymization), credited-at, naming status/deadline, and adoption
  count; no moderation notes or account identity. Define its exact DB source and ordering.

All list limits default 50, range 1..100. Empty lists are `[]`; optional values are explicit null,
never omitted. The owner may alter fields, but implementation cannot invent them.

### C13 — Board path/filter encoding is not canonical

`boards/{category}/{variables}/{epoch}/{mandate}` puts structured JSON in a path but supplies no
encoding. Variables include nullable faction plus booleans, and a cursor must be bound to the full
normalized query. Category determines ranking kind only after loading its pinned catalog.

**Proposed contract:** `variables` is unpadded base64url of canonical exact-key JSON
`{"advisor":bool,"commons":bool,"faction":string|null,"glitched":bool}` with keys in that byte
order. Decode is size-bounded and re-encode-equal; noncanonical input is
`invalid/variables`. Epoch and mandate are strict base-10 integers without signs/leading zeros.
The category must exist in the **requested epoch/hash category artifact**, and its timer selects the
repository method/ranking union. The normalized filter hash used by C8 covers decoded variables,
category, epoch, mandate, and limit.

### C14 — Public replay evidence has no byte/content contract

The three endpoints name concepts but not response bytes. `run_log_archive` is gzip JSON, genesis
is JSONB text, events live separately, and a third-party verifier needs a manifest binding them to
the engine and constants bundle. Returning convenient re-encoded objects could invalidate the
claimed immutable evidence hash.

**Proposed contract:** expose a small manifest at `/verdict` containing run ID, verdict, engine
version, constants hash, genesis URL+SHA256, replay-log URL+SHA256, and catalog URL. `/genesis`
returns the exact canonical genesis bytes captured by the run-genesis contract as
`application/json`. `/replay-log` returns the exact immutable archive bytes as
`application/gzip` with its stored SHA256; it includes ordered replay inputs, receipts, and events
per the archive encoding already documented. All three authorize only when at least one
`verified_runs` row exists for the run; queue status alone is insufficient. Nonboardable/private
runs are indistinguishable from unknown (404). ETags equal the quoted stored content hashes.

### C15 — Signed cursor cryptography and lifecycle are unspecified

“MAC'd with a deployment secret” leaves canonical bytes, algorithm, secret length, filter binding,
key rotation, and error behavior undefined. A restart with a new ephemeral secret would invalidate
every cursor; accepting attacker-controlled key fields before MAC validation risks parser abuse.

**Proposed contract:** envelope is unpadded base64url canonical JSON
`{v:1,kid,op,filter_sha256,key,mac}` with `mac = base64url(HMAC-SHA256(secret,
canonical_json_without_mac))`. Secrets are deployment configuration, minimum 32 bytes; one current
`kid` signs and a bounded keyring verifies during rotation. Decode caps the token at 2048 bytes,
checks exact keys/types and known `kid`, verifies MAC in constant time, then parses operation key
fields. Operation/filter mismatch returns 400 `invalid/cursor`; cursors have no time expiry and
key rotation keeps old verification keys for at least the maximum declared cache lifetime.
`CompositionConfig` receives the keyring explicitly; startup fails closed if absent.

### C16 — Cache/limiter/request-ID middleware is only half-ruled

C6 provides max-age literals but not `immutable`, conditional requests, whether 304 spends rate,
API-config identity, or request-ID behavior from original A4/C7. `balance/api/phase0.json` would be
mistaken for simulation balance unless its ownership is explicit.

**Proposed contract:** operational API policy lives at `config/api/phase0.json`, not in the epoch
artifact set, with strict schema and literals for public burst/per-minute/max entries, proxy hops,
cache ages, and cursor key IDs (secrets remain environment-only). Limiting and request-ID assignment
run before conditional-GET handling, so 304 spends one token. Strong ETag is SHA256 of exact served
bytes; matching `If-None-Match` returns 304 with ETag, Cache-Control, and request ID but no body.
Verification evidence uses `public,max-age=31536000,immutable`; other classes use the ruled ages
without `immutable`. `X-Request-ID` follows the original proposed UUID/lowercase bounded validation,
is generated UUIDv7 when invalid/missing, echoed on every response, and logged—not added to JSON.

### C17 — Registry generation still needs a type-schema authority

An operation row referencing a Go DTO does not itself define required/optional fields, integer
bounds, formats, exact objects, unions, or TS names. Reflection over arbitrary structs loses
semantic bounds; handwritten schemas beside structs recreate two authorities. Existing handlers
also return anonymous structs/raw receipts and must be migrated without changing behavior.

**Proposed contract:** named DTOs carry one closed repository-owned schema descriptor through Go
types implementing `APISchema() Schema`; operation registration accepts only such types. The same
descriptor validates runtime response fixtures, generates OpenAPI and TS, and compatibility-checks
pins. Descriptors are code (not a second JSON artifact), use a closed union of object/array/string/
integer/boolean/null/ref/oneOf with bounds/formats, and reject unsupported maps/interfaces. Raw
`IntentReceipt` is one named exact schema supplied by its owning package. A registry conformance
test invokes every operation with seeded success/error paths and verifies emitted JSON against its
descriptor, preventing declared-schema/handler drift.

## Owner rulings on C11–C17 (2026-08-03)

- **C11 — accepted, the fallback removed:** no pre-mint formula fallback. `GET catalogs/{hash}`
  serves ONLY hashes whose `catalog_artifacts` actually contain the bytes; a hash without a stored
  formula artifact 404s for the formula sub-resource. The formula artifact becomes an epoch
  artifact at the schema-v4 mint (C5) — until then the endpoint honestly has no historical
  formulas to serve, and says so. No same-identity mutable response.
- **C12 — accepted:** the public DTOs are DECLARED as generated schema descriptors from the
  operation registry (C17's authority) — epoch changelog, catalog-artifact encoding/order, the
  board filter/item/ranking-key union, Route Registry fields, null rules, limits, and unknown-ID
  behavior are each pinned in the registry row's DTO and compat-tested. (This RFC's implementation
  batch enumerates them; the grammar is C17's.)
- **C13 — accepted:** boards use a NORMALIZED query, not JSON-in-path — `GET boards/{category}?
  variables={canonical-encoded}&epoch=&mandate=&cursor=`; the canonical variable encoding
  (sorted keys, explicit-null faction, booleans as 0/1) is declared and the cursor binds to the
  full normalized query (C15's MAC covers it). Ranking kind resolves from the category's pinned
  catalog after load.
- **C14 — accepted:** the verification endpoints return the RAW immutable bytes with a manifest —
  `run_log_archive` gzip-JSON as-stored, genesis JSONB text as-stored, events as-stored, plus a
  `manifest` binding them to engine version + constants hash + verdict. A third party re-runs the
  shipped validator on exactly those bytes (no convenient re-encoding that could invalidate replay).
- **C15 — accepted:** cursor = `base64url(canonical_json(key_fields) ‖ HMAC-SHA256)`; the MAC
  secret is a DEPLOYMENT secret (32+ bytes, DP5 config, stable across restarts — NOT ephemeral, so
  cursors survive deploys), two-key rotation like the JWT keys; MAC validated BEFORE any key_field
  is parsed (no attacker-controlled parse); a bad MAC is a typed `invalid_cursor` rejection.
- **C16 — accepted:** `immutable` on the verification responses (they are), conditional requests
  supported, a **304 does NOT spend the rate budget**, `balance/api/phase0.json` is explicitly API
  config (a header comment + a distinct loader, never the balance harness's), request-ID echo per
  A4/C7.
- **C17 — accepted, the single authority resolved:** operation DTOs are declared as **explicit
  schema descriptors in Go** (a small typed schema-DSL, NOT reflection over arbitrary structs and
  NOT handwritten-JSON-beside-structs) — one authority generating both the OpenAPI schema and the
  TS types with bounds/formats/unions/exact-objects intact. The registry row references its
  descriptor; startup validates descriptor↔handler shape.

## Acceptance blockers (Codex review, 2026-08-03)

The platform direction is correct, but the draft does not yet define a reproducible HTTP surface.
Several named endpoints cannot be implemented against the shipped data model. The following
closures are required before acceptance.

### C1 — “Generated from router + typed handlers” has no source authority

The implemented chi router contains ordinary handler functions, anonymous request/response
structs, and raw JSON receipts. Go cannot derive exact OpenAPI response unions or compatibility
rules from that runtime shape. Generating a document from a second annotation set would violate
A2's one-authority claim.

**Proposed contract:** introduce a closed Go operation registry. Each operation owns method,
path template, stability surface, auth mode, named request DTO, named success DTO(s), and typed
error/status set. The chi router is mounted from that registry, and OpenAPI plus TS types are
generated from the same rows and named DTO schema descriptors. Startup rejects duplicate
method/path or operation IDs. Existing handlers are adapted behind registered operations; no
reflection over handler bodies and no hand-written OpenAPI fragments. The generator and registry
version are pinned, and generated output is byte-canonical JSON.

### C2 — The stability law and compatibility algorithm are unresolved

The public namespace is the one third parties consume, yet its stability is still an open
question. “Optional field addition passes” also needs an exact schema-comparison algorithm;
otherwise requiredness, enum growth, union arms, numeric bounds, and `additionalProperties`
changes will be classified inconsistently.

**Proposed contract:** both `/api/v1/` and `/api/public/v1/` are additive-only for the lifetime
of v1. Commit normalized per-operation request/response/error schemas as compatibility pins.
The checker allows new operations, new optional response properties, and widening a response
enum; it rejects operation removal, method/path change, property removal/rename, optional→required,
type/format change, constraint narrowing, response-status removal, and any new required request
property. Request enums do not grow inside v1 because generated exhaustive clients would become
incomplete. Any intentional exception is `/v2`, never a bypass tag.

### C3 — The public endpoint identifiers and response schemas are incomplete

`GET catalogs/{epoch_id}` is ambiguous because an epoch accepts multiple constants hashes.
Boards are keyed by category, variables, epoch, mandate level, and one of three ranking-key
families, but the proposed URL supplies only category and epoch. The Route Registry and epoch
rows also have no declared public DTOs.

**Proposed contract:** identify an immutable catalog bundle by `constants_hash`, not epoch:
`GET /api/public/v1/catalogs/{constants_hash}`. `GET /epochs` returns epochs with ordered
`accepted_hashes`, timestamps, and changelog content/ref. Board requests require exact query
fields for the shipped dimensions: `variables` in canonical encoded form, `mandate_level`,
`limit`, and optional cursor; category data determines the ranking-key family. Define closed DTOs
for every response, including null/empty behavior, ordering, bounds, unknown IDs/hashes, and
whether changelog bytes are inline or linked. Route Registry rows must name the exact public
fields, naming-reservation state, executor-credit/anonymization behavior, ordering, and cursor.

### C4 — Public replayability contradicts the privacy assertion

The listed verification endpoint returns only “evidence bundle references”, but no endpoint can
dereference them. A real verifier needs immutable genesis, ordered replay inputs/receipts/events,
engine version, and pinned catalog bytes. Those records contain founder ID, company stream ID,
exact action history, and run state. AC5 simultaneously claims the public surface exposes no
founder-identifying data beyond board identity.

**Proposed contract:** owner must choose one honest policy. Recommended: every ranked run is a
public artifact, and ranking consent makes its replay bundle public. Add an immutable evidence
endpoint keyed by public `run_id`, specify its canonical encoding/content hash and all included
fields, and state explicitly that founder public ID plus gameplay history are public while account,
session, email, recovery, IP, and non-ranked saves never are. The verification response links the
bundle hash and pinned catalogs; imported/drifted/unranked runs remain absent. If action history
must stay private, retract the “third party can fetch and re-verify” acceptance criterion rather
than returning unusable references.

### C5 — Formula artifacts are not epoch artifacts

The epoch authority pins seven balance artifacts. `docs/generated/production-formulas.json` is a
current generated document, not stored in `catalog_artifacts`, so an historical catalog endpoint
cannot truthfully return its epoch's generated formulas. Adding it to the constants identity is a
mint and affects every composition/parity site.

**Proposed contract:** generate an immutable formula artifact for each constants bundle during
the same source generation that mints it, include that artifact in the epoch seed authority, and
store its bytes in `catalog_artifacts`; adding it is a protocol-compliant mint. The public catalog
response returns only stored pinned bytes. If formulas remain derived rather than identity-bearing,
define a versioned derivation engine and prove historical regeneration byte-identical—do not serve
the current formula document under an old hash.

### C6 — Cache and limiter behavior is not executable

“Aggressively cacheable”, `max-age`, ETag, and “existing limiter” supply no literals. The existing
IP limiter is private to `account.API`; its unauthenticated policy and trusted-proxy parsing cannot
simply be attached to a new composed service without creating a second authority.

**Proposed contract:** extract one platform middleware owner for trusted-client-IP resolution,
request IDs, and bounded token buckets. Declare public burst/per-minute/max-entry literals in
config. Immutable catalog/evidence resources use strong ETags equal to their content hashes and
`Cache-Control: public, max-age=31536000, immutable`; mutable epoch/board/registry pages use a
declared short max-age and a strong ETag derived from canonical response bytes. Honor
`If-None-Match` with 304 and no body. Define whether a cache hit/304 spends a limiter token
(recommend yes, before handler work) and preserve the current trusted-proxy semantics.

### C7 — A4 conflates HTTP errors with intent receipts

Production rejections are persisted 200 responses with their own receipt envelope, while router,
auth, limiter, and server failures use top-level `{category,detail}` errors. Calling both one
“typed-rejection shape” would either break the replayed receipt or misdocument the HTTP surface.
The category registry and status mapping are not enumerated, and request-ID behavior is unnamed.

**Proposed contract:** document two distinct unions: `IntentReceipt` (unchanged authoritative
payload) and `APIError {category,detail}` for non-receipt HTTP failures. The operation registry in
C1 declares the closed error/status alternatives per endpoint, with one generated global category
catalog. Use `X-Request-ID`: accept only a bounded lowercase hex/UUID value or generate a UUIDv7,
echo it on every response including 304/errors, and include it in server logs—not response JSON.
Define the future deprecation headers now by standard names and date/link grammar.

### C8 — Pagination is not specified and cannot be shared blindly

Current boards use different cursor keys for time/count and magnitude orderings; registry and
epoch lists have different stable keys. “Keyset cursor” does not define encoding, tamper handling,
scope binding, limit bounds, or whether a cursor can be replayed against another query.

**Proposed contract:** cursors are opaque base64url canonical JSON carrying schema version,
operation ID, normalized filter hash, and that operation's ordered key tuple. Decode is exact-key,
size-bounded, and rejects cross-operation/filter reuse as `invalid_cursor`. Cursors are not secret
and need no signature because they convey only public keys; handlers revalidate every component.
Each list DTO is `{items, next_cursor}` with `next_cursor:null` at exhaustion and a shared
Phase-0 limit range/default.

### C9 — The generated-client acceptance criterion describes code that does not exist

There is no hand-written HTTP API layer under `client/src/api/` and no production `fetch` caller
to replace. AC4's required “diff shows deletion, not addition” is therefore impossible at HEAD.

**Proposed contract:** generate `client/src/api/generated/` types and operation metadata, then add
one thin hand-written transport client that owns base URL, auth token, request ID, retries, and
JSON decoding. Existing WebSocket transport remains separate. AC4 requires all new HTTP callers
to use this client and forbids handwritten request/response DTOs; delete the false deletion claim.

### C10 — Composition and privacy must be structural

The composed API currently owns account routes only. Public reads require leaderboard, epoch,
catalog, verification, and Route Registry repositories, but no public service interface or mount
ownership is declared. A response-enumeration test alone cannot prove a future DTO did not expose
a private field.

**Proposed contract:** `gameserver.Compose` constructs a dedicated read-only `publicapi.Service`
from narrow reader interfaces and mounts its generated registry beside account routes. The package
must not import account/session/save-state repositories; a build-graph lint enforces that boundary.
Public DTOs use explicit named fields—never database structs or `map[string]any`—and the privacy
test checks the generated schema allowlist plus seeded runtime responses. No public handler starts
a write transaction or exposes an operator mutation path.

## Changelog

- 2026-08-03: created (draft) — versioning law, generated schemas, the public read surface.
- 2026-08-03: C11–C17 ruled — no pre-mint formula fallback, normalized board queries, raw-bytes-plus-manifest verification evidence, HMAC cursors with a stable deployment secret validated-before-parse, explicit Go schema-descriptor DSL as the single generation authority. Fully accepted.
- 2026-08-03: C1–C10 ruled — operation-registry single source, additive-only both namespaces with the exact compat algorithm, dereferenceable verification endpoints (verified runs are public record), formula artifacts become epoch artifacts, signed opaque cursors, structural public/private split. Accepted.
- 2026-08-03: C1–C17 reconciled into A1–A8; implementation unblocked. Historical blocker sections
  remain as the review record, not active status.
- 2026-08-03: schema/cursor implementation found C18–C19 at the endpoint layer: embedded artifact
  JSON needs owner-schema union arms, and raw evidence needs a declared raw-response arm.
