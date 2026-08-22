# Deployment

Deployment Foundation is implementing. The repository does **not** yet claim a supported self-host
bundle or release-ready deployment; backup, rollback, operations and the exact-manifest clean-host
rehearsal remain unfinished.

## Runtime content closure

`server/releasepackage` derives the gameserver's repository-independent content from the live epoch
authority rather than maintaining another catalog list. The closure contains:

- `balance/epochs/phase0.json`, every artifact and epoch changelog it declares;
- `balance/transport/phase0.json` and `moderation/guild-names.txt`, the two additional files opened
  by gameserver composition; and
- `deployment/content-manifest.v1.json`, whose constants hash must equal the derived current bundle
  and whose copy hash must be canonical.

Every path is local, unique, nonempty and SHA-256 bound. Staging accepts only an empty destination,
copies exactly that sorted set with no test/source/planning trees, and then re-walks the output.
Missing, altered, extra or symlinked files fail validation. This means a later container can copy
the staged directory to `/opt/cloud-clicker/content` without depending on a repository checkout.

Run the staging boundary with:

```sh
make stage-release-content RELEASE_CONTENT_OUTPUT=/absolute/empty/directory
```

The command prints the epoch, constants/copy identities and every staged file hash as canonical
JSON. This is an input to the release manifest still being implemented; it is not itself release
evidence.

`deployment/Dockerfile.gameserver` is a `scratch`-based, numeric-nonroot image boundary that copies
only the statically linked binary and that staged content. `make build-gameserver-linux-amd64`
builds the reproducible, trimpath, VCS-metadata-free binary at the explicitly supplied
`RELEASE_SERVER_OUTPUT`. The image still receives release version/source commit as OCI labels, so
provenance lives in the release boundary rather than a host-dependent Go build record.
`make build-gameserver-image` requires `--platform linux/amd64` through its fixed recipe and a
caller-supplied `SOURCE_DATE_EPOCH` (normally the source commit time), then writes a `docker save`
archive. The assembler verifies the saved config's Linux/amd64 identity, nonroot entry point and
OCI version/revision/source/license labels against the release manifest; a digest from another
commit is rejected.

`deployment/compose.template.yml` defines the current core topology: only Caddy publishes two host
ports; Caddy and gameserver share an internal application network; gameserver and Postgres share a
separate internal database network. The gameserver is read-only, drops all capabilities and mounts
only current file-backed secrets. `compose.rotation.template.yml` adds previous JWT/bootstrap pairs
only during an actual overlap—ordinary installations do not manufacture placeholder previous
secrets. Every rendered Caddy/gameserver/Postgres image reference must include an immutable
`@sha256:` digest. The Caddy route list includes only the SPA, API, WebSocket, health and readiness;
metrics are deliberately absent from the public proxy.

`make render-release-compose` replaces the three image tokens only when each supplied reference is
digest-pinned, validates the private topology, and refuses to overwrite an existing output. The
checked-in `.env.example` contains only operator configuration and host paths to secret files;
`deployment/secrets/` is ignored so the documented layout cannot be committed accidentally. Docker
Compose's own `config` command has parsed the rendered boundary successfully. The release assembler
also runs the repository validator against the exact rendered bytes; the later clean-host rehearsal
runs Compose's parser and the stack itself.

`deployment/config.schema.json` describes the non-secret operator `.env` inputs and rejects unknown
members. `deployment/release-manifest.schema.json` describes the byte-binding release record. The
runtime `validate-config` command remains the authoritative validator after Compose has mapped
operator paths to `/run/secrets`; JSON Schema is not substituted for opening and validating the
actual secret files.

## Application licenses and SBOM

`make generate-release-metadata` inventories the module graph actually linked into
`cmd/gameserver` (not the much larger `go.sum` graph), adds the Go standard library, and reads the
three exact browser runtime dependencies from `client/package.json` plus their installed package
manifests. It reads shipped LICENSE/COPYING bytes directly, recognizes only the audited MIT,
Apache-2.0, BSD-2-Clause and BSD-3-Clause family, preserves multi-license modules as SPDX `AND`
expressions, and fails on missing, ambiguous, unknown or metadata-mismatched licenses.

The outputs are `third-party-licenses.txt` and an SPDX-2.3 JSON document with package-manager purls,
download locations and root `DEPENDS_ON` relationships. On the current graph the generator finds 37
linked Go modules, the Go standard library and three browser dependencies (41 dependencies total),
matching the prior license audit while retaining the previously hidden dual Apache-2.0/MIT libyaml
notice. Version, full commit and RFC3339 creation time are explicit inputs; an existing output
directory is never silently overlaid.

This is the application SBOM only. The assembler requires separate SPDX inputs for Caddy, the
gameserver image and Postgres and binds each SBOM hash beside that image's immutable digest.

## Release bundle assembly

`make assemble-release-bundle` accepts only an empty output directory and requires all of the
following explicit inputs: the Linux/amd64 gameserver binary and its `docker save` archive, built
client, generated application metadata, release version/full source commit, tested Docker
Engine/Compose versions, three digest-pinned image references and their three image SBOMs. It
rejects a non-ELF or non-amd64
binary, client symlinks, an absent SPA entry point, empty/missing inputs, mutable image references
and a pre-existing output tree.

The resulting directory contains the runtime content closure, site, binary, offline gameserver
image archive, Docker/Caddy/Compose inputs, schemas, root and third-party licenses, four SBOM documents and
`release-manifest.json`. The manifest records the current migration, both save-schema versions,
epoch/copy/constants identities and the SHA-256 of every other bundle file. Validation re-walks the
directory and rejects any missing, extra or changed byte, including attribution or an image SBOM.
It intentionally describes a release *candidate*: backup, rollback, operations and the exact
clean-host R-006 rehearsal remain required before the project can claim supported self-hosting.
