# R-012 — bounded conserved accrual observation

2026-10-07. Predeclaration: `61f01349`, baseline `b6a3c48d`.
Research only; **production AC6 remains RED**. No accepted save representation,
numeric policy, implementation approval or release claim follows from this file.

## Question and executed population

Can a conserved quantity, projected through the shipped twelve-digit numeric
boundary, avoid the partition differences without repricing a prior interval?
The exact population and controls were recorded before execution in
[plan.md](plan.md#r-012-first-wave--conserved-state-feasibility-not-a-runtime-repair).

The new [artifact](../../testdata/axis-stack/partition-research-v1.json) contains
all 520 frozen observations, 12 rate/debit or cap/debit observations, ten existing-
law projection controls, eight rejected-input controls, selected source SHA-256
identities and the original fixture bundle hash. No rows are excluded after
measurement. The test-local cap variants are separately rehashed and admitted;
they do not modify production balance files.

| Population | Go | TypeScript | What was actually exercised |
|---|---|---|---|
| 512 ordinary / 8 capped | 520 completed | 520 completed | Actual Go Evaluate/state encoding; TS actual scalar accrual/ledger projection; separate Rat/BigInt prototypes |
| Rate/debit transitions | 8 completed | 8 completed | Experimental visible-state debit and carry reset, not served Company intents |
| Cap/debit transitions | 4 completed | 4 completed | Saturation drops excess before the subsequent experimental debit |
| Projection controls | 10 matched | 10 matched | Existing quantizer, including signed midpoint and exponent carry |
| Invalid inputs | 8 rejected | 8 rejected | Declared domain/restore refusals, not a production codec security audit |

Thus each runtime executes 542 primary cases plus eight refusals. The client
file has 551 tests because its metadata/source/population check is additional.
Go's current-engine arms compare complete encoded state; TypeScript's current
arms compare cash only. Full served-state Go/TS parity is not inferred.

## Results

- The current implementation differs on 45/520 one-shot-versus-split full Go
  states. All 45 also differ in cash; there are no other full-state-only
  differences in this population. The actual TS scalar branches agree with
  those Go cash observations.
- The current one-shot branch matches the exact once-settled reference cash
  on all 520 rows. This is a population result, not a universal rounding theorem.
- The conserved prototype's final wire **and residue** match its closed-form
  reference and each other on all 520 frozen rows, including serialization and
  reconstruction at each cut. Its 12 action/cap boundary rows also match their
  separately calculated references under the experimental settlement policy.
- Dropping residue at each cut/restart produces 45 disagreements. Its cash
  matches the current split branch on all 520 rows. Future-rate repricing
  disagrees in all eight eligible transition controls. All four cap controls
  have strictly positive pre-cap excess; retaining it fails the zero-carry
  reconstruction check before a later debit can bank it.
- The reference contains 97 negative, 99 positive and 324 zero residues. Carry
  is not simply a nonnegative extra balance. All four whole-second controls
  at cash1e4 agree with the current implementation and reference.

Representative end3114ms/cut1553ms, cash1e4/rate1.15115/efficiency1:
one-shot cash `1.00035846811e4`; current split `1.00035846812e4`;
conserved split `1.00035846811e4`. A lower-order quantity lost at an intermediate
commit explains this case without changing the shipped midpoint goldens.

## Reproduction and discrimination

From the repository root, ordinary cold replay must reproduce the committed
artifact byte-for-byte. Explicit generation is allowed only after every declared
case/control and population check succeeds:

```sh
make test-go GO_PACKAGES=./production GO_TEST_FLAGS='-run TestAxisPartitionResearch -v -count=1'
make test-client typecheck
# Explicit new observation after an intentional instrument/source change:
make test-go GO_PACKAGES=./production GO_TEST_FLAGS='-run TestAxisPartitionResearch -v -count=1' UPDATE_PARTITION_RESEARCH=1
```

I changed only the artifact's acceptance label to `PROVEN`. Go rejected the
changed bytes; TS failed the exact NOT_PROVEN metadata assertion, while its other
550 research tests passed. The original artifact was restored before final
verification. This is a demonstrated false-promotion refusal, not a mutation
of production or independent review of the instrument.

Final cold client: 8799 passes / 340 visible skips; types/Svelte zero errors or
warnings. Final cold production/economy/decimal: economy and decimal pass;
production fails the existing RP-307 property in exactly 27/128 cases. The new
research test passes but does not disable or repair that acceptance regression.
Historical kernel-history RP-131 is separately unresolved; no complete or hosted
CI execution/green claim is made.

## Limits and contract consequences

The domain is exponent -32..32, terminating decimal scale <=48 and integer
intervals <=60000ms. The admitted numeric domain instead reaches exponent
8,999,999,999,999,999. Materializing an enormous rational is not a production
solution. These tests do not establish CPU/storage bounds across that domain.
The thirteen source hashes pin selected inputs/instruments, not every transitive
dependency; deterministic Go reconstruction and the fixture bundle hash are
additional bindings, not a claim of complete release provenance.

The JSON residue is test-only. The first wave does not test a complete canonical
restore codec, trailing JSON, inconsistent visible/residue projections or a
below-cap quantity that rounds visibly up to the cap. Those are explicit next
research cases, not silently accepted production behaviors. The experiments
also do not establish provision/decay integration, multi-resource settlement,
offline-cap banking, real SQL restart, old-save migration or browser workflows.

The separately predeclared [second wave](anchor-research.md) now executes those
prototype framing/projection and near-cap diagnostics without modifying this
first-wave instrument or artifact. It observes two invalid admissions and one
valid near-cap refusal in the old Go carry model; its separate frozen-anchor
comparison extends the sampled numeric range, not production/save acceptance.

A future buildable contract must specify, with executed evidence:

1. A bounded representation compatible with the whole admitted numeric domain;
   compare an unchanged-primitive frozen anchor with conserved carry rather than
   assuming this bounded rational prototype is that representation.
2. Exactly which real input/rate/debit/cap/mode changes settle or reset history,
   including upward rounding at a cap and no hidden overflow bank.
3. Canonical reconstruction, missing/old-save migration, content-pin changes,
   deterministic replay and bounded serialized size. New authoritative fields
   require explicit save/version/kernel implications; none are invented here.
4. Preservation of actual prior-interval pricing and full state/receipt parity,
   not only scalar wire equality or a changed tolerance.

This observation supports further representation research and an author finding.
It does not yet support an implementable persistence RFC or an AC6 waiver.
Claude's mandatory designated review must cover the complete new range after
`b6a3c48d`, including this record and later first-filter edge. Earlier independent
review obligations, owner/API/content holds and full nine-tier 1.0 scope remain.
