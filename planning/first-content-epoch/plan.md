# First Content Epoch implementation plan

RFC: `rfc/first-content-epoch.md`

- [x] Stage all thirteen owner-authored achievement copy rows as intentional orphans and regenerate
  the copy artifacts.
- [x] Stage and content-gate the ratified meters, achievements, pets, permits, and `minigame_api`
  candidate artifacts without touching the active epoch seed.
- [x] Complete the literal promotion manifest and composed harness evidence.
- [x] Obtain explicit owner sign-off for the single epoch-6 mint commit.
- [ ] Mint the complete dependency chain atomically, run every declared gate, receive the
  cross-party designated review, and archive.

The owner-authorized production mint and its original designated reviews are recorded in the
log. This RFC remains open for post-mint acceptance and dependent closeout; the supplements below
cannot archive it on their own.

### Historical regression supplement (2026-10-07)

- [x] Restore the exact sixteen-artifact epoch-6 activation and fresh-Founder tests without
  dropping epoch-8 coverage: `TestFirstContentEpochActivatesAtNewRunBoundary`,
  `TestFirstContentEpochInitializesFreshFounderWithFullSet`, and the two `TestCurrentContent*`
  companions pass cold against their distinct pins.
- [x] Exercise the epoch-5→6 terminal transition and restored replay:
  `TestFirstContentEpochExitPreservesOldResourceUniverse` passes cold, with the ending run's old
  resource universe retained and the new run initialized with permits/legal departments at zero.
- [x] Prove the epoch-5→6 boundary through real Postgres:
  `TestFirstContentEpochPersistedBoundaryIntegration` passes in `make test-game-ui-composed`.
  New connections reload both saved axes, actual epoch pins, immutable genesis/contributions
  and both replay histories; wrong hash/forged receipt reject and duplicate Exit does not mutate.
- [x] Repair RP-366 under Fiscal F8/F13: the first Fiscal period uses the logged Founder clock,
  not the Company's separate evaluation time. `TestFirstContentEpochExitFiscalActivationUsesFounderClock`
  fails on the old comparison for both offsets and passes with the repair, retaining the credit
  divergence rejection. Kernel 0.3.167 records this behavior correction.
- [ ] Designated review of the supplements. Diagnostic integration fixtures are not naturally
  earned progression, clean-host release proof, original mint gates or the full archival union.
