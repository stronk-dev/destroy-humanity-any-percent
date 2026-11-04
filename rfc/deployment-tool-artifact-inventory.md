# RFC: Deployment Tool Artifact Inventory and Delivery

- **Status:** draft — technical review and owner adoption pending; no implementation or shipping authority
- **Author:** Codex
- **Created:** 2026-11-04
- **Parent / amends:** [Deployment Foundation](deployment-foundation.md), DP1/AC8; [Native Supervisor Profile](deployment-native-supervisor-profile.md), SP1/SP5/SP8
- **Design refs:** `design/06-tech.md` §stack; `design/00-vision.md` pillar6
- **Planning:** existing `planning/deployment-foundation/plan.md` and append-only log

## Outcome and authority

An operator receives the exact CLI/configuration compiler used by the selected
native producer, with a complete, separate inventory of its components, terms,
notices and required source. Its identity and delivery validate without a checkout
or a runtime license lookup. Application and image inventories retain their own
scopes. No version, build variant or research fixture is selected for shipping.

This supplies the missing SP1 packaging contract. It does not adopt the native
architecture/host floor, decide unresolved third-party rights, change gameplay,
install tools on the operator's host, introduce an image/service/signing key, or
authorize implementation before acceptance. Technical approval is not adoption
or a legal conclusion. Unknown terms remain blockers, not configurable waivers.

The [tool audit](../planning/deployment-foundation/native-tool-provenance-audit.md)
owns research evidence. In particular: the CLI's empty module list is not zero
dependencies; Compose's four root MPL refusals cannot be labelled permissive;
headers and copied scopes matter; the publisher OCI counterpart differs from
the original binary despite matching code. This contract does not require
universal upstream signatures or independently reproduced vendor builds.

## T1 — One bundle authority and fixed tool roles

Retain release-manifest v1's sorted `artifacts` path/SHA256 list as the authority
for all delivered bytes. Add files, not a parallel manifest/ledger or a new
release identity. Existing non-native bundles are not restamped or rewritten.
The adopted native producer requires the complete tool group below; an older
validator's acceptance of extra files does not qualify as native validation.

`tools/inventory.json` is a closed JSON object with required integer
`schema_version:1` and `tools`, exactly two entries sorted by `id`. Each entry
has exactly `id`, `version`, `platform`, `binary_path`, `metadata_path`,
`sbom_path`, `notices_path`. `platform` is `linux/amd64`; `version` is the exact
nonempty ASCII release version, not `latest`, an index/tag alias or a range.
All paths equal the following table, not arbitrary caller-selected locations:

| ID | Binary | Metadata | SPDX | Readable notices |
|---|---|---|---|---|
| `docker-cli` | `tools/docker` | `tools/metadata/docker-cli.json` | `sbom/tools/docker-cli.spdx.json` | `licenses/tools/docker-cli.txt` |
| `docker-compose` | `tools/cli-plugins/docker-compose` | `tools/metadata/docker-compose.json` | `sbom/tools/docker-compose.spdx.json` | `licenses/tools/docker-compose.txt` |

Ship `tools/inventory.schema.json` for these private records. Deployment owns
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
describes the host daemon; Docker CLI has its own independent version.
Source/target manifests keep their own identities; do not demand every retained
bundle have the selected producer's tool version. No historical field changes.

## T2 — Metadata and delivery catalogue

Each metadata file is a closed object with required `schema_version:1`,
`tool_id`, `version`, `platform`, `acquisition`, `source_record`, `provenance`
and `components`. The first three identity values match the index.

`acquisition` has exactly `kind`, `uri`, `container_sha256`, `member`,
`build_record_path`; nullable values are explicit:

| Kind | Required values | Null values |
|---|---|---|
| `published_file` | HTTPS public URI; container SHA256 equals delivered binary hash | member, build record |
| `published_archive` | HTTPS URI, exact parent archive SHA256, exact regular member path | build record |
| `local_build` | manifest-relative build-record path | URI, container hash, member |

The generator verifies the real download/archive/member or build output before
emitting a record. A parent download is a build input, not automatically a bundle
payload: do not ship the Docker daemon/container-runtime executables merely
because the CLI arrived in their archive. Runtime validation never downloads it.

`source_record` has exactly `uri` (public HTTPS repository URI), `commit`
(full 40-character lowercase hex), `basis` (`binary_vcs`, `publisher_release`, `local_build`)
and `evidence_paths` (nonempty, sorted unique manifest-relative paths). The basis
states what was actually established; a tag/API claim or short revision is not
silently upgraded to compiled-source authentication. Full source inputs and
native/header evidence belong to the producer's reconciled component inventory.

`provenance` is an array of objects with exactly `path`, `kind`, `scope`, sorted
by unique `path`. `kind` is `in_toto`, `oci_chain` or `source_evidence`; `scope` is
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
an archive. Every source payload is used by at least one entry. Components with
no source-delivery obligation have empty source arrays; a reviewed obligation
cannot be removed merely by emitting empty arrays. The per-version reviewed
scope map and independent producer check establish required coverage; runtime
validation checks that declared coverage against the delivered files. Neither
protects against an attacker replacing the entire trusted release and its policy.

Grant/NOTICE/copyright bytes retain their complete original content, including
relative source scope and any multiple grants. `licenses/tools/<id>.txt` renders
components in SPDX-ID order, then sorted grant paths and notice paths. Each
section identifies component/version/path and includes the referenced raw bytes,
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

## T5 — Truthful acquisition/build provenance

