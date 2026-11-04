# Native producer tools: provenance and rights checkpoint

2026-11-04. Scope: the existing **research-only** Linux/amd64 Docker CLI28.4.0
and Compose2.39.4 fixtures, not approved release inputs. Owning authority is
SP1's proposed tool-artifact floor. No distribution, tool-version recommendation,
new source dependency, build/runtime change or complete license clearance.

## Question and result

Can the cached tool hashes, upstream statements and the existing application
inventory satisfy the proposed shipping floor? **No.** These are distinct
evidence gaps, not a reason to omit either tool or mark a partial inventory full.

The [source-bound observation](native-tool-inventory-observation.v1.json) records
actual binary metadata and publisher-file consistency. The
[manual probe](probes/native-tool-inventory.py.txt) runs `go version -m` against
the actual ELF files, checks their bytes and the downloaded statement subjects,
and compares Compose's module rows to the upstream SBOM. It is a bounded audit,
not an artifact-authentication, signature, license or release-eligibility verifier.
Its completed inspection explicitly leaves distribution unapproved and native
dependency closure unknown. Invalid inputs/guards fail before creating output;
an existing observation path refuses without overwriting its bytes.

| Actual fixture | Verified local observation | Unresolved shipping evidence |
|---|---|---|
| Docker CLI28.4.0 |43,556,144bytes; SHA256`27f239f97492c434e091a70b41b8796c698f0aa2b6052c6f9c28da4ee7ae888b`; Go1.24.7, linux/amd64, no ELF interpreter; `CGO_ENABLED=1`, tags`osusergo,pkcs11`; ZERO embedded dependency-module rows; matching source query now maps623 packages/89 vendor modules | Zero rows do not mean zero dependencies. Published-binary/source correspondence and exact native linked-component/notice closure remain missing. Cached archive has a local hash but no independently checked publisher checksum. |
| Compose2.39.4 |76,361,386bytes; SHA256`7af95166a730b87e172d4fc9aefea8725d3c6c7327d59149267b452114ddb7d4`; Go1.24.7, linux/amd64, CGO0, upstream build tag`e2e`; all180 embedded dependency name/version rows occur in the183-package upstream SBOM |182 declared licenses are NOASSERTION; all183 concluded licenses are NOASSERTION. A module inventory supplies neither its license texts nor notices. Runtime/config compatibility of any selected shipping build remains mandatory. |

