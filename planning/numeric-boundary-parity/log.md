# Numeric Boundary Parity — Running Log

## 2026-07-28

- Adversarial review found 502 legal input pairs where Go returned NaN and pinned JavaScript
  returned finite state, all in the upper half omitted by the generator's exponent clamp.
- The same review reproduced zero geometric cost for a representably-near-one ratio.
- Follow-up accepted and implementation started. No push.
