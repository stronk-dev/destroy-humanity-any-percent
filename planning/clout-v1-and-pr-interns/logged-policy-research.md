# R-012 — actual logged policy replay

2026-10-07. Baseline `7effff9a`; predeclaration `6141af35`.
Test-only research. **No full Go/TS output parity or AC6 acceptance.**

## Population and observed result

All24 declarations use actual Go ApplyLogged and actual TypeScript applyLogged,
not an exported scalar helper or copied private engine. Company v19 starts with
input6, PR1, beige1, cash1e4 and a test-local rehashed1e100 cash cap. Three
profiles (quiet;1500ms burst;1500ms burst with10purchased beige-v2 and1legal dept)
cross online/offline and3114/59999/120001/90000000ms. The higher-tier diagnostic
inventory is synthetic; existing admission passes without weakening it. It is
not proof of reachable player inventory or a natural acquisition path.

Go builds actual Founder carry, active-play schedule and catch-up evidence, runs
the buy-one-beige intent, validates and restores complete initial/final states.
All24 purchases apply; all initial/final states restore exactly. The final
inventory, burst exhaustion, and numeric positive/zero permit bindings hold.
Nine copied catch-up populations (missing/from+1/to-1 for the three online25h
rows) explicitly refuse invalid replay evidence without changing initial state.
There are57Go full restoration checks, not57persisted SQL transitions.

The [Go-authored artifact](../../testdata/axis-stack/logged-policy-research-v1.json)
contains all24 complete initial/poststates, payloads, replay inputs, receipts,
events,14 selected source identities and exact fixture bundle artifacts.
Replay inputs are actually version12, not an assumed version7. The receipt,
events and poststate comparisons are exact canonical JSON strings.

## RP-309 — valid v19 replay is refused in TypeScript

All24 TypeScript initial full-state round trips pass. All24 actual logged calls
then throw `active-play resolved presence` BEFORE payout/receipt/event comparison.
TypeScript replay.ts's applyLogged presence check tests `wireVersion===18`;
Go tests `WireVersion>=18`, and the TypeScript restore codec already accepts
active-play state at v19. The accepted CV4 v19 extension/replay contract is the
repair authority. This is a real producer integration gap hidden by prior scalar
comparisons, not a claim of24 numeric payout differences.

The negative oracle deliberately requires a catch-up-specific error. Six TS
missing/from-coordinate cases are masked by the earlier presence error; they
remain RED, not classified as successful catch-up refusals. The three to-coordinate
corruptions refuse in replay parsing and preserve initial state. The census test
passes. Thus the new TS file has34tests:4pass,30fail. Existing9499pass/340visible
skips remain; complete client result9503pass/30fail/340skip.

Sample actual Go results in the combined profile:

| Elapsed / mode | Permits | Generated beige | Credit ms |
|---|---|---:|---:|
| 3114ms online | 5.32061455279e-3 | 0 | 0 |
| 120001ms online | 1.40108363411e-1 | 2 | 0 |
| 25h online, explicit catch-up | 2.4e1 | 1440 | 1800000 |
| 25h offline | 2.4e1 | 1440 | 1800000 |

These are observed Go outputs, not paired output agreement, absence-history
equivalence, a new policy or an accepted accumulation representation.

## Discrimination, correction and verification

A23-row Go declaration fails `incomplete logged policy population` before the
writer, even with UPDATE enabled. Forging the TS expected INITIAL state changes
all24 positive failures to exact full-state assertion failures before applyLogged.
That demonstrates the initial comparator, NOT the unreachable final comparator.
Both edits restore byte-exactly to their pre-probe files/artifact. No production
mutation is claimed. Existing catch-up negatives execute in the actual producers.

Self-check found two observer issues: a Ms/MS property typo (not reached because
of the earlier production error), and comparing permits to `0e0` even though
the emitted zero is `0`. Corrected to actual Decimal positivity with zero-resource
controls; completed re-observation then records the corrected source identities.
Expected production outcomes and the runtime were NOT altered to hide failures.

