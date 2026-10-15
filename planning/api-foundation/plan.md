# API Foundation implementation plan

RFC: `rfc/api-foundation.md`

Current A3/C6/C16 correction (2026-10-15, RP-641): public errors now share the
existing per-IP response budget instead of bypassing WriteCached. Twenty-one
old mounted-router cases reproduce the gap. Cached and uncached writers share
one admission decision; authorized304 remains free, with limits/proxy attribution
unchanged. The owning log records final affected checks and the review range after
`61fa96cf`. This is not a pre-reader work bound, complete API acceptance or a
resolution of held contracts. Continue accepted construction and consolidated review.

Current C14 raw-source correction (2026-10-15, RP-640): the actual repository now
refuses invalid UTF-8 genesis instead of admitting it through json.Valid. The old
real-Postgres source returns corrupt bytes with nil error; corrected14-case source
population and valid-Unicode byte/hash controls pass. Existing mounted error/refusal
tests and cold affected packages/vet pass. Raw transport/archives are not decoded or
normalized; no synthetic seed is claimed as a naturally verified public run. The
owning log holds the range after `a0029bd5`, pending designated review; no public
catalog/complete-loop, kernel, schema or release promotion follows.

Current HTTP encoding correction (2026-10-15, RP-639): Go wire admission now
refuses malformed UTF-8 before decoding, matching the existing client boundary.
Thirty-six old schema/literal failures and two real Postgres/session-route401
refusals reproduce the gap. Corrected mounted routes return exact400 invalid/body,
preserve state/credentials and subsequently issue authentic sessions. Valid Unicode/
escapes, shared owner decoding and raw byte identity remain; cold affected Go/vet,
selected real-DB routes, client encoding and unchanged generation pass. Owning log
holds the range after `455b7108`, pending Claude. No schema/pin/kernel, recovery
policy, owner proposal or full API/CI acceptance change.

Current C2 array-arm correction (2026-10-15, RP-638): ten old-checker regressions
approve changes that invalidate actual old replies, including mixed-element loss
hidden by invalid empty/singleton inputs. Reuse the existing structured intersection
proof for array arms, with side-specific references and whole-array coverage.
Safe growth, existing/third-arm overlap and162 finite-language comparisons agree
with runtime validation. Cold affected packages/vet and byte-identical generation
pass; the owning log records the range after `4f29ee55`, pending Claude review.
No schema DSL, endpoint/DTO/pin, kernel or held proposal changed/adopted; full API
construction, general inclusion and prior review debt remain open.

Current C2 object-domain correction (2026-10-14, RP-626): twelve old-checker actual-byte
regressions reproduce optional-field, required-demotion, enum and integer growth invalidating
exclusive object replies. A side-specific intersection-coverage proof now refuses those while
preserving unchanged/safe growth and already-invalid overlap, including third-arm coverage.
242 complete finite-domain comparisons agree with actual runtime validation; named/nested and
exact integer controls execute. Owning log holds final checks and the range after `5210d561`,
pending Claude. Unknown intersections fail closed; no general inclusion/full C2/API claim,
DTO/pin/DSL/kernel change or owner-draft adoption follows. History-exception ruling remains held.

Current C2 integer-domain correction (2026-10-14, RP-618): widening existing integer ranges
can invalidate old replies under exclusive union matching. Fifteen old-checker byte failures
reproduce across both bound directions, wrappers, >2^53 and int64 extremes. Compare exact
old/current interval partitions; no floats, new budget or value-space brute force. Safe growth,
old-invalid/new-only overlap, named/unbounded domains and a complete small runtime population
agree. Cold affected packages/vet and unchanged generation pass; the log owns the range after
`a5b4749c`, pending Claude review. Object cross-arm/full C2/API/release remain open.

Current C2 string-domain correction (2026-10-14, RP-617): widening an existing enum arm can
invalidate an old reply without adding any alternative. Six old-checker byte regressions
reproduce; the corrected check counts actual exclusive matches for old/new enum literals,
following references and the runtime format grammar. Old unbounded strings are included.
Safe growth, previously invalid overlaps and format-inapplicable literals remain accepted.
Affected packages/vet and byte-identical generation pass; the log owns the range after
`dca0bafd`, pending Claude review. RP-618 now covers integer cross-arm interactions; object/full C2/API
acceptance remain separate; no wire, baseline, owner, kernel or release promotion.

