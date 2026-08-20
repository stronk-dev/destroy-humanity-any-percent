# Permits and First Content Epoch lifecycle audit

Coordinate: product tree `190a4fa`; planning checkpoints through `bf8b4d4`; 2026-08-20.
Scope: acceptance, executable evidence, commit/review coverage, plan truth, canonical docs, and
archival eligibility. This audit changes no product, balance, RFC body, or authored copy.

## Verdict

Neither active RFC is archival-ready, but for different reasons.

- **Permits is implemented and live.** The candidate implementation range is designated-reviewed,
  the epoch-6 activation range is designated-reviewed, epochs 7 and 8 retain the capability, and
  current cold Go/TypeScript/Postgres gates pass. Its remaining blockers are real closeout defects:
  the required two-resource Go/TypeScript replay row is absent, the normative body still
  contradicts PT-C1/PT-C2/PT-C4, and canonical docs still describe the pre-mint catalog.
- **First Content Epoch was genuinely minted and reviewed.** Epoch 6 and its historical bundle
  remain loadable, the mint/remediation/baseline range received the cross-party verdict, the
  review's persisted-Founder obligation was discharged, and the later cache-masked fixtures were
  cold-repaired and reviewed. It still cannot close transactionally: AC4's exact verdict/range
  citation requirement is not met by `changelog/epoch-6.md`; AC5's range-head ruling did not
  reconcile the normative body; its plan still describes the designated mint review as pending;
  and the fixed epoch-6 activation witness was later changed into a deploy-current epoch-8 test.

This is not permission to flip completion boxes. The audit commit is not part of either historical
implementation/review range, and repository law forbids completion-by-record-inference.

## Permits authority and review chain

| Stage | Evidence | Result |
|---|---|---|
| Candidate implementation | `7d9cb37` | Initial designated verdict: NOT APPROVED; four findings. |
| Candidate remediation | `90633a6`, `d30ab9e` | F1-F4 remediation, including shared overshoot vector and atomic two-resource debit. |
| Candidate designated gate | `planning/permits-and-t3-gate/log.md`, 2026-08-08 verdict recorded by `88e2054` | APPROVED over the complete candidate set `{7d9cb37, 90633a6, d30ab9e}`. |
| Owner byte ratification | RFC candidate-manifest ruling, 2026-08-08 | Economy, Routes, and copy-source byte pins ratified. |
| Atomic activation | `3ff34bf` | Epoch 6 copied the ratified Routes bytes and composed Economy bytes into production. |
| Activation designated gate | FCE mint verdict recorded by `40aaad0` | APPROVED over `{3ff34bf, 84cf570, c41b388, 08c995e}`. |
| Cold-cache repair | `19b218e`, `e8f7b4e` | Stale epoch-5 test assumptions repaired; separate designated verdict APPROVED. |
| Current retention | epochs 7 and 8 plus active epoch-8 artifacts | `company.permits`, `generator.legal_dept`, and `gate.t3_to_t4` remain live. |

The range union covers the implementation that exists. It does not manufacture the missing
acceptance witness or reconcile contradictory authority text.

## Permits acceptance

| AC | Current state | Executable and historical evidence | Blocker |
|---|---|---|---|
| AC1 | **Proven integration** | Candidate Go/TS loaders, route-resource and doctrine-route composition, literal 16-artifact load, epoch-6 source/production identity, current cold `./production ./replaycatalog ./economy ./routes ./doctrine ./epochseed`, `make test-client`, and `make verify-schema` all pass. The pre-Permits category set is a demonstrated rejection in both runtimes. | None for behavior. Docs/body closeout remains. |
| AC2 | **Proven integration** | Shared overshooting hardcap vector is probe-discriminating in Go and TS; production tests cover neutral and all-target multiplied rate, 24-hour offline policy, and cap reason. Current cold Go and full client suites pass. | None for behavior. |
| AC3 | **Partial** | Go proves insufficient Permits leaves both balances unchanged and success debits Cash plus Permits atomically. The required Go-authored/TS-consumed replay row for that exact two-requirement crossing is absent. The existing cross-runtime `gate.t3_to_t4` doctrine fixture constructs a cash-only gate; the First Content Exit fixture is a Tier-1 wind-down. | RP-028: add the exact shared replay witness under accepted authority and designated review. |
| AC4 | **Behavior/review proven; normative criterion contradicted** | Candidate chronology/schema/route gates and designated review are present; FCE activated the bytes atomically and its mint range was designated-approved. The body still requires a separate pre-FCE `BALANCE-CHANGE:`, which PT-C1 explicitly ruled impossible. | RP-027: ruling author reconciles P2, P4, and AC4. |

The PT-C5 plan also asks for explicit fresh-genesis, real cross-epoch Exit, and old-run resource-
universe witnesses. The epoch-6 range contains catalog-driven activation and historical replay
evidence, but not the promised exact two-resource cross-runtime row; no unchecked box is promoted
by this audit.

## First Content Epoch authority and review chain