```sh
make test-go GO_PACKAGES=./production GO_TEST_FLAGS='-run TestAxisLoggedPolicyReplayResearch -v -count=1'
# Explicit full Go observation writer, not a TS-success claim:
make test-go GO_PACKAGES=./production GO_TEST_FLAGS='-run TestAxisLoggedPolicyReplayResearch -v -count=1' UPDATE_LOGGED_POLICY_RESEARCH=1
make test-client
make typecheck vet
make test-save-integration SAVE_TEST_PACKAGES=./production SAVE_TEST_FLAGS='-run TestAxis.*Research -v'
```

Final cold Go production41.833s remains red ONLY at original27AC6 cases;
economy6.094s/decimal.114s pass. Client has the30new failures above; types/Svelte
zero errors/warnings and full go vet pass independently. Native service executes
all SEVEN research tests without skips (1.437s), including unchanged SQL1215/16.
The new logged observer itself has NO SQL save/transaction evidence. AMD64,
history/hosted CI and all prior author/content/release holds remain open.

Final SHA256: Go observer4a82177d3c018707c586a9f8d585e4888c233252248b358ccf61bb970ffb6958;
TS observer1b6372f431265dc438d76146371598dbe078767fab8845629535b81c9c09cc86;
artifactb400abe407d891515eaaffd4f07f2843302bcd58eb700dd6ed9d2602f847a58b.

## Next and limits

Separately predeclare the bounded RP-309 runtime dispatch repair under accepted
CV4, including the required kernel version protocol, old v18 compatibility,
missing/unexpected active-play evidence refusals and actual output comparisons.
Do not mix it into this test-only range. Re-observation must identify changed
source lineage and preserve historical failing evidence; do not restamp outputs.
Only after the seam is repaired can this population answer numeric Go/TS parity.
Mode switches, buffs, episode meaning RP-308, full resource/provision boundaries,
accepted accumulation/save migration and Service/Store replay remain unproved.
Original27AC6 failures stay red, not waived. Claude must independently review
the ENTIRE span after7effff9a, including later records; all prior spans independent.
No checkbox/body/status/mint/archive/push/deploy or full1.0 claim follows.

## Subsequent RP-309 correction — separate accepted-CV4 range

The preceding observation is historical atb55f2131 (red source/artifact retained
in Git). Predeclarationff50031e at974c1a45 authorizes ONLY the two ordinary
presence/version predicates. Correcting them requires kernel162→163; it does
not alter scheduler/terminal/numeric policies. Five companion tests use genuine
old pinned v18/v16 populations: exact first v18 receipt/events, missing v18/v19
evidence, unexpected v16 evidence, and v18 with input version4. All five pass
with unchanged state on refusals. All nine catch-up target refusals now pass.

Every positive v19 replay still fails, now later at `active schedule state
mismatch`: the private scheduler independently tests exactlyv18 (RP-310).
A separate exact-v18 terminal guard is a source finding only, not yet an
executed terminal failure. No final-output/numeric parity can be claimed.

Restoring the old ordinary predicate makes31tests fail:24presence failures,
sixmasked catch-up negatives, and the missing-v19 negative wrongly applies
its purchase. Replacing the predicate withfalse makes28fail:24still hit the
scheduler, and all four companion refusals fail (three wrongly apply, v16 fails
at the wrong guard). Both runtime probes compile and restore byte-exactly.

Complete explicit Go re-observation passes24/9; comparison with974c1a45 proves
ALL output/bundle/profile/payload/input/negative-count values unchanged. ONLY
client replay and observer source hashes change. Current artifactSHA256:
f8154b7ef11b0d823cd41af3814265606274d0b1198c8480814f9028c1607b3a.
Original artifactb400abe4 belongs to the preceding research checkpoint, not the
new code identity. Earlier research artifacts remain unchanged.

Cold client9514pass/24fail/340skip (newfile15pass/24fail); independent types/vet/
topology13controls pass. Go production42.813s still fails original27AC6cases;
economy6.241s/decimal.225s/kernel.214s pass. Native runs all seven research tests,
no skips (1.431s), unchanged SQL1215/16. Kernel-history checkout/fixtures pass,
full version history still fails RP-131 at50a3a514; independent version-guard
fixtures pass. No whole CI/AMD64/hosted or release claim.

Next separately scope the actual scheduler correction and terminal replay
population under CV4. This correction/record edge and all earlier spans need
Claude independently; nothing is archived or marked full-Clout approved.