Keep vendor in-toto/SLSA records verbatim. Validate supported envelope/predicate
and subject semantics using their standard parsing rules: private closed-object
rules do not reject compatible extension fields in standard SLSA documents.
Recorded signature verification or reproduction claims require actual evidence;
neither is universally required by this RFC. Do not infer a SLSA security level.

A binary subject matches the actual executable digest. An OCI subject can bind
that executable only through a verified selected-platform manifest/config/layer/
diff-id/member chain with exact member bytes equal to the selected tool. Matching
code/ranges, a revision string or another build's data is insufficient. Source-
only/related documents retain those scopes. Never rewrite their subjects to fit.
Published acquisition/source records remain honest even when no vendor build
attestation exists; they cannot claim more than the download and source evidence
establish. The accepted release review still checks those evidence limits.

For a release-owned rebuild, generate a standard in-toto Statement/v1 with SLSA
`https://slsa.dev/provenance/v1` build provenance, subject SHA256 equal to the tool.
Use the [SLSA build-provenance format](https://slsa.dev/spec/v1.2/build-provenance).
This section defines the release-owned `tool-build-v1` type. Its `buildType` URI
is the immutable GitHub blob URL for this RFC at the producer release's full
`source_commit`, with fragment `#t5--truthful-acquisitionbuild-provenance`.
Its closed external parameters are `tool_id`, `version`, `platform`, `source_uri`,
`recipe`: the last is a manifest-relative public recipe under the tool's
provenance prefix. It identifies the exact commands, flags, environment and
input preparation, without secrets. The runner verifies resolved source, recipe,
toolchain, module, native/header, patch and generated inputs, invokes that recipe
and hashes its actual output. List their actual immutable URIs and digests in
`resolvedDependencies`; bind the source commit to `source_record` and the recipe
digest to the manifest. Missing/undeclared inputs fail the producer, rather than
emitting a guessed complete record. Record actual invocation/timestamps.

The builder ID is the immutable blob URL of the release-owned runner at that
same commit. The validator recognizes this build type and runner relationship;
an arbitrary builder string does not qualify. These are trusted local producer
claims, not a signature or proof of an independently secured builder. No operator
secrets, secret hashes, private configuration or invented successful steps.
Acceptance must execute the real recipe and runner; a hand-written statement
cannot supply that evidence. The implementation supplies those executable
producers under the adopted contract, not another future unspecified build type.

Do not embed the existing whole-release `BuildRecord` here: it hashes the final
release manifest and would create a cycle. Order outputs as binary → tool
inputs/source/terms/provenance → tool metadata/SPDX/notices/index → final existing
release manifest. Existing retained build/rehearsal records then bind that bundle.

## T6 — Actual producer and worker consumers

Extend the owning release-package assembler/generator after adoption. Copy only
explicit tool inputs and their complete metadata closure into an empty output;
validate before final manifest publication. The selected native-producer entry
point requires `ValidateBundle` plus the native profile reader's full T1–T5
validation, not merely file hashes/SPDX headers; it has no skip/default fallback.
The generic historical bundle format remains valid in its original scope.
An unsupported/missing group
fails before admitted work, without creating Docker jobs to inspect metadata.

OP2/SP5's selected ONE producer manifest binds the index, binaries and closure.
Do not mix a host CLI, unrelated plugin, source/target helper or another bundle's
notices. Retained source/target bundles need this group only if selected as the
producer; no historical manifest rewrite or implicit same-version requirement.
Old helpers do not gain support because a new manifest lists extra files.

The real isolated preflight worker uses those exact CLI/plugin bytes. Its private
Docker config points discovery at the selected read-only plugin directory;
it contains no host credentials/context. Do not assume a PATH or configured
directory proves execution: CLI discovery also has fallback locations. Prove
actual execution/identity and refusal of missing/wrong/shadowed plugins without
falling back to host/system tools. This adds no admitted CLI mutation; Go remains
the sole Engine dispatcher and SP1–SP8 retain all worker/lifetime boundaries.

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
   and executable/plugin swaps refuse before worker effects.
3. SPDX-header-only/empty-graph/wrong-root-checksum/orphan-component fixtures
   refuse. Scoped copyright/NOTICE removal and conflicting grants refuse EVEN
   AFTER regenerating ordinary bundle hashes. Actual notice bytes re-render.
4. Every required source payload passes its exact version/hash/scope checks;
   removed/altered covered files or unexamined extra-source terms refuse.
   MPL failure cannot be laundered under a permissive root; unsupported native/
   OASIS/exception terms remain refusals. Policy adoption is recorded separately.
5. Real acquisition/archive/member and local-build records bind the selected
   bytes. Foreign image subjects and code-equal/whole-different artifacts refuse
   selected-binary provenance; compatible standard extension fields still parse.
6. Actual isolated CLI/plugin configuration workflows pass all declared inputs;
   sever discovery, swap/remove a plugin and plant a fallback. The wrong bytes
   must never execute or yield valid compilation evidence. No dependency skip.
7. Checkout-free/provider-off exact bundle operation preserves all parent
   acceptance populations, including native source/target/rollback compatibility,
   SP8 and final R-006. No archived/current source identity or threshold is waived.

Technical review covers this companion with SP1/SP5/OP2 and current assembler
ownership. Owner adoption must select the architecture/artifact floor and this
terms policy; actual pins/builds and unresolved rights must be settled before
shipping eligibility. No research fixture, self-review, draft metadata, code-only
proof or local passing check supplies adoption, feature acceptance or release.
