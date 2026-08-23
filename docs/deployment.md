# Deployment

Deployment Foundation is implementing. The repository does **not** yet claim a supported self-host
bundle or release-ready deployment. The config, package, backup, release/rollback and private
operations batches are implemented but remain subject to their required independent reviews; the
exact-manifest clean-host rehearsal remains unfinished.

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

`make render-release-compose` replaces the six image tokens only when each supplied reference is
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

`make generate-release-metadata` inventories the union of module graphs actually linked into the
gameserver, deployment-backup, deployment-release, deployment-operations and deployment-rehearsal
commands (not the much larger `go.sum`
graph), adds the Go standard library, and reads the
three exact browser runtime dependencies from `client/package.json` plus their installed package
manifests. It reads shipped LICENSE/COPYING bytes directly, recognizes only the audited MIT,
ISC, Apache-2.0, BSD-2-Clause and BSD-3-Clause family, preserves multi-license modules as SPDX `AND`
expressions, and fails on missing, ambiguous, unknown or metadata-mismatched licenses.

The outputs are `third-party-licenses.txt` and an SPDX-2.3 JSON document with package-manager purls,
download locations and root `DEPENDS_ON` relationships. On the current graph the generator finds 40
linked third-party Go modules, the Go standard library and three browser dependencies (44
dependencies total),
matching the prior license audit while retaining the previously hidden dual Apache-2.0/MIT libyaml
notice. Version, full commit and RFC3339 creation time are explicit inputs; an existing output
directory is never silently overlaid.

The application SBOM inventories the union of dependencies linked into all five shipped Go
binaries plus the bundled browser client. The assembler requires separate SPDX inputs for
Alertmanager, Caddy, the gameserver image, node-exporter, Postgres and Prometheus and binds each
SBOM hash beside that image's immutable digest.
For upstream multi-platform references it also records the selected linux/amd64 OCI config digest;
the SPDX document name must identify that exact runtime config, preventing a native-host SBOM from
being attached to the supported amd64 release.

`make generate-image-sbom` runs Syft v1.51.0 from an immutable image digest against either an exact
registry reference or the gameserver Docker archive. The repository normalizer retains Syft's
package/file graph while binding the document name and namespace to the selected OCI config digest
and its creation time to the declared release timestamp. It refuses empty graphs, mutable config
identities, or an existing destination. Independent real Caddy scans have produced byte-identical
normalized output; the exact six-image set is still built and retained per release candidate.

## Release bundle assembly

`make assemble-release-bundle` accepts only an empty output directory and requires all of the
following explicit inputs: the Linux/amd64 gameserver, deployment-backup, deployment-release,
deployment-operations and deployment-rehearsal binaries, the
gameserver's `docker save` archive, built
client, generated application metadata, release version/full source commit, tested Docker
Engine/Compose versions, six digest-pinned image references, their linux/amd64 config digests and
their six image SBOMs. It
rejects a non-ELF or non-amd64
binary, client symlinks, an absent SPA entry point, empty/missing inputs, mutable image references
and a pre-existing output tree.

The resulting directory contains the runtime content closure, site, gameserver, backup, release,
operations and R-006 rehearsal helper binaries, the offline gameserver image archive,
Docker/Caddy/Compose inputs, the strict deployment and rehearsal schemas, operations rules/templates, root and
third-party licenses, seven SBOM documents and
`release-manifest.json`. The manifest records the current migration, both save-schema versions,
epoch/copy/constants identities and the SHA-256 of every other bundle file. Validation re-walks the
directory and rejects any missing, extra or changed byte, including attribution or an image SBOM.
It intentionally describes a release *candidate*: designated approval of the implementation
batches and the exact clean-host R-006 rehearsal remain required before the project
can claim supported self-hosting.

The bundled `deployment-rehearsal` command validates retained release-build records and final R-006
evidence. Its `run-plan` lane executes a reviewed, manifest-bound command plan directly—never via a
shell—into a new empty result directory. The plan must enumerate every declared positive and
negative population in canonical order, require exit zero for positive checks and exact exit one
for severing checks, and give every command a one-second to four-hour guard. Output is capped at
one MiB per check; timeout, truncation, wrong exit, unsafe shell/sudo invocation, secret-shaped
argument, prior result byte or non-monotonic observation fails the lane. The final evidence binds
the reviewed plan hash, so an operator cannot silently substitute a shorter command list.

