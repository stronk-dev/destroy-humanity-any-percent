# RFC: Combat Care Tables

- **Status:** draft — proposal, not implementation or content authority
- **Author:** Codex (for author/owner reconciliation)
- **Created:** 2026-11-05
- **Parent / amends:** Combat Shared Data C5 / AC5; its C3 external-input boundary
- **Design refs:** `design/04-pets.md` §2; source-reading correction below
- **Planning:** `planning/combat-shared-data/`

## Outcome and limits

Specify the missing two-integer care mapping so the parent can implement its
fixture-only boundary honestly. This proposes one table grammar, exact rounding,
composition and a synthetic test fixture. It changes neither care decay nor
Founder Soul, battle stats, damage, stamina, rewards or a live catalog.

The parent remains implementing. C2's complete catalog, effect union, crit policy,
minimum population and registration/hash contract remain separate gaps. Duel
DU-C1 still blocks engine code, and DU-C2–DU-C8 remain deferred unchanged. This
draft does not adopt their rest/debit/event/snapshot proposals by implication.

## Source correction — RP-072

The original research does **not** specify a 50%→30% obedience/disobedience
probability. `design/research/creature-battler.md` §3.4 Finding 4 reports a
simulated **win-rate** sequence 50/44/40/34/30 for obedience settings
1.00/.95/.90/.85/.80. Its §8.3 explicitly labels 50%→30% as win rate. The
research's own caveat calls the experiment a falsifier, not a tuner; its
fallback is another known move, whereas the draft duel's fallback is rest.
Those results cannot calibrate that different engine. This is an inspection of
the committed report, not a reproduction or independent validation of its sim.

`design/04-pets.md` §2 calls those percentages **disobedience** and reverses the
intended reliability relationship. Parent C5 repeats the pair without naming
the measured quantity. Neither statement supplies implementable probabilities.
Do not quietly choose 30% or 50% as a policy endpoint or claim the research
proved the fixture below balanced.

## Proposed contract

### CT1 — Preserve the actual external input boundary

Use Pet Care's existing exact pair `{pet_trust_ppm, soul}`:

- `pet_trust_ppm`: integer in `[0, 1_000_000]`;
- `soul`: integer in `[-9_007_199_254_740_991, 9_007_199_254_740_991]`.

Both are required; reject missing/null/wrong-type/noninteger/out-of-domain
inputs before evaluation, including a directly constructed invalid pair.
TypeScript admits primitive numbers only, not booleans, strings, boxed numbers
or coercion. Reuse the existing Pet Care input validator, without narrowing its
accepted Soul domain, changing saved values or deriving a second Trust value.

C3's int32 storage rule applies to the resulting combat quantities. Its wording
must explicitly distinguish the already-shipped safe-integer external Soul
input from int32 battle state. Clamp that input to the selected table endpoints
**before subtraction or multiplication**; do not cast it to int32. Intermediate
arithmetic below is int64 in Go and BigInt in TypeScript, never floating math.

### CT2 — One exact, immutable care-policy object

The future pinned Combat catalog owns this object under `care`:

```text
{
  obedience_by_trust: {points: [{x, y_ppm}, ...]},
  obedience_factor_by_soul: {points: [{x, y_ppm}, ...]}
}
```

All object keys are required, exact and unique; no extra/null fields or open
maps. Arrays are non-null. Numeric fields use integer JSON tokens in the same
strict sense as the current structural Combat readers: no fraction/exponent
tokens, booleans, numeric strings or out-of-range integers. Duplicate decoded
keys refuse even when their spelling uses escapes. Member order/whitespace do
not change semantic admission or authorize rewriting original artifact bytes.

Each table contains at least two points in strictly increasing `x` order;
duplicates or unsorted points refuse rather than being sorted/repaired.
`y_ppm` is in `[0, 1_000_000]` and nondecreasing; equal adjacent values are legal.
Trust `x` values are in `[0, 1_000_000]`, with first `x=0` and last `x=1_000_000`.
Soul `x` values are signed int32. Tables return independently owned immutable
storage, and invalid input returns an error with no partial policy/result.

The full C2 amendment must bind this object to its versioned artifact and
original-byte content hash. This is not a new artifact family, live pin, API
resource or permission to add a partial catalog endpoint. A snapshot carries
frozen inputs plus the pinned catalog identity, not another editable policy
table; the exact snapshot envelope remains with the deferred child ruling.

### CT3 — Interpolation, composition and bounded arithmetic

For a validated table `T`, clamp `z` to its first/last `x`. At a point return
that point's `y_ppm` exactly. Between adjacent points `(a,A)` and `(b,B)`:

```text
lerp(T,z) = A + floor(((z-a) * (B-A)) / (b-a))
base     = lerp(obedience_by_trust, pet_trust_ppm)
factor   = lerp(obedience_factor_by_soul, soul)
obedience_ppm = floor((base * factor) / 1_000_000)
```

Use the shared integer division helpers outside `combat/`; neither combat
runtime gains native division. There are precisely these two interpolation
floors and the final composition floor, without early ratio rounding.
The result object has exactly `{obedience_ppm}` in `[0, 1_000_000]`.
No silent truncation, fallback to defaults, live policy lookup or mutation.

