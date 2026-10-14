# RFC: Original Artifact Bytes in Catalog Transport

- **Status:** draft — engineering proposal; adoption and designated review pending. Not implementation authority.
- **Author:** Codex
- **Created:** 2026-10-14
- **Design refs:** `design/06-tech.md` (versioned data, cross-runtime parity); API Foundation's public verification/transparency contract.
- **Parent / amends:** API Foundation C12/C17/C18's embedded catalog representation and HTTP lexical admission. Preserves C2 and C19.
- **Planning:** existing `planning/api-foundation/` plan/log.

## Outcome and evidence

Preserve the declared `{name,sha256,json}` shape and original artifact hashes without silently
narrowing owner grammar. The actual formatted category-source probe proves `json.Marshal` with
`RawMessage` compacts embedded bytes (RP-597). A second real-owner probe admits a duplicate
`schema_version` candidate through the Category loader and structural C18 descriptor, while
registry HTTP admission rejects it (RP-598). Neither probe claims a deployed catalog defect.

The existing SDK parses schema replies and discards their original byte representation. Parsed
objects, canonical reserialization, or recalculated child digests cannot replace source identity.
The implemented internal bundle verifier checks exact bytes/digests/framed identity, but does
not resolve wire encoding, epoch authorization or owner semantics.

## Proposed contract

### B1 — Closed owner reference, not arbitrary JSON

Extend C17 with one `artifactRef` node carrying only a named structural `Ref`, literal
`ArtifactName` and explicit `MemberPolicy` (`unique` or `last_value`). Each catalog owner exports
its member policy with its complete descriptor; composition cannot guess or override it.
`unique` means decoded member names are unique recursively; `last_value` retains the owner's
existing JSON last-value behavior. Both still require valid UTF-8, one valid JSON value and
the exact referenced structure. Owner semantic loaders remain mandatory.

Construction permits this node only as the required `json` child of the exact C18 artifact
arm whose required `name` is that same singleton literal and whose required `sha256` is the
prefixed digest. Ref targets must exist, with the existing cycle/depth checks. All other node
fields, nested artifactRef targets, unnamed families and conflicting owner declarations refuse.
Only public, unauthenticated catalog schema responses may reach these arms. Requests, query/path
parameters, error schemas, private/authenticated responses and raw responses cannot reach them.
No generic JSON kind, open map, generic duplicate exemption or new response carrier is added.

Registered HTTP validation selects the complete exact artifact arm, then applies its owner's
member policy only inside that `json` value. Every envelope/metadata field and every unrelated
operation retains the current recursive uniqueness rule, including escaped duplicate names.
Invalid/ambiguous arms cannot bypass envelope checks. Structural validity is never inferred
from a matching hash. Existing schema literals follow the same descriptor-aware rules.

### B2 — Recoverable original-byte framing

The catalog writer first checks the complete source bundle's existing framed hash and every
advertised child digest, then the complete registered owner descriptors. It writes each source
directly after the `json` member's colon and immediately before the following member delimiter,
without adding, trimming, escaping, compacting or reserializing any source byte. In particular,
all whitespace in that interval belongs to the source. Emit the surrounding envelope
deterministically with no extra whitespace in that interval; keep artifacts name-sorted.
Validate the finished body before writing headers/body so refusal cannot publish partial data.

A consumer locates the exact declared `json` intervals in original UTF-8 response bytes with
a JSON-aware scanner (strings/escapes, nested arrays/objects, commas and braces are not regex
delimiters). Recover the entire colon-to-delimiter interval, including surrounding source
whitespace, and verify its digest before trusting the value. Missing/duplicate names, extra or
missing artifacts against the claimed bundle hash, malformed JSON, wrong owner shape/policy,
and tampered bytes/digests refuse. Do not canonicalize a failed interval to make its digest pass.
Envelope pretty-printing that changes that interval changes evidence bytes and must refuse.

Retain the existing whole-response ETag/cache/limiter conventions; source digests are not the
response ETag. Acceptance of a constants hash still comes from the epoch repository, never from
the verifier or current filesystem data. No formula fallback, mint, storage rewrite or new HTTP
field is authorized by this amendment.

### B3 — Generated consumer and compatibility

Generate the referenced OpenAPI/TypeScript shape from the same owner descriptor, with closed
extension metadata identifying the literal artifact family and member policy. Do not substitute
an unconstrained object or binary field. Clone this metadata with existing immutable exports.
Compatibility pins retain it; a changed family/ref/policy in an already-used arm refuses unless
the ordinary versioned API contract explicitly authorizes the change. New arms retain C2's
existing response-widening rules. Existing unannotated generated files remain byte-identical.

For only these declared public catalog success responses, generate an additional SDK-result
`artifactBytes` collection of `{name,sha256,bytes:Uint8Array}` recovered from the original body.
It is local result metadata, not a new wire field. Do not retain private/error-response bytes,
add storage, retries or logging, or claim this adds general SDK DTO validation. The extraction
checks exact intervals and advertised digests; the catalog consumer verifies complete framed
identity and invokes the existing TypeScript owner/replay loaders on those exact bytes.
No handwritten operation-ID/family allowlist or JSON.stringify reconstruction is permitted.

## Acceptance evidence

1. Construction refuses every prohibited placement, malformed owner declaration and reference
   ambiguity; caller/export mutation cannot change built descriptor or member-policy metadata.
2. Actual registry validation preserves ordinary HTTP duplicate refusal and rejects duplicate
   metadata, escaped duplicates and wrong families/shape. Real last-value owners accept their
   original duplicate-member candidates; unique owners refuse their candidates. Neither policy
   waives existing semantic validation. The RP-598 diagnostic becomes a compatibility proof.
3. Writer → original HTTP bytes → generated SDK → existing TypeScript catalog/replay loaders
   proves current and stored bundles without source/hash changes. Cover leading/trailing
   whitespace, UTF-8, escapes, delimiter-like strings and nested composites. RawMessage
   compaction, interval trimming, digest tampering, missing/extra artifacts and invalid owner
   shapes must fail; duplicate metadata can never select an exemption.
4. The complete registered union covers the authoritative served set. Small fixture unions are
   construction tests, not permission to mount a partial real endpoint. Owner exports, integer
   representation and lawful formula mint remain their separately accepted prerequisites.
5. Run affected cold Go/vet, actual generated client/browser checks, `make api-check` and the
   live server/Postgres public verification journey before closeout. Record exact reviewed range
   and unavailable boundaries. No API, hosted CI or 1.0 completion follows from this draft.

## Adoption boundary

This is a recommended engineering amendment, not a gameplay choice. Acceptance must explicitly
authorize B1's owner-scoped lexical rule, B2's byte framing and B3's local generated result
metadata after designated normative review. No production implementation or accepted-body edit
is made under this draft. Global HTTP uniqueness and the closed current C17 union remain in
force until then. Independent implementation review and full C18 acceptance remain separate.
