# RFC: Monotonic Founder Attendance

- **Status:** draft — RP-365 reproduced; D-024 contract choice pending. Not implementation authority.
- **Author:** Codex
- **Created:** 2026-10-07
- **Design refs:** `design/02-economy-balancing.md` §1, §2.4 and §9; `design/04-pets.md` companion-care intent.
- **Parent / amends:** archived [Founder Attendance Foundation](archive/founder-attendance-foundation.md), A3/A5; consumers include Pet Adoption/Care, Garden and Minigame Platform.
- **Planning:** investigation in `planning/cosmetic-shop-v1/log.md`, RP-365.

## Problem and evidence

The current clone-only resolver provisionally counts a sub-ceiling gap as attended. Once
that same unresolved gap exceeds `catchup_ceiling_ms`, it subtracts the whole gap. Previously
accepted attendance disappears. This follows A3's prescribed algorithm but violates the
monotonic consumer clock promised by A5 and needed by pet-care watermarks.

Executed evidence at `270a0d67`:

- Pure resolver, unchanged Founder/Company: gap5000→5001 ms changes effective attendance
  6700→1700 ms. Existing separate short/long-gap tests pass and do not detect this transition.
- Real Postgres/server/compiled DOM: manual action → adoption → six-second pause → Feed.
  Adoption records6219 ms; Feed resolves5581 ms and returns `stale pet care transition`,
  surfaced as generic HTTP409. Founder remains4; care does not commit.
- Trace-only healthy flow and a pause after an already-offline adoption both pass. The
  trigger is retroactive classification, not simply waiting, a stale Founder revision,
  or the separate Cosmetic Buy timeout RP-364.

Reproduction is retained in `planning/cosmetic-shop-v1/rp-365-attendance-probe.patch`.
The registered sources are restored; the patch is a diagnostic, not a passing acceptance test.

## Required outcome

1. Once an authoritative command accepts attended time, later sampling or Exit cannot
   erase that interval. Every consumer uses the same clock, not a pet-specific clamp.
2. Genuine dormant time remains offline: absence does not decay pets or award faucet quota.
3. Keep `age_ms` as the sole completed-run sum, replay independent of live Company reads,
   checked exact bounds, and the Founder/Company identity/revision safeguards.
4. Preserve server authority and lazy evaluation. No periodic per-player loop, hidden
   gameplay commands from the client, time-budget increase, retry-until-green or balance retune.

## D-024 — proposed direction, not ruled

Record accepted Company activity durably before attendance-consuming Founder commands
freeze their sample, using one shared replayable classification boundary. This would amend
A3's clone-only mechanism: contact that has already been accepted must not later become offline.
Cosmetic acquire/equip/unequip do not consume attendance and must retain mechanical isolation.

The owner must authorize that boundary or choose a different explicit shared-clock contract.
Before acceptance, specify its transaction/lock ordering, idempotency, refusal and rollback
behavior; Company revision/events and ordinary accrual effects; all consumer bindings; how
Exit uses the same classification; and historical replay/version compatibility. This draft
does not invent an internal command, migration, new cursor or wire discriminator.

Clamping each pet to its old watermark is not a resolution: it leaves faucet/Garden/Exit clocks
inconsistent. Ignoring all long gaps breaks offline behavior. Retrying Feed or issuing synthetic
manual actions from the UI hides the clock defect. None is authorized by this proposal.

## Acceptance required for the eventual contract

- Retain cutoff−1/cutoff/cutoff+1 and successive-command regressions; demonstrate failure on
  today's code and success on the correction, including actual adopted-pet watermarks.
- Real Postgres/DOM reproduction above must feed successfully after the gap and reload the
  receipt's persisted care state. No setup command inserted merely to rescue Feed.
- Genuine 25-hour absence, repeated connections and duplicate/retried commands: no unearned
  attendance, double decay or double quota. Exercise pet, faucet and Garden consumers.
- Both Exit race orders preserve accepted attendance exactly once; stale identity/revision,
  failure injection and rollback tests cover any new multi-stream boundary.
- Replay historical and new runs in both runtimes; no live read in Founder replay. Cover
  shared fixtures, compatibility and ordinary production/accrual consequences explicitly.

## Deviations and open questions

No change to design's offline entitlement or balance is proposed. Changing archived A3's
read-only sampling mechanism is an explicit proposed specification deviation, not a routine
bug fix already authorized by that mechanism. D-024 and the concrete boundary above must be
resolved and independently reviewed before status can become accepted.

## Changelog

- 2026-10-07: draft from executed RP-365 browser/Postgres and deterministic boundary evidence.
