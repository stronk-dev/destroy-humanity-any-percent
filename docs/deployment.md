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
mounts only current file-backed secrets (database URL, JWT, bootstrap and the public cursor secret,
whose ID is pinned to the policy name `k1`). `compose.rotation.template.yml` adds previous JWT/bootstrap pairs
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
gameserver, deployment-backup, deployment-release, deployment-operations, deployment-rehearsal and
deployment-browser
commands (not the much larger `go.sum`
graph), adds the Go standard library, and derives the browser inventory from the built client's
own sourcemaps (`client/dist/**/*.map`): every shipped module under `node_modules/` resolves to its
exact installed package directory, so transitive packages the bundler inlines (currently `pad-end`
and `tslib`) are attributed, and an aliased version (`break_infinity.js` 2.2.0, not the 1.3.0 that
the notations package also installs) is attributed as shipped. The client must be built with
sourcemaps first; a missing build fails. It reads shipped LICENSE/COPYING bytes directly
(whitespace-normalized so wrapped clauses match), recognizes only the audited 0BSD, MIT, ISC,
Apache-2.0, BSD-2-Clause and BSD-3-Clause family, preserves multi-license modules as SPDX `AND`
expressions, and fails on missing, ambiguous, unknown or metadata-mismatched licenses.

The outputs are `third-party-licenses.txt` and an SPDX-2.3 JSON document with package-manager purls,
download locations and root `DEPENDS_ON` relationships. On the current graph the generator finds 47
linked third-party Go modules, the Go standard library and three browser dependencies (51
dependencies total),
matching the prior license audit while retaining the previously hidden dual Apache-2.0/MIT libyaml
notice. Version, full commit and RFC3339 creation time are explicit inputs; an existing output
directory is never silently overlaid.

The application SBOM inventories the union of dependencies linked into all six shipped Go
binaries plus the bundled browser client. The assembler requires separate SPDX inputs for
Alertmanager, Caddy, the gameserver image, node-exporter, Postgres and Prometheus and binds each
SBOM hash beside that image's immutable digest. The clean-host rehearsal additionally binds the
Linux/amd64 Playwright image, its selected runtime-config digest and its SPDX inventory as a
rehearsal-only image; it is not a production Compose service.
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
deployment-operations, deployment-rehearsal and deployment-browser binaries, the
gameserver's `docker save` archive, built
client, generated application metadata, release version/full source commit, tested Docker
Engine/Compose versions, six digest-pinned image references, their linux/amd64 config digests and
their six image SBOMs, plus the immutable Playwright Linux/amd64 reference, runtime-config identity
and image SBOM used to run the browser driver. It
rejects a non-ELF or non-amd64
binary, client symlinks, an absent SPA entry point, empty/missing inputs, mutable image references
and a pre-existing output tree.

The resulting directory contains the runtime content closure, site, gameserver, backup, release,
operations and R-006 rehearsal helper binaries, the offline gameserver image archive,
Docker/Caddy/Compose inputs, the strict deployment and rehearsal schemas, operations rules/templates, root and
third-party licenses, eight SBOM documents and
`release-manifest.json`. The manifest records the current migration, both save-schema versions,
epoch/copy/constants identities and the SHA-256 of every other bundle file. Validation re-walks the
directory and rejects any missing, extra or changed byte, including attribution or an image SBOM.
It then binds the manifest's own claims to those bytes: every service image in the hash-bound
`compose.yml` must equal the manifest reference of that name (the backup worker runs the Postgres
reference), and the runtime closure re-derived from the bundle's `content/` through the same epoch
authority must match the staged file set exactly and carry the manifest's epoch ID, constants hash
and copy hash. Removing a catalog therefore fails even if the manifest is rebound. The database
migration is compiled into the gameserver and is bound at runtime by the private database
inspection, not by bundle bytes.
For rollback, an exact retained pre-browser bundle may carry its original manifest schema without
the later optional `rehearsal_images` property, but only when its manifest also has no rehearsal
image, browser helper or Playwright SBOM. The current source schema and every browser-bearing
candidate still require that property; all other required schema fields remain mandatory.
The assembler records the current company and founder save versions in a new bundle. Validation
also accepts a retained previous bundle's lower positive versions (never a future version), so
a later save-version bump does not invalidate the exact previous manifest before rollback. The
governed rollback still requires the pre-upgrade backup, previous image and live smoke proof.
It intentionally describes a release *candidate*: designated approval of the implementation
batches and the exact clean-host R-006 rehearsal remain required before the project
can claim supported self-hosting.

