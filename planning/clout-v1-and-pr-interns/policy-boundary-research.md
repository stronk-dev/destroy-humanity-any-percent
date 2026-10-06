# R-012 — actual production policy boundaries

2026-10-07. Baseline `d61a5248`; predeclaration `a25c6d5e`.
Go-only characterization, not a repair or AC6 pass. No runtime or policy changed.

## Executed population

All44 profiles traverse actual Evaluate, foundation validation and existing full
Company Encode/Restore. Both branches start byte-identically. Each positive
phase binds actual elapsed/production/banking, remaining burst and cursor to
independent integer policy. The provision oracle predicts run-relative minute
boundaries and integer carry without calling materialization code. Clock cases
require no work and unchanged complete state. No cell is skipped.

| Population | Cases | Different final states | Observed difference |
|---|---:|---:|---|
| Offline cap edges / bank near-cap | 12 | 8 | Cash; five also differ in credit |
| Burst wall-time / expiration | 16 | 6 | Cash only; integer burst/clock/credit bindings hold |
| Provision boundary / carry | 12 | 4 | Cash only; generated counts and remainder agree |
| Same-time / rollback | 4 | 0 | No state change |

The source-pinned [artifact](../../testdata/axis-stack/policy-research-v1.json)
contains all44 original contexts and88 final states,132 actual evaluation
results/receipts and differing top-level field names. There are220 exact
restoration points (88initial +132phase), eleven selected source identities.
Receipts are the actual Go EvaluationResult's JSON diagnostic projection, not
a claim of served wire/transport acceptance.

The admitted x6/owned-PR1 diagnostic uses cash1e4 and rehashed cap1e100. In
provision cases it owns1or10beige-v2 providers, with consistent purchased total.
This is validated synthetic Company state, not a naturally reached player path.
The permits resource remains0; it does not exercise nonzero multi-resource accrual.

## Policy meaning versus rounding — RP-308

For48hours with initial credit0 and no burst/provider:

- One offline call produces86,400,000ms, cash9.9513424e4 and43,200,000ms credit.
- Two24-hour calls produce172,800,000ms total, cash1.89026848e5 and zero credit.

For96hours, one call banks129,600,000ms; two48-hour calls bank86,400,000ms.
The near-credit-cap arms exercise actual saturation instead of pretending an
uncapped bank remains equivalent. The24h+1/2ms cases distinguish production
allowance from floor banking. All outcomes obey the existing per-call policy.

These are not decimal tolerances, an exploit, or authority to change offline
rewards. Clients do not select mode or server time. One absence internally
partitioned and separate reconnect histories are different semantic questions.
CV3/AC6's unqualified "any interval" needs an author-reconciled episode meaning
before a production repair contract. RP-308 records that question separately
from RP-307's actual rounding failure; neither is waived here. The owner was
asked to delegate that wording reconciliation, with existing rules preserved;
no answer has been received, and no normative body was edited.

The other ten cash differences occur without the offline cap/history mismatch.
Example: provision count10/online60000ms yields3.38602047247e8 one-shot versus
3.38602047246e8 split. Both branches produce exactly one beige tower and the
same remainder. Keeping integer policy correct does not repair rounding.

## Discrimination and verification

```sh
make test-go GO_PACKAGES=./production GO_TEST_FLAGS='-run TestAxisPolicyBoundaryResearch -v -count=1'
# Explicit complete re-observation only:
make test-go GO_PACKAGES=./production GO_TEST_FLAGS='-run TestAxisPolicyBoundaryResearch -v -count=1' UPDATE_POLICY_RESEARCH=1
make test-save-integration SAVE_TEST_PACKAGES=./production SAVE_TEST_FLAGS='-run TestAxis.*Research -v'
```

Five separate faults in the NEW observer fail semantically before writing:
ignore credit saturation; retain expired burst; drop generated count/carry;
advance a zero-work cursor; remove one profile. These are observer faults, not
production mutations. Source and artifact restore byte-exactly afterward.

Cold production/economy/decimal remains RED only at the original27 AC6 failures;
production38.605s, economy6.095s, decimal.228s. Relevant vet passes. Unchanged
client suite9499pass/340visible skips; types/Svelte zero errors/warnings.
Declared native service executes all six research tests with no skips; the old
SQL population remains1215/16complete, every old corpus reproduces unchanged.
The new policy observer itself is Go-only even when executed in that service.
No whole/hosted/AMD64 CI pass is claimed; existing environment/history holds stay.

## Next contract work and limits

An accumulation repair must preserve purchased-only factor timing, minute-grid
provision materialization/carry, burst wall-time, hardcaps, bank floor/saturation
and the accepted absence policy. A constant-rate scalar anchor is insufficient
as the entire production representation. Full-state restoration in this wave
does not persist a new anchor or prove an old-save activation/migration.

Next empirical work: actual paired Go/TS logged replay with these producers,
action/buff/mode boundaries, nonzero multi-resource production, content/clock
pins and Service/Store persistence. Mixed provider/burst/over-cap intersections
and sub-millisecond server-clock cases are not covered by this44-case census.
RP-308 body reconciliation remains held; no episode state is invented in code.

Entire new span afterd61a5248 needs Claude's designated review, including
predeclaration and following records. Earlier spans remain independent. Full
nine-tier/platform1.0 remains active; no checkbox, status, mint, archive, push,
deployment or owner release decision follows from this observation.