The publisher [Compose release](https://github.com/docker/compose/releases/tag/v2.39.4)
provides binary, SBOM and provenance asset hashes. Fresh downloaded files match
those SHA256/size fields and both statements name the **actual binary** digest:

- [SBOM statement](https://github.com/docker/compose/releases/download/v2.39.4/docker-compose-linux-x86_64.sbom.json):
  `a4ee58649188059813e6fdba38646139a12ee561379e8b01cf5bf0ea3bca7f1a`,380,834bytes.
- [Provenance statement](https://github.com/docker/compose/releases/download/v2.39.4/docker-compose-linux-x86_64.provenance.json):
  `d9e71aac728de6591d7c7e2a306d54bede1adb497adecf00ae27c884283f37ee`,50,136bytes.

These are publisher metadata obtained over HTTPS, not independently verified
signed attestations or a reproduced build. The API says the release is not
immutable. The provenance's builder is a claimed upstream Actions run; this
audit does not execute or authenticate that historical workflow.

Compose's embedded full VCS revision matches the observed
[tag reference](https://api.github.com/repos/docker/compose/git/ref/tags/v2.39.4),
`a32dc3da722e6b2e4fbfc72e40b1a697c4686460`. The
[Docker CLI annotated tag](https://api.github.com/repos/docker/cli/git/tags/313ac5662df10a1a4c247903dac7441d54b7ddf6)
names commit`d8eb465f86cfceeb57f8582e373d41a558d35503`, corroborating only the
binary's embedded short revision`d8eb465`; the API labels the tag unsigned.
Neither comparison proves the complete binary derives from that source.

Both projects publish top-level Apache-2.0 license text and a separate NOTICE:
[CLI LICENSE](https://github.com/docker/cli/blob/v28.4.0/LICENSE),
[CLI NOTICE](https://github.com/docker/cli/blob/v28.4.0/NOTICE),
[Compose LICENSE](https://github.com/docker/compose/blob/v2.39.4/LICENSE),
[Compose NOTICE](https://github.com/docker/compose/blob/v2.39.4/NOTICE).
Those files are downloaded/hashed inputs, **not** a conclusion about every
dependency's terms. The pinned CLI [Dockerfile](https://github.com/docker/cli/blob/v28.4.0/Dockerfile)
has an Alpine build path with musl development tools; that is source context,
not proof of the exact native link closure in this fixture.

## Consequence and next construction

### Exact Compose source and root-text follow-up

The complete [v2 source collection](compose-source-notices-observation.v2.json)
now covers all180 embedded dependency identities:59 archives from the existing
Go download cache and121 separately staged public-proxy downloads. Every full
archive matches its binary-reported Go `h1` identity. Total archive bytes are
106,533,767. The collector retains237 named license/NOTICE candidates containing
1,487,398bytes:200 root files and37 nested files. All180 modules have a named
root license candidate; filename presence is not a legal conclusion.

[Hash1](https://github.com/golang/mod/blob/v0.27.0/sumdb/dirhash/hash.go) hashes all
sorted archive names and contents; the collector does not authenticate a zip by
its transport or filename. Its500MiB compressed/expanded safety limits come from
[Go's module-zip contract](https://github.com/golang/mod/blob/v0.27.0/zip/zip.go).
Research collection guard600s is unchanged, with no exclusions or exhaustion.
The [failed v1](compose-source-notices-observation.v1.json) and its exact source
remain: Python DNS failed on121 downloads, so it completed only59 and exited1.
Completed direct curl transfers supplied those121 inputs before the offline v2
collector ran. Failed/partial inputs were not relabelled available or licensed.

The [manual actual-detector sweep](compose-license-classification-observation.v1.json)
binds that complete source observation and each private root-file population/
byte hash, then invokes the existing `licenseAt` and `DetectPermissiveLicense`
through a compiler overlay. It recognizes176 roots and refuses4; Go test exits1/
Make2. These4 exact pinned roots contain MPL-2.0 text:

- `github.com/hashicorp/errwrap@v1.1.0`
- `github.com/hashicorp/go-cleanhttp@v0.5.2`
- `github.com/hashicorp/go-multierror@v1.1.1`
- `github.com/hashicorp/go-version@v1.7.0`

This is the current detector's result, not a permission decision or a proposal
to label those files Apache/MIT. [MPL2.0 §3](https://www.mozilla.org/en-US/MPL/2.0/)
and [Mozilla's redistribution FAQ Q7–Q8](https://www.mozilla.org/en-US/MPL/2.0/FAQ/)
identify covered-source availability and recipient notices as part of executable
distribution. A tool-specific delivery proposal must account for the actual
covered source/notices and preserved recipient rights. The existing application
allowlist has not been widened. No shipping architecture/license policy adopted.

Root recognition does not resolve the37 nested scopes, main Compose source,
Go1.24.7 standard-library notices or the CLI's different Go/native build. Nor do
module checksums prove a vendor tree was unpatched or reproduce the binary.
Keep the source/build correspondence and complete final delivery gates separate.
The probe, v1/v2 sources and actual-refusal observation remain manual research
inputs; they are not normal CI tests or a package eligibility verifier.

### Nested-scope observation and application counterexample

The historical [v1 binary/source observation](compose-nested-scopes-observation.v1.json) parses
the actual pinned Compose ELF/Go metadata from one byte slice, checks all180
module identities and the complete source-inventory binding, then inspects all37
nested text candidates. All90,355 decoded functions have actual code-section
ranges. Twelve candidates have matching package/subtree function-name evidence;
25 do not. Every zero-match record says **unknown**, never absent or excluded.
Samples are limited to5 functions with explicit full counts/truncation flags.
This is positive binary evidence, not a complete compiler/package provenance map.

`github.com/docker/buildx@v0.28.0/bake/hclparser/LICENSE` contains MPL2.0 text,
SHA256`28ca54ff46a27183631aa262673d13e3c829d909059f5c861da835fe1783fbec`.
Its scope has zero function-name matches here; that does not establish whether
it contributes inlined code/data. The four module-root refusals therefore cannot
be presented as an exhaustive inventory of every potentially applicable MPL scope.
The37 filename candidates also include a workflow `license-check.yml` and
`license.go` source: a filename/detector label is not a separate legal grant.

A separate actual-application diagnostic found a DP1/AC8 omission at the
recorded pre-correction revision:
the six-command compiler import graph includes
`github.com/prometheus/client_golang/internal/github.com/golang/gddo/httputil`
at modulev1.23.2, but the current generator's delivered attribution omits the
copy's full nested BSD text and its2013 Go Authors copyright. This uses actual
source/import and actual metadata/formatter, not a package assumed present merely
because its module appears in an inventory. RP-761 owns the corrective route.
RP-761 is now independently corrected under957dfa6b..a8168b63; actual current
metadata retains that scoped BSD text. The root-NOTICE correction RP-760 remains
closed within its original scope. These application fixes do not cover tools.

### CLI source/vendor inventory

The [CLI source observation](cli-source-inventory-observation.v1.json) and
[manual probe](probes/cli-source-inventory.py.txt) query the actual pinned public
[v28.4.0 source](https://github.com/docker/cli/tree/d8eb465f86cfceeb57f8582e373d41a558d35503),
tree`c22925ab20127fcb06c19dff169e2547086d7c70`. Every inspected CLI/vendor file
matches a tracked blob in that tree. The checkout is clean before and after.
This source commit matches the annotated tag and the binary's short revision;
that correspondence still does not reproduce or authenticate the published build.

The host SDK is Go1.24.7 darwin/arm64; its archive is76,295,172bytes,
SHA256`d06bad763f8820d3e29ee11f2c0c71438903c007e772a159c5760a300298302e`,
matching the cached [Go publisher metadata](https://go.dev/dl/?mode=json&include=all).
The probe verifies the entire extracted SDK against that archive before invoking
its Go executable. No global installation or toolchain setting changes.
The upstream-required GOPATH source layout and `GO111MODULE=auto` preserve
vendoring. Cross-target `go list -deps -json -tags=osusergo,pkcs11 -buildmode=pie
./cmd/docker` selects Linux/amd64/v1, CGO1 and no extra Go experiments.

The complete query exits0 without diagnostics:623 packages, comprising83 CLI,
338 vendor and202 standard packages. All338 vendor packages map to89 exact
module/version owners through the tree's `vendor/modules.txt`; no unmapped
package is excluded. The observation retains every package identity, all116
named scoped license/NOTICE candidates (760,643bytes,11 nested), and a hash of
the3,203 selected source-file identities. Named candidates are not license
classification or exhaustive inline-term review. The first exploratory filename
rule omitted three `LICENSE-APACHE-2.0.txt` candidates; the final collector
includes dot, hyphen and underscore suffixes. That first count is not clearance.

Three packages declare CGO inputs: `runtime/cgo` requests`-lpthread`, `net`
requests`-lresolv`, and vendored `github.com/miekg/pkcs11@v1.1.1` requests`-ldl`.
These are source link requests, not an exact native-library version inventory.
The actual pinned Dockerfile offers Alpine/musl and Debian build variants; its
defaults cannot establish which native objects entered this published fixture.
The observation retains hashes of each package's declared native source/header
files. In particular, `pkcs11.h`, `pkcs11f.h` and `pkcs11t.h` contain a separate
OASIS Open2016 copyright/IPR-policy reference; the module-root BSD license alone
does not resolve that reference. No terms or redistribution rights are inferred.

Reproduce from repository root with the existing public cache inputs:
`python3 planning/deployment-foundation/probes/cli-source-inventory.py.txt
--output /tmp/cli-source-inventory.json`. The output is deterministic and refuses
an existing destination. The [private controls](probes/cli-source-inventory-controls.py.txt)
execute the real query before injecting an unowned vendor package, incomplete
graph, missing CLI command or removed dependency; all four refuse. The graph
must close over the CLI command's complete declared dependency set.
Replacement/duplicate vendor
metadata and overwriting an existing observation also refuse, preserving bytes.
No shared source/tool files are mutated. Complete seven-path66625d72..ac45914c
is independently APPROVED as bounded research/tool: fresh full observation is
byte-identical and all seven controls execute successfully. See the owning log.

This completes a source-derived package/owner and named-file inventory, not the
published binary's compiler/link map. Actual binary/source correspondence,
inline/header terms, exact native linked components, final scoped delivery and
tool adoption remain OPEN under RP-759. No CLI execution, native host, SP8,
AC8 or R-006 result follows from `go list`.

### Compose compiler-selected files and corrected binary matching

The [source-scope observation](compose-import-scopes-observation.v1.json) binds
the matching public main source at commit
`a32dc3da722e6b2e4fbfc72e40b1a697c4686460`, tree
`e702be38be3400928f7ad9e94470ad0cd6cff512`, and the same publisher-checked
Go1.24.7 SDK. It freshly runs the upstream Linux/amd64/CGO0/e2e source query
offline, with modules read-only and the normal Go module cache. All1,290 package
rows close over the command's complete dependency set:24 main,221 standard and
1,045 dependency packages. Their180 module/version/h1 identities exactly match
the hash-bound published Compose fixture. No source replacement is admitted.

The [manual collector](probes/compose-import-scopes.py.txt) checks6,962 inspected
source/text files. Main files match exact tracked Git blobs; standard files match
the publisher-checked SDK archive; dependency files match the complete h1-verified
module archives. Archive hashing and reads share the same immutable byte snapshot.
Every prior237 module candidate, including scopes with no selected package,
also matches its physical module-cache file. Main LICENSE/NOTICE plus six selected
standard-library scoped files bring the named-file population to245. Filename
presence and source selection still do not decide legal applicability or rights.
Final inspected-file hashes are rechecked before output; errors/600s guard
exhaustion refuse, with no successful partial report.

Of37 nested module candidates,17 have a selected package in their source
directory/subtree and20 do not. All remain in the report; five-package samples
carry full counts and explicit truncation. The17 include copied GDDO BSD,
libyaml and other distinct scoped texts. Buildx's MPL-marked hclparser scope has
no selected package here. This is a matching-source result, not an absence proof
about a published binary that has not been reproduced.

The comparison exposed RP-762: the old function-name matcher misses compiler
escaping. Go1.24.7's
[objabi.PathToPrefix](https://github.com/golang/go/blob/go1.24.7/src/cmd/internal/objabi/path.go)
escapes periods in the final package-path component. The actual Compose binary
contains `sigs.k8s.io/yaml/goyaml%2ev2` functions; raw `goyaml.v2` did not match.
V1's zero label remains UNKNOWN and its positive evidence is retained; no
previous omission was permission to discard those terms.

The [v2 manual binary probe](probes/compose-nested-scopes.v2_test.go.txt) applies
that encoding separately to exact packages and descendant prefixes. A regression
fails on the old matcher for the actual escaped YAML name, percent and UTF-8;
the correction passes those plus subtree, sibling-refusal and ordinary-method
controls. Full overlaid metadata tests and vet pass. The
[fresh v2 ELF observation](compose-nested-scopes-observation.v2.json) still checks
all90,355 functions/all37 candidates:15 positive scopes,22 UNKNOWN. Each of the
three goyaml.v2 license/NOTICE scopes now matches the same184 actual functions;
these are three text scopes, not three separate function populations. The two
other newly source-selected candidates share K8s's spec package and still have
zero binary-function matches. No absence or inline/data contribution is inferred.

Reproduce the offline source observation with
`python3 planning/deployment-foundation/probes/compose-import-scopes.py.txt
--output /tmp/compose-import-scopes.json`. The
[private controls](probes/compose-import-scopes-controls.py.txt) run the actual
query before removing a dependency, forging a module h1, changing selected-source
read bytes or changing an unselected scope's text read bytes. Each refuses;
existing output also refuses unchanged. These are private returned-byte faults,
not mutations to shared SDK/checkout/module/archive files. The owning log gives
the Go test overlay and exact old-failure/new-pass runs. Complete nine-path
a6592af3..571b0d8c is independently APPROVED as bounded research/tool. Fresh
source and binary reports are byte-identical, all five source controls pass,
and old regression/new full overlay/vet executions agree. RP-762 closes only
the escaped-name instrument defect. The owning log retains the exact verdict.

This advances complete source/text evidence, not artifact authentication,
published-build/native correspondence, inline-term clearance, delivery-policy
adoption or tool redistribution. The four actual root MPL refusals, CLI's native
and OASIS issues, final notices/covered-source delivery and SP1/native/R-006 gates
remain OPEN. No application or normal CI behavior changes.

The existing `gen-release-metadata` discovers dependencies of six Cloud Clicker
Go commands and the built browser client. It does not inspect external CLI/
Compose bytes. Adding these binaries to an artifact list cannot silently make
that application's SBOM cover their dependencies. The Docker archive also
contains daemon/container-runtime executables; this proposal needs ONLY its CLI,
not accidental redistribution of the entire archive.

Before SP1 adoption/packaging readiness: establish selected tool source/build
provenance and actual linked Go/native closure; collect the corresponding license
and NOTICE bytes; specify release-owned paths, manifest hashes and separate
tool SBOM/notice inputs; extend the assembler/validator only after the applicable
adoption. Preserve the existing app/image inventories. No unknown license may be
guessed from a module name or replaced by the top-level Apache label. A different
build is a different artifact requiring its own compatibility/native proof.

DP7's projector/cache adoption prerequisites and native-host/R-006 remain
separate. This audit advances SP1's evidence; it does not complete either gate.