The bundled `deployment-rehearsal` command validates retained release-build records and final R-006
evidence. Its `run-plan` lane executes a reviewed, manifest-bound command plan directly—never via a
shell—into a new empty result directory. The plan must enumerate the 42 non-forgery positive and
negative populations in canonical lifecycle order, require exit zero for positive checks and exact
exit three for severing checks, and give every command a one-second to four-hour guard. Every row
is bound to its producer: the command must be an absolute path to the `deployment-rehearsal` tool
itself running the population's own subcommand (`install-candidate`, `run-browser`,
`recover-empty`, `recover-populated`, `lifecycle-release`, `lifecycle-rollback` for their positives,
`restore-non-clean` and `restart-admitted-work` for those runtime negatives; `probe --population=<row>`
for every
other population). A constant command such as `/usr/bin/true` or `/usr/bin/false`, a probe for a
different population, or a generic failure exit therefore cannot stand in for proof. Populations
that no producer can exercise yet (currently the release-drain, rollback, rotation, alert-delivery,
journal-budget, provider-off and non-clean-target runtime rows among others) are bound to `probe`,
which exits `2` for them, so a real run fails loudly until each gains a producer. Output is capped at
one MiB per check; timeout, truncation, wrong exit, unsafe shell/sudo invocation, secret-shaped
argument, prior result byte or non-monotonic observation fails the lane. The final evidence binds
the reviewed plan hash, so an operator cannot silently substitute a shorter command list.

Canonical order follows the real state dependencies: host preflight, initial install, browser,
empty and populated restore, incident recovery, release, rollback, rotation, operations and
provider-off supply chain. Refusal fixtures inside stateful release and rollback steps run before
the successful transition consumes their authority. A correct population name assigned to the
wrong step, an alphabetic reorder, duplicate or omission is invalid. Result files remain named by
population and are matched by exact name; filesystem directory sorting is not execution order.

The stateful rows consume one strict mode-0600 scenario-input document. It contains only the run
identity, candidate/previous bundle paths, distinct work/artifact/install-operator/lifecycle-operator/
backup/metrics paths,
the private age-identity file path, public origin, receiver-health URL, public age recipient and
server UUID and bounded operator ID. Unknown/trailing fields, relative or nested mutable paths, unsafe URL identities,
missing/symlinked directories and a group/world-readable age identity reject before orchestration.
Secret values are not scenario fields: production Compose continues to read its existing mounted
secret files, and the reviewed plan contains only the scenario document path.

`make deployment-rehearsal-observe-host REHEARSAL_SCENARIO_CONFIG=/absolute/private/scenario.json`
validates both bundles, directly records Linux/amd64 distribution/kernel and Docker/Compose
versions, and requires the exact Compose project and Postgres volume to be absent. It also rejects
checkout metadata anywhere in the configured run surface and nonempty optional identity, mail,
AI, payment or cloud-monitoring credential variables. Its only output is the exclusive typed
`host-observation.json`; running this producer on another platform or a dirty host cannot create a
clean-host claim.

`make deployment-rehearsal-install-candidate REHEARSAL_SCENARIO_CONFIG=/absolute/private/scenario.json`
then invokes the production install controller against the exact candidate. The controller uses the
candidate's Compose/image/config preflight, starts the clean Postgres and stack, reconciles epoch
and artifact identity, and performs authenticated Caddy smoke. Its append-only state is isolated in
`install_operator_state`; the rehearsal retains `install-ledger.jsonl` only after strict decoding
proves a single successful candidate manifest/version/six-image identity. The installed stack stays
up for the following browser, backup and recovery populations.