Current C2 correction (2026-10-14, RP-616): appended response-union arms now require a
disjointness proof against the old arms. Six old-checker regressions pass metadata compatibility
while making actual formerly valid replies fail runtime validation. Corrected refusals and
disjoint kind/enum/reference/tagged-object/exact-boundary controls pass, including empty-array
overlap. Cold affected packages/vet and unchanged generated artifacts are recorded in the
owning log; the range after `ea194d8d` needs Claude review. No endpoints, DTOs, pins, schema DSL,
kernel or owner proposal is changed/adopted. RP-617 now covers existing string-enum interactions;
other cross-arm domains remain work. This batch is not complete C2/API acceptance.

Current A5/C17 admission correction (2026-10-14): path parameters now require
unique names and resolve named-reference chains before enforcing their existing
scalar kind boundary. Ten old-source refusals fail; corrected refusals, valid
reference/generator/mounted-dispatch controls and four affected native packages
pass, with unchanged production API artifacts. The owning log records the exact
range after `3c1da2b0`, pending designated review. No endpoint, URI-value grammar,
RP-562 overlap policy, C18 family count or held proposal is changed/adopted.

Current C18 byte boundary (2026-10-14): `VerifyPublicCatalogBundle` verifies advertised child
digests and complete framed identity while preserving detached original sources. Current
manifest19 and corruption controls execute; this internal primitive is not mounted transport.
RP-598's real category-loader candidate additionally proves owner last-value admission conflicts
with HTTP's recursive uniqueness check. The bounded `rfc/api-artifact-byte-transport.md` proposal
specifies owner-scoped lexical validation and byte recovery; it remains draft. Do not implement
an encoder that silently narrows valid owner artifacts or claim JSON.parse recovers original
bytes. Adoption/review of that engineering boundary is separate from gameplay choices.

Current C18 composition (2026-10-14): `ComposeCatalogSchemas` constructs the closed family-to-
owner union/bundle and refuses incomplete or ambiguous input. Actual epoch8 manifest19 plus
real exports15 returns only the four missing owners, with no partial schemas. Existing exports
compose/validate together; fixture contexts exercise closed payloads, references, immutable
exports and the actual generators. No endpoint/production registry is mounted or family count
promoted. RP-597 additionally executes a real formatted-artifact diagnostic: schema-valid
RawMessage marshaling destroys original-byte identity. Implement a byte-preserving writer/
consumer after reconciling RP-598, rather than canonicalizing stored sources or calling parsed
JSON proof. Ordinary source-verification construction remains authorized under C12/C18.
The latest owning log holds the complete new range after `ae0af284`, pending designated review.

Current C18 construction (2026-10-14): Categories, Achievements, Meters, Opportunities,
Doctrines, Fiscal, Soul, Curriculum, Prestige, Pitch, Pets, Factions, Guilds, Commons and Minigames export fifteen of nineteen pinned-family descriptors. Four, the complete response
union/catalog HTTP reader and public re-verification remain open. All semantic loaders remain
mandatory; no open JSON escape, content mint or false metadata-only kernel bump. Catalog
admission repairs carry genuine versions0.3.191–0.3.208 and remain pending designated review.
The append-only owning log holds each old failure, final checks, limits and exact review range;
latest remaining-Economy/shared-condition repair range is after `c0b71434` through its containing source/test/docs/record
commit. Economy's closed cost-curve, resource/cap, generator/role and manual/offline components are constructed; its full versioned artifact,
including the reused condition grammar and other row families, is still unexported and uncounted.
Continue other owners within lawful scope; unchanged metadata-only additions still need explicit
scope resolution, and the integer-representation proposal remains unadopted.

The narrowly verified descriptor-only/kernel-neutral lane and owner schema companions were
presented to Marco on2026-10-14; no reply/adoption has been received. Presentation/preselection
does not authorize a guard exemption, namespace relocation or new package boundary.

