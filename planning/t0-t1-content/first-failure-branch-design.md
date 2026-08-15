# Design proposal: branch the scripted first failure (T01-C31 arm)

- **Status:** draft for owner ruling. Authorizes nothing; copy is drafted for owner adoption.
- **Author:** Claude (design lane). Owner direction 2026-08-15: "shouldn't they pivot, get
  acquired, fail, or w/e" — the collapse should be caused by the run, not announced by a timer.
- **Supersedes:** the single-outcome arm of T01-C31 (automatic beat, one generic collapse).

## What already exists (so we build the smallest real thing)

- The trigger is **already content-gated**, not a bare timer: run 1, zero Founder Exits,
  `gate.t0_to_t1` crossed, 900,000 attended ms. You must have scaled into the garage.
- The game already ships exit flavours with different payout mixes — `acquihire`, `acquisition`,
  `ipo`, `collapse`, `scripted_first` — in `balance/prestige/phase0.json`.
- design/02: "collapse pays less Reputation but more Route Knowledge (*every death teaches you the
  map*)". design/01: Tier-1-appropriate fictions are the early ones; IPO would be absurd here.

## The binding constraint, and how the branch respects it

design/11 §6b (ruled): **"The scripted failure pays full first-Exit Reputation, so it is a *beat*,
never a tax."** Therefore:

- The terminal exit type stays mechanically `scripted_first` (modifier `1000000`, full payout) in
  **every** branch. `balance/prestige/phase0.json` is a byte-preserved carryover artifact in the
  ratified epoch-7 manifest — **this design does not touch it**, so no manifest row, constants
  hash, or ratification reopens.
- Branches therefore differ in **narrative + a non-negative Route-Knowledge bonus**. No branch is
  ever worse than today's beat; some are better. That teaches "exits differ" in the direction the
  no-tax rule permits, and it matches design/02's collapse-teaches-the-map shape without ever
  reducing Reputation.

## The three outcomes and their predicates

Evaluated on the company's own state at the trigger, in this fixed order, first match wins.
All inputs are already persisted fields — **no new mechanics, no new tracked state**:

| # | Outcome | Predicate (at trigger) | Fiction | RK bonus |
|---|---|---|---|---|
| 1 | **Acquihire** — "they bought the people" | `GeneratorPurchasedTotal` ≥ **A** and ≥ one upgrade owned | You built a real machine; somebody noticed and bought the team out from under it. | 0 |
| 2 | **Burnout / collapse** — "you outscaled the money" | cash balance < **B** × the cheapest unowned generator's current price (you spent to the floor) | You scaled faster than the revenue. The lights were on; the account was not. | +**R** |
| 3 | **Pivot** — "you were the company" | default (neither above) | Nothing here ran without you touching it. That is not a company; that is a person with a keyboard. | +**R**/2 |

`A`, `B`, `R` are balance data, derived by measurement across the three ruled personas so that
**each branch is reachable** — the acceptance requirement below.

## Acceptance (so this cannot become decorative)

1. Each of the three outcomes is reached by at least one of the ruled first-hour personas at the
   pinned seed; the pacing report records which. A branch no persona can reach is a finding, not
   a feature.
2. The branch decision is a pure function of persisted state, recorded in the `run_ended` payload
   and **replayed byte-identically** (same discipline as every other terminal transition).
3. Reputation granted is identical across branches; a mutation reducing it in any branch fails.
4. Removing the branch selector must fail a test — the outcome is asserted, not merely emitted.

## Copy drafts (owner-authored content — for adoption, not for shipping as-is)

Title stays **"Your First Company Failed"** for outcomes 2 and 3; outcome 1 needs its own title
because it is not a failure. Existing ratified body/next-run copy is retained for the default
path where noted.

**1 — Acquihire.** Title: *"Your First Company Was Acquired"*
Body: "Someone with a bigger office read your rate sheet and made an offer for the people, not
the product. The product was turned off on a Tuesday. You are told this is a success, and by
every metric that gets published, it is."

**2 — Burnout / collapse.** Title: *"Your First Company Failed"* (ratified title retained)
Body: "You bought the future faster than the present could pay for it. The machines are still
warm. Statistically, this is the most realistic thing in this game."
*(This is the closest to the existing ratified body and deliberately keeps its final sentence —
the survivorship-bias joke is the beat and should not move.)*

**3 — Pivot.** Title: *"Your First Company Pivoted"*
Body: "Nothing here ran unless you were touching it, so what you actually built was a job. The
pivot is the part where you admit that and keep the contacts."

Shared next-run line (existing ratified copy, unchanged): "The second company gets founded by
someone who has already failed once. Historically, that is the good one."

## Cost and sequencing (stated plainly)

This is scope expansion at the RFC's closing stage. It delays the T01 close by one design+
implementation+review cycle beyond the single-outcome arm. It does **not** reopen any ratified
hash, and it does not touch prestige or the epoch-7 manifest. AC4's composed proof consumes
whichever branch its pinned seed produces, so the two land together.

## Owner decisions requested

1. Accept the three outcomes and their predicate shapes (literals `A`/`B`/`R` measured after).
2. Adopt, edit, or replace the three copy drafts (owner-authored content).
3. Confirm the no-tax reading: identical Reputation in all branches, Route Knowledge as a
   non-negative bonus only.
