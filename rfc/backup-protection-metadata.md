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
This draft identifies the required contract and unresolved wire/authority choices. It is
not accepted or sufficiently specified for implementation.

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

## Proposed contract — must be completed and adopted before construction

1. Bind the complete protection/classification metadata to the exact payload identity:
   backup/server/manifest/epoch identities, start/completion times, pre-upgrade and
   resolution state, payload length/checksum. A successful payload checksum alone must
   never count as verified metadata. Define the exact canonical encoding and proof.
2. Define where the worker obtains trustworthy verification authority. A publicly
   verifiable signature or a separately protected operator-state commitment are possible
   engineering directions, not adopted mechanisms. Do not reuse unrelated credential
   keys, invent a signing scheme or mount the restore identity into scheduled services.
3. Metadata mismatch/missing authority must be visible and block deletion, rather than
   classify the file as an ordinary expired backup. Preserve the newest-backup and
   unresolved-upgrade rules for every admitted population.
4. Specify atomic publication, crash recovery, concurrency and any governed resolution
   transition. Backup completion cannot leave a valid-looking envelope with an absent
   proof. Resolution needs explicit authority; the current lack of a resolution writer
   is an existing separate DP6 gap, not permission to trust a changed outer flag.
5. Version any changed envelope/proof and define the existing v1 population's treatment.
   Historical backups remain restorable; unverifiable old protection metadata must not
   silently authorize automatic deletion. Do not rewrite committed fixtures/history or
   promise the exact retention schedule before this compatibility contract is adopted.

## Deviations from design

No gameplay change. Any new proof, secret or persistent operator-state authority needs
explicit Deployment specification and designated review, not implied permission here.

## Acceptance criteria

- Intact ordinary and unresolved pre-upgrade populations satisfy the retained DP6 policy.
- Both RP-505 outer-flag edits refuse automatic deletion while preserving the protected
  bytes; representative identity/time/payload/proof changes also refuse.
- Failed publication, interrupted writes, duplicate IDs and concurrent classification
  cannot leave a falsely admitted proof or delete a protected backup.
- Actual old/new encrypted restore and governed resolution/retention paths execute;
  missing verification authority fails closed without an on-host restore identity.
- The exact release bundle proves the complete path under R-006, including negative
  cases. Diagnostic success and component tests alone are not this acceptance.

## Open questions

Engineering author must complete threat model, verification authority, exact schema,
atomic publication and v1 compatibility, and reconcile the existing resolution lifecycle.
Owner adoption/designated review are required before implementation. No branch is selected.

## Changelog

- 2026-10-12: draft from executed RP-505 protection loss; no format/key/policy change.