| Stage | Evidence | Result |
|---|---|---|
| Candidate composition | `3530b08`, `d0dccb4`, `a0ca14e` | Sixteen literal rows staged; designated candidate verdict `b0277a1` APPROVED. |
| Composed pacing report | `821dfb9` | 300-run report produced; designated verdict `270e97b` reproduced it byte-identically and recorded the unsimulated-content limitation. |
| Owner mint sign-off | `a9e8e88` | FCE5.6 authorized the epoch-6 mint. |
| Mint/remediation/baseline | `3ff34bf`, `84cf570`, `c41b388`, record `08c995e` | Atomic mint, replay repair, and separate required baseline. |
| Mint designated gate | verdict recorded by `40aaad0` | APPROVED full range; F1 persisted-Founder test obligation registered. |
| F1 closure and dependent archives | `7b48a9d`, `5c20ee3` | Persisted extension state became probe-discriminating; Meters/Achievements/Pet Care archive batch APPROVED. |
| Cold-cache correction | `19b218e`, `e8f7b4e` | Four stale packages repaired and designated-approved under genuinely cold execution. |
| Historical preservation | current replaycatalog tests | Epoch-6 artifact hash still reconstructs and loads; epoch 7 remains loadable after epoch 8. |

## First Content Epoch acceptance

| AC | Current state | Executable and historical evidence | Blocker |
|---|---|---|---|
| AC1 | **Proven integration** | Mint commit registered 16 artifacts at `sha256:1a4463bc...f21a`; manifest rows pin source SHA; production/source equality was independently reviewed; current cold replaycatalog/harness suites reconstruct and load the historical epoch-6 bundle. | None for behavior. |
| AC2 | **Historically proven; current witness drifted** | At `3ff34bf`, `activeFirstContentBundle` pinned epoch 6 and the activation/fresh-Founder tests exercised its exact bundle. Later epoch mints renamed it `activeContentBundle` and now assert/load epoch 8, so HEAD no longer provides an epoch-6-specific activation regression. | RP-032: restore a fixed historical epoch-6 activation fixture without weakening deploy-current tests. |
| AC3 | **Proven with preserved historical fixtures** | Epoch-5 byte fixtures remain pinned; replay corpus and loaders retain historical decoding; current cold production/replaycatalog/client suites pass, and the epoch-6 loader test demonstrates later mints do not erase prior artifact resolution. | None for the original acceptance; retain fixed historical witnesses. |
| AC4 | **Partial / literal criterion unmet** | The composed report and its designated verdict exist. The changelog's candidate rows cite one exact reviewed set, but several rows name only a log, a generic “archived verdict,” or a verdict commit without the exact reviewed range. It therefore does not cite every consumed verdict *and reviewed range* as AC4 requires. | RP-030: resolve every artifact to a specific verdict entry and exact reviewed range; obtain review for the record correction. |
| AC5 | **Proven under the accepted range-head interpretation; body unreconciled** | The mint alone was intentionally red until replay remediation and the separate baseline. The designated review accepted green at range head; later cold-cache fixture defects were fixed and separately reviewed. Current cold Go/client/Postgres relevant gates pass. FCE5.5 and AC5 still say “at the mint commit,” while only an appended reconciliation note says range head. | RP-031: ruling author edits the normative body to the accepted range-head contract. |

## Current cold evidence at this audit coordinate

- `make test-go GO_PACKAGES='./production ./replaycatalog ./harness' GO_TEST_FLAGS='-count=1'`
  — PASS (`production` 4.758 s, `replaycatalog` 5.759 s, `harness` 35.667 s).
- `make test-client` — PASS (39 files passed, 2 skipped; 6,655 tests passed, 15 skipped).
- `make test-save-integration SAVE_TEST_PACKAGES='./production ./gameserver'` — PASS against
  real Postgres (`production` and `gameserver` Integration tests, `-count=1`).
- `make test-go GO_PACKAGES='./economy ./routes ./doctrine ./epochseed' GO_TEST_FLAGS='-count=1'`
  — PASS.
- `make verify-schema` — PASS.
- `make copy-check` — PASS (208 keys; 161 declared orphan warnings; content manifest check green).

These runs prove current mechanical behavior. Historical commit/range claims come from exact Git
objects and the named cross-party verdict entries; they are not inferred from current greenness.

## Closeout order

1. Ruling author reconciles Permits P2/P4/AC4 and FCE FCE5.5/AC5; route FCE's post-mint epoch-7
   harness material to its actual successor record instead of leaving the bounded mint RFC open by
   scope accretion.
2. Add the missing exact Permits two-resource Go/TS replay witness and restore a fixed epoch-6
   activation witness under accepted RFC authority.
3. Correct canonical Economy/Routes/Purchasable/Pitch-era docs and make the epoch-6 changelog's
   consumed verdict/range coordinates exact.
4. Run cold root gates and obtain the mandatory self-filter plus designated cross-party review
   over the new implementation/record span.
5. Only then perform one transactional closeout: docs canonical, statuses implemented, RFC and
   planning directories archived, index/backlog/log reconciled in the same reviewed change.
