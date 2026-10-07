# Actual action, opportunity, buff and mode sequences

Predeclared `9974f1d3` at clean `0d866359`, with explicit instrument corrections
`eabebf80`, `83d553ff`, `eca45c0c` before the respective corrected experiments.
Accepted Clout CV3/CV4, test-only. No runtime, kernel, balance, schema, copy,
CI workflow or accepted RFC change. This is NOT AC6 or full Clout acceptance.

## Population and actual transitions [V]

The unchanged, UNMINTED axisContentBundle carries schema5/Company19/Founder21.
Enumerate all4096 declared Founder UUIDs through actual Opportunities.Spawn;
select the first UUID per effect, not whichever observed result looks good:

| Effect | First selected UUID suffix | First attended spawn (ms) | Prelude commands |
|---|---:|---:|---:|
| active.building | 000000000005 | 3262 | 0 |
| active.click | 000000000009 | 6684 | 1 |
| active.lucky | 000000000003 | 6843 | 1 |
| active.production | 000000000001 | 3850 | 0 |

UUID prefix is `01986666-2000-7000-8000-`. Four effects × wall gaps3114/90000000ms
× mode paths online/offline/online and offline/online/offline gives16 sequences.
Each has eight core commands plus the table's actual online manual preludes:
136 total command attempts,120 applied/16 rejected. Every prelude advances the
real scheduler by5000ms, never forging its attended cursor or pending state.

Admitted synthetic Tier1/run2 inventory: beige99/provider10/legal1; initial
purchased110, historical attainment6, PR1 owned, cash1e9, compute credit5000,
no initial buff/burst. Veteran Founder owns all registered lifetime grants.
These are legal diagnostic initial states, NOT natural player acquisition.

The eight core commands are real buyexact1 beige at the first spawn, claim the
actual pending ID at that coordinate, buyPR2 at +1ms, spend3000 compute credit
at the same time, buyexact1 beige at +500ms, manual count1/window1000 after the
declared gap, another manual after exactly5000ms, and same-time PR2 rebuy.
First four and rebuy are online; the middle three use each declared mode path.
The rebuy rejects exactly not_eligible/owned and changes no state byte.

Actual schedule, catchup and claim-evidence builders feed actual ApplyLogged.
All136 complete receipts, ordered events, pre/poststates and real save restores
match TypeScript's continuous transitions. TS continues its own result state;
it does not restart every action from a Go-provided poststate. Census independently
asserts effect/gap/mode ordering, prelude counts, kinds and exact timestamps.
Actual claim clears pending; buff arms create the expected non-neutral producer
contribution; Lucky pays out without creating a buff. Purchase re-attains to12
without duplicate lifetime earning, PR2 is owned, bank debit2000/burst3000 are
exact, original buff expires, and final beige count is101. Every row produces
positive permits. Short gaps provision0; long gaps provision1440 at the actual
60000ms tick. Evaluation mode alone does not define attended-clock pauses:
the pinned5000ms gap classifier remains binding.

For EACH claim, delete its evidence, forge the effect row, or increment its
next attended coordinate:48 refusals per runtime. Go requires ErrInvalidReplayInputs;
TS rejects; both retain the complete initial state. No excluded failure arm.
TS population is65 declarations: census +16 sequences +48 refusals.

Complete bundle,18 selected source identities and every observed output are in
`testdata/axis-stack/sequence-research-v1.json`, SHA256
`6ec5ed1ba93ab2ebcd6f50f3c7c0aac1458748bfc9823695b3a47b0f5cf81c54`.

## Instrument failures, not production defects [V]

- Initial `18836/779169` and diagnostic `50940/a4dc60`: short row provisions0,
  despite positive permits. The provisional all-row positive-provision assertion
  ignored the actual60s tick. Predeclared correction retains short exact-zero
  controls AND long positive controls; no population removed.
- `4068/be498e`, diagnostic `66723/122058`: final5001ms exceeds catchup ceiling,
  correctly recording offline time and preventing buff expiry. Final delay is
  now the exact pinned5000ms, not a changed policy/duration or weakened expiry.
- `95078/a0a26d`, diagnostic `29978/4df0ff`: direct jump to click's first6684ms
  spawn also pauses attended time. Predeclared actual manual preludes reach it
  through legal short gaps. All commands are observed; count increases128→136.
