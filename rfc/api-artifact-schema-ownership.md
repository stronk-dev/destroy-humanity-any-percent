# RFC: Artifact Schema Ownership and Kernel-neutral Descriptor Delivery

- **Status:** draft — engineering proposal; owner adoption and designated review pending. Not implementation authority.
- **Author:** Codex
- **Created:** 2026-10-14
- **Design refs:** `design/06-tech.md` (pure boundaries, replay identity); published data/formulas and community tooling in `design/05`.
- **Parent / amends:** API Foundation C17/C18; Gate Predicates and Routes C2's descriptor-export ownership, without changing its evaluator imports.
- **Depends on:** acceptance/implementation of the exact-path scope-exception mechanism proposed in Kernel History Guard Integrity for kernel-neutral watched paths.
- **Planning:** existing `planning/api-foundation/` plan/log. No separate research program.

## Outcome

Complete the served catalog schema set without copying another owner's grammar, adding an API
dependency to the pure Route evaluator, or issuing a false gameplay-version bump for metadata.
This resolves engineering authority, not a new gameplay choice. The existing numeric encoding,
dictionary and empty-response proposals remain separate; this draft does not adopt them.

## Current evidence

At `c0b71434`, fifteen of the nineteen epoch-8 families export complete owner descriptors.
Economy, Routes, Minigame API and Relevance remain. Economy upgrades reuse `routes.ParseConditions`;
the accepted Route C2 and unchanged `make verify-routes-boundary` permit only Decimal as an
internal dependency of `server/routes`. Adding `publicapi` there fails that literal boundary.
Copying the condition union into Economy would create a second authority. The kernel path
registry also watches entire owner prefixes, so a descriptor-only edit cannot honestly signal
changed replay mechanics and cannot currently pass merely because its author calls it neutral.
No exception or companion has been implemented under this proposal.

## Proposed contract

### S1 — One Route-owned schema companion

Permit **`server/routeschema`**, explicitly owned by the Route contract, to export fresh closed
Go descriptors for the complete Route catalog and its shared condition union. It imports
`routes` for owner kinds/constants and `publicapi` for C17 descriptor values. It contains no
handler, evaluator, loader, storage, clock, network or gameplay implementation. The engine
package remains unchanged in dependency direction: `production → routes`, never the reverse;
`routes` still does not import the companion, publicapi or Economy. Keep its current guard.

Economy's full descriptor composes the companion's condition export rather than defining its
own Route arms. Economy owns the version-5-only `axis_at_least` arm and its upgrade-only
placement. Public catalog assembly composes these same exports. No reflection generator,
handwritten JSON authority, open-map/raw-JSON arm or fixture-only union is permitted.

Companion ownership is a narrow exception for this demonstrated pure-package conflict, not
permission to move arbitrary source out of watched paths or invent additional companion
packages. Other owners retain their current export locations. Normative grammar changes still
require the owner's accepted contract and designated review, wherever their descriptor lives.

### S2 — Narrow kernel-neutral delivery

Adopt descriptor-only delivery as a specific permitted scope-exception class, using the
append-only exact-commit/exact-path mechanism and actual independent verdict proposed by
Kernel History Guard Integrity. Do not create another review register or predeclaration ritual.
One coherent C18 review range may carry its declarations, tests, docs and closeout together.

Eligibility requires inspection of **every** changed watched production path: only fresh
descriptor construction/export, with no call from gameplay admission, evaluation, persistence,
replay, state/receipt/event/snapshot production or other runtime mechanics. Run the affected
owner/consumer tests and generation checks, plus byte-preserving current/stored artifact
compatibility and meaningful structural/semantic negative controls. Test counts alone are not
proof of behavior identity. Descriptor changes may affect API metadata; those changes require
the existing generated-wire compatibility checks and explicit API review, not a gameplay bump.

An entry is not valid until its exact committed path set has designated independent approval.
Recorder-only/self-review, an uncited diff, a filename suffix, a whole-prefix exemption or a
commit message saying "metadata" cannot qualify. Runtime loader/validator edits do not qualify:
when admission or mechanics changes, retain the real same-commit kernel version bump. Mixed
batches must account for every semantic path; metadata cannot mask one. No history rewrite,
removal of guarded owner prefixes, reuse of an unrelated correction, or waiver of RP-131.

### S3 — Full export and integration, not another component milestone

Finish Economy's five-version root/row/condition grammar using its existing component exports
and the shared Route companion. The version arm must select its own generator/upgrade/progress
grammar; a permissive cross-version component union is not the complete artifact descriptor.
Retain loader-owned references, ordering, cardinality, Decimal relationships and composed
cross-artifact restrictions. Preserve each owner's original duplicate-member policy and bytes.

Export the complete Route and Minigame API families. Relevance follows only after its separate
integer-token amendment is accepted and verified against its existing exact grammar. Assemble
the actual served epoch family union, not a hand-picked subset. Finish the catalog HTTP reader
and public client re-verification before claiming C18 acceptance. Formula export/registration
and any lawful mint remain separate existing obligations. No release-floor narrowing.

## Acceptance evidence

1. The unchanged Route import guard passes. Companion exports are deterministic/fresh; caller
   mutations cannot alter the owner authority. All shared condition kinds and wrong-arm/null/
   alias controls run through the actual owner and exported descriptor. Semantic refusals stay
   visible rather than being mislabeled as schema acceptance.
2. The scope guard rejects a forged neutral entry, missing/unapproved review, wrong path/range,
   and reuse against a changed commit or an unlisted watched path in a mixed batch. Independent
   diff review determines semantic eligibility; the guard does not infer it from filenames.
   A reviewed exact descriptor-only range passes
   without a gameplay bump. Full history still surfaces unrelated violations, including RP-131.
3. Current and stored catalogs validate without rewriting bytes/hashes; full version roots and
   forbidden cross-version arms discriminate. Registration covers the actual complete served
   artifact set, and ordinary existing HTTP compatibility checks remain mandatory.
4. Before C18 closeout, run the public producer → complete union/HTTP reader → generated client
   → TypeScript re-verifier workflow, with missing/tampered/invalid artifacts refused. Local
   descriptors or a consistent manifest alone do not prove this workflow.

## Adoption requested

Authorize the one Route-owned companion and the narrowly reviewed descriptor-only delivery
class above, together with its named scope-guard prerequisite. Otherwise retain the current
boundaries and explicitly choose a different accepted shared-schema architecture. No choice
has been received; this file and its index row are proposals, not adopted authority.