The concrete `rfc/api-artifact-schema-ownership.md` draft now specifies the single Route-owned
companion and narrow independently reviewed neutral delivery lane, with the history-scope
guard draft as an explicit prerequisite. It is not adopted authority. RP-593/RP-594 correct
the remaining Economy original root/row/arm admission and shared Route condition admission;
they do not count either complete schema family or implement the proposed package boundary.

Current generator correction (2026-10-14, RP-570): APIError enum exports now follow validated
named references in the envelope and scalar fields. Two sorted actual-registry regressions
lose metadata before correction; declared/unknown runtime responses discriminate independently.
Cold generator packages/vet and `make api-check` pass with all production-generated bytes
unchanged. Owning log records the invalid initial fixture and complete range after `180f25e2`,
pending Claude review. This does not adopt new descriptors or complete C17/C18/API acceptance.

Current bounded correction (2026-10-14, RP-569): generated detail-enum metadata now binds the
shared operation adapter to the existing APIError descriptor. Five old Garden/Soul refusals
fail; the corrected affected units312/native Chromium-WebKit702, cold generator packages/vet,
types/build and deterministic generation pass. Existing valid typed rejections, minigame
ExactJSON checks, server bytes and retry/recovery policies remain unchanged. Owning log records
the whole range after `e56da0ec`, pending Claude review. This is global error-shape conformance,
not complete per-operation error-pair coverage, real server/SQL or whole API/CI acceptance.

Current bounded correction (2026-10-14, RP-568): generated schema-response decoding now refuses
malformed UTF-8 rather than repairing reply bytes. All36 old-client refusals reproduce; final
Node289/native Chromium-WebKit578, cold generator packages/vet, types/build and deterministic
generation checks pass. Valid Unicode/JSON escapes/BOM tolerance, body-read failures and raw
byte/hash identity remain unchanged. The owning log records exact evidence and the whole range
after `5872b66e`, pending Claude review. This is encoding conformance, not runtime DTO validation,
draft adoption, real-server/SQL or full API/CI acceptance. Continue accepted API construction;
the existing owner-schema/wire/dictionary/bodiless boundaries remain explicit below.

Current C18 representation proposal (2026-10-14, RP-566):
`rfc/api-integer-token-encoding.md` specifies opt-in exact-safe numeric tokens without globally
widening current integers. Six additional shared boundary controls execute in Go/Node/native
Chromium/WebKit. The draft covers construction, unchanged URI/default behavior, exact
validation, immutable exports, generated metadata and compatibility; adoption/designated review
remain open, and no production DSL or owner descriptor is implemented. Owning log records
the new proposal/test-only range after `f900340d`; API-only kernel scope and full C18 remain.

Current C18 readiness/correction (2026-10-14, RP-566/RP-567): actual relevance-policy
admission preserves exact integral decimal/exponent tokens that the existing registry
integer validator refuses. A complete descriptor needs a representation-contract proposal,
not a product arithmetic ruling or a historical-byte rewrite. Separately, the accepted R12
grammar now refuses Go quoted numeric fields and TS array axes; shared old-source regressions
and freshly hashed replay-bundle checks execute in both runtimes. Kernel0.3.190 records this
real admission correction and registers both relevance owners as affecting paths. The owning
log records final checks, the existing red history guard and the complete range after
`bd0c9b97`, pending Claude review. This does not finish C18, public catalog serving or1.0.

Current bounded correction (2026-10-14, RP-564): generated response metadata now preserves
the registry's existing ExactJSON literals, and the shared minigame operation port uses them
to bind typed category/detail errors to the selected operation/status. Wrong pairs fail as
body-free transport errors; server outputs, DTOs, OpenAPI and compatibility pins are unchanged.
The owning log records executed regression/generator/consumer checks and the Firefox startup
limitation. The new range after `5cd7119f` needs Claude review; this is not general SDK shape
validation, adoption of the tenant-binding draft or full API/CI acceptance.

