# RFC: Backup Protection Metadata Binding

- **Status:** draft — RP-505; not implementation authority
- **Author:** Codex
- **Created:** 2026-10-12
- **Design refs:** `design/06-tech.md` §Stack (single-node deployment and Postgres)
- **Depends on:** Deployment Foundation DP3/DP6/AC4
- **Parent / amends:** [Deployment Foundation](deployment-foundation.md), backup envelope/retention boundary
- **Supersedes / superseded by:** —
- **Planning:** owning evidence in `planning/deployment-foundation/`; successor plan once accepted

## Summary

Give automatic backup classification a verified binding between protection metadata and
the immutable encrypted backup, without bringing the restore-only age identity on-host.
The engineering recommendation is a signed v2 envelope, using a dedicated Ed25519
writer key and an externally configured public-key trust set. Existing v1 files remain
restorable and can be explicitly authenticated and copied into v2 without rewriting
their originals. This recommendation is not accepted implementation authority.

## Motivation

RP-505's two real encrypted populations retain and restore their intact unresolved
pre-upgrade backups. Changing only the plaintext `pre_upgrade` or `upgrade_resolved`
boolean leaves the ciphertext/checksum intact: `ReadHeader` accepts, `ApplyRetention`
deletes the protected file, and authenticated `Restore` rejects before dump output.
The current checksum binds encrypted payload bytes, not the metadata retention trusts.

Scope: metadata integrity, conservative classification, publication/failure atomicity
and old-backup compatibility. This does not change retention durations, data rights,
the seven-day rollback window, deployment topology or the off-host age identity rule.
It does not claim protection against an operator who can simply delete all backup data.

## Proposed specification — adoption and normative review required

### BM1 — Trust boundary and authority

Protect against corrupted or replaced backup metadata/payload when the signing key and
configured trust set remain trustworthy. A signature does not prove decryptability,
Postgres restorability, freshness or completeness of the target population. It cannot
prevent an operator or attacker with filesystem deletion rights from deleting backups,
nor protect against a compromised writer holding the signing key. These limitations
must appear in the operator docs; R-006 still proves actual restore.

Use Go's standard [crypto/ed25519](https://pkg.go.dev/crypto/ed25519) `Sign`/`Verify`,
not a custom signature algorithm or an age/JWT/bootstrap/cursor key. The private writer
seed is32 bytes, serialized as padded standard base64 plus one newline in a mode0600,
regular non-symlink secret file. Generation uses the library's secure random source;
creation is exclusive and never prints a seed. Reject wrong lengths/encodings before
calling the library. A key's ID is `sha256:` plus lowercase hex SHA-256 of its32 public
bytes; IDs are content-derived, not reusable human aliases.

The separate trust file is closed JSON `{"schema_version":1,"keys":[{"id":ID,"public_key":KEY}]}`;
`public_key` is padded standard base64 of32 bytes. Require a nonempty ID-sorted set,
matching content-derived IDs, no duplicates/unknown fields and no unsupported version.
Its canonical encoding is ordered-struct `json.Marshal` plus exactly one newline,
with the field order shown above. This is trusted operator configuration outside the backup
target, mounted read-only. It must never be discovered from an envelope, target sidecar
or key supplied alongside an untrusted backup. Private seed and trust files have absolute
non-symlink paths; trusted parent directories cannot be group/world writable. Seed file
permissions admit no group/other access; trust files cannot be group/world writable.
An unreadable/invalid trust file blocks mutation rather than silently creating a new one.

The backup worker gets only its writer seed plus trust set; its public key must already
be trusted. The gameserver/Caddy receive neither. Explicit restore/retention need only
the public set (restore additionally needs the existing off-host age identity). Adding
these mounts/flags and validating them in release preflight is part of this amendment,
not authority supplied by the current Compose template. Private seed bytes never enter
the release bundle, logs, operation records, metrics or command-line arguments. Public
key IDs may appear in backup metadata; the trust set remains external operator configuration.

For replacement, provision a new seed and append its public key to the trusted set
before switching writers. Keep verification keys while any retained backup uses them;
do not drop one merely because its writer stopped. This adds no timed overlap to DP3's
three existing key families. Compromise-driven revocation requires explicit operator
handling of the affected backups, not automatic deletion or an invented recovery SLA.

### BM2 — Exact signed envelope

One immutable `.ccbackup` contains, in order:

1. `CLOUD-CLICKER-BACKUP-V2\n`;
2. one canonical UTF-8 JSON line `{"metadata":M,"signature":S}\n`;
3. the unchanged standard age encrypted payload.