The helper's `probe` boundary distinguishes the subject result from probe setup. Exit `0` means a
positive subject passed or a negative fixture was unexpectedly accepted; exit `1` is reserved for
a successfully prepared named negative fixture that reached its gate and was rejected; invalid
input, setup failure and an unimplemented population exit `2`. This prevents a missing file,
unsupported check or broken probe from satisfying a negative row just because it failed. The
current fixed probes cover the intact six-image/SBOM/license/provenance population, eight exact
bundle mutations, three production config/secret matrices, a public-metrics-route severing and
seeded source/image secret detection.
The bundle mutations remove the catalog, client, root license, config or release helper, or change
an image digest, runtime-config digest or image SBOM. The config matrices use the production
startup decoder and require every missing/malformed secret, duplicate key identity/value and
invalid origin/proxy variant to reject. The public-metrics probe first rebinds its changed Caddyfile
in the manifest so it reaches the semantic route validator rather than passing on an unrelated
hash mismatch. Bundle mutation uses a private temporary hardlink tree, never edits the retained
bundle, and a cleanup failure invalidates the outcome. Runtime, browser, recovery, rotation and
operations populations remain DP-F2 work and are not inferred from these package checks.

The seeded-source probe scans a valid tracked-file fixture; the seeded-image probe scans a valid
tar member rather than relying on a malformed archive to fail. Each requires exactly the scanner's
named sentinel finding before the no-secrets gate may produce exit `1`. A parser/setup error is
therefore not accepted as secret-detection evidence.

Full bundle validation re-runs the semantic validators for the rendered Compose topology, public
Caddy routes and gameserver Dockerfile after checking manifest byte equality. Assembly-time
validation alone is insufficient because a later re-signed bundle must not be able to retain valid
hashes while changing a public route or container boundary.

Final `validate` is not structural JSON validation. It requires the evidence file, the exact
reviewed plan and the exclusive per-population result directory. The plan hash must match the
`rehearsal_plan` artifact; run and both manifest identities must agree; the directory must contain
exactly one mode-0600 result for every canonical plan row and no extra entry. Each result's name,
kind, step, command hash, expected exit, completed/non-guarded state, observation interval and raw
file hash must agree with the plan and its evidence population. A plausible evidence JSON with
invented hashes is invalid even when its standalone shape is correct.

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

## Governed release, rollback and key rotation

`deployment-release` is the manifest-bound operator entry point. It is deliberately not an
automatic deployment service: an operator invokes it with the exact install, current or candidate bundle,
the separately mounted backup and operator-state directories, the canonical Caddy origin and the
private alert-receiver health URL. Database sizing and identity checks run through one-off backup
helper containers on the private database network; Postgres remains unpublished to the host. The operator-state directory is
mode-0700 storage outside either bundle and holds two mode-0600 append-only JSONL records:
`release-ledger.jsonl` and `rotation-ledger.jsonl`. Records contain identifiers, timestamps,
digests and outcomes, never key, password, recovery-code or token values.
Mode-0600 nonblocking lock files serialize release/rollback and rotation mutations so two operator
invocations cannot race the ledgers or the live stack.

Initial installation is a separate fail-closed operation, not a release with a fictional current
version. It loads and verifies all six manifest-bound image/config identities without starting a
service, then checks twice—around candidate configuration preflight—that the exact Compose project
has no container and that the named Postgres volume does not exist. It then starts Postgres and the
remaining stack, reconciles migration/epoch/artifact identity and runs the authenticated Caddy
smoke. A post-start failure removes only that exact Compose project's containers and new volumes;
a pre-start failure is non-destructive. A success is durable only after an append-only `install`
row is synced. If that final write fails, the new stack is removed rather than left running without
operator authority. Failed install rows may be retried after correction, while any successful
install or later release/rollback row permanently closes initial-install authority.

A release runs these gates in order and records the first failed stage without ever emitting a
success row: exact-bundle/config/receiver/free-space preflight; encrypted pre-upgrade backup;
image load/pull plus runtime-config digest verification; Compose-governed SIGTERM stop; observed
readiness-down, authenticated `server_restarting` WebSocket publication, intent refusal and clean
bounded process exit; candidate startup and forward migrations; exact database migration,
epoch/hash/artifact reconciliation; then authenticated HTTP and WebSocket smoke through Caddy.
The clean exit is meaningful because the gameserver exits zero only after admitted requests,
background jobs, relay/outbox flush and transport shutdown complete. `restart: unless-stopped`
does not race this sequence: the helper uses `docker compose stop`, not a raw container kill.
Normal release is strictly forward by semantic release version and database migration; an older
version can enter service only through the governed rollback command.

Every successful release row binds the candidate version, manifest SHA-256, all six image
digests, exact pre-upgrade backup, exact previous version/manifest and a seven-day rollback
deadline. Rollback accepts only those recorded values. It stops the failed stack, removes only the
named Postgres data volume, starts a clean Postgres service, restores the exact encrypted backup,
starts the exact previous bundle, repeats epoch/artifact reconciliation and runs the same Caddy
smoke. There is no Down-migration operation in the rollback interface or command.
An exact failed rollback attempt is recorded and may be retried inside the same deadline; a
successful rollback or any unrelated intervening transition closes that authority.

Key rotation is also operator-driven. `rotation-activate` records new-current/former-current IDs;
`rotation-remove` refuses removal until the governed interval has elapsed: 30 minutes for JWT,
31 days for bootstrap receipts and 366 days for public cursors. Cursor rotation stays inactive
until a public reader exists, but its durable timing/config contract is already enforced. The
runtime current/previous key decoder remains the authority for actual values and rejects half
pairs, duplicate IDs and duplicate values; the ledger stores IDs only.