The helper's `probe` boundary distinguishes the subject result from probe setup. Exit `0` means a
positive subject passed or a negative fixture was unexpectedly accepted; exit `3` is reserved for
a successfully prepared named negative fixture that reached its gate and was rejected; invalid
input, setup failure and an unimplemented population exit `2`, and usage or validation errors exit
`1`. This prevents a missing file,
unsupported check or broken probe from satisfying a negative row just because it failed. The
current fixed probes cover the intact six-image/SBOM/license/provenance population, eight exact
bundle mutations, three production config/secret matrices, a public-metrics-route severing,
seeded source/image secret detection, the clean-host/source-checkout evidence boundary, alert and
journal evidence severings, incomplete/guarded observation, the RPO/RTO limits, four encrypted
backup-envelope failure populations, a wrong epoch declaration (packaged `current_epoch_id`
changed, manifest rebound, rejected because the content no longer derives the manifest identity),
an irreversible/Down migration (the production release controller, given the exact bundle pair but
a candidate migration below the running release, refuses at `compatibility` with no runtime step,
after the unmodified pair is shown to reach the runtime), and a missing previous backup or image
(the production `VerifyRestoreInputs` and `Prepare` gates accept the intact inputs, then refuse a
removed pre-upgrade envelope and an absent previous image before anything destructive). The
rotation,
alert-delivery, journal-budget and provider-off runtime positives, still have no producer and exit
`2`.

`lifecycle-release` (`make deployment-rehearsal-lifecycle-release`) requires the exact single
candidate install row in the install operator state and an empty lifecycle operator state. It
removes the installed candidate stack, installs the exact previous bundle into the lifecycle state,
and releases the candidate over it through the production controller; the lifecycle ledger must then
hold exactly that install and the succeeded release. `lifecycle-rollback` rolls the candidate back
to the previous bundle with that release's own pre-upgrade backup (header re-read and bound to the
previous manifest before the controller runs), then retains the three-row lifecycle ledger and the
decoded backup header as the operator evidence final validation binds. A successful rollback row
keeps the release it rolled back from, like failed rollback rows.

`restore-non-clean` runs on the installed candidate stack. It first proves the live database is
real and populated by taking a recovery backup of it, then restores that backup in place without a
clean volume. The backup tool's clean-target gate must refuse with its dedicated
`non_clean_target` error class and the live database's recovery identity must be unchanged; only
then does the row exit `3`. An accepted restore exits `0`; any other failure (including a refusal
for another reason or a changed identity) exits `2`. The probe removes its own recovery backup.

`restart-admitted-work` SIGKILLs the installed gameserver instead of the governed stop, observes
readiness and the exit code, and feeds them through the same drain derivation the release helper
uses. The row exits `3` only when that derivation classifies the restart as not a bounded drain
(no gameserver-owned `503`, no courtesy frame, non-zero exit) and the stack is then restored to
readiness with a passing authenticated smoke; an unrestored host is a setup failure (`2`).
The bundle mutations remove the catalog, client, root license, config or release helper, or change
an image digest, runtime-config digest or image SBOM. The config matrices use the production
startup decoder and require every missing/malformed secret, duplicate key identity/value and
invalid origin/proxy variant to reject. The public-metrics probe first rebinds its changed Caddyfile
in the manifest so it reaches the semantic route validator rather than passing on an unrelated
hash mismatch. Bundle mutation uses a private temporary hardlink tree, never edits the retained
bundle, and a cleanup failure invalidates the outcome. The typed host, alert, journal and objective
negatives exercise their production evidence validators; they are falsifiers, not claims that a
clean host or runtime operation has been observed. Positive runtime, browser, recovery, rotation
and operations populations remain DP-F2 work and are not inferred from these package checks.

Bundle-mutation probes (`removed_*`, `changed_sbom`) first require the unmutated candidate to
validate, so an already-invalid bundle yields a setup error (exit `2`) rather than a false
rejection. After the mutation they rebind every manifest artifact hash (and a mutated image SBOM's
hash), so the release gate can reject only on meaning: a missing required artifact, content that no
longer derives the manifest's closure, or an SBOM whose subject no longer names the image's config
identity. A mutation caught only by a stale byte hash therefore no longer counts.

The seeded-source probe scans a valid tracked-file fixture; the seeded-image probe hides its
sentinel in a file inside a gzip-compressed layer blob, the shape BuildKit's `type=docker` export
produces, rather than relying on a malformed archive or a flat outer member. Each requires exactly the scanner's
named sentinel finding before the no-secrets gate may produce exit `3`. A parser/setup error is
therefore not accepted as secret-detection evidence.

