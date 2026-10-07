# RP-311 — CV5 applied receipt projection

2026-10-07; runtime range AFTER3868e3d2, predeclared79285a14; record-placement
correctione55487fb, actual red tests03f513ef. [V] means locally executed,
not designated review, complete Clout acceptance or release proof.

## Implementation boundary

Go and TS authoritative applied receipt snapshots now emit EXACT CV5:
`{input_kind,input_value,input_cap,cap_reason_key,saturated,contributions:
[{source_id,upgrade_id,factor}],product}` for Company19. Ordinary commands,
refreshed offers and Exit's actual new-run receipt use the same producer.
Go propagates projection errors through all existing callers; TS throws.
No silently omitted/null object or inferred default factor on derivation failure.

Input_value is the RAW Company input, factors use the existing clamped input,
saturated means strictly ABOVE cap, owned sources sort in byte order and
product uses the SAME live contribution arithmetic/order. Decimals remain
canonical strings. Legacy/non-axis receipt bytes are not retrofitted. This
is not the GameUI feature arm with extra attained/intern rows. Kernel166 in
all three mirrors lands with the changed runtime receipt semantics.
No save/schema/migration/event/input/gameplay arithmetic/catalog/balance/copy/
public API schema/CI/accepted body changes or historical kernel-guard repair.

## Tests-first and full-output re-observation

[V] Baseline62911/54c115 compiles and fails all7 Go ordinary/Exit cases on
the missing object.97291/1f49b2 full client:30new failures(24ordinary/6Exit),
9647pass/340visibleSkip. The independent literal receipt oracles genuinely fail.

[V] After repair69500/6ea5a2 passes7Go cases including actual refresh; expanded
63021/e00d25 passes the bounds/ownership/contrast/error-propagation table.
Five direct projections cover empty/one/two owned sources, exact cap44 and
synthetic input45 clamped to44; latter is explicitly NOT admitted save evidence.
Contrast score input and typed ordinary/refresh errors also execute. TS's
additional admitted latched inventory covers both owned interns at exact cap:
factors2.1/1.88, product3.948, NOT saturated; this is not natural progression.

[V] Before fixture refresh6817/303fcc fails60 existing complete receipt
comparisons,9617pass/340skip; all31 new declaration checks pass. This is the
expected observable receipt-byte change, not permission to remove old oracles.
Actual Go generators execute logged-policy24, terminal6pairs, activation6
sequences (`49893/198aac`) and separately action/buff/mode16sequences
(`81335/99f19e`). First selector mistakenly named an absent sequence test;
that first run claims ONLY its three actual populations, not all four.

[V] `86d922` compares ALL four generated reports to3868e3d2. After allowing
ONLY source_sha256 and the new object in applied receipt objects/canonical
JSON mirrors, full reports match: populations/payloads/frozen inputs/bundles/
all Company/Founder states/events/refusals/acceptance labels unchanged.
228 actual object/JSON mirror additions; no expected gameplay output restamp.
All recorded source hashes recomputed. Initial read-only comparisonbd5619
hit Node's default1MiB capture buffer on the2,083,215byte sequence report;
it was not a complete comparison. Larger read buffer supports the measured
file size, not an altered experiment budget or acceptance bound.

Company migration sourceSHA alone becomes
`09ad7c14be117b54f68c38ab8aa139a51422a1543220b6404672443136edd338`.
All11legacy/4Founder/five Company cases, baseline20 and old Founder source
stay unchanged. Prior b185b69d… artifact is the90e102a8 migration checkpoint's
historical evidence in Git, not relabelled as the new producer's observation.

## Actual compiling fault controls