Current bounded correction (2026-10-13, RP-535): the generated SDK now compares
declared response media type/subtype case-insensitively, retaining exact evidence
bytes/digests and wrong-media refusal. Old-client Node/native failures and corrected
SDK/runtime/operation-port checks are recorded in the owning log; generated schemas
and compatibility pins are unchanged. The new range after `15bf18b2` needs Claude
designated review. This is HTTP conformance, not a new representation, draft adoption,
whole API/CI acceptance or a gameplay milestone. Continue the remaining accepted
generated consumers and catalog-authority work within their unresolved boundaries.

Preceding bounded correction (2026-10-13, RP-524): the generated SDK now requires
single-schema replies to carry the declared JSON media before reading their body,
matching A5/C9/C19. Twelve old-client refusals fail; eight declared-JSON controls
pass. Corrected Node/native-browser/Linux-three-engine, types/build/generation and
the whole real-service composed lane pass. Reply doubles now use JSON media without
changing payloads or assertions. The log owns the range after `9b48c9c6`, pending
Claude designated review; this is not independent approval of Codex's earlier
download/client code, adoption of drafts or complete API/public-catalog acceptance.

Current integration supplement (2026-10-12): the existing root composed lane now requires
the generated SDK to download a real queue-verified Exit fixture's immutable evidence from
its live Go HTTP server, then runs the actual TypeScript verifier with DB-pinned catalog bytes.
Receipt/event corruption, log gaps and wrong constants must refuse. This is C14 construction,
not adoption of the catalog/dictionary/bodiless drafts or a completed public download loop.
The owning log records the actual run and boundary; designated review remains required.

Current bounded fix (2026-10-12, RP-506): registered dispatch now rejects exact `.`/`..`
path components before fetch; ordinary dotted/escaped names and dot-valued queries stay
unchanged. Four old-source regressions fail in Node/Chromium/WebKit; final affected
Node/native-browser and Linux three-engine consumers, cold generator/vet, types/build
and deterministic regeneration pass with unchanged OpenAPI/compatibility pins.
This is A5/C1/C9 conformance, not a schema/HTTP contract amendment or independent approval
of Codex's earlier download/client work. The owning log records the finished outcome.

Current runtime correction (2026-10-12, RP-502): shared conditional caching now parses the
complete If-None-Match field and uses HTTP weak comparison for strong/weak tags, lists,
repeated lines and wildcard. Nine old-source policy cases and all six mounted operations fail;
corrected affected packages and actual Linux-amd64/Postgres public download/privacy checks
pass without changing limiter values, evidence bytes, registry or generated artifacts.
Malformed conditions cannot obtain a cache hit; wildcard private/unknown runs still404.
The owning log records the test-fixture correction, exact checks and pending Claude review.
RP-483 remains an unadopted schema/SDK prerequisite; no full API/CI acceptance follows.

Current specification review (2026-10-12): Claude's A6/AC4 body correction
`266b8d8c^..266b8d8c` is designated-APPROVED by Codex. It faithfully reconciles
the already-adopted C9/C18 requirements; no new contract or implementation is
approved. The epoch8 manifest has nineteen artifact families at both the original
boundary and current HEAD. Exact owner descriptors, generated-only caller/lint
completion, historical formulas, missing response arms and later implementation
reviews remain open. The owning log records this documentation-only review.

Current configuration supplement (2026-10-12, RP-497): explicit previous cursor pairs now
refuse duplicate current IDs/values in direct composition, matching Deployment DP3. The
wholly absent C20 first-deployment alias remains valid. Old-source duplicate-value assertions
fail; corrected cold packages/vet, unchanged API generation and selected real Linux-amd64/
Postgres composition/download checks pass. This supplement and the preceding RP-496 repair
need Claude review; neither resolves RP-495 or completes public re-verification/API acceptance.

Current router correction/review (2026-10-12, RP-495/RP-496): direct composition now refuses
previous cursor material without its ID instead of treating it as the first-deployment alias.
Old-source regressions fail; corrected cold affected packages and selected real Linux-amd64/
Postgres composition tests pass. This Codex correction needs Claude review. The full original
router `731bb3e7` and board `4c316c4c` ranges now have designated CHANGES REQUIRED verdicts in
the log, completing those original reviews, not approving their later corrections. RP-495
routes the already-disclosed fixed-name/rotation-ledger conflict to the API/Deployment authors;
no fixed-name policy, overlap or ledger rule was changed. Next: remaining API construction and
later implementation reviews under accepted authority; dictionary/bodiless/history drafts remain
unadopted. Full API acceptance, clean-host deployment and nine-tier1.0 remain open.