The backup-envelope probes first create and restore a valid encrypted production envelope. They
then independently require truncated and corrupted payloads, the wrong age identity and a wrong
release-manifest identity to fail through the production restore gate. The interrupted-writer row
injects a dump read failure only after valid bytes were consumed and accepts the negative only when
creation fails and leaves no completed or temporary backup file. These are envelope and atomicity
falsifiers; the empty/populated database restore and non-clean target checks still require their
real Postgres populations.

For a release run, `make release-secret-scan` receives the exact candidate manifest hash, candidate
gameserver archive and an exclusive result path. It scans the real Git-tracked population and
image. The image scan opens every outer archive member, gunzips compressed layers, walks plain tar
layers and nested gzip/tar files inside them (to a fixed depth and a 1 GiB decompressed bound),
and matches each file's bytes and tar entry names once, at the innermost readable form. A zstd
member, corrupt gzip, truncated tar or depth/size overflow fails the scan closed instead of being
skipped. It derives the source commit from Git, and writes a strict mode-0600 result containing counts
and identities but no paths or content. Final evidence rejects a source-only scan, findings,
incomplete/guarded state, a different manifest or source commit, unknown fields and output
overwrite. The ordinary developer invocation may still print a non-evidentiary scan result without
creating an R-006 artifact. The root Make target resolves a repository-relative result path to
an absolute path before invoking the scanner; the scanner itself still requires an absolute path.

`make deployment-rehearsal-supply-chain` derives the supply-chain result by reopening both exact
bundles, both independent-build records and the structured secret scan. It requires the candidate
record's schema-v2 Playwright closure, matching source/manifest identities, six production images,
one rehearsal image, eight SBOM documents and root plus site attribution. Every verifiable
build-record field is reduced to bundle bytes: the gameserver archive and rebuild-archive hashes
must equal the bundle's `images/gameserver.docker.tar`, the rebuild manifest hash must equal the
manifest, and every image and rehearsal image (reference, runtime config, SBOM hash) must equal the
manifest's, which `ValidateBundle` binds to Compose and SBOM bytes. `independent_rebuild` and
`normalized_sboms_equal` describe the build procedure and remain operator-attested. The emitted mode-0600
result binds the hashes of every input. Final validation also retains and opens both build records,
so a rehashed supply-chain summary cannot invent an independent rebuild or detach itself from the
candidate and previous manifests.

Full bundle validation re-runs the semantic validators for the rendered Compose topology, public
Caddy routes and gameserver Dockerfile after checking manifest byte equality. Assembly-time
validation alone is insufficient because a later re-signed bundle must not be able to retain valid
hashes while changing a public route or container boundary.

Final `validate` is not structural JSON validation. It requires the evidence file, exact reviewed
plan, exclusive per-population result directory, exclusive retained-artifact directory, exact
candidate bundle and exclusive two-file seal directory. The
plan hash must match the `rehearsal_plan` artifact; run and both manifest identities must agree;
the result directory must contain exactly one mode-0600 result for every canonical plan row and no
extra entry. Each result's name, kind, step, command hash, expected exit, completed/non-guarded
state, observation interval and raw file hash must agree with the plan and its evidence population.

The artifact directory contains exactly fifteen mode-0600 retained inputs/outputs: candidate and
previous manifests, their independent-build records, separate candidate-install and lifecycle
release ledgers, the rotation ledger, backup header, host and recovery-objective observations,
browser result, alert delivery, journal observation, secret scan and supply-chain result. Every raw
digest must equal its evidence
row, and both manifest files must also equal the run's top-level identities. Each of the eleven
step intervals and input/output hashes is derived from the ordered command and raw-result hashes
of the plan rows assigned to that step. A plausible evidence JSON with invented artifact hashes or
handwritten step summaries is invalid even when its standalone shape is correct. The three tool
identities must equal the executable bytes of `deployment-rehearsal`, `deployment-release` and
`deployment-browser` in that candidate bundle; a changed, non-executable or symlinked tool rejects.
The browser and journal artifacts are decoded through their owning production contracts after
their hashes match: unknown/trailing fields, an incomplete browser outcome, a different candidate
manifest, an invalid journal budget/retention observation or a time interval outside the run all
reject even if the forged bytes are rehashed into the top-level evidence.

