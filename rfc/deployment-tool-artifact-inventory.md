# RFC: Deployment Tool Artifact Inventory and Delivery

- **Status:** draft — prior two-tool proposal reviewed through1648e792; direct upstream-Compose revision pending technical review and owner adoption; no implementation or shipping authority
- **Author:** Codex
- **Created:** 2026-11-04
- **Parent / amends:** [Deployment Foundation](deployment-foundation.md), DP1/AC8; [Native Supervisor Profile](deployment-native-supervisor-profile.md), SP1/SP5/SP8
- **Design refs:** `design/06-tech.md` §stack; `design/00-vision.md` pillar6
- **Planning:** existing `planning/deployment-foundation/plan.md` and append-only log

## Outcome and authority

An operator receives the exact directly invoked Compose configuration compiler
used by the selected native producer, with a complete, separate inventory of its components, terms,
notices and required source. Its identity and delivery validate without a checkout
or a runtime license lookup. Application and image inventories retain their own
scopes. No version, build variant or research fixture is selected for shipping.

This supplies the missing SP1 packaging contract. It does not adopt the native
architecture/host floor, decide unresolved third-party rights, change gameplay,
install tools on the operator's host, introduce an image/service/signing key, or
authorize implementation before acceptance. Technical approval is not adoption
or a legal conclusion. Unknown terms remain blockers, not configurable waivers.

The [tool audit](../planning/deployment-foundation/native-tool-provenance-audit.md)
owns research evidence. Its independently reviewed standalone comparison
supports one directly invoked compiler for the current configuration population.
Compose's embedded Docker CLI Go components still require attribution; four
root MPL refusals cannot be labelled permissive, and copied/header scopes matter.
The separate native-linked CLI fixture is no longer a proposed bundle input.
Its unresolved provenance/rights findings remain historical evidence, rather than
being relabelled as solved. This contract does not require universal upstream
signatures or independently reproduced vendor builds.

This revision consumes the upstream publisher's binary release. The earlier
unadopted local-build alternative is withdrawn from this contract; selecting a
release-owned build later requires its own accepted producer/recipe contract.
No existing release criterion requires that unused build system. Static tool
inventory/assembly work can be authorized by adoption of this companion; changing
runtime dispatch still requires native-profile/parent adoption under SP8.

## T1 — One bundle authority and fixed tool roles

Retain release-manifest v1's sorted `artifacts` path/SHA256 list as the authority
for all delivered bytes. Add files, not a parallel manifest/ledger or a new
release identity. Existing non-native bundles are not restamped or rewritten.
The adopted native producer requires the complete tool group below; an older
validator's acceptance of extra files does not qualify as native validation.

`tools/inventory.json` is a closed JSON object with required integer
`schema_version:1` and `tools`, exactly one entry. The entry
has exactly `id`, `version`, `platform`, `binary_path`, `metadata_path`,
`sbom_path`, `notices_path`. `platform` is `linux/amd64`; `version` is the exact
nonempty ASCII upstream release version, not `latest`, a channel alias or a range.
All paths equal the following table, not arbitrary caller-selected locations:

| ID | Binary | Metadata | SPDX | Readable notices |
|---|---|---|---|---|
| `docker-compose` | `tools/docker-compose` | `tools/metadata/docker-compose.json` | `sbom/tools/docker-compose.spdx.json` | `licenses/tools/docker-compose.txt` |

The tool also has a fixed scope map at `tools/metadata/<id>.scopes.json`.
Ship `tools/inventory.schema.json` for these private records, including scope maps.
Deployment owns
its schema and generator; it is neither a public API nor a kernel/balance input.
Schema, index, metadata and all referenced payloads are ordinary hashed bundle
artifacts. Private objects reject duplicate/unknown/missing keys, wrong types,
invalid UTF-8 and trailing data; stable pretty JSON is sufficient. No OP1
operator-record canonicalization or new hash protocol is introduced here.

Binary hashes come from the manifest. Standard SPDX checksums and provenance
subjects must match those hashes; the index does not duplicate hash authority.
Binaries are regular Linux/amd64 ELF files installed `0755`; metadata/text/source
files are regular `0644`. Refuse symlinks, path escapes, missing files and orphan
files under tool-owned prefixes. New-profile readers require this closure before
bootstrap, without executing a tool to discover whether its metadata is valid.

For the newly adopted native producer, `docker_compose_version` records the
bundled compiler version, also used in its actual conformance/rehearsal. Compare
an optional leading `v` without rewriting existing bytes. This explicitly
reconciles DP1's former host-Compose prerequisite for this profile: no separate
host Compose supplies configuration compilation. `docker_engine_version` still
describes the host daemon. No external Docker CLI version is part of this group.
Source/target manifests keep their own identities; do not demand every retained
bundle have the selected producer's tool version. No historical field changes.