Current privacy-evidence correction (2026-10-12, RP-494): the structural gate now checks
contract-derived field sets and complete nested paths, including APIError, rather than relying
only on suspicious-name fragments. Three actual valid-registry negative fixtures fail against
the old gate and pass after repair; compatible optional additions prove API compatibility is
not privacy approval. No runtime DTO, handler or published field changes. The original Route
Registry reader batch `09d12987` is now reviewed CHANGES REQUIRED for this evidence gap;
the log owns executed checks and the test-only correction's pending designated review.
Whole API acceptance remains separate from the original review dispositions above.

Current evidence correction (2026-10-12, RP-493): the board cursor's changed-epoch test now
uses a valid second epoch, verifies each changed filter succeeds without a cursor, and requires
the old cursor to fail before page retrieval. Deliberately disabling epoch binding passes the
old test and fails the corrected one; production source is restored unchanged. Cold affected
packages, declared Postgres reader tests and selected Linux-amd64 composed-server privacy/socket
integration pass. The log owns the targeted original evidence finding and pending test-only
corrective review; full original router/boards/routes dispositions are recorded separately above.

Current bounded correction (2026-10-12, RP-492): query string validation uses JSON escaping,
not Go literal escaping, while retaining malformed UTF-8 refusal and exact descriptor checks.
The old-source regression rejects six descriptor-admitted strings and an enum member; affected
cold Go/vet, generated-client tests and unchanged API generation pass after correction. The log
records CHANGES REQUIRED on Claude's original query-descriptor commit `32560c4b` separately
from this Codex repair's pending designated review. The original epoch-reader commit `e63796d5`
now also has a CHANGES REQUIRED verdict: its unchanged SQL/keyset/changelog tests pass against
real Postgres, but its private Account import (RP-390, separately repaired) and missing304
declaration (RP-483, draft amendment) prevent whole-batch approval. Router/boards/routes have their
own later verdicts; no adoption of the dictionary/bodiless/history drafts follows.

Current prerequisite (2026-10-09, RP-483): `rfc/api-bodiless-responses.md` proposes a closed
empty204/304 arm and explicit conditional-read input. Fresh existing six-operation cache
checks pass; the actual generated SDK rejects304 and has no DELETE lifecycle operations.
Owner adoption/review are pending. Do not implement this third response arm under unamended
C19, label204 JSON, or treat the proposal as rights/retention or browser-cache policy approval.

Current bounded correction (2026-10-08, RP-476): generated SDK dispatch selects the declared
media representation when one status has alternatives. Single-representation schema decoding,
raw digest checks and mounted endpoint formats remain unchanged. Old-client failures, affected
checks and the real main player journey are recorded in the log; designated review is pending.
Related OpenAPI header aggregation RP-477 is now locally corrected in a separate coherent
batch: all raw headers survive, shared names combine case-insensitively, and conditional
requirements name their media type instead of applying to an unrelated schema alternative.
Six actual-registry generation populations and unchanged mounted API generation pass;
runtime verification is unchanged. Its designated review is also pending.

Current bounded correction (2026-10-08, RP-473): operational policy admission now requires
all exact, non-null root/nested fields before struct decoding. Missing/null proxy hops no
longer silently select zero; explicit zero and escaped canonical names remain valid.
The owning log records old-source failures and final consumer checks. No configured values,
generated contract, proxy topology or limiter semantics changed; designated review is pending.

Current bounded correction (2026-10-08, RP-449): duplicate-member rejection belongs at
registry HTTP admission, including exact-response literal construction. Shared descriptor
validation restores the existing Go/TypeScript formula-artifact last-value contract without
rewriting stored bytes or changing fixtures. The owning log records regression, consumer and
Postgres results and the pending designated-review range; this is not full API/CI acceptance.