The host and recovery-objective artifacts are also strict timed observations. Their decoded host
and objective values must equal the corresponding top-level evidence field by field (with times
compared as instants) and lie inside the run. This makes Linux/amd64, clean-start, provider-off,
restored identity, RPO and RTO values derived claims rather than editable summaries. The base
dossier therefore binds sixteen artifacts including the reviewed plan; the final sealed dossier
binds eighteen after adding the immutable base and forgery proof.

Operator records are semantic evidence, not arbitrary JSONL. The candidate-install ledger must be
the single exact candidate success. The separate lifecycle ledger must be the exact previous
install, candidate release and previous rollback in order, with matching versions, manifests, six
image identities, operator, pre-upgrade backup and seven-day rollback authority. The decoded
backup header binds that backup to the previous manifest and epoch. Rotation evidence contains the
JWT, bootstrap and cursor activation/removal pairs and cannot shorten any governed overlap. Scratch
failure ledgers used by negative fixtures are not retained as successful authority.

The 43rd `forged_successful_evidence` population is sealed after the 42 command results complete,
so it never hashes itself. A base dossier must first pass the full plan/result/artifact/tool
validator. `make deployment-rehearsal-forge-proof` then changes the base dossier's browser-result
hash and succeeds only when the same base validator rejects that exact mutation. The final dossier
binds the immutable base and proof files as two additional artifact hashes; final validation checks
their mode and bytes, confirms the final dossier is only the base plus the seal, and independently
replays the mutation. A proof document saying `rejection_observed: true` is not trusted by itself.

`deployment-operations alert-observe` first validates the exact candidate bundle and executes its
checked-in seven-family rule population with `promtool` from the candidate's digest-pinned
Prometheus image. It then posts all seven canonical alert identities through the private
Alertmanager API, waits for seven configured-receiver notifications, resolves the same identities,
and waits for seven resolved notifications. The retained mode-0600 observation names each family
and accepts only when both states were delivered, the rule fixtures passed, and the exact candidate
manifest and run interval match. Receiver health or a single synthetic notification cannot satisfy
this artifact.

`deployment-browser` is the clean-host browser-driver component under construction. It is a
reproducible static Linux/amd64 Go binary using Chrome DevTools directly; it does not require Node,
Vite, a package install or repository source on the rehearsal host. The browser executable comes
from the rehearsal-only Playwright 1.62.0 Noble image pinned to amd64 manifest
`sha256:02bbb2155cd7109e3e9c741941097ed1608cf8b6fa44ee2595896da2bdc1f471` (runtime config
`sha256:50cbb76d250a50002045a95f484c5f40573cde831adbe40c784052c037e36118`). This image is a
measurement tool, not another running Cloud Clicker service.

The driver navigates the canonical HTTPS Caddy origin in real headless Chromium, clicks the visible
`BEGIN ATTEMPT` and `Fix Computer` controls, then uses enabled player controls to buy generators,
cross the T0 gate, reach the first run-end and continue into run 2. It does not seed the database,
call the intent API directly or advance the product clock. It accepts a result only when the desk
and run-2 state render, bootstrap credentials are retired, durable credentials exist, a WebSocket
handshake completed, the manual and later intent responses succeeded, the gate and terminal were
observed, and no page exception fired. The action and two-hour execution guards fail closed; they
are not pacing acceptance bounds. The strict version-2 result records transition booleans, run
sequence numbers, action/response counts, timestamps and candidate manifest hash in a new
mode-0600 JSON file. It never serializes a credential, Founder/account identifier, request body or
WebSocket payload. A loopback-HTTP flag exists solely for the isolated driver fixture and is not
admissible in the reviewed R-006 plan. The separate populated-recovery step must still observe
the asynchronous verified board row; a completed browser journey alone cannot assert it.
The binary, pinned browser image identity and its SBOM are part of the rebuilt release bundle and
its exact manifest. That construction does not itself mark
`phase0_browser_flow_through_caddy` implemented or observed; the driver must still execute the
product flow through Caddy on the authorized clean host.

`make deployment-rehearsal-run-browser REHEARSAL_SCENARIO_CONFIG=/absolute/private/scenario.json`
derives the candidate's exact manifest and rehearsal-image identities, then runs the candidate's
own driver inside that digest-pinned Playwright image on the Linux host network. The container is
read-only, capability-free, no-new-privileges, and receives only tmpfs scratch plus read-only driver
and evidence-directory mounts. The producer accepts only the exclusive strict result bound to the
same candidate manifest; Docker stdout or a zero exit without that artifact is not browser proof.

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
- the complete temporary envelope is re-read and its header, payload length and checksum verified
  (so a torn or short write never commits), synced, atomically renamed to `<backup-id>.ccbackup`,
  and the target directory is synced so the rename survives a crash.