## T2 — Metadata and delivery catalogue

Each metadata file is a closed object with required `schema_version:1`,
`tool_id`, `version`, `platform`, `acquisition`, `source_record`, `provenance`,
`scope_map_path` and `components`. The identity values match the index;
`scope_map_path` equals the fixed T1 path for that tool.

`acquisition` has exactly `uri` and `release_metadata_path`. URI is the public
HTTPS upstream asset URL
`https://github.com/docker/compose/releases/download/<tag>/docker-compose-linux-x86_64`;
no userinfo/query/fragment or another publisher/project. The exact tag agrees
with the tool version, permitting one leading `v` for comparison. The fixed
metadata path is `provenance/tools/docker-compose/release.json`, containing the
verbatim public GitHub release response. Its `url` must be the official
`https://api.github.com/repos/docker/compose/releases/<id>` using its own positive
integer `id`, `html_url` the same
repository's `releases/tag/<tag>`, and `tag_name` that exact tag. The unique
matching asset name/browser_download_url must agree. Require draft=false
and the asset's recorded size/SHA256 to equal the actual acquired binary and
manifest. Record prerelease/mutability as publisher data, not automatic approval.

The generator obtains that metadata and actual asset through the upstream public
release API/download, verifies the bytes and retains the response before emitting
the record. Runtime checks the retained bindings offline; it does not recontact
GitHub or authenticate the publisher anew. A mutable release tag is not immutable
content: changed bytes require a new inventory and conformance review. No local
rebuild/repack can be admitted merely by choosing an acquisition kind or a URI.

`source_record` has exactly `uri` (`https://github.com/docker/compose`), `commit`
(full 40-character lowercase hex), `basis` (`binary_vcs` or `publisher_release`)
and `evidence_paths` (nonempty, sorted unique manifest-relative paths). The basis
states what was actually established; a tag/API claim or short revision is not
silently upgraded to compiled-source authentication. Full source inputs and
native/header evidence belong to the producer's reconciled component inventory.

`provenance` is an array of objects with exactly `path`, `kind`, `scope`, sorted
by unique `path`. `kind` is `in_toto` or `source_evidence`; `scope` is
`selected_binary`, `source_claim` or `related_only`. All referenced documents are
manifest-bound files below `provenance/tools/<id>/`. These roles are checked
against actual subjects and acquisition, not trusted as caller approval flags.
Related records may supply context; they cannot satisfy a selected-binary binding.

`components` is nonempty, sorted by unique `spdx_id`, one row for EVERY package in
the tool SPDX document, including the tool itself. A row has exactly `spdx_id`,
`grant_paths`, `notice_paths`, `source_paths`, `source_files`, `evidence_paths`.
Path arrays are sorted unique manifest-relative paths; `grant_paths` is nonempty.
Grants and notices live
under `licenses/tools/<id>/texts/`; delivered source under `sources/tools/<id>/`;
evidence under `provenance/tools/<id>/`. No references may escape those scopes.
Component/package association, applicability and obligations come from reviewed
source inputs and the adopted policy, never the filename alone. The top-level
`source_record.evidence_paths` follows the same tool evidence prefix.

`source_files` is an array sorted by unique `(path, member)` with exact keys
`path`, `member`, `sha256`. `path` is one of that row's `source_paths`; `member`
is null for a raw source file, otherwise an exact safe archive member path.
`sha256` is the `sha256:`-prefixed digest of the uncompressed file bytes. Every
required covered file is listed; all entries are checked, including those inside
an archive. The union across components equals EVERY regular member of every
source payload, not merely one used member per archive. Raw files have one null-
member entry. Shared entries agree on their hashes. Components with no source-
delivery obligation have empty source arrays unless delivering accounted extra
source; a reviewed obligation cannot be removed merely by emitting empty arrays.

### T2a — Required scope map and source-member closure

The scope map is a closed object with exactly `schema_version:1`, `tool_id`,
`version`, `platform`, `binary_sha256`, `source_commit`, `source_inputs`, `scopes`
and `source_members`. Identity/commit match metadata; `binary_sha256` equals the
manifest's binary hash. This is the reviewed per-version producer input, shipped
and checked by the native validator, not an optional external review note.
Scope/source-input IDs use nonempty lowercase ASCII letters/digits and `._-`,
starting with a letter or digit; SPDX IDs retain SPDX's grammar. Hashes use
manifest SHA256 syntax. Null path/member sorts before strings.

