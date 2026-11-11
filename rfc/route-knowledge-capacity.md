# RFC: Route Knowledge Capacity

- **Status:** draft — owner direction pending
- **Author:** Codex
- **Created:** 2026-11-11
- **Design refs:** `design/02` §3.2–3.3 and `design/11` §3
- **Parent:** [Prestige & Exits](prestige-and-exits.md) D1–D4, P3, AC2/AC5; [Gate Predicates and Routes](archive/gate-predicates-and-routes.md) D5/C6; [Minigame Platform](minigame-platform-foundation.md) C24/C28
- **Evidence:** [RP-771](../planning/prestige-and-exits/route-knowledge-capacity.md)

## Summary

An otherwise eligible Wind Down currently fails if a positive grant exceeds the
Founder's Route Knowledge headroom. The same capacity limit can independently
stop the balance projection. Choose how Route Knowledge accumulates before
changing the Exit or projection. The recommended direction preserves earned
grants and the existing Exit-offer guarantee through an exact wider representation.

## Recommended contract — preserve grants exactly

1. Keep the existing catalog inputs, hint prices, grant rules and compensation
   authority. Compute the offered floor, recomputed payout and scripted curriculum
   bonus in exact integers. For a valid existing offer, apply the promised Route
   Knowledge floor even if its historical JSON used a safe numeric value. A single
   scripted Exit may combine two individually valid safe-integer inputs whose sum
   needs a wider representation. The versioned ending and Founder credit must
   report that same exact payout.
2. Use exact, nonnegative integers for saved Route Knowledge and exact signed
   arithmetic for projection. The projection still computes uncompensated route
   grants plus committed Exit credits minus hint debits. Negative net is existing
   Registry debt and carries no spendable balance. Later grants pay that debt
   first. Existing hint purchases stay permanent. Delivery, duplicate claims
   and rebuild from immutable facts must converge to the same amount. The Founder
   save supplies no additional projector credit.
3. Use canonical unsigned decimal strings for the versioned quantities that can
   exceed the safe-integer domain. Zero is `"0"`. Reject signs, fractions,
   exponents and leading zeros. Go and TypeScript consumers must preserve the
   integer end to end, including formatting, without converting through a
   floating-point number. Use Go `math/big.Int`, TypeScript `bigint`, and
   integer-valued Postgres `numeric` for the widened arithmetic and storage.
   Read SQL quantities as decimal text. Validate versioned source quantities
   before casting or claiming them. Existing bounded catalog fields, frame
   limits and unrelated counters keep their domains. Route Knowledge no longer
   has a gameplay accumulation cap under this option.
4. Use an append-only compatibility path. Old saved numbers, immutable events,
   receipts, commands and replay results retain their historical readers and
   bytes. New writes use an explicit version at each changed boundary. Existing
   eligible Founders must be able to cross the boundary with their current
   run, saved balance and any live offer intact. Before implementation, the
   owning specification must name the activation discriminator, versioned
   boundaries and reader rollout order, including both replayers and a browser
   already open across an upgrade. The live event reader must have an explicit
   discriminator it can actually see. Replaying an old command selects its
   recorded version. The [deployment rollback](deployment-foundation.md#dp5--stop-drain-start-release-and-rollback)
   continues to restore the pre-upgrade backup on clean Postgres.
5. Keep the existing atomic Exit write and replayed receipt. The changed
   boundary includes the Founder save and carry, Company and Founder resolved
   replay inputs, stored offers and `run_ended` payout, `founder_advanced`
   credit, hint resolved inputs, SQL balances, debt and claim and rebuild reads,
   and client event rendering. The Route Knowledge copy frame takes exact
   validated digits through a string parameter while retaining the authored
   template and the other copy parameters' existing type checks.
6. For minigame scaling, admit exact Route Knowledge at the source and apply
   the existing C28 operation before clamping to the declared output bounds.
   Keep the existing output, operand and other source domains and the ranked
   fairness rules. A minigame whose policy does not use Route Knowledge must
   still start for a Founder holding a widened balance. Previously resolved
   scaling inputs stay fixed. Both first-hour and Tier 2 pacing measurements
   must preserve the same exact arithmetic for their current Exit models.

## Alternative — a visible hard cap with immediate loss of excess

Keep the safe-integer maximum as a visible hard cap. Pay the amount that fits
and discard the rest at settlement. Record and present the earned and credited
amounts separately. This changes the meaning of a positive offer promise unless
capacity was reserved or the owner explicitly changes that guarantee. Newly
generated acquisition and acquihire offers currently carry zero Route Knowledge.
Reservation can govern future promises. The owner must separately choose what
happens to already-live positive promises whose headroom is insufficient.

This alternative requires a separate cap contract before implementation.
It must name one authoritative order for credits and debits, persist the
actual credited amount, and define what a compensation reverses when the
original grant was clipped. Clipping each arrival in the current projection
loses different amounts for different delivery order. Applying a cap only to
the display or to the final net lets earlier overflow reappear after a purchase.

## Acceptance criteria

1. The retained real-Postgres eligible Wind Down cases succeed at the old
   maximum, including the Founder who owns all current hints. A positive
   stored offer must satisfy its selected guarantee with insufficient old
   headroom. Run an actual scripted Exit whose combined earned grant exceeds
   the old numeric boundary. Verify the exact credit for the wider option or
   separately verify the earned and credited amounts under a cap.
2. The same declared inputs produce equal complete states, receipts and
   events in Go and TypeScript across the old boundary. Adjacent integers
   above it remain distinct. Historical replay and archived evidence retain
   their old bytes. Test a live old-run transition and an interrupted or
   retried write.
3. In real Postgres, reordered and repeated delivery, missing cache rebuild,
   late Exit and hint callbacks and compensation converge under the selected
   policy. Test debt repayment and spending above the old boundary for exact
   accumulation, with complete persisted claims and rollback on malformed input.
4. The mounted browser displays exact promised and received amounts using the
   real event path and copy frame. If a gameplay cap is selected, show the cap
   and discarded amount. The owner adopts any new player-facing copy separately.
5. For the exact option, start an eligible minigame for a Founder above the
   old boundary with a policy that ignores the counter, then exercise a
   declared Route Knowledge source with an operator that distinguishes exact
   input from an early clamp. Compare resolved session inputs at and above the
   old boundary. Verify first-hour and Tier 2 pacing Exit outcomes against the
   selected grant rule, with ordinary-value controls.

## Owner decision

Choose whether to preserve all Route Knowledge grants exactly (recommended)
or set a visible hard cap that discards excess. The cap direction also needs
an explicit ruling on the guarantee for positive stored offers. This draft
is pending adoption.