No unencrypted dump is written to the off-host target. Interrupted reads and writes remove their
temporary output.

Every scheduled round classifies the complete target *before* writing anything:
- **Completed backups:** regular `*.ccbackup` files.
- **The helper's own temporaries** (`.backup-payload-*.tmp` / `.backup-envelope-*.tmp`):
  - *Stale:* unmodified for more than one hour. These are crash debris; the round removes them and
    reports the count.
  - *Active:* possibly a concurrent pre-upgrade backup. The round defers, records no failure, and
    retries in five minutes.
- **Foreign entries:** anything else. A foreign entry or an invalid/corrupt backup records a
  backup failure, which fires the any-failure alert, and the round creates nothing and retries in
  five minutes.

The worker never exits for a blocked target, so a blocked target cannot fill with repeated backups
or drive a restart loop. The backup target must therefore be a dedicated directory; a filesystem
root containing `lost+found` blocks backups until the operator points the target at a
subdirectory.

The plaintext header's payload length and SHA-256 are verified on every read. A truncated or
substituted payload is invalid for schedule, retention and restore, even when age authentication
and the encrypted header copy would still verify.

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

`deployment-backup recovery-identity --database-url-file=/run/secrets/database-url` streams a
canonical identity of the private database without exporting row content. It emits row counts and
SHA-256 identities for the complete public-table population and separately for player ownership,
Founder state, Company state, events, boards and epoch/catalog authority. The version-2 identity
also reports the separate `verified_runs` row count: a projection event by itself does not prove
a leaderboard row. Every table name and
every full `to_jsonb` row contributes length-delimited bytes in deterministic order, so equal
counts cannot hide substituted content. Empty rehearsal databases must have no player-state rows
but must retain epoch/catalog authority; populated rehearsals require all six semantic domains to
be nonempty **and** at least one actual verified run. Recovery evidence compares the complete
before/after identity.

The manifest-bound Docker runtime invokes backup creation, recovery identity and restore only as
one-off services on the bundle's private database network. Scheduled/recovery backups and release
pre-upgrade backups are distinct typed populations: the ordinary recovery path requires
`pre_upgrade=false`, while rollback requires `pre_upgrade=true`. Neither path accepts the other's
header even when every digest, epoch and server identity otherwise matches.

The destructive R-006 database sequence is split into two restart-safe commands. First run
`make deployment-rehearsal-recover-empty REHEARSAL_SCENARIO_CONFIG=/absolute/private/scenario.json`
after the browser population. It requires nonempty player, Founder, Company, event and epoch
identities immediately, then waits only for the asynchronously projected verified run. A six-minute
safety guard covers the verifier's five-minute claim lease and bounded retry cadence; exhaustion
fails before any destructive step and is not a release-latency promise. It captures the populated
encrypted backup, rechecks that the complete semantic identity stayed unchanged during capture,
declares the incident, recreates Postgres from a clean volume, starts only gameserver/Caddy to
migrate and seed an empty player database, and then
backs up, destroys and restores that empty population with exact identity equality. It records a
strict private checkpoint only after the restored empty database passes manifest/epoch/artifact and
semantic checks. The private checkpoint is version 2 because it embeds the version-2 recovery
identity and refuses older event-only board evidence.

Then run `make deployment-rehearsal-recover-populated` with the same scenario. It consumes that
checkpoint, destroys the empty-restored volume, restores the exact populated backup, compares the
complete database identity before any new smoke state can be written, and finally runs the
authenticated Caddy smoke. RPO is derived from backup completion to the declared incident; RTO is
derived from restore start to smoke completion through the production objective validator. Only
then is `objective-observation.json` written. Both lanes are destructive and belong only in the
explicit clean-host release rehearsal, never push CI.

Build the helper with:

```sh
make build-deployment-backup-linux-amd64 RELEASE_BACKUP_OUTPUT=/absolute/path/deployment-backup
```

Run the cold empty/populated Postgres restore population with:

```sh
make test-deployment-backup
```

That lane checks exact account, Founder, Company save, event, leaderboard, epoch and whole-database
identity. It also proves a non-clean target is refused. Corrupt/truncated envelopes, wrong age identities,
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
has no container and that none of its five declared named volumes (`postgres_data`, `caddy_data`,
`caddy_config`, `prometheus_data`, `alertmanager_data`) already exists. An operator who kept a
volume from an earlier attempt—for example retained ACME certificates in `caddy_data`—must move it
aside first; install refuses rather than risk deleting it. It then starts Postgres and the
remaining stack, reconciles migration/epoch/artifact identity and runs the authenticated Caddy
smoke. A post-start failure removes only that exact Compose project's containers and the volumes
this install created;
a pre-start failure is non-destructive. A success is durable only after an append-only `install`
row is synced. If that final write fails, the new stack is removed rather than left running without
operator authority. Failed install rows may be retried after correction, while any successful
install or later release/rollback row permanently closes initial-install authority.

A release runs these gates in order and records the first failed stage without ever emitting a
success row: image load/pull plus runtime-config digest verification (first, because the candidate
config preflight runs inside the candidate image, whose bare `sha256:` reference resolves only after
`docker load`); exact-bundle/config/receiver/free-space preflight; encrypted pre-upgrade backup;
Compose-governed SIGTERM stop; observed
readiness-down, authenticated `server_restarting` WebSocket publication, intent refusal and clean
bounded process exit; candidate startup and forward migrations; exact database migration,
epoch/hash/artifact reconciliation; then authenticated HTTP and WebSocket smoke through Caddy.
Readiness-down means the gameserver's own `503` from `/readyz` while draining; a proxy 502/504 or
a transport error after the process has gone is not accepted. Readiness, the courtesy frame, socket
closure and the drain bound are observed directly. Intent refusal, admitted-work completion and
job/outbox flush are attested by the process contract rather than observed separately: the
gameserver exits zero only after admission closed and admitted requests, background jobs,
relay/outbox flush and transport shutdown completed. The real Caddy population separately
witnesses the exact intent refusal during drain. `restart: unless-stopped`
does not race this sequence: the helper uses `docker compose stop`, not a raw container kill.
Normal release is strictly forward by semantic release version and database migration; an older
version can enter service only through the governed rollback command.

The pre-upgrade backup is rollback authority only as host bytes: after the backup container
reports completion, the release helper re-reads the host envelope, verifies its payload length and
SHA-256, and requires its header to equal the reported header (manifest, epoch, server, pre-upgrade
class). Because rollback is restore-based and never runs a Down migration, a forward migration or
content change is irreversible exactly when that backup is unusable; release therefore refuses at
`preupgrade_backup`, before drain, rather than proceeding without a working rollback.

Every successful release row binds the candidate version, manifest SHA-256, all six image
digests, exact pre-upgrade backup, exact previous version/manifest and a seven-day rollback
deadline. Rollback accepts only those recorded values. Before it stops anything it loads and
verifies the previous bundle's images, runs the previous-bundle preflight, and verifies the restore
inputs on the host: the backup path is a regular non-empty file in the backup target, the age
identity is an owner-only regular file, and the envelope header's payload length/SHA-256, backup ID,
server, previous manifest, epoch and pre-upgrade class all match (`restore_inputs` stage). A
missing, corrupt or wrong backup or identity therefore refuses with the live database untouched.
Only then does it stop the failed stack, remove only the
named Postgres data volume, starts a clean Postgres service, restores the exact encrypted backup,
starts the exact previous bundle, repeats epoch/artifact reconciliation and runs the same Caddy
smoke. There is no Down-migration operation in the rollback interface or command.
An exact failed rollback attempt is recorded and may be retried inside the same deadline; a
successful rollback or any unrelated intervening transition closes that authority.

Key rotation is also operator-driven. `rotation-activate` records new-current/former-current IDs;
`rotation-remove` refuses removal until the governed interval has elapsed: 30 minutes for JWT,
31 days for bootstrap receipts and 366 days for public cursors. The public cursor reader now
exists, but cursor rotation stays inactive. The API policy fixes the names `k1`/`k0` (C20), which
conflicts with the ledger's per-rotation IDs; this is an open DESIGN-GAP in
`planning/api-foundation/log.md`. Its durable timing/config contract is already enforced. The
runtime current/previous key decoder remains the authority for actual values and rejects half
pairs, duplicate IDs and duplicate values; the ledger stores IDs only.