`source_inputs` is nonempty, sorted by unique `id`. Each closed row has `id`,
`uri` (exact public HTTPS acquisition URI), `sha256` and `format` (`tar_gz`, `zip`
or `file`). It identifies the original immutable source/SDK/module/patch input,
not a mutable directory name. The generator verifies actual input bytes and any
module h1/Git/derivation evidence under T3. These parent inputs need not all be
distributed; source delivery is the explicit `source_members` population below.
For `file`, origin `path` and member `source_path` are null and scope prefix is
empty: the input URI identifies the original file. For archives those paths are
nonempty safe relative members. Readable notices show the input URI as well as
the original path, preserving the origin of a standalone file.

`scopes` is nonempty, sorted by unique `id`. Each closed row has `id`, `spdx_id`,
`source_input_id`, `source_prefix`, `license_expression`, `selected_expression`,
`source_required`, `grant_origins`, `notice_origins`. Prefix is an exact safe
source path or directory prefix (empty means input root); it matches that path
or descendants separated by `/`, never a similar spelling. Both license fields
are SPDX expressions. `selected_expression` selects an actual permitted branch
of `license_expression` under T4; it cannot erase an `AND` obligation.
`source_required` is boolean and must be true for a selected MPL-covered scope.
Other required source obligations cannot be set false by the caller.

Origin arrays are sorted by unique `(source_input_id, path, bundle_path)`. Each
closed origin has exactly those three keys plus `sha256`: original source input,
original relative file path, delivered raw grant/notice path and exact full-file
hash. Grant origins are nonempty. Input IDs exist; delivered paths belong to the
same tool and their bytes match origin hashes. Origins retain inline/multiple
grants rather than extracting a reassuring sentence. Each component's grant/
notice arrays equal the union of its scopes' respective delivered origin paths.
All SPDX components have at least one applicable scope; no unreferenced scopes.
The producer verifies origin path/content against the actual source input, not
just an emitted hash. Safe relative paths do not contain escapes or empty members.

`source_members` is sorted by unique `(path, member)` and has closed rows with
exact keys `path`, `member`, `sha256`, `source_input_id`, `source_path`, `scope_ids`,
`required`. The first three use the `source_files` grammar. `source_path` locates
the original input member/file; nonempty sorted unique `scope_ids` associate its
actual component/terms scopes. Each associated scope applies to that input/path.
`required` is boolean: required covered source versus deliberately delivered
extra material. The producer reconciles the complete required source population
from actual inputs and reviewed obligations, retaining required modifications;
it cannot declare coverage from an already shortened payload. Extra material
also needs an applicable supported scope. If covered source differs from its
original input, retain its actual resulting bytes and modification information;
public derivation evidence in `source_record.evidence_paths` identifies the
input/result and preparation. This describes source-delivery derivation, not
permission to replace T2's upstream executable. A release-owned build/repack
requires the separate accepted contract described in T5; no runner or local-build
mode is admitted by this source map.

The source-member set equals both the full extracted regular-file population and
the union of metadata component `source_files`. Component association equals the
mapped scopes' SPDX IDs. Every declared hash agrees. Archive duplicate names,
escapes, special files, missing/extra members and unassociated members refuse;
directories are structural only. A newly inserted README or copied header cannot
pass because some other file in the archive was accounted for.

Independent producer checks compare this map against actual source/import/native
discovery, including nested terms and additional distributed files. Runtime
validation checks the map's declared coverage against actual bundle bytes; it
does not rediscover absent build inputs or prove legal applicability. Neither
protects against replacement of the entire trusted release and its reviewed policy.

Grant/NOTICE/copyright bytes retain their complete original content, including
relative source scope and any multiple grants. `licenses/tools/<id>.txt` renders
components in SPDX-ID order, then original origins in scope-ID/input/path order
for grants and notices. Each section identifies component/version/origin and
delivered path and includes the referenced raw bytes,
adding only a separator LF when needed. The validator re-renders and compares;
a similarly named, omitted or root-only text cannot satisfy another scope.
Raw text payload names use their SHA256 without prefix plus `.txt`; their bytes
must match that name and the manifest. A known grant from a different scope does
not mask an unsupported obligation; genuine alternative grants follow T4.

Every file in tool-owned metadata/provenance/license/source roots is either a
fixed T1 output or referenced by this catalogue, with its one applicable role.
Opaque extra files are not an unexamined distribution channel. Source archives
retain member paths/content and their version/hash identities; extraction is
bounded, rejects escapes/special files and cannot execute install/build hooks.
Preserve required covered files, notices and modification information. Additional
source/test/data files are distributed only after their own terms are covered;
bundling an entire research corpus is not automatic clearance.

