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

Each tool also has a fixed scope map at `tools/metadata/<id>.scopes.json`.
Ship `tools/inventory.schema.json` for these private records, including scope maps
and the T5 recipe. Deployment owns
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
`tool_id`, `version`, `platform`, `acquisition`, `source_record`, `provenance`,
`scope_map_path` and `components`. The identity values match the index;
`scope_map_path` equals the fixed T1 path for that tool.

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
input/result and preparation. A local build must stay within T5's explicit mode;
an opaque derivation does not extend the runner's permitted steps.

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
is the immutable GitHub blob URL for this RFC at the build producer's full
revision, with fragment `#t5--truthful-acquisitionbuild-provenance`.
Its closed external parameters are `tool_id`, `version`, `platform`,
`source_repository`, `source_commit`, `recipe`: the last is exactly
`provenance/tools/<id>/recipe.json`. Identity and declared repository/commit
expectations derive from the recipe and match `source_record.uri`/`commit`.
They are distinct from the actually acquired source archive URI/hash. The fixed runner is the
`build-tool` mode of existing `server/cmd/assemble-release-bundle/main.go`;
`builder.id` is its immutable GitHub blob URL at that same producer commit.
The validator compares that exact path and commit, not any same-commit blob.
No additional operator-installed or distributed helper binary is introduced.

### Tool-build-v1 recipe and initiation

This mode builds **pure-Go, CGO-disabled Linux/amd64 tools from a vendor-ready
source tree** on Linux/amd64. It does not reproduce the current CGO-enabled CLI
fixture or support arbitrary preparation, package managers, patches or native
compiler/linker commands. Published artifacts still have T2's acquisition route.
A different build mode needs an explicit contract amendment, not guessed steps.

The closed recipe keys are `schema_version:1`, `tool_id`, `version`, `platform`,
`source_repository`, `source_commit`, `inputs`, `source_input_id`,
`go_toolchain_input_id`, `source_subdirectory`,
`import_path`, `module_mode`, `main_package`, `build_mode`, `build_tags`,
`strip_symbols`, `strip_dwarf`, `string_overrides`. Identity matches T1.
`inputs` has exactly one source and one Go SDK, sorted by unique `id`, with
closed rows `id`, `kind` (`source` or `go_toolchain`), `uri` (public HTTPS),
`sha256`, `strip_prefix` (exact archive directory prefix). Both are tar.gz
inputs; IDs identify files `<id>.tar.gz` in the declared input directory. Hashes
and safe members verify before extraction, with no links, duplicates or special
files. `source_input_id`/`go_toolchain_input_id` select the corresponding rows.
`source_repository` is the public HTTPS repository URI; `source_commit` is its
full 40-character lowercase hex revision, both declared expectations rather
than inferred tarball metadata. The runner records them as external parameters
and separately records actual acquired archive URI/hash. It does not claim Git
authentication merely from compiling a tarball. After the build, the T3 producer
verifies actual Git/tree/import/terms correspondence against those expectations
and archive bytes before emitting/accepting the final binary-bound scope map.
That map is not a cyclic pre-build input. SDK platform/full extracted contents
match the selected official archive. Missing source evidence blocks metadata
eligibility, even if compilation succeeded; a version alone proves no source tree.

`source_subdirectory` is an existing safe relative directory, empty for root.
`module_mode` is `module_vendor` or `gopath_vendor`: respectively require the
source's complete `vendor/modules.txt` under Go's vendor consistency checks, or
the source's own vendored GOPATH dependencies. No dependency download, generated
vendor tree, code-generation hook or source patch runs in this build type.
`import_path` is null for modules, otherwise the source's exact Go import path;
stage it at that path inside the private GOPATH workspace. `main_package` is one
existing relative main package (`.` or `./path`), never `...` or an absolute path.
T3 independently reconciles compiled source/dependency identities, including
vendor bytes against original module/Git evidence; a self-consistent modules.txt
alone is not source authentication. Unsupported/incomplete source trees fail.

`build_mode` is `default` or `pie`. `build_tags` is sorted unique ASCII identifiers
using letters/digits/underscore. Strip fields are boolean. `string_overrides` is
sorted by unique `symbol`, with closed `{symbol,value}` rows: a qualified Go
string variable and its nonempty ASCII value (letters/digits and `._+:/@=-`, no
whitespace/control characters). These become `-X symbol=value`; booleans add
`-s`/`-w`. No arbitrary linker flags, shell command, `-toolexec`, external linker
or caller environment override is accepted.
The reviewed recipe stamps the tool's real version variable to its declared
version, not a label detached from the executable. The actual isolated tool
version output must agree during conformance before shipping eligibility.

After adoption, set the existing `RELEASE_COMMIT` to the full producer revision
and initiate through a root Make target:

```sh
make build-deployment-tool RELEASE_COMMIT="$RELEASE_COMMIT" TOOL_RECIPE=/public/recipe.json TOOL_INPUT_DIRECTORY=/public/inputs TOOL_BUILD_OUTPUT=/fresh/output
```