Install, release, rollback and recovery read the operator-state `rotation-ledger.jsonl` before any
Compose command. While both the JWT and bootstrap overlaps are open (latest ledger row for each
family is `activated`), every Compose invocation adds the bundled `compose.rotation.yml`, so a
recreated gameserver keeps both previous keys for the whole governed overlap. The bundled overlay
binds both families together; while exactly one of them is open the helper refuses before any
Compose command rather than recreating the gameserver without that previous key. Operators either
rotate JWT and bootstrap keys together or wait for the single open overlap to be removed. A
per-family overlay is a recorded follow-up, not current behavior. The ledger path must be absolute;
a missing ledger means no open overlap.

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
gameserver's private `/metrics` route, Caddy's dedicated private `:2020` metrics listener,
node-exporter, Alertmanager and itself. The public Caddy site has no metrics route. Caddy's
management API is disabled (`admin off`), so no container sharing a Caddy network can read or
replace the live ingress configuration; template validation rejects any admin listener. Production
template validation rejects a published private port, an operations service on the wrong network,
a mutable image, a non-distinct journald tag or a missing operations mount/config/secret.

The gameserver uses an isolated Prometheus registry. Its bounded signals cover readiness, process
health, HTTP/WebSocket route classes and latency, Postgres reachability and collector success,
pending outbox and dead-letter populations, background-job results and last success, composed
credential cleanup, and named invariant families. User-controlled paths collapse to fixed route
classes. Metric labels and ordinary structured logs exclude credentials, recovery codes, raw IPs,
payloads and account/founder/stream/intent identifiers. No application label is named `job` or
`instance`: Prometheus owns those target labels and would rename a colliding application label to
`exported_*`, silently detaching it from rules. Background jobs therefore use `job_name`, and every
closed-set job's success and failure series is created at zero at startup so the first failure is
an observable `increase()`. Backup, restore and release helpers write
atomic node-exporter textfiles; a later failure does not erase the last successful timestamp.
The HTTP metrics wrapper preserves the underlying writer's WebSocket hijack capability; the
operations-enabled Postgres composition population performs a real Centrifuge upgrade so ordinary
HTTP instrumentation cannot silently disable the realtime path.

The seven blocking alert families cover five-minute ingress/readiness loss, missing/late/failed
six-hour backups, two-minute Postgres or collector failure, storage pressure above 80%, three
gameserver restarts in ten minutes, credential-cleanup failure and dead-letter growth across two
15-second intervals. Storage pressure also fires when the host observation behind it is stale:
node-exporter's `host.prom` textfile older than five minutes (the observer runs every minute), a
textfile scrape error, or an absent host textfile, so a broken observer cannot keep serving its
last healthy values. The checked-in `promtool` population exercises each firing path and is
discriminating: it pins the lower duration bounds (no public alert at 4m45s, no Postgres alert at
1m45s), the six-hour backup deadline (quiet at five hours, firing at seven), the backup
any-failure clause, sub-threshold storage (0.79) and restart (two in ten minutes) populations, the
journal clause, the `cloud_clicker_ready` clause, single-interval dead-letter growth, both stale
host-observation clauses, and fire-then-clear resolution for the public-readiness and cleanup
alerts. Thirteen single-clause/threshold mutations each fail it. The
private-network integration scrapes the real gameserver registry: after one real
`credential_cleanup` failure, Prometheus must report `CloudClickerCleanupJobFailed` firing through
the shipped scrape and rule configuration. Release preflight also sends a fresh nonce-bearing
synthetic alert and requires a healthy configured receiver plus successful delivery, measured as
the rise in Alertmanager's `notification_requests_total` minus `notification_requests_failed_total`.
Any rise in failed requests during the window fails the proof; the notification-level
`notifications_failed_total` is not used because it moves only after all retries are exhausted.
The integration proves the exact nonce reaches the receiver and that a receiver answering 500
fails the proof on recorded failures. Alertmanager metrics cannot attribute a delivery to one alert
family or nonce, so the seven-family delivery record is count-derived, not per-family proof.

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
