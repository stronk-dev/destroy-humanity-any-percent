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
