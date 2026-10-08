# RFC: Bodiless Responses in the API Registry

- **Status:** draft — owner adoption and designated review pending; not implementation authority.
- **Author:** Codex (proposal for Marco)
- **Created:** 2026-10-09
- **Design refs:** `design/06-tech.md` (REST, server authority and versioned contracts)
- **Depends on / amends:** API Foundation A2/A5/A8, C9/C19; preserves A1/C2 and C16/C20.
- **Planning:** existing `planning/api-foundation/` plan/log; no separate execution program.

## Summary and owner decision

Add one closed `empty` response arm for the existing HTTP204 account/session deletion
responses and HTTP304 public cache responses. Generate their actual no-body result rather
than invent JSON, suppress a declared status or treat a valid304 as a transport exception.
Recommendation: adopt this arm and explicit conditional-read input below. The alternative is
an explicitly ruled transport-only304 exception plus a separate204 contract, not an implicit
bypass of full-status conformance. Neither option is adopted by this draft.

The gap is executed, not hypothetical: current mounted public handlers emit304 correctly,
while the current generated client rejects it as `undeclared API response status (304)`.
Existing `DELETE /api/v1/session` and `/api/v1/account` return204 but are unregistered.
See the API log's 2026-10-09 checkpoint and RP-483 for the exact evidence and limits.

## Specification

### BR1 — Closed descriptor and validation

Extend C19's response union with `empty {status:204|304}` (`ResponseEmpty = "empty"`).
`ContentType`, `SchemaRef`, `ContentHashHeader` and `ExactJSON` must be absent on this arm.
No other status uses it; schema/raw arms refuse204/304 instead of describing nonexistent
content. At either status there is exactly one empty arm, never competing representations.
A304 arm requires a GET operation with a declared200 schema/raw representation. This does
not add HEAD,205, generic media types or arbitrary HTTP-response/header descriptors.

Add `ValidateEmptyResponse(operationID, status, data)` which admits only a registered empty
arm and exactly zero bytes (nil or zero-length). Whitespace, `{}`, `null`, an undeclared status
or an unknown operation refuse. Existing schema/raw validators never admit an empty arm.
Header semantics remain A8/C16 and the raw representation's existing C19 contract; this
body validator does not claim to validate headers or know the client's stored representation.

HTTP204/304 have no content or trailers under
[RFC9110 §15.3.5](https://www.rfc-editor.org/rfc/rfc9110.html#section-15.3.5) and
[§15.4.5](https://www.rfc-editor.org/rfc/rfc9110.html#section-15.4.5). A304 reuses metadata
for the stored representation; its resource digest is not the digest of an empty payload.

### BR2 — Generation and explicit conditional reads

OpenAPI emits an empty response's description with no `content` member, following the
[OpenAPI3.1 response object](https://spec.openapis.org/oas/v3.1.1.html#response-object).
TypeScript emits `{status:204; body:undefined}` / `{status:304; body:undefined}` alongside
the existing error and content alternatives. Preserve `headers` and Fetch's `ok` value:
204 is true;304 is false. Callers discriminate by status, not a fabricated200 or empty DTO.

Operations declaring304 gain optional `ifNoneMatch?: string` input and the corresponding
optional OpenAPI `If-None-Match` header parameter. Its exact accepted grammar is one quoted
lowercase SHA256 ETag, `^"[0-9a-f]{64}"$`, matching the current server's strong-ETag output
and exact-match behavior. Invalid values refuse before fetching. Operations without304
reject this input if supplied, even by an untyped caller. This is not a generic header map.
No weak tags, wildcard/list matching or change to server-side conditional parsing is added.

Dispatch selects a declared empty status without requiring a content type, reads its
available body bytes and refuses any nonempty result. It does not JSON-parse them or hash
them against the stored resource's hash. Body-read/network errors propagate unchanged;
undocumented statuses still refuse. Conformance tests inspect direct validator/controlled
transport bytes too: a real HTTP stack can discard prohibited content before Fetch sees it.

The SDK stores no cache and makes exactly one fetch. No automatic refetch, cached-body merge,
token renewal, credential clearing or replay of deletion is introduced. If a browser cache
already resolves revalidation to200, normal declared200 dispatch remains unchanged. A caller
receiving304 owns its stored representation;304 alone proves no bytes or digest match for it.

### BR3 — Existing routes and compatibility

After adoption, API Foundation registers existing account/session DELETE handlers behind the
same authentication, account limiter and failed-auth IP limiter. Preserve absent-body/empty-
object request admission and every current error/status; do not tighten a legacy request
merely to fit generated types. Generate calls using the same empty-object request convention
as existing account lifecycle calls. No player-facing deletion/logout UI is authorized here.

Every currently mounted cached public GET explicitly declares304. Preserve its current ETag,
Cache-Control, request-ID echo, optional raw resource hash and zero rate-token charge, including
when the uncached budget is exhausted. Unknown/private/error replies do not become cache hits.
Future catalog registration must also declare its already-required304; this does not bypass
the held C18 descriptors or authorize mounting a partial catalog response.

Keep the current canonical response pin encoding (`204:empty::`, `304:empty::`). Adding304
and the previously unregistered operations is additive; removing/changing a pinned empty arm
fails compatibility. Conditional input is derived from that pinned304 arm, not a second
operation list. Preserve all prior operation/schema pins; no bulk reset, generator-version
bypass, `/v2` rewrite or JSON wrapper for existing204 is permitted.

## Scope and deviations

This explicitly amends C19's two-arm union; it is not a routine unruled implementation fix.
No gameplay, save/replay, retention, deletion semantics, provider policy, cache ages, rate
budget, public/private identity boundary or owner-authored copy changes. D-008/D-009/D-015
rights decisions and browser renewal/recovery remain separate. No other design deviation.

## Acceptance and closeout

1. Registry/validator controls admit204/304 only as above and reject invalid statuses, every
   forbidden descriptor field, duplicate alternatives,304 without GET/200, nonempty bodies,
   unknown operations and cross-kind validation. Existing schema/raw populations remain green.
2. Actual-registry OpenAPI/TS generation has no empty-response content schema and exact status/
   undefined association. Generated calls exercise204,304 and unchanged200/raw/error controls,
   optional conditional headers and invalid/misplaced input, with one fetch and no retries.
3. Real Postgres HTTP account/session deletion proves actual204, unchanged authorization,
   request admission, limiter/errors and persisted revocation/deletion effects. Reuse current
   fixtures; registration is not a data-erasure/privacy completion claim.
4. The complete mounted public registry exercises200→304, exhausted-budget304, changed-content
   invalidation and unchanged metadata/hash/error behavior. Check direct response bytes and
   generated dispatch; controlled reader tests do not substitute for the existing DB evidence.
5. Compatibility admits the additive registration and rejects removal/change of a pinned empty
   response; `make api-check`, affected cold Go/vet, client/types/build and relevant actual
   integration populations pass. Record the exact range for cross-party review before acceptance/
   archival. No full API/1.0 claim follows while dictionaries/catalogs/other obligations remain.

## Open question and changelog

- Owner adoption of BR1–BR3, followed by designated normative review, is required before build.
- 2026-10-09: source-grounded proposal; current SDK304 refusal and six mounted cache populations
  executed. No product code, generated artifact, compatibility pin or handler changed.