## T3 — Separate, complete SPDX generation

Produce one valid SPDX 2.3 document per executable. Keep the six-command/browser
application document, its existing permissive policy, website notices and all
image SBOMs unchanged in scope. Never merge these binaries into the application
list or cite an image's top-level Apache label for its embedded dependencies.

The tool document has a single DESCRIBES root for the exact executable: tool
ID/name/version/platform, SHA256 package checksum equal to the manifest binary,
and a namespace containing that binary hash. Components have distinct stable
SPDX IDs, version/source identity, applicable license expressions and source
references; DEPENDS_ON/CONTAINS relationships connect them to the tool. All
package IDs and catalogue component IDs must coincide exactly, with no orphan
or dangling relationships. A header-only `ValidateSPDX` pass is insufficient.

The producer reconciles actual binary build metadata, the matching compiler
import/source population, copied/header scopes and native build inputs. Go module
checksums, Git snapshots and any extraction/patch/generation inputs are verified
before use. Module presence is not a file-scope exemption; no function-name match
is not code absence. Legacy/GOPATH empty module metadata does not authorize an
empty graph. Native/library/header contributions require an explicit disposition,
not guessed package names or a blanket root license. Guard exhaustion, missing
input or an unaccounted contribution fails generation.

Per-version source/scope maps are reviewable producer inputs bound to the selected
binary/source identities; a hand-written package table alone cannot assert
closure. Actual discovery and generated output must be compared independently in
acceptance. Do not infer completeness from two outputs of the same broken mapper.

## T4 — Proposed tool terms policy and required source

This proposal keeps the application's existing six recognized permissive classes.
For THESE separate tool documents it proposes those classes plus MPL-2.0 with
the source/notice obligations below. This addition requires explicit owner policy
adoption; technical review alone cannot admit it. It grants no third-party rights.

Every component's released `licenseDeclared` and `licenseConcluded` must be an
accurate SPDX expression, including scoped combinations; unknown,
ambiguous or `NOASSERTION` terms cannot be relabelled MIT/Apache. Classification
examines each actual grant, not the combined notices or vendor marketing. All
obligations in an `AND` must be supported. A genuine alternative `OR` may select
a supported branch explicitly in the reviewed scope map; preserve the original
grant and record the selected branch, rather than erasing the alternative.

