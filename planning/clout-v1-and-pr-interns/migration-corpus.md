# CV4/AC7 — exact Company migration corpus

2026-10-07; test-only range after `7ca728ca`, predeclared `36924143`;
reader corrections predeclared `44c0f8a8`. [V] below means locally executed,
not designated independent approval or whole-RFC acceptance.

Historical checkpoint at90e102a8/3868e3d2. The later separate RP-311 runtime
range adds the required receipt projection and fully re-observes this source,
updating only its migration sourceSHA. The baseline/results below are retained
as their original observation, not current source-identity claims.
[Current receipt producer evidence](receipt-projection.md).

## Population and actual consumers

`testdata/save-migrations.json` corpus10 adds the five accepted CV4 names.
Baseline20 retains the original eleven legacy and four Founder cases unchanged.
Source is existing `axis-stack/activation-research-v1.json`, SHA256
`b185b69dc7bebdedc6f69f3267a20d1de4a846451a6527c98fcf105102363f3f`,
exact row `veteran/online/3114`. Epoch8 Company18 transitions to an UNMINTED
Company19 fixture, not a newly published epoch. Current constants hash:
`sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6`;
next: `sha256:b3028943704ab5b105876a6b97d2eb7e3f27846a85e61b34670ef27495175f8b`.

| Exact case | Executed boundary |
|---|---|
| company-v18-to-v19-new-run | Public load preserves18; actual Exit creates19 with empty run attainment; both next actions replay continuously and re-attain score2. |
| company-v19-derived-mismatch-rejected | Real post-purchase Company19, score3 replacing derived2: Go structural load succeeds, pinned validation rejects; TS pinned restore rejects. |
| company-v19-field-before-version-rejected | Real old Company18 plus both v19 fields: structural restore rejects. |
| company-v19-attained-not-superset-of-earned-rejected | Empty attained set/score0, earned purchase ID/score2: pinned superset check rejects without a separate score mismatch. |
| pre-activation-run-replays-through-exit | Actual old manual transition, then Exit under the old semantics; terminal remains18, new run starts19 empty. |

[V] Go's external `save_test` reader invokes PUBLIC RestoreState,
ValidateFoundationState, ApplyLogged and ApplyLoggedExit. This puts the cases
in the existing `make validate-migrations` package selector without an import
cycle or production testing export. TS invokes its actual public restore/replay.
Both readers enforce all20 unique names, exact case order/metadata/patches,
source SHA and unambiguous source row. Missing fields, cases or valid companions
cannot silently become a smaller passing population.

[V] Positive cases compare COMPLETE receipt, Company terminal/new save and
Founder carry-output envelope, all three ordered Exit event streams, plus
next-action receipts/events/full saves/restores. Founder replay returns a
partial output projection, NOT a complete save; Go mechanically projects its
actual public fields, sorted sets/slots and all11 v21 extensions. Identity
metadata comes from actual frozen inputs, not expected output. This does not
claim full Founder save encoding. Negative cases compare specific stages/errors,
unchanged input/state and minimally corrected admitted companions.

## Failed instruments, not hidden product findings

- `97094/a32f81`: eleven legacy/four Founder/three new negatives pass; two
  positives fail because the test adapter attempted full save encoding of the
  intentionally partial Founder output (`generators are required`). Corrected
  public-field projection preserves COMPLETE expected output comparison.
- `77033/657ed0`: 9645 client checks pass; new early-field assertion expected
  an invented phrase. The real exact diagnostic is `save v18 fields are not exact`.
  Both corrections were committed before corrected baseline observations.
- Focused pnpm wrapper `60661` emitted no result and was explicitly interrupted
  (`ab5391`, exit130). It supplies NO test evidence; existing full Make lane used.
- Final native jobs `42533/dcc783` and `27272/45b818` were mistakenly overlapped
  against ONE shared disposable DB. Their stream/epoch/outbox failures are
  invalid final observations, not newly inferred game bugs. Same populations
  rerun SERIALLY below; no assertion/population/expected behavior changed.
- Read-only source check `bbc990` used a nonexistent `artifact_hashes` member.
  Corrected `f54834` verifies the actual `source_sha256` map's ALL22 files.
  No file or expected artifact changed around that tool error.

## Demonstrated discrimination

[V] Corrected baseline: Go `46870/788c51` passes save .277s; full TS
`56122/b865b4` passes9646/340visibleSkip. Then:

| Single controlled fault | Go result | TS result |
|---|---|---|
| Delete first Company case | `04d68c` exit2, incomplete census | `26854/377217` exit2, Company and Founder census fail;9643pass/2fail |
| Forge ONLY Company source SHA | `11623/39b215` exit2, source SHA mismatch | `67146/e08ed4` exit2, digest mismatch;9645pass/1fail |
| Omit actual derived-score check | `80733/c0f5f7` exit2, named derived-negative admits invalid state | `89670/b6657c` exit2, named derived-negative fails to throw;9645pass/1fail |

Both compiling runtime mutants change ONLY the real derivation predicate.
Other four Company cases still pass. These are runtime boundary failures, not
compiler failures or stale source-pin rejection. Every mutant/corpus fault was
restored exactly AFTER its handle terminated; no production diff remains.

## Final verification and reproducibility

Run from repository root; serialize the TWO shared-database jobs:

```sh
make validate-migrations
make test-save-integration SAVE_TEST_PACKAGES='./production' SAVE_TEST_FLAGS='-run "TestAxisActivation.*|TestAxisSequencePersistenceIntegration" -v'
make test-go GO_PACKAGES='./save ./production ./economy ./decimal ./kernel' GO_TEST_FLAGS='-count=1'
make test-client
make typecheck
make vet
make build-client
make verify-ci-topology
```

[V] Serial `85332/11f907` native migration population passes save .536s.
Then `91919/1600d6` passes all three selected actual populations/noSkip5.803s,
including activation14faults and persisted16paths/168attempts/168retries/
32conflicts/48faults/16outbox negatives. Native Postgres16.15/linuxarm64;
orphan warning is not cleanup authority. Durations are not performance claims.

[V] Final client `90047/79d652`:9646pass/340visibleSkip,103passfiles/22skip.
Static `73307/57ec45`:types/Svelte0errors/warnings, full Go vet, client213module
build and topology13negative controls pass. Cold host `18820/eb73c9`: save.275s,
economy6.119s/decimal.229s/kernel.174s pass; production37.414s remains RED on
ONLY the ORIGINAL27AC6 interval-partition failures. No tolerance, new skip or
wholeCIgreen claim. `f54834` verifies unchanged legacy11/Founder4/source/baseline15
and source SHA/ALL22 source-file hashes.

## Limits and next lane

These are five actual fixture-level boundaries, not all-version activation,
natural acquisition/default minted journey, complete Clout/CV4/AC6 or 1.0 proof.
No product/save/schema/migration/kernel/catalog/copy/API/CI/RFC bytes changed.
ENTIRE new range after7ca728ca, INCLUDING predeclarations/corrections/record
edges, requires designated Claude review; older spans remain independent.
No checkbox/status/archive/mint/push/deploy/release call is authorized.

Next separately predeclare RP-311's accepted CV5 derived receipt object repair.
GameUISnapshot.features.axis_stack is not applied receipt.snapshot.axis_stack.
AC3/DG-B author contradiction, R-012 representation and RP-308 offline-episode
meaning, all owner/content/platform/release and full nine-tier obligations remain.