`M` is a closed struct with fields in this exact order: `schema_version` (literal2),
`backup_id`, `server_id`, `release_manifest_sha256`, `epoch_id`, `started_at`,
`completed_at`, `payload_sha256`, `payload_bytes`, `pre_upgrade`, `upgrade_resolved`,
`signing_key_id`. Preserve existing domain validation, integer types and mandatory
boolean fields. Dates are UTC `time.Time` JSON encoding. Canonical bytes are Go
`encoding/json.Marshal` of that ordered struct, with its default string escaping,
no omitted fields or whitespace. Readers strictly decode typed fields and require
byte equality with canonical re-encoding; nulls, duplicates, aliases, reordered keys,
extra fields, noncanonical integer/date spellings and missing fields cannot normalize
into admitted metadata. The wrapper likewise round-trips byte-exactly and uses that
field order. Retain the existing64KiB bounded line framing; EOF/truncation fails.

Sign the exact concatenation `CLOUD-CLICKER-BACKUP-METADATA-V2\n` plus canonical `M`
bytes using regular Ed25519, without an additional prehash. `S` is canonical padded
standard base64 of the64 signature bytes. The reader selects only the configured key
whose ID equals `M.signing_key_id`, verifies this message/signature, then checks the
complete payload's exact length/SHA-256. Unknown keys and malformed signatures refuse.
Golden encoding/signing fixtures must fix these bytes independently of generated output.

The encrypted authenticated-header and package-metadata formats remain v1 and retain
all existing semantic fields. Outer v2 adds the proof, not a different dump or a new
meaning for those fields. Restore verifies v2 proof/checksum and the existing age
authenticated header/package equality before `pg_restore`; signature success never
replaces decryption or package/dump validation. Restore continues to support original
v1 files through its explicitly supplied age identity, without asserting verified
automatic-retention metadata for them.

### BM3 — Publication, classification and concurrency

Build the encrypted payload privately, compute its size/checksum, sign the complete
metadata and write the complete envelope to the existing mode0600 temporary file.
Sync/close, verify the exact staged signature and payload against the configured trust
set, then use the existing atomic no-replace commit and directory sync. A sign/verify/
write/sync/rename failure cannot emit completion or advance backup-success metrics.
There is no independently published proof sidecar and no unsigned v2 fallback.

All supported create, scheduled rounds, import and cleanup paths serialize mutation
with an exclusive advisory lock on the opened target-directory inode; a lock failure
refuses. The target-owning entry point acquires the lock once and passes its authority
to nested helpers rather than trying to reacquire it. The supported target filesystem
must demonstrate the lock and atomic no-replace behavior in R-006; unsupported storage
is not silently emulated. Verify the configured path still names that directory and use regular,
non-symlink files. Cleanup enumerates the complete target under the lock, verifies all
backups and computes the existing DP6 policy from verified metadata only. Any invalid,
unknown-key or legacy-unverified backup blocks all committed-backup deletion. Recheck
the selected files' identity/verified bytes before removal; changed files abort. Sync
the target directory after deletion before reporting successful cleanup. A partial
filesystem deletion failure is visible, never reported as an atomic completed pass.
The lock protects cooperating tools, not hostile writers that bypass it (BM1).

Never trust a caller-supplied subset to establish the newest/protected population.
The CLI and scheduled worker both use the target-owning path; pure classification
helpers are not mutation authority. Existing stale-temporary cleanup stays bounded to
the declared names/type/age and runs under the same lock, not as proof of backup validity.

Report legacy-unverified separately from malformed/invalid, without silently excluding
either from the census. A scheduled round blocked by either state creates no backup,
records failure and retains its existing visible retry behavior. An operator must
prepare a valid signed target before switching the worker; mixed legacy data cannot
secretly disable cleanup or claim the exact DP6 schedule is met.

### BM4 — Existing backups and explicit adoption

No automatic signature over a checksum-only v1 header. The proposed one-off
`deployment-backup adopt-v1` command requires explicit source, distinct destination
target, expected release manifest, off-host age identity, writer-seed file and trust
file. The caller explicitly invokes this operation; the schedule never obtains the
restore identity or converts old files itself.

Copy the regular v1 source to a private mode0600 staged snapshot. Fully authenticate
that snapshot through the existing age header and package extraction/dump checksum
checks against the expected manifest/epoch before signing anything. Preserve its
backup/server/manifest/epoch identity, timestamps, protection flags and encrypted
payload bytes exactly; only the outer format/version/key/proof changes. Publish into
the distinct destination using BM3, refusing an existing backup ID. Never modify,
rename or delete the original, refresh its timestamps, mark it resolved, or claim that
cryptographic/package validation performed an actual Postgres restore.