Bounded validator correction (2026-10-08, RP-439): raw duplicate object members now refuse
at HTTP admission before schema map decoding, including escaped/nested names. The old-behavior regression,
affected packages/vet and byte-unchanged API generation pass locally. No handler output,
DSL extension, request middleware or full conformance claim; designated review remains.

Current checkpoint (2026-10-08): the historical catalog database source is implemented and
locally verified against real Postgres; accepted identities return exact stored artifacts,
not newer files, and corrupt/unaccepted evidence refuses. The future formula artifact now has
a shared version-14 model and owner schema, checked against actual byte-identical generation
and malformed-input controls. Both replay loaders and Go resolution now accept optional stored
formulas with exact integer grammar/identity; real Postgres proves old/new coexistence and
malformed-version refusal. This does not mount the catalog
HTTP reader or mint product formulas. Categories now exports the first descriptor for the
nineteen pinned families; five actual current/historical/candidate fixtures and malformed-input
controls pass. Its existing loader retains semantic-policy validation. The newer checkpoints
above add Achievements, Meters, Opportunities, Doctrines, Fiscal, Soul, Curriculum and Prestige. The other ten
owner descriptors and complete union remain open. RP-383/RP-384 compatibility corrections pass
locally; generated OpenAPI/TypeScript/pin bytes remain unchanged. Remaining owner descriptors,
remaining route/client registration, third-party verification, 304 metadata and designated
reviews stay open; not full API acceptance or green hosted CI. Latest log owns exact evidence.
Public reads now use the shared error descriptor from `publicapi` instead of importing Account;
the direct private-repository import regression and affected packages pass. This does not close
transitive/runtime privacy. The seventeen remaining artifact-owner packages now fall
under kernel-watched prefixes at that checkpoint; the remaining count is now ten after the
newer owner exports above. Scope remains constrained:
API-only additions require an explicit scope resolution or a
genuine accepted semantic change, never a false version bump or file move to evade the guard.

Existing session creation/refresh now mount from the registry and generate explicit calls, with
exact error alternatives and the existing shared IP limiter. Real Postgres creation/rotation/
reuse/lost-reply populations, affected client checks and the whole composed journey pass. The
explicit C1/C2 additive pin refresh protects both operations and their statuses, while capturing
already-registered Garden/optional Garage shapes. No automatic browser renewal or retry policy
is implemented; its draft decisions and full API acceptance remain open.

- [x] Reconcile C1–C17 into the active A1–A8 contract.
- [x] Implement the closed schema-descriptor DSL and operation registry foundation.
- [x] Implement authenticated keyset cursors and normalized board variables.
- [x] Implement strict operational API config and shared request/cache/limiter middleware.
  - [x] Resolve C20's limiter/proxy/key-ID/request-ID literals.