For MPL-covered distributed material, deliver the applicable covered Source Code
Form and license/required notices in the same bundle; source URLs alone do not
meet this project's preservation choice. This is the proposed delivery policy,
not a claim that MPL universally requires same-bundle distribution; see the
[actual MPL 2.0 terms](https://www.mozilla.org/en-US/MPL/2.0/), §§3.1–3.4.
The reviewed scope map binds each
covered component/file to its actual source version/checksum and payload.
Validate payload identities and coverage, retaining any actual modifications.
Removing/changing a covered file, not just the zip or index name, must invalidate
the tool closure. Full archives are allowed only when all additionally delivered
material's terms are also accounted for. No MPL obligation is inferred absent
because a nested directory has zero function-name or source-query matches.

Unresolved OASIS references, candidate GPL/runtime-exception applicability,
unknown native contributors or other unapproved terms STILL refuse. Those need
actual source/terms resolution and, where it changes this policy, separate
adoption; an exception's name or a package licence label is insufficient.
The existing fixtures remain research-only and are not a newly selected version,
build mode, policy exception or ready shipping inventory.

## T5 — Truthful upstream provenance

Keep vendor in-toto/SLSA records verbatim. Validate supported envelope/predicate
and subject semantics using their standard parsing rules: private closed-object
rules do not reject compatible extension fields in standard SLSA documents.
Recorded signature verification or reproduction claims require actual evidence;
neither is universally required by this RFC. Do not infer a SLSA security level.

A selected-binary record must name the actual standalone executable SHA256.
A foreign image subject, another build's code equality or a revision string
cannot supply that binding. OCI-only records are related source evidence under
this binary-release profile, not selected-binary provenance. Image inventories
retain their own existing source/image provenance requirements.

Preserve the upstream release's available SPDX/provenance statements, matching
their own published asset sizes/digests before using them. Source claims and
unsigned publisher assertions retain their evidence limits. The generated T3
document supplies its own classifications through actual component/source/grant
reconciliation; the retained upstream NOASSERTION document stays unchanged and
is not clearance. Compatible
standard fields remain supported; do not apply private closed-object decoding
to standard GitHub/SPDX/in-toto/SLSA envelopes.

This contract implements the upstream acquisition path in T2. A release-owned
build, repack or alternate publisher requires a separate accepted artifact/build
contract if selected. The earlier draft's unused local recipe/runner is not an
implementation dependency. No new signature, signing service, SLSA security
level or universal independently reproduced vendor build is imposed.

Produce binary → source/terms/standard evidence → metadata/SPDX/notices/index
→ final existing release manifest. The existing whole-release BuildRecord and
rehearsal records then bind the completed bundle; no final-manifest self-cycle.

## T6 — Actual producer and worker consumers

Extend the owning release-package assembler/generator after adoption. Copy only
explicit tool inputs and their complete metadata closure into an empty output;
validate before final manifest publication. The selected native-producer entry
point requires `ValidateBundle` plus the native profile reader's full T1–T5
validation, not merely file hashes/SPDX headers; it has no skip/default fallback.
The generic historical bundle format remains valid in its original scope.
An unsupported/missing group
fails before admitted work, without creating Docker jobs to inspect metadata.

OP2/SP5's selected ONE producer manifest binds the index, executable and closure.
Do not mix a host tool, unrelated executable, source/target helper or another bundle's
notices. Retained source/target bundles need this group only if selected as the
producer; no historical manifest rewrite or implicit same-version requirement.
Old helpers do not gain support because a new manifest lists extra files.

The real isolated preflight worker invokes the exact Compose file by absolute
path, following [C01](deployment-native-supervisor-adapters.md#c01--direct-configuration-worker).
It has an empty private Docker configuration, no plugin discovery, Engine channel
or host/system fallback. Prove actual execution/identity and missing/wrong-binary
refusal even when a host/PATH fallback is planted. Go remains the sole Engine
dispatcher; SP1–SP8 retain all worker/lifetime boundaries. Private compiler output
uses C01's unlogged original attach stream, never Docker logs as in the public
research fixture. All actual source/target/environment/path inputs must match
the trusted native compiler; no normalized short model can hide a mismatch.

All admitted configuration compiler inputs, success and failure semantics and
source/target compatibility still require the actual selected tools. A different
build is a different artifact, with full existing conformance, not a source/legal
shortcut. This metadata work cannot substitute for the real native SP8/R-006.

## Acceptance and adoption gates

1. Actual selected-tool producers, assembler, validator and native preflight
   consume one closure. Independently reconcile real discovered components and
   imported/copied/native/header scopes; omitted discovery cannot pass by omitting
   the same row from every output. Research examples are not shipping proof.
2. Intact package passes; missing/extra/wrong-role tool or metadata, duplicate
   keys/IDs and cross-tool/cross-producer paths refuse. Real ELF wrong-platform
   and executable swaps refuse before worker effects.
3. SPDX-header-only/empty-graph/wrong-root-checksum/orphan-component fixtures
   refuse. Scoped copyright/NOTICE removal and conflicting grants refuse EVEN
   AFTER regenerating ordinary bundle hashes. Actual notice bytes and original
   source origins re-render; wrong/missing origin associations refuse.
4. Every required source payload passes its exact version/hash/scope checks;
   removed/altered covered files or unexamined extra-source terms refuse.
   Insert an unlisted regular member into a real archive and rebuild its bundle
   hash: total member/catalogue/map equality must still reject it. Also reject
   a wrong binary/source-bound map and missing/foreign component associations.
   MPL failure cannot be laundered under a permissive root; unsupported native/
   OASIS/exception terms remain refusals. Policy adoption is recorded separately.
5. Actual upstream release/asset/source records bind the selected bytes. Wrong
   repository/tag/asset/platform/size/digest and foreign image subjects refuse;
   code-equal/whole-different artifacts cannot supply selected-binary provenance.
   Compatible standard extension fields still parse. No rebuild/alternate-
   publisher record can silently enter this upstream acquisition profile.
6. Actual isolated direct configuration workflows pass all declared inputs.
   Swap/remove the selected executable and seed host/PATH fallbacks: wrong bytes
   must never execute or yield valid compilation evidence. Prove C01's no-Engine/
   no-secret-descriptor grants and private unlogged output handling with actual
   rejection/discrimination; the public fixture is not that privacy proof.
   No dependency skip or reduced source/target/environment population.
7. Checkout-free/provider-off exact bundle operation preserves all parent
   acceptance populations, including native source/target/rollback compatibility,
   SP8 and final R-006. No archived/current source identity or threshold is waived.

Technical review covers this companion with SP1/SP5/OP2 and current assembler
ownership. Owner adoption must select the architecture/artifact floor and this
terms policy; actual pins/payloads and unresolved rights must be settled before
shipping eligibility. No research fixture, self-review, draft metadata, code-only
proof or local passing check supplies adoption, feature acceptance or release.
