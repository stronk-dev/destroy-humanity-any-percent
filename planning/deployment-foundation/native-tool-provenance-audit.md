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
| Docker CLI28.4.0 |43,556,144bytes; SHA256`27f239f97492c434e091a70b41b8796c698f0aa2b6052c6f9c28da4ee7ae888b`; Go1.24.7, linux/amd64, no ELF interpreter; `CGO_ENABLED=1`, tags`osusergo,pkcs11`; ZERO embedded dependency-module rows | Zero rows do not mean zero dependencies. Actual Go/vendor and native linked-component/notice closure are missing. Cached archive has a local hash but no independently checked publisher checksum. |
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