- [ ] Register public DTOs/readers and raw verification evidence endpoints.
  - [ ] Export and compose C18's ruled owner artifact descriptors.
    - [x] Export Prestige's existing v1 descriptor with finite payout keys and existing duration
      representation bound (`TestPrestigeOwnerSchemaAdmitsCurrentStoredAndRetunedPoliciesWithoutChangingBytes`).
      Explicit zero and data retunes remain valid; semantic/composition checks stay mandatory.
      No changed formulas, new gameplay ceiling, public endpoint or designated approval claim.
    - [x] Export Curriculum's current v1 descriptor with ruled branch/starter pairings;
      current/stored/candidate data and paired bindings pass
      (`TestCurriculumOwnerSchemaAdmitsCurrentStoredAndCandidateWithoutChangingBytes`). Generic
      Reputation starters remain unchanged; valid data retunes pass. Semantic loader remains
      mandatory; no new content, public endpoint or designated approval claim.
    - [x] Export Soul's existing closed v1 descriptor with heartbeat/presentation fields;
      current/stored bytes and tracked copy bindings pass
      (`TestSoulOwnerSchemaAdmitsCurrentStoredAndRecoveryFixtureWithoutChangingBytes`). Explicit
      zero/false and original-byte uniqueness remain, semantic loader is mandatory; no source mint,
      public catalog endpoint or designated approval claim.
    - [x] Export Fiscal's existing v1 policy/row descriptor; current/stored bytes and corresponding
      Economy bindings pass (`TestFiscalOwnerSchemaAdmitsCurrentStoredAndFixtureWithoutChangingBytes`).
      Exact-name/non-null admission repair does not retune clocks, costs or factors. Semantic loader
      stays mandatory; no complete union/public endpoint or designated approval claim.
    - [x] Export the existing Doctrine identity-only v1 descriptor; current/stored bytes and
      corresponding Route bindings pass
      (`TestDoctrineOwnerSchemaAdmitsCurrentStoredAndFixtureWithoutChangingBytes`). Exact-key/
      non-null admission repair preserves explicit zero; semantic loader remains mandatory.
      No new choices/effects/content or complete C18/public HTTP/designated approval claim.
    - [x] Export the existing Opportunities v1 descriptor with four closed effect arms;
      actual current/stored catalogs and their corresponding Economy bindings pass
      (`TestActivePlayOwnerSchemaAdmitsCurrentStoredAndAllEffectArms`). A9 admission correction
      and explicit semantic-loader separation accompany it; no union/public serving or review approval.
    - [x] Export the existing Meters v1 descriptor with closed input and decay arms;
      current/stored bytes and registered signed/slot domains pass
      (`TestMeterOwnerSchemaAdmitsExistingArtifactsWithoutChangingBytes`,
      `TestMeterOwnerSchemaInputAndDecayDomains`). Semantic loader stays mandatory;
      no complete union/public HTTP or designated approval claim.
    - [x] Export the existing Achievements v1 closed structural descriptor, with finite
      depth-specific conditions and all proof arms; current/stored source bytes pass
      (`TestAchievementOwnerSchemaAdmitsExistingArtifactsAndFixtureWithoutChangingBytes`,
      `TestAchievementOwnerSchemaAllConditionAndProofArms`). Semantic loader stays mandatory;
      no full union, public HTTP, minted content or designated approval claim.
    - [x] Export the existing category catalog's closed v1 descriptor, with all five canonical
      row bindings; actual historical/current/candidate catalogs and 26 negative cases pass
      (`TestCategoryOwnerSchemaAdmitsStoredAndCandidateCatalogsWithoutChangingBytes`,
      `TestCategoryOwnerSchemaRejectsUndeclaredAndWrongRowShapes`). Loader semantic checks
      remain mandatory; no complete union or HTTP endpoint claim.
    - [x] Extract the future formula artifact's shared model and closed version-14 owner
      descriptor; actual generator preserves bytes and validates before writing
      (`TestPublishedFormulaModelAndDescriptorPreserveExistingBytes`, populated/refusal
      controls and `make formulas-check`). No product formula mint/HTTP registration.
    - [x] Extend both replay loaders and Go resolution for optional stored version-14 formulas,
      using Go-generated client metadata and original bytes. Shared controls, existing replay
      tests and `TestStoredFormulaReplayIntegration` pass; product mint/public HTTP stay open.
  - [x] C3/C11/C12 database source: accepted-hash authorization, exact bytes/digests/order,
    bundle identity validation and real-Postgres historical/refusal cases
    (`TestPublicCatalogIntegration`; no HTTP/formula mint or replay-verification claim).
  - [x] Resolve C19's raw response arm.
  - [x] C14 repository source: board-record authorization, exact immutable bytes/pinned metadata,
    hash consistency and real-Postgres refusal tests.
  - [x] C14 HTTP handlers/manifest descriptors and a real verified-run download round trip
    (TestComposedGameserverExitVerificationAndBoardIntegration; pinned catalogs supplied from DB).
  - [x] Historical C14 evidence after a distinct database epoch/hash rollover: the same
    verified run downloads/reverifies with freshly loaded original DB artifacts, retaining
    receipt-corruption refusal (TestComposedGameserverExitVerificationAndBoardIntegration).
    Existing composed lane requires the rollover and repeats generated-SDK/TS checks.
    No process upgrade, public catalog serving, release artifact or designated approval claim.
  - [ ] Complete the third-party loop using public catalogs and the TypeScript verifier;
    same/historical-epoch downloads with DB-supplied catalogs do not close this requirement.
