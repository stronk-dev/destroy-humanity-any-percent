# Deployment

Deployment Foundation is implementing. The repository does **not** yet claim a supported self-host
bundle or release-ready deployment; rollback, operations and the exact-manifest clean-host
rehearsal remain unfinished, and the backup implementation remains subject to its required review.

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
`make build-gameserver-image` uses the named `cloud-clicker-release-v1` Buildx builder, whose
BuildKit image is pinned for the release rehearsal. Its recipe fixes `linux/amd64`, disables
provenance attachment for the offline Docker archive, rewrites layer timestamps and requires a
caller-supplied `SOURCE_DATE_EPOCH` (normally the source commit time). The assembler verifies the
saved config's Linux/amd64 identity, nonroot entry point and
OCI version/revision/source/license labels against the release manifest; a digest from another
commit is rejected.

`deployment/compose.template.yml` defines the current core topology: only Caddy publishes two host
ports; Caddy and gameserver share an internal application network; gameserver and Postgres share a
separate internal database network; the backup worker shares only that private database network.
The gameserver and backup worker are read-only, nonroot and drop all capabilities. The gameserver
mounts only current file-backed secrets. `compose.rotation.template.yml` adds previous JWT/bootstrap pairs
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
For upstream multi-platform references it also records the selected linux/amd64 OCI config digest;
the SPDX document name must identify that exact runtime config, preventing a native-host SBOM from
being attached to the supported amd64 release.

## Release bundle assembly

`make assemble-release-bundle` accepts only an empty output directory and requires all of the
following explicit inputs: the Linux/amd64 gameserver and deployment-backup binaries, the
gameserver's `docker save` archive, built
client, generated application metadata, release version/full source commit, tested Docker
Engine/Compose versions, three digest-pinned image references, their linux/amd64 config digests and
their three image SBOMs. It
rejects a non-ELF or non-amd64
binary, client symlinks, an absent SPA entry point, empty/missing inputs, mutable image references
and a pre-existing output tree.

The resulting directory contains the runtime content closure, site, gameserver and backup-helper
binaries, offline gameserver
image archive, Docker/Caddy/Compose inputs, schemas, root and third-party licenses, four SBOM documents and
`release-manifest.json`. The manifest records the current migration, both save-schema versions,
epoch/copy/constants identities and the SHA-256 of every other bundle file. Validation re-walks the
directory and rejects any missing, extra or changed byte, including attribution or an image SBOM.
It intentionally describes a release *candidate*: designated approval of the backup batch,
rollback, operations and the exact clean-host R-006 rehearsal remain required before the project
can claim supported self-hosting.

## Encrypted Postgres backup and restore

`deployment-backup` is a statically linked Linux/amd64 operator helper included in, hashed by and
validated with the release bundle. The base Compose stack runs it in the pinned Postgres 16 Alpine
image so its `pg_dump` and `pg_restore` major is exact. The worker is nonroot, read-only, drops all
capabilities, has no published port and can reach only the private database network. Its `/tmp` is
an isolated tmpfs; the final backup target is a separate operator mount and is never the Postgres
data volume.

The scheduled command creates a Postgres custom-format dump immediately and every six hours. It
reads the database URL from a read-only Compose secret and passes credentials to libpq through
temporary mode-0600 service/pass files—never command arguments or logs. Outside `/run/secrets`,
the helper accepts only owner-readable secret files. A backup is committed only
after all of the following succeed:

- the live Goose migration equals the exact release manifest;
- `pg_dump` 16 completes a custom-format archive;
- the release manifest and its exact epoch declaration validate and agree;
- the dump, manifest, epoch and authenticated metadata are encrypted to the operator's age X25519
  recipient; and
- the encrypted checksum is verified, synced and atomically renamed to `<backup-id>.ccbackup`.

No unencrypted dump is written to the off-host target. Interrupted reads and writes remove their
temporary output; a crash-surviving temporary or otherwise invalid file makes the next retention
pass fail rather than disappear quietly.

The retention policy keeps every completed six-hour backup for seven days, then one completed
backup per UTC day through day 30. It always protects the newest valid backup and every unresolved
pre-upgrade backup. A population containing an invalid file blocks deletion. A missing population
or newest completion older than six hours is reported as a failing state.

Restore requires the age identity explicitly; that identity is not mounted in the continuously
running stack. The helper authenticates the outer envelope, encrypted metadata, release-manifest
digest, epoch bytes, artifact inventory and dump checksum before invoking `pg_restore`. It accepts
only an empty target database, checks the custom archive before mutation, uses `--exit-on-error`
and requires the restored Goose migration to equal the release manifest. It never cleans or
overwrites a live database.

Build the helper with:

```sh
make build-deployment-backup-linux-amd64 RELEASE_BACKUP_OUTPUT=/absolute/path/deployment-backup
```

Run the cold empty/populated Postgres restore population with:

```sh
make test-deployment-backup
```

That lane checks exact account, Founder, Company save, event, leaderboard and epoch identity. It
also proves a non-clean target is refused. Corrupt/truncated envelopes, wrong age identities,
wrong manifests, wrong epoch bytes, partial output, missing/late populations and incomplete or
out-of-bound objective measurements have focused negative tests.

DP-C does not claim the 6-hour RPO or 4-hour RTO from component execution. The helper exposes a
measurement validator, but an observation is invalid until incident time, restore start and the
authenticated post-restore Caddy smoke completion are all present. Those bounds become release
evidence only during the exact-manifest R-006 rehearsal.
