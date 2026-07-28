# Numeric Boundary Parity — Running Log

## 2026-07-28

- Adversarial review found 502 legal input pairs where Go returned NaN and pinned JavaScript
  returned finite state, all in the upper half omitted by the generator's exponent clamp.
- The same review reproduced zero geometric cost for a representably-near-one ratio.
- Follow-up accepted and implementation started. No push.
- Direct division and normalized arithmetic-result handling repair the demonstrated division and
  multiplication pairs while preserving diagnostic overflow/underflow behavior.
- Expanding the random exponent domain exposed one additional integer-power lower-bound carry bug;
  the exact integer-power path now retains that mantissa instead of losing it through float
  exponent decomposition.
- Near-one geometric sum/afford now takes the constant-price limit in both runtimes.
- Regenerated schema-3 vectors contain 6,295 cases, 22 mandatory edges, and enforced upper-half
  binary coverage. `make verify` passed with 19,038 browser tests across all engines.
- No push performed.
