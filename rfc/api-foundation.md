# RFC: API Foundation (versioning, schemas, the public read surface)

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-03
- **Design refs:** `design/06` (chi REST, closed surfaces), `design/05` (published formulas — the transparency law), the fan-tooling stance (`design/05 §Neopets adoptions`: "a documented read-API stance so the JellyNeo-equivalent builds WITH us from day one")
- **Research:** `neopets-social-history.md §5` (the shadow-infrastructure lesson), `wc3-custom-ecosystem.md §3` (community-built platform gaps), `speedrun-governance.md` (validator transparency)
- **Depends on:** Account API + Transport + Gameserver Composition (implemented)
- **Owner ruling honored:** breadth-first — this is the contract every future endpoint lands on, so the API never reshapes under consumers.
- **Planning:** `planning/api-foundation/` (once implementing)

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

- Public API stability tier (is `/api/public/v1/` under the same additive-only law? recommend
  yes, stricter: it's the one surface THIRD PARTIES build on).
- Display-name/identity policy for boards — the profile RFC owns it; until then boards expose
  the existing founder public id only.

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

- 2026-08-03: created (draft) — versioning law, generated schemas, the public read surface;
  the fan-tooling stance from the enclave research made structural.
- 2026-08-03: Codex acceptance review recorded C1–C10; implementation remains blocked pending
  owner rulings and reconciliation into one executable contract.