Failure before commit leaves the original intact and no admitted destination. A directory
sync failure after commit is visible indeterminate durability, not a completion; preserve
the original and committed destination, never overwrite on retry. Explicit inspection
must verify the complete destination and successfully sync it before it can be accepted.
Before reconfiguration, inventory the complete original target and reconcile every
committed backup ID, source-file digest and protection class to its authenticated copy;
no selected subset or silently omitted legacy file qualifies. Unreadable/invalid source
entries block this migration gate and need explicit operator disposition, not exclusion.
Complete destination verification plus an actual sampled Postgres restore precede
operator reconfiguration of the schedule to that target.
Disposal of original v1 files is NOT authorized here. Wrong age identity, wrong manifest,
corrupt metadata, incomplete tar/dump, duplicate destination ID and missing signing
authority must all refuse; successful copy and unchanged original bytes are observed.

### BM5 — Protection stays immutable; resolution remains a separate hold

Creation records `upgrade_resolved:false` and never patches a signed backup afterward.
This amendment supplies no resolution writer, expiry shortcut or automatic interpretation
of a successful release as resolution. Explicit v1 adoption preserves authenticated
historical flags; it is not new resolution authority. Unresolved pre-upgrade backups
remain protected indefinitely under the current policy. The existing missing governed
resolution lifecycle remains a separate DP6 author gap, so this repair alone cannot
close all of DP6/AC4 or promise bounded storage for unresolved upgrades.

## Deviations from design

No gameplay or retention-duration change. This explicitly proposes an outer envelope
version and a dedicated on-host signing secret/public trust configuration beyond the
current DP6 envelope and Compose mounts. Owner adoption and designated normative review
are required; no new secret/state authority is implied by an automatic goal continuation.

## Acceptance criteria

1. Fixed BM2 vectors prove exact encoding/signing; valid unseen backups verify using only
   configured public keys. Both RP-505 flag edits and each identity/time/checksum/size/key/
   signature edit refuse deletion with bytes preserved. Seed/key/signature length failures,
   malformed framing/JSON, duplicate/case-alias fields and untrusted embedded keys refuse.
2. Actual target-owned cleanup proves intact six-hourly/daily/newest/unresolved populations,
   full-population refusal, lock contention, file replacement and visible partial deletion
   failure. A subset cannot authorize deletion; bounds remain the exact existing DP6 values.
3. Signing/write/verification/sync/commit failures and concurrent duplicate writers never
   publish a false success or replace a completed backup. Restart cannot admit an unsigned
   or partially published envelope. Relevant regression cases—not every assertion—must
   establish the critical deletion oracle discriminates.
4. Real v1→v2 package adoption preserves original bytes and ciphertext, refuses BM4 negatives,
   and actually restores both generations against clean Postgres16 with identical populated
   relational identity. An unauthenticated legacy population blocks cleanup and scheduling
   visibly; no restore identity appears in continuous Compose mounts.
5. Replacement keys admit retained old and new backups; removed/unknown keys block mutation.
   The mounted worker and release/rollback/rehearsal readers use the same verified boundary.
   Preflight rejects missing/misplaced authority; source/bundle scans reject private key bytes.
6. Affected package/vet and declared Linux/Postgres backup/release tests pass, then exact
   current/previous release bundles exercise install, scheduled backup, protected retention,
   explicit restore and rollback under R-006. No component pass substitutes for this artifact
   proof. Designated review covers the complete implementation and records before closeout.

This does not close the separately held resolution lifecycle, actual snapshot-based RPO,
other Deployment findings or the full1.0 release floor.

## Open questions

Adopt or reject BM1–BM5 as the recommended metadata-integrity repair, including the new
dedicated signing-secret/public-trust configuration and explicit legacy-adoption operation.
If adopted, obtain designated normative review before implementation; adoption is not a
review verdict. The governed resolution lifecycle remains an explicitly separate author
question, not a branch silently selected here.

## Changelog

- 2026-10-12: draft from executed RP-505 protection loss; no format/key/policy change.
- 2026-10-12: complete the recommended v2 signature, trust, atomic publication, target-owned
  cleanup and authenticated v1-copy design. Source-grounded against current Create/Restore,
  package extraction, scheduled rounds, Compose mounts and no-replace rename. Standard
  library signature API checked against its official documentation. Proposal only; no
  product code, operator configuration or new key generated and no acceptance execution.