| Fault | Result |
|---|---|
| Go omits receipt axis object | 58791/62cab1 exit2; all7 actual paths plus direct projections fail on absence. |
| TS emits unowned source rows | 62204/5f774b exit2;90fail/9588pass/340skip. New30ordinary/Exit assertions fail; exact-cap two-owned case and census survive correctly. |
| TS omits receipt axis object | 56425/d3d2c5 exit2;91fail/9587pass/340skip, including31new object assertions. |
| Go forces contribution factor1 | 50851/54ff4b exit2; ordinary/one/two/cap/above-cap/contrast fail, empty and all6 empty-new-run Exit cases survive correctly. |
| Go noops actual receipt refresh | 8006/ad0869 exit2: actual ordinary purchase exposes stale input6 instead of8, and refresh-error propagation also fails. Six Exit cases/direct valid projections survive. |

First Go factor probe303c5f failed compilation (`factors` unused); NO runtime
discrimination claimed from it. Corrected probe reads the real factor then
forces1, preserving compilation. All faults restored AFTER terminal handles.
No artifact regeneration around a mutant. Surviving populations are disclosed,
not treated as failures or removed. Source-restoration hashes rechecked.

Self-inspection also tightened the happy direct-refresh input to a stale,
valid-shaped empty axis object. Normal36679/e7e1d2 passes. The extra noop probe
reveals stronger existing coverage: ordinary ApplyLogged itself uses refresh
AFTER the attainment hook, so its earlier assertion fails on6 rather than8,
BEFORE reaching the seeded direct-refresh assertion. Thus the hypothesis that
the whole ordinary/refresh test tolerated a noop was too broad; only the
isolated unchanged-input assertion could. No independent failing seeded-direct
assertion claimed from this probe. All source restored, final70636/84e6d9
passes both full receipt tests .211s; full Go vet1e7415 passes again.

## Final commands and outcomes

Root, normal Make commands, cold Go; TWO native DB jobs MUST be serial:

```sh
make validate-migrations
make test-save-integration SAVE_TEST_PACKAGES='./production ./gameui' SAVE_TEST_FLAGS='-run Integration'
make test-go GO_PACKAGES='./save ./production ./economy ./decimal ./kernel' GO_TEST_FLAGS='-count=1'
make test-client
make typecheck vet build-client verify-ci-topology
```

[V] 14630/07464c actual native migrations save.591s PASS. Then19214/e3dafe
actual production16.438s/gameui.181s Integration PASS. Native PG16.15/arm64;
not hosted/AMD64 or clean-host deployment. Orphan warning gives no cleanup authority.
[V]17416/07898b full client9678pass/340visibleSkip,104passfiles/22skip.
88456/92b73c types/Svelte0errors/warnings, full Go vet,213module build, CI
topology13negative controls PASS. Focused cold16189/b76780 save/production/
kernel runs ALL20 migration cases, new receipt cases and all4report populations.
[V]48431/b0ab92 cold host save.269s/economy6.338s/decimal.229s/kernel.070s
PASS, production47.083s fails ONLY original27AC6 full-state partitions.
No tolerance, gate weakening/new skip or wholeCIgreen claim. Durations are not
performance evidence. Historical RP-131 remains an independent CI failure.

[V] Re-executed `make verify-kernel-version`91511/294dd4: checkout-contract
and adversarial-fixture stages pass; history guard remains RED at historical
`50a3a5141d5c4c1cd25cbf3c9c39d9e3cf8e2444` (RP-131), not waived by this
real165→166 bump. Guard fails BEFORE later worktree validation; no claim it
passed the new range. Mirrors/same-commit runtime change checked separately.
Final8a3b1a recomputes all four source maps and migration SHA;900469 verifies
both append-only prefixes and unchanged corpus except Company sourceSHA/baseline.

## Remaining authority and scope

Locally repairs RP-311, not fullCV5/AC3/Clout acceptance. AC3/DG-B literal
identical-career-receipt contradiction stays with its author; adding CV5 does
not reconcile it. R-012 representation/RP-308 offline episode meaning, owner
content/mint, all other product/platform/manual/release obligations remain.
ENTIRE new span after3868e3d2 INCLUDING predeclaration/placement correction/
red tests/runtime/source observations/canonical records needs Claude's
designated independent pass. Older review spans remain independent. No plan
checkbox/RFC status/archival/mint/push/deploy/release call or goal completion.