This admits nondecreasing reliability with Trust and Soul. It does not change
raw stat ceilings or establish a quantitative fairness/engagement result.
It also introduces no crit, insurance, status, learned-trick or stamina rule.
Sampling, draw order, what a disobedient action does and its emitted events
remain engine-owned. An engine cannot treat this probability alone as their
accepted transition contract.

The widest segment has width at most `4_294_967_295`; output delta is at most
`1_000_000`. Their product is at most `4_294_967_295_000_000`, below signed-int64
maximum. Final composition is at most `1_000_000_000_000`. After endpoint
clamping all interpolation differences are nonnegative. Thus direct integer
multiply-before-divide fits the parent's intermediate contract even though the
unclamped external Soul range is wider than int32. Future table domains cannot
widen without a contract amendment and both-runtime vectors.

### CT4 — Literal synthetic fixture, not a gameplay mint

Use this small committed **test-only** policy to implement and cross-check CT3:

```json
{
  "obedience_by_trust": {
    "points": [
      {"x": 0, "y_ppm": 0},
      {"x": 800000, "y_ppm": 800000},
      {"x": 1000000, "y_ppm": 1000000}
    ]
  },
  "obedience_factor_by_soul": {
    "points": [{"x": 0, "y_ppm": 0}, {"x": 100, "y_ppm": 1000000}]
  }
}
```

These coefficients are selected for transparent arithmetic, **not** from the
research's win-rate column. They give zero obedience at depleted Soul; this is
a deliberate extreme test arm, not an adopted player experience. The Soul100
endpoint resembles the current Soul artifact's ceiling but neither reads it
dynamically nor authorizes coupling to deploy-current state. Different fixtures
must also exercise nonidentity slopes, intermediate rounding and wide signed
Soul endpoints; this identity-Trust fixture alone cannot prove interpolation.

For this fixture, exact examples are:

| `pet_trust_ppm` | `soul` | `obedience_ppm` |
|---:|---:|---:|
| 1000000 | 100 | 1000000 |
| 800000 | 100 | 800000 |
| 834567 | 37 | 308789 |
| 800001 | 80 | 640000 |
| 1000000 | -9007199254740991 | 0 |
| 800000 | 9007199254740991 | 800000 |

No production epoch may pin this test policy merely because the parent tests
pass. A live content proposal must name its exact points, their gameplay
effects and source policy, obtain applicable adoption/balance review, and prove
the real Pet Care→pinned table→battle workflow. Fixture completion is not that
consumer proof or Pet Care's carried end-to-end AC3.

### CT5 — Implementation acceptance after adoption

1. Strict Go/TS loaders admit the same complete population. Include extra,
   missing, duplicate/escaped-duplicate, null, wrong type, malformed integer,
   unsorted/duplicate point, domain, monotonicity and endpoint refusals.
   Check invalid direct inputs, independent storage and no partial result.
2. Shared literal expected values are derived independently of either runtime.
   Exercise every breakpoint and its integer neighbors, every distinct Soul
   input in the small fixture, trust endpoints and the six examples above.
   Include a nonidentity slope that distinguishes floor from round and an
   example that distinguishes floor-per-stage from one deferred final floor.
   For example, at `x=1` between `(0,0)` and `(3,2)`, floor gives0 and
   nearest-integer rounding gives1. At `x=2` and factor750001, stagewise
   composition is0 whereas deferring both floors gives1. These are additional arithmetic
   controls, not proposed live table points.
3. Exercise signed-int32 table endpoints and both safe-integer Soul extremes;
   old raw-input multiplication/int32-cast variants must not pass the same
   outputs. Assert monotonicity and `[0,1_000_000]` bounds against an independent
   exact-rational reference across the declared finite fixture population.
   Do not describe a finite sample as exhaustive over arbitrary future tables.
4. Both division gates, affected Go/vet, Node/native-browser consumers and
   typecheck pass. These are pure-library tests; no database, live match or
   battle-log parity is claimed without its own actual consumer execution.
5. Parent AC5 may record only its explicit fixture-only scope once these
   checks and designated review pass. Full parent acceptance still requires
   C2/AC1 and its complete reviewed implementation span. No child engine starts
   before DU-C1 is satisfied; no extra primitive is labelled combat gameplay.

## Exact body reconciliation requested on adoption

No author-owned body is edited by this draft. Request explicit adoption of the
table shape/arithmetic and **test-only** fixture, plus delegation to:

- replace design/04 §2's mislabelled `~50%→30% disobedience` clause with:
  “Trust increases the probability that the pet performs the chosen action.
  The reported 50%→30% research figures are simulated win rates, not policy
  probabilities; combat's accepted catalog contract owns the actual mapping.”
- replace parent C5's ambiguous numeric claim with CT1–CT4's exact boundary,
  immutable pinned tables, arithmetic and fixture-only limits;
- clarify parent C3's external-input/int32-state distinction as CT1 states;
- bind AC5 to CT5 without changing C2's unfinished catalog, the other parent
  acceptance criteria, or the deferred child rulings.

This is gameplay-policy structure plus normative-body reconciliation, so
technical review alone cannot adopt it. Live values, authored player-facing
copy, engine transitions and release approval retain their existing gates.

## Changelog

- 2026-11-05: initial draft. Re-derive the source-column error and propose a
  bounded fixture-only care contract without implementing unruled mechanics.