Build the release helper with:

```sh
make build-deployment-release-linux-amd64 RELEASE_HELPER_OUTPUT=/absolute/path/deployment-release
```

The host-side command reads `CLOUD_CLICKER_SERVER_ID` from the same non-secret operator environment.
Initial install requires `install --bundle` plus `--operator-state`, `--operator`,
`--public-origin`, `--receiver-health-url`, `--backup-target`, `--metrics-dir` and
`--age-recipient`. Release requires `--current-bundle`, `--candidate-bundle`, `--operator-state`, `--operator`,
`--public-origin`, `--receiver-health-url`, `--backup-target`, `--metrics-dir` and `--age-recipient`. Rollback replaces
the two release inputs with `--failed-bundle`, `--previous-bundle`, `--backup-id`, `--backup` and the
explicit off-host `--age-identity-file`; the age identity is rejected if it is not a private regular
file. Rotation uses `rotation-activate` or `rotation-remove` with the operator-state path, family,
current/previous IDs and operator identity. Secret values never appear in these arguments or ledgers.

DP-D component evidence does not claim supported deployment or rollback by itself. The exact
release-manifest clean-host R-006 rehearsal, measured RPO/RTO, operations evidence, and both review
gates remain mandatory before that claim.

## Private operations profile

The release bundle contains a provider-off operations profile: Prometheus, Alertmanager and
node-exporter share an internal Compose network and publish no host ports. Alertmanager alone also
joins the non-publishing edge network so an operator may choose either a local or remote receiver;
Prometheus and node-exporter have no external route. Prometheus scrapes the
gameserver's private `/metrics` route, Caddy's private `:2019` administration metrics,
node-exporter, Alertmanager and itself. The public Caddy site has no metrics route. Production
template validation rejects a published private port, an operations service on the wrong network,
a mutable image, a non-distinct journald tag or a missing operations mount/config/secret.

The gameserver uses an isolated Prometheus registry. Its bounded signals cover readiness, process
health, HTTP/WebSocket route classes and latency, Postgres reachability and collector success,
pending outbox and dead-letter populations, background-job results and last success, composed
credential cleanup, and named invariant families. User-controlled paths collapse to fixed route
classes. Metric labels and ordinary structured logs exclude credentials, recovery codes, raw IPs,
payloads and account/founder/stream/intent identifiers. Backup, restore and release helpers write
atomic node-exporter textfiles; a later failure does not erase the last successful timestamp.

The seven blocking alert families cover five-minute ingress/readiness loss, missing/late/failed
six-hour backups, two-minute Postgres or collector failure, storage pressure above 80%, three
gameserver restarts in ten minutes, credential-cleanup failure and dead-letter growth across two
15-second intervals. The checked-in `promtool` population proves each firing path and the resolved
population. Release preflight also sends a fresh nonce-bearing synthetic alert and requires both a
healthy configured receiver and an increase in Alertmanager's successful-notification counter.
The private-network integration proves that exact nonce reaches the receiver; health alone fails.

All seven services write to persistent journald with a unique bounded tag. Journal capacity is not
a guessed constant: `deployment-operations journal-observe` records the predeclared workload,
interval, sample count, observed bytes, peak bytes/day, filesystem size and proposed budget.
`journal-render` accepts only a complete, non-guarded observation with exact 14-day retention,
capacity for fourteen measured peak days and an alert at 80% before eviction; it writes both the
new journald drop-in and the exact budget file used by host monitoring, refusing to overwrite
either. The shipped stack has no raw-IP producer or enable switch, and the observation validator
rejects a claimed raw-IP security sink rather than trusting a boolean assertion. Adding one later
requires a separately governed producer, exact seven-day purge mechanism and executable witness;
the default journal cannot silently become that sink.

Install `cloud-clicker-observe.service` and `.timer`, copy `operations.env.example` to
`/etc/cloud-clicker/operations.env`, and set its absolute persistent-data, off-host-backup,
journal, metrics and measured-budget paths. Provision the metrics directory so the host release
operator and the backup container's numeric `70:70` user can both write it while node-exporter can
only read it; release and backup preflight fail instead of dropping an observation when that
boundary is wrong. The one-minute observer records filesystem use,
gameserver restart count and journal consumption through atomic textfiles. Build and exercise the
operations helper with:

```sh
make build-deployment-operations-linux-amd64 RELEASE_OPERATIONS_OUTPUT=/absolute/path/deployment-operations
make test-deployment-operations
```

The second command runs the exact pinned Prometheus rule evaluator, cold Go privacy/config/helper
tests and a real isolated Caddy/Prometheus/Alertmanager/node-exporter delivery population. A real Linux
host observation and exact-manifest alert/recovery rehearsal still belong to DP-F/R-006; the
component evidence here does not claim supported self-hosting or release readiness.