The target invokes the fixed runner's `build-tool` mode with `--root`,
`--producer-commit`, `--recipe`, `--input-directory`, `--output`; output must not
exist or must be empty. `RELEASE_COMMIT` uses the existing build source identity.
The runner checks full Git HEAD and clean runner/build-contract inputs
against that commit, records it as `internalParameters.producerSourceCommit`,
and the final validator requires the build-type URL, fixed runner URL and that
internal revision to agree. This is the reviewed tool-build producer revision,
which may precede a later application's manifest `source_commit`: unchanged
tools need not be rebuilt for an unrelated application edit. The final manifest
still binds the exact binary, recipe and receipt bytes; no uncited runner or
unsupported build type is admitted. These
paths are staging locations, not additional behavioral parameters or secrets.
The runner uses a private extracted workspace and the declared SDK's absolute
`bin/go`, with argv for `go build -a -trimpath -buildvcs=false`, the declared
build mode/tags/linker settings, a fixed private output and the exact package.
For modules add `-mod=vendor`. Set `GO111MODULE=on` or `off` for the selected mode;
set `GOOS=linux`, `GOARCH=amd64`, `GOAMD64=v1`, `CGO_ENABLED=0`, `GOENV=off`,
`GOWORK=off`, `GOTOOLCHAIN=local`, `GOPROXY=off`, `GOSUMDB=off`. Only the declared
SDK is on PATH. Use the root Makefile's existing GOCACHE; no task-specific
compiler cache or downloaded-module cache supplies source. Inherited Go flags,
toolchain switches, credentials, contexts and build hooks cannot alter this argv
or environment. Never execute the resulting tool as a build hook.

Successful output is exactly `tool` plus `build.intoto.json`; failed builds leave
no successful record. Verify actual ELF/platform and Go build settings, hash the
actual executable and emit the standard statement with actual invocation times.
The assembler copies these to T1's binary path and
`provenance/tools/<id>/build.intoto.json`; the latter is `build_record_path` and
an `in_toto`/`selected_binary` provenance row. `resolvedDependencies` includes the
exact source archive, SDK archive, recipe and discovered vendored/component
inputs with their actual URIs/digests. A declared Git expectation is not listed
as an authenticated resolved Git input without actual verification. The declared
source tuple matches `source_record` and T3's later verified scope map;
recipe bytes match the final manifest. No guessed successful steps or omitted
inputs. These are trusted local producer claims, not signatures, builder
security levels or independent reproduction claims. Acceptance executes the
actual runner/recipe and rejects foreign runner/recipe/input/settings/output.

Illustrative recipe (zero digests and example URI are placeholders, NOT a
selected artifact; missing/mismatched real inputs must refuse):

```json
{
  "schema_version": 1,
  "tool_id": "docker-cli",
  "version": "28.4.0",
  "platform": "linux/amd64",
  "source_repository": "https://github.com/docker/cli",
  "source_commit": "0000000000000000000000000000000000000000",
  "inputs": [
    {"id":"sdk","kind":"go_toolchain","uri":"https://go.dev/dl/go1.24.7.linux-amd64.tar.gz","sha256":"sha256:0000000000000000000000000000000000000000000000000000000000000000","strip_prefix":"go/"},
    {"id":"source","kind":"source","uri":"https://example.invalid/cli-source.tar.gz","sha256":"sha256:0000000000000000000000000000000000000000000000000000000000000000","strip_prefix":"cli/"}
  ],
  "source_input_id": "source",
  "go_toolchain_input_id": "sdk",
  "source_subdirectory": "",
  "import_path": "github.com/docker/cli",
  "module_mode": "gopath_vendor",
  "main_package": "./cmd/docker",
  "build_mode": "pie",
  "build_tags": ["osusergo"],
  "strip_symbols": false,
  "strip_dwarf": false,
  "string_overrides": [{"symbol":"github.com/docker/cli/cli/version.Version","value":"28.4.0"}]
}
```

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
   AFTER regenerating ordinary bundle hashes. Actual notice bytes and original
   source origins re-render; wrong/missing origin associations refuse.
4. Every required source payload passes its exact version/hash/scope checks;
   removed/altered covered files or unexamined extra-source terms refuse.
   Insert an unlisted regular member into a real archive and rebuild its bundle
   hash: total member/catalogue/map equality must still reject it. Also reject
   a wrong binary/source-bound map and missing/foreign component associations.
   MPL failure cannot be laundered under a permissive root; unsupported native/
   OASIS/exception terms remain refusals. Policy adoption is recorded separately.
5. Real acquisition/archive/member and local-build records bind the selected
   bytes. Foreign image subjects and code-equal/whole-different artifacts refuse
   selected-binary provenance; compatible standard extension fields still parse.
   Execute T5's actual recipe/runner. Foreign same-commit runner blob, wrong
   recipe/input/producer commit, altered settings/output, incomplete vendoring
   and inherited environment overrides must not produce a valid build record.
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
