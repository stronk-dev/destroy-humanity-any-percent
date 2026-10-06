# R-012 — frozen-anchor and reconstruction comparison

2026-10-07. Baseline `b54fc7ef`; predeclaration `cadd7111`.
This is test-only representation research, **not a repaired production engine,
accepted save format or AC6 pass**. The original failing regression is untouched.

## Executed method and population

The anchor retains a canonical initial balance and frozen source/efficiency/cap
inputs. A mere evaluation recomputes from that anchor through the unchanged
accrual primitive; it does not use the previously rounded visible balance as a
new starting balance. Each evaluation serializes/restores a complete experimental
snapshot. The action arm settles visible state on a debit/rate change and creates
a new anchor. Neither this field set nor this action policy is adopted here.

The [new artifact](../../testdata/axis-stack/anchor-research-v1.json) records 603
ordinary/domain/golden/near-cap cases, twelve action/cap rows, sixteen restore
negative payloads, three Go-only diagnostics and ten selected source hashes.
It binds the original first-wave corpus, both new instruments, the actual old
Go carry helper, the unchanged primitive goldens and selected arithmetic sources.
This is not complete transitive or release-artifact provenance.

| Predeclared population | Go | TypeScript | Result |
|---|---|---|---|
| Original frozen/capped rows | 520 complete | 520 complete | Anchor one/split agree with recorded actual Go one-shot cash; rebase negative differs45 |
| Numeric-domain edge rows | 64 complete | 64 complete | 60 accepted; four upper-domain/max-time invalid intermediates refused, never hidden by caps |
| Original primitive goldens | 16 complete | 16 complete | All eleven valid literals match; all five invalid inputs refused |
| Original rate/debit and cap/debit rows | 12 complete | 12 complete | Split/restored phases agree; zero differences from the first-wave boundary cash in this population |
| Near-cap rows | 3 complete | 3 complete | All project to cap; one rebase branch still remains below it; post-debit anchors reset |
| Restore negatives | 16 rejected | 16 rejected | Includes valid-but-wrong state/input, framing, duplicate/order, missing/version, size/count/time/domain cases |

Total primary population:615 per runtime, comprising606 accepted/nine refused.
There are sixteen additional restore refusals. The TS file has635 tests:615
primary,16 refusal,three post-debit checks and one metadata/source census.
No row is omitted or downgraded to a skip. No SQL, browser or whole-Evaluate
implementation runs on the anchor representation.

The maximum serialized size across all1215 valid observed snapshot instances
is223bytes. The512-byte gate is a conservative bound of this restricted schema
(up to three source strings), not a measured performance allowance or a claim
that every production resource/source context fits it. No enormous integers
are materialized. Sampling the exponent edges and maximum exact milliseconds
does not prove every possible numeric input or elapsed-history policy.

## Near-cap and carry-codec findings

Starting at9.99999999999e3 with cap1e4, over2000ms:

- Rates4e-9,5e-9,6e-9 all yield visible1e4 under one-shot/anchored split.
- Rebasing at1000ms with4e-9 yields9.99999999999e3 instead: this is one
  additional near-cap disagreement, separate from the45 original frozen rows.
- After visible debit11.3, all new anchors start9.9887e3. At1000ms the first
  two remain9.9887e3; the third projects9.98870000001e3. No new cap absorption
  policy is inferred from these existing-primitive projection observations.

The unchanged first-wave Go carry decoder admits wire0/residue1 and a valid
snapshot followed by another JSON value. It refuses the exact below-cap quantity
9999.999999998 because its projection is1e4 with residue-1/500000000. These
three diagnostics explicitly show why first-wave arithmetic agreement did not
prove a complete codec. They are Go-only observations of a research prototype,
not production save defects or a reason to rewrite the earlier evidence.

## Discrimination and reproduction

```sh
make test-go GO_PACKAGES=./production GO_TEST_FLAGS='-run TestAxisAnchorResearch -v -count=1'
make test-client typecheck
# Explicit new observation only after every declared cell/control succeeds:
make test-go GO_PACKAGES=./production GO_TEST_FLAGS='-run TestAxisAnchorResearch -v -count=1' UPDATE_ANCHOR_RESEARCH=1
```

I deliberately made each decoder overwrite its reconstructed wire with the
serialized wire. Go fails at wrong-valid-rate admission before generation; the
artifact cannot be overwritten. TS fails BOTH semantic wrong-valid-rate and
wrong-valid-wire tests, in addition to the self-source hash check. Thus those
tests catch the reconstruction defect itself, not only changed source bytes.
All temporary test faults were restored byte-exactly; both old and new artifact
hashes stay unchanged. An initial test-authoring compile error and malformed
combined test invocation are disclosed in the append-only log, not counted as
product findings or completed measurements.

Restored cold client:9434passes/340visible skips; types/Svelte zero errors or
warnings. Relevant vet passes. Cold production/economy/decimal remains RED only
at the original27/128 AC6 partition failures; both research tests pass. The
separate focused acceptance run also executes128 and fails27. No complete or
hosted CI claim, acceptance checkbox or archive follows.

## What this can and cannot authorize

An unchanged-primitive frozen anchor survives the declared arithmetic and codec
population without arbitrary-precision accumulation. That is a credible candidate
for further integration research, not yet the production contract.

The next empirical seam is **live rate producers**: this experiment starts from
already canonical source strings. Real Rates multiplies counts and slot factors
before accrual, retaining intermediate float precision. Serializing those raw
rates into twelve-digit strings may alter legitimate one-shot results; that must
be measured, not assumed away. A candidate may need a frozen producer context
rather than a quantized rate list. No such new context is invented here.

Likewise this codec requires canonical JSON byte order and whitespace. Real
Postgres saves use jsonb; a logical SQL round-trip must be tested before claiming
that this decoder can consume stored saves. The valid ephemeral round-trip here
is only the prototype's own serialization. SQL normalization is not a reason to
weaken numeric/input/recomputed-wire validation or to claim migration coverage.

Still unproven: full Company state/receipts, server clock, mode/24-hour offline
cap and banking, boost/provision boundaries, multi-resource identity/content pins,
old-save activation/migration and deterministic production replay. Any follow-up
RFC must name those boundaries and explicit save/kernel implications. This wave
does not rule owner/author choices, change K3, relax AC6 or authorize implementation.

Claude must independently review the entire new span after`b54fc7ef`, including
predeclaration, code/artifact, records and first-filter edge. Older spans and
holds remain independent. Full nine-tier/platform1.0 is active; nothing is
archived, minted, pushed, deployed or called release-ready.