- [x] Generate OpenAPI 3.1 and TypeScript from the registry; add compatibility pins/gate.
  - [x] Register existing session creation/refresh, preserve middleware and exact replies, and
    generate explicit typed calls (`TestSessionOperationRegistryIntegration`, persisted refresh
    population, shared-limiter/descriptor tests, generated SDK tests and actual pin controls).
    Local implementation is not designated approval or browser-renewal completion.
  - [x] Register existing account creation/current-Founder/New-Founder routes with unchanged
    handlers and shared limiting; export their exact historical timestamp/empty-display
    descriptors. `TestAccountLifecycleOperationRegistryIntegration` and the Founder limiter
    integration execute against real Postgres; compiler negatives, generated calls and actual
    committed-pin operation/status-removal controls pass. Empty-object SDK typing is corrected
    (RP-386). Pending designated review, not recovery UX or complete Account/API acceptance.
  - [x] Generate typed HTTP dispatch and raw/status response associations; bind real Game UI
    bootstrap/state reads (generated-api/runtime tests and real composed journey).
  - [x] Bind registered Minigame/Soul/Garden ports to generated dispatch, preserving their
    wire inputs/error classes (generated-operation-ports, minigame/soul-recovery-surface,
    garden-refresh-browser tests; real composed Pitch/default-Garden journey).
  - [x] Correct missing existing Soul error details; actual handler/mapper replies and
    authenticated Postgres errors conform (TestSoulRecoveryErrorRegistryMatchesHandlers /
    TestSoulRecoveryErrorRegistryIntegration). Full composed target remains RED at RP-377.
  - [ ] Migrate remaining HTTP callers; register missing routes/contracts and enforce C9's
    raw-fetch lint. Typed dispatch over the current subset is not AC4 completion.
- [ ] Compose the public router, full-registry privacy/conformance integration, canonical docs,
  independent full-range review, and archive.
  - [x] Remove the direct public→Account dependency for the shared error descriptor and guard
    production imports in both public packages (`TestPublicPackagesDoNotImportPrivateRepositories`
    fails on the old dependency, passes after extraction; alias/comment negatives discriminate).
    Transitive reader dependencies and complete runtime privacy remain open.
  - [x] Test caching and shared IP limiting through every currently mounted public operation,
    including changed-content and exhausted-budget cases (A8/C16/C20).
  - [ ] Resolve the undeclared bodiless 304 response arm before claiming full schema/status
    conformance. Runtime caching tests do not resolve this contract gap.

Carried dependency: historical formula serving stays unavailable until a protocol-compliant mint
stores formula bytes in the artifact set; no current-formula fallback is permitted.

## 2026-08-21 corrective lane — exact schema-response bytes

Authorized by accepted A4/A5 and split from rejected Q-002 range `34d04a5..dfaeafe` after
designated finding B (`ba8ca65`).

- [x] Add one validation-only exact-JSON arm to schema response descriptors; construction rejects
  empty, unsorted, duplicate, or schema-invalid literals and registry snapshots deep-clone bytes.
- [x] Bind the four Minigame operations' existing HTTP error mappings to operation/status-specific
  exact bytes without changing any handler output, status, category, detail, generated schema, or
  generated client artifact.
- [x] Demonstrate rejection of a schema-valid wrong category/detail pair, an appended response byte,
  invalid literal registration, and post-construction caller/snapshot mutation.
- [x] Update canonical API Foundation docs; run cold focused/root Go, vet, `api-check`, typecheck,
  and sequential Account/Gameserver Postgres populations.
- [x] Record an exact-range Codex first-filter and hand this API Foundation range plus Q-002's
  test-only remainder to Claude. Do not claim public endpoints, surface completion, or archival.

## 2026-10-07 — existing public board read failure

RP-367 locally repaired within A3/A6/C13; row-iteration regression and affected package tests/vet
pass. The normal composed lane now requires the existing real-Postgres board ranking/page case
to execute, not skip. It passes, but the whole lane is RED at the separate Pitch account-limiter
failure RP-368. Next: diagnose actual request cadence without changing limits/retry semantics;
obtain designated review of this repair. Catalog descriptors, verification readers, generated
client, privacy/acceptance and archival remain incomplete.
