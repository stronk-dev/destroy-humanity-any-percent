# Permits & T3 gate implementation plan

Current checkpoint,2026-11-04: Permits was ratified and promoted by the reviewed
epoch-6 mint; it is not awaiting another mint or the same owner hashes. These
historical milestones are distinct from the remaining post-mint acceptance,
author-body reconciliation and designated full closeout. No new playable tier
or current whole-CI/release claim follows.

- [x] Assemble byte-exact inactive economy and routes candidate fixtures.
- [x] Land the single bound cap-reason copy row as an intentional orphan.
- [x] Prove Go/TypeScript loader parity and doctrine-route composition.
- [x] Prove neutral/non-neutral multiplier, offline, and genuinely overshooting near-cap behavior through one Go/TypeScript vector.
- [x] Prove the two-requirement gate rejects without a partial debit and atomically debits both resources when satisfied.
- [x] File candidate SHA-256 manifest for owner ratification.
- [x] Owner ratified economy/routes/copy-source byte-pins2026-08-08; generated-copy
  hash is a point-in-time composition attestation, not a fourth immutable byte-pin
  (RFC candidate manifest and the designated re-review's N1).
- [x] Original candidate implementation reviewed: Claude's2026-08-08 verdict covers
  `{7d9cb37,90633a6,d30ab9e}`. This is not approval of later supplements or archival.
- [x] First Content Epoch atomically promoted the candidates and appended epoch6;
  its2026-08-10 designated mint verdict is recorded in the First Content log.
- [x] Post-mint exact epoch-6 initialization supplement: `TestFirstContentEpochInitializesFreshFounderWithFullSet` passes cold 2026-10-07; permits/legal departments start at zero. First Content plan/log owns the test-only range; no persisted-genesis claim.
- [x] Post-mint epoch-5→6 Exit supplement: `TestFirstContentEpochExitPreservesOldResourceUniverse`
  retains the ending run's old universe and initializes the new one. Fresh cold execution and
  corrective/test-tool approval2026-11-04 are recorded in the First Content log.
- [x] Post-mint controlled old-run replay: that pure test and
  `TestFirstContentEpochPersistedBoundaryIntegration` pass independently2026-11-04.
  This is later corrective evidence, not retroactive mint-day execution or all-history proof.
- [x] Shared two-requirement replay: `TestPermitsGateReplayCorpus` and
  `client/test/permits-replay.test.ts` pass independently2026-11-04; complete
  `a42e680e..ac3e3111` has corrective/test-tool approval. Five payment/refusal arms use
  ratified candidate composition, not a full release bundle or feature acceptance.
- [ ] Finish the complete post-mint acceptance sweep, RP-027 author reconciliation,
  canonical docs and designated full closeout before archival.

These post-mint supplements change no active balance artifact or epoch seed;
the original candidate-only scope must not be read as saying Permits is inactive.