- Corrected constructor `40767/0bee34` passes0.637s and publishes the FIRST valid
  corpus. No earlier failed attempt published a corpus or restamped outputs.

## Discrimination and restoration [V]

- Drop one row BEFORE TS enumeration: `80026/5fa473` fails exact missing-ID
  census despite only61 declarations being enumerated. No silent reduction.
- Forge the first claim's expected event kind: `54716/a5e385` fails full ordered
  event equality, not a loose prefix assertion.
- Actual Go active contribution result severed to nil: compiling `39951/1e2cc9`
  fails the semantic non-neutral contribution assertion BEFORE source-pin checks.
- Actual TS active contribution result severed: `46736/83ad61` fails10 complete
  receipts (all building/production and both long click arms). Both short click
  arms survive because the1000ms buff expires before their later manual command;
  four Lucky arms correctly survive a buff-only mutation. No coverage claim
  for the surviving combinations.
- Actual TS compute-burst start severed to0: `91166/64b2f0` fails ALL16 new full
  receipts, plus one existing doctrine corpus. Bank spending alone is not proof.

All sources restored exactly AFTER each handle terminated. A rejected patch
context during burst-probe setup wrote nothing and ran no experiment; corrected
target was read directly. No mutant, compiler failure counted as discrimination,
or modified expected output remains. Prior artifacts/source pins unchanged.

## Final checks / CI limits [V]

- Restored full client `33205/ee803b`:9640 pass/340 visible skips.
- `57401/131d98`: zero TS/Svelte errors/warnings, full go vet, client build213
  modules713ms, CI topology and all13 negative controls pass.
- Native Compose `28013/cb57bf`: ALL11 research/integration populations execute,
  no skip,5.476s. Existing SQL1215/16 and14 activation write faults still pass;
  new sequence observation itself is in-memory, not a new SQL population.
- Cold `87688/5dabf9`: production66.536s fails ONLY original27 AC6 cases;
  economy7.401s/decimal.267s/kernel.169s pass. Checks ran concurrently;
  no performance conclusion or bound change follows.
- Separate SHA/census recomputation `6e181d` verifies all18 source identities,
  16 rows,136 attempts,48 negatives and full report identity. Final non-writing
  observer `58247/be57d8` passes0.618s; `e0de35` confirms both append-only log
  prefixes,120 applied/16 rejected and parses every complete JSON observation.
  Existing server/client Actions lanes discover these tests; workflow source
  routing is NOT hosted execution.

No whole CI-green claim: original AC6, historical RP-131 and AMD64/hosted
obligations remain. Sandbox process-list diagnostic was denied; it yielded no
performance evidence and did not replace the terminal test result. Compose's
existing orphan warning is not cleanup authority.

Reproduce the non-writing checks from the repository root:

```sh
make test-go GO_PACKAGES='./production' GO_TEST_FLAGS='-run TestAxisActionBuffModeReplayResearch -count=1 -v'
make test-client
make typecheck vet build-client verify-ci-topology
make test-save-integration SAVE_TEST_PACKAGES='./production' SAVE_TEST_FLAGS='-run TestAxis.*Research\|TestAxisActivationWriteFaultsIntegration -v'
make test-go GO_PACKAGES='./production ./economy ./decimal ./kernel' GO_TEST_FLAGS='-count=1'
```

The last command is expected to expose the retained27 AC6 failures. The explicit
writer `UPDATE_SEQUENCE_RESEARCH=1` belongs only to intentional re-observation;
ordinary checks compare the complete report and never rewrite it.

## Limits / next boundary

Synthetic admitted history and one selected seed per effect, not all seeds,
combo-cap combinations, natural/default DOM progression or minted content.
This population proves continuous logged replay/restores, not live Service/Store
claim/burst transaction, retry, frozen-resource or history breadth. Next separately
predeclare those persisted claim/burst boundaries under accepted CV3/CV4; do not
reuse the earlier Exit rollback proof as if it covered them. Representation,
RP-308 episode meaning and exact AC6 repair remain separate unresolved gates.
All new span after0d866359 INCLUDING planning corrections and record edges
needs designated Claude review. No checkbox/status/archive/mint/push/deploy,
release call or shorter 1.0 scope. Full long-term objective remains active.
